package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type secretPattern struct {
	code       string
	pattern    *regexp.Regexp
	message    string
	suggestion string
}

var secretPatterns = []secretPattern{
	{
		code:       "HARDCODED_DATABASE_URL",
		pattern:    regexp.MustCompile(`(?i)(postgres|postgresql|mysql|mongodb|redis)://[^"\s]+`),
		message:    "Possible hardcoded database connection string.",
		suggestion: "Use `database { url env DATABASE_URL }` and keep the real value outside .black source.",
	},
	{
		code:       "HARDCODED_PRIVATE_KEY",
		pattern:    regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`),
		message:    "Possible hardcoded private key.",
		suggestion: "Move private keys to a secret manager or environment-provided file.",
	},
	{
		code:       "HARDCODED_TOKEN",
		pattern:    regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password)\s+["']?[A-Za-z0-9_\-./+=:]{12,}`),
		message:    "Possible hardcoded secret value.",
		suggestion: "Reference secrets with env names instead of storing secret values in .black source.",
	},
}

func SecurityScanSource(file string) SecurityScanResult {
	source, err := os.ReadFile(file)
	if err != nil {
		return SecurityScanResult{
			Success: false,
			Command: "security scan",
			Version: version,
			File:    file,
			Errors: []Diagnostic{{
				File:       file,
				Code:       "FILE_READ_ERROR",
				Message:    err.Error(),
				Suggestion: "Pass a readable .black file path or set source in blacklang.toml.",
			}},
			Findings: []Diagnostic{},
		}
	}

	findings := SecurityScanText(file, string(source))
	return SecurityScanResult{
		Success:  len(findings) == 0,
		Command:  "security scan",
		Version:  version,
		File:     file,
		Findings: findings,
		Errors:   []Diagnostic{},
	}
}

func SecurityScanText(file string, source string) []Diagnostic {
	findings := []Diagnostic{}
	for index, line := range strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		for _, secret := range secretPatterns {
			location := secret.pattern.FindStringIndex(trimmed)
			if location == nil {
				continue
			}
			findings = append(findings, Diagnostic{
				File:       file,
				Line:       index + 1,
				Column:     location[0] + 1,
				Code:       secret.code,
				Message:    secret.message,
				Suggestion: secret.suggestion,
			})
		}
	}

	return findings
}

func FormatSecurityScanIR(result SecurityScanResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("security scan ok\n")
	} else {
		builder.WriteString("security scan findings\n")
	}
	builder.WriteString(fmt.Sprintf("file %s\n", result.File))
	builder.WriteString(fmt.Sprintf("findings %d\n", len(result.Findings)))
	for _, finding := range result.Findings {
		builder.WriteString(fmt.Sprintf("  %s %s:%d:%d\n", finding.Code, finding.File, finding.Line, finding.Column))
	}
	return builder.String()
}

func EncryptedSourceMode() EncryptedSourceResult {
	return EncryptedSourceResult{
		Success:          true,
		Command:          "security encrypted-source",
		Version:          version,
		Status:           "planned",
		Extension:        ".black.enc",
		ProtectedFiles:   []string{".black", ".black.enc", "blacklang.toml"},
		ProductionPolicy: "Production packages exclude .black and .black.enc source files by default.",
		BuildPolicy:      "Draft v0.1 documents encrypted source mode; direct build from .black.enc is not implemented yet.",
		KeyPolicy:        "Encryption keys must come from the environment, OS keychain, or a future secret manager integration; keys must never be stored in .black source.",
		Rules: []string{
			"Treat .black source as protected source of truth.",
			"Never put secrets, passwords, API keys, tokens, or private keys in .black files.",
			"Build production artifacts from trusted developer or CI machines.",
			"Deploy generated artifacts, not protected BlackLang source.",
			"Use .black.enc for future encrypted-at-rest source storage.",
		},
		Errors: []Diagnostic{},
	}
}

func FormatEncryptedSourceIR(result EncryptedSourceResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	builder.WriteString("security encrypted-source\n")
	builder.WriteString(fmt.Sprintf("status %s\n", result.Status))
	builder.WriteString(fmt.Sprintf("extension %s\n", result.Extension))
	builder.WriteString(fmt.Sprintf("rules %d\n", len(result.Rules)))
	for _, rule := range result.Rules {
		builder.WriteString(fmt.Sprintf("  rule %s\n", rule))
	}
	return builder.String()
}
