package ingestion

import (
	"context"
	"testing"
	"time"
)

func TestSQLiteStagingStoreStageListAndMarkProcessed(t *testing.T) {
	store, closeFn, err := NewSQLiteStagingStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStagingStore() error = %v", err)
	}
	defer closeFn()

	record := SourceRecord{
		SourceName: "fbref",
		EntityType: "player",
		ExternalID: "p1",
		RawJSON:    `{"name":"Demo"}`,
		FetchedAt:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := store.StageRecord(context.Background(), record); err != nil {
		t.Fatalf("StageRecord() error = %v", err)
	}

	records, err := store.ListStaged(context.Background(), "fbref", "player")
	if err != nil {
		t.Fatalf("ListStaged() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("ListStaged() len = %d, want 1", len(records))
	}

	if err := store.MarkProcessed(context.Background(), "fbref", "p1"); err != nil {
		t.Fatalf("MarkProcessed() error = %v", err)
	}
	records, err = store.ListStaged(context.Background(), "fbref", "player")
	if err != nil {
		t.Fatalf("ListStaged() after mark error = %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("ListStaged() after mark len = %d, want 0", len(records))
	}
}

func TestSQLiteStagingStoreUpdatePublished(t *testing.T) {
	store, closeFn, err := NewSQLiteStagingStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewSQLiteStagingStore() error = %v", err)
	}
	defer closeFn()

	entity := PublishedEntity{
		SourceCode: "fbref",
		ExternalID: "p1",
		EntityType: "player",
		ResolvedID: 7,
		Name:       "Ada Demo",
		Attributes: map[string]string{
			"name":        "Ada Demo",
			"position":    "FW",
			"nationality": "England",
		},
		Validation: ValidatePlayerRecord(map[string]string{
			"name":        "Ada Demo",
			"position":    "FW",
			"nationality": "England",
		}),
		PublishedAt: time.Now(),
	}
	if err := store.SavePublished(context.Background(), entity); err != nil {
		t.Fatalf("SavePublished() error = %v", err)
	}

	published, err := store.ListPublishedBySource(context.Background(), "fbref")
	if err != nil {
		t.Fatalf("ListPublishedBySource() error = %v", err)
	}
	if len(published) != 1 {
		t.Fatalf("ListPublishedBySource() len = %d, want 1", len(published))
	}

	published[0].Name = "Ada Revised"
	published[0].ResolvedID = 11
	published[0].Attributes["name"] = "Ada Revised"
	if err := store.UpdatePublished(context.Background(), published[0]); err != nil {
		t.Fatalf("UpdatePublished() error = %v", err)
	}

	updated, err := store.ListPublishedBySource(context.Background(), "fbref")
	if err != nil {
		t.Fatalf("ListPublishedBySource() updated error = %v", err)
	}
	if updated[0].Name != "Ada Revised" {
		t.Fatalf("updated name = %q, want Ada Revised", updated[0].Name)
	}
	if updated[0].ResolvedID != 11 {
		t.Fatalf("updated resolved id = %d, want 11", updated[0].ResolvedID)
	}
	if updated[0].Attributes["name"] != "Ada Revised" {
		t.Fatalf("updated attribute name = %q, want Ada Revised", updated[0].Attributes["name"])
	}
}
