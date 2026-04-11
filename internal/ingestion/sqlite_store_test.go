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
