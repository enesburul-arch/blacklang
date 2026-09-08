# BlackLang Package Registry Manifests

This directory contains local, deterministic registry manifests for BlackLang packages.

The manifests are not an external publish. They are the source of truth that AI agents, release scripts, and future package registry automation can inspect before publishing npm packages or editor extensions.

Files:

- `package-index.blackdir`: stable package IDs, source paths, docs, validation commands, and trust flags.
- `public-index.blackdir`: prepared public/hosted ecosystem index source metadata for package and provider adapter discovery.
- `trust-policy.blackdir`: required checks before public package registry publication.
- `release-transparency.blackdir`: append-only release transparency log shape and hash-chain requirements.
- `key-rotation-policy.blackdir`: public key id, overlap, revocation, and private-key boundary rules.
- `key-revocations.blackdir`: append-only revoked release public key id list.
- `scripts/validate-registry.mjs`: local manifest validator.
- `../../scripts/verify-release-trust.mjs`: read-only release checksum and detached signature verifier.

Validation:

```bash
node packages/registry/scripts/validate-registry.mjs
node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict
```

Package wrappers and editor bridges must stay thin. They call the native `black` CLI and must not reimplement parsing, validation, formatting, diagnostics, completions, snippets, or affected analysis.
