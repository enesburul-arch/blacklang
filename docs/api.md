# Explicit API Declarations

Explicit `api` blocks declare REST endpoints that sit outside generated page CRUD routes.

They generate:

- parse JSON and BlackIR entries
- `black inspect` and `black inspect --affected` context
- OpenAPI paths with access, webhook, typed body, handler, transaction, service, and `x-blacklang-runtime: declared` metadata
- generated Express routes in `src/server.ts`
- generated contract and API smoke test coverage

These routes work in both `target web` and `target api` projects. Use `target api { backend node database sqlite }`, `database postgres`, or `database mysql` when the project should emit only the generated Express/Prisma/OpenAPI runtime without React and Vite files.

## Syntax

```black
api LowStockReport {
  method GET
  path "/api/reports/low-stock/{warehouseId}"
  param warehouseId text
  query limit integer
  private
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body tenantId text required
  body sku text required
  body quantity number required min 0
  body packSize number required min 1
  update Product where sku == body.sku {
    value incoming = body.quantity / body.packSize
    if incoming > 0 and tenantId == body.tenantId
      set stock = stock + incoming
    else
      set stock = stock
  }
  respond accepted
  webhook
  public
}
```

Supported methods are `GET`, `POST`, `PUT`, `PATCH`, and `DELETE`.

Paths must be safe `/api/...` routes. They cannot use `/api/auth`, trailing slashes, query strings, fragments, malformed `{param}` syntax, or generated page/auth/query/action/workflow routes.

Path params use `{name}` in the path and must have a matching `param name type` line. Query params use `query name type`. Parameter types use primitive BlackLang field types.

Body fields use `body name type` on `POST`, `PUT`, and `PATCH` APIs. Body types may include `file` and `image` values as data URL or absolute `http(s)` strings. Body fields may use `required`, `optional`, `min`, `max`, `length`, `regex`, `url`, and `message` modifiers; `accept` belongs to entity media form fields.

The shortest handler is one bounded entity update:

```black
update Product where sku == body.sku set stock = body.stock
```

The `where` field must be `id` or a stored field marked `unique`. Set targets must be stored primitive non-policy fields. Handler operands may be `body.name`, `param.name`, source fields inside set expressions, typed string/number/boolean literals, or bounded numeric expressions with `+`, `-`, `*`, `/`, parentheses, and deterministic precedence.

```black
update Product where sku == body.sku set stock = stock + body.quantity / body.packSize
```

When an update needs local values or conditional logic, use the block form:

```black
update Product where sku == body.sku {
  value incoming = body.quantity / body.packSize
  if incoming > 0 and tenantId == body.tenantId
    set stock = stock + incoming
  else
    set stock = stock
}
```

`value` declares an ordered local value for later `value`, `set`, and `if` statements. Branch bodies are indentation-based: statements under `if` or `else` must be indented deeper than the `if` line, and `else` must align with its matching `if`. Supported branch statements are `value`, `set`, and nested `if`.

Condition comparisons use `==`, `!=`, `<`, `<=`, `>`, and `>=`. Equality compares compatible scalar type families. Ordered comparisons currently require numeric values. Comparisons may be joined with `and`, `or`, `not`, and parentheses, using `not` > `and` > `or` precedence. Local value names must be safe identifiers and must not collide with body fields, params, source fields, computed fields, or reserved words.

Use a top-level [`transaction`](transaction.md) block when the generated update handler must run bounded lookup, update, and private auth audit logging inside one Prisma transaction. Use a top-level [`service`](service.md) block when explicit APIs should be grouped into a generated service manifest/module and OpenAPI service boundary without redefining endpoints.

## Runtime Behavior

Generated explicit API routes validate declared path, query, and body values, then return deterministic JSON:

```json
{
  "api": "LowStockReport",
  "status": "declared",
  "runtime": "declared",
  "method": "GET",
  "path": "/api/reports/low-stock/{warehouseId}",
  "access": "private",
  "webhook": false,
  "params": { "warehouseId": "north" },
  "query": { "limit": 10 },
  "body": null
}
```

Webhook routes return `202` with `status: "accepted"`. `POST`, `PUT`, and `PATCH` routes echo the parsed JSON body. `GET` and `DELETE` routes report `body: null`.

Handler routes update one bounded row and return deterministic metadata instead of echoing arbitrary runtime state:

```json
{
  "api": "StockWebhook",
  "status": "accepted",
  "runtime": "declared",
  "handler": "update",
  "entity": "Product",
  "id": "product-id",
  "updated": true
}
```

`public` routes mount before auth middleware. `private` routes mount behind generated auth and CSRF middleware when an `auth` block exists. In projects without `auth`, the route is still generated and OpenAPI metadata remains `private`, but there is no auth runtime to enforce.

When a transaction block targets an API update handler, request parsing and permission checks run before the database transaction; bounded row lookup, deterministic update, and private auth audit logging run inside `prisma.$transaction`. OpenAPI includes `x-blacklang-transaction: true` and `x-blacklang-transaction-name` for targeted update routes.

When a service block targets an API, generated output includes `services/manifest.json`, `src/services/<service>.ts`, OpenAPI root `x-blacklang-services`, route-level `x-blacklang-service`, and OpenAPI `tags`. Service blocks do not change request handling; they make API module intent explicit for AI agents, clients, API-only projects, and generated contract tests.

Tenant and owner policy rules are explicit:

- Private handlers use generated auth context for tenant/owner row scope.
- Public handlers on tenant-scoped entities must include the tenant field as a required body or path value.
- Public handlers cannot update owner-scoped entities.

## Current Limit

Explicit API handlers are intentionally bounded. They do not run raw SQL, arbitrary JavaScript, cross-route workflows, queues, third-party provider calls, or job triggers. Use page-bound `action` blocks for row-level UI actions, top-level `job` blocks for read-only query workers, and explicit `api` update handlers for safe webhook/internal API mutations.

Use:

```bash
black docs api --json
black explain api --json
black docs transaction --json
black docs service --json
black inspect app.black --affected LowStockReport --json
black inspect app.black --affected RestockAtomic --json
black inspect app.black --affected InventoryIntegration --json
```

