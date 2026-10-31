package playground

import (
	"github.com/go-icons/seti"
	"github.com/go-widgets/mvvmtk"
	"github.com/go-widgets/painter"
	"github.com/go-widgets/toolkit"
)

// filePanel is the project's explorer, left of the editor: the Project menu,
// the files (a toolkit.FileTree over the view-model's Files) and how much
// of the project's bounds they take.
type filePanel struct {
	box   *toolkit.VBox
	menu  *toolkit.DropDown
	tree  *toolkit.FileTree
	usage *toolkit.Label
}

// filePanelShare is the panel's share of the width the editor (5) and the
// preview (6) share with it: about 210 logical pixels in a 1440-pixel
// window.
const filePanelShare = 3

// setiIcon draws a file's Seti UI icon (go-icons/seti, MIT): the icons VS
// Code shows, through the toolkit's SVG renderer.
func setiIcon(name string, dir, _ bool) func(painter.Painter, toolkit.Rect, toolkit.RGBA) {
	if dir {
		return toolkit.SVGIcon(seti.Folder())
	}
	return toolkit.SVGIcon(seti.Icon(name))
}

// newFilePanel builds the panel and binds it to the view-model; insert puts
// Markdown at the deck's caret (a medium's "Insert into the deck").
func (s *Scene) newFilePanel(insert func(string)) *filePanel {
	vm, sc, inv := s.vm, toolkit.Scaled, s.invalidate
	fp := &filePanel{
		menu:  toolkit.NewDropDown(ProjectItems, 0),
		tree:  toolkit.NewFileTree(treeEntries(vm.Files.Get())),
		usage: toolkit.NewLabel(vm.Usage.Get()),
	}
	fp.usage.Ellipsis = true
	t := fp.tree
	t.Label = T(KFilesTreeLabel)
	t.ShowSizes = true
	t.Icon = setiIcon
	t.Selected().Set(vm.Open.Get())
	t.OnRequest = func(r toolkit.FileRequest) {
		var err error
		switch r.Action {
		case toolkit.FileOpen:
			vm.OpenFile(r.Path)
		case toolkit.FileCreate:
			err = vm.CreateFromPanel(r.Path)
		case toolkit.FileCreateDir:
			err = vm.CreateFolder(r.Path)
		case toolkit.FileRename, toolkit.FileMove:
			err = vm.RenamePath(r.Path, r.To)
		case toolkit.FileDelete:
			err = vm.DeletePath(r.Path)
		}
		vm.fail(err)
		s.dirty = true
	}
	t.MenuExtra = func(p string, dir bool) []toolkit.MenuItem {
		var items []toolkit.MenuItem
		switch {
		case p == "themes":
			for _, b := range BuiltinThemes {
				items = append(items, toolkit.MenuItem{Label: T(KFilesNewThemeFrom, b), Action: func() { vm.NewTheme(b) }})
			}
		case p == "kinds":
			items = append(items, toolkit.MenuItem{Label: T(KFilesNewKind), Action: vm.NewKind})
		case p == "":
			items = append(items, toolkit.MenuItem{Label: T(KFilesNewData), Action: vm.NewDataFile})
		case IsMedia(p):
			items = append(items, toolkit.MenuItem{Label: T(KFilesInsertIntoDeck), Action: func() { insert(MediaMarkdown(p)) }})
		}
		if !dir {
			items = append(items, toolkit.MenuItem{Label: T(KFilesDownload), Action: func() { vm.DownloadFile(p) }})
		}
		return items
	}
	s.unbind = append(s.unbind,
		vm.Files.Subscribe(func(fs []ProjectFile) { t.Entries().Set(treeEntries(fs)); s.dirty = true }),
		vm.Open.Subscribe(func(p string) { t.Selected().Set(p); s.dirty = true }),
		vm.CreateAsked.Subscribe(func(a CreateAsk) {
			t.BeginCreate(t.Selected().Get(), a.Folder)
			t.SetFocused(true)
			s.editor.Focused().Set(false)
			s.dirty = true
		}),
		mvvmtk.BindSelectedIndex(fp.menu, vm.ProjectAction, inv),
		mvvmtk.BindLabel(fp.usage, vm.Usage, inv),
	)
	fp.box = toolkit.NewVBox()
	fp.box.Spacing = sc(4)
	fp.box.AddFixed(fp.menu, sc(row-6))
	fp.box.AddFlex(t, 1)
	fp.box.AddFixed(fp.usage, sc(18))
	return fp
}

// treeEntries are the view-model's files as the toolkit's.
func treeEntries(fs []ProjectFile) []toolkit.FileEntry {
	out := make([]toolkit.FileEntry, len(fs))
	for i, f := range fs {
		out[i] = toolkit.FileEntry{Path: f.Path, Dir: f.Dir, Size: f.Size}
	}
	return out
}

// bindEditor shows a file in the editor: the deck, with its colours and its
// completion, or another file with its language's colours. The caret goes
// to the top of a file that was not shown.
func (s *Scene) bindEditor(p string) {
	if s.unbindEditor != nil {
		s.unbindEditor()
	}
	if p == DeckFile {
		s.editor.Syntax = deckSyntax{}
		s.editor.CompletionSource = s.deckCompletion
		s.unbindEditor = mvvmtk.BindTextView(s.editor.TextView, s.vm.Source, s.invalidate)
	} else {
		s.editor.Syntax = fileSyntax(p)
		s.editor.CompletionSource = nil
		s.unbindEditor = mvvmtk.BindTextView(s.editor.TextView, s.vm.FileText, s.invalidate)
	}
	if p != s.shown {
		s.shown = p
		s.editor.SetSelection(toolkit.Selection{})
		s.editor.CursorLine().Set(0)
		s.editor.CursorCol().Set(0)
		s.editor.ScrollToMatch(toolkit.Selection{})
	}
	s.markProblems()
	s.dirty = true
}

// FileDropDir is the folder a file dropped at device pixel (x, y) goes in,
// and whether (x, y) is over the file panel at all (else it is the editor's).
func (s *Scene) FileDropDir(x, y int) (string, bool) {
	t := s.files.tree
	r := t.Bounds()
	if !r.Contains(x, y) {
		return "", false
	}
	return t.DropDirAt(x-r.X, y-r.Y), true
}

// FileDragOver outlines the folder files dragged from outside would go in
// (over is false when the drag left the page or ended).
func (s *Scene) FileDragOver(x, y int, over bool) {
	t := s.files.tree
	r := t.Bounds()
	if over && r.Contains(x, y) {
		t.OnEvent(toolkit.Event{Kind: toolkit.EventDragMove, X: x - r.X, Y: y - r.Y})
	} else {
		t.OnEvent(toolkit.Event{Kind: toolkit.EventDragLeave})
	}
	s.dirty = true
}
