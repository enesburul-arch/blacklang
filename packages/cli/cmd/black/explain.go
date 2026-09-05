package main

import (
	"fmt"
	"strings"
)

var relatedDocsByKeyword = map[string][]string{
	"access":    {"page", "auth", "role"},
	"actions":   {"page", "api", "audit", "workflow"},
	"api":       {"openapi", "auth", "security", "cors"},
	"app":       {"syntax", "entity", "page"},
	"audit":     {"auth", "role", "actions"},
	"auth":      {"role", "access", "csrf", "security", "cors"},
	"blackir":   {"parse", "validate", "inspect", "docs"},
	"component": {"entity", "table", "form", "state"},
	"cors":      {"security", "api", "auth"},
	"csrf":      {"auth", "security"},
	"database":  {"security", "syntax"},
	"deploy":    {"database", "security", "package"},
	"docs":      {"syntax", "explain", "lint"},
	"entity":    {"page", "table", "form", "component"},
	"explain":   {"docs", "lint", "syntax"},
	"filter":    {"table", "search", "paginate"},
	"format":    {"lint", "validate"},
	"form":      {"entity", "page", "actions"},
	"layout":    {"page", "access"},
	"lint":      {"format", "validate", "security", "docs"},
	"openapi":   {"api", "page"},
	"page":      {"entity", "table", "form", "actions", "layout", "access"},
	"paginate":  {"table", "filter", "search"},
	"role":      {"auth", "access", "audit"},
	"search":    {"table", "filter"},
	"security":  {"database", "lint", "csrf", "cors", "auth", "deploy"},
	"state":     {"page", "component", "form"},
	"syntax":    {"app", "entity", "page", "docs"},
	"table":     {"page", "entity", "search", "filter", "paginate"},
	"validate":  {"lint", "format"},
	"version":   {"docs", "lint"},
	"workflow":  {"entity", "page", "actions", "audit"},
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
