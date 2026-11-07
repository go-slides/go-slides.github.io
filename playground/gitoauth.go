package playground

import (
	"context"
	"errors"

	"github.com/go-browserhttp/browsergit/oauth"
	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-widgets/mvvm"
)

// Sign in with GitLab: the authorization code flow with PKCE for a page that
// has no secret (go-browserhttp/browsergit/oauth; RFC 7636, RFC 9700).
//
//   - Sign in makes a fresh state and verifier, keeps them for THIS tab only
//     (sessionStorage, through the host) and sends the page to GitLab.
//   - Back on the page, the address is cleaned at once (the code leaves the
//     URL bar and the history before anything else runs), the attempt is read
//     AND deleted, and only then is the code exchanged: a callback this tab did
//     not start, a replayed one, or one on another URL exchanges nothing.
//   - The tokens live in memory (oauth.Tokens, secret.Token): never stored,
//     not even with Remember, which is for a typed personal token only. The
//     access token is refreshed (and rotated) before use when it is about to
//     expire; a failed refresh signs out.
//   - Disconnect revokes both tokens at GitLab and forgets them here at once.

// GitForges are the forges the pane offers, in GitForge order.
var GitForges = []string{"GitHub", "GitLab.com"}

// gitForgeNames are their remote names.
var gitForgeNames = []string{"github", "gitlab"}

// OAuthHost is what the page does for a sign-in.
type OAuthHost interface {
	// OAuthConfig is the OAuth application for forge; false when the site
	// has none registered (then only a personal token works).
	OAuthConfig(forge string) (oauth.Config, bool)
	// PageURL is the page's address now.
	PageURL() string
	// CleanURL drops the query from the address, without a history entry.
	CleanURL()
	// KeepAttempt keeps a's state and verifier for this tab only.
	KeepAttempt(a oauth.Attempt)
	// TakeAttempt returns the kept attempt and deletes it.
	TakeAttempt() (oauth.Attempt, bool)
	// Navigate sends the page to url.
	Navigate(url string)
}

// offNote is what the pane says while not connected, for the forge chosen.
func (vm *ViewModel) offNote() string {
	if vm.forgeName() == "gitlab" {
		return T(KGitNoteOffGitLab)
	}
	return T(KGitNoteOff)
}

func (vm *ViewModel) forgeName() string {
	if i := vm.GitForge.Get(); i >= 0 && i < len(gitForgeNames) {
		return gitForgeNames[i]
	}
	return gitForgeNames[0]
}

// initOAuth makes Sign in and finishes a sign-in the page came back with.
func (vm *ViewModel) initOAuth(idle func() bool) {
	oh, ok := vm.host.(OAuthHost)
	vm.GitSignIn = mvvm.NewCommand(vm.signIn, func() bool {
		if !ok || !idle() || vm.forgeName() != "gitlab" {
			return false
		}
		_, has := oh.OAuthConfig("gitlab")
		return has
	})
	if !ok {
		return
	}
	cfg, has := oh.OAuthConfig("gitlab")
	if !has {
		return
	}
	vm.git.oauthCfg = cfg
	page := oh.PageURL()
	if !cfg.IsCallback(page) {
		return
	}
	oh.CleanURL() // first of all: the code leaves the address bar and the history
	a, kept := oh.TakeAttempt()
	vm.quiet = true
	vm.GitForge.Set(1)
	vm.quiet = false
	vm.GitPane.Set(true)
	if !kept {
		vm.gitSay(T(KGitNoteSignInRefused))
		return
	}
	var tok *oauth.Tokens
	vm.gitRun(func(ctx context.Context) (err error) {
		tok, err = cfg.Finish(ctx, a, page)
		return err
	}, func(error) { vm.useOAuth(tok) })
}

// signIn sends the page to GitLab with a fresh attempt.
func (vm *ViewModel) signIn() {
	oh, ok := vm.host.(OAuthHost)
	if !ok {
		return
	}
	cfg, has := oh.OAuthConfig("gitlab")
	if !has {
		vm.gitSay(T(KGitNoteNoApp))
		return
	}
	a, link, err := cfg.Begin()
	if err != nil {
		vm.gitSay(T(KGitNoteFailed, ErrText(err)))
		return
	}
	oh.KeepAttempt(a)
	vm.GitNote.Set(T(KGitNoteRedirecting))
	oh.Navigate(link)
}

// useOAuth connects the pane with a signed-in session.
func (vm *ViewModel) useOAuth(tok *oauth.Tokens) {
	h, ok := vm.host.(GitHost)
	if !ok || tok == nil {
		return
	}
	rem, err := h.GitRemote("gitlab", tok.Access)
	if err != nil {
		vm.git.oauth = tok
		vm.gitSay(T(KGitNoteFailed, ErrText(err)))
		vm.signOut()
		return
	}
	vm.git.token.Forget()
	vm.git.oauth, vm.git.rem, vm.git.forge = tok, rem, "gitlab"
	vm.GitConnected.Set(true)
	vm.GitNote.Set(T(KGitNoteSignedIn))
	vm.raiseGitCommands()
}

// remoteForOp is what an operation calls, away from the UI's turn, to get the
// remote: as it is for a typed token; for a sign-in, refreshed first when the
// access token is about to expire.
func (vm *ViewModel) remoteForOp() func(context.Context) (remote.Remote, error) {
	rem, tok, cfg := vm.git.rem, vm.git.oauth, vm.git.oauthCfg
	h, _ := vm.host.(GitHost)
	if tok == nil {
		return func(context.Context) (remote.Remote, error) { return rem, nil }
	}
	return func(ctx context.Context) (remote.Remote, error) {
		if !cfg.Expired(tok) {
			return rem, nil
		}
		if err := cfg.Refresh(ctx, tok); err != nil {
			return nil, err // tok is forgotten: signed out
		}
		if h == nil {
			return nil, errors.New("no git host")
		}
		return h.GitRemote("gitlab", tok.Access)
	}
}

// afterOp notices a sign-in whose refresh failed (its tokens are gone) and
// shows the pane signed out.
func (vm *ViewModel) afterOp() {
	if t := vm.git.oauth; t != nil && t.Access.IsZero() {
		vm.git.oauth, vm.git.rem = nil, nil
		vm.GitConnected.Set(false)
		vm.raiseGitCommands()
	}
}

// signOut revokes a sign-in at GitLab (in the background) and forgets it here
// at once.
func (vm *ViewModel) signOut() {
	t := vm.git.oauth
	if t == nil {
		return
	}
	vm.git.oauth = nil
	cfg := vm.git.oauthCfg
	h, ok := vm.host.(GitHost)
	if !ok {
		t.Forget()
		return
	}
	h.GitRun(func() { _ = cfg.Revoke(context.Background(), t) }, func() {})
}
