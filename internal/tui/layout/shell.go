package layout

import (
	"fmt"
	"strings"

	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/tui/keymap"
)

// RenderBootstrapShell returns a static shell preview for the initial repo bootstrap.
func RenderBootstrapShell(cfg app.Config) string {
	var b strings.Builder
	b.WriteString(cfg.Name)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("=", len(cfg.Name)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Version: %s\n\n", cfg.Version))
	b.WriteString("Top Bar    : bootstrap shell\n")
	b.WriteString("Navigation :\n")
	for _, section := range keymap.DefaultSections {
		b.WriteString("  - ")
		b.WriteString(section)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString("Main View  : SEC-007 pending\n")
	b.WriteString("Status Bar : buildable scaffold only\n")
	return b.String()
}
