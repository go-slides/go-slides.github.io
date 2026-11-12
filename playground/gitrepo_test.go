package playground

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-widgets/mvvm"
	"github.com/go-widgets/toolkit"
	"github.com/go-widgets/vcs"
)

func jsonOf(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func sums(m map[string]string) map[string]string {
	out := map[string]string{}
	for p, c := range m {
		out[p] = sumOf([]byte(c))
	}
	return out
}

func TestDiffStatusAgainstTheBase(t *testing.T) {
	base := sums(map[string]string{"deck.md": "a", "same.css": "s", "gone.png": "g", "both.md": "b", "twin.md": "t", "up.md": "u"})
	work := map[string][]byte{
		"deck.md": []byte("a2"), "same.css": []byte("s"), "new.md": []byte("n"),
		"both.md": []byte("b-mine"), "twin.md": []byte("t-both"), "up.md": []byte("u"),
	}
	got, err := diffStatus(context.Background(), base, work, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]vcs.State{"deck.md": vcs.Modified, "new.md": vcs.Added, "gone.png": vcs.Deleted, "both.md": vcs.Modified, "twin.md": vcs.Modified}
	if !maps.Equal(got, want) {
		t.Fatalf("status %v, want %v", got, want)
	}

	// Check for changes found the branch moved: where it and the project both
	// changed a path, to different content, the path is a conflict; the same change
	// on both sides is not; a path only the branch changed is not the project's
	theirs := remote.Files{
		"deck.md": []byte("a2"), "same.css": []byte("s"), "gone.png": nil, "both.md": []byte("b-theirs"),
		"twin.md": []byte("t-both"), "up.md": []byte("u-theirs"), "new.md": []byte("n-theirs"),
	}
	delete(theirs, "gone.png") // the branch deleted it too
	got, err = diffStatus(context.Background(), base, work, theirs)
	if err != nil {
		t.Fatal(err)
	}
	want = map[string]vcs.State{
		"deck.md":  vcs.Modified,   // the branch has the same edit: nothing to resolve
		"new.md":   vcs.Conflicted, // added on both sides, differently
		"gone.png": vcs.Deleted,    // deleted on both sides
		"both.md":  vcs.Conflicted,
		"twin.md":  vcs.Modified,
	}
	if !maps.Equal(got, want) {
		t.Fatalf("with theirs %v, want %v", got, want)
	}
	// deleted here, edited there; edited here, deleted there
	got, _ = diffStatus(context.Background(), sums(map[string]string{"x": "1", "y": "1"}),
		map[string][]byte{"y": []byte("2")}, remote.Files{"x": []byte("9")})
	if got["x"] != vcs.Conflicted || got["y"] != vcs.Conflicted {
		t.Fatalf("delete/edit %v", got)
	}
	// a branch that left a path alone while the project changed it
	got, _ = diffStatus(context.Background(), sums(map[string]string{"x": "1"}), map[string][]byte{"x": []byte("2")}, remote.Files{"x": []byte("1")})
	if got["x"] != vcs.Modified {
		t.Fatalf("only we changed it: %v", got)
	}
}

func TestDiffStatusStopsWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := diffStatus(ctx, sums(map[string]string{"a": "1"}), map[string][]byte{"a": []byte("2")}, nil); err == nil {
		t.Error("a cancelled context must return its error, not a partial map")
	}
	if _, err := diffStatus(ctx, sums(map[string]string{"a": "1"}), map[string][]byte{}, nil); err != nil {
		t.Error("nothing to read, nothing to refuse")
	}
	// the conflict pass honours it too
	n := 0
	cctx := &stoppingContext{Context: context.Background(), after: &n}
	if _, err := diffStatus(cctx, sums(map[string]string{"a": "1"}), map[string][]byte{"a": []byte("2")}, remote.Files{"a": []byte("3")}); err == nil {
		t.Error("the conflict pass must stop on a cancelled context")
	}
	var r vcs.Repository = &projectRepo{}
	if _, err := r.Status(ctx); err == nil {
		t.Error("Status of a cancelled context")
	}
	if r.Root() != "/" {
		t.Errorf("root %q", r.Root())
	}
}

// stoppingContext reports cancelled from its second Err on.
type stoppingContext struct {
	context.Context
	after *int
}

func (c *stoppingContext) Err() error {
	*c.after++
	if *c.after > 1 {
		return context.Canceled
	}
	return nil
}

func TestParseSumsRefusesAnythingElse(t *testing.T) {
	good := sumOf([]byte("x"))
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"", false}, {"not json", false}, {`{"a":"zz"}`, false}, {`{"a":"abcd"}`, false},
		{`{"a":"` + good + `"}`, true}, {`{}`, true},
	} {
		if got := parseSums(tc.in) != nil; got != tc.want {
			t.Errorf("parseSums(%q) kept: %v", tc.in, got)
		}
	}
	big := map[string]string{}
	for i := range 4097 {
		big[strings.Repeat("a", i+1)] = good
	}
	if parseSums(jsonOf(t, big)) != nil {
		t.Error("an unbounded map is dropped")
	}
}

func TestTreeEntriesListWhatTheBaseHadAndTheProjectLost(t *testing.T) {
	got := treeEntries([]ProjectFile{{Path: "a.md", Size: 3}, {Path: "d", Dir: true}}, []string{"gone.png"})
	want := []toolkit.FileEntry{{Path: "a.md", Size: 3}, {Path: "d", Dir: true}, {Path: "gone.png"}}
	if !slices.Equal(got, want) {
		t.Fatalf("%v, want %v", got, want)
	}
}

// The explorer's marks come from the project against the base: nothing before a
// load, then Modified, Added and Deleted as the person works, Conflicted where
// Check for changes finds the branch moved on a file the project changed too, and
// all of it comes back after a reload.
func TestTheExplorerMarksTheProjectAgainstTheBase(t *testing.T) {
	m := store.NewMemory()
	s, vm, h := gitScene(t, m, "")
	tr := vm.transport.(*fakeTransport)
	tree := s.files.tree
	vm.GitPane.Set(true)
	vm.GitRepo.Set("o/r")
	vm.GitDir.Set("talk")
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	if got := tree.Marks().Get(); len(got) != 0 {
		t.Fatalf("nothing is loaded, nothing is marked: %v", got)
	}
	if tree.Repository() == nil || tree.StatusError().Get() != nil {
		t.Fatalf("repository %v err %v", tree.Repository(), tree.StatusError().Get())
	}
	vm.GitLoad.Execute()
	answer(tr, vm)
	if got := tree.Marks().Get(); len(got) != 0 {
		t.Fatalf("just loaded: %v", got)
	}

	vm.Source.Set(vm.Source.Get() + "\n---\n\n# Mine\n")
	answer(tr, vm)
	if _, err := vm.AddFile("", "notes.md", []byte("n")); err != nil {
		t.Fatal(err)
	}
	if err := vm.DeletePath("media/logo.svg"); err != nil {
		t.Fatal(err)
	}
	marks := tree.Marks().Get()
	if marks[DeckFile] != toolkit.FileModified || marks["notes.md"] != toolkit.FileAdded || marks["media/logo.svg"] != toolkit.FileDeleted {
		t.Fatalf("marks %v", marks)
	}
	if marks["media"] != toolkit.FileModified {
		t.Errorf("a folder is Modified when anything under it is: %v", marks)
	}
	// the deleted file is a row, and the explorer refuses to act on it
	if !slices.Contains(vm.deletedFromBase(), "media/logo.svg") {
		t.Fatalf("deleted %v", vm.deletedFromBase())
	}
	vm.Open.Set(DeckFile)
	tree.OnRequest(toolkit.FileRequest{Action: toolkit.FileOpen, Path: "media/logo.svg"})
	tree.OnRequest(toolkit.FileRequest{Action: toolkit.FileDelete, Path: "media/logo.svg"})
	if vm.Open.Get() != DeckFile {
		t.Error("a deleted row was opened")
	}

	// the branch moved on the deck, which the project changed too
	h.forge.files["talk/deck.md"] = "# Theirs\n"
	h.forge.commit++
	vm.GitRefresh.Execute()
	if got := tree.Marks().Get()[DeckFile]; got != toolkit.FileConflicted {
		t.Fatalf("both changed the deck: %v (note %q)", got, vm.GitNote.Get())
	}
	if tree.Marks().Get()["media"] != toolkit.FileModified {
		t.Error("a folder keeps its roll-up")
	}

	// it survives a reload: the base and what it held come back from the UI state
	vm.saveUI()
	h2 := *h
	h2.st = store.New(m, NewTabID(), load(m, DraftProject), load(m, PreviousProject), load(m, UIProject))
	tr2 := &fakeTransport{}
	vm2 := NewViewModel(tr2, &h2, "")
	tr2.vm = vm2
	answer(tr2, vm2)
	if vm2.git.base == nil || !maps.Equal(vm2.git.baseSums, vm.git.baseSums) {
		t.Fatalf("reload: base %v sums %v", vm2.git.base, vm2.git.baseSums)
	}

	// Take theirs makes theirs the base: clean again
	vm.GitTakeTheirs.Execute()
	answer(tr, vm)
	if got := tree.Marks().Get(); got[DeckFile] != 0 {
		t.Fatalf("after Take theirs: %v", got)
	}
	// another repository forgets the base, and the marks with it
	vm.dropBase()
	if len(tree.Marks().Get()) != 0 || vm.deletedFromBase() != nil {
		t.Fatalf("no base, no marks: %v", tree.Marks().Get())
	}
	if st, err := vm.GitRepository().Status(context.Background()); err != nil || len(st) != 0 {
		t.Fatalf("status with no base: %v %v", st, err)
	}
}

// bigScene is a page with a base loaded whose folder holds n media files of size
// bytes each (the deck and the logo besides), unchanged since the load.
func bigScene(t *testing.T, host func(*gitTab) Host, n, size int) (*Scene, *ViewModel, *gitTab) {
	t.Helper()
	s, vm, h := gitSceneHost(t, store.NewMemory(), "", "", "", host)
	for i := range n {
		b := make([]byte, size)
		for j := range b {
			b[j] = byte(i*31 + j*7)
		}
		h.forge.files[fmt.Sprintf("talk/media/m%02d.png", i)] = string(b)
	}
	tr := vm.transport.(*fakeTransport)
	vm.GitPane.Set(true)
	vm.GitRepo.Set("o/r")
	vm.GitDir.Set("talk")
	vm.GitToken.Set(planted)
	vm.GitConnect.Execute()
	vm.GitLoad.Execute()
	answer(tr, vm)
	return s, vm, h
}

// A refresh hashes what was written since the last one, not the project: the
// stored file is replaced on every write, so its slice is its version.
func TestStatusHashesOnlyWhatChanged(t *testing.T) {
	_, vm, _ := bigScene(t, func(h *gitTab) Host { return h }, 6, 4096)
	r := vm.repo
	ctx := context.Background()
	status := func() map[string]vcs.State {
		t.Helper()
		st, err := r.Status(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return st
	}
	hashed := func(f func()) int64 {
		before := r.hashed.Load()
		f()
		return r.hashed.Load() - before
	}
	files := 6 + 2 // the media, the deck, the logo
	r.capture()
	if n := hashed(func() { status() }); n > int64(files) {
		t.Fatalf("first refresh hashed %d contents of %d files", n, files)
	}
	// nothing written: the deck, which each capture rebuilds from the editor, is the only one hashed
	r.capture()
	if n := hashed(func() {
		if st := status(); len(st) != 0 {
			t.Fatalf("unchanged: %v", st)
		}
	}); n != 1 {
		t.Fatalf("second refresh, nothing changed: %d hashes, want 1 (the deck)", n)
	}
	// one media file written: it and the deck
	b := bytes.Repeat([]byte{9}, 4096)
	// (the explorer's own refreshes, which refreshFiles starts, are inside the count)
	if n := hashed(func() {
		if err := vm.project().Write("media/m03.png", b, vm.drafts.now()); err != nil {
			t.Fatal(err)
		}
		vm.refreshFiles()
		if st := status(); st["media/m03.png"] != vcs.Modified || len(st) != 1 {
			t.Fatalf("one file changed: %v", st)
		}
	}); n != 2 {
		t.Fatalf("one file changed: %d hashes across the refreshes, want 2 (it and the deck)", n)
	}
	// the cache holds the project's files and no old copy
	r.mu.Lock()
	size := len(r.cache)
	r.mu.Unlock()
	if size > files+1 {
		t.Errorf("the cache holds %d entries for %d files", size, files)
	}
}

// asyncHost is a host that can run a function on the UI's turn from another
// goroutine: what the page does with setTimeout. The test is the event loop.
type asyncHost struct {
	*gitTab
	turns chan func()
}

func (a *asyncHost) PostUI(f func()) { a.turns <- f }

// settle runs the turns the status goroutines posted, until none comes for a while.
func (a *asyncHost) settle() {
	for {
		select {
		case f := <-a.turns:
			f()
		case <-time.After(150 * time.Millisecond):
			return
		}
	}
}

// With a host that can post, the status is asked off the UI's turn and answered on
// it: the marks arrive when the turn runs, and a newer answer wins. Under -race
// this is also the proof that Status reads a snapshot, not the view-model.
func TestStatusIsAskedOffTheUITurn(t *testing.T) {
	a := &asyncHost{turns: make(chan func(), 64)}
	s, vm, _ := bigScene(t, func(h *gitTab) Host { a.gitTab = h; return a }, 3, 2048)
	tree := s.files.tree
	a.settle()
	if len(tree.Marks().Get()) != 0 {
		t.Fatalf("just loaded: %v", tree.Marks().Get())
	}
	for i := range 20 { // edits while answers are in flight
		vm.Source.Set(vm.Source.Get() + fmt.Sprint(i))
	}
	if err := vm.project().Write("media/m01.png", bytes.Repeat([]byte{1}, 2048), vm.drafts.now()); err != nil {
		t.Fatal(err)
	}
	vm.refreshFiles()
	a.settle()
	marks := tree.Marks().Get()
	if marks[DeckFile] != toolkit.FileModified || marks["media/m01.png"] != toolkit.FileModified || marks["media"] != toolkit.FileModified {
		t.Fatalf("marks %v", marks)
	}
	vm.Source.Set(strings.Repeat("x", 3)) // a same-size edit of a changed deck: still modified
	vm.Source.Set(strings.Repeat("y", 3))
	a.settle()
	if tree.Marks().Get()[DeckFile] != toolkit.FileModified {
		t.Fatalf("after same-size edits: %v", tree.Marks().Get())
	}
}

// TestMeasureStatus prints what a status costs on a project with 20 MB of media,
// with the hashes kept and without (every file hashed on each refresh, as before):
//
//	GITREPO_MEASURE=1 go test -run TestMeasureStatus -v -count=1 .
func TestMeasureStatus(t *testing.T) {
	if os.Getenv("GITREPO_MEASURE") == "" {
		t.Skip("a measurement: set GITREPO_MEASURE=1")
	}
	_, vm, _ := bigScene(t, func(h *gitTab) Host { return h }, 10, 2<<20)
	r := vm.repo
	ctx := context.Background()
	after := func() {
		r.capture()
		if _, err := r.Status(ctx); err != nil {
			t.Fatal(err)
		}
	}
	before := func() {
		if _, err := diffStatus(ctx, vm.git.baseSums, vm.captureState().Files, vm.git.theirs); err != nil {
			t.Fatal(err)
		}
	}
	after() // warm: the first refresh after a load hashes the project once
	for name, f := range map[string]func(){"before (every file hashed)": before, "after (hashes kept)": after} {
		const rounds = 20
		allocs := testing.AllocsPerRun(rounds, f)
		start := time.Now()
		for range rounds {
			f()
		}
		t.Logf("%-28s %10.1f µs/refresh  %6.0f allocs/refresh", name, float64(time.Since(start).Nanoseconds())/1000/rounds, allocs)
	}
}

// dropBase empties the hash cache itself: no scene is attached here, so no Status
// runs that could empty it as a side effect (it does, with no base).
func TestDropBaseEmptiesTheCacheItself(t *testing.T) {
	vm := &ViewModel{GitRev: mvvm.NewObservable(0)}
	vm.repo = &projectRepo{vm: vm, cache: map[sliceID]string{}}
	buf := []byte("the previous person's file")
	vm.repo.cache[idOf(buf)] = sumOf(buf)
	vm.git.base = &remote.Base{Repo: "o/r"}
	vm.git.baseSums = map[string]string{"deck.md": sumOf(buf)}

	vm.dropBase()

	vm.repo.mu.Lock()
	n := len(vm.repo.cache)
	vm.repo.mu.Unlock()
	if n != 0 {
		t.Fatalf("%d hashes of the previous buffers are still kept", n)
	}
	if vm.git.base != nil || vm.git.baseSums != nil || vm.GitRev.Get() != 1 {
		t.Fatalf("base %v sums %v rev %d", vm.git.base, vm.git.baseSums, vm.GitRev.Get())
	}
}

// A status with no base keeps nothing: one that was already running when the
// identity changed may have put its hashes back, and this empties them.
func TestStatusWithNoBaseEmptiesTheCache(t *testing.T) {
	vm := &ViewModel{GitRev: mvvm.NewObservable(0)}
	r := &projectRepo{vm: vm, cache: map[sliceID]string{}}
	buf := []byte("a file of someone who is gone")
	r.cache[idOf(buf)] = sumOf(buf)
	r.capture() // no base: an empty snapshot
	st, err := r.Status(context.Background())
	if err != nil || len(st) != 0 {
		t.Fatalf("status %v %v", st, err)
	}
	r.mu.Lock()
	n := len(r.cache)
	r.mu.Unlock()
	if n != 0 {
		t.Fatalf("%d hashes kept for a project that is under no version control", n)
	}
}
