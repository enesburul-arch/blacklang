# SalesCRM Example

SalesCRM is a reusable BlackLang app template for account, contact, deal, and activity tracking.

It demonstrates current v0.1 language features without introducing new syntax:

- email/password cookie auth
- secret-safe database environment reference
- account/contact/deal/activity entities
- relation fields between entities
- role and field-level access control
- pipeline and activity workflows
- page state declarations
- reusable components
- explicit API contracts
- table filtering, sorting, pagination, and column visibility
- generated CRUD actions with archive/restore behavior

Validate the template from `packages/cli/`:

```bash
go run ./cmd/black validate ../../examples/crm/app.black --json
go run ./cmd/black build ../../examples/crm/app.black --out ../../generated-crm --json
```
