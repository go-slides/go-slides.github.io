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
	hang     bool // Session answers only when its context ends
}

func (g *fakeGitd) Session(ctx context.Context) (gitd.Info, error) {
	g.sessions++
	if g.hang {
		<-ctx.Done()
		return gitd.Info{}, remote.ErrTransport
	}
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
		kept: kept, remoteOK: true, gd: gd, front: frontMarker, hold: holdStorage, park: parkRuns,
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
	if s.gitp.Visible().Get() != "server" {
		t.Fatalf("page %q", s.gitp.Visible().Get())
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

const alicePrivate = "---\ntheme: default\n---\n\n# ALICE PRIVATE ALPHA\n"

// aliceLoadsAndSaves is Alice, signed in, loading her private deck on a profile.
func aliceLoadsAndSaves(t *testing.T, m *store.Memory) (*ViewModel, *gitTab) {
	t.Helper()
	alice := signedIn("github:lab/alpha")
	alice.info.User = "sub-alice"
	_, vm, h := gitdScene(t, m, alice, "", "", "")
	h.forge.files = map[string]string{"talk/deck.md": alicePrivate}
	vm.GitDir.Set("talk")
	vm.GitLoad.Execute()
	vm.SaveDraft()
	if !strings.Contains(vm.Source.Get(), "ALICE PRIVATE") {
		t.Fatalf("control: alice's deck is not loaded: %s", vm.GitNote.Get())
	}
	return vm, h
}

// everythingStored is every project and version the store holds, as one string.
func everythingStored(m *store.Memory) string {
	var b strings.Builder
	for _, n := range m.Names() {
		if p := load(m, n); p != nil {
			for _, k := range p.Paths() {
				d, _ := p.Read(k)
				b.Write(d)
			}
			for k, v := range p.UI {
				b.WriteString(k + "=" + v + "\n")
			}
		}
	}
	return b.String()
}

// go-slides/go-slides.github.io#99: the next person on a shared browser profile
// must not open on the previous person's private deck.
func TestTheNextPersonDoesNotGetThePreviousPersonsDeck(t *testing.T) {
	m := store.NewMemory()
	aliceLoadsAndSaves(t, m)
	if !strings.Contains(everythingStored(m), "ALICE PRIVATE") {
		t.Fatal("control: alice's deck is not stored")
	}
	bob := signedIn("github:lab/beta")
	bob.info.User = "sub-bob"
	_, vm2, _ := gitdScene(t, m, bob, "", "", "")
	if vm2.GitMode.Get() != GitModeSignedIn || vm2.GitUser.Get() != "sub-bob" {
		t.Fatalf("control: bob is not signed in: %d", vm2.GitMode.Get())
	}
	if strings.Contains(vm2.Source.Get(), "ALICE") {
		t.Fatalf("bob's editor opens on alice's deck: %q", vm2.Source.Get())
	}
	if s := everythingStored(m); strings.Contains(s, "ALICE") {
		t.Fatalf("alice's deck is still stored for bob's page")
	}
	if vm2.GitRepo.Get() != "lab/beta" || vm2.git.base != nil {
		t.Fatalf("bob's pane: repo %q base %v", vm2.GitRepo.Get(), vm2.git.base)
	}
	if n := len(vm2.vers.shown); n != 0 {
		t.Fatalf("bob sees %d versions", n)
	}
	if vm2.CanUndo() {
		t.Fatal("bob can undo into alice's deck")
	}
}

// The explorer's status keeps a hash per file of the project, by the file's
// buffer (gitrepo.go): on a change of identity the previous person's buffers must
// not stay in memory until some later refresh prunes them.
func TestAChangeOfIdentityEmptiesTheStatusCache(t *testing.T) {
	cached := func(vm *ViewModel) int {
		vm.repo.mu.Lock()
		defer vm.repo.mu.Unlock()
		return len(vm.repo.cache)
	}
	// the same page, another person (wipeLocal)
	vm, _ := aliceLoadsAndSaves(t, store.NewMemory())
	if cached(vm) == 0 {
		t.Fatal("control: alice's files are hashed and kept")
	}
	vm.bindOwner("sub-bob")
	if n := cached(vm); n != 0 {
		t.Fatalf("a new identity: %d hashes of alice's buffers are still kept", n)
	}
	if s := vm.repo.snap.Load(); s == nil || s.base != nil || len(s.work) != 0 {
		t.Fatalf("a new identity: the snapshot still holds %+v", s)
	}
	// another repository of the same person (pickRepo)
	vm2, _ := aliceLoadsAndSaves(t, store.NewMemory())
	if cached(vm2) == 0 {
		t.Fatal("control: the files are hashed and kept")
	}
	vm2.git.grants = append(vm2.git.grants, gitd.Repo{Forge: "github", Name: "lab/other"})
	vm2.pickRepo(len(vm2.git.grants) - 1)
	if n := cached(vm2); n != 0 {
		t.Fatalf("another repository: %d hashes of the last one's buffers are still kept", n)
	}
}

func TestTheSamePersonKeepsTheirDeck(t *testing.T) {
	m := store.NewMemory()
	aliceLoadsAndSaves(t, m)
	again := signedIn("github:lab/alpha")
	again.info.User = "sub-alice"
	_, vm2, _ := gitdScene(t, m, again, "", "", "")
	if !strings.Contains(vm2.Source.Get(), "ALICE PRIVATE") {
		t.Fatalf("alice lost her own deck on returning: %q", vm2.Source.Get())
	}
}

func TestSigningOutOrAnExpiredSessionClearsThePersonsDeck(t *testing.T) {
	m := store.NewMemory()
	vm, h := aliceLoadsAndSaves(t, m)
	h.forge.fail = gitd.ErrSignIn
	vm.GitRefresh.Execute()
	if vm.GitMode.Get() != GitModeSignedOut {
		t.Fatalf("mode %d", vm.GitMode.Get())
	}
	if strings.Contains(vm.Source.Get(), "ALICE") || strings.Contains(everythingStored(m), "ALICE") || vm.CanUndo() || len(vm.vers.shown) != 0 {
		t.Fatalf("alice's deck stays after the sign-in ended: %q", vm.Source.Get())
	}
	// and a page opened signed out on that profile shows nothing of hers either
	_, vm3, _ := gitdScene(t, m, &fakeGitd{err: gitd.ErrSignIn}, "", "", "")
	if strings.Contains(vm3.Source.Get(), "ALICE") {
		t.Fatal("a signed-out page shows alice's deck")
	}
}

func TestADraftNobodyOwnsSurvivesASignedOutVisit(t *testing.T) {
	m := store.NewMemory()
	_, vm, _ := gitdScene(t, m, nil, "", "", "") // client mode: a visitor's own draft
	vm.Source.Set("# visitor's own notes\n")
	vm.SaveDraft()
	_, vm2, _ := gitdScene(t, m, &fakeGitd{err: gitd.ErrSignIn}, "", "", "")
	if !strings.Contains(vm2.Source.Get(), "visitor's own notes") {
		t.Fatalf("a signed-out visitor lost a draft that belongs to nobody: %q", vm2.Source.Get())
	}
}

func (g *gitTab) GitdFront(context.Context) bool { return g.front }

// frontMarker is what gitdScene gives the page's marker.
var frontMarker bool

// holdStorage and parkRuns are what gitdScene gives the page (#107).
var (
	holdStorage bool
	parkRuns    *[][2]func()
)

func (g *gitTab) HoldStorage() bool { return g.hold }

// go-slides/go-slides.github.io#104: on a page served by a slides-gitd front, a
// server that does not answer is "unavailable", never "no server".
func TestAFrontThatIsDownDoesNotBringBackTheTokenField(t *testing.T) {
	for name, err := range map[string]error{"a 502": remote.ErrTransport, "a page in place of the API": gitd.ErrNoServer} {
		gd := &fakeGitd{err: err}
		_, vm, ht := gitdSceneFront(t, store.NewMemory(), gd, true, "github", "https://api.github.com", planted)
		if vm.GitMode.Get() != GitModeUnavailable || vm.GitNote.Get() != T(KGitNoteServerDown) {
			t.Fatalf("%s: mode %d note %q", name, vm.GitMode.Get(), vm.GitNote.Get())
		}
		if ht.kept != "" {
			t.Errorf("%s: a remembered token stays", name)
		}
		vm.GitRepo.Set("o/r")
		vm.GitToken.Set(planted)
		vm.GitConnect.Execute()
		if vm.GitConnected.Get() || len(ht.tokens) != 0 || vm.GitLoad.CanExecute() || vm.GitServerSignIn.CanExecute() {
			t.Fatalf("%s: a token or a command works on a page whose server is down", name)
		}
	}
}

func TestWithoutTheMarkerAFailingServerIsNoServer(t *testing.T) {
	_, vm, ht := gitdSceneFront(t, store.NewMemory(), &fakeGitd{err: remote.ErrTransport}, false, "", "", "")
	if vm.GitMode.Get() != GitModeClient {
		t.Fatalf("mode %d", vm.GitMode.Get())
	}
	vm.GitRepo.Set("o/r")
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	if !vm.GitConnected.Get() || len(ht.tokens) != 1 {
		t.Fatal("client mode no longer works without the marker")
	}
}

// gitdSceneFront is gitdScene with the marker set or not.
func gitdSceneFront(t *testing.T, m *store.Memory, gd *fakeGitd, front bool, keptForge, keptOrigin, kept string) (*Scene, *ViewModel, *gitTab) {
	t.Helper()
	frontMarker = front
	defer func() { frontMarker = false }()
	return gitdScene(t, m, gd, keptForge, keptOrigin, kept)
}

// bobsPage opens a page on m (alice's profile) that holds the stored project; the
// network answers when the test says so.
func bobsPage(t *testing.T, m *store.Memory, gd *fakeGitd) (*ViewModel, *gitTab, *[][2]func()) {
	t.Helper()
	var parked [][2]func()
	holdStorage, parkRuns, frontMarker = true, &parked, true
	defer func() { holdStorage, parkRuns, frontMarker = false, nil, false }()
	_, vm, h := gitdScene(t, m, gd, "", "", "")
	return vm, h, &parked
}

func answer1(p *[][2]func()) {
	for _, r := range *p {
		r[0]()
		r[1]()
	}
	*p = nil
}

// go-slides/go-slides.github.io#107: before the server says who is signed in, a
// page served by a front shows and renders nothing of what is stored.
func TestNothingStoredIsShownOrRenderedBeforeTheSessionAnswers(t *testing.T) {
	m := store.NewMemory()
	aliceLoadsAndSaves(t, m)
	bob := signedIn("github:lab/beta")
	bob.info.User = "sub-bob"
	vm, _, parked := bobsPage(t, m, bob)
	if len(*parked) != 1 {
		t.Fatalf("control: the identity check should be pending, parked %d", len(*parked))
	}
	if strings.Contains(vm.Source.Get(), "ALICE") || strings.Contains(vm.Doc.Get(), "ALICE") || strings.Contains(vm.RailDoc.Get(), "ALICE") {
		t.Fatalf("alice's deck is in bob's editor or sent to the renderer while the session is pending")
	}
	if vm.Status.Get() != T(KGitNoteServerHold) {
		t.Fatalf("status %q", vm.Status.Get())
	}
	answer1(parked)
	if vm.GitMode.Get() != GitModeSignedIn || strings.Contains(vm.Source.Get(), "ALICE") || strings.Contains(everythingStored(m), "ALICE") {
		t.Fatalf("after bob's session: mode %d source %q", vm.GitMode.Get(), vm.Source.Get())
	}
}

func TestTheSamePersonGetsTheirStoredDeckBackAfterTheSession(t *testing.T) {
	m := store.NewMemory()
	aliceLoadsAndSaves(t, m)
	again := signedIn("github:lab/alpha")
	again.info.User = "sub-alice"
	vm, _, parked := bobsPage(t, m, again)
	if strings.Contains(vm.Source.Get(), "ALICE") {
		t.Fatal("the stored deck shows before the session answers")
	}
	answer1(parked)
	if !strings.Contains(vm.Source.Get(), "ALICE PRIVATE") || vm.GitRepo.Get() != "lab/alpha" {
		t.Fatalf("alice did not get her deck back: %q repo %q", vm.Source.Get(), vm.GitRepo.Get())
	}
}

// a server that hangs: past the deadline the page says unavailable and still
// shows nothing stored, and writes nothing over it
func TestAHungSessionShowsUnavailableAndNothingStored(t *testing.T) {
	m := store.NewMemory()
	aliceLoadsAndSaves(t, m)
	old := GitdDeadline
	GitdDeadline = 30 * time.Millisecond
	defer func() { GitdDeadline = old }()
	vm, _, parked := bobsPage(t, m, &fakeGitd{hang: true})
	answer1(parked)
	if vm.GitMode.Get() != GitModeUnavailable || vm.GitNote.Get() != T(KGitNoteServerDown) {
		t.Fatalf("mode %d note %q", vm.GitMode.Get(), vm.GitNote.Get())
	}
	if strings.Contains(vm.Source.Get(), "ALICE") {
		t.Fatal("a hung server shows the stored deck")
	}
	vm.Source.Set("# typed while unavailable\n")
	vm.SaveDraft()
	if !strings.Contains(everythingStored(m), "ALICE PRIVATE") || strings.Contains(everythingStored(m), "typed while unavailable") {
		t.Fatal("the held project was written over or wiped while the server was unknown")
	}
}

func TestASignedOutPageWipesWhatIsBoundToSomeoneAndShowsNothing(t *testing.T) {
	m := store.NewMemory()
	aliceLoadsAndSaves(t, m)
	vm, _, parked := bobsPage(t, m, &fakeGitd{err: gitd.ErrSignIn})
	if strings.Contains(vm.Source.Get(), "ALICE") {
		t.Fatal("shown while pending")
	}
	answer1(parked)
	if vm.GitMode.Get() != GitModeSignedOut || strings.Contains(vm.Source.Get(), "ALICE") || strings.Contains(everythingStored(m), "ALICE") {
		t.Fatalf("mode %d", vm.GitMode.Get())
	}
}

func TestAnOrdinaryPageWhoseMarkerWasInconclusiveGetsItsProjectBack(t *testing.T) {
	m := store.NewMemory()
	_, vm0, _ := gitdScene(t, m, nil, "", "", "")
	vm0.Source.Set("# my own notes\n")
	vm0.SaveDraft()
	var parked [][2]func()
	holdStorage, parkRuns = true, &parked
	defer func() { holdStorage, parkRuns = false, nil }()
	_, vm, _ := gitdScene(t, m, &fakeGitd{err: gitd.ErrNoServer}, "", "", "")
	if strings.Contains(vm.Source.Get(), "my own notes") {
		t.Fatal("held page shows the stored deck")
	}
	answer1(&parked)
	if !strings.Contains(vm.Source.Get(), "my own notes") || vm.GitMode.Get() != GitModeClient {
		t.Fatalf("mode %d source %q", vm.GitMode.Get(), vm.Source.Get())
	}
}
