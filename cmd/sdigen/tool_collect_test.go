package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

func loadTestPackage(t *testing.T, dir string) *packages.Package {
	t.Helper()

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		t.Fatalf("packages.Load: %v", err)
	}
	if len(pkgs) == 0 {
		t.Fatal("no packages loaded")
	}
	return pkgs[0]
}

func Test_findDepsStructFields(t *testing.T) {
	tmpDir := setupTestPkg(t)
	writeFile(t, filepath.Join(tmpDir, "deps.go"), `
package testpkg

type repo interface { Get() }
type cache interface { C() }

type deps struct {
	repo repo
	cache cache
}
`)

	pkg := loadTestPackage(t, tmpDir)
	fields := findDepsStructFields(pkg)
	if len(fields) != 2 {
		t.Fatalf("expected 2 deps fields, got %d (%v)", len(fields), fields)
	}
}

func Test_findDepsStructFields_skipsNonStructDeps(t *testing.T) {
	tmpDir := setupTestPkg(t)
	writeFile(t, filepath.Join(tmpDir, "deps.go"), `
package testpkg

type deps int
`)

	pkg := loadTestPackage(t, tmpDir)
	if fields := findDepsStructFields(pkg); len(fields) != 0 {
		t.Fatalf("expected no fields for non-struct deps, got %v", fields)
	}
}

func Test_typeParamsString_generics(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fixture.go", "package fixture\ntype service[T, V any] struct{}", 0)
	if err != nil {
		t.Fatal(err)
	}

	ts := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	got := typeParamsString(ts)
	if got != "[T, V]" {
		t.Fatalf("typeParamsString = %q", got)
	}
}

func Test_isInterface(t *testing.T) {
	tmpDir := setupTestPkg(t)
	writeFile(t, filepath.Join(tmpDir, "deps.go"), `
package testpkg

type repo interface { Get() }

type deps struct {
	repo repo
	name string
}
`)

	pkg := loadTestPackage(t, tmpDir)

	var repoField, nameField ast.Expr
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name.Name != "deps" {
					continue
				}
				st := ts.Type.(*ast.StructType)
				repoField = st.Fields.List[0].Type
				nameField = st.Fields.List[1].Type
			}
		}
	}

	if repoField == nil || nameField == nil {
		t.Fatal("deps struct fields not found")
	}
	if !isInterface(pkg, repoField) {
		t.Fatal("repo must be interface")
	}
	if isInterface(pkg, nameField) {
		t.Fatal("string field must not be interface")
	}
	if isInterface(pkg, &ast.Ident{Name: "not_in_package"}) {
		t.Fatal("unknown type must not be interface")
	}
}

func Test_typeParamsString_nonGeneric(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fixture.go", "package fixture\ntype service struct{}", 0)
	if err != nil {
		t.Fatal(err)
	}

	ts := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	if got := typeParamsString(ts); got != "" {
		t.Fatalf("expected empty type params, got %q", got)
	}
}
