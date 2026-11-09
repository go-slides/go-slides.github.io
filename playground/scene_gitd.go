package playground

import (
	"fmt"
	"strings"

	"github.com/go-widgets/mvvmtk"
	"github.com/go-widgets/toolkit"
)

// The git pane in server mode (gitdvm.go): who is signed in, the repositories
// the server grants, and the actions. There is no token field, no Remember box
// and no OAuth button: this page holds no credential.

// gitServerPanel is the server mode's widgets.
type gitServerPanel struct {
	box                     *toolkit.VBox
	title, user, reposLabel *toolkit.Label
	note                    *toolkit.TextView
	pick                    *toolkit.DropDown
	ref, dir                *toolkit.Entry
	signin, load, propose   *toolkit.Button
	refresh, theirs, open   *toolkit.Button
}

func (s *Scene) newGitServerPanel(pair func(a, b toolkit.Widget) *toolkit.HBox) *gitServerPanel {
	vm, inv, sc := s.vm, s.invalidate, toolkit.Scaled
	p := &gitServerPanel{
		title:      toolkit.NewLabel(T(KLblGitServerTitle)).SetFontSize(sc(baseFontPx + 2)),
		user:       toolkit.NewLabel(""),
		reposLabel: toolkit.NewLabel(T(KLblGitServerRepos)),
		note:       toolkit.NewTextView(""),
		pick:       toolkit.NewDropDown(nil, 0),
		ref:        toolkit.NewEntry(vm.GitRef.Get()),
		dir:        toolkit.NewEntry(vm.GitDir.Get()),
		signin:     toolkit.NewButton(T(KLblGitServerSignIn), nil),
		load:       toolkit.NewButton(T(KLblGitLoad), nil),
		propose:    toolkit.NewButton(T(KLblGitPropose), nil),
		refresh:    toolkit.NewButton(T(KLblGitRefresh), nil),
		theirs:     toolkit.NewButton(T(KLblGitTakeTheirs), nil),
		open:       toolkit.NewButton(T(KLblGitOpenLink), nil),
	}
	if f := uiFont(sc(baseFontPx + 2)); f != nil {
		p.title.FontSize, p.title.Font = 0, f
	}
	p.ref.Placeholder, p.dir.Placeholder = T(KLblGitBranch), T(KLblGitFolder)
	wrap := func() {
		cols := vm.GitWidth.Get() * 100 / max(baseFontPx*55, 1)
		p.note.SetText(Wrap(vm.GitNote.Get(), max(cols-2, 20)))
		inv()
	}
	wrap()
	who := func() {
		if u := vm.GitUser.Get(); u != "" {
			p.user.Text().Set(fmt.Sprintf(T(KLblGitServerUser), u))
		} else {
			p.user.Text().Set("")
		}
		inv()
	}
	who()
	repos := func() {
		if r := vm.GitRepos.Get(); r != "" {
			p.pick.Options = strings.Split(r, "\n")
		} else {
			p.pick.Options = nil
		}
		p.pick.Selected().Set(vm.GitRepoPick.Get())
		inv()
	}
	repos()
	s.unbind = append(s.unbind,
		mvvmtk.BindEntryText(p.ref, vm.GitRef, inv),
		mvvmtk.BindEntryText(p.dir, vm.GitDir, inv),
		mvvmtk.BindSelectedIndex(p.pick, vm.GitRepoPick, inv),
		mvvmtk.BindCommand(p.signin, vm.GitServerSignIn, inv),
		mvvmtk.BindCommand(p.load, vm.GitLoad, inv),
		mvvmtk.BindCommand(p.propose, vm.GitPropose, inv),
		mvvmtk.BindCommand(p.refresh, vm.GitRefresh, inv),
		mvvmtk.BindCommand(p.theirs, vm.GitTakeTheirs, inv),
		mvvmtk.BindCommand(p.open, vm.GitOpenLink, inv),
		vm.GitWidth.Subscribe(func(int) { wrap() }),
		vm.GitNote.Subscribe(func(string) { wrap() }),
		vm.GitUser.Subscribe(func(string) { who() }),
		vm.GitRepos.Subscribe(func(string) { repos() }),
	)
	p.box = toolkit.NewVBox()
	p.box.Spacing = sc(6)
	p.box.AddFixed(p.title, sc(row))
	p.box.AddFixed(p.user, sc(row-8))
	p.box.AddFixed(p.signin, sc(row-4))
	p.box.AddFixed(p.reposLabel, sc(row-8))
	p.box.AddFixed(p.pick, sc(row-4))
	p.box.AddFixed(p.ref, sc(row-4))
	p.box.AddFixed(p.dir, sc(row-4))
	p.box.AddFixed(pair(p.load, p.refresh), sc(row-4))
	p.box.AddFixed(pair(p.propose, p.theirs), sc(row-4))
	p.box.AddFixed(p.open, sc(row-4))
	p.box.AddFlex(p.note, 1)
	return p
}
