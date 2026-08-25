package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/ai"
	"github.com/savio/savio/backend/internal/platform/money"
	"github.com/savio/savio/backend/internal/transactions"
)

// Service implements the Telegram recap bot: it long-polls Telegram and writes
// parsed expense messages into the configured workspace as POSTED transactions.
// Finance limits to the bound workspace (AGENTS #19/#25) and never lets
// Telegram-provided ids select resources across workspaces.
type Service struct {
	db *gorm.DB
	ai *ai.Service
	tx *transactions.Service
}

func NewService(db *gorm.DB, aiSvc *ai.Service, txSvc *transactions.Service) *Service {
	return &Service{db: db, ai: aiSvc, tx: txSvc}
}

// Poll fetches pending updates once. The worker calls it in a tight loop. It is
// exactly-once per update (telegram_processed unique guard), even though
// multiple worker replicas could poll concurrently.
func (s *Service) Poll(ctx context.Context) error {
	st, err := s.load(ctx)
	if err != nil {
		return err
	}
	if !st.Enabled || st.BotToken == "" {
		return nil
	}
	client := newBotClient(st.BotToken)
	ups, err := client.updates(ctx, st.LastUpdateID+1)
	if err != nil {
		return err
	}
	if len(ups) == 0 {
		return nil
	}
	maxID := st.LastUpdateID
	for i := range ups {
		u := ups[i]
		if u.ID <= st.LastUpdateID {
			continue
		}
		if u.ID > maxID {
			maxID = u.ID
		}
		if !s.claim(ctx, u.ID) {
			continue // already handled by another tick
		}
		if u.Message == nil {
			continue
		}
		// ponytail: chat is authorized by configured chat_id; multi-bot fan-out
		// per chat would be the upgrade path if this grows beyond one bot.
		if st.ChatID != "" && strconv.FormatInt(u.Message.Chat.ID, 10) != st.ChatID {
			continue
		}
		if err := s.handle(ctx, st, u, client); err != nil {
			slog.Error("telegram handler", "error", err.Error(), "update_id", u.ID)
		}
	}
	if maxID != st.LastUpdateID {
		if err := s.db.WithContext(ctx).Model(&Settings{ID: 1}).Update("last_update_id", maxID).Error; err != nil {
			slog.Error("telegram offset", "error", err.Error())
		}
	}
	return nil
}

// claim inserts the update id with a unique constraint so a crash between "new
// update" and "offset ack" cannot duplicate a transaction. Returns false when
// the id was already claimed.
func (s *Service) claim(ctx context.Context, id int64) bool {
	res := s.db.WithContext(ctx).Exec(
		`INSERT INTO telegram_processed (update_id) VALUES (?) ON CONFLICT (update_id) DO NOTHING`, id)
	return res.Error == nil && res.RowsAffected == 1
}

func (s *Service) handle(ctx context.Context, st *Settings, u Update, client *botClient) error {
	chatID := strconv.FormatInt(u.Message.Chat.ID, 10)
	text := strings.TrimSpace(u.Message.Text)
	if text == "" || text == "/start" || text == "/help" {
		return client.sendMessage(ctx, chatID, helpText)
	}
	if st.WorkspaceID == uuid.Nil {
		return client.sendMessage(ctx, chatID, helpText)
	}
	desc, amountMinor, ok := parseExpense(text)
	if !ok {
		return client.sendMessage(ctx, chatID, "Format nominal belum dikenali. Contoh: `kopi 25000` atau `grab food 45.500` (nama di depan, nominal di akhir).")
	}
	wsID := st.WorkspaceID

	authorizer, err := s.ownerUserID(ctx, wsID)
	if err != nil {
		return client.sendMessage(ctx, chatID, "Gagal membaca ruang kerja. Pastikan konfigurasi Telegram valid.")
	}

	acctID, err := s.defaultAccountID(ctx, wsID)
	if err != nil {
		return client.sendMessage(ctx, chatID, "Belum ada akun aktif untuk ruang kerja ini. Buat akun dulu di aplikasi.")
	}

	catName := categorizeKeyword(desc)
	if guess, cerr := s.ai.Categorize(ctx, wsID, desc, ""); cerr == nil && guess.CategoryGuess != "" {
		catName = guess.CategoryGuess
	}
	catID := resolveCategoryID(ctx, s.db, wsID, catName)

	txDate, err := s.workspaceDate(ctx, wsID)
	if err != nil {
		return err
	}
	view, err := s.tx.Create(ctx, wsID, authorizer, &transactions.CreateInput{
		AccountID:       acctID,
		CategoryID:      catID,
		Type:            string(transactions.TypeExpense),
		AmountMinor:     amountMinor,
		TransactionDate: txDate,
		Description:     desc,
		Source:          string(transactions.SourceTelegram),
		Status:          string(transactions.StatusPosted),
	})
	if err != nil {
		return client.sendMessage(ctx, chatID, fmt.Sprintf("Gagal mencatat pengeluaran: %s", err.Error()))
	}
	reply := fmt.Sprintf(
		"✅ %s · %s\n%s\nID: %s",
		money.FormatMinorUnits(amountMinor), catName, desc, view.ID)
	return client.sendMessage(ctx, chatID, reply)
}

const helpText = "Cara pakai:\nKirim nama pengeluaran + nominal di akhir.\nContoh: `chocolate hazelnut dutch 24000`\nNominal boleh pakai titik/koma: `grab food 24.500`, `pulsa 50rb`, `kopi 2.5jt`."

func (s *Service) ownerUserID(ctx context.Context, wsID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.db.WithContext(ctx).Table("workspace_memberships").
		Where("workspace_id = ? AND role = 'OWNER' AND status = 'ACTIVE'", wsID).
		Order("created_at").Limit(1).Pluck("user_id", &id).Error
	return id, err
}

func (s *Service) defaultAccountID(ctx context.Context, wsID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.db.WithContext(ctx).Table("accounts").
		Where("workspace_id = ? AND status = 'ACTIVE'", wsID).
		Order("created_at").Limit(1).Pluck("id", &id).Error
	return id, err
}

// workspaceDate returns today in the workspace timezone so the date-only field
// does not shift with UTC (AGENTS #20).
func (s *Service) workspaceDate(ctx context.Context, wsID uuid.UUID) (time.Time, error) {
	var tz string
	if err := s.db.WithContext(ctx).Table("workspaces").
		Where("id = ?", wsID).Pluck("timezone", &tz).Error; err != nil {
		return time.Time{}, err
	}
	if loc, err := time.LoadLocation(tz); err == nil {
		now := time.Now().In(loc)
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc), nil
	}
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
}

// categorizeKeyword is the deterministic, AI-free fallback so the bot still
// categorizes when AI is disabled or down (AGENTS #77, #236).
func categorizeKeyword(desc string) string {
	s := strings.ToLower(desc)
	contains := func(keys ...string) bool {
		for _, k := range keys {
			if strings.Contains(s, k) {
				return true
			}
		}
		return false
	}
	switch {
	case contains("makan", "kopi", "coffee", "cafe", "restoran", "nasi", "ayam", "burger", "pizza", "sushi", "bakso", "mie", "chocolate", "dutch", "camilan", "gofood", "grabfood", "grab food", "go food"):
		return "Food & Dining"
	case contains("grab", "gojek", "goride", "go ride", "ojek", "gocar", "bensin", "toll", "parkir", "transjakarta", "commuter"):
		return "Transportation"
	case contains("listrik", "pln", "pdam", "pulsa", "internet", "wifi", "telkom", "indihome", "bpjs"):
		return "Utilities"
	case contains("pasar", "alfamart", "indomaret", "buah", "sayur", "beras", "susu"):
		return "Groceries"
	case contains("obat", "apotek", "klinik", "dokter", "vitamin", "rumah sakit"):
		return "Health"
	case contains("shopee", "tokopedia", "lazada", "baju", "sepatu", "tas"):
		return "Shopping"
	case contains("netflix", "spotify", "youtube", "xxi", "bioskop", "steam"):
		return "Entertainment"
	case contains("salon", "barbershop", "skincare", "parfum", "potong rambut"):
		return "Personal Care"
	}
	return "Other Expense"
}

// resolveCategoryID maps a preferred category NAME to a usable category id in the
// workspace (system or workspace-scoped). Falls back to "Other Expense". Returns
// nil when nothing matches, leaving the transaction uncategorized.
func resolveCategoryID(ctx context.Context, db *gorm.DB, wsID uuid.UUID, preferred string) *uuid.UUID {
	var id uuid.UUID
	lookup := func(name string) bool {
		return db.WithContext(ctx).Table("categories").
			Where("name = ? AND type = 'EXPENSE' AND status = 'ACTIVE' AND (is_system = TRUE OR workspace_id = ?)", name, wsID).
			Limit(1).Pluck("id", &id).Error == nil && id != uuid.Nil
	}
	if lookup(preferred) {
		return &id
	}
	if lookup("Other Expense") {
		return &id
	}
	return nil
}
