package playground

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/coded"
	"github.com/go-widgets/toolkit"
)

// The playground's words, in the seven languages of the site (en, fr, de, es,
// zh, ko, ja). A message is a typed Key; its text is in i18n_<lang>.go, English
// being the reference every other catalogue is checked against
// (i18n_test.go: a key missing in a language, a placeholder that changed, a
// string literal left in a view all fail the build).
//
// The choice of a plain table over golang.org/x/text/message: the playground
// needs one language per page load (the page's language, chosen by its URL), no
// runtime switching, and plurals of the "one/other" kind only, which Tn covers;
// x/text/message would add its CLDR plural and number tables to a module that
// ships as WebAssembly. The table costs the strings themselves and no code.

// Lang is a language of the playground: the code of the site's languages.
type Lang string

// The seven languages. Chinese is Simplified (the site's zh is zh-Hans).
const (
	LangEN Lang = "en"
	LangFR Lang = "fr"
	LangDE Lang = "de"
	LangES Lang = "es"
	LangIT Lang = "it"
	LangPT Lang = "pt" // Brazilian Portuguese (pt-BR)
	LangPL Lang = "pl"
	LangUK Lang = "uk"
	LangRU Lang = "ru"
	LangZH Lang = "zh"
	LangKO Lang = "ko"
	LangJA Lang = "ja"
)

// Langs are the twelve languages, in the order of the site's selector.
var Langs = []Lang{LangEN, LangFR, LangDE, LangES, LangIT, LangPT, LangPL, LangUK, LangRU, LangZH, LangKO, LangJA}

// Key names a message. The constants are in i18n_keys.go; a few families (the
// Format tab's labels and tips, by control name) are built by the functions
// beside them.
type Key string

// catalogs holds each language's messages. English is the reference: a key
// another language lacks falls back to it, which the test forbids.
var catalogs = map[Lang]map[Key]string{
	LangEN: enMessages, LangFR: frMessages, LangDE: deMessages, LangES: esMessages,
	LangIT: itMessages, LangPT: ptMessages, LangPL: plMessages, LangUK: ukMessages, LangRU: ruMessages,
	LangZH: zhMessages, LangKO: koMessages, LangJA: jaMessages,
}

// locale is the language the playground speaks, set once at start-up by the
// page (SetLocale) before the view-model and the scene are built.
var locale = LangEN

// Locale is the language in use.
func Locale() Lang { return locale }

// IsCJK is whether the language needs the CJK font.
func (l Lang) IsCJK() bool { return l == LangZH || l == LangKO || l == LangJA }

// IsCyrillic is whether the language is written in Cyrillic.
func (l Lang) IsCyrillic() bool { return l == LangUK || l == LangRU }

// ParseLang reads a language tag ("fr", "fr-CA", "zh-Hans", "ja_JP") as one of
// the seven; false for any other.
func ParseLang(tag string) (Lang, bool) {
	tag = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(tag), "_", "-"))
	base, _, _ := strings.Cut(tag, "-")
	for _, l := range Langs {
		if string(l) == base {
			return l, true
		}
	}
	return LangEN, false
}

// SetLocale makes l the language and rebuilds every table derived from the
// messages (the menus' items, the ribbon's labels...). A tag that is not one of
// the seven leaves English. It returns the language in effect.
func SetLocale(tag string) Lang {
	l, _ := ParseLang(tag)
	locale = l
	localizeToolkit()
	localise()
	return l
}

// T is the message k in the current language, with fmt's verbs filled from
// args. A key missing in the language is read in English; missing in English
// it is the key itself, which the test catches.
func T(k Key, args ...any) string {
	s, ok := catalogs[locale][k]
	if !ok {
		if s, ok = enMessages[k]; !ok {
			s = string(k)
		}
	}
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

// Tn is the message k in the plural form n takes in the current language: the
// catalogue holds k+".one", k+".few", k+".many" and k+".other" as the language's
// CLDR plural rules (https://cldr.unicode.org/index/cldr-spec/plural-rules)
// need them; "other" is always there and is what a missing form falls back to.
func Tn(k Key, n int, args ...any) string {
	cat := PluralCategory(locale, n)
	if s, ok := catalogs[locale][k+"."+Key(cat)]; ok {
		return fmtMsg(s, args)
	}
	return T(k+".other", args...)
}

func fmtMsg(s string, args []any) string {
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

// PluralCategory is the CLDR plural category of the integer n in l: "one",
// "few", "many" or "other".
func PluralCategory(l Lang, n int) string {
	if n < 0 {
		n = -n
	}
	m10, m100 := n%10, n%100
	switch l {
	case LangFR, LangPT:
		if n == 0 || n == 1 {
			return "one"
		}
	case LangEN, LangDE, LangES, LangIT:
		if n == 1 {
			return "one"
		}
	case LangPL:
		switch {
		case n == 1:
			return "one"
		case m10 >= 2 && m10 <= 4 && (m100 < 12 || m100 > 14):
			return "few"
		default:
			return "many"
		}
	case LangUK, LangRU:
		switch {
		case m10 == 1 && m100 != 11:
			return "one"
		case m10 >= 2 && m10 <= 4 && (m100 < 12 || m100 > 14):
			return "few"
		default:
			return "many"
		}
	}
	return "other"
}

// PluralForms are the categories a language's catalogue must give for a
// counted message.
func PluralForms(l Lang) []string {
	switch l {
	case LangPL, LangUK, LangRU:
		return []string{"one", "few", "many", "other"}
	case LangZH, LangKO, LangJA:
		return []string{"other"}
	}
	return []string{"one", "other"}
}

// months is the 12 abbreviations the date format reads, from the message.
func months(m time.Month) string {
	return strings.Split(T(KDateMonths), ",")[m-1]
}

// dateShort is a day, the way the language writes it: "Jan 2", "2 janv.".
func dateShort(t time.Time) string {
	return T(KDateShort, months(t.Month()), t.Day())
}

// clock is the time of day.
func clock(t time.Time) string { return t.Format("15:04") }

// ErrText is an error in the language in use: a coded error (store, media,
// templates) by its catalogue message "err.<code>", any other as it says itself.
func ErrText(err error) string {
	if err == nil {
		return ""
	}
	if c, ok := coded.As(err); ok {
		if s, ok := lookup(Key("err." + c.Code)); ok {
			return fmt.Sprintf(s, c.Args...)
		}
	}
	return err.Error()
}

// lookup is the message in the current language, English as the fallback,
// and whether there was one.
func lookup(k Key) (string, bool) {
	if s, ok := catalogs[locale][k]; ok {
		return s, true
	}
	s, ok := enMessages[k]
	return s, ok
}

// codedCodes are the codes of the errors the store, media and templates
// packages return (coded.Error): each has the message "err.<code>".
var codedCodes = []string{
	"store.pathlen", "store.pathrel", "store.pathctl", "store.pathleaves", "store.filesize", "store.maxfiles",
	"store.projectsize", "store.nostore", "media.notfile", "media.textover", "media.notmedium", "media.mediaover",
	"tpl.mapping", "tpl.config", "tpl.notmapping",
}

// explainCode is what a diagnostic's code means, in the language in use; ""
// in English (the compiler's own message says it) or for no code.
func explainCode(code string) string {
	if code == "" || locale == LangEN {
		return ""
	}
	s, _ := lookup(Key("code." + code))
	return s
}

// localizeToolkit makes the toolkit say its own words in the language in use.
func localizeToolkit() {
	if locale == LangEN {
		toolkit.SetLocalizer(nil)
		return
	}
	toolkit.SetLocalizer(func(key, english string) string {
		s, _ := catalogs[locale][Key("tk."+key)]
		return s
	})
}

// CatalogueRunes are the distinct runes the language's messages draw: what a
// font subset for the language must hold, whatever else it holds.
func CatalogueRunes(l Lang) []rune {
	seen := map[rune]bool{}
	var out []rune
	for _, s := range catalogs[l] {
		for _, r := range s {
			if r >= ' ' && !seen[r] {
				seen[r] = true
				out = append(out, r)
			}
		}
	}
	return out
}

// localise rebuilds the tables that hold messages as values (the menus' items,
// the ribbon's labels, the examples' titles...) after the language changed.
func localise() {
	ModeNames = []string{T(KModeOverview), T(KModeSlideshow)}
	ExampleNames = exampleTitles()
	SaveItems = saveItems()
	ProjectItems = projectItems()
	KindTitle = T(KKindTitle)
	GeneralHelp = T(KHelpGeneral)
	DocLinks = docLinks()
	formatLabels = formatTexts(fmtLabelKey)
	formatTips = formatTexts(fmtTipKey)
	ribbonTabs = tabLabels()
	RibbonLabels = ribbonLabels()
}

// msgError is an error that says itself in the language in use when it is
// read, not when it was made: a package-level error value is made before the
// language is known.
type msgError struct {
	key  Key
	args []any
}

func (e *msgError) Error() string { return T(e.key, e.args...) }

// Errorf is an error carrying the message k, for the page's code (wasm).
func Errorf(k Key, args ...any) error { return errf(k, args...) }

// errf is an error carrying the message k with its arguments.
func errf(k Key, args ...any) error { return &msgError{k, args} }
