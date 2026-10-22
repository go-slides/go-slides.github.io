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
  // ready, and shown: on a slow first load the loading carousel fades out first
  await until(page, "the editor", () => globalThis.webcanvasReady && globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
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
  // The editor scrolled sideways to follow the caret at the end of the heading
  // (its text area is narrower since the minimap): Home brings the left edge back.
  await page.keyboard.press("Home");
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

  // Present: the presenter view opens in its own window, synced with the
  // preview both ways through this page, with the notes and the timer.
  d = await debug(page);
  const pr = d.rects.present;
  const popupTarget = browser.waitForTarget((t) => t.opener() && t.opener() === page.target(), { timeout: 10000 });
  await page.mouse.click(pr[0] + pr[2] / 2, pr[1] + pr[3] / 2);
  const popup = await (await popupTarget).page();
  await until(page, "the presenter page", () => goSlidesDebug().presenterOpen && goSlidesDebug().presenterDocLength > 0);
  await popup.setViewport({ width: 1100, height: 680 });
  let pv;
  for (let i = 0; i < 100 && !pv; i++) {
    pv = popup.frames().find((f) => f !== popup.mainFrame());
    if (pv) {
      const ok = await pv.evaluate(() => !!document.querySelector(".slides-pv")).catch(() => false);
      if (!ok) pv = undefined;
    }
    if (!pv) await new Promise((r) => setTimeout(r, 100));
  }
  if (!pv) fail("no presenter view in the presenter window");
  const pvState = () => pv.evaluate(() => ({
    active: [...document.querySelectorAll("div.slides-pv > div.marpit > svg[data-marpit-svg]")].findIndex((s) => s.classList.contains("active")),
    page: (document.querySelector(".slides-pv-page") || {}).textContent,
    timer: (document.querySelector(".slides-pv-timer") || {}).textContent,
    note: [...document.querySelectorAll(".slides-note.active")].map((n) => n.textContent).join(" "),
    next: document.querySelectorAll(".slides-pv-next svg[data-marpit-svg]").length,
  }));
  // It opens on the slide the preview shows.
  let s0 = await until(page, "the presenter on the preview's slide", async () => true);
  for (let i = 0; i < 50; i++) { s0 = await pvState(); if (s0.active === a1) break; await new Promise((r) => setTimeout(r, 100)); }
  if (s0.active !== a1) fail("the presenter opened on slide " + s0.active + ", the preview shows " + a1);
  if (!/\d\d:\d\d:\d\d/.test(s0.timer) || s0.next !== 1) fail("presenter view: " + JSON.stringify(s0));
  // Back one slide from the presenter window: both go.
  await popup.bringToFront();
  await popup.mouse.click(300, 300);
  await popup.keyboard.press("ArrowLeft");
  let s1;
  for (let i = 0; i < 50; i++) { s1 = await pvState(); if (s1.active === a1 - 1 && (await active()) === a1 - 1) break; await new Promise((r) => setTimeout(r, 100)); }
  if (s1.active !== a1 - 1 || (await active()) !== a1 - 1) fail("presenter → audience sync: presenter " + s1.active + ", audience " + (await active()));
  if (!/speaker note/.test(s1.note)) fail("no speaker note on slide " + (a1 - 1) + ": " + JSON.stringify(s1));
  await popup.screenshot({ path: path.join(shots, "4-presenter.png") });
  // And forward from the audience: the presenter follows.
  await page.bringToFront();
  await page.mouse.click(fr[0] + fr[2] * 0.8, fr[1] + fr[3] / 2);
  let s2;
  for (let i = 0; i < 50; i++) { s2 = await pvState(); if (s2.active === a1) break; await new Promise((r) => setTimeout(r, 100)); }
  if (s2.active !== a1) fail("audience → presenter sync: presenter " + s2.active);
  // An edit re-renders both and keeps the slide.
  const lb = (await debug(page)).loads;
  await page.evaluate(() => goSlidesSetSource(goSlidesDebug().source + "\n\nedited while presenting"));
  await page.waitForFunction((b) => goSlidesDebug().loads > b, { timeout: 10000 }, lb);
  await new Promise((r) => setTimeout(r, 600));
  const pv2 = popup.frames().find((f) => f !== popup.mainFrame());
  pv = pv2;
  const s3 = await pvState();
  if ((await active()) !== a1 || s3.active !== a1) fail("an edit lost the slide: audience " + (await active()) + ", presenter " + s3.active);
  const presenter = { relayed: (await debug(page)).relayed, notes: s1.note.trim() };
  await popup.close();
  await until(page, "the presenter window closed", () => !goSlidesDebug().presenterOpen);

  // Math, images and backgrounds: the example deck, slide by slide in the
  // sandboxed slideshow. Every <img> decodes, every background figure has
  // its data: image, the math is typeset; and a screenshot of each.
  const media = {};
  {
    const src = fs.readFileSync(path.join(__dirname, "..", "examples", "6-media.md"), "utf8");
    const l0 = (await debug(page)).loads;
    await page.evaluate((t) => goSlidesSetSource(t), src);
    await page.waitForFunction((b) => goSlidesDebug().loads > b && !goSlidesDebug().busy, { timeout: 10000 }, l0);
    await new Promise((r) => setTimeout(r, 600));
    const sf = page.frames().find((f) => f !== page.mainFrame());
    const n = await sf.evaluate(() => document.querySelectorAll("svg[data-marpit-svg]").length);
    if (n !== 5) fail("the media example has " + n + " slides");
    await page.mouse.click(fr[0] + fr[2] / 2, fr[1] + fr[3] / 2);
    await page.keyboard.press("Home");
    media.slides = [];
    for (let i = 0; i < n; i++) {
      if (i > 0) await page.keyboard.press("ArrowRight");
      await new Promise((r) => setTimeout(r, 500));
      const st = await sf.evaluate(async () => {
        const s = document.querySelector("svg[data-marpit-svg].active");
        const imgs = [...s.querySelectorAll("img")];
        await Promise.all(imgs.map((im) => (im.decode ? im.decode().catch(() => null) : null)));
        const bgs = [...s.querySelectorAll("figure")].map((f) => getComputedStyle(f).backgroundImage).filter((b) => b && b !== "none");
        // a background image decodes too: load each one's URL
        const bgOK = await Promise.all(bgs.map((b) => new Promise((res) => {
          const u = b.replace(/^url\("?/, "").replace(/"?\)$/, "");
          const im = new Image(); im.onload = () => res(im.naturalWidth > 0); im.onerror = () => res(false); im.src = u;
        })));
        return {
          imgs: imgs.length, imgsOK: imgs.filter((im) => im.complete && im.naturalWidth > 0).length,
          filter: imgs.map((im) => getComputedStyle(im).filter).filter((f) => f !== "none").length,
          bgs: bgs.length, bgsOK: bgOK.filter(Boolean).length,
          math: s.querySelectorAll("mjx-container, .math, svg svg, math").length,
          text: s.textContent.replace(/\s+/g, " ").trim().slice(0, 60),
        };
      });
      media.slides.push(st);
      await page.screenshot({ path: path.join(shots, "5-media-" + (i + 1) + ".png") });
    }
    const s = media.slides;
    if (s[0].bgs !== 1 || s[0].bgsOK !== 1) fail("slide 1, full background: " + JSON.stringify(s[0]));
    if (s[1].math < 2 || /\$|\\sqrt/.test(s[1].text)) fail("slide 2, math: " + JSON.stringify(s[1]));
    if (s[2].imgs !== 2 || s[2].imgsOK !== 2 || s[2].filter !== 1) fail("slide 3, images: " + JSON.stringify(s[2]));
    if (s[3].bgs !== 1 || s[3].bgsOK !== 1) fail("slide 4, split background: " + JSON.stringify(s[3]));
    if (s[4].bgs !== 2 || s[4].bgsOK !== 2) fail("slide 5, two backgrounds: " + JSON.stringify(s[4]));

    // A PNG file dropped on the page lands in the deck as a data: URI.
    const before = (await debug(page)).source.length;
    await page.evaluate(() => {
      const b64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==";
      const bytes = Uint8Array.from(atob(b64), (c) => c.charCodeAt(0));
      const dt = new DataTransfer();
      dt.items.add(new File([bytes], "dot.png", { type: "image/png" }));
      window.dispatchEvent(new DragEvent("drop", { dataTransfer: dt, bubbles: true, cancelable: true }));
    });
    await until(page, "the dropped image in the source", () => goSlidesDebug().source.includes("![dot](data:image/png;base64,iVBORw0KGgo"));
    media.dropped = (await debug(page)).source.length - before;
  }

  // The theme picker, in the overview: choosing gaia writes theme: gaia
  // into the deck; editing the deck by hand moves the picker.
  {
    d = await debug(page);
    await page.mouse.click(d.rects.mode[0] + d.rects.mode[2] / 2, d.rects.mode[1] + d.rects.mode[3] / 2);
    const pm = await until(page, "the mode list", () => goSlidesDebug().rects.popover);
    await page.mouse.click(pm[0] + pm[2] / 2, pm[1] + pm[3] / 4);
    await until(page, "overview mode", () => goSlidesDebug().mode === 0);
    await page.evaluate((t) => goSlidesSetSource(t), "---\nslides: true\n---\n\n# Themes\n");
    await until(page, "the default theme", () => goSlidesDebug().theme === "default");
    const th = (await debug(page)).rects.theme;
    await page.mouse.click(th[0] + th[2] / 2, th[1] + th[3] / 2);
    const tp = await until(page, "the theme list", () => goSlidesDebug().rects.popover);
    await page.mouse.click(tp[0] + tp[2] / 2, tp[1] + tp[3] / 2); // gaia, the middle of three
    await until(page, "theme: gaia in the deck", () => /\ntheme: gaia\n/.test(goSlidesDebug().source) && goSlidesDebug().theme === "gaia");
    await page.evaluate(() => goSlidesSetSource(goSlidesDebug().source.replace("theme: gaia", "theme: uncover")));
    await until(page, "the picker following the deck", () => goSlidesDebug().theme === "uncover");
    await page.screenshot({ path: path.join(shots, "6-theme.png") });
  }

  // The preview keeps its place: scrolled to slide 30 of the 100-slide
  // deck, an edit in slide 30 (patched in) and a change of theme (a new
  // document) both leave slide 30 at the top.
  const scroll = {};
  if (process.env.BENCH_DECK) {
    const src = fs.readFileSync(process.env.BENCH_DECK, "utf8");
    let l0 = (await debug(page)).loads;
    await page.evaluate((t) => goSlidesSetSource(t), src);
    await page.waitForFunction((b) => goSlidesDebug().loads > b && !goSlidesDebug().busy, { timeout: 30000 }, l0);
    await new Promise((r) => setTimeout(r, 500));
    const of = () => page.frames().find((f) => f !== page.mainFrame());
    const topSlide = () => of().evaluate(() => {
      const k = document.getElementById("go-slides-preview").children;
      for (let i = 0; i < k.length; i++) if (k[i].getBoundingClientRect().bottom > 1) return i;
      return -1;
    });
    await of().evaluate(() => document.getElementById("go-slides-preview").children[29].scrollIntoView({ block: "start" }));
    await new Promise((r) => setTimeout(r, 300));
    scroll.before = await topSlide();
    if (scroll.before !== 29) fail("could not scroll to slide 30: " + scroll.before);
    // one character in slide 30: the first heading after its separator
    const p0 = (await debug(page)).patches;
    l0 = (await debug(page)).loads;
    scroll.edited = await page.evaluate(() => {
      const lines = goSlidesDebug().source.split("\n");
      let seps = 0;
      for (let i = 0; i < lines.length; i++) {
        if (lines[i] === "---") seps++;
        // two lines of the front matter, then a separator before each slide after the first
        if (seps === 2 + 29 && /^#+ /.test(lines[i])) { lines[i] += "!"; goSlidesSetSource(lines.join("\n")); return lines[i]; }
      }
      return "";
    });
    await page.waitForFunction((b) => goSlidesDebug().loads > b && !goSlidesDebug().busy, { timeout: 10000 }, l0);
    await new Promise((r) => setTimeout(r, 300));
    scroll.afterEdit = await topSlide();
    scroll.patched = (await debug(page)).patches > p0;
    if (!scroll.edited || scroll.afterEdit !== 29 || !scroll.patched) fail("an edit in slide 30 moved the preview: " + JSON.stringify(scroll));
    l0 = (await debug(page)).loads;
    await page.evaluate(() => goSlidesSetSource(goSlidesDebug().source.replace(/^---\n/, "---\ntheme: gaia\n")));
    await page.waitForFunction((b) => goSlidesDebug().loads > b && !goSlidesDebug().busy, { timeout: 10000 }, l0);
    await new Promise((r) => setTimeout(r, 500));
    scroll.afterTheme = await topSlide();
    if (scroll.afterTheme !== 29) fail("a new theme moved the preview: " + JSON.stringify(scroll));
    await page.screenshot({ path: path.join(shots, "7-slide30.png") });
  }

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

  console.log("PLAYGROUND_OK " + JSON.stringify({ readyMs, firstPreviewMs, downloadMs: Math.round(timings.downloadMs), editToPreviewMs: editMs, roundTripMs, iframeLoadMs, presenter, scroll, media, bench }));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
