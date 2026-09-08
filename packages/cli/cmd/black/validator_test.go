package main

import (
	"strings"
	"testing"
)

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

func TestValidateMigrationRenames(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required
  name text required
}

migration RenameProductName {
  rename entity ProductItem to Product
  rename field Product.title to name
}
`
	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) > 0 {
		t.Fatalf("unexpected parse diagnostics: %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) > 0 {
		t.Fatalf("unexpected validation diagnostics: %#v", diagnostics)
	}
}

func TestValidateMigrationRenameDiagnostics(t *testing.T) {
	source := `app Warehouse

entity Product {
  oldName text
  total money
  computed displayTotal money = total * 2
}

migration BadRename {
  rename field Product.oldName to missing
  rename field Product.oldTotal to displayTotal
  rename entity Product to RenamedProduct
}
`
	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) > 0 {
		t.Fatalf("unexpected parse diagnostics: %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	for _, code := range []string{
		"UNKNOWN_MIGRATION_RENAME_TARGET",
		"UNSUPPORTED_MIGRATION_RENAME_TARGET",
		"MIGRATION_RENAME_SOURCE_STILL_EXISTS",
	} {
		if !hasDiagnostic(diagnostics, code) {
			t.Fatalf("expected diagnostic %s in %#v", code, diagnostics)
		}
	}
}

func TestValidateEntityPolicies(t *testing.T) {
	source := `app Warehouse

auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}

entity WorkItem {
  ownerId text required
  tenantId text required
  title text required
  policy owner ownerId
  policy tenant tenantId
}

role Admin {
  allow all
}

page WorkItems {
  source WorkItem
  access Admin

  table {
    columns title
  }

  form {
    fields title
  }

  actions create, edit, delete
}
`
	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) > 0 {
		t.Fatalf("unexpected parse diagnostics: %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) > 0 {
		t.Fatalf("unexpected validation diagnostics: %#v", diagnostics)
	}
}

func TestValidateEntityPolicyDiagnostics(t *testing.T) {
	source := `app Warehouse

entity WorkItem {
  ownerId text unique
  computed displayTenant text = ownerId + ownerId
  title text required
  policy owner ownerId
  policy owner missingOwner
  policy tenant displayTenant
}

entity Ticket {
  title text required
  policy tenant missingTenant
}

action BadOwnershipUpdate {
  source WorkItem
  input title text required
  set ownerId = title
}

page WorkItems {
  source WorkItem

  table {
    columns title
  }

  form {
    fields title, ownerId
  }

  actions create, BadOwnershipUpdate
}
`
	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) > 0 {
		t.Fatalf("unexpected parse diagnostics: %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	for _, code := range []string{
		"AUTH_REQUIRED_FOR_ENTITY_POLICY",
		"MISSING_ENTITY_POLICY_REQUIRED_FIELD",
		"UNSUPPORTED_ENTITY_POLICY_UNIQUE_FIELD",
		"DUPLICATE_ENTITY_POLICY",
		"UNKNOWN_ENTITY_POLICY_FIELD",
		"UNSUPPORTED_ENTITY_POLICY_FIELD",
		"UNSUPPORTED_POLICY_FORM_FIELD",
		"UNSUPPORTED_ACTION_POLICY_FIELD",
	} {
		if !hasDiagnostic(diagnostics, code) {
			t.Fatalf("expected %s diagnostic, got %#v", code, diagnostics)
		}
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

func TestValidateComputedField(t *testing.T) {
	source := `app Warehouse

entity Product {
  stock number default 0
  price money min 0
  computed inventoryValue money = stock * price label "Inventory Value"
}

page Products {
  source Product

  table {
    columns stock, price, inventoryValue
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

func TestValidateEntityIndexes(t *testing.T) {
	source := `app Warehouse

entity Customer {
  name text required
}

entity Order {
  customer Customer required
  status text
  total money
  index customer
  index customer, status
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

func TestValidateEntityIndexErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required
  stock integer
  price money
  status text
  computed inventoryValue money = stock * price
  index stock, stock
  index inventoryValue
  index missing
  index sku, stock, price, status, missing
  index stock, price
  index stock, price
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	for _, code := range []string{"DUPLICATE_INDEX_FIELD", "UNSUPPORTED_COMPUTED_INDEX_FIELD", "UNKNOWN_INDEX_FIELD", "UNSUPPORTED_ENTITY_INDEX_FIELD_COUNT", "DUPLICATE_ENTITY_INDEX"} {
		if !hasDiagnostic(diagnostics, code) {
			t.Fatalf("expected %s diagnostic, got %#v", code, diagnostics)
		}
	}
}

func TestValidateRelationLoadPolicy(t *testing.T) {
	source := `app Warehouse

entity Customer {
  name text required
}

entity Order {
  customer Customer required load list detail query mutation
  reviewer Customer optional load none
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateRelationLoadPolicyErrors(t *testing.T) {
	source := `app Warehouse

auth {
  strategy emailPassword
  session cookie

  user {
    email email required unique load list
  }
}

entity Customer {
  name text required
}

entity Order {
  customer Customer required load
  reviewer Customer optional load list list
  approver Customer optional load list load detail
  archivedBy Customer optional load none detail
  shipper Customer optional load compact
  name text load detail
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	for _, code := range []string{"UNSUPPORTED_AUTH_USER_FIELD_MODIFIER", "MISSING_RELATION_LOAD_SCOPE", "DUPLICATE_RELATION_LOAD_SCOPE", "DUPLICATE_RELATION_LOAD", "CONFLICTING_RELATION_LOAD_SCOPE", "UNSUPPORTED_RELATION_LOAD_SCOPE", "UNSUPPORTED_RELATION_LOAD_FIELD"} {
		if !hasDiagnostic(diagnostics, code) {
			t.Fatalf("expected %s diagnostic, got %#v", code, diagnostics)
		}
	}
}

func TestValidateComputedFieldErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text
  stock number default 0
  computed stock money = stock * missing
  computed missingValue money = stock * missing
  computed labelText text = stock * 2
  computed bad money = name * 2 sparkle
}

page Products {
  source Product

  table {
    columns bad
    search bad
    filter bad
    sort bad asc
  }

  form {
    fields bad
  }
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
	for _, code := range []string{"DUPLICATE_FIELD", "UNKNOWN_COMPUTED_FIELD", "UNSUPPORTED_COMPUTED_FIELD_TYPE", "INCOMPATIBLE_COMPUTED_FIELD", "UNSUPPORTED_COMPUTED_FIELD_MODIFIER", "UNSUPPORTED_COMPUTED_FORM_FIELD", "UNSUPPORTED_COMPUTED_SEARCH_FIELD", "UNSUPPORTED_COMPUTED_FILTER_FIELD", "UNSUPPORTED_COMPUTED_SORT_FIELD"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
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

func TestValidateInlineUIIntent(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required ui text "#172026" 14 semibold left
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

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateInlineUIIntentErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text ui button primary white
  sku text ui text black | text white
  status text ui sparkle bright
}

page Products {
  source Product

  table {
    columns name, sku, status
    ui button primary white
  }

  actions create
  action edit ui button primary white
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
	for _, code := range []string{"UNSUPPORTED_UI_TARGET_MODE", "DUPLICATE_UI_INTENT", "UNSUPPORTED_UI_MODE", "UNKNOWN_ACTION_UI"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidatePageViewErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    order form, sidebar, form
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

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, code := range []string{"UNSUPPORTED_VIEW_SECTION", "DUPLICATE_VIEW_SECTION"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidatePageViewRequiresOrder(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
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

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 validation diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Code != "MISSING_VIEW_ORDER" {
		t.Fatalf("expected MISSING_VIEW_ORDER, got %q", diagnostics[0].Code)
	}
}

func TestValidatePageViewCompositionErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    compose carousel columns 5 gap huge stackAt xl
    section table span 5
    section detail span 1
    section detail span 2
    section form display panel side top
    group Record sections table, form compose tabs columns 5 gap huge span 5
    group Record sections detail
    group DetailGroup sections detail, detail
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

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, code := range []string{"UNSUPPORTED_VIEW_COMPOSE_MODE", "UNSUPPORTED_VIEW_COMPOSE_COLUMNS", "UNSUPPORTED_VIEW_GAP", "UNSUPPORTED_VIEW_STACK_AT", "UNSUPPORTED_VIEW_SECTION_SPAN", "DUPLICATE_VIEW_SECTION", "UNSUPPORTED_VIEW_SECTION_DISPLAY", "UNSUPPORTED_VIEW_SECTION_SIDE", "DUPLICATE_VIEW_GROUP", "UNSUPPORTED_VIEW_GROUP_COMPOSE_MODE", "UNSUPPORTED_VIEW_GROUP_COMPOSE_COLUMNS", "UNSUPPORTED_VIEW_GROUP_GAP", "UNSUPPORTED_VIEW_GROUP_SPAN", "UNSUPPORTED_VIEW_GROUP_SECTION", "DUPLICATE_VIEW_GROUP_SECTION"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidatePageViewGroupRejectsTabsAndOverlaySections(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    order table, detail, form
    compose tabs gap md
    section form display modal
    group Record sections detail, form compose stack
    tab List sections table
    tab Record sections detail, form
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

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, code := range []string{"UNSUPPORTED_VIEW_GROUP", "UNSUPPORTED_VIEW_GROUP_SECTION_DISPLAY", "UNSUPPORTED_VIEW_SECTION_DISPLAY"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidatePageViewTriggerErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    order table, detail, form
    trigger sidebar on rowSelect
    trigger detail on hover
    trigger table on rowSelect
    trigger form on editStart
    trigger form on editStart
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

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, code := range []string{"UNSUPPORTED_VIEW_TRIGGER_SECTION", "UNSUPPORTED_VIEW_TRIGGER_EVENT", "UNSUPPORTED_VIEW_TRIGGER", "UNSUPPORTED_VIEW_TRIGGER_ACTION", "DUPLICATE_VIEW_TRIGGER"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidatePageViewTabsErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    order table, detail, form
    compose tabs columns 2 stackAt md
    tab List sections table, table
    tab Record sections detail, sidebar
    tab Record sections form
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

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	codes := map[string]bool{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code] = true
	}
	for _, code := range []string{"UNSUPPORTED_VIEW_COMPOSE_COLUMNS", "UNSUPPORTED_VIEW_STACK_AT", "DUPLICATE_VIEW_TAB_SECTION", "UNSUPPORTED_VIEW_TAB_SECTION", "DUPLICATE_VIEW_TAB", "MISSING_VIEW_TAB_SECTION"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidatePageViewTabsRequireComposeTabs(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    order table
    tab List sections table
  }

  table {
    columns name
  }
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
	if !codes["UNSUPPORTED_VIEW_TAB"] {
		t.Fatalf("expected UNSUPPORTED_VIEW_TAB, got %#v", diagnostics)
	}
}

func TestValidatePageViewTabsRequireTabDeclarations(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    order table
    compose tabs
  }

  table {
    columns name
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 1 || diagnostics[0].Code != "MISSING_VIEW_TABS" {
		t.Fatalf("expected MISSING_VIEW_TABS, got %#v", diagnostics)
	}
}

func TestValidatePageViewComponentSection(t *testing.T) {
	source := `app Warehouse

entity Product {
  stock number
  price money
  computed inventoryValue money = stock * price
}

component StockBadge {
  input stock number
  variant low when stock < 10
}

component InventoryValueBadge {
  input inventoryValue money
  variant high when inventoryValue > 1000
}

page Products {
  source Product

  view {
    order table, StockSummary, InventoryPreview, StockCards, detail, form
    compose grid columns 2 gap md
    section table span 2
    section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
    section InventoryPreview component InventoryValueBadge bind first span 1 title "Inventory Preview"
    section StockCards component StockBadge bind each span 1 title "Stock Cards"
    group Record sections StockSummary, InventoryPreview, StockCards compose grid columns 2 title "Record Components"
  }

  table {
    columns stock, price, inventoryValue
  }

  form {
    fields stock, price
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

func TestValidatePageViewComponentSectionErrors(t *testing.T) {
	source := `app Warehouse

entity Customer {
  name text
}

entity Product {
  stock number
}

component StockBadge {
  input stock number
}

component MissingFieldBadge {
  input missing number
}

component WrongTypeBadge {
  input stock text
}

component ListBadge {
  input stock number[]
}

component EntityBadge {
  input customer Customer
}

page Products {
  source Product

  view {
    order table, GhostSection, UnknownSection, NoBind, BadBind, MissingField, WrongType, ListSection, EntitySection, BuiltInComponent, detail, form
    section UnknownSection component MissingBadge bind selected
    section NoBind component StockBadge
    section BadBind component StockBadge bind all
    section MissingField component MissingFieldBadge bind selected
    section WrongType component WrongTypeBadge bind selected
    section ListSection component ListBadge bind selected
    section EntitySection component EntityBadge bind selected
    section detail component StockBadge bind selected
    section form bind selected
  }

  table {
    columns stock
  }
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	for _, code := range []string{
		"UNSUPPORTED_VIEW_SECTION",
		"UNKNOWN_VIEW_COMPONENT",
		"MISSING_VIEW_COMPONENT_BIND",
		"UNSUPPORTED_VIEW_COMPONENT_BIND",
		"UNKNOWN_VIEW_COMPONENT_INPUT_FIELD",
		"VIEW_COMPONENT_INPUT_TYPE_MISMATCH",
		"UNSUPPORTED_VIEW_COMPONENT_INPUT",
		"CONFLICTING_VIEW_COMPONENT_SECTION",
		"UNSUPPORTED_VIEW_SECTION_BIND",
	} {
		if !hasDiagnostic(diagnostics, code) {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateExplicitUIIdentityErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text
}

page Products {
  source Product

  table {
    id SharedPanel
    class repeated repeated
    columns name
  }

  form {
    id SharedPanel
    class validClass
    fields name
  }

  actions create
  action create id invalid.id
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
	for _, code := range []string{"DUPLICATE_UI_ID", "DUPLICATE_UI_CLASS", "INVALID_UI_ID"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateI18NAndLabelTranslationErrors(t *testing.T) {
	source := `app Warehouse

i18n {
  default fr
  locales tr, tr
}

label Product.name {
  en "Product Name"
  en "Name"
}

label Product.missing {
  tr "Eksik"
  tr "Tekrar"
  de "Fehlt"
}

entity Product {
  name text
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
	for _, code := range []string{"DUPLICATE_LOCALE", "UNKNOWN_DEFAULT_LOCALE", "MISSING_DEFAULT_LABEL_TRANSLATION", "DUPLICATE_LABEL_LOCALE", "UNKNOWN_LABEL_TARGET", "UNKNOWN_LABEL_LOCALE"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateUILabelTranslationTargets(t *testing.T) {
	source := `app Warehouse

i18n {
  default tr
  locales tr, en
}

label app.title {
  tr "Depo"
  en "Warehouse"
}

label page.Products {
  tr "Ürünler"
  en "Products"
}

label action.create {
  tr "Oluştur"
  en "Create"
}

label action.new.Product {
  tr "Yeni Ürün"
  en "New Product"
}

label table.search {
  tr "Ara"
  en "Search"
}

label status.active {
  tr "Aktif"
  en "Active"
}

label app.unsupported {
  tr "Hatalı"
  en "Invalid"
}

label page.Missing {
  tr "Eksik"
  en "Missing"
}

label action.new.Missing {
  tr "Eksik"
  en "Missing"
}

entity Product {
  name text
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

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	codes := map[string]int{}
	for _, diagnostic := range diagnostics {
		codes[diagnostic.Code]++
	}
	if codes["UNKNOWN_LABEL_TARGET"] != 2 || codes["INVALID_LABEL_TARGET"] != 1 {
		t.Fatalf("expected unsupported UI label target diagnostics, got %#v", diagnostics)
	}
}

func TestValidateFieldTextTranslationErrors(t *testing.T) {
	source := `app Warehouse

i18n {
  default tr
  locales tr, en
}

placeholder Product.name {
  en "Enter product name"
  en "Name"
}

placeholder Product.name {
  tr "Ürün adını gir"
  en "Enter product name"
}

help Product.inventoryValue {
  tr "Hesaplanan değer"
  en "Computed value"
}

message MissingTarget {
  tr "Hatalı"
  en "Invalid"
}

message Product.missing {
  tr "Eksik"
  en "Missing"
}

message Product.name {
  tr "Geçerli ürün adı gir"
  de "Enter a valid product name"
}

entity Product {
  name text
  computed inventoryValue money = 1 * 2
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
	for _, code := range []string{
		"DUPLICATE_PLACEHOLDER_TARGET",
		"DUPLICATE_PLACEHOLDER_LOCALE",
		"MISSING_DEFAULT_PLACEHOLDER_TRANSLATION",
		"UNSUPPORTED_HELP_TARGET",
		"INVALID_MESSAGE_TARGET",
		"UNKNOWN_MESSAGE_TARGET",
		"UNKNOWN_MESSAGE_LOCALE",
	} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
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

func TestValidateExplicitAPIUpdateHandler(t *testing.T) {
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
  tenantId text required
  sku text required unique
  stock number default 0
  policy tenant tenantId
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body tenantId text required
  body sku text required
  body stock number required min 0
  update Product where sku == body.sku set stock = body.stock
  respond accepted
  webhook
  public
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

func TestValidateExplicitAPIUpdateHandlerErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  tenantId text required
  sku text required
  name text
  stock number default 0
  computed value number = stock * stock
  policy tenant tenantId
}

api BrokenWebhook {
  method GET
  path "/api/webhooks/broken"
  body sku text optional
  body stock number required label "Stock"
  body stock money
  update Product where name == body.sku set tenantId = body.stock, stock = body.missing, value = body.stock
  respond accepted
  webhook
  public
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
	for _, code := range []string{
		"UNSUPPORTED_API_BODY_METHOD",
		"UNSUPPORTED_API_HANDLER_METHOD",
		"DUPLICATE_API_BODY",
		"UNSUPPORTED_API_BODY_MODIFIER",
		"UNBOUNDED_API_UPDATE",
		"OPTIONAL_API_HANDLER_VALUE",
		"MISSING_API_HANDLER_POLICY_SCOPE",
		"UNSUPPORTED_API_HANDLER_POLICY_FIELD",
		"UNKNOWN_API_HANDLER_VALUE",
		"UNSUPPORTED_API_HANDLER_FIELD",
	} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
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

func TestValidateExplicitAPIRouteConflicts(t *testing.T) {
	source := `app Warehouse

api ProductsShadow {
  method GET
  path "/api/products"
  public
}

api ProductDetailShadow {
  method GET
  path "/api/products/{productId}"
  param productId text
  public
}

api ProductSummaryShadow {
  method GET
  path "/api/products/query/summary"
  public
}

api ReportById {
  method GET
  path "/api/reports/{id}"
  param id text
  public
}

api ReportBySlug {
  method GET
  path "/api/reports/{slug}"
  param slug text
  public
}

entity Product {
  name text required
}

query ProductSummary {
  source Product
  aggregate productCount count
}

page Products {
  source Product
  query ProductSummary

  table {
    columns name
  }

  actions create, edit, delete, archive, restore
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if !hasDiagnostic(diagnostics, "DUPLICATE_API_ROUTE") {
		t.Fatalf("expected duplicate API route diagnostics, got %#v", diagnostics)
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

func TestValidateSecurityCORSErrors(t *testing.T) {
	source := `app Warehouse

security {
  cors {
    origins env cors_origins
    credentials maybe
  }
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
	for _, code := range []string{"INVALID_ENV_NAME", "INVALID_CORS_CREDENTIALS"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateTargetErrors(t *testing.T) {
	source := `app Warehouse

target mobile {
  frontend vue
  backend go
  database mongodb
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
	for _, code := range []string{"UNSUPPORTED_TARGET", "UNSUPPORTED_TARGET_FRONTEND", "UNSUPPORTED_TARGET_BACKEND", "UNSUPPORTED_TARGET_DATABASE"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateTargetAllowsPostgresDatabase(t *testing.T) {
	source := `app Warehouse

target web {
  frontend react
  backend node
  database postgres
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == "UNSUPPORTED_TARGET_DATABASE" {
			t.Fatalf("expected postgres target database to be supported, got %#v", diagnostics)
		}
	}
}

func TestValidateTargetAllowsMySQLDatabase(t *testing.T) {
	source := `app Warehouse

target web {
  frontend react
  backend node
  database mysql
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == "UNSUPPORTED_TARGET_DATABASE" {
			t.Fatalf("expected mysql target database to be supported, got %#v", diagnostics)
		}
	}
}

func TestValidateTargetMySQLRejectsMigrations(t *testing.T) {
	source := `app Warehouse

target web {
  frontend react
  backend node
  database mysql
}

entity Product {
  sku text required
}

migration RenameInventory {
  rename entity Inventory to Product
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if !hasDiagnostic(diagnostics, "UNSUPPORTED_TARGET_DATABASE_MIGRATION") {
		t.Fatalf("expected mysql migration diagnostic, got %#v", diagnostics)
	}
}

func TestValidateTargetAllowsAPIWithoutFrontend(t *testing.T) {
	source := `app Warehouse

target api {
  backend node
  database sqlite
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	for _, diagnostic := range diagnostics {
		if strings.HasPrefix(diagnostic.Code, "MISSING_TARGET") || strings.HasPrefix(diagnostic.Code, "UNSUPPORTED_TARGET") {
			t.Fatalf("expected api target without frontend to be supported, got %#v", diagnostics)
		}
	}
}

func TestValidateTargetAPIRejectsFrontendAndBrowserTests(t *testing.T) {
	source := `app Warehouse

target api {
  frontend react
  backend node
  database sqlite
}

entity Product {
  name text required
}

page Products {
  source Product
  table {
    columns name
  }
}

test ProductsBrowser {
  page Products
  expect text "Products"
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
	for _, code := range []string{"UNSUPPORTED_TARGET_FRONTEND", "UNSUPPORTED_API_TARGET_TEST"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateTargetRequiresStack(t *testing.T) {
	source := `app Warehouse

target web {
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
	for _, code := range []string{"MISSING_TARGET_FRONTEND", "MISSING_TARGET_BACKEND", "MISSING_TARGET_DATABASE"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateDeployErrors(t *testing.T) {
	source := `app Warehouse

deploy {
  target zip
  port env port default 70000
  env DATABASE_URL maybe
  env DATABASE_URL required
  preview staging
  rollback keep 0
  cloud unknown app env cloud_app region env 1REGION
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
	for _, code := range []string{"UNSUPPORTED_DEPLOY_TARGET", "INVALID_ENV_NAME", "INVALID_DEPLOY_PORT_DEFAULT", "UNSUPPORTED_DEPLOY_ENV_MODE", "DUPLICATE_DEPLOY_ENV", "UNSUPPORTED_DEPLOY_PREVIEW", "INVALID_DEPLOY_ROLLBACK_KEEP", "DEPLOY_CLOUD_REQUIRES_DOCKER", "UNSUPPORTED_DEPLOY_CLOUD_PROVIDER"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func TestValidateOpsDiagnostics(t *testing.T) {
	source := `app Warehouse

ops {
  health path "/api/health"
  readiness path "/readyz"
  metrics path "/readyz"
  logging traces
  observe otel endpoint env observe_endpoint
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	codes := diagnosticCodes(diagnostics)
	for _, code := range []string{"INVALID_OPS_PATH", "DUPLICATE_OPS_PATH", "UNSUPPORTED_OPS_LOGGING", "UNSUPPORTED_OPS_OBSERVE_PROVIDER", "INVALID_ENV_NAME"} {
		if !codes[code] {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}

	emptyProgram, emptyParseDiagnostics := Parse("test.black", `app Warehouse

ops {
}
`)
	if len(emptyParseDiagnostics) != 0 {
		t.Fatalf("expected no empty ops parse diagnostics, got %#v", emptyParseDiagnostics)
	}
	emptyCodes := diagnosticCodes(Validate(emptyProgram))
	if !emptyCodes["MISSING_OPS_SIGNAL"] {
		t.Fatalf("expected MISSING_OPS_SIGNAL, got %#v", Validate(emptyProgram))
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

func TestValidateMediaFieldConstraints(t *testing.T) {
	source := `app Warehouse

entity Product {
  photo image optional accept "image/*"
  manual file optional accept "application/pdf"
}

page Products {
  source Product
  table {
    columns photo, manual
  }
  form {
    fields photo, manual
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

func TestValidateMediaFieldConstraintDiagnostics(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text accept "image/*"
  photo image accept "application/pdf"
  manual file accept
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	for _, code := range []string{"UNSUPPORTED_ACCEPT_CONSTRAINT", "UNSUPPORTED_IMAGE_ACCEPT", "MISSING_ACCEPT_VALUE"} {
		if !hasDiagnostic(diagnostics, code) {
			t.Fatalf("expected diagnostic %s in %#v", code, diagnostics)
		}
	}
}

func TestValidateActionMediaInputIsUnsupported(t *testing.T) {
	source := `app Warehouse

entity Product {
  photo image optional
}

action AttachPhoto {
  source Product
  input upload image required
  set photo = upload
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if !hasDiagnostic(diagnostics, "UNSUPPORTED_ACTION_INPUT_TYPE") {
		t.Fatalf("expected unsupported action media input diagnostic, got %#v", diagnostics)
	}
}

func TestValidateJobDeclaration(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
  stock number default 0
}

query LowStockProducts {
  source Product
  where stock < 10
  sort stock asc
  limit 50
}

job LowStockMonitor {
  schedule every 15 minutes
  run query LowStockProducts
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

func TestValidateJobErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
  stock number default 0
}

query LowStockProducts {
  source Product
  where stock < 10
}

job lowStockMonitor {
  schedule every 0 weeks
  run query MissingQuery
}

job LowStockProducts {
  schedule every 15 minutes
  run query LowStockProducts
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	for _, code := range []string{"INVALID_JOB_NAME", "UNSUPPORTED_JOB_SCHEDULE_UNIT", "UNKNOWN_JOB_QUERY", "JOB_NAME_COLLISION"} {
		if !hasDiagnostic(diagnostics, code) {
			t.Fatalf("expected validation code %s, got %#v", code, diagnostics)
		}
	}
}

func hasDiagnostic(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}
