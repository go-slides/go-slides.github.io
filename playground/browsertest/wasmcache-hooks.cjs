// The playground's own part of landingkit's wasmcache browser tests
// (wasmcache-e2e.cjs --hooks this file): the work a reader has in the editor
// when a newer version is announced is the deck, typed on the canvas, and it
// must be the deck that is in the editor after the reload into the new
// build, restored from the draft the page saved before it reloaded.
"use strict";

const TYPED = " kept over the update";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const fs = require("fs");
const path = require("path");

// the run is for another language's playground (--page /fr/playground/)
const LANGUAGE_PAGE = process.argv.some((a) => /^\/[a-z]{2}\/playground\/$/.test(a));

module.exports = {
  viewport: { width: 1440, height: 900, deviceScaleFactor: 1 },

  // The video worker is fetched when a deck needs a poster, the math font when
  // a deck has a formula, the emoji pack when it has an emoji, never before: the
  // kit asserts none was requested nor stored until this runs (the deck is the
  // welcome example's: no video, no formula, no emoji), and cached after it.
  // The font and the pack are fetched by the render worker, whose script is always
  // /playground/worker.js: in another language's page (/fr/playground/) it is
  // controlled by the English page's service worker, which a visit to /fr/ alone
  // has not installed, so there the worker's fetches are not that scope's to cache.
  // Its deck is then the video's alone, and the kit's draft carries no formula
  // and no emoji into the new build (a render of one would fetch them from the
  // network, which the kit counts).
  useFiles: LANGUAGE_PAGE
    ? ["/playground/video-worker.wasm"]
    : ["/playground/video-worker.wasm", "/playground/fonts/stix-math.otf", "/playground/emoji/twemoji.pack"],
  async use(page) {
    const clip = fs.readFileSync(path.join(__dirname, "..", "testdata", "clip.mp4"));
    const r = (await page.evaluate(() => goSlidesDebug())).rects.editor;
    await page.evaluate((b64, x, y) => {
      const dt = new DataTransfer();
      dt.items.add(new File([Uint8Array.from(atob(b64), (c) => c.charCodeAt(0))], "clip.mp4", { type: "video/mp4" }));
      window.dispatchEvent(new DragEvent("drop", { dataTransfer: dt, clientX: x, clientY: y, bubbles: true, cancelable: true }));
    }, clip.toString("base64"), r[0] + r[2] / 2, r[1] + r[3] / 2);
    await page.waitForFunction(() => goSlidesDebug().media.split("|").includes("media/clip.mp4"), { timeout: 60000 });
    const extra = LANGUAGE_PAGE ? "" : "$E=mc^2$\n\n\u{1F604}\n\n";
    await page.evaluate((x) => goSlidesSetSource("---\nslides: true\n---\n\n# Clip\n\n" + x + "![clip](media/clip.mp4#t=0.3)\n"), extra);
  },

  async prepare(page) {
    const d = await page.evaluate(() => goSlidesDebug());
    const lines = d.source.split("\n");
    const hl = lines.findIndex((l) => l.startsWith("# "));
    const [x, y] = await page.evaluate((l, c) => goSlidesCaretPixel(l, c), hl, lines[hl].length);
    await page.mouse.click(x + 2, y + 6);
    await page.keyboard.type(TYPED);
    await page.waitForFunction((t) => goSlidesDebug().source.includes(t), { timeout: 10000 }, TYPED);
  },

  async verify(page) {
    // the editor is up (the kit waited for .stage.ready); the draft comes back
    // a moment later with the project
    for (let i = 0; i < 100; i++) {
      const d = await page.evaluate(() => globalThis.goSlidesDebug && goSlidesDebug());
      if (d && d.source.includes(TYPED)) return true; // the status line is in the page's language: the text is the witness
      await sleep(100);
    }
    return false;
  },
};
