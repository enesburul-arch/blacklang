package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSeedDeclaration(t *testing.T) {
	source := `app Warehouse

entity Customer {
  name text required unique
}

entity Order {
  customer Customer required
  total money default 0
}

seed DemoCustomers {
  source Customer

  row DemoCustomerAcme {
    name "Acme Logistics"
  }
}

seed DemoOrders {
  source Order

  row DemoOrderDraft {
    customer ref DemoCustomerAcme
    total 125.50
  }
}
`

	program, diagnostics := Parse("seed.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	if len(program.Seeds) != 2 {
		t.Fatalf("expected two seeds, got %#v", program.Seeds)
	}
	if program.Seeds[0].Name != "DemoCustomers" || program.Seeds[0].Source != "Customer" || len(program.Seeds[0].Rows) != 1 {
		t.Fatalf("unexpected first seed: %#v", program.Seeds[0])
	}
	if got := program.Seeds[1].Rows[0].Values[0].Value; got.Kind != "ref" || got.Value != "DemoCustomerAcme" {
		t.Fatalf("expected relation ref, got %#v", got)
	}
}

func TestValidateSeedDeclaration(t *testing.T) {
	source := `app Warehouse

entity Customer {
  name text required unique length 3..40
}

entity Product {
  sku text required unique regex "^[A-Z0-9-]+$"
  photo image optional accept "image/*"
  manual file optional accept "application/pdf"
  stock number required min 0
  active boolean default true
}

entity Order {
  customer Customer required
  total money default 0
}

seed DemoCustomers {
  source Customer

  row DemoCustomerAcme {
    name "Acme Logistics"
  }
}

seed DemoProducts {
  source Product

  row DemoProductLow {
    sku "LOW-001"
    photo "data:image/png;base64,AA=="
    manual "data:application/pdf;base64,AA=="
    stock 3
    active true
  }
}

seed DemoOrders {
  source Order

  row DemoOrderDraft {
    customer ref DemoCustomerAcme
    total 125.50
  }
}
`

	program, diagnostics := Parse("seed.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validate diagnostics, got %#v", diagnostics)
	}
}

func TestValidateSeedErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required unique regex "^[A-Z0-9-]+$"
  photo image optional accept "image/*"
  stock number required min 1
  computed inventoryValue money = stock * stock
}

entity Order {
  product Product required
  status text required length 3..8
}

seed Product {
  source Missing

  row DemoMissing {
    sku "MISS"
  }
}

seed DemoProducts {
  source Product
  source Product

  row DemoDuplicate {
    sku "DUP"
    sku "DUP"
    stock 0
    inventoryValue 99
    missingField true
  }

  row DemoDuplicate {
    sku "DUP"
    stock "many"
  }

  row DemoNoSku {
    photo "not-an-image"
    stock 2
  }

  row DemoScalarRef {
    sku ref DemoDuplicate
    stock 2
  }

  row 123Bad {
    sku "bad"
    stock 2
  }
}

seed DemoOrders {
  source Order

  row DemoOrderLiteralRelation {
    product "DemoDuplicate"
    status "ok"
  }

  row DemoOrderUnknownRef {
    product ref MissingProduct
    status "draft"
  }
}
`

	program, parseDiagnostics := Parse("seed.black", source)
	if !hasDiagnostic(parseDiagnostics, "DUPLICATE_SEED_SOURCE") {
		t.Fatalf("expected duplicate source parse diagnostic, got %#v", parseDiagnostics)
	}
	diagnostics := append(parseDiagnostics, Validate(program)...)
	for _, code := range []string{
		"SEED_NAME_COLLISION",
		"UNKNOWN_SEED_SOURCE",
		"DUPLICATE_SEED_SOURCE",
		"DUPLICATE_SEED_ROW",
		"DUPLICATE_SEED_FIELD",
		"SEED_VALUE_CONSTRAINT_MISMATCH",
		"UNSUPPORTED_SEED_FIELD",
		"UNKNOWN_SEED_FIELD",
		"DUPLICATE_SEED_UNIQUE_VALUE",
		"SEED_VALUE_TYPE_MISMATCH",
		"MISSING_SEED_FIELD",
		"UNSUPPORTED_SEED_REF",
		"INVALID_SEED_ROW",
		"SEED_RELATION_REQUIRES_REF",
		"UNKNOWN_SEED_REF",
	} {
		if !hasDiagnostic(diagnostics, code) {
			t.Fatalf("expected %s, got %#v", code, diagnostics)
		}
	}
}

func TestBuildWebGeneratesSeedRuntime(t *testing.T) {
	source := `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity Customer {
  name text required unique
}

entity Order {
  customer Customer required
  total money default 0
}

seed DemoCustomers {
  source Customer

  row DemoCustomerAcme {
    name "Acme Logistics"
  }
}

seed DemoOrders {
  source Order

  row DemoOrderDraft {
    customer ref DemoCustomerAcme
    total 125.50
  }
}

page Customers {
  source Customer

  table {
    columns name
  }

  form {
    fields name
  }

  actions create
}
`

	program, diagnostics := Parse("seed.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("parse seed fixture: %#v", diagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("validate seed fixture: %#v", diagnostics)
	}
	out := t.TempDir()
	if _, diagnostics := BuildWeb(program, out); len(diagnostics) != 0 {
		t.Fatalf("generate seed fixture: %#v", diagnostics)
	}

	packageJSON, err := os.ReadFile(filepath.Join(out, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	seedTS, err := os.ReadFile(filepath.Join(out, "src", "seed.ts"))
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		`"db:setup": "tsx src/setup-db.ts && npm run db:seed"`,
		`"db:seed": "tsx src/seed.ts"`,
	} {
		if !strings.Contains(string(packageJSON), expected) {
			t.Fatalf("expected package.json to contain %q:\n%s", expected, packageJSON)
		}
	}
	for _, expected := range []string{
		`"seed": "DemoCustomers"`,
		`"table": "Order"`,
		`"columns": [`,
		`"customerId"`,
		`"DemoCustomerAcme"`,
		`ON CONFLICT(\"id\") DO UPDATE SET`,
		`BlackLang seed data applied`,
	} {
		if !strings.Contains(string(seedTS), expected) {
			t.Fatalf("expected seed.ts to contain %q:\n%s", expected, seedTS)
		}
	}
}
