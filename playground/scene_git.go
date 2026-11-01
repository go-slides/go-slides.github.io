package playground

import (
	"github.com/go-widgets/mvvmtk"
	"github.com/go-widgets/toolkit"
)

// The GitHub pane (gitvm.go): a column of the splitter, right of the
// versions. The token field is masked (toolkit.Entry.Mask: bullets on screen,
// never copied to the clipboard) and the view-model empties it as soon as
// Connect takes the token.

// gitWidth and its range, in logical pixels.
const gitWidth, gitMin, gitMax = 300, 240, 600

// gitPanel is the pane's widgets.
type gitPanel struct {
	box                                *toolkit.VBox
	title                              *toolkit.Label
	note                               *toolkit.TextView
	repo, ref, dir, token              *toolkit.Entry
	remember                           *toolkit.CheckButton
	connect, disconnect, load, propose *toolkit.Button
	refresh, theirs, open              *toolkit.Button
}

// newGitPanel builds the pane over the view-model.
func (s *Scene) newGitPanel() *gitPanel {
	vm, inv, sc := s.vm, s.invalidate, toolkit.Scaled
	entry := func(o string) *toolkit.Entry { return toolkit.NewEntry(o) }
	p := &gitPanel{
		title:      toolkit.NewLabel(T(KLblGit)).SetFontSize(sc(baseFontPx + 2)),
		note:       toolkit.NewTextView(""),
		repo:       entry(vm.GitRepo.Get()),
		ref:        entry(vm.GitRef.Get()),
		dir:        entry(vm.GitDir.Get()),
		token:      entry(""),
		remember:   toolkit.NewCheckButton(T(KLblGitRemember), vm.GitRemember.Get()),
		connect:    toolkit.NewButton(T(KLblGitConnect), nil),
		disconnect: toolkit.NewButton(T(KLblGitDisconnect), nil),
		load:       toolkit.NewButton(T(KLblGitLoad), nil),
		propose:    toolkit.NewButton(T(KLblGitPropose), nil),
		refresh:    toolkit.NewButton(T(KLblGitRefresh), nil),
		theirs:     toolkit.NewButton(T(KLblGitTakeTheirs), nil),
		open:       toolkit.NewButton(T(KLblGitOpenLink), nil),
	}
	if f := uiFont(sc(baseFontPx + 2)); f != nil {
		p.title.FontSize, p.title.Font = 0, f
	}
	p.token.Mask = '•'
	p.repo.Placeholder, p.ref.Placeholder, p.dir.Placeholder, p.token.Placeholder = T(KLblGitRepo), T(KLblGitBranch), T(KLblGitFolder), T(KLblGitToken)
	// the note wraps to the pane's width (a TextView does not wrap)
	wrap := func() {
		cols := vm.GitWidth.Get() * 100 / max(baseFontPx*55, 1)
		p.note.SetText(Wrap(vm.GitNote.Get(), max(cols-2, 20)))
		inv()
	}
	wrap()
	s.unbind = append(s.unbind,
		mvvmtk.BindEntryText(p.repo, vm.GitRepo, inv),
		mvvmtk.BindEntryText(p.ref, vm.GitRef, inv),
		mvvmtk.BindEntryText(p.dir, vm.GitDir, inv),
		mvvmtk.BindEntryText(p.token, vm.GitToken, inv),
		vm.GitWidth.Subscribe(func(int) { wrap() }),
		mvvmtk.BindChecked(p.remember, vm.GitRemember, inv),
		mvvmtk.BindCommand(p.connect, vm.GitConnect, inv),
		mvvmtk.BindCommand(p.disconnect, vm.GitDisconnect, inv),
		mvvmtk.BindCommand(p.load, vm.GitLoad, inv),
		mvvmtk.BindCommand(p.propose, vm.GitPropose, inv),
		mvvmtk.BindCommand(p.refresh, vm.GitRefresh, inv),
		mvvmtk.BindCommand(p.theirs, vm.GitTakeTheirs, inv),
		mvvmtk.BindCommand(p.open, vm.GitOpenLink, inv),
		vm.GitNote.Subscribe(func(string) { wrap() }),
	)
	pair := func(a, b toolkit.Widget) *toolkit.HBox {
		h := toolkit.NewHBox()
		h.Spacing = sc(6)
		h.AddFlex(a, 1)
		h.AddFlex(b, 1)
		return h
	}
	p.box = toolkit.NewVBox()
	p.box.Spacing = sc(6)
	p.box.AddFixed(p.title, sc(row))
	p.box.AddFixed(p.repo, sc(row-4))
	p.box.AddFixed(p.ref, sc(row-4))
	p.box.AddFixed(p.dir, sc(row-4))
	p.box.AddFixed(p.token, sc(row-4))
	p.box.AddFixed(p.remember, sc(row-8))
	p.box.AddFixed(pair(p.connect, p.disconnect), sc(row-4))
	p.box.AddFixed(pair(p.load, p.refresh), sc(row-4))
	p.box.AddFixed(pair(p.propose, p.theirs), sc(row-4))
	p.box.AddFixed(p.open, sc(row-4))
	p.box.AddFlex(p.note, 1)
	return p
}

// gitRects are the pane's named rectangles, for a test harness.
func (s *Scene) gitRects(r map[string]toolkit.Rect) {
	if s.gitp == nil || !s.vm.GitPane.Get() {
		return
	}
	p := s.gitp
	for k, w := range map[string]toolkit.Widget{
		"git-repo": p.repo, "git-ref": p.ref, "git-dir": p.dir, "git-token": p.token, "git-remember": p.remember,
		"git-connect": p.connect, "git-disconnect": p.disconnect, "git-load": p.load, "git-propose": p.propose,
		"git-refresh": p.refresh, "git-theirs": p.theirs, "git-open": p.open,
	} {
		r[k] = w.Bounds()
	}
}
