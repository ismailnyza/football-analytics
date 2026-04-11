package layout

import "testing"

func TestTerminalLayoutSize_zeroFallsBack(t *testing.T) {
	w, h := TerminalLayoutSize(0, 0)
	if w < 80 || h < 20 {
		t.Fatalf("got %d x %d", w, h)
	}
}

func TestTerminalLayoutSize_brokenTinyUsesFloor(t *testing.T) {
	w, h := TerminalLayoutSize(3, 5)
	if w != 80 || h != 24 {
		t.Fatalf("got %d x %d", w, h)
	}
}

func TestTerminalLayoutSize_normalPassthrough(t *testing.T) {
	w, h := TerminalLayoutSize(100, 40)
	if w != 100 || h != 40 {
		t.Fatalf("got %d x %d", w, h)
	}
}
