package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"path/filepath"
	"context"
	"time"

	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/assessor"
	plg "github.com/compliance-framework/plugin-llm-evidence-assessor/internal/plugin"
	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/provider"
	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/api"
	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/summarize"
)

func main() {
	var req assessor.AssessmentRequest
	src := os.Getenv("LLM_REQ_FILE")
	if src != "" {
		f, err := os.Open(filepath.Clean(src))
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot open request file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&req); err != nil {
			fmt.Fprintf(os.Stderr, "invalid request file: %v\n", err)
			os.Exit(1)
		}
	} else {
		apiURL := os.Getenv("API_URL")
		apiTok := os.Getenv("API_TOKEN")
		lbl := os.Getenv("EVIDENCE_LABELS_JSON")
		tw := os.Getenv("TIME_WINDOW")
		if apiURL != "" {
			var labels map[string]string
			if lbl != "" {
				_ = json.Unmarshal([]byte(lbl), &labels)
			}
			dur := time.Duration(0)
			if tw != "" {
				if d, err := time.ParseDuration(tw); err == nil {
					dur = d
				}
			}
			cl := api.NewClient(apiURL, apiTok)
			items, err := cl.ListEvidence(context.Background(), labels, time.Now().Add(-dur))
			if err != nil {
				fmt.Fprintf(os.Stderr, "list evidence failed: %v\n", err)
				os.Exit(1)
			}
			summ := summarize.Summarize(items)
			req = assessor.AssessmentRequest{
				PolicyContext: assessor.PolicyContext{
					Framework: os.Getenv("POLICY_FRAMEWORK"),
					Baseline:  os.Getenv("POLICY_BASELINE"),
				},
				Evidence:    summ,
				Constraints: assessor.Constraints{
					Provider:    getenvDefault("PROVIDER", "openai"),
					Model:       getenvDefault("MODEL", "gpt-4.1-mini"),
					MaxTokens:   getenvInt("MAX_TOKENS", 1000),
					Temperature: getenvFloat("TEMPERATURE", 0),
				},
			}
		} else {
			if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
				fmt.Fprintf(os.Stderr, "invalid request: %v\n", err)
				os.Exit(1)
			}
		}
	}
	dry := strings.EqualFold(os.Getenv("LLM_DRY_RUN"), "true")
	c := provider.NewOpenAIClient(dry, os.Getenv("OPENAI_API_KEY"), os.Getenv("OPENAI_BASE_URL"))
	s := plg.NewServer(c)
	resp, err := s.Assess(&req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "assess error: %v\n", err)
		os.Exit(2)
	}
	if os.Getenv("PUBLISH_HINTS") == "true" && os.Getenv("API_URL") != "" {
		apiURL := os.Getenv("API_URL")
		apiTok := os.Getenv("API_TOKEN")
		labels := map[string]string{}
		_ = json.Unmarshal([]byte(os.Getenv("PUBLISH_LABELS_JSON")), &labels)
		var items []api.HintRecord
		for _, h := range resp.Hints {
			items = append(items, api.HintRecord{
				Type:         "llm_policy_result",
				Labels:       labels,
				ControlID:    h.ControlID,
				HintCategory: h.HintCategory,
				Confidence:   h.Confidence,
				Rationale:    h.Rationale,
				Citations:    h.Citations,
				Provider:     resp.Provider,
				Model:        resp.Model,
			})
		}
		if len(items) > 0 {
			cl := api.NewClient(apiURL, apiTok)
			_ = cl.PublishHints(context.Background(), api.HintsEnvelope{Items: items})
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}

func getenvDefault(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func getenvInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	var n int
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil {
		return def
	}
	return n
}

func getenvFloat(k string, def float64) float64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	var f float64
	_, err := fmt.Sscanf(v, "%f", &f)
	if err != nil {
		return def
	}
	return f
}
