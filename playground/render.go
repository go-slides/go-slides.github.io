package playground

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-slides/slides"
	"github.com/go-slides/slides/deck"
	"github.com/go-slides/slides/editor"
	"github.com/go-slides/slides/project"
	"github.com/go-slides/slides/vfs"
	"github.com/go-slides/slides/viewer"
)

// MaxSource bounds the deck the playground renders: a slide deck typed in a
// browser is a few kilobytes, a few images dropped into it (MaxImage each,
// as data: URIs) some hundreds, and the bound keeps a pasted ten megabytes
// from holding the worker.
const MaxSource = 1 << 20

// Request asks the worker for a preview.
type Request struct {
	ID     int    `json:"id"`
	Source string `json:"source"`
	Mode   Mode   `json:"mode"`
	// Presenter asks for the presenter view's page too: the presenter
	// window is open.
	Presenter bool `json:"presenter,omitempty"`
}

// Response is the worker's answer.
type Response struct {
	ID int `json:"id"`
	// Doc is the slideshow's page. An overview comes as Head and Parts
	// instead (editor.Result's), which the page assembles into a document
	// (overviewPage) or a patch of the slides that changed. Parts is nil
	// for no overview, and empty for an overview with no slide.
	Doc   string   `json:"doc,omitempty"`
	Head  string   `json:"head,omitempty"`
	Parts []string `json:"parts"`
	// PresenterDoc is the presenter view's page, when the request asked.
	PresenterDoc string              `json:"presenterDoc,omitempty"`
	Slides       int                 `json:"slides"`
	Diagnostics  []editor.Diagnostic `json:"diagnostics,omitempty"`
	Millis       int64               `json:"millis"`
	Error        string              `json:"error,omitempty"`
}

// Page is the preview's document: the slideshow's, or the overview
// assembled from its parts ("" for neither: a render that failed).
func (r Response) Page() string { return r.PageAt(SlideTop{I: -1}) }

// PageAt is Page, an overview starting scrolled to top.
func (r Response) PageAt(top SlideTop) string {
	if r.Parts == nil {
		return r.Doc
	}
	return overviewPage(r.Head+strings.Join(r.Parts, "")+"</div>", top)
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
		// Both pages sync through the playground page, which relays
		// between the preview and the presenter window: a sandboxed page
		// has no localStorage (see viewer.SyncParent).
		page, err := presentation(req.Source, viewer.Bespoke{OSC: true, Transition: true, Presenter: true, Sync: viewer.SyncParent})
		if err != nil {
			res.Error = err.Error()
			break
		}
		res.Doc = sandboxPage(page)
		if req.Presenter {
			pv, err := presentation(req.Source, viewer.Bespoke{Presenter: true, View: "presenter", Sync: viewer.SyncParent})
			if err != nil {
				res.Error = err.Error()
				break
			}
			res.PresenterDoc = sandboxPage(pv)
		}
	default:
		res.Head, res.Parts = r.Head, r.SlideHTML
		if res.Parts == nil {
			res.Parts = []string{}
		}
	}
	res.Millis = time.Since(start).Milliseconds()
	return res
}

// Presentation builds the deck as the slides command would: one HTML page,
// with the keyboard, transitions and the presenter view. Raw HTML follows
// Marp's allowlist.
func Presentation(src string) (string, error) {
	return presentation(src, viewer.DefaultBespoke())
}

// presentation is Presentation with the presenting page's options.
func presentation(src string, b viewer.Bespoke) (string, error) {
	if len(src) > MaxSource {
		return "", errors.New("the deck is too large for the playground")
	}
	fs := vfs.NewMap(map[string][]byte{DeckPath: []byte(src)})
	cfg, err := project.Load(DeckPath, project.Options{FS: fs, NoConfig: true})
	if err != nil {
		return "", err
	}
	cfg.HTML = deck.HTMLAllowlist
	cfg.Bespoke = &b
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
//
// It also applies a patch (a PreviewPatch posted by the page as
// {goSlidesPatch: …}): each changed slide's markup is parsed into an inert
// <template> and put in place of the slide at its index, and slides past
// the count are removed. Parsing into a template runs nothing (scripts in
// one are inert, and the page's CSP allows only its own nonce), so a patch
// is markup exactly as the document it replaces would be, never code. The
// fitted headings are measured again when a changed slide has one.
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
window.addEventListener("message",function(e){
if(e.source!==parentWin||!e.data||!e.data.goSlidesPatch)return;
var p=e.data.goSlidesPatch,box=document.getElementById("go-slides-preview");
if(!box||typeof p.count!=="number"||!Array.isArray(p.changed))return;
var fit=false;
for(var i=0;i<p.changed.length;i++){var c=p.changed[i];if(typeof c.i!=="number"||typeof c.html!=="string")continue;
var t=document.createElement("template");t.innerHTML=c.html;var n=t.content.firstElementChild;if(!n)continue;
if(c.html.indexOf("data-fit-content")>=0)fit=true;
var old=box.children[c.i];if(old)old.replaceWith(n);else box.appendChild(n)}
while(box.children.length>p.count)box.lastElementChild.remove();
box.setAttribute("data-slides",String(p.count));
if(fit)fitHeadings();
parentWin.postMessage({goSlidesPatched:p.count},"*");
});
var topFrame=0;
function reportTop(){topFrame=0;var box=document.getElementById("go-slides-preview");if(!box)return;
for(var i=0;i<box.children.length;i++){var r=box.children[i].getBoundingClientRect();
if(r.bottom>0){parentWin.postMessage({goSlidesTop:{i:i,f:r.height>0?Math.max(0,-r.top)/r.height:0}},"*");return}}}
window.addEventListener("scroll",function(){if(!topFrame)topFrame=requestAnimationFrame(reportTop)},{passive:true});
})();
function startAt(i,f){var box=document.getElementById("go-slides-preview"),s=box&&box.children[i];
if(s){var r=s.getBoundingClientRect();window.scrollTo(0,window.scrollY+r.top+f*r.height)}}`

// SlideTop is where the overview is scrolled: the top-most visible slide,
// and how far down it the view starts (0 to 1). The frame reports it as it
// scrolls; a new document starts there, so a replaced document keeps the
// slide in view whatever its new height. I < 0 is none.
type SlideTop struct {
	I int
	F float64
}

// overviewPage wraps the editor package's preview markup in a page of its
// own, with the fitting-heading script and the scroll sync, scrolled to top.
func overviewPage(markup string, top SlideTop) string {
	nonce := newNonce()
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8">`)
	b.WriteString(`<meta http-equiv="Content-Security-Policy" content="` + csp("'nonce-"+nonce+"'") + `">`)
	b.WriteString(`<style>html,body{margin:0;background:#e9ebef}body{padding:14px}` +
		`@media (prefers-color-scheme:dark){html,body{background:#0d1117}}</style>`)
	b.WriteString(`</head><body>`)
	b.WriteString(markup)
	// fitHeadings is deck.FitScript, kept to run again after a patch.
	b.WriteString(`<script nonce="` + nonce + `">function fitHeadings(){` + deck.FitScript + "}\n" + overviewScript + "\nfitHeadings();")
	if top.I >= 0 {
		// numbers only: the place cannot carry markup or code
		f := min(max(top.F, 0), 1)
		b.WriteString("startAt(" + strconv.Itoa(top.I) + "," + strconv.FormatFloat(f, 'f', 4, 64) + ");")
	}
	b.WriteString("</script>")
	b.WriteString(`</body></html>`)
	return b.String()
}

var scriptRE = regexp.MustCompile(`(?s)<script>(.*?)</script>`)

// sandboxShim runs before the presentation's own script. In a sandboxed
// iframe (an opaque origin, about:srcdoc) the history API throws on a URL
// change and localStorage throws on access; the presentation uses the first
// for the slide number in the URL and the second to talk to its presenter
// view. The shim makes the first a no-op and gives the second an in-memory
// store, so the slideshow works in the preview; the presenter view syncs
// through the parent page instead (viewer.SyncParent).
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
