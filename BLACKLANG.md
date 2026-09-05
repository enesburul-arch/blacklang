# BlackLang Project Guide

This repository contains the BlackLang language, compiler, examples, and documentation.

v0.1 is the completed first web MVP. v0.2 planning lives in `ROADMAP-v0.2.md`.

Install paths are tracked in `docs/install.md`.

## What BlackLang Is

BlackLang is an AI-native deterministic intent language.

It describes what an application should do in a compact, predictable form that AI agents can read, edit, validate, and compile into working software.

AI agents should primarily edit `.black` files, then use the `black` CLI to validate and generate application code.

## Important AI Learning Principle

BlackLang is new, so an AI agent may not know it from training data.

That is expected.

The project must make the language easy to learn from local files and CLI output:

- Keep `BLACKLANG.md` short.
- Keep `SPEC.md` precise.
- Keep examples small.
- Use familiar programming words.
- Use stable JSON error codes.
- Prefer one syntax for one behavior.

The target is:

```text
First task: small learning cost
Later tasks: much lower change cost
```

## Parser Model

Draft v0.1 uses a lexer-backed token stream before AST parsing. This keeps quoted text, comments, braces, commas, and comparison operators deterministic for AI agents and compiler errors.

## Source Files

BlackLang source files use this extension:

```text
.black
```

BlackLang UI theme/profile files use this extension:

```text
.blackthm
```

Example source files live under:

```text
examples/
```

## Generated Files

Generated files should be treated as compiler output. Do not manually edit generated files unless a task explicitly asks for generator debugging.

## Source Security

BlackLang source files are high-value source assets because a compact `.black` file can represent a large generated application.

Keep secrets out of `.black` files. Use environment references such as:

```black
database {
  url env DATABASE_URL
}

security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}
```

Draft v0.1 parses and validates these `database` and `security.cors` declarations and includes them in JSON/BlackIR outputs. Literal database URLs are rejected. Generated web servers read comma-separated browser origins from the configured environment variable.

Production servers should receive generated production artifacts when possible, not the protected `.black` source of truth.

Useful source-security commands:

```bash
black security scan --json
black security encrypted-source --json
black package --production
```

The scan reports likely hardcoded secrets. `security encrypted-source` documents the `.black.enc` protected source policy. The production package excludes protected source, local secrets, local databases, dependencies, and generated Prisma client output.

Generated web projects keep `npm run db:push` mapped to BlackLang's deterministic SQLite setup in v0.1. They also expose `npm run db:push:native` for direct Prisma `db push` checks when the local Prisma schema-engine is stable.

## First Implementation Target

The first implementation target is a single-binary CLI written in Go.

The CLI should eventually run without requiring Python, Node.js, npm, or pip.

## Project Version File

BlackLang projects should eventually include:

```text
blacklang.toml
```

Draft example:

```toml
version = "0.1"
target = "web"
```

AI agents should use this file before deciding which syntax rules apply.

## AI-Friendly Commands

Every important command should support JSON output:

```bash
black parse app.black --json
black version --json
black format --check --json
black lint --json
black validate --json
black inspect --json
black inspect --affected Product.stock --json
black agent startup --json
black theme inspect --json
black docs ui --json
black docs ui-profile --json
black docs ui-modes --json
black docs entity --json
black docs diagnostics --json
black docs --all --json
black explain table --json
```

## Current Format Command

Draft v0.2 supports deterministic source formatting:

```bash
black format app.black
black format --check --json
black format app.black --stdout
```

When no file is provided, the CLI reads `blacklang.toml` and formats the configured source file. AI agents should prefer `black format --check --json` before validating or building AI-written source changes.

## Current Lint Command

Draft v0.2 supports a read-only lint command:

```bash
black lint --json
black lint app.black --json
```

Lint checks formatting, parse diagnostics, semantic validation diagnostics, and likely hardcoded source secrets in one report. It never writes files. AI agents should use the `checks` array to see which layer failed and the `findings` array to decide what to fix next.

## Current Docs All Command

Draft v0.2 supports a complete compact docs export:

```bash
black docs --all --json
```

The output contains every known `DocEntry`, sorted by keyword. AI agents should use this when they need the full local BlackLang reference, and use `black docs <keyword> --json` when only one concept is needed.

## Current Inspect Affected Command

Draft v0.2 supports focused impact analysis:

```bash
black inspect app.black --affected Product.stock --json
```

The affected output tells AI agents which entities, pages, roles, workflows, states, components, APIs, and generated files may change when a symbol is edited.

Use it before renaming or changing important fields such as `status`, relation fields, workflow source entities, or role-scoped fields.

## Current Agent Startup Command

Draft v0.2 supports a deterministic startup checklist for AI agents:

```bash
black agent startup --json
black agent startup app.black --json
```

The output tells an agent which local files to read first, which `.black` source file is the source of truth, which generated directory should be treated as rebuildable output, and which commands to run before and after edits.

Agents should run this when entering an unfamiliar BlackLang project instead of guessing the project workflow from memory.

## Current Theme Profile Format

Draft v0.2 supports a separate `.blackthm` source file for UI theme/profile metadata:

```blackthm
blackthm WarehouseTheme {
  version 1
  target web
  locked true

  token color primary "#2563eb"

  profile UICompact {
    version 1
    baseline box color width style pt pr pb pl radius place
    baseline text color size weight align
    baseline table color width style density zebra
    baseline button bg color radius size variant

    ui box = color width style pt pr pb pl radius place;
    ui text = color size weight align;
    ui table = color width style density zebra;
    ui button = bg color radius size variant;
  }
}
```

Projects can point to this file from `blacklang.toml`:

```toml
theme = "examples/warehouse/theme.blackthm"
```

AI agents can inspect it with:

```bash
black theme inspect --json
```

`theme inspect` also exposes compact UI profile rules:

- Slots are read left to right.
- Inline UI syntax is `ui <mode> <values...> [| <mode> <values...>...]`.
- Missing trailing values use defaults.
- Extra values are errors.
- Duplicate slots inside one mode are errors.
- Web UI profiles require standard `box`, `text`, `table`, and `button` modes.
- Locked profiles require `baseline <mode> <slot...>` lines.
- Current mode slots must keep baseline slots as their exact prefix.
- After profile lock, existing slots are immutable and new slots are append-only.

`ui <mode> = <slot...>;` is the generator reading order for compact inline UI values. Legacy `mode <name> <slot...>` lines are still accepted for backward compatibility.

This phase validates and exposes theme/profile metadata plus compact slot profile rules and locked append-only checks. `black build` uses the configured `.blackthm` profile order when mapping inline UI intent to generated CSS.

Standard mode groups:

- `box`: container border, spacing, radius, and placement
- `text`: typography for labels, headings, helper text, and body copy
- `table`: table-specific borders, density, and row patterns
- `button`: action control styling

## Current Inline UI Intent

Draft v0.2 can parse, validate, and generate CSS from compact `ui` intent inside `.black` source:

```black
entity Product {
  name text required ui text "#172026" 14 semibold left
}

page Products {
  source Product

  table {
    id ProductsTable
    class inventoryTable
    columns name
    ui table border 1 solid compact true
  }

  form {
    id ProductForm
    class inventoryForm
    fields name
    ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
  }

  actions create
  action create id CreateProductButton
  action create class primaryAction
  action create ui button primary white 6 md solid
}
```

Current rule:

- Field UI accepts `box` and `text`.
- Form UI accepts `box`, `text`, and `button`.
- Table UI accepts `box`, `text`, and `table`.
- Action button UI accepts `button`.
- Table and form blocks can declare `id Identifier` and `class ClassName...`.
- Page actions can declare `action <name> id Identifier` and `action <name> class ClassName...`.
- Generated web IDs and custom classes are normalized to kebab-case.
- Repeated row action IDs are expanded with generated suffixes so DOM IDs stay unique.
- Values are positional and should follow the active `.blackthm` profile mode slots.
- Generated web output appends stable `.bl-ui-*` classes to `src/styles.css`.
- The current generator uses the configured `.blackthm` slot order when available, and falls back to standard v0.2 slots and safe defaults.

## Current Diagnostic Documentation

Stable diagnostic code rules live in:

```text
docs/diagnostics.md
```

AI agents should use diagnostic `code` values, not message text, when deciding how to repair parser, validator, build, package, source-security, docs, or inspect errors.

The short CLI reference is available with:

```bash
black docs diagnostics --json
```

## Current Explain Command

Draft v0.2 supports action-oriented keyword explanations:

```bash
black explain entity --json
black explain table --json
```

Explain output includes purpose, syntax, example, agent steps, related keywords, agent notes, and error codes. AI agents should use it when a single BlackLang concept needs more guidance than the short docs entry.

## Current Page Actions

Draft v0.1 supports:

- `create`
- `edit`
- `delete`
- `archive`
- `restore`

## Current Field Labels

Draft v0.1 supports field labels as entity field modifiers:

```black
entity Product {
  name text required label "Product Name"
}
```

The generated web app uses labels in forms, table headers, and detail fields.

## Current Internationalization

Draft v0.2 supports a first i18n layer for generated field labels:

```black
i18n {
  default tr
  locales tr, en
}

label Product.name {
  tr "Ürün Adı"
  en "Product Name"
}
```

Current rule:

- Use one `i18n` block per project.
- The default locale must be included in `locales`.
- Top-level `label` blocks currently target entity fields with `Entity.field`.
- Generated web UI uses the default locale translation first.
- If no translation exists, generated UI falls back to the field `label "Text"` modifier, then title-cased field name.
- Runtime language switching, date format, number format, currency format, and RTL support come later.

## Current Field Placeholders

Draft v0.1 supports field placeholders as entity field modifiers:

```black
entity Product {
  name text required placeholder "Enter product name"
}
```

The generated web app uses placeholders in form inputs. For relation select fields, the placeholder becomes the empty select option.

## Current Field Help

Draft v0.1 supports field help text as an entity field modifier:

```black
entity Product {
  name text required help "Visible product name in lists"
}
```

The generated web app shows help text under form fields.

## Current Inline Form Validation

Draft v0.1 generates inline form validation messages from existing field rules.

Generated forms show field-level messages for required fields, email fields, number-like fields, required relation fields, numeric `min`/`max`, text/email `length` ranges, text/email `regex` patterns, and text `url` checks before sending invalid input to the API.

```black
entity Product {
  sku text required length 3..40 regex "^[A-Z0-9]+$" message "Use uppercase letters and numbers"
  stock number min 0
  website text optional url
}
```

The generated API validation uses the same field rules. `message "Text"` overrides the generated validation message for that field.

Draft v0.1 also supports entity-level cross-field validation:

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

Generated forms and API validation use the same cross-field and conditional rules.

## Current Relation Syntax

Draft v0.1 supports entity references as field types:

```black
entity Order {
  customer Customer required
}
```

The referenced type must be an existing entity.

When a relation field is used in a page form, the generated web app renders a select input and sends the matching foreign key to the API.

When a required relation field has no available related records, the generated form disables submit and tells the user which record must be created first.

When the related entity has its own generated page, the required relation guidance can navigate to that page.

When a relation field is used in a table or detail panel, the generated web app displays a readable label from the related record when available.

When a relation field is used in `table.search`, generated search uses that same readable relation label.

## Current Table Sorting

Draft v0.1 supports default table sorting inside table blocks:

```black
table {
  columns sku, name, stock
  search sku, name
  filter stock
  sort stock desc
  paginate 25
}
```

Generated React lists apply the sort after search filtering. Relation fields sort by their readable relation label when available.

`paginate 25` generates Previous/Next controls and shows 25 records per page after search and sort.

Generated tables also include column visibility controls derived from the `columns` list, so users can hide or show table columns without adding extra BlackLang syntax.

`filter stock` generates a field-level filter input. Relation filters use the readable relation label when available.

## Current Generated App Shell

Draft v0.1 can derive a shared application shell from the page list or from an explicit layout declaration.

Generated apps include sidebar navigation, a topbar, and a breadcrumb. If no explicit layout is declared, navigation falls back to the page order.

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

When a layout has sidebar items, generated navigation follows that order.

Generated app shells are responsive: desktop uses a sidebar, while narrow screens use a menu button and drawer navigation.

## Current OpenAPI Output

Draft v0.1 generates an OpenAPI contract for web targets:

```text
generated/openapi.json
```

The contract is derived from page source entities and page actions. It describes generated REST paths, request bodies, response schemas, relation ID fields, and field formats such as email.

The generated Express server also serves it at:

```text
/openapi.json
```

Draft v0.1 also supports explicit API contract declarations:

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

Explicit API declarations are contract-first in v0.1: they are parsed, validated, included in JSON/BlackIR, and written to `generated/openapi.json`. Runtime implementation generation comes later.

## Current Security Defaults

Draft v0.1 generated API servers include baseline secure defaults:

- Disabled `X-Powered-By`
- Basic security headers
- `100kb` JSON body limit
- Simple IP-based rate limit

These are generated automatically. Future versions should add explicit auth, role, permission, and policy syntax.

Draft v0.1 also supports explicit CORS intent:

```black
security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}
```

Generated Express servers read `CORS_ORIGINS` as a comma-separated list, reject unlisted browser origins, answer `OPTIONS` preflight requests, and add credential headers when `credentials true` is declared.

## Current Target Declaration

Draft v0.1 supports an explicit top-level target declaration:

```black
target web {
  frontend react
  backend node
  database sqlite
}
```

This block declares the intended generated application stack for the current `.black` source. In v0.1, the only supported stack is web output with React frontend, Node backend, and SQLite database runtime.

If the block is omitted, the generator keeps the legacy default: `web`, `react`, `node`, `sqlite`.

Rules:

- Use one `target` block per project.
- Existing v0.1 generator output supports only `target web`.
- `frontend react`, `backend node`, and `database sqlite` are required when the block exists.
- PostgreSQL, mobile, desktop, API-only, and alternate backend targets must wait until matching generator support exists.
- AI agents should use `black docs target --json` and `black inspect --affected target --json` before changing target metadata.

## Current Deployment Declaration

Draft v0.1 supports Docker deployment intent:

```black
deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  env CORS_ORIGINS optional
}
```

When `target docker` is declared, the web generator writes `Dockerfile`, `.dockerignore`, and `docker-compose.yml`, adds `PORT` to `.env.example`, adds a `start` script to generated `package.json`, makes the generated server read the configured port environment variable, and serves the built Vite frontend from `dist`.

Draft v0.1 keeps the generated runtime on SQLite. PostgreSQL deployment syntax should wait until the generator can produce a matching PostgreSQL runtime.

## Current Auth Declaration

Draft v0.1 supports a top-level auth declaration:

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

The compiler parses and validates this intent and includes it in JSON/IR outputs.

Draft v0.1 generates a basic login/register UI shell from this auth intent.

Draft v0.1 also generates basic register, login, logout, and current-user API endpoints with password hashing and cookie-backed session storage.

When `auth` exists, generated CRUD API routes require a valid cookie session.

Generated React apps check `/api/auth/me` on load and include a logout action.

Draft v0.1 generates CSRF protection for cookie-authenticated write requests.

Draft v0.1 does not generate password reset or OAuth from `auth` yet.

## Current Role And Access Declaration

Draft v0.1 supports permission intent in `.black` source:

```black
role Admin {
  allow all
}

role Worker {
  allow read Product
  deny read Product price
  allow read Customer
  allow read Order
}

page Products {
  source Product
  access Admin, Worker
}
```

The compiler parses and validates roles and page access, and includes them in JSON/IR outputs.

Generated auth stores one user role, assigns the first declared role to newly registered users, returns the role from `/api/auth/me`, and blocks page API routes when the current user's role is not allowed.

When roles exist, generated apps include a basic Users page where the first declared role can list users and update roles.

Generated API routes also use role permission actions:

- `read` for list/detail
- `create` for create
- `update` for edit/archive/restore
- `delete` for delete/bulk delete

Generated React pages hide action controls when the current role is not allowed to use them.

Field names after a permission resource scope the permission to those fields. For example, `deny read Product price` lets the role read Product records while hiding the `price` field from generated API responses and React views.

Field-level mutation is also enforced. For example, `allow update Product stock` lets the role update only `stock`; other submitted Product fields are ignored by generated API routes.

Generated apps also create an audit log table when auth and roles exist. Create, update, archive, restore, delete, bulk delete, register, and role update operations write audit records. The first declared role can open the generated Audit page to review recent activity.

Generated cookie auth uses an HttpOnly session cookie plus a readable CSRF cookie. Generated write requests send the CSRF token in `X-CSRF-Token`, and the API rejects authenticated state-changing requests when the cookie and header do not match.

This is the first runtime authorization foundation. Multiple roles per user, ownership, and tenant rules come later.

## Current Workflow Declaration

Draft v0.1 supports top-level workflow intent:

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

The compiler parses and validates workflow source entities, states, transitions, and transition allow roles. Workflow intent appears in JSON and BlackIR outputs.

In generated authenticated web apps, workflow source entities must have `status text`. The compiler generates `POST /api/<pages>/:id/workflow/<transition>` routes, matching API client methods, row action buttons, OpenAPI paths, role checks, state checks, status updates, and `workflow.<transition>` audit log entries.

## Current State Declaration

Draft v0.1 supports top-level client state intent:

```black
state OrdersPageState {
  selectedOrders Order[]
  activeFilter text
  modal createOrder closed
}
```

The compiler parses and validates state fields and modal defaults. State intent appears in JSON and BlackIR outputs. Generated React pages bind matching state declarations by page name, generate `useState` hooks, and create modal open/close helpers such as `openCreateOrder` and `closeCreateOrder`.

## Current Component Declaration

Draft v0.1 supports top-level reusable UI component intent:

```black
component StockBadge {
  input stock number

  variant low when stock < 10
  variant normal when stock >= 10
}
```

The compiler parses and validates component inputs and variants. Component intent appears in JSON and BlackIR outputs. The generator creates standalone React component files and can turn simple variant conditions such as `stock < 10` into runtime class selection. When a component has one input that matches an entity field name and type, generated table cells and detail fields render that field through the component, and generated form fields show a live preview while the value is edited.
