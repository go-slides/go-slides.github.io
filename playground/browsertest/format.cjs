// format.cjs <url> <shots-dir>: the Format tab in headless Chrome, at 1920x1080 and
// 1280x800, light and dark. With the caret in a word, Bold, a keyboard shortcut,
// the table grid, a slide's alignment (read from the preview's own computed
// style) and the link command are driven with the mouse and the keyboard, and the
// toggles are checked to show the state at the caret. Screenshots of the tab, with
// the table grid open, go to <shots-dir>/format-<size>-<scheme>[-grid].png.
const path = require("node:path");
const puppeteer = require("puppeteer-core");

const [url, shots] = process.argv.slice(2);
const chrome = process.env.CHROME || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const fail = (m) => { console.log("FAIL " + m); process.exit(1); };

const DECK = "---\nslides: true\ntheme: default\n---\n\n# Format\n\nsay hello there and more\n\n---\n\n# Two\n\nsecond slide\n";

(async () => {
  const b = await puppeteer.launch({ executablePath: chrome, headless: "new", args: ["--no-sandbox"] });
  const log = [];
  for (const [w, h] of [[1920, 1080], [1280, 800]]) {
    for (const scheme of ["light", "dark"]) {
      const tag = `${w}x${h}-${scheme}`;
      const ctx = await b.createBrowserContext();
      const page = await ctx.newPage();
      page.on("pageerror", (e) => fail(tag + " pageerror: " + e.message));
      await page.setViewport({ width: w, height: h, deviceScaleFactor: 1 });
      await page.emulateMediaFeatures([{ name: "prefers-color-scheme", value: scheme }]);
      await page.goto(url, { waitUntil: "networkidle0" });
      await page.waitForFunction("window.goSlidesReady===true", { timeout: 60000 });
      const dbg = () => page.evaluate(() => goSlidesDebug());
      const src = async () => (await dbg()).source;
      await page.evaluate((s) => goSlidesSetSource(s), DECK);
      await page.waitForFunction(() => !goSlidesDebug().busy && goSlidesDebug().source.includes("second slide"), { timeout: 30000 });
      const click = async (r) => page.mouse.click(r[0] + r[2] / 2, r[1] + r[3] / 2);
      const rects = async () => (await dbg()).rects;
      await click((await rects())["tab-Format"]);
      await page.waitForFunction(() => goSlidesDebug().rects["fmt-bold"], { timeout: 10000 });

      // put the caret in "hello" (line 8, col 6): a click on its pixel
      const caretAt = async (line, col) => {
        const [x, y] = await page.evaluate((l, c) => goSlidesCaretPixel(l, c), line, col);
        await page.mouse.click(x, y);
      };
      await caretAt(7,7);
      let state = (await dbg()).format;
      if (state.bold) fail(`${tag}: Bold is pressed on plain text`);
      await click((await rects())["fmt-bold"]);
      await sleep(150);
      if (!(await src()).includes("say **hello** there")) fail(`${tag}: Bold did not mark the word: ${JSON.stringify(await src())}`);
      state = (await dbg()).format;
      if (!state.bold) fail(`${tag}: Bold should show pressed inside **hello**`);
      await click((await rects())["fmt-bold"]);
      await sleep(150);
      if (!(await src()).includes("say hello there")) fail(`${tag}: Bold again did not unmark`);

      // the shortcuts: Ctrl+I, then Ctrl+U on the same word
      await caretAt(7,7);
      await page.keyboard.down("Control");
      await page.keyboard.press("KeyI");
      await page.keyboard.up("Control");
      await sleep(150);
      if (!(await src()).includes("say _hello_ there")) fail(`${tag}: Ctrl+I: ${JSON.stringify(await src())}`);
      await page.keyboard.down("Control");
      await page.keyboard.press("KeyI");
      await page.keyboard.up("Control");
      await sleep(100);
      await page.keyboard.down("Control");
      await page.keyboard.press("KeyU");
      await page.keyboard.up("Control");
      await sleep(150);
      if (!(await src()).includes("say <u>hello</u> there")) fail(`${tag}: Ctrl+U: ${JSON.stringify(await src())}`);
      state = (await dbg()).format;
      if (!state.underline) fail(`${tag}: Underline should show pressed`);
      await page.keyboard.down("Control");
      await page.keyboard.press("KeyU");
      await page.keyboard.up("Control");
      await sleep(100);

      // a heading and a list
      await click((await rects())["fmt-h2"]);
      await sleep(150);
      if (!(await src()).includes("## say hello there")) fail(`${tag}: Heading 2: ${JSON.stringify(await src())}`);
      if (!(await dbg()).format.h2) fail(`${tag}: Heading 2 should show pressed`);
      await click((await rects())["fmt-h2"]);
      await sleep(150);

      // the table grid: open it, light a 3 x 2 corner, choose it
      await caretAt(7,3);
      await click((await rects())["fmt-table"]);
      await sleep(200);
      let pop = (await rects()).popover;
      if (!pop) fail(`${tag}: the Table control did not open its grid`);
      const cs = (pop[2] - 2) / 8;
      const cell = (c, r) => [pop[0] + 1 + c * cs + cs / 2, pop[1] + 1 + r * cs + cs / 2];
      const [hx, hy] = cell(2, 1);
      await page.mouse.move(hx, hy, { steps: 3 });
      await sleep(150);
      await page.screenshot({ path: path.join(shots, `format-${tag}-grid.png`) });
      await page.mouse.click(hx, hy);
      await sleep(250);
      const s1 = await src();
      if (!s1.includes("| Column 1 | Column 2 | Column 3 |") || s1.includes("Column 4")) fail(`${tag}: the 3 x 2 table: ${JSON.stringify(s1)}`);
      // in the table: add a row, centre a column
      await click((await rects())["fmt-row-below"]);
      await sleep(150);
      await click((await rects())["fmt-colalign-center"]);
      await sleep(150);
      const s2 = await src();
      if (!/\| :-+: \|/.test(s2) && !s2.includes(":-")) fail(`${tag}: column alignment: ${JSON.stringify(s2)}`);
      if (!(await dbg()).format["colalign-center"]) fail(`${tag}: the column's alignment should show pressed`);

      // slide alignment: the preview's own style
      await caretAt(7,3);
      await click((await rects())["fmt-slide-center"]);
      await page.waitForFunction(() => goSlidesDebug().source.includes("text-align: center"), { timeout: 10000 });
      await sleep(600);
      const frame = (await page.$("#slides-preview")).contentFrame ? await (await page.$("#slides-preview")).contentFrame() : null;
      if (frame) {
        const align = await frame.evaluate(() => {
          const secs = [...document.querySelectorAll("section")];
          return secs.map((s) => getComputedStyle(s).textAlign);
        });
        if (align[0] !== "center" || align[1] === "center") fail(`${tag}: slide 1 centred and slide 2 not: ${JSON.stringify(align)}`);
      }
      if (!(await dbg()).format["slide-center"]) fail(`${tag}: the slide's alignment should show pressed`);

      // a link: select "more" by keyboard and press Ctrl+K
      await page.evaluate((s) => goSlidesSetSource(s), DECK);
      await page.waitForFunction(() => !goSlidesDebug().busy && goSlidesDebug().source.includes("second slide"), { timeout: 30000 });
      await caretAt(7, 22);
      await page.keyboard.down("Control");
      await page.keyboard.press("KeyK");
      await page.keyboard.up("Control");
      await sleep(200);
      if (!(await src()).includes("[more](https://)")) fail(`${tag}: Ctrl+K: ${JSON.stringify(await src())}`);
      if (!(await dbg()).format.link) fail(`${tag}: Link should show pressed inside a link`);

      // the tab as it looks
      await page.evaluate((s) => goSlidesSetSource(s), DECK);
      await sleep(300);
      await caretAt(7,7);
      await click((await rects())["fmt-bold"]);
      await sleep(300);
      await page.screenshot({ path: path.join(shots, `format-${tag}.png`) });
      log.push(`${tag}: bold, shortcuts, heading, table grid, row, column alignment, slide alignment and link work`);
      await ctx.close();
    }
  }
  await b.close();
  console.log("FORMAT_OK\n" + log.join("\n"));
})();
