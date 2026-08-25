package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/google/uuid"

	"github.com/savio/savio/backend/internal/platform/database"
	"github.com/savio/savio/backend/internal/telegram"
)

// telegramwebhook registers (or removes) the Telegram bot webhook for a
// workspace directly against the DB, bypassing HTTP auth — an ops convenience
// for dev. Usage:
//
//	go run ./cmd/telegramwebhook -workspace <id> -url https://xxx.ngrok-free.app
//	go run ./cmd/telegramwebhook -workspace <id>          # removes webhook
func main() {
	workspace := flag.String("workspace", "", "workspace UUID owning the bot")
	url := flag.String("url", "", "public HTTPS base URL to register (empty removes)")
	flag.Parse()

	if *workspace == "" {
		log.Fatal("usage: telegramwebhook -workspace <id> [-url https://...]")
	}
	wsID, err := uuid.Parse(*workspace)
	if err != nil {
		log.Fatalf("bad workspace id: %v", err)
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://savio:savio@localhost:5433/savio?sslmode=disable"
	}
	db, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}

	ctx := context.Background()
	svc := telegram.NewService(db, nil, nil)

	if *url == "" {
		st, err := svc.RegisterWebhook(ctx, wsID, "")
		if err != nil {
			log.Fatalf("remove webhook: %v", err)
		}
		log.Printf("webhook removed for workspace %s", st.WorkspaceID)
		return
	}

	st, err := svc.RegisterWebhook(ctx, wsID, *url)
	if err != nil {
		log.Fatalf("register webhook: %v", err)
	}
	if st.WebhookSecret == "" {
		log.Fatal("webhook registered but secret missing")
	}
	log.Printf("webhook registered for workspace %s", st.WorkspaceID)
	log.Printf("full webhook URL: %s", *url+"/api/v1/telegram/webhook/"+st.WebhookSecret)
	log.Print("hint: the API now pushes instant, long-poll stops")
}