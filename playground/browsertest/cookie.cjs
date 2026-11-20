// The server-mode session cookie in WebKit, Firefox and Chromium (go-slides#103).
//
// slides-gitd sets `__Host-gitd; Path=/; Secure; HttpOnly; SameSite=Strict`
// (gitd server/server.go) on GET /api/git/session. browsertest/gitd.cjs covers
// Chrome from http://127.0.0.1 only. This test serves a fake site over HTTPS
// and asks each engine two questions:
//
//   (a) is the cookie stored, kept out of document.cookie, and sent on ordinary
//       same-site requests (reload, fetch, POST);
//   (b) is it sent on the FIRST request after the sign-in chain coming back from
//       ANOTHER SITE (app.test -> /oauth2/sign_in -> idp.test -> /oauth2/callback
//       -> /page), and on the next one.
//
// Cross-site made real: app.test and idp.test are two registrable domains (a
// port would be the same site). Names resolve through a CONNECT proxy that
// every engine is told to use (Playwright's `proxy` option), which sends every
// host to the one local HTTPS server: no /etc/hosts edit, the same mechanism in
// WebKit (which has no host-resolver switch), Firefox and Chromium. The server
// logs the Cookie header of every request: the answer is what the SERVER saw.
//
//   NODE_PATH=<dir with playwright> node cookie.cjs [webkit firefox chromium]
//
// Needs `openssl` for the throwaway certificate. Prints COOKIE_OK {...}.
"use strict";
const https = require("https");
const http = require("http");
const net = require("net");
const fs = require("fs");
const os = require("os");
const path = require("path");
const { execFileSync } = require("child_process");
const pw = require("playwright");

const NAME = "__Host-gitd";
const APP = "app.test";
const IDP = "idp.test";

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}

// What each engine is expected to do. A change in an engine is a finding: fix
// the row and the comment beside it, never the assertion's direction.
// Keys: ordinary (a), then the first and the next request of (b).
const EXPECT = {
  // Measured (Playwright 1.65.1: WebKit 26.x, Firefox, Chromium), see the PR.
  // Common to all three: stored, HttpOnly, sent on every same-site request;
  // withheld when the idp's login form POSTs back (loginFirst/loginPage), on a
  // link clicked on the other site, and on a reload of the landing page; and
  // sent on the page's own first fetch after landing (chainNext, loginNext).
  // They differ on a 302 straight back (chain*): Chromium sends it, because the
  // redirect is not a new initiator; Firefox and WebKit withhold it.
  // WebKit also withholds it on a navigation made from the other site's page.
  chromium: { ordinary: true, chainFirst: true, chainPage: true, chainNext: true, loginFirst: false, loginPage: false, loginNext: true, reload: false, link: false, typed: true, bounceLanding: true, bouncePage: true, reloadAfterBounce: true },
  firefox: { ordinary: true, chainFirst: false, chainPage: false, chainNext: true, loginFirst: false, loginPage: false, loginNext: true, reload: false, link: false, typed: true, bounceLanding: false, bouncePage: true, reloadAfterBounce: true },
  webkit: { ordinary: true, chainFirst: false, chainPage: false, chainNext: true, loginFirst: false, loginPage: false, loginNext: true, reload: false, link: false, typed: false, bounceLanding: false, bouncePage: true, reloadAfterBounce: true },
};

function makeCert(dir) {
  const key = path.join(dir, "key.pem");
  const crt = path.join(dir, "crt.pem");
  execFileSync("openssl", ["req", "-x509", "-newkey", "ec", "-pkeyopt", "ec_paramgen_curve:prime256v1", "-nodes", "-keyout", key, "-out", crt, "-days", "2", "-subj", "/CN=" + APP, "-addext", `subjectAltName=DNS:${APP},DNS:${IDP}`], { stdio: "pipe" });
  return { key: fs.readFileSync(key), cert: fs.readFileSync(crt) };
}

const log = []; // {host, path, cookie, site}
function handler(req, res) {
  const host = (req.headers.host || "").split(":")[0];
  const u = new URL(req.url, "https://" + host);
  const cookie = req.headers.cookie || "";
  log.push({ host, path: u.pathname, has: cookie.includes(NAME + "="), site: req.headers["sec-fetch-site"] || "" });
  const html = (b) => { res.writeHead(200, { "Content-Type": "text/html", "Cache-Control": "no-store" }); res.end("<!doctype html><meta charset=utf-8><body>" + b); };
  const redirect = (to) => { res.writeHead(302, { Location: to, "Cache-Control": "no-store" }); res.end(); };
  if (host === APP) {
    switch (u.pathname) {
      case "/page": return html("page");
      case "/api/git/session":
        res.writeHead(200, { "Content-Type": "application/json", "Cache-Control": "no-store", "Set-Cookie": `${NAME}=sess1; Path=/; Secure; HttpOnly; SameSite=Strict` });
        return res.end('{"csrf":"x"}');
      case "/api/git/echo":
        res.writeHead(200, { "Content-Type": "application/json", "Cache-Control": "no-store" });
        return res.end(JSON.stringify({ has: cookie.includes(NAME + "=") }));
      case "/oauth2/sign_in": return redirect(`https://${IDP}/authorize${u.search}`);
      // oauth2-proxy: sets its own cookie (Lax) and sends the browser on
      case "/oauth2/callback": return redirect(u.searchParams.get("next") || "/page");
      // the workaround: a same-site page that navigates on by itself
      case "/bounce": return html("<script>location.replace('/page')</script>");
    }
  }
  if (host === IDP) {
    // already signed in at the provider: straight back; ?login=1: its login form first
    if (u.pathname === "/authorize") return u.searchParams.get("login") ? html(`<form method=post action="/login${u.search}"><button id=ok>Sign in</button></form>`) : redirect(`https://${APP}/oauth2/callback${u.search}`);
    if (u.pathname === "/login") return redirect(`https://${APP}/oauth2/callback${u.search}`);
    if (u.pathname === "/link") return html(`<a id=go href="https://${APP}/page">back</a>`);
  }
  res.writeHead(404); res.end("no");
}

function startProxy(httpsPort) {
  const p = http.createServer((q, r) => { r.writeHead(405); r.end(); });
  p.on("connect", (req, sock, head) => {
    const up = net.connect(httpsPort, "127.0.0.1", () => {
      sock.write("HTTP/1.1 200 Connection Established\r\n\r\n");
      if (head.length) up.write(head);
      up.pipe(sock); sock.pipe(up);
    });
    up.on("error", () => sock.destroy());
    sock.on("error", () => up.destroy());
  });
  return new Promise((ok) => p.listen(0, "127.0.0.1", () => ok(p)));
}

// requests logged after mark, at host+path
const since = (mark, host, p) => log.slice(mark).filter((e) => e.host === host && e.path === p);

async function runEngine(name, proxyPort) {
  const browser = await pw[name].launch({ proxy: { server: `http://127.0.0.1:${proxyPort}` } });
  const out = {};
  try {
    const ctx = await browser.newContext({ ignoreHTTPSErrors: true });
    const page = await ctx.newPage();
    const echo = () => page.evaluate(async () => (await (await fetch("/api/git/echo", { credentials: "same-origin" })).json()).has);

    // (a) stored, hidden from script, sent on same-site requests
    await page.goto(`https://${APP}/page`);
    await page.evaluate(() => fetch("/api/git/session").then((r) => r.json()));
    const jar = (await ctx.cookies(`https://${APP}/`)).find((c) => c.name === NAME);
    if (!jar) fail(name + ": __Host-gitd not stored");
    if (!jar.secure || !jar.httpOnly || jar.sameSite !== "Strict" || jar.path !== "/") fail(name + ": attributes changed " + JSON.stringify(jar));
    if ((await page.evaluate(() => document.cookie)).includes(NAME)) fail(name + ": cookie visible to script");
    let m = log.length;
    await page.reload();
    const reloaded = since(m, APP, "/page")[0];
    const fetched = await echo();
    const posted = await page.evaluate(async () => (await (await fetch("/api/git/echo", { method: "POST", body: "{}", headers: { "Content-Type": "application/json" } })).json()).has);
    out.ordinary = !!reloaded && reloaded.has && fetched && posted;

    // (b) the sign-in chain: app -> idp (other site) -> callback -> page. It
    // starts from the page, as the pane's Sign in does (location.href), so the
    // navigation is page-initiated and the final hop is initiated by the idp.
    const chain = async (key, q, login) => {
      await page.goto(`https://${APP}/page`);
      m = log.length;
      await page.evaluate((to) => { window.__old = true; location.href = to; }, "/oauth2/sign_in" + q);
      if (login) { await page.waitForURL(`https://${IDP}/authorize**`); await page.click("#ok"); }
      for (let i = 0; i < 400 && !(since(m, APP, "/oauth2/callback").length && since(m, APP, "/page").length); i++) await new Promise((r) => setTimeout(r, 25));
      if (!since(m, APP, "/page").length) fail(name + ": the " + key + " never came back to /page");
      await page.waitForFunction(() => !window.__old && document.readyState === "complete"); // the NEW document
      out[key + "First"] = since(m, APP, "/oauth2/callback")[0].has;
      out[key + "Page"] = since(m, APP, "/page")[0].has;
      out[key + "Next"] = await echo(); // the page's own first call: same-site
    };
    await chain("chain", "", false); // the idp still has its session: 302 straight back
    await chain("login", "?login=1", true); // the idp shows its login form: POST, then 302 back
    m = log.length;
    await page.reload();
    out.reload = since(m, APP, "/page")[0].has; // reloading the page that chain ended on
    // the workaround: the callback lands on a same-site page that navigates on
    await page.goto(`https://${APP}/page`);
    m = log.length;
    await page.evaluate(() => { location.href = "/oauth2/sign_in?next=/bounce"; });
    for (let i = 0; i < 400 && !since(m, APP, "/bounce").length; i++) await new Promise((r) => setTimeout(r, 25));
    for (let i = 0; i < 400 && !since(m, APP, "/page").length; i++) await new Promise((r) => setTimeout(r, 25));
    if (!since(m, APP, "/page").length) fail(name + ": the bounce never reached /page");
    await page.waitForFunction(() => document.readyState === "complete" && location.pathname === "/page");
    out.bounceLanding = since(m, APP, "/bounce")[0].has; // the landing page itself is still a cross-site first request
    out.bouncePage = since(m, APP, "/page")[0].has; // the page it navigates to is not
    m = log.length;
    await page.reload();
    out.reloadAfterBounce = since(m, APP, "/page")[0].has; // the bounce's landing is same-site: a reload is fine

    // a link clicked on the other site (no redirect), and a navigation Playwright
    // makes from the other site's page (what an address typed in the bar would do)
    await page.goto(`https://${IDP}/link`);
    m = log.length;
    await Promise.all([page.waitForURL(`https://${APP}/page`), page.click("#go")]);
    out.link = since(m, APP, "/page")[0].has;
    await page.goto(`https://${IDP}/link`);
    m = log.length;
    await page.goto(`https://${APP}/page`);
    out.typed = since(m, APP, "/page")[0].has;
    await ctx.close();
  } finally {
    await browser.close();
  }
  return out;
}

(async () => {
  const names = process.argv.slice(2).length ? process.argv.slice(2) : ["webkit", "firefox", "chromium"];
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "gitd-cookie-"));
  const srv = https.createServer(makeCert(dir), handler);
  await new Promise((ok) => srv.listen(0, "127.0.0.1", ok));
  const proxy = await startProxy(srv.address().port);
  const result = {};
  let bad = 0;
  for (const n of names) {
    if (!EXPECT[n]) fail("unknown engine " + n);
    const got = await runEngine(n, proxy.address().port);
    result[n] = got;
    for (const k of Object.keys(EXPECT[n])) {
      if (got[k] !== EXPECT[n][k]) { console.error(`MISMATCH ${n}.${k}: sent=${got[k]}, expected ${EXPECT[n][k]}`); bad++; }
    }
    console.log(n, JSON.stringify(got));
  }
  proxy.close(); srv.close(); fs.rmSync(dir, { recursive: true, force: true });
  if (bad) fail(bad + " mismatch(es) against the table above");
  console.log("COOKIE_OK " + JSON.stringify(result));
  process.exit(0);
})().catch((e) => fail(e.stack || String(e)));
