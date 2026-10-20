package playground

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-slides/slides"
	"github.com/go-slides/slides/deck"
	"github.com/go-slides/slides/editor"
	"github.com/go-slides/slides/project"
	"github.com/go-slides/slides/vfs"
)

// MaxSource bounds the deck the playground renders: a slide deck typed or
// pasted in a browser is a few kilobytes, and the bound keeps a pasted
// megabyte from holding the worker.
const MaxSource = 256 << 10

// Request asks the worker for a preview.
type Request struct {
	ID     int    `json:"id"`
	Source string `json:"source"`
	Mode   Mode   `json:"mode"`
}

// Response is the worker's answer.
type Response struct {
	ID          int                 `json:"id"`
	Doc         string              `json:"doc,omitempty"`
	Slides      int                 `json:"slides"`
	Diagnostics []editor.Diagnostic `json:"diagnostics,omitempty"`
	Millis      int64               `json:"millis"`
	Error       string              `json:"error,omitempty"`
}

// Serve renders a request; the worker runs it. The document it returns is
// the preview iframe's: a complete page with its own Content Security
// Policy, shown sandboxed (no same origin) by the host page.
func Serve(req Request) Response {
	start := time.Now()
	res := Response{ID: req.ID}
	if len(req.Source) > MaxSource {
		res.Error = fmt.Sprintf("the deck is %d KiB; the playground renders decks up to %d KiB — use the slides command for larger ones", len(req.Source)>>10, MaxSource>>10)
		return res
	}
	r := editor.Render(editor.Request{Path: DeckPath, Source: req.Source})
	res.Diagnostics, res.Slides = r.Diagnostics, r.Slides
	switch {
	case r.HTML == "":
		// An error: the previous preview stays, the problems say why.
	case req.Mode == ModeSlideshow:
		page, err := Presentation(req.Source)
		if err != nil {
			res.Error = err.Error()
			break
		}
		res.Doc = sandboxPage(page)
	default:
		res.Doc = overviewPage(r.HTML)
	}
	res.Millis = time.Since(start).Milliseconds()
	return res
}

// Presentation builds the deck as the slides command would: one HTML page,
// with the keyboard, transitions and the presenter view. Raw HTML follows
// Marp's allowlist.
func Presentation(src string) (string, error) {
	if len(src) > MaxSource {
		return "", errors.New("the deck is too large for the playground")
	}
	fs := vfs.NewMap(map[string][]byte{DeckPath: []byte(src)})
	cfg, err := project.Load(DeckPath, project.Options{FS: fs, NoConfig: true})
	if err != nil {
		return "", err
	}
	cfg.HTML = deck.HTMLAllowlist
	page, _, err := slides.HTML([]byte(src), cfg)
	return page, err
}

// csp is the preview's Content Security Policy: no network but images,
// media, style sheets and fonts (a deck links them); scripts only those the
// page itself carries, named by hash or nonce; no form, no frame, no base.
func csp(scripts string) string {
	return "default-src 'none'; img-src https: http: data: blob:; media-src https: data:; " +
		"style-src 'unsafe-inline' https:; font-src https: data:; script-src " + scripts +
		"; base-uri 'none'; form-action 'none'; frame-src 'none'"
}

// overviewScript scrolls the overview to the editor's caret, and tells the
// editor which line a click in the preview landed on. Messages come only
// from the parent page, and go only to it.
const overviewScript = `(function(){
var parentWin=window.parent;
function lineOf(el){while(el&&el!==document.body){if(el.classList&&el.classList.contains("code-line"))return +el.getAttribute("data-line");el=el.parentElement}return -1}
window.addEventListener("message",function(e){
if(e.source!==parentWin||!e.data||typeof e.data.line!=="number")return;
var best=null,els=document.querySelectorAll("svg.code-line,.code-line");
for(var i=0;i<els.length;i++){var l=+els[i].getAttribute("data-line");if(l<=e.data.line&&(!best||l>=+best.getAttribute("data-line")))best=els[i]}
if(best)best.scrollIntoView({block:"nearest",behavior:"smooth"});
});
document.addEventListener("click",function(e){var l=lineOf(e.target);if(l>=0)parentWin.postMessage({goSlidesLine:l},"*")});
})();`

// overviewPage wraps the editor package's preview markup in a page of its
// own, with the fitting-heading script and the scroll sync.
func overviewPage(markup string) string {
	nonce := newNonce()
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8">`)
	b.WriteString(`<meta http-equiv="Content-Security-Policy" content="` + csp("'nonce-"+nonce+"'") + `">`)
	b.WriteString(`<style>html,body{margin:0;background:#e9ebef}body{padding:14px}` +
		`@media (prefers-color-scheme:dark){html,body{background:#0d1117}}</style>`)
	b.WriteString(`</head><body>`)
	b.WriteString(markup)
	b.WriteString(`<script nonce="` + nonce + `">` + overviewScript + "\n" + deck.FitScript + `</script>`)
	b.WriteString(`</body></html>`)
	return b.String()
}

var scriptRE = regexp.MustCompile(`(?s)<script>(.*?)</script>`)

// sandboxShim runs before the presentation's own script. In a sandboxed
// iframe (an opaque origin, about:srcdoc) the history API throws on a URL
// change and localStorage throws on access; the presentation uses the first
// for the slide number in the URL and the second to talk to its presenter
// view. The shim makes the first a no-op and gives the second an in-memory
// store, so the slideshow works in the preview (the presenter view, which
// needs a second window, works in the downloaded page).
const sandboxShim = `(function(){try{var h=history;["replaceState","pushState"].forEach(function(n){var f=h[n].bind(h);h[n]=function(){try{return f.apply(null,arguments)}catch(e){}}})}catch(e){}
try{window.localStorage}catch(e){var m={};try{Object.defineProperty(window,"localStorage",{configurable:true,value:{getItem:function(k){return k in m?m[k]:null},setItem:function(k,v){m[k]=String(v)},removeItem:function(k){delete m[k]},clear:function(){m={}},key:function(i){return Object.keys(m)[i]||null},get length(){return Object.keys(m).length}}})}catch(e){}}})();`

// sandboxPage puts a Content Security Policy in a presentation page that
// allows its own scripts, by hash, and nothing else, and runs sandboxShim
// first.
func sandboxPage(page string) string {
	if i := strings.Index(page, "<head>"); i >= 0 {
		i += len("<head>")
		page = page[:i] + "<script>" + sandboxShim + "</script>" + page[i:]
	}
	var hashes []string
	for _, m := range scriptRE.FindAllStringSubmatch(page, -1) {
		sum := sha256.Sum256([]byte(m[1]))
		hashes = append(hashes, "'sha256-"+base64.StdEncoding.EncodeToString(sum[:])+"'")
	}
	scripts := "'none'"
	if len(hashes) > 0 {
		scripts = strings.Join(hashes, " ")
	}
	meta := `<meta http-equiv="Content-Security-Policy" content="` + csp(scripts) + `">`
	if i := strings.Index(page, "<head>"); i >= 0 {
		return page[:i+len("<head>")] + meta + page[i+len("<head>"):]
	}
	return meta + page
}

func newNonce() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return base64.RawURLEncoding.EncodeToString(b[:])
}
