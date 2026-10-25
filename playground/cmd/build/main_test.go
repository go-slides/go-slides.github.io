package main

import (
	"bytes"
	"encoding/json"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-slides/slides/export"
	"github.com/go-webengine/engine/paint"
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
	if !strings.Contains(report, "cards: 7 pictures") {
		t.Fatal(report)
	}
	wasm, err := os.ReadFile(filepath.Join(out, "export.wasm"))
	if err != nil {
		t.Fatal(err)
	}
	// the fonts the export engine fetches: every one it may ask for, the
	// bytes this build embeds
	if !strings.Contains(report, "fonts: 23 files") {
		t.Fatal(report)
	}
	for name, want := range map[string][]byte{"inter/inter.ttf": paint.BundledFontData("inter/inter.ttf"), "lato/lato-black.ttf": export.FontData("lato/lato-black.ttf")} {
		got, err := os.ReadFile(filepath.Join(out, "fonts", filepath.FromSlash(name)))
		if err != nil || len(want) == 0 || !bytes.Equal(got, want) {
			t.Fatalf("fonts/%s: %d bytes for %d, %v", name, len(got), len(want), err)
		}
		// and the module does not carry them: a stretch of the file's glyph data
		if bytes.Contains(wasm, want[len(want)/2:len(want)/2+256]) {
			t.Errorf("export.wasm embeds %s", name)
		}
	}
	// built with ExportTags: no JavaScript engine. A package linked in
	// leaves its import path in the module's type data (-s -w strips the
	// symbol names, not that): an export built without the tag holds
	// goja's 15 times and esbuild's 10.
	for _, pkg := range []string{"github.com/dop251/goja", "github.com/evanw/esbuild"} {
		if bytes.Contains(wasm, []byte(pkg)) {
			t.Errorf("export.wasm links %s", pkg)
		}
	}
	// and no renderer: the deck comes rendered from the page's module, which
	// has one (the planted positive)
	page, err := os.ReadFile(filepath.Join(out, "playground.wasm"))
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range []string{"github.com/yuin/goldmark", "github.com/go-rouge/rouge", "github.com/hashicorp/hcl"} {
		if !bytes.Contains(page, []byte(pkg)) {
			t.Errorf("playground.wasm names no %s: the check below sees nothing", pkg)
		}
		if bytes.Contains(wasm, []byte(pkg)) {
			t.Errorf("export.wasm links %s", pkg)
		}
	}
	if !bytes.Contains(wasm, []byte("github.com/go-webengine/engine")) {
		t.Error("the planted positive: export.wasm names no go-webengine symbol, so the check above sees nothing")
	}
	if _, err := execute(t, "--module", filepath.Join("..", ".."), "--out", out, "--max-gzip", "1"); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatalf("a 1-byte budget passed: %v", err)
	}
}

// TestCards renders the carousel's deck: one JPEG a card, small, at the
// card size; and the card list the site shows has as many entries.
func TestCards(t *testing.T) {
	out := t.TempDir()
	n, size, err := renderCards(t.Context(), filepath.Join("..", "..", "cards", "cards.md"), out)
	if err != nil {
		t.Fatal(err)
	}
	var cards []struct{ Title, Text, Href string }
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "data", "playground_cards.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &cards); err != nil {
		t.Fatal(err)
	}
	if n != len(cards) || n != 7 {
		t.Fatalf("%d pictures for %d cards", n, len(cards))
	}
	if size > 7*40<<10 {
		t.Fatalf("the pictures weigh %d KiB", size>>10)
	}
	f, err := os.Open(filepath.Join(out, "7.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cfg, err := jpeg.DecodeConfig(f)
	if err != nil || cfg.Width != 480 || cfg.Height != 270 {
		t.Fatalf("%+v %v", cfg, err)
	}
	for _, c := range cards {
		if c.Title == "" || c.Text == "" || !strings.HasPrefix(c.Href, "https://go-slides.github.io/docs/") {
			t.Errorf("card %+v", c)
		}
	}
	if _, _, err := renderCards(t.Context(), filepath.Join(out, "none.md"), out); err == nil {
		t.Fatal("a missing deck rendered")
	}
	// the cards alone, as the site's link check builds them
	site := t.TempDir()
	report, err := execute(t, "--module", filepath.Join("..", ".."), "--out", site, "--wasm=false")
	if err != nil || !strings.Contains(report, "cards: 7 pictures") || strings.Contains(report, "playground.wasm") {
		t.Fatal(report, err)
	}
	if _, err := os.Stat(filepath.Join(site, "cards", "1.jpg")); err != nil {
		t.Fatal(err)
	}
	if _, err := execute(t, "--module", filepath.Join("..", ".."), "--out", site, "--wasm=false", "--cards", "none.md"); err == nil {
		t.Fatal("a missing cards deck passed")
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
