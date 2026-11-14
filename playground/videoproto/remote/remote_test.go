package remote_test

import (
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/videoproto"
	"github.com/go-slides/go-slides.github.io/playground/videoproto/remote"
	"github.com/go-slides/go-slides.github.io/playground/videoproto/serve"
	"github.com/go-slides/slides"
	"github.com/go-slides/slides/deck"
	"github.com/go-slides/slides/vfs"
	"github.com/go-slides/slides/video"
	"github.com/go-slides/slides/video/avkit"
)

// loop is a Transport that is the worker in the same process: the request and
// the response go through the protocol's types only, as a message does.
type loop struct {
	mu       sync.Mutex
	srv      *serve.Server
	calls    []videoproto.Request
	payloads []int
	err      error
}

func (l *loop) Call(req videoproto.Request, payload []byte) (videoproto.Response, []byte, error) {
	l.mu.Lock()
	l.calls = append(l.calls, req)
	l.payloads = append(l.payloads, len(payload))
	err := l.err
	l.mu.Unlock()
	if err != nil {
		return videoproto.Response{}, nil, err
	}
	res, img := l.srv.Handle(req, append([]byte(nil), payload...))
	return res, img, nil
}

func clip(t testing.TB) []byte {
	t.Helper()
	b, err := os.ReadFile("../../testdata/clip.mp4")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func code(err error) string {
	var e *video.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

// the same deck, rendered with go-avkit in the process and with the worker's
// answers: the page is the same, byte for byte.
func TestRenderThroughTheWorkerIsByteIdentical(t *testing.T) {
	data := clip(t)
	src := "# Clip\n\n![my clip](clip.mp4#t=0.3)\n\n---\n\n![video muted loop](clip.mp4)\n"
	render := func() string {
		page, _, err := slides.HTML([]byte(src), slides.Config{Parse: deck.Options{FS: vfs.NewMap(map[string][]byte{"clip.mp4": data})}})
		if err != nil {
			t.Fatal(err)
		}
		return page
	}
	video.Register(avkit.Provider{})
	t.Cleanup(func() { video.Register(avkit.Provider{}) })
	direct := render()
	if !strings.Contains(direct, "data:image/jpeg;base64,") {
		t.Fatal("the direct render has no poster")
	}
	l := &loop{srv: &serve.Server{}}
	video.Register(remote.New(l))
	if got := render(); got != direct {
		t.Fatalf("the render through the worker differs (%d bytes against %d)", len(got), len(direct))
	}
	if len(l.calls) == 0 {
		t.Fatal("the worker was not asked")
	}
}

func TestTheFileCrossesOnceAndAnswersAreKept(t *testing.T) {
	data := clip(t)
	l := &loop{srv: &serve.Server{}}
	p := remote.New(l)
	lim := video.DefaultLimits
	pr, err := p.Probe(data, "mp4", lim)
	if err != nil || pr.Width <= 0 || pr.Codec == "" {
		t.Fatalf("%+v %v", pr, err)
	}
	if _, _, _, err := p.Poster(pr, data, 0, lim); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := p.Poster(pr, data, 0.3, lim); err != nil {
		t.Fatal(err)
	}
	// 3 calls: probe, poster 0, poster 0.3; the file went with the first only
	if len(l.calls) != 3 || l.payloads[0] != len(data) || l.payloads[1] != 0 || l.payloads[2] != 0 {
		t.Fatalf("calls %d, payloads %v", len(l.calls), l.payloads)
	}
	// asking again is answered from what is kept
	if _, err := p.Probe(data, "mp4", lim); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := p.Poster(pr, data, 0, lim); err != nil {
		t.Fatal(err)
	}
	if len(l.calls) != 3 {
		t.Fatalf("a repeated question reached the worker: %d calls", len(l.calls))
	}
}

func TestAWorkerThatLostTheFileIsSentItAgain(t *testing.T) {
	data := clip(t)
	l := &loop{srv: &serve.Server{}}
	p := remote.New(l)
	lim := video.DefaultLimits
	pr, err := p.Probe(data, "mp4", lim)
	if err != nil {
		t.Fatal(err)
	}
	l.srv = &serve.Server{} // the worker was replaced
	l.calls, l.payloads = nil, nil
	if _, _, _, err := p.Poster(pr, data, 0, lim); err != nil {
		t.Fatalf("the poster after the worker was replaced: %v", err)
	}
	if len(l.calls) != 2 || l.payloads[0] != 0 || l.payloads[1] != len(data) {
		t.Fatalf("calls %d, payloads %v: want one without the file, one with", len(l.calls), l.payloads)
	}
}

func TestRefusalsKeepTheirCodes(t *testing.T) {
	l := &loop{srv: &serve.Server{}}
	p := remote.New(l)
	lim := video.DefaultLimits
	if _, err := p.Probe([]byte("not a video at all, no"), "mp4", lim); code(err) != video.CodeSyntax {
		t.Errorf("garbage: %v", err)
	}
	lim.MaxBytes = 100
	if _, err := p.Probe(clip(t), "mp4", lim); code(err) != video.CodeLimit {
		t.Errorf("too large: %v", err)
	}
	// through Open, as slides asks
	video.Register(remote.New(l))
	t.Cleanup(func() { video.Register(avkit.Provider{}) })
	if _, err := video.Open(make([]byte, video.DefaultLimits.MaxBytes+1), video.DefaultLimits); code(err) != video.CodeLimit {
		t.Errorf("Open, 64 MiB and a byte: %v", err)
	}
}

func TestATransportFailureIsUnavailable(t *testing.T) {
	l := &loop{srv: &serve.Server{}, err: errors.New("the worker could not be started")}
	_, err := remote.New(l).Probe(clip(t), "mp4", video.DefaultLimits)
	if code(err) != video.CodeUnavailable || !strings.Contains(err.Error(), "could not be started") {
		t.Fatalf("%v", err)
	}
}

func TestAHostCannotRaiseTheWorkersLimits(t *testing.T) {
	s := &serve.Server{}
	big := video.DefaultLimits
	big.MaxBytes = 1 << 40
	big.MaxBoxes = 1 << 40
	res, _ := s.Handle(videoproto.Request{ID: 1, Op: videoproto.OpProbe, Key: "k", HasData: true, Limits: big}, make([]byte, video.DefaultLimits.MaxBytes+1))
	if res.Code != video.CodeLimit {
		t.Fatalf("a request that asks for a larger file than the worker reads: %+v", res)
	}
	res, _ = s.Handle(videoproto.Request{ID: 2, Op: "unlink", Key: "k"}, nil)
	if res.Code != video.CodeSyntax {
		t.Fatalf("an unknown operation: %+v", res)
	}
}
