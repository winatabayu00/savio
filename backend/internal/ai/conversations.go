package ai

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
)

type Conversation struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	WorkspaceID uuid.UUID `json:"-" gorm:"type:uuid"`
	UserID      uuid.UUID `json:"-" gorm:"type:uuid"`
	Title       *string   `json:"title"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Messages    []Message `json:"messages,omitempty" gorm:"foreignKey:ConversationID"`
}

func (Conversation) TableName() string { return "ai_conversations" }

type Message struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	ConversationID uuid.UUID       `json:"-" gorm:"type:uuid"`
	Role           string          `json:"role"`
	Content        string          `json:"content"`
	Response       json.RawMessage `json:"-" gorm:"type:jsonb"`
	CreatedAt      time.Time       `json:"created_at"`
}

func (Message) TableName() string { return "ai_messages" }

func (m Message) MarshalJSON() ([]byte, error) {
	type message Message
	out := struct {
		message
		Response *CopilotResult `json:"response,omitempty"`
	}{message: message(m)}
	if len(m.Response) > 0 {
		out.Response = new(CopilotResult)
		if err := json.Unmarshal(m.Response, out.Response); err != nil {
			return nil, err
		}
	}
	return json.Marshal(out)
}

func (s *Service) CreateConversation(ctx context.Context, workspaceID, userID uuid.UUID) (*Conversation, error) {
	c := &Conversation{ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID}
	if err := s.db.WithContext(ctx).Create(c).Error; err != nil {
		return nil, err
	}
	c.Messages = []Message{}
	return c, nil
}

func (s *Service) ListConversations(ctx context.Context, workspaceID, userID uuid.UUID) ([]Conversation, error) {
	var rows []Conversation
	err := s.db.WithContext(ctx).Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Order("updated_at DESC").Limit(100).Find(&rows).Error
	return rows, err
}

func (s *Service) Conversation(ctx context.Context, workspaceID, userID, id uuid.UUID) (*Conversation, error) {
	var c Conversation
	err := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ? AND user_id = ?", id, workspaceID, userID).
		Preload("Messages", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC, id ASC") }).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errs.NotFound("Conversation not found")
	}
	return &c, err
}

func (s *Service) DeleteConversation(ctx context.Context, workspaceID, userID, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Where("id = ? AND workspace_id = ? AND user_id = ?", id, workspaceID, userID).Delete(&Conversation{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errs.NotFound("Conversation not found")
	}
	return nil
}

func (s *Service) SendMessage(ctx context.Context, workspaceID, userID, id uuid.UUID, question string, horizon int, now time.Time) (*Conversation, error) {
	question = strings.TrimSpace(question)
	if question == "" || len(question) > 2000 {
		return nil, errs.ValidationFields(map[string]string{"question": "Question must be between 1 and 2000 characters."})
	}
	c, err := s.Conversation(ctx, workspaceID, userID, id)
	if err != nil {
		return nil, err
	}
	result, err := s.Copilot(ctx, workspaceID, question, horizon, now)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	now = now.UTC()
	userMessage := Message{ID: uuid.New(), ConversationID: id, Role: "USER", Content: question, CreatedAt: now}
	assistantMessage := Message{ID: uuid.New(), ConversationID: id, Role: "ASSISTANT", Content: result.Answer, Response: encoded, CreatedAt: now.Add(time.Nanosecond)}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&userMessage).Error; err != nil {
			return err
		}
		if err := tx.Create(&assistantMessage).Error; err != nil {
			return err
		}
		updates := map[string]any{"updated_at": now}
		if c.Title == nil {
			title := truncate(question, 80)
			updates["title"] = title
		}
		return tx.Model(&Conversation{}).Where("id = ? AND workspace_id = ? AND user_id = ?", id, workspaceID, userID).Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	return s.Conversation(ctx, workspaceID, userID, id)
}
