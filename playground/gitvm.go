package playground

import (
	"context"
	"encoding/json"
	"path"
	"slices"
	"strings"

	"github.com/go-browserhttp/browsergit/oauth"
	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-browserhttp/browsergit/secret"
	"github.com/go-slides/go-slides.github.io/playground/versions"
	"github.com/go-widgets/mvvm"
)

// The GitHub pane: load a deck's folder from a GitHub repository, propose the
// project back as a NEW branch and a pull request, and see what changed there
// since (go-browserhttp/browsergit; design: design/git.md).
//
// The token is the one secret the playground holds, and where it may go is the
// point of this file:
//
//   - it is typed into a masked field (GitToken), wrapped in a secret.Token
//     and the field emptied at once: nothing else holds the text;
//   - it lives in this tab's memory, in the page's instance of the module and
//     nowhere else: never in the render worker (a separate instance, sent only
//     render requests), never in the preview frames (sent only slides), never in
//     the project, the draft, the journal, the versions, the UI state, a share
//     link or an export — none of them has a field it could go in;
//   - "Remember on this device" is the one exception, and the person's choice:
//     the host keeps it (TokenKeeper), outside every store the deck can read;
//     Disconnect and Forget my drafts drop it;
//   - every message that comes back from GitHub is redacted before it is
//     shown, and the requests go straight to the forge's API (never through a
//     proxy that would see the token). The page's CSP names the forges' API
//     origins in connect-src, so fetch, XHR and beacons to any other host are
//     refused, and it closes forms, frames, objects, workers and <base> to other
//     origins; it does not stop a script injected into the page from sending a
//     token out as the URL of an image (see layouts/_default/playground.html).

// GitHost is what the page does for the GitHub pane. A host without it has no
// pane worth showing; its commands stay disabled.
type GitHost interface {
	// GitRemote is the REST remote of forge ("github", "gitlab") that
	// authenticates with tok.
	GitRemote(forge string, tok secret.Token) (remote.Remote, error)
	// GitRun runs op away from the UI's turn (it blocks on the network), then
	// done on the UI's turn.
	GitRun(op func(), done func())
}

// TokenKeeper keeps a token on this device, when the person asks for it.
type TokenKeeper interface {
	KeepToken(tok secret.Token)
	KeptToken() secret.Token
	ForgetToken()
}

// gitBranchPrefix is where every proposal's branch goes.
const gitBranchPrefix = "slides/"

// gitBaseKey is the UI state's key for the loaded base (no secret in it).
const gitBaseKey = "git-base"

// gitState is the view-model's part of the GitHub pane.
type gitState struct {
	token secret.Token // a personal access token typed in
	// oauth is a GitLab sign-in (gitoauth.go): its tokens refresh themselves
	// before use; rem is rebuilt from the current access token each time
	oauth    *oauth.Tokens
	oauthCfg oauth.Config
	forge    string // what rem talks to: "github" or "gitlab"
	rem      remote.Remote
	base     *remote.Base
	busy     bool
	// theirs is what Check for changes found, until Take theirs or another load
	theirs     remote.Files
	theirsBase *remote.Base
}

// initGit makes the observables and commands; NewViewModel calls it.
func (vm *ViewModel) initGit() {
	vm.GitPane = mvvm.NewObservable(false)
	vm.GitRepo = mvvm.NewObservable("")
	vm.GitRef = mvvm.NewObservable("")
	vm.GitDir = mvvm.NewObservable("")
	vm.GitToken = mvvm.NewObservable("")
	vm.GitRemember = mvvm.NewObservable(false)
	vm.GitConnected = mvvm.NewObservable(false)
	vm.GitNote = mvvm.NewObservable(T(KGitNoteOff))
	vm.GitLink = mvvm.NewObservable("")
	vm.GitForge = mvvm.NewObservable(0)
	_, hasHost := vm.host.(GitHost)
	idle := func() bool { return hasHost && !vm.git.busy }
	on := func() bool { return idle() && vm.git.rem != nil }
	vm.GitConnect = mvvm.NewCommand(vm.connectGit, idle)
	vm.GitDisconnect = mvvm.NewCommand(vm.disconnectGit, func() bool { return vm.git.rem != nil })
	vm.GitLoad = mvvm.NewCommand(vm.loadGit, on)
	vm.GitPropose = mvvm.NewCommand(vm.proposeGit, func() bool { return on() && vm.git.base != nil })
	vm.GitRefresh = mvvm.NewCommand(vm.refreshGit, func() bool { return on() && vm.git.base != nil })
	vm.GitTakeTheirs = mvvm.NewCommand(vm.takeTheirs, func() bool { return idle() && vm.git.theirs != nil })
	vm.GitOpenLink = mvvm.NewCommand(func() { vm.host.Open(vm.GitLink.Get()) }, func() bool { return vm.GitLink.Get() != "" })
	vm.GitLink.Subscribe(func(string) { vm.GitOpenLink.RaiseCanExecuteChanged() })
	if d := vm.drafts; d.st != nil {
		if p := d.st.Cached(UIProject); p != nil {
			vm.loadGitUI(func(k string) string { return p.UI[k] })
		}
	}
	for _, o := range []*mvvm.Observable[string]{vm.GitRepo, vm.GitRef, vm.GitDir} {
		o.Subscribe(func(string) { vm.saveUI() })
	}
	vm.GitNote.Set(vm.offNote())
	// before any sign-in news, which must not be overwritten by the above
	vm.initOAuth(idle)
	// another forge: the connection (and a token kept for it) goes
	vm.GitForge.Subscribe(func(int) {
		if vm.git.rem != nil || vm.git.oauth != nil {
			vm.disconnectGit()
		}
		vm.GitSignIn.RaiseCanExecuteChanged()
		if !vm.GitConnected.Get() {
			vm.GitNote.Set(vm.offNote())
		}
		vm.saveUI()
	})
	vm.GitRemember.Subscribe(func(b bool) {
		vm.saveUI()
		if k, ok := vm.host.(TokenKeeper); ok && !b {
			k.ForgetToken() // unticking forgets at once
		}
	})
	// a token remembered on this device connects at once; one kept while the
	// choice says otherwise is dropped
	if k, ok := vm.host.(TokenKeeper); ok && hasHost {
		if tok := k.KeptToken(); !tok.IsZero() {
			if vm.GitRemember.Get() {
				vm.useToken(tok)
			} else {
				tok.Forget()
				k.ForgetToken()
			}
		}
	}
}

func (vm *ViewModel) raiseGitCommands() {
	for _, c := range []*mvvm.Command{vm.GitConnect, vm.GitDisconnect, vm.GitLoad, vm.GitPropose, vm.GitRefresh, vm.GitTakeTheirs} {
		c.RaiseCanExecuteChanged()
	}
}

// gitSay puts a message in the pane and on the status line, redacted: a
// message that came back from the network could carry the token.
func (vm *ViewModel) gitSay(s string) {
	s = secret.Redact(s, vm.git.token)
	if t := vm.git.oauth; t != nil {
		s = secret.Redact(s, t.Access, t.Refresh)
	}
	vm.GitNote.Set(s)
	vm.notice(s)
}

// connectGit takes the token typed into the masked field.
func (vm *ViewModel) connectGit() {
	tok := secret.New(vm.GitToken.Get())
	vm.GitToken.Set("") // the field, and so its text, empty at once
	if tok.IsZero() {
		vm.gitSay(T(KGitNoteNoToken))
		return
	}
	vm.useToken(tok)
	if !vm.GitConnected.Get() {
		return
	}
	if k, ok := vm.host.(TokenKeeper); ok && vm.GitRemember.Get() {
		k.KeepToken(tok)
	}
}

func (vm *ViewModel) useToken(tok secret.Token) {
	h, ok := vm.host.(GitHost)
	if !ok {
		return
	}
	rem, err := h.GitRemote(vm.forgeName(), tok)
	if err != nil {
		// redacted with THIS token: it is not the pane's yet
		msg := secret.Redact(T(KGitNoteFailed, ErrText(err)), tok)
		tok.Forget()
		vm.gitSay(msg)
		return
	}
	vm.git.token.Forget()
	vm.git.token, vm.git.rem, vm.git.forge = tok, rem, vm.forgeName()
	vm.GitConnected.Set(true)
	if vm.GitRemember.Get() {
		vm.GitNote.Set(T(KGitNoteRemember))
	} else {
		vm.GitNote.Set(T(KGitNoteConnected))
	}
	vm.raiseGitCommands()
}

// disconnectGit drops the token: from memory, and from this device; a GitLab
// sign-in is revoked at GitLab (gitoauth.go) and forgotten here at once.
func (vm *ViewModel) disconnectGit() {
	vm.signOut()
	vm.git.token.Forget()
	vm.git.token, vm.git.rem = secret.Token{}, nil
	if k, ok := vm.host.(TokenKeeper); ok {
		k.ForgetToken()
	}
	vm.GitConnected.Set(false)
	vm.GitNote.Set(vm.offNote())
	vm.raiseGitCommands()
}

// gitRun runs op on the network and then done, with the pane busy meanwhile.
func (vm *ViewModel) gitRun(op func(context.Context) error, done func(error)) {
	h := vm.host.(GitHost)
	vm.git.busy = true
	vm.raiseGitCommands()
	vm.GitNote.Set(T(KGitNoteBusy))
	var err error
	h.GitRun(func() { err = op(context.Background()) }, func() {
		vm.git.busy = false
		vm.afterOp()
		vm.raiseGitCommands()
		if err != nil {
			vm.gitSay(T(KGitNoteFailed, ErrText(err)))
			return
		}
		done(nil)
	})
}

// stateOf turns a folder read from GitHub into a project state.
func stateOf(files remote.Files) ProjectState {
	s := ProjectState{Files: map[string][]byte{}, Open: DeckFile}
	dirs := map[string]bool{}
	for p, b := range files {
		s.Files[p] = b
		for d := path.Dir(p); d != "." && d != "/"; d = path.Dir(d) {
			dirs[d] = true
		}
	}
	for d := range dirs {
		s.Dirs = append(s.Dirs, d)
	}
	slices.Sort(s.Dirs)
	return s
}

// gitName names a base: "owner/name@1a2b3c4".
func gitName(b remote.Base) string {
	name := b.Repo
	if b.Dir != "" {
		name += "/" + b.Dir
	}
	c := b.Commit
	if len(c) > 7 {
		c = c[:7]
	}
	return name + "@" + c
}

// loadGit replaces the project with the folder's files: one step of Undo,
// the project as it was kept as a version first.
func (vm *ViewModel) loadGit() {
	repo, ref, dir := strings.TrimSpace(vm.GitRepo.Get()), strings.TrimSpace(vm.GitRef.Get()), strings.Trim(strings.TrimSpace(vm.GitDir.Get()), "/")
	use := vm.remoteForOp()
	var files remote.Files
	var base remote.Base
	vm.gitRun(func(ctx context.Context) error {
		rem, err := use(ctx)
		if err != nil {
			return err
		}
		files, base, err = rem.Load(ctx, repo, ref, dir)
		return err
	}, func(error) { vm.takeLoaded(files, base, T(KStepGitLoad, repo)) })
}

// takeLoaded makes files the project, and base what proposals start from.
func (vm *ViewModel) takeLoaded(files remote.Files, base remote.Base, step string) {
	if _, ok := files[DeckFile]; !ok {
		vm.gitSay(T(KGitNoteNoDeck, gitName(base), DeckFile))
		return
	}
	s := stateOf(files)
	vm.projectStep(step, func() bool {
		vm.applyState(s)
		return true
	})
	vm.git.base = &base
	vm.git.theirs, vm.git.theirsBase = nil, nil
	vm.GitLink.Set("")
	vm.keepVersion(s, versions.Git, gitName(base), nil)
	vm.saveUI()
	vm.raiseGitCommands()
	vm.gitSay(T(KGitNoteLoaded, gitName(base)))
}

// proposeGit sends the project as a new branch and opens a pull request.
func (vm *ViewModel) proposeGit() {
	base := *vm.git.base
	title := T(KGitDefaultTitle)
	if p, ok := vm.host.(Prompter); ok {
		t, ok := p.Prompt(T(KAskGitTitle), title)
		if !ok {
			return
		}
		if t = strings.TrimSpace(t); t != "" {
			title = t
		}
	}
	if len([]rune(title)) > 120 {
		title = string([]rune(title)[:120])
	}
	prop := remote.Proposal{
		Branch: remote.BranchName(gitBranchPrefix, title, vm.drafts.now()),
		Title:  title,
		Body:   T(KGitBody),
	}
	files := remote.Files(vm.captureState().Files)
	use := vm.remoteForOp()
	var res remote.Result
	vm.gitRun(func(ctx context.Context) error {
		rem, err := use(ctx)
		if err != nil {
			return err
		}
		res, err = rem.Propose(ctx, base, files, prop)
		return err
	}, func(error) {
		vm.GitLink.Set(res.URL)
		vm.gitSay(T(KGitNoteProposed, res.Branch))
	})
}

// refreshGit asks whether the branch moved; if so it compares what is there
// with now, and Take theirs loads it.
func (vm *ViewModel) refreshGit() {
	base := *vm.git.base
	use := vm.remoteForOp()
	var theirs remote.Files
	var head remote.Base
	vm.gitRun(func(ctx context.Context) error {
		rem, err := use(ctx)
		if err != nil {
			return err
		}
		theirs, head, err = rem.Refresh(ctx, base)
		return err
	}, func(error) {
		if theirs == nil {
			vm.gitSay(T(KGitNoteUpToDate))
			return
		}
		vm.git.theirs, vm.git.theirsBase = theirs, &head
		vm.raiseGitCommands()
		vm.CompareFile.Set(0)
		c := Compare(stateOf(theirs), vm.captureState(), T(KCompareTitle, T(KVersionGit, gitName(head))), -1)
		c.Name = T(KVersionGit, gitName(head))
		vm.Compare.Set(c)
		vm.gitSay(T(KGitNoteMoved, gitName(head)))
	})
}

// takeTheirs loads what Check for changes found.
func (vm *ViewModel) takeTheirs() {
	files, base := vm.git.theirs, vm.git.theirsBase
	if files == nil || base == nil {
		return
	}
	vm.backToNow()
	vm.takeLoaded(files, *base, T(KStepGitLoad, base.Repo))
}

// saveGitUI writes the pane's fields into the UI state: the repository, the
// branch, the folder, the base loaded and the Remember choice. Never the token.
func (vm *ViewModel) saveGitUI(set func(key, value string)) {
	set("git-repo", vm.GitRepo.Get())
	set("git-ref", vm.GitRef.Get())
	set("git-dir", vm.GitDir.Get())
	if vm.GitRemember.Get() {
		set("git-remember", "true")
	} else {
		set("git-remember", "false")
	}
	set("git-forge", vm.forgeName())
	if b := vm.git.base; b != nil {
		if j, err := json.Marshal(b); err == nil {
			set(gitBaseKey, string(j))
		}
	}
}

// loadGitUI reads them back; a base that does not parse is dropped.
func (vm *ViewModel) loadGitUI(get func(key string) string) {
	vm.quiet = true
	defer func() { vm.quiet = false }()
	vm.GitRepo.Set(get("git-repo"))
	vm.GitRef.Set(get("git-ref"))
	vm.GitDir.Set(get("git-dir"))
	vm.GitRemember.Set(get("git-remember") == "true")
	for i, n := range gitForgeNames {
		if get("git-forge") == n {
			vm.GitForge.Set(i)
		}
	}
	if j := get(gitBaseKey); j != "" {
		var b remote.Base
		if json.Unmarshal([]byte(j), &b) == nil && remote.CheckRepo(b.Repo, false) == nil && remote.IsCommitID(b.Commit) {
			vm.git.base = &b
		}
	}
}
