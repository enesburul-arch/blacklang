package main

import "os"

func LintFile(file string) LintResult {
	result := LintResult{
		Success:  true,
		Command:  "lint",
		Version:  version,
		File:     file,
		Summary:  Summary{},
		Checks:   []LintCheck{},
		Findings: []Diagnostic{},
		Errors:   []Diagnostic{},
	}

	source, err := os.ReadFile(file)
	if err != nil {
		result.Success = false
		result.Errors = []Diagnostic{{
			File:       file,
			Code:       "FILE_READ_ERROR",
			Message:    err.Error(),
			Suggestion: "Pass a readable .black file path or set source in blacklang.toml.",
		}}
		result.Checks = append(result.Checks,
			LintCheck{Name: "format", Success: false, Findings: 1},
			LintCheck{Name: "parse", Success: false, Findings: 1},
			LintCheck{Name: "validate", Success: false, Findings: 1},
			LintCheck{Name: "security", Success: false, Findings: 1},
		)
		return result
	}

	formatted, formatDiagnostics := FormatBlackSource(file, string(source))
	formatFindings := append([]Diagnostic{}, formatDiagnostics...)
	if len(formatDiagnostics) == 0 && formatted != string(source) {
		formatFindings = append(formatFindings, Diagnostic{
			File:       file,
			Code:       "FORMAT_REQUIRED",
			Message:    "BlackLang source is not formatted.",
			Suggestion: "Run `black format " + file + "`.",
		})
	}
	result.addLintCheck("format", formatFindings)

	program, parseDiagnostics := Parse(file, string(source))
	result.addLintCheck("parse", parseDiagnostics)

	validateDiagnostics := []Diagnostic{}
	if len(parseDiagnostics) == 0 {
		validateDiagnostics = Validate(program)
		result.Summary = Summary{
			App:      program.App.Name,
			Entities: len(program.Entities),
			Pages:    len(program.Pages),
		}
	}
	result.addLintCheck("validate", validateDiagnostics)

	securityFindings := SecurityScanText(file, string(source))
	result.addLintCheck("security", securityFindings)

	result.Success = len(result.Findings) == 0 && len(result.Errors) == 0
	return result
}

func (result *LintResult) addLintCheck(name string, findings []Diagnostic) {
	if findings == nil {
		findings = []Diagnostic{}
	}
	result.Checks = append(result.Checks, LintCheck{
		Name:     name,
		Success:  len(findings) == 0,
		Findings: len(findings),
	})
	result.Findings = append(result.Findings, findings...)
}
