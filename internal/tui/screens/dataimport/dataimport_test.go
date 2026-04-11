package dataimport

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ismael/football-analytics/internal/ingestion"
)

func TestNewWithStateDir_hasSources(t *testing.T) {
	m := NewWithStateDir(t.TempDir())
	if len(m.sources) == 0 {
		t.Fatal("must have at least one source")
	}
}

func TestUpdate_cursorNavigation(t *testing.T) {
	m := NewWithStateDir(t.TempDir())
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
}

func TestView_containsSourceNames(t *testing.T) {
	m := NewWithStateDir(t.TempDir())
	view := m.View(120, 40)
	for _, fragment := range []string{"Data / Import", "fbref", "transfermarkt"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("view missing %q", fragment)
		}
	}
}

func TestView_neverBeforeFetch(t *testing.T) {
	m := NewWithStateDir(t.TempDir())
	view := m.View(120, 40)
	if !strings.Contains(view, "never") {
		t.Fatal("expected never for zero last fetch")
	}
}

func TestUpdate_fetchFinished_refreshesRows(t *testing.T) {
	dir := t.TempDir()
	m := NewWithStateDir(dir)
	if err := ingestion.RunDemoFetch(context.Background(), dir, time.Now()); err != nil {
		t.Fatal(err)
	}
	m2, _ := m.Update(fetchFinishedMsg{err: nil})
	found := false
	for _, s := range m2.sources {
		if s.Name == "fbref" && s.Records > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected fbref rows after fetch")
	}
}

func TestUpdate_saveFbrefURLFromEditor(t *testing.T) {
	dir := t.TempDir()
	m := NewWithStateDir(dir)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	if !m.editingURL {
		t.Fatal("editingURL should be true after u")
	}
	for _, r := range "https://example.com/fbref" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.editingURL {
		t.Fatal("editingURL should be false after enter")
	}
	if !strings.Contains(m.fbrefURL, "example.com/fbref") {
		t.Fatalf("fbrefURL = %q", m.fbrefURL)
	}
	saved, err := ingestion.ReadFbrefURLFile(dir)
	if err != nil {
		t.Fatalf("ReadFbrefURLFile() error = %v", err)
	}
	if saved != m.fbrefURL {
		t.Fatalf("saved URL = %q, want %q", saved, m.fbrefURL)
	}
}

func TestView_showsStagedPayloadsForSelectedSource(t *testing.T) {
	dir := t.TempDir()
	if err := ingestion.RunDemoFetch(context.Background(), dir, time.Now()); err != nil {
		t.Fatalf("RunDemoFetch() error = %v", err)
	}
	m := NewWithStateDir(dir)
	view := m.View(120, 40)
	if !strings.Contains(view, "Staged Payloads") {
		t.Fatal("expected staged payload section in view")
	}
}

func TestView_showsNormalizedPreviewForFbref(t *testing.T) {
	dir := t.TempDir()
	store, closeFn, err := ingestion.NewSQLiteStagingStore(dir)
	if err != nil {
		t.Fatalf("NewSQLiteStagingStore() error = %v", err)
	}
	defer closeFn()
	if err := store.StageRecord(context.Background(), ingestion.SourceRecord{
		SourceName: "fbref",
		EntityType: "player",
		ExternalID: "fb-1",
		RawJSON:    `{"player":"Ada Demo","pos":"FW","nation":"eng England"}`,
		FetchedAt:  time.Now(),
	}); err != nil {
		t.Fatalf("StageRecord() error = %v", err)
	}
	reg, err := ingestion.LoadOrCreateRegistry(dir)
	if err != nil {
		t.Fatalf("LoadOrCreateRegistry() error = %v", err)
	}
	reg.RecordRun("fbref", time.Now(), 1, 0, 0, nil)
	if err := reg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	m := NewWithStateDir(dir)
	view := m.View(120, 40)
	if !strings.Contains(view, "Normalized Preview") {
		t.Fatal("expected normalized preview section")
	}
	if !strings.Contains(view, "Ada Demo") {
		t.Fatal("expected normalized player name in preview")
	}
	if !strings.Contains(view, "Publish Preview") {
		t.Fatal("expected publish preview section")
	}
}

func TestUpdate_pPublishesSelectedSource(t *testing.T) {
	dir := t.TempDir()
	store, closeFn, err := ingestion.NewSQLiteStagingStore(dir)
	if err != nil {
		t.Fatalf("NewSQLiteStagingStore() error = %v", err)
	}
	defer closeFn()
	if err := store.StageRecord(context.Background(), ingestion.SourceRecord{
		SourceName: "fbref",
		EntityType: "player",
		ExternalID: "fb-1",
		RawJSON:    `{"player":"Ada Demo","pos":"FW","nation":"eng England"}`,
		FetchedAt:  time.Now(),
	}); err != nil {
		t.Fatalf("StageRecord() error = %v", err)
	}
	reg, err := ingestion.LoadOrCreateRegistry(dir)
	if err != nil {
		t.Fatalf("LoadOrCreateRegistry() error = %v", err)
	}
	reg.RecordRun("fbref", time.Now(), 1, 0, 0, nil)
	if err := reg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	m := NewWithStateDir(dir)
	msg := runPublishCmd(dir, "fbref")()
	m, _ = m.Update(msg)
	if !strings.Contains(m.status, "published 1 entities") {
		t.Fatalf("status = %q", m.status)
	}

	published, err := store.ListPublishedBySource(context.Background(), "fbref")
	if err != nil {
		t.Fatalf("ListPublishedBySource() error = %v", err)
	}
	if len(published) != 1 {
		t.Fatalf("published len = %d, want 1", len(published))
	}

	view := m.View(120, 40)
	if !strings.Contains(view, "Published Entities") {
		t.Fatal("expected published entities section")
	}
	if !strings.Contains(view, "Published Detail") {
		t.Fatal("expected published detail section")
	}
}

func TestUpdate_editPublishedEntityName(t *testing.T) {
	dir := t.TempDir()
	store, closeFn, err := ingestion.NewSQLiteStagingStore(dir)
	if err != nil {
		t.Fatalf("NewSQLiteStagingStore() error = %v", err)
	}
	defer closeFn()
	if err := store.SavePublished(context.Background(), ingestion.PublishedEntity{
		SourceCode: "fbref",
		ExternalID: "fb-1",
		EntityType: "player",
		ResolvedID: 7,
		Name:       "Ada Demo",
		Attributes: map[string]string{
			"name":        "Ada Demo",
			"position":    "FW",
			"nationality": "England",
		},
		Validation: ingestion.ValidatePlayerRecord(map[string]string{
			"name":        "Ada Demo",
			"position":    "FW",
			"nationality": "England",
		}),
		PublishedAt: time.Now(),
	}); err != nil {
		t.Fatalf("SavePublished() error = %v", err)
	}
	reg, err := ingestion.LoadOrCreateRegistry(dir)
	if err != nil {
		t.Fatalf("LoadOrCreateRegistry() error = %v", err)
	}
	reg.RecordRun("fbref", time.Now(), 0, 0, 1, nil)
	if err := reg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	m := NewWithStateDir(dir)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !m.focusPublished {
		t.Fatal("expected published focus after tab")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if !m.editingPublished {
		t.Fatal("expected published edit mode after e")
	}
	m.entityInput.SetValue("Ada Revised")
	msg := runUpdatePublishedCmd(dir, ingestion.PublishedEntity{
		ID:         m.persisted[m.publishedCursor].ID,
		SourceCode: m.persisted[m.publishedCursor].SourceCode,
		ExternalID: m.persisted[m.publishedCursor].ExternalID,
		EntityType: m.persisted[m.publishedCursor].EntityType,
		ResolvedID: m.persisted[m.publishedCursor].ResolvedID,
		Name:       "Ada Revised",
		Attributes: map[string]string{
			"name":        "Ada Revised",
			"position":    "FW",
			"nationality": "England",
		},
		Validation: ingestion.ValidatePlayerRecord(map[string]string{
			"name":        "Ada Revised",
			"position":    "FW",
			"nationality": "England",
		}),
	}, "name")()
	m, _ = m.Update(msg)
	if !strings.Contains(m.status, "published entity updated") {
		t.Fatalf("status = %q", m.status)
	}

	published, err := store.ListPublishedBySource(context.Background(), "fbref")
	if err != nil {
		t.Fatalf("ListPublishedBySource() error = %v", err)
	}
	if len(published) != 1 {
		t.Fatalf("published len = %d, want 1", len(published))
	}
	if published[0].Name != "Ada Revised" {
		t.Fatalf("published name = %q, want Ada Revised", published[0].Name)
	}

	view := m.View(120, 40)
	if !strings.Contains(view, "Ada Revised") {
		t.Fatal("expected updated published name in view")
	}
}
