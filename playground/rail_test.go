package playground

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func longDeck(n int) string {
	var b strings.Builder
	b.WriteString("---\nslides: true\ntheme: default\n---\n\n")
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString("\n---\n\n")
		}
		fmt.Fprintf(&b, "# Slide %d\n\ntext\n", i+1)
	}
	return b.String()
}

func TestSlideLinesAndSlideAt(t *testing.T) {
	parts := []string{
		`<svg data-marpit-svg="" class="code-line" data-line="5"></svg>`,
		`<svg data-marpit-svg=""></svg>`,
		`<svg data-marpit-svg="" class="code-line" data-line="16"></svg>`,
	}
	lines := SlideLines(parts)
	if fmt.Sprint(lines) != "[5 -1 16]" {
		t.Fatalf("SlideLines = %v", lines)
	}
	for _, c := range []struct{ line, want int }{{0, 0}, {5, 0}, {15, 0}, {16, 2}, {400, 2}} {
		if got := SlideAt(lines, c.line); got != c.want {
			t.Errorf("SlideAt(%d) = %d, want %d", c.line, got, c.want)
		}
	}
	if SlideAt(nil, 3) != -1 {
		t.Fatal("no slides")
	}
}

// The rail page holds the same slides as the preview, on a transparent page,
// under the preview's own CSP, with a script that takes only the parent's
// messages.
func TestRailPage(t *testing.T) {
	r := Serve(Request{ID: 1, Source: longDeck(5)})
	page := railPage(r.Head, r.Parts)
	for _, want := range []string{
		"Content-Security-Policy", "background:transparent!important", "content-visibility:auto",
		`data-line="`, "goSlidesRail", "goSlidesPatch", "e.source!==parentWin",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("rail page lacks %q", want)
		}
	}
	if n := strings.Count(page, "data-marpit-svg"); n < 5 {
		t.Errorf("rail page holds %d slides, want 5", n)
	}
	if strings.Contains(page, "goSlidesLine") {
		t.Error("the rail must not report clicks: the strip takes them")
	}
}

// Slides, the current slide and the strip follow the deck, the caret and the
// strip's own selection, and choosing a thumbnail takes the editor to its slide.
func TestThumbnailsFollowTheDeckAndMoveTheEditor(t *testing.T) {
	s, vm, _, _ := newScene(t)
	vm.Source.Set(longDeck(30))
	if vm.SlidesCount() != 30 || s.thumbs.Count() != 30 {
		t.Fatalf("%d slides in the view-model, %d in the strip", vm.SlidesCount(), s.thumbs.Count())
	}
	buf := make([]byte, 4*1280*800)
	s.Draw(buf)
	// The caret chooses the current slide.
	s.GoToLine(vm.SlideLines[7] + 1)
	if vm.CurrentSlide.Get() != 7 || s.thumbs.Current().Get() != 7 {
		t.Fatalf("caret in slide 8: current %d, strip %d", vm.CurrentSlide.Get(), s.thumbs.Current().Get())
	}
	// A click on a thumbnail takes the editor to the slide's first line.
	s.Draw(buf)
	c := s.thumbs.CardRect(5)
	if b := s.thumbs.Bounds(); c.W == 0 || c.Y+c.H > b.Y+b.H {
		t.Fatalf("slide 6 has no card in the window: %+v in %+v", c, b)
	}
	s.Click(c.X+c.W/2, c.Y+c.H/2)
	s.Release(c.X+c.W/2, c.Y+c.H/2)
	if l, _ := s.EditorCursor(); l != vm.SlideLines[5] {
		t.Fatalf("clicked slide 6: caret at line %d, want %d", l, vm.SlideLines[5])
	}
	if vm.CurrentSlide.Get() != 5 {
		t.Fatalf("current slide %d, want 5", vm.CurrentSlide.Get())
	}
	// The preview's own scroll reports are held off for a moment after the caret
	// chose, then followed.
	vm.PreviewScrolled(SlideTop{I: 3})
	if vm.CurrentSlide.Get() != 5 {
		t.Fatal("a scroll report right after a caret move pulled the selection off")
	}
	vm.holdSlide = time.Now().Add(-time.Second)
	vm.PreviewScrolled(SlideTop{I: 3})
	if vm.CurrentSlide.Get() != 3 {
		t.Fatalf("the preview scrolled to slide 4: current %d", vm.CurrentSlide.Get())
	}
	// Keys go to the strip once it holds the keyboard.
	c = s.thumbs.CardRect(3)
	s.Click(c.X+c.W/2, c.Y+c.H/2)
	s.Release(c.X+c.W/2, c.Y+c.H/2)
	s.KeyDown("ArrowDown")
	if vm.CurrentSlide.Get() != 4 {
		t.Fatalf("ArrowDown in the strip: current %d, want 4", vm.CurrentSlide.Get())
	}
	if s.editing() {
		t.Fatal("the editor kept the keyboard")
	}
}

func TestTogglesShowAndHideTheMinimapAndThumbnails(t *testing.T) {
	s, vm, _, _ := newScene(t)
	vm.Source.Set(longDeck(10))
	buf := make([]byte, 4*1280*800)
	s.Draw(buf)
	if !s.editor.ShowMinimap {
		t.Fatal("the minimap starts on")
	}
	if _, ok := s.ThumbsRect(); !ok {
		t.Fatal("the thumbnails start on")
	}
	wide := s.thumbs.Bounds().W
	vm.Minimap.Set(false)
	if s.editor.ShowMinimap {
		t.Fatal("the minimap toggle did not reach the editor")
	}
	vm.Thumbs.Set(false)
	s.Draw(buf)
	if _, ok := s.ThumbsRect(); ok {
		t.Fatal("thumbnails off: the frame should be hidden")
	}
	if got := s.thumbs.Bounds().W; got >= wide/4 {
		t.Fatalf("thumbnails off: the column is still %d wide (was %d)", got, wide)
	}
	vm.Thumbs.Set(true)
	s.Draw(buf)
	if s.thumbs.Bounds().W != wide {
		t.Fatalf("thumbnails back on: width %d, want %d", s.thumbs.Bounds().W, wide)
	}
	// The cards the frame is laid from.
	first, gap := s.ThumbsCards()
	if first.W <= 0 || first.H <= 0 || gap <= 0 {
		t.Fatalf("cards: %+v gap %d", first, gap)
	}
	if first.H*16 > first.W*9+16 || first.H*16 < first.W*9-16 {
		t.Fatalf("a card is not 16:9: %+v", first)
	}
}

// The frame stacks its slides from the first card, the card size and the gap, so
// every card the canvas draws must be where that arithmetic puts it (within a
// device pixel), at every scroll and pixel ratio.
func TestFrameArithmeticMatchesEveryCard(t *testing.T) {
	for _, scale := range []float64{1, 1.5, 2} {
		SetupText(scale)
		_, vm, tr, h := func() (*Scene, *ViewModel, *fakeTransport, *fakeHost) {
			tr := &fakeTransport{}
			h := newHost()
			vm := NewViewModel(tr, h, "")
			answer(tr, vm)
			return nil, vm, tr, h
		}()
		_ = tr
		s := NewScene(vm, h, int(1500*scale), int(900*scale))
		vm.Source.Set(longDeck(100))
		buf := make([]byte, 4*s.w*s.h)
		for _, scroll := range []int{0, 1234, 5000, 1 << 20} {
			s.thumbs.Scroll().Set(scroll)
			s.Draw(buf)
			strip := s.thumbs.Bounds()
			first, gap := s.ThumbsCards()
			f, cards := s.VisibleThumbCards()
			if len(cards) == 0 {
				t.Fatalf("scale %v scroll %d: no cards", scale, scroll)
			}
			c0 := s.thumbs.CardRect(0) // may be above the window: still on the line
			for i, c := range cards {
				idx := f + i
				// the first card's position is read at the current scroll
				wantY := c0.Y + idx*(first.H+gap)
				if d := c.Y - wantY; d < -1 || d > 1 {
					t.Errorf("scale %v scroll %d card %d: y %d, frame arithmetic %d", scale, scroll, idx, c.Y, wantY)
				}
				if c.X != strip.X+first.X || c.W != first.W || c.H != first.H {
					t.Errorf("scale %v scroll %d card %d: %+v vs first %+v in strip %+v", scale, scroll, idx, c, first, strip)
				}
			}
		}
	}
	SetupText(1)
}

// A rail frame that missed a patch is reloaded with the slides the preview holds.
func TestRailPageIsTheCurrentOverview(t *testing.T) {
	_, vm, _, _ := newScene(t)
	if vm.RailPage() == "" {
		t.Fatal("the first render's slides are the rail's: it has them to load")
	}
	vm.Source.Set(longDeck(4))
	vm.PreviewLoaded()
	page := vm.RailPage()
	if page == "" || strings.Count(page, "data-marpit-svg") < 4 {
		t.Fatalf("rail page: %d bytes", len(page))
	}
	if vm.RailDoc.Get() == "" {
		t.Fatal("the rail document was not published")
	}
}
