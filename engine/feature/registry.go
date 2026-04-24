package feature

import (
	"fmt"
	"sort"

	"github.com/ismailnyza/football-analytics/engine/evidence"
)

type Feature interface {
	Name() string
	Description() string
	Grade() evidence.Grade
	Enabled() bool
	Category() string
}

type MatchContext struct {
	HomeTeam       string
	AwayTeam       string
	HomeRating     float64
	AwayRating     float64
	HomeForm       float64
	AwayForm       float64
	Season         string
	HomeLastNGames []MatchResult
	AwayLastNGames []MatchResult
}

type MatchResult struct {
	Result       string
	GoalsFor     int
	GoalsAgainst int
	Opponent     string
	Venue        string
}

type FeatureOutput struct {
	FeatureName string
	Value       float64
	Weight      float64
}

type Registry struct {
	features map[string]Feature
}

func NewRegistry() *Registry {
	return &Registry{features: make(map[string]Feature)}
}

func (r *Registry) Register(f Feature) error {
	if f.Name() == "" {
		return fmt.Errorf("feature name required")
	}
	if _, exists := r.features[f.Name()]; exists {
		return fmt.Errorf("feature %q already registered", f.Name())
	}
	if !f.Grade().EngineEligible() {
		return fmt.Errorf("feature %q has grade %s, not eligible for engine", f.Name(), f.Grade())
	}
	r.features[f.Name()] = f
	return nil
}

func (r *Registry) Get(name string) (Feature, bool) {
	f, ok := r.features[name]
	return f, ok
}

func (r *Registry) List() []Feature {
	names := make([]string, 0, len(r.features))
	for name := range r.features {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]Feature, len(names))
	for i, name := range names {
		result[i] = r.features[name]
	}
	return result
}

func (r *Registry) ListEnabled() []Feature {
	var result []Feature
	for _, f := range r.features {
		if f.Enabled() {
			result = append(result, f)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result
}

func (r *Registry) Count() int {
	return len(r.features)
}

func (r *Registry) CountEnabled() int {
	n := 0
	for _, f := range r.features {
		if f.Enabled() {
			n++
		}
	}
	return n
}
