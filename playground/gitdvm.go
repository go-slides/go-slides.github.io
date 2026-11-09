package playground

import (
	"context"
	"errors"
	"strings"

	"github.com/go-browserhttp/browsergit/gitd"
	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-browserhttp/browsergit/secret"
	"github.com/go-widgets/mvvm"
)

// Server mode: the page is served behind slides-gitd (github.com/go-slides/gitd),
// a backend that keeps the forge credentials, so the page holds none.
//
//   - The page finds the server by asking its OWN origin (gitd.Client.Session,
//     GET /api/git/session). Anything that is not slides-gitd (a static host's
//     404, a page that answers every path) leaves the pane exactly as it was.
//   - A 401 shows Sign in; a session shows who is signed in and the repositories
//     the server grants, in a list: there is no token field, no OAuth button, no
//     Remember box in this mode, and a token a client-mode visit remembered is
//     forgotten from this device as soon as the server is found.
//   - Every action goes to the server (gitd.Client.Remote), which holds the
//     credential, names the proposal's branch and checks the grant.
//   - When the sign-in ends (401, on a call or on a refresh of the session) the
//     pane says so and offers Sign in. It never reaches for another way of
//     authenticating: no token typed earlier is used instead, because none is
//     held.
//   - The session cookie is the browser's (HttpOnly); the CSRF token lives in
//     gitd.Client, in memory. Neither is in the view-model, the UI state, a
//     message, a version or the journal.

// The modes the pane is in.
const (
	GitModeClient    = iota // no server: a token in the tab, or a GitLab sign-in
	GitModeSignedOut        // a server, and the person must sign in
	GitModeSignedIn         // a server, and a session
)

// GitdClient is what the view-model asks of slides-gitd: browsergit's gitd.Client.
type GitdClient interface {
	Session(ctx context.Context) (gitd.Info, error)
	Remote(forge string) remote.Remote
	Forget()
}

// GitdHost is a host that may be served behind slides-gitd.
type GitdHost interface {
	// Gitd is the client for the page's own origin; nil when there can be none.
	Gitd() GitdClient
	// GitdSignIn sends the page to the sign-in (oauth2-proxy) and back.
	GitdSignIn()
}

// initGitd makes the server mode's observables and asks the server who we are.
func (vm *ViewModel) initGitd() {
	vm.GitMode = mvvm.NewObservable(GitModeClient)
	vm.GitUser = mvvm.NewObservable("")
	vm.GitRepos = mvvm.NewObservable("")
	vm.GitRepoPick = mvvm.NewObservable(0)
	gh, ok := vm.host.(GitdHost)
	vm.GitServerSignIn = mvvm.NewCommand(func() {
		if ok {
			vm.GitNote.Set(T(KGitNoteServerSignIn))
			gh.GitdSignIn()
		}
	}, func() bool { return ok && vm.GitMode.Get() == GitModeSignedOut && !vm.git.busy })
	vm.GitRepoPick.Subscribe(func(i int) {
		if vm.GitMode.Get() == GitModeSignedIn {
			vm.pickRepo(i)
		}
	})
}

// discoverGitd asks the server, when the host may have one.
func (vm *ViewModel) discoverGitd() {
	if gh, ok := vm.host.(GitdHost); ok {
		if c := gh.Gitd(); c != nil {
			vm.git.gitd = c
			vm.sessionOp(true)
		}
	}
}

// sessionOp asks the server for the session. first: this is the discovery, and
// a server that is not there is no news.
func (vm *ViewModel) sessionOp(first bool) {
	h, ok := vm.host.(GitHost)
	if !ok {
		return
	}
	var info gitd.Info
	var err error
	vm.git.busy = true
	vm.raiseGitCommands()
	h.GitRun(func() { info, err = vm.git.gitd.Session(context.Background()) }, func() {
		vm.git.busy = false
		vm.takeSession(info, err, first)
		vm.raiseGitCommands()
		vm.GitServerSignIn.RaiseCanExecuteChanged()
	})
}

// forgetClientCredentials drops what client mode held: the typed token, a
// GitLab sign-in, and a token remembered on this device.
func (vm *ViewModel) forgetClientCredentials() {
	vm.signOut()
	vm.git.token.Forget()
	vm.git.token, vm.git.rem = secret.Token{}, nil
	if k, ok := vm.host.(TokenKeeper); ok {
		k.ForgetToken()
	}
	vm.GitToken.Set("")
}

func (vm *ViewModel) takeSession(info gitd.Info, err error, first bool) {
	switch {
	case err == nil:
		vm.enterServerMode()
		vm.GitMode.Set(GitModeSignedIn)
		vm.GitUser.Set(info.User)
		vm.git.grants = info.Repos
		labels := make([]string, len(info.Repos))
		for i, r := range info.Repos {
			labels[i] = r.String()
		}
		vm.GitRepos.Set(strings.Join(labels, "\n"))
		vm.GitConnected.Set(len(labels) > 0)
		// a repository the server does not grant is dropped from what was
		// remembered: another person may have used this browser
		pick := -1
		for i, r := range info.Repos {
			if r.Name == vm.GitRepo.Get() {
				pick = i
			}
		}
		if pick < 0 {
			vm.git.base, vm.git.theirs, vm.git.theirsBase = nil, nil, nil
			vm.GitRepo.Set("")
			vm.GitLink.Set("")
			pick = 0
		}
		if len(labels) == 0 {
			vm.git.rem = nil
			vm.GitNote.Set(T(KGitNoteServerNoRepos))
			return
		}
		vm.GitRepoPick.Set(pick)
		vm.pickRepo(pick)
		vm.GitNote.Set(T(KGitNoteServer, info.User))
	case errors.Is(err, gitd.ErrSignIn):
		vm.enterServerMode()
		vm.dropSession()
		if first {
			vm.GitNote.Set(T(KGitNoteServerOut))
		} else {
			vm.GitNote.Set(T(KGitNoteServerEnded))
		}
	case first:
		// no slides-gitd here: the pane is what it was
	default:
		vm.gitSay(T(KGitNoteServerFailed, ErrText(err)))
	}
}

// enterServerMode forgets what client mode held (no token is used in this mode,
// so none is kept) and leaves the signed-in state's fields to the caller.
func (vm *ViewModel) enterServerMode() {
	if vm.GitMode.Get() == GitModeClient {
		vm.forgetClientCredentials()
	}
}

// dropSession is the signed-out state: nothing the signed-in person had stays
// in the pane, and no command that needs a session can run.
func (vm *ViewModel) dropSession() {
	vm.GitMode.Set(GitModeSignedOut)
	vm.GitUser.Set("")
	vm.GitRepos.Set("")
	vm.git.grants, vm.git.rem = nil, nil
	vm.git.theirs, vm.git.theirsBase = nil, nil
	vm.GitConnected.Set(false)
	vm.git.gitd.Forget()
}

// pickRepo makes the pane talk about grant i.
func (vm *ViewModel) pickRepo(i int) {
	if i < 0 || i >= len(vm.git.grants) || vm.git.gitd == nil {
		return
	}
	r := vm.git.grants[i]
	if b := vm.git.base; b != nil && b.Repo != r.Name {
		vm.git.base, vm.git.theirs, vm.git.theirsBase = nil, nil, nil
		vm.GitLink.Set("")
	}
	vm.git.forge, vm.git.rem = r.Forge, vm.git.gitd.Remote(r.Forge)
	vm.GitRepo.Set(r.Name)
	vm.raiseGitCommands()
}

// serverFailure handles an error from an operation in server mode; it reports
// whether it did.
func (vm *ViewModel) serverFailure(err error) bool {
	if vm.GitMode.Get() == GitModeClient {
		return false
	}
	switch {
	case errors.Is(err, gitd.ErrSignIn):
		vm.dropSession()
		vm.GitNote.Set(T(KGitNoteServerEnded))
		vm.notice(T(KGitNoteServerEnded))
	case errors.Is(err, gitd.ErrNoCredential):
		vm.gitSay(T(KGitNoteServerNoCred))
	default:
		vm.gitSay(T(KGitNoteServerFailed, ErrText(err)))
	}
	return true
}
