package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type QueryDecl struct {
	Name           string            `json:"name"`
	Source         string            `json:"source"`
	Where          []QueryFilterDecl `json:"where,omitempty"`
	Sort           SortDecl          `json:"sort,omitempty"`
	Limit          int               `json:"limit,omitempty"`
	Position       Position          `json:"position"`
	SourcePosition Position          `json:"sourcePosition,omitempty"`
	SortPosition   Position          `json:"sortPosition,omitempty"`
	LimitPosition  Position          `json:"limitPosition,omitempty"`
}

type QueryFilterDecl struct {
	Field    string       `json:"field"`
	Operator string       `json:"operator"`
	Value    QueryLiteral `json:"value"`
	Position Position     `json:"position"`
}

type QueryLiteral struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

var queryNamePattern = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
var queryIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var queryNumberPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?$`)
var queryDatePattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
var queryDateTimePattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?(Z|[+-]([01][0-9]|2[0-3]):[0-5][0-9])$`)

func (p *parser) parseQuery(start int, parts []string) int {
	line := p.lineNumber(start)
	statement := p.lines[start]
	if len(parts) != 3 || len(statement.Tokens) != 3 || !queryStatementIdentifiers(statement, 0, 1) || statement.Tokens[2].Kind != tokenSymbol || parts[2] != "{" {
		p.addError(line, 1, "INVALID_QUERY_DECLARATION", "Query declaration must be `query Name {`.", "Example: `query LowStockProducts {`.")
		return start
	}
	query := QueryDecl{Name: parts[1], Position: p.position(line, 1)}
	seen := map[string]bool{}
	for index := start + 1; index < len(p.lines); index++ {
		statement := p.lines[index]
		tokens := statement.Tokens
		line := p.lineNumber(index)
		if len(tokens) == 1 && tokens[0].Kind == tokenSymbol && tokens[0].Value == "}" {
			p.program.Queries = append(p.program.Queries, query)
			return index
		}
		if len(tokens) == 0 {
			continue
		}
		keyword := tokens[0].Value
		if tokens[0].Kind != tokenIdentifier {
			p.addError(line, 1, "UNEXPECTED_QUERY_TOKEN", "Query clauses must begin with source, where, sort, or limit.", "Use one query clause per line.")
			continue
		}
		if keyword == "source" || keyword == "sort" || keyword == "limit" {
			if seen[keyword] {
				p.addError(line, 1, "DUPLICATE_QUERY_"+strings.ToUpper(keyword), fmt.Sprintf("Query %s already declares %s.", query.Name, keyword), "Keep one "+keyword+" clause inside each query.")
				continue
			}
			seen[keyword] = true
		}
		switch keyword {
		case "source":
			if len(tokens) != 2 || !queryStatementIdentifiers(statement, 1) {
				p.addError(line, 1, "INVALID_QUERY_SOURCE", "Query source must be `source EntityName`.", "Example: `source Product`.")
				continue
			}
			query.Source = tokens[1].Value
			query.SourcePosition = tokens[1].Position
		case "where":
			if len(tokens) != 4 || !queryStatementIdentifiers(statement, 1) || tokens[2].Kind != tokenOperator {
				p.addError(line, 1, "INVALID_QUERY_WHERE", "Query filter must be `where field operator literal`.", "Example: `where stock < 10`; add another where line for AND.")
				continue
			}
			literal, ok := parseQueryLiteral(tokens[3])
			if !ok {
				p.addError(line, tokens[3].Position.Column, "INVALID_QUERY_LITERAL", "Query values must be quoted strings, finite decimal numbers, or true/false.", "Quote text/date values; field references and expressions are not supported.")
				continue
			}
			query.Where = append(query.Where, QueryFilterDecl{Field: tokens[1].Value, Operator: tokens[2].Value, Value: literal, Position: tokens[1].Position})
		case "sort":
			if len(tokens) != 3 || !queryStatementIdentifiers(statement, 1, 2) {
				p.addError(line, 1, "INVALID_QUERY_SORT", "Query sort must be `sort field asc` or `sort field desc`.", "Example: `sort stock asc`.")
				continue
			}
			query.Sort = SortDecl{Field: tokens[1].Value, Direction: tokens[2].Value}
			query.SortPosition = tokens[1].Position
		case "limit":
			if len(tokens) != 2 || tokens[1].Kind != tokenIdentifier {
				p.addError(line, 1, "INVALID_QUERY_LIMIT", "Query limit must be an integer from 1 to 1000.", "Example: `limit 50`; omitted limit defaults to 100.")
				continue
			}
			limit, err := strconv.Atoi(tokens[1].Value)
			if err != nil || tokens[1].Value != strconv.Itoa(limit) || limit < 1 || limit > 1000 {
				p.addError(line, tokens[1].Position.Column, "INVALID_QUERY_LIMIT", "Query limit must be an integer from 1 to 1000.", "Example: `limit 50`; omitted limit defaults to 100.")
				continue
			}
			query.Limit = limit
			query.LimitPosition = tokens[1].Position
		default:
			p.addError(line, 1, "UNEXPECTED_QUERY_TOKEN", fmt.Sprintf("Unexpected query token %q.", keyword), "Use source, where, sort, or limit inside a query.")
		}
	}
	p.addError(line, 1, "UNCLOSED_QUERY", fmt.Sprintf("Query %s is missing a closing brace.", query.Name), "Add `}` after the query body.")
	return len(p.lines) - 1
}

func queryStatementIdentifiers(statement sourceStatement, indexes ...int) bool {
	for _, index := range indexes {
		if index >= len(statement.Tokens) || statement.Tokens[index].Kind != tokenIdentifier {
			return false
		}
	}
	return true
}

func parseQueryLiteral(token sourceToken) (QueryLiteral, bool) {
	if token.Kind == tokenString {
		return QueryLiteral{Kind: "string", Value: token.Value}, true
	}
	if token.Kind != tokenIdentifier {
		return QueryLiteral{}, false
	}
	if token.Value == "true" || token.Value == "false" {
		return QueryLiteral{Kind: "boolean", Value: token.Value}, true
	}
	if queryNumberPattern.MatchString(token.Value) {
		value, err := strconv.ParseFloat(token.Value, 64)
		if err == nil && !math.IsInf(value, 0) && !math.IsNaN(value) {
			return QueryLiteral{Kind: "number", Value: token.Value}, true
		}
	}
	return QueryLiteral{}, false
}

func (v *semanticValidator) validateQueries(entities map[string]EntityDecl) {
	queries := map[string]QueryDecl{}
	normalizedNames := map[string]string{}
	symbols := queryOtherSymbols(v.program)
	for _, query := range v.program.Queries {
		if existing, ok := queries[query.Name]; ok {
			v.addDiagnostic(query.Position, "DUPLICATE_QUERY", fmt.Sprintf("Query %s is already defined.", query.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		queries[query.Name] = query
		if !queryNamePattern.MatchString(query.Name) {
			v.addDiagnostic(query.Position, "INVALID_QUERY_NAME", fmt.Sprintf("Query name %q must use PascalCase letters and digits.", query.Name), "Use a name such as LowStockProducts.")
		}
		normalized := strings.ToLower(query.Name)
		if existing, ok := normalizedNames[normalized]; ok {
			v.addDiagnostic(query.Position, "QUERY_NAME_COLLISION", fmt.Sprintf("Query %s conflicts with query %s after name normalization.", query.Name, existing), "Choose distinct query names, including after lowercasing.")
		}
		normalizedNames[normalized] = query.Name
		if symbols[query.Name] || symbols[normalized] {
			v.addDiagnostic(query.Position, "QUERY_NAME_COLLISION", fmt.Sprintf("Query %s conflicts with another application symbol.", query.Name), "Choose a unique query name for unambiguous inspect --affected output.")
		}
		if query.Limit < 0 || query.Limit > 1000 || (query.Limit == 0 && query.LimitPosition.Line != 0) {
			v.addDiagnostic(query.LimitPosition, "INVALID_QUERY_LIMIT", "Query limit must be an integer from 1 to 1000.", "Use limit 50 or omit limit for the default of 100.")
		}
		if query.Source == "" {
			v.addDiagnostic(query.Position, "MISSING_QUERY_SOURCE", fmt.Sprintf("Query %s is missing a source entity.", query.Name), "Add `source EntityName` inside the query.")
			continue
		}
		if !queryIdentifierPattern.MatchString(query.Source) {
			v.addDiagnostic(query.SourcePosition, "INVALID_QUERY_SOURCE", "Query source must be a valid entity identifier.", "Use an existing entity name such as Product.")
			continue
		}
		entity, ok := entities[query.Source]
		if !ok {
			v.addDiagnostic(query.SourcePosition, "UNKNOWN_QUERY_SOURCE", fmt.Sprintf("Query %s uses unknown entity %s.", query.Name, query.Source), "Declare the entity or change the query source.")
			continue
		}
		seenFilters := map[QueryFilterDecl]bool{}
		for _, filter := range query.Where {
			key := filter
			key.Position = Position{}
			if seenFilters[key] {
				v.addDiagnostic(filter.Position, "DUPLICATE_QUERY_WHERE", "Query repeats the same where condition.", "Keep each where condition once; distinct conditions combine with AND.")
			}
			seenFilters[key] = true
			field, ok := v.validateQueryField(entity, filter.Field, filter.Position)
			if !ok {
				continue
			}
			if !supportedComparisonOperators[filter.Operator] || ((filter.Operator != "==" && filter.Operator != "!=") && !supportedComputedFieldTypes[field.Type] && field.Type != "date" && field.Type != "datetime") {
				v.addDiagnostic(filter.Position, "UNSUPPORTED_QUERY_OPERATOR", fmt.Sprintf("Operator %q is not supported for query field %s.%s (%s).", filter.Operator, entity.Name, field.Name, field.Type), "Use == or != for all scalar types; numeric/date/datetime fields also allow <, <=, >, >=.")
			}
			if !queryLiteralMatchesField(filter.Value, field.Type) {
				v.addDiagnostic(filter.Position, "QUERY_LITERAL_TYPE_MISMATCH", fmt.Sprintf("Query value %q does not match field %s.%s (%s).", filter.Value.Value, entity.Name, field.Name, field.Type), "Use quoted text/email, YYYY-MM-DD date, RFC3339 datetime, a finite decimal number, or true/false matching the stored type; number/integer require int32 values.")
			}
		}
		if query.Sort.Field != "" {
			v.validateQueryField(entity, query.Sort.Field, query.SortPosition)
			if query.Sort.Direction != "asc" && query.Sort.Direction != "desc" {
				v.addDiagnostic(query.SortPosition, "UNSUPPORTED_QUERY_SORT_DIRECTION", fmt.Sprintf("Query sort direction %q is not supported.", query.Sort.Direction), "Use asc or desc.")
			}
		}
	}
	for _, page := range v.program.Pages {
		if page.Query == "" {
			continue
		}
		query, ok := queries[page.Query]
		if !ok {
			v.addDiagnostic(page.QueryPosition, "UNKNOWN_PAGE_QUERY", fmt.Sprintf("Page %s references unknown query %s.", page.Name, page.Query), "Declare the query at the top level or change the page query reference.")
		} else if query.Source != page.Source {
			v.addDiagnostic(page.QueryPosition, "PAGE_QUERY_SOURCE_MISMATCH", fmt.Sprintf("Page %s source %s does not match query %s source %s.", page.Name, page.Source, query.Name, query.Source), "Use a query bound to the page source entity.")
		}
	}
}

func (v *semanticValidator) validateQueryField(entity EntityDecl, name string, position Position) (FieldDecl, bool) {
	if !queryIdentifierPattern.MatchString(name) {
		v.addDiagnostic(position, "INVALID_QUERY_FIELD", fmt.Sprintf("Query field %q is not a valid stored field identifier.", name), "Use one scalar field name without property paths or expressions.")
		return FieldDecl{}, false
	}
	if _, ok := computedFieldIndex(entity)[name]; ok {
		v.addDiagnostic(position, "UNSUPPORTED_QUERY_FIELD", fmt.Sprintf("Query cannot read computed display field %s.%s.", entity.Name, name), "Use stored scalar fields; computed fields are display-only.")
		return FieldDecl{}, false
	}
	field, ok := fieldIndex(entity)[name]
	if !ok {
		v.addDiagnostic(position, "UNKNOWN_QUERY_FIELD", fmt.Sprintf("Query references unknown stored field %s.%s.", entity.Name, name), "Use a declared stored scalar field; generated system fields are not query inputs.")
		return FieldDecl{}, false
	}
	if !supportedFieldTypes[field.Type] {
		v.addDiagnostic(position, "UNSUPPORTED_QUERY_FIELD", fmt.Sprintf("Query cannot read relation field %s.%s.", entity.Name, name), "Use stored scalar fields; relation queries are not supported in this MVP.")
		return FieldDecl{}, false
	}
	return field, true
}

func queryLiteralMatchesField(literal QueryLiteral, fieldType string) bool {
	switch fieldType {
	case "text", "email":
		return literal.Kind == "string"
	case "date":
		if literal.Kind != "string" || !queryDatePattern.MatchString(literal.Value) {
			return false
		}
		_, err := time.Parse("2006-01-02", literal.Value)
		return err == nil
	case "datetime":
		if literal.Kind != "string" || !queryDateTimePattern.MatchString(literal.Value) {
			return false
		}
		_, err := time.Parse(time.RFC3339Nano, literal.Value)
		return err == nil
	case "boolean":
		return literal.Kind == "boolean" && (literal.Value == "true" || literal.Value == "false")
	case "number", "integer", "decimal", "money":
		if literal.Kind != "number" || !queryNumberPattern.MatchString(literal.Value) {
			return false
		}
		value, err := strconv.ParseFloat(literal.Value, 64)
		if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
			return false
		}
		if fieldType == "number" || fieldType == "integer" {
			parts := strings.SplitN(literal.Value, ".", 2)
			if len(parts) == 2 && strings.Trim(parts[1], "0") != "" {
				return false
			}
			_, err := strconv.ParseInt(parts[0], 10, 32)
			return err == nil
		}
		return true
	}
	return false
}

func queryOtherSymbols(program Program) map[string]bool {
	symbols := setOf("app", "target", "auth", "database", "security", "cors", "deploy")
	add := func(name string) {
		symbols[name] = true
		symbols[strings.ToLower(name)] = true
	}
	add(program.App.Name)
	for _, entity := range program.Entities {
		add(entity.Name)
	}
	for _, page := range program.Pages {
		add(page.Name)
	}
	for _, role := range program.Roles {
		add(role.Name)
	}
	for _, workflow := range program.Workflows {
		add(workflow.Name)
	}
	for _, state := range program.States {
		add(state.Name)
	}
	for _, component := range program.Components {
		add(component.Name)
	}
	for _, api := range program.APIs {
		add(api.Name)
	}
	for _, layout := range program.Layouts {
		add(layout.Name)
	}
	return symbols
}

func findQuery(program Program, name string) (QueryDecl, bool) {
	for _, query := range program.Queries {
		if query.Name == name {
			return query, true
		}
	}
	return QueryDecl{}, false
}

func queryFieldNames(query QueryDecl) []string {
	fields := []string{}
	for _, filter := range query.Where {
		if !containsString(fields, filter.Field) {
			fields = append(fields, filter.Field)
		}
	}
	if query.Sort.Field != "" && !containsString(fields, query.Sort.Field) {
		fields = append(fields, query.Sort.Field)
	}
	return fields
}
