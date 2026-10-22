package preview

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRoutesAndPDFExport(t *testing.T) {
	viewer := newServer("example.tex", io.Discard)
	httpServer := httptest.NewServer(viewer.routes("127.0.0.1"))
	defer httpServer.Close()

	resp, err := http.Get(httpServer.URL + "/api/status")
	if err != nil {
		t.Fatal(err)
	}
	var state Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if state.Status != "building" || state.HasPDF || state.Filename != "example.tex" {
		t.Fatalf("initial state: %+v", state)
	}

	viewer.publish(Snapshot{Status: "ready", UpdatedAt: time.Now()}, []byte("%PDF-1.4\nfixture"))
	resp, err = http.Get(httpServer.URL + "/preview.pdf")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !bytes.HasPrefix(data, []byte("%PDF-")) || resp.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("bad PDF response: %s, %q", resp.Status, data)
	}
	resp, err = http.Get(httpServer.URL + "/api/export?filename=final-review.pdf")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Content-Disposition"), "final-review.pdf") || !strings.Contains(resp.Header.Get("Content-Disposition"), "attachment") {
		t.Fatalf("bad export disposition: %s", resp.Header.Get("Content-Disposition"))
	}

	viewer.publish(Snapshot{Status: "error", Error: "broken TeX"}, nil)
	resp, err = http.Get(httpServer.URL + "/preview.pdf")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("last good PDF was lost after build error: %s", resp.Status)
	}
}

func TestWebAssetsAndHostSafety(t *testing.T) {
	viewer := newServer("example.tex", io.Discard)
	httpServer := httptest.NewServer(viewer.routes("127.0.0.1"))
	defer httpServer.Close()
	for _, path := range []string{"/", "/app.js", "/style.css", "/vendor/pdf.mjs", "/vendor/pdf.worker.mjs"} {
		resp, err := http.Get(httpServer.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 || resp.Header.Get("X-Content-Type-Options") != "nosniff" || resp.Header.Get("Cache-Control") != "no-store" {
			t.Errorf("asset %s: %s", path, resp.Status)
		}
	}
	root, err := http.Get(httpServer.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	html, err := io.ReadAll(root.Body)
	root.Body.Close()
	if err != nil || bytes.Contains(html, []byte("__LAMPO_ASSET_VERSION__")) || !bytes.Contains(html, []byte("/app.js?v=")) {
		t.Fatalf("asset version was not rendered into HTML: %v", err)
	}
	req, _ := http.NewRequest("GET", httpServer.URL+"/api/status", nil)
	req.Host = "malicious.example"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected host validation: %s", resp.Status)
	}
}

func TestSafeFilename(t *testing.T) {
	cases := map[string]string{
		"": "default.pdf", "report": "report.pdf", "hey.PDF": "hey.PDF",
		"../../secrets": "_.._secrets.pdf", "notes\r\nheader": "notes__header.pdf",
	}
	for input, want := range cases {
		if got := safeFilename(input, "default.pdf"); got != want {
			t.Errorf("safeFilename(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestAllowedHosts(t *testing.T) {
	if !allowedHost("127.0.0.1:7777", "127.0.0.1") {
		t.Fatal("loopback host rejected")
	}
	if allowedHost("attacker.example:7777", "127.0.0.1") {
		t.Fatal("DNS rebinding hostname accepted")
	}
	if !allowedHost("192.168.0.20:7777", "0.0.0.0") {
		t.Fatal("explicit LAN listener rejected a direct IP")
	}
	if allowedHost("attacker.example:7777", "0.0.0.0") {
		t.Fatal("wildcard listener accepted an untrusted DNS hostname")
	}
}
