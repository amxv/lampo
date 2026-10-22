package preview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/amxv/lampo/internal/latex"
)

// This fake latexmk builds an identifiable PDF without requiring an actual
// TeX installation in the Go CI environment. It verifies the switching,
// cache, discovery and HTTP behaviors against the real coordinator.
func TestMultiDocumentSession(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake shell compiler is Unix-only")
	}
	root := t.TempDir()
	bin := t.TempDir()
	buildLog := filepath.Join(root, "build-count.log")
	compiler := `#!/bin/sh
outdir=""
source=""
for argument in "$@"; do
  case "$argument" in
    -outdir=*) outdir="${argument#-outdir=}" ;;
    *.tex) source="$argument" ;;
  esac
done
printf '%s\n' "$source" >> "$LAMPO_TEST_BUILDLOG"
base="${source%.tex}"
printf '%%PDF-1.4\ncompiled: %s\n' "$source" > "$outdir/$base.pdf"
`
	if err := os.WriteFile(filepath.Join(bin, "latexmk"), []byte(compiler), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("LAMPO_TEST_BUILDLOG", buildLog)
	write := func(id string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(id))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("\\documentclass{article}\n\\begin{document}\nTest\\end{document}"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("main.tex")
	write("letter.tex")
	project, err := latex.OpenProject(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	viewer := newServer(filepath.Join(root, "main.tex"), io.Discard)
	viewer.state.Documents = project.Documents
	viewer.ctx = ctx
	changes := make(chan struct{}, 10)
	buildRoot := t.TempDir()
	done := make(chan struct{})
	go func() {
		viewer.buildLoop(ctx, project, buildRoot, "pdflatex", changes)
		close(done)
	}()
	server := httptest.NewServer(viewer.routes("127.0.0.1"))
	t.Cleanup(func() {
		cancel()
		<-done
		server.Close()
	})

	get := func() Snapshot {
		t.Helper()
		resp, err := http.Get(server.URL + "/api/status")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var state Snapshot
		if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
			t.Fatal(err)
		}
		return state
	}
	wait := func(condition func(Snapshot) bool) Snapshot {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			state := get()
			if condition(state) {
				return state
			}
			time.Sleep(8 * time.Millisecond)
		}
		t.Fatalf("timed out waiting for session state: %+v", get())
		return Snapshot{}
	}
	selectDocument := func(id string, expected int) Snapshot {
		t.Helper()
		payload, _ := json.Marshal(map[string]string{"id": id})
		resp, err := http.Post(server.URL+"/api/select", "application/json", bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != expected {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("select %q: HTTP %d, wanted %d: %s", id, resp.StatusCode, expected, body)
		}
		if expected != 200 {
			return Snapshot{}
		}
		var state Snapshot
		if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
			t.Fatal(err)
		}
		return state
	}
	buildCount := func() int {
		t.Helper()
		data, err := os.ReadFile(buildLog)
		if err != nil {
			return 0
		}
		return len(strings.Split(strings.TrimSpace(string(data)), "\n"))
	}

	first := wait(func(s Snapshot) bool { return s.Status == "ready" && s.ActiveID == "main.tex" })
	if len(first.Documents) != 2 || buildCount() != 1 {
		t.Fatalf("initial project incorrect: %+v, builds: %d", first, buildCount())
	}
	if state := selectDocument("../other.tex", 404); state.ActiveID != "" {
		t.Fatal("untrusted file path was accepted")
	}
	selectDocument("letter.tex", 200)
	letter := wait(func(s Snapshot) bool { return s.Status == "ready" && s.ActiveID == "letter.tex" })
	if buildCount() != 2 || letter.Revision <= first.Revision {
		t.Fatalf("second document wasn't compiled independently: %+v, builds: %d", letter, buildCount())
	}
	// Selecting an already compiled document is immediate and does not rerun
	// latexmk, even when both documents have the same main filename.
	cached := selectDocument("main.tex", 200)
	if cached.Status != "ready" || !cached.HasPDF || buildCount() != 2 {
		t.Fatalf("cached selection rebuilt instead of reusing PDF: %+v, builds: %d", cached, buildCount())
	}
	resp, err := http.Get(server.URL + "/preview.pdf")
	if err != nil {
		t.Fatal(err)
	}
	pdf, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !bytes.Contains(pdf, []byte("compiled: main.tex")) {
		t.Fatalf("wrong PDF served after switching: %q", pdf)
	}

	write("nested/report.tex")
	changes <- struct{}{}
	updated := wait(func(s Snapshot) bool { return len(s.Documents) == 3 })
	if updated.ActiveID != "main.tex" {
		t.Fatalf("new file changed selection: %+v", updated)
	}
	selectDocument("nested/report.tex", 200)
	wait(func(s Snapshot) bool { return s.ActiveID == "nested/report.tex" && s.Status == "ready" })
	if err := os.Remove(filepath.Join(root, "nested/report.tex")); err != nil {
		t.Fatal(err)
	}
	changes <- struct{}{}
	wait(func(s Snapshot) bool { return s.ActiveID == "main.tex" && len(s.Documents) == 2 && s.Status == "ready" })

	// When the last file is removed the server stays alive, and discovery
	// resumes as soon as a new document appears.
	for _, id := range []string{"main.tex", "letter.tex"} {
		if err := os.Remove(filepath.Join(root, id)); err != nil {
			t.Fatal(err)
		}
	}
	changes <- struct{}{}
	wait(func(s Snapshot) bool {
		return s.Status == "empty" && s.ActiveID == "" && len(s.Documents) == 0 && !s.HasPDF
	})
	write("new.tex")
	changes <- struct{}{}
	wait(func(s Snapshot) bool { return s.Status == "ready" && s.ActiveID == "new.tex" && len(s.Documents) == 1 })

	bad, _ := http.NewRequest("POST", server.URL+"/api/select", strings.NewReader(`{"id":"new.tex"}`))
	bad.Header.Set("Content-Type", "application/json")
	bad.Header.Set("Origin", "https://evil.example")
	badResp, err := http.DefaultClient.Do(bad)
	if err != nil {
		t.Fatal(err)
	}
	badResp.Body.Close()
	if badResp.StatusCode != 403 {
		t.Fatal(fmt.Sprintf("cross-origin mutation was permitted: %s", badResp.Status))
	}
}
