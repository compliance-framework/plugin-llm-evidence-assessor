package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base string
	tok  string
	http *http.Client
}

type EvidenceItem struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Labels    map[string]string `json:"labels"`
	ControlID string            `json:"controlId"`
	Data      map[string]any    `json:"data"`
	CreatedAt string            `json:"createdAt"`
}

type listResponse struct {
	Items []EvidenceItem `json:"items"`
}

func NewClient(base string, token string) *Client {
	return &Client{
		base: strings.TrimRight(base, "/"),
		tok:  token,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) ListEvidence(ctx context.Context, labels map[string]string, since time.Time) ([]EvidenceItem, error) {
	u, _ := url.Parse(c.base + "/evidence")
	q := u.Query()
	for k, v := range labels {
		q.Set("label."+k, v)
	}
	if !since.IsZero() {
		q.Set("since", since.UTC().Format(time.RFC3339))
	}
	u.RawQuery = q.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if c.tok != "" {
		req.Header.Set("Authorization", "Bearer "+c.tok)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var lr listResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, err
	}
	return lr.Items, nil
}
