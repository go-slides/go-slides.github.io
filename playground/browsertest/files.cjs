// Drives the file panel in headless Chrome: a theme made from the Project
// menu and chosen in the Theme list colours the slides; a kind made there
// is used by a slide; an image dropped on the editor lands in media/ and
// shows as a background in the sandboxed preview (a blob: URL the frame made
// itself); the project downloads as a .zip and opens again from it. Then
// the cost of 20 MiB of media: an edit's time to the preview, with and
// without them. Prints one JSON line.
//
//   node files.cjs http://127.0.0.1:PORT/playground/ SHOTS_DIR
//
// CHROME names the browser; NODE_PATH finds puppeteer-core.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const path = require("path");

const url = process.argv[2];
const shots = process.argv[3] || ".";
fs.mkdirSync(shots, { recursive: true });
const downloads = fs.mkdtempSync(path.join(shots, "dl-"));

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}

async function until(page, what, fn, ms = 30000, arg) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg);
    if (v) return v;
    if (Date.now() > end) {
      const d = await page.evaluate(() => goSlidesDebug()).catch(() => null);
      if (d) d.source = d.source.split("\n").slice(0, 12).join("\n");
      fail("timed out waiting for " + what + "\n" + JSON.stringify(d, null, 1));
    }
    await new Promise((r) => setTimeout(r, 25));
  }
}

const debug = (page) => page.evaluate(() => goSlidesDebug());
const center = (r) => [r[0] + r[2] / 2, r[1] + r[3] / 2];

// pick chooses item i of n in the drop-down named name.
async function pick(page, name, i, n) {
  const r = (await debug(page)).rects[name];
  await page.mouse.click(...center(r));
  const pop = await until(page, "the " + name + " list", () => goSlidesDebug().rects.popover);
  await page.mouse.click(pop[0] + pop[2] / 2, pop[1] + ((i + 0.5) * pop[3]) / n);
}

// rendered waits for the render a change causes.
async function rendered(page, change) {
  const l0 = (await debug(page)).loads;
  await change();
  await until(page, "the render", (b) => goSlidesDebug().loads > b && !goSlidesDebug().busy, 20000, l0);
  await new Promise((r) => setTimeout(r, 200));
}

// dropFiles drops files made in the page (name → base64 or a size of random
// bytes behind a PNG signature) at the centre of rect.
async function dropFiles(page, rect, files) {
  const [x, y] = center(rect);
  await page.evaluate((files, x, y) => {
    const dt = new DataTransfer();
    for (const [name, v] of Object.entries(files)) {
      let bytes;
      if (typeof v === "number") {
        bytes = new Uint8Array(v);
        for (let i = 0; i < v; i += 4096) bytes[i] = (i * 2654435761) >>> 24;
        bytes.set([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
      } else bytes = Uint8Array.from(atob(v), (c) => c.charCodeAt(0));
      dt.items.add(new File([bytes], name, { type: "image/png" }));
    }
    window.dispatchEvent(new DragEvent("drop", { dataTransfer: dt, clientX: x, clientY: y, bubbles: true, cancelable: true }));
  }, files, x, y);
}

const ITEMS = 12; // the Project menu's items, its title included
const DOT = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==";

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const page = await browser.newPage();
  page.on("pageerror", (e) => console.error("pageerror:", e.message));
  page.on("dialog", (d) => d.accept()); // Replace your deck…? yes
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
  const cdp = await page.target().createCDPSession();
  await cdp.send("Browser.setDownloadBehavior", { behavior: "allow", downloadPath: downloads });
  await page.goto(url, { waitUntil: "load" });
  await until(page, "the editor", () => globalThis.webcanvasReady && globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
  await until(page, "the first preview", () => goSlidesDebug().loads >= 1);
  const out = {};
  let d = await debug(page);
  if (!/^deck\.md\|kinds\/\|media\/\|themes\/$/.test(d.files)) fail("files: " + d.files);

  // A theme from the Project menu, then chosen in the Theme list.
  await pick(page, "project", 2, ITEMS); // New theme (from gaia)
  await until(page, "the new theme in the editor", () => goSlidesDebug().open === "themes/my-gaia.css" && goSlidesDebug().themes.includes("my-gaia"));
  d = await debug(page);
  const themes = d.themes.split(",");
  await rendered(page, () => pick(page, "theme", themes.indexOf("my-gaia"), themes.length));
  d = await debug(page);
  if (!/\ntheme: my-gaia\n/.test(d.source) || d.problems !== "") fail("the deck's theme: " + d.problems + "\n" + d.source.slice(0, 200));
  const frameOf = async (id) => {
    const h = await page.$("#" + id);
    return h.contentFrame();
  };
  let pf = await frameOf("slides-preview");
  out.h1 = await pf.evaluate(() => { const h = document.querySelector("section h1"); return h ? getComputedStyle(h).color : ""; });
  if (out.h1 !== "rgb(79, 70, 229)") fail("the theme's h1 colour is " + out.h1);
  await page.screenshot({ path: path.join(shots, "files-1-theme.png") });

  // A kind from the menu, used by a slide.
  await pick(page, "project", 4, ITEMS);
  await until(page, "the new kind", () => goSlidesDebug().open === "kinds/chapter.yaml");
  await rendered(page, () => page.evaluate(() => goSlidesSetSource("---\nslides: true\ntheme: my-gaia\n---\n\n# Title\n\n---\n\n<!-- _kind: chapter -->\n# Part one\n")));
  pf = await frameOf("slides-preview");
  out.kindClass = await pf.evaluate(() => [...document.querySelectorAll("section")].map((s) => s.className).join("|"));
  if (!/invert/.test(out.kindClass)) fail("the kind did not apply: " + out.kindClass);

  // An image dropped on the editor: into media/, named by the deck, shown
  // as a background through a blob: URL the frame made.
  d = await debug(page);
  await dropFiles(page, d.rects.editor, { "dot.png": DOT });
  await until(page, "the dropped image", () => goSlidesDebug().files.includes("media/dot.png") && goSlidesDebug().source.includes("![dot](media/dot.png)"));
  await rendered(page, () => page.evaluate(() => goSlidesSetSource("---\nslides: true\ntheme: my-gaia\n---\n\n![bg](media/dot.png)\n\n# Over a picture\n")));
  pf = await frameOf("slides-preview");
  // the frame makes its blob: URL once the page posts it the media
  await pf.waitForFunction(() => /blob:null\//.test((document.querySelector("figure") || { getAttribute: () => "" }).getAttribute("style") || ""), { timeout: 10000 })
    .catch(() => fail("the background stayed a path"));
  out.background = await pf.evaluate(async () => {
    const f = document.querySelector("figure");
    const bg = f ? getComputedStyle(f).backgroundImage : "";
    const u = bg.replace(/^url\("?/, "").replace(/"?\)$/, "");
    const ok = await new Promise((res) => { const im = new Image(); im.onload = () => res(im.naturalWidth > 0); im.onerror = () => res(false); im.src = u; });
    return { bg: bg.slice(0, 20), ok };
  });
  if (!out.background.ok || !out.background.bg.startsWith('url("blob:null/')) fail("the background: " + JSON.stringify(out.background));
  await page.screenshot({ path: path.join(shots, "files-2-media.png") });

  // Download the project, and open it again.
  await pick(page, "project", 11, ITEMS);
  const zip = path.join(downloads, "slides-project.zip");
  for (let i = 0; i < 100 && !fs.existsSync(zip); i++) await new Promise((r) => setTimeout(r, 100));
  if (!fs.existsSync(zip)) fail("no " + zip);
  out.zipBytes = fs.statSync(zip).size;
  const before = await debug(page);
  await rendered(page, () => page.evaluate(() => goSlidesSetSource("# replaced")));
  const chooser = page.waitForFileChooser({ timeout: 10000 });
  await pick(page, "project", 9, ITEMS);
  await (await chooser).accept([zip]);
  await until(page, "the project from the .zip", (src) => goSlidesDebug().source === src, 20000, before.source);
  d = await debug(page);
  if (d.files !== before.files) fail("files after the .zip: " + d.files + " (before: " + before.files + ")");
  out.reopened = d.files;

  // 20 MiB of media: an edit's time to the preview, without and with them.
  const edits = async () => {
    const t = [];
    for (let i = 0; i < 7; i++) {
      const s = Date.now();
      await rendered(page, () => page.evaluate((i) => goSlidesSetSource(goSlidesDebug().source + "\n\nedit " + i), i));
      t.push(Date.now() - s - 200);
    }
    t.sort((a, b) => a - b);
    return t[3];
  };
  out.editMsWithout = await edits();
  d = await debug(page);
  const big = {};
  for (let i = 0; i < 4; i++) big["big" + i + ".png"] = 5 << 20;
  const s = Date.now();
  await dropFiles(page, d.rects.files, big);
  await until(page, "20 MiB of media in the frame", () => goSlidesDebug().mediaShown >= 5, 60000);
  out.addMediaMs = Date.now() - s;
  out.usage = (await debug(page)).usage;
  out.editMsWith20MiB = await edits();
  // a new document (a theme change) posts the media again
  const t0 = Date.now();
  await rendered(page, () => page.evaluate(() => goSlidesSetSource(goSlidesDebug().source.replace("theme: my-gaia", "theme: uncover"))));
  await until(page, "the media after a new document", () => goSlidesDebug().mediaShown >= 5);
  out.newDocumentMsWith20MiB = Date.now() - t0;
  await page.screenshot({ path: path.join(shots, "files-3-20mib.png") });

  console.log("FILES_OK " + JSON.stringify(out));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
