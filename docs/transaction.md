# Transaction Blocks

Transaction blocks declare atomic runtime intent for generated write handlers. They let BlackLang keep action/API mutation declarations small while making database transaction boundaries explicit for AI agents, reviewers, OpenAPI clients, and generated tests.

## Syntax

```black
transaction RestockAtomic {
  action RestockProduct
  api StockWebhook
}
```

A transaction block is top-level. It references existing custom actions and explicit APIs by name. It does not create a page, a route, or a new mutation.

## Shape

```text
transaction <PascalCaseName> {
  action <ActionName>
  api <APIName>
}
```

A block must contain at least one `action` or `api` target. Each target is one line. Each action or API can be bound to at most one transaction block.

API targets must be explicit API declarations with an `update Entity where ... set ...` handler. Read-only declared APIs do not need an atomic write boundary.

Service blocks can group the same explicit APIs for module metadata. Use [service.md](service.md) for API grouping and `transaction` for atomic runtime boundaries; they are separate top-level declarations.

## Generated Behavior

When a transaction targets a custom action, each generated page-bound action route wraps source row lookup, deterministic update, and auth audit logging in `prisma.$transaction`.

When a transaction targets an explicit API update handler, the generated Express route validates request values first, then wraps bounded row lookup, deterministic update, and private auth audit logging in `prisma.$transaction`.

Validation, authentication, page access, request parsing, and field permission checks remain outside the database transaction. The database work that must commit or roll back together is inside the transaction callback.

OpenAPI operations for targeted routes include:

```json
{
  "x-blacklang-transaction": true,
  "x-blacklang-transaction-name": "RestockAtomic"
}
```

## AI Agent Workflow

```bash
black docs transaction --json
black explain transaction --json
black inspect --affected RestockAtomic --json
black inspect --affected RestockProduct --json
black inspect --affected StockWebhook --json
black format --check --json
black lint --json
black validate --json
black build --json
```

Use [action.md](action.md) before changing custom action mutation logic. Use [api.md](api.md) before changing explicit API update handlers.

Common repairs:

| Code | Repair |
| --- | --- |
| `INVALID_TRANSACTION_DECLARATION` | Start with `transaction Name {`. |
| `INVALID_TRANSACTION_NAME` | Use PascalCase, such as `RestockAtomic`. |
| `EMPTY_TRANSACTION` | Add at least one `action Name` or `api Name` line. |
| `UNKNOWN_TRANSACTION_ACTION` | Declare the action or remove the target. |
| `UNKNOWN_TRANSACTION_API` | Declare the API or remove the target. |
| `UNSUPPORTED_TRANSACTION_API` | Target only APIs that declare an update handler. |
| `DUPLICATE_TRANSACTION_TARGET` | Keep one target line per action/API inside the block. |
| `DUPLICATE_TRANSACTION_BINDING` | Bind each action/API to at most one transaction block. |
