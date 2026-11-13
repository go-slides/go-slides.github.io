// The playground's own part of landingkit's wasmcache browser tests
// (wasmcache-e2e.cjs --hooks this file): the work a reader has in the editor
// when a newer version is announced is the deck, typed on the canvas, and it
// must be the deck that is in the editor after the reload into the new
// build, restored from the draft the page saved before it reloaded.
"use strict";

const TYPED = " kept over the update";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

module.exports = {
  viewport: { width: 1440, height: 900, deviceScaleFactor: 1 },

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
      if (d && d.source.includes(TYPED) && /Restored your draft/.test(d.status)) return true;
      await sleep(100);
    }
    return false;
  },
};
