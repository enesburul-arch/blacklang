# SupportDesk Example

SupportDesk is a reusable BlackLang app template for customer support, ticket queues, SLA tracking, comments, and knowledge base publishing.

It demonstrates current v0.1 language features without introducing new syntax:

- email/password cookie auth
- secret-safe database environment reference
- organization/customer/team/agent/SLA/ticket/comment/article entities
- relation fields across support records
- role and field-level access control
- ticket lifecycle and article review workflows
- page state declarations
- reusable priority and published status components
- explicit API contracts
- table filtering, sorting, pagination, and column visibility
- generated CRUD actions with archive/restore behavior

Validate the template from `packages/cli/`:

```bash
go run ./cmd/black validate ../../examples/helpdesk/app.black --json
go run ./cmd/black build ../../examples/helpdesk/app.black --out ../../generated-helpdesk --json
```
