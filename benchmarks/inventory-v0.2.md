# InventoryControl Template Benchmark v0.2

## Purpose

This benchmark records the first reusable inventory app template for BlackLang.

The goal is to measure whether one compact `.black` source can describe a realistic stock management application pattern with auth, roles, relations, workflows, validation, API contracts, and generated web behavior.

## Source

```text
examples/inventory/app.black
```

## Measured Metrics

```text
Metric                 Value
BlackLang source lines 340
Generated files        47
Generated lines        9826
Entities               6
Pages                  6
Roles                  4
Workflows              2
Components             2
Explicit APIs          2
```

## Feature Coverage

- Auth with cookie sessions
- Secret-safe `database { url env DATABASE_URL }`
- Warehouse, Supplier, Category, Item, PurchaseOrder, and StockMovement entities
- Entity relation fields
- Field labels, placeholders, help text, min/max, length, URL, regex, and cross-field validation
- Admin, InventoryManager, StockClerk, and Viewer roles
- Field-level read and mutation restrictions
- Purchase order workflow
- Stock movement workflow
- Page state and modal declarations
- Stock level and movement completion display components
- Contract-first low stock report and stock adjustment webhook APIs
- CRUD pages with search, filter, sort, pagination, archive, and restore

## Notes

This is a template benchmark, not a claim that every inventory feature is complete in v0.1.

The template intentionally stays inside currently supported syntax so compiler validation can act as the source of truth.

## Verification

Measured on 2026-09-05:

```bash
black parse examples/inventory/app.black --json
black validate examples/inventory/app.black --json
black lint examples/inventory/app.black --json
black build examples/inventory/app.black --out <temp> --json
npm install
npm run build
```

All parse, validate, lint, generator, TypeScript, and Vite production build checks passed.

`npm install` reported 4 high-severity audit findings from generated app dependencies. They were not force-fixed in this template benchmark because that can introduce dependency changes outside the BlackLang source measurement.
