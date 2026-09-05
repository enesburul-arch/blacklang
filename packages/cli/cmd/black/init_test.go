package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitProjectWritesScaffold(t *testing.T) {
	root := t.TempDir()
	result := InitProject(root)
	if !result.Success {
		t.Fatalf("expected init success, got %#v", result.Errors)
	}
	if len(result.Files) != 4 {
		t.Fatalf("expected 4 files, got %d", len(result.Files))
	}

	expected := []string{
		"blacklang.toml",
		filepath.Join("src", "app.black"),
		"AGENTS.md",
		"BLACKLANG.md",
	}
	for _, relativePath := range expected {
		if _, err := os.Stat(filepath.Join(root, relativePath)); err != nil {
			t.Fatalf("expected %s to exist: %v", relativePath, err)
		}
	}
}

func TestInitProjectRefusesOverwrite(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "blacklang.toml"), []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	result := InitProject(root)
	if result.Success {
		t.Fatalf("expected init to fail")
	}
	if len(result.Errors) == 0 {
		t.Fatalf("expected overwrite diagnostic")
	}
	if result.Errors[0].Code != "INIT_FILE_EXISTS" {
		t.Fatalf("expected INIT_FILE_EXISTS, got %q", result.Errors[0].Code)
	}
}
