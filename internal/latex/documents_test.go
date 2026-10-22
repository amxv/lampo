package latex

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDocumentDiscovery(t *testing.T) {
	root := t.TempDir()
	write := func(path, source string) {
		t.Helper()
		file := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("main.tex", "\\documentclass{article}\n\\input{parts/introduction}")
	write("letter.tex", "% \\documentclass{article}\n\\documentclass[12pt]{letter}")
	write("parts/introduction.tex", "Here is an included chapter.")
	write("drafts/appendix.tex", "% only a helper\n")
	write("drafts/paper.tex", "  % comment\\documentclass{article}\n\\documentstyle{article}")
	write(".hidden/secret.tex", "\\documentclass{article}")
	write("node_modules/garbage.tex", "\\documentclass{article}")
	write("commented.tex", "% \\documentclass{article}\nThis is only a helper.")
	project, err := OpenProject(root)
	if err != nil {
		t.Fatal(err)
	}
	if project.InitialID != "main.tex" {
		t.Fatalf("main document not selected: %s", project.InitialID)
	}
	want := []Document{
		{ID: "drafts/paper.tex", Name: "paper.tex"},
		{ID: "letter.tex", Name: "letter.tex"},
		{ID: "main.tex", Name: "main.tex"},
	}
	if !reflect.DeepEqual(project.Documents, want) {
		t.Fatalf("documents = %#v, want %#v", project.Documents, want)
	}
	pinned, err := OpenProject(filepath.Join(root, "parts", "introduction.tex"))
	if err != nil {
		t.Fatal(err)
	}
	if pinned.InitialID != "introduction.tex" || len(pinned.Documents) != 1 {
		t.Fatalf("explicit helper should be selectable: %+v", pinned)
	}
}

func TestDocumentChangesAndEmptyFolder(t *testing.T) {
	root := t.TempDir()
	if _, err := OpenProject(root); err == nil {
		t.Fatal("expected an actionable error for empty directory")
	}
	a := filepath.Join(root, "a.tex")
	if err := os.WriteFile(a, []byte("plain LaTeX"), 0600); err != nil {
		t.Fatal(err)
	}
	docs, err := ScanDocuments(root, "")
	if err != nil || len(docs) != 1 || docs[0].ID != "a.tex" {
		t.Fatalf("legacy single-file fallback: %+v, %v", docs, err)
	}
	b := filepath.Join(root, "b.tex")
	if err := os.WriteFile(b, []byte("\\documentclass{article}"), 0600); err != nil {
		t.Fatal(err)
	}
	docs, err = ScanDocuments(root, "")
	if err != nil || len(docs) != 1 || docs[0].ID != "b.tex" {
		t.Fatalf("ignored helper incorrectly: %+v, %v", docs, err)
	}
	if err := os.Remove(b); err != nil {
		t.Fatal(err)
	}
	docs, err = ScanDocuments(root, "")
	if err != nil || len(docs) != 1 || docs[0].ID != "a.tex" {
		t.Fatalf("removal did not refresh: %+v, %v", docs, err)
	}
}
