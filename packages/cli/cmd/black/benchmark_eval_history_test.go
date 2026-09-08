package main

import (
	"strings"
	"testing"
)

func TestBenchmarkEvalHistoryReadsPublishedLocalManifest(t *testing.T) {
	result := BenchmarkEvalHistory(nil)
	if !result.Success {
		t.Fatalf("expected eval history success, got %#v", result.Errors)
	}
	if result.Command != "benchmark eval-history" || result.History.ID != "BLACKLANG-WEB-AI-EVAL-HISTORY" {
		t.Fatalf("expected eval history identity, got %#v", result)
	}
	if result.History.Status != "published-local" || result.History.PublishedAt == "" {
		t.Fatalf("expected published-local history metadata, got %#v", result.History)
	}
	if !strings.Contains(result.History.Policy, "must not be presented as billed model benchmark scores") {
		t.Fatalf("expected evidence-backed scoring policy, got %q", result.History.Policy)
	}
	if !strings.Contains(strings.Join(result.History.Commands, " "), "black benchmark eval-history --json") {
		t.Fatalf("expected eval-history command in manifest, got %#v", result.History.Commands)
	}
	if result.Summary.RunCount != 1 || result.Summary.ModelCount != 1 || result.Summary.FailedRuns != 0 {
		t.Fatalf("expected one clean local validation history run, got %#v", result.Summary)
	}
	if result.Summary.TotalRuns == 0 || result.Summary.PassedRuns != result.Summary.TotalRuns || result.Summary.AverageScore != 100 {
		t.Fatalf("expected passing local validation summary, got %#v", result.Summary)
	}
	if len(result.Runs) != 1 {
		t.Fatalf("expected one run, got %#v", result.Runs)
	}
	run := result.Runs[0]
	if run.SuiteID != "BLACKLANG-WEB-AI-EVAL-v0.2" || run.CoveragePercent != 99 {
		t.Fatalf("expected eval suite and coverage metadata, got %#v", run)
	}
	if !containsString(run.Evidence, "WEB-COMPLETION-WORKLOG.md") || !containsString(run.Evidence, "benchmarks/warehouse-v0.1.md") {
		t.Fatalf("expected evidence paths in eval history, got %#v", run.Evidence)
	}
	if len(run.Models) != 1 || run.Models[0].Source != "compiler-generated-checks" || run.Models[0].Provider != "local" {
		t.Fatalf("expected local validation source model entry, got %#v", run.Models)
	}
	if !strings.Contains(strings.Join(run.Notes, " "), "not an external paid-model score") {
		t.Fatalf("expected non-model-score note, got %#v", run.Notes)
	}
}

func TestFormatAIEvalHistoryIR(t *testing.T) {
	result := AIEvalHistoryResult{
		Success: true,
		History: AIEvalHistory{
			ID:          "BLACKLANG-WEB-AI-EVAL-HISTORY",
			Version:     "0.2",
			Status:      "published-local",
			PublishedAt: "2026-09-07",
		},
		Summary: AIEvalHistorySummary{
			RunCount:        1,
			ModelCount:      1,
			TotalRuns:       24,
			PassedRuns:      24,
			FailedRuns:      0,
			AverageScore:    100,
			CoveragePercent: 99,
			LatestRun:       "2026-09-07",
		},
		Runs: []AIEvalHistoryRun{
			{
				ID:              "warehouse-web-completion-2026-09-07",
				SuiteID:         "BLACKLANG-WEB-AI-EVAL-v0.2",
				Source:          "examples/warehouse/app.black",
				Status:          "published-local",
				CoveragePercent: 99,
				TotalRuns:       24,
				PassedRuns:      24,
				FailedRuns:      0,
				AverageScore:    100,
				Evidence:        []string{"WEB-COMPLETION-WORKLOG.md"},
				Models: []AIEvalHistoryModelResult{
					{
						Name:         "blacklang-local-validation",
						Provider:     "local",
						Source:       "compiler-generated-checks",
						Runs:         24,
						Passed:       24,
						Failed:       0,
						AverageScore: 100,
						MedianTokens: 0,
					},
				},
			},
		},
	}

	ir := FormatAIEvalHistoryIR(result)
	for _, expected := range []string{
		"blackir 0.1",
		"benchmark eval-history ok",
		"history BLACKLANG-WEB-AI-EVAL-HISTORY version 0.2 status published-local published_at 2026-09-07",
		"summary runs 1 models 1 total_runs 24 passed 24 failed 0 average_score 100 coverage 99 latest 2026-09-07",
		"run warehouse-web-completion-2026-09-07 suite BLACKLANG-WEB-AI-EVAL-v0.2 status published-local source examples/warehouse/app.black coverage 99 score 100 total_runs 24 passed 24 failed 0",
		"model blacklang-local-validation provider local source compiler-generated-checks runs 24 passed 24 failed 0 score 100 median_tokens 0",
		"evidence WEB-COMPLETION-WORKLOG.md",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected AI eval history IR to contain %q, got:\n%s", expected, ir)
		}
	}
}
