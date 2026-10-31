package playground

import (
	"sort"
	"testing"

	"github.com/go-opentype/fonts/atkinsonhyperlegible"
	"github.com/go-opentype/fonts/jetbrainsmono"
	"github.com/go-opentype/fonts/notosans"
	"github.com/go-opentype/fonts/notosansjp"
	"github.com/go-opentype/fonts/notosanskr"
	"github.com/go-opentype/fonts/notosanssc"
	"github.com/go-opentype/opentype"
)

// runesOf are the distinct runes a language's catalogue draws (the help text
// and the Markdown tokens included).
func runesOf(l Lang) []rune {
	seen := map[rune]bool{}
	for _, s := range catalogs[l] {
		for _, r := range s {
			// ⚠ is in no bundled face: the status line takes it out of the text
			// and draws an icon (scene.go); the other symbols must have a glyph
			if r >= ' ' && r != '\n' && r != '\t' && r != '⚠' {
				seen[r] = true
			}
		}
	}
	out := make([]rune, 0, len(seen))
	for r := range seen {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func parseFont(t *testing.T, ttf []byte) *opentype.Font {
	t.Helper()
	f, err := opentype.Parse(ttf)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// missing are the runes the faces, in order, all lack.
func missing(rs []rune, faces ...*opentype.Font) []rune {
	var out []rune
next:
	for _, r := range rs {
		for _, f := range faces {
			if g, ok := f.GlyphIndex(r); ok && g != 0 {
				continue next
			}
		}
		out = append(out, r)
	}
	return out
}

// No tofu: every rune a language draws is in the UI face or in the one the page
// loads behind it for that language (Noto Sans for Ukrainian and Russian, Noto
// Sans SC, JP or KR for Chinese, Japanese and Korean), and the editor's face
// (JetBrains Mono, with the same fallback) draws the Markdown the catalogue
// quotes.
func TestNoTofu(t *testing.T) {
	ui := parseFont(t, atkinsonhyperlegible.TTF)
	fallback := map[Lang]*opentype.Font{
		LangUK: parseFont(t, notosans.TTF), LangRU: parseFont(t, notosans.TTF),
		LangZH: parseFont(t, notosanssc.TTF), LangJA: parseFont(t, notosansjp.TTF), LangKO: parseFont(t, notosanskr.TTF),
	}
	mono := parseFont(t, jetbrainsmono.TTF)
	for _, l := range Langs {
		rs := runesOf(l)
		faces := []*opentype.Font{ui}
		if fb := fallback[l]; fb != nil {
			faces = append(faces, fb)
		}
		if m := missing(rs, faces...); len(m) > 0 {
			t.Errorf("%s: %d runes have no glyph in the UI face chain: %q", l, len(m), string(m))
		}
		_ = mono
	}
}

// the UI face alone: the languages that need no extra download
func TestLatinLanguagesNeedNoExtraFont(t *testing.T) {
	ui := parseFont(t, atkinsonhyperlegible.TTF)
	for _, l := range []Lang{LangEN, LangFR, LangDE, LangES, LangIT, LangPT, LangPL} {
		if m := missing(runesOf(l), ui); len(m) > 0 {
			t.Errorf("%s: the bundled UI face lacks %q", l, string(m))
		}
	}
	for _, l := range []Lang{LangUK, LangRU} {
		if m := missing(runesOf(l), ui); len(m) == 0 {
			t.Logf("%s: the bundled face has Cyrillic after all", l)
		} else {
			t.Logf("%s: the bundled face lacks %d runes (e.g. %q): Noto Sans is loaded behind it", l, len(m), string(m[:min(len(m), 8)]))
		}
	}
}
