package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// The canonical page keeps the existing entity module path for relation clients.
// Additional pages get their own modules so their routes and access never overwrite it.
func (g *webGenerator) pageModuleName(page PageDecl) string {
	name := strings.ToLower(page.Source)
	if canonical, ok := g.pageForEntity(page.Source); ok && canonical.Name != page.Name {
		return name + "." + strings.ToLower(page.Name)
	}
	return name
}

func (g *webGenerator) pageRouterName(page PageDecl) string {
	for index, candidate := range g.program.Pages {
		if candidate.Name == page.Name {
			return fmt.Sprintf("page%dRouter", index)
		}
	}
	return lowerCamelCase(page.Source) + "Router"
}

func (g *webGenerator) queryRoute(page PageDecl, entity EntityDecl) string {
	query, ok := findQuery(g.program, page.Query)
	if !ok {
		return ""
	}
	identifier := lowerCamelCase(entity.Name)
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%sRouter.get(\"/%s/query\", %sasync (req, res) => {\n", identifier, strings.ToLower(page.Name), g.permissionMiddleware("read", entity.Name)))
	if g.hasRuntimePermissions() {
		fields := queryFieldNames(query)
		if len(fields) > 0 {
			builder.WriteString(fmt.Sprintf("  if (!%s.every((field) => canAccessField(currentRoles(req), \"read\", %q, field))) {\n", tsStringArrayLiteral(fields), entity.Name))
			builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n")
		}
	}
	builder.WriteString("  const includeArchived = req.query.archived === \"all\";\n")
	builder.WriteString(fmt.Sprintf("  const items = await %sModel.findMany({\n", identifier))
	builder.WriteString("    where: { AND: [\n")
	builder.WriteString(fmt.Sprintf("      %s,\n", g.rowPolicyWhere(entity, "includeArchived ? {} : { archivedAt: null }")))
	for _, filter := range query.Where {
		field, _ := findField(entity, filter.Field)
		builder.WriteString(fmt.Sprintf("      { [%q]: { %s: %s } },\n", filter.Field, queryPrismaOperator(filter.Operator), queryPrismaLiteral(filter.Value, field)))
	}
	builder.WriteString("    ] },\n")
	builder.WriteString("    orderBy: [")
	if query.Sort.Field != "" {
		builder.WriteString(fmt.Sprintf("{ [%q]: %q }, ", query.Sort.Field, query.Sort.Direction))
	}
	builder.WriteString("{ id: \"asc\" }],\n")
	builder.WriteString(fmt.Sprintf("    take: %d\n", effectiveQueryLimit(query)))
	builder.WriteString("  });\n")
	builder.WriteString(g.relationLoadAttachStatement(entity, "query", "items", "  "))
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  res.json(items.map((item) => sanitize%s(item, currentRoles(req))));\n", entity.Name))
	} else {
		builder.WriteString("  res.json(items);\n")
	}
	builder.WriteString("});\n\n")
	builder.WriteString(g.querySummaryRoute(page, entity, query))
	return builder.String()
}

func (g *webGenerator) querySummaryRoute(page PageDecl, entity EntityDecl, query QueryDecl) string {
	if len(query.Aggregates) == 0 {
		return ""
	}
	identifier := lowerCamelCase(entity.Name)
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%sRouter.get(\"/%s/query/summary\", %sasync (req, res) => {\n", identifier, strings.ToLower(page.Name), g.permissionMiddleware("read", entity.Name)))
	if g.hasRuntimePermissions() {
		fields := queryFieldNames(query)
		if len(fields) > 0 {
			builder.WriteString(fmt.Sprintf("  if (!%s.every((field) => canAccessField(currentRoles(req), \"read\", %q, field))) {\n", tsStringArrayLiteral(fields), entity.Name))
			builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n")
		}
	}
	builder.WriteString("  const includeArchived = req.query.archived === \"all\";\n")
	builder.WriteString(fmt.Sprintf("  const summary = await %sModel.aggregate({\n", identifier))
	builder.WriteString("    where: { AND: [\n")
	builder.WriteString(fmt.Sprintf("      %s,\n", g.rowPolicyWhere(entity, "includeArchived ? {} : { archivedAt: null }")))
	for _, filter := range query.Where {
		field, _ := findField(entity, filter.Field)
		builder.WriteString(fmt.Sprintf("      { [%q]: { %s: %s } },\n", filter.Field, queryPrismaOperator(filter.Operator), queryPrismaLiteral(filter.Value, field)))
	}
	builder.WriteString("    ] }")
	clauses := queryAggregatePrismaClauses(query)
	if len(clauses) > 0 {
		builder.WriteString(",\n")
		for index, clause := range clauses {
			suffix := ","
			if index == len(clauses)-1 {
				suffix = ""
			}
			builder.WriteString(fmt.Sprintf("    %s%s\n", clause, suffix))
		}
	} else {
		builder.WriteString("\n")
	}
	builder.WriteString("  });\n")
	builder.WriteString("  res.json({\n")
	for _, aggregate := range query.Aggregates {
		builder.WriteString(fmt.Sprintf("    %q: %s,\n", aggregate.Name, queryAggregateResultExpression(aggregate)))
	}
	builder.WriteString("  });\n")
	builder.WriteString("});\n\n")
	return builder.String()
}

func queryAggregatePrismaClauses(query QueryDecl) []string {
	clauses := []string{}
	if queryHasCountAggregate(query) {
		clauses = append(clauses, "_count: { _all: true }")
	}
	for _, function := range []string{"sum", "avg", "min", "max"} {
		fields := []string{}
		for _, aggregate := range query.Aggregates {
			if aggregate.Function != function || aggregate.Field == "" || containsString(fields, aggregate.Field) {
				continue
			}
			fields = append(fields, aggregate.Field)
		}
		if len(fields) == 0 {
			continue
		}
		parts := []string{}
		for _, field := range fields {
			parts = append(parts, fmt.Sprintf("%q: true", field))
		}
		clauses = append(clauses, fmt.Sprintf("_%s: { %s }", function, strings.Join(parts, ", ")))
	}
	return clauses
}

func queryHasCountAggregate(query QueryDecl) bool {
	for _, aggregate := range query.Aggregates {
		if aggregate.Function == "count" {
			return true
		}
	}
	return false
}

func queryAggregateResultExpression(aggregate QueryAggregateDecl) string {
	if aggregate.Function == "count" {
		return "summary._count?._all ?? 0"
	}
	return fmt.Sprintf("summary._%s?.[%q] ?? null", aggregate.Function, aggregate.Field)
}

func queryPrismaOperator(operator string) string {
	return map[string]string{"==": "equals", "!=": "not", "<": "lt", "<=": "lte", ">": "gt", ">=": "gte"}[operator]
}

func queryPrismaLiteral(literal QueryLiteral, field FieldDecl) string {
	quoted, _ := json.Marshal(literal.Value)
	switch literal.Kind {
	case "number":
		// Emit the validated numeric literal as an explicit numeric conversion.
		return "Number(" + string(quoted) + ")"
	case "boolean":
		return literal.Value
	default:
		if field.Type == "date" || field.Type == "datetime" {
			return "new Date(" + string(quoted) + ")"
		}
		return string(quoted)
	}
}

func effectiveQueryLimit(query QueryDecl) int {
	if query.Limit == 0 {
		return 100
	}
	return query.Limit
}

func (g *webGenerator) queryClientMethod(page PageDecl, entity EntityDecl) string {
	if page.Query == "" {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("  queryList: (includeArchived = false) =>\n    request<%s[]>(endpoint + \"/query\" + (includeArchived ? \"?archived=all\" : \"\")),\n", entity.Name))
	if query, ok := findQuery(g.program, page.Query); ok && len(query.Aggregates) > 0 {
		builder.WriteString("  querySummary: (includeArchived = false) =>\n    request<Record<string, number | null>>(endpoint + \"/query/summary\" + (includeArchived ? \"?archived=all\" : \"\")),\n")
	}
	return builder.String()
}

// Query membership, stable order, and the bounded window must be recomputed by
// the server after every mutation. The effect revision also cancels stale loads.
func queryPageMutation(page PageDecl, ordinaryMutation string) string {
	if page.Query == "" {
		return ordinaryMutation
	}
	indent := ordinaryMutation[:len(ordinaryMutation)-len(strings.TrimLeft(ordinaryMutation, " \t"))]
	return indent + "refreshQuery();\n"
}

func (g *webGenerator) openapiQueryOperation(page PageDecl, entity EntityDecl, query QueryDecl) map[string]any {
	responses := openapiJSONResponses(map[string]any{
		"type": "array", "maxItems": effectiveQueryLimit(query),
		"items": map[string]any{"$ref": "#/components/schemas/" + entity.Name},
	})
	operation := map[string]any{
		"summary":                          "Run " + query.Name,
		"description":                      "Fixed server-side list query. Filters are combined with AND before stable sorting and the limit. The query does not restrict detail, mutations, or the base list. Table search and filters operate only on the returned window.",
		"x-blacklang-query":                query.Name,
		"x-blacklang-source":               query.Source,
		"x-blacklang-limit":                effectiveQueryLimit(query),
		"x-blacklang-required-read-fields": queryFieldNames(query),
		"parameters": []any{map[string]any{
			"name": "archived", "in": "query",
			"description": "Use all to include archived records while retaining the declared predicates, order, and limit.",
			"schema":      map[string]any{"type": "string", "enum": []string{"all"}},
		}},
		"responses": responses,
	}
	if g.program.Auth != nil {
		operation["security"] = []any{map[string]any{"cookieAuth": []any{}}}
		responses["401"] = map[string]any{"description": "Authentication required"}
	}
	if g.hasRuntimePermissions() || len(page.Access) > 0 {
		responses["403"] = map[string]any{"description": "Page access, entity read, or query field read permission denied"}
	}
	g.addOpenAPIRelationLoad(operation, entity, "query")
	return operation
}

func (g *webGenerator) openapiQuerySummaryOperation(page PageDecl, query QueryDecl) map[string]any {
	responses := openapiJSONResponses(querySummaryOpenAPISchema(query))
	operation := map[string]any{
		"summary":                          "Run " + query.Name + " summary",
		"description":                      "Fixed server-side aggregate summary for a bound query. The summary uses the same row policy, archive policy, and where filters as the query list. Sort and limit are not applied to aggregate results.",
		"x-blacklang-query":                query.Name,
		"x-blacklang-query-summary":        query.Name,
		"x-blacklang-source":               query.Source,
		"x-blacklang-aggregates":           queryAggregateOpenAPIMetadata(query),
		"x-blacklang-required-read-fields": queryFieldNames(query),
		"parameters": []any{map[string]any{
			"name": "archived", "in": "query",
			"description": "Use all to include archived records while retaining the declared predicates.",
			"schema":      map[string]any{"type": "string", "enum": []string{"all"}},
		}},
		"responses": responses,
	}
	if g.program.Auth != nil {
		operation["security"] = []any{map[string]any{"cookieAuth": []any{}}}
		responses["401"] = map[string]any{"description": "Authentication required"}
	}
	if g.hasRuntimePermissions() || len(page.Access) > 0 {
		responses["403"] = map[string]any{"description": "Page access, entity read, or aggregate field read permission denied"}
	}
	return operation
}

func querySummaryOpenAPISchema(query QueryDecl) map[string]any {
	properties := map[string]any{}
	required := []string{}
	for _, aggregate := range query.Aggregates {
		required = append(required, aggregate.Name)
		if aggregate.Function == "count" {
			properties[aggregate.Name] = map[string]any{"type": "integer", "minimum": 0}
		} else {
			properties[aggregate.Name] = map[string]any{"type": []string{"number", "null"}}
		}
	}
	return openapiObjectSchema(properties, required)
}

func queryAggregateOpenAPIMetadata(query QueryDecl) []any {
	metadata := []any{}
	for _, aggregate := range query.Aggregates {
		item := map[string]any{
			"name":     aggregate.Name,
			"function": aggregate.Function,
		}
		if aggregate.Field != "" {
			item["field"] = aggregate.Field
		}
		metadata = append(metadata, item)
	}
	return metadata
}

func (g *webGenerator) pageQuery(page PageDecl) (QueryDecl, bool) {
	if page.Query == "" {
		return QueryDecl{}, false
	}
	return findQuery(g.program, page.Query)
}

func (g *webGenerator) pageQueryHasAggregates(page PageDecl) bool {
	query, ok := g.pageQuery(page)
	return ok && len(query.Aggregates) > 0
}

func (g *webGenerator) querySummaryInitialState(page PageDecl) string {
	query, ok := g.pageQuery(page)
	if !ok || len(query.Aggregates) == 0 {
		return "{}"
	}
	parts := []string{}
	for _, aggregate := range query.Aggregates {
		parts = append(parts, fmt.Sprintf("%q: null", aggregate.Name))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func (g *webGenerator) querySummaryHelpers(page PageDecl) string {
	if !g.pageQueryHasAggregates(page) {
		return ""
	}
	return `function formatQuerySummaryValue(value: number | null | undefined) {
  if (value === null || value === undefined) return "—";
  return new Intl.NumberFormat().format(value);
}

`
}

func (g *webGenerator) querySummaryLoadStatement(page PageDecl, entityAPI string, activeRequest string) string {
	if !g.pageQueryHasAggregates(page) {
		return ""
	}
	return fmt.Sprintf(`    %sApi.querySummary(showArchived)
      .then((summary) => {
        if (%s) setQuerySummary(summary);
      })
      .catch((reason: unknown) => {
        if (%s) setError(reason instanceof Error ? reason.message : "Unable to load query summary");
      });

`, entityAPI, activeRequest, activeRequest)
}

func (g *webGenerator) querySummaryMarkup(page PageDecl) string {
	query, ok := g.pageQuery(page)
	if !ok || len(query.Aggregates) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("        <div className=\"query-summary\" aria-label=\"Query summary\">\n")
	for _, aggregate := range query.Aggregates {
		builder.WriteString(fmt.Sprintf("          <div className=\"query-summary-card\"><span>%s</span><strong>{formatQuerySummaryValue(querySummary[%q])}</strong></div>\n", identifierLabel(aggregate.Name), aggregate.Name))
	}
	builder.WriteString("        </div>\n")
	return builder.String()
}

func querySummarySmokeTests(g *webGenerator) string {
	if g.program.Auth == nil {
		return ""
	}
	for _, page := range g.program.Pages {
		if !g.pageQueryHasAggregates(page) {
			continue
		}
		path := "/api/" + strings.ToLower(page.Name) + "/query/summary"
		name := lowerCamelCase(page.Query) + "SummaryResponse"
		return fmt.Sprintf("  const %s = await fetch(baseURL + %q);\n  assert.equal(%s.status, 401, %q);\n\n", name, path, name, page.Query+" summary should reject anonymous requests")
	}
	return ""
}
