// Drives the loading carousel in headless Chrome:
//   - on a throttled network it shows, rotates, and fades into the editor;
//   - a reader hovering it gets a Start button instead of the fade;
//   - under prefers-reduced-motion it does not rotate;
//   - a cached module loads without it ever showing.
//
//   node tour.cjs http://127.0.0.1:PORT/playground/ SHOTS_DIR
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. Prints one JSON
// line, TOUR_OK {...}.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const path = require("path");

const url = process.argv[2];
const shots = process.argv[3] || ".";
fs.mkdirSync(shots, { recursive: true });

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}

async function until(page, what, fn, ms = 60000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn).catch(() => null);
    if (v) return v;
    if (Date.now() > end) fail("timed out waiting for " + what + ": " + JSON.stringify(await page.evaluate(() => globalThis.goSlidesTour).catch(() => null)));
    await new Promise((r) => setTimeout(r, 50));
  }
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const tourState = (page) => page.evaluate(() => ({
  ...globalThis.goSlidesTour,
  visible: !document.getElementById("pg-tour").hidden,
  ready: document.querySelector(".stage").classList.contains("ready"),
  slide: [...document.querySelectorAll(".pg-card")].findIndex((s) => !s.hidden),
  picture: (() => { const s = [...document.querySelectorAll(".pg-card")].find((s) => !s.hidden); const im = s && s.querySelector("img"); return im ? im.complete && im.naturalWidth : 0; })(),
  rotation: document.getElementById("pg-tour-rot").getAttribute("aria-label"),
}));

// A fresh page, its cache off, on a slow network (~3 MB/s), optionally
// asking for reduced motion.
async function slowPage(browser, reduced) {
  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  await page.setCacheEnabled(false);
  const cdp = await page.createCDPSession();
  await cdp.send("Network.enable");
  await cdp.send("Network.emulateNetworkConditions", { offline: false, latency: 40, downloadThroughput: 3e6, uploadThroughput: 1e6 });
  if (reduced) await page.emulateMediaFeatures([{ name: "prefers-reduced-motion", value: "reduce" }]);
  page.on("pageerror", (e) => console.error("pageerror:", e.message));
  return page;
}

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const out = {};

  // 1. Throttled: shown, its picture loaded, rotating; then it fades away.
  {
    const page = await slowPage(browser, false);
    const t0 = Date.now();
    await page.goto(url, { waitUntil: "domcontentloaded" });
    await until(page, "the carousel", () => globalThis.goSlidesTour && globalThis.goSlidesTour.shown);
    out.shownAfterMs = Date.now() - t0;
    await until(page, "the first card's picture", () => { const im = document.querySelector(".pg-card:not([hidden]) img"); return im && im.complete && im.naturalWidth > 0; });
    await page.screenshot({ path: path.join(shots, "tour-1-shown.png") });
    await until(page, "the carousel to rotate", () => globalThis.goSlidesTour.advanced >= 1, 8000);
    // the second picture was preloaded while the first was shown
    await until(page, "the second card's picture", () => { const im = document.querySelector(".pg-card:not([hidden]) img"); return im && im.complete && im.naturalWidth > 0; }, 1500);
    const s = await tourState(page);
    if (!s.visible || s.slide !== 1 || !s.picture) fail("rotation: " + JSON.stringify(s));
    await page.screenshot({ path: path.join(shots, "tour-2-rotated.png") });
    await until(page, "the editor", () => document.querySelector(".stage").classList.contains("ready"));
    out.readyAfterMs = Date.now() - t0;
    const e = await tourState(page);
    if (e.startShown) fail("an idle reader got a Start button");
    await sleep(500);
    await page.screenshot({ path: path.join(shots, "tour-3-editor.png") });
    out.rotated = e.advanced;
    await page.close();
  }

  // 2. A reader in the carousel when the editor is ready: Start waits.
  {
    const page = await slowPage(browser, false);
    await page.goto(url, { waitUntil: "domcontentloaded" });
    await until(page, "the carousel", () => globalThis.goSlidesTour && globalThis.goSlidesTour.shown);
    await page.hover("#pg-tour");
    await page.click("#pg-tour-next");
    await until(page, "the Start button", () => globalThis.goSlidesTour.startShown);
    const s = await tourState(page);
    if (s.ready || s.advanced !== 0) fail("hovered: " + JSON.stringify(s));
    await page.screenshot({ path: path.join(shots, "tour-4-start.png") });
    await page.click("#pg-start");
    await until(page, "the editor after Start", () => document.querySelector(".stage").classList.contains("ready"), 5000);
    out.startButton = true;
    await page.close();
  }

  // 3. Reduced motion: no rotation.
  {
    const page = await slowPage(browser, true);
    await page.goto(url, { waitUntil: "domcontentloaded" });
    await until(page, "the carousel", () => globalThis.goSlidesTour && globalThis.goSlidesTour.shown);
    await sleep(6000);
    const s = await tourState(page);
    if (s.advanced !== 0 || s.slide !== 0 || s.rotation !== "Start slide rotation") fail("reduced motion rotated: " + JSON.stringify(s));
    out.reducedMotion = "no rotation";
    await page.close();
  }

  // 4. Cached: a first load fills the cache, the second never shows it.
  {
    const page = await browser.newPage();
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
    await page.goto(url, { waitUntil: "load" });
    await until(page, "the first editor", () => document.querySelector(".stage").classList.contains("ready"));
    const t0 = Date.now();
    await page.reload({ waitUntil: "load" });
    await until(page, "the cached editor", () => document.querySelector(".stage").classList.contains("ready"));
    out.cachedReadyMs = Date.now() - t0;
    const s = await tourState(page);
    out.cachedShown = s.shown;
    out.cachedTimings = await page.evaluate(() => globalThis.goSlidesTimings);
    if (s.shown) fail("the carousel showed on a cached load: " + JSON.stringify(out));
    await page.close();
  }

  console.log("TOUR_OK " + JSON.stringify(out));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
