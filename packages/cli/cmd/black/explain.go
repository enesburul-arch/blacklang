package main

import (
	"fmt"
	"strings"
)

var relatedDocsByKeyword = map[string][]string{
	"access":              {"page", "auth", "role", "policy"},
	"accessibility":       {"view", "page", "lint", "generated-test", "test"},
	"action":              {"actions", "page", "role", "query", "policy", "transaction", "openapi", "audit"},
	"actions":             {"action", "page", "api", "audit", "workflow"},
	"agent-contract":      {"agent", "docs", "syntax", "inspect"},
	"api":                 {"openapi", "auth", "security", "cors", "ops", "policy", "action", "transaction", "service"},
	"app":                 {"syntax", "entity", "page"},
	"audit":               {"auth", "role", "actions"},
	"auth":                {"role", "access", "policy", "csrf", "security", "cors"},
	"blackir":             {"parse", "validate", "inspect", "docs"},
	"component":           {"entity", "table", "form", "view", "state"},
	"computed":            {"entity", "table", "form", "validate"},
	"cors":                {"security", "api", "auth"},
	"coverage":            {"benchmark", "docs", "agent-contract"},
	"csrf":                {"auth", "security"},
	"database":            {"target", "security", "syntax", "index", "migrate", "seed"},
	"deploy":              {"target", "database", "security", "ops", "package", "adapter-marketplace", "release-trust"},
	"docs":                {"syntax", "explain", "ide", "lint", "agent-contract"},
	"ecosystem":           {"package", "package-registry", "adapter-marketplace", "editor-marketplace", "release-trust", "deploy", "ide", "target", "security"},
	"package-registry":    {"ecosystem", "package", "ide", "editor-marketplace", "release-trust"},
	"release-trust":       {"ecosystem", "package-registry", "adapter-marketplace", "editor-marketplace", "security", "deploy"},
	"adapter-marketplace": {"ecosystem", "deploy", "ops", "ide", "editor-marketplace", "security", "release-trust"},
	"editor-marketplace":  {"ecosystem", "ide", "package-registry", "adapter-marketplace", "release-trust"},
	"entity":              {"page", "table", "form", "component", "computed", "media", "index", "policy", "relation-load", "seed"},
	"explain":             {"docs", "ide", "lint", "syntax"},
	"filter":              {"table", "search", "paginate"},
	"format":              {"lint", "validate"},
	"form":                {"entity", "page", "actions", "view", "media", "placeholder", "help", "message"},
	"generated-test":      {"openapi", "validate", "api", "ops", "test", "job"},
	"i18n":                {"label", "placeholder", "help", "message", "app", "page", "form", "table", "action"},
	"ide":                 {"docs", "explain", "lint", "validate", "diagnostics", "agent", "editor-marketplace"},
	"index":               {"entity", "database", "query", "inspect", "migrate"},
	"label":               {"i18n", "entity", "computed", "app", "page", "action", "table"},
	"layout":              {"page", "view", "access"},
	"lint":                {"format", "validate", "security", "docs"},
	"migrate":             {"database", "entity", "index", "inspect", "benchmark", "seed"},
	"media":               {"entity", "form", "table", "openapi", "generated-test"},
	"job":                 {"query", "database", "generated-test", "openapi", "deploy"},
	"openapi":             {"api", "service", "page", "query", "job", "action", "media", "relation-load", "ops"},
	"ops":                 {"deploy", "target", "database", "openapi", "generated-test", "security"},
	"page":                {"entity", "table", "form", "actions", "layout", "view", "component", "access", "policy", "test"},
	"paginate":            {"table", "filter", "search"},
	"placeholder":         {"i18n", "form", "entity"},
	"policy":              {"entity", "auth", "role", "access", "query", "action", "security"},
	"query":               {"entity", "page", "table", "filter", "access", "policy", "relation-load", "job", "inspect"},
	"relation-load":       {"entity", "page", "query", "openapi", "policy"},
	"role":                {"auth", "access", "policy", "audit"},
	"search":              {"table", "filter"},
	"security":            {"database", "lint", "csrf", "cors", "auth", "policy", "deploy", "ops", "package"},
	"seed":                {"database", "entity", "migrate", "generated-test", "benchmark"},
	"help":                {"i18n", "form", "entity"},
	"message":             {"i18n", "form", "entity", "validate"},
	"state":               {"page", "component", "form"},
	"syntax":              {"app", "entity", "page", "docs", "agent-contract"},
	"table":               {"page", "entity", "computed", "media", "search", "filter", "paginate", "view"},
	"target":              {"deploy", "database", "api", "package"},
	"transaction":         {"action", "api", "openapi", "database", "audit", "inspect"},
	"service":             {"api", "openapi", "target", "inspect", "generated-test"},
	"test":                {"page", "generated-test", "benchmark", "accessibility"},
	"theme":               {"ui-profile", "ui-modes", "theme-migration", "ui"},
	"theme-migration":     {"theme", "ui-profile", "ui-modes"},
	"validate":            {"lint", "format"},
	"version":             {"docs", "lint"},
	"view":                {"page", "layout", "table", "form", "component", "ui"},
	"workflow":            {"entity", "page", "actions", "audit"},
}

func ExplainKeyword(keyword string) ExplainResult {
	key := strings.ToLower(strings.TrimSpace(keyword))
	if key == "" {
		key = "syntax"
	}

	doc, ok := FindDoc(key)
	result := ExplainResult{
		Success:    ok,
		Command:    "explain",
		Version:    version,
		Keyword:    key,
		AgentSteps: []string{},
		AgentNotes: []string{},
		Related:    []string{},
		ErrorCodes: []string{},
		Errors:     []Diagnostic{},
	}
	if !ok {
		result.Errors = []Diagnostic{{
			Code:       "UNKNOWN_EXPLAIN_KEYWORD",
			Message:    fmt.Sprintf("No explanation exists for %q.", key),
			Suggestion: "Use `black docs --all --json` to list supported keywords.",
		}}
		return result
	}

	result.Keyword = doc.Keyword
	result.Purpose = doc.Purpose
	result.Syntax = doc.Syntax
	result.Example = doc.Example
	result.AgentSteps = explainAgentSteps(doc)
	result.AgentNotes = append([]string(nil), doc.AgentNotes...)
	result.Related = explainRelatedKeywords(doc.Keyword)
	result.ErrorCodes = append([]string(nil), doc.Errors...)
	return result
}

func explainAgentSteps(doc DocEntry) []string {
	steps := []string{
		"Read purpose, syntax, example, and agentNotes before editing or generating BlackLang source.",
		"Use the exact syntax shape shown here; do not invent an alternate spelling for the same behavior.",
		"Prefer the smallest .black source change that expresses the requested intent.",
	}
	if len(doc.Errors) > 0 {
		steps = append(steps, "If lint or validate reports one of the listed errorCodes, fix the source and rerun black lint --json.")
	}
	if doc.Keyword == "query" {
		steps = append(steps,
			"Inspect the source entity and its stored field types, then declare one top-level query using typed literal where clauses.",
			"Bind the query with query Name inside a page whose source matches; inspect --affected Name and the filter/sort fields before changing it.",
			"Build and verify the page query endpoints: row policy, archive policy, filters, deterministic sort, limit, and aggregate summary when declared; check permitted and denied roles.",
		)
	}
	if doc.Keyword == "job" {
		steps = append(steps,
			"Inspect the query that the job will run; keep filter, sort, and limit deterministic before binding it to a worker.",
			"Declare one top-level job with schedule every N minutes|hours|days and run query QueryName; avoid external scheduler or queue syntax in the core MVP.",
			"Run inspect --affected JobName --json and inspect --affected QueryName --json so generated worker, manifest, package script, and OpenAPI metadata impact is visible.",
			"Build and verify jobs/manifest.json, src/worker.ts, package jobs:run/jobs:loop scripts, OpenAPI x-blacklang-jobs metadata, and generated npm run jobs:run.",
		)
	}
	if doc.Keyword == "target" {
		steps = append(steps,
			"Choose target web when generated React UI is required; choose target api when the app should expose only the generated API/runtime surface.",
			"For target api, declare backend node and database sqlite, postgres, or mysql, and omit frontend react because no React or Vite output is generated.",
			"Run black inspect --affected target --json, build the app, and verify generated package scripts, OpenAPI, server routes, and omitted frontend/browser files.",
			"Keep browser test declarations on target web projects; target api reports UNSUPPORTED_API_TARGET_TEST.",
		)
	}
	if doc.Keyword == "service" {
		steps = append(steps,
			"Inspect each explicit API before adding the service so method, path, access, handler, webhook, and transaction metadata are clear.",
			"Declare one top-level service with one `api APIName` target per line; do not nest api blocks inside service.",
			"Keep each explicit API bound to at most one service block so generated module metadata and inspect --affected output stay unambiguous.",
			"Run inspect --affected ServiceName --json, build, and verify services/manifest.json, src/services/<service>.ts, OpenAPI x-blacklang-services, route-level x-blacklang-service, and generated contract assertions.",
		)
	}
	if doc.Keyword == "policy" {
		steps = append(steps,
			"Inspect the entity, auth block, roles, forms, actions, and existing ownerId or tenantId fields before adding row policy intent.",
			"Declare one stored text required non-unique field per policy, then add policy owner field or policy tenant field inside the entity.",
			"Keep policy fields out of form fields and custom action set clauses; generated routes stamp and filter them from authenticated user context.",
			"Run inspect --affected policy --json, build, and verify generated route where clauses, API input omission, OpenAPI input omission, auth tenant output, and Users page tenant editing when tenant policies exist.",
		)
	}
	if doc.Keyword == "index" {
		steps = append(steps,
			"Inspect the entity and choose stored fields that are used by queries, filters, sorts, or relation lookups.",
			"Declare index field or index fieldA, fieldB inside the entity; do not index computed display fields.",
			"Run black inspect --affected Entity.index --json and verify Prisma schema plus SQLite setup output after build.",
		)
	}
	if doc.Keyword == "relation-load" {
		steps = append(steps,
			"Inspect the relation field and identify which generated response contexts need the relation object instead of only the generated relation ID.",
			"Add one load modifier to the relation field with list, detail, query, mutation, or use load none by itself.",
			"Run black inspect --affected Entity.relationField --json, build, and verify generated route Prisma include clauses plus OpenAPI x-blacklang-relation-load metadata.",
		)
	}
	if doc.Keyword == "migrate" {
		steps = append(steps,
			"Save or obtain the old .black source and compare it with the proposed new source using black migrate plan old.black new.black --json.",
			"Treat success=false as invalid input; treat safe=false with success=true as a valid plan that needs manual/destructive review.",
			"After build, run generated npm run db:migrate:plan in the target environment to inspect database reachability, pending rename operations, ledger state, and unsafe checks.",
			"Run generated npm run db:migrate only when the plan is ready, then run npm run db:setup so the normal generated setup and seed path runs.",
		)
	}
	if doc.Keyword == "seed" {
		steps = append(steps,
			"Inspect the source entity, required/default fields, unique fields, relation fields, policies, and existing seed row keys before adding fixture rows.",
			"Declare one top-level seed block per entity dataset, then add row Key blocks with scalar field literals and relation refs.",
			"Keep row keys stable because generated db:seed uses them as id values and reruns idempotent upserts.",
			"Run inspect --affected SeedName --json, format/lint/validate, build, generated npm run build, and generated npm run db:setup.",
		)
	}
	if doc.Keyword == "test" {
		steps = append(steps,
			"Inspect the target page, page actions, i18n labels, and generated-test docs before adding browser-check expectations.",
			"Declare one top-level test with page PageName and expect text/page/action lines; keep expectations deterministic and generated from source intent.",
			"Run inspect --affected TestName --json, format/lint/validate, build, generated npm run build, generated npm test, generated npm run test:e2e:plan, and generated npm run test:e2e:matrix when a local browser executable is available.",
		)
	}
	if doc.Keyword == "action" {
		steps = append(steps,
			"Inspect the source entity, stored primitive fields, roles, and page actions before adding or editing a custom action.",
			"Declare one top-level action with source, primitive inputs, deterministic set clauses, optional allow, and optional success message.",
			"Bind the action by PascalCase name inside a page actions list with the same source entity; do not add another syntax or manual generated files.",
			"Use a top-level transaction block if the generated action route must run lookup, update, and audit logging atomically.",
			"Run inspect --affected ActionName --json and verify generated route, API client, OpenAPI operation, validation function, and page control.",
		)
	}
	if doc.Keyword == "transaction" {
		steps = append(steps,
			"Inspect each target action/API before adding the transaction so source entities, update handlers, permissions, and generated endpoints are clear.",
			"Declare one top-level transaction with action ActionName and/or api ApiName lines; do not put transaction inside the action or api block.",
			"Keep each action/API target bound to at most one transaction block and target only APIs that declare update handlers.",
			"Run inspect --affected TransactionName --json, build, and verify generated Prisma $transaction wrappers plus OpenAPI x-blacklang-transaction metadata.",
		)
	}
	if doc.Keyword == "media" {
		steps = append(steps,
			"Inspect the entity and page form/table usage, then add one file or image field with optional accept metadata.",
			"Use image for browser previews and file for generic downloads; keep provider storage, buckets, tokens, and upload credentials out of .black source.",
			"Build and verify generated Prisma string fields, file inputs, data URL validation, table/detail rendering, OpenAPI x-blacklang-media metadata, and generated tests.",
			"Run inspect --affected Entity.mediaField --json before editing a media field and avoid adding a second upload syntax for the same behavior.",
		)
	}
	if doc.Keyword == "api" {
		steps = append(steps,
			"Inspect existing generated routes, auth, row policies, and entity unique fields before adding an explicit API handler.",
			"Declare path params and typed body fields first, then add one bounded update handler using id or a unique stored field in the where clause.",
			"Keep handler writes to stored primitive non-policy fields; use body.name or param.name references and avoid raw SQL or JavaScript.",
			"Use a top-level transaction block when an explicit API update handler should run lookup, update, and audit logging atomically.",
			"For public tenant-scoped webhook handlers, include the tenant policy field as a required body or path value; use private auth for owner-scoped entities.",
			"Run inspect --affected ApiName --json, build, and verify generated server route, OpenAPI x-blacklang-handler metadata, and API smoke tests.",
		)
	}
	if doc.Keyword == "security" {
		steps = append(steps,
			"Run black security scan --json before packaging or deployment.",
			"Use black security encrypted-source --json to inspect protected source policy before introducing .black.enc files.",
			"When building from .black.enc source, set the header-declared key environment variable and verify build output contains no .black or .black.enc files.",
		)
	}
	if doc.Keyword == "deploy" {
		steps = append(steps,
			"Declare one top-level deploy block with target docker, optional port env default, deployment env references, optional preview local, optional rollback keep count, and optional cloud provider adapter metadata.",
			"Run black inspect --affected deploy --json to see generated Docker, compose, preview, rollback, cloud, package, and server artifacts before changing deployment intent.",
			"Build the generated app and verify Dockerfile, docker-compose.yml, docker-compose.preview.yml when preview local is declared, deploy/rollback.json plus scripts/rollback-plan.mjs when rollback keep is declared, and deploy/cloud.json plus scripts/cloud-plan.mjs plus scripts/cloud-exec.mjs when cloud is declared.",
			"Run npm run security:secrets:plan and npm run security:secrets:preflight to verify generated security/secrets.json references and provider CLI readiness before deployment without printing values.",
			"Use npm run deploy:rollback:plan for read-only rollback metadata inspection before any infrastructure mutation.",
			"Use npm run deploy:cloud:plan for read-only cloud metadata, npm run deploy:cloud:preflight for provider CLI readiness, and npm run deploy:cloud:exec only when apply-mode deployment is authorized.",
		)
	}
	if doc.Keyword == "ops" {
		steps = append(steps,
			"Choose root-level public probe paths and optional env-referenced webhook or OTLP observability intent inside one ops block.",
			"Run black inspect --affected ops --json to see generated server, OpenAPI, Docker, env, ops/observability.json, and test files impacted by the operation signals.",
			"Build the generated app and run npm test so the contract test checks OpenAPI/observability manifest metadata and the API smoke test probes health, readiness, metrics, trace context, and local observability delivery.",
			"When deploy target docker is present, verify Dockerfile and docker-compose.yml healthchecks point at the declared health path.",
			"Keep external observability endpoint values in environment variables and avoid hardcoded vendor URLs or tokens in .black source.",
		)
	}
	if doc.Keyword == "coverage" {
		steps = append(steps,
			"Run black benchmark coverage --json before making a broad web-completion claim.",
			"Use areas[].missing and areas[].next to choose the next compiler feature instead of guessing from memory.",
			"Run black benchmark issues --json when an agent or CI job only needs tracked issue totals and the current open issue list.",
			"Record completed issue IDs in roadmap docs and rerun coverage/issues after implementation.",
		)
	}
	if doc.Keyword == "ide" {
		steps = append(steps,
			"Run black ide --json to load compiler-owned extension metadata, completion items, snippets, and diagnostic codes.",
			"Run black ide diagnostics <file> --json for editor-friendly zero-based diagnostic ranges while keeping source files unchanged.",
			"Treat success=true with valid=false as a successful IDE analysis that found source diagnostics.",
			"Use completionItems and snippets from the compiler output instead of maintaining a stale editor-local copy.",
		)
	}
	if doc.Keyword == "ecosystem" {
		steps = append(steps,
			"Run black ecosystem --json before adding provider-specific behavior, release packaging, or editor/package integrations.",
			"Use packages[], registries[], adapters[], marketplaces[], publicIndex, and trustWorkflow to distinguish built-in core, local packageable source, prepared registry manifests, hosted index source, and provider adapter entries.",
			"Use registries[] to find package-index.blackdir, marketplaces[] to find adapter-index.blackdir, and publicIndex.path to find the static hosted-index source before publish/install work.",
			"Keep wrappers and extensions thin; they must call the native black CLI instead of reimplementing BlackLang parsing or validation.",
		)
	}
	if doc.Keyword == "package-registry" {
		steps = append(steps,
			"Run black ecosystem --json and read registries[] before package publishing work.",
			"Read packages/registry/package-index.blackdir, packages/registry/public-index.blackdir, and packages/registry/trust-policy.blackdir before changing npm, editor package, or hosted package index metadata.",
			"Run node packages/registry/scripts/validate-registry.mjs before any external package registry publish step.",
			"Verify release.blackdir, checksums.sha256, detached Ed25519 signatures, and a trusted release public key before enabling public downloads.",
		)
	}
	if doc.Keyword == "release-trust" {
		steps = append(steps,
			"Build release archives and write checksums only from trusted developer or CI machines.",
			"Create one detached Ed25519 signature next to each finalized archive using the canonical <artifact>.sig file name.",
			"Set BLACKLANG_RELEASE_PUBLIC_KEY or BLACKLANG_RELEASE_PUBLIC_KEY_FILE to the trusted public key; never store private signing keys in .black, manifests, or package metadata.",
			"Run node scripts/verify-release-trust.mjs <release-dir> --json --strict and require ready=true before GitHub Release, npm, editor, or provider adapter publishing.",
			"Check black ecosystem --json release.trust before changing signature algorithm, signature file pattern, or public key environment names.",
		)
	}
	if doc.Keyword == "adapter-marketplace" {
		steps = append(steps,
			"Run black ecosystem --json and read marketplaces[] before provider adapter work.",
			"Read adapters/marketplace/adapter-index.blackdir, adapters/marketplace/trust-policy.blackdir, and packages/registry/public-index.blackdir before adding provider-specific behavior or hosted adapter index metadata.",
			"Keep provider secrets, endpoints, DSNs, app names, regions, and tokens outside .black source.",
			"Add read-only manifest, plan, or preflight output, explicit apply mode, and signed package verification before any provider adapter is allowed to mutate external infrastructure.",
		)
	}
	if doc.Keyword == "editor-marketplace" {
		steps = append(steps,
			"Run black ecosystem --json and read packages[] plus adapters[] before editor package channel work.",
			"Read packages/registry/package-index.blackdir, adapters/marketplace/adapter-index.blackdir, docs/ide.md, and docs/editor-marketplace.md before changing editor publishing metadata.",
			"Run node packages/registry/scripts/validate-registry.mjs and cd editors/vscode-blacklang && npm test before packaging.",
			"Build the shared VSIX artifact with cd editors/vscode-blacklang && npm run package:vsix when the packaging tool is available.",
			"Keep marketplace tokens, signing keys, and publishing credentials outside .black source, generated output, manifests, and package metadata.",
		)
	}
	if doc.Keyword == "view" {
		steps = append(steps,
			"Use compose grid for multi-column panels, compose stack for vertical panels, or compose tabs with tab <Name> sections ... for tabbed page sections.",
			"Use section <Name> component <Component> bind selected|first|each to render a declared component as a generated page panel when its scalar inputs match source fields.",
			"Use stackAt sm, md, lg, or none on compose grid; generated CSS emits deterministic named breakpoint rules and clamps oversized spans.",
			"Use section detail/form display modal or drawer when the generated panel should open as an overlay instead of staying inline.",
			"Use group <Name> sections detail, form compose stack|grid when contiguous inline sections should share a nested wrapper.",
			"Use trigger detail on rowSelect, trigger form on createStart|editStart, trigger detail|table on saveSuccess, or trigger table on close for supported generated interactions.",
			"When using compose tabs, assign every ordered section to exactly one tab and keep tab sections inline and ungrouped before building.",
		)
	}
	if doc.Keyword == "accessibility" {
		steps = append(steps,
			"Run black audit accessibility --json after changing generated UI composition intent.",
			"Fix findings by editing .black view section or group declarations, then rerun audit and black build.",
			"Use explicit title text for generated dialog and grouped landmark names instead of relying on fallback labels.",
		)
	}
	if doc.Keyword == "i18n" {
		steps = append(steps,
			"Declare one i18n block with a default locale and every supported locale before adding translation blocks.",
			"Use label Entity.field for stored fields and computed display fields; use app.*, page.*, action.*, table.*, and status.* label targets for generated UI copy.",
			"Use placeholder, help, and message Entity.field for stored form fields.",
			"Build the web app and verify App.tsx includes a locale selector, lang/dir attributes, and pages receive locale when multiple locales are declared.",
			"Check generated pages for localized app chrome, navigation names, table tools, action buttons, field labels, placeholders, help text, field messages, and Intl value formatting.",
		)
	}
	if doc.Keyword == "placeholder" || doc.Keyword == "help" || doc.Keyword == "message" {
		steps = append(steps,
			"Use inline field modifiers for a single fallback text.",
			"Use top-level "+doc.Keyword+" Entity.field translation blocks when i18n declares multiple locales.",
			"Keep top-level "+doc.Keyword+" translations on stored fields because computed display fields have no form input.",
		)
	}
	if doc.Keyword == "label" {
		steps = append(steps,
			"Use inline field label modifiers for single-locale fallback text.",
			"Use top-level label Entity.field translation blocks when i18n declares multiple locales.",
			"Use label Entity.computedField to translate computed display labels without creating database columns or form inputs.",
			"Use app.*, page.*, action.*, table.*, and status.* label targets for generated app chrome, navigation, table tools, status copy, CRUD buttons, custom actions, and workflow transition buttons.",
		)
	}
	if doc.Keyword == "theme-migration" {
		steps = append(steps,
			"Run black theme inspect --json on the old and new theme when you need full profile metadata.",
			"Run black theme migrate <old.blackthm> <new.blackthm> --json before replacing a theme used by existing .black source.",
			"If UI_SLOT_MIGRATION_BREAK appears, keep old slots as the exact prefix and append any new slots at the end.",
		)
	}
	steps = append(steps, "After source edits, run black format --check --json, black lint --json, and black validate --json.")
	return steps
}

func explainRelatedKeywords(keyword string) []string {
	candidates := relatedDocsByKeyword[strings.ToLower(keyword)]
	related := []string{}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		if candidate == "" || seen[candidate] || candidate == strings.ToLower(keyword) {
			continue
		}
		if _, ok := FindDoc(candidate); !ok {
			continue
		}
		related = append(related, candidate)
		seen[candidate] = true
	}
	return related
}
