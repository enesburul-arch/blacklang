package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const version = "0.1.0-dev"

type response struct {
	Success bool   `json:"success"`
	Command string `json:"command"`
	Version string `json:"version"`
	Message string `json:"message"`
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		printHelp()
		return
	}

	switch args[0] {
	case "--version", "version":
		runVersion(args[1:])
	case "init":
		runInit(args[1:])
	case "parse":
		runParse(args[1:])
	case "format":
		runFormat(args[1:])
	case "lint":
		runLint(args[1:])
	case "validate":
		runValidate(args[1:])
	case "build":
		runBuild(args[1:])
	case "inspect":
		runInspect(args[1:])
	case "docs":
		runDocs(args[1:])
	case "explain":
		runExplain(args[1:])
	case "agent":
		runAgent(args[1:])
	case "theme":
		runTheme(args[1:])
	case "security":
		runSecurity(args[1:])
	case "package":
		runPackage(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printHelp()
		os.Exit(1)
	}
}

func runVersion(args []string) {
	result := VersionInfo()
	if hasJSONFlag(args) {
		printJSON(result)
		return
	}
	fmt.Println(result.Version)
}

func VersionInfo() VersionResult {
	return VersionResult{
		Success: true,
		Command: "version",
		Name:    "black",
		Version: version,
		Errors:  []Diagnostic{},
	}
}

func runFormat(args []string) {
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printFormatHelp()
		return
	}

	jsonOutput := hasJSONFlag(args)
	check := hasFlag(args, "--check")
	stdout := hasFlag(args, "--stdout")
	config := LoadConfig(".")
	file := firstNonOptionArg(args)
	if file == "" {
		file = config.Source
	}
	if file == "" {
		file = "examples/warehouse/app.black"
	}

	result, formatted := FormatBlackFile(file, !check && !stdout, check)
	result.Stdout = stdout
	if jsonOutput {
		printJSON(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}

	if !result.Success {
		printFormatResult(result)
		os.Exit(1)
	}
	if stdout {
		fmt.Print(formatted)
		return
	}
	printFormatResult(result)
}

func printFormatHelp() {
	fmt.Println(`BlackLang formatter

Usage:
  black format [file] [options]

Options:
  --check     Check formatting without writing
  --stdout    Print formatted source without writing
  --json      Print machine-readable JSON`)
}

func runLint(args []string) {
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printLintHelp()
		return
	}

	jsonOutput := hasJSONFlag(args)
	config := LoadConfig(".")
	file := firstNonOptionArg(args)
	if file == "" {
		file = config.Source
	}
	if file == "" {
		file = "examples/warehouse/app.black"
	}

	result := LintFile(file)
	if jsonOutput {
		printJSON(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}

	printLintResult(result)
	if !result.Success {
		os.Exit(1)
	}
}

func printLintHelp() {
	fmt.Println(`BlackLang linter

Usage:
  black lint [file] [options]

Options:
  --json      Print machine-readable JSON`)
}

func printHelp() {
	fmt.Println(`BlackLang CLI

Usage:
  black <command> [options]

Commands:
  init       Create a new BlackLang project scaffold
  parse       Parse .black source files
  format      Format .black source files deterministically
  lint        Check source formatting, syntax, semantics, and source security
  validate    Validate .black project semantics
  build       Generate target application code
  inspect     Print project summary or affected graph for humans or AI agents
  docs        Print compact language docs for one keyword or all keywords
  explain     Print action-oriented docs for one keyword
  agent       Print AI agent startup checklist output
  theme       Inspect .blackthm UI theme profiles
  security    Scan BlackLang source for source-security risks
  package     Create deployable artifacts without protected source files
  version     Print CLI version

Options:
  --json      Print machine-readable JSON when supported
  --ir        Print compact BlackIR when supported
  --help      Show this help message`)
}

func runSecurity(args []string) {
	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	if len(args) == 0 {
		diagnostic := Diagnostic{
			Code:       "UNKNOWN_SECURITY_COMMAND",
			Message:    "Security command must be `security scan` or `security encrypted-source`.",
			Suggestion: "Use `black security scan --json` or `black security encrypted-source --json`.",
		}
		printCommandError("security", jsonOutput, diagnostic)
		os.Exit(1)
	}
	if args[0] == "encrypted-source" {
		result := EncryptedSourceMode()
		if irOutput {
			fmt.Print(FormatEncryptedSourceIR(result))
			return
		}
		if jsonOutput {
			printJSON(result)
			return
		}
		fmt.Printf("encrypted source mode %s %s\n", result.Status, result.Extension)
		return
	}
	if args[0] != "scan" {
		diagnostic := Diagnostic{
			Code:       "UNKNOWN_SECURITY_COMMAND",
			Message:    "Security command must be `security scan` or `security encrypted-source`.",
			Suggestion: "Use `black security scan --json` or `black security encrypted-source --json`.",
		}
		printCommandError("security", jsonOutput, diagnostic)
		os.Exit(1)
	}
	config := LoadConfig(".")
	file := firstNonOptionArg(args[1:])
	if file == "" {
		file = config.Source
	}
	if file == "" {
		file = "examples/warehouse/app.black"
	}

	result := SecurityScanSource(file)
	if irOutput {
		fmt.Print(FormatSecurityScanIR(result))
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if jsonOutput {
		printJSON(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		for _, finding := range result.Findings {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", finding.File, finding.Line, finding.Column, finding.Code, finding.Message)
			if finding.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", finding.Suggestion)
			}
		}
		os.Exit(1)
	}
	fmt.Printf("security scan ok %s\n", result.File)
}

func runPackage(args []string) {
	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	if !hasFlag(args, "--production") {
		diagnostic := Diagnostic{
			Code:       "MISSING_PACKAGE_MODE",
			Message:    "Package command requires --production in v0.1.",
			Suggestion: "Use `black package --production`.",
		}
		printCommandError("package", jsonOutput, diagnostic)
		os.Exit(1)
	}

	project := LoadProject(args)
	if len(project.Diagnostics) > 0 {
		result := PackageResult{
			Success: false,
			Command: "package",
			Version: version,
			Mode:    "production",
			OutDir:  optionValue(args, "--out"),
			Files:   []GeneratedFile{},
			Errors:  project.Diagnostics,
		}
		if irOutput {
			printDiagnosticsIR("package", result.Errors)
		} else if jsonOutput {
			printJSON(result)
		} else {
			for _, diagnostic := range result.Errors {
				fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			}
		}
		os.Exit(1)
	}

	outDir := optionValue(args, "--out")
	result := PackageProduction(project.OutDir, outDir)
	if irOutput {
		fmt.Print(FormatPackageIR(result))
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if jsonOutput {
		printJSON(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s: %s\n", diagnostic.Code, diagnostic.Message)
		}
		os.Exit(1)
	}
	fmt.Printf("packaged %s\n", result.OutDir)
	for _, file := range result.Files {
		fmt.Printf("included %s\n", file.Path)
	}
}

func printPlaceholder(command string, jsonOutput bool) {
	message := fmt.Sprintf("%s is not implemented yet", command)
	if jsonOutput {
		payload := response{
			Success: false,
			Command: command,
			Version: version,
			Message: message,
		}
		encoded, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to encode JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(encoded))
		return
	}

	fmt.Println(message)
}

func hasJSONFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--json" {
			return true
		}
	}
	return false
}

func hasIRFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--ir" || arg == "--blackir" {
			return true
		}
	}
	return false
}

func hasFlag(args []string, name string) bool {
	for _, arg := range args {
		if arg == name {
			return true
		}
	}
	return false
}

func runParse(args []string) {
	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	config := LoadConfig(".")
	file := firstNonOptionArg(args)
	if file == "" {
		file = config.Source
	}
	if file == "" {
		file = "examples/warehouse/app.black"
	}

	source, err := os.ReadFile(file)
	if err != nil {
		printCommandError("parse", jsonOutput, Diagnostic{
			File:       file,
			Line:       0,
			Column:     0,
			Code:       "FILE_READ_ERROR",
			Message:    err.Error(),
			Suggestion: "Check that the file exists and is readable.",
		})
		os.Exit(1)
	}

	program, diagnostics := Parse(file, string(source))
	if len(diagnostics) > 0 {
		if irOutput {
			printDiagnosticsIR("parse", diagnostics)
			os.Exit(1)
		}
		printParseResult(jsonOutput, ParseResult{
			Success: false,
			Command: "parse",
			Version: version,
			File:    file,
			Program: program,
			Errors:  diagnostics,
		})
		os.Exit(1)
	}

	if irOutput {
		fmt.Print(FormatBlackIR(program))
		return
	}
	printParseResult(jsonOutput, ParseResult{
		Success: true,
		Command: "parse",
		Version: version,
		File:    file,
		Program: program,
		Errors:  []Diagnostic{},
	})
}

func runInit(args []string) {
	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	root := firstNonOptionArg(args)
	if root == "" {
		root = "."
	}

	result := InitProject(root)
	if irOutput {
		if result.Success {
			fmt.Print(FormatInitIR(result))
		} else {
			printDiagnosticsIR("init", result.Errors)
		}
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if jsonOutput {
		printJSON(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s: %s\n", diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		os.Exit(1)
	}

	fmt.Printf("initialized BlackLang project in %s\n", result.Root)
	for _, file := range result.Files {
		fmt.Printf("created %s\n", file.Path)
	}
}

func runValidate(args []string) {
	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	project := LoadProject(args)
	result := ValidateResult{
		Success: len(project.Diagnostics) == 0,
		Command: "validate",
		Version: version,
		File:    project.SourcePath,
		Summary: project.Summary(),
		Errors:  project.Diagnostics,
	}

	if irOutput {
		if result.Success {
			fmt.Print(FormatValidationIR(result))
		} else {
			printDiagnosticsIR("validate", result.Errors)
		}
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	printValidateResult(jsonOutput, result)
	if !result.Success {
		os.Exit(1)
	}
}

func runBuild(args []string) {
	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	project := LoadProject(args)
	if len(project.Diagnostics) > 0 {
		if irOutput {
			printDiagnosticsIR("build", project.Diagnostics)
			os.Exit(1)
		}
		printBuildResult(jsonOutput, BuildResult{
			Success: false,
			Command: "build",
			Version: version,
			File:    project.SourcePath,
			OutDir:  project.OutDir,
			Summary: project.Summary(),
			Errors:  project.Diagnostics,
		})
		os.Exit(1)
	}

	var theme *ThemeDecl
	if project.Config.Theme != "" {
		loadedTheme, themeDiagnostics := LoadTheme(project.Config.Theme)
		if len(themeDiagnostics) > 0 {
			if irOutput {
				printDiagnosticsIR("build", themeDiagnostics)
				os.Exit(1)
			}
			printBuildResult(jsonOutput, BuildResult{
				Success: false,
				Command: "build",
				Version: version,
				File:    project.SourcePath,
				OutDir:  project.OutDir,
				Summary: project.Summary(),
				Errors:  themeDiagnostics,
			})
			os.Exit(1)
		}
		theme = &loadedTheme
	}

	files, buildDiagnostics := BuildWebWithTheme(project.Program, project.OutDir, theme)
	if buildDiagnostics == nil {
		buildDiagnostics = []Diagnostic{}
	}
	result := BuildResult{
		Success: len(buildDiagnostics) == 0,
		Command: "build",
		Version: version,
		File:    project.SourcePath,
		OutDir:  project.OutDir,
		Summary: project.Summary(),
		Files:   files,
		Errors:  buildDiagnostics,
	}

	if irOutput {
		if result.Success {
			fmt.Print(FormatBuildIR(result))
		} else {
			printDiagnosticsIR("build", result.Errors)
		}
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	printBuildResult(jsonOutput, result)
	if !result.Success {
		os.Exit(1)
	}
}

func runInspect(args []string) {
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printInspectHelp()
		return
	}

	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	affectedRequested := hasFlag(args, "--affected")
	project := LoadProject(args)

	if affectedRequested {
		symbol := optionValue(args, "--affected")
		affected := newAffectedAnalysis(symbol)
		diagnostics := append([]Diagnostic{}, project.Diagnostics...)
		if len(diagnostics) == 0 {
			var affectedDiagnostics []Diagnostic
			affected, affectedDiagnostics = AnalyzeAffected(project.Program, symbol)
			diagnostics = append(diagnostics, affectedDiagnostics...)
		}
		result := InspectAffectedResult{
			Success:  len(diagnostics) == 0,
			Command:  "inspect",
			Version:  version,
			Config:   project.ConfigInfo(),
			Summary:  project.Summary(),
			Affected: affected,
			Errors:   diagnostics,
		}
		if irOutput {
			if result.Success {
				fmt.Print(FormatAffectedIR(result))
			} else {
				printDiagnosticsIR("inspect", result.Errors)
			}
			if !result.Success {
				os.Exit(1)
			}
			return
		}
		if jsonOutput {
			printJSON(result)
			if !result.Success {
				os.Exit(1)
			}
			return
		}
		printAffectedResult(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}

	result := InspectResult{
		Success: len(project.Diagnostics) == 0,
		Command: "inspect",
		Version: version,
		Config:  project.ConfigInfo(),
		Summary: project.Summary(),
		Program: project.Program,
		Errors:  project.Diagnostics,
	}

	if irOutput {
		if result.Success {
			fmt.Print(FormatInspectIR(result))
		} else {
			printDiagnosticsIR("inspect", result.Errors)
		}
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if jsonOutput {
		printJSON(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
		}
		os.Exit(1)
	}

	fmt.Printf("app: %s\n", result.Summary.App)
	fmt.Printf("source: %s\n", result.Config.Source)
	fmt.Printf("out: %s\n", result.Config.Out)
	fmt.Printf("entities: %d\n", result.Summary.Entities)
	fmt.Printf("pages: %d\n", result.Summary.Pages)
}

func printInspectHelp() {
	fmt.Println(`BlackLang inspect

Usage:
  black inspect [file] [options]

Options:
  --affected <symbol>  Print affected graph for an entity, field, page, role, workflow, state, component, api, target, or deploy
  --json               Print machine-readable JSON
  --ir                 Print compact BlackIR`)
}

func runDocs(args []string) {
	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	if hasFlag(args, "--all") {
		docs := AllDocs()
		result := DocsAllResult{
			Success: true,
			Command: "docs",
			Version: version,
			Count:   len(docs),
			Docs:    docs,
			Errors:  []Diagnostic{},
		}
		if irOutput {
			fmt.Print(FormatDocsAllIR(result))
			return
		}
		if jsonOutput {
			printJSON(result)
			return
		}
		fmt.Printf("docs %d\n", result.Count)
		for _, doc := range result.Docs {
			fmt.Printf("%s - %s\n", doc.Keyword, doc.Purpose)
		}
		return
	}

	keyword := firstNonOptionArg(args)
	if keyword == "" {
		keyword = "syntax"
	}

	doc, ok := FindDoc(keyword)
	result := DocsResult{
		Success: ok,
		Command: "docs",
		Version: version,
		Doc:     doc,
		Errors:  []Diagnostic{},
	}
	if !ok {
		result.Errors = []Diagnostic{{
			Code:       "UNKNOWN_DOC_KEYWORD",
			Message:    fmt.Sprintf("No docs entry exists for %q.", keyword),
			Suggestion: "Use syntax, version, docs, explain, agent, diagnostics, format, lint, app, target, auth, role, access, entity, layout, page, view, table, form, actions, ui, ui-profile, ui-modes, search, filter, paginate, workflow, state, component, blackir, openapi, package, security, cors, deploy, audit, or csrf.",
		}}
	}

	if irOutput {
		if result.Success {
			fmt.Print(FormatDocsIR(result))
		} else {
			printDiagnosticsIR("docs", result.Errors)
		}
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if jsonOutput {
		printJSON(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s: %s\n", diagnostic.Code, diagnostic.Message)
		}
		os.Exit(1)
	}
	fmt.Printf("%s\n\n%s\n\nSyntax:\n%s\n\nExample:\n%s\n", doc.Keyword, doc.Purpose, doc.Syntax, doc.Example)
}

func runExplain(args []string) {
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printExplainHelp()
		return
	}

	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	keyword := firstNonOptionArg(args)
	result := ExplainKeyword(keyword)

	if irOutput {
		if result.Success {
			fmt.Print(FormatExplainIR(result))
		} else {
			printDiagnosticsIR("explain", result.Errors)
		}
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if jsonOutput {
		printJSON(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}

	printExplainResult(result)
	if !result.Success {
		os.Exit(1)
	}
}

func runAgent(args []string) {
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printAgentHelp()
		return
	}

	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	startupArgs := args
	if len(args) > 0 && args[0] == "startup" {
		startupArgs = args[1:]
	} else if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		diagnostic := Diagnostic{
			Code:       "UNKNOWN_AGENT_COMMAND",
			Message:    fmt.Sprintf("No agent command exists for %q.", args[0]),
			Suggestion: "Use `black agent startup --json`.",
		}
		printCommandError("agent", jsonOutput, diagnostic)
		os.Exit(1)
	}

	result := AgentStartupChecklist(startupArgs)
	if irOutput {
		if result.Success {
			fmt.Print(FormatAgentStartupIR(result))
		} else {
			printDiagnosticsIR("agent startup", result.Errors)
		}
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if jsonOutput {
		printJSON(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}

	printAgentStartupResult(result)
	if !result.Success {
		os.Exit(1)
	}
}

func runTheme(args []string) {
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printThemeHelp()
		return
	}

	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	inspectArgs := args
	if len(args) > 0 && args[0] == "inspect" {
		inspectArgs = args[1:]
	} else if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		diagnostic := Diagnostic{
			Code:       "UNKNOWN_THEME_COMMAND",
			Message:    fmt.Sprintf("No theme command exists for %q.", args[0]),
			Suggestion: "Use `black theme inspect --json`.",
		}
		printCommandError("theme", jsonOutput, diagnostic)
		os.Exit(1)
	}

	result := InspectTheme(inspectArgs)
	if irOutput {
		if result.Success {
			fmt.Print(FormatThemeIR(result))
		} else {
			printDiagnosticsIR("theme inspect", result.Errors)
		}
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if jsonOutput {
		printJSON(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	printThemeInspectResult(result)
	if !result.Success {
		os.Exit(1)
	}
}

func printThemeHelp() {
	fmt.Println(`BlackLang theme

Usage:
  black theme inspect [file] [options]
  black theme [options]

Options:
  --json      Print machine-readable JSON
  --ir        Print compact BlackIR`)
}

func printAgentHelp() {
	fmt.Println(`BlackLang agent

Usage:
  black agent startup [file] [options]
  black agent [options]

Options:
  --json      Print machine-readable JSON
  --ir        Print compact BlackIR`)
}

func printExplainHelp() {
	fmt.Println(`BlackLang explain

Usage:
  black explain <keyword> [options]

Options:
  --json      Print machine-readable JSON
  --ir        Print compact BlackIR`)
}

func firstNonOptionArg(args []string) string {
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--out" || arg == "--affected" {
			index++
			continue
		}
		if len(arg) > 0 && arg[0] != '-' {
			return arg
		}
	}
	return ""
}

func optionValue(args []string, name string) string {
	for index, arg := range args {
		if arg == name && index+1 < len(args) {
			return args[index+1]
		}
	}
	return ""
}

func printCommandError(command string, jsonOutput bool, diagnostic Diagnostic) {
	if jsonOutput {
		printJSON(ParseResult{
			Success: false,
			Command: command,
			Version: version,
			Errors:  []Diagnostic{diagnostic},
		})
		return
	}

	fmt.Fprintf(os.Stderr, "%s: %s\n", diagnostic.Code, diagnostic.Message)
	if diagnostic.Suggestion != "" {
		fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
	}
}

func printFormatResult(result FormatResult) {
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}

	if result.Changed {
		fmt.Printf("formatted %s\n", result.File)
		return
	}
	fmt.Printf("format ok %s\n", result.File)
}

func printLintResult(result LintResult) {
	if result.Success {
		fmt.Printf("lint ok %s\n", result.File)
		return
	}

	for _, check := range result.Checks {
		status := "ok"
		if !check.Success {
			status = "fail"
		}
		fmt.Fprintf(os.Stderr, "%s %s findings=%d\n", status, check.Name, check.Findings)
	}
	for _, diagnostic := range result.Findings {
		fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
		if diagnostic.Suggestion != "" {
			fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
		}
	}
	for _, diagnostic := range result.Errors {
		fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
		if diagnostic.Suggestion != "" {
			fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
		}
	}
}

func printExplainResult(result ExplainResult) {
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s: %s\n", diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}

	fmt.Printf("%s\n\n%s\n\nSyntax:\n%s\n\nExample:\n%s\n", result.Keyword, result.Purpose, result.Syntax, result.Example)
	if len(result.AgentSteps) > 0 {
		fmt.Println("\nAgent steps:")
		for _, step := range result.AgentSteps {
			fmt.Printf("- %s\n", step)
		}
	}
	if len(result.Related) > 0 {
		fmt.Printf("\nRelated: %s\n", strings.Join(result.Related, ", "))
	}
}

func printParseResult(jsonOutput bool, result ParseResult) {
	if jsonOutput {
		printJSON(result)
		return
	}

	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}

	fmt.Printf("parsed %s\n", result.File)
	fmt.Printf("app: %s\n", result.Program.App.Name)
	fmt.Printf("entities: %d\n", len(result.Program.Entities))
	fmt.Printf("pages: %d\n", len(result.Program.Pages))
}

func printValidateResult(jsonOutput bool, result ValidateResult) {
	if jsonOutput {
		printJSON(result)
		return
	}

	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}

	fmt.Printf("valid %s\n", result.File)
	fmt.Printf("app: %s\n", result.Summary.App)
	fmt.Printf("entities: %d\n", result.Summary.Entities)
	fmt.Printf("pages: %d\n", result.Summary.Pages)
}

func printBuildResult(jsonOutput bool, result BuildResult) {
	if jsonOutput {
		printJSON(result)
		return
	}

	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}

	fmt.Printf("built %s\n", result.OutDir)
	for _, file := range result.Files {
		fmt.Printf("created %s\n", file.Path)
	}
}

func printAffectedResult(result InspectAffectedResult) {
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}

	affected := result.Affected
	fmt.Printf("affected %s (%s)\n", affected.Symbol, affected.Kind)
	printAffectedItems("entities", affected.Entities)
	printAffectedItems("pages", affected.Pages)
	printAffectedItems("roles", affected.Roles)
	printAffectedItems("workflows", affected.Workflows)
	printAffectedItems("states", affected.States)
	printAffectedItems("components", affected.Components)
	printAffectedItems("apis", affected.APIs)
	printAffectedItems("generated files", affected.GeneratedFiles)
}

func printAffectedItems(label string, items []AffectedItem) {
	if len(items) == 0 {
		return
	}
	fmt.Printf("%s:\n", label)
	for _, item := range items {
		if item.Reason == "" {
			fmt.Printf("- %s\n", item.Name)
			continue
		}
		fmt.Printf("- %s: %s\n", item.Name, item.Reason)
	}
}

func printAgentStartupResult(result AgentStartupResult) {
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}

	fmt.Println("agent startup ok")
	if result.Config.LanguageVersion != "" {
		fmt.Printf("language: %s\n", result.Config.LanguageVersion)
	}
	if result.Config.Target != "" {
		fmt.Printf("target: %s\n", result.Config.Target)
	}
	fmt.Printf("source: %s\n", result.Config.Source)
	fmt.Printf("out: %s\n", result.Config.Out)
	if result.Config.Theme != "" {
		fmt.Printf("theme: %s\n", result.Config.Theme)
	}
	fmt.Printf("app: %s\n", result.Summary.App)
	fmt.Println("read first:")
	for _, file := range result.ReadFirst {
		status := "missing"
		if file.Exists {
			status = "exists"
		}
		fmt.Printf("- %s (%s): %s\n", file.Path, status, file.Purpose)
	}
	fmt.Println("checklist:")
	for _, item := range result.Checklist {
		fmt.Printf("%d. %s\n", item.Step, item.Action)
	}
}

func printThemeInspectResult(result ThemeInspectResult) {
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}
	fmt.Printf("theme ok %s\n", result.File)
	fmt.Printf("name: %s\n", result.Theme.Name)
	fmt.Printf("version: %d\n", result.Theme.Version)
	fmt.Printf("target: %s\n", result.Theme.Target)
	fmt.Printf("locked: %t\n", result.Theme.Locked)
	fmt.Printf("tokens: %d\n", len(result.Theme.Tokens))
	fmt.Printf("profile: %s v%d\n", result.Theme.Profile.Name, result.Theme.Profile.Version)
	fmt.Printf("mode groups: %d\n", len(result.Theme.Profile.ModeGroups))
	fmt.Printf("baselines: %d\n", len(result.Theme.Profile.Baselines))
	fmt.Printf("rules: %s, missing trailing slots: %s, new locked slots: %s\n", result.Theme.Profile.Rules.SlotOrder, result.Theme.Profile.Rules.MissingTrailingSlots, result.Theme.Profile.Rules.NewSlotsAfterLock)
	for _, mode := range result.Theme.Profile.Modes {
		modeLabel := mode.Name
		if mode.Standard {
			modeLabel += " (standard)"
		}
		fmt.Printf("- mode %s: %s\n", modeLabel, strings.Join(mode.Slots, ", "))
	}
}

func printJSON(value any) {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(encoded))
}
