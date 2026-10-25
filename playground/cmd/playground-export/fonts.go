// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"context"
	"errors"
	"path"
	"strings"
	"sync"
	"syscall/js"
)

// The module is built without the fonts and without an HTTP client
// (ExportTags). cmd/build writes the fonts under the site's
// /playground/fonts/, the same bytes the slides command embeds, and a face
// is fetched from there the first time a slide draws it: the welcome deck
// downloads Inter Regular and Go Mono, not the 21 other faces. A deck's
// https: images are fetched by the browser too (go-webengine's
// Engine.Fetcher), as the preview's are.

var (
	fontsMu    sync.Mutex
	fontsBase  string // the fonts' URL, with a trailing slash; set by the first request
	fontsQuery string // the deploy's version, "?v=…", so that a deploy's fonts are its own
)

// setFontsBase records where the fonts are, from the page's request.
func setFontsBase(u, query string) {
	if u == "" {
		return
	}
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	fontsMu.Lock()
	fontsBase, fontsQuery = u, query
	fontsMu.Unlock()
}

// fetchFont fetches one of the fonts by its name ("inter/inter.ttf"). A
// name is checked against the lists the engine asks from, so the page's URL
// is the only part that varies.
func fetchFont(name string) ([]byte, error) {
	if !known[name] || path.Clean(name) != name {
		return nil, errors.New("not a font of the export: " + name)
	}
	fontsMu.Lock()
	base, query := fontsBase, fontsQuery
	fontsMu.Unlock()
	if base == "" {
		return nil, errors.New("the page gave no fonts URL")
	}
	b, _, err := browserFetch(context.Background(), base+name+query)
	if err != nil {
		return nil, errors.New("the font " + name + " could not be loaded: " + err.Error())
	}
	return b, nil
}

// browserFetch is the browser's fetch of url: its body and Content-Type.
// It blocks its goroutine, the export's, while the browser fetches: the
// worker's event loop runs meanwhile. A cross-origin URL is fetched as the
// page's own fetch would be (CORS applies), with no referrer.
func browserFetch(ctx context.Context, url string) ([]byte, string, error) {
	type result struct {
		b     []byte
		ctype string
		err   error
	}
	done := make(chan result, 1)
	var ctype string
	var onResp, onBuf, onErr js.Func
	onErr = js.FuncOf(func(_ js.Value, args []js.Value) any {
		msg := "fetch failed"
		if len(args) > 0 && args[0].Truthy() {
			msg = args[0].Call("toString").String()
		}
		done <- result{err: errors.New(msg)}
		return nil
	})
	onBuf = js.FuncOf(func(_ js.Value, args []js.Value) any {
		u8 := js.Global().Get("Uint8Array").New(args[0])
		b := make([]byte, u8.Get("length").Int())
		js.CopyBytesToGo(b, u8)
		done <- result{b: b, ctype: ctype}
		return nil
	})
	onResp = js.FuncOf(func(_ js.Value, args []js.Value) any {
		resp := args[0]
		if !resp.Get("ok").Bool() {
			done <- result{err: errors.New("HTTP " + resp.Get("status").Call("toString").String())}
			return nil
		}
		if ct := resp.Get("headers").Call("get", "Content-Type"); ct.Type() == js.TypeString {
			ctype = ct.String()
		}
		resp.Call("arrayBuffer").Call("then", onBuf, onErr)
		return nil
	})
	js.Global().Call("fetch", url, map[string]any{"referrerPolicy": "no-referrer"}).Call("then", onResp, onErr)
	var r result
	select {
	case r = <-done:
	case <-ctx.Done():
		r = result{err: ctx.Err()}
	}
	// the callbacks stay until the fetch settles, even when the context
	// ended first: a released Func called by JavaScript would crash
	go func() {
		if ctx.Err() != nil {
			<-done
		}
		onResp.Release()
		onBuf.Release()
		onErr.Release()
	}()
	return r.b, r.ctype, r.err
}
