package playground

import (
	"context"
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-slides/go-slides.github.io/playground/store"
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
	var r vcs.Repository = projectRepo{}
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
