// rail.cjs <url> <shots-dir>: the thumbnail frame must put each slide exactly on
// the card the canvas draws for it (within 1 CSS pixel), at device pixel ratios
// 1, 1.5 and 2, in both colour schemes, at the start, after the strip is
// scrolled to slide 50, after the page is resized narrower, and after the
// Minimap and Slides toggles are pressed. The card rectangles come from the
// toolkit's layout (goSlidesDebug().thumbCards), the slides' from the frame's
// own document (puppeteer reads a cross-origin frame), placed by the frame
// element's geometry in the page.
const path = require("node:path");
const puppeteer = require("puppeteer-core");

const [url, shots] = process.argv.slice(2);
const chrome = process.env.CHROME || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

function deck(n) {
  let s = "---\nslides: true\ntheme: default\npaginate: true\n---\n\n";
  for (let i = 0; i < n; i++) {
    if (i) s += "\n---\n\n";
    s += `# Slide ${i + 1}\n\ntext on slide ${i + 1}\n`;
  }
  return s;
}

function fail(msg) {
  console.log("FAIL " + msg);
  process.exit(1);
}

// misalignment of the frame's slides against the strip's cards, in CSS pixels
async function offsets(page) {
  const d = await page.evaluate(() => goSlidesDebug());
  const el = await page.$("#slides-rail");
  const box = await el.boundingBox();
  const frame = await el.contentFrame();
  const svgs = await frame.evaluate(() =>
    [...document.querySelectorAll("#go-slides-preview>svg")].map((e) => {
      const r = e.getBoundingClientRect();
      return [r.left, r.top, r.width, r.height];
    }));
  const { first, rects } = d.thumbCards;
  let worst = 0;
  rects.forEach((c, i) => {
    const s = svgs[first + i];
    if (!s) return;
    const got = [box.x + s[0], box.y + s[1], s[2], s[3]];
    for (let k = 0; k < 4; k++) worst = Math.max(worst, Math.abs(got[k] - c[k]));
  });
  return { worst, n: rects.length, first };
}

// every visible card must show its slide: the pixels of each card differ when the thumbnail frame
// is hidden, over most of the card (an empty column shows the strip's own placeholder either way).
// Returns the smallest fraction of a card's pixels the frame changes.
async function filled(page) {
  const d = await page.evaluate(() => goSlidesDebug());
  const strip = d.rects.thumbs; // a card the strip's edge cuts in two is not judged
  const rects = (d.thumbCards ? d.thumbCards.rects : []).filter((r) =>
    r[0] >= strip[0] && r[1] >= strip[1] && r[0] + r[2] <= strip[0] + strip[2] && r[1] + r[3] <= strip[1] + strip[3]);
  const shotOf = () => page.screenshot({ encoding: "base64" });
  const withFrame = await shotOf();
  await page.$eval("#slides-rail", (e) => { e.style.visibility = "hidden"; });
  const without = await shotOf();
  await page.$eval("#slides-rail", (e) => { e.style.visibility = ""; });
  const dpr = await page.evaluate(() => devicePixelRatio);
  const fr = await page.evaluate(async (a, b, rects, dpr) => {
    const load = async (s) => {
      const img = await createImageBitmap(await (await fetch("data:image/png;base64," + s)).blob());
      const c = new OffscreenCanvas(img.width, img.height);
      c.getContext("2d").drawImage(img, 0, 0);
      return [c.getContext("2d"), img.width, img.height];
    };
    const [ga, w, h] = await load(a), [gb] = await load(b);
    return rects.map((r) => {
      const x = Math.round(r[0] * dpr) + 3, y = Math.round(r[1] * dpr) + 3, cw = Math.round(r[2] * dpr) - 6, ch = Math.round(r[3] * dpr) - 6;
      if (cw <= 0 || ch <= 0 || x < 0 || y < 0 || x + cw > w || y + ch > h) return 1; // clipped by the window: not judged
      const pa = ga.getImageData(x, y, cw, ch).data, pb = gb.getImageData(x, y, cw, ch).data;
      let diff = 0;
      for (let i = 0; i < pa.length; i += 4) if (Math.abs(pa[i] - pb[i]) + Math.abs(pa[i + 1] - pb[i + 1]) + Math.abs(pa[i + 2] - pb[i + 2]) > 24) diff++;
      return diff / (pa.length / 4);
    });
  }, withFrame, without, rects, dpr);
  return { n: fr.length, least: fr.length ? Math.min(...fr) : 0 };
}

async function showsSlides(page, label) {
  let f;
  for (let t = 0; t < 40; t++) { // the frame loads and is told after the canvas draws
    f = await filled(page);
    if (f.n > 0 && f.least >= 0.5) return f;
    await sleep(150);
  }
  fail(`${label}: ${f.n} cards, the emptiest shows the slide on only ${(f.least * 100).toFixed(1)}% of its pixels`);
}

async function aligned(page, label) {
  let o;
  for (let t = 0; t < 40; t++) { // the frame is told by a message: allow it a moment
    o = await offsets(page);
    if (o.n > 0 && o.worst <= 1) return o;
    await sleep(100);
  }
  fail(`${label}: slides off their cards by ${o.worst.toFixed(2)} px (${o.n} cards from slide ${o.first})`);
}

(async () => {
  const b = await puppeteer.launch({ executablePath: chrome, headless: "new", args: ["--no-sandbox"] });
  const results = [];
  for (const dpr of [1, 1.5, 2]) {
    for (const scheme of ["light", "dark"]) {
      const page = await b.newPage();
      page.on("pageerror", (e) => fail("pageerror: " + e.message));
      await page.setViewport({ width: 1500, height: 900, deviceScaleFactor: dpr });
      await page.emulateMediaFeatures([{ name: "prefers-color-scheme", value: scheme }]);
      await page.goto(url, { waitUntil: "networkidle0" });
      await page.waitForFunction("window.goSlidesReady===true", { timeout: 60000 });
      await page.evaluate((src) => goSlidesSetSource(src), deck(100));
      await page.waitForFunction(() => !goSlidesDebug().busy && goSlidesDebug().thumbCards.rects.length > 3 &&
        goSlidesDebug().source.includes("Slide 100"), { timeout: 30000 });
      const tag = `dpr${dpr}-${scheme}`;
      const shot = (n) => page.screenshot({ path: path.join(shots, `rail-${tag}-${n}.png`) });
      const step = async (label, n) => {
        const o = await aligned(page, `${tag} ${label}`);
        const f = await showsSlides(page, `${tag} ${label}`);
        await shot(n);
        results.push(`${tag} ${label}: ${o.n} cards, worst ${o.worst.toFixed(2)} px, emptiest card ${(f.least * 100).toFixed(0)}% slide`);
      };

      await step("initial", "1-initial");

      // A divider dragged: the thumbnail frame and the preview frame follow the
      // canvas while the button is still down, and again once it is released; the
      // width is kept over a reload. (The drag is undone, so that each page starts
      // from the same state: the browser context keeps what the page stores.)
      const css = (r) => [r[0], r[1], r[2], r[3]];
      const dragDivider = async (name, dx, label) => {
        const dd = (await page.evaluate(() => goSlidesDebug())).rects[name];
        if (!dd) fail(`${tag}: no ${name} in ${JSON.stringify(Object.keys((await page.evaluate(() => goSlidesDebug())).rects))}`);
        const cx = dd[0] + dd[2] / 2, cy = dd[1] + dd[3] / 2;
        await page.mouse.move(cx, cy);
        await page.mouse.down();
        await page.mouse.move(cx + dx / 2, cy, { steps: 4 });
        await framesFollow(`${tag} ${label} (mid-drag)`);
        await page.mouse.move(cx + dx, cy, { steps: 4 });
        await page.mouse.up();
        await framesFollow(`${tag} ${label} (released)`);
      };
      const framesFollow = async (label) => {
        // the frames are placed in the same draw as the canvas: no wait
        const got = await page.evaluate(() => {
          const d = goSlidesDebug();
          const box = (sel) => { const r = document.querySelector(sel).getBoundingClientRect(); return [r.left, r.top, r.width, r.height]; };
          return { rects: d.rects, preview: box("#slides-preview"), rail: box("#slides-rail"), railShown: getComputedStyle(document.querySelector("#slides-rail")).display };
        });
        const near = (a, b) => a.every((v, i) => Math.abs(v - b[i]) <= 1.01);
        // a frame is the canvas's rectangle: the preview and the strip
        if (!near(got.preview, css(got.rects.preview))) fail(`${label}: the preview frame ${got.preview} is not on the canvas's ${css(got.rects.preview)}`);
        if (got.railShown !== "none" && !near(got.rail, css(got.rects.thumbs))) fail(`${label}: the thumbnail frame ${got.rail} is not on the strip ${css(got.rects.thumbs)}`);
      };
      const thumbsW = async () => (await page.evaluate(() => goSlidesDebug())).rects.thumbs[2];
      const w0 = await thumbsW();
      await dragDivider("split-2", 50, "thumbnail divider +50");
      const w1 = await thumbsW();
      if (Math.abs(w1 - w0 - 50) > 1.5) fail(`${tag}: the thumbnail column went from ${w0} to ${w1}, want +50`);
      await step("thumbnail divider dragged", "8-dragged");
      await page.reload({ waitUntil: "networkidle0" });
      await page.waitForFunction("window.goSlidesReady===true", { timeout: 60000 });
      const w2 = await thumbsW();
      if (Math.abs(w2 - w1) > 1.5) fail(`${tag}: after a reload the thumbnail column is ${w2}, it was ${w1}`);
      await page.evaluate((src) => goSlidesSetSource(src), deck(100));
      await page.waitForFunction(() => !goSlidesDebug().busy && goSlidesDebug().thumbCards.rects.length > 3 &&
        goSlidesDebug().source.includes("Slide 100"), { timeout: 30000 });
      await dragDivider("split-2", -50, "thumbnail divider -50");
      const w3 = await thumbsW();
      if (Math.abs(w3 - w0) > 1.5) fail(`${tag}: the thumbnail column is ${w3} after dragging back, was ${w0}`);
      await step("thumbnail divider back", "9-dragged-back");

      // scroll the strip down to slide 50: the wheel over it, a few rows at a time
      const th = (await page.evaluate(() => goSlidesDebug())).rects.thumbs;
      await page.mouse.move(th[0] + th[2] / 2, th[1] + th[3] / 2);
      for (let i = 0; i < 400; i++) {
        await page.mouse.wheel({ deltaY: 240 });
        const f = (await page.evaluate(() => goSlidesDebug())).thumbCards.first;
        if (f >= 48) break;
        await sleep(20);
      }
      const f = (await page.evaluate(() => goSlidesDebug())).thumbCards.first;
      if (f < 40) fail(`${tag}: the strip did not scroll (first card ${f})`);
      await step("scrolled to slide " + f, "2-scrolled");

      // narrower
      await page.setViewport({ width: 1000, height: 700, deviceScaleFactor: dpr });
      await sleep(600);
      await step("narrower window", "3-narrow");
      await page.setViewport({ width: 1500, height: 900, deviceScaleFactor: dpr });
      await sleep(600);

      // the Minimap and Slides toggles
      let d = await page.evaluate(() => goSlidesDebug());
      const click = async (r) => page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2);
      // the toggles are on the ribbon's View tab
      await click(d.rects["tab-View"]);
      await page.waitForFunction(() => goSlidesDebug().rects.minimapToggle, { timeout: 10000 });
      d = await page.evaluate(() => goSlidesDebug());
      await click(d.rects.minimapToggle);
      await sleep(300);
      await step("minimap off", "4-nominimap");
      await click(d.rects.thumbsToggle);
      await sleep(400);
      const hidden = await page.$eval("#slides-rail", (e) => getComputedStyle(e).display);
      if (hidden !== "none") fail(`${tag}: the frame stays visible with Slides off (${hidden})`);
      await shot("5-nothumbs");
      d = await page.evaluate(() => goSlidesDebug());
      await click(d.rects.thumbsToggle);
      await sleep(400);
      await step("slides back on", "6-thumbs-again");
      await click(d.rects.minimapToggle);
      await sleep(300);
      await step("minimap back on", "7-minimap-again");
      await page.close();
    }
  }

  // Each language page, cold, and the slideshow: the draft keeps the preview mode, so a reload
  // after Present comes back to the slideshow, where the column once stayed reserved and empty.
  // A browser context of its own per condition, so nothing is stored between them.
  const root = url.replace(/\/(?:[a-z]{2}\/)?playground\/?$/, "/");
  for (const lang of ["", "fr/", "ja/", "uk/"]) {
    for (const width of [1280, 1920]) {
      const ctx = await b.createBrowserContext();
      const page = await ctx.newPage();
      page.on("pageerror", (e) => fail("pageerror: " + e.message));
      await page.setViewport({ width, height: 900, deviceScaleFactor: 1 });
      const tag = `${lang.replace("/", "") || "en"}-${width}`;
      const ready = async () => {
        await page.waitForFunction("window.goSlidesReady===true", { timeout: 180000 });
        await page.waitForFunction(() => !goSlidesDebug().busy && goSlidesDebug().thumbCards.rects.length > 0, { timeout: 30000 })
          .catch(async () => {
            await page.screenshot({ path: path.join(shots, `FAIL-empty-${tag}.png`) });
            fail(`${tag}: the thumbnail column holds no card: an empty panel (screenshot FAIL-empty-${tag}.png)`);
          });
      };
      const check = async (label, n) => {
        const o = await aligned(page, `${tag} ${label}`);
        const f = await showsSlides(page, `${tag} ${label}`);
        await page.screenshot({ path: path.join(shots, `rail-${tag}-${n}.png`) });
        results.push(`${tag} ${label}: ${o.n} cards, worst ${o.worst.toFixed(2)} px, emptiest card ${(f.least * 100).toFixed(0)}% slide`);
      };
      await page.goto(`${root}${lang}playground/`, { waitUntil: "load", timeout: 120000 });
      await ready();
      await check("cold", "a-cold");
      await page.reload({ waitUntil: "load", timeout: 120000 });
      await ready();
      await check("restored draft", "b-reload");
      if (width === 1280) { // the slideshow, then back to it from a reload
        const d = await page.evaluate(() => goSlidesDebug());
        const p = d.rects.present;
        await page.mouse.click(p[0] + p[2] / 2, p[1] + p[3] / 2);
        await page.waitForFunction(() => goSlidesDebug().mode === 1, { timeout: 20000 });
        await sleep(1500);
        await check("slideshow", "c-slideshow");
        await page.reload({ waitUntil: "load", timeout: 120000 });
        await ready();
        if ((await page.evaluate(() => goSlidesDebug())).mode !== 1) fail(`${tag}: the reload did not keep the slideshow`);
        await check("slideshow after a reload", "d-slideshow-reload");
      }
      await ctx.close();
    }
  }
  // The thumbnail column resized with its divider: the thumbnails follow it, to three widths. Each card
  // fills the column (no empty band on its right, nothing clipped), shows its slide, and sits on the slide
  // the frame scales to it; a screenshot per width goes to shots/resize-*.png.
  {
    const ctx = await b.createBrowserContext();
    const page = await ctx.newPage();
    page.on("pageerror", (e) => fail("pageerror: " + e.message));
    await page.setViewport({ width: 1500, height: 900, deviceScaleFactor: 1 });
    await page.goto(`${root}playground/`, { waitUntil: "load", timeout: 120000 });
    await page.waitForFunction("window.goSlidesReady===true", { timeout: 180000 });
    await page.evaluate((src) => goSlidesSetSource(src), deck(100));
    await page.waitForFunction(() => !goSlidesDebug().busy && goSlidesDebug().thumbCards.rects.length > 3 &&
      goSlidesDebug().source.includes("Slide 100"), { timeout: 30000 });
    const geo = async () => {
      const d = await page.evaluate(() => goSlidesDebug());
      return { strip: d.rects.thumbs, card: d.thumbCards.rects[0], d };
    };
    const widths = [];
    for (const dx of [-60, 120, 140]) { // from the default: narrower, then wider, then wider still
      const d = await page.evaluate(() => goSlidesDebug());
      const dd = d.rects["split-2"];
      const cx = dd[0] + dd[2] / 2, cy = dd[1] + dd[3] / 2;
      await page.mouse.move(cx, cy);
      await page.mouse.down();
      await page.mouse.move(cx + dx, cy, { steps: 6 });
      await page.mouse.up();
      await sleep(400);
      const { strip, card } = await geo();
      const label = `resize to ${Math.round(strip[2])} px`;
      // the card fills the column: its right edge sits within the number gutter's mirror of the strip's edge
      const free = strip[0] + strip[2] - (card[0] + card[2]);
      if (free > 40) fail(`${label}: ${free.toFixed(0)} px of empty column right of the cards (card ${card[2].toFixed(0)} wide in ${strip[2].toFixed(0)})`);
      if (card[0] + card[2] > strip[0] + strip[2] + 0.5) fail(`${label}: the cards are clipped by the column`);
      widths.push(card[2]);
      const o = await aligned(page, label);
      const f = await showsSlides(page, label);
      await page.screenshot({ path: path.join(shots, `resize-${Math.round(strip[2])}.png`) });
      results.push(`${label}: card ${card[2].toFixed(0)} px, ${o.n} cards, worst ${o.worst.toFixed(2)} px, emptiest card ${(f.least * 100).toFixed(0)}% slide`);
    }
    if (!(widths[1] > widths[0]) || !(widths[2] > widths[1])) fail(`the cards did not follow the column: ${widths.join(", ")}`);
    await ctx.close();
  }
  await b.close();
  console.log("RAIL_OK\n" + results.join("\n"));
})();
