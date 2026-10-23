"use strict";
// Behaviour of hero-carousel.html on a page with cards. node hero.cjs URL
const puppeteer = require("puppeteer-core");
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
function fail(m) { console.error("FAIL " + m); process.exit(1); }
(async () => {
  const url = process.argv[2];
  const b = await puppeteer.launch({ executablePath: process.env.CHROME, headless: true, args: ["--no-sandbox"] });
  const out = {};
  {
    const p = await b.newPage();
    await p.setViewport({ width: 1440, height: 900 });
    await p.evaluateOnNewDocument(() => {
      window.__cls = 0;
      new PerformanceObserver((l) => { for (const e of l.getEntries()) if (!e.hadRecentInput) window.__cls += e.value; }).observe({ type: "layout-shift", buffered: true });
    });
    await p.goto(url, { waitUntil: "load" });
    const st = () => p.evaluate(() => ({
      on: [...document.querySelectorAll(".hero-car-card")].findIndex((c) => c.classList.contains("on")),
      // pictures still waiting among the shown cards and the next one
      pending: [...document.querySelectorAll(".hero-car-card")].filter((c, k, all) => k <= all.findIndex((x) => x.classList.contains("on")) + 1 && c.querySelector("img[data-src]")).length,
      later: [...document.querySelectorAll(".hero-car img[data-src]")].length,
      h: Math.round(document.querySelector(".hero-car").getBoundingClientRect().height),
      cls: window.__cls,
      stopped: window.heroCarousel && window.heroCarousel.stopped,
    }));
    const s0 = await st();
    out.start = s0;
    // the second card's picture is fetched up front only as "the next one"
    // (load(1) at start): no card past the next carries a URL yet
    await p.evaluate(() => heroCarousel.show(1));
    await sleep(600);
    const s1 = await st();
    await p.keyboard.press("Tab"); // focus enters: rotation stops
    await p.focus(".hero-car-next");
    await p.keyboard.press("ArrowRight");
    await sleep(500);
    const s2 = await st();
    out.after = { s1, s2 };
    if (s0.on !== 0 || s1.on !== 1 || s2.on !== 2) fail("navigation " + JSON.stringify(out));
    if (s1.h !== s0.h || s2.h !== s0.h) fail("the hero changed height " + JSON.stringify(out));
    if (s2.cls > 0.001) fail("layout shift " + s2.cls);
    if (!s2.stopped) fail("focus did not stop the rotation");
    if (s2.pending !== 0) fail("a picture was never fetched");
    await p.close();
  }
  {
    // reduced motion: nothing turns by itself
    const p = await b.newPage();
    await p.emulateMediaFeatures([{ name: "prefers-reduced-motion", value: "reduce" }]);
    await p.goto(url + "?i=1", { waitUntil: "load" });
    const before = await p.evaluate(() => heroCarousel.index);
    await sleep(Number(process.argv[3] || 6500));
    const after = await p.evaluate(() => ({ i: heroCarousel.index, stopped: heroCarousel.stopped, label: document.querySelector(".hero-car-rot").getAttribute("aria-label") }));
    out.reduced = { before, ...after };
    if (after.i !== before || !after.stopped || after.label !== "Start slide rotation") fail("reduced motion rotated " + JSON.stringify(out.reduced));
    await p.close();
  }
  {
    // no reduced motion: it turns by itself after the interval
    const p = await b.newPage();
    await p.goto(url + "?i=2", { waitUntil: "load" });
    await sleep(Number(process.argv[3] || 6500));
    out.rotated = await p.evaluate(() => heroCarousel.index);
    if (out.rotated !== 1) fail("did not rotate " + out.rotated);
    await p.close();
  }
  console.log("HERO_OK " + JSON.stringify(out));
  await b.close();
})();
