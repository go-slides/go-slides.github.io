// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"errors"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/go-slides/slides/emoji"
	"github.com/go-slides/slides/mathfont"

	playground "github.com/go-slides/go-slides.github.io/playground"
	"github.com/go-slides/go-slides.github.io/playground/emojipack"
	"github.com/go-slides/go-slides.github.io/playground/exportdoc"
	"github.com/go-slides/go-slides.github.io/playground/lazyasset"
)

// This module is built with the tags slides_nomathfont and slides_noemoji: it
// links no font for formulas (651 KB gzipped less) and no emoji graphics (2.06
// MB less); the emoji are one pack, fetched the same way (setupEmoji). The render worker gets the font as a file next
// to itself (fonts/stix-math.otf), the first time a render has a formula, and
// not before: a deck without math never fetches it.
//
// Rendering never waits. Until the file is here a formula is its TeX source
// (span.math-error); when it arrives the worker registers the provider again
// and tells the page (msgAssetReady), which renders the deck once more
// (ViewModel.Rerender). A failed download leaves the source, tells nobody, and
// is tried again by the next render that has a formula.
//
// lazyFile is the whole mechanism, so that the emoji font (go-slides v0.23.0,
// the same contract) is one more call of it: fetch once, register, render once.

const (
	// msgAssetReady is what a worker posts when a lazily fetched file has
	// arrived and its provider is registered; the name follows.
	msgAssetReady = `{"assetReady":`

	// mathFontFile is the font's place beside worker.js (cmd/build writes it,
	// with its licence: MathFontFile).
	mathFontFile = "fonts/stix-math.otf"

	// emojiPackFile is the Twemoji graphics, one file (cmd/build: EmojiPackFile).
	emojiPackFile = "emoji/twemoji.pack"

	// exportFontWait is how long an export's render waits for the font: its
	// whole render has docTimeout, and the page has no render to repeat.
	exportFontWait = 8 * time.Second
)

// theMath is the worker's math font; nil in the page, which typesets nothing.
var theMath *lazyasset.Asset

// lazyFile returns the Asset for file (a path beside this worker's script, with
// the deploy's ?v= as the other files have it, so the browser's cache and the
// offline cache key on it) and, when its bytes arrive, calls register with it
// (the provider takes them up: Register again) and tells the page, which
// renders again. name is what the page is told. When the download fails,
// failed is called (and the page is told nothing): the host drops the slides
// the render cache kept, so that the next render asks for the file again.
func lazyFile(name, file string, register func(*lazyasset.Asset), failed func(error)) *lazyasset.Asset {
	var a *lazyasset.Asset
	url := file + js.Global().Get("location").Get("search").String()
	a = lazyasset.New(url, jsFetch, func([]byte) {
		register(a)
		js.Global().Call("postMessage", msgAssetReady+`"`+name+`"}`)
	}).OnFailure(failed)
	return a
}

// theEmoji is the worker's emoji pack; nil in the page.
var theEmoji *lazyasset.Asset

// setupEmoji registers the worker's emoji provider (go-slides v0.23.0, built
// with slides_noemoji): the Twemoji graphics are one file, fetched the first time
// a render has an emoji. Until it is here an emoji is the text written. The very
// helper the math font uses: fetch once, register again, the page renders once.
func setupEmoji() {
	var prov *emojipack.Provider
	theEmoji = lazyFile("emoji", emojiPackFile, func(*lazyasset.Asset) {
		emoji.Register(prov) // again: drops what was cached from before
		playground.ForgetRenderedSlides()
	}, func(error) {
		playground.ForgetRenderedSlides()
	})
	prov = emojipack.NewProvider(theEmoji.Bytes)
	emoji.Register(prov)
}

// setupMathFont registers the worker's math provider. Nothing is fetched yet.
func setupMathFont() {
	theMath = lazyFile("mathfont", mathFontFile, func(a *lazyasset.Asset) {
		mathfont.Register(a)
		// the render cache holds the slides as the last render wrote them, the
		// formulas as source: the next render must write them again
		playground.ForgetRenderedSlides()
	}, func(error) {
		// a slide the render cache holds is not rendered again, so would not
		// ask for the font again: the next render must
		playground.ForgetRenderedSlides()
	})
	mathfont.Register(theMath)
}

// jsFetch is lazyasset.Fetch over the worker's fetch: the file's bytes, or the
// error (HTTP status included), never a block.
func jsFetch(url string, done func([]byte, error)) {
	var onResp, onBuf, onErr js.Func
	finish := func(b []byte, err error) {
		onResp.Release()
		onBuf.Release()
		onErr.Release()
		done(b, err)
	}
	onErr = js.FuncOf(func(_ js.Value, args []js.Value) any {
		msg := "the request failed"
		if len(args) > 0 && args[0].Truthy() {
			msg = args[0].Call("toString").String()
		}
		finish(nil, errors.New(msg))
		return nil
	})
	onBuf = js.FuncOf(func(_ js.Value, args []js.Value) any {
		u8 := js.Global().Get("Uint8Array").New(args[0])
		b := make([]byte, u8.Get("length").Int())
		js.CopyBytesToGo(b, u8)
		finish(b, nil)
		return nil
	})
	onResp = js.FuncOf(func(_ js.Value, args []js.Value) any {
		r := args[0]
		if !r.Get("ok").Bool() {
			finish(nil, errors.New("HTTP "+strconv.Itoa(r.Get("status").Int())))
			return nil
		}
		r.Call("arrayBuffer").Call("then", onBuf, onErr)
		return nil
	})
	js.Global().Call("fetch", url).Call("then", onResp, onErr)
}

// hasMath reports whether a project may have a formula: a dollar sign in the
// deck or in a file beside it. A false positive costs one font download.
func hasMath(p exportdoc.Project) bool {
	if strings.Contains(p.Source, "$") {
		return true
	}
	for _, f := range p.Files {
		if strings.Contains(f, "$") {
			return true
		}
	}
	return false
}

// hasEmojiText reports whether the deck or a file beside it may draw an emoji.
func hasEmojiText(p exportdoc.Project) bool {
	if emojipack.HasEmoji(p.Source) {
		return true
	}
	for _, f := range p.Files {
		if emojipack.HasEmoji(f) {
			return true
		}
	}
	return false
}

// waitForAssets is for an export, which has no page to render again: a deck
// with formulas waits (up to exportFontWait) for the math font, one with emoji
// for the emoji pack, and is refused when one cannot be had, rather than written
// with its formulas or emoji as source.
func waitForAssets(p exportdoc.Project) error {
	if theMath != nil && hasMath(p) {
		if _, err := theMath.Wait(exportFontWait); err != nil {
			return errors.New("the math font could not be loaded for the export: " + err.Error())
		}
	}
	if theEmoji != nil && hasEmojiText(p) {
		if _, err := theEmoji.Wait(exportFontWait); err != nil {
			return errors.New("the emoji graphics could not be loaded for the export: " + err.Error())
		}
	}
	return nil
}
