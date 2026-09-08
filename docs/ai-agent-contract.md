# BlackLang AI Agent Contract

This document defines how AI coding agents should work with BlackLang today.

It exists to prevent agents from inventing unsupported syntax or presenting a prototype as official BlackLang behavior.

## Official Source Path

Current official BlackLang projects use:

```text
.black source files
.black.enc protected source files
.blackthm theme/profile files
black CLI commands
generated web/API output
```

The official edit loop is:

```bash
black agent startup --json
black docs --all --json
black ide --json
black ide diagnostics app.black --json
black ecosystem --json
black inspect app.black --json
black format --check --json
black lint --json
black validate --json
black build
cd generated && npm test
black audit accessibility --json
black benchmark --json
black benchmark tasks --json
black benchmark eval --json
black benchmark eval-history --json
black security encrypted-source --json
black docs ops --json
```

If a repository includes `blacklang.toml`, use it to find the active source, output directory, target, and theme file. Run generated app commands such as `npm test` in the configured output directory. If the active source ends in `.black.enc`, set the key environment variable named in the encrypted header before running parse, lint, validate, inspect, benchmark, security scan, or build.

## Current Capability Boundary

BlackLang v0.2 is strongest for compact, deterministic web and API service intent:

```text
app
target web
target api
entity
relation fields
relation response load policies with generated batch prefetch
page
view order, DOM-order rendering, reusable component sections, collection-backed component sections, modal/drawer section display, nested section groups, and view interaction triggers
table
form
actions
auth
roles/access
workflow
state
component variants and page component sections
validation
local file/image media fields
computed display fields
entity indexes
schema migration planning
explicit entity/field rename migration runtime
first-class seed/fixture declarations and generated db:seed runtime
custom queries bound to pages
custom query aggregate summaries
background query jobs
custom actions bound to pages, including ordered local `value` declarations, compound `and`/`or`/`not` conditions, and indentation-based `if`/`else` branch statements
top-level transaction blocks for custom actions and explicit API update handlers
service blocks for explicit API module metadata
first-class browser-check and browser e2e test declarations
generated browser e2e matrix plan/run support
explicit API declared-runtime routes and bounded update handlers, including block-form `update` logic with local `value`, compound conditions, and `if`/`else`
owner and tenant entity row policies
i18n field/UI text
locale-aware display formatting
basic RTL direction support
inline UI intent
OpenAPI output
generated contract/API smoke tests
generated frontend/browser-check smoke tests for web targets
generated real browser e2e checks for web targets
generated browser e2e matrix checks for web targets
accessibility audit checks
Docker deployment intent
local preview deploy intent
rollback metadata and read-only rollback plan output
cloud adapter plan/preflight metadata, read-only cloud plan/preflight output, and explicit provider CLI apply runner
ops health/readiness/metrics signals
external observability webhook/OTLP hooks
source security scanning
protected source encryption
production packaging
source/generated benchmark reports
AI task benchmark, long-running eval corpus, and evidence-backed eval history reports
AI-readable docs/inspect/explain output
IDE metadata, autocomplete snippets, diagnostic code catalog, editor-friendly source diagnostics, and packageable VS Code/Open VSX/Cursor-compatible bridge metadata
prepared package registry, editor marketplace, provider adapter marketplace, and public ecosystem index manifests
```

Current BlackLang does not yet provide a general browser runtime or arbitrary frontend event language.

These are not official BlackLang features yet:

```text
<script type="text/black">
browser-side BlackLang interpreter
calculator/game expression runtime
arbitrary custom button click handlers outside declared page actions
arbitrary E2E step execution beyond generated `test` declarations
arbitrary JavaScript replacement
general-purpose programming
general-purpose variables, loops, user-defined functions, imports, or modules
provider-specific integrations without extensions
external media storage providers, image processing, or download permission policies
queue providers, retries, delayed queues, mutating jobs, external calls, or provider schedulers in core job syntax
automatic database migration apply/runtime for arbitrary schema/data transforms
multi-step service transaction orchestration beyond declared action/API update handlers
```

If a task requires one of these, the agent should say that it is outside the current BlackLang capability boundary.

## Do

- Read local project docs before editing.
- Treat `.black` and `.blackthm` files as source of truth.
- Use the official `black` CLI when available.
- Prefer `black docs <keyword> --json` or `black explain <keyword> --json` over guessing syntax.
- Use `black ide --json` when an editor, LSP bridge, or AI agent needs compiler-owned completions, snippets, file extensions, or diagnostic code metadata.
- Use `black ide diagnostics <file> --json` when an editor-style diagnostic panel needs zero-based ranges without rewriting source.
- Use `black ecosystem --json` before package, release, registry, editor extension, deploy adapter, provider adapter, marketplace, or hosted public index work; read `publicIndex.path` before changing `website/ecosystem-index.json`.
- Use `black docs package-registry --json` and `node packages/registry/scripts/validate-registry.mjs` before changing package publish/install behavior.
- Use `black docs adapter-marketplace --json` and `node packages/registry/scripts/validate-registry.mjs` before changing provider adapter marketplace behavior.
- Use `black docs editor-marketplace --json` before changing VS Code, Open VSX, Cursor-compatible, or other editor package channel metadata.
- Use `black docs release-trust --json`, `black ecosystem --json`, `node packages/registry/scripts/validate-registry.mjs`, and `node scripts/verify-release-trust.mjs <release-dir> --json --strict` before trusting GitHub Release, npm, editor, or provider adapter public install paths.
- Use `black docs media --json` or `black explain media --json` before adding file/image fields; keep storage providers and secrets outside `.black`.
- Use `black docs relation-load --json` or `black explain relation-load --json` before changing relation response include behavior; relation load policy controls generated response contexts and does not add query joins.
- Use `black docs job --json` or `black explain job --json` before adding background query jobs; bind jobs to declared queries and run generated `npm run jobs:run` after build for runtime proof.
- Use `black docs target --json` or `black explain target --json` before changing target metadata; choose `target api { backend node database sqlite }`, `database postgres`, or `database mysql` when React/Vite/browser output should be omitted. Generated rename migration runtime is limited to SQLite and PostgreSQL.
- Use normal generated or hand-written web code only when the requested behavior is outside current BlackLang support.
- Use generated `npm test` after `black build` when validating generated contract/API behavior; `target web` also validates frontend/browser-check behavior.
- Use `black docs api --json` or `black explain api --json` before changing explicit API route declarations.
- Use `black docs transaction --json` or `black explain transaction --json` before changing atomic generated route boundaries.
- Use `black docs service --json` or `black explain service --json` before changing explicit API service grouping.
- Use `black docs deploy --json` or `black explain deploy --json` before changing Docker, preview, rollback, cloud adapter, provider CLI execution, or deployment environment intent.
- After building a source with `deploy { cloud ... }`, run generated `npm run deploy:cloud:plan` and `npm run deploy:cloud:preflight` before any provider mutation; `npm run deploy:cloud:exec` is the explicit apply boundary.
- Use `black docs ops --json` or `black explain ops --json` before changing generated health, readiness, metrics, request logging, observability hook, trace context, or exporter intent.
- Use `black audit accessibility --json` after changing generated UI composition, overlays, groups, triggers, forms, or page actions.
- Use `black docs migration --json` and `black migrate plan old.black new.black --json` before deploying entity schema changes to an existing database; declare explicit entity/field renames in the new source when preserving old data. After build, use generated `npm run db:migrate:plan` for a read-only target database report, then `npm run db:migrate` and `npm run db:setup` only when the plan is ready.
- Use `black docs seed --json` or `black explain seed --json` before changing fixture rows; seed values are deterministic local/demo data and must not contain secrets or production credentials.
- Use `black docs test --json` or `black explain test --json` before changing browser test declarations; generated browser checks, `npm run test:e2e`, `npm run test:e2e:plan`, and `npm run test:e2e:matrix` are deterministic page/action/text assertions for `target web`, not arbitrary frontend automation. Run generated `npm run test:e2e:plan` before matrix execution, and run `npm run test:e2e:matrix` when local Chrome, Chromium, or Edge executables are available. Do not put top-level `test` declarations in `target api` sources.
- Use `black benchmark --json` when a report needs source/generated size evidence.
- Use `black benchmark tasks --json` when a report needs deterministic AI task scenarios and token estimate evidence.
- Use `black benchmark eval --json` when a long-running AI edit corpus needs repeat counts, case prompts, expected evidence, scoring criteria, and required validation commands.
- Use `black benchmark eval-history --json` when an agent needs evidence-backed eval result history. Local validation entries are not billed model benchmark scores.
- Use `black benchmark issues --json` when a report or automation only needs tracked web gap totals and current open issue follow-up.
- Use `black security encrypt <file> --out <file.black.enc> --json` to create protected source from an environment key.
- Use `black security decrypt <file.black.enc> --stdout` only in a trusted workspace when plaintext inspection is explicitly needed.
- Clearly label prototypes as prototypes when they are not produced by the official compiler.
- Keep secrets out of `.black` files.
- Keep private release signing keys out of `.black` files, generated output, manifests, package metadata, and adapter metadata. Release verification uses detached Ed25519 `<artifact>.sig` files and `BLACKLANG_RELEASE_PUBLIC_KEY` or `BLACKLANG_RELEASE_PUBLIC_KEY_FILE`; ecosystem metadata also exposes append-only release transparency and key rotation policy requirements.

## Do Not

- Do not invent new BlackLang syntax and present it as official.
- Do not use `<script type="text/black">` unless an official browser runtime exists.
- Do not claim a feature is generated by BlackLang when it was hand-written.
- Do not manually edit generated output as the normal solution.
- Do not run `black format` directly on `.black.enc`; format plaintext source and re-encrypt.
- Do not hide current limitations from the user.
- Do not treat BlackLang as a full Python/JavaScript replacement yet.

## Calculator Example Boundary

A simple calculator currently needs:

```text
button click events
local expression state
custom frontend behavior
keyboard handlers
```

Entity-based computed display fields are now supported for generated table/detail UI, but calculator-style local expression state and button event handling still belong to future frontend logic and event phases.

Today, an agent may build a calculator as a normal one-file HTML/JavaScript prototype, but it should not call that prototype official BlackLang output unless it was generated by the official `black` compiler from supported `.black` syntax.

## Correct Agent Response Pattern

When a user asks for something outside current BlackLang support, respond like this:

```text
BlackLang does not support this as official syntax yet.
I can either:
- build a normal web prototype and label it as non-BlackLang output, or
- add the missing BlackLang feature to the compiler first, then generate it.
```

For repository work, prefer the second option when the user is developing BlackLang itself.

## Current Status Claim

Safe current claim:

```text
BlackLang v0.2 can describe and generate useful CRUD/admin-style web applications and API-only Node services with auth, multi-role assignments, tenant admin UI for tenant policies, owner/tenant row policies, relations with response load policies and generated batch prefetch, workflow, validation, local file/image media fields, computed display fields with bounded arithmetic expressions, first-class seed/fixture declarations, custom queries, custom query aggregate summaries, deterministic background query worker jobs, custom row-level actions with bounded arithmetic set expressions, top-level transaction blocks for custom actions and explicit API update handlers, service blocks for explicit API module metadata, first-class browser-check and browser e2e test declarations for web targets, generated browser e2e matrix plan/run support, i18n field/UI text, locale-aware display formatting, basic RTL direction support, inline UI intent, page view order, DOM-order rendering, reusable page component sections, collection-backed component sections, modal/drawer section display, nested section groups, view interaction triggers, OpenAPI, generated contract/API/frontend/browser-check smoke tests, generated real browser e2e checks for web targets, generated browser e2e matrix checks for web targets, accessibility audit checks, Docker deployment, local preview deploy, rollback metadata, cloud adapter plan/preflight metadata, generated provider CLI deployment preflight/apply runner, generated secret reference manifests, read-only secret readiness plans, and read-only provider preflight checks, ops health/readiness/metrics signals, external observability webhook/OTLP hooks, W3C trace context, generated observability exporter manifests, source-security checks, protected source encryption, production packaging, signed release trust verification, release transparency log policy, key rotation policy, source/generated benchmark reports, AI task benchmark, long-running eval corpus, evidence-backed eval history reports, AI-readable CLI outputs, compiler-owned IDE metadata/diagnostics, packageable VS Code/Open VSX/Cursor-compatible editor channel metadata, prepared package registry manifests, prepared public ecosystem index metadata, and prepared provider adapter marketplace manifests.
```

Unsafe current claim:

```text
BlackLang can replace JavaScript, Python, or all browser logic today.
```

Long-term target claim:

```text
BlackLang aims to grow toward a Python-like general-purpose language after the web target matures, while preserving deterministic AI-friendly source design. Current web intent supports bounded arithmetic expressions for computed fields, custom action sets, and explicit API update handlers, plus ordered local `value` declarations, compound `and`/`or`/`not` conditions, and indentation-based `if`/`else` branches inside custom actions and explicit API update blocks. The post-web Core Program Logic v1 track should still generalize beyond row/API handlers into general-purpose variables, assignment rules, boolean logic, functions, parameters, return, scope, collections, loops, errors, and bounded async operations as separate future language work.
```
