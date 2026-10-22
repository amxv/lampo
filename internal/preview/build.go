package preview

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/amxv/lampo/internal/latex"
)

type cachedDocument struct {
	pdf       []byte
	result    latex.Result
	compiled  bool
	dirty     bool
	updatedAt time.Time
}

type completedBuild struct {
	id     string
	result latex.Result
}

// buildLoop owns all mutable project state. Switching, rediscovery and builds
// are serialized in this loop, so a canceled compile can never publish a PDF
// belonging to the previously selected document. Every document has its own
// latexmk output directory and last-good PDF cache.
func (s *server) buildLoop(ctx context.Context, project latex.Project, outputDir, engine string, changes <-chan struct{}) {
	documents := project.Documents
	active := project.InitialID
	cache := make(map[string]*cachedDocument)
	for _, document := range documents {
		cache[document.ID] = &cachedDocument{dirty: true}
	}

	completed := make(chan completedBuild, 1)
	var runningID string
	var cancel context.CancelFunc
	var running bool
	var debounce <-chan time.Time
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()

	// The view always reflects the selected document, even when the watcher
	// invalidates a cache or a build temporarily fails.
	show := func(newPDF bool) {
		next := Snapshot{ActiveID: active, Documents: documents}
		if active == "" {
			next.Status = "empty"
			next.Error = "No standalone LaTeX documents found. Add a .tex file to continue."
			s.publish(next, nil)
			return
		}
		entry := cache[active]
		if entry == nil {
			return
		}
		next.Log = entry.result.Log
		next.DurationMs = entry.result.Duration.Milliseconds()
		next.UpdatedAt = entry.updatedAt
		switch {
		case entry.dirty || (running && runningID == active):
			next.Status = "building"
		case !entry.compiled:
			next.Status = "building"
		case entry.result.Error != "":
			next.Status = "error"
			next.Error = entry.result.Error
		default:
			next.Status = "ready"
		}
		var pdf []byte
		if newPDF {
			pdf = entry.pdf
		}
		s.publish(next, pdf)
	}

	start := func() {
		if running || debounce != nil || active == "" {
			return
		}
		entry := cache[active]
		if entry == nil || !entry.dirty {
			return
		}
		entry.dirty = false
		id := active
		digest := sha256.Sum256([]byte(id))
		buildDir := filepath.Join(outputDir, hex.EncodeToString(digest[:8]))
		if err := os.MkdirAll(buildDir, 0700); err != nil {
			entry.compiled = true
			entry.result = latex.Result{Error: fmt.Sprintf("create build directory: %v", err)}
			show(false)
			return
		}
		builder := latex.Builder{
			Source:    filepath.Join(project.Root, filepath.FromSlash(id)),
			OutputDir: buildDir,
			Engine:    engine,
		}
		buildCtx, stop := context.WithCancel(ctx)
		cancel = stop
		runningID = id
		running = true
		show(false)
		go func() {
			result := builder.Build(buildCtx)
			select {
			case completed <- completedBuild{id: id, result: result}:
			case <-ctx.Done():
			}
		}()
	}

	show(false)
	start()
	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}
			return
		case request := <-s.selectDocument:
			valid := false
			for _, document := range documents {
				if request.id == document.ID {
					valid = true
					break
				}
			}
			if !valid {
				request.reply <- fmt.Errorf("unknown document %q", request.id)
				continue
			}
			if active != request.id {
				active = request.id
				if running && runningID != active {
					cache[runningID].dirty = true
					cancel()
				}
				show(true) // switch to this document's cached PDF immediately
			}
			request.reply <- nil
			start()
		case <-changes:
			// Only compile when the editor finishes its current write burst.
			// All cached PDFs remain visible, but need refreshing on next use.
			for _, entry := range cache {
				entry.dirty = true
			}
			if running {
				cancel()
			}
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(110 * time.Millisecond)
			debounce = timer.C
			show(false)
		case <-debounce:
			debounce = nil
			found, err := latex.ScanDocuments(project.Root, project.Pinned)
			if err != nil {
				_, _ = fmt.Fprintf(s.stdout, "  ! Could not rescan LaTeX documents: %v\n", err)
				start()
				continue
			}
			documents = found
			present := make(map[string]bool, len(documents))
			for _, document := range documents {
				present[document.ID] = true
				if cache[document.ID] == nil {
					cache[document.ID] = &cachedDocument{dirty: true}
				}
			}
			for id := range cache {
				if !present[id] && id != runningID {
					delete(cache, id)
				}
			}
			previousActive := active
			if !present[active] {
				active = ""
				if len(documents) != 0 {
					active = chooseDocument(documents)
				}
			}
			// All known PDFs remain cached, but the currently selected file
			// is always refreshed after a filesystem change.
			show(active != previousActive)
			start()
		case done := <-completed:
			running = false
			runningID = ""
			if cancel != nil {
				cancel()
				cancel = nil
			}
			entry := cache[done.id]
			// A canceled build or another write must never replace a newer
			// cached document. In particular, do not publish its stale PDF.
			valid := entry != nil && !entry.dirty && done.result.Error != "Build superseded by a newer edit"
			if valid {
				entry.compiled = true
				entry.result = done.result
				entry.updatedAt = time.Now()
				if done.result.Error == "" {
					entry.pdf = done.result.PDF
				}
				if active == done.id && debounce == nil {
					show(done.result.Error == "")
				}
				if done.result.Error != "" {
					_, _ = fmt.Fprintf(s.stdout, "  ✗ %s (%s)\n", done.result.Error, done.id)
				} else {
					_, _ = fmt.Fprintf(s.stdout, "  ✓ %s ready (%d ms)\n", done.id, done.result.Duration.Milliseconds())
				}
			}
			if entry != nil && !valid {
				entry.dirty = true
			}
			if entry != nil {
				stillPresent := false
				for _, document := range documents {
					if document.ID == done.id {
						stillPresent = true
						break
					}
				}
				if !stillPresent {
					delete(cache, done.id)
				}
			}
			// If the user switched while the old process was being canceled,
			// launch the new selection now instead of showing a stale result.
			start()
		}
	}
}

func chooseDocument(documents []latex.Document) string {
	for _, document := range documents {
		if document.ID == "main.tex" {
			return document.ID
		}
	}
	return documents[0].ID
}
