package playground

import (
	"github.com/go-widgets/mvvmtk"
	"github.com/go-widgets/toolkit"
)

// The Versions pane (a column of the splitter, right of the help) and the
// comparison, which takes the editor's place while it is open: a
// toolkit.DiffView of the file picked, side by side or inline.

// versionsWidth and its range, in logical pixels.
const versionsWidth, versionsMin, versionsMax = 300, 220, 600

// versionsPanel is the pane's widgets.
type versionsPanel struct {
	box                                             *toolkit.VBox
	title, note                                     *toolkit.Label
	list                                            *toolkit.ListBox
	onlyNamed                                       *toolkit.CheckButton
	preview, compare, restore, delete, back, saveAs *toolkit.Button
}

// comparePanel is the comparison's widgets.
type comparePanel struct {
	box    *toolkit.VBox
	title  *toolkit.Label
	file   *toolkit.DropDown
	layout *toolkit.CheckButton
	back   *toolkit.Button
	diff   *toolkit.DiffView
	note   *toolkit.Label
}

// newVersionsPanel builds the pane over the view-model.
func (s *Scene) newVersionsPanel() *versionsPanel {
	vm, inv, sc := s.vm, s.invalidate, toolkit.Scaled
	p := &versionsPanel{
		title:     toolkit.NewLabel(T(KLblVersions)).SetFontSize(sc(baseFontPx + 2)),
		note:      toolkit.NewLabel(vm.VersionsNote()),
		list:      toolkit.NewListBox(vm.VersionItems.Get()),
		onlyNamed: toolkit.NewCheckButton(T(KLblOnlyNamed), vm.OnlyNamed.Get()),
		saveAs:    toolkit.NewButton(T(KLblSaveVersion), nil),
		preview:   toolkit.NewButton(T(KLblVersionPreview), nil),
		compare:   toolkit.NewButton(T(KLblVersionCompare), nil),
		restore:   toolkit.NewButton(T(KLblVersionRestore), nil),
		delete:    toolkit.NewButton(T(KLblVersionDelete), nil),
		back:      toolkit.NewButton(T(KLblBackToNow), nil),
	}
	if f := uiFont(sc(baseFontPx + 2)); f != nil {
		p.title.FontSize, p.title.Font = 0, f
	}
	p.note.Ellipsis = true
	s.unbind = append(s.unbind,
		mvvmtk.BindChecked(p.onlyNamed, vm.OnlyNamed, inv),
		mvvmtk.BindListSelection(p.list, vm.VersionSelected, inv),
		mvvmtk.BindCommand(p.saveAs, vm.SaveVersion, inv),
		mvvmtk.BindCommand(p.preview, vm.PreviewVersion, inv),
		mvvmtk.BindCommand(p.compare, vm.CompareVersion, inv),
		mvvmtk.BindCommand(p.restore, vm.RestoreVersion, inv),
		mvvmtk.BindCommand(p.delete, vm.DeleteVersion, inv),
		mvvmtk.BindCommand(p.back, vm.BackToNow, inv),
		vm.VersionItems.Subscribe(func(items []string) {
			p.list.Items = items
			p.note.Text().Set(vm.VersionsNote())
			inv()
		}),
	)
	row1 := toolkit.NewHBox()
	row1.Spacing = sc(6)
	row1.AddFlex(p.preview, 1)
	row1.AddFlex(p.compare, 2)
	row2 := toolkit.NewHBox()
	row2.Spacing = sc(6)
	row2.AddFlex(p.restore, 1)
	row2.AddFlex(p.delete, 1)
	p.box = toolkit.NewVBox()
	p.box.Spacing = sc(6)
	p.box.AddFixed(p.title, sc(row))
	p.box.AddFixed(p.saveAs, sc(row-4))
	p.box.AddFixed(p.onlyNamed, sc(row-8))
	p.box.AddFixed(p.note, sc(row-10))
	p.box.AddFlex(p.list, 1)
	p.box.AddFixed(row1, sc(row-4))
	p.box.AddFixed(row2, sc(row-4))
	p.box.AddFixed(p.back, sc(row-4))
	return p
}

// newComparePanel builds the comparison over the view-model.
func (s *Scene) newComparePanel() *comparePanel {
	vm, inv, sc := s.vm, s.invalidate, toolkit.Scaled
	c := &comparePanel{
		title:  toolkit.NewLabel(""),
		file:   toolkit.NewDropDown(nil, 0),
		layout: toolkit.NewCheckButton(T(KLblSideBySide), vm.CompareSideBySide.Get()),
		back:   toolkit.NewButton(T(KLblBackToNow), nil),
		diff:   toolkit.NewDiffView("", ""),
		note:   toolkit.NewLabel(""),
	}
	c.title.Ellipsis = true
	c.diff.SetLayout(vm.CompareSideBySide.Get())
	if s.editorFont != nil {
		c.diff.SetFont(s.editorFont)
	}
	show := func() {
		cmp := vm.Compare.Get()
		if cmp == nil {
			s.leftStack.SetVisible("editor")
			inv()
			return
		}
		c.title.Text().Set(cmp.Title)
		c.file.Options = vm.CompareItems()
		i := vm.CompareFile.Get()
		c.note.Text().Set("")
		switch {
		case len(cmp.Files) == 0:
			c.note.Text().Set(T(KCompareSame))
			c.diff.SetTexts("", "")
		case i >= 0 && i < len(cmp.Files):
			f := cmp.Files[i]
			c.diff.OldTitle, c.diff.NewTitle = cmp.Name, T(KCompareNow)
			if f.Binary {
				c.note.Text().Set(c.file.Options[i])
				c.diff.SetTexts("", "")
			} else {
				c.diff.SetTexts(f.Old, f.New)
				c.diff.NextChange(1)
			}
		}
		s.leftStack.SetVisible("compare")
		inv()
	}
	s.unbind = append(s.unbind,
		mvvmtk.BindSelectedIndex(c.file, vm.CompareFile, inv),
		mvvmtk.BindChecked(c.layout, vm.CompareSideBySide, inv),
		mvvmtk.BindCommand(c.back, vm.BackToNow, inv),
		vm.Compare.Subscribe(func(*Comparison) { show() }),
		vm.CompareFile.Subscribe(func(int) { show() }),
		vm.CompareSideBySide.Subscribe(func(b bool) { c.diff.SetLayout(b); inv() }),
	)
	// two rows, so that a narrow column keeps every control whole: what is
	// compared and Back to now; the file and the layout
	head := toolkit.NewHBox()
	head.Spacing = sc(8)
	head.AddFlex(c.title, 1)
	head.AddFixed(c.back, toolkit.TextWidth(T(KLblBackToNow))+sc(28))
	pick := toolkit.NewHBox()
	pick.Spacing = sc(10)
	pick.AddFlex(c.file, 1)
	pick.AddFixed(c.layout, toolkit.TextWidth(T(KLblSideBySide))+sc(36))
	c.box = toolkit.NewVBox()
	c.box.Spacing = sc(6)
	c.box.AddFixed(head, sc(row-4))
	c.box.AddFixed(pick, sc(row-4))
	c.box.AddFixed(c.note, sc(row-10))
	c.box.AddFlex(c.diff, 1)
	return c
}

// versionRects are the named rectangles of the pane and the comparison, for
// a test harness.
func (s *Scene) versionRects(r map[string]toolkit.Rect) {
	if s.versions == nil {
		return
	}
	if s.vm.VersionsPane.Get() {
		p := s.versions
		for k, w := range map[string]toolkit.Widget{
			"versions-list": p.list, "versions-only-named": p.onlyNamed, "versions-save": p.saveAs,
			"version-preview": p.preview, "version-compare": p.compare, "version-restore": p.restore,
			"version-delete": p.delete, "version-back": p.back,
		} {
			r[k] = w.Bounds()
		}
	}
	if s.vm.Compare.Get() != nil {
		c := s.compare
		r["compare-diff"], r["compare-file"], r["compare-layout"], r["compare-back"] = c.diff.Bounds(), c.file.Bounds(), c.layout.Bounds(), c.back.Bounds()
	}
}
