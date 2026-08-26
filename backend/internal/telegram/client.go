package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// pollTimeout is the Telegram long-poll period; slightly under the HTTP client
// timeout so the client gives up first and the poll loop stays responsive.
const pollTimeout = 25 * time.Second

// Update is the subset of a Telegram update needed for recap processing.
type Update struct {
	ID      int64 `json:"update_id"`
	Message *struct {
		MessageID int64 `json:"message_id"`
		Chat      struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		From struct {
			FirstName string `json:"first_name"`
		} `json:"from"`
		Text string `json:"text"`
	} `json:"message"`
}

type botClient struct {
	token  string
	rawURL string
	client *http.Client
}

func newBotClient(token string) *botClient {
	return &botClient{
		token:  token,
		rawURL: "https://api.telegram.org",
		client: &http.Client{Timeout: pollTimeout + 10*time.Second},
	}
}

// updates long-polls Telegram from the given offset (last processed update ID).
func (b *botClient) updates(ctx context.Context, offset int64) ([]Update, error) {
	u := fmt.Sprintf("%s/bot%s/getUpdates?timeout=%d&offset=%d", b.rawURL, b.token, int(pollTimeout.Seconds()), offset)
	var payload struct {
		Ok          bool     `json:"ok"`
		Result      []Update `json:"result"`
		Description string   `json:"description"`
	}
	if err := b.do(ctx, "GET", u, nil, &payload); err != nil {
		return nil, err
	}
	if !payload.Ok {
		return nil, fmt.Errorf("telegram getUpdates: %s", payload.Description)
	}
	return payload.Result, nil
}

func (b *botClient) sendMessage(ctx context.Context, chatID, text string) error {
	body := map[string]any{"chat_id": chatID, "text": text}
	u := fmt.Sprintf("%s/bot%s/sendMessage", b.rawURL, b.token)
	var payload struct {
		Ok bool `json:"ok"`
	}
	return b.do(ctx, "POST", u, body, &payload)
}

func (b *botClient) setWebhook(ctx context.Context, webhookURL, secret string) error {
	u := fmt.Sprintf("%s/bot%s/setWebhook?url=%s&secret_token=%s", b.rawURL, b.token, url.QueryEscape(webhookURL), url.QueryEscape(secret))
	var payload struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := b.do(ctx, "GET", u, nil, &payload); err != nil {
		return err
	}
	if !payload.Ok {
		return fmt.Errorf("telegram setWebhook: %s", payload.Description)
	}
	return nil
}

func (b *botClient) deleteWebhook(ctx context.Context) error {
	u := fmt.Sprintf("%s/bot%s/deleteWebhook", b.rawURL, b.token)
	var payload struct {
		Ok bool `json:"ok"`
	}
	return b.do(ctx, "GET", u, nil, &payload)
}

func (b *botClient) do(ctx context.Context, method, u string, body any, out any) error {
	var rd io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("telegram http %d for %s", resp.StatusCode, method)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}
	return nil
}
