package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpAndVersion(t *testing.T) {
	for _, input := range []string{"--help", "help", "-h"} {
		var out, errors bytes.Buffer
		if err := Run([]string{input}, &out, &errors); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "--port") || !strings.Contains(out.String(), "Usage:") {
			t.Fatalf("incomplete help: %s", out.String())
		}
	}
	var out, errors bytes.Buffer
	if err := Run([]string{"--version"}, &out, &errors); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "lampo ") {
		t.Fatalf("unexpected version: %q", out.String())
	}
}

func TestParseOptions(t *testing.T) {
	o, err := parseOptions([]string{"paper/main.tex", "--port=3001", "--engine", "xelatex", "--no-open"})
	if err != nil {
		t.Fatal(err)
	}
	if o.Path != "paper/main.tex" || o.Port != 3001 || o.Engine != "xelatex" || o.Open {
		t.Fatalf("unexpected options: %+v", o)
	}
	defaultOptions, err := parseOptions(nil)
	if err != nil || defaultOptions.Path != "." || defaultOptions.Host != "127.0.0.1" || !defaultOptions.Open {
		t.Fatalf("unexpected defaults: %+v, %v", defaultOptions, err)
	}
}

func TestParseErrors(t *testing.T) {
	for _, args := range [][]string{
		{"--bad"}, {"--port", "foo"}, {"--port", "65536"}, {"--port"},
		{"--engine", "bad"}, {"first.tex", "second.tex"}, {"--host="},
	} {
		if _, err := parseOptions(args); err == nil {
			t.Errorf("expected error for %q", args)
		}
	}
}
