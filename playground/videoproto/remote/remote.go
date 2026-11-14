// Package remote is the client of a video worker: a video.Provider that asks
// one over a Transport (videoproto's requests) instead of reading the file
// itself. go-slides, built with the tag slides_novideo, links no container
// reader or decoder; the host registers this and the deck's videos are read in
// the worker.
//
// The renders of a page come again and again for the same videos (every
// keystroke builds the deck anew), so the answers are kept: a probe by the
// file's fingerprint, a poster by fingerprint, time and width. The file crosses
// to the worker once, and again only if the worker no longer holds it.
package remote

import (
	"container/list"
	"strconv"
	"sync"

	"github.com/go-slides/go-slides.github.io/playground/videoproto"
	"github.com/go-slides/slides/video"
)

// Transport carries one request to a worker and brings back its answer and
// the answer's payload. It blocks until the worker answers; its error is that
// of the carrying (the worker could not be started, did not answer in time),
// not a refusal of the file, which is the Response's.
type Transport interface {
	Call(req videoproto.Request, payload []byte) (videoproto.Response, []byte, error)
}

// maxPosterBytes bounds the posters kept: a poster is a few hundred KiB.
const maxPosterBytes = 24 << 20

// Provider is a video.Provider that asks a worker.
type Provider struct {
	T Transport

	mu     sync.Mutex
	next   int
	probes map[string]video.Probed
	sent   map[string]bool // the keys the worker holds, as far as this client knows
	order  []string        // the keys sent, oldest first (the worker keeps two)
	posts  map[string]*list.Element
	lru    *list.List
	bytes  int
}

type poster struct {
	key  string
	jpg  []byte
	w, h int
}

// New is a Provider over a Transport.
func New(t Transport) *Provider { return &Provider{T: t} }

var _ video.Provider = (*Provider)(nil)

// workerKeep is how many files the worker holds (serve.keep).
const workerKeep = 2

// key is carried in Probed.State, so that a poster does not hash the file
// again.
type state struct{ key string }

// Probe implements video.Provider.
func (p *Provider) Probe(data []byte, format string, lim video.Limits) (video.Probed, error) {
	key := videoproto.Fingerprint(data)
	p.mu.Lock()
	if pr, ok := p.probes[key]; ok {
		p.mu.Unlock()
		return pr, nil
	}
	p.mu.Unlock()
	res, _, err := p.call(videoproto.Request{Op: videoproto.OpProbe, Key: key, Format: format, Limits: lim}, data)
	if err != nil {
		return video.Probed{}, err
	}
	pr := video.Probed{Codec: res.Codec, Width: res.Width, Height: res.Height, Duration: res.Duration, Decodable: res.Decodable, State: state{key}}
	p.mu.Lock()
	if p.probes == nil || len(p.probes) >= 64 {
		p.probes = map[string]video.Probed{}
	}
	p.probes[key] = pr
	p.mu.Unlock()
	return pr, nil
}

// Poster implements video.Provider.
func (p *Provider) Poster(pr video.Probed, data []byte, t float64, lim video.Limits) ([]byte, int, int, error) {
	st, ok := pr.State.(state)
	if !ok {
		st = state{videoproto.Fingerprint(data)}
	}
	pk := st.key + "@" + ftoa(t) + "w" + itoa(lim.PosterWidth)
	p.mu.Lock()
	if e, ok := p.posts[pk]; ok {
		p.lru.MoveToFront(e)
		c := e.Value.(*poster)
		p.mu.Unlock()
		return c.jpg, c.w, c.h, nil
	}
	p.mu.Unlock()
	res, img, err := p.call(videoproto.Request{Op: videoproto.OpPoster, Key: st.key, T: t, Limits: lim}, data)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(img) == 0 {
		return nil, 0, 0, video.Fail(video.CodeSyntax, "the video worker answered a poster with no picture")
	}
	p.mu.Lock()
	if p.posts == nil {
		p.posts, p.lru = map[string]*list.Element{}, list.New()
	}
	p.posts[pk] = p.lru.PushFront(&poster{pk, img, res.Width, res.Height})
	p.bytes += len(img)
	for p.bytes > maxPosterBytes && p.lru.Len() > 1 {
		old := p.lru.Remove(p.lru.Back()).(*poster)
		delete(p.posts, old.key)
		p.bytes -= len(old.jpg)
	}
	p.mu.Unlock()
	return img, res.Width, res.Height, nil
}

// call sends a request, with the file when the worker is not known to hold
// it, and sends it again with the file when the worker says it does not. A
// refusal in the answer is returned as a *video.Error.
func (p *Provider) call(req videoproto.Request, data []byte) (videoproto.Response, []byte, error) {
	for attempt := 0; ; attempt++ {
		p.mu.Lock()
		p.next++
		req.ID = p.next
		have := p.sent[req.Key]
		p.mu.Unlock()
		var payload []byte
		req.HasData = !have || attempt > 0
		if req.HasData {
			payload = data
		}
		res, img, err := p.T.Call(req, payload)
		if err != nil {
			return res, nil, video.Fail(video.CodeUnavailable, "the video worker: %v", err)
		}
		if res.Code == videoproto.CodeNotHeld && !req.HasData {
			p.forget(req.Key)
			if attempt > 0 {
				return res, nil, video.Fail(video.CodeUnavailable, "the video worker lost the file twice")
			}
			continue
		}
		if res.Code != "" {
			// a refusal may come before the worker kept the file: it is sent
			// again next time, which costs nothing a refused file is worth
			return res, nil, video.Fail(res.Code, "%s", res.Error)
		}
		if req.HasData {
			p.remember(req.Key)
		}
		return res, img, nil
	}
}

func (p *Provider) remember(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.sent == nil {
		p.sent = map[string]bool{}
	}
	if p.sent[key] {
		return
	}
	p.sent[key] = true
	p.order = append(p.order, key)
	for len(p.order) > workerKeep {
		delete(p.sent, p.order[0])
		p.order = p.order[1:]
	}
}

func (p *Provider) forget(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.sent, key)
	for i, k := range p.order {
		if k == key {
			p.order = append(p.order[:i], p.order[i+1:]...)
			break
		}
	}
}

// Reset forgets what the worker holds and the answers kept: the worker was
// replaced.
func (p *Provider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sent, p.order = nil, nil
}

func itoa(n int) string     { return strconv.Itoa(n) }
func ftoa(f float64) string { return strconv.FormatFloat(f, 'g', -1, 64) }
