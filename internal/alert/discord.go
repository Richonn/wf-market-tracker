package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Discord struct {
	WebhookURL string
}

type discordPayload struct {
	Content string `json:"content"`
}

func (d *Discord) Send(title, message string) error {
	payload := discordPayload{Content: fmt.Sprintf("**%s**\n%s", title, message)}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("discord returned %d", resp.StatusCode)
	}
	return nil
}
