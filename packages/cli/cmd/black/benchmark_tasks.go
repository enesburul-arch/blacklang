package main

import (
	"fmt"
	"math"
)

type aiTaskScenarioTemplate struct {
	ID                    string
	Name                  string
	Purpose               string
	Trigger               string
	ProjectEvidence       []string
	BlackLangEdits        []string
	GeneratedImpact       []string
	Commands              []string
	BlackLangContextLines int
	BlackLangEditLines    int
	DocsContextTokens     int
	GeneratedContextLines int
	GeneratedEditLines    int
	StackContextTokens    int
}

func BenchmarkTasks(args []string) AITaskBenchmarkResult {
	project := LoadProject(args)
	result := AITaskBenchmarkResult{
		Success:   false,
		Command:   "benchmark tasks",
		Version:   version,
		Config:    project.ConfigInfo(),
		Summary:   project.Summary(),
		Scenarios: []AITaskScenario{},
		Errors:    []Diagnostic{},
	}
	if len(project.Diagnostics) > 0 {
		result.Errors = project.Diagnostics
		return result
	}

	benchmark := BenchmarkProject(args)
	if !benchmark.Success {
		result.Errors = benchmark.Errors
		return result
	}

	sourceTokensPerLine := benchmarkTokensPerLine(benchmark.Source.Bytes, benchmark.Source.Lines)
	generatedTokensPerLine := benchmarkTokensPerLine(benchmark.Generated.Bytes, benchmark.Generated.Lines)
	result.Baseline = AITaskBenchmarkBaseline{
		SourceFiles:            benchmark.Source.Files,
		SourceLines:            benchmark.Source.Lines,
		SourceBytes:            benchmark.Source.Bytes,
		GeneratedFiles:         benchmark.Generated.Files,
		GeneratedLines:         benchmark.Generated.Lines,
		GeneratedBytes:         benchmark.Generated.Bytes,
		GeneratedToSourceLines: benchmark.Ratios.GeneratedToSourceLines,
		SourceTokensPerLine:    sourceTokensPerLine,
		GeneratedTokensPerLine: generatedTokensPerLine,
	}
	result.Scenarios = buildAITaskScenarios(project.Program, sourceTokensPerLine, generatedTokensPerLine)
	result.Totals = summarizeAITaskScenarios(result.Scenarios)
	result.Success = true
	return result
}

func buildAITaskScenarios(program Program, sourceTokensPerLine int, generatedTokensPerLine int) []AITaskScenario {
	templates := []aiTaskScenarioTemplate{
		{
			ID:      "AI-TASK-QUERY-001",
			Name:    "Add bounded report query",
			Purpose: "Measure the cost of adding a stored-field filtered, sorted, limited report page with aggregate summary cards.",
			Trigger: "A user asks for a Low Stock report page that lists products below a stock threshold and shows count/sum/average summary values.",
			ProjectEvidence: []string{
				fmt.Sprintf("queries:%d", len(program.Queries)),
				fmt.Sprintf("pages:%d", len(program.Pages)),
			},
			BlackLangEdits: []string{
				"Add one top-level query with source, where, aggregate, sort, and limit lines.",
				"Bind the query from one page with the same source entity.",
			},
			GeneratedImpact: []string{
				"API route and API client for the query",
				"OpenAPI query and summary paths",
				"React page summary cards and query-backed table refresh",
				"Generated contract/API/frontend smoke tests",
			},
			Commands: []string{
				"black docs query --json",
				"black explain query --json",
				"black inspect --affected QueryName --json",
				"black build --json",
				"cd generated && npm test",
			},
			BlackLangContextLines: 90,
			BlackLangEditLines:    14,
			DocsContextTokens:     850,
			GeneratedContextLines: 720,
			GeneratedEditLines:    900,
			StackContextTokens:    1200,
		},
		{
			ID:      "AI-TASK-ACTION-001",
			Name:    "Add row-level custom action",
			Purpose: "Measure the cost of adding a validated row mutation exposed as a page action, API route, OpenAPI operation, and React control.",
			Trigger: "A user asks for a Restock action that accepts a quantity and increments product stock through the generated page.",
			ProjectEvidence: []string{
				fmt.Sprintf("actions:%d", len(program.Actions)),
				fmt.Sprintf("pages:%d", len(program.Pages)),
			},
			BlackLangEdits: []string{
				"Add one top-level action with source, optional transaction, input, set, allow, and success lines.",
				"Expose the action from a matching page actions list.",
			},
			GeneratedImpact: []string{
				"API route and API client for the action",
				"Server-side action input validation and permission guards",
				"OpenAPI action metadata and transaction flag",
				"React row action button and action form panel",
				"Generated contract/API/frontend smoke tests",
			},
			Commands: []string{
				"black docs action --json",
				"black explain action --json",
				"black inspect --affected ActionName --json",
				"black build --json",
				"cd generated && npm test",
			},
			BlackLangContextLines: 90,
			BlackLangEditLines:    12,
			DocsContextTokens:     900,
			GeneratedContextLines: 820,
			GeneratedEditLines:    760,
			StackContextTokens:    1300,
		},
		{
			ID:      "AI-TASK-SEED-001",
			Name:    "Add deterministic fixture rows",
			Purpose: "Measure the cost of adding local/demo data with stable row keys and idempotent generated setup.",
			Trigger: "A user asks for demo Product, Customer, and Order records that can be reapplied during local setup.",
			ProjectEvidence: []string{
				fmt.Sprintf("seeds:%d", len(program.Seeds)),
				fmt.Sprintf("entities:%d", len(program.Entities)),
			},
			BlackLangEdits: []string{
				"Add one seed block per entity dataset.",
				"Use row keys, typed scalar literals, and relation refs.",
			},
			GeneratedImpact: []string{
				"Generated src/seed.ts runtime",
				"package.json db:seed and db:setup wiring",
				"Database setup verification path",
			},
			Commands: []string{
				"black docs seed --json",
				"black explain seed --json",
				"black inspect --affected SeedName --json",
				"black build --json",
				"cd generated && npm run db:setup",
			},
			BlackLangContextLines: 70,
			BlackLangEditLines:    22,
			DocsContextTokens:     700,
			GeneratedContextLines: 260,
			GeneratedEditLines:    150,
			StackContextTokens:    700,
		},
		{
			ID:      "AI-TASK-TEST-001",
			Name:    "Add browser test expectations",
			Purpose: "Measure the cost of adding deterministic page/action/text checks for generated browser-check and real browser e2e output.",
			Trigger: "A user asks for a browser check that Products exists, LowStock remains linked, and RestockProduct appears as an action in the generated app.",
			ProjectEvidence: []string{
				fmt.Sprintf("tests:%d", len(program.Tests)),
				fmt.Sprintf("pages:%d", len(program.Pages)),
			},
			BlackLangEdits: []string{
				"Add one top-level test with a page target.",
				"Add expect text, expect page, and expect action lines.",
			},
			GeneratedImpact: []string{
				"Generated src/blacklang.browser.test.tsx",
				"Generated src/blacklang.e2e.test.ts",
				"Generated src/blacklang.e2e.matrix.ts",
				"Generated tests/browser-matrix.json",
				"package.json npm test browser-check step",
				"package.json test:e2e, test:e2e:plan, test:e2e:matrix, and test:all scripts",
			},
			Commands: []string{
				"black docs test --json",
				"black explain test --json",
				"black inspect --affected TestName --json",
				"black build --json",
				"cd generated && npm test",
				"cd generated && npm run test:e2e",
				"cd generated && npm run test:e2e:plan",
				"cd generated && npm run test:e2e:matrix",
			},
			BlackLangContextLines: 45,
			BlackLangEditLines:    7,
			DocsContextTokens:     500,
			GeneratedContextLines: 410,
			GeneratedEditLines:    583,
			StackContextTokens:    700,
		},
		{
			ID:      "AI-TASK-POLICY-001",
			Name:    "Add owner or tenant row policy",
			Purpose: "Measure the cost of scoping generated list, query, detail, mutation, workflow, and custom action routes by authenticated user context.",
			Trigger: "A user asks for records to be isolated by tenant or owner across generated pages and APIs.",
			ProjectEvidence: []string{
				fmt.Sprintf("auth:%t", program.Auth != nil),
				fmt.Sprintf("entities:%d", len(program.Entities)),
			},
			BlackLangEdits: []string{
				"Add one stored text required policy field.",
				"Add policy owner or policy tenant inside the entity.",
				"Keep policy fields out of form fields and action set clauses.",
			},
			GeneratedImpact: []string{
				"Route-level row scope filters",
				"Input stripping and create/update stamping",
				"Auth session tenant output",
				"OpenAPI input omission and generated permission tests",
			},
			Commands: []string{
				"black docs policy --json",
				"black explain policy --json",
				"black inspect --affected policy --json",
				"black build --json",
				"cd generated && npm test",
			},
			BlackLangContextLines: 130,
			BlackLangEditLines:    10,
			DocsContextTokens:     900,
			GeneratedContextLines: 980,
			GeneratedEditLines:    1100,
			StackContextTokens:    1500,
		},
		{
			ID:      "AI-TASK-API-001",
			Name:    "Add explicit webhook update handler",
			Purpose: "Measure the cost of adding a safe declared-runtime API route with typed body validation and a bounded update handler.",
			Trigger: "A user asks for a stock webhook that receives SKU and quantity, validates the body, and updates one matching Product row.",
			ProjectEvidence: []string{
				fmt.Sprintf("apis:%d", len(program.APIs)),
				fmt.Sprintf("entities:%d", len(program.Entities)),
			},
			BlackLangEdits: []string{
				"Add one top-level api block with method, path, body, update, respond, and access lines.",
				"Use id or unique stored field in the update where clause.",
			},
			GeneratedImpact: []string{
				"Generated Express declared-runtime route",
				"Typed request body validation",
				"OpenAPI body schema and x-blacklang-update metadata",
				"Generated API smoke validation probe",
			},
			Commands: []string{
				"black docs api --json",
				"black explain api --json",
				"black inspect --affected ApiName --json",
				"black build --json",
				"cd generated && npm test",
			},
			BlackLangContextLines: 115,
			BlackLangEditLines:    11,
			DocsContextTokens:     1000,
			GeneratedContextLines: 760,
			GeneratedEditLines:    680,
			StackContextTokens:    1300,
		},
		{
			ID:      "AI-TASK-OPS-001",
			Name:    "Add generated ops probes",
			Purpose: "Measure the cost of adding health, readiness, metrics, request logging, OpenAPI ops metadata, and Docker healthchecks.",
			Trigger: "A user asks for production probes so a VPS or container orchestrator can check app health.",
			ProjectEvidence: []string{
				fmt.Sprintf("ops:%t", program.Ops != nil),
				fmt.Sprintf("deploy:%t", program.Deploy != nil),
			},
			BlackLangEdits: []string{
				"Add one top-level ops block with health, readiness, metrics, and logging declarations.",
			},
			GeneratedImpact: []string{
				"Express health/readiness/metrics routes",
				"OpenAPI ops metadata",
				"Dockerfile and docker-compose healthchecks",
				"Generated API smoke probes",
			},
			Commands: []string{
				"black docs ops --json",
				"black explain ops --json",
				"black inspect --affected ops --json",
				"black build --json",
				"cd generated && npm test",
			},
			BlackLangContextLines: 55,
			BlackLangEditLines:    6,
			DocsContextTokens:     650,
			GeneratedContextLines: 430,
			GeneratedEditLines:    360,
			StackContextTokens:    800,
		},
		{
			ID:      "AI-TASK-DEPLOY-001",
			Name:    "Add local preview and rollback metadata",
			Purpose: "Measure the cost of adding deterministic local preview deployment and rollback planning output.",
			Trigger: "A user asks for a local preview deployment stack and a rollback plan that keeps the newest three releases.",
			ProjectEvidence: []string{
				fmt.Sprintf("deploy:%t", program.Deploy != nil),
				fmt.Sprintf("database:%s", benchmarkTargetDatabase(program)),
			},
			BlackLangEdits: []string{
				"Add preview local inside the existing deploy block.",
				"Add rollback keep 3 inside the existing deploy block.",
			},
			GeneratedImpact: []string{
				"Generated docker-compose.preview.yml",
				"Generated deploy manifest and rollback metadata",
				"Generated read-only rollback plan script",
				"package.json preview and rollback scripts",
			},
			Commands: []string{
				"black docs deploy --json",
				"black explain deploy --json",
				"black inspect --affected deploy --json",
				"black build --json",
				"cd generated && npm run deploy:rollback:plan",
			},
			BlackLangContextLines: 45,
			BlackLangEditLines:    2,
			DocsContextTokens:     620,
			GeneratedContextLines: 150,
			GeneratedEditLines:    120,
			StackContextTokens:    500,
		},
	}

	scenarios := make([]AITaskScenario, 0, len(templates))
	for _, template := range templates {
		scenarios = append(scenarios, AITaskScenario{
			ID:              template.ID,
			Name:            template.Name,
			Purpose:         template.Purpose,
			Trigger:         template.Trigger,
			ProjectEvidence: append([]string(nil), template.ProjectEvidence...),
			BlackLangEdits:  append([]string(nil), template.BlackLangEdits...),
			GeneratedImpact: append([]string(nil), template.GeneratedImpact...),
			Commands:        append([]string(nil), template.Commands...),
			EstimatedTokens: estimateAITaskTokens(template, sourceTokensPerLine, generatedTokensPerLine),
		})
	}
	return scenarios
}

func benchmarkTargetDatabase(program Program) string {
	if program.Target != nil && program.Target.Database != "" {
		return program.Target.Database
	}
	return "sqlite"
}

func estimateAITaskTokens(template aiTaskScenarioTemplate, sourceTokensPerLine int, generatedTokensPerLine int) AITokenEstimate {
	commandTokens := estimateTextTokens(template.Trigger)
	for _, command := range template.Commands {
		commandTokens += estimateTextTokens(command)
	}
	for _, edit := range template.BlackLangEdits {
		commandTokens += estimateTextTokens(edit)
	}
	blackLangInput := template.BlackLangContextLines*sourceTokensPerLine + template.DocsContextTokens + commandTokens
	blackLangOutput := template.BlackLangEditLines * sourceTokensPerLine
	conventionalInput := template.GeneratedContextLines*generatedTokensPerLine + template.StackContextTokens + estimateTextTokens(template.Trigger)
	conventionalOutput := template.GeneratedEditLines * generatedTokensPerLine
	return AITokenEstimate{
		BlackLangInput:          blackLangInput,
		BlackLangOutput:         blackLangOutput,
		BlackLangTotal:          blackLangInput + blackLangOutput,
		ConventionalInput:       conventionalInput,
		ConventionalOutput:      conventionalOutput,
		ConventionalTotal:       conventionalInput + conventionalOutput,
		EstimatedSavingsPercent: tokenSavingsPercent(blackLangInput+blackLangOutput, conventionalInput+conventionalOutput),
		Basis: []string{
			fmt.Sprintf("sourceTokensPerLine=%d from current source bytes/lines", sourceTokensPerLine),
			fmt.Sprintf("generatedTokensPerLine=%d from current generated bytes/lines", generatedTokensPerLine),
			"blackLangInput = source context lines + focused docs/context + command/edit prompt text",
			"blackLangOutput = estimated .black edit lines",
			"conventionalInput = generated stack context lines + stack setup context + task text",
			"conventionalOutput = estimated generated stack edit lines",
		},
	}
}

func benchmarkTokensPerLine(bytes int64, lines int) int {
	if lines <= 0 || bytes <= 0 {
		return 8
	}
	value := int(math.Ceil((float64(bytes) / float64(lines)) / 4.0))
	if value < 1 {
		return 1
	}
	return value
}

func estimateTextTokens(text string) int {
	if text == "" {
		return 0
	}
	return int(math.Ceil(float64(len(text)) / 4.0))
}

func tokenSavingsPercent(blackLangTotal int, conventionalTotal int) int {
	if conventionalTotal <= 0 || blackLangTotal >= conventionalTotal {
		return 0
	}
	return int(math.Round((1 - float64(blackLangTotal)/float64(conventionalTotal)) * 100))
}

func summarizeAITaskScenarios(scenarios []AITaskScenario) AITaskBenchmarkTotals {
	totals := AITaskBenchmarkTotals{ScenarioCount: len(scenarios)}
	for _, scenario := range scenarios {
		totals.BlackLangTotal += scenario.EstimatedTokens.BlackLangTotal
		totals.ConventionalTotal += scenario.EstimatedTokens.ConventionalTotal
	}
	totals.EstimatedSavingsPercent = tokenSavingsPercent(totals.BlackLangTotal, totals.ConventionalTotal)
	return totals
}
