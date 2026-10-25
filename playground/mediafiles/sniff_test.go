package mediafiles

import (
	"net/http"
	"strings"
	"testing"
)

// imageOracle is what Type used before: net/http's sniffer, its image
// answers only.
func imageOracle(data []byte) string {
	if s := http.DetectContentType(data); strings.HasPrefix(s, "image/") {
		return s
	}
	return ""
}

// sniffImage answers as net/http's DetectContentType does, image types
// only, on each signature, its truncations and near misses, and the other
// patterns the standard tries first.
func TestSniffImageMatchesNetHTTP(t *testing.T) {
	seeds := []string{
		"\x00\x00\x01\x00rest", "\x00\x00\x02\x00", "BMxx", "GIF87a..", "GIF89a", "RIFF\x01\x02\x03\x04WEBPVP8 ",
		"RIFF\x01\x02\x03\x04WEBQVP8 ", "\x89PNG\r\n\x1a\n\x00", "\xFF\xD8\xFF\xE0", "\xFF\xD8\xFE",
		"<html>", "  <!DOCTYPE HTML>", "<?xml", "%PDF-1.7", "%!PS-Adobe-", "\xFE\xFF\x00\x00", "\xFF\xFE\x00\x00",
		"\xEF\xBB\xBF\x00", "", "B", "GIF8", "plain text", "\x00\x00\x03\x00",
	}
	for _, s := range seeds {
		for n := 0; n <= len(s); n++ {
			d := []byte(s[:n])
			if got, want := sniffImage(d), imageOracle(d); got != want {
				t.Errorf("sniffImage(%q) = %q, net/http %q", d, got, want)
			}
		}
	}
}

func FuzzSniffImage(f *testing.F) {
	for _, s := range []string{"\x89PNG\r\n\x1a\n", "\xFF\xD8\xFF", "GIF89a", "RIFF0000WEBPVP8", "BM", "\x00\x00\x01\x00", "<p>"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, d []byte) {
		if got, want := sniffImage(d), imageOracle(d); got != want {
			t.Fatalf("sniffImage(%q) = %q, net/http %q", d, got, want)
		}
	})
}

// Type keeps its answers: an image's bytes win over its extension, other
// media keep theirs.
func TestTypeSniffs(t *testing.T) {
	for _, c := range []struct {
		p    string
		data string
		want string
	}{
		{"a.png", "\xFF\xD8\xFF\xE0", "image/jpeg"},
		{"a.jpg", "\x89PNG\r\n\x1a\n", "image/png"},
		{"a.png", "not an image", "image/png"},
		{"a.svg", "\x89PNG\r\n\x1a\n", "image/svg+xml"},
		{"a.mp4", "\x89PNG\r\n\x1a\n", "video/mp4"},
		{"a.bin", "x", "application/octet-stream"},
	} {
		if got := Type(c.p, []byte(c.data)); got != c.want {
			t.Errorf("Type(%q, %q) = %q, want %q", c.p, c.data, got, c.want)
		}
	}
}
