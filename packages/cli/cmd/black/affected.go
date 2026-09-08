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
			Suggestion: "Use an entity, field, page, view section, query, job, action, transaction, service, seed, test, migration, role, workflow, state, component, api, target, deploy, ops, i18n, or app symbol such as `Product`, `Product.stock`, `Products`, `Products.StockSummary`, `LowStockMonitor`, `RestockAtomic`, `InventoryIntegration`, `DemoProducts`, `WarehouseBrowserSmoke`, `migration`, `ops`, or `target`.",
		}}
	}

	if strings.Contains(symbol, ".") {
		entityName, fieldName, ok := strings.Cut(symbol, ".")
		entity, entityOK := affectedFindEntity(program, entityName)
		if ok && entityOK {
			if fieldName == "index" || fieldName == "indexes" {
				analysis.Kind = "entity-index"
				analysis.Found = len(entity.Indexes) > 0
				analysis.Entity = entity.Name
				populateEntityIndexAffected(&analysis, entity)
				return analysis, missingEntityIndexDiagnostic(analysis)
			}
			if field, fieldOK := findField(entity, fieldName); fieldOK {
				analysis.Kind = "field"
				if _, ok := affectedFindEntity(program, field.Type); ok {
					analysis.Kind = "relation-field"
				}
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
		if ok {
			if page, section, sectionOK := affectedFindPageViewSection(program, entityName, fieldName); sectionOK {
				analysis.Kind = "view-section"
				analysis.Found = true
				populateViewSectionAffected(&analysis, program, page, section)
				return analysis, nil
			}
		}
		return analysis, []Diagnostic{{
			Code:       "UNKNOWN_AFFECTED_SYMBOL",
			Message:    fmt.Sprintf("Affected symbol %q was not found.", symbol),
			Suggestion: "Use an existing entity field, computed-field, index, or page view section symbol such as `Product.stock`, `Product.inventoryValue`, `Product.index`, or `Products.StockSummary`.",
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
	if query, ok := findQuery(program, symbol); ok {
		analysis.Kind = "query"
		analysis.Found = true
		analysis.Entity = query.Source
		populateQueryAffected(&analysis, program, query)
		return analysis, nil
	}
	if job, ok := findJob(program, symbol); ok {
		analysis.Kind = "job"
		analysis.Found = true
		populateJobAffected(&analysis, program, job)
		return analysis, nil
	}
	if action, ok := findCustomAction(program, symbol); ok {
		analysis.Kind = "action"
		analysis.Found = true
		analysis.Entity = action.Source
		populateActionAffected(&analysis, program, action)
		return analysis, nil
	}
	if transaction, ok := findTransaction(program, symbol); ok {
		analysis.Kind = "transaction"
		analysis.Found = true
		populateTransactionAffected(&analysis, program, transaction)
		return analysis, nil
	}
	if service, ok := findService(program, symbol); ok {
		analysis.Kind = "service"
		analysis.Found = true
		populateServiceAffected(&analysis, program, service)
		return analysis, nil
	}
	if seed, ok := findSeed(program, symbol); ok {
		analysis.Kind = "seed"
		analysis.Found = true
		analysis.Entity = seed.Source
		populateSeedAffected(&analysis, program, seed)
		return analysis, nil
	}
	if test, ok := findTest(program, symbol); ok {
		analysis.Kind = "test"
		analysis.Found = true
		populateTestAffected(&analysis, program, test)
		return analysis, nil
	}
	if migration, ok := affectedFindMigration(program, symbol); ok {
		analysis.Kind = "migration"
		analysis.Found = true
		populateMigrationAffected(&analysis, migration)
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
	if page, section, ok := affectedFindViewSection(program, symbol); ok {
		analysis.Kind = "view-section"
		analysis.Found = true
		populateViewSectionAffected(&analysis, program, page, section)
		return analysis, nil
	}
	if api, ok := affectedFindAPI(program, symbol); ok {
		analysis.Kind = "api"
		analysis.Found = true
		populateAPIAffected(&analysis, program, api)
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
		populateDatabaseAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "database")
	}
	if symbol == "seed" || symbol == "seeds" || symbol == "fixture" || symbol == "fixtures" {
		analysis.Kind = "seed"
		analysis.Found = len(program.Seeds) > 0
		populateSeedsAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "seed")
	}
	if symbol == "test" || symbol == "tests" || symbol == "browser" || symbol == "browser-test" || symbol == "browser-tests" {
		analysis.Kind = "test"
		analysis.Found = len(program.Tests) > 0
		populateTestsAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "test")
	}
	if symbol == "job" || symbol == "jobs" || symbol == "worker" || symbol == "workers" || symbol == "background" || symbol == "background-jobs" {
		analysis.Kind = "job"
		analysis.Found = len(program.Jobs) > 0
		populateJobsAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "job")
	}
	if symbol == "migration" || symbol == "migrations" {
		analysis.Kind = "migration"
		analysis.Found = len(program.Migrations) > 0
		populateMigrationsAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "migration")
	}
	if symbol == "security" || symbol == "cors" {
		analysis.Kind = "security"
		analysis.Found = program.Security != nil
		populateSecurityAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "security")
	}
	if symbol == "policy" || symbol == "policies" || symbol == "owner" || symbol == "tenant" {
		analysis.Kind = "policy"
		analysis.Found = programHasEntityPolicies(program, symbol)
		populatePolicyAffected(&analysis, program, symbol)
		return analysis, missingPolicyDiagnostic(analysis, symbol)
	}
	if symbol == "deploy" || symbol == "cloud" {
		analysis.Kind = "deploy"
		analysis.Found = program.Deploy != nil && (symbol == "deploy" || program.Deploy.Cloud != nil)
		populateDeployAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "deploy")
	}
	if symbol == "ops" || symbol == "health" || symbol == "readiness" || symbol == "metrics" || symbol == "logging" || symbol == "observe" || symbol == "observability" {
		analysis.Kind = "ops"
		analysis.Found = programHasOpsSignal(program, symbol)
		populateOpsAffected(&analysis, program)
		return analysis, missingOpsDiagnostic(analysis, symbol)
	}
	if symbol == "i18n" {
		analysis.Kind = "i18n"
		analysis.Found = program.I18N != nil
		populateI18NAffected(&analysis, program)
		return analysis, missingTopLevelDiagnostic(analysis, "i18n")
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
		Suggestion: "Use an existing entity, field, page, view section, query, job, action, transaction, service, seed, test, migration, role, workflow, state, component, api, auth, database, security, deploy, ops, i18n, target, or app symbol.",
	}}
}

func populateEntityIndexAffected(analysis *AffectedAnalysis, entity EntityDecl) {
	analysis.addEntity(entity.Name, "Entity index declarations are defined on this entity.")
	analysis.addGeneratedFile("prisma/schema.prisma", "Prisma indexes are generated from entity index declarations.")
	analysis.addGeneratedFile("src/setup-db.ts", "SQLite setup creates deterministic indexes from entity index declarations.")
	analysis.AgentNotes = append(analysis.AgentNotes, "Index changes affect generated database schema/setup; they do not change API shape, React pages, or stored field names.")
}

func newAffectedAnalysis(symbol string) AffectedAnalysis {
	return AffectedAnalysis{
		Symbol:         symbol,
		Kind:           "unknown",
		Entities:       []AffectedItem{},
		Migrations:     []AffectedItem{},
		Pages:          []AffectedItem{},
		Queries:        []AffectedItem{},
		Jobs:           []AffectedItem{},
		Actions:        []AffectedItem{},
		Transactions:   []AffectedItem{},
		Services:       []AffectedItem{},
		Seeds:          []AffectedItem{},
		Tests:          []AffectedItem{},
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
		if _, ok := affectedFindEntity(program, field.Type); ok {
			analysis.AgentNotes = append(analysis.AgentNotes, "Relation field changes can affect generated Prisma relation IDs, API response include policy, relation label display fallback, OpenAPI relation-load metadata, and page relation select options.")
		}
	} else {
		analysis.addGeneratedFile("src/validation/"+strings.ToLower(entity.Name)+".ts", "Entity validation is generated from the source entity.")
	}

	analysis.addEntity(entity.Name, "The symbol is declared on this entity.")
	analysis.addGeneratedFile("src/types.ts", "Entity and relation types are generated from all entity declarations.")
	analysis.addGeneratedFile("prisma/schema.prisma", "Database schema is generated from entity fields and relations.")
	analysis.addGeneratedFile("src/setup-db.ts", "SQLite setup mirrors generated entity columns and relation IDs.")
	analysis.addGeneratedFile("openapi.json", "Generated REST schemas include entity fields, page actions, and relation load metadata.")
	if len(entity.Policies) > 0 && (field == nil || entityPolicyFieldMap(entity)[fieldName].Field != "") {
		analysis.addGeneratedFile("src/routes/"+strings.ToLower(entity.Name)+".ts", "Generated API routes apply owner/tenant row policy filters.")
		if program.Auth != nil {
			analysis.addGeneratedFile("src/routes/auth.ts", "Generated auth runtime provides user and tenant identity for row policies.")
		}
		analysis.AgentNotes = append(analysis.AgentNotes, "Entity owner/tenant policy fields are stamped by generated routes; do not put them in forms or custom action set clauses.")
	}
	for _, query := range program.Queries {
		if query.Source != entity.Name {
			continue
		}
		reason := "Query returns stored records from " + entity.Name + "."
		if field != nil && containsString(queryFieldNames(query), fieldName) {
			reason = "Query filters, sorts, or aggregates by " + entity.Name + "." + fieldName + "."
		}
		analysis.addQuery(query.Name, reason)
		addMatchingJobsForQuery(analysis, program, query.Name, "Job runs query "+query.Name+" and depends on its entity fields.")
	}
	for _, action := range program.Actions {
		if action.Source != entity.Name {
			continue
		}
		reason := "Action mutates stored records from " + entity.Name + "."
		if field != nil && customActionUsesField(action, fieldName) {
			reason = "Action reads or writes " + entity.Name + "." + fieldName + "."
		}
		if field == nil || customActionUsesField(action, fieldName) {
			analysis.addAction(action.Name, reason)
		}
	}
	for _, seed := range program.Seeds {
		if seed.Source != entity.Name {
			continue
		}
		reason := "Seed writes deterministic fixture rows for " + entity.Name + "."
		if field != nil {
			if !seedUsesField(seed, fieldName) {
				continue
			}
			reason = "Seed sets " + entity.Name + "." + fieldName + " fixture values."
		}
		analysis.addSeed(seed.Name, reason)
		analysis.addGeneratedFile("src/seed.ts", "Generated seed runtime applies declared fixture rows.")
		analysis.addGeneratedFile("package.json", "Generated db:setup/db:seed scripts run seed data when declared.")
	}

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
			if field != nil && pageUsesField(program, page, fieldName) {
				reason = "Page table, form, search, filter, sort, or component section references " + fieldName + "."
				if _, ok := affectedFindEntity(program, field.Type); ok {
					reason = "Page table, form, search, filter, sort, component section, or generated relation load policy references " + fieldName + "."
				}
			}
			analysis.addPage(page.Name, reason)
			addGeneratedPageFiles(analysis, program, page, entity.Name, "Generated page/API route depends on "+entity.Name+".")
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
					addGeneratedPageFiles(analysis, program, page, entity.Name, "Workflow controls and routes are generated for "+workflow.Name+".")
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

	for _, migration := range program.Migrations {
		for _, rename := range migration.Renames {
			if rename.Kind == "entity" && (rename.To == entity.Name || rename.From == entity.Name) {
				analysis.addMigration(migration.Name, "Migration renames entity "+rename.From+" to "+rename.To+".")
				addMigrationGeneratedFiles(analysis)
			}
			if rename.Kind == "field" && rename.Entity == entity.Name {
				if field == nil || rename.From == fieldName || rename.To == fieldName {
					analysis.addMigration(migration.Name, "Migration renames field "+rename.Entity+"."+rename.From+" to "+rename.To+".")
					addMigrationGeneratedFiles(analysis)
				}
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
	if page.Query != "" {
		analysis.addQuery(page.Query, "Page list is bound to this query.")
	}
	for _, action := range customActionsForPage(program, page) {
		analysis.addAction(action.Name, "Page exposes this custom action.")
	}
	for _, section := range pageViewComponentSections(page) {
		analysis.addComponent(section.Component, "Page view section "+section.Name+" renders this component.")
		analysis.addGeneratedFile("src/components/"+section.Component+".tsx", "Generated page imports this component for a component section.")
	}
	if entity, ok := affectedFindEntity(program, page.Source); ok {
		analysis.Entity = entity.Name
		analysis.addEntity(entity.Name, "Page source is "+entity.Name+".")
		addGeneratedPageFiles(analysis, program, page, entity.Name, "Generated page/API files are tied to this page.")
	}
	if page.Layout != "" {
		analysis.addGeneratedFile("src/App.tsx", "Application shell uses page layout and navigation.")
	}
	for _, roleName := range page.Access {
		analysis.addRole(roleName, "Page access references this role.")
	}
	for _, test := range program.Tests {
		if test.Page == page.Name {
			analysis.addTest(test.Name, "Browser test targets this page.")
			analysis.addGeneratedFile("src/blacklang.browser.test.tsx", "Generated browser checks assert page expectations.")
			analysis.addGeneratedFile("src/blacklang.e2e.test.ts", "Generated browser e2e checks execute declared page expectations in a real browser.")
			analysis.addGeneratedFile("src/blacklang.e2e.matrix.ts", "Generated browser e2e matrix runner executes declared expectations across available browser targets.")
			analysis.addGeneratedFile("tests/browser-matrix.json", "Generated browser matrix manifest exposes deterministic target and command metadata.")
			analysis.addGeneratedFile("package.json", "Generated package test scripts include browser checks and browser e2e when tests are declared.")
		}
	}
	// Adding/removing a query binding can change which page owns the canonical
	// entity client used by sibling pages and relation selectors.
	for _, sibling := range program.Pages {
		if sibling.Name == page.Name || sibling.Source != page.Source {
			continue
		}
		analysis.addPage(sibling.Name, "Page query binding can change canonical entity API module selection.")
		addGeneratedPageFiles(analysis, program, sibling, sibling.Source, "Sibling page shares canonical entity client selection.")
		if sibling.Query != "" {
			analysis.addQuery(sibling.Query, "Sibling query uses the same entity API module selection.")
		}
	}
	for _, consumer := range program.Pages {
		if entity, ok := affectedFindEntity(program, consumer.Source); ok && pageHasRelationTo(entity, page.Source) {
			analysis.addPage(consumer.Name, "Relation selectors use the canonical page for "+page.Source+".")
			analysis.addGeneratedFile("src/pages/"+consumer.Name+"Page.tsx", "Relation client/navigation depends on canonical page selection.")
		}
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

func populateQueryAffected(analysis *AffectedAnalysis, program Program, query QueryDecl) {
	analysis.addQuery(query.Name, "The symbol is this query declaration.")
	analysis.addEntity(query.Source, "Query reads stored records from this entity.")
	addMatchingJobsForQuery(analysis, program, query.Name, "Job runs this query from the generated worker.")
	for _, page := range program.Pages {
		if page.Query != query.Name {
			continue
		}
		analysis.addPage(page.Name, "Page list uses query "+query.Name+".")
		module := (&webGenerator{program: program}).pageModuleName(page)
		analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", "Page loads and refreshes this query result and any declared aggregate summary.")
		analysis.addGeneratedFile("src/api/"+module+".ts", "Page API client exposes the bound query list and aggregate summary when declared.")
		analysis.addGeneratedFile("src/routes/"+module+".ts", "Query filtering, ordering, limit, aggregate summary, and read guards execute here.")
		analysis.addGeneratedFile("openapi.json", "Bound query endpoints describe this query and any declared aggregate summary.")
		for _, role := range page.Access {
			analysis.addRole(role, "Bound query endpoint uses this page's access policy.")
		}
	}
	for _, role := range program.Roles {
		for _, permission := range role.Permissions {
			if permission.Action == "all" || (permission.Resource == query.Source && (permission.Action == "read" || permission.Action == "manage")) {
				analysis.addRole(role.Name, "Query checks entity and filter/sort field read permissions.")
			}
		}
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "Query changes affect bound lists and declared summary endpoints; they do not change database columns or mutation rules. Entity policies provide row authorization separately. Unbound queries expose no endpoint.")
}

func populateJobAffected(analysis *AffectedAnalysis, program Program, job JobDecl) {
	analysis.addJob(job.Name, "The symbol is this background job declaration.")
	analysis.addGeneratedFile("jobs/manifest.json", "Generated job manifest lists schedule and read-only run metadata.")
	analysis.addGeneratedFile("src/worker.ts", "Generated worker runs declared jobs.")
	analysis.addGeneratedFile("package.json", "Generated jobs:run and jobs:loop scripts are available when jobs are declared.")
	analysis.addGeneratedFile("openapi.json", "OpenAPI root metadata includes declared jobs for agent discovery.")
	if job.Run.Kind == "query" {
		analysis.addQuery(job.Run.Query, "Job runs this query.")
		if query, ok := findQuery(program, job.Run.Query); ok {
			analysis.addEntity(query.Source, "Job query reads this entity in internal worker scope.")
		}
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "Jobs are generated internal workers in this MVP. They run declared queries on a schedule and log compact metadata; queues, retries, external schedulers, and mutating jobs are future adapter work.")
}

func populateJobsAffected(analysis *AffectedAnalysis, program Program) {
	if len(program.Jobs) > 0 {
		analysis.addGeneratedFile("jobs/manifest.json", "Generated job manifest lists every declared job.")
		analysis.addGeneratedFile("src/worker.ts", "Generated worker runs declared jobs.")
		analysis.addGeneratedFile("package.json", "Generated jobs scripts are emitted when jobs exist.")
		analysis.addGeneratedFile("openapi.json", "OpenAPI root metadata includes declared jobs for agent discovery.")
	}
	for _, job := range program.Jobs {
		analysis.addJob(job.Name, "Declared background job.")
		if job.Run.Kind == "query" {
			analysis.addQuery(job.Run.Query, "Job runs this query.")
		}
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "Inspect a specific job name for query/source impact before changing generated worker behavior.")
}

func populateActionAffected(analysis *AffectedAnalysis, program Program, action CustomActionDecl) {
	analysis.addAction(action.Name, "The symbol is this custom action declaration.")
	analysis.addEntity(action.Source, "Action source is "+action.Source+".")
	analysis.addGeneratedFile("src/types.ts", "Action input types are generated from custom action inputs.")
	analysis.addGeneratedFile("src/validation/"+strings.ToLower(action.Source)+".ts", "Action input validation is generated with the entity validation module.")
	analysis.addGeneratedFile("openapi.json", "Bound custom action endpoints describe request and response schemas.")
	for _, page := range program.Pages {
		if page.Source != action.Source {
			continue
		}
		if containsString(page.Actions, action.Name) {
			analysis.addPage(page.Name, "Page actions list exposes "+action.Name+".")
			addGeneratedPageFiles(analysis, program, page, action.Source, "Custom action route, client method, and UI control are generated here.")
		}
	}
	for _, roleName := range action.Allow {
		analysis.addRole(roleName, "Custom action allow list references this role.")
	}
	if transactionName, ok := transactionActionBindings(program)[action.Name]; ok {
		analysis.addTransaction(transactionName, "Transaction block wraps this action route in a generated Prisma transaction.")
	}
	for _, role := range program.Roles {
		for _, permission := range role.Permissions {
			if permission.Action == "all" || (permission.Resource == action.Source && (permission.Action == "update" || permission.Action == "manage")) {
				analysis.addRole(role.Name, "Custom action checks update permission and writable fields.")
			}
		}
	}
	note := "Custom actions are row-level mutations. They use page/auth/update permission checks and do not create standalone endpoints unless a page lists the action."
	if _, ok := transactionActionBindings(program)[action.Name]; ok {
		note += " A top-level transaction block wraps this generated action route, so lookup, update, and audit logging run in one Prisma transaction."
	}
	analysis.AgentNotes = append(analysis.AgentNotes, note)
}

func populateTransactionAffected(analysis *AffectedAnalysis, program Program, transaction TransactionDecl) {
	analysis.addTransaction(transaction.Name, "The symbol is this top-level transaction declaration.")
	analysis.addGeneratedFile("src/server.ts", "Generated API server wraps targeted explicit API update handlers in Prisma transactions.")
	analysis.addGeneratedFile("openapi.json", "Transaction metadata is exposed on targeted explicit API and custom action operations.")
	analysis.addGeneratedFile("src/blacklang.contract.test.ts", "Generated contract tests assert transaction metadata for targeted operations.")
	for _, target := range transaction.Actions {
		if action, ok := findCustomAction(program, target.Name); ok {
			analysis.addAction(action.Name, "Transaction targets this custom action.")
			analysis.addEntity(action.Source, "Targeted action mutates this entity.")
			for _, page := range program.Pages {
				if page.Source == action.Source && containsString(page.Actions, action.Name) {
					analysis.addPage(page.Name, "This page exposes the targeted custom action route.")
					addGeneratedPageFiles(analysis, program, page, action.Source, "Custom action route selection changes transactional runtime behavior.")
				}
			}
		} else {
			analysis.addAction(target.Name, "Transaction references this action name.")
		}
	}
	for _, target := range transaction.APIs {
		if api, ok := affectedFindAPI(program, target.Name); ok {
			analysis.addAPI(api.Name, "Transaction targets this explicit API update handler.")
			if api.Update != nil {
				analysis.addEntity(api.Update.Source, "Targeted API update handler mutates this entity.")
			}
		} else {
			analysis.addAPI(target.Name, "Transaction references this api name.")
		}
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "Transaction blocks are top-level atomic runtime intent. They do not create new endpoints; they change generated runtime boundaries and OpenAPI metadata for targeted action/API update handlers.")
}

func populateServiceAffected(analysis *AffectedAnalysis, program Program, service ServiceDecl) {
	analysis.addService(service.Name, "The symbol is this top-level service declaration.")
	analysis.addGeneratedFile("services/manifest.json", "Generated service manifest groups explicit API declarations for agent discovery.")
	analysis.addGeneratedFile("src/services/"+kebabCase(service.Name)+".ts", "Generated service module exports deterministic explicit API metadata.")
	analysis.addGeneratedFile("openapi.json", "OpenAPI tags and x-blacklang-service metadata are generated for service-bound explicit APIs.")
	analysis.addGeneratedFile("src/blacklang.contract.test.ts", "Generated contract tests assert service metadata for targeted API operations.")
	for _, target := range service.APIs {
		if api, ok := affectedFindAPI(program, target.Name); ok {
			analysis.addAPI(api.Name, "Service exposes this explicit API declaration.")
			if api.Update != nil {
				analysis.addEntity(api.Update.Source, "Service API update handler reads and mutates this entity.")
			}
			if transactionName, ok := transactionAPIBindings(program)[api.Name]; ok {
				analysis.addTransaction(transactionName, "Service API is also wrapped by this transaction block.")
			}
		}
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "Service blocks are top-level API module intent. They do not create routes; they group existing explicit APIs for generated service manifests, service modules, OpenAPI tags, and compact agent discovery.")
}

func addMatchingJobsForQuery(analysis *AffectedAnalysis, program Program, queryName string, reason string) {
	for _, job := range program.Jobs {
		if job.Run.Kind == "query" && job.Run.Query == queryName {
			analysis.addJob(job.Name, reason)
			analysis.addGeneratedFile("jobs/manifest.json", "Generated job manifest references "+job.Name+".")
			analysis.addGeneratedFile("src/worker.ts", "Generated worker executes "+job.Name+".")
			analysis.addGeneratedFile("package.json", "Generated job scripts run declared jobs.")
		}
	}
}

func populateSeedAffected(analysis *AffectedAnalysis, program Program, seed SeedDecl) {
	analysis.addSeed(seed.Name, "The symbol is this seed declaration.")
	analysis.addEntity(seed.Source, "Seed source is "+seed.Source+".")
	analysis.addGeneratedFile("src/seed.ts", "Generated seed runtime applies "+seed.Name+" rows with deterministic upserts.")
	analysis.addGeneratedFile("src/setup-db.ts", "Generated database setup runs seed data after schema setup when seeds are declared.")
	analysis.addGeneratedFile("package.json", "Generated package scripts include db:seed and wire db:setup to seed data.")
	if entity, ok := affectedFindEntity(program, seed.Source); ok {
		for _, row := range seed.Rows {
			for _, value := range row.Values {
				if field, fieldOK := findField(entity, value.Field); fieldOK && !supportedFieldTypes[field.Type] {
					analysis.addEntity(field.Type, "Seed row "+row.Key+" references related row "+value.Value.Value+".")
				}
			}
		}
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "Seed changes affect local/demo database content through generated setup; they do not change schema, API routes, or React behavior.")
}

func populateMigrationAffected(analysis *AffectedAnalysis, migration MigrationDecl) {
	analysis.addMigration(migration.Name, "The symbol is this migration declaration.")
	for _, rename := range migration.Renames {
		switch rename.Kind {
		case "entity":
			analysis.addEntity(rename.To, "Migration renames previous entity "+rename.From+" to current entity "+rename.To+".")
		case "field":
			analysis.addEntity(rename.Entity, "Migration renames previous field "+rename.From+" to current field "+rename.To+".")
		}
	}
	addMigrationGeneratedFiles(analysis)
	analysis.AgentNotes = append(analysis.AgentNotes, "Validate migration declarations before building; generated db:migrate:plan is read-only, db:migrate applies declared renames, and generated setup applies them before schema setup.")
}

func populateMigrationsAffected(analysis *AffectedAnalysis, program Program) {
	for _, migration := range program.Migrations {
		analysis.addMigration(migration.Name, "Migration declaration is part of the generated migration manifest.")
	}
	addMigrationGeneratedFiles(analysis)
	analysis.AgentNotes = append(analysis.AgentNotes, "Migration changes affect generated setup, generated migration preview files, and generated online migration runner scripts.")
}

func populateSeedsAffected(analysis *AffectedAnalysis, program Program) {
	for _, seed := range program.Seeds {
		analysis.addSeed(seed.Name, "Seed declaration writes deterministic fixture rows.")
		analysis.addEntity(seed.Source, "Seed source is "+seed.Source+".")
	}
	analysis.addGeneratedFile("src/seed.ts", "Generated seed runtime applies declared fixture rows.")
	analysis.addGeneratedFile("src/setup-db.ts", "Generated database setup runs seed data after schema setup when seeds are declared.")
	analysis.addGeneratedFile("package.json", "Generated package scripts include db:seed and wire db:setup to seed data.")
	analysis.AgentNotes = append(analysis.AgentNotes, "Seed rows are deterministic local/demo data. Keep secrets out of seed literals and rerun db:setup to apply changes.")
}

func populateTestAffected(analysis *AffectedAnalysis, program Program, test TestDecl) {
	analysis.addTest(test.Name, "The symbol is this test declaration.")
	if page, ok := affectedFindPage(program, test.Page); ok {
		analysis.Entity = page.Source
		analysis.addPage(page.Name, "Test targets this generated page.")
		analysis.addEntity(page.Source, "Test expectations read generated UI intent for this page source.")
		for _, expectation := range test.Expectations {
			switch expectation.Kind {
			case "action":
				analysis.addPage(page.Name, "Test expects action "+expectation.Value+" on this page.")
				if action, ok := findCustomAction(program, expectation.Value); ok {
					analysis.addAction(action.Name, "Test expects this custom action to be exposed on the page.")
				}
			case "page":
				analysis.addPage(expectation.Value, "Test expects this page to exist in generated navigation metadata.")
			}
		}
	}
	analysis.addGeneratedFile("src/blacklang.browser.test.tsx", "Generated browser checks assert declared test expectations.")
	analysis.addGeneratedFile("src/blacklang.e2e.test.ts", "Generated browser e2e checks execute declared expectations in a real browser.")
	analysis.addGeneratedFile("src/blacklang.e2e.matrix.ts", "Generated browser e2e matrix runner executes declared expectations across available browser targets.")
	analysis.addGeneratedFile("tests/browser-matrix.json", "Generated browser matrix manifest exposes deterministic target and command metadata.")
	analysis.addGeneratedFile("src/App.tsx", "Generated browser checks and e2e load the React app shell.")
	analysis.addGeneratedFile("package.json", "Generated package test scripts include browser checks and browser e2e when tests are declared.")
	analysis.AgentNotes = append(analysis.AgentNotes, "Test changes affect generated smoke, browser e2e, and browser matrix checks; they do not change application runtime behavior.")
}

func populateTestsAffected(analysis *AffectedAnalysis, program Program) {
	for _, test := range program.Tests {
		analysis.addTest(test.Name, "Test declaration contributes generated browser checks and browser e2e checks.")
		if page, ok := affectedFindPage(program, test.Page); ok {
			analysis.addPage(page.Name, "Browser test targets this page.")
			analysis.addEntity(page.Source, "Browser test reads generated UI intent for this page source.")
		}
	}
	analysis.addGeneratedFile("src/blacklang.browser.test.tsx", "Generated browser checks assert declared test expectations.")
	analysis.addGeneratedFile("src/blacklang.e2e.test.ts", "Generated browser e2e checks execute declared expectations in a real browser.")
	analysis.addGeneratedFile("src/blacklang.e2e.matrix.ts", "Generated browser e2e matrix runner executes declared expectations across available browser targets.")
	analysis.addGeneratedFile("tests/browser-matrix.json", "Generated browser matrix manifest exposes deterministic target and command metadata.")
	analysis.addGeneratedFile("src/App.tsx", "Generated browser checks and e2e load the React app shell.")
	analysis.addGeneratedFile("package.json", "Generated package test scripts include browser checks and browser e2e when tests are declared.")
	analysis.AgentNotes = append(analysis.AgentNotes, "Run generated npm test, npm run test:e2e, npm run test:e2e:plan, and npm run test:e2e:matrix after editing test declarations.")
}

func populateRoleAffected(analysis *AffectedAnalysis, program Program, role RoleDecl) {
	analysis.addRole(role.Name, "The symbol is this role.")
	analysis.addGeneratedFile("src/auth/UsersPage.tsx", "Generated role management uses declared roles.")
	analysis.addGeneratedFile("src/auth/AuditPage.tsx", "Generated audit UI is role-aware.")
	analysis.addGeneratedFile("src/routes/auth.ts", "Auth routes store and update user roles.")
	for _, query := range program.Queries {
		for _, permission := range role.Permissions {
			if permission.Action == "all" || (permission.Resource == query.Source && (permission.Action == "read" || permission.Action == "manage")) {
				analysis.addQuery(query.Name, "Query execution checks entity and filter/sort field read permissions.")
			}
		}
		for _, page := range program.Pages {
			if page.Query == query.Name && containsString(page.Access, role.Name) {
				analysis.addQuery(query.Name, "Bound page access includes this role.")
			}
		}
	}
	for _, action := range program.Actions {
		for _, permission := range role.Permissions {
			if permission.Action == "all" || (permission.Resource == action.Source && (permission.Action == "update" || permission.Action == "manage")) {
				analysis.addAction(action.Name, "Custom action checks update permission and writable fields.")
			}
		}
		if containsString(action.Allow, role.Name) {
			analysis.addAction(action.Name, "Custom action allow list includes this role.")
		}
	}

	for _, permission := range role.Permissions {
		if permission.Resource == "" || permission.Action == "all" {
			for _, page := range program.Pages {
				analysis.addPage(page.Name, "Role has global permissions that can affect page actions.")
				addGeneratedPageFiles(analysis, program, page, page.Source, "Role guards affect generated page/API behavior.")
			}
			continue
		}
		analysis.addEntity(permission.Resource, "Role permission references this resource.")
		for _, page := range program.Pages {
			if page.Source == permission.Resource {
				analysis.addPage(page.Name, "Page source matches role permission resource.")
				addGeneratedPageFiles(analysis, program, page, page.Source, "Role permission guards affect generated page/API behavior.")
			}
		}
	}
	for _, page := range program.Pages {
		if containsString(page.Access, role.Name) {
			analysis.addPage(page.Name, "Page access list includes "+role.Name+".")
			addGeneratedPageFiles(analysis, program, page, page.Source, "Page-level role guard includes "+role.Name+".")
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
			addGeneratedPageFiles(analysis, program, page, workflow.Source, "Workflow routes, clients, and buttons are generated for this source.")
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
						if page.Source == entity.Name && pageUsesField(program, page, field.Name) {
							analysis.addPage(page.Name, "Generated page may render "+field.Name+" through "+component.Name+".")
							analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", "Page imports matching generated component.")
						}
					}
				}
			}
		}
	}
	for _, page := range program.Pages {
		for _, section := range pageViewComponentSections(page) {
			if section.Component != component.Name {
				continue
			}
			analysis.addPage(page.Name, "Page view section "+section.Name+" renders "+component.Name+".")
			analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", "Page imports this component for a generated component section.")
		}
	}
}

func affectedFindViewSection(program Program, symbol string) (PageDecl, ViewSectionDecl, bool) {
	for _, page := range program.Pages {
		if page.View == nil {
			continue
		}
		for _, section := range page.View.Sections {
			if section.Name == symbol && section.Component != "" {
				return page, section, true
			}
		}
	}
	return PageDecl{}, ViewSectionDecl{}, false
}

func affectedFindPageViewSection(program Program, pageName string, sectionName string) (PageDecl, ViewSectionDecl, bool) {
	for _, page := range program.Pages {
		if page.Name != pageName || page.View == nil {
			continue
		}
		for _, section := range page.View.Sections {
			if section.Name == sectionName {
				return page, section, true
			}
		}
	}
	return PageDecl{}, ViewSectionDecl{}, false
}

func populateViewSectionAffected(analysis *AffectedAnalysis, program Program, page PageDecl, section ViewSectionDecl) {
	analysis.Entity = page.Source
	analysis.addPage(page.Name, "Page view declares section "+section.Name+".")
	analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", "Generated page renders this view section in effective DOM order.")
	analysis.addGeneratedFile("src/styles.css", "Generated styles include stable order, span, display, and component-section classes.")
	if section.Component != "" {
		analysis.addComponent(section.Component, "View section "+section.Name+" renders this component.")
		analysis.addGeneratedFile("src/components/"+section.Component+".tsx", "Generated page imports this component for the section.")
	}
	if entity, ok := affectedFindEntity(program, page.Source); ok {
		analysis.addEntity(entity.Name, "View section belongs to a page sourced from "+entity.Name+".")
		if component, ok := affectedFindComponent(program, section.Component); ok {
			for _, input := range component.Inputs {
				if field, fieldOK := findField(entity, input.Name); fieldOK && field.Type == input.Type {
					analysis.addEntity(entity.Name, "Component section input "+input.Name+" binds to stored field "+entity.Name+"."+field.Name+".")
				}
				if computed, computedOK := findComputedField(entity, input.Name); computedOK && computed.Type == input.Type {
					analysis.addEntity(entity.Name, "Component section input "+input.Name+" binds to computed field "+entity.Name+"."+computed.Name+".")
				}
			}
		}
	}
	if section.Bind == "each" {
		analysis.AgentNotes = append(analysis.AgentNotes, "bind each renders one component instance per loaded list/query record and reuses the page items collection.")
	}
}

func populateAPIAffected(analysis *AffectedAnalysis, program Program, api APIDecl) {
	analysis.addAPI(api.Name, "The symbol is this explicit API declaration.")
	analysis.addGeneratedFile("src/server.ts", "Generated server mounts the declared explicit API route and any typed handler runtime.")
	analysis.addGeneratedFile("openapi.json", "Explicit API declarations, typed body schemas, and handler metadata are written to the OpenAPI contract.")
	analysis.addGeneratedFile("src/blacklang.contract.test.ts", "Generated contract tests assert explicit API paths and metadata.")
	analysis.addGeneratedFile("src/blacklang.api.test.ts", "Generated API smoke tests probe declared explicit API runtime routes.")
	if api.Update != nil {
		analysis.addEntity(api.Update.Source, "Explicit API update handler reads and mutates this entity.")
		if transactionName, ok := transactionAPIBindings(program)[api.Name]; ok {
			analysis.addTransaction(transactionName, "Transaction block wraps this explicit API update handler in a generated Prisma transaction.")
		}
		if serviceName, ok := serviceAPIBindings(program)[api.Name]; ok {
			analysis.addService(serviceName, "Service block groups this explicit API for generated module metadata.")
			analysis.addGeneratedFile("services/manifest.json", "Generated service manifest references this explicit API.")
			analysis.addGeneratedFile("src/services/"+kebabCase(serviceName)+".ts", "Generated service module references this explicit API.")
		}
		fields := apiUpdateSetFieldNames(*api.Update)
		analysis.AgentNotes = append(analysis.AgentNotes, "Explicit API update handlers validate typed body/path values, run one bounded stored-field update, apply policy scope, and return deterministic JSON. This API writes: "+strings.Join(fields, ", ")+".")
		return
	}
	if serviceName, ok := serviceAPIBindings(program)[api.Name]; ok {
		analysis.addService(serviceName, "Service block groups this explicit API for generated module metadata.")
		analysis.addGeneratedFile("services/manifest.json", "Generated service manifest references this explicit API.")
		analysis.addGeneratedFile("src/services/"+kebabCase(serviceName)+".ts", "Generated service module references this explicit API.")
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "Explicit API runtime routes validate declared path and query parameters and return a deterministic declared response.")
}

func populateAuthAffected(analysis *AffectedAnalysis, program Program) {
	analysis.addGeneratedFile("src/auth/AuthPage.tsx", "Auth declaration generates login/register UI.")
	analysis.addGeneratedFile("src/routes/auth.ts", "Auth declaration generates auth API routes.")
	analysis.addGeneratedFile("src/App.tsx", "Generated app restores auth state and renders protected shell.")
	analysis.addGeneratedFile("src/server.ts", "Generated server mounts auth routes and protects CRUD routes.")
	for _, page := range program.Pages {
		analysis.addPage(page.Name, "Auth affects generated access checks for pages.")
		if page.Query != "" {
			analysis.addQuery(page.Query, "Bound query endpoint uses the page's authentication checks.")
		}
		for _, action := range customActionsForPage(program, page) {
			analysis.addAction(action.Name, "Bound custom action endpoint uses the page's authentication checks.")
		}
	}
}

func populateDatabaseAffected(analysis *AffectedAnalysis, program Program) {
	analysis.addGeneratedFile(".env.example", "Database env reference is documented for local setup.")
	analysis.addGeneratedFile("prisma/schema.prisma", "Database provider and schema are emitted for generated persistence.")
	analysis.addGeneratedFile("src/db.ts", "Generated Prisma client reads database configuration.")
	analysis.addGeneratedFile("src/setup-db.ts", "Generated database setup prepares the selected SQLite or PostgreSQL runtime.")
	if len(program.Seeds) > 0 {
		analysis.addGeneratedFile("src/seed.ts", "Generated seed runtime applies declared fixture rows after schema setup.")
		analysis.addGeneratedFile("package.json", "Generated package scripts include db:seed when seeds are declared.")
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "Before deploying schema-affecting entity changes to an existing database, run black migrate plan old.black new.black --json.")
}

func addMigrationGeneratedFiles(analysis *AffectedAnalysis) {
	analysis.addGeneratedFile("migrations/manifest.json", "Generated migration manifest records first-class rename declarations.")
	analysis.addGeneratedFile("migrations/*.sql", "Generated migration SQL previews deterministic rename operations.")
	analysis.addGeneratedFile("src/migrate.ts", "Generated migration runner provides read-only db:migrate:plan and explicit db:migrate apply modes.")
	analysis.addGeneratedFile("src/setup-db.ts", "Generated setup applies rename migrations before schema setup.")
	analysis.addGeneratedFile("prisma/schema.prisma", "Generated schema includes the migration ledger when migrations exist.")
	analysis.addGeneratedFile("package.json", "Generated package scripts include db:migrate:plan and db:migrate when migrations exist.")
}

func populateSecurityAffected(analysis *AffectedAnalysis, program Program) {
	analysis.addGeneratedFile(".env.example", "Security env references are documented for local setup.")
	analysis.addGeneratedFile("security/secrets.json", "Generated secret manifest records env references, provider labels, and provider preflight script metadata without values.")
	analysis.addGeneratedFile("scripts/secrets-plan.mjs", "Generated secret plan reports environment readiness without printing values.")
	analysis.addGeneratedFile("scripts/secrets-provider.mjs", "Generated secret provider preflight checks selected provider label, prefix, and CLI availability without fetching values.")
	analysis.addGeneratedFile("package.json", "Generated package scripts include security:secrets:plan and security:secrets:preflight.")
	analysis.addGeneratedFile("src/server.ts", "Generated server emits security middleware from security intent.")
	if program.Deploy != nil && program.Deploy.Target == "docker" {
		analysis.addGeneratedFile("docker-compose.yml", "Generated compose environment mirrors security env references.")
	}
}

func populatePolicyAffected(analysis *AffectedAnalysis, program Program, symbol string) {
	generator := &webGenerator{program: program}
	for _, entity := range program.Entities {
		for _, policy := range entity.Policies {
			if symbol == "owner" && policy.Kind != "owner" {
				continue
			}
			if symbol == "tenant" && policy.Kind != "tenant" {
				continue
			}
			analysis.addEntity(entity.Name, title(policy.Kind)+" policy is declared on "+entity.Name+"."+policy.Field+".")
			analysis.addGeneratedFile("prisma/schema.prisma", "Policy fields remain normal stored fields in the generated schema.")
			analysis.addGeneratedFile("src/setup-db.ts", "Generated database setup creates policy fields as stored columns.")
			if program.Auth != nil {
				analysis.addGeneratedFile("src/routes/auth.ts", "Generated auth runtime exposes current user identity for row policy checks.")
			}
			for _, page := range program.Pages {
				if page.Source != entity.Name {
					continue
				}
				module := generator.pageModuleName(page)
				analysis.addPage(page.Name, "Generated CRUD/query/action routes for this page apply "+policy.Kind+" row policy.")
				analysis.addGeneratedFile("src/routes/"+module+".ts", "Generated route applies "+policy.Kind+" row scope to list, detail, mutations, actions, and workflow transitions.")
				analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", "Generated page omits policy fields from forms and receives scoped API data.")
			}
		}
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "Owner policy scopes rows by current user id; tenant policy scopes rows by current user tenant id.")
	analysis.AgentNotes = append(analysis.AgentNotes, "Policy fields are generated stored fields but generated create/update routes control their values.")
}

func programHasEntityPolicies(program Program, symbol string) bool {
	for _, entity := range program.Entities {
		for _, policy := range entity.Policies {
			if symbol == "owner" && policy.Kind != "owner" {
				continue
			}
			if symbol == "tenant" && policy.Kind != "tenant" {
				continue
			}
			return true
		}
	}
	return false
}

func populateDeployAffected(analysis *AffectedAnalysis, program Program) {
	analysis.addGeneratedFile(".env.example", "Deployment env defaults are documented for local setup.")
	analysis.addGeneratedFile(".dockerignore", "Docker build context exclusions are generated from deploy intent.")
	analysis.addGeneratedFile("Dockerfile", "Docker image build instructions are generated from deploy intent.")
	analysis.addGeneratedFile("docker-compose.yml", "Docker Compose service wiring is generated from deploy intent.")
	analysis.addGeneratedFile("package.json", "Generated package scripts include the production start command.")
	analysis.addGeneratedFile("src/server.ts", "Generated server reads the configured deploy port environment variable.")
	if program.Deploy != nil && program.Deploy.Preview != nil && program.Deploy.Preview.Mode == "local" {
		analysis.addGeneratedFile("docker-compose.preview.yml", "Generated local preview compose stack uses isolated preview ports and data volume.")
		analysis.addGeneratedFile("package.json", "Generated package scripts include deploy:preview and deploy:preview:down.")
	}
	if program.Deploy != nil && program.Deploy.Rollback != nil {
		analysis.addGeneratedFile("deploy/manifest.json", "Generated deployment manifest records target, env, preview, rollback, and generated artifact metadata.")
		analysis.addGeneratedFile("deploy/rollback.json", "Generated rollback metadata records the deterministic keep policy.")
		analysis.addGeneratedFile("scripts/rollback-plan.mjs", "Generated rollback plan script reports retained releases and rollback candidate without mutating infrastructure.")
		analysis.addGeneratedFile("package.json", "Generated package scripts include deploy:rollback:plan.")
	}
	if program.Deploy != nil && program.Deploy.Cloud != nil {
		analysis.addGeneratedFile("deploy/manifest.json", "Generated deployment manifest records cloud adapter metadata.")
		analysis.addGeneratedFile("deploy/cloud.json", "Generated cloud adapter manifest records provider, image, app env, region env, required env, provider CLI, and explicit apply policy.")
		analysis.addGeneratedFile("scripts/cloud-plan.mjs", "Generated cloud plan script reports provider readiness without mutating infrastructure.")
		analysis.addGeneratedFile("scripts/cloud-exec.mjs", "Generated cloud execution script runs read-only preflight by default and requires explicit apply mode for provider CLI deployment.")
		analysis.addGeneratedFile("package.json", "Generated package scripts include deploy:cloud:plan, deploy:cloud:preflight, and deploy:cloud:exec.")
	}
}

func programHasOpsSignal(program Program, symbol string) bool {
	if program.Ops == nil {
		return false
	}
	switch symbol {
	case "health":
		return program.Ops.Health != nil
	case "readiness":
		return program.Ops.Readiness != nil
	case "metrics":
		return program.Ops.Metrics != nil
	case "logging":
		return program.Ops.Logging != ""
	case "observe", "observability":
		return program.Ops.Observe != nil
	default:
		return program.Ops.Health != nil || program.Ops.Readiness != nil || program.Ops.Metrics != nil || program.Ops.Logging != "" || program.Ops.Observe != nil
	}
}

func populateOpsAffected(analysis *AffectedAnalysis, program Program) {
	analysis.addGeneratedFile(".env.example", "Ops env references such as observability endpoints are documented for local setup.")
	if program.Ops != nil && program.Ops.Observe != nil {
		analysis.addGeneratedFile("ops/observability.json", "Generated observability manifest records provider, endpoint env, trace context, delivery mode, and exporter metadata.")
	}
	analysis.addGeneratedFile("src/server.ts", "Generated server emits ops endpoints, readiness checks, metrics counters, request logging, trace context, and observability hooks.")
	analysis.addGeneratedFile("openapi.json", "Generated OpenAPI contract includes public ops endpoints and observability/exporter metadata when declared.")
	analysis.addGeneratedFile("src/blacklang.contract.test.ts", "Generated contract tests assert declared ops paths and observability/exporter metadata.")
	analysis.addGeneratedFile("src/blacklang.api.test.ts", "Generated API smoke tests probe declared ops endpoints and local observability hook or OTLP trace delivery.")
	if program.Deploy != nil && program.Deploy.Target == "docker" {
		analysis.addGeneratedFile("Dockerfile", "Generated Docker image healthcheck probes the declared health endpoint.")
		analysis.addGeneratedFile("docker-compose.yml", "Generated Docker Compose service healthcheck probes the declared health endpoint and carries ops env references.")
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "Ops endpoints are public runtime probes outside /api so auth and CSRF middleware do not block deployment health checks.")
	analysis.AgentNotes = append(analysis.AgentNotes, "Ops observability hooks use env-referenced endpoints, W3C traceparent context, and non-blocking delivery.")
}

func populateI18NAffected(analysis *AffectedAnalysis, program Program) {
	analysis.addGeneratedFile("src/App.tsx", "Runtime language selector, app chrome labels, navigation labels, lang, and dir attributes depend on i18n.")
	for _, page := range program.Pages {
		analysis.addPage(page.Name, "Generated page receives locale and may render localized field text, UI copy, or formatted values.")
		analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", "Generated page field labels, placeholders, help text, field messages, table/status/action copy, and display formatting depend on i18n.")
	}
	for _, label := range program.Labels {
		if entityName, _, ok := splitLabelTarget(label.Target); ok {
			if _, exists := affectedFindEntity(program, entityName); exists {
				analysis.addEntity(entityName, "Label translation target is "+label.Target+".")
			}
		}
	}
	blocks := append([]FieldTextTranslationDecl{}, program.Placeholders...)
	blocks = append(blocks, program.HelpTexts...)
	blocks = append(blocks, program.Messages...)
	for _, block := range blocks {
		if entityName, _, ok := splitLabelTarget(block.Target); ok {
			analysis.addEntity(entityName, "Field text translation target is "+block.Target+".")
		}
	}
	analysis.AgentNotes = append(analysis.AgentNotes, "i18n changes do not rename stored fields or database columns; UI label targets change generated app chrome/action/table/status copy and require rebuilding web output.")
}

func populateTargetAffected(analysis *AffectedAnalysis, program Program) {
	analysis.addGeneratedFile("README.md", "Generated summary documents the selected target stack.")
	analysis.addGeneratedFile("package.json", "Generated package dependencies and scripts depend on the target stack.")
	analysis.addGeneratedFile("prisma.config.ts", "Database target selection affects Prisma configuration.")
	analysis.addGeneratedFile("prisma/schema.prisma", "Database target selection affects generated schema provider.")
	analysis.addGeneratedFile("src/server.ts", "Backend target selection affects generated server runtime.")
	analysis.addGeneratedFile("src/blacklang.contract.test.ts", "Generated contract tests depend on the target stack.")
	analysis.addGeneratedFile("src/blacklang.api.test.ts", "Generated API smoke tests depend on the target stack.")
	if !(&webGenerator{program: program}).isAPIOnlyTarget() {
		analysis.addGeneratedFile("index.html", "Web target selection affects the generated HTML entry.")
		analysis.addGeneratedFile("vite.config.ts", "Frontend target selection affects Vite configuration.")
		analysis.addGeneratedFile("src/main.tsx", "Web target selection affects the React entry.")
		analysis.addGeneratedFile("src/App.tsx", "Web target selection affects the generated React shell.")
		analysis.addGeneratedFile("src/styles.css", "Web target selection affects generated frontend styles.")
		analysis.addGeneratedFile("src/vite-env.d.ts", "Web target selection affects Vite type declarations.")
		analysis.addGeneratedFile("src/blacklang.frontend.test.tsx", "Web target selection affects generated frontend smoke tests.")
	}
	if program.Deploy != nil && program.Deploy.Target == "docker" {
		analysis.addGeneratedFile("Dockerfile", "Docker build instructions depend on the selected target stack.")
		analysis.addGeneratedFile("docker-compose.yml", "Docker Compose service wiring depends on target stack runtime.")
	}
	if (&webGenerator{program: program}).isAPIOnlyTarget() {
		analysis.AgentNotes = append(analysis.AgentNotes, "target api emits Express, Prisma, OpenAPI, validation, contract/API tests, and optional seed/job/deploy output without React, Vite, frontend smoke, browser-check, browser e2e, or browser matrix files.")
	}
}

func populateAppAffected(analysis *AffectedAnalysis, program Program) {
	if program.App.Name != "" {
		analysis.addGeneratedFile("package.json", "Generated package metadata uses the app name.")
		analysis.addGeneratedFile("index.html", "Generated HTML title uses the app name.")
		analysis.addGeneratedFile("src/App.tsx", "Generated shell displays the app name.")
	}
}

func addGeneratedPageFiles(analysis *AffectedAnalysis, program Program, page PageDecl, entityName string, reason string) {
	apiOnly := (&webGenerator{program: program}).isAPIOnlyTarget()
	if !apiOnly {
		analysis.addGeneratedFile("src/pages/"+page.Name+"Page.tsx", reason)
	}
	if entityName != "" {
		fileName := (&webGenerator{program: program}).pageModuleName(page)
		analysis.addGeneratedFile("src/api/"+fileName+".ts", reason)
		analysis.addGeneratedFile("src/routes/"+fileName+".ts", reason)
		analysis.addGeneratedFile("src/validation/"+strings.ToLower(entityName)+".ts", reason)
	}
	if !apiOnly {
		analysis.addGeneratedFile("src/App.tsx", "Generated navigation and route registration include pages.")
	}
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
			analysis.addGeneratedFile("src/server.ts", "Matching explicit API runtime route is generated into the server.")
			analysis.addGeneratedFile("openapi.json", "Matching explicit API contract is generated into OpenAPI.")
			continue
		}
		if api.Update != nil && explicitAPIUpdateMentions(*api.Update, needle) {
			analysis.addAPI(api.Name, "Explicit API update handler mentions "+needle+".")
			analysis.addGeneratedFile("src/server.ts", "Matching explicit API handler is generated into the server.")
			analysis.addGeneratedFile("openapi.json", "Matching explicit API handler metadata is generated into OpenAPI.")
		}
	}
}

func explicitAPIUpdateMentions(update APIUpdateDecl, needle string) bool {
	lowerNeedle := strings.ToLower(needle)
	if strings.Contains(strings.ToLower(update.Source), lowerNeedle) || strings.Contains(strings.ToLower(update.Source+"."+update.Where.Field), lowerNeedle) || strings.Contains(strings.ToLower(update.Where.Field), lowerNeedle) {
		return true
	}
	if explicitAPIOperandMentions(update.Where.Value, lowerNeedle) {
		return true
	}
	for _, statement := range apiUpdateStatements(update) {
		if explicitAPIStatementMentions(statement, update.Source, lowerNeedle) {
			return true
		}
	}
	return false
}

func explicitAPIStatementMentions(statement APIStatementDecl, source string, lowerNeedle string) bool {
	switch statement.Kind {
	case "value":
		if statement.Value == nil {
			return false
		}
		if strings.Contains(strings.ToLower(statement.Value.Name), lowerNeedle) {
			return true
		}
		return explicitAPIExpressionDeclMentions(statement.Value.Expression, lowerNeedle)
	case "set":
		if statement.Set == nil {
			return false
		}
		if strings.Contains(strings.ToLower(source+"."+statement.Set.Field), lowerNeedle) || strings.Contains(strings.ToLower(statement.Set.Field), lowerNeedle) {
			return true
		}
		return explicitAPIExpressionDeclMentions(statement.Set.Expr, lowerNeedle)
	case "if":
		if statement.If == nil {
			return false
		}
		if statement.If.Condition.Tree != nil && explicitAPIConditionExpressionMentions(*statement.If.Condition.Tree, lowerNeedle) {
			return true
		}
		if explicitAPIExpressionDeclMentions(statement.If.Condition.Left, lowerNeedle) || explicitAPIExpressionDeclMentions(statement.If.Condition.Right, lowerNeedle) {
			return true
		}
		for _, branchStatement := range statement.If.Then {
			if explicitAPIStatementMentions(branchStatement, source, lowerNeedle) {
				return true
			}
		}
		for _, branchStatement := range statement.If.Else {
			if explicitAPIStatementMentions(branchStatement, source, lowerNeedle) {
				return true
			}
		}
	}
	return false
}

func explicitAPIConditionExpressionMentions(condition ConditionExpressionDecl, lowerNeedle string) bool {
	switch condition.Kind {
	case "comparison":
		return condition.Comparison != nil && (explicitAPIExpressionMentions(condition.Comparison.Left, lowerNeedle) || explicitAPIExpressionMentions(condition.Comparison.Right, lowerNeedle))
	case "and", "or":
		return (condition.Left != nil && explicitAPIConditionExpressionMentions(*condition.Left, lowerNeedle)) || (condition.Right != nil && explicitAPIConditionExpressionMentions(*condition.Right, lowerNeedle))
	case "not":
		return condition.Left != nil && explicitAPIConditionExpressionMentions(*condition.Left, lowerNeedle)
	default:
		return false
	}
}

func explicitAPIExpressionDeclMentions(expression APIExpressionDecl, lowerNeedle string) bool {
	if expression.Tree != nil && explicitAPIExpressionMentions(*expression.Tree, lowerNeedle) {
		return true
	}
	return explicitAPIOperandMentions(expression.Left, lowerNeedle) || (expression.Right != nil && explicitAPIOperandMentions(*expression.Right, lowerNeedle))
}

func explicitAPIOperandMentions(operand APIOperandDecl, lowerNeedle string) bool {
	if strings.Contains(strings.ToLower(operand.Value), lowerNeedle) {
		return true
	}
	if operand.Kind == "body" || operand.Kind == "param" {
		return strings.Contains(strings.ToLower(operand.Kind+"."+operand.Value), lowerNeedle)
	}
	return false
}

func explicitAPIExpressionMentions(expression ExpressionDecl, lowerNeedle string) bool {
	for _, operand := range expressionOperands(expression) {
		apiOperand := apiOperandFromExpression(operand)
		if explicitAPIOperandMentions(apiOperand, lowerNeedle) {
			return true
		}
	}
	return false
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

func missingPolicyDiagnostic(analysis AffectedAnalysis, symbol string) []Diagnostic {
	if analysis.Found {
		return nil
	}
	return []Diagnostic{{
		Code:       "UNKNOWN_AFFECTED_SYMBOL",
		Message:    fmt.Sprintf("Affected policy symbol %q was not found.", symbol),
		Suggestion: "Add `policy owner ownerId` or `policy tenant tenantId` inside an entity before inspecting policy impact.",
	}}
}

func missingOpsDiagnostic(analysis AffectedAnalysis, symbol string) []Diagnostic {
	if analysis.Found {
		return nil
	}
	return []Diagnostic{{
		Code:       "UNKNOWN_AFFECTED_SYMBOL",
		Message:    fmt.Sprintf("Affected ops symbol %q was not found.", symbol),
		Suggestion: "Add a top-level ops block with `health path \"/healthz\"` or inspect an existing ops signal.",
	}}
}

func missingEntityIndexDiagnostic(analysis AffectedAnalysis) []Diagnostic {
	if analysis.Found {
		return nil
	}
	return []Diagnostic{{
		Code:       "UNKNOWN_AFFECTED_SYMBOL",
		Message:    fmt.Sprintf("Affected symbol %q was not found.", analysis.Symbol),
		Suggestion: "Add an entity index such as `index sku` before inspecting Product.index.",
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

func affectedFindMigration(program Program, name string) (MigrationDecl, bool) {
	for _, migration := range program.Migrations {
		if migration.Name == name {
			return migration, true
		}
	}
	return MigrationDecl{}, false
}

func customActionUsesField(action CustomActionDecl, fieldName string) bool {
	for _, statement := range actionStatements(action) {
		if actionStatementUsesField(statement, fieldName) {
			return true
		}
	}
	return false
}

func actionStatementUsesField(statement ActionStatementDecl, fieldName string) bool {
	switch statement.Kind {
	case "value":
		return statement.Value != nil && actionExpressionUsesField(statement.Value.Expression, fieldName)
	case "set":
		if statement.Set == nil {
			return false
		}
		return statement.Set.Field == fieldName || actionExpressionUsesField(statement.Set.Expression, fieldName)
	case "if":
		if statement.If == nil {
			return false
		}
		if statement.If.Condition.Tree != nil && actionConditionExpressionUsesField(*statement.If.Condition.Tree, fieldName) {
			return true
		}
		if actionExpressionUsesField(statement.If.Condition.Left, fieldName) || actionExpressionUsesField(statement.If.Condition.Right, fieldName) {
			return true
		}
		for _, branchStatement := range statement.If.Then {
			if actionStatementUsesField(branchStatement, fieldName) {
				return true
			}
		}
		for _, branchStatement := range statement.If.Else {
			if actionStatementUsesField(branchStatement, fieldName) {
				return true
			}
		}
	}
	return false
}

func actionConditionExpressionUsesField(condition ConditionExpressionDecl, fieldName string) bool {
	switch condition.Kind {
	case "comparison":
		return condition.Comparison != nil && (expressionUsesActionField(condition.Comparison.Left, fieldName) || expressionUsesActionField(condition.Comparison.Right, fieldName))
	case "and", "or":
		return (condition.Left != nil && actionConditionExpressionUsesField(*condition.Left, fieldName)) || (condition.Right != nil && actionConditionExpressionUsesField(*condition.Right, fieldName))
	case "not":
		return condition.Left != nil && actionConditionExpressionUsesField(*condition.Left, fieldName)
	default:
		return false
	}
}

func expressionUsesActionField(expression ExpressionDecl, fieldName string) bool {
	for _, operand := range expressionOperands(expression) {
		if operand.ValueKind == "reference" && operand.Value == fieldName {
			return true
		}
	}
	return false
}

func actionExpressionUsesField(expression ActionExpressionDecl, fieldName string) bool {
	if expression.Tree != nil {
		for _, operand := range expressionOperands(*expression.Tree) {
			if operand.ValueKind == "reference" && operand.Value == fieldName {
				return true
			}
		}
		return false
	}
	if expression.Left.Kind == "reference" && expression.Left.Value == fieldName {
		return true
	}
	if expression.Right != nil && expression.Right.Kind == "reference" && expression.Right.Value == fieldName {
		return true
	}
	return false
}

func pageForState(program Program, state StateDecl) (PageDecl, bool) {
	for _, page := range program.Pages {
		if state.Name == page.Name+"State" || state.Name == page.Name+"PageState" {
			return page, true
		}
	}
	return PageDecl{}, false
}

func pageUsesField(program Program, page PageDecl, fieldName string) bool {
	if containsString(page.Table.Columns, fieldName) ||
		containsString(page.Table.Search, fieldName) ||
		containsString(page.Table.Filters, fieldName) ||
		page.Table.Sort.Field == fieldName ||
		containsString(page.Form.Fields, fieldName) {
		return true
	}
	return pageComponentSectionsUseField(program, page, fieldName)
}

func pageComponentSectionsUseField(program Program, page PageDecl, fieldName string) bool {
	entity, _ := affectedFindEntity(program, page.Source)
	for _, section := range pageViewComponentSections(page) {
		component, ok := affectedFindComponent(program, section.Component)
		if !ok {
			continue
		}
		for _, input := range component.Inputs {
			if input.Name == fieldName {
				return true
			}
			if computed, ok := findComputedField(entity, input.Name); ok && containsString(computedFieldReferences(computed), fieldName) {
				return true
			}
		}
	}
	return false
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

func (analysis *AffectedAnalysis) addMigration(name string, reason string) {
	analysis.Migrations = addAffectedItem(analysis.Migrations, name, reason)
}

func (analysis *AffectedAnalysis) addPage(name string, reason string) {
	analysis.Pages = addAffectedItem(analysis.Pages, name, reason)
}

func (analysis *AffectedAnalysis) addQuery(name string, reason string) {
	analysis.Queries = addAffectedItem(analysis.Queries, name, reason)
}

func (analysis *AffectedAnalysis) addJob(name string, reason string) {
	analysis.Jobs = addAffectedItem(analysis.Jobs, name, reason)
}

func (analysis *AffectedAnalysis) addAction(name string, reason string) {
	analysis.Actions = addAffectedItem(analysis.Actions, name, reason)
}

func (analysis *AffectedAnalysis) addTransaction(name string, reason string) {
	analysis.Transactions = addAffectedItem(analysis.Transactions, name, reason)
}

func (analysis *AffectedAnalysis) addService(name string, reason string) {
	analysis.Services = addAffectedItem(analysis.Services, name, reason)
}

func (analysis *AffectedAnalysis) addSeed(name string, reason string) {
	analysis.Seeds = addAffectedItem(analysis.Seeds, name, reason)
}

func (analysis *AffectedAnalysis) addTest(name string, reason string) {
	analysis.Tests = addAffectedItem(analysis.Tests, name, reason)
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
