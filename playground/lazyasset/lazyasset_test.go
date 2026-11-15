// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

package lazyasset

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-tex/math"

	"github.com/go-slides/slides"
	"github.com/go-slides/slides/mathfont"
	"github.com/go-slides/slides/mathfont/embedded"
)

// netStub is a fetch the test answers by hand.
type netStub struct {
	mu      sync.Mutex
	pending []func([]byte, error)
	urls    []string
}

func (n *netStub) fetch(url string, done func([]byte, error)) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.urls = append(n.urls, url)
	n.pending = append(n.pending, done)
}

func (n *netStub) answer(b []byte, err error) {
	n.mu.Lock()
	d := n.pending[0]
	n.pending = n.pending[1:]
	n.mu.Unlock()
	d(b, err)
}

// TestNothingIsFetchedUntilAsked: an Asset no render asks for downloads nothing
// (a deck without a formula never fetches the font).
func TestNothingIsFetchedUntilAsked(t *testing.T) {
	n := &netStub{}
	a := New("/playground/fonts/stix-math.otf", n.fetch, nil)
	if len(n.urls) != 0 || a.Fetches() != 0 {
		t.Fatalf("fetched before being asked: %v", n.urls)
	}
}

// TestOneFetchWhileUnderWay: many formulas in one render, one download, and
// the ready function once with the bytes.
func TestOneFetchWhileUnderWay(t *testing.T) {
	n := &netStub{}
	var got [][]byte
	a := New("/f", n.fetch, func(b []byte) { got = append(got, b) })
	for i := 0; i < 5; i++ {
		if _, err := a.Bytes(); !errors.Is(err, ErrLoading) {
			t.Fatalf("ask %d: %v", i, err)
		}
	}
	if a.Fetches() != 1 {
		t.Fatalf("%d fetches for five asks", a.Fetches())
	}
	n.answer([]byte("font"), nil)
	if len(got) != 1 || string(got[0]) != "font" {
		t.Fatalf("ready called %d times: %q", len(got), got)
	}
	for i := 0; i < 3; i++ {
		if b, err := a.Bytes(); err != nil || string(b) != "font" {
			t.Fatalf("after: %q, %v", b, err)
		}
	}
	if a.Fetches() != 1 || len(got) != 1 {
		t.Errorf("after the bytes: %d fetches, ready %d times", a.Fetches(), len(got))
	}
}

// TestAFailureRetriesOnTheNextAskAndNeverRendersAgain: a failed fetch leaves
// the error, not the ready call; the next ask fetches again, and only an ask
// does.
func TestAFailureRetriesOnTheNextAskAndNeverRendersAgain(t *testing.T) {
	n := &netStub{}
	ready := 0
	a := New("/f", n.fetch, func([]byte) { ready++ })
	a.Bytes()
	n.answer(nil, errors.New("HTTP 503"))
	if ready != 0 || a.Loaded() || len(n.pending) != 0 || a.Fetches() != 1 {
		t.Fatalf("a failure started something: ready %d, pending %d, fetches %d", ready, len(n.pending), a.Fetches())
	}
	_, err := a.Bytes()
	if !errors.Is(err, ErrLoading) || !strings.Contains(err.Error(), "HTTP 503") {
		t.Fatalf("the retry's answer does not say why the last one failed: %v", err)
	}
	if a.Fetches() != 2 {
		t.Fatalf("%d fetches: the next ask did not retry", a.Fetches())
	}
	n.answer([]byte("ok"), nil)
	if ready != 1 || !a.Loaded() {
		t.Errorf("after the retry: ready %d, loaded %v", ready, a.Loaded())
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.pending) != 0 {
		t.Error("a request is left under way")
	}
}

// TestOnFailure: the host hears of a failed download (to drop what it kept of
// the render that asked), once per failure, and of no success.
func TestOnFailure(t *testing.T) {
	n := &netStub{}
	var failed []error
	a := New("/f", n.fetch, func([]byte) { t.Error("ready on a failure") }).OnFailure(func(err error) { failed = append(failed, err) })
	a.Bytes()
	n.answer(nil, errors.New("HTTP 500"))
	a.Bytes()
	n.answer(nil, errors.New("HTTP 502"))
	if len(failed) != 2 || !strings.Contains(failed[1].Error(), "HTTP 502") {
		t.Fatalf("%v", failed)
	}
}

// TestAnEmptyBodyIsAFailure: a 200 with nothing in it is not the font.
func TestAnEmptyBodyIsAFailure(t *testing.T) {
	n := &netStub{}
	a := New("/f", n.fetch, func([]byte) { t.Error("ready on an empty body") })
	a.Bytes()
	n.answer(nil, nil)
	if a.Loaded() {
		t.Error("loaded an empty file")
	}
}

// TestSynchronousFetch: a fetch that answers before it returns (a cache hit)
// gives the bytes to the very ask that started it.
func TestSynchronousFetch(t *testing.T) {
	a := New("/f", func(_ string, done func([]byte, error)) { done([]byte("x"), nil) }, nil)
	if b, err := a.Bytes(); err != nil || string(b) != "x" {
		t.Fatalf("%q, %v", b, err)
	}
}

func TestWait(t *testing.T) {
	n := &netStub{}
	a := New("/f", n.fetch, nil)
	go func() {
		for {
			n.mu.Lock()
			k := len(n.pending)
			n.mu.Unlock()
			if k > 0 {
				n.answer([]byte("late"), nil)
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()
	b, err := a.Wait(5 * time.Second)
	if err != nil || string(b) != "late" {
		t.Fatalf("%q, %v", b, err)
	}
	// a fetch that never answers ends with the wait
	b2 := New("/g", func(string, func([]byte, error)) {}, nil)
	if _, err := b2.Wait(20 * time.Millisecond); !errors.Is(err, ErrLoading) {
		t.Errorf("the wait: %v", err)
	}
	// and one that fails says why
	c := New("/h", func(_ string, done func([]byte, error)) { go done(nil, errors.New("offline")) }, nil)
	if _, err := c.Wait(5 * time.Second); err == nil || !strings.Contains(err.Error(), "offline") {
		t.Errorf("a failed fetch: %v", err)
	}
}

// TestAsTheMathProvider: through go-slides' whole render, as the playground's
// worker does it. Before the font a formula is its source, in math-error; the
// bytes arrive, the provider is registered again, and the next render is
// typeset. A failed fetch leaves the source, and the next render fetches again.
func TestAsTheMathProvider(t *testing.T) {
	t.Cleanup(embedded.Register)
	const deck = "# Energy\n\n$E=mc^2$\n"
	n := &netStub{}
	var a *Asset
	a = New("/playground/fonts/stix-math.otf", n.fetch, func([]byte) { mathfont.Register(a) })
	mathfont.Register(a)

	render := func() string {
		t.Helper()
		h, _, err := slides.HTML([]byte(deck), slides.Config{})
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	h := render()
	if !strings.Contains(h, `class="math-error"`) || strings.Contains(h, `class="slides-math"`) || !strings.Contains(h, "$E=mc^2$") {
		t.Fatalf("before the font: %.300s", h)
	}
	if a.Fetches() != 1 {
		t.Fatalf("%d fetches for the first render", a.Fetches())
	}
	// the download fails: still the source, no render is asked for, and the next
	// render fetches again
	n.answer(nil, errors.New("offline"))
	h = render()
	if !strings.Contains(h, `class="math-error"`) || a.Fetches() != 2 {
		t.Fatalf("after a failure: %d fetches, %.300s", a.Fetches(), h)
	}
	n.answer(math.DefaultFont(), nil)
	h = render()
	if strings.Contains(h, `class="math-error"`) || !strings.Contains(h, `class="slides-math"`) {
		t.Fatalf("after the font: %.300s", h)
	}
	if a.Fetches() != 2 {
		t.Errorf("a typeset render fetched again: %d", a.Fetches())
	}
}
