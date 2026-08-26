package telegram

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/savio/savio/backend/internal/migrations"
	"github.com/savio/savio/backend/internal/users"
	"github.com/savio/savio/backend/internal/workspaces"
)

func TestMain(m *testing.M) {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		testURL, err := url.Parse(v)
		if err == nil {
			testURL.Path = "/savio_test_telegram"
			ensureTestDB(v, testURL.String())
			os.Setenv("DATABASE_URL", testURL.String())
		}
	}
	openDB()
	code := m.Run()
	os.Exit(code)
}

var db *gorm.DB

func openDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return
	}
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		panic(err)
	}
}

func ensureTestDB(adminDSN, testDSN string) {
	c, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return
	}
	defer c.Close()
	_, _ = c.Exec(`CREATE DATABASE savio_test_telegram`)
	if err := migrateTestDB(testDSN); err != nil {
		panic(err)
	}
}

func migrateTestDB(dsn string) error {
	src, err := iofs.New(migrations.FS, migrations.Dir)
	if err != nil {
		return err
	}
	defer src.Close()
	mm, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return err
	}
	defer mm.Close()
	for {
		_, _, verr := mm.Version()
		if errors.Is(verr, migrate.ErrNilVersion) {
			break
		}
		if verr != nil {
			return verr
		}
		if err := mm.Steps(-1); err != nil {
			return err
		}
	}
	return mm.Up()
}

func createUser(t *testing.T, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	err := db.Create(&users.User{ID: id, Name: name, Email: name + "-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}

func fixture(t *testing.T) (svc *Service, wsID, ownerID, memberID uuid.UUID) {
	t.Helper()
	svc = NewService(db, nil, nil)
	wsID = uuid.New()
	if err := db.Create(&workspaces.Workspace{ID: wsID, Name: "T", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error; err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	ownerID = createUser(t, "Owner")
	memberID = createUser(t, "Member")
	for _, m := range []struct {
		user uuid.UUID
		role string
	}{{ownerID, "OWNER"}, {memberID, "MEMBER"}} {
		if err := db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: m.user, Role: m.role, Status: "ACTIVE"}).Error; err != nil {
			t.Fatalf("create membership: %v", err)
		}
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM telegram_settings WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id IN ($1, $2)`, ownerID, memberID)
	})
	return svc, wsID, ownerID, memberID
}

func TestRegisterWebhookRequiresEnabledBotAndChat(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	svc, wsID, _, member := fixture(t)

	st := &Settings{WorkspaceID: wsID, Enabled: false, BotToken: "111:abc", ChatID: "42"}
	if err := db.Create(st).Error; err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	_, err := svc.RegisterWebhook(context.Background(), wsID, member, "https://a.ngrok-free.app")
	if err == nil || !strings.Contains(err.Error(), "aktifkan") {
		t.Fatalf("disabled bot must be rejected, got %v", err)
	}

	if err := db.Model(st).Update("enabled", true).Error; err != nil {
		t.Fatal(err)
	}
	st.ChatID = ""
	if err := db.Model(st).Update("chat_id", "").Error; err != nil {
		t.Fatal(err)
	}
	_, err = svc.RegisterWebhook(context.Background(), wsID, member, "https://a.ngrok-free.app")
	if err == nil || !strings.Contains(err.Error(), "chat_id") {
		t.Fatalf("missing chat_id must be rejected, got %v", err)
	}
}

func TestRegisterWebhookRejectsTokenUsedByAnotherWorkspace(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	svc, wsID, _, member := fixture(t)

	other := uuid.New()
	if err := db.Create(&workspaces.Workspace{ID: other, Name: "Other", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error; err != nil {
		t.Fatal(err)
	}
	token := "999:shared"
	if err := db.Create(&Settings{WorkspaceID: other, Enabled: true, BotToken: token, ChatID: "7",
		WebhookURL: "https://a.ngrok-free.app", WebhookSecret: "s3cr3t"}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec(`DELETE FROM telegram_settings WHERE workspace_id = $1`, other) })

	if err := db.Create(&Settings{WorkspaceID: wsID, Enabled: true, BotToken: token, ChatID: "42"}).Error; err != nil {
		t.Fatal(err)
	}
	_, err := svc.RegisterWebhook(context.Background(), wsID, member, "https://b.ngrok-free.app")
	if err == nil || !strings.Contains(err.Error(), "ruang kerja lain") {
		t.Fatalf("shared bot token must be rejected, got %v", err)
	}
}

func TestAuthorizerPrefersConfiguredByUser(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	svc, wsID, owner, member := fixture(t)
	ctx := context.Background()

	cfg := member
	gotID, gotName, err := svc.authorizerFor(ctx, wsID, &cfg)
	if err != nil || gotID != member || gotName != "Member" {
		t.Fatalf("configured member expected, got %s/%s err=%v", gotID, gotName, err)
	}

	left := uuid.New()
	gotID, _, err = svc.authorizerFor(ctx, wsID, &left)
	if err != nil || gotID != owner {
		t.Fatalf("non-member configured_by must fall back to owner %s, got %s err=%v", owner, gotID, err)
	}

	gotID, _, err = svc.authorizerFor(ctx, wsID, nil)
	if err != nil || gotID != owner {
		t.Fatalf("nil configured_by must fall back to owner %s, got %s err=%v", owner, gotID, err)
	}
}
