# CIMPLR Policy Check Service — Deployment

Standalone microservice (like Email Service). **No database** — CIMPLR Go owns all `policyengine_svc.*` Postgres read/write.

## Responsibilities

| This service (`:8184`) | CIMPLR Go (`:8185` + gateway) |
|------------------------|-------------------------------|
| Evaluate policies (threshold / slabs / …) | `policyengine_svc` schema, UI API |
| PEL validate + test harness | CDM / trigger / policy CRUD + audit |
| Pure check — no Postgres | Calls this service with `service_key` |

## Build & run locally

```bash
cd CIMPLR-Policy-Service
cp .env.example .env
go run ./cmd/
```

Health:

```bash
curl -s -X POST http://localhost:8184/v1/health \
  -H 'Content-Type: application/json' \
  -d '{"service_key":"YOUR_KEY"}'
```

## Env

| Variable | Purpose |
|----------|---------|
| `PORT` | Default `8184` |
| `POLICY_SERVICE_KEY` | Shared secret with `cimplrcorpsaas` |

CIMPLR `.env` must set:

- `POLICY_SERVICE_URL=http://localhost:8184`
- `POLICY_SERVICE_KEY=<same secret>`

## API (all POST, JSON body)

- `/v1/health`
- `/v1/evaluate` — evaluate a policy set against CDM values
- `/v1/test` — workbench test harness (no audit side-effects)
- `/v1/pel/validate` — validate PEL expression

Auth: `Authorization: Bearer <POLICY_SERVICE_KEY>` or `"service_key"` in JSON body.
