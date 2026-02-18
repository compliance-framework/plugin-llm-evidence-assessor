package assessor

type EvidenceSummary struct {
	ControlID      string            `json:"controlId"`
	EvidenceID     string            `json:"evidenceId"`
	ContentSummary string            `json:"contentSummary"`
	ArtifactMeta   map[string]string `json:"artifactMeta,omitempty"`
}

type PolicyContext struct {
	Framework string `json:"framework,omitempty"`
	Baseline  string `json:"baseline,omitempty"`
}

type Constraints struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
}

type AssessmentRequest struct {
	PolicyContext PolicyContext     `json:"policyContext"`
	Evidence      []EvidenceSummary `json:"evidence"`
	Constraints   Constraints       `json:"constraints"`
}

type AssessmentHint struct {
	ControlID            string   `json:"controlId"`
	HintCategory         string   `json:"hintCategory"`
	Confidence           float64  `json:"confidence"`
	Rationale            string   `json:"rationale"`
	Citations            []string `json:"citations,omitempty"`
	RecommendedNextSteps []string `json:"recommendedNextSteps,omitempty"`
}

type AssessmentResponse struct {
	Hints              []AssessmentHint `json:"hints"`
	UsageTokenEstimate int              `json:"usageTokenEstimate"`
	Provider           string           `json:"provider"`
	Model              string           `json:"model"`
}

type EvidenceAssessor interface {
	Assess(req *AssessmentRequest) (*AssessmentResponse, error)
}
