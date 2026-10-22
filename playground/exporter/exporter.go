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

// Export makes the deck's document in format ("pdf", "pptx" or "odp"), as
// `slides --pdf` (vector, with the outline), `slides --pptx` and `slides
// --odp` (editable, drawn at scale 2) do with their defaults. progress, if
// not nil, is told each slide done.
func Export(ctx context.Context, src, format string, progress func(done, total int)) (Result, error) {
	f, ok := Formats[format]
	if !ok {
		return Result{}, fmt.Errorf("cannot export %q: want pdf, pptx or odp", format)
	}
	if len(src) > MaxSource {
		return Result{}, errors.New("the deck is too large to export here: use the slides command")
	}
	const path = "deck.md"
	cfg, err := project.Load(path, project.Options{FS: vfs.NewMap(map[string][]byte{path: []byte(src)}), NoConfig: true})
	if err != nil {
		return Result{}, err
	}
	cfg.HTML = deck.HTMLAllowlist
	b, err := slides.Render([]byte(src), cfg, deck.FrameBox)
	if err != nil {
		return Result{}, err
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
