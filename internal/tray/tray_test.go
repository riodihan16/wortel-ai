package tray

import (
	"testing"
	"unsafe"
)

func TestNIDSize(t *testing.T) {
	if got := unsafe.Sizeof(notifyIconData{}); got != wantNIDSize {
		t.Fatalf("sizeof(notifyIconData)=%d want %d", got, wantNIDSize)
	}
}

func TestActionForMenuID(t *testing.T) {
	cases := map[uint32]Action{
		1: ActionShow, 2: ActionSettings, 3: ActionHide, 4: ActionExit,
		0: ActionNone, 5: ActionNone, 999: ActionNone,
	}
	for id, want := range cases {
		if got := ActionForMenuID(id); got != want {
			t.Fatalf("id %d: got %v want %v", id, got, want)
		}
	}
}

func TestDispatchNilSafe(t *testing.T) {
	var c Callbacks // all nil: must not panic
	for _, a := range []Action{ActionShow, ActionSettings, ActionHide, ActionExit, ActionNone} {
		c.dispatch(a)
	}
}

func TestDispatchCalls(t *testing.T) {
	var got []Action
	c := Callbacks{
		OnShow:     func() { got = append(got, ActionShow) },
		OnSettings: func() { got = append(got, ActionSettings) },
		OnHide:     func() { got = append(got, ActionHide) },
		OnExit:     func() { got = append(got, ActionExit) },
	}
	for _, a := range []Action{ActionShow, ActionSettings, ActionHide, ActionExit} {
		c.dispatch(a)
	}
	if len(got) != 4 {
		t.Fatalf("got %v", got)
	}
	for i, want := range []Action{ActionShow, ActionSettings, ActionHide, ActionExit} {
		if got[i] != want {
			t.Fatalf("slot %d: got %v want %v", i, got[i], want)
		}
	}
}
