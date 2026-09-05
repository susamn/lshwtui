package ui

import (
	"encoding/json"
	"fmt"
	"sort"

	"lshwtui/internal/lshw"
)

// FieldKind distinguishes the line types the detail pane can draw.
type FieldKind int

const (
	FieldPair    FieldKind = iota // a key/value row
	FieldSection                  // a heading such as "configuration"
	FieldBlank                    // vertical breathing room
)

// Field is one line of the detail pane, produced without any styling so the
// content and its presentation stay separable.
type Field struct {
	Kind  FieldKind
	Key   string
	Value string
}

// Fields renders a node as detail-pane lines: its scalar fields first, then
// the configuration and capabilities maps as sorted sections.
func Fields(n *lshw.Node) []Field {
	if n == nil {
		return nil
	}
	var out []Field
	for _, k := range n.Keys() {
		out = append(out, Field{Kind: FieldPair, Key: k, Value: formatValue(n, k)})
	}
	for _, section := range []string{"configuration", "capabilities"} {
		m := n.Map(section)
		if len(m) == 0 {
			continue
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		out = append(out, Field{Kind: FieldBlank}, Field{Kind: FieldSection, Key: section})
		for _, k := range keys {
			out = append(out, Field{Kind: FieldPair, Key: k, Value: m[k]})
		}
	}
	return out
}

// RawJSON renders the node's own subtree as indented JSON.
func RawJSON(n *lshw.Node) string {
	if n == nil {
		return ""
	}
	b, err := json.MarshalIndent(n.Raw, "", "  ")
	if err != nil {
		return fmt.Sprintf("cannot render JSON: %v", err)
	}
	return string(b)
}

// PlainText renders a node's fields as tab-free text, for yanking.
func PlainText(n *lshw.Node) string {
	fields := Fields(n)
	width := 0
	for _, f := range fields {
		if f.Kind == FieldPair && len(f.Key) > width {
			width = len(f.Key)
		}
	}
	out := ""
	for _, f := range fields {
		switch f.Kind {
		case FieldBlank:
			out += "\n"
		case FieldSection:
			out += "[" + f.Key + "]\n"
		default:
			out += fmt.Sprintf("%-*s  %s\n", width, f.Key, f.Value)
		}
	}
	return out
}

// formatValue annotates the raw value where lshw's units are unhelpful on
// their own — byte counts and Hz are far easier to read scaled.
func formatValue(n *lshw.Node, key string) string {
	raw := n.Str(key)
	v, ok := n.Num(key)
	if !ok || v <= 0 {
		return raw
	}
	switch key {
	case "size", "capacity":
		if u := n.Str("units"); u != "" && u != "bytes" {
			return raw
		}
		return fmt.Sprintf("%s  (%s)", humanBytes(v), raw)
	case "clock":
		return fmt.Sprintf("%s  (%s Hz)", humanHertz(v), raw)
	}
	return raw
}

func humanBytes(v int64) string {
	const unit = 1024
	if v < unit {
		return fmt.Sprintf("%d B", v)
	}
	div, exp := int64(unit), 0
	for n := v / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(v)/float64(div), "KMGTPE"[exp])
}

func humanHertz(v int64) string {
	switch {
	case v >= 1_000_000_000:
		return fmt.Sprintf("%.2f GHz", float64(v)/1e9)
	case v >= 1_000_000:
		return fmt.Sprintf("%.0f MHz", float64(v)/1e6)
	case v >= 1_000:
		return fmt.Sprintf("%.0f kHz", float64(v)/1e3)
	}
	return fmt.Sprintf("%d Hz", v)
}
