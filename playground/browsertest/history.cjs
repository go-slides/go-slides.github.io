// Undo, Redo and the versions in headless Chrome: Ctrl+Z and the ribbon's
// buttons take an edit back and forth; Reset, a whole-project step, is undone
// as one; Save version… keeps a named version, the Versions pane compares it
// with now (a DiffView, F7) and restores it, and Undo takes the restore back;
// the comparison and the Home tab are pictured in both schemes. Prints
// HISTORY_OK {...}.
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
  page.on("dialog", (d) => (d.type() === "prompt" ? d.accept("Before the review") : d.accept()));
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

  // Save version… (named by the prompt), change the deck, compare, restore
  const named = await source(page);
  await click(page, "save-version");
  await until(page, "the named version", () => (goSlidesDebug().status || "").includes("Before the review"));
  await until(page, "the Versions pane", () => goSlidesDebug().rects["versions-list"]);
  const lines2 = named.split("\n");
  const hl2 = lines2.findIndex((l) => l.startsWith("# "));
  const [x2, y2] = await page.evaluate((l, c) => goSlidesCaretPixel(l, c), hl2, lines2[hl2].length);
  await page.mouse.click(x2 + 2, y2 + 6);
  await page.keyboard.type(" (draft 2)");
  await until(page, "the change", (s) => goSlidesDebug().source !== s, named);
  const changed = await source(page);
  // the newest version is the named one: the first row of the list
  const list = (await debug(page)).rects["versions-list"];
  await page.mouse.click(list[0] + 40, list[1] + 8);
  await click(page, "version-compare");
  await until(page, "the comparison", () => goSlidesDebug().rects["compare-diff"]);
  out.compare = "a DiffView of deck.md";
  // inline (the default in the editor's column), then side by side
  for (const layout of ["inline", "side"]) {
    if (layout === "side") await click(page, "compare-layout");
    for (const scheme of ["light", "dark"]) {
      await page.emulateMediaFeatures([{ name: "prefers-color-scheme", value: scheme }]);
      await new Promise((r) => setTimeout(r, 400));
      await page.screenshot({ path: path.join(shots, "versions-compare-" + layout + "-" + scheme + ".png") });
    }
  }
  await page.emulateMediaFeatures([{ name: "prefers-color-scheme", value: "light" }]);
  // the keyboard moves between changes in the diff
  const diff = (await debug(page)).rects["compare-diff"];
  await page.mouse.click(diff[0] + diff[2] / 2, diff[1] + diff[3] / 2);
  await page.keyboard.press("F7");
  await click(page, "compare-back");
  await until(page, "the editor back", () => !goSlidesDebug().rects["compare-diff"]);
  await page.mouse.click(list[0] + 40, list[1] + 8);
  await click(page, "version-restore");
  await until(page, "the restore", (s) => goSlidesDebug().source === s, named);
  await chord(page, "z");
  await until(page, "Undo of the restore", (s) => goSlidesDebug().source === s, changed);
  out.restore = "restored, and the restore undone";

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
