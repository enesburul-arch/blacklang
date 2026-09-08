package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type goldenBuildManifest struct {
	App   string             `json:"app"`
	Files []goldenFileMetric `json:"files"`
}

type goldenFileMetric struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Lines  int    `json:"lines"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}

func TestWarehouseGeneratedGoldenManifest(t *testing.T) {
	root := testRepoRoot(t)
	sourcePath := filepath.Join(root, "examples", "warehouse", "app.black")
	themePath := filepath.Join(root, "examples", "warehouse", "theme.blackthm")
	goldenPath := filepath.Join("testdata", "golden", "warehouse-build-manifest.json")

	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	program, parseDiagnostics := Parse(sourcePath, string(source))
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	if validateDiagnostics := Validate(program); len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validate diagnostics, got %#v", validateDiagnostics)
	}
	theme, themeDiagnostics := LoadTheme(themePath)
	if len(themeDiagnostics) != 0 {
		t.Fatalf("expected no theme diagnostics, got %#v", themeDiagnostics)
	}

	outDir := t.TempDir()
	files, buildDiagnostics := BuildWebWithTheme(program, outDir, &theme)
	if len(buildDiagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", buildDiagnostics)
	}
	manifest := goldenManifestFromBuild(t, program.App.Name, outDir, "generated", files)
	actual := marshalGoldenManifest(t, manifest)

	if os.Getenv("UPDATE_BLACKLANG_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, actual, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(actual)) != strings.TrimSpace(string(expected)) {
		t.Fatalf("warehouse generated golden manifest changed.\nRun UPDATE_BLACKLANG_GOLDEN=1 go test ./cmd/black to update the reviewed manifest.")
	}
}

func goldenManifestFromBuild(t *testing.T, appName string, tempDir string, outLabel string, files []GeneratedFile) goldenBuildManifest {
	t.Helper()
	manifest := goldenBuildManifest{App: appName, Files: []goldenFileMetric{}}
	for _, file := range files {
		data, err := os.ReadFile(file.Path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		manifest.Files = append(manifest.Files, goldenFileMetric{
			Path:   filepath.ToSlash(benchmarkGeneratedDisplayPath(tempDir, outLabel, file.Path)),
			Kind:   file.Kind,
			Lines:  countLines(string(data)),
			Bytes:  int64(len(data)),
			SHA256: hex.EncodeToString(sum[:]),
		})
	}
	sort.Slice(manifest.Files, func(i, j int) bool {
		return manifest.Files[i].Path < manifest.Files[j].Path
	})
	return manifest
}

func marshalGoldenManifest(t *testing.T, manifest goldenBuildManifest) []byte {
	t.Helper()
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

func testRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}
