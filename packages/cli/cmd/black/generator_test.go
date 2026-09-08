package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildWebWritesExpectedFiles(t *testing.T) {
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

security {
  cors {
    origins env CORS_ORIGINS
    credentials true
  }
}

deploy {
  target docker
  port env PORT default 3001
  env DATABASE_URL required
  env CORS_ORIGINS optional
  preview local
  rollback keep 3
  cloud fly app env FLY_APP_NAME region env FLY_REGION
}

entity Product {
  sku text required unique length 3..40 regex "^[A-Z0-9]+$" message "Use uppercase letters and numbers"
  name text required label "Product Name" placeholder "Enter product name" help "Visible product name"
  photo image optional accept "image/*" label "Photo" help "Optional local product image"
  stock number default 0 min 0 label "Stock Count" placeholder "Enter stock count"
  price money min 0 label "Unit Price" placeholder "Enter unit price"
  computed inventoryValue money = stock * price label "Inventory Value"
  index stock
  index sku, stock
}

entity Customer {
  name text required
  email email unique
  website text optional url message "Enter a valid website"
}

entity Order {
  customer Customer required label "Customer" placeholder "Select a customer" help "Customer connected to this order"
  total money default 0 min 0 label "Order Total" placeholder "Enter order total" help "Total amount for this order"
  discount money default 0 min 0 label "Discount" placeholder "Enter discount"
  status text default draft label "Order Status" placeholder "Enter order status"
  trackingNumber text optional label "Tracking Number" placeholder "Enter tracking number"
  index customer, status
  validate discount <= total message "Discount cannot exceed total"
  validate trackingNumber required when status == shipped message "Tracking number is required when shipped"
}

role Admin {
  allow all
}

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
  body sku text required
  body stock number required min 0
  update Product where sku == body.sku set stock = body.stock
  respond accepted
  webhook
  public
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

workflow OrderPreparation {
  source Order
  states draft, picking, verified, packaged, shipped

  transition startPicking {
    from draft
    to picking
    allow Admin
  }
}

state OrdersPageState {
  selectedOrders Order[]
  activeFilter text
  modal createOrder closed
}

component StockBadge {
  input stock number

  variant low when stock < 10
  variant normal when stock >= 10
}

layout AdminLayout {
  sidebar {
    item Orders
    item Products
    item Customers
  }
}

page Products {
  layout AdminLayout
  source Product
  access Admin

  view {
    order table, StockSummary, StockCards, detail, form
    compose grid columns 2 gap md stackAt md
    section table span 2
    section StockSummary component StockBadge bind selected span 1 title "Stock Summary"
    section StockCards component StockBadge bind each span 1 title "Stock Cards"
    section detail span 1
    section form span 1
  }

  table {
    columns sku, name, photo, stock, price, inventoryValue
    search sku, name
    filter stock
    sort stock desc
    paginate 10
  }

  form {
    fields sku, name, photo, stock, price
  }

  actions create, edit, delete, archive, restore
}

page Customers {
  layout AdminLayout
  source Customer
  access Admin

  table {
    columns name, email, website
    search name, email
    filter email
    sort name asc
    paginate 10
  }

  form {
    fields name, email, website
  }

  actions create, edit, delete, archive, restore
}

page Orders {
  layout AdminLayout
  source Order
  access Admin

  table {
    columns customer, total, discount, status, trackingNumber
    search customer, status, trackingNumber
    filter customer, status
    sort customer asc
    paginate 10
  }

  form {
    fields customer, total, discount, status, trackingNumber
  }

  actions create, edit, delete, archive, restore
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	files, diagnostics := BuildWeb(program, outDir)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}
	if len(files) != 52 {
		t.Fatalf("expected 52 generated files, got %d", len(files))
	}

	expected := []string{
		"README.md",
		".env.example",
		".dockerignore",
		"Dockerfile",
		"docker-compose.yml",
		"docker-compose.preview.yml",
		filepath.Join("deploy", "manifest.json"),
		filepath.Join("deploy", "rollback.json"),
		filepath.Join("deploy", "cloud.json"),
		filepath.Join("scripts", "rollback-plan.mjs"),
		filepath.Join("scripts", "cloud-plan.mjs"),
		filepath.Join("scripts", "cloud-exec.mjs"),
		filepath.Join("security", "secrets.json"),
		filepath.Join("scripts", "secrets-plan.mjs"),
		filepath.Join("scripts", "secrets-provider.mjs"),
		"openapi.json",
		"package.json",
		"index.html",
		"tsconfig.json",
		"vite.config.ts",
		"prisma.config.ts",
		filepath.Join("prisma", "schema.prisma"),
		filepath.Join("src", "main.tsx"),
		filepath.Join("src", "App.tsx"),
		filepath.Join("src", "db.ts"),
		filepath.Join("src", "setup-db.ts"),
		filepath.Join("jobs", "manifest.json"),
		filepath.Join("src", "worker.ts"),
		filepath.Join("src", "server.ts"),
		filepath.Join("src", "styles.css"),
		filepath.Join("src", "vite-env.d.ts"),
		filepath.Join("src", "types.ts"),
		filepath.Join("src", "blacklang.contract.test.ts"),
		filepath.Join("src", "blacklang.api.test.ts"),
		filepath.Join("src", "blacklang.frontend.test.tsx"),
		filepath.Join("src", "auth", "AuthPage.tsx"),
		filepath.Join("src", "auth", "UsersPage.tsx"),
		filepath.Join("src", "auth", "AuditPage.tsx"),
		filepath.Join("src", "components", "StockBadge.tsx"),
		filepath.Join("src", "routes", "auth.ts"),
		filepath.Join("src", "api", "product.ts"),
		filepath.Join("src", "api", "customer.ts"),
		filepath.Join("src", "api", "order.ts"),
		filepath.Join("src", "routes", "product.ts"),
		filepath.Join("src", "routes", "customer.ts"),
		filepath.Join("src", "routes", "order.ts"),
		filepath.Join("src", "validation", "product.ts"),
		filepath.Join("src", "validation", "customer.ts"),
		filepath.Join("src", "validation", "order.ts"),
		filepath.Join("src", "pages", "ProductsPage.tsx"),
		filepath.Join("src", "pages", "CustomersPage.tsx"),
		filepath.Join("src", "pages", "OrdersPage.tsx"),
	}
	for _, relativePath := range expected {
		if _, err := os.Stat(filepath.Join(outDir, relativePath)); err != nil {
			t.Fatalf("expected generated file %s: %v", relativePath, err)
		}
	}

	envExample, err := os.ReadFile(filepath.Join(outDir, ".env.example"))
	if err != nil {
		t.Fatalf("expected generated env example: %v", err)
	}
	envExampleText := string(envExample)
	for _, value := range []string{
		`PORT=3001`,
		`BLACKLANG_PREVIEW_PORT=4001`,
		`DATABASE_URL="file:./dev.db"`,
		`CORS_ORIGINS="http://localhost:5173"`,
		`FLY_APP_NAME=`,
		`FLY_REGION=`,
	} {
		if !strings.Contains(envExampleText, value) {
			t.Fatalf("expected env example to contain %q, got:\n%s", value, envExampleText)
		}
	}

	readme, err := os.ReadFile(filepath.Join(outDir, "README.md"))
	if err != nil {
		t.Fatalf("expected generated README: %v", err)
	}
	readmeText := string(readme)
	for _, value := range []string{
		`npm run deploy:preview`,
		`npm run deploy:preview:down`,
		`npm run deploy:rollback:plan`,
		`npm run deploy:cloud:plan`,
		`npm run deploy:cloud:preflight`,
		`npm run deploy:cloud:exec`,
		`npm run security:secrets:plan`,
		`npm run security:secrets:preflight`,
		`npm run jobs:run`,
		`npm run jobs:loop`,
		`- Target: web`,
		`- Frontend: react`,
		`- Backend: node`,
		`- Database: sqlite`,
		`- Jobs: 1`,
	} {
		if !strings.Contains(readmeText, value) {
			t.Fatalf("expected README to contain %q, got:\n%s", value, readmeText)
		}
	}

	packageJSON, err := os.ReadFile(filepath.Join(outDir, "package.json"))
	if err != nil {
		t.Fatalf("expected generated package.json: %v", err)
	}
	if !strings.Contains(string(packageJSON), `"start": "tsx src/server.ts"`) {
		t.Fatalf("expected package.json to contain start script, got:\n%s", string(packageJSON))
	}
	if !strings.Contains(string(packageJSON), `"test": "npm run db:generate && tsx src/blacklang.contract.test.ts && tsx src/blacklang.api.test.ts && tsx src/blacklang.frontend.test.tsx"`) {
		t.Fatalf("expected package.json to contain generated contract test script, got:\n%s", string(packageJSON))
	}
	for _, value := range []string{
		`"deploy:preview": "docker compose -f docker-compose.preview.yml --project-name warehouse-preview up --build"`,
		`"deploy:preview:down": "docker compose -f docker-compose.preview.yml --project-name warehouse-preview down"`,
		`"deploy:rollback:plan": "node scripts/rollback-plan.mjs"`,
		`"deploy:cloud:plan": "node scripts/cloud-plan.mjs"`,
		`"deploy:cloud:preflight": "node scripts/cloud-exec.mjs --preflight"`,
		`"deploy:cloud:exec": "node scripts/cloud-exec.mjs --apply"`,
		`"security:secrets:plan": "node scripts/secrets-plan.mjs"`,
		`"security:secrets:preflight": "node scripts/secrets-provider.mjs --preflight"`,
		`"jobs:run": "npm run db:setup && tsx src/worker.ts --once"`,
		`"jobs:loop": "npm run db:setup && tsx src/worker.ts --loop"`,
	} {
		if !strings.Contains(string(packageJSON), value) {
			t.Fatalf("expected package.json to contain %q, got:\n%s", value, string(packageJSON))
		}
	}

	secrets, err := os.ReadFile(filepath.Join(outDir, "security", "secrets.json"))
	if err != nil {
		t.Fatalf("expected generated secret manifest: %v", err)
	}
	var secretManifest generatedSecretManifest
	if err := json.Unmarshal(secrets, &secretManifest); err != nil {
		t.Fatalf("expected valid secret manifest JSON: %v\n%s", err, string(secrets))
	}
	secretReferences := map[string]generatedSecretReference{}
	for _, ref := range secretManifest.References {
		secretReferences[ref.Name] = ref
	}
	if secretManifest.PlanFile != "scripts/secrets-plan.mjs" || secretManifest.ProviderFile != "scripts/secrets-provider.mjs" || secretManifest.Policy == "" {
		t.Fatalf("expected secret manifest metadata, got %#v", secretManifest)
	}
	if ref := secretReferences["DATABASE_URL"]; !ref.Required || !ref.Sensitive || !stringSliceContains(ref.Sources, "database.url") || !stringSliceContains(ref.Sources, "deploy.env") {
		t.Fatalf("expected DATABASE_URL to be tracked as required sensitive env, got %#v", ref)
	}
	if ref := secretReferences["CORS_ORIGINS"]; ref.Required || ref.Sensitive || !stringSliceContains(ref.Sources, "security.cors.origins") {
		t.Fatalf("expected CORS_ORIGINS to be tracked as optional non-sensitive env, got %#v", ref)
	}
	if ref := secretReferences["FLY_APP_NAME"]; !ref.Required || ref.Sensitive || !stringSliceContains(ref.Sources, "deploy.cloud.app") {
		t.Fatalf("expected FLY_APP_NAME cloud env reference, got %#v", ref)
	}
	secretsPlan, err := os.ReadFile(filepath.Join(outDir, "scripts", "secrets-plan.mjs"))
	if err != nil {
		t.Fatalf("expected generated secret plan script: %v", err)
	}
	for _, value := range []string{
		`command: "security:secrets:plan"`,
		`providerKey: String(prefix) + "/" + String(ref.name)`,
		`missingRequired`,
	} {
		if !strings.Contains(string(secretsPlan), value) {
			t.Fatalf("expected secrets plan to contain %q, got:\n%s", value, string(secretsPlan))
		}
	}
	secretsProvider, err := os.ReadFile(filepath.Join(outDir, "scripts", "secrets-provider.mjs"))
	if err != nil {
		t.Fatalf("expected generated secret provider script: %v", err)
	}
	for _, value := range []string{
		`command: "security:secrets:preflight"`,
		`cli-availability-only`,
		`env: commandEnv()`,
		`never fetches, logs, writes, or injects secret values`,
	} {
		if !strings.Contains(string(secretsProvider), value) {
			t.Fatalf("expected secrets provider preflight to contain %q, got:\n%s", value, string(secretsProvider))
		}
	}

	jobManifest, err := os.ReadFile(filepath.Join(outDir, "jobs", "manifest.json"))
	if err != nil {
		t.Fatalf("expected generated jobs manifest: %v", err)
	}
	var jobs generatedJobManifest
	if err := json.Unmarshal(jobManifest, &jobs); err != nil {
		t.Fatalf("expected valid jobs manifest JSON: %v\n%s", err, string(jobManifest))
	}
	if len(jobs.Jobs) != 1 || jobs.Jobs[0].Name != "LowStockMonitor" || jobs.Jobs[0].Query != "LowStockProducts" || jobs.Jobs[0].IntervalMs != 900000 {
		t.Fatalf("expected LowStockMonitor job manifest, got %#v", jobs)
	}

	worker, err := os.ReadFile(filepath.Join(outDir, "src", "worker.ts"))
	if err != nil {
		t.Fatalf("expected generated worker: %v", err)
	}
	workerText := string(worker)
	for _, value := range []string{
		`export const blackJobs`,
		`async function lowStockMonitorJob()`,
		`await prisma.product.findMany`,
		`{ ["stock"]: { lt: Number("10") } }`,
		`take: 50`,
		`select: { id: true }`,
		`export async function runAllJobsOnce()`,
		`process.argv.includes("--loop")`,
	} {
		if !strings.Contains(workerText, value) {
			t.Fatalf("expected worker to contain %q, got:\n%s", value, workerText)
		}
	}

	frontendSmokeTest, err := os.ReadFile(filepath.Join(outDir, "src", "blacklang.frontend.test.tsx"))
	if err != nil {
		t.Fatalf("expected generated frontend smoke test: %v", err)
	}
	frontendSmokeText := string(frontendSmokeTest)
	for _, value := range []string{
		`import { renderToString } from "react-dom/server";`,
		`const html = renderToString(<App />);`,
		`assert.ok(html.includes("Checking session"), "authenticated apps should render the session check state before client effects run");`,
	} {
		if !strings.Contains(frontendSmokeText, value) {
			t.Fatalf("expected frontend smoke test to contain %q, got:\n%s", value, frontendSmokeText)
		}
	}

	apiSmokeTest, err := os.ReadFile(filepath.Join(outDir, "src", "blacklang.api.test.ts"))
	if err != nil {
		t.Fatalf("expected generated API smoke test: %v", err)
	}
	apiSmokeText := string(apiSmokeTest)
	for _, value := range []string{
		`const { createApp } = await import("./server");`,
		`process.env.CORS_ORIGINS = "https://allowed.example";`,
		`const lowStockReportResponse = await fetch(baseURL + "/api/reports/low-stock/sample?limit=7");`,
		`assert.equal(lowStockReportResponse.status, 401, "LowStockReport should reject anonymous private API requests");`,
		`const stockWebhookResponse = await fetch(baseURL + "/api/webhooks/stock", { method: "POST", headers: { "content-type": "application/json" }, body: JSON.stringify({}) });`,
		`assert.equal(stockWebhookResponse.status, 400, "StockWebhook should validate explicit API handler body before mutation");`,
		`assert.equal(api.status, 401, "authenticated API routes should reject anonymous requests");`,
		`assert.equal(deniedCors.status, 403, "CORS should reject origins outside the configured environment list");`,
		`assert.equal(allowedCors.headers.get("access-control-allow-origin"), "https://allowed.example");`,
	} {
		if !strings.Contains(apiSmokeText, value) {
			t.Fatalf("expected API smoke test to contain %q, got:\n%s", value, apiSmokeText)
		}
	}

	contractTest, err := os.ReadFile(filepath.Join(outDir, "src", "blacklang.contract.test.ts"))
	if err != nil {
		t.Fatalf("expected generated contract test: %v", err)
	}
	if !strings.Contains(string(contractTest), `photo: "data:image/png;base64,AA=="`) {
		t.Fatalf("expected contract test to include media data URL literal, got:\n%s", string(contractTest))
	}
	if !strings.Contains(string(contractTest), `"x-blacklang-jobs"`) || !strings.Contains(string(contractTest), `LowStockMonitor`) {
		t.Fatalf("expected contract test to assert job OpenAPI metadata, got:\n%s", string(contractTest))
	}

	dockerfile, err := os.ReadFile(filepath.Join(outDir, "Dockerfile"))
	if err != nil {
		t.Fatalf("expected generated Dockerfile: %v", err)
	}
	dockerfileText := string(dockerfile)
	for _, value := range []string{
		`FROM node:22-alpine`,
		`RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi`,
		`RUN npm run build`,
		`ENV PORT=3001`,
		`EXPOSE 3001`,
		`CMD ["sh", "-c", "npm run db:setup && npm run start"]`,
	} {
		if !strings.Contains(dockerfileText, value) {
			t.Fatalf("expected Dockerfile to contain %q, got:\n%s", value, dockerfileText)
		}
	}

	compose, err := os.ReadFile(filepath.Join(outDir, "docker-compose.yml"))
	if err != nil {
		t.Fatalf("expected generated docker-compose.yml: %v", err)
	}
	composeText := string(compose)
	for _, value := range []string{
		`warehouse:`,
		`- "${PORT:-3001}:${PORT:-3001}"`,
		`PORT: "${PORT:-3001}"`,
		`DATABASE_URL: "${DATABASE_URL:-file:/app/data/dev.db}"`,
		`CORS_ORIGINS: "${CORS_ORIGINS:-http://localhost:5173}"`,
		`FLY_APP_NAME: "${FLY_APP_NAME}"`,
		`FLY_REGION: "${FLY_REGION}"`,
		`warehouse-data:/app/data`,
	} {
		if !strings.Contains(composeText, value) {
			t.Fatalf("expected docker-compose.yml to contain %q, got:\n%s", value, composeText)
		}
	}

	previewCompose, err := os.ReadFile(filepath.Join(outDir, "docker-compose.preview.yml"))
	if err != nil {
		t.Fatalf("expected generated docker-compose.preview.yml: %v", err)
	}
	previewComposeText := string(previewCompose)
	for _, value := range []string{
		`warehouse-preview:`,
		`- "${BLACKLANG_PREVIEW_PORT:-4001}:${PORT:-3001}"`,
		`BLACKLANG_DEPLOY_PREVIEW: "local"`,
		`DATABASE_URL: "${DATABASE_URL:-file:/app/data/preview.db}"`,
		`CORS_ORIGINS: "${CORS_ORIGINS:-http://localhost:4001}"`,
		`FLY_APP_NAME: "${FLY_APP_NAME}"`,
		`FLY_REGION: "${FLY_REGION}"`,
		`warehouse-preview-data:/app/data`,
	} {
		if !strings.Contains(previewComposeText, value) {
			t.Fatalf("expected docker-compose.preview.yml to contain %q, got:\n%s", value, previewComposeText)
		}
	}

	manifestContent, err := os.ReadFile(filepath.Join(outDir, "deploy", "manifest.json"))
	if err != nil {
		t.Fatalf("expected deploy manifest: %v", err)
	}
	var manifest generatedDeployManifest
	if err := json.Unmarshal(manifestContent, &manifest); err != nil {
		t.Fatalf("expected valid deploy manifest JSON: %v\n%s", err, string(manifestContent))
	}
	if manifest.Preview == nil || manifest.Preview.Mode != "local" || manifest.Preview.PortDefault != "4001" {
		t.Fatalf("expected local preview metadata, got %#v", manifest.Preview)
	}
	if manifest.Rollback == nil || manifest.Rollback.Strategy != "keep" || manifest.Rollback.Keep != 3 {
		t.Fatalf("expected rollback keep metadata, got %#v", manifest.Rollback)
	}
	if manifest.Cloud == nil || manifest.Cloud.Provider != "fly" || manifest.Cloud.AppEnv != "FLY_APP_NAME" || manifest.Cloud.RegionEnv != "FLY_REGION" {
		t.Fatalf("expected cloud adapter metadata, got %#v", manifest.Cloud)
	}
	if !containsString(manifest.GeneratedFiles, "docker-compose.preview.yml") || !containsString(manifest.GeneratedFiles, "deploy/rollback.json") || !containsString(manifest.GeneratedFiles, "deploy/cloud.json") || !containsString(manifest.GeneratedFiles, "scripts/cloud-exec.mjs") || !containsString(manifest.GeneratedFiles, "security/secrets.json") || !containsString(manifest.GeneratedFiles, "scripts/secrets-plan.mjs") || !containsString(manifest.GeneratedFiles, "scripts/secrets-provider.mjs") {
		t.Fatalf("expected deploy manifest generated files to include preview, rollback, cloud, and secret artifacts, got %#v", manifest.GeneratedFiles)
	}

	cloudContent, err := os.ReadFile(filepath.Join(outDir, "deploy", "cloud.json"))
	if err != nil {
		t.Fatalf("expected deploy cloud metadata: %v", err)
	}
	var cloud generatedDeployCloud
	if err := json.Unmarshal(cloudContent, &cloud); err != nil {
		t.Fatalf("expected valid cloud JSON: %v\n%s", err, string(cloudContent))
	}
	if cloud.Provider != "fly" || cloud.Adapter != "provider-cli" || cloud.RunnerFile != "scripts/cloud-exec.mjs" || cloud.Execution.Mode != "explicit-apply" || cloud.Execution.CLI.Primary != "fly" || !containsString(cloud.Execution.CLI.Alternatives, "flyctl") || !containsString(cloud.RequiredEnv, "FLY_APP_NAME") || !containsString(cloud.RequiredEnv, "FLY_REGION") || !containsString(cloud.AuthEnv, "FLY_API_TOKEN") {
		t.Fatalf("expected cloud provider CLI adapter metadata, got %#v", cloud)
	}
	if !containsString(cloud.Execution.Apply.Args, "deploy") || !containsString(cloud.Execution.Apply.Args, "--app") || !containsString(cloud.Execution.Apply.Args, "${FLY_APP_NAME}") {
		t.Fatalf("expected fly apply command template, got %#v", cloud.Execution.Apply)
	}
	cloudScript, err := os.ReadFile(filepath.Join(outDir, "scripts", "cloud-plan.mjs"))
	if err != nil {
		t.Fatalf("expected cloud plan script: %v", err)
	}
	if !strings.Contains(string(cloudScript), `command: "deploy:cloud:plan"`) || !strings.Contains(string(cloudScript), `preflightCommand`) || !strings.Contains(string(cloudScript), `presentAuthEnv`) {
		t.Fatalf("expected cloud plan script to print readiness and execution JSON, got:\n%s", string(cloudScript))
	}
	cloudExecScript, err := os.ReadFile(filepath.Join(outDir, "scripts", "cloud-exec.mjs"))
	if err != nil {
		t.Fatalf("expected cloud execution script: %v", err)
	}
	for _, value := range []string{
		`baseReport("deploy:cloud:preflight"`,
		`baseReport("deploy:cloud:exec"`,
		`spawnSync(cliCheck.selected, resolvedArgs`,
		`provider CLI executable was not found on PATH`,
		`<blacklang-output-truncated>`,
	} {
		if !strings.Contains(string(cloudExecScript), value) {
			t.Fatalf("expected cloud execution script to contain %q, got:\n%s", value, string(cloudExecScript))
		}
	}

	rollbackContent, err := os.ReadFile(filepath.Join(outDir, "deploy", "rollback.json"))
	if err != nil {
		t.Fatalf("expected deploy rollback metadata: %v", err)
	}
	var rollback generatedDeployRollbackPlan
	if err := json.Unmarshal(rollbackContent, &rollback); err != nil {
		t.Fatalf("expected valid rollback JSON: %v\n%s", err, string(rollbackContent))
	}
	if rollback.Keep != 3 || rollback.ReleaseRoot != "releases" || len(rollback.Steps) == 0 {
		t.Fatalf("expected rollback plan metadata, got %#v", rollback)
	}
	rollbackScript, err := os.ReadFile(filepath.Join(outDir, "scripts", "rollback-plan.mjs"))
	if err != nil {
		t.Fatalf("expected rollback plan script: %v", err)
	}
	if !strings.Contains(string(rollbackScript), `command: "deploy:rollback:plan"`) || !strings.Contains(string(rollbackScript), `rollbackCandidate`) {
		t.Fatalf("expected rollback script to print rollback plan JSON, got:\n%s", string(rollbackScript))
	}

	authPage, err := os.ReadFile(filepath.Join(outDir, "src", "auth", "AuthPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated auth page: %v", err)
	}
	authPageText := string(authPage)
	for _, value := range []string{
		`export function AuthPage`,
		`const [mode, setMode] = useState<"login" | "register">("login");`,
		`fetch("/api/auth/" + mode, {`,
		`credentials: "same-origin"`,
		`const body = await response.json();`,
		`onAuthenticated(body.user);`,
		`Generated from BlackLang auth intent with cookie session persistence.`,
		`{mode === "login" ? "Sign in" : "Create account"}`,
		`<input name="email" required type="email" placeholder="you@example.com" />`,
		`<input name="password" required minLength={8} type="password" placeholder="At least 8 characters" />`,
	} {
		if !strings.Contains(authPageText, value) {
			t.Fatalf("expected AuthPage to contain %q, got:\n%s", value, authPageText)
		}
	}

	openapi, err := os.ReadFile(filepath.Join(outDir, "openapi.json"))
	if err != nil {
		t.Fatalf("expected generated openapi spec: %v", err)
	}
	openapiText := string(openapi)
	for _, value := range []string{
		`"openapi": "3.1.0"`,
		`"title": "Warehouse API"`,
		`"/api/auth/login"`,
		`"/api/auth/register"`,
		`"AuthLoginInput"`,
		`"AuthRegisterInput"`,
		`"AuthRoleUpdateInput"`,
		`"AuthUser"`,
		`"cookieAuth"`,
		`"black_session"`,
		`"/api/auth/users"`,
		`"/api/auth/audit"`,
		`"/api/auth/users/{id}/role"`,
		`"AuditLog"`,
		`"/api/orders"`,
		`"/api/orders/{id}"`,
		`"/api/orders/{id}/archive"`,
		`"/api/orders/{id}/restore"`,
		`"/api/orders/{id}/workflow/startPicking"`,
		`"Run OrderPreparation transition startPicking"`,
		`"/api/reports/low-stock/{warehouseId}"`,
		`"/api/webhooks/stock"`,
		`"x-blacklang-api": "LowStockReport"`,
		`"x-blacklang-access": "private"`,
		`"x-blacklang-runtime": "declared"`,
		`"x-blacklang-handler": "update"`,
		`"x-blacklang-update"`,
		`"x-blacklang-webhook": true`,
		`"x-blacklang-jobs"`,
		`"LowStockMonitor"`,
		`"LowStockProducts"`,
		`"intervalMs": 900000`,
		`"name": "warehouseId"`,
		`"name": "limit"`,
		`"minimum": 0`,
		`"patch"`,
		`"OrderInput"`,
		`"StockWebhookInput"`,
		`"customerId"`,
		`"format": "email"`,
		`"x-blacklang-media": "image"`,
		`"x-blacklang-encoding": "data-url"`,
		`"contentMediaType": "image/*"`,
	} {
		if !strings.Contains(openapiText, value) {
			t.Fatalf("expected openapi spec to contain %q, got:\n%s", value, openapiText)
		}
	}

	server, err := os.ReadFile(filepath.Join(outDir, "src", "server.ts"))
	if err != nil {
		t.Fatalf("expected generated server: %v", err)
	}
	serverText := string(server)
	for _, value := range []string{
		`import { authRouter, requireAuth, requireCsrf } from "./routes/auth";`,
		`import path from "node:path";`,
		`import { pathToFileURL } from "node:url";`,
		`export function createApp() {`,
		`const port = Number(process.env["PORT"] ?? "3001");`,
		`const rateWindowMs = 60_000;`,
		`const rateLimit = 120;`,
		`app.disable("x-powered-by");`,
		`res.setHeader("X-Content-Type-Options", "nosniff");`,
		`res.setHeader("X-Frame-Options", "DENY");`,
		`res.setHeader("Referrer-Policy", "no-referrer");`,
		`res.setHeader("Permissions-Policy", "geolocation=(), microphone=(), camera=()");`,
		`const corsOrigins = (process.env.CORS_ORIGINS ?? "").split(",").map((origin) => origin.trim()).filter(Boolean);`,
		`if (!corsOrigins.includes(origin)) {`,
		`res.status(403).json({ error: "CORS origin is not allowed" });`,
		`res.setHeader("Access-Control-Allow-Origin", origin);`,
		`res.setHeader("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token");`,
		`res.setHeader("Access-Control-Allow-Credentials", "true");`,
		`res.status(204).end();`,
		`res.status(429).json({ error: "Too many requests" });`,
		`app.use(express.json({ limit: "2mb" }));`,
		`type ExplicitAPIParamSpec = { name: string; type: string; required: boolean; min?: number; max?: number; minLength?: number; maxLength?: number; pattern?: string; url?: boolean; message?: string };`,
		`spec.type === "image"`,
		`spec.type === "file"`,
		`app.post("/api/webhooks/stock", async (req, res) => {`,
		`const bodyValues = readExplicitAPIValues(rawBody, [{ name: "sku", type: "text", required: true }, { name: "stock", type: "number", required: true, min: 0 }]);`,
		`const existing = await prisma.product.findFirst({ where: { sku: bodyValues.value["sku"] } as any });`,
		`data.stock = bodyValues.value["stock"];`,
		`const item = await prisma.product.update({`,
		`res.status(202).json({ api: "StockWebhook", status: "accepted", runtime: "declared", handler: "update", entity: "Product", id: item.id, updated: true });`,
		`app.get("/api/reports/low-stock/:warehouseId", (req, res) => {`,
		`app.use("/api/auth", authRouter);`,
		`app.use("/api", requireAuth);`,
		`app.use("/api", requireCsrf);`,
		`const openAPISpec = {`,
		`app.get("/openapi.json", (_req, res) => {`,
		`res.json(openAPISpec);`,
		`app.use(express.static(path.join(process.cwd(), "dist")));`,
		`res.sendFile(path.join(process.cwd(), "dist", "index.html"));`,
		`if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {`,
		`const app = createApp();`,
	} {
		if !strings.Contains(serverText, value) {
			t.Fatalf("expected server to contain %q, got:\n%s", value, serverText)
		}
	}
	publicAPIIndex := strings.Index(serverText, `app.post("/api/webhooks/stock", async (req, res) => {`)
	authMountIndex := strings.Index(serverText, `app.use("/api/auth", authRouter);`)
	privateAPIIndex := strings.Index(serverText, `app.get("/api/reports/low-stock/:warehouseId", (req, res) => {`)
	if publicAPIIndex < 0 || authMountIndex < 0 || privateAPIIndex < 0 || !(publicAPIIndex < authMountIndex && authMountIndex < privateAPIIndex) {
		t.Fatalf("expected public explicit API before auth and private explicit API after auth, got public=%d auth=%d private=%d\n%s", publicAPIIndex, authMountIndex, privateAPIIndex, serverText)
	}

	authRoute, err := os.ReadFile(filepath.Join(outDir, "src", "routes", "auth.ts"))
	if err != nil {
		t.Fatalf("expected generated auth route: %v", err)
	}
	authRouteText := string(authRoute)
	for _, value := range []string{
		`export function requireAuth(req: express.Request, res: express.Response, next: express.NextFunction) {`,
		`const csrfCookie = "black_csrf";`,
		`function readCookie(cookieHeader: string | undefined, name: string) {`,
		`function setAuthCookies(res: express.Response, token: string, csrfToken: string) {`,
		`export function requireCsrf(req: express.Request, res: express.Response, next: express.NextFunction) {`,
		`res.status(403).json({ error: "Invalid CSRF token" });`,
		`SELECT u.id, u.name, u.email, u.role, u.roles, u.passwordHash FROM \"BlackSession\" s JOIN \"BlackUser\" u ON u.id = s.userId WHERE s.id = ? AND s.expiresAt > datetime('now')`,
		`export function requirePageAccess(allowedRoles: string[]) {`,
		`role: primaryRole(roles)`,
		`roles`,
		`authRouter.get("/users", requireAuth, requirePageAccess([defaultRole]), (_req, res) => {`,
		`authRouter.get("/audit", requireAuth, requirePageAccess([defaultRole]), (_req, res) => {`,
		`authRouter.put("/users/:id/role", requireAuth, requireCsrf, requirePageAccess([defaultRole]), (req, res) => {`,
		`export function writeAuditLog(actor: ReturnType<typeof publicUser> | undefined, action: string, resource: string, resourceId: string, summary = "") {`,
		`writeAuditLog((req as any).blackUser, "role.update", "BlackUser", user.id, "Roles changed to " + roles.join(", "));`,
		`export const authRouter = express.Router();`,
		`authRouter.post("/register", (req, res) => {`,
		`authRouter.post("/login", (req, res) => {`,
		`authRouter.post("/logout", requireCsrf, (req, res) => {`,
		`res.clearCookie(csrfCookie, { path: "/" });`,
		`authRouter.get("/me", (req, res) => {`,
		`crypto.pbkdf2Sync(password, salt, 120_000, 32, "sha256")`,
		`res.cookie(sessionCookie, token, {`,
		`res.cookie(csrfCookie, csrfToken, {`,
		`httpOnly: true`,
		`httpOnly: false`,
	} {
		if !strings.Contains(authRouteText, value) {
			t.Fatalf("expected auth route to contain %q, got:\n%s", value, authRouteText)
		}
	}

	auditPage, err := os.ReadFile(filepath.Join(outDir, "src", "auth", "AuditPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated audit page: %v", err)
	}
	auditPageText := string(auditPage)
	for _, value := range []string{
		`export function AuditPage()`,
		`fetch("/api/auth/audit", { credentials: "same-origin" })`,
		`Generated from BlackLang auth, role, and action intent.`,
		`No audit logs yet.`,
	} {
		if !strings.Contains(auditPageText, value) {
			t.Fatalf("expected AuditPage to contain %q, got:\n%s", value, auditPageText)
		}
	}

	usersPage, err := os.ReadFile(filepath.Join(outDir, "src", "auth", "UsersPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated users page: %v", err)
	}
	usersPageText := string(usersPage)
	for _, value := range []string{
		`export function UsersPage()`,
		`const roles = ["Admin"];`,
		`const tenantAdminEnabled = false;`,
		`function csrfHeaders()`,
		`fetch("/api/auth/users", { credentials: "same-origin" })`,
		`fetch("/api/auth/users/" + userId + "/role", {`,
		`headers: csrfHeaders(),`,
		`Generated from BlackLang role declarations.`,
	} {
		if !strings.Contains(usersPageText, value) {
			t.Fatalf("expected UsersPage to contain %q, got:\n%s", value, usersPageText)
		}
	}

	stockBadge, err := os.ReadFile(filepath.Join(outDir, "src", "components", "StockBadge.tsx"))
	if err != nil {
		t.Fatalf("expected generated StockBadge component: %v", err)
	}
	stockBadgeText := string(stockBadge)
	for _, value := range []string{
		`type StockBadgeProps = {`,
		`stock: number;`,
		`export function StockBadge(props: StockBadgeProps) {`,
		`className={"black-component black-component-stock-badge black-component-stock-badge-" + variant}`,
		`if (Number(props.stock) < 10) return "low";`,
		`if (Number(props.stock) >= 10) return "normal";`,
	} {
		if !strings.Contains(stockBadgeText, value) {
			t.Fatalf("expected StockBadge to contain %q, got:\n%s", value, stockBadgeText)
		}
	}

	schema, err := os.ReadFile(filepath.Join(outDir, "prisma", "schema.prisma"))
	if err != nil {
		t.Fatalf("expected generated prisma schema: %v", err)
	}
	schemaText := string(schema)
	for _, value := range []string{
		"photo String?",
		"customerId String",
		`customer Customer @relation("Order_customer", fields: [customerId], references: [id])`,
		`orderCustomerItems Order[] @relation("Order_customer")`,
		`@@index([stock], map: "Product_stock_idx")`,
		`@@index([sku, stock], map: "Product_sku_stock_idx")`,
		`@@index([customerId, status], map: "Order_customer_status_idx")`,
	} {
		if !strings.Contains(schemaText, value) {
			t.Fatalf("expected schema to contain %q, got:\n%s", value, schemaText)
		}
	}

	setupDB, err := os.ReadFile(filepath.Join(outDir, "src", "setup-db.ts"))
	if err != nil {
		t.Fatalf("expected generated setup db: %v", err)
	}
	setupDBText := string(setupDB)
	for _, value := range []string{
		`CREATE TABLE IF NOT EXISTS "BlackUser"`,
		`"email" TEXT NOT NULL UNIQUE`,
		`"role" TEXT NOT NULL DEFAULT`,
		`ALTER TABLE "BlackUser" ADD COLUMN "role"`,
		`"passwordHash" TEXT NOT NULL`,
		`"photo" TEXT`,
		`CREATE TABLE IF NOT EXISTS "BlackSession"`,
		`"userId" TEXT NOT NULL REFERENCES "BlackUser"("id")`,
		`CREATE TABLE IF NOT EXISTS "BlackAuditLog"`,
		`"actorUserId" TEXT NOT NULL`,
		`CREATE INDEX IF NOT EXISTS "Product_stock_idx" ON "Product" ("stock");`,
		`CREATE INDEX IF NOT EXISTS "Product_sku_stock_idx" ON "Product" ("sku", "stock");`,
		`CREATE INDEX IF NOT EXISTS "Order_customer_status_idx" ON "Order" ("customerId", "status");`,
	} {
		if !strings.Contains(setupDBText, value) {
			t.Fatalf("expected setup db to contain %q, got:\n%s", value, setupDBText)
		}
	}

	orderPage, err := os.ReadFile(filepath.Join(outDir, "src", "pages", "OrdersPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated orders page: %v", err)
	}
	orderPageText := string(orderPage)
	for _, value := range []string{
		"type PageProps = {",
		"onNavigate?: (page: string) => void;",
		"customerApi.list()",
		"const [customerOptions, setCustomerOptions] = useState<Customer[]>([]);",
		"const [selectedOrders, setSelectedOrders] = useState<Order[]>([]);",
		"const [activeFilter, setActiveFilter] = useState<string>(\"\");",
		"const [createOrderOpen, setCreateOrderOpen] = useState(false);",
		"function openCreateOrder() {",
		"setCreateOrderOpen(true);",
		"function closeCreateOrder() {",
		"setCreateOrderOpen(false);",
		`onClick={openCreateOrder}>New Order</button>`,
		"{((canUpdate && editingId) || (canCreate && createOrderOpen)) && (",
		"{(editingId || createOrderOpen) && <button className=\"secondary\" type=\"button\" onClick={resetForm}>Cancel</button>}",
		"<select required disabled={customerOptions.length === 0} value={form.customer}",
		`<option value="">Select a customer</option>`,
		"customerId: form.customer",
		"String(item.customer?.name ?? item.customerId ?? \"\")",
		"String(item.customer?.name ?? item.customerId ?? \"\").toLowerCase().includes(query)",
		"return [...filteredItems].sort((left, right) => String(left.customer?.name ?? left.customerId ?? \"\").localeCompare(String(right.customer?.name ?? right.customerId ?? \"\")) * 1);",
		"const pageSize = 10;",
		"const paginatedItems = useMemo(() => {",
		"{paginatedItems.map((item) => (",
		"Page {safeCurrentPage} of {totalPages}",
		"onClick={() => setCurrentPage((page) => Math.min(totalPages, page + 1))}>Next</button>",
		"const [visibleColumns, setVisibleColumns] = useState<Record<string, boolean>>({ customer: true, total: true, discount: true, status: true, trackingNumber: true });",
		"const [filters, setFilters] = useState<Record<string, string>>({ customer: \"\", status: \"\" });",
		"function updateFilter(field: string, value: string) {",
		"(filters.customer.trim() === \"\" || String(item.customer?.name ?? item.customerId ?? \"\").toLowerCase().includes(filters.customer.trim().toLowerCase()))",
		"(filters.status.trim() === \"\" || String(item.status ?? \"\").toLowerCase().includes(filters.status.trim().toLowerCase()))",
		"Customer Filter",
		`placeholder="Filter Customer"`,
		"function toggleColumn(column: string) {",
		"<span className=\"muted\">Columns</span>",
		"{permissions.fields.customer !== false && <label className=\"inline-control\"><input checked={visibleColumns.customer} type=\"checkbox\" onChange={() => toggleColumn(\"customer\")} /> Customer</label>}",
		"{visibleColumns.customer && permissions.fields.customer !== false && <th>Customer</th>}",
		"{visibleColumns.customer && permissions.fields.customer !== false && <td>{String(item.customer?.name ?? item.customerId ?? \"\")}</td>}",
		"const tableColspan = visibleColumnCount + 3;",
		"<tr><td colSpan={tableColspan}>No order records yet.</td></tr>",
		"Order Total",
		"Discount",
		"Tracking Number",
		`placeholder="Enter order total"`,
		`placeholder="Enter discount"`,
		`placeholder="Enter tracking number"`,
		"Total amount for this order",
		"Customer connected to this order",
		"type FormErrors = Record<string, string>;",
		"function validateForm(form: FormState): FormErrors {",
		`errors.customer = "Customer is required";`,
		`errors.total = "Order Total must be a number";`,
		`errors.total = "Order Total must be at least 0";`,
		`errors.discount = "Discount cannot exceed total";`,
		`errors.trackingNumber = "Tracking number is required when shipped";`,
		"const visibleFormErrors = useMemo(() => Object.fromEntries(Object.entries(formErrors).filter(([field]) => touchedFields[field] || submitted)), [formErrors, touchedFields, submitted]);",
		"<form noValidate onSubmit={saveItem}>",
		"{visibleFormErrors.customer && <span className=\"field-error\">{visibleFormErrors.customer}</span>}",
		"const nextErrors = validateForm(form);",
		"if (Object.keys(nextErrors).length > 0) return;",
		"const missingRequiredRelations = customerOptions.length === 0;",
		"Create a Customer record before creating Order.",
		"onClick={() => onNavigate(\"Customers\")}>Open Customers</button>",
		"disabled={saving || missingRequiredRelations}",
		"async function runStartPickingWorkflow(item: Order) {",
		`if (!canUpdate || item.archivedAt || String(item.status ?? "") !== "draft") return;`,
		"const updated = await orderApi.transitionStartPicking(item.id);",
		"setItems((current) => current.map((existing) => existing.id === item.id ? updated : existing));",
		`onClick={() => runStartPickingWorkflow(item)}>Start Picking</button>`,
	} {
		if !strings.Contains(orderPageText, value) {
			t.Fatalf("expected OrdersPage to contain %q, got:\n%s", value, orderPageText)
		}
	}

	app, err := os.ReadFile(filepath.Join(outDir, "src", "App.tsx"))
	if err != nil {
		t.Fatalf("expected generated app page: %v", err)
	}
	appText := string(app)
	for _, value := range []string{
		`import { useEffect, useState } from "react";`,
		`import { AuthPage } from "./auth/AuthPage";`,
		`import { UsersPage } from "./auth/UsersPage";`,
		`import { AuditPage } from "./auth/AuditPage";`,
		`const pages = [
  { name: "Orders", access: ["Admin"] },
  { name: "Products", access: ["Admin"] },
  { name: "Customers", access: ["Admin"] },
  { name: "Users", access: ["Admin"] },
  { name: "Audit", access: ["Admin"] },
];`,
		"const [navOpen, setNavOpen] = useState(false);",
		"const [authenticated, setAuthenticated] = useState<boolean | null>(null);",
		"const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null);",
		`const currentRoles = currentUser ? ((currentUser.roles?.length ?? 0) > 0 ? currentUser.roles : [currentUser.role]) : [];`,
		`const visiblePages = currentUser ? pages.filter((item) => item.access.length === 0 || item.access.includes("authenticated") || currentRoles.some((role) => item.access.includes(role))) : pages;`,
		`fetch("/api/auth/me", { credentials: "same-origin" })`,
		`const csrfToken = document.cookie.split("; ").find((item) => item.startsWith("black_csrf="))?.split("=")[1] ?? "";`,
		`await fetch("/api/auth/logout", { method: "POST", credentials: "same-origin", headers: csrfToken ? { "X-CSRF-Token": decodeURIComponent(csrfToken) } : {} });`,
		`Checking session...`,
		`onClick={logout}>Logout</button>`,
		`return <AuthPage appName="Warehouse" onAuthenticated={(user) => { setCurrentUser(user); setAuthenticated(true); }} />;`,
		"function navigateTo(pageName: string) {",
		`return <OrdersPage onNavigate={navigateTo} permissions={pagePermissions("Order", ["customer", "total", "discount", "status", "trackingNumber"])} />;`,
		`return <CustomersPage onNavigate={navigateTo} permissions={pagePermissions("Customer", ["name", "email", "website"])} />;`,
		"return <UsersPage />;",
		"return <AuditPage />;",
		`function canAccessAction(action: string, resource: string) {`,
		`function canAccessField(action: string, resource: string, field: string) {`,
		`function pagePermissions(resource: string, fields: string[]) {`,
		`const writeAccess = Object.fromEntries(fields.map((field) => [field, canAccessField("update", resource, field)]));`,
		`writeFields: writeAccess`,
		"{renderPage()}",
		"<div className={navOpen ? \"app-shell nav-open\" : \"app-shell\"}>",
		"{navOpen && <button className=\"nav-backdrop\" type=\"button\" aria-label=\"Close navigation\" onClick={() => setNavOpen(false)} />}",
		"<aside className=\"app-sidebar\">",
		"<div className=\"app-brand\">Warehouse</div>",
		"onClick={() => navigateTo(item.name)}",
		"<div className=\"app-topbar\">",
		"<button className=\"menu-button secondary\" type=\"button\" aria-expanded={navOpen} onClick={() => setNavOpen(true)}>Menu</button>",
		"<span className=\"breadcrumb\">Warehouse / {page.name}</span>",
		"<div className=\"app-content\">{renderPage()}</div>",
	} {
		if !strings.Contains(appText, value) {
			t.Fatalf("expected App to contain %q, got:\n%s", value, appText)
		}
	}

	productApi, err := os.ReadFile(filepath.Join(outDir, "src", "api", "product.ts"))
	if err != nil {
		t.Fatalf("expected generated product api: %v", err)
	}
	productApiText := string(productApi)
	if !strings.Contains(productApiText, `credentials: "same-origin"`) {
		t.Fatalf("expected api client to include credentials, got:\n%s", productApiText)
	}
	if !strings.Contains(productApiText, `function csrfHeaders()`) || !strings.Contains(productApiText, `headers["X-CSRF-Token"] = decodeURIComponent(token);`) {
		t.Fatalf("expected api client to include csrf headers, got:\n%s", productApiText)
	}

	productPage, err := os.ReadFile(filepath.Join(outDir, "src", "pages", "ProductsPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated products page: %v", err)
	}
	productPageText := string(productPage)
	for _, value := range []string{
		`<main className="page-view page-view-products bl-view-compose-grid">`,
		`const stockSummaryComponentItem = selectedItem;`,
		`<section className="panel bl-view-section-detail bl-view-span-1">`,
		`<section className="panel bl-view-section-stock-summary bl-view-component-section bl-view-component-stock-badge bl-view-span-1">`,
		`<h2>Stock Summary</h2>`,
		`{permissions.fields.stock !== false ? (`,
		`stockSummaryComponentItem ? (`,
		`<StockBadge stock={Number(stockSummaryComponentItem.stock ?? 0)} />`,
		`Select a record to render StockBadge.`,
		`import { StockBadge } from "../components/StockBadge";`,
		`function renderMediaField(value: unknown, kind: "file" | "image", label: string) {`,
		`return <img className="media-thumb" src={source} alt={label} loading="lazy" />;`,
		`function readFileAsDataURL(file: File): Promise<string> {`,
		`async function updateFileField(field: string, file: File | null) {`,
		`function computeInventoryValue(item: Product) {`,
		`const value = Number(item.stock ?? 0) * Number(item.price ?? 0);`,
		`function formatComputedValue(value: number | null) {`,
		`{permissions.fields.photo !== false && <label className="inline-control"><input checked={visibleColumns.photo} type="checkbox" onChange={() => toggleColumn("photo")} /> Photo</label>}`,
		`{visibleColumns.photo && permissions.fields.photo !== false && <td>{renderMediaField(item.photo, "image", "Photo")}</td>}`,
		`{permissions.fields.photo !== false && <div><dt>Photo</dt><dd>{renderMediaField(selectedItem.photo, "image", "Photo")}</dd></div>}`,
		`{permissions.fields.stock !== false && permissions.fields.price !== false && <label className="inline-control"><input checked={visibleColumns.inventoryValue} type="checkbox" onChange={() => toggleColumn("inventoryValue")} /> Inventory Value</label>}`,
		`{visibleColumns.inventoryValue && permissions.fields.stock !== false && permissions.fields.price !== false && <th>Inventory Value</th>}`,
		`{visibleColumns.inventoryValue && permissions.fields.stock !== false && permissions.fields.price !== false && <td>{formatComputedValue(computeInventoryValue(item))}</td>}`,
		`{permissions.fields.stock !== false && permissions.fields.price !== false && <div><dt>Inventory Value</dt><dd>{formatComputedValue(computeInventoryValue(selectedItem))}</dd></div>}`,
		`{visibleColumns.stock && permissions.fields.stock !== false && <td>{<StockBadge stock={Number(item.stock ?? 0)} />}</td>}`,
		`{permissions.fields.stock !== false && <div><dt>Stock Count</dt><dd>{<StockBadge stock={Number(selectedItem.stock ?? 0)} />}</dd></div>}`,
		`const stockCardsComponentItems = items;`,
		`<section className="panel bl-view-section-stock-cards bl-view-component-section bl-view-component-stock-badge bl-view-span-1">`,
		`<div className="component-section-list" role="list" aria-label="Stock Cards">`,
		`{stockCardsComponentItems.map((item) => (`,
		`<input type="text" required minLength={3} maxLength={40} pattern="^[A-Z0-9]+$" value={form.sku}`,
		`<input type="file" accept="image/*" onChange={(event) => void updateFileField("photo", event.currentTarget.files?.[0] ?? null)} />`,
		`{form.photo && <span className="field-preview">{renderMediaField(form.photo, "image", "Photo")}</span>}`,
		`<input type="number" min="0" placeholder="Enter stock count" value={form.stock}`,
		`errors.photo = "Photo must be a valid image";`,
		`errors.sku = "Use uppercase letters and numbers";`,
		`if (!errors.sku && !(new RegExp("^[A-Z0-9]+$")).test(form.sku)) errors.sku = "Use uppercase letters and numbers";`,
		`errors.stock = "Stock Count must be at least 0";`,
		`<span className="field-preview"><StockBadge stock={Number(form.stock || 0)} /></span>`,
	} {
		if !strings.Contains(productPageText, value) {
			t.Fatalf("expected ProductsPage to contain %q, got:\n%s", value, productPageText)
		}
	}
	sectionOrder := []string{
		`bl-view-section-table`,
		`bl-view-section-stock-summary`,
		`bl-view-section-stock-cards`,
		`bl-view-section-detail`,
		`bl-view-section-form`,
	}
	previousIndex := -1
	for _, marker := range sectionOrder {
		index := strings.Index(productPageText, marker)
		if index == -1 {
			t.Fatalf("expected ProductsPage DOM to contain section marker %q, got:\n%s", marker, productPageText)
		}
		if index < previousIndex {
			t.Fatalf("expected ProductsPage DOM section order %v, got %q before an earlier section", sectionOrder, marker)
		}
		previousIndex = index
	}

	orderApi, err := os.ReadFile(filepath.Join(outDir, "src", "api", "order.ts"))
	if err != nil {
		t.Fatalf("expected generated order api: %v", err)
	}
	orderApiText := string(orderApi)
	if !strings.Contains(orderApiText, `transitionStartPicking: (id: string) =>`) || !strings.Contains(orderApiText, `endpoint + "/" + id + "/workflow/startPicking"`) {
		t.Fatalf("expected order api to include workflow transition client, got:\n%s", orderApiText)
	}

	productRoute, err := os.ReadFile(filepath.Join(outDir, "src", "routes", "product.ts"))
	if err != nil {
		t.Fatalf("expected generated product route: %v", err)
	}
	productRouteText := string(productRoute)
	for _, value := range []string{
		`import { canAccessField, filterWritableFields, requirePageAccess, requirePermission, writeAuditLog } from "./auth";`,
		`function currentUser(req: express.Request) {`,
		`writeAuditLog(currentUser(req), "create", "Product", item.id, "Product record created");`,
		`writeAuditLog(currentUser(req), "update", "Product", item.id, "Product record updated");`,
		`writeAuditLog(currentUser(req), "archive", "Product", item.id, "Product record archived");`,
		`writeAuditLog(currentUser(req), "restore", "Product", item.id, "Product record restored");`,
		`writeAuditLog(currentUser(req), "bulkDelete", "Product", ids.join(","), String(result.count) + " product records deleted");`,
		`writeAuditLog(currentUser(req), "delete", "Product", String(req.params.id), "Product record deleted");`,
		`productRouter.use("/products", requirePageAccess(["Admin"]));`,
		`function sanitizeProduct(item: any, roles: string[]) {`,
		`if (!canAccessField(roles, "read", "Product", "price")) delete output.price;`,
		`const writableValue = filterWritableFields(currentRoles(req), "update", "Product", validation.value as Record<string, unknown>);`,
		`data: writableValue as any`,
		`productRouter.get("/products", requirePermission("read", "Product"), async (req, res) => {`,
		`productRouter.post("/products", requirePermission("create", "Product"), async (req, res) => {`,
		`productRouter.put("/products/:id", requirePermission("update", "Product"), async (req, res) => {`,
		`productRouter.delete("/products", requirePermission("delete", "Product"), async (req, res) => {`,
	} {
		if !strings.Contains(productRouteText, value) {
			t.Fatalf("expected product route to contain %q, got:\n%s", value, productRouteText)
		}
	}

	productValidation, err := os.ReadFile(filepath.Join(outDir, "src", "validation", "product.ts"))
	if err != nil {
		t.Fatalf("expected generated product validation: %v", err)
	}
	productValidationText := string(productValidation)
	for _, value := range []string{
		`errors.push("Use uppercase letters and numbers");`,
		`if (!(new RegExp("^[A-Z0-9]+$")).test(value.sku)) errors.push("Use uppercase letters and numbers");`,
		`value.photo = String(input.photo);`,
		`value.photo.startsWith("data:image/")`,
		`if (parsed < 0) errors.push("stock must be at least 0");`,
		`if (parsed < 0) errors.push("price must be at least 0");`,
	} {
		if !strings.Contains(productValidationText, value) {
			t.Fatalf("expected product validation to contain %q, got:\n%s", value, productValidationText)
		}
	}

	customerPage, err := os.ReadFile(filepath.Join(outDir, "src", "pages", "CustomersPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated customers page: %v", err)
	}
	customerPageText := string(customerPage)
	for _, value := range []string{
		`<input type="url" value={form.website}`,
		`try { new URL(form.website); } catch { errors.website = "Enter a valid website"; }`,
	} {
		if !strings.Contains(customerPageText, value) {
			t.Fatalf("expected CustomersPage to contain %q, got:\n%s", value, customerPageText)
		}
	}

	customerValidation, err := os.ReadFile(filepath.Join(outDir, "src", "validation", "customer.ts"))
	if err != nil {
		t.Fatalf("expected generated customer validation: %v", err)
	}
	customerValidationText := string(customerValidation)
	if !strings.Contains(customerValidationText, `try { new URL(value.website); } catch { errors.push("Enter a valid website"); }`) {
		t.Fatalf("expected customer validation to include url validation, got:\n%s", customerValidationText)
	}

	orderRoute, err := os.ReadFile(filepath.Join(outDir, "src", "routes", "order.ts"))
	if err != nil {
		t.Fatalf("expected generated order route: %v", err)
	}
	orderValidation, err := os.ReadFile(filepath.Join(outDir, "src", "validation", "order.ts"))
	if err != nil {
		t.Fatalf("expected generated order validation: %v", err)
	}
	if !strings.Contains(string(orderValidation), `if (!(Number(value.discount) <= Number(value.total))) errors.push("Discount cannot exceed total");`) {
		t.Fatalf("expected order validation to include cross-field validation, got:\n%s", string(orderValidation))
	}
	if !strings.Contains(string(orderValidation), `if ((String(value.status) == "shipped") && (value.trackingNumber === undefined || value.trackingNumber === null || String(value.trackingNumber).trim() === "")) errors.push("Tracking number is required when shipped");`) {
		t.Fatalf("expected order validation to include conditional validation, got:\n%s", string(orderValidation))
	}
	orderRouteText := string(orderRoute)
	for _, value := range []string{
		`orderRouter.post("/orders/:id/workflow/startPicking", requirePermission("update", "Order"), async (req, res) => {`,
		`if (!user || !(user.roles ?? [user.role]).some((role) => ["Admin"].includes(role))) {`,
		`if (String(existing.status ?? "") !== "draft") {`,
		`res.status(409).json({ error: "Transition startPicking requires status draft" });`,
		`data: { status: "picking" }`,
		`writeAuditLog(currentUser(req), "workflow.startPicking", "Order", item.id, "OrderPreparation: draft -> picking");`,
	} {
		if !strings.Contains(orderRouteText, value) {
			t.Fatalf("expected order route to contain %q, got:\n%s", value, orderRouteText)
		}
	}

	styles, err := os.ReadFile(filepath.Join(outDir, "src", "styles.css"))
	if err != nil {
		t.Fatalf("expected generated styles: %v", err)
	}
	stylesText := string(styles)
	for _, value := range []string{
		".app-shell {",
		"grid-template-columns: 240px minmax(0, 1fr);",
		".app-sidebar {",
		".nav-backdrop {",
		".app-topbar {",
		".menu-button {",
		".breadcrumb {",
		".field-preview {",
		".component-section-list {",
		".media-thumb {",
		".media-link {",
		"main > header {",
		"/* Generated from BlackLang page view order and composition. */",
		".page-view-products.bl-view-compose-grid {",
		"grid-template-columns: repeat(2, minmax(0, 1fr));",
		"gap: 16px;",
		".page-view-products .bl-view-section-table {",
		"order: 1;",
		".page-view-products .bl-view-section-stock-summary {",
		"order: 2;",
		".page-view-products .bl-view-section-stock-cards {",
		"order: 3;",
		".page-view-products .bl-view-section-detail {",
		"order: 4;",
		".page-view-products .bl-view-section-form {",
		"order: 5;",
		"grid-column: span 2;",
		"grid-column: span 1;",
		"@media (max-width: 768px) {",
		"@media (max-width: 760px) {",
		".app-shell.nav-open .app-sidebar {",
	} {
		if !strings.Contains(stylesText, value) {
			t.Fatalf("expected styles to contain %q, got:\n%s", value, stylesText)
		}
	}
}

func TestSQLiteDefaultValueQuotesStringLiterals(t *testing.T) {
	if got := sqliteDefaultValue(FieldDecl{Type: "text"}, "default"); got != "'default'" {
		t.Fatalf("expected SQLite text default to be a single-quoted literal, got %q", got)
	}
	if got := sqliteDefaultValue(FieldDecl{Type: "text"}, "O'Hare"); got != "'O''Hare'" {
		t.Fatalf("expected SQLite text default to escape single quotes, got %q", got)
	}
	if got := sqliteDefaultValue(FieldDecl{Type: "number"}, "12"); got != "12" {
		t.Fatalf("expected SQLite numeric default to stay numeric, got %q", got)
	}
	if got := sqliteDefaultValue(FieldDecl{Type: "boolean"}, "true"); got != "1" {
		t.Fatalf("expected SQLite boolean true default to become 1, got %q", got)
	}
}

func TestBuildWebGeneratesInlineUICSS(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required ui text "#172026" 14 semibold left
  stock number default 0
}

page Products {
  source Product

  table {
    id ProductsTable
    class dataPanel
    columns sku, stock
    ui table border 1 solid compact true
  }

  form {
    id ProductForm
    class editPanel
    fields sku, stock
    ui box black 1 solid 8 8 5 5 6 center | button primary white 6 md solid
  }

  actions create, edit, delete
  action create id CreateProductButton
  action create class primaryAction
  action create ui button primary white 6 md solid
  action delete id DeleteProductButton
  action delete class dangerAction
  action delete ui button danger white 6 sm solid
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	styles, err := os.ReadFile(filepath.Join(outDir, "src", "styles.css"))
	if err != nil {
		t.Fatalf("expected generated styles: %v", err)
	}
	stylesText := string(styles)
	for _, value := range []string{
		"/* Generated from BlackLang inline UI intent. */",
		".bl-ui-table-products table {",
		"border: 1px solid #d8dee4;",
		".bl-ui-table-products tbody tr:nth-child(even) {",
		".bl-ui-form-products {",
		"padding: 8px 8px 5px 5px;",
		".bl-ui-form-products button {",
		".bl-ui-field-product-sku {",
		"font-size: 14px;",
		"font-weight: 600;",
		".bl-ui-action-products-create {",
		"background: #2563eb;",
		".bl-ui-action-products-delete {",
		"font-size: 13px;",
	} {
		if !strings.Contains(stylesText, value) {
			t.Fatalf("expected generated inline UI CSS to contain %q, got:\n%s", value, stylesText)
		}
	}

	page, err := os.ReadFile(filepath.Join(outDir, "src", "pages", "ProductsPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated products page: %v", err)
	}
	pageText := string(page)
	for _, value := range []string{
		`<section id="products-table" className="panel bl-view-section-table data-panel bl-ui-table-products">`,
		`<section id="product-form" className="panel bl-view-section-form edit-panel bl-ui-form-products">`,
		`<label className="bl-ui-field-product-sku">`,
		`id={!editingId ? "create-product-button-submit" : undefined} className={!editingId ? "primary-action bl-ui-action-products-create" : undefined}`,
		`id="delete-product-button-bulk" className="danger danger-action bl-ui-action-products-delete"`,
		`id={"delete-product-button-item-" + item.id} className="danger danger-action bl-ui-action-products-delete"`,
	} {
		if !strings.Contains(pageText, value) {
			t.Fatalf("expected generated page to contain %q, got:\n%s", value, pageText)
		}
	}
}

func TestBuildWebGeneratesPageViewTabs(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required
}

page Products {
  source Product

  view {
    order table, detail, form
    compose tabs gap md
    tab List sections table
    tab Record sections detail, form
  }

  table {
    columns name
  }

  form {
    fields name
  }

  actions create, edit
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	page, err := os.ReadFile(filepath.Join(outDir, "src", "pages", "ProductsPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated products page: %v", err)
	}
	pageText := string(page)
	for _, value := range []string{
		`const [activeViewTab, setActiveViewTab] = useState("List");`,
		`<main className="page-view page-view-products bl-view-compose-tabs">`,
		`<nav className="view-tabs" role="tablist" aria-label="Products view sections">`,
		`<button className={activeViewTab === "List" ? "active" : ""} type="button" role="tab" aria-selected={activeViewTab === "List"} onClick={() => setActiveViewTab("List")}>List</button>`,
		`{activeViewTab === "List" && (`,
		`{activeViewTab === "Record" && (`,
		`<div className="view-tab-panel" role="tabpanel" aria-label="Record tab">`,
	} {
		if !strings.Contains(pageText, value) {
			t.Fatalf("expected generated tabs page to contain %q, got:\n%s", value, pageText)
		}
	}

	styles, err := os.ReadFile(filepath.Join(outDir, "src", "styles.css"))
	if err != nil {
		t.Fatalf("expected generated styles: %v", err)
	}
	stylesText := string(styles)
	for _, value := range []string{
		".view-tabs {",
		".view-tabs button.active {",
		".view-tab-panel {",
		".page-view-products.bl-view-compose-tabs {",
		"gap: 16px;",
	} {
		if !strings.Contains(stylesText, value) {
			t.Fatalf("expected generated tabs styles to contain %q, got:\n%s", value, stylesText)
		}
	}
}

func TestBuildWebGeneratesResponsiveGridBreakpointLadder(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required
  name text required
  stock integer
}

page Products {
  source Product

  view {
    order table, detail, form
    compose grid columns 4 gap md stackAt sm
    section table span 4
    section detail span 2
    section form span 1
  }

  table {
    columns sku, name, stock
  }

  form {
    fields sku, name, stock
  }

  actions create, edit
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	styles, err := os.ReadFile(filepath.Join(outDir, "src", "styles.css"))
	if err != nil {
		t.Fatalf("expected generated styles: %v", err)
	}
	stylesText := string(styles)
	for _, value := range []string{
		"grid-template-columns: repeat(4, minmax(0, 1fr));",
		"@media (max-width: 1024px) {",
		"grid-template-columns: repeat(3, minmax(0, 1fr));",
		".page-view-products .bl-view-section-table {\n    grid-column: span 3;",
		"@media (max-width: 768px) {",
		"grid-template-columns: repeat(2, minmax(0, 1fr));",
		".page-view-products .bl-view-section-table {\n    grid-column: span 2;",
		"@media (max-width: 640px) {",
		"grid-template-columns: 1fr;",
		".page-view-products.bl-view-compose-grid .panel {\n    grid-column: 1 / -1;",
	} {
		if !strings.Contains(stylesText, value) {
			t.Fatalf("expected responsive grid styles to contain %q, got:\n%s", value, stylesText)
		}
	}
}

func TestBuildWebGeneratesViewSectionModalAndDrawer(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required
  name text required
  stock integer
}

page Products {
  source Product

  view {
    order table, detail, form
    compose grid columns 2 gap md stackAt md
    section table span 2
    section detail display drawer side left title "Product Drawer"
    section form display modal title "Product Form"
  }

  table {
    columns sku, name, stock
  }

  form {
    fields sku, name, stock
  }

  actions create, edit
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	page, err := os.ReadFile(filepath.Join(outDir, "src", "pages", "ProductsPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated page: %v", err)
	}
	pageText := string(page)
	for _, value := range []string{
		"const [formPanelOpen, setFormPanelOpen] = useState(false);",
		"function startCreateItem()",
		"function closeDetailPanel()",
		"onClick={startCreateItem}>New Product</button>",
		"className=\"section-overlay section-overlay-drawer section-overlay-drawer-left\"",
		"className=\"section-overlay section-overlay-modal\"",
		"className=\"panel bl-view-section-detail bl-view-display-drawer bl-view-side-left\"",
		"className=\"panel bl-view-section-form bl-view-display-modal\"",
		"<h2>{\"Product Drawer\"}</h2>",
		"<h2>{\"Product Form\"}</h2>",
	} {
		if !strings.Contains(pageText, value) {
			t.Fatalf("expected generated page to contain %q, got:\n%s", value, pageText)
		}
	}

	styles, err := os.ReadFile(filepath.Join(outDir, "src", "styles.css"))
	if err != nil {
		t.Fatalf("expected generated styles: %v", err)
	}
	stylesText := string(styles)
	for _, value := range []string{
		".section-overlay {",
		".section-overlay-modal {",
		".section-overlay-drawer-left {",
		".panel-header {",
	} {
		if !strings.Contains(stylesText, value) {
			t.Fatalf("expected generated overlay styles to contain %q, got:\n%s", value, stylesText)
		}
	}
}

func TestBuildWebGeneratesViewSectionGroups(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required
  name text required
  stock integer
}

page Products {
  source Product

  view {
    order table, detail, form
    compose grid columns 2 gap md stackAt md
    section table span 1
    group Record sections detail, form compose grid columns 2 gap sm span 1 title "Record Workspace"
  }

  table {
    columns sku, name, stock
  }

  form {
    fields sku, name, stock
  }

  actions create, edit
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	page, err := os.ReadFile(filepath.Join(outDir, "src", "pages", "ProductsPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated page: %v", err)
	}
	pageText := string(page)
	for _, value := range []string{
		`<section className="view-group bl-view-group-record bl-view-group-compose-grid bl-view-span-1" aria-label="Record Workspace">`,
		`<h2 className="view-group-title">Record Workspace</h2>`,
		`<section className="panel bl-view-section-detail">`,
		`<section className="panel bl-view-section-form">`,
	} {
		if !strings.Contains(pageText, value) {
			t.Fatalf("expected generated page to contain %q, got:\n%s", value, pageText)
		}
	}

	styles, err := os.ReadFile(filepath.Join(outDir, "src", "styles.css"))
	if err != nil {
		t.Fatalf("expected generated styles: %v", err)
	}
	stylesText := string(styles)
	for _, value := range []string{
		".view-group {",
		".view-group-title {",
		".page-view-products .bl-view-group-record {\n  order: 2;",
		".page-view-products .bl-view-group-record {\n  display: grid;",
		"grid-template-columns: repeat(2, minmax(0, 1fr));",
		"gap: 8px;",
		".page-view-products .bl-view-group-record {\n  grid-column: span 1;",
		"@media (max-width: 768px) {\n  .page-view-products .bl-view-group-record {",
	} {
		if !strings.Contains(stylesText, value) {
			t.Fatalf("expected generated group styles to contain %q, got:\n%s", value, stylesText)
		}
	}
}

func TestBuildWebGeneratesViewTriggers(t *testing.T) {
	source := `app Warehouse

entity Product {
  sku text required
  name text required
}

page Products {
  source Product

  view {
    order table, detail, form
    compose tabs gap md
    tab List sections table
    tab Record sections detail, form
    trigger detail on rowSelect
    trigger form on createStart
    trigger form on editStart
    trigger detail on saveSuccess
    trigger table on close
  }

  table {
    columns sku, name
  }

  form {
    fields sku, name
  }

  actions create, edit
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	page, err := os.ReadFile(filepath.Join(outDir, "src", "pages", "ProductsPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated page: %v", err)
	}
	pageText := string(page)
	for _, value := range []string{
		`<main className="page-view page-view-products bl-view-compose-tabs bl-view-has-triggers">`,
		`const [activeViewTab, setActiveViewTab] = useState("List");`,
		`function startCreateItem()`,
		`setActiveViewTab("Record");`,
		`setActiveViewTab("List");`,
		`onClick={startCreateItem}>New Product</button>`,
		`const item = await productApi.get(id);`,
		`setSelectedItem(item);`,
		`setSelectedItem(saved);`,
		`setForm({`,
	} {
		if !strings.Contains(pageText, value) {
			t.Fatalf("expected generated page to contain %q, got:\n%s", value, pageText)
		}
	}
}

func TestBuildWebUsesDefaultLocaleFieldLabels(t *testing.T) {
	source := `app Warehouse

i18n {
  default tr
  locales tr, en
}

label Product.name {
  tr "Ürün Adı"
  en "Product Name"
}

placeholder Product.name {
  tr "Ürün adını gir"
  en "Enter product name"
}

help Product.name {
  tr "Listelerde görünen ürün adı"
  en "Visible product name in lists"
}

message Product.name {
  tr "Geçerli ürün adı gir"
  en "Enter a valid product name"
}

label app.title {
  tr "Depo"
  en "Warehouse"
}

label app.language {
  tr "Dil"
  en "Language"
}

label app.menu {
  tr "Menü"
  en "Menu"
}

label app.primaryNavigation {
  tr "Ana gezinme"
  en "Primary navigation"
}

label app.closeNavigation {
  tr "Gezinmeyi kapat"
  en "Close navigation"
}

label app.source {
  tr "kaynak"
  en "source"
}

label page.Products {
  tr "Ürünler"
  en "Products"
}

label table.search {
  tr "Ara"
  en "Search"
}

label table.filter {
  tr "Filtre"
  en "Filter"
}

label table.columns {
  tr "Kolonlar"
  en "Columns"
}

label table.actions {
  tr "İşlemler"
  en "Actions"
}

label table.empty {
  tr "Kayıt yok."
  en "No records yet."
}

label status.loadingRecords {
  tr "Kayıtlar yükleniyor..."
  en "Loading records..."
}

label action.new.Product {
  tr "Yeni Ürün"
  en "New Product"
}

label action.create.Product {
  tr "Ürün Oluştur"
  en "Create Product"
}

label action.edit.Product {
  tr "Ürün Düzenle"
  en "Edit Product"
}

label action.view {
  tr "Görüntüle"
  en "View"
}

label action.edit {
  tr "Düzenle"
  en "Edit"
}

label action.create {
  tr "Oluştur"
  en "Create"
}

label action.saveChanges {
  tr "Değişiklikleri Kaydet"
  en "Save Changes"
}

label action.saving {
  tr "Kaydediliyor..."
  en "Saving..."
}

label action.cancel {
  tr "İptal"
  en "Cancel"
}

entity Product {
  name text required
  price money required
  stock integer required
  computed inventoryValue money = stock * price label "Inventory Value"
}

label Product.inventoryValue {
  tr "Stok Değeri"
  en "Inventory Value"
}

page Products {
  source Product

  table {
    columns name, inventoryValue
    search name
    filter name
  }

  form {
    fields name, price, stock
  }

  actions create, edit
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	app, err := os.ReadFile(filepath.Join(outDir, "src", "App.tsx"))
	if err != nil {
		t.Fatalf("expected generated app: %v", err)
	}
	appText := string(app)
	for _, value := range []string{
		`const locales = ["tr", "en"];`,
		`const defaultLocale = "tr";`,
		`const rtlLocalePrefixes = ["ar", "fa", "he", "ur"];`,
		`"app.title": { "tr": "Depo", "en": "Warehouse" },`,
		`"page.Products": { "tr": "Ürünler", "en": "Products" },`,
		`"app.language": { "tr": "Dil", "en": "Language" },`,
		`function localeDirection(locale: string)`,
		`function uiLabel(key: string, fallback: string, locale: string)`,
		`const [activeLocale, setActiveLocale] = useState(defaultLocale);`,
		`const activeDirection = localeDirection(activeLocale);`,
		`const appTitle = uiLabel("app.title", "Depo", activeLocale);`,
		`const pageLabel = page ? uiLabel("page." + page.name, page.label, activeLocale) : "";`,
		`{ name: "Products", label: "Ürünler", access: [] },`,
		`<label className="locale-switcher">{uiLabel("app.language", "Dil", activeLocale)}`,
		`<select value={activeLocale} onChange={(event) => setActiveLocale(event.target.value)}>`,
		`locale={activeLocale}`,
		`<div className="app-brand">{appTitle}</div>`,
		`aria-label={uiLabel("app.primaryNavigation", "Ana gezinme", activeLocale)}`,
		`>{uiLabel("app.menu", "Menü", activeLocale)}</button>`,
		`<span className="breadcrumb">{appTitle} / {pageLabel}</span>`,
		`<div className={navOpen ? "app-shell nav-open" : "app-shell"} lang={activeLocale} dir={activeDirection}>`,
	} {
		if !strings.Contains(appText, value) {
			t.Fatalf("expected generated app to contain %q, got:\n%s", value, appText)
		}
	}

	page, err := os.ReadFile(filepath.Join(outDir, "src", "pages", "ProductsPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated products page: %v", err)
	}
	pageText := string(page)
	for _, value := range []string{
		`locale?: string;`,
		`const defaultLocale = "tr";`,
		`"name": { "tr": "Ürün Adı", "en": "Product Name" },`,
		`"inventoryValue": { "tr": "Stok Değeri", "en": "Inventory Value" },`,
		`const fieldPlaceholders: Record<string, Record<string, string>> = {`,
		`"name": { "tr": "Ürün adını gir", "en": "Enter product name" },`,
		`const fieldHelpTexts: Record<string, Record<string, string>> = {`,
		`"name": { "tr": "Listelerde görünen ürün adı", "en": "Visible product name in lists" },`,
		`const fieldMessages: Record<string, Record<string, string>> = {`,
		`"name": { "tr": "Geçerli ürün adı gir", "en": "Enter a valid product name" },`,
		`const uiLabels: Record<string, Record<string, string>> = {`,
		`"table.search": { "tr": "Ara", "en": "Search" },`,
		`"action.new.Product": { "tr": "Yeni Ürün", "en": "New Product" },`,
		`function fieldText(texts: Record<string, Record<string, string>>, field: string, fallback: string, locale: string)`,
		`return texts[field]?.[locale] ?? texts[field]?.[defaultLocale] ?? fallback;`,
		`function uiLabel(key: string, fallback: string, locale: string)`,
		`function fieldLabel(field: string, fallback: string, locale: string)`,
		`function fieldPlaceholder(field: string, fallback: string, locale: string)`,
		`function fieldHelpText(field: string, fallback: string, locale: string)`,
		`function fieldMessage(field: string, fallback: string, locale: string)`,
		`function formatFieldValue(value: unknown, kind: string, locale: string)`,
		`new Intl.NumberFormat(locale, options).format(number)`,
		`new Intl.DateTimeFormat(locale, options).format(date)`,
		`currency: localeCurrency(locale)`,
		`export function ProductsPage({ onNavigate, permissions = defaultPermissions, locale = "tr" }: PageProps)`,
		`const formErrors = useMemo(() => validateForm(form, locale), [form, locale]);`,
		`const nextErrors = validateForm(form, locale);`,
		`errors.name = fieldMessage("name", "Ürün Adı is required", locale);`,
		`<h1>{uiLabel("page.Products", "Ürünler", locale)}</h1>`,
		`<span>{uiLabel("app.title", "Depo", locale)} {uiLabel("app.source", "kaynak", locale)}: Product</span>`,
		`placeholder={uiLabel("table.search", "Ara", locale)}`,
		`{fieldLabel("name", "Name", locale)} {uiLabel("table.filter", "Filtre", locale)}`,
		`placeholder={uiLabel("table.filter", "Filtre", locale) + " " + fieldLabel("name", "Name", locale)}`,
		`<span className="muted">{uiLabel("table.columns", "Kolonlar", locale)}</span>`,
		`{loading && <div className="status">{uiLabel("status.loadingRecords", "Kayıtlar yükleniyor...", locale)}</div>}`,
		`<tr><td colSpan={tableColspan}>{uiLabel("table.empty", "Kayıt yok.", locale)}</td></tr>`,
		`<th>{fieldLabel("name", "Name", locale)}</th>`,
		`<th>{fieldLabel("inventoryValue", "Inventory Value", locale)}</th>`,
		`<th>{uiLabel("table.actions", "İşlemler", locale)}</th>`,
		`onClick={() => viewItem(item.id)}>{uiLabel("action.view", "Görüntüle", locale)}</button>`,
		`onClick={() => editItem(item)}>{uiLabel("action.edit", "Düzenle", locale)}</button>`,
		`<dt>{fieldLabel("name", "Name", locale)}</dt>`,
		`<dd>{formatFieldValue(selectedItem.price, "money", locale)}</dd>`,
		`<dd>{formatFieldValue(selectedItem.stock, "integer", locale)}</dd>`,
		`<dt>{fieldLabel("inventoryValue", "Inventory Value", locale)}</dt>`,
		`<td>{formatComputedValue(computeInventoryValue(item), "money", locale)}</td>`,
		`<dd>{formatComputedValue(computeInventoryValue(selectedItem), "money", locale)}</dd>`,
		`
              {fieldLabel("name", "Name", locale)}
              <input type="text" required placeholder={fieldPlaceholder("name", "", locale)} value={form.name}`,
		`{fieldHelpText("name", "", locale) && <span className="field-note">{fieldHelpText("name", "", locale)}</span>}`,
		`{saving ? uiLabel("action.saving", "Kaydediliyor...", locale) : editingId ? uiLabel("action.saveChanges", "Değişiklikleri Kaydet", locale) : uiLabel("action.create", "Oluştur", locale)}`,
		`onClick={resetForm}>{uiLabel("action.cancel", "İptal", locale)}</button>`,
	} {
		if !strings.Contains(pageText, value) {
			t.Fatalf("expected generated page to contain %q, got:\n%s", value, pageText)
		}
	}
}

func TestBuildWebUsesThemeSlotOrderForInlineUICSS(t *testing.T) {
	source := `app Warehouse

entity Product {
  name text required ui text 16 "#172026" left semibold
}

page Products {
  source Product

  table {
    columns name
    ui table 1 border solid compact true
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
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	theme, themeDiagnostics := ParseTheme("theme.blackthm", `blackthm WarehouseTheme {
  version 1
  target web
  locked false

  profile UICompact {
    version 1
    ui box = color width style pt pr pb pl radius place;
    ui text = size color align weight;
    ui table = width color style density zebra;
    ui button = bg color radius size variant;
  }
}
`)
	if len(themeDiagnostics) != 0 {
		t.Fatalf("expected no theme diagnostics, got %#v", themeDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWebWithTheme(program, outDir, &theme); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	styles, err := os.ReadFile(filepath.Join(outDir, "src", "styles.css"))
	if err != nil {
		t.Fatalf("expected generated styles: %v", err)
	}
	stylesText := string(styles)
	for _, value := range []string{
		".bl-ui-field-product-name {",
		"font-size: 16px;",
		"color: #172026;",
		"text-align: left;",
		"font-weight: 600;",
		"border: 1px solid #d8dee4;",
	} {
		if !strings.Contains(stylesText, value) {
			t.Fatalf("expected generated inline UI CSS to contain %q, got:\n%s", value, stylesText)
		}
	}
}

func TestBuildWebUsesLowerCamelIdentifiersForCompoundEntityNames(t *testing.T) {
	source := `app InventoryControl

auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}

entity PurchaseOrder {
  orderNumber text required
  status text default draft
}

role Admin {
  allow all
}

workflow PurchaseOrderFlow {
  source PurchaseOrder
  states draft, submitted

  transition submit {
    from draft
    to submitted
    allow Admin
  }
}

page PurchaseOrders {
  source PurchaseOrder

  table {
    columns orderNumber, status
  }

  form {
    fields orderNumber, status
  }

  actions create, edit, delete
}
`

	program, parseDiagnostics := Parse("compound.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	server, err := os.ReadFile(filepath.Join(outDir, "src", "server.ts"))
	if err != nil {
		t.Fatalf("expected generated server: %v", err)
	}
	route, err := os.ReadFile(filepath.Join(outDir, "src", "routes", "purchaseorder.ts"))
	if err != nil {
		t.Fatalf("expected generated purchase order route: %v", err)
	}
	page, err := os.ReadFile(filepath.Join(outDir, "src", "pages", "PurchaseOrdersPage.tsx"))
	if err != nil {
		t.Fatalf("expected generated purchase orders page: %v", err)
	}

	serverText := string(server)
	if !strings.Contains(serverText, `import { purchaseOrderRouter as page0Router } from "./routes/purchaseorder";`) {
		t.Fatalf("expected server to import lower-camel router, got:\n%s", serverText)
	}
	if !strings.Contains(serverText, `app.use("/api", page0Router);`) {
		t.Fatalf("expected server to register lower-camel router, got:\n%s", serverText)
	}
	routeText := string(route)
	for _, value := range []string{
		"const purchaseOrderModel = prisma.purchaseOrder;",
		"purchaseOrderRouter.post",
		"purchaseOrderModel.findUnique",
		"purchaseOrderModel.update",
	} {
		if !strings.Contains(routeText, value) {
			t.Fatalf("expected purchase order route to contain %q, got:\n%s", value, routeText)
		}
	}
	pageText := string(page)
	if !strings.Contains(pageText, `import { purchaseOrderApi } from "../api/purchaseorder";`) {
		t.Fatalf("expected page to import lower-camel API client, got:\n%s", pageText)
	}
	if !strings.Contains(pageText, "purchaseOrderApi.list(showArchived)") {
		t.Fatalf("expected page to call lower-camel API client, got:\n%s", pageText)
	}
}

func TestBuildWebGeneratesEntityPolicyRuntime(t *testing.T) {
	source := `app SecureWarehouse

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

entity WorkItem {
  ownerId text required
  tenantId text required
  title text required
  policy owner ownerId
  policy tenant tenantId
}

query OpenWorkItems {
  source WorkItem
  where title != "done"
  sort title asc
  limit 25
}

action RenameWorkItem {
  source WorkItem
  input newTitle text required
  set title = newTitle
}

role Admin {
  allow all
}

page WorkItems {
  source WorkItem
  query OpenWorkItems
  access Admin

  table {
    columns title
  }

  form {
    fields title
  }

  actions create, edit, delete, archive, restore, RenameWorkItem
}
`
	program, parseDiagnostics := Parse("test.black", source)
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

	route := readGeneratedFile(t, outDir, filepath.Join("src", "routes", "workitem.ts"))
	expectedRouteSnippets := []string{
		`function rowPolicyWhere(req: express.Request`,
		`{ ownerId: user.id },`,
		`{ tenantId: user.tenantId ?? "default" },`,
		`validateWorkItemInput(applyRowPolicyInput(req, req.body as Record<string, unknown>))`,
		`Object.assign(writableValue, applyRowPolicyInput(req, {}));`,
		`const writableUpdateValue = stripRowPolicyFields(writableValue);`,
		`const item = await workItemModel.findFirst({`,
		`where: rowPolicyWhere(req, { id: String(req.params.id) }) as any`,
		`where: rowPolicyWhere(req, { id: { in: ids } }) as any`,
		`post("/workitems/:id/actions/renameworkitem"`,
	}
	for _, snippet := range expectedRouteSnippets {
		if !strings.Contains(route, snippet) {
			t.Fatalf("expected generated route to contain %q, got:\n%s", snippet, route)
		}
	}

	client := readGeneratedFile(t, outDir, filepath.Join("src", "api", "workitem.ts"))
	if !strings.Contains(client, `export type WorkItemInput = Omit<WorkItem, "id" | "ownerId" | "tenantId">;`) {
		t.Fatalf("expected policy fields to be omitted from client input, got:\n%s", client)
	}

	authRoute := readGeneratedFile(t, outDir, filepath.Join("src", "routes", "auth.ts"))
	for _, snippet := range []string{
		`tenantId: string;`,
		`tenantId: user.tenantId`,
		`tenantId: "default"`,
		`authRouter.put("/users/:id/tenant", requireAuth, requireCsrf, requirePageAccess([defaultRole]), (req, res) => {`,
		`UPDATE \"BlackUser\" SET tenantId = ?, updatedAt = CURRENT_TIMESTAMP WHERE id = ?`,
		`writeAuditLog((req as any).blackUser, "tenant.update", "BlackUser", user.id, "Tenant changed to " + tenantId);`,
	} {
		if !strings.Contains(authRoute, snippet) {
			t.Fatalf("expected generated auth route to contain %q, got:\n%s", snippet, authRoute)
		}
	}

	usersPage := readGeneratedFile(t, outDir, filepath.Join("src", "auth", "UsersPage.tsx"))
	for _, snippet := range []string{
		`const tenantAdminEnabled = true;`,
		`fetch("/api/auth/users/" + userId + "/tenant", {`,
		`Generated from BlackLang role declarations and tenant policies.`,
		`<span className="field-note">Tenant ID</span>`,
	} {
		if !strings.Contains(usersPage, snippet) {
			t.Fatalf("expected generated users page to contain %q, got:\n%s", snippet, usersPage)
		}
	}

	openapi := readGeneratedFile(t, outDir, "openapi.json")
	var spec map[string]any
	if err := json.Unmarshal([]byte(openapi), &spec); err != nil {
		t.Fatalf("expected valid openapi json: %v", err)
	}
	schemas := spec["components"].(map[string]any)["schemas"].(map[string]any)
	workItemInput := schemas["WorkItemInput"].(map[string]any)
	workItemInputProperties := workItemInput["properties"].(map[string]any)
	if _, ok := workItemInputProperties["ownerId"]; ok {
		t.Fatalf("did not expect ownerId in WorkItemInput schema, got %#v", workItemInputProperties)
	}
	if _, ok := workItemInputProperties["tenantId"]; ok {
		t.Fatalf("did not expect tenantId in WorkItemInput schema, got %#v", workItemInputProperties)
	}
	authUser := schemas["AuthUser"].(map[string]any)
	authUserProperties := authUser["properties"].(map[string]any)
	if _, ok := authUserProperties["tenantId"]; !ok {
		t.Fatalf("expected tenantId in AuthUser schema, got %#v", authUserProperties)
	}
	if _, ok := schemas["AuthTenantUpdateInput"]; !ok {
		t.Fatalf("expected AuthTenantUpdateInput schema, got %#v", schemas)
	}
	paths := spec["paths"].(map[string]any)
	if _, ok := paths["/api/auth/users/{id}/tenant"]; !ok {
		t.Fatalf("expected tenant update path in OpenAPI, got %#v", paths)
	}
}

func TestBuildWebGeneratesOpsRuntime(t *testing.T) {
	source := `app OpsDemo

target web {
  frontend react
  backend node
  database sqlite
}

deploy {
  target docker
  port env PORT default 3001
}

ops {
  health path "/healthz"
  readiness path "/readyz"
  metrics path "/metrics"
  logging requests
  observe webhook endpoint env BLACKLANG_OBSERVABILITY_ENDPOINT
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
`
	program, parseDiagnostics := Parse("test.black", source)
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

	server := readGeneratedFile(t, outDir, filepath.Join("src", "server.ts"))
	for _, snippet := range []string{
		`import { prisma } from "./db";`,
		`const opsCounters = { requests: 0, errors: 0 };`,
		`console.log(JSON.stringify({`,
		`function opsObserveEndpoint() {`,
		`process.env["BLACKLANG_OBSERVABILITY_ENDPOINT"] ?? ""`,
		`type OpsTraceContext = {`,
		`function opsTraceContext(header: string | string[] | undefined): OpsTraceContext {`,
		`res.setHeader("traceparent", traceContext.traceparent);`,
		`void fetch(observeEndpoint, {`,
		`kind: "observability_delivery_error"`,
		`path: req.originalUrl,`,
		`traceId: traceContext.traceId,`,
		`app.get("/healthz", (_req, res) => {`,
		`app.get("/readyz", async (_req, res) => {`,
		`await prisma.$queryRawUnsafe("SELECT 1");`,
		`app.get("/metrics", (_req, res) => {`,
	} {
		if !strings.Contains(server, snippet) {
			t.Fatalf("expected server.ts to contain %q, got:\n%s", snippet, server)
		}
	}

	openapi := readGeneratedFile(t, outDir, "openapi.json")
	var spec map[string]any
	if err := json.Unmarshal([]byte(openapi), &spec); err != nil {
		t.Fatalf("expected valid openapi json: %v", err)
	}
	paths := spec["paths"].(map[string]any)
	for route, kind := range map[string]string{"/healthz": "health", "/readyz": "readiness", "/metrics": "metrics"} {
		operations := paths[route].(map[string]any)
		get := operations["get"].(map[string]any)
		if get["x-blacklang-ops"] != kind || get["x-blacklang-public"] != true {
			t.Fatalf("expected %s to include ops metadata %s, got %#v", route, kind, get)
		}
	}
	observability := spec["x-blacklang-observability"].(map[string]any)
	if observability["provider"] != "webhook" || observability["endpointEnv"] != "BLACKLANG_OBSERVABILITY_ENDPOINT" || observability["traceContext"] != "w3c-traceparent" || observability["protocol"] != "http-json" {
		t.Fatalf("expected OpenAPI observability metadata, got %#v", observability)
	}
	manifestJSON := readGeneratedFile(t, outDir, filepath.Join("ops", "observability.json"))
	var manifest generatedOpsObservabilityManifest
	if err := json.Unmarshal([]byte(manifestJSON), &manifest); err != nil {
		t.Fatalf("expected valid observability manifest JSON: %v\n%s", err, manifestJSON)
	}
	if manifest.Provider != "webhook" || manifest.EndpointEnv != "BLACKLANG_OBSERVABILITY_ENDPOINT" || manifest.Signal != "request-events" || manifest.TraceContext != "w3c-traceparent" || !manifest.Exporter.NonBlocking {
		t.Fatalf("expected webhook observability manifest metadata, got %#v", manifest)
	}

	dockerfile := readGeneratedFile(t, outDir, "Dockerfile")
	if !strings.Contains(dockerfile, `HEALTHCHECK --interval=30s`) || !strings.Contains(dockerfile, `/healthz`) {
		t.Fatalf("expected Dockerfile healthcheck to use /healthz, got:\n%s", dockerfile)
	}
	if strings.Contains(dockerfile, `\u003e`) {
		t.Fatalf("expected Dockerfile healthcheck to contain executable JavaScript, got escaped greater-than:\n%s", dockerfile)
	}
	compose := readGeneratedFile(t, outDir, "docker-compose.yml")
	if !strings.Contains(compose, "healthcheck:") || !strings.Contains(compose, `/healthz`) {
		t.Fatalf("expected docker-compose healthcheck to use /healthz, got:\n%s", compose)
	}
	if strings.Contains(compose, `\u003e`) {
		t.Fatalf("expected docker-compose healthcheck to contain executable JavaScript, got escaped greater-than:\n%s", compose)
	}

	contractTest := readGeneratedFile(t, outDir, filepath.Join("src", "blacklang.contract.test.ts"))
	for _, snippet := range []string{`assertPath("/healthz", "get");`, `x-blacklang-ops`, `x-blacklang-observability`} {
		if !strings.Contains(contractTest, snippet) {
			t.Fatalf("expected contract test to contain %q, got:\n%s", snippet, contractTest)
		}
	}
	apiTest := readGeneratedFile(t, outDir, filepath.Join("src", "blacklang.api.test.ts"))
	for _, snippet := range []string{`baseURL + "/healthz"`, `baseURL + "/readyz"`, `baseURL + "/metrics"`, `observeServer`, `process.env["BLACKLANG_OBSERVABILITY_ENDPOINT"]`, `openapi.headers.get("traceparent")`, `ops observe webhook should receive structured request events with trace context`} {
		if !strings.Contains(apiTest, snippet) {
			t.Fatalf("expected API smoke test to contain %q, got:\n%s", snippet, apiTest)
		}
	}
}

func TestBuildWebGeneratesOTLPTraceExporter(t *testing.T) {
	source := `app TraceDemo

target web {
  frontend react
  backend node
  database sqlite
}

ops {
  health path "/healthz"
  metrics path "/metrics"
  observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT
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
`
	program, parseDiagnostics := Parse("test.black", source)
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

	server := readGeneratedFile(t, outDir, filepath.Join("src", "server.ts"))
	for _, snippet := range []string{
		`import crypto from "node:crypto";`,
		`const opsObserveProvider = "otlp";`,
		`function opsTraceContext(header: string | string[] | undefined): OpsTraceContext {`,
		`res.setHeader("traceparent", traceContext.traceparent);`,
		`function opsOTLPTracePayload(event: Record<string, unknown>, traceContext: OpsTraceContext, started: number) {`,
		`resourceSpans: [{`,
		`scopeSpans: [{ scope: { name: "blacklang.generated.ops"`,
		`span.parentSpanId = traceContext.parentSpanId;`,
	} {
		if !strings.Contains(server, snippet) {
			t.Fatalf("expected OTLP server.ts to contain %q, got:\n%s", snippet, server)
		}
	}

	openapi := readGeneratedFile(t, outDir, "openapi.json")
	var spec map[string]any
	if err := json.Unmarshal([]byte(openapi), &spec); err != nil {
		t.Fatalf("expected valid openapi json: %v", err)
	}
	observability := spec["x-blacklang-observability"].(map[string]any)
	if observability["provider"] != "otlp" || observability["endpointEnv"] != "BLACKLANG_OTLP_ENDPOINT" || observability["protocol"] != "otlp-http-json" || observability["mode"] != "async-otlp-http-json-traces" || observability["signal"] != "traces" || observability["traceContext"] != "w3c-traceparent" {
		t.Fatalf("expected OTLP OpenAPI observability metadata, got %#v", observability)
	}

	manifestJSON := readGeneratedFile(t, outDir, filepath.Join("ops", "observability.json"))
	var manifest generatedOpsObservabilityManifest
	if err := json.Unmarshal([]byte(manifestJSON), &manifest); err != nil {
		t.Fatalf("expected valid observability manifest JSON: %v\n%s", err, manifestJSON)
	}
	if manifest.Provider != "otlp" || manifest.EndpointEnv != "BLACKLANG_OTLP_ENDPOINT" || manifest.Protocol != "otlp-http-json" || manifest.Signal != "traces" || manifest.TraceContext != "w3c-traceparent" || !manifest.Exporter.NonBlocking {
		t.Fatalf("expected OTLP observability manifest metadata, got %#v", manifest)
	}

	apiTest := readGeneratedFile(t, outDir, filepath.Join("src", "blacklang.api.test.ts"))
	for _, snippet := range []string{
		`process.env["BLACKLANG_OTLP_ENDPOINT"]`,
		`openapi.headers.get("traceparent")`,
		`ops observe otlp should receive OTLP trace payloads`,
		`Array.isArray(event.resourceSpans)`,
		`typeof span.traceId === "string"`,
	} {
		if !strings.Contains(apiTest, snippet) {
			t.Fatalf("expected OTLP API smoke test to contain %q, got:\n%s", snippet, apiTest)
		}
	}
}

func TestBuildWebGeneratesCloudPlanWithoutRollback(t *testing.T) {
	source := `app CloudDemo

target web {
  frontend react
  backend node
  database sqlite
}

deploy {
  target docker
  cloud render app env RENDER_SERVICE_ID
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
`
	program, parseDiagnostics := Parse("test.black", source)
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

	manifest := readGeneratedFile(t, outDir, filepath.Join("deploy", "manifest.json"))
	if !strings.Contains(manifest, `"provider": "render"`) || !strings.Contains(manifest, `"deploy/cloud.json"`) {
		t.Fatalf("expected deploy manifest to include render cloud metadata, got:\n%s", manifest)
	}
	cloud := readGeneratedFile(t, outDir, filepath.Join("deploy", "cloud.json"))
	if !strings.Contains(cloud, `"appEnv": "RENDER_SERVICE_ID"`) || !strings.Contains(cloud, `"scripts/cloud-plan.mjs"`) || !strings.Contains(cloud, `"scripts/cloud-exec.mjs"`) || !strings.Contains(cloud, `"primary": "render"`) || !strings.Contains(cloud, `"BLACKLANG_CLOUD_IMAGE"`) {
		t.Fatalf("expected cloud manifest to include app env, plan script, execution script, and render CLI metadata, got:\n%s", cloud)
	}
	if _, err := os.Stat(filepath.Join(outDir, "deploy", "rollback.json")); !os.IsNotExist(err) {
		t.Fatalf("expected rollback metadata to be omitted when rollback is not declared")
	}
}

func TestBuildWebSupportsPostgresRuntime(t *testing.T) {
	source := `app SupportDesk

target web {
  frontend react
  backend node
  database postgres
}

database {
  url env BLACK_DATABASE_URL
}

auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}

deploy {
  target docker
  port env PORT default 3001
  env BLACK_DATABASE_URL required
}

entity Ticket {
  title text required
  priority text default medium
  index priority
}

role Admin {
  allow all
}

page Tickets {
  source Ticket
  access Admin

  table {
    columns title, priority
  }

  form {
    fields title, priority
  }

  actions create, edit, delete
}
`

	program, parseDiagnostics := Parse("postgres.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	assertContains := func(relativePath string, values ...string) string {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(outDir, relativePath))
		if err != nil {
			t.Fatalf("expected generated %s: %v", relativePath, err)
		}
		text := string(content)
		for _, value := range values {
			if !strings.Contains(text, value) {
				t.Fatalf("expected %s to contain %q, got:\n%s", relativePath, value, text)
			}
		}
		return text
	}
	assertNotContains := func(text string, value string) {
		t.Helper()
		if strings.Contains(text, value) {
			t.Fatalf("expected text not to contain %q, got:\n%s", value, text)
		}
	}

	envExample := assertContains(".env.example",
		`BLACK_DATABASE_URL="postgresql://blacklang:blacklang@localhost:5432/blacklang"`,
		`POSTGRES_DB="blacklang"`,
		`POSTGRES_USER="blacklang"`,
		`POSTGRES_PASSWORD="blacklang"`,
	)
	assertNotContains(envExample, `DATABASE_URL="file:./dev.db"`)

	packageJSON := assertContains("package.json",
		`"@prisma/adapter-pg": "7.10.0"`,
		`"pg": "latest"`,
		`"@types/pg": "latest"`,
	)
	assertNotContains(packageJSON, "better-sqlite3")

	prismaConfig := assertContains("prisma.config.ts", `url: process.env.BLACK_DATABASE_URL ?? "postgresql://blacklang:blacklang@localhost:5432/blacklang"`)
	assertNotContains(prismaConfig, "file:./dev.db")

	schema := assertContains(filepath.Join("prisma", "schema.prisma"),
		`provider = "postgresql"`,
		`model BlackUser {`,
		`email String @unique`,
		`role String @default("Admin")`,
		`model BlackSession {`,
		`@@index([userId], map: "BlackSession_userId_idx")`,
		`model BlackAuditLog {`,
		`@@index([actorUserId], map: "BlackAuditLog_actorUserId_idx")`,
		`model Ticket {`,
		`@@index([priority], map: "Ticket_priority_idx")`,
	)
	assertNotContains(schema, `provider = "sqlite"`)

	dbClient := assertContains(filepath.Join("src", "db.ts"),
		`import { PrismaPg } from "@prisma/adapter-pg";`,
		`connectionString: process.env.BLACK_DATABASE_URL ?? "postgresql://blacklang:blacklang@localhost:5432/blacklang"`,
	)
	assertNotContains(dbClient, "PrismaBetterSqlite3")

	setupDB := assertContains(filepath.Join("src", "setup-db.ts"),
		`process.env.BLACK_DATABASE_URL = process.env.BLACK_DATABASE_URL ?? "postgresql://blacklang:blacklang@localhost:5432/blacklang";`,
		`const command = process.platform === "win32" ? "npm.cmd" : "npm";`,
		`spawnSync(command, ["run", "db:push:native"], {`,
		`PostgreSQL database schema ready`,
	)
	assertNotContains(setupDB, "better-sqlite3")

	authRoute := assertContains(filepath.Join("src", "routes", "auth.ts"),
		`import { prisma } from "../db";`,
		`export async function requireAuth(req: express.Request, res: express.Response, next: express.NextFunction) {`,
		`await prisma.blackUser.findUnique({ where: { email } })`,
		`await prisma.blackSession.create({ data: { id: token, userId: user.id, expiresAt: sessionExpiry() } });`,
		`void prisma.blackAuditLog.create({`,
	)
	assertNotContains(authRoute, "better-sqlite3")

	compose := assertContains("docker-compose.yml",
		`depends_on:`,
		`postgres:`,
		`condition: service_healthy`,
		`BLACK_DATABASE_URL: "${BLACK_DATABASE_URL:-postgresql://blacklang:blacklang@postgres:5432/blacklang}"`,
		`image: postgres:17-alpine`,
		`POSTGRES_PASSWORD: "${POSTGRES_PASSWORD:-blacklang}"`,
		`pg_isready -U ${POSTGRES_USER:-blacklang} -d ${POSTGRES_DB:-blacklang}`,
	)
	assertNotContains(compose, "/app/data")
}

func TestBuildWebSupportsMySQLRuntime(t *testing.T) {
	source := `app SupportDesk

target web {
  frontend react
  backend node
  database mysql
}

database {
  url env BLACK_DATABASE_URL
}

auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}

deploy {
  target docker
  port env PORT default 3001
  env BLACK_DATABASE_URL required
  preview local
}

entity Ticket {
  title text required
  priority text default medium
  resolved boolean default false
  index priority
}

seed DemoTickets {
  source Ticket

  row DemoTicketOne {
    title "Broken printer"
    resolved false
  }
}

role Admin {
  allow all
}

page Tickets {
  source Ticket
  access Admin

  table {
    columns title, priority, resolved
  }

  form {
    fields title, priority, resolved
  }

  actions create, edit, delete
}
`

	program, parseDiagnostics := Parse("mysql.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	assertContains := func(relativePath string, values ...string) string {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(outDir, relativePath))
		if err != nil {
			t.Fatalf("expected generated %s: %v", relativePath, err)
		}
		text := string(content)
		for _, value := range values {
			if !strings.Contains(text, value) {
				t.Fatalf("expected %s to contain %q, got:\n%s", relativePath, value, text)
			}
		}
		return text
	}
	assertNotContains := func(text string, value string) {
		t.Helper()
		if strings.Contains(text, value) {
			t.Fatalf("expected text not to contain %q, got:\n%s", value, text)
		}
	}

	envExample := assertContains(".env.example",
		`BLACK_DATABASE_URL="mysql://blacklang:blacklang@localhost:3306/blacklang"`,
		`MYSQL_DATABASE="blacklang"`,
		`MYSQL_USER="blacklang"`,
		`MYSQL_PASSWORD="blacklang"`,
		`MYSQL_ROOT_PASSWORD="blacklang_root"`,
	)
	assertNotContains(envExample, `DATABASE_URL="file:./dev.db"`)

	packageJSON := assertContains("package.json",
		`"@prisma/adapter-mariadb": "7.10.0"`,
		`"@prisma/client": "7.10.0"`,
		`"db:setup": "tsx src/setup-db.ts && npm run db:seed"`,
	)
	assertNotContains(packageJSON, "better-sqlite3")
	assertNotContains(packageJSON, `"@prisma/adapter-pg"`)
	assertNotContains(packageJSON, `"pg":`)

	prismaConfig := assertContains("prisma.config.ts", `url: process.env.BLACK_DATABASE_URL ?? "mysql://blacklang:blacklang@localhost:3306/blacklang"`)
	assertNotContains(prismaConfig, "file:./dev.db")

	schema := assertContains(filepath.Join("prisma", "schema.prisma"),
		`provider = "mysql"`,
		`model BlackUser {`,
		`model BlackSession {`,
		`model BlackAuditLog {`,
		`model Ticket {`,
		`@@index([priority], map: "Ticket_priority_idx")`,
	)
	assertNotContains(schema, `provider = "postgresql"`)
	assertNotContains(schema, `provider = "sqlite"`)

	dbClient := assertContains(filepath.Join("src", "db.ts"),
		`import { PrismaMariaDb } from "@prisma/adapter-mariadb";`,
		`const adapter = new PrismaMariaDb(process.env.BLACK_DATABASE_URL ?? "mysql://blacklang:blacklang@localhost:3306/blacklang");`,
		`export const prisma = new PrismaClient({ adapter });`,
	)
	assertNotContains(dbClient, "PrismaPg")
	assertNotContains(dbClient, "PrismaBetterSqlite3")

	setupDB := assertContains(filepath.Join("src", "setup-db.ts"),
		`process.env.BLACK_DATABASE_URL = process.env.BLACK_DATABASE_URL ?? "mysql://blacklang:blacklang@localhost:3306/blacklang";`,
		`spawnSync(command, ["run", "db:push:native"], {`,
		`MySQL database schema ready`,
	)
	assertNotContains(setupDB, "better-sqlite3")
	assertNotContains(setupDB, "PostgreSQL database schema ready")

	seed := assertContains(filepath.Join("src", "seed.ts"),
		`import { prisma } from "./db";`,
		`"model": "ticket"`,
		`"types": [`,
		`"boolean"`,
		`await modelDelegate(row).upsert({`,
		`BlackLang seed data applied`,
	)
	assertNotContains(seed, "ON CONFLICT")

	authRoute := assertContains(filepath.Join("src", "routes", "auth.ts"),
		`import { prisma } from "../db";`,
		`await prisma.blackUser.findUnique({ where: { email } })`,
		`void prisma.blackAuditLog.create({`,
	)
	assertNotContains(authRoute, "better-sqlite3")

	apiSmoke := assertContains(filepath.Join("src", "blacklang.api.test.ts"),
		`const { createApp } = await import("./server");`,
	)
	assertNotContains(apiSmoke, "file::memory:")

	secretManifest := assertContains(filepath.Join("security", "secrets.json"),
		`"name": "MYSQL_PASSWORD"`,
		`"kind": "mysql-password"`,
		`"name": "MYSQL_ROOT_PASSWORD"`,
		`"sensitive": true`,
	)
	assertNotContains(secretManifest, "generated.postgres")

	compose := assertContains("docker-compose.yml",
		`depends_on:`,
		`mysql:`,
		`condition: service_healthy`,
		`BLACK_DATABASE_URL: "${BLACK_DATABASE_URL:-mysql://blacklang:blacklang@mysql:3306/blacklang}"`,
		`image: mysql:8.4`,
		`MYSQL_ROOT_PASSWORD: "${MYSQL_ROOT_PASSWORD:-blacklang_root}"`,
		`mysqladmin ping -h 127.0.0.1 -u$${MYSQL_USER:-blacklang} -p$${MYSQL_PASSWORD:-blacklang} --silent`,
	)
	assertNotContains(compose, "/app/data")
	assertNotContains(compose, "postgres:17-alpine")

	previewCompose := assertContains("docker-compose.preview.yml",
		`mysql-preview:`,
		`BLACK_DATABASE_URL: "${BLACK_DATABASE_URL:-mysql://blacklang:blacklang@mysql-preview:3306/blacklang_preview}"`,
		`MYSQL_DATABASE: "${MYSQL_DATABASE:-blacklang_preview}"`,
	)
	assertNotContains(previewCompose, "postgres-preview")
}

func TestBuildWebWritesMigrationRuntime(t *testing.T) {
	source := `app Warehouse

target web {
  frontend react
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  name text required
}

migration RenameProductName {
  rename entity ProductItem to Product
  rename field Product.title to name
}

page Products {
  source Product

  table {
    columns sku, name
  }

  form {
    fields sku, name
  }

  actions create, edit
}
`

	program, parseDiagnostics := Parse("migration.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	validateDiagnostics := Validate(program)
	if len(validateDiagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", validateDiagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	assertContains := func(relativePath string, values ...string) string {
		t.Helper()
		content, err := os.ReadFile(filepath.Join(outDir, relativePath))
		if err != nil {
			t.Fatalf("expected generated %s: %v", relativePath, err)
		}
		text := string(content)
		for _, value := range values {
			if !strings.Contains(text, value) {
				t.Fatalf("expected %s to contain %q, got:\n%s", relativePath, value, text)
			}
		}
		return text
	}

	assertContains(filepath.Join("migrations", "manifest.json"),
		`"targetDatabase": "sqlite"`,
		`"planCommand": "npm run db:migrate:plan"`,
		`"applyCommand": "npm run db:migrate"`,
		`"policy": "plan mode is read-only; apply mode runs only declared rename migrations before generated schema setup"`,
		`"name": "RenameProductName"`,
		`"file": "001_renameproductname.sql"`,
		`"kind": "entity"`,
		`"from": "ProductItem"`,
		`"to": "Product"`,
		`"kind": "field"`,
		`"entity": "Product"`,
		`"oldColumn": "title"`,
		`"newColumn": "name"`,
	)
	assertContains(filepath.Join("migrations", "001_renameproductname.sql"),
		`ALTER TABLE "ProductItem" RENAME TO "Product";`,
		`ALTER TABLE "Product" RENAME COLUMN "title" TO "name";`,
	)
	assertContains(filepath.Join("src", "setup-db.ts"),
		`type BlackMigrationOperation = {`,
		`const blackMigrations: BlackMigration[] = [`,
		`CREATE TABLE IF NOT EXISTS \"BlackMigration\"`,
		`applyBlackMigrations();`,
		`ALTER TABLE " + quoteIdentifier(operation.from) + " RENAME TO " + quoteIdentifier(operation.to)`,
		`ALTER TABLE " + quoteIdentifier(operation.entity) + " RENAME COLUMN " + quoteIdentifier(operation.oldColumn) + " TO " + quoteIdentifier(operation.newColumn)`,
	)
	assertContains(filepath.Join("src", "migrate.ts"),
		`command: "db:migrate:plan"`,
		`command: "db:migrate"`,
		`databaseReachable: true`,
		`plan mode did not create it`,
		`Apply mode runs only declared rename migrations`,
		`ALTER TABLE " + quoteIdentifier(operation.from) + " RENAME TO " + quoteIdentifier(operation.to)`,
		`ALTER TABLE " + quoteIdentifier(operation.entity) + " RENAME COLUMN " + quoteIdentifier(operation.oldColumn) + " TO " + quoteIdentifier(operation.newColumn)`,
	)
	assertContains("package.json",
		`"db:migrate:plan": "tsx src/migrate.ts --plan"`,
		`"db:migrate": "tsx src/migrate.ts --apply"`,
	)
	assertContains(filepath.Join("prisma", "schema.prisma"),
		`model BlackMigration {`,
		`id String @id`,
		`appliedAt DateTime @default(now())`,
	)
}

func TestBuildWebSupportsAPIOnlyTarget(t *testing.T) {
	source := `app InventoryAPI

target api {
  backend node
  database sqlite
}

entity Product {
  sku text required unique
  stock number default 0
}

query LowStockProducts {
  source Product
  where stock < 10
  sort stock asc
  limit 25
}

job LowStockMonitor {
  schedule every 1 hours
  run query LowStockProducts
}

action RestockProduct {
  source Product
  input quantity number required min 1
  set stock = stock + quantity
}

page Products {
  source Product
  query LowStockProducts
  table {
    columns sku, stock
  }
  actions create, edit, RestockProduct
}

api StockWebhook {
  method POST
  path "/api/webhooks/stock"
  body sku text required
  body stock number required min 0
  update Product where sku == body.sku set stock = body.stock
  respond accepted
  public
  webhook
}
`

	program, parseDiagnostics := Parse("test.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}
	outDir := t.TempDir()
	files, diagnostics := BuildWeb(program, outDir)
	if len(diagnostics) != 0 {
		t.Fatalf("expected no build diagnostics, got %#v", diagnostics)
	}

	paths := map[string]bool{}
	for _, file := range files {
		relative, err := filepath.Rel(outDir, file.Path)
		if err != nil {
			t.Fatalf("expected relative path for %s: %v", file.Path, err)
		}
		paths[filepath.ToSlash(relative)] = true
	}
	for _, relativePath := range []string{
		"README.md",
		"package.json",
		"openapi.json",
		"prisma/schema.prisma",
		"jobs/manifest.json",
		"src/worker.ts",
		"src/server.ts",
		"src/types.ts",
		"src/api/product.ts",
		"src/routes/product.ts",
		"src/validation/product.ts",
		"src/blacklang.contract.test.ts",
		"src/blacklang.api.test.ts",
	} {
		if !paths[relativePath] {
			t.Fatalf("expected API-only build to include %s, got %#v", relativePath, paths)
		}
	}
	for _, relativePath := range []string{
		"index.html",
		"vite.config.ts",
		"src/main.tsx",
		"src/App.tsx",
		"src/styles.css",
		"src/vite-env.d.ts",
		"src/pages/ProductsPage.tsx",
		"src/blacklang.frontend.test.tsx",
		"src/blacklang.browser.test.tsx",
		"src/blacklang.e2e.test.ts",
		"src/blacklang.e2e.matrix.ts",
		"tests/browser-matrix.json",
	} {
		if paths[relativePath] {
			t.Fatalf("expected API-only build to omit %s, got %#v", relativePath, paths)
		}
	}

	readme := readGeneratedFile(t, outDir, "README.md")
	for _, value := range []string{`- Target: api`, `- Frontend: none`, `- Jobs: 1`} {
		if !strings.Contains(readme, value) {
			t.Fatalf("expected API-only README to contain %q, got:\n%s", value, readme)
		}
	}

	packageJSON := readGeneratedFile(t, outDir, "package.json")
	for _, value := range []string{
		`"dev": "tsx src/server.ts"`,
		`"build": "npm run db:generate && tsc"`,
		`"test": "npm run db:generate && tsx src/blacklang.contract.test.ts && tsx src/blacklang.api.test.ts"`,
		`"jobs:run": "npm run db:setup && tsx src/worker.ts --once"`,
	} {
		if !strings.Contains(packageJSON, value) {
			t.Fatalf("expected API-only package.json to contain %q, got:\n%s", value, packageJSON)
		}
	}
	for _, value := range []string{`"vite"`, `"react"`, `"react-dom"`, `"@vitejs/plugin-react"`, `"@types/react"`} {
		if strings.Contains(packageJSON, value) {
			t.Fatalf("expected API-only package.json to omit %q, got:\n%s", value, packageJSON)
		}
	}

	server := readGeneratedFile(t, outDir, filepath.Join("src", "server.ts"))
	if strings.Contains(server, "express.static") || strings.Contains(server, "dist\", \"index.html") {
		t.Fatalf("expected API-only server to omit static frontend fallback, got:\n%s", server)
	}
	if strings.Contains(server, `import path from "node:path";`) {
		t.Fatalf("expected API-only server to omit unused path import, got:\n%s", server)
	}
	if !strings.Contains(server, `const openAPISpec = {`) || !strings.Contains(server, `res.json(openAPISpec);`) {
		t.Fatalf("expected API-only server to return embedded OpenAPI JSON, got:\n%s", server)
	}
	if !strings.Contains(server, `res.status(404).json({ error: "Not found" });`) {
		t.Fatalf("expected API-only server to return JSON 404, got:\n%s", server)
	}

	contractTest := readGeneratedFile(t, outDir, filepath.Join("src", "blacklang.contract.test.ts"))
	if strings.Contains(contractTest, "blacklang.frontend") || !strings.Contains(contractTest, `"x-blacklang-jobs"`) {
		t.Fatalf("expected API-only contract test to stay API/job focused, got:\n%s", contractTest)
	}
}

func TestBuildWebGeneratesRelationLoadPolicies(t *testing.T) {
	source := `app RelationLoad

auth {
  strategy emailPassword
  session cookie

  user {
    name text required
    email email required unique
  }
}

entity Customer {
  tenantId text required
  name text required
  email email
  policy tenant tenantId
}

entity Order {
  customer Customer required load detail query label "Customer"
  total money default 0
}

role Admin {
  allow all
}

role Worker {
  allow read Order
  allow read Customer
  deny read Customer email
}

query RecentOrders {
  source Order
  sort total desc
  limit 25
}

page Customers {
  source Customer
  table {
    columns name
  }
  form {
    fields name
  }
  actions create edit
}

page Orders {
  source Order
  query RecentOrders
  table {
    columns customer, total
  }
  form {
    fields customer, total
  }
  actions create edit archive restore
}
`

	program, parseDiagnostics := Parse("relation-load.black", source)
	if len(parseDiagnostics) != 0 {
		t.Fatalf("expected no parse diagnostics, got %#v", parseDiagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("expected no validation diagnostics, got %#v", diagnostics)
	}

	outDir := t.TempDir()
	if _, diagnostics := BuildWeb(program, outDir); len(diagnostics) != 0 {
		t.Fatalf("expected build diagnostics to be empty, got %#v", diagnostics)
	}

	route := readGeneratedFile(t, outDir, filepath.Join("src", "routes", "order.ts"))
	querySegment := generatedSegment(t, route, `orderRouter.get("/orders/query"`, `orderRouter.get("/orders",`)
	listSegment := generatedSegment(t, route, `orderRouter.get("/orders",`, `orderRouter.get("/orders/:id"`)
	detailSegment := generatedSegment(t, route, `orderRouter.get("/orders/:id"`, `orderRouter.post("/orders",`)
	createSegment := generatedSegment(t, route, `orderRouter.post("/orders",`, `orderRouter.put("/orders/:id"`)

	if !strings.Contains(route, `async function attachOrderRelations(req: express.Request, items: any[], context: "list" | "detail" | "query" | "mutation")`) {
		t.Fatalf("expected route to include batch relation attach helper, got:\n%s", route)
	}
	if !strings.Contains(route, `const customerIds = Array.from(new Set(items.map((item) => item.customerId).filter((id): id is string => typeof id === "string" && id.length > 0)));`) ||
		!strings.Contains(route, `const customerRecords = await prisma.customer.findMany({`) ||
		!strings.Contains(route, `const customerMap = new Map(customerRecords.map((record) => [record.id, sanitizeCustomerRelation(record, roles)]));`) {
		t.Fatalf("expected route to batch prefetch customer relations, got:\n%s", route)
	}
	if !strings.Contains(route, `function sanitizeCustomerRelation(item: any, roles: string[])`) ||
		!strings.Contains(route, `if (!canAccessField(roles, "read", "Customer", "email")) delete output.email;`) ||
		!strings.Contains(route, `function rowPolicyWhereForCustomerRelation(req: express.Request, base: Record<string, unknown> = {})`) ||
		!strings.Contains(route, `where: rowPolicyWhereForCustomerRelation(req, { id: { in: customerIds } }) as any`) ||
		!strings.Contains(route, `["detail", "query"].includes(context) && canAccessField(roles, "read", "Order", "customer")`) {
		t.Fatalf("expected route to sanitize nested relation responses, got:\n%s", route)
	}
	if !strings.Contains(querySegment, `await attachOrderRelations(req, items, "query");`) {
		t.Fatalf("expected query endpoint to attach customer relation, got:\n%s", querySegment)
	}
	if strings.Contains(listSegment, `await attachOrderRelations(req, items, "list");`) {
		t.Fatalf("expected base list endpoint to omit customer relation attach, got:\n%s", listSegment)
	}
	if !strings.Contains(detailSegment, `await attachOrderRelations(req, [item], "detail");`) {
		t.Fatalf("expected detail endpoint to attach customer relation, got:\n%s", detailSegment)
	}
	if strings.Contains(createSegment, `await attachOrderRelations(req, [item], "mutation");`) {
		t.Fatalf("expected mutation endpoint to omit customer relation attach, got:\n%s", createSegment)
	}
	if strings.Contains(route, `include: { customer: true }`) {
		t.Fatalf("expected relation load policy to use explicit batch attach instead of Prisma include, got:\n%s", route)
	}

	var spec map[string]any
	if err := json.Unmarshal([]byte(readGeneratedFile(t, outDir, "openapi.json")), &spec); err != nil {
		t.Fatalf("expected valid openapi json: %v", err)
	}
	paths := spec["paths"].(map[string]any)
	queryOp := paths["/api/orders/query"].(map[string]any)["get"].(map[string]any)
	listOp := paths["/api/orders"].(map[string]any)["get"].(map[string]any)
	createOp := paths["/api/orders"].(map[string]any)["post"].(map[string]any)
	detailOp := paths["/api/orders/{id}"].(map[string]any)["get"].(map[string]any)
	if !openAPIRelationLoadIncluded(queryOp, "customer") || openAPIRelationLoadContext(queryOp) != "query" {
		t.Fatalf("expected query OpenAPI relation load metadata, got %#v", queryOp["x-blacklang-relation-load"])
	}
	if openAPIRelationLoadIncluded(listOp, "customer") || openAPIRelationLoadContext(listOp) != "list" {
		t.Fatalf("expected list OpenAPI relation load metadata to exclude customer, got %#v", listOp["x-blacklang-relation-load"])
	}
	if !openAPIRelationLoadIncluded(detailOp, "customer") || openAPIRelationLoadContext(detailOp) != "detail" {
		t.Fatalf("expected detail OpenAPI relation load metadata, got %#v", detailOp["x-blacklang-relation-load"])
	}
	if openAPIRelationLoadIncluded(createOp, "customer") || openAPIRelationLoadContext(createOp) != "mutation" {
		t.Fatalf("expected mutation OpenAPI relation load metadata to exclude customer, got %#v", createOp["x-blacklang-relation-load"])
	}
}

func generatedSegment(t *testing.T, content string, start string, end string) string {
	t.Helper()
	startIndex := strings.Index(content, start)
	if startIndex < 0 {
		t.Fatalf("expected generated content to contain segment start %q, got:\n%s", start, content)
	}
	endIndex := strings.Index(content[startIndex+len(start):], end)
	if endIndex < 0 {
		t.Fatalf("expected generated content to contain segment end %q after %q, got:\n%s", end, start, content[startIndex:])
	}
	return content[startIndex : startIndex+len(start)+endIndex]
}

func openAPIRelationLoadContext(operation map[string]any) string {
	value, _ := operation["x-blacklang-relation-load-context"].(string)
	return value
}

func openAPIRelationLoadIncluded(operation map[string]any, fieldName string) bool {
	values, _ := operation["x-blacklang-relation-load"].([]any)
	for _, value := range values {
		item, _ := value.(map[string]any)
		field, _ := item["field"].(string)
		if field != fieldName {
			continue
		}
		included, _ := item["included"].(bool)
		return included
	}
	return false
}

func readGeneratedFile(t *testing.T, outDir string, relativePath string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(outDir, relativePath))
	if err != nil {
		t.Fatalf("expected generated file %s: %v", relativePath, err)
	}
	return string(content)
}
