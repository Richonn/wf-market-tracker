package api

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

const (
	baseV1      = "https://api.warframe.market/v1"
	baseV2      = "https://api.warframe.market/v2"
	userAgent   = "wf-market-tracker/0.1 (github.com/Richonn/wf-market-tracker)"
	reqInterval = 400 * time.Millisecond
)

type Client struct {
	http    *http.Client
	limiter <-chan time.Time
}

func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 15 * time.Second},
		limiter: time.Tick(reqInterval),
	}
}

func (c *Client) get(url string) ([]byte, error) {
	<-c.limiter

	var body []byte
	var err error

	for attempt := range 3 {
		body, err = c.do(url)
		if err == nil {
			return body, nil
		}
		if attempt < 2 {
			time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
		}
	}
	return nil, err
}

func (c *Client) do(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Platform", "pc")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}
