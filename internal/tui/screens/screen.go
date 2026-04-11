package screens

import tea "github.com/charmbracelet/bubbletea"

// Screen is a self-contained interactive TUI panel rendered inside the app shell.
type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View(width, height int) string
}
