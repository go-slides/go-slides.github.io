// What the dependencies say in the page's language: the file panel's context
// menu (the toolkit's words), the status line and the problems list (with the
// explanation of a diagnostic's code), and the help of the directive under the
// caret (the editor package's).
//
//   node i18n2.cjs http://127.0.0.1:PORT SHOTS_DIR fr ja ...
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. Pictures go
// outside every repository.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const path = require("path");

const [base, shots, ...langs] = process.argv.slice(2);
fs.mkdirSync(shots, { recursive: true });
let lang = "";
const fail = (m) => { console.error("FAIL " + lang + ": " + m); process.exit(1); };
const wait = (ms) => new Promise((r) => setTimeout(r, ms));

const deck = "---\nslides: true\npaginate: true\ntheme: gaia\n---\n\n<!-- _kind: nosuch -->\n# Title\n";

async function one(l) {
  lang = l;
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox"] });
  const page = await browser.newPage();
  const errors = [];
  page.on("pageerror", (e) => errors.push(e.message));
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  await page.goto(base + (l === "en" ? "" : "/" + l) + "/playground/", { waitUntil: "load" });
  await page.waitForFunction(() => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"), { timeout: 120000 });
  await page.waitForFunction(() => goSlidesDebug().loads >= 1, { timeout: 60000 });
  const rect = (k) => page.evaluate((k) => goSlidesDebug().rects[k], k);
  const click = async (r, opt) => { await page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2, opt); await wait(400); };
  // the help pane is off at this width: the View tab switches it on
  const tabs = await page.evaluate(() => Object.entries(goSlidesDebug().rects).filter(([k]) => k.startsWith("tab-")).sort((a, b) => a[1][0] - b[1][0]).map(([k]) => k));
  await click(await rect(tabs[3]));
  await click(await rect("helpToggle"));
  await click(await rect(tabs[0]));
  // a deck with a problem, and the caret on a directive: its help in the pane
  await page.evaluate((d) => goSlidesSetSource(d), deck);
  await wait(1500);
  const p = await page.evaluate(() => goSlidesCaretPixel(2, 3)); // on "paginate"
  await page.mouse.click(p[0], p[1]);
  await wait(800);
  await page.screenshot({ path: path.join(shots, l + "-help-and-problems.png") });
  // the file panel's context menu on its first folder
  const folder = { x: 60, y: 258 };
  await page.mouse.click(folder.x, folder.y, { button: "right" });
  await wait(600);
  await page.screenshot({ path: path.join(shots, l + "-filetree-menu.png") });
  await page.keyboard.press("Escape");
  // a warning in the status line: renaming the deck is refused ("⚠" is an icon there)
  await page.mouse.click(80, 311);
  await wait(300);
  await page.keyboard.press("F2");
  await wait(300);
  await page.keyboard.type("other.md");
  await page.keyboard.press("Enter");
  await wait(800);
  await page.screenshot({ path: path.join(shots, l + "-status-warning.png") });
  const info = await page.evaluate(() => ({ status: goSlidesDebug().status, problems: goSlidesDebug().problems }));
  console.log(l, JSON.stringify(info));
  if (errors.length) fail("pageerror: " + errors.join("; "));
  await browser.close();
}

(async () => { for (const l of langs) await one(l); })().catch((e) => fail(String(e)));
