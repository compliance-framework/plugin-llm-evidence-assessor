package provider

import (
	"context"
	"testing"
)

func TestOpenAIClientDryRun(t *testing.T) {
	c := NewOpenAIClient(true, "", "")
	out, err := c.Generate(context.Background(), Params{
		Provider:    "openai",
		Model:       "gpt-4.1-mini",
		MaxTokens:   100,
		Temperature: 0,
	}, Payload{Prompt: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.JSON == "" {
		t.Fatalf("expected JSON in output")
	}
}
