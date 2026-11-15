package emojipack

import (
	"bytes"
	"compress/gzip"
	"errors"
	"strings"
	"testing"

	"github.com/go-slides/slides"
	"github.com/go-slides/slides/emoji"
	"github.com/go-slides/slides/emoji/embedded"

	"github.com/go-slides/go-slides.github.io/playground/lazyasset"
)

const smile = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 36 36"><circle cx="18" cy="18" r="18" fill="#FFCC4D"/></svg>`

func gz(t *testing.T, raw []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(raw)
	w.Close()
	return b.Bytes()
}

func TestRoundTrip(t *testing.T) {
	b, err := Encode([]Entry{{"1f604", smile}, {"1f469-200d-1f4bb", "<?xml version='1.0'?>" + smile}})
	if err != nil {
		t.Fatal(err)
	}
	p, err := Decode(b)
	if err != nil || p.Len() != 2 {
		t.Fatalf("%v %v", p, err)
	}
	if s, ok := p.SVGByStem("1f604"); !ok || s != smile {
		t.Errorf("%q %v", s, ok)
	}
	if _, ok := p.SVGByStem("1f605"); ok {
		t.Error("a stem that is not in the pack")
	}
}

func TestEncodeRefuses(t *testing.T) {
	for name, es := range map[string][]Entry{
		"none":      nil,
		"bad stem":  {{"1F604", smile}},
		"no stem":   {{"", smile}},
		"repeated":  {{"1f604", smile}, {"1f604", smile}},
		"not svg":   {{"1f604", "<html>"}},
		"too large": {{"1f604", "<svg" + strings.Repeat("x", emoji.MaxSVGBytes)}},
	} {
		if _, err := Encode(es); err == nil {
			t.Errorf("%s: encoded", name)
		}
	}
}

// TestDecodeRefuses: the bytes come from the network.
func TestDecodeRefuses(t *testing.T) {
	good, _ := Encode([]Entry{{"1f604", smile}})
	rec := func(stem, svg string) []byte {
		return []byte(Magic + "\x01" + string(rune(len(stem))) + stem + string(rune(len(svg))) + svg)
	}
	for name, b := range map[string][]byte{
		"empty":           nil,
		"not gzip":        []byte("hello"),
		"no magic":        gz(t, []byte("nothing")),
		"truncated gzip":  good[:len(good)-6],
		"truncated pack":  gz(t, []byte(Magic+"\x05\x04")),
		"zero entries":    gz(t, []byte(Magic+"\x00")),
		"huge count":      gz(t, []byte(Magic+"\xff\xff\x7f")),
		"a stem in caps":  gz(t, rec("1F604", smile)),
		"an html graphic": gz(t, rec("1f604", "<html>")),
		"trailing bytes":  gz(t, append(rec("1f604", smile), 'x')),
		"too large":       append(good, make([]byte, MaxFileBytes)...),
	} {
		if _, err := Decode(b); err == nil {
			t.Errorf("%s: decoded", name)
		}
	}
	// the planted positive: the same record, right, decodes
	if _, err := Decode(gz(t, rec("1f604", smile))); err != nil {
		t.Errorf("the control pack: %v", err)
	}
	// a bomb: past the inflated bound, refused without holding it
	if _, err := Decode(gz(t, make([]byte, MaxInflated+1024))); err == nil || !strings.Contains(err.Error(), "inflates past") {
		t.Errorf("a bomb: %v", err)
	}
}

func TestHasEmoji(t *testing.T) {
	for s, want := range map[string]bool{
		"plain text":       false,
		"a $ and a : b":    false,
		"hello 😄":          true,
		":smile: now":      true,
		"1️⃣ keycap":       true,
		"digit 1 alone":    false,
		"accents éàü only": false,
	} {
		if HasEmoji(s) != want {
			t.Errorf("HasEmoji(%q) = %v", s, !want)
		}
	}
}

// netStub answers by hand.
type netStub struct{ pending []func([]byte, error) }

func (n *netStub) fetch(_ string, done func([]byte, error)) { n.pending = append(n.pending, done) }
func (n *netStub) answer(b []byte, err error) {
	d := n.pending[0]
	n.pending = n.pending[1:]
	d(b, err)
}

// TestAsTheEmojiProvider: through go-slides' whole render, as the worker does.
// Before the pack an emoji is the text written; the pack arrives, the provider is
// registered again, and the next render draws a picture. A failed fetch leaves
// the text and the next render asks again; a pack that is not one leaves the text
// without asking again.
func TestAsTheEmojiProvider(t *testing.T) {
	t.Cleanup(embedded.Register)
	const deck = "# Hi 😄\n\nand :smile:\n"
	render := func() string {
		t.Helper()
		h, _, err := slides.HTML([]byte(deck), slides.Config{})
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	hasImg := func(h string) bool { return strings.Contains(h, "data:image/svg+xml;base64,") }

	n := &netStub{}
	var a *lazyasset.Asset
	var prov *Provider
	a = lazyasset.New("/playground/emoji/twemoji.pack", n.fetch, func([]byte) { emoji.Register(prov) })
	prov = NewProvider(a.Bytes)
	emoji.Register(prov)

	h := render()
	if hasImg(h) || !strings.Contains(h, "😄") || !strings.Contains(h, ":smile:") {
		t.Fatalf("before the pack: %.400s", h)
	}
	if a.Fetches() != 1 {
		t.Fatalf("%d fetches", a.Fetches())
	}
	n.answer(nil, errors.New("offline"))
	if hasImg(render()) || a.Fetches() != 2 {
		t.Fatalf("after a failure: %d fetches", a.Fetches())
	}
	pack, err := Encode([]Entry{{"1f604", smile}})
	if err != nil {
		t.Fatal(err)
	}
	n.answer(pack, nil)
	h = render()
	if !hasImg(h) || strings.Contains(h, ":smile:") {
		t.Fatalf("after the pack: %.600s", h)
	}
	if a.Fetches() != 2 {
		t.Errorf("a drawn render fetched again: %d", a.Fetches())
	}

	// bytes that are not a pack: text, the reason kept, no new fetch
	n2 := &netStub{}
	var a2 *lazyasset.Asset
	var p2 *Provider
	a2 = lazyasset.New("/x", n2.fetch, func([]byte) { emoji.Register(p2) })
	p2 = NewProvider(a2.Bytes)
	emoji.Register(p2)
	render()
	n2.answer([]byte("not a pack"), nil)
	if h := render(); hasImg(h) || p2.Err() == nil {
		t.Errorf("a bad pack drew: %v", p2.Err())
	}
	render()
	if a2.Fetches() != 1 {
		t.Errorf("a bad pack was fetched again: %d", a2.Fetches())
	}
}
