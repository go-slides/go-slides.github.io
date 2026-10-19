# go-slides.github.io

Hugo source for **<https://go-slides.github.io/>**, the landing page of the
[`go-slides`](https://github.com/go-slides) organisation.

```sh
hugo server          # preview
hugo --minify        # build into public/ (gitignored)
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

go-slides/slides is released as v0.5.0. The status line in the hero, the
`status` of each repository and the `when` of each output in `hugo.toml` say
so; they change together at each release. An output marked **planned**
is not built.

| Workflow | What |
| --- | --- |
| `deploy-pages.yml` | builds on pull requests; on `main` deploys to Pages, then fetches the live URL and fails unless Hugo generated it |
| `ci.yml` | the site builds and names every repository |
| `links.yml` | every link of the built site resolves (`landingkit links`) |
| `docs.yml` | the organisation's repositories are all on the landing and the profile (`docscan`) |

BSD-3-Clause.
