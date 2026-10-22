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
