# Service Blocks

Service blocks group existing explicit API declarations into a deterministic generated module boundary.

They do not create endpoints and they do not declare request handlers. The endpoint contract remains in top-level `api` blocks; `service` only adds generated service metadata for AI agents, OpenAPI clients, API-only projects, and reviewers.

## Syntax

```black
api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body sku text required
  body stock number required min 0
  update Product where sku == body.sku set stock = body.stock
  respond accepted
  webhook
  public
}

service InventoryIntegration {
  api StockWebhook
}
```

Use one target per line:

```black
service <PascalCaseName> {
  api <APIName>
}
```

Each API can belong to at most one service block so generated service metadata and `inspect --affected` stay unambiguous.

## Generated Output

When a service targets an explicit API, `black build` emits:

- `services/manifest.json`
- `src/services/<service>.ts`
- OpenAPI top-level `tags`
- OpenAPI top-level `x-blacklang-services`
- route-level `x-blacklang-service`
- generated contract assertions in `src/blacklang.contract.test.ts`

Example OpenAPI metadata:

```json
{
  "tags": ["InventoryIntegration"],
  "x-blacklang-service": "InventoryIntegration"
}
```

The generated TypeScript service module is metadata-only in the MVP. It exports the service name and API list with method, path, access, runtime, handler, webhook, and transaction metadata when present.

## AI Workflow

Use:

```bash
black docs service --json
black explain service --json
black inspect app.black --affected InventoryIntegration --json
black inspect app.black --affected StockWebhook --json
```

After editing service bindings, run:

```bash
black format --check --json
black lint --json
black validate --json
black build
```

Then inspect generated `services/manifest.json`, `src/services/<service>.ts`, `openapi.json`, and generated contract tests.

## Common Repairs

| Diagnostic | Repair |
| --- | --- |
| `INVALID_SERVICE_DECLARATION` | Start with `service Name {`. |
| `INVALID_SERVICE_NAME` | Use PascalCase, such as `InventoryIntegration`. |
| `SERVICE_NAME_COLLISION` | Choose a name that does not collide with app, entity, query, job, action, transaction, API, page, workflow, state, or component names. |
| `EMPTY_SERVICE` | Add at least one `api APIName` line. |
| `INVALID_SERVICE_API` | Use `api APIName` with a PascalCase API name. |
| `UNKNOWN_SERVICE_API` | Declare the top-level API first or remove the target. |
| `DUPLICATE_SERVICE_TARGET` | Keep one API target line per service. |
| `DUPLICATE_SERVICE_BINDING` | Bind each explicit API to at most one service. |
