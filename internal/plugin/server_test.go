package plugin

import (
	"context"
	"testing"

	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/assessor"
	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/provider"
)

type fakeClient struct {
	json string
}

func (f *fakeClient) Generate(ctx context.Context, p provider.Params, in provider.Payload) (provider.Output, error) {
	return provider.Output{JSON: f.json, UsageTokens: 42}, nil
}

func TestServerAssessParsesHints(t *testing.T) {
	f := &fakeClient{json: `{"hints":[{"controlId":"AC-1","hintCategory":"coverage","confidence":0.9,"rationale":"ok"}]}`}
	s := NewServer(f)
	resp, err := s.Assess(&assessor.AssessmentRequest{
		Constraints: assessor.Constraints{Provider: "openai", Model: "gpt-4.1-mini"},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(resp.Hints) != 1 || resp.Hints[0].ControlID != "AC-1" {
		t.Fatalf("unexpected hints: %+v", resp.Hints)
	}
	if resp.UsageTokenEstimate != 42 {
		t.Fatalf("expected usage tokens 42, got %d", resp.UsageTokenEstimate)
	}
}
