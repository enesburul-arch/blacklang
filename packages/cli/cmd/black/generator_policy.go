package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func entityPolicyFieldMap(entity EntityDecl) map[string]EntityPolicyDecl {
	fields := map[string]EntityPolicyDecl{}
	for _, policy := range entity.Policies {
		if policy.Field == "" {
			continue
		}
		fields[policy.Field] = policy
	}
	return fields
}

func entityPolicyField(entity EntityDecl, kind string) (string, bool) {
	for _, policy := range entity.Policies {
		if policy.Kind == kind && policy.Field != "" {
			return policy.Field, true
		}
	}
	return "", false
}

func entityPolicyFieldNames(entity EntityDecl) []string {
	names := []string{}
	seen := map[string]bool{}
	for _, policy := range entity.Policies {
		if policy.Field == "" || seen[policy.Field] {
			continue
		}
		seen[policy.Field] = true
		names = append(names, policy.Field)
	}
	return names
}

func (g *webGenerator) hasEntityPolicies(entity EntityDecl) bool {
	return len(entity.Policies) > 0
}

func (g *webGenerator) hasAnyEntityPolicies() bool {
	for _, entity := range g.program.Entities {
		if g.hasEntityPolicies(entity) {
			return true
		}
	}
	return false
}

func (g *webGenerator) hasTenantPolicies() bool {
	for _, entity := range g.program.Entities {
		if _, ok := entityPolicyField(entity, "tenant"); ok {
			return true
		}
	}
	return false
}

func (g *webGenerator) routeNeedsCurrentUser(entity EntityDecl) bool {
	return g.hasRuntimePermissions() || g.hasEntityPolicies(entity) || g.hasLoadedRelationTargetPolicies(entity)
}

func (g *webGenerator) currentUserTypeLiteral() string {
	if g.hasTenantPolicies() {
		return "{ id: string; name: string; email: string; role: string; roles: string[]; tenantId: string }"
	}
	return "{ id: string; name: string; email: string; role: string; roles: string[]; tenantId?: string }"
}

func (g *webGenerator) rowPolicyHelpers(entity EntityDecl) string {
	if !g.hasEntityPolicies(entity) {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(g.rowPolicyWhereHelperForEntity(entity, "rowPolicyWhere"))
	builder.WriteString(g.rowPolicyInputHelperForEntity(entity))
	builder.WriteString(g.stripRowPolicyFieldsHelperForEntity(entity))
	return builder.String()
}

func (g *webGenerator) rowPolicyWhereHelperForEntity(entity EntityDecl, functionName string) string {
	conditions := []string{}
	if field, ok := entityPolicyField(entity, "owner"); ok {
		conditions = append(conditions, fmt.Sprintf("    { %s: user.id },", field))
	}
	if field, ok := entityPolicyField(entity, "tenant"); ok {
		conditions = append(conditions, fmt.Sprintf("    { %s: user.tenantId ?? \"default\" },", field))
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("function %s(req: express.Request, base: Record<string, unknown> = {}) {\n", functionName))
	builder.WriteString("  const user = currentUser(req);\n")
	builder.WriteString("  if (!user) return { AND: [{ id: \"__blacklang_no_authenticated_user__\" }] };\n")
	builder.WriteString("  return { AND: [\n")
	builder.WriteString("    base,\n")
	for _, condition := range conditions {
		builder.WriteString(condition)
		builder.WriteString("\n")
	}
	builder.WriteString("  ] };\n")
	builder.WriteString("}\n\n")
	return builder.String()
}

func (g *webGenerator) rowPolicyInputHelperForEntity(entity EntityDecl) string {
	inputs := []string{}
	if field, ok := entityPolicyField(entity, "owner"); ok {
		inputs = append(inputs, fmt.Sprintf("    %s: user.id,", field))
	}
	if field, ok := entityPolicyField(entity, "tenant"); ok {
		inputs = append(inputs, fmt.Sprintf("    %s: user.tenantId ?? \"default\",", field))
	}

	var builder strings.Builder
	builder.WriteString("function applyRowPolicyInput(req: express.Request, input: Record<string, unknown>) {\n")
	builder.WriteString("  const user = currentUser(req);\n")
	builder.WriteString("  if (!user) return input;\n")
	builder.WriteString("  return {\n")
	builder.WriteString("    ...input,\n")
	for _, input := range inputs {
		builder.WriteString(input)
		builder.WriteString("\n")
	}
	builder.WriteString("  };\n")
	builder.WriteString("}\n\n")
	return builder.String()
}

func (g *webGenerator) stripRowPolicyFieldsHelperForEntity(entity EntityDecl) string {
	deletes := []string{}
	if field, ok := entityPolicyField(entity, "owner"); ok {
		deletes = append(deletes, fmt.Sprintf("  delete output.%s;", field))
	}
	if field, ok := entityPolicyField(entity, "tenant"); ok {
		deletes = append(deletes, fmt.Sprintf("  delete output.%s;", field))
	}

	var builder strings.Builder
	builder.WriteString("function stripRowPolicyFields(input: Record<string, unknown>) {\n")
	builder.WriteString("  const output = { ...input };\n")
	for _, line := range deletes {
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	builder.WriteString("  return output;\n")
	builder.WriteString("}\n\n")
	return builder.String()
}

func (g *webGenerator) rowPolicyWhere(entity EntityDecl, base string) string {
	if !g.hasEntityPolicies(entity) {
		return base
	}
	return "rowPolicyWhere(req, " + base + ") as any"
}

func (g *webGenerator) rowPolicyInputExpression(entity EntityDecl, input string) string {
	if !g.hasEntityPolicies(entity) {
		return input
	}
	return "applyRowPolicyInput(req, " + input + ")"
}

func (g *webGenerator) mutableInputExpression(entity EntityDecl, input string) string {
	if !g.hasEntityPolicies(entity) {
		return input
	}
	return "stripRowPolicyFields(" + input + ")"
}

func (g *webGenerator) apiClientInputOmit(entity EntityDecl) string {
	names := append([]string{"id"}, entityPolicyFieldNames(entity)...)
	parts := []string{}
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%q", name))
	}
	return strings.Join(parts, " | ")
}

func (g *webGenerator) authUserTenantFieldLine() string {
	if !g.hasTenantPolicies() {
		return ""
	}
	return "  tenantId: string;\n"
}

func (g *webGenerator) authPublicUserTenantLine() string {
	if !g.hasTenantPolicies() {
		return ""
	}
	return ",\n    tenantId: user.tenantId"
}

func (g *webGenerator) defaultAuthRolesJSON() string {
	content, err := json.Marshal([]string{g.defaultAuthRole()})
	if err != nil {
		return "[]"
	}
	return string(content)
}

func (g *webGenerator) sqliteAuthRolesColumnLine() string {
	return fmt.Sprintf("  \"roles\" TEXT NOT NULL DEFAULT '%s',\n", strings.ReplaceAll(g.defaultAuthRolesJSON(), "'", "''"))
}

func (g *webGenerator) sqliteAuthRolesMigration() string {
	defaultRoles := strings.ReplaceAll(g.defaultAuthRolesJSON(), "'", "''")
	return fmt.Sprintf("if (!blackUserColumns.some((column) => column.name === \"roles\")) {\n  db.prepare(`ALTER TABLE \"BlackUser\" ADD COLUMN \"roles\" TEXT NOT NULL DEFAULT '%s'`).run();\n  db.prepare(`UPDATE \"BlackUser\" SET \"roles\" = '[' || char(34) || \"role\" || char(34) || ']'`).run();\n}\n\n", defaultRoles)
}

func (g *webGenerator) sqliteAuthTenantColumnLine() string {
	if !g.hasTenantPolicies() {
		return ""
	}
	return "  \"tenantId\" TEXT NOT NULL DEFAULT 'default',\n"
}

func (g *webGenerator) sqliteAuthTenantMigration() string {
	if !g.hasTenantPolicies() {
		return ""
	}
	return "if (!blackUserColumns.some((column) => column.name === \"tenantId\")) {\n  db.prepare(`ALTER TABLE \"BlackUser\" ADD COLUMN \"tenantId\" TEXT NOT NULL DEFAULT 'default'`).run();\n}\n\n"
}

func (g *webGenerator) sqliteAuthUserSelectColumns(alias string) string {
	fields := []string{"id", "name", "email", "role", "roles", "passwordHash"}
	if g.hasTenantPolicies() {
		fields = append(fields, "tenantId")
	}
	if alias == "" {
		return strings.Join(fields, ", ")
	}
	prefixed := []string{}
	for _, field := range fields {
		prefixed = append(prefixed, alias+"."+field)
	}
	return strings.Join(prefixed, ", ")
}

func (g *webGenerator) sqliteAuthUserInsertColumns() string {
	if !g.hasTenantPolicies() {
		return "id, name, email, role, roles, passwordHash"
	}
	return "id, name, email, role, roles, passwordHash, tenantId"
}

func (g *webGenerator) sqliteAuthUserInsertPlaceholders() string {
	if !g.hasTenantPolicies() {
		return "?, ?, ?, ?, ?, ?"
	}
	return "?, ?, ?, ?, ?, ?, ?"
}

func (g *webGenerator) sqliteAuthUserInsertArgs() string {
	if !g.hasTenantPolicies() {
		return "user.id, user.name, user.email, user.role, user.roles, user.passwordHash"
	}
	return "user.id, user.name, user.email, user.role, user.roles, user.passwordHash, user.tenantId"
}

func (g *webGenerator) sqliteAuthUserObjectTenantLine() string {
	if !g.hasTenantPolicies() {
		return ""
	}
	return ",\n    tenantId: \"default\""
}

func (g *webGenerator) postgresAuthUserCreateTenantLine() string {
	if !g.hasTenantPolicies() {
		return ""
	}
	return ",\n        tenantId: \"default\""
}

func (g *webGenerator) authPrismaTenantLine() string {
	if !g.hasTenantPolicies() {
		return ""
	}
	return "  tenantId String @default(\"default\")\n"
}
