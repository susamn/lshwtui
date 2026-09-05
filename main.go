// Command lshwtui is a vim-navigable terminal UI for lshw.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"lshwtui/internal/lshw"
	"lshwtui/internal/ui"
)

// version is overridden at build time with -ldflags "-X main.version=…".
var version = "dev"

func main() {
	var (
		file     = flag.String("file", "", "open a saved lshw JSON snapshot instead of the live system")
		noGlyphs = flag.Bool("no-glyphs", false, "use ASCII icons instead of Nerd Font glyphs")
		showVer  = flag.Bool("version", false, "print the version and exit")
	)
	flag.StringVar(file, "f", "", "shorthand for -file")
	flag.Usage = usage
	flag.Parse()

	if *showVer {
		fmt.Println("lshwtui", version)
		return
	}

	if *file == "" {
		if err := lshw.Available(); err != nil {
			fail(err)
		}
	}

	prog := tea.NewProgram(
		ui.New(ui.Options{Glyphs: !*noGlyphs, File: *file}),
		tea.WithAltScreen(),
	)
	if _, err := prog.Run(); err != nil {
		fail(err)
	}
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		"lshwtui — a vim-navigable terminal UI for lshw\n\nUsage:\n  lshwtui [flags]\n\nFlags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(flag.CommandLine.Output(),
		"\nlshw only reports the full hardware tree as root; when run unprivileged\n"+
			"lshwtui shows what it can and offers to re-run as root with R.\n")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "lshwtui:", err)
	os.Exit(1)
}
