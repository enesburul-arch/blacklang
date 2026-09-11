# BlackLang

BlackLang is an AI-native deterministic intent language.

Current status: the v0.1 roadmap is complete. The next milestone is tracked in `ROADMAP-v0.2.md`.

It helps AI agents move faster by turning clear application intent into validated, working software.

The official CLI release artifact layout is documented in `docs/release-artifacts.md`.

The planned npm wrapper for `npx blacklang` is documented in `docs/npm-wrapper.md`.

Install paths for local development, GitHub Releases, and npm are documented in `docs/install.md`.

If `black` is not installed yet, the current public bootstrap path is source-based:

```bash
git clone https://github.com/enesburul-arch/blacklang.git
cd blacklang/packages/cli
go run ./cmd/black --help
go run ./cmd/black docs agent-contract --json
go run ./cmd/black validate ../../examples/warehouse/app.black --json
```

GitHub publish rules are documented in `docs/github-publish.md`.

The first static documentation site source is under `website/`.

The goal is not to replace Python or JavaScript as a general-purpose language. The goal is to give AI coding agents a smaller, clearer, safer representation of application intent.

BlackLang source files are high-value source assets. Secrets should stay outside `.black` files, and production deployments should prefer generated artifacts instead of shipping the source of truth.

```text
Human request
  -> AI coding agent
  -> .black source files
  -> black compiler
  -> generated web or API application
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

The same generator also supports an API-only application target for Node/Express services without React or Vite output.

Future targets may include mobile, desktop, and automation outputs.

## Templates

Reusable `.black` app templates live under `examples/`.

- `examples/warehouse/app.black` demonstrates Warehouse CRUD, local image media fields, computed inventory value with bounded arithmetic expressions, relation response load policies with generated batch prefetch, reusable page component sections, explicit rename migration intent, owner/tenant row policies, a LowStock page backed by a custom query with aggregate summary cards, a LowStockMonitor background query job, a transactional RestockProduct row action with local `value` and `if`/`else` logic, an explicit API block update handler, deterministic seed rows, and browser-check/browser e2e declarations with a generated browser matrix runner.
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
  photo image optional accept "image/*"
  stock number default 0
  price money
  computed inventoryValue money = stock * price label "Inventory Value"
}

entity Customer {
  name text required
  email email unique
}

entity Order {
  customer Customer required load detail query
  total money default 0
}

action RestockProduct {
  source Product
  input quantity number required min 1 label "Quantity"
  value restockValue = quantity
  if restockValue > 0 and stock >= 0
    set stock = stock + restockValue
  else
    set stock = stock
  success "Stock updated"
}

transaction RestockAtomic {
  action RestockProduct
}

page Products {
  source Product

  table {
    columns sku, name, stock, price, inventoryValue
    search sku, name
  }

  form {
    fields sku, name, stock, price
  }

  actions create, edit, delete, archive, restore, RestockProduct
}

test WarehouseBrowserSmoke {
  page Products
  expect text "Depo"
  expect page LowStock
  expect action RestockProduct
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
black theme migrate old.blackthm new.blackthm --json
black docs migrate --json
black explain migrate --json
black migrate plan old.black new.black --json
black docs seed --json
black explain seed --json
black inspect --affected DemoProducts --json
black docs test --json
black explain test --json
black inspect --affected WarehouseBrowserSmoke --json
black docs ui --json
black docs ui-modes --json
black docs entity --ir
black docs computed --json
black docs query --json
black explain query --json
black inspect --affected LowStockProducts --json
black docs action --json
black explain action --json
black inspect --affected RestockProduct --json
black docs policy --json
black explain policy --json
black inspect --affected policy --json
black docs benchmark --json
black benchmark --json
black benchmark tasks --json
black benchmark eval --json
black benchmark eval-history --json
black docs coverage --json
black benchmark coverage --json
black benchmark issues --json
black ide --json
black ide diagnostics examples/warehouse/app.black --json
black ecosystem --json
black docs release-trust --json
black docs package-registry --json
black docs adapter-marketplace --json
black docs editor-marketplace --json
node packages/registry/scripts/validate-registry.mjs
black audit accessibility --json
black security encrypt app.black --out app.black.enc --json
black security decrypt app.black.enc --stdout
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
- Standard UI mode groups for `box`, `text`, `table`, and `button`
- `theme migrate --json|--ir` command for safe old/new UI profile replacement checks
- `migrate plan --json|--ir` command for read-only old/new `.black` schema migration planning plus explicit rename declarations and generated `db:migrate:plan` / `db:migrate` runner scripts
- First-class seed/fixture declarations with generated deterministic `db:seed` runtime; see [docs/seed.md](docs/seed.md)
- First-class browser-check and browser e2e test declarations with generated deterministic page/action/text smoke checks, explicit real-browser execution, and a generated cross-browser matrix runner; see [docs/test.md](docs/test.md)
- Parsed and validated inline UI intent near fields, forms, tables, and action buttons
- Generated CSS classes from inline UI intent
- `.blackthm` `ui <mode> = <slot...>;` generator reading order
- Page view grid/stack/tabs composition with generated CSS, responsive breakpoint media queries, tab controls, modal/drawer section display, nested section groups, reusable component sections, and deterministic interaction triggers
- Computed display fields such as `computed inventoryValue money = stock * (price + tax)`
- Local `file` and `image` media fields with generated file inputs, data URL validation, table/detail rendering, and OpenAPI metadata; see [docs/media.md](docs/media.md)
- Custom queries with typed stored-field filters, deterministic sorting and limits, aggregate summary cards, page binding, permission checks, and JSON/BlackIR tooling; see [docs/query.md](docs/query.md)
- Background query jobs with deterministic `schedule every <integer> minutes|hours|days`, generated worker/manifest/package scripts, OpenAPI metadata, and JSON/BlackIR tooling; see [docs/job.md](docs/job.md)
- Custom actions with primitive inputs, ordered local `value` declarations, indentation-based `if`/`else` branches, deterministic row-level assignments, top-level transaction bindings, page binding, permission checks, generated route/client/UI/OpenAPI, and JSON/BlackIR tooling; see [docs/action.md](docs/action.md)
- Service blocks for grouping explicit API declarations into generated service manifests/modules, OpenAPI tags, route-level `x-blacklang-service` metadata, and JSON/BlackIR tooling; see [docs/service.md](docs/service.md)
- Owner and tenant entity row policies for generated list, query, detail, mutation, archive, delete, workflow transition, and custom action route scoping; see [docs/policy.md](docs/policy.md)
- Ops runtime signals with public health/readiness/metrics endpoints, structured request logs, OpenAPI metadata, generated smoke tests, and Docker healthchecks; see [docs/ops.md](docs/ops.md)
- `docs <keyword>` command
- `docs --all --json` command for deterministic compact docs export
- Stable diagnostic documentation in `docs/diagnostics.md`
- `explain <keyword> --json` command for focused agent guidance
- `benchmark [file] --json` command for deterministic source/generated output size reports
- `benchmark tasks --json|--ir` command for deterministic AI task benchmark scenarios and token estimate reports
- `benchmark eval --json|--ir` command for long-running AI eval corpus metadata
- `benchmark eval-history --json|--ir` command for evidence-backed local eval result history
- `benchmark coverage --json` command for weighted web coverage matrix and tracked issue status IDs
- `benchmark issues --json|--ir` command for compact tracked issue export
- `ide --json|--ir` command for compiler-owned IDE metadata, completion items, snippets, and diagnostic code catalog
- `ide diagnostics [file] --json|--ir` command for editor-friendly source diagnostics with zero-based ranges
- `ecosystem --json|--ir` command for release, package, registry, adapter, marketplace, extension, public index, signed release trust, release transparency, key rotation, and trust-policy discovery
- `packages/npm` thin npm/npx wrapper source for the native CLI
- Prepared local package registry manifests under `packages/registry`
- Prepared local provider adapter marketplace manifests under `adapters/marketplace`
- Prepared public ecosystem index source at `website/ecosystem-index.json`
- Local registry/marketplace trust validator at `packages/registry/scripts/validate-registry.mjs`
- Read-only signed release verifier at `scripts/verify-release-trust.mjs` for `release.blackdir`, `checksums.sha256`, detached Ed25519 `<artifact>.sig` files, trusted public key checks, plus prepared release transparency and key rotation policy manifests
- `audit accessibility --json|--ir` command for generated UI accessibility policy checks
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
- `black security encrypt <file>` and `black security decrypt <file.black.enc> --stdout` for `.black.enc` protected source workflows
- `black package --production` for deployable artifacts without protected source files
- First generated web files under `generated/`
- Generated React entry files, styles, and CRUD page behavior
- Generated validation files from field types and modifiers
- Generated `npm test` contract/API smoke tests for every target; `target web` also checks frontend render and declared browser-check expectations, and source `test` blocks add browser e2e plan/matrix commands
- Generated seed runtime and `db:setup` wiring for deterministic local/demo fixture rows
- Generated Prisma database schema from entities
- Generated Prisma and SQLite indexes from entity `index` declarations
- Read-only schema migration plan reports for safe, manual, destructive, and explicit rename entity/index changes
- Generated migration manifest, SQL previews, `src/migrate.ts`, and SQLite/PostgreSQL rename apply runtime
- Generated database workflow scripts and `.env.example`
- Generated secret/env reference manifest at `generated/security/secrets.json`, read-only `security:secrets:plan`, and provider CLI `security:secrets:preflight`
- Collection-backed page component sections with `bind each`
- Optional generated `db:push:native` script for direct Prisma schema push checks
- Generated `docker-compose.preview.yml` and preview package scripts from `deploy { preview local }`
- Generated deploy manifest, rollback metadata, and read-only rollback plan script from `deploy { rollback keep N }`
- Generated cloud adapter manifest, read-only cloud plan/preflight, and explicit provider CLI apply runner from `deploy { cloud fly app env FLY_APP_NAME region env FLY_REGION }`
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
- Relation response load policies such as `customer Customer load detail query`, generated batch prefetch, and nested relation response sanitization
- Generated empty-state guidance for required relation form fields
- Generated navigation from missing required relation fields to related pages
- Generated table search for relation display labels
- Field label modifiers such as `label "Product Name"`
- Generated form, table, and detail labels from field metadata
- Generated runtime language selector for i18n field/computed labels, generated app chrome/action/table/status copy, and stored field placeholder/help/message translations
- Locale-aware generated table/detail display formatting for number, integer, decimal, money, date, and datetime values
- Generated `lang`/`dir` app shell attributes for basic RTL locale support
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
- Generated background query worker at `generated/src/worker.ts`, job manifest at `generated/jobs/manifest.json`, and `jobs:run`/`jobs:loop` package scripts when `.black` declares jobs
- Generated contract/API tests for every target, plus frontend, optional browser-check, optional browser e2e, and optional browser matrix tests for `target web`; API-only output intentionally omits React/browser test files
- Generated public ops routes, external observability webhook/OTLP hooks, W3C trace context, and generated Docker healthchecks from `ops` declarations
- Explicit `api` declarations for generated declared-runtime routes, typed body fields, bounded single-line or block-form update handlers, local `value` and `if`/`else` handler logic, OpenAPI metadata, path/query params, access metadata, service grouping, and webhook 202 acknowledgements
- Parsed and validated auth intent with `auth { strategy emailPassword session cookie }`
- Generated basic login/register UI shell from auth intent
- Generated auth API routes with password hashing and cookie sessions
- Generated cookie-session enforcement for CRUD API routes
- Generated `/api/auth/me` session restore and logout behavior
- Parsed and validated role declarations
- Parsed and validated page access declarations
- Role and access intent in JSON/BlackIR outputs
- Generated multi-role storage for authenticated users with a backward-compatible primary `role`
- Generated page-level role guards from `access`
- Generated basic Users role management page for assigning one or more roles, plus tenant IDs when tenant policies exist
- Generated action-level API permission guards
- Generated role-aware action controls in React pages
- Generated custom action form panels and row controls from declared page actions
- Generated contract assertions for custom action OpenAPI metadata and action input validation
- Generated field-level read hiding in API responses and React pages
- Generated field-level mutation filtering for create/update payloads
- Generated owner/tenant policy stamping and route filtering for policy-scoped entities
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
- Parsed, validated, and generated page view order, DOM-order rendering, modal/drawer display, nested grouping, reusable component sections, and interaction triggers for table/detail/form/component sections
- Read-only accessibility audit diagnostics for generated overlays and nested view groups
- Compiler-owned IDE metadata, completion item, snippet, diagnostic code, editor diagnostic export, and packageable VS Code/Open VSX/Cursor-compatible bridge channel metadata
- Thin npm wrapper and ecosystem discovery metadata for release/package/registry/editor/adapter/marketplace/public-index tooling
- Signed release trust verification, release transparency policy, and key rotation policy for public install/package/adapter paths
- Documented AI agent capability boundary to prevent unsupported syntax invention

## License And Brand

BlackLang source code is licensed under the [MIT License](LICENSE). Copyright belongs to Muhammet Enes Burul.

MIT covers the software code. The BlackLang name, domain, logo, and related brand assets are not granted by the MIT License; see [Trademark And Brand Notice](TRADEMARKS.md).

## Documentation

- [Project Idea](blacklang-fikir.md)
- [Web Roadmap](blacklang-web-yol-haritasi.md)
- [Language Spec](SPEC.md)
- [Agent Guide](AGENTS.md)
- [Diagnostic Codes](docs/diagnostics.md)
- [Theme Profile](docs/theme-profile.md)
- [UI Profile](docs/ui-profile.md)
- [UI Modes](docs/ui-modes.md)
- [Inline UI Intent](docs/inline-ui.md)
- [Page View Composition](docs/view.md)
- [Schema Migration Plan](docs/migration.md)
- [Custom Queries](docs/query.md)
- [Custom Actions](docs/action.md)
- [State Declarations](docs/state.md)
- [File and Image Media Fields](docs/media.md)
- [Entity Row Policies](docs/policy.md)
- [Ops Runtime Signals](docs/ops.md)
- [Generated Tests](docs/generated-test.md)
- [Browser Check Tests](docs/test.md)
- [Accessibility Audit](docs/accessibility.md)
- [Benchmark Command](docs/benchmark.md)
- [IDE Contract](docs/ide.md)
- [Editor Marketplace Channels](docs/editor-marketplace.md)
- [Ecosystem Discovery](docs/ecosystem.md)
- [Package Registry](docs/package-registry.md)
- [Provider Adapter Marketplace](docs/adapter-marketplace.md)
- [Release Trust](docs/release-trust.md)
- [Protected Source](docs/protected-source.md)
- [Security CORS](docs/security-cors.md)
- [Secret Reference Manifest](docs/secret-management.md)
- [Deployment](docs/deployment.md)
- [AI Agent Contract](docs/ai-agent-contract.md)
- [Trademark And Brand Notice](TRADEMARKS.md)
- [Warehouse Benchmark v0.1](benchmarks/warehouse-v0.1.md)
- [SalesCRM Benchmark v0.2](benchmarks/crm-v0.2.md)
- [InventoryControl Benchmark v0.2](benchmarks/inventory-v0.2.md)
- [SupportDesk Benchmark v0.2](benchmarks/helpdesk-v0.2.md)
- [InvoiceFlow Benchmark v0.2](benchmarks/invoice-v0.2.md)
- [AppointmentBook Benchmark v0.2](benchmarks/appointment-v0.2.md)
- [ProjectPulse Benchmark v0.2](benchmarks/project-management-v0.2.md)
