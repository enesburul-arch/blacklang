package main

import (
	"fmt"
	"os"
	"strings"
)

func AgentStartupChecklist(args []string) AgentStartupResult {
	project := LoadProject(args)
	source := project.SourcePath
	out := project.OutDir
	theme := project.Config.Theme

	return AgentStartupResult{
		Success:       len(project.Diagnostics) == 0,
		Command:       "agent startup",
		Version:       version,
		Config:        project.ConfigInfo(),
		Summary:       project.Summary(),
		ReadFirst:     agentReadFirst(source, theme),
		SourceFiles:   []string{source},
		ThemeFiles:    agentThemeFiles(theme),
		GeneratedDirs: []string{out},
		Checklist:     agentChecklist(source, out),
		Commands:      agentCommands(source, out, theme),
		Policies: []string{
			"Edit .black source files first; treat generated output as rebuildable compiler output.",
			"Treat .blackthm files as source assets for deterministic UI profile metadata.",
			"Read blacklang.toml before assuming source, target, or output paths.",
			"Use diagnostic code values instead of message text when repairing failures.",
			"Run inspect --affected before changing high-impact entities, fields, pages, roles, workflows, states, components, or APIs.",
			"Keep secrets, passwords, API keys, tokens, and private keys out of .black source files.",
			"Prefer environment references such as env DATABASE_URL for secret-aware source.",
		},
		Errors: project.Diagnostics,
	}
}

func agentReadFirst(source string, theme string) []AgentReadFile {
	files := []AgentReadFile{
		{Path: "AGENTS.md", Purpose: "Local rules for AI agents working in this project."},
		{Path: "blacklang.toml", Purpose: "Project language version, target, source path, and output path."},
		{Path: "BLACKLANG.md", Purpose: "Compact project guide and current implemented behavior."},
		{Path: "SPEC.md", Purpose: "Precise language and CLI behavior for the current draft."},
		{Path: "docs/diagnostics.md", Purpose: "Stable diagnostic code reference and repair order."},
		{Path: source, Purpose: "Primary .black source of truth for the generated application."},
	}
	if theme != "" {
		files = append(files, AgentReadFile{Path: theme, Purpose: "BlackLang-native UI theme/profile source."})
	}
	for index := range files {
		files[index].Exists = fileExists(files[index].Path)
	}
	return files
}

func agentChecklist(source string, out string) []AgentChecklistItem {
	return []AgentChecklistItem{
		{Step: 1, Action: "Read readFirst files that exist.", Reason: "Learn local project rules before editing."},
		{Step: 2, Action: "Run black version --json.", Reason: "Confirm the installed CLI version."},
		{Step: 3, Action: "Run black docs --all --json when language context is missing.", Reason: "Load the compact local language reference."},
		{Step: 4, Action: fmt.Sprintf("Run black inspect %s --json.", quoteCLIArg(source)), Reason: "Learn the current app structure from the compiler."},
		{Step: 5, Action: fmt.Sprintf("Before high-impact edits, run black inspect %s --affected SymbolName --json.", quoteCLIArg(source)), Reason: "Ask the compiler what the edit can affect."},
		{Step: 6, Action: "Edit only .black source unless debugging the generator.", Reason: "Generated files should be recreated, not hand-maintained."},
		{Step: 7, Action: fmt.Sprintf("Run black format %s --check --json.", quoteCLIArg(source)), Reason: "Verify deterministic source style."},
		{Step: 8, Action: fmt.Sprintf("Run black lint %s --json.", quoteCLIArg(source)), Reason: "Check format, parse, validate, and source-security findings together."},
		{Step: 9, Action: fmt.Sprintf("Run black build %s --out %s --json.", quoteCLIArg(source), quoteCLIArg(out)), Reason: "Regenerate the target app and expose build diagnostics."},
	}
}

func agentCommands(source string, out string, theme string) []AgentCommand {
	commands := []AgentCommand{
		{Name: "version", Command: "black version --json", Purpose: "Check installed CLI version."},
		{Name: "docs", Command: "black docs --all --json", Purpose: "Read compact language reference."},
		{Name: "diagnostics", Command: "black docs diagnostics --json", Purpose: "Read stable diagnostic repair guidance."},
		{Name: "inspect", Command: fmt.Sprintf("black inspect %s --json", quoteCLIArg(source)), Purpose: "Inspect project structure."},
		{Name: "format", Command: fmt.Sprintf("black format %s --check --json", quoteCLIArg(source)), Purpose: "Check deterministic formatting."},
		{Name: "lint", Command: fmt.Sprintf("black lint %s --json", quoteCLIArg(source)), Purpose: "Run read-only source checks."},
		{Name: "build", Command: fmt.Sprintf("black build %s --out %s --json", quoteCLIArg(source), quoteCLIArg(out)), Purpose: "Generate target application code."},
		{Name: "security", Command: fmt.Sprintf("black security scan %s --json", quoteCLIArg(source)), Purpose: "Scan .black source for likely hardcoded secrets."},
	}
	if theme != "" {
		commands = append(commands, AgentCommand{Name: "theme", Command: fmt.Sprintf("black theme inspect %s --json", quoteCLIArg(theme)), Purpose: "Inspect BlackLang-native UI theme/profile metadata."})
	}
	return commands
}

func agentThemeFiles(theme string) []string {
	if strings.TrimSpace(theme) == "" {
		return nil
	}
	return []string{theme}
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func quoteCLIArg(value string) string {
	if value == "" {
		return value
	}
	if !strings.ContainsAny(value, " \t\n\"") {
		return value
	}
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}
