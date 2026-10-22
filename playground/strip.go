package playground

import (
	"github.com/go-widgets/painter"
	"github.com/go-widgets/toolkit"
)

// thumbsWidth is the logical width of the thumbnail column: a 120-pixel card,
// its number, and the scrollbar.
const thumbsWidth = 168

// stripSlot is the thumbnail strip as the main row lays it out: a column of its
// natural width, or a sliver when the thumbnails are switched off (a row cannot
// drop a child, so the strip draws nothing and asks for one pixel).
type stripSlot struct {
	*toolkit.PageStrip
	on func() bool
}

// newStripSlot builds the strip over the view-model's slides: its pages are the
// slides, the current page the current slide, and choosing one moves the editor
// (and through it the preview) to the slide's first line. The thumbnails
// themselves are the frame the host lays over the cards (see rail.go).
func newStripSlot(vm *ViewModel, s *Scene) *stripSlot {
	strip := toolkit.NewPageStrip(0)
	strip.PageW, strip.PageH = 16, 9
	vm.SlidesVersion.Subscribe(func(int) { strip.SetCount(len(vm.SlideLines)); s.dirty = true })
	vm.CurrentSlide.Subscribe(func(i int) { strip.Current().Set(i); s.dirty = true })
	strip.Current().Subscribe(func(i int) {
		if i == vm.CurrentSlide.Get() || i < 0 || i >= len(vm.SlideLines) || vm.SlideLines[i] < 0 {
			return // the view-model's own change coming back, or no such slide
		}
		vm.CurrentSlide.Set(i)
		s.GoToLine(vm.SlideLines[i])
	})
	return &stripSlot{PageStrip: strip, on: vm.Thumbs.Get}
}

// Measure is the column's width (Measurer, for HBox.AddNatural).
func (s *stripSlot) Measure(_, availH int) (int, int) {
	if !s.on() {
		return 1, availH
	}
	return toolkit.Scaled(thumbsWidth), availH
}

// Draw paints the strip when the thumbnails are on.
func (s *stripSlot) Draw(p painter.Painter, theme *toolkit.Theme) {
	if s.on() {
		s.PageStrip.Draw(p, theme)
	}
}

// ThumbsRect is where the host places the thumbnail frame, in device pixels;
// ok is false while the thumbnails are off.
func (s *Scene) ThumbsRect() (r toolkit.Rect, ok bool) {
	return s.thumbs.Bounds(), s.vm.Thumbs.Get() && s.vm.SlidesCount() > 0
}

// ThumbsCards is the first card's rectangle relative to ThumbsRect, and the gap
// between cards, in device pixels: the frame stacks the slides from them.
func (s *Scene) ThumbsCards() (first toolkit.Rect, gap int) {
	strip, c0 := s.thumbs.Bounds(), s.thumbs.CardRect(0)
	first = toolkit.Rect{X: c0.X - strip.X, Y: c0.Y - strip.Y, W: c0.W, H: c0.H}
	if s.thumbs.Count() > 1 {
		return first, s.thumbs.CardRect(1).Y - c0.Y - c0.H
	}
	return first, toolkit.Scaled(10)
}

// VisibleThumbCards is the first slide in the strip's window and the rectangles
// of the cards there, in device pixels (the surface's coordinates), as the
// canvas draws them: what a test harness compares the frame's slides with.
func (s *Scene) VisibleThumbCards() (first int, cards []toolkit.Rect) {
	f, l := s.thumbs.Visible()
	for i := f; i <= l; i++ {
		cards = append(cards, s.thumbs.CardRect(i))
	}
	return f, cards
}

// SlidesCount is the number of slides the last overview had.
func (vm *ViewModel) SlidesCount() int { return len(vm.SlideLines) }
