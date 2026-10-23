package playground

import (
	"slices"
	"strings"

	"github.com/go-slides/slides/editor"
	"github.com/go-widgets/mvvm"
)

// ImagePicker is what a host that can ask for a file offers: it shows its file
// picker and, once the person chose an image and the project took it into media/,
// gives its Markdown (MediaMarkdown: ![name](media/name.png), not a data: URI)
// back to Inserter.Picked with the mode it was asked for.
type ImagePicker interface {
	PickImage(mode ImageMode)
}

// Inserter is the Insert tab's part of the view-model: the commands that put a
// slide, an image, a background, math, a code block or a kind into the deck,
// each an edit of the Markdown at the caret (or around the selection). The
// editor is not known here: the scene gives the selection (Selection) and
// applies the changes (Apply), so that they join the editor's undo; with no
// editor the change is written into Source directly.
type Inserter struct {
	vm *ViewModel

	// Slide adds an empty slide after the current one; Math and Code insert a
	// $$ block and a fenced Go block with line numbers, around the selection if
	// there is one.
	Slide, Math, Code *mvvm.Command
	// Image, Background, BgLeft and BgRight ask the host for an image file: as an
	// image on the slide, as the slide's background, or as a background filling
	// the left or the right 40%.
	Image, Background, BgLeft, BgRight *mvvm.Command
	// Kind is the kind picker's selection: 0 its title, another index the kind
	// KindItems()[i] (the kinds the editor's completion knows), applied to the
	// current slide when chosen, the picker returning to its title. KindVersion
	// counts the changes of the list.
	Kind        *mvvm.Observable[int]
	KindVersion *mvvm.Observable[int]
	kinds       []string

	// Selection is the editor's selection (the scene keeps it); Apply writes a
	// change into the editor (the scene sets it).
	Selection *mvvm.Observable[Range]
	Apply     func(Change)
}

// KindTitle is the kind picker's title entry.
const KindTitle = "Kind…"

func newInserter(vm *ViewModel) *Inserter {
	in := &Inserter{
		vm:          vm,
		Kind:        mvvm.NewObservable(0),
		KindVersion: mvvm.NewObservable(0),
		Selection:   mvvm.NewObservable(Range{}),
	}
	in.Slide = mvvm.NewCommand(func() {
		in.apply(InsertSlide(vm.Source.Get(), in.caret()))
		vm.Status.Set("Added a slide after the current one.")
	}, nil)
	in.Math = mvvm.NewCommand(func() { in.block("$$", "$$", "a display formula") }, nil)
	in.Code = mvvm.NewCommand(func() { in.block("```go {lineNumbers}", "```", "a Go code block with line numbers") }, nil)
	in.Image = in.picker(ModeImage)
	in.Background = in.picker(ModeBackground)
	in.BgLeft = in.picker(ModeBgLeft)
	in.BgRight = in.picker(ModeBgRight)
	in.Kind.Subscribe(in.chooseKind)
	return in
}

// caret is where the editor's caret is.
func (in *Inserter) caret() Pos { return Pos{in.vm.CaretLine.Get(), in.vm.CaretCol.Get()} }

// apply makes a change: through the editor when the scene attached one, else in
// the source.
func (in *Inserter) apply(c Change) {
	if in.Apply != nil {
		in.Apply(c)
		return
	}
	in.vm.Source.Set(Replace(in.vm.Source.Get(), c))
	in.vm.CaretLine.Set(c.Caret.Line)
	in.vm.CaretCol.Set(c.Caret.Col)
}

func (in *Inserter) block(open, close, what string) {
	sel := in.Selection.Get()
	in.apply(Block(in.vm.Source.Get(), in.caret(), sel, open, close))
	if sel.Empty() {
		in.vm.Status.Set("Inserted " + what + ": type inside it.")
	} else {
		in.vm.Status.Set("Put the selection in " + what + ".")
	}
}

// picker is the command that asks the host for an image in a mode; with a host
// that cannot, it says so.
func (in *Inserter) picker(mode ImageMode) *mvvm.Command {
	return mvvm.NewCommand(func() {
		p, ok := in.vm.host.(ImagePicker)
		if !ok {
			in.vm.Status.Set("⚠ This page cannot ask for a file: drop or paste an image file into the editor.")
			return
		}
		p.PickImage(mode)
	}, nil)
}

// Picked is the host's answer to PickImage: the Markdown of the chosen image
// (MediaMarkdown), put into the deck as the mode asks. An image with text
// selected takes it as its alternative text.
func (in *Inserter) Picked(mode ImageMode, md string) {
	sel := in.Selection.Get()
	alt := string(mode)
	if mode == ModeImage && !sel.Empty() {
		alt = strings.TrimSpace(strings.NewReplacer("[", "", "]", "", "\n", " ").Replace(textIn(splitLines(in.vm.Source.Get()), sel)))
	}
	if alt != "" {
		md = RetagImage(md, alt)
	}
	if mode != ModeImage {
		sel = Range{} // a background goes on a line of its own, the selection stays
	}
	in.apply(Line(in.vm.Source.Get(), in.caret(), sel, md))
}

// KindItems is the kind picker's list: its title, then the kinds the editor's
// completion offers for a `_kind:` comment in this deck.
func (in *Inserter) KindItems() []string { return append([]string{KindTitle}, in.kinds...) }

// RefreshKinds asks the editor package for the kinds of the deck as it is, the
// way its completion lists them after `<!-- _kind: `, and tells the picker when
// they changed.
func (in *Inserter) RefreshKinds() {
	src := in.vm.Source.Get()
	probe := src + "\n\n<!-- _kind: "
	lines := splitLines(probe)
	last := len(lines) - 1
	var kinds []string
	req := in.vm.request() // with the project's files: their kinds count
	req.Source = probe
	for _, it := range editor.Complete(req, last, utf16Col(lines[last], len([]rune(lines[last])))) {
		if it.Kind == "kind" {
			kinds = append(kinds, it.Label)
		}
	}
	slices.Sort(kinds) // the completion's order is not stable; the picker's is
	if slices.Equal(kinds, in.kinds) {
		return
	}
	in.kinds = kinds
	in.KindVersion.Set(in.KindVersion.Get() + 1)
}

func (in *Inserter) chooseKind(i int) {
	if i <= 0 {
		return
	}
	defer in.Kind.Set(0)
	if i > len(in.kinds) {
		return
	}
	in.apply(SetKind(in.vm.Source.Get(), in.caret(), in.kinds[i-1]))
	in.vm.Status.Set("The slide is now of kind " + in.kinds[i-1] + ".")
}
