package playground

import (
	"errors"
	"strings"
	"testing"
)

// exportHost is a host with an export engine: it records the requests.
type exportHost struct {
	*fakeHost
	asked []string
}

func (h *exportHost) Export(format, source string) {
	h.asked = append(h.asked, format+":"+source[:min(len(source), 5)])
}

func TestExportFlow(t *testing.T) {
	h := &exportHost{fakeHost: newHost()}
	tr := &fakeTransport{}
	vm := NewViewModel(tr, h, "")
	tr.vm = vm
	answer(tr, vm)
	// Save as ▸ PDF
	vm.SaveAction.Set(2)
	if vm.SaveAction.Get() != 0 || len(h.asked) != 1 || !strings.HasPrefix(h.asked[0], "pdf:") || vm.Exporting() != "pdf" {
		t.Fatal(h.asked, vm.Exporting())
	}
	// one at a time
	vm.Export["odp"].Execute()
	if len(h.asked) != 1 || !strings.Contains(vm.Status.Get(), "An export is running") {
		t.Fatal(vm.Status.Get())
	}
	vm.ExportLoading(5<<20, 20<<20)
	if !strings.Contains(vm.Status.Get(), "25% of 20.0 MiB") {
		t.Fatal(vm.Status.Get())
	}
	vm.ExportLoading(5<<20, 0)
	vm.ExportProgress(3, 4)
	if vm.Status.Get() != "PDF: slide 3 of 4…" {
		t.Fatal(vm.Status.Get())
	}
	vm.Exported("deck.pdf", 123<<10, 4, nil)
	if !strings.HasPrefix(vm.Status.Get(), "Saved deck.pdf: 123 KiB, 4 slides in") || vm.Exporting() != "" {
		t.Fatal(vm.Status.Get())
	}
	// after the end, late reports are ignored
	vm.ExportProgress(1, 1)
	if !strings.HasPrefix(vm.Status.Get(), "Saved deck.pdf") {
		t.Fatal("a late report")
	}
	// a failure is said
	vm.SaveAction.Set(3)
	vm.Exported("", 0, 0, errors.New("boom"))
	if vm.Status.Get() != "⚠ The PowerPoint export failed: boom" {
		t.Fatal(vm.Status.Get())
	}
	// Save as ▸ HTML page is the presentation, as before
	vm.SaveAction.Set(1)
	if _, ok := h.downloads["deck.html"]; !ok {
		t.Fatal("no HTML")
	}
	// a host without an engine says so
	plain := NewViewModel(&fakeTransport{}, newHost(), "")
	plain.Export["pdf"].Execute()
	if !strings.Contains(plain.Status.Get(), "cannot export") {
		t.Fatal(plain.Status.Get())
	}
	if sizeText(10) != "10 bytes" || sizeText(3<<20) != "3.0 MiB" {
		t.Fatal(sizeText(10))
	}
}
