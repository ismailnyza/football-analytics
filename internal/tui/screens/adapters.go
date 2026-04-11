package screens

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ismael/football-analytics/internal/tui/screens/club"
	"github.com/ismael/football-analytics/internal/tui/screens/dashboard"
	"github.com/ismael/football-analytics/internal/tui/screens/dataimport"
	"github.com/ismael/football-analytics/internal/tui/screens/matchlab"
	"github.com/ismael/football-analytics/internal/tui/screens/playerdetail"
	"github.com/ismael/football-analytics/internal/tui/screens/scenarios"
	"github.com/ismael/football-analytics/internal/tui/screens/seasonlab"
	"github.com/ismael/football-analytics/internal/tui/screens/squad"
	"github.com/ismael/football-analytics/internal/tui/screens/transfers"
	"github.com/ismael/football-analytics/internal/tui/screens/world"
)

type matchlabAdapter struct{ model matchlab.Model }

func (a *matchlabAdapter) Init() tea.Cmd { return a.model.Init() }
func (a *matchlabAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}
func (a *matchlabAdapter) View(w, h int) string { return a.model.View(w, h) }

type seasonlabAdapter struct{ model seasonlab.Model }

func (a *seasonlabAdapter) Init() tea.Cmd { return a.model.Init() }
func (a *seasonlabAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}
func (a *seasonlabAdapter) View(w, h int) string { return a.model.View(w, h) }

type dashboardAdapter struct{ model dashboard.Model }

func (a *dashboardAdapter) Init() tea.Cmd { return a.model.Init() }
func (a *dashboardAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}
func (a *dashboardAdapter) View(w, h int) string { return a.model.View(w, h) }

type squadAdapter struct{ model squad.Model }

func (a *squadAdapter) Init() tea.Cmd { return a.model.Init() }
func (a *squadAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}
func (a *squadAdapter) View(w, h int) string { return a.model.View(w, h) }

type playerDetailAdapter struct{ model playerdetail.Model }

func (a *playerDetailAdapter) Init() tea.Cmd { return a.model.Init() }
func (a *playerDetailAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}
func (a *playerDetailAdapter) View(w, h int) string { return a.model.View(w, h) }

type clubAdapter struct{ model club.Model }

func (a *clubAdapter) Init() tea.Cmd { return a.model.Init() }
func (a *clubAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}
func (a *clubAdapter) View(w, h int) string { return a.model.View(w, h) }

type transfersAdapter struct{ model transfers.Model }

func (a *transfersAdapter) Init() tea.Cmd { return a.model.Init() }
func (a *transfersAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}
func (a *transfersAdapter) View(w, h int) string { return a.model.View(w, h) }

type worldAdapter struct{ model world.Model }

func (a *worldAdapter) Init() tea.Cmd { return a.model.Init() }
func (a *worldAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}
func (a *worldAdapter) View(w, h int) string { return a.model.View(w, h) }

type scenariosAdapter struct{ model scenarios.Model }

func (a *scenariosAdapter) Init() tea.Cmd { return a.model.Init() }
func (a *scenariosAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}
func (a *scenariosAdapter) View(w, h int) string { return a.model.View(w, h) }

type dataimportAdapter struct{ model dataimport.Model }

func (a *dataimportAdapter) Init() tea.Cmd { return a.model.Init() }
func (a *dataimportAdapter) Update(msg tea.Msg) (Screen, tea.Cmd) {
	updated, cmd := a.model.Update(msg)
	a.model = updated
	return a, cmd
}
func (a *dataimportAdapter) View(w, h int) string { return a.model.View(w, h) }
