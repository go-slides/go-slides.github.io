// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"errors"
	"fmt"
	"syscall/js"
	"time"

	playground "github.com/go-slides/go-slides.github.io/playground"
	"github.com/go-slides/go-slides.github.io/playground/store"
)

// saveDelay is how long after the typing stops the draft is saved.
const saveDelay = 500 * time.Millisecond

// The page's IndexedDB: projects (name → saved, tab, ui) and files
// ([project, path] → data, modified). A save is one readwrite transaction
// over both; the specification starts transactions whose scopes overlap in
// the order they were created, so a Copy then a Save run in that order.
const (
	dbName    = "go-slides-playground"
	dbVersion = 1
)

// guard runs f, turning a thrown JavaScript exception (indexedDB access
// throws in some private windows, or with storage blocked) into an error.
func guard(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if je, ok := r.(js.Error); ok {
				err = errors.New(je.Get("name").String())
				return
			}
			err = fmt.Errorf("%v", r)
		}
	}()
	f()
	return nil
}

// domErr names a DOMException (QuotaExceededError, AbortError, …).
func domErr(v js.Value, what string) error {
	if v.Truthy() && v.Get("name").Type() == js.TypeString {
		return errors.New(v.Get("name").String())
	}
	return errors.New(what)
}

// funcs holds the callbacks of one operation, released when it is over.
type funcs []js.Func

func (f *funcs) on(target js.Value, event string, fn func(js.Value)) {
	cb := js.FuncOf(func(_ js.Value, args []js.Value) any {
		var ev js.Value
		if len(args) > 0 {
			ev = args[0]
		}
		fn(ev)
		return nil
	})
	*f = append(*f, cb)
	target.Set("on"+event, cb)
}

func (f funcs) release() {
	for _, cb := range f {
		cb.Release()
	}
}

// openDB opens (creating it the first time) the page's database. It
// blocks, so it is called from the main goroutine, never in a callback.
func openDB() (js.Value, error) {
	type result struct {
		db  js.Value
		err error
	}
	ch := make(chan result, 1)
	var fs funcs
	err := guard(func() {
		req := js.Global().Get("indexedDB").Call("open", dbName, dbVersion)
		fs.on(req, "upgradeneeded", func(js.Value) {
			db := req.Get("result")
			names := db.Get("objectStoreNames")
			if !names.Call("contains", "projects").Bool() {
				db.Call("createObjectStore", "projects", map[string]any{"keyPath": "name"})
			}
			if !names.Call("contains", "files").Bool() {
				db.Call("createObjectStore", "files", map[string]any{"keyPath": []any{"project", "path"}})
			}
		})
		fs.on(req, "success", func(js.Value) { ch <- result{db: req.Get("result")} })
		fs.on(req, "error", func(js.Value) { ch <- result{err: domErr(req.Get("error"), "the database did not open")} })
		fs.on(req, "blocked", func(js.Value) { ch <- result{err: errors.New("the database is held by an older page")} })
	})
	if err != nil {
		fs.release()
		return js.Value{}, err
	}
	select {
	case r := <-ch:
		fs.release()
		return r.db, r.err
	case <-time.After(5 * time.Second):
		return js.Value{}, errors.New("the database did not answer")
	}
}

// idb is a store.Backend over the page's IndexedDB.
type idb struct{ db js.Value }

// filesOf is the key range of a project's files: [name, any path]. Arrays
// compare element by element and a shorter one first, so [name] is below
// every [name, path] and [name, []] above (an array sorts after a string).
func filesOf(name string) js.Value {
	return js.Global().Get("IDBKeyRange").Call("bound", []any{name}, []any{name, []any{}})
}

// tx starts a transaction over both stores; done is told once, when it
// completes or fails.
func (b *idb) tx(mode string, done func(error), body func(tx js.Value, fs *funcs)) {
	var fs funcs
	finish := func(err error) { fs.release(); done(err) }
	err := guard(func() {
		tx := b.db.Call("transaction", []any{"projects", "files"}, mode)
		fs.on(tx, "complete", func(js.Value) { finish(nil) })
		fs.on(tx, "abort", func(js.Value) { finish(domErr(tx.Get("error"), "the transaction was aborted")) })
		body(tx, &fs)
	})
	if err != nil {
		finish(err)
	}
}

func bytesOf(v js.Value) []byte {
	if !v.Truthy() {
		return nil
	}
	u8 := js.Global().Get("Uint8Array").New(v)
	b := make([]byte, u8.Get("length").Int())
	js.CopyBytesToGo(b, u8)
	return b
}

func (b *idb) Load(name string, done func(*store.Project, error)) {
	var meta js.Value
	files := map[string]store.File{}
	b.tx("readonly", func(err error) {
		if err != nil {
			done(nil, err)
			return
		}
		if !meta.Truthy() {
			done(nil, nil)
			return
		}
		ui := map[string]string{}
		if u := meta.Get("ui"); u.Truthy() {
			keys := js.Global().Get("Object").Call("keys", u)
			for i := range keys.Length() {
				k := keys.Index(i).String()
				if v := u.Get(k); v.Type() == js.TypeString {
					ui[k] = v.String()
				}
			}
		}
		saved := time.UnixMilli(int64(meta.Get("saved").Float()))
		done(store.Restore(name, saved, meta.Get("tab").String(), ui, files), nil)
	}, func(tx js.Value, fs *funcs) {
		get := tx.Call("objectStore", "projects").Call("get", name)
		fs.on(get, "success", func(js.Value) { meta = get.Get("result") })
		all := tx.Call("objectStore", "files").Call("getAll", filesOf(name))
		fs.on(all, "success", func(js.Value) {
			r := all.Get("result")
			for i := range r.Length() {
				f := r.Index(i)
				files[f.Get("path").String()] = store.File{Data: bytesOf(f.Get("data")), Modified: time.UnixMilli(int64(f.Get("modified").Float()))}
			}
		})
	})
}

func metaOf(p *store.Project) map[string]any {
	ui := map[string]any{}
	for k, v := range p.UI {
		ui[k] = v
	}
	return map[string]any{"name": p.Name, "saved": float64(p.Saved.UnixMilli()), "tab": p.Tab, "ui": ui}
}

func (b *idb) Save(p *store.Project, written, removed []string, done func(error)) {
	b.tx("readwrite", done, func(tx js.Value, _ *funcs) {
		tx.Call("objectStore", "projects").Call("put", metaOf(p))
		files := tx.Call("objectStore", "files")
		for _, w := range written {
			f, ok := p.File(w)
			if !ok {
				continue
			}
			data := js.Global().Get("Uint8Array").New(len(f.Data))
			js.CopyBytesToJS(data, f.Data)
			files.Call("put", map[string]any{"project": p.Name, "path": w, "data": data, "modified": float64(f.Modified.UnixMilli())})
		}
		for _, r := range removed {
			files.Call("delete", []any{p.Name, r})
		}
		// every request is made: commit now rather than when the task ends,
		// which a page being left (pagehide) may not reach
		if tx.Get("commit").Truthy() {
			tx.Call("commit")
		}
	})
}

func (b *idb) Copy(from, to string, done func(error)) {
	b.tx("readwrite", done, func(tx js.Value, fs *funcs) {
		projects, files := tx.Call("objectStore", "projects"), tx.Call("objectStore", "files")
		projects.Call("delete", to)
		files.Call("delete", filesOf(to))
		get := projects.Call("get", from)
		fs.on(get, "success", func(js.Value) {
			m := get.Get("result")
			if !m.Truthy() {
				return
			}
			m.Set("name", to)
			projects.Call("put", m)
			all := files.Call("getAll", filesOf(from))
			fs.on(all, "success", func(js.Value) {
				r := all.Get("result")
				for i := range r.Length() {
					f := r.Index(i)
					f.Set("project", to)
					files.Call("put", f)
				}
			})
		})
	})
}

func (b *idb) Delete(name string, done func(error)) {
	b.tx("readwrite", done, func(tx js.Value, _ *funcs) {
		tx.Call("objectStore", "projects").Call("delete", name)
		tx.Call("objectStore", "files").Call("delete", filesOf(name))
	})
}

// openStore opens the page's projects, loading the playground's before the
// view-model starts (it blocks: main goroutine only). Without IndexedDB
// the store keeps nothing, and the first save says so.
func openStore() *store.Store {
	tab := playground.NewTabID()
	db, err := openDB()
	if err != nil {
		js.Global().Get("console").Call("warn", "go-slides playground: no IndexedDB ("+err.Error()+"): drafts are not kept")
		return store.New(nil, tab)
	}
	b := &idb{db: db}
	var loaded []*store.Project
	for _, name := range []string{playground.DraftProject, playground.PreviousProject, playground.UIProject} {
		ch := make(chan *store.Project, 1)
		b.Load(name, func(p *store.Project, err error) {
			if err != nil {
				p = nil
			}
			ch <- p
		})
		loaded = append(loaded, <-ch)
	}
	return store.New(b, tab, loaded...)
}

// Projects implements playground.Projects.
func (h host) Projects() *store.Store { return h.st }

// Announce implements playground.Projects: the other tabs of the page
// hear it on the BroadcastChannel.
func (h host) Announce(name string, saved time.Time) {
	if h.bc.Truthy() {
		_ = guard(func() {
			h.bc.Call("postMessage", map[string]any{"name": name, "tab": h.st.Tab(), "saved": float64(saved.UnixMilli())})
		})
	}
}

// Prompt implements playground.Prompter (window.prompt): ok is false when the
// reader cancels.
func (host) Prompt(question, value string) (answer string, ok bool) {
	_ = guard(func() {
		v := js.Global().Call("prompt", question, value)
		if v.Type() == js.TypeString {
			answer, ok = v.String(), true
		}
	})
	return answer, ok
}

// Confirm implements playground.Confirmer.
func (host) Confirm(question string) bool {
	ok := false
	_ = guard(func() { ok = js.Global().Call("confirm", question).Truthy() })
	return ok
}

// journalKey is the write-ahead journal's localStorage key (see
// playground.Journal).
const journalKey = "go-slides.playground.journal"

// JournalWrite implements playground.Journal: a synchronous localStorage
// write (a QuotaExceededError, or blocked storage, comes back as an error).
func (host) JournalWrite(value string) error {
	return guard(func() { js.Global().Get("localStorage").Call("setItem", journalKey, value) })
}

// JournalRead implements playground.Journal.
func (host) JournalRead() (v string, ok bool) {
	_ = guard(func() {
		r := js.Global().Get("localStorage").Call("getItem", journalKey)
		if r.Type() == js.TypeString {
			v, ok = r.String(), true
		}
	})
	return v, ok
}

// JournalClear implements playground.Journal.
func (host) JournalClear() {
	_ = guard(func() { js.Global().Get("localStorage").Call("removeItem", journalKey) })
}

// newChannel opens the tabs' BroadcastChannel (zero Value without one).
func newChannel() js.Value {
	var bc js.Value
	_ = guard(func() {
		if c := js.Global().Get("BroadcastChannel"); c.Truthy() {
			bc = c.New(dbName)
		}
	})
	return bc
}

// keepDrafts saves the draft a moment after each edit, when the page is
// hidden or left (pagehide and visibilitychange, which mobile browsers fire
// where beforeunload is unreliable), and relays the other tabs' saves.
func keepDrafts(vm *playground.ViewModel, bc js.Value) {
	var timer js.Value
	// the journal first: synchronous, done when the page goes; then the
	// asynchronous IndexedDB save, which a closing page may cut short
	save := js.FuncOf(func(js.Value, []js.Value) any { vm.JournalDeck(); vm.SaveDraft(); return nil })
	vm.Source.Subscribe(func(string) {
		if timer.Truthy() {
			js.Global().Call("clearTimeout", timer)
		}
		timer = js.Global().Call("setTimeout", save, saveDelay.Milliseconds())
	})
	js.Global().Call("addEventListener", "pagehide", save)
	js.Global().Get("document").Call("addEventListener", "visibilitychange", js.FuncOf(func(js.Value, []js.Value) any {
		if js.Global().Get("document").Get("visibilityState").String() == "hidden" {
			vm.JournalDeck()
			vm.SaveDraft()
		}
		return nil
	}))
	if bc.Truthy() {
		bc.Set("onmessage", js.FuncOf(func(_ js.Value, args []js.Value) any {
			if len(args) == 0 {
				return nil
			}
			d := args[0].Get("data")
			if d.Type() != js.TypeObject || d.Get("name").Type() != js.TypeString || d.Get("tab").Type() != js.TypeString || d.Get("saved").Type() != js.TypeNumber {
				return nil
			}
			vm.DraftSavedElsewhere(d.Get("name").String(), d.Get("tab").String(), time.UnixMilli(int64(d.Get("saved").Float())))
			return nil
		}))
	}
}
