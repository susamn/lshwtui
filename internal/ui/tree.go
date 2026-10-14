package ui

import (
	"strings"

	"lshwtui/internal/lshw"
)

// Row is one rendered line of the tree pane: a node plus the drawing state
// needed to place it, precomputed so rendering stays a pure formatting step.
type Row struct {
	Node     *lshw.Node
	Depth    int
	Guide    string // indent guides for the ancestors, e.g. "│  │  "
	Branch   string // this node's own connector, "├─ " or "└─ "
	HasKids  bool
	Expanded bool
}

// Tree holds the hardware tree together with its expand/collapse state and a
// flattened view of the currently visible rows.
type Tree struct {
	Root     *lshw.Node
	expanded map[*lshw.Node]bool
	rows     []Row
	cursor   int
}

// NewTree builds a tree with every node expanded. lshw trees are small enough
// (tens of nodes) that showing everything up front is more useful than making
// the user drill down to find a disk.
func NewTree(root *lshw.Node) *Tree {
	t := &Tree{Root: root, expanded: map[*lshw.Node]bool{}}
	root.Walk(func(n *lshw.Node) {
		if len(n.Children) > 0 {
			t.expanded[n] = true
		}
	})
	t.rebuild()
	return t
}

// Rows returns the currently visible rows.
func (t *Tree) Rows() []Row { return t.rows }

// Len is the number of visible rows.
func (t *Tree) Len() int { return len(t.rows) }

// Cursor is the index of the selected row.
func (t *Tree) Cursor() int { return t.cursor }

// Selected returns the node under the cursor, or nil when the tree is empty.
func (t *Tree) Selected() *lshw.Node {
	if t.cursor < 0 || t.cursor >= len(t.rows) {
		return nil
	}
	return t.rows[t.cursor].Node
}

func (t *Tree) rebuild() {
	t.rows = t.rows[:0]
	t.appendNode(t.Root, "", "", true)
	t.clampCursor()
}

func (t *Tree) appendNode(n *lshw.Node, guide, branch string, isRoot bool) {
	t.rows = append(t.rows, Row{
		Node:     n,
		Depth:    n.Depth(),
		Guide:    guide,
		Branch:   branch,
		HasKids:  len(n.Children) > 0,
		Expanded: t.expanded[n],
	})
	if len(n.Children) == 0 || !t.expanded[n] {
		return
	}

	childGuide := guide
	if !isRoot {
		// Continue the vertical rule under this node only while it still has
		// siblings below it; otherwise the subtree hangs in clear space.
		if strings.HasPrefix(branch, "└") {
			childGuide += "   "
		} else {
			childGuide += "│  "
		}
	}
	for i, c := range n.Children {
		b := "├─ "
		if i == len(n.Children)-1 {
			b = "└─ "
		}
		t.appendNode(c, childGuide, b, false)
	}
}

// SetCursor moves the cursor to an absolute row index, clamped to the tree.
func (t *Tree) SetCursor(i int) {
	t.cursor = i
	t.clampCursor()
}

// Move shifts the cursor by delta rows.
func (t *Tree) Move(delta int) { t.SetCursor(t.cursor + delta) }

func (t *Tree) clampCursor() {
	if len(t.rows) == 0 {
		t.cursor = 0
		return
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
	if t.cursor >= len(t.rows) {
		t.cursor = len(t.rows) - 1
	}
}

// Toggle expands or collapses the selected node.
func (t *Tree) Toggle() {
	n := t.Selected()
	if n == nil || len(n.Children) == 0 {
		return
	}
	t.setExpanded(n, !t.expanded[n])
}

// Expand opens the selected node, or steps into it when already open.
func (t *Tree) Expand() {
	n := t.Selected()
	if n == nil || len(n.Children) == 0 {
		return
	}
	if !t.expanded[n] {
		t.setExpanded(n, true)
		return
	}
	t.Move(1)
}

// Collapse closes the selected node, or jumps to its parent when already
// closed — the usual vim-tree behaviour for `h`.
func (t *Tree) Collapse() {
	n := t.Selected()
	if n == nil {
		return
	}
	if t.expanded[n] && len(n.Children) > 0 {
		t.setExpanded(n, false)
		return
	}
	if n.Parent != nil {
		t.Reveal(n.Parent)
	}
}

// SetAllExpanded expands or collapses every node with children.
func (t *Tree) SetAllExpanded(open bool) {
	sel := t.Selected()
	t.Root.Walk(func(n *lshw.Node) {
		if len(n.Children) > 0 {
			t.expanded[n] = open
		}
	})
	t.expanded[t.Root] = true
	t.rebuild()
	if sel != nil {
		if open {
			t.Reveal(sel)
		} else {
			curr := sel
			for curr != nil && t.IndexOf(curr) == -1 {
				curr = curr.Parent
			}
			if idx := t.IndexOf(curr); idx != -1 {
				t.cursor = idx
			} else {
				t.clampCursor()
			}
		}
	}
}

func (t *Tree) setExpanded(n *lshw.Node, open bool) {
	sel := t.Selected()
	t.expanded[n] = open
	t.rebuild()
	if sel != nil {
		if open {
			t.Reveal(sel)
		} else {
			curr := sel
			for curr != nil && t.IndexOf(curr) == -1 {
				curr = curr.Parent
			}
			if idx := t.IndexOf(curr); idx != -1 {
				t.cursor = idx
			} else {
				t.clampCursor()
			}
		}
	}
}

// Reveal expands whatever is needed to make n visible and puts the cursor on
// it. Nodes hidden inside collapsed parents are the main reason search results
// would otherwise be unreachable.
func (t *Tree) Reveal(n *lshw.Node) {
	changed := false
	for p := n.Parent; p != nil; p = p.Parent {
		if !t.expanded[p] {
			t.expanded[p] = true
			changed = true
		}
	}
	if changed {
		t.rebuild()
	}
	for i, r := range t.rows {
		if r.Node == n {
			t.cursor = i
			return
		}
	}
}

// IndexOf returns the visible row index of n, or -1 when it is hidden.
func (t *Tree) IndexOf(n *lshw.Node) int {
	for i, r := range t.rows {
		if r.Node == n {
			return i
		}
	}
	return -1
}
