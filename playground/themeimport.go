package playground

import (
	"bytes"
	"path"
	"runtime/debug"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/go-slides/slides/deck"
	"github.com/go-slides/slides/diag"
	"github.com/go-slides/slides/themepack"
	"go.yaml.in/yaml/v3"
)

// Theme packages: a .slidestheme (go-slides/slides, package themepack) is one
// ZIP holding a theme.hcl, the CSS, fonts, images and kinds of a theme. The
// playground does not keep the archive: it checks it (themepack.Read, under
// the library's default limits: a package is untrusted input, a dropped file
// as much as a shared one) and writes it into the project as the files the
// project already has: themes/<name>.css, which is the theme as one Marp
// sheet with every font and image inlined as a data: URL (so the preview and
// the exports read nothing else), and, when the package brings kinds,
// kinds/<name>.yaml. Both are listed in slides.yaml like a theme or a kinds
// file made in the panel, and travel in the project's .zip.

// ThemeExt is the extension of a theme package.
const ThemeExt = themepack.Extension

// MaxThemePackage is the largest .slidestheme the page reads.
var MaxThemePackage = int(themepack.DefaultLimits.MaxArchive)

// IsThemePackage reports whether a file name is a theme package's.
func IsThemePackage(name string) bool {
	return strings.EqualFold(path.Ext(strings.ReplaceAll(name, `\`, "/")), ThemeExt)
}

// slidesVersion is the go-slides version this page was built with ("" when
// the build does not say: a package then is accepted whatever it needs).
func slidesVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, d := range bi.Deps {
		if d.Path == "github.com/go-slides/slides" {
			if d.Replace != nil {
				return d.Replace.Version
			}
			return d.Version
		}
	}
	return ""
}

// themeRefused says a package the library refused: its code, the library's
// own words (English, as its errors are, and they name the file and the
// value) and, in another language, what the code means in it.
func themeRefused(err error) error {
	code := diag.CodeOf(err)
	if code == "" {
		code = "theme"
	}
	msg := err.Error()
	if e := explainCode(code); e != "" {
		msg += " " + e
	}
	return errf(KErrThemeRefused, code, msg)
}

// kindsYAML is a package's kinds as a kinds file, in the package's order.
func kindsYAML(m *themepack.Meta) (string, error) {
	root := &yaml.Node{Kind: yaml.MappingNode,
		HeadComment: "Slide kinds of the theme package " + m.Name + " " + m.Version + ". A slide takes one with\n" +
			"  <!-- _kind: name -->   (this slide)\n  <!-- kind: name -->    (this slide and the next ones)"}
	for _, k := range m.Kinds {
		def := &yaml.Node{Kind: yaml.MappingNode}
		if k.Extends != "" {
			def.Content = append(def.Content, str("extends"), str(k.Extends))
		}
		for _, kv := range k.Dirs {
			def.Content = append(def.Content, str(kv[0]), str(kv[1]))
		}
		root.Content = append(root.Content, str(k.Name), def)
	}
	return encodeConfig(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}})
}

// textBytes is the text the renderer is sent besides the deck: what
// textFilesOf counts against MaxTextFiles.
func (vm *ViewModel) textBytes() int {
	p, total := vm.project(), 0
	for _, k := range p.Paths() {
		if k == DeckFile || IsMedia(k) {
			continue
		}
		if b, _ := p.Read(k); utf8.Valid(b) {
			total += len(b)
		}
	}
	return total
}

// ImportTheme puts a theme package into the project, as one step Undo takes
// back, and returns the path of the theme. Nothing is written when the
// package is refused: by the library (a checked archive, CSS, SVG, fonts,
// sums, versions), for a name that is a built-in theme or that the project
// has already, for a base theme the project lacks, or for text that would
// not fit what the renderer is sent. The same package again changes nothing.
func (vm *ViewModel) ImportTheme(data []byte) (string, error) {
	pkg, err := themepack.Read(bytes.NewReader(data), int64(len(data)), themepack.Limits{})
	if err == nil {
		err = pkg.CheckVersion(slidesVersion())
	}
	if err != nil {
		return "", themeRefused(err)
	}
	var out string
	var failed error
	vm.projectStep(T(KStepImportTheme, pkg.Meta.Name), func() bool {
		out, failed = vm.importTheme(pkg)
		return failed == nil
	})
	return out, failed
}

// OpenThemePackage is ImportTheme for a page that says what went wrong in
// its status line.
func (vm *ViewModel) OpenThemePackage(data []byte) {
	_, err := vm.ImportTheme(data)
	vm.fail(err)
}

func (vm *ViewModel) importTheme(pkg *themepack.Package) (string, error) {
	name := pkg.Meta.Name
	if slices.Contains(BuiltinThemes, name) || slices.Contains(themepack.ReservedNames, name) {
		return "", errf(KErrThemeBuiltin, name)
	}
	css, err := pkg.CSS()
	if err != nil {
		return "", themeRefused(err)
	}
	kinds := ""
	if len(pkg.Meta.Kinds) > 0 {
		if kinds, err = kindsYAML(pkg.Meta); err != nil {
			return "", themeRefused(err)
		}
		// what is written is read back as the library reads kinds
		if _, err := deck.ParseKindsYAML([]byte(kinds), name+".yaml"); err != nil {
			return "", themeRefused(err)
		}
	}
	p := vm.project()
	cssPath, kindsPath := path.Join("themes", name+".css"), path.Join("kinds", name+".yaml")

	// the files: free, or already these bytes
	same := map[string]bool{}
	for _, f := range []struct{ path, text string }{{cssPath, css}, {kindsPath, kinds}} {
		if f.text == "" {
			continue
		}
		old, ok := p.Read(f.path)
		switch {
		case ok && string(old) == f.text:
			same[f.path] = true
		case ok || vm.isFolder(f.path):
			if f.path == cssPath {
				return "", errf(KErrThemeTaken, name, f.path)
			}
			return "", errf(KErrExists, f.path)
		}
	}
	// a theme of this name, defined by another file, would be two themes of
	// one name
	for _, k := range p.Paths() {
		if k != cssPath && strings.EqualFold(path.Ext(k), ".css") {
			for _, m := range atTheme.FindAllStringSubmatch(p.Text(k), -1) {
				if m[1] == name {
					return "", errf(KErrThemeTaken, name, k)
				}
			}
		}
	}
	// its base theme: built in, or a theme of the project
	if b := pkg.Meta.Extends; b != "" && !slices.Contains(BuiltinThemes, b) {
		if !slices.Contains(ProjectThemeChoices("", vm.projectCSS()), b) {
			return "", errf(KErrThemeExtends, name, b)
		}
	}
	// the renderer is sent MaxTextFiles of text: a file past it would be
	// dropped there without a word, so it is refused here, with room for the
	// lines slides.yaml gains
	add := 0
	for _, f := range []struct{ path, text string }{{cssPath, css}, {kindsPath, kinds}} {
		if f.text != "" && !same[f.path] {
			add += len(f.text) + len(f.path) + 64
		}
	}
	if have := vm.textBytes(); have+add > MaxTextFiles {
		return "", errf(KErrThemeText, sizeText(add), sizeText(have), sizeText(MaxTextFiles))
	}

	now := vm.drafts.now()
	var made []string
	undo := func(err error) (string, error) {
		for _, k := range made {
			p.Remove(k)
		}
		return "", err
	}
	for _, f := range []struct{ path, text string }{{cssPath, css}, {kindsPath, kinds}} {
		if f.text == "" || same[f.path] {
			continue
		}
		if err := p.Write(f.path, []byte(f.text), now); err != nil {
			return undo(err)
		}
		made = append(made, f.path)
	}
	vm.listInConfig(cssPath)
	if kinds != "" {
		vm.listInConfig(kindsPath)
	}
	vm.filesChanged()
	msg := T(KStatusImportedTheme, name, cssPath, name)
	if kinds != "" {
		msg += " " + T(KStatusImportedKinds, kindsPath)
	}
	vm.notice(msg)
	return cssPath, nil
}
