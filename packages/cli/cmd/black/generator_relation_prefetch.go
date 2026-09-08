package main

import (
	"fmt"
	"strings"
)

func (g *webGenerator) relationFieldsForAnyLoad(entity EntityDecl) []FieldDecl {
	fields := []FieldDecl{}
	for _, field := range g.relationFields(entity) {
		for _, context := range relationLoadScopeOrder {
			if relationLoadIncludesContext(field, context) {
				fields = append(fields, field)
				break
			}
		}
	}
	return fields
}

func (g *webGenerator) hasRelationFieldsForAnyLoad(entity EntityDecl) bool {
	return len(g.relationFieldsForAnyLoad(entity)) > 0
}

func (g *webGenerator) hasLoadedRelationTargetPolicies(entity EntityDecl) bool {
	for _, field := range g.relationFieldsForAnyLoad(entity) {
		target, ok := g.findEntity(field.Type)
		if ok && g.hasEntityPolicies(target) {
			return true
		}
	}
	return false
}

func (g *webGenerator) relationLoadAttachStatement(entity EntityDecl, context string, itemsExpression string, indent string) string {
	if !g.hasRelationFieldsForLoad(entity, context) {
		return ""
	}
	return fmt.Sprintf("%sawait attach%sRelations(req, %s, %q);\n", indent, entity.Name, itemsExpression, context)
}

func (g *webGenerator) routeRelationLoadHelpers(entity EntityDecl) string {
	fields := g.relationFieldsForAnyLoad(entity)
	if len(fields) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(g.relationTargetSanitizerHelpers(entity))
	builder.WriteString(g.relationTargetRowPolicyHelpers(entity))
	builder.WriteString(fmt.Sprintf("async function attach%sRelations(req: express.Request, items: any[], context: \"list\" | \"detail\" | \"query\" | \"mutation\") {\n", entity.Name))
	builder.WriteString("  if (items.length === 0) return items;\n")
	if g.hasRuntimePermissions() {
		builder.WriteString("  const roles = currentRoles(req);\n")
	}
	for _, field := range fields {
		target, ok := g.findEntity(field.Type)
		if !ok {
			continue
		}
		contexts := relationLoadContextsForField(field)
		if len(contexts) == 0 {
			continue
		}
		condition := fmt.Sprintf("%s.includes(context)", tsStringArrayLiteral(contexts))
		if g.hasRuntimePermissions() {
			condition += fmt.Sprintf(" && canAccessField(roles, \"read\", %q, %q)", entity.Name, field.Name)
		}
		idsName := lowerCamelCase(field.Name) + "Ids"
		recordsName := lowerCamelCase(field.Name) + "Records"
		mapName := lowerCamelCase(field.Name) + "Map"
		builder.WriteString(fmt.Sprintf("  if (%s) {\n", condition))
		builder.WriteString(fmt.Sprintf("    const %s = Array.from(new Set(items.map((item) => item.%s).filter((id): id is string => typeof id === \"string\" && id.length > 0)));\n", idsName, relationIDFieldName(field)))
		builder.WriteString(fmt.Sprintf("    if (%s.length > 0) {\n", idsName))
		builder.WriteString(fmt.Sprintf("      const %s = await prisma.%s.findMany({\n", recordsName, lowerCamelCase(target.Name)))
		builder.WriteString(fmt.Sprintf("        where: %s\n", g.relationTargetRowPolicyWhere(target, fmt.Sprintf("{ id: { in: %s } }", idsName))))
		builder.WriteString("      });\n")
		if g.hasRuntimePermissions() {
			builder.WriteString(fmt.Sprintf("      const %s = new Map(%s.map((record) => [record.id, sanitize%sRelation(record, roles)]));\n", mapName, recordsName, target.Name))
		} else {
			builder.WriteString(fmt.Sprintf("      const %s = new Map(%s.map((record) => [record.id, record]));\n", mapName, recordsName))
		}
		builder.WriteString("      for (const item of items) {\n")
		builder.WriteString(fmt.Sprintf("        item.%s = %s.get(item.%s) ?? null;\n", field.Name, mapName, relationIDFieldName(field)))
		builder.WriteString("      }\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n")
	}
	builder.WriteString("  return items;\n")
	builder.WriteString("}\n\n")
	return builder.String()
}

func relationLoadContextsForField(field FieldDecl) []string {
	contexts := []string{}
	for _, context := range relationLoadScopeOrder {
		if relationLoadIncludesContext(field, context) {
			contexts = append(contexts, context)
		}
	}
	return contexts
}

func (g *webGenerator) relationTargetSanitizerHelpers(entity EntityDecl) string {
	if !g.hasRuntimePermissions() {
		return ""
	}
	targets := g.loadedRelationTargetEntities(entity)
	if len(targets) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, target := range targets {
		builder.WriteString(fmt.Sprintf("function sanitize%sRelation(item: any, roles: string[]) {\n", target.Name))
		builder.WriteString("  if (!item) return item;\n")
		builder.WriteString("  const output = { ...item };\n")
		for _, field := range target.Fields {
			builder.WriteString(fmt.Sprintf("  if (!canAccessField(roles, \"read\", %q, %q)) delete output.%s;\n", target.Name, field.Name, g.sqliteColumnName(field)))
			if g.isRelationField(field) {
				builder.WriteString(fmt.Sprintf("  if (!canAccessField(roles, \"read\", %q, %q)) delete output.%s;\n", target.Name, field.Name, field.Name))
			}
		}
		builder.WriteString("  return output;\n")
		builder.WriteString("}\n\n")
	}
	return builder.String()
}

func (g *webGenerator) relationTargetRowPolicyHelpers(entity EntityDecl) string {
	targets := g.loadedRelationTargetEntities(entity)
	if len(targets) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, target := range targets {
		if !g.hasEntityPolicies(target) {
			continue
		}
		builder.WriteString(g.rowPolicyWhereHelperForEntity(target, relationTargetRowPolicyWhereFunction(target)))
	}
	return builder.String()
}

func (g *webGenerator) loadedRelationTargetEntities(entity EntityDecl) []EntityDecl {
	targets := []EntityDecl{}
	seen := map[string]bool{}
	for _, field := range g.relationFieldsForAnyLoad(entity) {
		if seen[field.Type] {
			continue
		}
		target, ok := g.findEntity(field.Type)
		if !ok {
			continue
		}
		seen[target.Name] = true
		targets = append(targets, target)
	}
	return targets
}

func (g *webGenerator) relationTargetRowPolicyWhere(entity EntityDecl, base string) string {
	if !g.hasEntityPolicies(entity) {
		return base
	}
	return relationTargetRowPolicyWhereFunction(entity) + "(req, " + base + ") as any"
}

func relationTargetRowPolicyWhereFunction(entity EntityDecl) string {
	return "rowPolicyWhereFor" + entity.Name + "Relation"
}
