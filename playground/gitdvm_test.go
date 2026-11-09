package playground

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-browserhttp/browsergit/gitd"
	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-slides/go-slides.github.io/playground/store"
)

// fakeGitd is slides-gitd as the view-model sees it.
type fakeGitd struct {
	info     gitd.Info
	err      error // what Session says
	forge    *fakeForge
	forges   []string // which forge name Remote was asked for
	forgot   int
	sessions int
}

func (g *fakeGitd) Session(context.Context) (gitd.Info, error) {
	g.sessions++
	return g.info, g.err
}
func (g *fakeGitd) Remote(f string) remote.Remote { g.forges = append(g.forges, f); return g.forge }
func (g *fakeGitd) Forget()                       { g.forgot++ }

// a page served behind slides-gitd
func (g *gitTab) Gitd() GitdClient {
	if g.gd == nil {
		return nil
	}
	return g.gd
}
func (g *gitTab) GitdSignIn() { g.signIns++ }

func gitdScene(t *testing.T, m *store.Memory, gd *fakeGitd, keptForge, keptOrigin, kept string) (*Scene, *ViewModel, *gitTab) {
	t.Helper()
	SetupText(1)
	h := &gitTab{
		promptTab: &promptTab{filesTab: &filesTab{tab: &tab{fakeHost: newHost(), answer: true}, saved: map[string][]byte{}}},
		forge: &fakeForge{files: map[string]string{
			"talk/deck.md": "---\ntheme: default\n---\n\n# From the server\n",
		}, branches: map[string]bool{}},
		kept: kept, remoteOK: true, gd: gd,
	}
	if gd != nil {
		gd.forge = h.forge
	}
	h.keptForge, h.keptOrigin = keptForge, keptOrigin
	h.st = store.New(m, NewTabID(), load(m, DraftProject), load(m, PreviousProject), load(m, UIProject))
	tr := &fakeTransport{}
	vm := NewViewModel(tr, h, "")
	tr.vm = vm
	answer(tr, vm)
	vm.drafts.now = func() time.Time { return time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC) }
	return NewScene(vm, h, 1600, 900), vm, h
}

func signedIn(repos ...string) *fakeGitd {
	g := &fakeGitd{info: gitd.Info{User: "sub-1"}}
	for _, r := range repos {
		f, n, _ := strings.Cut(r, ":")
		g.info.Repos = append(g.info.Repos, gitd.Repo{Forge: f, Name: n})
	}
	return g
}

func TestNoServerLeavesThePaneAsItWas(t *testing.T) {
	for name, gd := range map[string]*fakeGitd{
		"a host without one":  nil,
		"a static host's 404": {err: gitd.ErrNoServer},
		"a transport error":   {err: remote.ErrTransport},
	} {
		_, vm, h := gitdScene(t, store.NewMemory(), gd, "", "", "")
		if vm.GitMode.Get() != GitModeClient {
			t.Fatalf("%s: mode %d", name, vm.GitMode.Get())
		}
		if vm.GitNote.Get() != vm.offNote() {
			t.Errorf("%s: the note changed to %q", name, vm.GitNote.Get())
		}
		// the token flow is what it was
		vm.GitRepo.Set("o/r")
		vm.GitToken.Set(planted)
		vm.GitConnect.Execute()
		if !vm.GitConnected.Get() || len(h.tokens) != 1 {
			t.Errorf("%s: a token no longer connects", name)
		}
	}
}

func TestServerModeIsSignedInWithoutAnyToken(t *testing.T) {
	m := store.NewMemory()
	gd := signedIn("github:o/r", "github:o/other")
	// a token a client-mode visit remembered on this device
	_, vm, h := gitdScene(t, m, gd, "github", "https://api.github.com", planted)
	if vm.GitMode.Get() != GitModeSignedIn || vm.GitUser.Get() != "sub-1" || vm.GitRepos.Get() != "github:o/r\ngithub:o/other" || !vm.GitConnected.Get() {
		t.Fatalf("pane: mode %d user %q repos %q", vm.GitMode.Get(), vm.GitUser.Get(), vm.GitRepos.Get())
	}
	if h.kept != "" {
		t.Error("a token remembered on this device is still there in server mode")
	}
	if vm.GitRepo.Get() != "o/r" || !vm.GitLoad.CanExecute() {
		t.Fatalf("first repository not picked: %q", vm.GitRepo.Get())
	}
	vm.GitDir.Set("talk")
	vm.GitLoad.Execute()
	if !strings.Contains(vm.Source.Get(), "From the server") {
		t.Fatalf("load: %q (%s)", vm.Source.Get(), vm.GitNote.Get())
	}
	vm.GitPropose.Execute()
	if len(h.forge.proposed) != 1 || vm.GitLink.Get() == "" {
		t.Fatalf("propose: %v (%s)", h.forge.proposed, vm.GitNote.Get())
	}
	vm.GitRefresh.Execute()
	if vm.GitNote.Get() != T(KGitNoteUpToDate) {
		t.Fatalf("refresh: %s", vm.GitNote.Get())
	}
	// GitRemote is how a token reaches a forge: it was never called
	if len(h.tokens) != 0 || len(h.forges) != 0 {
		t.Fatalf("a forge remote was built in server mode: %v %v", h.tokens, h.forges)
	}
	for _, f := range gd.forges {
		if f != "github" {
			t.Errorf("remote asked for forge %q", f)
		}
	}
	// a token typed anyway is not taken
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	if len(h.tokens) != 0 {
		t.Fatal("a typed token was taken in server mode")
	}
}

func TestSignedOutShowsSignInAndNeverFallsBack(t *testing.T) {
	gd := &fakeGitd{err: gitd.ErrSignIn}
	_, vm, h := gitdScene(t, store.NewMemory(), gd, "github", "https://api.github.com", planted)
	if vm.GitMode.Get() != GitModeSignedOut || vm.GitNote.Get() != T(KGitNoteServerOut) || h.kept != "" {
		t.Fatalf("mode %d note %q kept %q", vm.GitMode.Get(), vm.GitNote.Get(), h.kept)
	}
	if !vm.GitServerSignIn.CanExecute() || vm.GitLoad.CanExecute() {
		t.Fatal("signed out: Sign in on, Load off")
	}
	vm.GitServerSignIn.Execute()
	if h.signIns != 1 {
		t.Fatal("Sign in did not go to the sign-in")
	}
	// a typed token does not turn into a client-mode connection
	vm.GitRepo.Set("o/r")
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	if vm.GitConnected.Get() || len(h.tokens) != 0 || vm.GitLoad.CanExecute() {
		t.Fatal("a token was used while signed out")
	}
}

func TestAnExpiredSessionEndsTheSignInAndNothingElseTakesOver(t *testing.T) {
	gd := signedIn("github:o/r")
	_, vm, h := gitdScene(t, store.NewMemory(), gd, "", "", "")
	vm.GitDir.Set("talk")
	vm.GitLoad.Execute()
	if vm.git.base == nil {
		t.Fatal(vm.GitNote.Get())
	}
	h.forge.fail = gitd.ErrSignIn // the IdP session ended
	vm.GitRefresh.Execute()
	if vm.GitMode.Get() != GitModeSignedOut || vm.GitNote.Get() != T(KGitNoteServerEnded) || gd.forgot == 0 {
		t.Fatalf("mode %d note %q forgot %d", vm.GitMode.Get(), vm.GitNote.Get(), gd.forgot)
	}
	if vm.GitUser.Get() != "" || vm.GitRepos.Get() != "" || vm.GitConnected.Get() || vm.GitLoad.CanExecute() || vm.GitPropose.CanExecute() {
		t.Fatal("what the signed-in person had stays in the pane")
	}
	if len(h.tokens) != 0 {
		t.Fatal("a forge remote was built after the sign-in ended")
	}
}

func TestAServerThatHasNoCredentialForTheUserSaysSo(t *testing.T) {
	gd := signedIn("github:o/r")
	_, vm, h := gitdScene(t, store.NewMemory(), gd, "", "", "")
	h.forge.fail = gitd.ErrNoCredential
	vm.GitLoad.Execute()
	if vm.GitNote.Get() != T(KGitNoteServerNoCred) || vm.GitMode.Get() != GitModeSignedIn {
		t.Fatalf("%d %q", vm.GitMode.Get(), vm.GitNote.Get())
	}
}

func TestARememberedRepositoryTheServerDoesNotGrantIsDropped(t *testing.T) {
	m := store.NewMemory()
	// the page's earlier visit (another person's) loaded o/secret
	_, vm, _ := gitdScene(t, m, nil, "", "", "")
	vm.GitRepo.Set("o/secret")
	vm.git.base = &remote.Base{Forge: "github", Repo: "o/secret", Ref: "main", Commit: strings.Repeat("a", 40)}
	vm.saveUI()
	_, vm2, _ := gitdScene(t, m, signedIn("github:o/mine"), "", "", "")
	if vm2.git.base != nil || vm2.GitRepo.Get() != "o/mine" {
		t.Fatalf("base %v repo %q", vm2.git.base, vm2.GitRepo.Get())
	}
	if strings.Contains(vm2.GitRepos.Get(), "secret") || strings.Contains(vm2.GitNote.Get(), "secret") {
		t.Fatal("the other person's repository is shown")
	}
}

func TestNoRepositoriesGranted(t *testing.T) {
	_, vm, _ := gitdScene(t, store.NewMemory(), signedIn(), "", "", "")
	if vm.GitNote.Get() != T(KGitNoteServerNoRepos) || vm.GitLoad.CanExecute() || vm.GitConnected.Get() {
		t.Fatalf("%q", vm.GitNote.Get())
	}
}

func TestServerPaneHasNoTokenField(t *testing.T) {
	s, vm, _ := gitdScene(t, store.NewMemory(), signedIn("github:o/r"), "", "", "")
	if s.gitp.box.Visible().Get() != "server" {
		t.Fatalf("page %q", s.gitp.box.Visible().Get())
	}
	vm.GitPane.Set(true)
	r := map[string]struct{}{}
	for k := range s.Rects() {
		r[k] = struct{}{}
	}
	for _, k := range []string{"git-token", "git-remember", "git-connect", "git-forge", "git-signin"} {
		if _, ok := r[k]; ok {
			t.Errorf("%s is shown in server mode", k)
		}
	}
	for _, k := range []string{"git-pick", "git-load", "git-propose"} {
		if _, ok := r[k]; !ok {
			t.Errorf("%s missing in server mode", k)
		}
	}
}
