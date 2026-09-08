# Secret Reference Manifest

BlackLang source must reference secrets by environment name. Generated apps now make those references explicit without storing values.

Generated output includes:

```text
security/secrets.json
scripts/secrets-plan.mjs
scripts/secrets-provider.mjs
npm run security:secrets:plan
npm run security:secrets:preflight
```

`security/secrets.json` lists each environment-backed reference with its source, kind, required flag, and whether the value is sensitive. It records names such as `DATABASE_URL`, `CORS_ORIGINS`, cloud app/region env names, and observability endpoint env names, but never records the values.

`npm run security:secrets:plan` reads the manifest and current process environment, then prints a JSON readiness report:

- `ready`
- `missingRequired`
- `references[].present`
- `references[].providerKey`
- `policy`

The plan output never prints secret values.

`npm run security:secrets:preflight` executes the generated provider boundary in read-only preflight mode. It verifies the selected provider label, provider prefix, and provider CLI availability for `1password`, `vault`, `doppler`, or `aws-secrets-manager`. It does not fetch, print, write, inject, or transform secret values. For the default `environment` provider, it reports required environment presence directly.

Optional provider handoff is environment-driven:

```bash
BLACKLANG_SECRET_PROVIDER=vault
BLACKLANG_SECRET_PREFIX=Warehouse
npm run security:secrets:plan
npm run security:secrets:preflight
```

Supported provider labels for the read-only plan and preflight are `environment`, `1password`, `vault`, `doppler`, and `aws-secrets-manager`. Provider tooling remains outside generated source; inject values into the environment or mirror the manifest names under the reported provider keys before deployment. Provider preflight only checks provider CLI readiness; provider-owned runtime value injection belongs in a later adapter layer.

Use this with:

```bash
black security scan --json
black package --production
npm run security:secrets:plan
npm run security:secrets:preflight
```

Change `.black` source or generator code; do not manually edit generated manifests or scripts.
