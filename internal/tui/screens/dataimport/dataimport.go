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
	stateDir         string
	sources          []SourceStatus
	staged           []ingestion.SourceRecord
	preview          []string
	published        []string
	persisted        []ingestion.PublishedEntity
	cursor           int
	publishedCursor  int
	scroll           int
	status           string
	fbrefURL         string
	urlInput         textinput.Model
	entityInput      textinput.Model
	editingURL       bool
	editingPublished bool
	publishedField   string
	focusPublished   bool
}

type fetchFinishedMsg struct {
	err error
}

type publishFinishedMsg struct {
	count int
	err   error
}

type publishedSavedMsg struct {
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
	entityInput := textinput.New()
	entityInput.Width = 32
	entityInput.CharLimit = 128
	return Model{
		stateDir:    stateDir,
		urlInput:    input,
		entityInput: entityInput,
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
		m.persisted = nil
		m.publishedCursor = 0
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
	m.published = buildPublishedLines(records)
	publishedEntities, err := store.ListPublishedBySource(context.Background(), m.sources[m.cursor].Name)
	if err != nil {
		m.persisted = nil
		m.publishedCursor = 0
		return m
	}
	m.persisted = publishedEntities
	if len(m.persisted) == 0 {
		m.publishedCursor = 0
	} else if m.publishedCursor >= len(m.persisted) {
		m.publishedCursor = len(m.persisted) - 1
	}
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

func buildPublishedLines(records []ingestion.SourceRecord) []string {
	known := map[string]int64{
		"Arsenal FC":      1,
		"Liverpool FC":    2,
		"Manchester City": 3,
	}
	resolver := ingestion.NewEntityResolver(known)
	lines := make([]string, 0, min(len(records), 5))
	for _, record := range records[:min(len(records), 5)] {
		norm, validation, ok := normalizeForPreview(record)
		if !ok {
			continue
		}
		match := resolver.Resolve(norm)
		state := "valid"
		if !validation.Valid {
			state = "invalid"
		} else if match.IsNew {
			state = "new"
		}
		lines = append(lines, fmt.Sprintf(
			"  %-7s id:%-3d %-18s %s",
			state,
			match.ResolvedID,
			truncate(norm.Name, 18),
			truncate(norm.EntityType, 8),
		))
	}
	return lines
}

func buildPersistedLines(entities []ingestion.PublishedEntity, cursor int, focused bool) []string {
	lines := make([]string, 0, min(len(entities), 5))
	for i, entity := range entities[:min(len(entities), 5)] {
		state := "valid"
		if !entity.Validation.Valid {
			state = "invalid"
		}
		prefix := "  "
		if i == cursor {
			if focused {
				prefix = "> "
			} else {
				prefix = "* "
			}
		}
		lines = append(lines, fmt.Sprintf(
			"%s%-7s id:%-3d %-18s %s",
			prefix,
			state,
			entity.ResolvedID,
			truncate(entity.Name, 18),
			truncate(entity.EntityType, 8),
		))
	}
	return lines
}

func buildPersistedDetail(entity ingestion.PublishedEntity) []string {
	lines := []string{
		fmt.Sprintf("  id %d  resolved %d", entity.ID, entity.ResolvedID),
		fmt.Sprintf("  source %s  type %s", entity.SourceCode, entity.EntityType),
		fmt.Sprintf("  external %s", truncate(entity.ExternalID, 42)),
		fmt.Sprintf("  name %s", truncate(entity.Name, 42)),
	}
	if len(entity.Attributes) > 0 {
		parts := make([]string, 0, len(entity.Attributes))
		for _, key := range []string{"position", "nationality", "name"} {
			if value := strings.TrimSpace(entity.Attributes[key]); value != "" {
				parts = append(parts, fmt.Sprintf("%s=%s", key, truncate(value, 16)))
			}
		}
		if len(parts) > 0 {
			lines = append(lines, "  attrs "+strings.Join(parts, "  "))
		}
	}
	if entity.Validation.Valid {
		lines = append(lines, "  validation valid")
	} else if len(entity.Validation.Errors) > 0 {
		lines = append(lines, "  validation "+truncate(strings.Join(entity.Validation.Errors, "; "), 52))
	} else {
		lines = append(lines, "  validation invalid")
	}
	return lines
}

func normalizeForPreview(record ingestion.SourceRecord) (ingestion.NormalizedRecord, ingestion.ValidationResult, bool) {
	switch record.SourceName {
	case "fbref":
		norm, err := ingestion.NormalizeFbrefPlayerRecord(record)
		if err != nil {
			return ingestion.NormalizedRecord{}, ingestion.ValidationResult{}, false
		}
		return norm, ingestion.ValidatePlayerRecord(norm.Attributes), true
	default:
		return ingestion.NormalizedRecord{
				SourceName: record.SourceName,
				ExternalID: record.ExternalID,
				EntityType: record.EntityType,
				Name:       record.ExternalID,
				Attributes: map[string]string{"name": record.ExternalID},
			},
			ingestion.ValidationResult{EntityType: record.EntityType, ExternalID: record.ExternalID, Valid: true},
			true
	}
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
	case publishFinishedMsg:
		next := m.reload()
		if msg.err != nil {
			next.status = msg.err.Error()
		} else {
			next.status = fmt.Sprintf("published %d entities", msg.count)
		}
		return next, nil
	case publishedSavedMsg:
		next := m.reload()
		if msg.err != nil {
			next.status = msg.err.Error()
		} else {
			next.status = "published entity updated"
			next.focusPublished = true
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
		if m.editingPublished {
			switch msg.String() {
			case "esc":
				m.editingPublished = false
				m.publishedField = ""
				m.entityInput.Blur()
				m.status = "published edit cancelled"
				return m, nil
			case "enter":
				entity := m.selectedPublishedEntity()
				if entity == nil {
					m.editingPublished = false
					m.publishedField = ""
					m.entityInput.Blur()
					return m, nil
				}
				value := strings.TrimSpace(m.entityInput.Value())
				switch m.publishedField {
				case "name":
					if value == "" {
						m.status = "published name cannot be empty"
						return m, nil
					}
					entity.Name = value
					if entity.Attributes == nil {
						entity.Attributes = map[string]string{}
					}
					entity.Attributes["name"] = value
				case "resolved_id":
					var id int64
					if _, err := fmt.Sscanf(value, "%d", &id); err != nil {
						m.status = "resolved id must be numeric"
						return m, nil
					}
					entity.ResolvedID = id
				}
				m.editingPublished = false
				field := m.publishedField
				m.publishedField = ""
				m.entityInput.Blur()
				m.status = "saving published entity…"
				return m, runUpdatePublishedCmd(m.stateDir, *entity, field)
			}
			var cmd tea.Cmd
			m.entityInput, cmd = m.entityInput.Update(msg)
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
		case "p":
			if m.stateDir == "" || len(m.sources) == 0 || m.cursor >= len(m.sources) {
				m.status = "no source selected"
				return m, nil
			}
			m.status = "publishing…"
			return m, runPublishCmd(m.stateDir, m.sources[m.cursor].Name)
		case "u":
			m.editingURL = true
			m.urlInput.Focus()
			m.urlInput.SetValue(m.fbrefURL)
			m.status = "editing fbref URL"
			return m, nil
		case "tab":
			if len(m.persisted) == 0 {
				return m, nil
			}
			m.focusPublished = !m.focusPublished
			if m.focusPublished {
				m.status = "published entity focus"
			} else {
				m.status = "source list focus"
			}
			return m, nil
		case "e":
			if !m.focusPublished || len(m.persisted) == 0 {
				return m, nil
			}
			entity := m.selectedPublishedEntity()
			if entity == nil {
				return m, nil
			}
			m.editingPublished = true
			m.publishedField = "name"
			m.entityInput.SetValue(entity.Name)
			m.entityInput.Focus()
			m.status = "editing published name"
			return m, nil
		case "i":
			if !m.focusPublished || len(m.persisted) == 0 {
				return m, nil
			}
			entity := m.selectedPublishedEntity()
			if entity == nil {
				return m, nil
			}
			m.editingPublished = true
			m.publishedField = "resolved_id"
			m.entityInput.SetValue(fmt.Sprintf("%d", entity.ResolvedID))
			m.entityInput.Focus()
			m.status = "editing resolved id"
			return m, nil
		case "j", "down":
			if m.focusPublished {
				if m.publishedCursor < len(m.persisted)-1 {
					m.publishedCursor++
				}
				return m, nil
			}
			if m.cursor < len(m.sources)-1 {
				m.cursor++
				m = m.loadStagedForSelection()
			}
		case "k", "up":
			if m.focusPublished {
				if m.publishedCursor > 0 {
					m.publishedCursor--
				}
				return m, nil
			}
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
			if len(m.published) > 0 {
				lines = append(lines, titleStyle.Render("Publish Preview"))
				for _, line := range m.published {
					lines = append(lines, dimStyle.Render(line))
				}
			}
		}
		if len(m.persisted) > 0 {
			lines = append(lines, titleStyle.Render("Published Entities"))
			for _, line := range buildPersistedLines(m.persisted, m.publishedCursor, m.focusPublished) {
				lines = append(lines, dimStyle.Render(line))
			}
			if selected := m.selectedPublishedEntity(); selected != nil {
				lines = append(lines, titleStyle.Render("Published Detail"))
				for _, line := range buildPersistedDetail(*selected) {
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

	if m.editingPublished {
		lines = append(lines, "",
			labelStyle.Render("  Published Edit"),
			valueStyle.Render("  "+m.entityInput.View()),
			dimStyle.Render("  enter save  esc cancel"),
		)
	}

	lines = append(lines, "",
		dimStyle.Render("  r reload  f demo fetch  u edit fbref URL  s scrape current URL  p publish source  tab published focus  e edit name  i edit resolved id"),
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
	help := "j/k navigate  tab toggle focus  r reload  f demo  u edit URL  s scrape  p publish  e name  i id"
	if m.editingURL {
		help = "type/paste URL  enter save  esc cancel"
	} else if m.editingPublished {
		help = "type value  enter save  esc cancel"
	}
	helpBar := helpBarStyle.Render(help)
	return lipgloss.JoinVertical(lipgloss.Left, body, "", helpBar)
}

func (m Model) selectedPublishedEntity() *ingestion.PublishedEntity {
	if m.publishedCursor < 0 || m.publishedCursor >= len(m.persisted) {
		return nil
	}
	entity := m.persisted[m.publishedCursor]
	return &entity
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

func runUpdatePublishedCmd(stateDir string, entity ingestion.PublishedEntity, field string) tea.Cmd {
	return func() tea.Msg {
		store, closeFn, err := ingestion.NewSQLiteStagingStore(stateDir)
		if err != nil {
			return publishedSavedMsg{err: err}
		}
		defer closeFn()
		if field == "name" && entity.Validation.EntityType == "player" {
			entity.Validation = ingestion.ValidatePlayerRecord(entity.Attributes)
		}
		err = store.UpdatePublished(context.Background(), entity)
		return publishedSavedMsg{err: err}
	}
}

func runFbrefScrapeCmd(stateDir, pageURL string) tea.Cmd {
	return func() tea.Msg {
		err := ingestion.RunFbrefScrape(context.Background(), stateDir, pageURL, time.Now())
		return fetchFinishedMsg{err: err}
	}
}

func runPublishCmd(stateDir, source string) tea.Cmd {
	return func() tea.Msg {
		store, closeFn, err := ingestion.NewSQLiteStagingStore(stateDir)
		if err != nil {
			return publishFinishedMsg{err: err}
		}
		defer closeFn()
		count, err := ingestion.PublishSourceRecords(context.Background(), store, source)
		return publishFinishedMsg{count: count, err: err}
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
