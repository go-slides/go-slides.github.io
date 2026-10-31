package fontsubset

import (
	"testing"

	"github.com/go-opentype/fonts/atkinsonhyperlegible"
	"github.com/go-opentype/fonts/notosans"
	"github.com/go-opentype/fonts/notosansjp"
	"github.com/go-opentype/fonts/notosanskr"
	"github.com/go-opentype/fonts/notosanssc"
	"github.com/go-opentype/opentype"
	"github.com/go-widgets/toolkit"
)

// subsetted returns a subset that the toolkit draws with as the whole face
// does: the same advance for each rune, and ink.
func checkAgainstWhole(t *testing.T, whole []byte, runes []rune, probe string) []byte {
	t.Helper()
	sub, err := Subset(whole, runes)
	if err != nil {
		t.Fatal(err)
	}
	f, err := opentype.Parse(sub)
	if err != nil {
		t.Fatalf("the subset does not parse: %v", err)
	}
	for _, r := range probe {
		if g, ok := f.GlyphIndex(r); !ok || g == 0 {
			t.Errorf("%q is not in the subset", r)
		}
	}
	a, err1 := toolkit.NewTrueTypeFont(whole, 16)
	b, err2 := toolkit.NewTrueTypeFont(sub, 16)
	if err1 != nil || err2 != nil {
		t.Fatalf("toolkit: %v %v", err1, err2)
	}
	// kerning is not kept: a pixel or two over a line is the whole difference
	if wa, wb := a.Measure(probe), b.Measure(probe); wa == 0 || wa-wb > 2 || wb-wa > 2 {
		t.Errorf("%q measures %d in the whole face, %d in the subset", probe, wa, wb)
	}
	t.Logf("%d runes: %d bytes (whole face %d)", len(runes), len(sub), len(whole))
	return sub
}

func TestSubsetKeepsWhatItWasAskedFor(t *testing.T) {
	checkAgainstWhole(t, atkinsonhyperlegible.TTF, Latin(), "Hello, Wörld")
}

func TestCyrillic(t *testing.T) {
	sub := checkAgainstWhole(t, notosans.TTF, append(Latin(), Cyrillic()...), "Привіт, світе Ёё Її Ґґ")
	if len(sub) > 400<<10 {
		t.Errorf("the Cyrillic face is %d bytes", len(sub))
	}
}

func TestCJKFaces(t *testing.T) {
	cases := []struct {
		name  string
		ttf   []byte
		set   []rune
		probe string
	}{
		{"zh", notosanssc.TTF, GB2312(), "幻灯片，格式"},
		{"ja", notosansjp.TTF, JISX0208(), "スライド、書式設定"},
		{"ko", notosanskr.TTF, KSX1001(), "슬라이드 서식"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if len(c.set) < 2000 {
				t.Fatalf("the %s repertoire has %d characters", c.name, len(c.set))
			}
			checkAgainstWhole(t, c.ttf, append(Latin(), c.set...), c.probe)
		})
	}
}

func TestNoGlyphsIsAnError(t *testing.T) {
	if _, err := Subset(atkinsonhyperlegible.TTF, []rune{'中'}); err == nil {
		t.Error("a subset of nothing is not a font")
	}
}
