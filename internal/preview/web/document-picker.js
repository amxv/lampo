// Compact document picker for split-screen previews. Browser-native controls
// keep the UI small, accessible and independent of any framework.
export function createDocumentPicker(onSelect) {
  const host = document.querySelector("#document-picker");
  const trigger = document.querySelector("#document-trigger");
  const menu = document.querySelector("#document-menu");
  const search = document.querySelector("#document-search");
  const list = document.querySelector("#document-list");
  const footer = document.querySelector("#document-menu-footer");
  const label = document.querySelector("#filename");
  let documents = [];
  let activeId = "";
  let lastItemsKey = "";

  function positionMenu() {
    if (window.innerWidth > 700) {
      menu.style.top = "";
      menu.style.left = "";
      return;
    }
    const bounds = trigger.getBoundingClientRect();
    const width = Math.min(326, window.innerWidth - 24);
    menu.style.left = `${Math.max(12, Math.min(bounds.left, window.innerWidth - width - 12))}px`;
    menu.style.top = `${bounds.bottom + 8}px`;
  }

  function close(restoreFocus = false) {
    if (menu.hidden) return;
    menu.hidden = true;
    trigger.setAttribute("aria-expanded", "false");
    if (restoreFocus) trigger.focus();
  }

  function open() {
    positionMenu();
    menu.hidden = false;
    trigger.setAttribute("aria-expanded", "true");
    search.value = "";
    render();
    search.focus();
  }

  function render() {
    const query = search.value.trim().toLocaleLowerCase();
    const matches = documents
      .filter((doc) => doc.id.toLocaleLowerCase().includes(query))
      .sort((a, b) => a.id === activeId ? -1 : b.id === activeId ? 1 : a.id.localeCompare(b.id));
    const fragment = document.createDocumentFragment();
    for (const doc of matches) {
      const selected = doc.id === activeId;
      const button = document.createElement("button");
      button.type = "button";
      button.className = `document-option${selected ? " is-active" : ""}`;
      button.dataset.documentId = doc.id;
      button.setAttribute("aria-current", selected ? "true" : "false");
      button.title = doc.id;
      // Only static SVG markup is interpolated. Document names and paths
      // are always assigned via textContent, never injected as HTML.
      button.innerHTML = `<svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Z"/><path d="M14 2v6h6M8 13h8M8 17h6"/></svg>`;
      const copy = document.createElement("span");
      copy.className = "document-option-copy";
      const name = document.createElement("span");
      name.className = "document-option-name";
      name.textContent = doc.name;
      copy.append(name);
      if (doc.id !== doc.name) {
        const path = document.createElement("span");
        path.className = "document-option-path";
        path.textContent = doc.id.slice(0, -(doc.name.length + 1));
        copy.append(path);
      }
      button.append(copy);
      if (selected) {
        const tick = document.createElement("span");
        tick.className = "document-option-check";
        tick.textContent = "✓";
        tick.setAttribute("aria-label", "Current document");
        button.append(tick);
      }
      button.addEventListener("click", () => {
        close(true);
        if (doc.id !== activeId) void onSelect(doc.id);
      });
      fragment.append(button);
    }
    if (!matches.length) {
      const empty = document.createElement("div");
      empty.className = "document-empty";
      empty.textContent = query ? "No matching documents" : "No standalone LaTeX files yet";
      fragment.append(empty);
    }
    list.replaceChildren(fragment);
    const count = documents.length;
    footer.textContent = `${count} ${count === 1 ? "document" : "documents"} · Auto-discovered from your project`;
  }

  function update(nextDocuments, nextActiveId) {
    documents = Array.isArray(nextDocuments) ? nextDocuments : [];
    activeId = nextActiveId || "";
    const active = documents.find((doc) => doc.id === activeId);
    label.textContent = active?.name || "No document";
    trigger.title = active ? `${active.id} · Switch document` : "Choose a document";
    // Don't discard keyboard focus or search input on every build update.
    const key = JSON.stringify([documents, activeId]);
    if (key !== lastItemsKey) {
      lastItemsKey = key;
      render();
    }
  }

  trigger.addEventListener("click", () => menu.hidden ? open() : close());
  search.addEventListener("input", render);
  search.addEventListener("keydown", (event) => {
    if (event.key !== "ArrowDown" && event.key !== "Enter") return;
    const first = list.querySelector(".document-option");
    if (first) {
      event.preventDefault();
      if (event.key === "Enter") first.click();
      else first.focus();
    }
  });
  list.addEventListener("keydown", (event) => {
    if (!["ArrowDown", "ArrowUp"].includes(event.key)) return;
    const items = [...list.querySelectorAll(".document-option")];
    const index = items.indexOf(document.activeElement);
    const next = items[index + (event.key === "ArrowDown" ? 1 : -1)];
    if (next) {
      event.preventDefault();
      next.focus();
    }
  });
  document.addEventListener("pointerdown", (event) => {
    if (!menu.hidden && !host.contains(event.target)) close();
  });
  window.addEventListener("resize", () => { if (!menu.hidden) positionMenu(); });
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && !menu.hidden) {
      event.preventDefault();
      close(true);
    }
    if (event.key === "/" && menu.hidden && !event.metaKey && !event.ctrlKey && !event.altKey &&
        !(event.target instanceof HTMLInputElement) && !(event.target instanceof HTMLTextAreaElement)) {
      event.preventDefault();
      open();
    }
  });

  return { update, close };
}
