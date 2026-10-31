package playground

import "github.com/go-widgets/toolkit"

// Shortcut is Ctrl (or Cmd) held with a key, typed while the editor has the
// keyboard: the Format tab's shortcuts (Ctrl+B, Ctrl+I, Ctrl+U, Ctrl+E, Ctrl+K,
// Ctrl+Shift+X, …). code is the key's physical name (KeyB, Digit1). It reports
// whether it was one, so that the host keeps the browser's own for the others.
//
// Undo (Ctrl/Cmd+Z) and Redo (Shift+Ctrl/Cmd+Z, Ctrl+Y) are taken wherever the
// keyboard is: they also take back an operation on the whole project.
func (s *Scene) Shortcut(code string, shift, alt bool) bool {
	if undo, redo := isUndoKey(code, shift); (undo || redo) && !alt {
		if undo {
			s.vm.Undo.Execute()
		} else {
			s.vm.Redo.Execute()
		}
		s.dirty = true
		return true
	}
	if !s.editing() || !s.vm.Format.Shortcut(code, shift, alt) {
		return false
	}
	s.dirty = true
	return true
}

// applyChange makes an Insert tab edit in the editor: it selects what the change
// replaces and pastes the new text over it, which puts the edit in the editor's
// undo history, then puts the caret where the change says. The editor takes the
// keyboard, so that the person goes on typing where the text went in.
func (s *Scene) applyChange(c Change) {
	s.vm.OpenFile(DeckFile) // the deck, not a project file the panel opened
	s.files.tree.SetFocused(false)
	s.ribbon.SetFocused(false)
	s.editor.Focused().Set(true)
	// an insertion goes where the change says, not where the caret happens to be
	s.editor.CursorLine().Set(c.From.Line)
	s.editor.CursorCol().Set(c.From.Col)
	s.editor.SetSelection(toolkit.Selection{StartLine: c.From.Line, StartCol: c.From.Col, EndLine: c.To.Line, EndCol: c.To.Col})
	s.editor.Paste(c.Text)
	at := toolkit.Selection{StartLine: c.Caret.Line, StartCol: c.Caret.Col, EndLine: c.Caret.Line, EndCol: c.Caret.Col}
	if c.Select != nil { // the text the change marked stays selected
		at = toolkit.Selection{StartLine: c.Select.Start.Line, StartCol: c.Select.Start.Col, EndLine: c.Select.End.Line, EndCol: c.Select.End.Col}
	}
	s.editor.SetSelection(at)
	s.editor.CursorLine().Set(c.Caret.Line)
	s.editor.CursorCol().Set(c.Caret.Col)
	s.editor.ScrollToMatch(at)
	s.dirty = true
}
