package main

import (
	"fmt"
	"strings"
)

func WebCoverageReport() CoverageResult {
	areas := []CoverageArea{
		{
			Name:   "data-model-crud",
			Weight: 10,
			Score:  100,
			Status: "mature-mvp",
			Implemented: []string{
				"entities and stored primitive fields",
				"file and image media fields",
				"generated local data-url media inputs",
				"relations",
				"relation response load policies",
				"single-hop bulk relation prefetch for generated responses",
				"field validation",
				"CRUD pages",
				"archive/restore/delete actions",
				"computed display fields",
			},
			Missing: []string{
				"multi-hop relation graph loading after explicit relation graph syntax",
			},
			Next: []string{"add multi-hop relation graph loading only after explicit relation graph syntax exists"},
		},
		{
			Name:   "frontend-ui-layout",
			Weight: 15,
			Score:  100,
			Status: "mature-mvp",
			Implemented: []string{
				"React generation",
				"table/detail/form screens",
				"page view order",
				"page view stack/grid composition",
				"page view tabs composition",
				"page view modal/drawer section display",
				"page view nested section groups",
				"page view interaction triggers",
				"page view reusable component sections",
				"page view collection-backed component sections",
				"page view DOM-order rendering",
				"responsive grid breakpoint ladder",
				"accessibility audit checks for overlays and groups",
				"runtime language switching for field labels",
				"runtime placeholder, help, and field message translations",
				"runtime app chrome, table/status, CRUD/custom action, and workflow transition copy labels",
				"locale-aware number, integer, decimal, money, date, and datetime display formatting",
				"basic lang/dir support for RTL locales",
				"component declarations",
				"state declarations",
				".blackthm profiles",
				"inline UI intent",
				"theme migration safety checks",
			},
			Missing: []string{
				"client-side local expression state after Core Program Logic v1",
			},
			Next: []string{"add client-side local expression state in the separate Core Program Logic v1 track"},
		},
		{
			Name:   "backend-api-actions",
			Weight: 15,
			Score:  100,
			Status: "mature-mvp",
			Implemented: []string{
				"generated Express API",
				"API-only target generation",
				"CRUD routes",
				"custom query routes",
				"custom query aggregate summary routes",
				"custom row-level action routes",
				"bounded action/API local value, compound condition, and if/else logic",
				"transactional custom action routes",
				"generated background query worker jobs",
				"explicit API declared-runtime routes",
				"explicit API typed body and bounded update handlers",
				"explicit API service grouping and generated service modules",
				"webhook acknowledgement routes",
				"OpenAPI output",
				"server-side validation",
			},
			Missing: []string{
				"cross-service command orchestration beyond explicit API grouping",
			},
			Next: []string{"add cross-service command orchestration after explicit API grouping stabilizes"},
		},
		{
			Name:   "auth-permissions-security",
			Weight: 15,
			Score:  100,
			Status: "mature-mvp",
			Implemented: []string{
				"email/password cookie auth",
				"roles and page access",
				"multiple roles per user",
				"tenant administration UI",
				"secret reference manifest and provider handoff plan",
				"read-only secret manager provider preflight execution",
				"signed compiler and release verification trust workflow",
				"release transparency log policy",
				"release key rotation policy",
				"field permission checks",
				"nested relation response field sanitization",
				"owner row policies",
				"tenant row policies",
				"policy-scoped list/query/detail/mutation routes",
				"policy field input stripping",
				"CSRF middleware",
				"audit log UI",
				"CORS intent",
				"secure headers",
				"rate limit",
				"source security scan",
				".black.enc protected source MVP",
			},
			Missing: []string{
				"provider-owned runtime value injection adapters",
			},
			Next: []string{"add runtime value injection adapters only after provider-specific secret fetch policy is explicit"},
		},
		{
			Name:   "database-runtime",
			Weight: 10,
			Score:  100,
			Status: "mature",
			Implemented: []string{
				"SQLite Prisma schema generation",
				"deterministic local database setup",
				"first-class seed/fixture declarations",
				"generated deterministic database seed runtime",
				"PostgreSQL Prisma provider generation",
				"PostgreSQL Prisma driver adapter generation",
				"PostgreSQL Docker Compose service generation",
				"MySQL Prisma provider generation",
				"MySQL PrismaMariaDb driver adapter generation",
				"MySQL Docker Compose service generation",
				"MySQL deterministic seed runtime through Prisma upsert",
				"relation fields",
				"entity index declarations",
				"generated Prisma indexes",
				"generated SQLite setup indexes",
				"read-only schema migration plan",
				"explicit entity/field rename syntax",
				"generated migration manifest and SQL previews",
				"generated SQLite/PostgreSQL rename apply runtime",
				"generated online migration runner with read-only plan and explicit apply modes",
				"generated Prisma transactions for custom actions",
				"top-level transaction blocks for custom actions and explicit API update handlers",
			},
			Missing: []string{
				"MySQL generated rename migration runner",
			},
			Next: []string{"add MySQL rename migration runner only after provider-specific DDL safeguards are specified"},
		},
		{
			Name:   "deployment-operations",
			Weight: 10,
			Score:  100,
			Status: "mature",
			Implemented: []string{
				"Dockerfile generation",
				"docker-compose generation",
				"environment declaration",
				"production package exclusion",
				"generated start script",
				"local preview deployment compose generation",
				"isolated preview port and preview SQLite/PostgreSQL data defaults",
				"deployment manifest metadata",
				"rollback keep metadata",
				"read-only rollback plan script",
				"health endpoint declarations",
				"readiness endpoint declarations",
				"metrics endpoint declarations",
				"structured request logging",
				"Docker app service healthchecks",
				"OpenAPI metadata for ops endpoints",
				"cloud adapter metadata for fly/render/railway",
				"read-only cloud deploy plan script",
				"generated provider CLI deploy preflight and explicit apply runner",
				"external observability webhook hook",
				"external OTLP trace exporter hook",
				"OpenAPI metadata for observability hooks",
				"generated local smoke tests for observability hook delivery",
				"W3C traceparent request context",
				"OTLP HTTP JSON trace exporter",
				"observability exporter manifest metadata",
			},
			Missing: []string{
				"vendor-specific tracing backend packages after public adapter marketplace publishing",
			},
			Next: []string{"publish provider-specific tracing adapters after the public adapter marketplace trust review"},
		},
		{
			Name:   "testing-benchmarks",
			Weight: 10,
			Score:  100,
			Status: "mature",
			Implemented: []string{
				"Go compiler tests",
				"generated contract smoke tests",
				"generated API smoke tests",
				"explicit API runtime smoke tests",
				"first-class browser-check test syntax",
				"generated browser-check smoke tests",
				"generated full browser e2e execution",
				"generated browser e2e matrix plan and runner",
				"real browser auth, navigation, text, page, and action checks",
				"first-class seed/fixture syntax",
				"generated seed runtime smoke tests",
				"generated frontend render smoke tests",
				"golden manifest drift test",
				"source/generated benchmark command",
				"AI task benchmark scenarios",
				"token estimate reports",
				"long-running AI eval corpus",
				"published-local AI eval result history manifest",
				"AI eval result history JSON and BlackIR export",
			},
			Missing: []string{
				"external multi-model scored eval board after independent harness runs",
			},
			Next: []string{"append external multi-model scored runs after independent harness execution and release review"},
		},
		{
			Name:   "docs-ai-ergonomics",
			Weight: 10,
			Score:  100,
			Status: "mature",
			Implemented: []string{
				"docs --all JSON",
				"docs keyword JSON/IR",
				"explain keyword JSON/IR",
				"inspect --affected",
				"agent startup checklist",
				"diagnostics reference",
				"compiler-owned IDE metadata export",
				"compiler-owned IDE completions and snippets",
				"editor-friendly IDE diagnostics with zero-based ranges",
				"packaged VS Code extension source",
				"IDE format quick fix workflow",
				"IDE affected-symbol inspection workflow",
				"llms.txt",
				"documentation site source",
				"coverage status on docs site",
				"tracked issue export",
				"multi-editor package channel metadata",
				"editor marketplace publishing docs",
			},
			Missing: []string{
				"external editor marketplace owner publication",
			},
			Next: []string{"publish editor marketplace packages after release-owner review and signed artifact verification"},
		},
		{
			Name:   "ecosystem-integrations",
			Weight: 5,
			Score:  95,
			Status: "strong-mvp",
			Implemented: []string{
				"install path docs",
				"release artifact docs",
				"npm wrapper plan",
				"thin npm wrapper package source",
				"npm wrapper detached signature verification for downloaded release archives",
				"ecosystem discovery command",
				"built-in adapter registry metadata",
				"packageable VS Code/Open VSX/Cursor-compatible editor channel metadata",
				"prepared local package registry manifest",
				"prepared local provider adapter marketplace manifest",
				"registry and marketplace trust policy manifests",
				"signed release, package, and provider adapter trust workflow metadata",
				"local registry/marketplace manifest validator",
				"ecosystem JSON/BlackIR registry and marketplace discovery",
				"prepared public ecosystem index manifest",
				"documentation site ecosystem index JSON source",
			},
			Missing: []string{
				"external npm/editor registry publication",
				"deployed public provider adapter marketplace index after release-owner review",
			},
			Next: []string{"publish prepared registries after release review", "deploy hosted provider adapter marketplace index after trust review"},
		},
	}

	return CoverageResult{
		Success:           true,
		Command:           "benchmark coverage",
		Version:           version,
		Target:            "web",
		CompletionPercent: weightedCoveragePercent(areas),
		Areas:             areas,
		Issues:            webCoverageIssues(),
		Milestones: []CoverageMilestone{
			{Percent: 25, Meaning: "basic generated CRUD/admin web app is useful"},
			{Percent: 50, Meaning: "core generated web app is measurable and AI-editable"},
			{Percent: 75, Meaning: "layout, i18n, database, security, testing, and deployment are mature enough for larger apps"},
			{Percent: 100, Meaning: "full web target with runtime breadth, operations, ecosystem, and policy coverage"},
		},
		Errors: []Diagnostic{},
	}
}

func weightedCoveragePercent(areas []CoverageArea) int {
	weighted := 0
	weight := 0
	for _, area := range areas {
		weighted += area.Weight * area.Score
		weight += area.Weight
	}
	if weight == 0 {
		return 0
	}
	return (weighted + weight/2) / weight
}

func webCoverageIssues() []CoverageIssue {
	return []CoverageIssue{
		{ID: "WEB-UI-001", Phase: "Phase 20", Area: "frontend-ui-layout", Title: "Add UI profile migration rules", Priority: "high", Status: "done"},
		{ID: "WEB-LAYOUT-001", Phase: "Phase 28", Area: "frontend-ui-layout", Title: "Add modal and drawer section display for generated detail/form panels", Priority: "high", Status: "done"},
		{ID: "WEB-LAYOUT-002", Phase: "Phase 28", Area: "frontend-ui-layout", Title: "Add nested section groups with local stack/grid composition", Priority: "high", Status: "done"},
		{ID: "WEB-LAYOUT-003", Phase: "Phase 28", Area: "frontend-ui-layout", Title: "Add accessibility audit policy checks for overlays and nested groups", Priority: "high", Status: "done"},
		{ID: "WEB-LAYOUT-004", Phase: "Phase 28", Area: "frontend-ui-layout", Title: "Add advanced interactive composition triggers", Priority: "high", Status: "done"},
		{ID: "WEB-LAYOUT-005", Phase: "Component composition phase", Area: "frontend-ui-layout", Title: "Add reusable page component sections bound to selected or first source record data", Priority: "high", Status: "done"},
		{ID: "WEB-LAYOUT-006", Phase: "DOM-order rendering phase", Area: "frontend-ui-layout", Title: "Add generated JSX DOM order that follows effective page view order", Priority: "medium", Status: "done"},
		{ID: "WEB-LAYOUT-007", Phase: "Component collection phase", Area: "frontend-ui-layout", Title: "Add collection-backed page component sections with bind each", Priority: "medium", Status: "done"},
		{ID: "WEB-I18N-001", Phase: "Phase 26", Area: "frontend-ui-layout", Title: "Add runtime language switching for field labels", Priority: "medium", Status: "done"},
		{ID: "WEB-I18N-002", Phase: "Phase 26", Area: "frontend-ui-layout", Title: "Add placeholder, help, message, date, number, currency, and RTL support", Priority: "medium", Status: "done"},
		{ID: "WEB-I18N-003", Phase: "Phase 26", Area: "frontend-ui-layout", Title: "Add localized app chrome/action/table/status copy catalog", Priority: "medium", Status: "done"},
		{ID: "WEB-DATA-001", Phase: "Phase 21", Area: "database-runtime", Title: "Add generated schema migration plan before destructive database changes", Priority: "high", Status: "done"},
		{ID: "WEB-DATA-002", Phase: "Database runtime phase", Area: "database-runtime", Title: "Add PostgreSQL runtime before enabling postgres deploy target", Priority: "high", Status: "done"},
		{ID: "WEB-DATA-003", Phase: "Data runtime phase", Area: "database-runtime", Title: "Add applied migration runtime and first-class rename/refactor syntax", Priority: "high", Status: "done"},
		{ID: "WEB-DATA-006", Phase: "Migration runner phase", Area: "database-runtime", Title: "Add generated online migration runner safeguards", Priority: "high", Status: "done"},
		{ID: "WEB-DATA-009", Phase: "Database provider phase", Area: "database-runtime", Title: "Add MySQL target runtime, seed, and Docker Compose provider output", Priority: "medium", Status: "done"},
		{ID: "WEB-DATA-007", Phase: "Data logic phase", Area: "data-model-crud", Title: "Add relation response load policies for generated list, detail, query, and mutation routes", Priority: "medium", Status: "done"},
		{ID: "WEB-DATA-008", Phase: "Relation prefetch phase", Area: "data-model-crud", Title: "Add single-hop bulk relation prefetch and nested relation response sanitization", Priority: "medium", Status: "done"},
		{ID: "WEB-SEC-001", Phase: "Phase 29", Area: "auth-permissions-security", Title: "Add ownership, tenant, and policy language", Priority: "high", Status: "done"},
		{ID: "WEB-SEC-002", Phase: "Auth runtime phase", Area: "auth-permissions-security", Title: "Add multiple roles per authenticated user", Priority: "high", Status: "done"},
		{ID: "WEB-SEC-003", Phase: "Auth runtime phase", Area: "auth-permissions-security", Title: "Add tenant administration UI for tenant policies", Priority: "high", Status: "done"},
		{ID: "WEB-SEC-004", Phase: "Secret management phase", Area: "auth-permissions-security", Title: "Add generated secret reference manifest and provider handoff plan", Priority: "high", Status: "done"},
		{ID: "WEB-SEC-007", Phase: "Secret provider phase", Area: "auth-permissions-security", Title: "Add read-only secret manager provider preflight execution", Priority: "high", Status: "done"},
		{ID: "WEB-SEC-005", Phase: "Release trust phase", Area: "auth-permissions-security", Title: "Add signed compiler and release verification trust workflow", Priority: "high", Status: "done"},
		{ID: "WEB-SEC-006", Phase: "Release trust phase", Area: "auth-permissions-security", Title: "Add release transparency log and key rotation policy metadata", Priority: "high", Status: "done"},
		{ID: "WEB-API-001", Phase: "Phase 21", Area: "backend-api-actions", Title: "Generate declared runtime routes for explicit API and webhook declarations", Priority: "high", Status: "done"},
		{ID: "WEB-API-002", Phase: "API runtime phase", Area: "backend-api-actions", Title: "Add explicit API handler syntax for business side effects", Priority: "high", Status: "done"},
		{ID: "WEB-API-003", Phase: "Phase 21", Area: "backend-api-actions", Title: "Add aggregate summaries for page-bound custom queries", Priority: "high", Status: "done"},
		{ID: "WEB-API-004", Phase: "Background jobs phase", Area: "backend-api-actions", Title: "Add deterministic background query worker jobs", Priority: "high", Status: "done"},
		{ID: "WEB-API-005", Phase: "API target phase", Area: "backend-api-actions", Title: "Add API-only target generation", Priority: "high", Status: "done"},
		{ID: "WEB-API-006", Phase: "Service block phase", Area: "backend-api-actions", Title: "Add service/module grouping for explicit API declarations", Priority: "high", Status: "done"},
		{ID: "WEB-DATA-004", Phase: "Phase 21", Area: "database-runtime", Title: "Add generated transaction mode for custom actions", Priority: "high", Status: "done"},
		{ID: "WEB-DATA-005", Phase: "Transaction block phase", Area: "database-runtime", Title: "Add top-level transaction blocks for custom actions and explicit API update handlers", Priority: "high", Status: "done"},
		{ID: "WEB-OPS-001", Phase: "Deployment operations phase", Area: "deployment-operations", Title: "Add local preview deployment target and rollback metadata", Priority: "medium", Status: "done"},
		{ID: "WEB-OPS-002", Phase: "Phase 30", Area: "deployment-operations", Title: "Add health/readiness/metrics probes, request logging, and Docker healthchecks", Priority: "medium", Status: "done"},
		{ID: "WEB-OPS-003", Phase: "Future deploy phase", Area: "deployment-operations", Title: "Add cloud adapters and external observability provider hooks", Priority: "medium", Status: "done"},
		{ID: "WEB-OPS-004", Phase: "Deployment operations phase", Area: "deployment-operations", Title: "Add generated provider CLI deployment preflight and explicit apply runner", Priority: "medium", Status: "done"},
		{ID: "WEB-OPS-005", Phase: "Observability exporter phase", Area: "deployment-operations", Title: "Add distributed tracing context and OTLP exporter metadata", Priority: "medium", Status: "done"},
		{ID: "WEB-TEST-001", Phase: "Phase 22", Area: "testing-benchmarks", Title: "Add AI task benchmark scenarios and token estimate reports", Priority: "medium", Status: "done"},
		{ID: "WEB-TEST-002", Phase: "Test/runtime phase", Area: "testing-benchmarks", Title: "Add first-class fixture and seed syntax", Priority: "medium", Status: "done"},
		{ID: "WEB-TEST-003", Phase: "Test/runtime phase", Area: "testing-benchmarks", Title: "Add first-class browser-check syntax and generated browser checks", Priority: "medium", Status: "done"},
		{ID: "WEB-TEST-004", Phase: "Future test phase", Area: "testing-benchmarks", Title: "Add full browser e2e execution with real browser automation", Priority: "medium", Status: "done"},
		{ID: "WEB-TEST-005", Phase: "Browser matrix phase", Area: "testing-benchmarks", Title: "Add generated cross-browser e2e matrix plan and runner", Priority: "medium", Status: "done"},
		{ID: "WEB-TEST-006", Phase: "AI eval phase", Area: "testing-benchmarks", Title: "Add long-running AI eval corpus metadata", Priority: "medium", Status: "done"},
		{ID: "WEB-TEST-007", Phase: "AI eval history phase", Area: "testing-benchmarks", Title: "Add evidence-backed AI eval result history export", Priority: "medium", Status: "done"},
		{ID: "WEB-DOCS-001", Phase: "Phase 27", Area: "docs-ai-ergonomics", Title: "Publish coverage status on the documentation site", Priority: "medium", Status: "done"},
		{ID: "WEB-DOCS-002", Phase: "Docs AI ergonomics phase", Area: "docs-ai-ergonomics", Title: "Export tracked web coverage issues as compact JSON and IR", Priority: "medium", Status: "done"},
		{ID: "WEB-DOCS-003", Phase: "Docs tooling phase", Area: "docs-ai-ergonomics", Title: "Add IDE diagnostics and autocomplete support", Priority: "medium", Status: "done"},
		{ID: "WEB-DOCS-004", Phase: "Docs tooling phase", Area: "docs-ai-ergonomics", Title: "Add IDE refactor workflows and packaged editor extension", Priority: "medium", Status: "done"},
		{ID: "WEB-DOCS-005", Phase: "Editor marketplace phase", Area: "docs-ai-ergonomics", Title: "Add multi-editor marketplace package channel metadata", Priority: "medium", Status: "done"},
		{ID: "WEB-ECO-001", Phase: "Ecosystem phase", Area: "ecosystem-integrations", Title: "Ship installable release artifacts and plugin/adapter discovery", Priority: "medium", Status: "done"},
		{ID: "WEB-ECO-002", Phase: "Ecosystem phase", Area: "ecosystem-integrations", Title: "Prepare package registries and provider adapter marketplace", Priority: "medium", Status: "done"},
		{ID: "WEB-ECO-003", Phase: "Release trust phase", Area: "ecosystem-integrations", Title: "Add signed package and provider adapter verification workflow", Priority: "medium", Status: "done"},
		{ID: "WEB-ECO-004", Phase: "Hosted index phase", Area: "ecosystem-integrations", Title: "Add prepared public ecosystem index source for hosted package and adapter discovery", Priority: "medium", Status: "done"},
	}
}

func WebCoverageIssuesReport() CoverageIssuesResult {
	coverage := WebCoverageReport()
	issues := append([]CoverageIssue(nil), coverage.Issues...)
	openIssues := []CoverageIssue{}
	summary := CoverageIssuesSummary{Total: len(issues)}
	for _, issue := range issues {
		switch issue.Status {
		case "open":
			summary.Open++
			openIssues = append(openIssues, issue)
		case "done":
			summary.Done++
		}
	}
	return CoverageIssuesResult{
		Success:           true,
		Command:           "benchmark issues",
		Version:           version,
		Target:            coverage.Target,
		CompletionPercent: coverage.CompletionPercent,
		Summary:           summary,
		Issues:            issues,
		OpenIssues:        openIssues,
		Errors:            []Diagnostic{},
	}
}

func FormatCoverageIR(result CoverageResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("benchmark coverage ok\n")
	} else {
		builder.WriteString("benchmark coverage failed\n")
	}
	builder.WriteString(fmt.Sprintf("target %s\n", result.Target))
	builder.WriteString(fmt.Sprintf("completion %d\n", result.CompletionPercent))
	builder.WriteString(fmt.Sprintf("areas %d\n", len(result.Areas)))
	for _, area := range result.Areas {
		builder.WriteString(fmt.Sprintf("  area %s weight %d score %d status %s\n", area.Name, area.Weight, area.Score, area.Status))
	}
	builder.WriteString(fmt.Sprintf("issues %d\n", len(result.Issues)))
	for _, issue := range result.Issues {
		builder.WriteString(fmt.Sprintf("  issue %s %s %s %s\n", issue.ID, issue.Area, issue.Priority, issue.Status))
	}
	return builder.String()
}

func FormatCoverageIssuesIR(result CoverageIssuesResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("benchmark issues ok\n")
	} else {
		builder.WriteString("benchmark issues failed\n")
	}
	builder.WriteString(fmt.Sprintf("target %s\n", result.Target))
	builder.WriteString(fmt.Sprintf("completion %d\n", result.CompletionPercent))
	builder.WriteString(fmt.Sprintf("summary total %d open %d done %d\n", result.Summary.Total, result.Summary.Open, result.Summary.Done))
	builder.WriteString(fmt.Sprintf("openIssues %d\n", len(result.OpenIssues)))
	for _, issue := range result.OpenIssues {
		builder.WriteString(fmt.Sprintf("  issue %s %s %s %s\n", issue.ID, issue.Area, issue.Priority, issue.Status))
	}
	return builder.String()
}
