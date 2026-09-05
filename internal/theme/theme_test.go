package theme

import "testing"

// documentedClasses mirrors the minimum set of lshw device classes the
// theme package commits to covering.
var documentedClasses = []string{
	"system", "bus", "memory", "processor", "bridge", "display",
	"multimedia", "communication", "network", "storage", "disk",
	"volume", "input", "generic", "power", "battery", "printer", "tape",
}

func TestGlyphCoversDocumentedClasses(t *testing.T) {
	for _, mode := range []bool{true, false} {
		th := New(mode)
		for _, class := range documentedClasses {
			g := th.Glyph(class)
			if g == "" {
				t.Errorf("Glyph(%q) with glyphs=%v returned empty string", class, mode)
			}
		}
	}
}

func TestGlyphUnknownClassFallsBack(t *testing.T) {
	for _, mode := range []bool{true, false} {
		th := New(mode)
		got := th.Glyph("totally-unknown-class")
		if got == "" {
			t.Errorf("Glyph(unknown) with glyphs=%v returned empty string, want fallback", mode)
		}
		want := th.Glyph("default")
		if got != want {
			t.Errorf("Glyph(unknown) = %q, want fallback %q", got, want)
		}
	}
}

func TestClassStyleUnknownClassFallsBack(t *testing.T) {
	th := New(true)
	// Just verify it doesn't panic and returns a usable style for both a
	// known and an unknown class.
	if th.ClassStyle("processor").GetForeground() == nil {
		t.Errorf("ClassStyle(processor) has no foreground set")
	}
	if th.ClassStyle("totally-unknown-class").GetForeground() == nil {
		t.Errorf("ClassStyle(unknown) has no foreground set")
	}
}

func TestExpandedCollapsedMarkers(t *testing.T) {
	glyphTheme := New(true)
	if glyphTheme.Expanded() == "" || glyphTheme.Collapsed() == "" {
		t.Errorf("glyph-mode markers must be non-empty")
	}
	if glyphTheme.Expanded() == glyphTheme.Collapsed() {
		t.Errorf("Expanded() and Collapsed() must differ")
	}

	asciiTheme := New(false)
	if asciiTheme.Expanded() != "-" {
		t.Errorf("ascii Expanded() = %q, want %q", asciiTheme.Expanded(), "-")
	}
	if asciiTheme.Collapsed() != "+" {
		t.Errorf("ascii Collapsed() = %q, want %q", asciiTheme.Collapsed(), "+")
	}
}
