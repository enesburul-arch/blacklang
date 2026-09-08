package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBenchmarkProjectMeasuresGeneratedOutputWithoutMutatingOutDir(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "app.black")
	source := `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required
  stock number default 0
}

page Products {
  source Product

  table {
    columns sku, stock
  }

  form {
    fields sku, stock
  }

  actions create, edit
}
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(root, "generated-out")

	result := BenchmarkProject([]string{sourcePath, "--out", outDir})
	if !result.Success {
		t.Fatalf("expected benchmark success, got %#v", result.Errors)
	}
	if result.Source.Files != 1 || result.Source.Lines == 0 {
		t.Fatalf("expected source metrics, got %#v", result.Source)
	}
	if result.Generated.Files == 0 || result.Generated.Lines <= result.Source.Lines {
		t.Fatalf("expected generated output metrics, got source=%#v generated=%#v", result.Source, result.Generated)
	}
	if result.Ratios.GeneratedToSourceLines <= 1 {
		t.Fatalf("expected generated/source ratio above 1, got %#v", result.Ratios)
	}
	if _, err := os.Stat(outDir); !os.IsNotExist(err) {
		t.Fatalf("expected benchmark not to mutate configured out dir %s, stat err=%v", outDir, err)
	}
	if !benchmarkFileContains(result.GeneratedFiles, filepath.ToSlash(filepath.Join(outDir, "package.json")), "package") {
		t.Fatalf("expected package.json generated metric with output label, got %#v", result.GeneratedFiles)
	}
	if !benchmarkKindContains(result.GeneratedKinds, "react-page") {
		t.Fatalf("expected generated kind summary to include react-page, got %#v", result.GeneratedKinds)
	}
}

func TestFormatBenchmarkIR(t *testing.T) {
	result := BenchmarkResult{
		Success: true,
		Summary: Summary{App: "Warehouse"},
		Source: BenchmarkTotals{
			Files: 1,
			Lines: 10,
			Bytes: 100,
		},
		Generated: BenchmarkTotals{
			Files: 2,
			Lines: 120,
			Bytes: 2000,
		},
		Ratios: BenchmarkRatios{
			GeneratedToSourceLines: 12,
		},
		GeneratedKinds: []BenchmarkKindMetric{
			{Kind: "react-page", Files: 1, Lines: 80, Bytes: 1200},
		},
	}

	ir := FormatBenchmarkIR(result)
	for _, value := range []string{
		"blackir 0.1",
		"benchmark ok",
		"app Warehouse",
		"source files 1 lines 10 bytes 100",
		"generated files 2 lines 120 bytes 2000",
		"ratio generated_to_source_lines 12.00",
		"kind react-page files 1 lines 80 bytes 1200",
	} {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected benchmark IR to contain %q, got:\n%s", value, ir)
		}
	}
}

func benchmarkFileContains(files []BenchmarkFileMetric, path string, kind string) bool {
	for _, file := range files {
		if file.Path == path && file.Kind == kind {
			return true
		}
	}
	return false
}

func benchmarkKindContains(kinds []BenchmarkKindMetric, kind string) bool {
	for _, metric := range kinds {
		if metric.Kind == kind {
			return true
		}
	}
	return false
}
