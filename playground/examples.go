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
	Name, Source string
}

// exampleNames are the examples' titles, in the order of their files.
var exampleNames = map[string]string{
	"1-welcome.md":   "Welcome",
	"2-templates.md": "Templates: generated slides",
	"3-kinds.md":     "Slide kinds",
	"4-hcl.md":       "HCL front matter",
	"5-gaia.md":      "Themes and spot directives",
	"6-media.md":     "Math, images and backgrounds",

	// The gallery's single-file decks (go-slides/slides' examples/, shown in
	// the documentation's gallery). The ones that need files beside the deck
	// (kinds file, parts, images) are zips there, which the file panel
	// imports.
	"7-theme-default.md":    "Gallery: the default theme",
	"8-theme-gaia.md":       "Gallery: the gaia theme",
	"9-theme-uncover.md":    "Gallery: the uncover theme",
	"10-math.md":            "Gallery: math",
	"11-code.md":            "Gallery: code, line numbers and callouts",
	"12-emoji-and-fit.md":   "Gallery: emoji and fitting headings",
	"13-presenting.md":      "Gallery: transitions and speaker notes",
	"14-hcl-expressions.md": "Gallery: HCL expressions and typed data",
}

// Examples are the example decks, in order.
var Examples = loadExamples()

// ExampleNames are their titles, for the picker.
var ExampleNames = func() []string {
	var out []string
	for _, e := range Examples {
		out = append(out, e.Name)
	}
	return out
}()

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
		name := exampleNames[e.Name()]
		if name == "" {
			name = strings.TrimSuffix(e.Name(), ".md")
		}
		out = append(out, Example{Name: name, Source: string(b)})
	}
	return out
}
