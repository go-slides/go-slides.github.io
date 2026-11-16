package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/icons"
)

// Every icon the playground's sources name resolves in the generated subset, the
// subset holds nothing else, and the file is what the generator writes now.
func TestEveryUsedIconIsInTheSubset(t *testing.T) {
	names, err := Scan("../..")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 40 {
		t.Fatalf("scanned %d icon names, expected the ribbon's fifty or so", len(names))
	}
	for _, n := range names {
		if !strings.Contains(icons.Icon(n), "<svg") {
			t.Errorf("icon %q is used by the sources but missing from icons/icons_gen.go: run go generate ./icons", n)
		}
	}
	if icons.Len() != len(names) {
		t.Errorf("the subset has %d icons, the sources use %d: run go generate ./icons", icons.Len(), len(names))
	}
	if err := run(t, "--dir", "../..", "--out", "../../icons/icons_gen.go", "--check"); err != nil {
		t.Error(err)
	}
}

// The licence notice travels with the glyphs.
func TestLicenceNoticeIsKept(t *testing.T) {
	src, err := os.ReadFile("../../icons/icons_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"MIT License", "Luca Burgio", "ICONOIR-LICENSE"} {
		if !bytes.Contains(src, []byte(want)) {
			t.Errorf("icons_gen.go lacks %q", want)
		}
	}
	lic, err := os.ReadFile("../../icons/ICONOIR-LICENSE")
	if err != nil || !bytes.Contains(lic, []byte("Permission is hereby granted")) {
		t.Errorf("icons/ICONOIR-LICENSE is missing or not the MIT text: %v", err)
	}
}

// A name the sources use but iconoir lacks is an error, not an empty glyph; a
// name that is not a literal is one too.
func TestGeneratorRefusesWhatItCannotResolve(t *testing.T) {
	if _, err := Generate([]string{"no-such-icon"}); err == nil {
		t.Error("Generate accepted an unknown icon")
	}
	dir := t.TempDir()
	src := "package p\nfunc f(n string) { icon(n) }\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Scan(dir); err == nil {
		t.Error("Scan accepted an icon name that is not a literal")
	}
}

func TestWritesTheFile(t *testing.T) {
	dir := t.TempDir()
	src := "package p\nfunc f() { icon(\"undo\"); act(\"x\", 0, \"redo\", nil) }\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "gen.go")
	if err := run(t, "--dir", dir, "--out", out); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(out)
	if !bytes.Contains(got, []byte(`"undo"`)) || !bytes.Contains(got, []byte(`"redo"`)) {
		t.Errorf("generated file lacks the icons:\n%s", got)
	}
	if err := run(t, "--dir", dir, "--out", out, "--check"); err != nil {
		t.Error(err)
	}
	if err := os.WriteFile(out, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(t, "--dir", dir, "--out", out, "--check"); err == nil {
		t.Error("--check passed on a stale file")
	}
}

func run(t *testing.T, args ...string) error {
	t.Helper()
	cmd := RootCmd()
	cmd.SetArgs(args)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	return cmd.Execute()
}
