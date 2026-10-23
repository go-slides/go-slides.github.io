package playground

import (
	"strings"
	"testing"
)

// pickerHost is a host that can show a file picker.
type pickerHost struct {
	*fakeHost
	picked []ImageMode
}

func (h *pickerHost) PickImage(m ImageMode) { h.picked = append(h.picked, m) }

func insertVM(t *testing.T, src string) (*ViewModel, *pickerHost) {
	t.Helper()
	h := &pickerHost{fakeHost: newHost()}
	vm := NewViewModel(&fakeTransport{hold: true}, h, "")
	vm.Source.Set(src)
	return vm, h
}

func TestInsertSlideCommand(t *testing.T) {
	vm, _ := insertVM(t, "# One\n\ntext\n\n---\n# Two")
	vm.CaretLine.Set(0)
	vm.Insert.Slide.Execute()
	got := vm.Source.Get()
	if got != "# One\n\ntext\n\n---\n\n\n\n---\n# Two" {
		t.Errorf("source %q", got)
	}
	// the caret is on the new slide, below its separator
	if vm.CaretLine.Get() != 6 || vm.CaretCol.Get() != 0 {
		t.Errorf("caret %d,%d", vm.CaretLine.Get(), vm.CaretCol.Get())
	}
	if !strings.Contains(vm.Status.Get(), "Added a slide") {
		t.Errorf("status %q", vm.Status.Get())
	}
}

func TestInsertMathAndCodeCommands(t *testing.T) {
	vm, _ := insertVM(t, "# T\n\nfoo")
	vm.CaretLine.Set(1)
	vm.CaretCol.Set(0)
	vm.Insert.Math.Execute()
	if got := vm.Source.Get(); got != "# T\n$$\n\n$$\nfoo" {
		t.Errorf("math: %q", got)
	}
	if vm.CaretLine.Get() != 2 {
		t.Errorf("the caret is inside the block: line %d", vm.CaretLine.Get())
	}
	vm.Source.Set("x\nfoo\ny")
	vm.Insert.Selection.Set(Range{Pos{1, 0}, Pos{1, 3}})
	vm.Insert.Code.Execute()
	if got := vm.Source.Get(); got != "x\n```go {lineNumbers}\nfoo\n```\ny" {
		t.Errorf("code around the selection: %q", got)
	}
	if !strings.Contains(vm.Status.Get(), "selection") {
		t.Errorf("status %q", vm.Status.Get())
	}
}

// With an editor attached the changes go through it (and its undo), not the source.
func TestInsertThroughTheEditorHook(t *testing.T) {
	vm, _ := insertVM(t, "a")
	var got []Change
	vm.Insert.Apply = func(c Change) { got = append(got, c) }
	vm.Insert.Math.Execute()
	if len(got) != 1 || vm.Source.Get() != "a" {
		t.Fatalf("hook got %d changes, source %q", len(got), vm.Source.Get())
	}
	if !strings.Contains(got[0].Text, "$$") {
		t.Errorf("change %+v", got[0])
	}
}

func TestInsertImageCommandsAskTheHost(t *testing.T) {
	vm, h := insertVM(t, "# A\n")
	vm.Insert.Image.Execute()
	vm.Insert.Background.Execute()
	vm.Insert.BgLeft.Execute()
	vm.Insert.BgRight.Execute()
	want := []ImageMode{ModeImage, ModeBackground, ModeBgLeft, ModeBgRight}
	if len(h.picked) != 4 {
		t.Fatalf("picked %v", h.picked)
	}
	for i, m := range want {
		if h.picked[i] != m {
			t.Errorf("command %d asked for %q, want %q", i, h.picked[i], m)
		}
	}
	// a host with no picker says so
	plain := NewViewModel(&fakeTransport{hold: true}, newHost(), "")
	plain.Insert.Image.Execute()
	if !strings.HasPrefix(plain.Status.Get(), "⚠") {
		t.Errorf("status %q", plain.Status.Get())
	}
}

func TestPickedImages(t *testing.T) {
	const md = "![photo](data:image/png;base64,AA==)"
	vm, _ := insertVM(t, "# A\n\ntext")
	vm.CaretLine.Set(1)
	vm.CaretCol.Set(0)
	vm.Insert.Picked(ModeBackground, md)
	if got := vm.Source.Get(); !strings.Contains(got, "![bg](data:image/png;base64,AA==)") {
		t.Errorf("background: %q", got)
	}
	vm.Source.Set("# A\n\ntext")
	vm.Insert.Picked(ModeBgLeft, md)
	if got := vm.Source.Get(); !strings.Contains(got, "![bg left:40%](") {
		t.Errorf("left split: %q", got)
	}
	vm.Source.Set("# A\n\ntext")
	vm.Insert.Picked(ModeBgRight, md)
	if got := vm.Source.Get(); !strings.Contains(got, "![bg right:40%](") {
		t.Errorf("right split: %q", got)
	}
	// an image keeps its file name as alt text, or takes the selection
	vm.Source.Set("# A\n\ncaption [x]")
	vm.Insert.Picked(ModeImage, md)
	if got := vm.Source.Get(); !strings.Contains(got, "![photo](") {
		t.Errorf("image: %q", got)
	}
	vm.Source.Set("my [logo]")
	vm.Insert.Selection.Set(Range{Pos{0, 3}, Pos{0, 9}})
	vm.Insert.Picked(ModeImage, md)
	if got := vm.Source.Get(); !strings.Contains(got, "![logo](") {
		t.Errorf("image with a selection: %q", got)
	}
	// a background does not eat the selection
	vm.Source.Set("keep this text")
	vm.Insert.Selection.Set(Range{Pos{0, 0}, Pos{0, 4}})
	vm.Insert.Picked(ModeBackground, md)
	if got := vm.Source.Get(); !strings.Contains(got, "keep this text") {
		t.Errorf("the selection was replaced: %q", got)
	}
}

func TestKindPicker(t *testing.T) {
	vm, _ := insertVM(t, Examples[2].Source) // 3-kinds.md defines kinds
	vm.Insert.RefreshKinds()                 // a render does this when it comes back
	items := vm.Insert.KindItems()
	if items[0] != KindTitle || len(items) < 2 {
		t.Fatalf("kind items %q: the deck's kinds should be listed", items)
	}
	versions := vm.Insert.KindVersion.Get()
	vm.Insert.RefreshKinds()
	if vm.Insert.KindVersion.Get() != versions {
		t.Error("unchanged kinds must not notify")
	}
	vm.CaretLine.Set(len(splitLines(vm.Source.Get())) - 1)
	vm.Insert.Kind.Set(1)
	if !strings.Contains(vm.Source.Get(), KindComment(items[1])) {
		t.Errorf("kind %q not applied:\n%s", items[1], vm.Source.Get())
	}
	if vm.Insert.Kind.Get() != 0 {
		t.Error("the picker returns to its title")
	}
	// the title, and an index past the list, do nothing
	before := vm.Source.Get()
	vm.Insert.Kind.Set(0)
	vm.Insert.Kind.Set(999)
	if vm.Source.Get() != before || vm.Insert.Kind.Get() != 0 {
		t.Error("a title or a stale index must not edit the deck")
	}
	// a deck with no kinds: only the title
	vm.Source.Set("# plain")
	vm.Insert.RefreshKinds()
	if len(vm.Insert.KindItems()) < 1 {
		t.Error("the title is always there")
	}
}
