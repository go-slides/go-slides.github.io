// Package exportdoc renders a project's deck as the exports take it (a
// go-slides doc.Doc), as `slides --pdf` renders it: the project's
// slides.yaml, themes, kinds and data files, the box frame, raw HTML as the
// allowlist lets it through. The preview's renderer (playground.Serve)
// runs it in its worker and hands the Doc to the export worker as JSON, so
// that the export module carries no renderer of its own (playground/exporter).
package exportdoc

import (
	"errors"

	"github.com/go-slides/go-slides.github.io/playground/mediafiles"
	"github.com/go-slides/slides"
	"github.com/go-slides/slides/deck"
	"github.com/go-slides/slides/doc"
	"github.com/go-slides/slides/project"
	"github.com/go-slides/slides/vfs"
)

// MaxSource bounds the deck, as the playground's renderer does.
const MaxSource = 1 << 20

// DefaultDeckPath is where a deck with no project is.
const DefaultDeckPath = "/project/deck.md"

// Project is a deck with its project's text files, as the preview renders
// it: the deck's text at DeckPath, the project's text files (slides.yaml,
// themes, kinds, data) by their paths beside it, and its folders. The media
// go to the export itself (exporter.Export inlines them).
type Project struct {
	DeckPath string            `json:"deckPath"`
	Source   string            `json:"source"`
	Files    map[string]string `json:"files,omitempty"`
	Dirs     []string          `json:"dirs,omitempty"`
}

// Build renders p's deck in the box frame, with the project's
// configuration, into the Doc the exports take.
func Build(p Project) (*doc.Doc, error) {
	if len(p.Source) > MaxSource {
		return nil, errors.New("the deck is too large to export here: use the slides command")
	}
	if p.DeckPath == "" {
		p.DeckPath = DefaultDeckPath
	}
	listed, err := mediafiles.Check(p.DeckPath, p.Files, p.Dirs, nil)
	if err != nil {
		return nil, err
	}
	m := map[string][]byte{p.DeckPath: []byte(p.Source)}
	for k, v := range p.Files {
		m[k] = []byte(v)
	}
	fs := vfs.NewMap(m)
	for d := range listed {
		fs.Dir(d)
	}
	// the project's slides.yaml is read, as the slides command reads it
	cfg, err := project.Load(p.DeckPath, project.Options{FS: fs, NoConfig: len(p.Files) == 0})
	if err != nil {
		return nil, err
	}
	cfg.HTML = deck.HTMLAllowlist
	b, err := slides.Render([]byte(p.Source), cfg, deck.FrameBox)
	if err != nil {
		return nil, err
	}
	return b.Doc(), nil
}
