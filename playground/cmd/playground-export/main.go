// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

// Command playground-export is the playground's export worker: PDF,
// PowerPoint and OpenDocument, made by go-slides' export engine in a Web
// Worker. The page downloads it the first time an export is asked for,
// compiles it, and runs it through worker.js like the render worker.
//
// One JSON request in ({id, source, format}); messages out: {id, done,
// total} as slides are laid out, then {id, name, mime, slides, millis} with
// the document's bytes in data (transferred, not copied), or {id, error}.
package main

import (
	"context"
	"encoding/json"
	"syscall/js"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/exporter"
	"github.com/go-slides/go-slides.github.io/playground/mediafiles"
)

type request struct {
	ID       int               `json:"id"`
	Source   string            `json:"source"`
	Format   string            `json:"format"`
	DeckPath string            `json:"deckPath"`
	Files    map[string]string `json:"files"`
	Dirs     []string          `json:"dirs"`
	media    map[string][]byte
}

func main() {
	js.Global().Set("onmessage", js.FuncOf(func(_ js.Value, args []js.Value) any {
		// {req: JSON, media: {path: {type, bytes}}}: the project's text in
		// the JSON, its media as the ArrayBuffers the preview is posted
		var req request
		if len(args) == 0 {
			return nil
		}
		d := args[0].Get("data")
		if d.Type() != js.TypeObject || d.Get("req").Type() != js.TypeString ||
			json.Unmarshal([]byte(d.Get("req").String()), &req) != nil {
			return nil
		}
		req.media = readMedia(d.Get("media"))
		// The export runs on its own goroutine: the handler returns, and the
		// progress messages go out as the slides are laid out.
		go serve(req)
		return nil
	}))
	js.Global().Call("postMessage", map[string]any{"ready": true})
	select {}
}

// readMedia copies the media posted, within the bound on their total
// (exporter.Export checks it again, with the paths).
func readMedia(m js.Value) map[string][]byte {
	out := map[string][]byte{}
	if m.Type() != js.TypeObject {
		return out
	}
	keys := js.Global().Get("Object").Call("keys", m)
	total := 0
	for i := range keys.Length() {
		k := keys.Index(i).String()
		b := m.Get(k).Get("bytes")
		if b.Type() != js.TypeObject || !b.InstanceOf(js.Global().Get("ArrayBuffer")) {
			continue
		}
		n := b.Get("byteLength").Int()
		if total += n; total > mediafiles.MaxMedia {
			break
		}
		buf := make([]byte, n)
		js.CopyBytesToGo(buf, js.Global().Get("Uint8Array").New(b))
		out[k] = buf
	}
	return out
}

func serve(req request) {
	start := time.Now()
	post := js.Global().Get("postMessage")
	p := exporter.Project{DeckPath: req.DeckPath, Source: req.Source, Files: req.Files, Dirs: req.Dirs, Media: req.media}
	r, err := exporter.Export(context.Background(), p, req.Format, func(done, total int) {
		post.Invoke(map[string]any{"id": req.ID, "done": done, "total": total})
	})
	if err != nil {
		post.Invoke(map[string]any{"id": req.ID, "error": err.Error()})
		return
	}
	data := js.Global().Get("Uint8Array").New(len(r.Data))
	js.CopyBytesToJS(data, r.Data)
	post.Invoke(map[string]any{
		"id": req.ID, "name": r.Name, "mime": r.MIME, "slides": r.Slides,
		"millis": time.Since(start).Milliseconds(), "data": data,
	}, []any{data.Get("buffer")})
}
