# BlackLang Roadmap v0.2

## Purpose

v0.1 proved the core idea: a compact `.black` source can generate a working web application with auth, roles, CRUD, relations, workflow, validation, security defaults, packaging, and AI-readable JSON/IR outputs.

v0.2 should turn that proof into a more usable language and ecosystem.

The focus is:

- safer day-to-day editing for AI agents
- clearer install and release flow
- better UI/layout control without falling back to raw HTML/CSS
- stronger docs for humans and AI agents
- repeatable benchmarks and generated tests
- first real steps toward protected/encrypted source workflows

## Phase 18: Release and Install Path

This phase makes BlackLang installable and repeatable outside this local folder.

- [x] Create v0.2 roadmap
- [x] Define official CLI artifact layout
- [x] Add release build script for Windows binary
- [x] Add checksum generation for release artifacts
- [x] Add npm wrapper package plan for `npx blacklang`
- [x] Document install paths for GitHub Releases and npm
- [x] Add `black version --json`

## Phase 19: Developer and AI Ergonomics

This phase helps humans and AI agents edit `.black` files with less guesswork.

- [x] Add `black format` command
- [x] Add `black lint --json`
- [x] Add `black docs --all --json`
- [x] Add `black explain <keyword> --json`
- [x] Add `black inspect --affected <symbol> --json`
- [x] Add stable diagnostic documentation
- [x] Add agent startup checklist output
- [x] Add AI agent capability boundary and official workflow contract

## Phase 20: UI and Theme Language

This phase starts the BlackLang-native UI layer instead of forcing users or AI agents to write raw CSS.

- [x] Define `blackthm` or theme profile file format
- [x] Define compact UI slot profile rules
- [x] Make UI slots append-only after profile lock
- [x] Add mode groups such as `box`, `text`, `table`, and `button`
- [x] Support inline UI intent near fields, forms, tables, and buttons
- [x] Generate CSS from UI intent
- [x] Make CSS generation read `.blackthm` `ui <mode> = <slot...>;` order
- [x] Add migration rules for UI profile changes

## Phase 21: Query, Actions, and Data Logic

This phase moves beyond CRUD screens into richer application behavior.

- [x] Add computed fields
- [x] Add shared bounded arithmetic expression trees with precedence and parentheses for computed fields, custom action sets, and explicit API update handlers
- [x] Add ordered local `value` declarations, indentation-based `if`/`else` branches, and compound `and`/`or`/`not` conditions for custom actions and explicit API update blocks
- [x] Add relation response load policies for generated list, detail, query, and mutation routes
- [x] Add single-hop bulk relation prefetch and nested relation response sanitization
- [x] Add custom query declarations bound to entities and pages: typed stored-field filters, deterministic sort/limit, generated query lists, aggregate summaries, permission guards, and AI tooling
- [x] Add custom mutation/action declarations: top-level row-level actions with primitive inputs and deterministic set expressions
- [x] Add top-level transaction blocks for custom actions and explicit API update handlers
- [x] Add generated API routes for custom data actions
- [x] Add generated declared-runtime routes for explicit API and webhook declarations
- [x] Add generated UI controls for custom actions
- [x] Add permission checks for custom actions
- [x] Add validation for action input/output
- [x] Add entity index declarations for stored fields and relation fields
- [x] Generate deterministic Prisma and SQLite indexes from entity index declarations
- [x] Add generated schema migration plan before destructive database changes
- [x] Add applied migration runtime and first-class rename/refactor syntax
- [x] Add generated online migration runner safeguards with read-only `db:migrate:plan` and explicit `db:migrate`
- [x] Add first-class seed/fixture declarations with generated deterministic `db:seed` runtime

## Phase 22: Tests, Benchmarks, and Evals

This phase keeps BlackLang honest by measuring whether it really reduces repeated AI work.

- [x] Add qualitative language positioning comparison
- [x] Add measured 5327 LOC synthetic scale benchmark
- [x] Generate deterministic contract/API/frontend smoke tests in generated web projects
- [x] Generate basic frontend tests
- [x] Generate basic API route tests, including explicit API declared-runtime route probes
- [x] Add compiler golden tests for generated output
- [x] Add first-class fixture/seed syntax and generated setup runtime
- [x] Add first-class browser-check test syntax and generated browser checks
- [x] Add benchmark command for source/generated line counts
- [x] Add AI task benchmark scenarios
- [x] Add long-running AI eval corpus metadata
- [x] Add evidence-backed local AI eval result history export
- [x] Add full browser e2e execution with real browser automation
- [x] Add generated cross-browser e2e matrix plan and runner
- [x] Track token estimate reports
- [x] Publish benchmark notes in `benchmarks/`

## Phase 23: Protected Source Mode

This phase turns the `.black.enc` policy into real tooling.

- [x] Define encrypted source file header
- [x] Define key source rules
- [x] Add `black security encrypt`
- [x] Add `black security decrypt --stdout`
- [x] Add safe temporary build flow for encrypted source
- [x] Ensure decrypted source is never written during production package
- [x] Add tests for package exclusion and encrypted-source diagnostics

## Phase 24: Documentation Site

This phase creates the Python/W3Schools-style learning surface discussed for BlackLang.

- [x] Define docs site information architecture
- [x] Add quick start pages
- [x] Add language reference pages
- [x] Add examples gallery
- [x] Add generated output comparison pages
- [x] Add AI agent guide pages
- [x] Add versioned docs strategy for future syntax changes
- [x] Add public AI agent contract page for current capability boundaries

## Phase 25: Real App Templates

This phase grows the examples from Warehouse into reusable app patterns.

- [x] Add CRM example
- [x] Add inventory example
- [x] Add helpdesk example
- [x] Add invoice example
- [x] Add appointment example
- [x] Add project management example
- [x] Add template benchmark reports

## Phase 26: I18n MVP

This phase starts locale-aware generated UI text, field-level messages, display formatting, and runtime direction support.

- [x] Parse top-level `i18n` default/locales declarations
- [x] Parse top-level `label Entity.field` translation blocks
- [x] Validate locale lists, default locale, label targets, and duplicate translations
- [x] Include i18n and label translations in JSON/BlackIR outputs
- [x] Generate runtime field-label switching in web output
- [x] Add human and AI documentation for i18n
- [x] Add runtime language switching for field and computed display labels
- [x] Add placeholder/help/message translations
- [x] Add date, number, and currency formatting
- [x] Add RTL layout support
- [x] Add runtime app chrome, table/status, CRUD/custom action, and workflow transition copy labels

## Phase 27: Full Web Coverage Plan

This phase makes the long-term web target measurable instead of vague.

- [x] Define what "100% web coverage" means for BlackLang
- [x] Separate core language features from provider/plugin extensions
- [x] Add a coverage matrix for frontend, backend, database, security, deployment, testing, and ecosystem work
- [x] Add percentage milestones from MVP web app to full web coverage model
- [x] Add the priority order for reaching full web coverage
- [x] Turn the coverage matrix into tracked implementation issues
- [x] Add a `black benchmark` coverage report command
- [x] Publish coverage status on the documentation site
- [x] Add compact tracked issue export through `black benchmark issues --json|--ir`
- [x] Clarify web completion as deterministic web capability parity instead of JavaScript runtime parity
- [x] Record post-web Core Program Logic v1 as a separate future track
- [x] Add compiler-owned IDE metadata, completions, snippets, and diagnostics export
- [x] Add a packageable VS Code extension bridge for compiler-owned IDE diagnostics, completions, snippets, and source quick fixes
- [x] Add prepared VS Code, Open VSX, and Cursor-compatible editor package channel metadata for the shared compiler-owned bridge

## Phase 28: Page View Order and Layout Composition

This phase starts moving generated page composition into `.black` source instead of generated React or CSS edits.

- [x] Add `view { order ... }` syntax inside page blocks
- [x] Validate supported view sections and duplicates
- [x] Include view order in JSON/BlackIR output
- [x] Generate stable page section classes for table, detail, and form
- [x] Generate deterministic web CSS order rules from view intent
- [x] Add human and AI documentation for page view order
- [x] Add first page composition syntax: stack/grid, gap, responsive stackAt, and section span
- [x] Add first tabbed page composition syntax with generated tab controls
- [x] Add modal and drawer section display for generated detail/form panels
- [x] Add nested section groups with local stack/grid composition
- [x] Add read-only accessibility audit checks for generated overlays and nested groups
- [x] Add advanced interactive composition triggers
- [x] Add reusable page component sections bound to selected or first source record data
- [x] Add collection-backed page component sections with `bind each`
- [x] Add DOM-order rendering for richer composition when the layout model grows beyond simple ordering

## Phase 29: Ownership and Tenant Policies

This phase adds the first deterministic row-scope security layer for generated web applications.

- [x] Add entity `policy owner <field>` and `policy tenant <field>` syntax
- [x] Require policy fields to be stored `text required` non-unique fields
- [x] Require auth before policy use
- [x] Keep policy fields out of generated forms, API client input types, and OpenAPI input schemas
- [x] Prevent custom actions from setting policy fields
- [x] Scope generated list, query, detail, mutation, archive, delete, workflow transition, and custom action routes
- [x] Add generated tenant storage to auth runtime when tenant policies exist
- [x] Add generated tenant ID management UI when tenant policies exist
- [x] Add JSON, BlackIR, inspect/affected, docs, explain, diagnostics, tests, Warehouse example, and documentation site support

## Phase 30: Runtime Ops Signals

This phase makes generated web deployments observable by standard infrastructure checks without custom generated-file edits.

- [x] Add top-level `ops` declaration for health, readiness, metrics, and request logging
- [x] Validate safe root-level public ops paths and duplicate paths
- [x] Generate public health, readiness, and metrics routes before API auth and CSRF middleware
- [x] Generate structured JSON request logs when `logging requests` is declared
- [x] Add OpenAPI `x-blacklang-ops` and `x-blacklang-public` metadata
- [x] Probe ops routes from generated API smoke tests
- [x] Add Dockerfile and Docker Compose app healthchecks from `ops.health`
- [x] Add JSON, BlackIR, inspect/affected, docs, explain, diagnostics, tests, Warehouse example, and documentation site support

## Phase 30.1: Local Preview Deploy and Rollback Metadata

This phase makes generated web deployments easier to rehearse and recover without adding cloud-specific behavior before adapters exist.

- [x] Add `preview local` inside the existing top-level `deploy` block
- [x] Add `rollback keep <count>` inside the existing top-level `deploy` block
- [x] Validate supported preview mode and bounded rollback keep counts
- [x] Generate `docker-compose.preview.yml` with isolated preview port and preview database defaults
- [x] Generate `deploy/manifest.json`, `deploy/rollback.json`, and `scripts/rollback-plan.mjs`
- [x] Add `deploy:preview`, `deploy:preview:down`, and `deploy:rollback:plan` package scripts when declared
- [x] Add JSON, BlackIR, inspect/affected, docs, explain, diagnostics, tests, Warehouse example, and documentation site support
- [x] Add cloud deploy adapters after adapter metadata exists
- [x] Add external observability provider hooks and OTLP trace exporter metadata

## Phase 30.2: Cloud Adapter Manifest and Observability Hooks

This phase keeps cloud/provider work deterministic without making the core compiler mutate external infrastructure.

- [x] Add `cloud fly|render|railway app env NAME [region env NAME]` inside the existing top-level `deploy` block
- [x] Validate supported cloud providers, Docker target requirement, and env-only app/region values
- [x] Generate `deploy/cloud.json`, cloud metadata in `deploy/manifest.json`, read-only `scripts/cloud-plan.mjs`, and explicit-apply `scripts/cloud-exec.mjs`
- [x] Add `deploy:cloud:plan`, `deploy:cloud:preflight`, and `deploy:cloud:exec` package scripts when a cloud adapter is declared
- [x] Keep provider CLI deployment gated by read-only preflight, required env checks, CLI availability checks, and explicit apply mode
- [x] Add `observe webhook endpoint env NAME` inside the existing top-level `ops` block
- [x] Generate non-blocking request-event webhook delivery, OpenAPI observability metadata, and local API smoke coverage
- [x] Add `observe otlp endpoint env NAME` as a supported provider value using the same observe syntax
- [x] Generate W3C traceparent context, non-blocking OTLP HTTP JSON trace payloads, and `ops/observability.json` exporter metadata
- [x] Add JSON, BlackIR, inspect/affected, docs, explain, diagnostics, IDE metadata, tests, Warehouse example, and documentation site support

## Phase 30.3: Package Registry and Provider Adapter Marketplace Manifests

This phase makes ecosystem publishing and provider adapter growth discoverable without letting package tooling reimplement the language.

- [x] Add prepared local package registry manifest under `packages/registry/package-index.blackdir`
- [x] Add package registry trust policy under `packages/registry/trust-policy.blackdir`
- [x] Add prepared provider adapter marketplace manifest under `adapters/marketplace/adapter-index.blackdir`
- [x] Add provider adapter marketplace trust policy under `adapters/marketplace/trust-policy.blackdir`
- [x] Add local registry/marketplace manifest validator
- [x] Expose registries, marketplaces, and trust workflow checks from `black ecosystem --json|--ir`
- [x] Add `docs package-registry --json` and `docs adapter-marketplace --json` learning entries
- [x] Add `docs editor-marketplace --json` learning entry for multi-editor package channels
- [x] Add prepared public ecosystem index source metadata for future hosted package/provider adapter discovery
- [x] Add explain, docs, tests, AI agent contract, roadmap, and documentation site support

## Phase 30.4: File and Image Media Fields

This phase adds deterministic local media fields without introducing provider storage, secrets, or upload-side effects into `.black` source.

- [x] Add `file` and `image` stored entity field types
- [x] Add `accept "MIME hint"` as the single media input hint syntax
- [x] Validate `accept` only on `file` or `image` entity fields
- [x] Reject custom action media inputs and auth user media fields for this MVP
- [x] Generate local file inputs that read selected files as data URL strings
- [x] Generate API, frontend, seed, and explicit body validation for data URL or absolute http(s) media values
- [x] Generate table/detail image thumbnails, file links, and form previews
- [x] Emit OpenAPI `x-blacklang-media` and `x-blacklang-encoding: data-url` metadata
- [x] Add docs, explain, inspect/affected, diagnostics, IDE metadata, tests, Warehouse example, and documentation site support

## Phase 30.5: Background Query Jobs

This phase adds deterministic generated background worker jobs for read-only query work without introducing queue providers or external scheduler side effects into core syntax.

- [x] Add top-level `job Name { ... }` declarations
- [x] Add `schedule every <integer> minutes|hours|days` as the single MVP schedule syntax
- [x] Add `run query QueryName` as the single MVP run mode
- [x] Validate query references, supported schedule units, bounded intervals, duplicate clauses, and name collisions
- [x] Generate `jobs/manifest.json` and `src/worker.ts`
- [x] Add generated `jobs:run` and `jobs:loop` package scripts
- [x] Emit OpenAPI `x-blacklang-jobs` metadata
- [x] Add JSON, BlackIR, inspect/affected, docs, explain, diagnostics, IDE metadata, tests, Warehouse example, and documentation site support

## Phase 30.6: API-only Target

This phase lets the same deterministic app intent generate an API/runtime project without React, Vite, frontend smoke tests, browser-check tests, or browser e2e files.

- [x] Add `target api { backend node database sqlite|postgres|mysql }` as the single API-only target syntax
- [x] Keep `frontend` unsupported inside `target api`
- [x] Keep top-level browser `test` declarations bound to `target web`
- [x] Generate Express, Prisma, OpenAPI, validation, API clients/routes, contract/API tests, optional seed/job/deploy/ops output, and no frontend files
- [x] Add docs, explain, diagnostics, IDE metadata, tests, coverage issue tracking, and documentation site support

## Phase 30.7: Secret Reference Manifest

This phase makes environment-backed secret/config references reviewable without storing values in `.black` source or generated manifests.

- [x] Generate `security/secrets.json` with source, kind, required, and sensitive metadata for environment-backed references
- [x] Generate read-only `scripts/secrets-plan.mjs` and `security:secrets:plan`
- [x] Support provider handoff labels through `BLACKLANG_SECRET_PROVIDER` and `BLACKLANG_SECRET_PREFIX` without fetching or printing secret values
- [x] Generate read-only `scripts/secrets-provider.mjs` and `security:secrets:preflight` for provider label, prefix, and CLI readiness checks
- [x] Add generated contract tests, Go tests, docs, coverage issue tracking, and documentation site support

## Phase 30.8: Signed Release Trust

This phase makes public install paths verifiable before GitHub Release, npm, editor, or provider adapter publishing.

- [x] Add read-only `scripts/verify-release-trust.mjs` for release manifest, checksum, detached Ed25519 signature, and public key verification
- [x] Expose signature algorithm, signature file pattern, public key env names, and verification command through `black ecosystem --json|--ir`
- [x] Add release trust docs, explain/docs JSON support, package registry and adapter marketplace trust policy markers
- [x] Add release transparency log and key rotation policy metadata
- [x] Require npm wrapper downloads to verify detached signatures before extracting or executing downloaded binaries
- [x] Add tests, local registry validation, coverage issue tracking, and documentation site support

## Phase 30.9: Transaction Blocks

This phase separates write-handler declarations from atomic runtime boundary intent.

- [x] Add top-level `transaction Name { ... }` syntax for action/API update targets
- [x] Validate duplicate, unknown, empty, and unsupported transaction targets
- [x] Generate Prisma transaction wrappers for targeted custom actions and explicit API update handlers
- [x] Expose transaction bindings through JSON, BlackIR, OpenAPI metadata, inspect/affected, docs, explain, diagnostics, IDE metadata, Warehouse example, tests, and documentation site support

## Phase 30.10: Service Blocks

This phase separates explicit API endpoint declarations from service/module grouping intent.

- [x] Add top-level `service Name { api APIName }` syntax for grouping explicit APIs without redefining routes
- [x] Validate duplicate, unknown, empty, and duplicate-bound service API targets
- [x] Generate `services/manifest.json` and `src/services/<service>.ts` metadata modules for service-bound APIs
- [x] Expose service bindings through JSON, BlackIR, OpenAPI tags/extensions, inspect/affected, docs, explain, diagnostics, IDE metadata, Warehouse example, tests, and documentation site support

## v0.2 Exit Criteria

v0.2 is ready when:

- BlackLang can be installed from a documented release path.
- AI agents can inspect, explain, format, validate, build, and benchmark a project from CLI output.
- UI customization has a BlackLang-native first syntax.
- At least three real app templates exist.
- Protected source mode has a working first implementation or a clearly documented limitation.
- The documentation site can teach both humans and AI agents the current language version.
