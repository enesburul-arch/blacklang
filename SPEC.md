# BlackLang Specification

Version: draft v0.1

## Purpose

BlackLang is an AI-native deterministic intent language.

BlackLang is designed for AI coding agents. Its source files should be easy to read, easy to modify, deterministic, and compact without becoming cryptic.

## File Extension

BlackLang source files use the `.black` extension.

BlackLang's compact intermediate representation uses the `.blackir` extension when saved to disk.

## Design Rules

- One concept should have one syntax.
- Source files describe application intent, not low-level implementation.
- Generated code must not be manually edited.
- BlackLang source files are high-value source assets and should be protected like a full source repository.
- Secrets must not be stored directly in `.black` source files.
- Every compiler error should include a stable error code.
- Every AI-facing command should support JSON output.
- BlackLang syntax should reuse familiar programming words where possible.
- BlackLang must be easy for an AI agent to learn from a short local guide.

## Current Parser Model

Draft v0.1 uses a lexer-backed token stream before AST parsing.

The lexer recognizes identifiers, quoted strings, `{`, `}`, commas, comparison operators, newlines, and comments. Comments beginning with `#` or `//` are ignored only outside quoted strings. Inline braces are split into deterministic statements before parsing, and unclosed strings report `UNCLOSED_STRING`.

## Current Version Command

Draft v0.2 supports human and machine-readable CLI version output:

```bash
black version
black version --json
```

JSON shape:

```json
{
  "success": true,
  "command": "version",
  "name": "black",
  "version": "0.1.0-dev",
  "errors": []
}
```

Release scripts, npm wrappers, CI pipelines, and AI agents should prefer `black version --json` when they need structured version checks.

## Current Format Command

Draft v0.2 supports deterministic source formatting:

```bash
black format app.black
black format app.black --check
black format app.black --stdout
black format app.black --check --json
```

When the file argument is omitted, the CLI reads `blacklang.toml` and formats the configured `source` path.

JSON shape:

```json
{
  "success": true,
  "command": "format",
  "version": "0.1.0-dev",
  "file": "app.black",
  "changed": false,
  "check": true,
  "stdout": false,
  "errors": []
}
```

`--check` never writes files. If formatting is required, it returns `success: false` with `FORMAT_REQUIRED`. `--stdout` prints formatted source without writing.

## Current Lint Command

Draft v0.2 supports read-only linting:

```bash
black lint
black lint app.black --json
```

When the file argument is omitted, the CLI reads `blacklang.toml` and lints the configured `source` path.

Lint checks:

- formatting
- parse diagnostics
- semantic validation diagnostics
- likely hardcoded source secrets

JSON shape:

```json
{
  "success": true,
  "command": "lint",
  "version": "0.1.0-dev",
  "file": "app.black",
  "summary": {
    "app": "Warehouse",
    "entities": 3,
    "pages": 3
  },
  "checks": [
    {
      "name": "format",
      "success": true,
      "findings": 0
    }
  ],
  "findings": [],
  "errors": []
}
```

`findings` contains source diagnostics that should be fixed before build. `errors` contains command-level failures such as unreadable files. The lint command does not write source files.

## Current Docs All Command

Draft v0.2 supports a complete compact documentation export:

```bash
black docs --all --json
```

JSON shape:

```json
{
  "success": true,
  "command": "docs",
  "version": "0.1.0-dev",
  "count": 31,
  "docs": [
    {
      "keyword": "entity",
      "purpose": "Declares stored application data and its fields.",
      "syntax": "entity <PascalCaseName> { <fieldName> <fieldType> <modifiers...> }",
      "example": "entity Product { ... }",
      "agentNotes": [],
      "errors": []
    }
  ],
  "errors": []
}
```

The `docs` array is sorted by keyword for deterministic AI context. Agents should use `black docs --all --json` when entering an unfamiliar BlackLang project, and `black docs <keyword> --json` for focused edits.

## Current Inspect Affected Command

Draft v0.2 supports focused affected graph output:

```bash
black inspect app.black --affected Product.stock --json
black inspect app.black --affected Products --json
black inspect app.black --affected OrderLifecycle --json
```

JSON shape:

```json
{
  "success": true,
  "command": "inspect",
  "version": "0.1.0-dev",
  "affected": {
    "symbol": "Product.stock",
    "kind": "field",
    "found": true,
    "entity": "Product",
    "field": "stock",
    "entities": [],
    "pages": [],
    "roles": [],
    "workflows": [],
    "states": [],
    "components": [],
    "apis": [],
    "generatedFiles": [],
    "agentNotes": []
  },
  "errors": []
}
```

Supported symbols include entities, entity fields such as `Product.stock`, pages, roles, workflows, states, components, APIs, `auth`, `database`, and `app`.

Unknown symbols return `UNKNOWN_AFFECTED_SYMBOL`. Missing `--affected` values return `MISSING_AFFECTED_SYMBOL`.

AI agents should run this before editing high-impact symbols so they can validate the right generated pages, routes, OpenAPI output, workflows, and role guards after the change.

## Current Explain Command

Draft v0.2 supports action-oriented keyword explanations:

```bash
black explain entity --json
black explain table --json
```

JSON shape:

```json
{
  "success": true,
  "command": "explain",
  "version": "0.1.0-dev",
  "keyword": "entity",
  "purpose": "Declares stored application data and its fields.",
  "syntax": "entity <PascalCaseName> { <fieldName> <fieldType> <modifiers...> }",
  "example": "entity Product { ... }",
  "agentSteps": [],
  "agentNotes": [],
  "related": ["page", "table", "form"],
  "errorCodes": ["DUPLICATE_ENTITY"],
  "errors": []
}
```

`explain` is read-only. It is intended for focused AI guidance when one keyword needs more operational context than `black docs <keyword> --json`. Unknown keywords return `UNKNOWN_EXPLAIN_KEYWORD`.

## BlackLang Source Security

As BlackLang grows, a compact `.black` source file can represent a large generated application. This makes source protection important.

Draft security principles:

- `.black` files are source-of-truth project assets.
- Secrets, passwords, API keys, tokens, and private keys must be referenced from environment or secret managers, not written directly in `.black`.
- Production deployments should prefer generated artifacts and should not require shipping `.black` source files to production servers.
- `black security scan --json` detects likely leaked secrets and reports machine-readable diagnostics.
- `black security encrypted-source --json` reports the draft `.black.enc` protected source policy for AI agents and CI tools.
- `black package --production` creates deployable artifacts without protected source files.

Current source-safe style:

```black
database {
  url env DATABASE_URL
}
```

## Current Database Declaration

Draft v0.1 supports a top-level `database` declaration:

```black
database {
  url env DATABASE_URL
}
```

The parser, validator, JSON output, and BlackIR output understand this declaration.

Draft v0.1 validates:

- Only one database block may exist.
- `url` must use `env`.
- Environment variable names must use uppercase letters, numbers, and underscores.
- Literal database URLs are rejected by the parser.

## Current Source Security Commands

Draft v0.1 supports source-security scanning:

```bash
black security scan --json
black security encrypted-source --json
```

The scan checks `.black` source for likely hardcoded database URLs, private keys, API keys, tokens, secrets, and passwords.

`security encrypted-source` reports the draft `.black.enc` protected source policy. In v0.1, direct build from encrypted source is planned rather than implemented.

Draft v0.1 also supports production packaging:

```bash
black package --production
```

The production package copies generated output while excluding `.black` source files, `.env`, local database files, `node_modules`, and generated Prisma client output.

Generated package scripts keep `db:push` mapped to BlackLang's deterministic SQLite setup for MVP safety. `db:push:native` is emitted as an explicit opt-in Prisma `db push` command for local schema-engine checks.

## Generated Form Validation

Generated web forms should reuse BlackLang field rules for inline validation messages.

In draft v0.1, this includes required fields, email fields, number-like fields, required relation fields, numeric `min`/`max`, text/email `length` ranges, text/email `regex` patterns, text `url` checks, custom field messages, and entity-level cross-field validation.

Inline messages should be generated from the same `.black` source metadata used by backend validation.

## Generated Table Sorting

Generated web tables may declare a default sort order inside `table` blocks:

```black
table {
  columns sku, name, stock
  search sku, name
  sort stock desc
}
```

In draft v0.1, the sort field must exist on the page source entity and the direction must be `asc` or `desc`.

Generated React lists apply sorting after search filtering.

## Generated Table Pagination

Generated web tables may declare a page size inside `table` blocks:

```black
table {
  columns sku, name, stock
  search sku, name
  sort stock desc
  paginate 25
}
```

In draft v0.1, pagination size must be a positive whole number.

Generated React lists apply pagination after search filtering and sorting.

## Generated Column Visibility

Generated web tables should derive column visibility controls from the `columns` list.

In draft v0.1, every listed column is visible by default and the generated React UI lets users hide or show individual columns locally.

This behavior does not require a new source syntax because the existing `columns` list already declares the manageable table columns.

## Generated Table Filters

Generated web tables may declare field-level filters inside `table` blocks:

```black
table {
  columns customer, total, status
  search customer, status
  filter customer, status
}
```

In draft v0.1, every filter field must exist on the page source entity.

Generated React lists apply filters after global search and before sort and pagination.

## Generated Application Shell

Generated web apps should wrap pages in a shared application shell.

In draft v0.1, the shell is derived from the `page` list and includes sidebar navigation, a topbar, and a breadcrumb.

Explicit `layout` declarations may define sidebar navigation order:

```black
layout AdminLayout {
  sidebar {
    item Products
    item Customers
    item Orders
  }
}

page Products {
  layout AdminLayout
  source Product
}
```

In draft v0.1, sidebar items must reference existing page names.

Generated app shells should adapt navigation for small screens.

In draft v0.1, generated apps keep the sidebar visible on desktop and use a menu button plus drawer navigation on narrow screens.

## Generated OpenAPI Contract

Generated web apps should include a machine-readable REST API contract.

In draft v0.1, `black build` writes:

```text
generated/openapi.json
```

The OpenAPI document is derived from page source entities and page actions.

Draft v0.1 also supports explicit top-level API contract declarations:

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
  webhook
  public
}
```

Explicit API declarations are contract-first in draft v0.1. They appear in parse JSON, BlackIR, inspect output, and `generated/openapi.json`. Runtime route generation is planned for a later API phase.

Generated collection paths use the page name:

```text
GET    /api/products
POST   /api/products
DELETE /api/products
```

Generated item paths use stable `id` parameters:

```text
GET    /api/products/{id}
PUT    /api/products/{id}
DELETE /api/products/{id}
PATCH  /api/products/{id}/archive
PATCH  /api/products/{id}/restore
```

Create, update, delete, archive, and restore operations are included only when the page declares the matching action.

The generated Express server serves this file at:

```text
/openapi.json
```

## Generated Secure Defaults

Generated web API servers should start with safe baseline behavior before explicit auth syntax exists.

In draft v0.1, generated Express servers:

- Disable the `X-Powered-By` header
- Add basic browser security headers
- Limit JSON request bodies to `100kb`
- Apply a simple IP-based request rate limit

These defaults are generated automatically and do not require source syntax in v0.1.

## Auth Declaration

BlackLang may declare authentication intent with a top-level `auth` block:

```black
auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}
```

In draft v0.1, `auth` supports:

- `strategy emailPassword`
- `session cookie`
- `user { ... }` fields using primitive field types

The parser, validator, JSON output, and BlackIR output understand this declaration.

Draft v0.1 generates a basic login/register UI shell from `auth`.

Draft v0.1 also generates basic register, login, logout, and current-user API endpoints with password hashing and cookie-backed session storage.

When `auth` exists, generated CRUD API routes require a valid cookie session.

Generated React apps check `/api/auth/me` on load and include a logout action.

Draft v0.1 generates CSRF protection for cookie-authenticated write requests.

Draft v0.1 does not generate password reset or OAuth from `auth` yet.

## Current Role And Access Declaration

Draft v0.1 supports top-level role declarations:

```black
role Admin {
  allow all
}

role Worker {
  allow read Product
  deny read Product price
}
```

It also supports page-level access declarations:

```black
page Products {
  source Product
  access Admin, Worker
}
```

The parser, validator, JSON output, and BlackIR output understand this intent.

Draft v0.1 validates:

- Duplicate role names
- Supported permission effects: `allow`, `deny`
- Supported permission actions: `all`, `manage`, `read`, `create`, `update`, `delete`
- Permission resources must reference existing entities unless action is `all`
- Page access values must reference existing roles, or use `authenticated`
- `authenticated` access requires an `auth` block

Draft v0.1 generates a basic runtime role foundation:

- `BlackUser` stores one `role` value.
- Newly registered users receive the first declared role by default.
- `/api/auth/me` returns the current user role.
- Pages with `access` generate API route guards.
- Users without an allowed role receive `403 Forbidden`.
- When roles exist, a generated Users page lets the first declared role list users and update their roles.
- Permission actions generate API route guards:
  - `read` protects list and detail endpoints.
  - `create` protects create endpoints.
  - `update` protects update, archive, and restore endpoints.
  - `delete` protects single and bulk delete endpoints.
- Generated React pages hide create, edit, archive, restore, and delete controls when the current role does not allow the required action.
- `deny` permissions override matching `allow` permissions.
- `allow all` allows every generated action.
- `allow manage Product` allows generated actions for `Product`.
- Field names after a permission resource scope the permission to those fields, such as `deny read Product price`.
- Generated API responses remove fields that the current role cannot read.
- Generated React pages hide table columns, detail fields, and form fields that the current role cannot read.
- Field-level mutation rules filter generated create/update payloads before database writes.
- For example, `allow update Product stock` lets the role call the update endpoint but only `stock` is written; other submitted Product fields are ignored.
- Generated React edit forms hide fields that the current role cannot update.
- Generated apps create a `BlackAuditLog` table when auth and roles exist.
- Generated create, update, archive, restore, delete, bulk delete, register, and role update operations write audit records.
- When roles exist, a generated Audit page lets the first declared role review recent audit records.
- Generated cookie auth uses an HttpOnly session cookie plus a readable CSRF cookie.
- Generated write requests send the CSRF token in `X-CSRF-Token`.
- Generated API routes reject state-changing authenticated requests when the CSRF cookie and header do not match.

This is page-level, action-level, and field-level role enforcement plus basic role management, audit log UI, and CSRF/session protection. Multiple roles per user, ownership rules, and tenant rules are planned for later security phases.

## AI Learning Cost

BlackLang is new, so AI agents may not know it from training data at first.

This is an expected early disadvantage. The language must be designed so the first learning cost is paid once per project, then repeated changes become cheaper than normal web-stack editing.

Target model:

```text
Initial learning cost
  +
low repeated change cost
```

BlackLang should avoid alien or overly compressed syntax. Prefer familiar words:

- `app`
- `entity`
- `page`
- `table`
- `form`
- `action`
- `if`
- `when`
- `return`

Avoid cryptic syntax such as:

```text
@!P[x?=>z:7]
```

AI agents should be able to infer much of the intent even before reading the full documentation.

## Agent Learning Pack

Every BlackLang project should include a short local learning pack:

- `BLACKLANG.md`
- `AGENTS.md`
- `SPEC.md`
- `blacklang.toml`

The learning pack should tell AI agents:

- Which BlackLang version is used
- Which files are source files
- Which files are generated
- Which syntax rules matter for the project
- Which commands validate the project
- Which docs pages to read when stuck

The goal is to prevent AI agents from reading a large documentation set for every task.

## Version Awareness

BlackLang projects should declare their language version.

Planned file:

```toml
version = "0.1"
target = "web"
source = "src/app.black"
out = "generated"
```

AI agents should use the version before applying syntax rules:

```bash
black version
black docs --version 0.1 --agent
```

This matters because future BlackLang versions may add syntax that older examples do not use.

## AI-Focused Docs Commands

BlackLang should expose small, task-specific documentation through the CLI:

```bash
black docs entity --json
black docs page --json
black explain table --json
black explain entity --json
black inspect app.black --affected Product.stock --json
```

This allows an AI agent to read only the relevant part of the language instead of a full manual.

Example output shape:

```json
{
  "keyword": "table",
  "version": "0.1",
  "purpose": "Defines list rendering for a page source entity.",
  "syntax": "table { columns fieldName... search fieldName... }",
  "examples": [
    "table { columns sku, name search sku, name }"
  ],
  "errors": ["UNKNOWN_TABLE_COLUMN", "UNKNOWN_SEARCH_FIELD"]
}
```

## BlackIR

BlackIR is BlackLang's compact, AI-readable intermediate representation.

JSON remains supported for external tools, editor plugins, APIs, and integrations. BlackIR exists because JSON can become verbose for large projects.

Format roles:

```text
.black    source language written by humans and AI agents
.blackir  compact BlackLang intermediate representation
.json     standard integration format for external tools
```

CLI commands may support:

```bash
black parse app.black --ir
black validate app.black --ir
black build app.black --ir
```

Example BlackIR:

```blackir
blackir 0.1

app Warehouse

entity Product
  sku text required unique
  name text required
  stock number default 0
  price money

page Products source Product
  table sku name stock price
  search sku name
  form sku name stock price
  actions create edit delete archive restore
```

BlackIR should be:

- Shorter than equivalent JSON
- Stable enough for AI agents to rely on
- Easier to scan than deeply nested JSON
- Lossless enough for compiler and agent inspection tasks
- Optional, not a replacement for JSON integrations

## Top-Level Blocks

Draft v0.1 supports:

- `app`
- `auth`
- `database`
- `entity`
- `layout`
- `page`
- `role`
- `api`
- `workflow`
- `state`
- `component`

## Current Workflow Declaration

Draft v0.1 supports top-level workflow declarations for business state intent:

```black
workflow OrderPreparation {
  source Order
  states draft, picking, verified, packaged, shipped

  transition ship {
    from packaged
    to shipped
    allow Admin
  }
}
```

The parser, validator, JSON output, BlackIR output, generated API routes, generated API clients, generated row action controls, OpenAPI output, and audit log understand workflow declarations.

Draft v0.1 validates:

- Duplicate workflow names
- Workflow source must reference an existing entity
- Workflow source entity must contain `status text`
- Workflow must declare at least one state
- Duplicate state names
- Duplicate transition names inside a workflow
- Transition `from` and `to` states must exist in the workflow state list
- Transition `allow` values must reference existing roles or `authenticated`
- Transition `allow` requires an `auth` block

Generated authenticated apps expose transition routes in this shape:

```text
POST /api/<pages>/:id/workflow/<transition>
```

The generated route requires update permission for the source entity, checks transition `allow` roles when present, verifies the current `status` equals `from`, updates `status` to `to`, and writes a `workflow.<transition>` audit record. Generated React pages show transition buttons on matching rows when the current user can update the source entity and the row status equals the transition `from` value.

## Current State Declaration

Draft v0.1 supports top-level state declarations for client-side UI state intent:

```black
state OrdersPageState {
  selectedOrders Order[]
  activeFilter text
  modal createOrder closed
}
```

The parser, validator, JSON output, BlackIR output, and generated React pages understand state declarations.

Draft v0.1 validates:

- Duplicate state names
- Duplicate state field names inside a state
- State field types must be primitive field types or existing entity names
- Entity list state uses `Entity[]`
- Duplicate modal names inside a state
- Modal defaults must be `open` or `closed`

Generated React pages bind matching state declarations by page name. `OrdersPageState` or `OrdersState` can attach to `page Orders`. State fields generate `useState` hooks. Modal declarations generate open/close helpers; a `modal createOrder closed` declaration controls the generated create form visibility for the `Order` page source.

## Current Component Declaration

Draft v0.1 supports top-level component declarations for reusable UI intent:

```black
component StockBadge {
  input stock number

  variant low when stock < 10
  variant normal when stock >= 10
}
```

The parser, validator, JSON output, BlackIR output, and generated React component files understand component declarations.

Draft v0.1 validates:

- Duplicate component names
- Duplicate component input names
- Component input types must be primitive field types or existing entity names
- Entity list inputs use `Entity[]`
- Duplicate variant names inside a component
- Variant conditions must use `when condition`

Draft v0.1 preserves variant conditions as deterministic intent strings and generates standalone React component files. The first runtime expression support handles simple `input operator literal` variant checks such as `stock < 10`. When a component has one input that matches an entity field name and type, generated table cells and detail fields render that field through the component. Matching generated form fields show a live component preview while the user edits the field value.

## App

```black
app Warehouse
```

The `app` declaration names the application.

Rules:

- Exactly one `app` declaration should exist per project.
- The app name should use PascalCase.

## Entity

```black
entity Product {
  sku text required unique
  name text required
  stock number default 0
  price money
}
```

An `entity` describes stored application data.

Field syntax:

```black
fieldName fieldType modifier...
```

Supported field types in v0.1:

- `text`
- `number`
- `integer`
- `decimal`
- `money`
- `email`
- `boolean`
- `date`
- `datetime`

Supported field modifiers in v0.1:

- `required`
- `unique`
- `optional`
- `default value`
- `label "Display Text"`
- `placeholder "Input Hint"`
- `help "Helper Text"`
- `min value`
- `max value`
- `length min..max`
- `regex "pattern"`
- `url`
- `message "Validation message"`

The `label` modifier controls generated UI text without renaming the stored field:

```black
name text required label "Product Name"
```

The `placeholder` modifier controls generated form input hints:

```black
name text required placeholder "Enter product name"
```

The `help` modifier controls persistent helper text below generated form fields:

```black
name text required help "Visible product name in lists"
```

The `min` and `max` modifiers constrain number-like fields:

```black
stock number min 0 max 100
```

The `length` modifier constrains text and email fields:

```black
sku text required length 3..40
```

The `regex` modifier constrains text and email fields with a deterministic pattern:

```black
sku text required regex "^[A-Z0-9]+$"
```

The `url` modifier constrains text fields that store web addresses:

```black
website text optional url
```

The `message` modifier overrides the generated validation message for that field:

```black
sku text required regex "^[A-Z0-9]+$" message "Use uppercase letters and numbers"
```

Entity-level `validate` lines compare two fields on the same entity:

```black
entity Order {
  total money
  discount money
  status text default draft
  trackingNumber text optional
  validate discount <= total message "Discount cannot exceed total"
  validate trackingNumber required when status == shipped message "Tracking number is required when shipped"
}
```

Draft v0.1 supports `==`, `!=`, `<`, `<=`, `>`, and `>=`. Ordering comparisons require number-like fields. Equality comparisons may use same-type fields. Conditional required validation uses `validate field required when otherField == value message "Text"` and is emitted to generated forms and API validation.

Entity reference fields are also supported in v0.1:

```black
entity Order {
  customer Customer required
}
```

Rules:

- The field type may be the name of an existing entity.
- Entity reference fields generate a foreign key field in database output.
- Entity reference fields used in page forms generate select inputs.
- Required entity reference fields disable generated submit buttons when no related records are available.
- Required entity reference fields may link to the related entity page when that page exists.
- Entity reference fields used in page tables and detail panels display the related record label when it is loaded by the generated API.
- Entity reference fields used in `table.search` search the generated relation display label.

## Page

```black
page Products {
  source Product

  table {
    columns sku, name, stock, price
    search sku, name
  }

  form {
    fields sku, name, stock, price
  }

  actions create, edit, delete, archive, restore
}
```

A `page` describes a generated web screen.

Rules:

- `source` must reference an existing entity.
- `table.columns` must reference fields on the source entity.
- `table.search` must reference searchable fields on the source entity. Draft v0.1 supports `text`, `email`, and entity reference search fields.
- `form.fields` must reference fields on the source entity.
- Form labels use `label "Text"` when present, otherwise the field name is title-cased.
- Form inputs use `placeholder "Text"` when present.
- Form fields show `help "Text"` below the generated input when present.
- `actions` may include `create`, `edit`, `delete`, `archive`, and `restore` in v0.1.
- `archive` uses soft delete behavior by setting `archivedAt`.
- `restore` clears `archivedAt`.
- Default list output excludes archived records.

## Planned CLI Contract

```bash
black init
black format <file> --check --json
black lint <file> --json
black parse <file> --json
black validate <file> --json
black build <file> --out generated --json
black validate --ir
black build --ir
black inspect --ir
black docs entity --ir
black docs --all --json
black explain entity --json
black inspect --json
```

## Validation Rules

Draft v0.1 validates:

- Missing app declaration
- Duplicate entities
- Duplicate fields
- Unsupported field types
- Unknown entity reference field types
- Unsupported field modifiers
- Missing label values
- Missing placeholder values
- Missing help values
- Missing constraint values
- Invalid numeric constraint values
- Invalid length constraint values
- Invalid entity validation fields
- Unsupported entity validation operators
- Incompatible entity validation fields
- Numeric constraints on non-numeric fields
- Length constraints on non-text fields
- Unknown page source entity
- Unknown table columns
- Unknown form fields
- Unknown search fields
- Search fields with unsupported types
- Unsupported page actions

## JSON Error Shape

```json
{
  "success": false,
  "errors": [
    {
      "file": "examples/warehouse/app.black",
      "line": 17,
      "column": 13,
      "code": "UNKNOWN_FIELD",
      "message": "Page Products uses unknown field barcode.",
      "suggestion": "Add barcode to Product or remove it from columns."
    }
  ]
}
```
