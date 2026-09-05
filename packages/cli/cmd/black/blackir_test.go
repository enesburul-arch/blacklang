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
}

entity Product {
  sku text required unique
  stock number default 0 ui text "#172026" 14 semibold left
  status text default draft
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
    order form, table, detail
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
		"entity Product",
		"sku text required unique",
		"stock number default 0 ui text #172026 14 semibold left",
		"status text default draft",
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
		"view-order form table detail",
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
