# Warehouse Benchmark v0.1

Date: 2026-09-05

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
Total lines: 164
Code lines:  135
```

The source describes:

- 1 app
- auth intent: emailPassword with cookie session
- secret-safe database intent: `url env DATABASE_URL`
- 3 entities: Product, Customer, Order
- 2 roles: Admin, Worker
- 2 explicit API contracts: LowStockReport, StockWebhook
- 3 pages: Products, Customers, Orders
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
- explicit API path/query/path-param/access/webhook metadata in OpenAPI
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
- generated frontend and API validation from field constraints
- generated frontend and API validation from cross-field constraints
- generated frontend and API validation from conditional required constraints
- parsed and validated explicit API declarations
- explicit API intent in JSON/BlackIR outputs

## Generated Web Output

Generated folder:

```text
generated/
```

Counted generated source files:

```text
Generated source files: 32
Generated source lines: 5385
```

Excluded from this measurement:

- `generated/node_modules/`
- `generated/dist/`
- `generated/src/generated/prisma/`
- runtime database files
- generated README text
- `.env.example`

## Ratio

```text
BlackLang code lines:       135
Generated web source lines: 5385
Approximate ratio:          39.9x
```

In this benchmark, one BlackLang source line represents about 40 generated web stack source lines.

## Counted Generated Files

```text
generated/package.json
generated/index.html
generated/tsconfig.json
generated/vite.config.ts
generated/prisma.config.ts
generated/openapi.json
generated/prisma/schema.prisma
generated/src/main.tsx
generated/src/App.tsx
generated/src/db.ts
generated/src/setup-db.ts
generated/src/server.ts
generated/src/styles.css
generated/src/vite-env.d.ts
generated/src/types.ts
generated/src/auth/AuthPage.tsx
generated/src/auth/UsersPage.tsx
generated/src/auth/AuditPage.tsx
generated/src/components/StockBadge.tsx
generated/src/routes/auth.ts
generated/src/api/product.ts
generated/src/api/customer.ts
generated/src/api/order.ts
generated/src/routes/product.ts
generated/src/routes/customer.ts
generated/src/routes/order.ts
generated/src/validation/product.ts
generated/src/validation/customer.ts
generated/src/validation/order.ts
generated/src/pages/ProductsPage.tsx
generated/src/pages/CustomersPage.tsx
generated/src/pages/OrdersPage.tsx
```

## Verification Commands

```bash
dist/black.exe validate --json
dist/black.exe security scan --json
dist/black.exe build
dist/black.exe package --production
cd generated
npm run build
npm run db:setup
```

Last verified result:

```text
BlackLang validation: passed
BlackLang security scan: passed
BlackLang build:      passed
Production package:   passed
Generated web build:  passed
Generated DB setup:   passed
Workflow API smoke:   passed
```

## Notes

The generated line count is not the same as a hand-written minimum implementation.

A human could write a smaller version by omitting structure, type safety, repeated CRUD behavior, or generated consistency. But a realistic React + TypeScript + Express + Prisma implementation with the same behavior would still require many files and substantially more code than the `.black` source.
