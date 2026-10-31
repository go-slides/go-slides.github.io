package playground

import (
	"bytes"
	"maps"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/go-slides/go-slides.github.io/playground/versions"
	"github.com/go-widgets/mvvm"
	"github.com/go-widgets/toolkit/textdiff"
)

// The versions of the project (playground/versions), in this browser: kept
// automatically a while after the typing stops (AutoEvery), before each
// whole-project step (history.go), at each export and share, and by name
// (Save version…). The Versions pane lists them; one can be previewed in
// the preview pane, compared with now file by file (a toolkit.DiffView in
// the editor's place) and restored, the restore itself one step of Undo.

// AutoEvery is the least time between two automatic versions.
const AutoEvery = 2 * time.Minute

// Prompter asks the reader for a line of text (window.prompt).
type Prompter interface {
	Prompt(question, value string) (string, bool)
}

// CompareFile is one file of a comparison: its text then and now.
type CompareFile struct {
	Path             string
	Old, New         string
	Added, Removed   int
	Binary           bool
	OldSize, NewSize int
}

// Comparison is a version compared with the project now.
type Comparison struct {
	ID    int
	Name  string // the version's name, over its column
	Title string
	Files []CompareFile // the files that differ, sorted
}

// versionState is the view-model's part of the versions.
type versionState struct {
	v        *versions.Versions // nil: none (no store, or not loaded yet)
	off      bool               // the page keeps no storage
	lastAuto time.Time
	// the Versions pane: its items (newest first) and the versions they are
	shown []versions.Entry
	// previewing is the version the preview shows (nil: the project)
	previewing *ProjectState
}

// initVersions makes the observables and opens the store's versions.
func (vm *ViewModel) initVersions() {
	vm.VersionsPane = mvvm.NewObservable(false)
	vm.VersionItems = mvvm.NewObservableEq[[]string](nil, slices.Equal[[]string])
	vm.VersionSelected = mvvm.NewObservable(-1)
	vm.OnlyNamed = mvvm.NewObservable(false)
	vm.Compare = mvvm.NewObservable[*Comparison](nil)
	vm.CompareFile = mvvm.NewObservable(0)
	// inline first: the comparison takes the editor's column, often narrow
	vm.CompareSideBySide = mvvm.NewObservable(false)
	vm.OnlyNamed.Subscribe(func(bool) { vm.refreshVersions() })
	vm.SaveVersion = mvvm.NewCommand(vm.saveVersion, func() bool { return vm.vers.v != nil })
	pick := func() bool { return vm.selectedVersion() != nil }
	vm.PreviewVersion = mvvm.NewCommand(func() { vm.withSelected(vm.previewVersion) }, pick)
	vm.CompareVersion = mvvm.NewCommand(func() { vm.withSelected(vm.compareVersion) }, pick)
	vm.RestoreVersion = mvvm.NewCommand(func() { vm.withSelected(vm.restoreVersion) }, pick)
	vm.DeleteVersion = mvvm.NewCommand(func() { vm.withSelected(vm.deleteVersion) }, pick)
	vm.BackToNow = mvvm.NewCommand(vm.backToNow, func() bool { return vm.vers.previewing != nil || vm.Compare.Get() != nil })
	vm.VersionSelected.Subscribe(func(int) { vm.raiseVersionCommands() })
	vm.Compare.Subscribe(func(*Comparison) { vm.raiseVersionCommands() })
	if vm.drafts.st == nil {
		vm.vers.off = true
		vm.refreshVersions()
		return
	}
	versions.Open(vm.drafts.st, func(v *versions.Versions, err error) {
		if err != nil {
			// the draft's own save says what the storage refuses
			vm.vers.off = true
			vm.refreshVersions()
			return
		}
		vm.vers.v = v
		v.Changed = vm.refreshVersions
		vm.refreshVersions()
	})
}

func (vm *ViewModel) raiseVersionCommands() {
	for _, c := range []*mvvm.Command{vm.SaveVersion, vm.PreviewVersion, vm.CompareVersion, vm.RestoreVersion, vm.DeleteVersion, vm.BackToNow} {
		c.RaiseCanExecuteChanged()
	}
}

// versionName is how the pane and the messages name a version.
func (vm *ViewModel) versionName(e versions.Entry) string {
	switch e.Kind {
	case versions.Named:
		return e.Name
	case versions.Before:
		return T(KVersionBefore, e.Name)
	case versions.Export:
		return T(KVersionExport, e.Name)
	case versions.Share:
		return T(KVersionShared)
	}
	return T(KVersionAuto)
}

// refreshVersions rebuilds the pane's list.
func (vm *ViewModel) refreshVersions() {
	var items []string
	vm.vers.shown = nil
	if v := vm.vers.v; v != nil {
		for _, e := range v.Entries() {
			if vm.OnlyNamed.Get() && e.Kind != versions.Named {
				continue
			}
			vm.vers.shown = append(vm.vers.shown, e)
			items = append(items, T(KVersionLine, vm.drafts.short(e.Time), vm.versionName(e), sizeText(v.Size(e))))
		}
	}
	if sel := vm.VersionSelected.Get(); sel >= len(items) {
		vm.VersionSelected.Set(len(items) - 1)
	}
	vm.VersionItems.Set(items)
	vm.raiseVersionCommands()
}

// VersionsNote is what the pane says above an empty list ("" when it lists
// versions).
func (vm *ViewModel) VersionsNote() string {
	switch {
	case vm.vers.off:
		return T(KVersionsOff)
	case len(vm.vers.shown) == 0:
		return T(KVersionsEmpty)
	}
	return ""
}

func (vm *ViewModel) selectedVersion() *versions.Entry {
	i := vm.VersionSelected.Get()
	if i < 0 || i >= len(vm.vers.shown) {
		return nil
	}
	return &vm.vers.shown[i]
}

func (vm *ViewModel) withSelected(f func(versions.Entry)) {
	if e := vm.selectedVersion(); e != nil {
		f(*e)
	}
}

// snapshotOf is a project state as a version keeps it.
func snapshotOf(s ProjectState) versions.Snapshot {
	return versions.Snapshot{Files: s.Files, Dirs: s.Dirs, Open: s.Open}
}

// keepVersion keeps s as a version of kind k (name: what it is kept for).
func (vm *ViewModel) keepVersion(s ProjectState, k versions.Kind, name string, done func(versions.Entry)) {
	v := vm.vers.v
	if v == nil {
		return
	}
	v.Save(snapshotOf(s), k, name, vm.drafts.now(), func(e versions.Entry, err error) {
		switch {
		case err == versions.ErrNamed:
			vm.notice(T(KStatusVersionLimit, versions.MaxNamed))
		case err != nil:
			vm.notice(T(KStatusVersionsFailed, err.Error()))
		case done != nil:
			done(e)
		}
	})
}

// autoVersion keeps an automatic version when the last one is old enough:
// the draft's save tick calls it.
func (vm *ViewModel) autoVersion() {
	now := vm.drafts.now()
	if vm.vers.v == nil || now.Sub(vm.vers.lastAuto) < AutoEvery {
		return
	}
	vm.vers.lastAuto = now
	vm.keepVersion(vm.captureState(), versions.Auto, "", nil)
}

// saveVersion is Save version…: the reader names the version.
func (vm *ViewModel) saveVersion() {
	if vm.vers.v == nil {
		return
	}
	name := vm.drafts.when(vm.drafts.now()) // the default: when
	if p, ok := vm.host.(Prompter); ok {
		n, ok := p.Prompt(T(KAskVersionName), name)
		if !ok {
			return
		}
		if n != "" {
			name = n
		}
	}
	if utf8.RuneCountInString(name) > 80 {
		name = string([]rune(name)[:80])
	}
	vm.keepVersion(vm.captureState(), versions.Named, name, func(versions.Entry) {
		vm.notice(T(KStatusVersionSaved, name))
		vm.VersionsPane.Set(true)
	})
}

// loadVersion reads a version's files as a project state.
func (vm *ViewModel) loadVersion(e versions.Entry, done func(ProjectState)) {
	vm.vers.v.Load(e.ID, func(s versions.Snapshot, _ versions.Entry, err error) {
		if err != nil {
			vm.notice(T(KStatusVersionsFailed, err.Error()))
			return
		}
		if _, ok := s.Files[DeckFile]; !ok {
			s.Files[DeckFile] = nil
		}
		done(ProjectState{Files: s.Files, Dirs: s.Dirs, Open: s.Open})
	})
}

// previewVersion shows a version in the preview; the editor keeps the
// project. Back to now, or an edit, shows the project again.
func (vm *ViewModel) previewVersion(e versions.Entry) {
	vm.loadVersion(e, func(s ProjectState) {
		vm.vers.previewing = &s
		vm.raiseVersionCommands()
		vm.MediaVersion.Set(vm.MediaVersion.Get() + 1)
		vm.changed()
		vm.notice(T(KStatusPreviewing, vm.versionName(e)))
	})
}

// Previewing reports whether the preview shows a version.
func (vm *ViewModel) Previewing() bool { return vm.vers.previewing != nil }

// backToNow ends a preview and a comparison.
func (vm *ViewModel) backToNow() {
	vm.Compare.Set(nil)
	if vm.vers.previewing == nil {
		vm.raiseVersionCommands()
		return
	}
	vm.vers.previewing = nil
	vm.raiseVersionCommands()
	vm.MediaVersion.Set(vm.MediaVersion.Get() + 1)
	vm.changed()
}

// compareVersion compares a version with the project now, file by file.
func (vm *ViewModel) compareVersion(e versions.Entry) {
	vm.loadVersion(e, func(then ProjectState) {
		now := vm.captureState()
		vm.CompareFile.Set(0)
		c := Compare(then, now, T(KCompareTitle, vm.versionName(e)), e.ID)
		c.Name = vm.versionName(e)
		vm.Compare.Set(c)
	})
}

// Compare lists the files that differ between then and now.
func Compare(then, now ProjectState, title string, id int) *Comparison {
	paths := map[string]bool{}
	for k := range then.Files {
		paths[k] = true
	}
	for k := range now.Files {
		paths[k] = true
	}
	c := &Comparison{ID: id, Title: title}
	for _, p := range slices.Sorted(maps.Keys(paths)) {
		a, b := then.Files[p], now.Files[p]
		if bytes.Equal(a, b) {
			continue
		}
		f := CompareFile{Path: p, OldSize: len(a), NewSize: len(b)}
		if IsMedia(p) || !utf8.Valid(a) || !utf8.Valid(b) {
			f.Binary = true
		} else {
			f.Old, f.New = string(a), string(b)
			f.Added, f.Removed = textdiff.Stats(textdiff.Lines(textdiff.Split(f.Old), textdiff.Split(f.New)))
		}
		c.Files = append(c.Files, f)
	}
	return c
}

// CompareItems are the comparison's files, for its drop-down.
func (vm *ViewModel) CompareItems() []string {
	c := vm.Compare.Get()
	if c == nil {
		return nil
	}
	out := make([]string, len(c.Files))
	for i, f := range c.Files {
		if f.Binary {
			out[i] = T(KCompareBinary, f.Path, sizeText(f.OldSize), sizeText(f.NewSize))
		} else {
			out[i] = T(KCompareFileLine, f.Path, f.Added, f.Removed)
		}
	}
	return out
}

// restoreVersion puts a version back, as one step of Undo (the project as it
// was is kept as a version first, see recordStep).
func (vm *ViewModel) restoreVersion(e versions.Entry) {
	vm.loadVersion(e, func(s ProjectState) {
		vm.backToNow()
		name := vm.versionName(e)
		vm.projectStep(T(KStepRestoreVersion, name), func() bool {
			vm.applyState(s)
			return true
		})
		vm.notice(T(KStatusVersionRestored, name))
	})
}

// deleteVersion removes a version, after asking.
func (vm *ViewModel) deleteVersion(e versions.Entry) {
	name := vm.versionName(e)
	if c, ok := vm.host.(Confirmer); ok && !c.Confirm(T(KConfirmDeleteVersion, name)) {
		return
	}
	vm.vers.v.Delete(e.ID, func(err error) {
		if err != nil {
			vm.notice(T(KStatusVersionsFailed, err.Error()))
			return
		}
		if c := vm.Compare.Get(); c != nil && c.ID == e.ID {
			vm.Compare.Set(nil)
		}
		vm.notice(T(KStatusVersionDeleted, name))
	})
}
