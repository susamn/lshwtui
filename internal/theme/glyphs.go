package theme

import "github.com/charmbracelet/lipgloss"

// classGlyph pairs a Nerd Font v3 icon with a plain-ASCII fallback for the
// same lshw device class. Both strings are kept to a single rune so that
// "glyph + separator space" is a stable two-cell prefix in the tree,
// regardless of which mode is active.
type classGlyph struct {
	icon  string
	ascii string
}

// glyphTable maps lshw "class" values (as reported by `lshw -class`) to an
// icon. Codepoints are drawn from the fa (Font Awesome) and oct (Octicons)
// sets patched into Nerd Fonts v3, which are the most consistently present
// across popular Nerd Font builds.
var glyphTable = map[string]classGlyph{
	// system: root node of the tree; fa-desktop reads as "the machine itself".
	"system": {icon: "", ascii: "@"},

	// bus: fa-exchange — bidirectional data path.
	"bus": {icon: "", ascii: "="},

	// memory: fa-microchip.
	"memory": {icon: "", ascii: "M"},

	// processor: oct-cpu. Octicons' dedicated CPU glyph reads better at a
	// glance than any Font Awesome alternative available in the v3 set.
	"processor": {icon: "", ascii: "P"},

	// bridge: oct-git-merge — two paths joining into one, the closest
	// stand-in Nerd Fonts offers for a hardware bridge/hub.
	"bridge": {icon: "", ascii: "B"},

	// display: fa-television.
	"display": {icon: "", ascii: "D"},

	// multimedia: fa-music.
	"multimedia": {icon: "", ascii: "A"},

	// communication: fa-phone.
	"communication": {icon: "", ascii: "C"},

	// network: fa-wifi.
	"network": {icon: "", ascii: "N"},

	// storage: fa-hdd-o — the storage *controller*, distinct from the disk
	// device itself below.
	"storage": {icon: "", ascii: "S"},

	// disk: fa-database — the physical/logical disk device.
	"disk": {icon: "", ascii: "H"},

	// volume: fa-folder-open — a mounted partition/volume.
	"volume": {icon: "", ascii: "V"},

	// input: fa-keyboard-o.
	"input": {icon: "", ascii: "K"},

	// generic: fa-cog — catch-all hardware lshw can't classify further.
	"generic": {icon: "", ascii: "G"},

	// power: fa-bolt.
	"power": {icon: "", ascii: "E"},

	// battery: fa-battery-full.
	"battery": {icon: "", ascii: "b"},

	// printer: fa-print.
	"printer": {icon: "", ascii: "R"},

	// tape: fa-archive — a backup/archival device.
	"tape": {icon: "", ascii: "T"},

	// default: fa-question-circle, used for any class not listed above.
	"default": {icon: "", ascii: "?"},
}

// classColor gives each device class a distinct AdaptiveColor accent. The
// classes lshw surfaces most often — processor, memory, disk, network and
// display — are spread across opposite corners of the hue wheel (blue,
// cyan, orange, teal-green, purple) and deliberately avoid red, which is
// reserved for StatusErr elsewhere in the theme. Remaining, less critical
// classes reuse nearby hues at different lightness so the overall palette
// stays restrained rather than becoming a rainbow.
var classColor = map[string]lipgloss.AdaptiveColor{
	"system":        {Light: "#1e293b", Dark: "#c0caf5"},
	"bus":           {Light: "#475569", Dark: "#94a3b8"},
	"bridge":        {Light: "#7c3aed", Dark: "#c4b5fd"},
	"memory":        {Light: "#0e7490", Dark: "#7dd3fc"},
	"processor":     {Light: "#0058b0", Dark: "#7aa2f7"},
	"display":       {Light: "#a21caf", Dark: "#e0aaff"},
	"multimedia":    {Light: "#be185d", Dark: "#f9a8d4"},
	"communication": {Light: "#0891b2", Dark: "#67e8f9"},
	"network":       {Light: "#0f766e", Dark: "#5eead4"},
	"storage":       {Light: "#4338ca", Dark: "#9aa5f7"},
	"disk":          {Light: "#b45309", Dark: "#e0af68"},
	"volume":        {Light: "#a16207", Dark: "#fcd34d"},
	"input":         {Light: "#4d7c0f", Dark: "#bef264"},
	"generic":       {Light: "#57534e", Dark: "#a8a29e"},
	"power":         {Light: "#854d0e", Dark: "#facc15"},
	"battery":       {Light: "#15803d", Dark: "#86efac"},
	"printer":       {Light: "#334155", Dark: "#cbd5e1"},
	"tape":          {Light: "#78350f", Dark: "#d6a77a"},
	"default":       {Light: "#6b7280", Dark: "#9ca3af"},
}
