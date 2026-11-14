// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"errors"
	"syscall/js"
	"time"
)

// videoHub is the page's side of video support: video-worker.wasm (go-avkit's
// container reader and H.264 decoder), downloaded and started the first time a
// render worker asks for it, never for a deck without video, and linked to
// each render worker (the preview's, the export's) by a MessageChannel, so
// that the file goes from one worker to the other and never through this page.
//
// The module is fetched from the page's own origin (goSlidesVideoURL, versioned
// by the deploy like the export engine's, so that the browser's cache and the
// playground's service worker keep it): the Content Security Policy's
// connect-src 'self' and worker-src are enough. One worker serves the page; it
// is stopped when it stops answering (videoStuck) and the next ask starts another.
type videoHub struct {
	worker  js.Value
	module  js.Value
	loading bool
	ready   bool
	waiting []func(error)
	status  func(string)
	onMsg   js.Func
}

// videoRenderTimeout bounds a render that carries videos, which includes
// the poster's decoding (600 samples at most) in the video worker.
const videoRenderTimeout = 30 * time.Second

var theHub = newVideoHub()

func newVideoHub() *videoHub {
	h := &videoHub{}
	h.onMsg = js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 && args[0].Get("data").Type() == js.TypeString && args[0].Get("data").String() == `{"ready":true}` {
			h.ready, h.loading = true, false
			h.flush(nil)
		}
		return nil
	})
	return h
}

// isVideoMessage reports whether a worker's string message is about video, and
// handles it. from is the worker that sent it; done is called when a link to
// it is made (with the error if not), to re-arm the timer the wait paused.
func (h *videoHub) handle(s string, from js.Value, done func(error)) bool {
	switch s {
	case msgNeedVideo:
		h.link(from, done)
		return true
	case msgVideoStuck:
		h.stop()
		return true
	}
	return false
}

// link makes the video worker available (loading it if need be) and hands a
// port of a new channel to it and the other to target.
func (h *videoHub) link(target js.Value, done func(error)) {
	h.waiting = append(h.waiting, func(err error) {
		if err != nil {
			target.Call("postMessage", map[string]any{"videoPort": nil, "error": err.Error()})
		} else {
			ch := js.Global().Get("MessageChannel").New()
			h.worker.Call("postMessage", map[string]any{"videoPort": ch.Get("port2")}, []any{ch.Get("port2")})
			target.Call("postMessage", map[string]any{"videoPort": ch.Get("port1")}, []any{ch.Get("port1")})
		}
		if done != nil {
			done(err)
		}
	})
	if h.ready {
		h.flush(nil)
		return
	}
	if !h.loading {
		h.load()
	}
}

func (h *videoHub) flush(err error) {
	w := h.waiting
	h.waiting = nil
	for _, f := range w {
		f(err)
	}
}

// stop ends the video worker: it stopped answering.
func (h *videoHub) stop() {
	if h.worker.Truthy() {
		h.worker.Call("terminate")
	}
	h.worker, h.ready, h.loading = js.Value{}, false, false
}

// load downloads and compiles the module, then starts it in a worker (worker.js,
// as the renderer and the export engine).
func (h *videoHub) load() {
	h.loading = true
	if h.module.Truthy() {
		h.start(h.module)
		return
	}
	url := js.Global().Get("goSlidesVideoURL")
	if !url.Truthy() {
		h.loading = false
		h.flush(errors.New("this page has no video reader"))
		return
	}
	if h.status != nil {
		h.status("Loading video support…")
	}
	var onResp, onCompiled, onErr js.Func
	release := func() {
		for _, f := range []js.Func{onResp, onCompiled, onErr} {
			f.Release()
		}
	}
	onErr = js.FuncOf(func(_ js.Value, args []js.Value) any {
		msg := "the video reader could not be loaded"
		if len(args) > 0 && args[0].Truthy() {
			msg += ": " + args[0].Call("toString").String()
		}
		release()
		h.loading = false
		h.flush(errors.New(msg))
		return nil
	})
	onCompiled = js.FuncOf(func(_ js.Value, args []js.Value) any {
		release()
		h.module = args[0]
		h.start(args[0])
		return nil
	})
	onResp = js.FuncOf(func(_ js.Value, args []js.Value) any {
		resp := args[0]
		if !resp.Get("ok").Bool() {
			release()
			h.loading = false
			h.flush(errors.New("the video reader could not be loaded: HTTP " + resp.Get("status").Call("toString").String()))
			return nil
		}
		wasm := js.Global().Get("Response").New(resp.Get("body"), map[string]any{"headers": map[string]any{"Content-Type": "application/wasm"}})
		js.Global().Get("WebAssembly").Call("compileStreaming", wasm).Call("then", onCompiled, onErr)
		return nil
	})
	js.Global().Call("fetch", url).Call("then", onResp, onErr)
}

func (h *videoHub) start(mod js.Value) {
	h.worker = js.Global().Get("Worker").New(js.Global().Get("goSlidesWorkerURL"))
	h.worker.Call("addEventListener", "message", h.onMsg)
	h.worker.Call("postMessage", map[string]any{"module": mod})
}
