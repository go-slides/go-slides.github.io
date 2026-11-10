package playground

import (
	"archive/tar"
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-slides/slides/deck"
	"github.com/go-slides/slides/themepack"
	"github.com/ulikunitz/xz"
)

const acmeHCL = `theme {
  name        = "acme"
  version     = "1.0.0"
  description = "Example theme"
  license     = "MIT"
  extends     = "default"
  min_slides  = "0.17.0"
}
css "theme.css" {}
asset "logo" {
  file = "assets/logo.svg"
}
asset "background" {
  file = "assets/bg.png"
}
kind "acme-title" {
  class    = "lead"
  paginate = "false"
}
`

const acmeCSS = `section { background: url(assets/bg.png); }
section::after { content: ""; background: url(assets/logo.svg) no-repeat; }
h1 { color: #b00020; }
`

const acmeLogo = `<svg xmlns="http://www.w3.org/2000/svg" width="8" height="8"><rect width="8" height="8" fill="#b00020"/></svg>`

// acmeFiles are the files of a small theme package like the example of
// go-slides/slides (testdata/themes/acme), without the font.
func acmeFiles() themepack.Files {
	return themepack.Files{
		"theme.hcl":       []byte(acmeHCL),
		"theme.css":       []byte(acmeCSS),
		"assets/logo.svg": []byte(acmeLogo),
		"assets/bg.png":   png,
	}
}

// pack is a package archive made in code.
func pack(t *testing.T, files themepack.Files) []byte {
	t.Helper()
	b, err := themepack.WriteFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// entry is one member of a hostile archive.
type entry struct {
	hdr  tar.Header
	body string
}

// hostile is a .gsta.xz made by hand (archive/tar, ulikunitz/xz), with the
// entries themepack.WriteFiles would never write.
func hostile(t *testing.T, entries ...entry) []byte {
	t.Helper()
	var out bytes.Buffer
	xw, err := xz.NewWriter(&out)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(xw)
	for _, e := range entries {
		h := e.hdr
		h.Size = int64(len(e.body))
		if err := tw.WriteHeader(&h); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(e.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := xw.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

// with is files with one entry replaced.
func with(files themepack.Files, name, text string) themepack.Files {
	out := themepack.Files{}
	for k, v := range files {
		out[k] = v
	}
	out[name] = []byte(text)
	return out
}

// unchanged fails when the project is not what it was: a refused package
// leaves nothing behind.
func unchanged(t *testing.T, vm *ViewModel, before []string, config string) {
	t.Helper()
	if got := paths(vm); !slices.Equal(got, before) {
		t.Fatalf("files %q, were %q", got, before)
	}
	if got := vm.project().Text(ConfigPath); got != config {
		t.Fatalf("slides.yaml changed: %q", got)
	}
}

func TestImportThemePackage(t *testing.T) {
	vm, _, _ := openFiles(t, store.NewMemory())
	p, err := vm.ImportTheme(pack(t, acmeFiles()))
	if err != nil || p != "themes/acme.css" {
		t.Fatalf("import = %q, %v", p, err)
	}
	// the package's own files are kept as the source folder the theme editor edits
	if got := paths(vm); !slices.Equal(got, []string{"deck.md", "kinds/acme.yaml", "slides.yaml", "theme-src/acme/assets/bg.png", "theme-src/acme/assets/logo.svg", "theme-src/acme/theme.css", "theme-src/acme/theme.hcl", "themes/acme.css", "kinds/", "media/", "themes/"}) {
		t.Fatalf("files = %q", got)
	}
	css := vm.project().Text("themes/acme.css")
	for _, want := range []string{`@import-theme "default"`, "@theme acme", "url(\"data:image/svg+xml", "url(\"data:image/png"} {
		if !strings.Contains(css, want) {
			t.Errorf("the sheet lacks %q:\n%s", want, css)
		}
	}
	if strings.Contains(css, "assets/") {
		t.Errorf("the sheet still names a file of the package:\n%s", css)
	}
	cfg := vm.project().Text(ConfigPath)
	if !strings.Contains(cfg, "themes: [themes]") && !strings.Contains(cfg, "themes: themes") || !strings.Contains(cfg, "kinds/acme.yaml") {
		t.Errorf("slides.yaml = %q", cfg)
	}
	kb, _ := vm.project().Read("kinds/acme.yaml")
	ks, err := deck.ParseKindsYAML(kb, "acme.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if k := ks["acme-title"]; k == nil || k.Dirs.Get("class") != "lead" || k.Dirs.Get("paginate") != "false" {
		t.Errorf("kinds = %+v", ks)
	}
	if !slices.Contains(vm.Themes, "acme") {
		t.Errorf("themes %q", vm.Themes)
	}
	if !strings.Contains(vm.Status.Get(), "themes/acme.css") || !strings.Contains(vm.Status.Get(), "kinds/acme.yaml") {
		t.Errorf("status %q", vm.Status.Get())
	}
	// the renderer is sent the sheet and the kinds
	files, _ := vm.textFiles()
	if files[projectRoot+"/themes/acme.css"] != css {
		t.Error("the renderer is not sent the theme")
	}
	// the same package again is the same files
	if _, err := vm.ImportTheme(pack(t, acmeFiles())); err != nil {
		t.Fatalf("again: %v", err)
	}
	if strings.Count(vm.project().Text(ConfigPath), "kinds/acme.yaml") != 1 {
		t.Errorf("listed twice: %q", vm.project().Text(ConfigPath))
	}
}

func TestImportThemeThroughAddFileAndTheMenu(t *testing.T) {
	vm, h, _ := openFiles(t, store.NewMemory())
	z := pack(t, acmeFiles())
	p, err := vm.AddFile("", "Acme.GSTA.XZ", z)
	if err != nil || p != "themes/acme.css" {
		t.Fatalf("AddFile = %q, %v", p, err)
	}
	for _, f := range paths(vm) {
		if strings.HasSuffix(strings.ToLower(f), ThemeExt) {
			t.Fatalf("the archive was kept as %s", f)
		}
	}
	if _, err := vm.AddFile("themes", "x.gsta.xz", []byte("not a package")); err == nil {
		t.Fatal("a file that is not a package was taken")
	}
	_ = h // the "Add media" picker takes .xz and .gsta.xz too (the host's accept list)
	if !IsThemePackage("a.gsta.xz") || !IsThemePackage(`C:\t\A.GSTA.XZ`) || IsThemePackage("a.xz") || IsThemePackage("a.zip") || IsThemePackage("gsta.xz.txt") {
		t.Fatal("IsThemePackage")
	}
	// a page says what went wrong in its status line
	vm.OpenThemePackage([]byte("nope"))
	if !strings.Contains(vm.Status.Get(), "theme.archive") {
		t.Errorf("status %q", vm.Status.Get())
	}
}

func TestImportThemeIsOneUndoStep(t *testing.T) {
	vm, _, _ := openFiles(t, store.NewMemory())
	if _, err := vm.ImportTheme(pack(t, acmeFiles())); err != nil {
		t.Fatal(err)
	}
	if !vm.undo.project.CanUndo() {
		t.Fatal("no step to undo")
	}
	vm.undo.project.Undo()
	if got := paths(vm); slices.Contains(got, "themes/acme.css") || slices.Contains(got, "kinds/acme.yaml") {
		t.Fatalf("files %q", got)
	}
}

// TestImportThemeRefusesWhatTheLibraryRefuses: a package is untrusted (a
// dropped file as much as any), and a refusal writes nothing.
func TestImportThemeRefusesWhatTheLibraryRefuses(t *testing.T) {
	hcl := tar.Header{Name: "theme.hcl", Typeflag: tar.TypeReg, Mode: 0o644}
	slip := hostile(t, entry{hcl, acmeHCL}, entry{tar.Header{Name: "../../evil.css", Typeflag: tar.TypeReg, Mode: 0o644}, "x"})
	symlink := hostile(t, entry{hcl, acmeHCL}, entry{tar.Header{Name: "theme.css", Typeflag: tar.TypeSymlink, Linkname: "/etc/passwd"}, ""})
	hardlink := hostile(t, entry{hcl, acmeHCL}, entry{tar.Header{Name: "theme.css", Typeflag: tar.TypeLink, Linkname: "theme.hcl"}, ""})
	// a tar of zeros, far past the compression ratio the reader allows
	bomb := hostile(t, entry{hcl, acmeHCL}, entry{tar.Header{Name: "theme.css", Typeflag: tar.TypeReg, Mode: 0o644}, strings.Repeat("\x00", 64<<20)})
	// a CSS over its bound
	bigCSS := hostile(t, entry{hcl, acmeHCL}, entry{tar.Header{Name: "theme.css", Typeflag: tar.TypeReg, Mode: 0o644}, strings.Repeat("a", 3<<20)})

	for name, z := range map[string]struct {
		data []byte
		code string
	}{
		"tar-slip name":        {slip, "theme.archive"},
		"a symlink entry":      {symlink, "theme.archive"},
		"a hardlink entry":     {hardlink, "theme.archive"},
		"a tar of zeros":       {bomb, "theme."},
		"css @import":          {pack(t, with(acmeFiles(), "theme.css", "@import url(https://evil.example/x.css);\nh1{color:red}")), "theme.css"},
		"css url() outside":    {pack(t, with(acmeFiles(), "theme.css", "h1{background:url(https://evil.example/x.png)}")), "theme.css"},
		"script in an SVG":     {pack(t, with(acmeFiles(), "assets/logo.svg", `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)), "theme.svg"},
		"onload in an SVG":     {pack(t, with(acmeFiles(), "assets/logo.svg", `<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"></svg>`)), "theme.svg"},
		"a built-in name":      {pack(t, with(acmeFiles(), "theme.hcl", strings.Replace(acmeHCL, `"acme"`, `"gaia"`, 1))), "theme.meta"},
		"a wrong sum":          {pack(t, with(acmeFiles(), "theme.hcl", strings.Replace(acmeHCL, `css "theme.css" {}`, `css "theme.css" { sha256 = "`+strings.Repeat("0", 64)+`" }`, 1))), "theme.checksum"},
		"too new":              {pack(t, with(acmeFiles(), "theme.hcl", strings.Replace(acmeHCL, `"0.17.0"`, `"99.0.0"`, 1))), "theme.version"},
		"an undeclared file":   {pack(t, with(acmeFiles(), "extra.js", "alert(1)")), "theme.meta"},
		"a CSS over its bound": {bigCSS, "theme."},
		"not an archive":       {[]byte("PK nothing"), "theme.archive"},
		"a zip":                {[]byte("PK\x03\x04 a ZIP is no package"), "theme.archive"},
	} {
		t.Run(name, func(t *testing.T) {
			vm, _, _ := openFiles(t, store.NewMemory())
			before, config := paths(vm), vm.project().Text(ConfigPath)
			_, err := vm.ImportTheme(z.data)
			if err == nil {
				t.Fatal("imported")
			}
			if !strings.Contains(ErrText(err), z.code) {
				t.Errorf("refusal %q does not carry the code %q", ErrText(err), z.code)
			}
			unchanged(t, vm, before, config)
		})
	}
}

func TestImportThemeRefusesWhatTheProjectCannotTake(t *testing.T) {
	// a built-in name, whatever the library says
	vm, _, _ := openFiles(t, store.NewMemory())
	if err := vm.importThemeNamed(t, "uncover"); err == nil || !strings.Contains(ErrText(err), "built-in") {
		t.Errorf("built-in: %v", err)
	}

	// a theme of that name, in another file or with other bytes
	vm, _, _ = openFiles(t, store.NewMemory())
	if err := vm.CreateFile("themes/mine.css", "/* @theme acme */\nh1{}\n"); err != nil {
		t.Fatal(err)
	}
	before, config := paths(vm), vm.project().Text(ConfigPath)
	if _, err := vm.ImportTheme(pack(t, acmeFiles())); err == nil || !strings.Contains(ErrText(err), "themes/mine.css") {
		t.Errorf("taken name: %v", err)
	}
	unchanged(t, vm, before, config)

	vm, _, _ = openFiles(t, store.NewMemory())
	if _, err := vm.ImportTheme(pack(t, acmeFiles())); err != nil {
		t.Fatal(err)
	}
	before, config = paths(vm), vm.project().Text(ConfigPath)
	newer := pack(t, with(acmeFiles(), "theme.hcl", strings.Replace(acmeHCL, `"1.0.0"`, `"1.1.0"`, 1)))
	if _, err := vm.ImportTheme(newer); err == nil || !strings.Contains(ErrText(err), "themes/acme.css") {
		t.Errorf("other bytes under the same name: %v", err)
	}
	unchanged(t, vm, before, config)

	// a kinds file in the way
	vm, _, _ = openFiles(t, store.NewMemory())
	if err := vm.CreateFile("kinds/acme.yaml", "x:\n  class: lead\n"); err != nil {
		t.Fatal(err)
	}
	before, config = paths(vm), vm.project().Text(ConfigPath)
	if _, err := vm.ImportTheme(pack(t, acmeFiles())); err == nil || !strings.Contains(ErrText(err), "kinds/acme.yaml") {
		t.Errorf("kinds in the way: %v", err)
	}
	unchanged(t, vm, before, config)

	// a base theme the project lacks, then has
	vm, _, _ = openFiles(t, store.NewMemory())
	child := pack(t, with(with(acmeFiles(), "theme.hcl", strings.Replace(acmeHCL, `"default"`, `"corporate"`, 1)), "theme.css", "h1{color:red}"))
	before, config = paths(vm), vm.project().Text(ConfigPath)
	if _, err := vm.ImportTheme(child); err == nil || !strings.Contains(ErrText(err), "corporate") {
		t.Errorf("unknown base: %v", err)
	}
	unchanged(t, vm, before, config)
	if err := vm.CreateFile("themes/corporate.css", "/* @theme corporate */\n@import 'default';\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := vm.ImportTheme(child); err != nil {
		t.Errorf("a base the project has: %v", err)
	}
}

// importThemeNamed imports a package named name (the library refuses the
// built-in names itself; this reaches the playground's own refusal).
func (vm *ViewModel) importThemeNamed(t *testing.T, name string) error {
	t.Helper()
	pkg, err := themepack.FromFiles(acmeFiles(), themepack.Limits{})
	if err != nil {
		t.Fatal(err)
	}
	pkg.Meta.Name = name
	_, err = vm.importTheme(pkg)
	return err
}

// TestImportThemeStaysWithinWhatTheRendererIsSent: text past MaxTextFiles is
// dropped by the renderer without a word, so the import refuses it.
func TestImportThemeStaysWithinWhatTheRendererIsSent(t *testing.T) {
	vm, _, _ := openFiles(t, store.NewMemory())
	if err := vm.CreateFile("themes/big.css", "/* @theme big */\n/*"+strings.Repeat("a", MaxTextFiles-300)+"*/\n"); err != nil {
		t.Fatal(err)
	}
	before, config := paths(vm), vm.project().Text(ConfigPath)
	_, err := vm.ImportTheme(pack(t, acmeFiles()))
	if err == nil || !strings.Contains(ErrText(err), "at most 2.0 MiB") {
		t.Fatalf("over the bound: %v", err)
	}
	unchanged(t, vm, before, config)
	files, _ := vm.textFiles()
	if len(files) != 2 { // slides.yaml and big.css: all of them still sent
		t.Errorf("files sent: %d", len(files))
	}
}
