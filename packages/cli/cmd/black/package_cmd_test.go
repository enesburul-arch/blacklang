package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageProductionExcludesProtectedSourceAndLocalSecrets(t *testing.T) {
	dir := t.TempDir()
	sourceDir := filepath.Join(dir, "generated")
	outDir := filepath.Join(dir, "artifact")
	files := map[string]string{
		"package.json":       "{}",
		"src/server.ts":      "export {};",
		"dist/index.html":    `<html><script type="module" src="/assets/app.js"></script></html>`,
		"dist/assets/app.js": "console.log('app');",
		"dist/assets/old.js": "console.log('old');",
		"README.md":          "generated notes",
		".env":               "DATABASE_URL=secret",
		".env.example":       "DATABASE_URL=file:./dev.db",
		"app.black":          "app Secret",
		"src/generated/a.ts": "generated prisma",
		"node_modules/a.js":  "dependency",
		"dev.db":             "database",
		"smoke-test.db":      "database",
	}
	for name, content := range files {
		path := filepath.Join(sourceDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	result := PackageProduction(sourceDir, outDir)
	if !result.Success {
		t.Fatalf("expected package success, got %#v", result)
	}
	for _, included := range []string{"package.json", "src/server.ts", "dist/index.html", "dist/assets/app.js"} {
		if _, err := os.Stat(filepath.Join(outDir, filepath.FromSlash(included))); err != nil {
			t.Fatalf("expected %s to be included: %v", included, err)
		}
	}
	for _, excluded := range []string{"README.md", ".env", ".env.example", "app.black", "src/generated/a.ts", "node_modules/a.js", "dev.db", "smoke-test.db", "dist/assets/old.js"} {
		if _, err := os.Stat(filepath.Join(outDir, filepath.FromSlash(excluded))); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be excluded", excluded)
		}
	}
}
