// A video in the preview frame, in headless Chrome: the frame is a sandboxed
// srcdoc iframe (allow-scripts, an opaque origin) in a page under the
// playground page's own Content Security Policy (read from
// layouts/_default/playground.html), as the playground builds it. The page
// posts the project's file into the frame ({goSlidesMedia}), as the playground
// does. Checks that
//
//  1. the video's src becomes a blob: URL made from the file's bytes, the
//     poster (a data: URI, decoded in Go) stays;
//  2. nothing is refused: no securitypolicyviolation, no CSP console message
//     (media-src blob:), in the page or in the frame;
//  3. the browser loads the file when it can decode H.264 (Chrome does,
//     Chromium builds often do not: then only the URL is checked);
//  4. replacing the file revokes the previous blob: URL and gives the video
//     the new one; removing it gives the path back and revokes the URL;
//  5. no request leaves the page: the video's bytes never leave the browser.
//
//   VIDEO_PREVIEW_OUT=/dir go test -run TestWriteVideoPreviewForBrowser ./playground
//   node video.cjs /dir/frame.html /dir/clip.mp4 /path/to/layouts/_default/playground.html
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. Prints VIDEO_OK.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const os = require("os");
const path = require("path");

const [framePath, clipPath, layoutPath] = process.argv.slice(2);
const fail = (m) => { console.error("FAIL " + m); process.exit(1); };
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const csp = /<meta http-equiv="Content-Security-Policy" content="([^"]+)">/.exec(fs.readFileSync(layoutPath, "utf8"));
if (!csp) fail("no policy in " + layoutPath);
const srcdoc = fs.readFileSync(framePath, "utf8").replace(/&/g, "&amp;").replace(/"/g, "&quot;");
const parent = `<!doctype html><meta http-equiv="Content-Security-Policy" content="${csp[1]}"><body><iframe id="f" sandbox="allow-scripts" style="width:1000px;height:700px" srcdoc="${srcdoc}"></iframe>`;
const dir = fs.mkdtempSync(path.join(os.tmpdir(), "videocsp-"));
const parentPath = path.join(dir, "page.html");
fs.writeFileSync(parentPath, parent);
const clip = fs.readFileSync(clipPath);

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1100,800"] });
  const violations = [], requests = [];
  const page = await browser.newPage();
  await page.setViewport({ width: 1100, height: 800 });
  page.on("console", (m) => { if (/Content Security Policy|Refused to/i.test(m.text())) violations.push(m.text()); });
  page.on("request", (r) => { if (!r.url().startsWith("file:") && !r.url().startsWith("about:") && !r.url().startsWith("data:") && !r.url().startsWith("blob:")) requests.push(r.url()); });
  await page.goto("file://" + parentPath, { waitUntil: "load" });
  const frame = () => page.frames().find((x) => x !== page.mainFrame());
  const state = () => frame().evaluate(() => {
    const v = document.querySelector("video[data-slides-video]");
    return v && { src: v.getAttribute("src"), poster: (v.getAttribute("poster") || "").slice(0, 23), state: v.readyState, error: v.error && v.error.code, h264: !!v.canPlayType('video/mp4; codecs="avc1.42E01E"') };
  });
  // posts the project's media into the frame and waits for its answer
  const post = (media) => page.evaluate(async (media) => {
    const f = document.getElementById("f");
    const shown = new Promise((res) => addEventListener("message", (e) => { if (e.data && typeof e.data.goSlidesMediaShown === "number") res(e.data.goSlidesMediaShown); }, { once: true }));
    const m = {};
    for (const [k, b] of Object.entries(media)) m[k] = { type: "video/mp4", bytes: new Uint8Array(b).buffer };
    f.contentWindow.postMessage({ goSlidesMedia: m }, "*");
    return shown;
  }, Object.fromEntries(Object.entries(media).map(([k, b]) => [k, Array.from(b)])));

  let s = await state();
  if (!s || s.src !== "media/clip.mp4") fail("before the media: " + JSON.stringify(s));
  if (s.poster !== "data:image/jpeg;base64,") fail("no poster: " + JSON.stringify(s));

  if (await post({ "media/clip.mp4": clip }) !== 1) fail("the frame did not take the file");
  s = await state();
  const m1 = /^(blob:[^#]+)$/.exec(s.src);
  if (!m1) fail("the src is not a blob: URL of the file: " + JSON.stringify(s));
  await sleep(800);
  s = await state();
  if (s.h264 && (s.error || s.state < 1)) fail("the browser could not load the file: " + JSON.stringify(s));

  // the same path again with other bytes: a new URL, the old one revoked
  const old = m1[1];
  const other = Buffer.concat([clip, Buffer.from("x")]);
  if (await post({ "media/clip.mp4": other }) !== 1) fail("the frame did not take the second file");
  s = await state();
  const m2 = /^(blob:[^#]+)$/.exec(s.src);
  if (!m2 || m2[1] === old) fail("the video kept the replaced URL: " + s.src);
  // a revoked URL loads nothing (probed with a video element: media-src allows
  // blob:, connect-src does not, so fetch would be refused either way)
  if (s.h264) {
    const loads = await frame().evaluate((u) => new Promise((res) => {
      const v = document.createElement("video");
      v.onloadedmetadata = () => res(true);
      v.onerror = () => res(false);
      v.src = u;
      setTimeout(() => res("timeout"), 4000);
    }), old);
    if (loads !== false) fail("the replaced blob: URL was not revoked (" + loads + ")");
  }

  // the file leaves the project: the path is given back, the URL revoked
  if (await post({}) !== 0) fail("the frame kept a medium");
  s = await state();
  if (s.src !== "media/clip.mp4") fail("the path was not given back: " + s.src);

  await page.close();
  await browser.close();
  if (violations.length) fail("the policy refused something it should allow: " + violations.join(" | "));
  if (requests.length) fail("requests left the page: " + requests.join(" "));
  console.log("VIDEO_OK");
})().catch((e) => fail(e.stack || String(e)));
