package playground

import "strings"

// GeneralHelp is what the help pane says when the caret is on nothing it
// knows: how to use the playground. One paragraph or item per line: the
// pane wraps them to its width.
const GeneralHelp = `Type a deck on the left; the slides follow on the right.

- Slides are separated by a line of three dashes: ---
- The front matter, between the first two --- lines, holds the global directives (theme, paginate, header, footer, size) and go-slides' own keys: kinds, data, template.
- <!-- _class: lead --> in a slide is a directive for that slide alone.
- {{ .title }}, {{ page }}/{{ total }}, {{ range seq 3 }}: Go templates.
- kind: title applies a kind, a named set of directives.
- Math: $e^{i\pi}$ inline, $$ ... $$ on its own lines.
- Images: ![w:300](url) sizes one, ![sepia](url) filters it; ![bg](url) is a background, ![bg left:40%](url) a split one beside the text.

Files, left of the editor, is the deck's project, as the slides command reads it: deck.md, slides.yaml, themes/, kinds/, media/. Click a file to edit it. Right-click (or Shift+F10) for New File, New Folder, Rename (F2), Delete, New theme, New kind; drag a file onto a folder to move it. The Project menu makes a theme built on default, gaia or uncover (it joins the Theme list), a kind, a data file, and opens or downloads the whole project as a .zip (unzipped, slides build deck.md builds it).

Panes: drag the bar between two panes to resize them (Files, the editor, Slides, the preview, Help); double-click a bar, or press Enter on it, to pull the pane shut and bring it back; F6 moves between the bars, the arrow keys move one by 16 px, Home and End go to the smallest and the largest, Escape returns to the editor. The sizes you set are kept in this browser; the View tab still switches panes off.

Format tab: Bold, Italic, Strikethrough and Inline code work on the selection or the word at the caret and toggle (Ctrl+B, Ctrl+I, Ctrl+Shift+X, Ctrl+E; Cmd on a Mac): each button shows pressed inside its marks, and one Undo takes one action back. Underline (Ctrl+U) wraps the text in <u>…</u>, because Markdown has no underline and Marp's default HTML list keeps that tag. Headings 1 to 3 (Ctrl+Alt+1..3), bullets (Ctrl+Shift+8), numbering (Ctrl+Shift+7) and quote work on the selected lines; the rule is an <hr> tag, because ---, *** and ___ would start a new slide. Table opens a grid of columns by rows; at the caret in a table you add or remove a row or a column, align the column (:--, :-:, --:) and Tidy keeps the cells lined up. Slide text left, centre or right writes a <style scoped> block into the current slide: it works with raw HTML off and shows in the preview, the HTML page and the PDF. Link (Ctrl+K) writes [text](url) with the url selected.

Drop or paste an image or a video on the editor: it goes into media/ and the deck names it, ![](media/photo.png); drop files on a folder of the panel to put them there. A file is at most 8 MiB and a project 64 MiB (the line under the files says how much is used). Nothing is uploaded. An https: image works too, but your browser fetches it from its server, which then sees your address (no referrer is sent).

Theme picks default, gaia, uncover or a theme your deck defines (/* @theme name */ in a <style>): it writes theme: into the front matter, and follows the front matter when you edit it.

Completion opens as you type a directive, a kind, or inside {{ }}. Put the caret on a directive, a kind or a template function and this pane explains it. Click a slide in the overview to go to its line.

Present (or p in the slideshow) opens the presenter view in a second window: the current slide, the next step, the speaker notes (<!-- comments -->), a clock and a timer. Arrow keys, Space, Home and End move both windows; the slideshow and the presenter follow each other, also across edits. Allow pop-ups for this site if the window does not open.

Your deck stays in this browser: it is saved as you type in this browser's own storage, for this site only, and comes back when you reload; nothing is sent anywhere. Draft restores the draft a shared link, an example or another tab replaced, resets to the example, and Forget my drafts erases them. Share link puts the deck in the link itself (not its other files: download the project's .zip for those).

The ribbon's tabs: Home (example, theme, Present, Share link, Draft), Insert (a new slide, an image or background from a file, math, a code block with line numbers, a kind: each is typed into the deck at the caret, around the selection if there is one, and Undo takes it back), View (the preview, the minimap, the slide thumbnails) and Export. Alt shows a letter over each tab; type it, then the letter over a control. Ctrl+F1 collapses the ribbon.

Export tab: the HTML page of the presentation, its media inside; a PDF (vector, its text selectable), PowerPoint or OpenDocument deck (editable), made in your browser exactly as the slides command makes them. The first export downloads the export engine (about 10 MiB, once: your browser keeps it), compiling it as it arrives, and then only the fonts your slides use; https: images are fetched only if their server allows it (CORS), data: images always.`

// DocLink is a link the help pane offers.
type DocLink struct {
	Title, URL string
}

// DocLinks are the documentation pages the help pane links to.
var DocLinks = []DocLink{
	{"Getting started", "https://go-slides.github.io/docs/latest/getting-started/"},
	{"Marp compatibility", "https://go-slides.github.io/docs/latest/marp/"},
	{"Templates", "https://go-slides.github.io/docs/latest/templates/"},
	{"Kinds", "https://go-slides.github.io/docs/latest/kinds/"},
	{"HCL", "https://go-slides.github.io/docs/latest/hcl/"},
	{"Command line", "https://go-slides.github.io/docs/latest/cli/"},
	{"VS Code", "https://go-slides.github.io/docs/latest/vscode/"},
	{"Playground", "https://go-slides.github.io/docs/latest/playground/"},
}

// plainMarkdown turns the editor package's Markdown help into text for a
// plain text pane: code fences dropped, emphasis and code marks removed.
func plainMarkdown(md string) string {
	var out []string
	for _, line := range strings.Split(md, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			continue
		}
		line = strings.NewReplacer("**", "", "`", "").Replace(line)
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
