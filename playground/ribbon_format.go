package playground

import (
	"github.com/go-widgets/toolkit"
)

// The Format tab: what Markdown can say about text, paragraphs, tables, the
// alignment of a slide and links, as controls over the view-model's Formatter. A
// toggle shows the state at the caret (Bold is pressed inside **…**) and presses
// the same command a keyboard shortcut does.

// formatLabels are the labels of the Format controls, by the name a test harness
// finds them by (Scene.Rects).
// Rebuilt by SetLocale from the catalogue (fmtLabelKey).
var formatLabels = formatTexts(fmtLabelKey)

// fmtLabelKey and fmtTipKey are the messages of a Format control, by its name.
func fmtLabelKey(name string) Key { return Key("fmt.label." + name) }
func fmtTipKey(name string) Key   { return Key("fmt.tip." + name) }

// formatTexts reads one message per Format control in the current language.
func formatTexts(key func(string) Key) map[string]string {
	m := make(map[string]string, len(formatGlyphs))
	for name := range formatGlyphs {
		m[name] = T(key(name))
	}
	return m
}

// formatGlyphs are the Iconoir glyphs of the controls ("" for a text-only one).
var formatGlyphs = map[string]string{
	"bold": "bold", "italic": "italic", "underline": "underline", "strike": "strikethrough", "code": "code",
	"h1": "", "h2": "", "h3": "", "bullets": "list", "numbered": "numbered-list-left", "quote": "quote", "rule": "minus",
	"table": "table", "table-format": "wrench",
	"row-above": "table-rows", "row-below": "table-rows", "row-delete": "trash",
	"col-left": "table-2-columns", "col-right": "table-2-columns", "col-delete": "trash",
	"colalign-left": "align-left", "colalign-center": "align-center", "colalign-right": "align-right",
	"slide-left": "align-left", "slide-center": "align-center", "slide-right": "align-right",
	"link": "link",
}

// formatTips say what a control does, and what it writes where that is not obvious.
var formatTips = formatTexts(fmtTipKey)

// formatAction is the Action of a Format control: its label, glyph, tip's
// shortcut, over the view-model's command.
func formatAction(vm *ViewModel, name string) *toolkit.Action {
	a := toolkit.NewCommandAction(name, formatLabels[name], vm.Format.Cmd[name])
	if g := formatGlyphs[name]; g != "" {
		a.Icon = icon(g)
	}
	if s := ShortcutText(name); s != "" {
		if ch, err := toolkit.ParseChord(s); err == nil {
			a.Shortcut = ch
		}
	}
	return a
}

// formatToggle is a toggle that shows the state at the caret.
func formatToggle(vm *ViewModel, name string, large bool) *toolkit.RibbonControl {
	return &toolkit.RibbonControl{Kind: toolkit.RibbonToggle, Large: large, Action: formatAction(vm, name),
		Checked: vm.Format.State[name], Tip: formatTips[name]}
}

// formatButton is a button for a command with no state.
func formatButton(vm *ViewModel, name string, large bool) *toolkit.RibbonControl {
	return &toolkit.RibbonControl{Kind: toolkit.RibbonButton, Large: large, Action: formatAction(vm, name), Tip: formatTips[name]}
}

// The Table control's grid: columns by rows (the header is a row), up to 8 by 8; a
// larger table is made by adding rows and columns.
const tableGridCols, tableGridRows = 8, 8

// formatTab builds the Format tab.
func formatTab(vm *ViewModel) *toolkit.RibbonTab {
	f := vm.Format
	// the Table control opens a grid: the cells light from the corner to the one the
	// pointer is on, "3 × 4" says how big the table will be
	table := &toolkit.RibbonControl{Kind: toolkit.RibbonGrid, Large: true, Action: formatAction(vm, "table"),
		GridCols: tableGridCols, GridRows: tableGridRows, OnGrid: f.InsertTable, Tip: formatTips["table"]}
	return &toolkit.RibbonTab{Label: ribbonTabs[2], Groups: []*toolkit.RibbonGroup{
		group(KGrpText, "text", 7,
			formatToggle(vm, "bold", false), formatToggle(vm, "italic", false), formatToggle(vm, "underline", false),
			formatToggle(vm, "strike", false), formatToggle(vm, "code", false)),
		group(KGrpParagraph, "list", 6,
			formatToggle(vm, "h1", false), formatToggle(vm, "h2", false), formatToggle(vm, "h3", false),
			formatToggle(vm, "bullets", false), formatToggle(vm, "numbered", false), formatToggle(vm, "quote", false),
			formatButton(vm, "rule", false)),
		group(KGrpTable, "table", 5, table, formatButton(vm, "table-format", false)),
		group(KGrpRowsCols, "table-rows", 4,
			formatButton(vm, "row-above", false), formatButton(vm, "row-below", false), formatButton(vm, "row-delete", false),
			formatButton(vm, "col-left", false), formatButton(vm, "col-right", false), formatButton(vm, "col-delete", false)),
		group(KGrpColumn, "align-center", 3,
			formatToggle(vm, "colalign-left", false), formatToggle(vm, "colalign-center", false), formatToggle(vm, "colalign-right", false)),
		group(KGrpSlideText, "align-left", 2,
			formatToggle(vm, "slide-left", false), formatToggle(vm, "slide-center", false), formatToggle(vm, "slide-right", false)),
		group(KGrpLink, "link", 1, formatToggle(vm, "link", true)),
	}}
}
