package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func PackageProduction(sourceDir string, outDir string) PackageResult {
	if outDir == "" {
		outDir = filepath.Join("artifacts", "production")
	}
	sourceAbs, sourceErr := filepath.Abs(sourceDir)
	outAbs, outErr := filepath.Abs(outDir)
	if sourceErr != nil {
		return packageError(outDir, "PACKAGE_PATH_ERROR", sourceErr)
	}
	if outErr != nil {
		return packageError(outDir, "PACKAGE_PATH_ERROR", outErr)
	}
	if sourceAbs == outAbs || strings.HasPrefix(sourceAbs, outAbs+string(os.PathSeparator)) || strings.HasPrefix(outAbs, sourceAbs+string(os.PathSeparator)) {
		return packageError(outDir, "PACKAGE_OUTPUT_CONFLICT", fmt.Errorf("package output cannot be the generated source directory, its parent, or its child"))
	}

	if err := os.RemoveAll(outDir); err != nil {
		return packageError(outDir, "PACKAGE_CLEAN_ERROR", err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return packageError(outDir, "PACKAGE_CREATE_ERROR", err)
	}

	files := []GeneratedFile{}
	distAssets := referencedDistAssets(sourceDir)
	err := filepath.WalkDir(sourceDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == sourceDir {
			return nil
		}
		name := entry.Name()
		if entry.IsDir() {
			if name == "node_modules" || path == filepath.Join(sourceDir, "src", "generated") {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldExcludeProductionFile(path, sourceDir, distAssets) {
			return nil
		}
		relative, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(outDir, relative)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := copyFile(path, target); err != nil {
			return err
		}
		files = append(files, GeneratedFile{Path: target, Kind: "production-artifact"})
		return nil
	})
	if err != nil {
		return packageError(outDir, "PACKAGE_COPY_ERROR", err)
	}

	return PackageResult{
		Success: true,
		Command: "package",
		Version: version,
		Mode:    "production",
		OutDir:  outDir,
		Files:   files,
		Errors:  []Diagnostic{},
	}
}

func shouldExcludeProductionFile(path string, root string, distAssets map[string]bool) bool {
	name := filepath.Base(path)
	if name == ".env" || strings.HasSuffix(name, ".db") || strings.HasSuffix(name, ".black") || strings.HasSuffix(name, ".black.enc") {
		return true
	}
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	normalized := filepath.ToSlash(relative)
	if strings.HasPrefix(normalized, "dist/assets/") && len(distAssets) > 0 {
		return !distAssets[normalized]
	}
	return relative == "README.md" || relative == ".env.example"
}

func referencedDistAssets(root string) map[string]bool {
	indexPath := filepath.Join(root, "dist", "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return map[string]bool{}
	}
	assets := map[string]bool{}
	matches := regexp.MustCompile(`["']/?(assets/[^"']+)["']`).FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) == 2 {
			assets["dist/"+match[1]] = true
		}
	}
	return assets
}

func copyFile(source string, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(target)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	return err
}

func packageError(outDir string, code string, err error) PackageResult {
	return PackageResult{
		Success: false,
		Command: "package",
		Version: version,
		Mode:    "production",
		OutDir:  outDir,
		Files:   []GeneratedFile{},
		Errors: []Diagnostic{{
			Code:       code,
			Message:    err.Error(),
			Suggestion: "Run `black build` first and check that the output directory is writable.",
		}},
	}
}

func FormatPackageIR(result PackageResult) string {
	var builder strings.Builder
	builder.WriteString("blackir 0.1\n")
	if result.Success {
		builder.WriteString("package ok\n")
	} else {
		builder.WriteString("package failed\n")
	}
	builder.WriteString(fmt.Sprintf("mode %s\n", result.Mode))
	builder.WriteString(fmt.Sprintf("out %s\n", result.OutDir))
	builder.WriteString(fmt.Sprintf("files %d\n", len(result.Files)))
	for _, file := range result.Files {
		builder.WriteString(fmt.Sprintf("file %s %s\n", file.Kind, file.Path))
	}
	return builder.String()
}
