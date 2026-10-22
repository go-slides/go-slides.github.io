package playground

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

// TestMediaExample: the math, image and background example renders with
// every image reaching the page, in both modes, and its math typeset.
func TestMediaExample(t *testing.T) {
	var src string
	for _, e := range Examples {
		if e.Name == "Math, images and backgrounds" {
			src = e.Source
		}
	}
	if src == "" {
		t.Fatal("no media example")
	}
	sky := "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxMjgwIDcyMC"
	leaves := "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA2MDAgOTAwI"
	chart := "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA0MDAgMjQwI"
	for _, mode := range []Mode{ModeOverview, ModeSlideshow} {
		r := Serve(Request{Source: src, Mode: mode})
		if r.Error != "" || len(r.Diagnostics) != 0 || r.Slides != 5 {
			t.Fatalf("mode %v: %q %+v %d slides", mode, r.Error, r.Diagnostics, r.Slides)
		}
		for name, uri := range map[string]string{"sky": sky, "leaves": leaves, "chart": chart} {
			if !strings.Contains(r.Page(), uri) {
				t.Errorf("mode %v: the %s image is not in the page", mode, name)
			}
		}
		if strings.Contains(r.Page(), "$$") || strings.Contains(r.Page(), `\sqrt`) || strings.Contains(r.Page(), "{{ .img") {
			t.Errorf("mode %v: math or a template left as text", mode)
		}
		if !strings.Contains(r.Page(), "grayscale") {
			t.Errorf("mode %v: the image filter is lost", mode)
		}
	}
}

func TestImageMarkdown(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	md, err := ImageMarkdown("dir/my [photo](x).jpg", buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	// the type comes from the bytes, not the name; the name cannot end the link
	want := "![my photox](data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()) + ")"
	if md != want {
		t.Fatalf("%.80s", md)
	}
	svg := []byte("\n<?xml version=\"1.0\"?>\n<svg xmlns=\"http://www.w3.org/2000/svg\"/>")
	if md, err := ImageMarkdown("a.svg", svg); err != nil || !strings.HasPrefix(md, "![a](data:image/svg+xml;base64,") {
		t.Fatal(md, err)
	}
	for name, data := range map[string][]byte{
		"page.html": []byte("<html><script>alert(1)</script>"),
		"notes.txt": []byte("hello"),
		"x.xml":     []byte("<?xml version=\"1.0\"?><feed/>"),
		"empty.png": nil,
		"big.png":   append(append([]byte{}, buf.Bytes()...), make([]byte, MaxImage)...),
	} {
		if md, err := ImageMarkdown(name, data); err == nil {
			t.Errorf("%s accepted: %.60s", name, md)
		}
	}
}
