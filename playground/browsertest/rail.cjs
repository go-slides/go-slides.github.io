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
        await shot(n);
        results.push(`${tag} ${label}: ${o.n} cards, worst ${o.worst.toFixed(2)} px`);
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
  await b.close();
  console.log("RAIL_OK\n" + results.join("\n"));
})();
