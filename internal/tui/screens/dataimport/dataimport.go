package dataimport

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/ingestion"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	okStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

type SourceStatus struct {
	Name      string
	LastFetch time.Time
	Records   int
	Errors    int
	Published int
	CapPerRun int
	LastError string
}

type Model struct {
	stateDir   string
	sources    []SourceStatus
	staged     []ingestion.SourceRecord
	preview    []string
	cursor     int
	scroll     int
	status     string
	fbrefURL   string
	urlInput   textinput.Model
	editingURL bool
}

type fetchFinishedMsg struct {
	err error
}

func New(cfg app.Config) Model {
	dir, err := cfg.ResolveStateDir()
	if err != nil {
		return Model{status: fmt.Sprintf("state dir: %v", err)}
	}
	m := newModel(dir)
	return m.reload()
}

func NewWithStateDir(stateDir string) Model {
	m := newModel(stateDir)
	return m.reload()
}

func newModel(stateDir string) Model {
	input := textinput.New()
	input.Placeholder = "https://fbref.com/..."
	input.Width = 64
	input.CharLimit = 512
	return Model{
		stateDir: stateDir,
		urlInput: input,
	}
}

func (m Model) reload() Model {
	if m.stateDir == "" {
		m.sources = nil
		return m
	}
	if url, err := ingestion.ReadFbrefURLFile(m.stateDir); err == nil {
		m.fbrefURL = url
		m.urlInput.SetValue(url)
	}
	reg, err := ingestion.LoadOrCreateRegistry(m.stateDir)
	if err != nil {
		m.status = err.Error()
		m.sources = nil
		return m
	}
	m.sources = snapshotsToStatuses(reg.Snapshots())
	m = m.loadStagedForSelection()
	return m
}

func (m Model) loadStagedForSelection() Model {
	if m.stateDir == "" || len(m.sources) == 0 || m.cursor >= len(m.sources) {
		m.staged = nil
		return m
	}
	store, closeFn, err := ingestion.NewSQLiteStagingStore(m.stateDir)
	if err != nil {
		m.staged = nil
		return m
	}
	defer closeFn()
	records, err := store.ListStaged(context.Background(), m.sources[m.cursor].Name, "")
	if err != nil {
		m.staged = nil
		m.preview = nil
		return m
	}
	m.staged = records
	m.preview = buildPreviewLines(records)
	return m
}

func buildPreviewLines(records []ingestion.SourceRecord) []string {
	lines := make([]string, 0, min(len(records), 5))
	for _, record := range records[:min(len(records), 5)] {
		switch record.SourceName {
		case "fbref":
			norm, err := ingestion.NormalizeFbrefPlayerRecord(record)
			if err != nil {
				lines = append(lines, fmt.Sprintf("  ! %s (%s)", truncate(record.ExternalID, 18), truncate(err.Error(), 42)))
				continue
			}
			validation := ingestion.ValidatePlayerRecord(norm.Attributes)
			status := "ok"
			if !validation.Valid {
				status = "invalid"
			}
			lines = append(lines, fmt.Sprintf(
				"  %s  %-18s %s %s",
				status,
				truncate(norm.Name, 18),
				truncate(norm.Attributes["position"], 4),
				truncate(norm.Attributes["nationality"], 12),
			))
		default:
			lines = append(lines, fmt.Sprintf(
				"  %-8s %-18s %s",
				record.EntityType,
				truncate(record.ExternalID, 18),
				truncate(record.RawJSON, 36),
			))
		}
	}
	return lines
}

func snapshotsToStatuses(snaps []ingestion.SourceSnapshot) []SourceStatus {
	out := make([]SourceStatus, 0, len(snaps))
	for _, s := range snaps {
		out = append(out, SourceStatus{
			Name:      s.Name,
			LastFetch: s.LastRunAt,
			Records:   s.Records,
			Errors:    s.Errors,
			Published: s.Published,
			CapPerRun: s.CapPerRun,
			LastError: s.LastError,
		})
	}
	return out
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fetchFinishedMsg:
		next := m.reload()
		if msg.err != nil {
			next.status = msg.err.Error()
		} else {
			next.status = "ingest run finished"
		}
		return next, nil
	case tea.KeyMsg:
		if m.editingURL {
			switch msg.String() {
			case "esc":
				m.editingURL = false
				m.urlInput.Blur()
				m.urlInput.SetValue(m.fbrefURL)
				m.status = "fbref URL edit cancelled"
				return m, nil
			case "enter":
				value := strings.TrimSpace(m.urlInput.Value())
				if value == "" {
					m.status = "fbref URL cannot be empty"
					return m, nil
				}
				if m.stateDir != "" {
					if err := ingestion.WriteFbrefURLFile(m.stateDir, value); err != nil {
						m.status = err.Error()
						return m, nil
					}
				}
				m.fbrefURL = value
				m.editingURL = false
				m.urlInput.Blur()
				m.status = "fbref URL saved"
				return m, nil
			}
			var cmd tea.Cmd
			m.urlInput, cmd = m.urlInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "r":
			return m.reload(), nil
		case "f":
			if m.stateDir == "" {
				m.status = "no state directory"
				return m, nil
			}
			m.status = "fetch running…"
			return m, runDemoFetchCmd(m.stateDir)
		case "s":
			if m.stateDir == "" {
				m.status = "no state directory"
				return m, nil
			}
			u := strings.TrimSpace(m.fbrefURL)
			if u == "" {
				m.status = "press u to enter an FBref URL before scraping"
				return m, nil
			}
			m.status = "fbref scrape…"
			return m, runFbrefScrapeCmd(m.stateDir, u)
		case "u":
			m.editingURL = true
			m.urlInput.Focus()
			m.urlInput.SetValue(m.fbrefURL)
			m.status = "editing fbref URL"
			return m, nil
		case "j", "down":
			if m.cursor < len(m.sources)-1 {
				m.cursor++
				m = m.loadStagedForSelection()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m = m.loadStagedForSelection()
			}
		}
	}
	return m, nil
}

func (m Model) View(width, height int) string {
	if m.status != "" && len(m.sources) == 0 && m.stateDir == "" {
		return lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Data / Import"),
			"",
			errStyle.Render("  "+m.status),
			"",
			helpBarStyle.Render("  fix state directory configuration"),
		)
	}

	header := labelStyle.Render(fmt.Sprintf("  %-16s %-14s %8s %8s %10s",
		"Source", "Last fetch", "Records", "Errors", "Published"))

	lines := []string{
		titleStyle.Render("Data / Import"),
		"",
		header,
	}

	if len(m.sources) == 0 {
		lines = append(lines, dimStyle.Render("  (no sources)"))
	} else {
		for i, s := range m.sources {
			ageStr := formatAge(s.LastFetch)
			errCount := ""
			if s.Errors > 0 {
				errCount = errStyle.Render(fmt.Sprintf("%8d", s.Errors))
			} else {
				errCount = okStyle.Render(fmt.Sprintf("%8d", s.Errors))
			}
			row := fmt.Sprintf("  %-16s %-14s %8d %s %10d",
				s.Name, ageStr, s.Records, errCount, s.Published)
			var styled string
			if i == m.cursor {
				styled = valueStyle.Background(lipgloss.Color("62")).Render(row)
			} else if s.Errors > 10 {
				styled = warnStyle.Render(row)
			} else {
				styled = valueStyle.Render(row)
			}
			lines = append(lines, styled)
		}
	}

	if m.cursor >= 0 && m.cursor < len(m.sources) {
		s := m.sources[m.cursor]
		detail := fmt.Sprintf("  %s  cap/run %d", s.Name, s.CapPerRun)
		if s.LastError != "" {
			detail += "  " + s.LastError
		}
		lines = append(lines, "", dimStyle.Render(detail))
		if len(m.staged) > 0 {
			lines = append(lines, titleStyle.Render("Staged Payloads"))
			limit := min(len(m.staged), 5)
			for _, rec := range m.staged[:limit] {
				lines = append(lines, dimStyle.Render(fmt.Sprintf(
					"  %-8s %-18s %s",
					rec.EntityType,
					truncate(rec.ExternalID, 18),
					truncate(rec.RawJSON, 48),
				)))
			}
			if len(m.staged) > limit {
				lines = append(lines, dimStyle.Render(fmt.Sprintf("  ... %d more staged rows", len(m.staged)-limit)))
			}
			if len(m.preview) > 0 {
				lines = append(lines, titleStyle.Render("Normalized Preview"))
				for _, line := range m.preview {
					lines = append(lines, dimStyle.Render(line))
				}
			}
		}
	}

	if m.status != "" {
		lines = append(lines, "", labelStyle.Render("  "+m.status))
	}

	lines = append(lines, "",
		labelStyle.Render("  FBref URL"),
	)
	urlLine := "  " + m.fbrefURL
	if m.editingURL {
		urlLine = "  " + m.urlInput.View()
	} else if strings.TrimSpace(m.fbrefURL) == "" {
		urlLine = dimStyle.Render("  (press u to enter URL)")
	}
	if m.editingURL {
		lines = append(lines, valueStyle.Render(urlLine), dimStyle.Render("  enter save  esc cancel"))
	} else if strings.TrimSpace(m.fbrefURL) == "" {
		lines = append(lines, urlLine)
	} else {
		lines = append(lines, valueStyle.Render(urlLine))
	}

	lines = append(lines, "",
		dimStyle.Render("  r reload  f demo fetch  u edit fbref URL  s scrape current URL"),
	)

	bodyH := height - 4
	if bodyH < 4 {
		bodyH = 4
	}
	total := len(lines)
	maxScroll := total - bodyH
	if maxScroll < 0 {
		maxScroll = 0
	}
	start := m.scroll
	if start > maxScroll {
		start = maxScroll
	}
	end := start + bodyH
	if end > total {
		end = total
	}

	body := strings.Join(lines[start:end], "\n")
	help := "j/k navigate  r reload  f demo  u edit URL  s scrape"
	if m.editingURL {
		help = "type/paste URL  enter save  esc cancel"
	}
	helpBar := helpBarStyle.Render(help)
	return lipgloss.JoinVertical(lipgloss.Left, body, "", helpBar)
}

func formatAge(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func runDemoFetchCmd(stateDir string) tea.Cmd {
	return func() tea.Msg {
		err := ingestion.RunDemoFetch(context.Background(), stateDir, time.Now())
		return fetchFinishedMsg{err: err}
	}
}

func runFbrefScrapeCmd(stateDir, pageURL string) tea.Cmd {
	return func() tea.Msg {
		err := ingestion.RunFbrefScrape(context.Background(), stateDir, pageURL, time.Now())
		return fetchFinishedMsg{err: err}
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
