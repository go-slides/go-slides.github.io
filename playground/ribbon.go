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
// view on the same state the editor and the preview follow.

// icon is an Iconoir glyph (MIT, go-icons/iconoir) by name.
func icon(name string) toolkit.IconFunc { return toolkit.SVGIcon(iconoir.Icon(name)) }

// act is an Action over a view-model command, with its Iconoir glyph.
func act(id, label, glyph string, c *mvvm.Command) *toolkit.Action {
	a := toolkit.NewCommandAction(id, label, c)
	a.Icon = icon(glyph)
	return a
}

// view is an Action with no body, for a control whose state is an observable
// (a toggle, a drop-down): it carries the label and the glyph.
func view(id, label, glyph string) *toolkit.Action {
	a := toolkit.NewAction(id, label, nil)
	a.Icon = icon(glyph)
	return a
}

// ThemeOptions are the theme drop-down's entries, each saying what it is.
func ThemeOptions(themes []string) []string {
	out := make([]string, len(themes))
	for i, t := range themes {
		out[i] = "Theme: " + t
	}
	return out
}

// RibbonLabels are the labels of the controls a test harness clicks, by the name
// Scene.Rects gives them.
var RibbonLabels = map[string]string{
	"examples": "Example", "reset": "Reset", "theme": "Theme", "draft": "Draft",
	"share": "Share link", "present": "Present", "mode": "Preview",
	"minimapToggle": "Minimap", "thumbsToggle": "Slides", "filesToggle": "Files", "helpToggle": "Help",
	"newslide": "New slide", "image": "Image", "background": "Background",
	"math": "Math", "code": "Code block", "kind": "Kind",
	"save-html": "HTML page", "save-pdf": "PDF", "save-pptx": "PowerPoint", "save-odp": "OpenDocument",
}

func init() {
	// the Format tab's controls, named "fmt-bold", "fmt-table"…
	for name, label := range formatLabels {
		RibbonLabels["fmt-"+name] = label
	}
}

// ribbonTabs are the tab labels, in order.
var ribbonTabs = []string{"Home", "Insert", "Format", "View", "Export"}

// newRibbon builds the ribbon over vm. inv is called when a control's list
// changes, so the scene repaints. The returned functions undo the subscriptions.
func newRibbon(vm *ViewModel, inv func()) (*toolkit.Ribbon, []func()) {
	var unbind []func()

	// Home: the deck, its look, sharing and presenting, the drafts.
	example := &toolkit.RibbonControl{Kind: toolkit.RibbonDropdown, Label: "Example", Options: ExampleNames, Selected: vm.Example,
		Tip: "Start from one of the examples. Your text stays in this browser: Draft ▸ Restore replaced brings it back."}
	theme := &toolkit.RibbonControl{Kind: toolkit.RibbonDropdown, Label: "Theme", Options: ThemeOptions(vm.Themes), Selected: vm.Theme,
		Tip: "The deck's theme: choosing one writes it into the deck's front matter."}
	draft := &toolkit.RibbonControl{Kind: toolkit.RibbonDropdown, Label: "Draft", Options: vm.DraftItems(), Selected: vm.DraftAction,
		Tip: "Restore, forget or reset the deck this browser keeps for you."}
	unbind = append(unbind,
		vm.ThemesVersion.Subscribe(func(int) { theme.Options = ThemeOptions(vm.Themes); inv() }),
		vm.DraftVersion.Subscribe(func(int) { draft.Options = vm.DraftItems(); inv() }),
	)
	home := &toolkit.RibbonTab{Label: "Home", Groups: []*toolkit.RibbonGroup{
		{Label: "Deck", Icon: icon("folder"), Priority: 4, Controls: []*toolkit.RibbonControl{
			example,
			{Kind: toolkit.RibbonButton, Action: act("reset", "Reset", "refresh", vm.Reset), Tip: "Put the example's text back; your deck is kept as a draft."},
		}},
		{Label: "Look", Icon: icon("palette"), Priority: 2, Controls: []*toolkit.RibbonControl{theme}},
		{Label: "Present", Icon: icon("presentation"), Priority: 3, Controls: []*toolkit.RibbonControl{
			{Kind: toolkit.RibbonButton, Large: true, Action: act("present", "Present", "presentation", vm.Present),
				Tip: "Show the slideshow here and open the presenter view in its own window."},
			{Kind: toolkit.RibbonButton, Large: true, Action: act("share", "Share link", "share-ios", vm.Share),
				Tip: "Copy a link that carries the deck itself: nothing is uploaded."},
		}},
		{Label: "Draft", Icon: icon("floppy-disk"), Priority: 1, Controls: []*toolkit.RibbonControl{draft}},
	}}

	// Insert: edits of the deck's Markdown at the caret, or around the selection.
	kind := &toolkit.RibbonControl{Kind: toolkit.RibbonDropdown, Label: "Kind", Options: vm.Insert.KindItems(), Selected: vm.Insert.Kind,
		Tip: "Give the slide a kind: a template with its own look, set by a <!-- _kind: … --> comment."}
	unbind = append(unbind, vm.Insert.KindVersion.Subscribe(func(int) { kind.Options = vm.Insert.KindItems(); inv() }))
	bg := act("background", "Background", "media-image-list", vm.Insert.Background)
	insert := &toolkit.RibbonTab{Label: "Insert", Groups: []*toolkit.RibbonGroup{
		{Label: "Slides", Icon: icon("page-plus"), Priority: 4, Controls: []*toolkit.RibbonControl{
			{Kind: toolkit.RibbonButton, Large: true, Action: act("newslide", "New slide", "page-plus", vm.Insert.Slide),
				Tip: "Add an empty slide after the current one."},
			kind,
		}},
		{Label: "Pictures", Icon: icon("media-image"), Priority: 3, Controls: []*toolkit.RibbonControl{
			{Kind: toolkit.RibbonButton, Large: true, Action: act("image", "Image", "media-image", vm.Insert.Image),
				Tip: "Choose an image file: it goes into the deck as a data: URI, nothing is uploaded. Selected text becomes its description."},
			{Kind: toolkit.RibbonSplit, Large: true, Action: bg, Items: []toolkit.RibbonItem{
				{Action: act("bg-left", "Left split background", "layout-left", vm.Insert.BgLeft)},
				{Action: act("bg-right", "Right split background", "layout-right", vm.Insert.BgRight)},
			}, Tip: "Choose an image for the slide's background; the arrow fills only the left or the right 40%."},
		}},
		{Label: "Text", Icon: icon("text"), Priority: 2, Controls: []*toolkit.RibbonControl{
			{Kind: toolkit.RibbonButton, Action: act("math", "Math", "sigma-function", vm.Insert.Math),
				Tip: "A display formula between $$ lines; the selection goes inside it."},
			{Kind: toolkit.RibbonButton, Action: act("code", "Code block", "code-brackets", vm.Insert.Code),
				Tip: "A fenced code block with line numbers ({lineNumbers}); the selection goes inside it."},
		}},
	}}

	// View: what the page shows.
	mode := &toolkit.RibbonControl{Kind: toolkit.RibbonDropdown, Label: "Preview", Options: ModeNames, Selected: vm.Mode,
		Tip: "Overview stacks every slide and scrolls with the editor; Slideshow is the presentation itself."}
	viewTab := &toolkit.RibbonTab{Label: "View", Groups: []*toolkit.RibbonGroup{
		{Label: "Preview", Icon: icon("eye"), Priority: 3, Controls: []*toolkit.RibbonControl{
			mode,
			{Kind: toolkit.RibbonButton, Action: act("present-view", "Present", "presentation", vm.Present),
				Tip: "Show the slideshow here and open the presenter view in its own window."},
		}},
		{Label: "Editor", Icon: icon("view-grid"), Priority: 2, Controls: []*toolkit.RibbonControl{
			{Kind: toolkit.RibbonToggle, Action: view("minimap", "Minimap", "map"), Checked: vm.Minimap,
				Tip: "A map of the deck beside the editor's scrollbar."},
			{Kind: toolkit.RibbonToggle, Action: view("thumbs", "Slides", "view-grid"), Checked: vm.Thumbs,
				Tip: "A column of slide thumbnails; click one to go to it."},
			{Kind: toolkit.RibbonToggle, Action: view("files", "Files", "folder"), Checked: vm.FilesPane,
				Tip: "The project's files: themes, kinds, data and media. A narrow window starts without it."},
			{Kind: toolkit.RibbonToggle, Action: view("help", "Help", "notes"), Checked: vm.HelpPane,
				Tip: "The help for what the caret is on, and the documentation links. A narrow window starts without it."},
		}},
	}}

	// Export: the deck as a file.
	export := &toolkit.RibbonTab{Label: "Export", Groups: []*toolkit.RibbonGroup{
		{Label: "Save as", Icon: icon("download"), Priority: 1, Controls: []*toolkit.RibbonControl{
			{Kind: toolkit.RibbonButton, Large: true, Action: act("save-html", "HTML page", "code", vm.Download),
				Tip: "One HTML file that presents the deck in any browser."},
			{Kind: toolkit.RibbonButton, Large: true, Action: act("save-pdf", "PDF", "page", vm.Export["pdf"]),
				Tip: "A PDF, one page per slide, made here in your browser."},
			{Kind: toolkit.RibbonButton, Large: true, Action: act("save-pptx", "PowerPoint", "presentation", vm.Export["pptx"]),
				Tip: "An editable PowerPoint file, made here in your browser."},
			{Kind: toolkit.RibbonButton, Large: true, Action: act("save-odp", "OpenDocument", "page-edit", vm.Export["odp"]),
				Tip: "An editable OpenDocument presentation, made here in your browser."},
		}},
	}}

	r := toolkit.NewRibbon(home, insert, formatTab(vm), viewTab, export)
	r.TooltipDelayMs = 450
	return r, unbind
}
