// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

// Command playground-wasm is the browser shell of the go-slides playground.
// The same module runs twice:
//
//   - in the page, it is the editor: a go-widgets canvas application (package
//     playground) handed to go-widgets/webcanvas, plus the sandboxed iframe
//     that shows the preview over the canvas;
//   - in a Web Worker (worker.js), it renders: each request is a deck, each
//     answer a preview page, so typing never waits for go-slides.
//
// The page compiles the module once and hands the compiled module to the
// worker, so it is downloaded and compiled one time. The shell holds no
// logic: everything is in the tagless playground package.
package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	playground "github.com/go-slides/go-slides.github.io/playground"
	"github.com/go-widgets/toolkit"
	"github.com/go-widgets/webcanvas"
)

const canvasID = "slides-canvas"

// renderTimeout is how long a render may take before the worker is stopped
// and replaced: go-slides cannot be interrupted, the worker can.
const renderTimeout = 5 * time.Second

func main() {
	if js.Global().Get("document").IsUndefined() {
		worker()
		return
	}
	page()
}

// worker answers render requests, one JSON message in, one out.
func worker() {
	js.Global().Set("onmessage", js.FuncOf(func(_ js.Value, args []js.Value) any {
		var req playground.Request
		if len(args) == 0 || args[0].Get("data").Type() != js.TypeString ||
			json.Unmarshal([]byte(args[0].Get("data").String()), &req) != nil {
			return nil
		}
		b, _ := json.Marshal(playground.Serve(req))
		js.Global().Call("postMessage", string(b))
		return nil
	}))
	js.Global().Call("postMessage", `{"ready":true}`)
	select {}
}

// transport talks to the render worker, replacing it when it hangs.
type transport struct {
	vm      *playground.ViewModel
	w       js.Value
	ready   bool
	queued  string
	timer   js.Value
	sent    time.Time
	lastRTT time.Duration
	onMsg   js.Func
	// onTimeout is created once: setTimeout is armed for every request.
	onTimeout js.Func
}

func (t *transport) spawn() {
	t.ready = false
	t.w = js.Global().Get("Worker").New(js.Global().Get("goSlidesWorkerURL"))
	t.w.Call("addEventListener", "message", t.onMsg)
	// The page compiled the module; the worker instantiates the same one.
	t.w.Call("postMessage", map[string]any{"module": js.Global().Get("goSlidesModule")})
}

func (t *transport) Send(r playground.Request) {
	b, _ := json.Marshal(r)
	t.queued = string(b)
	t.flush()
}

func (t *transport) flush() {
	if !t.ready || t.queued == "" {
		return
	}
	t.w.Call("postMessage", t.queued)
	t.queued = ""
	t.sent = time.Now()
	t.timer = js.Global().Call("setTimeout", t.onTimeout, renderTimeout.Milliseconds())
}

// timeout stops a worker whose render took too long and starts another.
func (t *transport) timeout() {
	t.w.Call("terminate")
	t.spawn()
	t.vm.GiveUp("rendering took more than 5 s, so it was stopped: is a template looping?")
}

func (t *transport) message(data js.Value) {
	if data.Type() != js.TypeString {
		return
	}
	s := data.String()
	if s == `{"ready":true}` {
		t.ready = true
		t.flush()
		return
	}
	var r playground.Response
	if json.Unmarshal([]byte(s), &r) != nil {
		return
	}
	js.Global().Call("clearTimeout", t.timer)
	t.lastRTT = time.Since(t.sent)
	t.vm.Receive(r)
}

// host is what the page does for the view-model.
type host struct{ p *presenter }

func (h host) OpenPresenter() { h.p.open() }

func (host) SetFragment(f string) {
	js.Global().Get("history").Call("replaceState", js.Null(), "", "#"+f)
}

func (host) Download(name, mime, data string) {
	blob := js.Global().Get("Blob").New([]any{data}, map[string]any{"type": mime})
	url := js.Global().Get("URL").Call("createObjectURL", blob)
	a := js.Global().Get("document").Call("createElement", "a")
	a.Set("href", url)
	a.Set("download", name)
	a.Call("click")
	js.Global().Get("URL").Call("revokeObjectURL", url)
}

func (host) Open(url string) {
	js.Global().Call("open", url, "_blank", "noopener")
}

func (host) Clipboard(text string) {
	if c := js.Global().Get("navigator").Get("clipboard"); c.Truthy() {
		c.Call("writeText", text)
	}
}

// app is the scene, plus the preview iframe kept over its preview pane.
type app struct {
	*playground.Scene
	frame js.Value
	dpr   float64
	last  toolkit.Rect
}

func (a *app) Draw(buf []byte) {
	a.Scene.Draw(buf)
	r := a.PreviewRect()
	if r == a.last {
		return
	}
	a.last = r
	st := a.frame.Get("style")
	px := func(v int) string { return strconv.FormatFloat(float64(v)/a.dpr, 'f', 2, 64) + "px" }
	st.Set("left", px(r.X))
	st.Set("top", px(r.Y))
	st.Set("width", px(r.W))
	st.Set("height", px(r.H))
}

func page() {
	dpr := 1.0
	if r := js.Global().Get("devicePixelRatio"); r.Type() == js.TypeNumber && r.Float() > 0 {
		dpr = r.Float()
	}
	playground.SetupText(dpr)

	t := &transport{}
	t.onMsg = js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			t.message(args[0].Get("data"))
		}
		return nil
	})
	t.onTimeout = js.FuncOf(func(js.Value, []js.Value) any { t.timeout(); return nil })
	fragment := js.Global().Get("location").Get("hash").String()
	pres := &presenter{}
	hst := host{p: pres}
	vm := playground.NewViewModel(t, hst, fragment)
	pres.vm = vm
	t.vm = vm
	t.spawn()

	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", canvasID)
	w, h := 1200, 800
	if canvas.Truthy() {
		w, h = canvas.Get("clientWidth").Int(), canvas.Get("clientHeight").Int()
	}
	scene := playground.NewScene(vm, hst, max(int(float64(w)*dpr), 1), max(int(float64(h)*dpr), 1))

	// The preview: an iframe with an opaque origin (sandbox without
	// allow-same-origin), so the deck's page can reach neither this page nor
	// its storage, and with its own CSP (see playground.Serve).
	frame := doc.Call("createElement", "iframe")
	frame.Set("id", "slides-preview")
	frame.Call("setAttribute", "sandbox", "allow-scripts")
	frame.Call("setAttribute", "referrerpolicy", "no-referrer")
	frame.Call("setAttribute", "allow", "fullscreen")
	frame.Call("setAttribute", "title", "Slide preview")
	frame.Get("style").Set("cssText", "position:absolute;border:1px solid var(--line);border-radius:8px;background:#fff")
	doc.Call("querySelector", ".stage").Call("appendChild", frame)
	var loaded int
	var docSet time.Time
	var lastLoad time.Duration
	frame.Call("addEventListener", "load", js.FuncOf(func(js.Value, []js.Value) any {
		loaded++
		lastLoad = time.Since(docSet)
		frame.Get("contentWindow").Call("postMessage", map[string]any{"line": vm.CaretLine.Get()}, "*")
		pres.restore(frame)
		return nil
	}))
	pres.audience = frame
	vm.PresenterDoc.Subscribe(pres.setDoc)
	vm.Doc.Subscribe(func(d string) {
		docSet = time.Now()
		frame.Set("srcdoc", d)
	})
	if d := vm.Doc.Get(); d != "" {
		frame.Set("srcdoc", d)
	}
	vm.CaretLine.Subscribe(func(l int) {
		frame.Get("contentWindow").Call("postMessage", map[string]any{"line": l}, "*")
	})
	// A click on a slide in the overview moves the editor to its line. Only
	// the preview iframe is listened to, and only for a number.
	js.Global().Call("addEventListener", "message", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 || !args[0].Get("source").Equal(frame.Get("contentWindow")) {
			return nil
		}
		data := args[0].Get("data")
		if !data.Truthy() || data.Type() != js.TypeObject {
			return nil
		}
		if l := data.Get("goSlidesLine"); l.Type() == js.TypeNumber {
			scene.GoToLine(l.Int())
		}
		pres.fromAudience(data)
		return nil
	}))

	// The clipboard: the canvas leaves ⌘C/⌘V to the browser, whose events
	// carry the text.
	js.Global().Call("addEventListener", "paste", js.FuncOf(func(_ js.Value, args []js.Value) any {
		e := args[0]
		if cd := e.Get("clipboardData"); cd.Truthy() && scene.Paste(cd.Call("getData", "text/plain").String()) {
			e.Call("preventDefault")
		}
		return nil
	}))
	js.Global().Call("addEventListener", "copy", js.FuncOf(func(_ js.Value, args []js.Value) any {
		e := args[0]
		if s := scene.Copy(); s != "" && e.Get("clipboardData").Truthy() {
			e.Get("clipboardData").Call("setData", "text/plain", s)
			e.Call("preventDefault")
		}
		return nil
	}))

	js.Global().Set("goSlidesSetPalette", js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 1 {
			p := playground.LightPalette
			for key, dst := range map[string]*toolkit.RGBA{
				"bg": &p.Background, "surface": &p.Surface, "ink": &p.Ink,
				"line": &p.Line, "border": &p.Border, "accent": &p.Accent,
			} {
				if c, ok := playground.ParseColor(args[0].Get(key).String()); ok {
					*dst = c
				}
			}
			scene.SetPalette(p)
		}
		return nil
	}))

	// goSlidesDebug() is what a test harness asserts on, in CSS pixels.
	// Its positions are the page's, so a harness clicks them as they are.
	origin := func() (float64, float64) {
		r := canvas.Call("getBoundingClientRect")
		return r.Get("left").Float(), r.Get("top").Float()
	}
	js.Global().Set("goSlidesDebug", js.FuncOf(func(js.Value, []js.Value) any {
		ox, oy := origin()
		rects := map[string]any{}
		for name, r := range scene.Rects() {
			rects[name] = []any{ox + float64(r.X)/dpr, oy + float64(r.Y)/dpr, float64(r.W) / dpr, float64(r.H) / dpr}
		}
		line, col := scene.EditorCursor()
		probs := []any{}
		for _, s := range scene.ProblemHighlights() {
			probs = append(probs, []any{s.StartLine, s.StartCol, s.EndCol})
		}
		return map[string]any{
			"rects": rects, "source": vm.Source.Get(), "status": vm.Status.Get(),
			"problems": vm.Problems.Get(), "highlights": probs, "help": vm.Help.Get(),
			"caret": []any{line, col}, "busy": vm.Busy(), "docLength": len(vm.Doc.Get()),
			"docHasCSP": strings.Contains(vm.Doc.Get(), "Content-Security-Policy"),
			"loads":     loaded, "lastRoundTripMs": t.lastRTT.Milliseconds(), "lastLoadMs": lastLoad.Milliseconds(),
			"mode": vm.Mode.Get(), "presenterOpen": vm.PresenterOpen.Get(),
			"presenterDocLength": len(vm.PresenterDoc.Get()), "relayed": pres.relayed,
		}
	}))
	// goSlidesCaretPixel(line, col) is the CSS pixel of an editor position.
	js.Global().Set("goSlidesCaretPixel", js.FuncOf(func(_ js.Value, a []js.Value) any {
		x, y := scene.CaretPixel(a[0].Int(), a[1].Int())
		ox, oy := origin()
		return []any{ox + float64(x)/dpr, oy + float64(y)/dpr}
	}))
	// goSlidesSetSource(text) replaces the deck, as picking an example does,
	// so a harness can time a large one.
	js.Global().Set("goSlidesSetSource", js.FuncOf(func(_ js.Value, a []js.Value) any {
		if len(a) == 1 && a[0].Type() == js.TypeString {
			vm.Source.Set(a[0].String())
		}
		return nil
	}))
	js.Global().Set("goSlidesReady", true)
	webcanvas.Run(canvasID, &app{Scene: scene, frame: frame, dpr: dpr})
}
