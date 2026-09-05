package main

import (
	"strings"
	"testing"
)

func TestFormatBlackIR(t *testing.T) {
	source := `app Warehouse

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

entity Product {
  sku text required unique
  stock number default 0
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
  table {
    columns sku, stock
    search sku
    filter stock
    sort stock desc
    paginate 25
  }
  form {
    fields sku, stock
  }
  actions create, edit, delete
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
		"auth strategy emailPassword session cookie",
		"user",
		"email email required unique",
		"database",
		"url env DATABASE_URL",
		"entity Product",
		"sku text required unique",
		"stock number default 0",
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
		"table sku stock",
		"filter stock",
		"sort stock desc",
		"paginate 25",
		"actions create edit delete",
		"access Admin",
	}
	for _, value := range expected {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected IR to contain %q, got:\n%s", value, ir)
		}
	}
}
