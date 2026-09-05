package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintFilePassesFormattedValidSource(t *testing.T) {
	file := filepath.Join(t.TempDir(), "app.black")
	source := `app Warehouse

database {
  url env DATABASE_URL
}

entity Product {
  sku text
}

page Products {
  source Product

  table {
    columns sku
  }

  form {
    fields sku
  }

  actions create
}
`
	formatted, diagnostics := FormatBlackSource(file, source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected source to format cleanly, got %#v", diagnostics)
	}
	if err := os.WriteFile(file, []byte(formatted), 0644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	result := LintFile(file)
	if !result.Success {
		t.Fatalf("expected lint success, got %#v", result)
	}
	if len(result.Checks) != 4 {
		t.Fatalf("expected 4 checks, got %#v", result.Checks)
	}
	if result.Summary.App != "Warehouse" || result.Summary.Entities != 1 || result.Summary.Pages != 1 {
		t.Fatalf("expected lint summary, got %#v", result.Summary)
	}
}

func TestLintFileReportsFormatValidateAndSecurityFindings(t *testing.T) {
	file := filepath.Join(t.TempDir(), "app.black")
	source := `app   Warehouse
# token example_token_value_1234567890
entity Product { sku text }
page Products {
source Product
table {
columns missing
}
}
`
	if err := os.WriteFile(file, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	result := LintFile(file)
	if result.Success {
		t.Fatalf("expected lint failure")
	}

	codes := map[string]bool{}
	for _, finding := range result.Findings {
		codes[finding.Code] = true
	}
	for _, code := range []string{"FORMAT_REQUIRED", "UNKNOWN_TABLE_COLUMN", "HARDCODED_TOKEN"} {
		if !codes[code] {
			t.Fatalf("expected finding code %s, got %#v", code, result.Findings)
		}
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no command errors, got %#v", result.Errors)
	}
}

func TestLintFileReportsReadError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "missing.black")
	result := LintFile(file)
	if result.Success {
		t.Fatalf("expected read error")
	}
	if len(result.Errors) != 1 || result.Errors[0].Code != "FILE_READ_ERROR" {
		t.Fatalf("expected FILE_READ_ERROR, got %#v", result)
	}
}
