// Command lshwtui is a vim-navigable terminal UI for lshw.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

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
		check    = flag.Bool("check", false, "verify hardware capture and JSON parsing, then exit")
	)
	flag.StringVar(file, "f", "", "shorthand for -file")
	flag.Usage = usage
	flag.Parse()

	if *showVer {
		fmt.Println("lshwtui", version)
		return
	}

	if *check {
		var data []byte
		var src lshw.Source
		var err error
		if *file != "" {
			data, src, err = lshw.LoadFile(*file)
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			data, src, err = lshw.Capture(ctx)
		}
		if err != nil {
			fail(err)
		}
		root, err := lshw.Parse(data)
		if err != nil {
			fail(err)
		}
		count := 0
		root.Walk(func(*lshw.Node) { count++ })
		fmt.Printf("OK: %s (%d devices detected)\n", src.Describe(), count)
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
