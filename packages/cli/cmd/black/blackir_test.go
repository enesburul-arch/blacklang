package main

import (
	"strings"
	"testing"
)

func TestFormatBlackIR(t *testing.T) {
	source := `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

auth {
  strategy emailPassword
  session cookie

  user {
    email email required unique
  }
}

database {
  url env DATABASE_URL
}

deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  preview local
  rollback keep 3
  cloud fly app env FLY_APP_NAME region env FLY_REGION
}

ops {
  health path "/healthz"
  readiness path "/readyz"
  metrics path "/metrics"
  logging requests
  observe webhook endpoint env BLACKLANG_OBSERVABILITY_ENDPOINT
}

entity Product {
  sku text required unique
  customer Customer required load detail query label "Customer"
  stock number default 0 ui text "#172026" 14 semibold left
  price money default 0
  computed inventoryValue money = stock * price label "Inventory Value"
  status text default draft
  index stock
}

role Admin {
  allow all
}

workflow OrderPreparation {
  source Product
  states draft, picking, shipped

  transition ship {
    from picking
    to shipped
    allow Admin
  }
}

state ProductPageState {
  selectedProducts Product[]
  activeFilter text
  modal createProduct closed
}

component StockBadge {
  input stock number

  variant low when stock < 10
  variant normal when stock >= 10
}

layout AdminLayout {
  sidebar {
    item Products
  }
}

page Products {
  layout AdminLayout
  source Product
  access Admin
  view {
    order form, StockSummary, StockCards, table, detail
    section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
    section StockCards component StockBadge bind each span 1 title "Stock Cards"
    group Record sections detail, form compose stack gap md title "Record Workspace"
    trigger detail on rowSelect
    trigger form on editStart
  }
  table {
    columns sku, stock
    search sku
    filter stock
    sort stock desc
    paginate 25
    ui table border 1 solid compact true
  }
  form {
    fields sku, stock
    ui box black 1 solid 8 8 5 5 6 center
  }
  actions create, edit, delete
  action create ui button primary white 6 md solid
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}

	ir := FormatBlackIR(program)
	expected := []string{
		"blackir 0.1",
		"app Warehouse",
		"target web frontend react backend node database sqlite",
		"auth strategy emailPassword session cookie",
		"user",
		"email email required unique",
		"database",
		"url env DATABASE_URL",
		"deploy target docker",
		"port env PORT default 3001",
		"env DATABASE_URL required",
		"preview local",
		"rollback keep 3",
		"cloud fly app env FLY_APP_NAME region env FLY_REGION",
		"ops",
		"health path \"/healthz\"",
		"readiness path \"/readyz\"",
		"metrics path \"/metrics\"",
		"logging requests",
		"observe webhook endpoint env BLACKLANG_OBSERVABILITY_ENDPOINT",
		"entity Product",
		"sku text required unique",
		"customer Customer required load detail query label Customer",
		"stock number default 0 ui text #172026 14 semibold left",
		"price money default 0",
		"computed inventoryValue money = stock * price label Inventory Value",
		"status text default draft",
		"index stock",
		"role Admin",
		"allow all",
		"workflow OrderPreparation source Product",
		"states draft picking shipped",
		"transition ship from picking to shipped allow Admin",
		"state ProductPageState",
		"selectedProducts Product[]",
		"activeFilter text",
		"modal createProduct closed",
		"component StockBadge",
		"input stock number",
		"variant low when stock < 10",
		"variant normal when stock >= 10",
		"layout AdminLayout",
		"sidebar Products",
		"page Products layout AdminLayout source Product",
		"view-order form StockSummary StockCards table detail",
		"view-section StockSummary component StockBadge bind selected span 1 title \"Stock Summary\"",
		"view-section StockCards component StockBadge bind each span 1 title \"Stock Cards\"",
		"view-group Record sections detail form compose stack gap md title \"Record Workspace\"",
		"view-trigger detail on rowSelect",
		"view-trigger form on editStart",
		"table sku stock",
		"filter stock",
		"sort stock desc",
		"paginate 25",
		"table-ui table border 1 solid compact true",
		"form-ui box black 1 solid 8 8 5 5 6 center",
		"actions create edit delete",
		"action-ui create button primary white 6 md solid",
		"access Admin",
	}
	for _, value := range expected {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected IR to contain %q, got:\n%s", value, ir)
		}
	}
}

func TestFormatBlackIRIncludesMigrations(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
  name text
}

migration RenameProductName {
  rename entity ProductItem to Product
  rename field Product.title to name
}
`
	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}

	ir := FormatBlackIR(program)
	for _, expected := range []string{
		"migration RenameProductName",
		"rename entity ProductItem to Product",
		"rename field Product.title to name",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected IR to contain %q, got:\n%s", expected, ir)
		}
	}
}
