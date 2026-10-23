// Package mediafiles is what the playground's renderer and its export
// engine (a module of its own, kept small) share about a project's files:
// which are media, their types, putting them into a page as data: URIs,
// and the bounds a project's files are checked against before a build
// reads them.
package mediafiles

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"
)

// The bounds of what a build reads besides the deck: its text files (sent
// on every keystroke to the renderer) and its media (sent to an export).
const (
	MaxText  = 2 << 20
	MaxMedia = 64 << 20
)

// types are the media a project holds, by extension.
var types = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif",
	".webp": "image/webp", ".avif": "image/avif", ".svg": "image/svg+xml", ".ico": "image/x-icon",
	".mp4": "video/mp4", ".webm": "video/webm", ".mov": "video/quicktime", ".m4v": "video/mp4",
	".ogv": "video/ogg", ".mp3": "audio/mpeg", ".m4a": "audio/mp4", ".ogg": "audio/ogg",
	".wav": "audio/wav", ".woff": "font/woff", ".woff2": "font/woff2", ".ttf": "font/ttf", ".otf": "font/otf",
}

// IsMedia reports whether a project file is a medium (bytes), not text the
// editor opens and a build reads.
func IsMedia(p string) bool {
	_, ok := types[strings.ToLower(path.Ext(p))]
	return ok
}

// Kind is a medium's type by its extension alone ("" for none).
func Kind(p string) string { return types[strings.ToLower(path.Ext(p))] }

// Type is the type a medium is posted and inlined with: by its extension,
// except that bytes that say they are an image win (a .png that is a JPEG).
// An SVG keeps its type: shown as an image, it runs nothing.
func Type(p string, data []byte) string {
	t := Kind(p)
	if strings.HasPrefix(t, "image/") && t != "image/svg+xml" {
		if s := http.DetectContentType(data); strings.HasPrefix(s, "image/") {
			return s
		}
	}
	if t == "" {
		return "application/octet-stream"
	}
	return t
}

// EscapeRef writes a path as a URL path: the characters a Markdown link or
// an attribute would end at are escaped.
func EscapeRef(p string) string {
	var b strings.Builder
	for _, r := range p {
		if strings.ContainsRune(" ()<>\"'#?%", r) {
			fmt.Fprintf(&b, "%%%02X", r)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Inline puts the media a page or a style sheet names by relative path into
// it as data: URIs: an attribute's value and a CSS url(), quoted or not,
// as the path, ./path, or ../path (a theme in themes/ naming media/x). It
// is what the preview's frame does with blob: URLs (the playground's
// mediaScript), for a document that leaves the page: Save as, the exports.
func Inline(page string, media map[string][]byte) string {
	if len(media) == 0 {
		return page
	}
	names := make([]string, 0, len(media))
	for k := range media {
		names = append(names, k)
	}
	// longest first: media/a.png.bak is not media/a.png
	sort.Slice(names, func(i, j int) bool { return len(names[i]) > len(names[j]) })
	var pairs []string
	for _, n := range names {
		uri := "data:" + Type(n, media[n]) + ";base64," + base64.StdEncoding.EncodeToString(media[n])
		refs := []string{n}
		if e := EscapeRef(n); e != n {
			refs = append(refs, e)
		}
		for _, r := range refs {
			for _, ref := range []string{r, "./" + r, "../" + r} {
				for _, q := range []string{`"`, `'`, `&quot;`, `&#34;`, `(`} {
					end := q
					if q == "(" {
						end = ")"
					}
					pairs = append(pairs, q+ref+end, q+uri+end)
				}
			}
		}
	}
	return strings.NewReplacer(pairs...).Replace(page)
}

// Check checks the project files a build is handed, by their paths beside
// the deck at deckPath: every one under the deck's folder and not the deck,
// within the bounds; it returns the folders as the editor package lists
// them. The files come from the page, but a worker takes nothing for
// granted.
func Check(deckPath string, files map[string]string, dirs []string, media map[string][]byte) (map[string][]string, error) {
	root := path.Dir(deckPath) + "/"
	inside := func(k string) bool {
		c := path.Clean(k)
		return strings.HasPrefix(c, root) && c == k && c != deckPath
	}
	total := 0
	for k, v := range files {
		if !inside(k) {
			return nil, fmt.Errorf("%q is not a file of the project", k)
		}
		if total += len(v); total > MaxText {
			return nil, fmt.Errorf("the project's text files are over %d KiB", MaxText>>10)
		}
	}
	total = 0
	for k, v := range media {
		if path.IsAbs(k) || path.Clean(k) != k || strings.HasPrefix(k, "../") || k == ".." {
			return nil, fmt.Errorf("%q is not a medium of the project", k)
		}
		if total += len(v); total > MaxMedia {
			return nil, fmt.Errorf("the project's media are over %d MiB", MaxMedia>>20)
		}
	}
	listed := map[string][]string{}
	for _, d := range dirs {
		if c := path.Clean(d); strings.HasPrefix(c, root) {
			listed[c] = nil
		}
	}
	return listed, nil
}
