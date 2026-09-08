# BlackLang Theme Migration

`black theme migrate` checks whether a new `.blackthm` file can safely replace an old one.

It is read-only. It does not rewrite theme files or generated output.

## Command

```bash
black theme migrate <old.blackthm> <new.blackthm> --json
black theme migrate <old.blackthm> <new.blackthm> --ir
```

## Why It Exists

Inline UI values are positional:

```black
ui box black 1 solid 8 8 5 5 6 center
```

The active profile decides what each value means:

```blackthm
ui box = color width style pt pr pb pl radius place;
```

Changing that slot order can silently remap old source intent. The migration command makes that risk machine-readable before a theme is replaced.

## Safe Changes

Safe migration keeps existing positional meaning:

- Theme name stays the same.
- Theme target stays the same.
- Theme and profile versions do not go backward.
- A locked profile stays locked.
- Profile name stays the same.
- Every old UI mode still exists.
- Every old slot sequence remains the exact prefix of the new slot sequence.
- New slots are appended at the end.

Example safe change:

```blackthm
ui box = color width style;
ui box = color width style shadow;
```

The JSON result includes `safe: true` and a `slot-appended` change.

## Unsafe Changes

This is unsafe:

```blackthm
ui box = color width style;
ui box = color shadow width style;
```

The old second value used to mean `width`; after insertion it would mean `shadow`. The command reports `UI_SLOT_MIGRATION_BREAK`.

Removing a mode reports `UI_MODE_REMOVED`. Changing the theme name, target, profile name, lock state, or lowering versions also reports stable diagnostics.

## AI Agent Rule

Before replacing a theme used by existing `.black` files:

```bash
black theme inspect old.blackthm --json
black theme inspect new.blackthm --json
black theme migrate old.blackthm new.blackthm --json
```

Proceed only when `success` and `safe` are both `true`.
