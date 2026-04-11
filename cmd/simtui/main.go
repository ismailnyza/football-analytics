package main

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/tui"
)

func main() {
	cfg := app.DefaultConfig()
	cfg, closeStore, err := app.OpenLocalStore(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := closeStore(); err != nil {
			log.Printf("close store: %v", err)
		}
	}()

	program := tea.NewProgram(tui.NewModel(cfg), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
