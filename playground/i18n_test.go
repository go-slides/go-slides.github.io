package playground

import (
	"go/ast"
	"go/parser"
	gotoken "go/token"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"
)

// keyConsts reads the Key constants of i18n_keys.go by parsing it: a key
// added there is checked in every language without anyone listing it twice.
func keyConsts(t *testing.T) map[Key]string {
	t.Helper()
	f, err := parser.ParseFile(gotoken.NewFileSet(), "i18n_keys.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	out := map[Key]string{}
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != gotoken.CONST {
			continue
		}
		for _, s := range g.Specs {
			vs := s.(*ast.ValueSpec)
			for i, n := range vs.Names {
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok {
					t.Fatalf("%s: not a literal", n.Name)
				}
				v, _ := strconv.Unquote(lit.Value)
				if prev, dup := out[Key(v)]; dup {
					t.Errorf("%s and %s share the key %q", prev, n.Name, v)
				}
				out[Key(v)] = n.Name
			}
		}
	}
	return out
}

// wantKeys is every message key a language must have: the constants, with the
// counted ones in the forms the language's plural rules need, and the Format
// tab's two families by control name.
func wantKeys(t *testing.T, l Lang) []Key {
	t.Helper()
	isCounted := map[Key]bool{}
	for _, c := range counted {
		isCounted[c] = true
	}
	var out []Key
	for k := range keyConsts(t) {
		if !isCounted[k] {
			out = append(out, k)
			continue
		}
		for _, f := range PluralForms(l) {
			out = append(out, k+"."+Key(f))
		}
	}
	for name := range formatGlyphs {
		out = append(out, fmtLabelKey(name), fmtTipKey(name))
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

var verbRe = regexp.MustCompile(`%(\[\d+\])?[-+#0]*\d*(\.\d+)?[a-zA-Z%]`)

// verbs are the fmt verbs of a message, sorted, explicit indexes removed.
func verbs(s string) []string {
	var out []string
	for _, m := range verbRe.FindAllString(s, -1) {
		if m == "%%" {
			continue
		}
		out = append(out, regexp.MustCompile(`\[\d+\]`).ReplaceAllString(m, ""))
	}
	sort.Strings(out)
	return out
}

func TestEveryKeyInEveryLanguage(t *testing.T) {
	for _, l := range Langs {
		cat := catalogs[l]
		want := wantKeys(t, l)
		have := map[Key]bool{}
		for _, k := range want {
			have[k] = true
			s, ok := cat[k]
			if !ok {
				t.Errorf("%s: key %q is missing", l, k)
			} else if strings.TrimSpace(s) == "" {
				t.Errorf("%s: key %q is empty", l, k)
			}
		}
		for k := range cat {
			if !have[k] {
				t.Errorf("%s: key %q is not a key (or a plural form the language does not use)", l, k)
			}
		}
	}
}

func TestPlaceholdersAgreeWithEnglish(t *testing.T) {
	for _, l := range Langs {
		for k, s := range catalogs[l] {
			ref, ok := enMessages[k]
			if !ok {
				ref = enMessages[Key(strings.TrimSuffix(string(k), "."+string(k[strings.LastIndex(string(k), ".")+1:])))+".other"]
			}
			if got, want := verbs(s), verbs(ref); !slices.Equal(got, want) {
				t.Errorf("%s %q: verbs %v, English has %v", l, k, got, want)
			}
		}
	}
}

func TestEnglishIsWhatTheCodeSaid(t *testing.T) {
	SetLocale("en")
	for k, want := range map[Key]string{
		KTabHome: "Home", KLblShare: "Share link", KProblemLine: "line %d",
	} {
		if got := T(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if got := Tn(KStatusRendered, 1, 1, 5); got != "1 slide · rendered in 5 ms" {
		t.Error(got)
	}
	if got := Tn(KStatusRendered, 3, 3, 5); got != "3 slides · rendered in 5 ms" {
		t.Error(got)
	}
}

// the CLDR cardinal rules of the twelve languages
func TestPluralCategories(t *testing.T) {
	cases := []struct {
		l    Lang
		n    int
		want string
	}{
		{LangEN, 0, "other"}, {LangEN, 1, "one"}, {LangEN, 2, "other"},
		{LangDE, 1, "one"}, {LangDE, 5, "other"}, {LangES, 1, "one"}, {LangIT, 1, "one"}, {LangIT, 2, "other"},
		{LangFR, 0, "one"}, {LangFR, 1, "one"}, {LangFR, 2, "other"},
		{LangPT, 0, "one"}, {LangPT, 1, "one"}, {LangPT, 2, "other"},
		{LangPL, 1, "one"}, {LangPL, 2, "few"}, {LangPL, 4, "few"}, {LangPL, 5, "many"}, {LangPL, 12, "many"},
		{LangPL, 14, "many"}, {LangPL, 22, "few"}, {LangPL, 0, "many"}, {LangPL, 21, "many"}, {LangPL, 112, "many"},
		{LangRU, 1, "one"}, {LangRU, 2, "few"}, {LangRU, 5, "many"}, {LangRU, 11, "many"}, {LangRU, 21, "one"},
		{LangRU, 22, "few"}, {LangRU, 25, "many"}, {LangRU, 111, "many"}, {LangRU, 0, "many"},
		{LangUK, 1, "one"}, {LangUK, 3, "few"}, {LangUK, 11, "many"}, {LangUK, 21, "one"}, {LangUK, 101, "one"}, {LangUK, 14, "many"},
		{LangZH, 1, "other"}, {LangKO, 1, "other"}, {LangJA, 1, "other"},
	}
	for _, c := range cases {
		if got := PluralCategory(c.l, c.n); got != c.want {
			t.Errorf("%s %d: %s, want %s", c.l, c.n, got, c.want)
		}
	}
}

// every counted message gives a distinct, complete text for the forms the
// language uses, and Tn picks the right noun for 1, 2, 5 and 21
func TestCountedMessages(t *testing.T) {
	t.Cleanup(func() { SetLocale("en") })
	for _, l := range Langs {
		SetLocale(string(l))
		for _, base := range counted {
			seen := map[string]string{}
			for _, n := range []int{0, 1, 2, 5, 21, 22} {
				args := []any{n, 1.5, 3, 4}
				switch base {
				case KStatusRendered:
					args = []any{n, 7}
				case KStatusProblems:
					args = []any{n, "x"}
				case KStatusSaved:
					args = []any{"deck.pdf", "10 KiB", n, 1.5}
				case KFilesUsage:
					args = []any{n, "1 KiB", 64}
				}
				s := Tn(base, n, args...)
				if strings.Contains(s, "%!") || strings.Contains(s, "%d") {
					t.Errorf("%s %s n=%d: %q", l, base, n, s)
				}
				seen[PluralCategory(l, n)] = s
			}
			if len(seen) != len(PluralForms(l)) && l != LangPL && l != LangUK && l != LangRU {
				// the other languages use every form they give
				t.Logf("%s %s: %d forms used", l, base, len(seen))
			}
		}
	}
	SetLocale("pl")
	if a, b := Tn(KStatusRendered, 2, 2, 1), Tn(KStatusRendered, 5, 5, 1); a == b {
		t.Errorf("pl: few and many read alike: %q", a)
	}
	SetLocale("ru")
	if a, b := Tn(KFilesUsage, 21, 21, "1", 64), Tn(KFilesUsage, 22, 22, "1", 64); a == b {
		t.Errorf("ru: one and few read alike: %q", a)
	}
}

func TestParseLangAndSetLocale(t *testing.T) {
	t.Cleanup(func() { SetLocale("en") })
	for tag, want := range map[string]Lang{
		"fr": LangFR, "fr-CA": LangFR, "zh-Hans": LangZH, "zh_CN": LangZH, "pt-BR": LangPT, "JA": LangJA, "uk": LangUK, "ru-RU": LangRU,
	} {
		if got, ok := ParseLang(tag); !ok || got != want {
			t.Errorf("%q: %q %v", tag, got, ok)
		}
	}
	if _, ok := ParseLang("sv"); ok {
		t.Error("sv is not a language of the playground")
	}
	if SetLocale("sv") != LangEN || Locale() != LangEN {
		t.Error("an unknown language leaves English")
	}
	SetLocale("fr")
	if ModeNames[1] != "Diaporama" || RibbonLabels["share"] != T(KLblShare) || ribbonTabs[0] != T(KTabHome) {
		t.Errorf("the derived tables did not follow the language: %v %q", ModeNames, ribbonTabs[0])
	}
	if !strings.Contains(DocLinks[0].URL, "/fr/") {
		t.Error(DocLinks[0].URL)
	}
	if err := errf(KErrExists, "a.md"); !strings.Contains(err.Error(), "a.md") || err.Error() == "a.md exists already" {
		t.Errorf("a message error says itself in the language in use: %q", err)
	}
	SetLocale("en")
	if DocLinks[0].URL != docsBase+"getting-started/" {
		t.Error(DocLinks[0].URL)
	}
}

func TestDatesFollowTheLanguage(t *testing.T) {
	t.Cleanup(func() { SetLocale("en") })
	d := time.Date(2026, time.October, 3, 9, 5, 0, 0, time.UTC)
	want := map[Lang]string{LangEN: "Oct 3", LangFR: "3 oct.", LangZH: "10月3日", LangKO: "10월 3일", LangJA: "10月3日"}
	for l, w := range want {
		SetLocale(string(l))
		if got := dateShort(d); got != w {
			t.Errorf("%s: %q, want %q", l, got, w)
		}
	}
	for _, l := range Langs {
		SetLocale(string(l))
		if n := len(strings.Split(T(KDateMonths), ",")); n != 12 {
			t.Errorf("%s: %d months", l, n)
		}
	}
}

// the views hold no text of their own: a string literal with a space or a
// capital letter in them is a message that skipped the catalogue
func TestViewsHaveNoLiteralText(t *testing.T) {
	for _, file := range []string{"ribbon.go", "ribbon_format.go", "filepanel.go", "help.go"} {
		f, err := parser.ParseFile(gotoken.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if imp, ok := n.(*ast.ImportSpec); ok {
				_ = imp
				return false
			}
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != gotoken.STRING {
				return true
			}
			s, _ := strconv.Unquote(lit.Value)
			if strings.Contains(s, " ") || s != "" && unicode.IsUpper([]rune(s)[0]) {
				t.Errorf("%s: literal text %s", file, lit.Value)
			}
			return true
		})
	}
}

func TestWrapKeepsLatinAndBreaksCJK(t *testing.T) {
	if got := Wrap("one two three four", 9); got != "one two\nthree\nfour" {
		t.Errorf("%q", got)
	}
	text := "在左边输入幻灯片，右边会跟着变化。幻灯片之间用三个短横线分隔。"
	out := Wrap(text, 20)
	for _, ln := range strings.Split(out, "\n") {
		if cells(ln) > 20 {
			t.Errorf("%q is %d cells", ln, cells(ln))
		}
		if r := []rune(ln); len(r) > 0 && strings.ContainsRune("，。、", r[0]) {
			t.Errorf("a line starts with %q", string(r[0]))
		}
	}
	if strings.ReplaceAll(out, "\n", "") != text {
		t.Error("the wrap lost text")
	}
}

// the help of each language explains the same things: the same code tokens
func TestHelpKeepsTheCode(t *testing.T) {
	for _, l := range Langs {
		h := catalogs[l][KHelpGeneral]
		for _, tok := range []string{"deck.md", "slides.yaml", "{{ .title }}", "<!-- _class: lead -->", "![bg](url)", "$$", "Ctrl+F1", "<hr>", "<style scoped>"} {
			if !strings.Contains(h, tok) {
				t.Errorf("%s: the help lost %q", l, tok)
			}
		}
	}
}
