---
slides: true
theme: default
paginate: true
footer: go-slides playground
data:
  name: World
---

<!-- _class: lead -->
<!-- _paginate: false -->

# Hello, {{ .name }}!

A deck in **Markdown**, rendered by go-slides
— in your browser, by WebAssembly.

---

## What you can write

<!-- A speaker note: Present opens the presenter view, which shows it. -->

- Headings, lists, **bold**, _italic_, `code`
- Tables, images, :sparkles: emoji
- Math: $e^{i\pi} + 1 = 0$
- Go templates: this is page {{ page }} of {{ total }}

---

## Code is highlighted

```go
func main() {
	fmt.Println("hello, slides")
}
```

---

<!-- _class: invert -->

## Try it

1. Edit this text: the slide follows.
2. Pick another example above.
3. **Slideshow** presents the deck; **Share** puts it in a link.
