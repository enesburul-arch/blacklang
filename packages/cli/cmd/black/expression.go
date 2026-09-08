package main

import "fmt"

type coreExpressionOperandParser func(sourceToken) (string, string, bool)

type coreExpressionParser struct {
	tokens       []sourceToken
	index        int
	parseOperand coreExpressionOperandParser
}

type ConditionComparisonDecl struct {
	Left     ExpressionDecl `json:"left"`
	Operator string         `json:"operator"`
	Right    ExpressionDecl `json:"right"`
	Position Position       `json:"position"`
}

type ConditionExpressionDecl struct {
	Kind       string                   `json:"kind"`
	Left       *ConditionExpressionDecl `json:"left,omitempty"`
	Right      *ConditionExpressionDecl `json:"right,omitempty"`
	Comparison *ConditionComparisonDecl `json:"comparison,omitempty"`
	Position   Position                 `json:"position"`
}

type conditionExpressionParser struct {
	tokens       []sourceToken
	index        int
	parseOperand coreExpressionOperandParser
}

func parseCoreExpression(tokens []sourceToken, parseOperand coreExpressionOperandParser) (ExpressionDecl, bool) {
	parser := coreExpressionParser{
		tokens:       tokens,
		parseOperand: parseOperand,
	}
	expr, ok := parser.parseBinary(1)
	if !ok || parser.index != len(tokens) {
		return ExpressionDecl{}, false
	}
	return expr, true
}

func parseConditionExpression(tokens []sourceToken, parseOperand coreExpressionOperandParser) (ConditionExpressionDecl, bool) {
	parser := conditionExpressionParser{
		tokens:       tokens,
		parseOperand: parseOperand,
	}
	condition, ok := parser.parseOr()
	if !ok || parser.index != len(tokens) {
		return ConditionExpressionDecl{}, false
	}
	return condition, true
}

func tokensToParts(tokens []sourceToken) []string {
	parts := []string{}
	for _, token := range tokens {
		if token.Kind == tokenSymbol && token.Value == "," {
			continue
		}
		parts = append(parts, token.Value)
	}
	return parts
}

func dereferenceExpression(expr *ExpressionDecl) ExpressionDecl {
	if expr == nil {
		return ExpressionDecl{}
	}
	return *expr
}

func (p *coreExpressionParser) parseBinary(minPrecedence int) (ExpressionDecl, bool) {
	left, ok := p.parsePrimary()
	if !ok {
		return ExpressionDecl{}, false
	}

	for p.index < len(p.tokens) {
		token := p.tokens[p.index]
		precedence := expressionOperatorPrecedence(token.Value)
		if precedence < minPrecedence {
			break
		}
		operator := token.Value
		p.index++
		right, ok := p.parseBinary(precedence + 1)
		if !ok {
			return ExpressionDecl{}, false
		}
		leftCopy := left
		rightCopy := right
		left = ExpressionDecl{
			Kind:     "binary",
			Operator: operator,
			Left:     &leftCopy,
			Right:    &rightCopy,
			Position: token.Position,
		}
	}

	return left, true
}

func (p *conditionExpressionParser) parseOr() (ConditionExpressionDecl, bool) {
	left, ok := p.parseAnd()
	if !ok {
		return ConditionExpressionDecl{}, false
	}
	for p.matchConditionKeyword("or") {
		operator := p.tokens[p.index]
		p.index++
		right, ok := p.parseAnd()
		if !ok {
			return ConditionExpressionDecl{}, false
		}
		leftCopy := left
		rightCopy := right
		left = ConditionExpressionDecl{
			Kind:     "or",
			Left:     &leftCopy,
			Right:    &rightCopy,
			Position: operator.Position,
		}
	}
	return left, true
}

func (p *conditionExpressionParser) parseAnd() (ConditionExpressionDecl, bool) {
	left, ok := p.parseUnary()
	if !ok {
		return ConditionExpressionDecl{}, false
	}
	for p.matchConditionKeyword("and") {
		operator := p.tokens[p.index]
		p.index++
		right, ok := p.parseUnary()
		if !ok {
			return ConditionExpressionDecl{}, false
		}
		leftCopy := left
		rightCopy := right
		left = ConditionExpressionDecl{
			Kind:     "and",
			Left:     &leftCopy,
			Right:    &rightCopy,
			Position: operator.Position,
		}
	}
	return left, true
}

func (p *conditionExpressionParser) parseUnary() (ConditionExpressionDecl, bool) {
	if p.matchConditionKeyword("not") {
		operator := p.tokens[p.index]
		p.index++
		condition, ok := p.parseUnary()
		if !ok {
			return ConditionExpressionDecl{}, false
		}
		conditionCopy := condition
		return ConditionExpressionDecl{
			Kind:     "not",
			Left:     &conditionCopy,
			Position: operator.Position,
		}, true
	}
	return p.parsePrimary()
}

func (p *conditionExpressionParser) parsePrimary() (ConditionExpressionDecl, bool) {
	if p.index >= len(p.tokens) {
		return ConditionExpressionDecl{}, false
	}
	token := p.tokens[p.index]
	if token.Kind == tokenSymbol && token.Value == "(" && p.startsBooleanGroup() {
		closeIndex := matchingConditionParen(p.tokens, p.index)
		if closeIndex < 0 {
			return ConditionExpressionDecl{}, false
		}
		nested, ok := parseConditionExpression(p.tokens[p.index+1:closeIndex], p.parseOperand)
		if !ok {
			return ConditionExpressionDecl{}, false
		}
		p.index = closeIndex + 1
		return nested, true
	}
	return p.parseComparison()
}

func (p *conditionExpressionParser) startsBooleanGroup() bool {
	closeIndex := matchingConditionParen(p.tokens, p.index)
	if closeIndex < 0 {
		return false
	}
	next := closeIndex + 1
	if next < len(p.tokens) && p.tokens[next].Kind == tokenOperator && supportedComparisonOperators[p.tokens[next].Value] {
		return false
	}
	return true
}

func (p *conditionExpressionParser) parseComparison() (ConditionExpressionDecl, bool) {
	start := p.index
	end := p.index
	depth := 0
	for end < len(p.tokens) {
		token := p.tokens[end]
		if token.Kind == tokenSymbol {
			if token.Value == "(" {
				depth++
			}
			if token.Value == ")" {
				if depth == 0 {
					break
				}
				depth--
			}
		}
		if depth == 0 && token.Kind == tokenIdentifier && (token.Value == "and" || token.Value == "or") {
			break
		}
		end++
	}
	if end <= start {
		return ConditionExpressionDecl{}, false
	}
	operatorIndex := -1
	for index := start; index < end; index++ {
		token := p.tokens[index]
		if token.Kind != tokenOperator || !supportedComparisonOperators[token.Value] {
			continue
		}
		if operatorIndex != -1 {
			return ConditionExpressionDecl{}, false
		}
		operatorIndex = index
	}
	if operatorIndex <= start || operatorIndex >= end-1 {
		return ConditionExpressionDecl{}, false
	}
	left, leftOK := parseCoreExpression(p.tokens[start:operatorIndex], p.parseOperand)
	right, rightOK := parseCoreExpression(p.tokens[operatorIndex+1:end], p.parseOperand)
	if !leftOK || !rightOK {
		return ConditionExpressionDecl{}, false
	}
	p.index = end
	comparison := ConditionComparisonDecl{
		Left:     left,
		Operator: p.tokens[operatorIndex].Value,
		Right:    right,
		Position: p.tokens[operatorIndex].Position,
	}
	comparisonCopy := comparison
	return ConditionExpressionDecl{
		Kind:       "comparison",
		Comparison: &comparisonCopy,
		Position:   comparison.Position,
	}, true
}

func (p *conditionExpressionParser) matchConditionKeyword(keyword string) bool {
	return p.index < len(p.tokens) && p.tokens[p.index].Kind == tokenIdentifier && p.tokens[p.index].Value == keyword
}

func matchingConditionParen(tokens []sourceToken, openIndex int) int {
	if openIndex >= len(tokens) || tokens[openIndex].Kind != tokenSymbol || tokens[openIndex].Value != "(" {
		return -1
	}
	depth := 0
	for index := openIndex; index < len(tokens); index++ {
		token := tokens[index]
		if token.Kind == tokenSymbol && token.Value == "(" {
			depth++
			continue
		}
		if token.Kind == tokenSymbol && token.Value == ")" {
			depth--
			if depth == 0 {
				return index
			}
		}
	}
	return -1
}

func firstConditionComparison(condition ConditionExpressionDecl) (ConditionComparisonDecl, bool) {
	if condition.Kind == "comparison" && condition.Comparison != nil {
		return *condition.Comparison, true
	}
	if condition.Left != nil {
		if comparison, ok := firstConditionComparison(*condition.Left); ok {
			return comparison, true
		}
	}
	if condition.Right != nil {
		return firstConditionComparison(*condition.Right)
	}
	return ConditionComparisonDecl{}, false
}

func (p *coreExpressionParser) parsePrimary() (ExpressionDecl, bool) {
	if p.index >= len(p.tokens) {
		return ExpressionDecl{}, false
	}
	token := p.tokens[p.index]
	if token.Kind == tokenSymbol && token.Value == "(" {
		p.index++
		expr, ok := p.parseBinary(1)
		if !ok || p.index >= len(p.tokens) {
			return ExpressionDecl{}, false
		}
		if p.tokens[p.index].Kind != tokenSymbol || p.tokens[p.index].Value != ")" {
			return ExpressionDecl{}, false
		}
		p.index++
		return expr, true
	}
	if token.Kind == tokenSymbol && token.Value == ")" {
		return ExpressionDecl{}, false
	}

	kind, value, ok := p.parseOperand(token)
	if !ok {
		return ExpressionDecl{}, false
	}
	p.index++
	return ExpressionDecl{
		Kind:      "operand",
		ValueKind: kind,
		Value:     value,
		Position:  token.Position,
	}, true
}

func expressionOperatorPrecedence(operator string) int {
	switch operator {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}

func expressionHasBinary(expr ExpressionDecl) bool {
	if expr.Kind == "binary" {
		return true
	}
	if expr.Left != nil && expressionHasBinary(*expr.Left) {
		return true
	}
	if expr.Right != nil && expressionHasBinary(*expr.Right) {
		return true
	}
	return false
}

func expressionIsZeroNumber(expr ExpressionDecl) bool {
	return expr.Kind == "operand" && expr.ValueKind == "number" && numericLiteralIsZero(expr.Value)
}

func expressionDivisionDivisors(expr ExpressionDecl) []ExpressionDecl {
	divisors := []ExpressionDecl{}
	if expr.Left != nil {
		divisors = append(divisors, expressionDivisionDivisors(*expr.Left)...)
	}
	if expr.Kind == "binary" && expr.Operator == "/" && expr.Right != nil {
		divisors = append(divisors, *expr.Right)
	}
	if expr.Right != nil {
		divisors = append(divisors, expressionDivisionDivisors(*expr.Right)...)
	}
	return divisors
}

func expressionOperands(expr ExpressionDecl) []ExpressionDecl {
	if expr.Kind != "binary" {
		if expr.Kind == "operand" {
			return []ExpressionDecl{expr}
		}
		return nil
	}
	operands := []ExpressionDecl{}
	if expr.Left != nil {
		operands = append(operands, expressionOperands(*expr.Left)...)
	}
	if expr.Right != nil {
		operands = append(operands, expressionOperands(*expr.Right)...)
	}
	return operands
}

func formatCoreExpression(expr ExpressionDecl) string {
	return formatCoreExpressionWithParent(expr, 0, "", false)
}

func formatCoreExpressionWithParent(expr ExpressionDecl, parentPrecedence int, parentOperator string, rightChild bool) string {
	if expr.Kind != "binary" {
		if expr.ValueKind == "string" {
			return quoteBlackString(expr.Value)
		}
		return expr.Value
	}

	precedence := expressionOperatorPrecedence(expr.Operator)
	left := ""
	right := ""
	if expr.Left != nil {
		left = formatCoreExpressionWithParent(*expr.Left, precedence, expr.Operator, false)
	}
	if expr.Right != nil {
		right = formatCoreExpressionWithParent(*expr.Right, precedence, expr.Operator, true)
	}
	rendered := fmt.Sprintf("%s %s %s", left, expr.Operator, right)
	if precedence < parentPrecedence || (rightChild && precedence == parentPrecedence && (parentOperator == "-" || parentOperator == "/")) {
		return "(" + rendered + ")"
	}
	return rendered
}

func renderCoreExpressionJS(expr ExpressionDecl, operand func(ExpressionDecl) string, numeric bool) string {
	return renderCoreExpressionJSWithParent(expr, operand, numeric, 0, "", false)
}

func renderCoreExpressionJSWithParent(expr ExpressionDecl, operand func(ExpressionDecl) string, numeric bool, parentPrecedence int, parentOperator string, rightChild bool) string {
	if expr.Kind != "binary" {
		raw := operand(expr)
		if numeric {
			return fmt.Sprintf("Number(%s ?? 0)", raw)
		}
		return raw
	}
	precedence := expressionOperatorPrecedence(expr.Operator)
	left := "0"
	right := "0"
	if expr.Left != nil {
		left = renderCoreExpressionJSWithParent(*expr.Left, operand, true, precedence, expr.Operator, false)
	}
	if expr.Right != nil {
		right = renderCoreExpressionJSWithParent(*expr.Right, operand, true, precedence, expr.Operator, true)
	}
	rendered := fmt.Sprintf("%s %s %s", left, expr.Operator, right)
	if precedence < parentPrecedence || (rightChild && precedence == parentPrecedence && (parentOperator == "-" || parentOperator == "/")) {
		return "(" + rendered + ")"
	}
	return rendered
}
