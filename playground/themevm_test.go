package playground

import (
	"bytes"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-opentype/fonts/jetbrainsmono"
	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-slides/slides/themepack"
)

func themeVM(t *testing.T) (*ViewModel, *filesTab, *fakeTransport) {
	t.Helper()
	SetLocale("en")
	return openFiles(t, store.NewMemory())
}

func TestNewThemeInEditorMakesSourceSheetAndTheDecksTheme(t *testing.T) {
	vm, _, tr := themeVM(t)
	vm.NewThemeInEditor("gaia")
	name := vm.ThemeName.Get()
	if name != "my-gaia" || !vm.ThemePane.Get() {
		t.Fatalf("theme %q, pane %v", name, vm.ThemePane.Get())
	}
	p := vm.project()
	for _, f := range []string{"theme-src/my-gaia/theme.hcl", "theme-src/my-gaia/theme.css", "themes/my-gaia.css"} {
		if _, ok := p.Read(f); !ok {
			t.Errorf("%s is missing: %q", f, paths(vm))
		}
	}
	if got, _ := DeckTheme(vm.Source.Get()); got != "my-gaia" {
		t.Errorf("the deck's theme is %q", got)
	}
	if css := p.Text("themes/my-gaia.css"); !strings.Contains(css, `@import-theme "gaia"`) || !strings.Contains(css, "@theme my-gaia") {
		t.Errorf("compiled sheet:\n%s", css)
	}
	// the renderer is sent the compiled sheet, not the source folder's sheet it has no use for
	last := tr.sent[len(tr.sent)-1]
	if _, ok := last.Files[projectRoot+"/themes/my-gaia.css"]; !ok {
		t.Errorf("the preview is not sent the theme: %v", slices.Collect(mapsKeys(last.Files)))
	}
	if !strings.Contains(vm.project().Text(ConfigPath), "themes") {
		t.Errorf("slides.yaml: %q", vm.project().Text(ConfigPath))
	}
	// the new theme is one undo step
	vm.undoStep()
	if _, ok := vm.project().Read("theme-src/my-gaia/theme.hcl"); ok {
		t.Error("undo left the source behind")
	}
	if got, _ := DeckTheme(vm.Source.Get()); got == "my-gaia" {
		t.Error("undo left the deck on the theme")
	}
}

func mapsKeys[M ~map[string]V, V any](m M) func(yield func(string) bool) {
	return func(yield func(string) bool) {
		for k := range m {
			if !yield(k) {
				return
			}
		}
	}
}

func TestEditingTheControlsRewritesTheSheetThePreviewReads(t *testing.T) {
	vm, _, tr := themeVM(t)
	vm.NewThemeInEditor("default")
	before := vm.project().Text("themes/my-default.css")

	vm.SetThemeValue("background", "#102030")
	vm.SetThemeValue("font-body", "Georgia")
	css := vm.project().Text("themes/my-default.css")
	if css == before || !strings.Contains(css, "--gsta-background: #102030;") || !strings.Contains(css, `--gsta-font-body: "Georgia";`) {
		t.Fatalf("the compiled sheet did not follow:\n%s", css)
	}
	if got := tr.sent[len(tr.sent)-1].Files[projectRoot+"/themes/my-default.css"]; got != css {
		t.Error("the preview was not sent the new sheet")
	}
	v, st, ok := vm.ThemeValues()
	if !ok || st != BlockInSync || v["background"] != "#102030" {
		t.Fatalf("%v %v %v", v, st, ok)
	}
	// the sheet really renders: the colour reaches the page
	r := Serve(Request{ID: 9, Source: vm.Source.Get(), Files: mustText(vm)})
	if r.Error != "" || !strings.Contains(r.Page(), "#102030") {
		t.Errorf("render: %q; the colour is not in the page", r.Error)
	}
	if len(r.Diagnostics) != 0 {
		t.Errorf("diagnostics: %+v", r.Diagnostics)
	}
}

func mustText(vm *ViewModel) map[string]string {
	files, _ := vm.textFiles()
	return files
}

func TestEditsOfOneControlInARowAreOneUndoStep(t *testing.T) {
	vm, _, _ := themeVM(t)
	vm.NewThemeInEditor("default")
	vm.undo.project.Clear()
	clock := time.Unix(1_700_000_000, 0)
	vm.drafts.now = func() time.Time { return clock }
	for _, c := range []string{"#111111", "#222222", "#333333"} {
		clock = clock.Add(100 * time.Millisecond)
		vm.SetThemeValue("accent", c)
	}
	clock = clock.Add(100 * time.Millisecond)
	vm.SetThemeValue("link", "#0000ee") // another control: its own step
	if vm.UndoLabel.Get() == "" {
		t.Fatalf("no step: %q", vm.UndoLabel.Get())
	}
	vm.undoStep() // the link
	if v, _, _ := vm.ThemeValues(); v["link"] != "" || v["accent"] != "#333333" {
		t.Fatalf("after the first undo: %v", v)
	}
	vm.undoStep() // the three accents together
	if v, _, _ := vm.ThemeValues(); len(v) != 0 {
		t.Fatalf("after the second undo the drag is gone as one step: %v", v)
	}
	if vm.undo.project.CanUndo() {
		t.Error("two steps were expected, a third is left")
	}
	vm.redoStep()
	if v, _, _ := vm.ThemeValues(); v["accent"] != "#333333" {
		t.Errorf("redo: %v", v)
	}
	// a pause longer than the window is a new step
	clock = clock.Add(burstWindow + time.Second)
	vm.SetThemeValue("accent", "#444444")
	clock = clock.Add(10 * time.Second)
	vm.SetThemeValue("accent", "#555555")
	vm.undoStep()
	if v, _, _ := vm.ThemeValues(); v["accent"] != "#444444" {
		t.Errorf("a pause separates the steps: %v", v)
	}
}

func TestAHandEditOfTheSheetIsReadBackAndCompiled(t *testing.T) {
	vm, _, _ := themeVM(t)
	vm.NewThemeInEditor("default")
	vm.SetThemeValue("text", "#abcdef")
	vm.OpenThemeSheet()
	if vm.Open.Get() != "theme-src/my-default/theme.css" {
		t.Fatalf("open = %q", vm.Open.Get())
	}
	rev := vm.ThemeRev.Get()
	// the person edits the sheet in the editor, inside the block (a value) and outside it
	text := vm.FileText.Get()
	text = strings.Replace(text, "--gsta-text: #abcdef;", "--gsta-text: #fedcba;", 1)
	vm.FileText.Set("h2 { color: teal }\n" + text)
	v, st, _ := vm.ThemeValues()
	if v["text"] != "#fedcba" || st != BlockInSync {
		t.Fatalf("the controls did not follow the sheet: %v %v", v, st)
	}
	if vm.ThemeRev.Get() == rev {
		t.Error("the pane is not told")
	}
	css := vm.project().Text("themes/my-default.css")
	if !strings.Contains(css, "#fedcba") || !strings.Contains(css, "h2 { color: teal }") {
		t.Errorf("the compiled sheet did not follow the hand edit:\n%s", css)
	}

	// a rule added inside the fence: the controls stop writing, and say so
	vm.FileText.Set(strings.Replace(vm.FileText.Get(), visualEnd, "h1 { text-decoration: underline }\n"+visualEnd, 1))
	if _, st, _ := vm.ThemeValues(); st != BlockDiverged {
		t.Fatalf("state %v", st)
	}
	if !strings.Contains(vm.ThemeNote.Get(), "gsta:visual") {
		t.Errorf("the pane says nothing: %q", vm.ThemeNote.Get())
	}
	kept := vm.FileText.Get()
	vm.SetThemeValue("link", "#00ff00")
	if vm.project().Text("theme-src/my-default/theme.css") != kept {
		t.Error("the controls overwrote CSS the person wrote")
	}
	if !strings.Contains(vm.Status.Get(), "edited by hand") {
		t.Errorf("status = %q", vm.Status.Get())
	}
	vm.RestoreThemeBlock()
	if _, st, _ := vm.ThemeValues(); st != BlockInSync || vm.ThemeNote.Get() != "" {
		t.Errorf("after restore: %v %q", st, vm.ThemeNote.Get())
	}
	if strings.Contains(vm.project().Text("theme-src/my-default/theme.css"), "underline") || !strings.Contains(vm.project().Text("theme-src/my-default/theme.css"), "h2 { color: teal }") {
		t.Error("restore touched more than the block, or less")
	}
}

func TestMetadataIsEditedAndRefusedByTheLibrary(t *testing.T) {
	vm, _, _ := themeVM(t)
	vm.NewThemeInEditor("default")
	good := vm.project().Text("themes/my-default.css")
	vm.SetThemeMeta("version", "not-a-version")
	if !strings.Contains(vm.Problems.Get(), "semantic version") || !strings.Contains(vm.Problems.Get(), "theme-src/my-default") {
		t.Fatalf("problems = %q", vm.Problems.Get())
	}
	if vm.project().Text("themes/my-default.css") != good {
		t.Error("a theme the library refuses replaced the sheet that compiled")
	}
	if strings.Contains(vm.project().Text("theme-src/my-default/theme.hcl"), "not-a-version") {
		t.Error("an edit the library refuses is not stored: the controls could no longer read the theme")
	}
	if !strings.Contains(vm.Status.Get(), "semantic version") {
		t.Errorf("status %q", vm.Status.Get())
	}
	SetLocale("fr")
	vm.SetThemeMeta("version", "still-not")
	if !strings.Contains(vm.Problems.Get(), "Le theme.hcl du paquet de thème") {
		t.Errorf("the code is said in the language: %q", vm.Problems.Get())
	}
	SetLocale("en")
	vm.SetThemeMeta("version", "2.1.0")
	vm.SetThemeMeta("authors", "Ada <ada@example.org>; Bo")
	vm.SetThemeMeta("description", "Mine")
	vm.SetThemeMeta("license", "Apache-2.0")
	vm.SetThemeMeta("homepage", "https://example.org/t")
	vm.SetThemeMeta("min_slides", "0.17.0")
	vm.SetThemeMeta("extends", "gaia")
	if vm.Problems.Get() != "" {
		t.Errorf("problems = %q", vm.Problems.Get())
	}
	m, _, _ := vm.ThemeMetaValues()
	if m["version"] != "2.1.0" || m["authors"] != "Ada <ada@example.org>; Bo" || m["extends"] != "gaia" || m["license"] != "Apache-2.0" || m["name"] != "my-default" {
		t.Errorf("meta = %v", m)
	}
	if css := vm.project().Text("themes/my-default.css"); !strings.Contains(css, `@import-theme "gaia"`) || !strings.Contains(css, "@version 2.1.0") {
		t.Errorf("sheet:\n%s", css)
	}
}

func TestImportingAPackageKeepsItsSourceAndOpensTheEditor(t *testing.T) {
	vm, _, _ := themeVM(t)
	vm.ThemeImportFile("acme.gsta.xz", pack(t, acmeFiles()))
	if vm.ThemeName.Get() != "acme" || !vm.ThemePane.Get() {
		t.Fatalf("theme %q pane %v; status %q", vm.ThemeName.Get(), vm.ThemePane.Get(), vm.Status.Get())
	}
	for n, b := range acmeFiles() {
		got, ok := vm.project().Read("theme-src/acme/" + n)
		if !ok || !bytes.Equal(got, b) {
			t.Errorf("theme-src/acme/%s is not the package's own file", n)
		}
	}
	// the same package again changes nothing and still opens the editor
	vm.ThemeName.Set("")
	vm.ThemeImportFile("acme.gsta.xz", pack(t, acmeFiles()))
	if vm.ThemeName.Get() != "acme" {
		t.Errorf("again: %q", vm.ThemeName.Get())
	}
}

func TestImportingARefusedPackageShowsTheCodeInProblems(t *testing.T) {
	vm, _, _ := themeVM(t)
	before := paths(vm)
	bad := pack(t, with(acmeFiles(), "theme.css", "h1{background:url(https://evil.example/x.png)}"))
	vm.ThemeImportFile("evil.gsta.xz", bad)
	if vm.ThemeName.Get() != "" {
		t.Error("a refused package was opened")
	}
	if p := vm.Problems.Get(); !strings.Contains(p, "theme.css") || !strings.Contains(p, "evil.example") {
		t.Errorf("problems = %q", p)
	}
	unchanged(t, vm, before, vm.project().Text(ConfigPath))
	SetLocale("de")
	vm.ThemeImportFile("evil.gsta.xz", bad)
	if p := vm.Problems.Get(); !strings.Contains(p, "Eine CSS-Datei des Theme-Pakets") {
		t.Errorf("the code is said in German too: %q", p)
	}
	SetLocale("en")
	vm.ThemeImportFile("junk.gsta.xz", []byte("not an archive"))
	if !strings.Contains(vm.Problems.Get(), "archive") {
		t.Errorf("problems = %q", vm.Problems.Get())
	}
}

func TestImportingALooseSheet(t *testing.T) {
	vm, _, _ := themeVM(t)
	vm.ThemeImportFile("brand.css", []byte("/* @theme brand */\n@import \"gaia\";\nh1 { color: #c00 }\n"))
	if vm.ThemeName.Get() != "brand" {
		t.Fatalf("theme %q; status %q; problems %q", vm.ThemeName.Get(), vm.Status.Get(), vm.Problems.Get())
	}
	if !strings.Contains(vm.Status.Get(), "gaia") {
		t.Errorf("the base theme's move is said: %q", vm.Status.Get())
	}
	if css := vm.project().Text("themes/brand.css"); !strings.Contains(css, `@import-theme "gaia"`) || !strings.Contains(css, "#c00") {
		t.Errorf("sheet:\n%s", css)
	}
	for name, data := range map[string][]byte{"x.txt": []byte("hi"), "bin.css": {0xff, 0xfe, 0x00}} {
		vm.ThemeImportFile(name, data)
		if !strings.Contains(vm.Status.Get(), "style sheet") && !strings.Contains(vm.Status.Get(), name) {
			t.Errorf("%s: status %q", name, vm.Status.Get())
		}
	}
	vm.ThemeImportFile("gaia.css", []byte("h1{}"))
	if !strings.Contains(vm.Problems.Get(), "built-in") {
		t.Errorf("a built-in name is refused: %q", vm.Problems.Get())
	}
	vm.ThemeImportFile("brand.css", []byte("h1{}")) // taken
	if !strings.Contains(vm.Problems.Get(), "themes/brand.css") {
		t.Errorf("a taken name is refused: %q", vm.Problems.Get())
	}
	// a sheet the library would refuse is kept (the person's text) and the refusal shown
	vm.ThemeImportFile("remote.css", []byte("h1{background:url(https://evil.example/x.png)}"))
	if vm.ThemeName.Get() != "remote" || !strings.Contains(vm.Problems.Get(), "evil.example") {
		t.Errorf("%q %q", vm.ThemeName.Get(), vm.Problems.Get())
	}
}

func TestTheRoundTripImportEditExportReimport(t *testing.T) {
	vm, h, _ := themeVM(t)
	vm.ThemeImportFile("acme.gsta.xz", pack(t, acmeFiles()))
	vm.SetThemeValue("accent", "#ff0066")
	vm.SetThemeValue("font-heading", "Georgia")
	vm.SetThemeMeta("version", "1.1.0")
	vm.SetThemeMeta("authors", "Ada")
	// the package's own sheet names its asset "logo": the editor's logo is refused
	// rather than break that rule, and the package is as it was
	vm.ThemeSetLogoFile("mark.png", png)
	if !strings.Contains(vm.Status.Get(), "assets/logo.svg") {
		t.Errorf("a logo that would break the sheet is refused, saying what: %q", vm.Status.Get())
	}
	vm.ExportThemePackage()
	data, ok := h.saved["acme-1.1.0.gsta.xz"]
	if !ok {
		t.Fatalf("no download: %v %q", slices.Collect(mapsKeys(h.saved)), vm.Status.Get())
	}
	// what was downloaded verifies, as `slides theme verify` reads it
	pkg, err := themepack.Read(bytes.NewReader(data), int64(len(data)), themepack.Limits{})
	if err != nil {
		t.Fatalf("the export does not verify: %v", err)
	}
	if pkg.Meta.Version != "1.1.0" || pkg.Meta.Authors[0] != "Ada" || pkg.Meta.Extends != "default" || pkg.Meta.MinSlides != "0.17.0" || len(pkg.Meta.Kinds) != 1 {
		t.Errorf("meta lost: %+v", pkg.Meta)
	}
	if !bytes.Equal(pkg.Files["assets/bg.png"], png) || !bytes.Equal(pkg.Files["assets/logo.svg"], []byte(acmeLogo)) {
		t.Errorf("assets lost: %v", slices.Collect(mapsKeys(pkg.Files)))
	}
	if !strings.Contains(string(pkg.Files["theme.css"]), "--gsta-accent: #ff0066;") || !strings.Contains(string(pkg.Files["theme.css"]), "h1 { color: #b00020; }") {
		t.Errorf("sheet:\n%s", pkg.Files["theme.css"])
	}

	// into a fresh project: the editor shows what was edited
	vm2, _, _ := themeVM(t)
	vm2.ThemeImportFile("acme-1.1.0.gsta.xz", data)
	if vm2.ThemeName.Get() != "acme" {
		t.Fatalf("reimport: %q / %q", vm2.ThemeName.Get(), vm2.Problems.Get())
	}
	v, st, _ := vm2.ThemeValues()
	if st != BlockInSync || v["accent"] != "#ff0066" || v["font-heading"] != "Georgia" {
		t.Errorf("values after the round trip: %v %v", v, st)
	}
	m, _, _ := vm2.ThemeMetaValues()
	if m["version"] != "1.1.0" || m["authors"] != "Ada" {
		t.Errorf("meta after the round trip: %v", m)
	}
	// the compiled sheets are the same in both projects: what is shown is what is exported
	if vm.project().Text("themes/acme.css") != vm2.project().Text("themes/acme.css") {
		t.Error("the sheet differs after the round trip")
	}
	// and the exported package, exported again, is the same bytes
	vm2.ExportThemePackage()
	if again := vm2.host.(*filesTab).saved["acme-1.1.0.gsta.xz"]; !bytes.Equal(again, data) {
		t.Error("a package imported and exported again is not the same bytes")
	}
}

func TestExportCSSAndTheCannotSaveNotice(t *testing.T) {
	vm, h, _ := themeVM(t)
	vm.NewThemeInEditor("default")
	vm.SetThemeValue("background", "#010203")
	vm.ExportThemeCSS()
	css := string(h.saved["my-default.css"])
	if !strings.Contains(css, "#010203") || !strings.Contains(css, "@theme my-default") {
		t.Errorf("css:\n%s", css)
	}
	if !strings.Contains(vm.Status.Get(), "my-default.css") {
		t.Errorf("status %q", vm.Status.Get())
	}
	// a page that cannot save bytes says so
	plain := NewViewModel(&fakeTransport{}, &tab{fakeHost: newHost(), answer: true}, "")
	plain.NewThemeInEditor("default")
	plain.ExportThemePackage()
	plain.ExportThemeCSS()
	if !strings.Contains(plain.Status.Get(), "save") && plain.Status.Get() == "" {
		t.Errorf("status %q", plain.Status.Get())
	}
}

func TestExportOfAThemeTheLibraryRefusesShowsWhy(t *testing.T) {
	vm, h, _ := themeVM(t)
	vm.ThemeImportFile("remote.css", []byte("h1{background:url(https://evil.example/x.png)}"))
	vm.ExportThemePackage()
	vm.ExportThemeCSS()
	if len(h.saved) != 0 {
		t.Errorf("a refused theme was exported: %v", slices.Collect(mapsKeys(h.saved)))
	}
	if !strings.Contains(vm.Problems.Get(), "evil.example") || !strings.Contains(vm.Status.Get(), "evil.example") {
		t.Errorf("problems %q status %q", vm.Problems.Get(), vm.Status.Get())
	}
}

func TestFontsAndLogoThroughTheView(t *testing.T) {
	vm, _, _ := themeVM(t)
	vm.NewThemeInEditor("default")
	vm.ThemeAddFontFile("Mono.ttf", jetbrainsmono.TTF)
	if !slices.Contains(vm.ThemeFamilies(), "Mono") || !slices.Contains(vm.ThemeFamilies(), "serif") {
		t.Errorf("families %v", vm.ThemeFamilies())
	}
	if _, ok := vm.project().Read("theme-src/my-default/fonts/Mono.ttf"); !ok {
		t.Errorf("the font file: %q", paths(vm))
	}
	_, fonts, _ := vm.ThemeMetaValues()
	if len(fonts) != 1 || fonts[0].License != UnspecifiedLicense {
		t.Fatalf("fonts %+v", fonts)
	}
	vm.SetThemeMeta("font-license:Mono", "OFL-1.1")
	if _, fonts, _ := vm.ThemeMetaValues(); fonts[0].License != "OFL-1.1" {
		t.Errorf("fonts %+v", fonts)
	}
	vm.SetThemeValue("font-body", "Mono")
	if css := vm.project().Text("themes/my-default.css"); !strings.Contains(css, `font-family:"Mono"`) || !strings.Contains(css, "data:font/") {
		t.Errorf("the compiled sheet inlines the font for the preview")
	}
	vm.ThemeAddFontFile("x.txt", []byte("no"))
	if !strings.Contains(vm.Status.Get(), "not a font") {
		t.Errorf("status %q", vm.Status.Get())
	}
	vm.ThemeSetLogoFile("l.svg", []byte(acmeLogo))
	if vm.ThemeLogoName() != "logo.svg" {
		t.Errorf("logo %q", vm.ThemeLogoName())
	}
	vm.ThemeSetLogoFile("l.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" onload="x()"/>`))
	if vm.ThemeLogoName() != "logo.svg" {
		t.Error("a refused logo replaced the good one")
	}
	vm.RemoveThemeLogo()
	if vm.ThemeLogoName() != "" {
		t.Error("logo not removed")
	}
	vm.ThemeAddFontFile("Mono.ttf", jetbrainsmono.TTF) // again: the face is replaced, not doubled
	if _, fonts, _ := vm.ThemeMetaValues(); len(fonts) != 1 {
		t.Errorf("fonts %+v", fonts)
	}
}

func TestAProjectThemeWithoutSourceIsTakenOverAsASheet(t *testing.T) {
	vm, _, _ := themeVM(t)
	vm.NewTheme("default") // the file panel's own template: themes/my-default.css, no source
	if vm.hasThemeSource("my-default") {
		t.Fatal("the plain template has no source")
	}
	if !slices.Contains(vm.EditableThemes(), "my-default") {
		t.Fatalf("editable: %v (themes %v)", vm.EditableThemes(), vm.Themes)
	}
	vm.ThemePane.Set(true)
	if vm.ThemeName.Get() != "my-default" {
		t.Fatalf("the pane starts on the deck's project theme: %q", vm.ThemeName.Get())
	}
	original := vm.project().Text("themes/my-default.css")
	vm.SetThemeValue("accent", "#00aa00")
	css := vm.project().Text("themes/my-default.css")
	if !strings.Contains(css, "--gsta-accent: #00aa00;") {
		t.Errorf("sheet:\n%s", css)
	}
	if !strings.Contains(vm.project().Text("theme-src/my-default/theme.css"), "@theme my-default") {
		t.Errorf("the person's own sheet became the source:\n%s", vm.project().Text("theme-src/my-default/theme.css"))
	}
	_ = original
	// the CSS view of a theme with no source makes one first
	vm.ThemeName.Set("")
	vm.NewTheme("gaia")
	vm.ThemeName.Set("my-gaia")
	vm.OpenThemeSheet()
	if vm.Open.Get() != "theme-src/my-gaia/theme.css" || !vm.hasThemeSource("my-gaia") {
		t.Errorf("open %q", vm.Open.Get())
	}
}

func TestThemeVersionIsKeptInTheVersionsHistory(t *testing.T) {
	vm, _, _ := themeVM(t)
	vm.NewThemeInEditor("default")
	vm.SetThemeMeta("version", "0.3.0")
	vm.SaveThemeVersion()
	found := false
	for _, it := range vm.VersionItems.Get() {
		if strings.Contains(it, "my-default 0.3.0") {
			found = true
		}
	}
	if !found {
		t.Errorf("versions: %q; status %q", vm.VersionItems.Get(), vm.Status.Get())
	}
	// and every edit step keeps what it replaced as a version too
	vm.SetThemeValue("accent", "#123456")
	before := false
	for _, it := range vm.VersionItems.Get() {
		if strings.Contains(it, "my-default") {
			before = true
		}
	}
	if !before {
		t.Errorf("versions: %q", vm.VersionItems.Get())
	}
}

func TestThemeCommandsNeedATheme(t *testing.T) {
	vm, h, _ := themeVM(t)
	for _, c := range []interface{ CanExecute() bool }{vm.ThemePackage, vm.ThemeCSS, vm.ThemeSave, vm.ThemeFont, vm.ThemeLogo, vm.ThemeNoLogo, vm.ThemeOpenCSS, vm.ThemeRestore} {
		if c.CanExecute() {
			t.Error("a command is enabled with no theme")
		}
	}
	if !vm.ThemeNew.CanExecute() || !vm.ThemeImport.CanExecute() {
		t.Error("New and Import need no theme")
	}
	vm.SetThemeValue("accent", "#fff") // no theme: said, not a panic
	vm.SetThemeMeta("version", "1.0.0")
	vm.RestoreThemeBlock()
	vm.ExportThemePackage()
	vm.ExportThemeCSS()
	vm.SaveThemeVersion()
	vm.OpenThemeSheet()
	if !strings.Contains(vm.Status.Get(), "No theme") && vm.Status.Get() == "" {
		t.Errorf("status %q", vm.Status.Get())
	}
	if v, _, ok := vm.ThemeValues(); ok || v != nil {
		t.Error("values without a theme")
	}
	if m, f, ok := vm.ThemeMetaValues(); ok || m != nil || f != nil {
		t.Error("meta without a theme")
	}
	if vm.ThemeLogoName() != "" {
		t.Error("logo without a theme")
	}
	vm.ThemeNew.Execute()
	if vm.ThemeName.Get() == "" {
		t.Error("New made no theme")
	}
	vm.ThemeImport.Execute()
	vm.ThemeFont.Execute()
	vm.ThemeLogo.Execute()
	if want := []string{"themeedit", "themefont", "themelogo"}; !slices.Equal(h.picked, want) {
		t.Errorf("pickers asked for %v, want %v", h.picked, want)
	}
	for _, c := range []interface{ Execute() }{vm.ThemePackage, vm.ThemeCSS, vm.ThemeSave, vm.ThemeNoLogo, vm.ThemeOpenCSS, vm.ThemeRestore} {
		c.Execute()
	}
}

func TestAThemeSourceThatIsNotAPackageIsShownNotDropped(t *testing.T) {
	vm, _, _ := themeVM(t)
	vm.NewThemeInEditor("default")
	vm.OpenThemeSheet()
	vm.OpenFile("theme-src/my-default/theme.hcl")
	vm.FileText.Set(`theme { name = "my-default" version = `) // half typed
	if !strings.Contains(vm.Problems.Get(), "theme-src/my-default") {
		t.Errorf("problems = %q", vm.Problems.Get())
	}
	if _, _, ok := vm.ThemeValues(); ok {
		t.Error("an unreadable theme.hcl gives no values")
	}
	vm.FileText.Set(strings.Replace(vm.project().Text("theme-src/my-default/theme.hcl"), `version = `, `version = "1.0.0"`+"\n}\n#", 1))
	if _, _, ok := vm.ThemeValues(); !ok && vm.Problems.Get() == "" {
		t.Error("nothing is said")
	}
}

func TestRenameTheAndEditAFacesWeightAndStyle(t *testing.T) {
	vm, _, _ := themeVM(t)
	vm.NewThemeInEditor("default")
	vm.ThemeAddFontFile("Mono.ttf", jetbrainsmono.TTF)
	vm.SetThemeMeta("font-weight:fonts/Mono.ttf", "700")
	vm.SetThemeMeta("font-style:fonts/Mono.ttf", "italic")
	if _, fonts, _ := vm.ThemeMetaValues(); fonts[0].Weight != "700" || fonts[0].Style != "italic" {
		t.Fatalf("%+v", fonts)
	}
	vm.SetThemeMeta("font-weight:fonts/Mono.ttf", "bold") // the library refuses: not stored
	if _, fonts, _ := vm.ThemeMetaValues(); fonts[0].Weight != "700" {
		t.Errorf("%+v", fonts)
	}
	vm.SetThemeValue("accent", "#123456")
	vm.RenameTheme("brand")
	if vm.ThemeName.Get() != "brand" {
		t.Fatalf("name %q status %q", vm.ThemeName.Get(), vm.Status.Get())
	}
	for _, p := range paths(vm) {
		if strings.Contains(p, "my-default") {
			t.Errorf("the old name is left in %s", p)
		}
	}
	if got, _ := DeckTheme(vm.Source.Get()); got != "brand" {
		t.Errorf("deck theme %q", got)
	}
	css := vm.project().Text("themes/brand.css")
	if !strings.Contains(css, "@theme brand") || !strings.Contains(css, "#123456") || strings.Contains(css, "my-default") {
		t.Errorf("sheet:\n%s", css)
	}
	if _, fonts, _ := vm.ThemeMetaValues(); fonts[0].Style != "italic" {
		t.Errorf("%+v", fonts)
	}
	vm.undoStep()
	if vm.hasThemeSource("brand") || !vm.hasThemeSource("my-default") {
		t.Error("undo puts the old name back")
	}
	vm.redoStep()
	vm.ThemeName.Set("brand")
	for _, bad := range []string{"gaia", "Bad Name", "brand", ""} {
		vm.RenameTheme(bad)
		if vm.ThemeName.Get() != "brand" {
			t.Errorf("%q renamed", bad)
		}
	}
	vm.NewThemeInEditor("default")
	vm.RenameTheme("brand") // taken
	if vm.ThemeName.Get() == "brand" || !strings.Contains(vm.Status.Get(), "brand") {
		t.Errorf("taken: %q %q", vm.ThemeName.Get(), vm.Status.Get())
	}
	vm.ThemeName.Set("")
	vm.RenameTheme("x")
}
