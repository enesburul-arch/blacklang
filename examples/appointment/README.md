# AppointmentBook Example

AppointmentBook is a reusable BlackLang app template for clients, services, staff schedules, resources, appointments, reminders, and waitlists.

It demonstrates current v0.1 language features without introducing new syntax:

- email/password cookie auth
- secret-safe database environment reference
- client/location/service/staff/room/availability/appointment/reminder/waitlist entities
- relation fields across scheduling records
- role and field-level access control
- appointment and reminder workflows
- page state declarations
- reusable duration and confirmation display components
- explicit API contracts
- table filtering, sorting, pagination, and column visibility
- generated CRUD actions with archive/restore behavior

Validate the template from `packages/cli/`:

```bash
go run ./cmd/black validate ../../examples/appointment/app.black --json
go run ./cmd/black build ../../examples/appointment/app.black --out ../../generated-appointment --json
```
