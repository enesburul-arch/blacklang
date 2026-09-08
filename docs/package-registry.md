# BlackLang Package Registry

The package registry is a local, deterministic manifest layer for package publishing work.

It does not publish to npm, the VS Code Marketplace, GitHub Releases, or any external registry by itself. It records which packageable sources exist, which docs explain them, which commands validate them, and which trust checks must pass before publication.

```bash
black ecosystem --json
black docs package-registry --json
node packages/registry/scripts/validate-registry.mjs
node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict
```

Authoritative files:

- `packages/registry/package-index.blackdir`
- `packages/registry/trust-policy.blackdir`
- `packages/registry/release-transparency.blackdir`
- `packages/registry/key-rotation-policy.blackdir`
- `packages/registry/key-revocations.blackdir`
- `packages/registry/public-index.blackdir`
- `packages/registry/scripts/validate-registry.mjs`
- `website/ecosystem-index.json`
- `scripts/verify-release-trust.mjs`

Current package IDs:

| ID | Kind | Source | Status |
|---|---|---|---|
| `npm:blacklang` | npm wrapper | `packages/npm` | local package source |
| `vscode:blacklang-vscode` | editor extension | `editors/vscode-blacklang` | packageable source |
| `openvsx:blacklang-vscode` | editor extension channel | `editors/vscode-blacklang` | packageable source |
| `cursor:blacklang-vscode` | editor extension channel | `editors/vscode-blacklang` | packageable source |

Trust rules:

- Release artifacts must have `release.blackdir` and `checksums.sha256` before public package publication.
- Release artifacts must have detached Ed25519 `<artifact>.sig` signatures and pass `scripts/verify-release-trust.mjs` before public install paths are trusted.
- Public package publishing also requires release transparency policy and key rotation policy metadata from `packages/registry/release-transparency.blackdir` and `packages/registry/key-rotation-policy.blackdir`.
- The npm wrapper must remain thin and forward to the native BlackLang CLI.
- `BLACKLANG_ALLOW_DOWNLOAD=1` is required before the npm wrapper may download release artifacts.
- `BLACKLANG_BINARY` can point to a trusted local CLI binary.
- Editor bridges must consume compiler-owned `black ide` and `black ide diagnostics` output instead of reimplementing language behavior.
- VS Code, Open VSX, and Cursor-compatible editor channels share the same packageable bridge and signed VSIX trust workflow.
- Public index snapshots must be generated from `black ecosystem --json`, validated locally, and deployed only after release-owner review.

Before changing package publishing behavior, run:

```bash
node packages/registry/scripts/validate-registry.mjs
node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict
cd packages/npm && npm test
cd editors/vscode-blacklang && npm test
```
