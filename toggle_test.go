package main

import "testing"

func TestToggleFor(t *testing.T) {
	cases := []struct {
		name   string
		hidden bool
		mode   string
		want   toggleAction
	}{
		{"hidden window opens chat", true, "pet", toggleShowChat},
		{"hidden panel opens chat", true, "panel", toggleShowChat},
		{"pet opens chat", false, "pet", toggleShowChat},
		{"startup (no mode yet) is the pet", false, "", toggleShowChat},
		{"open panel goes back to the pet", false, "panel", toggleToPet},
	}
	for _, c := range cases {
		if got := toggleFor(c.hidden, c.mode); got != c.want {
			t.Errorf("%s: toggleFor(%v, %q) = %v, want %v", c.name, c.hidden, c.mode, got, c.want)
		}
	}
}
