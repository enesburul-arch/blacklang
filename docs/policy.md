# Entity Row Policies

Entity policies declare ownership and tenant row scope for generated web routes.

## Syntax

```black
auth {
  strategy emailPassword
  session cookie
  user {
    name text required
    email email required
  }
}

entity Order {
  tenantId text required default "default"
  ownerId text required default "system"
  total money default 0
  status text default draft

  policy tenant tenantId
  policy owner ownerId
}

page Orders {
  source Order
  actions create, edit, delete

  table {
    columns total, status
  }

  form {
    fields total, status
  }
}
```

The canonical policy lines are:

```black
policy owner ownerId
policy tenant tenantId
```

There is no short field-like syntax for policy declarations. `owner Customer required` remains a normal relation field; a row policy is always written with the `policy` keyword.

## Rules

- Policy lines live inside an `entity`.
- A project must declare `auth` before using entity policies.
- Each entity can declare at most one `owner` policy and one `tenant` policy.
- The referenced policy field must be a stored `text required` field on the same entity.
- Policy fields must not be `unique`; many rows can share the same owner or tenant.
- Computed fields, relation fields, and generated system fields cannot be policy fields.
- Policy fields stay out of generated forms and API client input types.
- Custom actions cannot set policy fields.

## Generated Behavior

`policy owner ownerId` scopes rows to the authenticated user's `id`.

`policy tenant tenantId` scopes rows to the authenticated user's `tenantId`. The generated auth runtime stores `tenantId` on `BlackUser`, defaults new users to `"default"`, and lets the first declared role update user tenant IDs from the generated Users page.

Generated create and update routes stamp policy fields from the authenticated user context. User-submitted values for policy fields are stripped from writable input. Generated list, custom query, detail, archive, restore, delete, workflow transition, and custom action routes apply the same row scope before loading or mutating rows.

Queries remain deterministic list-selection rules. A query can say “low stock products”; a policy says “only rows in the current user/tenant scope.” When both exist, the generated query endpoint applies row policy, archive visibility, query filters, sort, and limit.

## AI Agent Workflow

```bash
black docs policy --json
black explain policy --json
black inspect --affected policy --json
black inspect --affected Order.ownerId --json
black format --check --json
black lint --json
black validate --json
black build --json
```

Use [diagnostics.md](diagnostics.md) for stable diagnostic codes and repair guidance. Change `.black` source or generator code; do not manually edit generated route, API client, OpenAPI, Prisma, or setup output.

Common repairs:

| Code | Repair |
| --- | --- |
| `INVALID_ENTITY_POLICY` | Use `policy owner field` or `policy tenant field` inside an entity. |
| `AUTH_REQUIRED_FOR_ENTITY_POLICY` | Add an `auth` block or remove the policy. |
| `DUPLICATE_ENTITY_POLICY` | Keep only one policy per kind on the entity. |
| `DUPLICATE_ENTITY_POLICY_FIELD` | Use different fields for owner and tenant policies. |
| `UNKNOWN_ENTITY_POLICY_FIELD` | Declare the referenced stored field on the same entity. |
| `UNSUPPORTED_ENTITY_POLICY_FIELD` | Use a stored `text` field, not a computed or relation field. |
| `MISSING_ENTITY_POLICY_REQUIRED_FIELD` | Add `required` to the policy field. |
| `UNSUPPORTED_ENTITY_POLICY_UNIQUE_FIELD` | Remove `unique` from the policy field. |
| `UNSUPPORTED_POLICY_FORM_FIELD` | Remove the policy field from generated page forms. |
| `UNSUPPORTED_ACTION_POLICY_FIELD` | Remove the policy field from custom action `set` clauses. |
