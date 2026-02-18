package plugin

import (
	"context"
	"encoding/json"

	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/assessor"
	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/provider"
)

type Server struct {
	client provider.Client
}

func NewServer(c provider.Client) *Server {
	return &Server{client: c}
}

func buildPrompt(req *assessor.AssessmentRequest) string {
	return "Summarize evidence and produce JSON with hints"
}

func (s *Server) Assess(req *assessor.AssessmentRequest) (*assessor.AssessmentResponse, error) {
	params := provider.Params{
		Provider:    req.Constraints.Provider,
		Model:       req.Constraints.Model,
		MaxTokens:   req.Constraints.MaxTokens,
		Temperature: req.Constraints.Temperature,
	}
	out, err := s.client.Generate(context.Background(), params, provider.Payload{Prompt: buildPrompt(req)})
	if err != nil {
		return nil, err
	}
	var resp assessor.AssessmentResponse
	if err := json.Unmarshal([]byte(out.JSON), &resp); err != nil {
		return nil, err
	}
	resp.Provider = params.Provider
	resp.Model = params.Model
	resp.UsageTokenEstimate = out.UsageTokens
	return &resp, nil
}
