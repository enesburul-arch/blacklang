package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseServiceBlock(t *testing.T) {
	source := `api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  respond accepted
  public
}

service InventoryIntegration {
  api StockWebhook
}
`
	program := parseServiceProgram(t, source)
	if len(program.Services) != 1 {
		t.Fatalf("expected one service, got %#v", program.Services)
	}
	service := program.Services[0]
	if service.Name != "InventoryIntegration" || len(service.APIs) != 1 || service.APIs[0].Name != "StockWebhook" {
		t.Fatalf("expected InventoryIntegration to target StockWebhook, got %#v", service)
	}
}

func TestParseServiceDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		source string
		code   string
	}{
		{
			name: "invalid declaration",
			source: `service {
}`,
			code: "INVALID_SERVICE_DECLARATION",
		},
		{
			name: "invalid api target",
			source: `service InventoryIntegration {
  api StockWebhook extra
}`,
			code: "INVALID_SERVICE_API",
		},
		{
			name: "duplicate api target",
			source: `service InventoryIntegration {
  api StockWebhook
  api StockWebhook
}`,
			code: "DUPLICATE_SERVICE_TARGET",
		},
		{
			name: "unexpected token",
			source: `service InventoryIntegration {
  route StockWebhook
}`,
			code: "UNEXPECTED_SERVICE_TOKEN",
		},
		{
			name: "unclosed",
			source: `service InventoryIntegration {
  api StockWebhook
`,
			code: "UNCLOSED_SERVICE",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, diagnostics := Parse("test.black", test.source)
			assertServiceDiagnostic(t, diagnostics, test.code)
		})
	}
}

func TestValidateServiceDiagnostics(t *testing.T) {
	base := `app Inventory

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

api StockReport {
  method GET
  path "/api/reports/stock"
  respond declared
  private
}
`
	tests := []struct {
		name   string
		source string
		code   string
	}{
		{
			name: "duplicate service",
			source: `service InventoryIntegration {
  api StockWebhook
}

service InventoryIntegration {
  api StockReport
}
`,
			code: "DUPLICATE_SERVICE",
		},
		{
			name: "invalid name",
			source: `service inventoryIntegration {
  api StockWebhook
}
`,
			code: "INVALID_SERVICE_NAME",
		},
		{
			name: "name collision",
			source: `service StockWebhook {
  api StockWebhook
}
`,
			code: "SERVICE_NAME_COLLISION",
		},
		{
			name: "empty service",
			source: `service InventoryIntegration {
}
`,
			code: "EMPTY_SERVICE",
		},
		{
			name: "invalid api name",
			source: `service InventoryIntegration {
  api stockWebhook
}
`,
			code: "INVALID_SERVICE_API",
		},
		{
			name: "unknown api",
			source: `service InventoryIntegration {
  api MissingWebhook
}
`,
			code: "UNKNOWN_SERVICE_API",
		},
		{
			name: "duplicate binding",
			source: `service InventoryIntegration {
  api StockWebhook
}

service ReportingService {
  api StockWebhook
}
`,
			code: "DUPLICATE_SERVICE_BINDING",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program := parseServiceProgram(t, base+test.source)
			assertServiceDiagnostic(t, Validate(program), test.code)
		})
	}
}

func TestBuildWebGeneratesServiceModuleAndOpenAPIMetadata(t *testing.T) {
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

service InventoryIntegration {
  api StockWebhook
}
`
	program := parseServiceProgram(t, source)
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	manifest := readServiceGenerated(t, outDir, filepath.Join("services", "manifest.json"))
	module := readServiceGenerated(t, outDir, filepath.Join("src", "services", "inventory-integration.ts"))
	contract := readServiceGenerated(t, outDir, filepath.Join("src", "blacklang.contract.test.ts"))
	openapiText := readServiceGenerated(t, outDir, "openapi.json")
	if !strings.Contains(module, "export const inventoryIntegrationService") || !strings.Contains(module, `"transaction": "RestockAtomic"`) || !strings.Contains(module, `export type InventoryIntegrationAPIName`) {
		t.Fatalf("expected generated service module metadata, got:\n%s", module)
	}
	if !strings.Contains(manifest, `"name": "InventoryIntegration"`) || !strings.Contains(manifest, `"name": "StockWebhook"`) {
		t.Fatalf("expected generated service manifest metadata, got:\n%s", manifest)
	}
	for _, value := range []string{
		`assert.equal((spec as any)["x-blacklang-services"]?.length, 1);`,
		`assert.ok((spec as any).tags?.some((tag: any) => tag.name === "InventoryIntegration"));`,
		`assert.equal(paths["/api/webhooks/stock"]?.post?.["x-blacklang-service"], "InventoryIntegration");`,
	} {
		if !strings.Contains(contract, value) {
			t.Fatalf("expected contract assertion %q in:\n%s", value, contract)
		}
	}

	var spec struct {
		Tags     []map[string]any                     `json:"tags"`
		Services []map[string]any                     `json:"x-blacklang-services"`
		Paths    map[string]map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal([]byte(openapiText), &spec); err != nil {
		t.Fatal(err)
	}
	operation := spec.Paths["/api/webhooks/stock"]["post"]
	if operation["x-blacklang-service"] != "InventoryIntegration" {
		t.Fatalf("expected InventoryIntegration operation metadata, got %#v", operation)
	}
	if tags, ok := operation["tags"].([]any); !ok || len(tags) != 1 || tags[0] != "InventoryIntegration" {
		t.Fatalf("expected InventoryIntegration operation tag, got %#v", operation["tags"])
	}
	if len(spec.Tags) != 1 || spec.Tags[0]["name"] != "InventoryIntegration" || len(spec.Services) != 1 {
		t.Fatalf("expected service root metadata, got tags=%#v services=%#v", spec.Tags, spec.Services)
	}

	analysis, diagnostics := AnalyzeAffected(program, "InventoryIntegration")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no service affected diagnostics, got %#v", diagnostics)
	}
	if !affectedItemsContain(analysis.Services, "InventoryIntegration") || !affectedItemsContain(analysis.APIs, "StockWebhook") || !affectedItemsContain(analysis.GeneratedFiles, "services/manifest.json") || !affectedItemsContain(analysis.GeneratedFiles, "src/services/inventory-integration.ts") {
		t.Fatalf("expected service affected graph to include API and generated service files, got %#v", analysis)
	}

	apiAnalysis, diagnostics := AnalyzeAffected(program, "StockWebhook")
	if len(diagnostics) != 0 {
		t.Fatalf("expected no API affected diagnostics, got %#v", diagnostics)
	}
	if !affectedItemsContain(apiAnalysis.Services, "InventoryIntegration") {
		t.Fatalf("expected API affected graph to include InventoryIntegration, got %#v", apiAnalysis.Services)
	}
}

func parseServiceProgram(t *testing.T, source string) Program {
	t.Helper()
	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	return program
}

func assertServiceDiagnostic(t *testing.T, diagnostics []Diagnostic, code string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return
		}
	}
	t.Fatalf("expected diagnostic %s, got %#v", code, diagnostics)
}

func readServiceGenerated(t *testing.T, outDir string, relativePath string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(outDir, relativePath))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
