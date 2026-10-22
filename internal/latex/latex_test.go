package latex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindSource(t *testing.T) {
	dir := t.TempDir()
	if _, err := FindSource(dir); err == nil {
		t.Fatal("expected an error for empty dir")
	}
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("intro.tex", "Hello")
	if got, err := FindSource(dir); err != nil || filepath.Base(got) != "intro.tex" {
		t.Fatalf("single file: %q, %v", got, err)
	}
	write("paper.tex", `\documentclass{article}`)
	if got, err := FindSource(dir); err != nil || filepath.Base(got) != "paper.tex" {
		t.Fatalf("root detection: %q, %v", got, err)
	}
	write("other.tex", `\documentclass{article}`)
	if got, err := FindSource(dir); err != nil || filepath.Base(got) != "other.tex" {
		t.Fatalf("deterministic root selection: %q, %v", got, err)
	}
	write("main.tex", `\documentclass{article}`)
	if got, err := FindSource(dir); err != nil || filepath.Base(got) != "main.tex" {
		t.Fatalf("main detection: %q, %v", got, err)
	}
}

func TestSummarizeError(t *testing.T) {
	message := summarizeError(os.ErrInvalid, "something\n! LaTeX Error: Missing font.\nmore logs")
	if message != "LaTeX Error: Missing font." {
		t.Fatalf("unexpected: %s", message)
	}
	line := summarizeError(os.ErrInvalid, "something\n./sections/chapter.tex:37: Undefined control sequence.\nmore logs")
	if line != "./sections/chapter.tex:37: Undefined control sequence." {
		t.Fatalf("expected a source line in the error, got %q", line)
	}
}

func TestTailBufferIsBounded(t *testing.T) {
	var b tailBuffer
	input := strings.Repeat("A", 200000) + "end"
	_, _ = b.Write([]byte(input))
	if len(b.String()) != 96*1024 || !strings.HasSuffix(b.String(), "end") {
		t.Fatal("incorrect log tail")
	}
}
