// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

// Package lazyasset is a file the render worker fetches the first time a deck
// needs it, and not before: the STIX math font (go-slides' mathfont.Provider),
// and the emoji font that go-slides v0.23.0 will ask for the same way.
//
// Rendering is synchronous and never waits, so an Asset answers at once: its
// bytes when they are here, an error when they are not, and the first such
// answer starts the download. When the bytes arrive the Asset calls its ready
// function, once per successful fetch, and the host registers the provider
// again and renders the deck again (the page's view-model: Rerender). A fetch
// that fails leaves nothing behind but its error: the next ask starts another,
// so a failed download is retried by the next render and never by a timer, and
// nothing renders again because of it.
//
// The package has no JavaScript in it: the fetch is a function the host gives
// (cmd/playground-wasm gives the browser's), which is what makes the contract
// testable here.
package lazyasset

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// ErrLoading is what Bytes answers while a download is under way.
var ErrLoading = errors.New("lazyasset: still loading")

// A Fetch downloads url and calls done once with the body or the error, from
// any goroutine (before it returns, too). It must not block.
//
// max is the most it may read: a body longer than max is a failure, and the
// fetch stops reading at max (the answer is not buffered whole first).
type Fetch func(url string, max int, done func(body []byte, err error))

// DefaultMax bounds a fetch whose Asset was given no expected size.
const DefaultMax = 16 << 20

// An Asset is one file fetched on first use. It is safe for concurrent use.
type Asset struct {
	url    string
	fetch  Fetch
	ready  func([]byte)
	failed func(error)
	size   int      // the expected size, 0 if not given
	sum    [32]byte // the expected SHA-256, when size is given

	mu      sync.Mutex
	data    []byte
	cur     chan struct{} // closed when the fetch under way is over, in either way
	lastErr error
	fetches int
}

// New returns an Asset for url. ready, if not nil, is called with the bytes
// when they have arrived (from the goroutine that completed the fetch).
func New(url string, fetch Fetch, ready func([]byte)) *Asset {
	return &Asset{url: url, fetch: fetch, ready: ready}
}

// Bytes returns the file if it is here. If it is not, Bytes starts the fetch
// (one at a time: asking again while it is under way starts nothing) and
// returns an error that wraps ErrLoading; after a failed fetch the error says
// so, and a new fetch is started.
func (a *Asset) Bytes() ([]byte, error) {
	a.mu.Lock()
	if a.data != nil {
		b := a.data
		a.mu.Unlock()
		return b, nil
	}
	if a.cur != nil {
		a.mu.Unlock()
		return nil, fmt.Errorf("%w: %s", ErrLoading, a.url)
	}
	prev := a.lastErr
	a.lastErr = nil
	done := make(chan struct{})
	a.cur = done
	a.fetches++
	a.mu.Unlock()

	max := DefaultMax
	if a.size > 0 {
		max = a.size
	}
	a.fetch(a.url, max, func(body []byte, err error) { a.finish(done, body, err) })

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.data != nil { // the fetch answered before it returned
		return a.data, nil
	}
	if prev != nil {
		return nil, fmt.Errorf("%w: %s (the last attempt failed: %v)", ErrLoading, a.url, prev)
	}
	return nil, fmt.Errorf("%w: %s", ErrLoading, a.url)
}

func (a *Asset) finish(done chan struct{}, body []byte, err error) {
	a.mu.Lock()
	if a.cur != done { // a second answer to the same fetch: ignored
		a.mu.Unlock()
		return
	}
	a.cur = nil
	if err == nil && len(body) == 0 {
		err = errors.New("the file is empty")
	}
	// What a provider is given was built with the page and is checked here: the
	// worker may sit outside the offline cache's scope (a language page), whose
	// hash check would otherwise be the only one. A mismatch is a failed fetch:
	// the bytes are dropped, nothing renders, and the next ask fetches again.
	if err == nil && a.size > 0 {
		if len(body) != a.size {
			err = fmt.Errorf("the file is %d bytes, expected %d", len(body), a.size)
		} else if sha256.Sum256(body) != a.sum {
			err = errors.New("the file's SHA-256 is not the one built with this page")
		}
	}
	if err != nil {
		a.lastErr = fmt.Errorf("fetching %s: %w", a.url, err)
	} else {
		a.data = body
	}
	a.mu.Unlock()
	// ready first: a Wait that returns finds the provider registered again
	if err == nil && a.ready != nil {
		a.ready(body)
	}
	if err != nil && a.failed != nil {
		a.failed(err)
	}
	close(done)
}

// Expect sets the file's size and SHA-256 (64 hex digits) as the build wrote
// them: the fetch reads no more than size bytes, and a body that is not exactly
// that file is a failed fetch. A bad hash panics: it is a constant of the build.
// Call it before the first Bytes.
func (a *Asset) Expect(size int, sha256Hex string) *Asset {
	b, err := hex.DecodeString(sha256Hex)
	if err != nil || len(b) != 32 || size <= 0 {
		panic("lazyasset: Expect(" + strconv.Itoa(size) + ", " + sha256Hex + ")")
	}
	a.size = size
	copy(a.sum[:], b)
	return a
}

// OnFailure sets what is called, with the error, when a fetch has failed: the
// host registers its provider again, so that go-slides writes again the slides
// that fell back, the next render asks again (and so fetches again). It must not render: a failure is not an event
// that renders. Call it before the first Bytes.
func (a *Asset) OnFailure(f func(error)) *Asset {
	a.failed = f
	return a
}

// Wait is Bytes for a caller that may block (the export's render, which has no
// page to render again): it waits up to d for the fetch to end, and returns the
// bytes, or the error of the fetch or of the wait. Never call it from the
// goroutine that must run the fetch's callbacks.
func (a *Asset) Wait(d time.Duration) ([]byte, error) {
	b, err := a.Bytes()
	if err == nil {
		return b, nil
	}
	a.mu.Lock()
	done := a.cur
	a.mu.Unlock()
	if done != nil {
		select {
		case <-done:
		case <-time.After(d):
			return nil, fmt.Errorf("%w: %s (no answer in %v)", ErrLoading, a.url, d)
		}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.data != nil {
		return a.data, nil
	}
	if a.lastErr != nil {
		return nil, a.lastErr
	}
	return nil, err
}

// Font makes an Asset a go-slides mathfont.Provider: Font() ([]byte, error).
func (a *Asset) Font() ([]byte, error) { return a.Bytes() }

// Fetches is how many downloads were started, for tests and the debug page.
func (a *Asset) Fetches() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.fetches
}

// Loaded reports whether the bytes are here.
func (a *Asset) Loaded() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.data != nil
}
