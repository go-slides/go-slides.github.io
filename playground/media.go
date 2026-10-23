package playground

import (
	"path"
	"strings"

	"github.com/go-slides/go-slides.github.io/playground/mediafiles"
)

// Media are a project's images, videos, sounds and fonts: bytes the deck
// names by a relative path (![bg](media/photo.jpg)), kept in the project
// store, never in the deck's text. go-slides leaves such a path as it is in
// the markup, and the preview's sandboxed frame, whose origin is opaque,
// could not load it from anywhere: the page posts the bytes into the frame,
// whose own script (mediaScript) makes blob: URLs of them, in its own
// origin, and puts them in place of the paths. A blob: URL made by the page
// is refused there (measured in Chrome: "Not allowed to load local
// resource"), and data: URIs would put 4/3 of every byte in the markup the
// worker sends on each keystroke.

// IsMedia reports whether a project file is a medium (bytes), not text the
// editor opens and the renderer reads (mediafiles.IsMedia).
func IsMedia(p string) bool { return mediafiles.IsMedia(p) }

// MediaType is the type a medium is posted and inlined with
// (mediafiles.Type).
func MediaType(p string, data []byte) string { return mediafiles.Type(p, data) }

// MediaMarkdown is what a medium dropped into the editor adds to the deck:
// an image, or the HTML of a video or a sound (go-slides' HTML allowlist
// takes them).
func MediaMarkdown(p string) string {
	t := mediafiles.Kind(p)
	ref := mediafiles.EscapeRef(p)
	switch {
	case strings.HasPrefix(t, "video/"):
		return `<video src="` + ref + `" controls></video>`
	case strings.HasPrefix(t, "audio/"):
		return `<audio src="` + ref + `" controls></audio>`
	}
	alt := strings.TrimSuffix(path.Base(p), path.Ext(p))
	alt = strings.Map(func(r rune) rune {
		if strings.ContainsRune("[]()\\\n\r<>", r) {
			return -1
		}
		return r
	}, alt)
	return "![" + alt + "](" + ref + ")"
}

// InlineMedia puts the media a page names by relative path into the page as
// data: URIs, for a page that leaves the playground as one file (Save as ▸
// HTML page): mediafiles.Inline.
func InlineMedia(page string, media map[string][]byte) string { return mediafiles.Inline(page, media) }

// mediaScript runs in the preview's frame (and the slideshow's, and the
// presenter view's): it takes {goSlidesMedia: {path: {type, bytes}}} from
// its parent, makes a blob: URL of each in its own origin, and puts them in
// place of the relative paths that name them: in src, poster and href
// (an SVG <image>; never a link's), in style attributes and in style
// elements' url(). Elements added later (a patched slide, the slideshow's
// clones) are fixed as they arrive. A path that names no medium is left as
// it is; ../ is resolved, and a theme's ../media/x (written from themes/)
// finds media/x. The URLs of the previous set are revoked once replaced.
const mediaScript = `(function(){var P=window.parent,urls={},paths={};
function key(u){if(typeof u!=="string")return null;u=u.trim();if(u.slice(0,5)==="blob:")return paths[u]||null;
if(/^[a-z][a-z0-9+.-]*:/i.test(u)||u.charAt(0)==="/"||u.charAt(0)==="#")return null;
u=u.split("#")[0].split("?")[0];try{u=decodeURIComponent(u)}catch(e){}
var out=[];u.split("/").forEach(function(s){if(s===""||s===".")return;if(s===".."){if(out.length&&out[out.length-1]!=="..")out.pop();else out.push("..")}else out.push(s)});
var p=out.join("/");while(!Object.prototype.hasOwnProperty.call(urls,p)&&p.slice(0,3)==="../")p=p.slice(3);
return Object.prototype.hasOwnProperty.call(urls,p)?p:null}
function sub(css){return css.replace(/url\(\s*(['"]?)([^'")]*)\1\s*\)/g,function(m,q,u){var k=key(u);return k?'url("'+urls[k]+'")':m})}
function one(el){if(!el||el.nodeType!==1)return;var tag=el.tagName.toUpperCase();
["src","poster","href"].forEach(function(a){if(a==="href"&&tag!=="IMAGE")return;var v=el.getAttribute(a),k=v&&key(v);if(k)el.setAttribute(a,urls[k])});
var s=el.getAttribute("style");if(s&&s.indexOf("url(")>=0){var n=sub(s);if(n!==s)el.setAttribute("style",n)}
if(tag==="STYLE"&&el.textContent.indexOf("url(")>=0){var t=sub(el.textContent);if(t!==el.textContent)el.textContent=t}}
function fix(root){one(root);if(root&&root.querySelectorAll){var els=root.querySelectorAll("[src],[poster],image,[style],style");for(var i=0;i<els.length;i++)one(els[i])}}
new MutationObserver(function(ms){if(!Object.keys(urls).length)return;ms.forEach(function(m){for(var i=0;i<m.addedNodes.length;i++)fix(m.addedNodes[i])})}).observe(document.documentElement,{childList:true,subtree:true});
window.addEventListener("message",function(e){if(e.source!==P||!e.data||!e.data.goSlidesMedia||typeof e.data.goSlidesMedia!=="object")return;
var m=e.data.goSlidesMedia,old=paths;urls={};paths={};
for(var k in m){var f=m[k];if(!f||!(f.bytes instanceof ArrayBuffer)||typeof f.type!=="string"||!/^(image|video|audio|font)\/[a-z0-9.+-]+$/.test(f.type))continue;
var u=URL.createObjectURL(new Blob([f.bytes],{type:f.type}));urls[k]=u;paths[u]=k}
for(var o in old)paths[o]=old[o];fix(document.documentElement);
for(var o2 in old){delete paths[o2];URL.revokeObjectURL(o2)}
P.postMessage({goSlidesMediaShown:Object.keys(urls).length},"*")});
})();`
