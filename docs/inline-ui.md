# BlackLang Inline UI Intent

This document defines the current compact `ui` syntax inside `.black` files.

## Purpose

Inline UI intent keeps styling instructions near the field, form, table, or action button they belong to, without requiring raw CSS in the source file.

The web generator now turns supported inline UI intent into stable CSS classes in `src/styles.css`.

## Syntax

```black
ui <mode> <values...> [| <mode> <values...>...]
```

Values are positional. AI agents should inspect the active `.blackthm` profile before writing values:

```bash
black theme inspect --json
black docs ui --json
```

## Placements

Field-level typography:

```black
entity Product {
  name text required ui text "#172026" 14 semibold left
}
```

Field-level `ui` is trailing metadata. Put ordinary field modifiers such as `required`, `label`, `placeholder`, and `help` before `ui`.

Table-level intent:

```black
table {
  columns sku, name, stock
  ui table border 1 solid compact true
}
```

Form-level intent:

```black
form {
  fields sku, name, stock
  ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
}
```

Action button intent:

```black
actions create, edit
action create ui button primary white 6 md solid
```

## Current Mode Rules

```text
field   box, text
form    box, text, button
table   box, text, table
button  button
```

`action <name> ui button ...` must reference an action listed in the page `actions` line.

## JSON Shape

Parsed fields, forms, tables, and action UI declarations expose `ui` arrays:

```json
{
  "ui": [
    {
      "mode": "text",
      "values": ["#172026", "14", "semibold", "left"]
    }
  ]
}
```

## Generated CSS

Generated web output uses deterministic class names:

```text
table   .bl-ui-table-<page>
form    .bl-ui-form-<page>
field   .bl-ui-field-<entity>-<field>
action  .bl-ui-action-<page>-<action>
```

Current standard slot mapping:

```text
box     color width style pt pr pb pl radius place
text    color size weight align
table   color width style density zebra
button  bg color radius size variant
```

Missing trailing values use safe defaults. This phase includes built-in color aliases such as `primary`, `black`, `white`, `border`, `muted`, `danger`, and `success`. Full `.blackthm` token resolution is a later extension.

## Diagnostics

```text
INVALID_UI_INTENT
INVALID_ACTION_UI
UNSUPPORTED_UI_MODE
UNSUPPORTED_UI_TARGET_MODE
DUPLICATE_UI_INTENT
UNKNOWN_ACTION_UI
```
