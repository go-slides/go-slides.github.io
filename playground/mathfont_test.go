package playground

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/go-tex/math"

	"github.com/go-slides/slides/mathfont"
	"github.com/go-slides/slides/mathfont/embedded"

	"github.com/go-slides/go-slides.github.io/playground/emojipack"
)

// lateFont is a mathfont.Provider whose font arrives later, as the render
// worker's does (cmd/playground-wasm/lazyfetch.go).
type lateFont struct{ loaded atomic.Bool }

func (l *lateFont) Font() ([]byte, error) {
	if !l.loaded.Load() {
		return nil, errors.New("not fetched yet")
	}
	return math.DefaultFont(), nil
}

// TestWelcomeNeedsNoLazyFile: the deck a first visit opens has no formula and no
// emoji, so a first visit fetches neither the math font nor the emoji pack
// (and the offline cache's kit, which says no first visit may fetch a lazy file,
// holds: .github/workflows/playground.yml).
func TestWelcomeNeedsNoLazyFile(t *testing.T) {
	src := Examples[0].Source
	if strings.Contains(src, "$") {
		t.Error("the welcome deck has a formula")
	}
	if emojipack.HasEmoji(src) {
		t.Error("the welcome deck has an emoji")
	}
	// the planted positives: the other examples do, so the checks see something
	if !emojipack.HasEmoji(":sparkles: x") || !strings.Contains(Examples[9].Source, "$") {
		t.Error("the checks above see nothing")
	}
}

// TestFormulasAfterTheFontArrives: the preview renders a deck by slides, keeping
// each slide's HTML for the next render (the editor package's render cache),
// and that key does not say whether the font had arrived. After the font does,
// ForgetRenderedSlides makes the next render write the formulas again; without
// it a slide that did not change would keep its source.
func TestFormulasAfterTheFontArrives(t *testing.T) {
	t.Cleanup(embedded.Register)
	late := &lateFont{}
	mathfont.Register(late)
	req := Request{ID: 1, Source: "---\nslides: true\n---\n\n# Energy\n\n$E=mc^2$\n"}
	parts := func() string {
		t.Helper()
		req.ID++
		r := Serve(req)
		if r.Error != "" {
			t.Fatal(r.Error)
		}
		return strings.Join(r.Parts, "")
	}
	before := parts()
	if !strings.Contains(before, `class="math-error"`) || strings.Contains(before, `class="slides-math"`) || !strings.Contains(before, "$E=mc^2$") {
		t.Fatalf("before the font: %.400s", before)
	}
	late.loaded.Store(true)
	mathfont.Register(late)
	ForgetRenderedSlides()
	after := parts()
	if strings.Contains(after, `class="math-error"`) || !strings.Contains(after, `class="slides-math"`) {
		t.Fatalf("after the font: %.400s", after)
	}
	// and the render after that is the same, from the cache this time
	if again := parts(); again != after {
		t.Error("a third render differs from the second")
	}
	// a failure makes the same forgetting: the next render asks the provider again
	late.loaded.Store(false)
	mathfont.Register(late)
	ForgetRenderedSlides()
	if got := parts(); !strings.Contains(got, `class="math-error"`) {
		t.Errorf("a font that went away still typesets: %.400s", got)
	}
}
