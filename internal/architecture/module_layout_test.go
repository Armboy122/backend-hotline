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

// ─── S0.5 Boundary tests for the module-first shape ──────────────────────────
//
// These tests guard the SCC-style module-first layout introduced in Phase 0
// (see plan/00-structure-reset.md). They keep the task pilot well-shaped
// while later modules are migrated.

func TestTaskModuleScaffoldExists(t *testing.T) {
	root := projectRoot(t)
	required := []string{
		"internal/modules/task/controller.go",
		"internal/modules/task/service.go",
		"internal/modules/task/repository.go",
		"internal/modules/task/repository_impl.go",
		"internal/modules/task/dto.go",
		"internal/modules/task/errors.go",
	}

	for _, relPath := range required {
		if _, err := os.Stat(filepath.Join(root, relPath)); err != nil {
			t.Fatalf("expected module scaffold file %s to exist: %v", relPath, err)
		}
	}
}

func TestRouterUsesTaskModuleEntryPoint(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "internal/router/router.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read router: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "taskmodule.NewController") {
		t.Fatalf("expected router to use taskmodule.NewController for /v1/tasks")
	}
	if strings.Contains(src, "gormadapter.NewTaskRepository(") {
		t.Fatalf("router still wires /v1/tasks through the old layer-first constructor")
	}
}

// moduleFiles walks internal/modules and returns every non-test .go file along
// with its module-relative basename (e.g. "service.go").
func moduleFiles(t *testing.T) []moduleFile {
	t.Helper()
	root := projectRoot(t)
	target := filepath.Join(root, "internal/modules")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil
	}

	var out []moduleFile
	err := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		out = append(out, moduleFile{
			Path:     path,
			Basename: filepath.Base(path),
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walking internal/modules: %v", err)
	}
	return out
}

type moduleFile struct {
	Path     string
	Basename string
}

func parseImports(t *testing.T, path string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		t.Errorf("parse %s: %v", path, err)
		return nil
	}
	imports := make([]string, 0, len(f.Imports))
	for _, imp := range f.Imports {
		imports = append(imports, strings.Trim(imp.Path.Value, `"`))
	}
	return imports
}

// TestModuleServiceDoesNotImportGin keeps service-layer code transport-agnostic.
// service.go is the SCC-style home for business logic and must be callable
// from HTTP, gRPC, CLI, or tests without a Gin context.
func TestModuleServiceDoesNotImportGin(t *testing.T) {
	root := projectRoot(t)
	for _, mf := range moduleFiles(t) {
		if mf.Basename != "service.go" {
			continue
		}
		for _, imp := range parseImports(t, mf.Path) {
			if strings.Contains(imp, "gin-gonic/gin") || strings.Contains(imp, "internal/models") || strings.Contains(imp, "internal/handlers/") {
				t.Errorf(
					"BOUNDARY VIOLATION: module service file %q imports %q.\n"+
						"  Module services must stay transport-agnostic and persistence-free.\n"+
						"  Move Gin-specific logic to controller.go and persistence wiring to repository_impl.go.",
					rel(root, mf.Path), imp,
				)
			}
		}
	}
}

// TestModuleDTODoesNotImportGORM keeps DTOs lightweight and portable.
func TestModuleDTODoesNotImportGORM(t *testing.T) {
	root := projectRoot(t)
	for _, mf := range moduleFiles(t) {
		if mf.Basename != "dto.go" {
			continue
		}
		for _, imp := range parseImports(t, mf.Path) {
			if strings.Contains(imp, "gorm.io/gorm") {
				t.Errorf(
					"BOUNDARY VIOLATION: module DTO file %q imports %q.\n"+
						"  DTO files must not depend on GORM.\n"+
						"  Keep DTOs pure data shapes.",
					rel(root, mf.Path), imp,
				)
			}
		}
	}
}

// TestModuleControllerDoesNotImportGORM keeps the controller thin: it must
// not reach into the persistence layer directly. Only repository_impl.go is
// allowed to import gorm.io/gorm inside a module folder.
func TestModuleControllerDoesNotImportGORM(t *testing.T) {
	root := projectRoot(t)
	for _, mf := range moduleFiles(t) {
		if mf.Basename != "controller.go" {
			continue
		}
		for _, imp := range parseImports(t, mf.Path) {
			if strings.Contains(imp, "gorm.io/gorm") || strings.Contains(imp, "internal/models") {
				t.Errorf(
					"BOUNDARY VIOLATION: module controller file %q imports %q.\n"+
						"  Controllers must not depend on GORM or persistence models.\n"+
						"  Use the module's repository/service boundary instead.",
					rel(root, mf.Path), imp,
				)
			}
		}
	}
}

// TestModuleRepositoryInterfaceDoesNotLeakGORM verifies that repository.go
// (the module-local persistence contract) does not expose GORM types or
// internal/models. The interface should speak in domain entities or local DTOs.
func TestModuleRepositoryInterfaceDoesNotLeakGORM(t *testing.T) {
	root := projectRoot(t)
	for _, mf := range moduleFiles(t) {
		if mf.Basename != "repository.go" {
			continue
		}
		for _, imp := range parseImports(t, mf.Path) {
			if strings.Contains(imp, "gorm.io/gorm") || strings.Contains(imp, "internal/models") {
				t.Errorf(
					"BOUNDARY VIOLATION: module repository interface %q imports %q.\n"+
						"  Module repository contracts must not expose GORM types or internal/models.\n"+
						"  Use domain entities or local DTOs instead.",
					rel(root, mf.Path), imp,
				)
			}
		}
	}
}

// TestOnlyRepositoryImplImportsGORMInModules ensures gorm.io/gorm is confined
// to repository_impl.go inside any module folder. This keeps every other file
// in the module portable and easy to test without a real database.
func TestOnlyRepositoryImplImportsGORMInModules(t *testing.T) {
	root := projectRoot(t)
	for _, mf := range moduleFiles(t) {
		for _, imp := range parseImports(t, mf.Path) {
			if !strings.Contains(imp, "gorm.io/gorm") {
				continue
			}
			if mf.Basename == "repository_impl.go" {
				continue
			}
			t.Errorf(
				"BOUNDARY VIOLATION: module file %q imports %q.\n"+
					"  Inside internal/modules, only repository_impl.go may depend on GORM.\n"+
					"  Move persistence wiring into repository_impl.go.",
				rel(root, mf.Path), imp,
			)
		}
	}
}
