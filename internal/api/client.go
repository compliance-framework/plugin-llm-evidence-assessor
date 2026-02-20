package api

import (
	"context"
	"encoding/json"
	"net/http"
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

// labelfilter request payloads (minimal representation)
type lfCondition struct {
	Label    string `json:"label,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
}
type lfScope struct {
	Condition *lfCondition `json:"condition,omitempty"`
	Query     *lfQuery     `json:"query,omitempty"`
}
type lfQuery struct {
	Operator string    `json:"operator"`
	Scopes   []lfScope `json:"scopes"`
}
type lfFilter struct {
	Scope *lfScope `json:"scope"`
}
type searchRequest struct {
	Filter lfFilter `json:"filter"`
}

func (c *Client) ListEvidence(ctx context.Context, labels map[string]string, since time.Time) ([]EvidenceItem, error) {
	var scopes []lfScope
	for k, v := range labels {
		scopes = append(scopes, lfScope{Condition: &lfCondition{Label: k, Operator: "=", Value: v}})
	}
	body := searchRequest{
		Filter: lfFilter{
			Scope: &lfScope{
				Query: &lfQuery{
					Operator: "AND",
					Scopes:   scopes,
				},
			},
		},
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/evidence/search", strings.NewReader(string(b)))
	if c.tok != "" {
		req.Header.Set("Authorization", "Bearer "+c.tok)
	}
	req.Header.Set("Content-Type", "application/json")
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
