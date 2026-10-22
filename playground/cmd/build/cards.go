package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-slides/slides"
	"github.com/go-slides/slides/deck"
	"github.com/go-slides/slides/export"
	"github.com/go-slides/slides/project"
	"github.com/go-slides/slides/vfs"
)

// cardScale renders a 1280×720 slide as 480×270: the carousel shows it at
// 320 CSS pixels wide, sharp on a 1.5× screen, a few kilobytes as JPEG.
const cardScale = 0.375

// renderCards renders each slide of the cards deck to out/<n>.jpg (1-based),
// as `slides --images jpeg` would, and returns how many it wrote and their
// total size. The pictures are go-slides' own output: the carousel shows
// what the playground makes.
func renderCards(ctx context.Context, deckPath, out string) (n int, size int64, err error) {
	src, err := os.ReadFile(deckPath)
	if err != nil {
		return 0, 0, err
	}
	name := filepath.Base(deckPath)
	cfg, err := project.Load(name, project.Options{FS: vfs.NewMap(map[string][]byte{name: src}), NoConfig: true})
	if err != nil {
		return 0, 0, err
	}
	cfg.HTML = deck.HTMLAllowlist
	b, err := slides.Render(src, cfg, deck.FrameBox)
	if err != nil {
		return 0, 0, err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return 0, 0, err
	}
	opts := export.Options{Scale: cardScale, JPEGQuality: 78}
	err = export.Each(ctx, b, opts, func(i int, img *image.RGBA) error {
		var buf bytes.Buffer
		if err := export.JPEG(&buf, img, opts); err != nil {
			return err
		}
		n++
		size += int64(buf.Len())
		return os.WriteFile(filepath.Join(out, strconv.Itoa(i+1)+".jpg"), buf.Bytes(), 0o644)
	})
	if err != nil {
		return n, size, fmt.Errorf("rendering %s: %w", deckPath, err)
	}
	return n, size, nil
}
