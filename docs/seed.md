# Seed and Fixture Declarations

Seed declarations describe deterministic local/demo database rows in `.black` source.

They are for development fixtures, generated smoke testing, demos, and repeatable local setup. They are not a place for production secrets or credentials.

## Syntax

```black
seed DemoProducts {
  source Product

  row DemoProductLow {
    tenantId "default"
    sku "LOW-001"
    name "Low Stock Widget"
    stock 3
    price 19.99
  }
}

seed DemoOrders {
  source Order

  row DemoOrderDraft {
    customer ref DemoCustomerAcme
    total 125.50
  }
}
```

Each `seed` block targets one entity through `source`. Each `row` has one stable key. The generated seed runtime uses that key as the row `id`, so rerunning `db:setup` or `db:seed` updates the declared rows instead of creating duplicates.

## Values

Scalar fields use typed literals:

- quoted strings for `text`, `email`, `date`, and `datetime`
- finite numbers for `number`, `integer`, `decimal`, and `money`
- `true` or `false` for `boolean`

Relation fields must use `ref`:

```black
customer ref DemoCustomerAcme
```

The referenced row key must exist in a seed block for the related entity.

## Validation

BlackLang validates seed rows before build:

- seed names use PascalCase and stay unique
- source entities must exist
- row keys must be stable identifiers and unique per entity
- field names must be stored fields on the source entity
- computed display fields cannot be seeded
- required fields without defaults must be present
- scalar values must match the field type
- relation fields must use `ref`
- referenced rows must exist
- basic field constraints such as `min`, `max`, `length`, `regex`, and `url` must pass
- duplicate seed values for unique fields are rejected

## Generated Web Behavior

When a project declares seeds, `black build` emits:

```text
generated/src/seed.ts
```

The generated `package.json` also includes:

```json
{
  "db:setup": "tsx src/setup-db.ts && npm run db:seed",
  "db:seed": "tsx src/seed.ts"
}
```

SQLite and PostgreSQL seed runtimes use parameterized upserts. MySQL seed runtime uses generated Prisma upserts through the provider adapter. The `.black` syntax does not expose raw SQL.

Use:

```bash
black docs seed --json
black explain seed --json
black inspect app.black --affected DemoProducts --json
black build app.black --out generated
cd generated
npm run db:setup
```
