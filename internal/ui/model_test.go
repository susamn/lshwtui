package ui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"lshwtui/internal/lshw"
)

// drive feeds a sequence of key presses through the model, returning the
// resulting model and its rendered frame.
func drive(t *testing.T, m Model, keys ...string) (Model, string) {
	t.Helper()
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "enter", "esc", "tab", "backspace", "up", "down", "ctrl+d", "ctrl+u":
			msg = tea.KeyMsg{Type: keyTypes[k]}
		case " ":
			msg = tea.KeyMsg{Type: tea.KeySpace}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	return m, m.View()
}

var keyTypes = map[string]tea.KeyType{
	"enter": tea.KeyEnter, "esc": tea.KeyEsc, "tab": tea.KeyTab,
	"backspace": tea.KeyBackspace, "up": tea.KeyUp, "down": tea.KeyDown,
	"ctrl+d": tea.KeyCtrlD, "ctrl+u": tea.KeyCtrlU, "ctrl+w": tea.KeyCtrlW,
}

func loaded(t *testing.T) Model {
	t.Helper()
	data, err := os.ReadFile("../../testdata/sample.json")
	if err != nil {
		t.Skipf("no sample capture: %v", err)
	}
	m := New(Options{Glyphs: true})
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = sized.(Model)
	next, _ := m.Update(loadedMsg{data: data, source: lshw.Source{Origin: "lshw"}})
	m = next.(Model)
	if m.tree == nil {
		t.Fatalf("tree not built: %s", m.status)
	}
	return m
}

func TestViewRendersBothPanes(t *testing.T) {
	m := loaded(t)
	out := m.View()
	for _, want := range []string{"Devices", "Details", "help"} {
		if !strings.Contains(out, want) {
			t.Errorf("view missing %q", want)
		}
	}
	if got := strings.Count(out, "\n") + 1; got != 40 {
		t.Errorf("view is %d lines, want exactly 40 to fill the terminal", got)
	}
	// Every line must fill the width exactly, or the panes drift apart as the
	// user scrolls through rows of differing length.
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w != 120 {
			t.Errorf("line %d is %d cells wide, want 120", i, w)
		}
	}
}

func TestNavigationStaysInBounds(t *testing.T) {
	m := loaded(t)
	m, _ = drive(t, m, "G")
	if m.tree.Cursor() != m.tree.Len()-1 {
		t.Errorf("G left cursor at %d, want %d", m.tree.Cursor(), m.tree.Len()-1)
	}
	m, _ = drive(t, m, "j", "j", "j")
	if m.tree.Cursor() != m.tree.Len()-1 {
		t.Errorf("j past the end moved cursor to %d", m.tree.Cursor())
	}
	m, _ = drive(t, m, "g", "k", "k")
	if m.tree.Cursor() != 0 {
		t.Errorf("k past the top moved cursor to %d", m.tree.Cursor())
	}
}

func TestCollapseAllThenExpandAll(t *testing.T) {
	m := loaded(t)
	full := m.tree.Len()
	// Collapse-all deliberately leaves the root open, so the user is left with
	// a browsable top level rather than a single line.
	m, _ = drive(t, m, "C")
	want := 1 + len(m.tree.Root.Children)
	if m.tree.Len() != want {
		t.Errorf("collapse all left %d rows, want %d (root plus its children)", m.tree.Len(), want)
	}
	m, _ = drive(t, m, "E")
	if m.tree.Len() != full {
		t.Errorf("expand all restored %d rows, want %d", m.tree.Len(), full)
	}
}

func TestCollapseAllWhenCursorOnDeepChild(t *testing.T) {
	m := loaded(t)
	// Move deep into the tree.
	m, _ = drive(t, m, "j", "j", "j", "j", "j")
	if m.tree.Cursor() == 0 {
		t.Fatal("cursor should not be at root")
	}
	m, _ = drive(t, m, "C")
	want := 1 + len(m.tree.Root.Children)
	if m.tree.Len() != want {
		t.Errorf("collapse all from deep child left %d rows, want %d", m.tree.Len(), want)
	}
}

func TestTreeViewportStability(t *testing.T) {
	m := loaded(t)
	// Set small window so scrolling occurs
	sized, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m = sized.(Model)
	h := m.paneHeight() - 1 // title line

	// Move cursor past h
	for i := 0; i <= h+2; i++ {
		m, _ = drive(t, m, "j")
	}
	topAfterDown := m.treeTop
	if topAfterDown <= 0 {
		t.Fatalf("expected treeTop > 0 after scrolling past height, got %d", topAfterDown)
	}

	// Move cursor up by one: treeTop should NOT change because cursor is still within viewport
	m, _ = drive(t, m, "k")
	if m.treeTop != topAfterDown {
		t.Errorf("treeTop shifted from %d to %d on single k, expected viewport to stay stable", topAfterDown, m.treeTop)
	}
}

func TestSearchFindsAndClears(t *testing.T) {
	m := loaded(t)
	m, _ = drive(t, m, "/", "c", "p", "u")
	if m.search.Count() == 0 {
		t.Fatal(`searching "cpu" found nothing in a real capture`)
	}
	if m.tree.Selected() != m.search.Current() {
		t.Error("cursor did not follow the current match")
	}
	m, _ = drive(t, m, "enter", "esc")
	if m.search.Active() {
		t.Error("esc did not clear the search")
	}
}

func TestSearchPromptEditing(t *testing.T) {
	m := loaded(t)
	// Multi-byte UTF-8 rune handling on backspace
	m, _ = drive(t, m, "/", "c", "p", "u", " ", "世", "界")
	if m.input != "cpu 世界" {
		t.Fatalf("expected input 'cpu 世界', got %q", m.input)
	}
	m, _ = drive(t, m, "backspace")
	if m.input != "cpu 世" {
		t.Fatalf("expected input 'cpu 世' after backspacing rune, got %q", m.input)
	}
	// ctrl+w deletes previous word
	m, _ = drive(t, m, "ctrl+w")
	if m.input != "cpu " {
		t.Fatalf("expected input 'cpu ' after ctrl+w, got %q", m.input)
	}
	// ctrl+u clears entire line
	m, _ = drive(t, m, "ctrl+u")
	if m.input != "" {
		t.Fatalf("expected empty input after ctrl+u, got %q", m.input)
	}
}

func TestSearchMissReportsNoMatch(t *testing.T) {
	m := loaded(t)
	m, out := drive(t, m, "/", "z", "z", "z", "q")
	if m.search.Count() != 0 {
		t.Fatal("expected no matches for zzzq")
	}
	if !strings.Contains(out, "no matches") {
		t.Error("prompt did not report the empty result")
	}
}

func TestRawViewToggles(t *testing.T) {
	m := loaded(t)
	m, out := drive(t, m, "x")
	if !strings.Contains(out, "Raw JSON") {
		t.Error("x did not switch the detail pane to raw JSON")
	}
	if !strings.Contains(out, `"class"`) {
		t.Error("raw view is not showing JSON")
	}
	m, out = drive(t, m, "esc")
	if strings.Contains(out, "Raw JSON") {
		t.Error("esc did not leave raw view")
	}
}

func TestHelpOverlay(t *testing.T) {
	m := loaded(t)
	_, out := drive(t, m, "?")
	for _, want := range []string{"Keybindings", "Navigation", "yank"} {
		if !strings.Contains(out, want) {
			t.Errorf("help overlay missing %q", want)
		}
	}
}

func TestSnapshotReloadAsRootIgnored(t *testing.T) {
	m := loaded(t)
	m.source = lshw.Source{Origin: "snap.json", Snapshot: true}
	m, _ = drive(t, m, "R")
	if !strings.Contains(m.status, "cannot reload snapshot as root") {
		t.Errorf("expected status to warn about snapshot escalation, got %q", m.status)
	}
}

func TestNarrowTerminalDoesNotPanic(t *testing.T) {
	m := loaded(t)
	for _, size := range []tea.WindowSizeMsg{{Width: 10, Height: 5}, {Width: 12, Height: 5}, {Width: 16, Height: 5}, {Width: 20, Height: 5}, {Width: 35, Height: 10}, {Width: 200, Height: 60}} {
		next, _ := m.Update(size)
		out := next.(Model).View()
		if out == "" {
			t.Errorf("empty frame at %dx%d", size.Width, size.Height)
		}
		// Verify lines do not exceed terminal width in either pane.
		detail := next.(Model)
		detail.focus = paneDetail
		for _, frame := range []string{out, detail.View()} {
			for i, line := range strings.Split(frame, "\n") {
				if w := lipgloss.Width(line); w > size.Width {
					t.Errorf("at %dx%d: line %d is %d cells wide, exceeds %d", size.Width, size.Height, i, w, size.Width)
				}
			}
		}
	}
}

func TestSavePromptDefaultsToDatedFile(t *testing.T) {
	m := loaded(t)
	m, out := drive(t, m, "s")
	if m.mode != modeSave {
		t.Fatal("s did not open the save prompt")
	}
	if !strings.HasSuffix(m.input, ".json") || !strings.Contains(out, "save snapshot to") {
		t.Errorf("unexpected save prompt %q", m.input)
	}
}
