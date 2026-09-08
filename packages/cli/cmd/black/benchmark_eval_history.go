package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultAIEvalHistoryRelativePath = "benchmarks/eval-history.blackdir"

func BenchmarkEvalHistory(args []string) AIEvalHistoryResult {
	path := optionValue(args, "--history")
	if path == "" {
		path = discoverAIEvalHistoryManifest()
	}

	result := AIEvalHistoryResult{
		Success: false,
		Command: "benchmark eval-history",
		Version: version,
		Errors:  []Diagnostic{},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		result.Errors = []Diagnostic{{
			File:       path,
			Line:       1,
			Column:     1,
			Code:       "EVAL_HISTORY_READ_ERROR",
			Message:    "Could not read AI eval history manifest.",
			Suggestion: "Create benchmarks/eval-history.blackdir or pass --history <file>.",
		}}
		return result
	}

	history, runs, diagnostics := parseAIEvalHistoryManifest(path, string(data))
	result.History = history
	result.Runs = runs
	result.Summary = summarizeAIEvalHistory(runs)
	if len(diagnostics) > 0 {
		result.Errors = diagnostics
		return result
	}

	result.Success = true
	return result
}

func discoverAIEvalHistoryManifest() string {
	wd, err := os.Getwd()
	if err != nil {
		return defaultAIEvalHistoryRelativePath
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, defaultAIEvalHistoryRelativePath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return defaultAIEvalHistoryRelativePath
}

func parseAIEvalHistoryManifest(path string, content string) (AIEvalHistory, []AIEvalHistoryRun, []Diagnostic) {
	history := AIEvalHistory{Source: filepath.ToSlash(path), Commands: []string{}}
	runs := []AIEvalHistoryRun{}
	diagnostics := []Diagnostic{}
	currentRun := -1

	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	for index, rawLine := range lines {
		lineNumber := index + 1
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}

		switch tokens[0] {
		case "evalHistory":
			if len(tokens) < 2 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Missing eval history id."))
				continue
			}
			history.ID = tokens[1]
		case "version":
			if len(tokens) < 2 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Missing eval history version."))
				continue
			}
			history.Version = tokens[1]
		case "status":
			if len(tokens) < 2 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Missing eval history status."))
				continue
			}
			history.Status = tokens[1]
		case "source":
			if len(tokens) < 2 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Missing eval history source."))
				continue
			}
			history.Source = tokens[1]
		case "publishedAt":
			if len(tokens) < 2 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Missing eval history publishedAt date."))
				continue
			}
			history.PublishedAt = tokens[1]
		case "policy":
			history.Policy = unquoteEvalHistoryValue(evalHistoryTrailingValue(line, "policy"))
		case "command":
			history.Commands = append(history.Commands, unquoteEvalHistoryValue(evalHistoryTrailingValue(line, "command")))
		case "run":
			if len(tokens) < 2 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Missing eval history run id."))
				continue
			}
			fields, fieldDiagnostics := evalHistoryFields(path, lineNumber, tokens[2:])
			diagnostics = append(diagnostics, fieldDiagnostics...)
			run := AIEvalHistoryRun{
				ID:              tokens[1],
				SuiteID:         fields["suite"],
				Source:          fields["source"],
				Status:          fields["status"],
				StartedAt:       fields["startedAt"],
				CompletedAt:     fields["completedAt"],
				CoveragePercent: evalHistoryIntField(path, lineNumber, fields, "coverage", &diagnostics),
				CaseCount:       evalHistoryIntField(path, lineNumber, fields, "cases", &diagnostics),
				Repeat:          evalHistoryIntField(path, lineNumber, fields, "repeat", &diagnostics),
				TotalRuns:       evalHistoryIntField(path, lineNumber, fields, "totalRuns", &diagnostics),
				PassedRuns:      evalHistoryIntField(path, lineNumber, fields, "passedRuns", &diagnostics),
				FailedRuns:      evalHistoryIntField(path, lineNumber, fields, "failedRuns", &diagnostics),
				AverageScore:    evalHistoryIntField(path, lineNumber, fields, "averageScore", &diagnostics),
				Evidence:        []string{},
				Notes:           []string{},
				Models:          []AIEvalHistoryModelResult{},
			}
			runs = append(runs, run)
			currentRun = len(runs) - 1
		case "model":
			if currentRun < 0 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Model entry must follow a run entry."))
				continue
			}
			if len(tokens) < 2 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Missing eval history model name."))
				continue
			}
			fields, fieldDiagnostics := evalHistoryFields(path, lineNumber, tokens[2:])
			diagnostics = append(diagnostics, fieldDiagnostics...)
			model := AIEvalHistoryModelResult{
				Name:         tokens[1],
				Provider:     fields["provider"],
				Source:       fields["source"],
				Runs:         evalHistoryIntField(path, lineNumber, fields, "runs", &diagnostics),
				Passed:       evalHistoryIntField(path, lineNumber, fields, "passed", &diagnostics),
				Failed:       evalHistoryIntField(path, lineNumber, fields, "failed", &diagnostics),
				AverageScore: evalHistoryIntField(path, lineNumber, fields, "averageScore", &diagnostics),
				MedianTokens: evalHistoryIntField(path, lineNumber, fields, "medianTokens", &diagnostics),
				Notes:        []string{},
			}
			runs[currentRun].Models = append(runs[currentRun].Models, model)
		case "evidence":
			if currentRun < 0 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Evidence entry must follow a run entry."))
				continue
			}
			value := evalHistoryTrailingValue(line, "evidence")
			if value == "" {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Missing eval history evidence path."))
				continue
			}
			runs[currentRun].Evidence = append(runs[currentRun].Evidence, unquoteEvalHistoryValue(value))
		case "note":
			if currentRun < 0 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Note entry must follow a run entry."))
				continue
			}
			runs[currentRun].Notes = append(runs[currentRun].Notes, unquoteEvalHistoryValue(evalHistoryTrailingValue(line, "note")))
		case "modelNote":
			if currentRun < 0 || len(runs[currentRun].Models) == 0 {
				diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Model note entry must follow a model entry."))
				continue
			}
			modelIndex := len(runs[currentRun].Models) - 1
			runs[currentRun].Models[modelIndex].Notes = append(runs[currentRun].Models[modelIndex].Notes, unquoteEvalHistoryValue(evalHistoryTrailingValue(line, "modelNote")))
		default:
			diagnostics = append(diagnostics, evalHistoryDiagnostic(path, lineNumber, "INVALID_EVAL_HISTORY_MANIFEST", "Unknown eval history directive "+tokens[0]+"."))
		}
	}

	diagnostics = append(diagnostics, validateAIEvalHistoryManifest(path, history, runs)...)
	return history, runs, diagnostics
}

func evalHistoryFields(path string, line int, tokens []string) (map[string]string, []Diagnostic) {
	fields := map[string]string{}
	diagnostics := []Diagnostic{}
	if len(tokens)%2 != 0 {
		diagnostics = append(diagnostics, evalHistoryDiagnostic(path, line, "INVALID_EVAL_HISTORY_MANIFEST", "Expected key/value pairs in eval history directive."))
	}
	for index := 0; index+1 < len(tokens); index += 2 {
		fields[tokens[index]] = tokens[index+1]
	}
	return fields, diagnostics
}

func evalHistoryIntField(path string, line int, fields map[string]string, name string, diagnostics *[]Diagnostic) int {
	value := fields[name]
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		*diagnostics = append(*diagnostics, evalHistoryDiagnostic(path, line, "INVALID_EVAL_HISTORY_MANIFEST", "Expected integer value for "+name+"."))
		return 0
	}
	return parsed
}

func validateAIEvalHistoryManifest(path string, history AIEvalHistory, runs []AIEvalHistoryRun) []Diagnostic {
	diagnostics := []Diagnostic{}
	if history.ID == "" {
		diagnostics = append(diagnostics, evalHistoryDiagnostic(path, 1, "INVALID_EVAL_HISTORY_MANIFEST", "Eval history manifest must declare evalHistory <id>."))
	}
	if history.Version == "" {
		diagnostics = append(diagnostics, evalHistoryDiagnostic(path, 1, "INVALID_EVAL_HISTORY_MANIFEST", "Eval history manifest must declare version."))
	}
	if history.Status == "" {
		diagnostics = append(diagnostics, evalHistoryDiagnostic(path, 1, "INVALID_EVAL_HISTORY_MANIFEST", "Eval history manifest must declare status."))
	}
	if history.Policy == "" {
		diagnostics = append(diagnostics, evalHistoryDiagnostic(path, 1, "INVALID_EVAL_HISTORY_MANIFEST", "Eval history manifest must declare policy."))
	}
	if len(history.Commands) == 0 {
		diagnostics = append(diagnostics, evalHistoryDiagnostic(path, 1, "INVALID_EVAL_HISTORY_MANIFEST", "Eval history manifest must declare at least one command."))
	}
	if len(runs) == 0 {
		diagnostics = append(diagnostics, evalHistoryDiagnostic(path, 1, "INVALID_EVAL_HISTORY_MANIFEST", "Eval history manifest must declare at least one run."))
	}
	for _, run := range runs {
		if run.ID == "" || run.SuiteID == "" || run.Status == "" {
			diagnostics = append(diagnostics, evalHistoryDiagnostic(path, 1, "INVALID_EVAL_HISTORY_MANIFEST", "Every eval history run must declare id, suite, and status."))
		}
		if len(run.Models) == 0 {
			diagnostics = append(diagnostics, evalHistoryDiagnostic(path, 1, "INVALID_EVAL_HISTORY_MANIFEST", "Every eval history run must include at least one model or validation source result."))
		}
		if len(run.Evidence) == 0 {
			diagnostics = append(diagnostics, evalHistoryDiagnostic(path, 1, "INVALID_EVAL_HISTORY_MANIFEST", "Every eval history run must include evidence paths."))
		}
	}
	return diagnostics
}

func summarizeAIEvalHistory(runs []AIEvalHistoryRun) AIEvalHistorySummary {
	summary := AIEvalHistorySummary{RunCount: len(runs)}
	models := map[string]bool{}
	scoreWeight := 0
	for _, run := range runs {
		summary.TotalRuns += run.TotalRuns
		summary.PassedRuns += run.PassedRuns
		summary.FailedRuns += run.FailedRuns
		if run.CoveragePercent > summary.CoveragePercent {
			summary.CoveragePercent = run.CoveragePercent
		}
		if run.CompletedAt >= summary.LatestRun {
			summary.LatestRun = run.CompletedAt
		}
		scoreWeight += run.AverageScore * run.TotalRuns
		for _, model := range run.Models {
			key := model.Provider + "/" + model.Name + "/" + model.Source
			models[key] = true
		}
	}
	summary.ModelCount = len(models)
	if summary.TotalRuns > 0 {
		summary.AverageScore = scoreWeight / summary.TotalRuns
	}
	return summary
}

func evalHistoryTrailingValue(line string, keyword string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, keyword))
}

func unquoteEvalHistoryValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	unquoted, err := strconv.Unquote(value)
	if err == nil {
		return unquoted
	}
	return value
}

func evalHistoryDiagnostic(path string, line int, code string, message string) Diagnostic {
	return Diagnostic{
		File:       path,
		Line:       line,
		Column:     1,
		Code:       code,
		Message:    message,
		Suggestion: "Use the documented benchmarks/eval-history.blackdir key/value format.",
	}
}
