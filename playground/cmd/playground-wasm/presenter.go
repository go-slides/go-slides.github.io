// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"syscall/js"

	playground "github.com/go-slides/go-slides.github.io/playground"
)

// presenterShell is the presenter window's own page: nothing but a sandboxed
// iframe for the presenter view, like the preview's. The window is opened by
// this page, so it is this page's (trusted) origin; the deck runs only in
// the iframe, with an opaque origin and its own CSP.
const presenterShell = `<!doctype html><html lang="en"><head><meta charset="utf-8">` +
	`<title>Presenter view — go-slides playground</title>` +
	`<style>html,body{margin:0;height:100%;background:#222}iframe{display:block;border:0;width:100%;height:100%}</style>` +
	`</head><body><iframe id="pv" sandbox="allow-scripts" referrerpolicy="no-referrer" title="Presenter view"></iframe>` +
	// The presenter view takes its messages from its parent, this window:
	// what the playground page (the opener) posts here is passed on from
	// here, and only a slidesSync.
	`<script>addEventListener("message",function(e){var f=document.getElementById("pv");` +
	`if(e.source===opener&&e.data&&e.data.slidesSync&&f.contentWindow)f.contentWindow.postMessage(e.data,"*")})</script>` +
	`</body></html>`

// presenter is the presenter window and the relay between it and the
// preview. A sandboxed page has no storage to share, so each presenting page
// posts {slidesSync: {index, fragmentIndex}} to its parent on every change
// (go-slides' viewer.SyncParent); this page passes it to the other one. It
// relays only that, between its own two iframes.
type presenter struct {
	vm       *playground.ViewModel
	audience js.Value // the preview iframe
	win      js.Value // the presenter window, or undefined
	frame    js.Value // its iframe
	last     js.Value // the last slidesSync message, to restore after a reload
	relayed  int      // messages relayed, for the test harness
}

func (p *presenter) isOpen() bool {
	return p.win.Truthy() && !p.win.Get("closed").Bool()
}

// open opens the presenter window, or brings it forward.
func (p *presenter) open() {
	if p.isOpen() {
		p.win.Call("focus")
		return
	}
	w := js.Global().Call("open", "", "go-slides-presenter", "width=1100,height=680")
	if !w.Truthy() {
		p.vm.Status.Set("⚠ The browser blocked the presenter window: allow pop-ups for this page, then Present again.")
		return
	}
	d := w.Get("document")
	d.Call("open")
	d.Call("write", presenterShell)
	d.Call("close")
	p.win = w
	p.frame = d.Call("getElementById", "pv")
	p.frame.Call("addEventListener", "load", js.FuncOf(func(js.Value, []js.Value) any {
		p.restore(p.frame)
		return nil
	}))
	w.Call("addEventListener", "message", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 || !args[0].Get("source").Equal(p.frame.Get("contentWindow")) {
			return nil
		}
		if s := args[0].Get("data"); s.Truthy() && s.Get("slidesSync").Truthy() {
			p.last = s
			p.relay(p.audience, s)
		}
		return nil
	}))
	w.Call("addEventListener", "pagehide", js.FuncOf(func(js.Value, []js.Value) any {
		p.vm.PresenterOpen.Set(false)
		return nil
	}))
	if doc := p.vm.PresenterDoc.Get(); doc != "" {
		p.frame.Set("srcdoc", doc)
	}
	p.vm.PresenterOpen.Set(true)
}

// fromAudience takes a message the preview posted.
func (p *presenter) fromAudience(data js.Value) {
	switch {
	case data.Get("slidesSync").Truthy():
		p.last = data
		if p.isOpen() {
			p.toPresenter(data)
		}
	case data.Get("slidesPresent").Truthy():
		p.vm.Present.Execute()
	}
}

func (p *presenter) relay(to, data js.Value) {
	if w := to.Get("contentWindow"); w.Truthy() {
		w.Call("postMessage", data, "*")
		p.relayed++
	}
}

// toPresenter posts to the presenter view from its parent, the window.
func (p *presenter) toPresenter(data js.Value) {
	// The window is this page's origin: post to it with that origin.
	p.win.Call("postMessage", data, js.Global().Get("location").Get("origin"))
	p.relayed++
}

// restore puts a reloaded page (a re-render) back on the current slide.
func (p *presenter) restore(frame js.Value) {
	if !p.last.Truthy() {
		return
	}
	if frame.Equal(p.frame) {
		p.toPresenter(p.last)
	} else {
		p.relay(frame, p.last)
	}
}

// setDoc shows a new presenter page, when the window is open.
func (p *presenter) setDoc(doc string) {
	if p.isOpen() && doc != "" {
		p.frame.Set("srcdoc", doc)
	}
}
