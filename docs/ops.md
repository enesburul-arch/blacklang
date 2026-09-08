# BlackLang Ops

Ops declarations describe public runtime signals for generated web deployments.

```black
ops {
  health path "/healthz"
  readiness path "/readyz"
  metrics path "/metrics"
  logging requests
  observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT
}
```

Rules:

- Use one top-level `ops` block.
- `health path` generates a public GET endpoint with status, app, version, uptime, and start time.
- `readiness path` generates a public GET endpoint that checks database connectivity and returns `503` when the database is unavailable.
- `metrics path` generates a public GET endpoint with in-process request, error, status-code, uptime, and start-time counters.
- `logging requests` emits one structured JSON console log after each observed request finishes.
- `observe webhook endpoint env NAME` sends non-blocking structured request events to the configured endpoint when that environment variable is set.
- `observe otlp endpoint env NAME` sends non-blocking OTLP HTTP JSON trace payloads to the configured endpoint when that environment variable is set.
- Observe middleware creates or propagates W3C `traceparent` context and adds the response `traceparent` header.
- Paths must be root-level public paths. Do not use `/api`, `/openapi.json`, query strings, fragments, braces, whitespace, or unsafe characters.
- Observability endpoint values must come from environment variables. Do not hardcode vendor URLs, tokens, DSNs, or API keys in `.black` source.

Generated output:

- `src/server.ts` includes the declared routes before API auth and CSRF middleware.
- `openapi.json` includes `x-blacklang-ops` and `x-blacklang-public` metadata for each declared ops endpoint.
- `openapi.json` includes `x-blacklang-observability` metadata when an observe hook is declared.
- `ops/observability.json` records provider, endpoint env, signal, protocol, trace context, delivery, and exporter metadata.
- `src/blacklang.contract.test.ts` checks the OpenAPI ops paths and metadata.
- `src/blacklang.api.test.ts` probes declared health, readiness, metrics, trace context, and local observability delivery over HTTP.
- `.env.example`, `docker-compose.yml`, and local preview compose output include the observe endpoint env reference.
- When `deploy { target docker }` and `health path` both exist, `Dockerfile` and `docker-compose.yml` include a Node-based healthcheck.

Before editing ops intent, inspect the affected graph:

```bash
black inspect app.black --affected ops --json
```

After changing ops declarations:

```bash
black format --check --json
black lint --json
black validate --json
black build
cd generated
npm test
```

Diagnostics:

```text
INVALID_OPS_DECLARATION
DUPLICATE_OPS
UNCLOSED_OPS
UNEXPECTED_OPS_TOKEN
INVALID_OPS_HEALTH
DUPLICATE_OPS_HEALTH
INVALID_OPS_READINESS
DUPLICATE_OPS_READINESS
INVALID_OPS_METRICS
DUPLICATE_OPS_METRICS
INVALID_OPS_LOGGING
DUPLICATE_OPS_LOGGING
INVALID_OPS_OBSERVE
DUPLICATE_OPS_OBSERVE
MISSING_OPS_SIGNAL
UNSUPPORTED_OPS_LOGGING
UNSUPPORTED_OPS_OBSERVE_PROVIDER
MISSING_OPS_OBSERVE_ENDPOINT_ENV
INVALID_OPS_PATH
DUPLICATE_OPS_PATH
INVALID_ENV_NAME
```
