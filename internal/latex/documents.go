package latex

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Document is a standalone LaTeX entrypoint, identified only by its path
// relative to the watched directory. Absolute local paths are never sent to
// the browser.
type Document struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	Root      string
	InitialID string
	Documents []Document
	// Pinned is the explicitly provided source file, which remains selectable
	// even when it is not identifiable by a \documentclass directive.
	Pinned string
}

var documentDirective = regexp.MustCompile(`\\(?:documentclass|documentstyle)\b`)

var ignoredDirectories = map[string]bool{
	".git": true, ".svn": true, ".hg": true, ".lampo": true,
	".build": true, "node_modules": true, "dist": true, "target": true,
	"__pycache__": true, ".venv": true,
}

// OpenProject resolves a CLI path and chooses an entrypoint. An ambiguous
// directory no longer fails: main.tex wins, then alphabetically first root.
func OpenProject(path string) (Project, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Project{}, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return Project{}, fmt.Errorf("cannot find LaTeX project %q: %w", path, err)
	}
	project := Project{Root: abs}
	if !info.IsDir() {
		if strings.ToLower(filepath.Ext(abs)) != ".tex" {
			return Project{}, fmt.Errorf("expected a .tex file, got %q", path)
		}
		project.Root = filepath.Dir(abs)
		project.Pinned = filepath.Base(abs)
	}
	project.Root, err = filepath.EvalSymlinks(project.Root)
	if err != nil {
		return Project{}, err
	}
	project.Documents, err = ScanDocuments(project.Root, project.Pinned)
	if err != nil {
		return Project{}, err
	}
	if len(project.Documents) == 0 {
		return Project{}, fmt.Errorf("no standalone .tex documents found in %q; pass a path to a .tex file", project.Root)
	}
	project.InitialID = project.Pinned
	if project.InitialID == "" {
		project.InitialID = chooseInitial(project.Documents)
	}
	return project, nil
}

// ScanDocuments includes entrypoints in nested folders while omitting helper
// snippets (\input / \include) and hidden/generated directories. If there is
// exactly one .tex file, it remains usable for backwards compatibility.
func ScanDocuments(root, pinned string) ([]Document, error) {
	var roots, candidates []Document
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if path == root {
				return walkErr
			}
			return nil
		}
		if entry.IsDir() {
			if path != root && (ignoredDirectories[entry.Name()] || strings.HasPrefix(entry.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || strings.ToLower(filepath.Ext(entry.Name())) != ".tex" {
			return nil
		}
		id, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		id = filepath.ToSlash(id)
		doc := Document{ID: id, Name: entry.Name()}
		candidates = append(candidates, doc)
		if id == pinned || hasDocumentClass(path) {
			roots = append(roots, doc)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		// A source that delegates its preamble to \input is also a valid root.
		for _, candidate := range candidates {
			if candidate.ID == "main.tex" || candidate.ID == pinned {
				roots = append(roots, candidate)
			}
		}
		if len(roots) == 0 && len(candidates) == 1 {
			roots = candidates
		}
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].ID < roots[j].ID })
	return roots, nil
}

func chooseInitial(docs []Document) string {
	for _, doc := range docs {
		if doc.ID == "main.tex" {
			return doc.ID
		}
	}
	return docs[0].ID
}

func hasDocumentClass(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	reader := bufio.NewReader(io.LimitReader(f, 1024*1024))
	for {
		line, err := reader.ReadBytes('\n')
		// A percent sign starts a comment unless escaped by a backslash.
		for i, b := range line {
			if b == '%' {
				slashes := 0
				for j := i - 1; j >= 0 && line[j] == '\\'; j-- {
					slashes++
				}
				if slashes%2 == 0 {
					line = line[:i]
					break
				}
			}
		}
		if documentDirective.Match(bytes.TrimSpace(line)) {
			return true
		}
		if err != nil {
			return false
		}
	}
}
