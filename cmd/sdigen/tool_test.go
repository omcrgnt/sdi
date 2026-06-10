package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerator_embedDeps(t *testing.T) {
	tmpDir := setupTestPkg(t)

	sourceCode := `
package testpkg

type repo interface {
	GetData() string
}

type deps struct {
	repo repo
}

type service struct {
	deps
	other string
}
`
	writeFile(t, filepath.Join(tmpDir, "service.go"), sourceCode)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output := readGenFile(t, tmpDir, "service_sdi_gen.go")

	for _, exp := range []string{
		"package testpkg",
		"func (r *service) Deps() []any",
		"(*repo)(nil)",
		"case repo:",
		"r.repo = v",
	} {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q", exp)
		}
	}

	if strings.Contains(output, "other") {
		t.Error("field outside deps should not be in generated code")
	}
}

func TestGenerator_skipsWithoutDepsStruct(t *testing.T) {
	tmpDir := setupTestPkg(t)

	sourceCode := `
package testpkg

type repo interface {
	GetData() string
}

type service struct {
	deps
}
`
	writeFile(t, filepath.Join(tmpDir, "service.go"), sourceCode)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "service_sdi_gen.go")); err == nil {
		t.Fatal("expected no generated file without type deps struct")
	}
}

func TestGenerator_skipsWithoutEmbedDeps(t *testing.T) {
	tmpDir := setupTestPkg(t)

	sourceCode := `
package testpkg

type repo interface {
	GetData() string
}

type deps struct {
	repo repo
}

type service struct {
	repo repo
}
`
	writeFile(t, filepath.Join(tmpDir, "service.go"), sourceCode)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "service_sdi_gen.go")); err == nil {
		t.Fatal("expected no generated file when deps is not embedded")
	}
}

func TestGenerator_skipsNonInterfaceInDeps(t *testing.T) {
	tmpDir := setupTestPkg(t)

	sourceCode := `
package testpkg

type repo interface {
	GetData() string
}

type deps struct {
	repo repo
	name string
}

type worker struct {
	deps
}
`
	writeFile(t, filepath.Join(tmpDir, "service.go"), sourceCode)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output := readGenFile(t, tmpDir, "service_sdi_gen.go")

	if !strings.Contains(output, "(*repo)(nil)") {
		t.Error("expected repo in Deps")
	}
	if strings.Contains(output, "name") {
		t.Error("non-interface field in deps should not be in generated code")
	}
}

func TestGenerator_multipleEmbedders(t *testing.T) {
	tmpDir := setupTestPkg(t)

	sourceCode := `
package testpkg

type repo interface {
	GetData() string
}

type deps struct {
	repo repo
}

type service struct {
	deps
}

type worker struct {
	deps
}
`
	writeFile(t, filepath.Join(tmpDir, "service.go"), sourceCode)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output := readGenFile(t, tmpDir, "service_sdi_gen.go")

	if !strings.Contains(output, "func (r *service) Deps()") {
		t.Error("expected Deps for service")
	}
	if !strings.Contains(output, "func (r *worker) Deps()") {
		t.Error("expected Deps for worker")
	}
}

func setupTestPkg(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "sdi-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	writeFile(t, filepath.Join(tmpDir, "go.mod"), "module testgen\n\ngo 1.21")
	return tmpDir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readGenFile(t *testing.T, dir, name string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("generated file not found: %v", err)
	}
	return string(content)
}
