package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var supportedFieldTypes = setOf(
	"text",
	"number",
	"integer",
	"decimal",
	"money",
	"email",
	"boolean",
	"date",
	"datetime",
)

var supportedFieldModifiers = setOf(
	"required",
	"unique",
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

var searchableFieldTypes = setOf(
	"text",
	"email",
)

var supportedActions = setOf(
	"create",
	"edit",
	"delete",
	"archive",
	"restore",
)

var supportedAuthStrategies = setOf(
	"emailPassword",
)

var supportedAuthSessions = setOf(
	"cookie",
)

var supportedPermissionActions = setOf(
	"all",
	"manage",
	"read",
	"create",
	"update",
	"delete",
)

var supportedAPIMethods = setOf(
	"GET",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
)

func Validate(program Program) []Diagnostic {
	validator := semanticValidator{
		program:     program,
		diagnostics: []Diagnostic{},
	}
	validator.validate()
	return validator.diagnostics
}

type semanticValidator struct {
	program     Program
	diagnostics []Diagnostic
}

func (v *semanticValidator) validate() {
	v.validateApp()
	v.validateAuth()
	v.validateDatabase()
	entityIndex := v.validateEntities()
	roleIndex := v.validateRoles(entityIndex)
	v.validateAPIs()
	layoutIndex := v.validateLayouts()
	pageIndex := v.validatePages(entityIndex, layoutIndex, roleIndex)
	v.validateWorkflows(entityIndex, roleIndex)
	v.validateStates(entityIndex)
	v.validateComponents(entityIndex)
	v.validateLayoutReferences(layoutIndex, pageIndex)
}

func (v *semanticValidator) validateAPIs() {
	apiIndex := map[string]APIDecl{}
	routeIndex := map[string]APIDecl{}
	for _, api := range v.program.APIs {
		if existing, ok := apiIndex[api.Name]; ok {
			v.addDiagnostic(api.Position, "DUPLICATE_API", fmt.Sprintf("API %s is already defined.", api.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		apiIndex[api.Name] = api

		if api.Method == "" {
			v.addDiagnostic(api.Position, "MISSING_API_METHOD", fmt.Sprintf("API %s is missing a method.", api.Name), "Add `method GET`, `method POST`, `method PUT`, `method PATCH`, or `method DELETE`.")
		} else if !supportedAPIMethods[strings.ToUpper(api.Method)] {
			v.addDiagnostic(api.Position, "UNSUPPORTED_API_METHOD", fmt.Sprintf("API %s uses unsupported method %q.", api.Name, api.Method), "Use GET, POST, PUT, PATCH, or DELETE.")
		}

		if api.Path == "" {
			v.addDiagnostic(api.Position, "MISSING_API_PATH", fmt.Sprintf("API %s is missing a path.", api.Name), "Add `path \"/api/name\"`.")
		} else if !strings.HasPrefix(api.Path, "/") {
			v.addDiagnostic(api.Position, "INVALID_API_PATH", fmt.Sprintf("API %s path must start with `/`.", api.Name), "Use an absolute API path such as `/api/reports/low-stock`.")
		}

		routeKey := strings.ToUpper(api.Method) + " " + api.Path
		if api.Method != "" && api.Path != "" {
			if existing, ok := routeIndex[routeKey]; ok {
				v.addDiagnostic(api.Position, "DUPLICATE_API_ROUTE", fmt.Sprintf("API %s reuses route %s.", api.Name, routeKey), fmt.Sprintf("First definition is API %s at %s:%d.", existing.Name, existing.Position.File, existing.Position.Line))
			}
			routeIndex[routeKey] = api
		}

		if api.Access != "" && api.Access != "public" && api.Access != "private" {
			v.addDiagnostic(api.Position, "UNSUPPORTED_API_ACCESS", fmt.Sprintf("API %s uses unsupported access %q.", api.Name, api.Access), "Use public or private.")
		}

		paramIndex := map[string]APIParamDecl{}
		for _, param := range api.Params {
			if existing, ok := paramIndex[param.Name]; ok {
				v.addDiagnostic(param.Position, "DUPLICATE_API_PARAM", fmt.Sprintf("API %s path parameter %s is already defined.", api.Name, param.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
				continue
			}
			paramIndex[param.Name] = param
			if !supportedFieldTypes[param.Type] {
				v.addDiagnostic(param.Position, "UNSUPPORTED_API_PARAM_TYPE", fmt.Sprintf("API %s path parameter %s uses unsupported type %q.", api.Name, param.Name, param.Type), "Use primitive field types for API path parameters.")
			}
		}

		for _, pathParam := range apiPathParamNames(api.Path) {
			if _, ok := paramIndex[pathParam]; !ok {
				v.addDiagnostic(api.Position, "MISSING_API_PATH_PARAM", fmt.Sprintf("API %s path uses {%s} without a matching param declaration.", api.Name, pathParam), fmt.Sprintf("Add `param %s text` inside the api block.", pathParam))
			}
		}
		for _, param := range api.Params {
			if !containsString(apiPathParamNames(api.Path), param.Name) {
				v.addDiagnostic(param.Position, "UNUSED_API_PARAM", fmt.Sprintf("API %s declares param %s but the path does not use it.", api.Name, param.Name), fmt.Sprintf("Use {%s} in the path or remove the param.", param.Name))
			}
		}

		queryIndex := map[string]APIParamDecl{}
		for _, query := range api.Queries {
			if existing, ok := queryIndex[query.Name]; ok {
				v.addDiagnostic(query.Position, "DUPLICATE_API_QUERY", fmt.Sprintf("API %s query parameter %s is already defined.", api.Name, query.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
				continue
			}
			queryIndex[query.Name] = query
			if !supportedFieldTypes[query.Type] {
				v.addDiagnostic(query.Position, "UNSUPPORTED_API_QUERY_TYPE", fmt.Sprintf("API %s query parameter %s uses unsupported type %q.", api.Name, query.Name, query.Type), "Use primitive field types for API query parameters.")
			}
		}
	}
}

func (v *semanticValidator) validateApp() {
	if v.program.App.Name == "" {
		v.addDiagnostic(Position{}, "MISSING_APP", "Project is missing an app declaration.", "Add `app AppName` at the top of the file.")
	}
}

func (v *semanticValidator) validateAuth() {
	if v.program.Auth == nil {
		return
	}
	auth := *v.program.Auth
	if auth.Strategy == "" {
		v.addDiagnostic(auth.Position, "MISSING_AUTH_STRATEGY", "Auth block is missing a strategy.", "Add `strategy emailPassword` inside auth.")
	} else if !supportedAuthStrategies[auth.Strategy] {
		v.addDiagnostic(auth.Position, "UNSUPPORTED_AUTH_STRATEGY", fmt.Sprintf("Auth uses unsupported strategy %q.", auth.Strategy), "Use emailPassword in v0.1.")
	}

	if auth.Session == "" {
		v.addDiagnostic(auth.Position, "MISSING_AUTH_SESSION", "Auth block is missing a session type.", "Add `session cookie` inside auth.")
	} else if !supportedAuthSessions[auth.Session] {
		v.addDiagnostic(auth.Position, "UNSUPPORTED_AUTH_SESSION", fmt.Sprintf("Auth uses unsupported session %q.", auth.Session), "Use cookie in v0.1.")
	}

	if len(auth.User.Fields) == 0 {
		v.addDiagnostic(auth.Position, "MISSING_AUTH_USER", "Auth block is missing user fields.", "Add `user { email email unique }` inside auth.")
		return
	}

	fieldIndex := map[string]FieldDecl{}
	for _, field := range auth.User.Fields {
		if existing, ok := fieldIndex[field.Name]; ok {
			v.addDiagnostic(field.Position, "DUPLICATE_AUTH_USER_FIELD", fmt.Sprintf("Auth user field %s is already defined.", field.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		fieldIndex[field.Name] = field
		if !supportedFieldTypes[field.Type] {
			v.addDiagnostic(field.Position, "UNSUPPORTED_AUTH_USER_FIELD_TYPE", fmt.Sprintf("Auth user field %s uses unsupported type %q.", field.Name, field.Type), "Use primitive field types in auth user for v0.1.")
		}
		for _, modifier := range field.Modifiers {
			if !supportedFieldModifiers[modifier.Name] {
				v.addDiagnostic(field.Position, "UNSUPPORTED_AUTH_USER_FIELD_MODIFIER", fmt.Sprintf("Auth user field %s uses unsupported modifier %q.", field.Name, modifier.Name), "Use required, unique, optional, default, label, placeholder, help, min, max, length, regex, url, or message.")
			}
			if modifier.Name == "default" && modifier.Value == "" {
				v.addDiagnostic(field.Position, "MISSING_DEFAULT_VALUE", fmt.Sprintf("Auth user field %s has default without a value.", field.Name), "Write default followed by a value, such as `default active`.")
			}
			if modifier.Name == "label" && modifier.Value == "" {
				v.addDiagnostic(field.Position, "MISSING_LABEL_VALUE", fmt.Sprintf("Auth user field %s has label without a value.", field.Name), "Write label followed by text, such as `label \"Email\"`.")
			}
			if modifier.Name == "placeholder" && modifier.Value == "" {
				v.addDiagnostic(field.Position, "MISSING_PLACEHOLDER_VALUE", fmt.Sprintf("Auth user field %s has placeholder without a value.", field.Name), "Write placeholder followed by text, such as `placeholder \"you@example.com\"`.")
			}
			if modifier.Name == "help" && modifier.Value == "" {
				v.addDiagnostic(field.Position, "MISSING_HELP_VALUE", fmt.Sprintf("Auth user field %s has help without a value.", field.Name), "Write help followed by text, such as `help \"Used for login\"`.")
			}
			v.validateConstraintModifier("auth user", "", field, modifier)
		}
	}
}

func (v *semanticValidator) validateDatabase() {
	if v.program.Database == nil {
		return
	}
	database := *v.program.Database
	if database.URL.Name == "" {
		v.addDiagnostic(database.Position, "MISSING_DATABASE_URL", "Database block is missing a url environment reference.", "Add `url env DATABASE_URL` inside database.")
		return
	}
	if !validEnvName(database.URL.Name) {
		v.addDiagnostic(database.URL.Position, "INVALID_ENV_NAME", fmt.Sprintf("Database url references invalid environment variable %q.", database.URL.Name), "Use uppercase letters, numbers, and underscores, such as DATABASE_URL.")
	}
}

func validEnvName(value string) bool {
	if value == "" {
		return false
	}
	for index, char := range value {
		if char >= 'A' && char <= 'Z' {
			continue
		}
		if char >= '0' && char <= '9' && index > 0 {
			continue
		}
		if char == '_' && index > 0 {
			continue
		}
		return false
	}
	return true
}

func (v *semanticValidator) validateEntities() map[string]EntityDecl {
	entityIndex := map[string]EntityDecl{}
	for _, entity := range v.program.Entities {
		if existing, ok := entityIndex[entity.Name]; ok {
			v.addDiagnostic(entity.Position, "DUPLICATE_ENTITY", fmt.Sprintf("Entity %s is already defined.", entity.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		entityIndex[entity.Name] = entity
	}
	for _, entity := range v.program.Entities {
		if entityIndex[entity.Name].Name == entity.Name {
			v.validateFields(entity, entityIndex)
		}
	}
	return entityIndex
}

func (v *semanticValidator) validateRoles(entityIndex map[string]EntityDecl) map[string]RoleDecl {
	roleIndex := map[string]RoleDecl{}
	for _, role := range v.program.Roles {
		if existing, ok := roleIndex[role.Name]; ok {
			v.addDiagnostic(role.Position, "DUPLICATE_ROLE", fmt.Sprintf("Role %s is already defined.", role.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		roleIndex[role.Name] = role
	}

	for _, role := range v.program.Roles {
		if roleIndex[role.Name].Name != role.Name {
			continue
		}
		for _, permission := range role.Permissions {
			if permission.Effect != "allow" && permission.Effect != "deny" {
				v.addDiagnostic(permission.Position, "UNSUPPORTED_PERMISSION_EFFECT", fmt.Sprintf("Role %s uses unsupported permission effect %q.", role.Name, permission.Effect), "Use allow or deny.")
			}
			if !supportedPermissionActions[permission.Action] {
				v.addDiagnostic(permission.Position, "UNSUPPORTED_PERMISSION_ACTION", fmt.Sprintf("Role %s uses unsupported permission action %q.", role.Name, permission.Action), "Use all, manage, read, create, update, or delete.")
			}
			if permission.Action != "all" && permission.Resource == "" {
				v.addDiagnostic(permission.Position, "MISSING_PERMISSION_RESOURCE", fmt.Sprintf("Role %s permission %s is missing a resource.", role.Name, permission.Action), "Add an entity name, such as `allow read Product`.")
				continue
			}
			if permission.Resource != "" {
				entity, ok := entityIndex[permission.Resource]
				if !ok {
					v.addDiagnostic(permission.Position, "UNKNOWN_PERMISSION_RESOURCE", fmt.Sprintf("Role %s references unknown resource %s.", role.Name, permission.Resource), "Use an existing entity name as the permission resource.")
					continue
				}
				fieldIndex := map[string]bool{}
				for _, field := range entity.Fields {
					fieldIndex[field.Name] = true
				}
				for _, fieldName := range permission.Fields {
					if !fieldIndex[fieldName] {
						v.addDiagnostic(permission.Position, "UNKNOWN_PERMISSION_FIELD", fmt.Sprintf("Role %s references unknown field %s.%s.", role.Name, permission.Resource, fieldName), "Use an existing field name for field-level permission.")
					}
				}
			}
		}
	}

	return roleIndex
}

func (v *semanticValidator) validateLayouts() map[string]LayoutDecl {
	layoutIndex := map[string]LayoutDecl{}
	for _, layout := range v.program.Layouts {
		if existing, ok := layoutIndex[layout.Name]; ok {
			v.addDiagnostic(layout.Position, "DUPLICATE_LAYOUT", fmt.Sprintf("Layout %s is already defined.", layout.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		layoutIndex[layout.Name] = layout
	}
	return layoutIndex
}

func (v *semanticValidator) validateFields(entity EntityDecl, entityIndex map[string]EntityDecl) {
	fieldIndex := map[string]FieldDecl{}
	for _, field := range entity.Fields {
		if existing, ok := fieldIndex[field.Name]; ok {
			v.addDiagnostic(field.Position, "DUPLICATE_FIELD", fmt.Sprintf("Field %s.%s is already defined.", entity.Name, field.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		fieldIndex[field.Name] = field

		if !supportedFieldTypes[field.Type] {
			if _, ok := entityIndex[field.Type]; !ok {
				v.addDiagnostic(field.Position, "UNSUPPORTED_FIELD_TYPE", fmt.Sprintf("Field %s.%s uses unsupported type %q.", entity.Name, field.Name, field.Type), "Use a primitive type or the name of an existing entity.")
			}
		}

		for _, modifier := range field.Modifiers {
			if !supportedFieldModifiers[modifier.Name] {
				v.addDiagnostic(field.Position, "UNSUPPORTED_FIELD_MODIFIER", fmt.Sprintf("Field %s.%s uses unsupported modifier %q.", entity.Name, field.Name, modifier.Name), "Use required, unique, optional, default, label, placeholder, help, min, max, length, regex, url, or message.")
			}
			if modifier.Name == "default" && modifier.Value == "" {
				v.addDiagnostic(field.Position, "MISSING_DEFAULT_VALUE", fmt.Sprintf("Field %s.%s has default without a value.", entity.Name, field.Name), "Write default followed by a value, such as `default 0`.")
			}
			if modifier.Name == "label" && modifier.Value == "" {
				v.addDiagnostic(field.Position, "MISSING_LABEL_VALUE", fmt.Sprintf("Field %s.%s has label without a value.", entity.Name, field.Name), "Write label followed by text, such as `label \"Product Name\"`.")
			}
			if modifier.Name == "placeholder" && modifier.Value == "" {
				v.addDiagnostic(field.Position, "MISSING_PLACEHOLDER_VALUE", fmt.Sprintf("Field %s.%s has placeholder without a value.", entity.Name, field.Name), "Write placeholder followed by text, such as `placeholder \"Enter product name\"`.")
			}
			if modifier.Name == "help" && modifier.Value == "" {
				v.addDiagnostic(field.Position, "MISSING_HELP_VALUE", fmt.Sprintf("Field %s.%s has help without a value.", entity.Name, field.Name), "Write help followed by text, such as `help \"Shown under the input\"`.")
			}
			v.validateConstraintModifier("field", entity.Name, field, modifier)
		}
	}
	v.validateEntityValidations(entity, fieldIndex)
}

func (v *semanticValidator) validateEntityValidations(entity EntityDecl, fieldIndex map[string]FieldDecl) {
	validationIndex := map[string]EntityValidationDecl{}
	for _, validation := range entity.Validations {
		key := entityValidationKey(validation)
		if existing, ok := validationIndex[key]; ok {
			v.addDiagnostic(validation.Position, "DUPLICATE_ENTITY_VALIDATION", fmt.Sprintf("Entity %s repeats validation %s.", entity.Name, key), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		validationIndex[key] = validation

		if validation.Required {
			if _, ok := fieldIndex[validation.Left]; !ok {
				v.addDiagnostic(validation.Position, "UNKNOWN_VALIDATION_FIELD", fmt.Sprintf("Entity %s validation references unknown field %s.", entity.Name, validation.Left), "Use fields defined on the same entity.")
			}
			if validation.When == nil {
				v.addDiagnostic(validation.Position, "MISSING_VALIDATION_CONDITION", fmt.Sprintf("Entity %s conditional validation is missing a when condition.", entity.Name), "Use `validate field required when otherField == value`.")
				continue
			}
			v.validateValidationCondition(entity, fieldIndex, *validation.When)
			continue
		}

		left, leftOK := fieldIndex[validation.Left]
		if !leftOK {
			v.addDiagnostic(validation.Position, "UNKNOWN_VALIDATION_FIELD", fmt.Sprintf("Entity %s validation references unknown field %s.", entity.Name, validation.Left), "Use fields defined on the same entity.")
		}
		right, rightOK := fieldIndex[validation.Right]
		if !rightOK {
			v.addDiagnostic(validation.Position, "UNKNOWN_VALIDATION_FIELD", fmt.Sprintf("Entity %s validation references unknown field %s.", entity.Name, validation.Right), "Use fields defined on the same entity.")
		}
		if !supportedComparisonOperators[validation.Operator] {
			v.addDiagnostic(validation.Position, "UNSUPPORTED_VALIDATION_OPERATOR", fmt.Sprintf("Entity %s validation uses unsupported operator %q.", entity.Name, validation.Operator), "Use ==, !=, <, <=, >, or >=.")
		}
		if leftOK && rightOK {
			if numberLikeField(left) && numberLikeField(right) {
				continue
			}
			if left.Type == right.Type && (validation.Operator == "==" || validation.Operator == "!=") {
				continue
			}
			v.addDiagnostic(validation.Position, "INCOMPATIBLE_VALIDATION_FIELDS", fmt.Sprintf("Entity %s validation compares incompatible fields %s and %s.", entity.Name, validation.Left, validation.Right), "Use number-like fields for ordering comparisons, or same-type fields for equality checks.")
		}
	}
}

func entityValidationKey(validation EntityValidationDecl) string {
	if validation.Required && validation.When != nil {
		return validation.Left + " required when " + validation.When.Left + " " + validation.When.Operator + " " + validation.When.Right
	}
	return validation.Left + " " + validation.Operator + " " + validation.Right
}

func (v *semanticValidator) validateValidationCondition(entity EntityDecl, fieldIndex map[string]FieldDecl, condition ValidationConditionDecl) {
	left, leftOK := fieldIndex[condition.Left]
	if !leftOK {
		v.addDiagnostic(condition.Position, "UNKNOWN_VALIDATION_FIELD", fmt.Sprintf("Entity %s validation condition references unknown field %s.", entity.Name, condition.Left), "Use fields defined on the same entity.")
	}
	if !supportedComparisonOperators[condition.Operator] {
		v.addDiagnostic(condition.Position, "UNSUPPORTED_VALIDATION_OPERATOR", fmt.Sprintf("Entity %s validation condition uses unsupported operator %q.", entity.Name, condition.Operator), "Use ==, !=, <, <=, >, or >=.")
	}
	if !leftOK || !supportedComparisonOperators[condition.Operator] {
		return
	}
	if conditionRightField, ok := fieldIndex[condition.Right]; ok {
		if numberLikeField(left) && numberLikeField(conditionRightField) {
			return
		}
		if left.Type == conditionRightField.Type && (condition.Operator == "==" || condition.Operator == "!=") {
			return
		}
		v.addDiagnostic(condition.Position, "INCOMPATIBLE_VALIDATION_FIELDS", fmt.Sprintf("Entity %s validation condition compares incompatible fields %s and %s.", entity.Name, condition.Left, condition.Right), "Use number-like fields for ordering comparisons, or same-type fields for equality checks.")
		return
	}
	if numberLikeField(left) && (condition.Operator == "<" || condition.Operator == "<=" || condition.Operator == ">" || condition.Operator == ">=") {
		if _, err := strconv.ParseFloat(condition.Right, 64); err != nil {
			v.addDiagnostic(condition.Position, "INVALID_VALIDATION_LITERAL", fmt.Sprintf("Entity %s validation condition has invalid numeric value %q.", entity.Name, condition.Right), "Use a numeric literal or another number-like field.")
		}
		return
	}
	if condition.Operator != "==" && condition.Operator != "!=" && !numberLikeField(left) {
		v.addDiagnostic(condition.Position, "INCOMPATIBLE_VALIDATION_FIELDS", fmt.Sprintf("Entity %s validation condition uses ordering operator on non-numeric field %s.", entity.Name, condition.Left), "Use == or != for text-like conditions.")
	}
}

func (v *semanticValidator) validateConstraintModifier(scope string, entityName string, field FieldDecl, modifier Modifier) {
	fieldName := field.Name
	if entityName != "" {
		fieldName = entityName + "." + field.Name
	}
	switch modifier.Name {
	case "min", "max":
		if modifier.Value == "" {
			v.addDiagnostic(field.Position, "MISSING_CONSTRAINT_VALUE", fmt.Sprintf("%s %s has %s without a value.", title(scope), fieldName, modifier.Name), fmt.Sprintf("Write %s followed by a value, such as `%s 0`.", modifier.Name, modifier.Name))
			return
		}
		if !numberLikeField(field) {
			v.addDiagnostic(field.Position, "UNSUPPORTED_NUMERIC_CONSTRAINT", fmt.Sprintf("%s %s uses %s on non-numeric type %s.", title(scope), fieldName, modifier.Name, field.Type), "Use min/max on number, integer, decimal, or money fields.")
			return
		}
		if _, err := strconv.ParseFloat(modifier.Value, 64); err != nil {
			v.addDiagnostic(field.Position, "INVALID_NUMERIC_CONSTRAINT", fmt.Sprintf("%s %s has invalid %s value %q.", title(scope), fieldName, modifier.Name, modifier.Value), fmt.Sprintf("Use a numeric %s value, such as `%s 0`.", modifier.Name, modifier.Name))
		}
	case "length":
		if modifier.Value == "" {
			v.addDiagnostic(field.Position, "MISSING_CONSTRAINT_VALUE", fmt.Sprintf("%s %s has length without a value.", title(scope), fieldName), "Write length followed by a range, such as `length 3..40`.")
			return
		}
		if field.Type != "text" && field.Type != "email" {
			v.addDiagnostic(field.Position, "UNSUPPORTED_LENGTH_CONSTRAINT", fmt.Sprintf("%s %s uses length on non-text type %s.", title(scope), fieldName, field.Type), "Use length on text or email fields.")
			return
		}
		minLength, maxLength, ok := parseLengthConstraint(modifier.Value)
		if !ok || minLength < 0 || maxLength < minLength {
			v.addDiagnostic(field.Position, "INVALID_LENGTH_CONSTRAINT", fmt.Sprintf("%s %s has invalid length value %q.", title(scope), fieldName, modifier.Value), "Use a range such as `length 3..40`.")
		}
	case "regex":
		if modifier.Value == "" {
			v.addDiagnostic(field.Position, "MISSING_CONSTRAINT_VALUE", fmt.Sprintf("%s %s has regex without a value.", title(scope), fieldName), "Write regex followed by a quoted pattern, such as `regex \"^[A-Z0-9]+$\"`.")
			return
		}
		if field.Type != "text" && field.Type != "email" {
			v.addDiagnostic(field.Position, "UNSUPPORTED_REGEX_CONSTRAINT", fmt.Sprintf("%s %s uses regex on non-text type %s.", title(scope), fieldName, field.Type), "Use regex on text or email fields.")
			return
		}
		if _, err := regexp.Compile(modifier.Value); err != nil {
			v.addDiagnostic(field.Position, "INVALID_REGEX_CONSTRAINT", fmt.Sprintf("%s %s has invalid regex value %q.", title(scope), fieldName, modifier.Value), "Use a valid regular expression pattern.")
		}
	case "url":
		if field.Type != "text" {
			v.addDiagnostic(field.Position, "UNSUPPORTED_URL_CONSTRAINT", fmt.Sprintf("%s %s uses url on non-text type %s.", title(scope), fieldName, field.Type), "Use url on text fields that store web addresses.")
		}
	case "message":
		if modifier.Value == "" {
			v.addDiagnostic(field.Position, "MISSING_MESSAGE_VALUE", fmt.Sprintf("%s %s has message without a value.", title(scope), fieldName), "Write message followed by text, such as `message \"Enter a valid SKU\"`.")
		}
	}
}

func numberLikeField(field FieldDecl) bool {
	return field.Type == "number" || field.Type == "integer" || field.Type == "decimal" || field.Type == "money"
}

func parseLengthConstraint(value string) (int, int, bool) {
	parts := strings.Split(value, "..")
	if len(parts) != 2 {
		return 0, 0, false
	}
	minLength, minErr := strconv.Atoi(parts[0])
	maxLength, maxErr := strconv.Atoi(parts[1])
	if minErr != nil || maxErr != nil {
		return 0, 0, false
	}
	return minLength, maxLength, true
}

func apiPathParamNames(path string) []string {
	matches := regexp.MustCompile(`\{([A-Za-z][A-Za-z0-9_]*)\}`).FindAllStringSubmatch(path, -1)
	names := []string{}
	for _, match := range matches {
		if len(match) == 2 {
			names = append(names, match[1])
		}
	}
	return names
}

func (v *semanticValidator) validatePages(entityIndex map[string]EntityDecl, layoutIndex map[string]LayoutDecl, roleIndex map[string]RoleDecl) map[string]PageDecl {
	pageIndex := map[string]PageDecl{}
	for _, page := range v.program.Pages {
		if existing, ok := pageIndex[page.Name]; ok {
			v.addDiagnostic(page.Position, "DUPLICATE_PAGE", fmt.Sprintf("Page %s is already defined.", page.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		pageIndex[page.Name] = page

		if page.Layout != "" {
			if _, ok := layoutIndex[page.Layout]; !ok {
				v.addDiagnostic(page.Position, "UNKNOWN_PAGE_LAYOUT", fmt.Sprintf("Page %s uses unknown layout %s.", page.Name, page.Layout), "Create the layout or remove the page layout reference.")
			}
		}

		source, ok := entityIndex[page.Source]
		if page.Source == "" {
			v.addDiagnostic(page.Position, "MISSING_PAGE_SOURCE", fmt.Sprintf("Page %s is missing a source entity.", page.Name), "Add `source EntityName` inside the page.")
			continue
		}
		if !ok {
			v.addDiagnostic(page.Position, "UNKNOWN_SOURCE_ENTITY", fmt.Sprintf("Page %s uses unknown source entity %s.", page.Name, page.Source), "Create the entity or change the page source.")
			continue
		}

		v.validatePageFields(page, source)
		v.validateActions(page)
		v.validatePageAccess(page, roleIndex)
	}
	return pageIndex
}

func (v *semanticValidator) validatePageAccess(page PageDecl, roleIndex map[string]RoleDecl) {
	for _, access := range page.Access {
		if v.program.Auth == nil {
			v.addDiagnostic(page.Position, "AUTH_REQUIRED_FOR_ACCESS", fmt.Sprintf("Page %s uses access control without an auth block.", page.Name), "Add an auth block or remove page access.")
			continue
		}
		if access == "authenticated" {
			continue
		}
		if _, ok := roleIndex[access]; !ok {
			v.addDiagnostic(page.Position, "UNKNOWN_ACCESS_ROLE", fmt.Sprintf("Page %s references unknown access role %s.", page.Name, access), "Create the role or remove it from access.")
		}
	}
}

func (v *semanticValidator) validateLayoutReferences(layoutIndex map[string]LayoutDecl, pageIndex map[string]PageDecl) {
	for _, layout := range layoutIndex {
		seen := map[string]bool{}
		for _, item := range layout.Sidebar.Items {
			if seen[item] {
				v.addDiagnostic(layout.Position, "DUPLICATE_SIDEBAR_ITEM", fmt.Sprintf("Layout %s sidebar includes page %s more than once.", layout.Name, item), "Keep each sidebar page item once.")
				continue
			}
			seen[item] = true
			if _, ok := pageIndex[item]; !ok {
				v.addDiagnostic(layout.Position, "UNKNOWN_SIDEBAR_ITEM", fmt.Sprintf("Layout %s sidebar references unknown page %s.", layout.Name, item), "Create the page or remove it from the sidebar.")
			}
		}
	}
}

func (v *semanticValidator) validatePageFields(page PageDecl, source EntityDecl) {
	fields := map[string]FieldDecl{}
	for _, field := range source.Fields {
		fields[field.Name] = field
	}

	for _, column := range page.Table.Columns {
		if _, ok := fields[column]; !ok {
			v.addDiagnostic(page.Position, "UNKNOWN_TABLE_COLUMN", fmt.Sprintf("Page %s table uses unknown field %s.%s.", page.Name, source.Name, column), "Add the field to the source entity or remove it from columns.")
		}
	}

	for _, fieldName := range page.Form.Fields {
		if _, ok := fields[fieldName]; !ok {
			v.addDiagnostic(page.Position, "UNKNOWN_FORM_FIELD", fmt.Sprintf("Page %s form uses unknown field %s.%s.", page.Name, source.Name, fieldName), "Add the field to the source entity or remove it from fields.")
		}
	}

	for _, fieldName := range page.Table.Search {
		field, ok := fields[fieldName]
		if !ok {
			v.addDiagnostic(page.Position, "UNKNOWN_SEARCH_FIELD", fmt.Sprintf("Page %s search uses unknown field %s.%s.", page.Name, source.Name, fieldName), "Add the field to the source entity or remove it from search.")
			continue
		}
		if !searchableFieldTypes[field.Type] {
			if _, ok := v.findEntity(field.Type); !ok {
				v.addDiagnostic(field.Position, "UNSEARCHABLE_FIELD_TYPE", fmt.Sprintf("Field %s.%s with type %s cannot be searched in v0.1.", source.Name, field.Name, field.Type), "Use text, email, or entity reference fields for search in v0.1.")
			}
		}
	}

	for _, fieldName := range page.Table.Filters {
		if _, ok := fields[fieldName]; !ok {
			v.addDiagnostic(page.Position, "UNKNOWN_FILTER_FIELD", fmt.Sprintf("Page %s filter uses unknown field %s.%s.", page.Name, source.Name, fieldName), "Add the field to the source entity or remove it from filter.")
		}
	}

	if page.Table.Sort.Field != "" {
		if _, ok := fields[page.Table.Sort.Field]; !ok {
			v.addDiagnostic(page.Position, "UNKNOWN_SORT_FIELD", fmt.Sprintf("Page %s table sorts by unknown field %s.%s.", page.Name, source.Name, page.Table.Sort.Field), "Add the field to the source entity or change the sort field.")
		}
		if page.Table.Sort.Direction != "asc" && page.Table.Sort.Direction != "desc" {
			v.addDiagnostic(page.Position, "UNSUPPORTED_SORT_DIRECTION", fmt.Sprintf("Page %s table uses unsupported sort direction %q.", page.Name, page.Table.Sort.Direction), "Use asc or desc.")
		}
	}

	if page.Table.Paginate < 0 {
		v.addDiagnostic(page.Position, "UNSUPPORTED_PAGE_SIZE", fmt.Sprintf("Page %s table uses unsupported page size %d.", page.Name, page.Table.Paginate), "Use a positive whole number, such as `paginate 25`.")
	}
}

func (v *semanticValidator) validateActions(page PageDecl) {
	for _, action := range page.Actions {
		if !supportedActions[action] {
			v.addDiagnostic(page.Position, "UNSUPPORTED_ACTION", fmt.Sprintf("Page %s uses unsupported action %q.", page.Name, action), "Use create, edit, delete, archive, or restore in v0.1.")
		}
	}
}

func (v *semanticValidator) validateWorkflows(entityIndex map[string]EntityDecl, roleIndex map[string]RoleDecl) {
	workflowIndex := map[string]WorkflowDecl{}
	for _, workflow := range v.program.Workflows {
		if existing, ok := workflowIndex[workflow.Name]; ok {
			v.addDiagnostic(workflow.Position, "DUPLICATE_WORKFLOW", fmt.Sprintf("Workflow %s is already defined.", workflow.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		workflowIndex[workflow.Name] = workflow
	}

	for _, workflow := range v.program.Workflows {
		if workflowIndex[workflow.Name].Name != workflow.Name {
			continue
		}
		if workflow.Source == "" {
			v.addDiagnostic(workflow.Position, "MISSING_WORKFLOW_SOURCE", fmt.Sprintf("Workflow %s is missing a source entity.", workflow.Name), "Add `source EntityName` inside the workflow.")
		} else if source, ok := entityIndex[workflow.Source]; !ok {
			v.addDiagnostic(workflow.Position, "UNKNOWN_WORKFLOW_SOURCE", fmt.Sprintf("Workflow %s uses unknown source entity %s.", workflow.Name, workflow.Source), "Use an existing entity as the workflow source.")
		} else {
			statusField, hasStatusField := findField(source, "status")
			if !hasStatusField {
				v.addDiagnostic(workflow.Position, "MISSING_WORKFLOW_STATUS_FIELD", fmt.Sprintf("Workflow %s source entity %s has no status field.", workflow.Name, workflow.Source), "Add `status text default draft` to the source entity.")
			} else if statusField.Type != "text" {
				v.addDiagnostic(workflow.Position, "UNSUPPORTED_WORKFLOW_STATUS_FIELD_TYPE", fmt.Sprintf("Workflow %s source entity %s uses non-text status field.", workflow.Name, workflow.Source), "Use `status text` for workflow state storage in v0.1.")
			}
		}
		if len(workflow.States) == 0 {
			v.addDiagnostic(workflow.Position, "MISSING_WORKFLOW_STATES", fmt.Sprintf("Workflow %s has no states.", workflow.Name), "Add `states draft, active, done` inside the workflow.")
		}

		stateIndex := map[string]bool{}
		for _, state := range workflow.States {
			if stateIndex[state] {
				v.addDiagnostic(workflow.Position, "DUPLICATE_WORKFLOW_STATE", fmt.Sprintf("Workflow %s state %s is declared more than once.", workflow.Name, state), "Keep each workflow state once.")
				continue
			}
			stateIndex[state] = true
		}

		transitionIndex := map[string]TransitionDecl{}
		for _, transition := range workflow.Transitions {
			if existing, ok := transitionIndex[transition.Name]; ok {
				v.addDiagnostic(transition.Position, "DUPLICATE_WORKFLOW_TRANSITION", fmt.Sprintf("Workflow %s transition %s is already defined.", workflow.Name, transition.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
				continue
			}
			transitionIndex[transition.Name] = transition
		}

		for _, transition := range workflow.Transitions {
			if transitionIndex[transition.Name].Name != transition.Name {
				continue
			}
			if transition.From == "" {
				v.addDiagnostic(transition.Position, "MISSING_TRANSITION_FROM", fmt.Sprintf("Workflow %s transition %s is missing from state.", workflow.Name, transition.Name), "Add `from StateName` inside the transition.")
			} else if !stateIndex[transition.From] {
				v.addDiagnostic(transition.Position, "UNKNOWN_TRANSITION_FROM", fmt.Sprintf("Workflow %s transition %s uses unknown from state %s.", workflow.Name, transition.Name, transition.From), "Use a state declared in the workflow states list.")
			}
			if transition.To == "" {
				v.addDiagnostic(transition.Position, "MISSING_TRANSITION_TO", fmt.Sprintf("Workflow %s transition %s is missing to state.", workflow.Name, transition.Name), "Add `to StateName` inside the transition.")
			} else if !stateIndex[transition.To] {
				v.addDiagnostic(transition.Position, "UNKNOWN_TRANSITION_TO", fmt.Sprintf("Workflow %s transition %s uses unknown to state %s.", workflow.Name, transition.Name, transition.To), "Use a state declared in the workflow states list.")
			}
			for _, allowed := range transition.Allow {
				if v.program.Auth == nil {
					v.addDiagnostic(transition.Position, "AUTH_REQUIRED_FOR_WORKFLOW_ALLOW", fmt.Sprintf("Workflow %s transition %s uses allow without an auth block.", workflow.Name, transition.Name), "Add an auth block or remove transition allow.")
					continue
				}
				if allowed == "authenticated" {
					continue
				}
				if _, ok := roleIndex[allowed]; !ok {
					v.addDiagnostic(transition.Position, "UNKNOWN_WORKFLOW_ALLOW_ROLE", fmt.Sprintf("Workflow %s transition %s references unknown role %s.", workflow.Name, transition.Name, allowed), "Use an existing role or authenticated in transition allow.")
				}
			}
		}
	}
}

func (v *semanticValidator) validateStates(entityIndex map[string]EntityDecl) {
	stateIndex := map[string]StateDecl{}
	for _, state := range v.program.States {
		if existing, ok := stateIndex[state.Name]; ok {
			v.addDiagnostic(state.Position, "DUPLICATE_STATE", fmt.Sprintf("State %s is already defined.", state.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		stateIndex[state.Name] = state
	}

	for _, state := range v.program.States {
		if stateIndex[state.Name].Name != state.Name {
			continue
		}

		fieldIndex := map[string]StateField{}
		for _, field := range state.Fields {
			if existing, ok := fieldIndex[field.Name]; ok {
				v.addDiagnostic(field.Position, "DUPLICATE_STATE_FIELD", fmt.Sprintf("State %s field %s is already defined.", state.Name, field.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
				continue
			}
			fieldIndex[field.Name] = field
			if !supportedFieldTypes[field.Type] {
				if _, ok := entityIndex[field.Type]; !ok {
					v.addDiagnostic(field.Position, "UNSUPPORTED_STATE_FIELD_TYPE", fmt.Sprintf("State %s field %s uses unsupported type %q.", state.Name, field.Name, field.Type), "Use a primitive type or an existing entity type. Use Entity[] for lists.")
				}
			}
		}

		modalIndex := map[string]StateModal{}
		for _, modal := range state.Modals {
			if existing, ok := modalIndex[modal.Name]; ok {
				v.addDiagnostic(modal.Position, "DUPLICATE_STATE_MODAL", fmt.Sprintf("State %s modal %s is already defined.", state.Name, modal.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
				continue
			}
			modalIndex[modal.Name] = modal
			if modal.Default != "open" && modal.Default != "closed" {
				v.addDiagnostic(modal.Position, "UNSUPPORTED_STATE_MODAL_DEFAULT", fmt.Sprintf("State %s modal %s uses unsupported default %q.", state.Name, modal.Name, modal.Default), "Use open or closed.")
			}
		}
	}
}

func (v *semanticValidator) validateComponents(entityIndex map[string]EntityDecl) {
	componentIndex := map[string]ComponentDecl{}
	for _, component := range v.program.Components {
		if existing, ok := componentIndex[component.Name]; ok {
			v.addDiagnostic(component.Position, "DUPLICATE_COMPONENT", fmt.Sprintf("Component %s is already defined.", component.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		componentIndex[component.Name] = component
	}

	for _, component := range v.program.Components {
		if componentIndex[component.Name].Name != component.Name {
			continue
		}

		inputIndex := map[string]ComponentInput{}
		for _, input := range component.Inputs {
			if existing, ok := inputIndex[input.Name]; ok {
				v.addDiagnostic(input.Position, "DUPLICATE_COMPONENT_INPUT", fmt.Sprintf("Component %s input %s is already defined.", component.Name, input.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
				continue
			}
			inputIndex[input.Name] = input
			if !supportedFieldTypes[input.Type] {
				if _, ok := entityIndex[input.Type]; !ok {
					v.addDiagnostic(input.Position, "UNSUPPORTED_COMPONENT_INPUT_TYPE", fmt.Sprintf("Component %s input %s uses unsupported type %q.", component.Name, input.Name, input.Type), "Use a primitive type or an existing entity type. Use Entity[] for lists.")
				}
			}
		}

		variantIndex := map[string]ComponentVariant{}
		for _, variant := range component.Variants {
			if existing, ok := variantIndex[variant.Name]; ok {
				v.addDiagnostic(variant.Position, "DUPLICATE_COMPONENT_VARIANT", fmt.Sprintf("Component %s variant %s is already defined.", component.Name, variant.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
				continue
			}
			variantIndex[variant.Name] = variant
			if variant.Condition == "" {
				v.addDiagnostic(variant.Position, "MISSING_COMPONENT_VARIANT_CONDITION", fmt.Sprintf("Component %s variant %s is missing a condition.", component.Name, variant.Name), "Add `when condition`, such as `when stock < 10`.")
			}
		}
	}
}

func (v *semanticValidator) findEntity(name string) (EntityDecl, bool) {
	for _, entity := range v.program.Entities {
		if entity.Name == name {
			return entity, true
		}
	}
	return EntityDecl{}, false
}

func (v *semanticValidator) addDiagnostic(position Position, code string, message string, suggestion string) {
	v.diagnostics = append(v.diagnostics, Diagnostic{
		File:       position.File,
		Line:       position.Line,
		Column:     position.Column,
		Code:       code,
		Message:    message,
		Suggestion: suggestion,
	})
}

func setOf(values ...string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}
