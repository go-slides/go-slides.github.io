package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-icons/twemoji"
	"github.com/go-tex/math"

	"github.com/go-slides/go-slides.github.io/playground/emojipack"
)

// TestMathFontBytesAndLicence: the file the page fetches is the embedded
// provider's font, byte for byte, and travels with its licence.
func TestMathFontBytesAndLicence(t *testing.T) {
	dir := t.TempDir()
	n, err := writeMathFont(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, MathFontFile))
	if err != nil || n != int64(len(got)) || !bytes.Equal(got, math.DefaultFont()) {
		t.Fatalf("%d bytes written, %d reported: %v", len(got), n, err)
	}
	// an OpenType font with CFF outlines, and under what the provider accepts (8 MiB)
	if !bytes.HasPrefix(got, []byte("OTTO")) || len(got) > 8<<20 {
		t.Errorf("not a font the provider takes: %q, %d bytes", got[:4], len(got))
	}
	lic, err := os.ReadFile(filepath.Join(dir, MathFontLicense))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Copyright 2001-2021 The STIX Fonts Project Authors", "SIL OPEN FONT LICENSE Version 1.1"} {
		if !strings.Contains(string(lic), want) {
			t.Errorf("the licence lacks %q", want)
		}
	}
	f := filepath.Join(t.TempDir(), "file")
	os.WriteFile(f, nil, 0o644)
	if _, err := writeMathFont(filepath.Join(f, "x")); err == nil {
		t.Error("a directory under a file was written")
	}
}

// TestModuleTags: the page's module is built without the video reader and
// without the formulas' font, and fetches the font from where this build puts it.
func TestModuleTags(t *testing.T) {
	if got := strings.Fields(ModuleTags); len(got) != 3 || got[0] != "slides_novideo" || got[1] != "slides_nomathfont" || got[2] != "slides_noemoji" {
		t.Errorf("ModuleTags = %q", ModuleTags)
	}
	src, err := os.ReadFile(filepath.Join("..", "playground-wasm", "lazyfetch.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), `mathFontFile = "fonts/`+MathFontFile+`"`) {
		t.Errorf("cmd/playground-wasm/lazyfetch.go asks for another file than %s", MathFontFile)
	}
	if !strings.Contains(string(src), `emojiPackFile = "emoji/`+EmojiPackFile+`"`) {
		t.Errorf("cmd/playground-wasm/lazyfetch.go asks for another file than %s", EmojiPackFile)
	}
}

// TestEmojiPack: the pack the page fetches holds every graphic the embedded
// provider serves, byte for byte, and travels with its licence and credit.
func TestEmojiPack(t *testing.T) {
	dir := t.TempDir()
	size, n, err := writeEmojiPack(dir)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, EmojiPackFile))
	if err != nil || int64(len(b)) != size {
		t.Fatalf("%d bytes for %d: %v", len(b), size, err)
	}
	p, err := emojipack.Decode(b)
	if err != nil || p.Len() != n || n != len(twemoji.Stems()) {
		t.Fatalf("%v: %d graphics for %d", err, n, len(twemoji.Stems()))
	}
	for _, s := range twemoji.Stems() {
		want, _ := twemoji.SVGByStem(s)
		if got, ok := p.SVGByStem(s); !ok || got != want {
			t.Fatalf("the pack's %s differs from the embedded set's", s)
		}
	}
	lic, err := os.ReadFile(filepath.Join(dir, EmojiLicenseFile))
	if err != nil || !strings.Contains(string(lic), "Attribution 4.0 International") {
		t.Errorf("licence: %v", err)
	}
	notice, err := os.ReadFile(filepath.Join(dir, EmojiNoticeFile))
	if err != nil || !strings.Contains(string(notice), twemoji.Attribution) {
		t.Errorf("notice: %v", err)
	}
	if size > 2<<20 {
		t.Errorf("the pack is %d bytes: over the embedded archive's 2.16 MB is no gain", size)
	}
}
