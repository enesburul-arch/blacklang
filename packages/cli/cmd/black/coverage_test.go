package main

import (
	"strings"
	"testing"
)

func TestWebCoverageReportIsWeightedAndActionable(t *testing.T) {
	result := WebCoverageReport()
	if !result.Success {
		t.Fatalf("expected coverage success, got %#v", result.Errors)
	}
	if result.Command != "benchmark coverage" || result.Target != "web" {
		t.Fatalf("expected benchmark coverage web command, got %#v", result)
	}
	if result.CompletionPercent != 100 {
		t.Fatalf("expected current weighted coverage to be 100, got %d", result.CompletionPercent)
	}
	weight := 0
	areas := map[string]CoverageArea{}
	for _, area := range result.Areas {
		weight += area.Weight
		areas[area.Name] = area
		if len(area.Implemented) == 0 || len(area.Missing) == 0 || len(area.Next) == 0 {
			t.Fatalf("expected area %s to be actionable, got %#v", area.Name, area)
		}
	}
	if weight != 100 {
		t.Fatalf("expected coverage weights to sum to 100, got %d", weight)
	}
	if !strings.Contains(strings.Join(areas["auth-permissions-security"].Implemented, " "), ".black.enc") {
		t.Fatalf("expected protected source to count toward security coverage, got %#v", areas["auth-permissions-security"])
	}
	if !strings.Contains(strings.Join(areas["deployment-operations"].Implemented, " "), "health endpoint") {
		t.Fatalf("expected ops health checks to count toward deployment coverage, got %#v", areas["deployment-operations"])
	}
	if !strings.Contains(strings.Join(areas["deployment-operations"].Implemented, " "), "local preview deployment") || !strings.Contains(strings.Join(areas["deployment-operations"].Implemented, " "), "rollback keep metadata") {
		t.Fatalf("expected preview and rollback metadata to count toward deployment coverage, got %#v", areas["deployment-operations"])
	}
	if !strings.Contains(strings.Join(areas["deployment-operations"].Implemented, " "), "OTLP HTTP JSON trace exporter") || strings.Contains(strings.Join(areas["deployment-operations"].Missing, " "), "distributed tracing/exporter plugins") {
		t.Fatalf("expected OTLP trace exporter to count toward deployment coverage, got %#v", areas["deployment-operations"])
	}
	if !strings.Contains(strings.Join(areas["testing-benchmarks"].Implemented, " "), "AI eval result history") || strings.Contains(strings.Join(areas["testing-benchmarks"].Missing, " "), "published multi-model eval result history") {
		t.Fatalf("expected eval history export to count toward testing coverage, got %#v", areas["testing-benchmarks"])
	}
	if !strings.Contains(strings.Join(areas["docs-ai-ergonomics"].Implemented, " "), "multi-editor package channel metadata") || strings.Contains(strings.Join(areas["docs-ai-ergonomics"].Missing, " "), "multi-editor marketplace publishing") {
		t.Fatalf("expected multi-editor channel metadata to count toward docs coverage, got %#v", areas["docs-ai-ergonomics"])
	}
	if !strings.Contains(strings.Join(areas["ecosystem-integrations"].Implemented, " "), "public ecosystem index") || strings.Contains(strings.Join(areas["ecosystem-integrations"].Missing, " "), "public hosted provider adapter marketplace index") {
		t.Fatalf("expected prepared public index source to count toward ecosystem coverage, got %#v", areas["ecosystem-integrations"])
	}
	if !strings.Contains(strings.Join(areas["data-model-crud"].Implemented, " "), "file and image media fields") || strings.Contains(strings.Join(areas["data-model-crud"].Missing, " "), "file/media fields") {
		t.Fatalf("expected media fields to count toward data model coverage, got %#v", areas["data-model-crud"])
	}
	if !strings.Contains(strings.Join(areas["data-model-crud"].Implemented, " "), "relation response load policies") || strings.Contains(strings.Join(areas["data-model-crud"].Missing, " "), "advanced relation loading policies") {
		t.Fatalf("expected relation load policies to count toward data model coverage, got %#v", areas["data-model-crud"])
	}
	if !strings.Contains(strings.Join(areas["data-model-crud"].Implemented, " "), "bulk relation prefetch") || strings.Contains(strings.Join(areas["data-model-crud"].Missing, " "), "bulk relation prefetch") {
		t.Fatalf("expected bulk relation prefetch to count toward data model coverage, got %#v", areas["data-model-crud"])
	}
	if !strings.Contains(strings.Join(areas["auth-permissions-security"].Implemented, " "), "nested relation response field sanitization") {
		t.Fatalf("expected nested relation sanitization to count toward security coverage, got %#v", areas["auth-permissions-security"])
	}
	if !strings.Contains(strings.Join(areas["auth-permissions-security"].Implemented, " "), "secret manager provider preflight execution") || strings.Contains(strings.Join(areas["auth-permissions-security"].Missing, " "), "secret manager provider execution") {
		t.Fatalf("expected secret provider preflight to count toward security coverage, got %#v", areas["auth-permissions-security"])
	}
	if !strings.Contains(strings.Join(areas["backend-api-actions"].Implemented, " "), "background query worker jobs") || strings.Contains(strings.Join(areas["backend-api-actions"].Missing, " "), "background jobs") {
		t.Fatalf("expected background query worker jobs to count toward backend coverage, got %#v", areas["backend-api-actions"])
	}
	if !strings.Contains(strings.Join(areas["backend-api-actions"].Implemented, " "), "API-only target generation") || strings.Contains(strings.Join(areas["backend-api-actions"].Missing, " "), "API-only target") {
		t.Fatalf("expected API-only target generation to count toward backend coverage, got %#v", areas["backend-api-actions"])
	}
	if areas["backend-api-actions"].Score != 100 || !strings.Contains(strings.Join(areas["backend-api-actions"].Implemented, " "), "local value, compound condition, and if/else logic") {
		t.Fatalf("expected action/API local logic to complete backend coverage, got %#v", areas["backend-api-actions"])
	}
	if !strings.Contains(strings.Join(areas["frontend-ui-layout"].Implemented, " "), "reusable component sections") || strings.Contains(strings.Join(areas["frontend-ui-layout"].Missing, " "), "advanced reusable component composition") {
		t.Fatalf("expected reusable component sections to count toward frontend coverage, got %#v", areas["frontend-ui-layout"])
	}
	if !strings.Contains(strings.Join(areas["frontend-ui-layout"].Implemented, " "), "DOM-order rendering") {
		t.Fatalf("expected DOM-order rendering to count toward frontend coverage, got %#v", areas["frontend-ui-layout"])
	}
	if !strings.Contains(strings.Join(areas["frontend-ui-layout"].Implemented, " "), "collection-backed component sections") || strings.Contains(strings.Join(areas["frontend-ui-layout"].Missing, " "), "collection-backed component sections") {
		t.Fatalf("expected collection-backed component sections to count toward frontend coverage, got %#v", areas["frontend-ui-layout"])
	}
	if len(result.Issues) == 0 {
		t.Fatalf("expected tracked issues")
	}
	openIssues := 0
	uiMigrationDone := false
	runtimeI18NDone := false
	expandedI18NDone := false
	uiCopyI18NDone := false
	modalDrawerDone := false
	viewGroupsDone := false
	accessibilityAuditDone := false
	viewTriggersDone := false
	componentSectionsDone := false
	domOrderDone := false
	collectionComponentSectionsDone := false
	migrationPlanDone := false
	postgresRuntimeDone := false
	migrationRuntimeDone := false
	transactionActionDone := false
	transactionBlockDone := false
	migrationRunnerDone := false
	mysqlRuntimeDone := false
	relationLoadDone := false
	relationPrefetchDone := false
	rowPolicyDone := false
	multiRoleDone := false
	tenantAdminDone := false
	secretManifestDone := false
	secretProviderExecutionDone := false
	releaseTrustDone := false
	releaseTransparencyDone := false
	explicitAPIRuntimeDone := false
	explicitAPIHandlerDone := false
	aggregateQueryDone := false
	backgroundJobDone := false
	apiOnlyTargetDone := false
	serviceBlockDone := false
	deployPreviewRollbackDone := false
	opsSignalsDone := false
	cloudAdaptersDone := false
	providerCLIExecutionDone := false
	opsTraceExporterDone := false
	aiTaskBenchmarkDone := false
	seedFixtureDone := false
	browserCheckDone := false
	browserE2EDone := false
	browserMatrixDone := false
	aiEvalCorpusDone := false
	aiEvalHistoryDone := false
	trackedIssueExportDone := false
	ideDiagnosticsDone := false
	ideRefactorDone := false
	editorMarketplaceDone := false
	ecosystemPackageDone := false
	ecosystemRegistryDone := false
	ecosystemSignedTrustDone := false
	ecosystemPublicIndexDone := false
	for _, issue := range result.Issues {
		if issue.ID == "" || issue.Area == "" || issue.Title == "" {
			t.Fatalf("expected issue metadata, got %#v", issue)
		}
		if issue.Status == "open" {
			openIssues++
		}
		if issue.ID == "WEB-UI-001" && issue.Status == "done" {
			uiMigrationDone = true
		}
		if issue.ID == "WEB-I18N-001" && issue.Status == "done" {
			runtimeI18NDone = true
		}
		if issue.ID == "WEB-I18N-002" && issue.Status == "done" {
			expandedI18NDone = true
		}
		if issue.ID == "WEB-I18N-003" && issue.Status == "done" {
			uiCopyI18NDone = true
		}
		if issue.ID == "WEB-LAYOUT-001" && issue.Status == "done" {
			modalDrawerDone = true
		}
		if issue.ID == "WEB-LAYOUT-002" && issue.Status == "done" {
			viewGroupsDone = true
		}
		if issue.ID == "WEB-LAYOUT-003" && issue.Status == "done" {
			accessibilityAuditDone = true
		}
		if issue.ID == "WEB-LAYOUT-004" && issue.Status == "done" {
			viewTriggersDone = true
		}
		if issue.ID == "WEB-LAYOUT-005" && issue.Status == "done" {
			componentSectionsDone = true
		}
		if issue.ID == "WEB-LAYOUT-006" && issue.Status == "done" {
			domOrderDone = true
		}
		if issue.ID == "WEB-LAYOUT-007" && issue.Status == "done" {
			collectionComponentSectionsDone = true
		}
		if issue.ID == "WEB-DATA-001" && issue.Status == "done" {
			migrationPlanDone = true
		}
		if issue.ID == "WEB-DATA-002" && issue.Status == "done" {
			postgresRuntimeDone = true
		}
		if issue.ID == "WEB-DATA-003" && issue.Status == "done" {
			migrationRuntimeDone = true
		}
		if issue.ID == "WEB-DATA-004" && issue.Status == "done" {
			transactionActionDone = true
		}
		if issue.ID == "WEB-DATA-005" && issue.Status == "done" {
			transactionBlockDone = true
		}
		if issue.ID == "WEB-DATA-006" && issue.Status == "done" {
			migrationRunnerDone = true
		}
		if issue.ID == "WEB-DATA-009" && issue.Status == "done" {
			mysqlRuntimeDone = true
		}
		if issue.ID == "WEB-DATA-007" && issue.Status == "done" {
			relationLoadDone = true
		}
		if issue.ID == "WEB-DATA-008" && issue.Status == "done" {
			relationPrefetchDone = true
		}
		if issue.ID == "WEB-SEC-001" && issue.Status == "done" {
			rowPolicyDone = true
		}
		if issue.ID == "WEB-SEC-002" && issue.Status == "done" {
			multiRoleDone = true
		}
		if issue.ID == "WEB-SEC-003" && issue.Status == "done" {
			tenantAdminDone = true
		}
		if issue.ID == "WEB-SEC-004" && issue.Status == "done" {
			secretManifestDone = true
		}
		if issue.ID == "WEB-SEC-007" && issue.Status == "done" {
			secretProviderExecutionDone = true
		}
		if issue.ID == "WEB-SEC-005" && issue.Status == "done" {
			releaseTrustDone = true
		}
		if issue.ID == "WEB-SEC-006" && issue.Status == "done" {
			releaseTransparencyDone = true
		}
		if issue.ID == "WEB-API-001" && issue.Status == "done" {
			explicitAPIRuntimeDone = true
		}
		if issue.ID == "WEB-API-002" && issue.Status == "done" {
			explicitAPIHandlerDone = true
		}
		if issue.ID == "WEB-API-003" && issue.Status == "done" {
			aggregateQueryDone = true
		}
		if issue.ID == "WEB-API-004" && issue.Status == "done" {
			backgroundJobDone = true
		}
		if issue.ID == "WEB-API-005" && issue.Status == "done" {
			apiOnlyTargetDone = true
		}
		if issue.ID == "WEB-API-006" && issue.Status == "done" {
			serviceBlockDone = true
		}
		if issue.ID == "WEB-OPS-001" && issue.Status == "done" {
			deployPreviewRollbackDone = true
		}
		if issue.ID == "WEB-OPS-002" && issue.Status == "done" {
			opsSignalsDone = true
		}
		if issue.ID == "WEB-OPS-003" && issue.Status == "done" {
			cloudAdaptersDone = true
		}
		if issue.ID == "WEB-OPS-004" && issue.Status == "done" {
			providerCLIExecutionDone = true
		}
		if issue.ID == "WEB-OPS-005" && issue.Status == "done" {
			opsTraceExporterDone = true
		}
		if issue.ID == "WEB-TEST-001" && issue.Status == "done" {
			aiTaskBenchmarkDone = true
		}
		if issue.ID == "WEB-TEST-002" && issue.Status == "done" {
			seedFixtureDone = true
		}
		if issue.ID == "WEB-TEST-003" && issue.Status == "done" {
			browserCheckDone = true
		}
		if issue.ID == "WEB-TEST-004" && issue.Status == "done" {
			browserE2EDone = true
		}
		if issue.ID == "WEB-TEST-005" && issue.Status == "done" {
			browserMatrixDone = true
		}
		if issue.ID == "WEB-TEST-006" && issue.Status == "done" {
			aiEvalCorpusDone = true
		}
		if issue.ID == "WEB-TEST-007" && issue.Status == "done" {
			aiEvalHistoryDone = true
		}
		if issue.ID == "WEB-DOCS-002" && issue.Status == "done" {
			trackedIssueExportDone = true
		}
		if issue.ID == "WEB-DOCS-003" && issue.Status == "done" {
			ideDiagnosticsDone = true
		}
		if issue.ID == "WEB-DOCS-004" && issue.Status == "done" {
			ideRefactorDone = true
		}
		if issue.ID == "WEB-DOCS-005" && issue.Status == "done" {
			editorMarketplaceDone = true
		}
		if issue.ID == "WEB-ECO-001" && issue.Status == "done" {
			ecosystemPackageDone = true
		}
		if issue.ID == "WEB-ECO-002" && issue.Status == "done" {
			ecosystemRegistryDone = true
		}
		if issue.ID == "WEB-ECO-003" && issue.Status == "done" {
			ecosystemSignedTrustDone = true
		}
		if issue.ID == "WEB-ECO-004" && issue.Status == "done" {
			ecosystemPublicIndexDone = true
		}
	}
	if openIssues != 0 {
		t.Fatalf("expected no open coverage issues, got %#v", result.Issues)
	}
	if !uiMigrationDone {
		t.Fatalf("expected WEB-UI-001 to be done, got %#v", result.Issues)
	}
	if !runtimeI18NDone {
		t.Fatalf("expected WEB-I18N-001 to be done, got %#v", result.Issues)
	}
	if !expandedI18NDone {
		t.Fatalf("expected WEB-I18N-002 to be done, got %#v", result.Issues)
	}
	if !uiCopyI18NDone {
		t.Fatalf("expected WEB-I18N-003 to be done, got %#v", result.Issues)
	}
	if !modalDrawerDone {
		t.Fatalf("expected WEB-LAYOUT-001 to be done, got %#v", result.Issues)
	}
	if !viewGroupsDone {
		t.Fatalf("expected WEB-LAYOUT-002 to be done, got %#v", result.Issues)
	}
	if !accessibilityAuditDone {
		t.Fatalf("expected WEB-LAYOUT-003 to be done, got %#v", result.Issues)
	}
	if !viewTriggersDone {
		t.Fatalf("expected WEB-LAYOUT-004 to be done, got %#v", result.Issues)
	}
	if !componentSectionsDone {
		t.Fatalf("expected WEB-LAYOUT-005 to be done, got %#v", result.Issues)
	}
	if !domOrderDone {
		t.Fatalf("expected WEB-LAYOUT-006 to be done, got %#v", result.Issues)
	}
	if !collectionComponentSectionsDone {
		t.Fatalf("expected WEB-LAYOUT-007 to be done, got %#v", result.Issues)
	}
	if !migrationPlanDone {
		t.Fatalf("expected WEB-DATA-001 to be done, got %#v", result.Issues)
	}
	if !postgresRuntimeDone {
		t.Fatalf("expected WEB-DATA-002 to be done, got %#v", result.Issues)
	}
	if !migrationRuntimeDone {
		t.Fatalf("expected WEB-DATA-003 to be done, got %#v", result.Issues)
	}
	if !transactionActionDone {
		t.Fatalf("expected WEB-DATA-004 to be done, got %#v", result.Issues)
	}
	if !transactionBlockDone {
		t.Fatalf("expected WEB-DATA-005 to be done, got %#v", result.Issues)
	}
	if !migrationRunnerDone {
		t.Fatalf("expected WEB-DATA-006 to be done, got %#v", result.Issues)
	}
	if !mysqlRuntimeDone {
		t.Fatalf("expected WEB-DATA-009 to be done, got %#v", result.Issues)
	}
	if !relationLoadDone {
		t.Fatalf("expected WEB-DATA-007 to be done, got %#v", result.Issues)
	}
	if !relationPrefetchDone {
		t.Fatalf("expected WEB-DATA-008 to be done, got %#v", result.Issues)
	}
	if !rowPolicyDone {
		t.Fatalf("expected WEB-SEC-001 to be done, got %#v", result.Issues)
	}
	if !multiRoleDone {
		t.Fatalf("expected WEB-SEC-002 to be done, got %#v", result.Issues)
	}
	if !tenantAdminDone {
		t.Fatalf("expected WEB-SEC-003 to be done, got %#v", result.Issues)
	}
	if !secretManifestDone {
		t.Fatalf("expected WEB-SEC-004 to be done, got %#v", result.Issues)
	}
	if !secretProviderExecutionDone {
		t.Fatalf("expected WEB-SEC-007 to be done, got %#v", result.Issues)
	}
	if !releaseTrustDone {
		t.Fatalf("expected WEB-SEC-005 to be done, got %#v", result.Issues)
	}
	if !releaseTransparencyDone {
		t.Fatalf("expected WEB-SEC-006 to be done, got %#v", result.Issues)
	}
	if !explicitAPIRuntimeDone {
		t.Fatalf("expected WEB-API-001 to be done, got %#v", result.Issues)
	}
	if !explicitAPIHandlerDone {
		t.Fatalf("expected WEB-API-002 to be done, got %#v", result.Issues)
	}
	if !aggregateQueryDone {
		t.Fatalf("expected WEB-API-003 to be done, got %#v", result.Issues)
	}
	if !backgroundJobDone {
		t.Fatalf("expected WEB-API-004 to be done, got %#v", result.Issues)
	}
	if !apiOnlyTargetDone {
		t.Fatalf("expected WEB-API-005 to be done, got %#v", result.Issues)
	}
	if !serviceBlockDone {
		t.Fatalf("expected WEB-API-006 to be done, got %#v", result.Issues)
	}
	if !deployPreviewRollbackDone {
		t.Fatalf("expected WEB-OPS-001 to be done, got %#v", result.Issues)
	}
	if !opsSignalsDone {
		t.Fatalf("expected WEB-OPS-002 to be done, got %#v", result.Issues)
	}
	if !cloudAdaptersDone {
		t.Fatalf("expected WEB-OPS-003 to be done, got %#v", result.Issues)
	}
	if !providerCLIExecutionDone {
		t.Fatalf("expected WEB-OPS-004 to be done, got %#v", result.Issues)
	}
	if !opsTraceExporterDone {
		t.Fatalf("expected WEB-OPS-005 to be done, got %#v", result.Issues)
	}
	if !aiTaskBenchmarkDone {
		t.Fatalf("expected WEB-TEST-001 to be done, got %#v", result.Issues)
	}
	if !seedFixtureDone {
		t.Fatalf("expected WEB-TEST-002 to be done, got %#v", result.Issues)
	}
	if !browserCheckDone {
		t.Fatalf("expected WEB-TEST-003 to be done, got %#v", result.Issues)
	}
	if !browserE2EDone {
		t.Fatalf("expected WEB-TEST-004 to be done, got %#v", result.Issues)
	}
	if !browserMatrixDone {
		t.Fatalf("expected WEB-TEST-005 to be done, got %#v", result.Issues)
	}
	if !aiEvalCorpusDone {
		t.Fatalf("expected WEB-TEST-006 to be done, got %#v", result.Issues)
	}
	if !aiEvalHistoryDone {
		t.Fatalf("expected WEB-TEST-007 to be done, got %#v", result.Issues)
	}
	if !trackedIssueExportDone {
		t.Fatalf("expected WEB-DOCS-002 to be done, got %#v", result.Issues)
	}
	if !ideDiagnosticsDone {
		t.Fatalf("expected WEB-DOCS-003 to be done, got %#v", result.Issues)
	}
	if !ideRefactorDone {
		t.Fatalf("expected WEB-DOCS-004 to be done, got %#v", result.Issues)
	}
	if !editorMarketplaceDone {
		t.Fatalf("expected WEB-DOCS-005 to be done, got %#v", result.Issues)
	}
	if !ecosystemPackageDone {
		t.Fatalf("expected WEB-ECO-001 to be done, got %#v", result.Issues)
	}
	if !ecosystemRegistryDone {
		t.Fatalf("expected WEB-ECO-002 to be done, got %#v", result.Issues)
	}
	if !ecosystemSignedTrustDone {
		t.Fatalf("expected WEB-ECO-003 to be done, got %#v", result.Issues)
	}
	if !ecosystemPublicIndexDone {
		t.Fatalf("expected WEB-ECO-004 to be done, got %#v", result.Issues)
	}
}

func TestFormatCoverageIR(t *testing.T) {
	ir := FormatCoverageIR(WebCoverageReport())
	for _, expected := range []string{
		"benchmark coverage ok",
		"target web",
		"completion 100",
		"area frontend-ui-layout",
		"issue WEB-LAYOUT-001",
		"issue WEB-LAYOUT-005",
		"issue WEB-LAYOUT-006",
		"issue WEB-LAYOUT-007",
		"issue WEB-SEC-002",
		"issue WEB-SEC-003",
		"issue WEB-SEC-004",
		"issue WEB-SEC-007",
		"issue WEB-SEC-005",
		"issue WEB-SEC-006",
		"issue WEB-OPS-004",
		"issue WEB-OPS-005",
		"issue WEB-DATA-009",
		"issue WEB-DATA-007",
		"issue WEB-DATA-008",
		"issue WEB-ECO-002",
		"issue WEB-ECO-003",
		"issue WEB-ECO-004",
		"issue WEB-TEST-006",
		"issue WEB-TEST-007",
		"issue WEB-DOCS-005",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected coverage IR to contain %q, got:\n%s", expected, ir)
		}
	}
}

func TestWebCoverageIssuesReport(t *testing.T) {
	result := WebCoverageIssuesReport()
	if !result.Success {
		t.Fatalf("expected coverage issues success, got %#v", result.Errors)
	}
	if result.Command != "benchmark issues" || result.Target != "web" || result.CompletionPercent != 100 {
		t.Fatalf("expected benchmark issues identity, got %#v", result)
	}
	if result.Summary.Total != len(result.Issues) || result.Summary.Open != len(result.OpenIssues) || result.Summary.Done == 0 {
		t.Fatalf("expected issue summary to match issue lists, got %#v", result)
	}
	if result.Summary.Open != 0 {
		t.Fatalf("expected no open issues, got %#v", result)
	}
	if len(result.OpenIssues) != 0 {
		t.Fatalf("expected compact issue export to have no open issues, got %#v", result.OpenIssues)
	}
}

func TestFormatCoverageIssuesIR(t *testing.T) {
	ir := FormatCoverageIssuesIR(WebCoverageIssuesReport())
	for _, expected := range []string{
		"blackir 0.1",
		"benchmark issues ok",
		"target web",
		"completion 100",
		"summary total 54 open 0 done 54",
		"openIssues 0",
	} {
		if !strings.Contains(ir, expected) {
			t.Fatalf("expected coverage issues IR to contain %q, got:\n%s", expected, ir)
		}
	}
}

func coverageIssuesContain(issues []CoverageIssue, id string) bool {
	for _, issue := range issues {
		if issue.ID == id {
			return true
		}
	}
	return false
}
