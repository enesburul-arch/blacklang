package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindDoc(t *testing.T) {
	doc, ok := FindDoc("entity")
	if !ok {
		t.Fatalf("expected entity docs")
	}
	if doc.Keyword != "entity" {
		t.Fatalf("expected entity keyword, got %q", doc.Keyword)
	}
}

func TestFindVersionDoc(t *testing.T) {
	doc, ok := FindDoc("version")
	if !ok {
		t.Fatalf("expected version docs")
	}
	if !strings.Contains(doc.Syntax, "version --json") {
		t.Fatalf("expected version docs to mention json output, got %#v", doc)
	}
}

func TestFindFormatDoc(t *testing.T) {
	doc, ok := FindDoc("format")
	if !ok {
		t.Fatalf("expected format docs")
	}
	if !strings.Contains(doc.Syntax, "format [file]") {
		t.Fatalf("expected format docs to mention file usage, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "FORMAT_REQUIRED") {
		t.Fatalf("expected format docs to mention FORMAT_REQUIRED, got %#v", doc)
	}
}

func TestFindLintDoc(t *testing.T) {
	doc, ok := FindDoc("lint")
	if !ok {
		t.Fatalf("expected lint docs")
	}
	if !strings.Contains(doc.Syntax, "lint [file]") {
		t.Fatalf("expected lint docs to mention file usage, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "FORMAT_REQUIRED") {
		t.Fatalf("expected lint docs to mention FORMAT_REQUIRED, got %#v", doc)
	}
}

func TestAllDocsReturnsSortedEntries(t *testing.T) {
	docs := AllDocs()
	if len(docs) < 10 {
		t.Fatalf("expected many docs entries, got %d", len(docs))
	}
	for index := 1; index < len(docs); index++ {
		if docs[index-1].Keyword > docs[index].Keyword {
			t.Fatalf("expected sorted docs, got %q before %q", docs[index-1].Keyword, docs[index].Keyword)
		}
	}

	found := map[string]bool{}
	for _, doc := range docs {
		found[doc.Keyword] = true
		if doc.Purpose == "" || doc.Syntax == "" {
			t.Fatalf("expected doc %q to include purpose and syntax, got %#v", doc.Keyword, doc)
		}
	}
	for _, keyword := range []string{"accessibility", "action", "adapter-marketplace", "agent-contract", "benchmark", "computed", "coverage", "docs", "editor-marketplace", "ecosystem", "explain", "format", "generated-test", "help", "ide", "index", "job", "lint", "media", "message", "migrate", "migration", "ops", "package-registry", "placeholder", "release-trust", "service", "syntax", "entity", "page", "test", "view", "ui", "target", "deploy", "theme-migration"} {
		if !found[keyword] {
			t.Fatalf("expected docs to include %q", keyword)
		}
	}
}

func TestFindServiceDoc(t *testing.T) {
	doc, ok := FindDoc("service")
	if !ok {
		t.Fatalf("expected service docs")
	}
	notes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(doc.Syntax, "api <APIName>") || !strings.Contains(doc.Example, "service InventoryIntegration") || !strings.Contains(notes, "services/manifest.json") || !strings.Contains(notes, "x-blacklang-service") {
		t.Fatalf("expected service docs to explain API grouping and generated metadata, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNKNOWN_SERVICE_API") {
		t.Fatalf("expected service docs to mention service diagnostics, got %#v", doc.Errors)
	}
}

func TestFindJobDoc(t *testing.T) {
	doc, ok := FindDoc("job")
	if !ok {
		t.Fatalf("expected job docs")
	}
	notes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(doc.Syntax, "schedule every") || !strings.Contains(doc.Example, "run query LowStockProducts") || !strings.Contains(notes, "jobs:run") || !strings.Contains(notes, "future adapter work") {
		t.Fatalf("expected job docs to explain schedule, query worker, scripts, and MVP limits, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNKNOWN_JOB_QUERY") {
		t.Fatalf("expected job docs to mention job diagnostics, got %#v", doc.Errors)
	}
}

func TestFindMediaDoc(t *testing.T) {
	doc, ok := FindDoc("media")
	if !ok {
		t.Fatalf("expected media docs")
	}
	notes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(doc.Syntax, "file|image") || !strings.Contains(doc.Example, `accept "image/*"`) || !strings.Contains(notes, "data URL") || !strings.Contains(notes, "x-blacklang-media") {
		t.Fatalf("expected media docs to explain file/image fields, data URL storage, and OpenAPI metadata, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNSUPPORTED_IMAGE_ACCEPT") {
		t.Fatalf("expected media docs to mention image accept diagnostics, got %#v", doc.Errors)
	}
}

func TestFindEcosystemDoc(t *testing.T) {
	doc, ok := FindDoc("ecosystem")
	if !ok {
		t.Fatalf("expected ecosystem docs")
	}
	if !strings.Contains(doc.Syntax, "black ecosystem") {
		t.Fatalf("expected ecosystem docs to show CLI syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "packages/npm") || !strings.Contains(strings.Join(doc.AgentNotes, " "), "extensions or adapters") {
		t.Fatalf("expected ecosystem docs to explain package and adapter discovery, got %#v", doc)
	}
}

func TestFindPackageRegistryDoc(t *testing.T) {
	doc, ok := FindDoc("package-registry")
	if !ok {
		t.Fatalf("expected package registry docs")
	}
	notes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(doc.Syntax, "package-index.blackdir") || !strings.Contains(notes, "openvsx:blacklang-vscode") || !strings.Contains(notes, "validate-registry.mjs") || !strings.Contains(notes, "native black CLI") {
		t.Fatalf("expected package registry docs to explain manifests and trust checks, got %#v", doc)
	}
}

func TestFindReleaseTrustDoc(t *testing.T) {
	doc, ok := FindDoc("release-trust")
	if !ok {
		t.Fatalf("expected release trust docs")
	}
	notes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(doc.Syntax, "verify-release-trust.mjs") || !strings.Contains(notes, "Ed25519") || !strings.Contains(notes, "BLACKLANG_RELEASE_PUBLIC_KEY") {
		t.Fatalf("expected release trust docs to explain signed release verification, got %#v", doc)
	}
}

func TestFindAdapterMarketplaceDoc(t *testing.T) {
	doc, ok := FindDoc("adapter-marketplace")
	if !ok {
		t.Fatalf("expected adapter marketplace docs")
	}
	notes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(doc.Syntax, "adapter-index.blackdir") || !strings.Contains(notes, "editor:open-vsx") || !strings.Contains(notes, "Provider credentials") || !strings.Contains(notes, "read-only manifest") {
		t.Fatalf("expected adapter marketplace docs to explain manifests and provider trust checks, got %#v", doc)
	}
}

func TestFindEditorMarketplaceDoc(t *testing.T) {
	doc, ok := FindDoc("editor-marketplace")
	if !ok {
		t.Fatalf("expected editor marketplace docs")
	}
	notes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(doc.Syntax, "package-index.blackdir") || !strings.Contains(notes, "openvsx:blacklang-vscode") || !strings.Contains(notes, "editor:cursor-compatible") || !strings.Contains(notes, "release.blackdir") {
		t.Fatalf("expected editor marketplace docs to explain multi-editor channels and trust checks, got %#v", doc)
	}
}

func TestFindIDEDoc(t *testing.T) {
	doc, ok := FindDoc("ide")
	if !ok {
		t.Fatalf("expected IDE docs")
	}
	if !strings.Contains(doc.Syntax, "black ide") || !strings.Contains(doc.Syntax, "ide diagnostics") {
		t.Fatalf("expected IDE docs to show manifest and diagnostics syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "zero-based") || !strings.Contains(strings.Join(doc.Errors, ","), "UNKNOWN_IDE_COMMAND") {
		t.Fatalf("expected IDE docs to explain editor diagnostics and command errors, got %#v", doc)
	}
}

func TestFindAccessibilityDoc(t *testing.T) {
	doc, ok := FindDoc("accessibility")
	if !ok {
		t.Fatalf("expected accessibility docs")
	}
	if !strings.Contains(doc.Syntax, "audit accessibility") {
		t.Fatalf("expected accessibility docs to mention audit syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "ACCESSIBILITY_MISSING_OVERLAY_TITLE") {
		t.Fatalf("expected accessibility docs to mention accessibility diagnostics, got %#v", doc)
	}
}

func TestFindCoverageDoc(t *testing.T) {
	doc, ok := FindDoc("coverage")
	if !ok {
		t.Fatalf("expected coverage docs")
	}
	if !strings.Contains(doc.Syntax, "benchmark coverage") || !strings.Contains(doc.Syntax, "benchmark issues") {
		t.Fatalf("expected coverage docs to mention coverage and issue export syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "weighted progress") || !strings.Contains(strings.Join(doc.AgentNotes, " "), "open issue") {
		t.Fatalf("expected coverage docs to explain completion percent, got %#v", doc)
	}
}

func TestFindBenchmarkDoc(t *testing.T) {
	doc, ok := FindDoc("benchmark")
	if !ok {
		t.Fatalf("expected benchmark docs")
	}
	if !strings.Contains(doc.Syntax, "black benchmark") || !strings.Contains(doc.Syntax, "benchmark tasks") || !strings.Contains(doc.Syntax, "benchmark issues") || !strings.Contains(doc.Example, "--json") {
		t.Fatalf("expected benchmark docs to show CLI syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "temporary directory") || !strings.Contains(strings.Join(doc.AgentNotes, " "), "token estimate") || !strings.Contains(strings.Join(doc.AgentNotes, " "), "tracked issue export") {
		t.Fatalf("expected benchmark docs to mention side-effect-free measurement, got %#v", doc)
	}
}

func TestFindGeneratedTestDoc(t *testing.T) {
	doc, ok := FindDoc("generated-test")
	if !ok {
		t.Fatalf("expected generated-test docs")
	}
	if !strings.Contains(doc.Syntax, "npm test") || !strings.Contains(doc.Syntax, "test:e2e") || !strings.Contains(doc.Example, "black build") {
		t.Fatalf("expected generated-test docs to show build and npm test flow, got %#v", doc)
	}
	joinedNotes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(joinedNotes, "contract.test.ts") || !strings.Contains(joinedNotes, "api.test.ts") || !strings.Contains(joinedNotes, "blacklang.e2e.matrix.ts") {
		t.Fatalf("expected generated-test docs to mention contract coverage, got %#v", doc)
	}
}

func TestFindBrowserTestDoc(t *testing.T) {
	doc, ok := FindDoc("test")
	if !ok {
		t.Fatalf("expected test docs")
	}
	if !strings.Contains(doc.Syntax, "expect text") || !strings.Contains(doc.Example, "WarehouseBrowserSmoke") {
		t.Fatalf("expected browser test docs to show syntax, got %#v", doc)
	}
	joinedNotes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(joinedNotes, "src/blacklang.browser.test.tsx") || !strings.Contains(joinedNotes, "src/blacklang.e2e.test.ts") || !strings.Contains(joinedNotes, "src/blacklang.e2e.matrix.ts") {
		t.Fatalf("expected browser test docs to mention generated browser test files, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNKNOWN_TEST_EXPECT_ACTION") {
		t.Fatalf("expected browser test docs to mention test diagnostics, got %#v", doc)
	}
}

func TestFindOpsDoc(t *testing.T) {
	doc, ok := FindDoc("ops")
	if !ok {
		t.Fatalf("expected ops docs")
	}
	joinedNotes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(doc.Syntax, "health path") || !strings.Contains(doc.Syntax, "observe webhook|otlp") || !strings.Contains(joinedNotes, "readiness") || !strings.Contains(joinedNotes, "traceparent") || !strings.Contains(joinedNotes, "observability") || !strings.Contains(joinedNotes, "Dockerfile") {
		t.Fatalf("expected ops docs to mention health/readiness/Docker behavior, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "INVALID_OPS_PATH") || !strings.Contains(strings.Join(doc.Errors, ","), "INVALID_OPS_OBSERVE") {
		t.Fatalf("expected ops docs to mention ops diagnostics, got %#v", doc)
	}
}

func TestFindDeployDoc(t *testing.T) {
	doc, ok := FindDoc("deploy")
	if !ok {
		t.Fatalf("expected deploy docs")
	}
	joinedNotes := strings.Join(doc.AgentNotes, " ")
	if !strings.Contains(doc.Syntax, "preview local") || !strings.Contains(doc.Syntax, "rollback keep") || !strings.Contains(doc.Syntax, "cloud") {
		t.Fatalf("expected deploy docs to mention preview, rollback, and cloud syntax, got %#v", doc)
	}
	if !strings.Contains(joinedNotes, "docker-compose.preview.yml") || !strings.Contains(joinedNotes, "deploy/rollback.json") || !strings.Contains(joinedNotes, "deploy/cloud.json") {
		t.Fatalf("expected deploy docs to mention generated preview, rollback, and cloud files, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "INVALID_DEPLOY_ROLLBACK_KEEP") || !strings.Contains(strings.Join(doc.Errors, ","), "INVALID_DEPLOY_CLOUD") {
		t.Fatalf("expected deploy docs to mention rollback and cloud diagnostics, got %#v", doc)
	}
}

func TestFindIndexDoc(t *testing.T) {
	doc, ok := FindDoc("index")
	if !ok {
		t.Fatalf("expected index docs")
	}
	if !strings.Contains(doc.Syntax, "index <storedField>") {
		t.Fatalf("expected index docs to mention index syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "Computed display fields cannot be indexed") {
		t.Fatalf("expected index docs to reject computed indexes, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNKNOWN_INDEX_FIELD") {
		t.Fatalf("expected index docs to mention index diagnostics, got %#v", doc)
	}
}

func TestFindMigrateDoc(t *testing.T) {
	doc, ok := FindDoc("migrate")
	if !ok {
		t.Fatalf("expected migrate docs")
	}
	if !strings.Contains(doc.Syntax, "migrate plan") {
		t.Fatalf("expected migrate docs to mention plan syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "read-only") {
		t.Fatalf("expected migrate docs to mention read-only behavior, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "MISSING_SCHEMA_MIGRATION_FILES") {
		t.Fatalf("expected migrate docs to mention missing-file diagnostic, got %#v", doc)
	}
}

func TestFindMigrationDoc(t *testing.T) {
	doc, ok := FindDoc("migration")
	if !ok {
		t.Fatalf("expected migration docs")
	}
	if !strings.Contains(doc.Syntax, "rename entity") || !strings.Contains(doc.Syntax, "rename field") {
		t.Fatalf("expected migration docs to mention rename syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "rename-table or rename-column") {
		t.Fatalf("expected migration docs to mention migration plan rename output, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "DUPLICATE_MIGRATION_RENAME") {
		t.Fatalf("expected migration docs to mention migration diagnostics, got %#v", doc)
	}
}

func TestFindSecurityDocMentionsProtectedSourceEncryption(t *testing.T) {
	doc, ok := FindDoc("security")
	if !ok {
		t.Fatalf("expected security docs")
	}
	if !strings.Contains(doc.Syntax, "security encrypt") || !strings.Contains(doc.Syntax, "security decrypt") {
		t.Fatalf("expected security docs to mention encrypt/decrypt syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), ".black.enc source in memory") {
		t.Fatalf("expected security docs to mention encrypted in-memory source flow, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "MISSING_DECRYPT_STDOUT") {
		t.Fatalf("expected security docs to mention decrypt diagnostics, got %#v", doc)
	}
}

func TestFindActionDoc(t *testing.T) {
	doc, ok := FindDoc("action")
	if !ok {
		t.Fatalf("expected action docs")
	}
	if !strings.Contains(doc.Syntax, "action <Name>") || !strings.Contains(doc.Example, "RestockProduct") {
		t.Fatalf("expected action docs to show custom action syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "POST /api/<lowercase-page>/:id/actions/<lowercase-action>") {
		t.Fatalf("expected action docs to mention generated route contract, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "ACTION_INPUT_FIELD_COLLISION") {
		t.Fatalf("expected action docs to mention action diagnostics, got %#v", doc)
	}
}

func TestFindInspectDoc(t *testing.T) {
	doc, ok := FindDoc("inspect")
	if !ok {
		t.Fatalf("expected inspect docs")
	}
	if !strings.Contains(doc.Syntax, "--affected <symbol>") {
		t.Fatalf("expected inspect docs to mention --affected, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNKNOWN_AFFECTED_SYMBOL") {
		t.Fatalf("expected inspect docs to mention UNKNOWN_AFFECTED_SYMBOL, got %#v", doc)
	}
}

func TestFindDiagnosticsDoc(t *testing.T) {
	doc, ok := FindDoc("diagnostics")
	if !ok {
		t.Fatalf("expected diagnostics docs")
	}
	if !strings.Contains(doc.Syntax, "docs/diagnostics.md") {
		t.Fatalf("expected diagnostics docs to mention docs/diagnostics.md, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "HARDCODED_TOKEN") {
		t.Fatalf("expected diagnostics docs to mention source-security codes, got %#v", doc)
	}
}

func TestFindAgentDoc(t *testing.T) {
	doc, ok := FindDoc("agent")
	if !ok {
		t.Fatalf("expected agent docs")
	}
	if !strings.Contains(doc.Syntax, "agent startup") {
		t.Fatalf("expected agent docs to mention startup, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNKNOWN_AGENT_COMMAND") {
		t.Fatalf("expected agent docs to mention UNKNOWN_AGENT_COMMAND, got %#v", doc)
	}
}

func TestFindAgentContractDoc(t *testing.T) {
	doc, ok := FindDoc("agent-contract")
	if !ok {
		t.Fatalf("expected agent-contract docs")
	}
	if !strings.Contains(doc.Purpose, "capability boundary") {
		t.Fatalf("expected agent-contract docs to mention capability boundary, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "text/black") {
		t.Fatalf("expected agent-contract docs to mention unsupported text/black runtime, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "source bootstrap") {
		t.Fatalf("expected agent-contract docs to mention source bootstrap, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "validate rightValue != 0") {
		t.Fatalf("expected agent-contract docs to warn about unsupported literal validation, got %#v", doc)
	}
}

func TestFindThemeDoc(t *testing.T) {
	doc, ok := FindDoc("theme")
	if !ok {
		t.Fatalf("expected theme docs")
	}
	if !strings.Contains(doc.Syntax, "theme inspect") {
		t.Fatalf("expected theme docs to mention inspect, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "INVALID_UI_MODE") {
		t.Fatalf("expected theme docs to mention UI diagnostics, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "NON_APPEND_ONLY_UI_SLOT") {
		t.Fatalf("expected theme docs to mention lock diagnostics, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UI_SLOT_MIGRATION_BREAK") {
		t.Fatalf("expected theme docs to mention migration diagnostics, got %#v", doc)
	}
}

func TestFindThemeMigrationDoc(t *testing.T) {
	doc, ok := FindDoc("theme-migration")
	if !ok {
		t.Fatalf("expected theme-migration docs")
	}
	if !strings.Contains(doc.Syntax, "theme migrate") {
		t.Fatalf("expected theme-migration docs to mention migrate syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.AgentNotes, " "), "exact prefix") {
		t.Fatalf("expected theme-migration docs to explain prefix compatibility, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UI_SLOT_MIGRATION_BREAK") {
		t.Fatalf("expected theme-migration docs to mention slot break diagnostics, got %#v", doc)
	}
}

func TestFindUIProfileDoc(t *testing.T) {
	doc, ok := FindDoc("ui-profile")
	if !ok {
		t.Fatalf("expected ui-profile docs")
	}
	if !strings.Contains(doc.Syntax, "ui <mode> = <slot...>") {
		t.Fatalf("expected ui-profile docs to mention generator ui slot order, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "DUPLICATE_UI_SLOT") {
		t.Fatalf("expected ui-profile docs to mention duplicate slot diagnostics, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "MISSING_UI_LOCK_BASELINE") {
		t.Fatalf("expected ui-profile docs to mention lock baseline diagnostics, got %#v", doc)
	}
}

func TestFindUIModesDoc(t *testing.T) {
	doc, ok := FindDoc("ui-modes")
	if !ok {
		t.Fatalf("expected ui-modes docs")
	}
	if !strings.Contains(doc.Syntax, "ui box") {
		t.Fatalf("expected ui-modes docs to mention standard modes, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "MISSING_STANDARD_UI_MODE") {
		t.Fatalf("expected ui-modes docs to mention missing standard mode diagnostics, got %#v", doc)
	}
}

func TestFindUIDoc(t *testing.T) {
	doc, ok := FindDoc("ui")
	if !ok {
		t.Fatalf("expected ui docs")
	}
	if !strings.Contains(doc.Syntax, "ui <mode> <values...>") {
		t.Fatalf("expected ui docs to mention inline syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNKNOWN_ACTION_UI") {
		t.Fatalf("expected ui docs to mention action diagnostics, got %#v", doc)
	}
}

func TestFindViewDoc(t *testing.T) {
	doc, ok := FindDoc("view")
	if !ok {
		t.Fatalf("expected view docs")
	}
	if !strings.Contains(doc.Syntax, "view { order") {
		t.Fatalf("expected view docs to mention order syntax, got %#v", doc)
	}
	if !strings.Contains(doc.Syntax, "compose stack|grid|tabs") || !strings.Contains(strings.Join(doc.AgentNotes, " "), "tab <Name> sections") {
		t.Fatalf("expected view docs to mention tabs syntax, got %#v", doc)
	}
	if !strings.Contains(doc.Syntax, "display inline|modal|drawer") || !strings.Contains(doc.Example, "display drawer") {
		t.Fatalf("expected view docs to mention modal/drawer display syntax, got %#v", doc)
	}
	if !strings.Contains(doc.Syntax, "group <Name> sections") || !strings.Contains(strings.Join(doc.AgentNotes, " "), "wraps contiguous inline sections") {
		t.Fatalf("expected view docs to mention group syntax, got %#v", doc)
	}
	if !strings.Contains(doc.Syntax, "trigger <section> on") || !strings.Contains(strings.Join(doc.AgentNotes, " "), "rowSelect") {
		t.Fatalf("expected view docs to mention trigger syntax, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNSUPPORTED_VIEW_SECTION") {
		t.Fatalf("expected view docs to mention view diagnostics, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNSUPPORTED_VIEW_SECTION_DISPLAY") {
		t.Fatalf("expected view docs to mention section display diagnostics, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNSUPPORTED_VIEW_GROUP_SECTION") {
		t.Fatalf("expected view docs to mention group diagnostics, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "MISSING_VIEW_TAB_SECTION") {
		t.Fatalf("expected view docs to mention tab diagnostics, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNSUPPORTED_VIEW_TRIGGER") {
		t.Fatalf("expected view docs to mention trigger diagnostics, got %#v", doc)
	}
}

func TestFindTargetDoc(t *testing.T) {
	doc, ok := FindDoc("target")
	if !ok {
		t.Fatalf("expected target docs")
	}
	if !strings.Contains(doc.Syntax, "target web") || !strings.Contains(doc.Syntax, "target api") {
		t.Fatalf("expected target docs to mention web and API targets, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNSUPPORTED_TARGET_DATABASE") || !strings.Contains(strings.Join(doc.Errors, ","), "UNSUPPORTED_API_TARGET_TEST") {
		t.Fatalf("expected target docs to mention target diagnostics, got %#v", doc)
	}
}

func TestDiagnosticsReferenceMentionsCoreCodes(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "docs", "diagnostics.md"))
	if err != nil {
		t.Fatalf("expected diagnostics reference to be readable: %v", err)
	}
	text := string(content)
	for _, code := range []string{
		"UNKNOWN_TABLE_COLUMN",
		"FORMAT_REQUIRED",
		"HARDCODED_TOKEN",
		"MISSING_ENCRYPTION_KEY",
		"MISSING_DECRYPT_STDOUT",
		"UNSUPPORTED_FORMAT_ENCRYPTED_SOURCE",
		"MISSING_AFFECTED_SYMBOL",
		"UNSUPPORTED_FIELD_TYPE",
		"UNSUPPORTED_TARGET",
		"UNSUPPORTED_API_TARGET_TEST",
		"UNSUPPORTED_VIEW_SECTION",
		"UNKNOWN_ACTION_VALUE",
		"UNKNOWN_INDEX_FIELD",
		"UNSUPPORTED_COMPUTED_INDEX_FIELD",
		"MISSING_SCHEMA_MIGRATION_FILES",
	} {
		if !strings.Contains(text, code) {
			t.Fatalf("expected diagnostics reference to mention %s", code)
		}
	}
}

func TestFormatDocsIR(t *testing.T) {
	doc, ok := FindDoc("table")
	if !ok {
		t.Fatalf("expected table docs")
	}
	ir := FormatDocsIR(DocsResult{
		Success: true,
		Command: "docs",
		Version: version,
		Doc:     doc,
		Errors:  []Diagnostic{},
	})

	expected := []string{
		"docs ok",
		"keyword table",
		"purpose",
		"syntax",
		"UNKNOWN_TABLE_COLUMN",
	}
	for _, value := range expected {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected IR to contain %q, got:\n%s", value, ir)
		}
	}
}

func TestFormatDocsAllIR(t *testing.T) {
	docs := AllDocs()
	ir := FormatDocsAllIR(DocsAllResult{
		Success: true,
		Command: "docs",
		Version: version,
		Count:   len(docs),
		Docs:    docs,
		Errors:  []Diagnostic{},
	})

	expected := []string{
		"docs all ok",
		"count ",
		"keyword entity",
		"keyword lint",
		"FORMAT_REQUIRED",
	}
	for _, value := range expected {
		if !strings.Contains(ir, value) {
			t.Fatalf("expected IR to contain %q, got:\n%s", value, ir)
		}
	}
}
