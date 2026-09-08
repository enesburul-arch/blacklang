package main

import (
	"fmt"
	"strconv"
	"strings"
)

type parser struct {
	file        string
	lines       []sourceStatement
	program     Program
	diagnostics []Diagnostic
}

func Parse(file string, source string) (Program, []Diagnostic) {
	lines, diagnostics := tokenizeSource(file, source)
	p := parser{
		file:        file,
		lines:       lines,
		diagnostics: diagnostics,
	}
	p.parse()
	return p.program, p.diagnostics
}

func (p *parser) parse() {
	for index := 0; index < len(p.lines); index++ {
		lineNumber := p.lineNumber(index)
		parts := p.partsAt(index)
		if len(parts) == 0 {
			continue
		}
		switch parts[0] {
		case "app":
			p.parseApp(parts, lineNumber)
		case "target":
			index = p.parseTarget(index, parts)
		case "auth":
			index = p.parseAuth(index, parts)
		case "database":
			index = p.parseDatabase(index, parts)
		case "security":
			index = p.parseSecurity(index, parts)
		case "deploy":
			index = p.parseDeploy(index, parts)
		case "ops":
			index = p.parseOps(index, parts)
		case "i18n":
			index = p.parseI18N(index, parts)
		case "label":
			index = p.parseLabelTranslation(index, parts)
		case "placeholder", "help", "message":
			index = p.parseFieldTextTranslation(index, parts, parts[0])
		case "entity":
			index = p.parseEntity(index, parts)
		case "migration":
			index = p.parseMigration(index, parts)
		case "seed":
			index = p.parseSeed(index, parts)
		case "test":
			index = p.parseTest(index, parts)
		case "query":
			index = p.parseQuery(index, parts)
		case "job":
			index = p.parseJob(index, parts)
		case "action":
			index = p.parseCustomAction(index, parts)
		case "transaction":
			index = p.parseTransaction(index, parts)
		case "service":
			index = p.parseService(index, parts)
		case "role":
			index = p.parseRole(index, parts)
		case "api":
			index = p.parseAPI(index, parts)
		case "layout":
			index = p.parseLayout(index, parts)
		case "page":
			index = p.parsePage(index, parts)
		case "workflow":
			index = p.parseWorkflow(index, parts)
		case "state":
			index = p.parseState(index, parts)
		case "component":
			index = p.parseComponent(index, parts)
		default:
			p.addError(lineNumber, 1, "UNEXPECTED_TOP_LEVEL", fmt.Sprintf("Unexpected top-level token %q.", parts[0]), "Use app, target, auth, database, security, deploy, ops, i18n, label, placeholder, help, message, entity, migration, seed, test, query, job, action, transaction, service, role, api, layout, page, workflow, state, or component at the top level.")
		}
	}
}

func (p *parser) parseApp(parts []string, lineNumber int) {
	if len(parts) != 2 {
		p.addError(lineNumber, 1, "INVALID_APP_DECLARATION", "App declaration must be `app Name`.", "Example: `app Warehouse`.")
		return
	}
	if p.program.App.Name != "" {
		p.addError(lineNumber, 1, "DUPLICATE_APP", "Only one app declaration is allowed.", "Keep a single `app` declaration per project.")
		return
	}
	p.program.App = AppDecl{
		Name:     parts[1],
		Position: p.position(lineNumber, 1),
	}
}

func (p *parser) parseTarget(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_TARGET_DECLARATION", "Target declaration must be `target name {`.", "Example: `target web { frontend react backend node database sqlite }` or `target api { backend node database mysql }`.")
		return start
	}
	if p.program.Target != nil {
		p.addError(lineNumber, 1, "DUPLICATE_TARGET", "Only one target declaration is allowed.", "Keep a single `target` block per project in v0.2.")
		return start
	}

	target := TargetDecl{
		Name:     parts[1],
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Target = &target
			return index
		}

		switch rowParts[0] {
		case "frontend":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_TARGET_FRONTEND", "Target frontend must be `frontend name`.", "Example: `frontend react`.")
				continue
			}
			if target.Frontend != "" {
				p.addError(currentLine, 1, "DUPLICATE_TARGET_FRONTEND", "Target frontend is already declared.", "Keep one frontend line inside target.")
				continue
			}
			target.Frontend = rowParts[1]
		case "backend":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_TARGET_BACKEND", "Target backend must be `backend name`.", "Example: `backend node`.")
				continue
			}
			if target.Backend != "" {
				p.addError(currentLine, 1, "DUPLICATE_TARGET_BACKEND", "Target backend is already declared.", "Keep one backend line inside target.")
				continue
			}
			target.Backend = rowParts[1]
		case "database":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_TARGET_DATABASE", "Target database must be `database name`.", "Example: `database sqlite`, `database postgres`, or `database mysql`.")
				continue
			}
			if target.Database != "" {
				p.addError(currentLine, 1, "DUPLICATE_TARGET_DATABASE", "Target database is already declared.", "Keep one database line inside target.")
				continue
			}
			target.Database = rowParts[1]
		default:
			p.addError(currentLine, 1, "UNEXPECTED_TARGET_TOKEN", fmt.Sprintf("Unexpected target token %q.", rowParts[0]), "Use frontend, backend, or database inside target.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_TARGET", fmt.Sprintf("Target %s is missing a closing brace.", target.Name), "Add `}` after the target body.")
	return len(p.lines) - 1
}

func (p *parser) parseAuth(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_AUTH_DECLARATION", "Auth declaration must be `auth {`.", "Example: `auth { strategy emailPassword }`.")
		return start
	}
	if p.program.Auth != nil {
		p.addError(lineNumber, 1, "DUPLICATE_AUTH", "Only one auth declaration is allowed.", "Keep a single `auth` block per project.")
		return start
	}

	auth := AuthDecl{
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Auth = &auth
			return index
		}

		switch rowParts[0] {
		case "strategy":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_AUTH_STRATEGY", "Auth strategy must be `strategy name`.", "Example: `strategy emailPassword`.")
				continue
			}
			auth.Strategy = rowParts[1]
		case "session":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_AUTH_SESSION", "Auth session must be `session name`.", "Example: `session cookie`.")
				continue
			}
			auth.Session = rowParts[1]
		case "user":
			var user UserDecl
			user, index = p.parseAuthUser(index, rowParts)
			auth.User = user
		default:
			p.addError(currentLine, 1, "UNEXPECTED_AUTH_TOKEN", fmt.Sprintf("Unexpected auth token %q.", rowParts[0]), "Use strategy, session, or user inside auth.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_AUTH", "Auth block is missing a closing brace.", "Add `}` after the auth body.")
	return len(p.lines) - 1
}

func (p *parser) parseDatabase(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_DATABASE_DECLARATION", "Database declaration must be `database {`.", "Example: `database { url env DATABASE_URL }`.")
		return start
	}
	if p.program.Database != nil {
		p.addError(lineNumber, 1, "DUPLICATE_DATABASE", "Only one database declaration is allowed.", "Keep a single `database` block per project.")
		return start
	}

	database := DatabaseDecl{
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Database = &database
			return index
		}

		switch rowParts[0] {
		case "url":
			if len(rowParts) != 3 || rowParts[1] != "env" {
				p.addError(currentLine, 1, "INVALID_DATABASE_URL", "Database url must reference an environment variable.", "Use `url env DATABASE_URL`.")
				continue
			}
			database.URL = EnvRef{
				Name:     rowParts[2],
				Position: p.position(currentLine, 1),
			}
		default:
			p.addError(currentLine, 1, "UNEXPECTED_DATABASE_TOKEN", fmt.Sprintf("Unexpected database token %q.", rowParts[0]), "Use `url env DATABASE_URL` inside database.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_DATABASE", "Database block is missing a closing brace.", "Add `}` after the database block.")
	return len(p.lines) - 1
}

func (p *parser) parseMigration(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_MIGRATION_DECLARATION", "Migration declaration must be `migration Name {`.", "Example: `migration RenameProductName { rename field Product.oldName to name }`.")
		return start
	}

	migration := MigrationDecl{
		Name:     parts[1],
		Renames:  []MigrationRenameDecl{},
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Migrations = append(p.program.Migrations, migration)
			return index
		}

		switch rowParts[0] {
		case "rename":
			rename, ok := p.parseMigrationRename(rowParts, currentLine)
			if ok {
				migration.Renames = append(migration.Renames, rename)
			}
		default:
			p.addError(currentLine, 1, "UNEXPECTED_MIGRATION_TOKEN", fmt.Sprintf("Unexpected migration token %q.", rowParts[0]), "Use rename declarations inside migration.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_MIGRATION", fmt.Sprintf("Migration %s is missing a closing brace.", migration.Name), "Add `}` after the migration body.")
	return len(p.lines) - 1
}

func (p *parser) parseMigrationRename(parts []string, lineNumber int) (MigrationRenameDecl, bool) {
	if len(parts) != 5 || parts[3] != "to" {
		p.addError(lineNumber, 1, "INVALID_MIGRATION_RENAME", "Migration rename must be `rename entity Old to New` or `rename field Entity.old to new`.", "Use `rename entity ProductItem to Product` or `rename field Product.oldName to name`.")
		return MigrationRenameDecl{}, false
	}
	switch parts[1] {
	case "entity":
		return MigrationRenameDecl{
			Kind:     "entity",
			From:     parts[2],
			To:       parts[4],
			Position: p.position(lineNumber, 1),
		}, true
	case "field":
		entityName, fieldName, ok := strings.Cut(parts[2], ".")
		if !ok || entityName == "" || fieldName == "" {
			p.addError(lineNumber, 1, "INVALID_MIGRATION_FIELD_RENAME", "Field rename source must be `Entity.oldField`.", "Use `rename field Product.oldName to name`.")
			return MigrationRenameDecl{}, false
		}
		return MigrationRenameDecl{
			Kind:     "field",
			Entity:   entityName,
			From:     fieldName,
			To:       parts[4],
			Position: p.position(lineNumber, 1),
		}, true
	default:
		p.addError(lineNumber, 1, "INVALID_MIGRATION_RENAME_KIND", fmt.Sprintf("Unsupported migration rename kind %q.", parts[1]), "Use entity or field.")
		return MigrationRenameDecl{}, false
	}
}

func (p *parser) parseSecurity(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_SECURITY_DECLARATION", "Security declaration must be `security {`.", "Start with `security {`, then add a nested `cors {` block.")
		return start
	}
	if p.program.Security != nil {
		p.addError(lineNumber, 1, "DUPLICATE_SECURITY", "Only one security declaration is allowed.", "Keep a single `security` block per project.")
		return start
	}

	security := SecurityDecl{
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Security = &security
			return index
		}

		switch rowParts[0] {
		case "cors":
			cors, next := p.parseCORS(index, rowParts)
			index = next
			if security.CORS != nil {
				p.addError(currentLine, 1, "DUPLICATE_CORS", "Only one cors block is allowed inside security.", "Keep a single `cors` block per security declaration.")
				continue
			}
			security.CORS = &cors
		default:
			p.addError(currentLine, 1, "UNEXPECTED_SECURITY_TOKEN", fmt.Sprintf("Unexpected security token %q.", rowParts[0]), "Use cors inside security.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_SECURITY", "Security block is missing a closing brace.", "Add `}` after the security block.")
	return len(p.lines) - 1
}

func (p *parser) parseCORS(start int, parts []string) (CORSDecl, int) {
	lineNumber := p.lineNumber(start)
	cors := CORSDecl{
		Position: p.position(lineNumber, 1),
	}
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_CORS_DECLARATION", "CORS declaration must be `cors {`.", "Start with `cors {`, then add `origins env CORS_ORIGINS` inside it.")
		return cors, start
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			return cors, index
		}

		switch rowParts[0] {
		case "origins":
			if len(rowParts) != 3 || rowParts[1] != "env" {
				p.addError(currentLine, 1, "INVALID_CORS_ORIGINS", "CORS origins must reference an environment variable.", "Use `origins env CORS_ORIGINS`.")
				continue
			}
			if cors.Origins.Name != "" {
				p.addError(currentLine, 1, "DUPLICATE_CORS_ORIGINS", "CORS origins are already declared.", "Keep one origins line inside cors.")
				continue
			}
			cors.Origins = EnvRef{
				Name:     rowParts[2],
				Position: p.position(currentLine, 1),
			}
		case "credentials":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_CORS_CREDENTIALS", "CORS credentials must be `credentials true|false`.", "Example: `credentials true`.")
				continue
			}
			if cors.Credentials != "" {
				p.addError(currentLine, 1, "DUPLICATE_CORS_CREDENTIALS", "CORS credentials are already declared.", "Keep one credentials line inside cors.")
				continue
			}
			cors.Credentials = rowParts[1]
		default:
			p.addError(currentLine, 1, "UNEXPECTED_CORS_TOKEN", fmt.Sprintf("Unexpected cors token %q.", rowParts[0]), "Use origins or credentials inside cors.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_CORS", "CORS block is missing a closing brace.", "Add `}` after the cors block.")
	return cors, len(p.lines) - 1
}

func (p *parser) parseDeploy(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_DEPLOY_DECLARATION", "Deploy declaration must be `deploy {`.", "Start with `deploy {`, then add target, port, and env lines.")
		return start
	}
	if p.program.Deploy != nil {
		p.addError(lineNumber, 1, "DUPLICATE_DEPLOY", "Only one deploy declaration is allowed.", "Keep a single `deploy` block per project.")
		return start
	}

	deploy := DeployDecl{
		Env:      []DeployEnvDecl{},
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Deploy = &deploy
			return index
		}

		switch rowParts[0] {
		case "target":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_DEPLOY_TARGET", "Deploy target must be `target name`.", "Example: `target docker`.")
				continue
			}
			if deploy.Target != "" {
				p.addError(currentLine, 1, "DUPLICATE_DEPLOY_TARGET", "Deploy target is already declared.", "Keep one target line inside deploy.")
				continue
			}
			deploy.Target = rowParts[1]
		case "port":
			if len(rowParts) != 5 || rowParts[1] != "env" || rowParts[3] != "default" {
				p.addError(currentLine, 1, "INVALID_DEPLOY_PORT", "Deploy port must be `port env NAME default NUMBER`.", "Example: `port env PORT default 3001`.")
				continue
			}
			if deploy.Port != nil {
				p.addError(currentLine, 1, "DUPLICATE_DEPLOY_PORT", "Deploy port is already declared.", "Keep one port line inside deploy.")
				continue
			}
			deploy.Port = &DeployPortDecl{
				Env: EnvRef{
					Name:     rowParts[2],
					Position: p.position(currentLine, 1),
				},
				Default:  rowParts[4],
				Position: p.position(currentLine, 1),
			}
		case "env":
			if len(rowParts) != 3 {
				p.addError(currentLine, 1, "INVALID_DEPLOY_ENV", "Deploy env must be `env NAME required|optional`.", "Example: `env DATABASE_URL required`.")
				continue
			}
			deploy.Env = append(deploy.Env, DeployEnvDecl{
				Name:     rowParts[1],
				Mode:     rowParts[2],
				Position: p.position(currentLine, 1),
			})
		case "preview":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_DEPLOY_PREVIEW", "Deploy preview must be `preview local`.", "Example: `preview local`.")
				continue
			}
			if deploy.Preview != nil {
				p.addError(currentLine, 1, "DUPLICATE_DEPLOY_PREVIEW", "Deploy preview is already declared.", "Keep one preview line inside deploy.")
				continue
			}
			deploy.Preview = &DeployPreviewDecl{
				Mode:     rowParts[1],
				Position: p.position(currentLine, 1),
			}
		case "rollback":
			if len(rowParts) != 3 || rowParts[1] != "keep" {
				p.addError(currentLine, 1, "INVALID_DEPLOY_ROLLBACK", "Deploy rollback must be `rollback keep NUMBER`.", "Example: `rollback keep 3`.")
				continue
			}
			if deploy.Rollback != nil {
				p.addError(currentLine, 1, "DUPLICATE_DEPLOY_ROLLBACK", "Deploy rollback is already declared.", "Keep one rollback line inside deploy.")
				continue
			}
			keep, err := strconv.Atoi(rowParts[2])
			if err != nil {
				p.addError(currentLine, 1, "INVALID_DEPLOY_ROLLBACK", "Deploy rollback keep count must be an integer.", "Example: `rollback keep 3`.")
				continue
			}
			deploy.Rollback = &DeployRollbackDecl{
				Strategy: "keep",
				Keep:     keep,
				Position: p.position(currentLine, 1),
			}
		case "cloud":
			cloud, ok := p.parseDeployCloud(rowParts, currentLine)
			if !ok {
				continue
			}
			if deploy.Cloud != nil {
				p.addError(currentLine, 1, "DUPLICATE_DEPLOY_CLOUD", "Deploy cloud adapter is already declared.", "Keep one `cloud provider app env NAME` line inside deploy.")
				continue
			}
			deploy.Cloud = &cloud
		default:
			p.addError(currentLine, 1, "UNEXPECTED_DEPLOY_TOKEN", fmt.Sprintf("Unexpected deploy token %q.", rowParts[0]), "Use target, port, env, preview, rollback, or cloud inside deploy.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_DEPLOY", "Deploy block is missing a closing brace.", "Add `}` after the deploy block.")
	return len(p.lines) - 1
}

func (p *parser) parseDeployCloud(parts []string, lineNumber int) (DeployCloudDecl, bool) {
	if len(parts) != 5 && len(parts) != 8 {
		p.addError(lineNumber, 1, "INVALID_DEPLOY_CLOUD", "Deploy cloud must be `cloud PROVIDER app env NAME` with optional `region env NAME`.", "Example: `cloud fly app env FLY_APP_NAME region env FLY_REGION`.")
		return DeployCloudDecl{}, false
	}
	if parts[2] != "app" || parts[3] != "env" {
		p.addError(lineNumber, 1, "INVALID_DEPLOY_CLOUD", "Deploy cloud app must reference an environment variable.", "Example: `cloud fly app env FLY_APP_NAME`.")
		return DeployCloudDecl{}, false
	}
	cloud := DeployCloudDecl{
		Provider: parts[1],
		App: EnvRef{
			Name:     parts[4],
			Position: p.position(lineNumber, 1),
		},
		Position: p.position(lineNumber, 1),
	}
	if len(parts) == 8 {
		if parts[5] != "region" || parts[6] != "env" {
			p.addError(lineNumber, 1, "INVALID_DEPLOY_CLOUD", "Deploy cloud region must be `region env NAME`.", "Example: `cloud fly app env FLY_APP_NAME region env FLY_REGION`.")
			return DeployCloudDecl{}, false
		}
		cloud.Region = EnvRef{
			Name:     parts[7],
			Position: p.position(lineNumber, 1),
		}
	}
	return cloud, true
}

func (p *parser) parseOps(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_OPS_DECLARATION", "Ops declaration must be `ops {`.", "Start with `ops {`, then add health, readiness, metrics, or logging lines.")
		return start
	}
	if p.program.Ops != nil {
		p.addError(lineNumber, 1, "DUPLICATE_OPS", "Only one ops declaration is allowed.", "Keep a single `ops` block per project.")
		return start
	}

	ops := OpsDecl{
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Ops = &ops
			return index
		}

		switch rowParts[0] {
		case "health":
			endpoint, ok := p.parseOpsEndpoint(rowParts, currentLine, "health", "INVALID_OPS_HEALTH")
			if !ok {
				continue
			}
			if ops.Health != nil {
				p.addError(currentLine, 1, "DUPLICATE_OPS_HEALTH", "Ops health endpoint is already declared.", "Keep one `health path \"/healthz\"` line inside ops.")
				continue
			}
			ops.Health = &endpoint
		case "readiness":
			endpoint, ok := p.parseOpsEndpoint(rowParts, currentLine, "readiness", "INVALID_OPS_READINESS")
			if !ok {
				continue
			}
			if ops.Readiness != nil {
				p.addError(currentLine, 1, "DUPLICATE_OPS_READINESS", "Ops readiness endpoint is already declared.", "Keep one `readiness path \"/readyz\"` line inside ops.")
				continue
			}
			ops.Readiness = &endpoint
		case "metrics":
			endpoint, ok := p.parseOpsEndpoint(rowParts, currentLine, "metrics", "INVALID_OPS_METRICS")
			if !ok {
				continue
			}
			if ops.Metrics != nil {
				p.addError(currentLine, 1, "DUPLICATE_OPS_METRICS", "Ops metrics endpoint is already declared.", "Keep one `metrics path \"/metrics\"` line inside ops.")
				continue
			}
			ops.Metrics = &endpoint
		case "logging":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_OPS_LOGGING", "Ops logging must be `logging requests`.", "Use `logging requests` for generated request logs.")
				continue
			}
			if ops.Logging != "" {
				p.addError(currentLine, 1, "DUPLICATE_OPS_LOGGING", "Ops logging is already declared.", "Keep one logging line inside ops.")
				continue
			}
			ops.Logging = rowParts[1]
		case "observe":
			observe, ok := p.parseOpsObserve(rowParts, currentLine)
			if !ok {
				continue
			}
			if ops.Observe != nil {
				p.addError(currentLine, 1, "DUPLICATE_OPS_OBSERVE", "Ops observability hook is already declared.", "Keep one `observe PROVIDER endpoint env NAME` line inside ops.")
				continue
			}
			ops.Observe = &observe
		default:
			p.addError(currentLine, 1, "UNEXPECTED_OPS_TOKEN", fmt.Sprintf("Unexpected ops token %q.", rowParts[0]), "Use health, readiness, metrics, logging, or observe inside ops.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_OPS", "Ops block is missing a closing brace.", "Add `}` after the ops block.")
	return len(p.lines) - 1
}

func (p *parser) parseOpsEndpoint(parts []string, lineNumber int, name string, code string) (OpsEndpointDecl, bool) {
	if len(parts) != 3 || parts[1] != "path" {
		examplePath := map[string]string{"health": "/healthz", "readiness": "/readyz", "metrics": "/metrics"}[name]
		p.addError(lineNumber, 1, code, fmt.Sprintf("Ops %s endpoint must be `%s path %q`.", name, name, examplePath), fmt.Sprintf("Example: `%s path %q`.", name, examplePath))
		return OpsEndpointDecl{}, false
	}
	return OpsEndpointDecl{
		Path:     parts[2],
		Position: p.position(lineNumber, 1),
	}, true
}

func (p *parser) parseOpsObserve(parts []string, lineNumber int) (OpsObserveDecl, bool) {
	if len(parts) != 5 || parts[2] != "endpoint" || parts[3] != "env" {
		p.addError(lineNumber, 1, "INVALID_OPS_OBSERVE", "Ops observe must be `observe PROVIDER endpoint env NAME`.", "Example: `observe otlp endpoint env BLACKLANG_OTLP_ENDPOINT`.")
		return OpsObserveDecl{}, false
	}
	return OpsObserveDecl{
		Provider: parts[1],
		Endpoint: EnvRef{
			Name:     parts[4],
			Position: p.position(lineNumber, 1),
		},
		Position: p.position(lineNumber, 1),
	}, true
}

func (p *parser) parseI18N(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_I18N_DECLARATION", "I18n declaration must be `i18n {`.", "Example: `i18n { default tr locales tr, en }`.")
		return start
	}
	if p.program.I18N != nil {
		p.addError(lineNumber, 1, "DUPLICATE_I18N", "Only one i18n declaration is allowed.", "Keep a single `i18n` block per project.")
		return start
	}

	i18n := I18NDecl{
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.I18N = &i18n
			return index
		}

		switch rowParts[0] {
		case "default":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_I18N_DEFAULT", "I18n default locale must be `default locale`.", "Example: `default tr`.")
				continue
			}
			if i18n.Default != "" {
				p.addError(currentLine, 1, "DUPLICATE_I18N_DEFAULT", "I18n default locale is already declared.", "Keep one default locale inside i18n.")
				continue
			}
			i18n.Default = rowParts[1]
		case "locales":
			locales := parseList(rowParts[1:])
			if len(locales) == 0 {
				p.addError(currentLine, 1, "INVALID_I18N_LOCALES", "I18n locales must include at least one locale.", "Example: `locales tr, en`.")
				continue
			}
			if len(i18n.Locales) > 0 {
				p.addError(currentLine, 1, "DUPLICATE_I18N_LOCALES", "I18n locales are already declared.", "Keep one locales list inside i18n.")
				continue
			}
			i18n.Locales = locales
		default:
			p.addError(currentLine, 1, "UNEXPECTED_I18N_TOKEN", fmt.Sprintf("Unexpected i18n token %q.", rowParts[0]), "Use default or locales inside i18n.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_I18N", "I18n block is missing a closing brace.", "Add `}` after the i18n block.")
	return len(p.lines) - 1
}

func (p *parser) parseLabelTranslation(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_LABEL_DECLARATION", "Label translation declaration must be `label Entity.field {`.", "Example: `label Product.name { tr \"Ürün Adı\" }`.")
		return start
	}

	label := LabelTranslationDecl{
		Target:       parts[1],
		Translations: []TranslationValue{},
		Position:     p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Labels = append(p.program.Labels, label)
			return index
		}
		if len(rowParts) != 2 {
			p.addError(currentLine, 1, "INVALID_LABEL_TRANSLATION", "Label translation must be `locale \"Text\"`.", "Example: `tr \"Ürün Adı\"`.")
			continue
		}
		label.Translations = append(label.Translations, TranslationValue{
			Locale:   rowParts[0],
			Text:     rowParts[1],
			Position: p.position(currentLine, 1),
		})
	}

	p.addError(lineNumber, 1, "UNCLOSED_LABEL", fmt.Sprintf("Label translation %s is missing a closing brace.", label.Target), "Add `}` after label translations.")
	return len(p.lines) - 1
}

func (p *parser) parseFieldTextTranslation(start int, parts []string, kind string) int {
	lineNumber := p.lineNumber(start)
	codeKind := strings.ToUpper(kind)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_"+codeKind+"_DECLARATION", fmt.Sprintf("%s translation declaration must be `%s Entity.field {`.", title(kind), kind), fmt.Sprintf("Example: `%s Product.name { tr \"Ürün adını gir\" }`.", kind))
		return start
	}

	block := FieldTextTranslationDecl{
		Target:       parts[1],
		Translations: []TranslationValue{},
		Position:     p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.appendFieldTextTranslation(kind, block)
			return index
		}
		if len(rowParts) != 2 {
			p.addError(currentLine, 1, "INVALID_"+codeKind+"_TRANSLATION", fmt.Sprintf("%s translation must be `locale \"Text\"`.", title(kind)), "Example: `tr \"Ürün adını gir\"`.")
			continue
		}
		block.Translations = append(block.Translations, TranslationValue{
			Locale:   rowParts[0],
			Text:     rowParts[1],
			Position: p.position(currentLine, 1),
		})
	}

	p.addError(lineNumber, 1, "UNCLOSED_"+codeKind, fmt.Sprintf("%s translation %s is missing a closing brace.", title(kind), block.Target), fmt.Sprintf("Add `}` after %s translations.", kind))
	return len(p.lines) - 1
}

func (p *parser) appendFieldTextTranslation(kind string, block FieldTextTranslationDecl) {
	switch kind {
	case "placeholder":
		p.program.Placeholders = append(p.program.Placeholders, block)
	case "help":
		p.program.HelpTexts = append(p.program.HelpTexts, block)
	case "message":
		p.program.Messages = append(p.program.Messages, block)
	}
}

func (p *parser) parseAuthUser(start int, parts []string) (UserDecl, int) {
	lineNumber := p.lineNumber(start)
	user := UserDecl{
		Fields:   []FieldDecl{},
		Position: p.position(lineNumber, 1),
	}
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_AUTH_USER_DECLARATION", "Auth user declaration must be `user {`.", "Example: `user { email email unique }`.")
		return user, start
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		fieldParts := p.partsAt(index)
		if isClosingBrace(fieldParts) {
			return user, index
		}

		if len(fieldParts) < 2 {
			p.addError(currentLine, 1, "INVALID_AUTH_USER_FIELD", "Auth user field must include a name and type.", "Example: `email email unique`.")
			continue
		}
		fieldMetadata := p.parseModifiersAndUI(fieldParts[2:], currentLine)
		user.Fields = append(user.Fields, FieldDecl{
			Name:         fieldParts[0],
			Type:         fieldParts[1],
			Modifiers:    fieldMetadata.modifiers,
			RelationLoad: relationLoadScopesFromModifiers(fieldMetadata.modifiers),
			UI:           fieldMetadata.ui,
			Position:     p.position(currentLine, 1),
		})
	}

	p.addError(lineNumber, 1, "UNCLOSED_AUTH_USER", "Auth user block is missing a closing brace.", "Add `}` after auth user fields.")
	return user, len(p.lines) - 1
}

func (p *parser) parseEntity(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_ENTITY_DECLARATION", "Entity declaration must be `entity Name {`.", "Example: `entity Product {`.")
		return start
	}

	entity := EntityDecl{
		Name:           parts[1],
		Fields:         []FieldDecl{},
		ComputedFields: []ComputedFieldDecl{},
		Indexes:        []EntityIndexDecl{},
		Policies:       []EntityPolicyDecl{},
		Validations:    []EntityValidationDecl{},
		Position:       p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		fieldParts := p.partsAt(index)
		if isClosingBrace(fieldParts) {
			p.program.Entities = append(p.program.Entities, entity)
			return index
		}

		if fieldParts[0] == "validate" {
			validation, ok := p.parseEntityValidation(fieldParts, currentLine)
			if ok {
				entity.Validations = append(entity.Validations, validation)
			}
			continue
		}
		if fieldParts[0] == "computed" {
			computed, ok := p.parseComputedField(p.lines[index], currentLine)
			if ok {
				entity.ComputedFields = append(entity.ComputedFields, computed)
			}
			continue
		}
		if fieldParts[0] == "index" {
			indexDecl, ok := p.parseEntityIndex(fieldParts, currentLine)
			if ok {
				entity.Indexes = append(entity.Indexes, indexDecl)
			}
			continue
		}
		if fieldParts[0] == "policy" {
			policy, ok := p.parseEntityPolicy(fieldParts, currentLine)
			if ok {
				entity.Policies = append(entity.Policies, policy)
			}
			continue
		}
		if len(fieldParts) < 2 {
			p.addError(currentLine, 1, "INVALID_FIELD", "Field declaration must include a name and type.", "Example: `sku text required unique`.")
			continue
		}

		fieldMetadata := p.parseModifiersAndUI(fieldParts[2:], currentLine)
		field := FieldDecl{
			Name:         fieldParts[0],
			Type:         fieldParts[1],
			Modifiers:    fieldMetadata.modifiers,
			RelationLoad: relationLoadScopesFromModifiers(fieldMetadata.modifiers),
			UI:           fieldMetadata.ui,
			Position:     p.position(currentLine, 1),
		}
		entity.Fields = append(entity.Fields, field)
	}

	p.addError(lineNumber, 1, "UNCLOSED_ENTITY", fmt.Sprintf("Entity %s is missing a closing brace.", entity.Name), "Add `}` after the entity fields.")
	return len(p.lines) - 1
}

func (p *parser) parseComputedField(statement sourceStatement, lineNumber int) (ComputedFieldDecl, bool) {
	tokens := statement.Tokens
	if len(tokens) < 7 || !queryStatementIdentifiers(statement, 1, 2) || tokens[3].Kind != tokenOperator || tokens[3].Value != "=" {
		p.addError(lineNumber, 1, "INVALID_COMPUTED_FIELD", "Computed field must be `computed name type = expression`.", "Example: `computed inventoryValue money = stock * (price + tax)`.")
		return ComputedFieldDecl{}, false
	}

	expression, modifierStart, ok := p.parseComputedExpressionAndModifierStart(tokens, lineNumber)
	if !ok {
		return ComputedFieldDecl{}, false
	}
	left, operator, right := computedExpressionLegacyParts(expression)
	expressionCopy := expression

	return ComputedFieldDecl{
		Name: tokens[1].Value,
		Type: tokens[2].Value,
		Expression: ComputedExpressionDecl{
			Left:     left,
			Operator: operator,
			Right:    right,
			Tree:     &expressionCopy,
			Position: p.position(lineNumber, 1),
		},
		Modifiers: parseModifiers(tokensToParts(tokens[modifierStart:])),
		Position:  p.position(lineNumber, 1),
	}, true
}

func (p *parser) parseComputedExpressionAndModifierStart(tokens []sourceToken, lineNumber int) (ExpressionDecl, int, bool) {
	for index := 4; index <= len(tokens); index++ {
		if index != len(tokens) && !isComputedModifierStart(tokens[index]) && !isComputedUnknownModifierStart(tokens[index]) {
			continue
		}
		expressionTokens := tokens[4:index]
		if len(expressionTokens) == 0 {
			continue
		}
		expression, ok := parseCoreExpression(expressionTokens, parseComputedExpressionOperand)
		if !ok || !expressionHasBinary(expression) {
			continue
		}
		return expression, index, true
	}
	p.addError(lineNumber, 1, "INVALID_COMPUTED_EXPRESSION", "Computed expression must use number-like fields or numeric literals with +, -, *, /, and parentheses.", "Example: `computed inventoryValue money = stock * (price + tax)`.")
	return ExpressionDecl{}, len(tokens), false
}

func isComputedModifierStart(token sourceToken) bool {
	return token.Kind == tokenIdentifier && supportedComputedFieldModifiers[token.Value]
}

func isComputedUnknownModifierStart(token sourceToken) bool {
	return token.Kind == tokenIdentifier && expressionOperatorPrecedence(token.Value) == 0
}

func parseComputedExpressionOperand(token sourceToken) (string, string, bool) {
	if token.Kind != tokenIdentifier {
		return "", "", false
	}
	if queryNumberPattern.MatchString(token.Value) {
		return "number", token.Value, true
	}
	if queryIdentifierPattern.MatchString(token.Value) {
		return "reference", token.Value, true
	}
	return "", "", false
}

func computedExpressionLegacyParts(expression ExpressionDecl) (string, string, string) {
	if expression.Kind == "binary" {
		left := ""
		right := ""
		if expression.Left != nil {
			left = formatCoreExpression(*expression.Left)
		}
		if expression.Right != nil {
			right = formatCoreExpression(*expression.Right)
		}
		return left, expression.Operator, right
	}
	return formatCoreExpression(expression), "", ""
}

func (p *parser) parseEntityIndex(parts []string, lineNumber int) (EntityIndexDecl, bool) {
	fields := parseList(parts[1:])
	if len(fields) == 0 {
		p.addError(lineNumber, 1, "INVALID_ENTITY_INDEX", "Entity index must list at least one stored field.", "Example: `index sku` or `index customer, status`.")
		return EntityIndexDecl{}, false
	}
	return EntityIndexDecl{
		Fields:   fields,
		Position: p.position(lineNumber, 1),
	}, true
}

func (p *parser) parseEntityPolicy(parts []string, lineNumber int) (EntityPolicyDecl, bool) {
	if len(parts) != 3 || !queryIdentifierPattern.MatchString(parts[1]) || !queryIdentifierPattern.MatchString(parts[2]) {
		p.addError(lineNumber, 1, "INVALID_ENTITY_POLICY", "Entity policy must be `policy owner field` or `policy tenant field`.", "Example: `policy owner ownerId`.")
		return EntityPolicyDecl{}, false
	}
	return EntityPolicyDecl{
		Kind:     parts[1],
		Field:    parts[2],
		Position: p.position(lineNumber, 1),
	}, true
}

func (p *parser) parseEntityValidation(parts []string, lineNumber int) (EntityValidationDecl, bool) {
	if len(parts) >= 7 && parts[2] == "required" && parts[3] == "when" {
		if len(parts) != 7 && !(len(parts) == 9 && parts[7] == "message") {
			p.addError(lineNumber, 1, "INVALID_ENTITY_VALIDATION", "Conditional validation must be `validate field required when field operator value message \"Text\"`.", "Example: `validate trackingNumber required when status == shipped message \"Tracking number is required when shipped\"`.")
			return EntityValidationDecl{}, false
		}
		validation := EntityValidationDecl{
			Left:     parts[1],
			Required: true,
			When: &ValidationConditionDecl{
				Left:     parts[4],
				Operator: parts[5],
				Right:    parts[6],
				Position: p.position(lineNumber, 1),
			},
			Position: p.position(lineNumber, 1),
		}
		if len(parts) == 9 {
			validation.Message = parts[8]
		}
		return validation, true
	}
	if len(parts) != 4 && !(len(parts) == 6 && parts[4] == "message") {
		p.addError(lineNumber, 1, "INVALID_ENTITY_VALIDATION", "Entity validation must be `validate field operator field message \"Text\"` or `validate field required when field operator value message \"Text\"`.", "Example: `validate discount <= total message \"Discount cannot exceed total\"`.")
		return EntityValidationDecl{}, false
	}
	validation := EntityValidationDecl{
		Left:     parts[1],
		Operator: parts[2],
		Right:    parts[3],
		Position: p.position(lineNumber, 1),
	}
	if len(parts) == 6 {
		validation.Message = parts[5]
	}
	return validation, true
}

func (p *parser) parseAPI(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_API_DECLARATION", "API declaration must be `api Name {`.", "Example: `api LowStockProducts {`.")
		return start
	}

	api := APIDecl{
		Name:     parts[1],
		Queries:  []APIParamDecl{},
		Params:   []APIParamDecl{},
		Body:     []FieldDecl{},
		Position: p.position(lineNumber, 1),
	}
	seen := map[string]bool{}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		statement := p.lines[index]
		rowParts := statement.Parts()
		if isClosingBrace(rowParts) {
			p.program.APIs = append(p.program.APIs, api)
			return index
		}
		if len(rowParts) == 0 {
			continue
		}
		keyword := rowParts[0]
		if keyword == "method" || keyword == "path" || keyword == "public" || keyword == "private" || keyword == "webhook" || keyword == "update" || keyword == "respond" {
			seenKey := keyword
			if keyword == "public" || keyword == "private" {
				seenKey = "access"
			}
			if seen[seenKey] {
				p.addError(currentLine, 1, "DUPLICATE_API_"+strings.ToUpper(seenKey), fmt.Sprintf("API %s already declares %s.", api.Name, seenKey), "Keep one "+seenKey+" clause inside each api.")
				continue
			}
			seen[seenKey] = true
		}

		switch keyword {
		case "method":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_API_METHOD", "API method must be `method GET`.", "Use GET, POST, PUT, PATCH, or DELETE.")
				continue
			}
			api.Method = rowParts[1]
		case "path":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_API_PATH", "API path must be `path \"/api/name\"`.", "Use a quoted path, such as `path \"/api/reports/low-stock\"`.")
				continue
			}
			api.Path = rowParts[1]
		case "query":
			if len(rowParts) != 3 {
				p.addError(currentLine, 1, "INVALID_API_QUERY", "API query must be `query name type`.", "Example: `query limit integer`.")
				continue
			}
			api.Queries = append(api.Queries, APIParamDecl{Name: rowParts[1], Type: rowParts[2], Position: p.position(currentLine, 1)})
		case "param":
			if len(rowParts) != 3 {
				p.addError(currentLine, 1, "INVALID_API_PARAM", "API path parameter must be `param name type`.", "Example: `param id text`.")
				continue
			}
			api.Params = append(api.Params, APIParamDecl{Name: rowParts[1], Type: rowParts[2], Position: p.position(currentLine, 1)})
		case "body":
			field, ok := p.parseAPIBodyField(statement, currentLine)
			if ok {
				api.Body = append(api.Body, field)
			}
		case "update":
			if len(statement.Tokens) > 0 && statement.Tokens[len(statement.Tokens)-1].Kind == tokenSymbol && statement.Tokens[len(statement.Tokens)-1].Value == "{" {
				update, next, ok := p.parseAPIUpdateBlock(index, statement, currentLine)
				if ok {
					api.Update = &update
					index = next
				}
				continue
			}
			update, ok := p.parseAPIUpdate(statement, currentLine)
			if ok {
				api.Update = &update
			}
		case "respond":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_API_RESPOND", "API respond must be `respond declared`, `respond accepted`, or `respond updated`.", "Example: `respond accepted`.")
				continue
			}
			api.Respond = rowParts[1]
		case "public", "private":
			if len(rowParts) != 1 {
				p.addError(currentLine, 1, "INVALID_API_ACCESS", "API access must be `public` or `private`.", "Use one access token per line.")
				continue
			}
			api.Access = rowParts[0]
		case "webhook":
			if len(rowParts) != 1 {
				p.addError(currentLine, 1, "INVALID_API_WEBHOOK", "API webhook marker must be `webhook`.", "Use webhook on its own line.")
				continue
			}
			api.Webhook = true
		default:
			p.addError(currentLine, 1, "UNEXPECTED_API_TOKEN", fmt.Sprintf("Unexpected api token %q.", rowParts[0]), "Use method, path, query, param, body, update, respond, public, private, or webhook inside an api block.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_API", fmt.Sprintf("API %s is missing a closing brace.", api.Name), "Add `}` after the api body.")
	return len(p.lines) - 1
}

func (p *parser) parseRole(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_ROLE_DECLARATION", "Role declaration must be `role Name {`.", "Example: `role Admin {`.")
		return start
	}

	role := RoleDecl{
		Name:        parts[1],
		Permissions: []PermissionDecl{},
		Position:    p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Roles = append(p.program.Roles, role)
			return index
		}

		switch rowParts[0] {
		case "allow", "deny":
			if len(rowParts) < 2 {
				p.addError(currentLine, 1, "INVALID_PERMISSION_DECLARATION", "Permission must include an action.", "Example: `allow read Product`.")
				continue
			}
			permission := PermissionDecl{
				Effect:   rowParts[0],
				Action:   rowParts[1],
				Position: p.position(currentLine, 1),
			}
			if len(rowParts) > 2 {
				permission.Resource = rowParts[2]
			}
			if len(rowParts) > 3 {
				permission.Fields = parseList(rowParts[3:])
			}
			role.Permissions = append(role.Permissions, permission)
		default:
			p.addError(currentLine, 1, "UNEXPECTED_ROLE_TOKEN", fmt.Sprintf("Unexpected role token %q.", rowParts[0]), "Use allow or deny inside a role.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_ROLE", fmt.Sprintf("Role %s is missing a closing brace.", role.Name), "Add `}` after the role body.")
	return len(p.lines) - 1
}

func (p *parser) parseLayout(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_LAYOUT_DECLARATION", "Layout declaration must be `layout Name {`.", "Example: `layout AdminLayout {`.")
		return start
	}

	layout := LayoutDecl{
		Name:     parts[1],
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		sectionParts := p.partsAt(index)
		if isClosingBrace(sectionParts) {
			p.program.Layouts = append(p.program.Layouts, layout)
			return index
		}

		switch sectionParts[0] {
		case "sidebar":
			var sidebar SidebarDecl
			sidebar, index = p.parseSidebar(index, sectionParts)
			layout.Sidebar = sidebar
		default:
			p.addError(currentLine, 1, "UNEXPECTED_LAYOUT_TOKEN", fmt.Sprintf("Unexpected layout token %q.", sectionParts[0]), "Use sidebar inside a layout.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_LAYOUT", fmt.Sprintf("Layout %s is missing a closing brace.", layout.Name), "Add `}` after the layout body.")
	return len(p.lines) - 1
}

func (p *parser) parseSidebar(start int, parts []string) (SidebarDecl, int) {
	lineNumber := p.lineNumber(start)
	sidebar := SidebarDecl{}
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_SIDEBAR_DECLARATION", "Sidebar declaration must be `sidebar {`.", "Example: `sidebar {`.")
		return sidebar, start
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			return sidebar, index
		}

		switch rowParts[0] {
		case "item":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_SIDEBAR_ITEM", "Sidebar item must be `item PageName`.", "Example: `item Products`.")
				continue
			}
			sidebar.Items = append(sidebar.Items, rowParts[1])
		default:
			p.addError(currentLine, 1, "UNEXPECTED_SIDEBAR_TOKEN", fmt.Sprintf("Unexpected sidebar token %q.", rowParts[0]), "Use item inside a sidebar.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_SIDEBAR", "Sidebar is missing a closing brace.", "Add `}` after sidebar items.")
	return sidebar, len(p.lines) - 1
}

func (p *parser) parsePage(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_PAGE_DECLARATION", "Page declaration must be `page Name {`.", "Example: `page Products {`.")
		return start
	}

	page := PageDecl{
		Name:     parts[1],
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		sectionParts := p.partsAt(index)
		if isClosingBrace(sectionParts) {
			p.program.Pages = append(p.program.Pages, page)
			return index
		}

		switch sectionParts[0] {
		case "layout":
			if len(sectionParts) != 2 {
				p.addError(currentLine, 1, "INVALID_PAGE_LAYOUT", "Page layout must be `layout LayoutName`.", "Example: `layout AdminLayout`.")
				continue
			}
			page.Layout = sectionParts[1]
		case "source":
			if len(sectionParts) != 2 {
				p.addError(currentLine, 1, "INVALID_SOURCE", "Source must be `source EntityName`.", "Example: `source Product`.")
				continue
			}
			page.Source = sectionParts[1]
		case "query":
			if len(sectionParts) != 2 || len(p.lines[index].Tokens) != 2 || !queryStatementIdentifiers(p.lines[index], 0, 1) {
				p.addError(currentLine, 1, "INVALID_PAGE_QUERY", "Page query must be `query QueryName`.", "Example: `query LowStockProducts`.")
				continue
			}
			if page.Query != "" {
				p.addError(currentLine, 1, "DUPLICATE_PAGE_QUERY", "Page query is already declared.", "Keep one query reference inside each page.")
				continue
			}
			page.Query = sectionParts[1]
			page.QueryPosition = p.position(currentLine, 1)
		case "view":
			var view PageViewDecl
			view, index = p.parsePageView(index, sectionParts)
			if page.View != nil {
				p.addError(currentLine, 1, "DUPLICATE_VIEW", "Page view is already declared.", "Keep one `view` block inside each page.")
				continue
			}
			page.View = &view
		case "table":
			var table TableDecl
			table, index = p.parseTable(index, sectionParts)
			page.Table = table
		case "form":
			var form FormDecl
			form, index = p.parseForm(index, sectionParts)
			page.Form = form
		case "actions":
			page.Actions = parseList(sectionParts[1:])
		case "action":
			if len(sectionParts) < 4 {
				p.addError(currentLine, 1, "INVALID_ACTION_INTENT", "Action intent must be `action name ui button values...`, `action name id Value`, or `action name class Value`.", "Example: `action create ui button primary white 6 md solid`.")
				continue
			}
			position := p.position(currentLine, 1)
			actionName := sectionParts[1]
			switch sectionParts[2] {
			case "ui":
				if len(sectionParts) < 5 {
					p.addError(currentLine, 1, "INVALID_ACTION_UI", "Action UI intent must be `action name ui button values...`.", "Example: `action create ui button primary white 6 md solid`.")
					continue
				}
				ui, ok := p.parseUIIntents(sectionParts[2:], currentLine)
				if ok {
					p.mergeActionIntent(&page, actionName, ui, nil, position, currentLine)
				}
			case "id":
				if len(sectionParts) != 4 {
					p.addError(currentLine, 1, "INVALID_UI_ID", "Action id must be `action name id Identifier`.", "Example: `action create id CreateProductButton`.")
					continue
				}
				identity := &UIIdentity{ID: sectionParts[3], Position: position}
				p.mergeActionIntent(&page, actionName, nil, identity, position, currentLine)
			case "class":
				classes := parseList(sectionParts[3:])
				if len(classes) == 0 {
					p.addError(currentLine, 1, "INVALID_UI_CLASS", "Action class must be `action name class ClassName`.", "Example: `action create class primaryAction`.")
					continue
				}
				identity := &UIIdentity{Classes: classes, Position: position}
				p.mergeActionIntent(&page, actionName, nil, identity, position, currentLine)
			default:
				p.addError(currentLine, 1, "INVALID_ACTION_INTENT", "Action intent must be `action name ui button values...`, `action name id Value`, or `action name class Value`.", "Example: `action create class primaryAction`.")
			}
		case "access":
			page.Access = parseList(sectionParts[1:])
		default:
			p.addError(currentLine, 1, "UNEXPECTED_PAGE_TOKEN", fmt.Sprintf("Unexpected page token %q.", sectionParts[0]), "Use layout, source, query, view, table, form, actions, action, or access inside a page.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_PAGE", fmt.Sprintf("Page %s is missing a closing brace.", page.Name), "Add `}` after the page body.")
	return len(p.lines) - 1
}

func (p *parser) parsePageView(start int, parts []string) (PageViewDecl, int) {
	lineNumber := p.lineNumber(start)
	view := PageViewDecl{Position: p.position(lineNumber, 1)}
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_VIEW_DECLARATION", "View declaration must be `view {`.", "Example: `view {` followed by `order form, table, detail`.")
		return view, start
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			return view, index
		}

		switch rowParts[0] {
		case "order":
			if len(rowParts) < 2 {
				p.addError(currentLine, 1, "INVALID_VIEW_ORDER", "View order must list at least one section.", "Example: `order form, table, detail`.")
				continue
			}
			if len(view.Order) > 0 {
				p.addError(currentLine, 1, "DUPLICATE_VIEW_ORDER", "View order is already declared.", "Keep one order line inside view.")
				continue
			}
			view.Order = parseList(rowParts[1:])
		case "compose":
			if view.Compose != nil {
				p.addError(currentLine, 1, "DUPLICATE_VIEW_COMPOSE", "View compose is already declared.", "Keep one compose line inside view.")
				continue
			}
			compose, ok := p.parseViewCompose(rowParts, currentLine)
			if ok {
				view.Compose = &compose
			}
		case "section":
			section, ok := p.parseViewSection(rowParts, currentLine)
			if ok {
				view.Sections = append(view.Sections, section)
			}
		case "tab":
			tab, ok := p.parseViewTab(rowParts, currentLine)
			if ok {
				view.Tabs = append(view.Tabs, tab)
			}
		case "group":
			group, ok := p.parseViewGroup(rowParts, currentLine)
			if ok {
				view.Groups = append(view.Groups, group)
			}
		case "trigger":
			trigger, ok := p.parseViewTrigger(rowParts, currentLine)
			if ok {
				view.Triggers = append(view.Triggers, trigger)
			}
		default:
			p.addError(currentLine, 1, "UNEXPECTED_VIEW_TOKEN", fmt.Sprintf("Unexpected view token %q.", rowParts[0]), "Use order, compose, section, tab, group, or trigger inside a view block.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_VIEW", "View is missing a closing brace.", "Add `}` after view settings.")
	return view, len(p.lines) - 1
}

func (p *parser) parseViewCompose(parts []string, lineNumber int) (ViewComposeDecl, bool) {
	compose := ViewComposeDecl{Position: p.position(lineNumber, 1)}
	if len(parts) < 2 {
		p.addError(lineNumber, 1, "INVALID_VIEW_COMPOSE", "View compose must name a mode.", "Example: `compose grid columns 2 gap md stackAt md`.")
		return compose, false
	}
	compose.Mode = parts[1]
	for index := 2; index < len(parts); index += 2 {
		if index+1 >= len(parts) {
			p.addError(lineNumber, 1, "INVALID_VIEW_COMPOSE", "View compose options must use key value pairs.", "Example: `compose grid columns 2 gap md stackAt md`.")
			return compose, false
		}
		key := parts[index]
		value := parts[index+1]
		switch key {
		case "columns":
			if compose.Columns != 0 {
				p.addError(lineNumber, 1, "DUPLICATE_VIEW_COMPOSE_OPTION", "View compose columns option is already declared.", "Keep one columns option.")
				return compose, false
			}
			columns, err := strconv.Atoi(value)
			if err != nil || columns < 1 {
				p.addError(lineNumber, 1, "INVALID_VIEW_COMPOSE_COLUMNS", "View compose columns must be a number.", "Use columns 1, 2, 3, or 4.")
				return compose, false
			}
			compose.Columns = columns
		case "gap":
			if compose.Gap != "" {
				p.addError(lineNumber, 1, "DUPLICATE_VIEW_COMPOSE_OPTION", "View compose gap option is already declared.", "Keep one gap option.")
				return compose, false
			}
			compose.Gap = value
		case "stackAt":
			if compose.StackAt != "" {
				p.addError(lineNumber, 1, "DUPLICATE_VIEW_COMPOSE_OPTION", "View compose stackAt option is already declared.", "Keep one stackAt option.")
				return compose, false
			}
			compose.StackAt = value
		default:
			p.addError(lineNumber, 1, "INVALID_VIEW_COMPOSE_OPTION", fmt.Sprintf("Unknown view compose option %q.", key), "Use columns, gap, or stackAt.")
			return compose, false
		}
	}
	return compose, true
}

func (p *parser) parseViewSection(parts []string, lineNumber int) (ViewSectionDecl, bool) {
	section := ViewSectionDecl{Position: p.position(lineNumber, 1)}
	if len(parts) < 2 {
		p.addError(lineNumber, 1, "INVALID_VIEW_SECTION", "View section must be `section <name> [component <Component>] [bind selected|first|each] [span <columns>] [display inline|modal|drawer]`.", "Example: `section StockSummary component StockBadge bind selected title \"Stock Summary\"`.")
		return section, false
	}
	section.Name = parts[1]
	for index := 2; index < len(parts); index += 2 {
		if index+1 >= len(parts) {
			p.addError(lineNumber, 1, "INVALID_VIEW_SECTION", "View section options must use key value pairs.", "Use component, bind, span, display, side, or title options.")
			return section, false
		}
		key := parts[index]
		value := parts[index+1]
		switch key {
		case "component":
			if section.Component != "" {
				p.addError(lineNumber, 1, "DUPLICATE_VIEW_SECTION_OPTION", "View section component option is already declared.", "Keep one component option.")
				return section, false
			}
			section.Component = value
		case "bind":
			if section.Bind != "" {
				p.addError(lineNumber, 1, "DUPLICATE_VIEW_SECTION_OPTION", "View section bind option is already declared.", "Keep one bind option.")
				return section, false
			}
			section.Bind = value
		case "span":
			if section.Span != 0 {
				p.addError(lineNumber, 1, "DUPLICATE_VIEW_SECTION_OPTION", "View section span option is already declared.", "Keep one span option.")
				return section, false
			}
			span, err := strconv.Atoi(value)
			if err != nil {
				p.addError(lineNumber, 1, "INVALID_VIEW_SECTION_SPAN", "View section span must be a number.", "Use span 1, 2, 3, or 4.")
				return section, false
			}
			section.Span = span
		case "display":
			if section.Display != "" {
				p.addError(lineNumber, 1, "DUPLICATE_VIEW_SECTION_OPTION", "View section display option is already declared.", "Keep one display option.")
				return section, false
			}
			section.Display = value
		case "side":
			if section.Side != "" {
				p.addError(lineNumber, 1, "DUPLICATE_VIEW_SECTION_OPTION", "View section side option is already declared.", "Keep one side option.")
				return section, false
			}
			section.Side = value
		case "title":
			if section.Title != "" {
				p.addError(lineNumber, 1, "DUPLICATE_VIEW_SECTION_OPTION", "View section title option is already declared.", "Keep one title option.")
				return section, false
			}
			section.Title = value
		default:
			p.addError(lineNumber, 1, "INVALID_VIEW_SECTION_OPTION", fmt.Sprintf("Unknown view section option %q.", key), "Use component, bind, span, display, side, or title.")
			return section, false
		}
	}
	return section, true
}

func (p *parser) parseViewTab(parts []string, lineNumber int) (ViewTabDecl, bool) {
	tab := ViewTabDecl{Position: p.position(lineNumber, 1)}
	if len(parts) < 4 || parts[2] != "sections" || !isThemeIdentifier(parts[1]) {
		p.addError(lineNumber, 1, "INVALID_VIEW_TAB", "View tab must be `tab <Name> sections <section...>`.", "Example: `tab Overview sections table, detail`.")
		return tab, false
	}
	sections := parseList(parts[3:])
	if len(sections) == 0 {
		p.addError(lineNumber, 1, "INVALID_VIEW_TAB", "View tab must list at least one section.", "Example: `tab Details sections detail, form`.")
		return tab, false
	}
	tab.Name = parts[1]
	tab.Sections = sections
	return tab, true
}

func (p *parser) parseViewGroup(parts []string, lineNumber int) (ViewGroupDecl, bool) {
	group := ViewGroupDecl{Position: p.position(lineNumber, 1)}
	if len(parts) < 4 || parts[2] != "sections" || !isThemeIdentifier(parts[1]) {
		p.addError(lineNumber, 1, "INVALID_VIEW_GROUP", "View group must be `group <Name> sections <section...> [compose stack|grid]`.", "Example: `group Record sections detail, form compose stack title \"Record\"`.")
		return group, false
	}
	group.Name = parts[1]

	sectionEnd := 3
	for sectionEnd < len(parts) && !isViewGroupOption(parts[sectionEnd]) {
		sectionEnd++
	}
	if sectionEnd == 3 {
		p.addError(lineNumber, 1, "INVALID_VIEW_GROUP", "View group must list at least one section.", "Example: `group Record sections detail, form`.")
		return group, false
	}
	group.Sections = parseList(parts[3:sectionEnd])
	if len(group.Sections) == 0 {
		p.addError(lineNumber, 1, "INVALID_VIEW_GROUP", "View group must list at least one section.", "Example: `group Record sections detail, form`.")
		return group, false
	}

	options := parts[sectionEnd:]
	if len(options)%2 != 0 {
		p.addError(lineNumber, 1, "INVALID_VIEW_GROUP_OPTION", "View group options must use key value pairs.", "Use compose, columns, gap, span, or title options.")
		return group, false
	}

	seen := map[string]bool{}
	for index := 0; index < len(options); index += 2 {
		key := options[index]
		value := options[index+1]
		if seen[key] {
			p.addError(lineNumber, 1, "DUPLICATE_VIEW_GROUP_OPTION", fmt.Sprintf("View group option %q is already declared.", key), "Keep one value for each group option.")
			return group, false
		}
		seen[key] = true
		switch key {
		case "compose":
			if group.Compose == nil {
				group.Compose = &ViewComposeDecl{Position: group.Position}
			}
			group.Compose.Mode = value
		case "columns":
			if group.Compose == nil {
				group.Compose = &ViewComposeDecl{Position: group.Position}
			}
			columns, err := strconv.Atoi(value)
			if err != nil || columns < 1 {
				p.addError(lineNumber, 1, "INVALID_VIEW_GROUP_COLUMNS", "View group columns must be a number.", "Use columns 1, 2, 3, or 4.")
				return group, false
			}
			group.Compose.Columns = columns
		case "gap":
			if group.Compose == nil {
				group.Compose = &ViewComposeDecl{Position: group.Position}
			}
			group.Compose.Gap = value
		case "span":
			span, err := strconv.Atoi(value)
			if err != nil || span < 1 {
				p.addError(lineNumber, 1, "INVALID_VIEW_GROUP_SPAN", "View group span must be a number.", "Use span 1, 2, 3, or 4.")
				return group, false
			}
			group.Span = span
		case "title":
			group.Title = value
		default:
			p.addError(lineNumber, 1, "INVALID_VIEW_GROUP_OPTION", fmt.Sprintf("Unknown view group option %q.", key), "Use compose, columns, gap, span, or title.")
			return group, false
		}
	}
	return group, true
}

func isViewGroupOption(value string) bool {
	switch value {
	case "compose", "columns", "gap", "span", "title":
		return true
	default:
		return false
	}
}

func (p *parser) parseViewTrigger(parts []string, lineNumber int) (ViewTriggerDecl, bool) {
	trigger := ViewTriggerDecl{Position: p.position(lineNumber, 1)}
	if len(parts) != 4 || parts[2] != "on" {
		p.addError(lineNumber, 1, "INVALID_VIEW_TRIGGER", "View trigger must be `trigger <section> on <event>`.", "Example: `trigger detail on rowSelect`.")
		return trigger, false
	}
	trigger.Section = parts[1]
	trigger.Event = parts[3]
	return trigger, true
}

func (p *parser) parseTable(start int, parts []string) (TableDecl, int) {
	lineNumber := p.lineNumber(start)
	table := TableDecl{}
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_TABLE_DECLARATION", "Table declaration must be `table {`.", "Example: `table {`.")
		return table, start
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			return table, index
		}

		switch rowParts[0] {
		case "columns":
			table.Columns = parseList(rowParts[1:])
		case "search":
			table.Search = parseList(rowParts[1:])
		case "filter":
			table.Filters = parseList(rowParts[1:])
		case "id":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_UI_ID", "Table id must be `id Identifier`.", "Example: `id ProductsTable`.")
				continue
			}
			table.Identity = p.setUIIdentityID(table.Identity, rowParts[1], p.position(currentLine, 1), currentLine)
		case "class":
			classes := parseList(rowParts[1:])
			if len(classes) == 0 {
				p.addError(currentLine, 1, "INVALID_UI_CLASS", "Table class must include at least one class name.", "Example: `class importantTable compactPanel`.")
				continue
			}
			table.Identity = addUIIdentityClasses(table.Identity, classes, p.position(currentLine, 1))
		case "ui":
			ui, ok := p.parseUIIntents(rowParts, currentLine)
			if ok {
				table.UI = append(table.UI, ui...)
			}
		case "sort":
			if len(rowParts) != 3 {
				p.addError(currentLine, 1, "INVALID_TABLE_SORT", "Table sort must be `sort field asc` or `sort field desc`.", "Example: `sort name asc`.")
				continue
			}
			table.Sort = SortDecl{Field: rowParts[1], Direction: rowParts[2]}
		case "paginate":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_TABLE_PAGINATION", "Table pagination must be `paginate number`.", "Example: `paginate 25`.")
				continue
			}
			size, err := strconv.Atoi(rowParts[1])
			if err != nil {
				p.addError(currentLine, 1, "INVALID_TABLE_PAGINATION", "Table pagination must use a whole number.", "Example: `paginate 25`.")
				continue
			}
			if size <= 0 {
				p.addError(currentLine, 1, "INVALID_TABLE_PAGINATION", "Table pagination must use a positive whole number.", "Example: `paginate 25`.")
				continue
			}
			table.Paginate = size
		default:
			p.addError(currentLine, 1, "UNEXPECTED_TABLE_TOKEN", fmt.Sprintf("Unexpected table token %q.", rowParts[0]), "Use columns, search, filter, sort, paginate, id, class, or ui inside a table.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_TABLE", "Table is missing a closing brace.", "Add `}` after table settings.")
	return table, len(p.lines) - 1
}

func (p *parser) parseForm(start int, parts []string) (FormDecl, int) {
	lineNumber := p.lineNumber(start)
	form := FormDecl{}
	if len(parts) != 2 || parts[1] != "{" {
		p.addError(lineNumber, 1, "INVALID_FORM_DECLARATION", "Form declaration must be `form {`.", "Example: `form {`.")
		return form, start
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			return form, index
		}

		switch rowParts[0] {
		case "fields":
			form.Fields = parseList(rowParts[1:])
		case "id":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_UI_ID", "Form id must be `id Identifier`.", "Example: `id ProductForm`.")
				continue
			}
			form.Identity = p.setUIIdentityID(form.Identity, rowParts[1], p.position(currentLine, 1), currentLine)
		case "class":
			classes := parseList(rowParts[1:])
			if len(classes) == 0 {
				p.addError(currentLine, 1, "INVALID_UI_CLASS", "Form class must include at least one class name.", "Example: `class productForm elevatedPanel`.")
				continue
			}
			form.Identity = addUIIdentityClasses(form.Identity, classes, p.position(currentLine, 1))
		case "ui":
			ui, ok := p.parseUIIntents(rowParts, currentLine)
			if ok {
				form.UI = append(form.UI, ui...)
			}
		default:
			p.addError(currentLine, 1, "UNEXPECTED_FORM_TOKEN", fmt.Sprintf("Unexpected form token %q.", rowParts[0]), "Use fields, id, class, or ui inside a form.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_FORM", "Form is missing a closing brace.", "Add `}` after form settings.")
	return form, len(p.lines) - 1
}

func (p *parser) setUIIdentityID(identity *UIIdentity, id string, position Position, lineNumber int) *UIIdentity {
	if identity == nil {
		return &UIIdentity{ID: id, Position: position}
	}
	if identity.ID != "" {
		p.addError(lineNumber, 1, "DUPLICATE_UI_ID", fmt.Sprintf("UI id %s was already declared.", identity.ID), "Keep one id per generated UI element.")
	}
	identity.ID = id
	if identity.Position.Line == 0 {
		identity.Position = position
	}
	return identity
}

func addUIIdentityClasses(identity *UIIdentity, classes []string, position Position) *UIIdentity {
	if identity == nil {
		return &UIIdentity{Classes: classes, Position: position}
	}
	identity.Classes = append(identity.Classes, classes...)
	if identity.Position.Line == 0 {
		identity.Position = position
	}
	return identity
}

func (p *parser) mergeActionIntent(page *PageDecl, action string, ui []UIIntent, identity *UIIdentity, position Position, lineNumber int) {
	for index := range page.ActionUI {
		if page.ActionUI[index].Action != action {
			continue
		}
		page.ActionUI[index].UI = append(page.ActionUI[index].UI, ui...)
		if identity != nil {
			page.ActionUI[index].Identity = p.mergeUIIdentity(page.ActionUI[index].Identity, identity, lineNumber)
		}
		return
	}

	intent := ActionUIIntent{
		Action:   action,
		UI:       ui,
		Position: position,
	}
	if identity != nil {
		intent.Identity = p.mergeUIIdentity(nil, identity, lineNumber)
	}
	page.ActionUI = append(page.ActionUI, intent)
}

func (p *parser) mergeUIIdentity(current *UIIdentity, update *UIIdentity, lineNumber int) *UIIdentity {
	if update == nil {
		return current
	}
	if current == nil {
		return &UIIdentity{
			ID:       update.ID,
			Classes:  append([]string{}, update.Classes...),
			Position: update.Position,
		}
	}
	if update.ID != "" {
		current = p.setUIIdentityID(current, update.ID, update.Position, lineNumber)
	}
	if len(update.Classes) > 0 {
		current = addUIIdentityClasses(current, update.Classes, update.Position)
	}
	return current
}

func (p *parser) parseWorkflow(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_WORKFLOW_DECLARATION", "Workflow declaration must be `workflow Name {`.", "Example: `workflow OrderPreparation {`.")
		return start
	}

	workflow := WorkflowDecl{
		Name:        parts[1],
		Transitions: []TransitionDecl{},
		Position:    p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Workflows = append(p.program.Workflows, workflow)
			return index
		}

		switch rowParts[0] {
		case "source":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_WORKFLOW_SOURCE", "Workflow source must be `source EntityName`.", "Example: `source Order`.")
				continue
			}
			workflow.Source = rowParts[1]
		case "states":
			workflow.States = parseList(rowParts[1:])
		case "transition":
			var transition TransitionDecl
			transition, index = p.parseTransition(index, rowParts)
			workflow.Transitions = append(workflow.Transitions, transition)
		default:
			p.addError(currentLine, 1, "UNEXPECTED_WORKFLOW_TOKEN", fmt.Sprintf("Unexpected workflow token %q.", rowParts[0]), "Use source, states, or transition inside a workflow.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_WORKFLOW", fmt.Sprintf("Workflow %s is missing a closing brace.", workflow.Name), "Add `}` after the workflow body.")
	return len(p.lines) - 1
}

func (p *parser) parseTransition(start int, parts []string) (TransitionDecl, int) {
	lineNumber := p.lineNumber(start)
	transition := TransitionDecl{
		Position: p.position(lineNumber, 1),
	}
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_TRANSITION_DECLARATION", "Transition declaration must be `transition Name {`.", "Example: `transition ship {`.")
		return transition, start
	}
	transition.Name = parts[1]

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			return transition, index
		}

		switch rowParts[0] {
		case "from":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_TRANSITION_FROM", "Transition from must be `from StateName`.", "Example: `from draft`.")
				continue
			}
			transition.From = rowParts[1]
		case "to":
			if len(rowParts) != 2 {
				p.addError(currentLine, 1, "INVALID_TRANSITION_TO", "Transition to must be `to StateName`.", "Example: `to shipped`.")
				continue
			}
			transition.To = rowParts[1]
		case "allow":
			transition.Allow = parseList(rowParts[1:])
		default:
			p.addError(currentLine, 1, "UNEXPECTED_TRANSITION_TOKEN", fmt.Sprintf("Unexpected transition token %q.", rowParts[0]), "Use from, to, or allow inside a transition.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_TRANSITION", fmt.Sprintf("Transition %s is missing a closing brace.", transition.Name), "Add `}` after the transition body.")
	return transition, len(p.lines) - 1
}

func (p *parser) parseState(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_STATE_DECLARATION", "State declaration must be `state Name {`.", "Example: `state ProductPageState {`.")
		return start
	}

	state := StateDecl{
		Name:     parts[1],
		Fields:   []StateField{},
		Modals:   []StateModal{},
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.States = append(p.program.States, state)
			return index
		}

		if rowParts[0] == "modal" {
			if len(rowParts) != 3 {
				p.addError(currentLine, 1, "INVALID_STATE_MODAL", "State modal must be `modal name open|closed`.", "Example: `modal createProduct closed`.")
				continue
			}
			state.Modals = append(state.Modals, StateModal{
				Name:     rowParts[1],
				Default:  rowParts[2],
				Position: p.position(currentLine, 1),
			})
			continue
		}
		if len(rowParts) != 2 {
			p.addError(currentLine, 1, "INVALID_STATE_FIELD", "State field must include a name and type.", "Example: `activeFilter text`.")
			continue
		}
		fieldType := rowParts[1]
		list := strings.HasSuffix(fieldType, "[]")
		if list {
			fieldType = strings.TrimSuffix(fieldType, "[]")
		}
		state.Fields = append(state.Fields, StateField{
			Name:     rowParts[0],
			Type:     fieldType,
			List:     list,
			Position: p.position(currentLine, 1),
		})
	}

	p.addError(lineNumber, 1, "UNCLOSED_STATE", fmt.Sprintf("State %s is missing a closing brace.", state.Name), "Add `}` after the state body.")
	return len(p.lines) - 1
}

func (p *parser) parseComponent(start int, parts []string) int {
	lineNumber := p.lineNumber(start)
	if len(parts) != 3 || parts[2] != "{" {
		p.addError(lineNumber, 1, "INVALID_COMPONENT_DECLARATION", "Component declaration must be `component Name {`.", "Example: `component StockBadge {`.")
		return start
	}

	component := ComponentDecl{
		Name:     parts[1],
		Inputs:   []ComponentInput{},
		Variants: []ComponentVariant{},
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.Components = append(p.program.Components, component)
			return index
		}

		switch rowParts[0] {
		case "input":
			if len(rowParts) != 3 {
				p.addError(currentLine, 1, "INVALID_COMPONENT_INPUT", "Component input must be `input name type`.", "Example: `input stock number`.")
				continue
			}
			inputType := rowParts[2]
			list := strings.HasSuffix(inputType, "[]")
			if list {
				inputType = strings.TrimSuffix(inputType, "[]")
			}
			component.Inputs = append(component.Inputs, ComponentInput{
				Name:     rowParts[1],
				Type:     inputType,
				List:     list,
				Position: p.position(currentLine, 1),
			})
		case "variant":
			if len(rowParts) < 4 || rowParts[2] != "when" {
				p.addError(currentLine, 1, "INVALID_COMPONENT_VARIANT", "Component variant must be `variant name when condition`.", "Example: `variant low when stock < 10`.")
				continue
			}
			component.Variants = append(component.Variants, ComponentVariant{
				Name:      rowParts[1],
				Condition: strings.Join(rowParts[3:], " "),
				Position:  p.position(currentLine, 1),
			})
		default:
			p.addError(currentLine, 1, "UNEXPECTED_COMPONENT_TOKEN", fmt.Sprintf("Unexpected component token %q.", rowParts[0]), "Use input or variant inside a component.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_COMPONENT", fmt.Sprintf("Component %s is missing a closing brace.", component.Name), "Add `}` after the component body.")
	return len(p.lines) - 1
}

func parseModifiers(parts []string) []Modifier {
	modifiers := []Modifier{}
	for index := 0; index < len(parts); index++ {
		name := parts[index]
		modifier := Modifier{Name: name}
		if modifierTakesList(name) {
			modifier.Value, index = parseListModifierValue(parts, index)
			modifiers = append(modifiers, modifier)
			continue
		}
		if modifierTakesValue(name) && index+1 < len(parts) {
			modifier.Value = parts[index+1]
			index++
		}
		modifiers = append(modifiers, modifier)
	}
	return modifiers
}

type fieldMetadata struct {
	modifiers []Modifier
	ui        []UIIntent
}

func (p *parser) parseModifiersAndUI(parts []string, lineNumber int) fieldMetadata {
	metadata := fieldMetadata{
		modifiers: []Modifier{},
		ui:        []UIIntent{},
	}
	for index := 0; index < len(parts); index++ {
		name := parts[index]
		if name == "ui" {
			ui, ok := p.parseUIIntents(parts[index:], lineNumber)
			if ok {
				metadata.ui = ui
			}
			return metadata
		}
		modifier := Modifier{Name: name}
		if modifierTakesList(name) {
			modifier.Value, index = parseListModifierValue(parts, index)
			metadata.modifiers = append(metadata.modifiers, modifier)
			continue
		}
		if modifierTakesValue(name) && index+1 < len(parts) {
			modifier.Value = parts[index+1]
			index++
		}
		metadata.modifiers = append(metadata.modifiers, modifier)
	}
	return metadata
}

func (p *parser) parseUIIntents(parts []string, lineNumber int) ([]UIIntent, bool) {
	if len(parts) < 3 || parts[0] != "ui" {
		p.addError(lineNumber, 1, "INVALID_UI_INTENT", "UI intent must be `ui mode values...`.", "Example: `ui box black 1 solid 8 8 5 5 6 center`.")
		return nil, false
	}
	intents := []UIIntent{}
	index := 1
	for index < len(parts) {
		mode := parts[index]
		if mode == "|" {
			p.addError(lineNumber, 1, "INVALID_UI_INTENT", "UI intent mode is missing before `|`.", "Write `ui box values... | text values...`.")
			return nil, false
		}
		index++
		values := []string{}
		for index < len(parts) && parts[index] != "|" {
			values = append(values, parts[index])
			index++
		}
		if len(values) == 0 {
			p.addError(lineNumber, 1, "INVALID_UI_INTENT", fmt.Sprintf("UI intent mode %s has no values.", mode), fmt.Sprintf("Add compact values after `%s`, or remove the UI intent.", mode))
			return nil, false
		}
		intents = append(intents, UIIntent{
			Mode:     mode,
			Values:   values,
			Position: p.position(lineNumber, 1),
		})
		if index < len(parts) && parts[index] == "|" {
			index++
			if index >= len(parts) {
				p.addError(lineNumber, 1, "INVALID_UI_INTENT", "UI intent cannot end with `|`.", "Add another `mode values...` segment after `|`, or remove the trailing separator.")
				return nil, false
			}
		}
	}
	return intents, true
}

func modifierTakesValue(name string) bool {
	return name == "default" || name == "label" || name == "placeholder" || name == "help" || name == "min" || name == "max" || name == "length" || name == "regex" || name == "accept" || name == "message"
}

func parseList(parts []string) []string {
	items := []string{}
	for _, part := range parts {
		item := strings.TrimSpace(strings.TrimSuffix(part, ","))
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func (p *parser) position(line int, column int) Position {
	return Position{
		File:   p.file,
		Line:   line,
		Column: column,
	}
}

func (p *parser) partsAt(index int) []string {
	return p.lines[index].Parts()
}

func (p *parser) lineNumber(index int) int {
	return p.lines[index].Position.Line
}

func isClosingBrace(parts []string) bool {
	return len(parts) == 1 && parts[0] == "}"
}

func (p *parser) addError(line int, column int, code string, message string, suggestion string) {
	p.diagnostics = append(p.diagnostics, Diagnostic{
		File:       p.file,
		Line:       line,
		Column:     column,
		Code:       code,
		Message:    message,
		Suggestion: suggestion,
	})
}
