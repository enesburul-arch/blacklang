# Generated Contract and Browser E2E Tests

BlackLang generated web and API-only apps include deterministic smoke tests for the generated contract and API surface. `target web` also includes React render checks, declared browser-check expectations, explicit browser e2e execution, and browser e2e matrix plan/run commands when source `test` declarations exist.

After building a project, run this inside the generated app directory:

```bash
npm test
```

If `blacklang.toml` sets a custom `out` directory, run the command in that generated output directory. When the source declares seeds, also run `npm run db:setup` to prove schema setup and deterministic fixture upserts work together.

When source `test` declarations exist in a `target web` project, also run:

```bash
npm run test:e2e
npm run test:e2e:plan
npm run test:e2e:matrix
```

The generated `test:e2e` command runs `db:generate`, launches an installed Chrome/Chromium/Edge executable through `playwright-core`, starts the generated API server and Vite dev server on random localhost ports, registers a deterministic auth user when auth is enabled, and checks declared text/page/action expectations in the real DOM. It uses an ephemeral SQLite database by default so local e2e runs do not mutate `dev.db`.

The generated `test:e2e:plan` command is read-only. It prints JSON describing custom, Chrome, Edge, and Chromium target availability without launching browsers or touching the database. `test:e2e:matrix` runs the same generated e2e intent once per available target and reports passed/failed/skipped counts plus compact stdout/stderr tails.

`target api` projects do not emit React, frontend smoke, browser-check, browser e2e, or browser matrix tests. Their generated `npm test` script runs `npm run db:generate`, then only `src/blacklang.contract.test.ts` and `src/blacklang.api.test.ts`.

The generated test files are:

```text
src/blacklang.contract.test.ts
src/blacklang.api.test.ts
src/blacklang.frontend.test.tsx   # target web only
src/blacklang.browser.test.tsx    # target web only, when source declares test blocks
src/blacklang.e2e.test.ts         # target web only, when source declares test blocks
src/blacklang.e2e.matrix.ts       # target web only, when source declares test blocks
tests/browser-matrix.json         # target web only, when source declares test blocks
```

They check:

- `openapi.json` version and schemas
- page CRUD paths
- bound custom query paths
- bound custom query summary paths and aggregate metadata
- background query job metadata in `x-blacklang-jobs` when jobs are declared
- bound custom action paths and `x-blacklang-action`/`x-blacklang-transaction` metadata
- declared ops paths and `x-blacklang-ops` metadata
- explicit API contract paths, typed body schemas, `x-blacklang-runtime`, and `x-blacklang-handler` metadata
- service-bound explicit API `x-blacklang-service`, OpenAPI tags, root `x-blacklang-services`, and generated `services/manifest.json` metadata
- generated entity validation functions
- generated custom action input validation functions
- real HTTP serving of `/openapi.json`
- real HTTP serving of generated explicit API declared-runtime routes
- generated explicit API handler body validation before mutation
- HTTP health, readiness, and metrics probes when declared
- anonymous API rejection when auth is enabled
- JSON 404 behavior when auth is not enabled
- generated CORS allow/deny behavior when `security.cors` exists
- generated seed runtime wiring in `package.json` when seed declarations exist
- generated `jobs:run` and `jobs:loop` package scripts when job declarations exist
- server-side rendering of the generated React `App` in `target web`
- generated browser-check expectations from top-level `test` declarations when they exist in `target web`
- real browser auth/register, page navigation, DOM text, and action button checks through `npm run test:e2e` in `target web`
- read-only browser e2e matrix plan metadata and cross-browser e2e execution through `npm run test:e2e:plan` and `npm run test:e2e:matrix` in `target web`

This is generated output. Do not edit it manually. Change the `.black` source or the compiler generator.

The top-level `test` declaration is the current first-class BlackLang browser test syntax. It requires `target web`; `target api` reports `UNSUPPORTED_API_TARGET_TEST` because no browser runtime is generated. The generated fast browser-check smoke test stays in `npm test`; full browser e2e remains explicit as `npm run test:e2e`, and cross-browser coverage uses generated `test:e2e:plan` plus `test:e2e:matrix`. These runs stay bounded to generated auth, setup/seed, navigation, text, page, and action checks.

Background query jobs are checked by contract metadata inside `npm test`. Run `npm run jobs:run` separately after `black build` when the source declares jobs; this executes the generated worker once and logs compact JSON metadata.
