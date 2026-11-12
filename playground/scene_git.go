package playground

import (
	"github.com/go-widgets/toolkit"
)

// The Git pane (gitvm.go, gitdvm.go, gitframe.go): a column of the splitter,
// right of the versions. It is toolkit.ForgePane, the toolkit's widget for the
// pane of an editor that works against a code forge, over forgeModel (the
// view-model's observables and commands). Its token field is masked
// (toolkit.Entry.Mask: bullets on screen, never copied to the clipboard) and the
// view-model empties it as soon as Connect takes the token.

// gitWidth and its range, in logical pixels.
const gitWidth, gitMin, gitMax = 300, 240, 600

// newGitPane builds the pane over the view-model.
func (s *Scene) newGitPane() *toolkit.ForgePane {
	s.forge = newForgeModel(s.vm)
	p := toolkit.NewForgePane(s.forge)
	p.Invalidate = s.invalidate
	p.NoteWrap = Wrap // breaks Chinese, Japanese and Korean between runes, keeps list indents
	p.SetTitle(T(KLblGit))
	p.SetTitleFont(uiFont(toolkit.Scaled(baseFontPx + 2)))
	s.unbind = append(s.unbind, p.Close, s.forge.close)
	return p
}

// GitFrameRect is where the git frame's iframe goes: the pane's slot, while the
// pane is open in client mode with a frame. The host puts the iframe there.
func (s *Scene) GitFrameRect() (toolkit.Rect, bool) {
	if s.gitp == nil || !s.vm.GitPane.Get() {
		return toolkit.Rect{}, false
	}
	return s.gitp.FrameRect()
}

// gitRects are the pane's named rectangles, for a test harness: the controls of
// the layout in use, each as "git-<name>". The frame layout has no token,
// Remember, sign-in, Connect or forge control, so a test can assert it.
func (s *Scene) gitRects(r map[string]toolkit.Rect) {
	if s.gitp == nil || !s.vm.GitPane.Get() {
		return
	}
	for k, v := range s.gitp.Parts() {
		r["git-"+k] = v
	}
}
