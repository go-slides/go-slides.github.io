// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"syscall/js"
	"time"

	"github.com/go-browserhttp/browsergit/remote"
	playground "github.com/go-slides/go-slides.github.io/playground"
	"github.com/go-widgets/toolkit"
)

// The page's side of the git frame (playground/gitframe.go): an iframe of the
// second origin go-slides-usercontent.github.io laid over the Git pane, and the
// postMessage to and from it. The page names that origin once (goSlidesGitFrame
// Origin, from the site's configuration) and uses it three ways: as the iframe's
// address, as the targetOrigin of every message it sends (never "*"), and as the
// only origin it believes, together with the iframe's own window as the source.
// The frame is not sandboxed: it needs the storage of its own origin, and what
// keeps it from this page is that it IS another origin.

// frameTimeout bounds one command: a frame that never answers is an error, not
// a pane that is busy for ever.
const frameTimeout = 5 * time.Minute

type frameReply struct {
	data  json.RawMessage
	files remote.Files
	err   error
}

type gitFrame struct {
	origin, nonce string
	el            js.Value
	mu            sync.Mutex
	seq           int
	pending       map[string]chan frameReply
	watch         func(playground.FrameState)
	last          *playground.FrameState
	// placement, so that only a change is written
	placed toolkit.Rect
	clip   string
}

// theGitFrame is nil on a page with no frame origin (then client mode has no
// forge at all: there is no way to a token on this page).
var theGitFrame *gitFrame

// newGitFrame reads the origin the page was built with. The element is created
// when the pane first asks for it, so a visit that never opens the Git pane
// downloads nothing of the frame.
func newGitFrame() *gitFrame {
	o := js.Global().Get("goSlidesGitFrameOrigin")
	if o.Type() != js.TypeString || o.String() == "" {
		return nil
	}
	buf := js.Global().Get("Uint8Array").New(16)
	js.Global().Get("crypto").Call("getRandomValues", buf)
	b := make([]byte, 16)
	js.CopyBytesToGo(b, buf)
	const hex = "0123456789abcdef"
	n := make([]byte, 32)
	for i, c := range b {
		n[2*i], n[2*i+1] = hex[c>>4], hex[c&15]
	}
	f := &gitFrame{origin: o.String(), nonce: string(n), pending: map[string]chan frameReply{}}
	js.Global().Call("addEventListener", "message", js.FuncOf(func(_ js.Value, a []js.Value) any {
		f.onMessage(a[0])
		return nil
	}))
	return f
}

// ensure creates the iframe on first use.
func (f *gitFrame) ensure() {
	if f.el.Truthy() {
		return
	}
	lang := "en"
	if l := js.Global().Get("goSlidesLang"); l.Type() == js.TypeString {
		lang = l.String()
	}
	doc := js.Global().Get("document")
	el := doc.Call("createElement", "iframe")
	el.Set("id", "slides-gitframe")
	el.Call("setAttribute", "referrerpolicy", "no-referrer")
	el.Call("setAttribute", "title", playground.T(playground.KFrameTitleGit))
	el.Get("style").Set("cssText", "position:absolute;display:none;border:1px solid var(--line);border-radius:8px;background:#fff")
	// The frame does not ask the browser who framed it (location.ancestorOrigins
	// is missing in Firefox, and no referrer is sent): this page SAYS HELLO, with
	// the nonce, to the frame's origin once the frame's document has loaded, and
	// the frame believes it by the message's origin, its source and the nonce.
	// It says nothing to anybody before that.
	el.Call("addEventListener", "load", js.FuncOf(func(js.Value, []js.Value) any {
		hello := js.Global().Get("Object").New()
		hello.Set("t", "hello")
		hello.Set("nonce", f.nonce)
		el.Get("contentWindow").Call("postMessage", hello, f.origin)
		return nil
	}))
	el.Set("src", f.origin+"/shell.html?lang="+lang+"#"+f.nonce)
	doc.Call("querySelector", ".stage").Call("appendChild", el)
	f.el = el
}

// place puts the iframe over the pane's slot, or hides it.
func (f *gitFrame) place(a *app) {
	r, ok := a.GitFrameRect()
	if !ok {
		if f.el.Truthy() && f.placed != (toolkit.Rect{}) {
			f.placed = toolkit.Rect{}
			f.el.Get("style").Set("display", "none")
		}
		return
	}
	f.ensure()
	px := func(v int) string { return strconv.FormatFloat(float64(v)/a.dpr, 'f', 2, 64) + "px" }
	clip := "none"
	if pop, open := a.OpenPopover(); open {
		clip = playground.PreviewClip(r, pop, a.dpr)
	}
	st := f.el.Get("style")
	if clip != f.clip {
		f.clip = clip
		st.Set("clipPath", clip)
	}
	if r == f.placed {
		return
	}
	f.placed = r
	st.Set("display", "block")
	st.Set("left", px(r.X))
	st.Set("top", px(r.Y))
	st.Set("width", px(r.W))
	st.Set("height", px(r.H))
}

// onMessage believes a message only from the frame's own window, from the
// frame's origin.
func (f *gitFrame) onMessage(e js.Value) {
	if !f.el.Truthy() || !e.Get("source").Equal(f.el.Get("contentWindow")) || e.Get("origin").String() != f.origin {
		return
	}
	d := e.Get("data")
	if d.Type() != js.TypeObject || d.Get("t").Type() != js.TypeString {
		return
	}
	switch d.Get("t").String() {
	case "result":
		f.result(d)
	case "ready", "state":
		s := playground.FrameState{
			Connected:  d.Get("connected").Truthy(),
			Remembered: d.Get("remembered").Truthy(),
		}
		if fg := d.Get("forge"); fg.Type() == js.TypeString && (fg.String() == "github" || fg.String() == "gitlab") {
			s.Forge = fg.String()
		}
		if s.Connected && s.Forge == "" {
			s.Connected = false // a connection to no forge we know is none
		}
		f.last = &s
		if f.watch != nil {
			f.watch(s)
		}
	}
}

func (f *gitFrame) result(d js.Value) {
	id := d.Get("id")
	if id.Type() != js.TypeString {
		return
	}
	f.mu.Lock()
	ch, ok := f.pending[id.String()]
	delete(f.pending, id.String())
	f.mu.Unlock()
	if !ok {
		return
	}
	var r frameReply
	if !d.Get("ok").Truthy() {
		msg := "the frame refused the command"
		if e := d.Get("error"); e.Type() == js.TypeString {
			msg = e.String()
		}
		r.err = errors.New(msg)
		ch <- r
		return
	}
	r.data = json.RawMessage(js.Global().Get("JSON").Call("stringify", d.Get("data")).String())
	if fs := d.Get("files"); fs.Type() == js.TypeObject {
		r.files = remote.Files{}
		keys := js.Global().Get("Object").Call("keys", fs)
		for i := 0; i < keys.Length(); i++ {
			k := keys.Index(i).String()
			u := fs.Get(k)
			b := make([]byte, u.Get("length").Int())
			js.CopyBytesToGo(b, u)
			r.files[k] = b
		}
	}
	ch <- r
}

// FrameMode implements playground.FrameHost.
func (host) FrameMode() bool { return theGitFrame != nil }

// FrameWatch implements playground.FrameHost.
func (host) FrameWatch(w func(playground.FrameState)) {
	f := theGitFrame
	f.watch = w
	if f.last != nil {
		w(*f.last)
	}
}

// FrameCall implements playground.FrameHost. It runs off the UI's turn (GitRun),
// so it may wait for the answer.
func (host) FrameCall(cmd string, fields map[string]any, files remote.Files) (json.RawMessage, remote.Files, error) {
	f := theGitFrame
	if f == nil || !f.el.Truthy() {
		return nil, nil, errors.New("the git frame is not loaded")
	}
	f.mu.Lock()
	f.seq++
	id := "c" + strconv.Itoa(f.seq)
	ch := make(chan frameReply, 1)
	f.pending[id] = ch
	f.mu.Unlock()

	msg := js.Global().Get("Object").New()
	for k, v := range fields {
		msg.Set(k, js.ValueOf(v))
	}
	msg.Set("t", cmd)
	msg.Set("id", id)
	msg.Set("nonce", f.nonce)
	if files != nil || cmd == "propose" { // propose always carries the (possibly empty) files
		o := js.Global().Get("Object").New()
		for p, b := range files {
			u := js.Global().Get("Uint8Array").New(len(b))
			js.CopyBytesToJS(u, b)
			o.Set(p, u)
		}
		msg.Set("files", o)
	}
	// the frame's origin, never "*"
	f.el.Get("contentWindow").Call("postMessage", msg, f.origin)

	select {
	case r := <-ch:
		return r.data, r.files, r.err
	case <-time.After(frameTimeout):
		f.mu.Lock()
		delete(f.pending, id)
		f.mu.Unlock()
		return nil, nil, errors.New("the git frame did not answer")
	}
}
