# go-slides.github.io

Hugo source for **<https://go-slides.github.io/>**, the landing page of the
[`go-slides`](https://github.com/go-slides) organisation.

```sh
hugo server          # preview
hugo --minify        # build into public/ (gitignored)
```

## The playground: `/playground/`

`playground/` is a nested Go module: the playground at
<https://go-slides.github.io/playground/>. One WebAssembly module runs twice:

- in the page, as a [go-widgets](https://github.com/go-widgets) canvas
  application (the same shell as go-fft's and go-tex's playgrounds): the
  editor with completion from go-slides' `editor` package, the help pane,
  the problems, the toolbar;
- in a Web Worker (`static/playground/worker.js`), as the renderer, so typing
  never waits for go-slides. The page compiles the module once and hands it
  to the worker; a render that takes over 5 s stops the worker and starts
  another.

The preview is an iframe kept over the canvas's preview pane. It is
sandboxed without `allow-same-origin` and its document carries a Content
Security Policy that runs only its own scripts (by nonce or hash); the
deck's raw HTML follows Marp's allowlist. A shared link carries the deck in
its fragment, deflated: nothing is sent to any server.

**Present** opens the presenter view in a second window. The page opens the
window itself (`about:blank`, same origin) and writes into it a shell holding
another iframe, sandboxed the same way, with go-slides' presenter view. The
two sandboxed pages have no storage to share, so they sync through their
parent (`viewer.SyncParent` in go-slides): each posts `{slidesSync}` to its
parent and the page relays it to the other. No iframe that runs the deck is
given `allow-same-origin`.

An image file dropped or pasted into the editor goes into the deck as a
`data:` URI (`playground.ImageMarkdown`). Its type is sniffed from its bytes
(PNG, JPEG, GIF, WebP, SVG), and it is limited to 256 KiB. The preview's CSP
allows `data:` and `https:` images. The editor colours front-matter keys at
any depth, in YAML (block and flow) and HCL.

While the module downloads, a carousel presents go-slides. Its cards are
listed in `data/playground_cards.json`, and their pictures are the slides
of `playground/cards/cards.md`, which `cmd/build` renders to JPEG with
go-slides itself (about 50 KiB in all; `--wasm=false` renders only them).
It is plain HTML, CSS and a few lines of the loader's JavaScript, and it
follows the WAI-ARIA carousel pattern. It shows only while the module is
still downloading after 300 ms, so a cached module never flashes it. It does not
rotate under `prefers-reduced-motion`, and it stops on hover and on focus.
When the editor is ready, the carousel fades out, unless the reader is in
it: then a Start button waits for them. `browsertest/tour.cjs` covers all of
this on a throttled network.

The deck survives reloads. **`playground/store`** is the page's project
store: a project is files (path → bytes) with a little metadata, kept in
IndexedDB. The deck is `deck.md` in project `draft`, the draft it last
replaced is in `previous`, and the UI state is in `ui`. A save writes only
the changed files, in one transaction. go-slides reads a project through
`Project.FS` (a `vfs.FS`). The draft is saved 500 ms after the typing stops,
and on `pagehide` and `visibilitychange`. A shared link wins over the draft,
which stays restorable from the **Draft** menu. Tabs announce their saves on
a BroadcastChannel; a tab never erases another's draft, which becomes the
previous one. Reset and the examples keep the draft the same way.
`browsertest/draft.cjs` covers reloads, tabs, a shared link and Forget.

**Save as** offers the HTML page, PDF, PowerPoint and OpenDocument.
`playground/exporter` runs the command's own path: `slides.Render` with the
box frame, then `export.EachLayout` into the vector PDF (with its outline),
the editable PPTX or the editable ODP. It runs in a second module,
`export.wasm` (`cmd/playground-export`, about 21 MiB gzipped), which the
editor downloads only on the first export, with progress. The editor then
compiles it and runs it in a worker of its own; the URL carries the deploy's
sha, so the browser keeps it cached. The 100-slide benchmark deck's PDF comes
out byte-identical to `slides --pdf`. `browsertest/export.cjs` saves each
format and checks it.

The overview is patched, not reloaded. The worker sends go-slides'
preview in parts (`editor.Result.Head` and `SlideHTML`). While the head
(the theme's style sheet) is unchanged, the page posts only the slides that
changed into the frame. The frame's own nonce'd script parses them in an
inert `<template>` and swaps them in, so the preview keeps its scroll and
lays out one slide instead of a hundred. When the document must be
replaced (another theme, or back from the slideshow), the frame has
reported its top-most slide and the new document starts there. The
**Theme** picker writes `theme:` into the front matter (YAML or HCL), and
follows it when it is edited by hand.

The editor (go-widgets' `CodeEditor`) has vertical and horizontal scrollbars
and a **Minimap** (a toggle above the editor; errors are red ticks on it).
**Slides** shows a column of slide thumbnails beside the preview, built on
go-widgets' `PageStrip`: the canvas draws the strip (numbers, the current
slide, the scrollbar, the keys) and a transparent third iframe lays the
overview's own slides over its cards, scaled by the browser, so a thumbnail is
the real slide with nothing rasterised and nothing to download. The frame
holds the same head and slides as the preview and takes the same per-slide
patches (`playground/rail.go`); the browser skips the layout of the slides
outside the strip. Clicking a thumbnail moves the editor and the preview to
the slide; the strip follows the caret and the preview's scroll. (Rasterising
the thumbnails with go-webengine in WebAssembly was measured and rejected:
19 ms a slide and 21 MiB of module to download; an SVG slide drawn on a canvas
taints it, so the browser cannot hand the pixels back either.)

The only JavaScript is the page's loader (`layouts/_default/playground.html`)
and the worker's bootstrap; everything else is Go. Building the module is
`go run ./cmd/build` (cobra), run by the deploy, never committed:

```sh
cd playground
GOWORK=off go test ./...                 # the view-model, the scene, the renderer
GOWORK=off go run ./cmd/build            # static/playground/playground.wasm + wasm_exec.js
cd .. && hugo --minify                   # the site, into public/
playground/browsertest/drive.cjs         # headless Chrome (see playground.yml)
```

## The design is not here

It comes from the shared Hugo module
[`go-fleettools/landingkit`](https://github.com/go-fleettools/landingkit),
pinned in `go.mod`. `kit.go` exists only so that `go mod tidy` (which Renovate
runs after every bump) keeps that requirement: a Hugo site has no Go code.

## The colours are read, not chosen

`[params.brand]` in `hugo.toml` is the output of
`landingkit brand static/img/logo.svg`. go-slides is in the **document
family** (go-odf, go-richdoc, go-rtf, go-docutils, go-ooxml, go-tex) and takes
its `#4F46E5 → #1E1B4B`. The logo and favicon are `go-slides/brand`'s
`svg/color/go-slides.svg`.

## What the page claims

go-slides/slides is released as v0.7.1. The status line in the hero, the
`status` of each repository and the `when` of each output in `hugo.toml` say
so; they change together at each release. An output marked **planned**
is not built.

| Workflow | What |
| --- | --- |
| `deploy-pages.yml` | builds on pull requests; on `main` deploys to Pages, then fetches the live URL and fails unless Hugo generated it |
| `ci.yml` | the site builds and names every repository |
| `links.yml` | every link of the built site resolves (`landingkit links`) |
| `docs.yml` | the organisation's repositories are all on the landing and the profile (`docscan`) |
| `playground.yml` | the playground's tests, its module and size budget, and headless Chrome driving the built page |

BSD-3-Clause.
