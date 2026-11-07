package playground

import (
	"github.com/go-icons/iconoir"
	"github.com/go-widgets/mvvm"
	"github.com/go-widgets/toolkit"
)

// The playground's command ribbon: four tabs of go-widgets' Ribbon over the
// view-model's commands and observables. It holds no logic of its own: a button
// is an Action over an mvvm.Command (NewCommandAction), a drop-down or a toggle
// shares the observable the view-model already has, so the ribbon is one more
// view on the same state the editor and the preview follow. Its words are the
// catalogue's (i18n.go): no literal text in this file.

// icon is an Iconoir glyph (MIT, go-icons/iconoir) by name.
func icon(name string) toolkit.IconFunc { return toolkit.SVGIcon(iconoir.Icon(name)) }

// act is an Action over a view-model command, with its Iconoir glyph.
func act(id string, label Key, glyph string, c *mvvm.Command) *toolkit.Action {
	a := toolkit.NewCommandAction(id, T(label), c)
	a.Icon = icon(glyph)
	return a
}

// view is an Action with no body, for a control whose state is an observable
// (a toggle, a drop-down): it carries the label and the glyph.
func view(id string, label Key, glyph string) *toolkit.Action {
	a := toolkit.NewAction(id, T(label), nil)
	a.Icon = icon(glyph)
	return a
}

// ThemeOptions are the theme drop-down's entries, each saying what it is.
func ThemeOptions(themes []string) []string {
	out := make([]string, len(themes))
	for i, t := range themes {
		out[i] = T(KLblThemeOption, t)
	}
	return out
}

// RibbonLabels are the labels of the controls a test harness clicks, by the name
// Scene.Rects gives them. Rebuilt by SetLocale.
var RibbonLabels = ribbonLabels()

func ribbonLabels() map[string]string {
	m := map[string]string{
		"undo": T(KLblUndo), "redo": T(KLblRedo), "save-version": T(KLblSaveVersion), "versions": T(KLblVersions), "git": T(KLblGit),
		"examples": T(KLblExample), "reset": T(KLblReset), "theme": T(KLblTheme), "draft": T(KLblDraft),
		"share": T(KLblShare), "present": T(KLblPresent), "mode": T(KLblPreview),
		"minimapToggle": T(KLblMinimap), "thumbsToggle": T(KLblSlides), "filesToggle": T(KLblFiles), "helpToggle": T(KLblHelp),
		"newslide": T(KLblNewSlide), "image": T(KLblImage), "background": T(KLblBackground),
		"math": T(KLblMath), "code": T(KLblCode), "kind": T(KLblKind),
		"save-html": T(KLblHTML), "save-pdf": T(KLblPDF), "save-pptx": T(KLblPPTX), "save-odp": T(KLblODP),
	}
	// the Format tab's controls, named "fmt-bold", "fmt-table"…
	for name, label := range formatLabels {
		m["fmt-"+name] = label
	}
	return m
}

// ribbonTabs are the tab labels, in order. Rebuilt by SetLocale.
var ribbonTabs = tabLabels()

func tabLabels() []string {
	return []string{T(KTabHome), T(KTabInsert), T(KTabFormat), T(KTabView), T(KTabExport)}
}

// group is a ribbon group.
func group(label Key, glyph string, priority int, controls ...*toolkit.RibbonControl) *toolkit.RibbonGroup {
	return &toolkit.RibbonGroup{Label: T(label), Icon: icon(glyph), Priority: priority, Controls: controls}
}

// button is a ribbon button over an action, with its tip.
func button(a *toolkit.Action, large bool, tip Key) *toolkit.RibbonControl {
	return &toolkit.RibbonControl{Kind: toolkit.RibbonButton, Large: large, Action: a, Tip: T(tip)}
}

// toggle is a ribbon toggle over an observable.
func toggle(a *toolkit.Action, checked *mvvm.Observable[bool], tip Key) *toolkit.RibbonControl {
	return &toolkit.RibbonControl{Kind: toolkit.RibbonToggle, Action: a, Checked: checked, Tip: T(tip)}
}

// dropdown is a ribbon drop-down over a selection.
func dropdown(label Key, options []string, selected *mvvm.Observable[int], tip Key) *toolkit.RibbonControl {
	return &toolkit.RibbonControl{Kind: toolkit.RibbonDropdown, Label: T(label), Options: options, Selected: selected, Tip: T(tip)}
}

// newRibbon builds the ribbon over vm. inv is called when a control's list
// changes, so the scene repaints. The returned functions undo the subscriptions.
func newRibbon(vm *ViewModel, inv func()) (*toolkit.Ribbon, []func()) {
	var unbind []func()

	// Home: the deck, its look, sharing and presenting, the drafts.
	example := dropdown(KLblExample, ExampleNames, vm.Example, KTipExample)
	theme := dropdown(KLblTheme, ThemeOptions(vm.Themes), vm.Theme, KTipTheme)
	draft := dropdown(KLblDraft, vm.DraftItems(), vm.DraftAction, KTipDraft)
	unbind = append(unbind,
		vm.ThemesVersion.Subscribe(func(int) { theme.Options = ThemeOptions(vm.Themes); inv() }),
		vm.DraftVersion.Subscribe(func(int) { draft.Options = vm.DraftItems(); inv() }),
	)
	// History: Undo and Redo, their tips naming the project step they would take
	undoA, redoA := act("undo", KLblUndo, "undo", vm.Undo), act("redo", KLblRedo, "redo", vm.Redo)
	undoA.Shortcut, _ = toolkit.ParseChord(UndoShortcut)
	redoA.Shortcut, _ = toolkit.ParseChord(RedoShortcut)
	undoB, redoB := button(undoA, true, KTipUndo), button(redoA, false, KTipRedo)
	undoB.Tip, redoB.Tip = UndoTip(vm.UndoLabel.Get()), RedoTip(vm.RedoLabel.Get())
	unbind = append(unbind,
		vm.UndoLabel.Subscribe(func(l string) { undoB.Tip = UndoTip(l); inv() }),
		vm.RedoLabel.Subscribe(func(l string) { redoB.Tip = RedoTip(l); inv() }),
	)
	home := &toolkit.RibbonTab{Label: ribbonTabs[0], Groups: []*toolkit.RibbonGroup{
		group(KGrpHistory, "undo", 5, undoB, redoB,
			button(act("save-version", KLblSaveVersion, "bookmark", vm.SaveVersion), false, KTipSaveVersion),
			toggle(view("versions", KLblVersions, "clock-rotate-right"), vm.VersionsPane, KTipVersions),
			toggle(view("git", KLblGit, "git-fork"), vm.GitPane, KTipGit)),
		group(KGrpDeck, "folder", 4, example, button(act("reset", KLblReset, "refresh", vm.Reset), false, KTipReset)),
		group(KGrpLook, "palette", 2, theme),
		group(KGrpPresent, "presentation", 3,
			button(act("present", KLblPresent, "presentation", vm.Present), true, KTipPresent),
			button(act("share", KLblShare, "share-ios", vm.Share), true, KTipShare)),
		group(KGrpDraft, "floppy-disk", 1, draft),
	}}

	// Insert: edits of the deck's Markdown at the caret, or around the selection.
	kind := dropdown(KLblKind, vm.Insert.KindItems(), vm.Insert.Kind, KTipKind)
	unbind = append(unbind, vm.Insert.KindVersion.Subscribe(func(int) { kind.Options = vm.Insert.KindItems(); inv() }))
	bg := act("background", KLblBackground, "media-image-list", vm.Insert.Background)
	background := &toolkit.RibbonControl{Kind: toolkit.RibbonSplit, Large: true, Action: bg, Items: []toolkit.RibbonItem{
		{Action: act("bg-left", KLblBgLeft, "layout-left", vm.Insert.BgLeft)},
		{Action: act("bg-right", KLblBgRight, "layout-right", vm.Insert.BgRight)},
	}, Tip: T(KTipBackground)}
	insert := &toolkit.RibbonTab{Label: ribbonTabs[1], Groups: []*toolkit.RibbonGroup{
		group(KGrpSlides, "page-plus", 4,
			button(act("newslide", KLblNewSlide, "page-plus", vm.Insert.Slide), true, KTipNewSlide), kind),
		group(KGrpPictures, "media-image", 3,
			button(act("image", KLblImage, "media-image", vm.Insert.Image), true, KTipImage), background),
		group(KGrpText, "text", 2,
			button(act("math", KLblMath, "sigma-function", vm.Insert.Math), false, KTipMath),
			button(act("code", KLblCode, "code-brackets", vm.Insert.Code), false, KTipCode)),
	}}

	// View: what the page shows.
	mode := dropdown(KLblPreview, ModeNames, vm.Mode, KTipMode)
	viewTab := &toolkit.RibbonTab{Label: ribbonTabs[3], Groups: []*toolkit.RibbonGroup{
		group(KGrpPreview, "eye", 3, mode,
			button(act("present-view", KLblPresent, "presentation", vm.Present), false, KTipPresent)),
		group(KGrpEditor, "view-grid", 2,
			toggle(view("minimap", KLblMinimap, "map"), vm.Minimap, KTipMinimap),
			toggle(view("thumbs", KLblSlides, "view-grid"), vm.Thumbs, KTipThumbs),
			toggle(view("files", KLblFiles, "folder"), vm.FilesPane, KTipFiles),
			toggle(view("help", KLblHelp, "notes"), vm.HelpPane, KTipHelp)),
	}}

	// Export: the deck as a file.
	export := &toolkit.RibbonTab{Label: ribbonTabs[4], Groups: []*toolkit.RibbonGroup{
		group(KGrpSaveAs, "download", 1,
			button(act("save-html", KLblHTML, "code", vm.Download), true, KTipHTML),
			button(act("save-pdf", KLblPDF, "page", vm.Export["pdf"]), true, KTipPDF),
			button(act("save-pptx", KLblPPTX, "presentation", vm.Export["pptx"]), true, KTipPPTX),
			button(act("save-odp", KLblODP, "page-edit", vm.Export["odp"]), true, KTipODP)),
	}}

	r := toolkit.NewRibbon(home, insert, formatTab(vm), viewTab, export)
	r.TooltipDelayMs = 450
	return r, unbind
}
