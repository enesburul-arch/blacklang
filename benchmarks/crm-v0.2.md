# SalesCRM Template Benchmark v0.2

## Purpose

This benchmark records the first reusable CRM app template for BlackLang.

The goal is to measure whether one compact `.black` source can describe a realistic business application pattern with auth, roles, relations, workflows, validation, API contracts, and generated web behavior.

## Source

```text
examples/crm/app.black
```

## Measured Metrics

```text
Metric                 Value
BlackLang source lines 264
Generated files        39
Generated lines        7480
Entities               4
Pages                  4
Roles                  3
Workflows              2
Components             2
Explicit APIs          2
```

## Feature Coverage

- Auth with cookie sessions
- Secret-safe `database { url env DATABASE_URL }`
- Company, Contact, Deal, and Activity entities
- Entity relation fields
- Field labels, placeholders, help text, min/max, length, URL, and cross-field validation
- Admin, SalesManager, and SalesRep roles
- Field-level read and mutation restrictions
- Deal pipeline workflow
- Activity follow-up workflow
- Page state and modal declarations
- Probability and completion display components
- Contract-first forecast and lead webhook APIs
- CRUD pages with search, filter, sort, pagination, archive, and restore

## Notes

This is a template benchmark, not a claim that every CRM feature is complete in v0.1.

The template intentionally stays inside currently supported syntax so compiler validation can act as the source of truth.

## Verification

Measured on 2026-09-05:

```bash
black parse examples/crm/app.black --json
black validate examples/crm/app.black --json
black lint examples/crm/app.black --json
black build examples/crm/app.black --out <temp> --json
npm install
npm run build
```

All parse, validate, lint, generator, TypeScript, and Vite production build checks passed.

`npm install` reported 4 high-severity audit findings from generated app dependencies. They were not force-fixed in this template benchmark because that can introduce dependency changes outside the BlackLang source measurement.
