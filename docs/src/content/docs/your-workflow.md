---
title: Your live preview
description: Make Lampo feel right at home next to your editor or chat.
order: 2
category: Guides
summary: Live updates, sharp zoom, easy PDF export, and a preview that stays out of your way.
---

Your browser is your preview. Your editor stays your editor. Lampo simply keeps the two in sync.

## Work side by side

Open Lampo in one browser window, then place it beside your editor or an AI coding assistant. Make either window as narrow or wide as you like: the PDF fits the available space automatically.

You can resize freely. Lampo adjusts the page immediately and sharpens the text as soon as you finish resizing, including on high-density displays.

## Make a change, see the result

Saving a LaTeX file triggers a new preview. This includes chapters, figures, and other files your document uses.

Look for the small status indicator at the top:

| Status | What it means |
| --- | --- |
| **Up to date** | You're seeing the latest successfully compiled PDF. |
| **Updating** | A saved change is being compiled. Keep working. |
| **Build failed** | Something in the LaTeX needs fixing. The last working PDF stays visible. |

Click the indicator to see the compiler output if you need to troubleshoot a build.

## Zoom and navigate

Use the **−** and **+** buttons in the toolbar. Click the zoom value to return to **Fit width**. Text remains selectable, and links in the PDF are clickable.

These shortcuts work when you aren't typing in an input:

| Key | Action |
| --- | --- |
| `+` / `-` | Zoom in / zoom out |
| `0` | Fit the page to the window width |
| `/` | Open document search |
| `Esc` | Close an open dialog |

## Export whenever you're ready

Click **Export PDF**, type a name, and download the current PDF. You can export as often as you want while the preview keeps running.

## End your session

Go back to the terminal where you started Lampo and press **Ctrl+C**. The preview server stops and its temporary build files are cleared. Your `.tex` project is unchanged.

**Next:** [Switch between multiple documents](/docs/multiple-documents).
