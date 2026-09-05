// Package ui implements the Bubble Tea application: a two-pane, vim-navigable
// view of the lshw hardware tree.
package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lshwtui/internal/lshw"
	"lshwtui/internal/theme"
)

type pane int

const (
	paneTree pane = iota
	paneDetail
)

type mode int

const (
	modeNormal mode = iota
	modeSearch
	modeSave
	modeHelp
)

// Options configures a Model at construction time.
type Options struct {
	Glyphs bool   // draw Nerd Font icons rather than ASCII fallbacks
	File   string // open this snapshot instead of the live system
}

// Model is the root Bubble Tea model.
type Model struct {
	theme *theme.Theme
	opts  Options

	tree   *Tree
	search Search
	source lshw.Source
	raw    []byte // the JSON behind the current tree, kept for snapshot saving

	focus   pane
	mode    mode
	rawView bool

	input      string     // text buffer for the search and save prompts
	searchFrom *lshw.Node // cursor position when search began, restored on esc

	status      string
	statusIsErr bool
	fatal       string

	width, height int
	treeTop       int
	detailTop     int
}

// New builds a Model. Loading happens in the first Init command so the UI can
// draw a frame before lshw — which takes a second or two — has finished.
func New(opts Options) Model {
	return Model{
		theme:  theme.New(opts.Glyphs),
		opts:   opts,
		status: "loading hardware…",
	}
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd { return m.loadCmd() }

// --- messages ---

type loadedMsg struct {
	data   []byte
	source lshw.Source
	err    error
}

type flashMsg struct {
	text  string
	isErr bool
}

// --- commands ---

func (m Model) loadCmd() tea.Cmd {
	file := m.opts.File
	return func() tea.Msg {
		if file != "" {
			data, src, err := lshw.LoadFile(file)
			return loadedMsg{data: data, source: src, err: err}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		data, src, err := lshw.Capture(ctx)
		return loadedMsg{data: data, source: src, err: err}
	}
}

// rootLoadCmd hands the terminal back to the escalation helper so its password
// prompt is visible, then reads the JSON it wrote to a temporary file.
func (m Model) rootLoadCmd() tea.Cmd {
	tmp := filepath.Join(os.TempDir(), fmt.Sprintf("lshwtui-%d.json", os.Getpid()))
	cmd, src, err := lshw.PrivilegedCapture(tmp)
	if err != nil {
		return flash(err.Error(), true)
	}
	return tea.ExecProcess(cmd, func(runErr error) tea.Msg {
		defer os.Remove(tmp)
		if runErr != nil {
			return loadedMsg{err: fmt.Errorf("privileged capture failed: %w", runErr)}
		}
		data, readErr := os.ReadFile(tmp)
		if readErr != nil {
			return loadedMsg{err: fmt.Errorf("reading privileged capture: %w", readErr)}
		}
		return loadedMsg{data: data, source: src}
	})
}

func flash(text string, isErr bool) tea.Cmd {
	return func() tea.Msg { return flashMsg{text: text, isErr: isErr} }
}
