package main

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestServe(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.wasm"), []byte("\x00asm"), 0o644)
	portFile := filepath.Join(t.TempDir(), "port")
	got := make(chan net.Listener, 1)
	old := serve
	defer func() { serve = old }()
	serve = func(ln net.Listener, d string) error {
		got <- ln
		return old(ln, d)
	}
	c := RootCmd()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetArgs([]string{"--dir", dir, "--port-file", portFile})
	go c.Execute()
	ln := <-got
	defer ln.Close()
	b, _ := os.ReadFile(portFile)
	port, _ := strconv.Atoi(string(b))
	if ln.Addr().(*net.TCPAddr).Port != port || !ln.Addr().(*net.TCPAddr).IP.IsLoopback() {
		t.Fatal("not on the loopback port it wrote")
	}
	res, err := http.Get("http://127.0.0.1:" + string(b) + "/a.wasm")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if string(body) != "\x00asm" || res.Header.Get("Cache-Control") != "no-store" || res.Header.Get("Content-Type") != "application/wasm" {
		t.Fatalf("%q %v", body, res.Header)
	}
}

func TestServeErrors(t *testing.T) {
	c := RootCmd()
	c.SetArgs([]string{"--port-file", filepath.Join(t.TempDir(), "no", "such", "dir")})
	if err := c.Execute(); err == nil {
		t.Fatal("an unwritable port file passed")
	}
	old := osExit
	defer func() { osExit = old }()
	code := 0
	osExit = func(c int) { code = c }
	os.Args = []string{"serve", "--bogus"}
	main()
	if code != 1 {
		t.Fatal(code)
	}
}
