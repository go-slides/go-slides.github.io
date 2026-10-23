//go:build js && wasm

package main

import "github.com/go-slides/go-slides.github.io/playground"

// formatState is the Format tab's toggles that are pressed, by name, for the
// browser test (goSlidesDebug().format).
func formatState(vm *playground.ViewModel) map[string]any {
	out := map[string]any{}
	for _, name := range playground.States {
		if vm.Format.State[name].Get() {
			out[name] = true
		}
	}
	return out
}
