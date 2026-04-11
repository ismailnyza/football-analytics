package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/tui"
)

func main() {
	cfg := app.DefaultConfig()
	program := tea.NewProgram(tui.NewModel(cfg), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
