# Warehouse Benchmark v0.1

Date: 2026-09-06

## Purpose

This benchmark records the current BlackLang advantage for the generated Warehouse web application.

The goal is not to claim a final universal ratio. The goal is to keep a concrete, repeatable measurement as BlackLang grows.

## Source

BlackLang source file:

```text
examples/warehouse/app.black
```

Current source contains:

```text
Total lines: 770
Code lines:  639
```

The source describes:

- 1 app
- auth intent: emailPassword with cookie session
- secret-safe database intent: `url env DATABASE_URL`
- generated secret provider preflight: `npm run security:secrets:preflight`
- Docker deploy intent with local preview, rollback metadata, cloud plan/preflight, and explicit provider CLI apply runner
- 3 entities: Product, Customer, Order
- 2 roles: Admin, Worker
- 2 explicit API declared-runtime routes: LowStockReport, StockWebhook
- 4 pages: Products, LowStock, Customers, Orders
- 3 deterministic seed/fixture declarations for Product, Customer, and Order demo rows
- 1 deterministic browser-check/browser e2e test declaration for Products
- 1 workflow: OrderPreparation
- 1 client state declaration: OrdersPageState
- 1 component declaration: StockBadge
- page access intent for Products, Customers, and Orders
- CRUD actions
- archive and restore actions
- relation field: Order.customer -> Customer
- relation select input
- relation empty-state guidance
- relation navigation to related page
- relation display in table/detail
- relation response load policy for list/detail/query/mutation contexts
- generated batch relation response attachment
- nested relation response sanitization in permission-aware routes
- relation search
- field labels
- field placeholders
- field help text
- inline form validation messages
- field constraint modifiers: `min`, `max`, and `length`
- entity-level cross-field validation: `validate discount <= total`
- conditional required validation: `validate trackingNumber required when status == shipped`
- default table sorting
- table pagination
- column visibility controls
- field-level table filters
- generated application shell
- sidebar navigation
- topbar and breadcrumb
- explicit layout declaration
- explicit sidebar navigation order
- responsive drawer navigation
- generated OpenAPI contract
- generated OpenAPI route
- explicit API path/query/path-param/body/access/webhook metadata in OpenAPI and generated runtime routes
- bounded explicit API update handler for the StockWebhook endpoint
- generated secure API defaults
- generated login/register UI shell
- generated auth API routes
- generated cookie session tables
- generated password hashing
- protected generated CRUD API routes
- generated session restore and logout behavior
- parsed and validated role declarations
- parsed and validated page access declarations
- role and access intent in JSON/BlackIR outputs
- generated single-role storage for authenticated users
- generated page-level role guards from `access`
- generated basic Users role management page
- generated action-level API permission guards
- generated role-aware action controls in React pages
- generated field-level read hiding in API responses and React pages
- generated field-level mutation filtering for create/update payloads
- generated audit log storage, API endpoint, and Audit page
- generated CSRF protection for authenticated cookie write requests
- parsed and validated workflow source, states, transitions, and transition allow roles
- workflow intent in JSON/BlackIR outputs
- generated workflow transition API routes
- generated workflow transition API client methods
- generated workflow transition row action buttons
- generated workflow OpenAPI paths
- generated workflow status updates
- generated workflow audit entries
- parsed and validated client state fields and modals
- state intent in JSON/BlackIR outputs
- generated React state hooks from matching state declarations
- generated modal open/close helpers from state declarations
- parsed and validated component inputs and variants
- component intent in JSON/BlackIR outputs
- generated standalone React component file
- generated component variant class selection
- generated component table/detail rendering binding
- generated live component preview in matching form fields
- generated reusable page component section binding
- generated JSX DOM order following effective page view order
- generated frontend and API validation from field constraints
- generated frontend and API validation from cross-field constraints
- generated frontend and API validation from conditional required constraints
- parsed and validated explicit API declarations
- explicit API intent in JSON/BlackIR outputs and generated server routes
- custom query intent for LowStockProducts
- generated query list route and API client
- generated query aggregate summary route, API client method, OpenAPI metadata, and React summary cards
- transactional custom action route intent for RestockProduct
- first-class seed/fixture intent in JSON/BlackIR/inspect/affected output
- generated deterministic `src/seed.ts` runtime and `db:setup`/`db:seed` package scripts
- first-class browser-check/browser e2e intent in JSON/BlackIR/inspect/affected output
- generated deterministic `src/blacklang.browser.test.tsx` browser-check runtime and `npm test` wiring
- generated deterministic `src/blacklang.e2e.test.ts` browser e2e runtime with `test:e2e`/`test:all` package scripts
- generated `docker-compose.preview.yml`, `BLACKLANG_PREVIEW_PORT`, and preview package scripts
- generated `deploy/manifest.json`, `deploy/rollback.json`, and read-only rollback plan script
- generated `deploy/cloud.json`, read-only cloud plan/preflight scripts, and explicit apply provider CLI runner script
- runtime i18n labels for generated app chrome, page names, table/status copy, CRUD actions, custom actions, and workflow transitions

## Generated Web Output

Generated folder:

```text
generated/
```

Counted generated source files:

```text
Generated source files: 65
Generated source lines: 14503
```

Excluded from this measurement:

- `generated/node_modules/`
- `generated/dist/`
- `generated/src/generated/prisma/`
- runtime database files

## Ratio

```text
BlackLang source lines:     770
Generated web source lines: 14503
Approximate ratio:          18.84x
```

In this benchmark, one BlackLang source line represents about 18.84 generated web stack source lines.

## Counted Generated Files

The authoritative path list is emitted by:

```bash
black benchmark examples/warehouse/app.black --out generated --json
```

The current run reports 64 deterministic generator-owned files, including API clients, routes, OpenAPI, tests, migrations, deterministic seed runtime, generated browser-check/e2e runtime, Docker files, local preview compose, rollback metadata, cloud plan/preflight/explicit-apply scripts, React pages, CSS, and validation modules.

## AI Task Benchmark Signal

The deterministic AI task scenario report is emitted by:

```bash
black benchmark tasks examples/warehouse/app.black --out generated --json
black benchmark eval examples/warehouse/app.black --out generated --json
```

Current scenario summary:

```text
Scenarios:                       8
Estimated BlackLang tokens:      10519
Estimated conventional tokens:   90872
Estimated savings signal:        88%
```

Scenario IDs:

```text
AI-TASK-QUERY-001
AI-TASK-ACTION-001
AI-TASK-SEED-001
AI-TASK-TEST-001
AI-TASK-POLICY-001
AI-TASK-API-001
AI-TASK-OPS-001
AI-TASK-DEPLOY-001
```

These are deterministic planning estimates derived from current source/generated measurements and fixed scenario context sizes. They are not billed-token measurements and they do not call an AI model.

## Verification Commands

```bash
dist/black.exe validate --json
dist/black.exe security scan --json
dist/black.exe ide --json
dist/black.exe ide diagnostics examples/warehouse/app.black --json
dist/black.exe build
dist/black.exe package --production
cd generated
npm run build
npm run db:setup
npm test
npm run test:e2e
npm run test:e2e:plan
npm run test:e2e:matrix
npm run deploy:rollback:plan
npm run deploy:cloud:plan
npm run deploy:cloud:preflight
black benchmark tasks examples/warehouse/app.black --out generated --json
black benchmark eval examples/warehouse/app.black --out generated --json
```

Last verified result:

```text
BlackLang validation: passed
BlackLang security scan: passed
BlackLang IDE export:    passed
BlackLang IDE diagnostics: passed
BlackLang build:      passed
Production package:   passed
Generated web build:  passed
Generated DB setup:   passed
Generated npm test:    passed
Generated e2e test:   passed
Rollback plan script: passed
Cloud plan script:    passed
Cloud preflight:      passed
AI task benchmark:    passed
AI eval corpus:        passed
Browser matrix plan:   passed
Browser matrix run:    passed
Workflow API smoke:   passed
```

## Notes

The generated line count is not the same as a hand-written minimum implementation.

A human could write a smaller version by omitting structure, type safety, repeated CRUD behavior, or generated consistency. But a realistic React + TypeScript + Express + Prisma implementation with the same behavior would still require many files and substantially more code than the `.black` source.
