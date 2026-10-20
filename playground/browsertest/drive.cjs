// Drives the built playground in headless Chrome: it loads, the preview
// appears, typing updates it, a planted template error shows in the editor
// and the problems, the help follows the caret, the preview is sandboxed,
// and the slideshow presents. It prints the timings as one JSON line.
//
//   node drive.cjs http://127.0.0.1:PORT/playground/ SHOTS_DIR
//
// CHROME names the browser; NODE_PATH finds puppeteer-core.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const path = require("path");

const url = process.argv[2];
const shots = process.argv[3] || ".";
fs.mkdirSync(shots, { recursive: true });

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}

async function until(page, what, fn, ms = 30000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn);
    if (v) return v;
    if (Date.now() > end) {
      const d = await page.evaluate(() => (globalThis.goSlidesDebug ? goSlidesDebug() : null)).catch(() => null);
      if (d) d.source = d.source.split("\n").slice(0, 16).join("\n");
      fail("timed out waiting for " + what + "\n" + JSON.stringify(d, null, 1));
    }
    await new Promise((r) => setTimeout(r, 25));
  }
}

const debug = (page) => page.evaluate(() => goSlidesDebug());

(async () => {
  const browser = await puppeteer.launch({
    executablePath: process.env.CHROME,
    headless: true,
    args: ["--no-sandbox", "--window-size=1440,900"],
  });
  const page = await browser.newPage();
  page.on("pageerror", (e) => console.error("pageerror:", e.message));
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  const t0 = Date.now();
  await page.goto(url, { waitUntil: "load" });
  await until(page, "the editor", () => globalThis.webcanvasReady && globalThis.goSlidesReady);
  const readyMs = Date.now() - t0;
  await until(page, "the first preview", () => goSlidesDebug().loads >= 1);
  const firstPreviewMs = Date.now() - t0;
  const timings = await page.evaluate(() => goSlidesTimings);
  await page.screenshot({ path: path.join(shots, "1-loaded.png") });

  let d = await debug(page);
  if (!d.docHasCSP) fail("the preview document carries no CSP");
  const sandbox = await page.$eval("#slides-preview", (f) => f.getAttribute("sandbox"));
  if (sandbox !== "allow-scripts") fail("the preview iframe's sandbox is " + sandbox);
  if (!/slides? · rendered/.test(d.status)) fail("status: " + d.status);

  // Type at the end of the first heading line.
  const lines = d.source.split("\n");
  const hl = lines.findIndex((l) => l.startsWith("# Hello"));
  const [x, y] = await page.evaluate((l, c) => goSlidesCaretPixel(l, c), hl, lines[hl].length);
  await page.mouse.click(x + 2, y + 6);
  const loadsBefore = (await debug(page)).loads;
  const typed = Date.now();
  await page.keyboard.type(" from Chrome");
  await until(page, "the typed text in the source", () => goSlidesDebug().source.includes("Hello, {{ .name }}! from Chrome"));
  await until(page, "the preview after typing", () => !goSlidesDebug().busy && goSlidesDebug().loads > 0, 10000);
  const before = loadsBefore;
  await page.waitForFunction((b) => goSlidesDebug().loads > b, { timeout: 10000 }, before);
  const editMs = Date.now() - typed;
  const preview = page.frames().find((f) => f !== page.mainFrame());
  if (!preview) fail("no preview frame");
  await until(page, "the typed text in the preview", async () => true);
  const previewText = await preview.evaluate(() => document.body.innerText);
  if (!previewText.includes("from Chrome")) fail("the preview does not show the edit: " + previewText.slice(0, 200));
  d = await debug(page);
  const roundTripMs = d.lastRoundTripMs, iframeLoadMs = d.lastLoadMs;

  // Plant an error: a template that names data the deck does not have.
  await page.keyboard.type(" {{ .nope }}");
  await until(page, "the error in the problems", () => /nope/.test(goSlidesDebug().problems));
  await until(page, "the error highlighted in the editor", () => goSlidesDebug().highlights.length > 0);
  d = await debug(page);
  if (d.highlights[0][0] !== hl) fail("the highlight is on line " + d.highlights[0][0] + ", not " + hl);
  await page.screenshot({ path: path.join(shots, "2-error.png") });
  // Take it back out: the problem clears.
  for (let i = 0; i < " {{ .nope }}".length; i++) await page.keyboard.press("Backspace");
  await until(page, "the problem to clear", () => goSlidesDebug().problems === "" && goSlidesDebug().highlights.length === 0);

  // The help follows the caret: on "paginate" it explains the directive.
  const pl = lines.findIndex((l) => l.startsWith("paginate"));
  const [px, py] = await page.evaluate((l) => goSlidesCaretPixel(l, 3), pl);
  await page.mouse.click(px + 1, py + 6);
  await until(page, "the help on paginate", () => /page number/.test(goSlidesDebug().help));

  // Completion: typing a kind value offers the built-in kinds.
  // (Covered by the editor package's tests; here, the popup opens.)

  // The slideshow: pick it in the Preview drop-down.
  d = await debug(page);
  const m = d.rects.mode;
  await page.mouse.click(m[0] + m[2] / 2, m[1] + m[3] / 2);
  const pop = await until(page, "the mode list", () => goSlidesDebug().rects.popover);
  const loads = (await debug(page)).loads;
  await page.mouse.click(pop[0] + pop[2] / 2, pop[1] + (pop[3] * 3) / 4);
  await until(page, "slideshow mode", () => goSlidesDebug().mode === 1);
  await page.waitForFunction((b) => goSlidesDebug().loads > b, { timeout: 10000 }, loads);
  const show = page.frames().find((f) => f !== page.mainFrame());
  await new Promise((r) => setTimeout(r, 400));
  const hasViewer = await show.evaluate(() => document.querySelectorAll("svg[data-marpit-svg]").length);
  if (hasViewer < 4) fail("the slideshow shows " + hasViewer + " slides");
  // The keyboard moves to the next slide, inside the sandbox.
  const active = () => show.evaluate(() => [...document.querySelectorAll("svg[data-marpit-svg]")].findIndex((s) => s.classList.contains("active")));
  const a0 = await active();
  const fr = await page.$eval("#slides-preview", (f) => { const r = f.getBoundingClientRect(); return [r.left, r.top, r.width, r.height]; });
  await page.mouse.click(fr[0] + fr[2] / 2, fr[1] + fr[3] / 2);
  await page.keyboard.press("ArrowRight");
  await new Promise((r) => setTimeout(r, 400));
  const a1 = await active();
  if (!(a1 > a0)) fail("the slideshow did not advance: " + a0 + " → " + a1);
  await page.screenshot({ path: path.join(shots, "3-slideshow.png") });

  // The 100-slide benchmark deck, when given: from a change of the source
  // to the preview showing it, in the overview.
  const bench = {};
  if (process.env.BENCH_DECK) {
    const src = fs.readFileSync(process.env.BENCH_DECK, "utf8");
    d = await debug(page);
    await page.mouse.click(d.rects.mode[0] + d.rects.mode[2] / 2, d.rects.mode[1] + d.rects.mode[3] / 2);
    const p2 = await until(page, "the mode list", () => goSlidesDebug().rects.popover);
    await page.mouse.click(p2[0] + p2[2] / 2, p2[1] + p2[3] / 4);
    await until(page, "overview mode", () => goSlidesDebug().mode === 0);
    await page.waitForFunction(() => !goSlidesDebug().busy, { timeout: 30000 });
    let l0 = (await debug(page)).loads;
    let s = Date.now();
    await page.evaluate((t) => goSlidesSetSource(t), src);
    await page.waitForFunction((b) => goSlidesDebug().loads > b, { timeout: 30000 }, l0);
    bench.firstMs = Date.now() - s;
    const times = [], rtts = [];
    for (let i = 0; i < 9; i++) {
      l0 = (await debug(page)).loads;
      s = Date.now();
      await page.evaluate((t, i) => goSlidesSetSource(t + "\n\nedit " + i), src, i);
      await page.waitForFunction((b) => goSlidesDebug().loads > b, { timeout: 30000 }, l0);
      times.push(Date.now() - s);
      rtts.push((await debug(page)).lastRoundTripMs);
    }
    times.sort((a, b) => a - b);
    rtts.sort((a, b) => a - b);
    bench.editToPreviewMedianMs = times[4];
    bench.workerRoundTripMedianMs = rtts[4];
    bench.status = (await debug(page)).status;
  }

  console.log("PLAYGROUND_OK " + JSON.stringify({ readyMs, firstPreviewMs, downloadMs: Math.round(timings.downloadMs), editToPreviewMs: editMs, roundTripMs, iframeLoadMs, bench }));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
