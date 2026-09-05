package main

import (
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
}

entity Product {
  sku text required unique length 3..40 regex "^[A-Z0-9]+$" message "Use uppercase letters and numbers"
  name text required label "Product Name" placeholder "Enter product name" help "Visible product name"
  stock number default 0 min 0 label "Stock Count" placeholder "Enter stock count"
  price money min 0 label "Unit Price" placeholder "Enter unit price"
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
  webhook
  public
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
	if len(files) != 37 {
		t.Fatalf("expected 37 generated files, got %d", len(files))
	}

	expected := []string{
		"README.md",
		".env.example",
		".dockerignore",
		"Dockerfile",
		"docker-compose.yml",
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
		filepath.Join("src", "server.ts"),
		filepath.Join("src", "styles.css"),
		filepath.Join("src", "vite-env.d.ts"),
		filepath.Join("src", "types.ts"),
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
		`DATABASE_URL="file:./dev.db"`,
		`CORS_ORIGINS="http://localhost:5173"`,
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
		`- Target: web`,
		`- Frontend: react`,
		`- Backend: node`,
		`- Database: sqlite`,
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
		`warehouse-data:/app/data`,
	} {
		if !strings.Contains(composeText, value) {
			t.Fatalf("expected docker-compose.yml to contain %q, got:\n%s", value, composeText)
		}
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
		`"x-blacklang-webhook": true`,
		`"name": "warehouseId"`,
		`"name": "limit"`,
		`"patch"`,
		`"OrderInput"`,
		`"StockWebhookInput"`,
		`"customerId"`,
		`"format": "email"`,
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
		`app.use(express.json({ limit: "100kb" }));`,
		`app.use("/api/auth", authRouter);`,
		`app.use("/api", requireAuth);`,
		`app.use("/api", requireCsrf);`,
		`app.get("/openapi.json", (_req, res) => {`,
		`res.sendFile(path.join(process.cwd(), "openapi.json"));`,
		`app.use(express.static(path.join(process.cwd(), "dist")));`,
		`res.sendFile(path.join(process.cwd(), "dist", "index.html"));`,
	} {
		if !strings.Contains(serverText, value) {
			t.Fatalf("expected server to contain %q, got:\n%s", value, serverText)
		}
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
		`SELECT u.id, u.name, u.email, u.role, u.passwordHash FROM \"BlackSession\" s JOIN \"BlackUser\" u ON u.id = s.userId WHERE s.id = ? AND s.expiresAt > datetime('now')`,
		`export function requirePageAccess(allowedRoles: string[]) {`,
		`role: user.role`,
		`authRouter.get("/users", requireAuth, requirePageAccess([defaultRole]), (_req, res) => {`,
		`authRouter.get("/audit", requireAuth, requirePageAccess([defaultRole]), (_req, res) => {`,
		`authRouter.put("/users/:id/role", requireAuth, requireCsrf, requirePageAccess([defaultRole]), (req, res) => {`,
		`export function writeAuditLog(actor: ReturnType<typeof publicUser> | undefined, action: string, resource: string, resourceId: string, summary = "") {`,
		`writeAuditLog((req as any).blackUser, "role.update", "BlackUser", user.id, "Role changed to " + role);`,
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
		"customerId String",
		`customer Customer @relation("Order_customer", fields: [customerId], references: [id])`,
		`orderCustomerItems Order[] @relation("Order_customer")`,
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
		`CREATE TABLE IF NOT EXISTS "BlackSession"`,
		`"userId" TEXT NOT NULL REFERENCES "BlackUser"("id")`,
		`CREATE TABLE IF NOT EXISTS "BlackAuditLog"`,
		`"actorUserId" TEXT NOT NULL`,
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
		`const visiblePages = currentUser ? pages.filter((item) => item.access.length === 0 || item.access.includes("authenticated") || item.access.includes(currentUser.role)) : pages;`,
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
		`<main className="page-view page-view-products">`,
		`<section className="panel bl-view-section-detail">`,
		`import { StockBadge } from "../components/StockBadge";`,
		`{visibleColumns.stock && permissions.fields.stock !== false && <td>{<StockBadge stock={Number(item.stock ?? 0)} />}</td>}`,
		`{permissions.fields.stock !== false && <div><dt>Stock Count</dt><dd>{<StockBadge stock={Number(selectedItem.stock ?? 0)} />}</dd></div>}`,
		`<input type="text" required minLength={3} maxLength={40} pattern="^[A-Z0-9]+$" value={form.sku}`,
		`<input type="number" min="0" placeholder="Enter stock count" value={form.stock}`,
		`errors.sku = "Use uppercase letters and numbers";`,
		`if (!errors.sku && !(new RegExp("^[A-Z0-9]+$")).test(form.sku)) errors.sku = "Use uppercase letters and numbers";`,
		`errors.stock = "Stock Count must be at least 0";`,
		`<span className="field-preview"><StockBadge stock={Number(form.stock || 0)} /></span>`,
	} {
		if !strings.Contains(productPageText, value) {
			t.Fatalf("expected ProductsPage to contain %q, got:\n%s", value, productPageText)
		}
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
		`productRouter.use(requirePageAccess(["Admin"]));`,
		`function sanitizeProduct(item: any, role: string) {`,
		`if (!canAccessField(role, "read", "Product", "price")) delete output.price;`,
		`const writableValue = filterWritableFields(currentRole(req), "update", "Product", validation.value as Record<string, unknown>);`,
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
		`if (!user || !["Admin"].includes(user.role)) {`,
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
		"main > header {",
		"/* Generated from BlackLang page view order. */",
		".page-view-products .bl-view-section-form {",
		"order: 1;",
		".page-view-products .bl-view-section-table {",
		"order: 2;",
		".page-view-products .bl-view-section-detail {",
		"order: 3;",
		"@media (max-width: 760px) {",
		".app-shell.nav-open .app-sidebar {",
	} {
		if !strings.Contains(stylesText, value) {
			t.Fatalf("expected styles to contain %q, got:\n%s", value, stylesText)
		}
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

entity Product {
  name text required
}

page Products {
  source Product

  table {
    columns name
    search name
    filter name
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
		`Ürün Adı Filter`,
		`Filter Ürün Adı`,
		`<th>Ürün Adı</th>`,
		`<dt>Ürün Adı</dt>`,
		`
              Ürün Adı
              <input`,
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
	if !strings.Contains(serverText, `import { purchaseOrderRouter } from "./routes/purchaseorder";`) {
		t.Fatalf("expected server to import lower-camel router, got:\n%s", serverText)
	}
	if !strings.Contains(serverText, `app.use("/api", purchaseOrderRouter);`) {
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
