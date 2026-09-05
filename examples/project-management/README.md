# ProjectPulse Example

ProjectPulse is a reusable BlackLang app template for organizations, delivery teams, projects, milestones, tasks, time entries, risks, and project updates.

It demonstrates current v0.1 language features without introducing new syntax:

- email/password cookie auth
- secret-safe database environment reference
- organization/team/member/project/milestone/task/comment/time/risk/update entities
- relation fields across project records
- role and field-level access control
- project and task workflows
- page state declarations
- reusable progress and billable display components
- explicit API contracts
- table filtering, sorting, pagination, and column visibility
- generated CRUD actions with archive/restore behavior

Validate the template from `packages/cli/`:

```bash
go run ./cmd/black validate ../../examples/project-management/app.black --json
go run ./cmd/black build ../../examples/project-management/app.black --out ../../generated-project-management --json
```
