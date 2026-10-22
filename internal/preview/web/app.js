const assetVersion = new URL(import.meta.url).searchParams.get("v") || "1";
const pdfjs = await import(`/vendor/pdf.mjs?v=${encodeURIComponent(assetVersion)}`);
const { createDocumentPicker } = await import(`/document-picker.js?v=${encodeURIComponent(assetVersion)}`);

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
const emptyPreview = $("#empty-state").cloneNode(true);

let snapshot = null;
let pdfDocument = null;
let loadingTask = null;
let requestedRevision = -1;
let loadToken = 0;
let renderToken = 0;
let renderTask = null;
let customScale = null;
let displayedScale = 1;
let renderedPages = [];
let renderedDocument = null;
let renderFailure = "";
let sharpRenderTimer = null;
let renderMode = null;

const picker = createDocumentPicker(async (id) => {
  try {
    const response = await fetch("/api/select", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id }),
    });
    if (!response.ok) throw new Error(`Could not select document (${response.status})`);
    updateSnapshot(await response.json());
  } catch (error) {
    console.error("Lampo document switch error:", error);
    setStatus("error", "Switch failed");
    $("#error-banner").hidden = false;
    $("#error-message").textContent = error.message;
  }
});

// Resizing uses a cheap CSS transform immediately, then rasterizes each page
// at the *actual* viewport scale and device pixel ratio once resizing stops.
const maxFitScale = 1.55;
const sharpenDelay = 160;
const maxCanvasPixels = 16_000_000;

function setStatus(status, message) {
  statusButton.className = `status status--${status}`;
  statusText.textContent = message;
}

function updateSnapshot(next) {
	if (snapshot && next.sequence < snapshot.sequence) return;
	const switched = snapshot && next.activeId !== snapshot.activeId;
	if (switched) resetDocumentPreview(next.status);
  snapshot = next;
  const ready = next.status === "ready";
  picker.update(next.documents, next.activeId);
  document.title = `${next.filename || "LaTeX preview"} · Lampo`;
  exportButton.disabled = !next.hasPdf;

  if (next.status === "error") {
    setStatus("error", "Build failed");
  } else if (next.status === "building") {
    setStatus("building", next.hasPdf ? "Updating" : "Compiling");
  } else if (ready) {
    setStatus("ready", "Up to date");
  } else if (next.status === "empty") {
    setStatus("error", "No documents");
  }

  const error = next.status === "empty" ? "" : (next.error || renderFailure);
  $("#error-banner").hidden = !error;
  $("#error-message").textContent = error;
  logText.textContent = next.log || "No compiler output yet.";
  $("#compile-time").textContent = next.durationMs
    ? `${(next.durationMs / 1000).toFixed(2)}s build`
    : "Local preview";

  if (next.hasPdf && next.revision !== requestedRevision) {
    requestedRevision = next.revision;
    void loadPDF(next.revision);
  } else if (!next.hasPdf && (next.status === "error" || next.status === "empty")) {
    showErrorEmptyState();
  }
}

function resetDocumentPreview(status) {
  // An in-flight PDF from the previous document must never appear under the
  // newly selected filename. Keep cached bytes server-side until requested.
  ++loadToken;
  ++renderToken;
  renderTask?.cancel();
  clearTimeout(sharpRenderTimer);
  renderMode = null;
  pdfDocument = null;
  renderedDocument = null;
  renderedPages = [];
  displayedScale = 1;
  requestedRevision = -1;
  const placeholder = emptyPreview.cloneNode(true);
  if (status === "empty") {
    placeholder.querySelector("h1").textContent = "No documents found";
    placeholder.querySelector("p").textContent = "Create a standalone .tex file in this project to preview it here.";
    placeholder.querySelector(".loading-line").hidden = true;
  } else {
    placeholder.querySelector("h1").textContent = "Opening document";
    placeholder.querySelector("p").textContent = "Your selected PDF will appear here shortly.";
  }
  pages.replaceChildren(placeholder);
  $("#page-count").textContent = "—";
  $("#zoom-value").textContent = "Fit width";
  if (loadingTask) {
    const previous = loadingTask;
    loadingTask = null;
    void previous.destroy().catch(() => {});
  }
}

function showErrorEmptyState() {
  const message = $("#empty-state");
  if (!message) return;
  if (snapshot?.status === "empty") {
    message.querySelector("h1").textContent = "No documents found";
    message.querySelector("p").textContent = "Create a standalone .tex file here and Lampo will discover it automatically.";
  } else {
    message.querySelector("h1").textContent = "Waiting for a clean build";
    message.querySelector("p").textContent = "Fix the LaTeX error and save your file. Lampo will automatically try again.";
  }
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
  return Math.max(0.15, Math.min(maxFitScale, (viewport.clientWidth - margin) / pageWidth));
}

function outputPixelRatio(width, height) {
  // Render at the display's native Retina/HiDPI resolution where possible,
  // with a bounded bitmap for very large paper sizes or extreme zoom levels.
  return Math.min(
    Math.max(1, window.devicePixelRatio || 1),
    3,
    Math.sqrt(maxCanvasPixels / (width * height)),
    8192 / Math.max(width, height),
  );
}

// Keep the expensive canvas/text/annotation layers at their original render
// resolution. The outer page takes up the *scaled* space in the document flow,
// while one transform scales all inner layers together. This is cheap enough
// to run on every browser resize frame, with no debounced PDF re-render.
function layoutPages() {
  if (!renderedPages.length) return;
  const scale = customScale ?? fitScale(renderedPages[0].width);
  const label = customScale == null ? "Fit width" : `${Math.round(scale * 100)}%`;
  if ($("#zoom-value").textContent !== label) $("#zoom-value").textContent = label;
  if (Math.abs(scale - displayedScale) < 0.0001) return;
  displayedScale = scale;
  for (const { frame, content, width, height, renderScale } of renderedPages) {
    frame.style.width = `${width * scale}px`;
    frame.style.height = `${height * scale}px`;
    // Disable CSS compositing once the raster matches the displayed size.
    // Keeping scale(1) creates a layer that can soften fine PDF lettering.
    const ratio = scale / renderScale;
    content.style.transform = Math.abs(ratio - 1) < 0.0001 ? "none" : `scale(${ratio})`;
  }
}

function needsSharpRender() {
  if (renderedDocument !== pdfDocument || !renderedPages.length) return false;
  return renderedPages.some(({ width, height, renderScale, pixelRatio: renderedRatio }) =>
    Math.abs(renderScale - displayedScale) > Math.max(0.002, displayedScale * 0.005) ||
    Math.abs(outputPixelRatio(width * displayedScale, height * displayedScale) - renderedRatio) > 0.005
  );
}

function cancelSharpening() {
  if (renderMode !== "sharpen") return;
  ++renderToken;
  renderTask?.cancel();
  renderMode = null;
}

function scheduleSharpRender() {
  clearTimeout(sharpRenderTimer);
  if (!needsSharpRender() || renderMode !== null) return;
  sharpRenderTimer = setTimeout(() => {
    if (needsSharpRender()) void renderPages(pdfDocument, true);
  }, sharpenDelay);
}

async function renderPages(pdfDoc = pdfDocument, sharpen = false) {
  if (!pdfDoc) return;
  clearTimeout(sharpRenderTimer);
  const token = ++renderToken;
  if (renderTask) renderTask.cancel();
  renderMode = sharpen ? "sharpen" : "document";
  const ratio = viewport.scrollTop / Math.max(1, viewport.scrollHeight - viewport.clientHeight);
  const staging = document.createDocumentFragment();
  const newPages = [];

  try {
    const first = await pdfDoc.getPage(1);
    if (token !== renderToken) return;
    if (sharpen && (renderedDocument !== pdfDoc || renderedPages.length !== pdfDoc.numPages)) return;
    const renderScale = sharpen
      ? displayedScale
      : customScale ?? fitScale(first.getViewport({ scale: 1 }).width);
    for (let pageNumber = 1; pageNumber <= pdfDoc.numPages; pageNumber++) {
      if (token !== renderToken) return;
      const page = pageNumber === 1 ? first : await pdfDoc.getPage(pageNumber);
      if (token !== renderToken) return;
      const view = page.getViewport({ scale: renderScale });
      const pixelRatio = outputPixelRatio(view.width, view.height);
      const section = sharpen ? renderedPages[pageNumber - 1].frame : document.createElement("section");
      if (!sharpen) {
        section.className = "pdf-page";
        section.setAttribute("aria-label", `Page ${pageNumber}`);
      }
      const content = document.createElement("div");
      content.className = "pdf-page-content";
      content.style.width = `${view.width}px`;
      content.style.height = `${view.height}px`;
      if (!sharpen) section.append(content);
      const pageRecord = {
        frame: section,
        content,
        width: view.width / renderScale,
        height: view.height / renderScale,
        renderScale,
        pixelRatio,
      };

      const canvas = document.createElement("canvas");
      canvas.width = Math.round(view.width * pixelRatio);
      canvas.height = Math.round(view.height * pixelRatio);
      canvas.style.width = `${view.width}px`;
      canvas.style.height = `${view.height}px`;
      content.append(canvas);
      const context = canvas.getContext("2d", { alpha: false });
      if (!context) throw new Error("Canvas 2D rendering is unavailable");
      const task = page.render({
        canvasContext: context,
        viewport: view,
        transform: pixelRatio === 1 ? undefined : [pixelRatio, 0, 0, pixelRatio, 0, 0],
        background: "#ffffff",
      });
      renderTask = task;
      try {
        await task.promise;
      } finally {
        if (renderTask === task) renderTask = null;
      }
      if (token !== renderToken) return;

      // Text remains selectable, while annotations preserve clickable links.
      const textLayer = document.createElement("div");
      textLayer.className = "textLayer";
      textLayer.style.setProperty("--scale-factor", String(renderScale));
      content.append(textLayer);
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
        content.append(linkLayer);
      } catch (error) {
        console.warn("Lampo: some links cannot be previewed", error);
      }

      if (token !== renderToken) return;
      if (sharpen) {
        // Commit one completed page at a time. The old canvas stays visible
        // during the whole render, and frame dimensions never change.
        section.replaceChildren(content);
        renderedPages[pageNumber - 1] = pageRecord;
        const scaleRatio = displayedScale / renderScale;
        content.style.transform = Math.abs(scaleRatio - 1) < 0.0001 ? "none" : `scale(${scaleRatio})`;
      } else {
        newPages.push(pageRecord);
        staging.append(section);
      }
    }

    if (token !== renderToken) return;
    if (!sharpen) {
      pages.replaceChildren(staging);
      renderedPages = newPages;
      renderedDocument = pdfDoc;
      // Force layout for the new PDF even when its scale matches the last one.
      displayedScale = 0;
      layoutPages();
      $("#page-count").textContent = `${pdfDoc.numPages} ${pdfDoc.numPages === 1 ? "page" : "pages"}`;
      viewport.scrollTop = ratio * Math.max(0, viewport.scrollHeight - viewport.clientHeight);
    }
  } catch (error) {
    if (token !== renderToken || error?.name === "RenderingCancelledException") return;
    console.error("Lampo page render error:", error);
    if (sharpen) {
      console.warn("Lampo: keeping the previous PDF rendering after a sharpening failure.");
    } else {
      renderFailure = `Preview rendering failed: ${error.message}`;
      $("#error-banner").hidden = false;
      $("#error-message").textContent = renderFailure;
      setStatus("error", "Preview error");
    }
  } finally {
    if (token === renderToken) {
      renderMode = null;
      // If the window moved while a new PDF was loading, resharpen it now.
      if (needsSharpRender() && !sharpen) scheduleSharpRender();
    }
  }
}

function changeZoom(factor) {
  customScale = Math.max(0.25, Math.min(3, displayedScale * factor));
  cancelSharpening();
  layoutPages();
  scheduleSharpRender();
}

$("#zoom-out").addEventListener("click", () => changeZoom(1 / 1.2));
$("#zoom-in").addEventListener("click", () => changeZoom(1.2));
$("#zoom-value").addEventListener("click", () => {
  customScale = null;
  cancelSharpening();
  layoutPages();
  scheduleSharpRender();
});

new ResizeObserver(() => {
  if (customScale !== null) return;
  const previousScale = displayedScale;
  layoutPages();
  if (Math.abs(displayedScale - previousScale) > 0.0001) cancelSharpening();
  scheduleSharpRender();
}).observe(viewport);

// A tab moved between monitors can change pixel density without changing its
// CSS width. In that case refresh the backing canvas at the new device scale.
function watchDevicePixelRatio() {
  const query = matchMedia(`(resolution: ${window.devicePixelRatio}dppx)`);
  query.addEventListener("change", () => {
    cancelSharpening();
    scheduleSharpRender();
    watchDevicePixelRatio();
  }, { once: true });
}
watchDevicePixelRatio();

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
  if (event.key === "0") { customScale = null; cancelSharpening(); layoutPages(); scheduleSharpRender(); }
});

const events = new EventSource("/events");
events.addEventListener("state", (event) => {
  try { updateSnapshot(JSON.parse(event.data)); }
  catch (error) { console.error("Invalid Lampo server event:", error); }
});
events.addEventListener("error", () => {
  if (events.readyState !== EventSource.OPEN) setStatus("error", "Reconnecting");
});
