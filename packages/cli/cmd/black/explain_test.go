package main

import (
	"strings"
	"testing"
)

func TestExplainKeyword(t *testing.T) {
	result := ExplainKeyword("entity")
	if !result.Success {
		t.Fatalf("expected explain success, got %#v", result)
	}
	if result.Keyword != "entity" || result.Purpose == "" || result.Syntax == "" || result.Example == "" {
		t.Fatalf("expected populated entity explanation, got %#v", result)
	}
	if len(result.AgentSteps) == 0 || !strings.Contains(result.AgentSteps[len(result.AgentSteps)-1], "black lint --json") {
		t.Fatalf("expected agent steps to include validation flow, got %#v", result.AgentSteps)
	}
	if !containsExplainString(result.Related, "page") || !containsExplainString(result.Related, "table") {
		t.Fatalf("expected related keywords for entity, got %#v", result.Related)
	}
	if !containsExplainString(result.ErrorCodes, "DUPLICATE_ENTITY") {
		t.Fatalf("expected entity error codes, got %#v", result.ErrorCodes)
	}
}

func TestExplainKeywordDefaultsToSyntax(t *testing.T) {
	result := ExplainKeyword("")
	if !result.Success {
		t.Fatalf("expected default syntax explanation success, got %#v", result)
	}
	if result.Keyword != "syntax" {
		t.Fatalf("expected syntax default, got %#v", result)
	}
}

func TestExplainCustomActionKeyword(t *testing.T) {
	result := ExplainKeyword("action")
	if !result.Success {
		t.Fatalf("expected action explanation success, got %#v", result)
	}
	if result.Keyword != "action" || !strings.Contains(result.Syntax, "set <storedField>") {
		t.Fatalf("expected custom action syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "actions") || !containsExplainString(result.Related, "openapi") {
		t.Fatalf("expected action related keywords, got %#v", result.Related)
	}
	if !containsExplainString(result.ErrorCodes, "PAGE_ACTION_SOURCE_MISMATCH") {
		t.Fatalf("expected action error codes, got %#v", result.ErrorCodes)
	}
	foundStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "Declare one top-level action") {
			foundStep = true
			break
		}
	}
	if !foundStep {
		t.Fatalf("expected action-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainJobKeyword(t *testing.T) {
	result := ExplainKeyword("job")
	if !result.Success {
		t.Fatalf("expected job explanation success, got %#v", result)
	}
	if result.Keyword != "job" || !strings.Contains(result.Syntax, "schedule every") {
		t.Fatalf("expected job syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "query") || !containsExplainString(result.Related, "openapi") {
		t.Fatalf("expected job related keywords, got %#v", result.Related)
	}
	if !containsExplainString(result.ErrorCodes, "UNKNOWN_JOB_QUERY") {
		t.Fatalf("expected job error codes, got %#v", result.ErrorCodes)
	}
	foundStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "jobs/manifest.json") {
			foundStep = true
			break
		}
	}
	if !foundStep {
		t.Fatalf("expected job-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainServiceKeyword(t *testing.T) {
	result := ExplainKeyword("service")
	if !result.Success {
		t.Fatalf("expected service explanation success, got %#v", result)
	}
	if result.Keyword != "service" || !strings.Contains(result.Syntax, "api <APIName>") {
		t.Fatalf("expected service syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "api") || !containsExplainString(result.Related, "openapi") {
		t.Fatalf("expected service related keywords, got %#v", result.Related)
	}
	if !containsExplainString(result.ErrorCodes, "UNKNOWN_SERVICE_API") {
		t.Fatalf("expected service error codes, got %#v", result.ErrorCodes)
	}
	foundStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "services/manifest.json") {
			foundStep = true
			break
		}
	}
	if !foundStep {
		t.Fatalf("expected service-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainTargetKeyword(t *testing.T) {
	result := ExplainKeyword("target")
	if !result.Success {
		t.Fatalf("expected target explanation success, got %#v", result)
	}
	if result.Keyword != "target" || !strings.Contains(result.Syntax, "target api") {
		t.Fatalf("expected target syntax to include API-only target, got %#v", result)
	}
	if !containsExplainString(result.Related, "api") || !containsExplainString(result.ErrorCodes, "UNSUPPORTED_API_TARGET_TEST") {
		t.Fatalf("expected target related docs and diagnostics, got %#v", result)
	}
	foundStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "omitted frontend/browser files") {
			foundStep = true
			break
		}
	}
	if !foundStep {
		t.Fatalf("expected target-specific API-only agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainIndexKeyword(t *testing.T) {
	result := ExplainKeyword("index")
	if !result.Success {
		t.Fatalf("expected index explanation success, got %#v", result)
	}
	if result.Keyword != "index" || !strings.Contains(result.Syntax, "index <storedField>") {
		t.Fatalf("expected index syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "database") || !containsExplainString(result.ErrorCodes, "UNSUPPORTED_COMPUTED_INDEX_FIELD") {
		t.Fatalf("expected index related docs and errors, got %#v", result)
	}
	foundStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "Entity.index") {
			foundStep = true
			break
		}
	}
	if !foundStep {
		t.Fatalf("expected index-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainMigrateKeyword(t *testing.T) {
	result := ExplainKeyword("migrate")
	if !result.Success {
		t.Fatalf("expected migrate explanation success, got %#v", result)
	}
	if result.Keyword != "migrate" || !strings.Contains(result.Syntax, "migrate plan") {
		t.Fatalf("expected migrate syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "database") || !containsExplainString(result.ErrorCodes, "MISSING_SCHEMA_MIGRATION_FILES") {
		t.Fatalf("expected migrate related docs and errors, got %#v", result)
	}
	foundStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "safe=false") {
			foundStep = true
			break
		}
	}
	if !foundStep {
		t.Fatalf("expected migrate-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainAccessibilityKeyword(t *testing.T) {
	result := ExplainKeyword("accessibility")
	if !result.Success {
		t.Fatalf("expected accessibility explanation success, got %#v", result)
	}
	if result.Keyword != "accessibility" || !strings.Contains(result.Syntax, "audit accessibility") {
		t.Fatalf("expected accessibility syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "view") || !containsExplainString(result.ErrorCodes, "ACCESSIBILITY_MISSING_GROUP_TITLE") {
		t.Fatalf("expected accessibility related docs and errors, got %#v", result)
	}
	foundStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "Run black audit accessibility") {
			foundStep = true
			break
		}
	}
	if !foundStep {
		t.Fatalf("expected accessibility-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainOpsKeyword(t *testing.T) {
	result := ExplainKeyword("ops")
	if !result.Success {
		t.Fatalf("expected ops explanation success, got %#v", result)
	}
	if result.Keyword != "ops" || !strings.Contains(result.Syntax, "health path") {
		t.Fatalf("expected ops syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "deploy") || !containsExplainString(result.Related, "openapi") {
		t.Fatalf("expected ops related docs, got %#v", result.Related)
	}
	if !containsExplainString(result.ErrorCodes, "INVALID_OPS_PATH") {
		t.Fatalf("expected ops diagnostics, got %#v", result.ErrorCodes)
	}
	foundStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "inspect --affected ops") {
			foundStep = true
			break
		}
	}
	if !foundStep {
		t.Fatalf("expected ops-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainDeployKeyword(t *testing.T) {
	result := ExplainKeyword("deploy")
	if !result.Success {
		t.Fatalf("expected deploy explanation success, got %#v", result)
	}
	if result.Keyword != "deploy" || !strings.Contains(result.Syntax, "preview local") || !strings.Contains(result.Syntax, "rollback keep") {
		t.Fatalf("expected deploy syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "target") || !containsExplainString(result.ErrorCodes, "INVALID_DEPLOY_ROLLBACK_KEEP") {
		t.Fatalf("expected deploy related docs and errors, got %#v", result)
	}
	foundStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "deploy:rollback:plan") {
			foundStep = true
			break
		}
	}
	if !foundStep {
		t.Fatalf("expected deploy-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainBrowserTestKeyword(t *testing.T) {
	result := ExplainKeyword("test")
	if !result.Success {
		t.Fatalf("expected test explanation success, got %#v", result)
	}
	if result.Keyword != "test" || !strings.Contains(result.Syntax, "expect text") {
		t.Fatalf("expected browser test syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "page") || !containsExplainString(result.Related, "generated-test") {
		t.Fatalf("expected test related docs, got %#v", result.Related)
	}
	if !containsExplainString(result.ErrorCodes, "UNKNOWN_TEST_EXPECT_ACTION") {
		t.Fatalf("expected test diagnostics, got %#v", result.ErrorCodes)
	}
	foundDeclareStep := false
	foundE2EStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "Declare one top-level test") {
			foundDeclareStep = true
		}
		if strings.Contains(step, "npm run test:e2e") {
			foundE2EStep = true
		}
	}
	if !foundDeclareStep || !foundE2EStep {
		t.Fatalf("expected test-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainCoverageKeyword(t *testing.T) {
	result := ExplainKeyword("coverage")
	if !result.Success {
		t.Fatalf("expected coverage explanation success, got %#v", result)
	}
	if result.Keyword != "coverage" || !strings.Contains(result.Syntax, "benchmark issues") {
		t.Fatalf("expected coverage syntax to mention issue export, got %#v", result)
	}
	foundIssueStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "benchmark issues --json") {
			foundIssueStep = true
			break
		}
	}
	if !foundIssueStep {
		t.Fatalf("expected coverage-specific issue export step, got %#v", result.AgentSteps)
	}
}

func TestExplainIDEKeyword(t *testing.T) {
	result := ExplainKeyword("ide")
	if !result.Success {
		t.Fatalf("expected IDE explanation success, got %#v", result)
	}
	if result.Keyword != "ide" || !strings.Contains(result.Syntax, "ide diagnostics") {
		t.Fatalf("expected IDE syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "diagnostics") || !containsExplainString(result.ErrorCodes, "UNKNOWN_IDE_COMMAND") {
		t.Fatalf("expected IDE related docs and diagnostics, got %#v", result)
	}
	foundManifestStep := false
	foundDiagnosticsStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "black ide --json") {
			foundManifestStep = true
		}
		if strings.Contains(step, "black ide diagnostics") {
			foundDiagnosticsStep = true
		}
	}
	if !foundManifestStep || !foundDiagnosticsStep {
		t.Fatalf("expected IDE-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainEcosystemKeyword(t *testing.T) {
	result := ExplainKeyword("ecosystem")
	if !result.Success {
		t.Fatalf("expected ecosystem explanation success, got %#v", result)
	}
	if result.Keyword != "ecosystem" || !strings.Contains(result.Syntax, "black ecosystem") {
		t.Fatalf("expected ecosystem syntax, got %#v", result)
	}
	if !containsExplainString(result.Related, "package") || !containsExplainString(result.Related, "deploy") {
		t.Fatalf("expected ecosystem related docs, got %#v", result.Related)
	}
	foundEcosystemStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "black ecosystem --json") {
			foundEcosystemStep = true
			break
		}
	}
	if !foundEcosystemStep {
		t.Fatalf("expected ecosystem-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainPackageRegistryKeyword(t *testing.T) {
	result := ExplainKeyword("package-registry")
	if !result.Success {
		t.Fatalf("expected package registry explanation success, got %#v", result)
	}
	if result.Keyword != "package-registry" || !containsExplainString(result.Related, "ecosystem") {
		t.Fatalf("expected package registry identity and related docs, got %#v", result)
	}
	foundRegistryStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "packages/registry/package-index.blackdir") {
			foundRegistryStep = true
			break
		}
	}
	if !foundRegistryStep {
		t.Fatalf("expected package registry-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainReleaseTrustKeyword(t *testing.T) {
	result := ExplainKeyword("release-trust")
	if !result.Success {
		t.Fatalf("expected release trust explanation success, got %#v", result)
	}
	if result.Keyword != "release-trust" || !containsExplainString(result.Related, "ecosystem") || !containsExplainString(result.Related, "package-registry") {
		t.Fatalf("expected release trust identity and related docs, got %#v", result)
	}
	foundTrustStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "verify-release-trust.mjs") && strings.Contains(step, "ready=true") {
			foundTrustStep = true
			break
		}
	}
	if !foundTrustStep {
		t.Fatalf("expected release trust-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainAdapterMarketplaceKeyword(t *testing.T) {
	result := ExplainKeyword("adapter-marketplace")
	if !result.Success {
		t.Fatalf("expected adapter marketplace explanation success, got %#v", result)
	}
	if result.Keyword != "adapter-marketplace" || !containsExplainString(result.Related, "ecosystem") {
		t.Fatalf("expected adapter marketplace identity and related docs, got %#v", result)
	}
	foundMarketplaceStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "adapters/marketplace/adapter-index.blackdir") {
			foundMarketplaceStep = true
			break
		}
	}
	if !foundMarketplaceStep {
		t.Fatalf("expected adapter marketplace-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainEditorMarketplaceKeyword(t *testing.T) {
	result := ExplainKeyword("editor-marketplace")
	if !result.Success {
		t.Fatalf("expected editor marketplace explanation success, got %#v", result)
	}
	if result.Keyword != "editor-marketplace" || !containsExplainString(result.Related, "package-registry") || !containsExplainString(result.Related, "release-trust") {
		t.Fatalf("expected editor marketplace identity and related docs, got %#v", result)
	}
	foundEditorMarketplaceDocsStep := false
	foundEditorMarketplaceValidateStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "docs/editor-marketplace.md") {
			foundEditorMarketplaceDocsStep = true
		}
		if strings.Contains(step, "validate-registry.mjs") {
			foundEditorMarketplaceValidateStep = true
		}
	}
	if !foundEditorMarketplaceDocsStep || !foundEditorMarketplaceValidateStep {
		t.Fatalf("expected editor marketplace-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainMediaKeyword(t *testing.T) {
	result := ExplainKeyword("media")
	if !result.Success {
		t.Fatalf("expected media explanation success, got %#v", result)
	}
	if result.Keyword != "media" || !containsExplainString(result.Related, "entity") || !containsExplainString(result.Related, "openapi") {
		t.Fatalf("expected media identity and related docs, got %#v", result)
	}
	foundMediaStep := false
	for _, step := range result.AgentSteps {
		if strings.Contains(step, "x-blacklang-media") && strings.Contains(step, "data URL") {
			foundMediaStep = true
			break
		}
	}
	if !foundMediaStep {
		t.Fatalf("expected media-specific agent steps, got %#v", result.AgentSteps)
	}
}

func TestExplainKeywordReportsUnknownKeyword(t *testing.T) {
	result := ExplainKeyword("missing")
	if result.Success {
		t.Fatalf("expected unknown keyword failure")
	}
	if len(result.Errors) != 1 || result.Errors[0].Code != "UNKNOWN_EXPLAIN_KEYWORD" {
		t.Fatalf("expected UNKNOWN_EXPLAIN_KEYWORD, got %#v", result)
	}
}

func TestFormatExplainIR(t *testing.T) {
	result := ExplainKeyword("lint")
	ir := FormatExplainIR(result)
	for _, expected := range []string{
		"explain ok",
		"keyword lint",
		"agentSteps",
		"related format security",
		"FORMAT_REQUIRED",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected IR to contain %q, got:\n%s", expected, ir)
		}
	}
}

func containsExplainString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
