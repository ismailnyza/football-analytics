package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ismael/football-analytics/internal/tui/screens/matchlab"
)

// matchlabAdapter wraps matchlab.Model so it satisfies the Screen interface.
type matchlabAdapter struct {
	model matchlab.Model
}

func (a *matchlabAdapter) Init() tea.Cmd { return a.model.Init() }

func (a *matchlabAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}

func (a *matchlabAdapter) View(width, height int) string {
	return a.model.View(width, height)
}
