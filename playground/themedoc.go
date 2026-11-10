package playground

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-slides/slides/themepack"
)

// The theme editor edits a theme in the form of its package: a folder
// themes/<name>.src/ in the project holds exactly what `slides theme pack`
// reads (theme.hcl, the sheet, fonts/, assets/, licence texts), so that the
// folder is the one source of truth. Everything else is derived from it:
//
//   - themes/<name>.css, the sheet the preview and the exports read, is the
//     package compiled by themepack (fonts and images inlined as data: URLs);
//   - the controls of the Theme pane are read from the folder and write to it;
//   - a package is packed from it, by the same library the CLI verifies with.
//
// The controls and the CSS stay in step through one block of the sheet, fenced
// by two comments (visualBegin, visualEnd). The editor writes only inside the
// fence; everything outside is the person's, byte for byte. A person who edits
// inside the fence is read back when the block is still what the editor would
// write for the values it holds; when it is not (a rule changed, a property
// added), the block is "diverged": the controls say so and refuse to write,
// rather than regenerate the block over the person's CSS.

// ThemeSrcRoot is the project folder that holds the sources of the themes. It is
// not under themes/, which slides.yaml lists whole as a set of sheets: a source
// sheet has no @theme comment and is not one.
const ThemeSrcRoot = "theme-src"

// ThemeSrcDir is the project folder holding the source of theme name.
func ThemeSrcDir(name string) string { return ThemeSrcRoot + "/" + name }

// ThemeNameOfSrc is the theme a path of a source folder belongs to ("" when the
// path is not in one).
func ThemeNameOfSrc(p string) string {
	rest, ok := strings.CutPrefix(p, ThemeSrcRoot+"/")
	if !ok {
		return ""
	}
	if name, _, found := strings.Cut(rest, "/"); found && themepack.ValidThemeName(name) {
		return name
	}
	return ""
}

// ThemeDoc is a theme package being edited: its metadata and its files.
type ThemeDoc struct {
	// Meta is what theme.hcl declares.
	Meta themepack.Meta
	// Files are the package's files but theme.hcl: the sheets, fonts, assets and
	// licence texts, by path inside the package.
	Files themepack.Files
}

// NewThemeDoc is a theme of its own that starts from base (built in or one of the
// project's), with an empty sheet.
func NewThemeDoc(name, base string) *ThemeDoc {
	return &ThemeDoc{
		Meta: themepack.Meta{
			Name: name, Version: "0.1.0", License: "MIT", Extends: base,
			CSS: []themepack.CSSFile{{Path: "theme.css"}},
		},
		Files: themepack.Files{"theme.css": []byte("/* The theme's own rules go here. */\n")},
	}
}

// DocFromPackage is the editing form of a checked package.
func DocFromPackage(pkg *themepack.Package) *ThemeDoc {
	d := &ThemeDoc{Meta: *pkg.Meta, Files: themepack.Files{}}
	for n, b := range pkg.Files {
		if n != themepack.MetaFile {
			d.Files[n] = bytes.Clone(b)
		}
	}
	return d
}

// DocFromFiles reads a theme from the files of its source folder (theme.hcl
// included). Only theme.hcl is parsed here; Check says whether the whole is a
// package.
func DocFromFiles(files themepack.Files) (*ThemeDoc, error) {
	src, ok := files[themepack.MetaFile]
	if !ok {
		return nil, fmt.Errorf("theme package: no %s at the root", themepack.MetaFile)
	}
	m, err := themepack.ParseMeta(src, themepack.Limits{})
	if err != nil {
		return nil, err
	}
	d := &ThemeDoc{Meta: *m, Files: themepack.Files{}}
	for n, b := range files {
		if n != themepack.MetaFile {
			d.Files[n] = b
		}
	}
	return d, nil
}

// slidesSheetImport finds the lines a loose sheet uses to name its base theme.
var slidesSheetImport = regexp.MustCompile(`(?m)^[ \t]*@import(?:-theme)?[ \t]+["']([A-Za-z0-9][A-Za-z0-9_.-]*)["'][ \t]*;?[ \t]*\r?\n?`)

// DocFromCSS is the editing form of a loose sheet: the sheet becomes theme.css,
// the name comes from its /* @theme name */ comment (or from fallback, the file's
// name), and a base theme named by an @import line moves into Extends, because a
// package names its base there and the library refuses the @import.
// The second result lists what was changed in the sheet, for the person.
func DocFromCSS(css, fallback string) (*ThemeDoc, []string) {
	var notes []string
	name := fallback
	if m := atTheme.FindStringSubmatch(css); m != nil {
		name = m[1]
	}
	name = SlugName(name)
	d := NewThemeDoc(name, "")
	if m := slidesSheetImport.FindStringSubmatch(css); m != nil {
		d.Meta.Extends = m[1]
		css = slidesSheetImport.ReplaceAllString(css, "")
		notes = append(notes, T(KThemeNoteImport, m[1]))
	}
	d.Meta.License = UnspecifiedLicense // not known: the person says
	d.Files["theme.css"] = []byte(css)
	return d, notes
}

// SlugName makes a theme name of text: lower case letters, digits and dashes.
func SlugName(s string) string {
	var b strings.Builder
	dash := true
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash:
			b.WriteByte('-')
			dash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 64 {
		out = strings.TrimRight(out[:64], "-")
	}
	if out == "" {
		return "theme"
	}
	return out
}

// CSSPath is the path of the sheet the controls write to: the first css block.
func (d *ThemeDoc) CSSPath() string {
	if len(d.Meta.CSS) == 0 {
		return "theme.css"
	}
	return d.Meta.CSS[0].Path
}

// Sheet is the text of that sheet.
func (d *ThemeDoc) Sheet() string { return string(d.Files[d.CSSPath()]) }

// SetSheet replaces the sheet's text (a hand edit, or the controls' write).
func (d *ThemeDoc) SetSheet(s string) { d.Files[d.CSSPath()] = []byte(s) }

// HCL is theme.hcl as the editor writes it: the fields in a fixed order, no
// checksums (a hand edit that adds them is kept: see Encode).
func (d *ThemeDoc) HCL() string {
	m := &d.Meta
	var b strings.Builder
	b.WriteString("theme {\n")
	kv := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, "  %-11s = %s\n", k, hclString(v))
		}
	}
	kv("name", m.Name)
	kv("version", m.Version)
	kv("description", m.Description)
	if len(m.Authors) > 0 {
		q := make([]string, len(m.Authors))
		for i, a := range m.Authors {
			q[i] = hclString(a)
		}
		fmt.Fprintf(&b, "  %-11s = [%s]\n", "authors", strings.Join(q, ", "))
	}
	kv("license", m.License)
	kv("homepage", m.Homepage)
	kv("extends", m.Extends)
	kv("min_slides", m.MinSlides)
	b.WriteString("}\n")
	for _, c := range m.CSS {
		if c.SHA256 != "" {
			fmt.Fprintf(&b, "\ncss %s {\n  sha256 = %s\n}\n", hclString(c.Path), hclString(c.SHA256))
		} else {
			fmt.Fprintf(&b, "\ncss %s {}\n", hclString(c.Path))
		}
	}
	for _, f := range m.Fonts {
		fmt.Fprintf(&b, "\nfont %s {\n  weight  = %s\n  style   = %s\n  file    = %s\n  license = %s\n", hclString(f.Family), hclWeight(f.Weight), hclString(f.Style), hclString(f.File), hclString(f.License))
		if f.SHA256 != "" {
			fmt.Fprintf(&b, "  sha256  = %s\n", hclString(f.SHA256))
		}
		b.WriteString("}\n")
	}
	for _, a := range m.Assets {
		fmt.Fprintf(&b, "\nasset %s {\n  file = %s\n", hclString(a.ID), hclString(a.File))
		if a.SHA256 != "" {
			fmt.Fprintf(&b, "  sha256 = %s\n", hclString(a.SHA256))
		}
		b.WriteString("}\n")
	}
	for _, k := range m.Kinds {
		fmt.Fprintf(&b, "\nkind %s {\n", hclString(k.Name))
		if k.Extends != "" {
			fmt.Fprintf(&b, "  extends = %s\n", hclString(k.Extends))
		}
		for _, kv := range k.Dirs {
			fmt.Fprintf(&b, "  %s = %s\n", kv[0], hclString(kv[1]))
		}
		b.WriteString("}\n")
	}
	return b.String()
}

func hclString(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`, "${", "$${", "%{", "%%{")
	return `"` + r.Replace(s) + `"`
}

func hclWeight(w string) string {
	if _, err := strconv.Atoi(w); err == nil {
		return w
	}
	return hclString(w)
}

// Encode is the package's files: the editor's theme.hcl and the rest. The sums
// the metadata declares are brought up to date, so that what a person edited
// (a font replaced, the sheet) still verifies.
func (d *ThemeDoc) Encode() themepack.Files {
	out := themepack.Files{}
	for n, b := range d.Files {
		out[n] = b
	}
	for i, c := range d.Meta.CSS {
		if c.SHA256 != "" {
			d.Meta.CSS[i].SHA256 = sum(d.Files[c.Path])
		}
	}
	for i, f := range d.Meta.Fonts {
		if f.SHA256 != "" {
			d.Meta.Fonts[i].SHA256 = sum(d.Files[f.File])
		}
	}
	for i, a := range d.Meta.Assets {
		if a.SHA256 != "" {
			d.Meta.Assets[i].SHA256 = sum(d.Files[a.File])
		}
	}
	out[themepack.MetaFile] = []byte(d.HCL())
	return out
}

// Check is the library's check of the package: the same one `slides theme
// verify` makes. Its errors are coded (theme.meta, theme.css, theme.svg,
// theme.font, theme.limit...), and shown as they are.
func (d *ThemeDoc) Check() (*themepack.Package, error) {
	return themepack.FromFiles(d.Encode(), themepack.Limits{})
}

// Pack is the package as a .gsta.xz, after Check.
func (d *ThemeDoc) Pack() ([]byte, error) {
	pkg, err := d.Check()
	if err != nil {
		return nil, err
	}
	data, err := themepack.WriteFiles(pkg.Files)
	if err != nil {
		return nil, err
	}
	// what is handed out is read back, under the limits a reader applies
	if _, err := themepack.Read(bytes.NewReader(data), int64(len(data)), themepack.Limits{}); err != nil {
		return nil, err
	}
	return data, nil
}

// PlainCSS is the theme as one sheet of CSS for a download: the package
// compiled, fonts and images inlined (themepack's CSS).
func (d *ThemeDoc) PlainCSS() (string, error) {
	pkg, err := d.Check()
	if err != nil {
		return "", err
	}
	return pkg.CSS()
}

// --- the controls' block ------------------------------------------------

const (
	visualBegin = "/* gsta:visual begin: written by the theme editor; edit the values here or the controls, the rest of the sheet is yours */"
	visualEnd   = "/* gsta:visual end */"
)

// PropKind says what a property's control is.
type PropKind int

// The kinds of property.
const (
	PropColor PropKind = iota
	PropFont
	PropNumber
	PropEnum
)

// ThemeProp is one property the controls edit: a custom property of the block.
type ThemeProp struct {
	ID    string
	Kind  PropKind
	Group string // palette, type, logo, footer
	// Number and Enum.
	Units    []string
	Min, Max int
	Step     int
	Decimals int
	Choices  []string
}

// The property groups, in the order the pane shows them.
const (
	GroupPalette = "palette"
	GroupType    = "type"
	GroupLogo    = "logo"
	GroupFooter  = "footer"
)

// Placements of the logo, and of the page number.
var (
	LogoPlacements       = []string{"none", "top-left", "top-right", "bottom-left", "bottom-right", "background"}
	PaginationPlacements = []string{"bottom-right", "bottom-left", "bottom-center", "top-right", "top-left", "top-center"}
	FooterAlignments     = []string{"left", "center", "right"}
)

// ThemeProps are the properties, in pane order.
var ThemeProps = []ThemeProp{
	{ID: "background", Kind: PropColor, Group: GroupPalette},
	{ID: "text", Kind: PropColor, Group: GroupPalette},
	{ID: "heading", Kind: PropColor, Group: GroupPalette},
	{ID: "accent", Kind: PropColor, Group: GroupPalette},
	{ID: "link", Kind: PropColor, Group: GroupPalette},
	{ID: "font-body", Kind: PropFont, Group: GroupType},
	{ID: "font-heading", Kind: PropFont, Group: GroupType},
	{ID: "font-code", Kind: PropFont, Group: GroupType},
	{ID: "size-body", Kind: PropNumber, Group: GroupType, Units: []string{"px", "pt"}, Min: 8, Max: 200, Step: 1},
	{ID: "size-heading", Kind: PropNumber, Group: GroupType, Units: []string{"px", "pt"}, Min: 8, Max: 300, Step: 1},
	{ID: "size-code", Kind: PropNumber, Group: GroupType, Units: []string{"px", "pt"}, Min: 8, Max: 200, Step: 1},
	{ID: "weight-body", Kind: PropNumber, Group: GroupType, Min: 100, Max: 900, Step: 100},
	{ID: "weight-heading", Kind: PropNumber, Group: GroupType, Min: 100, Max: 900, Step: 100},
	{ID: "logo-placement", Kind: PropEnum, Group: GroupLogo, Choices: LogoPlacements},
	{ID: "logo-size", Kind: PropNumber, Group: GroupLogo, Units: []string{"px"}, Min: 8, Max: 600, Step: 4},
	{ID: "pagination-placement", Kind: PropEnum, Group: GroupFooter, Choices: PaginationPlacements},
	{ID: "pagination-size", Kind: PropNumber, Group: GroupFooter, Units: []string{"px", "pt"}, Min: 6, Max: 120, Step: 1},
	{ID: "footer-align", Kind: PropEnum, Group: GroupFooter, Choices: FooterAlignments},
	{ID: "footer-size", Kind: PropNumber, Group: GroupFooter, Units: []string{"px", "pt"}, Min: 6, Max: 120, Step: 1},
}

// PropByID finds a property.
func PropByID(id string) (ThemeProp, bool) {
	for _, p := range ThemeProps {
		if p.ID == id {
			return p, true
		}
	}
	return ThemeProp{}, false
}

var (
	colorValue  = regexp.MustCompile(`^[A-Za-z0-9#%.,() /+-]{1,64}$`)
	fontValue   = regexp.MustCompile(`^[^"'\\;{}<>*/()@:\x00-\x1f\x7f]{1,64}$`)
	numberValue = regexp.MustCompile(`^-?[0-9]{1,5}(\.[0-9]{1,2})?(px|pt|em|rem|%)?$`)
)

// ErrThemeValue is a value the property does not take: a message in the
// reader's language.
func validValue(p ThemeProp, v string) bool {
	switch p.Kind {
	case PropColor:
		return colorValue.MatchString(v) && !strings.Contains(strings.ToLower(v), "url(")
	case PropFont:
		return fontValue.MatchString(v) && strings.TrimSpace(v) == v
	case PropNumber:
		return numberValue.MatchString(v)
	case PropEnum:
		return slices.Contains(p.Choices, v)
	}
	return false
}

// Visual is the values the block holds, by property id.
type Visual map[string]string

// SyncState says how the block stands with the sheet.
type SyncState int

// The block's states.
const (
	// BlockAbsent: the sheet has no block yet; the controls write one on the
	// first edit.
	BlockAbsent SyncState = iota
	// BlockInSync: the block is what the editor writes for its values.
	BlockInSync
	// BlockDiverged: the block was edited past what the controls can show; they
	// do not write until it is restored or emptied, so nothing of the person's is
	// overwritten.
	BlockDiverged
)

// blockRange finds the fenced block: [start, end) of the whole text, fence
// lines included, and whether there is exactly one.
func blockRange(css string) (start, end int, ok bool) {
	i := strings.Index(css, visualBeginPrefix)
	if i < 0 || strings.Count(css, visualBeginPrefix) != 1 {
		return 0, 0, false
	}
	j := strings.Index(css[i:], visualEnd)
	if j < 0 || strings.Count(css, visualEnd) != 1 {
		return 0, 0, false
	}
	end = i + j + len(visualEnd)
	if end < len(css) && css[end] == '\n' {
		end++
	}
	return i, end, true
}

const visualBeginPrefix = "/* gsta:visual begin"

var customProp = regexp.MustCompile(`--gsta-([a-z-]+)\s*:\s*([^;}]*);`)

// ReadVisual reads the values out of the sheet's block and says whether the
// block is still the editor's own.
func ReadVisual(css string, logoPath string) (Visual, SyncState) {
	v := Visual{}
	i, j, ok := blockRange(css)
	if !ok {
		if strings.Contains(css, visualBeginPrefix) || strings.Contains(css, visualEnd) {
			return v, BlockDiverged // a fence cut in half, or doubled
		}
		return v, BlockAbsent
	}
	block := css[i:j]
	for _, m := range customProp.FindAllStringSubmatch(block, -1) {
		if p, ok := PropByID(m[1]); ok {
			if val := strings.TrimSpace(m[2]); validValue(p, unquoteFont(p, val)) {
				v[m[1]] = unquoteFont(p, val)
			}
		}
	}
	if strings.TrimRight(blockText(v, logoPath), "\n") == strings.TrimRight(block, "\n") {
		return v, BlockInSync
	}
	return v, BlockDiverged
}

// unquoteFont is a font property's value without the quotes it is written with.
func unquoteFont(p ThemeProp, val string) string {
	if p.Kind == PropFont && len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
		return val[1 : len(val)-1]
	}
	return val
}

// WriteVisual puts the block for v into the sheet: in place of the one there,
// else at the end (it must win over the rules before it). With no value the
// block is removed. It reports false, and leaves the sheet alone, when the block
// is diverged.
func WriteVisual(css string, v Visual, logoPath string) (string, bool) {
	i, j, ok := blockRange(css)
	if !ok {
		if strings.Contains(css, visualBeginPrefix) || strings.Contains(css, visualEnd) {
			return css, false
		}
		if len(v) == 0 {
			return css, true
		}
		if css != "" && !strings.HasSuffix(css, "\n") {
			css += "\n"
		}
		return css + "\n" + blockText(v, logoPath), true
	}
	if _, st := ReadVisual(css, logoPath); st == BlockDiverged {
		return css, false
	}
	if len(v) == 0 {
		return strings.TrimRight(css[:i], "\n") + "\n" + css[j:], true
	}
	return css[:i] + blockText(v, logoPath) + css[j:], true
}

// blockText is the block for v: the custom properties, then one rule per
// group that has a value, so that a sheet without values writes nothing and a
// sheet with a colour does not also change the fonts.
func blockText(v Visual, logoPath string) string {
	if len(v) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(visualBegin + "\n")
	b.WriteString("section {\n")
	for _, p := range ThemeProps {
		if val, ok := v[p.ID]; ok {
			if p.Kind == PropFont {
				val = fontQuoted(val)
			}
			fmt.Fprintf(&b, "  --gsta-%s: %s;\n", p.ID, val)
		}
	}
	b.WriteString("}\n")
	rule := func(sel string, decls ...[2]string) {
		var out []string
		for _, d := range decls {
			if d[1] != "" {
				out = append(out, "  "+d[0]+": "+d[1]+";")
			}
		}
		if len(out) > 0 {
			b.WriteString(sel + " {\n" + strings.Join(out, "\n") + "\n}\n")
		}
	}
	use := func(id, fallback string) string {
		if _, ok := v[id]; !ok {
			return ""
		}
		if fallback != "" {
			return "var(--gsta-" + id + "), " + fallback
		}
		return "var(--gsta-" + id + ")"
	}
	rule("section",
		[2]string{"background-color", use("background", "")},
		[2]string{"color", use("text", "")},
		[2]string{"font-family", use("font-body", "sans-serif")},
		[2]string{"font-size", use("size-body", "")},
		[2]string{"font-weight", use("weight-body", "")},
	)
	rule("h1, h2, h3, h4, h5, h6",
		[2]string{"color", use("heading", "")},
		[2]string{"font-family", use("font-heading", "sans-serif")},
		[2]string{"font-weight", use("weight-heading", "")},
	)
	if _, ok := v["size-heading"]; ok {
		rule("h1", [2]string{"font-size", "var(--gsta-size-heading)"})
	}
	rule("a", [2]string{"color", use("link", "")})
	rule("blockquote, hr", [2]string{"border-color", use("accent", "")})
	rule("strong, li::marker", [2]string{"color", use("accent", "")})
	rule("code, pre, kbd", [2]string{"font-family", use("font-code", "monospace")}, [2]string{"font-size", use("size-code", "")})
	if pl := v["pagination-placement"]; pl != "" || v["pagination-size"] != "" {
		decls := [][2]string{{"font-size", use("pagination-size", "")}}
		if pl != "" {
			decls = append(decls, paginationPlace(pl)...)
		}
		rule("section::after", decls...)
	}
	if v["footer-align"] != "" || v["footer-size"] != "" {
		decls := [][2]string{{"font-size", use("footer-size", "")}}
		if a := v["footer-align"]; a != "" {
			decls = append(decls, [2]string{"text-align", a}, [2]string{"left", "0"}, [2]string{"right", "0"})
		}
		rule("section > footer", decls...)
	}
	if pl := v["logo-placement"]; pl != "" && pl != "none" && logoPath != "" {
		size := "var(--gsta-logo-size, 64px)"
		if pl == "background" {
			rule("section", [2]string{"background-image", `url("` + logoPath + `")`},
				[2]string{"background-repeat", "no-repeat"}, [2]string{"background-position", "center"},
				[2]string{"background-size", "cover"})
		} else {
			y, x, _ := strings.Cut(pl, "-")
			rule("section::before", [2]string{"content", `""`}, [2]string{"position", "absolute"},
				[2]string{y, "24px"}, [2]string{x, "24px"}, [2]string{"width", size}, [2]string{"height", size},
				[2]string{"background", `url("` + logoPath + `") no-repeat center / contain`})
		}
	}
	b.WriteString(visualEnd + "\n")
	return b.String()
}

// fontQuoted is a family as the block writes it: a generic keyword as it is, a
// name in quotes.
func fontQuoted(name string) string {
	switch name {
	case "serif", "sans-serif", "monospace", "cursive", "fantasy", "system-ui", "ui-serif", "ui-sans-serif", "ui-monospace":
		return name
	}
	return `"` + name + `"`
}

func paginationPlace(pl string) [][2]string {
	y, x, _ := strings.Cut(pl, "-")
	out := [][2]string{{y, "20px"}}
	switch x {
	case "left":
		out = append(out, [2]string{"left", "30px"})
	case "right":
		out = append(out, [2]string{"right", "30px"})
	default:
		out = append(out, [2]string{"left", "0"}, [2]string{"right", "0"}, [2]string{"text-align", "center"})
	}
	return out
}

// --- reading and writing the controls on the document -----------------------

// LogoPath is the path the sheet names the logo by ("" without one): the asset
// "logo" of the package, as a reference relative to the sheet.
func (d *ThemeDoc) LogoPath() string {
	for _, a := range d.Meta.Assets {
		if a.ID == "logo" {
			return relTo(d.CSSPath(), a.File)
		}
	}
	return ""
}

// relTo writes target as a reference from the directory of from.
func relTo(from, target string) string {
	dir := path.Dir(from)
	if dir == "." || dir == "" {
		return target
	}
	return strings.Repeat("../", strings.Count(dir, "/")+1) + target
}

// Visual reads the values and the block's state from the sheet.
func (d *ThemeDoc) Visual() (Visual, SyncState) { return ReadVisual(d.Sheet(), d.LogoPath()) }

// ErrDiverged is the refusal of a write into a block the person edited by hand.
var ErrDiverged = fmt.Errorf("the block is edited by hand")

// SetVisual sets one property (an empty value removes it). It writes only the
// block, and refuses (ErrDiverged) rather than overwrite a block edited by hand.
func (d *ThemeDoc) SetVisual(id, val string) error {
	p, ok := PropByID(id)
	if !ok {
		return fmt.Errorf("unknown property %q", id)
	}
	if val != "" && !validValue(p, val) {
		return fmt.Errorf("%q is not a value of %s", val, id)
	}
	v, _ := d.Visual()
	if val == "" {
		delete(v, id)
	} else {
		v[id] = val
	}
	css, ok := WriteVisual(d.Sheet(), v, d.LogoPath())
	if !ok {
		return ErrDiverged
	}
	d.SetSheet(css)
	return nil
}

// RestoreVisual rewrites the block from the values it holds, which drops what
// the person added inside the fence (and only that).
func (d *ThemeDoc) RestoreVisual() {
	v, _ := d.Visual()
	if !d.putBlock(v) {
		// a fence cut in half or doubled: the strays go, a block is written anew
		css := strings.ReplaceAll(d.Sheet(), visualEnd+"\n", "")
		for {
			i := strings.Index(css, visualBeginPrefix)
			if i < 0 {
				break
			}
			end := strings.IndexByte(css[i:], '\n')
			if end < 0 {
				end = len(css) - i - 1
			}
			css = css[:i] + css[i+end+1:]
		}
		css = strings.ReplaceAll(css, visualEnd, "")
		css, _ = WriteVisual(css, v, d.LogoPath())
		d.SetSheet(css)
	}
}

// putBlock writes the block for v in the place of the one in the sheet; false
// when the sheet has no (single) block.
func (d *ThemeDoc) putBlock(v Visual) bool {
	css := d.Sheet()
	i, j, ok := blockRange(css)
	if !ok {
		return false
	}
	text := blockText(v, d.LogoPath())
	if text == "" {
		d.SetSheet(strings.TrimRight(css[:i], "\n") + "\n" + css[j:])
	} else {
		d.SetSheet(css[:i] + text + css[j:])
	}
	return true
}

// --- fonts and the logo ------------------------------------------------------

// FontExts are the font files a package takes, by extension.
var FontExts = []string{".woff2", ".woff", ".ttf", ".otf"}

// ImageExts are the images a logo may be.
var ImageExts = []string{".svg", ".png", ".jpg", ".jpeg", ".webp", ".gif"}

// FamilyOfFile guesses a family from a font file's name: "Inter-Bold.woff2" is
// "Inter".
func FamilyOfFile(name string) string {
	base := strings.TrimSuffix(path.Base(name), path.Ext(name))
	for _, cut := range []string{"-", "_"} {
		if i := strings.Index(base, cut); i > 0 {
			base = base[:i]
		}
	}
	fam := strings.Map(func(r rune) rune {
		if strings.ContainsRune("\"'\\{};<>/*()@\n\r", r) {
			return -1
		}
		return r
	}, base)
	if len(fam) > 64 {
		fam = fam[:64]
	}
	if fam = strings.TrimSpace(fam); fam == "" {
		return "Custom"
	}
	return fam
}

// UnspecifiedLicense is what an added font carries until the person says its
// licence: a valid licence reference that claims nothing. The Check passes it;
// the pane shows it as something to fill.
const UnspecifiedLicense = "LicenseRef-unspecified"

// AddFont puts a font file in the package, as a face of family (the file's own
// family when empty), and makes the files' check refuse nothing the library
// would not. It returns the family. Two faces of one family may differ in
// weight and style; the same file name twice is replaced.
func (d *ThemeDoc) AddFont(file string, data []byte, family, weight, style string) (string, error) {
	ext := strings.ToLower(path.Ext(file))
	if !slices.Contains(FontExts, ext) {
		return "", fmt.Errorf("%s is not a font file (%s)", path.Base(file), strings.Join(FontExts, ", "))
	}
	if family == "" {
		family = FamilyOfFile(file)
	}
	if weight == "" {
		weight = "400"
	}
	if style == "" {
		style = "normal"
	}
	rel := "fonts/" + safeFileName(path.Base(file))
	trial := d.clone()
	trial.Files[rel] = bytes.Clone(data)
	face := themepack.Font{Family: family, Weight: weight, Style: style, File: rel, License: UnspecifiedLicense}
	replaced := false
	for i, f := range trial.Meta.Fonts {
		if f.File == rel {
			face.License = f.License
			trial.Meta.Fonts[i] = face
			replaced = true
		}
	}
	if !replaced {
		trial.Meta.Fonts = append(trial.Meta.Fonts, face)
	}
	if _, err := trial.Check(); err != nil {
		return "", err
	}
	d.Meta, d.Files = trial.Meta, trial.Files
	return family, nil
}

// SetLogo puts an image in the package as the asset "logo" (and takes the old
// one out), and gives the logo a placement if it had none.
func (d *ThemeDoc) SetLogo(file string, data []byte) error {
	ext := strings.ToLower(path.Ext(file))
	if !slices.Contains(ImageExts, ext) {
		return fmt.Errorf("%s is not an image a logo can be (%s)", path.Base(file), strings.Join(ImageExts, ", "))
	}
	trial := d.clone()
	trial.swapLogo(func(x *ThemeDoc) {
		x.dropLogo()
		rel := "assets/logo" + ext
		x.Files[rel] = bytes.Clone(data)
		x.Meta.Assets = append(x.Meta.Assets, themepack.Asset{ID: "logo", File: rel})
	})
	if _, err := trial.Check(); err != nil {
		return err
	}
	d.Meta, d.Files = trial.Meta, trial.Files
	if v, _ := d.Visual(); v["logo-placement"] == "" {
		_ = d.SetVisual("logo-placement", "top-right")
	}
	return nil
}

// RemoveLogo takes the logo out and its rule with it.
func (d *ThemeDoc) RemoveLogo() {
	d.swapLogo(func(x *ThemeDoc) { x.dropLogo() })
}

// swapLogo changes the logo with change, and when the block is the editor's
// own writes it again for the logo there is now (its url() names the file).
// A block edited by hand is left as it is: a url() in it that names a file now
// gone is then the library's to refuse, with its code.
func (d *ThemeDoc) swapLogo(change func(*ThemeDoc)) {
	v, st := d.Visual()
	change(d)
	if st == BlockInSync {
		d.putBlock(v)
	}
}

func (d *ThemeDoc) dropLogo() {
	var keep []themepack.Asset
	for _, a := range d.Meta.Assets {
		if a.ID == "logo" {
			delete(d.Files, a.File)
			continue
		}
		keep = append(keep, a)
	}
	d.Meta.Assets = keep
}

func (d *ThemeDoc) clone() *ThemeDoc {
	c := &ThemeDoc{Meta: d.Meta, Files: themepack.Files{}}
	c.Meta.Authors = slices.Clone(d.Meta.Authors)
	c.Meta.CSS = slices.Clone(d.Meta.CSS)
	c.Meta.Fonts = slices.Clone(d.Meta.Fonts)
	c.Meta.Assets = slices.Clone(d.Meta.Assets)
	c.Meta.Kinds = slices.Clone(d.Meta.Kinds)
	for n, b := range d.Files {
		c.Files[n] = b
	}
	return c
}

// FontFamilies are the families the package declares, in order.
func (d *ThemeDoc) FontFamilies() []string {
	var out []string
	for _, f := range d.Meta.Fonts {
		if !slices.Contains(out, f.Family) {
			out = append(out, f.Family)
		}
	}
	return out
}

// safeFileName keeps a file name to what a package path takes.
func safeFileName(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		}
		return '-'
	}, s)
}

// BundledFamilies are the families a theme may name without bringing a file:
// the generic keywords and the families every browser and office suite has.
var BundledFamilies = []string{
	"sans-serif", "serif", "monospace", "system-ui",
	"Arial", "Helvetica", "Verdana", "Tahoma", "Trebuchet MS", "Georgia", "Times New Roman", "Palatino", "Courier New",
}

// sum is a sha256 as the metadata writes it.
func sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
