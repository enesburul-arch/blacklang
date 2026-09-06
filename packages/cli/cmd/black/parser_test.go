package main

import "testing"

func TestParseWarehouseExample(t *testing.T) {
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
  computed inventoryValue money = stock * price label "Inventory Value"
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
	if program.Target == nil || program.Target.Name != "web" || program.Target.Frontend != "react" || program.Target.Backend != "node" || program.Target.Database != "sqlite" {
		t.Fatalf("expected web react node sqlite target, got %#v", program.Target)
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
	if len(program.Entities[0].ComputedFields) != 1 {
		t.Fatalf("expected 1 computed field, got %#v", program.Entities[0].ComputedFields)
	}
	computed := program.Entities[0].ComputedFields[0]
	if computed.Name != "inventoryValue" || computed.Type != "money" || computed.Expression.Left != "stock" || computed.Expression.Operator != "*" || computed.Expression.Right != "price" {
		t.Fatalf("expected inventoryValue computed field, got %#v", computed)
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
	if program.Pages[0].View == nil || len(program.Pages[0].View.Order) != 3 || program.Pages[0].View.Order[0] != "form" || program.Pages[0].View.Order[1] != "table" || program.Pages[0].View.Order[2] != "detail" {
		t.Fatalf("expected view order form table detail, got %#v", program.Pages[0].View)
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

func TestParseI18NAndLabelTranslations(t *testing.T) {
	source := `app Warehouse

i18n {
  default tr
  locales tr, en
}

label Product.name {
  tr "Ürün Adı"
  en "Product Name"
}

entity Product {
  name text required
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if program.I18N == nil || program.I18N.Default != "tr" || len(program.I18N.Locales) != 2 {
		t.Fatalf("expected i18n declaration, got %#v", program.I18N)
	}
	if len(program.Labels) != 1 || program.Labels[0].Target != "Product.name" {
		t.Fatalf("expected Product.name label translation, got %#v", program.Labels)
	}
	if program.Labels[0].Translations[0].Locale != "tr" || program.Labels[0].Translations[0].Text != "Ürün Adı" {
		t.Fatalf("expected Turkish translation, got %#v", program.Labels[0].Translations)
	}
}

func TestParseSecurityCORS(t *testing.T) {
	source := `app Warehouse

security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}

entity Product {
  name text required
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if program.Security == nil || program.Security.CORS == nil {
		t.Fatalf("expected security cors declaration, got %#v", program.Security)
	}
	if program.Security.CORS.Origins.Name != "CORS_ORIGINS" {
		t.Fatalf("expected CORS_ORIGINS, got %#v", program.Security.CORS.Origins)
	}
	if program.Security.CORS.Credentials != "true" {
		t.Fatalf("expected credentials true, got %#v", program.Security.CORS)
	}
}

func TestParseTargetWeb(t *testing.T) {
	source := `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  name text required
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if program.Target == nil {
		t.Fatalf("expected target declaration")
	}
	if program.Target.Name != "web" || program.Target.Frontend != "react" || program.Target.Backend != "node" || program.Target.Database != "sqlite" {
		t.Fatalf("expected web react node sqlite target, got %#v", program.Target)
	}
}

func TestParseDeployDocker(t *testing.T) {
	source := `app Warehouse

deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  env CORS_ORIGINS optional
}

entity Product {
  name text required
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	if program.Deploy == nil {
		t.Fatalf("expected deploy declaration")
	}
	if program.Deploy.Target != "docker" {
		t.Fatalf("expected docker target, got %#v", program.Deploy)
	}
	if program.Deploy.Port == nil || program.Deploy.Port.Env.Name != "PORT" || program.Deploy.Port.Default != "3001" {
		t.Fatalf("expected port env PORT default 3001, got %#v", program.Deploy.Port)
	}
	if len(program.Deploy.Env) != 2 || program.Deploy.Env[0].Name != "DATABASE_URL" || program.Deploy.Env[0].Mode != "required" {
		t.Fatalf("expected deploy env declarations, got %#v", program.Deploy.Env)
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

func TestParseInlineUIIntent(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required label "Product Name" ui text "#172026" 14 semibold left
}

page Products {
  source Product

  table {
    columns name
    ui table border 1 solid compact true
  }

  form {
    fields name
    ui box black 1 solid 8 8 5 5 6 center | text "#172026" 14 regular left
  }

  actions create, edit
  action create ui button primary white 6 md solid
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	fieldUI := program.Entities[0].Fields[0].UI
	if len(fieldUI) != 1 || fieldUI[0].Mode != "text" || fieldUI[0].Values[0] != "#172026" {
		t.Fatalf("expected field text ui, got %#v", fieldUI)
	}
	tableUI := program.Pages[0].Table.UI
	if len(tableUI) != 1 || tableUI[0].Mode != "table" || tableUI[0].Values[3] != "compact" {
		t.Fatalf("expected table ui intent, got %#v", tableUI)
	}
	formUI := program.Pages[0].Form.UI
	if len(formUI) != 2 || formUI[0].Mode != "box" || formUI[1].Mode != "text" {
		t.Fatalf("expected two form ui modes, got %#v", formUI)
	}
	actionUI := program.Pages[0].ActionUI
	if len(actionUI) != 1 || actionUI[0].Action != "create" || actionUI[0].UI[0].Mode != "button" {
		t.Fatalf("expected create button ui, got %#v", actionUI)
	}
}

func TestParseExplicitUIIdentity(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  table {
    id ProductsTable
    class denseTable trackedPanel
    columns name
  }

  form {
    id ProductForm
    class editPanel
    fields name
  }

  actions create, edit
  action create id CreateProductButton
  action create class primaryAction
  action create ui button primary white 6 md solid
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}

	page := program.Pages[0]
	if page.Table.Identity == nil || page.Table.Identity.ID != "ProductsTable" || len(page.Table.Identity.Classes) != 2 {
		t.Fatalf("expected table identity, got %#v", page.Table.Identity)
	}
	if page.Form.Identity == nil || page.Form.Identity.ID != "ProductForm" || page.Form.Identity.Classes[0] != "editPanel" {
		t.Fatalf("expected form identity, got %#v", page.Form.Identity)
	}
	if len(page.ActionUI) != 1 || page.ActionUI[0].Identity == nil || page.ActionUI[0].Identity.ID != "CreateProductButton" {
		t.Fatalf("expected merged action identity, got %#v", page.ActionUI)
	}
	if page.ActionUI[0].UI[0].Mode != "button" || page.ActionUI[0].Identity.Classes[0] != "primaryAction" {
		t.Fatalf("expected merged action UI and class, got %#v", page.ActionUI[0])
	}
}

func TestParsePageViewOrder(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    order form, table, detail
  }

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
		t.Fatalf("expected no diagnostics, got %#v", diagnostics)
	}
	view := program.Pages[0].View
	if view == nil || len(view.Order) != 3 {
		t.Fatalf("expected view order, got %#v", view)
	}
	if view.Order[0] != "form" || view.Order[1] != "table" || view.Order[2] != "detail" {
		t.Fatalf("expected form table detail order, got %#v", view.Order)
	}
}

func TestParseInvalidInlineUIIntent(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text ui text
}

page Products {
  source Product
  action create button primary
}
`

	_, diagnostics := Parse("test.black", source)
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, code := range []string{"INVALID_UI_INTENT", "INVALID_ACTION_INTENT"} {
		if !codes[code] {
			t.Fatalf("expected parse code %s, got %#v", code, diagnostics)
		}
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
