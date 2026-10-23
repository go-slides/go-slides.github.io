// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"slices"
	"strings"

	playground "github.com/go-slides/go-slides.github.io/playground"
)

// mediaList is the project's media paths, sorted and joined by "|", for
// goSlidesDebug: what a browser test asserts the explorer holds.
func mediaList(vm *playground.ViewModel) string {
	var paths []string
	for p := range vm.Media() {
		paths = append(paths, p)
	}
	slices.Sort(paths)
	return strings.Join(paths, "|")
}
