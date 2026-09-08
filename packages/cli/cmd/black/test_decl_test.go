package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBrowserTestDeclaration(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product
  actions create
}

test WarehouseBrowserSmoke {
  page Products
  expect text "Products"
  expect page Products
  expect action create
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	if len(program.Tests) != 1 {
		t.Fatalf("expected one test declaration, got %#v", program.Tests)
	}
	test := program.Tests[0]
	if test.Name != "WarehouseBrowserSmoke" || test.Page != "Products" || len(test.Expectations) != 3 {
		t.Fatalf("unexpected test declaration: %#v", test)
	}
	if got := test.Expectations[0]; got.Kind != "text" || got.Value != "Products" {
		t.Fatalf("expected text expectation, got %#v", got)
	}
}

func TestValidateBrowserTestDeclaration(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

action RestockProduct {
  source Product
  input quantity number required min 1
  set name = "Restocked"
}

page Products {
  source Product
  actions create, RestockProduct
}

test WarehouseBrowserSmoke {
  page Products
  expect text "Products"
  expect page Products
  expect action RestockProduct
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validate diagnostics, got %#v", diagnostics)
	}
}

func TestValidateBrowserTestErrors(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product
  actions create
}

test Products {
  page Missing
  expect text ""
  expect text ""
  expect page Missing
  expect action delete
}

test EmptyBrowserSmoke {
  page Products
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", diagnostics)
	}
	diagnostics = Validate(program)
	for _, code := range []string{
		"TEST_NAME_COLLISION",
		"UNKNOWN_TEST_PAGE",
		"MISSING_TEST_EXPECT",
	} {
		if !hasDiagnostic(diagnostics, code) {
			t.Fatalf("expected diagnostic %s, got %#v", code, diagnostics)
		}
	}
}

func TestBuildWebGeneratesBrowserChecks(t *testing.T) {
	source := `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  name text required
}

page Products {
  source Product
  actions create
}

test WarehouseBrowserSmoke {
  page Products
  expect text "Products"
  expect page Products
  expect action create
}
`

	program, diagnostics := Parse("test.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("parse browser test fixture: %#v", diagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("validate browser test fixture: %#v", diagnostics)
	}
	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("generate browser test fixture: %#v", diagnostics)
	}
	browserTest, err := os.ReadFile(filepath.Join(outDir, "src", "blacklang.browser.test.tsx"))
	if err != nil {
		t.Fatalf("read generated browser test: %v", err)
	}
	browserText := string(browserTest)
	for _, snippet := range []string{
		"BlackLang generated browser checks passed",
		"WarehouseBrowserSmoke",
		`"kind": "action"`,
		`"value": "create"`,
	} {
		if !strings.Contains(browserText, snippet) {
			t.Fatalf("expected generated browser test to contain %q, got:\n%s", snippet, browserText)
		}
	}
	e2eTest, err := os.ReadFile(filepath.Join(outDir, "src", "blacklang.e2e.test.ts"))
	if err != nil {
		t.Fatalf("read generated browser e2e test: %v", err)
	}
	e2eText := string(e2eTest)
	for _, snippet := range []string{
		"BlackLang generated browser e2e tests passed",
		`import { chromium } from "playwright-core";`,
		`const databaseEnvName = "DATABASE_URL";`,
		`const pageLabels: Record<string, string>`,
		`const actionLabelsByPage: Record<string, Record<string, string>>`,
		`await authenticateIfNeeded(page, webBaseUrl);`,
		`await expectActionVisible(page, check, expectation.value);`,
	} {
		if !strings.Contains(e2eText, snippet) {
			t.Fatalf("expected generated browser e2e test to contain %q, got:\n%s", snippet, e2eText)
		}
	}
	matrixTest, err := os.ReadFile(filepath.Join(outDir, "src", "blacklang.e2e.matrix.ts"))
	if err != nil {
		t.Fatalf("read generated browser e2e matrix test: %v", err)
	}
	matrixText := string(matrixTest)
	for _, snippet := range []string{
		`command: "test:e2e:plan"`,
		`command: "test:e2e:matrix"`,
		`BLACKLANG_E2E_CHROME_PATH`,
		`BLACKLANG_E2E_EDGE_PATH`,
		`spawnSync(process.execPath`,
		`stdoutTail`,
	} {
		if !strings.Contains(matrixText, snippet) {
			t.Fatalf("expected generated browser e2e matrix to contain %q, got:\n%s", snippet, matrixText)
		}
	}
	matrixManifest, err := os.ReadFile(filepath.Join(outDir, "tests", "browser-matrix.json"))
	if err != nil {
		t.Fatalf("read generated browser matrix manifest: %v", err)
	}
	for _, snippet := range []string{
		`"planCommand": "npm run test:e2e:plan"`,
		`"runCommand": "npm run test:e2e:matrix"`,
		`"name": "chrome"`,
		`"name": "edge"`,
		`"src/blacklang.e2e.matrix.ts"`,
	} {
		if !strings.Contains(string(matrixManifest), snippet) {
			t.Fatalf("expected browser matrix manifest to contain %q, got:\n%s", snippet, matrixManifest)
		}
	}
	packageJSON, err := os.ReadFile(filepath.Join(outDir, "package.json"))
	if err != nil {
		t.Fatalf("read generated package: %v", err)
	}
	if !strings.Contains(string(packageJSON), "tsx src/blacklang.browser.test.tsx") {
		t.Fatalf("expected package test script to include browser checks, got:\n%s", string(packageJSON))
	}
	for _, snippet := range []string{
		`"test:e2e": "npm run db:generate && tsx src/blacklang.e2e.test.ts"`,
		`"test:e2e:plan": "tsx src/blacklang.e2e.matrix.ts --plan"`,
		`"test:e2e:matrix": "npm run db:generate && tsx src/blacklang.e2e.matrix.ts --run"`,
		`"test:all": "npm test && npm run test:e2e:matrix"`,
		`"playwright-core": "latest"`,
	} {
		if !strings.Contains(string(packageJSON), snippet) {
			t.Fatalf("expected package.json to contain %q, got:\n%s", snippet, packageJSON)
		}
	}
}
