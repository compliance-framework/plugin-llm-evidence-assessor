# LLM Plugin Flow and Examples

## Overview
This document shows the end-to-end flow and concrete JSON examples for how the LLM plugin integrates with evidence plugins and Rego policies.

## Architecture (Mermaid)
```mermaid
graph TD
    Agent[Agent] -->|schedule & config| EP[Evidence Plugins]
    EP -->|evidence JSON| Agent
    Agent -->|summarize & redact| LLMPL[LLM Evidence Assessor]
    LLMPL -->|provider SDK| LLM[(LLM Provider)]
    LLM -->|hints & usage| LLMPL
    LLMPL -->|AssessmentResponse| Agent
    Agent -->|OPA/Rego input\n(evidence + hints)| POL[Policy Engine]
    POL -->|pass/fail & findings| Agent
    Agent -->|report| API[(Compliance API)]
```

## AssessmentRequest (Agent → LLM Plugin)
```json
{
  "policyContext": {
    "framework": "NIST 800-53",
    "baseline": "Moderate"
  },
  "evidence": [
    {
      "controlId": "AC-1",
      "evidenceId": "sarif:todo-app",
      "contentSummary": "SARIF shows 2 High, 1 Medium findings for access controls in todo-app",
      "artifactMeta": {
        "source": "file-attestation",
        "path": "/workspace/out_sarif.json",
        "sha256": "8d1f...c0a"
      }
    },
    {
      "controlId": "AC-2",
      "evidenceId": "attest:golangci",
      "contentSummary": "Attestation exists and signer=alice@example.com for .golangci.yml",
      "artifactMeta": {
        "path": "/workspace/.golangci.yml",
        "attestation": "/workspace/.sigstore/.golangci.yml.bundle",
        "signer": "alice@example.com"
      }
    }
  ],
  "constraints": {
    "provider": "openai",
    "model": "gpt-4.1-mini",
    "maxTokens": 1000,
    "temperature": 0
  }
}
```

## AssessmentResponse (LLM Plugin → Agent)
```json
{
  "hints": [
    {
      "controlId": "AC-1",
      "hintCategory": "coverage",
      "confidence": 0.85,
      "rationale": "SARIF indicates repeated high-severity issues around access control constraints; potential gaps in implementation.",
      "citations": ["artifact:sarif:todo-app"],
      "recommendedNextSteps": [
        "Prioritize remediation of high-severity AC findings",
        "Review role/permission model in todo-app"
      ]
    },
    {
      "controlId": "AC-2",
      "hintCategory": "assurance",
      "confidence": 0.72,
      "rationale": "Attestation present; signer matches approved identity; moderate assurance.",
      "citations": ["artifact:attest:golangci"]
    }
  ],
  "usageTokenEstimate": 327,
  "provider": "openai",
  "model": "gpt-4.1-mini"
}
```

## OPA/Rego Input (Agent → Policy Engine)
```json
{
  "evidence": {
    "files": {
      "sarif": {
        "id": "sarif:todo-app",
        "path": "/workspace/out_sarif.json",
        "summary": { "high": 2, "medium": 1 }
      }
    },
    "attestation": {
      "golangci": {
        "id": "attest:golangci",
        "exists": true,
        "signer": "alice@example.com"
      }
    }
  },
  "hints": [
    {
      "controlId": "AC-1",
      "hintCategory": "coverage",
      "confidence": 0.85,
      "rationale": "SARIF indicates repeated high-severity issues around access control constraints; potential gaps in implementation.",
      "citations": ["artifact:sarif:todo-app"]
    },
    {
      "controlId": "AC-2",
      "hintCategory": "assurance",
      "confidence": 0.72,
      "rationale": "Attestation present; signer matches approved identity; moderate assurance.",
      "citations": ["artifact:attest:golangci"]
    }
  ],
  "context": {
    "framework": "NIST 800-53",
    "baseline": "Moderate"
  }
}
```

## Rego Examples (Using Hints Safely)
Raise a review flag when a high-confidence coverage hint exists for AC-1:
```rego
package compliance_framework.controls.ac_coverage

warn[msg] {
  some i
  h := input.hints[i]
  h.controlId == "AC-1"
  h.hintCategory == "coverage"
  h.confidence >= 0.8
  msg := sprintf("AC-1 requires review: %s", [h.rationale])
}
```

Combine deterministic evidence with hints:
```rego
package compliance_framework.attestation

deny[msg] {
  input.evidence.attestation.golangci.exists == false
  msg := "Attestation for .golangci.yml is missing"
}

observe[msg] {
  some i
  input.evidence.attestation.golangci.exists
  h := input.hints[i]
  h.controlId == "AC-2"
  h.hintCategory == "assurance"
  h.confidence >= 0.7
  msg := sprintf("AC-2 assurance noted: %s", [h.rationale])
}
```
