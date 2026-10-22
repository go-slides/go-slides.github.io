package exporter

import (
	"archive/zip"
	"bytes"
	"context"
	"strings"
	"testing"
)

const sample = "---\nslides: true\ntheme: gaia\n---\n\n# One\n\nSome **text** and $e^{i\\pi}$.\n\n---\n\n## Two\n\n- a\n- b\n"

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
			r, err := Export(context.Background(), sample, format, func(done, total int) {
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
	if _, err := Export(context.Background(), sample, "docx", nil); err == nil {
		t.Fatal("docx accepted")
	}
	if _, err := Export(context.Background(), strings.Repeat("x", MaxSource+1), "pdf", nil); err == nil {
		t.Fatal("an oversize deck accepted")
	}
	if _, err := Export(context.Background(), "---\nslides: true\n---\n{{ nope }}", "pdf", nil); err == nil {
		t.Fatal("a broken template exported")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Export(ctx, sample, "pdf", nil); err == nil {
		t.Fatal("a cancelled export finished")
	}
}
