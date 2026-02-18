package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"os"
)

type openAIClient struct {
	dry     bool
	apiKey  string
	baseURL string
	httpc   *http.Client
}

func NewOpenAIClient(dry bool, apiKey string, baseURL string) Client {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/chat/completions"
	}
	return &openAIClient{
		dry:     dry,
		apiKey:  apiKey,
		baseURL: baseURL,
		httpc:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *openAIClient) Generate(ctx context.Context, params Params, in Payload) (Output, error) {
	if c.dry || c.apiKey == "" {
		fake := os.Getenv("LLM_FAKE_JSON")
		if fake == "" {
			if fp := os.Getenv("LLM_FAKE_JSON_FILE"); fp != "" {
				b, err := os.ReadFile(fp)
				if err == nil {
					fake = string(b)
				}
			}
		}
		if fake == "" {
			fake = `{"hints":[]}`
		}
		return Output{JSON: fake, UsageTokens: 0}, nil
	}
	body := map[string]any{
		"model": params.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a JSON-only responder. Output strictly valid JSON matching AssessmentResponse shape with {\"hints\":[]} or populated hints."},
			{"role": "user", "content": in.Prompt},
		},
		"temperature": params.Temperature,
		"max_tokens":  params.MaxTokens,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return Output{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, buf)
	if err != nil {
		return Output{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpc.Do(req)
	if err != nil {
		return Output{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Output{}, errors.New(resp.Status)
	}
	var r struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Output{}, err
	}
	out := Output{JSON: "", UsageTokens: r.Usage.TotalTokens}
	if len(r.Choices) > 0 {
		out.JSON = r.Choices[0].Message.Content
	}
	if out.JSON == "" {
		out.JSON = `{"hints":[]}`
	}
	return out, nil
}
