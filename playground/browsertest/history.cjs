// Undo, Redo and the versions in headless Chrome: Ctrl+Z and the ribbon's
// buttons take an edit back and forth; Reset, a whole-project step, is undone
// as one; the Home tab is pictured in both schemes. Prints HISTORY_OK {...}.
//
//   node history.cjs http://127.0.0.1:PORT/playground/ SHOTS_DIR
//
// CHROME names the browser; NODE_PATH finds puppeteer-core.
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

async function until(page, what, fn, arg, ms = 30000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg).catch(() => null);
    if (v) return v;
    if (Date.now() > end) fail("timed out waiting for " + what + ": " + (await page.evaluate(() => goSlidesDebug().status).catch(() => "?")));
    await new Promise((r) => setTimeout(r, 25));
  }
}

const debug = (page) => page.evaluate(() => goSlidesDebug());
const source = async (page) => (await debug(page)).source;

async function click(page, key) {
  const r = await until(page, key, (k) => goSlidesDebug().rects[k], key);
  await page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2);
}

async function chord(page, key, shift) {
  await page.keyboard.down("Control");
  if (shift) await page.keyboard.down("Shift");
  await page.keyboard.press(key);
  if (shift) await page.keyboard.up("Shift");
  await page.keyboard.up("Control");
}

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const page = await browser.newPage();
  page.on("pageerror", (e) => fail("pageerror: " + e.message));
  page.on("dialog", (d) => d.accept(d.defaultValue() || undefined));
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  await page.goto(url, { waitUntil: "load" });
  await until(page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
  await until(page, "the first preview", () => goSlidesDebug().loads >= 1);
  const out = {};

  // type one character at the end of the heading, then take it back and forth
  const start = await source(page);
  const lines = start.split("\n");
  const hl = lines.findIndex((l) => l.startsWith("# "));
  const [x, y] = await page.evaluate((l, c) => goSlidesCaretPixel(l, c), hl, lines[hl].length);
  await page.mouse.click(x + 2, y + 6);
  await page.keyboard.type("Q");
  await until(page, "the typed Q", (s) => goSlidesDebug().source !== s, start);
  await chord(page, "z");
  await until(page, "Ctrl+Z", (s) => goSlidesDebug().source === s, start);
  await chord(page, "z", true);
  await until(page, "Ctrl+Shift+Z", (s) => goSlidesDebug().source !== s, start);
  await click(page, "undo");
  await until(page, "the Undo button", (s) => goSlidesDebug().source === s, start);
  await click(page, "redo");
  await until(page, "the Redo button", (s) => goSlidesDebug().source.includes("Q"), start);
  out.edit = "Ctrl+Z, Ctrl+Shift+Z, Undo, Redo";

  // Reset replaces the deck (a step of the whole project): one Undo brings it back
  const typed = await source(page);
  await click(page, "reset");
  await until(page, "Reset", (s) => goSlidesDebug().source !== s, typed);
  await chord(page, "z");
  await until(page, "Undo of Reset", (s) => goSlidesDebug().source === s, typed);
  out.project = "Reset undone in one step";

  // the Home tab in both schemes, the pointer on Redo: its tip names the step
  const redo = (await debug(page)).rects["redo"];
  for (const scheme of ["light", "dark"]) {
    await page.emulateMediaFeatures([{ name: "prefers-color-scheme", value: scheme }]);
    await page.mouse.move(0, 400);
    await new Promise((r) => setTimeout(r, 300));
    await page.mouse.move(redo[0] + redo[2] / 2, redo[1] + redo[3] / 2);
    await new Promise((r) => setTimeout(r, 900));
    await page.screenshot({ path: path.join(shots, "history-home-" + scheme + ".png"), clip: { x: 0, y: 60, width: 900, height: 200 } });
  }

  console.log("HISTORY_OK " + JSON.stringify(out));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
