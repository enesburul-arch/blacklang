package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseComputedExpressionTreeWithPrecedence(t *testing.T) {
	source := `app Warehouse

entity Product {
  stock number
  price money
  tax money
  computed inventoryValue money = stock * (price + tax) label "Inventory Value"
}

page Products {
  source Product
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}

	computed := program.Entities[0].ComputedFields[0]
	if computed.Expression.Tree == nil {
		t.Fatalf("expected computed expression tree, got %#v", computed.Expression)
	}
	if computed.Expression.Operator != "*" || computed.Expression.Right != "price + tax" {
		t.Fatalf("expected legacy root expression fields to summarize tree, got %#v", computed.Expression)
	}
	ir := FormatBlackIR(program)
	if !strings.Contains(ir, `computed inventoryValue money = stock * (price + tax) label Inventory Value`) {
		t.Fatalf("expected BlackIR to preserve grouped expression, got:\n%s", ir)
	}
}

func TestBuildWebGeneratesNestedActionExpressionRuntime(t *testing.T) {
	source := `app Inventory

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  stock number default 0
}

action RestockProduct {
  source Product
  input quantity number required
  input packSize number required
  set stock = stock + quantity / packSize
}

page Products {
  source Product
  actions RestockProduct
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}
	route, err := os.ReadFile(filepath.Join(outDir, "src", "routes", "product.ts"))
	if err != nil {
		t.Fatalf("expected generated product route: %v", err)
	}
	text := string(route)
	for _, expected := range []string{
		`const restockProductLogicDivisor1 = Number(actionInput.packSize ?? 0);`,
		`if (restockProductLogicDivisor1 === 0) {`,
		`data.stock = Number(existing.stock ?? 0) + Number(actionInput.quantity ?? 0) / Number(actionInput.packSize ?? 0);`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected generated action route to contain %q, got:\n%s", expected, text)
		}
	}
}

func TestBuildWebGeneratesNestedAPIExpressionRuntime(t *testing.T) {
	source := `app Inventory

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  stock number default 0
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body sku text required
  body quantity number required
  body packSize number required
  update Product where sku == body.sku set stock = stock + body.quantity / body.packSize
  respond updated
}

page Products {
  source Product
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}
	server, err := os.ReadFile(filepath.Join(outDir, "src", "server.ts"))
	if err != nil {
		t.Fatalf("expected generated server: %v", err)
	}
	text := string(server)
	for _, expected := range []string{
		`const stockWebhookLogicDivisor1 = Number(bodyValues.value["packSize"] ?? 0);`,
		`if (stockWebhookLogicDivisor1 === 0) {`,
		`data.stock = Number(existing.stock ?? 0) + Number(bodyValues.value["quantity"] ?? 0) / Number(bodyValues.value["packSize"] ?? 0);`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected generated API route to contain %q, got:\n%s", expected, text)
		}
	}
}

func TestBuildWebGeneratesActionValueIfElseRuntime(t *testing.T) {
	source := `app Inventory

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  stock number default 0
  price money default 0
  status text default "standard"
}

action PriceProduct {
  source Product
  input quantity number required
  value subtotal = price * quantity
  if subtotal > 1000
    set status = "bulk"
    set stock = stock + quantity
  else
    set status = "standard"
}

page Products {
  source Product
  actions PriceProduct
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	if len(program.Actions[0].Statements) != 2 {
		t.Fatalf("expected value and if statements, got %#v", program.Actions[0].Statements)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}
	route, err := os.ReadFile(filepath.Join(outDir, "src", "routes", "product.ts"))
	if err != nil {
		t.Fatalf("expected generated product route: %v", err)
	}
	text := string(route)
	for _, expected := range []string{
		`const subtotal = Number(existing.price ?? 0) * Number(actionInput.quantity ?? 0);`,
		`if (subtotal > 1000) {`,
		`data.status = "bulk";`,
		`data.stock = Number(existing.stock ?? 0) + Number(actionInput.quantity ?? 0);`,
		`} else {`,
		`data.status = "standard";`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected generated action route to contain %q, got:\n%s", expected, text)
		}
	}
	ir := FormatBlackIR(program)
	for _, expected := range []string{
		`value subtotal = price * quantity`,
		`if subtotal > 1000`,
		`set status = "bulk"`,
		`else`,
		`set status = "standard"`,
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected BlackIR to contain %q, got:\n%s", expected, ir)
		}
	}
}

func TestBuildWebGeneratesAPIValueIfElseRuntime(t *testing.T) {
	source := `app Inventory

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  stock number default 0
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body sku text required
  body quantity number required
  body packSize number required
  update Product where sku == body.sku {
    value incoming = body.quantity / body.packSize
    if incoming > 0
      set stock = stock + incoming
    else
      set stock = stock
  }
  respond updated
}

page Products {
  source Product
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	if program.APIs[0].Update == nil || len(program.APIs[0].Update.Statements) != 2 {
		t.Fatalf("expected API update value and if statements, got %#v", program.APIs[0].Update)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}
	server, err := os.ReadFile(filepath.Join(outDir, "src", "server.ts"))
	if err != nil {
		t.Fatalf("expected generated server: %v", err)
	}
	text := string(server)
	for _, expected := range []string{
		`const stockWebhookLogicDivisor1 = Number(bodyValues.value["packSize"] ?? 0);`,
		`const incoming = Number(bodyValues.value["quantity"] ?? 0) / Number(bodyValues.value["packSize"] ?? 0);`,
		`if (incoming > 0) {`,
		`data.stock = Number(existing.stock ?? 0) + Number(incoming ?? 0);`,
		`} else {`,
		`data.stock = existing.stock;`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected generated API route to contain %q, got:\n%s", expected, text)
		}
	}
	metadata := openapiAPIUpdateMetadata(*program.APIs[0].Update)
	sets := metadata["sets"].([]any)
	if len(sets) != 1 || sets[0] != "stock" {
		t.Fatalf("expected OpenAPI metadata to include unique stock set, got %#v", metadata)
	}
}

func TestValidateRejectsReservedLogicValueNamesCaseInsensitive(t *testing.T) {
	source := `app Inventory

entity Product {
  sku text required unique
  stock number default 0
}

action BadAction {
  source Product
  value If = stock
  set stock = stock
}

api BadAPI {
  method POST
  path "/api/bad"
  body sku text required
  update Product where sku == body.sku {
    value Else = stock
    set stock = stock
  }
  respond updated
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	expectedCodes := map[string]bool{
		"INVALID_ACTION_VALUE_NAME":      false,
		"INVALID_API_HANDLER_VALUE_NAME": false,
	}
	for _, diagnostic := range diagnostics {
		if _, ok := expectedCodes[diagnostic.Code]; ok {
			expectedCodes[diagnostic.Code] = true
		}
	}
	for code, found := range expectedCodes {
		if !found {
			t.Fatalf("expected diagnostic %s, got %#v", code, diagnostics)
		}
	}
}

func TestBuildWebNarrowsAPIDirectNumericLocalValue(t *testing.T) {
	source := `app Inventory

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  stock number default 0
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body sku text required
  body stock number required min 0
  update Product where sku == body.sku {
    value incomingStock = body.stock
    if incomingStock >= 0
      set stock = incomingStock
    else
      set stock = stock
  }
  respond accepted
  webhook
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}
	server, err := os.ReadFile(filepath.Join(outDir, "src", "server.ts"))
	if err != nil {
		t.Fatalf("expected generated server: %v", err)
	}
	text := string(server)
	for _, expected := range []string{
		`const incomingStock = Number(bodyValues.value["stock"] ?? 0);`,
		`if (incomingStock >= 0) {`,
		`data.stock = incomingStock;`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected generated API route to contain %q, got:\n%s", expected, text)
		}
	}
}

func TestBuildWebNarrowsActionSourceNumericLocalValue(t *testing.T) {
	source := `app Inventory

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  stock number default 0
  price money
}

action PriceProduct {
  source Product
  value currentPrice = price
  if currentPrice > 0
    set stock = stock
  else
    set stock = stock
}

page Products {
  source Product
  actions PriceProduct
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}
	route, err := os.ReadFile(filepath.Join(outDir, "src", "routes", "product.ts"))
	if err != nil {
		t.Fatalf("expected generated product route: %v", err)
	}
	text := string(route)
	for _, expected := range []string{
		`const currentPrice = Number(existing.price ?? 0);`,
		`if (currentPrice > 0) {`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected generated action route to contain %q, got:\n%s", expected, text)
		}
	}
}

func TestBuildWebGeneratesActionCompoundIfRuntime(t *testing.T) {
	source := `app Inventory

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  stock number default 0
  price money default 0
  status text default "standard"
}

action PriceProduct {
  source Product
  input quantity number required
  value subtotal = price * quantity
  if subtotal > 1000 and not stock == 0
    set status = "bulk"
  else
    set status = "standard"
}

page Products {
  source Product
  actions PriceProduct
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	condition := program.Actions[0].Statements[1].If.Condition
	if condition.Tree == nil || condition.Tree.Kind != "and" {
		t.Fatalf("expected compound action condition tree, got %#v", condition)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}
	route, err := os.ReadFile(filepath.Join(outDir, "src", "routes", "product.ts"))
	if err != nil {
		t.Fatalf("expected generated product route: %v", err)
	}
	text := string(route)
	for _, expected := range []string{
		`if ((subtotal > 1000 && !(existing.stock == 0))) {`,
		`data.status = "bulk";`,
		`} else {`,
		`data.status = "standard";`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected generated action route to contain %q, got:\n%s", expected, text)
		}
	}
	ir := FormatBlackIR(program)
	if !strings.Contains(ir, `if subtotal > 1000 and not stock == 0`) {
		t.Fatalf("expected BlackIR to preserve compound action condition, got:\n%s", ir)
	}
}

func TestBuildWebGeneratesAPICompoundIfRuntime(t *testing.T) {
	source := `app Inventory

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  stock number default 0
  status text default "standard"
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body sku text required
  body stock number required min 0
  body status text required
  update Product where sku == body.sku {
    value incomingStock = body.stock
    if incomingStock >= 0 and (body.status == "active" or stock == 0)
      set stock = incomingStock
    else
      set stock = stock
  }
  respond accepted
  webhook
}

page Products {
  source Product
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	condition := program.APIs[0].Update.Statements[1].If.Condition
	if condition.Tree == nil || condition.Tree.Kind != "and" {
		t.Fatalf("expected compound API condition tree, got %#v", condition)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}
	server, err := os.ReadFile(filepath.Join(outDir, "src", "server.ts"))
	if err != nil {
		t.Fatalf("expected generated server: %v", err)
	}
	text := string(server)
	for _, expected := range []string{
		`const incomingStock = Number(bodyValues.value["stock"] ?? 0);`,
		`if ((incomingStock >= 0 && (bodyValues.value["status"] == "active" || existing.stock == 0))) {`,
		`data.stock = incomingStock;`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected generated API route to contain %q, got:\n%s", expected, text)
		}
	}
	ir := FormatBlackIR(program)
	if !strings.Contains(ir, `if incomingStock >= 0 and (body.status == "active" or stock == 0)`) {
		t.Fatalf("expected BlackIR to preserve compound API condition, got:\n%s", ir)
	}
}

func TestValidateRejectsCompoundIfTypeMismatch(t *testing.T) {
	source := `app Inventory

entity Product {
  sku text required unique
  stock number default 0
  status text default "standard"
}

action BadAction {
  source Product
  if stock > "high" or status == "active"
    set stock = stock
}

api BadAPI {
  method POST
  path "/api/bad"
  body sku text required
  body status text required
  update Product where sku == body.sku {
    if stock > body.status or stock >= 0
      set stock = stock
  }
  respond updated
}
`

	program, parseDiagnostics := Parse("expression.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	diagnostics := Validate(program)
	expectedCodes := map[string]bool{
		"INCOMPATIBLE_ACTION_CONDITION":      false,
		"INCOMPATIBLE_API_HANDLER_CONDITION": false,
	}
	for _, diagnostic := range diagnostics {
		if _, ok := expectedCodes[diagnostic.Code]; ok {
			expectedCodes[diagnostic.Code] = true
		}
	}
	for code, found := range expectedCodes {
		if !found {
			t.Fatalf("expected diagnostic %s, got %#v", code, diagnostics)
		}
	}
}
