package playground

import (
	"github.com/go-widgets/painter"
	"github.com/go-widgets/toolkit"
)

// slot is a pane as the main row lays it out: a column of a natural width, or a
// sliver when the pane is switched off (a row cannot drop a child, so the pane
// draws nothing, hears nothing and asks for one pixel).
type slot struct {
	toolkit.Widget
	on    func() bool
	width int // logical pixels
}

// Measure is the column's width (Measurer, for HBox.AddNatural).
func (s *slot) Measure(_, availH int) (int, int) {
	if !s.on() {
		return 1, availH
	}
	return toolkit.Scaled(s.width), availH
}

// Draw paints the pane when it is on.
func (s *slot) Draw(p painter.Painter, theme *toolkit.Theme) {
	if s.on() {
		s.Widget.Draw(p, theme)
	}
}

// OnEvent gives the pane its events when it is on.
func (s *slot) OnEvent(ev toolkit.Event) {
	if s.on() {
		s.Widget.OnEvent(ev)
	}
}

// HitTest is the pane's when it is on.
func (s *slot) HitTest(x, y int) bool { return s.on() && s.Widget.HitTest(x, y) }

// Children is the pane, for the walks over the tree (the open popovers, the
// accessibility tree), while it is on.
func (s *slot) Children() []toolkit.Widget {
	if !s.on() {
		return nil
	}
	return []toolkit.Widget{s.Widget}
}

// Pane widths, in logical pixels.
const (
	filesWidth = 190
	helpWidth  = 280
)
