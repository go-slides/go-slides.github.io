package playground

import (
	"testing"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

// A remembered token is kept with the forge and the API origin it is for. The
// forge the pane shows is a UI setting every tab shares, so another tab choosing
// GitLab.com (or a GitLab sign-in returning) changes it under a page that holds a
// GitHub token: the token used to be sent as a Bearer to gitlab.com.
func TestRememberedTokenFollowsItsForge(t *testing.T) {
	const gh, ghAPI, gl, glAPI = "github", "https://api.github.com", "gitlab", "https://gitlab.com"

	// kept for the forge the connection talks to, with its origin
	m := store.NewMemory()
	_, vm1, h1 := gitScene(t, m, "")
	vm1.GitRemember.Set(true)
	vm1.GitToken.Set(planted)
	vm1.GitConnect.Execute()
	if h1.kept != planted || h1.keptForge != gh || h1.keptOrigin != ghAPI {
		t.Fatalf("kept %q for %q at %q", h1.kept, h1.keptForge, h1.keptOrigin)
	}

	// another tab, not connected, picks GitLab: the stored forge choice changes
	_, vm2, _ := gitScene(t, m, "")
	vm2.GitForge.Set(1)

	// this page reloads: the device still holds the GitHub token. It is not
	// handed to anything for GitLab: no remote is built, and the token is gone
	_, vm3, h3 := gitSceneKept(t, m, h1.kept, h1.keptForge, h1.keptOrigin)
	if vm3.forgeName() != gl {
		t.Fatalf("setup: the forge is %q", vm3.forgeName())
	}
	for i, f := range h3.forges {
		if f != gh && h3.tokens[i] == planted {
			t.Errorf("LEAK: the GitHub token was handed to the %s remote", f)
		}
	}
	if len(h3.tokens) != 0 || vm3.GitConnected.Get() || h3.kept != "" {
		t.Errorf("a token kept for another forge must be dropped unused: tokens %v, connected %v, kept %q",
			h3.tokens, vm3.GitConnected.Get(), h3.kept)
	}

	// the other way round: a GitLab token, a page that shows GitHub
	m = store.NewMemory()
	_, vm4, h4 := gitScene(t, m, "")
	vm4.GitForge.Set(1)
	vm4.GitRemember.Set(true)
	vm4.GitToken.Set(planted)
	vm4.GitConnect.Execute()
	if h4.keptForge != gl || h4.keptOrigin != glAPI {
		t.Fatalf("a GitLab token is kept for %q at %q", h4.keptForge, h4.keptOrigin)
	}
	_, vm5, _ := gitScene(t, m, "")
	vm5.GitForge.Set(0)
	_, vm6, h6 := gitSceneKept(t, m, h4.kept, h4.keptForge, h4.keptOrigin)
	if len(h6.tokens) != 0 || vm6.GitConnected.Get() || h6.kept != "" {
		t.Errorf("a GitLab token on a GitHub page: tokens %v, connected %v, kept %q", h6.tokens, vm6.GitConnected.Get(), h6.kept)
	}

	// the right forge with the wrong origin, a record that names no forge (a bare
	// token an earlier page kept), and one that names no origin: all dropped
	for name, rec := range map[string][3]string{
		"another origin": {planted, gh, "https://api.github.com.evil.example"},
		"no forge":       {planted, "", ghAPI},
		"no origin":      {planted, gh, ""},
		"unknown forge":  {planted, "bitbucket", ghAPI},
	} {
		m := store.NewMemory()
		_, vm, h := gitSceneKept(t, m, rec[0], rec[1], rec[2])
		// the stored choice must say Remember, or the token is dropped for that
		// reason alone
		vm.GitRemember.Set(true)
		_, vmr, hr := gitSceneKept(t, m, rec[0], rec[1], rec[2])
		_ = h
		_ = vm
		if len(hr.tokens) != 0 || vmr.GitConnected.Get() || hr.kept != "" {
			t.Errorf("%s: tokens %v, connected %v, kept %q", name, hr.tokens, vmr.GitConnected.Get(), hr.kept)
		}
	}

	// control: the same page, the same choice, a record that matches, connects
	m = store.NewMemory()
	_, vm7, _ := gitScene(t, m, "")
	vm7.GitRemember.Set(true)
	_, vm8, h8 := gitSceneKept(t, m, planted, gh, ghAPI)
	if !vm8.GitConnected.Get() || len(h8.tokens) != 1 || h8.forges[0] != gh {
		t.Errorf("a matching record must connect: tokens %v forges %v", h8.tokens, h8.forges)
	}
}

// The record the host writes: the forge and the origin travel with the token, and
// anything that is not a whole record (a bare token, damage, a future version)
// reads as nothing.
func TestKeptTokenRecord(t *testing.T) {
	s := EncodeKept("github", "https://api.github.com", planted)
	f, o, tok, ok := DecodeKept(s)
	if !ok || f != "github" || o != "https://api.github.com" || tok != planted {
		t.Fatalf("%q: %q %q %q %v", s, f, o, tok, ok)
	}
	for name, bad := range map[string]string{
		"a bare token":    planted,
		"empty":           "",
		"damaged":         `{"v":1,"forge":"github"`,
		"another version": `{"v":2,"forge":"github","origin":"https://api.github.com","token":"x"}`,
		"no token":        `{"v":1,"forge":"github","origin":"https://api.github.com"}`,
		"no forge":        `{"v":1,"origin":"https://api.github.com","token":"x"}`,
		"no origin":       `{"v":1,"forge":"github","token":"x"}`,
		"not an object":   `["github"]`,
	} {
		if _, _, _, ok := DecodeKept(bad); ok {
			t.Errorf("%s read as a record: %q", name, bad)
		}
	}
	if KeptTokenKey == LegacyTokenKey {
		t.Error("the record must not reuse the legacy key")
	}
}
