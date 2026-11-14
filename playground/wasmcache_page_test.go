package playground

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readRepo(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.FromSlash("../" + rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The offline cache is landingkit's wasmcache: the page loads its script and
// starts it once the editor is up, guarded (a site built without the
// `landingkit wasmcache` step has no script), saving the project first; the
// wasm exposes that save; and the deploy's step covers every language's
// playground, lists no /api/ path, and keeps the lazily used export engine and
// fonts out of the first download.
func TestPlaygroundOfflineCache(t *testing.T) {
	page := readRepo(t, "layouts/_default/playground.html")
	for _, want := range []string{
		`<script src="/playground/wasmcache-client.js"></script>`,
		`if(globalThis.wasmcache) wasmcache.start({`,
		`goSlidesSaveNow()`,
		`beforeReload:`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("playground.html lacks %s", want)
		}
	}
	// the cache widens nothing the page can reach: no directive names the worker,
	// and connect-src is still the page's own origin
	if !strings.Contains(page, "worker-src 'self' blob:; object-src 'none'; base-uri 'none'; form-action 'none'; connect-src 'self'; frame-src 'self' blob: {{ $fo }}") {
		t.Error("the policy changed")
	}
	if !strings.Contains(readRepo(t, "playground/cmd/playground-wasm/store.go"), `Set("goSlidesSaveNow"`) {
		t.Error("the wasm does not expose goSlidesSaveNow")
	}
	action := readRepo(t, ".github/actions/wasmcache/action.yml")
	langs := []string{""} // English, at the root
	cfg, err := filepath.Glob("../config/_default/params.*.toml")
	if err != nil || len(cfg) == 0 {
		t.Fatalf("no language configs: %v", err)
	}
	for _, c := range cfg {
		langs = append(langs, strings.Split(filepath.Base(c), ".")[1]+"/")
	}
	for _, l := range langs {
		if !strings.Contains(action, "-scope /"+l+"playground/") {
			t.Errorf("no worker scope for /%splayground/", l)
		}
	}
	for _, want := range []string{"-exclude /api/", "-lazy playground/export.wasm", "-lazy playground/video-worker.wasm","-lazy playground/fonts/", "-asset playground/playground.wasm"} {
		if !strings.Contains(action, want) {
			t.Errorf("the wasmcache step lacks %s", want)
		}
	}
	if strings.Contains(action, "-asset /api") || strings.Contains(action, "-asset api") {
		t.Error("an /api/ path is listed for caching")
	}
	for _, wf := range []string{"deploy-pages.yml", "playground.yml"} {
		if !strings.Contains(readRepo(t, ".github/workflows/"+wf), "./.github/actions/wasmcache") && !strings.Contains(readRepo(t, ".github/workflows/"+wf), "/.github/actions/wasmcache") {
			t.Errorf("%s does not run the wasmcache step", wf)
		}
	}
}
