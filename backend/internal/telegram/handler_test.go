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
