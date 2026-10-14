package ui

import (
	"strings"

	"lshwtui/internal/lshw"
)

// Search holds the active query and its matches in tree order.
//
// Matching is a case-insensitive substring test over everything lshw knows
// about a node — not just its label — because the fields people actually hunt
// for (a serial number, "nvme", a MAC address, a PCI bus id) live in the
// detail fields and in the configuration/capabilities maps.
type Search struct {
	query   string
	matches []*lshw.Node
	set     map[*lshw.Node]bool
	idx     int
}

// Query returns the current query string.
func (s *Search) Query() string { return s.query }

// Active reports whether a query is in effect.
func (s *Search) Active() bool { return s.query != "" }

// Count is the number of matching nodes.
func (s *Search) Count() int { return len(s.matches) }

// Ordinal is the 1-based position of the current match, or 0 when there is none.
func (s *Search) Ordinal() int {
	if len(s.matches) == 0 {
		return 0
	}
	return s.idx + 1
}

// Matched reports whether n matches the active query.
func (s *Search) Matched(n *lshw.Node) bool { return s.set[n] }

// Current returns the match the cursor should sit on, or nil.
func (s *Search) Current() *lshw.Node {
	if len(s.matches) == 0 {
		return nil
	}
	return s.matches[s.idx]
}

// Clear drops the query and all matches.
func (s *Search) Clear() {
	s.query = ""
	s.matches = nil
	s.set = nil
	s.idx = 0
}

// Run re-evaluates the query against the tree, collecting matches in
// depth-first order so that n/N walk the list the way the eye reads it. near is
// the currently selected node: the match cursor lands on the first match at or
// after it, so refining a query does not throw the user back to the top.
func (s *Search) Run(root *lshw.Node, query string, near *lshw.Node) {
	s.query = query
	s.matches = nil
	s.set = map[*lshw.Node]bool{}
	s.idx = 0
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" || root == nil {
		return
	}

	passedNear := near == nil
	landed := false
	root.Walk(func(n *lshw.Node) {
		if n == near {
			passedNear = true
		}
		if !strings.Contains(haystack(n), needle) {
			return
		}
		s.matches = append(s.matches, n)
		s.set[n] = true
		if passedNear && !landed {
			s.idx = len(s.matches) - 1
			landed = true
		}
	})
}

// Next advances to the following match, wrapping around.
func (s *Search) Next() *lshw.Node {
	if len(s.matches) == 0 {
		return nil
	}
	s.idx = (s.idx + 1) % len(s.matches)
	return s.matches[s.idx]
}

// Prev steps back to the previous match, wrapping around.
func (s *Search) Prev() *lshw.Node {
	if len(s.matches) == 0 {
		return nil
	}
	s.idx = (s.idx - 1 + len(s.matches)) % len(s.matches)
	return s.matches[s.idx]
}

// haystack flattens everything searchable about a node into one lowercase
// string. Rebuilt per query rather than cached: the trees are small, and a
// cache would have to be invalidated on every reload.
func haystack(n *lshw.Node) string {
	var b strings.Builder
	for _, k := range n.Keys() {
		b.WriteString(strings.ToLower(k))
		b.WriteByte(' ')
		b.WriteString(strings.ToLower(n.Str(k)))
		b.WriteByte(' ')
	}
	for _, section := range []string{"configuration", "capabilities"} {
		for k, v := range n.Map(section) {
			b.WriteString(strings.ToLower(k))
			b.WriteByte(' ')
			b.WriteString(strings.ToLower(v))
			b.WriteByte(' ')
		}
	}
	return b.String()
}
