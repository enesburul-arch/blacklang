# BlackLang Deployment

BlackLang deployment intent describes how the generated web app should run in production.

```black
deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  env CORS_ORIGINS optional
  preview local
  rollback keep 3
  cloud fly app env FLY_APP_NAME region env FLY_REGION
}
```

Rules:

- Use one top-level `deploy` block.
- Draft v0.1 supports `target docker`.
- `port env NAME default PORT` makes the generated server read its listen port from the environment.
- `env NAME required|optional` documents runtime configuration without storing secrets in `.black` source.
- `preview local` asks the generator for an isolated local Docker Compose preview stack.
- `rollback keep COUNT` asks the generator for read-only rollback metadata that keeps the newest declared number of releases.
- `cloud PROVIDER app env NAME` asks the generator for cloud adapter metadata, read-only plan/preflight output, and an explicit provider CLI apply runner for `fly`, `render`, or `railway`.
- `region env NAME` is optional cloud adapter metadata for providers that need a deployment region.
- Environment variable names use uppercase letters, numbers, and underscores.
- Rollback keep counts must be between 1 and 20.
- Cloud provider app, region, tokens, and endpoints must stay in environment variables or provider tooling, not `.black` source.

Generated web behavior when `target docker` is declared:

- Writes `Dockerfile`.
- Writes `.dockerignore`.
- Writes `docker-compose.yml`.
- Adds `PORT` to `.env.example`.
- Adds a `start` script to `package.json`.
- Writes `security/secrets.json`, `scripts/secrets-plan.mjs`, and `scripts/secrets-provider.mjs`.
- Adds `security:secrets:plan` and `security:secrets:preflight` to `package.json`.
- Makes `src/server.ts` read the configured port environment variable.
- Serves the built Vite frontend from `dist` after API routes.
- When `ops { health path "/healthz" }` exists, adds Dockerfile and Docker Compose healthchecks that probe the declared health path.

Generated web behavior when `preview local` is declared:

- Writes `docker-compose.preview.yml`.
- Adds `BLACKLANG_PREVIEW_PORT` to `.env.example`.
- Adds `deploy:preview` and `deploy:preview:down` package scripts.
- Uses an isolated preview host port and preview data defaults so local preview does not share the normal app data volume by default.

Generated web behavior when `rollback keep COUNT` is declared:

- Writes `deploy/manifest.json`.
- Writes `deploy/rollback.json`.
- Writes `scripts/rollback-plan.mjs`.
- Adds `deploy:rollback:plan` to `package.json`.
- The generated rollback plan script is read-only. It reports retained release directories and the current rollback candidate without changing infrastructure.

Generated web behavior when `cloud PROVIDER app env NAME` is declared:

- Writes `deploy/manifest.json` with cloud adapter metadata.
- Writes `deploy/cloud.json`.
- Writes `scripts/cloud-plan.mjs` and `scripts/cloud-exec.mjs`.
- Adds `deploy:cloud:plan`, `deploy:cloud:preflight`, and `deploy:cloud:exec` to `package.json`.
- Adds the cloud app and region env names to `.env.example`, `docker-compose.yml`, and local preview compose output.
- The generated cloud plan script is read-only. It reports missing provider env values, provider CLI metadata, auth env names, optional env names, and planned adapter steps without changing infrastructure.
- The generated cloud preflight script is read-only. It checks required env and provider CLI availability without changing infrastructure.
- The generated cloud exec script requires explicit apply mode through `npm run deploy:cloud:exec`; it invokes the selected provider CLI and redacts configured environment values from captured output.

Generated provider CLI templates:

- `fly`: `fly deploy . --app ${APP_ENV} --yes`, with `flyctl` accepted as a fallback executable.
- `render`: `render deploys create ${APP_ENV} --wait --confirm -o json`, with optional `BLACKLANG_CLOUD_IMAGE` and `BLACKLANG_CLOUD_COMMIT` arguments when those env values are present.
- `railway`: `railway up . --project ${APP_ENV} --ci --yes --json`, with optional `BLACKLANG_CLOUD_SERVICE` and `BLACKLANG_CLOUD_MESSAGE` arguments when those env values are present.

SQLite, PostgreSQL, and MySQL targets are supported by generated runtime output. PostgreSQL Docker Compose output includes a PostgreSQL service with its own service healthcheck. MySQL Docker Compose output includes a MySQL 8.4 service, `MYSQL_*` local env defaults, a service healthcheck, and an app database URL fallback.

When the source declares `migration` blocks, generated deployment output also includes `src/migrate.ts`, `db:migrate:plan`, and `db:migrate`. Run `npm run db:migrate:plan` in the target environment before mutating an existing database. Run `npm run db:migrate` only after the plan is ready, then run `npm run db:setup` so the normal generated setup and seed path runs. Migration plan output is JSON and redacts database connection values. Generated rename migration runtime currently supports SQLite and PostgreSQL; `target database mysql` reports `UNSUPPORTED_TARGET_DATABASE_MIGRATION` when migration blocks are present.

Generated secret reference behavior:

- `security/secrets.json` lists required and optional environment references without values.
- `npm run security:secrets:plan` reports missing required environment references without printing secret values.
- `npm run security:secrets:preflight` checks the selected provider label, prefix, and provider CLI availability without fetching or printing secret values.
- `BLACKLANG_SECRET_PROVIDER` and `BLACKLANG_SECRET_PREFIX` label external provider handoff in the read-only plan and preflight.

Use `docs/ops.md` or `black docs ops --json` when you need runtime health, readiness, metrics, request logging, external observability hooks, trace context, or OTLP exporter metadata.
