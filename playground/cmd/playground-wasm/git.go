// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/go-browserhttp/browsergit/forge"
	"github.com/go-browserhttp/browsergit/remote"
	"github.com/go-browserhttp/browsergit/secret"
)

// The GitHub pane's side of the page (playground/gitvm.go). It runs here, in
// the page's instance of the module, and only here: the render worker is
// another instance, built from the same module but started by worker() and
// never by page(), so no host — and so no token — exists in it.

// githubAPI is the one origin the token is sent to. The page's Content
// Security Policy (layouts/_default/playground.html) allows no other, so a
// script injected into the page could not send it anywhere else either. There
// is deliberately no proxy here: GitHub's REST API answers CORS itself.
const githubAPI = "https://api.github.com"

// tokenKey is where "Remember on this device" keeps the token: localStorage,
// which the sandboxed preview frames (an opaque origin) and the render worker
// (no localStorage in a worker) cannot read. Nothing is written there unless
// the person ticks the box.
const tokenKey = "goSlidesGitHubToken"

// GitRemote implements playground.GitHost.
func (host) GitRemote(tok secret.Token) (remote.Remote, error) {
	// The default transport in the browser is the Fetch API, called directly
	// (redirect "error", credentials "omit"): no net/http in the module.
	return forge.New(forge.Config{
		Kind:         forge.GitHub,
		API:          githubAPI,
		Token:        tok,
		BranchPrefix: "slides/",
	})
}

// GitRun implements playground.GitHost: op blocks on the network, so it runs
// in a goroutine (the event loop keeps turning), and done comes back on a turn
// of its own.
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

func localStorage() (js.Value, bool) {
	var ls js.Value
	if guard(func() { ls = js.Global().Get("localStorage") }) != nil || ls.IsUndefined() || ls.IsNull() {
		return js.Value{}, false
	}
	return ls, true
}

// KeepToken implements playground.TokenKeeper (only on the person's choice).
func (host) KeepToken(tok secret.Token) {
	if ls, ok := localStorage(); ok {
		_ = guard(func() { ls.Call("setItem", tokenKey, tok.Reveal()) })
	}
}

// KeptToken implements playground.TokenKeeper.
func (host) KeptToken() secret.Token {
	ls, ok := localStorage()
	if !ok {
		return secret.Token{}
	}
	var v js.Value
	if guard(func() { v = ls.Call("getItem", tokenKey) }) != nil || v.Type() != js.TypeString {
		return secret.Token{}
	}
	return secret.New(v.String())
}

// ForgetToken implements playground.TokenKeeper.
func (host) ForgetToken() {
	if ls, ok := localStorage(); ok {
		_ = guard(func() { ls.Call("removeItem", tokenKey) })
	}
}
