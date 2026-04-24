package evidence

import (
	"fmt"
	"strings"
)

type Grade string

const (
	GradeA Grade = "A"
	GradeB Grade = "B"
	GradeC Grade = "C"
	GradeD Grade = "D"
	GradeF Grade = "F"
)

type FactorRecord struct {
	Name       string
	Hypothesis string
	Grade      Grade
	TestMethod string
}

func (g Grade) Valid() bool {
	switch g {
	case GradeA, GradeB, GradeC, GradeD, GradeF:
		return true
	default:
		return false
	}
}

func (g Grade) EngineEligible() bool {
	switch g {
	case GradeA, GradeB, GradeC:
		return true
	default:
		return false
	}
}

func (r FactorRecord) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("factor name is required")
	}
	if strings.TrimSpace(r.Hypothesis) == "" {
		return fmt.Errorf("factor hypothesis is required")
	}
	if !r.Grade.Valid() {
		return fmt.Errorf("invalid evidence grade: %q", r.Grade)
	}
	if strings.TrimSpace(r.TestMethod) == "" {
		return fmt.Errorf("test method is required")
	}
	return nil
}
