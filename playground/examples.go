package playground

import (
	"cmp"
	"embed"
	"io/fs"
	"math"
	"path"
	"slices"
	"strconv"
	"strings"
)

//go:embed examples/*.md
var exampleFS embed.FS

// Example is a deck the playground offers.
type Example struct {
	// Name is the title in the picker, in the current language; the decks
	// themselves stay in English (no translation of them exists).
	Name, Source string
	// File is the deck's file name in examples/.
	File string
}

// exampleNames are the examples' titles (messages), by the name of their files.
var exampleNames = map[string]Key{
	"1-welcome.md":   KExWelcome,
	"2-templates.md": KExTemplates,
	"3-kinds.md":     KExKinds,
	"4-hcl.md":       KExHCL,
	"5-gaia.md":      KExGaia,
	"6-media.md":     KExMedia,

	// The gallery's single-file decks (go-slides/slides' examples/, shown in
	// the documentation's gallery). The ones that need files beside the deck
	// (kinds file, parts, images) are zips there, which the file panel
	// imports.
	"7-theme-default.md":    KExGalleryDefault,
	"8-theme-gaia.md":       KExGalleryGaia,
	"9-theme-uncover.md":    KExGalleryUncover,
	"10-math.md":            KExGalleryMath,
	"11-code.md":            KExGalleryCode,
	"12-emoji-and-fit.md":   KExGalleryEmoji,
	"13-presenting.md":      KExGalleryPresenting,
	"14-hcl-expressions.md": KExGalleryHCL,
}

// Examples are the example decks, in order.
var Examples = loadExamples()

// ExampleNames are their titles, for the picker. Rebuilt by SetLocale.
var ExampleNames = exampleTitles()

// exampleTitles names the examples in the current language.
func exampleTitles() []string {
	var out []string
	for i := range Examples {
		e := &Examples[i]
		if k, ok := exampleNames[e.File]; ok {
			e.Name = T(k)
		}
		out = append(out, e.Name)
	}
	return out
}

// examplePosition is the number a file name starts with ("10-math.md" is 10);
// a name without one goes last.
func examplePosition(name string) int {
	n, _, ok := strings.Cut(name, "-")
	if i, err := strconv.Atoi(n); ok && err == nil {
		return i
	}
	return math.MaxInt
}

func loadExamples() []Example {
	entries, _ := exampleFS.ReadDir("examples")
	// In the order of the number the file starts with: "10-" follows "9-".
	slices.SortStableFunc(entries, func(a, b fs.DirEntry) int {
		return cmp.Compare(examplePosition(a.Name()), examplePosition(b.Name()))
	})
	var out []Example
	for _, e := range entries {
		b, _ := exampleFS.ReadFile(path.Join("examples", e.Name()))
		name := strings.TrimSuffix(e.Name(), ".md")
		if k, ok := exampleNames[e.Name()]; ok {
			name = T(k)
		}
		out = append(out, Example{Name: name, Source: string(b), File: e.Name()})
	}
	return out
}
