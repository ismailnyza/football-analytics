package sqlite

import "testing"

func TestDefaultPath(t *testing.T) {
	got := DefaultPath("/tmp/project-data")
	want := "/tmp/project-data/football.db"
	if got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}
