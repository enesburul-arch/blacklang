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
