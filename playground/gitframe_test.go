package playground

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-widgets/toolkit"
)

// frameTab is a page whose client mode runs through the git frame. Its frame is
// the fake forge behind the typed commands: what reaches it is exactly what a
// real frame would be sent, and it is all there is to send (no token).
type frameTab struct {
	*gitTab
	watch   func(FrameState)
	calls   []string
	sent    []map[string]any
	sentF   []remote.Files
	failing error
}

func (f *frameTab) FrameMode() bool { return true }

func (f *frameTab) FrameWatch(w func(FrameState)) { f.watch = w }

func (f *frameTab) FrameCall(cmd string, fields map[string]any, files remote.Files) (json.RawMessage, remote.Files, error) {
	f.calls = append(f.calls, cmd)
	f.sent = append(f.sent, fields)
	f.sentF = append(f.sentF, files)
	if f.failing != nil {
		return nil, nil, f.failing
	}
	ctx := context.Background()
	str := func(k string) string { s, _ := fields[k].(string); return s }
	base := func() remote.Base {
		var b remote.Base
		j, _ := json.Marshal(fields["base"])
		_ = json.Unmarshal(j, &b)
		return b
	}
	enc := func(v any) json.RawMessage { j, _ := json.Marshal(v); return j }
	switch cmd {
	case "load":
		files, b, err := f.forge.Load(ctx, str("repo"), str("ref"), str("dir"))
		return enc(map[string]any{"base": b}), files, err
	case "propose":
		// the frame names the branch, as the real one does
		res, err := f.forge.Propose(ctx, base(), files, remote.Proposal{
			Branch: remote.BranchName("slides/", str("title"), time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)),
			Title:  str("title"), Body: str("body"),
		})
		return enc(map[string]string{"branch": res.Branch, "commit": res.Commit, "url": res.URL}), nil, err
	case "refresh":
		files, b, err := f.forge.Refresh(ctx, base())
		return enc(map[string]any{"base": b, "moved": files != nil}), files, err
	case "disconnect":
		if f.watch != nil {
			f.watch(FrameState{})
		}
		return enc(map[string]any{"connected": false}), nil, nil
	}
	return nil, nil, errors.New("unknown command")
}

func frameScene(t *testing.T, m *store.Memory, kept string) (*Scene, *ViewModel, *frameTab) {
	t.Helper()
	SetupText(1)
	g := &gitTab{
		promptTab: &promptTab{filesTab: &filesTab{tab: &tab{fakeHost: newHost(), answer: true}, saved: map[string][]byte{}}},
		forge: &fakeForge{files: map[string]string{
			"README.md":    "# repo",
			"talk/deck.md": "---\ntheme: default\n---\n\n# From GitHub\n",
		}, branches: map[string]bool{}},
		kept: kept, remoteOK: true,
	}
	if kept != "" {
		g.keptForge, g.keptOrigin = "github", "https://api.github.com"
	}
	g.st = store.New(m, NewTabID(), load(m, DraftProject), load(m, PreviousProject), load(m, UIProject))
	h := &frameTab{gitTab: g}
	tr := &fakeTransport{}
	vm := NewViewModel(tr, h, "")
	tr.vm = vm
	answer(tr, vm)
	vm.drafts.now = func() time.Time { return time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC) }
	return NewScene(vm, h, 1600, 900), vm, h
}

// The whole journey through the frame: nothing in the page ever holds a token,
// so there is nothing for the sweep to find; the commands are the typed ones.
func TestTheFramePaneHoldsNoToken(t *testing.T) {
	m := store.NewMemory()
	s, vm, h := frameScene(t, m, "")
	if h.watch == nil || !vm.frameOn() {
		t.Fatal("the pane is not in frame mode")
	}
	vm.GitPane.Set(true)
	if vm.GitNote.Get() != T(KGitNoteFrameOff) {
		t.Fatalf("the off note: %q", vm.GitNote.Get())
	}
	// Connect on this page is not a thing: even a typed token is not taken.
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	if vm.GitConnected.Get() || len(h.tokens) != 0 || vm.git.token.Reveal() != "" {
		t.Fatal("this page took a token while the frame is on")
	}
	vm.GitToken.Set("")

	// the frame connects (the person typed the token THERE)
	h.watch(FrameState{Connected: true, Forge: "github"})
	if !vm.GitConnected.Get() || vm.GitNote.Get() != T(KGitNoteFrameOn) || vm.forgeName() != "github" {
		t.Fatalf("connected? %v %q", vm.GitConnected.Get(), vm.GitNote.Get())
	}
	vm.GitRepo.Set("o/r")
	vm.GitDir.Set("talk")
	vm.GitLoad.Execute()
	if !strings.Contains(vm.Source.Get(), "From GitHub") || vm.git.base == nil || h.calls[len(h.calls)-1] != "load" {
		t.Fatalf("load: %q %v", vm.Source.Get(), h.calls)
	}
	tr := vm.transport.(*fakeTransport)
	answer(tr, vm)
	vm.Source.Set(vm.Source.Get() + "\n---\n\n# Mine\n")
	answer(tr, vm)
	h.name = "Add a slide"
	vm.GitPropose.Execute()
	if len(h.forge.proposed) != 1 || h.forge.proposed[0].Title != "Add a slide" || !strings.HasPrefix(h.forge.proposed[0].Branch, "slides/add-a-slide-") {
		t.Fatalf("proposed %+v (note %q)", h.forge.proposed, vm.GitNote.Get())
	}
	// the page sent no branch: the frame names it
	last := h.sent[len(h.sent)-1]
	if _, has := last["branch"]; has || last["title"] != "Add a slide" {
		t.Fatalf("the propose command: %v", last)
	}
	if vm.GitLink.Get() != "https://github.com/o/r/pull/1" {
		t.Fatalf("link %q", vm.GitLink.Get())
	}
	// someone else pushes: Check for changes
	h.forge.files["talk/deck.md"] = "# Theirs\n"
	h.forge.commit++
	vm.GitRefresh.Execute()
	if vm.Compare.Get() == nil || !vm.GitTakeTheirs.CanExecute() {
		t.Fatalf("refresh did not compare (note %q)", vm.GitNote.Get())
	}
	// every command that went to the frame is a typed one, and none carries a token
	for i, c := range h.calls {
		if c != "load" && c != "propose" && c != "refresh" && c != "disconnect" {
			t.Errorf("an untyped command %q", c)
		}
		j, _ := json.Marshal(h.sent[i])
		if strings.Contains(string(j), "token") && c != "propose" { // "token" never as a field name
			t.Errorf("%s carries a token field: %s", c, j)
		}
	}
	if leaked(everything(t, m, h.gitTab, vm, tr.sent, nil)) {
		t.Fatal("a token is in what the page wrote or showed")
	}

	// the pane shows no token field: its rectangles are not offered
	rects := map[string]toolkit.Rect{}
	s.gitRects(rects)
	for _, k := range []string{"git-token", "git-remember", "git-connect", "git-signin", "git-forge"} {
		if _, has := rects[k]; has {
			t.Errorf("the pane still has %s", k)
		}
	}
	if _, ok := rects["git-repo"]; !ok {
		t.Error("the pane lost its repository field")
	}
	vm.GitPane.Set(false)
	if _, ok := s.GitFrameRect(); ok {
		t.Error("the frame is placed while the pane is closed")
	}
	vm.GitPane.Set(true)

	// Disconnect goes to the frame, which announces it
	vm.GitDisconnect.Execute()
	if h.calls[len(h.calls)-1] != "disconnect" || vm.GitConnected.Get() || vm.GitNote.Get() != T(KGitNoteFrameOff) {
		t.Fatalf("disconnect: %v %v %q", h.calls, vm.GitConnected.Get(), vm.GitNote.Get())
	}
	if vm.GitLoad.CanExecute() {
		t.Error("Load is enabled with no connection")
	}
}

// A token an earlier page kept in this origin's storage is the thing the frame
// exists to keep out of the page: dropped at first load, and the person told.
func TestAKeptTokenIsDroppedAndTheUserTold(t *testing.T) {
	m := store.NewMemory()
	_, vm, h := frameScene(t, m, planted)
	if h.kept != "" || h.keptForge != "" {
		t.Fatal("the kept token is still in this page's storage")
	}
	if len(h.tokens) != 0 || vm.GitConnected.Get() {
		t.Fatal("the kept token was used")
	}
	if !strings.HasPrefix(vm.GitNote.Get(), T(KGitNoteMigrated)) || !strings.Contains(vm.GitNote.Get(), T(KGitNoteFrameOff)) {
		t.Fatalf("the person is not told: %q", vm.GitNote.Get())
	}
	// and a page with nothing kept says nothing of it
	_, vm2, _ := frameScene(t, store.NewMemory(), "")
	if strings.Contains(vm2.GitNote.Get(), T(KGitNoteMigrated)) {
		t.Fatal("a migration notice with nothing migrated")
	}
}

func TestAFrameFailureIsShownNotHidden(t *testing.T) {
	_, vm, h := frameScene(t, store.NewMemory(), "")
	h.watch(FrameState{Connected: true, Forge: "gitlab", Remembered: true})
	if vm.forgeName() != "gitlab" || vm.GitNote.Get() != T(KGitNoteFrameKept) {
		t.Fatalf("forge %q note %q", vm.forgeName(), vm.GitNote.Get())
	}
	vm.GitRepo.Set("o/r")
	h.failing = errors.New("frame: command refused: repo")
	vm.GitLoad.Execute()
	if !strings.Contains(vm.GitNote.Get(), "command refused") {
		t.Fatalf("the failure was not shown: %q", vm.GitNote.Get())
	}
}

// Server mode is unchanged: a state the frame announces while a slides-gitd
// server holds the credentials is not taken up.
func TestTheFrameIsIgnoredInServerMode(t *testing.T) {
	_, vm, h := frameScene(t, store.NewMemory(), "")
	vm.GitMode.Set(GitModeSignedIn)
	h.watch(FrameState{Connected: true, Forge: "github"})
	if _, isFrame := vm.git.rem.(frameRemote); isFrame || vm.git.forge == "github" {
		t.Fatal("the frame's connection replaced the server's")
	}
}

func TestFrameRemoteRefusesAnswersThatAreNotWhatWasAsked(t *testing.T) {
	h := &frameTab{gitTab: &gitTab{forge: &fakeForge{}}}
	r := frameRemote{h}
	h.failing = errors.New("down")
	if _, _, err := r.Load(context.Background(), "o/r", "", ""); err == nil {
		t.Error("load: an error was swallowed")
	}
	if _, _, err := r.Refresh(context.Background(), remote.Base{}); err == nil {
		t.Error("refresh: an error was swallowed")
	}
	if _, err := r.Propose(context.Background(), remote.Base{}, nil, remote.Proposal{Title: "t"}); err == nil {
		t.Error("propose: an error was swallowed")
	}
}
