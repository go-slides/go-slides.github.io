package playground

import (
	"github.com/go-widgets/mvvm"
	"github.com/go-widgets/toolkit"
)

// The Git column: the Git pane, and under it, only while the server has refused a
// call for want of the user's own token (vm.GitCredForge), three buttons: Store on
// the server (opens the server's credential page in a new tab), I stored it, retry,
// and Forget my token on the server (gitdvm.go). The toolkit's ForgePane has no
// control for this, and the buttons carry no text field: the token is never typed
// into this page.

// gitCredRows is how many buttons the strip has.
const gitCredRows = 3

// newGitColumn is the Git pane over its credential strip, as one widget for the
// splitter.
func (s *Scene) newGitColumn() toolkit.Widget {
	vm := s.vm
	button := func(label Key, c *mvvm.Command) *toolkit.Button {
		b := toolkit.NewButton(T(label), nil)
		s.unbind = append(s.unbind, mvvm.BindCommand(c, &b.OnClick, func(on bool) {
			b.Disabled().Set(!on)
			s.invalidate()
		}))
		return b
	}
	strip := toolkit.NewVBox()
	strip.Spacing = toolkit.Scaled(6)
	s.credBtns = map[string]*toolkit.Button{
		"store":  button(KLblGitCredStore, vm.GitStoreToken),
		"retry":  button(KLblGitCredRetry, vm.GitRetry),
		"forget": button(KLblGitCredForget, vm.GitForgetToken),
	}
	for _, k := range []string{"store", "retry", "forget"} {
		strip.AddFixed(s.credBtns[k], toolkit.Scaled(28))
	}

	with := toolkit.NewVBox()
	with.Spacing = toolkit.Scaled(6)
	with.AddFlex(s.gitp, 1)
	with.AddFixed(strip, gitCredRows*toolkit.Scaled(28)+(gitCredRows-1)*toolkit.Scaled(6))

	col := toolkit.NewStack()
	col.AddPage("plain", s.gitp)
	col.AddPage("strip", with)
	show := func(forge string) {
		if forge != "" {
			col.SetVisible("strip")
		} else {
			col.SetVisible("plain")
		}
		col.SetBounds(col.Bounds()) // the page that came forward takes the column's rectangle
		s.invalidate()
	}
	s.unbind = append(s.unbind, vm.GitCredForge.Subscribe(show))
	show(vm.GitCredForge.Get())
	s.gitcol = col
	return col
}

// credColumn is the Git column (for the tests): its visible page is "strip" while
// the credential buttons are shown.
func (s *Scene) credColumn() *toolkit.Stack { return s.gitcol }
