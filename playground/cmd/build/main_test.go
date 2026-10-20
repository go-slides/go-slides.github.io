package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func execute(t *testing.T, args ...string) (string, error) {
	t.Helper()
	c := RootCmd()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetErr(&out)
	c.SetArgs(args)
	err := c.Execute()
	return out.String(), err
}

// TestBuild compiles the real module into a temporary site directory.
func TestBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles the WebAssembly module")
	}
	out := t.TempDir()
	report, err := execute(t, "--module", filepath.Join("..", ".."), "--out", out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(report, "gzip") {
		t.Fatal(report)
	}
	if b, err := os.ReadFile(filepath.Join(out, "wasm_exec.js")); err != nil || !bytes.Contains(b, []byte("globalThis.Go")) {
		t.Fatal("wasm_exec.js", err)
	}
	if _, err := execute(t, "--module", filepath.Join("..", ".."), "--out", out, "--max-gzip", "1"); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatalf("a 1-byte budget passed: %v", err)
	}
}

func TestErrors(t *testing.T) {
	if _, err := execute(t, "--go", filepath.Join(t.TempDir(), "no-go"), "--out", t.TempDir()); err == nil {
		t.Fatal("a missing go command passed")
	}
	f := filepath.Join(t.TempDir(), "file")
	os.WriteFile(f, nil, 0o644)
	if _, err := execute(t, "--out", filepath.Join(f, "x")); err == nil {
		t.Fatal("an output under a file passed")
	}
	if _, err := execute(t, "--module", t.TempDir(), "--out", t.TempDir()); err == nil {
		t.Fatal("a directory with no module compiled")
	}
	if _, err := execute(t, "extra"); err == nil {
		t.Fatal("an argument was accepted")
	}
	old := osExit
	defer func() { osExit = old }()
	code := 0
	osExit = func(c int) { code = c }
	os.Args = []string{"build", "--bogus"}
	main()
	if code != 1 {
		t.Fatal(code)
	}
	if s := Measure(bytes.Repeat([]byte("ab"), 1000)); s.Raw != 2000 || s.Gzip >= s.Raw || s.Brotli >= s.Raw {
		t.Fatalf("%+v", s)
	}
}
