package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func front(t *testing.T) *httptest.Server {
	t.Helper()
	site := t.TempDir()
	if err := os.WriteFile(filepath.Join(site, "index.html"), []byte("site"), 0o644); err != nil {
		t.Fatal(err)
	}
	var h http.Handler
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { h.ServeHTTP(w, r) }))
	t.Cleanup(srv.Close)
	var err error
	if h, err = Handler(srv.URL, site, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	return srv
}

func get(t *testing.T, srv *httptest.Server, path string, hdr map[string]string, cookie string) *http.Response {
	t.Helper()
	r, _ := http.NewRequest("GET", srv.URL+path, nil)
	for k, v := range hdr {
		r.Header.Set(k, v)
	}
	if cookie != "" {
		r.AddCookie(&http.Cookie{Name: "test-user", Value: cookie})
	}
	rsp, err := (&http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}).Do(r)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { rsp.Body.Close() })
	return rsp
}

func TestTheFrontIsWhatCaddyAndOauth2ProxyAre(t *testing.T) {
	srv := front(t)
	if r := get(t, srv, "/", nil, ""); r.StatusCode != 200 {
		t.Fatalf("site: %d", r.StatusCode)
	}
	if r := get(t, srv, "/api/git/session", nil, ""); r.StatusCode != 401 {
		t.Fatalf("no session cookie: %d", r.StatusCode)
	}
	// a client-chosen identity is deleted before the proxy's own is set
	if r := get(t, srv, "/api/git/session", map[string]string{"X-Auth-Request-User": Alice}, ""); r.StatusCode != 401 {
		t.Fatalf("a forged identity header was believed: %d", r.StatusCode)
	}
	if r := get(t, srv, "/api/git/session", map[string]string{"X-Auth-Request-User": Alice}, Bob); r.StatusCode != 200 {
		t.Fatalf("bob: %d", r.StatusCode)
	} else {
		var b bytes.Buffer
		_, _ = b.ReadFrom(r.Body)
		if !strings.Contains(b.String(), "lab/beta") || strings.Contains(b.String(), "lab/alpha") {
			t.Fatalf("bob with a forged alice header got %s", b.String())
		}
	}
	if r := get(t, srv, "/api/git/session", nil, "sub-mallory"); r.StatusCode != 401 {
		t.Fatalf("an unknown user: %d", r.StatusCode)
	}
	// sign-in sends the browser back to a path of this site only
	for rd, want := range map[string]string{"/playground/": "/playground/", "https://evil.example/": "/", "//evil.example": "/"} {
		r := get(t, srv, "/oauth2/sign_in?rd="+rd, nil, "")
		if r.StatusCode != 302 || r.Header.Get("Location") != want || !strings.Contains(r.Header.Get("Set-Cookie"), "test-user=sub-alice") {
			t.Errorf("rd %q: %d %q %q", rd, r.StatusCode, r.Header.Get("Location"), r.Header.Get("Set-Cookie"))
		}
	}
	if r := get(t, srv, "/__test/advance?repo=lab/alpha&title=X", nil, ""); r.StatusCode != 200 {
		t.Fatal(r.StatusCode)
	}
	if r := get(t, srv, "/__test/state", nil, ""); r.StatusCode != 200 {
		t.Fatal(r.StatusCode)
	}
}
