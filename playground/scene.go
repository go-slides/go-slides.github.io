package playground

import (
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/go-opentype/fonts/jetbrainsmono"
	"github.com/go-slides/slides/editor"
	"github.com/go-widgets/mvvmtk"
	"github.com/go-widgets/painter"
	"github.com/go-widgets/toolkit"
)

// baseFontPx is the logical text size; SetupText multiplies it by the device
// pixel ratio.
const baseFontPx = 14

// SetupText installs the toolkit's metric scale and an anti-aliased face at
// that scale. The wasm shell calls it with window.devicePixelRatio before
// NewScene: one logical pixel is one CSS pixel, drawn at device resolution.
func SetupText(scale float64) {
	if scale <= 0 {
		scale = 1
	}
	toolkit.SetMetricScale(scale)
	_ = toolkit.UseOpenTypeTextSize(baseFontPx)
}

// Palette is the handful of colours the landing's stylesheet defines; the
// host reads them from its custom properties, so the canvas wears the page's
// light or dark theme and its brand accent.
type Palette struct {
	Background, Surface, Ink, Line, Border, Accent toolkit.RGBA
}

// LightPalette and DarkPalette are the landing's tokens with go-slides'
// indigo, the fallback when the host passes none and what tests draw with.
var (
	LightPalette = Palette{
		Background: toolkit.RGB(0xff, 0xff, 0xff), Surface: toolkit.RGB(0xf7, 0xf8, 0xfb),
		Ink: toolkit.RGB(0x0f, 0x11, 0x15), Line: toolkit.RGB(0xe7, 0xe9, 0xee),
		Border: toolkit.RGB(0xcf, 0xd4, 0xdd), Accent: toolkit.RGB(0x4f, 0x46, 0xe5),
	}
	DarkPalette = Palette{
		Background: toolkit.RGB(0x0b, 0x0e, 0x14), Surface: toolkit.RGB(0x14, 0x1a, 0x24),
		Ink: toolkit.RGB(0xe6, 0xed, 0xf3), Line: toolkit.RGB(0x23, 0x2a, 0x36),
		Border: toolkit.RGB(0x3a, 0x45, 0x56), Accent: toolkit.RGB(0x9e, 0x99, 0xf1),
	}
)

// Theme is the toolkit theme the palette paints with.
func (p Palette) Theme() *toolkit.Theme {
	return &toolkit.Theme{
		Background: p.Background, Surface: p.Surface, SurfaceAlt: p.Line,
		OnBackground: p.Ink, OnSurface: p.Ink, Accent: p.Accent, Border: p.Border,
	}
}

// Scene is the View: the toolkit widget tree bound to a ViewModel, plus the
// bookkeeping a canvas application needs. It implements go-widgets/
// webcanvas's App, Resizer, Scroller and Animator, so the wasm shell is one
// webcanvas.Run call. The preview itself is not drawn here: it is an iframe
// the host places over PreviewRect.
type Scene struct {
	vm    *ViewModel
	w, h  int
	theme *toolkit.Theme

	root *toolkit.PopoverHost
	body *toolkit.VBox

	examples, mode, themes   *toolkit.DropDown
	reset, share, download   *toolkit.Button
	present                  *toolkit.Button
	minimapT, thumbsT        *toolkit.ToggleButton
	thumbs                   *stripSlot
	status                   *toolkit.Label
	editor                   *toolkit.CodeEditor
	problems, help           *toolkit.TextView
	preview                  *toolkit.Label
	links                    []*toolkit.Link
	problemsTitle, helpTitle *toolkit.Label

	pressed       toolkit.Widget
	pressedBounds toolkit.Rect

	dirty  bool
	unbind []func()
}

// row is the height of one control row, in logical pixels.
const row = 32

// NewScene builds the widget tree for a w×h device-pixel surface and binds it
// to vm. Call SetupText first.
func NewScene(vm *ViewModel, host Host, w, h int) *Scene {
	s := &Scene{vm: vm, w: w, h: h, theme: LightPalette.Theme()}
	inv := s.invalidate
	sc := toolkit.Scaled
	mono, _ := toolkit.NewTrueTypeFont(jetbrainsmono.TTF, sc(baseFontPx-1))

	s.examples = toolkit.NewDropDown(ExampleNames, vm.Example.Get())
	s.mode = toolkit.NewDropDown(ModeNames, vm.Mode.Get())
	s.themes = toolkit.NewDropDown(vm.Themes, vm.Theme.Get())
	s.reset = toolkit.NewButton("Reset", nil)
	s.share = toolkit.NewButton("Share link", nil)
	s.download = toolkit.NewButton("Download HTML", nil)
	s.present = toolkit.NewButton("Present", nil)
	s.status = toolkit.NewLabel(vm.Status.Get())
	s.status.Ellipsis = true
	s.unbind = append(s.unbind,
		mvvmtk.BindSelectedIndex(s.examples, vm.Example, inv),
		mvvmtk.BindSelectedIndex(s.mode, vm.Mode, inv),
		mvvmtk.BindSelectedIndex(s.themes, vm.Theme, inv),
		// the deck's own themes join the choices
		vm.ThemesVersion.Subscribe(func(int) { s.themes.Options = vm.Themes; s.dirty = true }),
		mvvmtk.BindCommand(s.reset, vm.Reset, inv),
		mvvmtk.BindCommand(s.share, vm.Share, inv),
		mvvmtk.BindCommand(s.download, vm.Download, inv),
		mvvmtk.BindCommand(s.present, vm.Present, inv),
		mvvmtk.BindLabel(s.status, vm.Status, inv),
	)

	// The editor: syntax colours, completion from go-slides' editor
	// package, the caret followed by the help and the preview.
	s.editor = toolkit.NewCodeEditor(vm.Source.Get())
	if mono != nil {
		s.editor.SetFont(mono)
	}
	s.editor.Syntax = deckSyntax{}
	s.editor.CompletionSource = func(_ []string, line, col int) []toolkit.CompletionItem {
		return completionItems(vm.Complete(line, col))
	}
	s.editor.MatchColor = toolkit.RGBA{R: 0xf8, G: 0x51, B: 0x49, A: 0x55}
	s.unbind = append(s.unbind,
		mvvmtk.BindTextView(s.editor.TextView, vm.Source, inv),
	)
	s.editor.CursorLine().Subscribe(func(l int) { vm.CaretLine.Set(l); vm.followCaret(); s.dirty = true })
	s.editor.CursorCol().Subscribe(func(c int) { vm.CaretCol.Set(c); s.dirty = true })
	vm.CaretLine.Subscribe(func(l int) {
		if s.editor.CursorLine().Get() != l {
			s.GoToLine(l)
		}
	})
	vm.DiagVersion.Subscribe(func(int) { s.markProblems(); s.dirty = true })

	s.problemsTitle = toolkit.NewLabel("Problems")
	s.problems = toolkit.NewTextView(vm.Problems.Get())
	if mono != nil {
		s.problems.SetFont(mono)
	}
	s.unbind = append(s.unbind, vm.Problems.Subscribe(func(string) { s.wrapProblems() }))

	s.preview = toolkit.NewLabel("")

	s.helpTitle = toolkit.NewLabel("Help").SetFontSize(sc(baseFontPx + 2))
	// The help pane wraps its text to its width (a TextView does not wrap):
	// it is rewritten when the help or the layout changes.
	s.help = toolkit.NewTextView(vm.Help.Get())
	s.unbind = append(s.unbind, vm.Help.Subscribe(func(string) { s.wrapHelp() }))
	for _, l := range DocLinks {
		url := l.URL
		s.links = append(s.links, toolkit.NewLink(l.Title+" →", func() { host.Open(url) }))
	}

	toolbar := toolkit.NewHBox()
	toolbar.Spacing = sc(8)
	toolbar.AddFixed(toolkit.NewLabel("Example"), sc(62))
	toolbar.AddFixed(s.examples, sc(220))
	toolbar.AddFixed(toolkit.NewLabel("Preview"), sc(58))
	toolbar.AddFixed(s.mode, sc(120))
	toolbar.AddFixed(toolkit.NewLabel("Theme"), sc(50))
	toolbar.AddFixed(s.themes, sc(110))
	for _, b := range []*toolkit.Button{s.present, s.reset, s.share, s.download} {
		toolbar.AddFixed(b, b.PreferredWidth()+sc(10))
	}
	toolbar.AddFlex(s.status, 1)

	// The editor's options: the minimap, and the column of slide thumbnails.
	s.minimapT = toolkit.NewToggleButton("Minimap", vm.Minimap.Get())
	s.thumbsT = toolkit.NewToggleButton("Slides", vm.Thumbs.Get())
	s.editor.ShowMinimap = vm.Minimap.Get()
	s.thumbs = newStripSlot(vm, s)
	s.unbind = append(s.unbind,
		mvvmtk.BindToggle(s.minimapT, vm.Minimap, inv),
		mvvmtk.BindToggle(s.thumbsT, vm.Thumbs, inv),
		vm.Minimap.Subscribe(func(on bool) { s.editor.ShowMinimap = on; s.dirty = true }),
		vm.Thumbs.Subscribe(func(bool) { s.layout(); s.dirty = true }),
	)
	options := toolkit.NewHBox()
	options.Spacing = sc(8)
	options.AddFixed(s.thumbsT, sc(72))
	options.AddFixed(s.minimapT, sc(84))
	options.AddFlex(toolkit.NewLabel(""), 1)

	left := toolkit.NewVBox()
	left.Spacing = sc(4)
	left.AddFixed(options, sc(row-6))
	left.AddFlex(s.editor, 1)
	left.AddFixed(s.problemsTitle, sc(row-8))
	left.AddFixed(s.problems, sc(84))

	right := toolkit.NewVBox()
	right.Spacing = sc(6)
	right.AddFixed(s.helpTitle, sc(row))
	right.AddFlex(s.help, 1)
	right.AddFixed(toolkit.NewLabel("Documentation"), sc(row-8))
	for _, l := range s.links {
		right.AddFixed(l, sc(22))
	}

	main := toolkit.NewHBox()
	main.Spacing = sc(12)
	main.AddFlex(left, 5)
	main.AddNatural(s.thumbs)
	main.AddFlex(s.preview, 6)
	main.AddFixed(right, sc(280))

	s.body = toolkit.NewVBox()
	s.body.Spacing = sc(10)
	s.body.AddFixed(toolbar, sc(row))
	s.body.AddFlex(main, 1)
	s.root = toolkit.NewPopoverHost(s.body)
	s.layout()
	return s
}

// wrapProblems rewrites the problems pane at its width (a monospaced face,
// whose advance is 0.6 of its size).
func (s *Scene) wrapProblems() {
	cols := s.problems.Bounds().W * 100 / max(toolkit.Scaled(baseFontPx-1)*60, 1)
	s.problems.SetText(Wrap(s.vm.Problems.Get(), max(cols-2, 20)))
	s.dirty = true
}

// wrapHelp rewrites the help pane at its width.
func (s *Scene) wrapHelp() {
	// An average glyph of the UI face is a little over half its size.
	cols := s.help.Bounds().W * 100 / max(toolkit.Scaled(baseFontPx)*50, 1)
	s.help.SetText(Wrap(s.vm.Help.Get(), max(cols-2, 20)))
	s.dirty = true
}

// Wrap breaks text into lines of at most cols runes, at spaces, keeping its
// own line breaks and the indentation of a list item's continuation.
func Wrap(text string, cols int) string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		indent := ""
		if t := strings.TrimLeft(line, " "); strings.HasPrefix(t, "- ") {
			indent = strings.Repeat(" ", len(line)-len(t)+2)
		}
		for len([]rune(line)) > cols {
			r := []rune(line)
			cut := strings.LastIndex(string(r[:cols]), " ")
			if cut <= len(indent) {
				cut = len(string(r[:cols]))
			}
			out = append(out, strings.TrimRight(line[:cut], " "))
			line = indent + strings.TrimLeft(line[cut:], " ")
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// completionItems turns the editor package's items into the toolkit's.
func completionItems(items []editor.Item) []toolkit.CompletionItem {
	kinds := map[string]toolkit.CompletionItemKind{
		"directive": toolkit.CompletionProperty, "value": toolkit.CompletionEnum,
		"kind": toolkit.CompletionClass, "theme": toolkit.CompletionConstant,
		"function": toolkit.CompletionFunction, "data": toolkit.CompletionField,
	}
	out := make([]toolkit.CompletionItem, 0, len(items))
	for _, it := range items {
		out = append(out, toolkit.CompletionItem{
			Label: it.Label, Kind: kinds[it.Kind], Detail: it.Detail,
			Documentation: plainMarkdown(it.Doc), InsertText: it.Insert,
		})
	}
	return out
}

// markProblems highlights the diagnostics of the deck in the editor.
func (s *Scene) markProblems() {
	lines := strings.Split(s.vm.Source.Get(), "\n")
	var sels []toolkit.Selection
	for _, d := range s.vm.Diagnostics {
		if d.File != DeckPath || d.Line < 0 || d.Line >= len(lines) {
			continue
		}
		text := lines[d.Line]
		start, end := runeCol(text, d.Column), runeCol(text, d.EndColumn)
		if end <= start {
			start, end = 0, len([]rune(text))
		}
		sels = append(sels, toolkit.Selection{StartLine: d.Line, StartCol: start, EndLine: d.Line, EndCol: end})
	}
	s.editor.SetMatchHighlights(sels)
}

// runeCol converts a UTF-16 column of a line to a rune column.
func runeCol(line string, u16 int) int {
	n, i := 0, 0
	for _, r := range line {
		if n >= u16 {
			return i
		}
		n += utf16.RuneLen(r)
		i++
	}
	return i
}

// GoToLine moves the editor's caret to the start of a line and scrolls it
// into view (a click on a slide in the preview).
func (s *Scene) GoToLine(line int) {
	n := strings.Count(s.editor.Text().Get(), "\n") + 1
	line = min(max(line, 0), max(n-1, 0))
	s.editor.SetSelection(toolkit.Selection{StartLine: line, EndLine: line})
	s.editor.CursorLine().Set(line)
	s.editor.CursorCol().Set(0)
	s.editor.ScrollToMatch(toolkit.Selection{StartLine: line, EndLine: line})
	s.dirty = true
}

func (s *Scene) invalidate() { s.dirty = true }

// layout fits the tree to the surface, with a margin.
func (s *Scene) layout() {
	m := toolkit.Scaled(12)
	s.root.SetBounds(toolkit.Rect{W: s.w, H: s.h})
	s.body.SetBounds(toolkit.Rect{X: m, Y: m, W: s.w - 2*m, H: s.h - 2*m})
	s.wrapHelp()
	s.wrapProblems()
}

// PreviewRect is where the host places the preview iframe, in device pixels.
func (s *Scene) PreviewRect() toolkit.Rect { return s.preview.Bounds() }

// SetPalette recolours the scene.
func (s *Scene) SetPalette(p Palette) {
	s.theme = p.Theme()
	s.dirty = true
}

// Theme is the toolkit theme the scene paints with.
func (s *Scene) Theme() *toolkit.Theme { return s.theme }

// Size is the surface size in device pixels.
func (s *Scene) Size() (int, int) { return s.w, s.h }

// Resize relays the scene out for a canvas of cw×ch CSS pixels and returns
// the device-pixel size it renders at.
func (s *Scene) Resize(cw, ch int) (int, int) {
	scale := toolkit.MetricScale()
	s.w, s.h = max(int(float64(cw)*scale+0.5), 1), max(int(float64(ch)*scale+0.5), 1)
	s.layout()
	return s.w, s.h
}

// Draw paints the whole surface into buf (RGBA, row-major).
func (s *Scene) Draw(buf []byte) {
	bg := s.theme.Background
	for i := 0; i+3 < len(buf); i += 4 {
		buf[i], buf[i+1], buf[i+2], buf[i+3] = bg.R, bg.G, bg.B, 0xff
	}
	s.root.Draw(painter.NewPixelPainter(buf, s.w, s.h), s.theme)
	s.dirty = false
}

// AnimationStep repaints when something outside an input event changed the
// scene: a render that came back, the theme, a paste.
func (s *Scene) AnimationStep(float64) bool { return s.dirty }

// editing reports whether the editor holds the keyboard.
func (s *Scene) editing() bool { return s.editor.Focused().Get() }

// Click is a primary-button press at device pixel (x, y).
func (s *Scene) Click(x, y int) bool {
	s.editor.Focused().Set(s.editor.Bounds().Contains(x, y))
	s.thumbs.SetFocused(s.vm.Thumbs.Get() && s.thumbs.Bounds().Contains(x, y))
	s.root.OnEvent(toolkit.Event{Kind: toolkit.EventClick, X: x, Y: y})
	s.pressed, s.pressedBounds = deepest(s.body, x, y)
	return true
}

// deepest is the leaf widget under (x, y), for drag capture.
func deepest(w toolkit.Widget, x, y int) (toolkit.Widget, toolkit.Rect) {
	type parent interface{ Children() []toolkit.Widget }
	for {
		p, ok := w.(parent)
		if !ok {
			return w, w.Bounds()
		}
		var next toolkit.Widget
		for _, c := range p.Children() {
			if c.Bounds().Contains(x, y) {
				next = c
				break
			}
		}
		if next == nil {
			return w, w.Bounds()
		}
		w = next
	}
}

// Move is pointer motion: a drag to the pressed widget, else hover.
func (s *Scene) Move(x, y int) bool {
	if s.pressed != nil {
		r := s.pressedBounds
		s.pressed.OnEvent(toolkit.Event{Kind: toolkit.EventMouseDrag, X: x - r.X, Y: y - r.Y})
		return true
	}
	s.root.OnEvent(toolkit.Event{Kind: toolkit.EventMouseMove, X: x, Y: y})
	return true
}

// Release ends a press.
func (s *Scene) Release(x, y int) bool {
	if s.pressed == nil {
		return false
	}
	r := s.pressedBounds
	s.pressed.OnEvent(toolkit.Event{Kind: toolkit.EventMouseUp, X: x - r.X, Y: y - r.Y})
	s.pressed = nil
	return true
}

// Context is a secondary click; the playground has no context menu.
func (s *Scene) Context(int, int) bool { return false }

// key delivers a keyboard event to whatever holds the keyboard.
func (s *Scene) key(ev toolkit.Event) {
	if s.editing() {
		s.editor.OnEvent(ev)
		return
	}
	if s.thumbs.Focused() {
		s.thumbs.OnEvent(ev)
		return
	}
	s.root.OnEvent(ev)
}

// Char is typed text.
func (s *Scene) Char(ch string) bool {
	s.key(toolkit.Event{Kind: toolkit.EventChar, Code: ch})
	return true
}

// KeyDown is a named key. A single character here is a shortcut (⌘V arrives
// as "v"): it is left to the browser, which turns copy and paste into the
// events the host forwards to Paste and Copy.
func (s *Scene) KeyDown(code string) bool {
	if len([]rune(code)) == 1 {
		return false
	}
	s.key(toolkit.Event{Kind: toolkit.EventKeyDown, Code: code})
	return true
}

// Scroll is the wheel, in rows.
func (s *Scene) Scroll(x, y, dx, dy int) bool {
	s.root.OnEvent(toolkit.Event{Kind: toolkit.EventScroll, X: x, Y: y, Delta: dy, DeltaX: dx})
	return true
}

// Paste inserts clipboard text into the editor.
func (s *Scene) Paste(text string) bool {
	if !s.editing() {
		return false
	}
	s.editor.Paste(text)
	s.dirty = true
	return true
}

// InsertImage puts an image's Markdown (see ImageMarkdown) at the editor's
// caret, on a line of its own, focusing the editor: a file dropped on the
// page or pasted.
func (s *Scene) InsertImage(markdown string) {
	s.editor.Focused().Set(true)
	s.editor.Paste("\n" + markdown + "\n")
	s.dirty = true
}

// Copy returns the editor's selection, for the clipboard ("" if none).
func (s *Scene) Copy() string {
	if !s.editing() {
		return ""
	}
	return s.editor.SelectionText()
}

// Rects names the on-screen rectangle of each control, in device pixels,
// for a test harness that drives the page with synthetic events.
func (s *Scene) Rects() map[string]toolkit.Rect {
	r := map[string]toolkit.Rect{
		"examples": s.examples.Bounds(), "mode": s.mode.Bounds(), "reset": s.reset.Bounds(),
		"share": s.share.Bounds(), "download": s.download.Bounds(), "status": s.status.Bounds(),
		"present": s.present.Bounds(), "theme": s.themes.Bounds(),
		"editor": s.editor.Bounds(), "problems": s.problems.Bounds(), "preview": s.preview.Bounds(),
		"help": s.help.Bounds(), "thumbs": s.thumbs.Bounds(),
		"minimapToggle": s.minimapT.Bounds(), "thumbsToggle": s.thumbsT.Bounds(),
	}
	for _, p := range toolkit.PopoverOwners(s.root) {
		r["popover"] = p.PopoverBounds()
	}
	return r
}

// OpenPopover is the device-pixel bounds of the open drop-down list, if one
// is open: the host keeps the preview iframe, which lies over the canvas,
// out of its way (PreviewClip).
func (s *Scene) OpenPopover() (toolkit.Rect, bool) {
	for _, p := range toolkit.PopoverOwners(s.root) {
		return p.PopoverBounds(), true
	}
	return toolkit.Rect{}, false
}

// PreviewClip is the CSS clip-path that keeps the preview iframe (frame,
// device pixels) from covering an open drop-down list (pop): the list's
// rectangle cut out of the frame as a notch from its top edge, where a
// toolbar's list overlaps it ("none" when they do not overlap). dpr turns
// device pixels into CSS pixels.
func PreviewClip(frame, pop toolkit.Rect, dpr float64) string {
	x0, x1 := max(pop.X, frame.X), min(pop.X+pop.W, frame.X+frame.W)
	y1 := min(pop.Y+pop.H, frame.Y+frame.H)
	if x0 >= x1 || y1 <= frame.Y || dpr <= 0 {
		return "none"
	}
	// a small margin, so the list's shadow shows
	const margin = 6
	px := func(v int) string { return strconv.FormatFloat(float64(v)/dpr, 'f', 1, 64) + "px" }
	l, r := px(max(x0-frame.X-margin, 0)), px(min(x1-frame.X+margin, frame.W))
	b := px(min(y1-frame.Y+margin, frame.H))
	return "polygon(0 0," + l + " 0," + l + " " + b + "," + r + " " + b + "," + r + " 0,100% 0,100% 100%,0 100%)"
}

// CaretPixel is the device pixel of a caret position in the editor, for a
// test harness that clicks there.
func (s *Scene) CaretPixel(line, col int) (int, int) { return s.editor.CaretPixel(line, col) }

// EditorCursor is the editor's caret (line, rune column).
func (s *Scene) EditorCursor() (int, int) {
	return s.editor.CursorLine().Get(), s.editor.CursorCol().Get()
}

// ProblemHighlights are the ranges the editor shows as problems.
func (s *Scene) ProblemHighlights() []toolkit.Selection { return s.editor.MatchHighlights() }
