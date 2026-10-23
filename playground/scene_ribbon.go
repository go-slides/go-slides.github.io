package playground

import "github.com/go-widgets/toolkit"

// applyChange makes an Insert tab edit in the editor: it selects what the change
// replaces and pastes the new text over it, which puts the edit in the editor's
// undo history, then puts the caret where the change says. The editor takes the
// keyboard, so that the person goes on typing where the text went in.
func (s *Scene) applyChange(c Change) {
	s.vm.OpenFile(DeckFile) // the deck, not a project file the panel opened
	s.files.tree.SetFocused(false)
	s.ribbon.SetFocused(false)
	s.editor.Focused().Set(true)
	s.editor.SetSelection(toolkit.Selection{StartLine: c.From.Line, StartCol: c.From.Col, EndLine: c.To.Line, EndCol: c.To.Col})
	s.editor.Paste(c.Text)
	at := toolkit.Selection{StartLine: c.Caret.Line, StartCol: c.Caret.Col, EndLine: c.Caret.Line, EndCol: c.Caret.Col}
	s.editor.SetSelection(at)
	s.editor.CursorLine().Set(c.Caret.Line)
	s.editor.CursorCol().Set(c.Caret.Col)
	s.editor.ScrollToMatch(at)
	s.dirty = true
}
