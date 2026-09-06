package main

import (
	"reflect"
	"strings"
	"testing"
)

const queryTestEntities = `app Warehouse
entity Product {
  name text
  email email
  stock number
  units integer
  price money
  ratio decimal
  active boolean
  due date
  updated datetime
  customer Customer
  computed inventoryValue money = stock * price
}
entity Customer {
  name text
}
`

func TestParseQueryPreservesTypedLiteralsAndPositions(t *testing.T) {
	source := queryTestEntities + `query LowStockProducts {
  source Product
  where stock < 10
  where active == true
  where name != "false"
  sort stock asc
  limit 50
}
page Products {
  source Product
  query LowStockProducts
  table { columns name, stock }
}
`
	program, diagnostics := Parse("query.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("parse: %#v", diagnostics)
	}
	if diagnostics = Validate(program); len(diagnostics) != 0 {
		t.Fatalf("validate: %#v", diagnostics)
	}
	query, ok := findQuery(program, "LowStockProducts")
	if !ok || query.Source != "Product" || query.Limit != 50 || query.Sort.Field != "stock" || query.Sort.Direction != "asc" {
		t.Fatalf("query: %#v", query)
	}
	if len(query.Where) != 3 || query.Where[0].Value != (QueryLiteral{Kind: "number", Value: "10"}) || query.Where[1].Value != (QueryLiteral{Kind: "boolean", Value: "true"}) || query.Where[2].Value != (QueryLiteral{Kind: "string", Value: "false"}) {
		t.Fatalf("typed filters: %#v", query.Where)
	}
	if query.SourcePosition.Line == 0 || query.SortPosition.Line == 0 || query.LimitPosition.Line == 0 || query.Where[0].Position.Column != 9 || program.Pages[0].QueryPosition.Line == 0 {
		t.Fatalf("source positions missing: %#v %#v", query, program.Pages[0])
	}
	if !reflect.DeepEqual(queryFieldNames(query), []string{"stock", "active", "name"}) {
		t.Fatalf("field dependencies: %#v", queryFieldNames(query))
	}
}

func TestQueryLiteralDelimitersSurviveLexingAndFormatting(t *testing.T) {
	for _, value := range []string{"{", "}", ",", "source", "query", "true", "10", "a,b", "line\nquote\"slash\\", "}); process.exit(1); //"} {
		t.Run(value, func(t *testing.T) {
			source := queryTestEntities + "query TextMatches {\nsource Product\nwhere name == " + quoteBlackString(value) + "\n}\n"
			formatted, diagnostics := FormatBlackSource("query.black", source)
			if len(diagnostics) != 0 {
				t.Fatalf("format: %#v", diagnostics)
			}
			program, diagnostics := Parse("query.black", formatted)
			if len(diagnostics) != 0 || len(program.Queries) != 1 || len(program.Queries[0].Where) != 1 {
				t.Fatalf("parse formatted %s: %#v", formatted, diagnostics)
			}
			if got := program.Queries[0].Where[0].Value; got.Kind != "string" || got.Value != value {
				t.Fatalf("literal got %#v, want %q", got, value)
			}
			if diagnostics = Validate(program); len(diagnostics) != 0 {
				t.Fatalf("validate: %#v", diagnostics)
			}
			second, diagnostics := FormatBlackSource("query.black", formatted)
			if len(diagnostics) != 0 || second != formatted {
				t.Fatalf("format is not idempotent: %q -> %q", formatted, second)
			}
		})
	}
}

func TestQueryParseDiagnostics(t *testing.T) {
	cases := []struct{ name, source, code string }{
		{"header", "query LowStock\n", "INVALID_QUERY_DECLARATION"},
		{"quoted keyword", "\"query\" LowStock {\nsource Product\n}\n", "INVALID_QUERY_DECLARATION"},
		{"quoted name", "query \"LowStock\" {\nsource Product\n}\n", "INVALID_QUERY_DECLARATION"},
		{"source shape", "query LowStock {\nsource Product extra\n}\n", "INVALID_QUERY_SOURCE"},
		{"source quote", "query LowStock {\nsource \"Product\"\n}\n", "INVALID_QUERY_SOURCE"},
		{"source twice", "query LowStock {\nsource Product\nsource Product\n}\n", "DUPLICATE_QUERY_SOURCE"},
		{"where shape", "query LowStock {\nsource Product\nwhere stock < 10 or stock > 20\n}\n", "INVALID_QUERY_WHERE"},
		{"where field quote", "query LowStock {\nsource Product\nwhere \"stock\" < 10\n}\n", "INVALID_QUERY_WHERE"},
		{"bare string", "query LowStock {\nsource Product\nwhere name == active\n}\n", "INVALID_QUERY_LITERAL"},
		{"NaN", "query LowStock {\nsource Product\nwhere price == NaN\n}\n", "INVALID_QUERY_LITERAL"},
		{"infinity", "query LowStock {\nsource Product\nwhere price == Infinity\n}\n", "INVALID_QUERY_LITERAL"},
		{"numeric overflow", "query LowStock {\nsource Product\nwhere price == " + strings.Repeat("9", 400) + "\n}\n", "INVALID_QUERY_LITERAL"},
		{"exponent", "query LowStock {\nsource Product\nwhere price == 1e3\n}\n", "INVALID_QUERY_LITERAL"},
		{"boolean case", "query LowStock {\nsource Product\nwhere active == TRUE\n}\n", "INVALID_QUERY_LITERAL"},
		{"sort shape", "query LowStock {\nsource Product\nsort stock\n}\n", "INVALID_QUERY_SORT"},
		{"sort twice", "query LowStock {\nsource Product\nsort stock asc\nsort name desc\n}\n", "DUPLICATE_QUERY_SORT"},
		{"limit twice", "query LowStock {\nsource Product\nlimit 5\nlimit 10\n}\n", "DUPLICATE_QUERY_LIMIT"},
		{"unknown clause", "query LowStock {\nsource Product\njoin Customer\n}\n", "UNEXPECTED_QUERY_TOKEN"},
		{"quoted clause", "query LowStock {\n\"source\" Product\n}\n", "UNEXPECTED_QUERY_TOKEN"},
		{"unclosed", "query LowStock {\nsource Product\n", "UNCLOSED_QUERY"},
		{"page query shape", "page Products {\nsource Product\nquery LowStock extra\n}\n", "INVALID_PAGE_QUERY"},
		{"page query comma", "page Products {\nsource Product\nquery LowStock,\n}\n", "INVALID_PAGE_QUERY"},
		{"page query quoted keyword", "page Products {\nsource Product\n\"query\" LowStock\n}\n", "INVALID_PAGE_QUERY"},
		{"page query twice", "page Products {\nsource Product\nquery LowStock\nquery LowStock\n}\n", "DUPLICATE_PAGE_QUERY"},
	}
	for _, value := range []string{"0", "-1", "1001", "1.5", "\"50\"", "+50", "050", "99999999999999999999999"} {
		cases = append(cases, struct{ name, source, code string }{"limit " + value, "query LowStock {\nsource Product\nlimit " + value + "\n}\n", "INVALID_QUERY_LIMIT"})
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, diagnostics := Parse("query.black", queryTestEntities+test.source)
			assertQueryDiagnostic(t, diagnostics, test.code)
		})
	}
}

func TestQueryValidationDiagnostics(t *testing.T) {
	cases := []struct{ name, query, suffix, code string }{
		{"missing source", "", "", "MISSING_QUERY_SOURCE"},
		{"unknown source", "source Missing", "", "UNKNOWN_QUERY_SOURCE"},
		{"invalid source", "source Product.bad", "", "INVALID_QUERY_SOURCE"},
		{"unknown field", "source Product\nwhere missing == 1", "", "UNKNOWN_QUERY_FIELD"},
		{"system field", "source Product\nwhere id == \"x\"", "", "UNKNOWN_QUERY_FIELD"},
		{"computed field", "source Product\nwhere inventoryValue > 0", "", "UNSUPPORTED_QUERY_FIELD"},
		{"relation field", "source Product\nwhere customer == \"x\"", "", "UNSUPPORTED_QUERY_FIELD"},
		{"invalid field", "source Product\nwhere customer.name == \"x\"", "", "INVALID_QUERY_FIELD"},
		{"operator", "source Product\nwhere stock = 1", "", "UNSUPPORTED_QUERY_OPERATOR"},
		{"text order", "source Product\nwhere name > \"A\"", "", "UNSUPPORTED_QUERY_OPERATOR"},
		{"boolean order", "source Product\nwhere active < true", "", "UNSUPPORTED_QUERY_OPERATOR"},
		{"wrong literal", "source Product\nwhere stock == \"10\"", "", "QUERY_LITERAL_TYPE_MISMATCH"},
		{"duplicate condition", "source Product\nwhere stock < 10\nwhere stock < 10", "", "DUPLICATE_QUERY_WHERE"},
		{"sort direction", "source Product\nsort stock descending", "", "UNSUPPORTED_QUERY_SORT_DIRECTION"},
		{"sort computed", "source Product\nsort inventoryValue asc", "", "UNSUPPORTED_QUERY_FIELD"},
		{"sort relation", "source Product\nsort customer asc", "", "UNSUPPORTED_QUERY_FIELD"},
		{"sort unknown", "source Product\nsort missing asc", "", "UNKNOWN_QUERY_FIELD"},
		{"source mismatch", "source Product", "page Customers {\nsource Customer\nquery LowStock\n}\n", "PAGE_QUERY_SOURCE_MISMATCH"},
		{"unknown page query", "source Product", "page Products {\nsource Product\nquery Unknown\n}\n", "UNKNOWN_PAGE_QUERY"},
		{"duplicate name", "source Product", "query LowStock {\nsource Product\n}\n", "DUPLICATE_QUERY"},
		{"normalized duplicate", "source Product", "query Lowstock {\nsource Product\n}\n", "QUERY_NAME_COLLISION"},
		{"page collision", "source Product", "page LowStock {\nsource Product\n}\n", "QUERY_NAME_COLLISION"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			program, diagnostics := Parse("query.black", queryTestEntities+"query LowStock {\n"+test.query+"\n}\n"+test.suffix)
			if len(diagnostics) != 0 {
				t.Fatalf("parse: %#v", diagnostics)
			}
			assertQueryDiagnostic(t, Validate(program), test.code)
		})
	}
	for _, name := range []string{"Product", "Warehouse", "Auth", "Cors", "Target", "Database", "Security", "Deploy"} {
		t.Run("symbol "+name, func(t *testing.T) {
			program, diagnostics := Parse("query.black", queryTestEntities+"query "+name+" {\nsource Product\n}\n")
			if len(diagnostics) != 0 {
				t.Fatalf("parse: %#v", diagnostics)
			}
			assertQueryDiagnostic(t, Validate(program), "QUERY_NAME_COLLISION")
		})
	}
	for _, name := range []string{"lowStock", "Low-Stock", "Low.Stock", "../LowStock", "Low_Stock"} {
		t.Run("name "+name, func(t *testing.T) {
			program, diagnostics := Parse("query.black", queryTestEntities+"query "+name+" {\nsource Product\n}\n")
			if len(diagnostics) != 0 {
				t.Fatalf("parse: %#v", diagnostics)
			}
			assertQueryDiagnostic(t, Validate(program), "INVALID_QUERY_NAME")
		})
	}
}

func TestQueryLiteralTypeBoundaries(t *testing.T) {
	cases := []struct {
		fieldType, kind, value string
		valid                  bool
	}{
		{"number", "number", "2147483647", true}, {"integer", "number", "-2147483648", true},
		{"number", "number", "2147483648", false}, {"integer", "number", "-2147483649", false},
		{"number", "number", "1.0", true}, {"number", "number", "1.1", false},
		{"integer", "number", "2147483647.0000000000001", false},
		{"decimal", "number", "-0.125", true}, {"money", "number", "12.50", true},
		{"money", "string", "12.50", false}, {"money", "number", "NaN", false},
		{"money", "number", "1e3", false}, {"text", "string", "", true},
		{"email", "string", "person@example.com", true}, {"text", "boolean", "true", false},
		{"boolean", "boolean", "false", true}, {"boolean", "string", "true", false}, {"boolean", "boolean", "TRUE", false},
		{"date", "string", "2024-02-29", true}, {"date", "string", "2025-02-29", false},
		{"date", "string", "2025-2-01", false}, {"date", "string", "2025-01-01T00:00:00Z", false},
		{"datetime", "string", "2026-09-06T09:30:00Z", true}, {"datetime", "string", "2026-09-06T12:30:00.123+03:00", true},
		{"datetime", "string", "2026-09-06T09:30:00", false}, {"datetime", "string", "2026-09-06 09:30:00Z", false},
		{"datetime", "string", "2026-09-06T09:30:00+24:00", false}, {"datetime", "string", "2026-09-06T09:30:00+03:60", false},
	}
	for _, test := range cases {
		t.Run(test.fieldType+" "+test.kind+" "+test.value, func(t *testing.T) {
			if got := queryLiteralMatchesField(QueryLiteral{Kind: test.kind, Value: test.value}, test.fieldType); got != test.valid {
				t.Fatalf("got %v, want %v", got, test.valid)
			}
		})
	}
}

func TestQueryOptionalDefaultsAndStoredTypeSupport(t *testing.T) {
	source := queryTestEntities + `query AllProducts { source Product }
query TypedProducts {
  source Product
  where stock >= -1
  where stock < 100
  where units == 2
  where price > 0.25
  where ratio <= 1.5
  where name == "Widget"
  where email != ""
  where active == false
  where due >= "2026-09-06"
  where updated < "2026-09-07T00:00:00Z"
  sort active desc
  limit 1000
}
`
	program, diagnostics := Parse("query.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("parse: %#v", diagnostics)
	}
	if diagnostics = Validate(program); len(diagnostics) != 0 {
		t.Fatalf("validate: %#v", diagnostics)
	}
	if program.Queries[0].Limit != 0 || program.Queries[0].Sort.Field != "" || len(program.Queries[0].Where) != 0 {
		t.Fatalf("omitted defaults should stay omitted in AST: %#v", program.Queries[0])
	}
}

func TestValidateDuplicatePageRoute(t *testing.T) {
	program, diagnostics := Parse("query.black", queryTestEntities+"page LowStock { source Product }\npage Lowstock { source Product }\n")
	if len(diagnostics) != 0 {
		t.Fatalf("parse: %#v", diagnostics)
	}
	assertQueryDiagnostic(t, Validate(program), "DUPLICATE_PAGE_ROUTE")
}

func assertQueryDiagnostic(t *testing.T, diagnostics []Diagnostic, code string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			if diagnostic.File != "query.black" || diagnostic.Line < 1 || diagnostic.Column < 1 {
				t.Fatalf("diagnostic lacks source position: %#v", diagnostic)
			}
			return
		}
	}
	t.Fatalf("expected %s, got %#v", code, diagnostics)
}
