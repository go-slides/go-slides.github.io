package playground

import (
	"path"
	"strings"

	"github.com/go-widgets/toolkit"
)

// The editor's language for a project file, by its extension: the deck's
// Markdown (deckSyntax), YAML and HCL (keys and comments, as in the front
// matter: frontKeys), JSON (keys) and CSS (comments, at-rules, properties).

// FileLanguage names the editor's language for a file.
func FileLanguage(p string) string {
	switch strings.ToLower(path.Ext(p)) {
	case ".md", ".markdown":
		return "Markdown"
	case ".yaml", ".yml":
		return "YAML"
	case ".hcl":
		return "HCL"
	case ".json":
		return "JSON"
	case ".css":
		return "CSS"
	}
	return "Text"
}

// fileSyntax is the highlighter of a file's language, nil for plain text.
func fileSyntax(p string) toolkit.Highlighter {
	switch FileLanguage(p) {
	case "Markdown":
		return deckSyntax{}
	case "YAML", "JSON":
		return keysSyntax{}
	case "HCL":
		return keysSyntax{hcl: true}
	case "CSS":
		return cssSyntax{}
	}
	return nil
}

// keysSyntax colours the keys and comments of YAML, JSON and HCL.
type keysSyntax struct{ hcl bool }

// Highlight implements toolkit.Highlighter.
func (k keysSyntax) Highlight(_ string, lines []string, theme *toolkit.Theme) [][]toolkit.TextSpan {
	key, _, _, comment, _, _ := syntaxColors(theme)
	out := make([][]toolkit.TextSpan, len(lines))
	for i, l := range lines {
		out[i] = frontKeys(l, k.hcl, key, comment)
	}
	return out
}

// cssSyntax colours a style sheet: comments (over several lines too),
// at-rules and property names.
type cssSyntax struct{}

// Highlight implements toolkit.Highlighter.
func (cssSyntax) Highlight(_ string, lines []string, theme *toolkit.Theme) [][]toolkit.TextSpan {
	key, _, heading, comment, _, _ := syntaxColors(theme)
	out := make([][]toolkit.TextSpan, len(lines))
	inComment, depth := false, 0
	for i, line := range lines {
		spans := delimited(line, "/*", "*/", comment, &inComment)
		inside := func(col int) bool {
			for _, s := range spans {
				if col >= s.Start && col < s.End {
					return true
				}
			}
			return false
		}
		r := []rune(line)
		for j := 0; j < len(r); j++ {
			if inside(j) {
				continue
			}
			switch c := r[j]; {
			case c == '{':
				depth++
			case c == '}':
				depth = max(depth-1, 0)
			case c == '@':
				e := j + 1
				for e < len(r) && (r[e] == '-' || r[e] >= 'a' && r[e] <= 'z') {
					e++
				}
				spans = append(spans, toolkit.TextSpan{Start: j, End: e, Color: heading})
				j = e - 1
			case depth > 0 && (c == '-' || c >= 'a' && c <= 'z') && (j == 0 || strings.ContainsRune(" \t;{", r[j-1])):
				e := j
				for e < len(r) && (r[e] == '-' || r[e] >= 'a' && r[e] <= 'z' || r[e] >= '0' && r[e] <= '9') {
					e++
				}
				k := e
				for k < len(r) && (r[k] == ' ' || r[k] == '\t') {
					k++
				}
				if k < len(r) && r[k] == ':' {
					spans = append(spans, toolkit.TextSpan{Start: j, End: e, Color: key})
				}
				j = e - 1
			}
		}
		out[i] = spans
	}
	return out
}
