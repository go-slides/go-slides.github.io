// The video worker in headless Chrome: video support is a module of its own
// (video-worker.wasm), loaded the first time a deck needs a poster.
//
//  1. a deck without video never requests video-worker.wasm;
//  2. a video dropped into the project and named by the deck loads it once
//     (same origin, so the page's connect-src 'self' holds), and the preview's
//     <video> has its poster (a JPEG data: URI decoded in the worker);
//  3. a file that is not a video, and one made of boxes past the limit, give
//     video.syntax and video.limit, and the page stays alive;
//  4. the exports still carry what they carried: the PPTX and the ODP hold the
//     video and its poster, the PDF the poster picture.
//
//   node videoworker.cjs http://127.0.0.1:PORT/playground/ DOWNLOAD_DIR /path/to/clip.mp4
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. Prints VIDEOWORKER_OK.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");

const url = process.argv[2];
const dir = path.resolve(process.argv[3] || ".");
const clip = fs.readFileSync(process.argv[4]);
fs.mkdirSync(dir, { recursive: true });

const fail = (m) => { console.error("FAIL " + m); process.exit(1); };
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function until(page, what, fn, arg, ms = 120000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg).catch(() => null);
    if (v) return v;
    if (Date.now() > end) fail("timed out waiting for " + what + ": " + JSON.stringify(await page.evaluate(() => { const d = goSlidesDebug(); return { status: d.status, problems: d.problems, source: d.source.slice(0, 200) }; }).catch(() => "?")));
    await sleep(50);
  }
}

// a file whose boxes are past the limit: an ftyp, then 300,000 empty boxes
function boxBomb() {
  const ftyp = Buffer.alloc(16);
  ftyp.writeUInt32BE(16, 0);
  ftyp.write("ftypisom", 4, "latin1");
  const empty = Buffer.alloc(8);
  empty.writeUInt32BE(8, 0);
  empty.write("free", 4, "latin1");
  return Buffer.concat([ftyp, ...Array(300000).fill(empty)]);
}
const notVideo = Buffer.concat([Buffer.from([0, 0, 0, 24]), Buffer.from("ftypisom"), Buffer.alloc(64, 7)]);

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  const errors = [], videoRequests = [], violations = [];
  page.on("pageerror", (e) => errors.push(e.message));
  page.on("console", (m) => { if (/Content Security Policy|Refused to/i.test(m.text())) violations.push(m.text()); });
  page.on("request", (r) => { if (r.url().includes("video-worker.wasm")) videoRequests.push(r.url()); });
  const cdp = await page.createCDPSession();
  await cdp.send("Browser.setDownloadBehavior", { behavior: "allow", downloadPath: dir, eventsEnabled: true });
  const done = new Set();
  cdp.on("Browser.downloadProgress", (e) => { if (e.state === "completed") done.add(e.guid); });
  await page.goto(url, { waitUntil: "load" });
  await until(page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready") && !goSlidesDebug().busy);
  const debug = () => page.evaluate(() => goSlidesDebug());

  // 1. no video, no video worker
  await page.evaluate(() => goSlidesSetSource("---\nslides: true\n---\n\n# No video\n\ntext\n\n---\n\n# Two\n"));
  await until(page, "the deck without video", () => /# No video/.test(goSlidesDebug().source) && !goSlidesDebug().busy && /2 slides/.test(goSlidesDebug().status));
  await sleep(1500);
  if (videoRequests.length) fail("a deck without video fetched " + videoRequests.join(" "));
  // the offline cache (landingkit's wasmcache), when the site has it: the
  // video worker is a lazy file, so the first install did not fetch or store it
  const cacheHolds = () => page.evaluate(async () => {
    for (const k of await caches.keys()) {
      for (const r of await (await caches.open(k)).keys()) if (new URL(r.url).pathname.endsWith("/video-worker.wasm")) return true;
    }
    return false;
  });
  const hasCache = await page.evaluate(() => !!globalThis.wasmcache && wasmcache.cached());
  if (hasCache && (await cacheHolds())) fail("the video worker was precached for a deck without video");

  // a medium dropped on the editor lands in media/ (insert.cjs)
  const drop = async (name, bytes, type) => {
    const r = (await debug()).rects.editor;
    await page.evaluate((name, b64, type, x, y) => {
      const dt = new DataTransfer();
      dt.items.add(new File([Uint8Array.from(atob(b64), (c) => c.charCodeAt(0))], name, { type }));
      window.dispatchEvent(new DragEvent("drop", { dataTransfer: dt, clientX: x, clientY: y, bubbles: true, cancelable: true }));
    }, name, bytes.toString("base64"), type, r[0] + r[2] / 2, r[1] + r[3] / 2);
    await until(page, name + " in media/", (n) => goSlidesDebug().media.split("|").includes("media/" + n), name);
  };

  // 2. a video: loaded once, the poster arrives
  await drop("clip.mp4", clip, "video/mp4");
  await page.evaluate(() => goSlidesSetSource("---\nslides: true\n---\n\n# Clip\n\n![clip](media/clip.mp4#t=0.3)\n"));
  const poster = () => page.frames().filter((f) => f !== page.mainFrame()).reduce(async (acc, f) => (await acc) || f.evaluate(() => { const v = document.querySelector("video[data-slides-video], video"); return v && v.getAttribute("poster") && v.getAttribute("poster").slice(0, 23); }).catch(() => null), Promise.resolve(null));
  const end = Date.now() + 120000;
  while ((await poster()) !== "data:image/jpeg;base64,") {
    if (Date.now() > end) fail("no poster in the preview: " + JSON.stringify(await debug().then((d) => ({ status: d.status, problems: d.problems }))));
    await sleep(100);
  }
  if (videoRequests.length !== 1) fail("video-worker.wasm was requested " + videoRequests.length + " times: " + videoRequests.join(" "));
  if (!new URL(videoRequests[0]).pathname.endsWith("/playground/video-worker.wasm")) fail("video worker from " + videoRequests[0]);
  if (hasCache) {
    // after its first use it is in the cache, for the next visit
    const t = Date.now() + 30000;
    while (!(await cacheHolds())) {
      if (Date.now() > t) fail("the video worker was not cached after its first use");
      await sleep(200);
    }
  }
  let d = await debug();
  if (/video\./.test(JSON.stringify(d.problems))) fail("a good video has problems: " + JSON.stringify(d.problems));
  // typing again does not load it again
  await page.evaluate(() => goSlidesSetSource("---\nslides: true\n---\n\n# Clip!\n\n![clip](media/clip.mp4#t=0.3)\n"));
  await until(page, "the second render", () => /Clip!/.test(goSlidesDebug().source) && !goSlidesDebug().busy);
  await sleep(500);
  if (videoRequests.length !== 1) fail("a keystroke fetched the video worker again");

  // 3. refusals, with their codes
  await drop("notvideo.mp4", notVideo, "video/mp4");
  await page.evaluate(() => goSlidesSetSource("---\nslides: true\n---\n\n# Bad\n\n![bad](media/notvideo.mp4)\n"));
  await until(page, "video.syntax", () => /video\.syntax/.test(JSON.stringify(goSlidesDebug().problems)));
  await drop("bomb.mp4", boxBomb(), "video/mp4");
  await page.evaluate(() => goSlidesSetSource("---\nslides: true\n---\n\n# Bomb\n\n![bomb](media/bomb.mp4)\n"));
  await until(page, "video.limit", () => /video\.limit/.test(JSON.stringify(goSlidesDebug().problems)));
  // and the page is alive: a good deck renders again
  await page.evaluate(() => goSlidesSetSource("---\nslides: true\n---\n\n# Alive\n\n![clip](media/clip.mp4)\n"));
  await until(page, "the page alive after the refusals", () => /Alive/.test(goSlidesDebug().source) && !goSlidesDebug().busy && !/video\./.test(JSON.stringify(goSlidesDebug().problems)));

  // 4. the exports
  async function save(key, name) {
    for (const f of fs.readdirSync(dir)) if (f.startsWith("deck.")) fs.rmSync(path.join(dir, f));
    const tab = (await debug()).rects["tab-Export"];
    await page.mouse.click(tab[0] + tab[2] / 2, tab[1] + tab[3] / 2);
    const r = await until(page, "the " + key + " button", (k) => goSlidesDebug().rects[k], key);
    const before = done.size;
    await page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2);
    const status = await until(page, "saved " + name, (n) => { const s = goSlidesDebug().status; return (s.startsWith("Saved " + n) || s.startsWith("⚠")) && s; }, name);
    if (status.startsWith("⚠")) fail(status);
    const file = path.join(dir, name);
    const stop = Date.now() + 30000;
    let last = -1;
    for (;;) {
      const there = fs.existsSync(file) && !fs.existsSync(file + ".crdownload");
      const size = there ? fs.statSync(file).size : -1;
      if (done.size > before && there && size > 0 && size === last) break;
      last = size;
      if (Date.now() > stop) fail("the download of " + name + " did not finish");
      await sleep(150);
    }
    return file;
  }
  const unzipList = (f) => execFileSync("unzip", ["-Z1", f]).toString().split("\n");
  const pptx = await save("save-pptx", "deck.pptx");
  const pl = unzipList(pptx);
  if (!pl.some((n) => /^ppt\/media\/.*\.mp4$/.test(n))) fail("the PPTX holds no video: " + pl.join(" "));
  if (!pl.some((n) => /^ppt\/media\/(?!slidesvideo).*\.(png|jpe?g)$/.test(n))) fail("the PPTX holds no poster: " + pl.join(" "));
  const odp = await save("save-odp", "deck.odp");
  const ol = unzipList(odp);
  if (!ol.some((n) => /^Video\/.*\.mp4$/.test(n))) fail("the ODP holds no video: " + ol.join(" "));
  const pdf = await save("save-pdf", "deck.pdf");
  const pdfBytes = fs.readFileSync(pdf);
  if (!pdfBytes.subarray(0, 5).equals(Buffer.from("%PDF-"))) fail("not a PDF");
  // the deck holds nothing but the video, so an image in the PDF is its poster
  if (!/\/Subtype\s*\/Image/.test(pdfBytes.toString("latin1"))) fail("the PDF carries no picture: the poster is missing");
  if (videoRequests.length !== 1) fail("the exports fetched the video worker again: " + videoRequests.length);

  if (errors.length) fail("page errors: " + errors.join(" | "));
  if (violations.length) fail("the policy refused something it should allow: " + violations.join(" | "));
  console.log("VIDEOWORKER_OK " + JSON.stringify({ videoWorkerRequests: videoRequests.length, pptx: pl.length, odp: ol.length, pdf: pdfBytes.length }));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
