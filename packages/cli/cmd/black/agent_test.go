package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentStartupChecklistUsesProjectConfig(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "AGENTS.md"), "# Rules\n")
	writeTestFile(t, filepath.Join(root, "BLACKLANG.md"), "# BlackLang\n")
	writeTestFile(t, filepath.Join(root, "SPEC.md"), "# Spec\n")
	writeTestFile(t, filepath.Join(root, "docs", "diagnostics.md"), "# Diagnostics\n")
	writeTestFile(t, filepath.Join(root, "blacklang.toml"), `version = "0.1"
target = "web"
source = "src/app.black"
out = "generated"
`)
	writeTestFile(t, filepath.Join(root, "src", "app.black"), `app Warehouse

entity Product {
  sku text required unique
}

page Products {
  source Product
  actions create, edit, delete
}
`)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("expected cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir temp project: %v", err)
	}

	result := AgentStartupChecklist([]string{})
	if !result.Success {
		t.Fatalf("expected startup checklist success, got %#v", result.Errors)
	}
	if result.Config.Source != "src/app.black" {
		t.Fatalf("expected config source, got %#v", result.Config)
	}
	if result.Config.Out != "generated" {
		t.Fatalf("expected config out, got %#v", result.Config)
	}
	if result.Summary.App != "Warehouse" || result.Summary.Entities != 1 || result.Summary.Pages != 1 {
		t.Fatalf("expected project summary, got %#v", result.Summary)
	}
	assertStartupReadFile(t, result.ReadFirst, "AGENTS.md", true)
	assertStartupReadFile(t, result.ReadFirst, "docs/diagnostics.md", true)
	assertStartupReadFile(t, result.ReadFirst, "src/app.black", true)
	assertStartupCommand(t, result.Commands, "inspect", "black inspect src/app.black --json")
	assertStartupCommand(t, result.Commands, "build", "black build src/app.black --out generated --json")
	if len(result.Checklist) < 8 {
		t.Fatalf("expected detailed checklist, got %#v", result.Checklist)
	}
	if !strings.Contains(strings.Join(result.Policies, "\n"), "generated output") {
		t.Fatalf("expected generated output policy, got %#v", result.Policies)
	}
}

func TestAgentStartupChecklistReportsSourceErrors(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "blacklang.toml"), `source = "missing/app.black"
out = "generated"
`)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("expected cwd: %v", err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir temp project: %v", err)
	}

	result := AgentStartupChecklist([]string{})
	if result.Success {
		t.Fatalf("expected startup checklist to report invalid project")
	}
	if len(result.Errors) != 1 || result.Errors[0].Code != "FILE_READ_ERROR" {
		t.Fatalf("expected FILE_READ_ERROR, got %#v", result.Errors)
	}
	assertStartupReadFile(t, result.ReadFirst, "missing/app.black", false)
	assertStartupCommand(t, result.Commands, "diagnostics", "black docs diagnostics --json")
}

func TestFormatAgentStartupIR(t *testing.T) {
	result := AgentStartupResult{
		Success: true,
		Command: "agent startup",
		Version: version,
		Config: ConfigInfo{
			LanguageVersion: "0.1",
			Target:          "web",
			Source:          "src/app.black",
			Out:             "generated",
		},
		Summary: Summary{App: "Warehouse", Entities: 1, Pages: 1},
		ReadFirst: []AgentReadFile{
			{Path: "AGENTS.md", Purpose: "Rules", Exists: true},
		},
		Checklist: []AgentChecklistItem{
			{Step: 1, Action: "Read files.", Reason: "Learn rules."},
		},
		Commands: []AgentCommand{
			{Name: "inspect", Command: "black inspect src/app.black --json", Purpose: "Inspect project structure."},
		},
		Errors: []Diagnostic{},
	}

	ir := FormatAgentStartupIR(result)
	for _, value := range []string{
		"agent startup ok",
		"language 0.1",
		"target web",
		"source src/app.black",
		"readFirst",
		"commands",
		"black inspect src/app.black --json",
	} {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected IR to contain %q, got:\n%s", value, ir)
		}
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertStartupReadFile(t *testing.T, files []AgentReadFile, path string, exists bool) {
	t.Helper()
	for _, file := range files {
		if file.Path == path {
			if file.Exists != exists {
				t.Fatalf("expected %s exists=%v, got %#v", path, exists, file)
			}
			return
		}
	}
	t.Fatalf("expected readFirst to include %s, got %#v", path, files)
}

func assertStartupCommand(t *testing.T, commands []AgentCommand, name string, expected string) {
	t.Helper()
	for _, command := range commands {
		if command.Name == name {
			if command.Command != expected {
				t.Fatalf("expected command %s to be %q, got %#v", name, expected, command)
			}
			return
		}
	}
	t.Fatalf("expected command list to include %s, got %#v", name, commands)
}
