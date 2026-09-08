# BlackLang Entity Indexes

Entity indexes keep database performance intent in `.black` source.

## Syntax

```black
entity Order {
  customer Customer required
  status text default draft
  index customer, status
}
```

## Rules

- Declare indexes inside an `entity` block.
- Use `index field` for one stored field.
- Use `index fieldA, fieldB` for a composite stored-field index.
- Index fields must be stored fields declared on the same entity.
- Relation fields are allowed; the generated database uses the relation ID column.
- Computed display fields cannot be indexed because they are not database columns.
- One index may contain at most 4 fields.
- Duplicate fields inside one index are validation errors.
- Duplicate index declarations with the same ordered field list are validation errors.

## Generated Web Output

The web generator emits:

- `@@index([...], map: "..._idx")` in `prisma/schema.prisma`
- `CREATE INDEX IF NOT EXISTS ...` in `src/setup-db.ts`

For relation fields, the generated Prisma and SQLite outputs use the generated foreign-key column:

```black
entity Order {
  customer Customer required
  status text
  index customer, status
}
```

```prisma
@@index([customerId, status], map: "Order_customer_status_idx")
```

## AI Agent Notes

Use indexes for fields that appear in page filters, table sorts, custom query `where` clauses, custom query `sort` clauses, or relation lookups.

Before changing index declarations, run:

```bash
black inspect app.black --affected Order.index --json
```

Before deploying an index change to an existing database, compare old and new source:

```bash
black migrate plan old.black new.black --json
```

After changing index declarations, run:

```bash
black format --check --json
black lint --json
black validate --json
black build
```

Index declarations affect database schema/setup output. They do not change API response shape or generated React pages.

## Diagnostics

```text
INVALID_ENTITY_INDEX
UNKNOWN_INDEX_FIELD
UNSUPPORTED_COMPUTED_INDEX_FIELD
DUPLICATE_INDEX_FIELD
DUPLICATE_ENTITY_INDEX
UNSUPPORTED_ENTITY_INDEX_FIELD_COUNT
```
