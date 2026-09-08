# Editor Marketplace Channels

BlackLang editor packages are compiler-owned bridges. They must call `black ide --json` and `black ide diagnostics <file> --json` instead of copying parser, validator, completion, snippet, or diagnostic rules.

This repository prepares package metadata for multiple editor channels. It does not publish to any external marketplace by itself.

```bash
black ecosystem --json
black docs editor-marketplace --json
node packages/registry/scripts/validate-registry.mjs
cd editors/vscode-blacklang && npm test
cd editors/vscode-blacklang && npm run package:vsix
```

Current editor package channel IDs:

| ID | Channel | Source | Status |
|---|---|---|---|
| `vscode:blacklang-vscode` | VS Code Marketplace | `editors/vscode-blacklang` | packageable source |
| `openvsx:blacklang-vscode` | Open VSX | `editors/vscode-blacklang` | packageable source |
| `cursor:blacklang-vscode` | Cursor-compatible VSIX | `editors/vscode-blacklang` | packageable source |

Current editor adapter IDs:

| ID | Provider | Source | Status |
|---|---|---|---|
| `editor:vscode` | `vscode` | extension | packageable source |
| `editor:open-vsx` | `openvsx` | extension | packageable source |
| `editor:cursor-compatible` | `cursor-compatible-vsix` | extension | packageable source |

All channels reuse the same packageable bridge and release trust workflow. Public publication requires:

- `release.blackdir`
- `checksums.sha256`
- detached Ed25519 signatures for release archives and package artifacts
- `node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict`
- release transparency metadata
- release key rotation metadata
- no private signing keys, tokens, or credentials in `.black`, generated output, manifests, or package metadata

Add new editor channels as registry and adapter metadata first. Keep provider-specific publishing tokens outside the repository and require an explicit release-owner publish step.
