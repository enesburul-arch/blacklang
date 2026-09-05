# BlackLang Roadmap

Status: v0.1 roadmap complete. The next milestone is tracked in `ROADMAP-v0.2.md`.

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
- [x] Add production packaging that excludes protected source by default
- [x] Explore encrypted source mode such as `app.black.enc`
