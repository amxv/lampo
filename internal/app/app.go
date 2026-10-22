package app

import (
	"context"
	"fmt"
	"io"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/amxv/lampo/internal/buildinfo"
	"github.com/amxv/lampo/internal/preview"
)

const commandName = "lampo"

var version = buildinfo.CurrentVersion()

// Run starts a local preview and blocks until the user interrupts the CLI.
func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 1 {
		switch args[0] {
		case "--help", "-h", "help":
			printHelp(stdout)
			return nil
		case "--version", "-v", "version":
			_, _ = fmt.Fprintf(stdout, "%s %s\n", commandName, version)
			return nil
		}
	}
	options, err := parseOptions(args)
	if err != nil {
		return fmt.Errorf("%w (run `lampo --help`)", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return preview.Serve(ctx, options, stdout, stderr)
}

func parseOptions(args []string) (preview.Options, error) {
	options := preview.Options{Path: ".", Host: "127.0.0.1", Open: true, Engine: "pdflatex"}
	pathSet := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		name, value, found := strings.Cut(arg, "=")
		if !found {
			name = arg
		}
		switch name {
		case "--port", "-p", "--host", "--engine":
			if !found {
				i++
				if i >= len(args) {
					return options, fmt.Errorf("%s requires a value", name)
				}
				value = args[i]
			}
			switch name {
			case "--port", "-p":
				port, err := strconv.Atoi(value)
				if err != nil || port < 0 || port > 65535 {
					return options, fmt.Errorf("invalid port %q (expected 0-65535)", value)
				}
				options.Port = port
			case "--host":
				options.Host = value
			case "--engine":
				options.Engine = value
			}
		case "--no-open":
			options.Open = false
		case "--open":
			options.Open = true
		default:
			if strings.HasPrefix(arg, "-") {
				return options, fmt.Errorf("unknown option %q", arg)
			}
			if pathSet {
				return options, fmt.Errorf("only one LaTeX project or file may be specified")
			}
			options.Path = arg
			pathSet = true
		}
	}
	if options.Host == "" {
		return options, fmt.Errorf("host cannot be empty")
	}
	switch options.Engine {
	case "pdflatex", "xelatex", "lualatex":
	default:
		return options, fmt.Errorf("unknown engine %q (choose pdflatex, xelatex or lualatex)", options.Engine)
	}
	return options, nil
}

func printHelp(w io.Writer) {
	_, _ = fmt.Fprint(w, `Lampo - live LaTeX preview in your browser

Usage:
  lampo [file.tex | directory] [options]

Run without a file to open main.tex (or the only .tex file) in the current
directory. Lampo watches your project, recompiles on save, and instantly
updates the browser preview. Press Ctrl+C to stop.

Options:
  -p, --port <number>     HTTP port (default: any free port)
      --host <address>    Bind address (default: 127.0.0.1)
      --engine <name>     pdflatex (default), xelatex, or lualatex
      --no-open           Print the URL without opening the browser
      --open              Open the browser (default)
  -v, --version           Print the version
  -h, --help              Show this help

Examples:
  lampo
  lampo resume/main.tex
  lampo ./paper --port 3030
  lampo article.tex --engine xelatex --no-open

Requires latexmk and your chosen LaTeX engine on PATH (TeX Live or MiKTeX).
No Node.js, cloud services, or external network connections are required.
`)
}
