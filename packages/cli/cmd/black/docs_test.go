package main

import (
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
	for _, keyword := range []string{"docs", "explain", "format", "lint", "syntax", "entity", "page"} {
		if !found[keyword] {
			t.Fatalf("expected docs to include %q", keyword)
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
