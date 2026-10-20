"use strict";
// The playground's render worker. The page compiles playground.wasm once and
// posts the compiled WebAssembly.Module here as the first message; this
// script runs it with Go's wasm_exec.js. The Go program sees no document,
// knows it is the worker, and answers render requests from then on (see
// playground/cmd/playground-wasm). Nothing is fetched here: no second
// download, no second compilation.
importScripts("wasm_exec.js");
self.onmessage = function (e) {
  if (!e.data || !(e.data.module instanceof WebAssembly.Module)) return;
  self.onmessage = null; // the Go program installs its own handler
  var go = new Go();
  WebAssembly.instantiate(e.data.module, go.importObject).then(function (instance) {
    go.run(instance); // never resolves: the worker serves forever
  });
};
