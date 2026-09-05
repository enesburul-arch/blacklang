# AppointmentBook Template Benchmark v0.2

## Purpose

This benchmark records the first reusable appointment scheduling app template for BlackLang.

The goal is to measure whether one compact `.black` source can describe a realistic scheduling application pattern with auth, roles, relations, workflows, validation, API contracts, and generated web behavior.

## Source

```text
examples/appointment/app.black
```

## Measured Metrics

```text
Metric                 Value
BlackLang source lines 457
Generated files        59
Generated lines        13944
Entities               9
Pages                  9
Roles                  4
Workflows              2
Components             2
Explicit APIs          2
```

## Feature Coverage

- Auth with cookie sessions
- Secret-safe `database { url env DATABASE_URL }`
- Client, Location, Service, StaffMember, Room, AvailabilityBlock, Appointment, AppointmentReminder, and WaitlistEntry entities
- Entity relation fields
- Field labels, placeholders, help text, min/max, length, regex, email, date, datetime, boolean, and conditional validation
- Admin, SchedulingManager, Practitioner, and ClientViewer roles
- Field-level read and mutation restrictions
- Appointment lifecycle workflow
- Reminder lifecycle workflow
- Page state and modal declarations
- Duration and confirmation display components
- Contract-first availability search and appointment webhook APIs
- CRUD pages with search, filter, sort, pagination, archive, and restore

## Notes

This is a template benchmark, not a claim that every appointment scheduling feature is complete in v0.1.

The template intentionally stays inside currently supported syntax so compiler validation can act as the source of truth.

## Verification

Measured on 2026-09-05:

```bash
black parse examples/appointment/app.black --json
black validate examples/appointment/app.black --json
black lint examples/appointment/app.black --json
black build examples/appointment/app.black --out <temp> --json
npm install
npm run build
```

All parse, validate, lint, generator, TypeScript, and Vite production build checks passed.

`npm install` reported 4 high-severity audit findings from generated app dependencies. They were not force-fixed in this template benchmark because that can introduce dependency changes outside the BlackLang source measurement.
