package exporter

import (
	"context"
	"fmt"
	"testing"
)

func TestNoticesKnowTheExportsLines(t *testing.T) {
	for _, c := range []struct {
		line, code string
		args       []string
	}{
		{"svg-animation: spin.svg (slide 1): animated natively in PPTX: 4 elements (rotate, opacity)", "export.native", []string{"spin.svg", "1", "PPTX", "4", "rotate, opacity", ""}},
		{"svg-animation: a b.svg (slides 1, 3): animated natively in ODP: 1 element (scale); not carried: <rect> starts on \"click\", an event", "export.native", []string{"a b.svg", "1, 3", "ODP", "1", "scale", "; not carried: <rect> starts on \"click\", an event"}},
		{"svg-animation: morph.svg is an animated GIF of 24 frames at 12 fps, 600x600 pixels, 55 KiB: <path id=p>: it animates d", "export.gif", []string{"morph.svg", "24", "12", "600x600", "55", "<path id=p>: it animates d"}},
		{"svg-animation: x.svg (slide 2): animated as frames in PPTX, an animated GIF: it is skewed", "export.frames", []string{"x.svg", "2", "PPTX", "it is skewed"}},
		{"svg-animation: x.svg is kept as a still picture, which says nothing moves: too small", "export.still", []string{"x.svg", "too small"}},
		{"svg-animation: x.svg (slide 1) is animated, which a PDF cannot show: it is drawn at its initial state (use HTML, PPTX or ODP to see it move)", "export.initial", []string{"x.svg", "1"}},
		{`local-file-blocked: "a.svg" (slides 1, 2) is a local file, blocked by security reason: ...`, "export.blocked", []string{"a.svg", "1, 2"}},
		{`local-file-missing: "gone.svg" (slide 4) does not exist: ...`, "export.missing", []string{"gone.svg", "4"}},
		{"something else", "export.other", []string{"something else"}},
	} {
		n := notice(c.line)
		if n.Code != c.code || fmt.Sprint(n.Args) != fmt.Sprint(toAny(c.args)) || n.Text != c.line {
			t.Errorf("%q\n got %s %q\nwant %s %q", c.line, n.Code, n.Args, c.code, c.args)
		}
	}
}

func toAny(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

// What the export really says is what Notices knows: no line of a real export
// of animated images falls through to "export.other".
func TestEveryLineOfARealExportIsKnown(t *testing.T) {
	p := src("# One\n\n![w:300](media/spin.svg)\n\n---\n\n![w:300](media/morph.svg)\n")
	p.Media = map[string][]byte{
		"media/spin.svg":  []byte(animatedSVG),
		"media/morph.svg": []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><path d="M10 10L90 10L90 90Z"><animate attributeName="d" values="M10 10L90 10L90 90Z;M10 10L90 50L50 90Z" dur="1s"/></path></svg>`),
	}
	for _, format := range []string{"pptx", "odp", "pdf"} {
		r, err := run(context.Background(), p, format, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Warnings) != 2 {
			t.Fatalf("%s: warnings %q", format, r.Warnings)
		}
		for _, n := range Notices(r.Warnings) {
			if n.Code == "export.other" {
				t.Errorf("%s: a line Notices does not know: %q", format, n.Text)
			}
		}
	}
}
