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
)

type request struct {
	ID     int    `json:"id"`
	Source string `json:"source"`
	Format string `json:"format"`
}

func main() {
	js.Global().Set("onmessage", js.FuncOf(func(_ js.Value, args []js.Value) any {
		var req request
		if len(args) == 0 || args[0].Get("data").Type() != js.TypeString ||
			json.Unmarshal([]byte(args[0].Get("data").String()), &req) != nil {
			return nil
		}
		// The export runs on its own goroutine: the handler returns, and the
		// progress messages go out as the slides are laid out.
		go serve(req)
		return nil
	}))
	js.Global().Call("postMessage", map[string]any{"ready": true})
	select {}
}

func serve(req request) {
	start := time.Now()
	post := js.Global().Get("postMessage")
	r, err := exporter.Export(context.Background(), req.Source, req.Format, func(done, total int) {
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
