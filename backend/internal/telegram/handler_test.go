package telegram

import (
	"encoding/json"
	"testing"
)

func TestWebhookRejectsMissingTelegramSecretHeader(t *testing.T) {
	if validWebhookSecret("secret", "secret", "") {
		t.Fatal("missing Telegram secret header accepted")
	}
}

func TestWebhookRequiresMatchingPathAndHeaderSecrets(t *testing.T) {
	if !validWebhookSecret("secret", "secret", "secret") {
		t.Fatal("matching webhook secrets rejected")
	}
	if validWebhookSecret("secret", "other", "secret") {
		t.Fatal("wrong webhook path secret accepted")
	}
}

func TestWebhookURLForIncludesAPIPrefix(t *testing.T) {
	got := webhookURLFor("https://abc.ngrok-free.app", "s3cr3t")
	want := "https://abc.ngrok-free.app/api/v1/telegram/webhook/s3cr3t"
	if got != want {
		t.Fatalf("webhookURLFor = %q, want %q", got, want)
	}
	if webhookURLFor("", "s") != "" {
		t.Fatal("empty base should produce empty url")
	}
}

func TestWebhookURLForSupportsReverseProxyBasePath(t *testing.T) {
	got := webhookURLFor("https://api.example.com/savio/", "s3cr3t")
	want := "https://api.example.com/savio/api/v1/telegram/webhook/s3cr3t"
	if got != want {
		t.Fatalf("webhookURLFor = %q, want %q", got, want)
	}
}

func TestUpdateReadsSenderFirstName(t *testing.T) {
	var update Update
	if err := json.Unmarshal([]byte(`{"update_id":1,"message":{"chat":{"id":2},"from":{"first_name":"Wina"},"text":"kopi 25000"}}`), &update); err != nil {
		t.Fatal(err)
	}
	if update.Message.From.FirstName != "Wina" {
		t.Fatalf("first name = %q", update.Message.From.FirstName)
	}
}
