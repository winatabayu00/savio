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

// Poll fetches pending updates once via long-poll for every enabled workspace
// bot. The worker calls it in a tight loop. It is exactly-once per update per
// workspace (telegram_processed unique guard), even though multiple worker
// replicas could poll concurrently. Webhook mode (config.webhook_url set)
// replaces polling entirely.
func (s *Service) Poll(ctx context.Context) error {
	var settings []Settings
	if err := s.db.WithContext(ctx).Where("enabled = TRUE AND bot_token <> '' AND webhook_url = ''").
		Find(&settings).Error; err != nil {
		return err
	}
	for i := range settings {
		st := settings[i]
		client := newBotClient(st.BotToken)
		ups, err := client.updates(ctx, st.LastUpdateID+1)
		if err != nil {
			slog.Error("telegram poll", "error", err.Error(), "workspace_id", st.WorkspaceID)
			continue
		}
		if len(ups) == 0 {
			continue
		}
		maxID := st.LastUpdateID
		for j := range ups {
			u := ups[j]
			if u.ID <= st.LastUpdateID {
				continue
			}
			if u.ID > maxID {
				maxID = u.ID
			}
			if !s.claim(ctx, st.WorkspaceID, u.ID) {
				continue // already handled by another tick
			}
			if err := s.Handle(ctx, &st, u, client); err != nil {
				slog.Error("telegram handler", "error", err.Error(), "update_id", u.ID, "workspace_id", st.WorkspaceID)
			}
		}
		if maxID != st.LastUpdateID {
			if err := s.db.WithContext(ctx).Model(&Settings{WorkspaceID: st.WorkspaceID}).
				Update("last_update_id", maxID).Error; err != nil {
				slog.Error("telegram offset", "error", err.Error())
			}
		}
	}
	return nil
}

// Handle processes one Telegram update (shared by long-poll and webhook modes).
// The chat is authorized by the configured chat_id; a secret-less webhook never
// bypasses this.
func (s *Service) Handle(ctx context.Context, st *Settings, u Update, client *botClient) error {
	if !st.Enabled || st.BotToken == "" || u.Message == nil {
		return nil
	}
	if st.ChatID != "" && strconv.FormatInt(u.Message.Chat.ID, 10) != st.ChatID {
		return nil
	}
	return s.handle(ctx, st, u, client)
}

// claim inserts the (workspace, update id) pair with a unique constraint so a
// crash between "new update" and "offset ack" cannot duplicate a transaction.
// Returns false when the id was already claimed.
func (s *Service) claim(ctx context.Context, workspaceID uuid.UUID, id int64) bool {
	res := s.db.WithContext(ctx).Exec(
		`INSERT INTO telegram_processed (workspace_id, update_id) VALUES (?, ?) ON CONFLICT (workspace_id, update_id) DO NOTHING`, workspaceID, id)
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

	authorizer, userName, err := s.authorizerFor(ctx, wsID, st.ConfiguredByUserID)
	if err != nil {
		return client.sendMessage(ctx, chatID, "Gagal membaca ruang kerja. Pastikan konfigurasi Telegram valid.")
	}

	acctID, err := s.defaultAccountID(ctx, wsID)
	if err != nil {
		return client.sendMessage(ctx, chatID, "Belum ada akun aktif untuk ruang kerja ini. Buat akun dulu di aplikasi.")
	}

	catName := categorizeKeyword(desc)
	// ponytail: mock provider returns a canned "Food & Dining" for every input;
	// never let the demo facade override deterministic keyword categorization.
	if !s.ai.UsesMockProvider(ctx) {
		// Real AI sees a JSON reference of candidate categories AND active
		// accounts, so it can also say which wallet the entry belongs to.
		// A wrong account guess only falls back to the default — the AI never
		// writes finance state (AGENTS #67).
		if res, cerr := s.ai.CategorizeEntry(ctx, wsID, desc, ""); cerr == nil && res != nil {
			if res.CategoryGuess != "" {
				catName = res.CategoryGuess
			}
			if res.AccountGuess != "" {
				acctID = s.findAccountIDByName(ctx, wsID, res.AccountGuess, acctID)
			}
		}
	}
	catID := resolveCategoryID(ctx, s.db, wsID, catName)

	txDate, err := s.workspaceDate(ctx, wsID)
	if err != nil {
		return err
	}
	_, err = s.tx.Create(ctx, wsID, authorizer, &transactions.CreateInput{
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
	name := strings.TrimSpace(userName)
	if name == "" {
		name = "teman"
	}
	reply := fmt.Sprintf(
		"✅ Halo %s, pencatatan pengeluaran sebesar %s sudah berhasil dicatat.\n\nKategori: %s\nCatatan: %s",
		name, money.FormatMinorUnits(amountMinor), catName, desc)
	return client.sendMessage(ctx, chatID, reply)
}

const helpText = "Cara pakai:\nKirim nama pengeluaran + nominal di akhir.\nContoh: `chocolate hazelnut dutch 24000`\nNominal boleh pakai titik/koma: `grab food 24.500`, `pulsa 50rb`, `kopi 2.5jt`."

func (s *Service) ownerUserID(ctx context.Context, wsID uuid.UUID) (uuid.UUID, string, error) {
	var owner struct {
		UserID uuid.UUID
		Name   string
	}
	err := s.db.WithContext(ctx).Table("workspace_memberships wm").
		Select("wm.user_id, u.name").
		Joins("JOIN users u ON u.id = wm.user_id").
		Where("wm.workspace_id = ? AND wm.role = 'OWNER' AND wm.status = 'ACTIVE'", wsID).
		Order("wm.created_at").Limit(1).Take(&owner).Error
	return owner.UserID, owner.Name, err
}

// authorizerFor attributes Telegram-driven transactions to the user who
// configured the bot (the logged-in user at setup), falling back to the first
// OWNER when they left the workspace.
func (s *Service) authorizerFor(ctx context.Context, wsID uuid.UUID, preferred *uuid.UUID) (uuid.UUID, string, error) {
	if preferred != nil && *preferred != uuid.Nil {
		var m struct {
			UserID uuid.UUID
			Name   string
		}
		err := s.db.WithContext(ctx).Table("workspace_memberships wm").
			Select("wm.user_id, u.name").
			Joins("JOIN users u ON u.id = wm.user_id").
			Where("wm.workspace_id = ? AND wm.user_id = ? AND wm.status = 'ACTIVE'", wsID, *preferred).
			Take(&m).Error
		if err == nil {
			return m.UserID, m.Name, nil
		}
	}
	return s.ownerUserID(ctx, wsID)
}

func (s *Service) defaultAccountID(ctx context.Context, wsID uuid.UUID) (uuid.UUID, error) {
	var acc struct{ ID uuid.UUID }
	err := s.db.WithContext(ctx).Table("accounts").
		Where("workspace_id = ? AND status = 'ACTIVE'", wsID).
		Order("created_at").Limit(1).Select("id").Take(&acc).Error
	return acc.ID, err
}

func (s *Service) findAccountIDByName(ctx context.Context, wsID uuid.UUID, name string, fallback uuid.UUID) uuid.UUID {
	var acc struct{ ID uuid.UUID }
	if err := s.db.WithContext(ctx).Table("accounts").
		Where("workspace_id = ? AND status = 'ACTIVE' AND name = ?", wsID, name).
		Select("id").Take(&acc).Error; err == nil && acc.ID != uuid.Nil {
		return acc.ID
	}
	return fallback
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
	var row struct{ ID uuid.UUID }
	lookup := func(name string) bool {
		err := db.WithContext(ctx).Table("categories").
			Where("name = ? AND type = 'EXPENSE' AND status = 'ACTIVE' AND (is_system = TRUE OR workspace_id = ?)", name, wsID).
			Limit(1).Select("id").Take(&row).Error
		return err == nil && row.ID != uuid.Nil
	}
	if lookup(preferred) {
		return &row.ID
	}
	if lookup("Other Expense") {
		return &row.ID
	}
	return nil
}
