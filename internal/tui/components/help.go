package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var helpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("252")).
	Background(lipgloss.Color("236")).
	Padding(0, 1)

// RenderHelp renders a single-line help/status strip.
func RenderHelp(content string) string {
	return helpStyle.Render(strings.TrimSpace(content))
}
