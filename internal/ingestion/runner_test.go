package ingestion

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestCappedStagingStore_respectsCap(t *testing.T) {
	inner := NewInMemoryStagingStore()
	recs := make([]SourceRecord, 5)
	for i := range recs {
		recs[i] = SourceRecord{EntityType: "player", ExternalID: fmt.Sprintf("p%d", i), RawJSON: "{}"}
	}
	ad := NewStaticAdapter("captest", recs)
	wrapped := NewCappedStagingStore(inner, 2)
	ctx := context.Background()
	n, err := ad.Fetch(ctx, wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("n=%d", n)
	}
	all, err := inner.ListStaged(ctx, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("len=%d", len(all))
	}
}

func TestRunDemoFetch_updatesRegistry(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	now := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	if err := RunDemoFetch(ctx, dir, now); err != nil {
		t.Fatal(err)
	}
	reg, err := LoadOrCreateRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range reg.Snapshots() {
		if s.Name == "fbref" {
			if s.Records != 3 {
				t.Fatalf("fbref records %d", s.Records)
			}
			if !s.LastRunAt.Equal(now) {
				t.Fatalf("time %v", s.LastRunAt)
			}
			return
		}
	}
	t.Fatal("fbref not found")
}

func TestRunDemoFetch_persistsRawPayloadsInSQLite(t *testing.T) {
	dir := t.TempDir()
	if err := RunDemoFetch(context.Background(), dir, time.Now()); err != nil {
		t.Fatalf("RunDemoFetch() error = %v", err)
	}
	store, closeFn, err := NewSQLiteStagingStore(dir)
	if err != nil {
		t.Fatalf("NewSQLiteStagingStore() error = %v", err)
	}
	defer closeFn()

	records, err := store.ListStaged(context.Background(), "fbref", "")
	if err != nil {
		t.Fatalf("ListStaged() error = %v", err)
	}
	if len(records) == 0 {
		t.Fatal("expected sqlite-backed staged payloads after demo fetch")
	}
}
