// Package playground is the go-slides playground: a deck editor beside a live
// preview, in the browser, with no server. It is a go-widgets application
// (scene.go, the View) over a go-widgets/mvvm view-model (this file); the
// slides are rendered by go-slides itself in a Web Worker (render.go), and
// shown in a sandboxed iframe the host page places over the preview pane.
//
// Everything here is tagless and tested natively; cmd/playground-wasm is the
// thin js/wasm shell that wires it to the page.
package playground

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/go-slides/slides/editor"
	"github.com/go-widgets/mvvm"
)

// DeckPath is the path the playground's deck goes by: there is no file, but
// error messages and relative paths need one.
const DeckPath = "/playground/deck.md"

// Mode is what the preview shows.
type Mode int

const (
	// ModeOverview stacks every slide, scrolling with the editor.
	ModeOverview Mode = iota
	// ModeSlideshow is the presentation itself: one slide at a time, the
	// keyboard, transitions, the presenter view (p).
	ModeSlideshow
)

// ModeNames are the modes in Mode order.
var ModeNames = []string{"Overview", "Slideshow"}

// Transport carries render requests to the worker. Send never blocks; the
// answer comes back through ViewModel.Receive.
type Transport interface {
	Send(Request)
}

// Host is what the page does for the view-model.
type Host interface {
	// SetFragment replaces the URL's fragment (no history entry).
	SetFragment(fragment string)
	// Download offers a file to save.
	Download(name, mime, data string)
	// Open opens a URL in a new tab.
	Open(url string)
	// Clipboard copies text.
	Clipboard(text string)
	// OpenPresenter opens (or brings forward) the presenter window.
	OpenPresenter()
	// PatchPreview sends a patch to the preview frame's document (see
	// overviewScript), which replaces only the slides that changed.
	PatchPreview(PreviewPatch)
}

// PreviewPatch turns the overview the frame shows into the next one: the
// slides that changed, by index, and how many there are now.
type PreviewPatch struct {
	Count   int          `json:"count"`
	Changed []SlidePatch `json:"changed"`
}

// SlidePatch is one slide's new markup.
type SlidePatch struct {
	I    int    `json:"i"`
	HTML string `json:"html"`
}

// preview is what the preview frame holds, so that an overview whose
// style sheet and theme are unchanged is patched slide by slide instead of
// loaded again: a 100-slide document re-parsed and laid out on every
// keystroke is most of the time from a change to the preview.
type preview struct {
	// head and parts are the overview the frame has loaded (head == "":
	// none, or a slideshow).
	head  string
	parts []string
	// loading is a document set and not loaded yet; loadingHead/Parts are
	// what it holds. A render that arrives meanwhile waits in wantHead/
	// wantParts and is applied by a patch when the load is reported.
	loading                 bool
	loadingHead, wantHead   string
	loadingParts, wantParts []string
	wanted                  bool
}

// ViewModel holds the playground's state as go-widgets/mvvm observables. It
// references no widget; scene.go binds the widgets to it.
type ViewModel struct {
	Source  *mvvm.Observable[string]
	Example *mvvm.Observable[int]
	Mode    *mvvm.Observable[int]
	// Theme is the index in Themes of the deck's theme. The deck's text is
	// the source of truth: choosing a theme writes it into the front matter
	// (SetTheme), and editing the front matter moves the choice. Themes are
	// the choices (ThemeChoices); ThemesVersion counts their changes.
	Theme         *mvvm.Observable[int]
	Themes        []string
	ThemesVersion *mvvm.Observable[int]
	themeSyncing  bool
	// CaretLine and CaretCol are the editor's caret (0-based; the column
	// counts runes), which the context help and the scroll sync follow.
	CaretLine *mvvm.Observable[int]
	CaretCol  *mvvm.Observable[int]

	// Derived.
	Doc *mvvm.Observable[string] // the preview iframe's document
	// PresenterOpen says whether the presenter window is open: the host
	// sets it when the window opens and closes. PresenterDoc is its page.
	PresenterOpen *mvvm.Observable[bool]
	PresenterDoc  *mvvm.Observable[string]
	Status        *mvvm.Observable[string] // "12 slides · 48 ms"
	Problems      *mvvm.Observable[string] // one line per diagnostic
	Help          *mvvm.Observable[string] // the context help at the caret
	// Diagnostics are the last render's problems, for the editor's
	// highlights; Version counts the renders that changed them.
	Diagnostics []editor.Diagnostic
	DiagVersion *mvvm.Observable[int]

	// Minimap and Thumbs are the editor's minimap and the column of slide
	// thumbnails, each on or off. SlideLines are the source lines the slides
	// start at (SlidesVersion counts their changes); CurrentSlide is the slide
	// the caret is in, or the preview is scrolled to; RailDoc is the thumbnail
	// frame's document (see rail.go).
	Minimap       *mvvm.Observable[bool]
	Thumbs        *mvvm.Observable[bool]
	SlideLines    []int
	holdSlide     time.Time
	SlidesVersion *mvvm.Observable[int]
	CurrentSlide  *mvvm.Observable[int]
	RailDoc       *mvvm.Observable[string]

	// Commands.
	Share    *mvvm.Command
	Download *mvvm.Command
	Reset    *mvvm.Command
	// Present shows the slideshow and opens the presenter view beside it.
	Present *mvvm.Command

	// DraftAction is the Draft menu's selection: 0 its title, another
	// index an action (DraftItems), run when chosen. DraftVersion counts
	// the changes of DraftItems.
	DraftAction  *mvvm.Observable[int]
	DraftVersion *mvvm.Observable[int]

	transport Transport
	host      Host
	drafts    *drafts

	pv preview // what the preview frame holds
	// pendingNotice is a status the next render keeps (see notice).
	pendingNotice string
	// top is where the overview is scrolled, as the frame last reported.
	top      SlideTop
	next     int     // the next request ID
	inFlight int     // the request awaiting an answer, 0 if none
	pending  bool    // the source changed while a request was in flight
	last     Request // the request in flight
}

// NewViewModel builds the view-model on a starting deck: the shared deck of
// the URL's fragment if there is one, else the first example.
func NewViewModel(t Transport, h Host, fragment string) *ViewModel {
	vm := &ViewModel{
		Source:        mvvm.NewObservable(""),
		Example:       mvvm.NewObservable(0),
		Mode:          mvvm.NewObservable(int(ModeOverview)),
		Theme:         mvvm.NewObservable(0),
		Themes:        BuiltinThemes,
		ThemesVersion: mvvm.NewObservable(0),
		CaretLine:     mvvm.NewObservable(0),
		CaretCol:      mvvm.NewObservable(0),
		Doc:           mvvm.NewObservable(""),
		Status:        mvvm.NewObservable("Starting…"),
		Problems:      mvvm.NewObservable(""),
		Help:          mvvm.NewObservable(GeneralHelp),
		DiagVersion:   mvvm.NewObservable(0),
		Minimap:       mvvm.NewObservable(true),
		Thumbs:        mvvm.NewObservable(true),
		SlidesVersion: mvvm.NewObservable(0),
		CurrentSlide:  mvvm.NewObservable(-1),
		RailDoc:       mvvm.NewObservable(""),
		PresenterOpen: mvvm.NewObservable(false),
		PresenterDoc:  mvvm.NewObservable(""),
		DraftAction:   mvvm.NewObservable(0),
		DraftVersion:  mvvm.NewObservable(0),
		transport:     t,
		host:          h,
		top:           SlideTop{I: -1},
		drafts:        &drafts{now: time.Now},
	}
	if p, ok := h.(Projects); ok {
		vm.drafts.st, vm.drafts.ann = p.Projects(), p.Announce
	}
	var shared string
	hasShared, status := false, ""
	if fragment != "" {
		if s, err := DecodeShare(fragment); err == nil {
			shared, hasShared = s, true
			status = "Opened the deck shared in the link."
		} else {
			status = "The link's deck could not be read: " + err.Error()
		}
	}
	src, ds := vm.startDrafts(shared, hasShared)
	if ds != "" {
		status = ds
	}
	if status != "" {
		vm.notice(status)
	}
	vm.Source.Set(src)
	vm.loadUI()
	vm.syncTheme()
	vm.refreshDraftMenu()
	vm.Source.Subscribe(func(string) { vm.drafts.dirty = true; vm.syncTheme(); vm.changed() })
	vm.Theme.Subscribe(vm.chooseTheme)
	vm.Mode.Subscribe(func(int) { vm.saveUI(); vm.changed() })
	vm.DraftAction.Subscribe(vm.chooseDraftAction)
	vm.PresenterOpen.Subscribe(func(open bool) {
		if open {
			vm.changed()
		}
	})
	vm.Example.Subscribe(func(i int) {
		if i >= 0 && i < len(Examples) {
			// the draft stays available (Draft ▸ Restore replaced)
			vm.archiveDraft(Examples[i].Source)
			vm.Source.Set(Examples[i].Source)
		}
	})
	vm.CaretLine.Subscribe(func(int) { vm.updateHelp() })
	vm.CaretCol.Subscribe(func(int) { vm.updateHelp() })

	vm.Share = mvvm.NewCommand(vm.share, nil)
	vm.Download = mvvm.NewCommand(vm.download, nil)
	vm.Reset = mvvm.NewCommand(func() {
		ex := Examples[max(vm.Example.Get(), 0)].Source
		if vm.Source.Get() == ex {
			return
		}
		if c, ok := h.(Confirmer); ok && !c.Confirm("Replace your deck with the example? Your text stays in this browser: Draft ▸ Restore replaced.") {
			return
		}
		vm.archiveDraft(ex)
		vm.Source.Set(ex)
	}, nil)
	vm.Present = mvvm.NewCommand(vm.present, nil)
	vm.changed()
	return vm
}

// changed asks for a render of the current source. One request is in flight
// at most: a change while one is out is sent when its answer comes back, so
// typing never queues renders that are already stale.
func (vm *ViewModel) changed() {
	if vm.inFlight != 0 {
		vm.pending = true
		return
	}
	vm.next++
	vm.inFlight = vm.next
	mode := Mode(vm.Mode.Get())
	// The presenter view follows the slideshow: no slideshow, no page for it.
	presenter := vm.PresenterOpen.Get() && mode == ModeSlideshow
	vm.last = Request{ID: vm.next, Source: vm.Source.Get(), Mode: mode, Presenter: presenter}
	vm.transport.Send(vm.last)
}

// Receive takes a render's answer.
func (vm *ViewModel) Receive(r Response) {
	if r.ID != vm.inFlight {
		return // an answer to a request given up on
	}
	vm.inFlight = 0
	if r.Error != "" {
		vm.Status.Set("⚠ " + r.Error)
	} else {
		switch {
		case r.Parts != nil:
			vm.showOverview(r.Head, r.Parts)
		case r.Doc != "":
			vm.pv = preview{} // a slideshow: nothing to patch
			vm.Doc.Set(r.Doc)
		}
		if r.PresenterDoc != "" {
			vm.PresenterDoc.Set(r.PresenterDoc)
		}
		status := fmt.Sprintf("%d slide%s · rendered in %d ms", r.Slides, plural(r.Slides), r.Millis)
		if r.Doc == "" && r.Parts == nil {
			status = "the preview shows the last deck that rendered"
		}
		if n := len(r.Diagnostics); n > 0 {
			status = fmt.Sprintf("%d problem%s · %s", n, plural(n), status)
		} else if vm.pendingNotice != "" {
			// a notice said as the deck changed outlives the render it caused
			status = vm.pendingNotice
		}
		vm.pendingNotice = ""
		if vm.drafts.failed {
			status += " · ⚠ draft not kept in this browser"
		}
		vm.Status.Set(status)
	}
	vm.Diagnostics = r.Diagnostics
	vm.Problems.Set(formatProblems(r.Diagnostics))
	vm.DiagVersion.Set(vm.DiagVersion.Get() + 1)
	if vm.pending {
		vm.pending = false
		vm.changed()
	}
}

// notice says something that the render about to come must not replace
// (a restored draft, a shared deck): the status keeps it through that
// render.
func (vm *ViewModel) notice(s string) {
	vm.pendingNotice = s
	vm.Status.Set(s)
}

// syncTheme moves the theme picker to the deck's theme ("default" when the
// deck names none), with the deck's own themes among the choices.
func (vm *ViewModel) syncTheme() {
	src := vm.Source.Get()
	choices := ThemeChoices(src)
	name, ok := DeckTheme(src)
	if !ok {
		name = "default"
	}
	vm.themeSyncing = true
	defer func() { vm.themeSyncing = false }()
	if !slices.Equal(choices, vm.Themes) {
		vm.Themes = choices
		vm.ThemesVersion.Set(vm.ThemesVersion.Get() + 1)
	}
	if i := slices.Index(vm.Themes, name); i >= 0 {
		vm.Theme.Set(i)
	}
}

// chooseTheme writes the chosen theme into the deck.
func (vm *ViewModel) chooseTheme(i int) {
	if vm.themeSyncing || i < 0 || i >= len(vm.Themes) {
		return
	}
	src := vm.Source.Get()
	if cur, _ := DeckTheme(src); cur == vm.Themes[i] {
		return
	}
	vm.Source.Set(SetTheme(src, vm.Themes[i]))
}

// showOverview puts an overview in the preview: a patch of the slides that
// changed when the frame holds an overview with the same head, else a new
// document. While a document loads, the latest overview waits for it.
func (vm *ViewModel) showOverview(head string, parts []string) {
	pv := &vm.pv
	vm.setSlides(parts)
	switch {
	case pv.loading:
		pv.wanted, pv.wantHead, pv.wantParts = true, head, parts
	case pv.head != "" && headKey(pv.head) == headKey(head):
		if p, ok := diffParts(pv.parts, parts); ok {
			vm.host.PatchPreview(p)
		}
		pv.parts = parts
	default:
		*pv = preview{loading: true, loadingHead: head, loadingParts: parts}
		vm.Doc.Set(Response{Head: head, Parts: parts}.PageAt(vm.top))
		vm.RailDoc.Set(railPage(head, parts))
	}
}

// setSlides records where each slide starts and tells the thumbnails when that
// changed (the count, or a line: a slide inserted above moves the others).
func (vm *ViewModel) setSlides(parts []string) {
	lines := SlideLines(parts)
	if slices.Equal(lines, vm.SlideLines) {
		return
	}
	vm.SlideLines = lines
	vm.SlidesVersion.Set(vm.SlidesVersion.Get() + 1)
	vm.followCaret()
}

// followCaret moves CurrentSlide to the slide the caret is in.
func (vm *ViewModel) followCaret() {
	vm.CurrentSlide.Set(SlideAt(vm.SlideLines, vm.CaretLine.Get()))
	// The preview scrolls to the caret smoothly, reporting the slides it passes:
	// those reports must not pull the selection off the slide the caret chose.
	vm.holdSlide = time.Now().Add(slideHold)
}

// slideHold is how long after the caret moved the preview's scroll reports are
// left unheard by the thumbnails (a smooth scroll to a far slide).
const slideHold = time.Second

// RailPage is the thumbnail frame's document for the overview the preview
// holds now: what a frame that missed a patch is reloaded with ("" for none).
func (vm *ViewModel) RailPage() string {
	if vm.pv.head == "" {
		return ""
	}
	return railPage(vm.pv.head, vm.pv.parts)
}

// PreviewScrolled is the frame's report of where the overview is
// scrolled; the next document the overview loads starts there.
func (vm *ViewModel) PreviewScrolled(top SlideTop) {
	if top.I >= 0 && top.I < 100000 {
		vm.top = top
		if top.I < len(vm.SlideLines) && time.Now().After(vm.holdSlide) {
			vm.CurrentSlide.Set(top.I) // the thumbnails follow the preview's scroll
		}
	}
}

// PreviewPlace is where the overview was last scrolled (I < 0 if never):
// a document that starts there needs no scroll to the caret.
func (vm *ViewModel) PreviewPlace() SlideTop { return vm.top }

// PreviewLoaded is the host's word that the preview frame loaded the
// document last set: an overview waiting for it is applied now.
func (vm *ViewModel) PreviewLoaded() {
	pv := &vm.pv
	if !pv.loading {
		return
	}
	*pv = preview{head: pv.loadingHead, parts: pv.loadingParts, wanted: pv.wanted, wantHead: pv.wantHead, wantParts: pv.wantParts}
	if pv.wanted {
		head, parts := pv.wantHead, pv.wantParts
		pv.wanted, pv.wantHead, pv.wantParts = false, "", nil
		vm.showOverview(head, parts)
	}
}

// headKey is an overview's head without its slide count (the container's
// data-slides, which a patch updates): two overviews whose keys are equal
// differ by their slides alone.
func headKey(head string) string {
	const attr = ` data-slides="`
	i := strings.Index(head, attr)
	if i < 0 {
		return head
	}
	j := strings.IndexByte(head[i+len(attr):], '"')
	if j < 0 {
		return head
	}
	return head[:i] + head[i+len(attr)+j+1:]
}

// diffParts is the patch from one list of slides to the next; ok is false
// when nothing changed.
func diffParts(from, to []string) (PreviewPatch, bool) {
	p := PreviewPatch{Count: len(to)}
	for i, s := range to {
		if i >= len(from) || from[i] != s {
			p.Changed = append(p.Changed, SlidePatch{I: i, HTML: s})
		}
	}
	return p, len(p.Changed) > 0 || len(to) != len(from)
}

// GiveUp abandons the request in flight (the host stopped a worker that took
// too long) and says so.
func (vm *ViewModel) GiveUp(reason string) {
	id := vm.inFlight
	if id == 0 {
		return
	}
	vm.Receive(Response{ID: id, Error: reason})
}

// Busy reports whether a render is in flight.
func (vm *ViewModel) Busy() bool { return vm.inFlight != 0 }

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// formatProblems writes one line per diagnostic: "line 12: message".
func formatProblems(ds []editor.Diagnostic) string {
	var b strings.Builder
	for _, d := range ds {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		where := fmt.Sprintf("line %d", d.Line+1)
		if d.File != DeckPath && d.File != "" {
			where = d.File + ":" + where
		}
		sev := "error"
		if d.Severity == "warning" {
			sev = "warning"
		}
		fmt.Fprintf(&b, "%s %s: %s", sev, where, d.Message)
	}
	return b.String()
}

// request is the editor request for the current source.
func (vm *ViewModel) request() editor.Request {
	return editor.Request{Path: DeckPath, Source: vm.Source.Get()}
}

// utf16Col converts a rune column of a line to the UTF-16 column the editor
// package counts in.
func utf16Col(line string, runeCol int) int {
	n := 0
	for i, r := range []rune(line) {
		if i >= runeCol {
			break
		}
		n += utf16.RuneLen(r)
	}
	return n
}

// lineAt returns line i of the source, "" past the end.
func (vm *ViewModel) lineAt(i int) string {
	lines := strings.Split(vm.Source.Get(), "\n")
	if i < 0 || i >= len(lines) {
		return ""
	}
	return lines[i]
}

// Complete returns the completions at a caret, rune column.
func (vm *ViewModel) Complete(line, runeCol int) []editor.Item {
	return editor.Complete(vm.request(), line, utf16Col(vm.lineAt(line), runeCol))
}

// updateHelp shows the help for what the caret is on, else the general help.
func (vm *ViewModel) updateHelp() {
	line := vm.CaretLine.Get()
	h := editor.HoverAt(vm.request(), line, utf16Col(vm.lineAt(line), vm.CaretCol.Get()))
	if h == nil {
		vm.Help.Set(GeneralHelp)
		return
	}
	vm.Help.Set(plainMarkdown(h.Doc))
}

// present shows the slideshow in the preview and opens the presenter
// window beside it; the page relays slide changes between the two.
func (vm *ViewModel) present() {
	vm.Mode.Set(int(ModeSlideshow))
	vm.host.OpenPresenter()
	vm.Status.Set("Presenting: the presenter view is in its own window; arrow keys in either move both.")
}

// share puts the deck in the URL's fragment and copies the link.
func (vm *ViewModel) share() {
	frag, err := EncodeShare(vm.Source.Get())
	if err != nil {
		vm.Status.Set("⚠ " + err.Error())
		return
	}
	vm.host.SetFragment(frag)
	vm.host.Clipboard(ShareBase + "#" + frag)
	vm.Status.Set(fmt.Sprintf("Link copied: the deck travels in the link itself (%d characters), nothing is uploaded.", len(frag)+1))
}

// download saves the presentation as one HTML page, the slides command's.
// It is built here, on the main thread: a download is one click, not a
// keystroke, so it does not need the worker.
func (vm *ViewModel) download() {
	page, err := Presentation(vm.Source.Get())
	if err != nil {
		vm.Status.Set("⚠ " + err.Error())
		return
	}
	vm.host.Download("deck.html", "text/html", page)
	vm.Status.Set("Saved deck.html: open it in any browser, or present from it.")
}

// ShareBase is the page a shared link opens.
const ShareBase = "https://go-slides.github.io/playground/"
