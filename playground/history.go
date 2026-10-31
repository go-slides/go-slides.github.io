package playground

import (
	"bytes"
	"maps"
	"slices"
	"strings"

	"github.com/go-widgets/mvvm"
	"github.com/go-widgets/toolkit"
)

// Undo and Redo, in the ribbon and on Ctrl/Cmd+Z, Shift+Ctrl/Cmd+Z and
// Ctrl+Y. Two histories answer them, in this order:
//
//   - the editor's, one per file: what was typed, pasted, inserted or
//     formatted in the file shown. Each file keeps its own while another is
//     shown (toolkit.TextHistory), so Undo never puts one file's text into
//     another;
//   - the project's (a toolkit.UndoStack): the operations on the whole
//     project — an example picked, Reset, a .zip or a folder opened, a draft
//     restored, a file or a folder deleted or renamed, a version restored.
//     Each is one step, undone whole: the files, the folders and the file
//     shown come back as they were before it.
//
// Undo takes the editor's last step while there is one, then the project's;
// Redo the same. A project step clears the files' histories (their steps
// were taken on files it replaced); undoing it brings them back, and Redo
// puts back everything as it was when Undo was pressed, edits made in other
// files since included, so neither loses work.
//
// Why undo rather than ask: these operations already keep what they replace
// (Draft ▸ Restore replaced, and a version is saved before each, see
// versions.go), and a question asked before every Reset or example is a
// question people learn to dismiss. The questions that were there (a .zip
// or a folder replacing the deck, Reset) stay: they come from outside the
// page or throw away a deck the reader may not know is kept.

// EditorHistory is the editor as the undo needs it: the scene's CodeEditor.
type EditorHistory interface {
	CanUndo() bool
	CanRedo() bool
	Undo()
	Redo()
	History() toolkit.TextHistory
	SetHistory(toolkit.TextHistory)
}

// ProjectState is the whole project at one moment: every file (the deck,
// DeckFile, included), the folders made in the panel and the file shown.
// Undo puts one back, and a version keeps one.
type ProjectState struct {
	Files map[string][]byte
	Dirs  []string
	Open  string
}

// Size is the bytes the files take.
func (s ProjectState) Size() int {
	n := 0
	for _, b := range s.Files {
		n += len(b)
	}
	return n
}

// Same reports whether two states hold the same files and folders (the
// file shown aside).
func (s ProjectState) Same(o ProjectState) bool {
	return maps.EqualFunc(s.Files, o.Files, bytes.Equal) && slices.Equal(s.Dirs, o.Dirs)
}

// undoState is the view-model's part of the undo.
type undoState struct {
	project *toolkit.UndoStack
	editor  EditorHistory                  // nil until a scene binds one
	files   map[string]toolkit.TextHistory // the histories of the files not shown
	// applying is set while a state is put back: the observables it sets
	// (the example) must not run their own operations again
	applying bool
}

// initUndo makes the commands; NewViewModel calls it.
func (vm *ViewModel) initUndo() {
	vm.undo = undoState{project: toolkit.NewUndoStack(50), files: map[string]toolkit.TextHistory{}}
	vm.Undo = mvvm.NewCommand(vm.undoStep, vm.CanUndo)
	vm.Redo = mvvm.NewCommand(vm.redoStep, vm.CanRedo)
	vm.UndoLabel = mvvm.NewObservable("")
	vm.RedoLabel = mvvm.NewObservable("")
	vm.undo.project.Version().Subscribe(func(int) { vm.refreshUndo() })
	vm.refreshUndo()
}

// BindEditor gives the undo the editor (the scene calls it once).
func (vm *ViewModel) BindEditor(e EditorHistory) {
	vm.undo.editor = e
	vm.refreshUndo()
}

// EditorSwitched keeps the history of the file the editor showed and gives
// it the one of the file it shows now (the scene calls it when it binds
// another file).
func (vm *ViewModel) EditorSwitched(from, to string) {
	e := vm.undo.editor
	if e == nil || from == to {
		return
	}
	vm.undo.files[from] = e.History()
	e.SetHistory(vm.undo.files[to])
	delete(vm.undo.files, to)
	vm.refreshUndo()
}

// EditorHistoryChanged is the editor's word that its history changed; edit
// says a new step was recorded, which drops the project's redo steps.
func (vm *ViewModel) EditorHistoryChanged(edit bool) {
	if edit {
		vm.undo.project.DropRedo()
	}
	vm.refreshUndo()
}

// CanUndo reports whether Undo would do something.
func (vm *ViewModel) CanUndo() bool {
	return (vm.undo.editor != nil && vm.undo.editor.CanUndo()) || vm.undo.project.CanUndo()
}

// CanRedo reports whether Redo would do something.
func (vm *ViewModel) CanRedo() bool {
	return (vm.undo.editor != nil && vm.undo.editor.CanRedo()) || vm.undo.project.CanRedo()
}

func (vm *ViewModel) undoStep() {
	if e := vm.undo.editor; e != nil && e.CanUndo() {
		e.Undo()
	} else if vm.undo.project.Undo() {
		vm.notice(T(KStatusUndone, vm.undo.project.RedoLabel()))
	}
	vm.refreshUndo()
}

func (vm *ViewModel) redoStep() {
	if e := vm.undo.editor; e != nil && e.CanRedo() {
		e.Redo()
	} else if vm.undo.project.Redo() {
		vm.notice(T(KStatusRedone, vm.undo.project.UndoLabel()))
	}
	vm.refreshUndo()
}

// refreshUndo updates what the Undo and Redo controls show: enabled or not,
// and the step they would take ("" for an edit in the editor).
func (vm *ViewModel) refreshUndo() {
	if vm.Undo == nil {
		return
	}
	u, r := "", ""
	if e := vm.undo.editor; e == nil || !e.CanUndo() {
		u = vm.undo.project.UndoLabel()
	}
	if e := vm.undo.editor; e == nil || !e.CanRedo() {
		r = vm.undo.project.RedoLabel()
	}
	vm.UndoLabel.Set(u)
	vm.RedoLabel.Set(r)
	vm.Undo.RaiseCanExecuteChanged()
	vm.Redo.RaiseCanExecuteChanged()
}

// captureState is the project as it is now.
func (vm *ViewModel) captureState() ProjectState {
	p := vm.project()
	s := ProjectState{Files: map[string][]byte{DeckFile: []byte(vm.Source.Get())}, Open: vm.Open.Get()}
	for _, k := range p.Paths() {
		if k != DeckFile {
			s.Files[k], _ = p.Read(k)
		}
	}
	for d := range vm.dirs {
		s.Dirs = append(s.Dirs, d)
	}
	slices.Sort(s.Dirs)
	return s
}

// applyState makes the project s: its files written, the others removed, its
// folders, its deck in the editor and the file it showed open. The draft
// saves it as it saves an edit.
func (vm *ViewModel) applyState(s ProjectState) {
	vm.undo.applying = true
	defer func() { vm.undo.applying = false }()
	p, now := vm.project(), vm.drafts.now()
	for _, k := range p.Paths() {
		if _, ok := s.Files[k]; !ok && k != DeckFile {
			p.Remove(k)
		}
	}
	for k, b := range s.Files {
		if k == DeckFile {
			continue
		}
		if cur, ok := p.Read(k); !ok || !bytes.Equal(cur, b) {
			if vm.fail(p.Write(k, b, now)) {
				return
			}
		}
	}
	vm.dirs = map[string]bool{}
	for _, d := range s.Dirs {
		vm.dirs[d] = true
	}
	vm.saveDirs()
	vm.Source.Set(string(s.Files[DeckFile]))
	if _, ok := s.Files[s.Open]; ok && s.Open != DeckFile {
		vm.OpenFile(s.Open)
	} else {
		vm.Open.Set(DeckFile)
	}
	vm.filesChanged()
}

// takeHistories is every file's history, the shown file's included.
func (vm *ViewModel) takeHistories() map[string]toolkit.TextHistory {
	h := maps.Clone(vm.undo.files)
	if e := vm.undo.editor; e != nil {
		h[vm.Open.Get()] = e.History()
	}
	return h
}

// putHistories gives every file its history back.
func (vm *ViewModel) putHistories(h map[string]toolkit.TextHistory) {
	h = maps.Clone(h)
	if h == nil {
		h = map[string]toolkit.TextHistory{}
	}
	if e := vm.undo.editor; e != nil {
		e.SetHistory(h[vm.Open.Get()])
	}
	delete(h, vm.Open.Get())
	vm.undo.files = h
}

// projectStep runs op, an operation on the whole project, and records it as
// one step named label when it changed the project. op reports false when
// it did nothing (a question answered no).
func (vm *ViewModel) projectStep(label string, op func() bool) {
	if vm.undo.applying || vm.undo.project.Busy() {
		op()
		return
	}
	before, hist := vm.captureState(), vm.takeHistories()
	if op() {
		vm.recordStep(label, before, hist)
	}
}

// recordStep records the step from before (and its files' histories) to the
// project as it is now. Nothing is recorded when nothing changed.
func (vm *ViewModel) recordStep(label string, before ProjectState, hist map[string]toolkit.TextHistory) {
	if vm.captureState().Same(before) {
		return
	}
	for k, h := range hist {
		hist[k] = h.WithoutRedo() // a new step drops what could be redone
	}
	var redoState ProjectState
	var redoHist map[string]toolkit.TextHistory
	vm.undo.project.Push(toolkit.UndoCommand{
		Label: label,
		Undo: func() {
			redoState, redoHist = vm.captureState(), vm.takeHistories()
			vm.applyState(before)
			vm.putHistories(hist)
		},
		Redo: func() {
			hist = vm.takeHistories()
			vm.applyState(redoState)
			vm.putHistories(redoHist)
		},
	})
	// the files' steps were taken on the files the operation replaced
	vm.putHistories(nil)
	vm.refreshUndo()
}

// UndoTip is the Undo control's tooltip: what it would undo, and the
// shortcut.
func UndoTip(label string) string {
	if label == "" {
		return T(KTipUndo)
	}
	return T(KTipUndoStep, label)
}

// RedoTip is the Redo control's tooltip.
func RedoTip(label string) string {
	if label == "" {
		return T(KTipRedo)
	}
	return T(KTipRedoStep, label)
}

// The shortcuts the tooltips show (Cmd on a Mac; Ctrl+Y is Redo too).
const (
	UndoShortcut = "Ctrl+Z"
	RedoShortcut = "Ctrl+Shift+Z"
)

// isUndoKey reports whether a Ctrl/Cmd key is Undo (z) or Redo (Shift+z, y).
func isUndoKey(code string, shift bool) (undo, redo bool) {
	switch strings.TrimPrefix(code, "Key") {
	case "Z":
		return !shift, shift
	case "Y":
		return false, !shift
	}
	return false, false
}
