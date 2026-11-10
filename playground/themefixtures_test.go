package playground

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteThemeFixturesForBrowser writes the files browsertest/themeeditor.cjs
// imports (THEME_FIXTURES_OUT names the folder): a theme package like the CLI's
// example, a package the library refuses, and a loose sheet.
func TestWriteThemeFixturesForBrowser(t *testing.T) {
	dir := os.Getenv("THEME_FIXTURES_OUT")
	if dir == "" {
		t.Skip("THEME_FIXTURES_OUT is not set")
	}
	for name, data := range map[string][]byte{
		"acme.gsta.xz":  pack(t, acmeFiles()),
		"evil.gsta.xz":  pack(t, with(acmeFiles(), "theme.css", "h1{background:url(https://evil.example/x.png)}")),
		"loose.css":     []byte("/* @theme loose */\n@import \"default\";\nh1 { color: #b00020 }\n"),
		"not-a-sheet.t": []byte("hello"),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
