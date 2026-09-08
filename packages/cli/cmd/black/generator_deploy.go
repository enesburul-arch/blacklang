package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type generatedDeployPort struct {
	Env     string `json:"env"`
	Default string `json:"default"`
}

type generatedDeployEnv struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
}

type generatedDeployPreview struct {
	Mode        string `json:"mode"`
	ComposeFile string `json:"composeFile"`
	ProjectName string `json:"projectName"`
	PortEnv     string `json:"portEnv"`
	PortDefault string `json:"portDefault"`
}

type generatedDeployRollback struct {
	Strategy    string `json:"strategy"`
	Keep        int    `json:"keep"`
	PlanFile    string `json:"planFile"`
	ReleaseRoot string `json:"releaseRoot"`
}

type generatedDeployCloud struct {
	Provider      string                        `json:"provider"`
	Adapter       string                        `json:"adapter"`
	ManifestFile  string                        `json:"manifestFile"`
	PlanFile      string                        `json:"planFile"`
	PreflightFile string                        `json:"preflightFile"`
	RunnerFile    string                        `json:"runnerFile"`
	AppEnv        string                        `json:"appEnv"`
	RegionEnv     string                        `json:"regionEnv,omitempty"`
	Image         string                        `json:"image"`
	RequiredEnv   []string                      `json:"requiredEnv"`
	AuthEnv       []string                      `json:"authEnv"`
	OptionalEnv   []string                      `json:"optionalEnv,omitempty"`
	Execution     generatedDeployCloudExecution `json:"execution"`
	Steps         []string                      `json:"steps"`
}

type generatedDeployCloudExecution struct {
	Mode             string                      `json:"mode"`
	Policy           string                      `json:"policy"`
	PlanCommand      string                      `json:"planCommand"`
	PreflightCommand string                      `json:"preflightCommand"`
	ApplyCommand     string                      `json:"applyCommand"`
	ApplyFlag        string                      `json:"applyFlag"`
	CLI              generatedDeployCloudCLI     `json:"cli"`
	Apply            generatedDeployCloudCommand `json:"apply"`
}

type generatedDeployCloudCLI struct {
	Primary      string   `json:"primary"`
	Alternatives []string `json:"alternatives,omitempty"`
	VersionArgs  []string `json:"versionArgs"`
}

type generatedDeployCloudCommand struct {
	Executable string   `json:"executable"`
	Args       []string `json:"args"`
	Cwd        string   `json:"cwd"`
}

type generatedDeployManifest struct {
	Version        string                   `json:"version"`
	App            string                   `json:"app"`
	Target         string                   `json:"target"`
	Frontend       string                   `json:"frontend"`
	Backend        string                   `json:"backend"`
	Database       string                   `json:"database"`
	Port           generatedDeployPort      `json:"port"`
	Env            []generatedDeployEnv     `json:"env"`
	Preview        *generatedDeployPreview  `json:"preview,omitempty"`
	Rollback       *generatedDeployRollback `json:"rollback,omitempty"`
	Cloud          *generatedDeployCloud    `json:"cloud,omitempty"`
	GeneratedFiles []string                 `json:"generatedFiles"`
}

type generatedDeployRollbackPlan struct {
	Version      string   `json:"version"`
	App          string   `json:"app"`
	Target       string   `json:"target"`
	Strategy     string   `json:"strategy"`
	Keep         int      `json:"keep"`
	ReleaseRoot  string   `json:"releaseRoot"`
	CurrentLink  string   `json:"currentLink"`
	PreviousLink string   `json:"previousLink"`
	Steps        []string `json:"steps"`
}

func (g *webGenerator) hasDeployManifest() bool {
	return g.hasDeployPreviewLocal() || g.hasDeployRollback() || g.hasDeployCloud()
}

func (g *webGenerator) hasDeployPreviewLocal() bool {
	return g.program.Deploy != nil && g.program.Deploy.Preview != nil && g.program.Deploy.Preview.Mode == "local"
}

func (g *webGenerator) hasDeployRollback() bool {
	return g.program.Deploy != nil && g.program.Deploy.Rollback != nil && g.program.Deploy.Rollback.Strategy == "keep" && g.program.Deploy.Rollback.Keep > 0
}

func (g *webGenerator) hasDeployCloud() bool {
	return g.program.Deploy != nil && g.program.Deploy.Cloud != nil && g.program.Deploy.Cloud.Provider != ""
}

func (g *webGenerator) deployServiceName() string {
	serviceName := kebabCase(g.program.App.Name)
	if serviceName == "" {
		return "blacklang-app"
	}
	return serviceName
}

func (g *webGenerator) deployPreviewServiceName() string {
	return g.deployServiceName() + "-preview"
}

func (g *webGenerator) deployPreviewProjectName() string {
	return g.deployPreviewServiceName()
}

func (g *webGenerator) deployPreviewDatabaseServiceName() string {
	return g.composeDatabaseServiceName(g.deployPreviewServiceName(), true)
}

func (g *webGenerator) deployPreviewPortDefault() string {
	port, err := strconv.Atoi(g.deployPortDefault())
	if err == nil {
		if port >= 1 && port <= 64535 {
			return strconv.Itoa(port + 1000)
		}
		if port > 1000 && port <= 65535 {
			return strconv.Itoa(port - 1000)
		}
	}
	return "4001"
}

func (g *webGenerator) packageDeployScripts() string {
	if g.deployTarget() != "docker" {
		return ""
	}
	scripts := []string{}
	if g.hasDeployPreviewLocal() {
		projectName := g.deployPreviewProjectName()
		scripts = append(scripts,
			fmt.Sprintf(`"deploy:preview": "docker compose -f docker-compose.preview.yml --project-name %s up --build"`, projectName),
			fmt.Sprintf(`"deploy:preview:down": "docker compose -f docker-compose.preview.yml --project-name %s down"`, projectName),
		)
	}
	if g.hasDeployRollback() {
		scripts = append(scripts, `"deploy:rollback:plan": "node scripts/rollback-plan.mjs"`)
	}
	if g.hasDeployCloud() {
		scripts = append(scripts,
			`"deploy:cloud:plan": "node scripts/cloud-plan.mjs"`,
			`"deploy:cloud:preflight": "node scripts/cloud-exec.mjs --preflight"`,
			`"deploy:cloud:exec": "node scripts/cloud-exec.mjs --apply"`,
		)
	}
	if len(scripts) == 0 {
		return ""
	}
	return ",\n    " + strings.Join(scripts, ",\n    ")
}

func (g *webGenerator) deployReadmeSection() string {
	if g.deployTarget() != "docker" {
		return ""
	}
	lines := []string{}
	if g.hasDeployPreviewLocal() {
		lines = append(lines, "npm run deploy:preview", "npm run deploy:preview:down")
	}
	if g.hasDeployRollback() {
		lines = append(lines, "npm run deploy:rollback:plan")
	}
	if g.hasDeployCloud() {
		lines = append(lines, "npm run deploy:cloud:plan", "npm run deploy:cloud:preflight", "npm run deploy:cloud:exec")
	}
	if len(lines) == 0 {
		return ""
	}
	return "## Deployment\n\n~~~bash\n" + strings.Join(lines, "\n") + "\n~~~\n\n"
}

func (g *webGenerator) dockerComposePreviewYAML() string {
	serviceName := g.deployPreviewServiceName()
	databaseServiceName := g.deployPreviewDatabaseServiceName()
	volumeName := serviceName + "-data"

	var builder strings.Builder
	builder.WriteString("# Generated by BlackLang. Do not edit manually.\n\n")
	builder.WriteString("services:\n")
	builder.WriteString(fmt.Sprintf("  %s:\n", serviceName))
	builder.WriteString("    build: .\n")
	if g.usesSQLComposeService() {
		builder.WriteString("    depends_on:\n")
		builder.WriteString(fmt.Sprintf("      %s:\n", databaseServiceName))
		builder.WriteString("        condition: service_healthy\n")
	}
	builder.WriteString("    ports:\n")
	builder.WriteString(fmt.Sprintf("      - \"${BLACKLANG_PREVIEW_PORT:-%s}:${%s:-%s}\"\n", g.deployPreviewPortDefault(), g.deployPortEnv(), g.deployPortDefault()))
	builder.WriteString("    environment:\n")
	builder.WriteString("      NODE_ENV: production\n")
	builder.WriteString("      BLACKLANG_DEPLOY_PREVIEW: \"local\"\n")
	for _, name := range g.deployEnvNames() {
		builder.WriteString(fmt.Sprintf("      %s: \"%s\"\n", name, g.composePreviewEnvValue(name, databaseServiceName)))
	}
	builder.WriteString(g.composeHealthcheck("    "))
	if !g.usesSQLComposeService() {
		builder.WriteString("    volumes:\n")
		builder.WriteString(fmt.Sprintf("      - %s:/app/data\n\n", volumeName))
		builder.WriteString("volumes:\n")
		builder.WriteString(fmt.Sprintf("  %s:\n", volumeName))
		return builder.String()
	}
	builder.WriteString("\n")
	builder.WriteString(g.sqlComposeServiceYAML(serviceName, databaseServiceName, true))
	builder.WriteString("volumes:\n")
	builder.WriteString(fmt.Sprintf("  %s-%s:\n", serviceName, g.targetDatabase()))
	return builder.String()
}

func (g *webGenerator) composePreviewEnvValue(name string, databaseHost string) string {
	if name == g.deployPortEnv() {
		return fmt.Sprintf("${%s:-%s}", name, g.deployPortDefault())
	}
	if name == g.databaseEnvName() {
		return fmt.Sprintf("${%s:-%s}", name, g.composePreviewDatabaseURL(databaseHost))
	}
	if cors := g.corsConfig(); cors != nil && cors.Origins.Name == name {
		return fmt.Sprintf("${%s:-http://localhost:%s}", name, g.deployPreviewPortDefault())
	}
	return fmt.Sprintf("${%s}", name)
}

func (g *webGenerator) composePreviewDatabaseURL(databaseHost string) string {
	if g.isPostgres() {
		return fmt.Sprintf("postgresql://blacklang:blacklang@%s:5432/blacklang_preview", databaseHost)
	}
	if g.isMySQL() {
		return fmt.Sprintf("mysql://blacklang:blacklang@%s:3306/blacklang_preview", databaseHost)
	}
	return "file:/app/data/preview.db"
}

func (g *webGenerator) deployManifestJSON() string {
	manifest := generatedDeployManifest{
		Version:  version,
		App:      g.program.App.Name,
		Target:   g.deployTarget(),
		Frontend: g.targetFrontend(),
		Backend:  g.targetBackend(),
		Database: g.targetDatabase(),
		Port: generatedDeployPort{
			Env:     g.deployPortEnv(),
			Default: g.deployPortDefault(),
		},
		Env:            g.deployDeclaredEnv(),
		GeneratedFiles: g.deployGeneratedFiles(),
	}
	if g.hasDeployPreviewLocal() {
		manifest.Preview = &generatedDeployPreview{
			Mode:        "local",
			ComposeFile: "docker-compose.preview.yml",
			ProjectName: g.deployPreviewProjectName(),
			PortEnv:     "BLACKLANG_PREVIEW_PORT",
			PortDefault: g.deployPreviewPortDefault(),
		}
	}
	if g.hasDeployRollback() {
		manifest.Rollback = &generatedDeployRollback{
			Strategy:    g.program.Deploy.Rollback.Strategy,
			Keep:        g.program.Deploy.Rollback.Keep,
			PlanFile:    "scripts/rollback-plan.mjs",
			ReleaseRoot: "releases",
		}
	}
	if g.hasDeployCloud() {
		cloud := g.deployCloudManifest()
		manifest.Cloud = &cloud
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "{}\n"
	}
	return string(data) + "\n"
}

func (g *webGenerator) deployCloudManifest() generatedDeployCloud {
	if !g.hasDeployCloud() {
		return generatedDeployCloud{}
	}
	cloud := g.program.Deploy.Cloud
	requiredEnv := []string{cloud.App.Name}
	if cloud.Region.Name != "" {
		requiredEnv = append(requiredEnv, cloud.Region.Name)
	}
	cli := g.deployCloudCLI()
	apply := generatedDeployCloudCommand{
		Executable: cli.Primary,
		Args:       g.deployCloudApplyArgs(),
		Cwd:        ".",
	}
	return generatedDeployCloud{
		Provider:      cloud.Provider,
		Adapter:       "provider-cli",
		ManifestFile:  "deploy/cloud.json",
		PlanFile:      "scripts/cloud-plan.mjs",
		PreflightFile: "scripts/cloud-exec.mjs",
		RunnerFile:    "scripts/cloud-exec.mjs",
		AppEnv:        cloud.App.Name,
		RegionEnv:     cloud.Region.Name,
		Image:         g.deployServiceName() + ":latest",
		RequiredEnv:   requiredEnv,
		AuthEnv:       g.deployCloudAuthEnv(),
		OptionalEnv:   g.deployCloudOptionalEnv(),
		Execution: generatedDeployCloudExecution{
			Mode:             "explicit-apply",
			Policy:           "Plan and preflight are read-only. Apply mode requires --apply, resolved required environment variables, and a provider CLI on PATH. Generated output redacts configured environment values from command output.",
			PlanCommand:      "npm run deploy:cloud:plan",
			PreflightCommand: "npm run deploy:cloud:preflight",
			ApplyCommand:     "npm run deploy:cloud:exec",
			ApplyFlag:        "--apply",
			CLI:              cli,
			Apply:            apply,
		},
		Steps: []string{
			"Build the generated Docker image from .black source.",
			"Read deploy/cloud.json and resolve app/region from environment variables.",
			"Run npm run deploy:cloud:plan for read-only declared cloud metadata.",
			"Run npm run deploy:cloud:preflight to check required env and provider CLI availability without mutating infrastructure.",
			"Run npm run deploy:cloud:exec only in an authorized environment; it invokes the selected provider CLI through scripts/cloud-exec.mjs --apply.",
		},
	}
}

func (g *webGenerator) deployCloudCLI() generatedDeployCloudCLI {
	if !g.hasDeployCloud() {
		return generatedDeployCloudCLI{VersionArgs: []string{"--version"}}
	}
	switch g.program.Deploy.Cloud.Provider {
	case "fly":
		return generatedDeployCloudCLI{Primary: "fly", Alternatives: []string{"flyctl"}, VersionArgs: []string{"version"}}
	case "render":
		return generatedDeployCloudCLI{Primary: "render", VersionArgs: []string{"--version"}}
	case "railway":
		return generatedDeployCloudCLI{Primary: "railway", VersionArgs: []string{"--version"}}
	default:
		return generatedDeployCloudCLI{Primary: g.program.Deploy.Cloud.Provider, VersionArgs: []string{"--version"}}
	}
}

func (g *webGenerator) deployCloudAuthEnv() []string {
	if !g.hasDeployCloud() {
		return nil
	}
	switch g.program.Deploy.Cloud.Provider {
	case "fly":
		return []string{"FLY_API_TOKEN", "FLY_ACCESS_TOKEN"}
	case "render":
		return []string{"RENDER_API_KEY"}
	case "railway":
		return []string{"RAILWAY_TOKEN", "RAILWAY_API_TOKEN"}
	default:
		return nil
	}
}

func (g *webGenerator) deployCloudOptionalEnv() []string {
	if !g.hasDeployCloud() {
		return nil
	}
	switch g.program.Deploy.Cloud.Provider {
	case "render":
		return []string{"BLACKLANG_CLOUD_IMAGE", "BLACKLANG_CLOUD_COMMIT"}
	case "railway":
		return []string{"BLACKLANG_CLOUD_SERVICE", "BLACKLANG_CLOUD_MESSAGE"}
	default:
		return nil
	}
}

func (g *webGenerator) deployCloudApplyArgs() []string {
	if !g.hasDeployCloud() {
		return nil
	}
	cloud := g.program.Deploy.Cloud
	app := "${" + cloud.App.Name + "}"
	region := "${" + cloud.Region.Name + "}"
	switch cloud.Provider {
	case "fly":
		return []string{"deploy", ".", "--app", app, "--yes"}
	case "render":
		return []string{"deploys", "create", app, "--wait", "--confirm", "-o", "json"}
	case "railway":
		args := []string{"up", ".", "--project", app, "--ci", "--yes", "--json"}
		if cloud.Region.Name != "" {
			args = append(args, "--environment", region)
		}
		return args
	default:
		return []string{}
	}
}

func (g *webGenerator) deployCloudJSON() string {
	cloud := g.deployCloudManifest()
	data, err := json.MarshalIndent(cloud, "", "  ")
	if err != nil {
		return "{}\n"
	}
	return string(data) + "\n"
}

func (g *webGenerator) cloudPlanMJS() string {
	return `// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import fs from "node:fs";
import path from "node:path";

const configPath = path.join(process.cwd(), "deploy", "cloud.json");
const config = JSON.parse(fs.readFileSync(configPath, "utf8"));
const execution = config.execution ?? {};
const requiredEnv = Array.isArray(config.requiredEnv) ? config.requiredEnv : [];
const authEnv = Array.isArray(config.authEnv) ? config.authEnv : [];
const optionalEnv = Array.isArray(config.optionalEnv) ? config.optionalEnv : [];
const missingEnv = requiredEnv.filter((name) => !process.env[name]);
const presentAuthEnv = authEnv.filter((name) => Boolean(process.env[name]));
const presentOptionalEnv = optionalEnv.filter((name) => Boolean(process.env[name]));

console.log(JSON.stringify({
  success: true,
  command: "deploy:cloud:plan",
  ready: missingEnv.length === 0,
  app: process.env[config.appEnv] ?? null,
  region: config.regionEnv ? (process.env[config.regionEnv] ?? null) : null,
  provider: config.provider,
  adapter: config.adapter,
  image: config.image,
  requiredEnv,
  authEnv,
  optionalEnv,
  missingEnv,
  presentAuthEnv,
  presentOptionalEnv,
  execution: {
    mode: execution.mode,
    policy: execution.policy,
    planCommand: execution.planCommand,
    preflightCommand: execution.preflightCommand,
    applyCommand: execution.applyCommand,
    applyFlag: execution.applyFlag,
    cli: execution.cli,
    apply: execution.apply
  },
  steps: config.steps
}, null, 2));
`
}

func (g *webGenerator) cloudExecMJS() string {
	return `// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { spawnSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

const configPath = path.join(process.cwd(), "deploy", "cloud.json");
const config = JSON.parse(fs.readFileSync(configPath, "utf8"));
const execution = config.execution ?? {};
const requiredEnv = Array.isArray(config.requiredEnv) ? config.requiredEnv : [];
const authEnv = Array.isArray(config.authEnv) ? config.authEnv : [];
const optionalEnv = Array.isArray(config.optionalEnv) ? config.optionalEnv : [];
const cli = execution.cli ?? {};
const apply = execution.apply ?? {};
const argv = new Set(process.argv.slice(2));
const mode = argv.has("--apply") ? "apply" : "preflight";

function missingEnv(names) {
  return names.filter((name) => !process.env[name]);
}

function presentEnv(names) {
  return names.filter((name) => Boolean(process.env[name]));
}

function commandCandidates() {
  const candidates = [];
  if (typeof cli.primary === "string" && cli.primary.length > 0) candidates.push(cli.primary);
  if (Array.isArray(cli.alternatives)) {
    for (const item of cli.alternatives) {
      if (typeof item === "string" && item.length > 0 && !candidates.includes(item)) candidates.push(item);
    }
  }
  return candidates;
}

function checkCLI(candidates) {
  const versionArgs = Array.isArray(cli.versionArgs) && cli.versionArgs.length > 0 ? cli.versionArgs : ["--version"];
  const attempts = [];
  for (const candidate of candidates) {
    const result = spawnSync(candidate, versionArgs, { encoding: "utf8", shell: false });
    if (result.error && result.error.code === "ENOENT") {
      attempts.push({ executable: candidate, available: false, reason: "not-found" });
      continue;
    }
    if (result.error) {
      attempts.push({ executable: candidate, available: false, reason: result.error.code ?? result.error.message });
      continue;
    }
    attempts.push({ executable: candidate, available: true, status: result.status });
    return { available: true, selected: candidate, versionArgs, attempts };
  }
  return { available: false, selected: null, versionArgs, attempts };
}

function baseApplyArgs() {
  return Array.isArray(apply.args) ? [...apply.args] : [];
}

function applyArgsWithOptionalEnv() {
  const args = baseApplyArgs();
  if (config.provider === "render") {
    if (process.env.BLACKLANG_CLOUD_IMAGE) args.push("--image", "${BLACKLANG_CLOUD_IMAGE}");
    if (process.env.BLACKLANG_CLOUD_COMMIT) args.push("--commit", "${BLACKLANG_CLOUD_COMMIT}");
  }
  if (config.provider === "railway") {
    if (process.env.BLACKLANG_CLOUD_SERVICE) args.push("--service", "${BLACKLANG_CLOUD_SERVICE}");
    if (process.env.BLACKLANG_CLOUD_MESSAGE) args.push("--message", "${BLACKLANG_CLOUD_MESSAGE}");
  }
  return args;
}

function resolveArg(arg) {
  if (typeof arg !== "string") return arg;
  return arg.replace(/\$\{([A-Z0-9_]+)\}/g, (_match, name) => process.env[name] ?? "");
}

function redact(text) {
  if (!text) return "";
  let safe = String(text);
  for (const name of [...requiredEnv, ...authEnv, ...optionalEnv]) {
    const value = process.env[name];
    if (value) {
      safe = safe.split(value).join("<" + name + ">");
    }
  }
  return safe.length > 12000 ? safe.slice(0, 12000) + "\n<blacklang-output-truncated>" : safe;
}

function redactedCommand(executable, args) {
  return {
    executable: executable ?? cli.primary ?? null,
    args,
    cwd: apply.cwd ?? "."
  };
}

function baseReport(command, cliCheck, blocked, blockReasons, commandArgs) {
  return {
    success: command === "deploy:cloud:exec" ? !blocked : true,
    command,
    mode,
    ready: !blocked,
    blocked,
    blockReasons,
    provider: config.provider,
    adapter: config.adapter,
    appEnv: config.appEnv,
    regionEnv: config.regionEnv ?? null,
    image: config.image,
    requiredEnv,
    authEnv,
    optionalEnv,
    missingEnv: missingEnv(requiredEnv),
    presentAuthEnv: presentEnv(authEnv),
    presentOptionalEnv: presentEnv(optionalEnv),
    cli: cliCheck,
    apply: redactedCommand(cliCheck.selected, commandArgs),
    policy: execution.policy,
    steps: config.steps
  };
}

const commandArgs = applyArgsWithOptionalEnv();
const cliCheck = checkCLI(commandCandidates());
const missingRequired = missingEnv(requiredEnv);
const blockReasons = [];
if (missingRequired.length > 0) blockReasons.push("missing required environment variables");
if (!cliCheck.available) blockReasons.push("provider CLI executable was not found on PATH");
const blocked = blockReasons.length > 0;

if (mode === "preflight") {
  console.log(JSON.stringify(baseReport("deploy:cloud:preflight", cliCheck, blocked, blockReasons, commandArgs), null, 2));
  process.exit(0);
}

if (blocked) {
  console.log(JSON.stringify(baseReport("deploy:cloud:exec", cliCheck, true, blockReasons, commandArgs), null, 2));
  process.exit(1);
}

const resolvedArgs = commandArgs.map(resolveArg);
const child = spawnSync(cliCheck.selected, resolvedArgs, {
  cwd: process.cwd(),
  env: process.env,
  encoding: "utf8",
  shell: false
});
const success = child.status === 0;
const report = {
  ...baseReport("deploy:cloud:exec", cliCheck, false, [], commandArgs),
  success,
  status: child.status,
  signal: child.signal,
  output: {
    stdout: redact(child.stdout),
    stderr: redact(child.stderr)
  },
  error: child.error ? (child.error.code ?? child.error.message) : null
};
console.log(JSON.stringify(report, null, 2));
process.exit(success ? 0 : 1);
`
}

func (g *webGenerator) deployRollbackJSON() string {
	keep := 0
	strategy := ""
	if g.program.Deploy != nil && g.program.Deploy.Rollback != nil {
		keep = g.program.Deploy.Rollback.Keep
		strategy = g.program.Deploy.Rollback.Strategy
	}
	plan := generatedDeployRollbackPlan{
		Version:      version,
		App:          g.program.App.Name,
		Target:       g.deployTarget(),
		Strategy:     strategy,
		Keep:         keep,
		ReleaseRoot:  "releases",
		CurrentLink:  "current",
		PreviousLink: "previous",
		Steps: []string{
			"Build a new generated artifact from .black source.",
			"Store each release under deploy.releaseRoot using an immutable release id.",
			"Keep only the newest deploy.keep releases after a successful deployment.",
			"Use deploy.previousLink as the rollback candidate before switching deploy.currentLink.",
			"Run npm run deploy:rollback:plan to inspect the deterministic local rollback plan before mutating infrastructure.",
		},
	}
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return "{}\n"
	}
	return string(data) + "\n"
}

func (g *webGenerator) rollbackPlanMJS() string {
	return `// Generated by BlackLang. Do not edit manually.

import fs from "node:fs";
import path from "node:path";

const configPath = path.join(process.cwd(), "deploy", "rollback.json");
const config = JSON.parse(fs.readFileSync(configPath, "utf8"));
const releaseRoot = path.join(process.cwd(), config.releaseRoot);
const releases = fs.existsSync(releaseRoot)
  ? fs.readdirSync(releaseRoot, { withFileTypes: true })
      .filter((entry) => entry.isDirectory())
      .map((entry) => entry.name)
      .sort()
      .reverse()
  : [];
const retainedReleases = releases.slice(0, config.keep);
const rollbackCandidate = retainedReleases[1] ?? null;

console.log(JSON.stringify({
  success: true,
  command: "deploy:rollback:plan",
  app: config.app,
  target: config.target,
  strategy: config.strategy,
  keep: config.keep,
  releaseRoot: config.releaseRoot,
  retainedReleases,
  rollbackCandidate,
  steps: config.steps
}, null, 2));
`
}

func (g *webGenerator) deployDeclaredEnv() []generatedDeployEnv {
	if g.program.Deploy == nil {
		return nil
	}
	env := []generatedDeployEnv{}
	for _, decl := range g.program.Deploy.Env {
		env = append(env, generatedDeployEnv{Name: decl.Name, Mode: decl.Mode})
	}
	return env
}

func (g *webGenerator) deployGeneratedFiles() []string {
	files := []string{
		".env.example",
		".dockerignore",
		"Dockerfile",
		"docker-compose.yml",
		"package.json",
		"security/secrets.json",
		"scripts/secrets-plan.mjs",
		"scripts/secrets-provider.mjs",
		"src/server.ts",
	}
	if g.hasDeployPreviewLocal() {
		files = append(files, "docker-compose.preview.yml")
	}
	if g.hasDeployManifest() {
		files = append(files, "deploy/manifest.json")
	}
	if g.hasDeployRollback() {
		files = append(files, "deploy/rollback.json", "scripts/rollback-plan.mjs")
	}
	if g.hasDeployCloud() {
		files = append(files, "deploy/cloud.json", "scripts/cloud-plan.mjs", "scripts/cloud-exec.mjs")
	}
	if len(g.program.Migrations) > 0 {
		files = append(files, "migrations/manifest.json", "migrations/*.sql", "src/migrate.ts")
	}
	if g.hasServices() {
		files = append(files, "services/manifest.json")
		for _, service := range g.program.Services {
			files = append(files, "src/services/"+kebabCase(service.Name)+".ts")
		}
	}
	return files
}
