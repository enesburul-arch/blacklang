package main

import (
	"fmt"
	"sort"
	"strings"
)

func IDESupport() IDESupportResult {
	return IDESupportResult{
		Success: true,
		Command: "ide",
		Version: version,
		Language: IDELanguageInfo{
			ID:              "blacklang",
			Name:            "BlackLang",
			LanguageVersion: "0.1",
			Extensions:      []string{".black", ".blackthm"},
			CommentPrefix:   "#",
		},
		Capabilities: IDECapabilities{
			Diagnostics:       "black ide diagnostics <file> --json",
			CompletionItems:   true,
			Snippets:          true,
			DiagnosticCodes:   true,
			PositionBase:      "zero-based",
			RangeEndExclusive: true,
		},
		Commands:        ideCommands(),
		CompletionItems: ideCompletionItems(),
		Snippets:        ideSnippets(),
		DiagnosticCodes: ideDiagnosticCodes(),
		Errors:          []Diagnostic{},
	}
}

func IDEDiagnosticsFile(file string) IDEDiagnosticsResult {
	result := IDEDiagnosticsResult{
		Success:     true,
		Command:     "ide diagnostics",
		Version:     version,
		File:        file,
		Valid:       true,
		Project:     Summary{},
		Summary:     IDEDiagnosticsSummary{},
		Diagnostics: []IDEDiagnostic{},
		Errors:      []Diagnostic{},
	}

	source, readDiagnostics := ReadBlackSource(file)
	if len(readDiagnostics) > 0 {
		result.Success = false
		result.Valid = false
		result.Diagnostics = append(result.Diagnostics, convertIDEDiagnostics("read", readDiagnostics)...)
		result.Errors = readDiagnostics
		result.Summary = summarizeIDEDiagnostics(result.Diagnostics)
		return result
	}

	formatted, formatDiagnostics := FormatBlackSource(file, source)
	formatFindings := append([]Diagnostic{}, formatDiagnostics...)
	if len(formatDiagnostics) == 0 && formatted != source {
		suggestion := "Run `black format " + file + "`."
		if isEncryptedSourcePath(file) {
			suggestion = "Decrypt to a trusted plaintext workspace, run `black format`, then re-encrypt with `black security encrypt`."
		}
		formatFindings = append(formatFindings, Diagnostic{
			File:       file,
			Code:       "FORMAT_REQUIRED",
			Message:    "BlackLang source is not formatted.",
			Suggestion: suggestion,
		})
	}
	result.Diagnostics = append(result.Diagnostics, convertIDEDiagnostics("format", formatFindings)...)

	program, parseDiagnostics := Parse(file, source)
	result.Diagnostics = append(result.Diagnostics, convertIDEDiagnostics("parse", parseDiagnostics)...)
	if len(parseDiagnostics) == 0 {
		result.Project = Summary{
			App:      program.App.Name,
			Entities: len(program.Entities),
			Pages:    len(program.Pages),
		}
		result.Diagnostics = append(result.Diagnostics, convertIDEDiagnostics("validate", Validate(program))...)
	}
	result.Diagnostics = append(result.Diagnostics, convertIDEDiagnostics("security", SecurityScanText(file, source))...)
	result.Summary = summarizeIDEDiagnostics(result.Diagnostics)
	result.Valid = result.Summary.Total == 0
	return result
}

func FormatIDESupportIR(result IDESupportResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("ide ok\n")
	} else {
		builder.WriteString("ide failed\n")
	}
	builder.WriteString(fmt.Sprintf("language %s version %s extensions %s\n", result.Language.ID, result.Language.LanguageVersion, strings.Join(result.Language.Extensions, ",")))
	builder.WriteString(fmt.Sprintf("diagnosticsCommand %q\n", result.Capabilities.Diagnostics))
	builder.WriteString(fmt.Sprintf("completionItems %d\n", len(result.CompletionItems)))
	for _, item := range result.CompletionItems {
		builder.WriteString(fmt.Sprintf("  completion %s %s %s\n", item.Context, item.Label, item.Kind))
	}
	builder.WriteString(fmt.Sprintf("snippets %d\n", len(result.Snippets)))
	for _, snippet := range result.Snippets {
		builder.WriteString(fmt.Sprintf("  snippet %s %s\n", snippet.Context, snippet.Prefix))
	}
	builder.WriteString(fmt.Sprintf("diagnosticCodes %d\n", len(result.DiagnosticCodes)))
	for _, diagnostic := range result.DiagnosticCodes {
		builder.WriteString(fmt.Sprintf("  diagnostic %s %s\n", diagnostic.Code, diagnostic.Severity))
	}
	return builder.String()
}

func FormatIDEDiagnosticsIR(result IDEDiagnosticsResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("ide diagnostics ok\n")
	} else {
		builder.WriteString("ide diagnostics failed\n")
	}
	builder.WriteString(fmt.Sprintf("file %s\n", result.File))
	builder.WriteString(fmt.Sprintf("valid %t\n", result.Valid))
	builder.WriteString(fmt.Sprintf("summary diagnostics %d errors %d warnings %d\n", result.Summary.Total, result.Summary.Errors, result.Summary.Warnings))
	for _, diagnostic := range result.Diagnostics {
		builder.WriteString(fmt.Sprintf(
			"diagnostic %s %s %d:%d-%d:%d %q\n",
			diagnostic.Severity,
			diagnostic.Code,
			diagnostic.Range.Start.Line,
			diagnostic.Range.Start.Character,
			diagnostic.Range.End.Line,
			diagnostic.Range.End.Character,
			diagnostic.Message,
		))
	}
	return builder.String()
}

func convertIDEDiagnostics(source string, diagnostics []Diagnostic) []IDEDiagnostic {
	result := make([]IDEDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		startLine := diagnostic.Line - 1
		startCharacter := diagnostic.Column - 1
		if startLine < 0 {
			startLine = 0
		}
		if startCharacter < 0 {
			startCharacter = 0
		}
		endCharacter := startCharacter + 1
		result = append(result, IDEDiagnostic{
			File:       diagnostic.File,
			Range:      IDERange{Start: IDEPosition{Line: startLine, Character: startCharacter}, End: IDEPosition{Line: startLine, Character: endCharacter}},
			Severity:   ideDiagnosticSeverity(diagnostic.Code),
			Code:       diagnostic.Code,
			Message:    diagnostic.Message,
			Suggestion: diagnostic.Suggestion,
			Source:     source,
		})
	}
	return result
}

func summarizeIDEDiagnostics(diagnostics []IDEDiagnostic) IDEDiagnosticsSummary {
	summary := IDEDiagnosticsSummary{Total: len(diagnostics)}
	for _, diagnostic := range diagnostics {
		switch diagnostic.Severity {
		case "warning":
			summary.Warnings++
		default:
			summary.Errors++
		}
	}
	return summary
}

func ideDiagnosticSeverity(code string) string {
	if code == "FORMAT_REQUIRED" || strings.HasPrefix(code, "HARDCODED_") {
		return "warning"
	}
	return "error"
}

func ideCommands() []IDECommand {
	return []IDECommand{
		{Name: "manifest", Command: "black ide --json", Purpose: "Export IDE language metadata, completion items, snippets, and diagnostic code catalog."},
		{Name: "diagnostics", Command: "black ide diagnostics <file> --json", Purpose: "Return zero-based editor diagnostics for one .black source file."},
		{Name: "docs", Command: "black docs ide --json", Purpose: "Read the IDE contract docs from the local compiler."},
		{Name: "explain", Command: "black explain ide --json", Purpose: "Read action-oriented IDE setup guidance for agents."},
	}
}

func ideCompletionItems() []IDECompletionItem {
	items := []IDECompletionItem{}
	for _, keyword := range []string{
		"app", "target", "database", "security", "auth", "role", "entity", "query", "job", "action", "transaction", "service", "api",
		"migration", "seed", "test", "page", "workflow", "state", "component", "i18n", "label",
		"placeholder", "help", "message", "deploy", "ops",
	} {
		doc, ok := FindDoc(keyword)
		documentation := ""
		if ok {
			documentation = doc.Purpose
		}
		items = append(items, IDECompletionItem{
			Label:         keyword,
			Kind:          "keyword",
			Context:       "top-level",
			Detail:        "BlackLang top-level declaration",
			InsertText:    keyword,
			Documentation: documentation,
		})
	}
	for _, keyword := range []string{
		"computed", "source", "table", "form", "actions", "view", "compose", "section", "component", "bind", "group", "trigger",
		"where", "sort", "limit", "aggregate", "schedule", "every", "run", "input", "set", "allow", "success",
		"path", "method", "param", "body", "respond", "update", "frontend", "backend", "url", "env",
		"preview", "rollback", "cloud", "region", "health", "readiness", "metrics", "logging", "observe", "endpoint",
	} {
		items = append(items, IDECompletionItem{
			Label:         keyword,
			Kind:          "keyword",
			Context:       "block",
			Detail:        "BlackLang block keyword",
			InsertText:    keyword,
			Documentation: "Use in the matching BlackLang declaration block.",
		})
	}
	for _, value := range sortedMapKeys(supportedFieldTypes) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "type",
			Context:       "field-type",
			Detail:        "Stored field type",
			InsertText:    value,
			Documentation: "Supported primitive field type for stored fields, auth fields, API fields, state fields, and component inputs.",
		})
	}
	for _, value := range sortedMapKeys(supportedComputedFieldTypes) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "type",
			Context:       "computed-type",
			Detail:        "Computed display field type",
			InsertText:    value,
			Documentation: "Supported computed display field type.",
		})
	}
	for _, value := range sortedMapKeys(supportedFieldModifiers) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "modifier",
			Context:       "field-modifier",
			Detail:        "Stored field modifier",
			InsertText:    value,
			Documentation: "Supported modifier for stored fields and compatible typed inputs.",
		})
	}
	for _, value := range sortedMapKeys(supportedRelationLoadScopes) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "relation-load",
			Detail:        "Relation response load scope",
			InsertText:    value,
			Documentation: "Use after a relation field `load` modifier to select generated response contexts.",
		})
	}
	for _, value := range sortedMapKeys(supportedComparisonOperators) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "operator",
			Context:       "comparison",
			Detail:        "Comparison operator",
			InsertText:    value,
			Documentation: "Supported deterministic comparison operator.",
		})
	}
	for _, value := range sortedMapKeys(supportedTargetNames) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "target-name",
			Detail:        "Supported generated application target",
			InsertText:    value,
			Documentation: "Use in a top-level target declaration.",
		})
	}
	for _, value := range sortedMapKeys(supportedTargetDatabases) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "target-database",
			Detail:        "Supported generated database provider",
			InsertText:    value,
			Documentation: "Use after a target block `database` option.",
		})
	}
	for _, value := range sortedMapKeys(supportedViewComposeModes) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "view-compose",
			Detail:        "View compose mode",
			InsertText:    value,
			Documentation: "Supported generated page composition mode.",
		})
	}
	for _, value := range sortedMapKeys(supportedViewSectionDisplays) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "view-display",
			Detail:        "View section display mode",
			InsertText:    value,
			Documentation: "Supported generated detail/form display mode.",
		})
	}
	for _, value := range sortedMapKeys(supportedViewComponentBinds) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "view-component-bind",
			Detail:        "View component section bind mode",
			InsertText:    value,
			Documentation: "Use after a view component section `bind` option to choose selected, first, or each source record data.",
		})
	}
	for _, value := range sortedMapKeys(supportedDeployPreviewModes) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "deploy-preview",
			Detail:        "Deploy preview mode",
			InsertText:    value,
			Documentation: "Supported deploy preview mode.",
		})
	}
	for _, value := range sortedMapKeys(supportedDeployRollbackStrategies) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "deploy-rollback",
			Detail:        "Deploy rollback strategy",
			InsertText:    value,
			Documentation: "Supported deploy rollback metadata strategy.",
		})
	}
	for _, value := range sortedMapKeys(supportedDeployCloudProviders) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "deploy-cloud",
			Detail:        "Deploy cloud provider adapter",
			InsertText:    value,
			Documentation: "Supported cloud adapter provider for generated deploy metadata.",
		})
	}
	for _, value := range sortedMapKeys(supportedOpsObserveProviders) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "ops-observe",
			Detail:        "Ops observability provider",
			InsertText:    value,
			Documentation: "Supported observability hook provider.",
		})
	}
	for _, value := range sortedMapKeys(supportedJobScheduleUnits) {
		items = append(items, IDECompletionItem{
			Label:         value,
			Kind:          "value",
			Context:       "job-schedule",
			Detail:        "Job schedule unit",
			InsertText:    value,
			Documentation: "Supported generated background job schedule unit.",
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Context != items[j].Context {
			return items[i].Context < items[j].Context
		}
		if items[i].Kind != items[j].Kind {
			return items[i].Kind < items[j].Kind
		}
		return items[i].Label < items[j].Label
	})
	return items
}

func ideSnippets() []IDESnippet {
	snippets := []IDESnippet{
		{
			Prefix:      "app-api",
			Context:     "top-level",
			Description: "Minimal generated Node API-only app target.",
			Body: []string{
				"app ${1:AppName}",
				"target api {",
				"  backend node",
				"  database sqlite",
				"}",
			},
		},
		{
			Prefix:      "app-web",
			Context:     "top-level",
			Description: "Minimal generated React/Node web app target.",
			Body: []string{
				"app ${1:AppName}",
				"target web {",
				"  frontend react",
				"  backend node",
				"  database sqlite",
				"}",
			},
		},
		{
			Prefix:      "auth-email",
			Context:     "top-level",
			Description: "Email/password cookie auth with roles.",
			Body: []string{
				"auth {",
				"  strategy emailPassword",
				"  session cookie",
				"  user {",
				"    email email required unique",
				"    password text required",
				"  }",
				"}",
			},
		},
		{
			Prefix:      "entity",
			Context:     "top-level",
			Description: "Stored entity with one required field.",
			Body: []string{
				"entity ${1:EntityName} {",
				"  ${2:name} text required",
				"}",
			},
		},
		{
			Prefix:      "computed",
			Context:     "entity",
			Description: "Computed display field that is not a database column or form input.",
			Body: []string{
				"computed ${1:displayValue} money = ${2:stock} * ${3:price} label \"${4:Display Value}\"",
			},
		},
		{
			Prefix:      "media-image",
			Context:     "entity",
			Description: "Local image field stored as a generated data URL string.",
			Body: []string{
				"${1:photo} image optional accept \"image/*\" label \"${2:Photo}\"",
			},
		},
		{
			Prefix:      "relation-load",
			Context:     "entity",
			Description: "Relation field with explicit generated API response include scopes.",
			Body: []string{
				"${1:customer} ${2:Customer} required load ${3:detail} ${4:query} label \"${5:Customer}\"",
			},
		},
		{
			Prefix:      "query",
			Context:     "top-level",
			Description: "Stored-field filtered, sorted, limited query.",
			Body: []string{
				"query ${1:QueryName} {",
				"  source ${2:EntityName}",
				"  where ${3:status} == \"${4:active}\"",
				"  sort ${5:createdAt} desc",
				"  limit ${6:20}",
				"}",
			},
		},
		{
			Prefix:      "job-query",
			Context:     "top-level",
			Description: "Generated background worker job that runs a declared query on a schedule.",
			Body: []string{
				"job ${1:JobName} {",
				"  schedule every ${2:15} minutes",
				"  run query ${3:QueryName}",
				"}",
			},
		},
		{
			Prefix:      "page",
			Context:     "top-level",
			Description: "Generated page bound to an entity.",
			Body: []string{
				"page ${1:PageName} {",
				"  source ${2:EntityName}",
				"",
				"  view {",
				"    order form, table, detail",
				"  }",
				"}",
			},
		},
		{
			Prefix:      "view-component-section",
			Context:     "view",
			Description: "Reusable generated component panel bound to selected, first, or each source record data.",
			Body: []string{
				"section ${1:StockSummary} component ${2:StockBadge} bind ${3:selected} span ${4:1} title \"${5:Stock Summary}\"",
			},
		},
		{
			Prefix:      "action",
			Context:     "top-level",
			Description: "Deterministic custom action with one input and one stored-field assignment.",
			Body: []string{
				"action ${1:ActionName} {",
				"  source ${2:EntityName}",
				"  input ${3:quantity} integer required",
				"  set ${4:stock} = ${4:stock} + ${3:quantity}",
				"  success \"${5:Updated}\"",
				"}",
			},
		},
		{
			Prefix:      "transaction",
			Context:     "top-level",
			Description: "Atomic generated runtime boundary for a declared action or API update handler.",
			Body: []string{
				"transaction ${1:TransactionName} {",
				"  action ${2:ActionName}",
				"  api ${3:APIName}",
				"}",
			},
		},
		{
			Prefix:      "service",
			Context:     "top-level",
			Description: "Generated service module boundary for explicit APIs.",
			Body: []string{
				"service ${1:ServiceName} {",
				"  api ${2:APIName}",
				"}",
			},
		},
		{
			Prefix:      "api-update",
			Context:     "top-level",
			Description: "Explicit API update handler with typed body input.",
			Body: []string{
				"api ${1:UpdateThing} {",
				"  method POST",
				"  path \"/api/${2:things}/{${3:id}}\"",
				"  param ${3:id} text required",
				"  body ${4:name} text required",
				"  update ${5:EntityName} where id == param.${3:id} set ${4:name} = body.${4:name}",
				"  respond updated",
				"}",
			},
		},
		{
			Prefix:      "seed",
			Context:     "top-level",
			Description: "Stable idempotent seed rows for generated setup.",
			Body: []string{
				"seed ${1:DemoRows} {",
				"  source ${2:EntityName}",
				"  row ${3:sample} {",
				"    ${4:name} \"${5:Sample}\"",
				"  }",
				"}",
			},
		},
		{
			Prefix:      "test",
			Context:     "top-level",
			Description: "Generated browser smoke and E2E expectations.",
			Body: []string{
				"test ${1:PageSmoke} {",
				"  page ${2:PageName}",
				"  expect text \"${3:Visible text}\"",
				"  expect page ${2:PageName}",
				"}",
			},
		},
		{
			Prefix:      "deploy-docker",
			Context:     "top-level",
			Description: "Docker deployment intent with local preview, rollback metadata, cloud adapter plan/preflight metadata, and explicit provider CLI apply runner.",
			Body: []string{
				"deploy {",
				"  target docker",
				"  port env PORT default 3001",
				"  env DATABASE_URL required",
				"  preview local",
				"  rollback keep 3",
				"  cloud fly app env FLY_APP_NAME region env FLY_REGION",
				"}",
			},
		},
		{
			Prefix:      "ops",
			Context:     "top-level",
			Description: "Generated health, readiness, metrics, request logging, and observability signals.",
			Body: []string{
				"ops {",
				"  health path \"/healthz\"",
				"  readiness path \"/readyz\"",
				"  metrics path \"/metrics\"",
				"  logging requests",
				"  observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT",
				"}",
			},
		},
	}
	sort.SliceStable(snippets, func(i, j int) bool {
		if snippets[i].Context != snippets[j].Context {
			return snippets[i].Context < snippets[j].Context
		}
		return snippets[i].Prefix < snippets[j].Prefix
	})
	return snippets
}

func ideDiagnosticCodes() []IDEDiagnosticCode {
	codeKeywords := map[string][]string{}
	for _, doc := range AllDocs() {
		for _, code := range doc.Errors {
			if code == "" {
				continue
			}
			codeKeywords[code] = append(codeKeywords[code], doc.Keyword)
		}
	}
	codes := make([]string, 0, len(codeKeywords))
	for code := range codeKeywords {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	result := make([]IDEDiagnosticCode, 0, len(codes))
	for _, code := range codes {
		keywords := append([]string(nil), codeKeywords[code]...)
		sort.Strings(keywords)
		result = append(result, IDEDiagnosticCode{
			Code:          code,
			Severity:      ideDiagnosticSeverity(code),
			Documentation: "Referenced by docs keywords: " + strings.Join(keywords, ", "),
		})
	}
	return result
}

func sortedMapKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
