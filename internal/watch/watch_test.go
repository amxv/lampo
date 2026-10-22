package watch

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRelevantPaths(t *testing.T) {
	root := filepath.Join("tmp", "project")
	for _, name := range []string{".git/HEAD", "dist/preview.pdf", "node_modules/index.tex"} {
		if !excluded(filepath.Join(root, name), root) {
			t.Fatalf("should exclude %s", name)
		}
	}
	if excluded(filepath.Join(root, "figures", "plot.pdf"), root) {
		t.Fatal("excluded relevant figures directory")
	}
}

func TestAtomicSaveAndNestedDirectories(t *testing.T) {
	root := t.TempDir()
	m, err := New(root, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { m.Run(ctx); close(done) }()

	// A nested directory added after startup must be watched.
	nested := filepath.Join(root, "sections")
	if err := os.Mkdir(nested, 0700); err != nil {
		t.Fatal(err)
	}
	assertChange(t, m.Changes())
	path := filepath.Join(nested, "intro.tex")
	if err := os.WriteFile(path, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	assertChange(t, m.Changes())
	if err := os.WriteFile(filepath.Join(nested, ".intro.swp"), []byte("updated"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(nested, ".intro.swp"), path); err != nil {
		t.Fatal(err)
	}
	assertChange(t, m.Changes())
	cancel()
	<-done
}

func assertChange(t *testing.T, events <-chan struct{}) {
	t.Helper()
	select {
	case <-events:
	case <-time.After(3 * time.Second):
		t.Fatal("watcher missed file change")
	}
}
