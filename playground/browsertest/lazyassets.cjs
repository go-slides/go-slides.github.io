// The files the render worker fetches on demand, in headless Chrome. The page's
// module is built without go-slides' math font and emoji graphics
// (slides_nomathfont, slides_noemoji); the worker fetches fonts/stix-math.otf the
// first time a render has a formula, and emoji/twemoji.pack the first time it has
// an emoji. For each of the two:
//
//  1. a deck without it never requests the file;
//  2. a deck with it requests it once; until it arrives the formula or emoji is
//     the text written (span.math-error; the emoji character), when it arrives
//     the preview is rendered again, once, and it is drawn (span.slides-math; an
//     SVG picture), the text gone; an edit afterwards does not fetch again;
//  3. a failed fetch (a 503) leaves the text, nothing asks again by itself (no
//     loop), and the next render fetches again and draws;
//  4. a body that is not exactly the file the build wrote (a byte changed, or
//     longer) is refused like a failed fetch.
//
//   node lazyassets.cjs SITE_DIR
//
// SITE_DIR is the built site (playground/index.html under it): this script serves
// it itself, to count the requests, delay one and refuse another, without the
// offline cache (its worker answers a failed lazy file with a request of its own:
// the cache and these files are the wasmcache kit's, with
// browsertest/wasmcache-hooks.cjs).
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. Prints LAZYASSETS_OK.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const http = require("http");
const path = require("path");

const root = path.resolve(process.argv[2] || "");
if (!process.argv[2] || !fs.existsSync(path.join(root, "playground", "index.html"))) {
  console.error("usage: node lazyassets.cjs SITE_DIR  (the built site: playground/index.html under it)");
  process.exit(2);
}

// a shared link with a deck that has no formula and no emoji (a case's start)
const NOMATH_FRAGMENT = "deck=0tXV5SrOyUxJLbZSKCkqTeUCCXApK_inpeVk5qUqJCcmZ6RycTkqpKQmZyuUZ5Zk5JeWKCQqpOUX5ZbmJOpxAQYA";

const KINDS = {
  math: {
    file: "/playground/fonts/stix-math.otf",
    deck: "---\nslides: true\n---\n\n# Energy\n\n$E=mc^2$\n",
    // what the preview and thumbnail frames show: drawn, and as the text written
    count: () => [document.querySelectorAll(".slides-math").length, document.querySelectorAll(".math-error").length],
  },
  emoji: {
    file: "/playground/emoji/twemoji.pack",
    deck: "---\nslides: true\n---\n\n# Smile\n\n\u{1F604}\n",
    count: () => [document.querySelectorAll('img[src^="data:image/svg+xml"]').length, document.body.textContent.includes("\u{1F604}") ? 1 : 0],
  },
};

const fail = (m) => { console.error("FAIL " + m); process.exit(1); };
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// the server: the site, with each lazy file's requests counted, delayed or refused
const state = {};
for (const k of Object.values(KINDS)) state[k.file] = { hits: 0, delayMs: 0, failures: 0, tamper: "" };
const server = http.createServer((req, res) => {
  const u = new URL(req.url, "http://x");
  const p = u.pathname.endsWith("/") ? u.pathname + "index.html" : u.pathname;
  if (p.endsWith("/wasmcache-sw.js")) { res.writeHead(404); return res.end("no"); }
  const f = state[u.pathname];
  if (f) {
    f.hits++;
    if (f.failures > 0) {
      f.failures--;
      // after the delay, as a slow server would: the render that asked is over
      return setTimeout(() => {
        res.writeHead(503, { "content-type": "text/plain", "cache-control": "no-store" });
        res.end("down");
      }, f.delayMs);
    }
  }
  let body;
  try { body = fs.readFileSync(path.join(root, path.normalize(p))); } catch (e) { res.writeHead(404); return res.end("no"); }
  if (f && f.tamper === "flip") { body = Buffer.from(body); body[body.length >> 1] ^= 1; }
  if (f && f.tamper === "longer") body = Buffer.concat([body, Buffer.from("x")]);
  const types = { ".html": "text/html", ".js": "text/javascript", ".json": "application/json", ".wasm": "application/wasm", ".css": "text/css", ".otf": "font/otf" };
  const send = () => {
    res.writeHead(200, { "content-type": types[path.extname(p)] || "application/octet-stream", "cache-control": "no-cache" });
    res.end(body);
  };
  if (f && f.delayMs) setTimeout(send, f.delayMs); else send();
});

(async () => {
  await new Promise((r) => server.listen(0, "127.0.0.1", r));
  const base = "http://127.0.0.1:" + server.address().port;
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const errors = [], seen = []; // seen: the page under test, for the report of a timeout
  const total = () => Object.values(state).reduce((n, f) => n + f.hits, 0);

  // a fresh browser context for each case: its own storage
  const open = async (fragment) => {
    const ctx = await browser.createBrowserContext();
    const page = await ctx.newPage();
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
    page.on("pageerror", (e) => errors.push(e.message));
    seen.length = 0;
    seen.push(page);
    await page.goto(base + "/playground/" + (fragment ? "#" + fragment : ""), { waitUntil: "load" });
    await page.waitForFunction(() => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready") && !goSlidesDebug().busy, { timeout: 120000 });
    return page;
  };
  const setSource = async (page, src) => {
    await page.evaluate((s) => goSlidesSetSource(s), src);
    await page.waitForFunction((s) => goSlidesDebug().source === s, { timeout: 10000 }, src);
    await page.waitForFunction(() => !goSlidesDebug().busy, { timeout: 60000 });
  };
  // what the preview and thumbnail frames show: [drawn, as text]
  const shown = async (page, kind) => {
    let drawn = 0, text = 0;
    for (const f of page.frames()) {
      if (f === page.mainFrame()) continue;
      const c = await f.evaluate(kind.count).catch(() => [0, 0]);
      drawn += c[0];
      text += c[1];
    }
    return { drawn, text };
  };
  const waitFor = async (what, pred, ms = 60000) => {
    const end = Date.now() + ms;
    for (;;) {
      const v = await pred();
      if (v) return v;
      if (Date.now() > end) {
        let more = "";
        for (const pg of seen) {
          more += " | " + JSON.stringify(await pg.evaluate(() => { const d = goSlidesDebug(); return { status: d.status, problems: d.problems, busy: d.busy }; }).catch(() => "?"));
          for (const f of pg.frames()) if (f !== pg.mainFrame()) more += " | frame: " + (await f.evaluate(() => (document.querySelector("section") || document.body).innerHTML.slice(0, 300)).catch(() => "?"));
        }
        fail("timed out waiting for " + what + more);
      }
      await sleep(50);
    }
  };
  const renders = (page) => page.evaluate(() => { const d = goSlidesDebug(); return d.loads + d.patches; });

  for (const [name, kind] of Object.entries(KINDS)) {
    const f = state[kind.file];
    const reset = (delayMs, failures) => { for (const s of Object.values(state)) s.hits = 0; f.delayMs = delayMs; f.failures = failures; };

    // 1. nothing of it in the deck, nothing fetched (neither file)
    reset(0, 0);
    let page = await open(NOMATH_FRAGMENT);
    await setSource(page, "---\nslides: true\n---\n\n# Plain\n\ntext\n\n---\n\n# Two\n");
    await sleep(1500);
    if (total() !== 0) fail(name + ": a deck without it fetched " + JSON.stringify(state));
    await page.browserContext().close();

    // 2. one fetch; the text until it arrives, then drawn by one more render
    reset(1500, 0);
    page = await open(NOMATH_FRAGMENT);
    if (total() !== 0) fail(name + ": the page fetched a file before any deck used one");
    await setSource(page, kind.deck);
    await waitFor(name + " as text", async () => (await shown(page, kind)).text > 0 && f.hits === 1, 20000);
    const during = await shown(page, kind);
    if (during.drawn !== 0) fail(name + " was drawn before the file arrived: " + JSON.stringify(during));
    await waitFor(name + " drawn", async () => { const s = await shown(page, kind); return s.drawn > 0 && s.text === 0; });
    if (f.hits !== 1) fail(name + ": the file was requested " + f.hits + " times for one deck");
    if (name === "math" && state[KINDS.emoji.file].hits) fail("a formula fetched the emoji pack");
    if (name === "emoji" && state[KINDS.math.file].hits) fail("an emoji fetched the math font");
    // the render happened once and nothing keeps rendering
    await waitFor("the render to settle", async () => !(await page.evaluate(() => goSlidesDebug().busy)));
    const settled = await renders(page);
    await sleep(2000);
    if ((await renders(page)) !== settled) fail(name + ": the preview kept rendering after the file arrived");
    // an edit does not fetch it again, and what it draws is drawn at once
    await setSource(page, kind.deck + "\nmore\n");
    await waitFor(name + " drawn after an edit", async () => { const s = await shown(page, kind); return s.drawn > 0 && s.text === 0; });
    await sleep(500);
    if (f.hits !== 1) fail(name + ": an edit fetched the file again: " + f.hits);
    await page.browserContext().close();

    // 3. a failed fetch: the text stays, nothing asks again by itself, the next render does
    reset(800, 1);
    page = await open(NOMATH_FRAGMENT);
    await setSource(page, kind.deck);
    await waitFor(name + ": the failed fetch", async () => f.hits === 1);
    await waitFor(name + " as text", async () => (await shown(page, kind)).text > 0);
    await sleep(1500); // the 503 is back by now
    const settled2 = await renders(page);
    await sleep(2500);
    if (f.hits !== 1) fail(name + ": a failed fetch was tried again by itself: " + f.hits);
    if ((await renders(page)) !== settled2) fail(name + ": a failed fetch made the preview render again");
    const after = await shown(page, kind);
    if (after.drawn !== 0 || after.text === 0) fail(name + ": after the failure it is not the text written: " + JSON.stringify(after));
    await setSource(page, kind.deck + "\nmore\n"); // the next render
    await waitFor(name + ": the retry", async () => f.hits === 2);
    await waitFor(name + " drawn after the retry", async () => { const s = await shown(page, kind); return s.drawn > 0 && s.text === 0; });
    await sleep(1000);
    if (f.hits !== 2) fail(name + ": the file was requested " + f.hits + " times after one failure and one retry");
    await page.browserContext().close();

    // 4. a body that is not the file the build wrote (same size, one byte changed; and
    // longer) is refused as a failed fetch: the text stays, nothing asks again by
    // itself, the next render fetches again and takes the real file
    for (const mode of ["flip", "longer"]) {
      reset(0, 0);
      f.tamper = mode;
      page = await open(NOMATH_FRAGMENT);
      await setSource(page, kind.deck);
      await waitFor(name + ": the " + mode + " fetch", async () => f.hits === 1);
      await sleep(2500);
      if (f.hits !== 1) fail(name + " (" + mode + "): a refused body was fetched again by itself: " + f.hits);
      const refused = await shown(page, kind);
      if (refused.drawn !== 0 || refused.text === 0) fail(name + " (" + mode + "): a body that is not the file was used: " + JSON.stringify(refused));
      f.tamper = "";
      await setSource(page, kind.deck + "\nmore\n");
      await waitFor(name + " (" + mode + ") drawn after the real file", async () => { const s = await shown(page, kind); return s.drawn > 0 && s.text === 0; });
      if (f.hits !== 2) fail(name + " (" + mode + "): " + f.hits + " requests");
      await page.browserContext().close();
    }
    f.tamper = "";
  }

  if (errors.length) fail("page errors: " + errors.join(" | "));
  console.log("LAZYASSETS_OK " + JSON.stringify({ kinds: Object.keys(KINDS), noRequests: 0, oneFetch: 1, afterFailure: 2 }));
  await browser.close();
  server.close();
})().catch((e) => fail(e.stack || String(e)));
