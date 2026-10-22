package playground

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// ConfigPath is the project's configuration, which the slides command and
// the VS Code extension read next to the deck.
const ConfigPath = "slides.yaml"

// The templates the file panel starts a theme, a kind or a data file from.
// Each says what it is and how the deck uses it, for the reader who opens
// one without having read the documentation.

// ThemeTemplate is a theme built on a built-in one (@import), named name.
func ThemeTemplate(name, base string) string {
	return "/* @theme " + name + " */\n" +
		"/*\n" +
		" * A theme of this project, built on " + base + ": what is written here\n" +
		" * changes it, the rest stays " + base + "'s. The deck uses it with\n" +
		" * `theme: " + name + "` (or the Theme list); slides.yaml lists the themes\n" +
		" * folder, so the slides command finds it too.\n" +
		" */\n" +
		"@import '" + base + "';\n\n" +
		"section {\n  /* background: #fdfcf9; */\n  /* font-family: Georgia, serif; */\n}\n\n" +
		"h1 {\n  color: #4f46e5;\n}\n"
}

// KindTemplate is a kinds file holding one kind, named name.
func KindTemplate(name string) string {
	return "# Slide kinds: named sets of directives. A slide takes one with\n" +
		"#   <!-- _kind: " + name + " -->   (this slide)\n" +
		"#   <!-- kind: " + name + " -->    (this slide and the next ones)\n" +
		"# slides.yaml lists this file under kinds, for the slides command too.\n" +
		name + ":\n" +
		"  extends: section\n" +
		"  class: invert\n" +
		"  paginate: false\n"
}

// DataTemplate is a template data file.
func DataTemplate() string {
	return "# Template data: the deck reads it as {{ .title }}, {{ range .points }}…\n" +
		"# slides.yaml lists this file under data, for the slides command too.\n" +
		"title: My talk\n" +
		"points:\n" +
		"  - The first point\n" +
		"  - The second point\n"
}

// ConfigTemplate is a new project's configuration.
const ConfigTemplate = "# The project's configuration, read by the slides command and the\n" +
	"# VS Code extension too:  slides build deck.md\n"

// configKeys are the configuration's lists of project files.
var configKeys = []string{"themes", "kinds", "data"}

// ListInConfig adds value to the list under key in slides.yaml's text
// (making the file's text when it is empty), keeping its comments; a key
// that holds one path becomes a list of two; a value listed already
// changes nothing.
func ListInConfig(src, key, value string) (string, error) {
	if strings.TrimSpace(src) == "" {
		src = ConfigTemplate
	}
	doc, root, err := configRoot(src)
	if err != nil {
		return src, err
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != key {
			continue
		}
		v := root.Content[i+1]
		switch {
		case v.Kind == yaml.ScalarNode && v.Value == value:
			return src, nil
		case v.Kind == yaml.ScalarNode && v.Tag != "!!null":
			old := *v
			*v = yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{&old, str(value)}}
		case v.Kind == yaml.ScalarNode:
			*v = yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{str(value)}}
		case v.Kind == yaml.SequenceNode:
			for _, c := range v.Content {
				if c.Value == value {
					return src, nil
				}
			}
			v.Content = append(v.Content, str(value))
		default:
			return src, fmt.Errorf("%s: %s is a mapping; list %s in it by hand", ConfigPath, key, value)
		}
		return encodeConfig(doc)
	}
	root.Content = append(root.Content, str(key), &yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle, Content: []*yaml.Node{str(value)}})
	return encodeConfig(doc)
}

// RenameInConfig replaces a path listed under themes, kinds or data (a file
// renamed or moved); remove drops it instead (a file deleted). The text is
// returned as it is when nothing names the path.
func RenameInConfig(src, from, to string, remove bool) (string, error) {
	if strings.TrimSpace(src) == "" {
		return src, nil
	}
	doc, root, err := configRoot(src)
	if err != nil {
		return src, err
	}
	changed := false
	match := func(v string) (string, bool) {
		switch {
		case v == from:
			return to, true
		case strings.HasPrefix(v, from+"/"): // in a folder renamed
			return to + v[len(from):], true
		}
		return "", false
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if !slices.Contains(configKeys, root.Content[i].Value) {
			continue
		}
		v := root.Content[i+1]
		switch v.Kind {
		case yaml.ScalarNode:
			if n, ok := match(v.Value); ok {
				changed = true
				if remove {
					*v = yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
				} else {
					v.Value = n
				}
			}
		case yaml.SequenceNode:
			kept := v.Content[:0]
			for _, c := range v.Content {
				if n, ok := match(c.Value); ok {
					changed = true
					if remove {
						continue
					}
					c.Value = n
				}
				kept = append(kept, c)
			}
			v.Content = kept
		}
	}
	if !changed {
		return src, nil
	}
	return encodeConfig(doc)
}

// configRoot parses slides.yaml: the document and its top mapping.
func configRoot(src string) (*yaml.Node, *yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(src), &doc); err != nil {
		return nil, nil, fmt.Errorf("%s: %v", ConfigPath, err)
	}
	if doc.Kind == 0 { // only comments
		head := strings.TrimRight(src, "\n")
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, HeadComment: head}}}
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, nil, errors.New(ConfigPath + ": the configuration is not a mapping")
	}
	return &doc, doc.Content[0], nil
}

func str(s string) *yaml.Node { return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s} }

func encodeConfig(doc *yaml.Node) (string, error) {
	var b bytes.Buffer
	e := yaml.NewEncoder(&b)
	e.SetIndent(2)
	if err := e.Encode(doc); err != nil {
		return "", err
	}
	return b.String(), nil
}
