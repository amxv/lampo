// Package preview serves a self-contained, local browser UI for LaTeX projects.
package preview

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amxv/lampo/internal/latex"
	"github.com/amxv/lampo/internal/watch"
)

//go:embed web/* web/vendor/*
var ui embed.FS

type Options struct {
	Path   string
	Host   string
	Port   int
	Engine string
	Open   bool
}

type Snapshot struct {
	Sequence   uint64           `json:"sequence"`
	Status     string           `json:"status"`
	Revision   uint64           `json:"revision"`
	Filename   string           `json:"filename"`
	ActiveID   string           `json:"activeId"`
	Documents  []latex.Document `json:"documents"`
	Error      string           `json:"error,omitempty"`
	Log        string           `json:"log,omitempty"`
	DurationMs int64            `json:"durationMs,omitempty"`
	UpdatedAt  time.Time        `json:"updatedAt,omitempty"`
	HasPDF     bool             `json:"hasPdf"`
}

type server struct {
	mu             sync.RWMutex
	state          Snapshot
	pdf            []byte
	clients        map[chan Snapshot]struct{}
	stdout         io.Writer
	assetID        string
	selectDocument chan selectRequest
	ctx            context.Context
}

func newServer(source string, stdout io.Writer) *server {
	return &server{
		state:          Snapshot{Status: "building", Filename: filepath.Base(source), ActiveID: filepath.Base(source), Documents: []latex.Document{{ID: filepath.Base(source), Name: filepath.Base(source)}}},
		clients:        make(map[chan Snapshot]struct{}),
		stdout:         stdout,
		assetID:        strconv.FormatInt(time.Now().UnixNano(), 36),
		selectDocument: make(chan selectRequest),
		ctx:            context.Background(),
	}
}

// Serve listens only on localhost by default, and removes all generated TeX
// files when stopped. It never writes into the source project.
func Serve(ctx context.Context, options Options, stdout, stderr io.Writer) error {
	project, err := latex.OpenProject(options.Path)
	if err != nil {
		return err
	}
	source := filepath.Join(project.Root, filepath.FromSlash(project.InitialID))
	builder := latex.Builder{Source: source, Engine: options.Engine}
	if err := builder.Check(); err != nil {
		return err
	}
	output, err := os.MkdirTemp("", "lampo-build-*")
	if err != nil {
		return fmt.Errorf("create temporary build directory: %w", err)
	}
	defer os.RemoveAll(output)
	builder.OutputDir = output
	monitor, err := watch.New(project.Root, stderr)
	if err != nil {
		return err
	}
	// The listener is acquired before starting any goroutines: if the port is
	// already occupied, fail cleanly rather than leaving orphan background work.
	listener, err := net.Listen("tcp", net.JoinHostPort(options.Host, strconv.Itoa(options.Port)))
	if err != nil {
		_ = monitor.Close()
		return fmt.Errorf("start preview listener: %w", err)
	}
	defer listener.Close()
	viewer := newServer(source, stdout)
	viewer.ctx = ctx
	viewer.state.ActiveID = project.InitialID
	viewer.state.Documents = project.Documents
	viewer.state.Filename = filepath.Base(source)
	httpServer := &http.Server{
		Handler:           viewer.routes(options.Host),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
	addr := listener.Addr().(*net.TCPAddr)
	visitHost := options.Host
	if visitHost == "0.0.0.0" {
		visitHost = "127.0.0.1"
	} else if visitHost == "::" {
		visitHost = "::1"
	}
	url := "http://" + net.JoinHostPort(visitHost, strconv.Itoa(addr.Port))
	_, _ = fmt.Fprintf(stdout, "\n  ⚡ Lampo live preview\n  %s\n\n  Source: %s\n  Engine: %s\n  Watching for changes. Press Ctrl+C to stop.\n\n", url, source, options.Engine)
	go monitor.Run(ctx)
	go viewer.buildLoop(ctx, project, output, builder.Engine, monitor.Changes())
	serveDone := make(chan error, 1)
	go func() { serveDone <- httpServer.Serve(listener) }()
	if options.Open {
		if err := openBrowser(url); err != nil {
			_, _ = fmt.Fprintf(stderr, "Could not open browser automatically: %v\n", err)
		}
	}
	var serveErr error
	select {
	case <-ctx.Done():
	case serveErr = <-serveDone:
	}
	shutdownCtx, stop := context.WithTimeout(context.Background(), 3*time.Second)
	defer stop()
	_ = httpServer.Shutdown(shutdownCtx)
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return serveErr
	}
	return nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func (s *server) publish(next Snapshot, pdf []byte) {
	s.mu.Lock()
	next.Sequence = s.state.Sequence + 1
	if next.ActiveID != s.state.ActiveID {
		// Never display a cached PDF from a *different* document while the
		// selected entrypoint is still compiling.
		s.pdf = nil
		s.state.Revision++
	}
	if pdf != nil {
		s.pdf = pdf
		s.state.Revision++
	}
	next.Revision = s.state.Revision
	next.HasPDF = len(s.pdf) > 0
	if next.ActiveID != "" {
		next.Filename = filepath.Base(filepath.FromSlash(next.ActiveID))
	} else {
		next.Filename = ""
	}
	s.state = next
	for ch := range s.clients {
		select {
		case ch <- next:
		default:
			// Replace a stale queued status instead of blocking compilation on
			// a slow or suspended browser tab.
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- next:
			default:
			}
		}
	}
	s.mu.Unlock()
}

func (s *server) snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *server) routes(host string) http.Handler {
	site, _ := fs.Sub(ui, "web")
	files := http.FileServer(http.FS(site))
	index, _ := fs.ReadFile(site, "index.html")
	index = []byte(strings.ReplaceAll(string(index), "__LAMPO_ASSET_VERSION__", s.assetID))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s.snapshot())
	})
	mux.HandleFunc("POST /api/select", s.selectHandler)
	mux.HandleFunc("GET /events", s.events)
	mux.HandleFunc("GET /preview.pdf", s.servePDF(false))
	mux.HandleFunc("GET /api/export", s.servePDF(true))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(index)
			return
		}
		if strings.HasSuffix(r.URL.Path, ".mjs") {
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		}
		files.ServeHTTP(w, r)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; font-src 'self' data:; img-src 'self' data: blob:; connect-src 'self'; worker-src 'self' blob:; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		// Prevent a random website from rebinding a hostile hostname to a
		// loopback service and reading a private document over HTTP.
		if !allowedHost(r.Host, host) {
			http.Error(w, "invalid preview host", http.StatusForbidden)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

type selectRequest struct {
	id    string
	reply chan error
}

func (s *server) selectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	// This is a localhost-only state mutation, not a filesystem path API.
	// Check Origin and content type so cross-site forms cannot switch documents.
	if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host && origin != "https://"+r.Host {
		http.Error(w, `{"error":"invalid origin"}`, http.StatusForbidden)
		return
	}
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, `{"error":"expected application/json"}`, http.StatusUnsupportedMediaType)
		return
	}
	var request struct {
		ID string `json:"id"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&request) != nil || request.ID == "" {
		http.Error(w, `{"error":"expected a document id"}`, http.StatusBadRequest)
		return
	}
	message := selectRequest{id: request.ID, reply: make(chan error, 1)}
	select {
	case s.selectDocument <- message:
	case <-s.ctx.Done():
		http.Error(w, `{"error":"preview is shutting down"}`, http.StatusServiceUnavailable)
		return
	case <-r.Context().Done():
		return
	}
	select {
	case err := <-message.reply:
		if err != nil {
			http.Error(w, `{"error":"document not found"}`, http.StatusNotFound)
			return
		}
	case <-s.ctx.Done():
		http.Error(w, `{"error":"preview is shutting down"}`, http.StatusServiceUnavailable)
		return
	case <-r.Context().Done():
		return
	}
	_ = json.NewEncoder(w).Encode(s.snapshot())
}

func allowedHost(requestHost, bindHost string) bool {
	host, _, err := net.SplitHostPort(requestHost)
	if err != nil {
		host = requestHost
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	// An explicit wildcard bind permits direct IP access from a trusted LAN,
	// but still rejects arbitrary DNS names to limit DNS rebinding attacks.
	if bindHost == "0.0.0.0" || bindHost == "::" {
		return net.ParseIP(host) != nil
	}
	return host == bindHost && bindHost != "0.0.0.0" && bindHost != "::"
}

func (s *server) servePDF(download bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		pdf := s.pdf
		name := strings.TrimSuffix(s.state.Filename, filepath.Ext(s.state.Filename)) + ".pdf"
		s.mu.RUnlock()
		if len(pdf) == 0 {
			http.Error(w, "no successfully compiled PDF yet", http.StatusNotFound)
			return
		}
		if download {
			name = safeFilename(r.URL.Query().Get("filename"), name)
		}
		kind := "inline"
		if download {
			kind = "attachment"
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Disposition", mime.FormatMediaType(kind, map[string]string{"filename": name}))
		w.Header().Set("Content-Length", strconv.Itoa(len(pdf)))
		_, _ = w.Write(pdf)
	}
}

func safeFilename(input, fallback string) string {
	input = strings.TrimSpace(input)
	input = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 || r == '/' || r == '\\' || r == ':' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, input)
	input = strings.Trim(input, ". ")
	if input == "" || input == "_" || input == ".." {
		return fallback
	}
	if !strings.HasSuffix(strings.ToLower(input), ".pdf") {
		input += ".pdf"
	}
	if len(input) > 200 {
		return fallback
	}
	return input
}

func (s *server) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	ch := make(chan Snapshot, 1)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	initial := s.state
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
	}()
	write := func(snapshot Snapshot) bool {
		data, _ := json.Marshal(snapshot)
		_, err := fmt.Fprintf(w, "event: state\ndata: %s\n\n", data)
		if err == nil {
			flusher.Flush()
		}
		return err == nil
	}
	if !write(initial) {
		return
	}
	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case snapshot := <-ch:
			if !write(snapshot) {
				return
			}
		case <-keepalive.C:
			_, _ = io.WriteString(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
