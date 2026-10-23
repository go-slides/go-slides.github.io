// split.cjs <url> <shots-dir>: the panes' dividers in headless Chrome. At 1920x1080
// and 1280x800, light and dark: a divider is dragged with the mouse (the pointer
// crossing the preview's iframe, which must not take the drag), the pane gets the
// width, the frames follow the canvas, a double-click pulls a pane shut and
// another brings it back, the keyboard (F6, arrows, Enter) works, and what was
// dragged is still there after a reload. Screenshots after the drags go to
// <shots-dir>/split-<size>-<scheme>.png.
const path = require("node:path");
const puppeteer = require("puppeteer-core");

const [url, shots] = process.argv.slice(2);
const chrome = process.env.CHROME || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const fail = (m) => { console.log("FAIL " + m); process.exit(1); };

(async () => {
  const b = await puppeteer.launch({ executablePath: chrome, headless: "new", args: ["--no-sandbox"] });
  const log = [];
  for (const [w, h] of [[1920, 1080], [1280, 800]]) {
    for (const scheme of ["light", "dark"]) {
      const tag = `${w}x${h}-${scheme}`;
      const ctx = await b.createBrowserContext(); // each page starts with nothing stored
      const page = await ctx.newPage();
      page.on("pageerror", (e) => fail(tag + " pageerror: " + e.message));
      await page.setViewport({ width: w, height: h, deviceScaleFactor: 1 });
      await page.emulateMediaFeatures([{ name: "prefers-color-scheme", value: scheme }]);
      const open = async () => {
        await page.goto(url, { waitUntil: "networkidle0" });
        await page.waitForFunction("window.goSlidesReady===true", { timeout: 60000 });
      };
      const dbg = () => page.evaluate(() => goSlidesDebug());
      const rects = async () => (await dbg()).rects;
      await open();
      // 1280 starts with the file panel and the help off; switch the file panel
      // on from the View tab so that every size has the same dividers to try
      if (w < 1440) {
        const r = await rects();
        await page.mouse.click(r["tab-View"][0] + r["tab-View"][2] / 2, r["tab-View"][1] + r["tab-View"][3] / 2);
        await page.waitForFunction(() => goSlidesDebug().rects.filesToggle, { timeout: 10000 });
        const t = (await rects()).filesToggle;
        await page.mouse.click(t[0] + t[2] / 2, t[1] + t[3] / 2);
        await sleep(300);
      }
      const frames = async (label) => {
        const got = await page.evaluate(() => {
          const d = goSlidesDebug();
          const box = (s) => { const r = document.querySelector(s).getBoundingClientRect(); return [r.left, r.top, r.width, r.height]; };
          return { rects: d.rects, preview: box("#slides-preview"), rail: box("#slides-rail"), railShown: getComputedStyle(document.querySelector("#slides-rail")).display };
        });
        const near = (a, c) => a.every((v, i) => Math.abs(v - c[i]) <= 1.01);
        if (!near(got.preview, got.rects.preview)) fail(`${tag} ${label}: preview frame ${got.preview} off the canvas's ${got.rects.preview}`);
        if (got.railShown !== "none" && !near(got.rail, got.rects.thumbs)) fail(`${tag} ${label}: thumbnail frame ${got.rail} off the strip ${got.rects.thumbs}`);
      };
      const drag = async (name, dx, label) => {
        const d = (await rects())[name];
        if (!d) fail(`${tag}: no ${name}`);
        const cx = d[0] + d[2] / 2, cy = d[1] + 120;
        await page.mouse.move(cx, cy);
        await page.mouse.down();
        await page.mouse.move(cx + dx / 2, cy, { steps: 6 });
        await frames(label + " mid-drag");
        await page.mouse.move(cx + dx, cy, { steps: 6 });
        await page.mouse.up();
        await frames(label);
      };
      const width = async (name) => (await rects())[name][2];

      const f0 = await width("files");
      await drag("split-0", 70, "files +70");
      const f1 = await width("files");
      if (Math.abs(f1 - f0 - 70) > 1.5) fail(`${tag}: files ${f0} -> ${f1}, want +70`);
      const cursor = await page.evaluate(() => document.querySelector("canvas").style.cursor);
      void cursor;

      // the preview-side divider crossing the preview iframe: the thumbnails grow
      const t0 = await width("thumbs");
      await drag("split-2", 40, "thumbs +40");
      const t1 = await width("thumbs");
      if (Math.abs(t1 - t0 - 40) > 1.5) fail(`${tag}: thumbs ${t0} -> ${t1}, want +40`);

      // the divider after the preview: the help is wider when dragged left
      let helpMoved = false;
      const rs = await rects();
      if (rs["split-3"] && rs.help) {
        const hw0 = rs.help[2];
        await drag("split-3", -60, "help +60");
        const hw1 = await width("help");
        if (Math.abs(hw1 - hw0 - 60) > 1.5) fail(`${tag}: help ${hw0} -> ${hw1}, want +60`);
        helpMoved = true;
      }

      // a double-click pulls the file panel shut, another brings it back
      const dbl = async (name) => {
        const d = (await rects())[name];
        await page.mouse.click(d[0] + d[2] / 2, d[1] + 120, { count: 2, delay: 40 });
        await sleep(150);
      };
      await dbl("split-0");
      if ((await rects()).files[2] > 1) fail(`${tag}: the file panel is still ${(await rects()).files[2]} wide after a double-click`);
      await frames("files shut");
      await dbl("split-0");
      const f2 = await width("files");
      if (Math.abs(f2 - f1) > 1.5) fail(`${tag}: the file panel came back at ${f2}, was ${f1}`);

      // the keyboard: F6 to the first divider, ArrowRight, Enter shuts, Enter opens
      await page.keyboard.press("Escape"); // the press on the divider focused it: back to the editor
      await page.keyboard.press("F6");
      await page.keyboard.press("ArrowRight");
      await sleep(100);
      const f3 = await width("files");
      if (Math.abs(f3 - f2 - 16) > 1.5) fail(`${tag}: ArrowRight moved the panel by ${f3 - f2}, want 16`);
      await page.keyboard.press("Enter");
      await sleep(100);
      if ((await rects()).files[2] > 1) fail(`${tag}: Enter did not shut the panel`);
      await page.keyboard.press("Enter");
      await sleep(100);
      const f4 = await width("files");
      if (Math.abs(f4 - f3) > 1.5) fail(`${tag}: Enter did not bring the panel back (${f4}, was ${f3})`);
      await page.keyboard.press("Escape");

      await sleep(300);
      await page.screenshot({ path: path.join(shots, `split-${tag}.png`) });

      // reload: the dragged widths are the page's now
      const before = await rects();
      await open();
      const after = await rects();
      if (Math.abs(after.files[2] - before.files[2]) > 1.5 || Math.abs(after.thumbs[2] - before.thumbs[2]) > 1.5) {
        fail(`${tag}: after a reload files ${after.files[2]} / thumbs ${after.thumbs[2]}, were ${before.files[2]} / ${before.thumbs[2]}`);
      }
      await frames("reloaded");
      // the editor keeps its room: about 60 columns at 7.8 px each at 1920 wide
      if (w >= 1900 && after.editor[2] < 60 * 7.8) fail(`${tag}: the editor is ${after.editor[2]} px wide`);
      log.push(`${tag}: files ${f0}->${f1}, thumbs ${t0}->${t1}${helpMoved ? ", help dragged" : ""}, editor ${after.editor[2]}, kept over a reload`);
      await ctx.close();
    }
  }
  await b.close();
  console.log("SPLIT_OK\n" + log.join("\n"));
})();
