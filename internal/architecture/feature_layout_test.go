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

func TestTaskFeatureScaffoldExists(t *testing.T) {
	root := projectRoot(t)
	required := []string{
		"internal/feature/task/controller/initiator.go",
		"internal/feature/task/controller/v1.go",
		"internal/feature/task/service/initiator.go",
		"internal/feature/task/repository/initiator.go",
		"internal/feature/task/dto/dto.go",
		"internal/feature/task/entity/entity.go",
		"internal/feature/task/mapper/mapper.go",
	}

	for _, relPath := range required {
		if _, err := os.Stat(filepath.Join(root, relPath)); err != nil {
			t.Fatalf("expected Hinghoi-style feature scaffold file %s to exist: %v", relPath, err)
		}
	}
}

func TestRouterUsesTaskFeatureEntryPoint(t *testing.T) {
	root := projectRoot(t)
	path := filepath.Join(root, "internal/router/router.go")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read router: %v", err)
	}
	src := string(content)

	for _, want := range []string{
		"taskrepository.NewRepository",
		"taskservice.NewService",
		"taskcontroller.NewController",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected router to use %s for /v1/tasks", want)
		}
	}
	if strings.Contains(src, "taskmodule.NewController") {
		t.Fatalf("router still wires /v1/tasks through the old internal/modules entrypoint")
	}
}

type featureFile struct {
	Path  string
	Layer string
}

func featureFiles(t *testing.T) []featureFile {
	t.Helper()
	root := projectRoot(t)
	target := filepath.Join(root, "internal/feature")
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil
	}

	var out []featureFile
	err := filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relPath := rel(root, path)
		parts := strings.Split(relPath, string(filepath.Separator))
		layer := ""
		if len(parts) >= 4 {
			layer = parts[3]
		}
		out = append(out, featureFile{Path: path, Layer: layer})
		return nil
	})
	if err != nil {
		t.Fatalf("walking internal/feature: %v", err)
	}
	return out
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

func TestFeatureServiceDoesNotImportTransportOrPersistence(t *testing.T) {
	root := projectRoot(t)
	for _, ff := range featureFiles(t) {
		if ff.Layer != "service" {
			continue
		}
		for _, imp := range parseImports(t, ff.Path) {
			if strings.Contains(imp, "gin-gonic/gin") ||
				strings.Contains(imp, "gorm.io/gorm") ||
				strings.Contains(imp, "internal/handlers/") ||
				strings.Contains(imp, "internal/models") {
				t.Errorf(
					"BOUNDARY VIOLATION: feature service file %q imports %q.\n"+
						"  Hinghoi-style services must stay transport-agnostic and keep persistence behind repositories.",
					rel(root, ff.Path), imp,
				)
			}
		}
	}
}

func TestFeatureControllerDoesNotImportGORM(t *testing.T) {
	root := projectRoot(t)
	for _, ff := range featureFiles(t) {
		if ff.Layer != "controller" {
			continue
		}
		for _, imp := range parseImports(t, ff.Path) {
			if strings.Contains(imp, "gorm.io/gorm") || strings.Contains(imp, "internal/models") {
				t.Errorf(
					"BOUNDARY VIOLATION: feature controller file %q imports %q.\n"+
						"  Controllers parse/map/delegate only; data access belongs in repository packages.",
					rel(root, ff.Path), imp,
				)
			}
		}
	}
}

func TestFeatureDTOEntityMapperDoNotImportFrameworks(t *testing.T) {
	root := projectRoot(t)
	checkedLayers := map[string]bool{"dto": true, "entity": true, "mapper": true}
	for _, ff := range featureFiles(t) {
		if !checkedLayers[ff.Layer] {
			continue
		}
		for _, imp := range parseImports(t, ff.Path) {
			if strings.Contains(imp, "gin-gonic/gin") ||
				strings.Contains(imp, "gorm.io/gorm") ||
				strings.Contains(imp, "spf13/viper") ||
				strings.Contains(imp, "aws/aws-sdk-go") {
				t.Errorf(
					"BOUNDARY VIOLATION: feature %s file %q imports %q.\n"+
						"  DTO/entity/mapper packages should stay framework-free.",
					ff.Layer, rel(root, ff.Path), imp,
				)
			}
		}
	}
}

func TestHinghoiFeatureStructureIsCanonical(t *testing.T) {
	root := projectRoot(t)
	featureDir := filepath.Join(root, "internal/feature")
	if _, err := os.Stat(featureDir); os.IsNotExist(err) {
		t.Fatalf("internal/feature does not exist; Hinghoi-style feature folders are canonical")
	}

	entries, err := os.ReadDir(featureDir)
	if err != nil {
		t.Fatalf("reading internal/feature: %v", err)
	}

	requiredDirs := []string{"controller", "service", "repository"}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		for _, req := range requiredDirs {
			p := filepath.Join(featureDir, entry.Name(), req)
			if _, err := os.Stat(p); err != nil {
				t.Errorf("feature %q is missing required directory %q: %v", entry.Name(), req, err)
			}
		}
	}
}
