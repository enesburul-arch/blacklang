package main

import (
	"strings"
	"testing"
)

func TestFormatInspectIR(t *testing.T) {
	program, diagnostics := Parse("test.black", `app Warehouse

database {
  url env DATABASE_URL
}

entity Product {
  sku text required unique
}

page Products {
  source Product
  actions create, edit, delete
}
`)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}

	result := InspectResult{
		Success: true,
		Command: "inspect",
		Version: version,
		Config: ConfigInfo{
			LanguageVersion: "0.1",
			Target:          "web",
			Source:          "src/app.black",
			Out:             "generated",
		},
		Summary: Summary{
			App:      "Warehouse",
			Entities: 1,
			Pages:    1,
		},
		Program: program,
		Errors:  []Diagnostic{},
	}

	ir := FormatInspectIR(result)
	expected := []string{
		"inspect ok",
		"language 0.1",
		"target web",
		"source src/app.black",
		"database url env DATABASE_URL",
		"entity Product fields 1",
		"page Products source Product actions create edit delete",
	}
	for _, value := range expected {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected IR to contain %q, got:\n%s", value, ir)
		}
	}
}
