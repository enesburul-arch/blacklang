package main

import (
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func BenchmarkProject(args []string) BenchmarkResult {
	project := LoadProject(args)
	result := BenchmarkResult{
		Success:        false,
		Command:        "benchmark",
		Version:        version,
		Config:         project.ConfigInfo(),
		Summary:        project.Summary(),
		SourceFiles:    []BenchmarkFileMetric{},
		GeneratedFiles: []BenchmarkFileMetric{},
		GeneratedKinds: []BenchmarkKindMetric{},
		Errors:         []Diagnostic{},
	}

	if len(project.Diagnostics) > 0 {
		result.Errors = project.Diagnostics
		return result
	}

	sourceFiles := []BenchmarkFileMetric{}
	sourceMetric, diagnostic := countBenchmarkFile(project.SourcePath, "black-source", project.SourcePath)
	if diagnostic != nil {
		result.Errors = append(result.Errors, *diagnostic)
		return result
	}
	sourceFiles = append(sourceFiles, sourceMetric)

	var theme *ThemeDecl
	if project.Config.Theme != "" {
		loadedTheme, themeDiagnostics := LoadTheme(project.Config.Theme)
		if len(themeDiagnostics) > 0 {
			result.Errors = append(result.Errors, themeDiagnostics...)
			return result
		}
		theme = &loadedTheme
		themeMetric, diagnostic := countBenchmarkFile(project.Config.Theme, "theme", project.Config.Theme)
		if diagnostic != nil {
			result.Errors = append(result.Errors, *diagnostic)
			return result
		}
		sourceFiles = append(sourceFiles, themeMetric)
	}
	result.SourceFiles = sourceFiles
	result.Source = benchmarkTotals(sourceFiles)

	tempDir, err := os.MkdirTemp("", "blacklang-benchmark-*")
	if err != nil {
		result.Errors = append(result.Errors, Diagnostic{
			Code:       "BENCHMARK_TEMP_ERROR",
			Message:    err.Error(),
			Suggestion: "Ensure the system temp directory is writable, then retry black benchmark.",
		})
		return result
	}
	defer os.RemoveAll(tempDir)

	files, diagnostics := BuildWebWithTheme(project.Program, tempDir, theme)
	if len(diagnostics) > 0 {
		result.Errors = append(result.Errors, diagnostics...)
		return result
	}

	generatedFiles := []BenchmarkFileMetric{}
	for _, file := range files {
		displayPath := benchmarkGeneratedDisplayPath(tempDir, project.OutDir, file.Path)
		metric, diagnostic := countBenchmarkFile(file.Path, file.Kind, displayPath)
		if diagnostic != nil {
			result.Errors = append(result.Errors, *diagnostic)
			return result
		}
		generatedFiles = append(generatedFiles, metric)
	}
	sort.Slice(generatedFiles, func(i, j int) bool {
		return generatedFiles[i].Path < generatedFiles[j].Path
	})
	result.GeneratedFiles = generatedFiles
	result.Generated = benchmarkTotals(generatedFiles)
	result.GeneratedKinds = benchmarkKindTotals(generatedFiles)
	result.Ratios = benchmarkRatios(result.Source.Lines, result.Generated.Lines)
	result.Success = true
	return result
}

func countBenchmarkFile(path string, kind string, displayPath string) (BenchmarkFileMetric, *Diagnostic) {
	data, err := os.ReadFile(path)
	if err != nil {
		return BenchmarkFileMetric{}, &Diagnostic{
			File:       displayPath,
			Code:       "BENCHMARK_FILE_READ_ERROR",
			Message:    err.Error(),
			Suggestion: "Ensure benchmark input and generated output files are readable.",
		}
	}
	return BenchmarkFileMetric{
		Path:  filepath.ToSlash(displayPath),
		Kind:  kind,
		Lines: countLines(string(data)),
		Bytes: int64(len(data)),
	}, nil
}

func countLines(text string) int {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.TrimSuffix(normalized, "\n")
	if normalized == "" {
		return 0
	}
	return strings.Count(normalized, "\n") + 1
}

func benchmarkTotals(files []BenchmarkFileMetric) BenchmarkTotals {
	totals := BenchmarkTotals{Files: len(files)}
	for _, file := range files {
		totals.Lines += file.Lines
		totals.Bytes += file.Bytes
	}
	return totals
}

func benchmarkKindTotals(files []BenchmarkFileMetric) []BenchmarkKindMetric {
	byKind := map[string]BenchmarkKindMetric{}
	for _, file := range files {
		metric := byKind[file.Kind]
		metric.Kind = file.Kind
		metric.Files++
		metric.Lines += file.Lines
		metric.Bytes += file.Bytes
		byKind[file.Kind] = metric
	}
	kinds := make([]string, 0, len(byKind))
	for kind := range byKind {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	result := make([]BenchmarkKindMetric, 0, len(kinds))
	for _, kind := range kinds {
		result = append(result, byKind[kind])
	}
	return result
}

func benchmarkRatios(sourceLines int, generatedLines int) BenchmarkRatios {
	ratios := BenchmarkRatios{}
	if sourceLines > 0 {
		ratios.GeneratedToSourceLines = roundBenchmarkRatio(float64(generatedLines) / float64(sourceLines))
	}
	if generatedLines > 0 {
		ratios.SourceToGeneratedLines = roundBenchmarkRatio(float64(sourceLines) / float64(generatedLines))
	}
	return ratios
}

func roundBenchmarkRatio(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return math.Round(value*100) / 100
}

func benchmarkGeneratedDisplayPath(tempDir string, outDir string, generatedPath string) string {
	relative, err := filepath.Rel(tempDir, generatedPath)
	if err != nil || strings.HasPrefix(relative, "..") {
		return generatedPath
	}
	if outDir == "" {
		outDir = "generated"
	}
	return filepath.Join(outDir, relative)
}
