package main

import (
	"encoding/json"
	"strings"
)

type generatedSecretReference struct {
	Name        string   `json:"name"`
	Sources     []string `json:"sources"`
	Kind        string   `json:"kind"`
	Required    bool     `json:"required"`
	Sensitive   bool     `json:"sensitive"`
	Description string   `json:"description"`
}

type generatedSecretManifest struct {
	Version      string                     `json:"version"`
	App          string                     `json:"app"`
	Target       string                     `json:"target"`
	Provider     string                     `json:"provider"`
	ProviderEnv  string                     `json:"providerEnv"`
	PrefixEnv    string                     `json:"prefixEnv"`
	PlanFile     string                     `json:"planFile"`
	ProviderFile string                     `json:"providerFile"`
	Policy       string                     `json:"policy"`
	References   []generatedSecretReference `json:"references"`
	Steps        []string                   `json:"steps"`
}

func (g *webGenerator) secretManifestJSON() string {
	manifest := generatedSecretManifest{
		Version:      version,
		App:          g.program.App.Name,
		Target:       g.targetName(),
		Provider:     "environment",
		ProviderEnv:  "BLACKLANG_SECRET_PROVIDER",
		PrefixEnv:    "BLACKLANG_SECRET_PREFIX",
		PlanFile:     "scripts/secrets-plan.mjs",
		ProviderFile: "scripts/secrets-provider.mjs",
		Policy:       "secret values must stay out of .black source and generated manifests or provider preflight output",
		References:   g.secretReferences(),
		Steps: []string{
			"Keep secret values out of .black source files.",
			"Set required references in the process environment or mirror the same names under the selected secret manager prefix.",
			"Run npm run security:secrets:plan before deployment to verify environment presence without printing values.",
			"Run npm run security:secrets:preflight before provider handoff to verify provider label, prefix, and CLI availability without fetching secret values.",
		},
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "{}\n"
	}
	return string(data) + "\n"
}

func (g *webGenerator) secretReferences() []generatedSecretReference {
	refs := []generatedSecretReference{}
	index := map[string]int{}
	add := func(name string, source string, kind string, required bool, sensitive bool, description string) {
		if name == "" {
			return
		}
		if existing, ok := index[name]; ok {
			ref := &refs[existing]
			ref.Required = ref.Required || required
			ref.Sensitive = ref.Sensitive || sensitive
			if !stringSliceContains(ref.Sources, source) {
				ref.Sources = append(ref.Sources, source)
			}
			if ref.Description == "" {
				ref.Description = description
			}
			return
		}
		index[name] = len(refs)
		refs = append(refs, generatedSecretReference{
			Name:        name,
			Sources:     []string{source},
			Kind:        kind,
			Required:    required,
			Sensitive:   sensitive,
			Description: description,
		})
	}

	add(g.deployPortEnv(), "deploy.port", "runtime-port", false, false, "Generated HTTP server port.")
	if g.hasDeployPreviewLocal() {
		add("BLACKLANG_PREVIEW_PORT", "deploy.preview", "runtime-port", false, false, "Local preview host port.")
	}
	add(g.databaseEnvName(), "database.url", "database-url", g.deployEnvRequired(g.databaseEnvName()), true, "Database connection string.")
	if g.isPostgres() {
		add("POSTGRES_DB", "generated.postgres", "postgres-database", false, false, "Local PostgreSQL database name.")
		add("POSTGRES_USER", "generated.postgres", "postgres-user", false, false, "Local PostgreSQL user.")
		add("POSTGRES_PASSWORD", "generated.postgres", "postgres-password", false, true, "Local PostgreSQL password.")
	}
	if g.isMySQL() {
		add("MYSQL_DATABASE", "generated.mysql", "mysql-database", false, false, "Local MySQL database name.")
		add("MYSQL_USER", "generated.mysql", "mysql-user", false, false, "Local MySQL user.")
		add("MYSQL_PASSWORD", "generated.mysql", "mysql-password", false, true, "Local MySQL password.")
		add("MYSQL_ROOT_PASSWORD", "generated.mysql", "mysql-root-password", false, true, "Local MySQL root password.")
	}
	if cors := g.corsConfig(); cors != nil && cors.Origins.Name != "" {
		add(cors.Origins.Name, "security.cors.origins", "cors-origin-list", false, false, "Comma-separated browser origins allowed by generated CORS middleware.")
	}
	if g.program.Deploy != nil && g.program.Deploy.Cloud != nil {
		add(g.program.Deploy.Cloud.App.Name, "deploy.cloud.app", "cloud-app", true, false, "Cloud provider app or service identifier.")
		if g.program.Deploy.Cloud.Region.Name != "" {
			add(g.program.Deploy.Cloud.Region.Name, "deploy.cloud.region", "cloud-region", true, false, "Cloud provider region identifier.")
		}
	}
	if g.program.Ops != nil && g.program.Ops.Observe != nil {
		add(g.program.Ops.Observe.Endpoint.Name, "ops.observe.endpoint", "observability-endpoint", false, true, "External observability endpoint.")
	}
	if g.program.Deploy != nil {
		for _, env := range g.program.Deploy.Env {
			add(env.Name, "deploy.env", "declared-env", env.Mode == "required", inferEnvReferenceSensitive(env.Name), "Declared deployment environment reference.")
		}
	}
	return refs
}

func (g *webGenerator) deployEnvRequired(name string) bool {
	if g.program.Deploy == nil {
		return false
	}
	for _, env := range g.program.Deploy.Env {
		if env.Name == name && env.Mode == "required" {
			return true
		}
	}
	return false
}

func inferEnvReferenceSensitive(name string) bool {
	upper := strings.ToUpper(name)
	for _, marker := range []string{"PASSWORD", "SECRET", "TOKEN", "PRIVATE", "KEY", "DATABASE_URL", "DSN", "WEBHOOK", "ENDPOINT"} {
		if strings.Contains(upper, marker) {
			return true
		}
	}
	return false
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (g *webGenerator) secretsPlanMJS() string {
	return `// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import fs from "node:fs";
import path from "node:path";

const configPath = path.join(process.cwd(), "security", "secrets.json");
const config = JSON.parse(fs.readFileSync(configPath, "utf8"));
const references = Array.isArray(config.references) ? config.references : [];
const provider = process.env[config.providerEnv] || config.provider || "environment";
const prefix = process.env[config.prefixEnv] || config.app || "blacklang";
const supportedProviders = ["environment", "1password", "vault", "doppler", "aws-secrets-manager"];

const checkedReferences = references.map((ref) => {
  const present = typeof process.env[ref.name] === "string" && process.env[ref.name].length > 0;
  return {
    name: ref.name,
    sources: ref.sources,
    kind: ref.kind,
    required: Boolean(ref.required),
    sensitive: Boolean(ref.sensitive),
    present,
    providerKey: String(prefix) + "/" + String(ref.name)
  };
});
const missingRequired = checkedReferences.filter((ref) => ref.required && !ref.present).map((ref) => ref.name);

console.log(JSON.stringify({
  success: true,
  command: "security:secrets:plan",
  app: config.app,
  target: config.target,
  provider,
  providerKnown: supportedProviders.includes(provider),
  ready: missingRequired.length === 0,
  missingRequired,
  references: checkedReferences,
  policy: config.policy,
  steps: config.steps
}, null, 2));
`
}

func (g *webGenerator) secretsProviderMJS() string {
	return `// Generated by BlackLang. Do not edit manually.

import "dotenv/config";
import { spawnSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

const args = new Set(process.argv.slice(2));
if (!args.has("--preflight")) {
  console.error(JSON.stringify({
    success: false,
    command: "security:secrets:preflight",
    error: "Missing --preflight. Generated secret provider execution is read-only and only supports preflight checks."
  }, null, 2));
  process.exit(1);
}

const configPath = path.join(process.cwd(), "security", "secrets.json");
const config = JSON.parse(fs.readFileSync(configPath, "utf8"));
const references = Array.isArray(config.references) ? config.references : [];
const provider = process.env[config.providerEnv] || config.provider || "environment";
const prefix = process.env[config.prefixEnv] || config.app || "blacklang";

const adapters = {
  environment: {
    label: "environment",
    cli: null,
    versionArgs: [],
    safeCheck: "process-environment-presence"
  },
  "1password": {
    label: "1Password CLI",
    cli: "op",
    versionArgs: ["--version"],
    safeCheck: "cli-availability-only"
  },
  vault: {
    label: "HashiCorp Vault CLI",
    cli: "vault",
    versionArgs: ["version"],
    safeCheck: "cli-availability-only"
  },
  doppler: {
    label: "Doppler CLI",
    cli: "doppler",
    versionArgs: ["--version"],
    safeCheck: "cli-availability-only"
  },
  "aws-secrets-manager": {
    label: "AWS CLI for Secrets Manager",
    cli: "aws",
    versionArgs: ["--version"],
    safeCheck: "cli-availability-only"
  }
};

function commandEnv() {
  const allowed = ["PATH", "Path", "PATHEXT", "SystemRoot", "WINDIR", "HOME", "USERPROFILE"];
  const env = {};
  for (const key of allowed) {
    if (typeof process.env[key] === "string") {
      env[key] = process.env[key];
    }
  }
  return env;
}

function commandAvailable(adapter) {
  if (!adapter || !adapter.cli) {
    return { required: false, available: true, command: null, status: 0, errorCode: null };
  }
  const result = spawnSync(adapter.cli, adapter.versionArgs, {
    cwd: process.cwd(),
    encoding: "utf8",
    shell: false,
    env: commandEnv()
  });
  return {
    required: true,
    available: !result.error && result.status === 0,
    command: adapter.cli,
    args: adapter.versionArgs,
    status: typeof result.status === "number" ? result.status : null,
    errorCode: result.error ? result.error.code || "EXEC_ERROR" : null
  };
}

const adapter = adapters[provider];
const cli = commandAvailable(adapter);
const checkedReferences = references.map((ref) => {
  const present = typeof process.env[ref.name] === "string" && process.env[ref.name].length > 0;
  return {
    name: ref.name,
    sources: ref.sources,
    kind: ref.kind,
    required: Boolean(ref.required),
    sensitive: Boolean(ref.sensitive),
    environmentPresent: present,
    providerKey: String(prefix) + "/" + String(ref.name),
    providerCheck: adapter ? adapter.safeCheck : "unsupported-provider"
  };
});
const missingRequiredEnvironment = checkedReferences.filter((ref) => ref.required && !ref.environmentPresent).map((ref) => ref.name);
const environmentReady = missingRequiredEnvironment.length === 0;
const providerKnown = Boolean(adapter);
const providerReady = providerKnown && cli.available;

console.log(JSON.stringify({
  success: true,
  command: "security:secrets:preflight",
  mode: "preflight",
  app: config.app,
  target: config.target,
  provider,
  providerKnown,
  providerLabel: adapter ? adapter.label : null,
  prefix,
  ready: provider === "environment" ? environmentReady : providerReady,
  environmentReady,
  missingRequiredEnvironment,
  providerExecutable: cli,
  references: checkedReferences,
  policy: config.policy,
  executionPolicy: "read-only provider preflight: generated code verifies provider labels and CLI availability, but never fetches, logs, writes, or injects secret values",
  steps: config.steps
}, null, 2));
`
}
func (g *webGenerator) secretReadmeSection() string {
	return "## Secret References\n\nGenerated secret and environment references are listed without values in `security/secrets.json`.\n\n~~~bash\nnpm run security:secrets:plan\nnpm run security:secrets:preflight\n~~~\n\n`security:secrets:preflight` checks the selected provider label, prefix, and CLI availability without fetching or printing secret values.\n\n"
}
