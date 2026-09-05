package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	root := t.TempDir()
	content := `version = "0.1"
target = "web"
source = "src/app.black"
out = "generated"
theme = "theme.blackthm"
`
	if err := os.WriteFile(filepath.Join(root, "blacklang.toml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	config := LoadConfig(root)
	if config.Version != "0.1" {
		t.Fatalf("expected version 0.1, got %q", config.Version)
	}
	if config.Target != "web" {
		t.Fatalf("expected target web, got %q", config.Target)
	}
	if config.Source != "src/app.black" {
		t.Fatalf("expected source src/app.black, got %q", config.Source)
	}
	if config.Out != "generated" {
		t.Fatalf("expected out generated, got %q", config.Out)
	}
	if config.Theme != "theme.blackthm" {
		t.Fatalf("expected theme theme.blackthm, got %q", config.Theme)
	}
}
