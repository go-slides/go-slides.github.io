// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/go-browserhttp/browsergit/forge"
	"github.com/go-browserhttp/browsergit/oauth"
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

// gitlabBase is GitLab.com, the one GitLab the public page talks to (its CSP
// allows https://gitlab.com and no other).
const gitlabBase = "https://gitlab.com"

// GitRemote implements playground.GitHost.
func (host) GitRemote(name string, tok secret.Token) (remote.Remote, error) {
	// The default transport in the browser is the Fetch API, called directly
	// (redirect "error", credentials "omit"): no net/http in the module.
	cfg := forge.Config{Kind: forge.GitHub, API: githubAPI, Token: tok, BranchPrefix: "slides/"}
	if name == "gitlab" {
		cfg.Kind, cfg.API = forge.GitLab, gitlabBase+"/api/v4"
	}
	return forge.New(cfg)
}

// attemptKey is where a sign-in's state and verifier wait for the way back:
// sessionStorage, this tab only, deleted as soon as it is read.
const attemptKey = "goSlidesOAuthAttempt"

// OAuthConfig implements playground.OAuthHost. The page sets
// goSlidesGitLabClientID from the site's configuration (empty: no application
// registered, sign-in off). The redirect URI is this very page, exactly: one
// per language, each registered with the GitLab application.
func (host) OAuthConfig(name string) (oauth.Config, bool) {
	if name != "gitlab" {
		return oauth.Config{}, false
	}
	id := js.Global().Get("goSlidesGitLabClientID")
	if id.Type() != js.TypeString || id.String() == "" {
		return oauth.Config{}, false
	}
	loc := js.Global().Get("location")
	redirect := loc.Get("origin").String() + loc.Get("pathname").String()
	return oauth.GitLab(gitlabBase, id.String(), redirect), true
}

// PageURL implements playground.OAuthHost.
func (host) PageURL() string { return js.Global().Get("location").Get("href").String() }

// CleanURL implements playground.OAuthHost: the query (code, state) goes, the
// fragment (a shared deck) stays, and no history entry is made.
func (host) CleanURL() {
	loc := js.Global().Get("location")
	_ = guard(func() {
		js.Global().Get("history").Call("replaceState", js.Null(), "", loc.Get("pathname").String()+loc.Get("hash").String())
	})
}

func sessionStorage() (js.Value, bool) {
	var ss js.Value
	if guard(func() { ss = js.Global().Get("sessionStorage") }) != nil || ss.IsUndefined() || ss.IsNull() {
		return js.Value{}, false
	}
	return ss, true
}

// KeepAttempt implements playground.OAuthHost.
func (host) KeepAttempt(a oauth.Attempt) {
	b, err := json.Marshal(a)
	if err != nil {
		return
	}
	if ss, ok := sessionStorage(); ok {
		_ = guard(func() { ss.Call("setItem", attemptKey, string(b)) })
	}
}

// TakeAttempt implements playground.OAuthHost: read once, deleted at once.
func (host) TakeAttempt() (oauth.Attempt, bool) {
	ss, ok := sessionStorage()
	if !ok {
		return oauth.Attempt{}, false
	}
	var v js.Value
	_ = guard(func() {
		v = ss.Call("getItem", attemptKey)
		ss.Call("removeItem", attemptKey)
	})
	if v.Type() != js.TypeString {
		return oauth.Attempt{}, false
	}
	var a oauth.Attempt
	if json.Unmarshal([]byte(v.String()), &a) != nil {
		return oauth.Attempt{}, false
	}
	return a, true
}

// Navigate implements playground.OAuthHost.
func (host) Navigate(url string) { js.Global().Get("location").Call("assign", url) }

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
