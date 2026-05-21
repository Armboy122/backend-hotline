package migrations_test

import (
	"os"
	"strings"
	"testing"
)

func TestMigrationWorkflowDocumentsGooseAndMakeTargets(t *testing.T) {
	script, err := os.ReadFile("../../../scripts/migrate.sh")
	if err != nil {
		t.Fatalf("read scripts/migrate.sh: %v", err)
	}
	scriptText := string(script)
	for _, fragment := range []string{
		"GOOSE_DRIVER=postgres",
		"pkg/db/migrations",
		"goose -dir",
		"DATABASE_URL",
	} {
		if !strings.Contains(scriptText, fragment) {
			t.Fatalf("migrate.sh missing fragment %q\nScript:\n%s", fragment, scriptText)
		}
	}

	makefile, err := os.ReadFile("../../../Makefile")
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	makeText := string(makefile)
	for _, fragment := range []string{"migrate:", "migrate-status:", "migrate-prod:", "predeploy-prod:"} {
		if !strings.Contains(makeText, fragment) {
			t.Fatalf("Makefile missing target %q", fragment)
		}
	}

	readme, err := os.ReadFile("../../../README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readmeText := string(readme)
	for _, fragment := range []string{"Goose", "scripts/migrate.sh", "Cloud Run", "ก่อน deploy"} {
		if !strings.Contains(readmeText, fragment) {
			t.Fatalf("README missing migration documentation fragment %q", fragment)
		}
	}
}
