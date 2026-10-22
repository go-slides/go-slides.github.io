package playground

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

func zipOf(t testing.TB, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, body := range entries {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte(body))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestZipRoundTrip(t *testing.T) {
	p := ProjectFiles{
		Files: map[string][]byte{
			"deck.md": []byte("# hi\n"), "slides.yaml": []byte("themes: [themes]\n"),
			"themes/a.css": []byte("/* @theme a */"), "media/x.png": {0x89, 'P', 'N', 'G'},
		},
		Dirs: []string{"kinds"},
	}
	at := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	z, err := ExportZip(p, func(string) time.Time { return at })
	if err != nil {
		t.Fatal(err)
	}
	r, _ := zip.NewReader(bytes.NewReader(z), int64(len(z)))
	methods := map[string]uint16{}
	for _, f := range r.File {
		methods[f.Name] = f.Method
	}
	if methods["media/x.png"] != zip.Store || methods["deck.md"] != zip.Deflate || methods["kinds/"] != zip.Store {
		t.Fatalf("methods = %v", methods)
	}
	back, err := ImportZip(z)
	if err != nil {
		t.Fatal(err)
	}
	if len(back.Files) != 4 || string(back.Files["themes/a.css"]) != "/* @theme a */" || len(back.Dirs) != 1 || back.Dirs[0] != "kinds" {
		t.Fatalf("back = %+v", back)
	}
}

func TestImportZipLayouts(t *testing.T) {
	// a zipped folder: the folder is the project; the only .md is the deck
	got, err := ImportZip(zipOf(t, map[string]string{
		"talk/talk.md": "# x", "talk/media/a.png": "png", "talk/": "", "talk/kinds/": "",
		"__MACOSX/talk/._talk.md": "junk", "talk/.DS_Store": "junk",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Files["deck.md"]) != "# x" || got.Files["media/a.png"] == nil || len(got.Files) != 2 {
		t.Fatalf("files = %v", got.Files)
	}
	if len(got.Dirs) != 1 || got.Dirs[0] != "kinds" {
		t.Fatalf("dirs = %v", got.Dirs)
	}
	// two decks and no deck.md: which?
	if _, err := ImportZip(zipOf(t, map[string]string{"a.md": "", "b.md": ""})); err != ErrNoDeck {
		t.Fatalf("err = %v", err)
	}
	// files at the top and in a folder: nothing is stripped
	got, err = ImportZip(zipOf(t, map[string]string{"deck.md": "", "media/a.png": "", "media/": ""}))
	if err != nil || got.Files["media/a.png"] == nil || len(got.Dirs) != 0 {
		t.Fatalf("got %+v, %v", got, err)
	}
	// a folder that is not the top one keeps the files where they are
	if _, err := ImportZip(zipOf(t, map[string]string{"a/deck.md": "", "b/": ""})); err != ErrNoDeck {
		t.Fatalf("err = %v", err)
	}
}

func TestImportZipRefuses(t *testing.T) {
	cases := map[string][]byte{
		"not a zip":     []byte("hello"),
		"traversal":     zipOf(t, map[string]string{"deck.md": "", "../evil": ""}),
		"absolute":      zipOf(t, map[string]string{"deck.md": "", "/etc/passwd": ""}),
		"backslash":     zipOf(t, map[string]string{"deck.md": "", `..\evil`: ""}),
		"dir traversal": zipOf(t, map[string]string{"deck.md": "", "../d/": ""}),
		"twice":         twice(t),
		"too big file":  zipOf(t, map[string]string{"deck.md": strings.Repeat("x", store.MaxFileSize+1)}),
		"symlink":       symlinkZip(t),
	}
	many := map[string]string{"deck.md": ""}
	for i := range store.MaxFiles {
		many["f"+strings.Repeat("x", i%10)+string(rune('a'+i%26))+string(rune('a'+i/26))] = ""
	}
	cases["too many files"] = zipOf(t, many)
	entries := map[string]string{}
	for i := range MaxZipEntries + 1 {
		entries["d"+string(rune('a'+i%26))+string(rune('a'+(i/26)%26))+string(rune('a'+i/676))+"/"] = ""
	}
	cases["too many entries"] = zipOf(t, entries)
	// 9 MiB of zeros in pieces under the file bound: a bomb by ratio
	bomb := map[string]string{"deck.md": ""}
	for i := range 3 {
		bomb["z"+string(rune('a'+i))] = strings.Repeat("\x00", 3<<20)
	}
	cases["zip bomb"] = zipOf(t, bomb)
	for name, z := range cases {
		if _, err := ImportZip(z); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := ImportZip(make([]byte, MaxZipSize+1)); err == nil {
		t.Error("an archive over the size bound was accepted")
	}
}

func twice(t *testing.T) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for range 2 {
		f, _ := w.Create("deck.md")
		f.Write([]byte("x"))
	}
	w.Close()
	return buf.Bytes()
}

func symlinkZip(t *testing.T) []byte {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	h := &zip.FileHeader{Name: "deck.md"}
	h.SetMode(0o777 | 1<<27)
	f, _ := w.CreateHeader(h)
	f.Write([]byte("/etc/passwd"))
	w.Close()
	return buf.Bytes()
}

func TestImportFolder(t *testing.T) {
	got, err := ImportFolder(map[string][]byte{"p/deck.md": []byte("x"), "p/themes/a.css": nil, "p/.DS_Store": nil})
	if err != nil || len(got.Files) != 2 {
		t.Fatalf("got %+v, %v", got, err)
	}
	if _, ok := got.Files["themes/a.css"]; !ok {
		t.Fatalf("files = %v", got.Files)
	}
	for name, in := range map[string]map[string][]byte{
		"traversal": {"deck.md": nil, "../x": nil},
		"big":       {"deck.md": make([]byte, store.MaxFileSize+1)},
	} {
		if _, err := ImportFolder(in); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	many := map[string][]byte{}
	for i := range store.MaxFiles + 1 {
		many[strings.Repeat("a", i%5+1)+string(rune('a'+i%26))+string(rune('a'+i/26))+".md"] = nil
	}
	if _, err := ImportFolder(many); err == nil {
		t.Error("too many files accepted")
	}
	huge := map[string][]byte{}
	for i := range 9 {
		huge[string(rune('a'+i))] = make([]byte, store.MaxFileSize)
	}
	if _, err := ImportFolder(huge); err == nil {
		t.Error("a project over the bound accepted")
	}
}

// The importer never panics, and what it accepts is inside the bounds.
func FuzzImportZip(f *testing.F) {
	f.Add(zipOf(f, map[string]string{"deck.md": "# x", "media/a.png": "png"}))
	f.Add(zipOf(f, map[string]string{"p/x.md": "", "p/kinds/": ""}))
	f.Add(zipOf(f, map[string]string{"../evil": "", "deck.md": ""}))
	f.Add([]byte("PK\x03\x04"))
	f.Fuzz(func(t *testing.T, data []byte) {
		p, err := ImportZip(data)
		if err != nil {
			return
		}
		total := 0
		for k, v := range p.Files {
			if c, err := store.CleanPath(k); err != nil || c != k {
				t.Fatalf("path %q accepted", k)
			}
			total += len(v)
		}
		if len(p.Files) > store.MaxFiles || total > store.MaxProjectSize {
			t.Fatalf("bounds: %d files, %d bytes", len(p.Files), total)
		}
		if _, ok := p.Files[store.DeckPath]; !ok {
			t.Fatal("no deck")
		}
	})
}
