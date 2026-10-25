// Package exporter is the playground's PDF, PowerPoint and OpenDocument
// export: the slides command's own path (cmd/slides exportBytes) with its
// defaults, run in a Web Worker by a module of its own
// (cmd/playground-export) that the page downloads the first time an export
// is asked for. It takes the deck already rendered (a go-slides doc.Doc,
// which playground/exportdoc makes in the preview's worker), so that the
// module carries go-webengine and the writers, not a renderer.
package exporter

import (
	"bytes"
	"context"
	"fmt"

	"github.com/go-slides/go-slides.github.io/playground/mediafiles"
	"github.com/go-slides/slides/doc"
	"github.com/go-slides/slides/export"
	"github.com/go-webengine/engine"
)

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

// Input is what an export needs besides the format.
type Input struct {
	// Doc is the deck, rendered (exportdoc.Build).
	Doc *doc.Doc
	// Media are the project's media by their paths relative to the deck's
	// folder (media/photo.png): they go into the slides as data: URIs (the
	// preview's frame makes blob: URLs of the same bytes).
	Media map[string][]byte
	// Fetcher, when set, fetches the slides' https: images (and style
	// sheets, fonts) the way the host can: in the browser, its fetch,
	// since the export module carries no HTTP client.
	Fetcher func(ctx context.Context, url string) ([]byte, string, error)
}

// Export makes the deck's document in format ("pdf", "pptx" or "odp"), as
// `slides --pdf` (vector, with the outline), `slides --pptx` and `slides
// --odp` (editable, drawn at scale 2) do with their defaults. progress, if
// not nil, is told each slide done.
func Export(ctx context.Context, in Input, format string, progress func(done, total int)) (Result, error) {
	f, ok := Formats[format]
	if !ok {
		return Result{}, fmt.Errorf("cannot export %q: want pdf, pptx or odp", format)
	}
	if in.Doc == nil {
		return Result{}, fmt.Errorf("no deck to export")
	}
	if _, err := mediafiles.Check("/project/deck.md", nil, nil, in.Media); err != nil {
		return Result{}, err
	}
	d := *in.Doc
	if len(in.Media) > 0 {
		d.CSS = mediafiles.Inline(d.CSS, in.Media)
		d.Slides = append([]doc.Slide(nil), d.Slides...)
		for i := range d.Slides {
			d.Slides[i].HTML = mediafiles.Inline(d.Slides[i].HTML, in.Media)
		}
	}
	// One slide at a time: a browser worker has one thread, and one
	// slide's layout is what it holds.
	opts := export.Options{Workers: 1, Scale: 1, Fetcher: in.Fetcher}
	if format != "pdf" {
		opts.Scale = 2 // the command's default for a deck shown full screen
	}
	done := 0
	step := func() {
		done++
		if progress != nil {
			progress(done, len(d.Slides))
		}
	}
	var add func(*engine.PageLayout) error
	var write func(*bytes.Buffer) error
	switch format {
	case "pdf":
		w := export.NewPDFWith(&d, export.DefaultPDFOptions)
		add, write = w.AddLayout, func(buf *bytes.Buffer) error { return w.Write(buf) }
	case "pptx":
		w := export.NewPPTX(&d)
		add, write = w.AddLayout, func(buf *bytes.Buffer) error { return w.Write(buf) }
	default:
		w := export.NewODP(&d)
		add, write = w.AddLayout, func(buf *bytes.Buffer) error { return w.Write(buf) }
	}
	if err := export.EachLayout(ctx, &d, opts, func(_ int, l *engine.PageLayout) error {
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
	return Result{Data: buf.Bytes(), Name: f.Name, MIME: f.MIME, Slides: len(d.Slides)}, nil
}
