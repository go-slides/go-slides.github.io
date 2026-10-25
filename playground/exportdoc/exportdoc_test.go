package exportdoc

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-slides/slides"
	"github.com/go-slides/slides/deck"
)

// Build is slides.Render in the box frame, as the slides command exports:
// the same Doc, with raw HTML as the allowlist lets it through.
func TestBuild(t *testing.T) {
	src := "---\ntheme: gaia\n---\n\n# One <b>bold</b>\n\n<!-- note -->\n\n---\n\n## Two\n"
	d, err := Build(Project{Source: src})
	if err != nil {
		t.Fatal(err)
	}
	cfg := slides.Config{HTML: deck.HTMLAllowlist}
	b, err := slides.Render([]byte(src), cfg, deck.FrameBox)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(d, b.Doc()) {
		t.Fatal("the export's deck is not the command's")
	}
	if len(d.Slides) != 2 || !strings.Contains(d.Slides[0].HTML, "<b>bold</b>") || len(d.Slides[0].Notes) != 1 {
		t.Fatalf("%d slides, %q", len(d.Slides), d.Slides[0].Notes)
	}
	if _, err := Build(Project{Source: strings.Repeat("x", MaxSource+1)}); err == nil {
		t.Fatal("an oversize deck was built")
	}
	if _, err := Build(Project{Source: src, Files: map[string]string{"/etc/passwd": "x"}}); err == nil {
		t.Fatal("a file outside the project was read")
	}
}
