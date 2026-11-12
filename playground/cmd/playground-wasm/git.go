// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"errors"
	"syscall/js"
	"time"

	"github.com/go-browserhttp/browsergit/forge"
	"github.com/go-browserhttp/browsergit/gitd"
	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-browserhttp/browsergit/secret"
	playground "github.com/go-slides/go-slides.github.io/playground"
)

// The GitHub pane's side of the page (playground/gitvm.go).
//
// ⛔ THIS PAGE HOLDS NO FORGE TOKEN, and makes no call to a forge. In client mode
// every forge call goes through the git frame (gitframe.go here, and
// playground/gitframe.go): a document on go-slides-usercontent.github.io, a
// second origin, holds the token, does the GitLab sign-in, remembers the token
// on the person's say-so in its own localStorage, and runs the calls. The page's
// CSP no longer names api.github.com or gitlab.com in connect-src, so even this
// module could not reach them (go-slides#90).
//
// What stays here is the server mode's client (gitd, same origin only), and the
// removal of a token an earlier version kept in this origin's localStorage.

// githubAPI and gitlabBase are the API origins a kept token was recorded for;
// nothing here sends to them.
const (
	githubAPI  = "https://api.github.com"
	gitlabBase = "https://gitlab.com"
)

// GitOrigin implements playground.GitHost.
func (host) GitOrigin(name string) string {
	if name == "gitlab" {
		return gitlabBase
	}
	return githubAPI
}

// GitRemote implements playground.GitHost. It is never the way to a forge on
// this page: client mode runs through the frame (FrameHost, gitframe.go), and a
// pane without one has no forge at all.
func (host) GitRemote(string, secret.Token) (remote.Remote, error) {
	return nil, errors.New("this page makes no forge call: the git frame does")
}

// GitRun implements playground.GitHost: op blocks (on the frame's answer, or on
// the network for the server mode), so it runs in a goroutine (the event loop
// keeps turning), and done comes back on a turn of its own.
func (host) GitRun(op func(), done func()) {
	go func() {
		op()
		var f js.Func
		f = js.FuncOf(func(js.Value, []js.Value) any {
			f.Release()
			done()
			return nil
		})
		js.Global().Call("setTimeout", f, 0)
	}()
}

// PostUI implements playground.UIPoster: f runs on a turn of the event loop, never
// on the goroutine that asks.
func (host) PostUI(f func()) {
	var fn js.Func
	fn = js.FuncOf(func(js.Value, []js.Value) any {
		fn.Release()
		f()
		return nil
	})
	js.Global().Call("setTimeout", fn, 0)
}

func localStorage() (js.Value, bool) {
	var ls js.Value
	if guard(func() { ls = js.Global().Get("localStorage") }) != nil || ls.IsUndefined() || ls.IsNull() {
		return js.Value{}, false
	}
	return ls, true
}

// KeepToken implements playground.TokenKeeper. It keeps nothing: the frame
// keeps a remembered token, on its own origin. (The interface stays for the
// removal below.)
func (host) KeepToken(string, string, secret.Token) {}

// KeptToken implements playground.TokenKeeper: what an EARLIER version of this
// page kept here, so that the pane can drop it and say so.
func (host) KeptToken() (forge, origin string, tok secret.Token) {
	ls, ok := localStorage()
	if !ok {
		return "", "", secret.Token{}
	}
	var v js.Value
	_ = guard(func() { v = ls.Call("getItem", playground.KeptTokenKey) })
	if v.Type() != js.TypeString {
		return "", "", secret.Token{}
	}
	forge, origin, t, ok := playground.DecodeKept(v.String())
	if !ok {
		return "", "", secret.Token{}
	}
	return forge, origin, secret.New(t)
}

// ForgetToken implements playground.TokenKeeper: both keys, the record and the
// bare token the first version wrote.
func (host) ForgetToken() {
	if ls, ok := localStorage(); ok {
		_ = guard(func() {
			ls.Call("removeItem", playground.KeptTokenKey)
			ls.Call("removeItem", playground.LegacyTokenKey)
		})
	}
}

// Gitd implements playground.GitdHost: the client for slides-gitd on the page's
// own origin (/api/git). Whether there is one is found out by asking it
// (playground/gitdvm.go). Its calls use the Fetch API in mode "same-origin": the
// browser refuses any other origin, so no forge host is reached, and the CSRF
// token it keeps lives in this module's memory only.
func (host) Gitd() playground.GitdClient {
	c, err := gitd.New(gitd.Config{})
	if err != nil {
		return nil
	}
	return c
}

// GitdSignIn implements playground.GitdHost: oauth2-proxy's sign-in, which sends
// the browser back to this page (a path on this origin, as it requires).
func (host) GitdSignIn() {
	loc := js.Global().Get("location")
	back := js.Global().Call("encodeURIComponent", loc.Get("pathname").String()+loc.Get("search").String())
	loc.Call("assign", "/oauth2/sign_in?rd="+back.String())
}

// GitdFront implements playground.GitdHost: whether this page is served by a
// slides-gitd front. The front serves a static marker, /.well-known/slides-gitd,
// on the page's own origin (selfhost/Caddyfile.gitd); it is asked for separately
// from the API, so a server that is down is still known to be the server and the
// pane does not fall back to a token. Anything but a 200 whose body names
// slides-gitd is "no".
func (host) GitdFront(ctx context.Context) bool {
	rsp, err := forge.DefaultTransport().RoundTrip(ctx, &forge.Request{
		Method: "GET", URL: "/.well-known/slides-gitd", SameOrigin: true, Limit: 4096,
		Header: map[string]string{"Accept": "application/json"},
	})
	if err != nil || rsp.Status != 200 {
		return false
	}
	var m map[string]any
	if json.Unmarshal(rsp.Body, &m) != nil {
		return false
	}
	_, ok := m["slides-gitd"]
	return ok
}

// HoldStorage implements playground.StorageHolder: the page may be served by a
// slides-gitd front, so nothing stored is shown until the server says who is
// signed in (playground/gitdvm.go).
func (h host) HoldStorage() bool { return h.hold }

// probeFront asks for the front's marker before the view-model starts. Only a
// definite 404 (an ordinary static host) lets the page start on its stored
// project at once; the marker, or no clear answer, holds it.
func probeFront() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rsp, err := forge.DefaultTransport().RoundTrip(ctx, &forge.Request{
		Method: "GET", URL: "/.well-known/slides-gitd", SameOrigin: true, Limit: 4096,
		Header: map[string]string{"Accept": "application/json"},
	})
	return err != nil || rsp.Status != 404
}
