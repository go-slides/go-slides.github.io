package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-icons/twemoji"

	"github.com/go-slides/go-slides.github.io/playground/emojipack"
)

// EmojiPackFile is the Twemoji graphics as one file (package emojipack), and
// EmojiLicenseFile and EmojiNoticeFile what travels with it (CC-BY 4.0 asks for
// its text and a credit). They are written under out/emoji/. The module is built
// with go-slides' tag slides_noemoji (ModuleTags), so it links none of the
// graphics (2.06 MB gzipped less); the render worker fetches the pack the first
// time a render has an emoji (cmd/playground-wasm/lazyfetch.go, which asks for
// "emoji/" + EmojiPackFile).
const (
	EmojiPackFile    = "twemoji.pack"
	EmojiLicenseFile = "twemoji.LICENSE-GRAPHICS.txt"
	EmojiNoticeFile  = "NOTICE.txt"
)

// emojiLicense is the CC-BY 4.0 legal text of the graphics, verbatim from
// go-icons/twemoji's TWEMOJI-LICENSE-GRAPHICS (the module the SVGs come from).
//
//go:embed TWEMOJI-LICENSE-GRAPHICS.txt
var emojiLicense []byte

// writeEmojiPack writes the pack under dir: every graphic of go-icons/twemoji,
// the very SVGs go-slides' embedded provider (emoji/embedded) serves, by stem.
// It returns the pack's size and how many graphics it holds.
func writeEmojiPack(dir string) (size int64, n int, err error) {
	stems := twemoji.Stems()
	entries := make([]emojipack.Entry, 0, len(stems))
	for _, s := range stems {
		svg, ok := twemoji.SVGByStem(s)
		if !ok {
			return 0, 0, fmt.Errorf("go-icons/twemoji has no graphic for its own stem %s", s)
		}
		entries = append(entries, emojipack.Entry{Stem: s, SVG: svg})
	}
	b, err := emojipack.Encode(entries)
	if err != nil {
		return 0, 0, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, 0, err
	}
	notice := fmt.Sprintf("%s\n\nTwemoji graphics, version %s, as one file (%s: see playground/emojipack).\nThe licence is in %s (Creative Commons Attribution 4.0 International,\nhttps://creativecommons.org/licenses/by/4.0/). Whoever shows these graphics must credit them as above.\n",
		twemoji.Attribution, twemoji.Version, EmojiPackFile, EmojiLicenseFile)
	for name, data := range map[string][]byte{EmojiPackFile: b, EmojiLicenseFile: emojiLicense, EmojiNoticeFile: []byte(notice)} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			return 0, 0, err
		}
	}
	return int64(len(b)), len(entries), nil
}
