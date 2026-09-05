# InvoiceFlow Example

InvoiceFlow is a reusable BlackLang app template for clients, billing contacts, catalog items, invoices, invoice lines, payments, and credit notes.

It demonstrates current v0.1 language features without introducing new syntax:

- email/password cookie auth
- secret-safe database environment reference
- client/contact/tax/catalog/invoice/payment/credit entities
- relation fields across billing records
- role and field-level access control
- invoice and payment workflows
- page state declarations
- reusable balance and payment display components
- explicit API contracts
- table filtering, sorting, pagination, and column visibility
- generated CRUD actions with archive/restore behavior

Validate the template from `packages/cli/`:

```bash
go run ./cmd/black validate ../../examples/invoice/app.black --json
go run ./cmd/black build ../../examples/invoice/app.black --out ../../generated-invoice --json
```
