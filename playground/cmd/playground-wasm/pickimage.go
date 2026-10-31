// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"syscall/js"

	playground "github.com/go-slides/go-slides.github.io/playground"
	"github.com/go-slides/go-slides.github.io/playground/store"
)

// PickImage implements playground.ImagePicker, for the Insert tab's Image and
// Background controls: the browser's file picker (the click that opened it is the
// person's own, so the browser allows it), then the chosen files go into the
// project's media/ and their Markdown into the deck, as the mode asks.
func (host) PickImage(mode playground.ImageMode) {
	input := js.Global().Get("document").Call("createElement", "input")
	input.Set("type", "file")
	input.Set("accept", "image/*")
	var onchange js.Func
	onchange = js.FuncOf(func(js.Value, []js.Value) any {
		defer onchange.Release()
		go files.addPicked(input.Get("files"), mode)
		return nil
	})
	input.Call("addEventListener", "change", onchange)
	input.Call("click")
}

// addPicked stores the picked files in the project (media/) and puts each into
// the deck as the mode asks (playground.Inserter.Picked): the same store as a
// file dropped on the panel, so the deck refers to media/x.png, not to bytes.
func (u *filesUI) addPicked(list js.Value, mode playground.ImageMode) {
	for i := range list.Get("length").Int() {
		f := list.Call("item", i)
		b, err := readBlob(f, store.MaxFileSize)
		if err != nil {
			u.vm.Status.Set("⚠ " + err.Error())
			continue
		}
		p, err := u.vm.AddFile("", f.Get("name").String(), b)
		if err != nil {
			u.vm.Status.Set("⚠ " + err.Error())
			continue
		}
		if !playground.IsMedia(p) {
			u.vm.Status.Set(playground.T(playground.KWarn, playground.T(playground.KErrNotImage, p)))
			continue
		}
		u.vm.Insert.Picked(mode, playground.MediaMarkdown(p))
		u.vm.Status.Set(playground.T(playground.KStatusAddedPut, p, len(b)>>10))
	}
}
