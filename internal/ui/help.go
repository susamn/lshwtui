package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type binding struct{ keys, desc string }

type bindingGroup struct {
	title    string
	bindings []binding
}

var helpGroups = []bindingGroup{
	{"Navigation", []binding{
		{"j / k", "move down / up"},
		{"g / G", "jump to first / last"},
		{"ctrl+d / ctrl+u", "half page down / up"},
		{"l / h", "expand, or collapse and step to parent"},
		{"enter / space", "toggle the selected node"},
		{"E / C", "expand all / collapse all"},
		{"tab", "switch focus between the panes"},
	}},
	{"Search", []binding{
		{"/", "search every field, live as you type"},
		{"n / N", "next / previous match"},
		{"esc", "leave raw view, or clear the search"},
	}},
	{"View", []binding{
		{"x", "toggle the raw JSON of the selected subtree"},
	}},
	{"Data", []binding{
		{"y", "yank the selected node's details"},
		{"Y", "yank the selected node's raw JSON"},
		{"s", "save the captured JSON as a snapshot"},
		{"r", "reload from lshw"},
		{"R", "reload as root for the complete tree"},
	}},
	{"General", []binding{
		{"?", "close this help"},
		{"q", "quit"},
	}},
}

func (m Model) helpView() string {
	keyWidth := 0
	for _, g := range helpGroups {
		for _, b := range g.bindings {
			if len(b.keys) > keyWidth {
				keyWidth = len(b.keys)
			}
		}
	}

	var lines []string
	for i, g := range helpGroups {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, m.theme.DetailSection.Render(g.title))
		for _, b := range g.bindings {
			lines = append(lines,
				"  "+m.theme.HelpKey.Width(keyWidth).Render(b.keys)+"  "+m.theme.Help.Render(b.desc))
		}
	}

	h := m.paneHeight()
	for len(lines) < h-1 {
		lines = append(lines, "")
	}
	inner := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.PaneTitleActive.Render("Keybindings"),
		strings.Join(lines[:maxInt(0, h-1)], "\n"),
	)
	return m.theme.BorderActive.Width(m.width - 2).Height(h).Render(inner)
}
