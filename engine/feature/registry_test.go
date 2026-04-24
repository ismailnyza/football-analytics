package feature

import (
	"testing"

	"github.com/ismailnyza/football-analytics/engine/evidence"
)

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()
	f := NewBaseFeature("team_strength", "historical team strength baseline", "team", evidence.GradeB, true)
	if err := r.Register(f); err != nil {
		t.Fatalf("register: %v", err)
	}
	if r.Count() != 1 {
		t.Fatalf("count = %d, want 1", r.Count())
	}
	if r.CountEnabled() != 1 {
		t.Fatalf("enabled count = %d, want 1", r.CountEnabled())
	}
}

func TestRegistryDuplicateReject(t *testing.T) {
	r := NewRegistry()
	f1 := NewBaseFeature("dup", "first", "test", evidence.GradeB, true)
	f2 := NewBaseFeature("dup", "second", "test", evidence.GradeB, true)
	if err := r.Register(f1); err != nil {
		t.Fatalf("register f1: %v", err)
	}
	if err := r.Register(f2); err == nil {
		t.Fatalf("expected duplicate error")
	}
}

func TestRegistryGradeGate(t *testing.T) {
	r := NewRegistry()
	f := NewBaseFeature("bad", "should fail", "test", evidence.GradeD, true)
	if err := r.Register(f); err == nil {
		t.Fatalf("expected grade D rejection")
	}
	f2 := NewBaseFeature("bad2", "should also fail", "test", evidence.GradeF, false)
	if err := r.Register(f2); err == nil {
		t.Fatalf("expected grade F rejection")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	f1 := NewBaseFeature("team_strength", "a", "team", evidence.GradeB, true)
	f2 := NewBaseFeature("form", "b", "form", evidence.GradeC, false)
	r.Register(f1)
	r.Register(f2)

	list := r.List()
	if len(list) != 2 {
		t.Fatalf("list len = %d, want 2", len(list))
	}
	if list[0].Name() != "form" || list[1].Name() != "team_strength" {
		t.Fatalf("list not sorted: %v %v", list[0].Name(), list[1].Name())
	}

	enabled := r.ListEnabled()
	if len(enabled) != 1 || enabled[0].Name() != "team_strength" {
		t.Fatalf("enabled list wrong: %d items", len(enabled))
	}
}
