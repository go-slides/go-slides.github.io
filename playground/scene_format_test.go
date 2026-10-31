package playground

import (
	"strings"
	"testing"

	"github.com/go-widgets/toolkit"
)

// put makes the deck source, with the caret at (line, col) and, when sel is not
// empty, the text from (line, col) to (line, col+sel) selected.
func put(s *Scene, vm *ViewModel, src string, line, col, sel int) {
	vm.Source.Set(src)
	s.editor.SetSelection(toolkit.Selection{StartLine: line, StartCol: col, EndLine: line, EndCol: col + sel})
	s.editor.CursorLine().Set(line)
	s.editor.CursorCol().Set(col + sel)
}

func formatScene(t *testing.T) (*Scene, *ViewModel) {
	t.Helper()
	s, vm, _, _ := newScene(t)
	s.Resize(1920, 900)
	gotoTab(t, s, "Format")
	return s, vm
}

func TestFormatTabShowsItsControls(t *testing.T) {
	s, _ := formatScene(t)
	for name := range formatLabels {
		if s.Rects()["fmt-"+name].W == 0 {
			t.Errorf("Format tab: no %s among %v", name, s.ribbon.Levels())
		}
	}
}

func TestFormatButtonsEditThroughTheEditorInOneUndoStep(t *testing.T) {
	s, vm := formatScene(t)
	put(s, vm, "say hello now", 0, 4, 5) // "hello" selected
	clickRect(s, "fmt-bold")
	if got := vm.Source.Get(); got != "say **hello** now" {
		t.Fatalf("Bold: %q", got)
	}
	if !vm.Format.State["bold"].Get() {
		t.Error("Bold should show pressed on the text it marked")
	}
	if !s.editing() {
		t.Error("the editor keeps the keyboard")
	}
	// the marked text stays selected: pressing again unmarks it
	clickRect(s, "fmt-bold")
	if got := vm.Source.Get(); got != "say hello now" || vm.Format.State["bold"].Get() {
		t.Fatalf("Bold again: %q pressed=%v", got, vm.Format.State["bold"].Get())
	}
	// one Undo reverts each action
	s.editor.Undo()
	if got := vm.Source.Get(); got != "say **hello** now" {
		t.Fatalf("Undo of the unbold: %q", got)
	}
	s.editor.Undo()
	if got := vm.Source.Get(); got != "say hello now" {
		t.Fatalf("Undo of the bold: %q", got)
	}
}

func TestFormatTogglesFollowTheCaret(t *testing.T) {
	s, vm := formatScene(t)
	put(s, vm, "a **bold** and _it_ and `c`\n\n## Head\n\n- item\n\n> quote", 0, 5, 0)
	if !vm.Format.State["bold"].Get() || vm.Format.State["italic"].Get() {
		t.Error("in the bold word")
	}
	put(s, vm, vm.Source.Get(), 0, 17, 0)
	if !vm.Format.State["italic"].Get() || vm.Format.State["bold"].Get() {
		t.Error("in the italic word")
	}
	put(s, vm, vm.Source.Get(), 0, 25, 0)
	if !vm.Format.State["code"].Get() {
		t.Error("in the code")
	}
	put(s, vm, vm.Source.Get(), 2, 5, 0)
	if !vm.Format.State["h2"].Get() || vm.Format.State["h1"].Get() {
		t.Error("in a heading 2")
	}
	put(s, vm, vm.Source.Get(), 4, 3, 0)
	if !vm.Format.State["bullets"].Get() {
		t.Error("in a list")
	}
	put(s, vm, vm.Source.Get(), 6, 4, 0)
	if !vm.Format.State["quote"].Get() {
		t.Error("in a quote")
	}
	// the ribbon's toggle is that state: drawn pressed (no panic, state shared)
	buf := make([]byte, 1920*900*4)
	s.Draw(buf)
	// a project file shown in the editor is not formatted
	vm.Open.Set("themes/x.css")
	if vm.Format.State["quote"].Get() || vm.Format.Cmd["bold"].CanExecute() {
		t.Error("a theme file is not formatted as Markdown")
	}
}

func TestFormatShortcuts(t *testing.T) {
	s, vm := formatScene(t)
	put(s, vm, "say hello now", 0, 4, 5)
	s.editor.Focused().Set(true)
	for _, c := range []struct {
		code       string
		shift, alt bool
		want       string
	}{
		{"KeyB", false, false, "say **hello** now"},
		{"KeyB", false, false, "say hello now"},
		{"KeyI", false, false, "say _hello_ now"},
		{"KeyI", false, false, "say hello now"},
		{"KeyU", false, false, "say <u>hello</u> now"},
		{"KeyU", false, false, "say hello now"},
		{"KeyX", true, false, "say ~~hello~~ now"},
		{"KeyX", true, false, "say hello now"},
		{"KeyE", false, false, "say `hello` now"},
		{"KeyE", false, false, "say hello now"},
		{"Digit2", false, true, "## say hello now"},
		{"Digit2", false, true, "say hello now"},
		{"Digit8", true, false, "- say hello now"},
		{"Digit7", true, false, "1. say hello now"},
		{"Period", true, false, "> 1. say hello now"},
	} {
		if !s.Shortcut(c.code, c.shift, c.alt) {
			t.Fatalf("%s: not a shortcut", c.code)
		}
		if got := vm.Source.Get(); got != c.want {
			t.Fatalf("%s: %q, want %q", c.code, got, c.want)
		}
	}
	// not a shortcut: left to the browser (Ctrl+Z is Undo now, history_test.go);
	// and not while the editor has no keyboard
	if s.Shortcut("KeyQ", false, false) || s.Shortcut("KeyB", true, true) {
		t.Error("only the Format shortcuts are taken")
	}
	s.editor.Focused().Set(false)
	if s.Shortcut("KeyB", false, false) {
		t.Error("a shortcut with the editor out of use is the page's")
	}
	if ShortcutText("bold") != "Ctrl+B" || ShortcutText("table") != "" {
		t.Error("the tooltips' shortcuts")
	}
}

func TestFormatTableCommandsAreForTables(t *testing.T) {
	s, vm := formatScene(t)
	put(s, vm, "plain text", 0, 2, 0)
	if vm.Format.Cmd["row-below"].CanExecute() || vm.Format.Cmd["colalign-center"].CanExecute() {
		t.Error("table commands are disabled outside a table")
	}
	// the Table control opens a grid of cells: the third column of the second row
	// is a table of 3 columns and 2 rows
	clickRect(s, "fmt-table")
	pop, open := s.Rects()["popover"]
	if !open {
		t.Fatal("the Table control should open its grid")
	}
	cs := (pop.W - 2) / tableGridCols
	x, y := pop.X+1+2*cs+cs/2, pop.Y+1+cs+cs/2
	s.Click(x, y)
	s.Release(x, y)
	if !strings.Contains(vm.Source.Get(), "| Column 1 | Column 2 | Column 3 |") || strings.Contains(vm.Source.Get(), "Column 4") {
		t.Fatalf("table: %q", vm.Source.Get())
	}
	if n := strings.Count(vm.Source.Get(), "\n"); n != 1+1+1+1 { // text, blank, header, delimiter, one body row
		t.Fatalf("3 x 2 is the header and one row: %q", vm.Source.Get())
	}
	// the first header cell is selected for typing over
	if sel := s.editor.Selection().Get(); sel.StartLine != 2 || sel.EndLine != 2 || sel.EndCol-sel.StartCol != len("Column 1") {
		t.Errorf("the header's first cell should be selected: %+v", sel)
	}
	if !vm.Format.Cmd["row-below"].CanExecute() {
		t.Error("in the table the row commands run")
	}
	before := strings.Count(vm.Source.Get(), "\n")
	clickRect(s, "fmt-row-below")
	if strings.Count(vm.Source.Get(), "\n") != before+1 {
		t.Errorf("Row below: %q", vm.Source.Get())
	}
	clickRect(s, "fmt-colalign-center")
	if !strings.Contains(vm.Source.Get(), ":------:") && !strings.Contains(vm.Source.Get(), ":-") {
		t.Errorf("centre alignment: %q", vm.Source.Get())
	}
	if !vm.Format.State["colalign-center"].Get() {
		t.Error("the column's alignment shows pressed")
	}
	// Undo takes the alignment back in one step
	s.editor.Undo()
	if strings.Contains(vm.Source.Get(), ":-") {
		t.Errorf("Undo: %q", vm.Source.Get())
	}
	// outside a table a table command says why
	put(s, vm, "plain", 0, 1, 0)
	vm.Status.Set("")
	vm.Format.Cmd["row-below"].Execute()
	if vm.Source.Get() != "plain" {
		t.Error("a table command outside a table changes nothing")
	}
	// the menu's other sizes
	vm.Format.InsertTable(5, 4)
	if !strings.Contains(vm.Source.Get(), "Column 5") {
		t.Errorf("5 x 4: %q", vm.Source.Get())
	}
}

func TestFormatSlideAlignmentToggles(t *testing.T) {
	s, vm := formatScene(t)
	put(s, vm, "# One\n\ntext\n\n---\n\n# Two", 2, 2, 0)
	clickRect(s, "fmt-slide-center")
	if !strings.Contains(vm.Source.Get(), "<style scoped>\nsection { text-align: center; }\n</style>") {
		t.Fatalf("centre: %q", vm.Source.Get())
	}
	if !vm.Format.State["slide-center"].Get() || vm.Format.State["slide-left"].Get() {
		t.Error("the slide's alignment shows pressed")
	}
	// the other slide is not affected
	put(s, vm, vm.Source.Get(), strings.Count(vm.Source.Get(), "\n"), 0, 0)
	if !vm.Format.State["slide-left"].Get() {
		t.Error("the second slide is left-aligned")
	}
	put(s, vm, vm.Source.Get(), 7, 0, 0)
	clickRect(s, "fmt-slide-center") // again: back to left
	if strings.Contains(vm.Source.Get(), "<style scoped>") {
		t.Errorf("centre again should remove the block: %q", vm.Source.Get())
	}
	// left on a left slide does nothing
	before := vm.Source.Get()
	clickRect(s, "fmt-slide-left")
	if vm.Source.Get() != before {
		t.Error("left on a left slide")
	}
}

func TestFormatLinkSelectsTheURL(t *testing.T) {
	s, vm := formatScene(t)
	put(s, vm, "see docs now", 0, 4, 4)
	clickRect(s, "fmt-link")
	if got := vm.Source.Get(); got != "see [docs](https://) now" {
		t.Fatalf("link: %q", got)
	}
	sel := s.editor.Selection().Get()
	if sel.StartCol != 11 || sel.EndCol != 19 {
		t.Errorf("the url should be selected to type over: %+v", sel)
	}
	if !vm.Format.State["link"].Get() {
		t.Error("Link shows pressed inside a link")
	}
}

func TestFormatWithoutAnEditorWritesTheSource(t *testing.T) {
	vm := NewViewModel(&fakeTransport{}, newHost(), "")
	vm.Source.Set("say hello now")
	vm.CaretLine.Set(0)
	vm.CaretCol.Set(6)
	vm.Format.Cmd["bold"].Execute()
	if got := vm.Source.Get(); got != "say **hello** now" {
		t.Fatalf("no editor: %q", got)
	}
	vm.Format.HTMLAllowed = false
	vm.Format.Refresh()
	if vm.Format.Cmd["underline"].CanExecute() {
		t.Error("underline is disabled when the HTML policy drops <u>")
	}
	vm.Format.HTMLAllowed = true
	vm.Format.Refresh()
	vm.Source.Set("")
	vm.Format.Cmd["link"].Execute()
	if !strings.Contains(vm.Source.Get(), "[text](https://)") {
		t.Fatalf("link: %q", vm.Source.Get())
	}
}
