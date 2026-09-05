# ProjectPulse Template Benchmark v0.2

## Purpose

This benchmark records the first reusable project management app template for BlackLang.

The goal is to measure whether one compact `.black` source can describe a realistic project delivery application pattern with auth, roles, relations, workflows, validation, API contracts, and generated web behavior.

## Source

```text
examples/project-management/app.black
```

## Measured Metrics

```text
Metric                 Value
BlackLang source lines 519
Generated files        63
Generated lines        15621
Entities               10
Pages                  10
Roles                  4
Workflows              2
Components             2
Explicit APIs          2
```

## Feature Coverage

- Auth with cookie sessions
- Secret-safe `database { url env DATABASE_URL }`
- Organization, Team, TeamMember, Project, Milestone, ProjectTask, TaskComment, TimeEntry, Risk, and ProjectUpdate entities
- Entity relation fields
- Field labels, placeholders, help text, min/max, length, regex, email, date, boolean, money, cross-field validation, and conditional validation
- Admin, ProjectManager, TeamMemberRole, and ClientViewer roles
- Field-level read and mutation restrictions
- Project lifecycle workflow
- Task lifecycle workflow
- Page state and modal declarations
- Progress and billable display components
- Contract-first project health report and project webhook APIs
- CRUD pages with search, filter, sort, pagination, archive, and restore

## Notes

This is a template benchmark, not a claim that every project management feature is complete in v0.1.

The template intentionally stays inside currently supported syntax so compiler validation can act as the source of truth.

## Verification

Measured on 2026-09-05:

```bash
black parse examples/project-management/app.black --json
black validate examples/project-management/app.black --json
black lint examples/project-management/app.black --json
black build examples/project-management/app.black --out <temp> --json
npm install
npm run build
```

All parse, validate, lint, generator, TypeScript, and Vite production build checks passed.

`npm audit` reported 4 high-severity findings from generated app dependencies. They were not force-fixed in this template benchmark because that can introduce dependency changes outside the BlackLang source measurement.
