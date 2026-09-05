// Package lshw models the hardware tree produced by `lshw -json`.
package lshw

import (
	"fmt"
	"sort"
	"strings"
)

// Node is one entry in the lshw hardware tree.
//
// lshw's JSON schema is loose: which keys appear depends entirely on the
// device class, and new keys show up between lshw releases. So the raw decoded
// object is kept as-is and typed access happens through accessors, rather than
// pinning a struct that silently drops fields we did not anticipate.
type Node struct {
	Raw      map[string]any
	Children []*Node
	Parent   *Node
}

// Str returns a scalar field rendered as a string, or "" when absent.
func (n *Node) Str(key string) string {
	v, ok := n.Raw[key]
	if !ok {
		return ""
	}
	return scalarString(v)
}

// Num returns a numeric field, or 0 and false when absent or non-numeric.
func (n *Node) Num(key string) (int64, bool) {
	f, ok := n.Raw[key].(float64)
	if !ok {
		return 0, false
	}
	return int64(f), true
}

// Map returns a nested string map field such as configuration or capabilities.
func (n *Node) Map(key string) map[string]string {
	m, ok := n.Raw[key].(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = scalarString(v)
	}
	return out
}

// List returns a string-slice field such as logicalname.
func (n *Node) List(key string) []string {
	switch v := n.Raw[key].(type) {
	case nil:
		return nil
	case []any:
		out := make([]string, 0, len(v))
		for _, e := range v {
			out = append(out, scalarString(e))
		}
		return out
	default:
		return []string{scalarString(v)}
	}
}

func (n *Node) ID() string          { return n.Str("id") }
func (n *Node) Class() string       { return n.Str("class") }
func (n *Node) Description() string { return n.Str("description") }
func (n *Node) Product() string     { return n.Str("product") }
func (n *Node) Vendor() string      { return n.Str("vendor") }

// Label is the one-line name shown for this node in the tree pane.
func (n *Node) Label() string {
	for _, s := range []string{n.Product(), n.Description(), n.ID()} {
		if s != "" {
			return s
		}
	}
	return "(unnamed)"
}

// Walk visits n and every descendant in depth-first order.
func (n *Node) Walk(fn func(*Node)) {
	fn(n)
	for _, c := range n.Children {
		c.Walk(fn)
	}
}

// Depth is the number of ancestors above this node.
func (n *Node) Depth() int {
	d := 0
	for p := n.Parent; p != nil; p = p.Parent {
		d++
	}
	return d
}

// Keys returns the node's scalar field names, minus structural ones, in a
// stable display order: well-known fields first, then the rest alphabetically.
func (n *Node) Keys() []string {
	var known, extra []string
	seen := map[string]bool{}
	for _, k := range preferredOrder {
		if _, ok := n.Raw[k]; ok && isScalar(n.Raw[k]) {
			known = append(known, k)
			seen[k] = true
		}
	}
	for k, v := range n.Raw {
		if seen[k] || structuralKeys[k] || !isScalar(v) {
			continue
		}
		extra = append(extra, k)
	}
	sort.Strings(extra)
	return append(known, extra...)
}

var preferredOrder = []string{
	"id", "class", "description", "product", "vendor", "serial", "version",
	"businfo", "physid", "handle", "logicalname", "dev", "slot",
	"size", "capacity", "clock", "width", "units", "claimed", "disabled",
}

var structuralKeys = map[string]bool{
	"children": true, "configuration": true, "capabilities": true,
}

func isScalar(v any) bool {
	switch v.(type) {
	case map[string]any:
		return false
	default:
		return true
	}
}

func scalarString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case []any:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, scalarString(e))
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(t)
	}
}
