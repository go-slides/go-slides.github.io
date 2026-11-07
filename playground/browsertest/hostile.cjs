// Shared links that carry a hostile deck, opened in headless Chrome: a deck that
// includes itself (go-slides/slides#67 overflowed the stack and killed the render
// worker), one that asks for a billion list items, one that writes without end,
// one that nests a hundred thousand quotes. Each must show its error in the
// problems and leave the page alive: no crash, no hang, the editor still takes
// keystrokes and the render worker still answers; where the hostile part is one
// short line, deleting it brings the preview back. Prints HOSTILE_OK {...}.
//
//   node hostile.cjs http://127.0.0.1:PORT/playground/
//
// CHROME names the browser; NODE_PATH finds puppeteer-core.
"use strict";
const puppeteer = require("puppeteer-core");
const zlib = require("zlib");

const url = process.argv[2];

function fail(msg) {
  console.error("FAIL " + msg);
  process.exit(1);
}

// a shared link's fragment: raw deflate, base64url (playground/share.go)
const link = (src) => url + "#deck=" + zlib.deflateRawSync(Buffer.from(src), { level: 9 }).toString("base64url");

const front = "---\ntheme: default\n---\n\n# Fine so far\n\n";
// tail is the hostile line, the last of the deck: deleting it must bring the
// preview back (none for a deck whose hostile part is the whole body)
const decks = {
  "a deck that includes itself": { tail: '{{ include "deck.md" }}', problem: /include cycle/i },
  "a billion list items": { tail: "{{ range seq 1000000000 }}- x{{ end }}", problem: /seq 1000000000/i },
  "a template that writes without end": { tail: '{{ range seq 1000000 }}{{ repeat 1000 "x" }}{{ end }}', problem: /wrote more than/i },
  "a nested range over seq 2000 x seq 2000": { tail: "{{ range seq 2000 }}{{ range seq 2000 }}{{ end }}{{ end }}", problem: /ran more than \d+ steps|ran for more than|rendering took more than 5 s/i }, // in the browser the page's own 5 s watchdog may stop it first (the worker is slower than native)
  "a hundred thousand nested quotes": { body: "> ".repeat(100000) + "x", problem: /nested block quotes/i },
};

async function until(page, what, fn, arg, ms = 30000) {
  const end = Date.now() + ms;
  for (;;) {
    const v = await page.evaluate(fn, arg).catch(() => null);
    if (v) return v;
    if (Date.now() > end) {
      const d = await page.evaluate(() => { const g = goSlidesDebug(); return { status: g.status, problems: g.problems.slice(0, 200), source: g.source.slice(-80) }; }).catch(() => "page unreachable");
      fail("timed out waiting for " + what + ": " + JSON.stringify(d));
    }
    await new Promise((r) => setTimeout(r, 25));
  }
}

(async () => {
  const browser = await puppeteer.launch({
    executablePath: process.env.CHROME,
    headless: true,
    args: ["--no-sandbox", "--window-size=1440,900"],
  });
  const took = {};
  for (const [name, d] of Object.entries(decks)) {
    if (process.env.HOSTILE_ONLY && !name.includes(process.env.HOSTILE_ONLY)) continue; // one deck, to look at it
    const src = d.tail ? front + d.tail : front + d.body;
    // a context of its own: no saved draft from the last deck decides what loads
    const ctx = await browser.createBrowserContext();
    const page = await ctx.newPage();
    const crashes = [];
    page.on("error", (e) => crashes.push("page crashed: " + e.message));
    page.on("pageerror", (e) => crashes.push("pageerror: " + e.message));
    await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 1 });
    const t0 = Date.now();
    await page.goto(link(src), { waitUntil: "load" });
    await until(page, name + ": the editor", () => globalThis.webcanvasReady && globalThis.goSlidesReady && document.querySelector(".stage").classList.contains("ready"));
    // the deck from the link is what the editor holds
    await until(page, name + ": the shared deck in the editor", (n) => goSlidesDebug().source.length >= n, src.length);
    // refused: the problem is shown and the render is over
    await until(page, name + ": the problem", (re) => new RegExp(re.source, re.flags).test(goSlidesDebug().problems + "\n" + goSlidesDebug().status) && !goSlidesDebug().busy, { source: d.problem.source, flags: d.problem.flags });
    took[name] = Date.now() - t0;
    if (crashes.length) fail(name + ": " + crashes.join("; "));

    // alive: a keystroke at the start of the deck is taken, the worker answers
    // (a new round trip), and the problem is still the same refusal
    const [x, y] = await page.evaluate(() => goSlidesCaretPixel(0, 0));
    await page.mouse.click(x + 2, y + 6);
    const trips = (await page.evaluate(() => goSlidesDebug())).lastRoundTripMs;
    await page.keyboard.type(" ");
    await until(page, name + ": the keystroke in the source", (n) => goSlidesDebug().source.length === n + 1, src.length);
    await until(page, name + ": the worker to answer again", (t) => !goSlidesDebug().busy && goSlidesDebug().lastRoundTripMs !== t, trips).catch(() => null);
    await until(page, name + ": the problem again", (re) => new RegExp(re.source, re.flags).test(goSlidesDebug().problems + "\n" + goSlidesDebug().status) && !goSlidesDebug().busy, { source: d.problem.source, flags: d.problem.flags });
    await page.keyboard.press("Backspace"); // the character behind the caret: the space typed
    await until(page, name + ": the source back", (n) => goSlidesDebug().source.length === n, src.length);
    await until(page, name + ": the render of it to end", () => !goSlidesDebug().busy);

    if (d.tail) {
      // delete the hostile line: click at its end, then a Backspace a character
      const last = src.split("\n").length - 1;
      // click at the line's start (its end can lie past the editor's text area) and
      // press End
      const [ex, ey] = await page.evaluate((l) => goSlidesCaretPixel(l, 0), last);
      await page.mouse.click(ex + 2, ey + 6);
      await page.keyboard.press("End");
      if (process.env.HOSTILE_DEBUG) console.error("clicked at", ex, ey, "line", last, "col", d.tail.length);
      // HOSTILE_SHOTS names a directory outside every repository for a picture
      if (process.env.HOSTILE_SHOTS) await page.screenshot({ path: require("path").join(process.env.HOSTILE_SHOTS, "hostile-" + Object.keys(decks).indexOf(name) + ".png") });
      for (let i = 0; i < d.tail.length; i++) await page.keyboard.press("Backspace");
      if (process.env.HOSTILE_DEBUG) console.error("after the Backspaces:", JSON.stringify(await page.evaluate(() => { const g = goSlidesDebug(); return { n: g.source.length, tail: g.source.slice(-70), busy: g.busy, status: g.status }; })));
      await until(page, name + ": the problems to clear", () => goSlidesDebug().problems === "" && !goSlidesDebug().busy);
      const preview = await until(page, name + ": the preview", () => true).then(() => page.frames().find((f) => f !== page.mainFrame()));
      const text = preview ? await preview.evaluate(() => document.body.innerText).catch(() => "") : "";
      if (!text.includes("Fine so far")) fail(name + ": the preview shows " + JSON.stringify(text.slice(0, 80)));
    }
    if (crashes.length) fail(name + ": " + crashes.join("; "));
    await ctx.close();
  }
  await browser.close();
  console.log("HOSTILE_OK " + JSON.stringify(took));
})().catch((e) => fail(e.stack || String(e)));
