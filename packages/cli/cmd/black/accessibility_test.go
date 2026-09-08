package main

import (
	"strings"
	"testing"
)

func TestAccessibilityAuditPassesForTitledOverlaysAndGroups(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    order table, detail, form
    compose grid columns 2 gap md stackAt md
    section detail display drawer side right title "Product Details"
    section form display modal title "Product Form"
  }

  table {
    columns name
  }

  form {
    fields name
  }

  actions create, edit
}

page Customers {
  source Product

  view {
    order table, detail, form
    compose grid columns 2 gap md stackAt md
    group Record sections detail, form compose stack gap md title "Record Workspace"
  }

  table {
    columns name
  }

  form {
    fields name
  }

  actions create, edit
}
`
	program, diagnostics := Parse("a11y.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("parse: %#v", diagnostics)
	}
	if diagnostics = Validate(program); len(diagnostics) != 0 {
		t.Fatalf("validate: %#v", diagnostics)
	}
	findings := AccessibilityDiagnostics(program)
	if len(findings) != 0 {
		t.Fatalf("expected no accessibility findings, got %#v", findings)
	}
}

func TestAccessibilityAuditFindsMissingOverlayAndGroupTitles(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    order table, detail, form
    compose grid columns 2 gap md stackAt md
    section form display modal
  }

  table {
    columns name
  }

  form {
    fields name
  }

  actions create, edit
}

page Customers {
  source Product

  view {
    order table, detail, form
    compose grid columns 2 gap md stackAt md
    group Record sections detail, form compose stack gap md
  }

  table {
    columns name
  }

  form {
    fields name
  }

  actions create, edit
}
`
	program, diagnostics := Parse("a11y.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("parse: %#v", diagnostics)
	}
	if diagnostics = Validate(program); len(diagnostics) != 0 {
		t.Fatalf("validate: %#v", diagnostics)
	}

	findings := AccessibilityDiagnostics(program)
	codes := map[string]bool{}
	for _, finding := range findings {
		codes[finding.Code] = true
		if finding.File != "a11y.black" || finding.Line < 1 || finding.Column < 1 {
			t.Fatalf("expected positioned accessibility finding, got %#v", finding)
		}
	}
	for _, code := range []string{"ACCESSIBILITY_MISSING_OVERLAY_TITLE", "ACCESSIBILITY_MISSING_GROUP_TITLE"} {
		if !codes[code] {
			t.Fatalf("expected accessibility code %s, got %#v", code, findings)
		}
	}
}

func TestFormatAccessibilityAuditIR(t *testing.T) {
	result := AccessibilityAuditResult{
		Success: false,
		Command: "audit accessibility",
		Version: version,
		File:    "a11y.black",
		Findings: []Diagnostic{{
			File:   "a11y.black",
			Line:   10,
			Column: 1,
			Code:   "ACCESSIBILITY_MISSING_OVERLAY_TITLE",
		}},
		Errors: []Diagnostic{},
	}
	ir := FormatAccessibilityAuditIR(result)
	for _, expected := range []string{
		"audit accessibility findings",
		"file a11y.black",
		"findings 1",
		"ACCESSIBILITY_MISSING_OVERLAY_TITLE",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected IR to contain %q, got:\n%s", expected, ir)
		}
	}
}
