package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
)

func site(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	for name, body := range map[string]string{
		"index.html":                   strings.Repeat("<p>landing</p>", 200),
		"playground/index.html":        strings.Repeat("<p>playground</p>", 200),
		"playground/export.wasm":       strings.Repeat("\x00asm wasm", 500),
		"playground/cards/1.jpg":       "jpeg bytes, not compressed again",
		"playground/fonts/inter/a.ttf": strings.Repeat("glyf", 400),
		"tiny.svg":                     "<svg/>",
	} {
		p := filepath.Join(d, filepath.FromSlash(name))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

func read(t *testing.T, file string) map[string][]byte {
	t.Helper()
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]byte{}
	tr := tar.NewReader(zr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return out
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Uid != 0 || h.Gid != 0 || h.Uname != "" || !h.ModTime.Equal(time.Unix(1700000000, 0)) {
			t.Fatalf("%s: owner %d/%d %q, time %v", h.Name, h.Uid, h.Gid, h.Uname, h.ModTime)
		}
		b, _ := io.ReadAll(tr)
		out[h.Name] = b
	}
}

// The same site gives the same bytes; each compressible file has its .br and
// .gz, which decompress to it; the others have none.
func TestBundle(t *testing.T) {
	s := site(t)
	a, b := filepath.Join(t.TempDir(), "a.tar.gz"), filepath.Join(t.TempDir(), "b.tar.gz")
	at := time.Unix(1700000000, 0)
	sa, n, err := Bundle(s, "go-slides-site-v1", a, at)
	if err != nil {
		t.Fatal(err)
	}
	sb, _, err := Bundle(s, "go-slides-site-v1", b, at)
	if err != nil {
		t.Fatal(err)
	}
	ba, _ := os.ReadFile(a)
	bb, _ := os.ReadFile(b)
	if sa != sb || !bytes.Equal(ba, bb) {
		t.Fatal("two bundles of one site differ")
	}
	files := read(t, a)
	if n == 0 || len(files) < n {
		t.Fatalf("%d entries, %d read", n, len(files))
	}
	wasm := files["go-slides-site-v1/playground/export.wasm"]
	br, err := io.ReadAll(brotli.NewReader(bytes.NewReader(files["go-slides-site-v1/playground/export.wasm.br"])))
	if err != nil || !bytes.Equal(br, wasm) || len(wasm) == 0 {
		t.Fatalf("export.wasm.br: %v", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(files["go-slides-site-v1/playground/fonts/inter/a.ttf.gz"]))
	if err != nil {
		t.Fatal(err)
	}
	if gz, _ := io.ReadAll(zr); !bytes.Equal(gz, files["go-slides-site-v1/playground/fonts/inter/a.ttf"]) {
		t.Fatal("a.ttf.gz")
	}
	for _, none := range []string{"playground/cards/1.jpg.br", "playground/cards/1.jpg.gz", "tiny.svg.br"} {
		if _, ok := files["go-slides-site-v1/"+none]; ok {
			t.Errorf("%s: a precompressed form that is not smaller, or of an image", none)
		}
	}
	if _, ok := files["go-slides-site-v1/index.html.br"]; !ok {
		t.Error("no index.html.br")
	}
}

func TestBundleRefuses(t *testing.T) {
	s := site(t)
	out := filepath.Join(t.TempDir(), "x.tar.gz")
	if _, _, err := Bundle(s, "a/b", out, time.Unix(0, 0)); err == nil {
		t.Error("a top directory with a slash")
	}
	if _, _, err := Bundle(t.TempDir(), "x", out, time.Unix(0, 0)); err == nil {
		t.Error("an empty site")
	}
	if err := os.Symlink("/etc/passwd", filepath.Join(s, "link")); err == nil {
		if _, _, err := Bundle(s, "x", out, time.Unix(0, 0)); err == nil || !strings.Contains(err.Error(), "symbolic link") {
			t.Errorf("a symbolic link: %v", err)
		}
	}
}

func TestCommand(t *testing.T) {
	s := site(t)
	out := filepath.Join(t.TempDir(), "c.tar.gz")
	t.Setenv("SOURCE_DATE_EPOCH", "1700000000")
	c := RootCmd()
	var stdout bytes.Buffer
	c.SetOut(&stdout)
	c.SetArgs([]string{"--site", s, "--name", "go-slides-site-v1", "--out", out})
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "c.tar.gz") || len(read(t, out)) == 0 {
		t.Fatal(stdout.String())
	}
	t.Setenv("SOURCE_DATE_EPOCH", "soon")
	c = RootCmd()
	c.SetArgs([]string{"--site", s, "--out", out})
	if err := c.Execute(); err == nil {
		t.Fatal("a bad SOURCE_DATE_EPOCH was taken")
	}
}
