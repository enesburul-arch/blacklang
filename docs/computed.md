# Computed Fields

Computed fields declare read-only display values derived from stored entity fields.

They are intended for generated table columns and detail views, not database storage or form input.

## Syntax

```black
entity Product {
  stock number default 0
  price money default 0
  computed inventoryValue money = stock * price label "Inventory Value"
}
```

Shape:

```text
computed <name> <type> = <left> <operator> <right> [label "Text"] [help "Text"]
```

Supported computed types in v0.2:

```text
number
integer
decimal
money
```

Supported operators in v0.2:

```text
+
-
*
/
```

Operands must be stored number-like fields on the same entity or numeric literals.

## Page Usage

```black
page Products {
  source Product

  table {
    columns stock, price, inventoryValue
  }
}
```

Generated web output computes the value in the React page from the loaded record.

The computed field is also shown in the generated detail view.

## Rules

- Computed fields are not Prisma columns.
- Computed fields are not SQLite columns.
- Computed fields are not submitted by generated forms.
- Computed fields can be listed in table `columns`.
- Computed fields cannot be used in form `fields`.
- Computed field `search`, `filter`, and `sort` support is planned for a later data logic phase.
- Custom query `where` and `sort` only accept stored primitive fields; computed display fields may still appear in the bound page's table columns.
- If a computed field reads a source field hidden by field-level permissions, the computed value is hidden too.

## Agent Notes

Use `black docs computed --json` before writing computed field syntax.

Use `black inspect --affected Entity.computedField --json` before changing a computed field in an existing project.

After edits, run:

```bash
black format --check --json
black lint --json
black validate --json
black build
```
