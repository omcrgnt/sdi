package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func parseStructType(t *testing.T, decl string) *ast.StructType {
	t.Helper()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fixture.go", "package fixture\n"+decl, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	ts := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		t.Fatalf("expected struct type")
	}
	return st
}

func Test_skipFile(t *testing.T) {
	if !skipFile("service_test.go") {
		t.Error("expected skip *_test.go")
	}
	if !skipFile("service_sdi_gen.go") {
		t.Error("expected skip *_sdi_gen.go")
	}
	if skipFile("service.go") {
		t.Error("expected not skip service.go")
	}
}

func Test_structEmbedsDeps(t *testing.T) {
	if !structEmbedsDeps(parseStructType(t, "type S struct { deps }")) {
		t.Fatal("anonymous embed deps")
	}
	if !structEmbedsDeps(parseStructType(t, "type S struct { deps deps }")) {
		t.Fatal("named embed deps deps")
	}
	if structEmbedsDeps(parseStructType(t, "type S struct { repo repo }")) {
		t.Fatal("no deps embed")
	}
}

func Test_typeName(t *testing.T) {
	st := parseStructType(t, "type S struct { deps }")
	field := st.Fields.List[0]
	if got := typeName(field.Type); got != "deps" {
		t.Fatalf("typeName = %q", got)
	}
}

func Test_typeName_nonIdent(t *testing.T) {
	st := parseStructType(t, "type S struct { p *deps }")
	field := st.Fields.List[0]
	if got := typeName(field.Type); got != "" {
		t.Fatalf("typeName(*deps) = %q, want empty", got)
	}
}

func Test_structEmbedsDeps_notFirstField(t *testing.T) {
	if !structEmbedsDeps(parseStructType(t, "type S struct { n int; deps }")) {
		t.Fatal("deps embed after other field")
	}
}
