# BlackLang Documentation Site Plan

This folder will contain the official BlackLang learning and reference site.

The site should work for both humans and AI coding agents.

The first published static site source lives in `../website/`.

## Purpose

The documentation site is the public memory of the language.

It should explain:

- How to install BlackLang
- How `.black` files work
- How target platform and generated stack intent is declared, including API-only output
- How each keyword behaves
- How CLI commands behave
- How JSON diagnostics work
- How `.blackthm` theme profile files work
- How compact UI profile slots are read
- How standard UI mode groups work
- How inline UI intent is written near fields, forms, tables, and action buttons
- How page view order, DOM-order rendering, reusable component sections, collection-backed component sections, grid/stack/tabs composition, modal/drawer display, nested section groups, and interaction triggers arrange generated form, table, detail, and component panels
- How relation load policy controls generated list/detail/query/mutation response includes
- How computed display fields are declared from stored numeric data
- How entity indexes generate deterministic Prisma and database setup indexes
- How schema migration plans, explicit rename runtime, and generated migration runner safeguards protect existing database shape changes
- How deterministic seed/fixture rows are declared and applied by generated setup
- How custom queries select bounded, ordered records for pages
- How background query jobs run deterministic read-only query workers on a schedule
- How custom actions mutate rows through validated, permission-checked page actions
- How client-side state declarations bind to generated React page state and modal helpers
- How transaction blocks wrap selected action/API update routes atomically
- How service blocks group explicit APIs into generated service manifests, service modules, and OpenAPI service metadata
- How local file and image fields generate data URL form/API/media UI behavior
- How explicit API declarations generate deterministic declared-runtime routes, typed body validation, and bounded update handlers
- How entity row policies scope generated routes by owner and tenant
- How generated web/API apps expose deterministic contract/API tests, with frontend/browser-check/e2e/matrix layers for web targets
- How first-class browser test declarations assert generated page, action, and text expectations
- How generated UI accessibility policy checks catch source-level layout risks
- How source/generated output benchmarks are measured
- How AI task benchmark scenarios and token estimate reports are generated
- How long-running AI eval corpus metadata is exported
- How evidence-backed AI eval result history is exported
- How current web coverage is reported as a weighted matrix
- How tracked coverage issues are exported as compact JSON/IR
- How IDE metadata, completions, snippets, editor diagnostics, and the packageable VS Code/Open VSX/Cursor-compatible bridge metadata work
- How release/package/editor/adapter discovery is exposed for AI agents
- How prepared package registry manifests and trust checks are validated locally before external publication
- How the prepared public ecosystem index source feeds future hosted package/provider adapter discovery
- How prepared provider adapter marketplace manifests separate provider behavior from core syntax
- How signed release verification, release transparency, and key rotation policy work before public install paths are trusted
- How locked UI profiles enforce append-only slot changes
- How theme migration checks prevent positional UI remaps
- How i18n locale, field/UI text translations, value formatting, and RTL direction work
- How security CORS policy is declared without hardcoding deployment origins
- How protected `.black.enc` source files are encrypted, decrypted, and built in memory
- How generated secret/env reference manifests, read-only readiness plans, and provider preflight checks work
- How Docker deployment intent generates production runtime files, local preview compose, rollback metadata, cloud adapter plans, and provider CLI preflight/apply runners
- How ops health, readiness, metrics, request logging, observability webhook/OTLP hooks, W3C trace context, and Docker healthchecks are declared
- How BlackLang compares with common web languages without overstating current capability
- How measured scale benchmarks differ from positioning notes
- How AI agents should respect the current capability boundary
- How AI agents should use the language

## Planned Structure

```text
docs/
├── learn/
├── reference/
├── cli/
├── errors/
├── install.md
├── github-publish.md
├── target.md
├── i18n.md
├── inline-ui.md
├── theme-migration.md
├── secret-management.md
├── view.md
├── relation-load.md
├── computed.md
├── index.md
├── migration.md
├── seed.md
├── query.md
├── job.md
├── action.md
├── state.md
├── service.md
├── media.md
├── api.md
├── policy.md
├── generated-test.md
├── test.md
├── accessibility.md
├── benchmark.md
├── ide.md
├── editor-marketplace.md
├── ecosystem.md
├── ../website/ecosystem-index.json
├── package-registry.md
├── adapter-marketplace.md
├── release-trust.md
├── ai-agent-contract.md
├── deployment.md
├── ops.md
├── security-cors.md
├── protected-source.md
├── release-artifacts.md
├── npm-wrapper.md
├── ai-agents/
├── examples/
├── benchmarks/
│   └── language-comparison-v0.2.md
│   └── scale-5327-loc.md
├── llms.txt
└── llms-full.txt
```

## First Pages

- Learn: What is BlackLang?
- Learn: Quick start
- Reference: Syntax
- Reference: app
- Reference: target
- Reference: entity
- Reference: page
- Reference: table
- Reference: form
- Reference: view
- Reference: relation load policy
- Reference: computed
- Reference: index
- Reference: migration
- Reference: seed and fixture declarations
- CLI: schema migration plan, rename runtime, and generated migration runner
- Reference: query
- Reference: background query jobs
- Reference: action
- Reference: client state declarations
- Reference: service blocks
- Reference: local file and image media fields
- Reference: explicit API declared-runtime routes and bounded update handlers
- Reference: entity row policies
- Reference: generated contract/API tests and target web frontend/browser-check/e2e/matrix checks
- Reference: browser test declarations
- CLI: accessibility audit
- CLI: benchmark
- CLI: benchmark tasks
- CLI: benchmark eval corpus
- CLI: benchmark eval history
- CLI: benchmark coverage
- CLI: IDE metadata and diagnostics
- CLI: ecosystem discovery
- CLI: package registry manifests
- CLI: provider adapter marketplace manifests
- CLI: editor marketplace channel manifests
- CLI: public ecosystem index source metadata
- CLI: signed release trust, release transparency, and key rotation policy
- CLI: protected source encryption
- Reference: ops runtime signals
- CLI: parse
- CLI: format
- CLI: lint
- CLI: validate
- CLI: docs all JSON export
- CLI: explain keyword JSON
- CLI: theme inspect JSON
- CLI: theme migrate JSON/IR
- CLI: install paths
- CLI: GitHub publish policy
- CLI: release artifact layout
- CLI: npm wrapper plan
- AI Agents: Codex guide
- AI Agents: Capability boundary and official workflow
- Errors: Error code reference
- Reference: Theme profile
- Reference: UI profile rules
- Reference: UI modes
- Reference: Inline UI intent
- Reference: Theme migration
- Reference: I18n

## Documentation Rule

Every language or CLI change should update the relevant documentation in the same change.
