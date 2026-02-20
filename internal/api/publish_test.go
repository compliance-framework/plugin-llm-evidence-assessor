package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublishHints(t *testing.T) {
	var got HintsEnvelope
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/evidence" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		w.WriteHeader(201)
	}))
	defer ts.Close()
	c := NewClient(ts.URL, "")
	err := c.PublishHints(context.Background(), HintsEnvelope{
		Items: []HintRecord{{
			Type:         "llm_policy_result",
			Labels:       map[string]string{"team": "ccf"},
			ControlID:    "AC-1",
			HintCategory: "coverage",
			Confidence:   0.9,
			Rationale:    "ok",
			Citations:    []string{"evidence:sarif:todo-app"},
			Provider:     "openai",
			Model:        "gpt-4.1-mini",
		}},
	})
	if err != nil {
		t.Fatalf("publish error: %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].ControlID != "AC-1" {
		t.Fatalf("unexpected payload")
	}
}
