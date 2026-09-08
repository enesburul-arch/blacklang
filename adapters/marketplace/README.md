# BlackLang Provider Adapter Marketplace

This directory contains the local provider adapter marketplace index for BlackLang web target work.

The marketplace is manifest-first. It describes built-in and packageable adapters by stable ID, source ownership, provider values, docs, capabilities, and trust rules. It does not publish packages or mutate cloud infrastructure by itself.

Current provider-facing entries include:

- `deploy-cloud:docker-provider`: Docker image deployment plan metadata plus read-only preflight and explicit apply runner metadata for `fly`, `render`, and `railway` provider values.
- `observability:webhook`: env-referenced, non-blocking request event delivery for generated web apps.
- `observability:otlp`: env-referenced, non-blocking OTLP HTTP JSON trace export for generated web apps.
- `editor:vscode`: packageable compiler-backed VS Code bridge.

Validation:

```bash
node packages/registry/scripts/validate-registry.mjs
```

Provider credentials, regions, endpoints, DSNs, and tokens stay outside `.black` source. Adapter execution must start with read-only manifest/plan/preflight checks before any external mutation, and provider CLI apply mode must be explicit.
