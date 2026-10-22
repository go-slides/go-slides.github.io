package playground

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"encoding/base64"
	"regexp"
	"strings"
	"testing"

	"github.com/go-widgets/toolkit"
)

// fakeTransport answers every request in-process, as the worker would.
type fakeTransport struct {
	vm   *ViewModel
	sent []Request
	hold bool // keep requests unanswered
}

func (f *fakeTransport) Send(r Request) {
	f.sent = append(f.sent, r)
	if !f.hold && f.vm != nil {
		f.vm.Receive(Serve(r))
	}
}

type fakeHost struct {
	fragment, clip, opened string
	downloads              map[string]string
	presenters             int
	patches                []PreviewPatch
}

func (h *fakeHost) OpenPresenter() { h.presenters++ }

func (h *fakeHost) PatchPreview(p PreviewPatch) { h.patches = append(h.patches, p) }

func (h *fakeHost) SetFragment(f string)     { h.fragment = f }
func (h *fakeHost) Download(n, _, d string)  { h.downloads[n] = d }
func (h *fakeHost) Open(u string)            { h.opened = u }
func (h *fakeHost) Clipboard(t string)       { h.clip = t }
func newHost() *fakeHost                     { return &fakeHost{downloads: map[string]string{}} }
func answer(t *fakeTransport, vm *ViewModel) { t.vm = vm; vm.Receive(Serve(t.sent[len(t.sent)-1])) }

func TestEveryExampleRendersCleanly(t *testing.T) {
	if len(Examples) != 6 {
		t.Fatalf("%d examples", len(Examples))
	}
	for _, e := range Examples {
		r := Serve(Request{ID: 1, Source: e.Source})
		if r.Error != "" || len(r.Diagnostics) != 0 || r.Slides < 2 || !strings.Contains(r.Page(), "<svg") {
			t.Errorf("%s: error %q, %d slides, diagnostics %+v", e.Name, r.Error, r.Slides, r.Diagnostics)
		}
		s := Serve(Request{ID: 2, Source: e.Source, Mode: ModeSlideshow})
		if s.Error != "" || !strings.Contains(s.Doc, "Content-Security-Policy") {
			t.Errorf("%s slideshow: %q", e.Name, s.Error)
		}
	}
}

func TestOverviewPageIsLockedDown(t *testing.T) {
	r := Serve(Request{ID: 1, Source: Examples[0].Source})
	page := r.Page() // each page has a nonce of its own
	m := regexp.MustCompile(`script-src 'nonce-([^']+)'`).FindStringSubmatch(page)
	if m == nil {
		t.Fatal("no nonce in the CSP")
	}
	// two scripts (the media's, the page's), both nonced
	if n := strings.Count(page, "<script"); n != 2 || strings.Count(page, `<script nonce="`+m[1]+`">`) != n {
		t.Fatal("the overview's scripts are not the nonced ones")
	}
	if !strings.Contains(page, `class="code-line" data-line=`) {
		t.Fatal("no source lines for the scroll sync")
	}
	if strings.Contains(r.Page(), m[1]) {
		t.Fatal("two pages with the same nonce")
	}
	if Serve(Request{Source: "x"}).Head == r.Head && strings.Join(Serve(Request{Source: "x"}).Parts, "") == strings.Join(r.Parts, "") {
		t.Fatal("same overview for two decks")
	}
}

// TestSlideshowScriptsAreHashed: every script of the presentation, the shim
// included, is allowed by its hash, and nothing else is.
func TestSlideshowScriptsAreHashed(t *testing.T) {
	r := Serve(Request{ID: 1, Source: Examples[0].Source, Mode: ModeSlideshow})
	scripts := scriptRE.FindAllStringSubmatch(r.Page(), -1)
	if len(scripts) < 2 {
		t.Fatalf("%d scripts", len(scripts))
	}
	for _, s := range scripts {
		sum := sha256.Sum256([]byte(s[1]))
		if !strings.Contains(r.Page(), "'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'") {
			t.Fatalf("a script is not in the CSP: %.80s", s[1])
		}
	}
	if !strings.Contains(r.Page(), sandboxShim) || strings.Index(r.Page(), "Content-Security-Policy") > strings.Index(r.Page(), sandboxShim) {
		t.Fatal("the CSP must come first, then the shim")
	}
	if got := sandboxPage("<p>no head"); !strings.HasPrefix(got, "<meta") {
		t.Fatal(got)
	}
}

// TestDeckHTMLStaysOnTheAllowlist: a script in the deck is not in the
// preview, in either mode.
func TestDeckHTMLStaysOnTheAllowlist(t *testing.T) {
	src := "---\nslides: true\n---\n\n<script>alert(1)</script>\n\n<img src=x onerror=alert(2)>\n"
	for _, m := range []Mode{ModeOverview, ModeSlideshow} {
		r := Serve(Request{Source: src, Mode: m})
		if strings.Contains(r.Page(), "alert(1)</script>") || strings.Contains(r.Page(), "onerror") {
			t.Fatalf("mode %d let deck HTML through", m)
		}
	}
}

func TestServeBounds(t *testing.T) {
	r := Serve(Request{ID: 7, Source: strings.Repeat("a", MaxSource+1)})
	if r.ID != 7 || r.Error == "" || r.Page() != "" {
		t.Fatalf("%+v", r)
	}
	if _, err := Presentation(strings.Repeat("a", MaxSource+1)); err == nil {
		t.Fatal("an oversized presentation")
	}
	// An error leaves no document: the previous preview stays.
	e := Serve(Request{Source: "---\nslides: true\n---\n\n# {{ .nope }}\n"})
	if e.Page() != "" || len(e.Diagnostics) == 0 || e.Diagnostics[0].Line != 4 {
		t.Fatalf("%+v", e)
	}
}

func TestShare(t *testing.T) {
	src := Examples[1].Source
	f, err := EncodeShare(src)
	if err != nil || !strings.HasPrefix(f, "deck=") {
		t.Fatal(f, err)
	}
	for _, in := range []string{f, "#" + f} {
		if got, err := DecodeShare(in); err != nil || got != src {
			t.Fatalf("round trip: %v", err)
		}
	}
	if _, err := EncodeShare(strings.Repeat("é and some entropy ", 1) + randomText(200<<10)); err == nil {
		t.Fatal("a deck too large for a link was encoded")
	}
	bomb := func(n int, b byte) string {
		var buf bytes.Buffer
		w, _ := flate.NewWriter(&buf, flate.BestCompression)
		w.Write(bytes.Repeat([]byte{b}, n))
		w.Close()
		return "deck=" + base64.RawURLEncoding.EncodeToString(buf.Bytes())
	}
	for name, frag := range map[string]string{
		"no deck":   "other=1",
		"damaged":   "deck=!!!",
		"truncated": f[:len(f)/2],
		"too long":  "deck=" + strings.Repeat("A", MaxShare),
		"bomb":      bomb(64<<20, 'a'),
		"binary":    bomb(10, 0xff),
	} {
		if _, err := DecodeShare(frag); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func randomText(n int) string {
	var b strings.Builder
	x := uint32(1)
	for b.Len() < n {
		x = x*1664525 + 1013904223
		b.WriteByte(byte('a' + x>>24%26))
	}
	return b.String()
}

func TestViewModelCoalescesRenders(t *testing.T) {
	tr := &fakeTransport{hold: true}
	h := newHost()
	vm := NewViewModel(tr, h, "")
	if len(tr.sent) != 1 || !vm.Busy() {
		t.Fatalf("%d requests", len(tr.sent))
	}
	// Three edits while the first render is out: one more request, the last.
	vm.Source.Set("a")
	vm.Source.Set("ab")
	vm.Source.Set("abc")
	if len(tr.sent) != 1 {
		t.Fatalf("%d requests while one was in flight", len(tr.sent))
	}
	answer(tr, vm)
	if len(tr.sent) != 2 || tr.sent[1].Source != "abc" {
		t.Fatalf("%+v", tr.sent)
	}
	// An answer to an abandoned request is dropped.
	vm.Receive(Response{ID: 999, Doc: "stale"})
	if vm.Doc.Get() == "stale" {
		t.Fatal("a stale answer was shown")
	}
	vm.GiveUp("too slow")
	if !strings.Contains(vm.Status.Get(), "too slow") || vm.Busy() {
		t.Fatal(vm.Status.Get())
	}
	vm.GiveUp("nothing in flight") // no-op
}

// TestPreviewPatch: once the frame has loaded an overview, an edit that
// keeps the style sheet patches the slides that changed instead of loading
// a document; a render arriving while a document loads waits for the load.
func TestPreviewPatch(t *testing.T) {
	deck := func(two, three string) string {
		return "---\nslides: true\n---\n\n# One\n\n---\n\n# " + two + "\n" + three
	}
	tr := &fakeTransport{}
	h := newHost()
	vm := NewViewModel(tr, h, "")
	tr.hold = true
	answer(tr, vm) // the first deck: a document, loading
	vm.Source.Set(deck("Two", ""))
	answer(tr, vm)
	doc := vm.Doc.Get()
	if !strings.Contains(doc, "Hello") || len(h.patches) != 0 {
		t.Fatal("the first overview is a document")
	}
	// An edit while the document loads: no new document, no patch yet.
	vm.Source.Set(deck("Two!", ""))
	answer(tr, vm)
	if vm.Doc.Get() != doc || len(h.patches) != 0 {
		t.Fatal("a render during the load replaced or patched the document")
	}
	// The load: the latest overview is applied as a patch.
	vm.PreviewLoaded()
	if len(h.patches) != 1 || h.patches[0].Count != 2 || !strings.Contains(h.patches[0].Changed[len(h.patches[0].Changed)-1].HTML, "Two!") {
		t.Fatalf("patch after the load: %+v", h.patches)
	}
	tr.hold = false
	// One slide edited: that slide alone.
	vm.Source.Set(deck("Two?", ""))
	if p := h.patches[len(h.patches)-1]; len(p.Changed) != 1 || p.Changed[0].I != 1 || !strings.Contains(p.Changed[0].HTML, "Two?") {
		t.Fatalf("one slide: %+v", p)
	}
	vm.Source.Set(deck("Two!", ""))
	// A third slide: added by the patch; then removed again.
	vm.Source.Set(deck("Two!", "\n---\n\n# Three\n"))
	p := h.patches[len(h.patches)-1]
	if p.Count != 3 || len(p.Changed) != 1 || p.Changed[0].I != 2 {
		t.Fatalf("added slide: %+v", p)
	}
	vm.Source.Set(deck("Two!", ""))
	if p := h.patches[len(h.patches)-1]; p.Count != 2 || len(p.Changed) != 0 {
		t.Fatalf("removed slide: %+v", p)
	}
	// The same deck again: nothing to send.
	n := len(h.patches)
	vm.Source.Set(deck("Two!", "") + "\n")
	if len(h.patches) != n {
		t.Fatalf("an unchanged deck was patched: %+v", h.patches[n:])
	}
	// Another theme changes the head: a new document.
	vm.Source.Set("---\nslides: true\ntheme: gaia\n---\n\n# One\n")
	if vm.Doc.Get() == doc || !strings.Contains(vm.Doc.Get(), `data-theme="gaia"`) || len(h.patches) != n {
		t.Fatal("a new theme was patched in")
	}
	// The slideshow replaces the document, and the overview after it is a
	// new document too.
	vm.PreviewLoaded()
	vm.Mode.Set(int(ModeSlideshow))
	vm.PreviewLoaded()
	vm.Mode.Set(int(ModeOverview))
	if !strings.Contains(vm.Doc.Get(), "go-slides-preview") || len(h.patches) != n {
		t.Fatal("the overview after the slideshow was patched into it")
	}
}

func TestViewModelRendersAndExplains(t *testing.T) {
	tr := &fakeTransport{}
	h := newHost()
	vm := NewViewModel(tr, h, "")
	answer(tr, vm)
	if !strings.Contains(vm.Doc.Get(), "<svg") || !strings.Contains(vm.Status.Get(), "4 slides") {
		t.Fatal(vm.Status.Get())
	}
	tr.vm = vm
	vm.Source.Set("---\nslides: true\npaginate: true\n---\n\n# {{ .nope }}\n")
	if !strings.Contains(vm.Problems.Get(), "error line 6") || !strings.Contains(vm.Status.Get(), "1 problem ·") {
		t.Fatalf("%q / %q", vm.Problems.Get(), vm.Status.Get())
	}
	// The help follows the caret: on paginate, the directive's doc.
	vm.CaretLine.Set(2)
	vm.CaretCol.Set(2)
	if !strings.Contains(vm.Help.Get(), "page number") || strings.Contains(vm.Help.Get(), "**") {
		t.Fatal(vm.Help.Get())
	}
	vm.CaretLine.Set(4)
	if vm.Help.Get() != GeneralHelp {
		t.Fatal("no general help off a directive")
	}
	if items := vm.Complete(2, len("paginate: ")); len(items) == 0 {
		t.Fatal("no completion for paginate")
	}
	vm.Mode.Set(int(ModeSlideshow))
	if tr.sent[len(tr.sent)-1].Mode != ModeSlideshow {
		t.Fatal("the mode did not ask for a slideshow")
	}
	vm.Example.Set(2)
	if vm.Source.Get() != Examples[2].Source {
		t.Fatal("the example was not loaded")
	}
	vm.Source.Set("edited")
	vm.Reset.Execute()
	if vm.Source.Get() != Examples[2].Source {
		t.Fatal("reset")
	}
	vm.Receive(Response{ID: vm.inFlight, Error: "boom"})
}

func TestViewModelShareAndDownload(t *testing.T) {
	tr := &fakeTransport{}
	h := newHost()
	vm := NewViewModel(tr, h, "")
	vm.Share.Execute()
	if !strings.HasPrefix(h.fragment, "deck=") || h.clip != ShareBase+"#"+h.fragment {
		t.Fatal(h.fragment, h.clip)
	}
	// The link opens the deck it carries.
	vm2 := NewViewModel(&fakeTransport{}, h, "#"+h.fragment)
	if vm2.Source.Get() != vm.Source.Get() || !strings.Contains(vm2.Status.Get(), "shared") {
		t.Fatal(vm2.Status.Get())
	}
	bad := NewViewModel(&fakeTransport{}, h, "#deck=%%%")
	if bad.Source.Get() != Examples[0].Source || !strings.Contains(bad.Status.Get(), "could not be read") {
		t.Fatal(bad.Status.Get())
	}
	vm.Download.Execute()
	if p := h.downloads["deck.html"]; !strings.Contains(p, "<svg") || strings.Contains(p, "Content-Security-Policy") {
		t.Fatal("the download is not the plain presentation")
	}
	vm.Source.Set(strings.Repeat("x", MaxSource+1))
	vm.Download.Execute()
	vm.Share.Execute()
	if !strings.HasPrefix(vm.Status.Get(), "⚠") {
		t.Fatal(vm.Status.Get())
	}
}

// TestPresent: Present shows the slideshow and opens the presenter window;
// with the window open, each render brings the presenter page too, and
// both pages sync through the parent, never localStorage.
func TestPresent(t *testing.T) {
	tr := &fakeTransport{}
	h := newHost()
	vm := NewViewModel(tr, h, "")
	answer(tr, vm)
	vm.Present.Execute()
	if h.presenters != 1 || vm.Mode.Get() != int(ModeSlideshow) {
		t.Fatal("present")
	}
	if tr.sent[len(tr.sent)-1].Presenter {
		t.Fatal("a presenter page asked for before the window opened")
	}
	vm.PresenterOpen.Set(true) // the host opened it
	last := tr.sent[len(tr.sent)-1]
	if !last.Presenter || last.Mode != ModeSlideshow {
		t.Fatalf("%+v", last)
	}
	pd := vm.PresenterDoc.Get()
	if !strings.Contains(pd, `"view":"presenter","sync":"parent"`) || !strings.Contains(pd, "Content-Security-Policy") {
		t.Fatal("the presenter page")
	}
	if !strings.Contains(vm.Doc.Get(), `"sync":"parent"`) {
		t.Fatal("the audience page does not sync through the parent")
	}
	// Back to the overview: no presenter page is rendered.
	vm.Mode.Set(int(ModeOverview))
	if tr.sent[len(tr.sent)-1].Presenter {
		t.Fatal("presenter page for the overview")
	}
	vm.PresenterOpen.Set(false)
}

func TestFormatProblems(t *testing.T) {
	r := Serve(Request{Source: "---\nslides: true\nkinds: k.yaml\ntheme: nope\n---\n\n# x\n"})
	s := formatProblems(r.Diagnostics)
	// The playground has no files: a kinds file is missing, and says so.
	if !strings.Contains(s, "error line") || !strings.Contains(s, "k.yaml") {
		t.Fatal(s)
	}
	w := Serve(Request{Source: "---\nslides: true\ntheme: nope\n---\n\n# x\n"})
	if s := formatProblems(w.Diagnostics); !strings.HasPrefix(s, "warning line 3") {
		t.Fatal(s)
	}
}

func TestWrap(t *testing.T) {
	got := Wrap("- one two three four five six\nshort", 14)
	want := "- one two\n  three four\n  five six\nshort"
	if got != want {
		t.Fatalf("%q", got)
	}
	if Wrap("abcdefghijklmnop", 5) != "abcde\nfghij\nklmno\np" {
		t.Fatal(Wrap("abcdefghijklmnop", 5))
	}
}

func TestHighlight(t *testing.T) {
	lines := strings.Split("---\ntitle: x\n# comment\n---\n\n# Head\n<!-- _class: a\n -->\n{{ .x }} and {{ .y }}\n```go\n{{ .z }}\n```\n---", "\n")
	for _, theme := range []*toolkit.Theme{LightPalette.Theme(), DarkPalette.Theme(), nil} {
		spans := deckSyntax{}.Highlight("", lines, theme)
		if len(spans) != len(lines) {
			t.Fatal(len(spans))
		}
		if len(spans[1]) != 1 || spans[1][0].End != 5 {
			t.Fatalf("key: %+v", spans[1])
		}
		if len(spans[8]) != 2 || spans[8][1].Start != 13 {
			t.Fatalf("actions: %+v", spans[8])
		}
		if len(spans[7]) != 1 || spans[7][0].End != 4 {
			t.Fatalf("comment continuation: %+v", spans[7])
		}
		if len(spans[10]) != 1 || spans[10][0].Start != 0 || spans[10][0].End != 8 {
			t.Fatalf("code is not template-coloured: %+v", spans[10])
		}
	}
}

// TestHighlightFrontKeys: keys at every depth of a YAML or HCL front
// matter, and nothing that only looks like one.
func TestHighlightFrontKeys(t *testing.T) {
	theme := LightPalette.Theme()
	keyColor, _, _, _, _, _ := syntaxColors(theme)
	keys := func(front string) []string {
		lines := strings.Split(front+"\n---\n# x", "\n")
		spans := deckSyntax{}.Highlight("", lines, theme)
		var got []string
		for i, ss := range spans {
			r := []rune(lines[i])
			for _, s := range ss {
				if s.Color == keyColor {
					got = append(got, string(r[s.Start:s.End]))
				}
			}
		}
		return got
	}
	for _, c := range []struct{ front, want string }{
		{"---\ndata:\n  name: World\n  deep:\n    deeper: 1", "data name deep deeper"},
		{"---\ndata: {name: World, n: 2, sub: {x: 1}}", "data name n sub x"},
		{"---\nitems:\n  - id: 1\n    label: a\n  - b", "items id label"},
		{"---\nurl: https://x.org/a:b\ntime: 12:30\nnote: \"a: b\" # c: d", "url time note"},
		{"---\n\"quoted key\": 1\n'single': 2", `"quoted key" 'single'`},
		{"---\ntitle: x # comment: no", "title"},
		{"---hcl\nslides = true\ndata = {\n  company = \"acme\"\n  year    = 2026\n}\nfooter = \"${a == b}\"\nkind \"big\" {\n  class = \"lead\"\n}\nobj = { a = 1, b: 2 } // c = 3", "slides data company year footer kind class obj a b"},
	} {
		if got := strings.Join(keys(c.front), " "); got != c.want {
			t.Errorf("%q:\n got %q\nwant %q", c.front, got, c.want)
		}
	}
	// the comment of a line is coloured as one
	lines := []string{"---", "title: x # note", "---"}
	ss := deckSyntax{}.Highlight("", lines, theme)[1]
	if len(ss) != 2 || ss[1].Start != 9 || ss[1].Color == keyColor {
		t.Fatalf("comment: %+v", ss)
	}
}

func TestParseColor(t *testing.T) {
	for in, want := range map[string]toolkit.RGBA{
		"#4f46e5": toolkit.RGB(0x4f, 0x46, 0xe5), " #abc ": toolkit.RGB(0xaa, 0xbb, 0xcc),
		"rgb(1, 2, 300)": toolkit.RGB(1, 2, 255), "rgb(1 2 3 / 0.5)": toolkit.RGB(1, 2, 3),
	} {
		if got, ok := ParseColor(in); !ok || got != want {
			t.Errorf("%q: %v %v", in, got, ok)
		}
	}
	for _, bad := range []string{"", "#12", "#gggggg", "rgb(1,2)", "rgb(a,b,c)", "red"} {
		if _, ok := ParseColor(bad); ok {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestPlainMarkdownAndCols(t *testing.T) {
	if got := plainMarkdown("```\nupper S\n```\n\n**bold** `code`"); got != "upper S\n\nbold code" {
		t.Fatalf("%q", got)
	}
	if utf16Col("é😀x", 2) != 3 || runeCol("é😀x", 3) != 2 || runeCol("ab", 9) != 2 {
		t.Fatal("columns")
	}
}
