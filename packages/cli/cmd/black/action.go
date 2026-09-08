package main

import (
	"fmt"
	"strings"
)

type CustomActionDecl struct {
	Name                string                `json:"name"`
	Source              string                `json:"source,omitempty"`
	Inputs              []FieldDecl           `json:"inputs,omitempty"`
	Sets                []ActionSetDecl       `json:"sets,omitempty"`
	Statements          []ActionStatementDecl `json:"statements,omitempty"`
	Allow               []string              `json:"allow,omitempty"`
	Success             string                `json:"success,omitempty"`
	Transaction         bool                  `json:"transaction,omitempty"`
	Position            Position              `json:"position"`
	SourcePosition      Position              `json:"sourcePosition,omitempty"`
	SuccessPosition     Position              `json:"successPosition,omitempty"`
	TransactionPosition Position              `json:"transactionPosition,omitempty"`
}

type ActionSetDecl struct {
	Field      string               `json:"field"`
	Expression ActionExpressionDecl `json:"expression"`
	Position   Position             `json:"position"`
}

type ActionStatementDecl struct {
	Kind     string           `json:"kind"`
	Value    *ActionValueDecl `json:"value,omitempty"`
	Set      *ActionSetDecl   `json:"set,omitempty"`
	If       *ActionIfDecl    `json:"if,omitempty"`
	Position Position         `json:"position"`
}

type ActionValueDecl struct {
	Name       string               `json:"name"`
	Expression ActionExpressionDecl `json:"expression"`
	Position   Position             `json:"position"`
}

type ActionConditionDecl struct {
	Left     ActionExpressionDecl     `json:"left"`
	Operator string                   `json:"operator"`
	Right    ActionExpressionDecl     `json:"right"`
	Tree     *ConditionExpressionDecl `json:"tree,omitempty"`
	Position Position                 `json:"position"`
}

type ActionIfDecl struct {
	Condition ActionConditionDecl   `json:"condition"`
	Then      []ActionStatementDecl `json:"then,omitempty"`
	Else      []ActionStatementDecl `json:"else,omitempty"`
	Position  Position              `json:"position"`
}

type ActionExpressionDecl struct {
	Left     ActionOperandDecl  `json:"left"`
	Operator string             `json:"operator,omitempty"`
	Right    *ActionOperandDecl `json:"right,omitempty"`
	Tree     *ExpressionDecl    `json:"tree,omitempty"`
	Position Position           `json:"position"`
}

type ActionOperandDecl struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

var supportedActionOperators = setOf(
	"+",
	"-",
	"*",
	"/",
)

var supportedActionInputModifiers = setOf(
	"required",
	"optional",
	"default",
	"label",
	"placeholder",
	"help",
	"min",
	"max",
	"length",
	"regex",
	"url",
	"message",
)

var reservedLogicValueNames = setOf(
	"if",
	"else",
	"and",
	"or",
	"not",
	"value",
	"set",
	"source",
	"input",
	"allow",
	"success",
	"update",
	"where",
	"body",
	"param",
	"true",
	"false",
	"null",
	"undefined",
	"constructor",
	"prototype",
	"__proto__",
)

func (p *parser) parseCustomAction(start int, parts []string) int {
	line := p.lineNumber(start)
	statement := p.lines[start]
	if len(parts) != 3 || len(statement.Tokens) != 3 || !queryStatementIdentifiers(statement, 0, 1) || statement.Tokens[2].Kind != tokenSymbol || parts[2] != "{" {
		p.addError(line, 1, "INVALID_ACTION_DECLARATION", "Action declaration must be `action Name {`.", "Example: `action RestockProduct {`.")
		return start
	}

	action := CustomActionDecl{
		Name:       parts[1],
		Inputs:     []FieldDecl{},
		Sets:       []ActionSetDecl{},
		Statements: []ActionStatementDecl{},
		Position:   p.position(line, 1),
	}
	seen := map[string]bool{}

	for index := start + 1; index < len(p.lines); index++ {
		statement := p.lines[index]
		tokens := statement.Tokens
		line := p.lineNumber(index)
		parts := statement.Parts()
		if len(tokens) == 1 && tokens[0].Kind == tokenSymbol && tokens[0].Value == "}" {
			p.program.Actions = append(p.program.Actions, action)
			return index
		}
		if len(tokens) == 0 {
			continue
		}
		if tokens[0].Kind != tokenIdentifier {
			p.addError(line, 1, "UNEXPECTED_ACTION_TOKEN", "Action clauses must begin with source, input, value, set, if, allow, or success.", "Use one action clause per line.")
			continue
		}
		keyword := tokens[0].Value
		if keyword == "source" || keyword == "success" {
			if seen[keyword] {
				p.addError(line, 1, "DUPLICATE_ACTION_"+strings.ToUpper(keyword), fmt.Sprintf("Action %s already declares %s.", action.Name, keyword), "Keep one "+keyword+" clause inside each action.")
				continue
			}
			seen[keyword] = true
		}

		switch keyword {
		case "source":
			if len(tokens) != 2 || !queryStatementIdentifiers(statement, 1) {
				p.addError(line, 1, "INVALID_ACTION_SOURCE", "Action source must be `source EntityName`.", "Example: `source Product`.")
				continue
			}
			action.Source = tokens[1].Value
			action.SourcePosition = tokens[1].Position
		case "transaction":
			p.addError(line, 1, "UNSUPPORTED_ACTION_TRANSACTION", "Action-level transaction is no longer canonical BlackLang syntax.", "Use a top-level transaction block such as `transaction RestockAtomic { action RestockProduct }`.")
		case "input":
			if len(parts) < 3 || len(tokens) < 3 || !queryStatementIdentifiers(statement, 1, 2) {
				p.addError(line, 1, "INVALID_ACTION_INPUT", "Action input must be `input name type modifiers...`.", "Example: `input quantity number required min 1`.")
				continue
			}
			action.Inputs = append(action.Inputs, FieldDecl{
				Name:      parts[1],
				Type:      parts[2],
				Modifiers: parseModifiers(parts[3:]),
				Position:  p.position(line, 1),
			})
		case "set":
			set, ok := p.parseActionSet(statement, line)
			if ok {
				action.Sets = append(action.Sets, set)
				action.Statements = append(action.Statements, actionStatementFromSet(set))
			}
		case "value":
			value, ok := p.parseActionValue(statement, line)
			if ok {
				action.Statements = append(action.Statements, actionStatementFromValue(value))
			}
		case "if":
			statement, next, ok := p.parseActionIf(index)
			if ok {
				action.Statements = append(action.Statements, statement)
				index = next - 1
			}
		case "else":
			p.addError(line, 1, "UNEXPECTED_ACTION_ELSE", "Action else must follow an if block at the same indentation.", "Place `else` directly after the indented if branch.")
		case "allow":
			allowed := parseList(parts[1:])
			if len(allowed) == 0 {
				p.addError(line, 1, "INVALID_ACTION_ALLOW", "Action allow must list at least one role or authenticated.", "Example: `allow Admin, Worker`.")
				continue
			}
			action.Allow = append(action.Allow, allowed...)
		case "success":
			if len(tokens) != 2 || tokens[1].Kind != tokenString {
				p.addError(line, 1, "INVALID_ACTION_SUCCESS", "Action success message must be `success \"Message\"`.", "Example: `success \"Stock updated\"`.")
				continue
			}
			action.Success = tokens[1].Value
			action.SuccessPosition = tokens[1].Position
		default:
			p.addError(line, 1, "UNEXPECTED_ACTION_TOKEN", fmt.Sprintf("Unexpected action token %q.", keyword), "Use source, input, value, set, if, allow, or success inside an action.")
		}
	}

	p.addError(line, 1, "UNCLOSED_ACTION", fmt.Sprintf("Action %s is missing a closing brace.", action.Name), "Add `}` after the action body.")
	return len(p.lines) - 1
}

func (p *parser) parseActionSet(statement sourceStatement, line int) (ActionSetDecl, bool) {
	tokens := statement.Tokens
	if len(tokens) < 4 {
		p.addError(line, 1, "INVALID_ACTION_SET", "Action set must be `set field = expression`.", "Example: `set stock = stock + (quantity * packSize)`.")
		return ActionSetDecl{}, false
	}
	if tokens[1].Kind != tokenIdentifier || tokens[2].Kind != tokenOperator || tokens[2].Value != "=" {
		p.addError(line, 1, "INVALID_ACTION_SET", "Action set must assign to one stored field with `=`.", "Example: `set stock = stock + quantity`.")
		return ActionSetDecl{}, false
	}
	expression, ok := parseCoreExpression(tokens[3:], parseActionExpressionOperand)
	if !ok {
		p.addError(line, tokens[3].Position.Column, "INVALID_ACTION_EXPRESSION", "Action set expressions must use inputs, source fields, literals, +, -, *, /, and parentheses.", "Example: `set stock = stock + (quantity * packSize)`.")
		return ActionSetDecl{}, false
	}
	return ActionSetDecl{
		Field:      tokens[1].Value,
		Expression: actionExpressionFromCore(expression),
		Position:   tokens[1].Position,
	}, true
}

func actionStatementFromSet(set ActionSetDecl) ActionStatementDecl {
	setCopy := set
	return ActionStatementDecl{
		Kind:     "set",
		Set:      &setCopy,
		Position: set.Position,
	}
}

func actionStatementFromValue(value ActionValueDecl) ActionStatementDecl {
	valueCopy := value
	return ActionStatementDecl{
		Kind:     "value",
		Value:    &valueCopy,
		Position: value.Position,
	}
}

func actionStatementFromIf(ifDecl ActionIfDecl) ActionStatementDecl {
	ifCopy := ifDecl
	return ActionStatementDecl{
		Kind:     "if",
		If:       &ifCopy,
		Position: ifDecl.Position,
	}
}

func (p *parser) parseActionValue(statement sourceStatement, line int) (ActionValueDecl, bool) {
	tokens := statement.Tokens
	if len(tokens) < 4 || tokens[1].Kind != tokenIdentifier || tokens[2].Kind != tokenOperator || tokens[2].Value != "=" {
		p.addError(line, 1, "INVALID_ACTION_VALUE", "Action value must be `value name = expression`.", "Example: `value subtotal = price * quantity`.")
		return ActionValueDecl{}, false
	}
	expression, ok := parseCoreExpression(tokens[3:], parseActionExpressionOperand)
	if !ok {
		p.addError(line, tokens[3].Position.Column, "INVALID_ACTION_VALUE_EXPRESSION", "Action values must use inputs, source fields, literals, +, -, *, /, and parentheses.", "Example: `value subtotal = price * quantity`.")
		return ActionValueDecl{}, false
	}
	return ActionValueDecl{
		Name:       tokens[1].Value,
		Expression: actionExpressionFromCore(expression),
		Position:   tokens[1].Position,
	}, true
}

func (p *parser) parseActionIf(start int) (ActionStatementDecl, int, bool) {
	statement := p.lines[start]
	tokens := statement.Tokens
	line := p.lineNumber(start)
	if len(tokens) < 4 {
		p.addError(line, 1, "INVALID_ACTION_IF", "Action if must be `if condition`.", "Example: `if subtotal > 1000 and stock >= 0`.")
		return ActionStatementDecl{}, start + 1, false
	}
	condition, ok := p.parseActionCondition(tokens[1:], line)
	if !ok {
		return ActionStatementDecl{}, start + 1, false
	}
	parentIndent := tokens[0].Position.Column
	thenStatements, next := p.parseActionBranchStatements(start+1, parentIndent)
	if len(thenStatements) == 0 {
		p.addError(line, 1, "MISSING_ACTION_IF_BODY", "Action if must contain at least one indented value, set, or nested if statement.", "Indent branch statements under the if line.")
		return ActionStatementDecl{}, next, false
	}
	elseStatements := []ActionStatementDecl{}
	if next < len(p.lines) && p.isActionElseAt(next, parentIndent) {
		elseLine := p.lineNumber(next)
		elseTokens := p.lines[next].Tokens
		if len(elseTokens) != 1 {
			p.addError(elseLine, 1, "INVALID_ACTION_ELSE", "Action else must be written as `else` on its own line.", "Put branch statements on the following indented lines.")
			return ActionStatementDecl{}, next + 1, false
		}
		elseStatements, next = p.parseActionBranchStatements(next+1, parentIndent)
		if len(elseStatements) == 0 {
			p.addError(elseLine, 1, "MISSING_ACTION_ELSE_BODY", "Action else must contain at least one indented value, set, or nested if statement.", "Indent branch statements under the else line.")
			return ActionStatementDecl{}, next, false
		}
	}
	return actionStatementFromIf(ActionIfDecl{
		Condition: condition,
		Then:      thenStatements,
		Else:      elseStatements,
		Position:  tokens[0].Position,
	}), next, true
}

func (p *parser) parseActionBranchStatements(start int, parentIndent int) ([]ActionStatementDecl, int) {
	statements := []ActionStatementDecl{}
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
			p.addError(line, tokens[0].Position.Column, "UNEXPECTED_ACTION_BRANCH_TOKEN", "Action branch statements must begin with value, set, or if.", "Use one branch statement per line.")
			index++
			continue
		}
		switch tokens[0].Value {
		case "value":
			value, ok := p.parseActionValue(statement, line)
			if ok {
				statements = append(statements, actionStatementFromValue(value))
			}
			index++
		case "set":
			set, ok := p.parseActionSet(statement, line)
			if ok {
				statements = append(statements, actionStatementFromSet(set))
			}
			index++
		case "if":
			statement, next, ok := p.parseActionIf(index)
			if ok {
				statements = append(statements, statement)
			}
			index = next
		case "else":
			p.addError(line, tokens[0].Position.Column, "UNEXPECTED_ACTION_ELSE", "Action else must align with its matching if line.", "Move else to the same indentation as the if it belongs to.")
			index++
		default:
			p.addError(line, tokens[0].Position.Column, "UNEXPECTED_ACTION_BRANCH_TOKEN", fmt.Sprintf("Unexpected action branch token %q.", tokens[0].Value), "Use value, set, or if inside an action if branch.")
			index++
		}
	}
	return statements, len(p.lines)
}

func (p *parser) isActionElseAt(index int, parentIndent int) bool {
	if index >= len(p.lines) || len(p.lines[index].Tokens) == 0 {
		return false
	}
	token := p.lines[index].Tokens[0]
	return token.Kind == tokenIdentifier && token.Value == "else" && token.Position.Column == parentIndent
}

func (p *parser) parseActionCondition(tokens []sourceToken, line int) (ActionConditionDecl, bool) {
	tree, ok := parseConditionExpression(tokens, parseActionExpressionOperand)
	if !ok {
		p.addError(line, 1, "INVALID_ACTION_CONDITION", "Action if condition must be `expression operator expression` with optional and/or/not groups.", "Example: `if subtotal > 1000 and stock >= 0`.")
		return ActionConditionDecl{}, false
	}
	comparison, ok := firstConditionComparison(tree)
	if !ok {
		p.addError(line, 1, "INVALID_ACTION_CONDITION", "Action if condition must include at least one comparison.", "Example: `if subtotal > 1000`.")
		return ActionConditionDecl{}, false
	}
	result := actionConditionFromComparison(comparison)
	if tree.Kind != "comparison" {
		treeCopy := tree
		result.Tree = &treeCopy
	}
	return result, true
}

func actionConditionFromComparison(comparison ConditionComparisonDecl) ActionConditionDecl {
	return ActionConditionDecl{
		Left:     actionExpressionFromCore(comparison.Left),
		Operator: comparison.Operator,
		Right:    actionExpressionFromCore(comparison.Right),
		Position: comparison.Position,
	}
}

func parseActionExpressionOperand(token sourceToken) (string, string, bool) {
	operand, ok := parseActionOperand(token)
	if !ok {
		return "", "", false
	}
	return operand.Kind, operand.Value, true
}

func actionExpressionFromCore(expression ExpressionDecl) ActionExpressionDecl {
	result := ActionExpressionDecl{
		Left:     actionOperandFromExpression(expression),
		Position: expression.Position,
	}
	if expression.Kind == "binary" {
		result.Operator = expression.Operator
		if expression.Left != nil {
			result.Left = actionOperandFromExpression(*expression.Left)
		}
		if expression.Right != nil {
			right := actionOperandFromExpression(*expression.Right)
			result.Right = &right
		}
	}
	expressionCopy := expression
	result.Tree = &expressionCopy
	return result
}

func actionOperandFromExpression(expression ExpressionDecl) ActionOperandDecl {
	if expression.Kind == "operand" {
		return ActionOperandDecl{Kind: expression.ValueKind, Value: expression.Value}
	}
	return ActionOperandDecl{Kind: "expression", Value: formatCoreExpression(expression)}
}

func parseActionOperand(token sourceToken) (ActionOperandDecl, bool) {
	if token.Kind == tokenString {
		return ActionOperandDecl{Kind: "string", Value: token.Value}, true
	}
	if token.Kind != tokenIdentifier {
		return ActionOperandDecl{}, false
	}
	if token.Value == "true" || token.Value == "false" {
		return ActionOperandDecl{Kind: "boolean", Value: token.Value}, true
	}
	if queryNumberPattern.MatchString(token.Value) {
		return ActionOperandDecl{Kind: "number", Value: token.Value}, true
	}
	if queryIdentifierPattern.MatchString(token.Value) {
		return ActionOperandDecl{Kind: "reference", Value: token.Value}, true
	}
	return ActionOperandDecl{}, false
}

func (v *semanticValidator) validateCustomActions(entities map[string]EntityDecl, roles map[string]RoleDecl) map[string]CustomActionDecl {
	actions := map[string]CustomActionDecl{}
	normalizedNames := map[string]string{}
	symbols := customActionOtherSymbols(v.program)
	for _, action := range v.program.Actions {
		if existing, ok := actions[action.Name]; ok {
			v.addDiagnostic(action.Position, "DUPLICATE_ACTION", fmt.Sprintf("Action %s is already defined.", action.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		actions[action.Name] = action
		if !queryNamePattern.MatchString(action.Name) {
			v.addDiagnostic(action.Position, "INVALID_ACTION_NAME", fmt.Sprintf("Action name %q must use PascalCase letters and digits.", action.Name), "Use a name such as RestockProduct.")
		}
		normalized := strings.ToLower(action.Name)
		if existing, ok := normalizedNames[normalized]; ok {
			v.addDiagnostic(action.Position, "ACTION_NAME_COLLISION", fmt.Sprintf("Action %s conflicts with action %s after name normalization.", action.Name, existing), "Choose distinct action names, including after lowercasing.")
		}
		normalizedNames[normalized] = action.Name
		if symbols[action.Name] || symbols[normalized] {
			v.addDiagnostic(action.Position, "ACTION_NAME_COLLISION", fmt.Sprintf("Action %s conflicts with another application symbol.", action.Name), "Choose a unique action name for unambiguous inspect --affected output.")
		}
		if action.Source == "" {
			v.addDiagnostic(action.Position, "MISSING_ACTION_SOURCE", fmt.Sprintf("Action %s is missing a source entity.", action.Name), "Add `source EntityName` inside the action.")
			continue
		}
		if !queryIdentifierPattern.MatchString(action.Source) {
			v.addDiagnostic(action.SourcePosition, "INVALID_ACTION_SOURCE", "Action source must be a valid entity identifier.", "Use an existing entity name such as Product.")
			continue
		}
		entity, ok := entities[action.Source]
		if !ok {
			v.addDiagnostic(action.SourcePosition, "UNKNOWN_ACTION_SOURCE", fmt.Sprintf("Action %s uses unknown entity %s.", action.Name, action.Source), "Declare the entity or change the action source.")
			continue
		}
		v.validateActionInputs(action, entity)
		v.validateActionSets(action, entity)
		v.validateActionAllow(action, roles)
	}
	return actions
}

func (v *semanticValidator) validateActionInputs(action CustomActionDecl, entity EntityDecl) {
	seen := map[string]FieldDecl{}
	sourceFields := fieldIndex(entity)
	computedFields := computedFieldIndex(entity)
	for _, input := range action.Inputs {
		if existing, ok := seen[input.Name]; ok {
			v.addDiagnostic(input.Position, "DUPLICATE_ACTION_INPUT", fmt.Sprintf("Action %s input %s is already defined.", action.Name, input.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		seen[input.Name] = input
		if _, ok := sourceFields[input.Name]; ok {
			v.addDiagnostic(input.Position, "ACTION_INPUT_FIELD_COLLISION", fmt.Sprintf("Action %s input %s conflicts with source field %s.%s.", action.Name, input.Name, entity.Name, input.Name), "Choose an input name that cannot be confused with a source field.")
		}
		if _, ok := computedFields[input.Name]; ok {
			v.addDiagnostic(input.Position, "ACTION_INPUT_FIELD_COLLISION", fmt.Sprintf("Action %s input %s conflicts with computed display field %s.%s.", action.Name, input.Name, entity.Name, input.Name), "Choose an input name that cannot be confused with a source field.")
		}
		if !queryIdentifierPattern.MatchString(input.Name) {
			v.addDiagnostic(input.Position, "INVALID_ACTION_INPUT", fmt.Sprintf("Action %s input %q is not a valid identifier.", action.Name, input.Name), "Use one input name such as quantity.")
		}
		if !supportedFieldTypes[input.Type] {
			v.addDiagnostic(input.Position, "UNSUPPORTED_ACTION_INPUT_TYPE", fmt.Sprintf("Action %s input %s uses unsupported type %q.", action.Name, input.Name, input.Type), "Use text, email, number, integer, decimal, money, boolean, date, or datetime for action inputs.")
		} else if input.Type == "file" || input.Type == "image" {
			v.addDiagnostic(input.Position, "UNSUPPORTED_ACTION_INPUT_TYPE", fmt.Sprintf("Action %s input %s uses media type %q.", action.Name, input.Name, input.Type), "Use media fields on entities; generated custom action media inputs are not supported yet.")
		}
		for _, modifier := range input.Modifiers {
			if !supportedActionInputModifiers[modifier.Name] {
				v.addDiagnostic(input.Position, "UNSUPPORTED_ACTION_INPUT_MODIFIER", fmt.Sprintf("Action %s input %s uses unsupported modifier %q.", action.Name, input.Name, modifier.Name), "Use required, optional, default, label, placeholder, help, min, max, length, regex, url, or message.")
			}
			v.validateConstraintModifier("action input", action.Name, input, modifier)
			if modifier.Name == "default" && modifier.Value == "" {
				v.addDiagnostic(input.Position, "MISSING_DEFAULT_VALUE", fmt.Sprintf("Action %s input %s has default without a value.", action.Name, input.Name), "Write default followed by a value, such as `default 1`.")
			}
			if modifier.Name == "label" && modifier.Value == "" {
				v.addDiagnostic(input.Position, "MISSING_LABEL_VALUE", fmt.Sprintf("Action %s input %s has label without a value.", action.Name, input.Name), "Write label followed by text, such as `label \"Quantity\"`.")
			}
			if modifier.Name == "placeholder" && modifier.Value == "" {
				v.addDiagnostic(input.Position, "MISSING_PLACEHOLDER_VALUE", fmt.Sprintf("Action %s input %s has placeholder without a value.", action.Name, input.Name), "Write placeholder followed by text, such as `placeholder \"Enter quantity\"`.")
			}
			if modifier.Name == "help" && modifier.Value == "" {
				v.addDiagnostic(input.Position, "MISSING_HELP_VALUE", fmt.Sprintf("Action %s input %s has help without a value.", action.Name, input.Name), "Write help followed by text, such as `help \"How much stock to add\"`.")
			}
			if modifier.Name == "message" && modifier.Value == "" {
				v.addDiagnostic(input.Position, "MISSING_MESSAGE_VALUE", fmt.Sprintf("Action %s input %s has message without a value.", action.Name, input.Name), "Write message followed by text, such as `message \"Quantity must be positive\"`.")
			}
		}
	}
}

func (v *semanticValidator) validateActionSets(action CustomActionDecl, entity EntityDecl) {
	statements := actionStatements(action)
	if !actionStatementsHaveSet(statements) {
		v.addDiagnostic(action.Position, "MISSING_ACTION_SET", fmt.Sprintf("Action %s does not change any fields.", action.Name), "Add a deterministic assignment such as `set stock = stock + quantity`.")
		return
	}
	fields := fieldIndex(entity)
	policyFields := entityPolicyFieldMap(entity)
	inputs := fieldIndex(EntityDecl{Fields: action.Inputs})
	values := map[string]FieldDecl{}
	v.validateActionStatements(action, entity, statements, fields, policyFields, inputs, values)
}

func (v *semanticValidator) validateActionStatements(action CustomActionDecl, entity EntityDecl, statements []ActionStatementDecl, fields map[string]FieldDecl, policyFields map[string]EntityPolicyDecl, inputs map[string]FieldDecl, values map[string]FieldDecl) {
	seenSets := map[string]Position{}
	for _, statement := range statements {
		switch statement.Kind {
		case "value":
			if statement.Value != nil {
				v.validateActionValue(action, entity, *statement.Value, fields, inputs, values)
			}
		case "set":
			if statement.Set != nil {
				v.validateActionSetStatement(action, entity, *statement.Set, fields, policyFields, inputs, values, seenSets)
			}
		case "if":
			if statement.If == nil {
				v.addDiagnostic(statement.Position, "INVALID_ACTION_IF", fmt.Sprintf("Action %s has an invalid if statement.", action.Name), "Use `if condition` with an indented branch.")
				continue
			}
			v.validateActionCondition(action, statement.If.Condition, fields, inputs, values)
			thenValues := cloneFieldMap(values)
			elseValues := cloneFieldMap(values)
			v.validateActionStatements(action, entity, statement.If.Then, fields, policyFields, inputs, thenValues)
			v.validateActionStatements(action, entity, statement.If.Else, fields, policyFields, inputs, elseValues)
		default:
			v.addDiagnostic(statement.Position, "INVALID_ACTION_STATEMENT", fmt.Sprintf("Action %s has unsupported statement kind %q.", action.Name, statement.Kind), "Use value, set, or if inside action logic.")
		}
	}
}

func (v *semanticValidator) validateActionValue(action CustomActionDecl, entity EntityDecl, value ActionValueDecl, fields map[string]FieldDecl, inputs map[string]FieldDecl, values map[string]FieldDecl) {
	if !queryIdentifierPattern.MatchString(value.Name) || reservedLogicValueNames[strings.ToLower(value.Name)] {
		v.addDiagnostic(value.Position, "INVALID_ACTION_VALUE_NAME", fmt.Sprintf("Action %s value name %q is not a safe identifier.", action.Name, value.Name), "Use a lowerCamelCase name such as subtotal.")
		return
	}
	if existing, ok := values[value.Name]; ok {
		v.addDiagnostic(value.Position, "DUPLICATE_ACTION_VALUE", fmt.Sprintf("Action %s value %s is already defined.", action.Name, value.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
		return
	}
	if _, ok := inputs[value.Name]; ok {
		v.addDiagnostic(value.Position, "ACTION_VALUE_NAME_COLLISION", fmt.Sprintf("Action %s value %s conflicts with an input.", action.Name, value.Name), "Choose a value name distinct from action inputs.")
		return
	}
	if _, ok := fields[value.Name]; ok {
		v.addDiagnostic(value.Position, "ACTION_VALUE_NAME_COLLISION", fmt.Sprintf("Action %s value %s conflicts with source field %s.%s.", action.Name, value.Name, entity.Name, value.Name), "Choose a value name distinct from source fields.")
		return
	}
	if _, ok := computedFieldIndex(entity)[value.Name]; ok {
		v.addDiagnostic(value.Position, "ACTION_VALUE_NAME_COLLISION", fmt.Sprintf("Action %s value %s conflicts with computed display field %s.%s.", action.Name, value.Name, entity.Name, value.Name), "Choose a value name distinct from computed fields.")
		return
	}
	valueType, ok, _ := v.inferActionExpression(action, value.Expression, fields, inputs, values)
	if ok {
		values[value.Name] = FieldDecl{Name: value.Name, Type: valueType, Position: value.Position}
	}
}

func (v *semanticValidator) validateActionSetStatement(action CustomActionDecl, entity EntityDecl, set ActionSetDecl, fields map[string]FieldDecl, policyFields map[string]EntityPolicyDecl, inputs map[string]FieldDecl, values map[string]FieldDecl, seenSets map[string]Position) {
	if existing, ok := seenSets[set.Field]; ok {
		v.addDiagnostic(set.Position, "DUPLICATE_ACTION_SET", fmt.Sprintf("Action %s sets %s more than once in the same block.", action.Name, set.Field), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
		return
	}
	seenSets[set.Field] = set.Position
	field, ok := fields[set.Field]
	if !ok {
		if _, computed := computedFieldIndex(entity)[set.Field]; computed {
			v.addDiagnostic(set.Position, "UNSUPPORTED_ACTION_FIELD", fmt.Sprintf("Action %s cannot set computed display field %s.%s.", action.Name, entity.Name, set.Field), "Use stored scalar fields; computed fields are display-only.")
			return
		}
		v.addDiagnostic(set.Position, "UNKNOWN_ACTION_FIELD", fmt.Sprintf("Action %s sets unknown field %s.%s.", action.Name, entity.Name, set.Field), "Use a stored source field.")
		return
	}
	if policy, ok := policyFields[set.Field]; ok {
		v.addDiagnostic(set.Position, "UNSUPPORTED_ACTION_POLICY_FIELD", fmt.Sprintf("Action %s cannot set %s policy field %s.%s.", action.Name, policy.Kind, entity.Name, field.Name), "Policy fields are written by generated auth-aware routes.")
		return
	}
	if _, relation := v.findEntity(field.Type); relation {
		v.addDiagnostic(set.Position, "UNSUPPORTED_ACTION_FIELD", fmt.Sprintf("Action %s cannot set relation field %s.%s in the MVP.", action.Name, entity.Name, field.Name), "Use stored primitive fields for custom action assignments.")
		return
	}
	if !supportedFieldTypes[field.Type] {
		v.addDiagnostic(set.Position, "UNSUPPORTED_ACTION_FIELD", fmt.Sprintf("Action %s cannot set field %s.%s with type %s.", action.Name, entity.Name, field.Name, field.Type), "Use stored primitive fields for custom action assignments.")
		return
	}
	v.validateActionExpression(action, field, set.Expression, fields, inputs, values)
}

func (v *semanticValidator) validateActionExpression(action CustomActionDecl, target FieldDecl, expression ActionExpressionDecl, fields map[string]FieldDecl, inputs map[string]FieldDecl, values map[string]FieldDecl) {
	if expression.Tree != nil {
		valueType, ok, binary := v.validateActionExpressionNode(action, target, *expression.Tree, fields, inputs, values)
		if ok && !binary {
			operand := actionOperandFromExpression(*expression.Tree)
			if !actionOperandAssignable(operand, valueType, target.Type) {
				v.addDiagnostic(expression.Position, "ACTION_VALUE_TYPE_MISMATCH", fmt.Sprintf("Action %s assigns %s value to %s field %s.", action.Name, valueType, target.Type, target.Name), "Use a value with the same stored field type family.")
			}
		}
		return
	}
	leftType, leftOK := v.validateActionOperand(action, expression.Left, fields, inputs, values, expression.Position)
	if expression.Operator == "" {
		if leftOK && !actionOperandAssignable(expression.Left, leftType, target.Type) {
			v.addDiagnostic(expression.Position, "ACTION_VALUE_TYPE_MISMATCH", fmt.Sprintf("Action %s assigns %s value to %s field %s.", action.Name, leftType, target.Type, target.Name), "Use a value with the same stored field type family.")
		}
		return
	}
	if expression.Right == nil {
		v.addDiagnostic(expression.Position, "INVALID_ACTION_EXPRESSION", fmt.Sprintf("Action %s expression is missing a right operand.", action.Name), "Use `set field = numericExpression`.")
		return
	}
	if !supportedActionOperators[expression.Operator] {
		v.addDiagnostic(expression.Position, "UNSUPPORTED_ACTION_OPERATOR", fmt.Sprintf("Action %s uses unsupported operator %q.", action.Name, expression.Operator), "Use +, -, *, or /.")
	}
	rightType, rightOK := v.validateActionOperand(action, *expression.Right, fields, inputs, values, expression.Position)
	if !numberLikeType(target.Type) || !numberLikeType(leftType) || !numberLikeType(rightType) {
		v.addDiagnostic(expression.Position, "INCOMPATIBLE_ACTION_EXPRESSION", fmt.Sprintf("Action %s uses arithmetic for non-numeric field %s.", action.Name, target.Name), "Use arithmetic only with number, integer, decimal, or money values.")
		return
	}
	if leftOK && rightOK && expression.Operator == "/" && expression.Right.Kind == "number" && numericLiteralIsZero(expression.Right.Value) {
		v.addDiagnostic(expression.Position, "INVALID_ACTION_EXPRESSION", fmt.Sprintf("Action %s divides by zero.", action.Name), "Use a non-zero literal divisor.")
	}
}

func (v *semanticValidator) validateActionExpressionNode(action CustomActionDecl, target FieldDecl, expression ExpressionDecl, fields map[string]FieldDecl, inputs map[string]FieldDecl, values map[string]FieldDecl) (string, bool, bool) {
	if expression.Kind != "binary" {
		operand := actionOperandFromExpression(expression)
		if operand.Kind == "expression" {
			v.addDiagnostic(expression.Position, "INVALID_ACTION_VALUE", fmt.Sprintf("Action %s has invalid expression value %q.", action.Name, operand.Value), "Use an action input, source field, or typed literal.")
			return "", false, false
		}
		valueType, ok := v.validateActionOperand(action, operand, fields, inputs, values, expression.Position)
		return valueType, ok, false
	}

	operatorOK := supportedActionOperators[expression.Operator]
	if !operatorOK {
		v.addDiagnostic(expression.Position, "UNSUPPORTED_ACTION_OPERATOR", fmt.Sprintf("Action %s uses unsupported operator %q.", action.Name, expression.Operator), "Use +, -, *, or /.")
	}
	leftType, leftOK, _ := v.validateActionExpressionNode(action, target, dereferenceExpression(expression.Left), fields, inputs, values)
	rightType, rightOK, _ := v.validateActionExpressionNode(action, target, dereferenceExpression(expression.Right), fields, inputs, values)
	ok := operatorOK && leftOK && rightOK
	if leftOK && rightOK && (!numberLikeType(target.Type) || !numberLikeType(leftType) || !numberLikeType(rightType)) {
		v.addDiagnostic(expression.Position, "INCOMPATIBLE_ACTION_EXPRESSION", fmt.Sprintf("Action %s uses arithmetic for non-numeric field %s.", action.Name, target.Name), "Use arithmetic only with number, integer, decimal, or money values.")
		ok = false
	}
	if expression.Operator == "/" && expression.Right != nil && expressionIsZeroNumber(*expression.Right) {
		v.addDiagnostic(expression.Position, "INVALID_ACTION_EXPRESSION", fmt.Sprintf("Action %s divides by zero.", action.Name), "Use a non-zero literal divisor.")
		ok = false
	}
	return "number", ok, true
}

func (v *semanticValidator) inferActionExpression(action CustomActionDecl, expression ActionExpressionDecl, fields map[string]FieldDecl, inputs map[string]FieldDecl, values map[string]FieldDecl) (string, bool, bool) {
	if expression.Tree != nil {
		return v.inferActionExpressionNode(action, *expression.Tree, fields, inputs, values)
	}
	leftType, leftOK := v.validateActionOperand(action, expression.Left, fields, inputs, values, expression.Position)
	if expression.Operator == "" {
		return leftType, leftOK, false
	}
	if expression.Right == nil {
		v.addDiagnostic(expression.Position, "INVALID_ACTION_EXPRESSION", fmt.Sprintf("Action %s expression is missing a right operand.", action.Name), "Use a deterministic expression.")
		return "", false, true
	}
	if !supportedActionOperators[expression.Operator] {
		v.addDiagnostic(expression.Position, "UNSUPPORTED_ACTION_OPERATOR", fmt.Sprintf("Action %s uses unsupported operator %q.", action.Name, expression.Operator), "Use +, -, *, or /.")
		return "", false, true
	}
	rightType, rightOK := v.validateActionOperand(action, *expression.Right, fields, inputs, values, expression.Position)
	if leftOK && rightOK && (!numberLikeType(leftType) || !numberLikeType(rightType)) {
		v.addDiagnostic(expression.Position, "INCOMPATIBLE_ACTION_EXPRESSION", fmt.Sprintf("Action %s uses arithmetic with non-numeric values.", action.Name), "Use arithmetic only with number, integer, decimal, or money values.")
		return "", false, true
	}
	if leftOK && rightOK && expression.Operator == "/" && expression.Right.Kind == "number" && numericLiteralIsZero(expression.Right.Value) {
		v.addDiagnostic(expression.Position, "INVALID_ACTION_EXPRESSION", fmt.Sprintf("Action %s divides by zero.", action.Name), "Use a non-zero literal divisor.")
		return "", false, true
	}
	return "number", leftOK && rightOK, true
}

func (v *semanticValidator) inferActionExpressionNode(action CustomActionDecl, expression ExpressionDecl, fields map[string]FieldDecl, inputs map[string]FieldDecl, values map[string]FieldDecl) (string, bool, bool) {
	if expression.Kind != "binary" {
		operand := actionOperandFromExpression(expression)
		if operand.Kind == "expression" {
			v.addDiagnostic(expression.Position, "INVALID_ACTION_VALUE", fmt.Sprintf("Action %s has invalid expression value %q.", action.Name, operand.Value), "Use an action input, source field, value, or typed literal.")
			return "", false, false
		}
		valueType, ok := v.validateActionOperand(action, operand, fields, inputs, values, expression.Position)
		return valueType, ok, false
	}
	operatorOK := supportedActionOperators[expression.Operator]
	if !operatorOK {
		v.addDiagnostic(expression.Position, "UNSUPPORTED_ACTION_OPERATOR", fmt.Sprintf("Action %s uses unsupported operator %q.", action.Name, expression.Operator), "Use +, -, *, or /.")
	}
	leftType, leftOK, _ := v.inferActionExpressionNode(action, dereferenceExpression(expression.Left), fields, inputs, values)
	rightType, rightOK, _ := v.inferActionExpressionNode(action, dereferenceExpression(expression.Right), fields, inputs, values)
	ok := operatorOK && leftOK && rightOK
	if leftOK && rightOK && (!numberLikeType(leftType) || !numberLikeType(rightType)) {
		v.addDiagnostic(expression.Position, "INCOMPATIBLE_ACTION_EXPRESSION", fmt.Sprintf("Action %s uses arithmetic with non-numeric values.", action.Name), "Use arithmetic only with number, integer, decimal, or money values.")
		ok = false
	}
	if expression.Operator == "/" && expression.Right != nil && expressionIsZeroNumber(*expression.Right) {
		v.addDiagnostic(expression.Position, "INVALID_ACTION_EXPRESSION", fmt.Sprintf("Action %s divides by zero.", action.Name), "Use a non-zero literal divisor.")
		ok = false
	}
	return "number", ok, true
}

func (v *semanticValidator) validateActionCondition(action CustomActionDecl, condition ActionConditionDecl, fields map[string]FieldDecl, inputs map[string]FieldDecl, values map[string]FieldDecl) {
	if condition.Tree != nil {
		v.validateActionConditionTree(action, *condition.Tree, fields, inputs, values)
		return
	}
	v.validateActionConditionComparison(action, condition, fields, inputs, values)
}

func (v *semanticValidator) validateActionConditionTree(action CustomActionDecl, condition ConditionExpressionDecl, fields map[string]FieldDecl, inputs map[string]FieldDecl, values map[string]FieldDecl) {
	switch condition.Kind {
	case "comparison":
		if condition.Comparison == nil {
			v.addDiagnostic(condition.Position, "INVALID_ACTION_CONDITION", fmt.Sprintf("Action %s has an invalid if comparison.", action.Name), "Use `if condition` with one or more comparisons.")
			return
		}
		v.validateActionConditionComparison(action, actionConditionFromComparison(*condition.Comparison), fields, inputs, values)
	case "and", "or":
		if condition.Left == nil || condition.Right == nil {
			v.addDiagnostic(condition.Position, "INVALID_ACTION_CONDITION", fmt.Sprintf("Action %s has an incomplete %s condition.", action.Name, condition.Kind), "Use both sides of and/or.")
			return
		}
		v.validateActionConditionTree(action, *condition.Left, fields, inputs, values)
		v.validateActionConditionTree(action, *condition.Right, fields, inputs, values)
	case "not":
		if condition.Left == nil {
			v.addDiagnostic(condition.Position, "INVALID_ACTION_CONDITION", fmt.Sprintf("Action %s has an incomplete not condition.", action.Name), "Use `not expression operator expression`.")
			return
		}
		v.validateActionConditionTree(action, *condition.Left, fields, inputs, values)
	default:
		v.addDiagnostic(condition.Position, "INVALID_ACTION_CONDITION", fmt.Sprintf("Action %s has an unsupported condition node %q.", action.Name, condition.Kind), "Use comparison conditions joined by and/or/not.")
	}
}

func (v *semanticValidator) validateActionConditionComparison(action CustomActionDecl, condition ActionConditionDecl, fields map[string]FieldDecl, inputs map[string]FieldDecl, values map[string]FieldDecl) {
	leftType, leftOK, _ := v.inferActionExpression(action, condition.Left, fields, inputs, values)
	rightType, rightOK, _ := v.inferActionExpression(action, condition.Right, fields, inputs, values)
	if !supportedComparisonOperators[condition.Operator] {
		v.addDiagnostic(condition.Position, "UNSUPPORTED_ACTION_CONDITION_OPERATOR", fmt.Sprintf("Action %s uses unsupported if operator %q.", action.Name, condition.Operator), "Use ==, !=, <, <=, >, or >=.")
		return
	}
	if !leftOK || !rightOK {
		return
	}
	if condition.Operator == "==" || condition.Operator == "!=" {
		if !actionValueAssignable(leftType, rightType) && !actionValueAssignable(rightType, leftType) {
			v.addDiagnostic(condition.Position, "ACTION_CONDITION_TYPE_MISMATCH", fmt.Sprintf("Action %s compares %s to %s.", action.Name, leftType, rightType), "Compare values from the same scalar type family.")
		}
		return
	}
	if !numberLikeType(leftType) || !numberLikeType(rightType) {
		v.addDiagnostic(condition.Position, "INCOMPATIBLE_ACTION_CONDITION", fmt.Sprintf("Action %s uses ordered comparison with %s and %s.", action.Name, leftType, rightType), "Use <, <=, >, or >= only with numeric values in action conditions.")
	}
}

func (v *semanticValidator) validateActionOperand(action CustomActionDecl, operand ActionOperandDecl, fields map[string]FieldDecl, inputs map[string]FieldDecl, values map[string]FieldDecl, position Position) (string, bool) {
	switch operand.Kind {
	case "string":
		return "text", true
	case "number":
		return "number", true
	case "boolean":
		return "boolean", true
	case "reference":
		if value, ok := values[operand.Value]; ok {
			return value.Type, true
		}
		if input, ok := inputs[operand.Value]; ok {
			return input.Type, true
		}
		if field, ok := fields[operand.Value]; ok {
			if _, relation := v.findEntity(field.Type); relation {
				v.addDiagnostic(position, "UNSUPPORTED_ACTION_REFERENCE", fmt.Sprintf("Action %s cannot use relation field %s in an expression.", action.Name, operand.Value), "Use stored primitive fields or inputs.")
				return "", false
			}
			return field.Type, true
		}
		v.addDiagnostic(position, "UNKNOWN_ACTION_VALUE", fmt.Sprintf("Action %s references unknown value %s.", action.Name, operand.Value), "Use an action input, source field, local value, or typed literal.")
		return "", false
	default:
		v.addDiagnostic(position, "INVALID_ACTION_VALUE", fmt.Sprintf("Action %s has invalid expression value %q.", action.Name, operand.Value), "Use an action input, source field, local value, or typed literal.")
		return "", false
	}
}

func (v *semanticValidator) validateActionAllow(action CustomActionDecl, roles map[string]RoleDecl) {
	seen := map[string]bool{}
	for _, allowed := range action.Allow {
		if seen[allowed] {
			v.addDiagnostic(action.Position, "DUPLICATE_ACTION_ALLOW", fmt.Sprintf("Action %s allow repeats %s.", action.Name, allowed), "Keep each role once in an action allow list.")
			continue
		}
		seen[allowed] = true
		if v.program.Auth == nil {
			v.addDiagnostic(action.Position, "AUTH_REQUIRED_FOR_ACTION_ALLOW", fmt.Sprintf("Action %s uses allow without an auth block.", action.Name), "Add an auth block or remove action allow.")
			continue
		}
		if allowed == "authenticated" {
			continue
		}
		if _, ok := roles[allowed]; !ok {
			v.addDiagnostic(action.Position, "UNKNOWN_ACTION_ALLOW_ROLE", fmt.Sprintf("Action %s references unknown role %s.", action.Name, allowed), "Use an existing role or authenticated in action allow.")
		}
	}
}

func customActionOtherSymbols(program Program) map[string]bool {
	symbols := queryOtherSymbols(program)
	for _, action := range program.Actions {
		delete(symbols, action.Name)
		delete(symbols, strings.ToLower(action.Name))
	}
	for _, query := range program.Queries {
		symbols[query.Name] = true
		symbols[strings.ToLower(query.Name)] = true
	}
	for _, job := range program.Jobs {
		symbols[job.Name] = true
		symbols[strings.ToLower(job.Name)] = true
	}
	return symbols
}

func findCustomAction(program Program, name string) (CustomActionDecl, bool) {
	for _, action := range program.Actions {
		if action.Name == name {
			return action, true
		}
	}
	return CustomActionDecl{}, false
}

func customActionsForPage(program Program, page PageDecl) []CustomActionDecl {
	actions := []CustomActionDecl{}
	for _, name := range page.Actions {
		if action, ok := findCustomAction(program, name); ok && action.Source == page.Source {
			actions = append(actions, action)
		}
	}
	return actions
}

func customActionsForEntity(program Program, entityName string) []CustomActionDecl {
	actions := []CustomActionDecl{}
	for _, action := range program.Actions {
		if action.Source == entityName {
			actions = append(actions, action)
		}
	}
	return actions
}

func customActionSetFieldNames(action CustomActionDecl) []string {
	fields := []string{}
	for _, set := range actionStatementSets(actionStatements(action)) {
		if !containsString(fields, set.Field) {
			fields = append(fields, set.Field)
		}
	}
	return fields
}

func actionStatements(action CustomActionDecl) []ActionStatementDecl {
	if len(action.Statements) > 0 {
		return action.Statements
	}
	statements := []ActionStatementDecl{}
	for _, set := range action.Sets {
		statements = append(statements, actionStatementFromSet(set))
	}
	return statements
}

func actionStatementsHaveSet(statements []ActionStatementDecl) bool {
	return len(actionStatementSets(statements)) > 0
}

func actionStatementSets(statements []ActionStatementDecl) []ActionSetDecl {
	sets := []ActionSetDecl{}
	for _, statement := range statements {
		switch statement.Kind {
		case "set":
			if statement.Set != nil {
				sets = append(sets, *statement.Set)
			}
		case "if":
			if statement.If != nil {
				sets = append(sets, actionStatementSets(statement.If.Then)...)
				sets = append(sets, actionStatementSets(statement.If.Else)...)
			}
		}
	}
	return sets
}

func cloneFieldMap(values map[string]FieldDecl) map[string]FieldDecl {
	clone := map[string]FieldDecl{}
	for name, field := range values {
		clone[name] = field
	}
	return clone
}

func numberLikeType(fieldType string) bool {
	return fieldType == "number" || fieldType == "integer" || fieldType == "decimal" || fieldType == "money"
}

func actionValueAssignable(valueType string, targetType string) bool {
	if valueType == targetType {
		return true
	}
	return numberLikeType(valueType) && numberLikeType(targetType)
}

func actionOperandAssignable(operand ActionOperandDecl, valueType string, targetType string) bool {
	switch operand.Kind {
	case "string":
		if targetType == "text" || targetType == "email" {
			return true
		}
		return queryLiteralMatchesField(QueryLiteral{Kind: "string", Value: operand.Value}, targetType)
	case "number":
		return queryLiteralMatchesField(QueryLiteral{Kind: "number", Value: operand.Value}, targetType)
	case "boolean":
		return queryLiteralMatchesField(QueryLiteral{Kind: "boolean", Value: operand.Value}, targetType)
	default:
		return actionValueAssignable(valueType, targetType)
	}
}

func numericLiteralIsZero(value string) bool {
	trimmed := strings.Trim(value, "0")
	return trimmed == "" || trimmed == "." || trimmed == "-." || trimmed == "-"
}
