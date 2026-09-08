package main

import (
	"fmt"
	"regexp"
	"strings"
)

type TestDecl struct {
	Name         string                `json:"name"`
	Page         string                `json:"page,omitempty"`
	PagePosition Position              `json:"pagePosition,omitempty"`
	Expectations []TestExpectationDecl `json:"expectations,omitempty"`
	Position     Position              `json:"position"`
}

type TestExpectationDecl struct {
	Kind     string   `json:"kind"`
	Value    string   `json:"value"`
	Position Position `json:"position"`
}

var testNamePattern = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)

func (p *parser) parseTest(start int, parts []string) int {
	line := p.lineNumber(start)
	statement := p.lines[start]
	if len(parts) != 3 || len(statement.Tokens) != 3 || !queryStatementIdentifiers(statement, 0, 1) || statement.Tokens[2].Kind != tokenSymbol || parts[2] != "{" {
		p.addError(line, 1, "INVALID_TEST_DECLARATION", "Test declaration must be `test Name {`.", "Example: `test WarehouseBrowserSmoke {`.")
		return start
	}

	test := TestDecl{
		Name:         parts[1],
		Expectations: []TestExpectationDecl{},
		Position:     p.position(line, 1),
	}
	seen := map[string]bool{}

	for index := start + 1; index < len(p.lines); index++ {
		statement := p.lines[index]
		tokens := statement.Tokens
		line := p.lineNumber(index)
		rowParts := statement.Parts()
		if isClosingBrace(rowParts) {
			p.program.Tests = append(p.program.Tests, test)
			return index
		}
		if len(tokens) == 0 {
			continue
		}
		if tokens[0].Kind != tokenIdentifier {
			p.addError(line, 1, "UNEXPECTED_TEST_TOKEN", "Test clauses must begin with page or expect.", "Use one test clause per line.")
			continue
		}

		switch rowParts[0] {
		case "page":
			if seen["page"] {
				p.addError(line, 1, "DUPLICATE_TEST_PAGE", fmt.Sprintf("Test %s already declares page.", test.Name), "Keep one page clause inside each test.")
				continue
			}
			seen["page"] = true
			if len(tokens) != 2 || !queryStatementIdentifiers(statement, 1) {
				p.addError(line, 1, "INVALID_TEST_PAGE", "Test page must be `page PageName`.", "Example: `page Products`.")
				continue
			}
			test.Page = tokens[1].Value
			test.PagePosition = tokens[1].Position
		case "expect":
			expectation, ok := parseTestExpectation(tokens)
			if !ok {
				p.addError(line, 1, "INVALID_TEST_EXPECT", "Test expectation must be `expect text \"Text\"`, `expect page PageName`, or `expect action actionName`.", "Use one deterministic expectation per line.")
				continue
			}
			test.Expectations = append(test.Expectations, expectation)
		default:
			p.addError(line, 1, "UNEXPECTED_TEST_TOKEN", fmt.Sprintf("Unexpected test token %q.", rowParts[0]), "Use page or expect inside a test block.")
		}
	}

	p.addError(line, 1, "UNCLOSED_TEST", fmt.Sprintf("Test %s is missing a closing brace.", test.Name), "Add `}` after the test body.")
	return len(p.lines) - 1
}

func parseTestExpectation(tokens []sourceToken) (TestExpectationDecl, bool) {
	if len(tokens) != 3 || tokens[0].Kind != tokenIdentifier || tokens[0].Value != "expect" || tokens[1].Kind != tokenIdentifier {
		return TestExpectationDecl{}, false
	}
	kind := tokens[1].Value
	switch kind {
	case "text":
		if tokens[2].Kind != tokenString {
			return TestExpectationDecl{}, false
		}
	case "page", "action":
		if tokens[2].Kind != tokenIdentifier {
			return TestExpectationDecl{}, false
		}
	default:
		return TestExpectationDecl{}, false
	}
	return TestExpectationDecl{
		Kind:     kind,
		Value:    tokens[2].Value,
		Position: tokens[0].Position,
	}, true
}

func (v *semanticValidator) validateTests(pageIndex map[string]PageDecl) map[string]TestDecl {
	testIndex := map[string]TestDecl{}
	normalizedNames := map[string]string{}
	symbols := testOtherSymbols(v.program)
	for _, test := range v.program.Tests {
		if existing, ok := testIndex[test.Name]; ok {
			v.addDiagnostic(test.Position, "DUPLICATE_TEST", fmt.Sprintf("Test %s is already defined.", test.Name), fmt.Sprintf("First definition is at %s:%d.", existing.Position.File, existing.Position.Line))
			continue
		}
		testIndex[test.Name] = test
		if !testNamePattern.MatchString(test.Name) {
			v.addDiagnostic(test.Position, "INVALID_TEST_NAME", fmt.Sprintf("Test name %q must use PascalCase letters and digits.", test.Name), "Use a name such as WarehouseBrowserSmoke.")
		}
		normalized := strings.ToLower(test.Name)
		if existing, ok := normalizedNames[normalized]; ok {
			v.addDiagnostic(test.Position, "TEST_NAME_COLLISION", fmt.Sprintf("Test %s conflicts with test %s after name normalization.", test.Name, existing), "Choose distinct test names, including after lowercasing.")
		}
		normalizedNames[normalized] = test.Name
		if symbols[test.Name] || symbols[normalized] {
			v.addDiagnostic(test.Position, "TEST_NAME_COLLISION", fmt.Sprintf("Test %s conflicts with another application symbol.", test.Name), "Choose a unique test name for unambiguous inspect --affected output.")
		}
		v.validateTest(test, pageIndex)
	}
	return testIndex
}

func (v *semanticValidator) validateTest(test TestDecl, pageIndex map[string]PageDecl) {
	if test.Page == "" {
		v.addDiagnostic(test.Position, "MISSING_TEST_PAGE", fmt.Sprintf("Test %s is missing a page.", test.Name), "Add `page PageName` inside the test.")
		return
	}
	page, ok := pageIndex[test.Page]
	if !ok {
		v.addDiagnostic(test.PagePosition, "UNKNOWN_TEST_PAGE", fmt.Sprintf("Test %s references unknown page %s.", test.Name, test.Page), "Use a page declared in the current source.")
		return
	}
	if len(test.Expectations) == 0 {
		v.addDiagnostic(test.Position, "MISSING_TEST_EXPECT", fmt.Sprintf("Test %s has no expectations.", test.Name), "Add at least one `expect text`, `expect page`, or `expect action` line.")
		return
	}
	seen := map[string]Position{}
	for _, expectation := range test.Expectations {
		key := expectation.Kind + ":" + expectation.Value
		if existing, ok := seen[key]; ok {
			v.addDiagnostic(expectation.Position, "DUPLICATE_TEST_EXPECT", fmt.Sprintf("Test %s repeats expectation %s.", test.Name, key), fmt.Sprintf("First expectation is at %s:%d.", existing.File, existing.Line))
			continue
		}
		seen[key] = expectation.Position
		switch expectation.Kind {
		case "text":
			value := strings.TrimSpace(expectation.Value)
			if value == "" {
				v.addDiagnostic(expectation.Position, "INVALID_TEST_EXPECT_TEXT", fmt.Sprintf("Test %s has an empty text expectation.", test.Name), "Use a non-empty generated UI text literal.")
			} else if len([]rune(value)) > 200 {
				v.addDiagnostic(expectation.Position, "INVALID_TEST_EXPECT_TEXT", fmt.Sprintf("Test %s text expectation is too long.", test.Name), "Keep browser text expectations at 200 characters or fewer.")
			}
		case "page":
			if _, ok := pageIndex[expectation.Value]; !ok {
				v.addDiagnostic(expectation.Position, "UNKNOWN_TEST_EXPECT_PAGE", fmt.Sprintf("Test %s expects unknown page %s.", test.Name, expectation.Value), "Use a declared page name.")
			}
		case "action":
			if !hasAction(page, expectation.Value) {
				v.addDiagnostic(expectation.Position, "UNKNOWN_TEST_EXPECT_ACTION", fmt.Sprintf("Test %s expects action %s on page %s, but the page does not expose it.", test.Name, expectation.Value, page.Name), "Use an action listed in the target page actions.")
			}
		default:
			v.addDiagnostic(expectation.Position, "UNSUPPORTED_TEST_EXPECT", fmt.Sprintf("Test %s uses unsupported expectation kind %s.", test.Name, expectation.Kind), "Use text, page, or action.")
		}
	}
}

func testOtherSymbols(program Program) map[string]bool {
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
	for _, seed := range program.Seeds {
		add(seed.Name)
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
	return symbols
}

func findTest(program Program, name string) (TestDecl, bool) {
	for _, test := range program.Tests {
		if test.Name == name {
			return test, true
		}
	}
	return TestDecl{}, false
}
