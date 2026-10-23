package playground

import (
	"github.com/go-widgets/mvvm"
)

// Formatter is the Format tab's part of the view-model: the commands that mark
// text, shape paragraphs, build tables, align a slide and make links, and the state
// the caret is in, which the tab's toggles show (Bold is pressed inside **…**).
// Every command is a pure edit of the deck (format_*.go) applied through the
// editor, so one press is one step of Undo; with no editor it is written into the
// source, as the Insert tab's commands are.
type Formatter struct {
	vm *ViewModel
	in *Inserter

	// Cmd are the commands, by the names Commands lists; State the pressed state of
	// the toggles, by the names States lists.
	Cmd   map[string]*mvvm.Command
	State map[string]*mvvm.Observable[bool]

	// HTMLAllowed is whether the deck's HTML policy keeps the <u> tag: true for the
	// playground, which renders with Marp's default allow-list (marp-core 4.4:
	// u, mark, sub, sup, ins, del, kbd… but no style attribute). With it false the
	// Underline command is disabled.
	HTMLAllowed bool
}

// Commands are the names of the Format commands, in the tab's order.
var Commands = []string{
	"bold", "italic", "underline", "strike", "code",
	"h1", "h2", "h3", "bullets", "numbered", "quote", "rule",
	"table", "row-above", "row-below", "row-delete", "col-left", "col-right", "col-delete",
	"colalign-left", "colalign-center", "colalign-right", "table-format",
	"slide-left", "slide-center", "slide-right",
	"link",
}

// States are the names of the toggles' states.
var States = []string{
	"bold", "italic", "underline", "strike", "code",
	"h1", "h2", "h3", "bullets", "numbered", "quote",
	"colalign-left", "colalign-center", "colalign-right",
	"slide-left", "slide-center", "slide-right", "link",
}

// FormatShortcut is a keyboard shortcut of a Format command, written for Ctrl (or
// Cmd on a Mac) held: the key by its physical name (KeyB, Digit1), with Shift and
// Alt when they are.
type FormatShortcut struct {
	Code       string
	Shift, Alt bool
	Command    string
	Text       string // as shown in a tooltip
}

// FormatShortcuts follow the editors whose toolbars the bibliography read: Ctrl+B,
// Ctrl+I, Ctrl+U and Ctrl+K as in Typora and Markdown All in One, Ctrl+E for code and
// Ctrl+Shift+X for strikethrough, Ctrl+Shift+7 and 8 for the lists as GitHub's editor
// has them, and Ctrl+Alt+1..3 for the headings (Ctrl+1..3 are the browser's tabs).
var FormatShortcuts = []FormatShortcut{
	{"KeyB", false, false, "bold", "Ctrl+B"},
	{"KeyI", false, false, "italic", "Ctrl+I"},
	{"KeyU", false, false, "underline", "Ctrl+U"},
	{"KeyX", true, false, "strike", "Ctrl+Shift+X"},
	{"KeyE", false, false, "code", "Ctrl+E"},
	{"KeyK", false, false, "link", "Ctrl+K"},
	{"Digit1", false, true, "h1", "Ctrl+Alt+1"},
	{"Digit2", false, true, "h2", "Ctrl+Alt+2"},
	{"Digit3", false, true, "h3", "Ctrl+Alt+3"},
	{"Digit8", true, false, "bullets", "Ctrl+Shift+8"},
	{"Digit7", true, false, "numbered", "Ctrl+Shift+7"},
	{"Period", true, false, "quote", "Ctrl+Shift+."},
}

// ShortcutText is the shortcut of a command as shown in a tooltip ("" for none).
func ShortcutText(command string) string {
	for _, s := range FormatShortcuts {
		if s.Command == command {
			return s.Text
		}
	}
	return ""
}

// inlineMarks are the commands that toggle an inline mark.
var inlineMarks = map[string]Marker{
	"bold": MarkBold, "italic": MarkItalic, "underline": MarkUnderline, "strike": MarkStrike, "code": MarkCode,
}

func newFormatter(vm *ViewModel) *Formatter {
	f := &Formatter{vm: vm, in: vm.Insert, Cmd: map[string]*mvvm.Command{}, State: map[string]*mvvm.Observable[bool]{}, HTMLAllowed: true}
	for _, n := range States {
		f.State[n] = mvvm.NewObservable(false)
	}
	always := func() bool { return f.editable() }
	inTable := func() bool {
		in, _ := TableState(f.src(), f.caret())
		return f.editable() && in
	}
	add := func(name string, can func() bool, run func()) { f.Cmd[name] = mvvm.NewCommand(run, can) }
	for name, m := range inlineMarks {
		can := always
		if name == "underline" {
			can = func() bool { return f.editable() && f.HTMLAllowed }
		}
		add(name, can, func() {
			f.apply(ToggleInline(f.src(), f.caret(), f.sel(), m), true, "")
		})
	}
	for n := 1; n <= 3; n++ {
		add("h"+string(rune('0'+n)), always, func() {
			c, ok := ToggleHeading(f.src(), f.caret(), f.sel(), n)
			f.apply(c, ok, "A heading goes on a line of the deck's text, not in the front matter or a code block.")
		})
	}
	add("bullets", always, func() {
		c, ok := ToggleList(f.src(), f.caret(), f.sel(), false)
		f.apply(c, ok, "A list goes on a line of the deck's text, not in the front matter or a code block.")
	})
	add("numbered", always, func() {
		c, ok := ToggleList(f.src(), f.caret(), f.sel(), true)
		f.apply(c, ok, "A list goes on a line of the deck's text, not in the front matter or a code block.")
	})
	add("quote", always, func() {
		c, ok := ToggleQuote(f.src(), f.caret(), f.sel())
		f.apply(c, ok, "A quote goes on a line of the deck's text, not in the front matter or a code block.")
	})
	add("rule", always, func() {
		f.apply(InsertRule(f.src(), f.caret(), f.sel()), true, "")
		f.vm.Status.Set("Inserted a horizontal rule (***: three dashes on a line would start a new slide).")
	})
	add("table", always, func() { f.InsertTable(3, 3) })
	tableOp := func(name, msg string, op func(src string, caret Pos) (Change, bool)) {
		add(name, inTable, func() { c, ok := op(f.src(), f.caret()); f.apply(c, ok, msg) })
	}
	notIn := "Put the caret in a table first."
	tableOp("row-above", notIn, func(s string, c Pos) (Change, bool) { return TableAddRow(s, c, false) })
	tableOp("row-below", notIn, func(s string, c Pos) (Change, bool) { return TableAddRow(s, c, true) })
	tableOp("row-delete", "Put the caret in a row of the table's body: the header stays.", TableDeleteRow)
	tableOp("col-left", notIn, func(s string, c Pos) (Change, bool) { return TableAddColumn(s, c, false) })
	tableOp("col-right", notIn, func(s string, c Pos) (Change, bool) { return TableAddColumn(s, c, true) })
	tableOp("col-delete", "A table keeps at least one column.", TableDeleteColumn)
	tableOp("colalign-left", notIn, func(s string, c Pos) (Change, bool) { return TableAlign(s, c, AlignLeft) })
	tableOp("colalign-center", notIn, func(s string, c Pos) (Change, bool) { return TableAlign(s, c, AlignCenter) })
	tableOp("colalign-right", notIn, func(s string, c Pos) (Change, bool) { return TableAlign(s, c, AlignRight) })
	tableOp("table-format", notIn, TableFormat)
	for name, a := range map[string]SlideAlignment{"slide-left": SlideLeft, "slide-center": SlideCenter, "slide-right": SlideRight} {
		add(name, always, func() {
			c := SetSlideAlignment(f.src(), f.caret(), a)
			f.apply(c, c.From != c.To || c.Text != "", "") // left on a left slide: nothing to do
		})
	}
	add("link", always, func() { f.apply(InsertLink(f.src(), f.caret(), f.sel()), true, "") })

	// the toggles follow the caret, the selection, the deck and the file shown
	vm.CaretLine.Subscribe(func(int) { f.Refresh() })
	vm.CaretCol.Subscribe(func(int) { f.Refresh() })
	vm.Source.Subscribe(func(string) { f.Refresh() })
	vm.Insert.Selection.Subscribe(func(Range) { f.Refresh() })
	vm.Open.Subscribe(func(string) { f.Refresh() })
	f.Refresh()
	return f
}

func (f *Formatter) editable() bool { return f.vm.Open.Get() == DeckFile }
func (f *Formatter) src() string    { return f.vm.Source.Get() }
func (f *Formatter) caret() Pos     { return f.in.caret() }
func (f *Formatter) sel() Range     { return f.in.Selection.Get() }

// apply writes a change into the deck, through the editor when there is one; ok
// false means there was nothing to do, and the status says why.
func (f *Formatter) apply(c Change, ok bool, why string) {
	if !ok {
		f.vm.Status.Set(why)
		return
	}
	f.in.apply(c)
	if f.in.Apply == nil && c.Select != nil { // no editor: the selection is the caret's
		f.vm.CaretLine.Set(c.Select.End.Line)
		f.vm.CaretCol.Set(c.Select.End.Col)
	}
	f.Refresh()
}

// InsertTable writes a table of cols columns and rows rows (header included) after
// the caret's line.
func (f *Formatter) InsertTable(cols, rows int) {
	f.apply(InsertTable(f.src(), f.caret(), f.sel(), cols, rows), true, "")
	f.vm.Status.Set("Inserted a table: type in its cells; the table-aligning commands keep it tidy.")
}

// Shortcut runs the command a Ctrl (or Cmd) shortcut stands for, and says whether
// there was one.
func (f *Formatter) Shortcut(code string, shift, alt bool) bool {
	for _, s := range FormatShortcuts {
		if s.Code == code && s.Shift == shift && s.Alt == alt {
			f.Cmd[s.Command].Execute()
			return true
		}
	}
	return false
}

// Refresh sets the toggles' state from the deck at the caret, and tells the
// commands that depend on it.
func (f *Formatter) Refresh() {
	set := func(name string, on bool) {
		if f.State[name] != nil {
			f.State[name].Set(on && f.editable())
		}
	}
	src, caret, sel := f.src(), f.caret(), f.sel()
	for name, m := range inlineMarks {
		set(name, InlineActive(src, caret, sel, m))
	}
	h, bullets, numbered, quote := BlockState(src, caret)
	set("h1", h == 1)
	set("h2", h == 2)
	set("h3", h == 3)
	set("bullets", bullets)
	set("numbered", numbered)
	set("quote", quote)
	_, a := TableState(src, caret)
	set("colalign-left", a == AlignLeft)
	set("colalign-center", a == AlignCenter)
	set("colalign-right", a == AlignRight)
	al := SlideAlignmentAt(src, caret)
	set("slide-left", al == SlideLeft)
	set("slide-center", al == SlideCenter)
	set("slide-right", al == SlideRight)
	set("link", InLink(src, caret))
	for _, c := range f.Cmd {
		c.RaiseCanExecuteChanged()
	}
}
