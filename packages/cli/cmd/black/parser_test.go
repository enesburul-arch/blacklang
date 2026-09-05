package main

import "testing"

func TestParseWarehouseExample(t *testing.T) {
	source := `app Warehouse

auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}

database {
  url env DATABASE_URL
}

entity Product {
  sku text required unique
  name text required
  stock number default 0
  price money
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
    columns sku, name, stock, price
    search sku, name
    filter stock
    sort stock desc
    paginate 10
  }

  form {
    fields sku, name, stock, price
  }

  actions create, edit, delete, archive, restore
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if program.App.Name != "Warehouse" {
		t.Fatalf("expected app Warehouse, got %q", program.App.Name)
	}
	if program.Auth == nil {
		t.Fatalf("expected auth declaration")
	}
	if program.Auth.Strategy != "emailPassword" || program.Auth.Session != "cookie" {
		t.Fatalf("expected emailPassword cookie auth, got %#v", program.Auth)
	}
	if len(program.Auth.User.Fields) != 2 {
		t.Fatalf("expected 2 auth user fields, got %#v", program.Auth.User.Fields)
	}
	if program.Database == nil || program.Database.URL.Name != "DATABASE_URL" {
		t.Fatalf("expected database url env DATABASE_URL, got %#v", program.Database)
	}
	if len(program.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(program.Entities))
	}
	if program.Entities[0].Name != "Product" {
		t.Fatalf("expected Product entity, got %q", program.Entities[0].Name)
	}
	if len(program.Entities[0].Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(program.Entities[0].Fields))
	}
	if len(program.Roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(program.Roles))
	}
	if program.Roles[0].Name != "Admin" || len(program.Roles[0].Permissions) != 1 || program.Roles[0].Permissions[0].Action != "all" {
		t.Fatalf("expected Admin allow all, got %#v", program.Roles[0])
	}
	if len(program.Workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(program.Workflows))
	}
	if program.Workflows[0].Name != "OrderPreparation" || program.Workflows[0].Source != "Product" {
		t.Fatalf("expected OrderPreparation workflow for Product, got %#v", program.Workflows[0])
	}
	if len(program.Workflows[0].States) != 3 || program.Workflows[0].States[2] != "shipped" {
		t.Fatalf("expected workflow states, got %#v", program.Workflows[0].States)
	}
	if len(program.Workflows[0].Transitions) != 1 || program.Workflows[0].Transitions[0].Allow[0] != "Admin" {
		t.Fatalf("expected workflow transition allow Admin, got %#v", program.Workflows[0].Transitions)
	}
	if len(program.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(program.States))
	}
	if program.States[0].Name != "ProductPageState" || len(program.States[0].Fields) != 2 || len(program.States[0].Modals) != 1 {
		t.Fatalf("expected ProductPageState fields and modal, got %#v", program.States[0])
	}
	if program.States[0].Fields[0].Type != "Product" || !program.States[0].Fields[0].List {
		t.Fatalf("expected selectedProducts Product[] field, got %#v", program.States[0].Fields[0])
	}
	if program.States[0].Modals[0].Name != "createProduct" || program.States[0].Modals[0].Default != "closed" {
		t.Fatalf("expected createProduct closed modal, got %#v", program.States[0].Modals[0])
	}
	if len(program.Components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(program.Components))
	}
	if program.Components[0].Name != "StockBadge" || len(program.Components[0].Inputs) != 1 || len(program.Components[0].Variants) != 2 {
		t.Fatalf("expected StockBadge component, got %#v", program.Components[0])
	}
	if program.Components[0].Variants[0].Name != "low" || program.Components[0].Variants[0].Condition != "stock < 10" {
		t.Fatalf("expected low stock variant, got %#v", program.Components[0].Variants[0])
	}
	if len(program.Pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(program.Pages))
	}
	if len(program.Layouts) != 1 {
		t.Fatalf("expected 1 layout, got %d", len(program.Layouts))
	}
	if program.Layouts[0].Name != "AdminLayout" || len(program.Layouts[0].Sidebar.Items) != 1 || program.Layouts[0].Sidebar.Items[0] != "Products" {
		t.Fatalf("expected AdminLayout sidebar Products, got %#v", program.Layouts[0])
	}
	if program.Pages[0].Layout != "AdminLayout" {
		t.Fatalf("expected Products layout AdminLayout, got %q", program.Pages[0].Layout)
	}
	if program.Pages[0].Source != "Product" {
		t.Fatalf("expected Products source Product, got %q", program.Pages[0].Source)
	}
	if len(program.Pages[0].Access) != 1 || program.Pages[0].Access[0] != "Admin" {
		t.Fatalf("expected Products access Admin, got %#v", program.Pages[0].Access)
	}
	if len(program.Pages[0].Table.Columns) != 4 {
		t.Fatalf("expected 4 table columns, got %d", len(program.Pages[0].Table.Columns))
	}
	if program.Pages[0].Table.Sort.Field != "stock" || program.Pages[0].Table.Sort.Direction != "desc" {
		t.Fatalf("expected sort stock desc, got %#v", program.Pages[0].Table.Sort)
	}
	if len(program.Pages[0].Table.Filters) != 1 || program.Pages[0].Table.Filters[0] != "stock" {
		t.Fatalf("expected filter stock, got %#v", program.Pages[0].Table.Filters)
	}
	if program.Pages[0].Table.Paginate != 10 {
		t.Fatalf("expected paginate 10, got %d", program.Pages[0].Table.Paginate)
	}
	if len(program.Pages[0].Actions) != 5 {
		t.Fatalf("expected 5 actions, got %d", len(program.Pages[0].Actions))
	}
}

func TestParseReportsInvalidTableSort(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
}

page Products {
  source Product
  table {
    columns sku
    sort sku
  }
}
`

	_, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "INVALID_TABLE_SORT" {
		t.Fatalf("expected INVALID_TABLE_SORT, got %q", diagnostics[0].Code)
	}
}

func TestParseReportsInvalidTablePagination(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
}

page Products {
  source Product
  table {
    columns sku
    paginate zero
  }
}
`

	_, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "INVALID_TABLE_PAGINATION" {
		t.Fatalf("expected INVALID_TABLE_PAGINATION, got %q", diagnostics[0].Code)
	}
}

func TestParseReportsUnclosedEntity(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required unique
`

	_, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNCLOSED_ENTITY" {
		t.Fatalf("expected UNCLOSED_ENTITY, got %q", diagnostics[0].Code)
	}
}

func TestParseQuotedFieldTextModifiers(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required label "Product Name" placeholder "Enter product name" help "Visible product name"
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	field := program.Entities[0].Fields[0]
	if len(field.Modifiers) != 4 {
		t.Fatalf("expected 4 modifiers, got %#v", field.Modifiers)
	}
	if field.Modifiers[1].Name != "label" || field.Modifiers[1].Value != "Product Name" {
		t.Fatalf("expected quoted label modifier, got %#v", field.Modifiers[1])
	}
	if field.Modifiers[2].Name != "placeholder" || field.Modifiers[2].Value != "Enter product name" {
		t.Fatalf("expected quoted placeholder modifier, got %#v", field.Modifiers[2])
	}
	if field.Modifiers[3].Name != "help" || field.Modifiers[3].Value != "Visible product name" {
		t.Fatalf("expected quoted help modifier, got %#v", field.Modifiers[3])
	}
}

func TestParseTokenStreamKeepsCommentsOutsideQuotedStrings(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text label "Product #1" help "Docs live at https://example.com/product" // trailing comment
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	field := program.Entities[0].Fields[0]
	if field.Modifiers[0].Value != "Product #1" {
		t.Fatalf("expected quoted # text to survive tokenization, got %#v", field.Modifiers)
	}
	if field.Modifiers[1].Value != "Docs live at https://example.com/product" {
		t.Fatalf("expected quoted // text to survive tokenization, got %#v", field.Modifiers)
	}
}

func TestParseTokenStreamSplitsInlineBraces(t *testing.T) {
	source := `app Warehouse

entity Product { sku text required }
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if len(program.Entities) != 1 || len(program.Entities[0].Fields) != 1 || program.Entities[0].Fields[0].Name != "sku" {
		t.Fatalf("expected inline entity to parse, got %#v", program.Entities)
	}
}

func TestParseTokenStreamReportsUnclosedString(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text label "Product Name
}
`

	_, diagnostics := Parse("test.black", source)
	if len(diagnostics) == 0 {
		t.Fatalf("expected unclosed string diagnostic")
	}
	if diagnostics[0].Code != "UNCLOSED_STRING" {
		t.Fatalf("expected UNCLOSED_STRING, got %#v", diagnostics)
	}
}

func TestParseFieldConstraintModifiers(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required length 3..40
  stock number min 0 max 100
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	sku := program.Entities[0].Fields[0]
	stock := program.Entities[0].Fields[1]
	if sku.Modifiers[1].Name != "length" || sku.Modifiers[1].Value != "3..40" {
		t.Fatalf("expected length modifier value, got %#v", sku.Modifiers)
	}
	if stock.Modifiers[0].Name != "min" || stock.Modifiers[0].Value != "0" {
		t.Fatalf("expected min modifier value, got %#v", stock.Modifiers)
	}
	if stock.Modifiers[1].Name != "max" || stock.Modifiers[1].Value != "100" {
		t.Fatalf("expected max modifier value, got %#v", stock.Modifiers)
	}
}

func TestParseAdvancedValidationModifiers(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required regex "^[A-Z0-9]+$" message "Use uppercase letters and numbers"
  website text optional url
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	sku := program.Entities[0].Fields[0]
	website := program.Entities[0].Fields[1]
	if sku.Modifiers[1].Name != "regex" || sku.Modifiers[1].Value != "^[A-Z0-9]+$" {
		t.Fatalf("expected regex modifier value, got %#v", sku.Modifiers)
	}
	if sku.Modifiers[2].Name != "message" || sku.Modifiers[2].Value != "Use uppercase letters and numbers" {
		t.Fatalf("expected message modifier value, got %#v", sku.Modifiers)
	}
	if website.Modifiers[1].Name != "url" {
		t.Fatalf("expected url modifier, got %#v", website.Modifiers)
	}
}

func TestParseExplicitAPIDeclaration(t *testing.T) {
	source := `app Warehouse

api LowStockReport {
  method GET
  path "/api/reports/low-stock/{warehouseId}"
  param warehouseId text
  query limit integer
  private
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  webhook
  public
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	if len(program.APIs) != 2 {
		t.Fatalf("expected 2 apis, got %#v", program.APIs)
	}
	report := program.APIs[0]
	if report.Name != "LowStockReport" || report.Method != "GET" || report.Path != "/api/reports/low-stock/{warehouseId}" || report.Access != "private" {
		t.Fatalf("unexpected report api: %#v", report)
	}
	if len(report.Params) != 1 || report.Params[0].Name != "warehouseId" || report.Params[0].Type != "text" {
		t.Fatalf("unexpected report params: %#v", report.Params)
	}
	if len(report.Queries) != 1 || report.Queries[0].Name != "limit" || report.Queries[0].Type != "integer" {
		t.Fatalf("unexpected report queries: %#v", report.Queries)
	}
	if !program.APIs[1].Webhook || program.APIs[1].Access != "public" {
		t.Fatalf("unexpected webhook api: %#v", program.APIs[1])
	}
}

func TestParseEntityValidation(t *testing.T) {
	source := `app Warehouse

entity Order {
  total money
  discount money
  status text
  trackingNumber text optional
  validate discount <= total message "Discount cannot exceed total"
  validate trackingNumber required when status == shipped message "Tracking number is required when shipped"
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	validations := program.Entities[0].Validations
	if len(validations) != 2 {
		t.Fatalf("expected 2 validations, got %#v", validations)
	}
	if validations[0].Left != "discount" || validations[0].Operator != "<=" || validations[0].Right != "total" || validations[0].Message != "Discount cannot exceed total" {
		t.Fatalf("unexpected validation: %#v", validations[0])
	}
	if validations[1].Left != "trackingNumber" || !validations[1].Required || validations[1].When == nil || validations[1].When.Left != "status" || validations[1].When.Operator != "==" || validations[1].When.Right != "shipped" || validations[1].Message != "Tracking number is required when shipped" {
		t.Fatalf("unexpected conditional validation: %#v", validations[1])
	}
}
