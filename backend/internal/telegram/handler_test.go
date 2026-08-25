package telegram

import "testing"

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