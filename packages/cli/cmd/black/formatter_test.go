package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatBlackSource(t *testing.T) {
	source := "app   Warehouse\r\nentity Product { sku text label \"Product #1\" // keep comment\r\nstock number default 0}\r\n"
	formatted, diagnostics := FormatBlackSource("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}

	expected := `app Warehouse

entity Product {
  sku text label "Product #1" // keep comment
  stock number default 0
}
`
	if formatted != expected {
		t.Fatalf("unexpected formatted source:\n%s", formatted)
	}
}

func TestFormatBlackFileCheckAndWrite(t *testing.T) {
	file := filepath.Join(t.TempDir(), "app.black")
	source := "app   Warehouse\nentity Product { sku text }\n"
	if err := os.WriteFile(file, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write temp source: %v", err)
	}

	result, _ := FormatBlackFile(file, false, true)
	if result.Success {
		t.Fatalf("expected check to fail when formatting is required")
	}
	if !result.Changed || len(result.Errors) != 1 || result.Errors[0].Code != "FORMAT_REQUIRED" {
		t.Fatalf("expected FORMAT_REQUIRED result, got %#v", result)
	}

	result, _ = FormatBlackFile(file, true, false)
	if !result.Success || !result.Changed {
		t.Fatalf("expected write format to succeed and change file, got %#v", result)
	}

	result, _ = FormatBlackFile(file, false, true)
	if !result.Success || result.Changed {
		t.Fatalf("expected formatted file to pass check, got %#v", result)
	}
}

func TestFormatBlackSourceKeepsBraceComments(t *testing.T) {
	source := `app Warehouse
entity Product { // product fields
sku text
} // end product
`
	formatted, diagnostics := FormatBlackSource("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}

	expected := `app Warehouse

entity Product { // product fields
  sku text
} // end product
`
	if formatted != expected {
		t.Fatalf("unexpected formatted source:\n%s", formatted)
	}
}

func TestFormatBlackSourceReportsLexerDiagnostics(t *testing.T) {
	_, diagnostics := FormatBlackSource("test.black", "app \"Warehouse\n")
	if len(diagnostics) != 1 || diagnostics[0].Code != "UNCLOSED_STRING" {
		t.Fatalf("expected UNCLOSED_STRING diagnostic, got %#v", diagnostics)
	}
}
