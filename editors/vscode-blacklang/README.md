# BlackLang VS Code Extension

This extension is a thin IDE bridge over the BlackLang compiler.

It does not reimplement the language. It reads:

```bash
black ide --json
black ide diagnostics <file> --json
```

Features:

- `.black` and `.blackthm` syntax highlighting
- compiler-owned completion items
- compiler-owned canonical snippets
- editor diagnostics with zero-based ranges
- quick fix for `FORMAT_REQUIRED` through `black format`
- diagnostic info quick actions
- affected-symbol inspection through `black inspect --affected`

Set `blacklang.cliPath` when the `black` executable is not on `PATH`.

Package locally with:

```bash
npm run package:vsix
```

The package command uses `@vscode/vsce` through `npx`. Marketplace publishing is intentionally outside this repository workflow.
