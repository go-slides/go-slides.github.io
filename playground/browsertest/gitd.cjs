// The git pane in SERVER MODE, in headless Chrome, against the real slides-gitd
// server code (cmd/gitdfront: the site, slides-gitd over an in-memory forge, and
// oauth2-proxy faked by its trusted-header mode: a "test-user" cookie).
//
// Two people, Alice and Bob, each in a browser context of their own (own cookies,
// own storage), with different repositories and different forge tokens. The test
// asserts what a person and what an attacker could observe:
//
//   - discovery: signed out the pane offers Sign in and no token field; Sign in
//     goes to the proxy and back; signed in it shows the identity and ONLY that
//     person's repositories;
//   - Load, Propose (the server names the branch under that person's directory;
//     the forge saw that person's token), Check for changes, Take theirs, all to
//     the page's own origin: no request leaves it, no forge host is contacted;
//   - every write carries the exact Origin, X-Requested-With, a JSON body and the
//     CSRF token; the session cookie is HttpOnly, Secure, SameSite=Strict and not
//     in document.cookie; the CSRF token and the forge tokens are in no storage,
//     no IndexedDB, no URL, no frame, no worker, no page state;
//   - isolation: neither person ever sees the other's repository, deck, branch or
//     session, not even when one is handed the other's cookie, CSRF token or
//     identity header; a stolen cookie destroys the session it was stolen from
//     and the owner signs in again by themselves (a new session) without loss;
//   - an ended sign-in shows Sign in again and uses no token instead; the CSP
//     still refuses every other host.
//
//   node gitd.cjs http://127.0.0.1:PORT/playground/
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. Prints GITD_OK {...}.
"use strict";
const puppeteer = require("puppeteer-core");

const url = process.argv[2];
const origin = new URL(url).origin;
const ALICE_TOKEN = "ALICE-FORGE-TOKEN-PLACEHOLDER-0001";
const BOB_TOKEN = "BOB-FORGE-TOKEN-PLACEHOLDER-0002";

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}

async function until(page, what, fn, arg, ms = 30000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg).catch(() => null);
    if (v) return v;
    if (Date.now() > end) fail("timed out waiting for " + what + ": " + (await page.evaluate(() => JSON.stringify(goSlidesDebug().git)).catch(() => "?")));
    await new Promise((r) => setTimeout(r, 25));
  }
}

async function click(page, key) {
  const r = await until(page, "the control " + key, (k) => goSlidesDebug().rects[k], key);
  await page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2);
}

async function typeInto(page, key, text) {
  await click(page, key);
  await page.keyboard.press("End"); // empty the field whatever it held
  for (let i = 0; i < 40; i++) await page.keyboard.press("Backspace");
  await page.keyboard.type(text);
}

const gitState = (page) => page.evaluate(() => goSlidesDebug().git);

async function setCookie(page, name, value) {
  const cdp = await page.createCDPSession();
  const host = name.startsWith("__Host-"); // the prefix demands Secure, no Domain, Path=/
  await cdp.send("Network.setCookie", { name, value, url: origin + "/", path: "/", httpOnly: true, secure: host, sameSite: host ? "Strict" : "Lax" });
  await cdp.detach();
}
async function deleteCookie(page, name) {
  const cdp = await page.createCDPSession();
  await cdp.send("Network.deleteCookies", { name, url: origin + "/" });
  await cdp.detach();
}
async function allCookies(page) {
  const cdp = await page.createCDPSession();
  const { cookies } = await cdp.send("Network.getCookies", { urls: [origin + "/"] });
  await cdp.detach();
  return cookies;
}

// everything the page, its frames and its worker can reach, as one string
// storeOnly: the page's own state and storage, not what the frames and the worker
// received while it was open (their logs outlive nothing: the page's lifetime)
async function sweep(page, storeOnly) {
  const top = await page.evaluate(async () => {
    const dec = new TextDecoder();
    const str = (v) => {
      if (v instanceof ArrayBuffer) return dec.decode(new Uint8Array(v));
      if (ArrayBuffer.isView(v)) return dec.decode(v);
      if (v && typeof v === "object") return Object.keys(v).map((k) => k + "=" + str(v[k])).join("\n");
      return String(v);
    };
    let out = "URL " + location.href + "\nCOOKIE " + document.cookie + "\n";
    for (const st of [localStorage, sessionStorage]) for (let i = 0; i < st.length; i++) out += st.key(i) + "=" + st.getItem(st.key(i)) + "\n";
    for (const info of await indexedDB.databases()) {
      const db = await new Promise((ok, ko) => { const r = indexedDB.open(info.name); r.onsuccess = () => ok(r.result); r.onerror = () => ko(r.error); });
      for (const name of db.objectStoreNames) {
        const all = await new Promise((ok, ko) => { const r = db.transaction(name).objectStore(name).getAll(); r.onsuccess = () => ok(r.result); r.onerror = () => ko(r.error); });
        out += "IDB " + info.name + "/" + name + " " + str(all) + "\n";
      }
      db.close();
    }
    out += "DEBUG " + JSON.stringify(goSlidesDebug()) + "\n";
    return out;
  });
  if (storeOnly) return top;
  let frames = "";
  for (const f of page.frames()) frames += await f.evaluate(() => (globalThis.__seen || []).join("\n")).catch(() => "");
  let workers = "";
  for (const w of page.workers()) workers += await w.evaluate(() => (globalThis.__seen || []).join("\n")).catch(() => "");
  return top + "\n" + frames + "\n" + workers;
}

// a person: a browser context of their own, a page, and a record of every request
async function person(browser, name) {
  const ctx = await browser.createBrowserContext();
  const page = await ctx.newPage();
  const p = { name, ctx, page, requests: [], posts: [], dialogs: 0 };
  page.on("pageerror", (e) => fail(name + " pageerror: " + e.message));
  page.on("dialog", (d) => { p.dialogs++; d.accept(); });
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  await page.setRequestInterception(true);
  page.on("request", (req) => {
    p.requests.push(req.url());
    if (req.method() === "POST") p.posts.push({ url: req.url(), headers: req.headers(), body: req.postData() || "" });
    req.continue();
  });
  const record = () => addEventListener("message", (e) => { try { (globalThis.__seen = globalThis.__seen || []).push(typeof e.data === "string" ? e.data : JSON.stringify(e.data)); } catch (_) {} });
  page.on("frameattached", (f) => f.evaluate(record).catch(() => {}));
  page.on("framenavigated", (f) => f.evaluate(record).catch(() => {}));
  page.on("workercreated", (w) => w.evaluate(record).catch(() => {}));
  return p;
}

async function open(p) {
  await p.page.goto(url, { waitUntil: "load" });
  await until(p.page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
  for (const f of p.page.frames()) await f.evaluate(() => addEventListener("message", (e) => { try { (globalThis.__seen = globalThis.__seen || []).push(typeof e.data === "string" ? e.data : JSON.stringify(e.data)); } catch (_) {} })).catch(() => {});
  await click(p.page, "git");
}

const serverState = async () => (await fetch(origin + "/__test/state")).json();

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const out = {};
  const alice = await person(browser, "alice"), bob = await person(browser, "bob"), carol = await person(browser, "signed-out");

  // 1. Signed out: Sign in, no token field; the control goes to the proxy and back.
  await open(carol);
  let st = await until(carol.page, "signed-out pane", () => goSlidesDebug().git.mode === 1 && goSlidesDebug().git);
  let rects = await carol.page.evaluate(() => Object.keys(goSlidesDebug().rects));
  for (const k of ["git-token", "git-remember", "git-connect", "git-forge", "git-signin"]) if (rects.includes(k)) fail("signed out, the pane shows " + k);
  if (!rects.includes("git-ssignin")) fail("signed out, no Sign in");
  await setCookie(carol.page, "test-intent", "sub-alice");
  await Promise.all([carol.page.waitForNavigation({ waitUntil: "load" }), click(carol.page, "git-ssignin")]);
  await until(carol.page, "the editor after the sign-in", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
  await click(carol.page, "git");
  st = await until(carol.page, "signed in after Sign in", () => goSlidesDebug().git.mode === 2 && goSlidesDebug().git);
  if (st.user !== "sub-alice") fail("signed in as " + st.user);
  out.signin = "signed out shows Sign in and no token field; Sign in goes to the proxy and back, signed in";

  // 2. Alice and Bob: each sees only their own repositories.
  await setCookie(alice.page, "test-user", "sub-alice");
  await setCookie(bob.page, "test-user", "sub-bob");
  await open(alice);
  await open(bob);
  const sa = await until(alice.page, "alice signed in", () => goSlidesDebug().git.mode === 2 && goSlidesDebug().git);
  const sb = await until(bob.page, "bob signed in", () => goSlidesDebug().git.mode === 2 && goSlidesDebug().git);
  if (sa.repos !== "github:lab/alpha" || sa.user !== "sub-alice") fail("alice's pane: " + JSON.stringify(sa));
  if (sb.repos !== "github:lab/beta" || sb.user !== "sub-bob") fail("bob's pane: " + JSON.stringify(sb));
  rects = await alice.page.evaluate(() => Object.keys(goSlidesDebug().rects));
  for (const k of ["git-token", "git-remember", "git-connect", "git-forge", "git-signin", "git-frame"]) if (rects.includes(k)) fail("server mode shows " + k);
  // The git frame (go-slides#90) is client mode's: in server mode the tokens are
  // on the server, and the page lays no frame over the pane (it may not even
  // have created one).
  const gf = await alice.page.evaluate(() => { const e = document.getElementById("slides-gitframe"); return e ? e.style.display : "absent"; });
  if (gf !== "absent" && gf !== "none") fail("server mode shows the git frame: display " + gf);
  out.repositories = "alice sees lab/alpha only, bob lab/beta only; no token, Remember, Connect, forge, OAuth or git-frame control";

  // 3. Load, Propose, Check for changes, Take theirs.
  for (const [p, repo, deck] of [[alice, "lab/alpha", "Alpha deck of Alice"], [bob, "lab/beta", "Beta deck of Bob"]]) {
    await typeInto(p.page, "git-dir", "talk");
    await click(p.page, "git-load");
    await until(p.page, p.name + "'s deck", (d) => goSlidesDebug().source.includes(d), deck);
    const other = p === alice ? "Beta deck" : "Alpha deck";
    if ((await p.page.evaluate(() => goSlidesDebug().source)).includes(other)) fail(p.name + " has the other's deck");
  }
  await click(alice.page, "git-propose");
  await until(alice.page, "alice's pull request", () => goSlidesDebug().git.link !== "");
  await click(bob.page, "git-propose");
  await until(bob.page, "bob's pull request", () => goSlidesDebug().git.link !== "");
  let state = await serverState();
  const pa = state.proposals.filter((x) => x.Repo === "lab/alpha"), pb = state.proposals.filter((x) => x.Repo === "lab/beta");
  if (pa.length !== 1 || pb.length !== 1 || state.proposals.length !== 2) fail("proposals: " + JSON.stringify(state.proposals));
  if (pa[0].Token !== ALICE_TOKEN || pb[0].Token !== BOB_TOKEN) fail("the forge saw the wrong token for a proposal: " + JSON.stringify(state.proposals));
  if (!pa[0].Branch.startsWith("slides/alice-") || !pb[0].Branch.startsWith("slides/bob-")) fail("branch directories: " + pa[0].Branch + " " + pb[0].Branch);
  if (pa[0].Branch.includes("bob") || pb[0].Branch.includes("alice")) fail("a branch carries the other's name");
  await click(alice.page, "git-refresh");
  await until(alice.page, "up to date", () => /Nothing changed|rien/i.test(goSlidesDebug().git.note));
  await fetch(origin + "/__test/advance?repo=lab/alpha&title=" + encodeURIComponent("Alpha v2 by a colleague"));
  await click(alice.page, "git-refresh");
  await until(alice.page, "the changes found", () => /changes \(/.test(goSlidesDebug().git.note));
  await click(alice.page, "git-theirs");
  await until(alice.page, "their deck taken", () => goSlidesDebug().source.includes("Alpha v2 by a colleague"));
  out.actions = "Load, Propose (branch under the person's own directory, the forge saw their token), Check for changes, Take theirs";

  // 4. Same origin only; the writes carry what slides-gitd demands.
  for (const p of [alice, bob]) {
    for (const u of p.requests) if (new URL(u).origin !== origin && !u.startsWith("data:") && !u.startsWith("blob:")) fail(p.name + " contacted " + u);
    if (p.requests.some((u) => /github|gitlab/.test(new URL(u).hostname))) fail(p.name + " contacted a forge");
    const gitd = p.posts.filter((r) => r.url.startsWith(origin + "/api/git/"));
    if (gitd.length < 2) fail(p.name + " made " + gitd.length + " writes");
    for (const r of gitd) {
      const h = r.headers;
      if (h["origin"] !== origin || h["x-requested-with"] !== "slides-gitd" || !/^application\/json/.test(h["content-type"] || "") || !h["x-csrf-token"] || h["x-csrf-token"].length < 16) fail(p.name + " write headers: " + JSON.stringify(h));
      JSON.parse(r.body);
    }
  }
  out.sameOrigin = "every request to the page's own origin; each write with Origin, X-Requested-With, JSON, X-CSRF-Token";

  // 5. The session cookie, the CSRF token and the forge tokens are in no JS-visible place.
  const sessions = {};
  for (const p of [alice, bob]) {
    const ck = (await allCookies(p.page)).find((c) => c.name === "__Host-gitd");
    if (!ck || !ck.httpOnly || !ck.secure || ck.sameSite !== "Strict" || ck.path !== "/" || ck.domain.startsWith(".")) fail(p.name + " session cookie: " + JSON.stringify(ck));
    sessions[p.name] = ck.value;
    const csrf = (await p.page.evaluate(async () => (await (await fetch("/api/git/session")).json()).csrf));
    if (!csrf || csrf.length < 16) fail("no csrf");
    p.csrf = csrf;
    const s = await sweep(p.page);
    for (const secret of [ck.value, csrf, ALICE_TOKEN, BOB_TOKEN]) if (s.includes(secret)) fail(p.name + ": a secret is in the page's storage, state, frames or worker: " + secret.slice(0, 12));
    for (const u of p.requests) if (u.includes(ck.value) || u.includes(csrf) || u.includes(ALICE_TOKEN) || u.includes(BOB_TOKEN)) fail(p.name + " secret in a URL");
    if (/gitd/.test(await p.page.evaluate(() => document.cookie))) fail("the session cookie is readable by script");
    if (!s.includes("Alpha") && !s.includes("Beta")) fail("control: the sweep should see the deck");
  }
  if (sessions.alice === sessions.bob) fail("two people share a session");
  out.secrets = "the cookie is HttpOnly+Secure+SameSite=Strict+__Host-, unreadable by script; cookie, CSRF and forge tokens in no storage, IndexedDB, URL, frame, worker or page state";

  // 6. Isolation under attack, from Bob's page.
  const attack = (body, headers) => bob.page.evaluate(async (b, h) => {
    const r = await fetch("/api/git/load", { method: "POST", headers: Object.assign({ "X-Requested-With": "slides-gitd", "Content-Type": "application/json" }, h), body: JSON.stringify(b) });
    return { status: r.status, text: await r.text() };
  }, body, headers);
  const alpha = { forge: "github", repo: "lab/alpha", dir: "talk" };
  // (a) his own session and token, her repository
  let r = await attack(alpha, { "X-CSRF-Token": bob.csrf });
  if (r.status !== 403 || /Alpha/.test(r.text)) fail("bob read alice's repository: " + r.status + " " + r.text.slice(0, 80));
  // (b) her CSRF token with his session
  r = await attack(alpha, { "X-CSRF-Token": alice.csrf });
  if (r.status !== 403 || /Alpha/.test(r.text)) fail("alice's CSRF token worked in bob's session: " + r.status);
  // (c) a forged identity header is deleted by the front, as by Caddy
  r = await attack(alpha, { "X-CSRF-Token": bob.csrf, "X-Auth-Request-User": "sub-alice" });
  if (r.status !== 403 || /Alpha/.test(r.text)) fail("a forged identity header was believed: " + r.status);
  // (d) her cookie in his browser: refused, and her session is destroyed
  await setCookie(bob.page, "__Host-gitd", sessions.alice);
  r = await bob.page.evaluate(async () => { const x = await fetch("/api/git/session"); return { status: x.status, text: await x.text() }; });
  if (r.status !== 200 || /alpha|Alpha|sub-alice/.test(r.text)) fail("bob, with alice's cookie, was served alice's session: " + r.text.slice(0, 100));
  r = await attack(alpha, { "X-CSRF-Token": alice.csrf });
  if (r.status !== 403 || /Alpha/.test(r.text)) fail("alice's cookie and CSRF token worked in bob's browser: " + r.status);
  state = await serverState();
  if (state.proposals.length !== 2) fail("an attack proposed something");
  // her session is gone on the server; she carries on without signing in again
  await click(alice.page, "git-refresh");
  await until(alice.page, "alice carries on", () => /Nothing changed|changes \(|rien/i.test(goSlidesDebug().git.note) || goSlidesDebug().git.note.includes("Pull request"));
  const sa2 = await gitState(alice.page);
  if (sa2.mode !== 2 || sa2.repos !== "github:lab/alpha") fail("alice after the theft: " + JSON.stringify(sa2));
  const sb2 = await gitState(bob.page);
  if (sb2.repos !== "github:lab/beta") fail("bob's pane after the attack: " + JSON.stringify(sb2));
  const cookieNow = (await allCookies(alice.page)).find((c) => c.name === "__Host-gitd");
  if (!cookieNow || cookieNow.value === sessions.alice) fail("alice kept the stolen session id");
  const sweepB = await sweep(bob.page);
  if (/Alpha|lab\/alpha|sub-alice|ALICE/.test(sweepB)) fail("something of alice's is in bob's page");
  const sweepA = await sweep(alice.page);
  if (/Beta deck|lab\/beta|sub-bob|BOB/.test(sweepA)) fail("something of bob's is in alice's page");
  out.isolation = "bob with his own session, her CSRF token, a forged identity header or her stolen cookie reads nothing of alice's; the theft destroys that session, alice re-establishes her own; no trace of one in the other's page";

  // 7. A sign-in that ends: Sign in again, no token in its place.
  await deleteCookie(alice.page, "test-user");
  await click(alice.page, "git-propose");
  const ended = await until(alice.page, "the sign-in ended", () => goSlidesDebug().git.mode === 1 && goSlidesDebug().git);
  if (ended.user !== "" || ended.repos !== "") fail("the signed-out pane keeps " + JSON.stringify(ended));
  rects = await alice.page.evaluate(() => Object.keys(goSlidesDebug().rects));
  if (!rects.includes("git-ssignin") || rects.includes("git-token") || rects.includes("git-connect")) fail("after the sign-in ended: " + rects.filter((k) => k.startsWith("git")));
  if (alice.requests.some((u) => /github|gitlab/.test(new URL(u).hostname))) fail("a forge was contacted after the sign-in ended");
  state = await serverState();
  if (state.proposals.length !== 2) fail("a proposal went through after the sign-in ended");
  out.ended = "an ended sign-in shows Sign in again, empties the pane, proposes nothing and falls back to no token";

  // 7b. The person's deck does not stay behind for the next person on the same
  // browser profile: after the sign-in ended, and for another person signing in.
  // the store deletes asynchronously: a few seconds, not forever
  const leftOver = async () => (await alice.page.evaluate(() => goSlidesDebug().source)).includes("Alpha") || /Alpha|lab\/alpha/.test(await sweep(alice.page, true));
  for (let i = 0; i < 40 && (await leftOver()); i++) await new Promise((r) => setTimeout(r, 250));
  if (await leftOver()) fail("alice's deck stays in the page or its storage after the sign-in ended");
  await setCookie(alice.page, "test-user", "sub-alice");
  await open(alice);
  await until(alice.page, "alice signed in again", () => goSlidesDebug().git.mode === 2);
  await typeInto(alice.page, "git-dir", "talk");
  await click(alice.page, "git-load");
  await until(alice.page, "alice's deck again", () => goSlidesDebug().source.includes("Alpha deck of Alice") || goSlidesDebug().source.includes("v2"));
  await new Promise((r) => setTimeout(r, 2500)); // the draft is stored once the page is quiet
  if (!/Alpha/.test(await sweep(alice.page, true))) fail("control: alice's deck should be stored in her profile");
  await setCookie(alice.page, "test-user", "sub-bob"); // bob sits down at the same browser profile
  await open(alice);
  await until(alice.page, "bob signed in on alice's profile", () => goSlidesDebug().git.mode === 2 && goSlidesDebug().git.user === "sub-bob");
  // the store deletes the blobs of the versions asynchronously: give it a few seconds, not forever
  // wait for the SAME predicate the check below asserts (the wipe clears the
  // deck, the repository names and the owner in separate asynchronous steps),
  // for a bounded time, then say which of them is left
  const aliceLeft = async () => {
    const swept = await sweep(alice.page, true);
    const src = await alice.page.evaluate(() => goSlidesDebug().source);
    const m = /Alpha|lab\/alpha|sub-alice/.exec(swept);
    if (m) return "storage: " + swept.slice(Math.max(0, swept.lastIndexOf("IDB ", m.index)), m.index + 40);
    return /Alpha/.test(src) ? "the editor's source" : "";
  };
  let left = await aliceLeft();
  for (const end = Date.now() + 30000; left && Date.now() < end; ) { await new Promise((r) => setTimeout(r, 250)); left = await aliceLeft(); }
  if (left) fail("bob, on alice's browser profile, got something of alice's: " + left);
  await setCookie(alice.page, "test-user", "sub-alice"); // alice comes back: her page is hers again, and empty of bob's
  out.sharedProfile = "an ended sign-in and another person on the same browser profile find nothing of the previous person's deck, draft, versions or state";

  // 7c. A front that is down: the page is served by slides-gitd (the marker), the
  // API answers 502. That is not "no server": no token field comes back.
  const down = await person(browser, "front-down");
  await down.page.setRequestInterception(true);
  down.page.removeAllListeners("request");
  down.page.on("request", (req) => {
    if (new URL(req.url()).pathname === "/api/git/session") return req.respond({ status: 502, contentType: "text/plain", body: "Bad Gateway" });
    req.continue();
  });
  await open(down);
  const sd = await until(down.page, "the unavailable pane", () => goSlidesDebug().git.mode === 3 && goSlidesDebug().git);
  if (!/unavailable/i.test(sd.note)) fail("the unavailable note: " + sd.note);
  const dr = await down.page.evaluate(() => Object.keys(goSlidesDebug().rects));
  for (const k of ["git-token", "git-remember", "git-connect", "git-forge", "git-signin"]) if (dr.includes(k)) fail("a page whose server is down shows " + k);
  out.down = "served by the front but the API answers 502: 'unavailable', no token field, no client mode";

  // 7d. Before the server says who is signed in, a page on a shared profile shows
  // and renders nothing of what is stored: with a slow session, and with a hung one.
  async function aliceStoresHerDeck() {
    await alice.page.setRequestInterception(true);
    alice.page.removeAllListeners("request");
    alice.page.on("request", (req) => req.continue());
    await setCookie(alice.page, "test-user", "sub-alice");
    await open(alice);
    await until(alice.page, "alice signed in", () => goSlidesDebug().git.mode === 2);
    await typeInto(alice.page, "git-dir", "talk");
    await click(alice.page, "git-load");
    await until(alice.page, "alice's deck", () => goSlidesDebug().source.includes("Alpha"));
    await new Promise((r) => setTimeout(r, 2500));
    if (!/Alpha/.test(await sweep(alice.page, true))) fail("control: alice's deck should be stored");
  }
  async function bobSitsDown(sessionHandler) {
    await setCookie(alice.page, "test-user", "sub-bob");
    alice.page.removeAllListeners("request");
    alice.page.on("request", (req) => {
      if (new URL(req.url()).pathname === "/api/git/session") return sessionHandler(req);
      req.continue();
    });
    await alice.page.goto(url, { waitUntil: "load" });
    await until(alice.page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
  }
  const nothingOfAlice = async (what) => {
    const src = await alice.page.evaluate(() => JSON.stringify(goSlidesDebug()));
    if (/Alpha|lab\/alpha|sub-alice/.test(src)) fail(what + ": alice's deck is in bob's editor or state");
    const all = await sweep(alice.page);
    const frames = all.split("\nDEBUG ")[0]; // the page's storage is not yet wiped: it is not what is looked at
    let seen = "";
    for (const f of alice.page.frames()) seen += await f.evaluate(() => (globalThis.__seen || []).join("\n")).catch(() => "");
    for (const w of alice.page.workers()) seen += await w.evaluate(() => (globalThis.__seen || []).join("\n")).catch(() => "");
    if (/Alpha/.test(seen)) fail(what + ": alice's deck was sent to the preview or the render worker");
  };
  // slow: the session answers after 5 s
  await aliceStoresHerDeck();
  let answeredAt = 0;
  await bobSitsDown((req) => setTimeout(() => { answeredAt = Date.now(); req.continue().catch(() => {}); }, 5000));
  let samples = 0;
  while (!answeredAt && samples < 40) {
    await nothingOfAlice("while the session is pending");
    samples++;
    await new Promise((r) => setTimeout(r, 150));
  }
  if (samples < 2) fail("control: the check never saw the pending state (" + samples + ")");
  const pend = await gitState(alice.page);
  await until(alice.page, "bob signed in after the delay", () => goSlidesDebug().git.mode === 2 && goSlidesDebug().git.user === "sub-bob");
  await nothingOfAlice("after bob's session");
  // hung: the session never answers; past the deadline the page says unavailable, and still nothing stored shows
  await aliceStoresHerDeck();
  await bobSitsDown(() => {}); // never answered
  const hung = await until(alice.page, "unavailable past the deadline", () => goSlidesDebug().git.mode === 3 && goSlidesDebug().git, undefined, 40000);
  if (!/unavailable/i.test(hung.note)) fail("hung session note: " + hung.note);
  await nothingOfAlice("with the session hung");
  out.heldUntilIdentified = "slow session (" + samples + " samples while pending) and hung session (unavailable past the deadline): nothing stored is shown, rendered or sent to the preview until the server names the person";

  // 8. The CSP still refuses another host and allows the page's own.
  const blocked = await alice.page.evaluate(async () => { try { await fetch("https://example.com/x"); return false; } catch (_) { return true; } });
  if (!blocked) fail("connect-src let the page fetch another host");
  out.csp = "connect-src unchanged: another host is refused";

  console.log("GITD_OK " + JSON.stringify(out));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
