package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestQueryGenerationIsDeterministic(t *testing.T) {
	g, firstDir := queryGeneratorFixture(t)
	secondDir := t.TempDir()
	files, diagnostics := BuildWeb(g.program, secondDir)
	if len(diagnostics) != 0 {
		t.Fatalf("second build failed: %#v", diagnostics)
	}
	for _, file := range files {
		relative, err := filepath.Rel(secondDir, file.Path)
		if err != nil {
			t.Fatal(err)
		}
		first, err := os.ReadFile(filepath.Join(firstDir, relative))
		if err != nil {
			t.Fatal(err)
		}
		second, err := os.ReadFile(file.Path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first, second) {
			t.Fatalf("repeated query generation changed %s", relative)
		}
	}
}

func TestQueryLearningContract(t *testing.T) {
	doc, ok := FindDoc("query")
	if !ok {
		t.Fatal("query documentation missing")
	}
	if strings.Contains(doc.Syntax, ";") {
		t.Fatal("query syntax must show newlines, not unsupported semicolon separators")
	}
	program, diagnostics := Parse("query-doc.black", doc.Example)
	if len(diagnostics) != 0 {
		t.Fatalf("query learning example must parse: %#v", diagnostics)
	}
	if diagnostics = Validate(program); len(diagnostics) != 0 {
		t.Fatalf("query learning example must validate: %#v", diagnostics)
	}
	explanation := ExplainKeyword("query")
	if !explanation.Success || len(explanation.AgentSteps) < 6 || !containsString(explanation.Related, "inspect") || len(explanation.ErrorCodes) == 0 {
		t.Fatalf("incomplete query explanation: %#v", explanation)
	}
	for _, note := range []string{"403", "1..1000", "row authorization", "AND", "relation selectors"} {
		if !strings.Contains(strings.Join(doc.AgentNotes, " "), note) {
			t.Fatalf("query docs must explain %q", note)
		}
	}
}

func TestQueryInspectAndAffected(t *testing.T) {
	program := queryToolingFixture(t)
	analysis, diagnostics := AnalyzeAffected(program, "LowStockProducts")
	if len(diagnostics) != 0 || analysis.Kind != "query" || !analysis.Found || analysis.Entity != "Product" {
		t.Fatalf("query identity missing: %#v %#v", analysis, diagnostics)
	}
	if !affectedItemsContain(analysis.Queries, "LowStockProducts") || !affectedItemsContain(analysis.Pages, "LowStock") || affectedItemsContain(analysis.Pages, "Products") {
		t.Fatalf("query should affect its bound page: %#v", analysis)
	}
	for _, path := range []string{"src/pages/LowStockPage.tsx", "src/routes/product.lowstock.ts", "src/api/product.lowstock.ts", "openapi.json"} {
		if !affectedItemsContain(analysis.GeneratedFiles, path) {
			t.Fatalf("missing query generated file %s: %#v", path, analysis.GeneratedFiles)
		}
	}
	for _, path := range []string{"prisma/schema.prisma", "src/setup-db.ts", "src/validation/product.ts"} {
		if affectedItemsContain(analysis.GeneratedFiles, path) {
			t.Fatalf("list query must not affect database/mutation schema: %s", path)
		}
	}
	for _, symbol := range []string{"Product", "Product.stock", "LowStock", "Reader", "auth"} {
		result, errors := AnalyzeAffected(program, symbol)
		if len(errors) != 0 || !affectedItemsContain(result.Queries, "LowStockProducts") {
			t.Fatalf("%s must identify dependent query: %#v %#v", symbol, result, errors)
		}
	}
	computed, errors := AnalyzeAffected(program, "Product.inventoryValue")
	if len(errors) != 0 || len(computed.Queries) != 0 {
		t.Fatalf("computed fields cannot be query dependencies: %#v %#v", computed, errors)
	}
	page, errors := AnalyzeAffected(program, "LowStock")
	if len(errors) != 0 || !affectedItemsContain(page.Pages, "Products") || !affectedItemsContain(page.GeneratedFiles, "src/pages/ProductsPage.tsx") || !affectedItemsContain(page.GeneratedFiles, "src/api/product.ts") {
		t.Fatalf("query binding changes must expose sibling module selection effects: %#v %#v", page, errors)
	}
	unused, errors := AnalyzeAffected(program, "AllProducts")
	if len(errors) != 0 || len(unused.GeneratedFiles) != 0 {
		t.Fatalf("unbound query should not claim runtime output: %#v %#v", unused, errors)
	}
	again, _ := AnalyzeAffected(program, "LowStockProducts")
	if !reflect.DeepEqual(analysis, again) {
		t.Fatal("affected output must be deterministic")
	}
	encoded, err := json.Marshal(InspectResult{Success: true, Program: program})
	if err != nil || !strings.Contains(string(encoded), `"queries":[`) || !strings.Contains(string(encoded), `"query":"LowStockProducts"`) {
		t.Fatalf("inspect JSON must retain declarations and bindings: %s %v", encoded, err)
	}
	ir := FormatAffectedIR(InspectAffectedResult{Success: true, Affected: analysis})
	if !strings.Contains(ir, "queries 1") || !strings.Contains(ir, "kind query") {
		t.Fatalf("affected IR missing query: %s", ir)
	}
	inspectIR := FormatInspectIR(InspectResult{Success: true, Program: program})
	if !strings.Contains(inspectIR, "queries 2") || !strings.Contains(inspectIR, "query LowStockProducts source Product filters 2") {
		t.Fatalf("inspect IR missing query: %s", inspectIR)
	}
}

func TestQueryBlackIRRetainsLiteralKinds(t *testing.T) {
	program := queryToolingFixture(t)
	ir := FormatBlackIR(program)
	for _, expected := range []string{
		"query LowStockProducts source Product\n",
		"  where stock < 10\n",
		`  where name != "true, {stock} # literal"`,
		"  sort stock asc\n  limit 50\n",
		"  query LowStockProducts\n",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("query IR missing %q: %s", expected, ir)
		}
	}
}

func queryToolingFixture(t *testing.T) Program {
	t.Helper()
	source := `app Inventory
auth {
  strategy emailPassword
  session cookie
  user {
    email email required unique
  }
}
entity Product {
  name text required
  stock number default 0
  price money default 0
  computed inventoryValue money = stock * price
}
query LowStockProducts {
  source Product
  where stock < 10
  where name != "true, {stock} # literal"
  sort stock asc
  limit 50
}
query AllProducts {
  source Product
}
role Reader {
  allow read Product
}
page Products {
  source Product
  table { columns name, stock }
}
page LowStock {
  source Product
  query LowStockProducts
  access Reader
  table { columns name, stock }
}
`
	program, diagnostics := Parse("query-tooling.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("invalid tooling fixture: %#v", diagnostics)
	}
	if diagnostics = Validate(program); len(diagnostics) != 0 {
		t.Fatalf("invalid tooling fixture: %#v", diagnostics)
	}
	return program
}
