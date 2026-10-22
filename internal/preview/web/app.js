const assetVersion = new URL(import.meta.url).searchParams.get("v") || "1";
const pdfjs = await import(`/vendor/pdf.mjs?v=${encodeURIComponent(assetVersion)}`);

pdfjs.GlobalWorkerOptions.workerSrc = `/vendor/pdf.worker.mjs?v=${encodeURIComponent(assetVersion)}`;

const $ = (selector) => document.querySelector(selector);
const viewport = $("#viewport");
const pages = $("#pages");
const statusButton = $("#build-status");
const statusText = $("#status-text");
const logPanel = $("#log-panel");
const logText = $("#log-text");
const exportButton = $("#export");
const dialog = $("#export-dialog");
const nameInput = $("#export-name");

let snapshot = null;
let pdfDocument = null;
let loadingTask = null;
let requestedRevision = -1;
let loadToken = 0;
let renderToken = 0;
let renderTask = null;
let customScale = null;
let displayedScale = 1;
let renderFailure = "";
let resizeTimer = null;

function setStatus(status, message) {
  statusButton.className = `status status--${status}`;
  statusText.textContent = message;
}

function updateSnapshot(next) {
  snapshot = next;
  const ready = next.status === "ready";
  $("#filename").textContent = next.filename;
  document.title = `${next.filename} · Lampo`;
  exportButton.disabled = !next.hasPdf;

  if (next.status === "error") {
    setStatus("error", "Build failed");
  } else if (next.status === "building") {
    setStatus("building", next.hasPdf ? "Updating" : "Compiling");
  } else if (ready) {
    setStatus("ready", "Up to date");
  }

  const error = next.error || renderFailure;
  $("#error-banner").hidden = !error;
  $("#error-message").textContent = error;
  logText.textContent = next.log || "No compiler output yet.";
  $("#compile-time").textContent = next.durationMs
    ? `${(next.durationMs / 1000).toFixed(2)}s build`
    : "Local preview";

  if (next.hasPdf && next.revision !== requestedRevision) {
    requestedRevision = next.revision;
    void loadPDF(next.revision);
  } else if (!next.hasPdf && next.status === "error") {
    showErrorEmptyState();
  }
}

function showErrorEmptyState() {
  const message = $("#empty-state");
  if (!message) return;
  message.querySelector("h1").textContent = "Waiting for a clean build";
  message.querySelector("p").textContent = "Fix the LaTeX error and save your file. Lampo will automatically try again.";
  message.querySelector(".loading-line").hidden = true;
}

async function loadPDF(revision) {
  const token = ++loadToken;
  try {
    const response = await fetch(`/preview.pdf?revision=${revision}`, { cache: "no-store" });
    if (!response.ok) throw new Error(`PDF request failed (${response.status})`);
    const bytes = new Uint8Array(await response.arrayBuffer());
    if (token !== loadToken) return;
    const task = pdfjs.getDocument({ data: bytes, useSystemFonts: true });
    const nextDocument = await task.promise;
    if (token !== loadToken) {
      await task.destroy();
      return;
    }
    const previousTask = loadingTask;
    pdfDocument = nextDocument;
    loadingTask = task;
    renderFailure = "";
    await renderPages(nextDocument);
    if (previousTask && previousTask !== task) await previousTask.destroy();
  } catch (error) {
    if (token !== loadToken) return;
    console.error("Lampo PDF render error:", error);
    renderFailure = `Could not render the PDF: ${error.message}`;
    $("#error-banner").hidden = false;
    $("#error-message").textContent = renderFailure;
    setStatus("error", "Preview error");
  }
}

function fitScale(pageWidth) {
  const margin = viewport.clientWidth <= 700 ? 24 : 64;
  return Math.max(0.15, Math.min(1.55, (viewport.clientWidth - margin) / pageWidth));
}

async function renderPages(pdfDoc = pdfDocument) {
  if (!pdfDoc) return;
  const token = ++renderToken;
  if (renderTask) renderTask.cancel();
  const ratio = viewport.scrollTop / Math.max(1, viewport.scrollHeight - viewport.clientHeight);
  const staging = document.createDocumentFragment();

  try {
    const first = await pdfDoc.getPage(1);
    if (token !== renderToken) return;
    const base = first.getViewport({ scale: 1 });
    const scale = customScale ?? fitScale(base.width);
    displayedScale = scale;
    $("#zoom-value").textContent = customScale == null ? "Fit width" : `${Math.round(scale * 100)}%`;
    const pixelRatio = Math.min(window.devicePixelRatio || 1, 2);

    for (let pageNumber = 1; pageNumber <= pdfDoc.numPages; pageNumber++) {
      if (token !== renderToken) return;
      const page = pageNumber === 1 ? first : await pdfDoc.getPage(pageNumber);
      const view = page.getViewport({ scale });
      const section = document.createElement("section");
      section.className = "pdf-page";
      section.setAttribute("aria-label", `Page ${pageNumber}`);
      section.style.width = `${view.width}px`;
      section.style.height = `${view.height}px`;

      const canvas = document.createElement("canvas");
      canvas.width = Math.round(view.width * pixelRatio);
      canvas.height = Math.round(view.height * pixelRatio);
      canvas.style.width = `${view.width}px`;
      canvas.style.height = `${view.height}px`;
      section.append(canvas);
      const context = canvas.getContext("2d", { alpha: false });
      if (!context) throw new Error("Canvas 2D rendering is unavailable");
      renderTask = page.render({
        canvasContext: context,
        viewport: view,
        transform: pixelRatio === 1 ? undefined : [pixelRatio, 0, 0, pixelRatio, 0, 0],
        background: "#ffffff",
      });
      await renderTask.promise;
      renderTask = null;
      if (token !== renderToken) return;

      // Text remains selectable, while annotations preserve clickable links.
      const textLayer = document.createElement("div");
      textLayer.className = "textLayer";
      textLayer.style.setProperty("--scale-factor", String(scale));
      section.append(textLayer);
      try {
        const layer = new pdfjs.TextLayer({
          textContentSource: page.streamTextContent(), container: textLayer, viewport: view,
        });
        await layer.render();
      } catch (error) {
        console.warn("Lampo: text selection is unavailable for this page", error);
        textLayer.remove();
      }
      if (token !== renderToken) return;

      try {
        const annotations = await page.getAnnotations({ intent: "display" });
        const linkLayer = document.createElement("div");
        linkLayer.className = "linkLayer";
        for (const annotation of annotations) {
          if (annotation.subtype !== "Link" || !annotation.url || !annotation.rect) continue;
          let url;
          try { url = new URL(annotation.url); } catch { continue; }
          if (!["https:", "http:", "mailto:"].includes(url.protocol)) continue;
          const [x1, y1] = view.convertToViewportPoint(annotation.rect[0], annotation.rect[1]);
          const [x2, y2] = view.convertToViewportPoint(annotation.rect[2], annotation.rect[3]);
          const anchor = document.createElement("a");
          anchor.href = url.toString();
          anchor.target = "_blank";
          anchor.rel = "noopener noreferrer";
          anchor.setAttribute("aria-label", `Open ${url.hostname || url.protocol}`);
          anchor.style.left = `${Math.min(x1, x2)}px`;
          anchor.style.top = `${Math.min(y1, y2)}px`;
          anchor.style.width = `${Math.abs(x2 - x1)}px`;
          anchor.style.height = `${Math.abs(y2 - y1)}px`;
          linkLayer.append(anchor);
        }
        section.append(linkLayer);
      } catch (error) {
        console.warn("Lampo: some links cannot be previewed", error);
      }
      staging.append(section);
    }

    if (token !== renderToken) return;
    pages.replaceChildren(staging);
    $("#page-count").textContent = `${pdfDoc.numPages} ${pdfDoc.numPages === 1 ? "page" : "pages"}`;
    viewport.scrollTop = ratio * Math.max(0, viewport.scrollHeight - viewport.clientHeight);
  } catch (error) {
    if (token !== renderToken || error?.name === "RenderingCancelledException") return;
    console.error("Lampo page render error:", error);
    renderFailure = `Preview rendering failed: ${error.message}`;
    $("#error-banner").hidden = false;
    $("#error-message").textContent = renderFailure;
    setStatus("error", "Preview error");
  }
}

function changeZoom(factor) {
  customScale = Math.max(0.25, Math.min(3, displayedScale * factor));
  void renderPages();
}

$("#zoom-out").addEventListener("click", () => changeZoom(1 / 1.2));
$("#zoom-in").addEventListener("click", () => changeZoom(1.2));
$("#zoom-value").addEventListener("click", () => {
  customScale = null;
  void renderPages();
});

new ResizeObserver(() => {
  if (customScale !== null || !pdfDocument) return;
  clearTimeout(resizeTimer);
  resizeTimer = setTimeout(() => void renderPages(), 100);
}).observe(viewport);

function toggleLog(open) { logPanel.hidden = !open; }
statusButton.addEventListener("click", () => toggleLog(logPanel.hidden));
$("#show-error-log").addEventListener("click", () => toggleLog(true));
$("#close-log").addEventListener("click", () => toggleLog(false));
$("#copy-log").addEventListener("click", async () => {
  try {
    await navigator.clipboard.writeText(logText.textContent);
    $("#copy-log").textContent = "Copied!";
    setTimeout(() => { $("#copy-log").textContent = "Copy log"; }, 1500);
  } catch (error) { console.warn("Clipboard unavailable:", error); }
});

function closeDialog() { dialog.hidden = true; }
exportButton.addEventListener("click", () => {
  if (!snapshot?.hasPdf) return;
  nameInput.value = snapshot.filename.replace(/\.tex$/i, "") + ".pdf";
  dialog.hidden = false;
  nameInput.focus();
  nameInput.select();
});
$("#export-cancel").addEventListener("click", closeDialog);
dialog.addEventListener("click", (event) => { if (event.target === dialog) closeDialog(); });
$("#export-form").addEventListener("submit", (event) => {
  event.preventDefault();
  let filename = nameInput.value.trim();
  if (!filename) return;
  if (!/\.pdf$/i.test(filename)) filename += ".pdf";
  const link = document.createElement("a");
  link.href = `/api/export?filename=${encodeURIComponent(filename)}`;
  link.download = filename;
  document.body.append(link);
  link.click();
  link.remove();
  closeDialog();
});

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") { closeDialog(); toggleLog(false); }
  if (event.target instanceof HTMLInputElement || event.metaKey || event.ctrlKey || event.altKey) return;
  if (event.key === "+" || event.key === "=") changeZoom(1.2);
  if (event.key === "-") changeZoom(1 / 1.2);
  if (event.key === "0") { customScale = null; void renderPages(); }
});

const events = new EventSource("/events");
events.addEventListener("state", (event) => {
  try { updateSnapshot(JSON.parse(event.data)); }
  catch (error) { console.error("Invalid Lampo server event:", error); }
});
events.addEventListener("error", () => {
  if (events.readyState !== EventSource.OPEN) setStatus("error", "Reconnecting");
});
