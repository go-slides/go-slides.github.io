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
)

// exporter runs go-slides' export engine (cmd/playground-export) in a
// worker of its own. The module is downloaded the first time an export is
// asked for (the page's goSlidesExportURL, versioned by the deploy, so the
// browser's cache keeps it), compiled once, and the worker kept for the
// next exports of the page.
type exporter struct {
	vm      *playground.ViewModel
	worker  js.Value
	ready   bool
	loading bool
	queued  string // the request waiting for the worker
	next    int
	timer   js.Value
	onMsg   js.Func
	onTime  js.Func
}

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
	return e
}

// Export implements playground.Exporter (through host).
func (e *exporter) Export(format, source string) {
	e.next++
	b, _ := json.Marshal(map[string]any{"id": e.next, "source": source, "format": format})
	e.queued = string(b)
	switch {
	case e.ready:
		e.flush()
	case !e.loading:
		e.load()
	}
}

func (e *exporter) flush() {
	if e.queued == "" {
		return
	}
	e.worker.Call("postMessage", e.queued)
	e.queued = ""
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
	e.worker, e.ready, e.loading, e.queued = js.Value{}, false, false, ""
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
