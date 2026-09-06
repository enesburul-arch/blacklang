package main

import (
	"fmt"
	"strings"
)

func AnalyzeAffected(program Program, symbol string) (AffectedAnalysis, []Diagnostic) {
	symbol = strings.TrimSpace(symbol)
	analysis := newAffectedAnalysis(symbol)
	if symbol == "" {
		return analysis, []Diagnostic{{
			Code:       "MISSING_AFFECTED_SYMBOL",
			Message:    "`--affected` requires a symbol.",
			Suggestion: "Use an entity, field, page, role, workflow, state, component, api, target, deploy, or app symbol such as `Product`, `Product.stock`, `Products`, or `target`.",
		}}
	}

	if strings.Contains(symbol, ".") {
		entityName, fieldName, ok := strings.Cut(symbol, ".")
		entity, entityOK := affectedFindEntity(program, entityName)
		if ok && entityOK {
			if field, fieldOK := findField(entity, fieldName); fieldOK {
				analysis.Kind = "field"
				analysis.Found = true
				analysis.Entity = entity.Name
				analysis.Field = field.Name
				populateEntityAffected(&analysis, program, entity, &field)
				return analysis, nil
			}
			if field, fieldOK := findComputedField(entity, fieldName); fieldOK {
				analysis.Kind = "computed-field"
				analysis.Found = true
				analysis.Entity = entity.Name
				analysis.Field = field.Name
				populateComputedFieldAffected(&analysis, program, entity, field)
				return analysis, nil
			}
		}
		return analysis, []Diagnostic{{
			Code:       "UNKNOWN_AFFECTED_SYMBOL",
			Message:    fmt.Sprintf("Affected symbol %q was not found.", symbol),
			Suggestion: "Use an existing entity field or computed-field symbol such as `Product.stock`, or run `black inspect --json` to see available names.",
		}}
	}

	if entity, ok := affectedFindEntity(program, symbol); ok {
		analysis.Kind = "entity"
		analysis.Found = true
		analysis.Entity = entity.Name
		populateEntityAffected(&analysis, program, entity, nil)
		return analysis, nil
	}
	if page, ok := affectedFindPage(program, symbol); ok {
		analysis.Kind = "page"
		analysis.Found = true
		populatePageAffected(&analysis, program, page)
		return analysis, nil
	}
	if role, ok := affectedFindRole(program, symbol); ok {
		analysis.Kind = "role"
		analysis.Found = true
		populateRoleAffected(&analysis, program, role)
		return analysis, nil
	}
	if workflow, ok := affectedFindWorkflow(program, symbol); ok {
		analysis.Kind = "workflow"
		analysis.Found = true
		populateWorkflowAffected(&analysis, program, workflow)
		return analysis, nil
	}
	if state, ok := affectedFindState(program, symbol); ok {
		analysis.Kind = "state"
		analysis.Found = true
		populateStateAffected(&analysis, program, state)
		return analysis, nil
	}
	if component, ok := affectedFindComponent(program, symbol); ok {
		analysis.Kind = "component"
		analysis.Found = true
		populateComponentAffected(&analysis, program, component)
		return analysis, nil
	}
	if api, ok := affectedFindAPI(program, symbol); ok {
		analysis.Kind = "api"
		analysis.Found = true
		populateAPIAffected(&analysis, api)
		return analysis, nil
	}
	if symbol == "auth" {
		analysis.Kind = "auth"
		analysis.Found = program.Auth != nil
		populateAuthAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "auth")
	}
	if symbol == "database" {
		analysis.Kind = "database"
		analysis.Found = program.Database != nil
		populateDatabaseAffected(&analysis)
		return analysis, missingTopLevelDiagnostic(analysis, "database")
	}
	if symbol == "security" || symbol == "cors" {
		analysis.Kind = "security"
		analysis.Found = program.Security != nil
		populateSecurityAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "security")
	}
	if symbol == "deploy" {
		analysis.Kind = "deploy"
		analysis.Found = program.Deploy != nil
		populateDeployAffected(&analysis)
		return analysis, missingTopLevelDiagnostic(analysis, "deploy")
	}
	if symbol == "target" {
		analysis.Kind = "target"
		analysis.Found = program.Target != nil
		populateTargetAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "target")
	}
	if symbol == "app" || symbol == program.App.Name {
		analysis.Kind = "app"
		analysis.Found = program.App.Name != ""
		populateAppAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "app")
	}

	return analysis, []Diagnostic{{
		Code:       "UNKNOWN_AFFECTED_SYMBOL",
		Message:    fmt.Sprintf("Affected symbol %q was not found.", symbol),
		Suggestion: "Use an existing entity, field, page, role, workflow, state, component, api, auth, database, security, deploy, target, or app symbol.",
	}}
}

func newAffectedAnalysis(symbol string) AffectedAnalysis {
	return AffectedAnalysis{
		Symbol:         symbol,
		Kind:           "unknown",
		Entities:       []AffectedItem{},
		Pages:          []AffectedItem{},
		Roles:          []AffectedItem{},
		Workflows:      []AffectedItem{},
		States:         []AffectedItem{},
		Components:     []AffectedItem{},
		APIs:           []AffectedItem{},
		GeneratedFiles: []AffectedItem{},
		AgentNotes: []string{
			"Use this affected graph before editing generated output.",
			"Validate and build after changing any listed source symbol.",
		},
	}
}

func populateEntityAffected(analysis *AffectedAnalysis, program Program, entity EntityDecl, field *FieldDecl) {
	fieldName := ""
	if field != nil {
		fieldName = field.Name
		analysis.addGeneratedFile("src/validation/"+strings.ToLower(entity.Name)+".ts", "Field validation is generated from the source entity.")
	} else {
		analysis.addGeneratedFile("src/validation/"+strings.ToLower(entity.Name)+".ts", "Entity validation is generated from the source entity.")
	}

	analysis.addEntity(entity.Name, "The symbol is declared on this entity.")
	analysis.addGeneratedFile("src/types.ts", "Entity and relation types are generated from all entity declarations.")
	analysis.addGeneratedFile("prisma/schema.prisma", "Database schema is generated from entity fields and relations.")
	analysis.addGeneratedFile("src/setup-db.ts", "SQLite setup mirrors generated entity columns and relation IDs.")
	analysis.addGeneratedFile("openapi.json", "Generated REST schemas include entity fields and page actions.")

	if field != nil && validationUsesField(entity.Validations, fieldName) {
		analysis.addEntity(entity.Name, "Entity-level validation references "+fieldName+".")
	}

	for _, page := range program.Pages {
		pageEntity, ok := affectedFindEntity(program, page.Source)
		if !ok {
			continue
		}
		if page.Source == entity.Name {
			reason := "Page source is " + entity.Name + "."
			if field != nil && pageUsesField(page, fieldName) {
				reason = "Page table, form, search, filter, or sort references " + fieldName + "."
			}
			analysis.addPage(page.Name, reason)
			addGeneratedPageFiles(analysis, page, entity.Name, "Generated page/API route depends on "+entity.Name+".")
			continue
		}
		if pageHasRelationTo(pageEntity, entity.Name) {
			if field == nil || fieldName == relationLabelField(entity) {
				analysis.addPage(page.Name, "Page source has a relation to "+entity.Name+" and may display its relation label.")
				analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", "Relation select/display uses "+entity.Name+" records.")
			}
		}
	}

	for _, otherEntity := range program.Entities {
		if otherEntity.Name == entity.Name {
			continue
		}
		for _, otherField := range otherEntity.Fields {
			if otherField.Type == entity.Name {
				analysis.addEntity(otherEntity.Name, "Field "+otherField.Name+" references "+entity.Name+".")
				analysis.addGeneratedFile("prisma/schema.prisma", "Relation back-reference changes can affect both entities.")
			}
		}
	}

	for _, role := range program.Roles {
		for _, permission := range role.Permissions {
			if permission.Resource != entity.Name {
				continue
			}
			if field == nil || len(permission.Fields) == 0 || containsString(permission.Fields, fieldName) {
				analysis.addRole(role.Name, "Permission references "+entity.Name+affectedFieldSuffix(fieldName)+".")
			}
		}
	}

	for _, workflow := range program.Workflows {
		if workflow.Source == entity.Name && (field == nil || fieldName == "status") {
			analysis.addWorkflow(workflow.Name, "Workflow source is "+entity.Name+" and generated transitions mutate status.")
			for _, page := range program.Pages {
				if page.Source == entity.Name {
					addGeneratedPageFiles(analysis, page, entity.Name, "Workflow controls and routes are generated for "+workflow.Name+".")
				}
			}
		}
	}

	for _, state := range program.States {
		for _, stateField := range state.Fields {
			if stateField.Type == entity.Name {
				analysis.addState(state.Name, "State field "+stateField.Name+" uses "+entity.Name+".")
			}
		}
		if page, ok := pageForState(program, state); ok && page.Source == entity.Name {
			analysis.addState(state.Name, "State is bound to page "+page.Name+".")
		}
	}

	for _, component := range program.Components {
		for _, input := range component.Inputs {
			if field != nil && input.Name == field.Name && input.Type == field.Type {
				analysis.addComponent(component.Name, "Component input matches affected field "+field.Name+".")
				analysis.addGeneratedFile("src/components/"+component.Name+".tsx", "Generated component uses this input.")
			}
			if field == nil && input.Type == entity.Name {
				analysis.addComponent(component.Name, "Component input type references "+entity.Name+".")
				analysis.addGeneratedFile("src/components/"+component.Name+".tsx", "Generated component imports this entity type.")
			}
		}
	}

	addMatchingAPIs(analysis, program, entity.Name)
	if field != nil {
		addMatchingAPIs(analysis, program, entity.Name+"."+fieldName)
		addMatchingAPIs(analysis, program, fieldName)
	}
}

func populatePageAffected(analysis *AffectedAnalysis, program Program, page PageDecl) {
	analysis.addPage(page.Name, "The symbol is this page.")
	if entity, ok := affectedFindEntity(program, page.Source); ok {
		analysis.Entity = entity.Name
		analysis.addEntity(entity.Name, "Page source is "+entity.Name+".")
		addGeneratedPageFiles(analysis, page, entity.Name, "Generated page/API files are tied to this page.")
	}
	if page.Layout != "" {
		analysis.addGeneratedFile("src/App.tsx", "Application shell uses page layout and navigation.")
	}
	for _, roleName := range page.Access {
		analysis.addRole(roleName, "Page access references this role.")
	}
}

func populateComputedFieldAffected(analysis *AffectedAnalysis, program Program, entity EntityDecl, field ComputedFieldDecl) {
	analysis.addEntity(entity.Name, "The computed field is declared on this entity.")
	for _, page := range program.Pages {
		if page.Source != entity.Name {
			continue
		}
		analysis.addPage(page.Name, "Generated detail UI can display computed field "+field.Name+".")
		analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", "Generated React page computes "+entity.Name+"."+field.Name+" from source fields.")
	}
	for _, reference := range computedFieldReferences(field) {
		analysis.addEntity(entity.Name, "Computed field "+field.Name+" reads "+reference+".")
	}
}

func populateRoleAffected(analysis *AffectedAnalysis, program Program, role RoleDecl) {
	analysis.addRole(role.Name, "The symbol is this role.")
	analysis.addGeneratedFile("src/auth/UsersPage.tsx", "Generated role management uses declared roles.")
	analysis.addGeneratedFile("src/auth/AuditPage.tsx", "Generated audit UI is role-aware.")
	analysis.addGeneratedFile("src/routes/auth.ts", "Auth routes store and update user roles.")

	for _, permission := range role.Permissions {
		if permission.Resource == "" || permission.Action == "all" {
			for _, page := range program.Pages {
				analysis.addPage(page.Name, "Role has global permissions that can affect page actions.")
				addGeneratedPageFiles(analysis, page, page.Source, "Role guards affect generated page/API behavior.")
			}
			continue
		}
		analysis.addEntity(permission.Resource, "Role permission references this resource.")
		for _, page := range program.Pages {
			if page.Source == permission.Resource {
				analysis.addPage(page.Name, "Page source matches role permission resource.")
				addGeneratedPageFiles(analysis, page, page.Source, "Role permission guards affect generated page/API behavior.")
			}
		}
	}
	for _, page := range program.Pages {
		if containsString(page.Access, role.Name) {
			analysis.addPage(page.Name, "Page access list includes "+role.Name+".")
			addGeneratedPageFiles(analysis, page, page.Source, "Page-level role guard includes "+role.Name+".")
		}
	}
}

func populateWorkflowAffected(analysis *AffectedAnalysis, program Program, workflow WorkflowDecl) {
	analysis.addWorkflow(workflow.Name, "The symbol is this workflow.")
	analysis.Entity = workflow.Source
	analysis.addEntity(workflow.Source, "Workflow source is "+workflow.Source+".")
	for _, page := range program.Pages {
		if page.Source == workflow.Source {
			analysis.addPage(page.Name, "Generated page can render workflow transition controls.")
			addGeneratedPageFiles(analysis, page, workflow.Source, "Workflow routes, clients, and buttons are generated for this source.")
		}
	}
	for _, transition := range workflow.Transitions {
		for _, roleName := range transition.Allow {
			analysis.addRole(roleName, "Workflow transition "+transition.Name+" allows this role.")
		}
	}
	analysis.addGeneratedFile("openapi.json", "Workflow transition endpoints are added to the OpenAPI contract.")
}

func populateStateAffected(analysis *AffectedAnalysis, program Program, state StateDecl) {
	analysis.addState(state.Name, "The symbol is this state declaration.")
	if page, ok := pageForState(program, state); ok {
		analysis.addPage(page.Name, "State name binds to this generated page.")
		if entity, entityOK := affectedFindEntity(program, page.Source); entityOK {
			analysis.addEntity(entity.Name, "Bound page source is "+entity.Name+".")
		}
		analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", "Generated React state hooks are emitted in this page.")
	}
	for _, field := range state.Fields {
		if _, ok := affectedFindEntity(program, field.Type); ok {
			analysis.addEntity(field.Type, "State field "+field.Name+" uses this entity type.")
		}
	}
}

func populateComponentAffected(analysis *AffectedAnalysis, program Program, component ComponentDecl) {
	analysis.addComponent(component.Name, "The symbol is this component declaration.")
	analysis.addGeneratedFile("src/components/"+component.Name+".tsx", "Generated component source.")
	for _, input := range component.Inputs {
		for _, entity := range program.Entities {
			if input.Type == entity.Name {
				analysis.addEntity(entity.Name, "Component input type references this entity.")
			}
			for _, field := range entity.Fields {
				if input.Name == field.Name && input.Type == field.Type {
					analysis.addEntity(entity.Name, "Component input can bind to "+entity.Name+"."+field.Name+".")
					for _, page := range program.Pages {
						if page.Source == entity.Name && pageUsesField(page, field.Name) {
							analysis.addPage(page.Name, "Generated page may render "+field.Name+" through "+component.Name+".")
							analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", "Page imports matching generated component.")
						}
					}
				}
			}
		}
	}
}

func populateAPIAffected(analysis *AffectedAnalysis, api APIDecl) {
	analysis.addAPI(api.Name, "The symbol is this explicit API contract.")
	analysis.addGeneratedFile("openapi.json", "Explicit API declarations are written to the OpenAPI contract.")
}

func populateAuthAffected(analysis *AffectedAnalysis, program Program) {
	analysis.addGeneratedFile("src/auth/AuthPage.tsx", "Auth declaration generates login/register UI.")
	analysis.addGeneratedFile("src/routes/auth.ts", "Auth declaration generates auth API routes.")
	analysis.addGeneratedFile("src/App.tsx", "Generated app restores auth state and renders protected shell.")
	analysis.addGeneratedFile("src/server.ts", "Generated server mounts auth routes and protects CRUD routes.")
	for _, page := range program.Pages {
		analysis.addPage(page.Name, "Auth affects generated access checks for pages.")
	}
}

func populateDatabaseAffected(analysis *AffectedAnalysis) {
	analysis.addGeneratedFile(".env.example", "Database env reference is documented for local setup.")
	analysis.addGeneratedFile("prisma/schema.prisma", "Database provider and schema are emitted for generated persistence.")
	analysis.addGeneratedFile("src/db.ts", "Generated Prisma client reads database configuration.")
	analysis.addGeneratedFile("src/setup-db.ts", "Generated SQLite setup writes the local MVP database.")
}

func populateSecurityAffected(analysis *AffectedAnalysis, program Program) {
	analysis.addGeneratedFile(".env.example", "Security env references are documented for local setup.")
	analysis.addGeneratedFile("src/server.ts", "Generated server emits security middleware from security intent.")
	if program.Deploy != nil && program.Deploy.Target == "docker" {
		analysis.addGeneratedFile("docker-compose.yml", "Generated compose environment mirrors security env references.")
	}
}

func populateDeployAffected(analysis *AffectedAnalysis) {
	analysis.addGeneratedFile(".env.example", "Deployment env defaults are documented for local setup.")
	analysis.addGeneratedFile(".dockerignore", "Docker build context exclusions are generated from deploy intent.")
	analysis.addGeneratedFile("Dockerfile", "Docker image build instructions are generated from deploy intent.")
	analysis.addGeneratedFile("docker-compose.yml", "Docker Compose service wiring is generated from deploy intent.")
	analysis.addGeneratedFile("package.json", "Generated package scripts include the production start command.")
	analysis.addGeneratedFile("src/server.ts", "Generated server reads the configured deploy port environment variable.")
}

func populateTargetAffected(analysis *AffectedAnalysis, program Program) {
	analysis.addGeneratedFile("README.md", "Generated summary documents the selected target stack.")
	analysis.addGeneratedFile("package.json", "Generated package dependencies and scripts depend on the target stack.")
	analysis.addGeneratedFile("vite.config.ts", "Frontend target selection affects Vite configuration.")
	analysis.addGeneratedFile("prisma.config.ts", "Database target selection affects Prisma configuration.")
	analysis.addGeneratedFile("prisma/schema.prisma", "Database target selection affects generated schema provider.")
	analysis.addGeneratedFile("src/server.ts", "Backend target selection affects generated server runtime.")
	if program.Deploy != nil && program.Deploy.Target == "docker" {
		analysis.addGeneratedFile("Dockerfile", "Docker build instructions depend on the selected target stack.")
		analysis.addGeneratedFile("docker-compose.yml", "Docker Compose service wiring depends on target stack runtime.")
	}
}

func populateAppAffected(analysis *AffectedAnalysis, program Program) {
	if program.App.Name != "" {
		analysis.addGeneratedFile("package.json", "Generated package metadata uses the app name.")
		analysis.addGeneratedFile("index.html", "Generated HTML title uses the app name.")
		analysis.addGeneratedFile("src/App.tsx", "Generated shell displays the app name.")
	}
}

func addGeneratedPageFiles(analysis *AffectedAnalysis, page PageDecl, entityName string, reason string) {
	analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", reason)
	if entityName != "" {
		fileName := strings.ToLower(entityName)
		analysis.addGeneratedFile("src/api/"+fileName+".ts", reason)
		analysis.addGeneratedFile("src/routes/"+fileName+".ts", reason)
		analysis.addGeneratedFile("src/validation/"+fileName+".ts", reason)
	}
	analysis.addGeneratedFile("src/App.tsx", "Generated navigation and route registration include pages.")
	analysis.addGeneratedFile("src/server.ts", "Generated server mounts API routes for pages.")
	analysis.addGeneratedFile("openapi.json", "Generated page actions are reflected in the OpenAPI contract.")
}

func addMatchingAPIs(analysis *AffectedAnalysis, program Program, needle string) {
	if needle == "" {
		return
	}
	lowerNeedle := strings.ToLower(needle)
	for _, api := range program.APIs {
		if strings.Contains(strings.ToLower(api.Name), lowerNeedle) || strings.Contains(strings.ToLower(api.Path), lowerNeedle) {
			analysis.addAPI(api.Name, "API name or path mentions "+needle+".")
			analysis.addGeneratedFile("openapi.json", "Matching explicit API contract is generated into OpenAPI.")
		}
	}
}

func missingTopLevelDiagnostic(analysis AffectedAnalysis, symbol string) []Diagnostic {
	if analysis.Found {
		return nil
	}
	return []Diagnostic{{
		Code:       "UNKNOWN_AFFECTED_SYMBOL",
		Message:    fmt.Sprintf("Top-level symbol %q is not declared in this project.", symbol),
		Suggestion: "Run `black inspect --json` to see declared top-level blocks.",
	}}
}

func affectedFindEntity(program Program, name string) (EntityDecl, bool) {
	for _, entity := range program.Entities {
		if entity.Name == name {
			return entity, true
		}
	}
	return EntityDecl{}, false
}

func affectedFindPage(program Program, name string) (PageDecl, bool) {
	for _, page := range program.Pages {
		if page.Name == name {
			return page, true
		}
	}
	return PageDecl{}, false
}

func affectedFindRole(program Program, name string) (RoleDecl, bool) {
	for _, role := range program.Roles {
		if role.Name == name {
			return role, true
		}
	}
	return RoleDecl{}, false
}

func affectedFindWorkflow(program Program, name string) (WorkflowDecl, bool) {
	for _, workflow := range program.Workflows {
		if workflow.Name == name {
			return workflow, true
		}
	}
	return WorkflowDecl{}, false
}

func affectedFindState(program Program, name string) (StateDecl, bool) {
	for _, state := range program.States {
		if state.Name == name {
			return state, true
		}
	}
	return StateDecl{}, false
}

func affectedFindComponent(program Program, name string) (ComponentDecl, bool) {
	for _, component := range program.Components {
		if component.Name == name {
			return component, true
		}
	}
	return ComponentDecl{}, false
}

func affectedFindAPI(program Program, name string) (APIDecl, bool) {
	for _, api := range program.APIs {
		if api.Name == name {
			return api, true
		}
	}
	return APIDecl{}, false
}

func pageForState(program Program, state StateDecl) (PageDecl, bool) {
	for _, page := range program.Pages {
		if state.Name == page.Name+"State" || state.Name == page.Name+"PageState" {
			return page, true
		}
	}
	return PageDecl{}, false
}

func pageUsesField(page PageDecl, fieldName string) bool {
	return containsString(page.Table.Columns, fieldName) ||
		containsString(page.Table.Search, fieldName) ||
		containsString(page.Table.Filters, fieldName) ||
		page.Table.Sort.Field == fieldName ||
		containsString(page.Form.Fields, fieldName)
}

func pageHasRelationTo(entity EntityDecl, targetEntity string) bool {
	for _, field := range entity.Fields {
		if field.Type == targetEntity {
			return true
		}
	}
	return false
}

func validationUsesField(validations []EntityValidationDecl, fieldName string) bool {
	for _, validation := range validations {
		if validation.Left == fieldName || validation.Right == fieldName {
			return true
		}
		if validation.When != nil && (validation.When.Left == fieldName || validation.When.Right == fieldName) {
			return true
		}
	}
	return false
}

func affectedFieldSuffix(fieldName string) string {
	if fieldName == "" {
		return ""
	}
	return "." + fieldName
}

func (analysis *AffectedAnalysis) addEntity(name string, reason string) {
	analysis.Entities = addAffectedItem(analysis.Entities, name, reason)
}

func (analysis *AffectedAnalysis) addPage(name string, reason string) {
	analysis.Pages = addAffectedItem(analysis.Pages, name, reason)
}

func (analysis *AffectedAnalysis) addRole(name string, reason string) {
	analysis.Roles = addAffectedItem(analysis.Roles, name, reason)
}

func (analysis *AffectedAnalysis) addWorkflow(name string, reason string) {
	analysis.Workflows = addAffectedItem(analysis.Workflows, name, reason)
}

func (analysis *AffectedAnalysis) addState(name string, reason string) {
	analysis.States = addAffectedItem(analysis.States, name, reason)
}

func (analysis *AffectedAnalysis) addComponent(name string, reason string) {
	analysis.Components = addAffectedItem(analysis.Components, name, reason)
}

func (analysis *AffectedAnalysis) addAPI(name string, reason string) {
	analysis.APIs = addAffectedItem(analysis.APIs, name, reason)
}

func (analysis *AffectedAnalysis) addGeneratedFile(name string, reason string) {
	analysis.GeneratedFiles = addAffectedItem(analysis.GeneratedFiles, name, reason)
}

func addAffectedItem(items []AffectedItem, name string, reason string) []AffectedItem {
	if name == "" {
		return items
	}
	for index, item := range items {
		if item.Name != name {
			continue
		}
		if reason != "" && item.Reason != reason && !strings.Contains(item.Reason, reason) {
			items[index].Reason = item.Reason + " " + reason
		}
		return items
	}
	return append(items, AffectedItem{Name: name, Reason: reason})
}
