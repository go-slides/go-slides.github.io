// Drives the Insert tab's pictures in headless Chrome: Image, Background and the
// left split background each open the browser's file picker, and the file they
// get lands in the project's media/ (the explorer lists it) with a reference to
// it in the deck, never a data: URI; an image dropped on the editor does the
// same. Prints one line, INSERT_OK.
//
//   node insert.cjs http://127.0.0.1:PORT/playground/
//
// CHROME names the browser; NODE_PATH finds puppeteer-core.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const os = require("os");
const path = require("path");

const url = process.argv[2];

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}

async function until(page, what, fn, arg, ms = 30000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg).catch(() => null);
    if (v) return v;
    if (Date.now() > end) fail("timed out waiting for " + what + ": " + JSON.stringify(await page.evaluate(() => goSlidesDebug()).then((d) => ({ status: d.status, media: d.media, source: d.source.slice(0, 300) })).catch(() => null)));
    await new Promise((r) => setTimeout(r, 40));
  }
}

const debug = (page) => page.evaluate(() => goSlidesDebug());
const center = (r) => [r[0] + r[2] / 2, r[1] + r[3] / 2];

// a 1×1 PNG
const PNG = Buffer.from("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==", "base64");

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1500,900"] });
  const page = await browser.newPage();
  await page.setViewport({ width: 1500, height: 900, deviceScaleFactor: 1 });
  page.on("pageerror", (e) => console.error("pageerror:", e.message));
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "insert-"));
  const file = (name) => { const p = path.join(dir, name); fs.writeFileSync(p, PNG); return p; };
  await page.goto(url, { waitUntil: "load" });
  await until(page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready") && !goSlidesDebug().busy);
  await page.evaluate(() => goSlidesSetSource("---\nslides: true\n---\n\n# Pictures\n\ntext\n"));
  await until(page, "the deck", () => /# Pictures/.test(goSlidesDebug().source) && !goSlidesDebug().busy);

  // the picker: click the control, accept the file the browser asks for
  async function pickWith(click, name) {
    const [chooser] = await Promise.all([page.waitForFileChooser({ timeout: 10000 }), click()]);
    await chooser.accept([file(name)]);
  }
  const tab = async (t, key) => {
    const r = (await debug(page)).rects["tab-" + t];
    await page.mouse.click(...center(r));
    await page.waitForFunction((k) => goSlidesDebug().rects[k], { timeout: 10000 }, key);
    return (await debug(page)).rects[key];
  };
  const noData = async (what) => {
    const d = await debug(page);
    if (/data:/.test(d.source)) fail(what + " put a data: URI in the deck:\n" + d.source.slice(0, 400));
    return d;
  };

  // Insert ▸ Image
  let r = await tab("Insert", "image");
  await pickWith(() => page.mouse.click(...center(r)), "photo.png");
  await until(page, "Image in media/", () => /!\[photo\]\(media\/photo\.png\)/.test(goSlidesDebug().source));
  let d = await noData("Image");
  if (!d.media.split("|").includes("media/photo.png")) fail("the explorer does not hold media/photo.png: " + d.media);

  // Insert ▸ Background
  r = (await debug(page)).rects.background;
  await pickWith(() => page.mouse.click(r[0] + r[2] / 2, r[1] + 8), "back.png");
  await until(page, "Background in media/", () => /!\[bg\]\(media\/back\.png\)/.test(goSlidesDebug().source));
  await noData("Background");

  // Insert ▸ Background ▸ Left split background (the arrow half opens its menu)
  r = (await debug(page)).rects.background;
  await page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] - 8);
  const pop = await until(page, "the Background menu", () => goSlidesDebug().rects.popover);
  await pickWith(() => page.mouse.click(pop[0] + pop[2] / 2, pop[1] + pop[3] / 4), "left.png");
  await until(page, "the left split in media/", () => /!\[bg left:40%\]\(media\/left\.png\)/.test(goSlidesDebug().source));
  await noData("Left split background");

  // an image dropped on the editor goes the same way
  const ed = (await debug(page)).rects.editor;
  const [x, y] = center(ed);
  await page.evaluate((png, x, y) => {
    const dt = new DataTransfer();
    dt.items.add(new File([Uint8Array.from(atob(png), (c) => c.charCodeAt(0))], "dropped.png", { type: "image/png" }));
    window.dispatchEvent(new DragEvent("drop", { dataTransfer: dt, clientX: x, clientY: y, bubbles: true, cancelable: true }));
  }, PNG.toString("base64"), x, y);
  await until(page, "the dropped image in media/", () => /\(media\/dropped\.png\)/.test(goSlidesDebug().source));
  d = await noData("a drop");
  for (const m of ["photo", "back", "left", "dropped"]) if (!d.media.split("|").includes("media/" + m + ".png")) fail("media/" + m + ".png is not in the explorer: " + d.media);

  console.log("INSERT_OK " + JSON.stringify({ media: d.media }));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
