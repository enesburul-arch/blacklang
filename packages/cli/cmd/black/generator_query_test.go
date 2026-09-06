package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const generatorQuerySource = `app Inventory
auth {
  strategy emailPassword
  session cookie
  user {
    name text required
    email email required unique
  }
}
role Admin {
  allow all
}
role Worker {
  allow read Product
  deny read Product price
}
entity Product {
  name text required
  stock number default 0
  price money default 0
  active boolean default true
  received date optional
  updated datetime optional
  status text default draft
}
entity Order {
  product Product required
}
query LowStockProducts {
  source Product
  where stock < 10
  where stock >= 0
  where active == true
  sort price asc
  limit 2
}
query UnusedProducts {
  source Product
}
page LowStock {
  source Product
  query LowStockProducts
  access Admin, Worker
  table {
    columns name, stock
  }
  form {
    fields name, stock, price, active, status
  }
  actions create, edit, delete, archive, restore
}
page Products {
  source Product
  access Admin
  table {
    columns name, stock
  }
  form {
    fields name, stock, price, active, status
  }
  actions create, edit, delete
}
page Orders {
  source Order
  table {
    columns product
  }
  form {
    fields product
  }
  actions create
}
workflow ProductLifecycle {
  source Product
  states draft, ready
  transition markReady {
    from draft
    to ready
    allow Admin
  }
}
`

func queryGeneratorFixture(t *testing.T) (*webGenerator, string) {
	t.Helper()
	program, diagnostics := Parse("query.black", generatorQuerySource)
	if len(diagnostics) != 0 {
		t.Fatalf("parse query fixture: %+v", diagnostics)
	}
	if diagnostics := Validate(program); len(diagnostics) != 0 {
		t.Fatalf("validate query fixture: %+v", diagnostics)
	}
	outDir := t.TempDir()
	files, diagnostics := BuildWeb(program, outDir)
	if len(diagnostics) != 0 {
		t.Fatalf("generate query fixture: %+v", diagnostics)
	}
	seen := map[string]bool{}
	for _, file := range files {
		if seen[file.Path] {
			t.Fatalf("duplicate generated path: %s", file.Path)
		}
		seen[file.Path] = true
	}
	return &webGenerator{program: program}, outDir
}

func readQueryGenerated(t *testing.T, outDir, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(outDir, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestQueryPagesHaveIndependentRoutesAndPreserveRelationLists(t *testing.T) {
	_, outDir := queryGeneratorFixture(t)
	baseRoute := readQueryGenerated(t, outDir, "src/routes/product.ts")
	queryRoute := readQueryGenerated(t, outDir, "src/routes/product.lowstock.ts")
	baseClient := readQueryGenerated(t, outDir, "src/api/product.ts")
	queryClient := readQueryGenerated(t, outDir, "src/api/product.lowstock.ts")
	page := readQueryGenerated(t, outDir, "src/pages/LowStockPage.tsx")
	orders := readQueryGenerated(t, outDir, "src/pages/OrdersPage.tsx")
	server := readQueryGenerated(t, outDir, "src/server.ts")
	for _, pair := range [][2]string{
		{baseRoute, `productRouter.use("/products", requirePageAccess(["Admin"]));`},
		{queryRoute, `productRouter.use("/lowstock", requirePageAccess(["Admin", "Worker"]));`},
		{queryRoute, `productRouter.get("/lowstock/query", requirePermission("read", "Product")`},
		{queryRoute, `["stock", "active", "price"].every((field) => canAccessField(currentRole(req), "read", "Product", field))`},
		{queryClient, `queryList: (includeArchived = false)`},
		{page, `import { productApi } from "../api/product.lowstock";`},
		{page, `productApi.queryList(showArchived)`},
		{orders, `import { productApi } from "../api/product";`},
		{orders, `productApi.list()`},
		{server, `import { productRouter as page0Router } from "./routes/product.lowstock";`},
		{server, `import { productRouter as page1Router } from "./routes/product";`},
	} {
		if !strings.Contains(pair[0], pair[1]) {
			t.Errorf("missing generated wiring: %s", pair[1])
		}
	}
	if strings.Contains(baseRoute, "/query") || strings.Contains(baseClient, "queryList") {
		t.Fatal("ordinary entity list was replaced by query behavior")
	}
	if strings.Index(queryRoute, `get("/lowstock/query"`) > strings.Index(queryRoute, `get("/lowstock/:id"`) {
		t.Fatal("query endpoint must be registered before the dynamic detail endpoint")
	}
	if !strings.Contains(queryRoute, `where: includeArchived ? {} : { archivedAt: null },`) {
		t.Fatal("base list must remain available on a query-bound page")
	}
	if strings.Contains(server, "UnusedProducts") || strings.Contains(queryRoute, "UnusedProducts") {
		t.Fatal("unbound query unexpectedly exposes an endpoint")
	}
	var spec struct {
		Paths map[string]map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal([]byte(readQueryGenerated(t, outDir, "openapi.json")), &spec); err != nil {
		t.Fatal(err)
	}
	op := spec.Paths["/api/lowstock/query"]["get"]
	if op["x-blacklang-query"] != "LowStockProducts" || op["x-blacklang-limit"] != float64(2) || op["security"] == nil {
		t.Fatalf("query contract must describe its runtime: %#v", op)
	}
	if _, ok := spec.Paths["/api/products/query"]; ok {
		t.Fatal("unbound page must not expose query contract")
	}
}

func TestQueryPagesInvalidateEveryMutation(t *testing.T) {
	_, outDir := queryGeneratorFixture(t)
	page := readQueryGenerated(t, outDir, "src/pages/LowStockPage.tsx")
	// create/update, single/bulk delete, archive/restore and workflow each cause
	// the server to recompute membership, ordering and the bounded window.
	if got := strings.Count(page, "refreshQuery();"); got != 7 {
		t.Fatalf("expected all seven mutation paths to refresh, got %d", got)
	}
	for _, value := range []string{"[showArchived, queryRevision]", "setQueryRevision((current) => current + 1);", "setItems([]);", "setSelectedIds([]);"} {
		if !strings.Contains(page, value) {
			t.Errorf("missing query refresh lifecycle %q", value)
		}
	}
	for _, value := range []string{"setItems((current)", "setItems((current) => [...current, saved])"} {
		if strings.Contains(page, value) {
			t.Fatalf("query pages must not locally guess membership after mutation: %s", value)
		}
	}
}

func TestQueryOnlyPagePreservesBaseListAndOmitsUndeclaredMutations(t *testing.T) {
	program, diagnostics := Parse("single.black", "app Single\nentity Product {\n stock number\n}\nquery AllProducts {\n source Product\n}\npage Products {\n source Product\n query AllProducts\n table {\n columns stock\n }\n}\n")
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	g := webGenerator{program: program}
	route := g.route(program.Pages[0], program.Entities[0])
	for _, required := range []string{`get("/products/query"`, `get("/products"`, `get("/products/:id"`, `orderBy: [{ id: "asc" }]`, `take: 100`} {
		if !strings.Contains(route, required) {
			t.Errorf("missing %s", required)
		}
	}
	for _, denied := range []string{"Router.post(", "Router.put(", "Router.delete("} {
		if strings.Contains(route, denied) {
			t.Errorf("undeclared mutation route %s must match its absent OpenAPI operation", denied)
		}
	}
}

func TestGeneratedQueryRouteExecutesTypedFiltersAndPermissionGuards(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node.js unavailable; generated runtime execution requires Node.js")
	}
	g, _ := queryGeneratorFixture(t)
	query := &g.program.Queries[0]
	query.Where = append(query.Where,
		QueryFilterDecl{Field: "name", Operator: "!=", Value: QueryLiteral{Kind: "string", Value: "quote\"\\\n</script>"}},
		QueryFilterDecl{Field: "received", Operator: ">=", Value: QueryLiteral{Kind: "string", Value: "2026-09-01"}},
		QueryFilterDecl{Field: "updated", Operator: "<=", Value: QueryLiteral{Kind: "string", Value: "2026-09-06T11:00:00+03:00"}},
	)
	route := g.queryRoute(g.program.Pages[0], g.program.Entities[0])
	script := `const assert = require("node:assert/strict");
let handler, permission, calls = 0, lastArgs;
const rows = [{ id: "a", stock: 1, active: true, price: 9 }];
const productModel = { async findMany(args) { calls++; lastArgs = args; return rows; } };
const productRouter = { get(path, guard, callback) { assert.equal(path, "/lowstock/query"); permission = guard; handler = callback; } };
function requirePermission(action, resource) { assert.equal(action, "read"); assert.equal(resource, "Product"); return "entityReadGuard"; }
function currentRole(req) { return req.role; }
function canAccessField(role, _action, _resource, field) { return role !== "Worker" || field !== "price"; }
function sanitizeProduct(item, role) { return { ...item, sanitizedFor: role }; }
` + route + `
async function run(role, query) {
  const response = { statusCode: 200, body: null, status(value) { this.statusCode = value; return this; }, json(value) { this.body = value; } };
  await handler({ role, query }, response);
  return response;
}

(async () => {
  assert.equal(permission, "entityReadGuard");
  let response = await run("Admin", { stock: "999", limit: "999999", sort: "name", where: "injected" });
  assert.equal(response.statusCode, 200);
  assert.equal(response.body[0].sanitizedFor, "Admin");
  assert.equal(lastArgs.take, 2);
  assert.deepEqual(lastArgs.orderBy, [{ price: "asc" }, { id: "asc" }]);
  const filters = lastArgs.where.AND;
  assert.deepEqual(filters.slice(0, 4), [{ archivedAt: null }, { stock: { lt: 10 } }, { stock: { gte: 0 } }, { active: { equals: true } }]);
  assert.equal(filters[4].name.not, "quote\"\\\n</script>");
  assert.equal(filters[5].received.gte.toISOString(), "2026-09-01T00:00:00.000Z");
  assert.equal(filters[6].updated.lte.toISOString(), "2026-09-06T08:00:00.000Z");
  response = await run("Admin", { archived: "all" });
  assert.deepEqual(lastArgs.where.AND[0], {});
  assert.equal(lastArgs.where.AND.length, 7);
  const callsBeforeDenied = calls;
  response = await run("Worker", {});
  assert.equal(response.statusCode, 403);
  assert.equal(calls, callsBeforeDenied, "hidden sort field must deny before any database query");
  console.log("query runtime passed");
})().catch((error) => { console.error(error); process.exitCode = 1; });
`
	path := filepath.Join(t.TempDir(), "query-runtime.cjs")
	if err := os.WriteFile(path, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(node, path).CombinedOutput(); err != nil {
		t.Fatalf("generated query runtime failed: %v\n%s", err, output)
	}
}

func TestGeneratedQueryEffectRejectsStaleLoadsAndClearsDeniedReload(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node.js unavailable; generated runtime execution requires Node.js")
	}
	_, outDir := queryGeneratorFixture(t)
	page := readQueryGenerated(t, outDir, "src/pages/LowStockPage.tsx")
	start := strings.Index(page, "  useEffect(() => {")
	endMarker := "  function refreshQuery() {"
	end := strings.Index(page, endMarker)
	refreshEnd := strings.Index(page[end:], "\n  }\n") + end + len("\n  }\n")
	if start < 0 || end < 0 || refreshEnd <= end {
		t.Fatal("missing generated query effect and invalidation function")
	}
	// This block has only one erased TypeScript annotation and no JSX. Execute
	// the actual generated effect independently of React's rendering machinery.
	effect := strings.ReplaceAll(page[start:refreshEnd], "reason: unknown", "reason")
	script := `const assert = require("node:assert/strict");
let effect, items = [{id:"old"}], selectedIds = ["old"], loading = false, error = null;
let showArchived = false, queryRevision = 0;
const queryRequestVersion = {current: 0};
const requests = [];
const productApi = {queryList(archived) {return new Promise((resolve,reject) => requests.push({resolve,reject,archived}));}};
function useEffect(callback) {effect = callback;}
function setItems(value) {items = value;}
function setSelectedIds(value) {selectedIds = value;}
function setLoading(value) {loading = value;}
function setError(value) {error = value;}
function setQueryRevision(callback) {queryRevision = callback(queryRevision);}
` + effect + `
const flush = () => new Promise(resolve => setImmediate(resolve));
(async () => {
  let cleanup = effect();
  assert.deepEqual(items, []);
  assert.deepEqual(selectedIds, []);
  refreshQuery();
  // React has not yet run the previous effect's cleanup. Immediate invalidation
  // must still prevent this response from restoring records changed by a mutation.
  requests[0].resolve([{id:"pre-mutation"}]);
  await flush();
  assert.deepEqual(items, []);
  assert.equal(loading, true);
  cleanup();
  cleanup = effect();
  requests[1].resolve([{id:"current"}]);
  await flush();
  assert.deepEqual(items, [{id:"current"}]);
  assert.equal(loading, false);
  // An archive toggle followed by denial must not retain the previous list.
  cleanup();
  showArchived = true;
  cleanup = effect();
  assert.equal(requests[2].archived, true);
  requests[2].reject(new Error("Forbidden"));
  await flush();
  assert.deepEqual(items, []);
  assert.equal(error, "Forbidden");
  assert.equal(loading, false);
  cleanup();
  cleanup = effect();
  cleanup();
  requests[3].resolve([{id:"after-unmount"}]);
  await flush();
  assert.deepEqual(items, []);
  console.log("query effect lifecycle passed");
})().catch(error => {console.error(error); process.exitCode = 1;});
`
	path := filepath.Join(t.TempDir(), "query-effect.cjs")
	if err := os.WriteFile(path, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(node, path).CombinedOutput(); err != nil {
		t.Fatalf("generated query effect failed: %v\n%s", err, output)
	}
}
