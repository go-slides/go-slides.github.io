// Drives the Theme editor in headless Chrome, with a pointer and a keyboard:
//
//   import a .gsta.xz  →  edit (a colour, a version, a placement)  →  the sheet
//   the preview reads follows  →  export a .gsta.xz  →  `slides theme verify`
//   accepts it  →  a second browser (empty storage) imports the download and
//   shows the same values and compiles the same sheet  →  export CSS  →  undo.
//
// Also: a package the library refuses is shown in Problems with its code and
// changes nothing; a loose .css opens; nothing leaves the page (no request but
// the page's own is made while the editor is used).
//
//   node themeeditor.cjs http://127.0.0.1:PORT/playground/ FIXTURES_DIR DOWNLOADS_DIR
//
// FIXTURES_DIR holds acme.gsta.xz, evil.gsta.xz and loose.css (go test -run
// TestWriteThemeFixturesForBrowser with THEME_FIXTURES_OUT). SLIDES names the
// `slides` binary that verifies the export (the test fails without it unless
// SLIDES_OPTIONAL is set). CHROME names the browser; NODE_PATH finds
// puppeteer-core. No capture is written: a harness that wants one clips the
// pane (never the whole screen) and deletes it after looking.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");

const url = process.argv[2];
const fixtures = process.argv[3];
const downloads = process.argv[4];
fs.mkdirSync(downloads, { recursive: true });

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}
function check(cond, msg) {
  if (!cond) fail(msg);
}

async function until(page, what, fn, ms = 30000, arg) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg);
    if (v) return v;
    if (Date.now() > end) {
      const d = await page.evaluate(() => goSlidesDebug()).catch(() => null);
      if (d) d.source = d.source.split("\n").slice(0, 8).join("\n");
      fail("timed out waiting for " + what + "\n" + JSON.stringify(d && { status: d.status, problems: d.problems, loads: d.loads, fullLoads: d.fullLoads, patches: d.patches, busy: d.busy, themeEditor: d.themeEditor, files: d.files }, null, 1));
    }
    await new Promise((r) => setTimeout(r, 25));
  }
}

const debug = (page) => page.evaluate(() => goSlidesDebug());
const center = (r) => [r[0] + r[2] / 2, r[1] + r[3] / 2];
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// rect is the named control's rectangle, waiting for it to be laid out.
async function rect(page, name) {
  return until(page, "the control " + name, (n) => goSlidesDebug().rects[n], 15000, name);
}

async function click(page, name) {
  await page.mouse.click(...center(await rect(page, name)));
}

// rendered waits for the render a change causes.
async function rendered(page, change) {
  const l0 = (await debug(page)).loads;
  await change();
  await until(page, "the render", (b) => goSlidesDebug().loads > b && !goSlidesDebug().busy, 20000, l0);
}

// chooseFile clicks a button that opens a file picker and answers it.
async function chooseFile(page, button, file) {
  const [chooser] = await Promise.all([page.waitForFileChooser({ timeout: 10000 }), click(page, button)]);
  await chooser.accept([file]);
}

// reveal scrolls the pane's column until the row is inside it.
async function reveal(page, row) {
  for (let i = 0; i < 40; i++) {
    const r = await rect(page, row);
    const sc = await rect(page, "theme-scroll");
    const above = r[1] < sc[1], below = r[1] + r[3] > sc[1] + sc[3];
    if (!above && !below) return r;
    await page.mouse.move(sc[0] + sc[2] / 2, sc[1] + sc[3] / 2);
    await page.mouse.wheel({ deltaY: above ? -120 : 120 });
    await sleep(60);
  }
  fail("the row " + row + " cannot be scrolled into view");
}

// type puts text in the row's field: click it, select all, replace.
async function typeIn(page, row, text) {
  const r = await reveal(page, row);
  await page.mouse.click(r[0] + r[2] - 12, r[1] + r[3] / 2); // the right end of the field: the caret goes to the end of the text
  for (let i = 0; i < 64; i++) await page.keyboard.press("Backspace"); // the field is emptied, key by key
  await page.keyboard.type(text);
}

async function openPage(browser) {
  const ctx = await browser.createBrowserContext(); // empty storage: another person
  const page = await ctx.newPage();
  page.on("pageerror", (e) => fail("pageerror: " + e.message));
  const requests = [];
  page.on("request", (r) => requests.push(r.url()));
  await page.setViewport({ width: 1760, height: 1000, deviceScaleFactor: 1 });
  const cdp = await page.target().createCDPSession();
  // the context is its own browser context: the download behaviour is set for it
  await cdp.send("Browser.setDownloadBehavior", { behavior: "allow", downloadPath: downloads, browserContextId: ctx.id });
  await page.goto(url, { waitUntil: "load" });
  await until(page, "the editor", () => globalThis.webcanvasReady && globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
  await until(page, "the first preview", () => goSlidesDebug().loads >= 1);
  return { page, ctx, requests };
}

// download waits for a file the page saved.
async function download(name, page) {
  const file = path.join(downloads, name);
  const end = Date.now() + 15000;
  while (Date.now() < end) {
    if (fs.existsSync(file) && !fs.existsSync(file + ".crdownload")) {
      await sleep(100);
      return fs.readFileSync(file);
    }
    await sleep(50);
  }
  const d = page ? await debug(page) : {};
  fail("no download " + name + " in " + downloads + ": " + fs.readdirSync(downloads) + "\nstatus: " + d.status + "\nproblems: " + d.problems);
}

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1760,1000"] });
  const out = {};
  const A = await openPage(browser);
  const page = A.page;
  const ownOrigin = new URL(url).origin;

  // the pane opens from the ribbon's Theme editor toggle
  await click(page, "theme-editor");
  const t0 = await until(page, "the Theme pane", () => goSlidesDebug().themeEditor.pane && goSlidesDebug().themeEditor);
  check(t0.name === "", "no theme is being edited yet: " + t0.name);
  check(/No theme|Aucun thème/.test(t0.note), "the pane says there is no theme: " + t0.note);

  // a package the library refuses: shown in Problems, with its code, and nothing changes
  const filesBefore = (await debug(page)).files;
  await chooseFile(page, "theme-import", path.join(fixtures, "evil.gsta.xz"));
  const refused = await until(page, "the refusal", () => /evil\.example/.test(goSlidesDebug().problems) && goSlidesDebug());
  check(/theme\.css|style sheet|CSS/i.test(refused.problems), "Problems names the code's meaning: " + refused.problems);
  check(refused.files === filesBefore, "a refused package changed the project: " + refused.files);
  check(refused.themeEditor.name === "", "a refused package was opened");

  // import the package into the editor
  await chooseFile(page, "theme-import", path.join(fixtures, "acme.gsta.xz"));
  const t1 = await until(page, "the theme opened", () => goSlidesDebug().themeEditor.name === "acme" && goSlidesDebug());
  check(t1.files.includes("theme-src/acme/theme.hcl") && t1.files.includes("themes/acme.css"), "the package's files: " + t1.files);
  check(t1.themeEditor.meta.version === "1.0.0" && t1.themeEditor.meta.extends === "default", "metadata: " + JSON.stringify(t1.themeEditor.meta));
  check(t1.problems === "", "problems after a good import: " + t1.problems);
  out.imported = t1.themeEditor.meta.version;

  // edit: a colour typed in the colour row, the text and background for the contrast badge
  await rendered(page, () => typeIn(page, "theme-row-accent", "#ff0066"));
  await typeIn(page, "theme-row-background", "#ffffff");
  await typeIn(page, "theme-row-text", "#222222");
  const t2 = await until(page, "the values", () => {
    const v = goSlidesDebug().themeEditor.values;
    return v.accent === "#ff0066" && v.background === "#ffffff" && v.text === "#222222" && goSlidesDebug().themeEditor;
  });
  check(t2.sheet.includes("--gsta-accent: #ff0066;"), "the sheet the preview reads follows the control:\n" + t2.sheet);
  check(/^15\.\d\d:1 AAA$/.test(t2.contrast[0]), "the contrast badge of text on background: " + t2.contrast[0]);
  check(/Set both|Réglez/.test(t2.contrast[2]), "an unset pair asks for its colours: " + t2.contrast[2]);
  // a pair that fails says so in words
  await typeIn(page, "theme-row-link", "#aaaaaa");
  const t3 = await until(page, "the failing pair", () => /Fail|Échec/.test(goSlidesDebug().themeEditor.contrast[2]) && goSlidesDebug().themeEditor);
  check(/^2\.\d\d:1 (Fail|Échec)$/.test(t3.contrast[2]), "link #aaa on white: " + t3.contrast[2]);
  out.contrast = t3.contrast[0] + " / " + t3.contrast[2];

  // THEME_SHOT=file writes ONE capture of the Theme pane alone (never the whole
  // screen), for a person to look at; it is theirs to delete, and nothing here
  // keeps it. The default is no capture.
  if (process.env.THEME_SHOT) {
    const r = await debug(page);
    const top = r.rects["theme-pick"], right = r.rects["theme-new"], bot = r.rects["theme-contrast-3"];
    const x = top[0] - 8, y = top[1] - 40;
    await page.screenshot({ path: process.env.THEME_SHOT, clip: { x, y, width: right[0] + right[2] - x + 8, height: bot[1] + bot[3] - y + 8 } });
  }

  // the preview frame shows the edit (the page's own CSS is in the frame's document)
  await until(page, "the preview to carry the colour", () => goSlidesDebug().themeEditor.sheet.includes("#ff0066") && goSlidesDebug().busy === false);

  // metadata: the version
  await typeIn(page, "theme-row-m:version", "1.1.0");
  await until(page, "the version", () => goSlidesDebug().themeEditor.meta.version === "1.1.0");
  // a half-typed version is refused by the library and not stored: the controls keep reading the theme
  await typeIn(page, "theme-row-m:version", "1.1");
  const half = await until(page, "the refusal of a half-typed version", () => /semantic version/.test(goSlidesDebug().problems) && goSlidesDebug());
  check(half.themeEditor.ok && half.themeEditor.meta.version === "1.1.0", "a refused edit is not stored: " + JSON.stringify(half.themeEditor.meta));
  await typeIn(page, "theme-row-m:version", "1.1.0");
  await until(page, "Problems to clear", () => goSlidesDebug().problems === "");

  // undo through the ribbon's Undo: the version step goes back
  const sheetBefore = (await debug(page)).themeEditor.sheet;
  await click(page, "undo");
  await until(page, "undo", () => goSlidesDebug().themeEditor.meta.version === "1.0.0");
  await click(page, "redo");
  await until(page, "redo", () => goSlidesDebug().themeEditor.meta.version === "1.1.0");
  check((await debug(page)).themeEditor.sheet === sheetBefore, "redo restores the sheet");

  // export the package: built in the browser, nothing sent
  const sheetShown = (await debug(page)).themeEditor.sheet;
  await click(page, "theme-package");
  const pkg = await download("acme-1.1.0.gsta.xz", page);
  check(pkg.length > 100 && pkg[0] === 0xfd && pkg.slice(1, 6).toString("latin1") === "7zXZ\0", "the download is an xz stream");
  const exported = path.join(downloads, "acme-1.1.0.gsta.xz");
  const slides = process.env.SLIDES;
  if (slides) {
    const said = execFileSync(slides, ["theme", "verify", exported], { encoding: "utf8" });
    out.verify = said.trim().split("\n")[0];
  } else if (!process.env.SLIDES_OPTIONAL) {
    fail("SLIDES must name the slides binary: the export has to pass `slides theme verify`");
  } else {
    out.verify = "skipped (no SLIDES)";
  }

  // the plain CSS export
  await click(page, "theme-css");
  const css = (await download("acme.css", page)).toString("utf8");
  check(css.includes("@import-theme \"default\"") && css.includes("--gsta-accent: #ff0066;") && css.includes("@theme acme"), "the CSS export:\n" + css.slice(0, 400));

  // a second browser, with empty storage, imports the download: the same theme
  const B = await openPage(browser);
  await click(B.page, "theme-editor");
  await chooseFile(B.page, "theme-import", exported);
  const b1 = await until(B.page, "the theme in the second browser", () => goSlidesDebug().themeEditor.name === "acme" && goSlidesDebug());
  const te = b1.themeEditor;
  check(te.values.accent === "#ff0066" && te.values.background === "#ffffff" && te.values.text === "#222222" && te.values.link === "#aaaaaa", "the values after the round trip: " + JSON.stringify(te.values));
  check(te.meta.version === "1.1.0" && te.meta.extends === "default", "the metadata after the round trip: " + JSON.stringify(te.meta));
  check(te.sheet === sheetShown, "the sheet the preview reads is the same after the round trip");
  check(b1.problems === "", "problems after the round trip: " + b1.problems);
  out.roundTrip = "same values, metadata and sheet";
  // exported again from the second browser: the same bytes
  fs.unlinkSync(exported);
  await click(B.page, "theme-package");
  const again = await download("acme-1.1.0.gsta.xz", B.page);
  check(Buffer.compare(again, pkg) === 0, "a package imported and exported again is not the same bytes");
  out.stableBytes = again.length;

  // a loose sheet opens as a theme, its @import moved into the base
  await chooseFile(B.page, "theme-import", path.join(fixtures, "loose.css"));
  const l1 = await until(B.page, "the loose sheet", () => goSlidesDebug().themeEditor.name === "loose" && goSlidesDebug());
  check(l1.themeEditor.meta.extends === "default" && !l1.themeEditor.sheet.includes('@import "default"') && l1.themeEditor.sheet.includes('@import-theme "default"'), "the loose sheet's base: " + JSON.stringify(l1.themeEditor.meta));

  // nothing left the page: every request was to the page's own origin (or a blob/data URL)
  for (const { requests } of [A, B]) {
    const foreign = requests.filter((u) => !u.startsWith(ownOrigin) && !/^(blob|data|about):/.test(u));
    check(foreign.length === 0, "requests that left the page: " + foreign.join(", "));
  }
  out.foreignRequests = 0;

  await browser.close();
  console.log(JSON.stringify(out));
})().catch((e) => fail(e.stack || String(e)));
