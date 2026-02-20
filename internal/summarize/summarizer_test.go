package summarize

import (
	"testing"

	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/api"
)

func TestSummarizeSarifAndAttestation(t *testing.T) {
	items := []api.EvidenceItem{
		{
			ID:        "sarif:todo-app",
			Type:      "sarif_summary",
			ControlID: "AC-1",
			Data: map[string]any{
				"high":       2.0,
				"medium":     1.0,
				"tool":       "golangci-lint",
				"categories": []any{"access-control"},
				"path":       "/workspace/out_sarif.json",
			},
		},
		{
			ID:        "attest:golangci",
			Type:      "attestation",
			ControlID: "AC-2",
			Data: map[string]any{
				"exists":          true,
				"signerApproved":  true,
				"issuerApproved":  true,
				"path":            "/workspace/.golangci.yml",
			},
		},
	}
	out := Summarize(items)
	if len(out) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(out))
	}
	if out[0].ControlID != "AC-1" || out[1].ControlID != "AC-2" {
		t.Fatalf("unexpected control IDs")
	}
	if out[0].ArtifactMeta["basename"] != "out_sarif.json" {
		t.Fatalf("unexpected basename for sarif")
	}
	if out[1].ArtifactMeta["basename"] != ".golangci.yml" {
		t.Fatalf("unexpected basename for attestation")
	}
}
