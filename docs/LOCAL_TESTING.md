# Local Testing and Seeding Guide

## Bring up API locally
```bash
cd /Users/abdulhanif/demo/local-dev
docker compose -f compose/common.yml -f compose/smtp4dev.yml -f compose/api.yml up -d --force-recreate
curl -s http://localhost:8080/api/health
```

## Seed minimal Evidence
Create a small Evidence record using the API. Adjust labels to your environment.

```bash
cat > /tmp/evidence_minimal.json <<'JSON'
{
  "uuid": "00000000-0000-0000-0000-000000000001",
  "title": "Demo SARIF summary",
  "description": "Minimal seed for local tests",
  "labels": {
    "team": "ccf",
    "repository": "todo-app",
    "framework": "NIST-800-53"
  },
  "start": "2026-02-18T12:00:00Z",
  "end": "2026-02-18T12:05:00Z",
  "origins": [],
  "activities": [],
  "inventoryItems": [],
  "components": [],
  "subjects": [],
  "status": {
    "state": "not-satisfied"
  }
}
JSON

curl -s -X POST http://localhost:8080/api/evidence \
  -H "Content-Type: application/json" \
  --data-binary @/tmp/evidence_minimal.json
```

## Search Evidence by labels
The evidence search uses a label filter body. This returns the matching Evidence set.

```bash
cat > /tmp/evidence_search.json <<'JSON'
{
  "filter": {
    "scope": {
      "query": {
        "operator": "AND",
        "scopes": [
          { "condition": { "label": "team", "operator": "=", "value": "ccf" } },
          { "condition": { "label": "repository", "operator": "=", "value": "todo-app" } }
        ]
      }
    }
  }
}
JSON

curl -s -X POST http://localhost:8080/api/evidence/search \
  -H "Content-Type: application/json" \
  --data-binary @/tmp/evidence_search.json | jq .
```

## Run the LLM plugin locally (dry-run)
Fetch → summarize → assess with no provider calls:

```bash
cd /Users/abdulhanif/demo/plugin-llm-evidence-assessor
go build -o llm-assessor

LLM_DRY_RUN=true \
API_URL=http://localhost:8080/api \
EVIDENCE_LABELS_JSON='{"team":"ccf","repository":"todo-app"}' \
TIME_WINDOW=24h \
POLICY_FRAMEWORK="NIST 800-53" POLICY_BASELINE="Moderate" \
PROVIDER=openai MODEL=gpt-4.1-mini MAX_TOKENS=800 TEMPERATURE=0.2 \
./llm-assessor
```

Publish advisory results (dry-run with fake hints):

```bash
printf '%s' \
'{"hints":[{"controlId":"AC-1","hintCategory":"coverage","confidence":0.85,"rationale":"Coverage concern","citations":["evidence:sarif:todo-app"]}]}' \
> /tmp/fake_hints.json

LLM_DRY_RUN=true \
API_URL=http://localhost:8080/api \
PUBLISH_HINTS=true \
PUBLISH_LABELS_JSON='{"team":"ccf","repository":"todo-app","framework":"NIST-800-53"}' \
LLM_FAKE_JSON_FILE=/tmp/fake_hints.json \
./llm-assessor
```

## Notes
- For deployments requiring authentication, set API_TOKEN and ensure the plugin and curl add the Authorization header.
- Keep inputs minimal and allowlisted; avoid seeding sensitive content. 
