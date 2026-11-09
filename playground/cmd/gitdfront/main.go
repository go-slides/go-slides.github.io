// Command gitdfront is the browser test's stand-in for Caddy and oauth2-proxy in
// front of slides-gitd (selfhost/Caddyfile.gitd): it serves a built site and
// mounts the REAL slides-gitd server code (github.com/go-slides/gitd/gitdtest,
// over an in-memory forge) at /api/git/, on one loopback origin.
//
// oauth2-proxy is faked by its trusted-header mode, as the gitd README describes:
// a request carries a "test-user" cookie (what oauth2-proxy's session cookie is),
// and the front turns it into X-Auth-Request-User after deleting every identity
// header the client sent, as the Caddyfile does. A request with no such cookie
// is a 401, as /oauth2/auth answers. /oauth2/sign_in signs in as the user named
// by the "test-intent" cookie (default alice) and sends the browser back to rd.
//
// Test only: it binds 127.0.0.1, its tokens are fixed strings, and /__test/* is
// the harness's window on what the fake forge received.
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-slides/gitd/gitdtest"
	"github.com/spf13/cobra"
)

var osExit = os.Exit

func main() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gitdfront:", err)
		osExit(1)
	}
}

// The two people, their repositories and their forge tokens. The decks differ
// so that a page showing the wrong one is seen at once.
const (
	Alice = "sub-alice"
	Bob   = "sub-bob"
	// AliceToken and BobToken are not shaped like real tokens.
	AliceToken = "ALICE-FORGE-TOKEN-PLACEHOLDER-0001"
	BobToken   = "BOB-FORGE-TOKEN-PLACEHOLDER-0002"
)

func options(origin, dir string) gitdtest.Options {
	return gitdtest.Options{
		Origin: origin, Dir: dir,
		Users: []gitdtest.User{
			{Subject: Alice, Slug: "alice", Repos: []string{"lab/alpha"}, Token: AliceToken},
			{Subject: Bob, Slug: "bob", Repos: []string{"lab/beta"}, Token: BobToken},
		},
		Repos: []gitdtest.Repo{
			{Name: "lab/alpha", Token: AliceToken, Files: map[string]string{"talk/deck.md": "---\ntheme: default\n---\n\n# Alpha deck of Alice\n"}},
			{Name: "lab/beta", Token: BobToken, Files: map[string]string{"talk/deck.md": "---\ntheme: default\n---\n\n# Beta deck of Bob\n"}},
		},
	}
}

// RootCmd returns the command.
func RootCmd() *cobra.Command {
	var site, portFile, data string
	cmd := &cobra.Command{
		Use:           "gitdfront",
		Short:         "Serve a built site with slides-gitd behind a fake oauth2-proxy, for the browser test",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				return err
			}
			port := ln.Addr().(*net.TCPAddr).Port
			h, err := Handler(fmt.Sprintf("http://127.0.0.1:%d", port), site, data)
			if err != nil {
				ln.Close()
				return err
			}
			if portFile != "" {
				if err := os.WriteFile(portFile, []byte(strconv.Itoa(port)), 0o644); err != nil {
					ln.Close()
					return err
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "serving %s on http://127.0.0.1:%d/ with slides-gitd\n", site, port)
			return http.Serve(ln, h)
		},
	}
	cmd.Flags().StringVar(&site, "dir", "public", "the built site to serve")
	cmd.Flags().StringVar(&data, "data", "", "a directory for the vault (an empty one)")
	cmd.Flags().StringVar(&portFile, "port-file", "", "write the chosen port here")
	return cmd
}

// Handler is the front: the site, the fake oauth2-proxy and slides-gitd.
func Handler(origin, site, data string) (http.Handler, error) {
	env, err := gitdtest.New(options(origin, data))
	if err != nil {
		return nil, err
	}
	known := map[string]bool{Alice: true, Bob: true}
	files := http.FileServer(http.Dir(site))
	mux := http.NewServeMux()
	mux.HandleFunc("/api/git/", func(w http.ResponseWriter, r *http.Request) {
		// what the Caddyfile does: the client never chooses its identity
		r.Header.Del("X-Auth-Request-User")
		r.Header.Del("X-Auth-Request-Email")
		r.Header.Del("X-Gitd-Proxy-Secret")
		c, err := r.Cookie("test-user")
		if err != nil || !known[c.Value] {
			// oauth2-proxy's /oauth2/auth, as forward_auth passes it on
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		r.Header.Set("X-Auth-Request-User", c.Value)
		env.Handler.ServeHTTP(w, r)
	})
	mux.HandleFunc("/oauth2/sign_in", func(w http.ResponseWriter, r *http.Request) {
		who := Alice
		if c, err := r.Cookie("test-intent"); err == nil && known[c.Value] {
			who = c.Value
		}
		http.SetCookie(w, &http.Cookie{Name: "test-user", Value: who, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
		rd := r.URL.Query().Get("rd")
		// oauth2-proxy only sends the browser back to a path on this site
		if u, err := url.Parse(rd); err != nil || !strings.HasPrefix(rd, "/") || strings.HasPrefix(rd, "//") || u.Host != "" {
			rd = "/"
		}
		http.Redirect(w, r, rd, http.StatusFound)
	})
	// the marker the real front serves (Caddyfile.gitd): this page is served by slides-gitd
	mux.HandleFunc("/.well-known/slides-gitd", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slides-gitd":1}`))
	})
	mux.HandleFunc("/__test/advance", func(w http.ResponseWriter, r *http.Request) {
		repo := r.URL.Query().Get("repo")
		ok := env.Advance(repo, map[string]string{"talk/deck.md": "---\ntheme: default\n---\n\n# " + r.URL.Query().Get("title") + "\n"})
		fmt.Fprintf(w, "%v", ok)
	})
	mux.HandleFunc("/__test/state", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"proposals": env.Proposals(), "audit": env.Audit()})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		files.ServeHTTP(w, r)
	})
	return mux, nil
}
