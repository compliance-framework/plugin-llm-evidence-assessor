package summarize

import (
	"fmt"
	"path/filepath"

	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/api"
	"github.com/compliance-framework/plugin-llm-evidence-assessor/internal/assessor"
)

func basename(p any) string {
	if s, ok := p.(string); ok {
		return filepath.Base(s)
	}
	return ""
}

func asInt(m map[string]any, k string) int {
	if v, ok := m[k]; ok {
		switch t := v.(type) {
		case float64:
			return int(t)
		case int:
			return t
		}
	}
	return 0
}

func asStringSlice(m map[string]any, k string, limit int) []string {
	var out []string
	if v, ok := m[k]; ok {
		if s, ok := v.([]any); ok {
			for _, e := range s {
				if str, ok := e.(string); ok {
					out = append(out, str)
					if len(out) >= limit {
						break
					}
				}
			}
		}
	}
	return out
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func asBool(m map[string]any, k string) bool {
	if v, ok := m[k]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func Summarize(items []api.EvidenceItem) []assessor.EvidenceSummary {
	var out []assessor.EvidenceSummary
	for _, it := range items {
		switch it.Type {
		case "sarif_summary":
			high := asInt(it.Data, "high")
			med := asInt(it.Data, "medium")
			tool := it.Data["tool"]
			cats := asStringSlice(it.Data, "categories", 5)
			summary := fmt.Sprintf("High:%d, Medium:%d, tool:%v, categories:%v", high, med, tool, cats)
			out = append(out, assessor.EvidenceSummary{
				ControlID:      it.ControlID,
				EvidenceID:     it.ID,
				ContentSummary: summary,
				ArtifactMeta: map[string]string{
					"basename": basename(it.Data["path"]),
				},
			})
		case "attestation":
			exists := asBool(it.Data, "exists")
			signerApproved := asBool(it.Data, "signerApproved")
			issuerApproved := asBool(it.Data, "issuerApproved")
			summary := fmt.Sprintf("attestation_exists:%s, signer_approved:%s, issuer_approved:%s", boolStr(exists), boolStr(signerApproved), boolStr(issuerApproved))
			out = append(out, assessor.EvidenceSummary{
				ControlID:      it.ControlID,
				EvidenceID:     it.ID,
				ContentSummary: summary,
				ArtifactMeta: map[string]string{
					"basename": basename(it.Data["path"]),
				},
			})
		default:
		}
	}
	return out
}
