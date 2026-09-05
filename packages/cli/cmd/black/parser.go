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
		case "auth":
			index = p.parseAuth(index, parts)
		case "database":
			index = p.parseDatabase(index, parts)
		case "entity":
			index = p.parseEntity(index, parts)
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
			p.addError(lineNumber, 1, "UNEXPECTED_TOP_LEVEL", fmt.Sprintf("Unexpected top-level token %q.", parts[0]), "Use app, auth, database, entity, role, api, layout, page, workflow, state, or component at the top level.")
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
			Name:      fieldParts[0],
			Type:      fieldParts[1],
			Modifiers: fieldMetadata.modifiers,
			UI:        fieldMetadata.ui,
			Position:  p.position(currentLine, 1),
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
		Name:        parts[1],
		Fields:      []FieldDecl{},
		Validations: []EntityValidationDecl{},
		Position:    p.position(lineNumber, 1),
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
		if len(fieldParts) < 2 {
			p.addError(currentLine, 1, "INVALID_FIELD", "Field declaration must include a name and type.", "Example: `sku text required unique`.")
			continue
		}

		fieldMetadata := p.parseModifiersAndUI(fieldParts[2:], currentLine)
		field := FieldDecl{
			Name:      fieldParts[0],
			Type:      fieldParts[1],
			Modifiers: fieldMetadata.modifiers,
			UI:        fieldMetadata.ui,
			Position:  p.position(currentLine, 1),
		}
		entity.Fields = append(entity.Fields, field)
	}

	p.addError(lineNumber, 1, "UNCLOSED_ENTITY", fmt.Sprintf("Entity %s is missing a closing brace.", entity.Name), "Add `}` after the entity fields.")
	return len(p.lines) - 1
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
		Position: p.position(lineNumber, 1),
	}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		rowParts := p.partsAt(index)
		if isClosingBrace(rowParts) {
			p.program.APIs = append(p.program.APIs, api)
			return index
		}

		switch rowParts[0] {
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
			p.addError(currentLine, 1, "UNEXPECTED_API_TOKEN", fmt.Sprintf("Unexpected api token %q.", rowParts[0]), "Use method, path, query, param, public, private, or webhook inside an api block.")
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
			p.addError(currentLine, 1, "UNEXPECTED_PAGE_TOKEN", fmt.Sprintf("Unexpected page token %q.", sectionParts[0]), "Use layout, source, table, form, actions, action, or access inside a page.")
		}
	}

	p.addError(lineNumber, 1, "UNCLOSED_PAGE", fmt.Sprintf("Page %s is missing a closing brace.", page.Name), "Add `}` after the page body.")
	return len(p.lines) - 1
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
	return name == "default" || name == "label" || name == "placeholder" || name == "help" || name == "min" || name == "max" || name == "length" || name == "regex" || name == "message"
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
