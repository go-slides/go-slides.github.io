// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"syscall/js"
	"time"

	playground "github.com/go-slides/go-slides.github.io/playground"
	"github.com/go-slides/go-slides.github.io/playground/exportdoc"
)

// exporter runs go-slides' export engine (cmd/playground-export) in a
// worker of its own. The module is downloaded the first time an export is
// asked for (the page's goSlidesExportURL, versioned by the deploy, so the
// browser's cache keeps it), compiled once, and the worker kept for the
// next exports of the page.
//
// The engine carries no renderer: the deck is rendered for it (a go-slides
// doc.Doc, playground/exportdoc) by this page's own module in a worker of
// its own, while the engine downloads, and handed over as JSON.
type exporter struct {
	vm      *playground.ViewModel
	worker  js.Value
	ready   bool
	loading bool
	pending *exportJob // the export waiting for its deck or for the worker
	next    int
	timer   js.Value
	onMsg   js.Func
	onTime  js.Func

	// the worker that renders the deck for an export (the preview's module)
	docW     js.Value
	docReady bool
	docTimer js.Value
	onDoc    js.Func
	onDocTO  js.Func
}

// exportJob is one export: its request, its media, and its deck once
// rendered.
type exportJob struct {
	id         int
	format     string
	docReq     string // the render request, until the doc worker is ready
	doc        json.RawMessage
	media      map[string]any
	fontsURL   string
	fontsQuery string
}

// docTimeout bounds the deck's render for an export, as the preview's is
// bounded (renderTimeout), with room for a long deck.
const docTimeout = 10 * time.Second

// newExporter makes the exporter; its view-model is set once made (the
// host, which holds the exporter, is handed to the view-model first).
func newExporter() *exporter {
	e := &exporter{}
	e.onMsg = js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			e.message(args[0].Get("data"))
		}
		return nil
	})
	e.onTime = js.FuncOf(func(js.Value, []js.Value) any {
		e.fail(errors.New("it took longer than " + playground.ExportTimeout.String() + "; the export engine was stopped"))
		return nil
	})
	e.onDoc = js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			e.docMessage(args[0].Get("data"))
		}
		return nil
	})
	e.onDocTO = js.FuncOf(func(js.Value, []js.Value) any {
		e.dropDocWorker()
		e.fail(errors.New("rendering the deck took more than " + docTimeout.String() + ": is a template looping?"))
		return nil
	})
	return e
}

// Export implements playground.Exporter (through host). The deck is
// rendered from the project the preview renders (its text files), and the
// media go to the engine as the same ArrayBuffers the preview frames are
// posted (files.mediaMessage), so the export is what the preview shows.
func (e *exporter) Export(format, source string) {
	e.next++
	deckPath, texts, dirs := e.vm.ProjectText()
	job := &exportJob{id: e.next, format: format, media: files.mediaMessage()}
	// where the export engine fetches the fonts a deck draws (cmd/build
	// writes them there; the engine itself carries none)
	if u := js.Global().Get("goSlidesFontsURL"); u.Type() == js.TypeString {
		job.fontsURL = u.String()
	}
	if q := js.Global().Get("goSlidesFontsQuery"); q.Type() == js.TypeString {
		job.fontsQuery = q.String()
	}
	b, _ := json.Marshal(docRequest{ExportDoc: true, ID: job.id,
		Project: exportdoc.Project{DeckPath: deckPath, Source: source, Files: texts, Dirs: dirs}})
	job.docReq = string(b)
	e.pending = job
	e.renderDoc()
	if !e.ready && !e.loading {
		e.load()
	}
}

// renderDoc asks the doc worker for the pending export's deck, starting the
// worker the first time.
func (e *exporter) renderDoc() {
	if !e.docW.Truthy() {
		e.docW = js.Global().Get("Worker").New(js.Global().Get("goSlidesWorkerURL"))
		e.docW.Call("addEventListener", "message", e.onDoc)
		// the page compiled its module; this worker instantiates the same one
		e.docW.Call("postMessage", map[string]any{"module": js.Global().Get("goSlidesModule")})
	}
	if !e.docReady || e.pending == nil || e.pending.docReq == "" {
		return
	}
	e.docW.Call("postMessage", e.pending.docReq)
	e.pending.docReq = ""
	e.docTimer = js.Global().Call("setTimeout", e.onDocTO, docTimeout.Milliseconds())
}

// docMessage takes the doc worker's answer: ready, or a deck rendered.
func (e *exporter) docMessage(data js.Value) {
	if data.Type() != js.TypeString {
		return
	}
	s := data.String()
	if s == `{"ready":true}` {
		e.docReady = true
		e.renderDoc()
		return
	}
	var r docResponse
	if json.Unmarshal([]byte(s), &r) != nil || e.pending == nil || r.ID != e.pending.id {
		return
	}
	if e.docTimer.Truthy() {
		js.Global().Call("clearTimeout", e.docTimer)
	}
	if r.Error != "" {
		e.pending = nil
		e.vm.Exported("", 0, 0, errors.New(r.Error))
		return
	}
	e.pending.doc = r.Doc
	e.flush()
}

// dropDocWorker stops the doc worker; the next export starts another.
func (e *exporter) dropDocWorker() {
	if e.docW.Truthy() {
		e.docW.Call("terminate")
	}
	e.docW, e.docReady = js.Value{}, false
}

// flush hands the pending export to the engine once it has both its deck
// and a worker.
func (e *exporter) flush() {
	j := e.pending
	if j == nil || j.doc == nil || !e.ready {
		return
	}
	b, _ := json.Marshal(map[string]any{"id": j.id, "format": j.format, "doc": j.doc,
		"fontsURL": j.fontsURL, "fontsQuery": j.fontsQuery})
	e.worker.Call("postMessage", map[string]any{"req": string(b), "media": j.media})
	e.pending = nil
	e.timer = js.Global().Call("setTimeout", e.onTime, playground.ExportTimeout.Milliseconds())
}

// fail ends the export in progress, and drops the worker: the next export
// starts a new one from the compiled module.
func (e *exporter) fail(err error) {
	if e.timer.Truthy() {
		js.Global().Call("clearTimeout", e.timer)
	}
	if e.worker.Truthy() {
		e.worker.Call("terminate")
	}
	e.worker, e.ready, e.loading, e.pending = js.Value{}, false, false, nil
	e.vm.Exported("", 0, 0, err)
}

// load fetches the module with progress, compiles it and starts the worker.
func (e *exporter) load() {
	e.loading = true
	if mod := js.Global().Get("goSlidesExportModule"); mod.Truthy() {
		e.start(mod)
		return
	}
	url := js.Global().Get("goSlidesExportURL")
	if !url.Truthy() {
		e.fail(errors.New("this page has no export engine"))
		return
	}
	total := int64(0)
	if n := js.Global().Get("goSlidesExportBytes"); n.Type() == js.TypeNumber {
		total = int64(n.Float())
	}
	if streams() {
		e.loadStreaming(url, total)
		return
	}
	var chunks []js.Value
	var got int64
	var reader js.Value
	var onRead, onResp, onErr, onCompiled js.Func
	release := func() {
		for _, f := range []js.Func{onRead, onResp, onErr, onCompiled} {
			f.Release()
		}
	}
	onErr = js.FuncOf(func(_ js.Value, args []js.Value) any {
		msg := "the export engine could not be loaded"
		if len(args) > 0 && args[0].Truthy() {
			msg += ": " + args[0].Call("toString").String()
		}
		release()
		e.fail(errors.New(msg))
		return nil
	})
	onCompiled = js.FuncOf(func(_ js.Value, args []js.Value) any {
		release()
		js.Global().Set("goSlidesExportModule", args[0])
		e.start(args[0])
		return nil
	})
	onRead = js.FuncOf(func(_ js.Value, args []js.Value) any {
		r := args[0]
		if r.Get("done").Bool() {
			buf := js.Global().Get("Uint8Array").New(got)
			pos := 0
			for _, c := range chunks {
				buf.Call("set", c, pos)
				pos += c.Get("length").Int()
			}
			chunks = nil
			e.vm.Status.Set("Compiling the export engine…")
			js.Global().Get("WebAssembly").Call("compile", buf.Get("buffer")).Call("then", onCompiled, onErr)
			return nil
		}
		v := r.Get("value")
		chunks = append(chunks, v)
		got += int64(v.Get("length").Int())
		e.vm.ExportLoading(got, total)
		reader.Call("read").Call("then", onRead, onErr)
		return nil
	})
	onResp = js.FuncOf(func(_ js.Value, args []js.Value) any {
		resp := args[0]
		if !resp.Get("ok").Bool() {
			release()
			e.fail(errors.New("the export engine could not be loaded: HTTP " + resp.Get("status").Call("toString").String()))
			return nil
		}
		if total == 0 {
			if cl := resp.Get("headers").Call("get", "Content-Length"); cl.Type() == js.TypeString {
				total = int64(js.Global().Call("parseInt", cl, 10).Float())
			}
		}
		reader = resp.Get("body").Call("getReader")
		reader.Call("read").Call("then", onRead, onErr)
		return nil
	})
	js.Global().Call("fetch", url).Call("then", onResp, onErr)
}

// streams reports whether the browser compiles a module as it downloads
// (WebAssembly.compileStreaming) and can count the bytes on the way
// (TransformStream).
func streams() bool {
	return js.Global().Get("WebAssembly").Get("compileStreaming").Type() == js.TypeFunction &&
		js.Global().Get("TransformStream").Type() == js.TypeFunction &&
		js.Global().Get("Response").Type() == js.TypeFunction
}

// loadStreaming downloads the module and compiles it as it arrives
// (WebAssembly.compileStreaming), instead of after the last byte. The
// bytes are counted on their way to the compiler for the progress.
func (e *exporter) loadStreaming(url js.Value, total int64) {
	var got int64
	var onResp, onErr, onCompiled, onChunk, onFlush js.Func
	release := func() {
		for _, f := range []js.Func{onResp, onErr, onCompiled, onChunk, onFlush} {
			f.Release()
		}
	}
	onErr = js.FuncOf(func(_ js.Value, args []js.Value) any {
		msg := "the export engine could not be loaded"
		if len(args) > 0 && args[0].Truthy() {
			msg += ": " + args[0].Call("toString").String()
		}
		release()
		e.fail(errors.New(msg))
		return nil
	})
	onCompiled = js.FuncOf(func(_ js.Value, args []js.Value) any {
		release()
		js.Global().Set("goSlidesExportModule", args[0])
		e.start(args[0])
		return nil
	})
	// transform(chunk, controller): count, pass on
	onChunk = js.FuncOf(func(_ js.Value, args []js.Value) any {
		got += int64(args[0].Get("length").Int())
		e.vm.ExportLoading(got, total)
		args[1].Call("enqueue", args[0])
		return nil
	})
	onFlush = js.FuncOf(func(js.Value, []js.Value) any {
		e.vm.Status.Set("Compiling the export engine…")
		return nil
	})
	onResp = js.FuncOf(func(_ js.Value, args []js.Value) any {
		resp := args[0]
		if !resp.Get("ok").Bool() {
			release()
			e.fail(errors.New("the export engine could not be loaded: HTTP " + resp.Get("status").Call("toString").String()))
			return nil
		}
		if total == 0 {
			if cl := resp.Get("headers").Call("get", "Content-Length"); cl.Type() == js.TypeString {
				total = int64(js.Global().Call("parseInt", cl, 10).Float())
			}
		}
		counter := js.Global().Get("TransformStream").New(map[string]any{"transform": onChunk, "flush": onFlush})
		body := resp.Get("body").Call("pipeThrough", counter)
		wasm := js.Global().Get("Response").New(body, map[string]any{"headers": map[string]any{"Content-Type": "application/wasm"}})
		js.Global().Get("WebAssembly").Call("compileStreaming", wasm).Call("then", onCompiled, onErr)
		return nil
	})
	js.Global().Call("fetch", url).Call("then", onResp, onErr)
}

// start runs the compiled module in a worker (worker.js, as the renderer).
func (e *exporter) start(mod js.Value) {
	e.worker = js.Global().Get("Worker").New(js.Global().Get("goSlidesWorkerURL"))
	e.worker.Call("addEventListener", "message", e.onMsg)
	e.worker.Call("postMessage", map[string]any{"module": mod})
}

func (e *exporter) message(d js.Value) {
	if d.Type() != js.TypeObject {
		return
	}
	if d.Get("ready").Truthy() {
		e.ready, e.loading = true, false
		e.flush()
		return
	}
	if d.Get("id").Type() != js.TypeNumber || d.Get("id").Int() != e.next {
		return
	}
	if dn := d.Get("done"); dn.Type() == js.TypeNumber {
		e.vm.ExportProgress(dn.Int(), d.Get("total").Int())
		return
	}
	if e.timer.Truthy() {
		js.Global().Call("clearTimeout", e.timer)
	}
	if msg := d.Get("error"); msg.Type() == js.TypeString {
		e.vm.Exported("", 0, 0, errors.New(msg.String()))
		return
	}
	data := d.Get("data")
	name, mime := d.Get("name").String(), d.Get("mime").String()
	blob := js.Global().Get("Blob").New([]any{data}, map[string]any{"type": mime})
	url := js.Global().Get("URL").Call("createObjectURL", blob)
	a := js.Global().Get("document").Call("createElement", "a")
	a.Set("href", url)
	a.Set("download", name)
	a.Call("click")
	// revoked once the download has taken it
	js.Global().Call("setTimeout", js.FuncOf(func(js.Value, []js.Value) any {
		js.Global().Get("URL").Call("revokeObjectURL", url)
		return nil
	}), (10 * time.Second).Milliseconds())
	e.vm.Exported(name, data.Get("length").Int(), d.Get("slides").Int(), nil)
}
