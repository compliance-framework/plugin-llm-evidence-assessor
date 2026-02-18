# LLM Evidence Assessor Plugin – Design

## Objective
- Generate advisory hints (confidence, rationale, citations) from redacted evidence summaries using LLMs.
- Keep deterministic compliance decisions in OPA/Rego; plugin never decides pass/fail.
- Maintain strong security boundaries through out-of-process execution and data minimization.

## Architecture
- Agent schedules evidence plugins (e.g., file-attestation) and collects evidence.
- Agent summarizes and redacts evidence, then invokes the LLM plugin.
- LLM plugin calls the provider, parses JSON hints, and returns an AssessmentResponse.
- Agent applies OPA/Rego policies with evidence + hints and reports findings.

## Data Contracts
- EvidenceSummary: controlId, evidenceId, contentSummary, artifactMeta.
- AssessmentRequest: policyContext, evidence[], constraints (provider, model, budgets).
- AssessmentHint: controlId, hintCategory, confidence, rationale, citations, recommendedNextSteps.
- AssessmentResponse: hints[], usageTokenEstimate, provider, model.

## Security
- Out-of-process/plugin container with least privilege.
- Redaction in agent; no raw secrets or large artifacts to the plugin.
- Guardrails: token budgets, rate limits, timeouts, output schema validation.
- Secrets from env/secret store; never logged; rotated.
- Provenance: model, provider, usage, timestamps, policy rule IDs.

## Scalability & Resilience
- Worker pools with backpressure; batching/streaming for large inputs.
- Retry/backoff with circuit breakers; fallback providers/models.
- Optional memoization keyed by (artifact hash, controlId, model, context).
- Strict timeouts; accept partial results when necessary.

## Observability
- Metrics: latency, errors, token usage, hint acceptance rate.
- Tracing with correlation IDs across agent ↔ plugin ↔ provider.
- Structured, redacted logs; decision provenance retained.

## Integration
- Packaged as OCI image referenced by `plugins.<id>.source`.
- Policies remain the authority; hints are advisory only.
- Compatible with existing file-attestation plugins and policy bundles.

## Rollout
- Phase 1: Go-only plugin, single provider, AC/IA families, hints-only.
- Phase 2: Expand families, streaming/batching/caching, SLIs dashboards.
- Phase 3: Optional Python gRPC adapter if needed, benchmarking/cost optimization.
- Phase 4: Governance hardening and model evaluation harness.
