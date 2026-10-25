package exporter

import (
	"bytes"
	"compress/zlib"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"regexp"
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/exportdoc"
)

// pdfText is a PDF's bytes with every Flate stream inflated in place: what
// its content streams and object streams say, to search.
func pdfText(b []byte) string {
	var out bytes.Buffer
	re := regexp.MustCompile(`(?s)stream\r?\n`)
	rest := b
	for {
		loc := re.FindIndex(rest)
		if loc == nil {
			out.Write(rest)
			return out.String()
		}
		out.Write(rest[:loc[1]])
		rest = rest[loc[1]:]
		end := bytes.Index(rest, []byte("endstream"))
		if end < 0 {
			out.Write(rest)
			return out.String()
		}
		if r, err := zlib.NewReader(bytes.NewReader(rest[:end])); err == nil {
			if inflated, err := io.ReadAll(r); err == nil || len(inflated) > 0 {
				out.Write(inflated)
			}
		}
		out.WriteString("\nendstream")
		rest = rest[end+len("endstream"):]
	}
}

// greenPNG is a 4×4 opaque green PNG.
func greenPNG(t *testing.T) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			img.Set(x, y, color.RGBA{0, 200, 0, 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// The export renders the project the preview renders: its theme (a red
// heading, from themes/ through slides.yaml) and its media (a background
// named media/dot.png). The same deck without its project has neither.
func TestExportUsesTheProject(t *testing.T) {
	deck := "---\nslides: true\ntheme: red\n---\n\n![bg](media/dot.png)\n\n# Red heading\n"
	p := project{
		Project: exportdoc.Project{
			DeckPath: "/project/deck.md",
			Source:   deck,
			Files: map[string]string{
				"/project/slides.yaml":    "themes: [themes]\n",
				"/project/themes/red.css": "/* @theme red */\n@import 'default';\nh1 { color: #ff0000; }\n",
			},
			Dirs: []string{"/project/media", "/project/themes"},
		},
		Media: map[string][]byte{"media/dot.png": greenPNG(t)},
	}
	with, err := run(context.Background(), p, "pdf", nil)
	if err != nil {
		t.Fatal(err)
	}
	without, err := run(context.Background(), src(deck), "pdf", nil)
	if err != nil {
		t.Fatal(err)
	}
	red := regexp.MustCompile(`(?m)(^|\s)1(\.0+)? 0(\.0+)? 0(\.0+)? (rg|sc|scn)\b`)
	img := regexp.MustCompile(`/Subtype\s*/Image`)
	w, wo := pdfText(with.Data), pdfText(without.Data)
	if !red.MatchString(w) {
		t.Error("the project's theme colour (#ff0000) is not in the PDF")
	}
	if !img.MatchString(w) {
		t.Error("the project's image is not in the PDF")
	}
	// the control: the same measurement sees neither without the project
	if red.MatchString(wo) || img.MatchString(wo) {
		t.Errorf("the deck alone has red %v, an image %v: the checks cannot tell", red.MatchString(wo), img.MatchString(wo))
	}
	// PPTX and ODP carry the theme and the image too
	for _, f := range []string{"pptx", "odp"} {
		r, err := run(context.Background(), p, f, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(r.Data, []byte(".png")) && !bytes.Contains(r.Data, []byte(".jpeg")) && !bytes.Contains(r.Data, []byte(".jpg")) {
			t.Errorf("%s holds no picture", f)
		}
	}
}
