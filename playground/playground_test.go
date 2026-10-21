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
}

func (h *fakeHost) OpenPresenter() { h.presenters++ }

func (h *fakeHost) SetFragment(f string)     { h.fragment = f }
func (h *fakeHost) Download(n, _, d string)  { h.downloads[n] = d }
func (h *fakeHost) Open(u string)            { h.opened = u }
func (h *fakeHost) Clipboard(t string)       { h.clip = t }
func newHost() *fakeHost                     { return &fakeHost{downloads: map[string]string{}} }
func answer(t *fakeTransport, vm *ViewModel) { t.vm = vm; vm.Receive(Serve(t.sent[len(t.sent)-1])) }

func TestEveryExampleRendersCleanly(t *testing.T) {
	if len(Examples) != 5 {
		t.Fatalf("%d examples", len(Examples))
	}
	for _, e := range Examples {
		r := Serve(Request{ID: 1, Source: e.Source})
		if r.Error != "" || len(r.Diagnostics) != 0 || r.Slides < 2 || !strings.Contains(r.Doc, "<svg") {
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
	m := regexp.MustCompile(`script-src 'nonce-([^']+)'`).FindStringSubmatch(r.Doc)
	if m == nil {
		t.Fatal("no nonce in the CSP")
	}
	if strings.Count(r.Doc, "<script") != 1 || !strings.Contains(r.Doc, `<script nonce="`+m[1]+`">`) {
		t.Fatal("the overview's one script is not the nonced one")
	}
	if !strings.Contains(r.Doc, `class="code-line" data-line=`) {
		t.Fatal("no source lines for the scroll sync")
	}
	if Serve(Request{Source: "x"}).Doc == r.Doc {
		t.Fatal("same document for two decks")
	}
}

// TestSlideshowScriptsAreHashed: every script of the presentation, the shim
// included, is allowed by its hash, and nothing else is.
func TestSlideshowScriptsAreHashed(t *testing.T) {
	r := Serve(Request{ID: 1, Source: Examples[0].Source, Mode: ModeSlideshow})
	scripts := scriptRE.FindAllStringSubmatch(r.Doc, -1)
	if len(scripts) < 2 {
		t.Fatalf("%d scripts", len(scripts))
	}
	for _, s := range scripts {
		sum := sha256.Sum256([]byte(s[1]))
		if !strings.Contains(r.Doc, "'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'") {
			t.Fatalf("a script is not in the CSP: %.80s", s[1])
		}
	}
	if !strings.Contains(r.Doc, sandboxShim) || strings.Index(r.Doc, "Content-Security-Policy") > strings.Index(r.Doc, sandboxShim) {
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
		if strings.Contains(r.Doc, "alert(1)</script>") || strings.Contains(r.Doc, "onerror") {
			t.Fatalf("mode %d let deck HTML through", m)
		}
	}
}

func TestServeBounds(t *testing.T) {
	r := Serve(Request{ID: 7, Source: strings.Repeat("a", MaxSource+1)})
	if r.ID != 7 || r.Error == "" || r.Doc != "" {
		t.Fatalf("%+v", r)
	}
	if _, err := Presentation(strings.Repeat("a", MaxSource+1)); err == nil {
		t.Fatal("an oversized presentation")
	}
	// An error leaves no document: the previous preview stays.
	e := Serve(Request{Source: "---\nslides: true\n---\n\n# {{ .nope }}\n"})
	if e.Doc != "" || len(e.Diagnostics) == 0 || e.Diagnostics[0].Line != 4 {
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
