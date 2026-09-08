# BlackLang Ecosystem Discovery

`black ecosystem` reports deterministic release, package, registry, adapter, marketplace, public index, extension, and trust-policy metadata.

```bash
black ecosystem --json
black ecosystem --ir
```

Use it before adding provider-specific behavior or package/editor integrations.

The JSON result includes:

- `release`: release channel, manifest file, checksum file, release scripts, signed release trust metadata, release transparency log policy, and key rotation policy
- `packages`: local packageable sources and channels such as `packages/npm`, `vscode:blacklang-vscode`, `openvsx:blacklang-vscode`, and `cursor:blacklang-vscode`
- `adapters`: built-in target, deploy, cloud deploy, observability, and editor adapter metadata
- `registries`: prepared package registry manifests such as `packages/registry/package-index.blackdir`
- `marketplaces`: prepared provider adapter marketplace manifests such as `adapters/marketplace/adapter-index.blackdir`
- `publicIndex`: prepared static index metadata for `website/ecosystem-index.json` before public hosting
- `trustWorkflow`: policy files, required validation checks, and install/publish review steps
- `policies`: rules for keeping provider-specific behavior in extensions/adapters until it is ready to become official syntax

Current package registry metadata lives in:

```text
packages/registry/package-index.blackdir
packages/registry/trust-policy.blackdir
packages/registry/release-transparency.blackdir
packages/registry/key-rotation-policy.blackdir
packages/registry/key-revocations.blackdir
packages/registry/public-index.blackdir
website/ecosystem-index.json
```

Current provider adapter marketplace metadata lives in:

```text
adapters/marketplace/adapter-index.blackdir
adapters/marketplace/trust-policy.blackdir
```

Validate the manifests with:

```bash
node packages/registry/scripts/validate-registry.mjs
```

Verify a prepared release directory before public install paths are trusted:

```bash
node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict
```

The release trust contract uses detached Ed25519 signatures named `<artifact>.sig`. The verifier reads the public key from `BLACKLANG_RELEASE_PUBLIC_KEY` or `BLACKLANG_RELEASE_PUBLIC_KEY_FILE`; private signing keys stay outside the repository, `.black` files, manifests, and package metadata. `release.trust.transparencyLog` exposes the append-only `transparency.blackdir` entry shape and hash-chain fields. `release.trust.keyRotation` exposes the public key id format, 30 day overlap policy, and revocation manifest path.

Current built-in cloud deploy metadata is exposed through `deploy-cloud:docker-provider`. It supports `fly`, `render`, and `railway` as provider values for generated Docker image deployment plans, read-only preflight checks, and explicit apply-mode provider CLI runners. Generated apps declare this with:

```black
deploy {
  target docker
  cloud fly app env FLY_APP_NAME region env FLY_REGION
}
```

Current external observability metadata is exposed through `observability:webhook` and `observability:otlp`. Generated apps declare this with:

```black
ops {
  observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT
}
```

Use one `observe` line per `ops` block; `webhook` is the request-event alternative. Deploy cloud declarations produce local plan metadata, read-only preflight output, and an explicit apply runner. Observability declarations produce local metadata, W3C traceparent context, generated smoke-test coverage, and an `ops/observability.json` manifest. Provider secrets, app names, regions, endpoints, DSNs, and tokens stay in environment variables or provider tooling.

The npm wrapper at `packages/npm` is intentionally thin. It resolves a trusted native BlackLang binary and forwards arguments instead of reimplementing the language in JavaScript.

The editor bridge at `editors/vscode-blacklang` is also thin. It consumes `black ide --json` and `black ide diagnostics <file> --json`. Package registry metadata prepares VS Code, Open VSX, and Cursor-compatible VSIX channels from that same bridge; external publication remains a release-owner action.

The public ecosystem index snapshot at `website/ecosystem-index.json` is generated from `black ecosystem --json`. It is a prepared source for future hosted package/provider adapter discovery, not a deployment or an external marketplace publish by itself.

Validation commands:

```bash
node packages/registry/scripts/validate-registry.mjs

cd packages/npm
npm test

cd editors/vscode-blacklang
npm test

black ecosystem --json
```

Future provider adapters should follow the same discipline: parser, validator, docs, JSON/BlackIR, diagnostics, affected analysis, tests, compact AI learning docs, manifest metadata, signed package verification, and trust checks before syntax is treated as official or before external systems are mutated.
