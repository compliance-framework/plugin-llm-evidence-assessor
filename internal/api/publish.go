package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type HintRecord struct {
	Type          string            `json:"type"`
	Labels        map[string]string `json:"labels"`
	ControlID     string            `json:"controlId"`
	HintCategory  string            `json:"hintCategory"`
	Confidence    float64           `json:"confidence"`
	Rationale     string            `json:"rationale"`
	Citations     []string          `json:"citations"`
	Provider      string            `json:"provider"`
	Model         string            `json:"model"`
}

type HintsEnvelope struct {
	Items []HintRecord `json:"items"`
}

func (c *Client) PublishHints(ctx context.Context, env HintsEnvelope) error {
	u := c.base + "/evidence"
	b, _ := json.Marshal(env)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if c.tok != "" {
		req.Header.Set("Authorization", "Bearer "+c.tok)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &httpError{code: resp.StatusCode, status: resp.Status}
	}
	return nil
}

type httpError struct {
	code   int
	status string
}

func (e *httpError) Error() string {
	return strings.TrimSpace(e.status)
}
