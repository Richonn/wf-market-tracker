package alert

import (
	"fmt"
	"net/http"
	"strings"
)

type Ntfy struct {
	URL   string
	Token string
}

func (n *Ntfy) Send(title, message string) error {
	req, err := http.NewRequest(http.MethodPost, n.URL, strings.NewReader(message))
	if err != nil {
		return err
	}
	req.Header.Set("Title", title)
	if n.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", n.Token))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy returned %d", resp.StatusCode)
	}
	return nil
}
