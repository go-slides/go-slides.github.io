// Time to the editor ready, cold against warm, with the offline cache:
//
//   node wasmcache-bench.cjs <built site, wasmcache already run> [Mbit/s=50] [rounds=7]
//
// cold = an empty browser profile (nothing cached, no worker); warm = the
// same profile on its second visit, the worker installed and the build in
// Cache Storage. The server gzips what Pages gzips, and paces the bytes it
// sends at the given line rate (the wasm is 11.9 MiB gzipped), so the figure
// is a download's, not loopback's. The rounds interleave cold and warm; the
// median and every sample are printed.
//
// ⛔ Times are not measured when the machine is busy: with a 1-minute load
// above 8 it prints DEFERRED and exits 3 (the machine is shared).
//
// CHROME names the browser; NODE_PATH finds puppeteer-core.
"use strict";
const http = require("http");
const fs = require("fs");
const os = require("os");
const path = require("path");
const zlib = require("zlib");
const puppeteer = require("puppeteer-core");

const root = path.resolve(process.argv[2] || ".");
const mbit = Number(process.argv[3] || 50), rounds = Number(process.argv[4] || 7);
const load = os.loadavg()[0];
if (load > 8) { console.log("DEFERRED: 1-minute load " + load.toFixed(1) + " is above 8"); process.exit(3); }

const gz = new Map();
const server = http.createServer((req, res) => {
  const u = new URL(req.url, "http://x");
  const p = u.pathname.endsWith("/") ? u.pathname + "index.html" : u.pathname;
  let body;
  try { body = fs.readFileSync(path.join(root, path.normalize(p))); } catch (e) { res.writeHead(404); return res.end(); }
  const types = { ".html": "text/html", ".js": "text/javascript", ".json": "application/json", ".wasm": "application/wasm", ".css": "text/css", ".ttf": "font/ttf" };
  const h = { "content-type": types[path.extname(p)] || "application/octet-stream", "cache-control": "max-age=600" };
  if (body.length > 4096 && req.headers["accept-encoding"] && /gzip/.test(req.headers["accept-encoding"])) {
    if (!gz.has(p)) gz.set(p, zlib.gzipSync(body, { level: 6 }));
    body = gz.get(p);
    h["content-encoding"] = "gzip";
  }
  h["content-length"] = body.length;
  res.writeHead(200, h);
  // paced: 64 KiB at a time, at the line rate
  const chunk = 65536, ms = (chunk * 8) / (mbit * 1000);
  let off = 0;
  (function send() {
    if (off >= body.length) return res.end();
    res.write(body.subarray(off, off + chunk));
    off += chunk;
    setTimeout(send, ms);
  })();
});

const median = (a) => a.slice().sort((x, y) => x - y)[a.length >> 1];
(async () => {
  await new Promise((r) => server.listen(0, "127.0.0.1", r));
  const url = "http://127.0.0.1:" + server.address().port + "/playground/";
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const visit = async (ctx) => {
    const page = await ctx.newPage();
    const t0 = Date.now();
    await page.goto(url, { waitUntil: "load" });
    await page.waitForFunction(() => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready") && !goSlidesDebug().busy, { timeout: 120000, polling: 20 });
    const ms = Date.now() - t0;
    const t = await page.evaluate(() => globalThis.goSlidesTimings);
    return { page, ms, inPage: t };
  };
  const warm = await browser.createBrowserContext();
  const prime = await visit(warm);
  await prime.page.evaluate(() => wasmcache.cached());
  await prime.page.close();
  const cold = [], hot = [];
  for (let i = 0; i < rounds; i++) {
    const c = await browser.createBrowserContext();
    const a = await visit(c);
    cold.push(a.ms);
    await a.page.close();
    await c.close();
    const w = await visit(warm);
    hot.push(w.ms);
    await w.page.close();
  }
  console.log(JSON.stringify({ lineMbit: mbit, rounds, load1: load, coldMs: cold, warmMs: hot, coldMedian: median(cold), warmMedian: median(hot) }));
  await browser.close();
  server.close();
})().catch((e) => { console.error("FAIL " + (e && e.stack || e)); process.exit(1); });
