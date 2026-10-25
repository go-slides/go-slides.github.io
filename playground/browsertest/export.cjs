// Drives the export in headless Chrome: Save as ▸ PDF downloads the export
// engine once and saves a PDF; PPTX and ODP then reuse it; with BENCH_DECK,
// the 100-slide deck's PDF is timed. Each saved file is checked by its
// magic bytes. Prints one JSON line, EXPORT_OK {...}.
//
//   node export.cjs http://127.0.0.1:PORT/playground/ DOWNLOAD_DIR
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. NET_PROFILE
// ("bytesPerSecond,latencyMs", e.g. 3145728,40) throttles the network once
// the editor is ready, so that the first export's download is timed on a
// known link (Chrome's own emulation), the same for every build compared.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const path = require("path");

const url = process.argv[2];
const dir = path.resolve(process.argv[3] || ".");
fs.mkdirSync(dir, { recursive: true });

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}

async function until(page, what, fn, arg, ms = 180000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg).catch(() => null);
    if (v) return v;
    if (Date.now() > end) fail("timed out waiting for " + what + ": " + (await page.evaluate(() => goSlidesDebug().status).catch(() => "?")));
    await new Promise((r) => setTimeout(r, 50));
  }
}

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  page.on("pageerror", (e) => console.error("pageerror:", e.message));
  const cdp = await page.createCDPSession();
  await cdp.send("Browser.setDownloadBehavior", { behavior: "allow", downloadPath: dir, eventsEnabled: true });
  const done = new Set();
  cdp.on("Browser.downloadProgress", (e) => { if (e.state === "completed") done.add(e.guid); });
  await page.goto(url, { waitUntil: "load" });
  await until(page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));

  const out = {};
  if (process.env.NET_PROFILE) {
    const [bps, latency] = process.env.NET_PROFILE.split(",").map(Number);
    if (!(bps > 0) || !(latency >= 0)) fail("NET_PROFILE wants bytesPerSecond,latencyMs: " + process.env.NET_PROFILE);
    await cdp.send("Network.enable");
    await cdp.send("Network.emulateNetworkConditions", { offline: false, latency, downloadThroughput: bps, uploadThroughput: bps });
    out.netProfile = { bytesPerSecond: bps, latencyMs: latency };
  }
  // The Export tab's buttons: key is the control's name in goSlidesDebug().rects.
  async function save(key, name, magic) {
    for (const f of fs.readdirSync(dir)) if (f.startsWith("deck.")) fs.rmSync(path.join(dir, f));
    const tab = (await page.evaluate(() => goSlidesDebug())).rects["tab-Export"];
    await page.mouse.click(tab[0] + tab[2] / 2, tab[1] + tab[3] / 2);
    const r = await until(page, "the " + key + " button", (k) => goSlidesDebug().rects[k], key);
    const t0 = Date.now();
    const before = done.size;
    await page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2);
    const status = await until(page, "saved " + name, (n) => { const s = goSlidesDebug().status; return (s.startsWith("Saved " + n) || s.startsWith("⚠")) && s; }, name);
    if (status.startsWith("⚠")) fail(status);
    const ms = Date.now() - t0;
    const file = path.join(dir, name);
    // The download is finished when Chrome says it completed (it writes
    // name.crdownload and renames it), the final name is there, and its size has
    // not changed across two polls: a file read while it is still being written
    // is empty or short.
    const end = Date.now() + 30000;
    let last = -1;
    for (;;) {
      const there = fs.existsSync(file) && !fs.existsSync(file + ".crdownload");
      const size = there ? fs.statSync(file).size : -1;
      if (done.size > before && there && size > 0 && size === last) break;
      last = size;
      if (Date.now() > end) fail("the download of " + name + " did not finish in " + dir + " (completed " + (done.size - before) + ", size " + size + ")");
      await new Promise((r) => setTimeout(r, 150));
    }
    const b = fs.readFileSync(file);
    if (b.length === 0) fail(name + " is empty");
    if (!b.subarray(0, magic.length).equals(Buffer.from(magic))) fail(name + " starts " + JSON.stringify(b.subarray(0, 8).toString("latin1")));
    return { ms, bytes: b.length, status };
  }

  out.pdfFirst = await save("save-pdf", "deck.pdf", "%PDF-");
  out.pptx = await save("save-pptx", "deck.pptx", "PK");
  out.odp = await save("save-odp", "deck.odp", "PK");
  out.pdfAgain = await save("save-pdf", "deck.pdf", "%PDF-");
  if (process.env.BENCH_DECK) {
    await page.evaluate((t) => goSlidesSetSource(t), fs.readFileSync(process.env.BENCH_DECK, "utf8"));
    await until(page, "the 100-slide deck", () => /100 slides/.test(goSlidesDebug().status));
    out.bench100Pdf = await save("save-pdf", "deck.pdf", "%PDF-");
    fs.copyFileSync(path.join(dir, "deck.pdf"), path.join(dir, "bench100-browser.pdf"));
    out.bench100Pptx = await save("save-pptx", "deck.pptx", "PK");
  }
  out.engine = await page.evaluate(() => performance.getEntriesByType("resource").filter((e) => e.name.includes("export.wasm")).map((e) => ({ file: new URL(e.name).pathname, transferSize: e.transferSize, encodedBodySize: e.encodedBodySize, decodedBodySize: e.decodedBodySize, ms: Math.round(e.duration) })));
  console.log("EXPORT_OK " + JSON.stringify(out));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
