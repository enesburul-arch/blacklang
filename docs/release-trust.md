# BlackLang Release Trust

Release trust defines the checks that must pass before public install paths are trusted. It now also records release transparency log and public key rotation policy metadata for AI/CI review.

It covers CLI release archives, npm wrapper downloads, editor extension packages, and provider adapter packages.

```bash
black ecosystem --json
black docs release-trust --json
node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict
```

The current trust contract is:

- `release.blackdir` lists every finalized archive and its SHA-256 hash.
- `checksums.sha256` contains the same archive/hash pairs.
- Every archive has one detached Ed25519 signature named `<artifact>.sig`.
- The verifier reads the trusted public key from `BLACKLANG_RELEASE_PUBLIC_KEY` or `BLACKLANG_RELEASE_PUBLIC_KEY_FILE`.
- `transparency.blackdir` should contain one append-only entry per public archive with release, channel, artifact, sha256, signature, keyId, previousEntryHash, entryHash, and publishedAt fields.
- Key rotation uses `sha256-public-key-spki-prefix` key ids, a 30 day old/new key overlap policy, and `key-revocations.blackdir` for revoked public key ids.
- Private signing keys must never be stored in `.black` source, generated output, package metadata, registry manifests, or adapter manifests.

`scripts/verify-release-trust.mjs` is read-only. It reports:

- release directory and version
- release channel and CLI command
- each artifact from `release.blackdir`
- whether the archive exists
- whether `release.blackdir`, `checksums.sha256`, and actual archive bytes agree
- whether the detached `.sig` file exists
- whether Ed25519 verification passed when a public key is present
- missing files or trust inputs

Use non-strict mode while preparing a release:

```bash
node scripts/verify-release-trust.mjs artifacts/releases/v0.2.0 --json
```

Use strict mode before publishing:

```bash
node scripts/verify-release-trust.mjs artifacts/releases/v0.2.0 --json --strict
```

Strict mode exits non-zero unless `ready` is `true`. Run `node packages/registry/scripts/validate-registry.mjs` as the companion policy check so release transparency and key rotation manifest requirements are also verified before publishing.

This phase does not create signatures, publish a log, rotate keys, or read private keys. Signing and public log publication remain release-owner actions outside the repository. BlackLang defines the deterministic verification contract and safe metadata that AI agents, CI, npm wrapper code, editor packages, and provider adapter tooling can inspect.
