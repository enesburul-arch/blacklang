package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBenchmarkEvalCorpusReportsLongRunningCasesWithoutMutatingOutDir(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "app.black")
	source := `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  stock number default 0
}

query LowStockProducts {
  source Product
  where stock < 10
  sort stock asc
  limit 50
}

action RestockProduct {
  source Product
  input quantity number required min 1
  set stock = stock + quantity
}

page Products {
  source Product
  query LowStockProducts

  table {
    columns sku, stock
  }

  form {
    fields sku, stock
  }

  actions create, edit, RestockProduct
}

test WarehouseBrowserSmoke {
  page Products
  expect text "Warehouse"
  expect action RestockProduct
}
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(root, "generated-out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	sentinelPath := filepath.Join(outDir, "sentinel.txt")
	if err := os.WriteFile(sentinelPath, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	tasks := BenchmarkTasks([]string{sourcePath, "--out", outDir})
	if !tasks.Success {
		t.Fatalf("expected benchmark tasks success, got %#v", tasks.Errors)
	}

	result := BenchmarkEvalCorpus([]string{sourcePath, "--out", outDir})
	if !result.Success {
		t.Fatalf("expected benchmark eval success, got %#v", result.Errors)
	}
	if result.Command != "benchmark eval" || result.Summary.App != "Warehouse" {
		t.Fatalf("expected benchmark eval identity, got %#v", result)
	}
	if result.Suite.ID != "BLACKLANG-WEB-AI-EVAL-v0.2" || result.Suite.Mode != "long-running" || result.Suite.Repeat != 3 {
		t.Fatalf("expected long-running eval suite metadata, got %#v", result.Suite)
	}
	if !strings.Contains(result.Suite.Policy, "read-only metadata") || !strings.Contains(result.Suite.Policy, "does not call an AI model") {
		t.Fatalf("expected read-only eval policy, got %q", result.Suite.Policy)
	}
	if !strings.Contains(strings.Join(result.Suite.Commands, " "), "black benchmark eval --json") {
		t.Fatalf("expected eval JSON command in suite metadata, got %#v", result.Suite.Commands)
	}
	if len(result.Cases) != len(tasks.Scenarios) || len(result.Cases) < 8 {
		t.Fatalf("expected one eval case per task scenario, got cases=%d scenarios=%d", len(result.Cases), len(tasks.Scenarios))
	}
	matrixCaseSeen := false
	for _, item := range result.Cases {
		if !strings.HasPrefix(item.ID, "AI-EVAL-") || item.ScenarioID == "" || item.Prompt == "" {
			t.Fatalf("expected populated eval case identity, got %#v", item)
		}
		if len(item.ExpectedEvidence) == 0 || len(item.RequiredCommands) == 0 {
			t.Fatalf("expected eval case evidence and commands, got %#v", item)
		}
		if scoreMax(item.Scoring) != 100 {
			t.Fatalf("expected eval scoring to total 100, got %#v", item.Scoring)
		}
		if strings.Contains(strings.Join(item.RequiredCommands, " "), "test:e2e:matrix") {
			matrixCaseSeen = true
		}
	}
	if !matrixCaseSeen {
		t.Fatalf("expected at least one eval case to require the browser matrix command, got %#v", result.Cases)
	}
	if result.Totals.CaseCount != len(result.Cases) || result.Totals.TotalRuns != len(result.Cases)*result.Suite.Repeat {
		t.Fatalf("expected eval total runs to follow case count and repeat, got %#v", result.Totals)
	}
	if result.Totals.MaxScorePerRun != len(result.Cases)*100 || result.Totals.MaxScoreAcrossRepeats != len(result.Cases)*result.Suite.Repeat*100 {
		t.Fatalf("expected eval score totals, got %#v", result.Totals)
	}
	if result.Totals.BlackLangTotal != tasks.Totals.BlackLangTotal || result.Totals.ConventionalTotal != tasks.Totals.ConventionalTotal {
		t.Fatalf("expected eval totals to reuse task token totals, got eval=%#v tasks=%#v", result.Totals, tasks.Totals)
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "sentinel.txt" {
		t.Fatalf("expected benchmark eval not to mutate configured out dir, got %#v", entries)
	}
}

func TestFormatAIEvalCorpusIR(t *testing.T) {
	result := AIEvalCorpusResult{
		Success: true,
		Summary: Summary{App: "Warehouse"},
		Suite: AIEvalSuite{
			ID:             "BLACKLANG-WEB-AI-EVAL-v0.2",
			Mode:           "long-running",
			Repeat:         3,
			TimeoutMinutes: 120,
		},
		Baseline: AITaskBenchmarkBaseline{
			SourceLines:            10,
			GeneratedFiles:         2,
			GeneratedLines:         120,
			GeneratedToSourceLines: 12,
		},
		Cases: []AIEvalCase{
			{
				ID:         "AI-EVAL-AI-TASK-TEST-001",
				ScenarioID: "AI-TASK-TEST-001",
				Scoring: []AIEvalScoreCriterion{
					{ID: "source-intent", Points: 30},
					{ID: "agent-learning", Points: 20},
					{ID: "validation", Points: 30},
					{ID: "generated-boundary", Points: 20},
				},
				EstimatedTokens: AITokenEstimate{
					BlackLangTotal:          100,
					ConventionalTotal:       900,
					EstimatedSavingsPercent: 89,
				},
			},
		},
		Totals: AIEvalTotals{
			CaseCount:             1,
			TotalRuns:             3,
			MaxScoreAcrossRepeats: 300,
			EstimatedMinutesTotal: 36,
		},
	}

	ir := FormatAIEvalCorpusIR(result)
	for _, value := range []string{
		"blackir 0.1",
		"benchmark eval ok",
		"app Warehouse",
		"suite BLACKLANG-WEB-AI-EVAL-v0.2 mode long-running repeat 3 timeout_minutes 120",
		"cases 1 total_runs 3 max_score 300 estimated_minutes 36",
		"case AI-EVAL-AI-TASK-TEST-001 scenario AI-TASK-TEST-001 score 100 blacklang_tokens 100 conventional_tokens 900 savings 89",
	} {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected AI eval corpus IR to contain %q, got:\n%s", value, ir)
		}
	}
}
