package lshw

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSingleObject(t *testing.T) {
	root, err := Parse([]byte(`{"id":"box","class":"system","children":[
		{"id":"cpu","class":"processor","product":"Test CPU","clock":3600000000}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if root.ID() != "box" || len(root.Children) != 1 {
		t.Fatalf("unexpected root %+v", root.Raw)
	}
	cpu := root.Children[0]
	if cpu.Parent != root {
		t.Error("child is not linked back to its parent")
	}
	if cpu.Label() != "Test CPU" {
		t.Errorf("label = %q, want the product name", cpu.Label())
	}
	if cpu.Depth() != 1 {
		t.Errorf("depth = %d, want 1", cpu.Depth())
	}
}

func TestParseTopLevelArrayGetsSyntheticRoot(t *testing.T) {
	root, err := Parse([]byte(`[{"id":"a","class":"system"},{"id":"b","class":"system"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(root.Children) != 2 {
		t.Fatalf("got %d children, want the two top-level nodes", len(root.Children))
	}
	if root.ID() != "hardware" {
		t.Errorf("synthetic root id = %q", root.ID())
	}
}

func TestParseSingleElementArrayIsUnwrapped(t *testing.T) {
	root, err := Parse([]byte(`[{"id":"only","class":"system"}]`))
	if err != nil {
		t.Fatal(err)
	}
	if root.ID() != "only" {
		t.Errorf("root id = %q, want the sole node itself", root.ID())
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	for name, input := range map[string]string{
		"empty":   ``,
		"invalid": `{not json`,
		"scalar":  `42`,
		"noNodes": `[]`,
	} {
		if _, err := Parse([]byte(input)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestAccessorsTolerateMissingAndOddTypes(t *testing.T) {
	root, err := Parse([]byte(`{"id":"x","class":"disk","logicalname":["/dev/sda","/dev/sdb"],
		"configuration":{"driver":"nvme","ncq":true},"size":1024}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := root.List("logicalname"); len(got) != 2 || got[0] != "/dev/sda" {
		t.Errorf("List = %v", got)
	}
	if got := root.Map("configuration")["ncq"]; got != "true" {
		t.Errorf("bool config value rendered as %q", got)
	}
	if got := root.Str("nonexistent"); got != "" {
		t.Errorf("missing key returned %q", got)
	}
	if _, ok := root.Num("class"); ok {
		t.Error("Num on a string field reported success")
	}
	if got := root.Str("size"); got != "1024" {
		t.Errorf("numeric field rendered as %q, want an integer without a decimal point", got)
	}
}

func TestSnapshotRoundTrip(t *testing.T) {
	data := []byte(`{"id":"snap","class":"system"}`)
	path := filepath.Join(t.TempDir(), "nested", "snap.json")

	if err := SaveFile(path, data); err != nil {
		t.Fatal(err)
	}
	got, src, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Errorf("round-tripped %q", got)
	}
	if !src.Snapshot || src.Describe() != "snapshot snap.json" {
		t.Errorf("source = %+v", src)
	}
}

func TestLoadFileReportsMissingFile(t *testing.T) {
	if _, _, err := LoadFile(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Error("expected an error for a missing snapshot")
	}
}

func TestSourceDescribe(t *testing.T) {
	cases := []struct {
		src  Source
		want string
	}{
		{Source{Origin: "lshw"}, "live (partial)"},
		{Source{Origin: "sudo lshw", Privileged: true}, "live (root)"},
		{Source{Origin: "/tmp/a/b.json", Snapshot: true}, "snapshot b.json"},
	}
	for _, c := range cases {
		if got := c.src.Describe(); got != c.want {
			t.Errorf("Describe() = %q, want %q", got, c.want)
		}
	}
}

func TestParseRealCapture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/sample.json")
	if err != nil {
		t.Skipf("no sample capture: %v", err)
	}
	root, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	root.Walk(func(node *Node) {
		n++
		if node.Class() == "" {
			t.Errorf("node %q has no class", node.ID())
		}
	})
	if n < 10 {
		t.Errorf("walked only %d nodes", n)
	}
}
