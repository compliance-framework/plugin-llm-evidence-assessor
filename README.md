# LLM Evidence Assessor Plugin (PoC)

Generates advisory “hints” (confidence, rationale, citations) from redacted evidence summaries using LLMs. Deterministic policy decisions remain in OPA/Rego.

## Objectives
- Produce non-authoritative hints from LLMs to augment compliance evaluation
- Preserve deterministic decisions in Rego; the plugin never decides pass/fail
- Maintain strict boundaries via out-of-process execution and data minimization

## Architecture
```mermaid
flowchart TD
    Agent[Agent] -->|schedule + plugin config| EP[Evidence Plugins (e.g., File Attestation)]
    EP -->|evidence JSON| Agent
    Agent -->|summarize + redact| LLMPL[LLM Evidence Assessor Plugin]
    LLMPL -->|Provider SDK| LLM[(LLM Provider)]
    LLM -->|hints + usage| LLMPL
    LLMPL -->|AssessmentResponse| Agent
    Agent -->|OPA/Rego input: evidence + hints| POL[Policy Engine]
    POL -->|pass/fail + findings| Agent
    Agent -->|report| API[(Compliance API)]
```

## Data Contracts
- EvidenceSummary: controlId, evidenceId, contentSummary, artifactMeta
- AssessmentRequest: policyContext, evidence[], constraints (provider, model, maxTokens, temperature)
- AssessmentHint: controlId, hintCategory, confidence[0..1], rationale, citations[], recommendedNextSteps[]
- AssessmentResponse: hints[], usageTokenEstimate, provider, model

## Security & Governance
- Out-of-process plugin; least-privilege FS/network
- Input must be redacted: no secrets or large artifacts
- Guardrails: token budgets, rate limits, timeouts, strict JSON validation
- Secrets from env/secret store; never logged; rotated/scoped
- Provenance: model/provider/version, token usage, timestamps

## Build
```bash
go build -o llm-assessor
```

## Local Usage
Dry-run (no provider call):
```bash
cat >/tmp/llm_req.json <<'JSON'
{
  "policyContext": {},
  "evidence": [],
  "constraints": {
    "provider": "openai",
    "model": "gpt-4.1-mini",
    "maxTokens": 1000,
    "temperature": 0
  }
}
JSON

LLM_DRY_RUN=true LLM_REQ_FILE=/tmp/llm_req.json ./llm-assessor
```

Injecting fake hints (end-to-end without provider):
```bash
cat >/tmp/llm_req_hints.json <<'JSON'
{
  "policyContext": {},
  "evidence": [
    {"controlId": "AC-1", "evidenceId": "E1", "contentSummary": "SARIF shows 2 High, 1 Medium"}
  ],
  "constraints": {
    "provider": "openai",
    "model": "gpt-4.1-mini",
    "maxTokens": 1000,
    "temperature": 0
  }
}
JSON

printf '%s' '{"hints":[{"controlId":"AC-1","hintCategory":"coverage","confidence":0.85,"rationale":"Findings indicate partial control coverage","citations":["artifact:E1"]}]}' > /tmp/fake_hints.json

LLM_DRY_RUN=true \
LLM_REQ_FILE=/tmp/llm_req_hints.json \
LLM_FAKE_JSON_FILE=/tmp/fake_hints.json \
./llm-assessor
```

## Container
```bash
docker build -t plugin-llm-evidence-assessor:local .

docker run --rm -i \
  -v /tmp/llm_req_hints.json:/req.json:ro \
  -v /tmp/fake_hints.json:/fake.json:ro \
  -e LLM_DRY_RUN=true \
  -e LLM_REQ_FILE=/req.json \
  -e LLM_FAKE_JSON_FILE=/fake.json \
  plugin-llm-evidence-assessor:local
```

## Environment Variables
- LLM_DRY_RUN=true: bypass provider calls and return fake JSON
- LLM_REQ_FILE=/path/to/req.json: file-based request input (alternative to stdin)
- LLM_FAKE_JSON/LLM_FAKE_JSON_FILE: inject fake AssessmentResponse JSON for testing
- OPENAI_API_KEY: provider key (for live calls)
- OPENAI_BASE_URL: override endpoint if using a proxy

## Agent Integration (YAML)
```yaml
daemon: false
verbosity: 0

api:
  url: http://api:8080

plugins:
  llm_evidence_assessor:
    schedule: "*/5 * * * *"
    source: plugin-llm-evidence-assessor:local
    policies:
      - ghcr.io/compliance-framework/plugin-file-attestation-policies:v0.1.0
    config:
      evidence_inputs:
        - plugin_ref: file_attestation_sarif
          control_families: ["AC","IA"]
      provider: "openai"
      model: "gpt-4.1-mini"
      max_tokens: 1000
      temperature: 0.2
      redaction_rules: "default"
      policy_labels: '{"tier":"vcs","team":"ccf","repository":"todo-app","organization":"compliance-framework"}'
```

## Rego Interaction
Hints are non-authoritative inputs; policies remain deterministic. Example patterns:
```rego
package compliance_framework.controls.ac_coverage

warn[msg] {
  some h
  input.hints[h].controlId == "AC-1"
  input.hints[h].hintCategory == "coverage"
  input.hints[h].confidence >= 0.8
  msg := "AC-1 requires review due to LLM coverage hint"
}
```

## Tests
```bash
go test ./... -v
```

## Design
See docs/DESIGN.md for detailed architecture, security, and rollout.
