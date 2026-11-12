package playground

import (
	"slices"
	"strings"

	"github.com/go-widgets/mvvm"
	"github.com/go-widgets/toolkit"
)

// forgeModel is the view-model's side of toolkit.ForgePane: the pane the toolkit
// draws is over this interface, and what answers it here is the observables and
// commands the view-model already has (gitvm.go, gitdvm.go, gitframe.go). Two are
// derived: the layout in use (the view-model has a mode and a frame, the pane one
// choice of three) and the granted repositories (a text of lines here, a list
// there).
type forgeModel struct {
	vm     *ViewModel
	mode   *mvvm.Observable[toolkit.ForgeMode]
	repos  *mvvm.Observable[[]string]
	unbind []func()
}

var _ toolkit.ForgeModel = (*forgeModel)(nil)

// newForgeModel derives the two and follows the view-model; close detaches it.
func newForgeModel(vm *ViewModel) *forgeModel {
	m := &forgeModel{
		vm:    vm,
		mode:  mvvm.NewObservable(toolkit.ForgeClient),
		repos: mvvm.NewObservableEq([]string(nil), slices.Equal[[]string]),
	}
	m.syncMode(vm.GitMode.Get())
	m.syncRepos(vm.GitRepos.Get())
	m.unbind = []func(){
		vm.GitMode.Subscribe(func(i int) { m.syncMode(i) }),
		vm.GitRepos.Subscribe(m.syncRepos),
	}
	return m
}

// close stops following the view-model.
func (m *forgeModel) close() {
	for _, u := range m.unbind {
		u()
	}
	m.unbind = nil
}

// syncMode: a server in charge is the server layout; else the frame, when this
// page runs its client mode through one; else the page's own token field.
func (m *forgeModel) syncMode(gitMode int) {
	switch {
	case gitMode != GitModeClient:
		m.mode.Set(toolkit.ForgeServer)
	case m.vm.frameOn():
		m.mode.Set(toolkit.ForgeFrame)
	default:
		m.mode.Set(toolkit.ForgeClient)
	}
}

func (m *forgeModel) syncRepos(lines string) {
	if lines == "" {
		m.repos.Set(nil)
		return
	}
	m.repos.Set(strings.Split(lines, "\n"))
}

func (m *forgeModel) Forges() []string                          { return GitForges }
func (m *forgeModel) Forge() *mvvm.Observable[int]              { return m.vm.GitForge }
func (m *forgeModel) Repo() *mvvm.Observable[string]            { return m.vm.GitRepo }
func (m *forgeModel) Ref() *mvvm.Observable[string]             { return m.vm.GitRef }
func (m *forgeModel) Dir() *mvvm.Observable[string]             { return m.vm.GitDir }
func (m *forgeModel) Token() *mvvm.Observable[string]           { return m.vm.GitToken }
func (m *forgeModel) Remember() *mvvm.Observable[bool]          { return m.vm.GitRemember }
func (m *forgeModel) Note() *mvvm.Observable[string]            { return m.vm.GitNote }
func (m *forgeModel) Mode() *mvvm.Observable[toolkit.ForgeMode] { return m.mode }
func (m *forgeModel) User() *mvvm.Observable[string]            { return m.vm.GitUser }
func (m *forgeModel) Repos() *mvvm.Observable[[]string]         { return m.repos }
func (m *forgeModel) RepoPick() *mvvm.Observable[int]           { return m.vm.GitRepoPick }
func (m *forgeModel) Connect() *mvvm.Command                    { return m.vm.GitConnect }
func (m *forgeModel) Disconnect() *mvvm.Command                 { return m.vm.GitDisconnect }
func (m *forgeModel) SignIn() *mvvm.Command                     { return m.vm.GitSignIn }
func (m *forgeModel) ServerSignIn() *mvvm.Command               { return m.vm.GitServerSignIn }
func (m *forgeModel) Load() *mvvm.Command                       { return m.vm.GitLoad }
func (m *forgeModel) Propose() *mvvm.Command                    { return m.vm.GitPropose }
func (m *forgeModel) Refresh() *mvvm.Command                    { return m.vm.GitRefresh }
func (m *forgeModel) TakeTheirs() *mvvm.Command                 { return m.vm.GitTakeTheirs }
func (m *forgeModel) OpenLink() *mvvm.Command                   { return m.vm.GitOpenLink }
