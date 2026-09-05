package main

import (
	"os"
	"path/filepath"
)

func InitProject(root string) InitResult {
	result := InitResult{
		Success: true,
		Command: "init",
		Version: version,
		Root:    root,
		Files:   []GeneratedFile{},
		Errors:  []Diagnostic{},
	}

	files := []struct {
		path    string
		kind    string
		content string
	}{
		{"blacklang.toml", "config", defaultBlacklangTOML()},
		{filepath.Join("src", "app.black"), "source", defaultAppBlack()},
		{"AGENTS.md", "agent-guide", defaultAgentsMD()},
		{"BLACKLANG.md", "language-guide", defaultBlacklangMD()},
	}

	for _, file := range files {
		fullPath := filepath.Join(root, file.path)
		if exists(fullPath) {
			result.Success = false
			result.Errors = append(result.Errors, Diagnostic{
				File:       fullPath,
				Code:       "INIT_FILE_EXISTS",
				Message:    "Refusing to overwrite existing file.",
				Suggestion: "Run init in an empty directory or move the existing file first.",
			})
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, Diagnostic{
				File:       fullPath,
				Code:       "INIT_DIRECTORY_ERROR",
				Message:    err.Error(),
				Suggestion: "Choose a writable project directory.",
			})
			continue
		}
		if err := os.WriteFile(fullPath, []byte(file.content), 0644); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, Diagnostic{
				File:       fullPath,
				Code:       "INIT_WRITE_ERROR",
				Message:    err.Error(),
				Suggestion: "Check file permissions and try again.",
			})
			continue
		}
		result.Files = append(result.Files, GeneratedFile{
			Path: fullPath,
			Kind: file.kind,
		})
	}

	return result
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func defaultBlacklangTOML() string {
	return `version = "0.1"
target = "web"
source = "src/app.black"
out = "generated"
`
}

func defaultAppBlack() string {
	return `app Warehouse

entity Product {
  sku text required unique
  name text required
  stock number default 0
  price money
}

page Products {
  source Product

  table {
    columns sku, name, stock, price
    search sku, name
  }

  form {
    fields sku, name, stock, price
  }

  actions create, edit, delete, archive, restore
}
`
}

func defaultAgentsMD() string {
	return `# AI Agent Instructions

1. Read blacklang.toml.
2. Read BLACKLANG.md.
3. Modify only .black source files.
4. Run black validate src/app.black --ir.
5. Run black build src/app.black --out generated --ir.
6. Do not manually edit generated files.
`
}

func defaultBlacklangMD() string {
	return `# BlackLang Project

This project uses BlackLang v0.1.

Source file:

- src/app.black

Generated output:

- generated/

BlackLang source is intended to be read and edited by humans and AI agents. Generated files should be recreated with black build instead of edited manually.
`
}
