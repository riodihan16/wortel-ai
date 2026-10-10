package main

import "testing"

func TestToggleFor(t *testing.T) {
	cases := []struct {
		name    string
		hidden  bool
		mode    string
		focused bool
		want    toggleAction
	}{
		{"hidden window opens chat", true, "pet", false, toggleShowChat},
		{"hidden panel opens chat", true, "panel", false, toggleShowChat},
		{"pet opens chat", false, "pet", true, toggleShowChat},
		{"startup (no mode yet) is the pet", false, "", false, toggleShowChat},
		{"focused open panel goes back to the pet", false, "panel", true, toggleToPet},
		{"open panel behind another app comes forward", false, "panel", false, toggleShowChat},
	}
	for _, c := range cases {
		if got := toggleFor(c.hidden, c.mode, c.focused); got != c.want {
			t.Errorf("%s: toggleFor(%v, %q, %v) = %v, want %v", c.name, c.hidden, c.mode, c.focused, got, c.want)
		}
	}
}
