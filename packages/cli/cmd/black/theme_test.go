package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTheme(t *testing.T) {
	theme, diagnostics := ParseTheme("theme.blackthm", `blackthm WarehouseTheme {
  version 1
  target web
  locked false

  token color primary "#2563eb"
  token space sm 8

  profile UICompact {
    version 1
    mode box color width style pt pr pb pl radius place
    mode text color size weight align
  }
}
`)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if theme.Name != "WarehouseTheme" || theme.Version != 1 || theme.Target != "web" || theme.Locked {
		t.Fatalf("unexpected theme metadata: %#v", theme)
	}
	if len(theme.Tokens) != 2 {
		t.Fatalf("expected two tokens, got %#v", theme.Tokens)
	}
	if theme.Tokens[0].Value != "#2563eb" {
		t.Fatalf("expected quoted hex token value, got %#v", theme.Tokens[0])
	}
	if theme.Profile.Name != "UICompact" || theme.Profile.Version != 1 {
		t.Fatalf("unexpected profile metadata: %#v", theme.Profile)
	}
	if len(theme.Profile.Modes) != 2 {
		t.Fatalf("expected two modes, got %#v", theme.Profile.Modes)
	}
	if strings.Join(theme.Profile.Modes[0].Slots, " ") != "color width style pt pr pb pl radius place" {
		t.Fatalf("unexpected box slots: %#v", theme.Profile.Modes[0])
	}
}

func TestInspectThemeUsesConfigPath(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "blacklang.toml"), `theme = "theme.blackthm"
`)
	writeTestFile(t, filepath.Join(root, "theme.blackthm"), `blackthm AppTheme {
  version 1
  target web
  locked false

  profile UICompact {
    version 1
    mode box color width
  }
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

	result := InspectTheme([]string{})
	if !result.Success {
		t.Fatalf("expected theme inspect success, got %#v", result.Errors)
	}
	if result.File != "theme.blackthm" || result.Theme.Name != "AppTheme" {
		t.Fatalf("expected config theme path, got %#v", result)
	}
}

func TestParseThemeReportsStableDiagnostics(t *testing.T) {
	_, diagnostics := ParseTheme("theme.blackthm", `blackthm BrokenTheme {
  version one
  token color primary "#2563eb"
  token color primary "#111111"

  profile UICompact {
    version 1
    mode box color width
    mode box color
  }
}
`)
	codes := diagnosticCodes(diagnostics)
	for _, code := range []string{"INVALID_THEME_VERSION", "DUPLICATE_THEME_TOKEN", "DUPLICATE_UI_MODE"} {
		if !codes[code] {
			t.Fatalf("expected diagnostic %s, got %#v", code, diagnostics)
		}
	}
}

func TestParseThemeRequiresProfile(t *testing.T) {
	_, diagnostics := ParseTheme("theme.blackthm", `blackthm EmptyTheme {
  version 1
}
`)
	codes := diagnosticCodes(diagnostics)
	if !codes["MISSING_UI_PROFILE"] {
		t.Fatalf("expected MISSING_UI_PROFILE, got %#v", diagnostics)
	}
}

func TestFormatThemeIR(t *testing.T) {
	theme, diagnostics := ParseTheme("theme.blackthm", `blackthm WarehouseTheme {
  version 1
  target web
  locked false

  token color primary "#2563eb"

  profile UICompact {
    version 1
    mode box color width
  }
}
`)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}

	ir := FormatThemeIR(ThemeInspectResult{
		Success: true,
		Command: "theme inspect",
		Version: version,
		File:    "theme.blackthm",
		Theme:   theme,
		Errors:  []Diagnostic{},
	})
	for _, value := range []string{
		"theme inspect ok",
		"theme WarehouseTheme version 1 target web locked false",
		"color primary \"#2563eb\"",
		"profile UICompact version 1",
		"mode box slots color width",
	} {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected IR to contain %q, got:\n%s", value, ir)
		}
	}
}

func diagnosticCodes(diagnostics []Diagnostic) map[string]bool {
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	return codes
}
