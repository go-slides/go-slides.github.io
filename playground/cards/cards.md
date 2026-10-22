---
slides: true
theme: gaia
class: lead
# The pictures of the playground's loading carousel: one slide a card, in
# the order of data/playground_cards.json, rendered to JPEG by cmd/build
# with go-slides itself.
data:
  langs: [Go, Rust, Zig]
  sky: data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxMjgwIDcyMCI+PGRlZnM+PGxpbmVhckdyYWRpZW50IGlkPSJzIiB4Mj0iMCIgeTI9IjEiPjxzdG9wIG9mZnNldD0iMCIgc3RvcC1jb2xvcj0iIzFlM2E4YSIvPjxzdG9wIG9mZnNldD0iLjYiIHN0b3AtY29sb3I9IiNmNTllMGIiLz48c3RvcCBvZmZzZXQ9IjEiIHN0b3AtY29sb3I9IiNmZGU2OGEiLz48L2xpbmVhckdyYWRpZW50PjwvZGVmcz48cmVjdCB3aWR0aD0iMTI4MCIgaGVpZ2h0PSI3MjAiIGZpbGw9InVybCgjcykiLz48Y2lyY2xlIGN4PSI5MDAiIGN5PSI0MzAiIHI9IjExMCIgZmlsbD0iI2ZmZjdlZCIgb3BhY2l0eT0iLjkiLz48cGF0aCBkPSJNMCA1NjAgMjYwIDMzMCA0NzAgNTIwIDcwMCAzMDAgOTgwIDU2MCAxMTYwIDQyMCAxMjgwIDUyMFY3MjBIMHoiIGZpbGw9IiMzMTJlODEiLz48cGF0aCBkPSJNMCA2NDAgMzAwIDQ3MCA1NjAgNjQwIDgyMCA1MDAgMTI4MCA2NjBWNzIwSDB6IiBmaWxsPSIjMWUxYjRiIi8+PC9zdmc+
  leaves: data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA2MDAgOTAwIj48cmVjdCB3aWR0aD0iNjAwIiBoZWlnaHQ9IjkwMCIgZmlsbD0iIzE0NTMyZCIvPjxnIGZpbGw9IiMyMmM1NWUiIG9wYWNpdHk9Ii44NSI+PGVsbGlwc2UgY3g9IjE1MCIgY3k9IjE2MCIgcng9IjEyMCIgcnk9IjUwIiB0cmFuc2Zvcm09InJvdGF0ZSgtMzAgMTUwIDE2MCkiLz48ZWxsaXBzZSBjeD0iNDIwIiBjeT0iMzAwIiByeD0iMTQwIiByeT0iNTUiIHRyYW5zZm9ybT0icm90YXRlKDI1IDQyMCAzMDApIi8+PGVsbGlwc2UgY3g9IjIwMCIgY3k9IjUyMCIgcng9IjE1MCIgcnk9IjYwIiB0cmFuc2Zvcm09InJvdGF0ZSgtMjAgMjAwIDUyMCkiLz48ZWxsaXBzZSBjeD0iNDUwIiBjeT0iNzIwIiByeD0iMTMwIiByeT0iNTAiIHRyYW5zZm9ybT0icm90YXRlKDM1IDQ1MCA3MjApIi8+PC9nPjxnIGZpbGw9IiM4NmVmYWMiIG9wYWNpdHk9Ii43Ij48Y2lyY2xlIGN4PSI1MjAiIGN5PSIxMTAiIHI9IjQwIi8+PGNpcmNsZSBjeD0iODAiIGN5PSI4MjAiIHI9IjU1Ii8+PGNpcmNsZSBjeD0iMzMwIiBjeT0iNDIwIiByPSIzMCIvPjwvZz48L3N2Zz4=
kinds:
  big:
    class: lead
---

## Go templates

{{ range .langs }}
- Hello from **{{ . }}**, slide {{ page }} of {{ total }}
{{ end }}

---

<!-- _kind: big -->

# kind: big

One name, a whole set of directives.

---

<!-- _class: invert -->

## Math, typeset in Go

$$
\int_{-\infty}^{\infty} e^{-x^2}\,dx = \sqrt{\pi}
\qquad
\sum_{k=1}^{n} k = \frac{n(n+1)}{2}
$$

---

![bg left:42%]({{ .leaves }})

## Images and backgrounds

`![bg left:42%](…)` beside the text, filters and sizes.

---

<!-- _class: lead -->

# PDF · PPTX · ODP

Vector PDF with selectable text; PowerPoint and OpenDocument you can edit.

---

![bg brightness:.55]({{ .sky }})

<!-- _class: invert -->

## Presenter view

Notes, the next step and a timer, in a second window.

---

<!-- _class: lead invert -->

# VS Code

The same preview beside your Markdown.
