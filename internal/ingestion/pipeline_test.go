package ingestion

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryStagingStore_stageAndList(t *testing.T) {
	store := NewInMemoryStagingStore()
	ctx := context.Background()

	rec := SourceRecord{
		SourceName: "fbref",
		EntityType: "player",
		ExternalID: "fbref-001",
		RawJSON:    `{"name":"Test Player"}`,
		FetchedAt:  time.Now(),
	}
	if err := store.StageRecord(ctx, rec); err != nil {
		t.Fatalf("stage: %v", err)
	}

	records, err := store.ListStaged(ctx, "fbref", "player")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}

func TestInMemoryStagingStore_markProcessed(t *testing.T) {
	store := NewInMemoryStagingStore()
	ctx := context.Background()

	rec := SourceRecord{SourceName: "fbref", EntityType: "player", ExternalID: "p1", RawJSON: "{}"}
	_ = store.StageRecord(ctx, rec)
	_ = store.MarkProcessed(ctx, "fbref", "p1")

	records, _ := store.ListStaged(ctx, "fbref", "player")
	if len(records) != 0 {
		t.Fatalf("expected 0 records after marking processed, got %d", len(records))
	}
}

func TestInMemoryStagingStore_validationError(t *testing.T) {
	store := NewInMemoryStagingStore()
	err := store.StageRecord(context.Background(), SourceRecord{})
	if err == nil {
		t.Fatal("expected error for empty source name/external ID")
	}
}

func TestStaticAdapter_fetchStages(t *testing.T) {
	records := []SourceRecord{
		{EntityType: "player", ExternalID: "p1", RawJSON: `{}`},
		{EntityType: "club", ExternalID: "c1", RawJSON: `{}`},
	}
	adapter := NewStaticAdapter("test", records)
	store := NewInMemoryStagingStore()

	count, err := adapter.Fetch(context.Background(), store)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

func TestNormalize_appliesFunc(t *testing.T) {
	records := []SourceRecord{
		{SourceName: "test", ExternalID: "p1", EntityType: "player", RawJSON: `{"name":"Alice"}`},
	}
	norm := func(r SourceRecord) (NormalizedRecord, error) {
		return NormalizedRecord{
			SourceName: r.SourceName,
			ExternalID: r.ExternalID,
			EntityType: r.EntityType,
			Name:       "Alice",
		}, nil
	}
	out, errs := Normalize(records, norm)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(out) != 1 || out[0].Name != "Alice" {
		t.Fatalf("normalize output wrong: %+v", out)
	}
}

func TestEntityResolver_existingMatch(t *testing.T) {
	resolver := NewEntityResolver(map[string]int64{"Arsenal FC": 1, "Liverpool FC": 2})
	match := resolver.Resolve(NormalizedRecord{Name: "Arsenal FC"})
	if match.ResolvedID != 1 {
		t.Fatalf("resolved ID = %d, want 1", match.ResolvedID)
	}
	if match.IsNew {
		t.Fatal("should not be marked new")
	}
}

func TestEntityResolver_newEntity(t *testing.T) {
	resolver := NewEntityResolver(map[string]int64{})
	match := resolver.Resolve(NormalizedRecord{Name: "Unknown Club FC"})
	if !match.IsNew {
		t.Fatal("expected new entity flag")
	}
	if match.ResolvedID == 0 {
		t.Fatal("expected a provisional ID")
	}
}

func TestValidatePlayerRecord_missingField(t *testing.T) {
	result := ValidatePlayerRecord(map[string]string{"name": "Alice"})
	if result.Valid {
		t.Fatal("expected invalid when position missing")
	}
}

func TestValidatePlayerRecord_valid(t *testing.T) {
	result := ValidatePlayerRecord(map[string]string{"name": "Alice", "position": "ST"})
	if !result.Valid {
		t.Fatalf("expected valid, errors: %v", result.Errors)
	}
}

func TestValidateClubRecord_missingName(t *testing.T) {
	result := ValidateClubRecord(map[string]string{})
	if result.Valid {
		t.Fatal("expected invalid when name missing")
	}
}
