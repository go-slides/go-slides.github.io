package playground

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/go-browserhttp/browsergit/gitd"
	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-browserhttp/browsergit/secret"
	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-widgets/mvvm"
	"github.com/go-widgets/toolkit"
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
	GitModeClient      = iota // no server: a token in the tab, or a GitLab sign-in
	GitModeSignedOut          // a server, and the person must sign in
	GitModeSignedIn           // a server, and a session
	GitModeUnavailable        // served by a slides-gitd front that does not answer
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
	// GitdFront reports whether this page is served by a slides-gitd front: a
	// same-origin marker the front serves (/.well-known/slides-gitd), NOT a call
	// to the API, so that a server that is down is still known to be the server.
	GitdFront(ctx context.Context) bool
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
			if vm.held != nil {
				vm.GitNote.Set(vm.heldNote())
			}
			vm.sessionOp(true)
		}
	}
}

// GitdDeadline bounds the identity check at start: a server that hangs is a
// server that is unavailable, and nothing stored is shown while it does.
var GitdDeadline = 10 * time.Second

// sessionOp asks the server for the session. first: this is the discovery, and
// a server that is not there is no news.
func (vm *ViewModel) sessionOp(first bool) {
	h, ok := vm.host.(GitHost)
	if !ok {
		return
	}
	var info gitd.Info
	var err error
	var front bool
	vm.git.busy = true
	// until the first answer, nobody knows whether this page has a server: the
	// git frame (client mode's credential frame) must not mount (FrameAllowed)
	vm.git.discovering = vm.git.discovering || first
	vm.raiseGitCommands()
	h.GitRun(func() {
		ctx, cancel := context.WithTimeout(context.Background(), GitdDeadline)
		defer cancel()
		if gh, ok := vm.host.(GitdHost); ok {
			front = gh.GitdFront(ctx)
		}
		info, err = vm.git.gitd.Session(ctx)
	}, func() {
		vm.git.busy = false
		vm.git.front = front
		vm.takeSession(info, err, first)
		vm.git.discovering = false // after takeSession: the mode it chose is in place
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
	if vm.held != nil && vm.releaseHold(info, err) {
		return
	}
	switch {
	case err == nil:
		vm.enterServerMode()
		vm.bindOwner(info.User)
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
			vm.dropBase()
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
	case first && vm.git.front:
		// This page IS served by a slides-gitd front, which does not answer (a 502,
		// an outage, a page in its place): that is not "no server". Client mode
		// and its token field never come back on such a page.
		vm.setUnavailable()
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
	vm.releaseOwner()
	vm.GitMode.Set(GitModeSignedOut)
	vm.GitUser.Set("")
	vm.GitRepos.Set("")
	vm.git.grants, vm.git.rem = nil, nil
	vm.git.theirs, vm.git.theirsBase = nil, nil
	vm.baseChanged()
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
		vm.dropBase()
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

// The playground keeps the project per origin (the draft, the journal, the
// versions, among them the Git version of every repository loaded), not per
// person. In server mode the deck on screen may come from a repository only the
// signed-in person may read, so it is bound to them:
//
//   - the UI state records a hash of the subject the data belongs to
//     (git-owner), written once a session says who is signed in;
//   - a session for another subject, or no session at all while data bound to
//     someone is there (signed out, a 401), wipes the draft, the journal, the
//     versions, the undo steps and the editor before anything is shown to the
//     next person (go-slides/go-slides.github.io#99);
//   - data nobody owns (written in client mode or by an older page) is adopted
//     by the first signed-in person, and is left alone while signed out.

// ownerOf is what is recorded of a subject: a hash, so the UI state holds no
// identity.
func ownerOf(subject string) string {
	h := sha256.Sum256([]byte("go-slides/owner/" + subject))
	return hex.EncodeToString(h[:8])
}

// bindOwner binds the stored project to the signed-in subject, after wiping it
// if it was bound to someone else.
func (vm *ViewModel) bindOwner(subject string) {
	o := ownerOf(subject)
	if vm.git.owner != "" && vm.git.owner != o {
		vm.wipeLocal()
	}
	vm.git.owner = o
	vm.saveUI()
}

// releaseOwner is the signed-out state: what is bound to a person does not stay
// in the page for whoever uses it next.
func (vm *ViewModel) releaseOwner() {
	if vm.git.owner == "" {
		return
	}
	vm.wipeLocal()
	vm.git.owner = ""
}

// wipeLocal removes the project from this browser and from the page: the
// stored draft, journal, UI state and versions, the undo steps, any preview or
// comparison, what the pane loaded, and the editor, which shows the default
// example again.
func (vm *ViewModel) wipeLocal() {
	vm.backToNow()
	vm.ForgetDrafts()
	if vm.vers.v == nil && !vm.vers.off {
		vm.vers.wipe = true // the versions are not open yet: cleared when they are
	}
	vm.undo.project.Clear()
	vm.undo.files = map[string]toolkit.TextHistory{}
	if e := vm.undo.editor; e != nil {
		e.SetHistory(toolkit.TextHistory{})
	}
	vm.refreshUndo()
	vm.dropBase()
	vm.GitLink.Set("")
	vm.GitRepo.Set("")
	ex := Examples[max(vm.Example.Get(), 0)].Source
	vm.applyState(ProjectState{Files: map[string][]byte{DeckFile: []byte(ex)}, Open: DeckFile})
	vm.undo.project.Clear() // applying the example is not a step to undo either
	vm.refreshUndo()
}

// A page served by a slides-gitd front does not read, show or render anything
// stored in this browser until the server has said who is signed in
// (go-slides/go-slides.github.io#107): the previous person's deck, from a
// repository this person may not read, would be on screen for a round trip, and
// for as long as a hung request lasts.
//
//   - at start the stored project, the journal and the versions are set aside
//     (heldStorage); the page starts on the example deck, with "Signing in…";
//   - the session names someone: the project is restored when the stored owner
//     hash is theirs, and wiped when it is not or belongs to nobody;
//   - a 401: the project bound to someone is wiped; nothing is restored;
//   - no answer within GitdDeadline, or a front that answers wrongly: "The git
//     server is unavailable", and the stored project stays set aside, unread and
//     unwritten;
//   - a page that turns out not to be served by a front (the marker's probe was
//     inconclusive, then the server is plainly absent): restored as client mode.

// StorageHolder is a host that knows, before the view-model starts, that the
// page may be served by a slides-gitd front (the marker is there, or could not
// be told) and that nothing stored may be shown yet.
type StorageHolder interface {
	HoldStorage() bool
}

// heldStorage is the project set aside.
type heldStorage struct {
	st    *store.Store
	jr    Journal
	owner string // the owner hash recorded in its UI state ("": nobody's)
}

// holdStorage sets the stored project aside when the host asks for it.
func (vm *ViewModel) holdStorage() {
	sh, ok := vm.host.(StorageHolder)
	if !ok || !sh.HoldStorage() || (vm.drafts.st == nil && vm.drafts.jr == nil) {
		return
	}
	if _, ok := vm.host.(GitdHost); !ok {
		return
	}
	h := &heldStorage{st: vm.drafts.st, jr: vm.drafts.jr}
	if h.st != nil {
		if p := h.st.Cached(UIProject); p != nil {
			h.owner = p.UI["git-owner"]
		}
	}
	vm.held = h
	vm.drafts.st, vm.drafts.jr = nil, nil
}

// heldNote is what the page says while it waits.
func (vm *ViewModel) heldNote() string { return T(KGitNoteServerHold) }

// attachHeld gives the page its storage back, unread.
func (vm *ViewModel) attachHeld() *heldStorage {
	h := vm.held
	vm.held = nil
	vm.drafts.st, vm.drafts.jr = h.st, h.jr
	vm.git.owner = h.owner
	vm.openVersions() // its versions are open now: they can be listed, or cleared
	return h
}

// restoreHeld loads the project the page set aside, as a start would have.
func (vm *ViewModel) restoreHeld() {
	src, status := vm.startDrafts("", false)
	vm.Source.Set(src)
	vm.drafts.dirty = false
	vm.drafts.journaled = vm.journalSig()
	vm.loadUI()
	if p := vm.drafts.st; p != nil {
		if c := p.Cached(UIProject); c != nil {
			vm.loadGitUI(func(k string) string { return c.UI[k] })
		}
	}
	vm.refreshDraftMenu()
	if status != "" {
		vm.notice(status)
	}
}

// releaseHold decides what the answer means for the project set aside. It
// reports whether takeSession has nothing more to do.
func (vm *ViewModel) releaseHold(info gitd.Info, err error) bool {
	switch {
	case err == nil:
		h := vm.attachHeld()
		if h.owner != "" && h.owner == ownerOf(info.User) {
			vm.restoreHeld()
		} else {
			// another person's, or nobody's: the pane's bindOwner wipes the former;
			// the latter is wiped here too, nothing of it is shown to anyone
			vm.git.owner = "unowned" // is no subject's hash: bindOwner wipes it
		}
		return false
	case errors.Is(err, gitd.ErrSignIn):
		vm.attachHeld() // dropSession wipes what is bound to someone
		return false
	case vm.git.front || (vm.held.owner != "" && !errors.Is(err, gitd.ErrNoServer)):
		// a front that does not answer, or no answer in time (and a project bound
		// to someone is set aside): nothing stored shows
		vm.enterServerMode()
		vm.setUnavailable()
		return true
	}
	// plainly no server and no front: an ordinary page, its project comes back
	vm.attachHeld()
	vm.restoreHeld()
	return false
}

// setUnavailable is the state of a page served by a front that does not answer.
func (vm *ViewModel) setUnavailable() {
	vm.enterServerMode()
	vm.releaseOwner()
	vm.GitMode.Set(GitModeUnavailable)
	vm.GitUser.Set("")
	vm.GitRepos.Set("")
	vm.git.grants, vm.git.rem = nil, nil
	vm.GitConnected.Set(false)
	vm.GitNote.Set(T(KGitNoteServerDown))
	vm.notice(T(KGitNoteServerDown))
}
