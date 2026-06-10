package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_printsGeneratedPath(t *testing.T) {
	tmpDir := setupTestPkg(t)
	writeFile(t, filepath.Join(tmpDir, "service.go"), `
package testpkg

type repo interface { Get() }

type deps struct { repo repo }

type service struct { deps }
`)

	out := captureStdout(t, func() {
		if err := Run(tmpDir); err != nil {
			t.Fatalf("Run failed: %v", err)
		}
	})

	want := filepath.Join(tmpDir, "service_sdi_gen.go")
	if !strings.Contains(out, "Generated: "+want) {
		t.Fatalf("stdout = %q, want Generated line for %s", out, want)
	}
}

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

func TestRun_invalidDirectory(t *testing.T) {
	err := Run(filepath.Join(t.TempDir(), "missing-sub"))
	if err == nil {
		t.Fatal("expected error for invalid package directory")
	}
}

func TestGenerator_depsStructAfterOtherTypes(t *testing.T) {
	tmpDir := setupTestPkg(t)

	writeFile(t, filepath.Join(tmpDir, "service.go"), `
package testpkg

type marker int

type repo interface { Get() }

type deps struct {
	repo repo
}

type service struct { deps }
`)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output := readGenFile(t, tmpDir, "service_sdi_gen.go")
	if !strings.Contains(output, "(*repo)(nil)") {
		t.Error("expected repo from deps defined after other types")
	}
}

func TestGenerator_depsInSeparateFile(t *testing.T) {
	tmpDir := setupTestPkg(t)

	writeFile(t, filepath.Join(tmpDir, "deps.go"), `
package testpkg

type repo interface { Get() }

type deps struct {
	repo repo
}
`)
	writeFile(t, filepath.Join(tmpDir, "service.go"), `
package testpkg

type service struct {
	deps
}
`)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output := readGenFile(t, tmpDir, "service_sdi_gen.go")
	if !strings.Contains(output, "(*repo)(nil)") {
		t.Error("expected repo stub from deps.go")
	}
}

func TestGenerator_multipleInterfaceDeps(t *testing.T) {
	tmpDir := setupTestPkg(t)

	sourceCode := `
package testpkg

type repo interface { Get() }
type cache interface { C() }

type deps struct {
	repo repo
	cache cache
}

type service struct {
	deps
}
`
	writeFile(t, filepath.Join(tmpDir, "service.go"), sourceCode)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output := readGenFile(t, tmpDir, "service_sdi_gen.go")
	for _, exp := range []string{"(*repo)(nil)", "(*cache)(nil)", "r.repo = v", "r.cache = v"} {
		if !strings.Contains(output, exp) {
			t.Errorf("expected %q in output", exp)
		}
	}
}

func TestGenerator_genericsTypeParams(t *testing.T) {
	tmpDir := setupTestPkg(t)

	writeFile(t, filepath.Join(tmpDir, "go.mod"), "module testgen\n\ngo 1.21")
	sourceCode := `
package testpkg

type repo interface { Get() }

type deps struct {
	repo repo
}

type service[T any] struct {
	deps
}
`
	writeFile(t, filepath.Join(tmpDir, "service.go"), sourceCode)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output := readGenFile(t, tmpDir, "service_sdi_gen.go")
	if !strings.Contains(output, "func (r *service[T]) Deps()") {
		t.Errorf("expected generic receiver, got:\n%s", output)
	}
}

func TestGenerator_emptyDepsStruct(t *testing.T) {
	tmpDir := setupTestPkg(t)

	sourceCode := `
package testpkg

type deps struct {}

type worker struct {
	deps
}
`
	writeFile(t, filepath.Join(tmpDir, "service.go"), sourceCode)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "service_sdi_gen.go")); err == nil {
		t.Fatal("expected no generated file when deps has no interface fields")
	}
}

func TestGenerator_packageWithoutEmbedder(t *testing.T) {
	tmpDir := setupTestPkg(t)

	sourceCode := `
package testpkg

type repo interface { Get() }

type deps struct {
	repo repo
}

type orphan struct {
	repo repo
}
`
	writeFile(t, filepath.Join(tmpDir, "service.go"), sourceCode)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "service_sdi_gen.go")); err == nil {
		t.Fatal("expected no generated file without deps embed")
	}
}

func TestGenerator_namedEmbedDeps(t *testing.T) {
	tmpDir := setupTestPkg(t)

	sourceCode := `
package testpkg

type repo interface { Get() }

type deps struct {
	repo repo
}

type service struct {
	deps deps
}
`
	writeFile(t, filepath.Join(tmpDir, "service.go"), sourceCode)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output := readGenFile(t, tmpDir, "service_sdi_gen.go")
	if !strings.Contains(output, "r.repo = v") {
		t.Error("expected inject for named embed deps deps")
	}
}

func TestGenerator_writeError(t *testing.T) {
	tmpDir := setupTestPkg(t)

	writeFile(t, filepath.Join(tmpDir, "service.go"), `
package testpkg

type repo interface { Get() }

type deps struct { repo repo }

type service struct { deps }
`)

	if err := os.Mkdir(filepath.Join(tmpDir, "service_sdi_gen.go"), 0755); err != nil {
		t.Fatal(err)
	}

	err := Run(tmpDir)
	if err == nil {
		t.Fatal("expected write error when output path is a directory")
	}
}

func TestGenerator_depsOnlyInTestFile(t *testing.T) {
	tmpDir := setupTestPkg(t)

	writeFile(t, filepath.Join(tmpDir, "service.go"), `
package testpkg

type service struct { deps }
`)
	writeFile(t, filepath.Join(tmpDir, "deps_test.go"), `
package testpkg

type repo interface { Get() }

type deps struct { repo repo }
`)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "service_sdi_gen.go")); err == nil {
		t.Fatal("deps in _test.go must not be used for generation")
	}
}

func TestRun_notPackageDirectory(t *testing.T) {
	tmpDir := setupTestPkg(t)
	path := filepath.Join(tmpDir, "single.go")
	writeFile(t, path, "package testpkg\n")

	if err := Run(path); err == nil {
		t.Fatal("expected error when dir is a file path")
	}
}

func TestGenerator_skipsTestGoSourceFile(t *testing.T) {
	tmpDir := setupTestPkg(t)

	writeFile(t, filepath.Join(tmpDir, "service_test.go"), `
package testpkg

type repo interface { Get() }

type deps struct { repo repo }

type service struct { deps }
`)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "service_sdi_gen.go")); err == nil {
		t.Fatal("must not generate from _test.go only")
	}
}

func TestGenerator_twoOutputFiles(t *testing.T) {
	tmpDir := setupTestPkg(t)

	writeFile(t, filepath.Join(tmpDir, "deps.go"), `
package testpkg

type repo interface { Get() }

type deps struct { repo repo }
`)
	writeFile(t, filepath.Join(tmpDir, "service.go"), `
package testpkg

type service struct { deps }
`)
	writeFile(t, filepath.Join(tmpDir, "worker.go"), `
package testpkg

type worker struct { deps }
`)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	readGenFile(t, tmpDir, "service_sdi_gen.go")
	readGenFile(t, tmpDir, "worker_sdi_gen.go")
}

func TestGenerator_depsNotStruct(t *testing.T) {
	tmpDir := setupTestPkg(t)

	writeFile(t, filepath.Join(tmpDir, "service.go"), `
package testpkg

type deps int

type service struct { deps }
`)

	if err := Run(tmpDir); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "service_sdi_gen.go")); err == nil {
		t.Fatal("expected no file when deps is not a struct")
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

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	os.Stdout = old
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	return <-done
}

func readGenFile(t *testing.T, dir, name string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("generated file not found: %v", err)
	}
	return string(content)
}
