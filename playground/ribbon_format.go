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
var formatLabels = map[string]string{
	"bold": "Bold", "italic": "Italic", "underline": "Underline", "strike": "Strikethrough", "code": "Inline code",
	"h1": "Heading 1", "h2": "Heading 2", "h3": "Heading 3", "bullets": "Bullets", "numbered": "Numbering",
	"quote": "Quote", "rule": "Rule",
	"table": "Table", "table-format": "Tidy",
	"row-above": "Row above", "row-below": "Row below", "row-delete": "Delete row",
	"col-left": "Column left", "col-right": "Column right", "col-delete": "Delete column",
	"colalign-left": "Column left-aligned", "colalign-center": "Column centred", "colalign-right": "Column right-aligned",
	"slide-left": "Slide text left", "slide-center": "Slide text centre", "slide-right": "Slide text right",
	"link": "Link",
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
var formatTips = map[string]string{
	"bold":            "Make the selected text, or the word at the caret, bold: **text**. Press again to undo.",
	"italic":          "Italic: _text_. Press again to undo.",
	"underline":       "Markdown has no underline: this wraps the text in <u>…</u>, a tag Marp's default HTML allow-list keeps, so it works with raw HTML off. Press again to undo.",
	"strike":          "Strikethrough: ~~text~~. Press again to undo.",
	"code":            "Inline code: `text`. Press again to undo.",
	"h1":              "Make the line (or each selected line) a heading 1: # text. Press again for plain text.",
	"h2":              "Heading 2: ## text.",
	"h3":              "Heading 3: ### text.",
	"bullets":         "Bulleted list: - item, one per selected line. Press again to undo.",
	"numbered":        "Numbered list: 1. item, 2. item… Press again to undo.",
	"quote":           "Quote: > text. Press again to take one level off. Ctrl+Shift+.",
	"rule":            "A horizontal rule. Marp reads ---, *** and ___ as a new slide, so this writes the <hr> tag, which the default HTML allow-list keeps.",
	"table":           "Insert a table after the current line: choose its columns and rows.",
	"table-format":    "Align the table around the caret: every column padded to its widest cell, the alignment marks kept.",
	"row-above":       "Add an empty row above the caret's row (under the header from the header).",
	"row-below":       "Add an empty row below the caret's row.",
	"row-delete":      "Remove the caret's row; the header stays.",
	"col-left":        "Add a column left of the caret's column.",
	"col-right":       "Add a column right of the caret's column.",
	"col-delete":      "Remove the caret's column; a table keeps one.",
	"colalign-left":   "Align the caret's table column left (:--). Press again to clear it.",
	"colalign-center": "Centre the caret's table column (:-:). Press again to clear it.",
	"colalign-right":  "Align the caret's table column right (--:). Press again to clear it.",
	"slide-left":      "Put the slide's text back at the left, the theme's own: removes the style block this tab wrote.",
	"slide-center":    "Centre the text of the slide the caret is in, with a <style scoped> block at the top of the slide. It works with raw HTML off and shows in the preview, the HTML page and the PDF.",
	"slide-right":     "Right-align the text of the slide the caret is in (a <style scoped> block).",
	"link":            "Make the selection, or the word, a link: [text](url) with the url selected for typing. In a link, select its url.",
}

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
	return &toolkit.RibbonTab{Label: "Format", Groups: []*toolkit.RibbonGroup{
		{Label: "Text", Icon: icon("text"), Priority: 7, Controls: []*toolkit.RibbonControl{
			formatToggle(vm, "bold", false), formatToggle(vm, "italic", false), formatToggle(vm, "underline", false),
			formatToggle(vm, "strike", false), formatToggle(vm, "code", false),
		}},
		{Label: "Paragraph", Icon: icon("list"), Priority: 6, Controls: []*toolkit.RibbonControl{
			formatToggle(vm, "h1", false), formatToggle(vm, "h2", false), formatToggle(vm, "h3", false),
			formatToggle(vm, "bullets", false), formatToggle(vm, "numbered", false), formatToggle(vm, "quote", false),
			formatButton(vm, "rule", false),
		}},
		{Label: "Table", Icon: icon("table"), Priority: 5, Controls: []*toolkit.RibbonControl{
			table, formatButton(vm, "table-format", false),
		}},
		{Label: "Rows and columns", Icon: icon("table-rows"), Priority: 4, Controls: []*toolkit.RibbonControl{
			formatButton(vm, "row-above", false), formatButton(vm, "row-below", false), formatButton(vm, "row-delete", false),
			formatButton(vm, "col-left", false), formatButton(vm, "col-right", false), formatButton(vm, "col-delete", false),
		}},
		{Label: "Column", Icon: icon("align-center"), Priority: 3, Controls: []*toolkit.RibbonControl{
			formatToggle(vm, "colalign-left", false), formatToggle(vm, "colalign-center", false), formatToggle(vm, "colalign-right", false),
		}},
		{Label: "Slide text", Icon: icon("align-left"), Priority: 2, Controls: []*toolkit.RibbonControl{
			formatToggle(vm, "slide-left", false), formatToggle(vm, "slide-center", false), formatToggle(vm, "slide-right", false),
		}},
		{Label: "Link", Icon: icon("link"), Priority: 1, Controls: []*toolkit.RibbonControl{
			formatToggle(vm, "link", true),
		}},
	}}
}
