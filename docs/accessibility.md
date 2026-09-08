# BlackLang Accessibility Audit

`black audit accessibility` is a read-only policy check for generated UI intent.

It exists so AI agents can catch source-level accessibility risks before editing generated React or CSS.

## Syntax

```bash
black audit accessibility [file] [--json|--ir]
```

When the file argument is omitted, the CLI reads `blacklang.toml` and audits the configured `source`.

## Current Checks

Draft v0.2 checks these generated UI policies:

- `section detail|form display modal|drawer` should declare `title "..."` so generated dialogs have stable accessible names.
- `group <Name> sections ...` should declare `title "..."` when it wraps multiple sections so generated nested landmarks are easy to understand.

The command parses and validates source first. If source is unreadable or semantically invalid, diagnostics are returned in `errors` and accessibility findings are not guessed from invalid AST state.

## JSON Shape

```json
{
  "success": false,
  "command": "audit accessibility",
  "version": "0.1.0-dev",
  "file": "examples/warehouse/app.black",
  "summary": {
    "app": "Warehouse",
    "entities": 3,
    "pages": 4
  },
  "findings": [
    {
      "file": "examples/warehouse/app.black",
      "line": 42,
      "column": 1,
      "code": "ACCESSIBILITY_MISSING_OVERLAY_TITLE",
      "message": "Page Orders view section form uses display modal without an explicit title.",
      "suggestion": "Add title \"...\" to the overlay section so generated dialogs have stable accessible names."
    }
  ],
  "errors": []
}
```

## AI Agent Notes

Run:

```bash
black audit accessibility --json
```

after changing page view composition, modal/drawer section display, groups, triggers, forms, or page actions.

Fix findings by editing `.black` source, then rerun:

```bash
black audit accessibility --json
black validate --json
black build
```

Do not fix accessibility policy findings by editing generated files.

## Diagnostics

```text
ACCESSIBILITY_MISSING_OVERLAY_TITLE
ACCESSIBILITY_MISSING_GROUP_TITLE
UNKNOWN_AUDIT_COMMAND
FILE_READ_ERROR
UNCLOSED_STRING
UNEXPECTED_CHARACTER
MISSING_APP
```
