package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	statusLines  = 1
	borderLines  = 2
	minTreeWidth = 26
	maxTreeWidth = 64
)

// paneHeight is the number of content lines inside a pane's border.
func (m Model) paneHeight() int {
	return maxInt(1, m.height-statusLines-borderLines)
}

func (m Model) treeWidth() int {
	return clamp(m.width*2/5, minTreeWidth, maxTreeWidth)
}

func (m Model) detailLineCount() int {
	if m.tree == nil {
		return 0
	}
	if m.rawView {
		return strings.Count(RawJSON(m.tree.Selected()), "\n") + 1
	}
	return len(Fields(m.tree.Selected()))
}

// View satisfies tea.Model.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "" // no size yet; the first WindowSizeMsg is imminent
	}
	if m.fatal != "" {
		return m.theme.StatusErr.Render("lshwtui: "+m.fatal) + "\n"
	}
	if m.tree == nil {
		return m.theme.Status.Render("  " + m.status)
	}
	if m.mode == modeHelp {
		return lipgloss.JoinVertical(lipgloss.Left, m.helpView(), m.statusBar())
	}

	// One column of gutter between the panes so their borders do not fuse.
	tw := m.treeWidth()
	dw := maxInt(20, m.width-tw-1)
	gutter := strings.Repeat(" \n", m.paneHeight()+borderLines-1) + " "
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		m.pane("Devices", m.treeLines(tw-4), tw, m.focus == paneTree),
		gutter,
		m.pane(m.detailTitle(), m.detailLines(dw-4), dw, m.focus == paneDetail),
	)
	return lipgloss.JoinVertical(lipgloss.Left, body, m.statusBar())
}

func (m Model) detailTitle() string {
	if m.rawView {
		return "Raw JSON"
	}
	return "Details"
}

// pane frames content in a bordered box, padding it to a fixed height so the
// two panes always line up regardless of how much content each holds.
func (m Model) pane(title string, lines []string, width int, active bool) string {
	h := m.paneHeight()
	for len(lines) < h {
		lines = append(lines, "")
	}
	lines = lines[:h]

	titleStyle, border := m.theme.PaneTitleInactive, m.theme.BorderInactive
	if active {
		titleStyle, border = m.theme.PaneTitleActive, m.theme.BorderActive
	}
	inner := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		strings.Join(lines[:h-1], "\n"),
	)
	return border.Width(width - 2).Height(h).Render(inner)
}

func (m Model) treeLines(width int) []string {
	rows := m.tree.Rows()
	h := maxInt(1, m.paneHeight()-1) // the pane's title consumes one line
	top := scrollTo(m.tree.Cursor(), m.treeTop, h)

	out := make([]string, 0, h)
	for i := top; i < len(rows) && i < top+h; i++ {
		out = append(out, m.treeRow(rows[i], i == m.tree.Cursor(), width))
	}
	return out
}

func (m Model) treeRow(r Row, selected bool, width int) string {
	marker := "  "
	if r.HasKids {
		if r.Expanded {
			marker = m.theme.Expanded() + " "
		} else {
			marker = m.theme.Collapsed() + " "
		}
	}
	class := r.Node.Class()
	plain := r.Guide + r.Branch + marker + m.theme.Glyph(class) + " " + r.Node.Label()
	plain = truncate(plain, width)

	switch {
	case selected:
		return m.theme.TreeSelected.Width(width).Render(plain)
	case m.search.Matched(r.Node):
		if r.Node == m.search.Current() {
			return m.theme.TreeCurMatch.Width(width).Render(plain)
		}
		return m.theme.TreeMatch.Render(plain)
	}

	// Unselected rows keep their per-segment colour: dim guides, a
	// class-coloured icon, plain label.
	head := r.Guide + r.Branch
	body := truncate(marker+m.theme.Glyph(class)+" "+r.Node.Label(), maxInt(0, width-lipgloss.Width(head)))
	return m.theme.Guide.Render(head) + m.theme.ClassStyle(class).Render(body)
}

func (m Model) detailLines(width int) []string {
	h := m.paneHeight() - 1
	var lines []string

	if m.rawView {
		for _, l := range strings.Split(RawJSON(m.tree.Selected()), "\n") {
			lines = append(lines, m.theme.DetailValue.Render(truncate(l, width)))
		}
	} else {
		fields := Fields(m.tree.Selected())
		keyWidth := 0
		for _, f := range fields {
			if f.Kind == FieldPair && len(f.Key) > keyWidth {
				keyWidth = len(f.Key)
			}
		}
		keyWidth = minInt(keyWidth, maxInt(8, width/3))
		for _, f := range fields {
			switch f.Kind {
			case FieldBlank:
				lines = append(lines, "")
			case FieldSection:
				lines = append(lines, m.theme.DetailSection.Render(f.Key))
			default:
				key := m.theme.DetailKey.Width(keyWidth).Render(truncate(f.Key, keyWidth))
				val := m.theme.DetailValue.Render(truncate(f.Value, maxInt(0, width-keyWidth-1)))
				lines = append(lines, key+" "+val)
			}
		}
	}

	top := clamp(m.detailTop, 0, m.detailMaxTop())
	if top >= len(lines) {
		return nil
	}
	return lines[top:minInt(len(lines), top+h)]
}

func (m Model) statusBar() string {
	if m.mode == modeSearch {
		return m.theme.Prompt.Width(m.width).Render("/" + m.input + "█" + m.searchCounter())
	}
	if m.mode == modeSave {
		return m.theme.Prompt.Width(m.width).Render("save snapshot to: " + m.input + "█")
	}

	left := m.status
	style := m.theme.Status
	if m.statusIsErr {
		style = m.theme.StatusErr
	}

	right := m.source.Describe()
	if !m.source.Privileged && !m.source.Snapshot {
		right = m.theme.StatusWarn.Render("partial — press R for root") + m.theme.Status.Render("  ")
	} else {
		right = m.theme.Status.Render(right)
	}
	right += m.theme.StatusKey.Render(" ? ") + m.theme.Status.Render("help")

	gap := maxInt(1, m.width-lipgloss.Width(left)-lipgloss.Width(right)-2)
	return m.theme.Status.Width(m.width).Render(" " + style.Render(left) + strings.Repeat(" ", gap) + right)
}

func (m Model) searchCounter() string {
	if !m.search.Active() {
		return ""
	}
	if m.search.Count() == 0 {
		return "   no matches"
	}
	return m.theme.Status.Render(
		"   " + itoa(m.search.Ordinal()) + "/" + itoa(m.search.Count()) + " matches")
}

// scrollTo returns a viewport offset that keeps cursor visible within height
// lines, scrolling by the minimum amount needed.
func scrollTo(cursor, top, height int) int {
	if height <= 0 {
		return 0
	}
	if cursor < top {
		return cursor
	}
	if cursor >= top+height {
		return cursor - height + 1
	}
	return top
}

// truncate cuts s to width display cells, marking the cut with an ellipsis.
func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	for i := len(runes); i > 0; i-- {
		if lipgloss.Width(string(runes[:i]))+1 <= width {
			return string(runes[:i]) + "…"
		}
	}
	return ""
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
