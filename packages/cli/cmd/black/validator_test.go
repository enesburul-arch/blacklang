package main

import "testing"

func TestValidateWarehouseExample(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required unique
  name text required
  stock number default 0
  price money
}

page Products {
  source Product

  table {
    columns sku, name, stock, price
    search sku, name
  }

  form {
    fields sku, name, stock, price
  }

  actions create, edit, delete, archive, restore
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateUnknownPageField(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required unique
}

page Products {
  source Product

  table {
    columns sku, barcode
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_TABLE_COLUMN" {
		t.Fatalf("expected UNKNOWN_TABLE_COLUMN, got %q", diagnostics[0].Code)
	}
}

func TestValidateEntityReferenceField(t *testing.T) {
	source := `app Warehouse

entity Customer {
  name text required
}

entity Order {
  customer Customer required
  total money default 0
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateEntityReferenceSearchField(t *testing.T) {
	source := `app Warehouse

entity Customer {
  name text required
}

entity Order {
  customer Customer required
}

page Orders {
  source Order

  table {
    columns customer
    search customer
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateUnsupportedAction(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required unique
}

page Products {
  source Product
  actions create, publish
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNSUPPORTED_ACTION" {
		t.Fatalf("expected UNSUPPORTED_ACTION, got %q", diagnostics[0].Code)
	}
}

func TestValidateLabelModifierRequiresValue(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text label
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "MISSING_LABEL_VALUE" {
		t.Fatalf("expected MISSING_LABEL_VALUE, got %q", diagnostics[0].Code)
	}
}

func TestValidatePlaceholderModifierRequiresValue(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text placeholder
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "MISSING_PLACEHOLDER_VALUE" {
		t.Fatalf("expected MISSING_PLACEHOLDER_VALUE, got %q", diagnostics[0].Code)
	}
}

func TestValidateHelpModifierRequiresValue(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text help
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "MISSING_HELP_VALUE" {
		t.Fatalf("expected MISSING_HELP_VALUE, got %q", diagnostics[0].Code)
	}
}

func TestValidateConstraintModifiers(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text length 3..40
  stock number min 0 max 100
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateAdvancedValidationModifiers(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text regex "^[A-Z0-9]+$" message "Use uppercase letters and numbers"
  website text url
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateEntityValidation(t *testing.T) {
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

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateEntityValidationErrors(t *testing.T) {
	source := `app Warehouse

entity Order {
  total money
  status text
  validate missing <= total
  validate status < total
  validate total around status
  validate missing required when status == shipped
  validate status required when missing == shipped
  validate status required when status around shipped
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, code := range []string{"UNKNOWN_VALIDATION_FIELD", "INCOMPATIBLE_VALIDATION_FIELDS", "UNSUPPORTED_VALIDATION_OPERATOR"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateConstraintModifierErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text length 40..3
  name text min 0
  stock number length 3..40
  price money max nope
  code text regex "["
  count number url
  slug text message
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, code := range []string{"INVALID_LENGTH_CONSTRAINT", "UNSUPPORTED_NUMERIC_CONSTRAINT", "UNSUPPORTED_LENGTH_CONSTRAINT", "INVALID_NUMERIC_CONSTRAINT", "INVALID_REGEX_CONSTRAINT", "UNSUPPORTED_URL_CONSTRAINT", "MISSING_MESSAGE_VALUE"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateDatabaseEnvReference(t *testing.T) {
	source := `app Warehouse

database {
  url env DATABASE_URL
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateDatabaseRequiresURL(t *testing.T) {
	source := `app Warehouse

database {
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "MISSING_DATABASE_URL" {
		t.Fatalf("expected MISSING_DATABASE_URL, got %q", diagnostics[0].Code)
	}
}

func TestValidateExplicitAPIDeclaration(t *testing.T) {
	source := `app Warehouse

api LowStockReport {
  method GET
  path "/api/reports/low-stock/{warehouseId}"
  param warehouseId text
  query limit integer
  private
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateExplicitAPIErrors(t *testing.T) {
	source := `app Warehouse

api Broken {
  method TRACE
  path "api/broken/{id}"
  param other text
  query limit Unknown
}

api Broken {
  method GET
  path "/api/broken"
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, code := range []string{"UNSUPPORTED_API_METHOD", "INVALID_API_PATH", "UNUSED_API_PARAM", "UNSUPPORTED_API_QUERY_TYPE", "DUPLICATE_API"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateDatabaseEnvName(t *testing.T) {
	source := `app Warehouse

database {
  url env database_url
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "INVALID_ENV_NAME" {
		t.Fatalf("expected INVALID_ENV_NAME, got %q", diagnostics[0].Code)
	}
}

func TestParseRejectsLiteralDatabaseURL(t *testing.T) {
	source := `app Warehouse

database {
  url "postgres://user:password@example.com/app"
}
`

	_, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 parse diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "INVALID_DATABASE_URL" {
		t.Fatalf("expected INVALID_DATABASE_URL, got %q", diagnostics[0].Code)
	}
}

func TestValidateUnknownSortField(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
}

page Products {
  source Product
  table {
    columns sku
    sort name asc
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_SORT_FIELD" {
		t.Fatalf("expected UNKNOWN_SORT_FIELD, got %q", diagnostics[0].Code)
	}
}

func TestValidateUnsupportedSortDirection(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
}

page Products {
  source Product
  table {
    columns sku
    sort sku newest
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNSUPPORTED_SORT_DIRECTION" {
		t.Fatalf("expected UNSUPPORTED_SORT_DIRECTION, got %q", diagnostics[0].Code)
	}
}

func TestValidateUnknownFilterField(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
}

page Products {
  source Product
  table {
    columns sku
    filter name
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_FILTER_FIELD" {
		t.Fatalf("expected UNKNOWN_FILTER_FIELD, got %q", diagnostics[0].Code)
	}
}

func TestValidateUnknownPageLayout(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
}

page Products {
  layout MissingLayout
  source Product
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_PAGE_LAYOUT" {
		t.Fatalf("expected UNKNOWN_PAGE_LAYOUT, got %q", diagnostics[0].Code)
	}
}

func TestValidateUnknownSidebarItem(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
}

layout AdminLayout {
  sidebar {
    item MissingPage
  }
}

page Products {
  layout AdminLayout
  source Product
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_SIDEBAR_ITEM" {
		t.Fatalf("expected UNKNOWN_SIDEBAR_ITEM, got %q", diagnostics[0].Code)
	}
}

func TestValidateAuthDeclaration(t *testing.T) {
	source := `app Warehouse

auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateUnsupportedAuthStrategy(t *testing.T) {
	source := `app Warehouse

auth {
  strategy magicLink
  session jwt

  user {
    email email required unique
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 2 {
		t.Fatalf("expected 2 validation diagnostics, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNSUPPORTED_AUTH_STRATEGY" {
		t.Fatalf("expected UNSUPPORTED_AUTH_STRATEGY, got %q", diagnostics[0].Code)
	}
	if diagnostics[1].Code != "UNSUPPORTED_AUTH_SESSION" {
		t.Fatalf("expected UNSUPPORTED_AUTH_SESSION, got %q", diagnostics[1].Code)
	}
}

func TestValidateRoleAndPageAccess(t *testing.T) {
	source := `app Warehouse

auth {
  strategy emailPassword
  session cookie

  user {
    email email required unique
  }
}

entity Product {
  sku text
}

role Admin {
  allow all
}

role Worker {
  allow read Product
  deny delete Product
}

page Products {
  source Product
  access Admin, Worker
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateUnknownAccessRole(t *testing.T) {
	source := `app Warehouse

auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}

entity Product {
  sku text
}

page Products {
  source Product
  access MissingRole
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_ACCESS_ROLE" {
		t.Fatalf("expected UNKNOWN_ACCESS_ROLE, got %q", diagnostics[0].Code)
	}
}

func TestValidateAccessRequiresAuth(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
}

role Admin {
  allow all
}

page Products {
  source Product
  access Admin
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "AUTH_REQUIRED_FOR_ACCESS" {
		t.Fatalf("expected AUTH_REQUIRED_FOR_ACCESS, got %q", diagnostics[0].Code)
	}
}

func TestValidateUnknownPermissionResource(t *testing.T) {
	source := `app Warehouse

role Admin {
  allow read MissingEntity
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_PERMISSION_RESOURCE" {
		t.Fatalf("expected UNKNOWN_PERMISSION_RESOURCE, got %q", diagnostics[0].Code)
	}
}

func TestValidateUnknownPermissionField(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
}

role Admin {
  allow read Product missingField
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_PERMISSION_FIELD" {
		t.Fatalf("expected UNKNOWN_PERMISSION_FIELD, got %q", diagnostics[0].Code)
	}
}

func TestValidateWorkflowDeclaration(t *testing.T) {
	source := `app Warehouse

auth {
  strategy emailPassword
  session cookie
  user {
    email email required unique
  }
}

entity Order {
  status text default draft
}

role Admin {
  allow all
}

workflow OrderPreparation {
  source Order
  states draft, picking, shipped

  transition ship {
    from picking
    to shipped
    allow Admin
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateUnknownWorkflowTransitionState(t *testing.T) {
	source := `app Warehouse

entity Order {
  status text
}

workflow OrderPreparation {
  source Order
  states draft, shipped

  transition ship {
    from packaged
    to shipped
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_TRANSITION_FROM" {
		t.Fatalf("expected UNKNOWN_TRANSITION_FROM, got %q", diagnostics[0].Code)
	}
}

func TestValidateWorkflowRequiresStatusField(t *testing.T) {
	source := `app Warehouse

entity Order {
  total money
}

workflow OrderPreparation {
  source Order
  states draft, shipped

  transition ship {
    from draft
    to shipped
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "MISSING_WORKFLOW_STATUS_FIELD" {
		t.Fatalf("expected MISSING_WORKFLOW_STATUS_FIELD, got %q", diagnostics[0].Code)
	}
}

func TestValidateUnknownWorkflowAllowRole(t *testing.T) {
	source := `app Warehouse

auth {
  strategy emailPassword
  session cookie
  user {
    email email required unique
  }
}

entity Order {
  status text
}

workflow OrderPreparation {
  source Order
  states draft, shipped

  transition ship {
    from draft
    to shipped
    allow MissingRole
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNKNOWN_WORKFLOW_ALLOW_ROLE" {
		t.Fatalf("expected UNKNOWN_WORKFLOW_ALLOW_ROLE, got %q", diagnostics[0].Code)
	}
}

func TestValidateStateDeclaration(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text
}

state ProductPageState {
  selectedProducts Product[]
  activeFilter text
  modal createProduct closed
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateInvalidStateDeclaration(t *testing.T) {
	source := `app Warehouse

state ProductPageState {
  selectedProducts MissingEntity[]
  modal createProduct hidden
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 2 {
		t.Fatalf("expected 2 validation diagnostics, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNSUPPORTED_STATE_FIELD_TYPE" {
		t.Fatalf("expected UNSUPPORTED_STATE_FIELD_TYPE, got %q", diagnostics[0].Code)
	}
	if diagnostics[1].Code != "UNSUPPORTED_STATE_MODAL_DEFAULT" {
		t.Fatalf("expected UNSUPPORTED_STATE_MODAL_DEFAULT, got %q", diagnostics[1].Code)
	}
}

func TestValidateComponentDeclaration(t *testing.T) {
	source := `app Warehouse

entity Product {
  stock number
}

component StockBadge {
  input product Product
  input stock number

  variant low when stock < 10
  variant normal when stock >= 10
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateInvalidComponentDeclaration(t *testing.T) {
	source := `app Warehouse

component StockBadge {
  input stock MissingEntity
  input stock number

  variant low when stock < 10
  variant low when stock >= 10
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 3 {
		t.Fatalf("expected 3 validation diagnostics, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "UNSUPPORTED_COMPONENT_INPUT_TYPE" {
		t.Fatalf("expected UNSUPPORTED_COMPONENT_INPUT_TYPE, got %q", diagnostics[0].Code)
	}
	if diagnostics[1].Code != "DUPLICATE_COMPONENT_INPUT" {
		t.Fatalf("expected DUPLICATE_COMPONENT_INPUT, got %q", diagnostics[1].Code)
	}
	if diagnostics[2].Code != "DUPLICATE_COMPONENT_VARIANT" {
		t.Fatalf("expected DUPLICATE_COMPONENT_VARIANT, got %q", diagnostics[2].Code)
	}
}
