package playground

import (
	"fmt"
	"time"
)

// Exporter is the host's export engine: a second WebAssembly module
// (cmd/playground-export, go-slides' export engine, about 16 MiB gzipped)
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

// SaveItems are the Save as menu: its title, then what it saves.
var SaveItems = []string{"Save as…", "HTML page", "PDF", "PowerPoint (PPTX)", "OpenDocument (ODP)"}

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
		vm.Status.Set("⚠ This page cannot export: use the slides command.")
		return
	}
	if vm.exporting != "" {
		vm.Status.Set("An export is running: " + exportNames[vm.exporting] + ".")
		return
	}
	src := vm.Source.Get()
	if len(src) > MaxSource {
		vm.Status.Set("⚠ The deck is too large to export here: use the slides command.")
		return
	}
	vm.exporting = format
	vm.exportStart = time.Now()
	vm.Status.Set("Preparing the " + exportNames[format] + "…")
	ex.Export(format, src)
}

// ExportLoading reports the export engine's download, the first time.
func (vm *ViewModel) ExportLoading(got, total int64) {
	if vm.exporting == "" {
		return
	}
	if total > 0 {
		vm.Status.Set(fmt.Sprintf("Loading the export engine (once; your browser keeps it): %d%% of %.1f MiB", min(100, got*100/total), float64(total)/(1<<20)))
		return
	}
	vm.Status.Set(fmt.Sprintf("Loading the export engine (once; your browser keeps it): %.1f MiB", float64(got)/(1<<20)))
}

// ExportProgress reports the slides laid out.
func (vm *ViewModel) ExportProgress(done, total int) {
	if vm.exporting == "" {
		return
	}
	vm.Status.Set(fmt.Sprintf("%s: slide %d of %d…", exportNames[vm.exporting], done, total))
}

// Exported ends an export: the document was offered to save (name, size,
// slides), or err says why not.
func (vm *ViewModel) Exported(name string, size, slides int, err error) {
	format := vm.exporting
	vm.exporting = ""
	if err != nil {
		vm.Status.Set("⚠ The " + exportNames[format] + " export failed: " + err.Error())
		return
	}
	secs := time.Since(vm.exportStart).Seconds()
	vm.Status.Set(fmt.Sprintf("Saved %s: %s, %d slide%s in %.1f s, made in your browser.", name, sizeText(size), slides, plural(slides), secs))
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
	return fmt.Sprintf("%d bytes", n)
}
