// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

// Package emojipack is the Twemoji graphics as ONE file the render worker
// fetches when a deck has an emoji (go-slides' emoji.Provider; the module is
// built with slides_noemoji and links none): a gzip stream of
//
//	"emojipack1\n", count,  then count x (stem, svg)
//
// where count and each length are unsigned varints and a stem is how twemoji.js
// names a graphic ("1f604", "1f469-200d-1f4bb"). One file, because 4,009 files
// would be 4,009 requests; one gzip stream over all of them, because the SVGs
// share most of their text: 10.1 MB of SVG is 1.74 MB that way, 4.2 MB with each
// SVG deflated alone (so that a graphic could be read without the others), and
// the embedded archive (go-icons/twemoji, which inflates a block at a time) is
// 2.16 MB. The file is compressed inside, because a host that serves a binary
// type does not compress it. Reading inflates it whole, once, into one string the
// graphics are slices of.
//
// The bytes come from the network, so they are untrusted: Decode bounds what it
// inflates, how many graphics it reads, and each graphic (go-slides' emoji
// package takes an SVG of at most MaxSVGBytes that starts "<svg" or "<?xml";
// a pack holding one that is not is refused whole).
package emojipack

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/go-slides/slides/emoji"
)

// Magic starts the inflated stream.
const Magic = "emojipack1\n"

// The bounds of a pack: the file itself, what it inflates to, how many graphics
// it holds (Twemoji has 4,009), and the longest stem.
const (
	MaxFileBytes = 8 << 20
	MaxInflated  = 16 << 20
	MaxEntries   = 8192
	MaxStemBytes = 128
)

// An Entry is a graphic: its stem and its SVG document.
type Entry struct{ Stem, SVG string }

// Encode writes the pack of entries (the build's side). Stems must be
// distinct and every SVG one the emoji package takes.
func Encode(entries []Entry) ([]byte, error) {
	if len(entries) == 0 || len(entries) > MaxEntries {
		return nil, fmt.Errorf("emojipack: %d entries", len(entries))
	}
	var raw bytes.Buffer
	raw.WriteString(Magic)
	var v [binary.MaxVarintLen64]byte
	put := func(n int) { raw.Write(v[:binary.PutUvarint(v[:], uint64(n))]) }
	put(len(entries))
	seen := make(map[string]bool, len(entries))
	for _, e := range entries {
		if !validStem(e.Stem) || seen[e.Stem] {
			return nil, fmt.Errorf("emojipack: stem %q is invalid or repeated", e.Stem)
		}
		if !validSVG(e.SVG) {
			return nil, fmt.Errorf("emojipack: the graphic of %q is not an SVG of at most %d bytes", e.Stem, emoji.MaxSVGBytes)
		}
		seen[e.Stem] = true
		put(len(e.Stem))
		raw.WriteString(e.Stem)
		put(len(e.SVG))
		raw.WriteString(e.SVG)
	}
	var out bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&out, gzip.BestCompression)
	zw.Write(raw.Bytes())
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// A Pack is a decoded file.
type Pack struct{ svg map[string]string }

// Len is the number of graphics.
func (p *Pack) Len() int { return len(p.svg) }

// SVGByStem is the graphic of a stem, as go-slides' emoji.Provider asks.
func (p *Pack) SVGByStem(stem string) (string, bool) {
	s, ok := p.svg[stem]
	return s, ok
}

func validStem(s string) bool {
	if s == "" || len(s) > MaxStemBytes {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c == '-') {
			return false
		}
	}
	return true
}

func validSVG(s string) bool {
	if len(s) == 0 || len(s) > emoji.MaxSVGBytes {
		return false
	}
	t := strings.TrimLeft(s, " \t\r\n\xef\xbb\xbf")
	return strings.HasPrefix(t, "<svg") || strings.HasPrefix(t, "<?xml")
}

// Decode reads a pack, or says why it is not one.
func Decode(b []byte) (*Pack, error) {
	if len(b) > MaxFileBytes {
		return nil, fmt.Errorf("emojipack: the file is %d bytes, over the %d allowed", len(b), MaxFileBytes)
	}
	zr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("emojipack: not a gzip stream: %w", err)
	}
	raw, err := io.ReadAll(io.LimitReader(zr, MaxInflated+1))
	if err != nil {
		return nil, fmt.Errorf("emojipack: %w", err)
	}
	if len(raw) > MaxInflated {
		return nil, fmt.Errorf("emojipack: it inflates past %d bytes", MaxInflated)
	}
	if !bytes.HasPrefix(raw, []byte(Magic)) {
		return nil, errors.New("emojipack: not a pack (no magic)")
	}
	s := string(raw)
	pos := len(Magic)
	uv := func() (int, error) {
		v, n := binary.Uvarint(raw[pos:])
		if n <= 0 || v > MaxInflated {
			return 0, errors.New("emojipack: a length is damaged")
		}
		pos += n
		return int(v), nil
	}
	str := func(max int) (string, error) {
		n, err := uv()
		if err != nil {
			return "", err
		}
		if n > max || n > len(raw)-pos {
			return "", errors.New("emojipack: a record runs past the end or its bound")
		}
		out := s[pos : pos+n]
		pos += n
		return out, nil
	}
	count, err := uv()
	if err != nil {
		return nil, err
	}
	if count == 0 || count > MaxEntries {
		return nil, fmt.Errorf("emojipack: %d entries", count)
	}
	m := make(map[string]string, count)
	for i := 0; i < count; i++ {
		stem, err := str(MaxStemBytes)
		if err != nil {
			return nil, err
		}
		svg, err := str(emoji.MaxSVGBytes)
		if err != nil {
			return nil, err
		}
		if !validStem(stem) || !validSVG(svg) {
			return nil, fmt.Errorf("emojipack: entry %d (%q) is not a stem and an SVG", i, stem)
		}
		if _, dup := m[stem]; dup {
			return nil, fmt.Errorf("emojipack: %q is repeated", stem)
		}
		m[stem] = svg
	}
	if pos != len(raw) {
		return nil, errors.New("emojipack: bytes after the last entry")
	}
	return &Pack{svg: m}, nil
}

// HasEmoji reports whether text may draw an emoji: a character go-slides reads
// as one (emoji.Cut) or a :shortcode: it knows. An export, which has no page to
// render again, waits for the pack only when this is true.
func HasEmoji(text string) bool {
	for i := 0; i < len(text); {
		if emoji.Cut(text[i:]) > 0 {
			return true
		}
		_, w := utf8.DecodeRuneInString(text[i:])
		i += w
	}
	found := false
	emoji.Replace(text, func(name, seq string) string { found = true; return "" })
	return found
}

// A Provider is a go-slides emoji.Provider over a pack that arrives later: src
// gives the file's bytes, or an error while they are not here (and starts their
// fetch; a lazyasset.Asset's Bytes). The pack is decoded on the first graphic
// asked for after the bytes are here, once; bytes that are not a pack leave the
// emoji as text, with the reason in Err.
type Provider struct {
	src func() ([]byte, error)

	mu   sync.Mutex
	pack *Pack
	bad  error
}

// NewProvider returns a Provider that reads its pack from src.
func NewProvider(src func() ([]byte, error)) *Provider { return &Provider{src: src} }

// SVGByStem implements emoji.Provider.
func (p *Provider) SVGByStem(stem string) (string, bool) {
	p.mu.Lock()
	pk, bad := p.pack, p.bad
	p.mu.Unlock()
	if pk == nil && bad == nil {
		b, err := p.src()
		if err != nil {
			return "", false // not here yet: the text stays, and the fetch is under way
		}
		p.mu.Lock()
		if p.pack == nil && p.bad == nil {
			p.pack, p.bad = nil, nil
			if d, err := Decode(b); err != nil {
				p.bad = err
			} else {
				p.pack = d
			}
		}
		pk = p.pack
		p.mu.Unlock()
	}
	if pk == nil {
		return "", false
	}
	return pk.SVGByStem(stem)
}

// Err is why the bytes were refused, or nil.
func (p *Provider) Err() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.bad
}
