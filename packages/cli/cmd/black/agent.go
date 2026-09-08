package main

import (
	"fmt"
	"os"
	"path/filepath"
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
			"Run inspect --affected before changing high-impact entities, fields, queries, jobs, actions, pages, roles, workflows, states, components, APIs, deployment, or ops signals.",
			"Run migrate plan before deploying entity field, relation, uniqueness, default, required, or index changes to an existing database.",
			"Keep secrets, passwords, API keys, tokens, and private keys out of .black source files.",
			"Prefer environment references such as env DATABASE_URL for secret-aware source.",
		},
		Errors: project.Diagnostics,
	}
}

func agentReadFirst(source string, theme string) []AgentReadFile {
	root := agentReadRoot(source, theme)
	rootPath := func(path string) string {
		if root == "" {
			return path
		}
		return relativeAgentPath(filepath.Join(root, path))
	}
	files := []AgentReadFile{
		{Path: rootPath("AGENTS.md"), Purpose: "Local rules for AI agents working in this project."},
		{Path: rootPath("blacklang.toml"), Purpose: "Project language version, target, source path, and output path."},
		{Path: rootPath("BLACKLANG.md"), Purpose: "Compact project guide and current implemented behavior."},
		{Path: rootPath("SPEC.md"), Purpose: "Precise language and CLI behavior for the current draft."},
		{Path: rootPath(filepath.Join("docs", "diagnostics.md")), Purpose: "Stable diagnostic code reference and repair order."},
		{Path: readableAgentPath(source, root), Purpose: "Primary .black source of truth for the generated application."},
	}
	if theme != "" {
		files = append(files, AgentReadFile{Path: readableAgentPath(theme, root), Purpose: "BlackLang-native UI theme/profile source."})
	}
	for index := range files {
		files[index].Exists = fileExists(files[index].Path)
	}
	return files
}

func agentReadRoot(source string, theme string) string {
	for _, candidate := range []string{".", source, theme} {
		if root, ok := findAgentReadRoot(candidate); ok {
			return root
		}
	}
	return ""
}

func findAgentReadRoot(path string) (string, bool) {
	if strings.TrimSpace(path) == "" {
		return "", false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	if info, err := os.Stat(abs); err == nil && !info.IsDir() {
		abs = filepath.Dir(abs)
	}
	for {
		if fileExists(filepath.Join(abs, "blacklang.toml")) || (fileExists(filepath.Join(abs, "AGENTS.md")) && fileExists(filepath.Join(abs, "SPEC.md"))) {
			return abs, true
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return "", false
}

func readableAgentPath(path string, root string) string {
	if strings.TrimSpace(path) == "" || fileExists(path) || root == "" || filepath.IsAbs(path) {
		return filepath.ToSlash(filepath.Clean(path))
	}
	candidate := filepath.Join(root, path)
	if fileExists(candidate) {
		return relativeAgentPath(candidate)
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func relativeAgentPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.ToSlash(filepath.Clean(path))
	}
	cwd, err := os.Getwd()
	if err != nil {
		return filepath.ToSlash(filepath.Clean(path))
	}
	rel, err := filepath.Rel(cwd, abs)
	if err != nil {
		return filepath.ToSlash(filepath.Clean(path))
	}
	return filepath.ToSlash(filepath.Clean(rel))
}

func agentChecklist(source string, out string) []AgentChecklistItem {
	return []AgentChecklistItem{
		{Step: 1, Action: "Read readFirst files that exist.", Reason: "Learn local project rules before editing."},
		{Step: 2, Action: "Run black version --json.", Reason: "Confirm the installed CLI version."},
		{Step: 3, Action: "Run black docs --all --json when language context is missing.", Reason: "Load the compact local language reference."},
		{Step: 4, Action: "Run black ide --json when editor or completion metadata is needed.", Reason: "Load compiler-owned IDE snippets, completions, and diagnostic code metadata."},
		{Step: 5, Action: "Run black ecosystem --json when package, release, registry, adapter, marketplace, or extension metadata is needed.", Reason: "Load deterministic ecosystem discovery output."},
		{Step: 6, Action: fmt.Sprintf("Run black inspect %s --json.", quoteCLIArg(source)), Reason: "Learn the current app structure from the compiler."},
		{Step: 7, Action: fmt.Sprintf("Before high-impact edits, run black inspect %s --affected SymbolName --json.", quoteCLIArg(source)), Reason: "Ask the compiler what the edit can affect."},
		{Step: 8, Action: "Edit only .black source unless debugging the generator.", Reason: "Generated files should be recreated, not hand-maintained."},
		{Step: 9, Action: fmt.Sprintf("Run black format %s --check --json.", quoteCLIArg(source)), Reason: "Verify deterministic source style."},
		{Step: 10, Action: fmt.Sprintf("Run black ide diagnostics %s --json.", quoteCLIArg(source)), Reason: "Return editor-friendly parse, validate, format, and source-security diagnostics."},
		{Step: 11, Action: fmt.Sprintf("Run black lint %s --json.", quoteCLIArg(source)), Reason: "Check format, parse, validate, and source-security findings together."},
		{Step: 12, Action: fmt.Sprintf("Run black build %s --out %s --json.", quoteCLIArg(source), quoteCLIArg(out)), Reason: "Regenerate the target app and expose build diagnostics."},
		{Step: 13, Action: fmt.Sprintf("Run cd %s && npm test.", quoteCLIArg(out)), Reason: "Smoke test the generated OpenAPI, validation, API server, and target-specific generated test contract."},
		{Step: 14, Action: fmt.Sprintf("Run black benchmark %s --out %s --json.", quoteCLIArg(source), quoteCLIArg(out)), Reason: "Measure source and generated output size from the compiler."},
		{Step: 15, Action: fmt.Sprintf("Run black benchmark tasks %s --out %s --json.", quoteCLIArg(source), quoteCLIArg(out)), Reason: "Measure deterministic AI task scenarios and token estimate signals."},
		{Step: 16, Action: fmt.Sprintf("Run black benchmark eval %s --out %s --json.", quoteCLIArg(source), quoteCLIArg(out)), Reason: "Load long-running AI eval cases, evidence requirements, scoring criteria, and validation commands."},
		{Step: 17, Action: "Run black benchmark coverage --json.", Reason: "Inspect the current measured web coverage matrix."},
		{Step: 18, Action: "Run black benchmark issues --json.", Reason: "Inspect compact tracked issue totals and current open follow-up work."},
	}
}

func agentCommands(source string, out string, theme string) []AgentCommand {
	commands := []AgentCommand{
		{Name: "version", Command: "black version --json", Purpose: "Check installed CLI version."},
		{Name: "docs", Command: "black docs --all --json", Purpose: "Read compact language reference."},
		{Name: "diagnostics", Command: "black docs diagnostics --json", Purpose: "Read stable diagnostic repair guidance."},
		{Name: "ide", Command: "black ide --json", Purpose: "Read compiler-owned IDE metadata, completions, snippets, and diagnostic codes."},
		{Name: "ide-diagnostics", Command: fmt.Sprintf("black ide diagnostics %s --json", quoteCLIArg(source)), Purpose: "Return editor-friendly diagnostics for the source file."},
		{Name: "ecosystem", Command: "black ecosystem --json", Purpose: "Read release, package, registry, adapter, marketplace, extension, and trust-policy discovery metadata."},
		{Name: "inspect", Command: fmt.Sprintf("black inspect %s --json", quoteCLIArg(source)), Purpose: "Inspect project structure."},
		{Name: "format", Command: fmt.Sprintf("black format %s --check --json", quoteCLIArg(source)), Purpose: "Check deterministic formatting."},
		{Name: "lint", Command: fmt.Sprintf("black lint %s --json", quoteCLIArg(source)), Purpose: "Run read-only source checks."},
		{Name: "build", Command: fmt.Sprintf("black build %s --out %s --json", quoteCLIArg(source), quoteCLIArg(out)), Purpose: "Generate target application code."},
		{Name: "generated-test", Command: fmt.Sprintf("cd %s && npm test", quoteCLIArg(out)), Purpose: "Smoke test the generated OpenAPI, validation, API server, and target-specific test contract."},
		{Name: "benchmark", Command: fmt.Sprintf("black benchmark %s --out %s --json", quoteCLIArg(source), quoteCLIArg(out)), Purpose: "Measure source and generated output size."},
		{Name: "benchmark-tasks", Command: fmt.Sprintf("black benchmark tasks %s --out %s --json", quoteCLIArg(source), quoteCLIArg(out)), Purpose: "Report deterministic AI task scenarios and token estimate signals."},
		{Name: "benchmark-eval", Command: fmt.Sprintf("black benchmark eval %s --out %s --json", quoteCLIArg(source), quoteCLIArg(out)), Purpose: "Report long-running AI eval corpus cases, evidence requirements, scoring criteria, and validation commands."},
		{Name: "coverage", Command: "black benchmark coverage --json", Purpose: "Inspect the current web coverage matrix."},
		{Name: "coverage-issues", Command: "black benchmark issues --json", Purpose: "Export compact tracked issue totals and current open follow-up work."},
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
