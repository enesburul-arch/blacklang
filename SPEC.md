# BlackLang Specification

Version: draft v0.1

## Purpose

BlackLang is an AI-native deterministic intent language.

BlackLang is designed for AI coding agents. Its source files should be easy to read, easy to modify, deterministic, and compact without becoming cryptic.

## File Extension

BlackLang source files use the `.black` extension.

Protected BlackLang source files use the `.black.enc` extension.

BlackLang UI theme/profile files use the `.blackthm` extension.

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
- AI agents must not invent unsupported BlackLang syntax and present it as official behavior.

## Current Parser Model

Draft v0.1 uses a lexer-backed token stream before AST parsing.

The lexer recognizes identifiers, quoted strings, `{`, `}`, `(`, `)`, commas, comparison operators, newlines, and comments. Comments beginning with `#` or `//` are ignored only outside quoted strings. Inline braces are split into deterministic statements before parsing, and unclosed strings report `UNCLOSED_STRING`.

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

## Current AI Agent Contract

The official BlackLang path is:

```text
.black source -> black CLI -> generated web/API output
```

AI agents should treat `docs/ai-agent-contract.md` as the current capability boundary.

Current BlackLang does not yet provide:

```text
<script type="text/black">
browser-side BlackLang interpreter
calculator/game expression runtime
custom frontend event handlers
arbitrary custom button behavior outside declared page actions
arbitrary E2E step execution beyond generated `test` declarations
general-purpose programming
```

When a requested task requires unsupported behavior, the agent should state the limitation and either build a clearly labeled normal web prototype or add the missing BlackLang compiler feature first.

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
    "pages": 4
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

## Current Accessibility Audit Command

Draft v0.2 supports read-only generated UI accessibility policy checks:

```bash
black audit accessibility --json
black audit accessibility app.black --ir
```

When the file argument is omitted, the CLI reads `blacklang.toml` and audits the configured `source` path.

The command parses and validates source first. If source is invalid, diagnostics are returned in `errors`. Valid source may return policy `findings` such as `ACCESSIBILITY_MISSING_OVERLAY_TITLE` or `ACCESSIBILITY_MISSING_GROUP_TITLE`. The command never writes `.black` source or generated files.

## Current Docs All Command

Draft v0.2 supports a complete compact documentation export:

```bash
black docs --all --json
black ide --json
black ide diagnostics app.black --json
black ecosystem --json
```

JSON shape:

```json
{
  "success": true,
  "command": "docs",
  "version": "0.1.0-dev",
  "count": 68,
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

## Current IDE Command

BlackLang IDE support is exposed through compiler-owned CLI output:

```bash
black ide --json
black ide --ir
black ide diagnostics app.black --json
black ide diagnostics app.black --ir
```

`black ide --json` returns language metadata, source extensions, completion items, snippets, and a diagnostic code catalog. `black ide diagnostics <file> --json` reads one source file and returns editor-friendly diagnostics for formatting, parse, validation, and source-security findings. Diagnostic ranges are zero-based and end-exclusive. A readable but invalid source returns `success: true` and `valid: false`; command-level failures such as unreadable files return `success: false`.

The packageable VS Code bridge lives under `editors/vscode-blacklang`. It contributes `.black` and `.blackthm` language ids, TextMate grammars, compiler-backed completion/snippet providers, compiler-backed diagnostics, a `FORMAT_REQUIRED` quick fix, and an affected-symbol inspection command backed by `black inspect --affected`.

## Current Ecosystem Command

BlackLang ecosystem discovery is exposed through:

```bash
black ecosystem --json
black ecosystem --ir
```

The JSON result includes release manifest/checksum/signature trust metadata, release transparency log policy, key rotation policy, release scripts, packageable source packages, prepared package registry manifests, built-in target/deploy/observability/editor adapters, prepared provider adapter marketplace manifests, and extension trust workflow checks. This command is read-only. It exists so AI agents and package tooling can discover install/release/editor/adapter boundaries without guessing from repository layout.

Release trust is exposed through:

```bash
black docs release-trust --json
node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict
```

The verifier is read-only. It checks `release.blackdir`, `checksums.sha256`, archive bytes, detached Ed25519 `<artifact>.sig` files, and a trusted public key from `BLACKLANG_RELEASE_PUBLIC_KEY` or `BLACKLANG_RELEASE_PUBLIC_KEY_FILE`. Companion registry policy metadata defines append-only release transparency entries and key rotation requirements. Private signing keys must never be stored in `.black` source, generated output, manifests, registry metadata, adapter metadata, or package metadata.

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

Provider-specific behavior must remain in extensions or adapters until the parser, validator, docs, JSON/BlackIR, diagnostics, affected analysis, tests, compact AI learning docs, manifest metadata, and trust checks are ready for that syntax to become official or before external systems are mutated.

## Current Computed Fields

Draft v0.2 supports read-only computed display fields inside entity blocks:

```black
entity Product {
  stock number default 0
  price money default 0
  tax money default 0
  computed inventoryValue money = stock * (price + tax) label "Inventory Value"
}
```

Computed fields are generated as React display helpers for table columns and detail views.

They are not database columns, not API input fields, and not generated form inputs.

Current expression support is bounded to stored number-like fields, numeric literals, parentheses, and deterministic arithmetic precedence:

```text
<numericExpression> ::= field | number | (expression) | expression + expression | expression - expression | expression * expression | expression / expression
```

If a computed field references a source field hidden by field-level permissions, generated UI hides the computed value too.

## Current Media Fields

Draft v0.2 supports local file and image fields inside entity blocks:

```black
entity Product {
  photo image optional accept "image/*" label "Product Photo"
  specSheet file optional accept "application/pdf" label "Spec Sheet"
}
```

`file` and `image` are stored scalar field types. Generated Prisma schemas store them as strings, and generated SQLite setup uses text columns. Generated React forms render file inputs; selected files are read as data URLs and submitted through the normal JSON API path. Generated table/detail UI renders `image` values as thumbnails and `file` values as links.

`accept "..."` is optional and only valid on `file` or `image` fields. `image` defaults to `accept "image/*"` in generated inputs when no accept value is declared. Image accept values must include `image/`.

Generated frontend and backend validation accepts media values that are data URLs or absolute `http(s)` URLs. `image` values must start with `data:image/` when using data URLs. Generated OpenAPI schemas include `x-blacklang-media` and `x-blacklang-encoding: data-url`. Generated Express JSON body size is `2mb` when any entity has a media field.

Storage provider adapters, buckets, upload credentials, private endpoints, image processing, and download permission policies are outside this MVP and must not be represented as secrets in `.black` source.

## Current Entity Indexes

Draft v0.2 supports stored-field database index intent inside entity blocks:

```black
entity Order {
  customer Customer required
  status text default draft
  index customer, status
}
```

Rules:

- Use `index field` for one stored field.
- Use `index fieldA, fieldB` for a composite stored-field index.
- Index fields must be stored fields declared on the same entity.
- Relation fields are allowed and map to generated relation ID columns such as `customerId`.
- Computed display fields cannot be indexed because they are not database columns.
- One index may contain at most 4 fields.
- Generated Prisma schema emits `@@index([...], map: "<Entity>_<field...>_idx")`.
- Generated SQLite setup emits matching `CREATE INDEX IF NOT EXISTS` statements.
- `black inspect app.black --affected Entity.index --json` reports database schema/setup impact.

## Current Schema Migration Plan, Rename Runtime, And Runner

Draft v0.2 supports a read-only schema planning command:

```bash
black migrate plan old.black new.black --json
black migrate plan old.black new.black --ir
```

The command compares two validated `.black` sources and reports generated database shape changes. It does not connect to a database, apply SQL, or mutate files. First-class migration rename declarations in the new/current source are applied by generated `db:migrate`, `db:setup`, and `db:push` before the current database schema is created or pushed.

Rename syntax:

```black
migration RenameProductName {
  rename entity ProductItem to Product
  rename field Product.title to name
}
```

Rules:

- `rename entity OldEntity to NewEntity` declares that the current `NewEntity` continues the old generated table.
- `rename field Entity.oldField to newField` declares that the current stored `newField` continues the old generated column on current `Entity`.
- The `to` entity or stored field must exist in the current source.
- The `from` entity or field must not remain in the current source.
- Rename intent is never inferred from similar names.
- Generated setup applies entity renames before field renames inside a migration.
- Fresh databases skip rename operations when the old table does not exist and then create the current schema normally.
- Existing databases fail loudly if both old and new table/column names exist.

JSON result shape:

```json
{
  "success": true,
  "command": "migrate plan",
  "version": "0.1.0-dev",
  "oldFile": "old.black",
  "newFile": "new.black",
  "safe": false,
  "destructive": true,
  "summary": {
    "entitiesAdded": 0,
    "entitiesRemoved": 0,
    "fieldsAdded": 1,
    "fieldsRemoved": 1,
    "fieldChanges": 1,
    "indexesAdded": 1,
    "indexesRemoved": 0,
    "renames": 0,
    "safeChanges": 2,
    "manualChanges": 1,
    "destructiveChanges": 1
  },
  "changes": [],
  "steps": [],
  "generatedFiles": [],
  "agentNotes": [],
  "errors": []
}
```

Risk rules:

- `safe`: adding a table, adding an optional/defaulted column, relaxing nullability, adding/removing an index, adding/changing/removing a default, removing a unique constraint, or applying an explicit entity/field rename.
- `manual`: adding a required column without a default, adding a unique constraint, or changing a relation target.
- `destructive`: dropping a table, dropping a column, changing a generated column name, or changing a generated database/source type.

`success` is false only when either input file cannot be read, parsed, or validated. A valid plan with `safe: false` is still a successful command result and should be reviewed before deployment.

Build output for migration blocks includes `migrations/manifest.json`, one deterministic `migrations/*.sql` preview file per migration declaration, a generated `src/migrate.ts` online migration runner, `db:migrate:plan` and `db:migrate` package scripts, a `BlackMigration` ledger model in `prisma/schema.prisma`, and generated setup code in `src/setup-db.ts`.

Generated `db:migrate:plan` is read-only and reports JSON with database reachability, per-migration ledger state, pending rename checks, `ready`, `summary`, `applied`, `skipped`, and redacted `errors`. SQLite plan mode does not create a missing database file. PostgreSQL plan mode reports unreachable databases without printing the connection string. Generated `db:migrate` applies only declared rename migrations and records successful runs in `BlackMigration`; deploy flows should run `db:setup` after apply mode so the normal generated setup and seed path runs.

Missing old/new file arguments return `MISSING_SCHEMA_MIGRATION_FILES`. Invalid rename declarations return stable diagnostics such as `INVALID_MIGRATION_RENAME`, `UNKNOWN_MIGRATION_RENAME_TARGET`, `MIGRATION_RENAME_SOURCE_STILL_EXISTS`, or `DUPLICATE_MIGRATION_RENAME`.

## Current Seed And Fixture Runtime

Draft v0.2 supports top-level deterministic seed declarations for local/demo fixture rows:

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

Grammar:

```text
seed <Name> {
  source <Entity>
  row <RowKey> {
    <storedField> <typedLiteral>
    <relationField> ref <OtherRowKey>
  }
}
```

Seed names use PascalCase letters and digits and must remain unique after lowercasing, including against other top-level symbols. Each seed block targets exactly one source entity. Row keys use letters, digits, and underscore, start with a letter, and are unique per generated entity because they become stable generated `id` values.

Seed values target stored fields only. Computed display fields, generated system fields, raw SQL, expressions, environment reads, and arbitrary code are unsupported. Scalar values reuse typed literal rules: quoted strings for `text`, `email`, `date`, `datetime`, `file`, and `image`, finite unquoted numbers for numeric fields, and unquoted `true` or `false` for booleans. Seeded `file` and `image` values must be data URLs or absolute `http(s)` URLs; image data URLs must start with `data:image/`. Relation fields must use `ref RowKey`; the referenced row key must be declared for the related entity.

Validation checks required stored fields without defaults, duplicate row fields, unknown fields, computed-field writes, scalar type mismatches, relation refs, basic field constraints such as `min`, `max`, `length`, `regex`, `url`, and media data URL shape, and duplicate fixture values for fields marked `unique`. Secrets, passwords, tokens, API keys, and production credentials must not be written as seed literals.

Generated web output emits `src/seed.ts` and adds `db:seed` to `package.json` when seeds are declared. `npm run db:setup` runs schema setup first and then `npm run db:seed`. The seed runtime uses parameterized inserts with `ON CONFLICT(id) DO UPDATE`, making repeated local setup deterministic and idempotent for declared rows.

Seeds appear in JSON, BlackIR, inspect, affected analysis, benchmark generated file kinds, generated `package.json`, and generated database setup behavior. Use `black docs seed --json`, `black explain seed --json`, and `black inspect --affected DemoProducts --json` for focused agent context. The detailed reference is `docs/seed.md`.

## Current Browser Test Declarations

Draft v0.2 supports top-level deterministic browser-check and browser e2e declarations for generated `target web` output:

```black
test WarehouseBrowserSmoke {
  page Products
  expect text "Depo"
  expect page LowStock
  expect action RestockProduct
}
```

Grammar:

```text
test <Name> {
  page <PageName>
  expect text "<generated UI text>"
  expect page <PageName>
  expect action <actionName>
}
```

Test names use PascalCase letters and digits and must remain unique after lowercasing, including against other top-level symbols. Each test targets exactly one page and must declare at least one expectation.

Supported expectations are intentionally bounded:

- `expect text "..."` checks generated React server-rendered HTML or the generated browser-check text catalog in `npm test`, and visible DOM text in `npm run test:e2e` or `npm run test:e2e:matrix`.
- `expect page PageName` checks generated page metadata in `npm test`, and generated navigation/page labels in `npm run test:e2e` or `npm run test:e2e:matrix`.
- `expect action actionName` checks actions exposed by the target page in generated metadata and checks visible action buttons in browser e2e and browser matrix runs.

When tests exist in a `target web` source, generated output emits `src/blacklang.browser.test.tsx` and appends it to `npm test` after the contract, API, and frontend smoke tests. It also emits `src/blacklang.e2e.test.ts`, `src/blacklang.e2e.matrix.ts`, `tests/browser-matrix.json`, and `test:e2e`, `test:e2e:plan`, `test:e2e:matrix`, and `test:all` package scripts. The e2e script launches an installed Chrome, Chromium, or Edge executable through `playwright-core`, creates an ephemeral SQLite database by default, runs setup/seed modules, registers a generated auth user when auth is enabled, and verifies declared expectations in the real DOM. The matrix plan is read-only and reports custom, Chrome, Edge, and Chromium target availability; matrix run executes the same e2e intent once per available target and reports compact JSON. `target api` has no generated browser runtime; top-level `test` declarations under `target api` report `UNSUPPORTED_API_TARGET_TEST`. Arbitrary frontend event logic and arbitrary E2E step execution remain outside the language.

Tests appear in JSON, BlackIR, inspect, affected analysis, docs, explain, diagnostics, generated package scripts, and Warehouse example output. Use `black docs test --json`, `black explain test --json`, and `black inspect --affected WarehouseBrowserSmoke --json` for focused agent context. The detailed reference is `docs/test.md`.

## Current Custom Queries

Draft v0.2 supports top-level custom queries bound to stored entities and consumed by pages:

```black
query LowStockProducts {
  source Product
  where stock < 10
  sort stock asc
  limit 50
}

page LowStock {
  source Product
  query LowStockProducts

  table {
    columns sku, name, stock
    search sku, name
    paginate 10
  }
}
```

Grammar:

```text
query <Name> {
  source <Entity>
  where <storedField> <operator> <typedLiteral>
  aggregate <name> count
  aggregate <name> <sum|avg|min|max> <numericField>
  sort <storedField> <asc|desc>
  limit <integer>
}
```

Query names use PascalCase letters and digits and must remain unique after lowercasing, including against other top-level symbols. The `source` line is required exactly once. `where` may be omitted or repeated; multiple distinct conditions use AND and identical repeats are rejected. `aggregate` may be omitted or repeated; aggregate names must be unique inside the query. `sort` and `limit` are optional and may each occur at most once. The only source keyword is `source`. Page binding is one `query <Name>` line; the query must exist and its source must equal the page source. Multiple pages may reuse the same query.

Allowed fields are stored `text`, `email`, `file`, `image`, `number`, `integer`, `decimal`, `money`, `boolean`, `date`, and `datetime` fields on the source entity. Relation, computed, and generated system fields cannot be used in conditions, declared sort, or aggregates. Conditions compare a field with a literal, never another field.

Literal and operator rules:

- Text and email use quoted strings. Dates use quoted valid `YYYY-MM-DD` calendar dates; datetimes use quoted RFC3339 timestamps such as `"2026-09-06T12:00:00Z"`.
- Numeric literals use finite, unquoted decimal notation without exponent or hexadecimal syntax. `number` and `integer` require signed 32-bit whole-number literals, matching the current stored integer runtime; `decimal` and `money` also allow decimal fractions. Boolean literals are unquoted `true` or `false`.
- `==` and `!=` are supported for every allowed field type.
- `<`, `<=`, `>`, and `>=` require numeric, date, or datetime fields.
- `aggregate name count` counts rows without a field. `sum`, `avg`, `min`, and `max` require one stored `number`, `integer`, `decimal`, or `money` field.
- Null literals, parameters, OR, relation paths, joins, and arbitrary SQL or code are unsupported.

Runtime rules:

- Query filters execute in the server before ordering and limiting; repeated conditions form an AND expression.
- A query may sort by one allowed field, ascending or descending. Generated `id asc` is the stable tie breaker. Without a sort declaration, order is `id asc`.
- The limit range is 1..1000 inclusive; the default is 100.
- Existing archived-record visibility applies. String comparison and null ordering follow the current SQLite runtime.
- Bound pages use a generated `GET /api/<lowercase-page-name>/query` route and `queryList` client method. Queries with aggregates also generate `GET /api/<lowercase-page-name>/query/summary`, a `querySummary` client method, and summary cards above the page table. Unused queries create no standalone endpoint.
- Aggregate summaries use the same source, row policy, archive mode, and `where` filters as the list route. Sort and limit are not applied to summary values.
- Existing entity list and CRUD routes retain their behavior. Relation selectors use the baseline entity list.
- Page search, table filters, optional `table.sort`, and pagination run on the returned subset. Without `table.sort`, query order is preserved. Successful mutations on the bound page trigger a query refetch.
- Query routes retain authentication, page access, and entity read checks. Every condition/sort/aggregate field must also be readable by the current role; otherwise the route returns 403. Response field hiding still applies.
- Query predicates select lists; they do not declare ownership or tenant authorization. Entity row policies, when declared, are applied separately to generated list, query, detail, and mutation routes.

Queries and page bindings appear in JSON, BlackIR, inspect, affected analysis, and generated OpenAPI output. Queries may also be consumed by background jobs. Use `black docs query --json`, `black explain query --json`, and `black inspect --affected LowStockProducts --json` for focused agent context. The detailed reference is `docs/query.md`.

## Current Background Query Jobs

Draft v0.2 supports top-level background jobs that run declared queries in a generated worker:

```black
query LowStockProducts {
  source Product
  where stock < 10
  sort stock asc
  limit 50
}

job LowStockMonitor {
  schedule every 15 minutes
  run query LowStockProducts
}
```

Grammar:

```text
job <Name> {
  schedule every <integer> minutes|hours|days
  run query <QueryName>
}
```

Job names use PascalCase letters and digits and must remain unique after lowercasing, including against other top-level symbols and query names. A job must contain exactly one schedule line and exactly one run line. The only schedule kind is `every`; supported units are `minutes`, `hours`, and `days`. Supported intervals are `1..1440 minutes`, `1..168 hours`, and `1..365 days`.

The only MVP run mode is `run query QueryName`. The referenced query must exist. The worker applies that query's stored-field `where`, deterministic `sort`, and `limit` rules, selects only record IDs, and logs compact JSON metadata with job name, schedule, query, source, count, limit, and timestamp.

Generated output:

- `jobs/manifest.json` lists job names, schedules, interval milliseconds, run kind, query, source, limit, and safe execution notes.
- `src/worker.ts` exports `blackJobs`, `runJob(name)`, and `runAllJobsOnce()`, supports `--once`, `--loop`, and optional `--job Name`, and disconnects Prisma after one-shot runs.
- `package.json` includes `jobs:run` for one execution and `jobs:loop` for a long-running worker process.
- `generated/openapi.json` includes root-level `x-blacklang-jobs` metadata.

Jobs run in internal worker scope and do not expose public HTTP endpoints. Queue providers, retries, delayed queues, mutating jobs, external calls, provider scheduler configuration, cron expression syntax, and arbitrary code are outside the MVP.

Jobs appear in JSON, BlackIR, inspect, affected analysis, docs, explain, diagnostics, IDE metadata, generated worker output, package scripts, and OpenAPI metadata. Use `black docs job --json`, `black explain job --json`, `black inspect --affected LowStockMonitor --json`, and generated `npm run jobs:run` for focused agent context. The detailed reference is `docs/job.md`.

## Current Custom Actions

Draft v0.2 supports top-level custom actions bound to stored entities and exposed through page `actions` lists:

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

Grammar:

```text
action <Name> {
  source <Entity>
  input <name> <primitiveFieldType> <modifiers...>
  value <name> = <value|numericExpression>
  if <condition>
    set <storedField> = <value|numericExpression>
  else
    set <storedField> = <value|numericExpression>
  set <storedField> = <value>
  set <storedField> = <left> <+|-|*|/> <right>
  allow <RoleName...|authenticated>
  success "Message"
}
```

Action names use PascalCase letters and digits and must remain unique after lowercasing, including against other top-level symbols and query names. The `source` line is required exactly once. An action must contain at least one reachable `set` statement. Inputs, local `value` declarations, `if`/`else` branches, `allow`, and `success` are optional. A page exposes a custom action by listing the action name in `actions`; the page source and action source must match. Unbound actions create no runtime endpoints. Atomic route behavior is declared with a separate top-level transaction block.

Input types are the primitive field types `text`, `email`, `number`, `integer`, `decimal`, `money`, `boolean`, `date`, and `datetime`. Custom action inputs do not support `file` or `image` in this MVP. `number` and `integer` validate as signed 32-bit whole numbers in the current generated Prisma runtime; `decimal` and `money` allow finite fractional values. Supported input modifiers are `required`, `optional`, `default`, `label`, `placeholder`, `help`, `min`, `max`, `length`, `regex`, `url`, and `message`. Input names cannot match stored or computed fields on the source entity.

Assignments may write stored primitive fields only. Relation fields, computed display fields, and entity policy fields cannot be set. Values may be typed literals, action inputs, stored primitive source fields, ordered local values, or bounded numeric expressions with `+`, `-`, `*`, `/`, parentheses, and deterministic precedence for number-like target and operand types. `value` declarations are evaluated in order and may be referenced by later `value`, `set`, and `if` statements. `if` uses indentation-based then/else branches; `else` must align with its matching `if`, and branch statements may be `value`, `set`, or nested `if`. Condition comparisons use `==`, `!=`, `<`, `<=`, `>`, and `>=`; ordered comparisons currently require numeric values. Comparisons may be joined with `and`, `or`, `not`, and parentheses, using `not` before `and` before `or` precedence. Literal division by zero is rejected by validation; dynamic division by zero is guarded by the generated route. Joins, aggregates, raw SQL, function calls, loops, external calls, and multi-row updates are outside this MVP.

Runtime rules:

- A page-bound custom action generates `POST /api/<lowercase-page-name>/:id/actions/<lowercase-action-name>`.
- The generated API client exposes `run<ActionName>(id, input)`.
- The generated React page renders a row button and a form panel for action inputs.
- Generated backend validation checks the action input before mutation.
- A top-level `transaction Name { action ActionName }` block makes the generated route run source row lookup, update, and auth audit write inside one Prisma transaction. Validation, auth, page access, and input parsing still run before the transaction.
- Authenticated role projects require page access, update permission on the source entity, field-level update permission for each written field, any optional action `allow` rule, and any entity row policy.
- Query-bound pages refetch their server query after a custom action succeeds.
- Generated OpenAPI output includes action request and response schemas plus `x-blacklang-action`, `x-blacklang-source`, `x-blacklang-mutates-fields`, and optional `x-blacklang-transaction`/`x-blacklang-transaction-name` metadata.

Custom actions appear in JSON, BlackIR, inspect, affected analysis, generated validation files, route files, API clients, React pages, OpenAPI output, and audit logs when auth and roles exist. Use `black docs action --json`, `black docs transaction --json`, `black explain action --json`, and `black inspect --affected RestockProduct --json` for focused agent context. The detailed references are `docs/action.md` and `docs/transaction.md`.

## Current Transaction Blocks

Draft v0.2 supports top-level transaction blocks for generated action/API update route boundaries:

```black
transaction RestockAtomic {
  action RestockProduct
  api StockWebhook
}
```

Grammar:

```text
transaction <Name> {
  action <ActionName>
  api <APIName>
}
```

Transaction names use PascalCase and must remain unique after lowercasing. A transaction block must target at least one custom action or explicit API update handler. Each action/API target can be bound to at most one transaction block across the project. API targets must have an `update Entity where ... set ...` handler or an `update Entity where ... { ... }` block.

Transaction blocks do not create endpoints and do not declare mutation logic. They change generated runtime boundaries and OpenAPI metadata for targeted routes. Targeted operations include `x-blacklang-transaction: true` and `x-blacklang-transaction-name`. Use `black docs transaction --json`, `black explain transaction --json`, and `black inspect --affected RestockAtomic --json` for focused agent context. The detailed reference is `docs/transaction.md`.

## Current Service Blocks

Draft v0.2 supports top-level service blocks for grouping existing explicit APIs into generated service module metadata:

```black
service InventoryIntegration {
  api StockWebhook
}
```

Grammar:

```text
service <Name> {
  api <APIName>
}
```

Service names use PascalCase and must remain unique after lowercasing. A service block must target at least one explicit API declaration. Each API target can be bound to at most one service block across the project.

Service blocks do not create endpoints and do not declare request handlers. They change generated service discovery metadata for existing explicit APIs. Generated output includes `services/manifest.json`, `src/services/<service>.ts`, OpenAPI top-level `tags`, OpenAPI top-level `x-blacklang-services`, and route-level `x-blacklang-service`. Use `black docs service --json`, `black explain service --json`, and `black inspect --affected InventoryIntegration --json` for focused agent context. The detailed reference is `docs/service.md`.

## Current Entity Row Policies

Draft v0.2 supports owner and tenant row policy intent inside entity blocks:

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

Grammar:

```text
policy owner <storedTextRequiredField>
policy tenant <storedTextRequiredField>
```

Policy lines require an `auth` block. Each entity may declare at most one `owner` policy and one `tenant` policy. The referenced field must be a stored `text required` field on the same entity and must not be `unique`. Computed fields, relation fields, generated system fields, missing fields, and non-text fields are rejected.

`policy owner ownerId` binds rows to the authenticated user's generated auth user id. `policy tenant tenantId` binds rows to the authenticated user's tenant id; generated auth creates a `tenantId` field on `BlackUser`, defaults new users to `"default"`, and lets the first declared role update user tenant IDs from the generated Users page.

Policy fields are normal stored database columns, but generated forms, API client input types, and OpenAPI input schemas omit them. Generated create and update routes stamp policy values from current user context. Generated list, custom query, detail, archive, restore, delete, workflow transition, and custom action routes apply policy filters before returning or mutating rows. Relation response attachment also applies the related target entity row policy when the target declares one. Custom action `set` clauses cannot write policy fields.

Policies appear in JSON, BlackIR, inspect summaries, affected analysis, generated auth routes, API routes, API clients, OpenAPI schemas, database setup, and Prisma schema output. Use `black docs policy --json`, `black explain policy --json`, and `black inspect --affected policy --json` for focused agent context. The detailed reference is `docs/policy.md`.

## Current Agent Startup Command

Draft v0.2 supports a deterministic project entry checklist:

```bash
black agent startup --json
black agent startup app.black --json
black agent startup --ir
```

When the file argument is omitted, the CLI reads `blacklang.toml` and uses the configured `source` and `out` paths.

JSON shape:

```json
{
  "success": true,
  "command": "agent startup",
  "version": "0.1.0-dev",
  "config": {
    "languageVersion": "0.1",
    "target": "web",
    "source": "examples/warehouse/app.black",
    "out": "generated",
    "theme": "examples/warehouse/theme.blackthm"
  },
  "summary": {
    "app": "Warehouse",
    "entities": 3,
    "pages": 4
  },
  "readFirst": [
    {
      "path": "AGENTS.md",
      "purpose": "Local rules for AI agents working in this project.",
      "exists": true
    }
  ],
  "sourceFiles": ["examples/warehouse/app.black"],
  "themeFiles": ["examples/warehouse/theme.blackthm"],
  "generatedDirs": ["generated"],
  "checklist": [],
  "commands": [],
  "policies": [],
  "errors": []
}
```

If the project currently has parse, validation, or file-read diagnostics, the command still returns the startup checklist but sets `success` to `false` and fills `errors`.

AI agents should run this before editing an unfamiliar project so they can learn local files, generated output boundaries, validation commands, and source-security rules from one stable compiler output.

## Current Theme Profile Format

Draft v0.2 supports separate `.blackthm` files for UI theme and compact UI profile metadata.

Theme files are source assets. They are not generated CSS. They exist so the compiler, generator, and AI agents can inspect UI token names and positional mode slots before compact inline UI values are mapped to CSS.

Projects can point to a theme file from `blacklang.toml`:

```toml
theme = "examples/warehouse/theme.blackthm"
```

Current syntax:

```blackthm
blackthm WarehouseTheme {
  version 1
  target web
  locked true

  token color primary "#2563eb"
  token color surface "#ffffff"
  token space sm 8
  token radius md 6

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

CLI:

```bash
black theme inspect --json
black theme inspect examples/warehouse/theme.blackthm --json
black theme inspect examples/warehouse/theme.blackthm --ir
black theme migrate old.blackthm new.blackthm --json
black theme migrate old.blackthm new.blackthm --ir
```

JSON shape:

```json
{
  "success": true,
  "command": "theme inspect",
  "version": "0.1.0-dev",
  "file": "examples/warehouse/theme.blackthm",
  "theme": {
    "name": "WarehouseTheme",
    "version": 1,
    "target": "web",
    "locked": true,
    "tokens": [],
    "profile": {
      "name": "UICompact",
      "version": 1,
      "rules": {
        "inlineSyntax": "ui <mode> <values...> [| <mode> <values...>...]",
        "slotOrder": "left-to-right",
        "modeSeparator": "|",
        "missingTrailingSlots": "default",
        "extraValues": "error",
        "duplicateSlots": "error",
        "lockBaseline": "required-when-locked",
        "existingSlotsAfterLock": "immutable",
        "newSlotsAfterLock": "append-only"
      },
      "modeGroups": [
        {
          "name": "box",
          "purpose": "Container box styling for border, spacing, radius, and placement.",
          "appliesTo": ["field", "form", "table", "component", "panel"],
          "defaultSlots": ["color", "width", "style", "pt", "pr", "pb", "pl", "radius", "place"],
          "required": true
        }
      ],
      "baselines": [],
      "modes": [
        {
          "name": "box",
          "standard": true,
          "purpose": "Container box styling for border, spacing, radius, and placement.",
          "appliesTo": ["field", "form", "table", "component", "panel"],
          "slots": ["color", "width", "style", "pt", "pr", "pb", "pl", "radius", "place"]
        }
      ]
    }
  },
  "errors": []
}
```

Hex colors should be quoted because `#` starts a comment outside quoted strings.

Theme migration JSON shape:

```json
{
  "success": true,
  "command": "theme migrate",
  "version": "0.1.0-dev",
  "oldFile": "theme-v1.blackthm",
  "newFile": "theme-v2.blackthm",
  "safe": true,
  "summary": {
    "oldTheme": "WarehouseTheme",
    "newTheme": "WarehouseTheme",
    "oldVersion": 1,
    "newVersion": 2,
    "target": "web",
    "oldLocked": true,
    "newLocked": true,
    "oldProfile": "UICompact",
    "newProfile": "UICompact",
    "oldProfileVersion": 1,
    "newProfileVersion": 2
  },
  "changes": [
    {
      "type": "slot-appended",
      "profile": "UICompact",
      "mode": "box",
      "slot": "shadow",
      "index": 10
    }
  ],
  "errors": []
}
```

Current compact UI profile rules:

- Mode slot order is positional and read left to right.
- Inline syntax is `ui <mode> <values...> [| <mode> <values...>...]`.
- `ui <mode> = <slot...>;` defines the generator reading order for compact inline UI values.
- `|` separates multiple UI mode groups on one source line.
- Missing trailing values use CSS generation defaults.
- Extra values are errors because they cannot map to known slots.
- A slot name may appear only once inside the same mode.
- Web UI profiles must include the standard `box`, `text`, `table`, and `button` modes.
- `profile.modeGroups` tells AI agents what each standard mode means and where it applies.
- Locked profiles must include `baseline <mode> <slot...>` lines.
- Current mode slots must start with the exact baseline slot sequence.
- After a profile is locked, existing slots are immutable and new slots are append-only.
- Theme migration checks require old mode slots to remain the exact prefix of new mode slots.

For example, this locked profile change is valid because `shadow` is appended:

```blackthm
profile UICompact {
  version 2
  baseline box color width style
  ui box = color width style shadow;
}
```

This change is invalid because `shadow` is inserted before locked slots:

```blackthm
profile UICompact {
  version 2
  baseline box color width style
  ui box = color shadow width style;
}
```

It returns `NON_APPEND_ONLY_UI_SLOT`.

Legacy `mode <name> <slot...>` lines are still accepted for backward compatibility, but the preferred generator order syntax is `ui <mode> = <slot...>;`.

`black theme migrate <old.blackthm> <new.blackthm> --json` reports `UI_SLOT_MIGRATION_BREAK` when a theme replacement inserts, removes, or reorders existing slots. It also reports incompatible theme/profile name, target, version, lock, and mode removal diagnostics. Intentional reorders require future source migration tooling before they can be marked safe.

Current standard UI mode groups:

```text
box     container border, spacing, radius, and placement
text    typography for labels, headings, helper text, and body copy
table   table-specific border, density, and row pattern styling
button  action control styling for submit and page action buttons
```

If any standard mode is missing, `black theme inspect --json` returns `MISSING_STANDARD_UI_MODE`.

## Current Inline UI Intent

Draft v0.2 supports compact inline UI intent inside `.black` source. This records visual intent near the field, form, table, or action button it belongs to and generates deterministic CSS classes in the web target.

Syntax:

```black
ui <mode> <values...> [| <mode> <values...>...]
```

Current supported placements:

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

JSON shape on parsed fields, forms, tables, and action UI declarations:

```json
{
  "ui": [
    {
      "mode": "text",
      "values": ["#172026", "14", "semibold", "left"]
    }
  ]
}
```

Current semantic rules:

- Field UI can use `box` and `text`.
- Form UI can use `box`, `text`, and `button`.
- Table UI can use `box`, `text`, and `table`.
- Action button UI can use `button`.
- Table and form blocks can declare `id Identifier` and `class ClassName...`.
- Page actions can declare `action <name> id Identifier` and `action <name> class ClassName...`.
- Generated web IDs and custom classes are normalized to kebab-case.
- `action <name> ui button ...` must reference an action listed in `actions`.
- Repeating the same UI mode in the same scope returns `DUPLICATE_UI_INTENT`.
- Unknown mode names return `UNSUPPORTED_UI_MODE`.
- Mode names that do not apply to the current target return `UNSUPPORTED_UI_TARGET_MODE`.
- Repeating the same UI id in one page returns `DUPLICATE_UI_ID`.
- Malformed UI lines return `INVALID_UI_INTENT`, `INVALID_ACTION_UI`, or `INVALID_ACTION_INTENT`.

Generated web output appends CSS rules to `src/styles.css` using stable class names:

```text
table   .bl-ui-table-<page>
form    .bl-ui-form-<page>
field   .bl-ui-field-<entity>-<field>
action  .bl-ui-action-<page>-<action>
```

Explicit action IDs are expanded by generated usage site to avoid duplicated DOM IDs:

```text
create action opener    <id>-open
create/edit form submit <id>-submit
bulk action button      <id>-bulk
row action button       <id>-item-<recordId>
```

The current generator uses the configured `.blackthm` slot order when available, and falls back to standard v0.2 slots for `box`, `text`, `table`, and `button`. Missing trailing values fall back to safe defaults. Full `.blackthm` token value resolution remains a later extension.

## Current Page View Order and Composition

Draft v0.2 supports a `view` block inside a page for ordering and composing generated page sections from `.black` source.

Syntax:

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

  table {
    columns sku, name, stock
  }

  form {
    fields sku, name, stock
  }
}
```

Current supported sections:

```text
table
detail
form
explicitly declared component sections
```

Current supported composition modes:

```text
stack
grid
tabs
```

Rules:

- `view` is declared inside a `page` block.
- Only one `view` block may exist per page.
- Only one `order` line may exist inside `view`.
- Only one `compose` line may exist inside `view`.
- Empty `view` blocks return `MISSING_VIEW_ORDER`.
- Listed sections render first.
- Omitted supported sections are appended in default `table, detail, form` order.
- `section <Name> component <Component> bind selected|first|each` declares a reusable generated component section for a page.
- Component section names may appear in `order`, `group`, and `tab` declarations after they are declared with `section`.
- `bind selected` passes the current selected/detail record; `bind first` passes the first loaded list record; `bind each` renders one component instance per loaded list/query record.
- Component section inputs must be scalar primitive inputs whose names and types match stored or computed fields on the page source entity.
- Component sections render inline in this MVP and cannot use `display modal`, `display drawer`, or `side`.
- `compose stack` keeps vertical page flow and may set `gap sm|md|lg`.
- `compose grid` emits deterministic CSS grid rules and may set `columns 1..4`, `gap sm|md|lg`, and `stackAt sm|md|lg|none`.
- `compose tabs` emits generated React tab state and controls and may set `gap sm|md|lg`.
- `section detail|form display modal|drawer` emits generated overlay wrappers for detail/form panels.
- `section title "Text"` sets the generated modal/drawer panel heading.
- `side left|right` is valid only with `display drawer` and defaults to `right` for drawer sections.
- `group <Name> sections <section...> compose stack|grid` emits one generated wrapper around contiguous inline sections.
- Group `span 1..4` applies to the wrapper inside an outer grid; group `title "Text"` emits a generated heading.
- Grid `columns` defaults to `2` when omitted.
- Grid `stackAt` defaults to `md` when omitted.
- Grid `stackAt` emits responsive breakpoint media queries; `stackAt sm` can step 4 columns through lg/md/sm before the final single-column stack.
- Section spans that exceed the responsive column count are clamped at that breakpoint.
- Tabs reject `columns` and `stackAt`.
- `section table|detail|form|<ComponentSection> span 1..4` emits deterministic grid span classes for supported sections.
- `section table|detail|form display inline|modal|drawer` emits deterministic display classes; modal/drawer are supported only for detail/form in this MVP.
- `group <Name> sections table, StockSummary, detail` requires supported contiguous sections; grouped sections must stay inline in this MVP.
- `tab <Name> sections <section...>` assigns supported sections to one generated tab.
- Tabs require every ordered section to appear in exactly one tab.
- Tabs reject modal/drawer section display and groups in this MVP.
- `trigger <section> on <event>` binds supported generated page events to deterministic section state changes.
- Supported trigger events are `rowSelect`, `createStart`, `editStart`, `saveSuccess`, and `close`.
- `rowSelect` supports `detail` or `form`.
- `createStart` and `editStart` support `form`.
- `saveSuccess` supports `table` or `detail`.
- `close` supports `table`.
- A trigger event may be declared once per page.
- Trigger events that require missing CRUD actions return `UNSUPPORTED_VIEW_TRIGGER_ACTION`.
- Unsupported section names return `UNSUPPORTED_VIEW_SECTION`.
- Duplicate section names return `DUPLICATE_VIEW_SECTION`.
- Invalid component section names or built-in component section conflicts return `INVALID_VIEW_COMPONENT_SECTION` or `CONFLICTING_VIEW_COMPONENT_SECTION`.
- Unknown component references, missing/unsupported binds, unsupported component input shapes, missing matching source fields, and component input type mismatches return `UNKNOWN_VIEW_COMPONENT`, `MISSING_VIEW_COMPONENT_BIND`, `UNSUPPORTED_VIEW_COMPONENT_BIND`, `UNSUPPORTED_VIEW_COMPONENT_INPUT`, `UNKNOWN_VIEW_COMPONENT_INPUT_FIELD`, and `VIEW_COMPONENT_INPUT_TYPE_MISMATCH`.
- Unsupported composition modes return `UNSUPPORTED_VIEW_COMPOSE_MODE`.
- Unsupported columns, gaps, breakpoints, spans, section display/side values, group values, tab declarations, tab sections, and trigger declarations return `UNSUPPORTED_VIEW_COMPOSE_COLUMNS`, `UNSUPPORTED_VIEW_GAP`, `UNSUPPORTED_VIEW_STACK_AT`, `UNSUPPORTED_VIEW_SECTION_SPAN`, `UNSUPPORTED_VIEW_SECTION_DISPLAY`, `UNSUPPORTED_VIEW_SECTION_SIDE`, `UNSUPPORTED_VIEW_GROUP_*`, `UNSUPPORTED_VIEW_TAB`, `UNSUPPORTED_VIEW_TAB_SECTION`, and `UNSUPPORTED_VIEW_TRIGGER*`.
- The generated web page root includes `page-view`, `page-view-<page>`, optional `bl-view-compose-<mode>`, and optional `bl-view-has-triggers` classes.
- Generated table, detail, form, and component sections include stable `.bl-view-section-*` classes plus optional `.bl-view-span-*`, `.bl-view-display-*`, and `.bl-view-side-*` classes.
- Generated component sections also include `.bl-view-component-section` and `.bl-view-component-<component>` classes.
- Generated groups include stable `.view-group`, `.bl-view-group-*`, `.bl-view-group-compose-*`, and optional `.bl-view-span-*` classes.
- Generated tabbed pages include `.view-tabs`, local `activeViewTab` state, and `.view-tab-panel` wrappers.
- Generated modal/drawer sections include `.section-overlay`, `.section-overlay-modal`, or `.section-overlay-drawer-*` wrappers and close buttons.
- Generated triggers may switch `activeViewTab`, open generated form overlays, select saved records for detail, or return close flows to the table section.
- When a page declares `view`, generated `src/styles.css` receives deterministic order, composition, tabs, and overlay rules for that page.
- Arbitrary coordinates and arbitrary frontend event handlers are planned for later layout phases.

## Current Diagnostic Documentation

Draft v0.2 records stable diagnostic behavior in `docs/diagnostics.md`.

Every diagnostic should keep `file`, `line`, `column`, `code`, `message`, and optional `suggestion` fields. AI agents should branch on `code`, not on mutable human message text.

The compact CLI entry is available through:

```bash
black docs diagnostics --json
```

## Current Inspect Affected Command

Draft v0.2 supports focused affected graph output:

```bash
black inspect app.black --affected Product.stock --json
black inspect app.black --affected Products --json
black inspect app.black --affected OrderLifecycle --json
black inspect app.black --affected LowStockProducts --json
black inspect app.black --affected RestockProduct --json
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
    "queries": [],
    "actions": [],
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

Supported symbols include entities, entity fields such as `Product.stock`, entity indexes such as `Product.index`, queries, jobs, actions, transactions, seeds, tests, pages, roles, workflows, states, components, APIs, `target`, `auth`, `database`, `security`, `deploy`, `ops`, `i18n`, and `app`.

Unknown symbols return `UNKNOWN_AFFECTED_SYMBOL`. Missing `--affected` values return `MISSING_AFFECTED_SYMBOL`.

AI agents should run this before editing high-impact symbols so they can validate the right generated pages, routes, OpenAPI output, workflows, and role guards after the change.

## Current Explain Command

Draft v0.2 supports action-oriented keyword explanations:

```bash
black explain entity --json
black explain table --json
black explain action --json
black explain migrate --json
black explain seed --json
black explain test --json
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
- `black security encrypted-source --json` reports the `.black.enc` protected source policy for AI agents and CI tools.
- `black security encrypt <file>` creates `.black.enc` source using environment-provided key material.
- `black security decrypt <file.black.enc> --stdout` prints decrypted source only when explicitly requested.
- `black package --production` creates deployable artifacts without protected source files.
- Generated apps include `security/secrets.json`, `npm run security:secrets:plan`, and `npm run security:secrets:preflight` so deployment agents can verify required environment-backed secret references and provider CLI readiness without printing values.

Current source-safe style:

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

## Current Security Declaration

Draft v0.1 supports a top-level `security` declaration with explicit CORS intent:

```black
security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}
```

The parser, validator, JSON output, BlackIR output, generated `.env.example`, and generated Express server understand this declaration.

Rules:

- Use one `security` block per project.
- Use one `cors` block inside `security`.
- `origins` must use `env`.
- Environment variable names must use uppercase letters, numbers, and underscores.
- The generated server reads comma-separated origins from that environment variable.
- `credentials` may be `true` or `false`; use `true` for cookie-auth browser clients.
- Browser requests from origins not listed in the environment value are rejected.

## Current Source Security Commands

Draft v0.2 supports source-security scanning and protected source encryption:

```bash
black security scan --json
black security encrypted-source --json
black security encrypt app.black --out app.black.enc --json
black security decrypt app.black.enc --stdout
```

The scan checks `.black` source for likely hardcoded database URLs, private keys, API keys, tokens, secrets, and passwords.

`security encrypted-source` reports the `.black.enc` protected source policy. Encrypted source files use a plaintext header:

```text
BLACKLANG-ENC v1
alg AES-256-GCM
kdf SHA256-ENV
keyEnv BLACKLANG_SOURCE_KEY
nonce <base64>
---
<base64 ciphertext>
```

Key material comes from the environment variable named in the header or passed with `--key-env`. The key value must never be stored in `.black` source.

Parse, lint, validate, inspect, benchmark, security scan, and build can read `.black.enc` source in memory when the required key environment variable is set. The compiler does not write decrypted source during this flow.

`black security decrypt` requires `--stdout`. It prints plaintext only for explicit inspection in a trusted developer or CI environment. `black format` rejects `.black.enc` paths with `UNSUPPORTED_FORMAT_ENCRYPTED_SOURCE`; format plaintext source and re-encrypt instead.

Draft v0.1 also supports production packaging:

```bash
black package --production
```

The production package copies generated output while excluding `.black` and `.black.enc` source files, `.env`, local database files, `node_modules`, and generated Prisma client output.

Generated package scripts keep `db:push` mapped to BlackLang's generated database setup. SQLite targets use deterministic local table setup; PostgreSQL and MySQL targets use the generated Prisma schema through `prisma db push`. `db:push:native` is emitted as an explicit Prisma `db push` command for local schema-engine checks. When migration blocks exist, generated apps also expose `db:migrate:plan` and `db:migrate` for read-only deployment planning and explicit rename application before setup.

## Current Target Declaration

Draft v0.2 supports top-level `target` declarations for generated web and API-only output:

```black
target web {
  frontend react
  backend node
  database sqlite
}
```

PostgreSQL and MySQL runtime output use the same declaration shape with a different database target:

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

API-only output omits the frontend line:

```black
target api {
  backend node
  database sqlite
}
```

The parser, validator, JSON output, BlackIR output, affected graph, generated README summary, and documentation commands understand this declaration.

Rules:

- Use one `target` block per project.
- `target web` and `target api` are the supported application targets in v0.2.
- `target web` requires `frontend react`.
- `target api` must omit `frontend`; it emits Express/Prisma/OpenAPI/runtime files without React, Vite, frontend smoke, browser-check, browser e2e, or browser matrix files.
- `backend node` is the only supported backend in v0.2.
- `database sqlite`, `database postgres`, and `database mysql` are the supported generated runtime databases in v0.2.
- Generated rename migration runtime supports SQLite and PostgreSQL. `database mysql` reports `UNSUPPORTED_TARGET_DATABASE_MIGRATION` when `migration` blocks are present.
- If `target` is omitted, the generator keeps the legacy default stack: web, React, Node, SQLite.
- Unsupported future stacks must stay invalid until the corresponding generator adapter exists.

## Current Deployment Declaration

Draft v0.2 supports a top-level `deploy` declaration:

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

The parser, validator, JSON output, BlackIR output, generated `.env.example`, generated server, package scripts, Dockerfile, `.dockerignore`, Docker Compose output, local preview compose output, deployment manifest, rollback metadata, and cloud adapter plan/preflight/apply metadata understand this declaration.

Rules:

- Use one `deploy` block per project.
- `target docker` is the only supported target in v0.1.
- `port env NAME default PORT` declares the environment variable the generated server reads for its listen port.
- The default port must be between 1 and 65535.
- `env NAME required|optional` declares runtime environment variables without storing secret values in `.black` source.
- `required` and `optional` are documentation/runtime wiring modes in v0.2.
- `preview local` emits `docker-compose.preview.yml`, `BLACKLANG_PREVIEW_PORT`, and `deploy:preview` / `deploy:preview:down` package scripts.
- Local preview uses a separate host port and preview database defaults so it can be run beside the normal Docker Compose stack.
- `rollback keep COUNT` emits `deploy/manifest.json`, `deploy/rollback.json`, `scripts/rollback-plan.mjs`, and `deploy:rollback:plan`.
- Rollback keep counts must be between 1 and 20.
- Rollback plan output is read-only metadata. It reports retained release directories and a rollback candidate without mutating infrastructure.
- `cloud PROVIDER app env NAME [region env NAME]` emits `deploy/manifest.json`, `deploy/cloud.json`, `scripts/cloud-plan.mjs`, `scripts/cloud-exec.mjs`, `deploy:cloud:plan`, `deploy:cloud:preflight`, and `deploy:cloud:exec`.
- Current cloud provider adapter values are `fly`, `render`, and `railway`.
- Cloud app and region values must use environment variable references. Provider tokens, endpoints, app names, and regions must not be hardcoded in `.black` source.
- Cloud plan output is read-only metadata. Cloud preflight checks required environment values and provider CLI availability without mutating provider infrastructure. Cloud exec requires explicit apply mode and redacts configured environment values from captured provider output.
- Generated Docker Compose persists the default SQLite database under `/app/data` for SQLite targets unless `DATABASE_URL` overrides it.
- Generated Docker Compose emits a PostgreSQL service, healthcheck, and app `DATABASE_URL` fallback for PostgreSQL targets.
- Generated Docker Compose emits a MySQL 8.4 service, healthcheck, `MYSQL_*` local env defaults, and app `DATABASE_URL` fallback for MySQL targets.
- When an `ops` health endpoint is declared, generated Dockerfile and Docker Compose app service healthchecks probe that public path with Node's built-in `fetch`.

## Current Ops Declaration

Draft v0.2 supports deterministic generated runtime signals for web deployments:

```black
ops {
  health path "/healthz"
  readiness path "/readyz"
  metrics path "/metrics"
  logging requests
  observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT
}
```

Rules:

- Use one top-level `ops` block per project.
- `health path` generates a public root-level GET endpoint with process status, app name, CLI version, uptime, and start time.
- `readiness path` generates a public root-level GET endpoint that checks database connectivity and returns `503` when the database is not reachable.
- `metrics path` generates a public root-level GET endpoint with in-process request count, server-error count, status-code counts, uptime, and start time.
- `logging requests` emits one structured JSON console log when each observed request finishes.
- `observe webhook endpoint env NAME` sends non-blocking structured request events to an env-referenced endpoint when that environment variable is set.
- `observe otlp endpoint env NAME` sends non-blocking OTLP HTTP JSON trace payloads to an env-referenced endpoint when that environment variable is set.
- Observe middleware creates or propagates W3C `traceparent` context and adds the response `traceparent` header.
- Observability endpoint values must use environment variable references. Vendor URLs, DSNs, tokens, and API keys must not be hardcoded in `.black` source.
- Ops paths must start with `/`, cannot be `/`, `/api`, `/api/...`, or `/openapi.json`, and may only contain letters, digits, slash, dash, underscore, and dot.
- Ops endpoints are generated before API auth and CSRF middleware so infrastructure probes are not blocked by application login.
- Ops endpoints and observability hooks appear in parse JSON, BlackIR, inspect summaries, affected analysis, generated server output, OpenAPI output, `ops/observability.json`, generated contract tests, and generated API smoke tests. Use `black docs ops --json`, `black explain ops --json`, and `black inspect --affected ops --json` for focused agent context. The detailed reference is `docs/ops.md`.

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

The OpenAPI document is derived from ops endpoints, page source entities, page actions, bound custom queries, background query job metadata, bound custom actions, transaction bindings, service bindings, and explicit API declarations.

Draft v0.2 also supports explicit top-level API declarations:

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

Explicit API declarations are generated declared-runtime routes in draft v0.2. They appear in parse JSON, BlackIR, inspect/affected output, `generated/openapi.json`, generated server routes, generated contract tests, and generated API smoke tests. Paths must use a safe `/api/...` shape, must not use the reserved `/api/auth` namespace, and must not conflict with generated page, auth, query, action, or workflow routes. Generated routes validate declared path, query, and body parameters and return deterministic JSON with `runtime: "declared"`. Webhook routes return `202 accepted`. Private routes are protected when `auth` exists.

Explicit API handlers support a bounded update MVP. `body` fields are supported for `POST`, `PUT`, and `PATCH` APIs and may include `file` and `image` values as data URL or absolute `http(s)` strings. Body fields may use `required`, `optional`, `min`, `max`, `length`, `regex`, `url`, and `message`; `accept` belongs to entity media form fields. `update Entity where field == value set field = value` updates one row selected by `id` or a stored `unique` field. The block form `update Entity where field == value { ... }` supports ordered local `value` declarations plus indentation-based `if`/`else` branches with `value`, `set`, and nested `if` statements. Set targets must be stored primitive non-policy fields. Handler values may reference `body.name`, `param.name`, source fields in set expressions, local values, typed string/number/boolean literals, and bounded numeric expressions with `+`, `-`, `*`, `/`, parentheses, and deterministic precedence. Condition comparisons use `==`, `!=`, `<`, `<=`, `>`, and `>=`; ordered comparisons currently require numeric values. Comparisons may be joined with `and`, `or`, `not`, and parentheses, using `not` before `and` before `or` precedence. A top-level transaction block can wrap the generated bounded lookup, update, and private auth audit logging in one Prisma transaction. A top-level service block can group explicit APIs into generated service manifest/module metadata and OpenAPI service tags without changing request handling. Public handlers on tenant-scoped entities must include the tenant policy field as a required body field or path param. Public handlers cannot update owner-scoped entities.

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
POST   /api/products/{id}/actions/restockproduct
```

Create, update, delete, archive, restore, and custom action operations are included only when the page declares the matching action.

Custom action operations include request input schemas, response schemas, `x-blacklang-action`, `x-blacklang-source`, `x-blacklang-mutates-fields`, and optional `x-blacklang-transaction`/`x-blacklang-transaction-name` metadata. Explicit API update operations can expose the same transaction metadata when targeted by a top-level transaction block. Service-bound explicit API operations include OpenAPI `tags`, route-level `x-blacklang-service`, and root `x-blacklang-services` metadata.

Ops operations include `x-blacklang-ops` and `x-blacklang-public` metadata for health, readiness, and metrics endpoints.

The generated Express server serves this file at:

```text
/openapi.json
```

## Generated Contract/API/Frontend/Browser-Check Smoke Tests

Generated web and API-only apps include Node/TypeScript smoke tests at:

```text
generated/src/blacklang.contract.test.ts
generated/src/blacklang.api.test.ts
generated/src/blacklang.frontend.test.tsx  # target web only
generated/src/blacklang.browser.test.tsx   # target web only, when source declares test blocks
generated/src/blacklang.e2e.test.ts        # target web only, when source declares test blocks
generated/src/blacklang.e2e.matrix.ts      # target web only, when source declares test blocks
generated/tests/browser-matrix.json        # target web only, when source declares test blocks
```

The generated `package.json` exposes fast smoke checks and explicit browser e2e checks as:

```bash
npm test
npm run test:e2e
npm run test:e2e:plan
npm run test:e2e:matrix
```

In draft v0.2, generated `npm test` starts with `npm run db:generate`, then reads `openapi.json`, asserts generated ops paths and metadata, entity schemas, action schemas, CRUD paths, bound query and query summary paths, custom action paths, explicit API paths, explicit API runtime metadata, and action metadata including transaction metadata flags. It also imports generated validation functions and checks representative valid and invalid entity/action inputs. When seed declarations exist, generated output also includes `src/seed.ts`, `db:seed`, and `db:setup` wiring; run `npm run db:setup` for runtime proof that deterministic fixture rows apply after schema setup. When test declarations exist in `target web`, generated output also includes `src/blacklang.browser.test.tsx`, `src/blacklang.e2e.test.ts`, `src/blacklang.e2e.matrix.ts`, `tests/browser-matrix.json`, `test:e2e`, `test:e2e:plan`, `test:e2e:matrix`, and `test:all`; `npm test` runs deterministic page/action/text browser-check expectations, `npm run test:e2e` runs generated auth/navigation/action checks in one real browser, and `npm run test:e2e:matrix` runs the same checks across every available matrix target. `target api` package tests stay contract/API-only after `db:generate`.

The API smoke test imports `createApp()` from the generated server, starts it on a random localhost port, checks `/openapi.json`, probes explicit API runtime routes, probes declared ops endpoints, checks anonymous API rejection when auth exists, checks JSON 404 behavior when auth is absent, and checks generated CORS allow/deny behavior when `security.cors` exists.

The frontend smoke test renders the generated React `App` with `react-dom/server` and checks that auth-enabled apps render the session-check state or unauthenticated apps render the app shell. It is emitted only for `target web`.

The browser-check smoke test also renders the generated React `App`, checks declared page metadata, checks declared action exposure on the target page, and checks generated text expectations against the render surface or generated text catalog. It is emitted only for `target web`. The browser e2e test starts generated API/Vite servers, runs generated setup/seed modules, launches a local browser executable, performs generated auth registration when needed, navigates pages, and checks visible text/action expectations. The generated browser matrix plan reads executable candidates only; the matrix run reuses the same e2e test for custom, Chrome, Edge, and Chromium targets.

The tests are generated output; do not edit them by hand. Change `.black` source or the generator.

## Source and Generated Output Benchmarks

BlackLang CLI can measure source size, generated output size, deterministic AI task scenarios, token estimates, web coverage, evidence-backed AI eval history, and tracked issue exports:

```bash
black benchmark app.black --out generated --json
black benchmark app.black --out generated --ir
black benchmark tasks app.black --out generated --json
black benchmark tasks app.black --out generated --ir
black benchmark eval app.black --out generated --json
black benchmark eval app.black --out generated --ir
black benchmark eval-history --json
black benchmark eval-history --ir
black benchmark coverage --json
black benchmark coverage --ir
black benchmark issues --json
black benchmark issues --ir
```

`black benchmark` loads, parses, and validates the project, then builds web output in a temporary directory and counts only files emitted by the BlackLang generator. It does not mutate the configured output directory. Source metrics include the primary `.black` file and the configured `.blackthm` theme file when one exists.

The JSON result includes `source`, `generated`, `ratios`, `sourceFiles`, `generatedFiles`, and `generatedKinds`. AI agents should use these fields for benchmark notes instead of guessing generated size.

`black benchmark tasks` loads and validates the project, runs the same side-effect-free temporary generated output measurement, and returns deterministic AI task benchmark scenarios. The JSON result includes `baseline`, `scenarios`, and `totals`.

Each scenario contains:

- `id`
- `name`
- `purpose`
- `trigger`
- `projectEvidence`
- `blacklangEdits`
- `generatedImpact`
- `commands`
- `estimatedTokens`

`estimatedTokens` compares a BlackLang edit loop with a conventional generated-stack edit loop using current source/generated line and byte counts, fixed scenario context sizes, and deterministic text estimates. The estimates are planning signals, not billed-token measurements, and the command never calls an AI model.

`black benchmark eval` loads and validates the project, reuses the same side-effect-free source/generated measurements, and returns a long-running AI eval corpus for repeated BlackLang web editing trials. The JSON result includes `suite`, `cases`, and `totals`. `suite` declares the corpus id, mode, repeat count, timeout, policy, and discovery commands. Each case includes a prompt, expected evidence, required commands, a 100-point scoring rubric, and deterministic token estimates. The command is read-only metadata: it does not call an AI model, mutate generated output, commit, push, deploy, or store secrets.

`black benchmark eval-history` reads `benchmarks/eval-history.blackdir` or the path passed with `--history <file>` and returns evidence-backed result history. The JSON result includes `history`, `summary`, and `runs`; each run records suite id, status, source, coverage signal, case/repeat counts, pass/fail totals, evidence paths, notes, and model or validation-source results. Local validation entries may record compiler and generated-app checks, but must not be presented as billed model benchmark scores. External model-to-model scores require repeated harness runs, command evidence, and release review before being appended.

`black benchmark coverage` reports the current weighted web coverage matrix, remaining tracked issue list, and percentage milestones. The JSON result includes `completionPercent`, `areas`, `issues`, and `milestones`. `completionPercent` is a weighted progress signal for planning, not a claim that every web application requirement is solved.

`black benchmark issues` reports the same tracked issue IDs as a compact export for AI/CI issue sync. The JSON result includes `summary.total`, `summary.open`, `summary.done`, `issues`, and `openIssues`.

## Generated Secure Defaults

Generated web API servers should start with safe baseline behavior before explicit auth syntax exists.

In draft v0.1, generated Express servers:

- Disable the `X-Powered-By` header
- Add basic browser security headers
- Limit JSON request bodies to `100kb`
- Apply a simple IP-based request rate limit
- Apply CORS middleware when `security.cors` is declared

Baseline defaults are generated automatically. CORS middleware requires explicit `security.cors` source syntax.

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

- `BlackUser` stores a primary `role` value and a `roles` set.
- Newly registered users receive the first declared role by default.
- `/api/auth/me` returns both the primary role and assigned role set.
- Pages with `access` generate API route guards.
- Users without an allowed role receive `403 Forbidden`.
- When roles exist, a generated Users page lets the first declared role list users and assign one or more roles. When tenant policies exist, the same page can update user tenant IDs.
- Permission actions generate API route guards:
  - `read` protects list and detail endpoints.
  - `create` protects create endpoints.
  - `update` protects update, archive, and restore endpoints.
  - `delete` protects single and bulk delete endpoints.
- Generated React pages hide create, edit, archive, restore, and delete controls when none of the current user roles allows the required action.
- Permission checks evaluate the assigned role set. `deny` permissions from any assigned role override matching `allow` permissions; otherwise an `allow` from any assigned role grants access.
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

This is page-level, action-level, field-level, owner, tenant, and multi-role enforcement plus basic role/tenant management, audit log UI, CSRF/session protection, source secret scanning, protected source encryption, generated secret reference manifests, read-only secret manager provider preflight execution, signed release trust verification, release transparency log policy, and key rotation policy. Provider-owned runtime value injection remains a later adapter layer.

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
theme = "src/theme.blackthm"
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
black agent startup --json
black theme inspect --json
black docs agent-contract --json
black docs ui --json
black docs ui-profile --json
black docs ui-modes --json
black docs view --json
black docs migrate --json
black explain migrate --json
black docs query --json
black explain query --json
black migrate plan old.black new.black --json
black docs action --json
black docs transaction --json
black docs service --json
black explain action --json
black explain transaction --json
black explain service --json
black inspect app.black --affected RestockProduct --json
black inspect app.black --affected RestockAtomic --json
black inspect app.black --affected InventoryIntegration --json
black docs ops --json
black explain ops --json
black inspect app.black --affected ops --json
black docs benchmark --json
black benchmark app.black --json
black benchmark tasks app.black --json
black benchmark eval app.black --json
black benchmark eval-history --json
black ide --json
black ide diagnostics app.black --json
black ecosystem --json
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
.black     source language written by humans and AI agents
.blackthm  UI theme/profile source language
.blackir   compact BlackLang intermediate representation
.json      standard integration format for external tools
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
  actions create edit delete archive restore RestockProduct

action RestockProduct source Product
  input quantity number required min 1
  value restockValue = quantity
  if restockValue > 0 and stock >= 0
    set stock = stock + restockValue
  else
    set stock = stock

transaction RestockAtomic
  action RestockProduct

service InventoryIntegration
  api StockWebhook
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
- `query` (v0.2)
- `job` (v0.2)
- `action` (v0.2)
- `transaction` (v0.2)
- `service` (v0.2)
- `ops` (v0.2)
- `role`
- `api`
- `migration` (v0.2)
- `seed` (v0.2)
- `test` (v0.2)
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

Page `view` blocks can also render a declared component as a reusable inline section:

```black
view {
  order table, StockSummary, StockCards, detail, form
  section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
  section StockCards component StockBadge bind each span 1 title "Stock Cards"
}
```

For this MVP, page component sections accept `bind selected`, `bind first`, or `bind each`; every component input must be a scalar primitive input whose name and type matches a stored or computed field on the page source entity.

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

Current generated Prisma output stores `number` and `integer` as signed integer values. Use `decimal` or `money` for fractional numeric values.

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

Top-level `message Entity.field { ... }` localizes generated frontend field-level validation messages for stored form fields. Entity-level `validate ... message "Text"` remains explicit validation metadata in the current draft.

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

Rules:

- Use one `i18n` block per project.
- `default` declares the initial locale used by generated web output.
- `locales` declares accepted locale names.
- The default locale must be included in `locales`.
- Top-level `label` blocks target stored fields and computed display fields with `Entity.field`.
- Top-level `label` blocks can also target generated UI copy with `app.*`, `page.*`, `action.*`, `table.*`, and `status.*`.
- Top-level `placeholder`, `help`, and `message` blocks target stored form fields with `Entity.field`.
- When more than one locale is declared, generated web UI includes a language selector.
- App chrome, navigation page names, table tools, status labels, CRUD buttons, custom action buttons, workflow transition buttons, and table/detail/form/filter/column labels re-render when the user changes locale.
- Generated form placeholders, form help text, and field-level frontend validation messages re-render when translated blocks exist.
- Generated table/detail `number`, `integer`, `decimal`, `money`, `date`, and `datetime` values use `Intl` formatting for the active locale.
- Generated App shell includes `lang` and deterministic `dir`; `ar`, `fa`, `he`, `ur`, and region variants are RTL.
- Runtime field text falls back to the default locale translation when the active locale is missing.
- If no translation exists, generated UI falls back to matching inline field metadata such as `label "Text"`, `placeholder "Text"`, `help "Text"`, or `message "Text"`.
- If no field label modifier exists, labels fall back to title-cased field name.

Supported UI label targets are `app.title`, `app.menu`, `app.language`, `app.logout`, `app.closeNavigation`, `app.primaryNavigation`, `app.checkingSession`, `app.noPages`, `app.source`, `app.recordForm`, `page.<DeclaredPage>`, generated auth pages `page.Users` and `page.Audit`, static CRUD/status targets such as `action.create`, `action.saveChanges`, `table.search`, `table.columns`, `table.previous`, `status.active`, and dynamic targets `action.<CustomAction>`, `action.<workflowTransition>`, `action.new.<Entity>`, `action.create.<Entity>`, `action.edit.<Entity>`, and `action.view.<Entity>`.

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
  customer Customer required load detail query
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
- Entity reference fields may use `load list`, `load detail`, `load query`, `load mutation`, or `load none` to control generated API response relation attachment contexts.
- Omitting `load` preserves the default relation attachment behavior in list, detail, query, and mutation responses.
- `load none` must appear by itself and returns only the generated relation ID field in generated responses.
- Generated list and query routes batch relation IDs into one target lookup per loaded relation field. Detail and mutation responses use the same deterministic attach helper for a single item.
- Permission-aware generated routes attach a relation object only when the current role can read the source relation field, and sanitize attached target records with the target entity's read field policy.
- Relation load policy does not change the generated database schema or add relation paths to custom query filters.
- Generated OpenAPI operations expose `x-blacklang-relation-load-context` and `x-blacklang-relation-load` metadata for relation fields.

Entity row policy lines are written inside the entity after declaring their stored fields:

```black
entity Order {
  tenantId text required default "default"
  ownerId text required default "system"
  total money default 0
  policy tenant tenantId
  policy owner ownerId
}
```

`policy owner <field>` and `policy tenant <field>` require auth and a stored `text required` non-unique field. Generated create/update routes stamp these fields from authenticated user context, and generated list/query/detail/mutation routes scope rows by them.

## Page

```black
page Products {
  source Product

  view {
    order form, table, detail
  }

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
- Optional `query QueryName` binds the page list to a query with the same entity source.
- `actions` may include declared custom action names when the action source matches the page source.
- Page names must remain distinct after lowercasing to avoid generated route/module collisions (`DUPLICATE_PAGE_ROUTE`).
- `view.order` may list `table`, `detail`, `form`, and declared component section names to control generated section order.
- Generated JSX/DOM section order follows the effective page view order. CSS order rules are still emitted as stable metadata and a layout backstop.
- `table.columns` must reference fields on the source entity.
- `table.search` must reference searchable fields on the source entity. Draft v0.1 supports `text`, `email`, and entity reference search fields.
- `form.fields` must reference fields on the source entity.
- Form labels use `label "Text"` when present, otherwise the field name is title-cased.
- Form inputs use `placeholder "Text"` when present.
- Form fields show `help "Text"` below the generated input when present.
- `actions` may include `create`, `edit`, `delete`, and custom action names. It may also include `archive` and `restore` for soft delete behavior.
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
black agent startup --json
black theme inspect --json
black migrate plan old.black new.black --json
black docs entity --ir
black docs action --json
black explain action --json
black docs --all --json
black explain entity --json
black inspect --affected RestockProduct --json
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
- Duplicate/missing query declarations and unknown query source entities
- Invalid query field, operator, typed literal, sort direction, or limit
- Unknown page queries and page/query source mismatches
- Missing/duplicate custom action declarations and unknown action source entities
- Invalid custom action inputs, set fields, expression values, allow roles, and page/action source mismatches
- Invalid, empty, duplicate, or unsupported transaction block bindings
- Invalid or duplicate ops declarations, unsafe ops paths, empty ops blocks, and unsupported ops logging modes
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
      "code": "UNKNOWN_TABLE_COLUMN",
      "message": "Page Products table uses unknown field Product.barcode.",
      "suggestion": "Add the field to the source entity or remove it from columns."
    }
  ]
}
```
