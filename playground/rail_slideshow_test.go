package playground

import (
	"strings"
	"testing"
)

// A deck reopened in the slideshow (the draft keeps the mode) fills the
// thumbnail column as the overview does: the render still returns the slides.
func TestThumbnailColumnIsFilledInTheSlideshow(t *testing.T) {
	r := Serve(Request{ID: 1, Source: longDeck(6), Mode: ModeSlideshow})
	if r.Doc == "" || len(r.Parts) != 6 || r.Head == "" {
		t.Fatalf("slideshow answer: doc %d bytes, %d parts, head %d bytes", len(r.Doc), len(r.Parts), len(r.Head))
	}

	s, vm, tr, h := newScene(t)
	vm.Mode.Set(int(ModeSlideshow))
	tr.hold = false
	vm.Source.Set(longDeck(6))
	if vm.SlidesCount() != 6 || s.thumbs.Count() != 6 {
		t.Fatalf("slideshow: %d slides in the view-model, %d in the strip", vm.SlidesCount(), s.thumbs.Count())
	}
	if page := vm.RailPage(); strings.Count(page, "data-marpit-svg") < 6 {
		t.Fatalf("slideshow: the rail document holds %d slides", strings.Count(page, "data-marpit-svg"))
	}
	if vm.RailDoc.Get() == "" {
		t.Fatal("slideshow: the rail document was not published")
	}
	buf := make([]byte, 4*s.w*s.h)
	s.Draw(buf)
	if _, ok := s.ThumbsRect(); !ok {
		t.Fatal("slideshow: the thumbnail frame is hidden while the column is on")
	}

	// An edit patches the frame's slides, as in the overview, and never reloads it.
	before := len(h.patches)
	docs := vm.RailDoc.Get()
	vm.Source.Set(strings.Replace(longDeck(6), "# Slide 3", "# Third", 1))
	if len(h.patches) != before+1 || len(h.patches[before].Changed) != 1 {
		t.Fatalf("slideshow edit: %d patches (%d before), want one changed slide", len(h.patches), before)
	}
	if vm.RailDoc.Get() != docs {
		t.Fatal("slideshow edit reloaded the thumbnail frame")
	}
	// Back to the overview: the frame holds slides again with no gap.
	vm.Mode.Set(int(ModeOverview))
	if vm.SlidesCount() != 6 || vm.RailPage() == "" {
		t.Fatalf("overview again: %d slides", vm.SlidesCount())
	}
}
