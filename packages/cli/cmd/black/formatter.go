package main

import (
	"os"
	"strings"
)

func FormatBlackFile(file string, write bool, check bool) (FormatResult, string) {
	result := FormatResult{
		Success: true,
		Command: "format",
		Version: version,
		File:    file,
		Check:   check,
		Errors:  []Diagnostic{},
	}

	source, err := os.ReadFile(file)
	if err != nil {
		result.Success = false
		result.Errors = []Diagnostic{{
			File:       file,
			Code:       "FILE_READ_ERROR",
			Message:    err.Error(),
			Suggestion: "Check that the file exists and is readable.",
		}}
		return result, ""
	}

	formatted, diagnostics := FormatBlackSource(file, string(source))
	if len(diagnostics) > 0 {
		result.Success = false
		result.Errors = diagnostics
		return result, ""
	}

	result.Changed = formatted != string(source)
	if check && result.Changed {
		result.Success = false
		result.Errors = []Diagnostic{{
			File:       file,
			Code:       "FORMAT_REQUIRED",
			Message:    "BlackLang source is not formatted.",
			Suggestion: "Run `black format " + file + "`.",
		}}
		return result, formatted
	}

	if write && result.Changed {
		mode := os.FileMode(0644)
		if info, statErr := os.Stat(file); statErr == nil {
			mode = info.Mode().Perm()
		}
		if err := os.WriteFile(file, []byte(formatted), mode); err != nil {
			result.Success = false
			result.Errors = []Diagnostic{{
				File:       file,
				Code:       "FILE_WRITE_ERROR",
				Message:    err.Error(),
				Suggestion: "Check that the file is writable.",
			}}
			return result, formatted
		}
	}

	return result, formatted
}

func FormatBlackSource(file string, source string) (string, []Diagnostic) {
	tokens, diagnostics := lexSourceForFormat(file, source)
	if len(diagnostics) > 0 {
		return "", diagnostics
	}

	statements := tokensToStatements(tokens)
	lines := []string{}
	indent := 0

	for _, statement := range statements {
		codeTokens, commentTokens := splitFormatTokens(statement.Tokens)
		if len(codeTokens) == 0 && len(commentTokens) == 0 {
			continue
		}

		if len(codeTokens) > 0 && codeTokens[0].Value == "}" {
			indent--
			if indent < 0 {
				indent = 0
			}
		}

		if shouldInsertFormatBlankLine(lines, indent, codeTokens) {
			lines = append(lines, "")
		}

		line := formatCodeTokens(codeTokens)
		comment := formatCommentTokens(commentTokens)
		if comment != "" {
			if line != "" {
				line += " "
			}
			line += comment
		}
		if line != "" {
			line = strings.Repeat("  ", indent) + line
		}
		lines = append(lines, line)

		if len(codeTokens) > 0 && codeTokens[len(codeTokens)-1].Value == "{" {
			indent++
		}
	}

	if len(lines) == 0 {
		return "", nil
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func splitFormatTokens(tokens []sourceToken) ([]sourceToken, []sourceToken) {
	codeTokens := []sourceToken{}
	commentTokens := []sourceToken{}
	for _, token := range tokens {
		if token.Kind == tokenComment {
			commentTokens = append(commentTokens, token)
			continue
		}
		if token.Kind == tokenNewline {
			continue
		}
		codeTokens = append(codeTokens, token)
	}
	return codeTokens, commentTokens
}

func shouldInsertFormatBlankLine(lines []string, indent int, codeTokens []sourceToken) bool {
	if len(lines) == 0 || lines[len(lines)-1] == "" || len(codeTokens) == 0 {
		return false
	}
	previous := strings.TrimSpace(lines[len(lines)-1])
	if strings.HasSuffix(previous, "{") {
		return false
	}
	if indent == 0 && codeTokens[0].Value != "}" {
		return true
	}
	if indent == 1 && codeTokens[len(codeTokens)-1].Value == "{" {
		return true
	}
	return false
}

func formatCodeTokens(tokens []sourceToken) string {
	parts := []string{}
	for _, token := range tokens {
		if token.Value == "," {
			if len(parts) > 0 {
				parts[len(parts)-1] += ","
			}
			continue
		}
		parts = append(parts, formatTokenValue(token))
	}
	return strings.Join(parts, " ")
}

func formatTokenValue(token sourceToken) string {
	if token.Kind == tokenString {
		return quoteBlackString(token.Value)
	}
	return token.Value
}

func quoteBlackString(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, char := range value {
		switch char {
		case '\\':
			builder.WriteString(`\\`)
		case '"':
			builder.WriteString(`\"`)
		case '\n':
			builder.WriteString(`\n`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			builder.WriteRune(char)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

func formatCommentTokens(tokens []sourceToken) string {
	comments := []string{}
	for _, token := range tokens {
		comment := strings.TrimSpace(token.Value)
		if comment != "" {
			comments = append(comments, comment)
		}
	}
	return strings.Join(comments, " ")
}
