//go:build js && wasm

package main

import (
	"github.com/go-slides/go-slides.github.io/playground"
)

// themeState is the Theme pane's state for a test harness (goSlidesDebug).
func themeState(vm *playground.ViewModel, s *playground.Scene) map[string]any {
	values, st, ok := vm.ThemeValues()
	vals := map[string]any{}
	for k, v := range values {
		vals[k] = v
	}
	meta, _, _ := vm.ThemeMetaValues()
	m := map[string]any{}
	for k, v := range meta {
		m[k] = v
	}
	contrast := []any{}
	for i := range 4 {
		contrast = append(contrast, s.ContrastText(i))
	}
	return map[string]any{
		"pane": vm.ThemePane.Get(), "name": vm.ThemeName.Get(), "ok": ok, "block": int(st),
		"values": vals, "meta": m, "note": vm.ThemeNote.Get(), "contrast": contrast,
		"sheet": vm.ThemeSheet(),
	}
}
