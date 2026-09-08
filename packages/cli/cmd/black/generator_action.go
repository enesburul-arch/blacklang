package main

import (
	"fmt"
	"strings"
)

func (g *webGenerator) hasCustomActions() bool {
	return len(g.program.Actions) > 0
}

func (g *webGenerator) validationImportNames(entity EntityDecl) string {
	names := []string{"validate" + entity.Name + "Input"}
	for _, action := range customActionsForEntity(g.program, entity.Name) {
		names = append(names, "validate"+action.Name+"Input")
	}
	return strings.Join(names, ", ")
}

func (g *webGenerator) actionInputTypes() string {
	var builder strings.Builder
	for _, action := range g.program.Actions {
		builder.WriteString(fmt.Sprintf("export type %sInput = {\n", action.Name))
		for _, input := range action.Inputs {
			optional := "?"
			if hasModifier(input, "required") && modifierValue(input, "default") == "" {
				optional = ""
			}
			builder.WriteString(fmt.Sprintf("  %s%s: %s;\n", input.Name, optional, tsType(input.Type)))
		}
		builder.WriteString("};\n\n")
	}
	return builder.String()
}

func (g *webGenerator) actionValidationFunctions(entity EntityDecl) string {
	var builder strings.Builder
	for _, action := range customActionsForEntity(g.program, entity.Name) {
		builder.WriteString(fmt.Sprintf("\nexport function validate%sInput(input: any): ValidationResult<Record<string, unknown>> {\n", action.Name))
		builder.WriteString("  const errors: string[] = [];\n")
		builder.WriteString("  const value: any = {};\n\n")
		for _, field := range action.Inputs {
			builder.WriteString(g.actionValidationFieldBlock(action, field))
		}
		builder.WriteString("  if (errors.length > 0) {\n")
		builder.WriteString("    return { valid: false, errors };\n")
		builder.WriteString("  }\n\n")
		builder.WriteString("  return { valid: true, value };\n")
		builder.WriteString("}\n")
	}
	return builder.String()
}

func (g *webGenerator) actionValidationFieldBlock(action CustomActionDecl, field FieldDecl) string {
	var builder strings.Builder
	name := field.Name
	builder.WriteString(fmt.Sprintf("  if (input.%s === undefined || input.%s === null || input.%s === \"\") {\n", name, name, name))
	if defaultValue := modifierValue(field, "default"); defaultValue != "" {
		builder.WriteString(fmt.Sprintf("    value.%s = %s;\n", name, typedLiteral(field, defaultValue)))
	} else if hasModifier(field, "required") {
		builder.WriteString(fmt.Sprintf("    errors.push(%q);\n", fieldValidationMessage(field, action.Name+"."+name+" is required")))
	} else {
		builder.WriteString(fmt.Sprintf("    value.%s = undefined;\n", name))
	}
	builder.WriteString("  } else {\n")
	switch field.Type {
	case "number", "integer":
		builder.WriteString(fmt.Sprintf("    const parsed = Number(input.%s);\n", name))
		builder.WriteString("    if (!Number.isInteger(parsed) || parsed < -2147483648 || parsed > 2147483647) {\n")
		builder.WriteString(fmt.Sprintf("      errors.push(%q);\n", fieldValidationMessage(field, action.Name+"."+name+" must be a whole number")))
		builder.WriteString("    } else {\n")
		if minValue := modifierValue(field, "min"); minValue != "" {
			builder.WriteString(fmt.Sprintf("      if (parsed < %s) errors.push(%q);\n", minValue, fieldValidationMessage(field, action.Name+"."+name+" must be at least "+minValue)))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			builder.WriteString(fmt.Sprintf("      if (parsed > %s) errors.push(%q);\n", maxValue, fieldValidationMessage(field, action.Name+"."+name+" must be at most "+maxValue)))
		}
		builder.WriteString(fmt.Sprintf("      value.%s = parsed;\n", name))
		builder.WriteString("    }\n")
	case "decimal", "money":
		builder.WriteString(fmt.Sprintf("    const parsed = Number(input.%s);\n", name))
		builder.WriteString("    if (!Number.isFinite(parsed)) {\n")
		builder.WriteString(fmt.Sprintf("      errors.push(%q);\n", fieldValidationMessage(field, action.Name+"."+name+" must be a number")))
		builder.WriteString("    } else {\n")
		if minValue := modifierValue(field, "min"); minValue != "" {
			builder.WriteString(fmt.Sprintf("      if (parsed < %s) errors.push(%q);\n", minValue, fieldValidationMessage(field, action.Name+"."+name+" must be at least "+minValue)))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			builder.WriteString(fmt.Sprintf("      if (parsed > %s) errors.push(%q);\n", maxValue, fieldValidationMessage(field, action.Name+"."+name+" must be at most "+maxValue)))
		}
		builder.WriteString(fmt.Sprintf("      value.%s = parsed;\n", name))
		builder.WriteString("    }\n")
	case "boolean":
		builder.WriteString(fmt.Sprintf("    value.%s = Boolean(input.%s);\n", name, name))
	case "email":
		builder.WriteString(fmt.Sprintf("    value.%s = String(input.%s);\n", name, name))
		builder.WriteString(fmt.Sprintf("    if (!value.%s.includes(\"@\")) errors.push(%q);\n", name, fieldValidationMessage(field, action.Name+"."+name+" must be an email")))
		builder.WriteString(g.actionTextConstraints(field, name, action.Name+"."+name))
	default:
		builder.WriteString(fmt.Sprintf("    value.%s = String(input.%s);\n", name, name))
		builder.WriteString(g.actionTextConstraints(field, name, action.Name+"."+name))
	}
	builder.WriteString("  }\n\n")
	return builder.String()
}

func (g *webGenerator) actionTextConstraints(field FieldDecl, name string, label string) string {
	var builder strings.Builder
	if minLength, maxLength, ok := fieldLengthBounds(field); ok {
		builder.WriteString(fmt.Sprintf("    if (value.%s.length < %d || value.%s.length > %d) errors.push(%q);\n", name, minLength, name, maxLength, fieldValidationMessage(field, fmt.Sprintf("%s length must be between %d and %d", label, minLength, maxLength))))
	}
	if hasModifier(field, "url") {
		builder.WriteString(fmt.Sprintf("    try { new URL(value.%s); } catch { errors.push(%q); }\n", name, fieldValidationMessage(field, label+" must be a valid URL")))
	}
	if pattern := modifierValue(field, "regex"); pattern != "" {
		builder.WriteString(fmt.Sprintf("    if (!(new RegExp(%q)).test(value.%s)) errors.push(%q);\n", pattern, name, fieldValidationMessage(field, label+" has an invalid format")))
	}
	return builder.String()
}

func (g *webGenerator) customActionRoutes(page PageDecl, entity EntityDecl) string {
	actions := customActionsForPage(g.program, page)
	if len(actions) == 0 {
		return ""
	}

	identifier := lowerCamelCase(entity.Name)
	path := strings.ToLower(page.Name)
	var builder strings.Builder
	for _, action := range actions {
		builder.WriteString(fmt.Sprintf("%sRouter.post(\"/%s/:id/actions/%s\", %sasync (req, res) => {\n", identifier, path, strings.ToLower(action.Name), g.permissionMiddleware("update", entity.Name)))
		if g.hasRuntimePermissions() {
			builder.WriteString(g.customActionAllowGuard(action, "  "))
			for _, field := range customActionSetFieldNames(action) {
				builder.WriteString(fmt.Sprintf("  if (!canAccessField(currentRoles(req), \"update\", %q, %q)) {\n", entity.Name, field))
				builder.WriteString("    res.status(403).json({ error: \"Forbidden\" });\n")
				builder.WriteString("    return;\n")
				builder.WriteString("  }\n\n")
			}
		}
		builder.WriteString(fmt.Sprintf("  const validation = validate%sInput(req.body);\n", action.Name))
		builder.WriteString("  if (!validation.valid) {\n")
		builder.WriteString("    res.status(400).json({ error: validation.errors });\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n\n")
		if _, transactional := g.customActionTransactionName(action); transactional {
			builder.WriteString(g.transactionalCustomActionRouteBody(action, entity, identifier))
		} else {
			builder.WriteString(g.standardCustomActionRouteBody(action, entity, identifier))
		}
		builder.WriteString("});\n\n")
	}
	return builder.String()
}

func (g *webGenerator) standardCustomActionRouteBody(action CustomActionDecl, entity EntityDecl, identifier string) string {
	var builder strings.Builder
	findMethod := "findUnique"
	whereExpression := "{ id: String(req.params.id) }"
	if g.hasEntityPolicies(entity) {
		findMethod = "findFirst"
		whereExpression = g.rowPolicyWhere(entity, "{ id: String(req.params.id) }")
	}
	builder.WriteString(fmt.Sprintf("  const existing = await %sModel.%s({ where: %s });\n", identifier, findMethod, whereExpression))
	builder.WriteString("  if (!existing) {\n")
	builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n\n")
	builder.WriteString("  const actionInput = validation.value as any;\n")
	builder.WriteString("  const data: Record<string, unknown> = {};\n")
	guardCounter := 0
	builder.WriteString(g.actionStatementsRuntime(action, entity, actionStatements(action), "  ", "existing", "actionInput", "http", &guardCounter, map[string]string{}))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("  const item = await %sModel.update({\n", identifier))
	if g.hasEntityPolicies(entity) {
		builder.WriteString("    where: { id: existing.id },\n")
	} else {
		builder.WriteString("    where: { id: String(req.params.id) },\n")
	}
	builder.WriteString("    data: data as any")
	builder.WriteString("\n")
	builder.WriteString("  });\n\n")
	builder.WriteString(g.customActionResponseBlock(action, entity, identifier, "  "))
	return builder.String()
}

func (g *webGenerator) transactionalCustomActionRouteBody(action CustomActionDecl, entity EntityDecl, identifier string) string {
	var builder strings.Builder
	txModel := identifier + "TxModel"
	findMethod := "findUnique"
	whereExpression := "{ id: String(req.params.id) }"
	if g.hasEntityPolicies(entity) {
		findMethod = "findFirst"
		whereExpression = g.rowPolicyWhere(entity, "{ id: String(req.params.id) }")
	}
	success := customActionSuccessMessage(action, identifier)
	builder.WriteString("  const actionInput = validation.value as any;\n")
	builder.WriteString("  const actionResult = await prisma.$transaction(async (tx) => {\n")
	builder.WriteString(fmt.Sprintf("    const %s = tx.%s;\n", txModel, identifier))
	builder.WriteString(fmt.Sprintf("    const existing = await %s.%s({ where: %s });\n", txModel, findMethod, whereExpression))
	builder.WriteString("    if (!existing) {\n")
	builder.WriteString("      return { status: \"notFound\" as const };\n")
	builder.WriteString("    }\n\n")
	builder.WriteString("    const data: Record<string, unknown> = {};\n")
	guardCounter := 0
	builder.WriteString(g.actionStatementsRuntime(action, entity, actionStatements(action), "    ", "existing", "actionInput", "transaction", &guardCounter, map[string]string{}))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("    const item = await %s.update({\n", txModel))
	if g.hasEntityPolicies(entity) {
		builder.WriteString("      where: { id: existing.id },\n")
	} else {
		builder.WriteString("      where: { id: String(req.params.id) },\n")
	}
	builder.WriteString("      data: data as any")
	builder.WriteString("\n")
	builder.WriteString("    });\n")
	if g.hasRuntimePermissions() {
		builder.WriteString("    const actor = currentUser(req);\n")
		builder.WriteString("    if (actor) {\n")
		if g.isMySQL() {
			builder.WriteString("      await tx.blackAuditLog.create({\n")
			builder.WriteString("        data: {\n")
			builder.WriteString("          id: crypto.randomUUID(),\n")
			builder.WriteString("          actorUserId: actor.id,\n")
			builder.WriteString("          actorRole: actor.role,\n")
			builder.WriteString(fmt.Sprintf("          action: %q,\n", "action."+action.Name))
			builder.WriteString(fmt.Sprintf("          resource: %q,\n", entity.Name))
			builder.WriteString("          resourceId: item.id,\n")
			builder.WriteString(fmt.Sprintf("          summary: %q\n", success))
			builder.WriteString("        }\n")
			builder.WriteString("      });\n")
		} else {
			builder.WriteString("      await tx.$executeRaw`\n")
			builder.WriteString("        INSERT INTO \"BlackAuditLog\" (\"id\", \"actorUserId\", \"actorRole\", \"action\", \"resource\", \"resourceId\", \"summary\")\n")
			builder.WriteString(fmt.Sprintf("        VALUES (${crypto.randomUUID()}, ${actor.id}, ${actor.role}, ${%q}, ${%q}, ${item.id}, ${%q})\n", "action."+action.Name, entity.Name, success))
			builder.WriteString("      `;\n")
		}
		builder.WriteString("    }\n")
	}
	builder.WriteString("    return { status: \"ok\" as const, item };\n")
	builder.WriteString("  });\n\n")
	builder.WriteString("  if (actionResult.status === \"notFound\") {\n")
	builder.WriteString(fmt.Sprintf("    res.status(404).json({ error: \"%s not found\" });\n", entity.Name))
	builder.WriteString("    return;\n")
	builder.WriteString("  }\n")
	if actionHasDivisionGuard(action) {
		builder.WriteString("  if (actionResult.status === \"divisionByZero\") {\n")
		builder.WriteString("    res.status(400).json({ error: \"Division by zero\" });\n")
		builder.WriteString("    return;\n")
		builder.WriteString("  }\n")
	}
	builder.WriteString("\n")
	builder.WriteString("  const item = actionResult.item;\n")
	builder.WriteString(g.customActionResponseBlock(action, entity, identifier, "  "))
	return builder.String()
}

func (g *webGenerator) customActionResponseBlock(action CustomActionDecl, entity EntityDecl, identifier string, indent string) string {
	var builder strings.Builder
	success := customActionSuccessMessage(action, identifier)
	builder.WriteString(g.relationLoadAttachStatement(entity, "mutation", "[item]", indent))
	if g.hasRuntimePermissions() {
		if _, transactional := g.customActionTransactionName(action); !transactional {
			builder.WriteString(fmt.Sprintf(indent+"writeAuditLog(currentUser(req), %q, %q, item.id, %q);\n", "action."+action.Name, entity.Name, success))
		}
		builder.WriteString(fmt.Sprintf(indent+"res.json({ item: sanitize%s(item, currentRoles(req)), message: %q });\n", entity.Name, success))
	} else {
		builder.WriteString(fmt.Sprintf(indent+"res.json({ item, message: %q });\n", success))
	}
	return builder.String()
}

func customActionSuccessMessage(action CustomActionDecl, identifier string) string {
	if action.Success != "" {
		return action.Success
	}
	return identifierLabel(action.Name) + " complete"
}

func actionHasDivisionGuard(action CustomActionDecl) bool {
	return actionStatementsHaveDivisionGuard(actionStatements(action))
}

func actionStatementsHaveDivisionGuard(statements []ActionStatementDecl) bool {
	for _, statement := range statements {
		switch statement.Kind {
		case "value":
			if statement.Value != nil && len(actionExpressionDivisionDivisors(statement.Value.Expression)) > 0 {
				return true
			}
		case "set":
			if statement.Set != nil && len(actionExpressionDivisionDivisors(statement.Set.Expression)) > 0 {
				return true
			}
		case "if":
			if statement.If == nil {
				continue
			}
			if len(actionConditionDivisionDivisors(statement.If.Condition)) > 0 {
				return true
			}
			if actionStatementsHaveDivisionGuard(statement.If.Then) || actionStatementsHaveDivisionGuard(statement.If.Else) {
				return true
			}
		}
	}
	return false
}

func (g *webGenerator) routeUsesTransactionalCustomActionAudit(page PageDecl) bool {
	if !g.hasRuntimePermissions() {
		return false
	}
	for _, action := range customActionsForPage(g.program, page) {
		if _, transactional := g.customActionTransactionName(action); transactional {
			return true
		}
	}
	return false
}
func (g *webGenerator) customActionAllowGuard(action CustomActionDecl, indent string) string {
	if len(action.Allow) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(indent + "const user = currentUser(req);\n")
	if containsString(action.Allow, "authenticated") {
		builder.WriteString(indent + "if (!user) {\n")
	} else {
		builder.WriteString(fmt.Sprintf(indent+"if (!user || !(user.roles ?? [user.role]).some((role) => %s.includes(role))) {\n", tsStringArrayLiteral(action.Allow)))
	}
	builder.WriteString(indent + "  res.status(403).json({ error: \"Forbidden\" });\n")
	builder.WriteString(indent + "  return;\n")
	builder.WriteString(indent + "}\n\n")
	return builder.String()
}

func (g *webGenerator) actionStatementsRuntime(action CustomActionDecl, entity EntityDecl, statements []ActionStatementDecl, indent string, existing string, input string, divisionMode string, guardCounter *int, values map[string]string) string {
	var builder strings.Builder
	for _, statement := range statements {
		switch statement.Kind {
		case "value":
			if statement.Value == nil {
				continue
			}
			builder.WriteString(g.actionExpressionGuards(action, statement.Value.Expression, indent, existing, input, divisionMode, guardCounter, values))
			valueName := tsSafeIdentifier(statement.Value.Name)
			valueType := g.actionExpressionRuntimeType(action, entity, statement.Value.Expression, values)
			builder.WriteString(fmt.Sprintf("%sconst %s = %s;\n", indent, valueName, g.actionLocalValueExpression(action, entity, statement.Value.Expression, existing, input, values)))
			values[statement.Value.Name] = valueType
		case "set":
			if statement.Set == nil {
				continue
			}
			builder.WriteString(g.actionExpressionGuards(action, statement.Set.Expression, indent, existing, input, divisionMode, guardCounter, values))
			builder.WriteString(fmt.Sprintf("%sdata.%s = %s;\n", indent, statement.Set.Field, g.actionExpressionWithValues(action, statement.Set.Expression, existing, input, values)))
		case "if":
			if statement.If == nil {
				continue
			}
			builder.WriteString(g.actionConditionGuards(action, statement.If.Condition, indent, existing, input, divisionMode, guardCounter, values))
			builder.WriteString(fmt.Sprintf("%sif (%s) {\n", indent, g.actionConditionExpression(action, entity, statement.If.Condition, existing, input, values)))
			thenValues := cloneStringMap(values)
			builder.WriteString(g.actionStatementsRuntime(action, entity, statement.If.Then, indent+"  ", existing, input, divisionMode, guardCounter, thenValues))
			if len(statement.If.Else) > 0 {
				builder.WriteString(indent + "} else {\n")
				elseValues := cloneStringMap(values)
				builder.WriteString(g.actionStatementsRuntime(action, entity, statement.If.Else, indent+"  ", existing, input, divisionMode, guardCounter, elseValues))
				builder.WriteString(indent + "}\n")
			} else {
				builder.WriteString(indent + "}\n")
			}
		}
	}
	return builder.String()
}

func (g *webGenerator) actionConditionExpression(action CustomActionDecl, entity EntityDecl, condition ActionConditionDecl, existing string, input string, values map[string]string) string {
	if condition.Tree != nil {
		return g.actionConditionTreeExpression(action, entity, *condition.Tree, existing, input, values)
	}
	return g.actionConditionComparisonExpression(action, entity, condition, existing, input, values)
}

func (g *webGenerator) actionConditionComparisonExpression(action CustomActionDecl, entity EntityDecl, condition ActionConditionDecl, existing string, input string, values map[string]string) string {
	ordered := condition.Operator == "<" || condition.Operator == "<=" || condition.Operator == ">" || condition.Operator == ">="
	left := g.actionConditionSideExpression(action, entity, condition.Left, existing, input, values, ordered)
	right := g.actionConditionSideExpression(action, entity, condition.Right, existing, input, values, ordered)
	return fmt.Sprintf("%s %s %s", left, condition.Operator, right)
}

func (g *webGenerator) actionConditionTreeExpression(action CustomActionDecl, entity EntityDecl, condition ConditionExpressionDecl, existing string, input string, values map[string]string) string {
	switch condition.Kind {
	case "comparison":
		if condition.Comparison == nil {
			return "false"
		}
		return g.actionConditionComparisonExpression(action, entity, actionConditionFromComparison(*condition.Comparison), existing, input, values)
	case "and":
		if condition.Left == nil || condition.Right == nil {
			return "false"
		}
		return fmt.Sprintf("(%s && %s)", g.actionConditionTreeExpression(action, entity, *condition.Left, existing, input, values), g.actionConditionTreeExpression(action, entity, *condition.Right, existing, input, values))
	case "or":
		if condition.Left == nil || condition.Right == nil {
			return "false"
		}
		return fmt.Sprintf("(%s || %s)", g.actionConditionTreeExpression(action, entity, *condition.Left, existing, input, values), g.actionConditionTreeExpression(action, entity, *condition.Right, existing, input, values))
	case "not":
		if condition.Left == nil {
			return "false"
		}
		return fmt.Sprintf("!(%s)", g.actionConditionTreeExpression(action, entity, *condition.Left, existing, input, values))
	default:
		return "false"
	}
}

func (g *webGenerator) actionConditionGuards(action CustomActionDecl, condition ActionConditionDecl, indent string, existing string, input string, divisionMode string, guardCounter *int, values map[string]string) string {
	if condition.Tree != nil {
		return g.actionConditionTreeGuards(action, *condition.Tree, indent, existing, input, divisionMode, guardCounter, values)
	}
	return g.actionConditionComparisonGuards(action, condition, indent, existing, input, divisionMode, guardCounter, values)
}

func (g *webGenerator) actionConditionComparisonGuards(action CustomActionDecl, condition ActionConditionDecl, indent string, existing string, input string, divisionMode string, guardCounter *int, values map[string]string) string {
	var builder strings.Builder
	builder.WriteString(g.actionExpressionGuards(action, condition.Left, indent, existing, input, divisionMode, guardCounter, values))
	builder.WriteString(g.actionExpressionGuards(action, condition.Right, indent, existing, input, divisionMode, guardCounter, values))
	return builder.String()
}

func (g *webGenerator) actionConditionTreeGuards(action CustomActionDecl, condition ConditionExpressionDecl, indent string, existing string, input string, divisionMode string, guardCounter *int, values map[string]string) string {
	var builder strings.Builder
	switch condition.Kind {
	case "comparison":
		if condition.Comparison != nil {
			builder.WriteString(g.actionConditionComparisonGuards(action, actionConditionFromComparison(*condition.Comparison), indent, existing, input, divisionMode, guardCounter, values))
		}
	case "and", "or":
		if condition.Left != nil {
			builder.WriteString(g.actionConditionTreeGuards(action, *condition.Left, indent, existing, input, divisionMode, guardCounter, values))
		}
		if condition.Right != nil {
			builder.WriteString(g.actionConditionTreeGuards(action, *condition.Right, indent, existing, input, divisionMode, guardCounter, values))
		}
	case "not":
		if condition.Left != nil {
			builder.WriteString(g.actionConditionTreeGuards(action, *condition.Left, indent, existing, input, divisionMode, guardCounter, values))
		}
	}
	return builder.String()
}

func (g *webGenerator) actionExpressionGuards(action CustomActionDecl, expression ActionExpressionDecl, indent string, existing string, input string, divisionMode string, guardCounter *int, values map[string]string) string {
	var builder strings.Builder
	for _, divisor := range actionExpressionDivisionDivisors(expression) {
		*guardCounter = *guardCounter + 1
		guardName := fmt.Sprintf("%sLogicDivisor%d", lowerCamelCase(action.Name), *guardCounter)
		builder.WriteString(fmt.Sprintf("%sconst %s = %s;\n", indent, guardName, g.actionCoreExpressionWithValues(action, divisor, existing, input, true, values)))
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

func (g *webGenerator) actionExpression(action CustomActionDecl, expression ActionExpressionDecl, existing string, input string) string {
	return g.actionExpressionWithValues(action, expression, existing, input, nil)
}

func (g *webGenerator) actionExpressionWithValues(action CustomActionDecl, expression ActionExpressionDecl, existing string, input string, values map[string]string) string {
	if expression.Tree != nil {
		return g.actionCoreExpressionWithValues(action, *expression.Tree, existing, input, expression.Tree.Kind == "binary", values)
	}
	left := g.actionOperandExpressionWithValues(action, expression.Left, existing, input, values)
	if expression.Operator == "" || expression.Right == nil {
		return left
	}
	right := g.actionOperandExpressionWithValues(action, *expression.Right, existing, input, values)
	return fmt.Sprintf("Number(%s ?? 0) %s Number(%s ?? 0)", left, expression.Operator, right)
}

func (g *webGenerator) actionCoreExpression(action CustomActionDecl, expression ExpressionDecl, existing string, input string, numeric bool) string {
	return g.actionCoreExpressionWithValues(action, expression, existing, input, numeric, nil)
}

func (g *webGenerator) actionCoreExpressionWithValues(action CustomActionDecl, expression ExpressionDecl, existing string, input string, numeric bool, values map[string]string) string {
	return renderCoreExpressionJS(expression, func(node ExpressionDecl) string {
		return g.actionOperandExpressionWithValues(action, actionOperandFromExpression(node), existing, input, values)
	}, numeric)
}

func actionExpressionDivisionDivisors(expression ActionExpressionDecl) []ExpressionDecl {
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

func (g *webGenerator) actionOperandExpression(action CustomActionDecl, operand ActionOperandDecl, existing string, input string) string {
	return g.actionOperandExpressionWithValues(action, operand, existing, input, nil)
}

func (g *webGenerator) actionOperandExpressionWithValues(action CustomActionDecl, operand ActionOperandDecl, existing string, input string, values map[string]string) string {
	switch operand.Kind {
	case "string":
		return fmt.Sprintf("%q", operand.Value)
	case "number", "boolean":
		return operand.Value
	case "reference":
		if values != nil {
			if _, ok := values[operand.Value]; ok {
				return tsSafeIdentifier(operand.Value)
			}
		}
		for _, field := range action.Inputs {
			if field.Name == operand.Value {
				return input + "." + operand.Value
			}
		}
		return existing + "." + operand.Value
	default:
		return "undefined"
	}
}

func (g *webGenerator) actionExpressionRuntimeType(action CustomActionDecl, entity EntityDecl, expression ActionExpressionDecl, values map[string]string) string {
	if expression.Tree != nil {
		if expression.Tree.Kind == "binary" {
			return "number"
		}
		return g.actionOperandRuntimeType(action, entity, actionOperandFromExpression(*expression.Tree), values)
	}
	if expression.Operator != "" {
		return "number"
	}
	return g.actionOperandRuntimeType(action, entity, expression.Left, values)
}

func (g *webGenerator) actionOperandRuntimeType(action CustomActionDecl, entity EntityDecl, operand ActionOperandDecl, values map[string]string) string {
	switch operand.Kind {
	case "string":
		return "text"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "reference":
		if values != nil {
			if valueType, ok := values[operand.Value]; ok {
				return valueType
			}
		}
		for _, input := range action.Inputs {
			if input.Name == operand.Value {
				return input.Type
			}
		}
		if field, ok := fieldIndex(entity)[operand.Value]; ok {
			return field.Type
		}
	}
	return ""
}

func actionExpressionSingleOperand(expression ActionExpressionDecl) (ActionOperandDecl, bool) {
	if expression.Tree != nil {
		if expression.Tree.Kind == "binary" {
			return ActionOperandDecl{}, false
		}
		return actionOperandFromExpression(*expression.Tree), true
	}
	if expression.Operator == "" && expression.Right == nil {
		return expression.Left, true
	}
	return ActionOperandDecl{}, false
}

func (g *webGenerator) actionExpressionNeedsNumericNarrowing(action CustomActionDecl, entity EntityDecl, expression ActionExpressionDecl, values map[string]string) bool {
	if !numberLikeType(g.actionExpressionRuntimeType(action, entity, expression, values)) {
		return false
	}
	operand, ok := actionExpressionSingleOperand(expression)
	if !ok || operand.Kind != "reference" {
		return false
	}
	if values != nil {
		if _, local := values[operand.Value]; local {
			return false
		}
	}
	for _, input := range action.Inputs {
		if input.Name == operand.Value {
			return false
		}
	}
	return true
}

func (g *webGenerator) actionLocalValueExpression(action CustomActionDecl, entity EntityDecl, expression ActionExpressionDecl, existing string, input string, values map[string]string) string {
	rendered := g.actionExpressionWithValues(action, expression, existing, input, values)
	if g.actionExpressionNeedsNumericNarrowing(action, entity, expression, values) {
		return fmt.Sprintf("Number(%s ?? 0)", rendered)
	}
	return rendered
}

func (g *webGenerator) actionConditionSideExpression(action CustomActionDecl, entity EntityDecl, expression ActionExpressionDecl, existing string, input string, values map[string]string, ordered bool) string {
	rendered := g.actionExpressionWithValues(action, expression, existing, input, values)
	if ordered && g.actionExpressionNeedsNumericNarrowing(action, entity, expression, values) {
		return fmt.Sprintf("Number(%s ?? 0)", rendered)
	}
	return rendered
}

func actionConditionDivisionDivisors(condition ActionConditionDecl) []ExpressionDecl {
	if condition.Tree != nil {
		return conditionExpressionDivisionDivisors(*condition.Tree)
	}
	divisors := []ExpressionDecl{}
	divisors = append(divisors, actionExpressionDivisionDivisors(condition.Left)...)
	divisors = append(divisors, actionExpressionDivisionDivisors(condition.Right)...)
	return divisors
}

func conditionExpressionDivisionDivisors(condition ConditionExpressionDecl) []ExpressionDecl {
	divisors := []ExpressionDecl{}
	switch condition.Kind {
	case "comparison":
		if condition.Comparison != nil {
			divisors = append(divisors, expressionDivisionDivisors(condition.Comparison.Left)...)
			divisors = append(divisors, expressionDivisionDivisors(condition.Comparison.Right)...)
		}
	case "and", "or":
		if condition.Left != nil {
			divisors = append(divisors, conditionExpressionDivisionDivisors(*condition.Left)...)
		}
		if condition.Right != nil {
			divisors = append(divisors, conditionExpressionDivisionDivisors(*condition.Right)...)
		}
	case "not":
		if condition.Left != nil {
			divisors = append(divisors, conditionExpressionDivisionDivisors(*condition.Left)...)
		}
	}
	return divisors
}

func (g *webGenerator) customActionClientMethods(page PageDecl, entity EntityDecl) string {
	actions := customActionsForPage(g.program, page)
	if len(actions) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, action := range actions {
		builder.WriteString(fmt.Sprintf(",\n  run%s: (id: string, input: %sInput) =>\n    request<{ item: %s; message: string }>(endpoint + \"/\" + id + \"/actions/%s\", {\n      method: \"POST\",\n      body: JSON.stringify(input)\n    })", action.Name, action.Name, entity.Name, strings.ToLower(action.Name)))
	}
	return builder.String()
}

func (g *webGenerator) apiClientTypeImportList(page PageDecl, entity EntityDecl) string {
	names := []string{entity.Name}
	for _, action := range customActionsForPage(g.program, page) {
		names = append(names, action.Name+"Input")
	}
	return strings.Join(names, ", ")
}

func (g *webGenerator) customActionPermissionMetadata(page PageDecl) string {
	actions := customActionsForPage(g.program, page)
	if len(actions) == 0 {
		return "[]"
	}
	parts := []string{}
	for _, action := range actions {
		parts = append(parts, fmt.Sprintf("{ name: %q, allow: %s, writes: %s }", action.Name, tsStringArrayLiteral(action.Allow), tsStringArrayLiteral(customActionSetFieldNames(action))))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (g *webGenerator) pagePermissionsCall(page PageDecl, entity EntityDecl) string {
	base := fmt.Sprintf("pagePermissions(%q, %s", page.Source, tsStringArrayLiteral(fieldNames(entity)))
	if len(customActionsForPage(g.program, page)) > 0 {
		base += ", " + g.customActionPermissionMetadata(page)
	}
	return base + ")"
}

func (g *webGenerator) customActionPageState(page PageDecl) string {
	actions := customActionsForPage(g.program, page)
	if len(actions) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, action := range actions {
		state := lowerCamelCase(action.Name)
		builder.WriteString(fmt.Sprintf("  const [%sTarget, set%sTarget] = useState<%s | null>(null);\n", state, action.Name, page.Source))
		builder.WriteString(fmt.Sprintf("  const [%sForm, set%sForm] = useState<FormState>(%s);\n", state, action.Name, actionFormStateLiteral(action)))
		builder.WriteString(fmt.Sprintf("  const [%sTouchedFields, set%sTouchedFields] = useState<Record<string, boolean>>({});\n", state, action.Name))
		builder.WriteString(fmt.Sprintf("  const [%sSubmitted, set%sSubmitted] = useState(false);\n", state, action.Name))
	}
	builder.WriteString("\n")
	return builder.String()
}

func (g *webGenerator) customActionPageDerivedState(page PageDecl) string {
	actions := customActionsForPage(g.program, page)
	if len(actions) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, action := range actions {
		state := lowerCamelCase(action.Name)
		builder.WriteString(fmt.Sprintf("  const canRun%s = permissions.customActions.%s !== false;\n", action.Name, action.Name))
		builder.WriteString(fmt.Sprintf("  const %sErrors = useMemo(() => validate%sForm(%sForm), [%sForm]);\n", state, action.Name, state, state))
		builder.WriteString(fmt.Sprintf("  const visible%sErrors = useMemo(() => Object.fromEntries(Object.entries(%sErrors).filter(([field]) => %sTouchedFields[field] || %sSubmitted)), [%sErrors, %sTouchedFields, %sSubmitted]);\n", action.Name, state, state, state, state, state, state))
	}
	builder.WriteString("\n")
	return builder.String()
}

func (g *webGenerator) customActionFormValidationFunctions(page PageDecl) string {
	actions := customActionsForPage(g.program, page)
	if len(actions) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, action := range actions {
		builder.WriteString(fmt.Sprintf("function validate%sForm(form: FormState): FormErrors {\n", action.Name))
		builder.WriteString("  const errors: FormErrors = {};\n\n")
		for _, input := range action.Inputs {
			builder.WriteString(g.actionFormValidationFieldBlock(action, input))
		}
		builder.WriteString("  return errors;\n")
		builder.WriteString("}\n\n")
	}
	return builder.String()
}

func (g *webGenerator) actionFormValidationFieldBlock(action CustomActionDecl, field FieldDecl) string {
	var builder strings.Builder
	name := field.Name
	label := actionInputLabel(field)
	builder.WriteString(fmt.Sprintf("  if (form.%s.trim() === \"\") {\n", name))
	if defaultValue := modifierValue(field, "default"); defaultValue != "" {
		builder.WriteString("    // Empty input will use the action input default on submit.\n")
	} else if hasModifier(field, "required") {
		builder.WriteString(fmt.Sprintf("    errors.%s = %q;\n", name, fieldValidationMessage(field, label+" is required")))
	}
	builder.WriteString("  } else {\n")
	switch field.Type {
	case "number", "integer":
		builder.WriteString(fmt.Sprintf("    const parsed = Number(form.%s);\n", name))
		builder.WriteString(fmt.Sprintf("    if (!Number.isInteger(parsed)) errors.%s = %q;\n", name, fieldValidationMessage(field, label+" must be a whole number")))
		if minValue := modifierValue(field, "min"); minValue != "" {
			builder.WriteString(fmt.Sprintf("    else if (parsed < %s) errors.%s = %q;\n", minValue, name, fieldValidationMessage(field, label+" must be at least "+minValue)))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			builder.WriteString(fmt.Sprintf("    else if (parsed > %s) errors.%s = %q;\n", maxValue, name, fieldValidationMessage(field, label+" must be at most "+maxValue)))
		}
	case "decimal", "money":
		builder.WriteString(fmt.Sprintf("    const parsed = Number(form.%s);\n", name))
		builder.WriteString(fmt.Sprintf("    if (!Number.isFinite(parsed)) errors.%s = %q;\n", name, fieldValidationMessage(field, label+" must be a number")))
		if minValue := modifierValue(field, "min"); minValue != "" {
			builder.WriteString(fmt.Sprintf("    else if (parsed < %s) errors.%s = %q;\n", minValue, name, fieldValidationMessage(field, label+" must be at least "+minValue)))
		}
		if maxValue := modifierValue(field, "max"); maxValue != "" {
			builder.WriteString(fmt.Sprintf("    else if (parsed > %s) errors.%s = %q;\n", maxValue, name, fieldValidationMessage(field, label+" must be at most "+maxValue)))
		}
	case "email":
		builder.WriteString(fmt.Sprintf("    if (!form.%s.includes(\"@\")) errors.%s = %q;\n", name, name, fieldValidationMessage(field, label+" must be an email")))
		builder.WriteString(g.actionFormTextConstraints(field, name, label))
	default:
		builder.WriteString(g.actionFormTextConstraints(field, name, label))
	}
	builder.WriteString("  }\n\n")
	return builder.String()
}

func (g *webGenerator) actionFormTextConstraints(field FieldDecl, name string, label string) string {
	var builder strings.Builder
	if minLength, maxLength, ok := fieldLengthBounds(field); ok {
		builder.WriteString(fmt.Sprintf("    if (form.%s.length < %d || form.%s.length > %d) errors.%s = %q;\n", name, minLength, name, maxLength, name, fieldValidationMessage(field, fmt.Sprintf("%s length must be between %d and %d", label, minLength, maxLength))))
	}
	if hasModifier(field, "url") {
		builder.WriteString(fmt.Sprintf("    if (!errors.%s) {\n", name))
		builder.WriteString(fmt.Sprintf("      try { new URL(form.%s); } catch { errors.%s = %q; }\n", name, name, fieldValidationMessage(field, label+" must be a valid URL")))
		builder.WriteString("    }\n")
	}
	if pattern := modifierValue(field, "regex"); pattern != "" {
		builder.WriteString(fmt.Sprintf("    if (!errors.%s && !(new RegExp(%q)).test(form.%s)) errors.%s = %q;\n", name, pattern, name, name, fieldValidationMessage(field, label+" has an invalid format")))
	}
	return builder.String()
}

func (g *webGenerator) customActionPageFunctions(page PageDecl, entity EntityDecl) string {
	actions := customActionsForPage(g.program, page)
	if len(actions) == 0 {
		return ""
	}
	entityAPI := lowerCamelCase(entity.Name)
	var builder strings.Builder
	for _, action := range actions {
		state := lowerCamelCase(action.Name)
		builder.WriteString(fmt.Sprintf("  function open%sAction(item: %s) {\n", action.Name, entity.Name))
		builder.WriteString(fmt.Sprintf("    if (!canRun%s || item.archivedAt) return;\n", action.Name))
		builder.WriteString("    setNotice(null);\n")
		builder.WriteString(fmt.Sprintf("    set%sTarget(item);\n", action.Name))
		builder.WriteString(fmt.Sprintf("    set%sForm(%s);\n", action.Name, actionFormStateLiteral(action)))
		builder.WriteString(fmt.Sprintf("    set%sTouchedFields({});\n", action.Name))
		builder.WriteString(fmt.Sprintf("    set%sSubmitted(false);\n", action.Name))
		builder.WriteString("  }\n\n")
		builder.WriteString(fmt.Sprintf("  function close%sAction() {\n", action.Name))
		builder.WriteString(fmt.Sprintf("    set%sTarget(null);\n", action.Name))
		builder.WriteString(fmt.Sprintf("    set%sTouchedFields({});\n", action.Name))
		builder.WriteString(fmt.Sprintf("    set%sSubmitted(false);\n", action.Name))
		builder.WriteString("  }\n\n")
		builder.WriteString(fmt.Sprintf("  function update%sField(field: string, value: string) {\n", action.Name))
		builder.WriteString(fmt.Sprintf("    set%sForm((current) => ({ ...current, [field]: value }));\n", action.Name))
		builder.WriteString(fmt.Sprintf("    set%sTouchedFields((current) => ({ ...current, [field]: true }));\n", action.Name))
		builder.WriteString("  }\n\n")
		builder.WriteString(fmt.Sprintf("  async function run%sAction(event: FormEvent<HTMLFormElement>) {\n", action.Name))
		builder.WriteString("    event.preventDefault();\n")
		builder.WriteString(fmt.Sprintf("    if (!canRun%s || !%sTarget) return;\n", action.Name, state))
		builder.WriteString(fmt.Sprintf("    set%sSubmitted(true);\n", action.Name))
		builder.WriteString("    setNotice(null);\n")
		builder.WriteString("    setError(null);\n")
		builder.WriteString(fmt.Sprintf("    const nextErrors = validate%sForm(%sForm);\n", action.Name, state))
		builder.WriteString("    if (Object.keys(nextErrors).length > 0) return;\n\n")
		builder.WriteString("    setSaving(true);\n")
		builder.WriteString("    const input = {\n")
		for _, input := range action.Inputs {
			builder.WriteString(fmt.Sprintf("      %s: %s,\n", input.Name, actionFormValueExpression(input, state+"Form")))
		}
		builder.WriteString("    };\n\n")
		builder.WriteString("    try {\n")
		builder.WriteString(fmt.Sprintf("      const result = await %sApi.run%s(%sTarget.id, input);\n", entityAPI, action.Name, state))
		builder.WriteString(queryPageMutation(page, "      setItems((current) => current.map((existing) => existing.id === result.item.id ? result.item : existing));\n"))
		builder.WriteString("      setSelectedItem((current) => current?.id === result.item.id ? result.item : current);\n")
		builder.WriteString("      setNotice(result.message);\n")
		builder.WriteString(fmt.Sprintf("      close%sAction();\n", action.Name))
		builder.WriteString("    } catch (reason: unknown) {\n")
		builder.WriteString(fmt.Sprintf("      setError(reason instanceof Error ? reason.message : %q);\n", "Unable to run "+identifierLabel(action.Name)))
		builder.WriteString("    } finally {\n")
		builder.WriteString("      setSaving(false);\n")
		builder.WriteString("    }\n")
		builder.WriteString("  }\n\n")
	}
	return builder.String()
}

func (g *webGenerator) customActionPageButtons(page PageDecl, entity EntityDecl) string {
	actions := customActionsForPage(g.program, page)
	if len(actions) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, action := range actions {
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("                  {canRun%s && !item.archivedAt && <button%s type=\"button\" disabled={saving} onClick={() => open%sAction(item)}>{%s}</button>}\n", action.Name, g.actionRowButtonAttributes(page, action.Name, "secondary"), action.Name, g.uiLabelExpression("action."+action.Name, identifierLabel(action.Name), "locale")))
		} else {
			builder.WriteString(fmt.Sprintf("                  {canRun%s && !item.archivedAt && <button%s type=\"button\" disabled={saving} onClick={() => open%sAction(item)}>%s</button>}\n", action.Name, g.actionRowButtonAttributes(page, action.Name, "secondary"), action.Name, identifierLabel(action.Name)))
		}
	}
	return builder.String()
}

func (g *webGenerator) customActionPanels(page PageDecl, entity EntityDecl) string {
	actions := customActionsForPage(g.program, page)
	if len(actions) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, action := range actions {
		state := lowerCamelCase(action.Name)
		builder.WriteString(fmt.Sprintf("      {%sTarget && (\n", state))
		builder.WriteString("      <section className=\"panel bl-view-section-action\">\n")
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("        <h2>{%s}</h2>\n", g.uiLabelExpression("action."+action.Name, identifierLabel(action.Name), "locale")))
		} else {
			builder.WriteString(fmt.Sprintf("        <h2>%s</h2>\n", identifierLabel(action.Name)))
		}
		builder.WriteString(fmt.Sprintf("        <p className=\"muted\">{%sTarget.id}</p>\n", state))
		builder.WriteString(fmt.Sprintf("        <form noValidate onSubmit={run%sAction}>\n", action.Name))
		builder.WriteString("          <div className=\"form-grid\">\n")
		for _, input := range action.Inputs {
			builder.WriteString(fmt.Sprintf("            <label>\n              %s\n              <input%s%s value={%sForm.%s} onChange={(event) => update%sField(%q, event.target.value)} />\n              {visible%sErrors.%s && <span className=\"field-error\">{visible%sErrors.%s}</span>}\n              %s</label>\n", actionInputLabel(input), inputAttributes(input), placeholderAttribute(input), state, input.Name, action.Name, input.Name, action.Name, input.Name, action.Name, input.Name, helpElement(input)))
		}
		builder.WriteString("          </div>\n")
		builder.WriteString("          <div className=\"toolbar\">\n")
		if g.hasRuntimeI18N() {
			builder.WriteString(fmt.Sprintf("            <button%s type=\"submit\" disabled={saving}>{saving ? %s : %s}</button>\n", g.actionButtonAttributesWithSuffix(page, action.Name, "submit"), g.uiLabelExpression("action.running", "Running...", "locale"), g.uiLabelExpression("action.run", "Run", "locale")))
			builder.WriteString(fmt.Sprintf("            <button className=\"secondary\" type=\"button\" onClick={close%sAction}>{%s}</button>\n", action.Name, g.uiLabelExpression("action.cancel", "Cancel", "locale")))
		} else {
			builder.WriteString(fmt.Sprintf("            <button%s type=\"submit\" disabled={saving}>{saving ? \"Running...\" : \"Run\"}</button>\n", g.actionButtonAttributesWithSuffix(page, action.Name, "submit")))
			builder.WriteString(fmt.Sprintf("            <button className=\"secondary\" type=\"button\" onClick={close%sAction}>Cancel</button>\n", action.Name))
		}
		builder.WriteString("          </div>\n")
		builder.WriteString("        </form>\n")
		builder.WriteString("      </section>\n")
		builder.WriteString("      )}\n\n")
	}
	return builder.String()
}

func actionFormStateLiteral(action CustomActionDecl) string {
	if len(action.Inputs) == 0 {
		return "{}"
	}
	parts := []string{}
	for _, input := range action.Inputs {
		parts = append(parts, fmt.Sprintf("%s: %q", input.Name, modifierValue(input, "default")))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func actionInputLabel(field FieldDecl) string {
	return fieldLabel(field)
}

func actionFormValueExpression(field FieldDecl, formName string) string {
	switch field.Type {
	case "number", "integer", "decimal", "money":
		if defaultValue := modifierValue(field, "default"); defaultValue != "" {
			return fmt.Sprintf("%s.%s === \"\" ? %s : Number(%s.%s)", formName, field.Name, defaultValue, formName, field.Name)
		}
		if hasModifier(field, "required") {
			return fmt.Sprintf("Number(%s.%s)", formName, field.Name)
		}
		return fmt.Sprintf("%s.%s === \"\" ? undefined : Number(%s.%s)", formName, field.Name, formName, field.Name)
	case "boolean":
		return fmt.Sprintf("%s.%s === \"true\"", formName, field.Name)
	default:
		return fmt.Sprintf("%s.%s", formName, field.Name)
	}
}

func (g *webGenerator) openapiActionOperation(page PageDecl, entity EntityDecl, action CustomActionDecl) map[string]any {
	operation := map[string]any{
		"summary":                    "Run " + action.Name,
		"description":                "Runs a deterministic row-level BlackLang custom action and returns the updated record plus a message.",
		"x-blacklang-action":         action.Name,
		"x-blacklang-source":         action.Source,
		"x-blacklang-mutates-fields": customActionSetFieldNames(action),
		"parameters":                 []any{openapiIDParameter()},
		"requestBody":                openapiRequestBody("#/components/schemas/" + action.Name + "Input"),
		"responses":                  openapiJSONResponses(map[string]any{"$ref": "#/components/schemas/" + action.Name + "Response"}),
	}
	if g.program.Auth != nil {
		operation["security"] = []any{map[string]any{"cookieAuth": []any{}}}
	}
	if len(action.Allow) > 0 {
		operation["x-blacklang-allow"] = action.Allow
	}
	if transactionName, transactional := g.customActionTransactionName(action); transactional {
		operation["x-blacklang-transaction"] = true
		if transactionName != "" {
			operation["x-blacklang-transaction-name"] = transactionName
		}
	}
	if len(page.Access) > 0 {
		operation["x-blacklang-page-access"] = page.Access
	}
	g.addOpenAPIRelationLoad(operation, entity, "mutation")
	return operation
}

func openapiActionInputSchema(action CustomActionDecl) map[string]any {
	properties := map[string]any{}
	required := []string{}
	for _, input := range action.Inputs {
		properties[input.Name] = openapiFieldSchema(input, false)
		if hasModifier(input, "required") && modifierValue(input, "default") == "" {
			required = append(required, input.Name)
		}
	}
	return openapiObjectSchema(properties, required)
}
