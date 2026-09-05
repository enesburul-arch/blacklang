# BlackLang UI Profile Rules

This document defines the compact positional UI slot rules used by `.blackthm` profile modes.

## Purpose

The UI profile keeps inline UI intent short and deterministic without making the compiler guess which value belongs to which style property.

## Profile Syntax

```blackthm
profile UICompact {
  version 1
  baseline box color width style pt pr pb pl radius place
  baseline text color size weight align
  baseline table color width style density zebra
  baseline button bg color radius size variant

  ui box = color width style pt pr pb pl radius place;
  ui text = color size weight align;
  ui table = color width style density zebra;
  ui button = bg color radius size variant;
}
```

Each `ui <mode> = <slot...>;` line defines a left-to-right generator slot order.

Legacy `mode <name> <slot...>` lines are still accepted for backward compatibility.

Each `baseline` line defines the frozen prefix for the matching mode after the theme has `locked true`.

## Inline Syntax

```black
ui <mode> <values...> [| <mode> <values...>...]
```

Example:

```black
form {
  fields email, password
  ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
}
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
- `ui <mode> = <slot...>;` is the preferred profile syntax for generator reading order.
- `|` separates multiple UI mode groups.
- Missing trailing values use CSS generation defaults.
- Extra values are errors.
- Each slot name may appear only once inside a mode.
- After a profile is locked, existing slots are immutable.
- After a profile is locked, new slots are append-only.
- Locked profiles require one baseline for each current mode.
- A locked mode must start with the exact baseline slot sequence.
- Web profiles must include standard `box`, `text`, `table`, and `button` mode groups.

## Standard Mode Groups

```text
box     container border, spacing, radius, and placement
text    typography for labels, headings, helper text, and body copy
table   table-specific border, density, and row pattern styling
button  action control styling for generated and explicit buttons
```

`black theme inspect --json` returns these groups in `profile.modeGroups`.

## Locked Append-Only Example

This is valid because `shadow` is appended after the baseline:

```blackthm
blackthm WarehouseTheme {
  version 2
  locked true

  profile UICompact {
    version 2
    baseline box color width style
    ui box = color width style shadow;
  }
}
```

This is invalid because `shadow` was inserted before existing slots:

```blackthm
blackthm WarehouseTheme {
  version 2
  locked true

  profile UICompact {
    version 2
    baseline box color width style
    ui box = color shadow width style;
  }
}
```

The compiler reports `NON_APPEND_ONLY_UI_SLOT`.

## CLI

```bash
black theme inspect --json
black docs ui --json
black docs ui-profile --json
```

AI agents should use `profile.rules`, `profile.modeGroups`, and `profile.modes[].slots` from `black theme inspect --json` before writing inline UI intent.
