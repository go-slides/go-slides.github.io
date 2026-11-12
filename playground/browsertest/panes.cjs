// panes.cjs <url> <shots-dir>: the panes' close buttons in headless Chrome, at
// 1920x1080, light and dark. Every pane (the file tree, the editor, the thumbnails,
// the preview, the help, Versions, GitHub and Theme) is closed from the button in its
// header: the others fill the room and none shrinks; the View tab's toggle opens it
// again at the width it had. What is open is kept over a reload. The last of the
// editor and the preview refuses to close, from its button and from the View tab.
// The keyboard: F6, Shift+Tab reaches a close button,
// Enter closes the pane and gives the keyboard back to the editor.
// One element capture per scheme (the canvas, never the screen) goes to
// <shots-dir>/panes-<scheme>.png.
const fs = require("node:fs");
const path = require("node:path");
const puppeteer = require("puppeteer-core");

const [url, shots] = process.argv.slice(2);
const chrome = process.env.CHROME || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const fail = (m) => { console.log("FAIL " + m); process.exit(1); };

// the panes, in the splitter's order; toggle is the View tab control that shows one
const PANES = [
  { id: "files", toggle: "filesToggle" },
  { id: "editor", toggle: "editorToggle" },
  { id: "thumbs", toggle: "thumbsToggle" },
  { id: "preview", toggle: "previewToggle" },
  { id: "help", toggle: "helpToggle" },
  { id: "versions", toggle: "versions" },
  { id: "git", toggle: "git" },
  { id: "theme", toggle: "theme-editor" },
];

(async () => {
  fs.mkdirSync(shots, { recursive: true });
  const b = await puppeteer.launch({ executablePath: chrome, headless: "new", args: ["--no-sandbox"] });
  const log = [];
  for (const scheme of ["light", "dark"]) {
    const tag = `1920x1080-${scheme}`;
    const ctx = await b.createBrowserContext(); // each page starts with nothing stored
    const page = await ctx.newPage();
    page.on("pageerror", (e) => fail(tag + " pageerror: " + e.message));
    await page.setViewport({ width: 1920, height: 1080, deviceScaleFactor: 1 });
    await page.emulateMediaFeatures([{ name: "prefers-color-scheme", value: scheme }]);
    const open = async () => {
      await page.goto(url, { waitUntil: "networkidle0" });
      await page.waitForFunction("window.goSlidesReady===true", { timeout: 60000 });
    };
    const rects = async () => (await page.evaluate(() => goSlidesDebug())).rects;
    const center = (r) => [r[0] + r[2] / 2, r[1] + r[3] / 2];
    const until = async (what, fn, ms = 10000) => {
      const end = Date.now() + ms;
      for (;;) {
        const v = await fn(await rects());
        if (v) return v;
        if (Date.now() > end) fail(`${tag}: timed out waiting for ${what}`);
        await sleep(30);
      }
    };
    // the View tab, then a control of it
    const viewTab = async () => {
      const r = await rects();
      await page.mouse.click(...center(r["tab-View"]));
      await until("the View tab", (rs) => rs.filesToggle);
    };
    const toggle = async (name) => { await viewTab(); await page.mouse.click(...center((await rects())[name])); };
    // widths of the panes that are open
    const widths = (rs) => Object.fromEntries(PANES.filter((p) => rs["pane-" + p.id]).map((p) => [p.id, rs["pane-" + p.id][2]]));
    const edges = (rs) => {
      const open = PANES.filter((p) => rs["pane-" + p.id]).map((p) => rs["pane-" + p.id]);
      return [Math.min(...open.map((r) => r[0])), Math.max(...open.map((r) => r[0] + r[2]))];
    };
    // the frames follow the canvas: the preview's iframe over the preview pane
    const previewFrame = () => page.evaluate(() => { const r = document.querySelector("#slides-preview").getBoundingClientRect(); return [r.left, r.top, r.width, r.height]; });

    await open();
    // every pane on: the three task panes from the View tab
    for (const id of ["versions", "git", "theme"]) {
      await toggle(PANES.find((p) => p.id === id).toggle);
      await until(`pane ${id}`, (rs) => rs["pane-" + id]);
    }
    const all = await rects();
    for (const p of PANES) if (!all["pane-" + p.id] || !all["close-" + p.id]) fail(`${tag}: pane ${p.id} has no header or close button`);

    // each pane closes from its button; the others fill and none shrinks; reopened
    // from the View tab it is as wide as it was
    for (const p of PANES) {
      const before = await rects();
      const w0 = widths(before), e0 = edges(before);
      await page.mouse.click(...center(before["close-" + p.id]));
      const after = await until(`${p.id} closed`, (rs) => !rs["pane-" + p.id] && rs);
      const w1 = widths(after), e1 = edges(after);
      if (Math.abs(e0[0] - e1[0]) > 1.5 || Math.abs(e0[1] - e1[1]) > 1.5) fail(`${tag}: closing ${p.id} left a gap: edges ${e0} -> ${e1}`);
      let grew = false;
      for (const q of Object.keys(w1)) {
        if (w1[q] < w0[q] - 1.5) fail(`${tag}: closing ${p.id} shrank ${q}: ${w0[q]} -> ${w1[q]}`);
        grew = grew || w1[q] > w0[q] + 1.5;
      }
      if (!grew) fail(`${tag}: closing ${p.id} freed room nobody took`);
      if (p.id === "preview") {
        const f = await previewFrame();
        if (f[2] > 1) fail(`${tag}: the preview is closed and its frame is still ${f[2]} wide`);
      }
      if (p.id === "editor") await sleep(50);
      await toggle(p.toggle);
      const back = await until(`${p.id} reopened`, (rs) => rs["pane-" + p.id] && rs);
      const w2 = widths(back);
      for (const q of Object.keys(w0)) {
        if (Math.abs(w2[q] - w0[q]) > 1.5) fail(`${tag}: reopening ${p.id}: ${q} is ${w2[q]} wide, was ${w0[q]}`);
      }
      if (p.id === "preview") {
        const f = await previewFrame(), r = back.preview;
        if (Math.abs(f[0] - r[0]) > 1.01 || Math.abs(f[2] - r[2]) > 1.01) fail(`${tag}: the preview frame ${f} is off its pane ${r}`);
      }
    }

    // the last of the editor and the preview refuses: from its button and the View tab
    let rs = await rects();
    await page.mouse.click(...center(rs["close-editor"]));
    await until("editor closed", (r) => !r["pane-editor"]);
    const only = (await rects())["pane-preview"];
    if (!only) fail(`${tag}: the preview vanished with the editor`);
    await page.mouse.click(...center((await rects())["close-preview"]));
    await sleep(150);
    if (!(await rects())["pane-preview"]) fail(`${tag}: the last pane closed from its button`);
    await toggle("previewToggle");
    await sleep(150);
    if (!(await rects())["pane-preview"]) fail(`${tag}: the last pane closed from the View tab`);
    // its tooltip says why: hover the disabled button (one capture of the canvas, then deleted by the caller)
    const cb = (await rects())["close-preview"];
    await page.mouse.move(...center(cb));
    await sleep(900);
    const canvas = await page.$("canvas");
    const shot = path.join(shots, `panes-${scheme}.png`);
    await canvas.screenshot({ path: shot });
    await page.mouse.move(5, 500);
    // the editor comes back from the View tab
    await toggle("editorToggle");
    await until("editor back", (r) => r["pane-editor"]);

    // the keyboard: F6 to the first divider, Shift+Tab to the file panel's close
    // button, Enter closes it and the keyboard goes back to the editor
    await page.keyboard.press("F6");
    await sleep(80);
    await page.keyboard.down("Shift");
    await page.keyboard.press("Tab");
    await page.keyboard.up("Shift");
    await sleep(80);
    await page.keyboard.press("Enter");
    await until("files closed from the keyboard", (r) => !r["pane-files"]);
    await toggle("filesToggle");
    await until("files back", (r) => r["pane-files"]);

    // what is open is kept over a reload: the panes the person closed stay closed
    // (Versions, GitHub and Theme are opened for a task: they start closed on every page)
    for (const id of ["versions", "git", "theme", "thumbs", "files", "help", "preview"]) {
      await page.mouse.click(...center((await rects())["close-" + id]));
      await until(`${id} closed`, (x) => !x["pane-" + id]);
    }
    const kept = await rects();
    await sleep(400); // the stored state is written
    await open();
    const reloaded = await rects();
    for (const id of ["preview", "thumbs", "files", "help"]) {
      if (reloaded["pane-" + id]) fail(`${tag}: pane ${id} was closed and is open after a reload`);
    }
    if (!reloaded["pane-editor"]) fail(`${tag}: the editor is gone after a reload`);
    if (Math.abs(reloaded["pane-editor"][2] - kept["pane-editor"][2]) > 1.5) {
      fail(`${tag}: the editor is ${reloaded["pane-editor"][2]} wide after a reload, was ${kept["pane-editor"][2]}`);
    }
    for (const p of PANES.filter((q) => ["files", "thumbs", "preview", "help"].includes(q.id))) {
      await toggle(p.toggle);
      await until(`${p.id} reopened after the reload`, (r) => r["pane-" + p.id]);
    }
    log.push(`${tag}: eight panes closed and reopened, the last one refused, kept over a reload`);
    await ctx.close();
  }
  await b.close();
  console.log("PANES_OK\n" + log.join("\n"));
})();
