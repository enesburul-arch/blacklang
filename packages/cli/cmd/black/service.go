package main

import (
	"fmt"
	"strings"
)

type ServiceDecl struct {
	Name     string              `json:"name"`
	APIs     []ServiceTargetDecl `json:"apis,omitempty"`
	Position Position            `json:"position"`
}

type ServiceTargetDecl struct {
	Name     string   `json:"name"`
	Position Position `json:"position"`
}

func (p *parser) parseService(start int, parts []string) int {
	line := p.lineNumber(start)
	statement := p.lines[start]
	if len(parts) != 3 || len(statement.Tokens) != 3 || !queryStatementIdentifiers(statement, 0, 1) || statement.Tokens[2].Kind != tokenSymbol || parts[2] != "{" {
		p.addError(line, 1, "INVALID_SERVICE_DECLARATION", "Service declaration must be `service Name {`.", "Example: `service InventoryIntegration {`.")
		return start
	}

	service := ServiceDecl{
		Name:     parts[1],
		APIs:     []ServiceTargetDecl{},
		Position: p.position(line, 1),
	}
	seenAPIs := map[string]Position{}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		statement := p.lines[index]
		tokens := statement.Tokens
		rowParts := statement.Parts()
		if isClosingBrace(rowParts) {
			p.program.Services = append(p.program.Services, service)
			return index
		}
		if len(tokens) == 0 {
			continue
		}
		if tokens[0].Kind != tokenIdentifier {
			p.addError(currentLine, 1, "UNEXPECTED_SERVICE_TOKEN", "Service clauses must begin with api.", "Use one `api APIName` target per line.")
			continue
		}

		if tokens[0].Value != "api" {
			p.addError(currentLine, 1, "UNEXPECTED_SERVICE_TOKEN", fmt.Sprintf("Unexpected service token %q.", tokens[0].Value), "Use api inside a service block.")
			continue
		}
		if len(tokens) != 2 || !queryStatementIdentifiers(statement, 1) {
			p.addError(currentLine, 1, "INVALID_SERVICE_API", "Service API targets must be `api Name`.", "Example: `api StockWebhook`.")
			continue
		}

		target := ServiceTargetDecl{Name: tokens[1].Value, Position: tokens[1].Position}
		if existing, ok := seenAPIs[target.Name]; ok {
			p.addError(currentLine, 1, "DUPLICATE_SERVICE_TARGET", fmt.Sprintf("Service %s already targets api %s.", service.Name, target.Name), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
			continue
		}
		seenAPIs[target.Name] = target.Position
		service.APIs = append(service.APIs, target)
	}

	p.addError(line, 1, "UNCLOSED_SERVICE", fmt.Sprintf("Service %s is missing a closing brace.", service.Name), "Add `}` after the service body.")
	return len(p.lines) - 1
}

func (v *semanticValidator) validateServices(apis map[string]APIDecl) map[string]ServiceDecl {
	services := map[string]ServiceDecl{}
	normalizedNames := map[string]string{}
	boundAPIs := map[string]ServiceDecl{}
	symbols := serviceOtherSymbols(v.program)

	for _, service := range v.program.Services {
		if existing, ok := services[service.Name]; ok {
			v.addDiagnostic(service.Position, "DUPLICATE_SERVICE", fmt.Sprintf("Service %s is already defined.", service.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		services[service.Name] = service

		if !queryNamePattern.MatchString(service.Name) {
			v.addDiagnostic(service.Position, "INVALID_SERVICE_NAME", fmt.Sprintf("Service name %q must use PascalCase letters and digits.", service.Name), "Use a name such as InventoryIntegration.")
		}
		normalized := strings.ToLower(service.Name)
		if existing, ok := normalizedNames[normalized]; ok {
			v.addDiagnostic(service.Position, "SERVICE_NAME_COLLISION", fmt.Sprintf("Service %s conflicts with service %s after name normalization.", service.Name, existing), "Choose distinct service names, including after lowercasing.")
		}
		normalizedNames[normalized] = service.Name
		if symbols[service.Name] || symbols[normalized] {
			v.addDiagnostic(service.Position, "SERVICE_NAME_COLLISION", fmt.Sprintf("Service %s conflicts with another application symbol.", service.Name), "Choose a unique service name for unambiguous inspect --affected output.")
		}

		if len(service.APIs) == 0 {
			v.addDiagnostic(service.Position, "EMPTY_SERVICE", fmt.Sprintf("Service %s does not target any api.", service.Name), "Add at least one `api Name` line.")
		}

		for _, target := range service.APIs {
			if !queryNamePattern.MatchString(target.Name) {
				v.addDiagnostic(target.Position, "INVALID_SERVICE_API", fmt.Sprintf("Service %s api target %q must use PascalCase.", service.Name, target.Name), "Reference a declared api such as StockWebhook.")
				continue
			}
			if _, ok := apis[target.Name]; !ok {
				v.addDiagnostic(target.Position, "UNKNOWN_SERVICE_API", fmt.Sprintf("Service %s references unknown api %s.", service.Name, target.Name), "Declare the api or remove the service target.")
				continue
			}
			if existing, ok := boundAPIs[target.Name]; ok {
				v.addDiagnostic(target.Position, "DUPLICATE_SERVICE_BINDING", fmt.Sprintf("API %s is already bound to service %s.", target.Name, existing.Name), "Bind each explicit api to at most one service block.")
				continue
			}
			boundAPIs[target.Name] = service
		}
	}

	return services
}

func serviceOtherSymbols(program Program) map[string]bool {
	symbols := map[string]bool{}
	add := func(value string) {
		if value == "" {
			return
		}
		symbols[value] = true
		symbols[strings.ToLower(value)] = true
	}
	add(program.App.Name)
	for _, entity := range program.Entities {
		add(entity.Name)
	}
	for _, query := range program.Queries {
		add(query.Name)
	}
	for _, job := range program.Jobs {
		add(job.Name)
	}
	for _, action := range program.Actions {
		add(action.Name)
	}
	for _, transaction := range program.Transactions {
		add(transaction.Name)
	}
	for _, role := range program.Roles {
		add(role.Name)
	}
	for _, api := range program.APIs {
		add(api.Name)
	}
	for _, layout := range program.Layouts {
		add(layout.Name)
	}
	for _, page := range program.Pages {
		add(page.Name)
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
	return symbols
}

func findService(program Program, name string) (ServiceDecl, bool) {
	for _, service := range program.Services {
		if service.Name == name {
			return service, true
		}
	}
	return ServiceDecl{}, false
}

func serviceAPIBindings(program Program) map[string]string {
	bindings := map[string]string{}
	for _, service := range program.Services {
		for _, target := range service.APIs {
			if _, exists := bindings[target.Name]; !exists {
				bindings[target.Name] = service.Name
			}
		}
	}
	return bindings
}

func (g *webGenerator) explicitAPIServiceName(api APIDecl) (string, bool) {
	for _, service := range g.program.Services {
		for _, target := range service.APIs {
			if target.Name == api.Name {
				return service.Name, true
			}
		}
	}
	return "", false
}
