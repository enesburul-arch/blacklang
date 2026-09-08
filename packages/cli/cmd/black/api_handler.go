package main

import (
	"fmt"
	"strings"
)

type APIUpdateDecl struct {
	Source         string             `json:"source"`
	Where          APIWhereDecl       `json:"where"`
	Sets           []APISetDecl       `json:"sets,omitempty"`
	Statements     []APIStatementDecl `json:"statements,omitempty"`
	Position       Position           `json:"position"`
	SourcePosition Position           `json:"sourcePosition,omitempty"`
}

type APIWhereDecl struct {
	Field         string         `json:"field"`
	Operator      string         `json:"operator"`
	Value         APIOperandDecl `json:"value"`
	Position      Position       `json:"position"`
	FieldPosition Position       `json:"fieldPosition,omitempty"`
	ValuePosition Position       `json:"valuePosition,omitempty"`
}

type APISetDecl struct {
	Field    string            `json:"field"`
	Expr     APIExpressionDecl `json:"expr"`
	Position Position          `json:"position"`
}

type APIStatementDecl struct {
	Kind     string        `json:"kind"`
	Value    *APIValueDecl `json:"value,omitempty"`
	Set      *APISetDecl   `json:"set,omitempty"`
	If       *APIIfDecl    `json:"if,omitempty"`
	Position Position      `json:"position"`
}

type APIValueDecl struct {
	Name       string            `json:"name"`
	Expression APIExpressionDecl `json:"expression"`
	Position   Position          `json:"position"`
}

type APIConditionDecl struct {
	Left     APIExpressionDecl        `json:"left"`
	Operator string                   `json:"operator"`
	Right    APIExpressionDecl        `json:"right"`
	Tree     *ConditionExpressionDecl `json:"tree,omitempty"`
	Position Position                 `json:"position"`
}

type APIIfDecl struct {
	Condition APIConditionDecl   `json:"condition"`
	Then      []APIStatementDecl `json:"then,omitempty"`
	Else      []APIStatementDecl `json:"else,omitempty"`
	Position  Position           `json:"position"`
}

type APIExpressionDecl struct {
	Left     APIOperandDecl  `json:"left"`
	Operator string          `json:"operator,omitempty"`
	Right    *APIOperandDecl `json:"right,omitempty"`
	Tree     *ExpressionDecl `json:"tree,omitempty"`
	Position Position        `json:"position"`
}

type APIOperandDecl struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

var supportedAPIBodyModifiers = setOf(
	"required",
	"optional",
	"min",
	"max",
	"length",
	"regex",
	"url",
	"message",
)

var supportedAPIRespondModes = setOf(
	"declared",
	"accepted",
	"updated",
)

func (p *parser) parseAPIBodyField(statement sourceStatement, line int) (FieldDecl, bool) {
	parts := statement.Parts()
	tokens := statement.Tokens
	if len(parts) < 3 || len(tokens) < 3 || !queryStatementIdentifiers(statement, 1, 2) {
		p.addError(line, 1, "INVALID_API_BODY", "API body field must be `body name type modifiers...`.", "Example: `body sku text required`.")
		return FieldDecl{}, false
	}
	return FieldDecl{
		Name:      parts[1],
		Type:      parts[2],
		Modifiers: parseModifiers(parts[3:]),
		Position:  p.position(line, 1),
	}, true
}

func (p *parser) parseAPIUpdate(statement sourceStatement, line int) (APIUpdateDecl, bool) {
	tokens := statement.Tokens
	if len(tokens) < 10 || !queryStatementIdentifiers(statement, 1, 3, 6, 7) {
		p.addError(line, 1, "INVALID_API_UPDATE", "API update must be `update Entity where field == value set field = value`.", "Example: `update Product where sku == body.sku set stock = body.stock`.")
		return APIUpdateDecl{}, false
	}
	if tokens[2].Value != "where" || tokens[4].Kind != tokenOperator || tokens[4].Value != "==" || tokens[6].Value != "set" {
		p.addError(line, 1, "INVALID_API_UPDATE", "API update must use `where field == value set field = value`.", "Example: `update Product where sku == body.sku set stock = body.stock`.")
		return APIUpdateDecl{}, false
	}
	whereValue, ok := parseAPIOperand(tokens[5])
	if !ok {
		p.addError(line, tokens[5].Position.Column, "INVALID_API_HANDLER_VALUE", "API handler values must be body.name, param.name, source fields, quoted strings, finite numbers, or true/false.", "Example: `body.sku`.")
		return APIUpdateDecl{}, false
	}
	update := APIUpdateDecl{
		Source: tokens[1].Value,
		Where: APIWhereDecl{
			Field:         tokens[3].Value,
			Operator:      tokens[4].Value,
			Value:         whereValue,
			Position:      tokens[3].Position,
			FieldPosition: tokens[3].Position,
			ValuePosition: tokens[5].Position,
		},
		Sets:           []APISetDecl{},
		Statements:     []APIStatementDecl{},
		Position:       tokens[0].Position,
		SourcePosition: tokens[1].Position,
	}
	index := 7
	for index < len(tokens) {
		if tokens[index].Kind != tokenIdentifier {
			p.addError(line, tokens[index].Position.Column, "INVALID_API_UPDATE", "API update set target must be a field name.", "Example: `set stock = body.stock`.")
			return APIUpdateDecl{}, false
		}
		fieldToken := tokens[index]
		if index+1 >= len(tokens) || tokens[index+1].Kind != tokenOperator || tokens[index+1].Value != "=" {
			p.addError(line, fieldToken.Position.Column, "INVALID_API_UPDATE", "API update assignments must use `field = value`.", "Example: `set stock = body.stock`.")
			return APIUpdateDecl{}, false
		}
		exprStart := index + 2
		exprEnd := exprStart
		for exprEnd < len(tokens) && !(tokens[exprEnd].Kind == tokenSymbol && tokens[exprEnd].Value == ",") {
			exprEnd++
		}
		expr, ok := p.parseAPIExpression(tokens[exprStart:exprEnd], line)
		if !ok {
			return APIUpdateDecl{}, false
		}
		update.Sets = append(update.Sets, APISetDecl{
			Field:    fieldToken.Value,
			Expr:     expr,
			Position: fieldToken.Position,
		})
		update.Statements = append(update.Statements, apiStatementFromSet(update.Sets[len(update.Sets)-1]))
		index = exprEnd
		if index >= len(tokens) {
			break
		}
		if tokens[index].Kind != tokenSymbol || tokens[index].Value != "," {
			p.addError(line, tokens[index].Position.Column, "INVALID_API_UPDATE", "API update assignments must be separated by commas.", "Example: `set stock = body.stock, price = body.price`.")
			return APIUpdateDecl{}, false
		}
		index++
		if index >= len(tokens) {
			p.addError(line, tokens[exprEnd].Position.Column, "INVALID_API_UPDATE", "API update cannot end with a trailing comma.", "Remove the trailing comma or add another assignment.")
			return APIUpdateDecl{}, false
		}
	}
	if len(update.Sets) == 0 {
		p.addError(line, 1, "INVALID_API_UPDATE", "API update must assign at least one field.", "Example: `update Product where sku == body.sku set stock = body.stock`.")
		return APIUpdateDecl{}, false
	}
	return update, true
}

func (p *parser) parseAPIUpdateBlock(start int, statement sourceStatement, line int) (APIUpdateDecl, int, bool) {
	tokens := statement.Tokens
	if len(tokens) != 7 || !queryStatementIdentifiers(statement, 1, 3) || tokens[2].Value != "where" || tokens[4].Kind != tokenOperator || tokens[4].Value != "==" || tokens[6].Kind != tokenSymbol || tokens[6].Value != "{" {
		p.addError(line, 1, "INVALID_API_UPDATE", "API update block must be `update Entity where field == value {`.", "Example: `update Product where sku == body.sku {`.")
		return APIUpdateDecl{}, start, false
	}
	whereValue, ok := parseAPIOperand(tokens[5])
	if !ok {
		p.addError(line, tokens[5].Position.Column, "INVALID_API_HANDLER_VALUE", "API handler values must be body.name, param.name, source fields, quoted strings, finite numbers, or true/false.", "Example: `body.sku`.")
		return APIUpdateDecl{}, start, false
	}
	update := APIUpdateDecl{
		Source: tokens[1].Value,
		Where: APIWhereDecl{
			Field:         tokens[3].Value,
			Operator:      tokens[4].Value,
			Value:         whereValue,
			Position:      tokens[3].Position,
			FieldPosition: tokens[3].Position,
			ValuePosition: tokens[5].Position,
		},
		Sets:           []APISetDecl{},
		Statements:     []APIStatementDecl{},
		Position:       tokens[0].Position,
		SourcePosition: tokens[1].Position,
	}
	statements, next := p.parseAPIUpdateBranchStatements(start+1, tokens[0].Position.Column)
	update.Statements = statements
	update.Sets = apiStatementSets(statements)
	if next >= len(p.lines) || !isClosingBrace(p.lines[next].Parts()) {
		p.addError(line, 1, "UNCLOSED_API_UPDATE", fmt.Sprintf("API update block for %s is missing a closing brace.", update.Source), "Add `}` after the update statements.")
		return update, next, false
	}
	return update, next, true
}

func apiStatementFromSet(set APISetDecl) APIStatementDecl {
	setCopy := set
	return APIStatementDecl{
		Kind:     "set",
		Set:      &setCopy,
		Position: set.Position,
	}
}

func apiStatementFromValue(value APIValueDecl) APIStatementDecl {
	valueCopy := value
	return APIStatementDecl{
		Kind:     "value",
		Value:    &valueCopy,
		Position: value.Position,
	}
}

func apiStatementFromIf(ifDecl APIIfDecl) APIStatementDecl {
	ifCopy := ifDecl
	return APIStatementDecl{
		Kind:     "if",
		If:       &ifCopy,
		Position: ifDecl.Position,
	}
}

func (p *parser) parseAPISetStatement(statement sourceStatement, line int) (APISetDecl, bool) {
	tokens := statement.Tokens
	if len(tokens) < 4 || tokens[1].Kind != tokenIdentifier || tokens[2].Kind != tokenOperator || tokens[2].Value != "=" {
		p.addError(line, 1, "INVALID_API_HANDLER_SET", "API update block set must be `set field = expression`.", "Example: `set stock = stock + incoming`.")
		return APISetDecl{}, false
	}
	expr, ok := p.parseAPIExpression(tokens[3:], line)
	if !ok {
		return APISetDecl{}, false
	}
	return APISetDecl{
		Field:    tokens[1].Value,
		Expr:     expr,
		Position: tokens[1].Position,
	}, true
}

func (p *parser) parseAPIValue(statement sourceStatement, line int) (APIValueDecl, bool) {
	tokens := statement.Tokens
	if len(tokens) < 4 || tokens[1].Kind != tokenIdentifier || tokens[2].Kind != tokenOperator || tokens[2].Value != "=" {
		p.addError(line, 1, "INVALID_API_HANDLER_VALUE_DECLARATION", "API value must be `value name = expression`.", "Example: `value incoming = body.quantity / body.packSize`.")
		return APIValueDecl{}, false
	}
	expression, ok := parseCoreExpression(tokens[3:], parseAPIExpressionOperand)
	if !ok {
		p.addError(line, tokens[3].Position.Column, "INVALID_API_HANDLER_VALUE_EXPRESSION", "API values must use body values, params, source fields, literals, +, -, *, /, and parentheses.", "Example: `value incoming = body.quantity / body.packSize`.")
		return APIValueDecl{}, false
	}
	return APIValueDecl{
		Name:       tokens[1].Value,
		Expression: apiExpressionFromCore(expression),
		Position:   tokens[1].Position,
	}, true
}

func (p *parser) parseAPIIf(start int) (APIStatementDecl, int, bool) {
	statement := p.lines[start]
	tokens := statement.Tokens
	line := p.lineNumber(start)
	if len(tokens) < 4 {
		p.addError(line, 1, "INVALID_API_HANDLER_IF", "API if must be `if condition`.", "Example: `if incoming > 0 and stock >= 0`.")
		return APIStatementDecl{}, start + 1, false
	}
	condition, ok := p.parseAPICondition(tokens[1:], line)
	if !ok {
		return APIStatementDecl{}, start + 1, false
	}
	parentIndent := tokens[0].Position.Column
	thenStatements, next := p.parseAPIUpdateBranchStatements(start+1, parentIndent)
	if len(thenStatements) == 0 {
		p.addError(line, 1, "MISSING_API_HANDLER_IF_BODY", "API if must contain at least one indented value, set, or nested if statement.", "Indent branch statements under the if line.")
		return APIStatementDecl{}, next, false
	}
	elseStatements := []APIStatementDecl{}
	if next < len(p.lines) && p.isAPIElseAt(next, parentIndent) {
		elseLine := p.lineNumber(next)
		elseTokens := p.lines[next].Tokens
		if len(elseTokens) != 1 {
			p.addError(elseLine, 1, "INVALID_API_HANDLER_ELSE", "API else must be written as `else` on its own line.", "Put branch statements on the following indented lines.")
			return APIStatementDecl{}, next + 1, false
		}
		elseStatements, next = p.parseAPIUpdateBranchStatements(next+1, parentIndent)
		if len(elseStatements) == 0 {
			p.addError(elseLine, 1, "MISSING_API_HANDLER_ELSE_BODY", "API else must contain at least one indented value, set, or nested if statement.", "Indent branch statements under the else line.")
			return APIStatementDecl{}, next, false
		}
	}
	return apiStatementFromIf(APIIfDecl{
		Condition: condition,
		Then:      thenStatements,
		Else:      elseStatements,
		Position:  tokens[0].Position,
	}), next, true
}

func (p *parser) parseAPIUpdateBranchStatements(start int, parentIndent int) ([]APIStatementDecl, int) {
	statements := []APIStatementDecl{}
	for index := start; index < len(p.lines); {
		statement := p.lines[index]
		tokens := statement.Tokens
		line := p.lineNumber(index)
		parts := statement.Parts()
		if len(tokens) == 0 {
			index++
			continue
		}
		if isClosingBrace(parts) || tokens[0].Position.Column <= parentIndent {
			return statements, index
		}
		if tokens[0].Kind != tokenIdentifier {
			p.addError(line, tokens[0].Position.Column, "UNEXPECTED_API_HANDLER_BRANCH_TOKEN", "API update branch statements must begin with value, set, or if.", "Use one branch statement per line.")
			index++
			continue
		}
		switch tokens[0].Value {
		case "value":
			value, ok := p.parseAPIValue(statement, line)
			if ok {
				statements = append(statements, apiStatementFromValue(value))
			}
			index++
		case "set":
			set, ok := p.parseAPISetStatement(statement, line)
			if ok {
				statements = append(statements, apiStatementFromSet(set))
			}
			index++
		case "if":
			statement, next, ok := p.parseAPIIf(index)
			if ok {
				statements = append(statements, statement)
			}
			index = next
		case "else":
			p.addError(line, tokens[0].Position.Column, "UNEXPECTED_API_HANDLER_ELSE", "API else must align with its matching if line.", "Move else to the same indentation as the if it belongs to.")
			index++
		default:
			p.addError(line, tokens[0].Position.Column, "UNEXPECTED_API_HANDLER_BRANCH_TOKEN", fmt.Sprintf("Unexpected API update branch token %q.", tokens[0].Value), "Use value, set, or if inside an API update block.")
			index++
		}
	}
	return statements, len(p.lines)
}

func (p *parser) isAPIElseAt(index int, parentIndent int) bool {
	if index >= len(p.lines) || len(p.lines[index].Tokens) == 0 {
		return false
	}
	token := p.lines[index].Tokens[0]
	return token.Kind == tokenIdentifier && token.Value == "else" && token.Position.Column == parentIndent
}

func (p *parser) parseAPICondition(tokens []sourceToken, line int) (APIConditionDecl, bool) {
	tree, ok := parseConditionExpression(tokens, parseAPIExpressionOperand)
	if !ok {
		p.addError(line, 1, "INVALID_API_HANDLER_CONDITION", "API if condition must be `expression operator expression` with optional and/or/not groups.", "Example: `if incoming > 0 and stock >= 0`.")
		return APIConditionDecl{}, false
	}
	comparison, ok := firstConditionComparison(tree)
	if !ok {
		p.addError(line, 1, "INVALID_API_HANDLER_CONDITION", "API if condition must include at least one comparison.", "Example: `if incoming > 0`.")
		return APIConditionDecl{}, false
	}
	result := apiConditionFromComparison(comparison)
	if tree.Kind != "comparison" {
		treeCopy := tree
		result.Tree = &treeCopy
	}
	return result, true
}

func apiConditionFromComparison(comparison ConditionComparisonDecl) APIConditionDecl {
	return APIConditionDecl{
		Left:     apiExpressionFromCore(comparison.Left),
		Operator: comparison.Operator,
		Right:    apiExpressionFromCore(comparison.Right),
		Position: comparison.Position,
	}
}

func (p *parser) parseAPIExpression(tokens []sourceToken, line int) (APIExpressionDecl, bool) {
	expression, ok := parseCoreExpression(tokens, parseAPIExpressionOperand)
	if !ok {
		column := 1
		if len(tokens) > 0 {
			column = tokens[0].Position.Column
		}
		p.addError(line, column, "INVALID_API_HANDLER_EXPRESSION", "API handler expressions must use body values, params, source fields, literals, +, -, *, /, and parentheses.", "Example: `stock + (body.quantity * body.packSize)`.")
		return APIExpressionDecl{}, false
	}
	return apiExpressionFromCore(expression), true
}

func parseAPIExpressionOperand(token sourceToken) (string, string, bool) {
	operand, ok := parseAPIOperand(token)
	if !ok {
		return "", "", false
	}
	return operand.Kind, operand.Value, true
}

func parseAPIOperand(token sourceToken) (APIOperandDecl, bool) {
	if token.Kind == tokenString {
		return APIOperandDecl{Kind: "string", Value: token.Value}, true
	}
	if token.Kind != tokenIdentifier {
		return APIOperandDecl{}, false
	}
	if token.Value == "true" || token.Value == "false" {
		return APIOperandDecl{Kind: "boolean", Value: token.Value}, true
	}
	if queryNumberPattern.MatchString(token.Value) {
		return APIOperandDecl{Kind: "number", Value: token.Value}, true
	}
	if kind, value, ok := parseAPIPrefixedOperand(token.Value); ok {
		return APIOperandDecl{Kind: kind, Value: value}, true
	}
	if queryIdentifierPattern.MatchString(token.Value) {
		return APIOperandDecl{Kind: "field", Value: token.Value}, true
	}
	return APIOperandDecl{}, false
}

func parseAPIPrefixedOperand(value string) (string, string, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return "", "", false
	}
	switch parts[0] {
	case "body", "param":
		if queryIdentifierPattern.MatchString(parts[1]) {
			return parts[0], parts[1], true
		}
	}
	return "", "", false
}

func apiExpressionFromCore(expression ExpressionDecl) APIExpressionDecl {
	result := APIExpressionDecl{
		Left:     apiOperandFromExpression(expression),
		Position: expression.Position,
	}
	if expression.Kind == "binary" {
		result.Operator = expression.Operator
		if expression.Left != nil {
			result.Left = apiOperandFromExpression(*expression.Left)
		}
		if expression.Right != nil {
			right := apiOperandFromExpression(*expression.Right)
			result.Right = &right
		}
	}
	expressionCopy := expression
	result.Tree = &expressionCopy
	return result
}

func apiOperandFromExpression(expression ExpressionDecl) APIOperandDecl {
	if expression.Kind == "operand" {
		return APIOperandDecl{Kind: expression.ValueKind, Value: expression.Value}
	}
	return APIOperandDecl{Kind: "expression", Value: formatCoreExpression(expression)}
}

func (v *semanticValidator) validateAPIBody(api APIDecl) map[string]FieldDecl {
	seen := map[string]FieldDecl{}
	if len(api.Body) > 0 && !explicitAPIHasBody(api.Method) {
		v.addDiagnostic(api.Position, "UNSUPPORTED_API_BODY_METHOD", fmt.Sprintf("API %s declares body fields for %s.", api.Name, strings.ToUpper(api.Method)), "Use body fields with POST, PUT, or PATCH APIs.")
	}
	for _, body := range api.Body {
		if existing, ok := seen[body.Name]; ok {
			v.addDiagnostic(body.Position, "DUPLICATE_API_BODY", fmt.Sprintf("API %s body field %s is already defined.", api.Name, body.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		seen[body.Name] = body
		if !queryIdentifierPattern.MatchString(body.Name) {
			v.addDiagnostic(body.Position, "INVALID_API_BODY", fmt.Sprintf("API %s body field %q is not a valid identifier.", api.Name, body.Name), "Use one body field name such as sku.")
		}
		if !supportedFieldTypes[body.Type] {
			v.addDiagnostic(body.Position, "UNSUPPORTED_API_BODY_TYPE", fmt.Sprintf("API %s body field %s uses unsupported type %q.", api.Name, body.Name, body.Type), "Use text, email, file, image, number, integer, decimal, money, boolean, date, or datetime.")
		}
		if hasModifier(body, "required") && hasModifier(body, "optional") {
			v.addDiagnostic(body.Position, "CONFLICTING_API_BODY_MODIFIER", fmt.Sprintf("API %s body field %s is both required and optional.", api.Name, body.Name), "Use either required or optional.")
		}
		for _, modifier := range body.Modifiers {
			if !supportedAPIBodyModifiers[modifier.Name] {
				v.addDiagnostic(body.Position, "UNSUPPORTED_API_BODY_MODIFIER", fmt.Sprintf("API %s body field %s uses unsupported modifier %q.", api.Name, body.Name, modifier.Name), "Use required, optional, min, max, length, regex, url, or message.")
				continue
			}
			v.validateConstraintModifier("api body", api.Name, body, modifier)
		}
	}
	return seen
}

func (v *semanticValidator) validateAPIRespond(api APIDecl) {
	if api.Respond == "" {
		return
	}
	if !supportedAPIRespondModes[api.Respond] {
		v.addDiagnostic(api.Position, "UNSUPPORTED_API_RESPOND", fmt.Sprintf("API %s uses unsupported respond mode %q.", api.Name, api.Respond), "Use declared, accepted, or updated.")
		return
	}
	if api.Update == nil && api.Respond == "updated" {
		v.addDiagnostic(api.Position, "UNSUPPORTED_API_RESPOND", fmt.Sprintf("API %s responds updated without a handler update.", api.Name), "Use `respond updated` only with an `update` handler.")
	}
}

func (v *semanticValidator) validateAPIUpdate(api APIDecl, entities map[string]EntityDecl, params map[string]APIParamDecl, body map[string]FieldDecl) {
	if api.Update == nil {
		return
	}
	update := *api.Update
	method := strings.ToUpper(api.Method)
	if method != "POST" && method != "PUT" && method != "PATCH" {
		v.addDiagnostic(update.Position, "UNSUPPORTED_API_HANDLER_METHOD", fmt.Sprintf("API %s update handler cannot run on %s.", api.Name, method), "Use POST, PUT, or PATCH for update handlers.")
	}
	entity, ok := entities[update.Source]
	if update.Source == "" {
		v.addDiagnostic(update.Position, "MISSING_API_HANDLER_SOURCE", fmt.Sprintf("API %s update handler is missing a source entity.", api.Name), "Use `update Entity where field == value set field = value`.")
		return
	}
	if !queryIdentifierPattern.MatchString(update.Source) {
		v.addDiagnostic(update.SourcePosition, "INVALID_API_HANDLER_SOURCE", fmt.Sprintf("API %s update source %q is not a valid entity identifier.", api.Name, update.Source), "Use an existing entity name such as Product.")
		return
	}
	if !ok {
		v.addDiagnostic(update.SourcePosition, "UNKNOWN_API_HANDLER_SOURCE", fmt.Sprintf("API %s update uses unknown entity %s.", api.Name, update.Source), "Declare the entity or change the update source.")
		return
	}
	whereField, whereOK := v.validateAPIUpdateWhere(api, update, entity, params, body)
	_ = whereField
	v.validateAPIUpdatePolicyScope(api, update, entity, params, body)
	v.validateAPIUpdateSets(api, update, entity, params, body)
	if !whereOK {
		return
	}
}

func (v *semanticValidator) validateAPIUpdateWhere(api APIDecl, update APIUpdateDecl, entity EntityDecl, params map[string]APIParamDecl, body map[string]FieldDecl) (FieldDecl, bool) {
	if update.Where.Operator != "==" {
		v.addDiagnostic(update.Where.Position, "UNSUPPORTED_API_HANDLER_OPERATOR", fmt.Sprintf("API %s update where operator %q is not supported.", api.Name, update.Where.Operator), "Use == for bounded update lookup.")
		return FieldDecl{}, false
	}
	field, ok := v.apiUpdateField(api, entity, update.Where.Field, update.Where.FieldPosition, true)
	if !ok {
		return FieldDecl{}, false
	}
	if update.Where.Field != "id" && !hasModifier(field, "unique") {
		v.addDiagnostic(update.Where.FieldPosition, "UNBOUNDED_API_UPDATE", fmt.Sprintf("API %s update where field %s.%s is not unique.", api.Name, entity.Name, update.Where.Field), "Use id or a stored field marked unique so the handler updates one bounded record.")
	}
	valueType, valueOK := v.validateAPIOperand(api, update.Where.Value, entity, params, body, nil, update.Where.ValuePosition, false)
	if valueOK && !actionValueAssignable(valueType, field.Type) {
		v.addDiagnostic(update.Where.ValuePosition, "API_HANDLER_VALUE_TYPE_MISMATCH", fmt.Sprintf("API %s update compares %s value to %s field %s.%s.", api.Name, valueType, field.Type, entity.Name, field.Name), "Use a request value or literal with the same stored field type family.")
	}
	return field, true
}

func (v *semanticValidator) validateAPIUpdatePolicyScope(api APIDecl, update APIUpdateDecl, entity EntityDecl, params map[string]APIParamDecl, body map[string]FieldDecl) {
	if len(entity.Policies) == 0 {
		return
	}
	if explicitAPIAccess(api) == "private" && v.program.Auth != nil {
		return
	}
	for _, policy := range entity.Policies {
		switch policy.Kind {
		case "tenant":
			if _, ok := body[policy.Field]; ok {
				if !hasModifier(body[policy.Field], "required") {
					v.addDiagnostic(update.Position, "MISSING_API_HANDLER_POLICY_SCOPE", fmt.Sprintf("API %s public handler body field %s must be required for tenant policy scope.", api.Name, policy.Field), "Mark the policy body field required.")
				}
				continue
			}
			if _, ok := params[policy.Field]; ok {
				continue
			}
			v.addDiagnostic(update.Position, "MISSING_API_HANDLER_POLICY_SCOPE", fmt.Sprintf("API %s public handler updates tenant-scoped entity %s without request field %s.", api.Name, entity.Name, policy.Field), fmt.Sprintf("Add `body %s text required` or a matching path param so the generated route scopes the update.", policy.Field))
		case "owner":
			v.addDiagnostic(update.Position, "UNSUPPORTED_API_HANDLER_POLICY_SCOPE", fmt.Sprintf("API %s public handler cannot update owner-scoped entity %s.", api.Name, entity.Name), "Use a private authenticated API handler for owner-scoped entities.")
		}
	}
}

func (v *semanticValidator) validateAPIUpdateSets(api APIDecl, update APIUpdateDecl, entity EntityDecl, params map[string]APIParamDecl, body map[string]FieldDecl) {
	statements := apiUpdateStatements(update)
	if !apiStatementsHaveSet(statements) {
		v.addDiagnostic(update.Position, "MISSING_API_HANDLER_SET", fmt.Sprintf("API %s update handler does not set any fields.", api.Name), "Add at least one `set field = value` assignment.")
		return
	}
	values := map[string]FieldDecl{}
	v.validateAPIUpdateStatements(api, update, entity, statements, params, body, values)
}

func (v *semanticValidator) validateAPIUpdateStatements(api APIDecl, update APIUpdateDecl, entity EntityDecl, statements []APIStatementDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl) {
	seen := map[string]Position{}
	for _, statement := range statements {
		switch statement.Kind {
		case "value":
			if statement.Value != nil {
				v.validateAPIValue(api, update, entity, *statement.Value, params, body, values)
			}
		case "set":
			if statement.Set != nil {
				v.validateAPISetStatement(api, entity, *statement.Set, params, body, values, seen)
			}
		case "if":
			if statement.If == nil {
				v.addDiagnostic(statement.Position, "INVALID_API_HANDLER_IF", fmt.Sprintf("API %s has an invalid if statement.", api.Name), "Use `if condition` with an indented branch.")
				continue
			}
			v.validateAPICondition(api, entity, statement.If.Condition, params, body, values)
			thenValues := cloneFieldMap(values)
			elseValues := cloneFieldMap(values)
			v.validateAPIUpdateStatements(api, update, entity, statement.If.Then, params, body, thenValues)
			v.validateAPIUpdateStatements(api, update, entity, statement.If.Else, params, body, elseValues)
		default:
			v.addDiagnostic(statement.Position, "INVALID_API_HANDLER_STATEMENT", fmt.Sprintf("API %s has unsupported update statement kind %q.", api.Name, statement.Kind), "Use value, set, or if inside API update logic.")
		}
	}
}

func (v *semanticValidator) validateAPIValue(api APIDecl, update APIUpdateDecl, entity EntityDecl, value APIValueDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl) {
	if !queryIdentifierPattern.MatchString(value.Name) || reservedLogicValueNames[strings.ToLower(value.Name)] {
		v.addDiagnostic(value.Position, "INVALID_API_HANDLER_VALUE_NAME", fmt.Sprintf("API %s value name %q is not a safe identifier.", api.Name, value.Name), "Use a lowerCamelCase name such as incoming.")
		return
	}
	if existing, ok := values[value.Name]; ok {
		v.addDiagnostic(value.Position, "DUPLICATE_API_HANDLER_VALUE", fmt.Sprintf("API %s value %s is already defined.", api.Name, value.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
		return
	}
	if _, ok := params[value.Name]; ok {
		v.addDiagnostic(value.Position, "API_HANDLER_VALUE_NAME_COLLISION", fmt.Sprintf("API %s value %s conflicts with a path parameter.", api.Name, value.Name), "Choose a value name distinct from params.")
		return
	}
	if _, ok := body[value.Name]; ok {
		v.addDiagnostic(value.Position, "API_HANDLER_VALUE_NAME_COLLISION", fmt.Sprintf("API %s value %s conflicts with a body field.", api.Name, value.Name), "Choose a value name distinct from body fields.")
		return
	}
	if _, ok := fieldIndex(entity)[value.Name]; ok {
		v.addDiagnostic(value.Position, "API_HANDLER_VALUE_NAME_COLLISION", fmt.Sprintf("API %s value %s conflicts with source field %s.%s.", api.Name, value.Name, entity.Name, value.Name), "Choose a value name distinct from source fields.")
		return
	}
	if _, ok := computedFieldIndex(entity)[value.Name]; ok {
		v.addDiagnostic(value.Position, "API_HANDLER_VALUE_NAME_COLLISION", fmt.Sprintf("API %s value %s conflicts with computed display field %s.%s.", api.Name, value.Name, entity.Name, value.Name), "Choose a value name distinct from computed fields.")
		return
	}
	valueType, ok, _ := v.inferAPIExpression(api, entity, value.Expression, params, body, values)
	if ok {
		values[value.Name] = FieldDecl{Name: value.Name, Type: valueType, Position: value.Position}
	}
}

func (v *semanticValidator) validateAPISetStatement(api APIDecl, entity EntityDecl, set APISetDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl, seen map[string]Position) {
	if existing, ok := seen[set.Field]; ok {
		v.addDiagnostic(set.Position, "DUPLICATE_API_HANDLER_SET", fmt.Sprintf("API %s update sets %s more than once in the same block.", api.Name, set.Field), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
		return
	}
	seen[set.Field] = set.Position
	field, ok := v.apiUpdateField(api, entity, set.Field, set.Position, false)
	if !ok {
		return
	}
	if policy, ok := entityPolicyFieldMap(entity)[set.Field]; ok {
		v.addDiagnostic(set.Position, "UNSUPPORTED_API_HANDLER_POLICY_FIELD", fmt.Sprintf("API %s update cannot set %s policy field %s.%s.", api.Name, policy.Kind, entity.Name, set.Field), "Policy fields are used for generated row scope and are not mutable from handlers.")
		return
	}
	v.validateAPIExpression(api, entity, field, set.Expr, params, body, values)
}

func (v *semanticValidator) apiUpdateField(api APIDecl, entity EntityDecl, name string, position Position, allowID bool) (FieldDecl, bool) {
	if allowID && name == "id" {
		return FieldDecl{Name: "id", Type: "text", Position: position}, true
	}
	if !queryIdentifierPattern.MatchString(name) {
		v.addDiagnostic(position, "INVALID_API_HANDLER_FIELD", fmt.Sprintf("API %s update field %q is not a valid identifier.", api.Name, name), "Use one stored field name without property paths or expressions.")
		return FieldDecl{}, false
	}
	if _, computed := computedFieldIndex(entity)[name]; computed {
		v.addDiagnostic(position, "UNSUPPORTED_API_HANDLER_FIELD", fmt.Sprintf("API %s update cannot write computed display field %s.%s.", api.Name, entity.Name, name), "Use stored scalar fields; computed fields are display-only.")
		return FieldDecl{}, false
	}
	field, ok := fieldIndex(entity)[name]
	if !ok {
		v.addDiagnostic(position, "UNKNOWN_API_HANDLER_FIELD", fmt.Sprintf("API %s update references unknown stored field %s.%s.", api.Name, entity.Name, name), "Use a declared stored scalar field.")
		return FieldDecl{}, false
	}
	if _, relation := v.findEntity(field.Type); relation {
		v.addDiagnostic(position, "UNSUPPORTED_API_HANDLER_FIELD", fmt.Sprintf("API %s update cannot use relation field %s.%s in the MVP.", api.Name, entity.Name, name), "Use stored primitive fields for handler lookup and assignments.")
		return FieldDecl{}, false
	}
	if !supportedFieldTypes[field.Type] {
		v.addDiagnostic(position, "UNSUPPORTED_API_HANDLER_FIELD", fmt.Sprintf("API %s update cannot use field %s.%s with type %s.", api.Name, entity.Name, name, field.Type), "Use stored primitive fields for handler lookup and assignments.")
		return FieldDecl{}, false
	}
	return field, true
}

func (v *semanticValidator) validateAPIExpression(api APIDecl, entity EntityDecl, target FieldDecl, expression APIExpressionDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl) {
	if expression.Tree != nil {
		valueType, ok, binary := v.validateAPIExpressionNode(api, entity, target, *expression.Tree, params, body, values)
		if ok && !binary {
			operand := apiOperandFromExpression(*expression.Tree)
			if !actionOperandAssignable(apiOperandAsAction(operand), valueType, target.Type) {
				v.addDiagnostic(expression.Position, "API_HANDLER_VALUE_TYPE_MISMATCH", fmt.Sprintf("API %s assigns %s value to %s field %s.%s.", api.Name, valueType, target.Type, entity.Name, target.Name), "Use a request value, source field, local value, or literal with the same stored field type family.")
			}
		}
		return
	}
	leftType, leftOK := v.validateAPIOperand(api, expression.Left, entity, params, body, values, expression.Position, true)
	if expression.Operator == "" {
		if leftOK && !actionOperandAssignable(apiOperandAsAction(expression.Left), leftType, target.Type) {
			v.addDiagnostic(expression.Position, "API_HANDLER_VALUE_TYPE_MISMATCH", fmt.Sprintf("API %s assigns %s value to %s field %s.%s.", api.Name, leftType, target.Type, entity.Name, target.Name), "Use a request value, source field, local value, or literal with the same stored field type family.")
		}
		return
	}
	if expression.Right == nil {
		v.addDiagnostic(expression.Position, "INVALID_API_HANDLER_EXPRESSION", fmt.Sprintf("API %s expression is missing a right operand.", api.Name), "Use `field = numericExpression`.")
		return
	}
	if !supportedActionOperators[expression.Operator] {
		v.addDiagnostic(expression.Position, "UNSUPPORTED_API_HANDLER_OPERATOR", fmt.Sprintf("API %s uses unsupported operator %q.", api.Name, expression.Operator), "Use +, -, *, or /.")
	}
	rightType, rightOK := v.validateAPIOperand(api, *expression.Right, entity, params, body, values, expression.Position, true)
	if !numberLikeType(target.Type) || !numberLikeType(leftType) || !numberLikeType(rightType) {
		v.addDiagnostic(expression.Position, "INCOMPATIBLE_API_HANDLER_EXPRESSION", fmt.Sprintf("API %s uses arithmetic for non-numeric field %s.%s.", api.Name, entity.Name, target.Name), "Use arithmetic only with number, integer, decimal, or money values.")
		return
	}
	if leftOK && rightOK && expression.Operator == "/" && expression.Right.Kind == "number" && numericLiteralIsZero(expression.Right.Value) {
		v.addDiagnostic(expression.Position, "INVALID_API_HANDLER_EXPRESSION", fmt.Sprintf("API %s divides by zero.", api.Name), "Use a non-zero literal divisor.")
	}
}

func (v *semanticValidator) validateAPIExpressionNode(api APIDecl, entity EntityDecl, target FieldDecl, expression ExpressionDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl) (string, bool, bool) {
	if expression.Kind != "binary" {
		operand := apiOperandFromExpression(expression)
		if operand.Kind == "expression" {
			v.addDiagnostic(expression.Position, "INVALID_API_HANDLER_VALUE", fmt.Sprintf("API %s has invalid handler value %q.", api.Name, operand.Value), "Use body.name, param.name, a source field, local value, or a typed literal.")
			return "", false, false
		}
		valueType, ok := v.validateAPIOperand(api, operand, entity, params, body, values, expression.Position, true)
		return valueType, ok, false
	}

	operatorOK := supportedActionOperators[expression.Operator]
	if !operatorOK {
		v.addDiagnostic(expression.Position, "UNSUPPORTED_API_HANDLER_OPERATOR", fmt.Sprintf("API %s uses unsupported operator %q.", api.Name, expression.Operator), "Use +, -, *, or /.")
	}
	leftType, leftOK, _ := v.validateAPIExpressionNode(api, entity, target, dereferenceExpression(expression.Left), params, body, values)
	rightType, rightOK, _ := v.validateAPIExpressionNode(api, entity, target, dereferenceExpression(expression.Right), params, body, values)
	ok := operatorOK && leftOK && rightOK
	if leftOK && rightOK && (!numberLikeType(target.Type) || !numberLikeType(leftType) || !numberLikeType(rightType)) {
		v.addDiagnostic(expression.Position, "INCOMPATIBLE_API_HANDLER_EXPRESSION", fmt.Sprintf("API %s uses arithmetic for non-numeric field %s.%s.", api.Name, entity.Name, target.Name), "Use arithmetic only with number, integer, decimal, or money values.")
		ok = false
	}
	if expression.Operator == "/" && expression.Right != nil && expressionIsZeroNumber(*expression.Right) {
		v.addDiagnostic(expression.Position, "INVALID_API_HANDLER_EXPRESSION", fmt.Sprintf("API %s divides by zero.", api.Name), "Use a non-zero literal divisor.")
		ok = false
	}
	return "number", ok, true
}

func (v *semanticValidator) inferAPIExpression(api APIDecl, entity EntityDecl, expression APIExpressionDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl) (string, bool, bool) {
	if expression.Tree != nil {
		return v.inferAPIExpressionNode(api, entity, *expression.Tree, params, body, values)
	}
	leftType, leftOK := v.validateAPIOperand(api, expression.Left, entity, params, body, values, expression.Position, true)
	if expression.Operator == "" {
		return leftType, leftOK, false
	}
	if expression.Right == nil {
		v.addDiagnostic(expression.Position, "INVALID_API_HANDLER_EXPRESSION", fmt.Sprintf("API %s expression is missing a right operand.", api.Name), "Use a deterministic expression.")
		return "", false, true
	}
	if !supportedActionOperators[expression.Operator] {
		v.addDiagnostic(expression.Position, "UNSUPPORTED_API_HANDLER_OPERATOR", fmt.Sprintf("API %s uses unsupported operator %q.", api.Name, expression.Operator), "Use +, -, *, or /.")
		return "", false, true
	}
	rightType, rightOK := v.validateAPIOperand(api, *expression.Right, entity, params, body, values, expression.Position, true)
	if leftOK && rightOK && (!numberLikeType(leftType) || !numberLikeType(rightType)) {
		v.addDiagnostic(expression.Position, "INCOMPATIBLE_API_HANDLER_EXPRESSION", fmt.Sprintf("API %s uses arithmetic with non-numeric values.", api.Name), "Use arithmetic only with number, integer, decimal, or money values.")
		return "", false, true
	}
	if leftOK && rightOK && expression.Operator == "/" && expression.Right.Kind == "number" && numericLiteralIsZero(expression.Right.Value) {
		v.addDiagnostic(expression.Position, "INVALID_API_HANDLER_EXPRESSION", fmt.Sprintf("API %s divides by zero.", api.Name), "Use a non-zero literal divisor.")
		return "", false, true
	}
	return "number", leftOK && rightOK, true
}

func (v *semanticValidator) inferAPIExpressionNode(api APIDecl, entity EntityDecl, expression ExpressionDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl) (string, bool, bool) {
	if expression.Kind != "binary" {
		operand := apiOperandFromExpression(expression)
		if operand.Kind == "expression" {
			v.addDiagnostic(expression.Position, "INVALID_API_HANDLER_VALUE", fmt.Sprintf("API %s has invalid handler value %q.", api.Name, operand.Value), "Use body.name, param.name, a source field, local value, or a typed literal.")
			return "", false, false
		}
		valueType, ok := v.validateAPIOperand(api, operand, entity, params, body, values, expression.Position, true)
		return valueType, ok, false
	}
	operatorOK := supportedActionOperators[expression.Operator]
	if !operatorOK {
		v.addDiagnostic(expression.Position, "UNSUPPORTED_API_HANDLER_OPERATOR", fmt.Sprintf("API %s uses unsupported operator %q.", api.Name, expression.Operator), "Use +, -, *, or /.")
	}
	leftType, leftOK, _ := v.inferAPIExpressionNode(api, entity, dereferenceExpression(expression.Left), params, body, values)
	rightType, rightOK, _ := v.inferAPIExpressionNode(api, entity, dereferenceExpression(expression.Right), params, body, values)
	ok := operatorOK && leftOK && rightOK
	if leftOK && rightOK && (!numberLikeType(leftType) || !numberLikeType(rightType)) {
		v.addDiagnostic(expression.Position, "INCOMPATIBLE_API_HANDLER_EXPRESSION", fmt.Sprintf("API %s uses arithmetic with non-numeric values.", api.Name), "Use arithmetic only with number, integer, decimal, or money values.")
		ok = false
	}
	if expression.Operator == "/" && expression.Right != nil && expressionIsZeroNumber(*expression.Right) {
		v.addDiagnostic(expression.Position, "INVALID_API_HANDLER_EXPRESSION", fmt.Sprintf("API %s divides by zero.", api.Name), "Use a non-zero literal divisor.")
		ok = false
	}
	return "number", ok, true
}

func (v *semanticValidator) validateAPICondition(api APIDecl, entity EntityDecl, condition APIConditionDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl) {
	if condition.Tree != nil {
		v.validateAPIConditionTree(api, entity, *condition.Tree, params, body, values)
		return
	}
	v.validateAPIConditionComparison(api, entity, condition, params, body, values)
}

func (v *semanticValidator) validateAPIConditionTree(api APIDecl, entity EntityDecl, condition ConditionExpressionDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl) {
	switch condition.Kind {
	case "comparison":
		if condition.Comparison == nil {
			v.addDiagnostic(condition.Position, "INVALID_API_HANDLER_CONDITION", fmt.Sprintf("API %s has an invalid if comparison.", api.Name), "Use `if condition` with one or more comparisons.")
			return
		}
		v.validateAPIConditionComparison(api, entity, apiConditionFromComparison(*condition.Comparison), params, body, values)
	case "and", "or":
		if condition.Left == nil || condition.Right == nil {
			v.addDiagnostic(condition.Position, "INVALID_API_HANDLER_CONDITION", fmt.Sprintf("API %s has an incomplete %s condition.", api.Name, condition.Kind), "Use both sides of and/or.")
			return
		}
		v.validateAPIConditionTree(api, entity, *condition.Left, params, body, values)
		v.validateAPIConditionTree(api, entity, *condition.Right, params, body, values)
	case "not":
		if condition.Left == nil {
			v.addDiagnostic(condition.Position, "INVALID_API_HANDLER_CONDITION", fmt.Sprintf("API %s has an incomplete not condition.", api.Name), "Use `not expression operator expression`.")
			return
		}
		v.validateAPIConditionTree(api, entity, *condition.Left, params, body, values)
	default:
		v.addDiagnostic(condition.Position, "INVALID_API_HANDLER_CONDITION", fmt.Sprintf("API %s has an unsupported condition node %q.", api.Name, condition.Kind), "Use comparison conditions joined by and/or/not.")
	}
}

func (v *semanticValidator) validateAPIConditionComparison(api APIDecl, entity EntityDecl, condition APIConditionDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl) {
	leftType, leftOK, _ := v.inferAPIExpression(api, entity, condition.Left, params, body, values)
	rightType, rightOK, _ := v.inferAPIExpression(api, entity, condition.Right, params, body, values)
	if !supportedComparisonOperators[condition.Operator] {
		v.addDiagnostic(condition.Position, "UNSUPPORTED_API_HANDLER_CONDITION_OPERATOR", fmt.Sprintf("API %s uses unsupported if operator %q.", api.Name, condition.Operator), "Use ==, !=, <, <=, >, or >=.")
		return
	}
	if !leftOK || !rightOK {
		return
	}
	if condition.Operator == "==" || condition.Operator == "!=" {
		if !actionValueAssignable(leftType, rightType) && !actionValueAssignable(rightType, leftType) {
			v.addDiagnostic(condition.Position, "API_HANDLER_CONDITION_TYPE_MISMATCH", fmt.Sprintf("API %s compares %s to %s.", api.Name, leftType, rightType), "Compare values from the same scalar type family.")
		}
		return
	}
	if !numberLikeType(leftType) || !numberLikeType(rightType) {
		v.addDiagnostic(condition.Position, "INCOMPATIBLE_API_HANDLER_CONDITION", fmt.Sprintf("API %s uses ordered comparison with %s and %s.", api.Name, leftType, rightType), "Use <, <=, >, or >= only with numeric values in API handler conditions.")
	}
}

func (v *semanticValidator) validateAPIOperand(api APIDecl, operand APIOperandDecl, entity EntityDecl, params map[string]APIParamDecl, body map[string]FieldDecl, values map[string]FieldDecl, position Position, allowSourceField bool) (string, bool) {
	switch operand.Kind {
	case "string":
		return "text", true
	case "number":
		return "number", true
	case "boolean":
		return "boolean", true
	case "body":
		field, ok := body[operand.Value]
		if !ok {
			v.addDiagnostic(position, "UNKNOWN_API_HANDLER_VALUE", fmt.Sprintf("API %s references unknown body value body.%s.", api.Name, operand.Value), "Declare the value with `body name type required` before using it in a handler.")
			return "", false
		}
		if !hasModifier(field, "required") {
			v.addDiagnostic(position, "OPTIONAL_API_HANDLER_VALUE", fmt.Sprintf("API %s handler uses optional body value body.%s.", api.Name, operand.Value), "Mark handler body values required so generated updates do not write null by accident.")
		}
		return field.Type, true
	case "param":
		param, ok := params[operand.Value]
		if !ok {
			v.addDiagnostic(position, "UNKNOWN_API_HANDLER_VALUE", fmt.Sprintf("API %s references unknown path parameter param.%s.", api.Name, operand.Value), "Declare the value with `param name type` and include `{name}` in the path.")
			return "", false
		}
		return param.Type, true
	case "field":
		if value, ok := values[operand.Value]; ok {
			return value.Type, true
		}
		if !allowSourceField {
			v.addDiagnostic(position, "UNSUPPORTED_API_HANDLER_VALUE", fmt.Sprintf("API %s where value cannot be source field %s.%s.", api.Name, entity.Name, operand.Value), "Use body.name, param.name, or a typed literal for bounded lookup.")
			return "", false
		}
		field, ok := v.apiUpdateField(api, entity, operand.Value, position, false)
		if !ok {
			return "", false
		}
		return field.Type, true
	default:
		v.addDiagnostic(position, "INVALID_API_HANDLER_VALUE", fmt.Sprintf("API %s has invalid handler value %q.", api.Name, operand.Value), "Use body.name, param.name, a source field, local value, or a typed literal.")
		return "", false
	}
}

func apiOperandAsAction(operand APIOperandDecl) ActionOperandDecl {
	switch operand.Kind {
	case "string", "number", "boolean":
		return ActionOperandDecl{Kind: operand.Kind, Value: operand.Value}
	default:
		return ActionOperandDecl{Kind: "reference", Value: operand.Value}
	}
}

func apiUpdateStatements(update APIUpdateDecl) []APIStatementDecl {
	if len(update.Statements) > 0 {
		return update.Statements
	}
	statements := []APIStatementDecl{}
	for _, set := range update.Sets {
		statements = append(statements, apiStatementFromSet(set))
	}
	return statements
}

func apiStatementsHaveSet(statements []APIStatementDecl) bool {
	return len(apiStatementSets(statements)) > 0
}

func apiStatementSets(statements []APIStatementDecl) []APISetDecl {
	sets := []APISetDecl{}
	for _, statement := range statements {
		switch statement.Kind {
		case "set":
			if statement.Set != nil {
				sets = append(sets, *statement.Set)
			}
		case "if":
			if statement.If != nil {
				sets = append(sets, apiStatementSets(statement.If.Then)...)
				sets = append(sets, apiStatementSets(statement.If.Else)...)
			}
		}
	}
	return sets
}

func apiUpdateSetFieldNames(update APIUpdateDecl) []string {
	fields := []string{}
	for _, set := range apiStatementSets(apiUpdateStatements(update)) {
		if !containsString(fields, set.Field) {
			fields = append(fields, set.Field)
		}
	}
	return fields
}
