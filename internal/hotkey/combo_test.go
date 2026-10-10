package hotkey

import "testing"

func TestStringParseRoundTrip(t *testing.T) {
	cases := []string{"Ctrl+Shift+W", "Ctrl+Alt+Z", "Alt+F4", "Ctrl+Shift+Win+F12", "Ctrl+1"}
	for _, in := range cases {
		h, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		if got := h.String(); got != in {
			t.Fatalf("round trip %q -> %q", in, got)
		}
		if h.Mods&ModNoRepeat == 0 {
			t.Fatalf("%q missing MOD_NOREPEAT", in)
		}
	}
}

func TestParseCaseInsensitive(t *testing.T) {
	a, err := Parse("ctrl+shift+w")
	if err != nil {
		t.Fatalf("lowercase parse: %v", err)
	}
	b, err := Parse("Ctrl+Shift+W")
	if err != nil {
		t.Fatalf("canonical parse: %v", err)
	}
	if a != b {
		t.Fatalf("case mismatch: %+v vs %+v", a, b)
	}
}

func TestParseRejects(t *testing.T) {
	bad := []string{"", "W", "Ctrl", "Ctrl+Shift+é", "Ctrl+Shift+Esc", "Shift+Nope"}
	for _, in := range bad {
		if _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) should fail", in)
		}
	}
}

func TestParseOrDefault(t *testing.T) {
	if h, ok := ParseOrDefault(""); !ok || h != Default {
		t.Fatalf("empty should give default+ok, got %+v ok=%v", h, ok)
	}
	if h, ok := ParseOrDefault("garbage"); ok || h != Default {
		t.Fatalf("garbage should give default+not-ok, got %+v ok=%v", h, ok)
	}
	h, ok := ParseOrDefault("Ctrl+Alt+Q")
	if !ok || h.VK != 0x51 {
		t.Fatalf("valid parse failed: %+v ok=%v", h, ok)
	}
}

func TestDefaultString(t *testing.T) {
	if got := Default.String(); got != "Ctrl+Shift+W" {
		t.Fatalf("default = %q", got)
	}
}
