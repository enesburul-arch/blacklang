# Custom Queries

Custom queries declare reusable, bounded lists of stored entity records. A page can bind a named query to display that list, and a background job can run the same query as a read-only worker task.

## Syntax

```black
query LowStockProducts {
  source Product
  where stock < 10
  aggregate lowStockCount count
  aggregate totalStock sum stock
  aggregate averagePrice avg price
  sort stock asc
  limit 50
}

job LowStockMonitor {
  schedule every 15 minutes
  run query LowStockProducts
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

The declaration is top-level. Query names use PascalCase letters and digits and must remain unique after lowercasing, including against other top-level symbols. Use one `source Entity`, zero or more `where field operator literal` lines, zero or more `aggregate` lines, one optional `sort field asc|desc`, and one optional `limit integer`. A page keeps its `source Entity` and adds one `query QueryName`; both sources must match. A query can be reused by several pages and by read-only background jobs. Omit forms and actions for a read-only page. Page names must also remain distinct after lowercasing because they determine route and module paths.

## Types and Operators

| Stored field type | Literal | Operators |
| --- | --- | --- |
| `text`, `email`, `file`, `image` | Quoted string, such as `"draft"`, `"buyer@example.com"`, or a data URL | `==`, `!=` |
| `number`, `integer` | Signed 32-bit whole-number literal, such as `10` | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| `decimal`, `money` | Finite decimal literal, such as `10` or `12.50` | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| `boolean` | Unquoted `true` or `false` | `==`, `!=` |
| `date` | Quoted date, such as `"2026-09-06"` | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| `datetime` | Quoted timestamp with timezone, such as `"2026-09-06T12:00:00Z"` | `==`, `!=`, `<`, `<=`, `>`, `>=` |

Multiple distinct `where` lines are combined with AND; identical repeated conditions are rejected. Strings must be quoted; numeric and boolean literals must not be quoted. Numeric literals use ordinary decimal notation; exponent notation, hexadecimal, NaN, and Infinity are invalid. Dates must be valid calendar dates, and datetimes must use RFC3339. The validator checks literals against the stored field type. Conditions cannot compare fields with each other.

Relation fields, computed fields, generated system fields, `null` literals, parameters, OR expressions, joins, raw SQL, and arbitrary code are outside query filter/sort expressions in this MVP. Relation response inclusion is controlled separately with relation field `load` scopes. `source` is the only source syntax; do not use `from`.

## Aggregate Summary

Aggregate lines add a generated query summary endpoint and summary cards on bound pages:

```black
aggregate lowStockCount count
aggregate totalStock sum stock
aggregate averagePrice avg price
aggregate lowestStock min stock
aggregate highestPrice max price
```

`count` has no field. `sum`, `avg`, `min`, and `max` require one stored numeric field: `number`, `integer`, `decimal`, or `money`. Aggregate names use normal identifiers, must be unique inside the query, and cannot use JavaScript prototype-reserved names such as `__proto__`. Aggregate fields follow the same read-permission guard as `where` and `sort` fields.

The generated summary route uses the same `source`, `where`, archived-record mode, and row policy as the query list. It does not apply `sort` or `limit`, because those belong to the bounded list window, not to the aggregate result. Empty result sets return `0` for `count` and `null` for `sum`, `avg`, `min`, and `max`.

## Deterministic Results

- Filters run in the generated server before sorting and limiting.
- `sort` accepts one stored primitive field and `asc` or `desc`.
- Without `sort`, records use generated `id asc` order. A declared sort adds `id asc` as a tie breaker.
- `limit` is a whole number from 1 through 1000. The default is 100.
- The query follows the existing archived-record visibility behavior.
- Text comparison and null ordering follow the current SQLite database behavior.

These rules keep the returned subset stable for unchanged data and archive mode. They do not add cursor pagination or an implicit total count; declare `aggregate someCount count` when the page needs a deterministic match count.

## Generated Web Behavior

A bound page uses `GET /api/<lowercase-page-name>/query`, generated before the item route. For `page LowStock`, the route is `GET /api/lowstock/query`. When the query declares aggregates, the page also gets `GET /api/<lowercase-page-name>/query/summary`. The generated page API client exposes `queryList` and, when needed, `querySummary`; the React page loads records through the list method and renders summary cards from the summary method. Query declarations without a bound page do not create standalone runtime endpoints. Query declarations referenced by a job generate worker behavior instead of public routes; see [job.md](job.md).

The existing entity list and CRUD routes keep their behavior. Relation selectors continue to load the ordinary entity list, so a low-stock page does not remove other products from relation options.

Relation fields can use `load query` when bound query results should include the related object for label display. Without `load query`, generated query rows still include the relation ID field and generated UI falls back to that ID. Omitting `load` on the relation field keeps the backward-compatible default and includes the relation object in query responses.

Page search, table filters, pagination, and an optional `table.sort` operate on the returned subset. `table.sort` may reorder that subset but cannot change which records the server selected. Without `table.sort`, the table preserves query order. After successful mutations on a bound page, the page refetches its query so records entering or leaving the subset and limit boundaries stay correct.

Query routes retain existing authentication, page access, entity read permission checks, and any entity row policy. They also require read permission for every field used in `where`, `sort`, or `aggregate`; a denied field returns `403 Forbidden`. Response field hiding still applies.

A query is a list-selection rule. It does not declare ownership, tenant, row authorization, or new restrictions on existing detail and mutation routes. Use `policy owner ownerId` or `policy tenant tenantId` on the entity when those routes should share row scope.

## AI Agent Workflow

```bash
black docs query --json
black explain query --json
black inspect --affected LowStockProducts --json
black inspect --affected LowStockMonitor --json
black inspect --affected Product.stock --json
black format --check --json
black lint --json
black validate --json
black build --json
```

Parse, inspect, affected analysis, and BlackIR expose query declarations, page bindings, and job consumers. Use [diagnostics.md](diagnostics.md) for stable error codes and repair guidance. Keep runtime secrets in environment variables; query literals are source data and must not contain secrets.

Common repairs:

| Code | Repair |
| --- | --- |
| `UNKNOWN_QUERY_SOURCE` | Use an existing entity in `source`. |
| `UNKNOWN_QUERY_FIELD` | Use a field declared on that entity. |
| `UNSUPPORTED_QUERY_FIELD` | Replace a computed or relation field with a stored primitive field. |
| `QUERY_LITERAL_TYPE_MISMATCH` | Match the literal's type and range to the field. |
| `UNSUPPORTED_QUERY_OPERATOR` | Use an operator supported for the field type. |
| `INVALID_QUERY_AGGREGATE` | Use `aggregate name count` or `aggregate name sum field`. |
| `DUPLICATE_QUERY_AGGREGATE` | Keep each aggregate name unique inside the query. |
| `UNSUPPORTED_QUERY_AGGREGATE` | Use `count`, `sum`, `avg`, `min`, or `max`. |
| `MISSING_QUERY_AGGREGATE_FIELD` | Add a stored numeric field to `sum`, `avg`, `min`, or `max`. |
| `UNSUPPORTED_QUERY_AGGREGATE_FIELD` | Remove the field from `count` or choose a stored numeric field. |
| `INVALID_QUERY_LIMIT` | Use a whole number from 1 through 1000. |
| `UNKNOWN_PAGE_QUERY` | Declare the referenced query or correct the page binding. |
| `PAGE_QUERY_SOURCE_MISMATCH` | Make the query source and page source match. |
| `QUERY_NAME_COLLISION` | Give the query a name not used by another top-level symbol. |
