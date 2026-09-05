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

## Phase 20: UI and Theme Language

This phase starts the BlackLang-native UI layer instead of forcing users or AI agents to write raw CSS.

- [x] Define `blackthm` or theme profile file format
- [x] Define compact UI slot profile rules
- [x] Make UI slots append-only after profile lock
- [x] Add mode groups such as `box`, `text`, `table`, and `button`
- [x] Support inline UI intent near fields, forms, tables, and buttons
- [x] Generate CSS from UI intent
- [x] Make CSS generation read `.blackthm` `ui <mode> = <slot...>;` order
- [ ] Add migration rules for UI profile changes

## Phase 21: Query, Actions, and Data Logic

This phase moves beyond CRUD screens into richer application behavior.

- [ ] Add computed fields
- [ ] Add custom query declarations bound to entities
- [ ] Add custom mutation/action declarations
- [ ] Add generated API routes for custom data actions
- [ ] Add generated UI controls for custom actions
- [ ] Add permission checks for custom actions
- [ ] Add validation for action input/output

## Phase 22: Tests, Benchmarks, and Evals

This phase keeps BlackLang honest by measuring whether it really reduces repeated AI work.

- [x] Add qualitative language positioning comparison
- [x] Add measured 5327 LOC synthetic scale benchmark
- [ ] Generate basic frontend tests
- [ ] Generate basic API route tests
- [ ] Add compiler golden tests for generated output
- [ ] Add benchmark command for source/generated line counts
- [ ] Add AI task benchmark scenarios
- [ ] Track token estimate reports
- [ ] Publish benchmark notes in `benchmarks/`

## Phase 23: Protected Source Mode

This phase turns the `.black.enc` policy into real tooling.

- [ ] Define encrypted source file header
- [ ] Define key source rules
- [ ] Add `black security encrypt`
- [ ] Add `black security decrypt --stdout`
- [ ] Add safe temporary build flow for encrypted source
- [ ] Ensure decrypted source is never written during production package
- [ ] Add tests for package exclusion and encrypted-source diagnostics

## Phase 24: Documentation Site

This phase creates the Python/W3Schools-style learning surface discussed for BlackLang.

- [x] Define docs site information architecture
- [x] Add quick start pages
- [x] Add language reference pages
- [x] Add examples gallery
- [x] Add generated output comparison pages
- [x] Add AI agent guide pages
- [x] Add versioned docs strategy for future syntax changes

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

This phase starts locale-aware generated UI text without adding runtime language switching yet.

- [x] Parse top-level `i18n` default/locales declarations
- [x] Parse top-level `label Entity.field` translation blocks
- [x] Validate locale lists, default locale, label targets, and duplicate translations
- [x] Include i18n and label translations in JSON/BlackIR outputs
- [x] Generate default-locale field labels in web output
- [x] Add human and AI documentation for i18n
- [ ] Add runtime language switching
- [ ] Add placeholder/help/message translations
- [ ] Add date, number, and currency formatting
- [ ] Add RTL layout support

## Phase 27: Full Web Coverage Plan

This phase makes the long-term web target measurable instead of vague.

- [x] Define what "100% web coverage" means for BlackLang
- [x] Separate core language features from provider/plugin extensions
- [x] Add a coverage matrix for frontend, backend, database, security, deployment, testing, and ecosystem work
- [x] Add percentage milestones from MVP web app to full web coverage model
- [x] Add the priority order for reaching full web coverage
- [ ] Turn the coverage matrix into tracked implementation issues
- [ ] Add a `black benchmark` coverage report command
- [ ] Publish coverage status on the documentation site

## Phase 28: Page View Order and Layout Composition

This phase starts moving generated page composition into `.black` source instead of generated React or CSS edits.

- [x] Add `view { order ... }` syntax inside page blocks
- [x] Validate supported view sections and duplicates
- [x] Include view order in JSON/BlackIR output
- [x] Generate stable page section classes for table, detail, and form
- [x] Generate deterministic web CSS order rules from view intent
- [x] Add human and AI documentation for page view order
- [ ] Add nested sections, grid, stack, tabs, modal, drawer, and responsive composition syntax
- [ ] Add DOM-order rendering for richer composition when the layout model grows beyond simple ordering

## v0.2 Exit Criteria

v0.2 is ready when:

- BlackLang can be installed from a documented release path.
- AI agents can inspect, explain, format, validate, build, and benchmark a project from CLI output.
- UI customization has a BlackLang-native first syntax.
- At least three real app templates exist.
- Protected source mode has a working first implementation or a clearly documented limitation.
- The documentation site can teach both humans and AI agents the current language version.
