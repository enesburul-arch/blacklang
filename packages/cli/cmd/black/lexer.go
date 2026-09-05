package main

import (
	"fmt"
	"strings"
	"unicode"
)

type tokenKind int

const (
	tokenIdentifier tokenKind = iota
	tokenString
	tokenSymbol
	tokenOperator
	tokenNewline
	tokenComment
)

type sourceToken struct {
	Kind     tokenKind
	Value    string
	Position Position
}

type sourceStatement struct {
	Tokens   []sourceToken
	Position Position
}

func (statement sourceStatement) Parts() []string {
	parts := []string{}
	for _, token := range statement.Tokens {
		if token.Value == "," {
			continue
		}
		parts = append(parts, token.Value)
	}
	return parts
}

func tokenizeSource(file string, source string) ([]sourceStatement, []Diagnostic) {
	tokens, diagnostics := lexSource(file, source)
	return tokensToStatements(tokens), diagnostics
}

func lexSource(file string, source string) ([]sourceToken, []Diagnostic) {
	return lexSourceWithComments(file, source, false)
}

func lexSourceForFormat(file string, source string) ([]sourceToken, []Diagnostic) {
	return lexSourceWithComments(file, source, true)
}

func lexSourceWithComments(file string, source string, keepComments bool) ([]sourceToken, []Diagnostic) {
	normalized := strings.ReplaceAll(strings.ReplaceAll(source, "\r\n", "\n"), "\r", "\n")
	runes := []rune(normalized)
	tokens := []sourceToken{}
	diagnostics := []Diagnostic{}
	line := 1
	column := 1
	index := 0

	position := func() Position {
		return Position{File: file, Line: line, Column: column}
	}
	advance := func() {
		if index >= len(runes) {
			return
		}
		if runes[index] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
		index++
	}
	addToken := func(kind tokenKind, value string, pos Position) {
		tokens = append(tokens, sourceToken{Kind: kind, Value: value, Position: pos})
	}
	addDiagnostic := func(pos Position, code string, message string, suggestion string) {
		diagnostics = append(diagnostics, Diagnostic{
			File:       file,
			Line:       pos.Line,
			Column:     pos.Column,
			Code:       code,
			Message:    message,
			Suggestion: suggestion,
		})
	}

	for index < len(runes) {
		char := runes[index]
		pos := position()

		if char == '\n' {
			addToken(tokenNewline, "\n", pos)
			advance()
			continue
		}
		if unicode.IsSpace(char) {
			advance()
			continue
		}
		if char == '#' || (char == '/' && index+1 < len(runes) && runes[index+1] == '/') {
			start := pos
			var builder strings.Builder
			for index < len(runes) && runes[index] != '\n' {
				if keepComments {
					builder.WriteRune(runes[index])
				}
				advance()
			}
			if keepComments {
				addToken(tokenComment, strings.TrimRightFunc(builder.String(), unicode.IsSpace), start)
			}
			continue
		}
		if char == '"' {
			start := pos
			advance()
			var builder strings.Builder
			closed := false
			for index < len(runes) {
				current := runes[index]
				if current == '\n' {
					break
				}
				if current == '\\' && index+1 < len(runes) {
					advance()
					escaped := runes[index]
					switch escaped {
					case '"':
						builder.WriteRune('"')
					case '\\':
						builder.WriteRune('\\')
					case 'n':
						builder.WriteRune('\n')
					case 't':
						builder.WriteRune('\t')
					default:
						builder.WriteRune(escaped)
					}
					advance()
					continue
				}
				if current == '"' {
					closed = true
					advance()
					break
				}
				builder.WriteRune(current)
				advance()
			}
			if !closed {
				addDiagnostic(start, "UNCLOSED_STRING", "String literal is missing a closing quote.", "Close the string with `\"` on the same line.")
			}
			addToken(tokenString, builder.String(), start)
			continue
		}
		if isBlackSymbol(char) {
			addToken(tokenSymbol, string(char), pos)
			advance()
			continue
		}
		if isBlackOperatorStart(char) {
			value := string(char)
			advance()
			if index < len(runes) && runes[index] == '=' && (char == '=' || char == '!' || char == '<' || char == '>') {
				value += "="
				advance()
			}
			addToken(tokenOperator, value, pos)
			continue
		}

		var builder strings.Builder
		for index < len(runes) {
			current := runes[index]
			if current == '\n' || unicode.IsSpace(current) || current == '"' || isBlackSymbol(current) || isBlackOperatorStart(current) || current == '#' || (current == '/' && index+1 < len(runes) && runes[index+1] == '/') {
				break
			}
			builder.WriteRune(current)
			advance()
		}
		if builder.Len() == 0 {
			addDiagnostic(pos, "UNEXPECTED_CHARACTER", fmt.Sprintf("Unexpected character %q.", char), "Use identifiers, quoted strings, commas, braces, or supported comparison operators.")
			advance()
			continue
		}
		addToken(tokenIdentifier, builder.String(), pos)
	}

	return tokens, diagnostics
}

func tokensToStatements(tokens []sourceToken) []sourceStatement {
	statements := []sourceStatement{}
	current := []sourceToken{}

	flush := func() {
		if len(current) == 0 {
			return
		}
		hasMeaningfulToken := false
		for _, token := range current {
			if token.Value != "," {
				hasMeaningfulToken = true
				break
			}
		}
		if !hasMeaningfulToken {
			current = []sourceToken{}
			return
		}
		tokensCopy := append([]sourceToken(nil), current...)
		statements = append(statements, sourceStatement{
			Tokens:   tokensCopy,
			Position: tokensCopy[0].Position,
		})
		current = []sourceToken{}
	}

	for index := 0; index < len(tokens); index++ {
		token := tokens[index]
		if token.Kind == tokenNewline {
			flush()
			continue
		}
		if token.Value == "}" {
			flush()
			current = append(current, token)
			if index+1 < len(tokens) && tokens[index+1].Kind == tokenComment && tokens[index+1].Position.Line == token.Position.Line {
				index++
				current = append(current, tokens[index])
			}
			flush()
			continue
		}
		current = append(current, token)
		if token.Value == "{" {
			if index+1 < len(tokens) && tokens[index+1].Kind == tokenComment && tokens[index+1].Position.Line == token.Position.Line {
				index++
				current = append(current, tokens[index])
			}
			flush()
		}
	}
	flush()
	return statements
}

func isBlackSymbol(char rune) bool {
	return char == '{' || char == '}' || char == ',' || char == ';'
}

func isBlackOperatorStart(char rune) bool {
	return char == '=' || char == '!' || char == '<' || char == '>'
}
