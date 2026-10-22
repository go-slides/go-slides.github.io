package playground

import (
	"regexp"
	"strings"
)

// BuiltinThemes are Marp's themes, which go-slides bundles.
var BuiltinThemes = []string{"default", "gaia", "uncover"}

// themeName is what the theme picker writes into a deck: a name, never
// anything that could end the value or start another key.
var themeName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)

// atTheme finds the themes a deck defines in its own <style> (Marpit's
// /* @theme name */ metadata comment).
var atTheme = regexp.MustCompile(`@theme\s+([A-Za-z0-9][A-Za-z0-9_.-]{0,63})`)

// ThemeChoices are the themes the picker offers for a deck: the built-in
// ones, the ones the deck defines, and the one it names if it is neither
// (a typo, or a theme of the command's --theme-set), so the picker always
// shows what the deck says.
func ThemeChoices(src string) []string {
	out := append([]string(nil), BuiltinThemes...)
	seen := map[string]bool{}
	for _, t := range out {
		seen[t] = true
	}
	add := func(t string) {
		if t != "" && !seen[t] && themeName.MatchString(t) {
			seen[t] = true
			out = append(out, t)
		}
	}
	for _, m := range atTheme.FindAllStringSubmatch(src, 32) {
		add(m[1])
	}
	if t, _ := DeckTheme(src); t != "" {
		add(t)
	}
	return out
}

// frontMatter locates a deck's front matter: the index of its first and
// closing lines (a line of three or more dashes, as deck.Parse delimits
// it), and whether it is HCL (---hcl). ok is false when there is none.
func frontMatter(lines []string) (open, close int, hcl, ok bool) {
	if len(lines) == 0 {
		return 0, 0, false, false
	}
	first := strings.TrimRight(lines[0], " \t\r")
	switch {
	case first == "---hcl":
		hcl = true
	case len(first) >= 3 && strings.Trim(first, "-") == "":
	default:
		return 0, 0, false, false
	}
	for i := 1; i < len(lines); i++ {
		l := strings.TrimRight(lines[i], " \t\r")
		if len(l) >= 3 && strings.Trim(l, "-") == "" {
			return 0, i, hcl, true
		}
	}
	return 0, 0, false, false
}

// themeLine is the index of the front matter line that sets the theme (a
// top-level key: not indented), or -1.
func themeLine(lines []string, open, close int, hcl bool) int {
	for i := open + 1; i < close; i++ {
		l := lines[i]
		if l == "" || l[0] == ' ' || l[0] == '\t' {
			continue
		}
		if hcl {
			rest, found := strings.CutPrefix(l, "theme")
			if r := strings.TrimLeft(rest, " \t"); found && strings.HasPrefix(r, "=") && !strings.HasPrefix(r, "==") {
				return i
			}
		} else if strings.HasPrefix(l, "theme:") {
			return i
		}
	}
	return -1
}

// DeckTheme is the theme a deck's front matter names, and whether it names
// one. A quoted value is unquoted and a trailing comment dropped; an HCL
// value that is an expression, not a string, is returned as written.
func DeckTheme(src string) (string, bool) {
	lines := strings.Split(src, "\n")
	open, close, hcl, ok := frontMatter(lines)
	if !ok {
		return "", false
	}
	i := themeLine(lines, open, close, hcl)
	if i < 0 {
		return "", false
	}
	var v string
	if hcl {
		_, v, _ = strings.Cut(lines[i], "=")
		if k := strings.Index(v, "#"); k >= 0 {
			v = v[:k]
		}
		if k := strings.Index(v, "//"); k >= 0 {
			v = v[:k]
		}
	} else {
		_, v, _ = strings.Cut(lines[i], ":")
		if k := strings.Index(v, " #"); k >= 0 {
			v = v[:k]
		}
	}
	v = strings.TrimSpace(v)
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		v = v[1 : len(v)-1]
	}
	return v, v != ""
}

// SetTheme sets a deck's theme in its front matter, the deck's text being
// the one source of truth (a shared link keeps the theme): the theme line
// is rewritten, or added first in the front matter, or a front matter is
// added. HCL front matter gets HCL. A name that is not a plain name is
// refused, and the deck returned as it is.
func SetTheme(src, name string) string {
	if !themeName.MatchString(name) {
		return src
	}
	lines := strings.Split(src, "\n")
	open, close, hcl, ok := frontMatter(lines)
	if !ok {
		return "---\ntheme: " + name + "\n---\n\n" + src
	}
	line := "theme: " + name
	if hcl {
		line = `theme = "` + name + `"`
	}
	if i := themeLine(lines, open, close, hcl); i >= 0 {
		// keep a trailing comment
		old := lines[i]
		comment := ""
		if hcl {
			if k := strings.Index(old, "#"); k >= 0 {
				comment = " " + old[k:]
			} else if k := strings.Index(old, "//"); k >= 0 {
				comment = " " + old[k:]
			}
		} else if k := strings.Index(old, " #"); k >= 0 {
			comment = old[k:]
		}
		lines[i] = line + strings.TrimRight(comment, " ")
		return strings.Join(lines, "\n")
	}
	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:open+1]...)
	out = append(out, line)
	out = append(out, lines[open+1:]...)
	return strings.Join(out, "\n")
}
