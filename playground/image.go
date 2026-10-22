package playground

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"path"
	"strings"
)

// MaxImage bounds an image dropped or pasted into the editor. It goes into
// the deck as a data: URI, a third larger than the file, and the deck is
// rendered on every keystroke, so it stays well under MaxSource.
const MaxImage = 256 << 10

// imageTypes are the formats a browser shows in an <img> and a CSS
// background, by the type their bytes say they are (http.DetectContentType,
// the WHATWG MIME sniffing algorithm), never by the name or the type the
// file claims.
var imageTypes = map[string]bool{
	"image/png": true, "image/jpeg": true, "image/gif": true, "image/webp": true,
}

// ImageMarkdown is the Markdown of an image dropped or pasted into the
// editor: ![name](data:...), the file's bytes in the deck itself, so the
// deck stays one file that works offline, in the preview's sandbox and in
// the downloaded page. An SVG is accepted when it starts as one; in an
// image or a background a browser runs none of its scripts.
func ImageMarkdown(name string, data []byte) (string, error) {
	if len(data) > MaxImage {
		return "", fmt.Errorf("%s is %d KiB; images up to %d KiB go into the deck — link a larger one by its https: URL", name, len(data)>>10, MaxImage>>10)
	}
	mime := http.DetectContentType(data)
	if !imageTypes[mime] {
		head := bytes.TrimLeft(bytes.TrimPrefix(data[:min(len(data), 512)], []byte("\xef\xbb\xbf")), " \t\r\n")
		if !(bytes.HasPrefix(head, []byte("<svg")) || bytes.HasPrefix(head, []byte("<?xml"))) || !bytes.Contains(data[:min(len(data), 4096)], []byte("<svg")) {
			return "", fmt.Errorf("%s is not a PNG, JPEG, GIF, WebP or SVG image", name)
		}
		mime = "image/svg+xml"
	}
	// The alternative text: the file's name without its extension, with
	// nothing that would end the Markdown of the image.
	alt := strings.TrimSuffix(path.Base(name), path.Ext(name))
	alt = strings.Map(func(r rune) rune {
		if strings.ContainsRune("[]()\\\n\r<>", r) {
			return -1
		}
		return r
	}, alt)
	return "![" + alt + "](data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data) + ")", nil
}
