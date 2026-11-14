// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"encoding/json"
	"errors"
	"sync"
	"syscall/js"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/videoproto"
	"github.com/go-slides/go-slides.github.io/playground/videoproto/remote"
	"github.com/go-slides/slides/video"
)

// The page's messages about video, as strings like {"ready":true}: a render
// worker asks for the video worker, and says it stopped answering.
const (
	msgNeedVideo  = `{"needVideo":true}`
	msgVideoStuck = `{"videoStuck":true}`
)

const (
	// videoLoadWait is how long a render waits for the page to load the
	// video worker and link it (a download and a compilation).
	videoLoadWait = 90 * time.Second
	// videoCallWait is how long one request waits for the video worker: a
	// poster is bounded to 600 samples, so this is a hung worker.
	videoCallWait = 25 * time.Second
)

// videoLink is a render worker's line to the video worker: a MessagePort the
// page gave it, and the requests waiting for answers. It is the
// remote.Transport of the video provider this module registers in place of
// go-avkit (it is built with slides_novideo).
type videoLink struct {
	mu      sync.Mutex
	port    js.Value
	asked   bool
	linked  chan error
	pending map[int]chan videoReply
	onMsg   js.Func
	prov    *remote.Provider
}

type videoReply struct {
	res videoproto.Response
	img []byte
}

var theVideo = newVideoLink()

func newVideoLink() *videoLink {
	l := &videoLink{pending: map[int]chan videoReply{}, linked: make(chan error, 1)}
	l.prov = remote.New(l)
	l.onMsg = js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			l.reply(args[0].Get("data"))
		}
		return nil
	})
	return l
}

// register makes the video worker the provider of the decks rendered here.
func (l *videoLink) register() { video.Register(l.prov) }

// setPort takes the page's answer to the request for the video worker: its
// port, or why there is none.
func (l *videoLink) setPort(port js.Value, why string) {
	l.mu.Lock()
	if port.Truthy() {
		l.port = port
		port.Call("addEventListener", "message", l.onMsg)
		port.Call("start")
		l.prov.Reset()
	}
	l.mu.Unlock()
	var err error
	if !port.Truthy() {
		err = errors.New(why)
	}
	select {
	case l.linked <- err:
	default:
	}
}

func (l *videoLink) reply(msg js.Value) {
	if msg.Type() != js.TypeObject || msg.Get("json").Type() != js.TypeString {
		return
	}
	var res videoproto.Response
	if json.Unmarshal([]byte(msg.Get("json").String()), &res) != nil {
		return
	}
	var img []byte
	if d := msg.Get("data"); d.Truthy() {
		img = make([]byte, d.Get("byteLength").Int())
		js.CopyBytesToGo(img, js.Global().Get("Uint8Array").New(d))
	}
	l.mu.Lock()
	ch := l.pending[res.ID]
	delete(l.pending, res.ID)
	l.mu.Unlock()
	if ch != nil {
		ch <- videoReply{res, img}
	}
}

// ensure has the page load the video worker and link it, the first time.
func (l *videoLink) ensure() error {
	l.mu.Lock()
	if l.port.Truthy() {
		l.mu.Unlock()
		return nil
	}
	ask := !l.asked
	l.asked = true
	l.mu.Unlock()
	if ask {
		select { // an answer left from before is not this one's
		case <-l.linked:
		default:
		}
		js.Global().Call("postMessage", msgNeedVideo)
	}
	select {
	case err := <-l.linked:
		if err != nil {
			l.mu.Lock()
			l.asked = false
			l.mu.Unlock()
			return err
		}
		// a second waiter finds the port set
		select {
		case l.linked <- nil:
		default:
		}
		return nil
	case <-time.After(videoLoadWait):
		l.mu.Lock()
		l.asked = false
		l.mu.Unlock()
		return errors.New("the video reader did not load")
	}
}

// Call implements remote.Transport.
func (l *videoLink) Call(req videoproto.Request, payload []byte) (videoproto.Response, []byte, error) {
	if err := l.ensure(); err != nil {
		return videoproto.Response{}, nil, err
	}
	b, _ := json.Marshal(req)
	msg := map[string]any{"json": string(b)}
	var transfer []any
	if len(payload) > 0 {
		buf := js.Global().Get("Uint8Array").New(len(payload))
		js.CopyBytesToJS(buf, payload)
		msg["data"] = buf.Get("buffer")
		transfer = []any{buf.Get("buffer")}
	}
	ch := make(chan videoReply, 1)
	l.mu.Lock()
	l.pending[req.ID] = ch
	port := l.port
	l.mu.Unlock()
	if transfer != nil {
		port.Call("postMessage", msg, transfer)
	} else {
		port.Call("postMessage", msg)
	}
	select {
	case r := <-ch:
		return r.res, r.img, nil
	case <-time.After(videoCallWait):
		l.mu.Lock()
		delete(l.pending, req.ID)
		l.port, l.asked = js.Value{}, false
		l.mu.Unlock()
		// the page stops the video worker, and the next request starts another
		js.Global().Call("postMessage", msgVideoStuck)
		return videoproto.Response{}, nil, errors.New("it did not answer in " + videoCallWait.String())
	}
}
