package main

import (
	"fmt"
	"strings"
)

type TransactionDecl struct {
	Name     string                  `json:"name"`
	Actions  []TransactionTargetDecl `json:"actions,omitempty"`
	APIs     []TransactionTargetDecl `json:"apis,omitempty"`
	Position Position                `json:"position"`
}

type TransactionTargetDecl struct {
	Name     string   `json:"name"`
	Position Position `json:"position"`
}

func (p *parser) parseTransaction(start int, parts []string) int {
	line := p.lineNumber(start)
	statement := p.lines[start]
	if len(parts) != 3 || len(statement.Tokens) != 3 || !queryStatementIdentifiers(statement, 0, 1) || statement.Tokens[2].Kind != tokenSymbol || parts[2] != "{" {
		p.addError(line, 1, "INVALID_TRANSACTION_DECLARATION", "Transaction declaration must be `transaction Name {`.", "Example: `transaction RestockAtomic {`.")
		return start
	}

	transaction := TransactionDecl{
		Name:     parts[1],
		Actions:  []TransactionTargetDecl{},
		APIs:     []TransactionTargetDecl{},
		Position: p.position(line, 1),
	}
	seenTargets := map[string]Position{}

	for index := start + 1; index < len(p.lines); index++ {
		currentLine := p.lineNumber(index)
		statement := p.lines[index]
		tokens := statement.Tokens
		rowParts := statement.Parts()
		if isClosingBrace(rowParts) {
			p.program.Transactions = append(p.program.Transactions, transaction)
			return index
		}
		if len(tokens) == 0 {
			continue
		}
		if tokens[0].Kind != tokenIdentifier {
			p.addError(currentLine, 1, "UNEXPECTED_TRANSACTION_TOKEN", "Transaction clauses must begin with action or api.", "Use one transaction target per line.")
			continue
		}

		keyword := tokens[0].Value
		switch keyword {
		case "action", "api":
			if len(tokens) != 2 || !queryStatementIdentifiers(statement, 1) {
				code := "INVALID_TRANSACTION_ACTION"
				if keyword == "api" {
					code = "INVALID_TRANSACTION_API"
				}
				p.addError(currentLine, 1, code, "Transaction targets must be `action Name` or `api Name`.", "Example: `action RestockProduct`.")
				continue
			}
			target := TransactionTargetDecl{Name: tokens[1].Value, Position: tokens[1].Position}
			targetKey := keyword + ":" + target.Name
			if existing, ok := seenTargets[targetKey]; ok {
				p.addError(currentLine, 1, "DUPLICATE_TRANSACTION_TARGET", fmt.Sprintf("Transaction %s already targets %s %s.", transaction.Name, keyword, target.Name), fmt.Sprintf("First definition is at %s:%d.", existing.File, existing.Line))
				continue
			}
			seenTargets[targetKey] = target.Position
			if keyword == "action" {
				transaction.Actions = append(transaction.Actions, target)
			} else {
				transaction.APIs = append(transaction.APIs, target)
			}
		default:
			p.addError(currentLine, 1, "UNEXPECTED_TRANSACTION_TOKEN", fmt.Sprintf("Unexpected transaction token %q.", keyword), "Use action or api inside a transaction block.")
		}
	}

	p.addError(line, 1, "UNCLOSED_TRANSACTION", fmt.Sprintf("Transaction %s is missing a closing brace.", transaction.Name), "Add `}` after the transaction body.")
	return len(p.lines) - 1
}

func (v *semanticValidator) validateTransactions(actions map[string]CustomActionDecl, apis map[string]APIDecl) map[string]TransactionDecl {
	transactions := map[string]TransactionDecl{}
	normalizedNames := map[string]string{}
	boundActions := map[string]TransactionDecl{}
	boundAPIs := map[string]TransactionDecl{}
	symbols := transactionOtherSymbols(v.program)

	for _, transaction := range v.program.Transactions {
		if existing, ok := transactions[transaction.Name]; ok {
			v.addDiagnostic(transaction.Position, "DUPLICATE_TRANSACTION", fmt.Sprintf("Transaction %s is already defined.", transaction.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		transactions[transaction.Name] = transaction

		if !queryNamePattern.MatchString(transaction.Name) {
			v.addDiagnostic(transaction.Position, "INVALID_TRANSACTION_NAME", fmt.Sprintf("Transaction name %q must use PascalCase letters and digits.", transaction.Name), "Use a name such as RestockAtomic.")
		}
		normalized := strings.ToLower(transaction.Name)
		if existing, ok := normalizedNames[normalized]; ok {
			v.addDiagnostic(transaction.Position, "TRANSACTION_NAME_COLLISION", fmt.Sprintf("Transaction %s conflicts with transaction %s after name normalization.", transaction.Name, existing), "Choose distinct transaction names, including after lowercasing.")
		}
		normalizedNames[normalized] = transaction.Name
		if symbols[transaction.Name] || symbols[normalized] {
			v.addDiagnostic(transaction.Position, "TRANSACTION_NAME_COLLISION", fmt.Sprintf("Transaction %s conflicts with another application symbol.", transaction.Name), "Choose a unique transaction name for unambiguous inspect --affected output.")
		}

		if len(transaction.Actions) == 0 && len(transaction.APIs) == 0 {
			v.addDiagnostic(transaction.Position, "EMPTY_TRANSACTION", fmt.Sprintf("Transaction %s does not target any action or api.", transaction.Name), "Add at least one `action Name` or `api Name` line.")
		}

		for _, target := range transaction.Actions {
			if !queryNamePattern.MatchString(target.Name) {
				v.addDiagnostic(target.Position, "INVALID_TRANSACTION_ACTION", fmt.Sprintf("Transaction %s action target %q must use PascalCase.", transaction.Name, target.Name), "Reference a declared custom action such as RestockProduct.")
				continue
			}
			if _, ok := actions[target.Name]; !ok {
				v.addDiagnostic(target.Position, "UNKNOWN_TRANSACTION_ACTION", fmt.Sprintf("Transaction %s references unknown action %s.", transaction.Name, target.Name), "Declare the action or remove the transaction target.")
				continue
			}
			if existing, ok := boundActions[target.Name]; ok {
				v.addDiagnostic(target.Position, "DUPLICATE_TRANSACTION_BINDING", fmt.Sprintf("Action %s is already bound to transaction %s.", target.Name, existing.Name), "Bind each action to at most one transaction block.")
				continue
			}
			boundActions[target.Name] = transaction
		}

		for _, target := range transaction.APIs {
			if !queryNamePattern.MatchString(target.Name) {
				v.addDiagnostic(target.Position, "INVALID_TRANSACTION_API", fmt.Sprintf("Transaction %s api target %q must use PascalCase.", transaction.Name, target.Name), "Reference a declared api such as StockWebhook.")
				continue
			}
			api, ok := apis[target.Name]
			if !ok {
				v.addDiagnostic(target.Position, "UNKNOWN_TRANSACTION_API", fmt.Sprintf("Transaction %s references unknown api %s.", transaction.Name, target.Name), "Declare the api or remove the transaction target.")
				continue
			}
			if api.Update == nil {
				v.addDiagnostic(target.Position, "UNSUPPORTED_TRANSACTION_API", fmt.Sprintf("Transaction %s targets api %s without an update handler.", transaction.Name, target.Name), "Only APIs with `update Entity where ... set ...` need a generated transaction boundary.")
				continue
			}
			if existing, ok := boundAPIs[target.Name]; ok {
				v.addDiagnostic(target.Position, "DUPLICATE_TRANSACTION_BINDING", fmt.Sprintf("API %s is already bound to transaction %s.", target.Name, existing.Name), "Bind each api to at most one transaction block.")
				continue
			}
			boundAPIs[target.Name] = transaction
		}
	}

	return transactions
}

func transactionOtherSymbols(program Program) map[string]bool {
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
	for _, service := range program.Services {
		add(service.Name)
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

func findTransaction(program Program, name string) (TransactionDecl, bool) {
	for _, transaction := range program.Transactions {
		if transaction.Name == name {
			return transaction, true
		}
	}
	return TransactionDecl{}, false
}

func transactionActionBindings(program Program) map[string]string {
	bindings := map[string]string{}
	for _, transaction := range program.Transactions {
		for _, target := range transaction.Actions {
			if _, exists := bindings[target.Name]; !exists {
				bindings[target.Name] = transaction.Name
			}
		}
	}
	return bindings
}

func transactionAPIBindings(program Program) map[string]string {
	bindings := map[string]string{}
	for _, transaction := range program.Transactions {
		for _, target := range transaction.APIs {
			if _, exists := bindings[target.Name]; !exists {
				bindings[target.Name] = transaction.Name
			}
		}
	}
	return bindings
}

func (g *webGenerator) customActionTransactionName(action CustomActionDecl) (string, bool) {
	for _, transaction := range g.program.Transactions {
		for _, target := range transaction.Actions {
			if target.Name == action.Name {
				return transaction.Name, true
			}
		}
	}
	if action.Transaction {
		return action.Name, true
	}
	return "", false
}

func (g *webGenerator) explicitAPITransactionName(api APIDecl) (string, bool) {
	for _, transaction := range g.program.Transactions {
		for _, target := range transaction.APIs {
			if target.Name == api.Name {
				return transaction.Name, true
			}
		}
	}
	return "", false
}
