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
	"file",
	"image",
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
	"accept",
	"message",
	"load",
)

var searchableFieldTypes = setOf(
	"text",
	"email",
)

var supportedComputedFieldTypes = setOf(
	"number",
	"integer",
	"decimal",
	"money",
)

var supportedComputedOperators = setOf(
	"+",
	"-",
	"*",
	"/",
)

var supportedComputedFieldModifiers = setOf(
	"label",
	"help",
)

var supportedActions = setOf(
	"create",
	"edit",
	"delete",
	"archive",
	"restore",
)

var supportedInlineUIModes = setOf(
	"box",
	"text",
	"table",
	"button",
)

var supportedViewSections = setOf(
	"table",
	"detail",
	"form",
)

var supportedViewSectionDisplays = setOf(
	"inline",
	"modal",
	"drawer",
)

var supportedViewSectionSides = setOf(
	"left",
	"right",
)

var supportedViewComponentBinds = setOf(
	"selected",
	"first",
	"each",
)

var supportedViewGroupComposeModes = setOf(
	"stack",
	"grid",
)

var supportedViewComposeModes = setOf(
	"stack",
	"grid",
	"tabs",
)

var supportedViewTriggerEvents = setOf(
	"rowSelect",
	"createStart",
	"editStart",
	"saveSuccess",
	"close",
)

var supportedViewGaps = setOf(
	"sm",
	"md",
	"lg",
)

var supportedViewBreakpoints = setOf(
	"sm",
	"md",
	"lg",
	"none",
)

var supportedAuthStrategies = setOf(
	"emailPassword",
)

var supportedAuthSessions = setOf(
	"cookie",
)

var supportedTargetNames = setOf(
	"api",
	"web",
)

var supportedTargetFrontends = setOf(
	"react",
)

var supportedTargetBackends = setOf(
	"node",
)

var supportedTargetDatabases = setOf(
	"mysql",
	"postgres",
	"sqlite",
)

var supportedDeployTargets = setOf(
	"docker",
)

var supportedDeployEnvModes = setOf(
	"required",
	"optional",
)

var supportedDeployPreviewModes = setOf(
	"local",
)

var supportedDeployRollbackStrategies = setOf(
	"keep",
)

var supportedDeployCloudProviders = setOf(
	"fly",
	"railway",
	"render",
)

var supportedOpsLoggingModes = setOf(
	"requests",
)

var supportedOpsObserveProviders = setOf(
	"otlp",
	"webhook",
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
	v.validateTarget()
	v.validateAuth()
	v.validateDatabase()
	v.validateSecurity()
	v.validateDeploy()
	v.validateOps()
	entityIndex := v.validateEntities()
	v.validateMigrations(entityIndex)
	v.validateSeeds(entityIndex)
	v.validateQueries(entityIndex)
	v.validateJobs(entityIndex)
	v.validateI18N()
	v.validateLabelTranslations(entityIndex)
	v.validateFieldTextTranslations(entityIndex)
	roleIndex := v.validateRoles(entityIndex)
	actionIndex := v.validateCustomActions(entityIndex, roleIndex)
	apiIndex := v.validateAPIs(entityIndex)
	v.validateServices(apiIndex)
	v.validateTransactions(actionIndex, apiIndex)
	layoutIndex := v.validateLayouts()
	componentIndex := indexComponents(v.program.Components)
	pageIndex := v.validatePages(entityIndex, layoutIndex, roleIndex, actionIndex, componentIndex)
	v.validateTests(pageIndex)
	v.validateWorkflows(entityIndex, roleIndex)
	v.validateStates(entityIndex)
	v.validateComponents(entityIndex)
	v.validateLayoutReferences(layoutIndex, pageIndex)
}

func (v *semanticValidator) validateAPIs(entityIndex map[string]EntityDecl) map[string]APIDecl {
	apiIndex := map[string]APIDecl{}
	routeIndex := map[string]APIDecl{}
	generatedRoutes := generatedAPIRouteShapes(v.program)
	for _, api := range v.program.APIs {
		if existing, ok := apiIndex[api.Name]; ok {
			v.addDiagnostic(api.Position, "DUPLICATE_API", fmt.Sprintf("API %s is already defined.", api.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		apiIndex[api.Name] = api

		methodOK := false
		if api.Method == "" {
			v.addDiagnostic(api.Position, "MISSING_API_METHOD", fmt.Sprintf("API %s is missing a method.", api.Name), "Add `method GET`, `method POST`, `method PUT`, `method PATCH`, or `method DELETE`.")
		} else if !supportedAPIMethods[strings.ToUpper(api.Method)] {
			v.addDiagnostic(api.Position, "UNSUPPORTED_API_METHOD", fmt.Sprintf("API %s uses unsupported method %q.", api.Name, api.Method), "Use GET, POST, PUT, PATCH, or DELETE.")
		} else {
			methodOK = true
		}

		pathOK := false
		if api.Path == "" {
			v.addDiagnostic(api.Position, "MISSING_API_PATH", fmt.Sprintf("API %s is missing a path.", api.Name), "Add `path \"/api/name\"`.")
		} else if !validExplicitAPIPath(api.Path) {
			v.addDiagnostic(api.Position, "INVALID_API_PATH", fmt.Sprintf("API %s path must be a safe `/api/...` path.", api.Name), "Use a stable API path such as `/api/reports/low-stock/{warehouseId}`. The `/api/auth` namespace is reserved.")
		} else {
			pathOK = true
		}

		if methodOK && pathOK {
			routeKey := strings.ToUpper(api.Method) + " " + normalizedAPIPathShape(api.Path)
			if generated, ok := generatedRoutes[routeKey]; ok {
				v.addDiagnostic(api.Position, "DUPLICATE_API_ROUTE", fmt.Sprintf("API %s conflicts with generated route %s.", api.Name, routeKey), generated)
			} else if existing, ok := routeIndex[routeKey]; ok {
				v.addDiagnostic(api.Position, "DUPLICATE_API_ROUTE", fmt.Sprintf("API %s reuses route %s.", api.Name, routeKey), fmt.Sprintf("First definition is API %s at %s:%d.", existing.Name, existing.Position.File, existing.Position.Line))
			} else {
				routeIndex[routeKey] = api
			}
		}

		if api.Access != "" && api.Access != "public" && api.Access != "private" {
			v.addDiagnostic(api.Position, "UNSUPPORTED_API_ACCESS", fmt.Sprintf("API %s uses unsupported access %q.", api.Name, api.Access), "Use public or private.")
		}

		pathParams := apiPathParamNames(api.Path)
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

		for _, pathParam := range pathParams {
			if _, ok := paramIndex[pathParam]; !ok {
				v.addDiagnostic(api.Position, "MISSING_API_PATH_PARAM", fmt.Sprintf("API %s path uses {%s} without a matching param declaration.", api.Name, pathParam), fmt.Sprintf("Add `param %s text` inside the api block.", pathParam))
			}
		}
		for _, param := range api.Params {
			if !containsString(pathParams, param.Name) {
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
		bodyIndex := v.validateAPIBody(api)
		v.validateAPIRespond(api)
		v.validateAPIUpdate(api, entityIndex, paramIndex, bodyIndex)
	}
	return apiIndex
}

func (v *semanticValidator) validateI18N() {
	if v.program.I18N == nil {
		return
	}

	i18n := v.program.I18N
	if i18n.Default == "" {
		v.addDiagnostic(i18n.Position, "MISSING_I18N_DEFAULT", "I18n declaration is missing a default locale.", "Add `default tr` inside i18n.")
	} else if !isThemeIdentifier(i18n.Default) {
		v.addDiagnostic(i18n.Position, "INVALID_LOCALE", fmt.Sprintf("I18n default locale %q is invalid.", i18n.Default), "Use locale names such as tr, en, or en-US.")
	}

	if len(i18n.Locales) == 0 {
		v.addDiagnostic(i18n.Position, "MISSING_I18N_LOCALES", "I18n declaration is missing locales.", "Add `locales tr, en` inside i18n.")
		return
	}

	locales := map[string]Position{}
	for _, locale := range i18n.Locales {
		if !isThemeIdentifier(locale) {
			v.addDiagnostic(i18n.Position, "INVALID_LOCALE", fmt.Sprintf("I18n locale %q is invalid.", locale), "Use locale names such as tr, en, or en-US.")
			continue
		}
		if existing, ok := locales[locale]; ok {
			v.addDiagnostic(i18n.Position, "DUPLICATE_LOCALE", fmt.Sprintf("I18n locale %s is declared more than once.", locale), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
			continue
		}
		locales[locale] = i18n.Position
	}

	if i18n.Default != "" && locales[i18n.Default].Line == 0 {
		v.addDiagnostic(i18n.Position, "UNKNOWN_DEFAULT_LOCALE", fmt.Sprintf("I18n default locale %s is not listed in locales.", i18n.Default), "Add the default locale to the locales list.")
	}
}

func (v *semanticValidator) validateLabelTranslations(entityIndex map[string]EntityDecl) {
	if len(v.program.Labels) == 0 {
		return
	}
	if v.program.I18N == nil {
		v.addDiagnostic(v.program.Labels[0].Position, "MISSING_I18N", "Label translations require an i18n block.", "Add `i18n { default tr locales tr, en }` before label translation blocks.")
	}

	locales := map[string]bool{}
	defaultLocale := ""
	if v.program.I18N != nil {
		defaultLocale = v.program.I18N.Default
		for _, locale := range v.program.I18N.Locales {
			locales[locale] = true
		}
	}

	targets := map[string]LabelTranslationDecl{}
	for _, label := range v.program.Labels {
		if existing, ok := targets[label.Target]; ok {
			v.addDiagnostic(label.Position, "DUPLICATE_LABEL_TARGET", fmt.Sprintf("Label target %s is already defined.", label.Target), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		targets[label.Target] = label

		if uiLabelTarget, targetOK := isUILabelTarget(v.program, entityIndex, label.Target); targetOK {
			_ = uiLabelTarget
		} else {
			entityName, fieldName, ok := splitLabelTarget(label.Target)
			if !ok {
				v.addDiagnostic(label.Position, "INVALID_LABEL_TARGET", fmt.Sprintf("Label target %s is invalid.", label.Target), "Use `label Entity.field { ... }`, `label app.title { ... }`, `label page.Products { ... }`, or `label action.create { ... }`.")
			} else if entity, exists := entityIndex[entityName]; !exists {
				v.addDiagnostic(label.Position, "UNKNOWN_LABEL_TARGET", fmt.Sprintf("Label target %s references unknown entity or UI label key %s.", label.Target, entityName), "Use an existing entity field or a supported UI label target such as app.title, page.Products, action.create, table.search, or status.active.")
			} else if !displayFieldExists(entity, fieldName) {
				v.addDiagnostic(label.Position, "UNKNOWN_LABEL_TARGET", fmt.Sprintf("Label target %s references unknown field %s.%s.", label.Target, entityName, fieldName), "Use an existing entity field or computed field.")
			}
		}

		if len(label.Translations) == 0 {
			v.addDiagnostic(label.Position, "MISSING_LABEL_TRANSLATION", fmt.Sprintf("Label target %s has no translations.", label.Target), "Add locale text lines such as `tr \"Ürün Adı\"`.")
			continue
		}

		translationLocales := map[string]Position{}
		hasDefault := defaultLocale == ""
		for _, translation := range label.Translations {
			if translation.Locale == "" || translation.Text == "" {
				v.addDiagnostic(translation.Position, "MISSING_LABEL_TRANSLATION", fmt.Sprintf("Label target %s has an empty locale or text.", label.Target), "Write `locale \"Text\"`.")
				continue
			}
			if !isThemeIdentifier(translation.Locale) {
				v.addDiagnostic(translation.Position, "INVALID_LOCALE", fmt.Sprintf("Label target %s uses invalid locale %q.", label.Target, translation.Locale), "Use locale names such as tr, en, or en-US.")
				continue
			}
			if len(locales) > 0 && !locales[translation.Locale] {
				v.addDiagnostic(translation.Position, "UNKNOWN_LABEL_LOCALE", fmt.Sprintf("Label target %s uses locale %s that is not listed in i18n.locales.", label.Target, translation.Locale), "Add the locale to `locales`, or remove this translation line.")
				continue
			}
			if existing, ok := translationLocales[translation.Locale]; ok {
				v.addDiagnostic(translation.Position, "DUPLICATE_LABEL_LOCALE", fmt.Sprintf("Label target %s repeats locale %s.", label.Target, translation.Locale), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
				continue
			}
			translationLocales[translation.Locale] = translation.Position
			if translation.Locale == defaultLocale {
				hasDefault = true
			}
		}
		if !hasDefault {
			v.addDiagnostic(label.Position, "MISSING_DEFAULT_LABEL_TRANSLATION", fmt.Sprintf("Label target %s has no %s translation.", label.Target, defaultLocale), "Add a translation for the default locale.")
		}
	}
}

func (v *semanticValidator) validateFieldTextTranslations(entityIndex map[string]EntityDecl) {
	v.validateFieldTextTranslationBlocks(entityIndex, "placeholder", "Placeholder", v.program.Placeholders)
	v.validateFieldTextTranslationBlocks(entityIndex, "help", "Help text", v.program.HelpTexts)
	v.validateFieldTextTranslationBlocks(entityIndex, "message", "Message", v.program.Messages)
}

func (v *semanticValidator) validateFieldTextTranslationBlocks(entityIndex map[string]EntityDecl, kind string, display string, blocks []FieldTextTranslationDecl) {
	if len(blocks) == 0 {
		return
	}

	codeKind := strings.ToUpper(kind)
	if v.program.I18N == nil {
		v.addDiagnostic(blocks[0].Position, "MISSING_I18N", fmt.Sprintf("%s translations require an i18n block.", display), fmt.Sprintf("Add `i18n { default tr locales tr, en }` before %s translation blocks.", kind))
	}

	locales := map[string]bool{}
	defaultLocale := ""
	if v.program.I18N != nil {
		defaultLocale = v.program.I18N.Default
		for _, locale := range v.program.I18N.Locales {
			locales[locale] = true
		}
	}

	targets := map[string]FieldTextTranslationDecl{}
	for _, block := range blocks {
		if existing, ok := targets[block.Target]; ok {
			v.addDiagnostic(block.Position, "DUPLICATE_"+codeKind+"_TARGET", fmt.Sprintf("%s target %s is already defined.", display, block.Target), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		targets[block.Target] = block

		entityName, fieldName, ok := splitLabelTarget(block.Target)
		if !ok {
			v.addDiagnostic(block.Position, "INVALID_"+codeKind+"_TARGET", fmt.Sprintf("%s target %s is invalid.", display, block.Target), fmt.Sprintf("Use `%s Entity.field { ... }`.", kind))
		} else if entity, exists := entityIndex[entityName]; !exists {
			v.addDiagnostic(block.Position, "UNKNOWN_"+codeKind+"_TARGET", fmt.Sprintf("%s target %s references unknown entity %s.", display, block.Target, entityName), "Use an existing entity and stored field.")
		} else if _, exists := fieldIndex(entity)[fieldName]; !exists {
			if _, computed := computedFieldIndex(entity)[fieldName]; computed {
				v.addDiagnostic(block.Position, "UNSUPPORTED_"+codeKind+"_TARGET", fmt.Sprintf("%s target %s references computed field %s.%s.", display, block.Target, entityName, fieldName), "Use label translations for computed display fields; placeholder, help, and message translations target stored form fields.")
			} else {
				v.addDiagnostic(block.Position, "UNKNOWN_"+codeKind+"_TARGET", fmt.Sprintf("%s target %s references unknown field %s.%s.", display, block.Target, entityName, fieldName), "Use an existing stored entity field.")
			}
		}

		if len(block.Translations) == 0 {
			v.addDiagnostic(block.Position, "MISSING_"+codeKind+"_TRANSLATION", fmt.Sprintf("%s target %s has no translations.", display, block.Target), "Add locale text lines such as `tr \"Ürün adını gir\"`.")
			continue
		}

		translationLocales := map[string]Position{}
		hasDefault := defaultLocale == ""
		for _, translation := range block.Translations {
			if translation.Locale == "" || translation.Text == "" {
				v.addDiagnostic(translation.Position, "MISSING_"+codeKind+"_TRANSLATION", fmt.Sprintf("%s target %s has an empty locale or text.", display, block.Target), "Write `locale \"Text\"`.")
				continue
			}
			if !isThemeIdentifier(translation.Locale) {
				v.addDiagnostic(translation.Position, "INVALID_LOCALE", fmt.Sprintf("%s target %s uses invalid locale %q.", display, block.Target, translation.Locale), "Use locale names such as tr, en, or en-US.")
				continue
			}
			if len(locales) > 0 && !locales[translation.Locale] {
				v.addDiagnostic(translation.Position, "UNKNOWN_"+codeKind+"_LOCALE", fmt.Sprintf("%s target %s uses locale %s that is not listed in i18n.locales.", display, block.Target, translation.Locale), "Add the locale to `locales`, or remove this translation line.")
				continue
			}
			if existing, ok := translationLocales[translation.Locale]; ok {
				v.addDiagnostic(translation.Position, "DUPLICATE_"+codeKind+"_LOCALE", fmt.Sprintf("%s target %s repeats locale %s.", display, block.Target, translation.Locale), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
				continue
			}
			translationLocales[translation.Locale] = translation.Position
			if translation.Locale == defaultLocale {
				hasDefault = true
			}
		}
		if !hasDefault {
			v.addDiagnostic(block.Position, "MISSING_DEFAULT_"+codeKind+"_TRANSLATION", fmt.Sprintf("%s target %s has no %s translation.", display, block.Target, defaultLocale), "Add a translation for the default locale.")
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
		} else if field.Type == "file" || field.Type == "image" {
			v.addDiagnostic(field.Position, "UNSUPPORTED_AUTH_USER_FIELD_TYPE", fmt.Sprintf("Auth user field %s uses media type %q.", field.Name, field.Type), "Use media fields on entities; generated auth user media fields are not supported yet.")
		}
		for _, modifier := range field.Modifiers {
			if modifier.Name == "load" {
				v.addDiagnostic(field.Position, "UNSUPPORTED_AUTH_USER_FIELD_MODIFIER", fmt.Sprintf("Auth user field %s uses unsupported modifier %q.", field.Name, modifier.Name), "Use load only on entity relation fields.")
				continue
			}
			if !supportedFieldModifiers[modifier.Name] {
				v.addDiagnostic(field.Position, "UNSUPPORTED_AUTH_USER_FIELD_MODIFIER", fmt.Sprintf("Auth user field %s uses unsupported modifier %q.", field.Name, modifier.Name), "Use required, unique, optional, default, label, placeholder, help, min, max, length, regex, url, accept, or message for supported auth user field types.")
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
		v.validateInlineUI("auth user field", "user."+field.Name, field.UI, inlineUIModesForTarget("field"))
	}
}

func (v *semanticValidator) validateTarget() {
	if v.program.Target == nil {
		return
	}

	target := *v.program.Target
	if target.Name == "" {
		v.addDiagnostic(target.Position, "MISSING_TARGET_NAME", "Target block is missing a name.", "Use `target web { ... }` or `target api { ... }`.")
	} else if !supportedTargetNames[target.Name] {
		v.addDiagnostic(target.Position, "UNSUPPORTED_TARGET", fmt.Sprintf("Target %q is not supported.", target.Name), "Use web or api in v0.2.")
	}

	if target.Name == "api" {
		if target.Frontend != "" {
			v.addDiagnostic(target.Position, "UNSUPPORTED_TARGET_FRONTEND", fmt.Sprintf("Target api does not support frontend %q.", target.Frontend), "Omit the frontend line inside target api.")
		}
		if len(v.program.Tests) > 0 {
			v.addDiagnostic(target.Position, "UNSUPPORTED_API_TARGET_TEST", "Target api does not generate browser test runtime.", "Use target web for browser test declarations or remove top-level test blocks.")
		}
	} else if target.Frontend == "" {
		v.addDiagnostic(target.Position, "MISSING_TARGET_FRONTEND", "Target block is missing a frontend.", "Add `frontend react` inside target web.")
	} else if !supportedTargetFrontends[target.Frontend] {
		v.addDiagnostic(target.Position, "UNSUPPORTED_TARGET_FRONTEND", fmt.Sprintf("Target frontend %q is not supported.", target.Frontend), "Use react in target web.")
	}

	if target.Backend == "" {
		v.addDiagnostic(target.Position, "MISSING_TARGET_BACKEND", "Target block is missing a backend.", "Add `backend node` inside target.")
	} else if !supportedTargetBackends[target.Backend] {
		v.addDiagnostic(target.Position, "UNSUPPORTED_TARGET_BACKEND", fmt.Sprintf("Target backend %q is not supported.", target.Backend), "Use node in v0.2.")
	}

	if target.Database == "" {
		v.addDiagnostic(target.Position, "MISSING_TARGET_DATABASE", "Target block is missing a database.", "Add `database sqlite`, `database postgres`, or `database mysql` inside target.")
	} else if !supportedTargetDatabases[target.Database] {
		v.addDiagnostic(target.Position, "UNSUPPORTED_TARGET_DATABASE", fmt.Sprintf("Target database %q is not supported.", target.Database), "Use sqlite, postgres, or mysql in v0.2.")
	} else if target.Database == "mysql" && len(v.program.Migrations) > 0 {
		v.addDiagnostic(target.Position, "UNSUPPORTED_TARGET_DATABASE_MIGRATION", "Target database mysql does not support generated rename migration runtime yet.", "Use sqlite or postgres for migration blocks, or remove migration blocks before targeting mysql.")
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

func (v *semanticValidator) validateSecurity() {
	if v.program.Security == nil {
		return
	}
	if v.program.Security.CORS == nil {
		return
	}

	cors := *v.program.Security.CORS
	if cors.Origins.Name == "" {
		v.addDiagnostic(cors.Position, "MISSING_CORS_ORIGINS", "CORS block is missing an origins environment reference.", "Add `origins env CORS_ORIGINS` inside cors.")
	} else if !validEnvName(cors.Origins.Name) {
		v.addDiagnostic(cors.Origins.Position, "INVALID_ENV_NAME", fmt.Sprintf("CORS origins reference invalid environment variable %q.", cors.Origins.Name), "Use uppercase letters, numbers, and underscores, such as CORS_ORIGINS.")
	}

	if cors.Credentials != "" && cors.Credentials != "true" && cors.Credentials != "false" {
		v.addDiagnostic(cors.Position, "INVALID_CORS_CREDENTIALS", fmt.Sprintf("CORS credentials uses unsupported value %q.", cors.Credentials), "Use `credentials true` or `credentials false`.")
	}
}

func (v *semanticValidator) validateDeploy() {
	if v.program.Deploy == nil {
		return
	}

	deploy := *v.program.Deploy
	if deploy.Target == "" {
		v.addDiagnostic(deploy.Position, "MISSING_DEPLOY_TARGET", "Deploy block is missing a target.", "Add `target docker` inside deploy.")
	} else if !supportedDeployTargets[deploy.Target] {
		v.addDiagnostic(deploy.Position, "UNSUPPORTED_DEPLOY_TARGET", fmt.Sprintf("Deploy target %q is not supported.", deploy.Target), "Use docker in v0.1.")
	}

	if deploy.Port != nil {
		if deploy.Port.Env.Name == "" {
			v.addDiagnostic(deploy.Port.Position, "MISSING_DEPLOY_PORT_ENV", "Deploy port is missing an environment variable name.", "Use `port env PORT default 3001`.")
		} else if !validEnvName(deploy.Port.Env.Name) {
			v.addDiagnostic(deploy.Port.Env.Position, "INVALID_ENV_NAME", fmt.Sprintf("Deploy port references invalid environment variable %q.", deploy.Port.Env.Name), "Use uppercase letters, numbers, and underscores, such as PORT.")
		}
		port, err := strconv.Atoi(deploy.Port.Default)
		if err != nil || port < 1 || port > 65535 {
			v.addDiagnostic(deploy.Port.Position, "INVALID_DEPLOY_PORT_DEFAULT", fmt.Sprintf("Deploy port default %q is invalid.", deploy.Port.Default), "Use a TCP port between 1 and 65535.")
		}
	}

	envIndex := map[string]DeployEnvDecl{}
	for _, env := range deploy.Env {
		if !validEnvName(env.Name) {
			v.addDiagnostic(env.Position, "INVALID_ENV_NAME", fmt.Sprintf("Deploy env references invalid environment variable %q.", env.Name), "Use uppercase letters, numbers, and underscores, such as DATABASE_URL.")
		}
		if !supportedDeployEnvModes[env.Mode] {
			v.addDiagnostic(env.Position, "UNSUPPORTED_DEPLOY_ENV_MODE", fmt.Sprintf("Deploy env %s uses unsupported mode %q.", env.Name, env.Mode), "Use required or optional.")
		}
		if existing, ok := envIndex[env.Name]; ok {
			v.addDiagnostic(env.Position, "DUPLICATE_DEPLOY_ENV", fmt.Sprintf("Deploy env %s is already declared.", env.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		envIndex[env.Name] = env
	}

	if deploy.Preview != nil && !supportedDeployPreviewModes[deploy.Preview.Mode] {
		v.addDiagnostic(deploy.Preview.Position, "UNSUPPORTED_DEPLOY_PREVIEW", fmt.Sprintf("Deploy preview mode %q is not supported.", deploy.Preview.Mode), "Use `preview local` in v0.2.")
	}

	if deploy.Rollback != nil {
		if !supportedDeployRollbackStrategies[deploy.Rollback.Strategy] {
			v.addDiagnostic(deploy.Rollback.Position, "UNSUPPORTED_DEPLOY_ROLLBACK", fmt.Sprintf("Deploy rollback strategy %q is not supported.", deploy.Rollback.Strategy), "Use `rollback keep NUMBER` in v0.2.")
		}
		if deploy.Rollback.Keep < 1 || deploy.Rollback.Keep > 20 {
			v.addDiagnostic(deploy.Rollback.Position, "INVALID_DEPLOY_ROLLBACK_KEEP", fmt.Sprintf("Deploy rollback keep count %d is invalid.", deploy.Rollback.Keep), "Use a rollback keep count between 1 and 20.")
		}
	}

	if deploy.Cloud != nil {
		if deploy.Target != "docker" {
			v.addDiagnostic(deploy.Cloud.Position, "DEPLOY_CLOUD_REQUIRES_DOCKER", "Deploy cloud adapters package the generated Docker web target.", "Declare `target docker` before `cloud fly app env FLY_APP_NAME`.")
		}
		if !supportedDeployCloudProviders[deploy.Cloud.Provider] {
			v.addDiagnostic(deploy.Cloud.Position, "UNSUPPORTED_DEPLOY_CLOUD_PROVIDER", fmt.Sprintf("Deploy cloud provider %q is not supported.", deploy.Cloud.Provider), "Use fly, render, or railway.")
		}
		if deploy.Cloud.App.Name == "" {
			v.addDiagnostic(deploy.Cloud.Position, "MISSING_DEPLOY_CLOUD_APP_ENV", "Deploy cloud app is missing an environment variable name.", "Use `cloud fly app env FLY_APP_NAME`.")
		} else if !validEnvName(deploy.Cloud.App.Name) {
			v.addDiagnostic(deploy.Cloud.App.Position, "INVALID_ENV_NAME", fmt.Sprintf("Deploy cloud app references invalid environment variable %q.", deploy.Cloud.App.Name), "Use uppercase letters, numbers, and underscores, such as FLY_APP_NAME.")
		}
		if deploy.Cloud.Region.Name != "" && !validEnvName(deploy.Cloud.Region.Name) {
			v.addDiagnostic(deploy.Cloud.Region.Position, "INVALID_ENV_NAME", fmt.Sprintf("Deploy cloud region references invalid environment variable %q.", deploy.Cloud.Region.Name), "Use uppercase letters, numbers, and underscores, such as FLY_REGION.")
		}
	}
}

func (v *semanticValidator) validateOps() {
	if v.program.Ops == nil {
		return
	}

	ops := *v.program.Ops
	if ops.Health == nil && ops.Readiness == nil && ops.Metrics == nil && ops.Logging == "" && ops.Observe == nil {
		v.addDiagnostic(ops.Position, "MISSING_OPS_SIGNAL", "Ops block does not declare any runtime signal.", "Add `health path \"/healthz\"`, `readiness path \"/readyz\"`, `metrics path \"/metrics\"`, `logging requests`, or `observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT`.")
		return
	}

	paths := map[string]string{}
	v.validateOpsEndpoint("health", ops.Health, paths)
	v.validateOpsEndpoint("readiness", ops.Readiness, paths)
	v.validateOpsEndpoint("metrics", ops.Metrics, paths)

	if ops.Logging != "" && !supportedOpsLoggingModes[ops.Logging] {
		v.addDiagnostic(ops.Position, "UNSUPPORTED_OPS_LOGGING", fmt.Sprintf("Ops logging mode %q is not supported.", ops.Logging), "Use `logging requests`.")
	}
	v.validateOpsObserve(ops.Observe)
}

func (v *semanticValidator) validateOpsEndpoint(name string, endpoint *OpsEndpointDecl, paths map[string]string) {
	if endpoint == nil {
		return
	}
	if !validOpsPath(endpoint.Path) {
		v.addDiagnostic(endpoint.Position, "INVALID_OPS_PATH", fmt.Sprintf("Ops %s path %q is invalid.", name, endpoint.Path), "Use an absolute public path such as `/healthz`, `/readyz`, or `/metrics`; do not use `/api`, `/openapi.json`, query strings, fragments, braces, or whitespace.")
		return
	}
	if existing, ok := paths[endpoint.Path]; ok {
		v.addDiagnostic(endpoint.Position, "DUPLICATE_OPS_PATH", fmt.Sprintf("Ops %s path %q is already used by %s.", name, endpoint.Path, existing), "Use distinct paths for health, readiness, and metrics.")
		return
	}
	paths[endpoint.Path] = name
}

func (v *semanticValidator) validateOpsObserve(observe *OpsObserveDecl) {
	if observe == nil {
		return
	}
	if !supportedOpsObserveProviders[observe.Provider] {
		v.addDiagnostic(observe.Position, "UNSUPPORTED_OPS_OBSERVE_PROVIDER", fmt.Sprintf("Ops observe provider %q is not supported.", observe.Provider), "Use `observe webhook endpoint env BLACKLANG_OBSERVABILITY_ENDPOINT` or `observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT`.")
	}
	if observe.Endpoint.Name == "" {
		v.addDiagnostic(observe.Position, "MISSING_OPS_OBSERVE_ENDPOINT_ENV", "Ops observe endpoint is missing an environment variable name.", "Use `observe webhook endpoint env BLACKLANG_OBSERVABILITY_ENDPOINT` or `observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT`.")
	} else if !validEnvName(observe.Endpoint.Name) {
		v.addDiagnostic(observe.Endpoint.Position, "INVALID_ENV_NAME", fmt.Sprintf("Ops observe endpoint references invalid environment variable %q.", observe.Endpoint.Name), "Use uppercase letters, numbers, and underscores, such as BLACKLANG_OTLP_ENDPOINT.")
	}
}

func validOpsPath(value string) bool {
	if value == "" || value == "/" || value == "/openapi.json" || value == "/api" || strings.HasPrefix(value, "/api/") || !strings.HasPrefix(value, "/") {
		return false
	}
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
			continue
		case char >= 'A' && char <= 'Z':
			continue
		case char >= '0' && char <= '9':
			continue
		case char == '/' || char == '-' || char == '_' || char == '.':
			continue
		default:
			return false
		}
	}
	return true
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

func (v *semanticValidator) validateMigrations(entityIndex map[string]EntityDecl) {
	migrationIndex := map[string]MigrationDecl{}
	renameSourceIndex := map[string]MigrationRenameDecl{}
	renameTargetIndex := map[string]MigrationRenameDecl{}
	for _, migration := range v.program.Migrations {
		if !isThemeIdentifier(migration.Name) {
			v.addDiagnostic(migration.Position, "INVALID_MIGRATION_NAME", fmt.Sprintf("Migration %q is not a valid identifier.", migration.Name), "Use a stable identifier such as `RenameProductName`.")
		}
		if existing, ok := migrationIndex[migration.Name]; ok {
			v.addDiagnostic(migration.Position, "DUPLICATE_MIGRATION", fmt.Sprintf("Migration %s is already defined.", migration.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		migrationIndex[migration.Name] = migration
		if len(migration.Renames) == 0 {
			v.addDiagnostic(migration.Position, "MISSING_MIGRATION_RENAME", fmt.Sprintf("Migration %s has no rename declarations.", migration.Name), "Add `rename entity Old to New` or `rename field Entity.oldField to newField`.")
		}

		for _, rename := range migration.Renames {
			sourceKey := migrationRenameSourceKey(rename)
			if existing, ok := renameSourceIndex[sourceKey]; ok {
				v.addDiagnostic(rename.Position, "DUPLICATE_MIGRATION_RENAME", fmt.Sprintf("Migration rename source %s is already declared.", migrationRenameSourceLabel(rename)), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
				continue
			}
			renameSourceIndex[sourceKey] = rename
			targetKey := migrationRenameTargetKey(rename)
			if existing, ok := renameTargetIndex[targetKey]; ok {
				v.addDiagnostic(rename.Position, "DUPLICATE_MIGRATION_RENAME", fmt.Sprintf("Migration rename target %s is already declared.", migrationRenameTargetLabel(rename)), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
				continue
			}
			renameTargetIndex[targetKey] = rename

			switch rename.Kind {
			case "entity":
				v.validateEntityMigrationRename(entityIndex, migration, rename)
			case "field":
				v.validateFieldMigrationRename(entityIndex, migration, rename)
			default:
				v.addDiagnostic(rename.Position, "INVALID_MIGRATION_RENAME_KIND", fmt.Sprintf("Migration %s uses unsupported rename kind %q.", migration.Name, rename.Kind), "Use entity or field.")
			}
		}
	}
}

func (v *semanticValidator) validateEntityMigrationRename(entityIndex map[string]EntityDecl, migration MigrationDecl, rename MigrationRenameDecl) {
	if rename.From == "" || rename.To == "" {
		v.addDiagnostic(rename.Position, "INVALID_MIGRATION_RENAME", fmt.Sprintf("Migration %s has an incomplete entity rename.", migration.Name), "Use `rename entity OldEntity to NewEntity`.")
		return
	}
	if rename.From == rename.To {
		v.addDiagnostic(rename.Position, "INVALID_MIGRATION_RENAME", fmt.Sprintf("Migration %s renames entity %s to itself.", migration.Name, rename.From), "Rename from an old entity name to the current entity name.")
		return
	}
	if _, ok := entityIndex[rename.To]; !ok {
		v.addDiagnostic(rename.Position, "UNKNOWN_MIGRATION_RENAME_TARGET", fmt.Sprintf("Migration %s renames entity %s to unknown current entity %s.", migration.Name, rename.From, rename.To), "The `to` entity must exist in the current source.")
	}
	if _, ok := entityIndex[rename.From]; ok {
		v.addDiagnostic(rename.Position, "MIGRATION_RENAME_SOURCE_STILL_EXISTS", fmt.Sprintf("Migration %s rename source entity %s still exists in the current source.", migration.Name, rename.From), "Remove the old entity declaration and keep only the current `to` entity.")
	}
}

func (v *semanticValidator) validateFieldMigrationRename(entityIndex map[string]EntityDecl, migration MigrationDecl, rename MigrationRenameDecl) {
	if rename.Entity == "" || rename.From == "" || rename.To == "" {
		v.addDiagnostic(rename.Position, "INVALID_MIGRATION_RENAME", fmt.Sprintf("Migration %s has an incomplete field rename.", migration.Name), "Use `rename field Entity.oldField to newField`.")
		return
	}
	if rename.From == rename.To {
		v.addDiagnostic(rename.Position, "INVALID_MIGRATION_RENAME", fmt.Sprintf("Migration %s renames field %s.%s to itself.", migration.Name, rename.Entity, rename.From), "Rename from the old field name to the current field name.")
		return
	}
	entity, ok := entityIndex[rename.Entity]
	if !ok {
		v.addDiagnostic(rename.Position, "UNKNOWN_MIGRATION_RENAME_ENTITY", fmt.Sprintf("Migration %s references unknown entity %s.", migration.Name, rename.Entity), "Use a current entity name in `rename field Entity.oldField to newField`.")
		return
	}
	fields := fieldIndex(entity)
	if _, ok := fields[rename.To]; !ok {
		if _, computed := computedFieldIndex(entity)[rename.To]; computed {
			v.addDiagnostic(rename.Position, "UNSUPPORTED_MIGRATION_RENAME_TARGET", fmt.Sprintf("Migration %s renames stored field %s.%s to computed field %s.%s.", migration.Name, rename.Entity, rename.From, rename.Entity, rename.To), "Rename migrations can target stored fields only; computed fields are display-only.")
			return
		}
		v.addDiagnostic(rename.Position, "UNKNOWN_MIGRATION_RENAME_TARGET", fmt.Sprintf("Migration %s renames field %s.%s to unknown current field %s.%s.", migration.Name, rename.Entity, rename.From, rename.Entity, rename.To), "The `to` field must exist as a stored field on the current entity.")
	}
	if _, ok := fields[rename.From]; ok {
		v.addDiagnostic(rename.Position, "MIGRATION_RENAME_SOURCE_STILL_EXISTS", fmt.Sprintf("Migration %s rename source field %s.%s still exists in the current source.", migration.Name, rename.Entity, rename.From), "Remove the old field declaration and keep only the current `to` field.")
	}
}

func migrationRenameSourceKey(rename MigrationRenameDecl) string {
	return rename.Kind + "\x00" + rename.Entity + "\x00" + rename.From
}

func migrationRenameTargetKey(rename MigrationRenameDecl) string {
	return rename.Kind + "\x00" + rename.Entity + "\x00" + rename.To
}

func migrationRenameSourceLabel(rename MigrationRenameDecl) string {
	if rename.Kind == "field" {
		return rename.Entity + "." + rename.From
	}
	return rename.From
}

func migrationRenameTargetLabel(rename MigrationRenameDecl) string {
	if rename.Kind == "field" {
		return rename.Entity + "." + rename.To
	}
	return rename.To
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

		_, relationField := entityIndex[field.Type]
		if !supportedFieldTypes[field.Type] {
			if !relationField {
				v.addDiagnostic(field.Position, "UNSUPPORTED_FIELD_TYPE", fmt.Sprintf("Field %s.%s uses unsupported type %q.", entity.Name, field.Name, field.Type), "Use a primitive type or the name of an existing entity.")
			}
		}

		for _, modifier := range field.Modifiers {
			if !supportedFieldModifiers[modifier.Name] {
				v.addDiagnostic(field.Position, "UNSUPPORTED_FIELD_MODIFIER", fmt.Sprintf("Field %s.%s uses unsupported modifier %q.", entity.Name, field.Name, modifier.Name), "Use required, unique, optional, default, label, placeholder, help, min, max, length, regex, url, accept, message, or load.")
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
			if modifier.Name == "load" {
				v.validateRelationLoadModifier(entity, field, relationField, modifier)
			}
			v.validateConstraintModifier("field", entity.Name, field, modifier)
		}
		if relationLoadModifierCount(field) > 1 {
			v.addDiagnostic(field.Position, "DUPLICATE_RELATION_LOAD", fmt.Sprintf("Field %s.%s repeats relation load policy.", entity.Name, field.Name), "Keep one load modifier, such as `load list detail query`.")
		}
		v.validateInlineUI("field", entity.Name+"."+field.Name, field.UI, inlineUIModesForTarget("field"))
	}
	v.validateComputedFields(entity, fieldIndex)
	v.validateEntityPolicies(entity, fieldIndex)
	v.validateEntityIndexes(entity, fieldIndex)
	v.validateEntityValidations(entity, fieldIndex)
}

func (v *semanticValidator) validateRelationLoadModifier(entity EntityDecl, field FieldDecl, relationField bool, modifier Modifier) {
	if !relationField {
		v.addDiagnostic(field.Position, "UNSUPPORTED_RELATION_LOAD_FIELD", fmt.Sprintf("Field %s.%s uses load but is not a relation field.", entity.Name, field.Name), "Use load only on fields whose type is another entity, such as `supplier Supplier load list detail`.")
		return
	}

	scopes := relationLoadScopesFromValue(modifier.Value)
	if len(scopes) == 0 {
		v.addDiagnostic(field.Position, "MISSING_RELATION_LOAD_SCOPE", fmt.Sprintf("Field %s.%s has load without a scope.", entity.Name, field.Name), "Use one or more of list, detail, query, mutation, or use `load none`.")
		return
	}

	seen := map[string]bool{}
	hasNone := false
	for _, scope := range scopes {
		if !supportedRelationLoadScopes[scope] {
			v.addDiagnostic(field.Position, "UNSUPPORTED_RELATION_LOAD_SCOPE", fmt.Sprintf("Field %s.%s uses unsupported relation load scope %q.", entity.Name, field.Name, scope), "Use list, detail, query, mutation, or none.")
			continue
		}
		if seen[scope] {
			v.addDiagnostic(field.Position, "DUPLICATE_RELATION_LOAD_SCOPE", fmt.Sprintf("Field %s.%s repeats relation load scope %q.", entity.Name, field.Name, scope), "Keep each load scope once.")
		}
		seen[scope] = true
		if scope == "none" {
			hasNone = true
		}
	}

	if hasNone && len(scopes) > 1 {
		v.addDiagnostic(field.Position, "CONFLICTING_RELATION_LOAD_SCOPE", fmt.Sprintf("Field %s.%s combines load none with other scopes.", entity.Name, field.Name), "Use `load none` by itself, or list explicit scopes without none.")
	}
}

func (v *semanticValidator) validateEntityPolicies(entity EntityDecl, fieldIndex map[string]FieldDecl) {
	policies := map[string]EntityPolicyDecl{}
	policyFields := map[string]EntityPolicyDecl{}
	for _, policy := range entity.Policies {
		if policy.Kind != "owner" && policy.Kind != "tenant" {
			v.addDiagnostic(policy.Position, "UNSUPPORTED_ENTITY_POLICY", fmt.Sprintf("Entity %s uses unsupported policy %q.", entity.Name, policy.Kind), "Use owner or tenant.")
			continue
		}
		if v.program.Auth == nil {
			v.addDiagnostic(policy.Position, "AUTH_REQUIRED_FOR_ENTITY_POLICY", fmt.Sprintf("Entity %s uses %s policy without an auth block.", entity.Name, policy.Kind), "Add auth { strategy emailPassword session cookie ... } or remove the entity policy.")
		}
		if existing, ok := policies[policy.Kind]; ok {
			v.addDiagnostic(policy.Position, "DUPLICATE_ENTITY_POLICY", fmt.Sprintf("Entity %s repeats %s policy.", entity.Name, policy.Kind), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		policies[policy.Kind] = policy
		if existing, ok := policyFields[policy.Field]; ok {
			v.addDiagnostic(policy.Position, "DUPLICATE_ENTITY_POLICY_FIELD", fmt.Sprintf("Entity %s uses field %s for more than one row policy.", entity.Name, policy.Field), fmt.Sprintf("First policy field use is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		policyFields[policy.Field] = policy

		field, ok := fieldIndex[policy.Field]
		if !ok {
			if _, computed := computedFieldIndex(entity)[policy.Field]; computed {
				v.addDiagnostic(policy.Position, "UNSUPPORTED_ENTITY_POLICY_FIELD", fmt.Sprintf("Entity %s %s policy references computed field %s.%s.", entity.Name, policy.Kind, entity.Name, policy.Field), "Use a stored text field such as ownerId or tenantId.")
				continue
			}
			v.addDiagnostic(policy.Position, "UNKNOWN_ENTITY_POLICY_FIELD", fmt.Sprintf("Entity %s %s policy references unknown field %s.%s.", entity.Name, policy.Kind, entity.Name, policy.Field), "Declare a stored text field such as ownerId text required, then add `policy owner ownerId`.")
			continue
		}
		if field.Type != "text" {
			v.addDiagnostic(policy.Position, "UNSUPPORTED_ENTITY_POLICY_FIELD", fmt.Sprintf("Entity %s %s policy field %s.%s uses %s.", entity.Name, policy.Kind, entity.Name, field.Name, field.Type), "Use a stored text field because generated auth user and tenant identifiers are strings.")
		}
		if !hasModifier(field, "required") {
			v.addDiagnostic(policy.Position, "MISSING_ENTITY_POLICY_REQUIRED_FIELD", fmt.Sprintf("Entity %s %s policy field %s.%s must be required.", entity.Name, policy.Kind, entity.Name, field.Name), "Use `text required` so every row receives a deterministic policy value.")
		}
		if hasModifier(field, "unique") {
			v.addDiagnostic(policy.Position, "UNSUPPORTED_ENTITY_POLICY_UNIQUE_FIELD", fmt.Sprintf("Entity %s %s policy field %s.%s cannot be unique.", entity.Name, policy.Kind, entity.Name, field.Name), "Remove unique; many rows can share the same owner or tenant.")
		}
	}
}

func (v *semanticValidator) validateComputedFields(entity EntityDecl, fieldIndex map[string]FieldDecl) {
	computedIndex := map[string]ComputedFieldDecl{}
	for _, computed := range entity.ComputedFields {
		if existing, ok := fieldIndex[computed.Name]; ok {
			v.addDiagnostic(computed.Position, "DUPLICATE_FIELD", fmt.Sprintf("Computed field %s.%s conflicts with stored field %s.%s.", entity.Name, computed.Name, entity.Name, existing.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		if existing, ok := computedIndex[computed.Name]; ok {
			v.addDiagnostic(computed.Position, "DUPLICATE_COMPUTED_FIELD", fmt.Sprintf("Computed field %s.%s is already defined.", entity.Name, computed.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		computedIndex[computed.Name] = computed

		if !supportedComputedFieldTypes[computed.Type] {
			v.addDiagnostic(computed.Position, "UNSUPPORTED_COMPUTED_FIELD_TYPE", fmt.Sprintf("Computed field %s.%s uses unsupported type %q.", entity.Name, computed.Name, computed.Type), "Use number, integer, decimal, or money for computed fields in v0.2.")
		}
		if computed.Expression.Tree != nil {
			v.validateComputedExpression(entity, computed, *computed.Expression.Tree, fieldIndex)
		} else {
			if !supportedComputedOperators[computed.Expression.Operator] {
				v.addDiagnostic(computed.Expression.Position, "UNSUPPORTED_COMPUTED_OPERATOR", fmt.Sprintf("Computed field %s.%s uses unsupported operator %q.", entity.Name, computed.Name, computed.Expression.Operator), "Use +, -, *, or /.")
			}
			v.validateComputedOperand(entity, computed, computed.Expression.Left, fieldIndex)
			v.validateComputedOperand(entity, computed, computed.Expression.Right, fieldIndex)
		}

		for _, modifier := range computed.Modifiers {
			if !supportedComputedFieldModifiers[modifier.Name] {
				v.addDiagnostic(computed.Position, "UNSUPPORTED_COMPUTED_FIELD_MODIFIER", fmt.Sprintf("Computed field %s.%s uses unsupported modifier %q.", entity.Name, computed.Name, modifier.Name), "Use label or help on computed fields in v0.2.")
			}
			if modifier.Name == "label" && modifier.Value == "" {
				v.addDiagnostic(computed.Position, "MISSING_LABEL_VALUE", fmt.Sprintf("Computed field %s.%s has label without a value.", entity.Name, computed.Name), "Write label followed by text, such as `label \"Inventory Value\"`.")
			}
			if modifier.Name == "help" && modifier.Value == "" {
				v.addDiagnostic(computed.Position, "MISSING_HELP_VALUE", fmt.Sprintf("Computed field %s.%s has help without a value.", entity.Name, computed.Name), "Write help followed by text, such as `help \"Calculated from stock and price\"`.")
			}
		}
	}
}

func (v *semanticValidator) validateComputedExpression(entity EntityDecl, computed ComputedFieldDecl, expression ExpressionDecl, fieldIndex map[string]FieldDecl) bool {
	if expression.Kind != "binary" {
		if expression.ValueKind != "number" && expression.ValueKind != "reference" {
			v.addDiagnostic(expression.Position, "INVALID_COMPUTED_EXPRESSION", fmt.Sprintf("Computed field %s.%s has invalid expression value %q.", entity.Name, computed.Name, expression.Value), "Use stored number-like fields or numeric literals.")
			return false
		}
		return v.validateComputedOperand(entity, computed, expression.Value, fieldIndex)
	}

	operatorOK := supportedComputedOperators[expression.Operator]
	if !operatorOK {
		v.addDiagnostic(expression.Position, "UNSUPPORTED_COMPUTED_OPERATOR", fmt.Sprintf("Computed field %s.%s uses unsupported operator %q.", entity.Name, computed.Name, expression.Operator), "Use +, -, *, or /.")
	}
	leftOK := v.validateComputedExpression(entity, computed, dereferenceExpression(expression.Left), fieldIndex)
	rightOK := v.validateComputedExpression(entity, computed, dereferenceExpression(expression.Right), fieldIndex)
	if expression.Operator == "/" && expression.Right != nil && expressionIsZeroNumber(*expression.Right) {
		v.addDiagnostic(expression.Position, "INVALID_COMPUTED_EXPRESSION", fmt.Sprintf("Computed field %s.%s divides by zero.", entity.Name, computed.Name), "Use a non-zero literal divisor.")
		return false
	}
	return operatorOK && leftOK && rightOK
}

func (v *semanticValidator) validateComputedOperand(entity EntityDecl, computed ComputedFieldDecl, operand string, fieldIndex map[string]FieldDecl) bool {
	if operand == "" {
		v.addDiagnostic(computed.Expression.Position, "INVALID_COMPUTED_EXPRESSION", fmt.Sprintf("Computed field %s.%s has an empty expression operand.", entity.Name, computed.Name), "Use `computed name type = numericExpression`.")
		return false
	}
	if isNumericLiteral(operand) {
		return true
	}
	field, ok := fieldIndex[operand]
	if !ok {
		v.addDiagnostic(computed.Expression.Position, "UNKNOWN_COMPUTED_FIELD", fmt.Sprintf("Computed field %s.%s references unknown field %s.", entity.Name, computed.Name, operand), "Use stored number-like fields defined on the same entity, or a numeric literal.")
		return false
	}
	if !numberLikeField(field) {
		v.addDiagnostic(computed.Expression.Position, "INCOMPATIBLE_COMPUTED_FIELD", fmt.Sprintf("Computed field %s.%s references non-numeric field %s.%s.", entity.Name, computed.Name, entity.Name, operand), "Use number, integer, decimal, or money operands for computed fields in v0.2.")
		return false
	}
	return true
}

func (v *semanticValidator) validateEntityIndexes(entity EntityDecl, fieldIndex map[string]FieldDecl) {
	computedIndex := computedFieldIndex(entity)
	indexes := map[string]EntityIndexDecl{}
	for _, index := range entity.Indexes {
		if len(index.Fields) == 0 {
			v.addDiagnostic(index.Position, "INVALID_ENTITY_INDEX", fmt.Sprintf("Entity %s index has no fields.", entity.Name), "Use `index field` or `index fieldA, fieldB`.")
			continue
		}
		if len(index.Fields) > 4 {
			v.addDiagnostic(index.Position, "UNSUPPORTED_ENTITY_INDEX_FIELD_COUNT", fmt.Sprintf("Entity %s index has %d fields.", entity.Name, len(index.Fields)), "Use at most 4 stored fields in one index.")
		}

		indexKey := strings.Join(index.Fields, "\x00")
		if existing, ok := indexes[indexKey]; ok {
			v.addDiagnostic(index.Position, "DUPLICATE_ENTITY_INDEX", fmt.Sprintf("Entity %s repeats index %s.", entity.Name, strings.Join(index.Fields, ", ")), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		indexes[indexKey] = index

		fields := map[string]Position{}
		for _, fieldName := range index.Fields {
			if existing, ok := fields[fieldName]; ok {
				v.addDiagnostic(index.Position, "DUPLICATE_INDEX_FIELD", fmt.Sprintf("Entity %s index repeats field %s.", entity.Name, fieldName), fmt.Sprintf("First index field definition is at %s:%d.", existing.File, existing.Line))
				continue
			}
			fields[fieldName] = index.Position
			if _, ok := computedIndex[fieldName]; ok {
				v.addDiagnostic(index.Position, "UNSUPPORTED_COMPUTED_INDEX_FIELD", fmt.Sprintf("Entity %s index references computed field %s.", entity.Name, fieldName), "Computed fields are display-only values; index stored fields instead.")
				continue
			}
			if _, ok := fieldIndex[fieldName]; !ok {
				v.addDiagnostic(index.Position, "UNKNOWN_INDEX_FIELD", fmt.Sprintf("Entity %s index references unknown field %s.", entity.Name, fieldName), "Use stored fields declared on the same entity.")
			}
		}
	}
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
	case "accept":
		if modifier.Value == "" {
			v.addDiagnostic(field.Position, "MISSING_ACCEPT_VALUE", fmt.Sprintf("%s %s has accept without a value.", title(scope), fieldName), "Write accept followed by a MIME hint, such as `accept \"image/*\"`.")
			return
		}
		if field.Type != "file" && field.Type != "image" {
			v.addDiagnostic(field.Position, "UNSUPPORTED_ACCEPT_CONSTRAINT", fmt.Sprintf("%s %s uses accept on non-media type %s.", title(scope), fieldName, field.Type), "Use accept on file or image fields.")
			return
		}
		if field.Type == "image" && !strings.Contains(modifier.Value, "image/") {
			v.addDiagnostic(field.Position, "UNSUPPORTED_IMAGE_ACCEPT", fmt.Sprintf("%s %s uses image field accept value %q.", title(scope), fieldName, modifier.Value), "Use an image MIME hint such as `accept \"image/*\"`.")
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

func validExplicitAPIPath(path string) bool {
	if !strings.HasPrefix(path, "/api/") {
		return false
	}
	if strings.HasSuffix(path, "/") || strings.Contains(path, "//") {
		return false
	}
	if path == "/api/auth" || strings.HasPrefix(path, "/api/auth/") {
		return false
	}
	for _, char := range path {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= 'A' && char <= 'Z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		switch char {
		case '/', '-', '_', '.', '{', '}':
			continue
		default:
			return false
		}
	}
	stripped := regexp.MustCompile(`\{[A-Za-z][A-Za-z0-9_]*\}`).ReplaceAllString(path, "")
	return !strings.ContainsAny(stripped, "{}")
}

func normalizedAPIPathShape(path string) string {
	lowerPath := strings.ToLower(path)
	return regexp.MustCompile(`\{[a-z][a-z0-9_]*\}`).ReplaceAllString(lowerPath, "{}")
}

func generatedAPIRouteShapes(program Program) map[string]string {
	routes := map[string]string{}
	add := func(method string, path string, suggestion string) {
		routes[strings.ToUpper(method)+" "+normalizedAPIPathShape(path)] = suggestion
	}
	if program.Auth != nil {
		add("POST", "/api/auth/register", "Choose a distinct `/api/...` path outside generated auth routes.")
		add("POST", "/api/auth/login", "Choose a distinct `/api/...` path outside generated auth routes.")
		add("POST", "/api/auth/logout", "Choose a distinct `/api/...` path outside generated auth routes.")
		add("GET", "/api/auth/me", "Choose a distinct `/api/...` path outside generated auth routes.")
		if len(program.Roles) > 0 {
			add("GET", "/api/auth/users", "Choose a distinct `/api/...` path outside generated auth role-management routes.")
			add("GET", "/api/auth/audit", "Choose a distinct `/api/...` path outside generated auth audit routes.")
			add("PUT", "/api/auth/users/{id}/role", "Choose a distinct `/api/...` path outside generated auth role-management routes.")
		}
	}
	for _, page := range program.Pages {
		basePath := "/api/" + strings.ToLower(page.Name)
		itemPath := basePath + "/{id}"
		add("GET", basePath, "Choose a distinct `/api/...` path outside generated page CRUD routes.")
		add("GET", itemPath, "Choose a distinct `/api/...` path outside generated page CRUD routes.")
		if page.Query != "" {
			add("GET", basePath+"/query", "Choose a distinct `/api/...` path outside generated bound-query routes.")
			if query, ok := findQuery(program, page.Query); ok && len(query.Aggregates) > 0 {
				add("GET", basePath+"/query/summary", "Choose a distinct `/api/...` path outside generated bound-query summary routes.")
			}
		}
		if hasAction(page, "create") {
			add("POST", basePath, "Choose a distinct `/api/...` path outside generated page create routes.")
		}
		if hasAction(page, "edit") {
			add("PUT", itemPath, "Choose a distinct `/api/...` path outside generated page update routes.")
		}
		if hasAction(page, "delete") {
			add("DELETE", basePath, "Choose a distinct `/api/...` path outside generated page delete routes.")
			add("DELETE", itemPath, "Choose a distinct `/api/...` path outside generated page delete routes.")
		}
		if hasAction(page, "archive") {
			add("PATCH", itemPath+"/archive", "Choose a distinct `/api/...` path outside generated page archive routes.")
		}
		if hasAction(page, "restore") {
			add("PATCH", itemPath+"/restore", "Choose a distinct `/api/...` path outside generated page restore routes.")
		}
		for _, action := range customActionsForPage(program, page) {
			add("POST", itemPath+"/actions/"+strings.ToLower(action.Name), "Choose a distinct `/api/...` path outside generated custom action routes.")
		}
		if program.Auth != nil && len(program.Roles) > 0 {
			for _, workflow := range program.Workflows {
				if workflow.Source != page.Source {
					continue
				}
				for _, transition := range workflow.Transitions {
					add("POST", itemPath+"/workflow/"+transition.Name, "Choose a distinct `/api/...` path outside generated workflow transition routes.")
				}
			}
		}
	}
	return routes
}

func splitLabelTarget(target string) (string, string, bool) {
	parts := strings.Split(target, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	if !isThemeIdentifier(parts[0]) || !isThemeIdentifier(parts[1]) {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func fieldIndex(entity EntityDecl) map[string]FieldDecl {
	fields := map[string]FieldDecl{}
	for _, field := range entity.Fields {
		fields[field.Name] = field
	}
	return fields
}

func computedFieldIndex(entity EntityDecl) map[string]ComputedFieldDecl {
	fields := map[string]ComputedFieldDecl{}
	for _, field := range entity.ComputedFields {
		fields[field.Name] = field
	}
	return fields
}

func displayFieldExists(entity EntityDecl, name string) bool {
	if _, ok := fieldIndex(entity)[name]; ok {
		return true
	}
	if _, ok := computedFieldIndex(entity)[name]; ok {
		return true
	}
	return false
}

func (v *semanticValidator) validatePages(entityIndex map[string]EntityDecl, layoutIndex map[string]LayoutDecl, roleIndex map[string]RoleDecl, actionIndex map[string]CustomActionDecl, componentIndex map[string]ComponentDecl) map[string]PageDecl {
	pageIndex := map[string]PageDecl{}
	pageRoutes := map[string]PageDecl{}
	for _, page := range v.program.Pages {
		if existing, ok := pageIndex[page.Name]; ok {
			v.addDiagnostic(page.Position, "DUPLICATE_PAGE", fmt.Sprintf("Page %s is already defined.", page.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		pageIndex[page.Name] = page
		route := strings.ToLower(page.Name)
		if existing, ok := pageRoutes[route]; ok {
			v.addDiagnostic(page.Position, "DUPLICATE_PAGE_ROUTE", fmt.Sprintf("Page %s conflicts with page %s at generated route /%s.", page.Name, existing.Name, route), "Choose page names that remain distinct after lowercasing.")
		} else {
			pageRoutes[route] = page
		}

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

		v.validatePageView(page, source, componentIndex)
		v.validatePageFields(page, source)
		v.validateActions(page, source, actionIndex)
		v.validatePageUIIdentities(page)
		v.validatePageAccess(page, roleIndex)
	}
	return pageIndex
}

func (v *semanticValidator) validatePageView(page PageDecl, source EntityDecl, componentIndex map[string]ComponentDecl) {
	if page.View == nil {
		return
	}
	if len(page.View.Order) == 0 && page.View.Compose == nil && len(page.View.Sections) == 0 && len(page.View.Tabs) == 0 && len(page.View.Groups) == 0 && len(page.View.Triggers) == 0 {
		v.addDiagnostic(page.View.Position, "MISSING_VIEW_ORDER", fmt.Sprintf("Page %s view block is missing order.", page.Name), "Add `order table, detail, form` or remove the empty view block.")
		return
	}

	seen := map[string]Position{}
	for _, section := range page.View.Order {
		if !pageViewSectionSupported(page, section) {
			v.addDiagnostic(page.View.Position, "UNSUPPORTED_VIEW_SECTION", fmt.Sprintf("Page %s view uses unsupported section %q.", page.Name, section), "Use table, detail, form, or declare a component section such as `section StockSummary component StockBadge bind selected`.")
			continue
		}
		if existing, ok := seen[section]; ok {
			v.addDiagnostic(page.View.Position, "DUPLICATE_VIEW_SECTION", fmt.Sprintf("Page %s view repeats section %s.", page.Name, section), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
			continue
		}
		seen[section] = page.View.Position
	}
	if page.View.Compose != nil {
		v.validateViewCompose(page, *page.View.Compose)
	}
	v.validateViewSections(page, source, componentIndex)
	v.validateViewGroups(page)
	v.validateViewTabs(page)
	v.validateViewTriggers(page)
}

func (v *semanticValidator) validateViewCompose(page PageDecl, compose ViewComposeDecl) {
	if !supportedViewComposeModes[compose.Mode] {
		v.addDiagnostic(compose.Position, "UNSUPPORTED_VIEW_COMPOSE_MODE", fmt.Sprintf("Page %s view compose mode %q is not supported.", page.Name, compose.Mode), "Use stack, grid, or tabs.")
	}
	if compose.Columns < 0 || compose.Columns > 4 {
		v.addDiagnostic(compose.Position, "UNSUPPORTED_VIEW_COMPOSE_COLUMNS", fmt.Sprintf("Page %s view compose columns value %d is not supported.", page.Name, compose.Columns), "Use columns 1, 2, 3, or 4.")
	}
	if compose.Mode == "stack" && compose.Columns > 0 {
		v.addDiagnostic(compose.Position, "UNSUPPORTED_VIEW_COMPOSE_COLUMNS", fmt.Sprintf("Page %s stack compose cannot use columns.", page.Name), "Remove columns or use compose grid.")
	}
	if compose.Mode == "tabs" && compose.Columns > 0 {
		v.addDiagnostic(compose.Position, "UNSUPPORTED_VIEW_COMPOSE_COLUMNS", fmt.Sprintf("Page %s tabs compose cannot use columns.", page.Name), "Remove columns from compose tabs.")
	}
	if compose.Mode == "tabs" && compose.StackAt != "" {
		v.addDiagnostic(compose.Position, "UNSUPPORTED_VIEW_STACK_AT", fmt.Sprintf("Page %s tabs compose cannot use stackAt.", page.Name), "Remove stackAt from compose tabs.")
	}
	if compose.Gap != "" && !supportedViewGaps[compose.Gap] {
		v.addDiagnostic(compose.Position, "UNSUPPORTED_VIEW_GAP", fmt.Sprintf("Page %s view gap %q is not supported.", page.Name, compose.Gap), "Use gap sm, gap md, or gap lg.")
	}
	if compose.StackAt != "" && !supportedViewBreakpoints[compose.StackAt] {
		v.addDiagnostic(compose.Position, "UNSUPPORTED_VIEW_STACK_AT", fmt.Sprintf("Page %s view stackAt %q is not supported.", page.Name, compose.StackAt), "Use stackAt sm, stackAt md, stackAt lg, or stackAt none.")
	}
}

func (v *semanticValidator) validateViewSections(page PageDecl, source EntityDecl, componentIndex map[string]ComponentDecl) {
	seen := map[string]Position{}
	for _, section := range page.View.Sections {
		if existing, ok := seen[section.Name]; ok {
			v.addDiagnostic(section.Position, "DUPLICATE_VIEW_SECTION", fmt.Sprintf("Page %s view repeats section %s.", page.Name, section.Name), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
			continue
		}
		if section.Component != "" {
			v.validateViewComponentSection(page, source, section, componentIndex)
		} else {
			if !supportedViewSections[section.Name] {
				v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_SECTION", fmt.Sprintf("Page %s view section uses unsupported section %q.", page.Name, section.Name), "Use table, detail, form, or add `component <Component> bind selected` to declare a component section.")
				continue
			}
			if section.Bind != "" {
				v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_SECTION_BIND", fmt.Sprintf("Page %s built-in view section %s cannot use bind %q.", page.Name, section.Name, section.Bind), "Use bind only on component sections.")
			}
		}
		if section.Span != 0 && (section.Span < 1 || section.Span > 4) {
			v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_SECTION_SPAN", fmt.Sprintf("Page %s view section %s uses unsupported span %d.", page.Name, section.Name, section.Span), "Use span 1, 2, 3, or 4.")
		}
		if section.Display != "" && !supportedViewSectionDisplays[section.Display] {
			v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_SECTION_DISPLAY", fmt.Sprintf("Page %s view section %s uses unsupported display %q.", page.Name, section.Name, section.Display), "Use display inline, display modal, or display drawer.")
		}
		if (section.Display == "modal" || section.Display == "drawer") && section.Name == "table" {
			v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_SECTION_DISPLAY", fmt.Sprintf("Page %s table section cannot use display %s.", page.Name, section.Display), "Use modal or drawer for detail or form sections.")
		}
		if section.Display == "modal" || section.Display == "drawer" {
			if page.View.Compose != nil && page.View.Compose.Mode == "tabs" {
				v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_SECTION_DISPLAY", fmt.Sprintf("Page %s cannot combine compose tabs with section %s display %s.", page.Name, section.Name, section.Display), "Use either compose tabs or modal/drawer section display in this MVP.")
			}
		}
		if section.Side != "" {
			if section.Display != "drawer" {
				v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_SECTION_SIDE", fmt.Sprintf("Page %s view section %s declares side without display drawer.", page.Name, section.Name), "Use side left/right only with display drawer.")
			} else if !supportedViewSectionSides[section.Side] {
				v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_SECTION_SIDE", fmt.Sprintf("Page %s view section %s uses unsupported drawer side %q.", page.Name, section.Name, section.Side), "Use side left or side right.")
			}
		}
		seen[section.Name] = section.Position
	}
}

func (v *semanticValidator) validateViewComponentSection(page PageDecl, source EntityDecl, section ViewSectionDecl, componentIndex map[string]ComponentDecl) {
	if supportedViewSections[section.Name] {
		v.addDiagnostic(section.Position, "CONFLICTING_VIEW_COMPONENT_SECTION", fmt.Sprintf("Page %s view section %s is a built-in section and cannot also render component %s.", page.Name, section.Name, section.Component), "Use a custom section name such as `StockSummary`.")
	}
	if !isThemeIdentifier(section.Name) {
		v.addDiagnostic(section.Position, "INVALID_VIEW_COMPONENT_SECTION", fmt.Sprintf("Page %s component section name %q is invalid.", page.Name, section.Name), "Use letters, numbers, underscores, or hyphens; start with a letter or underscore.")
	}
	component, ok := componentIndex[section.Component]
	if !ok {
		v.addDiagnostic(section.Position, "UNKNOWN_VIEW_COMPONENT", fmt.Sprintf("Page %s view section %s references unknown component %s.", page.Name, section.Name, section.Component), "Declare the component before using it in a view section.")
		return
	}
	if section.Bind == "" {
		v.addDiagnostic(section.Position, "MISSING_VIEW_COMPONENT_BIND", fmt.Sprintf("Page %s component section %s is missing bind.", page.Name, section.Name), "Use bind selected, bind first, or bind each.")
	} else if !supportedViewComponentBinds[section.Bind] {
		v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_COMPONENT_BIND", fmt.Sprintf("Page %s component section %s uses unsupported bind %q.", page.Name, section.Name, section.Bind), "Use bind selected, bind first, or bind each.")
	}
	if section.Display == "modal" || section.Display == "drawer" {
		v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_COMPONENT_DISPLAY", fmt.Sprintf("Page %s component section %s cannot use display %s.", page.Name, section.Name, section.Display), "Component sections render inline in this MVP.")
	}
	if section.Side != "" {
		v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_COMPONENT_SIDE", fmt.Sprintf("Page %s component section %s declares drawer side %q.", page.Name, section.Name, section.Side), "Remove side; component sections render inline in this MVP.")
	}

	fields := fieldIndex(source)
	computedFields := computedFieldIndex(source)
	for _, input := range component.Inputs {
		if input.List {
			v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_COMPONENT_INPUT", fmt.Sprintf("Page %s component section %s cannot bind list input %s.%s.", page.Name, section.Name, component.Name, input.Name), "Use scalar primitive component inputs for page component sections in this MVP.")
			continue
		}
		if !supportedFieldTypes[input.Type] {
			v.addDiagnostic(section.Position, "UNSUPPORTED_VIEW_COMPONENT_INPUT", fmt.Sprintf("Page %s component section %s cannot bind entity input %s.%s.", page.Name, section.Name, component.Name, input.Name), "Use scalar primitive component inputs that match stored or computed fields on the page source entity.")
			continue
		}
		if field, ok := fields[input.Name]; ok {
			if field.Type != input.Type {
				v.addDiagnostic(section.Position, "VIEW_COMPONENT_INPUT_TYPE_MISMATCH", fmt.Sprintf("Page %s component section %s input %s expects %s but %s.%s is %s.", page.Name, section.Name, input.Name, input.Type, source.Name, field.Name, field.Type), "Match the component input type to the source field type.")
			}
			continue
		}
		if computed, ok := computedFields[input.Name]; ok {
			if computed.Type != input.Type {
				v.addDiagnostic(section.Position, "VIEW_COMPONENT_INPUT_TYPE_MISMATCH", fmt.Sprintf("Page %s component section %s input %s expects %s but computed field %s.%s is %s.", page.Name, section.Name, input.Name, input.Type, source.Name, computed.Name, computed.Type), "Match the component input type to the computed display field type.")
			}
			continue
		}
		v.addDiagnostic(section.Position, "UNKNOWN_VIEW_COMPONENT_INPUT_FIELD", fmt.Sprintf("Page %s component section %s input %s has no matching field on %s.", page.Name, section.Name, input.Name, source.Name), "Add a stored or computed field with the same name, or use a component whose inputs match the page source.")
	}
}

func (v *semanticValidator) validateViewGroups(page PageDecl) {
	seenGroups := map[string]Position{}
	sectionOwners := map[string]Position{}
	for _, group := range page.View.Groups {
		if existing, ok := seenGroups[group.Name]; ok {
			v.addDiagnostic(group.Position, "DUPLICATE_VIEW_GROUP", fmt.Sprintf("Page %s view repeats group %s.", page.Name, group.Name), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
			continue
		}
		seenGroups[group.Name] = group.Position

		if page.View.Compose != nil && page.View.Compose.Mode == "tabs" {
			v.addDiagnostic(group.Position, "UNSUPPORTED_VIEW_GROUP", fmt.Sprintf("Page %s cannot combine compose tabs with view group %s.", page.Name, group.Name), "Use either compose tabs or view groups in this MVP.")
		}
		if len(group.Sections) == 0 {
			v.addDiagnostic(group.Position, "MISSING_VIEW_GROUP_SECTION", fmt.Sprintf("Page %s view group %s has no sections.", page.Name, group.Name), "Add supported sections such as detail, form.")
		}
		if group.Span != 0 && (group.Span < 1 || group.Span > 4) {
			v.addDiagnostic(group.Position, "UNSUPPORTED_VIEW_GROUP_SPAN", fmt.Sprintf("Page %s view group %s uses unsupported span %d.", page.Name, group.Name, group.Span), "Use span 1, 2, 3, or 4.")
		}
		if group.Compose != nil {
			v.validateViewGroupCompose(page, group, *group.Compose)
		}

		seenInGroup := map[string]bool{}
		for _, sectionName := range group.Sections {
			if !pageViewSectionSupported(page, sectionName) {
				v.addDiagnostic(group.Position, "UNSUPPORTED_VIEW_GROUP_SECTION", fmt.Sprintf("Page %s view group %s uses unsupported section %q.", page.Name, group.Name, sectionName), "Use table, detail, form, or a declared component section.")
				continue
			}
			if seenInGroup[sectionName] {
				v.addDiagnostic(group.Position, "DUPLICATE_VIEW_GROUP_SECTION", fmt.Sprintf("Page %s view group %s repeats section %s.", page.Name, group.Name, sectionName), "Keep each section once inside a group.")
				continue
			}
			seenInGroup[sectionName] = true
			if existing, ok := sectionOwners[sectionName]; ok {
				v.addDiagnostic(group.Position, "DUPLICATE_VIEW_GROUP_SECTION", fmt.Sprintf("Page %s view section %s appears in more than one group.", page.Name, sectionName), fmt.Sprintf("First group section is at %s:%d.", existing.File, existing.Line))
				continue
			}
			sectionOwners[sectionName] = group.Position
			if section, ok := pageViewSection(page, sectionName); ok && (section.Display == "modal" || section.Display == "drawer") {
				v.addDiagnostic(group.Position, "UNSUPPORTED_VIEW_GROUP_SECTION_DISPLAY", fmt.Sprintf("Page %s view group %s contains overlay section %s.", page.Name, group.Name, sectionName), "Keep grouped sections inline in this MVP.")
			}
		}
		if !viewGroupSectionsAreContiguous(page, group) {
			v.addDiagnostic(group.Position, "UNSUPPORTED_VIEW_GROUP_SECTION", fmt.Sprintf("Page %s view group %s uses non-contiguous sections.", page.Name, group.Name), "Group table/detail, detail/form, or all table/detail/form together in this MVP.")
		}
	}
}

func viewGroupSectionsAreContiguous(page PageDecl, group ViewGroupDecl) bool {
	positions := []int{}
	indexBySection := map[string]int{}
	for index, sectionName := range pageViewOrder(page) {
		indexBySection[sectionName] = index
	}
	seen := map[int]bool{}
	for _, sectionName := range group.Sections {
		index, ok := indexBySection[sectionName]
		if !ok || seen[index] {
			continue
		}
		positions = append(positions, index)
		seen[index] = true
	}
	if len(positions) <= 1 {
		return true
	}
	minPosition := positions[0]
	maxPosition := positions[0]
	for _, position := range positions[1:] {
		if position < minPosition {
			minPosition = position
		}
		if position > maxPosition {
			maxPosition = position
		}
	}
	return maxPosition-minPosition+1 == len(positions)
}

func (v *semanticValidator) validateViewGroupCompose(page PageDecl, group ViewGroupDecl, compose ViewComposeDecl) {
	if !supportedViewGroupComposeModes[compose.Mode] {
		v.addDiagnostic(group.Position, "UNSUPPORTED_VIEW_GROUP_COMPOSE_MODE", fmt.Sprintf("Page %s view group %s compose mode %q is not supported.", page.Name, group.Name, compose.Mode), "Use compose stack or compose grid inside a group.")
	}
	if compose.Columns < 0 || compose.Columns > 4 {
		v.addDiagnostic(group.Position, "UNSUPPORTED_VIEW_GROUP_COMPOSE_COLUMNS", fmt.Sprintf("Page %s view group %s columns value %d is not supported.", page.Name, group.Name, compose.Columns), "Use columns 1, 2, 3, or 4.")
	}
	if compose.Mode == "stack" && compose.Columns > 0 {
		v.addDiagnostic(group.Position, "UNSUPPORTED_VIEW_GROUP_COMPOSE_COLUMNS", fmt.Sprintf("Page %s view group %s stack compose cannot use columns.", page.Name, group.Name), "Remove columns or use compose grid.")
	}
	if compose.Gap != "" && !supportedViewGaps[compose.Gap] {
		v.addDiagnostic(group.Position, "UNSUPPORTED_VIEW_GROUP_GAP", fmt.Sprintf("Page %s view group %s gap %q is not supported.", page.Name, group.Name, compose.Gap), "Use gap sm, gap md, or gap lg.")
	}
}

func (v *semanticValidator) validateViewTabs(page PageDecl) {
	if len(page.View.Tabs) == 0 {
		if page.View.Compose != nil && page.View.Compose.Mode == "tabs" {
			v.addDiagnostic(page.View.Compose.Position, "MISSING_VIEW_TABS", fmt.Sprintf("Page %s uses compose tabs without tab declarations.", page.Name), "Add `tab List sections table` lines inside the view block.")
		}
		return
	}
	if page.View.Compose == nil || page.View.Compose.Mode != "tabs" {
		v.addDiagnostic(page.View.Tabs[0].Position, "UNSUPPORTED_VIEW_TAB", fmt.Sprintf("Page %s declares tabs without compose tabs.", page.Name), "Add `compose tabs` before tab declarations.")
	}

	orderedSections := pageViewOrder(page)
	expectedSections := map[string]bool{}
	for _, section := range orderedSections {
		if pageViewSectionSupported(page, section) {
			expectedSections[section] = true
		}
	}
	seenTabs := map[string]Position{}
	sectionOwners := map[string]Position{}
	for _, tab := range page.View.Tabs {
		if existing, ok := seenTabs[tab.Name]; ok {
			v.addDiagnostic(tab.Position, "DUPLICATE_VIEW_TAB", fmt.Sprintf("Page %s view repeats tab %s.", page.Name, tab.Name), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
			continue
		}
		seenTabs[tab.Name] = tab.Position
		seenInTab := map[string]bool{}
		for _, section := range tab.Sections {
			if !pageViewSectionSupported(page, section) {
				v.addDiagnostic(tab.Position, "UNSUPPORTED_VIEW_TAB_SECTION", fmt.Sprintf("Page %s view tab %s uses unsupported section %q.", page.Name, tab.Name, section), "Use table, detail, form, or a declared component section.")
				continue
			}
			if seenInTab[section] {
				v.addDiagnostic(tab.Position, "DUPLICATE_VIEW_TAB_SECTION", fmt.Sprintf("Page %s view tab %s repeats section %s.", page.Name, tab.Name, section), "Keep each section once inside a tab.")
				continue
			}
			seenInTab[section] = true
			if existing, ok := sectionOwners[section]; ok {
				v.addDiagnostic(tab.Position, "DUPLICATE_VIEW_TAB_SECTION", fmt.Sprintf("Page %s view section %s appears in more than one tab.", page.Name, section), fmt.Sprintf("First tab section is at %s:%d.", existing.File, existing.Line))
				continue
			}
			sectionOwners[section] = tab.Position
		}
	}
	if page.View.Compose == nil || page.View.Compose.Mode != "tabs" {
		return
	}
	for _, section := range orderedSections {
		if !expectedSections[section] {
			continue
		}
		if _, ok := sectionOwners[section]; !ok {
			v.addDiagnostic(page.View.Position, "MISSING_VIEW_TAB_SECTION", fmt.Sprintf("Page %s compose tabs does not include section %s.", page.Name, section), "Assign every ordered section to exactly one tab.")
		}
	}
}

func (v *semanticValidator) validateViewTriggers(page PageDecl) {
	seen := map[string]Position{}
	for _, trigger := range page.View.Triggers {
		if !supportedViewSections[trigger.Section] {
			v.addDiagnostic(trigger.Position, "UNSUPPORTED_VIEW_TRIGGER_SECTION", fmt.Sprintf("Page %s view trigger targets unsupported section %q.", page.Name, trigger.Section), "Use table, detail, or form.")
			continue
		}
		if !supportedViewTriggerEvents[trigger.Event] {
			v.addDiagnostic(trigger.Position, "UNSUPPORTED_VIEW_TRIGGER_EVENT", fmt.Sprintf("Page %s view trigger uses unsupported event %q.", page.Name, trigger.Event), "Use rowSelect, createStart, editStart, saveSuccess, or close.")
			continue
		}
		if existing, ok := seen[trigger.Event]; ok {
			v.addDiagnostic(trigger.Position, "DUPLICATE_VIEW_TRIGGER", fmt.Sprintf("Page %s repeats trigger event %s.", page.Name, trigger.Event), fmt.Sprintf("First trigger is at %s:%d. Keep one target section per event.", existing.File, existing.Line))
			continue
		}
		seen[trigger.Event] = trigger.Position

		if !viewTriggerCombinationSupported(trigger.Section, trigger.Event) {
			v.addDiagnostic(trigger.Position, "UNSUPPORTED_VIEW_TRIGGER", fmt.Sprintf("Page %s cannot trigger section %s on %s.", page.Name, trigger.Section, trigger.Event), "Use trigger detail|form on rowSelect, trigger form on createStart|editStart, trigger table|detail on saveSuccess, or trigger table on close.")
			continue
		}
		if !viewTriggerActionsSupported(page, trigger) {
			v.addDiagnostic(trigger.Position, "UNSUPPORTED_VIEW_TRIGGER_ACTION", fmt.Sprintf("Page %s trigger %s on %s has no matching page action.", page.Name, trigger.Section, trigger.Event), "Add the matching create/edit action or remove the trigger.")
		}
	}
}

func viewTriggerCombinationSupported(section string, event string) bool {
	switch event {
	case "rowSelect":
		return section == "detail" || section == "form"
	case "createStart":
		return section == "form"
	case "editStart":
		return section == "form"
	case "saveSuccess":
		return section == "table" || section == "detail"
	case "close":
		return section == "table"
	default:
		return false
	}
}

func viewTriggerActionsSupported(page PageDecl, trigger ViewTriggerDecl) bool {
	switch trigger.Event {
	case "createStart":
		return hasAction(page, "create")
	case "editStart":
		return hasAction(page, "edit")
	case "rowSelect":
		if trigger.Section == "form" {
			return hasAction(page, "edit")
		}
		return true
	case "saveSuccess":
		return hasAction(page, "create") || hasAction(page, "edit")
	default:
		return true
	}
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
	fields := fieldIndex(source)
	computedFields := computedFieldIndex(source)
	policyFields := entityPolicyFieldMap(source)

	for _, column := range page.Table.Columns {
		if _, ok := fields[column]; ok {
			continue
		}
		if _, ok := computedFields[column]; ok {
			continue
		}
		v.addDiagnostic(page.Position, "UNKNOWN_TABLE_COLUMN", fmt.Sprintf("Page %s table uses unknown field %s.%s.", page.Name, source.Name, column), "Add the field to the source entity or remove it from columns.")
	}

	for _, fieldName := range page.Form.Fields {
		if _, ok := fields[fieldName]; ok {
			if policy, policyOK := policyFields[fieldName]; policyOK {
				v.addDiagnostic(page.Position, "UNSUPPORTED_POLICY_FORM_FIELD", fmt.Sprintf("Page %s form includes %s policy field %s.%s.", page.Name, policy.Kind, source.Name, fieldName), "Remove policy fields from forms; generated create/update routes stamp them from the authenticated user.")
			}
			continue
		}
		if _, ok := computedFields[fieldName]; ok {
			v.addDiagnostic(page.Position, "UNSUPPORTED_COMPUTED_FORM_FIELD", fmt.Sprintf("Page %s form uses computed field %s.%s as an input.", page.Name, source.Name, fieldName), "Computed fields are read-only display values; remove them from form fields.")
			continue
		}
		v.addDiagnostic(page.Position, "UNKNOWN_FORM_FIELD", fmt.Sprintf("Page %s form uses unknown field %s.%s.", page.Name, source.Name, fieldName), "Add the field to the source entity or remove it from fields.")
	}

	for _, fieldName := range page.Table.Search {
		if _, ok := computedFields[fieldName]; ok {
			v.addDiagnostic(page.Position, "UNSUPPORTED_COMPUTED_SEARCH_FIELD", fmt.Sprintf("Page %s search uses computed field %s.%s.", page.Name, source.Name, fieldName), "Search computed display values in a later data logic phase; use stored text, email, or entity reference fields for now.")
			continue
		}
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
		if _, ok := computedFields[fieldName]; ok {
			v.addDiagnostic(page.Position, "UNSUPPORTED_COMPUTED_FILTER_FIELD", fmt.Sprintf("Page %s filter uses computed field %s.%s.", page.Name, source.Name, fieldName), "Filter computed display values in a later data logic phase; use stored fields for now.")
			continue
		}
		if _, ok := fields[fieldName]; !ok {
			v.addDiagnostic(page.Position, "UNKNOWN_FILTER_FIELD", fmt.Sprintf("Page %s filter uses unknown field %s.%s.", page.Name, source.Name, fieldName), "Add the field to the source entity or remove it from filter.")
		}
	}

	if page.Table.Sort.Field != "" {
		if _, ok := computedFields[page.Table.Sort.Field]; ok {
			v.addDiagnostic(page.Position, "UNSUPPORTED_COMPUTED_SORT_FIELD", fmt.Sprintf("Page %s sort uses computed field %s.%s.", page.Name, source.Name, page.Table.Sort.Field), "Sort computed display values in a later data logic phase; use stored fields for now.")
		} else if _, ok := fields[page.Table.Sort.Field]; !ok {
			v.addDiagnostic(page.Position, "UNKNOWN_SORT_FIELD", fmt.Sprintf("Page %s table sorts by unknown field %s.%s.", page.Name, source.Name, page.Table.Sort.Field), "Add the field to the source entity or change the sort field.")
		}
		if page.Table.Sort.Direction != "asc" && page.Table.Sort.Direction != "desc" {
			v.addDiagnostic(page.Position, "UNSUPPORTED_SORT_DIRECTION", fmt.Sprintf("Page %s table uses unsupported sort direction %q.", page.Name, page.Table.Sort.Direction), "Use asc or desc.")
		}
	}

	if page.Table.Paginate < 0 {
		v.addDiagnostic(page.Position, "UNSUPPORTED_PAGE_SIZE", fmt.Sprintf("Page %s table uses unsupported page size %d.", page.Name, page.Table.Paginate), "Use a positive whole number, such as `paginate 25`.")
	}

	v.validateInlineUI("table", page.Name+".table", page.Table.UI, inlineUIModesForTarget("table"))
	v.validateInlineUI("form", page.Name+".form", page.Form.UI, inlineUIModesForTarget("form"))
}

func (v *semanticValidator) validateActions(page PageDecl, source EntityDecl, actionIndex map[string]CustomActionDecl) {
	actions := map[string]bool{}
	for _, action := range page.Actions {
		if actions[action] {
			v.addDiagnostic(page.Position, "DUPLICATE_PAGE_ACTION", fmt.Sprintf("Page %s repeats action %s.", page.Name, action), "Keep each page action once.")
			continue
		}
		actions[action] = true
		customAction, custom := actionIndex[action]
		if custom {
			if customAction.Source != source.Name {
				v.addDiagnostic(page.Position, "PAGE_ACTION_SOURCE_MISMATCH", fmt.Sprintf("Page %s source %s does not match action %s source %s.", page.Name, source.Name, customAction.Name, customAction.Source), "Use a custom action bound to the page source entity.")
			}
			continue
		}
		if !supportedActions[action] {
			v.addDiagnostic(page.Position, "UNSUPPORTED_ACTION", fmt.Sprintf("Page %s uses unsupported action %q.", page.Name, action), "Use create, edit, delete, archive, restore, or a declared PascalCase custom action.")
		}
	}

	actionModeIndex := map[string]Position{}
	for _, actionUI := range page.ActionUI {
		if !actions[actionUI.Action] {
			v.addDiagnostic(actionUI.Position, "UNKNOWN_ACTION_UI", fmt.Sprintf("Page %s declares UI for action %s but that action is not listed.", page.Name, actionUI.Action), fmt.Sprintf("Add `%s` to the page actions list, or remove the action UI line.", actionUI.Action))
		}
		v.validateInlineUI("button", page.Name+"."+actionUI.Action, actionUI.UI, inlineUIModesForTarget("button"))
		for _, intent := range actionUI.UI {
			key := actionUI.Action + "." + intent.Mode
			if existing, ok := actionModeIndex[key]; ok {
				if existing.File == intent.Position.File && existing.Line == intent.Position.Line && existing.Column == intent.Position.Column {
					continue
				}
				v.addDiagnostic(intent.Position, "DUPLICATE_UI_INTENT", fmt.Sprintf("Page %s action %s repeats UI mode %s.", page.Name, actionUI.Action, intent.Mode), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
				continue
			}
			actionModeIndex[key] = intent.Position
		}
	}
}

func (v *semanticValidator) validatePageUIIdentities(page PageDecl) {
	seenIDs := map[string]Position{}
	identities := []struct {
		scope    string
		name     string
		identity *UIIdentity
	}{
		{scope: "table", name: page.Name + ".table", identity: page.Table.Identity},
		{scope: "form", name: page.Name + ".form", identity: page.Form.Identity},
	}

	for _, actionUI := range page.ActionUI {
		identities = append(identities, struct {
			scope    string
			name     string
			identity *UIIdentity
		}{scope: "action", name: page.Name + "." + actionUI.Action, identity: actionUI.Identity})
	}

	for _, item := range identities {
		v.validateUIIdentity(item.scope, item.name, item.identity)
		if item.identity == nil || item.identity.ID == "" {
			continue
		}
		normalized := kebabCase(item.identity.ID)
		if existing, ok := seenIDs[normalized]; ok {
			v.addDiagnostic(item.identity.Position, "DUPLICATE_UI_ID", fmt.Sprintf("%s %s repeats generated UI id %s.", title(item.scope), item.name, normalized), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
			continue
		}
		seenIDs[normalized] = item.identity.Position
	}
}

func (v *semanticValidator) validateUIIdentity(scope string, name string, identity *UIIdentity) {
	if identity == nil {
		return
	}
	if identity.ID != "" && !isThemeIdentifier(identity.ID) {
		v.addDiagnostic(identity.Position, "INVALID_UI_ID", fmt.Sprintf("%s %s uses invalid UI id %q.", title(scope), name, identity.ID), "Use letters, numbers, underscores, or hyphens; start with a letter or underscore.")
	}

	seenClasses := map[string]Position{}
	for _, className := range identity.Classes {
		if !isThemeIdentifier(className) {
			v.addDiagnostic(identity.Position, "INVALID_UI_CLASS", fmt.Sprintf("%s %s uses invalid UI class %q.", title(scope), name, className), "Use letters, numbers, underscores, or hyphens; start with a letter or underscore.")
			continue
		}
		normalized := kebabCase(className)
		if existing, ok := seenClasses[normalized]; ok {
			v.addDiagnostic(identity.Position, "DUPLICATE_UI_CLASS", fmt.Sprintf("%s %s repeats generated UI class %s.", title(scope), name, normalized), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
			continue
		}
		seenClasses[normalized] = identity.Position
	}
}

func inlineUIModesForTarget(target string) map[string]bool {
	switch target {
	case "field":
		return setOf("box", "text")
	case "form":
		return setOf("box", "text", "button")
	case "table":
		return setOf("box", "text", "table")
	case "button":
		return setOf("button")
	default:
		return supportedInlineUIModes
	}
}

func (v *semanticValidator) validateInlineUI(scope string, name string, intents []UIIntent, allowedModes map[string]bool) {
	seenModes := map[string]Position{}
	for _, intent := range intents {
		if existing, ok := seenModes[intent.Mode]; ok {
			v.addDiagnostic(intent.Position, "DUPLICATE_UI_INTENT", fmt.Sprintf("%s %s repeats UI mode %s.", title(scope), name, intent.Mode), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
			continue
		}
		seenModes[intent.Mode] = intent.Position

		if !supportedInlineUIModes[intent.Mode] {
			v.addDiagnostic(intent.Position, "UNSUPPORTED_UI_MODE", fmt.Sprintf("%s %s uses unsupported UI mode %q.", title(scope), name, intent.Mode), "Use box, text, table, or button.")
			continue
		}
		if !allowedModes[intent.Mode] {
			v.addDiagnostic(intent.Position, "UNSUPPORTED_UI_TARGET_MODE", fmt.Sprintf("%s %s cannot use UI mode %s.", title(scope), name, intent.Mode), fmt.Sprintf("Use a UI mode that applies to %s.", scope))
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
