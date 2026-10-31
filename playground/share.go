package playground

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"io"
	"strings"
	"unicode/utf8"
)

// sharePrefix starts a shared deck's fragment: "#deck=" followed by the deck,
// deflated and base64url-encoded. The fragment never leaves the browser: a
// shared link sends nothing to any server, GitHub Pages included.
const sharePrefix = "deck="

// MaxShare bounds the fragment a link may carry, encoded; browsers and chat
// tools cut long URLs well before this anyway.
const MaxShare = 64 << 10

// EncodeShare returns the fragment (without "#") that carries src.
func EncodeShare(src string) (string, error) {
	if len(src) > MaxSource {
		return "", errf(KErrShareTooLargeDeck)
	}
	var buf bytes.Buffer
	w, _ := flate.NewWriter(&buf, flate.BestCompression)
	_, _ = w.Write([]byte(src))
	_ = w.Close()
	s := sharePrefix + base64.RawURLEncoding.EncodeToString(buf.Bytes())
	if len(s) > MaxShare {
		return "", errf(KErrShareTooLargeLink)
	}
	return s, nil
}

// DecodeShare reads the deck a fragment carries ("#" optional). It refuses a
// fragment that is not a shared deck, that does not decode, that inflates
// past MaxSource (a small link must not expand to gigabytes), or that is not
// UTF-8 text.
func DecodeShare(fragment string) (string, error) {
	fragment = strings.TrimPrefix(fragment, "#")
	data, ok := strings.CutPrefix(fragment, sharePrefix)
	if !ok {
		return "", errf(KErrShareNoDeck)
	}
	if len(data) > MaxShare {
		return "", errf(KErrShareTooLong)
	}
	raw, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return "", errf(KErrShareDamaged)
	}
	r := flate.NewReader(bytes.NewReader(raw))
	defer r.Close()
	out, err := io.ReadAll(io.LimitReader(r, MaxSource+1))
	if err != nil {
		return "", errf(KErrShareDamaged)
	}
	if len(out) > MaxSource {
		return "", errf(KErrShareLarger)
	}
	if !utf8.Valid(out) {
		return "", errf(KErrShareNotText)
	}
	return string(out), nil
}
