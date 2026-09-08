# BlackLang Provider Adapter Marketplace

The provider adapter marketplace is a local, deterministic manifest layer for adapter discovery and trust decisions.

It does not execute provider CLIs, publish packages, create cloud apps, or send traffic to external services by itself. It records stable adapter IDs, provider values, docs, capabilities, source ownership, validation commands, and trust requirements.

```bash
black ecosystem --json
black docs adapter-marketplace --json
node packages/registry/scripts/validate-registry.mjs
node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict
```

Authoritative files:

- `adapters/marketplace/adapter-index.blackdir`
- `adapters/marketplace/trust-policy.blackdir`
- `packages/registry/public-index.blackdir`
- `website/ecosystem-index.json`

Current provider-facing adapter IDs:

| ID | Kind | Provider values | Status |
|---|---|---|---|
| `deploy-cloud:docker-provider` | cloud deploy metadata and explicit apply runner | `fly`, `render`, `railway` | built-in |
| `observability:webhook` | observability hook | `webhook` | built-in |
| `observability:otlp` | observability trace exporter | `otlp` | built-in |
| `editor:vscode` | editor bridge | `vscode` | packageable source |
| `editor:open-vsx` | editor package channel | `openvsx` | packageable source |
| `editor:cursor-compatible` | editor package channel | `cursor-compatible-vsix` | packageable source |
| `target:web-react-node-mysql` | generated web/API target | `mysql` | built-in |

Trust rules:

- Provider credentials, app names, regions, endpoints, DSNs, and tokens stay outside `.black` source.
- Adapter execution must start with read-only manifest, plan, or preflight output before any external mutation.
- Adapter packages must use detached Ed25519 release trust verification before public marketplace publishing.
- Editor package channels must reuse compiler-owned IDE output and the same signed VSIX artifact instead of forking syntax rules.
- Hosted adapter index snapshots must be generated from compiler-owned ecosystem metadata and validated before deployment.
- Marketplace entries must include stable ID, docs, capabilities, status, source ownership, trust flags, and validation commands.
- Provider-specific behavior should stay in adapters/extensions until parser, validator, docs, JSON/BlackIR, diagnostics, affected graph, tests, and compact learning docs are ready.

Generated web apps currently use this model through:

```black
deploy {
  target docker
  cloud fly app env FLY_APP_NAME region env FLY_REGION
}

ops {
  observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT
}
```

These declarations produce local metadata and tests. `observe webhook` emits request-event JSON, while `observe otlp` emits OTLP HTTP JSON trace payloads with W3C `traceparent` context. Deploy cloud declarations also generate `scripts/cloud-exec.mjs`; its preflight mode is read-only and its exec mode requires explicit apply before invoking the provider CLI.
