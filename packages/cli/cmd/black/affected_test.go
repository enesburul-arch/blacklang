package main

import (
	"strings"
	"testing"
)

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
	if !affectedItemsContain(analysis.Queries, "OpenOrders") {
		t.Fatalf("expected OpenOrders query, got %#v", analysis.Queries)
	}
	if !affectedItemsContain(analysis.Jobs, "OrderStatusSweep") {
		t.Fatalf("expected OrderStatusSweep job, got %#v", analysis.Jobs)
	}
	if !affectedItemsContain(analysis.GeneratedFiles, "src/pages/OrdersPage.tsx") {
		t.Fatalf("expected generated Orders page, got %#v", analysis.GeneratedFiles)
	}
	if !affectedItemsContain(analysis.GeneratedFiles, "src/worker.ts") {
		t.Fatalf("expected generated worker, got %#v", analysis.GeneratedFiles)
	}
}

func TestAnalyzeAffectedRelationField(t *testing.T) {
	program := parseAffectedTestProgram(t)

	analysis, diagnostics := AnalyzeAffected(program, "Order.customer")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "relation-field" || analysis.Entity != "Order" || analysis.Field != "customer" {
		t.Fatalf("unexpected relation affected identity: %#v", analysis)
	}
	for _, file := range []string{"src/routes/order.ts", "src/pages/OrdersPage.tsx", "openapi.json"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected relation field to affect %s, got %#v", file, analysis.GeneratedFiles)
		}
	}
	if !affectedStringsContain(analysis.AgentNotes, "API response include policy") {
		t.Fatalf("expected relation load note, got %#v", analysis.AgentNotes)
	}
}

func TestAnalyzeAffectedComponentSection(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text
  stock number
}

component StockBadge {
  input stock number
}

page Products {
  source Product

  view {
    order table, StockSummary, StockCards, detail, form
    section StockSummary component StockBadge bind selected
    section StockCards component StockBadge bind each
  }

  table {
    columns name
  }
}
`
	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}

	fieldAnalysis, diagnostics := AnalyzeAffected(program, "Product.stock")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no field diagnostics, got %#v", diagnostics)
	}
	if !affectedItemsContain(fieldAnalysis.Pages, "Products") {
		t.Fatalf("expected component section input to affect Products page, got %#v", fieldAnalysis.Pages)
	}

	componentAnalysis, diagnostics := AnalyzeAffected(program, "StockBadge")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no component diagnostics, got %#v", diagnostics)
	}
	if !affectedItemsContain(componentAnalysis.Pages, "Products") || !affectedItemsContain(componentAnalysis.GeneratedFiles, "src/pages/ProductsPage.tsx") {
		t.Fatalf("expected component to affect Products page output, got %#v", componentAnalysis)
	}

	sectionAnalysis, diagnostics := AnalyzeAffected(program, "StockCards")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no view section diagnostics, got %#v", diagnostics)
	}
	if sectionAnalysis.Kind != "view-section" || sectionAnalysis.Entity != "Product" || !affectedItemsContain(sectionAnalysis.Components, "StockBadge") || !affectedItemsContain(sectionAnalysis.GeneratedFiles, "src/pages/ProductsPage.tsx") {
		t.Fatalf("expected view section to affect page and component output, got %#v", sectionAnalysis)
	}

	qualifiedSectionAnalysis, diagnostics := AnalyzeAffected(program, "Products.StockCards")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no qualified view section diagnostics, got %#v", diagnostics)
	}
	if qualifiedSectionAnalysis.Kind != "view-section" || !affectedItemsContain(qualifiedSectionAnalysis.Pages, "Products") {
		t.Fatalf("expected qualified view section lookup, got %#v", qualifiedSectionAnalysis)
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

func TestAnalyzeAffectedEntityIndex(t *testing.T) {
	program := parseAffectedTestProgram(t)

	analysis, diagnostics := AnalyzeAffected(program, "Order.index")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "entity-index" || analysis.Entity != "Order" {
		t.Fatalf("unexpected index affected identity: %#v", analysis)
	}
	for _, file := range []string{"prisma/schema.prisma", "src/setup-db.ts"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}
	if affectedItemsContain(analysis.GeneratedFiles, "src/pages/OrdersPage.tsx") {
		t.Fatalf("did not expect page output for index-only change, got %#v", analysis.GeneratedFiles)
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

func TestAnalyzeAffectedJob(t *testing.T) {
	program := parseAffectedTestProgram(t)

	analysis, diagnostics := AnalyzeAffected(program, "OrderStatusSweep")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "job" {
		t.Fatalf("expected job affected analysis, got %#v", analysis)
	}
	if !affectedItemsContain(analysis.Jobs, "OrderStatusSweep") || !affectedItemsContain(analysis.Queries, "OpenOrders") || !affectedItemsContain(analysis.Entities, "Order") {
		t.Fatalf("expected job/query/entity impact, got %#v", analysis)
	}
	for _, file := range []string{"jobs/manifest.json", "src/worker.ts", "package.json", "openapi.json"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}

	query, diagnostics := AnalyzeAffected(program, "OpenOrders")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no query diagnostics, got %#v", diagnostics)
	}
	if !affectedItemsContain(query.Jobs, "OrderStatusSweep") {
		t.Fatalf("expected query analysis to include job, got %#v", query.Jobs)
	}
}

func TestAnalyzeAffectedDeploy(t *testing.T) {
	source := `app Warehouse

deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  preview local
  rollback keep 3
  cloud fly app env FLY_APP_NAME region env FLY_REGION
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
	for _, file := range []string{".env.example", ".dockerignore", "Dockerfile", "docker-compose.yml", "docker-compose.preview.yml", "deploy/manifest.json", "deploy/rollback.json", "deploy/cloud.json", "scripts/rollback-plan.mjs", "scripts/cloud-plan.mjs", "scripts/cloud-exec.mjs", "package.json", "src/server.ts"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}
	cloud, diagnostics := AnalyzeAffected(program, "cloud")
	if len(diagnostics) != 0 || !cloud.Found || cloud.Kind != "deploy" {
		t.Fatalf("expected cloud to map to deploy affected analysis, got %#v diagnostics %#v", cloud, diagnostics)
	}
}

func TestAnalyzeAffectedOps(t *testing.T) {
	source := `app Warehouse

deploy {
  target docker
}

ops {
  health path "/healthz"
  readiness path "/readyz"
  metrics path "/metrics"
  logging requests
  observe webhook endpoint env BLACKLANG_OBSERVABILITY_ENDPOINT
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}

	analysis, diagnostics := AnalyzeAffected(program, "ops")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "ops" {
		t.Fatalf("expected ops affected analysis, got %#v", analysis)
	}
	for _, file := range []string{"src/server.ts", "openapi.json", "ops/observability.json", "src/blacklang.contract.test.ts", "src/blacklang.api.test.ts", "Dockerfile", "docker-compose.yml"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}

	health, diagnostics := AnalyzeAffected(program, "health")
	if len(diagnostics) != 0 || !health.Found || health.Kind != "ops" {
		t.Fatalf("expected health to map to ops affected analysis, got %#v diagnostics %#v", health, diagnostics)
	}
	observe, diagnostics := AnalyzeAffected(program, "observe")
	if len(diagnostics) != 0 || !observe.Found || observe.Kind != "ops" {
		t.Fatalf("expected observe to map to ops affected analysis, got %#v diagnostics %#v", observe, diagnostics)
	}
}

func TestAnalyzeAffectedExplicitAPI(t *testing.T) {
	source := `app Warehouse

api LowStockReport {
  method GET
  path "/api/reports/low-stock/{warehouseId}"
  param warehouseId text
  query limit integer
  private
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

	analysis, diagnostics := AnalyzeAffected(program, "LowStockReport")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "api" {
		t.Fatalf("expected explicit API affected analysis, got %#v", analysis)
	}
	for _, file := range []string{"src/server.ts", "openapi.json", "src/blacklang.contract.test.ts", "src/blacklang.api.test.ts"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}
}

func TestAnalyzeAffectedI18N(t *testing.T) {
	source := `app Warehouse

i18n {
  default tr
  locales tr, en
}

label Product.name {
  tr "Ürün Adı"
  en "Product Name"
}

placeholder Product.name {
  tr "Ürün adını gir"
  en "Enter product name"
}

entity Product {
  name text required
}

page Products {
  source Product

  table {
    columns name
  }

  form {
    fields name
  }

  actions create
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

	analysis, diagnostics := AnalyzeAffected(program, "i18n")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "i18n" {
		t.Fatalf("expected i18n affected analysis, got %#v", analysis)
	}
	if !affectedItemsContain(analysis.Entities, "Product") {
		t.Fatalf("expected Product entity, got %#v", analysis.Entities)
	}
	for _, file := range []string{"src/App.tsx", "src/pages/ProductsPage.tsx"} {
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

func TestAnalyzeAffectedAPIOnlyTarget(t *testing.T) {
	source := `app InventoryAPI

target api {
  backend node
  database sqlite
}

deploy {
  target docker
}

entity Product {
  sku text required
}

page Products {
  source Product
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
	for _, file := range []string{"README.md", "package.json", "prisma.config.ts", "prisma/schema.prisma", "src/server.ts", "src/blacklang.contract.test.ts", "src/blacklang.api.test.ts", "Dockerfile", "docker-compose.yml"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}
	for _, file := range []string{"index.html", "vite.config.ts", "src/main.tsx", "src/App.tsx", "src/styles.css", "src/vite-env.d.ts", "src/blacklang.frontend.test.tsx"} {
		if affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected API-only target affected analysis to omit %s, got %#v", file, analysis.GeneratedFiles)
		}
	}
}

func TestAnalyzeAffectedAPIOnlyPageOmitsFrontendFiles(t *testing.T) {
	source := `app InventoryAPI

target api {
  backend node
  database sqlite
}

entity Product {
  sku text required
}

page Products {
  source Product
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

	analysis, diagnostics := AnalyzeAffected(program, "Products")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "page" {
		t.Fatalf("expected page affected analysis, got %#v", analysis)
	}
	for _, file := range []string{"src/api/product.ts", "src/routes/product.ts", "src/validation/product.ts", "src/server.ts", "openapi.json"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}
	for _, file := range []string{"src/pages/ProductsPage.tsx", "src/App.tsx"} {
		if affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected API-only page affected analysis to omit %s, got %#v", file, analysis.GeneratedFiles)
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

func TestAnalyzeAffectedMigration(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
  name text
}

migration RenameProductName {
  rename field Product.title to name
}

page Products {
  source Product

  table {
    columns sku, name
  }

  form {
    fields sku, name
  }
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

	analysis, diagnostics := AnalyzeAffected(program, "RenameProductName")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "migration" {
		t.Fatalf("expected migration affected analysis, got %#v", analysis)
	}
	if !affectedItemsContain(analysis.Migrations, "RenameProductName") || !affectedItemsContain(analysis.Entities, "Product") {
		t.Fatalf("expected migration and Product entity, got %#v", analysis)
	}
	for _, file := range []string{"migrations/manifest.json", "migrations/*.sql", "src/migrate.ts", "src/setup-db.ts", "prisma/schema.prisma", "package.json"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}

	entityAnalysis, diagnostics := AnalyzeAffected(program, "Product.name")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no field diagnostics, got %#v", diagnostics)
	}
	if !affectedItemsContain(entityAnalysis.Migrations, "RenameProductName") {
		t.Fatalf("expected field analysis to include migration, got %#v", entityAnalysis.Migrations)
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
  index customer, status
}

role Worker {
  allow read Order status
}

query OpenOrders {
  source Order
  where status == "draft"
  sort status asc
  limit 10
}

job OrderStatusSweep {
  schedule every 1 hours
  run query OpenOrders
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

func affectedStringsContain(items []string, needle string) bool {
	for _, item := range items {
		if strings.Contains(item, needle) {
			return true
		}
	}
	return false
}
