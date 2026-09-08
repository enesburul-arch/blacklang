package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEcosystemStatusReportsPackagesAdaptersAndPolicies(t *testing.T) {
	result := EcosystemStatus()
	if !result.Success {
		t.Fatalf("expected ecosystem success, got %#v", result.Errors)
	}
	if result.Command != "ecosystem" || result.Release.ArtifactManifest != "release.blackdir" || result.Release.ChecksumFile != "checksums.sha256" {
		t.Fatalf("expected ecosystem identity and release metadata, got %#v", result)
	}
	if result.Release.Trust.SignatureAlgorithm != "ed25519" || result.Release.Trust.SignatureFilePattern != "<artifact>.sig" || !strings.Contains(result.Release.Trust.VerifyCommand, "verify-release-trust.mjs") {
		t.Fatalf("expected signed release trust metadata, got %#v", result.Release.Trust)
	}
	if result.Release.Trust.TransparencyLog.PolicyFile != "packages/registry/release-transparency.blackdir" || result.Release.Trust.TransparencyLog.HashAlgorithm != "sha256" || !containsEcosystemString(result.Release.Trust.TransparencyLog.EntryFields, "entryHash") {
		t.Fatalf("expected release transparency log metadata, got %#v", result.Release.Trust.TransparencyLog)
	}
	if result.Release.Trust.KeyRotation.PolicyFile != "packages/registry/key-rotation-policy.blackdir" || result.Release.Trust.KeyRotation.KeyIDFormat != "sha256-public-key-spki-prefix" || result.Release.Trust.KeyRotation.MinimumOverlapDays != 30 {
		t.Fatalf("expected release key rotation metadata, got %#v", result.Release.Trust.KeyRotation)
	}
	if !containsEcosystemPackage(result.Packages, "npm:blacklang") || !containsEcosystemPackage(result.Packages, "vscode:blacklang-vscode") || !containsEcosystemPackage(result.Packages, "openvsx:blacklang-vscode") || !containsEcosystemPackage(result.Packages, "cursor:blacklang-vscode") {
		t.Fatalf("expected npm wrapper and multi-editor package metadata, got %#v", result.Packages)
	}
	if !containsEcosystemAdapter(result.Adapters, "target:web-react-node-postgres") || !containsEcosystemAdapter(result.Adapters, "target:web-react-node-mysql") || !containsEcosystemAdapter(result.Adapters, "deploy-cloud:docker-provider") || !containsEcosystemAdapter(result.Adapters, "observability:webhook") || !containsEcosystemAdapter(result.Adapters, "observability:otlp") || !containsEcosystemAdapter(result.Adapters, "editor:vscode") || !containsEcosystemAdapter(result.Adapters, "editor:open-vsx") || !containsEcosystemAdapter(result.Adapters, "editor:cursor-compatible") {
		t.Fatalf("expected target, deploy cloud, observability, and editor adapter metadata, got %#v", result.Adapters)
	}
	cloudAdapter, ok := findEcosystemAdapter(result.Adapters, "deploy-cloud:docker-provider")
	if !ok || !containsEcosystemString(cloudAdapter.Capabilities, "explicit apply provider CLI runner") || !containsEcosystemString(cloudAdapter.Capabilities, "read-only cloud preflight script") {
		t.Fatalf("expected deploy cloud adapter to expose provider CLI execution metadata, got %#v", cloudAdapter)
	}
	if !containsEcosystemRegistry(result.Registries, "registry:packages") {
		t.Fatalf("expected package registry metadata, got %#v", result.Registries)
	}
	if !containsEcosystemMarketplace(result.Marketplaces, "marketplace:provider-adapters") {
		t.Fatalf("expected provider adapter marketplace metadata, got %#v", result.Marketplaces)
	}
	if result.PublicIndex.ID != "public-index:blacklang-ecosystem" || result.PublicIndex.Path != "website/ecosystem-index.json" || !containsEcosystemString(result.PublicIndex.Packages, "openvsx:blacklang-vscode") || !containsEcosystemString(result.PublicIndex.Adapters, "editor:cursor-compatible") {
		t.Fatalf("expected public ecosystem index metadata, got %#v", result.PublicIndex)
	}
	if !containsEcosystemString(result.TrustWorkflow.PolicyFiles, "packages/registry/public-index.blackdir") || !containsEcosystemString(result.TrustWorkflow.PolicyFiles, "packages/registry/trust-policy.blackdir") || !containsEcosystemString(result.TrustWorkflow.PolicyFiles, "packages/registry/release-transparency.blackdir") || !containsEcosystemString(result.TrustWorkflow.PolicyFiles, "packages/registry/key-rotation-policy.blackdir") || !containsEcosystemString(result.TrustWorkflow.RequiredChecks, "node packages/registry/scripts/validate-registry.mjs") {
		t.Fatalf("expected trust workflow policy and validation checks, got %#v", result.TrustWorkflow)
	}
	if !containsEcosystemString(result.TrustWorkflow.RequiredChecks, "node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict") {
		t.Fatalf("expected signed release verification check, got %#v", result.TrustWorkflow.RequiredChecks)
	}
	if !strings.Contains(strings.Join(result.Policies, " "), "extensions or adapters") {
		t.Fatalf("expected extension policy, got %#v", result.Policies)
	}
}

func TestFormatEcosystemIR(t *testing.T) {
	ir := FormatEcosystemIR(EcosystemStatus())
	for _, expected := range []string{
		"blackir 0.1",
		"ecosystem ok",
		"release channel dev manifest release.blackdir checksums checksums.sha256",
		"release trust ed25519 pattern <artifact>.sig public-key-env BLACKLANG_RELEASE_PUBLIC_KEY public-key-file-env BLACKLANG_RELEASE_PUBLIC_KEY_FILE",
		"verify node scripts/verify-release-trust.mjs artifacts/releases/<version> --json --strict",
		"release transparency packages/registry/release-transparency.blackdir log transparency.blackdir hash sha256 status prepared-local-policy",
		"field entryHash",
		"release key-rotation packages/registry/key-rotation-policy.blackdir key-id sha256-public-key-spki-prefix overlap-days 30 revocations key-revocations.blackdir status prepared-local-policy",
		"package npm:blacklang npm-wrapper",
		"package openvsx:blacklang-vscode editor-extension-channel",
		"package cursor:blacklang-vscode editor-extension-channel",
		"adapter deploy:docker deploy built-in core",
		"adapter target:web-react-node-mysql target built-in core",
		"adapter deploy-cloud:docker-provider deploy-cloud built-in core",
		"adapter observability:webhook observability built-in core",
		"adapter observability:otlp observability built-in core",
		"adapter editor:vscode editor packageable-source extension",
		"adapter editor:open-vsx editor packageable-source extension",
		"adapter editor:cursor-compatible editor packageable-source extension",
		"registry registry:packages package-registry prepared-local-manifest packages/registry/package-index.blackdir",
		"marketplace marketplace:provider-adapters adapter-marketplace prepared-local-manifest adapters/marketplace/adapter-index.blackdir",
		"public-index public-index:blacklang-ecosystem ecosystem-public-index prepared-local-source website/ecosystem-index.json",
		"source packages/registry/public-index.blackdir",
		"adapter editor:cursor-compatible",
		"policy packages/registry/trust-policy.blackdir",
		"check node packages/registry/scripts/validate-registry.mjs",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected ecosystem IR to contain %q, got:\n%s", expected, ir)
		}
	}
}

func TestEcosystemManifestFilesAreReadable(t *testing.T) {
	root := testRepoRoot(t)
	for _, item := range []struct {
		path     string
		expected []string
	}{
		{
			path: filepath.Join("packages", "registry", "package-index.blackdir"),
			expected: []string{
				"registry blacklang-packages",
				"package npm:blacklang",
				"package vscode:blacklang-vscode",
				"package openvsx:blacklang-vscode",
				"package cursor:blacklang-vscode",
				"trust release-transparency-log true",
				"trust release-key-rotation true",
			},
		},
		{
			path: filepath.Join("packages", "registry", "public-index.blackdir"),
			expected: []string{
				"publicIndex blacklang-ecosystem-public-index",
				"path website/ecosystem-index.json",
				"format json",
				"package openvsx:blacklang-vscode",
				"adapter editor:open-vsx",
				"adapter target:web-react-node-mysql",
				"trust compiler-owned-ecosystem-source true",
				"publish external-deploy-after-release-owner-review",
			},
		},
		{
			path: filepath.Join("packages", "registry", "trust-policy.blackdir"),
			expected: []string{
				"policy package-registry-trust",
				"require release-manifest release.blackdir",
				"require checksum-file checksums.sha256",
				"require detached-signature-pattern \"<artifact>.sig\"",
				"require release-signature-algorithm ed25519",
				"require release-transparency-policy packages/registry/release-transparency.blackdir",
				"require release-key-rotation-policy packages/registry/key-rotation-policy.blackdir",
			},
		},
		{
			path: filepath.Join("packages", "registry", "release-transparency.blackdir"),
			expected: []string{
				"policy release-transparency-log",
				"log transparency.blackdir",
				"hash sha256",
				"append-only true",
				"entry field entryHash",
				"require entry-hash-covers \"release artifact sha256 signature keyId previousEntryHash\"",
			},
		},
		{
			path: filepath.Join("packages", "registry", "key-rotation-policy.blackdir"),
			expected: []string{
				"policy release-key-rotation",
				"key-id-format sha256-public-key-spki-prefix",
				"minimum-overlap-days 30",
				"revocation-file key-revocations.blackdir",
				"require private-key-outside-repository true",
			},
		},
		{
			path: filepath.Join("packages", "registry", "key-revocations.blackdir"),
			expected: []string{
				"revocations release-keys",
				"status prepared-local",
				"No revoked release keys",
			},
		},
		{
			path: filepath.Join("adapters", "marketplace", "adapter-index.blackdir"),
			expected: []string{
				"marketplace provider-adapters",
				"adapter deploy-cloud:docker-provider",
				"adapter observability:webhook",
				"adapter observability:otlp",
				"adapter editor:open-vsx",
				"adapter editor:cursor-compatible",
				"capability ops/observability.json",
				"capability \"VSIX package reuse\"",
				"trust release-transparency-log true",
				"trust release-key-rotation true",
			},
		},
		{
			path: filepath.Join("adapters", "marketplace", "trust-policy.blackdir"),
			expected: []string{
				"policy adapter-marketplace-trust",
				"require env-only-secrets true",
				"require read-only-plan-before-mutation true",
				"require signed-package-verification true",
				"require release-transparency-policy packages/registry/release-transparency.blackdir",
				"require release-key-rotation-policy packages/registry/key-rotation-policy.blackdir",
			},
		},
		{
			path: filepath.Join("website", "ecosystem-index.json"),
			expected: []string{
				"\"id\": \"public-index:blacklang-ecosystem\"",
				"\"path\": \"website/ecosystem-index.json\"",
				"\"openvsx:blacklang-vscode\"",
				"\"editor:cursor-compatible\"",
				"\"target:web-react-node-mysql\"",
				"\"Public index content must contain no provider credentials",
			},
		},
		{
			path: filepath.Join("scripts", "verify-release-trust.mjs"),
			expected: []string{
				"const command = \"release:verify-trust\"",
				"signatureAlgorithm: \"ed25519\"",
				"BLACKLANG_RELEASE_PUBLIC_KEY",
			},
		},
	} {
		content, err := os.ReadFile(filepath.Join(root, item.path))
		if err != nil {
			t.Fatalf("expected readable ecosystem manifest %s: %v", item.path, err)
		}
		text := string(content)
		for _, expected := range item.expected {
			if !strings.Contains(text, expected) {
				t.Fatalf("expected %s to contain %q, got:\n%s", item.path, expected, text)
			}
		}
	}
}

func containsEcosystemPackage(packages []EcosystemPackage, id string) bool {
	for _, pkg := range packages {
		if pkg.ID == id {
			return true
		}
	}
	return false
}

func containsEcosystemAdapter(adapters []EcosystemAdapter, id string) bool {
	_, ok := findEcosystemAdapter(adapters, id)
	return ok
}

func findEcosystemAdapter(adapters []EcosystemAdapter, id string) (EcosystemAdapter, bool) {
	for _, adapter := range adapters {
		if adapter.ID == id {
			return adapter, true
		}
	}
	return EcosystemAdapter{}, false
}

func containsEcosystemRegistry(registries []EcosystemRegistry, id string) bool {
	for _, registry := range registries {
		if registry.ID == id {
			return true
		}
	}
	return false
}

func containsEcosystemMarketplace(marketplaces []EcosystemMarketplace, id string) bool {
	for _, marketplace := range marketplaces {
		if marketplace.ID == id {
			return true
		}
	}
	return false
}

func containsEcosystemString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
