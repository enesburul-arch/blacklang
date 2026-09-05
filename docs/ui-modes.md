# BlackLang UI Modes

This document defines the standard UI mode groups used by `.blackthm` web profiles.

## Purpose

Mode groups let BlackLang keep UI intent compact while still telling the compiler what kind of element is being styled.

## Standard Groups

```text
box     container border, spacing, radius, and placement
text    typography for labels, headings, helper text, and body copy
table   table-specific border, density, and row pattern styling
button  action control styling for generated and explicit buttons
```

## Profile Syntax

```blackthm
profile UICompact {
  version 1
  ui box = color width style pt pr pb pl radius place;
  ui text = color size weight align;
  ui table = color width style density zebra;
  ui button = bg color radius size variant;
}
```

Web profiles must include all four standard modes. Custom modes can be added later, but the standard modes keep common web UI work predictable for AI agents.

Legacy `mode <name> <slot...>` lines are still accepted, but `ui <mode> = <slot...>;` is the preferred generator order syntax.

## Inspect Output

`black theme inspect --json` returns `profile.modeGroups`:

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

Each parsed standard mode also includes `standard: true`, `purpose`, and `appliesTo`.

## Diagnostics

If a profile is missing a standard mode, the compiler reports:

```text
MISSING_STANDARD_UI_MODE
```

AI agents should read `profile.modeGroups` before writing inline UI intent so they can choose the correct group instead of guessing.

For inline `.black` placement rules, use:

```bash
black docs ui --json
```
