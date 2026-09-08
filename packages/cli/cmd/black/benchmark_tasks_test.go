package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBenchmarkTasksReportsAITaskScenariosWithoutMutatingOutDir(t *testing.T) {
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

	result := BenchmarkTasks([]string{sourcePath, "--out", outDir})
	if !result.Success {
		t.Fatalf("expected benchmark tasks success, got %#v", result.Errors)
	}
	if result.Command != "benchmark tasks" || result.Summary.App != "Warehouse" {
		t.Fatalf("expected benchmark tasks identity, got %#v", result)
	}
	if result.Baseline.SourceLines == 0 || result.Baseline.GeneratedLines == 0 || result.Baseline.GeneratedToSourceLines <= 1 {
		t.Fatalf("expected populated baseline metrics, got %#v", result.Baseline)
	}
	if len(result.Scenarios) < 8 {
		t.Fatalf("expected multiple AI task scenarios, got %#v", result.Scenarios)
	}
	testScenario, ok := findAITaskScenario(result.Scenarios, "AI-TASK-TEST-001")
	if !ok {
		t.Fatalf("expected browser test task scenario, got %#v", result.Scenarios)
	}
	if !strings.Contains(strings.Join(testScenario.Commands, " "), "npm run test:e2e") || !strings.Contains(strings.Join(testScenario.GeneratedImpact, " "), "blacklang.e2e.test.ts") {
		t.Fatalf("expected browser test task scenario to include generated e2e evidence, got %#v", testScenario)
	}
	deployScenario, ok := findAITaskScenario(result.Scenarios, "AI-TASK-DEPLOY-001")
	if !ok {
		t.Fatalf("expected deploy task scenario, got %#v", result.Scenarios)
	}
	if !strings.Contains(strings.Join(deployScenario.Commands, " "), "deploy:rollback:plan") || !strings.Contains(strings.Join(deployScenario.GeneratedImpact, " "), "docker-compose.preview.yml") {
		t.Fatalf("expected deploy task scenario to include preview and rollback evidence, got %#v", deployScenario)
	}
	if result.Totals.ScenarioCount != len(result.Scenarios) || result.Totals.EstimatedSavingsPercent <= 0 {
		t.Fatalf("expected actionable token totals, got %#v", result.Totals)
	}
	if result.Totals.BlackLangTotal >= result.Totals.ConventionalTotal {
		t.Fatalf("expected BlackLang token total below conventional estimate, got %#v", result.Totals)
	}
	for _, scenario := range result.Scenarios {
		if scenario.EstimatedTokens.BlackLangTotal == 0 || scenario.EstimatedTokens.ConventionalTotal == 0 || len(scenario.EstimatedTokens.Basis) == 0 {
			t.Fatalf("expected scenario token estimate basis, got %#v", scenario)
		}
	}
	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		t.Fatalf("expected benchmark tasks not to mutate configured out dir %s, stat err=%v", outDir, err)
	}
}

func TestFormatAITaskBenchmarkIR(t *testing.T) {
	result := AITaskBenchmarkResult{
		Success: true,
		Summary: Summary{App: "Warehouse"},
		Baseline: AITaskBenchmarkBaseline{
			SourceLines:            10,
			GeneratedFiles:         2,
			GeneratedLines:         120,
			GeneratedToSourceLines: 12,
		},
		Scenarios: []AITaskScenario{
			{
				ID: "AI-TASK-TEST-001",
				EstimatedTokens: AITokenEstimate{
					BlackLangTotal:          100,
					ConventionalTotal:       900,
					EstimatedSavingsPercent: 89,
				},
			},
		},
		Totals: AITaskBenchmarkTotals{
			ScenarioCount:           1,
			BlackLangTotal:          100,
			ConventionalTotal:       900,
			EstimatedSavingsPercent: 89,
		},
	}

	ir := FormatAITaskBenchmarkIR(result)
	for _, value := range []string{
		"blackir 0.1",
		"benchmark tasks ok",
		"app Warehouse",
		"baseline source_lines 10 generated_lines 120 generated_files 2 ratio 12.00",
		"scenarios 1",
		"scenario AI-TASK-TEST-001 blacklang_tokens 100 conventional_tokens 900 savings 89",
		"totals blacklang_tokens 100 conventional_tokens 900 savings 89",
	} {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected AI task benchmark IR to contain %q, got:\n%s", value, ir)
		}
	}
}

func findAITaskScenario(scenarios []AITaskScenario, id string) (AITaskScenario, bool) {
	for _, scenario := range scenarios {
		if scenario.ID == id {
			return scenario, true
		}
	}
	return AITaskScenario{}, false
}
