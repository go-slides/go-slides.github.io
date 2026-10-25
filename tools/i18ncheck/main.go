// Command i18ncheck checks the landing's translations against the English
// words: every language of hugo.toml has i18n/<lang>.yaml with the keys of
// i18n/en.yaml, and a config/_default/params.<lang>.toml that translates the
// cards of hugo.toml (outputs by id, repositories by name, the hero's cards in
// order), and what is not prose is what the English has: the HTML tags and
// their links, the contents of <code>, the {{ .x }} placeholders.
//
// Usage, from the repository's root:
//
//	go run ./tools/i18ncheck
//
// It exits 1 and lists what differs. Standard library only.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	root := flag.String("root", ".", "the landing repository's root")
	flag.Parse()
	problems, err := Check(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "i18ncheck:", err)
		os.Exit(1)
	}
	for _, p := range problems {
		fmt.Println(p)
	}
	if len(problems) > 0 {
		fmt.Printf("i18ncheck: %d problem(s)\n", len(problems))
		os.Exit(1)
	}
	fmt.Println("i18ncheck: every language has its words and its cards")
}

// Check returns the differences between the translations and the English.
func Check(root string) ([]string, error) {
	hugo, err := os.ReadFile(filepath.Join(root, "hugo.toml"))
	if err != nil {
		return nil, err
	}
	langs := Languages(string(hugo))
	if len(langs) == 0 {
		return nil, fmt.Errorf("hugo.toml declares no language besides en")
	}
	en, err := YAML(filepath.Join(root, "i18n", "en.yaml"))
	if err != nil {
		return nil, err
	}
	outs, repos, heroCards := Cards(string(hugo))
	var problems []string
	add := func(format string, a ...any) { problems = append(problems, fmt.Sprintf(format, a...)) }
	for _, lang := range langs {
		tr, err := YAML(filepath.Join(root, "i18n", lang+".yaml"))
		if err != nil {
			add("%s: %v", lang, err)
			continue
		}
		for _, k := range keys(en) {
			v, ok := tr[k]
			if !ok {
				add("i18n/%s.yaml: %s is missing", lang, k)
				continue
			}
			for _, d := range Differences(en[k], v) {
				add("i18n/%s.yaml: %s: %s", lang, k, d)
			}
		}
		for _, k := range keys(tr) {
			if _, ok := en[k]; !ok {
				add("i18n/%s.yaml: %s is not an English key", lang, k)
			}
		}
		cfg, err := os.ReadFile(filepath.Join(root, "config", "_default", "params."+lang+".toml"))
		if err != nil {
			add("%s: %v", lang, err)
			continue
		}
		t := Tables(string(cfg))
		for id, c := range outs {
			tbl, ok := t["tr.out."+id]
			if !ok {
				add("params.%s.toml: [tr.out.%s] is missing", lang, id)
				continue
			}
			for _, f := range []string{"name", "what"} {
				if tbl[f] == "" {
					add("params.%s.toml: [tr.out.%s] has no %s", lang, id, f)
				}
			}
			for _, d := range Differences(c["what"], tbl["what"]) {
				add("params.%s.toml: [tr.out.%s] what: %s", lang, id, d)
			}
		}
		for name, c := range repos {
			tbl, ok := t["tr.repo."+name]
			if !ok {
				add("params.%s.toml: [tr.repo.%s] is missing", lang, name)
				continue
			}
			for _, d := range Differences(c["blurb"], tbl["blurb"]) {
				add("params.%s.toml: [tr.repo.%s] blurb: %s", lang, name, d)
			}
		}
		for name := range t {
			if id, ok := strings.CutPrefix(name, "tr.out."); ok {
				if _, known := outs[id]; !known {
					add("params.%s.toml: [%s] is not an output of hugo.toml", lang, name)
				}
			}
			if id, ok := strings.CutPrefix(name, "tr.repo."); ok {
				if _, known := repos[id]; !known {
					add("params.%s.toml: [%s] is not a repository of hugo.toml", lang, name)
				}
			}
		}
		got := ArrayTables(string(cfg), "tr.hero.cards")
		if len(got) != len(heroCards) {
			add("params.%s.toml: %d [[tr.hero.cards]], hugo.toml has %d cards", lang, len(got), len(heroCards))
		} else {
			for i, c := range heroCards {
				for _, f := range []string{"alt", "caption", "title", "text", "more"} {
					if (c[f] == "") != (got[i][f] == "") {
						add("params.%s.toml: hero card %d: %s must be present exactly when the English card has one", lang, i+1, f)
					}
				}
			}
		}
		if t["tr.hero"]["label"] == "" {
			add("params.%s.toml: [tr.hero] has no label", lang)
		}
	}
	return problems, nil
}

func keys[V any](m map[string]V) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

var langRE = regexp.MustCompile(`(?m)^\s*\[languages\.([a-z]{2}(?:-[A-Za-z]+)?)\]\s*$`)

// Languages returns the languages hugo.toml declares, other than en.
func Languages(hugo string) []string {
	var out []string
	for _, m := range langRE.FindAllStringSubmatch(hugo, -1) {
		if m[1] != "en" {
			out = append(out, m[1])
		}
	}
	return out
}

var yamlRE = regexp.MustCompile(`^([A-Za-z0-9_]+): "(.*)"$`)

// YAML reads a flat file of key: "value" lines.
func YAML(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for i, line := range strings.Split(string(b), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m := yamlRE.FindStringSubmatch(line)
		if m == nil {
			return nil, fmt.Errorf("%s:%d is not a `key: \"value\"` line", path, i+1)
		}
		v, err := strconv.Unquote(`"` + m[2] + `"`)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %v", path, i+1, err)
		}
		if _, dup := out[m[1]]; dup {
			return nil, fmt.Errorf("%s:%d: %s twice", path, i+1, m[1])
		}
		out[m[1]] = v
	}
	return out, nil
}

var (
	headerRE = regexp.MustCompile(`^\[(\[?)([A-Za-z0-9_.\-]+)\]\]?\s*$`)
	assignRE = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_-]*)\s*=\s*(.*)$`)
)

type block struct {
	name   string
	array  bool
	fields map[string]string
}

// blocks reads a TOML file as its [tables] and [[array tables]], with the
// string and number values of their key = value lines (strings unquoted).
func blocks(toml string) []block {
	var out []block
	cur := -1
	for _, line := range strings.Split(toml, "\n") {
		if m := headerRE.FindStringSubmatch(line); m != nil {
			out = append(out, block{name: m[2], array: m[1] == "[", fields: map[string]string{}})
			cur = len(out) - 1
			continue
		}
		if cur < 0 || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if m := assignRE.FindStringSubmatch(line); m != nil {
			v := strings.TrimSpace(m[2])
			if u, err := strconv.Unquote(v); err == nil {
				v = u
			}
			out[cur].fields[m[1]] = v
		}
	}
	return out
}

// Tables returns the [tables] of a file by name.
func Tables(toml string) map[string]map[string]string {
	out := map[string]map[string]string{}
	for _, b := range blocks(toml) {
		if !b.array {
			out[b.name] = b.fields
		}
	}
	return out
}

// ArrayTables returns the [[name]] tables of a file, in order.
func ArrayTables(toml, name string) []map[string]string {
	var out []map[string]string
	for _, b := range blocks(toml) {
		if b.array && b.name == name {
			out = append(out, b.fields)
		}
	}
	return out
}

// Cards reads hugo.toml's outputs (by id), repositories (by name) and hero
// cards (in order).
func Cards(hugo string) (outs, repos map[string]map[string]string, hero []map[string]string) {
	outs, repos = map[string]map[string]string{}, map[string]map[string]string{}
	for _, b := range blocks(hugo) {
		switch {
		case b.array && b.name == "params.outputs":
			if id := b.fields["id"]; id != "" {
				outs[id] = b.fields
			}
		case b.array && b.name == "params.repos":
			repos[b.fields["name"]] = b.fields
		case b.array && b.name == "params.hero.cards":
			hero = append(hero, b.fields)
		}
	}
	return
}

var (
	tagRE         = regexp.MustCompile(`<[^>]+>`)
	placeholderRE = regexp.MustCompile(`\{\{\s*\.[a-z]+\s*\}\}`)
	codeRE        = regexp.MustCompile(`(?s)<code>(.*?)</code>`)
	hrefRE        = regexp.MustCompile(`href="([^"]*)"`)
)

// Differences lists what a translation changes in what must stay: the tags,
// the links, the contents of <code> and the placeholders.
func Differences(en, tr string) []string {
	var out []string
	for _, c := range []struct {
		what string
		re   *regexp.Regexp
		pick int
	}{
		{"HTML tags", tagRE, 0},
		{"placeholders", placeholderRE, 0},
		{"<code> contents", codeRE, 1},
		{"links", hrefRE, 1},
	} {
		a, b := find(c.re, en, c.pick), find(c.re, tr, c.pick)
		if strings.Join(a, "\x00") != strings.Join(b, "\x00") {
			out = append(out, fmt.Sprintf("%s differ: English %q, translation %q", c.what, short(a), short(b)))
		}
	}
	return out
}

func find(re *regexp.Regexp, s string, pick int) []string {
	var out []string
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		out = append(out, m[pick])
	}
	sort.Strings(out)
	return out
}

func short(s []string) string {
	j := strings.Join(s, " ")
	if r := []rune(j); len(r) > 90 {
		return string(r[:90]) + "…"
	}
	return j
}
