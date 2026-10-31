package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-opentype/fonts/notosans"
	"github.com/go-opentype/fonts/notosansjp"
	"github.com/go-opentype/fonts/notosanskr"
	"github.com/go-opentype/fonts/notosanssc"
	"github.com/go-slides/go-slides.github.io/playground"
	"github.com/go-slides/go-slides.github.io/playground/fontsubset"
)

// uiFont is a face the page downloads behind the UI face for some languages,
// cut down to the repertoire of the language (fontsubset): the bundled
// Atkinson Hyperlegible has no Cyrillic, no kana, hanzi or hangul.
type uiFont struct {
	file  string
	langs []playground.Lang
	ttf   []byte
	set   func() []rune
}

var uiFonts = []uiFont{
	{"ui-cyrillic.ttf", []playground.Lang{playground.LangUK, playground.LangRU}, notosans.TTF, fontsubset.Cyrillic},
	{"ui-zh.ttf", []playground.Lang{playground.LangZH}, notosanssc.TTF, fontsubset.GB2312},
	{"ui-ja.ttf", []playground.Lang{playground.LangJA}, notosansjp.TTF, fontsubset.JISX0208},
	{"ui-ko.ttf", []playground.Lang{playground.LangKO}, notosanskr.TTF, fontsubset.KSX1001},
}

// writeUIFonts writes the subsets under dir; the page fetches the one its
// language needs (ui-<file>), only for the languages that need one.
func writeUIFonts(dir string) (n int, size int64, err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, 0, err
	}
	for _, f := range uiFonts {
		runes := append(fontsubset.Latin(), f.set()...)
		for _, l := range f.langs {
			runes = append(runes, playground.CatalogueRunes(l)...)
		}
		b, err := fontsubset.Subset(f.ttf, runes)
		if err != nil {
			return 0, 0, fmt.Errorf("%s: %w", f.file, err)
		}
		if err := os.WriteFile(filepath.Join(dir, f.file), b, 0o644); err != nil {
			return 0, 0, err
		}
		n++
		size += int64(len(b))
	}
	return n, size, nil
}
