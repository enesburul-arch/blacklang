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
			builder.WriteString(fmt.Sprintf("  if (!%s.every((field) => canAccessField(currentRole(req), \"read\", %q, field))) {\n", tsStringArrayLiteral(fields), entity.Name))
			builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n")
		}
	}
	builder.WriteString("  const includeArchived = req.query.archived === \"all\";\n")
	builder.WriteString(fmt.Sprintf("  const items = await %sModel.findMany({\n", identifier))
	builder.WriteString("    where: { AND: [\n")
	builder.WriteString("      includeArchived ? {} : { archivedAt: null },\n")
	for _, filter := range query.Where {
		field, _ := findField(entity, filter.Field)
		builder.WriteString(fmt.Sprintf("      { [%q]: { %s: %s } },\n", filter.Field, queryPrismaOperator(filter.Operator), queryPrismaLiteral(filter.Value, field)))
	}
	builder.WriteString("    ] },\n")
	if g.hasRelationFields(entity) {
		builder.WriteString(g.prismaIncludeLine(entity, "    "))
	}
	builder.WriteString("    orderBy: [")
	if query.Sort.Field != "" {
		builder.WriteString(fmt.Sprintf("{ [%q]: %q }, ", query.Sort.Field, query.Sort.Direction))
	}
	builder.WriteString("{ id: \"asc\" }],\n")
	builder.WriteString(fmt.Sprintf("    take: %d\n", effectiveQueryLimit(query)))
	builder.WriteString("  });\n")
	if g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  res.json(items.map((item) => sanitize%s(item, currentRole(req))));\n", entity.Name))
	} else {
		builder.WriteString("  res.json(items);\n")
	}
	builder.WriteString("});\n\n")
	return builder.String()
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
	return fmt.Sprintf("  queryList: (includeArchived = false) =>\n    request<%s[]>(endpoint + \"/query\" + (includeArchived ? \"?archived=all\" : \"\")),", entity.Name)
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
	return operation
}
