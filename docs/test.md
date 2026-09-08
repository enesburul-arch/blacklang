# Browser Test Declarations

`test` declares deterministic generated browser-check and browser e2e expectations for one page in `target web` projects.

The source syntax is intentionally small. `npm test` checks generated page metadata, generated action lists, and generated UI text/render catalogs after `black build`. When `test` declarations exist, generated web output also emits explicit browser e2e commands: `npm run test:e2e` for one selected Chromium-based browser, `npm run test:e2e:plan` for a read-only matrix report, and `npm run test:e2e:matrix` for every available matrix target. `target api` does not generate a browser runtime, so top-level `test` declarations report `UNSUPPORTED_API_TARGET_TEST`.

## Syntax

```black
test <Name> {
  page <PageName>
  expect text "Generated text"
  expect page <PageName>
  expect action <actionName>
}
```

Exactly one `page` line is required. At least one `expect` line is required.

Supported expectations:

- `expect text "..."` checks that the literal appears in the generated React server-rendered HTML or generated browser-check text catalog.
- `expect page PageName` checks that a declared page exists in generated page metadata.
- `expect action actionName` checks that the target page exposes that CRUD or custom action name.

## Example

```black
test WarehouseBrowserSmoke {
  page Products
  expect text "Depo"
  expect page LowStock
  expect action RestockProduct
}
```

When at least one `test` declaration exists in a `target web` project, generated output includes:

```text
src/blacklang.browser.test.tsx
src/blacklang.e2e.test.ts
src/blacklang.e2e.matrix.ts
tests/browser-matrix.json
```

The generated `package.json` appends the browser-check file to `npm test` after the contract, API, and frontend smoke tests. It also adds:

```bash
npm run test:e2e
npm run test:e2e:plan
npm run test:e2e:matrix
npm run test:all
```

`test:e2e` runs `db:generate`, creates an ephemeral SQLite database by default, runs generated setup/seed modules when present, starts the generated API server and Vite dev server on random localhost ports, registers a deterministic auth user when auth is enabled, then checks declared text/page/action expectations in the real browser DOM.

`test:e2e:plan` is read-only. It reports JSON with matrix target availability for custom, Chrome, Edge, and Chromium candidates without launching the app or mutating the database. `test:e2e:matrix` runs the same generated e2e intent once for each available browser target and returns compact JSON with passed/failed/skipped counts plus stdout/stderr tails for triage. Missing targets are skipped unless no browser is available.

Set `BLACKLANG_E2E_BROWSER_PATH` for a custom Chromium-based executable. Set `BLACKLANG_E2E_CHROME_PATH`, `BLACKLANG_E2E_EDGE_PATH`, or `BLACKLANG_E2E_CHROMIUM_PATH` for explicit matrix targets. Set `BLACKLANG_E2E_DATABASE_URL` when a non-default e2e database is required.

## AI workflow

Use these commands before and after editing browser-check intent:

```bash
black docs test --json
black explain test --json
black inspect examples/warehouse/app.black --affected WarehouseBrowserSmoke --json
black format --check --json examples/warehouse/app.black
black lint examples/warehouse/app.black --json
black validate examples/warehouse/app.black --json
black build examples/warehouse/app.black --out generated --json
cd generated
npm run build
npm test
npm run test:e2e:plan
npm run test:e2e:matrix
```

## Diagnostics

Common repair paths:

- `MISSING_TEST_PAGE`: add one `page PageName` line.
- `UNKNOWN_TEST_PAGE`: use a declared page name.
- `MISSING_TEST_EXPECT`: add at least one deterministic expectation.
- `UNKNOWN_TEST_EXPECT_PAGE`: use a declared page name in `expect page`.
- `UNKNOWN_TEST_EXPECT_ACTION`: use a CRUD or custom action exposed by the target page.
- `DUPLICATE_TEST_EXPECT`: remove repeated identical expectations.
- `UNSUPPORTED_TEST_EXPECT`: use only `text`, `page`, or `action`.
- `UNSUPPORTED_API_TARGET_TEST`: use `target web` for browser test declarations or remove top-level `test` blocks from API-only sources.

See [diagnostics.md](diagnostics.md) for the complete stable code list.
