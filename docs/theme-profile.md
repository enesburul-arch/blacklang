# BlackLang Theme Profile

This document defines the first `.blackthm` file format.

`.blackthm` files are source assets. They are not generated CSS. Their purpose is to describe UI tokens and compact UI mode slot profiles in a deterministic way that humans, AI agents, and the compiler can inspect.

## File Role

```text
.black     application intent
.blackthm  UI theme/profile intent
.blackir   compact compiler output
.json      external tool integration output
.css       generated style output in later phases
```

## Project Config

Projects may point to a theme file from `blacklang.toml`:

```toml
version = "0.1"
target = "web"
source = "examples/warehouse/app.black"
out = "generated"
theme = "examples/warehouse/theme.blackthm"
```

## Syntax

```blackthm
blackthm WarehouseTheme {
  version 1
  target web
  locked true

  token color primary "#2563eb"
  token color surface "#ffffff"
  token space sm 8
  token radius md 6

  profile UICompact {
    version 1
    baseline box color width style pt pr pb pl radius place
    baseline text color size weight align
    baseline table color width style density zebra
    baseline button bg color radius size variant

    mode box color width style pt pr pb pl radius place
    mode text color size weight align
    mode table color width style density zebra
    mode button bg color radius size variant
  }
}
```

## Rules

- A theme file starts with `blackthm <Name> {`.
- `version <number>` declares the theme file version.
- `target web` is the only supported target in draft v0.2.
- `locked false` means the profile is still editable.
- `locked true` means existing positional slots are frozen and checked against baselines.
- `token <kind> <name> <value>` declares reusable design tokens.
- Hex colors must be quoted because `#` starts a comment outside quoted strings.
- A theme must contain one `profile <Name> { ... }`.
- A profile must contain `version <number>`.
- A profile must contain one or more `mode` lines.
- A locked profile must contain `baseline <mode> <slot...>` lines.
- `mode <name> <slot...>` defines the left-to-right slot order for compact inline UI intent.
- A slot name may appear only once inside one mode.
- Web UI profiles must contain standard `box`, `text`, `table`, and `button` mode groups.
- In locked profiles, every baseline must be a prefix of the matching current mode.
- In locked profiles, new slots are valid only when appended after the baseline prefix.

## Compact UI Slot Rules

`black theme inspect --json` returns these rules in `profile.rules`:

```json
{
  "inlineSyntax": "ui <mode> <values...> [| <mode> <values...>...]",
  "slotOrder": "left-to-right",
  "modeSeparator": "|",
  "missingTrailingSlots": "default",
  "extraValues": "error",
  "duplicateSlots": "error",
  "lockBaseline": "required-when-locked",
  "existingSlotsAfterLock": "immutable",
  "newSlotsAfterLock": "append-only"
}
```

The same output includes `profile.modeGroups`, a stable machine-readable list of standard groups:

```json
[
    {
      "name": "box",
      "purpose": "Container box styling for border, spacing, radius, and placement.",
      "appliesTo": ["field", "form", "table", "component", "panel"],
      "defaultSlots": ["color", "width", "style", "pt", "pr", "pb", "pl", "radius", "place"],
      "required": true
    }
]
```

An inline UI line:

```black
ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
```

is read from left to right with the active profile:

```text
box.color  = black
box.width  = 1
box.style  = solid
box.pt     = 8
box.pr     = 8
box.pb     = 5
box.pl     = 5
box.radius = 6
box.place  = center
text.color = "#172026"
text.size  = 14
text.weight = regular
text.align = left
```

This keeps UI intent short without making the generator guess which value belongs to which style property.

## Locked Profile Example

When a profile is locked, baseline lines record the frozen prefix:

```blackthm
profile UICompact {
  version 2
  baseline box color width style pt pr pb pl radius place
  mode box color width style pt pr pb pl radius place shadow
}
```

This is valid because `shadow` is appended at the end.

This is invalid:

```blackthm
profile UICompact {
  version 2
  baseline box color width style pt pr pb pl radius place
  mode box color width shadow style pt pr pb pl radius place
}
```

It reports `NON_APPEND_ONLY_UI_SLOT` because the old slot order changed.

## Standard Mode Groups

```text
box     container border, spacing, radius, and placement
text    typography for labels, headings, helper text, and body copy
table   table-specific border, density, and row pattern styling
button  action control styling for generated and explicit buttons
```

Inline `.black` source uses these groups near fields, forms, tables, and page action buttons. Use `black docs ui --json` for the exact inline placement rules.

If a web profile is missing one of these modes, `black theme inspect --json` reports `MISSING_STANDARD_UI_MODE`.

## Current CLI

```bash
black theme inspect --json
black theme inspect examples/warehouse/theme.blackthm --json
black theme inspect examples/warehouse/theme.blackthm --ir
```

In this phase, `black theme inspect` reads and validates `.blackthm` metadata. CSS generation from `.blackthm` is planned for later Phase 20 steps.

## AI Agent Rule

AI agents should inspect the theme before writing inline UI intent:

```bash
black theme inspect --json
```

Then use the returned `profile.modes[].slots` arrays as the source of truth for positional UI values.
