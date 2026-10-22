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
