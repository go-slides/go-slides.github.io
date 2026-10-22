package playground

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/go-slides/slides/deck"
)

// The slide rail is the column of thumbnails beside the preview. The slides
// themselves are the overview's own markup: an SVG per slide, which the browser
// scales to any width, so a thumbnail is the real slide (its fonts, math and
// images included) at 120 pixels, with nothing to rasterise and nothing to
// download. The canvas draws the strip around them — numbers, the selection,
// the scrollbar, the keyboard — through the toolkit's PageStrip, and the host
// lays a transparent frame of those SVGs over its cards (see cmd/playground-wasm).
//
// The frame holds the same head and slides as the preview and takes the same
// per-slide patches, so editing one slide re-parses one slide in each. The
// browser skips the layout and paint of the slides outside the strip
// (content-visibility), so a 500-slide deck costs the thumbnails in view.

// slideLineRE finds the source line a slide starts at (the preview's own
// data-line attribute).
var slideLineRE = regexp.MustCompile(`data-line="(\d+)"`)

// SlideLines is the 0-based source line each slide starts at, read from the
// overview's slides (-1 for a slide that names none).
func SlideLines(parts []string) []int {
	lines := make([]int, len(parts))
	for i, p := range parts {
		lines[i] = -1
		if m := slideLineRE.FindStringSubmatch(p); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil {
				lines[i] = n
			}
		}
	}
	return lines
}

// SlideAt is the slide a source line belongs to: the last slide starting at or
// before it (0 for a line before the first; -1 when there are no slides).
func SlideAt(lines []int, line int) int {
	at := -1
	for i, l := range lines {
		if l < 0 {
			continue
		}
		if l > line {
			break
		}
		at = i
	}
	if at < 0 && len(lines) > 0 {
		return 0
	}
	return at
}

// railScript places the slides under the strip's cards and applies the same
// per-slide patches as the preview's (overviewScript). Messages come only from
// the parent page. A {goSlidesRail: {x, y, w, h, gap}} message gives the
// first card's position and size and the gap between cards, in CSS pixels; the
// slides are laid out from them, so scrolling the strip is one style change.
const railScript = `(function(){
var parentWin=window.parent,root=document.documentElement;
function num(v){return typeof v==="number"&&isFinite(v)&&Math.abs(v)<1e6}
window.addEventListener("message",function(e){
if(e.source!==parentWin||!e.data)return;
var r=e.data.goSlidesRail;
if(r&&num(r.x)&&num(r.y)&&num(r.w)&&num(r.h)&&num(r.gap)){
var s=root.style;s.setProperty("--rx",r.x+"px");s.setProperty("--ry",r.y+"px");
s.setProperty("--rw",r.w+"px");s.setProperty("--rh",r.h+"px");s.setProperty("--rgap",r.gap+"px")}
var p=e.data.goSlidesPatch,box=document.getElementById("go-slides-preview");
if(!p||!box||typeof p.count!=="number"||!Array.isArray(p.changed))return;
var fit=false;
for(var i=0;i<p.changed.length;i++){var c=p.changed[i];if(typeof c.i!=="number"||typeof c.html!=="string")continue;
var t=document.createElement("template");t.innerHTML=c.html;var n=t.content.firstElementChild;if(!n)continue;
if(c.html.indexOf("data-fit-content")>=0)fit=true;
var old=box.children[c.i];if(old)old.replaceWith(n);else box.appendChild(n)}
while(box.children.length>p.count)box.lastElementChild.remove();
box.setAttribute("data-slides",String(p.count));
if(fit)fitHeadings();
});
})();`

// railCSS overrides the overview's layout: the slides stack from the first
// card's position at the card's size, on a transparent page, and the browser
// skips the ones that are out of the frame.
const railCSS = `html,body{margin:0!important;padding:0!important;background:transparent!important;overflow:hidden}` +
	`#go-slides-preview{position:absolute;left:var(--rx,0);top:var(--ry,0);margin:0!important}` +
	`#go-slides-preview>svg[data-marpit-svg]{width:var(--rw,120px)!important;height:var(--rh,68px)!important;` +
	`margin:0 0 var(--rgap,10px)!important;box-shadow:none!important;content-visibility:auto;` +
	`contain-intrinsic-size:var(--rw,120px) var(--rh,68px)}`

// railPage is the thumbnail frame's document for an overview's head and slides.
func railPage(head string, parts []string) string {
	nonce := newNonce()
	var b strings.Builder
	b.WriteString(`<!doctype html><html><head><meta charset="utf-8">`)
	b.WriteString(`<meta http-equiv="Content-Security-Policy" content="` + csp("'nonce-"+nonce+"'") + `">`)
	b.WriteString(`</head><body>`)
	b.WriteString(head + strings.Join(parts, "") + "</div>")
	// the override comes after the deck's own style sheet, which the head holds
	b.WriteString(`<style>` + railCSS + `</style>`)
	b.WriteString(`<script nonce="` + nonce + `">function fitHeadings(){` + deck.FitScript + "}\n" + railScript + "\nfitHeadings();</script>")
	b.WriteString(`</body></html>`)
	return b.String()
}
