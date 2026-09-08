package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type SeedDecl struct {
	Name           string        `json:"name"`
	Source         string        `json:"source,omitempty"`
	Rows           []SeedRowDecl `json:"rows,omitempty"`
	Position       Position      `json:"position"`
	SourcePosition Position      `json:"sourcePosition,omitempty"`
}

type SeedRowDecl struct {
	Key      string          `json:"key"`
	Values   []SeedValueDecl `json:"values,omitempty"`
	Position Position        `json:"position"`
}

type SeedValueDecl struct {
	Field    string      `json:"field"`
	Value    SeedLiteral `json:"value"`
	Position Position    `json:"position"`
}

type SeedLiteral struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

var seedNamePattern = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
var seedRowKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

func (p *parser) parseSeed(start int, parts []string) int {
	line := p.lineNumber(start)
	statement := p.lines[start]
	if len(parts) != 3 || len(statement.Tokens) != 3 || !queryStatementIdentifiers(statement, 0, 1) || statement.Tokens[2].Kind != tokenSymbol || parts[2] != "{" {
		p.addError(line, 1, "INVALID_SEED_DECLARATION", "Seed declaration must be `seed Name {`.", "Example: `seed DemoProducts {`.")
		return start
	}

	seed := SeedDecl{
		Name:     parts[1],
		Rows:     []SeedRowDecl{},
		Position: p.position(line, 1),
	}
	seen := map[string]bool{}

	for index := start + 1; index < len(p.lines); index++ {
		statement := p.lines[index]
		tokens := statement.Tokens
		line := p.lineNumber(index)
		rowParts := statement.Parts()
		if isClosingBrace(rowParts) {
			p.program.Seeds = append(p.program.Seeds, seed)
			return index
		}
		if len(tokens) == 0 {
			continue
		}
		if tokens[0].Kind != tokenIdentifier {
			p.addError(line, 1, "UNEXPECTED_SEED_TOKEN", "Seed clauses must begin with source or row.", "Use one seed clause per line.")
			continue
		}

		switch rowParts[0] {
		case "source":
			if seen["source"] {
				p.addError(line, 1, "DUPLICATE_SEED_SOURCE", fmt.Sprintf("Seed %s already declares source.", seed.Name), "Keep one source clause inside each seed.")
				continue
			}
			seen["source"] = true
			if len(tokens) != 2 || !queryStatementIdentifiers(statement, 1) {
				p.addError(line, 1, "INVALID_SEED_SOURCE", "Seed source must be `source EntityName`.", "Example: `source Product`.")
				continue
			}
			seed.Source = tokens[1].Value
			seed.SourcePosition = tokens[1].Position
		case "row":
			row, next, ok := p.parseSeedRow(index, statement)
			index = next
			if ok {
				seed.Rows = append(seed.Rows, row)
			}
		default:
			p.addError(line, 1, "UNEXPECTED_SEED_TOKEN", fmt.Sprintf("Unexpected seed token %q.", rowParts[0]), "Use source or row inside a seed block.")
		}
	}

	p.addError(line, 1, "UNCLOSED_SEED", fmt.Sprintf("Seed %s is missing a closing brace.", seed.Name), "Add `}` after the seed body.")
	return len(p.lines) - 1
}

func (p *parser) parseSeedRow(start int, statement sourceStatement) (SeedRowDecl, int, bool) {
	line := p.lineNumber(start)
	tokens := statement.Tokens
	parts := statement.Parts()
	if len(parts) != 3 || len(tokens) != 3 || !queryStatementIdentifiers(statement, 0, 1) || tokens[2].Kind != tokenSymbol || parts[2] != "{" {
		p.addError(line, 1, "INVALID_SEED_ROW", "Seed row must be `row Key {`.", "Example: `row DemoProduct {`.")
		return SeedRowDecl{}, start, false
	}

	row := SeedRowDecl{
		Key:      parts[1],
		Values:   []SeedValueDecl{},
		Position: tokens[1].Position,
	}

	for index := start + 1; index < len(p.lines); index++ {
		statement := p.lines[index]
		tokens := statement.Tokens
		line := p.lineNumber(index)
		parts := statement.Parts()
		if isClosingBrace(parts) {
			return row, index, true
		}
		if len(tokens) == 0 {
			continue
		}
		if tokens[0].Kind != tokenIdentifier {
			p.addError(line, 1, "UNEXPECTED_SEED_ROW_TOKEN", "Seed row values must begin with a stored field name.", "Use `field value` or `relationField ref RowKey`.")
			continue
		}
		value, ok := parseSeedValue(tokens)
		if !ok {
			p.addError(line, 1, "INVALID_SEED_VALUE", "Seed value must be `field literal` or `relationField ref RowKey`.", "Example: `stock 3` or `customer ref DemoCustomer`.")
			continue
		}
		row.Values = append(row.Values, value)
	}

	p.addError(line, 1, "UNCLOSED_SEED_ROW", fmt.Sprintf("Seed row %s is missing a closing brace.", row.Key), "Add `}` after the row values.")
	return SeedRowDecl{}, len(p.lines) - 1, false
}

func parseSeedValue(tokens []sourceToken) (SeedValueDecl, bool) {
	if len(tokens) == 3 && tokens[0].Kind == tokenIdentifier && tokens[1].Kind == tokenIdentifier && tokens[1].Value == "ref" && tokens[2].Kind == tokenIdentifier {
		return SeedValueDecl{
			Field:    tokens[0].Value,
			Value:    SeedLiteral{Kind: "ref", Value: tokens[2].Value},
			Position: tokens[0].Position,
		}, true
	}
	if len(tokens) != 2 || tokens[0].Kind != tokenIdentifier {
		return SeedValueDecl{}, false
	}
	literal, ok := parseQueryLiteral(tokens[1])
	if !ok {
		return SeedValueDecl{}, false
	}
	return SeedValueDecl{
		Field:    tokens[0].Value,
		Value:    SeedLiteral{Kind: literal.Kind, Value: literal.Value},
		Position: tokens[0].Position,
	}, true
}

func (v *semanticValidator) validateSeeds(entityIndex map[string]EntityDecl) map[string]SeedDecl {
	seedIndex := map[string]SeedDecl{}
	normalizedNames := map[string]string{}
	symbols := seedOtherSymbols(v.program)
	rowsByEntity := map[string]map[string]SeedRowDecl{}

	for _, seed := range v.program.Seeds {
		if existing, ok := seedIndex[seed.Name]; ok {
			v.addDiagnostic(seed.Position, "DUPLICATE_SEED", fmt.Sprintf("Seed %s is already defined.", seed.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		seedIndex[seed.Name] = seed
		if !seedNamePattern.MatchString(seed.Name) {
			v.addDiagnostic(seed.Position, "INVALID_SEED_NAME", fmt.Sprintf("Seed name %q must use PascalCase letters and digits.", seed.Name), "Use a name such as DemoProducts.")
		}
		normalized := strings.ToLower(seed.Name)
		if existing, ok := normalizedNames[normalized]; ok {
			v.addDiagnostic(seed.Position, "SEED_NAME_COLLISION", fmt.Sprintf("Seed %s conflicts with seed %s after name normalization.", seed.Name, existing), "Choose distinct seed names, including after lowercasing.")
		}
		normalizedNames[normalized] = seed.Name
		if symbols[seed.Name] || symbols[normalized] {
			v.addDiagnostic(seed.Position, "SEED_NAME_COLLISION", fmt.Sprintf("Seed %s conflicts with another application symbol.", seed.Name), "Choose a unique seed name for unambiguous inspect --affected output.")
		}
		if seed.Source == "" {
			v.addDiagnostic(seed.Position, "MISSING_SEED_SOURCE", fmt.Sprintf("Seed %s is missing a source entity.", seed.Name), "Add `source EntityName` inside the seed.")
			continue
		}
		if !queryIdentifierPattern.MatchString(seed.Source) {
			v.addDiagnostic(seed.SourcePosition, "INVALID_SEED_SOURCE", "Seed source must be a valid entity identifier.", "Use an existing entity name such as Product.")
			continue
		}
		if _, ok := entityIndex[seed.Source]; !ok {
			v.addDiagnostic(seed.SourcePosition, "UNKNOWN_SEED_SOURCE", fmt.Sprintf("Seed %s uses unknown entity %s.", seed.Name, seed.Source), "Declare the entity or change the seed source.")
			continue
		}
		if _, ok := rowsByEntity[seed.Source]; !ok {
			rowsByEntity[seed.Source] = map[string]SeedRowDecl{}
		}
		for _, row := range seed.Rows {
			if !seedRowKeyPattern.MatchString(row.Key) {
				v.addDiagnostic(row.Position, "INVALID_SEED_ROW", fmt.Sprintf("Seed row key %q must use letters, digits, or underscore and start with a letter.", row.Key), "Use a stable key such as DemoProductLow.")
			}
			if existing, ok := rowsByEntity[seed.Source][row.Key]; ok {
				v.addDiagnostic(row.Position, "DUPLICATE_SEED_ROW", fmt.Sprintf("Seed row %s for entity %s is already defined.", row.Key, seed.Source), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
				continue
			}
			rowsByEntity[seed.Source][row.Key] = row
		}
	}

	uniqueValues := map[string]map[string]SeedValueDecl{}
	for _, seed := range v.program.Seeds {
		entity, ok := entityIndex[seed.Source]
		if !ok {
			continue
		}
		fields := fieldIndex(entity)
		computedFields := computedFieldIndex(entity)
		for _, row := range seed.Rows {
			v.validateSeedRow(seed, entity, row, fields, computedFields, rowsByEntity, uniqueValues)
		}
	}

	return seedIndex
}

func (v *semanticValidator) validateSeedRow(seed SeedDecl, entity EntityDecl, row SeedRowDecl, fields map[string]FieldDecl, computedFields map[string]ComputedFieldDecl, rowsByEntity map[string]map[string]SeedRowDecl, uniqueValues map[string]map[string]SeedValueDecl) {
	seenValues := map[string]SeedValueDecl{}
	for _, value := range row.Values {
		if !queryIdentifierPattern.MatchString(value.Field) {
			v.addDiagnostic(value.Position, "INVALID_SEED_FIELD", fmt.Sprintf("Seed row %s field %q is not a valid identifier.", row.Key, value.Field), "Use one stored field name such as sku.")
			continue
		}
		if existing, ok := seenValues[value.Field]; ok {
			v.addDiagnostic(value.Position, "DUPLICATE_SEED_FIELD", fmt.Sprintf("Seed row %s repeats field %s.", row.Key, value.Field), fmt.Sprintf("First value for this field is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		seenValues[value.Field] = value
		if _, ok := computedFields[value.Field]; ok {
			v.addDiagnostic(value.Position, "UNSUPPORTED_SEED_FIELD", fmt.Sprintf("Seed row %s cannot set computed display field %s.%s.", row.Key, entity.Name, value.Field), "Seed only stored fields; computed values are derived in generated UI.")
			continue
		}
		field, ok := fields[value.Field]
		if !ok {
			v.addDiagnostic(value.Position, "UNKNOWN_SEED_FIELD", fmt.Sprintf("Seed row %s references unknown field %s.%s.", row.Key, entity.Name, value.Field), "Use a stored field declared on the seed source entity.")
			continue
		}
		if supportedFieldTypes[field.Type] {
			if value.Value.Kind == "ref" {
				v.addDiagnostic(value.Position, "UNSUPPORTED_SEED_REF", fmt.Sprintf("Seed row %s uses ref for scalar field %s.%s.", row.Key, entity.Name, field.Name), "Use a typed literal for scalar fields and ref only for relation fields.")
				continue
			}
			literal := QueryLiteral{Kind: value.Value.Kind, Value: value.Value.Value}
			if !queryLiteralMatchesField(literal, field.Type) {
				v.addDiagnostic(value.Position, "SEED_VALUE_TYPE_MISMATCH", fmt.Sprintf("Seed row %s value for %s.%s does not match %s.", row.Key, entity.Name, field.Name, field.Type), "Use quoted text/email/date/datetime/file/image, finite numbers, or true/false matching the field type.")
				continue
			}
			v.validateSeedValueConstraints(entity, row, field, value)
			if hasModifier(field, "unique") {
				key := entity.Name + "." + field.Name
				if _, ok := uniqueValues[key]; !ok {
					uniqueValues[key] = map[string]SeedValueDecl{}
				}
				uniqueKey := value.Value.Kind + ":" + value.Value.Value
				if existing, ok := uniqueValues[key][uniqueKey]; ok {
					v.addDiagnostic(value.Position, "DUPLICATE_SEED_UNIQUE_VALUE", fmt.Sprintf("Seed row %s repeats unique value for %s.%s.", row.Key, entity.Name, field.Name), fmt.Sprintf("First seed value is at %s:%d.", existing.Position.File, existing.Position.Line))
				} else {
					uniqueValues[key][uniqueKey] = value
				}
			}
			continue
		}
		if value.Value.Kind != "ref" {
			v.addDiagnostic(value.Position, "SEED_RELATION_REQUIRES_REF", fmt.Sprintf("Seed row %s relation field %s.%s must use ref.", row.Key, entity.Name, field.Name), "Use `"+field.Name+" ref OtherSeedRow` for relation fields.")
			continue
		}
		targetRows, ok := rowsByEntity[field.Type]
		if !ok {
			v.addDiagnostic(value.Position, "UNKNOWN_SEED_REF", fmt.Sprintf("Seed row %s references %s but no seed rows exist for entity %s.", row.Key, value.Value.Value, field.Type), "Declare a seed row for the related entity.")
			continue
		}
		if _, ok := targetRows[value.Value.Value]; !ok {
			v.addDiagnostic(value.Position, "UNKNOWN_SEED_REF", fmt.Sprintf("Seed row %s references unknown %s row %s.", row.Key, field.Type, value.Value.Value), "Use the key of an existing seed row for the related entity.")
		}
	}

	for _, field := range entity.Fields {
		if hasModifier(field, "required") && !hasModifier(field, "default") {
			if _, ok := seenValues[field.Name]; !ok {
				v.addDiagnostic(row.Position, "MISSING_SEED_FIELD", fmt.Sprintf("Seed row %s for %s is missing required field %s.", row.Key, entity.Name, field.Name), "Add the required field to the row or give the entity field a default.")
			}
		}
	}
}

func (v *semanticValidator) validateSeedValueConstraints(entity EntityDecl, row SeedRowDecl, field FieldDecl, value SeedValueDecl) {
	if value.Value.Kind == "string" && !mediaLiteralMatchesField(value.Value.Value, field.Type) {
		v.addDiagnostic(value.Position, "SEED_VALUE_CONSTRAINT_MISMATCH", fmt.Sprintf("Seed row %s value for %s.%s is not a valid %s data URL or URL.", row.Key, entity.Name, field.Name, field.Type), "Use a data URL or absolute http(s) URL for seeded file and image values.")
	}
	for _, modifier := range field.Modifiers {
		switch modifier.Name {
		case "min":
			actual, ok := seedNumericValue(value.Value)
			expected, err := strconv.ParseFloat(modifier.Value, 64)
			if ok && err == nil && actual < expected {
				v.addDiagnostic(value.Position, "SEED_VALUE_CONSTRAINT_MISMATCH", fmt.Sprintf("Seed row %s value for %s.%s is below min %s.", row.Key, entity.Name, field.Name, modifier.Value), "Use a value that satisfies the field constraints.")
			}
		case "max":
			actual, ok := seedNumericValue(value.Value)
			expected, err := strconv.ParseFloat(modifier.Value, 64)
			if ok && err == nil && actual > expected {
				v.addDiagnostic(value.Position, "SEED_VALUE_CONSTRAINT_MISMATCH", fmt.Sprintf("Seed row %s value for %s.%s is above max %s.", row.Key, entity.Name, field.Name, modifier.Value), "Use a value that satisfies the field constraints.")
			}
		case "length":
			minLength, maxLength, ok := parseLengthConstraint(modifier.Value)
			if ok && value.Value.Kind == "string" {
				length := len([]rune(value.Value.Value))
				if length < minLength || length > maxLength {
					v.addDiagnostic(value.Position, "SEED_VALUE_CONSTRAINT_MISMATCH", fmt.Sprintf("Seed row %s value for %s.%s does not satisfy length %s.", row.Key, entity.Name, field.Name, modifier.Value), "Use a value that satisfies the field constraints.")
				}
			}
		case "regex":
			pattern, err := regexp.Compile(modifier.Value)
			if err == nil && value.Value.Kind == "string" && !pattern.MatchString(value.Value.Value) {
				v.addDiagnostic(value.Position, "SEED_VALUE_CONSTRAINT_MISMATCH", fmt.Sprintf("Seed row %s value for %s.%s does not match regex.", row.Key, entity.Name, field.Name), "Use a value that satisfies the field constraints.")
			}
		case "url":
			if value.Value.Kind == "string" {
				parsed, err := url.ParseRequestURI(value.Value.Value)
				if err != nil || parsed.Scheme == "" || parsed.Host == "" {
					v.addDiagnostic(value.Position, "SEED_VALUE_CONSTRAINT_MISMATCH", fmt.Sprintf("Seed row %s value for %s.%s is not a valid URL.", row.Key, entity.Name, field.Name), "Use an absolute URL such as https://example.com.")
				}
			}
		}
	}
}

func mediaLiteralMatchesField(value string, fieldType string) bool {
	switch fieldType {
	case "image":
		return strings.HasPrefix(value, "data:image/") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://")
	case "file":
		return strings.HasPrefix(value, "data:") || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://")
	default:
		return true
	}
}

func seedNumericValue(value SeedLiteral) (float64, bool) {
	if value.Kind != "number" {
		return 0, false
	}
	actual, err := strconv.ParseFloat(value.Value, 64)
	if err != nil {
		return 0, false
	}
	return actual, true
}

func seedOtherSymbols(program Program) map[string]bool {
	symbols := setOf("app", "target", "auth", "database", "security", "cors", "deploy", "seed", "test")
	add := func(name string) {
		symbols[name] = true
		symbols[strings.ToLower(name)] = true
	}
	add(program.App.Name)
	for _, entity := range program.Entities {
		add(entity.Name)
	}
	for _, query := range program.Queries {
		add(query.Name)
	}
	for _, action := range program.Actions {
		add(action.Name)
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
	for _, job := range program.Jobs {
		add(job.Name)
	}
	for _, test := range program.Tests {
		add(test.Name)
	}
	return symbols
}

func findSeed(program Program, name string) (SeedDecl, bool) {
	for _, seed := range program.Seeds {
		if seed.Name == name {
			return seed, true
		}
	}
	return SeedDecl{}, false
}

func seedUsesField(seed SeedDecl, fieldName string) bool {
	for _, row := range seed.Rows {
		for _, value := range row.Values {
			if value.Field == fieldName {
				return true
			}
		}
	}
	return false
}
