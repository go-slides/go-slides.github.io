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
  the problems, and a command ribbon (go-widgets' `Ribbon`, `playground/ribbon.go`);
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

**Files** is the deck's project, left of the editor (a go-widgets
`toolkit.FileTree` over the draft project of the store): `deck.md`,
`slides.yaml`, `themes/`, `kinds/`, `media/` — the layout the `slides`
command and the VS Code extension read. A file opens in the editor with its
language (Markdown, YAML, HCL, JSON, CSS) and its diagnostics; rename,
new file or folder, delete (confirmed), drag to move, context menu, APG
treeview keys. New theme (built on default, gaia or uncover with `@import`),
new kind and new data file start from templates and are listed in
`slides.yaml`; the project's themes join the Theme list. The whole project
downloads as a `.zip` (unzipped, `slides build deck.md` builds it) and opens
from a `.zip` or a folder (File System Access API where there is one, an
`<input webkitdirectory>` elsewhere). An import refuses paths that leave the
project, more than 256 files, 8 MiB a file, 64 MiB in all, and archives that
inflate more than 100 times (`playground/zip.go`, fuzzed).

Media dropped or pasted on the editor go into `media/` as bytes, and the
deck names them (`![](media/x.png)`): the deck stays small and readable. The
preview's frame has an opaque origin, where a `blob:` URL made by the page
does not load (measured), so the page posts the media's bytes into the
frame, whose own script makes `blob:` URLs of them in its own origin and
puts them in place of the paths (`playground/media.go`); the CSP is
unchanged but for `blob:` in `media-src` and `font-src`. Export ▸ HTML page puts
the media into the page as `data:` URIs. The PDF, PowerPoint and
OpenDocument exports render the same project: the export worker is posted
the project's text files and the same media `ArrayBuffer`s as the preview
frames, reads `slides.yaml`, themes, kinds and data files through the same
`vfs.FS`, and puts the media into the slides as `data:` URIs
(`playground/mediafiles`, shared by both modules, with the bounds).
Share link carries the deck only, and says so when the project holds more.

The write-ahead journal (`playground/journal.go`) holds the project's text
files with the deck (version 2 of its entry): a theme edited and the tab
closed at once keeps the edit; a file deleted since the last save stays
deleted. Media stay in IndexedDB alone.

While the module downloads, a carousel presents go-slides. Its cards are
listed in `data/playground_cards.json`, and their pictures are the slides
of `playground/cards/cards.md`, which `cmd/build` renders to JPEG with
go-slides itself (about 50 KiB in all; `--wasm=false` renders only them).
The carousel is the landing kit's `loading-carousel.html` (landingkit
v0.4.0). The loader calls `landingkitCarousel.downloaded()` and
`landingkitCarousel.done(reveal, onWait)`. It follows the WAI-ARIA carousel
pattern. It shows only while the module is
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

IndexedDB is asynchronous, and a closing page may not wait for the write. So
the deck's text is first written, synchronously, to a write-ahead journal
in `localStorage` (`playground.Journal`): on each save tick, and on
`pagehide` and when the page is hidden. At the next load, a journal newer
than the stored draft is replayed into it, stored, then cleared. Media stay
in IndexedDB. A deck over the journal's budget (1 Mi characters) falls back
to IndexedDB alone, and the status says so once.

The **ribbon** has four tabs, all views on the view-model: **Home** (the
example, Reset, the theme, Present, Share link, the Draft menu), **Insert**
(New slide, Image, Background with its left and right split, Math, a Go code
block with line numbers, a Kind picker), **View** (the preview mode, Present,
the Minimap and Slides toggles) and **Export**. A button is an `Action` over an
`mvvm.Command`; a drop-down or a toggle shares the observable the view-model
already has. It scales down by group priority as the window narrows, collapses
to its tabs (Ctrl+F1 or the chevron), shows keytips on Alt, and its tooltips
give a description. The Insert tab's commands (`playground/insert.go`,
`insertvm.go`) are text edits at the caret, or around the selection, made
through the editor so Undo takes them back, and always in the deck (a project
file open in the editor is left alone). Image and Background open the browser's
file picker and store the file in the project's `media/` like a file dropped on
the panel (`cmd/playground-wasm/pickimage.go`), the deck then referring to
`media/x.png`. The browser tests click the tabs and controls by the names
`goSlidesDebug().rects` gives them (`tab-View`, `mode`, `save-pdf`, ...).

The panes beside the editor (the file panel, the thumbnails, the help) each
have a width of their own and a toggle on the View tab, and the ribbon
collapses (Ctrl+F1): the editor gets the larger share of what is left, about
60 columns even at 1280 px. A window starts with the file panel off below
1440 CSS px and the help off below 1700; once the person switches a pane or the
ribbon, the choice is kept with the rest of the UI state (`ui` project) and wins
over the width (`playground/panels.go`).

**Export** offers the HTML page, PDF, PowerPoint and OpenDocument.
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

The kit (v0.5.0) adds two things to this site:

- **The hero is a carousel.** The kit's `hero-carousel.html` runs over the
  `[[params.hero.cards]]` in `hugo.toml`:
  - the terminal, first, so the first paint and its LCP are unchanged; it
    lives in `layouts/partials/hero-terminal.html`;
  - a picture of the playground, `static/img/hero/playground.webp`;
  - four of the loading carousel's pictures (`/playground/cards/<n>.jpg`,
    rendered by the deploy);
  - a VS Code card.

  `playground/browsertest/hero.cjs` drives it in CI.
- **The mantra strip.** `LICENSE=BSD-3-Clause` and `CGO_ENABLED=0` sit under
  the top bar. The playground page, a `playground` layout, has no strip.

## The colours are read, not chosen

`[params.brand]` in `hugo.toml` is the output of
`landingkit brand static/img/logo.svg`. go-slides is in the **document
family** (go-odf, go-richdoc, go-rtf, go-docutils, go-ooxml, go-tex) and takes
its `#4F46E5 → #1E1B4B`. The logo and favicon are `go-slides/brand`'s
`svg/color/go-slides.svg`.

## What the page claims

go-slides/slides is released as v0.9.0. The status line in the hero, the
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
