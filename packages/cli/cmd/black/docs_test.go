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
	for _, keyword := range []string{"docs", "explain", "format", "lint", "syntax", "entity", "page", "ui", "target", "deploy"} {
		if !found[keyword] {
			t.Fatalf("expected docs to include %q", keyword)
		}
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

func TestFindTargetDoc(t *testing.T) {
	doc, ok := FindDoc("target")
	if !ok {
		t.Fatalf("expected target docs")
	}
	if !strings.Contains(doc.Syntax, "target web") {
		t.Fatalf("expected target docs to mention target web, got %#v", doc)
	}
	if !strings.Contains(strings.Join(doc.Errors, ","), "UNSUPPORTED_TARGET_DATABASE") {
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
		"MISSING_AFFECTED_SYMBOL",
		"UNSUPPORTED_FIELD_TYPE",
		"UNSUPPORTED_TARGET",
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
