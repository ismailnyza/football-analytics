package feature

import "github.com/ismailnyza/football-analytics/engine/evidence"

type BaseFeature struct {
	name        string
	description string
	grade       evidence.Grade
	enabled     bool
	category    string
}

func NewBaseFeature(name, description, category string, grade evidence.Grade, enabled bool) BaseFeature {
	return BaseFeature{
		name:        name,
		description: description,
		category:    category,
		grade:       grade,
		enabled:     enabled,
	}
}

func (f BaseFeature) Name() string          { return f.name }
func (f BaseFeature) Description() string   { return f.description }
func (f BaseFeature) Grade() evidence.Grade { return f.grade }
func (f BaseFeature) Enabled() bool         { return f.enabled }
func (f BaseFeature) Category() string      { return f.category }
