# BlackLang

BlackLang is an AI-native deterministic intent language.

Current status: the v0.1 roadmap is complete. The next milestone is tracked in `ROADMAP-v0.2.md`.

It helps AI agents move faster by turning clear application intent into validated, working software.

The official CLI release artifact layout is documented in `docs/release-artifacts.md`.

The planned npm wrapper for `npx blacklang` is documented in `docs/npm-wrapper.md`.

Install paths for local development, GitHub Releases, and npm are documented in `docs/install.md`.

GitHub publish rules are documented in `docs/github-publish.md`.

The first static documentation site source is under `website/`.

The goal is not to replace Python or JavaScript as a general-purpose language. The goal is to give AI coding agents a smaller, clearer, safer representation of application intent.

BlackLang source files are high-value source assets. Secrets should stay outside `.black` files, and production deployments should prefer generated artifacts instead of shipping the source of truth.

```text
Human request
  -> AI coding agent
  -> .black source files
  -> black compiler
  -> generated web application
```

## Current Status

BlackLang is at the project foundation stage.

The first milestone is a single-binary CLI that can:

- Read `.black` files
- Parse them into an AST
- Validate semantic rules
- Generate a basic web application
- Return machine-readable JSON for AI agents

## Initial Target

The first target is web application generation:

- React
- TypeScript
- API layer
- Database schema
- Validation
- CRUD pages

Future targets may include mobile, desktop, API-only, and automation outputs.

## Templates

Reusable `.black` app templates live under `examples/`.

- `examples/warehouse/app.black` demonstrates the first Warehouse MVP.
- `examples/crm/app.black` demonstrates a SalesCRM template with auth, roles, relations, workflows, components, API contracts, and validation.
- `examples/inventory/app.black` demonstrates an InventoryControl template with warehouses, suppliers, stock items, purchase orders, movement workflows, field-level access, and API contracts.
- `examples/helpdesk/app.black` demonstrates a SupportDesk template with customers, teams, SLAs, tickets, comments, knowledge articles, workflows, field-level access, and API contracts.
- `examples/invoice/app.black` demonstrates an InvoiceFlow template with clients, invoice lines, payments, credit notes, payment workflows, field-level access, and API contracts.
- `examples/appointment/app.black` demonstrates an AppointmentBook template with clients, services, staff, rooms, availability blocks, appointment workflows, reminders, waitlists, field-level access, and API contracts.
- `examples/project-management/app.black` demonstrates a ProjectPulse template with organizations, teams, projects, milestones, tasks, time entries, risks, project updates, workflows, field-level access, and API contracts.

## Example

```black
app Warehouse

entity Product {
  sku text required unique
  name text required
  stock number default 0
  price money
}

entity Customer {
  name text required
  email email unique
}

entity Order {
  customer Customer required
  total money default 0
}

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

## Planned CLI

```bash
black init
black format --check --json
black lint --json
black parse examples/warehouse/app.black --json
black parse examples/warehouse/app.black --ir
black validate --ir
black build --ir
black inspect --ir
black inspect --affected Product.stock --json
black agent startup --json
black theme inspect --json
black docs entity --ir
black docs diagnostics --json
black docs --all --json
black explain entity --json
```

Implemented so far:

- Initial Go CLI skeleton
- `init` command
- `parse <file>` command
- `format [file]` command with `--check`, `--stdout`, and `--json`
- `lint [file] --json` command for format, parse, validate, and source-security checks
- `validate <file>` command
- `build <file>` command
- `inspect` command
- `inspect --affected <symbol> --json` command for AI-readable impact analysis
- `agent startup --json` command for deterministic AI project entry checklists
- `theme inspect --json` command for `.blackthm` UI theme/profile files
- Compact UI slot profile rules in `theme inspect` output
- Append-only slot checks for locked UI profiles
- `docs <keyword>` command
- `docs --all --json` command for deterministic compact docs export
- Stable diagnostic documentation in `docs/diagnostics.md`
- `explain <keyword> --json` command for focused agent guidance
- `blacklang.toml` source/out config support
- `version --json` command
- JSON parse result shape
- Lexer-backed token stream before AST parsing
- JSON validation result shape
- JSON build result shape
- Compact BlackIR output with `--ir`
- Draft AST for `app`, `entity`, and `page`
- Parsed and validated secret-safe `database { url env DATABASE_URL }` declarations
- `black security scan --json` for likely hardcoded source secrets
- `black security encrypted-source --json` for protected source mode policy
- `black package --production` for deployable artifacts without protected source files
- First generated web files under `generated/`
- Generated React entry files, styles, and CRUD page behavior
- Generated validation files from field types and modifiers
- Generated Prisma database schema from entities
- Generated database workflow scripts and `.env.example`
- Optional generated `db:push:native` script for direct Prisma schema push checks
- Generated API client/server wiring for CRUD actions
- Generated loading, saving, and error UI states
- Generated Prisma-backed API routes for CRUD persistence
- Generated deterministic SQLite setup script for local MVP databases
- Generated read/detail UI behavior for selected records
- Generated bulk delete behavior from `actions delete`
- Generated archive/restore soft delete behavior from `actions archive, restore`
- Entity reference fields such as `customer Customer required`
- Generated Prisma relation fields and SQLite foreign key columns
- Generated relation select inputs in forms
- Generated relation display in tables and detail panels
- Generated empty-state guidance for required relation form fields
- Generated navigation from missing required relation fields to related pages
- Generated table search for relation display labels
- Field label modifiers such as `label "Product Name"`
- Generated form, table, and detail labels from field metadata
- Field placeholder modifiers such as `placeholder "Enter product name"`
- Generated input placeholders and relation select placeholder options
- Field help modifiers such as `help "Visible product name in lists"`
- Generated persistent form help text from field metadata
- Field constraint modifiers such as `min 0`, `max 100`, and `length 3..40`
- Advanced validation modifiers such as `regex "^[A-Z0-9]+$"`, `url`, and `message "Use uppercase letters and numbers"`
- Entity-level cross-field validation such as `validate discount <= total message "Discount cannot exceed total"`
- Conditional required validation such as `validate trackingNumber required when status == shipped message "Tracking number is required when shipped"`
- Generated frontend and API validation from field constraint metadata
- Generated OpenAPI contract at `generated/openapi.json`
- Generated Express route that serves `/openapi.json`
- Explicit `api` declarations for contract-first OpenAPI paths, query params, path params, access metadata, and webhook metadata
- Parsed and validated auth intent with `auth { strategy emailPassword session cookie }`
- Generated basic login/register UI shell from auth intent
- Generated auth API routes with password hashing and cookie sessions
- Generated cookie-session enforcement for CRUD API routes
- Generated `/api/auth/me` session restore and logout behavior
- Parsed and validated role declarations
- Parsed and validated page access declarations
- Role and access intent in JSON/BlackIR outputs
- Generated single-role storage for authenticated users
- Generated page-level role guards from `access`
- Generated basic Users role management page
- Generated action-level API permission guards
- Generated role-aware action controls in React pages
- Generated field-level read hiding in API responses and React pages
- Generated field-level mutation filtering for create/update payloads
- Generated audit log storage, API endpoint, and Audit page
- Generated CSRF protection for authenticated cookie write requests
- Parsed and validated workflow declarations
- Workflow intent in JSON/BlackIR outputs
- Generated workflow transition API routes, API clients, row action buttons, OpenAPI paths, and audit entries
- Parsed and validated client state declarations
- State intent in JSON/BlackIR outputs
- Generated React state hooks and modal helpers from matching state declarations
- Parsed and validated component declarations
- Component intent in JSON/BlackIR outputs
- Generated standalone React component files from component declarations
- Generated table/detail rendering through matching component inputs
- Generated live form previews through matching component inputs

## Documentation

- [Project Idea](blacklang-fikir.md)
- [Web Roadmap](blacklang-web-yol-haritasi.md)
- [Language Spec](SPEC.md)
- [Agent Guide](AGENTS.md)
- [Diagnostic Codes](docs/diagnostics.md)
- [Theme Profile](docs/theme-profile.md)
- [Warehouse Benchmark v0.1](benchmarks/warehouse-v0.1.md)
- [SalesCRM Benchmark v0.2](benchmarks/crm-v0.2.md)
- [InventoryControl Benchmark v0.2](benchmarks/inventory-v0.2.md)
- [SupportDesk Benchmark v0.2](benchmarks/helpdesk-v0.2.md)
- [InvoiceFlow Benchmark v0.2](benchmarks/invoice-v0.2.md)
- [AppointmentBook Benchmark v0.2](benchmarks/appointment-v0.2.md)
- [ProjectPulse Benchmark v0.2](benchmarks/project-management-v0.2.md)
