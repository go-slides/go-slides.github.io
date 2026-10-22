// Drives the draft keeping in headless Chrome:
//   - typed text survives a reload (saved after the typing stops), and a
//     reload right after typing (saved on pagehide);
//   - a second tab's save is announced in the first, which does not erase it;
//   - a shared link wins and the draft stays restorable;
//   - Forget my drafts erases it: the next load is the example.
//
//   node draft.cjs http://127.0.0.1:PORT/playground/ SHOTS_DIR
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. Prints one JSON
// line, DRAFT_OK {...}.
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
    if (Date.now() > end) fail("timed out waiting for " + what + ": " + JSON.stringify(await page.evaluate(() => globalThis.goSlidesDebug && { status: goSlidesDebug().status, draft: goSlidesDebug().draftItems, source: goSlidesDebug().source.slice(0, 120) }).catch(() => null)));
    await new Promise((r) => setTimeout(r, 50));
  }
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const ready = (page) => until(page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready") && !goSlidesDebug().busy);

async function typeAtHeading(page, text) {
  const d = await page.evaluate(() => goSlidesDebug());
  const lines = d.source.split("\n");
  const hl = lines.findIndex((l) => l.startsWith("# "));
  const [x, y] = await page.evaluate((l, c) => goSlidesCaretPixel(l, c), hl, lines[hl].length);
  await page.mouse.click(x + 2, y + 6);
  await page.keyboard.type(text);
  await until(page, "the typed text", (t) => goSlidesDebug().source.includes(t), text);
}

async function pick(page, prefix) {
  const d = await page.evaluate(() => goSlidesDebug());
  const items = d.draftItems.split("|");
  const i = items.findIndex((s) => s.startsWith(prefix));
  if (i < 0) fail("no Draft item " + prefix + " in " + d.draftItems);
  const r = d.rects.draft;
  await page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2);
  const pop = await until(page, "the Draft menu", () => goSlidesDebug().rects.popover);
  await page.mouse.click(pop[0] + pop[2] / 2, pop[1] + ((i + 0.5) * pop[3]) / items.length);
}

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const out = {};
  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  page.on("pageerror", (e) => console.error("pageerror:", e.message));
  page.on("dialog", (d) => d.accept());
  await page.goto(url, { waitUntil: "load" });
  await ready(page);

  // 1. Typed, saved after the pause, back after a reload.
  await typeAtHeading(page, " kept");
  await sleep(900);
  await page.reload({ waitUntil: "load" });
  await ready(page);
  const s1 = await page.evaluate(() => goSlidesDebug());
  if (!/ kept/.test(s1.source) || !/Restored your draft/.test(s1.status)) fail("after a reload: " + JSON.stringify({ status: s1.status }));
  out.reload = "kept";
  await page.screenshot({ path: path.join(shots, "draft-1-restored.png") });

  // 2. Typed and reloaded at once: pagehide saved it.
  await typeAtHeading(page, " fast");
  await page.reload({ waitUntil: "load" });
  await ready(page);
  if (!/ kept fast/.test(await page.evaluate(() => goSlidesDebug().source))) fail("a reload right after typing lost the text");
  out.pagehide = "kept";

  // 3. A second tab saves: the first says so, and its next save keeps the
  //    other tab's draft restorable.
  const other = await browser.newPage();
  await other.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  other.on("dialog", (d) => d.accept());
  await other.goto(url, { waitUntil: "load" });
  await ready(other);
  await typeAtHeading(other, " other");
  await sleep(900);
  await until(page, "the other tab's save announced", () => /Another tab saved its draft/.test(goSlidesDebug().status));
  await page.bringToFront();
  await typeAtHeading(page, " first");
  await sleep(900);
  await until(page, "Restore replaced in the menu", () => goSlidesDebug().draftItems.includes("Restore replaced"));
  await pick(page, "Restore replaced");
  await until(page, "the other tab's text back", () => / other/.test(goSlidesDebug().source));
  out.tabs = "announced, kept, restored";
  await other.close();

  // 4. A shared link wins; the draft is offered back.
  // a link to a deck that is not the draft: Share writes the fragment
  await page.evaluate(() => goSlidesSetSource("---\nslides: true\n---\n\n# Shared deck\n"));
  const sh = (await page.evaluate(() => goSlidesDebug())).rects.share;
  await page.mouse.click(sh[0] + sh[2] / 2, sh[1] + sh[3] / 2);
  const link = await until(page, "the share link", () => location.hash.length > 10 && location.href);
  await page.evaluate(() => goSlidesSetSource("---\nslides: true\n---\n\n# My draft again\n"));
  await sleep(900);
  await page.goto(link, { waitUntil: "load" });
  await page.reload({ waitUntil: "load" });
  await ready(page);
  const s4 = await page.evaluate(() => goSlidesDebug());
  if (!/# Shared deck/.test(s4.source) || !/Your draft of .* is kept/.test(s4.status)) fail("shared link: " + JSON.stringify({ status: s4.status }));
  await pick(page, "Restore my draft");
  await until(page, "the draft back", () => /# My draft again/.test(goSlidesDebug().source));
  out.sharedLink = "wins, draft restorable";

  // 5. Forget: nothing stored, the next load is the example.
  await pick(page, "Forget my drafts");
  await until(page, "forgotten", () => /Forgot your drafts/.test(goSlidesDebug().status));
  await page.goto(url.replace(/#.*$/, ""), { waitUntil: "load" });
  await ready(page);
  const s5 = await page.evaluate(() => new Promise((res) => {
    const source = goSlidesDebug().source;
    const req = indexedDB.open("go-slides-playground");
    req.onsuccess = () => {
      const keys = req.result.transaction("projects").objectStore("projects").getAllKeys();
      keys.onsuccess = () => res({ source, projects: keys.result });
    };
  }));
  // only the UI state may be back (the preview mode, saved as the page started)
  if (!/# Hello, \{\{ \.name \}\}!/.test(s5.source) || s5.projects.some((k) => k !== "ui")) fail("after Forget: " + JSON.stringify(s5));
  out.forget = "erased";

  console.log("DRAFT_OK " + JSON.stringify(out));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
