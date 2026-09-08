package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIDESupportIncludesCompletionSnippetsAndDiagnostics(t *testing.T) {
	result := IDESupport()
	if !result.Success {
		t.Fatalf("expected IDE support success, got %#v", result.Errors)
	}
	if result.Command != "ide" || result.Language.ID != "blacklang" {
		t.Fatalf("expected IDE identity, got %#v", result)
	}
	if !containsIDEExtension(result.Language.Extensions, ".black") || !containsIDEExtension(result.Language.Extensions, ".blackthm") {
		t.Fatalf("expected BlackLang source extensions, got %#v", result.Language.Extensions)
	}
	if result.Capabilities.Diagnostics != "black ide diagnostics <file> --json" || !result.Capabilities.CompletionItems || !result.Capabilities.Snippets {
		t.Fatalf("expected IDE capabilities, got %#v", result.Capabilities)
	}
	if !containsIDECommand(result.Commands, "diagnostics") {
		t.Fatalf("expected diagnostics command, got %#v", result.Commands)
	}
	if !containsIDECompletion(result.CompletionItems, "top-level", "entity") {
		t.Fatalf("expected entity top-level completion, got %#v", result.CompletionItems)
	}
	if !containsIDECompletion(result.CompletionItems, "top-level", "job") || !containsIDECompletion(result.CompletionItems, "top-level", "service") || !containsIDECompletion(result.CompletionItems, "block", "schedule") || !containsIDECompletion(result.CompletionItems, "job-schedule", "minutes") {
		t.Fatalf("expected job/service top-level, schedule block, and schedule unit completions, got %#v", result.CompletionItems)
	}
	if !containsIDECompletion(result.CompletionItems, "target-name", "api") || !containsIDECompletion(result.CompletionItems, "target-name", "web") {
		t.Fatalf("expected target name completions, got %#v", result.CompletionItems)
	}
	if !containsIDECompletion(result.CompletionItems, "target-database", "mysql") {
		t.Fatalf("expected mysql target database completion, got %#v", result.CompletionItems)
	}
	if !containsIDECompletion(result.CompletionItems, "field-type", "money") {
		t.Fatalf("expected money field type completion, got %#v", result.CompletionItems)
	}
	if !containsIDECompletion(result.CompletionItems, "field-type", "image") || !containsIDECompletion(result.CompletionItems, "field-modifier", "accept") {
		t.Fatalf("expected media field type and accept modifier completions, got %#v", result.CompletionItems)
	}
	if !containsIDECompletion(result.CompletionItems, "field-modifier", "load") || !containsIDECompletion(result.CompletionItems, "relation-load", "query") || !containsIDECompletion(result.CompletionItems, "relation-load", "none") {
		t.Fatalf("expected relation load modifier and scope completions, got %#v", result.CompletionItems)
	}
	if !containsIDECompletion(result.CompletionItems, "block", "bind") || !containsIDECompletion(result.CompletionItems, "view-component-bind", "selected") || !containsIDECompletion(result.CompletionItems, "view-component-bind", "first") {
		t.Fatalf("expected view component bind completions, got %#v", result.CompletionItems)
	}
	if !containsIDECompletion(result.CompletionItems, "deploy-preview", "local") {
		t.Fatalf("expected deploy preview completion, got %#v", result.CompletionItems)
	}
	if !containsIDECompletion(result.CompletionItems, "deploy-cloud", "fly") || !containsIDECompletion(result.CompletionItems, "ops-observe", "webhook") || !containsIDECompletion(result.CompletionItems, "ops-observe", "otlp") {
		t.Fatalf("expected cloud and observe completions, got %#v", result.CompletionItems)
	}
	if !containsIDESnippet(result.Snippets, "query") || !containsIDESnippet(result.Snippets, "job-query") || !containsIDESnippet(result.Snippets, "service") || !containsIDESnippet(result.Snippets, "deploy-docker") || !containsIDESnippet(result.Snippets, "media-image") || !containsIDESnippet(result.Snippets, "relation-load") || !containsIDESnippet(result.Snippets, "view-component-section") || !containsIDESnippet(result.Snippets, "app-api") {
		t.Fatalf("expected query, job, service, deploy, media, relation load, view component, and API target snippets, got %#v", result.Snippets)
	}
	if !containsIDEDiagnosticCode(result.DiagnosticCodes, "MISSING_APP") || !containsIDEDiagnosticCode(result.DiagnosticCodes, "UNKNOWN_DOC_KEYWORD") || !containsIDEDiagnosticCode(result.DiagnosticCodes, "MISSING_ACCEPT_VALUE") || !containsIDEDiagnosticCode(result.DiagnosticCodes, "MISSING_RELATION_LOAD_SCOPE") || !containsIDEDiagnosticCode(result.DiagnosticCodes, "UNKNOWN_VIEW_COMPONENT") || !containsIDEDiagnosticCode(result.DiagnosticCodes, "UNSUPPORTED_VIEW_COMPONENT_INPUT") || !containsIDEDiagnosticCode(result.DiagnosticCodes, "UNKNOWN_JOB_QUERY") || !containsIDEDiagnosticCode(result.DiagnosticCodes, "UNSUPPORTED_API_TARGET_TEST") || !containsIDEDiagnosticCode(result.DiagnosticCodes, "UNSUPPORTED_TARGET_DATABASE_MIGRATION") {
		t.Fatalf("expected compiler diagnostic catalog, got %#v", result.DiagnosticCodes)
	}
}

func TestFormatIDESupportIR(t *testing.T) {
	ir := FormatIDESupportIR(IDESupport())
	for _, expected := range []string{
		"blackir 0.1",
		"ide ok",
		"language blacklang version 0.1",
		"diagnosticsCommand",
		"completion top-level entity keyword",
		"snippet top-level app-api",
		"snippet top-level job-query",
		"snippet top-level deploy-docker",
		"diagnostic MISSING_APP error",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected IDE IR to contain %q, got:\n%s", expected, ir)
		}
	}
}

func TestIDEDiagnosticsReportsFindingsWithoutCommandFailure(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "app.black")
	if err := os.WriteFile(file, []byte(`app Warehouse
target web {
  frontend react
  backend node
  database sqlite
}
entity Product {
  sku text required unique
}
page Products {
  source Missing
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	result := IDEDiagnosticsFile(file)
	if !result.Success {
		t.Fatalf("expected IDE diagnostics command success for readable invalid source, got %#v", result.Errors)
	}
	if result.Valid {
		t.Fatalf("expected invalid source to be valid=false, got %#v", result)
	}
	if result.Project.App != "Warehouse" || result.Project.Entities != 1 || result.Project.Pages != 1 {
		t.Fatalf("expected project summary from parseable source, got %#v", result.Project)
	}
	if result.Summary.Total == 0 || result.Summary.Errors == 0 {
		t.Fatalf("expected validation diagnostics, got %#v", result.Summary)
	}
	if !containsIDEDiagnostic(result.Diagnostics, "validate", "UNKNOWN_SOURCE_ENTITY") {
		t.Fatalf("expected UNKNOWN_SOURCE_ENTITY validation diagnostic, got %#v", result.Diagnostics)
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Range.Start.Line < 0 || diagnostic.Range.Start.Character < 0 {
			t.Fatalf("expected zero-based non-negative ranges, got %#v", diagnostic)
		}
	}
}

func TestIDEDiagnosticsReadErrorIsCommandError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "missing.black")
	result := IDEDiagnosticsFile(file)
	if result.Success || result.Valid {
		t.Fatalf("expected missing file to fail, got %#v", result)
	}
	if len(result.Errors) != 1 || result.Errors[0].Code != "FILE_READ_ERROR" {
		t.Fatalf("expected FILE_READ_ERROR, got %#v", result.Errors)
	}
	if !containsIDEDiagnostic(result.Diagnostics, "read", "FILE_READ_ERROR") {
		t.Fatalf("expected read diagnostic, got %#v", result.Diagnostics)
	}
}

func TestFormatIDEDiagnosticsIR(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "app.black")
	if err := os.WriteFile(file, []byte(`app Warehouse
target web {
  frontend react
  backend node
  database sqlite
}
page Products {
  source Missing
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	ir := FormatIDEDiagnosticsIR(IDEDiagnosticsFile(file))
	for _, expected := range []string{
		"blackir 0.1",
		"ide diagnostics ok",
		"valid false",
		"summary diagnostics",
		"UNKNOWN_SOURCE_ENTITY",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected IDE diagnostics IR to contain %q, got:\n%s", expected, ir)
		}
	}
}

func containsIDEExtension(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func containsIDECommand(commands []IDECommand, name string) bool {
	for _, command := range commands {
		if command.Name == name {
			return true
		}
	}
	return false
}

func containsIDECompletion(items []IDECompletionItem, context string, label string) bool {
	for _, item := range items {
		if item.Context == context && item.Label == label {
			return true
		}
	}
	return false
}

func containsIDESnippet(snippets []IDESnippet, prefix string) bool {
	for _, snippet := range snippets {
		if snippet.Prefix == prefix {
			return true
		}
	}
	return false
}

func containsIDEDiagnosticCode(codes []IDEDiagnosticCode, code string) bool {
	for _, diagnostic := range codes {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func containsIDEDiagnostic(diagnostics []IDEDiagnostic, source string, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Source == source && diagnostic.Code == code {
			return true
		}
	}
	return false
}
