// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"errors"
	"strconv"
	"syscall/js"

	playground "github.com/go-slides/go-slides.github.io/playground"
	"github.com/go-slides/go-slides.github.io/playground/store"
)

// The file panel's side of the page: files picked or dropped are read and
// handed to the view-model; the project's media are posted into the preview
// (and the presenter view), whose own script makes blob: URLs of them in its
// own origin (playground's mediaScript); a .zip is saved as bytes.

// DownloadBytes saves a binary file.
func (host) DownloadBytes(name, mime string, data []byte) {
	u8 := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(u8, data)
	blob := js.Global().Get("Blob").New([]any{u8}, map[string]any{"type": mime})
	url := js.Global().Get("URL").Call("createObjectURL", blob)
	a := js.Global().Get("document").Call("createElement", "a")
	a.Set("href", url)
	a.Set("download", name)
	a.Call("click")
	// revoked later: a click starts the download, it does not finish it
	var revoke js.Func
	revoke = js.FuncOf(func(js.Value, []js.Value) any {
		revoke.Release()
		js.Global().Get("URL").Call("revokeObjectURL", url)
		return nil
	})
	js.Global().Call("setTimeout", revoke, 60000)
}

// filesUI is what the page keeps for the panel: the view-model and the scene
// files go to, and the media already copied into JavaScript.
type filesUI struct {
	vm    *playground.ViewModel
	scene *playground.Scene
	// media are the JavaScript copies of the media, by path, kept between
	// posts while the file is the one copied
	media map[string]jsMedium
}

type jsMedium struct {
	size  int
	stamp int64    // when the file was written
	bytes js.Value // an ArrayBuffer
	typ   string
}

var files = &filesUI{media: map[string]jsMedium{}}

// Pick lets the reader pick media, a project's .zip or its folder.
func (host) Pick(what string) {
	if what == "folder" {
		if dp := js.Global().Get("showDirectoryPicker"); dp.Truthy() {
			// The File System Access API (Chromium): the folder itself, read
			// here; elsewhere an <input webkitdirectory> lists it.
			go files.readDirectory(js.Global().Call("showDirectoryPicker", map[string]any{"mode": "read"}))
			return
		}
	}
	in := js.Global().Get("document").Call("createElement", "input")
	in.Set("type", "file")
	switch what {
	case "media":
		in.Set("multiple", true)
		in.Set("accept", "image/*,video/*,audio/*,.woff,.woff2,.ttf,.otf")
	case "zip":
		in.Set("accept", ".zip,application/zip")
	case "folder":
		in.Set("webkitdirectory", true)
		in.Set("multiple", true)
	}
	var change js.Func
	change = js.FuncOf(func(js.Value, []js.Value) any {
		change.Release()
		go files.picked(what, in.Get("files"))
		return nil
	})
	in.Call("addEventListener", "change", change)
	in.Call("click")
}

// await waits for a promise, from a goroutine (never from a JavaScript
// callback, which must return first).
func await(p js.Value) (js.Value, error) {
	type result struct {
		v   js.Value
		err error
	}
	ch := make(chan result, 1)
	var ok, fail js.Func
	ok = js.FuncOf(func(_ js.Value, a []js.Value) any {
		ok.Release()
		fail.Release()
		var v js.Value
		if len(a) > 0 {
			v = a[0]
		}
		ch <- result{v: v}
		return nil
	})
	fail = js.FuncOf(func(_ js.Value, a []js.Value) any {
		ok.Release()
		fail.Release()
		msg := "failed"
		if len(a) > 0 && a[0].Truthy() && a[0].Get("message").Type() == js.TypeString {
			msg = a[0].Get("message").String()
		}
		ch <- result{err: errors.New(msg)}
		return nil
	})
	p.Call("then", ok, fail)
	r := <-ch
	return r.v, r.err
}

func mib(n int) string { return strconv.Itoa(n>>20) + " MiB" }

// readBlob reads a File's bytes, refusing one over limit before reading it.
func readBlob(f js.Value, limit int) ([]byte, error) {
	if f.Get("size").Int() > limit {
		return nil, errors.New(f.Get("name").String() + " is over " + mib(limit))
	}
	ab, err := await(f.Call("arrayBuffer"))
	if err != nil {
		return nil, err
	}
	u8 := js.Global().Get("Uint8Array").New(ab)
	b := make([]byte, u8.Get("length").Int())
	js.CopyBytesToGo(b, u8)
	return b, nil
}

// picked takes the files of an <input>.
func (u *filesUI) picked(what string, list js.Value) {
	n := list.Get("length").Int()
	switch what {
	case "zip":
		if n == 0 {
			return
		}
		b, err := readBlob(list.Call("item", 0), playground.MaxZipSize)
		if err != nil {
			u.vm.Status.Set("⚠ " + err.Error())
			return
		}
		u.vm.OpenZip(b)
	case "folder":
		if n > store.MaxFiles*4 {
			u.vm.Status.Set("⚠ the folder holds " + strconv.Itoa(n) + " files: a project holds at most " + strconv.Itoa(store.MaxFiles))
			return
		}
		in := map[string][]byte{}
		total := 0
		for i := range n {
			f := list.Call("item", i)
			if total += f.Get("size").Int(); total > store.MaxProjectSize {
				u.vm.Status.Set("⚠ the folder is over " + mib(store.MaxProjectSize))
				return
			}
			b, err := readBlob(f, store.MaxFileSize)
			if err != nil {
				u.vm.Status.Set("⚠ " + err.Error())
				return
			}
			in[f.Get("webkitRelativePath").String()] = b
		}
		u.vm.OpenFolder(in)
	default:
		u.add("", list, false)
	}
}

// readDirectory reads a folder picked with the File System Access API.
func (u *filesUI) readDirectory(picked js.Value) {
	dir, err := await(picked)
	if err != nil {
		return // the reader closed the picker
	}
	in := map[string][]byte{}
	total := 0
	var walk func(h js.Value, prefix string) error
	walk = func(h js.Value, prefix string) error {
		it := h.Call("values")
		for {
			next, err := await(it.Call("next"))
			if err != nil {
				return err
			}
			if next.Get("done").Bool() {
				return nil
			}
			e := next.Get("value")
			name := prefix + e.Get("name").String()
			if e.Get("kind").String() == "directory" {
				if err := walk(e, name+"/"); err != nil {
					return err
				}
				continue
			}
			if len(in) >= store.MaxFiles*4 {
				return errors.New("the folder holds more than " + strconv.Itoa(store.MaxFiles*4) + " files")
			}
			f, err := await(e.Call("getFile"))
			if err != nil {
				return err
			}
			if total += f.Get("size").Int(); total > store.MaxProjectSize {
				return errors.New("the folder is over " + mib(store.MaxProjectSize))
			}
			b, err := readBlob(f, store.MaxFileSize)
			if err != nil {
				return err
			}
			in[name] = b
		}
	}
	if err := walk(dir, ""); err != nil {
		u.vm.Status.Set("⚠ " + err.Error())
		return
	}
	u.vm.OpenFolder(in)
}

// add stores files dropped, pasted or picked: in folder dir when over the
// panel, else media in media/, and their Markdown at the deck's caret
// (insert).
func (u *filesUI) add(dir string, list js.Value, insert bool) {
	for i := range list.Get("length").Int() {
		f := list.Call("item", i)
		b, err := readBlob(f, store.MaxFileSize)
		if err != nil {
			u.vm.Status.Set("⚠ " + err.Error())
			continue
		}
		p, err := u.vm.AddFile(dir, f.Get("name").String(), b)
		if err != nil {
			u.vm.Status.Set("⚠ " + err.Error())
			continue
		}
		if insert && playground.IsMedia(p) {
			u.scene.InsertImage(playground.MediaMarkdown(p))
		}
		u.vm.Status.Set("Added " + p + " to the project (" + strconv.Itoa(len(b)>>10) + " KiB): it is kept in this browser and goes in the project's .zip.")
	}
}

// dropped takes a drop or a paste of files at device pixel (x, y) (a paste
// has no place: the deck's caret); it reports whether there were files.
func (u *filesUI) dropped(list js.Value, x, y int, placed bool) bool {
	if !list.Truthy() || list.Get("length").Int() == 0 {
		return false
	}
	if placed {
		if dir, over := u.scene.FileDropDir(x, y); over {
			go u.add(dir, list, false)
			return true
		}
	}
	go u.add("", list, true)
	return true
}

// postMedia posts the project's media into a frame's window (or, for the
// presenter view, its window, which passes them on).
func (u *filesUI) postMedia(target js.Value) {
	if !target.Truthy() {
		return
	}
	media := u.vm.Media()
	msg := map[string]any{}
	for p, b := range media {
		m, ok := u.media[p]
		if stamp := u.vm.Modified(p); !ok || m.size != len(b) || m.stamp != stamp {
			u8 := js.Global().Get("Uint8Array").New(len(b))
			js.CopyBytesToJS(u8, b)
			m = jsMedium{size: len(b), stamp: stamp, bytes: u8.Get("buffer"), typ: playground.MediaType(p, b)}
			u.media[p] = m
		}
		msg[p] = map[string]any{"type": m.typ, "bytes": m.bytes}
	}
	for p := range u.media {
		if _, ok := media[p]; !ok {
			delete(u.media, p)
		}
	}
	target.Call("postMessage", map[string]any{"goSlidesMedia": msg}, "*")
}

// filePaths lists the panel's entries for the test harness: "path" for a
// file, "path/" for a folder, joined by "|".
func filePaths(vm *playground.ViewModel) string {
	var b []byte
	for _, f := range vm.Files.Get() {
		if len(b) > 0 {
			b = append(b, '|')
		}
		b = append(b, f.Path...)
		if f.Dir {
			b = append(b, '/')
		}
	}
	return string(b)
}

// keepFiles saves the draft a moment after a change to the files.
func keepFiles(vm *playground.ViewModel) {
	var timer js.Value
	save := js.FuncOf(func(js.Value, []js.Value) any { vm.SaveDraft(); return nil })
	vm.FilesSaved = func() {
		if timer.Truthy() {
			js.Global().Call("clearTimeout", timer)
		}
		timer = js.Global().Call("setTimeout", save, saveDelay.Milliseconds())
	}
}
