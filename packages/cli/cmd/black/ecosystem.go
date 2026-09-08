package main

import (
	"fmt"
	"strings"
)

func EcosystemStatus() EcosystemResult {
	return EcosystemResult{
		Success: true,
		Command: "ecosystem",
		Version: version,
		Release: EcosystemRelease{
			Channel:          "dev",
			ArtifactManifest: "release.blackdir",
			ChecksumFile:     "checksums.sha256",
			Scripts: []AgentCommand{
				{Name: "build-windows-release", Command: "scripts/build-release-windows.ps1", Purpose: "Build a Windows CLI release archive with manifest.blackdir."},
				{Name: "write-release-checksums", Command: "scripts/write-release-checksums.ps1", Purpose: "Write checksums.sha256 and release.blackdir for finalized release archives."},
				{Name: "verify-release-trust", Command: "node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict", Purpose: "Verify release.blackdir, checksums.sha256, detached Ed25519 signatures, and trusted public key presence before publish."},
			},
			Trust: ReleaseTrust{
				SignatureAlgorithm:   "ed25519",
				SignatureFilePattern: "<artifact>.sig",
				PublicKeyEnv:         "BLACKLANG_RELEASE_PUBLIC_KEY",
				PublicKeyFileEnv:     "BLACKLANG_RELEASE_PUBLIC_KEY_FILE",
				VerifyCommand:        "node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict",
				RequiredBeforePublish: []string{
					"release.blackdir",
					"checksums.sha256",
					"detached Ed25519 signature for every archive",
					"trusted release public key from BLACKLANG_RELEASE_PUBLIC_KEY or BLACKLANG_RELEASE_PUBLIC_KEY_FILE",
					"release transparency log entry for every public archive",
					"key rotation policy check before changing trusted public keys",
				},
				TransparencyLog: ReleaseTransparencyLog{
					PolicyFile:     "packages/registry/release-transparency.blackdir",
					ReleaseLogFile: "transparency.blackdir",
					Status:         "prepared-local-policy",
					HashAlgorithm:  "sha256",
					EntryFields: []string{
						"release",
						"channel",
						"artifact",
						"sha256",
						"signature",
						"keyId",
						"previousEntryHash",
						"entryHash",
						"publishedAt",
					},
					RequiredBeforePublish: []string{
						"append-only transparency.blackdir entry per archive",
						"entry hash covers release, artifact, checksum, signature, key id, and previous entry hash",
						"release.blackdir, checksums.sha256, signature, and transparency entry must agree",
					},
				},
				KeyRotation: ReleaseKeyRotationPolicy{
					PolicyFile:         "packages/registry/key-rotation-policy.blackdir",
					Status:             "prepared-local-policy",
					KeyIDFormat:        "sha256-public-key-spki-prefix",
					MinimumOverlapDays: 30,
					RevocationFile:     "key-revocations.blackdir",
					RequiredBeforePublish: []string{
						"new public key id is listed before it signs public archives",
						"old and new public keys overlap for at least 30 days unless an emergency revocation is declared",
						"revoked key ids are refused for new release entries",
						"private signing keys stay outside .black source, generated output, manifests, and package metadata",
					},
				},
			},
		},
		Packages: []EcosystemPackage{
			{
				ID:       "npm:blacklang",
				Kind:     "npm-wrapper",
				Path:     "packages/npm",
				Status:   "local-package-source",
				Commands: []string{"cd packages/npm && npm test", "BLACKLANG_BINARY=/path/to/black npx blacklang --help"},
			},
			{
				ID:       "vscode:blacklang-vscode",
				Kind:     "editor-extension",
				Path:     "editors/vscode-blacklang",
				Status:   "packageable-source",
				Commands: []string{"cd editors/vscode-blacklang && npm test", "cd editors/vscode-blacklang && npm run package:vsix"},
			},
			{
				ID:       "openvsx:blacklang-vscode",
				Kind:     "editor-extension-channel",
				Path:     "editors/vscode-blacklang",
				Status:   "packageable-source",
				Commands: []string{"cd editors/vscode-blacklang && npm test", "cd editors/vscode-blacklang && npm run package:vsix"},
			},
			{
				ID:       "cursor:blacklang-vscode",
				Kind:     "editor-extension-channel",
				Path:     "editors/vscode-blacklang",
				Status:   "packageable-source",
				Commands: []string{"cd editors/vscode-blacklang && npm test", "cd editors/vscode-blacklang && npm run package:vsix"},
			},
		},
		Adapters: []EcosystemAdapter{
			{
				ID:           "target:web-react-node-sqlite",
				Kind:         "target",
				Name:         "React + Node + SQLite",
				Status:       "built-in",
				Source:       "core",
				Docs:         []string{"docs/target.md", "docs/deployment.md"},
				Capabilities: []string{"generated React frontend", "generated Node API", "SQLite setup/runtime"},
			},
			{
				ID:           "target:web-react-node-postgres",
				Kind:         "target",
				Name:         "React + Node + PostgreSQL",
				Status:       "built-in",
				Source:       "core",
				Docs:         []string{"docs/target.md", "docs/deployment.md"},
				Capabilities: []string{"generated React frontend", "generated Node API", "Prisma PostgreSQL adapter", "PostgreSQL Docker Compose service"},
			},
			{
				ID:           "target:web-react-node-mysql",
				Kind:         "target",
				Name:         "React + Node + MySQL",
				Status:       "built-in",
				Source:       "core",
				Docs:         []string{"docs/target.md", "docs/deployment.md"},
				Capabilities: []string{"generated React frontend", "generated Node API", "Prisma MySQL provider", "PrismaMariaDb adapter", "MySQL Docker Compose service"},
			},
			{
				ID:           "deploy:docker",
				Kind:         "deploy",
				Name:         "Docker deploy target",
				Status:       "built-in",
				Source:       "core",
				Docs:         []string{"docs/deployment.md"},
				Capabilities: []string{"Dockerfile", "docker-compose.yml", "environment references", "healthchecks", "deployment manifest metadata"},
			},
			{
				ID:           "deploy-cloud:docker-provider",
				Kind:         "deploy-cloud",
				Name:         "Docker provider CLI cloud adapter",
				Status:       "built-in",
				Source:       "core",
				Docs:         []string{"docs/deployment.md"},
				Capabilities: []string{"fly provider metadata", "render provider metadata", "railway provider metadata", "deploy/cloud.json", "read-only cloud plan script", "read-only cloud preflight script", "explicit apply provider CLI runner"},
			},
			{
				ID:           "deploy-preview:local",
				Kind:         "deploy-preview",
				Name:         "Local preview stack",
				Status:       "built-in",
				Source:       "core",
				Docs:         []string{"docs/deployment.md"},
				Capabilities: []string{"docker-compose.preview.yml", "isolated preview port", "preview database defaults"},
			},
			{
				ID:           "observability:ops-probes",
				Kind:         "observability",
				Name:         "Generated ops probes",
				Status:       "built-in",
				Source:       "core",
				Docs:         []string{"docs/ops.md"},
				Capabilities: []string{"health endpoint", "readiness endpoint", "metrics endpoint", "request logging"},
			},
			{
				ID:           "observability:webhook",
				Kind:         "observability",
				Name:         "External observability webhook hook",
				Status:       "built-in",
				Source:       "core",
				Docs:         []string{"docs/ops.md"},
				Capabilities: []string{"env-referenced endpoint", "non-blocking request event POST", "W3C traceparent correlation", "OpenAPI observability metadata", "generated local smoke test receiver"},
			},
			{
				ID:           "observability:otlp",
				Kind:         "observability",
				Name:         "OTLP HTTP trace exporter",
				Status:       "built-in",
				Source:       "core",
				Docs:         []string{"docs/ops.md"},
				Capabilities: []string{"ops/observability.json", "env-referenced OTLP HTTP endpoint", "W3C traceparent request propagation", "non-blocking OTLP HTTP JSON trace export", "generated local smoke test receiver"},
			},
			{
				ID:           "editor:vscode",
				Kind:         "editor",
				Name:         "VS Code bridge",
				Status:       "packageable-source",
				Source:       "extension",
				Docs:         []string{"docs/ide.md", "docs/editor-marketplace.md"},
				Capabilities: []string{"syntax highlighting", "compiler-owned completions", "compiler-owned snippets", "compiler-owned diagnostics", "format quick fix", "affected-symbol inspection"},
			},
			{
				ID:           "editor:open-vsx",
				Kind:         "editor",
				Name:         "Open VSX package channel",
				Status:       "packageable-source",
				Source:       "extension",
				Docs:         []string{"docs/ide.md", "docs/editor-marketplace.md"},
				Capabilities: []string{"VSIX package reuse", "compiler-owned completions", "compiler-owned diagnostics", "format quick fix"},
			},
			{
				ID:           "editor:cursor-compatible",
				Kind:         "editor",
				Name:         "Cursor-compatible VSIX channel",
				Status:       "packageable-source",
				Source:       "extension",
				Docs:         []string{"docs/ide.md", "docs/editor-marketplace.md"},
				Capabilities: []string{"VSIX package reuse", "compiler-owned completions", "compiler-owned diagnostics", "format quick fix"},
			},
		},
		Registries: []EcosystemRegistry{
			{
				ID:       "registry:packages",
				Kind:     "package-registry",
				Path:     "packages/registry/package-index.blackdir",
				Status:   "prepared-local-manifest",
				Packages: []string{"npm:blacklang", "vscode:blacklang-vscode", "openvsx:blacklang-vscode", "cursor:blacklang-vscode"},
				Docs:     []string{"docs/package-registry.md", "docs/npm-wrapper.md", "docs/ide.md", "docs/editor-marketplace.md"},
				Commands: []string{"node packages/registry/scripts/validate-registry.mjs", "cd packages/npm && npm test", "cd editors/vscode-blacklang && npm test", "cd editors/vscode-blacklang && npm run package:vsix"},
				Trust: []string{
					"Release artifacts must be listed in release.blackdir and checksums.sha256 before public registry publish.",
					"Release archives, npm wrapper downloads, and editor packages must pass detached Ed25519 signature verification before public install paths are trusted.",
					"Release transparency entries and key rotation policy metadata must be present before public package registry publishing.",
					"npm wrapper downloads are disabled unless BLACKLANG_ALLOW_DOWNLOAD=1 is set by the installer.",
					"Editor bridges must call the native black CLI for diagnostics, completions, snippets, and affected inspection.",
					"VS Code, Open VSX, and Cursor-compatible editor package channels share the same compiler-owned bridge and release trust checks.",
				},
			},
		},
		Marketplaces: []EcosystemMarketplace{
			{
				ID:        "marketplace:provider-adapters",
				Kind:      "adapter-marketplace",
				Path:      "adapters/marketplace/adapter-index.blackdir",
				Status:    "prepared-local-manifest",
				Adapters:  []string{"deploy-cloud:docker-provider", "observability:webhook", "observability:otlp", "editor:vscode", "editor:open-vsx", "editor:cursor-compatible", "target:web-react-node-sqlite", "target:web-react-node-postgres", "target:web-react-node-mysql", "deploy:docker", "deploy-preview:local"},
				Providers: []string{"fly", "render", "railway", "webhook", "otlp", "mysql", "vscode", "openvsx", "cursor-compatible-vsix"},
				Docs:      []string{"docs/adapter-marketplace.md", "docs/deployment.md", "docs/ops.md", "docs/ecosystem.md", "docs/editor-marketplace.md"},
				Commands:  []string{"node packages/registry/scripts/validate-registry.mjs", "black ecosystem --json", "black docs adapter-marketplace --json"},
				Trust: []string{
					"Provider adapters must expose deterministic manifest metadata, read-only preflight, and explicit apply boundaries before any provider CLI execution.",
					"Provider app, region, endpoint, DSN, token, and secret values must stay in environment variables or provider tooling.",
					"Provider adapter packages must use the same detached Ed25519 release trust workflow before external marketplace publishing.",
					"Provider adapter packages must use the same release transparency and key rotation policy metadata before external marketplace publishing.",
					"Marketplace entries must include docs, capabilities, source ownership, status, and verification commands.",
				},
			},
		},
		PublicIndex: EcosystemPublicIndex{
			ID:      "public-index:blacklang-ecosystem",
			Kind:    "ecosystem-public-index",
			Path:    "website/ecosystem-index.json",
			Status:  "prepared-local-source",
			Formats: []string{"json", "blackir"},
			Sources: []string{
				"packages/registry/public-index.blackdir",
				"packages/registry/package-index.blackdir",
				"adapters/marketplace/adapter-index.blackdir",
				"packages/registry/trust-policy.blackdir",
				"adapters/marketplace/trust-policy.blackdir",
			},
			Packages: []string{"npm:blacklang", "vscode:blacklang-vscode", "openvsx:blacklang-vscode", "cursor:blacklang-vscode"},
			Adapters: []string{"deploy-cloud:docker-provider", "observability:webhook", "observability:otlp", "editor:vscode", "editor:open-vsx", "editor:cursor-compatible", "target:web-react-node-sqlite", "target:web-react-node-postgres", "target:web-react-node-mysql"},
			Docs:     []string{"docs/ecosystem.md", "docs/package-registry.md", "docs/adapter-marketplace.md", "docs/editor-marketplace.md", "docs/release-trust.md"},
			Commands: []string{"black ecosystem --json", "black ecosystem --ir", "node packages/registry/scripts/validate-registry.mjs"},
			Trust: []string{
				"Public index source must be generated from compiler-owned ecosystem metadata, not hand-written package/adaptor copies.",
				"Public hosted provider adapter marketplace index publication requires release-owner review and signed release trust checks.",
				"Public index content must contain no provider credentials, marketplace tokens, DSNs, app names, regions, or secret values.",
				"Registry and marketplace manifests must validate locally before any hosted index deployment.",
			},
		},
		TrustWorkflow: EcosystemTrustWorkflow{
			PolicyFiles: []string{
				"packages/registry/public-index.blackdir",
				"packages/registry/trust-policy.blackdir",
				"packages/registry/release-transparency.blackdir",
				"packages/registry/key-rotation-policy.blackdir",
				"adapters/marketplace/trust-policy.blackdir",
			},
			RequiredChecks: []string{
				"black ecosystem --json",
				"black ecosystem --ir",
				"black docs package-registry --json",
				"black docs adapter-marketplace --json",
				"black docs release-trust --json",
				"node packages/registry/scripts/validate-registry.mjs",
				"node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict",
				"cd packages/npm && npm test",
				"cd editors/vscode-blacklang && npm test",
				"cd editors/vscode-blacklang && npm run package:vsix",
			},
			InstallSteps: []string{
				"Read black ecosystem --json and choose a registry or marketplace entry by stable id.",
				"Read the entry docs and policy file before installing or enabling a package/adapter.",
				"Verify local validation commands, release checksums, detached signatures, and public key trust before trusting external downloads.",
				"Keep provider credentials outside .black source and pass them through environment variables or provider tooling.",
			},
		},
		Policies: []string{
			"Provider-specific behavior belongs in extensions or adapters instead of expanding core syntax prematurely.",
			"Each extension or adapter must preserve parser, validator, docs, JSON/BlackIR, diagnostics, and affected-graph discipline before its syntax becomes official.",
			"Installable wrappers must stay thin and forward to the native BlackLang CLI instead of reimplementing the language.",
			"Release artifacts must include checksums, detached Ed25519 signatures, and compact manifest metadata for AI agents and package managers.",
			"Public release trust must include append-only transparency entries and key rotation policy metadata before external publishing.",
			"Multi-editor package channels must reuse compiler-owned IDE output and the same signed VSIX artifact instead of forking syntax rules.",
			"Hosted ecosystem indexes must be generated from compiler-owned metadata and deployed only after release-owner trust review.",
			"Registry and marketplace manifests must be validated locally before any external publish step.",
		},
		Errors: []Diagnostic{},
	}
}

func FormatEcosystemIR(result EcosystemResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("ecosystem ok\n")
	} else {
		builder.WriteString("ecosystem failed\n")
	}
	builder.WriteString(fmt.Sprintf("release channel %s manifest %s checksums %s\n", result.Release.Channel, result.Release.ArtifactManifest, result.Release.ChecksumFile))
	if result.Release.Trust.SignatureAlgorithm != "" {
		builder.WriteString(fmt.Sprintf(
			"release trust %s pattern %s public-key-env %s public-key-file-env %s\n",
			result.Release.Trust.SignatureAlgorithm,
			result.Release.Trust.SignatureFilePattern,
			result.Release.Trust.PublicKeyEnv,
			result.Release.Trust.PublicKeyFileEnv,
		))
		if result.Release.Trust.VerifyCommand != "" {
			builder.WriteString(fmt.Sprintf("  verify %s\n", result.Release.Trust.VerifyCommand))
		}
		for _, requirement := range result.Release.Trust.RequiredBeforePublish {
			builder.WriteString(fmt.Sprintf("  require %s\n", requirement))
		}
		if result.Release.Trust.TransparencyLog.PolicyFile != "" {
			builder.WriteString(fmt.Sprintf(
				"release transparency %s log %s hash %s status %s\n",
				result.Release.Trust.TransparencyLog.PolicyFile,
				result.Release.Trust.TransparencyLog.ReleaseLogFile,
				result.Release.Trust.TransparencyLog.HashAlgorithm,
				result.Release.Trust.TransparencyLog.Status,
			))
			for _, field := range result.Release.Trust.TransparencyLog.EntryFields {
				builder.WriteString(fmt.Sprintf("  field %s\n", field))
			}
			for _, requirement := range result.Release.Trust.TransparencyLog.RequiredBeforePublish {
				builder.WriteString(fmt.Sprintf("  require %s\n", requirement))
			}
		}
		if result.Release.Trust.KeyRotation.PolicyFile != "" {
			builder.WriteString(fmt.Sprintf(
				"release key-rotation %s key-id %s overlap-days %d revocations %s status %s\n",
				result.Release.Trust.KeyRotation.PolicyFile,
				result.Release.Trust.KeyRotation.KeyIDFormat,
				result.Release.Trust.KeyRotation.MinimumOverlapDays,
				result.Release.Trust.KeyRotation.RevocationFile,
				result.Release.Trust.KeyRotation.Status,
			))
			for _, requirement := range result.Release.Trust.KeyRotation.RequiredBeforePublish {
				builder.WriteString(fmt.Sprintf("  require %s\n", requirement))
			}
		}
	}
	builder.WriteString(fmt.Sprintf("packages %d\n", len(result.Packages)))
	for _, pkg := range result.Packages {
		builder.WriteString(fmt.Sprintf("  package %s %s %s %s\n", pkg.ID, pkg.Kind, pkg.Status, pkg.Path))
	}
	builder.WriteString(fmt.Sprintf("adapters %d\n", len(result.Adapters)))
	for _, adapter := range result.Adapters {
		builder.WriteString(fmt.Sprintf("  adapter %s %s %s %s\n", adapter.ID, adapter.Kind, adapter.Status, adapter.Source))
	}
	builder.WriteString(fmt.Sprintf("registries %d\n", len(result.Registries)))
	for _, registry := range result.Registries {
		builder.WriteString(fmt.Sprintf("  registry %s %s %s %s\n", registry.ID, registry.Kind, registry.Status, registry.Path))
	}
	builder.WriteString(fmt.Sprintf("marketplaces %d\n", len(result.Marketplaces)))
	for _, marketplace := range result.Marketplaces {
		builder.WriteString(fmt.Sprintf("  marketplace %s %s %s %s\n", marketplace.ID, marketplace.Kind, marketplace.Status, marketplace.Path))
	}
	if result.PublicIndex.ID != "" {
		builder.WriteString(fmt.Sprintf("public-index %s %s %s %s\n", result.PublicIndex.ID, result.PublicIndex.Kind, result.PublicIndex.Status, result.PublicIndex.Path))
		for _, format := range result.PublicIndex.Formats {
			builder.WriteString(fmt.Sprintf("  format %s\n", format))
		}
		for _, source := range result.PublicIndex.Sources {
			builder.WriteString(fmt.Sprintf("  source %s\n", source))
		}
		for _, pkg := range result.PublicIndex.Packages {
			builder.WriteString(fmt.Sprintf("  package %s\n", pkg))
		}
		for _, adapter := range result.PublicIndex.Adapters {
			builder.WriteString(fmt.Sprintf("  adapter %s\n", adapter))
		}
		for _, command := range result.PublicIndex.Commands {
			builder.WriteString(fmt.Sprintf("  command %s\n", command))
		}
	}
	if len(result.TrustWorkflow.PolicyFiles) > 0 || len(result.TrustWorkflow.RequiredChecks) > 0 || len(result.TrustWorkflow.InstallSteps) > 0 {
		builder.WriteString("trust workflow\n")
		for _, file := range result.TrustWorkflow.PolicyFiles {
			builder.WriteString(fmt.Sprintf("  policy %s\n", file))
		}
		for _, check := range result.TrustWorkflow.RequiredChecks {
			builder.WriteString(fmt.Sprintf("  check %s\n", check))
		}
		for _, step := range result.TrustWorkflow.InstallSteps {
			builder.WriteString(fmt.Sprintf("  step %s\n", step))
		}
	}
	if len(result.Policies) > 0 {
		builder.WriteString("policies\n")
		for _, policy := range result.Policies {
			builder.WriteString(fmt.Sprintf("  - %s\n", policy))
		}
	}
	return builder.String()
}
