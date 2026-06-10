package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_writeGenerated_formatsValidSource(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.go")

	err := writeGenerated(path, []byte("package p\n\nfunc F(){}\n"))
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0644 {
		t.Fatalf("file mode = %o, want 0644", info.Mode().Perm())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "func F()") {
		t.Fatalf("unexpected content: %s", data)
	}
}

func Test_writeGenerated_writesUnformattedOnFormatError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.go")

	raw := []byte("package p\n\nfunc F(){\n")
	if err := writeGenerated(path, raw); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(raw) {
		t.Fatalf("expected raw bytes on format error, got %q", data)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0644 {
		t.Fatalf("file mode = %o, want 0644", info.Mode().Perm())
	}
}

func Test_writeGenerated_logsFormatError(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	dir := t.TempDir()
	path := filepath.Join(dir, "out.go")
	raw := []byte("package p\n\nfunc F(){\n")

	if err := writeGenerated(path, raw); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "failed to format") {
		t.Fatalf("expected format error log, got %q", buf.String())
	}
}
