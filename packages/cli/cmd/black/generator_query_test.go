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
  aggregate lowStockCount count
  aggregate totalStock sum stock
  aggregate averagePrice avg price
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
	apiSmoke := readQueryGenerated(t, outDir, "src/blacklang.api.test.ts")
	for _, pair := range [][2]string{
		{baseRoute, `productRouter.use("/products", requirePageAccess(["Admin"]));`},
		{queryRoute, `productRouter.use("/lowstock", requirePageAccess(["Admin", "Worker"]));`},
		{queryRoute, `productRouter.get("/lowstock/query", requirePermission("read", "Product")`},
		{queryRoute, `productRouter.get("/lowstock/query/summary", requirePermission("read", "Product")`},
		{queryRoute, `const summary = await productModel.aggregate({`},
		{queryRoute, `_count: { _all: true }`},
		{queryRoute, `_sum: { "stock": true }`},
		{queryRoute, `_avg: { "price": true }`},
		{queryRoute, `["stock", "active", "price"].every((field) => canAccessField(currentRoles(req), "read", "Product", field))`},
		{queryClient, `queryList: (includeArchived = false)`},
		{queryClient, `querySummary: (includeArchived = false)`},
		{page, `import { productApi } from "../api/product.lowstock";`},
		{page, `productApi.queryList(showArchived)`},
		{page, `productApi.querySummary(showArchived)`},
		{page, `className="query-summary"`},
		{page, `formatQuerySummaryValue(querySummary["lowStockCount"])`},
		{orders, `import { productApi } from "../api/product";`},
		{orders, `productApi.list()`},
		{server, `import { productRouter as page0Router } from "./routes/product.lowstock";`},
		{server, `import { productRouter as page1Router } from "./routes/product";`},
		{apiSmoke, `const lowStockProductsSummaryResponse = await fetch(baseURL + "/api/lowstock/query/summary");`},
		{apiSmoke, `assert.equal(lowStockProductsSummaryResponse.status, 401, "LowStockProducts summary should reject anonymous requests");`},
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
	summaryOp := spec.Paths["/api/lowstock/query/summary"]["get"]
	if summaryOp["x-blacklang-query-summary"] != "LowStockProducts" || summaryOp["security"] == nil {
		t.Fatalf("query summary contract must describe its runtime: %#v", summaryOp)
	}
	aggregates, ok := summaryOp["x-blacklang-aggregates"].([]any)
	if !ok || len(aggregates) != 3 {
		t.Fatalf("query summary contract must include aggregate metadata: %#v", summaryOp)
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
	for _, value := range []string{"[showArchived, queryRevision]", "setQueryRevision((current) => current + 1);", "setItems([]);", "setSelectedIds([]);", `setQuerySummary({ "lowStockCount": null, "totalStock": null, "averagePrice": null });`} {
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
let permission, calls = 0, aggregateCalls = 0, lastArgs, lastAggregateArgs;
const handlers = {};
const rows = [{ id: "a", stock: 1, active: true, price: 9 }];
const productModel = {
  async findMany(args) { calls++; lastArgs = args; return rows; },
  async aggregate(args) { aggregateCalls++; lastAggregateArgs = args; return { _count: { _all: 3 }, _sum: { stock: 11 }, _avg: { price: 7.5 } }; }
};
const productRouter = { get(path, guard, callback) { handlers[path] = callback; if (path === "/lowstock/query") permission = guard; } };
function requirePermission(action, resource) { assert.equal(action, "read"); assert.equal(resource, "Product"); return "entityReadGuard"; }
function currentRoles(req) { return [req.role]; }
function canAccessField(roles, _action, _resource, field) { return !roles.includes("Worker") || field !== "price"; }
function sanitizeProduct(item, roles) { return { ...item, sanitizedFor: roles.join(",") }; }
` + route + `
async function run(role, query) {
  const response = { statusCode: 200, body: null, status(value) { this.statusCode = value; return this; }, json(value) { this.body = value; } };
  await handlers["/lowstock/query"]({ role, query }, response);
  return response;
}
async function runSummary(role, query) {
  const response = { statusCode: 200, body: null, status(value) { this.statusCode = value; return this; }, json(value) { this.body = value; } };
  await handlers["/lowstock/query/summary"]({ role, query }, response);
  return response;
}

(async () => {
  assert.equal(permission, "entityReadGuard");
  assert.equal(typeof handlers["/lowstock/query/summary"], "function");
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
  response = await runSummary("Admin", {});
  assert.equal(response.statusCode, 200);
  assert.deepEqual(response.body, { lowStockCount: 3, totalStock: 11, averagePrice: 7.5 });
  assert.deepEqual(lastAggregateArgs.where.AND.slice(0, 4), [{ archivedAt: null }, { stock: { lt: 10 } }, { stock: { gte: 0 } }, { active: { equals: true } }]);
  assert.deepEqual(lastAggregateArgs._count, { _all: true });
  assert.deepEqual(lastAggregateArgs._sum, { stock: true });
  assert.deepEqual(lastAggregateArgs._avg, { price: true });
  assert.equal(lastAggregateArgs.orderBy, undefined);
  assert.equal(lastAggregateArgs.take, undefined);
  const aggregateCallsBeforeDenied = aggregateCalls;
  response = await runSummary("Worker", {});
  assert.equal(response.statusCode, 403);
  assert.equal(aggregateCalls, aggregateCallsBeforeDenied, "hidden aggregate field must deny before any database query");
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
let effect, items = [{id:"old"}], selectedIds = ["old"], querySummary = { old: 1 }, loading = false, error = null;
let showArchived = false, queryRevision = 0;
const queryRequestVersion = {current: 0};
const listRequests = [];
const summaryRequests = [];
const emptySummary = { "lowStockCount": null, "totalStock": null, "averagePrice": null };
const productApi = {
  queryList(archived) {return new Promise((resolve,reject) => listRequests.push({resolve,reject,archived}));},
  querySummary(archived) {return new Promise((resolve,reject) => summaryRequests.push({resolve,reject,archived}));}
};
function useEffect(callback) {effect = callback;}
function setItems(value) {items = value;}
function setSelectedIds(value) {selectedIds = value;}
function setQuerySummary(value) {querySummary = value;}
function setLoading(value) {loading = value;}
function setError(value) {error = value;}
function setQueryRevision(callback) {queryRevision = callback(queryRevision);}
` + effect + `
const flush = () => new Promise(resolve => setImmediate(resolve));
(async () => {
  let cleanup = effect();
  assert.deepEqual(items, []);
  assert.deepEqual(selectedIds, []);
  assert.deepEqual(querySummary, emptySummary);
  refreshQuery();
  // React has not yet run the previous effect's cleanup. Immediate invalidation
  // must still prevent this response from restoring records changed by a mutation.
  listRequests[0].resolve([{id:"pre-mutation"}]);
  summaryRequests[0].resolve({ lowStockCount: 99, totalStock: 99, averagePrice: 99 });
  await flush();
  assert.deepEqual(items, []);
  assert.deepEqual(querySummary, emptySummary);
  assert.equal(loading, true);
  cleanup();
  cleanup = effect();
  summaryRequests[1].resolve({ lowStockCount: 1, totalStock: 3, averagePrice: 4.5 });
  listRequests[1].resolve([{id:"current"}]);
  await flush();
  assert.deepEqual(items, [{id:"current"}]);
  assert.deepEqual(querySummary, { lowStockCount: 1, totalStock: 3, averagePrice: 4.5 });
  assert.equal(loading, false);
  // An archive toggle followed by denial must not retain the previous list.
  cleanup();
  showArchived = true;
  cleanup = effect();
  assert.equal(listRequests[2].archived, true);
  assert.equal(summaryRequests[2].archived, true);
  listRequests[2].reject(new Error("Forbidden"));
  await flush();
  assert.deepEqual(items, []);
  assert.deepEqual(querySummary, emptySummary);
  assert.equal(error, "Forbidden");
  assert.equal(loading, false);
  cleanup();
  cleanup = effect();
  cleanup();
  listRequests[3].resolve([{id:"after-unmount"}]);
  summaryRequests[3].resolve({ lowStockCount: 2, totalStock: 2, averagePrice: 2 });
  await flush();
  assert.deepEqual(items, []);
  assert.deepEqual(querySummary, emptySummary);
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
