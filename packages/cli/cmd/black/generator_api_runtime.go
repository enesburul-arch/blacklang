package main

import (
	"fmt"
	"strings"
)

func (g *webGenerator) hasExplicitAPIs() bool {
	return len(g.program.APIs) > 0
}

func (g *webGenerator) hasExplicitAPIHandlers() bool {
	for _, api := range g.program.APIs {
		if api.Update != nil {
			return true
		}
	}
	return false
}

func (g *webGenerator) hasPrivateExplicitAPIHandlers() bool {
	for _, api := range g.program.APIs {
		if api.Update != nil && explicitAPIAccess(api) == "private" {
			return true
		}
	}
	return false
}

func (g *webGenerator) hasTransactionalPrivateExplicitAPIHandlers() bool {
	if !g.hasRuntimePermissions() {
		return false
	}
	for _, api := range g.program.APIs {
		if api.Update == nil || explicitAPIAccess(api) != "private" {
			continue
		}
		if _, transactional := g.explicitAPITransactionName(api); transactional {
			return true
		}
	}
	return false
}

func (g *webGenerator) explicitAPIHelpers() string {
	if !g.hasExplicitAPIs() {
		return ""
	}
	return `type ExplicitAPIParamSpec = { name: string; type: string; required: boolean; min?: number; max?: number; minLength?: number; maxLength?: number; pattern?: string; url?: boolean; message?: string };
type ExplicitAPIValue = string | number | boolean | null;
type ExplicitAPIValuesResult = { ok: true; value: Record<string, ExplicitAPIValue> } | { ok: false; error: string };

function firstExplicitAPIValue(raw: unknown): unknown {
  if (Array.isArray(raw)) {
    return raw[0] ?? null;
  }
  return raw ?? null;
}

function coerceExplicitAPIValue(raw: unknown, spec: ExplicitAPIParamSpec): { ok: true; value: ExplicitAPIValue } | { ok: false; error: string } {
  const value = firstExplicitAPIValue(raw);
  if (value === null || value === undefined || value === "") {
    if (spec.required) {
      return { ok: false, error: spec.message ?? spec.name + " is required" };
    }
    return { ok: true, value: null };
  }

  const text = String(value);
  let coerced: ExplicitAPIValue = text;
  if (spec.type === "number" || spec.type === "integer") {
    const numberValue = Number(text);
    if (!Number.isInteger(numberValue)) {
      return { ok: false, error: spec.message ?? spec.name + " must be an integer" };
    }
    coerced = numberValue;
  } else if (spec.type === "decimal" || spec.type === "money") {
    const numberValue = Number(text);
    if (!Number.isFinite(numberValue)) {
      return { ok: false, error: spec.message ?? spec.name + " must be a finite number" };
    }
    coerced = numberValue;
  } else if (spec.type === "boolean") {
    if (text === "true") coerced = true;
    else if (text === "false") coerced = false;
    else return { ok: false, error: spec.message ?? spec.name + " must be true or false" };
  } else if (spec.type === "date" && !/^\d{4}-\d{2}-\d{2}$/.test(text)) {
    return { ok: false, error: spec.message ?? spec.name + " must be YYYY-MM-DD" };
  } else if (spec.type === "datetime" && Number.isNaN(Date.parse(text))) {
    return { ok: false, error: spec.message ?? spec.name + " must be a valid datetime" };
  } else if (spec.type === "email" && !text.includes("@")) {
    return { ok: false, error: spec.message ?? spec.name + " must be an email" };
  } else if (spec.type === "image" && !(text.startsWith("data:image/") || text.startsWith("https://") || text.startsWith("http://"))) {
    return { ok: false, error: spec.message ?? spec.name + " must be a valid image" };
  } else if (spec.type === "file" && !(text.startsWith("data:") || text.startsWith("https://") || text.startsWith("http://"))) {
    return { ok: false, error: spec.message ?? spec.name + " must be a valid file" };
  }
  if (typeof coerced === "number") {
    if (spec.min !== undefined && coerced < spec.min) return { ok: false, error: spec.message ?? spec.name + " must be at least " + spec.min };
    if (spec.max !== undefined && coerced > spec.max) return { ok: false, error: spec.message ?? spec.name + " must be at most " + spec.max };
  }
  if (typeof coerced === "string") {
    if (spec.minLength !== undefined && coerced.length < spec.minLength) return { ok: false, error: spec.message ?? spec.name + " is too short" };
    if (spec.maxLength !== undefined && coerced.length > spec.maxLength) return { ok: false, error: spec.message ?? spec.name + " is too long" };
    if (spec.pattern && !(new RegExp(spec.pattern)).test(coerced)) return { ok: false, error: spec.message ?? spec.name + " has invalid format" };
    if (spec.url) {
      try { new URL(coerced); } catch { return { ok: false, error: spec.message ?? spec.name + " must be a valid URL" }; }
    }
  }
  return { ok: true, value: coerced };
}

function readExplicitAPIValues(source: Record<string, unknown>, specs: ExplicitAPIParamSpec[]): ExplicitAPIValuesResult {
  const value: Record<string, ExplicitAPIValue> = {};
  for (const spec of specs) {
    const result = coerceExplicitAPIValue(source[spec.name], spec);
    if (!result.ok) {
      return result;
    }
    value[spec.name] = result.value;
  }
  return { ok: true, value };
}

`
}

func (g *webGenerator) explicitAPIRoutes(public bool) string {
	if !g.hasExplicitAPIs() {
		return ""
	}
	var builder strings.Builder
	for _, api := range g.program.APIs {
		routeIsPublic := explicitAPIAccess(api) == "public"
		if routeIsPublic != public {
			continue
		}
		method := strings.ToLower(api.Method)
		if !supportedAPIMethods[strings.ToUpper(api.Method)] || api.Path == "" {
			continue
		}
		builder.WriteString(g.explicitAPIRoute(api, method))
	}
	if builder.Len() > 0 {
		builder.WriteString("\n")
	}
	return builder.String()
}

func (g *webGenerator) explicitAPIRoute(api APIDecl, method string) string {
	var builder strings.Builder
	routeAsync := ""
	if api.Update != nil {
		routeAsync = "async "
	}
	bodyExpression := "null"
	if explicitAPIHasBody(api.Method) {
		bodyExpression = "req.body ?? null"
	}
	statusCode := explicitAPIStatusCode(api)
	statusText := explicitAPIStatusText(api)

	builder.WriteString(fmt.Sprintf("app.%s(%s, %s(req, res) => {\n", method, contractJSONString(expressAPIPath(api.Path)), routeAsync))
	builder.WriteString(fmt.Sprintf("  const params = readExplicitAPIValues(req.params as Record<string, unknown>, %s);\n", explicitAPIParamSpecsLiteral(api.Params, true)))
	builder.WriteString("  if (!params.ok) {\n")
	builder.WriteString("    res.status(400).json({ error: params.error });\n")
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n")
	builder.WriteString(fmt.Sprintf("  const query = readExplicitAPIValues(req.query as Record<string, unknown>, %s);\n", explicitAPIParamSpecsLiteral(api.Queries, false)))
	builder.WriteString("  if (!query.ok) {\n")
	builder.WriteString("    res.status(400).json({ error: query.error });\n")
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n")
	builder.WriteString(fmt.Sprintf("  const body = %s;\n", bodyExpression))
	if len(api.Body) > 0 {
		builder.WriteString("  const rawBody = body && typeof body === \"object\" && !Array.isArray(body) ? body as Record<string, unknown> : {};\n")
		builder.WriteString(fmt.Sprintf("  const bodyValues = readExplicitAPIValues(rawBody, %s);\n", g.explicitAPIBodySpecsLiteral(api.Body)))
		builder.WriteString("  if (!bodyValues.ok) {\n")
		builder.WriteString("    res.status(400).json({ error: bodyValues.error });\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n")
	}
	if api.Update != nil {
		if _, transactional := g.explicitAPITransactionName(api); transactional {
			builder.WriteString(g.transactionalExplicitAPIUpdateRouteBody(api, *api.Update, statusCode, statusText))
		} else {
			builder.WriteString(g.explicitAPIUpdateRouteBody(api, *api.Update, statusCode, statusText))
		}
	} else {
		builder.WriteString(fmt.Sprintf("  res.status(%d).json({ api: %s, status: %s, runtime: \"declared\", method: req.method, path: %s, access: %s, webhook: %t, params: params.value, query: query.value, body });\n", statusCode, contractJSONString(api.Name), contractJSONString(statusText), contractJSONString(api.Path), contractJSONString(explicitAPIAccess(api)), api.Webhook))
	}
	builder.WriteString("});\n")
	return builder.String()
}

func (g *webGenerator) transactionalExplicitAPIUpdateRouteBody(api APIDecl, update APIUpdateDecl, statusCode int, statusText string) string {
	entity, ok := g.findEntity(update.Source)
	if !ok {
		return fmt.Sprintf("  res.status(%d).json({ api: %s, status: %s, runtime: \"declared\", method: req.method, path: %s, access: %s, webhook: %t, params: params.value, query: query.value, body });\n", statusCode, contractJSONString(api.Name), contractJSONString(statusText), contractJSONString(api.Path), contractJSONString(explicitAPIAccess(api)), api.Webhook)
	}
	identifier := lowerCamelCase(entity.Name)
	txModel := identifier + "TxModel"
	whereExpression := g.explicitAPIUpdateWhereExpression(api, update, entity)
	var builder strings.Builder
	if explicitAPIAccess(api) == "private" && g.hasRuntimePermissions() {
		builder.WriteString("  const actorRole = String((req as any).blackUser?.role ?? \"\");\n")
		for _, field := range apiUpdateSetFieldNames(update) {
			builder.WriteString(fmt.Sprintf("  if (!canAccessField(actorRole, \"update\", %q, %q)) {\n", entity.Name, field))
			builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("  const apiResult = await prisma.$transaction(async (tx) => {\n")
	builder.WriteString(fmt.Sprintf("    const %s = tx.%s;\n", txModel, identifier))
	builder.WriteString(fmt.Sprintf("    const existing = await %s.findFirst({ where: %s as any });\n", txModel, whereExpression))
	builder.WriteString("    if (!existing) {\n")
	builder.WriteString("      return { status: \"notFound\" as const };\n")
	builder.WriteString("    }\n\n")
	builder.WriteString("    const data: Record<string, unknown> = {};\n")
	guardCounter := 0
	builder.WriteString(g.apiStatementsRuntime(api, entity, apiUpdateStatements(update), "    ", "existing", "transaction", &guardCounter, map[string]string{}))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("    const item = await %s.update({\n", txModel))
	builder.WriteString("      where: { id: existing.id },\n")
	builder.WriteString("      data: data as any")
	if g.hasRelationFields(entity) {
		builder.WriteString(",\n")
		builder.WriteString(g.prismaIncludeLine(entity, "      ", false))
	} else {
		builder.WriteString("\n")
	}
	builder.WriteString("    });\n")
	if explicitAPIAccess(api) == "private" && g.hasRuntimePermissions() {
		builder.WriteString("    const actor = (req as any).blackUser ?? null;\n")
		builder.WriteString("    if (actor) {\n")
		if g.isMySQL() {
			builder.WriteString("      await tx.blackAuditLog.create({\n")
			builder.WriteString("        data: {\n")
			builder.WriteString("          id: crypto.randomUUID(),\n")
			builder.WriteString("          actorUserId: actor.id,\n")
			builder.WriteString("          actorRole: actor.role,\n")
			builder.WriteString(fmt.Sprintf("          action: %q,\n", "api."+api.Name))
			builder.WriteString(fmt.Sprintf("          resource: %q,\n", entity.Name))
			builder.WriteString("          resourceId: item.id,\n")
			builder.WriteString(fmt.Sprintf("          summary: %q\n", api.Name+" updated "+entity.Name))
			builder.WriteString("        }\n")
			builder.WriteString("      });\n")
		} else {
			builder.WriteString("      await tx.$executeRaw`\n")
			builder.WriteString("        INSERT INTO \"BlackAuditLog\" (\"id\", \"actorUserId\", \"actorRole\", \"action\", \"resource\", \"resourceId\", \"summary\")\n")
			builder.WriteString(fmt.Sprintf("        VALUES (${crypto.randomUUID()}, ${actor.id}, ${actor.role}, ${%q}, ${%q}, ${item.id}, ${%q})\n", "api."+api.Name, entity.Name, api.Name+" updated "+entity.Name))
			builder.WriteString("      `;\n")
		}
		builder.WriteString("    }\n")
	}
	builder.WriteString("    return { status: \"ok\" as const, item };\n")
	builder.WriteString("  });\n\n")
	builder.WriteString("  if (apiResult.status === \"notFound\") {\n")
	builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n")
	if explicitAPIUpdateHasDivisionGuard(update) {
		builder.WriteString("  if (apiResult.status === \"divisionByZero\") {\n")
		builder.WriteString("    res.status(400).json({ error: \"Division by zero\" });\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n")
	}
	builder.WriteString("\n")
	builder.WriteString("  const item = apiResult.item;\n")
	builder.WriteString(fmt.Sprintf("  res.status(%d).json({ api: %s, status: %s, runtime: \"declared\", handler: \"update\", entity: %s, id: item.id, updated: true });\n", statusCode, contractJSONString(api.Name), contractJSONString(statusText), contractJSONString(entity.Name)))
	builder.WriteString("  return;\n")
	return builder.String()
}

func (g *webGenerator) explicitAPIUpdateRouteBody(api APIDecl, update APIUpdateDecl, statusCode int, statusText string) string {
	entity, ok := g.findEntity(update.Source)
	if !ok {
		return fmt.Sprintf("  res.status(%d).json({ api: %s, status: %s, runtime: \"declared\", method: req.method, path: %s, access: %s, webhook: %t, params: params.value, query: query.value, body });\n", statusCode, contractJSONString(api.Name), contractJSONString(statusText), contractJSONString(api.Path), contractJSONString(explicitAPIAccess(api)), api.Webhook)
	}
	identifier := lowerCamelCase(entity.Name)
	whereExpression := g.explicitAPIUpdateWhereExpression(api, update, entity)
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("  const existing = await prisma.%s.findFirst({ where: %s as any });\n", identifier, whereExpression))
	builder.WriteString("  if (!existing) {\n")
	builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n\n")
	if explicitAPIAccess(api) == "private" && g.hasRuntimePermissions() {
		builder.WriteString("  const actorRole = String((req as any).blackUser?.role ?? \"\");\n")
		for _, field := range apiUpdateSetFieldNames(update) {
			builder.WriteString(fmt.Sprintf("  if (!canAccessField(actorRole, \"update\", %q, %q)) {\n", entity.Name, field))
			builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
			builder.WriteString("    return;\n")
			builder.WriteString("  }\n")
		}
		builder.WriteString("\n")
	}
	builder.WriteString("  const data: Record<string, unknown> = {};\n")
	guardCounter := 0
	builder.WriteString(g.apiStatementsRuntime(api, entity, apiUpdateStatements(update), "  ", "existing", "http", &guardCounter, map[string]string{}))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("  const item = await prisma.%s.update({\n", identifier))
	builder.WriteString("    where: { id: existing.id },\n")
	builder.WriteString("    data: data as any")
	if g.hasRelationFields(entity) {
		builder.WriteString(",\n")
		builder.WriteString(g.prismaIncludeLine(entity, "    ", false))
	} else {
		builder.WriteString("\n")
	}
	builder.WriteString("  });\n")
	if explicitAPIAccess(api) == "private" && g.hasRuntimePermissions() {
		builder.WriteString(fmt.Sprintf("  writeAuditLog((req as any).blackUser, %q, %q, item.id, %q);\n", "api."+api.Name, entity.Name, api.Name+" updated "+entity.Name))
	}
	builder.WriteString(fmt.Sprintf("  res.status(%d).json({ api: %s, status: %s, runtime: \"declared\", handler: \"update\", entity: %s, id: item.id, updated: true });\n", statusCode, contractJSONString(api.Name), contractJSONString(statusText), contractJSONString(entity.Name)))
	builder.WriteString("  return;\n")
	return builder.String()
}

func (g *webGenerator) explicitAPIUpdateWhereExpression(api APIDecl, update APIUpdateDecl, entity EntityDecl) string {
	conditions := []string{fmt.Sprintf("{ %s: %s }", update.Where.Field, g.apiOperandExpression(update.Where.Value, "existing"))}
	for _, policy := range entity.Policies {
		if policy.Field == "" {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("{ %s: %s }", policy.Field, g.explicitAPIPolicyValueExpression(api, policy)))
	}
	if len(conditions) == 1 {
		return conditions[0]
	}
	return "{ AND: [" + strings.Join(conditions, ", ") + "] }"
}

func explicitAPIUpdateHasDivisionGuard(update APIUpdateDecl) bool {
	return apiStatementsHaveDivisionGuard(apiUpdateStatements(update))
}

func apiStatementsHaveDivisionGuard(statements []APIStatementDecl) bool {
	for _, statement := range statements {
		switch statement.Kind {
		case "value":
			if statement.Value != nil && len(apiExpressionDivisionDivisors(statement.Value.Expression)) > 0 {
				return true
			}
		case "set":
			if statement.Set != nil && len(apiExpressionDivisionDivisors(statement.Set.Expr)) > 0 {
				return true
			}
		case "if":
			if statement.If == nil {
				continue
			}
			if len(apiConditionDivisionDivisors(statement.If.Condition)) > 0 {
				return true
			}
			if apiStatementsHaveDivisionGuard(statement.If.Then) || apiStatementsHaveDivisionGuard(statement.If.Else) {
				return true
			}
		}
	}
	return false
}

func (g *webGenerator) explicitAPIPolicyValueExpression(api APIDecl, policy EntityPolicyDecl) string {
	if explicitAPIAccess(api) == "private" && g.program.Auth != nil {
		if policy.Kind == "owner" {
			return "((req as any).blackUser?.id ?? \"__blacklang_no_authenticated_user__\")"
		}
		if policy.Kind == "tenant" {
			return "((req as any).blackUser?.tenantId ?? \"default\")"
		}
	}
	for _, body := range api.Body {
		if body.Name == policy.Field {
			return fmt.Sprintf("bodyValues.value[%q]", policy.Field)
		}
	}
	for _, param := range api.Params {
		if param.Name == policy.Field {
			return fmt.Sprintf("params.value[%q]", policy.Field)
		}
	}
	return "\"__blacklang_missing_policy_scope__\""
}

func (g *webGenerator) apiExpression(expression APIExpressionDecl, existing string) string {
	return g.apiExpressionWithValues(expression, existing, nil)
}

func (g *webGenerator) apiExpressionWithValues(expression APIExpressionDecl, existing string, values map[string]string) string {
	if expression.Tree != nil {
		return g.apiCoreExpressionWithValues(*expression.Tree, existing, expression.Tree.Kind == "binary", values)
	}
	left := g.apiOperandExpressionWithValues(expression.Left, existing, values)
	if expression.Operator == "" || expression.Right == nil {
		return left
	}
	right := g.apiOperandExpressionWithValues(*expression.Right, existing, values)
	return fmt.Sprintf("Number(%s ?? 0) %s Number(%s ?? 0)", left, expression.Operator, right)
}

func (g *webGenerator) apiCoreExpression(expression ExpressionDecl, existing string, numeric bool) string {
	return g.apiCoreExpressionWithValues(expression, existing, numeric, nil)
}

func (g *webGenerator) apiCoreExpressionWithValues(expression ExpressionDecl, existing string, numeric bool, values map[string]string) string {
	return renderCoreExpressionJS(expression, func(node ExpressionDecl) string {
		return g.apiOperandExpressionWithValues(apiOperandFromExpression(node), existing, values)
	}, numeric)
}

func apiExpressionDivisionDivisors(expression APIExpressionDecl) []ExpressionDecl {
	if expression.Tree != nil {
		return expressionDivisionDivisors(*expression.Tree)
	}
	if expression.Operator == "/" && expression.Right != nil {
		return []ExpressionDecl{{
			Kind:      "operand",
			ValueKind: expression.Right.Kind,
			Value:     expression.Right.Value,
			Position:  expression.Position,
		}}
	}
	return nil
}

func (g *webGenerator) apiOperandExpression(operand APIOperandDecl, existing string) string {
	return g.apiOperandExpressionWithValues(operand, existing, nil)
}

func (g *webGenerator) apiOperandExpressionWithValues(operand APIOperandDecl, existing string, values map[string]string) string {
	switch operand.Kind {
	case "string":
		return contractJSONString(operand.Value)
	case "number", "boolean":
		return operand.Value
	case "body":
		return fmt.Sprintf("bodyValues.value[%q]", operand.Value)
	case "param":
		return fmt.Sprintf("params.value[%q]", operand.Value)
	case "field":
		if values != nil {
			if _, ok := values[operand.Value]; ok {
				return tsSafeIdentifier(operand.Value)
			}
		}
		return existing + "." + operand.Value
	default:
		return "undefined"
	}
}

func (g *webGenerator) apiExpressionRuntimeType(api APIDecl, entity EntityDecl, expression APIExpressionDecl, values map[string]string) string {
	if expression.Tree != nil {
		if expression.Tree.Kind == "binary" {
			return "number"
		}
		return g.apiOperandRuntimeType(api, entity, apiOperandFromExpression(*expression.Tree), values)
	}
	if expression.Operator != "" {
		return "number"
	}
	return g.apiOperandRuntimeType(api, entity, expression.Left, values)
}

func (g *webGenerator) apiOperandRuntimeType(api APIDecl, entity EntityDecl, operand APIOperandDecl, values map[string]string) string {
	switch operand.Kind {
	case "string":
		return "text"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "body":
		for _, field := range api.Body {
			if field.Name == operand.Value {
				return field.Type
			}
		}
	case "param":
		for _, param := range api.Params {
			if param.Name == operand.Value {
				return param.Type
			}
		}
	case "field":
		if values != nil {
			if valueType, ok := values[operand.Value]; ok {
				return valueType
			}
		}
		if field, ok := fieldIndex(entity)[operand.Value]; ok {
			return field.Type
		}
	}
	return ""
}

func apiExpressionSingleOperand(expression APIExpressionDecl) (APIOperandDecl, bool) {
	if expression.Tree != nil {
		if expression.Tree.Kind == "binary" {
			return APIOperandDecl{}, false
		}
		return apiOperandFromExpression(*expression.Tree), true
	}
	if expression.Operator == "" && expression.Right == nil {
		return expression.Left, true
	}
	return APIOperandDecl{}, false
}

func (g *webGenerator) apiExpressionNeedsNumericNarrowing(api APIDecl, entity EntityDecl, expression APIExpressionDecl, values map[string]string) bool {
	if !numberLikeType(g.apiExpressionRuntimeType(api, entity, expression, values)) {
		return false
	}
	operand, ok := apiExpressionSingleOperand(expression)
	if !ok {
		return false
	}
	switch operand.Kind {
	case "body", "param":
		return true
	case "field":
		if values != nil {
			if _, local := values[operand.Value]; local {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (g *webGenerator) apiLocalValueExpression(api APIDecl, entity EntityDecl, expression APIExpressionDecl, existing string, values map[string]string) string {
	rendered := g.apiExpressionWithValues(expression, existing, values)
	if g.apiExpressionNeedsNumericNarrowing(api, entity, expression, values) {
		return fmt.Sprintf("Number(%s ?? 0)", rendered)
	}
	return rendered
}

func (g *webGenerator) apiConditionSideExpression(api APIDecl, entity EntityDecl, expression APIExpressionDecl, existing string, values map[string]string, ordered bool) string {
	rendered := g.apiExpressionWithValues(expression, existing, values)
	if ordered && g.apiExpressionNeedsNumericNarrowing(api, entity, expression, values) {
		return fmt.Sprintf("Number(%s ?? 0)", rendered)
	}
	return rendered
}

func (g *webGenerator) apiStatementsRuntime(api APIDecl, entity EntityDecl, statements []APIStatementDecl, indent string, existing string, divisionMode string, guardCounter *int, values map[string]string) string {
	var builder strings.Builder
	for _, statement := range statements {
		switch statement.Kind {
		case "value":
			if statement.Value == nil {
				continue
			}
			builder.WriteString(g.apiExpressionGuards(api, statement.Value.Expression, indent, existing, divisionMode, guardCounter, values))
			valueName := tsSafeIdentifier(statement.Value.Name)
			valueType := g.apiExpressionRuntimeType(api, entity, statement.Value.Expression, values)
			builder.WriteString(fmt.Sprintf("%sconst %s = %s;\n", indent, valueName, g.apiLocalValueExpression(api, entity, statement.Value.Expression, existing, values)))
			values[statement.Value.Name] = valueType
		case "set":
			if statement.Set == nil {
				continue
			}
			builder.WriteString(g.apiExpressionGuards(api, statement.Set.Expr, indent, existing, divisionMode, guardCounter, values))
			builder.WriteString(fmt.Sprintf("%sdata.%s = %s;\n", indent, statement.Set.Field, g.apiExpressionWithValues(statement.Set.Expr, existing, values)))
		case "if":
			if statement.If == nil {
				continue
			}
			builder.WriteString(g.apiConditionGuards(api, statement.If.Condition, indent, existing, divisionMode, guardCounter, values))
			builder.WriteString(fmt.Sprintf("%sif (%s) {\n", indent, g.apiConditionExpression(api, entity, statement.If.Condition, existing, values)))
			thenValues := cloneStringMap(values)
			builder.WriteString(g.apiStatementsRuntime(api, entity, statement.If.Then, indent+"  ", existing, divisionMode, guardCounter, thenValues))
			if len(statement.If.Else) > 0 {
				builder.WriteString(indent + "} else {\n")
				elseValues := cloneStringMap(values)
				builder.WriteString(g.apiStatementsRuntime(api, entity, statement.If.Else, indent+"  ", existing, divisionMode, guardCounter, elseValues))
				builder.WriteString(indent + "}\n")
			} else {
				builder.WriteString(indent + "}\n")
			}
		}
	}
	return builder.String()
}

func (g *webGenerator) apiConditionExpression(api APIDecl, entity EntityDecl, condition APIConditionDecl, existing string, values map[string]string) string {
	if condition.Tree != nil {
		return g.apiConditionTreeExpression(api, entity, *condition.Tree, existing, values)
	}
	return g.apiConditionComparisonExpression(api, entity, condition, existing, values)
}

func (g *webGenerator) apiConditionComparisonExpression(api APIDecl, entity EntityDecl, condition APIConditionDecl, existing string, values map[string]string) string {
	ordered := condition.Operator == "<" || condition.Operator == "<=" || condition.Operator == ">" || condition.Operator == ">="
	left := g.apiConditionSideExpression(api, entity, condition.Left, existing, values, ordered)
	right := g.apiConditionSideExpression(api, entity, condition.Right, existing, values, ordered)
	return fmt.Sprintf("%s %s %s", left, condition.Operator, right)
}

func (g *webGenerator) apiConditionTreeExpression(api APIDecl, entity EntityDecl, condition ConditionExpressionDecl, existing string, values map[string]string) string {
	switch condition.Kind {
	case "comparison":
		if condition.Comparison == nil {
			return "false"
		}
		return g.apiConditionComparisonExpression(api, entity, apiConditionFromComparison(*condition.Comparison), existing, values)
	case "and":
		if condition.Left == nil || condition.Right == nil {
			return "false"
		}
		return fmt.Sprintf("(%s && %s)", g.apiConditionTreeExpression(api, entity, *condition.Left, existing, values), g.apiConditionTreeExpression(api, entity, *condition.Right, existing, values))
	case "or":
		if condition.Left == nil || condition.Right == nil {
			return "false"
		}
		return fmt.Sprintf("(%s || %s)", g.apiConditionTreeExpression(api, entity, *condition.Left, existing, values), g.apiConditionTreeExpression(api, entity, *condition.Right, existing, values))
	case "not":
		if condition.Left == nil {
			return "false"
		}
		return fmt.Sprintf("!(%s)", g.apiConditionTreeExpression(api, entity, *condition.Left, existing, values))
	default:
		return "false"
	}
}

func (g *webGenerator) apiConditionGuards(api APIDecl, condition APIConditionDecl, indent string, existing string, divisionMode string, guardCounter *int, values map[string]string) string {
	if condition.Tree != nil {
		return g.apiConditionTreeGuards(api, *condition.Tree, indent, existing, divisionMode, guardCounter, values)
	}
	return g.apiConditionComparisonGuards(api, condition, indent, existing, divisionMode, guardCounter, values)
}

func (g *webGenerator) apiConditionComparisonGuards(api APIDecl, condition APIConditionDecl, indent string, existing string, divisionMode string, guardCounter *int, values map[string]string) string {
	var builder strings.Builder
	builder.WriteString(g.apiExpressionGuards(api, condition.Left, indent, existing, divisionMode, guardCounter, values))
	builder.WriteString(g.apiExpressionGuards(api, condition.Right, indent, existing, divisionMode, guardCounter, values))
	return builder.String()
}

func (g *webGenerator) apiConditionTreeGuards(api APIDecl, condition ConditionExpressionDecl, indent string, existing string, divisionMode string, guardCounter *int, values map[string]string) string {
	var builder strings.Builder
	switch condition.Kind {
	case "comparison":
		if condition.Comparison != nil {
			builder.WriteString(g.apiConditionComparisonGuards(api, apiConditionFromComparison(*condition.Comparison), indent, existing, divisionMode, guardCounter, values))
		}
	case "and", "or":
		if condition.Left != nil {
			builder.WriteString(g.apiConditionTreeGuards(api, *condition.Left, indent, existing, divisionMode, guardCounter, values))
		}
		if condition.Right != nil {
			builder.WriteString(g.apiConditionTreeGuards(api, *condition.Right, indent, existing, divisionMode, guardCounter, values))
		}
	case "not":
		if condition.Left != nil {
			builder.WriteString(g.apiConditionTreeGuards(api, *condition.Left, indent, existing, divisionMode, guardCounter, values))
		}
	}
	return builder.String()
}

func (g *webGenerator) apiExpressionGuards(api APIDecl, expression APIExpressionDecl, indent string, existing string, divisionMode string, guardCounter *int, values map[string]string) string {
	var builder strings.Builder
	for _, divisor := range apiExpressionDivisionDivisors(expression) {
		*guardCounter = *guardCounter + 1
		guardName := fmt.Sprintf("%sLogicDivisor%d", lowerCamelCase(api.Name), *guardCounter)
		builder.WriteString(fmt.Sprintf("%sconst %s = %s;\n", indent, guardName, g.apiCoreExpressionWithValues(divisor, existing, true, values)))
		builder.WriteString(fmt.Sprintf("%sif (%s === 0) {\n", indent, guardName))
		if divisionMode == "transaction" {
			builder.WriteString(indent + "  return { status: \"divisionByZero\" as const };\n")
		} else {
			builder.WriteString(indent + "  res.status(400).json({ error: \"Division by zero\" });\n")
			builder.WriteString(indent + "  return;\n")
		}
		builder.WriteString(indent + "}\n")
	}
	return builder.String()
}

func cloneStringMap(values map[string]string) map[string]string {
	clone := map[string]string{}
	for name, value := range values {
		clone[name] = value
	}
	return clone
}

func apiConditionDivisionDivisors(condition APIConditionDecl) []ExpressionDecl {
	if condition.Tree != nil {
		return conditionExpressionDivisionDivisors(*condition.Tree)
	}
	divisors := []ExpressionDecl{}
	divisors = append(divisors, apiExpressionDivisionDivisors(condition.Left)...)
	divisors = append(divisors, apiExpressionDivisionDivisors(condition.Right)...)
	return divisors
}

func explicitAPIStatusCode(api APIDecl) int {
	if api.Respond == "accepted" || (api.Respond == "" && api.Webhook) {
		return 202
	}
	return 200
}

func explicitAPIStatusText(api APIDecl) string {
	if api.Respond != "" {
		return api.Respond
	}
	if api.Webhook {
		return "accepted"
	}
	if api.Update != nil {
		return "updated"
	}
	return "declared"
}

func (g *webGenerator) explicitAPIBodySpecsLiteral(fields []FieldDecl) string {
	parts := []string{}
	for _, field := range fields {
		properties := []string{
			fmt.Sprintf("name: %s", contractJSONString(field.Name)),
			fmt.Sprintf("type: %s", contractJSONString(field.Type)),
			fmt.Sprintf("required: %t", hasModifier(field, "required")),
		}
		if minValue := modifierValue(field, "min"); minValue != "" {
			properties = append(properties, "min: "+minValue)
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			properties = append(properties, "max: "+maxValue)
		}
		if minLength, maxLength, ok := fieldLengthBounds(field); ok {
			properties = append(properties, fmt.Sprintf("minLength: %d", minLength), fmt.Sprintf("maxLength: %d", maxLength))
		}
		if pattern := modifierValue(field, "regex"); pattern != "" {
			properties = append(properties, fmt.Sprintf("pattern: %s", contractJSONString(pattern)))
		}
		if hasModifier(field, "url") {
			properties = append(properties, "url: true")
		}
		if message := modifierValue(field, "message"); message != "" {
			properties = append(properties, fmt.Sprintf("message: %s", contractJSONString(message)))
		}
		parts = append(parts, "{ "+strings.Join(properties, ", ")+" }")
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func explicitAPIParamSpecsLiteral(params []APIParamDecl, required bool) string {
	parts := []string{}
	for _, param := range params {
		parts = append(parts, fmt.Sprintf("{ name: %s, type: %s, required: %t }", contractJSONString(param.Name), contractJSONString(param.Type), required))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func explicitAPIHasBody(method string) bool {
	switch strings.ToUpper(method) {
	case "POST", "PUT", "PATCH":
		return true
	default:
		return false
	}
}

func expressAPIPath(path string) string {
	var builder strings.Builder
	for index := 0; index < len(path); index++ {
		if path[index] != '{' {
			builder.WriteByte(path[index])
			continue
		}
		end := strings.IndexByte(path[index:], '}')
		if end <= 0 {
			builder.WriteByte(path[index])
			continue
		}
		name := path[index+1 : index+end]
		builder.WriteString(":")
		builder.WriteString(name)
		index += end
	}
	return builder.String()
}

func explicitAPISmokeTests(g *webGenerator) string {
	if !g.hasExplicitAPIs() {
		return ""
	}
	var builder strings.Builder
	for _, api := range g.program.APIs {
		method := strings.ToUpper(api.Method)
		if !supportedAPIMethods[method] || api.Path == "" {
			continue
		}
		varName := lowerCamelCase(api.Name) + "Response"
		url := explicitAPISmokeURL(api)
		fetchOptions := explicitAPIFetchOptions(api)
		if explicitAPIAccess(api) == "private" && g.program.Auth != nil {
			builder.WriteString(fmt.Sprintf("  const %s = await fetch(baseURL + %s%s);\n", varName, contractJSONString(url), fetchOptions))
			builder.WriteString(fmt.Sprintf("  assert.equal(%s.status, 401, %q);\n\n", varName, api.Name+" should reject anonymous private API requests"))
			continue
		}
		if api.Update != nil && apiHasRequiredBody(api) {
			builder.WriteString(fmt.Sprintf("  const %s = await fetch(baseURL + %s, { method: %s, headers: { \"content-type\": \"application/json\" }, body: JSON.stringify({}) });\n", varName, contractJSONString(url), contractJSONString(method)))
			builder.WriteString(fmt.Sprintf("  assert.equal(%s.status, 400, %q);\n\n", varName, api.Name+" should validate explicit API handler body before mutation"))
			continue
		}
		expectedStatus := explicitAPIStatusCode(api)
		expectedStatusText := explicitAPIStatusText(api)
		bodyVar := lowerCamelCase(api.Name) + "Body"
		builder.WriteString(fmt.Sprintf("  const %s = await fetch(baseURL + %s%s);\n", varName, contractJSONString(url), fetchOptions))
		builder.WriteString(fmt.Sprintf("  assert.equal(%s.status, %d, %q);\n", varName, expectedStatus, api.Name+" should expose generated explicit API runtime"))
		builder.WriteString(fmt.Sprintf("  const %s = await %s.json() as { api?: string; status?: string; runtime?: string };\n", bodyVar, varName))
		builder.WriteString(fmt.Sprintf("  assert.equal(%s.api, %s);\n", bodyVar, contractJSONString(api.Name)))
		builder.WriteString(fmt.Sprintf("  assert.equal(%s.status, %s);\n", bodyVar, contractJSONString(expectedStatusText)))
		builder.WriteString(fmt.Sprintf("  assert.equal(%s.runtime, \"declared\");\n\n", bodyVar))
	}
	return builder.String()
}

func apiHasRequiredBody(api APIDecl) bool {
	for _, field := range api.Body {
		if hasModifier(field, "required") {
			return true
		}
	}
	return false
}

func explicitAPISmokeURL(api APIDecl) string {
	path := api.Path
	for _, param := range api.Params {
		path = strings.ReplaceAll(path, "{"+param.Name+"}", explicitAPISampleValue(param.Type))
	}
	if len(api.Queries) == 0 {
		return path
	}
	query := []string{}
	for _, param := range api.Queries {
		query = append(query, param.Name+"="+explicitAPISampleValue(param.Type))
	}
	return path + "?" + strings.Join(query, "&")
}

func explicitAPISampleValue(paramType string) string {
	switch paramType {
	case "number", "integer":
		return "7"
	case "decimal", "money":
		return "12.5"
	case "boolean":
		return "true"
	case "date":
		return "2026-01-01"
	case "datetime":
		return "2026-01-01T00:00:00.000Z"
	case "email":
		return "agent@example.com"
	default:
		return "sample"
	}
}

func explicitAPIFetchOptions(api APIDecl) string {
	method := strings.ToUpper(api.Method)
	if method == "GET" {
		return ""
	}
	if explicitAPIHasBody(method) {
		body := "{ blacklang: true }"
		if len(api.Body) > 0 {
			body = explicitAPISampleBodyLiteral(api.Body)
		}
		return fmt.Sprintf(", { method: %s, headers: { \"content-type\": \"application/json\" }, body: JSON.stringify(%s) }", contractJSONString(method), body)
	}
	return fmt.Sprintf(", { method: %s }", contractJSONString(method))
}

func explicitAPISampleBodyLiteral(fields []FieldDecl) string {
	parts := []string{}
	for _, field := range fields {
		parts = append(parts, fmt.Sprintf("%s: %s", field.Name, explicitAPISampleBodyValue(field.Type)))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func explicitAPISampleBodyValue(fieldType string) string {
	switch fieldType {
	case "number", "integer":
		return "7"
	case "decimal", "money":
		return "12.5"
	case "boolean":
		return "true"
	case "date":
		return contractJSONString("2026-01-01")
	case "datetime":
		return contractJSONString("2026-01-01T00:00:00.000Z")
	case "email":
		return contractJSONString("agent@example.com")
	default:
		return contractJSONString("sample")
	}
}
