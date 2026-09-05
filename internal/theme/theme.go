// Package theme centralises all lipgloss styling and Nerd Font glyphs for
// lshwtui, so the UI layer never hardcodes a colour or icon directly. Every
// colour is an AdaptiveColor: a single value that resolves differently
// depending on whether the terminal reports a light or dark background,
// keeping the TUI readable in both.
package theme

import "github.com/charmbracelet/lipgloss"

// Theme holds every style used by the UI, plus whether Nerd Font glyphs are
// available. Construct it with New.
type Theme struct {
	glyphs bool

	TreeRow      lipgloss.Style
	TreeSelected lipgloss.Style
	TreeMatch    lipgloss.Style
	TreeCurMatch lipgloss.Style
	Guide        lipgloss.Style
	Expander     lipgloss.Style

	DetailKey     lipgloss.Style
	DetailValue   lipgloss.Style
	DetailSection lipgloss.Style

	PaneTitleActive   lipgloss.Style
	PaneTitleInactive lipgloss.Style
	BorderActive      lipgloss.Style
	BorderInactive    lipgloss.Style

	Status     lipgloss.Style
	StatusKey  lipgloss.Style
	StatusWarn lipgloss.Style
	StatusErr  lipgloss.Style

	Help    lipgloss.Style
	HelpKey lipgloss.Style

	Prompt lipgloss.Style
}

// Chrome colours, shared across styles that aren't tied to a device class.
var (
	colorFg     = lipgloss.AdaptiveColor{Light: "#1e293b", Dark: "#c0caf5"}
	colorMuted  = lipgloss.AdaptiveColor{Light: "#64748b", Dark: "#787c99"}
	colorDim    = lipgloss.AdaptiveColor{Light: "#94a3b8", Dark: "#565f89"}
	colorAccent = lipgloss.AdaptiveColor{Light: "#1d4ed8", Dark: "#7aa2f7"}

	colorSelBg   = lipgloss.AdaptiveColor{Light: "#dbeafe", Dark: "#283457"}
	colorMatchBg = lipgloss.AdaptiveColor{Light: "#fef3c7", Dark: "#3b3220"}
	colorCurBg   = lipgloss.AdaptiveColor{Light: "#fde68a", Dark: "#5a4a1f"}

	colorWarn = lipgloss.AdaptiveColor{Light: "#b45309", Dark: "#e0af68"}
	colorErr  = lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#f7768e"}
)

// New builds a theme. glyphs=false selects ASCII fallback icons for
// terminals without a Nerd Font.
func New(glyphs bool) *Theme {
	return &Theme{
		glyphs: glyphs,

		TreeRow:      lipgloss.NewStyle().Foreground(colorFg),
		TreeSelected: lipgloss.NewStyle().Foreground(colorFg).Background(colorSelBg).Bold(true),
		TreeMatch:    lipgloss.NewStyle().Foreground(colorFg).Background(colorMatchBg),
		TreeCurMatch: lipgloss.NewStyle().Foreground(colorFg).Background(colorCurBg).Bold(true),
		Guide:        lipgloss.NewStyle().Foreground(colorDim),
		Expander:     lipgloss.NewStyle().Foreground(colorAccent),

		DetailKey:     lipgloss.NewStyle().Foreground(colorMuted),
		DetailValue:   lipgloss.NewStyle().Foreground(colorFg),
		DetailSection: lipgloss.NewStyle().Foreground(colorAccent).Bold(true),

		PaneTitleActive:   lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
		PaneTitleInactive: lipgloss.NewStyle().Foreground(colorMuted),
		BorderActive:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorAccent),
		BorderInactive:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorDim),

		Status:     lipgloss.NewStyle().Foreground(colorFg),
		StatusKey:  lipgloss.NewStyle().Foreground(colorAccent).Bold(true),
		StatusWarn: lipgloss.NewStyle().Foreground(colorWarn).Bold(true),
		StatusErr:  lipgloss.NewStyle().Foreground(colorErr).Bold(true),

		Help:    lipgloss.NewStyle().Foreground(colorFg),
		HelpKey: lipgloss.NewStyle().Foreground(colorAccent).Bold(true),

		Prompt: lipgloss.NewStyle().Foreground(colorFg).Bold(true),
	}
}

// Glyph returns the icon for an lshw device class (e.g. "processor"),
// falling back to a generic icon for any class not in the table.
func (t *Theme) Glyph(class string) string {
	g, ok := glyphTable[class]
	if !ok {
		g = glyphTable["default"]
	}
	if t.glyphs {
		return g.icon
	}
	return g.ascii
}

// ClassStyle returns the accent style used to colourise a device class.
func (t *Theme) ClassStyle(class string) lipgloss.Style {
	c, ok := classColor[class]
	if !ok {
		c = classColor["default"]
	}
	return lipgloss.NewStyle().Foreground(c)
}

// Expanded returns the marker for an expanded (open) tree node.
func (t *Theme) Expanded() string {
	if t.glyphs {
		return "▾"
	}
	return "-"
}

// Collapsed returns the marker for a collapsed (closed) tree node.
func (t *Theme) Collapsed() string {
	if t.glyphs {
		return "▸"
	}
	return "+"
}
