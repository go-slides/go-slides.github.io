// Package serve answers videoproto requests with go-avkit: the worker's side of
// the protocol. It is the only package of the playground that links a container
// reader and a decoder for video; the page's module links none and runs this in
// a worker of its own (cmd/playground-video).
package serve

import (
	"errors"
	"sync"

	"github.com/go-slides/go-slides.github.io/playground/videoproto"
	"github.com/go-slides/slides/video"
	"github.com/go-slides/slides/video/avkit"
)

// keep is how many files a Server holds: the project's videos that the
// renders of a page ask about, the most recent ones. Two 64 MiB files at most.
const keep = 2

// Server holds the files it was sent and answers requests about them.
type Server struct {
	mu    sync.Mutex
	files []held // most recent last
}

type held struct {
	key  string
	file *video.File
}

// Handle answers one request. payload is the file when req.HasData, and the
// picture of a poster's answer is the second result. The file is checked
// against the limits (video.Open) before anything is decoded, whatever the
// client said of it: the key and the size are not taken for granted.
func (s *Server) Handle(req videoproto.Request, payload []byte) (videoproto.Response, []byte) {
	res := videoproto.Response{ID: req.ID}
	lim := videoproto.Clamp(req.Limits)
	if req.Op != videoproto.OpPoster && req.Op != videoproto.OpProbe {
		res.Code, res.Error = video.CodeSyntax, "unknown operation "+quote(req.Op)
		return res, nil
	}
	file, err := s.file(req, payload, lim)
	if err != nil {
		if errors.Is(err, errNotHeld) {
			res.Code, res.Error = videoproto.CodeNotHeld, "the worker does not hold this file"
			return res, nil
		}
		res.Code, res.Error = codeOf(err)
		return res, nil
	}
	res.Format, res.MIME, res.Codec = file.Format, file.MIME, file.Codec
	res.Width, res.Height, res.Duration, res.Decodable = file.Width, file.Height, file.Duration, file.Decodable
	if req.Op == videoproto.OpProbe {
		return res, nil
	}
	jpg, w, h, err := file.Poster(req.T, lim)
	if err != nil {
		res.Code, res.Error = codeOf(err)
		return res, nil
	}
	res.Width, res.Height, res.Image = w, h, "image/jpeg"
	return res, jpg
}

var errNotHeld = errors.New("not held")

// file is the checked file a request names: the payload, opened and kept, or
// the copy held under the key.
func (s *Server) file(req videoproto.Request, payload []byte, lim video.Limits) (*video.File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !req.HasData {
		for _, h := range s.files {
			if h.key == req.Key {
				return h.file, nil
			}
		}
		return nil, errNotHeld
	}
	// with go-avkit, never with the registered provider, which may be the
	// client of this very worker
	f, err := video.OpenWith(avkit.Provider{}, payload, lim)
	if err != nil {
		return nil, err
	}
	// a file is held under the key the client gave it: the worker does not
	// trust it to be the file's checksum, only to name what to look up
	s.files = append(s.files[:0:0], append(without(s.files, req.Key), held{req.Key, f})...)
	if len(s.files) > keep {
		s.files = s.files[len(s.files)-keep:]
	}
	return f, nil
}

func without(h []held, key string) []held {
	out := make([]held, 0, len(h))
	for _, x := range h {
		if x.key != key {
			out = append(out, x)
		}
	}
	return out
}

// codeOf is a refusal's code and message; an error with no code is
// video.syntax.
func codeOf(err error) (code, msg string) {
	var e *video.Error
	if errors.As(err, &e) {
		return e.Code, e.Msg
	}
	return video.CodeSyntax, err.Error()
}

func quote(s string) string {
	if len(s) > 32 {
		s = s[:32] + "..."
	}
	return `"` + s + `"`
}
