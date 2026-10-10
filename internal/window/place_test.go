package window

import "testing"

// 1920x1080 monitor with a 48px taskbar at the bottom.
var work = Rect{Left: 0, Top: 0, Right: 1920, Bottom: 1032}

func TestFitKeepsPositionWhenWindowFits(t *testing.T) {
	x, y := Fit(100, 100, 400, 600, work)
	if x != 100 || y != 100 {
		t.Fatalf("got %d,%d want 100,100", x, y)
	}
}

func TestFitPushesUpFromBottomEdge(t *testing.T) {
	x, y := Fit(100, 850, 400, 600, work)
	if x != 100 || y != 432 {
		t.Fatalf("got %d,%d want 100,432", x, y)
	}
}

func TestFitPushesLeftFromRightEdge(t *testing.T) {
	x, y := Fit(1800, 100, 400, 600, work)
	if x != 1520 || y != 100 {
		t.Fatalf("got %d,%d want 1520,100", x, y)
	}
}

func TestFitPullsBackWindowHangingOffTopLeft(t *testing.T) {
	x, y := Fit(-50, -30, 400, 600, work)
	if x != 0 || y != 0 {
		t.Fatalf("got %d,%d want 0,0", x, y)
	}
}

func TestFitRespectsOffsetMonitor(t *testing.T) {
	// Second monitor left of the primary, taskbar-free.
	left := Rect{Left: -1280, Top: 0, Right: 0, Bottom: 1024}
	x, y := Fit(-200, 900, 400, 600, left)
	if x != -400 || y != 424 {
		t.Fatalf("got %d,%d want -400,424", x, y)
	}
}

func TestFitPinsOversizedWindowToTopLeft(t *testing.T) {
	small := Rect{Left: 0, Top: 0, Right: 800, Bottom: 500}
	x, y := Fit(300, 300, 1000, 600, small)
	if x != 0 || y != 0 {
		t.Fatalf("got %d,%d want 0,0", x, y)
	}
}

func TestPlacerReturnsPetToItsSpotAfterPanel(t *testing.T) {
	var p Placer
	// Pet sits in the bottom-right corner.
	px, py := 1700, 820
	cx, cy := p.Place(Panel, px, py, 400, 600, work)
	if cx != 1520 || cy != 432 {
		t.Fatalf("panel at %d,%d want 1520,432", cx, cy)
	}
	bx, by := p.Place(Pet, cx, cy, 180, 180, work)
	if bx != px || by != py {
		t.Fatalf("pet back at %d,%d want %d,%d", bx, by, px, py)
	}
}

func TestPlacerPetFollowsPanelMovedByUser(t *testing.T) {
	var p Placer
	cx, cy := p.Place(Panel, 100, 850, 400, 600, work) // pushed up by 418
	// User drags the panel 300px right while in chat.
	bx, by := p.Place(Pet, cx+300, cy, 180, 180, work)
	if bx != 400 || by != 850 {
		t.Fatalf("pet at %d,%d want 400,850", bx, by)
	}
}

func TestPlacerRepeatedPanelKeepsOriginalShift(t *testing.T) {
	var p Placer
	cx, cy := p.Place(Panel, 100, 850, 400, 600, work)
	// Tray "Show" while already in panel mode must not reset the shift.
	cx, cy = p.Place(Panel, cx, cy, 400, 600, work)
	bx, by := p.Place(Pet, cx, cy, 180, 180, work)
	if bx != 100 || by != 850 {
		t.Fatalf("pet at %d,%d want 100,850", bx, by)
	}
}

func TestPlacerClampsPetWhenUndoWouldLeaveScreen(t *testing.T) {
	var p Placer
	cx, cy := p.Place(Panel, 100, 850, 400, 600, work)
	// Panel dragged to the very bottom: undoing the shift would put the pet
	// below the work area, so it is clamped back inside.
	bx, by := p.Place(Pet, cx, cy+400, 180, 180, work)
	if bx != 100 || by != 852 {
		t.Fatalf("pet at %d,%d want 100,852", bx, by)
	}
}

func TestConstrainKeepsDraggedRectInsideWorkArea(t *testing.T) {
	// Pet dragged so it hangs 100px below the work area and 50px past the
	// right edge: same size, pulled back to the corner.
	got := Constrain(Rect{Left: 1790, Top: 952, Right: 1970, Bottom: 1132}, work)
	want := Rect{Left: 1740, Top: 852, Right: 1920, Bottom: 1032}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestConstrainLeavesRectInsideUntouched(t *testing.T) {
	in := Rect{Left: 10, Top: 20, Right: 190, Bottom: 200}
	if got := Constrain(in, work); got != in {
		t.Fatalf("got %+v want %+v", got, in)
	}
}

func TestSizeRejectsUnknownMode(t *testing.T) {
	if _, err := SizeOf(Mode("full")); err == nil {
		t.Fatal("want error for unknown mode")
	}
	if s, err := SizeOf(Panel); err != nil || s != PanelSize {
		t.Fatalf("panel size %v err %v", s, err)
	}
}
