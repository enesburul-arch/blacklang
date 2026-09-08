package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const transactionSource = `app Inventory

entity Product {
  sku text required unique
  stock number default 0
}

action RestockProduct {
  source Product
  input quantity number required min 1
  set stock = stock + quantity
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body sku text required
  body stock number required min 0
  update Product where sku == body.sku set stock = body.stock
  respond accepted
  webhook
  public
}

transaction RestockAtomic {
  action RestockProduct
  api StockWebhook
}
`

func TestParseTransactionBlock(t *testing.T) {
	program := parseTransactionProgram(t, transactionSource)
	if len(program.Transactions) != 1 {
		t.Fatalf("expected one transaction, got %#v", program.Transactions)
	}
	transaction := program.Transactions[0]
	if transaction.Name != "RestockAtomic" || len(transaction.Actions) != 1 || transaction.Actions[0].Name != "RestockProduct" || len(transaction.APIs) != 1 || transaction.APIs[0].Name != "StockWebhook" {
		t.Fatalf("unexpected transaction AST: %#v", transaction)
	}
}

func TestParseTransactionDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		source string
		code   string
	}{
		{
			name: "invalid declaration",
			source: `transaction {
  action RestockProduct
}
`,
			code: "INVALID_TRANSACTION_DECLARATION",
		},
		{
			name: "invalid action target",
			source: `transaction RestockAtomic {
  action
}
`,
			code: "INVALID_TRANSACTION_ACTION",
		},
		{
			name: "invalid api target",
			source: `transaction RestockAtomic {
  api
}
`,
			code: "INVALID_TRANSACTION_API",
		},
		{
			name: "duplicate target",
			source: `transaction RestockAtomic {
  action RestockProduct
  action RestockProduct
}
`,
			code: "DUPLICATE_TRANSACTION_TARGET",
		},
		{
			name: "unexpected token",
			source: `transaction RestockAtomic {
  query LowStockProducts
}
`,
			code: "UNEXPECTED_TRANSACTION_TOKEN",
		},
		{
			name: "unclosed",
			source: `transaction RestockAtomic {
  action RestockProduct
`,
			code: "UNCLOSED_TRANSACTION",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, diagnostics := Parse("transaction.black", test.source)
			assertTransactionDiagnostic(t, diagnostics, test.code)
		})
	}
}

func TestValidateTransactionDiagnostics(t *testing.T) {
	base := `app Inventory

entity Product {
  sku text required unique
  stock number default 0
}

action RestockProduct {
  source Product
  input quantity number required min 1
  set stock = stock + quantity
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body sku text required
  body stock number required min 0
  update Product where sku == body.sku set stock = body.stock
  respond accepted
  webhook
  public
}

api LowStockReport {
  method GET
  path "/api/reports/low-stock"
  public
}
`
	tests := []struct {
		name   string
		source string
		code   string
	}{
		{
			name: "duplicate transaction",
			source: `transaction RestockAtomic {
  action RestockProduct
}

transaction RestockAtomic {
  api StockWebhook
}
`,
			code: "DUPLICATE_TRANSACTION",
		},
		{
			name: "empty transaction",
			source: `transaction EmptyAtomic {
}
`,
			code: "EMPTY_TRANSACTION",
		},
		{
			name: "unknown action",
			source: `transaction RestockAtomic {
  action MissingAction
}
`,
			code: "UNKNOWN_TRANSACTION_ACTION",
		},
		{
			name: "unknown api",
			source: `transaction RestockAtomic {
  api MissingAPI
}
`,
			code: "UNKNOWN_TRANSACTION_API",
		},
		{
			name: "api without update handler",
			source: `transaction RestockAtomic {
  api LowStockReport
}
`,
			code: "UNSUPPORTED_TRANSACTION_API",
		},
		{
			name: "duplicate binding",
			source: `transaction RestockAtomic {
  action RestockProduct
}

transaction RestockAgain {
  action RestockProduct
}
`,
			code: "DUPLICATE_TRANSACTION_BINDING",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program := parseTransactionProgram(t, base+test.source)
			assertTransactionDiagnostic(t, Validate(program), test.code)
		})
	}
}

func TestBuildWebGeneratesTransactionalExplicitAPI(t *testing.T) {
	source := `app Inventory

target api {
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
  update Product where sku == body.sku set stock = body.stock
  respond accepted
  webhook
  public
}

transaction RestockAtomic {
  api StockWebhook
}
`
	program := parseTransactionProgram(t, source)
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	server := readTransactionGenerated(t, outDir, "src/server.ts")
	contract := readTransactionGenerated(t, outDir, "src/blacklang.contract.test.ts")
	openapiText := readTransactionGenerated(t, outDir, "openapi.json")
	for _, value := range []string{
		`app.post("/api/webhooks/stock", async (req, res) => {`,
		`const apiResult = await prisma.$transaction(async (tx) => {`,
		`const productTxModel = tx.product;`,
		`const existing = await productTxModel.findFirst({ where: { sku: bodyValues.value["sku"] } as any });`,
		`const item = await productTxModel.update({`,
		`res.status(202).json({ api: "StockWebhook", status: "accepted", runtime: "declared", handler: "update", entity: "Product", id: item.id, updated: true });`,
	} {
		if !strings.Contains(server, value) {
			t.Fatalf("expected generated transactional API runtime %q in:\n%s", value, server)
		}
	}
	for _, value := range []string{
		`assert.equal(paths["/api/webhooks/stock"]?.post?.["x-blacklang-transaction"], true)`,
		`assert.equal(paths["/api/webhooks/stock"]?.post?.["x-blacklang-transaction-name"], "RestockAtomic")`,
	} {
		if !strings.Contains(contract, value) {
			t.Fatalf("expected contract assertion %q in:\n%s", value, contract)
		}
	}

	var spec struct {
		Paths map[string]map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal([]byte(openapiText), &spec); err != nil {
		t.Fatal(err)
	}
	operation := spec.Paths["/api/webhooks/stock"]["post"]
	if operation["x-blacklang-transaction"] != true || operation["x-blacklang-transaction-name"] != "RestockAtomic" {
		t.Fatalf("expected RestockAtomic metadata, got %#v", operation)
	}

	analysis, diagnostics := AnalyzeAffected(program, "StockWebhook")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no affected diagnostics, got %#v", diagnostics)
	}
	if !affectedItemsContain(analysis.Transactions, "RestockAtomic") {
		t.Fatalf("expected StockWebhook affected graph to include RestockAtomic, got %#v", analysis.Transactions)
	}
}

func parseTransactionProgram(t *testing.T, source string) Program {
	t.Helper()
	program, diagnostics := Parse("transaction.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	return program
}

func assertTransactionDiagnostic(t *testing.T, diagnostics []Diagnostic, code string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			if diagnostic.File != "transaction.black" || diagnostic.Line < 1 || diagnostic.Column < 1 {
				t.Fatalf("diagnostic lacks source position: %#v", diagnostic)
			}
			return
		}
	}
	t.Fatalf("expected %s, got %#v", code, diagnostics)
}

func readTransactionGenerated(t *testing.T, outDir string, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(outDir, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
