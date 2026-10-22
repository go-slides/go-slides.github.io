package playground

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strings"
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

// mediaTypes are the media a project holds, by extension.
var mediaTypes = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif",
	".webp": "image/webp", ".avif": "image/avif", ".svg": "image/svg+xml", ".ico": "image/x-icon",
	".mp4": "video/mp4", ".webm": "video/webm", ".mov": "video/quicktime", ".m4v": "video/mp4",
	".ogv": "video/ogg", ".mp3": "audio/mpeg", ".m4a": "audio/mp4", ".ogg": "audio/ogg",
	".wav": "audio/wav", ".woff": "font/woff", ".woff2": "font/woff2", ".ttf": "font/ttf", ".otf": "font/otf",
}

// IsMedia reports whether a project file is a medium (bytes), not text the
// editor opens and the renderer reads.
func IsMedia(p string) bool {
	_, ok := mediaTypes[strings.ToLower(path.Ext(p))]
	return ok
}

// MediaType is the type a medium is posted and inlined with: by its
// extension, except that bytes that say they are an image win (a .png that
// is a JPEG). An SVG keeps its type: shown as an image, it runs nothing.
func MediaType(p string, data []byte) string {
	t := mediaTypes[strings.ToLower(path.Ext(p))]
	if strings.HasPrefix(t, "image/") && t != "image/svg+xml" {
		if s := http.DetectContentType(data); strings.HasPrefix(s, "image/") {
			return s
		}
	}
	if t == "" {
		return "application/octet-stream"
	}
	return t
}

// MediaMarkdown is what a medium dropped into the editor adds to the deck:
// an image, or the HTML of a video or a sound (go-slides' HTML allowlist
// takes them).
func MediaMarkdown(p string) string {
	t := mediaTypes[strings.ToLower(path.Ext(p))]
	ref := escapeRef(p)
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

// escapeRef writes a path as a URL path: the characters a Markdown link or
// an attribute would end at are escaped.
func escapeRef(p string) string {
	var b strings.Builder
	for _, r := range p {
		if strings.ContainsRune(" ()<>\"'#?%", r) {
			fmt.Fprintf(&b, "%%%02X", r)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// InlineMedia puts the media a page names by relative path into the page as
// data: URIs, for a page that leaves the playground as one file (Download
// HTML). Only references the renderer writes are replaced: an attribute's
// value and a CSS url(), quoted or not.
func InlineMedia(page string, media map[string][]byte) string {
	if len(media) == 0 {
		return page
	}
	names := make([]string, 0, len(media))
	for k := range media {
		names = append(names, k)
	}
	// longest first: media/a.png.bak is not media/a.png
	sort.Slice(names, func(i, j int) bool { return len(names[i]) > len(names[j]) })
	var pairs []string
	for _, n := range names {
		uri := "data:" + MediaType(n, media[n]) + ";base64," + base64.StdEncoding.EncodeToString(media[n])
		refs := []string{n, "./" + n}
		if e := escapeRef(n); e != n {
			refs = append(refs, e, "./"+e)
		}
		for _, ref := range refs {
			for _, q := range []string{`"`, `'`, `&quot;`, `&#34;`, `(`} {
				end := q
				if q == "(" {
					end = ")"
				}
				pairs = append(pairs, q+ref+end, q+uri+end)
			}
		}
	}
	return strings.NewReplacer(pairs...).Replace(page)
}

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
