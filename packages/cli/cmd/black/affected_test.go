package main

import "testing"

func TestAnalyzeAffectedField(t *testing.T) {
	program := parseAffectedTestProgram(t)

	analysis, diagnostics := AnalyzeAffected(program, "Order.status")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "field" || analysis.Entity != "Order" || analysis.Field != "status" {
		t.Fatalf("unexpected affected identity: %#v", analysis)
	}
	if !affectedItemsContain(analysis.Pages, "Orders") {
		t.Fatalf("expected Orders page, got %#v", analysis.Pages)
	}
	if !affectedItemsContain(analysis.Workflows, "OrderLifecycle") {
		t.Fatalf("expected OrderLifecycle workflow, got %#v", analysis.Workflows)
	}
	if !affectedItemsContain(analysis.Roles, "Worker") {
		t.Fatalf("expected Worker role, got %#v", analysis.Roles)
	}
	if !affectedItemsContain(analysis.Components, "StatusBadge") {
		t.Fatalf("expected StatusBadge component, got %#v", analysis.Components)
	}
	if !affectedItemsContain(analysis.GeneratedFiles, "src/pages/OrdersPage.tsx") {
		t.Fatalf("expected generated Orders page, got %#v", analysis.GeneratedFiles)
	}
}

func TestAnalyzeAffectedEntityIncludesRelationPages(t *testing.T) {
	program := parseAffectedTestProgram(t)

	analysis, diagnostics := AnalyzeAffected(program, "Customer")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "entity" || analysis.Entity != "Customer" {
		t.Fatalf("unexpected affected identity: %#v", analysis)
	}
	if !affectedItemsContain(analysis.Entities, "Order") {
		t.Fatalf("expected Order relation entity, got %#v", analysis.Entities)
	}
	if !affectedItemsContain(analysis.Pages, "Orders") {
		t.Fatalf("expected Orders page through relation display, got %#v", analysis.Pages)
	}
}

func TestAnalyzeAffectedUnknownSymbol(t *testing.T) {
	program := parseAffectedTestProgram(t)

	analysis, diagnostics := AnalyzeAffected(program, "Missing")
	if len(diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_AFFECTED_SYMBOL" {
		t.Fatalf("expected UNKNOWN_AFFECTED_SYMBOL, got %q", diagnostics[0].Code)
	}
	if analysis.Found {
		t.Fatalf("expected missing symbol to be unfound")
	}
}

func TestAnalyzeAffectedDeploy(t *testing.T) {
	source := `app Warehouse

deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}

	analysis, diagnostics := AnalyzeAffected(program, "deploy")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "deploy" {
		t.Fatalf("expected deploy affected analysis, got %#v", analysis)
	}
	for _, file := range []string{".env.example", ".dockerignore", "Dockerfile", "docker-compose.yml", "package.json", "src/server.ts"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}
}

func TestAnalyzeAffectedTarget(t *testing.T) {
	source := `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

deploy {
  target docker
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	diagnostics = Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}

	analysis, diagnostics := AnalyzeAffected(program, "target")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "target" {
		t.Fatalf("expected target affected analysis, got %#v", analysis)
	}
	for _, file := range []string{"README.md", "package.json", "vite.config.ts", "prisma.config.ts", "prisma/schema.prisma", "src/server.ts", "Dockerfile", "docker-compose.yml"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}
}

func TestFirstNonOptionArgSkipsAffectedValue(t *testing.T) {
	file := firstNonOptionArg([]string{"--affected", "Product.stock", "--json"})
	if file != "" {
		t.Fatalf("expected no file arg, got %q", file)
	}

	file = firstNonOptionArg([]string{"examples/warehouse/app.black", "--affected", "Product.stock", "--json"})
	if file != "examples/warehouse/app.black" {
		t.Fatalf("expected explicit file arg, got %q", file)
	}
}

func parseAffectedTestProgram(t *testing.T) Program {
	t.Helper()
	source := `app Warehouse

auth {
  strategy emailPassword
  session cookie

  user {
    email email required unique
  }
}

entity Customer {
  name text required
}

entity Order {
  customer Customer required
  status text default draft
  total money default 0
}

role Worker {
  allow read Order status
}

workflow OrderLifecycle {
  source Order
  states draft, shipped

  transition ship {
    from draft
    to shipped
    allow Worker
  }
}

state OrdersPageState {
  selectedOrders Order[]
  modal createOrder closed
}

component StatusBadge {
  input status text

  variant draft when status == draft
}

page Customers {
  source Customer

  table {
    columns name
    search name
  }

  form {
    fields name
  }

  actions create, edit
}

page Orders {
  source Order
  access Worker

  table {
    columns customer, status, total
    search customer, status
    filter status
    sort status asc
  }

  form {
    fields customer, status, total
  }

  actions create, edit, delete
}
`
	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	diagnostics = Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	return program
}

func affectedItemsContain(items []AffectedItem, name string) bool {
	for _, item := range items {
		if item.Name == name {
			return true
		}
	}
	return false
}
