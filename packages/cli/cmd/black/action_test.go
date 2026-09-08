package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const customActionSource = `app Inventory

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

role Admin {
  allow all
}

role Worker {
  allow read Product
  allow update Product stock
  deny read Product price
}

entity Product {
  name text required
  stock number default 0
  price money default 0
  status text default draft
}

query LowStockProducts {
  source Product
  where stock < 10
  sort stock asc
  limit 5
}

action RestockProduct {
  source Product
  input quantity number required min 1 label "Quantity" placeholder "Received units" help "Amount added to current stock"
  set stock = stock + quantity
  set status = "received"
  allow Admin, Worker
  success "Stock updated"
}

transaction RestockAtomic {
  action RestockProduct
}

page Products {
  source Product
  access Admin, Worker

  table {
    columns name, stock, price
  }

  form {
    fields name, stock, price, status
  }

  actions edit, RestockProduct
}

page LowStock {
  source Product
  query LowStockProducts
  access Admin, Worker

  table {
    columns name, stock
  }

  actions RestockProduct
}
`

func TestParseCustomAction(t *testing.T) {
	program, diagnostics := Parse("action.black", customActionSource)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	if len(program.Actions) != 1 {
		t.Fatalf("expected one action, got %#v", program.Actions)
	}
	action := program.Actions[0]
	if action.Name != "RestockProduct" || action.Source != "Product" {
		t.Fatalf("expected RestockProduct source Product, got %#v", action)
	}
	if len(action.Inputs) != 1 || action.Inputs[0].Name != "quantity" || action.Inputs[0].Type != "number" {
		t.Fatalf("expected quantity number input, got %#v", action.Inputs)
	}
	if len(action.Sets) != 2 {
		t.Fatalf("expected two set clauses, got %#v", action.Sets)
	}
	if action.Sets[0].Field != "stock" || action.Sets[0].Expression.Left.Value != "stock" || action.Sets[0].Expression.Operator != "+" || action.Sets[0].Expression.Right == nil || action.Sets[0].Expression.Right.Value != "quantity" {
		t.Fatalf("expected stock increment set, got %#v", action.Sets[0])
	}
	if action.Success != "Stock updated" || len(action.Allow) != 2 || action.Allow[0] != "Admin" || action.Allow[1] != "Worker" {
		t.Fatalf("expected allow roles and success message, got %#v", action)
	}
	if !containsString(program.Pages[0].Actions, "RestockProduct") || !containsString(program.Pages[1].Actions, "RestockProduct") {
		t.Fatalf("expected pages to expose RestockProduct, got %#v and %#v", program.Pages[0].Actions, program.Pages[1].Actions)
	}
	if len(program.Transactions) != 1 || program.Transactions[0].Name != "RestockAtomic" || len(program.Transactions[0].Actions) != 1 || program.Transactions[0].Actions[0].Name != "RestockProduct" {
		t.Fatalf("expected RestockAtomic transaction target, got %#v", program.Transactions)
	}
}

func TestParseActionLevelTransactionDiagnostic(t *testing.T) {
	tests := []struct {
		name   string
		source string
		code   string
	}{
		{
			name: "invalid transaction shape",
			source: `action RestockProduct {
  source Product
  transaction atomic
  set stock = stock
}
`,
			code: "UNSUPPORTED_ACTION_TRANSACTION",
		},
		{
			name: "standalone transaction",
			source: `action RestockProduct {
  source Product
  transaction
  set stock = stock
}
`,
			code: "UNSUPPORTED_ACTION_TRANSACTION",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, diagnostics := Parse("action.black", test.source)
			assertActionDiagnostic(t, diagnostics, test.code)
		})
	}
}

func TestValidateCustomAction(t *testing.T) {
	program := parseActionProgram(t, customActionSource)
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
}

func TestValidateCustomActionDiagnostics(t *testing.T) {
	base := `app Inventory

auth {
  strategy emailPassword
  session cookie

  user {
    email email required unique
  }
}

role Admin {
  allow all
}

entity Supplier {
  name text
}

entity Product {
  stock number default 0
  price money default 0
  name text
  supplier Supplier
  computed inventoryValue money = stock * price
}

`

	tests := []struct {
		name   string
		source string
		code   string
	}{
		{
			name: "missing source",
			source: `action RestockProduct {
  set stock = 1
}
`,
			code: "MISSING_ACTION_SOURCE",
		},
		{
			name: "duplicate input",
			source: `action RestockProduct {
  source Product
  input quantity number
  input quantity number
  set stock = stock + quantity
}
`,
			code: "DUPLICATE_ACTION_INPUT",
		},
		{
			name: "input collides with source field",
			source: `action RestockProduct {
  source Product
  input stock number
  set price = stock
}
`,
			code: "ACTION_INPUT_FIELD_COLLISION",
		},
		{
			name: "unsupported input type",
			source: `action RestockProduct {
  source Product
  input supplier Supplier
  set stock = stock
}
`,
			code: "UNSUPPORTED_ACTION_INPUT_TYPE",
		},
		{
			name: "unknown set field",
			source: `action RestockProduct {
  source Product
  set missing = stock
}
`,
			code: "UNKNOWN_ACTION_FIELD",
		},
		{
			name: "computed set field",
			source: `action RestockProduct {
  source Product
  set inventoryValue = stock
}
`,
			code: "UNSUPPORTED_ACTION_FIELD",
		},
		{
			name: "relation set field",
			source: `action RestockProduct {
  source Product
  set supplier = supplier
}
`,
			code: "UNSUPPORTED_ACTION_FIELD",
		},
		{
			name: "unknown operand",
			source: `action RestockProduct {
  source Product
  set stock = missing
}
`,
			code: "UNKNOWN_ACTION_VALUE",
		},
		{
			name: "type mismatch",
			source: `action RestockProduct {
  source Product
  set name = stock
}
`,
			code: "ACTION_VALUE_TYPE_MISMATCH",
		},
		{
			name: "division by zero",
			source: `action RestockProduct {
  source Product
  set stock = stock / 0
}
`,
			code: "INVALID_ACTION_EXPRESSION",
		},
		{
			name: "unknown allow role",
			source: `action RestockProduct {
  source Product
  set stock = stock
  allow Missing
}
`,
			code: "UNKNOWN_ACTION_ALLOW_ROLE",
		},
		{
			name: "query name collision",
			source: `query RestockProduct {
  source Product
}

action RestockProduct {
  source Product
  set stock = stock
}
`,
			code: "ACTION_NAME_COLLISION",
		},
		{
			name: "page source mismatch",
			source: `entity Order {
  stock number
}

action RestockProduct {
  source Product
  set stock = stock
}

page Orders {
  source Order
  actions RestockProduct
}
`,
			code: "PAGE_ACTION_SOURCE_MISMATCH",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program := parseActionProgram(t, base+test.source)
			assertActionDiagnostic(t, Validate(program), test.code)
		})
	}
}

func TestFormatCustomActionIRInspectAndAffected(t *testing.T) {
	program := parseActionProgram(t, customActionSource)
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}

	ir := FormatBlackIR(program)
	for _, value := range []string{
		"action RestockProduct source Product",
		`input quantity number required min 1 label "Quantity" placeholder "Received units" help "Amount added to current stock"`,
		"set stock = stock + quantity",
		`set status = "received"`,
		"allow Admin Worker",
		`success "Stock updated"`,
		"transaction RestockAtomic",
		"  action RestockProduct",
		"actions edit RestockProduct",
	} {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected BlackIR to contain %q, got:\n%s", value, ir)
		}
	}

	inspectIR := FormatInspectIR(InspectResult{
		Success: true,
		Command: "inspect",
		Version: version,
		Summary: Summary{
			App:      "Inventory",
			Entities: 1,
			Pages:    2,
		},
		Program: program,
		Errors:  []Diagnostic{},
	})
	for _, value := range []string{
		"actions 1",
		"action RestockProduct source Product inputs 1 sets 2 transaction RestockAtomic",
		"transactions 1",
		"transaction RestockAtomic actions 1 apis 0",
		"page Products source Product actions edit RestockProduct",
		"page LowStock source Product actions RestockProduct",
	} {
		if !strings.Contains(inspectIR, value) {
			t.Fatalf("expected inspect IR to contain %q, got:\n%s", value, inspectIR)
		}
	}

	analysis, diagnostics := AnalyzeAffected(program, "RestockProduct")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no affected diagnostics, got %#v", diagnostics)
	}
	if !analysis.Found || analysis.Kind != "action" || !affectedItemsContain(analysis.Actions, "RestockProduct") {
		t.Fatalf("expected RestockProduct action analysis, got %#v", analysis)
	}
	if !affectedItemsContain(analysis.Transactions, "RestockAtomic") {
		t.Fatalf("expected RestockAtomic to affect RestockProduct, got %#v", analysis.Transactions)
	}
	for _, name := range []string{"Products", "LowStock"} {
		if !affectedItemsContain(analysis.Pages, name) {
			t.Fatalf("expected affected page %s, got %#v", name, analysis.Pages)
		}
	}
	for _, file := range []string{"src/types.ts", "src/validation/product.ts", "src/routes/product.ts", "src/routes/product.lowstock.ts", "src/pages/LowStockPage.tsx", "openapi.json"} {
		if !affectedItemsContain(analysis.GeneratedFiles, file) {
			t.Fatalf("expected affected generated file %s, got %#v", file, analysis.GeneratedFiles)
		}
	}

	fieldAnalysis, diagnostics := AnalyzeAffected(program, "Product.stock")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no field affected diagnostics, got %#v", diagnostics)
	}
	if !affectedItemsContain(fieldAnalysis.Actions, "RestockProduct") {
		t.Fatalf("expected Product.stock to affect RestockProduct, got %#v", fieldAnalysis.Actions)
	}

	transactionAnalysis, diagnostics := AnalyzeAffected(program, "RestockAtomic")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no transaction affected diagnostics, got %#v", diagnostics)
	}
	if !transactionAnalysis.Found || transactionAnalysis.Kind != "transaction" || !affectedItemsContain(transactionAnalysis.Actions, "RestockProduct") {
		t.Fatalf("expected RestockAtomic transaction analysis, got %#v", transactionAnalysis)
	}
}

func TestBuildWebGeneratesCustomActionRuntimeAndUI(t *testing.T) {
	program := parseActionProgram(t, customActionSource)
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	route := readActionGenerated(t, outDir, "src/routes/product.ts")
	queryRoute := readActionGenerated(t, outDir, "src/routes/product.lowstock.ts")
	client := readActionGenerated(t, outDir, "src/api/product.ts")
	queryClient := readActionGenerated(t, outDir, "src/api/product.lowstock.ts")
	page := readActionGenerated(t, outDir, "src/pages/ProductsPage.tsx")
	queryPage := readActionGenerated(t, outDir, "src/pages/LowStockPage.tsx")
	types := readActionGenerated(t, outDir, "src/types.ts")
	validation := readActionGenerated(t, outDir, "src/validation/product.ts")
	app := readActionGenerated(t, outDir, "src/App.tsx")
	contractTest := readActionGenerated(t, outDir, "src/blacklang.contract.test.ts")
	apiTest := readActionGenerated(t, outDir, "src/blacklang.api.test.ts")
	openapiText := readActionGenerated(t, outDir, "openapi.json")

	for _, pair := range [][2]string{
		{route, `post("/products/:id/actions/restockproduct"`},
		{route, `const actionResult = await prisma.$transaction(async (tx) => {`},
		{route, `const productTxModel = tx.product;`},
		{route, `validateRestockProductInput(req.body)`},
		{route, `canAccessField(currentRoles(req), "update", "Product", "stock")`},
		{route, `await tx.$executeRaw`},
		{route, `INSERT INTO "BlackAuditLog"`},
		{route, `VALUES (${crypto.randomUUID()}, ${actor.id}, ${actor.role}, ${"action.RestockProduct"}, ${"Product"}, ${item.id}, ${"Stock updated"})`},
		{queryRoute, `post("/lowstock/:id/actions/restockproduct"`},
		{queryRoute, `const actionResult = await prisma.$transaction(async (tx) => {`},
		{client, `runRestockProduct: (id: string, input: RestockProductInput)`},
		{client, `"/actions/restockproduct"`},
		{queryClient, `runRestockProduct: (id: string, input: RestockProductInput)`},
		{page, `const canRunRestockProduct = permissions.customActions.RestockProduct !== false;`},
		{page, `function openRestockProductAction(item: Product)`},
		{page, `async function runRestockProductAction(event: FormEvent<HTMLFormElement>)`},
		{page, `productApi.runRestockProduct(restockProductTarget.id, input)`},
		{page, `Restock Product`},
		{queryPage, `productApi.runRestockProduct(restockProductTarget.id, input)`},
		{queryPage, `refreshQuery();`},
		{types, `export type RestockProductInput = {`},
		{types, `quantity: number;`},
		{validation, `export function validateRestockProductInput(input: any)`},
		{validation, `if (parsed < 1) errors.push("RestockProduct.quantity must be at least 1");`},
		{app, `type CustomActionPermission = { name: string; allow: string[]; writes: string[] };`},
		{app, `pagePermissions("Product", ["name", "stock", "price", "status"], [{ name: "RestockProduct", allow: ["Admin", "Worker"], writes: ["stock", "status"] }])`},
		{contractTest, `assertPath("/api/products/{id}/actions/restockproduct", "post")`},
		{contractTest, `assert.equal(paths["/api/products/{id}/actions/restockproduct"]?.post?.["x-blacklang-action"], "RestockProduct")`},
		{contractTest, `assert.equal(paths["/api/products/{id}/actions/restockproduct"]?.post?.["x-blacklang-transaction"], true)`},
		{contractTest, `assert.equal(paths["/api/products/{id}/actions/restockproduct"]?.post?.["x-blacklang-transaction-name"], "RestockAtomic")`},
		{contractTest, `assertValid(validateRestockProductInput(validRestockProductInput), "RestockProduct valid input")`},
		{apiTest, `const { createApp } = await import("./server");`},
		{apiTest, `assert.equal(api.status, 401, "authenticated API routes should reject anonymous requests");`},
	} {
		if !strings.Contains(pair[0], pair[1]) {
			t.Fatalf("missing generated custom action contract %q in:\n%s", pair[1], pair[0])
		}
	}

	var spec struct {
		Paths      map[string]map[string]map[string]any `json:"paths"`
		Components struct {
			Schemas map[string]any `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal([]byte(openapiText), &spec); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/products/{id}/actions/restockproduct", "/api/lowstock/{id}/actions/restockproduct"} {
		operation := spec.Paths[path]["post"]
		if operation["x-blacklang-action"] != "RestockProduct" {
			t.Fatalf("expected RestockProduct OpenAPI action for %s, got %#v", path, operation)
		}
		if operation["x-blacklang-transaction"] != true {
			t.Fatalf("expected RestockProduct OpenAPI transaction metadata for %s, got %#v", path, operation)
		}
		if operation["x-blacklang-transaction-name"] != "RestockAtomic" {
			t.Fatalf("expected RestockAtomic OpenAPI transaction name for %s, got %#v", path, operation)
		}
	}
	if _, ok := spec.Components.Schemas["RestockProductInput"]; !ok {
		t.Fatalf("expected RestockProductInput schema, got %#v", spec.Components.Schemas)
	}
	if _, ok := spec.Components.Schemas["RestockProductResponse"]; !ok {
		t.Fatalf("expected RestockProductResponse schema, got %#v", spec.Components.Schemas)
	}
}

func TestGeneratedCustomActionRouteExecutesValidatedMutation(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node.js unavailable; generated runtime execution requires Node.js")
	}
	program := parseActionProgram(t, customActionSource)
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	g := webGenerator{program: program}
	route := g.customActionRoutes(program.Pages[0], program.Entities[0])
	route = strings.ReplaceAll(route, ": Record<string, unknown>", "")
	route = strings.ReplaceAll(route, " as any", "")
	route = strings.ReplaceAll(route, " as const", "")

	script := `const assert = require("node:assert/strict");
let handler, guard, updates = 0, transactionCount = 0, audits = [], fieldAccess = true, checkedFields = [];
let row = { id: "p1", name: "Widget", stock: 5, price: 9, status: "draft", archivedAt: null };
const productModel = {
  async findUnique(args) {
    assert.deepEqual(args, { where: { id: "p1" } });
    return row;
  },
  async update(args) {
    updates++;
    assert.deepEqual(args.where, { id: "p1" });
    row = { ...row, ...args.data };
    return row;
  }
};
const crypto = { randomUUID: () => "audit-id" };
const prisma = {
  async $transaction(callback) {
    transactionCount++;
    return callback({
      product: productModel,
      async $executeRaw(_strings, ...values) {
        audits.push(values);
      }
    });
  }
};
const productRouter = {
  post(path, middleware, callback) {
    assert.equal(path, "/products/:id/actions/restockproduct");
    guard = middleware;
    handler = callback;
  }
};
function requirePermission(action, resource) {
  assert.equal(action, "update");
  assert.equal(resource, "Product");
  return "updateGuard";
}
function currentUser(req) { return req.user; }
function currentRoles(req) { const user = currentUser(req); return user ? (user.roles ?? [user.role]) : []; }
function canAccessField(_roles, action, resource, field) {
  assert.equal(action, "update");
  assert.equal(resource, "Product");
  checkedFields.push(field);
  return fieldAccess;
}
function validateRestockProductInput(input) {
  const quantity = Number(input.quantity);
  if (!Number.isInteger(quantity) || quantity < 1) {
    return { valid: false, errors: ["RestockProduct.quantity must be at least 1"] };
  }
  return { valid: true, value: { quantity } };
}
function sanitizeProduct(item, roles) { return { ...item, sanitizedFor: roles.join(",") }; }
function writeAuditLog(user, action, resource, resourceId, summary) {
  audits.push({ user, action, resource, resourceId, summary });
}
` + route + `
async function run(user, body) {
  const response = {
    statusCode: 200,
    body: null,
    status(value) { this.statusCode = value; return this; },
    json(value) { this.body = value; }
  };
  await handler({ params: { id: "p1" }, body, user }, response);
  return response;
}

(async () => {
  assert.equal(guard, "updateGuard");
  let response = await run({ id: "u1", role: "Worker" }, { quantity: 7 });
  assert.equal(response.statusCode, 200);
  assert.equal(response.body.item.stock, 12);
  assert.equal(response.body.item.status, "received");
  assert.equal(response.body.item.sanitizedFor, "Worker");
  assert.equal(response.body.message, "Stock updated");
  assert.deepEqual(audits, [["audit-id", "u1", "Worker", "action.RestockProduct", "Product", "p1", "Stock updated"]]);
  assert.equal(updates, 1);
  assert.equal(transactionCount, 1);
  assert.deepEqual(checkedFields.slice(0, 2), ["stock", "status"]);

  response = await run({ id: "u1", role: "Worker" }, { quantity: 0 });
  assert.equal(response.statusCode, 400);
  assert.equal(updates, 1);
  assert.equal(transactionCount, 1);

  response = await run({ id: "u2", role: "Guest" }, { quantity: 3 });
  assert.equal(response.statusCode, 403);
  assert.equal(updates, 1);
  assert.equal(transactionCount, 1);

  fieldAccess = false;
  response = await run({ id: "u1", role: "Worker" }, { quantity: 3 });
  assert.equal(response.statusCode, 403);
  assert.equal(updates, 1);
  assert.equal(transactionCount, 1);

  fieldAccess = true;
  row = null;
  response = await run({ id: "u1", role: "Worker" }, { quantity: 3 });
  assert.equal(response.statusCode, 404);
  assert.equal(updates, 1);
  assert.equal(transactionCount, 2);
  console.log("custom action runtime passed");
})().catch((error) => { console.error(error); process.exitCode = 1; });
`
	path := filepath.Join(t.TempDir(), "action-runtime.cjs")
	if err := os.WriteFile(path, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(node, path).CombinedOutput(); err != nil {
		t.Fatalf("generated custom action runtime failed: %v\n%s", err, output)
	}
}

func parseActionProgram(t *testing.T, source string) Program {
	t.Helper()
	program, diagnostics := Parse("action.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	return program
}

func assertActionDiagnostic(t *testing.T, diagnostics []Diagnostic, code string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			if diagnostic.File != "action.black" || diagnostic.Line < 1 || diagnostic.Column < 1 {
				t.Fatalf("diagnostic lacks source position: %#v", diagnostic)
			}
			return
		}
	}
	t.Fatalf("expected %s, got %#v", code, diagnostics)
}

func readActionGenerated(t *testing.T, outDir string, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(outDir, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
