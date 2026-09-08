package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

const defaultSourceKeyEnv = "BLACKLANG_SOURCE_KEY"
const encryptedSourceMagic = "BLACKLANG-ENC v1"
const encryptedSourceAlg = "AES-256-GCM"
const encryptedSourceKDF = "SHA256-ENV"
const encryptedSourceDelimiter = "---"
const encryptedSourceAAD = "BLACKLANG-ENC v1\nalg AES-256-GCM\nkdf SHA256-ENV"

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
	source, diagnostics := ReadBlackSource(file)
	if len(diagnostics) > 0 {
		return SecurityScanResult{
			Success:  false,
			Command:  "security scan",
			Version:  version,
			File:     file,
			Errors:   diagnostics,
			Findings: []Diagnostic{},
		}
	}

	findings := SecurityScanText(file, source)
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

func ReadBlackSource(file string) (string, []Diagnostic) {
	source, err := os.ReadFile(file)
	if err != nil {
		return "", []Diagnostic{{
			File:       file,
			Code:       "FILE_READ_ERROR",
			Message:    err.Error(),
			Suggestion: "Pass a readable .black file path or set source in blacklang.toml.",
		}}
	}
	if !isEncryptedSourcePath(file) {
		return string(source), []Diagnostic{}
	}
	plaintext, result := DecryptBlackSourceBytes(file, source, "")
	if !result.Success {
		return "", result.Errors
	}
	return plaintext, []Diagnostic{}
}

func EncryptBlackSourceFile(file string, outFile string, keyEnv string) SecurityEncryptResult {
	if keyEnv == "" {
		keyEnv = defaultSourceKeyEnv
	}
	if outFile == "" {
		outFile = defaultEncryptedSourceOut(file)
	}
	plaintext, err := os.ReadFile(file)
	if err != nil {
		return SecurityEncryptResult{
			Success: false,
			Command: "security encrypt",
			Version: version,
			File:    file,
			OutFile: outFile,
			KeyEnv:  keyEnv,
			Errors: []Diagnostic{{
				File:       file,
				Code:       "FILE_READ_ERROR",
				Message:    err.Error(),
				Suggestion: "Pass a readable .black file path.",
			}},
		}
	}

	encrypted, header, diagnostics := EncryptBlackSourceBytes(plaintext, keyEnv)
	if len(diagnostics) > 0 {
		return SecurityEncryptResult{
			Success: false,
			Command: "security encrypt",
			Version: version,
			File:    file,
			OutFile: outFile,
			KeyEnv:  keyEnv,
			Header:  header,
			Errors:  diagnostics,
		}
	}
	if err := os.WriteFile(outFile, encrypted, 0o600); err != nil {
		return SecurityEncryptResult{
			Success: false,
			Command: "security encrypt",
			Version: version,
			File:    file,
			OutFile: outFile,
			KeyEnv:  keyEnv,
			Header:  header,
			Errors: []Diagnostic{{
				File:       outFile,
				Code:       "ENCRYPTION_WRITE_ERROR",
				Message:    err.Error(),
				Suggestion: "Choose a writable --out path for the encrypted .black.enc file.",
			}},
		}
	}

	return SecurityEncryptResult{
		Success:         true,
		Command:         "security encrypt",
		Version:         version,
		File:            file,
		OutFile:         outFile,
		KeyEnv:          keyEnv,
		Header:          header,
		PlaintextBytes:  int64(len(plaintext)),
		CiphertextBytes: int64(len(encrypted)),
		Errors:          []Diagnostic{},
	}
}

func EncryptBlackSourceBytes(plaintext []byte, keyEnv string) ([]byte, EncryptedSourceHeader, []Diagnostic) {
	if keyEnv == "" {
		keyEnv = defaultSourceKeyEnv
	}
	key, diagnostic := sourceKeyFromEnv(keyEnv, "")
	if diagnostic != nil {
		return nil, EncryptedSourceHeader{}, []Diagnostic{*diagnostic}
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, EncryptedSourceHeader{}, []Diagnostic{{
			Code:       "ENCRYPTION_RANDOM_ERROR",
			Message:    err.Error(),
			Suggestion: "Retry encryption on a system with a working cryptographic random source.",
		}}
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, EncryptedSourceHeader{}, []Diagnostic{{
			Code:       "ENCRYPTION_KEY_ERROR",
			Message:    err.Error(),
			Suggestion: "Use a non-empty environment key value.",
		}}
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, EncryptedSourceHeader{}, []Diagnostic{{
			Code:       "ENCRYPTION_CIPHER_ERROR",
			Message:    err.Error(),
			Suggestion: "Retry with the built-in AES-GCM implementation.",
		}}
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, []byte(encryptedSourceAAD))
	header := EncryptedSourceHeader{
		Version:     "v1",
		Alg:         encryptedSourceAlg,
		KDF:         encryptedSourceKDF,
		KeyEnv:      keyEnv,
		NonceBase64: base64.StdEncoding.EncodeToString(nonce),
	}
	return []byte(formatEncryptedSource(header, ciphertext)), header, []Diagnostic{}
}

func DecryptBlackSourceFile(file string, keyEnv string) (string, SecurityDecryptResult) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", SecurityDecryptResult{
			Success: false,
			Command: "security decrypt",
			Version: version,
			File:    file,
			KeyEnv:  keyEnv,
			Errors: []Diagnostic{{
				File:       file,
				Code:       "FILE_READ_ERROR",
				Message:    err.Error(),
				Suggestion: "Pass a readable .black.enc file path.",
			}},
		}
	}
	return DecryptBlackSourceBytes(file, data, keyEnv)
}

func DecryptBlackSourceBytes(file string, data []byte, keyEnv string) (string, SecurityDecryptResult) {
	header, ciphertext, diagnostics := parseEncryptedSource(file, string(data))
	if len(diagnostics) > 0 {
		return "", SecurityDecryptResult{
			Success: false,
			Command: "security decrypt",
			Version: version,
			File:    file,
			KeyEnv:  keyEnv,
			Header:  header,
			Errors:  diagnostics,
		}
	}
	if keyEnv == "" {
		keyEnv = header.KeyEnv
	}
	key, diagnostic := sourceKeyFromEnv(keyEnv, file)
	if diagnostic != nil {
		return "", SecurityDecryptResult{
			Success: false,
			Command: "security decrypt",
			Version: version,
			File:    file,
			KeyEnv:  keyEnv,
			Header:  header,
			Errors:  []Diagnostic{*diagnostic},
		}
	}
	nonce, err := base64.StdEncoding.DecodeString(header.NonceBase64)
	if err != nil {
		return "", SecurityDecryptResult{
			Success: false,
			Command: "security decrypt",
			Version: version,
			File:    file,
			KeyEnv:  keyEnv,
			Header:  header,
			Errors: []Diagnostic{{
				File:       file,
				Code:       "INVALID_ENCRYPTED_SOURCE",
				Message:    "Encrypted source nonce is not valid base64.",
				Suggestion: "Recreate the .black.enc file with `black security encrypt`.",
			}},
		}
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", SecurityDecryptResult{
			Success: false,
			Command: "security decrypt",
			Version: version,
			File:    file,
			KeyEnv:  keyEnv,
			Header:  header,
			Errors: []Diagnostic{{
				File:       file,
				Code:       "ENCRYPTION_KEY_ERROR",
				Message:    err.Error(),
				Suggestion: "Use a non-empty environment key value.",
			}},
		}
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", SecurityDecryptResult{
			Success: false,
			Command: "security decrypt",
			Version: version,
			File:    file,
			KeyEnv:  keyEnv,
			Header:  header,
			Errors: []Diagnostic{{
				File:       file,
				Code:       "ENCRYPTION_CIPHER_ERROR",
				Message:    err.Error(),
				Suggestion: "Retry with the built-in AES-GCM implementation.",
			}},
		}
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, []byte(encryptedSourceAAD))
	if err != nil {
		return "", SecurityDecryptResult{
			Success: false,
			Command: "security decrypt",
			Version: version,
			File:    file,
			KeyEnv:  keyEnv,
			Header:  header,
			Errors: []Diagnostic{{
				File:       file,
				Code:       "DECRYPTION_FAILED",
				Message:    "Encrypted source could not be decrypted with the configured key.",
				Suggestion: "Check the key environment variable and encrypted file header.",
			}},
		}
	}
	return string(plaintext), SecurityDecryptResult{
		Success:        true,
		Command:        "security decrypt",
		Version:        version,
		File:           file,
		KeyEnv:         keyEnv,
		Header:         header,
		PlaintextBytes: int64(len(plaintext)),
		Errors:         []Diagnostic{},
	}
}

func formatEncryptedSource(header EncryptedSourceHeader, ciphertext []byte) string {
	return strings.Join([]string{
		encryptedSourceMagic,
		"alg " + header.Alg,
		"kdf " + header.KDF,
		"keyEnv " + header.KeyEnv,
		"nonce " + header.NonceBase64,
		encryptedSourceDelimiter,
		base64.StdEncoding.EncodeToString(ciphertext),
		"",
	}, "\n")
}

func parseEncryptedSource(file string, content string) (EncryptedSourceHeader, []byte, []Diagnostic) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) < 7 || strings.TrimSpace(lines[0]) != encryptedSourceMagic {
		return EncryptedSourceHeader{}, nil, []Diagnostic{invalidEncryptedSource(file, "Encrypted source must start with `BLACKLANG-ENC v1`.")}
	}
	header := EncryptedSourceHeader{Version: "v1"}
	delimiterIndex := -1
	for index := 1; index < len(lines); index++ {
		line := strings.TrimSpace(lines[index])
		if line == encryptedSourceDelimiter {
			delimiterIndex = index
			break
		}
		key, value, ok := strings.Cut(line, " ")
		if !ok {
			return header, nil, []Diagnostic{invalidEncryptedSource(file, "Encrypted source header line must use `key value` format.")}
		}
		switch key {
		case "alg":
			header.Alg = value
		case "kdf":
			header.KDF = value
		case "keyEnv":
			header.KeyEnv = value
		case "nonce":
			header.NonceBase64 = value
		default:
			return header, nil, []Diagnostic{invalidEncryptedSource(file, "Encrypted source header contains an unknown field.")}
		}
	}
	if delimiterIndex == -1 {
		return header, nil, []Diagnostic{invalidEncryptedSource(file, "Encrypted source header must end with `---`.")}
	}
	if header.Alg != encryptedSourceAlg || header.KDF != encryptedSourceKDF || header.KeyEnv == "" || header.NonceBase64 == "" {
		return header, nil, []Diagnostic{invalidEncryptedSource(file, "Encrypted source header is incomplete or unsupported.")}
	}
	body := strings.TrimSpace(strings.Join(lines[delimiterIndex+1:], ""))
	ciphertext, err := base64.StdEncoding.DecodeString(body)
	if err != nil || len(ciphertext) == 0 {
		return header, nil, []Diagnostic{invalidEncryptedSource(file, "Encrypted source body is not valid base64 ciphertext.")}
	}
	return header, ciphertext, []Diagnostic{}
}

func invalidEncryptedSource(file string, message string) Diagnostic {
	return Diagnostic{
		File:       file,
		Code:       "INVALID_ENCRYPTED_SOURCE",
		Message:    message,
		Suggestion: "Recreate the .black.enc file with `black security encrypt`.",
	}
}

func sourceKeyFromEnv(keyEnv string, file string) ([32]byte, *Diagnostic) {
	value := os.Getenv(keyEnv)
	if value == "" {
		return [32]byte{}, &Diagnostic{
			File:       file,
			Code:       "MISSING_ENCRYPTION_KEY",
			Message:    "Encrypted source key environment variable is not set.",
			Suggestion: "Set " + keyEnv + " in the environment; do not store the key in .black source.",
		}
	}
	key := sha256.Sum256([]byte(value))
	return key, nil
}

func defaultEncryptedSourceOut(file string) string {
	if strings.HasSuffix(file, ".black") {
		return file + ".enc"
	}
	return file + ".enc"
}

func isEncryptedSourcePath(file string) bool {
	return strings.HasSuffix(file, ".black.enc")
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
	builder.WriteString(fmt.Sprintf("errors %d\n", len(result.Errors)))
	for _, diagnostic := range result.Errors {
		builder.WriteString(fmt.Sprintf("  %s %s:%d:%d\n", diagnostic.Code, diagnostic.File, diagnostic.Line, diagnostic.Column))
	}
	return builder.String()
}

func FormatSecurityEncryptIR(result SecurityEncryptResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("security encrypt ok\n")
	} else {
		builder.WriteString("security encrypt failed\n")
	}
	builder.WriteString(fmt.Sprintf("file %s\n", result.File))
	builder.WriteString(fmt.Sprintf("out %s\n", result.OutFile))
	builder.WriteString(fmt.Sprintf("keyEnv %s\n", result.KeyEnv))
	builder.WriteString(fmt.Sprintf("alg %s\n", result.Header.Alg))
	builder.WriteString(fmt.Sprintf("kdf %s\n", result.Header.KDF))
	builder.WriteString(fmt.Sprintf("plaintextBytes %d\n", result.PlaintextBytes))
	builder.WriteString(fmt.Sprintf("ciphertextBytes %d\n", result.CiphertextBytes))
	builder.WriteString(fmt.Sprintf("errors %d\n", len(result.Errors)))
	for _, diagnostic := range result.Errors {
		builder.WriteString(fmt.Sprintf("  %s %s:%d:%d\n", diagnostic.Code, diagnostic.File, diagnostic.Line, diagnostic.Column))
	}
	return builder.String()
}

func FormatSecurityDecryptIR(result SecurityDecryptResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("security decrypt ok\n")
	} else {
		builder.WriteString("security decrypt failed\n")
	}
	builder.WriteString(fmt.Sprintf("file %s\n", result.File))
	builder.WriteString(fmt.Sprintf("keyEnv %s\n", result.KeyEnv))
	builder.WriteString(fmt.Sprintf("alg %s\n", result.Header.Alg))
	builder.WriteString(fmt.Sprintf("kdf %s\n", result.Header.KDF))
	builder.WriteString(fmt.Sprintf("plaintextBytes %d\n", result.PlaintextBytes))
	builder.WriteString(fmt.Sprintf("errors %d\n", len(result.Errors)))
	for _, diagnostic := range result.Errors {
		builder.WriteString(fmt.Sprintf("  %s %s:%d:%d\n", diagnostic.Code, diagnostic.File, diagnostic.Line, diagnostic.Column))
	}
	return builder.String()
}

func EncryptedSourceMode() EncryptedSourceResult {
	return EncryptedSourceResult{
		Success:          true,
		Command:          "security encrypted-source",
		Version:          version,
		Status:           "mvp-implemented",
		Extension:        ".black.enc",
		ProtectedFiles:   []string{".black", ".black.enc", "blacklang.toml"},
		ProductionPolicy: "Production packages exclude .black and .black.enc source files by default.",
		BuildPolicy:      "Draft v0.2 can decrypt .black.enc source in memory for validate, inspect, lint, benchmark, security scan, and build flows without writing plaintext source files.",
		KeyPolicy:        "MVP encryption derives an AES-256-GCM key from an environment variable such as BLACKLANG_SOURCE_KEY. Keys must never be stored in .black source.",
		Rules: []string{
			"Treat .black source as protected source of truth.",
			"Never put secrets, passwords, API keys, tokens, or private keys in .black files.",
			"Use black security encrypt with a key environment variable to create .black.enc source.",
			"Use black security decrypt --stdout only when plaintext inspection is explicitly needed.",
			"Build production artifacts from trusted developer or CI machines.",
			"Deploy generated artifacts, not protected BlackLang source.",
			"Do not write decrypted source into production packages.",
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
