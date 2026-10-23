package playground

import (
	"regexp"
	"strings"
)

// Aligning the content of a slide. Marp has no alignment syntax; the ways to do it
// (bibliography, opened for this change) are
//
//   - a class from the theme (<!-- _class: lead --> centres the title slide of the
//     default theme only);
//   - <div style="text-align: center">, which needs raw HTML on (marp --html): the
//     allow-list of marp-core 4.4 that go-slides keeps by default (deck/sanitize.go)
//     drops the style attribute of every tag;
//   - a <style scoped> block in the slide: Marpit takes the block over before any
//     HTML is filtered (deck/comments.go, isStyleBlock), so it works with raw HTML
//     off, and the preview, the HTML page and the PDF all show it.
//
// The Format tab uses the third: the block it writes is the three lines below, and
// it only ever changes a block of exactly that form, so a style the person wrote
// themselves is left alone.

// SlideAlignment is how the content of a slide is aligned.
type SlideAlignment string

// The alignments of a slide. Left is the theme's own: it is written as no block.
const (
	SlideLeft   SlideAlignment = "left"
	SlideCenter SlideAlignment = "center"
	SlideRight  SlideAlignment = "right"
)

var alignRule = regexp.MustCompile(`^\s*section\s*\{\s*text-align:\s*(center|right)\s*;?\s*\}\s*$`)

// alignBlock finds the block this command wrote in the slide [start, end): the line
// of its `<style scoped>` and the alignment it sets.
func alignBlock(lines []string, start, end int) (line int, a SlideAlignment, ok bool) {
	for i := start; i+2 < end; i++ {
		if strings.TrimSpace(lines[i]) != "<style scoped>" || strings.TrimSpace(lines[i+2]) != "</style>" {
			continue
		}
		if m := alignRule.FindStringSubmatch(lines[i+1]); m != nil {
			return i, SlideAlignment(m[1]), true
		}
	}
	return 0, SlideLeft, false
}

// SlideAlignmentAt is the alignment of the slide the caret is in.
func SlideAlignmentAt(src string, caret Pos) SlideAlignment {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	start, end := slideBounds(lines, caret.Line)
	_, a, _ := alignBlock(lines, start, end)
	return a
}

// SetSlideAlignment aligns the content of the slide the caret is in: a block
// `<style scoped> section { text-align: … } </style>` at the top of its text (the
// block is rewritten when there is one, removed for left, the theme's own). Asking
// for the alignment the slide has returns it to left, as the toggle buttons do.
func SetSlideAlignment(src string, caret Pos, a SlideAlignment) Change {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	start, end := slideBounds(lines, caret.Line)
	at, has, ok := alignBlock(lines, start, end)
	if ok && has == a {
		a = SlideLeft
	}
	rule := "section { text-align: " + string(a) + "; }"
	switch {
	case ok && a == SlideLeft:
		// remove the three lines and the blank line after them
		to := at + 3
		if to < len(lines) && isBlank(lines[to]) {
			to++
		}
		c := Change{From: Pos{at, 0}, To: Pos{to, 0}, Caret: caret}
		if caret.Line >= to {
			c.Caret.Line -= to - at
		} else if caret.Line >= at {
			c.Caret = Pos{at, 0}
		}
		if to >= len(lines) { // the block was the last thing: take the line before's break
			c.To = Pos{len(lines) - 1, len([]rune(lines[len(lines)-1]))}
		}
		return c
	case ok:
		c := Change{From: Pos{at + 1, 0}, To: Pos{at + 1, len([]rune(lines[at+1]))}, Text: rule, Caret: caret}
		return c
	case a == SlideLeft:
		return Change{From: caret, To: caret, Caret: caret}
	}
	// a new block, before the first line of text of the slide
	first := start
	for first < end && isBlank(lines[first]) {
		first++
	}
	block := "<style scoped>\n" + rule + "\n</style>\n\n"
	c := Change{From: Pos{first, 0}, To: Pos{first, 0}, Text: block, Caret: caret}
	if first >= len(lines) {
		c.Text = "<style scoped>\n" + rule + "\n</style>"
	}
	if caret.Line >= first {
		c.Caret.Line += 4
	}
	return c
}
