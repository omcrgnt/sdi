package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sdi-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	goMod := "module testgen\n\ngo 1.21"
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)

	sourceCode := `
package testpkg

type repo interface {
	GetData() string
}

type service struct {
	db repo
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "service.go"), []byte(sourceCode), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = Run(tmpDir)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	genFile := filepath.Join(tmpDir, "service_sdi_gen.go")
	content, err := os.ReadFile(genFile)
	if err != nil {
		t.Fatalf("Generated file not found: %v", err)
	}

	output := string(content)

	expected := []string{
		"package testpkg",
		"func (r *service) Deps() []any",
		"(*repo)(nil)",
		"func (r *service) Inject(args []any)",
		"case repo:",
		"r.db = v",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain %q, but it didn't", exp)
		}
	}
}
