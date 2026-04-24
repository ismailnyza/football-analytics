package evidence

import "testing"

func TestGradeEngineEligibility(t *testing.T) {
	cases := []struct {
		grade    Grade
		eligible bool
	}{
		{GradeA, true},
		{GradeB, true},
		{GradeC, true},
		{GradeD, false},
		{GradeF, false},
	}

	for _, tc := range cases {
		if got := tc.grade.EngineEligible(); got != tc.eligible {
			t.Fatalf("grade %s eligibility = %v, want %v", tc.grade, got, tc.eligible)
		}
	}
}

func TestFactorRecordValidate(t *testing.T) {
	valid := FactorRecord{
		Name:       "historical team strength baseline",
		Hypothesis: "stronger teams should improve outcome prediction",
		Grade:      GradeD,
		TestMethod: "planned backtest",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid factor record, got error: %v", err)
	}

	invalid := []FactorRecord{
		{},
		{Name: "x", Hypothesis: "y", Grade: Grade("Z"), TestMethod: "z"},
		{Name: "x", Hypothesis: " ", Grade: GradeA, TestMethod: "z"},
		{Name: "x", Hypothesis: "y", Grade: GradeA, TestMethod: " "},
	}

	for _, record := range invalid {
		if err := record.Validate(); err == nil {
			t.Fatalf("expected validation error for %#v", record)
		}
	}
}
