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
	case "benchmark":
		runBenchmark(args[1:])
	case "ecosystem":
		runEcosystem(args[1:])
	case "migrate":
		runMigrate(args[1:])
	case "inspect":
		runInspect(args[1:])
	case "docs":
		runDocs(args[1:])
	case "explain":
		runExplain(args[1:])
	case "ide":
		runIDE(args[1:])
	case "agent":
		runAgent(args[1:])
	case "theme":
		runTheme(args[1:])
	case "audit":
		runAudit(args[1:])
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

func runAudit(args []string) {
	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	if len(args) == 0 || hasFlag(args, "--help") || hasFlag(args, "-h") {
		printAuditHelp()
		if len(args) == 0 {
			os.Exit(1)
		}
		return
	}
	if args[0] != "accessibility" {
		diagnostic := Diagnostic{
			Code:       "UNKNOWN_AUDIT_COMMAND",
			Message:    "Audit command must be `audit accessibility`.",
			Suggestion: "Use `black audit accessibility --json`.",
		}
		printCommandError("audit", jsonOutput, diagnostic)
		os.Exit(1)
	}

	commandArgs := args[1:]
	config := LoadConfig(".")
	file := firstNonOptionArg(commandArgs)
	if file == "" {
		file = config.Source
	}
	if file == "" {
		file = "examples/warehouse/app.black"
	}

	result := AccessibilityAuditFile(file)
	if irOutput {
		fmt.Print(FormatAccessibilityAuditIR(result))
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
	printAccessibilityAuditResult(result)
	if !result.Success {
		os.Exit(1)
	}
}

func printAuditHelp() {
	fmt.Println(`BlackLang audit

Usage:
  black audit accessibility [file] [options]

Options:
  --json      Print machine-readable JSON
  --ir        Print compact BlackIR`)
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
  benchmark   Measure source/generated size, AI task/eval estimates, web coverage, and tracked issues
  ecosystem   Print release, package, registry, adapter, marketplace, public index, and extension discovery metadata
  migrate     Plan safe, manual, and destructive schema changes
  inspect     Print project summary or affected graph for humans or AI agents
  docs        Print compact language docs for one keyword or all keywords
  explain     Print action-oriented docs for one keyword
  ide         Export IDE metadata and source diagnostics
  agent       Print AI agent startup checklist output
  theme       Inspect and migrate .blackthm UI theme profiles
  audit       Run read-only policy audits
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
	if len(args) == 0 || hasFlag(args, "--help") || hasFlag(args, "-h") {
		printSecurityHelp()
		if len(args) == 0 {
			os.Exit(1)
		}
		return
	}

	command := args[0]
	commandArgs := args[1:]
	switch command {
	case "encrypted-source":
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
	case "encrypt":
		runSecurityEncrypt(commandArgs, jsonOutput, irOutput)
	case "decrypt":
		runSecurityDecrypt(commandArgs, jsonOutput, irOutput)
	case "scan":
		runSecurityScan(commandArgs, jsonOutput, irOutput)
	default:
		diagnostic := Diagnostic{
			Code:       "UNKNOWN_SECURITY_COMMAND",
			Message:    "Security command must be `security scan`, `security encrypted-source`, `security encrypt`, or `security decrypt`.",
			Suggestion: "Use `black security scan --json`, `black security encrypted-source --json`, `black security encrypt app.black --json`, or `black security decrypt app.black.enc --stdout`.",
		}
		printCommandError("security", jsonOutput, diagnostic)
		os.Exit(1)
	}
}

func printSecurityHelp() {
	fmt.Println(`BlackLang security

Usage:
  black security scan [file] [--json|--ir]
  black security encrypted-source [--json|--ir]
  black security encrypt <file> [--out <file.black.enc>] [--key-env <ENV>] [--json|--ir]
  black security decrypt <file.black.enc> --stdout [--key-env <ENV>] [--json|--ir]

Options:
  --out <file>      Write encrypted source to this file
  --key-env <ENV>   Read encryption key material from this environment variable
  --stdout          Print decrypted plaintext instead of writing a file
  --json            Print machine-readable JSON
  --ir              Print compact BlackIR`)
}

func runSecurityEncrypt(args []string, jsonOutput bool, irOutput bool) {
	config := LoadConfig(".")
	file := firstNonOptionArg(args)
	if file == "" {
		file = config.Source
	}
	if file == "" {
		diagnostic := Diagnostic{
			Code:       "MISSING_SECURITY_SOURCE",
			Message:    "Security encrypt requires a .black source file.",
			Suggestion: "Use `black security encrypt app.black --out app.black.enc`.",
		}
		printCommandError("security encrypt", jsonOutput, diagnostic)
		os.Exit(1)
	}

	result := EncryptBlackSourceFile(file, optionValue(args, "--out"), optionValue(args, "--key-env"))
	if irOutput {
		fmt.Print(FormatSecurityEncryptIR(result))
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
		printDiagnostics(result.Errors)
		os.Exit(1)
	}
	fmt.Printf("encrypted %s -> %s\n", result.File, result.OutFile)
	fmt.Printf("key env %s\n", result.KeyEnv)
}

func runSecurityDecrypt(args []string, jsonOutput bool, irOutput bool) {
	stdout := hasFlag(args, "--stdout")
	file := firstNonOptionArg(args)
	if file == "" {
		diagnostic := Diagnostic{
			Code:       "MISSING_SECURITY_SOURCE",
			Message:    "Security decrypt requires a .black.enc source file.",
			Suggestion: "Use `black security decrypt app.black.enc --stdout`.",
		}
		printCommandError("security decrypt", jsonOutput, diagnostic)
		os.Exit(1)
	}
	if !stdout {
		result := SecurityDecryptResult{
			Success: false,
			Command: "security decrypt",
			Version: version,
			File:    file,
			KeyEnv:  optionValue(args, "--key-env"),
			Errors: []Diagnostic{{
				File:       file,
				Code:       "MISSING_DECRYPT_STDOUT",
				Message:    "Security decrypt requires --stdout so plaintext is never written to disk by default.",
				Suggestion: "Use `black security decrypt " + file + " --stdout` inside a trusted developer or CI environment.",
			}},
		}
		if irOutput {
			fmt.Print(FormatSecurityDecryptIR(result))
		} else if jsonOutput {
			printJSON(result)
		} else {
			printDiagnostics(result.Errors)
		}
		os.Exit(1)
	}

	plaintext, result := DecryptBlackSourceFile(file, optionValue(args, "--key-env"))
	if result.Success && jsonOutput {
		result.Plaintext = plaintext
	}
	if irOutput {
		fmt.Print(FormatSecurityDecryptIR(result))
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
		printDiagnostics(result.Errors)
		os.Exit(1)
	}
	fmt.Print(plaintext)
}

func runSecurityScan(args []string, jsonOutput bool, irOutput bool) {
	config := LoadConfig(".")
	file := firstNonOptionArg(args)
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
		printDiagnostics(result.Errors)
		printDiagnostics(result.Findings)
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

	source, readDiagnostics := ReadBlackSource(file)
	if len(readDiagnostics) > 0 {
		if irOutput {
			printDiagnosticsIR("parse", readDiagnostics)
		} else if jsonOutput {
			printJSON(ParseResult{
				Success: false,
				Command: "parse",
				Version: version,
				File:    file,
				Errors:  readDiagnostics,
			})
		} else {
			printDiagnostics(readDiagnostics)
		}
		os.Exit(1)
	}

	program, diagnostics := Parse(file, source)
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

func runBenchmark(args []string) {
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printBenchmarkHelp()
		return
	}

	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	if len(args) > 0 && args[0] == "coverage" {
		result := WebCoverageReport()
		if irOutput {
			fmt.Print(FormatCoverageIR(result))
			return
		}
		if jsonOutput {
			printJSON(result)
			return
		}
		printCoverageResult(result)
		return
	}
	if len(args) > 0 && args[0] == "issues" {
		result := WebCoverageIssuesReport()
		if irOutput {
			fmt.Print(FormatCoverageIssuesIR(result))
			return
		}
		if jsonOutput {
			printJSON(result)
			return
		}
		printCoverageIssuesResult(result)
		return
	}
	if len(args) > 0 && args[0] == "eval-history" {
		result := BenchmarkEvalHistory(args[1:])
		if irOutput {
			if result.Success {
				fmt.Print(FormatAIEvalHistoryIR(result))
			} else {
				printDiagnosticsIR("benchmark eval-history", result.Errors)
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
		printAIEvalHistoryResult(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if len(args) > 0 && args[0] == "tasks" {
		result := BenchmarkTasks(args[1:])
		if irOutput {
			if result.Success {
				fmt.Print(FormatAITaskBenchmarkIR(result))
			} else {
				printDiagnosticsIR("benchmark tasks", result.Errors)
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
		printAITaskBenchmarkResult(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if len(args) > 0 && args[0] == "eval" {
		result := BenchmarkEvalCorpus(args[1:])
		if irOutput {
			if result.Success {
				fmt.Print(FormatAIEvalCorpusIR(result))
			} else {
				printDiagnosticsIR("benchmark eval", result.Errors)
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
		printAIEvalCorpusResult(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	result := BenchmarkProject(args)
	if irOutput {
		if result.Success {
			fmt.Print(FormatBenchmarkIR(result))
		} else {
			printDiagnosticsIR("benchmark", result.Errors)
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
	printBenchmarkResult(result)
	if !result.Success {
		os.Exit(1)
	}
}

func runEcosystem(args []string) {
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printEcosystemHelp()
		return
	}

	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	result := EcosystemStatus()
	if irOutput {
		fmt.Print(FormatEcosystemIR(result))
		return
	}
	if jsonOutput {
		printJSON(result)
		return
	}
	printEcosystemResult(result)
}

func runMigrate(args []string) {
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printMigrateHelp()
		return
	}

	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		diagnostic := Diagnostic{
			Code:       "UNKNOWN_MIGRATE_COMMAND",
			Message:    "Migrate command must be `migrate plan`.",
			Suggestion: "Use `black migrate plan old.black new.black --json`.",
		}
		printCommandError("migrate", jsonOutput, diagnostic)
		os.Exit(1)
	}
	if args[0] != "plan" {
		diagnostic := Diagnostic{
			Code:       "UNKNOWN_MIGRATE_COMMAND",
			Message:    fmt.Sprintf("No migrate command exists for %q.", args[0]),
			Suggestion: "Use `black migrate plan old.black new.black --json`.",
		}
		printCommandError("migrate", jsonOutput, diagnostic)
		os.Exit(1)
	}

	result := MigratePlan(args[1:])
	if irOutput {
		fmt.Print(FormatSchemaMigrationPlanIR(result))
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
	printSchemaMigrationPlanResult(result)
	if !result.Success {
		os.Exit(1)
	}
}

func printMigrateHelp() {
	fmt.Println(`BlackLang migrate

Usage:
  black migrate plan <old.black> <new.black> [--json|--ir]

Options:
  --json      Print machine-readable JSON
  --ir        Print compact BlackIR`)
}

func printBenchmarkHelp() {
	fmt.Println(`BlackLang benchmark

Usage:
  black benchmark [file] [options]
  black benchmark coverage [--json|--ir]
  black benchmark issues [--json|--ir]
  black benchmark tasks [file] [--out <dir>] [--json|--ir]
  black benchmark eval [file] [--out <dir>] [--json|--ir]
  black benchmark eval-history [--history <file>] [--json|--ir]

Options:
  --out <dir>  Use the configured output directory label in reported generated paths
  --history    Read an eval history manifest instead of benchmarks/eval-history.blackdir
  --json       Print machine-readable JSON
  --ir         Print compact BlackIR`)
}

func printEcosystemHelp() {
	fmt.Println(`BlackLang ecosystem

Usage:
  black ecosystem [options]

Options:
  --json      Print machine-readable JSON
  --ir        Print compact BlackIR`)
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
  --affected <symbol>  Print affected graph for an entity, field, query, job, action, page, role, workflow, state, component, api, target, deploy, or ops
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
			Suggestion: "Use syntax, version, docs, explain, ide, ecosystem, agent, agent-contract, diagnostics, format, lint, migrate, migration, app, target, auth, role, access, accessibility, entity, computed, query, action, ops, layout, page, test, view, table, form, actions, ui, ui-profile, ui-modes, theme, theme-migration, search, filter, paginate, workflow, state, component, blackir, openapi, generated-test, benchmark, package, package-registry, adapter-marketplace, editor-marketplace, security, cors, deploy, audit, or csrf.",
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

func runIDE(args []string) {
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printIDEHelp()
		return
	}

	jsonOutput := hasJSONFlag(args)
	irOutput := hasIRFlag(args)
	command := ""
	commandArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		command = args[0]
		commandArgs = args[1:]
	}

	switch command {
	case "":
		result := IDESupport()
		if irOutput {
			fmt.Print(FormatIDESupportIR(result))
			return
		}
		if jsonOutput {
			printJSON(result)
			return
		}
		printIDESupportResult(result)
	case "diagnostics":
		config := LoadConfig(".")
		file := firstNonOptionArg(commandArgs)
		if file == "" {
			file = config.Source
		}
		if file == "" {
			file = "examples/warehouse/app.black"
		}
		result := IDEDiagnosticsFile(file)
		if irOutput {
			fmt.Print(FormatIDEDiagnosticsIR(result))
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
		printIDEDiagnosticsResult(result)
		if !result.Success {
			os.Exit(1)
		}
	default:
		diagnostic := Diagnostic{
			Code:       "UNKNOWN_IDE_COMMAND",
			Message:    fmt.Sprintf("No IDE command exists for %q.", command),
			Suggestion: "Use `black ide --json` or `black ide diagnostics app.black --json`.",
		}
		printCommandError("ide", jsonOutput, diagnostic)
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
	if len(args) == 0 || strings.HasPrefix(args[0], "-") || args[0] == "inspect" {
		inspectArgs := args
		if len(args) > 0 && args[0] == "inspect" {
			inspectArgs = args[1:]
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
		return
	}
	if args[0] == "migrate" {
		result := MigrateTheme(args[1:])
		if irOutput {
			fmt.Print(FormatThemeMigrationIR(result))
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
		printThemeMigrationResult(result)
		if !result.Success {
			os.Exit(1)
		}
		return
	}
	if !strings.HasPrefix(args[0], "-") {
		diagnostic := Diagnostic{
			Code:       "UNKNOWN_THEME_COMMAND",
			Message:    fmt.Sprintf("No theme command exists for %q.", args[0]),
			Suggestion: "Use `black theme inspect --json` or `black theme migrate old.blackthm new.blackthm --json`.",
		}
		printCommandError("theme", jsonOutput, diagnostic)
		os.Exit(1)
	}
}

func printThemeHelp() {
	fmt.Println(`BlackLang theme

Usage:
  black theme inspect [file] [options]
  black theme migrate <old.blackthm> <new.blackthm> [options]
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

func printIDEHelp() {
	fmt.Println(`BlackLang IDE

Usage:
  black ide [options]
  black ide diagnostics [file] [options]

Options:
  --json      Print machine-readable JSON
  --ir        Print compact BlackIR`)
}

func firstNonOptionArg(args []string) string {
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--out" || arg == "--affected" || arg == "--key-env" || arg == "--history" {
			index++
			continue
		}
		if len(arg) > 0 && arg[0] != '-' {
			return arg
		}
	}
	return ""
}

func nonOptionArgs(args []string) []string {
	values := []string{}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--out" || arg == "--affected" || arg == "--key-env" || arg == "--history" {
			index++
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		values = append(values, arg)
	}
	return values
}

func printDiagnostics(diagnostics []Diagnostic) {
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
		if diagnostic.Suggestion != "" {
			fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
		}
	}
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

func printAccessibilityAuditResult(result AccessibilityAuditResult) {
	if result.Success {
		fmt.Printf("accessibility audit ok %s\n", result.File)
		return
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

func printBenchmarkResult(result BenchmarkResult) {
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}

	fmt.Printf("benchmark %s\n", result.Summary.App)
	fmt.Printf("source files: %d\n", result.Source.Files)
	fmt.Printf("source lines: %d\n", result.Source.Lines)
	fmt.Printf("generated files: %d\n", result.Generated.Files)
	fmt.Printf("generated lines: %d\n", result.Generated.Lines)
	fmt.Printf("generated/source line ratio: %.2f\n", result.Ratios.GeneratedToSourceLines)
}

func printCoverageResult(result CoverageResult) {
	if !result.Success {
		printDiagnostics(result.Errors)
		return
	}
	fmt.Printf("web coverage %d%%\n", result.CompletionPercent)
	for _, area := range result.Areas {
		fmt.Printf("%s: %d%% (%s, weight %d)\n", area.Name, area.Score, area.Status, area.Weight)
	}
}

func printCoverageIssuesResult(result CoverageIssuesResult) {
	if !result.Success {
		printDiagnostics(result.Errors)
		return
	}
	fmt.Printf("web coverage issues %d%%\n", result.CompletionPercent)
	fmt.Printf("total: %d\n", result.Summary.Total)
	fmt.Printf("open: %d\n", result.Summary.Open)
	fmt.Printf("done: %d\n", result.Summary.Done)
	for _, issue := range result.Issues {
		fmt.Printf("- %s [%s] %s %s: %s\n", issue.ID, issue.Status, issue.Priority, issue.Area, issue.Title)
	}
}

func printIDESupportResult(result IDESupportResult) {
	if !result.Success {
		printDiagnostics(result.Errors)
		return
	}
	fmt.Printf("ide %s %s\n", result.Language.ID, result.Language.LanguageVersion)
	fmt.Printf("extensions: %s\n", strings.Join(result.Language.Extensions, ", "))
	fmt.Printf("diagnostics: %s\n", result.Capabilities.Diagnostics)
	fmt.Printf("completion items: %d\n", len(result.CompletionItems))
	fmt.Printf("snippets: %d\n", len(result.Snippets))
	fmt.Printf("diagnostic codes: %d\n", len(result.DiagnosticCodes))
}

func printIDEDiagnosticsResult(result IDEDiagnosticsResult) {
	if !result.Success {
		printDiagnostics(result.Errors)
		return
	}
	fmt.Printf("ide diagnostics %s\n", result.File)
	fmt.Printf("valid: %t\n", result.Valid)
	fmt.Printf("diagnostics: %d\n", result.Summary.Total)
	for _, diagnostic := range result.Diagnostics {
		fmt.Printf(
			"%s:%d:%d %s %s: %s\n",
			diagnostic.File,
			diagnostic.Range.Start.Line,
			diagnostic.Range.Start.Character,
			diagnostic.Severity,
			diagnostic.Code,
			diagnostic.Message,
		)
		if diagnostic.Suggestion != "" {
			fmt.Printf("suggestion: %s\n", diagnostic.Suggestion)
		}
	}
}

func printEcosystemResult(result EcosystemResult) {
	if !result.Success {
		printDiagnostics(result.Errors)
		return
	}
	fmt.Printf("ecosystem %s\n", result.Release.Channel)
	if result.Release.Trust.SignatureAlgorithm != "" {
		fmt.Printf("release trust: %s %s\n", result.Release.Trust.SignatureAlgorithm, result.Release.Trust.SignatureFilePattern)
		fmt.Printf("verify: %s\n", result.Release.Trust.VerifyCommand)
		if result.Release.Trust.TransparencyLog.PolicyFile != "" {
			fmt.Printf("release transparency: %s %s\n", result.Release.Trust.TransparencyLog.HashAlgorithm, result.Release.Trust.TransparencyLog.PolicyFile)
		}
		if result.Release.Trust.KeyRotation.PolicyFile != "" {
			fmt.Printf("release key rotation: %s overlap %d days\n", result.Release.Trust.KeyRotation.KeyIDFormat, result.Release.Trust.KeyRotation.MinimumOverlapDays)
		}
	}
	fmt.Printf("packages: %d\n", len(result.Packages))
	for _, pkg := range result.Packages {
		fmt.Printf("- %s [%s] %s\n", pkg.ID, pkg.Status, pkg.Path)
	}
	fmt.Printf("adapters: %d\n", len(result.Adapters))
	for _, adapter := range result.Adapters {
		fmt.Printf("- %s [%s] %s\n", adapter.ID, adapter.Status, adapter.Name)
	}
	fmt.Printf("registries: %d\n", len(result.Registries))
	for _, registry := range result.Registries {
		fmt.Printf("- %s [%s] %s\n", registry.ID, registry.Status, registry.Path)
	}
	fmt.Printf("marketplaces: %d\n", len(result.Marketplaces))
	for _, marketplace := range result.Marketplaces {
		fmt.Printf("- %s [%s] %s\n", marketplace.ID, marketplace.Status, marketplace.Path)
	}
}

func printAITaskBenchmarkResult(result AITaskBenchmarkResult) {
	if !result.Success {
		printDiagnostics(result.Errors)
		return
	}
	fmt.Printf("ai task benchmark %s\n", result.Summary.App)
	fmt.Printf("baseline source lines: %d\n", result.Baseline.SourceLines)
	fmt.Printf("baseline generated lines: %d\n", result.Baseline.GeneratedLines)
	fmt.Printf("scenarios: %d\n", result.Totals.ScenarioCount)
	fmt.Printf("estimated BlackLang tokens: %d\n", result.Totals.BlackLangTotal)
	fmt.Printf("estimated conventional tokens: %d\n", result.Totals.ConventionalTotal)
	fmt.Printf("estimated savings: %d%%\n", result.Totals.EstimatedSavingsPercent)
	for _, scenario := range result.Scenarios {
		fmt.Printf("- %s: %d%% savings\n", scenario.ID, scenario.EstimatedTokens.EstimatedSavingsPercent)
	}
}

func printAIEvalCorpusResult(result AIEvalCorpusResult) {
	if !result.Success {
		printDiagnostics(result.Errors)
		return
	}
	fmt.Printf("ai eval corpus %s\n", result.Summary.App)
	fmt.Printf("suite: %s\n", result.Suite.ID)
	fmt.Printf("mode: %s\n", result.Suite.Mode)
	fmt.Printf("cases: %d\n", result.Totals.CaseCount)
	fmt.Printf("repeat: %d\n", result.Totals.Repeat)
	fmt.Printf("total runs: %d\n", result.Totals.TotalRuns)
	fmt.Printf("estimated minutes: %d\n", result.Totals.EstimatedMinutesTotal)
	fmt.Printf("estimated savings: %d%%\n", result.Totals.EstimatedSavingsPercent)
	for _, item := range result.Cases {
		fmt.Printf("- %s: %s\n", item.ID, item.Name)
	}
}

func printAIEvalHistoryResult(result AIEvalHistoryResult) {
	if !result.Success {
		printDiagnostics(result.Errors)
		return
	}
	fmt.Printf("ai eval history %s\n", result.History.ID)
	fmt.Printf("status: %s\n", result.History.Status)
	fmt.Printf("published: %s\n", result.History.PublishedAt)
	fmt.Printf("runs: %d\n", result.Summary.RunCount)
	fmt.Printf("models: %d\n", result.Summary.ModelCount)
	fmt.Printf("total runs: %d\n", result.Summary.TotalRuns)
	fmt.Printf("passed: %d\n", result.Summary.PassedRuns)
	fmt.Printf("failed: %d\n", result.Summary.FailedRuns)
	fmt.Printf("average score: %d\n", result.Summary.AverageScore)
	fmt.Printf("coverage: %d%%\n", result.Summary.CoveragePercent)
	for _, run := range result.Runs {
		fmt.Printf("- %s: %s score %d%%\n", run.ID, run.Status, run.AverageScore)
	}
}

func printSchemaMigrationPlanResult(result SchemaMigrationPlanResult) {
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}

	fmt.Printf("schema migration plan %s -> %s\n", result.OldFile, result.NewFile)
	fmt.Printf("safe: %t\n", result.Safe)
	fmt.Printf("destructive: %t\n", result.Destructive)
	fmt.Printf("changes: %d\n", len(result.Changes))
	for _, change := range result.Changes {
		target := change.Entity
		if change.Field != "" {
			target += "." + change.Field
		}
		if change.Index != "" {
			target += "." + change.Index
		}
		value := ""
		if change.OldValue != "" || change.NewValue != "" {
			value = fmt.Sprintf(" value=%s->%s", change.OldValue, change.NewValue)
		}
		fmt.Printf("- %s %s risk=%s%s\n", change.Type, target, change.Risk, value)
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
	printAffectedItems("queries", affected.Queries)
	printAffectedItems("jobs", affected.Jobs)
	printAffectedItems("actions", affected.Actions)
	printAffectedItems("transactions", affected.Transactions)
	printAffectedItems("services", affected.Services)
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

func printThemeMigrationResult(result ThemeMigrationResult) {
	if !result.Success {
		for _, diagnostic := range result.Errors {
			fmt.Fprintf(os.Stderr, "%s:%d:%d %s: %s\n", diagnostic.File, diagnostic.Line, diagnostic.Column, diagnostic.Code, diagnostic.Message)
			if diagnostic.Suggestion != "" {
				fmt.Fprintf(os.Stderr, "suggestion: %s\n", diagnostic.Suggestion)
			}
		}
		return
	}
	fmt.Printf("theme migration ok %s -> %s\n", result.OldFile, result.NewFile)
	fmt.Printf("safe: %t\n", result.Safe)
	fmt.Printf("theme: %s v%d -> v%d\n", result.Summary.NewTheme, result.Summary.OldVersion, result.Summary.NewVersion)
	fmt.Printf("profile: %s v%d -> v%d\n", result.Summary.NewProfile, result.Summary.OldProfileVersion, result.Summary.NewProfileVersion)
	if len(result.Changes) == 0 {
		fmt.Println("changes: none")
		return
	}
	fmt.Printf("changes: %d\n", len(result.Changes))
	for _, change := range result.Changes {
		switch change.Type {
		case "slot-appended":
			fmt.Printf("- %s %s.%s slot %s at %d\n", change.Type, change.Profile, change.Mode, change.Slot, change.Index)
		case "mode-added":
			fmt.Printf("- %s %s.%s\n", change.Type, change.Profile, change.Mode)
		default:
			fmt.Printf("- %s %s -> %s\n", change.Type, change.OldValue, change.NewValue)
		}
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
