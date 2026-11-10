// What an export says about the animated images of the deck, in headless
// Chrome: the project holds two animated SVG images (a rotation the export
// carries natively, a path that morphs, which it carries as an animated GIF);
// Save as ▸ PowerPoint saves the deck, and afterwards the status line points
// to Problems, whose lines say, in the language of the page, how each image
// was carried. Run in English and in French (the page's own language, not the
// export engine's English lines); a line left in the export's English would fail
// the French run. Prints EXPORTNOTES_OK.
//
//   node exportnotes.cjs http://127.0.0.1:PORT/playground/ DOWNLOAD_DIR
//
// CHROME names the browser; NODE_PATH finds puppeteer-core.
"use strict";
const puppeteer = require("puppeteer-core");
const fs = require("fs");
const path = require("path");

const url = process.argv[2];
const dir = path.resolve(process.argv[3] || ".");
fs.mkdirSync(dir, { recursive: true });

const fail = (msg) => { console.error("FAIL " + msg); process.exit(1); };

const spin = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect x="45" y="5" width="10" height="45" fill="#c00"><animateTransform attributeName="transform" type="rotate" from="0 50 50" to="360 50 50" dur="2s" repeatCount="indefinite"/></rect></svg>';
const morph = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><path d="M10 10L90 10L90 90Z" fill="#06c"><animate attributeName="d" values="M10 10L90 10L90 90Z;M10 10L90 50L50 90Z;M10 10L90 10L90 90Z" dur="2s" repeatCount="indefinite"/></path></svg>';
const data = (s) => "data:image/svg+xml;base64," + Buffer.from(s).toString("base64");
const deck = "# Animated\n\n![w:300](" + data(spin) + ")\n\n---\n\n![w:300](" + data(morph) + ")\n";

const langs = [
  { name: "en", exportTab: "Export", base: url, sentence: "See Problems", native: "animated natively in PPTX", gif: "animated GIF" },
  { name: "fr", exportTab: "Exporter", base: url.replace("/playground/", "/fr/playground/"), sentence: "Voir Problèmes", native: "animé nativement dans PPTX", gif: "GIF animé" },
];

async function until(page, what, fn, arg, ms = 180000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg).catch(() => null);
    if (v) return v;
    if (Date.now() > end) fail("timed out waiting for " + what + ": " + (await page.evaluate(() => goSlidesDebug().status).catch(() => "?")));
    await new Promise((r) => setTimeout(r, 50));
  }
}

(async () => {
  const browser = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox", "--window-size=1440,900"] });
  const out = {};
  for (const l of langs) {
    const page = await browser.newPage();
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
    page.on("pageerror", (e) => console.error("pageerror:", e.message));
    const cdp = await page.createCDPSession();
    await cdp.send("Browser.setDownloadBehavior", { behavior: "allow", downloadPath: dir });
    await page.goto(l.base, { waitUntil: "load" });
    await until(page, "the editor", () => globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
    await page.evaluate((t) => goSlidesSetSource(t), deck);
    await until(page, "the deck", () => /2/.test(goSlidesDebug().status) && !goSlidesDebug().busy);
    const tab = await until(page, "the Export tab", (k) => goSlidesDebug().rects[k], "tab-" + l.exportTab);
    await page.mouse.click(tab[0] + tab[2] / 2, tab[1] + tab[3] / 2);
    const r = await until(page, "the PowerPoint button", () => goSlidesDebug().rects["save-pptx"]);
    await page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2);
    const st = await until(page, "the notes after the export", (s) => { const d = goSlidesDebug(); return (d.status.includes(s) || d.status.startsWith("⚠")) && d; }, l.sentence);
    if (st.status.startsWith("⚠")) fail(l.name + ": " + st.status);
    const problems = st.problems;
    if (!problems.includes(l.native)) fail(l.name + ": no line saying " + JSON.stringify(l.native) + " in:\n" + problems);
    if (!problems.includes(l.gif)) fail(l.name + ": no line saying " + JSON.stringify(l.gif) + " in:\n" + problems);
    if (/svg-animation:|animated natively|an animated GIF of/.test(problems) && l.name !== "en") fail(l.name + ": a line of the export in English:\n" + problems);
    out[l.name] = { status: st.status, problems: problems.split("\n") };
    await page.close();
  }
  console.log("EXPORTNOTES_OK " + JSON.stringify(out));
  await browser.close();
})().catch((e) => fail(e.stack || String(e)));
