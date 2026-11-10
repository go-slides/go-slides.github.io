package playground

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-opentype/fonts/jetbrainsmono"
	"github.com/go-slides/slides/diag"
	"github.com/go-slides/slides/themepack"
)

func TestSlugNameAndSrcPaths(t *testing.T) {
	for in, want := range map[string]string{"My Theme!": "my-theme", "  --A_b--": "a-b", "Ünï": "n", "": "theme", "???": "theme", "ABC123": "abc123"} {
		if got := SlugName(in); got != want {
			t.Errorf("SlugName(%q) = %q, want %q", in, got, want)
		}
	}
	if got := SlugName(strings.Repeat("a", 90)); len(got) != 64 {
		t.Errorf("a long name is cut to 64, got %d", len(got))
	}
	if ThemeSrcDir("acme") != "theme-src/acme" {
		t.Error("ThemeSrcDir")
	}
	for in, want := range map[string]string{"theme-src/acme/theme.hcl": "acme", "theme-src/acme/fonts/a.ttf": "acme", "theme-src/acme": "", "themes/acme.css": "", "theme-src/Bad Name/x": ""} {
		if got := ThemeNameOfSrc(in); got != want {
			t.Errorf("ThemeNameOfSrc(%q) = %q, want %q", in, got, want)
		}
	}
}

// the package of the import tests, as a document
func acmeDoc(t *testing.T) *ThemeDoc {
	t.Helper()
	pkg, err := themepack.FromFiles(acmeFiles(), themepack.Limits{})
	if err != nil {
		t.Fatal(err)
	}
	return DocFromPackage(pkg)
}

func TestDocPackRoundTripsAndPassesTheLibraryCheck(t *testing.T) {
	d := acmeDoc(t)
	if d.Meta.Name != "acme" || d.Meta.Extends != "default" || len(d.Meta.Kinds) != 1 || len(d.Meta.Assets) != 2 {
		t.Fatalf("meta = %+v", d.Meta)
	}
	data, err := d.Pack()
	if err != nil {
		t.Fatal(err)
	}
	// what is handed out is what `slides theme verify` reads
	back, err := themepack.Read(bytes.NewReader(data), int64(len(data)), themepack.Limits{})
	if err != nil {
		t.Fatalf("the package does not verify: %v", err)
	}
	if back.Meta.Name != "acme" || back.Meta.Version != "1.0.0" || back.Meta.MinSlides != "0.17.0" || back.Meta.Kinds[0].Name != "acme-title" {
		t.Errorf("round trip lost metadata: %+v", back.Meta)
	}
	for n, b := range acmeFiles() {
		if n != "theme.hcl" && !bytes.Equal(back.Files[n], b) {
			t.Errorf("%s changed in the round trip", n)
		}
	}
	// the same document packs to the same bytes
	again, _ := d.Pack()
	if !bytes.Equal(data, again) {
		t.Error("packing is not deterministic")
	}
	// the plain CSS is the compiled sheet
	css, err := d.PlainCSS()
	if err != nil || !strings.Contains(css, `@import-theme "default"`) || !strings.Contains(css, "@theme acme") {
		t.Errorf("plain css: %v\n%s", err, css)
	}
}

func TestDocHCLQuotesWhatNeedsIt(t *testing.T) {
	d := NewThemeDoc("q", "")
	d.Meta.Description = `a "quoted" ${x} line\n`
	d.Meta.Authors = []string{`Ada "A" <a@x.org>`, "B"}
	d.Meta.CSS[0].SHA256 = strings.Repeat("0", 64)
	d.Meta.Fonts = []themepack.Font{{Family: "F", Weight: "100 900", Style: "italic", File: "fonts/f.ttf", License: "OFL-1.1", SHA256: strings.Repeat("1", 64)}}
	d.Meta.Assets = []themepack.Asset{{ID: "x", File: "assets/x.png", SHA256: strings.Repeat("2", 64)}}
	m, err := themepack.ParseMeta([]byte(d.HCL()), themepack.Limits{})
	if err != nil {
		t.Fatalf("%v\n%s", err, d.HCL())
	}
	if m.Description != d.Meta.Description || m.Authors[0] != d.Meta.Authors[0] || m.Fonts[0].Weight != "100 900" || m.CSS[0].SHA256 == "" || m.Assets[0].SHA256 == "" {
		t.Errorf("round trip: %+v", m)
	}
}

func TestEncodeBringsDeclaredSumsUpToDate(t *testing.T) {
	d := NewThemeDoc("sums", "")
	d.Meta.CSS[0].SHA256 = sum(d.Files["theme.css"])
	d.SetSheet("h1 { color: red }\n")
	if _, err := d.Check(); err != nil {
		t.Fatalf("an edited sheet with a declared sum must still verify: %v", err)
	}
	d2 := acmeDoc(t)
	d2.Meta.Assets[0].SHA256 = sum(d2.Files[d2.Meta.Assets[0].File])
	d2.Files[d2.Meta.Assets[0].File] = []byte(strings.Replace(acmeLogo, "8", "9", 1))
	d2.Meta.Fonts = nil
	if _, err := d2.Check(); err != nil {
		t.Fatalf("an asset replaced with a declared sum: %v", err)
	}
}

func TestDocFromFilesRefusesWhatIsNotAPackage(t *testing.T) {
	if _, err := DocFromFiles(themepack.Files{"theme.css": nil}); err == nil {
		t.Error("no theme.hcl")
	}
	_, err := DocFromFiles(themepack.Files{"theme.hcl": []byte(`theme { name = "default" }`)})
	if diag.CodeOf(err) != "theme.meta" {
		t.Errorf("a coded refusal, got %v (%q)", err, diag.CodeOf(err))
	}
	d, err := DocFromFiles(acmeFiles())
	if err != nil || d.Meta.Name != "acme" || d.Files["theme.hcl"] != nil || len(d.Files) != 3 {
		t.Errorf("%v %+v", err, d)
	}
}

func TestDocFromCSSMovesTheBaseIntoExtends(t *testing.T) {
	d, notes := DocFromCSS("/* @theme Corporate */\n@import \"gaia\";\nh1 { color: red }\n", "file")
	if d.Meta.Name != "corporate" || d.Meta.Extends != "gaia" || len(notes) != 1 || !strings.Contains(notes[0], "gaia") {
		t.Fatalf("%+v %v", d.Meta, notes)
	}
	if strings.Contains(d.Sheet(), "@import") || !strings.Contains(d.Sheet(), "h1 { color: red }") {
		t.Errorf("the sheet: %q", d.Sheet())
	}
	if _, err := d.Check(); err != nil {
		t.Errorf("a loose sheet that is fine must pass the package check: %v", err)
	}
	d2, notes := DocFromCSS("h1 { color: red }", "My File")
	if d2.Meta.Name != "my-file" || len(notes) != 0 || d2.Meta.Extends != "" {
		t.Errorf("%+v %v", d2.Meta, notes)
	}
	// what the library refuses is refused on export, coded
	bad, _ := DocFromCSS("h1 { background: url(https://evil.example/x.png) }", "bad")
	if _, err := bad.Pack(); diag.CodeOf(err) == "" {
		t.Errorf("a remote url() must be refused with a code, got %v", err)
	}
}

func TestVisualBlockAbsentInSyncAndEditedByHand(t *testing.T) {
	d := NewThemeDoc("v", "default")
	if v, st := d.Visual(); len(v) != 0 || st != BlockAbsent {
		t.Fatalf("a new sheet has no block: %v %v", v, st)
	}
	before := d.Sheet()
	if err := d.SetVisual("background", "#101820"); err != nil {
		t.Fatal(err)
	}
	if err := d.SetVisual("font-body", "Georgia"); err != nil {
		t.Fatal(err)
	}
	if err := d.SetVisual("size-body", "28px"); err != nil {
		t.Fatal(err)
	}
	v, st := d.Visual()
	if st != BlockInSync || v["background"] != "#101820" || v["font-body"] != "Georgia" || v["size-body"] != "28px" {
		t.Fatalf("%v %v", v, st)
	}
	if !strings.HasPrefix(d.Sheet(), before) {
		t.Errorf("what was in the sheet stays where it was:\n%s", d.Sheet())
	}
	if !strings.Contains(d.Sheet(), `font-family: var(--gsta-font-body), sans-serif;`) || !strings.Contains(d.Sheet(), `--gsta-font-body: "Georgia";`) {
		t.Errorf("the block:\n%s", d.Sheet())
	}
	if _, err := d.Check(); err != nil {
		t.Errorf("the block is CSS the library accepts: %v", err)
	}

	// text the person wrote outside the fence is theirs, byte for byte, through every control edit
	d.SetSheet("/* mine */\nh2 { color: teal }\n" + d.Sheet() + "\n/* after */ p { margin: 0 }\n")
	if err := d.SetVisual("background", "#000000"); err != nil {
		t.Fatal(err)
	}
	if s := d.Sheet(); !strings.HasPrefix(s, "/* mine */\nh2 { color: teal }\n") || !strings.HasSuffix(s, "/* after */ p { margin: 0 }\n") {
		t.Errorf("the person's CSS moved:\n%s", s)
	}
	// a value edited by hand inside the block, still in the editor's shape, is read back
	d.SetSheet(strings.Replace(d.Sheet(), "--gsta-background: #000000;", "--gsta-background: #123456;", 1))
	if v, st := d.Visual(); v["background"] != "#123456" || st != BlockInSync {
		t.Errorf("a hand-edited value: %v %v", v, st)
	}

	// a rule added inside the fence is something the controls cannot show: they stop writing
	diverged := strings.Replace(d.Sheet(), visualEnd, "h1 { text-decoration: underline }\n"+visualEnd, 1)
	d.SetSheet(diverged)
	if _, st := d.Visual(); st != BlockDiverged {
		t.Fatalf("state = %v", st)
	}
	if err := d.SetVisual("text", "#ffffff"); err != ErrDiverged {
		t.Fatalf("a diverged block is refused, got %v", err)
	}
	if d.Sheet() != diverged {
		t.Error("a refused write must leave the sheet untouched")
	}
	// the person decides: restoring drops what was written inside the fence, nothing else
	d.RestoreVisual()
	s := d.Sheet()
	if strings.Contains(s, "underline") || !strings.Contains(s, "--gsta-background: #123456;") || !strings.HasPrefix(s, "/* mine */") || !strings.HasSuffix(s, "p { margin: 0 }\n") {
		t.Errorf("restore:\n%s", s)
	}
	if _, st := d.Visual(); st != BlockInSync {
		t.Errorf("after restore: %v", st)
	}
	// emptying every value removes the block and nothing more
	for _, p := range ThemeProps {
		if err := d.SetVisual(p.ID, ""); err != nil {
			t.Fatal(err)
		}
	}
	if strings.Contains(d.Sheet(), "gsta:visual") || !strings.Contains(d.Sheet(), "h2 { color: teal }") || !strings.Contains(d.Sheet(), "/* after */") {
		t.Errorf("after removing every value:\n%s", d.Sheet())
	}
}

func TestVisualRefusesAValueThatIsNotOne(t *testing.T) {
	d := NewThemeDoc("v", "")
	for _, c := range [][2]string{
		{"background", "red; } body { display: none"}, {"background", "url(https://x/y)"}, {"text", "red}"},
		{"font-body", `x"; } * { `}, {"font-body", "a{b}"}, {"size-body", "12 px"}, {"size-body", "1e9px"},
		{"logo-placement", "middle"}, {"nosuch", "x"},
	} {
		if err := d.SetVisual(c[0], c[1]); err == nil {
			t.Errorf("%s = %q was taken", c[0], c[1])
		}
	}
	if d.Sheet() != string(NewThemeDoc("v", "").Files["theme.css"]) {
		t.Error("a refused value changed the sheet")
	}
}

func TestVisualFencesCutInHalfAreDivergedNotOverwritten(t *testing.T) {
	d := NewThemeDoc("v", "")
	for _, s := range []string{visualBegin + "\nsection { --gsta-text: red; }\n", "section{}\n" + visualEnd + "\n", visualBegin + "\n" + visualEnd + "\n" + visualBegin + "\n" + visualEnd + "\n"} {
		d.SetSheet(s)
		if _, st := d.Visual(); st != BlockDiverged {
			t.Errorf("%q: state %v", s, st)
		}
		if err := d.SetVisual("text", "#fff"); err != ErrDiverged || d.Sheet() != s {
			t.Errorf("%q: %v", s, err)
		}
	}
}

func TestEveryPropertyIsWrittenReadAndChecked(t *testing.T) {
	d := NewThemeDoc("all", "")
	if err := d.SetLogo("logo.svg", []byte(acmeLogo)); err != nil {
		t.Fatal(err)
	}
	vals := map[string]string{
		"background": "#fff", "text": "rgb(10 20 30)", "heading": "hsl(10 50% 40%)", "accent": "#ff00aa", "link": "navy",
		"font-body": "Arial", "font-heading": "Times New Roman", "font-code": "monospace",
		"size-body": "28px", "size-heading": "56pt", "size-code": "20px", "weight-body": "400", "weight-heading": "700",
		"logo-placement": "bottom-left", "logo-size": "72px", "pagination-placement": "top-center", "pagination-size": "18px",
		"footer-align": "center", "footer-size": "14px",
	}
	if len(vals) != len(ThemeProps) {
		t.Fatalf("a property is untested: %d of %d", len(vals), len(ThemeProps))
	}
	for id, v := range vals {
		if err := d.SetVisual(id, v); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
	}
	got, st := d.Visual()
	if st != BlockInSync {
		t.Fatalf("state %v\n%s", st, d.Sheet())
	}
	for id, v := range vals {
		if got[id] != v {
			t.Errorf("%s = %q, wrote %q", id, got[id], v)
		}
	}
	for _, want := range []string{"section::before", "section::after", "section > footer", `url("../assets/logo.svg")`} {
		if want == `url("../assets/logo.svg")` {
			want = `url("assets/logo.svg")`
		}
		if !strings.Contains(d.Sheet(), want) {
			t.Errorf("the block lacks %q:\n%s", want, d.Sheet())
		}
	}
	if _, err := d.Check(); err != nil {
		t.Fatalf("every rule the block writes is accepted by the library: %v\n%s", err, d.Sheet())
	}
	// placements: the logo as a background, the other page-number corners
	for _, pl := range []string{"top-left", "top-right", "bottom-right", "background", "none"} {
		if err := d.SetVisual("logo-placement", pl); err != nil {
			t.Fatal(err)
		}
	}
	for _, pl := range PaginationPlacements {
		if err := d.SetVisual("pagination-placement", pl); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := d.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestLogoAndFontsAreDeclaredFiles(t *testing.T) {
	d := NewThemeDoc("assets", "")
	if err := d.SetLogo("x.exe", []byte("MZ")); err == nil {
		t.Error("a logo is an image")
	}
	if err := d.SetLogo("logo.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>x</script></svg>`)); diag.CodeOf(err) != "theme.svg" {
		t.Errorf("a script in a logo is refused, coded: %v (%q)", err, diag.CodeOf(err))
	}
	if len(d.Meta.Assets) != 0 {
		t.Error("a refused logo left a declaration behind")
	}
	if err := d.SetLogo("brand.svg", []byte(acmeLogo)); err != nil {
		t.Fatal(err)
	}
	if v, _ := d.Visual(); v["logo-placement"] != "top-right" || d.LogoPath() != "assets/logo.svg" {
		t.Errorf("a first logo is placed: %v %q", v, d.LogoPath())
	}
	if err := d.SetLogo("brand.png", png); err != nil { // replaced, the rule follows
		t.Fatal(err)
	}
	if _, ok := d.Files["assets/logo.svg"]; ok || d.LogoPath() != "assets/logo.png" || !strings.Contains(d.Sheet(), `url("assets/logo.png")`) {
		t.Errorf("replacing the logo: %v %q\n%s", d.Files, d.LogoPath(), d.Sheet())
	}
	if _, st := d.Visual(); st != BlockInSync {
		t.Errorf("state %v", st)
	}
	d.RemoveLogo()
	if d.LogoPath() != "" || strings.Contains(d.Sheet(), "logo.png") || len(d.Meta.Assets) != 0 {
		t.Errorf("remove: %q\n%s", d.LogoPath(), d.Sheet())
	}
	if _, err := d.Check(); err != nil {
		t.Fatal(err)
	}

	if _, err := d.AddFont("notes.txt", []byte("x"), "", "", ""); err == nil {
		t.Error("not a font")
	}
	if _, err := d.AddFont("Fake.ttf", []byte("not a font"), "", "", ""); diag.CodeOf(err) != "theme.font" {
		t.Errorf("a file that is not a font is refused, coded: %v", err)
	}
	if len(d.Meta.Fonts) != 0 || len(d.Files) != 1 {
		t.Errorf("a refused font left something behind: %v", d.Files)
	}
	fam, err := d.AddFont("JetBrainsMono-Regular.ttf", jetbrainsmono.TTF, "", "", "")
	if err != nil || fam != "JetBrainsMono" || d.Meta.Fonts[0].License != UnspecifiedLicense {
		t.Fatalf("%q %v %+v", fam, err, d.Meta.Fonts)
	}
	if _, err := d.AddFont("JetBrainsMono-Regular.ttf", jetbrainsmono.TTF, "Mono", "700", "italic"); err != nil || len(d.Meta.Fonts) != 1 || d.Meta.Fonts[0].Family != "Mono" {
		t.Fatalf("the same file again replaces its face: %v %+v", err, d.Meta.Fonts)
	}
	d.Meta.Fonts[0].License = "OFL-1.1"
	if got := d.FontFamilies(); len(got) != 1 || got[0] != "Mono" {
		t.Errorf("families %v", got)
	}
	data, err := d.Pack()
	if err != nil {
		t.Fatal(err)
	}
	css, err := d.PlainCSS()
	if err != nil || !strings.Contains(css, "@font-face{font-family:\"Mono\"") || !strings.Contains(css, "data:font/") {
		t.Errorf("the compiled sheet inlines the font: %v", err)
	}
	if len(data) == 0 {
		t.Error("empty package")
	}
}

func TestFamilyOfFile(t *testing.T) {
	for in, want := range map[string]string{
		"Inter-Bold.woff2": "Inter", "FiraCode_Regular.ttf": "FiraCode", "JetBrainsMono-Regular.ttf": "JetBrainsMono",
		"a.otf": "a", ".ttf": "Custom", `We"ird{}.ttf`: "Weird",
	} {
		if got := FamilyOfFile(in); got != want {
			t.Errorf("FamilyOfFile(%q) = %q, want %q", in, got, want)
		}
	}
	if got := FamilyOfFile(strings.Repeat("a", 100) + ".ttf"); len(got) != 64 {
		t.Errorf("a long family is cut: %d", len(got))
	}
}
