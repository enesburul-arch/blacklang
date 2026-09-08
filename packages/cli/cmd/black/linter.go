package main

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

	source, readDiagnostics := ReadBlackSource(file)
	if len(readDiagnostics) > 0 {
		result.Success = false
		result.Errors = readDiagnostics
		result.Checks = append(result.Checks,
			LintCheck{Name: "format", Success: false, Findings: len(readDiagnostics)},
			LintCheck{Name: "parse", Success: false, Findings: len(readDiagnostics)},
			LintCheck{Name: "validate", Success: false, Findings: len(readDiagnostics)},
			LintCheck{Name: "security", Success: false, Findings: len(readDiagnostics)},
		)
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
	result.addLintCheck("format", formatFindings)

	program, parseDiagnostics := Parse(file, source)
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

	securityFindings := SecurityScanText(file, source)
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
