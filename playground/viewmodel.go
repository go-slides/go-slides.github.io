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
	"strings"
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
}

// ViewModel holds the playground's state as go-widgets/mvvm observables. It
// references no widget; scene.go binds the widgets to it.
type ViewModel struct {
	Source  *mvvm.Observable[string]
	Example *mvvm.Observable[int]
	Mode    *mvvm.Observable[int]
	// CaretLine and CaretCol are the editor's caret (0-based; the column
	// counts runes), which the context help and the scroll sync follow.
	CaretLine *mvvm.Observable[int]
	CaretCol  *mvvm.Observable[int]

	// Derived.
	Doc      *mvvm.Observable[string] // the preview iframe's document
	Status   *mvvm.Observable[string] // "12 slides · 48 ms"
	Problems *mvvm.Observable[string] // one line per diagnostic
	Help     *mvvm.Observable[string] // the context help at the caret
	// Diagnostics are the last render's problems, for the editor's
	// highlights; Version counts the renders that changed them.
	Diagnostics []editor.Diagnostic
	DiagVersion *mvvm.Observable[int]

	// Commands.
	Share    *mvvm.Command
	Download *mvvm.Command
	Reset    *mvvm.Command

	transport Transport
	host      Host

	next     int     // the next request ID
	inFlight int     // the request awaiting an answer, 0 if none
	pending  bool    // the source changed while a request was in flight
	last     Request // the request in flight
}

// NewViewModel builds the view-model on a starting deck: the shared deck of
// the URL's fragment if there is one, else the first example.
func NewViewModel(t Transport, h Host, fragment string) *ViewModel {
	vm := &ViewModel{
		Source:      mvvm.NewObservable(""),
		Example:     mvvm.NewObservable(0),
		Mode:        mvvm.NewObservable(int(ModeOverview)),
		CaretLine:   mvvm.NewObservable(0),
		CaretCol:    mvvm.NewObservable(0),
		Doc:         mvvm.NewObservable(""),
		Status:      mvvm.NewObservable("Starting…"),
		Problems:    mvvm.NewObservable(""),
		Help:        mvvm.NewObservable(GeneralHelp),
		DiagVersion: mvvm.NewObservable(0),
		transport:   t,
		host:        h,
	}
	src := Examples[0].Source
	if fragment != "" {
		if s, err := DecodeShare(fragment); err == nil {
			src = s
			vm.Status.Set("Opened the deck shared in the link.")
		} else {
			vm.Status.Set("The link's deck could not be read: " + err.Error())
		}
	}
	vm.Source.Set(src)
	vm.Source.Subscribe(func(string) { vm.changed() })
	vm.Mode.Subscribe(func(int) { vm.changed() })
	vm.Example.Subscribe(func(i int) {
		if i >= 0 && i < len(Examples) {
			vm.Source.Set(Examples[i].Source)
		}
	})
	vm.CaretLine.Subscribe(func(int) { vm.updateHelp() })
	vm.CaretCol.Subscribe(func(int) { vm.updateHelp() })

	vm.Share = mvvm.NewCommand(vm.share, nil)
	vm.Download = mvvm.NewCommand(vm.download, nil)
	vm.Reset = mvvm.NewCommand(func() {
		i := vm.Example.Get()
		vm.Source.Set(Examples[max(i, 0)].Source)
	}, nil)
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
	vm.last = Request{ID: vm.next, Source: vm.Source.Get(), Mode: Mode(vm.Mode.Get())}
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
		if r.Doc != "" {
			vm.Doc.Set(r.Doc)
		}
		status := fmt.Sprintf("%d slide%s · rendered in %d ms", r.Slides, plural(r.Slides), r.Millis)
		if r.Doc == "" {
			status = "the preview shows the last deck that rendered"
		}
		if n := len(r.Diagnostics); n > 0 {
			status = fmt.Sprintf("%d problem%s · %s", n, plural(n), status)
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
