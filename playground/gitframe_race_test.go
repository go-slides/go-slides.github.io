package playground

import (
	"strings"
	"testing"

	"github.com/go-browserhttp/browsergit/gitd"
	"github.com/go-slides/go-slides.github.io/playground/store"
)

// go-slides/go-slides.github.io#133: a page served by a slides-gitd front has no
// client mode, so it must never mount the git frame (client mode's credential
// frame, an iframe on another origin). Until the page knows it has no server, it
// is in its default mode, which is client mode: the frame was mounted in that
// window whenever the Git pane was open, and gitd.cjs saw the request to
// shell.html. The wasm host mounts the frame only where GitFrameRect says there is
// a slot, so "no slot" is "no frame request".
func TestTheGitFrameIsNotMountedBeforeTheServerHasAnswered(t *testing.T) {
	// the stored project is set aside and the session is slow: the marker was there,
	// the answer to "who is signed in" is not
	for name, answer := range map[string]*fakeGitd{
		"a signed-in person":         signedIn("github:o/r"),
		"a sign-in that ended":       {err: gitd.ErrSignIn},
		"a front that is down":       {err: gitd.ErrNoServer},
		"a server that is not there": {err: gitd.ErrNoServer},
	} {
		var parked [][2]func()
		front := name != "a server that is not there"
		holdStorage, parkRuns, frontMarker = true, &parked, front
		var ft *frameTab
		s, vm, _ := gitdSceneHost(t, store.NewMemory(), answer, "", "", "", func(h *gitTab) Host {
			if ft == nil {
				ft = &frameTab{gitTab: h}
			}
			return ft
		})
		holdStorage, parkRuns, frontMarker = false, nil, false

		vm.GitPane.Set(true)
		s.layout()
		if len(parked) != 1 {
			t.Fatalf("%s: control: the session should be pending, parked %d", name, len(parked))
		}
		if vm.GitMode.Get() != GitModeClient || !vm.frameOn() {
			t.Fatalf("%s: control: the page is in client mode with a frame while it waits: mode %d", name, vm.GitMode.Get())
		}
		if vm.FrameAllowed() {
			t.Errorf("%s: the frame is allowed before the server has answered", name)
		}
		if r, ok := s.GitFrameRect(); ok {
			t.Errorf("%s: a slot for the frame, %v, before the server has answered: the host would mount it", name, r)
		}

		answer1(&parked)
		r, ok := s.GitFrameRect()
		if front {
			// a front: signed in, signed out or down, never client mode, never a frame
			if ok || vm.GitMode.Get() == GitModeClient {
				t.Errorf("%s: mode %d, frame slot %v %v", name, vm.GitMode.Get(), r, ok)
			}
			continue
		}
		// plainly no server: an ordinary page, whose client mode runs through the frame
		if vm.GitMode.Get() != GitModeClient || !vm.FrameAllowed() {
			t.Errorf("%s: mode %d, allowed %v", name, vm.GitMode.Get(), vm.FrameAllowed())
		}
		if _, has := s.Rects()["git-frame"]; !has {
			t.Errorf("%s: no frame slot once the page knows it has no server (rects %v)", name, strings.Join(rectNames(s.Rects()), ","))
		}
	}
}

func rectNames[V any](m map[string]V) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
