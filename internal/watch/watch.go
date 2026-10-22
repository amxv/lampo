// Package watch provides low-latency, recursive project change notifications.
package watch

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var excludedDirs = map[string]bool{
	".git": true, ".svn": true, ".hg": true, ".lampo": true,
	".build": true, "node_modules": true, "dist": true, "target": true,
	"__pycache__": true, ".venv": true,
}

var watchedExtensions = map[string]bool{
	".tex": true, ".bib": true, ".cls": true, ".sty": true,
	".bst": true, ".bbx": true, ".cbx": true, ".def": true,
	".cfg": true, ".clo": true, ".fd": true, ".tikz": true,
	".pgf": true, ".eps": true, ".ps": true, ".svg": true,
	".pdf": true, ".png": true, ".jpg": true, ".jpeg": true,
	".webp": true, ".csv": true, ".dat": true, ".txt": true,
	".json": true, ".yaml": true, ".yml": true,
}

// Monitor watches directories, not individual files, so atomic editor saves
// cannot accidentally detach a watch from the source file.
type Monitor struct {
	fs      *fsnotify.Watcher
	root    string
	dirs    map[string]bool
	changes chan struct{}
	stderr  io.Writer
}

func New(root string, stderr io.Writer) (*Monitor, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	fs, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create project watcher: %w", err)
	}
	m := &Monitor{fs: fs, root: abs, dirs: make(map[string]bool), changes: make(chan struct{}, 1), stderr: stderr}
	if err := m.addTree(abs); err != nil {
		_ = fs.Close()
		return nil, err
	}
	return m, nil
}

func (m *Monitor) addTree(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// A project may contain unreadable or ephemeral unrelated directories.
			if path == root {
				return err
			}
			return filepath.SkipDir
		}
		if !d.IsDir() {
			return nil
		}
		if path != m.root && excludedDirs[d.Name()] {
			return filepath.SkipDir
		}
		if !m.dirs[path] {
			if err := m.fs.Add(path); err != nil {
				return fmt.Errorf("watch %s: %w", path, err)
			}
			m.dirs[path] = true
		}
		return nil
	})
}

func (m *Monitor) Changes() <-chan struct{} { return m.changes }

// Close is used if server startup fails before the monitor loop begins.
func (m *Monitor) Close() error { return m.fs.Close() }

// Run should be called once. It exits when ctx is canceled and closes its watches.
func (m *Monitor) Run(ctx context.Context) {
	defer m.fs.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-m.fs.Events:
			if !ok {
				return
			}
			m.handle(event)
		case err, ok := <-m.fs.Errors:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(m.stderr, "Lampo watcher: %v\n", err)
		}
	}
}

func (m *Monitor) handle(e fsnotify.Event) {
	if !e.Has(fsnotify.Create) && !e.Has(fsnotify.Write) && !e.Has(fsnotify.Remove) && !e.Has(fsnotify.Rename) {
		return
	}
	if excluded(e.Name, m.root) {
		return
	}
	if e.Has(fsnotify.Remove) || e.Has(fsnotify.Rename) {
		for path := range m.dirs {
			if path == e.Name || strings.HasPrefix(path, e.Name+string(filepath.Separator)) {
				delete(m.dirs, path)
			}
		}
	}
	if e.Has(fsnotify.Create) {
		if info, err := os.Stat(e.Name); err == nil && info.IsDir() {
			if err := m.addTree(e.Name); err != nil {
				_, _ = fmt.Fprintf(m.stderr, "Lampo watcher: %v\n", err)
			}
			m.notify()
			return
		}
	}
	if watchedExtensions[strings.ToLower(filepath.Ext(e.Name))] {
		m.notify()
	}
}

func (m *Monitor) notify() {
	select {
	case m.changes <- struct{}{}:
	default:
	}
}

func excluded(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return true
	}
	for _, segment := range strings.Split(filepath.Dir(rel), string(filepath.Separator)) {
		if excludedDirs[segment] {
			return true
		}
	}
	return false
}
