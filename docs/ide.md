# BlackLang IDE Contract

BlackLang IDE support is compiler-owned. Editors and language-server bridges should read metadata from the CLI instead of copying keyword, snippet, or diagnostic lists by hand.

```bash
black ide --json
black ide --ir
black ide diagnostics examples/warehouse/app.black --json
black ide diagnostics examples/warehouse/app.black --ir
```

`black ide --json` returns:

- `language`: language id, source extensions, language version, and comment prefix
- `capabilities`: diagnostic command, completion/snippet support, diagnostic catalog support, and zero-based range metadata
- `completionItems`: deterministic keyword, type, modifier, relation-load scope, view component bind, operator, deploy, and view value completions
- `snippets`: canonical snippets for common declarations such as app/target, entity, relation-load, query, page, view component section, action, API, seed, test, deploy, cloud deploy plan/preflight metadata, and ops observability
- `diagnosticCodes`: stable compiler diagnostic codes gathered from local docs entries

`black ide diagnostics <file> --json` reads one `.black` source file and reports editor-friendly diagnostics with zero-based, end-exclusive ranges. It checks deterministic formatting, parser diagnostics, validator diagnostics, and source-security findings. A readable but invalid source returns `success: true` and `valid: false`; command-level failures such as missing files return `success: false`.

Use this as the MVP integration point for VS Code, Open VSX/Cursor-compatible package channels, LSP bridges, and AI coding agents. Packaged editor extensions should consume this contract instead of reimplementing BlackLang syntax rules.

## VS Code Bridge

The repository includes a packageable VS Code extension source at:

```text
editors/vscode-blacklang/
```

It contributes `.black` and `.blackthm` language ids, TextMate syntax highlighting, compiler-backed completion items, compiler-backed snippets, and diagnostics from `black ide diagnostics`.

The deploy snippet includes `cloud fly app env FLY_APP_NAME region env FLY_REGION`. The ops snippet includes `observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT`, and completion metadata still exposes both `webhook` and `otlp` observe provider values. Both keep provider values and external endpoints outside `.black` source.

The extension also adds early refactor/repair workflows:

- `BlackLang: Refresh Diagnostics`
- `BlackLang: Show IDE Manifest`
- `BlackLang: Inspect Affected Symbol`
- `BlackLang: Format Source`
- quick fix for `FORMAT_REQUIRED`

Validate the extension source without launching VS Code:

```bash
cd editors/vscode-blacklang
npm test
```

Package it locally when the VS Code packaging tool is available:

```bash
npm run package:vsix
```

Multi-editor package channel metadata is prepared under `packages/registry/package-index.blackdir` and `adapters/marketplace/adapter-index.blackdir`. Marketplace publication remains an explicit release-owner action after signed release, transparency, and key-rotation checks pass.
