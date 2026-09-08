package main

import (
	"fmt"
	"strings"
)

func AccessibilityAuditFile(file string) AccessibilityAuditResult {
	result := AccessibilityAuditResult{
		Success:  true,
		Command:  "audit accessibility",
		Version:  version,
		File:     file,
		Summary:  Summary{},
		Findings: []Diagnostic{},
		Errors:   []Diagnostic{},
	}

	source, readDiagnostics := ReadBlackSource(file)
	if len(readDiagnostics) > 0 {
		result.Success = false
		result.Errors = readDiagnostics
		return result
	}

	program, parseDiagnostics := Parse(file, source)
	if len(parseDiagnostics) > 0 {
		result.Success = false
		result.Errors = parseDiagnostics
		return result
	}

	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) > 0 {
		result.Success = false
		result.Errors = validateDiagnostics
		return result
	}

	result.Summary = Summary{
		App:      program.App.Name,
		Entities: len(program.Entities),
		Pages:    len(program.Pages),
	}
	result.Findings = AccessibilityDiagnostics(program)
	result.Success = len(result.Findings) == 0 && len(result.Errors) == 0
	return result
}

func AccessibilityDiagnostics(program Program) []Diagnostic {
	findings := []Diagnostic{}
	for _, page := range program.Pages {
		if page.View == nil {
			continue
		}
		for _, section := range page.View.Sections {
			if (section.Display == "modal" || section.Display == "drawer") && strings.TrimSpace(section.Title) == "" {
				findings = append(findings, diagnosticAt(section.Position, "ACCESSIBILITY_MISSING_OVERLAY_TITLE", fmt.Sprintf("Page %s view section %s uses display %s without an explicit title.", page.Name, section.Name, section.Display), "Add title \"...\" to the overlay section so generated dialogs have stable accessible names."))
			}
		}
		for _, group := range page.View.Groups {
			if len(group.Sections) > 1 && strings.TrimSpace(group.Title) == "" {
				findings = append(findings, diagnosticAt(group.Position, "ACCESSIBILITY_MISSING_GROUP_TITLE", fmt.Sprintf("Page %s view group %s wraps multiple sections without an explicit title.", page.Name, group.Name), "Add title \"...\" to the group so generated nested section landmarks are easy to understand."))
			}
		}
	}
	if findings == nil {
		return []Diagnostic{}
	}
	return findings
}

func FormatAccessibilityAuditIR(result AccessibilityAuditResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("audit accessibility ok\n")
	} else {
		builder.WriteString("audit accessibility findings\n")
	}
	builder.WriteString(fmt.Sprintf("file %s\n", result.File))
	builder.WriteString(fmt.Sprintf("findings %d\n", len(result.Findings)))
	for _, finding := range result.Findings {
		builder.WriteString(fmt.Sprintf("  %s %s:%d:%d\n", finding.Code, finding.File, finding.Line, finding.Column))
	}
	builder.WriteString(fmt.Sprintf("errors %d\n", len(result.Errors)))
	for _, diagnostic := range result.Errors {
		builder.WriteString(fmt.Sprintf("  %s %s:%d:%d\n", diagnostic.Code, diagnostic.File, diagnostic.Line, diagnostic.Column))
	}
	return builder.String()
}

func diagnosticAt(position Position, code string, message string, suggestion string) Diagnostic {
	return Diagnostic{
		File:       position.File,
		Line:       position.Line,
		Column:     position.Column,
		Code:       code,
		Message:    message,
		Suggestion: suggestion,
	}
}
