package main

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Version string
	Target  string
	Source  string
	Out     string
	Theme   string
}

type LoadedProject struct {
	Config      Config
	SourcePath  string
	OutDir      string
	Program     Program
	Diagnostics []Diagnostic
}

func LoadConfig(root string) Config {
	config := Config{}
	path := filepath.Join(root, "blacklang.toml")
	content, err := os.ReadFile(path)
	if err != nil {
		return config
	}

	for _, line := range strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(trimConfigComment(line))
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"`)

		switch key {
		case "version":
			config.Version = value
		case "target":
			config.Target = value
		case "source":
			config.Source = value
		case "out":
			config.Out = value
		case "theme":
			config.Theme = value
		}
	}

	return config
}

func LoadProject(args []string) LoadedProject {
	config := LoadConfig(".")
	sourcePath := firstNonOptionArg(args)
	if sourcePath == "" {
		sourcePath = config.Source
	}
	if sourcePath == "" {
		sourcePath = "examples/warehouse/app.black"
	}

	outDir := optionValue(args, "--out")
	if outDir == "" {
		outDir = config.Out
	}
	if outDir == "" {
		outDir = "generated"
	}

	project := LoadedProject{
		Config:      config,
		SourcePath:  sourcePath,
		OutDir:      outDir,
		Diagnostics: []Diagnostic{},
	}

	source, err := os.ReadFile(sourcePath)
	if err != nil {
		project.Diagnostics = append(project.Diagnostics, Diagnostic{
			File:       sourcePath,
			Line:       0,
			Column:     0,
			Code:       "FILE_READ_ERROR",
			Message:    err.Error(),
			Suggestion: "Pass a readable .black file path or set source in blacklang.toml.",
		})
		return project
	}

	program, parseDiagnostics := Parse(sourcePath, string(source))
	validateDiagnostics := Validate(program)
	diagnostics := append(parseDiagnostics, validateDiagnostics...)
	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}
	project.Program = program
	project.Diagnostics = diagnostics
	return project
}

func (p LoadedProject) Summary() Summary {
	return Summary{
		App:      p.Program.App.Name,
		Entities: len(p.Program.Entities),
		Pages:    len(p.Program.Pages),
	}
}

func trimConfigComment(value string) string {
	inQuote := false
	escaped := false
	for index, char := range value {
		if escaped {
			escaped = false
			continue
		}
		if char == '\\' && inQuote {
			escaped = true
			continue
		}
		if char == '"' {
			inQuote = !inQuote
			continue
		}
		if !inQuote && char == '#' {
			return value[:index]
		}
	}
	return value
}

func (p LoadedProject) ConfigInfo() ConfigInfo {
	return ConfigInfo{
		LanguageVersion: p.Config.Version,
		Target:          p.Config.Target,
		Source:          p.SourcePath,
		Out:             p.OutDir,
		Theme:           p.Config.Theme,
	}
}
