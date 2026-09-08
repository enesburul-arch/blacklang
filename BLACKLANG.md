# BlackLang Project Guide

This repository contains the BlackLang language, compiler, examples, and documentation.

v0.1 is the completed first web MVP. v0.2 planning lives in `ROADMAP-v0.2.md`.

Install paths are tracked in `docs/install.md`.

## What BlackLang Is

BlackLang is an AI-native deterministic intent language.

It describes what an application should do in a compact, predictable form that AI agents can read, edit, validate, and compile into working software.

AI agents should primarily edit `.black` files, then use the `black` CLI to validate and generate application code.

## Current AI Agent Contract

External AI agents should treat the official BlackLang path as:

```text
.black source -> black CLI -> generated web/API output
```

They should not invent browser runtime syntax such as:

```html
<script type="text/black">
```

unless an official BlackLang browser runtime exists.

Current BlackLang is not a full JavaScript, Python, game, or calculator runtime. It is currently strongest for generated CRUD/admin-style web applications and API-only Node services with auth, multi-role assignments, tenant admin UI for tenant policies, owner/tenant row policies, relations with response load policies, workflow, validation, local file/image media fields, computed display fields with bounded arithmetic expressions, entity indexes, schema migration planning, explicit rename migration runtime, generated online migration runner safeguards, first-class seed/fixture declarations, custom queries, custom query aggregate summaries, deterministic background query worker jobs, custom row-level actions with bounded arithmetic set expressions, top-level transaction blocks for custom actions and explicit API update handlers, service blocks for explicit API module metadata, first-class browser-check and browser e2e test declarations for web targets, generated browser e2e matrix plan/run support, explicit API declared-runtime routes and bounded update handlers, i18n field/UI text, locale-aware display formatting, basic RTL direction support, inline UI intent, page view order, page view DOM-order rendering, modal/drawer section display, nested section groups, view interaction triggers, OpenAPI, generated contract/API/frontend/browser-check smoke tests, generated real browser e2e checks, Docker deployment, local preview deploy intent, rollback metadata, cloud adapter plan/preflight metadata, generated provider CLI deployment preflight/apply runner, generated secret reference manifests, read-only secret provider preflight execution, ops health/readiness/metrics signals, external observability webhook/OTLP hooks, W3C trace context, generated observability exporter manifests, source-security checks, protected source encryption, production packaging, signed release trust verification, release transparency log policy, key rotation policy, source/generated benchmark reports, AI task benchmark, long-running eval corpus, evidence-backed eval history reports, AI-readable CLI outputs, and compiler-owned IDE metadata/diagnostics.

Full rules live in:

```text
docs/ai-agent-contract.md
```

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

Generated apps also write `security/secrets.json`, `scripts/secrets-plan.mjs`, and `scripts/secrets-provider.mjs`. The manifest lists environment-backed secret/config references with source, kind, required, and sensitive metadata, but never stores values. Run `npm run security:secrets:plan` for environment readiness and `npm run security:secrets:preflight` for provider label, prefix, and CLI readiness without fetching or printing secret values.

Production servers should receive generated production artifacts when possible, not the protected `.black` source of truth.

Useful source-security commands:

```bash
black security scan --json
black security encrypted-source --json
black security encrypt app.black --out app.black.enc --json
black security decrypt app.black.enc --stdout
black package --production
```

The scan reports likely hardcoded secrets. `security encrypted-source` documents the `.black.enc` protected source policy. `security encrypt` creates AES-GCM encrypted BlackLang source from an environment key, and `security decrypt --stdout` prints plaintext only when explicitly requested. Parse, lint, validate, inspect, benchmark, security scan, and build can read `.black.enc` source in memory when the header-declared key environment variable is set. Generated `security/secrets.json`, `npm run security:secrets:plan`, and `npm run security:secrets:preflight` expose secret/config reference readiness and provider CLI preflight without printing values. The production package excludes protected source, local secrets, local databases, dependencies, and generated Prisma client output.

Generated web projects keep `npm run db:push` mapped to BlackLang's generated database setup. SQLite targets use deterministic local table setup; PostgreSQL and MySQL targets use the generated Prisma schema through `prisma db push`. They also expose `npm run db:push:native` for direct Prisma checks. When migration blocks exist, generated apps expose `npm run db:migrate:plan` and `npm run db:migrate` so deploy agents can inspect pending rename operations, apply only declared renames, and then run `npm run db:setup`.

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
black docs agent-contract --json
black theme inspect --json
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
black docs ui-profile --json
black docs ui-modes --json
black docs view --json
black docs computed --json
black docs query --json
black explain query --json
black docs action --json
black docs transaction --json
black docs service --json
black explain action --json
black explain transaction --json
black explain service --json
black inspect --affected RestockProduct --json
black inspect --affected RestockAtomic --json
black docs ops --json
black explain ops --json
black inspect --affected ops --json
black docs benchmark --json
black benchmark --json
black benchmark tasks --json
black ide --json
black ide diagnostics app.black --json
black ecosystem --json
black audit accessibility --json
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

## Current Accessibility Audit Command

Draft v0.2 supports a read-only generated UI accessibility policy audit:

```bash
black audit accessibility --json
black audit accessibility app.black --ir
```

The command parses and validates source first. It then reports policy findings such as modal/drawer sections without explicit `title` text and multi-section view groups without explicit `title` text. It never writes source or generated files.

## Current Docs All Command

Draft v0.2 supports a complete compact docs export:

```bash
black docs --all --json
```

The output contains every known `DocEntry`, sorted by keyword. AI agents should use this when they need the full local BlackLang reference, and use `black docs <keyword> --json` when only one concept is needed.

## Current IDE Command

Draft v0.2 supports compiler-owned IDE metadata and source diagnostics:

```bash
black ide --json
black ide --ir
black ide diagnostics app.black --json
black ide diagnostics app.black --ir
```

`black ide --json` returns BlackLang language id, source extensions, completion items, snippets, and a diagnostic code catalog. `black ide diagnostics <file> --json` returns zero-based editor ranges for format, parse, validate, and source-security diagnostics. A readable source file with findings returns `success: true` and `valid: false`, so editor integrations can display ordinary source diagnostics without treating the CLI call as a transport failure.

`editors/vscode-blacklang` contains a packageable VS Code bridge that consumes these commands for syntax highlighting, dynamic completions/snippets, diagnostics, a `FORMAT_REQUIRED` quick fix, and affected-symbol inspection.

## Current Ecosystem Command

Draft v0.2 supports deterministic ecosystem discovery:

```bash
black ecosystem --json
black ecosystem --ir
```

The result lists release manifest/checksum/signature trust metadata, release scripts, local packageable sources such as `packages/npm` and `editors/vscode-blacklang`, prepared package registry manifests, built-in target/deploy/observability/editor adapters, prepared provider adapter marketplace manifests, and extension trust workflow checks. Provider-specific behavior should stay in extensions or adapters until parser, validator, docs, JSON/BlackIR, diagnostics, affected graph, tests, manifest metadata, signed package verification, and trust checks make the behavior ready.

Release trust uses a single verification contract:

```bash
black docs release-trust --json
node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict
```

Public install paths are trusted only when `release.blackdir`, `checksums.sha256`, archive bytes, detached Ed25519 `<artifact>.sig` files, and a public key from `BLACKLANG_RELEASE_PUBLIC_KEY` or `BLACKLANG_RELEASE_PUBLIC_KEY_FILE` agree. Ecosystem metadata also exposes prepared release transparency and key rotation policies: append-only `transparency.blackdir` entries, `sha256-public-key-spki-prefix` key ids, a 30 day overlap rule, and `key-revocations.blackdir`. Private signing keys stay outside `.black` source, generated output, manifests, and package metadata.

Current registry and marketplace manifests live at:

```text
packages/registry/package-index.blackdir
packages/registry/trust-policy.blackdir
adapters/marketplace/adapter-index.blackdir
adapters/marketplace/trust-policy.blackdir
```

Validate them with:

```bash
node packages/registry/scripts/validate-registry.mjs
```

## Current Inspect Affected Command

Draft v0.2 supports focused impact analysis:

```bash
black inspect app.black --affected Product.stock --json
```

The affected output tells AI agents which entities, queries, background jobs, custom actions, transaction blocks, seeds, pages, roles, workflows, states, components, APIs, ops metadata, i18n metadata, and generated files may change when a symbol is edited.

Use it before renaming or changing important fields such as `status`, relation fields, workflow source entities, seed rows, custom actions, or role-scoped fields.

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
black theme migrate old.blackthm new.blackthm --json
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

This phase validates and exposes theme/profile metadata plus compact slot profile rules, locked append-only checks, and old/new migration safety checks. `black theme migrate <old.blackthm> <new.blackthm> --json` reports whether existing inline UI values keep the same positional meaning before a theme is replaced. `black build` uses the configured `.blackthm` profile order when mapping inline UI intent to generated CSS.

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

## Current Page View Order and Composition

Draft v0.2 supports a `view` block inside a page for ordering and composing generated page sections:

```black
page Products {
  source Product

  view {
    order table, StockSummary, StockCards, detail, form
    compose grid columns 2 gap md stackAt md
    section table span 2
    section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
    section StockCards component StockBadge bind each span 1 title "Stock Cards"
    section detail display drawer side right title "Product Details"
    section form display modal title "Product Form"
    trigger detail on rowSelect
    trigger form on createStart
    trigger form on editStart
    trigger detail on saveSuccess
    trigger table on close
  }
}
```

Current rule:

- `view` belongs inside one `page`.
- `order` may list `table`, `detail`, `form`, and declared component sections.
- Listed sections render first and generated JSX/DOM follows the effective order.
- Omitted supported sections are appended in default `table, detail, form` order.
- `section <Name> component <Component> bind selected|first|each` places a declared component as a reusable page panel.
- `bind selected` passes the current selected/detail record; `bind first` passes the first loaded list record; `bind each` renders one component instance per loaded list/query record.
- Component section inputs must be scalar primitive inputs whose names and types match stored or computed fields on the page source entity.
- Component sections render inline in this MVP; modal/drawer component sections are intentionally rejected.
- `compose stack` keeps vertical section flow.
- `compose grid` emits deterministic CSS grid rules with `columns 1..4`, `gap sm|md|lg`, and `stackAt sm|md|lg|none`.
- Grid `stackAt` emits responsive breakpoint media queries; `stackAt sm` can step 4 columns through lg/md/sm before the final single-column stack.
- Section spans that exceed the responsive column count are clamped at that breakpoint.
- `compose tabs` emits generated React tab controls with local `activeViewTab` state.
- `section table|detail|form|<ComponentSection> span 1..4` controls grid span for one generated section.
- `section detail|form display modal|drawer` opens generated detail/form panels as overlays instead of inline panels.
- `side left|right` controls drawer placement and is valid only with `display drawer`.
- `title "Text"` sets the generated modal/drawer heading.
- `group <Name> sections StockSummary, detail compose stack|grid` wraps contiguous inline sections in one generated layout container.
- Group `span 1..4` applies to the wrapper inside an outer grid; group `title "Text"` renders a generated heading.
- `tab <Name> sections table, StockSummary, detail` assigns supported sections to a tab and requires every ordered section to appear exactly once.
- `trigger <section> on <event>` binds supported generated interactions to section state changes.
- Supported trigger events are `rowSelect`, `createStart`, `editStart`, `saveSuccess`, and `close`.
- Use `trigger detail on rowSelect` to switch/show detail after a generated View action loads a row.
- Use `trigger form on createStart` and `trigger form on editStart` to switch/show the generated form.
- Use `trigger detail on saveSuccess` or `trigger table on saveSuccess` to choose the post-save section.
- Use `trigger table on close` to return tabbed or overlay flows to the list section.
- Duplicate or unsupported section names are validation errors.
- The current web generator emits stable `.bl-view-section-*`, `.bl-view-component-section`, `.bl-view-component-*`, `.bl-view-group-*`, `.bl-view-compose-*`, `.bl-view-span-*`, `.bl-view-display-*`, `.bl-view-side-*`, and `.bl-view-has-triggers` classes and deterministic CSS order/composition/overlay/group/trigger behavior.
- Arbitrary coordinates and arbitrary frontend event handlers are later layout features.

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

Draft v0.2 supports runtime i18n for generated field text and display formatting:

```black
i18n {
  default tr
  locales tr, en
}

label Product.name {
  tr "Ürün Adı"
  en "Product Name"
}

label app.title {
  tr "Depo"
  en "Warehouse"
}

label page.Products {
  tr "Ürünler"
  en "Products"
}

label action.create {
  tr "Oluştur"
  en "Create"
}

placeholder Product.name {
  tr "Ürün adını gir"
  en "Enter product name"
}

help Product.name {
  tr "Listelerde görünen ürün adı"
  en "Visible product name in lists"
}

message Product.stock {
  tr "Geçerli bir stok adedi gir"
  en "Enter a valid stock count"
}
```

Current rules:

- Use one `i18n` block per project.
- The default locale must be included in `locales`.
- Top-level `label` blocks target stored fields and computed display fields with `Entity.field`.
- Top-level `label` blocks can also target generated UI copy with supported `app.*`, `page.*`, `action.*`, `table.*`, and `status.*` keys.
- Top-level `placeholder`, `help`, and `message` blocks target stored form fields with `Entity.field`.
- When more than one locale is declared, generated web UI includes a language selector.
- App chrome, navigation page names, table tools, status labels, CRUD buttons, custom action buttons, workflow transition buttons, and table/detail/form/filter/column labels re-render when the user changes locale.
- Generated form placeholders, form help text, and field-level frontend validation messages re-render when translated blocks exist.
- Generated table/detail `number`, `integer`, `decimal`, `money`, `date`, and `datetime` values use `Intl` formatting for the active locale.
- Generated App shell includes `lang` and deterministic `dir`; `ar`, `fa`, `he`, `ur`, and region variants are RTL.
- Runtime field text falls back to the default locale translation when the active locale is missing.
- If no translation exists, generated UI falls back to matching inline field metadata such as `label "Text"`, `placeholder "Text"`, `help "Text"`, or `message "Text"`.
- If no field label modifier exists, labels fall back to title-cased field name.

## Current Media Fields

Draft v0.2 supports deterministic local media fields for generated web apps:

```black
entity Product {
  photo image optional accept "image/*" label "Product Photo"
  specSheet file optional accept "application/pdf" label "Spec Sheet"
}
```

`image` and `file` are stored scalar field types. Generated forms render file inputs, read selected files as data URLs, and submit those strings through the normal JSON API path. Generated table/detail UI renders images as thumbnails and files as links. Generated OpenAPI schemas expose `x-blacklang-media` and `x-blacklang-encoding: data-url`.

`accept` is optional and only valid on `file` or `image` fields. `image` defaults to `accept "image/*"` when no accept value is declared. Storage providers, buckets, upload tokens, private endpoints, and secret values stay outside `.black` source.

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

Current generated Prisma output stores `number` and `integer` as signed integer values. Use `decimal` or `money` for fractional numeric values.

```black
entity Product {
  sku text required length 3..40 regex "^[A-Z0-9]+$" message "Use uppercase letters and numbers"
  stock number min 0
  website text optional url
}
```

The generated API validation uses the same field rules. `message "Text"` overrides the generated validation message for that field.

Top-level `message Entity.field { ... }` localizes generated frontend field-level validation messages for stored form fields. It does not localize entity-level `validate ... message "Text"` lines in the current draft.

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
  customer Customer required load detail query
}
```

The referenced type must be an existing entity.

When a relation field is used in a page form, the generated web app renders a select input and sends the matching foreign key to the API.

When a required relation field has no available related records, the generated form disables submit and tells the user which record must be created first.

When the related entity has its own generated page, the required relation guidance can navigate to that page.

When a relation field is used in a table or detail panel, the generated web app displays a readable label from the related record when available.

Relation fields may add `load list`, `load detail`, `load query`, `load mutation`, or `load none` to control which generated API response contexts attach the related object. Omitting `load` keeps the default attach behavior for list, detail, query, and mutation responses. `load none` returns only the generated relation ID field in responses. Generated list and query routes batch relation IDs into one target lookup per relation field, and permission-aware builds sanitize attached target records before returning nested relation data.

When a relation field is used in `table.search`, generated search uses that same readable relation label when the relation object is loaded, and falls back to the relation ID when it is not loaded.

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

## Current Entity Indexes

Draft v0.2 supports stored-field database index intent inside entity blocks:

```black
entity Order {
  customer Customer required
  status text default draft
  index customer, status
}
```

Current rule:

- Use `index field` for one stored field.
- Use `index fieldA, fieldB` for a composite stored-field index.
- Relation fields are allowed and map to generated relation ID columns such as `customerId`.
- Computed display fields cannot be indexed because they are not database columns.
- Generated Prisma schema emits deterministic `@@index` declarations.
- Generated SQLite setup emits deterministic `CREATE INDEX IF NOT EXISTS` statements.
- `black inspect app.black --affected Order.index --json` reports database schema/setup impact.

## Current Schema Migration Plan, Rename Runtime, And Runner

Draft v0.2 supports a read-only migration planning command:

```bash
black migrate plan old.black new.black --json
black migrate plan old.black new.black --ir
```

The command compares old and new `.black` sources after parse and validation. It reports generated database shape changes for entities, stored fields, relation fields, required/default/unique modifiers, entity indexes, and explicit rename declarations.

Result fields:

- `success` is false only when an input source cannot be read, parsed, or validated.
- `safe` is true only when every detected change has safe risk.
- `destructive` is true when a change can remove stored rows, remove stored column values, rename a generated column, or require type conversion.
- `changes[].risk` is `safe`, `manual`, or `destructive`.

The CLI plan does not connect to a database. First-class rename declarations in the new/current source are applied by generated `db:migrate`, `db:setup`, and `db:push` before the current schema is created or pushed:

```black
migration RenameProductName {
  rename entity ProductItem to Product
  rename field Product.title to name
}
```

The `to` entity or stored field must exist in the current source. The `from` entity or field must not remain in the current source. Renames are never inferred from similar names. Generated setup applies entity renames before field renames, records successful runs in `BlackMigration`, skips fresh DB renames when the old table does not exist, and fails when both old and new names exist.

Generated apps with migration blocks also write `src/migrate.ts`, `db:migrate:plan`, and `db:migrate`. `db:migrate:plan` is read-only and reports JSON with database reachability, ledger state, pending rename checks, `ready`, and redacted errors. `db:migrate` applies only declared rename migrations and records them in `BlackMigration`; run `db:setup` after it so the normal generated setup and seed path runs.

## Current Seed And Fixture Runtime

Use a top-level seed block to declare deterministic local/demo fixture rows for one entity:

```black
seed DemoProducts {
  source Product

  row DemoProductLow {
    sku "LOW-001"
    name "Low Stock Widget"
    stock 3
    price 19.99
  }
}

seed DemoOrders {
  source Order

  row DemoOrderDraft {
    customer ref DemoCustomerAcme
    total 125.50
  }
}
```

Each `row` key becomes the generated stable `id`, so rerunning `npm run db:setup` or `npm run db:seed` upserts the same declared rows. Scalar values use the same typed literal discipline as queries: quoted strings for text/email/date/datetime, finite numbers for numeric fields, and `true`/`false` for booleans. Relation fields use `ref OtherRowKey`; the referenced row must exist in seed rows for the related entity.

Seed can set stored fields only. Computed display fields, generated system fields, expressions, raw SQL, environment reads, and arbitrary code are outside the MVP. Validation catches missing required fields, unknown fields, relation ref mistakes, scalar type mismatches, basic field constraint failures, and duplicate seed values for unique fields. Do not put secrets or production credentials in seed literals.

When seeds exist, the web generator writes `src/seed.ts`, adds `db:seed`, and wires `db:setup` to run schema setup followed by deterministic seed upserts. Read `docs/seed.md` or `black docs seed --json`; use `black explain seed --json` and `black inspect --affected DemoProducts --json` before changing fixture data.

## Current Custom Queries

Use a top-level query and bind it to a page with the same source:

```black
query LowStockProducts {
  source Product
  where stock < 10
  aggregate lowStockCount count
  aggregate totalStock sum stock
  aggregate averagePrice avg price
  sort stock asc
  limit 50
}

page LowStock {
  source Product
  query LowStockProducts

  table {
    columns name, stock
  }
}
```

Conditions use stored primitive fields and typed literals; repeated `where` lines use AND. Text/email/date/datetime literals are quoted. `==` and `!=` work for supported stored fields; ordering comparisons also work for numeric/date/datetime fields. `aggregate` is optional and repeatable: `count` has no field, while `sum`, `avg`, `min`, and `max` require a stored numeric field. `sort` is optional, uses one stored field, and adds `id asc` for ties; the default order is `id asc`. `limit` is 1..1000, default 100.

The generated page reads `/api/<page>/query` and refetches after its mutations. When aggregates are declared, it also reads `/api/<page>/query/summary` and renders summary cards above the table. Existing lists and relation options remain available. Table search/filter/sort/pagination apply to the returned subset; summary values use the same source, row/archive policy, and `where` filters but ignore sort and limit. Query routes enforce existing auth/access/read permissions and reject unreadable condition/sort/aggregate fields with 403. Query filters are list-selection rules; entity row policies are the shared ownership/tenant authorization layer for query, detail, and mutation routes. Relations, computed/system fields, parameters, joins, raw SQL, and arbitrary code are outside the MVP.

Read `docs/query.md` or `black docs query --json`; use `black explain query --json` and `black inspect --affected LowStockProducts --json` for focused edits.

## Current Background Query Jobs

Use a top-level job to run a declared query in the generated worker:

```black
job LowStockMonitor {
  schedule every 15 minutes
  run query LowStockProducts
}
```

The MVP accepts one deterministic schedule and one read-only query run clause:

```text
schedule every <integer> minutes|hours|days
run query <QueryName>
```

Supported ranges are `1..1440 minutes`, `1..168 hours`, and `1..365 days`. The referenced query must exist and keeps its stored-field `where`, `sort`, and `limit` rules. Generated output writes `jobs/manifest.json`, `src/worker.ts`, `jobs:run`, `jobs:loop`, and OpenAPI root `x-blacklang-jobs` metadata. The worker selects record IDs and logs compact JSON metadata with the job name, schedule, query, source, count, limit, and timestamp.

Jobs appear in JSON, BlackIR, inspect, affected analysis, docs, explain, diagnostics, IDE metadata, generated worker output, package scripts, and OpenAPI metadata. Queue providers, retries, delayed queues, mutating jobs, external calls, and provider schedulers are future adapter work.

Read `docs/job.md` or `black docs job --json`; use `black explain job --json`, `black inspect --affected LowStockMonitor --json`, and generated `npm run jobs:run` for focused edits.

## Current Custom Actions

Use a top-level action and bind it from a page with the same source:

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

Custom actions are deterministic row-level mutations. Inputs use primitive field types and normal validation modifiers. `value` declares ordered local values, and Python-like indentation-based `if`/`else` branches can choose deterministic `set` statements. `set` writes stored primitive fields only and may use typed literals, action inputs, source fields, local values, and bounded numeric expressions with `+`, `-`, `*`, `/`, parentheses, and deterministic precedence. Condition comparisons use `==`, `!=`, `<`, `<=`, `>`, and `>=`; ordered comparisons currently require numeric values. Comparisons may be joined with `and`, `or`, `not`, and parentheses, using `not` before `and` before `or` precedence. Use a top-level `transaction Name { action ActionName }` block when row lookup, update, and audit logging should run inside one generated Prisma transaction. Computed fields, relation fields, entity policy fields, function calls, loops, external calls, and multi-row updates are outside the MVP.

A bound action generates a POST route, API client method, backend input validator, React row button and form panel, OpenAPI schemas, and audit logging when auth and roles exist. Page access, update permission, field-level update permission, optional action `allow` rules, and entity row policies are enforced. OpenAPI includes `x-blacklang-transaction` and `x-blacklang-transaction-name` when a top-level transaction block targets the action. Query-bound pages refetch after a custom action succeeds.

Read `docs/action.md`, `docs/transaction.md`, `black docs action --json`, or `black docs transaction --json`; use `black explain action --json`, `black explain transaction --json`, and `black inspect --affected RestockProduct --json` for focused edits.

## Current Transaction Blocks

Use a top-level transaction block to bind generated action/API update routes to one Prisma transaction boundary:

```black
transaction RestockAtomic {
  action RestockProduct
  api StockWebhook
}
```

The block does not create a new route and does not define mutation logic. It references existing custom actions and explicit API update handlers. Each target may appear in at most one transaction block, and API targets must have a bounded single-line `update Entity where ... set ...` handler or block-form `update Entity where ... { ... }` handler.

Generated route metadata includes `x-blacklang-transaction: true` and `x-blacklang-transaction-name` for targeted operations. Read `docs/transaction.md`, `black docs transaction --json`, `black explain transaction --json`, and `black inspect --affected RestockAtomic --json` for focused edits.

## Current Service Blocks

Use a top-level service block to group existing explicit API declarations into generated module metadata:

```black
service InventoryIntegration {
  api StockWebhook
}
```

The block does not create a new route and does not define request handling. It references existing top-level `api` declarations. Each API may appear in at most one service block so generated service metadata and `inspect --affected` stay unambiguous.

Generated output includes `services/manifest.json`, `src/services/<service>.ts`, OpenAPI top-level `tags`, OpenAPI top-level `x-blacklang-services`, route-level `x-blacklang-service`, and generated contract assertions. Read `docs/service.md`, `black docs service --json`, `black explain service --json`, and `black inspect --affected InventoryIntegration --json` for focused edits.

## Current Entity Row Policies

Use `policy owner` and `policy tenant` inside an entity to scope generated web routes by authenticated user context:

```black
entity Order {
  tenantId text required default "default"
  ownerId text required default "system"
  total money default 0
  status text default draft
  policy tenant tenantId
  policy owner ownerId
}
```

Current rule:

- The project must declare `auth` before using entity policies.
- Each policy references a stored `text required` field on the same entity.
- `owner` uses the authenticated user's generated `id`.
- `tenant` uses the authenticated user's `tenantId`; generated auth defaults new users to `"default"` and the generated Users page can update tenant IDs when tenant policies exist.
- Policy fields are stored database columns but are omitted from generated form fields, API input types, and OpenAPI input schemas.
- Generated create/update routes stamp policy fields from current user context.
- Generated list, query, detail, archive, restore, delete, workflow transition, and custom action routes apply the same row scope.
- Custom actions cannot set policy fields.

Read `docs/policy.md` or `black docs policy --json`; use `black explain policy --json` and `black inspect --affected policy --json` for focused edits.

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

Draft v0.2 generates an OpenAPI contract for web targets:

```text
generated/openapi.json
```

The contract is derived from page source entities, page actions, bound custom queries, background query job metadata, bound custom actions, transaction bindings, service bindings, and explicit API declarations. It describes generated REST paths, request bodies, response schemas, relation ID fields, action/API transaction metadata, service metadata, job metadata in `x-blacklang-jobs`, and field formats such as email.

The generated Express server also serves it at:

```text
/openapi.json
```

Draft v0.2 also supports explicit API declarations with generated declared-runtime routes:

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
    if incoming > 0
      set stock = stock + incoming
    else
      set stock = stock
  }
  respond accepted
  webhook
  public
}

transaction RestockAtomic {
  api StockWebhook
}

service InventoryIntegration {
  api StockWebhook
}
```

Explicit API declarations are generated declared-runtime routes in v0.2: they are parsed, validated, included in JSON/BlackIR/inspect/affected output, written to `generated/openapi.json`, mounted in the generated Express server, and probed by generated API smoke tests. The generated route validates declared path, query, and body parameters and returns deterministic JSON with `runtime: "declared"`. Webhook routes return `202 accepted`. Private routes are protected when `auth` exists.

Explicit API handlers support a bounded update MVP. `body` fields are supported for `POST`, `PUT`, and `PATCH` APIs and may include `file` and `image` values as data URL or absolute `http(s)` strings. Body fields may use `required`, `optional`, `min`, `max`, `length`, `regex`, `url`, and `message`; `accept` belongs to entity media form fields. `update Entity where field == value set field = value` updates one row selected by `id` or a stored `unique` field. The block form `update Entity where field == value { ... }` supports ordered local `value` declarations plus indentation-based `if`/`else` branches with `value`, `set`, and nested `if` statements. Set targets must be stored primitive non-policy fields. Handler values may reference `body.name`, `param.name`, source fields in set expressions, local values, typed string/number/boolean literals, and bounded numeric expressions with `+`, `-`, `*`, `/`, parentheses, and deterministic precedence. Condition comparisons use `==`, `!=`, `<`, `<=`, `>`, and `>=`; ordered comparisons currently require numeric values. Comparisons may be joined with `and`, `or`, `not`, and parentheses, using `not` before `and` before `or` precedence. A top-level transaction block can wrap bounded lookup, update, and private auth audit logging in one Prisma transaction. A top-level service block can group explicit APIs into generated service manifest/module metadata and OpenAPI service tags without changing request handling. Public handlers on tenant-scoped entities must include the tenant policy field as a required body field or path param. Public handlers cannot update owner-scoped entities.

## Current Generated Contract/API/Frontend/Browser-Check Smoke Tests

Draft v0.2 generated web and API-only apps include deterministic contract and API smoke tests. Web targets also include frontend smoke tests, explicit browser e2e checks, and browser e2e matrix plan/run commands:

```bash
npm test
npm run test:e2e
npm run test:e2e:plan
npm run test:e2e:matrix
```

The generated tests always include `src/blacklang.contract.test.ts` and `src/blacklang.api.test.ts`, and the generated `npm test` script starts with `npm run db:generate` so Prisma client output exists before API smoke tests import the server. `target web` also emits `src/blacklang.frontend.test.tsx`. When a web source declares `test` blocks, generated output also includes `src/blacklang.browser.test.tsx`, `src/blacklang.e2e.test.ts`, `src/blacklang.e2e.matrix.ts`, `tests/browser-matrix.json`, `test:e2e`, `test:e2e:plan`, `test:e2e:matrix`, and `test:all`. `target api` omits React, frontend smoke, browser-check, browser e2e, and browser matrix files; its `npm test` script stays contract/API-only after `db:generate`. These tests check `openapi.json` for entity schemas, page CRUD paths, bound query and query summary paths, bound custom action paths, explicit API contracts, declared ops paths, and action/API metadata including transaction flags. They also run generated validation functions for representative valid and invalid entity/action payloads, including integer-like `number` fields rejecting fractional values. The API smoke test starts the generated Express app on a random localhost port and checks `/openapi.json`, explicit API runtime routes, ops endpoints, anonymous API behavior, JSON 404 behavior when auth is absent, and generated CORS behavior when configured. When seeds exist, generated output also includes `src/seed.ts`, `db:seed`, and `db:setup` wiring; run generated `npm run db:setup` to prove fixture rows apply after schema setup. The frontend smoke test renders the generated React `App` with `react-dom/server`. Browser-check tests assert declared page, action, and generated text expectations in `npm test`; browser e2e launches an installed Chrome/Chromium/Edge executable through `playwright-core`, starts generated API/Vite servers, registers a generated auth user when auth is enabled, and checks declared text/page/action expectations in the real DOM. The generated matrix plan is read-only; matrix run executes the same generated e2e intent once for each available custom, Chrome, Edge, or Chromium target and reports compact JSON.

Use a top-level `test Name { page PageName; expect text "Generated text"; expect page OtherPage; expect action actionName }` block for deterministic generated browser-check, browser e2e, and browser matrix intent in `target web` projects. Read `docs/test.md`, `black docs test --json`, or `black explain test --json` before editing browser test declarations. `target api` reports `UNSUPPORTED_API_TARGET_TEST` for top-level test blocks.

## Current Benchmark Command

Draft v0.2 can measure BlackLang source size, generated output size, AI task scenarios, token estimates, evidence-backed AI eval history, web coverage, and tracked issue exports:

```bash
black benchmark --json
black benchmark app.black --out generated --ir
black benchmark tasks --json
black benchmark eval --json
black benchmark eval-history --json
black benchmark coverage --json
black benchmark issues --json
black ide --json
black ide diagnostics app.black --json
black ecosystem --json
```

The command validates the project, builds web output in a temporary directory, and reports source/generated file counts, line counts, byte counts, line ratios, and generated kind summaries. It does not mutate the configured generated output directory.

`black benchmark tasks --json` reports deterministic AI task scenarios and token estimates for common generated web changes. It uses current source/generated measurements and fixed scenario context sizes; it is a planning signal, not a billed-token measurement.

`black benchmark eval --json` reports a long-running AI eval corpus with repeat count, case prompts, expected evidence, required commands, scoring criteria, and token estimates. It is read-only metadata and does not call an AI model, mutate generated output, commit, push, deploy, or store secrets.

`black benchmark eval-history --json` reads `benchmarks/eval-history.blackdir` and reports evidence-backed eval result history with run summaries, validation/model sources, evidence paths, and pass/fail totals. Local validation entries are not billed model benchmark scores; external model-to-model scores should be appended only after repeated harness runs and release review.

`black benchmark coverage --json` reports the current weighted web coverage matrix, tracked issue status IDs, and percentage milestones. Use it before making broad claims about how complete the web target is.

`black benchmark issues --json` reports the same tracked issue IDs as a smaller export with total/open/done counts and the current open issue list.

## Current Security Defaults

Draft v0.2 generated API servers include baseline secure defaults:

- Disabled `X-Powered-By`
- Basic security headers
- `100kb` JSON body limit
- Simple IP-based rate limit

These are generated automatically. Future versions should add more explicit auth, role, permission, and policy syntax.

Draft v0.2 also supports explicit CORS intent:

```black
security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}
```

Generated Express servers read `CORS_ORIGINS` as a comma-separated list, reject unlisted browser origins, answer `OPTIONS` preflight requests, and add credential headers when `credentials true` is declared.

## Current Protected Source Commands

Draft v0.2 supports encrypted `.black.enc` source files for protected-source workflows:

```bash
black security encrypted-source --json
black security encrypt app.black --out app.black.enc --json
black security decrypt app.black.enc --stdout
```

The encrypted source format uses a plaintext header with `BLACKLANG-ENC v1`, `AES-256-GCM`, `SHA256-ENV`, the key environment variable name, and a base64 nonce. The source body is encrypted and base64 encoded below the header delimiter.

`black security decrypt` requires `--stdout` and never writes plaintext source by default. Parse, lint, validate, inspect, benchmark, security scan, and build read `.black.enc` source in memory when the header-declared key environment variable is set.

`black format` intentionally rejects `.black.enc` files. Format plaintext `.black` source in a trusted workspace, then re-encrypt it. Production packages exclude both `.black` and `.black.enc` files.

## Current Target Declaration

Draft v0.2 supports explicit top-level target declarations for generated web and API-only output:

```black
target web {
  frontend react
  backend node
  database sqlite
}
```

This block declares the intended generated application stack for the current `.black` source. The current web generator supports React frontend, Node backend, and SQLite, PostgreSQL, or MySQL database runtime.

Use PostgreSQL or MySQL with the same target block shape:

```black
target web {
  frontend react
  backend node
  database postgres
}
```

```black
target web {
  frontend react
  backend node
  database mysql
}
```

Use API-only output by omitting the frontend line and choosing `target api`:

```black
target api {
  backend node
  database sqlite
}
```

If the block is omitted, the generator keeps the legacy default: `web`, `react`, `node`, `sqlite`.

Rules:

- Use one `target` block per project.
- Current generator output supports `target web` and `target api`.
- `target web` requires `frontend react`, `backend node`, and `database sqlite`, `database postgres`, or `database mysql`.
- `target api` requires `backend node` and `database sqlite`, `database postgres`, or `database mysql`; do not add a `frontend` line.
- Generated rename migration runtime supports SQLite and PostgreSQL; `database mysql` reports `UNSUPPORTED_TARGET_DATABASE_MIGRATION` when migration blocks are present.
- Mobile, desktop, and alternate backend targets must wait until matching generator support exists.
- AI agents should use `black docs target --json` and `black inspect --affected target --json` before changing target metadata.

## Current Deployment Declaration

Draft v0.2 supports Docker deployment intent:

```black
deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  env CORS_ORIGINS optional
  preview local
  rollback keep 3
  cloud fly app env FLY_APP_NAME region env FLY_REGION
}
```

When `target docker` is declared, the generator writes `Dockerfile`, `.dockerignore`, and `docker-compose.yml`, adds `PORT` to `.env.example`, adds a `start` script to generated `package.json`, and makes the generated server read the configured port environment variable. `target web` serves the built Vite frontend from `dist`; `target api` returns JSON 404 for non-API routes.

When migration blocks exist, generated deployable output includes `src/migrate.ts`, `db:migrate:plan`, and `db:migrate`. Run the plan command in the target environment before mutating an existing database, then run apply and `db:setup`.

When `preview local` is declared, the generator writes `docker-compose.preview.yml`, adds `BLACKLANG_PREVIEW_PORT` to `.env.example`, and adds `deploy:preview` / `deploy:preview:down` package scripts. The preview stack uses an isolated host port and preview database defaults.

When `rollback keep COUNT` is declared, the generator writes `deploy/manifest.json`, `deploy/rollback.json`, `scripts/rollback-plan.mjs`, and a `deploy:rollback:plan` package script. The generated rollback plan script is read-only metadata inspection; it reports retained release directories and a rollback candidate without mutating infrastructure.

When `cloud PROVIDER app env NAME` is declared, the generator writes `deploy/cloud.json`, adds cloud metadata to `deploy/manifest.json`, writes `scripts/cloud-plan.mjs` and `scripts/cloud-exec.mjs`, and adds `deploy:cloud:plan`, `deploy:cloud:preflight`, and `deploy:cloud:exec` package scripts. The current provider values are `fly`, `render`, and `railway`. Plan and preflight are read-only; exec requires explicit apply mode and a provider CLI on PATH. Cloud app, region, token, and endpoint values should stay in environment variables or provider tooling rather than `.black` source.

SQLite targets keep their local database under `/app/data` in Docker Compose. PostgreSQL targets generate a PostgreSQL service, healthcheck, PrismaPg runtime adapter, and app `DATABASE_URL` fallback. MySQL targets generate a MySQL 8.4 service, healthcheck, PrismaMariaDb runtime adapter, `MYSQL_*` local env defaults, and app `DATABASE_URL` fallback.

## Current Ops Declaration

Draft v0.2 supports public runtime probes and structured request logs:

```black
ops {
  health path "/healthz"
  readiness path "/readyz"
  metrics path "/metrics"
  logging requests
  observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT
}
```

The web generator emits root-level health, readiness, and metrics routes before API auth and CSRF middleware. OpenAPI includes `x-blacklang-ops` metadata, generated tests probe the declared endpoints, and Docker deploy output uses the health endpoint for app service healthchecks.

When `observe webhook endpoint env NAME` is declared, the generated server sends non-blocking structured request events to the configured endpoint only when that environment variable is set. When `observe otlp endpoint env NAME` is declared, it sends non-blocking OTLP HTTP JSON trace payloads instead. Observe middleware creates or propagates W3C `traceparent` context, adds the response `traceparent` header, writes `ops/observability.json`, includes `x-blacklang-observability` metadata, and generated API smoke tests verify local hook/exporter delivery without calling an external service.

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

Generated auth stores a primary `role` plus a `roles` set for each user, assigns the first declared role to newly registered users, returns both values from `/api/auth/me`, and blocks page API routes unless at least one current role is allowed.

When roles exist, generated apps include a basic Users page where the first declared role can list users and assign one or more roles. When tenant policies exist, the same page can update user tenant IDs.

Generated API routes also use role permission actions:

- `read` for list/detail
- `create` for create
- `update` for edit/archive/restore
- `delete` for delete/bulk delete

Generated React pages hide action controls when none of the current user roles is allowed to use them.

Field names after a permission resource scope the permission to those fields. For example, `deny read Product price` lets matching roles read Product records while hiding the `price` field from generated API responses and React views.

Field-level mutation is also enforced. For example, `allow update Product stock` lets the role update only `stock`; other submitted Product fields are ignored by generated API routes.

Generated apps also create an audit log table when auth and roles exist. Create, update, archive, restore, delete, bulk delete, register, and role update operations write audit records. The first declared role can open the generated Audit page to review recent activity.

Generated cookie auth uses an HttpOnly session cookie plus a readable CSRF cookie. Generated write requests send the CSRF token in `X-CSRF-Token`, and the API rejects authenticated state-changing requests when the cookie and header do not match.

Permission checks evaluate the full assigned role set: an explicit deny from any assigned role wins, otherwise an allow from any assigned role grants access. The primary `role` value remains for backward-compatible UI/API/audit surfaces. Secret manager provider execution comes later.

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

A page `view` can place a declared component as an inline reusable section:

```black
view {
  order table, StockSummary, StockCards, detail, form
  section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
  section StockCards component StockBadge bind each span 1 title "Stock Cards"
}
```

For this MVP, component sections bind to the selected source record, the first loaded list record, or each loaded list/query record. Component inputs must be scalar primitive inputs that match stored or computed fields on the page source entity by name and type. Generated JSX/DOM follows the effective `view.order`; CSS `order` rules are still emitted as stable metadata and a layout backstop.
