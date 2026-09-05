package main

import "testing"

func TestVersionInfo(t *testing.T) {
	result := VersionInfo()
	if !result.Success {
		t.Fatalf("expected version result success, got %#v", result)
	}
	if result.Command != "version" || result.Name != "black" || result.Version != version {
		t.Fatalf("unexpected version result: %#v", result)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no version errors, got %#v", result.Errors)
	}
}
