package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAssetsGenIsCurrent: the size and SHA-256 compiled into the module are
// those of the files the build writes now. A copy that went stale (go-tex/math or
// go-icons/twemoji updated, the pack's format changed, a toolchain that deflates
// otherwise) makes this red; `go run ./cmd/build` rewrites it.
func TestAssetsGenIsCurrent(t *testing.T) {
	dir := t.TempDir()
	font, err := writeMathFont(dir)
	if err != nil {
		t.Fatal(err)
	}
	pack, _, err := writeEmojiPack(dir)
	if err != nil {
		t.Fatal(err)
	}
	want, err := assetsGen(font, pack)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join("..", "..", AssetsGenFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("%s is stale: run `go run ./cmd/build` in playground/ and commit it.\nwant:\n%s\ngot:\n%s", AssetsGenFile, want, got)
	}
	// the planted negative: a one-byte change in a file is another hash
	other := digestOf(append([]byte("x"), 0))
	if b, _ := assetsGen(other, pack); string(b) == string(want) {
		t.Error("the generated source does not depend on the font's digest")
	}
	// and the file written is the file hashed
	for name, d := range map[string]digest{MathFontFile: font, EmojiPackFile: pack} {
		p := dir
		if name == EmojiPackFile {
			p = dir // writeEmojiPack writes beside the font here: one directory
		}
		b, err := os.ReadFile(filepath.Join(p, name))
		if err != nil || digestOf(b) != d {
			t.Errorf("%s: %v", name, err)
		}
	}
	if !strings.Contains(string(want), font.SHA) || !strings.Contains(string(want), pack.SHA) {
		t.Error("the digests are not in the source")
	}
}
