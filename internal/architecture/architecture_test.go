// Package architecture_test contains static boundary checks that enforce the
// Hinghoi-style feature-first rules defined in plan/00-backend-architecture.md.
//
// These tests do NOT connect to any database or start any server.
// They inspect Go source files directly via go/parser.
package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// projectRoot returns the absolute path to the module root by walking up
// two levels from this file's package directory (internal/architecture → internal → root).
func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	// wd == .../backend-hotline/internal/architecture when run with go test
	return filepath.Join(wd, "..", "..")
}

// goSourceFiles walks dir (relative to root) and returns absolute paths of
// all non-test .go files. Returns nil (no error) when the directory does not exist.
func goSourceFiles(t *testing.T, root, dir string) []string {
	t.Helper()
	target := filepath.Join(root, dir)
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil
	}
	var files []string
	if err := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		t.Fatalf("walking %s: %v", target, err)
	}
	return files
}

// importsOf returns the import paths declared in a Go source file.
func importsOf(t *testing.T, path string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		t.Errorf("parse %s: %v", path, err)
		return nil
	}
	var imports []string
	for _, imp := range f.Imports {
		imports = append(imports, strings.Trim(imp.Path.Value, `"`))
	}
	return imports
}

// rel returns path relative to root for readable error messages.
func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return r
}

// ─── A1.1 ────────────────────────────────────────────────────────────────────

// TestUsecasesDoNotImportGin verifies that no production usecase/service file
// imports github.com/gin-gonic/gin.
// Usecases and services must be transport-agnostic so they can be called from HTTP,
// gRPC, CLI, or tests without a real HTTP context.
func TestUsecasesDoNotImportGin(t *testing.T) {
	root := projectRoot(t)
	files := goSourceFiles(t, root, "internal/app")
	files = append(files, goSourceFiles(t, root, "internal/feature")...)
	if len(files) == 0 {
		t.Skip("no usecase/service source files yet")
	}

	for _, f := range files {
		if !strings.Contains(f, string(filepath.Separator)+"service"+string(filepath.Separator)) &&
			!strings.Contains(f, string(filepath.Separator)+"usecase"+string(filepath.Separator)) {
			continue
		}
		for _, imp := range importsOf(t, f) {
			if strings.Contains(imp, "gin-gonic/gin") {
				t.Errorf(
					"BOUNDARY VIOLATION: service/usecase file %q imports Gin.\n"+
						"  Services and usecases must not depend on HTTP transport packages.\n"+
						"  Move Gin-specific logic to the controller layer.",
					rel(root, f),
				)
			}
		}
	}
}

// ─── A1.2 ────────────────────────────────────────────────────────────────────

// TestDomainDoesNotImportFrameworks verifies that no production file under
// internal/domain imports framework or infrastructure packages.
// The domain layer must only depend on the Go standard library and
// small value-semantic libraries (e.g. time, errors).
func TestDomainDoesNotImportFrameworks(t *testing.T) {
	root := projectRoot(t)
	files := goSourceFiles(t, root, "internal/domain")
	if len(files) == 0 {
		t.Skip("internal/domain has no source files yet")
	}

	forbidden := []struct {
		substr string
		reason string
	}{
		{"gin-gonic/gin", "HTTP framework"},
		{"gorm.io/gorm", "ORM / persistence"},
		{"spf13/viper", "config library"},
		{"aws/aws-sdk-go", "AWS SDK / storage"},
		{"internal/handlers", "handler layer (circular)"},
		{"internal/models", "GORM models (persistence layer)"},
		{"internal/adapter", "adapter layer"},
	}

	for _, f := range files {
		for _, imp := range importsOf(t, f) {
			for _, fw := range forbidden {
				if strings.Contains(imp, fw.substr) {
					t.Errorf(
						"BOUNDARY VIOLATION: domain file %q imports %q (%s).\n"+
							"  Domain entities must not depend on infrastructure or framework packages.",
						rel(root, f), imp, fw.reason,
					)
				}
			}
		}
	}
}

// ─── A1.3 ────────────────────────────────────────────────────────────────────

// TestRouterHasNoBusinessQueries verifies that internal/router files do not
// contain inline GORM query calls. The router is composition-only: it wires
// handlers and middleware. Business logic and data access belong in usecases
// and repositories.
func TestRouterHasNoBusinessQueries(t *testing.T) {
	root := projectRoot(t)
	target := filepath.Join(root, "internal/router")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Skip("internal/router does not exist yet")
	}

	// GORM method patterns that indicate a direct database query in the router.
	queryPatterns := []string{
		".Where(",
		".Find(",
		".First(",
		".Create(",
		".Save(",
		".Updates(",
		".Update(",
		".Delete(",
		".Raw(",
		".Exec(",
	}

	var routerFiles []string
	_ = filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		routerFiles = append(routerFiles, path)
		return nil
	})

	for _, f := range routerFiles {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("reading %s: %v", f, err)
			continue
		}
		src := string(content)
		for _, pattern := range queryPatterns {
			if strings.Contains(src, pattern) {
				t.Errorf(
					"BOUNDARY VIOLATION: router file %q contains GORM query pattern %q.\n"+
						"  The router must only wire handlers and middleware.\n"+
						"  Move database access into a repository or usecase.",
					rel(root, f), pattern,
				)
			}
		}
	}
}

// ─── A1.4 ────────────────────────────────────────────────────────────────────

// TestRepositoryInterfacesDoNotLeakGORMModels verifies that port/outbound
// repository interface files do not import internal/models.
// Repository interfaces must express their contracts in terms of domain
// entities or primitive query/result structs, not GORM-tagged persistence
// models. GORM models belong exclusively in the adapter/out/persistence layer.
func TestRepositoryInterfacesDoNotLeakGORMModels(t *testing.T) {
	root := projectRoot(t)
	files := goSourceFiles(t, root, "internal/port/outbound/repository")
	if len(files) == 0 {
		t.Skip("internal/port/outbound/repository has no source files yet")
	}

	for _, f := range files {
		for _, imp := range importsOf(t, f) {
			if strings.Contains(imp, "internal/models") {
				t.Errorf(
					"BOUNDARY VIOLATION: repository interface file %q imports %q.\n"+
						"  Port interfaces must not expose internal/models types.\n"+
						"  Use domain entities (internal/domain/*) or plain structs instead.\n"+
						"  GORM models may only appear in internal/adapter/out/persistence.",
					rel(root, f), imp,
				)
			}
		}
	}
}
