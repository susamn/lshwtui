package lshw

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Parse decodes `lshw -json` output into a hardware tree.
//
// lshw emits either a single object or, when several top-level nodes exist, an
// array of them. An array is wrapped in a synthetic root so callers always get
// exactly one tree to render.
func Parse(data []byte) (*Node, error) {
	if len(data) == 0 {
		return nil, errors.New("lshw: empty output")
	}

	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("lshw: decoding JSON: %w", err)
	}

	switch v := raw.(type) {
	case map[string]any:
		return build(v, nil), nil
	case []any:
		roots := make([]*Node, 0, len(v))
		for _, e := range v {
			obj, ok := e.(map[string]any)
			if !ok {
				continue
			}
			roots = append(roots, build(obj, nil))
		}
		switch len(roots) {
		case 0:
			return nil, errors.New("lshw: no hardware nodes in output")
		case 1:
			return roots[0], nil
		}
		root := &Node{Raw: map[string]any{
			"id": "hardware", "class": "system", "description": "Hardware",
		}}
		for _, r := range roots {
			r.Parent = root
		}
		root.Children = roots
		return root, nil
	default:
		return nil, fmt.Errorf("lshw: unexpected top-level JSON type %T", raw)
	}
}

func build(obj map[string]any, parent *Node) *Node {
	n := &Node{Raw: obj, Parent: parent}
	kids, _ := obj["children"].([]any)
	for _, k := range kids {
		child, ok := k.(map[string]any)
		if !ok {
			continue
		}
		n.Children = append(n.Children, build(child, n))
	}
	return n
}
