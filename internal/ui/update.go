package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"lshwtui/internal/lshw"
)

// Update satisfies tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case flashMsg:
		m.status, m.statusIsErr = msg.text, msg.isErr
	case loadedMsg:
		m = m.applyLoad(msg)
	case tea.KeyMsg:
		var next tea.Model
		switch m.mode {
		case modeHelp:
			m.mode = modeNormal
		case modeSearch:
			next, cmd = m.updateSearchPrompt(msg)
			m = next.(Model)
		case modeSave:
			next, cmd = m.updateSavePrompt(msg)
			m = next.(Model)
		default:
			next, cmd = m.updateNormal(msg)
			m = next.(Model)
		}
	}
	m.syncTreeTop()
	return m, cmd
}

func (m Model) applyLoad(msg loadedMsg) Model {
	if msg.err != nil {
		m.status, m.statusIsErr = msg.err.Error(), true
		if m.tree == nil {
			m.fatal = msg.err.Error()
		}
		return m
	}
	root, err := lshw.Parse(msg.data)
	if err != nil {
		m.status, m.statusIsErr = err.Error(), true
		if m.tree == nil {
			m.fatal = err.Error()
		}
		return m
	}

	m.raw = msg.data
	m.source = msg.source
	m.tree = NewTree(root)
	m.treeTop, m.detailTop = 0, 0
	m.search.Clear()
	m.status, m.statusIsErr = fmt.Sprintf("loaded %d devices from %s", countNodes(root), msg.source.Describe()), false
	return m
}

func countNodes(root *lshw.Node) int {
	n := 0
	root.Walk(func(*lshw.Node) { n++ })
	return n
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		m.mode = modeHelp
		return m, nil
	case "tab":
		m.focus = 1 - m.focus
		return m, nil
	case "r":
		m.status, m.statusIsErr = "reloading…", false
		return m, m.loadCmd()
	case "R":
		if lshw.IsRoot() {
			m.status, m.statusIsErr = "already running as root", false
			return m, nil
		}
		return m, m.rootLoadCmd()
	}

	if m.tree == nil {
		return m, nil
	}

	switch msg.String() {
	case "/":
		m.mode = modeSearch
		m.input = m.search.Query()
		m.searchFrom = m.tree.Selected()
		return m, nil
	case "n":
		return m.jumpMatch(m.search.Next()), nil
	case "N":
		return m.jumpMatch(m.search.Prev()), nil
	case "esc":
		switch {
		case m.rawView:
			m.rawView = false
			m.detailTop = 0
		case m.search.Active():
			m.search.Clear()
			m.status, m.statusIsErr = "search cleared", false
		}
		return m, nil
	case "x":
		m.rawView = !m.rawView
		m.detailTop = 0
		return m, nil
	case "y":
		return m, m.yank(PlainText(m.tree.Selected()), "detail")
	case "Y":
		return m, m.yank(RawJSON(m.tree.Selected()), "raw JSON")
	case "s":
		m.mode = modeSave
		m.input = defaultSnapshotPath()
		return m, nil
	}

	if m.focus == paneDetail {
		return m.moveDetail(msg), nil
	}
	return m.moveTree(msg), nil
}

func (m Model) moveTree(msg tea.KeyMsg) Model {
	switch msg.String() {
	case "j", "down":
		m.tree.Move(1)
	case "k", "up":
		m.tree.Move(-1)
	case "ctrl+d":
		m.tree.Move(m.paneHeight() / 2)
	case "ctrl+u":
		m.tree.Move(-m.paneHeight() / 2)
	case "g", "home":
		m.tree.SetCursor(0)
	case "G", "end":
		m.tree.SetCursor(m.tree.Len() - 1)
	case "l", "right":
		m.tree.Expand()
	case "h", "left":
		m.tree.Collapse()
	case "enter", " ":
		m.tree.Toggle()
	case "E":
		m.tree.SetAllExpanded(true)
	case "C":
		m.tree.SetAllExpanded(false)
	default:
		return m
	}
	m.detailTop = 0
	return m
}

func (m Model) moveDetail(msg tea.KeyMsg) Model {
	step := 0
	switch msg.String() {
	case "j", "down":
		step = 1
	case "k", "up":
		step = -1
	case "ctrl+d":
		step = m.paneHeight() / 2
	case "ctrl+u":
		step = -m.paneHeight() / 2
	case "g", "home":
		m.detailTop = 0
		return m
	case "G", "end":
		m.detailTop = m.detailMaxTop()
		return m
	default:
		return m
	}
	m.detailTop = clamp(m.detailTop+step, 0, m.detailMaxTop())
	return m
}

func (m Model) updateSearchPrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		m.search.Clear()
		if m.searchFrom != nil {
			m.tree.Reveal(m.searchFrom)
		}
		return m, nil
	case "enter":
		m.mode = modeNormal
		if m.search.Count() == 0 && m.search.Active() {
			m.status, m.statusIsErr = fmt.Sprintf("no match for %q", m.search.Query()), true
		}
		return m, nil
	case "backspace":
		if m.input != "" {
			m.input = m.input[:len(m.input)-1]
		}
	case "ctrl+u":
		m.input = ""
	default:
		if r := msg.Runes; len(r) > 0 {
			m.input += string(r)
		}
	}

	// Re-run on every keystroke so the tree follows the query as it is typed.
	m.search.Run(m.tree.Root, m.input, m.searchFrom)
	if n := m.search.Current(); n != nil {
		m.tree.Reveal(n)
		m.detailTop = 0
	}
	return m, nil
}

func (m Model) updateSavePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeNormal
		return m, nil
	case "enter":
		m.mode = modeNormal
		path := strings.TrimSpace(m.input)
		if path == "" {
			return m, flash("save cancelled: empty path", true)
		}
		if err := lshw.SaveFile(expandHome(path), m.raw); err != nil {
			return m, flash(err.Error(), true)
		}
		return m, flash("saved snapshot to "+path, false)
	case "backspace":
		if m.input != "" {
			m.input = m.input[:len(m.input)-1]
		}
	case "ctrl+u":
		m.input = ""
	default:
		if r := msg.Runes; len(r) > 0 {
			m.input += string(r)
		}
	}
	return m, nil
}

func (m Model) jumpMatch(n *lshw.Node) Model {
	if n == nil {
		m.status, m.statusIsErr = "no matches", true
		return m
	}
	m.tree.Reveal(n)
	m.detailTop = 0
	return m
}

func (m Model) yank(text, what string) tea.Cmd {
	if err := copyToClipboard(text); err != nil {
		return flash("yank failed: "+err.Error(), true)
	}
	return flash(fmt.Sprintf("yanked %s (%d bytes)", what, len(text)), false)
}

func defaultSnapshotPath() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "hardware"
	}
	return fmt.Sprintf("lshw-%s-%s.json", host, time.Now().Format("20060102"))
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}

func clamp(v, lo, hi int) int { return maxInt(lo, minInt(v, hi)) }

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
