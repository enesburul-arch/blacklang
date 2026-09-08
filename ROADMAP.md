# BlackLang Roadmap

Status: v0.1 roadmap complete. The next milestone is tracked in `ROADMAP-v0.2.md`; v0.2 now includes computed display fields, relation response load policies, entity index declarations, schema migration planning, explicit rename migration runtime, first-class seed/fixture runtime, first-class browser-check and browser e2e test declarations, deterministic AI task/token benchmark reports, long-running AI eval corpus metadata, evidence-backed local eval result history export, custom query and aggregate summary MVP, deterministic background query worker jobs, custom row-level action MVP, bounded action/API local logic with `value`, indentation-based `if`/`else`, and compound `and`/`or`/`not` conditions, owner/tenant policies, ops runtime signals, cloud adapter manifests, provider CLI deployment preflight/apply runners, observability webhook/OTLP hooks, W3C trace context, prepared package registry/provider adapter marketplace manifests, prepared multi-editor package channel metadata, prepared public ecosystem index metadata, and deterministic local file/image media fields.

Current web completion work targets deterministic web capability parity, not JavaScript runtime parity. The web target now has bounded arithmetic expressions for computed fields, custom action sets, and explicit API update handlers, plus ordered local `value` declarations, Python-like indentation-based `if`/`else` branches, and compound `and`/`or`/`not` conditions inside custom actions and explicit API update blocks. After the web target is mature, BlackLang should add a separate Core Program Logic v1 track that generalizes this beyond row/API handlers into general-purpose variables, assignment rules, boolean logic, functions, parameters, return, scope, collections, loops, errors, and bounded async operations. That future core track should be measured with separate evals such as calculator local expression state, todo local state, conditional pricing, collection loops, and bounded async API calls.

## Phase 0: Project Foundation

- [x] Create repository structure
- [x] Define project purpose
- [x] Write initial language specification
- [x] Write AI agent instructions
- [x] Add first `.black` example
- [x] Create initial single-binary CLI structure
- [x] Add `black init` project scaffold command
- [x] Read `blacklang.toml` for default source and output paths
- [x] Add `black inspect --ir` project summary command
- [x] Add `black docs <keyword> --ir` compact learning command

## Phase 1: Minimal Language Core

- Support `app`
- Support `entity`
- Support basic field types
- Support field modifiers: `required`, `unique`, `default`, `optional`
- Support `page`
- Support `source`
- Support `table`
- Support `form`
- Support `actions`
- Support `search`

## Phase 2: Parser and AST

- [x] Parse source into AST
- [x] Preserve file, line, and column positions for parsed nodes
- [x] Return JSON parse output
- [x] Report syntax errors in JSON
- [x] Replace line-based parser with a stricter token stream when syntax grows

## Phase 3: Validator

- [x] Detect missing app declaration
- [x] Detect duplicate entities
- [x] Detect duplicate fields
- [x] Validate supported field types
- [x] Validate supported field modifiers
- [x] Validate page sources
- [x] Validate table columns
- [x] Validate form fields
- [x] Validate searchable fields
- [x] Validate supported actions

## Phase 4: Web Generator

- [x] Generate TypeScript types
- [x] Generate validation schemas
- [x] Generate database schema
- [x] Generate API route skeletons
- [x] Generate React page skeletons
- [x] Generate package metadata
- [x] Generate output README
- [x] Generate database workflow scripts and environment example
- [x] Generate Docker deployment files from deploy intent
- [x] Generate Vite entry files
- [x] Generate basic create/edit/delete UI behavior from actions
- [x] Verify generated Vite app with npm install and production build

## Phase 5: Measurement

- [x] Compare BlackLang source with normal web stack code
- [x] Measure source lines
- [x] Measure generated lines
- Measure files touched by AI
- Estimate token usage
- Track validation/build errors

## Phase 6: CRUD Depth

- [x] Generate API client layer for page actions
- [x] Generate Express API server entry
- [x] Wire React pages to generated API clients
- [x] Generate loading, saving, and error UI states
- [x] Generate Prisma Client database singleton
- [x] Generate Prisma-backed list/read/create/update/delete routes
- [x] Generate deterministic SQLite setup script
- [x] Alias `db:push` to generated SQLite setup for MVP
- [x] Revisit native `prisma db push` when Prisma schema-engine is stable in the local Windows environment
- [x] Add read/detail behavior
- [x] Add bulk delete behavior
- [x] Add archive/restore semantics

## Phase 7: Relations

- [x] Accept existing entity names as field types
- [x] Validate unknown relation target entities
- [x] Generate Prisma relation fields and back references
- [x] Generate SQLite foreign key columns in setup script
- [x] Generate relation select inputs in forms
- [x] Generate relation display in tables and details
- [x] Generate empty-state guidance for required relation form fields
- [x] Generate navigation from missing required relation fields to related pages
- [x] Support relation fields in generated table search

## Phase 8: Forms

- [x] Parse field label modifiers
- [x] Validate missing label values
- [x] Generate form labels from field label modifiers
- [x] Reuse field labels in table headers and detail panels
- [x] Generate placeholders from field metadata
- [x] Generate help text from field metadata
- [x] Generate inline validation messages

## Phase 9: Tables

- [x] Parse table sort declarations
- [x] Validate table sort fields and directions
- [x] Generate default table sorting in React lists
- [x] Parse table pagination declarations
- [x] Generate pagination controls in React lists
- [x] Generate column visibility controls
- [x] Parse table filter declarations
- [x] Generate table filter controls

## Phase 10: Layout

- [x] Generate application shell from page list
- [x] Generate sidebar navigation
- [x] Generate topbar and breadcrumb
- [x] Parse explicit layout declarations
- [x] Generate sidebar order from explicit layout declarations
- [x] Generate responsive drawer navigation

## Phase 11: API

- [x] Generate OpenAPI document from entities and page actions
- [x] Serve generated OpenAPI document from the Express API server
- [x] Parse explicit `api` declarations
- [x] Generate query parameter contracts
- [x] Generate path parameter contracts beyond default `id`
- [x] Generate public/private endpoint metadata
- [x] Generate webhook endpoint contracts
- [x] Generate declared runtime routes for explicit API and webhook declarations

## Phase 12: Security Foundation

- [x] Disable generated Express framework fingerprint header
- [x] Generate baseline browser security headers
- [x] Generate JSON request body size limit
- [x] Generate simple IP-based rate limiting
- [x] Parse explicit auth syntax
- [x] Validate explicit auth syntax
- [x] Generate login and register UI shell from auth syntax
- [x] Generate auth API routes from auth syntax
- [x] Generate password hashing for auth routes
- [x] Generate cookie session persistence from auth syntax
- [x] Protect generated CRUD API routes with cookie sessions
- [x] Restore frontend auth state from `/api/auth/me`
- [x] Generate logout behavior
- [x] Parse role and permission syntax
- [x] Validate role and page access references
- [x] Include role and page access in JSON/BlackIR outputs
- [x] Generate basic runtime role storage
- [x] Generate page-level role enforcement for page access
- [x] Generate basic role management UI
- [x] Generate tenant ID management UI when tenant policies exist
- [x] Generate action-level role enforcement
- [x] Generate field-level read hiding
- [x] Generate field-level mutation enforcement
- [x] Generate audit log support
- [x] Generate CSRF/session protection for cookie auth

## Phase 13: Workflow System

- [x] Parse workflow declarations
- [x] Validate workflow source entities
- [x] Validate workflow states
- [x] Validate workflow transitions
- [x] Validate transition allow roles
- [x] Include workflow intent in JSON/BlackIR outputs
- [x] Generate workflow transition API routes
- [x] Generate workflow transition UI controls
- [x] Generate workflow audit log entries
- [x] Generate workflow tests

## Phase 14: State and Client Behavior

- [x] Parse state declarations
- [x] Validate state fields and modal defaults
- [x] Include state intent in JSON/BlackIR outputs
- [x] Bind explicit state declarations to generated React state
- [x] Generate modal open/close helpers from state declarations
- [x] Generate state tests for runtime behavior

## Phase 15: Component System

- [x] Parse component declarations
- [x] Validate component inputs and variants
- [x] Include component intent in JSON/BlackIR outputs
- [x] Generate React components from component declarations
- [x] Bind components to table/detail rendering
- [x] Bind components to form rendering
- [x] Place declared components as reusable page view sections
- [x] Generate component tests for runtime behavior

## Phase 16: Validation System

- [x] Parse field constraint modifiers: `min`, `max`, `length`
- [x] Validate numeric `min`/`max` constraints
- [x] Validate text/email `length min..max` constraints
- [x] Generate frontend inline validation from field constraints
- [x] Generate API validation from field constraints
- [x] Parse regex validation
- [x] Parse URL validation
- [x] Generate custom validation messages
- [x] Generate cross-field validation
- [x] Generate conditional validation

## Phase 17: BlackLang Source Security

- [x] Record `.black` files as high-value source assets
- [x] Document that secrets must not be stored in `.black` files
- [x] Document production artifact preference over shipping source files
- [x] Parse secret-safe environment references such as `env DATABASE_URL`
- [x] Validate database url env references
- [x] Include database env references in JSON/BlackIR outputs
- [x] Add `black security scan --json`
- [x] Detect likely hardcoded passwords, API keys, tokens, and private keys
- [x] Generate secret/env reference manifest and read-only readiness plan
- [x] Add production packaging that excludes protected source by default
- [x] Explore encrypted source mode such as `app.black.enc`

## Phase 18: Deployment and Environment

- [x] Parse `deploy { target docker }`
- [x] Validate deploy target, port env, and deploy env declarations
- [x] Include deploy intent in JSON/BlackIR outputs
- [x] Generate `.env.example` port/default env wiring
- [x] Generate Dockerfile, `.dockerignore`, and `docker-compose.yml`
- [x] Make generated server read `PORT`
- [x] Serve built Vite frontend from the generated Express server
- [x] Add PostgreSQL runtime support before enabling postgres deploy target
- [x] Add MySQL runtime, seed, and Docker Compose support before enabling mysql deploy target
- [x] Generate health/readiness/metrics ops endpoints and Docker healthchecks from `ops`
- [x] Add local preview deployment target and rollback metadata
- [ ] Add cloud deployment adapters and external observability hooks

## Phase 19: Target and Plugin Foundation

- [x] Parse `target web { frontend react backend node database sqlite|postgres|mysql }` and `target api { backend node database sqlite|postgres|mysql }`
- [x] Validate target platform and generated stack declarations
- [x] Include target intent in JSON and BlackIR outputs
- [x] Add target docs and agent-facing explain support
- [x] Include target changes in affected graph analysis
- [x] Record generated target stack in generated README output
- [ ] Add generator adapter/plugin discovery
- [x] Add API-only target support
- [ ] Add mobile and desktop target planning after web stabilizes
- [x] Add PostgreSQL target only after generated runtime support exists
