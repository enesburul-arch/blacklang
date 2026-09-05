# BlackLang View Order

`view` is a page-level block that controls the generated section order inside one page.

It exists so humans and AI agents can move page parts such as the form, table, and detail panel from `.black` source instead of editing generated React or CSS.

## Syntax

```black
page Products {
  source Product

  view {
    order form, table, detail
  }

  table {
    columns sku, name, stock
  }

  form {
    fields sku, name, stock
  }
}
```

## Current Sections

Draft v0.2 supports these page sections:

```text
table
detail
form
```

Listed sections render first. Omitted supported sections are appended in the default order:

```text
table, detail, form
```

For example:

```black
view {
  order form
}
```

means:

```text
form, table, detail
```

## Generated Web Output

The web generator adds stable section classes:

```text
table   .bl-view-section-table
detail  .bl-view-section-detail
form    .bl-view-section-form
```

When a page declares `view`, generated `src/styles.css` receives deterministic order rules:

```css
.page-view-products .bl-view-section-form {
  order: 1;
}
```

## Rules

- Use `view` only inside a `page` block.
- Use one `view` block per page.
- Use one `order` line inside `view`.
- Current supported section names are `table`, `detail`, and `form`.
- Duplicate section names are validation errors.
- Unsupported section names are validation errors.
- Nested layout, grid, tabs, modal, drawer, and exact positioning are later composition features.

## AI Agent Notes

Use `black docs view --json` or `black explain view --json` before editing page section order.

Prefer changing:

```black
view {
  order form, table, detail
}
```

instead of editing generated React or CSS.

After changing view order, run:

```bash
black format --check --json
black lint --json
black validate --json
black build
```

## Diagnostics

```text
INVALID_VIEW_DECLARATION
INVALID_VIEW_ORDER
DUPLICATE_VIEW
DUPLICATE_VIEW_ORDER
MISSING_VIEW_ORDER
UNSUPPORTED_VIEW_SECTION
DUPLICATE_VIEW_SECTION
UNCLOSED_VIEW
UNEXPECTED_VIEW_TOKEN
```
