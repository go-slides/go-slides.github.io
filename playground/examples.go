package playground

import (
	"embed"
	"path"
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

func loadExamples() []Example {
	entries, _ := exampleFS.ReadDir("examples")
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
