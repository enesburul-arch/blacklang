# BlackLang UI Profile Rules

This document defines the compact positional UI slot rules used by `.blackthm` profile modes.

## Purpose

The UI profile keeps future inline UI intent short and deterministic without making the compiler guess which value belongs to which style property.

## Profile Syntax

```blackthm
profile UICompact {
  version 1
  mode box color width style pt pr pb pl radius place
  mode text color size weight align
  mode table color width style density zebra
  mode button bg color radius size variant
}
```

Each `mode` line defines a left-to-right slot order.

## Planned Inline Syntax

```black
ui <mode> <values...> [| <mode> <values...>...]
```

Example:

```black
ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
```

With the profile above, the first group maps to:

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
```

The second group maps to:

```text
text.color  = "#172026"
text.size   = 14
text.weight = regular
text.align  = left
```

## Rules

- Slots are positional and read left to right.
- `|` separates multiple UI mode groups.
- Missing trailing values use defaults in later CSS generation phases.
- Extra values are errors.
- Each slot name may appear only once inside a mode.
- After a profile is locked, existing slots are immutable.
- After a profile is locked, new slots are append-only.

## CLI

```bash
black theme inspect --json
black docs ui-profile --json
```

AI agents should use `profile.rules` and `profile.modes[].slots` from `black theme inspect --json` before writing future inline UI intent.
