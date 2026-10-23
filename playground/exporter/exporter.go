// Package exporter is the playground's PDF, PowerPoint and OpenDocument
// export: the slides command's own path (cmd/slides exportBytes) with its
// defaults, run in a Web Worker by a module of its own
// (cmd/playground-export) that the page downloads the first time an export
// is asked for. go-slides' export engine (go-webengine and the fonts it
// bundles) would otherwise more than double the editor's download.
package exporter

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/go-slides/go-slides.github.io/playground/mediafiles"
	"github.com/go-slides/slides"
	"github.com/go-slides/slides/deck"
	"github.com/go-slides/slides/export"
	"github.com/go-slides/slides/project"
	"github.com/go-slides/slides/vfs"
	"github.com/go-webengine/engine"
)

// MaxSource bounds the deck, as the playground's renderer does.
const MaxSource = 1 << 20

// Formats are the documents an export makes, with their file name and
// media type.
var Formats = map[string]struct{ Name, MIME string }{
	"pdf":  {"deck.pdf", "application/pdf"},
	"pptx": {"deck.pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
	"odp":  {"deck.odp", "application/vnd.oasis.opendocument.presentation"},
}

// Result is an exported document.
type Result struct {
	Data       []byte
	Name, MIME string
	Slides     int
}

// Project is a deck with its project, as the preview renders it: the
// deck's text at DeckPath, the project's text files (slides.yaml, themes,
// kinds, data) by their paths beside it, its folders, and its media by
// their paths relative to the deck's folder (media/photo.png).
type Project struct {
	DeckPath string
	Source   string
	Files    map[string]string
	Dirs     []string
	Media    map[string][]byte
}

// Export makes the deck's document in format ("pdf", "pptx" or "odp"), as
// `slides --pdf` (vector, with the outline), `slides --pptx` and `slides
// --odp` (editable, drawn at scale 2) do with their defaults, over the
// project the preview renders: its slides.yaml, themes, kinds and data
// files, and its media, which go into the slides as data: URIs (the
// preview's frame makes blob: URLs of the same bytes). progress, if not
// nil, is told each slide done.
func Export(ctx context.Context, p Project, format string, progress func(done, total int)) (Result, error) {
	f, ok := Formats[format]
	if !ok {
		return Result{}, fmt.Errorf("cannot export %q: want pdf, pptx or odp", format)
	}
	if len(p.Source) > MaxSource {
		return Result{}, errors.New("the deck is too large to export here: use the slides command")
	}
	if p.DeckPath == "" {
		p.DeckPath = "/project/deck.md"
	}
	listed, err := mediafiles.Check(p.DeckPath, p.Files, p.Dirs, p.Media)
	if err != nil {
		return Result{}, err
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
		return Result{}, err
	}
	cfg.HTML = deck.HTMLAllowlist
	b, err := slides.Render([]byte(p.Source), cfg, deck.FrameBox)
	if err != nil {
		return Result{}, err
	}
	if len(p.Media) > 0 {
		b.Packed.CSS = mediafiles.Inline(b.Packed.CSS, p.Media)
		for i := range b.Slides {
			b.Slides[i].HTML = mediafiles.Inline(b.Slides[i].HTML, p.Media)
		}
	}
	// One slide at a time: a browser worker has one thread, and one
	// slide's layout is what it holds.
	opts := export.Options{Workers: 1, Scale: 1}
	if format != "pdf" {
		opts.Scale = 2 // the command's default for a deck shown full screen
	}
	done := 0
	step := func() {
		done++
		if progress != nil {
			progress(done, len(b.Slides))
		}
	}
	var add func(*engine.PageLayout) error
	var write func(*bytes.Buffer) error
	switch format {
	case "pdf":
		w := export.NewPDFWith(b, export.DefaultPDFOptions)
		add, write = w.AddLayout, func(buf *bytes.Buffer) error { return w.Write(buf) }
	case "pptx":
		w := export.NewPPTX(b)
		add, write = w.AddLayout, func(buf *bytes.Buffer) error { return w.Write(buf) }
	default:
		w := export.NewODP(b)
		add, write = w.AddLayout, func(buf *bytes.Buffer) error { return w.Write(buf) }
	}
	if err := export.EachLayout(ctx, b, opts, func(_ int, l *engine.PageLayout) error {
		if err := add(l); err != nil {
			return err
		}
		step()
		return nil
	}); err != nil {
		return Result{}, err
	}
	var buf bytes.Buffer
	if err := write(&buf); err != nil {
		return Result{}, err
	}
	return Result{Data: buf.Bytes(), Name: f.Name, MIME: f.MIME, Slides: len(b.Slides)}, nil
}
