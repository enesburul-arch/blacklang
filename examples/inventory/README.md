# InventoryControl Example

InventoryControl is a reusable BlackLang app template for warehouse, supplier, item, purchase order, and stock movement tracking.

It demonstrates current v0.1 language features without introducing new syntax:

- email/password cookie auth
- secret-safe database environment reference
- warehouse/supplier/category/item/purchase-order/stock-movement entities
- relation fields between inventory records
- role and field-level access control
- purchase order and stock movement workflows
- page state declarations
- reusable stock status components
- explicit API contracts
- table filtering, sorting, pagination, and column visibility
- generated CRUD actions with archive/restore behavior

Validate the template from `packages/cli/`:

```bash
go run ./cmd/black validate ../../examples/inventory/app.black --json
go run ./cmd/black build ../../examples/inventory/app.black --out ../../generated-inventory --json
```
