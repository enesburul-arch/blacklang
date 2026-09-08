package main

import "strings"

var relationLoadScopeOrder = []string{"list", "detail", "query", "mutation"}

var supportedRelationLoadScopes = setOf(
	"list",
	"detail",
	"query",
	"mutation",
	"none",
)

func modifierTakesList(name string) bool {
	return name == "load"
}

func parseListModifierValue(parts []string, index int) (string, int) {
	values := []string{}
	for index+1 < len(parts) {
		candidate := strings.Trim(parts[index+1], ",")
		if candidate == "" {
			index++
			continue
		}
		if isModifierBoundary(candidate) {
			break
		}
		values = append(values, candidate)
		index++
	}
	return strings.Join(values, ","), index
}

func isModifierBoundary(value string) bool {
	switch value {
	case "required", "unique", "optional", "default", "label", "placeholder", "help", "min", "max", "length", "regex", "url", "accept", "message", "load", "ui":
		return true
	default:
		return false
	}
}

func relationLoadScopesFromModifiers(modifiers []Modifier) []string {
	scopes := []string{}
	for _, modifier := range modifiers {
		if modifier.Name != "load" {
			continue
		}
		scopes = append(scopes, relationLoadScopesFromValue(modifier.Value)...)
	}
	return scopes
}

func relationLoadScopesFromValue(value string) []string {
	scopes := []string{}
	for _, scope := range strings.FieldsFunc(value, func(char rune) bool {
		return char == ',' || char == ' ' || char == '\t'
	}) {
		scope = strings.TrimSpace(scope)
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

func relationLoadExplicit(field FieldDecl) bool {
	for _, modifier := range field.Modifiers {
		if modifier.Name == "load" {
			return true
		}
	}
	return false
}

func relationLoadScopes(field FieldDecl) []string {
	if !relationLoadExplicit(field) {
		return append([]string{}, relationLoadScopeOrder...)
	}
	return relationLoadScopesFromModifiers(field.Modifiers)
}

func relationLoadIncludesContext(field FieldDecl, context string) bool {
	if !relationLoadExplicit(field) {
		return true
	}
	for _, scope := range relationLoadScopes(field) {
		if scope == "none" {
			return false
		}
		if scope == context {
			return true
		}
	}
	return false
}

func relationLoadModifierCount(field FieldDecl) int {
	count := 0
	for _, modifier := range field.Modifiers {
		if modifier.Name == "load" {
			count++
		}
	}
	return count
}

func (g *webGenerator) openapiRelationLoadMetadata(entity EntityDecl, context string) []any {
	metadata := []any{}
	for _, field := range g.relationFields(entity) {
		metadata = append(metadata, map[string]any{
			"field":    field.Name,
			"target":   field.Type,
			"contexts": relationLoadScopes(field),
			"explicit": relationLoadExplicit(field),
			"included": relationLoadIncludesContext(field, context),
		})
	}
	return metadata
}

func (g *webGenerator) addOpenAPIRelationLoad(operation map[string]any, entity EntityDecl, context string) {
	metadata := g.openapiRelationLoadMetadata(entity, context)
	if len(metadata) == 0 {
		return
	}
	operation["x-blacklang-relation-load-context"] = context
	operation["x-blacklang-relation-load"] = metadata
}

func (g *webGenerator) withOpenAPIRelationLoad(entity EntityDecl, context string, operation map[string]any) map[string]any {
	g.addOpenAPIRelationLoad(operation, entity, context)
	return operation
}
