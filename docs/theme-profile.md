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
  locked false

  token color primary "#2563eb"
  token color surface "#ffffff"
  token space sm 8
  token radius md 6

  profile UICompact {
    version 1
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
- `locked true` will mean existing positional slots are frozen in later phases.
- `token <kind> <name> <value>` declares reusable design tokens.
- Hex colors must be quoted because `#` starts a comment outside quoted strings.
- A theme must contain one `profile <Name> { ... }`.
- A profile must contain `version <number>`.
- A profile must contain one or more `mode` lines.
- `mode <name> <slot...>` defines the left-to-right slot order for compact inline UI intent.

## Current CLI

```bash
black theme inspect --json
black theme inspect examples/warehouse/theme.blackthm --json
black theme inspect examples/warehouse/theme.blackthm --ir
```

In this phase, `black theme inspect` reads and validates `.blackthm` metadata. CSS generation from `.blackthm` is planned for later Phase 20 steps.

## AI Agent Rule

AI agents should inspect the theme before writing future inline UI intent:

```bash
black theme inspect --json
```

Then use the returned `profile.modes[].slots` arrays as the source of truth for positional UI values.
