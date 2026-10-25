package exporter

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/exportdoc"
	"github.com/go-slides/slides/doc"
)

const sample = "---\nslides: true\ntheme: gaia\n---\n\n# One\n\nSome **text** and $e^{i\\pi}$.\n\n---\n\n## Two\n\n- a\n- b\n"

// project is a deck with its project and media, as the page holds them.
type project struct {
	exportdoc.Project
	Media map[string][]byte
}

// run is the playground's export: the deck rendered by the preview's
// worker (exportdoc.Build), handed over as JSON, and exported.
func run(ctx context.Context, p project, format string, progress func(done, total int)) (Result, error) {
	d, err := exportdoc.Build(p.Project)
	if err != nil {
		return Result{}, err
	}
	j, err := json.Marshal(d)
	if err != nil {
		return Result{}, err
	}
	var back doc.Doc
	if err := json.Unmarshal(j, &back); err != nil {
		return Result{}, err
	}
	return Export(ctx, Input{Doc: &back, Media: p.Media}, format, progress)
}

func src(s string) project { return project{Project: exportdoc.Project{Source: s}} }

func TestExport(t *testing.T) {
	for format, check := range map[string]func(*testing.T, []byte){
		"pdf": func(t *testing.T, b []byte) {
			// (the outline is in a compressed object stream: not greppable)
			if !bytes.HasPrefix(b, []byte("%PDF-")) || !bytes.HasSuffix(bytes.TrimSpace(b), []byte("%%EOF")) {
				t.Fatalf("not a PDF: %.40q", b)
			}
		},
		"pptx": func(t *testing.T, b []byte) { zipHas(t, b, "ppt/slides/slide2.xml", "Two") },
		"odp":  func(t *testing.T, b []byte) { zipHas(t, b, "content.xml", "Two") },
	} {
		t.Run(format, func(t *testing.T) {
			var seen []int
			r, err := run(context.Background(), src(sample), format, func(done, total int) {
				if total != 2 {
					t.Errorf("total %d", total)
				}
				seen = append(seen, done)
			})
			if err != nil {
				t.Fatal(err)
			}
			if r.Slides != 2 || len(seen) != 2 || seen[1] != 2 || r.Name != Formats[format].Name || r.MIME != Formats[format].MIME {
				t.Fatalf("%+v %v", r.Name, seen)
			}
			check(t, r.Data)
		})
	}
}

func zipHas(t *testing.T, b []byte, name, text string) {
	t.Helper()
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range z.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			var buf bytes.Buffer
			buf.ReadFrom(rc)
			rc.Close()
			if !strings.Contains(buf.String(), text) {
				t.Fatalf("%s lacks %q", name, text)
			}
			return
		}
	}
	t.Fatalf("no %s", name)
}

func TestExportRefuses(t *testing.T) {
	if _, err := run(context.Background(), src(sample), "docx", nil); err == nil {
		t.Fatal("docx accepted")
	}
	if _, err := run(context.Background(), src(strings.Repeat("x", exportdoc.MaxSource+1)), "pdf", nil); err == nil {
		t.Fatal("an oversize deck accepted")
	}
	if _, err := run(context.Background(), src("---\nslides: true\n---\n{{ nope }}"), "pdf", nil); err == nil {
		t.Fatal("a broken template exported")
	}
	if _, err := Export(context.Background(), Input{}, "pdf", nil); err == nil {
		t.Fatal("no deck exported")
	}
	files := func(f map[string]string) project {
		return project{Project: exportdoc.Project{Source: sample, Files: f}}
	}
	media := func(m map[string][]byte) project {
		return project{Project: exportdoc.Project{Source: sample}, Media: m}
	}
	for name, p := range map[string]project{
		"outside":      files(map[string]string{"/etc/passwd": "x"}),
		"traversal":    files(map[string]string{"/project/../x": "x"}),
		"media up":     media(map[string][]byte{"../x.png": nil}),
		"media total":  media(map[string][]byte{"a.png": make([]byte, 33<<20), "b.png": make([]byte, 33<<20)}),
		"text total":   files(map[string]string{"/project/a.css": strings.Repeat("x", 3<<20)}),
		"bad config":   files(map[string]string{"/project/slides.yaml": "kinds: ["}),
		"missing kind": files(map[string]string{"/project/slides.yaml": "kinds: [kinds/nope.yaml]\n"}),
	} {
		if _, err := run(context.Background(), p, "pdf", nil); err == nil {
			t.Errorf("%s: exported", name)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := run(ctx, src(sample), "pdf", nil); err == nil {
		t.Fatal("a cancelled export finished")
	}
}
