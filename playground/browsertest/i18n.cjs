// Loads the playground in a language and takes pictures of what a translation
// can break: the Home and Format tabs, the help pane and the file panel (they
// are all on screen at 1440 px), and the ribbon at the narrowest width the
// page lays out (1100 px), where its groups fold. The text is checked too: the
// language the wasm speaks, the ribbon's tab labels, and no pageerror.
//
//   node i18n.cjs http://127.0.0.1:PORT/fr/playground/ fr SHOTS_DIR
//
// CHROME names the browser; NODE_PATH finds puppeteer-core. Pictures go
// outside every repository (the caller passes the directory).
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const path = require("path");

// node i18n.cjs BASE SHOTS_DIR lang...   (BASE: http://127.0.0.1:PORT, no slash)
const [base, shots, ...langs] = process.argv.slice(2);
fs.mkdirSync(shots, { recursive: true });
let lang = "";

function fail(msg) {
  console.error("FAIL " + lang + ": " + msg);
  process.exit(1);
}

async function one(l) {
  lang = l;
  const url = base + (l === "en" ? "" : "/" + l) + "/playground/";
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox"] });
  const page = await browser.newPage();
  const errors = [];
  page.on("pageerror", (e) => errors.push(e.message));
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  await page.goto(url, { waitUntil: "load" });
  await page.waitForFunction(
    () => globalThis.webcanvasReady && globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"),
    { timeout: 120000 },
  );
  await page.waitForFunction(() => goSlidesDebug().loads >= 1, { timeout: 60000 });
  const info = await page.evaluate(() => ({
    lang: globalThis.goSlidesLang, htmlLang: document.documentElement.lang,
    cjk: globalThis.goSlidesCJK ? globalThis.goSlidesCJK.length : 0,
    // the tabs, left to right (the rectangles are a map)
    tabs: Object.entries(goSlidesDebug().rects).filter(([k]) => k.startsWith("tab-")).sort((a, b) => a[1][0] - b[1][0]).map(([k]) => k),
    status: goSlidesDebug().status,
  }));
  if (info.lang !== lang) fail("goSlidesLang is " + info.lang);
  console.log(JSON.stringify(info));

  const tabs = info.tabs;
  const click = async (key) => {
    const r = await page.evaluate((k) => goSlidesDebug().rects[k], key);
    if (!r) fail("no rect " + key);
    await page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2);
    await new Promise((r) => setTimeout(r, 400));
  };
  const shot = async (name) => page.screenshot({ path: path.join(shots, lang + "-" + name + ".png") });

  // the help pane starts off at this width: the View tab switches it on
  await click(tabs[3]);
  if (!(await page.evaluate(() => goSlidesDebug().rects.helpToggle))) fail("no Help toggle on the View tab");
  await click("helpToggle");
  await shot("view");
  await click(tabs[0]);
  await shot("home"); // the default tab, the file panel left, the help pane right
  await click(tabs[2]); // Format
  await shot("format");
  await click(tabs[1]); // Insert
  await shot("insert");
  await click(tabs[0]);
  await page.setViewport({ width: 1100, height: 800, deviceScaleFactor: 1 });
  await new Promise((r) => setTimeout(r, 800));
  await shot("home-narrow");
  await click(tabs[2]);
  await shot("format-narrow");
  if (errors.length) fail("pageerror: " + errors.join("; "));
  await browser.close();
}

(async () => {
  for (const l of langs) await one(l);
})().catch((e) => fail(String(e)));
