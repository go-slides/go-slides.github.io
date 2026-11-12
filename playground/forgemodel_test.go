package playground

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-widgets/toolkit"
)

// noFrame is a FrameHost that is only there to be one.
type noFrame struct{}

func (noFrame) FrameMode() bool { return true }
func (noFrame) FrameCall(string, map[string]any, remote.Files) (json.RawMessage, remote.Files, error) {
	return nil, nil, nil
}
func (noFrame) FrameWatch(func(FrameState)) {}

// The toolkit's pane is over the view-model through forgeModel: the layout is the
// view-model's mode and frame, the repositories its lines, and the rest is its own
// observables and commands, not copies.
func TestForgeModelIsTheViewModel(t *testing.T) {
	_, vm, _ := gitScene(t, store.NewMemory(), "")
	m := newForgeModel(vm)
	defer m.close()
	if m.Mode().Get() != toolkit.ForgeClient {
		t.Fatalf("a page with a token field: %v", m.Mode().Get())
	}
	vm.GitMode.Set(GitModeSignedIn)
	if m.Mode().Get() != toolkit.ForgeServer {
		t.Error("a server in charge is the server layout")
	}
	vm.git.frame = noFrame{}
	vm.GitMode.Set(GitModeClient)
	if m.Mode().Get() != toolkit.ForgeFrame {
		t.Error("client mode through a frame is the frame layout")
	}
	vm.git.frame = nil
	vm.GitMode.Set(GitModeSignedOut)
	vm.GitMode.Set(GitModeClient)
	if m.Mode().Get() != toolkit.ForgeClient {
		t.Error("back to the token field")
	}

	vm.GitRepos.Set("github:o/one\ngitlab:o/two")
	if got := m.Repos().Get(); !slices.Equal(got, []string{"github:o/one", "gitlab:o/two"}) {
		t.Errorf("repos %v", got)
	}
	vm.GitRepos.Set("")
	if m.Repos().Get() != nil {
		t.Errorf("no repositories: %v", m.Repos().Get())
	}

	if !slices.Equal(m.Forges(), GitForges) || m.Forge() != vm.GitForge || m.Repo() != vm.GitRepo || m.Ref() != vm.GitRef ||
		m.Dir() != vm.GitDir || m.Token() != vm.GitToken || m.Remember() != vm.GitRemember || m.Note() != vm.GitNote ||
		m.User() != vm.GitUser || m.RepoPick() != vm.GitRepoPick {
		t.Error("an observable is not the view-model's")
	}
	if m.Connect() != vm.GitConnect || m.Disconnect() != vm.GitDisconnect || m.SignIn() != vm.GitSignIn ||
		m.ServerSignIn() != vm.GitServerSignIn || m.Load() != vm.GitLoad || m.Propose() != vm.GitPropose ||
		m.Refresh() != vm.GitRefresh || m.TakeTheirs() != vm.GitTakeTheirs || m.OpenLink() != vm.GitOpenLink {
		t.Error("a command is not the view-model's")
	}

	m.close()
	vm.GitRepos.Set("github:o/three")
	if m.Repos().Get() != nil {
		t.Error("a closed model still follows the view-model")
	}
}

// The pane's words are the playground's own, in the language in use: one
// translation for each word, whichever key asks.
func TestTheForgePaneSpeaksTheLanguage(t *testing.T) {
	t.Cleanup(func() { SetLocale("en") })
	for tk, k := range forgeWords {
		for _, l := range Langs {
			if catalogs[l][tk] == "" {
				t.Errorf("%s has no %s", l, tk)
			}
			if tk != "tk.forge.title" && catalogs[l][tk] != catalogs[l][k] {
				t.Errorf("%s: %s is %q, %s is %q", l, tk, catalogs[l][tk], k, catalogs[l][k])
			}
		}
	}
	SetLocale("fr")
	s, vm, _ := gitScene(t, store.NewMemory(), "")
	_ = vm
	vm.GitPane.Set(true)
	s.layout()
	found := false
	for _, n := range toolkit.WalkA11y(s.gitp) {
		found = found || n.Name == T(KLblGitLoad)
	}
	if !found || T(KLblGitLoad) == "Load" {
		t.Errorf("the pane did not say Load in French: %q", T(KLblGitLoad))
	}
}
