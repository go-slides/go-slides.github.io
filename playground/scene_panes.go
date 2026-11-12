package playground

import "github.com/go-widgets/toolkit"

// The panes' close buttons and dividers, from the keyboard.
//
// Each pane has a close button in its header and the dividers sit between the
// panes: together they are the window's chrome, one ring of stops from left to
// right (a pane's button, the divider after it, the next pane's button…). F6 takes
// the keyboard to the first divider; Tab and Shift+Tab then walk the whole ring,
// the buttons included; Escape gives the keyboard back to the editor. Enter or
// Space on a button closes its pane, and the keyboard goes back to the editor,
// since a closed pane never keeps it. A button that is disabled (the last of the
// editor and the preview) is not a stop.

// chromeStop is one stop of the ring: the close button of pane (pane >= 0) or the
// divider handle (pane < 0).
type chromeStop struct{ pane, handle int }

// chromeStops lists the stops that are on, left to right.
func (s *Scene) chromeStops() []chromeStop {
	var open []int
	for i := range s.panes {
		if s.split.IsOpen(i) {
			open = append(open, i)
		}
	}
	var out []chromeStop
	for n, i := range open {
		if b := s.panes[i].Button; !b.Disabled().Get() && b.Bounds().W > 0 {
			out = append(out, chromeStop{pane: i, handle: -1})
		}
		if n+1 < len(open) {
			out = append(out, chromeStop{pane: -1, handle: n})
		}
	}
	return out
}

// chromeCurrent is the stop that holds the keyboard, or -1.
func (s *Scene) chromeCurrent() int {
	j := s.split.FocusedHandle()
	for k, st := range s.chromeStops() {
		if st.pane < 0 && st.handle == j || st.pane >= 0 && s.panes[st.pane].Button.Focused() {
			return k
		}
	}
	return -1
}

// blurChrome takes the keyboard off every button and divider.
func (s *Scene) blurChrome() {
	s.split.Blur()
	for _, p := range s.panes {
		p.Button.SetFocused(false)
	}
}

// focusStop gives the keyboard to stop k; past either end it goes back to the
// editor.
func (s *Scene) focusStop(k int) {
	stops := s.chromeStops()
	s.blurChrome()
	if k < 0 || k >= len(stops) {
		s.leaveChrome()
		return
	}
	if st := stops[k]; st.pane >= 0 {
		s.panes[st.pane].Button.SetFocused(true)
	} else {
		s.split.FocusHandle(st.handle)
	}
}

// leaveChrome hands the keyboard back to the editor, when it is open: a closed
// pane takes none.
func (s *Scene) leaveChrome() {
	s.blurChrome()
	s.editor.SetFocused(s.split.IsOpen(paneEditor))
}

// SetShift tells whether Shift is held for the key about to be delivered (the
// canvas host hands over the key but not its modifiers): Shift+Tab and Shift+F6
// walk the close buttons and dividers backwards.
func (s *Scene) SetShift(shift bool) { s.shift = shift }

// chromeKey is a key for the stop cur that holds the keyboard: Tab and Shift+Tab
// move to the next and the previous stop, Escape gives the keyboard back to the
// editor; on a button Enter and Space close its pane, on a divider the rest is
// the splitter's (arrows, Home, End, Enter).
func (s *Scene) chromeKey(cur int, ev toolkit.Event) {
	defer func() { s.dirty = true }()
	st := s.chromeStops()[cur]
	switch ev.Code {
	case "Tab":
		s.focusStop(cur + 1 - 2*b2i(ev.Shift || s.shift))
	case "Shift+Tab":
		s.focusStop(cur - 1)
	case "Escape":
		s.leaveChrome()
	default:
		if st.pane < 0 {
			s.split.Handle(st.handle).OnEvent(ev)
		} else if ev.Code == "Enter" || ev.Code == " " || ev.Code == "Space" {
			s.panes[st.pane].Button.OnEvent(ev)
			s.leaveChrome()
		}
	}
}

// paneIDs name the panes in Rects (and so for a test harness): "close-files"
// is the close button of the file panel.
var paneIDs = [paneCount]string{"files", "editor", "thumbs", "preview", "help", "versions", "git", "theme"}
