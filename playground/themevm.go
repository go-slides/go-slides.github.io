package playground

import (
	"bytes"
	"fmt"
	"path"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-slides/go-slides.github.io/playground/versions"
	"github.com/go-slides/slides/diag"
	"github.com/go-slides/slides/themepack"
	"github.com/go-widgets/mvvm"
)

// The Theme pane (scene_theme.go): a column of the splitter that edits a theme
// of the project visually while the preview follows. The theme is kept as a
// package folder (themedoc.go: ThemeSrcDir) and the sheet the preview reads,
// themes/<name>.css, is compiled from it by the library the CLI verifies with,
// so what is exported is what is shown.
//
// Every edit of the controls is a step of the project's undo (history.go),
// which also keeps the files it replaced as a version; edits of one control in
// a row are one step. A hand edit of the sheet or of theme.hcl in the editor is
// an edit of the editor's own history, and the pane follows it.

// themeWidth and its range, in logical pixels.
const themeWidth, themeMin, themeMax = 340, 260, 640

// burstWindow is how long edits of one control in a row stay one undo step.
const burstWindow = 2 * time.Second

// themeState is the pane's part of the view-model.
type themeState struct {
	// ThemePane is the pane on or off; ThemeName the theme it edits ("": none);
	// ThemeRev counts the changes of what it shows (the theme's files, which
	// theme), for the pane to read the values again; ThemeNote says what the
	// pane needs the person to know (the block edited by hand, no theme).
	ThemePane *mvvm.Observable[bool]
	ThemeName *mvvm.Observable[string]
	ThemeRev  *mvvm.Observable[int]
	ThemeNote *mvvm.Observable[string]
	// The pane's commands.
	ThemeNew, ThemeImport, ThemePackage, ThemeCSS, ThemeSave *mvvm.Command
	ThemeFont, ThemeLogo, ThemeNoLogo, ThemeOpenCSS          *mvvm.Command
	ThemeRestore                                             *mvvm.Command

	// themeDiag are the lines the library's check last gave (shown in Problems
	// with the render's), themeSig the files of the edited theme as last seen.
	themeDiag []string
	themeSig  string
	themeBusy bool
	// the burst of edits that is the top step of the undo
	burst struct {
		key string
		at  time.Time
		ver int
	}
}

func (vm *ViewModel) initTheme() {
	vm.ThemePane = mvvm.NewObservable(false)
	vm.ThemeName = mvvm.NewObservable("")
	vm.ThemeRev = mvvm.NewObservable(0)
	vm.ThemeNote = mvvm.NewObservable("")
	pick := func(what string) func() {
		return func() {
			if p, ok := vm.host.(Picker); ok {
				p.Pick(what)
			}
		}
	}
	has := func() bool { return vm.ThemeName.Get() != "" }
	vm.ThemeNew = mvvm.NewCommand(func() { vm.NewThemeInEditor(vm.deckBase()) }, nil)
	vm.ThemeImport = mvvm.NewCommand(pick("themeedit"), nil)
	vm.ThemePackage = mvvm.NewCommand(vm.ExportThemePackage, has)
	vm.ThemeCSS = mvvm.NewCommand(vm.ExportThemeCSS, has)
	vm.ThemeSave = mvvm.NewCommand(vm.SaveThemeVersion, has)
	vm.ThemeFont = mvvm.NewCommand(pick("themefont"), has)
	vm.ThemeLogo = mvvm.NewCommand(pick("themelogo"), has)
	vm.ThemeNoLogo = mvvm.NewCommand(vm.RemoveThemeLogo, has)
	vm.ThemeOpenCSS = mvvm.NewCommand(vm.OpenThemeSheet, has)
	vm.ThemeRestore = mvvm.NewCommand(vm.RestoreThemeBlock, has)
	vm.ThemeName.Subscribe(func(name string) {
		for _, c := range []*mvvm.Command{vm.ThemePackage, vm.ThemeCSS, vm.ThemeSave, vm.ThemeFont, vm.ThemeLogo, vm.ThemeNoLogo, vm.ThemeOpenCSS, vm.ThemeRestore} {
			c.RaiseCanExecuteChanged()
		}
		vm.themeSig = ""
		if name != "" {
			vm.themeSig = vm.sourceSig(name)
		}
		vm.refreshThemeNote()
		vm.ThemeRev.Set(vm.ThemeRev.Get() + 1)
	})
	vm.ThemePane.Subscribe(func(on bool) {
		if on && vm.ThemeName.Get() == "" {
			vm.ThemeName.Set(vm.themeToEdit())
			vm.refreshThemeNote() // still no theme: the pane says what to do
		}
	})
	vm.refreshThemeNote()
}

// PickTheme makes name the theme the pane edits, and the deck's theme, so that
// the preview shows the edits (the deck's front matter, as the Theme picker
// writes it).
func (vm *ViewModel) PickTheme(name string) {
	vm.ThemeName.Set(name)
	if name == "" || !themeName.MatchString(name) {
		return
	}
	if cur, _ := DeckTheme(vm.Source.Get()); cur != name {
		vm.Source.Set(SetTheme(vm.Source.Get(), name))
	}
}

// deckBase is the theme a new theme starts from: the deck's, when it is built
// in or the project's, else default.
func (vm *ViewModel) deckBase() string {
	if t, ok := DeckTheme(vm.Source.Get()); ok && themeName.MatchString(t) {
		return t
	}
	return "default"
}

// EditableThemes are the project's themes the pane can edit: those it keeps a
// source for, and the other themes the project's sheets define (which it takes
// over as loose sheets).
func (vm *ViewModel) EditableThemes() []string {
	var out []string
	for _, t := range vm.Themes {
		if !slices.Contains(BuiltinThemes, t) && vm.themeSheetPath(t) != "" || vm.hasThemeSource(t) {
			out = append(out, t)
		}
	}
	return out
}

// themeToEdit is the theme the pane starts on: the deck's if it is the project's.
func (vm *ViewModel) themeToEdit() string {
	ts := vm.EditableThemes()
	if t, ok := DeckTheme(vm.Source.Get()); ok && slices.Contains(ts, t) {
		return t
	}
	if len(ts) > 0 {
		return ts[0]
	}
	return ""
}

func (vm *ViewModel) hasThemeSource(name string) bool {
	_, ok := vm.project().Read(ThemeSrcDir(name) + "/" + themepack.MetaFile)
	return ok
}

// themeSheetPath is the file of the project that defines theme name ("" when
// none does): themes/<name>.css, else a sheet whose @theme names it.
func (vm *ViewModel) themeSheetPath(name string) string {
	p := vm.project()
	own := path.Join("themes", name+".css")
	if _, ok := p.Read(own); ok {
		return own
	}
	for _, k := range p.Paths() {
		if strings.HasPrefix(k, "themes/") && strings.EqualFold(path.Ext(k), ".css") {
			for _, m := range atTheme.FindAllStringSubmatch(p.Text(k), -1) {
				if m[1] == name {
					return k
				}
			}
		}
	}
	return ""
}

// themeFiles reads the source folder of theme name: the files by their path
// inside the package.
func (vm *ViewModel) themeFiles(name string) themepack.Files {
	p, root := vm.project(), ThemeSrcDir(name)+"/"
	files := themepack.Files{}
	for _, k := range p.Paths() {
		if rel, ok := strings.CutPrefix(k, root); ok {
			b, _ := p.Read(k)
			files[rel] = b
		}
	}
	return files
}

// themeDoc is the edited theme as its source folder holds it. A theme that is
// the project's but has no source is taken over: its sheet becomes the source's.
// ok is false, with the reason in Problems, when there is no theme or its
// theme.hcl is not readable (the person fixes it in the editor).
func (vm *ViewModel) themeDoc() (*ThemeDoc, bool) {
	name := vm.ThemeName.Get()
	if name == "" {
		return nil, false
	}
	files := vm.themeFiles(name)
	if len(files) == 0 {
		sheet := vm.themeSheetPath(name)
		if sheet == "" {
			return nil, false
		}
		d, _ := DocFromCSS(vm.project().Text(sheet), name)
		d.Meta.Name = name
		return d, true
	}
	d, err := DocFromFiles(files)
	if err != nil {
		vm.themeProblem(err)
		return nil, false
	}
	return d, true
}

// themeProblem puts a library error where the render's problems are, with what
// its code means in the language.
func (vm *ViewModel) themeProblem(err error) {
	vm.themeDiag = nil
	if err != nil {
		vm.themeDiag = []string{themeProblemLine(vm.ThemeName.Get(), err)}
	}
	vm.showProblems()
}

func themeProblemLine(name string, err error) string {
	code := diag.CodeOf(err)
	line := T(KProblemError) + " " + ThemeSrcDir(name) + ": " + err.Error()
	if why := explainCode(code); why != "" {
		line += " — " + why
	}
	return line
}

// showProblems sets Problems: the last render's diagnostics, then the theme's.
func (vm *ViewModel) showProblems() {
	out := formatProblems(vm.Diagnostics)
	for _, l := range vm.themeDiag {
		if out != "" {
			out += "\n"
		}
		out += l
	}
	vm.Problems.Set(out)
}

// themeStore writes the document: its files into the source folder (what it
// dropped taken out), and the sheet compiled from it into themes/. A document
// the library refuses is still stored (the person's text is never lost) and
// the refusal is shown; the last sheet that compiled stays what the preview
// shows.
func (vm *ViewModel) themeStore(d *ThemeDoc) error {
	vm.themeBusy = true
	defer func() { vm.themeBusy = false }()
	p, now, root := vm.project(), vm.drafts.now(), ThemeSrcDir(d.Meta.Name)+"/"
	files := d.Encode()
	for _, k := range p.Paths() {
		if rel, ok := strings.CutPrefix(k, root); ok {
			if _, keep := files[rel]; !keep {
				p.Remove(k)
			}
		}
	}
	for rel, b := range files {
		if cur, ok := p.Read(root + rel); !ok || !bytes.Equal(cur, b) {
			if err := p.Write(root+rel, b, now); err != nil {
				return err
			}
		}
	}
	if err := vm.themeCompile(d); err != nil {
		return err
	}
	vm.themeSig = vm.sourceSig(d.Meta.Name)
	vm.filesChanged()
	vm.refreshThemeNote()
	vm.ThemeRev.Set(vm.ThemeRev.Get() + 1)
	return nil
}

// themeCompile checks the document and writes the sheet the preview reads (and
// the kinds file). The library's refusal goes to Problems.
func (vm *ViewModel) themeCompile(d *ThemeDoc) error {
	pkg, err := d.Check()
	if err != nil {
		vm.themeProblem(err)
		return nil // stored, shown, not compiled
	}
	css, err := pkg.CSS()
	if err != nil {
		vm.themeProblem(err)
		return nil
	}
	name, p, now := d.Meta.Name, vm.project(), vm.drafts.now()
	target := vm.themeSheetPath(name)
	if target == "" {
		target = path.Join("themes", name+".css")
	}
	if cur, ok := p.Read(target); !ok || string(cur) != css {
		if err := p.Write(target, []byte(css), now); err != nil {
			return err
		}
		vm.listInConfig(target)
	}
	if len(pkg.Meta.Kinds) > 0 {
		kinds, err := kindsYAML(pkg.Meta)
		if err != nil {
			vm.themeProblem(err)
			return nil
		}
		kp := path.Join("kinds", name+".yaml")
		if cur, ok := p.Read(kp); !ok || string(cur) != kinds {
			if err := p.Write(kp, []byte(kinds), now); err != nil {
				return err
			}
			vm.listInConfig(kp)
		}
	}
	vm.themeProblem(nil)
	return nil
}

// sourceSig identifies the files of a theme's source, to see a change.
func (vm *ViewModel) sourceSig(name string) string {
	p, root := vm.project(), ThemeSrcDir(name)+"/"
	var b strings.Builder
	for _, k := range p.Paths() {
		if strings.HasPrefix(k, root) {
			f, _ := p.File(k)
			fmt.Fprintf(&b, "%s %d %d\n", k, len(f.Data), f.Modified.UnixNano())
		}
	}
	return b.String()
}

// themeSync follows a change of the files (filesChanged calls it): a hand edit
// of the theme's sheet or of theme.hcl compiles again, and the pane reads the
// values again.
func (vm *ViewModel) themeSync() {
	name := vm.ThemeName.Get()
	if name == "" || vm.themeBusy {
		return
	}
	sig := vm.sourceSig(name)
	if sig == vm.themeSig {
		return
	}
	vm.themeSig = sig
	if sig != "" {
		if files := vm.themeFiles(name); len(files) > 0 {
			if d, err := DocFromFiles(files); err != nil {
				vm.themeProblem(err)
			} else {
				vm.themeBusy = true
				_ = vm.themeCompile(d)
				vm.themeBusy = false
			}
		}
	}
	vm.refreshThemeNote()
	vm.ThemeRev.Set(vm.ThemeRev.Get() + 1)
}

// refreshThemeNote says what the pane needs the person to know.
func (vm *ViewModel) refreshThemeNote() {
	note := ""
	switch d, ok := vm.themeDoc(); {
	case vm.ThemeName.Get() == "":
		note = T(KThemeNoteNone)
	case ok:
		if _, st := d.Visual(); st == BlockDiverged {
			note = T(KThemeNoteDiverged)
		}
	}
	vm.ThemeNote.Set(note)
}

// A step of the controls: the same control in a row, within the window, is one
// step of the undo (a dragged colour is one step, not a hundred).
func (vm *ViewModel) themeStep(key, label string, op func() bool) {
	now := vm.drafts.now()
	if b := &vm.burst; b.key == key && now.Sub(b.at) < burstWindow && b.ver == vm.undo.project.Version().Get() && !vm.undo.applying {
		if op() {
			b.at = now
		}
		return
	}
	was := vm.undo.project.Version().Get()
	vm.projectStep(label, op)
	if v := vm.undo.project.Version().Get(); v != was { // a step was recorded: it is the burst's
		vm.burst.key, vm.burst.at, vm.burst.ver = key, now, v
	} else {
		vm.burst.key = ""
	}
}

// editTheme runs an edit of the document as a step.
func (vm *ViewModel) editTheme(key, label string, mutate func(d *ThemeDoc) error) {
	d, ok := vm.themeDoc()
	if !ok {
		vm.notice(T(KThemeNoteNone))
		return
	}
	_, wasFine := d.Check()
	vm.themeStep(key, label, func() bool {
		if err := mutate(d); err != nil {
			vm.fail(err)
			return false
		}
		// A theme that checked before an edit and does not after it is not stored:
		// the files stay as they were (so that the controls can still read them, as a
		// half-typed version is not a version) and the library's word is shown. A
		// theme that did not check before (a hand edit the person is still making)
		// takes the edit, and the word is shown as well.
		if _, err := d.Check(); err != nil && wasFine == nil {
			vm.themeProblem(err)
			vm.fail(themeRefused(err))
			return false
		}
		if vm.fail(vm.themeStore(d)) {
			return false
		}
		return true
	})
}

// SetThemeValue sets one property of the block (an empty value removes it).
func (vm *ViewModel) SetThemeValue(id, val string) {
	label := T(KStepThemeEdit, themePropLabel(id))
	vm.editTheme("v:"+id, label, func(d *ThemeDoc) error {
		if err := d.SetVisual(id, val); err != nil {
			if err == ErrDiverged {
				return errf(KErrThemeDiverged)
			}
			return errf(KErrThemeValue, val, themePropLabel(id))
		}
		return nil
	})
}

// SetThemeMeta sets a field of theme.hcl: version, description, authors (one
// line, "; " between), license, homepage, extends, min_slides.
func (vm *ViewModel) SetThemeMeta(field, val string) {
	vm.editTheme("m:"+field, T(KStepThemeEdit, themeMetaLabel(field)), func(d *ThemeDoc) error {
		m := &d.Meta
		switch field {
		case "version":
			m.Version = val
		case "description":
			m.Description = val
		case "authors":
			m.Authors = nil
			for a := range strings.SplitSeq(val, ";") {
				if a = strings.TrimSpace(a); a != "" {
					m.Authors = append(m.Authors, a)
				}
			}
		case "license":
			m.License = val
		case "homepage":
			m.Homepage = val
		case "extends":
			m.Extends = val
		case "min_slides":
			m.MinSlides = val
		default:
			if fam, ok := strings.CutPrefix(field, "font-license:"); ok {
				for i := range m.Fonts {
					if m.Fonts[i].Family == fam {
						m.Fonts[i].License = val
					}
				}
			}
		}
		return nil
	})
}

// RestoreThemeBlock rewrites the block from the values it holds, which drops
// what was written inside its fence and nothing else.
func (vm *ViewModel) RestoreThemeBlock() {
	vm.editTheme("restore", T(KStepThemeEdit, T(KThemeBtnRestore)), func(d *ThemeDoc) error {
		d.RestoreVisual()
		return nil
	})
}

// NewThemeInEditor makes a theme built on base, as a package folder and its
// sheet, makes it the deck's theme and the pane's.
func (vm *ViewModel) NewThemeInEditor(base string) {
	if !themeName.MatchString(base) {
		base = "default"
	}
	name, _ := vm.uniqueName("themes", "my-"+SlugName(base), ".css")
	for vm.hasThemeSource(name) {
		name += "-2"
	}
	vm.projectStep(T(KStepThemeNew, name), func() bool {
		d := NewThemeDoc(name, base)
		if vm.fail(vm.themeStore(d)) {
			return false
		}
		vm.Source.Set(SetTheme(vm.Source.Get(), name))
		return true
	})
	vm.ThemePane.Set(true)
	vm.ThemeName.Set(name)
	vm.notice(T(KStatusThemeMade, name, base))
}

// ThemeImportFile opens a file in the editor: a theme package (checked as the
// library checks every package, its refusals shown in Problems) or a loose
// style sheet (taken as the theme's sheet).
func (vm *ViewModel) ThemeImportFile(name string, data []byte) {
	if IsThemePackage(name) {
		css, err := vm.ImportTheme(data)
		if err != nil {
			vm.themeImportRefused(err)
			return
		}
		vm.themeProblem(nil)
		vm.ThemePane.Set(true)
		vm.PickTheme(strings.TrimSuffix(path.Base(css), ".css"))
		return
	}
	if !strings.EqualFold(path.Ext(name), ".css") || !utf8.Valid(data) {
		vm.fail(errf(KErrThemeNotSheet, path.Base(name)))
		return
	}
	base := strings.TrimSuffix(path.Base(strings.ReplaceAll(name, `\`, "/")), path.Ext(name))
	d, notes := DocFromCSS(string(data), base)
	if slices.Contains(BuiltinThemes, d.Meta.Name) || slices.Contains(themepack.ReservedNames, d.Meta.Name) {
		vm.themeImportRefused(errf(KErrThemeBuiltin, d.Meta.Name))
		return
	}
	if vm.hasThemeSource(d.Meta.Name) || vm.themeSheetPath(d.Meta.Name) != "" {
		vm.themeImportRefused(errf(KErrThemeTaken, d.Meta.Name, vm.themeSheetPath(d.Meta.Name)))
		return
	}
	// stored, and the library's word on what the package would lack is shown
	vm.projectStep(T(KStepImportTheme, d.Meta.Name), func() bool { return !vm.fail(vm.themeStore(d)) })
	vm.ThemePane.Set(true)
	vm.PickTheme(d.Meta.Name)
	msg := T(KStatusThemeOpened, base+".css", d.Meta.Name)
	if len(notes) > 0 {
		msg += " " + strings.Join(notes, " ")
	}
	vm.notice(msg)
}

// themeImportRefused shows a refusal where the library's diagnostics are, and in
// the status line.
func (vm *ViewModel) themeImportRefused(err error) {
	vm.themeDiag = []string{T(KProblemError) + " " + ErrText(err)}
	vm.showProblems()
	vm.fail(err)
}

// ThemeAddFontFile adds a font file to the theme.
func (vm *ViewModel) ThemeAddFontFile(name string, data []byte) {
	base := path.Base(strings.ReplaceAll(name, `\`, "/"))
	vm.editTheme("font:"+base, T(KStepThemeFont, base), func(d *ThemeDoc) error {
		_, err := d.AddFont(base, data, "", "", "")
		return themeErr(err)
	})
}

// ThemeSetLogoFile sets the logo from an image file.
func (vm *ViewModel) ThemeSetLogoFile(name string, data []byte) {
	base := path.Base(strings.ReplaceAll(name, `\`, "/"))
	vm.editTheme("logo", T(KStepThemeLogo), func(d *ThemeDoc) error {
		return themeErr(d.SetLogo(base, data))
	})
}

// RemoveThemeLogo takes the logo out.
func (vm *ViewModel) RemoveThemeLogo() {
	vm.editTheme("logo", T(KStepThemeLogo), func(d *ThemeDoc) error {
		d.RemoveLogo()
		return nil
	})
}

// themeErr says a library refusal as the status line does, and keeps it in
// Problems.
func themeErr(err error) error {
	if err == nil {
		return nil
	}
	return themeRefused(err)
}

// OpenThemeSheet shows the theme's sheet in the editor: the CSS view. Edits
// there are read back by the controls, inside the block, and outside it they
// are the person's.
func (vm *ViewModel) OpenThemeSheet() {
	d, ok := vm.themeDoc()
	if !ok {
		return
	}
	if !vm.hasThemeSource(d.Meta.Name) { // a loose theme: its source is made first
		vm.projectStep(T(KStepThemeNew, d.Meta.Name), func() bool { return !vm.fail(vm.themeStore(d)) })
	}
	vm.OpenFile(ThemeSrcDir(d.Meta.Name) + "/" + d.CSSPath())
}

// ExportThemePackage downloads the theme as a .gsta.xz built by the library and
// read back by it, so that `slides theme verify` agrees. Nothing is uploaded.
func (vm *ViewModel) ExportThemePackage() {
	d, ok := vm.themeDoc()
	if !ok {
		return
	}
	data, err := d.Pack()
	if err != nil {
		vm.themeProblem(err)
		vm.notice(T(KWarn, ErrText(themeRefused(err))))
		return
	}
	vm.themeProblem(nil)
	dl, can := vm.host.(BytesDownloader)
	if !can {
		vm.notice(T(KStatusCannotSaveFile))
		return
	}
	file := d.Meta.Name + "-" + d.Meta.Version + ThemeExt
	dl.DownloadBytes(file, "application/x-xz", data)
	vm.notice(T(KStatusThemeExported, file))
}

// ExportThemeCSS downloads the theme as one style sheet, fonts and images
// inlined.
func (vm *ViewModel) ExportThemeCSS() {
	d, ok := vm.themeDoc()
	if !ok {
		return
	}
	css, err := d.PlainCSS()
	if err != nil {
		vm.themeProblem(err)
		vm.notice(T(KWarn, ErrText(themeRefused(err))))
		return
	}
	vm.themeProblem(nil)
	dl, can := vm.host.(BytesDownloader)
	if !can {
		vm.notice(T(KStatusCannotSaveFile))
		return
	}
	file := d.Meta.Name + ".css"
	dl.DownloadBytes(file, "text/css", []byte(css))
	vm.notice(T(KStatusThemeCSSExported, file))
}

// SaveThemeVersion keeps the project as a named version, called after the theme
// and its version.
func (vm *ViewModel) SaveThemeVersion() {
	d, ok := vm.themeDoc()
	if !ok {
		return
	}
	label := d.Meta.Name + " " + d.Meta.Version
	vm.keepVersion(vm.captureState(), versions.Named, label, func(versions.Entry) {
		vm.notice(T(KStatusThemeSaved, label))
	})
}

// ThemeValues are the values the pane shows, read from the sheet; the state is
// the block's.
func (vm *ViewModel) ThemeValues() (Visual, SyncState, bool) {
	d, ok := vm.themeDoc()
	if !ok {
		return nil, BlockAbsent, false
	}
	v, st := d.Visual()
	return v, st, true
}

// ThemeMetaValues are the fields of theme.hcl the pane shows, by field.
func (vm *ViewModel) ThemeMetaValues() (map[string]string, []themepack.Font, bool) {
	d, ok := vm.themeDoc()
	if !ok {
		return nil, nil, false
	}
	m := d.Meta
	return map[string]string{
		"name": m.Name, "version": m.Version, "description": m.Description,
		"authors": strings.Join(m.Authors, "; "), "license": m.License, "homepage": m.Homepage,
		"extends": m.Extends, "min_slides": m.MinSlides,
	}, m.Fonts, true
}

// ThemeFamilies are the families the font controls offer: the bundled ones and
// the ones the theme's own fonts bring.
func (vm *ViewModel) ThemeFamilies() []string {
	out := slices.Clone(BundledFamilies)
	if d, ok := vm.themeDoc(); ok {
		out = append(out, d.FontFamilies()...)
	}
	return out
}

// ThemeSheet is the compiled sheet of the edited theme as the preview reads it
// ("" without one), for a test harness to compare with what is exported.
func (vm *ViewModel) ThemeSheet() string {
	name := vm.ThemeName.Get()
	if name == "" {
		return ""
	}
	return vm.project().Text(vm.themeSheetPath(name))
}

// ThemeLogoName is the file name of the theme's logo ("" without one).
func (vm *ViewModel) ThemeLogoName() string {
	if d, ok := vm.themeDoc(); ok {
		for _, a := range d.Meta.Assets {
			if a.ID == "logo" {
				return path.Base(a.File)
			}
		}
	}
	return ""
}
