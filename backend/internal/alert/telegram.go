package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Telegram struct {
	Token  string
	Client *http.Client
}

func NewTelegram(token string) *Telegram {
	return &Telegram{Token: token, Client: &http.Client{Timeout: 10 * time.Second}}
}

func (t *Telegram) Send(chatID, text string) error {
	if t.Token == "" {
		return fmt.Errorf("telegram not configured")
	}
	payload := map[string]string{"chat_id": chatID, "text": text}
	b, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Token)
	res, err := t.Client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("telegram status %d", res.StatusCode)
	}
	return nil
}
