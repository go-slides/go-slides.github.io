package playground

import (
	"archive/zip"
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

// filesTab is a page that also saves bytes and lets the reader pick files.
type filesTab struct {
	*tab
	saved  map[string][]byte
	picked []string
}

func (f *filesTab) DownloadBytes(name, _ string, data []byte) { f.saved[name] = data }
func (f *filesTab) Pick(what string)                          { f.picked = append(f.picked, what) }

func openFiles(t *testing.T, m *store.Memory) (*ViewModel, *filesTab, *fakeTransport) {
	t.Helper()
	h := &filesTab{tab: &tab{fakeHost: newHost(), answer: true}, saved: map[string][]byte{}}
	h.st = store.New(m, NewTabID(), load(m, DraftProject), load(m, PreviousProject), load(m, UIProject))
	tr := &fakeTransport{}
	vm := NewViewModel(tr, h, "")
	tr.vm = vm
	answer(tr, vm)
	return vm, h, tr
}

// png is a 1×1 PNG.
var png = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\rIDATx\xdac\xfc\xcf\xc0P\x0f\x00\x04\x85\x01\x80\x84\xa9\x8c!\x00\x00\x00\x00IEND\xaeB`\x82")

func paths(vm *ViewModel) []string {
	var out []string
	for _, f := range vm.Files.Get() {
		p := f.Path
		if f.Dir {
			p += "/"
		}
		out = append(out, p)
	}
	return out
}

func TestPanelListsTheProject(t *testing.T) {
	vm, _, _ := openFiles(t, store.NewMemory())
	if got := paths(vm); !slices.Equal(got, []string{"deck.md", "kinds/", "media/", "themes/"}) {
		t.Fatalf("files = %q", got)
	}
	if !strings.HasPrefix(vm.Usage.Get(), "1 file · ") || !strings.HasSuffix(vm.Usage.Get(), "of 64 MiB") {
		t.Fatalf("usage = %q", vm.Usage.Get())
	}
}

// deckSize is the panel's size of deck.md.
func deckSize(vm *ViewModel) int64 {
	for _, f := range vm.Files.Get() {
		if f.Path == DeckFile {
			return f.Size
		}
	}
	return -1
}

// TestDeckSizeFollowsTheText: the panel's deck size and the usage line are
// the text's after the save tick, whatever replaced it (typing, a ribbon
// command, an example, Reset, an import), with a store and without one.
func TestDeckSizeFollowsTheText(t *testing.T) {
	vm, h, _ := openFiles(t, store.NewMemory())
	for _, change := range []func(){
		func() { vm.Source.Set("# x\n") },                    // typed
		func() { vm.Source.Set(vm.Source.Get() + "more\n") }, // a command's edit
		func() { vm.Example.Set(1) },                         // an example
		func() { vm.Reset.Execute() },                        // Reset
		func() { vm.OpenZip(zipOf(t, map[string]string{"deck.md": "# zipped\n"})) },
	} {
		h.answer = true
		change()
		vm.SaveDraft() // the host's save tick
		n := len(vm.Source.Get())
		if deckSize(vm) != int64(n) || !strings.Contains(vm.Usage.Get(), sizeText(n)+" of") {
			t.Fatalf("deck %d bytes, panel %d, usage %q", n, deckSize(vm), vm.Usage.Get())
		}
	}
	// no store: the tick still refreshes
	tr := &fakeTransport{}
	vm2 := NewViewModel(tr, newHost(), "")
	tr.vm = vm2
	vm2.Source.Set("# tiny\n")
	vm2.SaveDraft()
	if deckSize(vm2) != int64(len("# tiny\n")) {
		t.Fatalf("no store: panel %d", deckSize(vm2))
	}
}

func TestNewThemeJoinsThePickerAndRenders(t *testing.T) {
	vm, _, tr := openFiles(t, store.NewMemory())
	vm.ProjectAction.Set(2) // New theme (from gaia)
	if vm.ProjectAction.Get() != 0 {
		t.Fatal("the menu did not return to its title")
	}
	if vm.Open.Get() != "themes/my-gaia.css" || !strings.Contains(vm.FileText.Get(), "@import 'gaia'") {
		t.Fatalf("open %q: %q", vm.Open.Get(), vm.FileText.Get())
	}
	if cfg := vm.project().Text(ConfigPath); !strings.Contains(cfg, "themes: [themes]") {
		t.Fatalf("slides.yaml = %q", cfg)
	}
	i := slices.Index(vm.Themes, "my-gaia")
	if i != len(BuiltinThemes) {
		t.Fatalf("themes = %q", vm.Themes)
	}
	vm.Theme.Set(i)
	if th, _ := DeckTheme(vm.Source.Get()); th != "my-gaia" {
		t.Fatalf("deck theme = %q", th)
	}
	answer(tr, vm)
	last := tr.sent[len(tr.sent)-1]
	if last.Files["/playground/themes/my-gaia.css"] == "" || last.Files["/playground/slides.yaml"] == "" {
		t.Fatalf("the request carries %v", last.Files)
	}
	if len(vm.Diagnostics) != 0 || strings.Contains(vm.Status.Get(), "⚠") {
		t.Fatalf("diagnostics %+v, status %q", vm.Diagnostics, vm.Status.Get())
	}
	// Editing the theme re-renders with it; a second one gets a number.
	vm.FileText.Set(vm.FileText.Get() + "\nh2 { color: red; }\n")
	if !strings.Contains(vm.project().Text("themes/my-gaia.css"), "h2 { color: red; }") {
		t.Fatal("the edit was not kept")
	}
	vm.NewTheme("nope")
	if vm.Open.Get() != "themes/my-default.css" {
		t.Fatalf("open = %q", vm.Open.Get())
	}
	vm.NewTheme("gaia")
	if vm.Open.Get() != "themes/my-gaia-2.css" {
		t.Fatalf("open = %q", vm.Open.Get())
	}
}

func TestNewKindIsListedAndUsed(t *testing.T) {
	vm, _, tr := openFiles(t, store.NewMemory())
	vm.ProjectAction.Set(4)
	if vm.Open.Get() != "kinds/chapter.yaml" || !strings.Contains(vm.project().Text(ConfigPath), "kinds: [kinds/chapter.yaml]") {
		t.Fatalf("open %q, config %q", vm.Open.Get(), vm.project().Text(ConfigPath))
	}
	vm.OpenFile(DeckFile)
	vm.Source.Set("---\nslides: true\n---\n\n<!-- _kind: chapter -->\n# Part one\n")
	answer(tr, vm)
	r := Serve(tr.sent[len(tr.sent)-1])
	if len(r.Diagnostics) != 0 || !strings.Contains(strings.Join(r.Parts, ""), "invert") {
		t.Fatalf("diagnostics %+v", r.Diagnostics)
	}
	// A broken kinds file is a problem named by its file.
	vm.OpenFile("kinds/chapter.yaml")
	vm.FileText.Set("chapter:\n  nope: 1\n")
	answer(tr, vm)
	if !strings.Contains(vm.Problems.Get(), "kinds/chapter.yaml") || strings.Contains(vm.Problems.Get(), "/playground/") {
		t.Fatalf("problems = %q", vm.Problems.Get())
	}
	// Renamed, then deleted: slides.yaml follows.
	if err := vm.RenamePath("kinds/chapter.yaml", "kinds/part.yaml"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(vm.project().Text(ConfigPath), "kinds/part.yaml") || vm.Open.Get() != "kinds/part.yaml" {
		t.Fatalf("config %q, open %q", vm.project().Text(ConfigPath), vm.Open.Get())
	}
	if err := vm.DeletePath("kinds/part.yaml"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(vm.project().Text(ConfigPath), "part.yaml") || vm.Open.Get() != DeckFile {
		t.Fatalf("config %q, open %q", vm.project().Text(ConfigPath), vm.Open.Get())
	}
	vm.NewDataFile()
	if vm.Open.Get() != "data/values.yaml" || !strings.Contains(vm.project().Text(ConfigPath), "data: [data/values.yaml]") {
		t.Fatalf("data: %q", vm.project().Text(ConfigPath))
	}
}

func TestMediaAreFilesNotText(t *testing.T) {
	m := store.NewMemory()
	vm, h, _ := openFiles(t, m)
	v0 := vm.MediaVersion.Get()
	p, err := vm.AddFile("", "My photo.png", png)
	if err != nil || p != "media/My photo.png" {
		t.Fatalf("%q, %v", p, err)
	}
	if vm.MediaVersion.Get() == v0 || len(vm.Media()) != 1 || vm.Modified(p) == 0 {
		t.Fatal("the media did not change")
	}
	if got := MediaMarkdown(p); got != "![My photo](media/My%20photo.png)" {
		t.Fatalf("markdown = %q", got)
	}
	if again, _ := vm.AddFile("", "My photo.png", png); again != p {
		t.Fatalf("the same bytes again: %q", again)
	}
	if other, _ := vm.AddFile("", "My photo.png", append(slices.Clone(png), 0)); other != "media/My photo-2.png" {
		t.Fatalf("other bytes: %q", other)
	}
	if _, err := vm.AddFile("", "big.png", make([]byte, store.MaxFileSize+1)); err == nil {
		t.Fatal("a file over the bound was taken")
	}
	if _, err := vm.AddFile("", "..", png); err == nil {
		t.Fatal("a name that is no name was taken")
	}
	// A medium is not opened as text.
	vm.OpenFile(p)
	if vm.Open.Get() != DeckFile || !strings.Contains(vm.Status.Get(), "is a medium") {
		t.Fatalf("open %q, status %q", vm.Open.Get(), vm.Status.Get())
	}
	// The deck names it; the downloaded page holds it.
	vm.Source.Set("---\nslides: true\n---\n\n![bg](media/My%20photo.png)\n")
	vm.Download.Execute()
	page := h.downloads["deck.html"]
	if !strings.Contains(page, "data:image/png;base64,") || strings.Contains(page, `url("media/`) {
		t.Fatalf("the page does not hold the medium: %.200s", page)
	}
	// Kept across a reload, with the draft.
	vm.SaveDraft()
	vm2, _, _ := openFiles(t, m)
	if !slices.Contains(paths(vm2), "media/My photo.png") {
		t.Fatalf("after a reload: %q", paths(vm2))
	}
}

func TestFoldersAndRefusals(t *testing.T) {
	vm, _, _ := openFiles(t, store.NewMemory())
	if err := vm.CreateFolder("notes"); err != nil || !slices.Contains(paths(vm), "notes/") {
		t.Fatalf("%v, %q", err, paths(vm))
	}
	if err := vm.CreateFolder("notes"); err == nil {
		t.Fatal("a folder twice")
	}
	if err := vm.CreateFromPanel("notes/a.md"); err != nil {
		t.Fatal(err)
	}
	if err := vm.RenamePath("notes", "docs"); err != nil {
		t.Fatal(err)
	}
	if got := paths(vm); !slices.Contains(got, "docs/a.md") || slices.Contains(got, "notes/") {
		t.Fatalf("after the rename: %q", got)
	}
	if err := vm.RenamePath("media", "pictures"); err != nil || !slices.Contains(paths(vm), "pictures/") {
		t.Fatalf("%v %q", err, paths(vm))
	}
	for name, err := range map[string]error{
		"rename the deck":  vm.RenamePath(DeckFile, "x.md"),
		"onto the deck":    vm.RenamePath("docs/a.md", DeckFile),
		"delete the deck":  vm.DeletePath(DeckFile),
		"no such file":     vm.RenamePath("nope", "x"),
		"leaves":           vm.RenamePath("docs/a.md", "../x"),
		"exists":           vm.RenamePath("docs/a.md", "docs"),
		"create bad":       vm.CreateFile("/etc/x", ""),
		"create existing":  vm.CreateFile("docs/a.md", ""),
		"folder bad":       vm.CreateFolder(".."),
		"theme from panel": vm.CreateFromPanel("themes/x.css"),
	} {
		if (err == nil) != (name == "theme from panel") {
			t.Errorf("%s: %v", name, err)
		}
	}
	if !strings.Contains(vm.project().Text("themes/x.css"), "@theme x") {
		t.Fatal("a theme made in the panel starts from the template")
	}
	if err := vm.DeletePath("docs"); err != nil || slices.Contains(paths(vm), "docs/a.md") {
		t.Fatalf("%v %q", err, paths(vm))
	}
	// slides.hcl is the project's: it is not edited.
	vm.project().Write("slides.hcl", []byte("lang = \"fr\"\n"), vm.drafts.now())
	vm.NewKind()
	if strings.Contains(vm.project().Text(ConfigPath), "kinds") {
		t.Fatal("slides.yaml was edited beside slides.hcl")
	}
}

func TestZipOutAndIn(t *testing.T) {
	m := store.NewMemory()
	vm, h, _ := openFiles(t, m)
	vm.NewTheme("uncover")
	vm.AddFile("", "dot.png", png)
	vm.CreateFolder("empty")
	vm.OpenFile(DeckFile)
	vm.Source.Set("---\ntheme: my-uncover\n---\n\n![bg](media/dot.png)\n")
	vm.ProjectAction.Set(11) // Download project (.zip)
	z := h.saved[ZipName]
	r, err := zip.NewReader(bytes.NewReader(z), int64(len(z)))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	want := []string{"deck.md", "media/dot.png", "slides.yaml", "themes/my-uncover.css", "empty/", "kinds/"}
	if !slices.Equal(names, want) {
		t.Fatalf("zip = %q", names)
	}
	// Opened again in a fresh page: the same project.
	vm2, h2, _ := openFiles(t, store.NewMemory())
	vm2.OpenZip(z)
	if h2.asked != 1 || vm2.Source.Get() != vm.Source.Get() || !slices.Contains(paths(vm2), "themes/my-uncover.css") || !slices.Contains(paths(vm2), "empty/") {
		t.Fatalf("asked %d, files %q", h2.asked, paths(vm2))
	}
	if !slices.Contains(vm2.Themes, "my-uncover") {
		t.Fatalf("themes %q", vm2.Themes)
	}
	// Refused, the project stays.
	h2.answer = false
	vm2.OpenZip(zipOf(t, map[string]string{"deck.md": "other"}))
	if vm2.Source.Get() == "other" {
		t.Fatal("replaced without consent")
	}
	vm2.OpenZip([]byte("not a zip"))
	if !strings.Contains(vm2.Status.Get(), "not a .zip") {
		t.Fatalf("status %q", vm2.Status.Get())
	}
	h2.answer = true
	vm2.OpenFolder(map[string][]byte{"talk/talk.md": []byte("# folder"), "talk/media/a.png": png})
	if vm2.Source.Get() != "# folder" || !slices.Contains(paths(vm2), "media/a.png") {
		t.Fatalf("folder: %q", paths(vm2))
	}
	vm2.OpenFolder(map[string][]byte{"../x": nil})
	vm2.OpenZip(zipOf(t, map[string]string{"deck.md": strings.Repeat("x", MaxSource+1)}))
	if vm2.Source.Get() != "# folder" {
		t.Fatal("a deck over the bound replaced the project")
	}
	// The pickers are the host's.
	for i := 8; i <= 10; i++ {
		vm2.ProjectAction.Set(i)
	}
	if !slices.Equal(h2.picked, []string{"media", "zip", "folder"}) {
		t.Fatalf("picked %q", h2.picked)
	}
	vm2.ProjectAction.Set(6)
	vm2.ProjectAction.Set(7)
	if a := vm2.CreateAsked.Get(); a.N != 2 || !a.Folder {
		t.Fatalf("create asks %+v", a)
	}
	vm2.DownloadFile("media/a.png")
	vm2.DownloadFile(DeckFile)
	if h2.saved["a.png"] == nil || string(h2.saved["deck.md"]) != "# folder" {
		t.Fatal("DownloadFile")
	}
}

func TestShareSaysTheLinkHoldsTheDeckOnly(t *testing.T) {
	vm, h, _ := openFiles(t, store.NewMemory())
	vm.Share.Execute()
	if h.asked != 0 || h.fragment == "" {
		t.Fatal("a deck alone is shared without a question")
	}
	vm.AddFile("", "dot.png", png)
	vm.Share.Execute()
	if h.asked != 1 || h.saved[ZipName] == nil {
		t.Fatalf("asked %d", h.asked)
	}
	h.answer = false
	vm.Share.Execute()
	if !strings.Contains(vm.Status.Get(), "Only the deck is in the link") {
		t.Fatalf("status %q", vm.Status.Get())
	}
}

func TestServeTakesOnlyTheProject(t *testing.T) {
	for _, files := range []map[string]string{
		{"/etc/passwd": "x"}, {"/playground/../etc/x": "x"}, {DeckPath: "x"},
		{"/playground/big": strings.Repeat("x", MaxTextFiles+1)},
	} {
		if r := Serve(Request{Source: "# x", Files: files}); r.Error == "" {
			t.Errorf("%v accepted", files)
		}
	}
	if _, err := ProjectPresentation("# x", map[string]string{"/x": ""}, nil); err == nil {
		t.Error("ProjectPresentation took a file outside the project")
	}
}

func TestInlineMediaAndScript(t *testing.T) {
	page := `<img src="media/a.png"><figure style="background-image:url(&quot;./media/a.png&quot;)"></figure><a href="media/a.png.bak">`
	got := InlineMedia(page, map[string][]byte{"media/a.png": png})
	if strings.Count(got, "data:image/png;base64,") != 2 || !strings.Contains(got, `"media/a.png.bak"`) {
		t.Fatalf("inlined: %s", got)
	}
	if InlineMedia(page, nil) != page {
		t.Fatal("nothing to inline changed the page")
	}
	if MediaType("x.png", []byte("\xff\xd8\xff\xe0")) != "image/jpeg" || MediaType("x.bin", nil) != "application/octet-stream" || MediaType("v.mp4", nil) != "video/mp4" {
		t.Fatal("MediaType")
	}
	if MediaMarkdown("media/v.mp4") != `<video src="media/v.mp4" controls></video>` || MediaMarkdown("media/s.mp3") != `<audio src="media/s.mp3" controls></audio>` {
		t.Fatal("MediaMarkdown")
	}
	r := Serve(Request{Source: "# x", Mode: ModeSlideshow})
	if !strings.Contains(r.Doc, "goSlidesMedia") || !strings.Contains(r.Doc, "media-src https: data: blob:") {
		t.Fatal("the slideshow has no media script")
	}
}

func TestFileSyntax(t *testing.T) {
	th := LightPalette.Theme()
	for p, lang := range map[string]string{"a.md": "Markdown", "a.yml": "YAML", "a.hcl": "HCL", "a.json": "JSON", "a.css": "CSS", "a.txt": "Text"} {
		if FileLanguage(p) != lang {
			t.Errorf("%s: %s", p, FileLanguage(p))
		}
		if s := fileSyntax(p); (s == nil) != (lang == "Text") {
			t.Errorf("%s: highlighter %v", p, s)
		}
	}
	css := cssSyntax{}.Highlight("", []string{"/* @theme x", "*/ @import 'default';", "section {", "  color: red; /* c */", "}"}, th)
	if len(css[1]) < 2 || len(css[3]) < 2 {
		t.Fatalf("css spans %+v", css)
	}
	if k := (keysSyntax{}).Highlight("", []string{"a: 1 # c"}, th); len(k[0]) != 2 {
		t.Fatalf("yaml spans %+v", k)
	}
}
