package playground

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-browserhttp/browsergit/remote"
)

// The git frame (go-slides#90). In client mode the pane holds no token and makes
// no forge call: a document on a second origin, go-slides-usercontent.github.io
// (github.io is on the Public Suffix List, so a different site with storage of
// its own), holds the token, takes it from a password input of its own,
// remembers it in its own localStorage when the person ticks the box, does the
// GitLab sign-in, and runs the calls. This page embeds it in an iframe over the
// pane and sends it typed commands; what comes back is files, a base, a branch
// and a link. There is no command that takes or returns a token, so a script
// injected into the page has none to read, in memory, in storage or in a
// message (browsertest/git.cjs asserts it).
//
// The frame's own side is go-slides-usercontent/go-slides-usercontent.github.io;
// its CSP and the origins it accepts are declared in .second-origin.json and
// checked in CI by `secondorigin verify`.

// FrameState is what the frame announces about its connection.
type FrameState struct {
	Connected  bool
	Forge      string // "github" or "gitlab"
	Remembered bool   // the token is kept on the frame's device storage
}

// FrameHost is the page's side of the frame: the iframe, and the postMessage
// to and from it.
type FrameHost interface {
	// FrameMode reports whether client-mode forge calls go through the frame.
	FrameMode() bool
	// FrameCall sends one typed command and blocks until the answer (call it
	// from GitRun, off the UI's turn). fields are JSON-shaped values; files
	// travel apart. The answer is the command's data as JSON, and files.
	FrameCall(cmd string, fields map[string]any, files remote.Files) (json.RawMessage, remote.Files, error)
	// FrameWatch registers f, called on the UI's turn with each state the frame
	// announces, and with the current one when there is one.
	FrameWatch(f func(FrameState))
}

// frameRemote is a remote.Remote whose every call is a command to the frame.
type frameRemote struct{ h FrameHost }

// fieldsOf turns a value into the JSON-shaped map a command carries.
func fieldsOf(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r frameRemote) Load(_ context.Context, repo, ref, dir string) (remote.Files, remote.Base, error) {
	data, files, err := r.h.FrameCall("load", map[string]any{"repo": repo, "ref": ref, "dir": dir}, nil)
	if err != nil {
		return nil, remote.Base{}, err
	}
	var a struct {
		Base remote.Base `json:"base"`
	}
	if err := json.Unmarshal(data, &a); err != nil || a.Base.Repo == "" {
		return nil, remote.Base{}, errors.New("the frame's answer to load was not a base")
	}
	if files == nil {
		files = remote.Files{}
	}
	return files, a.Base, nil
}

func (r frameRemote) Propose(_ context.Context, base remote.Base, files remote.Files, p remote.Proposal) (remote.Result, error) {
	b, err := fieldsOf(base)
	if err != nil {
		return remote.Result{}, err
	}
	// The branch is the frame's to name (under slides/, never the base's): only
	// the title and the body travel.
	data, _, err := r.h.FrameCall("propose", map[string]any{"base": b, "title": p.Title, "body": p.Body}, files)
	if err != nil {
		return remote.Result{}, err
	}
	var res struct{ Branch, Commit, URL string }
	if err := json.Unmarshal(data, &res); err != nil || res.Branch == "" {
		return remote.Result{}, errors.New("the frame's answer to propose was not a result")
	}
	return remote.Result{Branch: res.Branch, Commit: res.Commit, URL: res.URL}, nil
}

func (r frameRemote) Refresh(_ context.Context, base remote.Base) (remote.Files, remote.Base, error) {
	b, err := fieldsOf(base)
	if err != nil {
		return nil, remote.Base{}, err
	}
	data, files, err := r.h.FrameCall("refresh", map[string]any{"base": b}, nil)
	if err != nil {
		return nil, remote.Base{}, err
	}
	var a struct {
		Base  remote.Base `json:"base"`
		Moved bool        `json:"moved"`
	}
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, remote.Base{}, fmt.Errorf("the frame's answer to refresh was not a base: %w", err)
	}
	if !a.Moved {
		return nil, base, nil
	}
	if files == nil {
		files = remote.Files{}
	}
	return files, a.Base, nil
}

// frameOn reports whether this view-model's client mode runs through the frame.
func (vm *ViewModel) frameOn() bool { return vm.git.frame != nil }

// initFrame puts the pane in frame mode: no token field, no sign-in, no Remember
// of its own. A token an earlier page kept in THIS origin's storage is dropped
// (it is the very thing the frame exists to keep out of this page) and the
// person is told.
func (vm *ViewModel) initFrame(fh FrameHost) {
	vm.git.frame = fh
	migrated := false
	if k, ok := vm.host.(TokenKeeper); ok {
		if _, _, tok := k.KeptToken(); !tok.IsZero() {
			tok.Forget()
			migrated = true
		}
		k.ForgetToken() // also the legacy bare-token key
	}
	vm.GitRemember.Set(false)
	off := T(KGitNoteFrameOff)
	if migrated {
		off = T(KGitNoteMigrated) + " " + off
	}
	vm.git.frameOff = off
	vm.GitNote.Set(off)
	fh.FrameWatch(vm.frameState)
}

// frameState is the frame's announcement: the pane is connected exactly when
// the frame is, to the forge the frame says.
func (vm *ViewModel) frameState(s FrameState) {
	if vm.GitMode.Get() != GitModeClient {
		return // a slides-gitd server holds the credentials; the frame is not shown
	}
	if s.Connected {
		vm.git.rem, vm.git.forge = frameRemote{vm.git.frame}, s.Forge
		for i, n := range gitForgeNames {
			if n == s.Forge {
				vm.quiet = true
				vm.GitForge.Set(i)
				vm.quiet = false
			}
		}
		vm.GitConnected.Set(true)
		if s.Remembered {
			vm.GitNote.Set(T(KGitNoteFrameKept))
		} else {
			vm.GitNote.Set(T(KGitNoteFrameOn))
		}
	} else {
		vm.git.rem, vm.git.forge = nil, ""
		vm.GitConnected.Set(false)
		vm.GitNote.Set(vm.git.frameOff)
	}
	vm.raiseGitCommands()
}
