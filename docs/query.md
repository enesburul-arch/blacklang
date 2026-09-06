# Custom Queries

Custom queries declare reusable, bounded lists of stored entity records. A page binds a named query to display that list.

## Syntax

```black
query LowStockProducts {
  source Product
  where stock < 10
  sort stock asc
  limit 50
}

page LowStock {
  source Product
  query LowStockProducts

  table {
    columns sku, name, stock, price, inventoryValue
    search sku, name
    paginate 10
  }
}
```

`Product` and the table fields must already exist. `inventoryValue` may be a computed display field; query conditions and sort may only use stored fields.

The declaration is top-level. Query names use PascalCase letters and digits and must remain unique after lowercasing, including against other top-level symbols. Use one `source Entity`, zero or more `where field operator literal` lines, one optional `sort field asc|desc`, and one optional `limit integer`. A page keeps its `source Entity` and adds one `query QueryName`; both sources must match. A query can be reused by several pages. Omit forms and actions for a read-only page. Page names must also remain distinct after lowercasing because they determine route and module paths.

## Types and Operators

| Stored field type | Literal | Operators |
| --- | --- | --- |
| `text`, `email` | Quoted string, such as `"draft"` or `"buyer@example.com"` | `==`, `!=` |
| `number`, `integer` | Signed 32-bit whole-number literal, such as `10` | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| `decimal`, `money` | Finite decimal literal, such as `10` or `12.50` | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| `boolean` | Unquoted `true` or `false` | `==`, `!=` |
| `date` | Quoted date, such as `"2026-09-06"` | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| `datetime` | Quoted timestamp with timezone, such as `"2026-09-06T12:00:00Z"` | `==`, `!=`, `<`, `<=`, `>`, `>=` |

Multiple distinct `where` lines are combined with AND; identical repeated conditions are rejected. Strings must be quoted; numeric and boolean literals must not be quoted. Numeric literals use ordinary decimal notation; exponent notation, hexadecimal, NaN, and Infinity are invalid. Dates must be valid calendar dates, and datetimes must use RFC3339. The validator checks literals against the stored field type. Conditions cannot compare fields with each other.

Relations, computed fields, generated system fields, `null` literals, parameters, OR expressions, joins, aggregates, raw SQL, and arbitrary code are outside this MVP. `source` is the only source syntax; do not use `from`.

## Deterministic Results

- Filters run in the generated server before sorting and limiting.
- `sort` accepts one stored primitive field and `asc` or `desc`.
- Without `sort`, records use generated `id asc` order. A declared sort adds `id asc` as a tie breaker.
- `limit` is a whole number from 1 through 1000. The default is 100.
- The query follows the existing archived-record visibility behavior.
- Text comparison and null ordering follow the current SQLite database behavior.

These rules keep the returned subset stable for unchanged data and archive mode. They do not add cursor pagination or a total count of all matches.

## Generated Web Behavior

A bound page uses `GET /api/<lowercase-page-name>/query`, generated before the item route. For `page LowStock`, the route is `GET /api/lowstock/query`. The generated page API client exposes `queryList`, and the React page loads its records through that method. Query declarations without a bound page do not create standalone runtime endpoints.

The existing entity list and CRUD routes keep their behavior. Relation selectors continue to load the ordinary entity list, so a low-stock page does not remove other products from relation options.

Page search, table filters, pagination, and an optional `table.sort` operate on the returned subset. `table.sort` may reorder that subset but cannot change which records the server selected. Without `table.sort`, the table preserves query order. After successful mutations on a bound page, the page refetches its query so records entering or leaving the subset and limit boundaries stay correct.

Query routes retain existing authentication, page access, and entity read permission checks. They also require read permission for every field used in `where` or `sort`; a denied field returns `403 Forbidden`. Response field hiding still applies.

A query is a list-selection rule. It does not impose ownership, tenant, row authorization, or new restrictions on existing detail and mutation routes.

## AI Agent Workflow

```bash
black docs query --json
black explain query --json
black inspect --affected LowStockProducts --json
black inspect --affected Product.stock --json
black format --check --json
black lint --json
black validate --json
black build --json
```

Parse, inspect, affected analysis, and BlackIR expose query declarations and page bindings. Use [diagnostics.md](diagnostics.md) for stable error codes and repair guidance. Keep runtime secrets in environment variables; query literals are source data and must not contain secrets.

Common repairs:

| Code | Repair |
| --- | --- |
| `UNKNOWN_QUERY_SOURCE` | Use an existing entity in `source`. |
| `UNKNOWN_QUERY_FIELD` | Use a field declared on that entity. |
| `UNSUPPORTED_QUERY_FIELD` | Replace a computed or relation field with a stored primitive field. |
| `QUERY_LITERAL_TYPE_MISMATCH` | Match the literal's type and range to the field. |
| `UNSUPPORTED_QUERY_OPERATOR` | Use an operator supported for the field type. |
| `INVALID_QUERY_LIMIT` | Use a whole number from 1 through 1000. |
| `UNKNOWN_PAGE_QUERY` | Declare the referenced query or correct the page binding. |
| `PAGE_QUERY_SOURCE_MISMATCH` | Make the query source and page source match. |
| `QUERY_NAME_COLLISION` | Give the query a name not used by another top-level symbol. |
