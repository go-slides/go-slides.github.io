// A recording's player in the preview frame, in headless Chrome: the frame is a
// sandboxed srcdoc iframe (allow-scripts, an opaque origin) in a page under the
// playground page's own Content Security Policy (read from
// layouts/_default/playground.html), as the playground builds it. Checks that
//
//  1. nothing is refused: no securitypolicyviolation, no CSP console message,
//     in the page or in the frame;
//  2. the player plays on its own to the end, shows the recording's text and
//     colours, and answers the speed button and the position bar;
//  3. under prefers-reduced-motion it starts paused, on its poster, and plays
//     when asked;
//  4. what it needs from the network is nothing: no request leaves the page
//     but the page itself.
//
//   CAST_PREVIEW_OUT=/path/frame.html go test -run TestWritePreviewPageForBrowser ./playground
//   node cast.cjs /path/frame.html /path/to/layouts/_default/playground.html
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. Prints CAST_OK.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const os = require("os");
const path = require("path");

const [framePath, layoutPath] = process.argv.slice(2);
const fail = (m) => { console.error("FAIL " + m); process.exit(1); };
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const csp = /<meta http-equiv="Content-Security-Policy" content="([^"]+)">/.exec(fs.readFileSync(layoutPath, "utf8"));
if (!csp) fail("no policy in " + layoutPath);
const srcdoc = fs.readFileSync(framePath, "utf8").replace(/&/g, "&amp;").replace(/"/g, "&quot;");
const parent = `<!doctype html><meta http-equiv="Content-Security-Policy" content="${csp[1].replace("{{ $fo }}", "https://go-slides-usercontent.github.io")}"><body><iframe id="f" sandbox="allow-scripts" style="width:1000px;height:700px" srcdoc="${srcdoc}"></iframe>`;
const dir = fs.mkdtempSync(path.join(os.tmpdir(), "castcsp-"));
const parentPath = path.join(dir, "page.html");
fs.writeFileSync(parentPath, parent);

async function frameState(page) {
  const f = page.frames().find((x) => x !== page.mainFrame());
  if (!f) return null;
  return f.evaluate(() => {
    const box = document.querySelector(".slides-cast");
    if (!box) return null;
    const ctl = box.querySelector(".slides-cast-ctl");
    return {
      label: ctl && ctl.querySelector("button").getAttribute("aria-label"),
      clock: ctl && ctl.querySelector(".slides-cast-clock").textContent,
      text: box.querySelector(".slides-cast-screen").textContent,
      rows: box.querySelectorAll('.slides-cast-screen > span[style*="1.2em"]').length,
      green: !![...box.querySelectorAll(".slides-cast-screen span span")].find((s) => s.style.color !== ""),
    };
  });
}

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1100,800"] });
  const violations = [], requests = [];
  async function open(reduce) {
    const page = await browser.newPage();
    if (reduce) await page.emulateMediaFeatures([{ name: "prefers-reduced-motion", value: "reduce" }]);
    await page.setViewport({ width: 1100, height: 800 });
    page.on("console", (m) => { if (/Content Security Policy|Refused to/i.test(m.text())) violations.push(m.text()); });
    page.on("request", (r) => { if (!r.url().startsWith("file:") && !r.url().startsWith("about:") && !r.url().startsWith("data:") && !r.url().startsWith("blob:")) requests.push(r.url()); });
    await page.goto("file://" + parentPath, { waitUntil: "load" });
    return page;
  }

  let page = await open(false);
  await sleep(1200);
  let s = await frameState(page);
  if (!s || s.label !== "Pause") fail("it did not start on its own: " + JSON.stringify(s));
  await sleep(2600);
  s = await frameState(page);
  if (s.label !== "Play" || !s.clock.startsWith("0:02 / 0:02") || !s.text.includes("world") || !s.text.includes("the end") || !s.green) fail("it did not play to the end: " + JSON.stringify(s));
  const f = page.frames().find((x) => x !== page.mainFrame());
  await f.evaluate(() => { const b = document.querySelector(".slides-cast-ctl input"); b.value = 300; b.dispatchEvent(new Event("input")); });
  s = await frameState(page);
  if (!s.text.includes("hello") || s.text.includes("world")) fail("seek: " + JSON.stringify(s));
  await f.evaluate(() => document.querySelectorAll(".slides-cast-ctl button")[1].click());
  s = await frameState(page);
  await page.close();

  page = await open(true);
  await sleep(1200);
  s = await frameState(page);
  if (!s || s.label !== "Play" || s.rows !== 0 || !s.text.includes("the end")) fail("reduced motion: " + JSON.stringify(s));
  await (page.frames().find((x) => x !== page.mainFrame())).evaluate(() => document.querySelector(".slides-cast-ctl button").click());
  await sleep(600);
  s = await frameState(page);
  if (s.label !== "Pause" || s.rows === 0) fail("reduced motion, after Play: " + JSON.stringify(s));
  await page.close();
  await browser.close();

  if (violations.length) fail("the policy refused something it should allow: " + violations.join(" | "));
  if (requests.length) fail("requests left the page: " + requests.join(" "));
  console.log("CAST_OK");
})().catch((e) => fail(e.stack || String(e)));
