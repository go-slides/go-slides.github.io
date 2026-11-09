package playground

import (
	"regexp"
	"strings"
	"testing"

	"github.com/go-slides/slides/editor"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

const testCast = "{\"version\":2,\"width\":30,\"height\":4}\n[0.2,\"o\",\"\\u001b[32mhello\\u001b[0m\\r\\n\"]\n[0.9,\"o\",\"world\"]\n"

// A recording added to the project goes in media/, is text (the renderer reads
// it) and is named in the deck as an image.
func TestRecordingsAreProjectTextInMedia(t *testing.T) {
	m := store.NewMemory()
	vm, _, _ := openFiles(t, m)
	p, err := vm.AddFile("", "My Demo.cast", []byte(testCast))
	if err != nil || p != "media/My Demo.cast" {
		t.Fatalf("%q, %v", p, err)
	}
	if IsMedia(p) || !IsRecording(p) || !IsRecording("x.CAST") || IsRecording("x.castle") {
		t.Fatal("a recording is text the renderer reads, not a medium")
	}
	if got := MediaMarkdown(p); got != "![My Demo](media/My%20Demo.cast)" {
		t.Fatalf("markdown = %q", got)
	}
	if string(vm.project().Text(p)) != testCast {
		t.Fatal("the recording is not in the project's text files")
	}
	if again, _ := vm.AddFile("", "My Demo.cast", []byte(testCast)); again != p {
		t.Fatalf("the same bytes again: %q", again)
	}
}

// The preview page carries the player, by its nonce, only when a slide has a
// recording; the page's policy names no other script.
func TestPreviewPageHasThePlayerForRecordings(t *testing.T) {
	res := editor.Render(editor.Request{
		Path: "deck.md", Source: "# Demo\n\n![Demo](media/demo.cast#t=0.5)\n",
		Files: map[string]string{"media/demo.cast": testCast},
	})
	if len(res.Diagnostics) != 0 || !strings.Contains(res.HTML, "slides-cast-data") {
		t.Fatalf("%+v", res.Diagnostics)
	}
	page := overviewPage(res.HTML, SlideTop{I: -1})
	nonce := regexp.MustCompile(`script-src 'nonce-([^']+)'`).FindStringSubmatch(page)
	if nonce == nil {
		t.Fatal("no nonce in the policy")
	}
	for _, s := range regexp.MustCompile(`(?i)<script[^>]*>`).FindAllString(page, -1) {
		if s != `<script nonce="`+nonce[1]+`">` && s != `<script type="application/json" class="slides-cast-data">` {
			t.Errorf("a script the policy does not name: %s", s)
		}
	}
	for _, want := range []string{"slidesCastInit", "prefers-reduced-motion", ".slides-cast-ctl"} {
		if !strings.Contains(page, want) {
			t.Errorf("the page lacks %q", want)
		}
	}
	plain := overviewPage(editor.Render(editor.Request{Path: "deck.md", Source: "# No recording\n"}).HTML, SlideTop{I: -1})
	if strings.Contains(plain, "window.slidesCastInit = function") || strings.Contains(plain, "slides-cast-ctl") {
		t.Error("the player is in a preview with no recording")
	}
	// a patch makes the player look for the recordings it brought
	if !strings.Contains(plain, "slidesCastInit()") {
		t.Error("a patch does not start the player on a new recording")
	}
}

// The slideshow (viewer.Document, with the frame's hash policy) names the
// player's script by its hash, and the data element is inert.
func TestSlideshowPageHasThePlayer(t *testing.T) {
	page, err := ProjectPresentation("# Demo\n\n![](media/demo.cast)\n", map[string]string{"/playground/media/demo.cast": testCast}, nil)
	if err != nil {
		t.Fatal(err)
	}
	page = sandboxPage(page)
	if !strings.Contains(page, "window.slidesCastInit = function") || !strings.Contains(page, `class="slides-cast-data"`) {
		i := strings.Index(page, "slides-cast")
		t.Errorf("the slideshow has no player: %.300s", page[max(i, 0):])
	}
	// only plain <script> elements are named by hash: the data element has a type
	if regexp.MustCompile(`(?s)<script>\{"w":`).MatchString(page) {
		t.Error("the recording's data is a script that runs")
	}
}
