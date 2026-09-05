package main

import (
	"strings"
	"testing"
)

func TestExplainKeyword(t *testing.T) {
	result := ExplainKeyword("entity")
	if !result.Success {
		t.Fatalf("expected explain success, got %#v", result)
	}
	if result.Keyword != "entity" || result.Purpose == "" || result.Syntax == "" || result.Example == "" {
		t.Fatalf("expected populated entity explanation, got %#v", result)
	}
	if len(result.AgentSteps) == 0 || !strings.Contains(result.AgentSteps[len(result.AgentSteps)-1], "black lint --json") {
		t.Fatalf("expected agent steps to include validation flow, got %#v", result.AgentSteps)
	}
	if !containsExplainString(result.Related, "page") || !containsExplainString(result.Related, "table") {
		t.Fatalf("expected related keywords for entity, got %#v", result.Related)
	}
	if !containsExplainString(result.ErrorCodes, "DUPLICATE_ENTITY") {
		t.Fatalf("expected entity error codes, got %#v", result.ErrorCodes)
	}
}

func TestExplainKeywordDefaultsToSyntax(t *testing.T) {
	result := ExplainKeyword("")
	if !result.Success {
		t.Fatalf("expected default syntax explanation success, got %#v", result)
	}
	if result.Keyword != "syntax" {
		t.Fatalf("expected syntax default, got %#v", result)
	}
}

func TestExplainKeywordReportsUnknownKeyword(t *testing.T) {
	result := ExplainKeyword("missing")
	if result.Success {
		t.Fatalf("expected unknown keyword failure")
	}
	if len(result.Errors) != 1 || result.Errors[0].Code != "UNKNOWN_EXPLAIN_KEYWORD" {
		t.Fatalf("expected UNKNOWN_EXPLAIN_KEYWORD, got %#v", result)
	}
}

func TestFormatExplainIR(t *testing.T) {
	result := ExplainKeyword("lint")
	ir := FormatExplainIR(result)
	for _, expected := range []string{
		"explain ok",
		"keyword lint",
		"agentSteps",
		"related format security",
		"FORMAT_REQUIRED",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected IR to contain %q, got:\n%s", expected, ir)
		}
	}
}

func containsExplainString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
