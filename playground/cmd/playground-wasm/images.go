// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"syscall/js"

	playground "github.com/go-slides/go-slides.github.io/playground"
)

// insertFiles puts the image files of a paste or a drop into the deck, one
// after the other, and reports whether there were files to take. Each is
// read only when its size allows it; the bytes decide what it is
// (playground.ImageMarkdown).
func insertFiles(files js.Value, scene *playground.Scene, vm *playground.ViewModel) bool {
	if !files.Truthy() || files.Get("length").Int() == 0 {
		return false
	}
	for i := range files.Get("length").Int() {
		f := files.Call("item", i)
		name := f.Get("name").String()
		if size := f.Get("size").Int(); size > playground.MaxImage {
			_, err := playground.ImageMarkdown(name, make([]byte, size))
			vm.Status.Set("⚠ " + err.Error())
			continue
		}
		var then, fail js.Func
		then = js.FuncOf(func(_ js.Value, args []js.Value) any {
			defer then.Release()
			defer fail.Release()
			u8 := js.Global().Get("Uint8Array").New(args[0])
			b := make([]byte, u8.Get("length").Int())
			js.CopyBytesToGo(b, u8)
			md, err := playground.ImageMarkdown(name, b)
			if err != nil {
				vm.Status.Set("⚠ " + err.Error())
				return nil
			}
			scene.InsertImage(md)
			vm.Status.Set("Inserted " + name + " into the deck as a data: URI: it travels with the deck, nothing is uploaded.")
			return nil
		})
		fail = js.FuncOf(func(js.Value, []js.Value) any {
			defer then.Release()
			defer fail.Release()
			vm.Status.Set("⚠ " + name + " could not be read.")
			return nil
		})
		f.Call("arrayBuffer").Call("then", then, fail)
	}
	return true
}
