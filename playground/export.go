package playground

import (
	"fmt"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/versions"
)

// Exporter is the host's export engine: a second WebAssembly module
// (cmd/playground-export, go-slides' export engine, about 5 MiB gzipped)
// that the page downloads the first time an export is asked for, compiling
// it as it arrives, and runs in a worker. It reports back through ExportLoading,
// ExportProgress and Exported. A host without one offers no export.
type Exporter interface {
	// Export makes the deck's document in format ("pdf", "pptx", "odp")
	// and offers it to save.
	Export(format, source string)
}

// ExportFormats are the formats the toolbar offers, in its order.
var ExportFormats = []string{"pdf", "pptx", "odp"}

var exportNames = map[string]string{"pdf": "PDF", "pptx": "PowerPoint", "odp": "OpenDocument"}

// SaveItems are the Save as menu: its title, then what it saves. Rebuilt by
// SetLocale.
var SaveItems = saveItems()

func saveItems() []string {
	return []string{T(KSaveAs), T(KLblHTML), T(KLblPDF), T(KLblPPTX) + " (PPTX)", T(KLblODP) + " (ODP)"}
}

// chooseSave runs the chosen Save as item and returns the menu to its
// title.
func (vm *ViewModel) chooseSave(i int) {
	if i <= 0 || i >= len(SaveItems) {
		return
	}
	vm.SaveAction.Set(0)
	if i == 1 {
		vm.Download.Execute()
		return
	}
	vm.export(ExportFormats[i-2])
}

// export asks the host for a document.
func (vm *ViewModel) export(format string) {
	ex, ok := vm.host.(Exporter)
	if !ok {
		vm.Status.Set(T(KStatusNoExport))
		return
	}
	if vm.exporting != "" {
		vm.Status.Set(T(KStatusExportRunning, exportNames[vm.exporting]))
		return
	}
	src := vm.Source.Get()
	if len(src) > MaxSource {
		vm.Status.Set(T(KStatusTooLargeExport))
		return
	}
	vm.exporting = format
	vm.exportStart = time.Now()
	vm.keepVersion(vm.captureState(), versions.Export, exportNames[format], nil)
	vm.Status.Set(T(KStatusPreparing, exportNames[format]))
	ex.Export(format, src)
}

// ExportLoading reports the export engine's download, the first time.
func (vm *ViewModel) ExportLoading(got, total int64) {
	if vm.exporting == "" {
		return
	}
	if total > 0 {
		vm.Status.Set(T(KStatusLoadingEnginePct, min(100, got*100/total), float64(total)/(1<<20)))
		return
	}
	vm.Status.Set(T(KStatusLoadingEngine, float64(got)/(1<<20)))
}

// ExportProgress reports the slides laid out.
func (vm *ViewModel) ExportProgress(done, total int) {
	if vm.exporting == "" {
		return
	}
	vm.Status.Set(T(KStatusExportProgress, exportNames[vm.exporting], done, total))
}

// Exported ends an export: the document was offered to save (name, size,
// slides), or err says why not.
func (vm *ViewModel) Exported(name string, size, slides int, err error) {
	format := vm.exporting
	vm.exporting = ""
	if err != nil {
		vm.Status.Set(T(KStatusExportFailed, exportNames[format], ErrText(err)))
		return
	}
	secs := time.Since(vm.exportStart).Seconds()
	vm.Status.Set(Tn(KStatusSaved, slides, name, sizeText(size), slides, secs))
}

// ExportTimeout is how long an export may run before the host gives up
// on its worker (a deck of a few hundred slides takes seconds).
const ExportTimeout = 3 * time.Minute

// Exporting is the format of the export running ("" for none).
func (vm *ViewModel) Exporting() string { return vm.exporting }

func sizeText(n int) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MiB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%d KiB", n>>10)
	}
	return T(KBytes, n)
}
