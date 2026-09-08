# Custom Actions

Custom actions declare deterministic row-level mutations. They are for domain actions such as restocking a product, approving a request, or marking a task ready without writing manual route, client, validation, and UI code.

## Syntax

```black
action RestockProduct {
  source Product
  input quantity number required min 1 label "Quantity"
  value restockValue = quantity
  if restockValue > 0 and stock >= 0
    set stock = stock + restockValue
  else
    set stock = stock
  allow Admin, Worker
  success "Stock updated"
}

transaction RestockAtomic {
  action RestockProduct
}

page Products {
  source Product
  actions edit, RestockProduct
}
```

The declaration is top-level. The page binding is explicit: a custom action does not create a runtime endpoint until a page with the same source entity lists it in `actions`. Atomic runtime behavior is declared separately with a top-level [`transaction`](transaction.md) block.

## Shape

```text
action <PascalCaseName> {
  source <Entity>
  input <name> <text|email|number|integer|decimal|money|boolean|date|datetime> <modifiers...>
  value <name> = <value|numericExpression>
  if <condition>
    set <storedField> = <value|numericExpression>
  else
    set <storedField> = <value|numericExpression>
  set <storedField> = <value|numericExpression>
  allow <RoleName...|authenticated>
  success "Message"
}
```

`source` is required once. At least one reachable `set` is required. `input`, `value`, `if`, `allow`, and `success` are optional.

Action names use PascalCase and must remain unique after lowercasing. They also cannot collide with other top-level symbols such as entities, pages, queries, roles, workflows, components, APIs, layouts, transactions, `app`, `auth`, `database`, `target`, `security`, or `deploy`.

## Inputs

Action inputs use the same primitive field types as entity fields:

```text
text
email
number
integer
decimal
money
boolean
date
datetime
```

Supported input modifiers are:

```text
required
optional
default
label
placeholder
help
min
max
length
regex
url
message
```

Input names must not reuse source entity field names, including computed display fields. This keeps identifiers in `set` expressions unambiguous.

`number` and `integer` inputs currently validate as signed 32-bit whole numbers, matching the generated SQLite/Prisma integer runtime. Use `decimal` or `money` when fractional values are required.

## Set Expressions

Assignments can use typed literals, action inputs, and stored primitive fields on the source entity.

```black
set stock = stock + quantity
set price = price * (multiplier - discountRate)
set stock = stock + quantity / packSize
set active = true
set name = "Replacement name"
```

Arithmetic supports `+`, `-`, `*`, `/`, parentheses, and deterministic operator precedence. The target field and every arithmetic operand must be number-like: `number`, `integer`, `decimal`, or `money`. Literal division by zero is rejected at validation time, and dynamic division by zero is guarded in the generated route.

## Local Values And If/Else

Use `value` for a named local value inside one action. Values are evaluated in order, can be used by later `value`, `set`, and `if` statements, and do not create database fields or request inputs.

```black
action PriceProduct {
  source Product
  input quantity number required
  value subtotal = price * quantity

  if subtotal > 1000 and stock >= 0
    set status = "bulk"
    set stock = stock + quantity
  else
    set status = "standard"
}
```

Branch bodies are indentation-based: statements under `if` or `else` must be indented deeper than the `if` line, and `else` must align with its matching `if`. Supported branch statements are `value`, `set`, and nested `if`.

Condition comparisons use `==`, `!=`, `<`, `<=`, `>`, and `>=`. Equality compares compatible scalar type families. Ordered comparisons currently require numeric values. Comparisons may be joined with `and`, `or`, `not`, and parentheses, using `not` > `and` > `or` precedence. Local value names must be safe identifiers and must not collide with action inputs, source fields, computed fields, or reserved words.

The MVP does not support computed fields, relation fields, entity policy fields, joins, aggregates, raw SQL, function calls, loops, external service calls, or multi-row mutations inside custom actions.

## Transaction Binding

Use a top-level transaction block when the generated action route must run the row lookup, update, and audit write as one atomic database unit:

```black
transaction RestockAtomic {
  action RestockProduct
}
```

The transaction block does not create a second action syntax and does not change the action's mutation shape. Validation, auth, page access, and input parsing still run before the transaction; the source row lookup, deterministic assignment update, and auth audit write run inside `prisma.$transaction`.

## Generated Web Behavior

For each page that binds a custom action, BlackLang generates:

- a POST route at `/api/<lowercase-page>/:id/actions/<lowercase-action>`;
- a typed API client method such as `runRestockProduct(id, input)`;
- a validation function such as `validateRestockProductInput`;
- a React row button and action form panel;
- OpenAPI request and response schemas;
- audit log entries when auth and roles exist.

The generated route loads the existing row, validates input, applies deterministic assignments, updates the record through Prisma, and returns:

```json
{
  "item": {
    "id": "product-id"
  },
  "message": "Stock updated"
}
```

When a custom action is run from a query-bound page, the page refetches the server query after success. It does not guess whether the record still belongs in the filtered list.

OpenAPI custom action operations include `x-blacklang-transaction: true` and `x-blacklang-transaction-name` when a top-level transaction block targets the action.

## Authorization

Custom actions reuse existing generated security rules:

- page access still guards the bound page route;
- entity update permission is required when roles exist;
- field-level update permission is required for every field written by `set`;
- optional `allow` narrows access to listed roles or `authenticated`;
- entity row policies scope the row lookup before mutation when declared;
- response field hiding still applies.

If `allow` is present, the project must have an `auth` block. A role named in `allow` must be declared unless the value is `authenticated`.

## AI Agent Workflow

```bash
black docs action --json
black docs transaction --json
black explain action --json
black explain transaction --json
black inspect --affected RestockProduct --json
black inspect --affected RestockAtomic --json
black inspect --affected Product.stock --json
black format --check --json
black lint --json
black validate --json
black build --json
```

Use [diagnostics.md](diagnostics.md) for stable error codes and repair guidance. Change `.black` source or generator code; do not manually edit generated output.

Common repairs:

| Code | Repair |
| --- | --- |
| `MISSING_ACTION_SOURCE` | Add `source EntityName` inside the action. |
| `UNKNOWN_ACTION_SOURCE` | Use an existing entity in `source`. |
| `UNSUPPORTED_ACTION_TRANSACTION` | Move atomic intent to a top-level `transaction Name { action ActionName }` block. |
| `ACTION_INPUT_FIELD_COLLISION` | Rename the input so it cannot be confused with a source field. |
| `UNKNOWN_ACTION_FIELD` | Assign only to a stored field on the source entity. |
| `UNSUPPORTED_ACTION_FIELD` | Replace computed or relation fields with stored primitive fields. |
| `UNSUPPORTED_ACTION_POLICY_FIELD` | Let generated policy stamping manage owner/tenant fields instead of setting them in the action. |
| `UNKNOWN_ACTION_VALUE` | Use an action input, source field, or typed literal. |
| `INVALID_ACTION_VALUE_NAME` | Rename the local value to a safe non-reserved identifier. |
| `ACTION_VALUE_NAME_COLLISION` | Rename the local value so it does not collide with an input or source/computed field. |
| `INVALID_ACTION_IF` | Write `if condition` and indent branch statements below it. |
| `INCOMPATIBLE_ACTION_CONDITION` | Use ordered comparisons only with numeric values. |
| `ACTION_VALUE_TYPE_MISMATCH` | Assign a value compatible with the target field type. |
| `INCOMPATIBLE_ACTION_EXPRESSION` | Use arithmetic only with number-like target and operand types. |
| `UNKNOWN_ACTION_ALLOW_ROLE` | Declare the role or remove it from `allow`. |
| `PAGE_ACTION_SOURCE_MISMATCH` | Bind the custom action only on pages with the same source entity. |
