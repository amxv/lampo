// Package latex locates LaTeX projects and builds them with latexmk.
package latex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Result is an immutable compilation result.
type Result struct {
	PDF      []byte
	Log      string
	Duration time.Duration
	Error    string
}

// Builder runs a local TeX installation, without installing packages or
// sending project data to an external service.
type Builder struct {
	Source    string
	OutputDir string
	Engine    string
	Binary    string // injectable for tests; defaults to latexmk
}

// FindSource resolves a path to a .tex document root.
func FindSource(path string) (string, error) {
	project, err := OpenProject(path)
	if err != nil {
		return "", err
	}
	return filepath.Join(project.Root, filepath.FromSlash(project.InitialID)), nil
}

// Check makes the error for missing TeX installations actionable before launch.
func (b Builder) Check() error {
	binary := b.Binary
	if binary == "" {
		binary = "latexmk"
	}
	if _, err := exec.LookPath(binary); err != nil {
		return fmt.Errorf("latexmk was not found on PATH; install TeX Live/MacTeX or MiKTeX (including latexmk): %w", err)
	}
	if b.Binary == "" {
		if _, err := exec.LookPath(b.Engine); err != nil {
			return fmt.Errorf("%s was not found on PATH; install a LaTeX distribution or select a different --engine: %w", b.Engine, err)
		}
	}
	return nil
}

// Build returns compiled PDF bytes. Failures carry a short human-readable error
// and the compiler log; the preview server keeps the last successful PDF.
func (b Builder) Build(ctx context.Context) Result {
	start := time.Now()
	mode := map[string]string{"pdflatex": "-pdf", "xelatex": "-xelatex", "lualatex": "-lualatex"}[b.Engine]
	if mode == "" {
		return Result{Error: fmt.Sprintf("unsupported LaTeX engine %q", b.Engine)}
	}
	binary := b.Binary
	if binary == "" {
		binary = "latexmk"
	}
	cmd := exec.CommandContext(ctx, binary,
		mode, "-interaction=nonstopmode", "-halt-on-error", "-file-line-error",
		"-outdir="+b.OutputDir, filepath.Base(b.Source))
	cmd.Dir = filepath.Dir(b.Source)
	var log tailBuffer
	cmd.Stdout = &log
	cmd.Stderr = &log
	err := cmd.Run()
	result := Result{Log: log.String(), Duration: time.Since(start)}
	if ctx.Err() != nil {
		result.Error = "Build superseded by a newer edit"
		return result
	}
	if err != nil {
		result.Error = summarizeError(err, result.Log)
		return result
	}
	file := filepath.Join(b.OutputDir, strings.TrimSuffix(filepath.Base(b.Source), filepath.Ext(b.Source))+".pdf")
	result.PDF, err = os.ReadFile(file)
	if err != nil {
		result.Error = fmt.Sprintf("compiler succeeded but could not read the PDF: %v", err)
	} else if !bytes.HasPrefix(result.PDF, []byte("%PDF-")) {
		result.PDF = nil
		result.Error = "the compiler did not produce a valid PDF"
	}
	return result
}

func summarizeError(err error, log string) string {
	lines := strings.Split(log, "\n")
	// -file-line-error writes the most useful diagnostic as
	// ./sections/introduction.tex:42: Undefined control sequence.
	fileError := regexp.MustCompile(`(?i)(?:^|[ /\\])[^\s:]+\.(?:tex|sty|cls|bib):[0-9]+:\s+.+`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if fileError.MatchString(line) {
			return line
		}
		if strings.HasPrefix(line, "! ") && !strings.Contains(line, "==> Fatal error") {
			return strings.TrimPrefix(line, "! ")
		}
		if strings.Contains(line, ": error:") {
			return line
		}
	}
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "Error:") {
			return line
		}
	}
	if errors.Is(err, exec.ErrNotFound) {
		return "latexmk is not installed or not on PATH"
	}
	return fmt.Sprintf("LaTeX build failed: %v (see build log)", err)
}

// tailBuffer bounds the memory and SSE payload cost of verbose compiler output.
type tailBuffer struct{ data []byte }

func (b *tailBuffer) Write(p []byte) (int, error) {
	const limit = 96 * 1024
	n := len(p)
	b.data = append(b.data, p...)
	if len(b.data) > limit {
		b.data = append([]byte(nil), b.data[len(b.data)-limit:]...)
	}
	return n, nil
}

func (b *tailBuffer) String() string { return string(b.data) }
