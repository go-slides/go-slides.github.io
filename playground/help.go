package playground

import "strings"

// GeneralHelp is what the help pane says when the caret is on nothing it
// knows: how to use the playground. One paragraph or item per line: the
// pane wraps them to its width. The text is the catalogue's (KHelpGeneral);
// rebuilt by SetLocale.
var GeneralHelp = T(KHelpGeneral)

// DocLink is a link the help pane offers.
type DocLink struct {
	Title, URL string
}

// docsBase is where the documentation is; the languages other than English
// have a folder of their own in it, as the site's (docs-lang.html).
const docsBase = "https://go-slides.github.io/docs/latest/"

// docPages are the documentation pages the help pane links to: the title's
// message and the page's path.
var docPages = []struct {
	title Key
	page  string
}{
	{KDocGettingStarted, "getting-started/"},
	{KDocMarp, "marp/"},
	{KDocTemplates, "templates/"},
	{KDocKinds, "kinds/"},
	{KDocHCL, "hcl/"},
	{KDocCLI, "cli/"},
	{KDocVSCode, "vscode/"},
	{KDocPlayground, "playground/"},
}

// DocLinks are the documentation pages the help pane links to, in the
// language of the playground. Rebuilt by SetLocale.
var DocLinks = docLinks()

func docLinks() []DocLink {
	base := docsBase
	if locale != LangEN {
		base += string(locale) + "/"
	}
	out := make([]DocLink, len(docPages))
	for i, p := range docPages {
		out[i] = DocLink{Title: T(p.title), URL: base + p.page}
	}
	return out
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
