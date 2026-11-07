package playground

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-browserhttp/browsergit/oauth"
	"github.com/go-slides/go-slides.github.io/playground/store"
)

// fakeGitLab is GitLab's OAuth side: codes bound to their challenge, single
// use; refresh tokens rotate; revocations are recorded.
type fakeGitLab struct {
	mu        sync.Mutex
	srv       *httptest.Server
	codes     map[string]string
	refreshes map[string]bool
	revoked   []string
	requests  int
	n         int
	expiresIn int
}

const redirectURI = "https://pg.example/playground/"

func newFakeGitLab(t *testing.T) *fakeGitLab {
	g := &fakeGitLab{codes: map[string]string{}, refreshes: map[string]bool{}, expiresIn: 7200}
	g.srv = httptest.NewServer(http.HandlerFunc(g.serve))
	t.Cleanup(g.srv.Close)
	return g
}

func (g *fakeGitLab) issue(authURL string) string {
	u, _ := url.Parse(authURL)
	q := u.Query()
	g.mu.Lock()
	defer g.mu.Unlock()
	g.n++
	code := fmt.Sprintf("GLCODE-%d", g.n)
	g.codes[code] = q.Get("code_challenge")
	return redirectURI + "?" + url.Values{"code": {code}, "state": {q.Get("state")}}.Encode()
}

func (g *fakeGitLab) serve(w http.ResponseWriter, r *http.Request) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.requests++
	_ = r.ParseForm()
	w.Header().Set("Content-Type", "application/json")
	fail := func() { w.WriteHeader(400); _, _ = w.Write([]byte(`{"error":"invalid_grant"}`)) }
	ok := func() {
		g.n++
		rt := fmt.Sprintf("GLREFRESH-%d", g.n)
		g.refreshes[rt] = true
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": fmt.Sprintf("GLACCESS-%d", g.n), "token_type": "Bearer", "expires_in": g.expiresIn, "refresh_token": rt})
	}
	switch r.URL.Path {
	case "/oauth/token":
		switch r.Form.Get("grant_type") {
		case "authorization_code":
			ch, found := g.codes[r.Form.Get("code")]
			delete(g.codes, r.Form.Get("code"))
			if !found || r.Form.Get("redirect_uri") != redirectURI || oauth.Challenge(r.Form.Get("code_verifier")) != ch {
				fail()
				return
			}
			ok()
		case "refresh_token":
			if !g.refreshes[r.Form.Get("refresh_token")] {
				fail()
				return
			}
			delete(g.refreshes, r.Form.Get("refresh_token"))
			ok()
		}
	case "/oauth/revoke":
		g.revoked = append(g.revoked, r.Form.Get("token"))
		_, _ = w.Write([]byte(`{}`))
	}
}

// oauthTab is a page with sign-in: its address, its sessionStorage and where
// it navigated, recorded in order.
type oauthTab struct {
	*gitTab
	gl      *fakeGitLab
	page    string
	session map[string]oauth.Attempt
	events  []string
}

func (o *oauthTab) OAuthConfig(name string) (oauth.Config, bool) {
	if name != "gitlab" || o.gl == nil {
		return oauth.Config{}, false
	}
	c := oauth.GitLab(o.gl.srv.URL, "app-1", redirectURI)
	return c, true
}
func (o *oauthTab) PageURL() string { return o.page }
func (o *oauthTab) CleanURL() {
	o.events = append(o.events, "clean")
	o.page = strings.SplitN(o.page, "?", 2)[0]
}
func (o *oauthTab) KeepAttempt(a oauth.Attempt) { o.session["a"] = a }
func (o *oauthTab) TakeAttempt() (oauth.Attempt, bool) {
	o.events = append(o.events, "take")
	a, ok := o.session["a"]
	delete(o.session, "a")
	return a, ok
}
func (o *oauthTab) Navigate(u string) { o.events = append(o.events, "navigate"); o.page = u }

// oauthPage opens a page at addr over m, with the session store and GitLab given.
func oauthPage(t *testing.T, m *store.Memory, gl *fakeGitLab, addr string, session map[string]oauth.Attempt) (*ViewModel, *oauthTab) {
	t.Helper()
	SetupText(1)
	g := &gitTab{
		promptTab: &promptTab{filesTab: &filesTab{tab: &tab{fakeHost: newHost(), answer: true}, saved: map[string][]byte{}}},
		forge:     &fakeForge{files: map[string]string{"talk/deck.md": "# From GitLab\n"}, branches: map[string]bool{}},
		remoteOK:  true,
	}
	g.st = store.New(m, NewTabID(), load(m, DraftProject), load(m, PreviousProject), load(m, UIProject))
	h := &oauthTab{gitTab: g, gl: gl, page: addr, session: session}
	tr := &fakeTransport{}
	vm := NewViewModel(tr, h, "")
	tr.vm = vm
	answer(tr, vm)
	return vm, h
}

func TestSignInWithGitLab(t *testing.T) {
	m, gl := store.NewMemory(), newFakeGitLab(t)
	session := map[string]oauth.Attempt{}
	vm, h := oauthPage(t, m, gl, redirectURI, session)
	if vm.GitSignIn.CanExecute() {
		t.Fatal("Sign in is GitLab's: not offered for GitHub")
	}
	vm.GitForge.Set(1)
	if !vm.GitSignIn.CanExecute() || vm.GitNote.Get() != T(KGitNoteOffGitLab) {
		t.Fatalf("GitLab chosen: sign-in offered, note %q", vm.GitNote.Get())
	}
	vm.GitSignIn.Execute()
	a, kept := session["a"]
	if !kept || h.events[len(h.events)-1] != "navigate" || !strings.HasPrefix(h.page, gl.srv.URL+"/oauth/authorize?") {
		t.Fatalf("sign-in: kept %v, page %q", kept, h.page)
	}
	if strings.Contains(h.page, a.Verifier) || !strings.Contains(h.page, oauth.Challenge(a.Verifier)) {
		t.Fatal("the authorize URL carries the challenge, never the verifier")
	}

	// GitLab sends the page back; it loads again, in the same tab.
	back := gl.issue(h.page)
	vm2, h2 := oauthPage(t, m, gl, back, session)
	if len(h2.events) < 2 || h2.events[0] != "clean" || h2.events[1] != "take" {
		t.Fatalf("the URL must be cleaned first, then the attempt taken: %v", h2.events)
	}
	if strings.Contains(h2.page, "code=") || len(session) != 0 {
		t.Fatal("the code stays in the URL, or the attempt in the session")
	}
	if !vm2.GitConnected.Get() || vm2.git.oauth == nil || vm2.git.forge != "gitlab" || vm2.GitForge.Get() != 1 || vm2.GitNote.Get() != T(KGitNoteSignedIn) {
		t.Fatalf("signed in: %v %q", vm2.GitConnected.Get(), vm2.GitNote.Get())
	}
	if last := h2.forges[len(h2.forges)-1]; last != "gitlab" || !strings.HasPrefix(h2.tokens[len(h2.tokens)-1], "GLACCESS-") {
		t.Fatalf("the remote is GitLab's, with the access token: %v", h2.forges)
	}
	// Remember is for a typed token only: a sign-in is never kept.
	vm2.GitRemember.Set(true)
	if h2.kept != "" {
		t.Fatal("a sign-in's token was kept on the device")
	}

	// The same callback again (back button, a copied URL): nothing is exchanged.
	before := gl.requests
	vm3, _ := oauthPage(t, m, gl, back, session)
	if vm3.GitConnected.Get() || gl.requests != before || vm3.GitNote.Get() != T(KGitNoteSignInRefused) {
		t.Fatalf("a replayed callback: connected %v, %d requests, note %q", vm3.GitConnected.Get(), gl.requests-before, vm3.GitNote.Get())
	}

	// Work, then sign out: both tokens revoked, nothing left anywhere.
	vm2.GitRepo.Set("g/p")
	vm2.GitDir.Set("talk")
	vm2.GitLoad.Execute()
	if !strings.Contains(vm2.Source.Get(), "From GitLab") {
		t.Fatalf("load through the sign-in: %q", vm2.GitNote.Get())
	}
	acc, ref := vm2.git.oauth.Access.Reveal(), vm2.git.oauth.Refresh.Reveal()
	tok := vm2.git.oauth
	vm2.GitDisconnect.Execute()
	if vm2.git.oauth != nil || vm2.GitConnected.Get() || !tok.Access.IsZero() || !tok.Refresh.IsZero() {
		t.Fatal("Disconnect must forget the sign-in")
	}
	if len(gl.revoked) != 2 || gl.revoked[0] != ref || gl.revoked[1] != acc {
		t.Fatalf("revoked %v", gl.revoked)
	}
	all := everything(t, m, h2.gitTab, vm2, nil, nil)
	for _, s := range []string{acc, ref, a.Verifier, a.State} {
		if strings.Contains(all, s) {
			t.Fatalf("%q reached storage or the screen", s)
		}
	}
}

func TestACallbackThisTabDidNotStartExchangesNothing(t *testing.T) {
	m, gl := store.NewMemory(), newFakeGitLab(t)
	// Another tab (or an attacker's link) began the sign-in: this tab has no attempt.
	other := map[string]oauth.Attempt{}
	vm, h := oauthPage(t, m, gl, redirectURI, other)
	vm.GitForge.Set(1)
	vm.GitSignIn.Execute()
	back := gl.issue(h.page)
	mine := map[string]oauth.Attempt{}
	vm2, _ := oauthPage(t, m, gl, back, mine)
	if vm2.GitConnected.Get() || gl.requests != 0 {
		t.Fatalf("connected %v after %d requests", vm2.GitConnected.Get(), gl.requests)
	}
	// A forged state, with this tab's attempt: refused by the state check.
	vm3, h3 := oauthPage(t, m, gl, redirectURI, mine)
	vm3.GitForge.Set(1)
	vm3.GitSignIn.Execute()
	forged := gl.issue(h3.page)
	forged = strings.Replace(forged, "state=", "state=forged", 1)
	vm4, _ := oauthPage(t, m, gl, forged, mine)
	if vm4.GitConnected.Get() || gl.requests != 0 || len(mine) != 0 {
		t.Fatalf("a forged state: connected %v, %d requests, attempt left %v", vm4.GitConnected.Get(), gl.requests, mine)
	}
	// A callback on another path is not a callback at all.
	vm5, h5 := oauthPage(t, m, gl, "https://pg.example/other/?code=x&state=y", mine)
	if vm5.GitConnected.Get() || len(h5.events) != 0 {
		t.Fatal("another path must be ignored")
	}
}

func TestAnExpiringSignInRefreshesAndAFailedRefreshSignsOut(t *testing.T) {
	m, gl := store.NewMemory(), newFakeGitLab(t)
	gl.expiresIn = 30 // under the minute of margin: refresh before each use
	session := map[string]oauth.Attempt{}
	vm, h := oauthPage(t, m, gl, redirectURI, session)
	vm.GitForge.Set(1)
	vm.GitSignIn.Execute()
	vm2, h2 := oauthPage(t, m, gl, gl.issue(h.page), session)
	first := vm2.git.oauth.Refresh.Reveal()
	vm2.GitRepo.Set("g/p")
	vm2.GitDir.Set("talk")
	vm2.GitLoad.Execute()
	if vm2.git.oauth.Refresh.Reveal() == first || !strings.Contains(vm2.Source.Get(), "From GitLab") {
		t.Fatal("an expiring access token must be refreshed (and rotated) before use")
	}
	if got := h2.tokens[len(h2.tokens)-1]; !strings.HasPrefix(got, "GLACCESS-") || got == h2.tokens[0] {
		t.Fatal("the remote must use the refreshed access token")
	}
	// GitLab forgets the refresh token: the next use signs out.
	gl.mu.Lock()
	gl.refreshes = map[string]bool{}
	gl.mu.Unlock()
	vm2.GitLoad.Execute()
	if vm2.GitConnected.Get() || vm2.git.oauth != nil {
		t.Fatal("a failed refresh must sign out")
	}
	_ = time.Now
}

func TestNoAppMeansNoSignIn(t *testing.T) {
	m := store.NewMemory()
	vm, _ := oauthPage(t, m, nil, redirectURI+"?code=x&state=y", map[string]oauth.Attempt{})
	vm.GitForge.Set(1)
	if vm.GitSignIn.CanExecute() {
		t.Fatal("no application registered: no sign-in")
	}
	vm.signIn()
	if vm.GitNote.Get() != T(KGitNoteNoApp) {
		t.Fatalf("note %q", vm.GitNote.Get())
	}
	// A typed GitLab token still works, for GitLab.
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	h := vm.host.(*oauthTab)
	if !vm.GitConnected.Get() || h.forges[len(h.forges)-1] != "gitlab" {
		t.Fatal("a GitLab personal token connects to GitLab")
	}
	// Changing forge disconnects.
	vm.GitForge.Set(0)
	if vm.GitConnected.Get() || vm.GitNote.Get() != T(KGitNoteOff) {
		t.Fatal("another forge must disconnect")
	}
}
