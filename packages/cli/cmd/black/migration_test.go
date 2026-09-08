package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeSchemaMigrationPlanReportsRiskLevels(t *testing.T) {
	dir := t.TempDir()
	oldFile := writeMigrationTestFile(t, dir, "old.black", `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity Customer {
  name text required
}

entity Product {
  sku text unique
  name text required
  stock integer
  price money
  index stock
}

page Products {
  source Product
  table {
    columns sku, name, stock, price
  }
  form {
    fields sku, name, stock, price
  }
  actions create
}
`)
	newFile := writeMigrationTestFile(t, dir, "new.black", `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity Customer {
  name text required
}

entity Product {
  sku text unique
  name text required
  price decimal
  status text required default "active"
  owner Customer required
  index sku, price
}

page Products {
  source Product
  table {
    columns sku, name, price, status
  }
  form {
    fields sku, name, price, status, owner
  }
  actions create
}
`)

	result := AnalyzeSchemaMigrationPlan(oldFile, newFile)
	if !result.Success {
		t.Fatalf("expected migration plan success, got errors %#v", result.Errors)
	}
	if result.Safe {
		t.Fatalf("expected unsafe plan because destructive/manual changes exist, got %#v", result)
	}
	if !result.Destructive {
		t.Fatalf("expected destructive plan, got %#v", result)
	}
	expectedChanges := []struct {
		changeType string
		risk       string
		entity     string
		field      string
	}{
		{"drop-column", "destructive", "Product", "stock"},
		{"change-column-type", "destructive", "Product", "price"},
		{"add-column", "safe", "Product", "status"},
		{"add-column", "manual", "Product", "owner"},
		{"add-index", "safe", "Product", ""},
		{"drop-index", "safe", "Product", ""},
	}
	for _, expected := range expectedChanges {
		if !schemaMigrationChangeExists(result.Changes, expected.changeType, expected.risk, expected.entity, expected.field) {
			t.Fatalf("expected change %s risk=%s target=%s.%s in %#v", expected.changeType, expected.risk, expected.entity, expected.field, result.Changes)
		}
	}
	if result.Summary.FieldsAdded != 2 || result.Summary.FieldsRemoved != 1 || result.Summary.IndexesAdded != 1 || result.Summary.IndexesRemoved != 1 {
		t.Fatalf("unexpected summary counts: %#v", result.Summary)
	}
	if result.Summary.DestructiveChanges == 0 || result.Summary.ManualChanges == 0 || result.Summary.SafeChanges == 0 {
		t.Fatalf("expected all risk levels in summary, got %#v", result.Summary)
	}
}

func TestAnalyzeSchemaMigrationPlanReportsNoOpAsSafe(t *testing.T) {
	dir := t.TempDir()
	source := `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text unique
  name text required
  index sku
}

page Products {
  source Product
  table {
    columns sku, name
  }
  form {
    fields sku, name
  }
  actions create
}
`
	oldFile := writeMigrationTestFile(t, dir, "old.black", source)
	newFile := writeMigrationTestFile(t, dir, "new.black", source)

	result := AnalyzeSchemaMigrationPlan(oldFile, newFile)
	if !result.Success || !result.Safe || result.Destructive {
		t.Fatalf("expected safe successful no-op plan, got %#v", result)
	}
	if len(result.Changes) != 0 {
		t.Fatalf("expected no changes, got %#v", result.Changes)
	}
	if len(result.Steps) != 1 || !strings.Contains(result.Steps[0].Action, "No schema migration") {
		t.Fatalf("expected no-op step, got %#v", result.Steps)
	}
}

func TestAnalyzeSchemaMigrationPlanUsesExplicitRenames(t *testing.T) {
	dir := t.TempDir()
	oldFile := writeMigrationTestFile(t, dir, "old.black", `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity ProductItem {
  title text required
  sku text unique
  index title
}

entity Purchase {
  product ProductItem required
  index product
}

page Products {
  source ProductItem
  table {
    columns title, sku
  }
  form {
    fields title, sku
  }
  actions create
}
`)
	newFile := writeMigrationTestFile(t, dir, "new.black", `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  name text required
  sku text unique
  index name
}

entity Purchase {
  product Product required
  index product
}

migration RenameProductItem {
  rename entity ProductItem to Product
  rename field Product.title to name
}

page Products {
  source Product
  table {
    columns name, sku
  }
  form {
    fields name, sku
  }
  actions create
}
`)

	result := AnalyzeSchemaMigrationPlan(oldFile, newFile)
	if !result.Success {
		t.Fatalf("expected migration plan success, got errors %#v", result.Errors)
	}
	if !result.Safe || result.Destructive {
		t.Fatalf("expected safe non-destructive rename plan, got %#v", result)
	}
	for _, expected := range []struct {
		changeType string
		entity     string
		field      string
	}{
		{"rename-table", "Product", ""},
		{"rename-column", "Product", "name"},
	} {
		if !schemaMigrationChangeExists(result.Changes, expected.changeType, "safe", expected.entity, expected.field) {
			t.Fatalf("expected change %s target=%s.%s in %#v", expected.changeType, expected.entity, expected.field, result.Changes)
		}
	}
	for _, forbidden := range []string{"create-table", "drop-table", "add-column", "drop-column", "add-index", "drop-index", "change-column-type", "change-relation-target"} {
		if schemaMigrationChangeTypeExists(result.Changes, forbidden) {
			t.Fatalf("did not expect %s in explicit rename plan: %#v", forbidden, result.Changes)
		}
	}
	if result.Summary.Renames != 2 || result.Summary.DestructiveChanges != 0 || result.Summary.ManualChanges != 0 {
		t.Fatalf("unexpected summary counts: %#v", result.Summary)
	}
}

func TestMigratePlanRequiresOldAndNewFiles(t *testing.T) {
	result := MigratePlan([]string{"old.black", "--json"})
	if result.Success {
		t.Fatalf("expected missing-file diagnostic")
	}
	if len(result.Errors) != 1 || result.Errors[0].Code != "MISSING_SCHEMA_MIGRATION_FILES" {
		t.Fatalf("expected MISSING_SCHEMA_MIGRATION_FILES, got %#v", result.Errors)
	}
}

func TestFormatSchemaMigrationPlanIR(t *testing.T) {
	result := SchemaMigrationPlanResult{
		Success:     true,
		Command:     "migrate plan",
		Version:     version,
		OldFile:     "old.black",
		NewFile:     "new.black",
		Safe:        false,
		Destructive: true,
		Changes: []SchemaMigrationChange{{
			Type:    "drop-column",
			Risk:    "destructive",
			Entity:  "Product",
			Field:   "stock",
			Column:  "stock",
			OldType: "integer",
		}},
		Steps:  []SchemaMigrationStep{{Order: 1, Action: "Back up data.", Reason: "Destructive change."}},
		Errors: []Diagnostic{},
	}

	ir := FormatSchemaMigrationPlanIR(result)
	for _, expected := range []string{
		"migrate plan ok",
		"safe false",
		"destructive true",
		"drop-column risk destructive target Product.stock",
		"steps",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected IR to contain %q, got:\n%s", expected, ir)
		}
	}
}

func writeMigrationTestFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write migration test file: %v", err)
	}
	return path
}

func schemaMigrationChangeExists(changes []SchemaMigrationChange, changeType string, risk string, entity string, field string) bool {
	for _, change := range changes {
		if change.Type == changeType && change.Risk == risk && change.Entity == entity && change.Field == field {
			return true
		}
	}
	return false
}

func schemaMigrationChangeTypeExists(changes []SchemaMigrationChange, changeType string) bool {
	for _, change := range changes {
		if change.Type == changeType {
			return true
		}
	}
	return false
}
