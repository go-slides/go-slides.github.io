package playground

import (
	"slices"
	"strconv"
	"strings"

	"github.com/go-widgets/mvvmtk"
	"github.com/go-widgets/toolkit"
)

// The Theme pane (themevm.go): a column of the splitter, right of the GitHub
// pane, that edits a theme of the project with go-widgets' Inspector (colour,
// font, number with a unit, enumeration and text rows, each by the editor that
// cannot produce an invalid value), ContrastBadge (WCAG, live, for every pair of
// text and ground) and the buttons of the theme: new, import, export, save a
// version, add a font, set the logo, and the CSS view (the editor on the theme's
// sheet). It holds no logic: every edit is a view-model call, and the pane reads
// the values back when the view-model says they changed (ThemeRev).

// contrastPairs are the pairs of ink and ground the pane checks, by property id;
// the headings are large text (3:1 is AA for them).
var contrastPairs = []struct {
	fg, bg string
	large  bool
}{
	{"text", "background", false},
	{"heading", "background", true},
	{"link", "background", false},
	{"accent", "background", false},
}

// themePair is one row of the contrast section.
type themePair struct {
	row   *toolkit.HBox
	name  *toolkit.Label
	stack *toolkit.Stack // "badge" or "need"
	badge *toolkit.ContrastBadge
	need  *toolkit.Label
}

// themePanel is the pane's widgets.
type themePanel struct {
	box                                     *toolkit.VBox
	title, contrastTitle                    *toolkit.Label
	pick                                    *toolkit.DropDown
	newB, importB, pkgB, cssB, saveB        *toolkit.Button
	fontB, logoB, noLogoB, sheetB, restoreB *toolkit.Button
	note                                    *toolkit.TextView
	scroll                                  *toolkit.ScrollView
	insp                                    *toolkit.Inspector
	pairs                                   []*themePair
	sig                                     string // what the inspector was built for
	picking                                 bool   // the drop-down is being moved to follow the view-model
	optValOf                                map[string]string
}

// themeRowIDs are the ids of the metadata rows (the properties' are theirs).
const themeMetaPrefix = "m:"

// newThemePanel builds the pane over the view-model.
func (s *Scene) newThemePanel() *themePanel {
	vm, inv, sc := s.vm, s.invalidate, toolkit.Scaled
	p := &themePanel{
		title:         toolkit.NewLabel(T(KLblThemeEditor)).SetFontSize(sc(baseFontPx + 2)),
		contrastTitle: toolkit.NewLabel(T(KThemeSecContrast)),
		pick:          toolkit.NewDropDown(nil, 0),
		newB:          toolkit.NewButton(T(KThemeBtnNew), nil),
		importB:       toolkit.NewButton(T(KThemeBtnImport), nil),
		pkgB:          toolkit.NewButton(T(KThemeBtnPackage), nil),
		cssB:          toolkit.NewButton(T(KThemeBtnCSS), nil),
		saveB:         toolkit.NewButton(T(KThemeBtnVersion), nil),
		fontB:         toolkit.NewButton(T(KThemeBtnFont), nil),
		logoB:         toolkit.NewButton(T(KThemeBtnLogo), nil),
		noLogoB:       toolkit.NewButton(T(KThemeBtnNoLogo), nil),
		sheetB:        toolkit.NewButton(T(KThemeBtnOpenCSS), nil),
		restoreB:      toolkit.NewButton(T(KThemeBtnRestore), nil),
		note:          toolkit.NewTextView(""),
		insp:          toolkit.NewInspector(nil),
	}
	if f := uiFont(sc(baseFontPx + 2)); f != nil {
		p.title.FontSize, p.title.Font = 0, f
	}
	p.scroll = toolkit.NewScrollView(p.insp)
	for _, cp := range contrastPairs {
		pr := &themePair{
			name:  toolkit.NewLabel(T(KThemeContrastPair, themePropLabel(cp.fg), themePropLabel(cp.bg))),
			stack: toolkit.NewStack(),
			badge: toolkit.NewContrastBadge(toolkit.RGBA{}, toolkit.RGBA{}),
			need:  toolkit.NewLabel(T(KThemeContrastNeed)),
		}
		pr.name.Ellipsis, pr.need.Ellipsis = true, true
		pr.badge.FailText = T(KThemeContrastFail)
		pr.badge.Large().Set(cp.large)
		pr.badge.SetBounds(toolkit.Rect{W: sc(150), H: sc(row - 12)})
		pr.stack.AddPage("badge", pr.badge)
		pr.stack.AddPage("need", pr.need)
		pr.stack.SetVisible("need")
		pr.row = toolkit.NewHBox()
		pr.row.Spacing = sc(6)
		pr.row.AddFlex(pr.name, 1)
		pr.row.AddFixed(pr.stack, sc(150))
		p.pairs = append(p.pairs, pr)
	}
	p.optValOf = map[string]string{}
	for v := range optionKeys {
		p.optValOf[ThemeOptionLabel(v)] = v
	}

	s.unbind = append(s.unbind,
		mvvmtk.BindCommand(p.newB, vm.ThemeNew, inv),
		mvvmtk.BindCommand(p.importB, vm.ThemeImport, inv),
		mvvmtk.BindCommand(p.pkgB, vm.ThemePackage, inv),
		mvvmtk.BindCommand(p.cssB, vm.ThemeCSS, inv),
		mvvmtk.BindCommand(p.saveB, vm.ThemeSave, inv),
		mvvmtk.BindCommand(p.fontB, vm.ThemeFont, inv),
		mvvmtk.BindCommand(p.logoB, vm.ThemeLogo, inv),
		mvvmtk.BindCommand(p.noLogoB, vm.ThemeNoLogo, inv),
		mvvmtk.BindCommand(p.sheetB, vm.ThemeOpenCSS, inv),
		mvvmtk.BindCommand(p.restoreB, vm.ThemeRestore, inv),
		vm.ThemeRev.Subscribe(func(int) { p.refresh(s); inv() }),
		vm.ThemeNote.Subscribe(func(string) { p.wrapNote(vm); inv() }),
		vm.ThemeWidth.Subscribe(func(int) { p.wrapNote(vm); inv() }),
		vm.ThemesVersion.Subscribe(func(int) { p.refreshPick(vm); inv() }),
		vm.ThemeName.Subscribe(func(string) { p.refreshPick(vm); inv() }),
		p.pick.Selected().Subscribe(func(i int) {
			if p.picking || i < 0 || i >= len(p.pick.Options) {
				return
			}
			vm.PickTheme(p.pick.Options[i])
		}),
	)

	pair := func(a, b toolkit.Widget) *toolkit.HBox {
		h := toolkit.NewHBox()
		h.Spacing = sc(6)
		h.AddFlex(a, 1)
		h.AddFlex(b, 1)
		return h
	}
	pickRow := toolkit.NewHBox()
	pickRow.Spacing = sc(6)
	pickRow.AddFlex(p.pick, 2)
	pickRow.AddFlex(p.newB, 1)
	p.box = toolkit.NewVBox()
	p.box.Spacing = sc(6)
	p.box.AddFixed(p.title, sc(row))
	p.box.AddFixed(pickRow, sc(row-4))
	p.box.AddFixed(pair(p.importB, p.sheetB), sc(row-4))
	p.box.AddFixed(pair(p.pkgB, p.cssB), sc(row-4))
	p.box.AddFixed(pair(p.fontB, p.logoB), sc(row-4))
	p.box.AddFixed(pair(p.noLogoB, p.saveB), sc(row-4))
	p.box.AddFixed(p.restoreB, sc(row-4))
	p.box.AddFixed(p.note, sc(78))
	p.box.AddFlex(p.scroll, 1)
	p.box.AddFixed(p.contrastTitle, sc(row-8))
	for _, pr := range p.pairs {
		p.box.AddFixed(pr.row, sc(row-10))
	}
	p.refreshPick(vm)
	p.wrapNote(vm)
	p.refresh(s)
	return p
}

// wrapNote wraps the note to the pane's width (a TextView does not wrap).
func (p *themePanel) wrapNote(vm *ViewModel) {
	cols := vm.ThemeWidth.Get() * 100 / max(baseFontPx*55, 1)
	p.note.SetText(Wrap(vm.ThemeNote.Get(), max(cols-2, 20)))
}

// refreshPick lists the themes the pane can edit and shows the one it does.
func (p *themePanel) refreshPick(vm *ViewModel) {
	p.picking = true
	defer func() { p.picking = false }()
	names := vm.EditableThemes()
	cur := vm.ThemeName.Get()
	if cur != "" && !slices.Contains(names, cur) {
		names = append(names, cur)
	}
	p.pick.Options = names
	if i := slices.Index(names, cur); i >= 0 {
		p.pick.Selected().Set(i)
	}
}

// metaChoices are the base themes a theme may name: the built-in ones and the
// project's, but not the theme itself; "none" is no base.
func (p *themePanel) metaChoices(vm *ViewModel) []string {
	out := []string{ThemeOptionLabel("none")}
	for _, t := range vm.Themes {
		if t != vm.ThemeName.Get() {
			out = append(out, t)
		}
	}
	return out
}

// rows are the inspector's rows for the theme being edited.
func (p *themePanel) rows(vm *ViewModel, fonts []string, licenses []string) []toolkit.InspectorRow {
	var out []toolkit.InspectorRow
	section := func(k Key) { out = append(out, toolkit.InspectorRow{Kind: toolkit.InspectorSection, Label: T(k)}) }
	groups := []struct {
		group string
		title Key
	}{{GroupPalette, KThemeSecPalette}, {GroupType, KThemeSecType}, {GroupLogo, KThemeSecLogo}, {GroupFooter, KThemeSecFooter}}
	for _, g := range groups {
		section(g.title)
		for _, pr := range ThemeProps {
			if pr.Group != g.group {
				continue
			}
			r := toolkit.InspectorRow{ID: pr.ID, Label: themePropLabel(pr.ID), Min: pr.Min, Max: pr.Max, Step: pr.Step, Units: pr.Units}
			switch pr.Kind {
			case PropColor:
				r.Kind = toolkit.InspectorColor
			case PropFont:
				r.Kind, r.Choices = toolkit.InspectorFont, fonts
			case PropNumber:
				r.Kind = toolkit.InspectorNumber
			case PropEnum:
				r.Kind = toolkit.InspectorEnum
				for _, c := range pr.Choices {
					r.Choices = append(r.Choices, ThemeOptionLabel(c))
				}
			}
			out = append(out, r)
		}
	}
	section(KThemeSecMeta)
	for _, f := range []string{"version", "description", "authors", "license", "homepage", "min_slides"} {
		out = append(out, toolkit.InspectorRow{ID: themeMetaPrefix + f, Label: themeMetaLabel(f), Kind: toolkit.InspectorText})
	}
	out = append(out, toolkit.InspectorRow{ID: themeMetaPrefix + "extends", Label: themeMetaLabel("extends"), Kind: toolkit.InspectorEnum, Choices: p.metaChoices(vm)})
	for _, fam := range licenses {
		out = append(out, toolkit.InspectorRow{ID: themeMetaPrefix + "font-license:" + fam, Label: themeMetaLabel("font-license:" + fam), Kind: toolkit.InspectorText})
	}
	return out
}

// refresh reads the theme back into the controls: the inspector is built again
// when what it offers changed (the families, the fonts' licences, the bases), and
// otherwise only the values that differ are set, so a control being used is not
// moved under the person.
func (p *themePanel) refresh(s *Scene) {
	vm := s.vm
	vals, _, ok := vm.ThemeValues()
	meta, fonts, ok2 := vm.ThemeMetaValues()
	var families, licenses []string
	if ok && ok2 {
		families = vm.ThemeFamilies()
		for _, f := range fonts {
			if !slices.Contains(licenses, f.Family) {
				licenses = append(licenses, f.Family)
			}
		}
	}
	sig := strings.Join(families, "\x00") + "\x01" + strings.Join(licenses, "\x00") + "\x01" + strings.Join(p.metaChoices(vm), "\x00") + "\x01" + strconv.FormatBool(ok && ok2)
	if sig != p.sig {
		p.sig = sig
		var rows []toolkit.InspectorRow
		if ok && ok2 {
			rows = p.rows(vm, families, licenses)
		}
		p.insp = toolkit.NewInspector(rows)
		p.insp.OnChange = func(id, v string) { p.edited(vm, id, v) }
		p.scroll.Child = p.insp
		p.scroll.SetBounds(p.scroll.Bounds())
	}
	if ok && ok2 {
		set := func(id, v string) {
			if p.insp.Value(id) != nil && p.insp.Get(id) != v {
				p.insp.Set(id, v)
			}
		}
		for _, pr := range ThemeProps {
			v := vals[pr.ID]
			if pr.Kind == PropEnum && v != "" {
				v = ThemeOptionLabel(v)
			}
			if pr.Kind == PropEnum && v == "" {
				continue // not set: the control shows its first choice, which is not a value
			}
			set(pr.ID, v)
		}
		for _, f := range []string{"version", "description", "authors", "license", "homepage", "min_slides"} {
			set(themeMetaPrefix+f, meta[f])
		}
		ext := meta["extends"]
		if ext == "" {
			ext = ThemeOptionLabel("none")
		}
		set(themeMetaPrefix+"extends", ext)
		for _, f := range fonts {
			set(themeMetaPrefix+"font-license:"+f.Family, f.License)
		}
	}
	p.contrast(vals)
	p.scroll.SetBounds(p.scroll.Bounds())
}

// edited is a change the person made in the inspector: a view-model call.
func (p *themePanel) edited(vm *ViewModel, id, v string) {
	if field, ok := strings.CutPrefix(id, themeMetaPrefix); ok {
		if field == "extends" && v == ThemeOptionLabel("none") {
			v = ""
		}
		vm.SetThemeMeta(field, v)
		return
	}
	if pr, ok := PropByID(id); ok && pr.Kind == PropEnum {
		if val, ok := p.optValOf[v]; ok {
			v = val
		}
	}
	vm.SetThemeValue(id, v)
}

// contrast shows, for each pair, its WCAG ratio and level, or asks for the two
// colours when one is not set (or is written as something the pane cannot read,
// like var(--x)).
func (p *themePanel) contrast(vals Visual) {
	for i, cp := range contrastPairs {
		fg, ok1 := toolkit.ParseCSSColor(vals[cp.fg])
		bg, ok2 := toolkit.ParseCSSColor(vals[cp.bg])
		pr := p.pairs[i]
		if ok1 && ok2 {
			pr.badge.Foreground().Set(fg)
			pr.badge.Background().Set(bg)
			pr.stack.SetVisible("badge")
		} else {
			pr.stack.SetVisible("need")
		}
	}
}

// ContrastText is what the contrast row i says, for a test harness.
func (s *Scene) ContrastText(i int) string {
	if s.themep == nil || i < 0 || i >= len(s.themep.pairs) {
		return ""
	}
	pr := s.themep.pairs[i]
	if pr.stack.Visible().Get() != "badge" {
		return pr.need.Text().Get()
	}
	return pr.badge.Text()
}

// themeRects name the pane's controls for a test harness.
func (s *Scene) themeRects(r map[string]toolkit.Rect) {
	if s.themep == nil || !s.vm.ThemePane.Get() {
		return
	}
	p := s.themep
	for k, w := range map[string]toolkit.Widget{
		"theme-pick": p.pick, "theme-new": p.newB, "theme-import": p.importB, "theme-package": p.pkgB, "theme-css": p.cssB,
		"theme-save": p.saveB, "theme-font": p.fontB, "theme-logo": p.logoB, "theme-nologo": p.noLogoB,
		"theme-sheet": p.sheetB, "theme-restore": p.restoreB, "theme-note": p.note, "theme-scroll": p.scroll,
	} {
		r[k] = w.Bounds()
	}
	for i, pr := range p.pairs {
		r["theme-contrast-"+strconv.Itoa(i)] = pr.row.Bounds()
	}
	for _, id := range slices.Concat(propIDs(), []string{themeMetaPrefix + "version", themeMetaPrefix + "description", themeMetaPrefix + "authors", themeMetaPrefix + "license", themeMetaPrefix + "extends"}) {
		if rc, ok := p.insp.EditorRect(id); ok {
			// the inspector is laid out as long as it is; the pane shows what the scroll
			// offset leaves, which is where a harness clicks
			rc.X -= p.scroll.OffsetX().Get()
			rc.Y -= p.scroll.OffsetY().Get()
			r["theme-row-"+id] = rc
		}
	}
}

func propIDs() []string {
	out := make([]string, len(ThemeProps))
	for i, pr := range ThemeProps {
		out[i] = pr.ID
	}
	return out
}
