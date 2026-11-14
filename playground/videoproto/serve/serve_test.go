package serve

import (
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/videoproto"
	"github.com/go-slides/slides/video"
)

type panicker struct{ onPoster bool }

func (p panicker) Probe(data []byte, format string, lim video.Limits) (video.Probed, error) {
	if !p.onPoster {
		panic("a defect in the reader")
	}
	return video.Probed{Codec: "avc1", Width: 64, Height: 48, Decodable: true}, nil
}

func (p panicker) Poster(video.Probed, []byte, float64, video.Limits) ([]byte, int, int, error) {
	panic("a defect in the decoder")
}

func ftyp() []byte {
	b := make([]byte, 16)
	b[3] = 16
	copy(b[4:], "ftypisom")
	return b
}

func TestAPanicIsACodedErrorAndTheServerKeepsServing(t *testing.T) {
	for name, p := range map[string]panicker{"probe": {}, "poster": {onPoster: true}} {
		s := &Server{Provider: p}
		op := videoproto.OpProbe
		if p.onPoster {
			op = videoproto.OpPoster
		}
		res, img := s.Handle(videoproto.Request{ID: 7, Op: op, Key: "k", HasData: true}, ftyp())
		if res.ID != 7 || res.Code != video.CodeSyntax || res.Error == "" || img != nil {
			t.Fatalf("%s: %+v", name, res)
		}
		// the next request is answered: the worker did not end
		res, _ = s.Handle(videoproto.Request{ID: 8, Op: "unknown"}, nil)
		if res.ID != 8 || res.Code != video.CodeSyntax {
			t.Fatalf("%s: the next request: %+v", name, res)
		}
	}
}
