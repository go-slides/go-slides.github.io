// The browser tests' one way to answer a file picker.
//
//   const { chooseFile } = require("./chooser.cjs");
//   await chooseFile(page, () => page.mouse.click(x, y), [path]);
//
// Never page.waitForFileChooser: it sends Page.setInterceptFileChooserDialog and
// returns without waiting for the answer, while the click that follows reaches the
// page by another route (Input.dispatchMouseEvent, through the browser's input
// pipeline). On a page's FIRST chooser the click could be handled before the page
// had the interception: Chrome then opens its own chooser, which headless Chrome
// never reports, and nothing arrives for 10 s ("Waiting for FileChooser failed",
// issue 128). Here the interception is ACKNOWLEDGED before trigger() runs, so the
// order is by construction and not by luck.
"use strict";

// chooseFile arms the interception on a session of its own, runs trigger() (the
// click that makes the page ask for a picker), waits for the picker the page
// opened, and sets the files on that very <input type=file> (DOM.setFileInputFiles
// on the event's backend node: the page's own input, which is how its change
// event fires). It throws when no picker opens within timeout ms.
async function chooseFile(page, trigger, files, timeout = 10000) {
  const cdp = await page.createCDPSession();
  try {
    await cdp.send("Page.enable");
    await cdp.send("Page.setInterceptFileChooserDialog", { enabled: true });
    const opened = new Promise((resolve) => cdp.once("Page.fileChooserOpened", resolve));
    await trigger();
    let timer;
    const late = new Promise((_, reject) => { timer = setTimeout(() => reject(new Error("the browser opened no file chooser within " + timeout + " ms")), timeout); });
    const ev = await Promise.race([opened, late]).finally(() => clearTimeout(timer));
    await cdp.send("DOM.setFileInputFiles", { files, backendNodeId: ev.backendNodeId });
  } finally {
    await cdp.detach().catch(() => {});
  }
}

module.exports = { chooseFile };
