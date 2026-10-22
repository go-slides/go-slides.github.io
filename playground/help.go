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

Drop or paste an image file (PNG, JPEG, GIF, WebP or SVG, up to 256 KiB) into the editor: it goes into the deck as a data: URI, so it travels with the deck and nothing is uploaded. An https: image works too, but your browser fetches it from its server, which then sees your address (no referrer is sent).

Completion opens as you type a directive, a kind, or inside {{ }}. Put the caret on a directive, a kind or a template function and this pane explains it. Click a slide in the overview to go to its line.

Present (or p in the slideshow) opens the presenter view in a second window: the current slide, the next step, the speaker notes (<!-- comments -->), a clock and a timer. Arrow keys, Space, Home and End move both windows; the slideshow and the presenter follow each other, also across edits. Allow pop-ups for this site if the window does not open.

Your deck stays in this browser: Share link puts it in the link itself, and Download HTML saves the presentation.`

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
