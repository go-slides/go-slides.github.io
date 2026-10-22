---
slides: true
theme: default
paginate: true
# These images are in the deck itself, as data: URIs, so the example is one
# file. An image dropped on the editor goes into media/ instead, and the deck
# names it: ![bg](media/photo.jpg).
data:
  img:
    chart: data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA0MDAgMjQwIj48cmVjdCB3aWR0aD0iNDAwIiBoZWlnaHQ9IjI0MCIgcng9IjEyIiBmaWxsPSIjZjZmOGZhIi8+PGcgZmlsbD0iIzA5NjlkYSI+PHJlY3QgeD0iNDAiIHk9IjE0MCIgd2lkdGg9IjUwIiBoZWlnaHQ9IjcwIi8+PHJlY3QgeD0iMTEwIiB5PSI5MCIgd2lkdGg9IjUwIiBoZWlnaHQ9IjEyMCIvPjxyZWN0IHg9IjE4MCIgeT0iMTEwIiB3aWR0aD0iNTAiIGhlaWdodD0iMTAwIi8+PHJlY3QgeD0iMjUwIiB5PSI1MCIgd2lkdGg9IjUwIiBoZWlnaHQ9IjE2MCIvPjwvZz48cmVjdCB4PSIzMjAiIHk9IjMwIiB3aWR0aD0iNTAiIGhlaWdodD0iMTgwIiBmaWxsPSIjYmYzOTg5Ii8+PHBhdGggZD0iTTMwIDIxMGgzNTAiIHN0cm9rZT0iIzU3NjA2YSIgc3Ryb2tlLXdpZHRoPSIzIi8+PC9zdmc+
    sky: data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxMjgwIDcyMCI+PGRlZnM+PGxpbmVhckdyYWRpZW50IGlkPSJzIiB4Mj0iMCIgeTI9IjEiPjxzdG9wIG9mZnNldD0iMCIgc3RvcC1jb2xvcj0iIzFlM2E4YSIvPjxzdG9wIG9mZnNldD0iLjYiIHN0b3AtY29sb3I9IiNmNTllMGIiLz48c3RvcCBvZmZzZXQ9IjEiIHN0b3AtY29sb3I9IiNmZGU2OGEiLz48L2xpbmVhckdyYWRpZW50PjwvZGVmcz48cmVjdCB3aWR0aD0iMTI4MCIgaGVpZ2h0PSI3MjAiIGZpbGw9InVybCgjcykiLz48Y2lyY2xlIGN4PSI5MDAiIGN5PSI0MzAiIHI9IjExMCIgZmlsbD0iI2ZmZjdlZCIgb3BhY2l0eT0iLjkiLz48cGF0aCBkPSJNMCA1NjAgMjYwIDMzMCA0NzAgNTIwIDcwMCAzMDAgOTgwIDU2MCAxMTYwIDQyMCAxMjgwIDUyMFY3MjBIMHoiIGZpbGw9IiMzMTJlODEiLz48cGF0aCBkPSJNMCA2NDAgMzAwIDQ3MCA1NjAgNjQwIDgyMCA1MDAgMTI4MCA2NjBWNzIwSDB6IiBmaWxsPSIjMWUxYjRiIi8+PC9zdmc+
    leaves: data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA2MDAgOTAwIj48cmVjdCB3aWR0aD0iNjAwIiBoZWlnaHQ9IjkwMCIgZmlsbD0iIzE0NTMyZCIvPjxnIGZpbGw9IiMyMmM1NWUiIG9wYWNpdHk9Ii44NSI+PGVsbGlwc2UgY3g9IjE1MCIgY3k9IjE2MCIgcng9IjEyMCIgcnk9IjUwIiB0cmFuc2Zvcm09InJvdGF0ZSgtMzAgMTUwIDE2MCkiLz48ZWxsaXBzZSBjeD0iNDIwIiBjeT0iMzAwIiByeD0iMTQwIiByeT0iNTUiIHRyYW5zZm9ybT0icm90YXRlKDI1IDQyMCAzMDApIi8+PGVsbGlwc2UgY3g9IjIwMCIgY3k9IjUyMCIgcng9IjE1MCIgcnk9IjYwIiB0cmFuc2Zvcm09InJvdGF0ZSgtMjAgMjAwIDUyMCkiLz48ZWxsaXBzZSBjeD0iNDUwIiBjeT0iNzIwIiByeD0iMTMwIiByeT0iNTAiIHRyYW5zZm9ybT0icm90YXRlKDM1IDQ1MCA3MjApIi8+PC9nPjxnIGZpbGw9IiM4NmVmYWMiIG9wYWNpdHk9Ii43Ij48Y2lyY2xlIGN4PSI1MjAiIGN5PSIxMTAiIHI9IjQwIi8+PGNpcmNsZSBjeD0iODAiIGN5PSI4MjAiIHI9IjU1Ii8+PGNpcmNsZSBjeD0iMzMwIiBjeT0iNDIwIiByPSIzMCIvPjwvZz48L3N2Zz4=
---

![bg]({{ .img.sky }})

<!-- _class: invert -->
<!-- _paginate: false -->

# Math, images and backgrounds

A full-slide background: `![bg](url)`

---

## Math, typeset by go-tex

Inline, $e^{i\pi} + 1 = 0$, and on its own line:

$$
\int_{-\infty}^{\infty} e^{-x^2}\,dx = \sqrt{\pi}
\qquad
\sum_{k=1}^{n} k = \frac{n(n+1)}{2}
$$

---

## Images, sized and filtered

![w:340]({{ .img.chart }}) ![w:340 grayscale]({{ .img.chart }})

`![w:340](url)` sets the width; `grayscale`, `sepia`, `blur`, `invert`… are filters.

---

![bg left:40%]({{ .img.leaves }})

## A split background

`![bg left:40%](url)`: the image takes the left 40 % of the slide, the text the rest. `right:` puts it on the other side.

Drop or paste an image file into the editor to put it in your deck.

---

![bg brightness:.5]({{ .img.leaves }})
![bg brightness:.5]({{ .img.sky }})

<!-- _class: invert -->

## Several backgrounds

Side by side; `![bg vertical]` stacks them, and filters work here too: `brightness:.5`.
