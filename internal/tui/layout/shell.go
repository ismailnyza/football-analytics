package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ismael/football-analytics/internal/tui/components"
)

// ViewModel contains the shell state needed for rendering.
type ViewModel struct {
	AppName     string
	Version     string
	Width       int
	Height      int
	Sections    []string
	Selected    int
	Active      string
	FocusedNav  bool
	MainContent string
	StatusNote  string
	HelpContent string
	ShowHelp    bool
	FocusLabel  string
}

var (
	pageFg = lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#e8e8e8"}
	pageBg = lipgloss.AdaptiveColor{Light: "#f2f2f2", Dark: "#1e1e2e"}

	pageStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Foreground(pageFg).
			Background(pageBg)

	topBarStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#2563eb", Dark: "#3d59a1"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"}).
			Padding(0, 1).
			Bold(true)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#64748b", Dark: "#6b7089"}).
			Background(lipgloss.AdaptiveColor{Light: "#e8e8ec", Dark: "#2a2a3d"}).
			Padding(1)

	panelTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1d4ed8", Dark: "#7aa2f7"}).
			Bold(true)

	selectedNavStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#c0caf5"}).
				Background(lipgloss.AdaptiveColor{Light: "#1d4ed8", Dark: "#565f89"}).
				Padding(0, 1)

	navStyle = lipgloss.NewStyle().
			Foreground(pageFg).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#334155", Dark: "#c0caf5"}).
			Background(lipgloss.AdaptiveColor{Light: "#cbd5e1", Dark: "#363b54"}).
			Padding(0, 1)
)

// RenderAppShell renders the root terminal shell.
func RenderAppShell(vm ViewModel) string {
	contentWidth := max(vm.Width-4, 60)
	navWidth := min(26, max(20, contentWidth/4))
	mainWidth := max(contentWidth-navWidth-1, 30)
	bodyHeight := max(vm.Height-7, 12)

	top := topBarStyle.Width(contentWidth).Render(vm.AppName + "  |  " + vm.Version + "  |  focus: " + vm.FocusLabel)

	navItems := make([]string, 0, len(vm.Sections))
	for idx, section := range vm.Sections {
		style := navStyle
		if idx == vm.Selected {
			style = selectedNavStyle
		}
		navItems = append(navItems, style.Render(section))
	}

	navTitle := panelTitleStyle.Render("Navigation")
	if vm.FocusedNav {
		navTitle += " [focused]"
	}
	navPanel := panelStyle.Width(navWidth).Height(bodyHeight).Render(
		navTitle + "\n\n" + strings.Join(navItems, "\n"),
	)

	mainTitle := panelTitleStyle.Render("Main Content")
	mainPanel := panelStyle.Width(mainWidth).Height(bodyHeight).Render(
		mainTitle + "\n\n" + vm.MainContent,
	)

	body := lipgloss.JoinHorizontal(lipgloss.Top, navPanel, mainPanel)

	status := statusStyle.Width(contentWidth).Render("Status: " + vm.StatusNote)
	if vm.ShowHelp {
		status += "\n" + components.RenderHelp(vm.HelpContent)
	}

	return pageStyle.Width(vm.Width).Render(lipgloss.JoinVertical(lipgloss.Left, top, body, status))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
