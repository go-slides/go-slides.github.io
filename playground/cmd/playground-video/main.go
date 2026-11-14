// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

// Command playground-video is the playground's video worker: go-avkit's
// container reader and H.264 decoder in a WebAssembly module of their own
// (video-worker.wasm), run in a Web Worker the first time a deck needs a
// poster. The page's module links none of it.
//
// The page links this worker to each worker that renders a deck with a
// MessageChannel: it posts {videoPort: port} here, and the requests of that
// renderer (package videoproto) arrive on the port as {json, data}: the
// request's JSON and, in data, the file. The answer goes back the same way:
// {json, data} with the poster's picture in data. Nothing is fetched here.
package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/go-slides/go-slides.github.io/playground/videoproto"
	"github.com/go-slides/go-slides.github.io/playground/videoproto/serve"
	"github.com/go-slides/slides/video"
)

func main() {
	srv := &serve.Server{}
	js.Global().Set("onmessage", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		port := args[0].Get("data").Get("videoPort")
		if !port.Truthy() {
			return nil
		}
		port.Call("addEventListener", "message", js.FuncOf(func(_ js.Value, a []js.Value) any {
			if len(a) > 0 {
				answer(srv, port, a[0].Get("data"))
			}
			return nil
		}))
		port.Call("start")
		return nil
	}))
	js.Global().Call("postMessage", `{"ready":true}`)
	select {}
}

// answer reads one request off a port and posts its answer.
func answer(srv *serve.Server, port, msg js.Value) {
	var req videoproto.Request
	if msg.Type() != js.TypeObject || msg.Get("json").Type() != js.TypeString || json.Unmarshal([]byte(msg.Get("json").String()), &req) != nil {
		return
	}
	var payload []byte
	if d := msg.Get("data"); d.Truthy() {
		// the file is bounded before it is copied into this module's memory
		n := d.Get("byteLength").Int()
		if n > video.DefaultLimits.MaxBytes {
			post(port, videoproto.Response{ID: req.ID, Code: video.CodeLimit, Error: "the file is larger than the 64 MiB read"}, nil)
			return
		}
		payload = make([]byte, n)
		js.CopyBytesToGo(payload, js.Global().Get("Uint8Array").New(d))
	}
	res, img := srv.Handle(req, payload)
	post(port, res, img)
}

func post(port js.Value, res videoproto.Response, img []byte) {
	b, _ := json.Marshal(res)
	m := map[string]any{"json": string(b)}
	if len(img) > 0 {
		buf := js.Global().Get("Uint8Array").New(len(img))
		js.CopyBytesToJS(buf, img)
		m["data"] = buf.Get("buffer")
		port.Call("postMessage", m, []any{buf.Get("buffer")})
		return
	}
	port.Call("postMessage", m)
}
