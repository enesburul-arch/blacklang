package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecurityScanAllowsEnvDatabaseReference(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "app.black")
	source := `app Warehouse

database {
  url env DATABASE_URL
}
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	result := SecurityScanSource(sourcePath)
	if !result.Success {
		t.Fatalf("expected security scan success, got %#v", result)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("expected no findings, got %#v", result.Findings)
	}
}

func TestSecurityScanFindsHardcodedSecrets(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "app.black")
	source := `app Warehouse

database {
  url "postgres://user:password@example.com/app"
}

service Mail {
  apiKey "example_token_value_1234567890"
}
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	result := SecurityScanSource(sourcePath)
	if result.Success {
		t.Fatalf("expected security scan findings")
	}
	codes := map[string]bool{}
	for _, finding := range result.Findings {
		codes[finding.Code] = true
	}
	for _, code := range []string{"HARDCODED_DATABASE_URL", "HARDCODED_TOKEN"} {
		if !codes[code] {
			t.Fatalf("expected finding code %s, got %#v", code, result.Findings)
		}
	}
}

func TestEncryptedSourceModeDocumentsProtectedSource(t *testing.T) {
	result := EncryptedSourceMode()
	if !result.Success {
		t.Fatalf("expected encrypted source mode result success, got %#v", result)
	}
	if result.Status != "mvp-implemented" {
		t.Fatalf("expected mvp-implemented status, got %#v", result.Status)
	}
	if result.Extension != ".black.enc" {
		t.Fatalf("expected .black.enc extension, got %#v", result)
	}
	if result.ProductionPolicy == "" || result.BuildPolicy == "" || len(result.Rules) == 0 {
		t.Fatalf("expected encrypted source policy details, got %#v", result)
	}
	ir := FormatEncryptedSourceIR(result)
	if !strings.Contains(ir, "security encrypted-source") || !strings.Contains(ir, "extension .black.enc") {
		t.Fatalf("expected encrypted source IR details, got:\n%s", ir)
	}
}

func TestSecurityEncryptDecryptRoundTrip(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "app.black")
	encryptedPath := filepath.Join(dir, "app.black.enc")
	source := formattedValidWarehouseSource(t)
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}
	t.Setenv("BLACKLANG_TEST_SOURCE_KEY", "round-trip-test-key")

	encrypt := EncryptBlackSourceFile(sourcePath, encryptedPath, "BLACKLANG_TEST_SOURCE_KEY")
	if !encrypt.Success {
		t.Fatalf("expected encrypt success, got %#v", encrypt)
	}
	if encrypt.OutFile != encryptedPath || encrypt.KeyEnv != "BLACKLANG_TEST_SOURCE_KEY" {
		t.Fatalf("expected explicit out/key env, got %#v", encrypt)
	}
	if encrypt.Header.Alg != encryptedSourceAlg || encrypt.Header.KDF != encryptedSourceKDF || encrypt.Header.NonceBase64 == "" {
		t.Fatalf("expected deterministic encrypted header metadata, got %#v", encrypt.Header)
	}
	encryptedBytes, err := os.ReadFile(encryptedPath)
	if err != nil {
		t.Fatalf("expected encrypted output file: %v", err)
	}
	encryptedText := string(encryptedBytes)
	if !strings.Contains(encryptedText, encryptedSourceMagic) {
		t.Fatalf("expected encrypted source header, got:\n%s", encryptedText)
	}
	if strings.Contains(encryptedText, "app Warehouse") || strings.Contains(encryptedText, "entity Product") {
		t.Fatalf("encrypted source should not contain plaintext source:\n%s", encryptedText)
	}

	plaintext, decrypt := DecryptBlackSourceFile(encryptedPath, "")
	if !decrypt.Success {
		t.Fatalf("expected decrypt success, got %#v", decrypt)
	}
	if plaintext != source {
		t.Fatalf("expected plaintext round-trip")
	}
	if !strings.Contains(FormatSecurityEncryptIR(encrypt), "security encrypt ok") {
		t.Fatalf("expected encrypt IR success")
	}
	if !strings.Contains(FormatSecurityDecryptIR(decrypt), "security decrypt ok") {
		t.Fatalf("expected decrypt IR success")
	}
}

func TestEncryptedSourceCanLoadLintAndScanWithoutPlaintextFile(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "app.black")
	encryptedPath := filepath.Join(dir, "app.black.enc")
	source := formattedValidWarehouseSource(t)
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}
	t.Setenv("BLACKLANG_TEST_SOURCE_KEY", "load-lint-scan-test-key")
	encrypt := EncryptBlackSourceFile(sourcePath, encryptedPath, "BLACKLANG_TEST_SOURCE_KEY")
	if !encrypt.Success {
		t.Fatalf("expected encrypt success, got %#v", encrypt)
	}
	if err := os.Remove(sourcePath); err != nil {
		t.Fatalf("failed to remove plaintext source: %v", err)
	}

	project := LoadProject([]string{encryptedPath})
	if len(project.Diagnostics) != 0 {
		t.Fatalf("expected encrypted project load success, got %#v", project.Diagnostics)
	}
	if project.Summary().App != "Warehouse" || project.Summary().Entities != 1 || project.Summary().Pages != 1 {
		t.Fatalf("expected loaded encrypted project summary, got %#v", project.Summary())
	}

	lint := LintFile(encryptedPath)
	if !lint.Success {
		t.Fatalf("expected encrypted source lint success, got %#v", lint)
	}
	scan := SecurityScanSource(encryptedPath)
	if !scan.Success {
		t.Fatalf("expected encrypted source security scan success, got %#v", scan)
	}
}

func TestEncryptedSourceRequiresKey(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "app.black")
	encryptedPath := filepath.Join(dir, "app.black.enc")
	if err := os.WriteFile(sourcePath, []byte(formattedValidWarehouseSource(t)), 0o644); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	missing := EncryptBlackSourceFile(sourcePath, encryptedPath, "BLACKLANG_MISSING_SOURCE_KEY")
	if missing.Success || len(missing.Errors) != 1 || missing.Errors[0].Code != "MISSING_ENCRYPTION_KEY" {
		t.Fatalf("expected missing encryption key diagnostic, got %#v", missing)
	}

	t.Setenv("BLACKLANG_TEST_SOURCE_KEY", "available-key")
	encrypt := EncryptBlackSourceFile(sourcePath, encryptedPath, "BLACKLANG_TEST_SOURCE_KEY")
	if !encrypt.Success {
		t.Fatalf("expected encrypt success, got %#v", encrypt)
	}
	if err := os.Unsetenv("BLACKLANG_TEST_SOURCE_KEY"); err != nil {
		t.Fatalf("failed to unset key: %v", err)
	}
	_, decrypt := DecryptBlackSourceFile(encryptedPath, "")
	if decrypt.Success || len(decrypt.Errors) != 1 || decrypt.Errors[0].Code != "MISSING_ENCRYPTION_KEY" {
		t.Fatalf("expected missing decrypt key diagnostic, got %#v", decrypt)
	}
}

func TestFormatBlackFileRejectsEncryptedSource(t *testing.T) {
	dir := t.TempDir()
	encryptedPath := filepath.Join(dir, "app.black.enc")
	if err := os.WriteFile(encryptedPath, []byte("not plaintext"), 0o600); err != nil {
		t.Fatalf("failed to write encrypted placeholder: %v", err)
	}

	result, formatted := FormatBlackFile(encryptedPath, false, true)
	if result.Success || formatted != "" {
		t.Fatalf("expected encrypted format failure, got %#v and %q", result, formatted)
	}
	if len(result.Errors) != 1 || result.Errors[0].Code != "UNSUPPORTED_FORMAT_ENCRYPTED_SOURCE" {
		t.Fatalf("expected unsupported encrypted format diagnostic, got %#v", result)
	}
}

func formattedValidWarehouseSource(t *testing.T) string {
	t.Helper()
	source := `app Warehouse

database {
  url env DATABASE_URL
}

entity Product {
  sku text required
}

page Products {
  source Product

  table {
    columns sku
  }

  form {
    fields sku
  }

  actions create
}
`
	formatted, diagnostics := FormatBlackSource("app.black", source)
	if len(diagnostics) != 0 {
		t.Fatalf("expected test source to format cleanly, got %#v", diagnostics)
	}
	return formatted
}
