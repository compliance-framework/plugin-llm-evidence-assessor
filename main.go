package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/assessor"
	plg "github.com/compliance-framework/plugin-llm-evidence-assessor/internal/plugin"
	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/provider"
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
		if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
			fmt.Fprintf(os.Stderr, "invalid request: %v\n", err)
			os.Exit(1)
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
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}
