// Package ingestion provides the raw data staging pipeline, source adapters,
// normalization, entity resolution, and validation stages.
package ingestion

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ---------------------------------------------------------------------------
// SEC-038: Raw staging pipeline
// ---------------------------------------------------------------------------

// SourceRecord is a raw, unprocessed record from an external data source.
type SourceRecord struct {
	SourceName string
	EntityType string // "player", "club", "match", etc.
	ExternalID string
	RawJSON    string
	FetchedAt  time.Time
}

// StagingStore is the write surface for raw records before normalization.
type StagingStore interface {
	StageRecord(ctx context.Context, record SourceRecord) error
	ListStaged(ctx context.Context, source, entityType string) ([]SourceRecord, error)
	MarkProcessed(ctx context.Context, source, externalID string) error
}

// InMemoryStagingStore is a simple in-process staging store for testing.
type InMemoryStagingStore struct {
	records   []SourceRecord
	processed map[string]bool
}

// NewInMemoryStagingStore returns an empty in-memory staging store.
func NewInMemoryStagingStore() *InMemoryStagingStore {
	return &InMemoryStagingStore{processed: make(map[string]bool)}
}

func (s *InMemoryStagingStore) StageRecord(_ context.Context, r SourceRecord) error {
	if r.SourceName == "" || r.ExternalID == "" {
		return fmt.Errorf("staging: source name and external ID are required")
	}
	s.records = append(s.records, r)
	return nil
}

func (s *InMemoryStagingStore) ListStaged(_ context.Context, source, entityType string) ([]SourceRecord, error) {
	out := make([]SourceRecord, 0)
	for _, r := range s.records {
		if source != "" && r.SourceName != source {
			continue
		}
		if entityType != "" && r.EntityType != entityType {
			continue
		}
		key := r.SourceName + ":" + r.ExternalID
		if s.processed[key] {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *InMemoryStagingStore) MarkProcessed(_ context.Context, source, externalID string) error {
	s.processed[source+":"+externalID] = true
	return nil
}

// ---------------------------------------------------------------------------
// SEC-039: Source adapters
// ---------------------------------------------------------------------------

// SourceAdapter fetches records from one external data source and stages them.
type SourceAdapter interface {
	Name() string
	Fetch(ctx context.Context, store StagingStore) (int, error)
}

// StaticAdapter is a test/demo adapter backed by a pre-loaded record set.
type StaticAdapter struct {
	name    string
	records []SourceRecord
}

// NewStaticAdapter creates an adapter that stages the given records on Fetch.
func NewStaticAdapter(name string, records []SourceRecord) *StaticAdapter {
	return &StaticAdapter{name: name, records: records}
}

func (a *StaticAdapter) Name() string { return a.name }

func (a *StaticAdapter) Fetch(ctx context.Context, store StagingStore) (int, error) {
	staged := 0
	for _, r := range a.records {
		r.SourceName = a.name
		if r.FetchedAt.IsZero() {
			r.FetchedAt = time.Now()
		}
		if err := store.StageRecord(ctx, r); err != nil {
			if errors.Is(err, ErrCapReached) {
				return staged, nil
			}
			return staged, fmt.Errorf("adapter %s: %w", a.name, err)
		}
		staged++
	}
	return staged, nil
}

// ---------------------------------------------------------------------------
// SEC-040: Normalization pipeline
// ---------------------------------------------------------------------------

// NormalizedRecord holds a cleaned, typed entity ready for entity resolution.
type NormalizedRecord struct {
	SourceName string
	ExternalID string
	EntityType string
	Name       string
	Attributes map[string]string
}

// NormalizerFunc transforms a raw staged record into a normalized one.
type NormalizerFunc func(raw SourceRecord) (NormalizedRecord, error)

// Normalize applies a normalizer to a batch of staged records.
func Normalize(records []SourceRecord, fn NormalizerFunc) ([]NormalizedRecord, []error) {
	out := make([]NormalizedRecord, 0, len(records))
	errs := make([]error, 0)
	for _, r := range records {
		norm, err := fn(r)
		if err != nil {
			errs = append(errs, fmt.Errorf("normalize %s/%s: %w", r.SourceName, r.ExternalID, err))
			continue
		}
		out = append(out, norm)
	}
	return out, errs
}

// ---------------------------------------------------------------------------
// SEC-041: Entity resolution
// ---------------------------------------------------------------------------

// EntityMatch holds the result of matching a normalized record to an existing entity.
type EntityMatch struct {
	NormalizedRecord
	ResolvedID int64 // 0 = new entity
	MatchScore float64
	IsNew      bool
}

// EntityResolver attempts to match normalized records to existing entities.
type EntityResolver struct {
	// nameIndex maps canonical names → existing IDs
	nameIndex map[string]int64
	nextID    int64
}

// NewEntityResolver creates a resolver with an initial set of known entities.
func NewEntityResolver(known map[string]int64) *EntityResolver {
	idx := make(map[string]int64, len(known))
	for name, id := range known {
		idx[canonicalize(name)] = id
	}
	var maxID int64
	for _, id := range known {
		if id > maxID {
			maxID = id
		}
	}
	return &EntityResolver{nameIndex: idx, nextID: maxID + 1}
}

// Resolve matches a normalized record to an existing entity or marks it new.
func (r *EntityResolver) Resolve(norm NormalizedRecord) EntityMatch {
	key := canonicalize(norm.Name)
	if id, ok := r.nameIndex[key]; ok {
		return EntityMatch{NormalizedRecord: norm, ResolvedID: id, MatchScore: 1.0}
	}
	// Assign a new provisional ID.
	id := r.nextID
	r.nextID++
	r.nameIndex[key] = id
	return EntityMatch{NormalizedRecord: norm, ResolvedID: id, MatchScore: 0, IsNew: true}
}

func canonicalize(s string) string {
	out := make([]byte, 0, len(s))
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			out = append(out, byte(c+32))
		} else if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			out = append(out, byte(c))
		}
	}
	return string(out)
}

// ---------------------------------------------------------------------------
// SEC-042: Validation and publish pipeline
// ---------------------------------------------------------------------------

// ValidationResult records the outcome of a validation check on an entity.
type ValidationResult struct {
	EntityType string
	ExternalID string
	Valid      bool
	Errors     []string
}

// ValidatePlayerRecord validates the required fields of a player entity.
func ValidatePlayerRecord(attrs map[string]string) ValidationResult {
	result := ValidationResult{EntityType: "player", Valid: true}
	required := []string{"name", "position"}
	for _, field := range required {
		if attrs[field] == "" {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("missing required field: %s", field))
		}
	}
	return result
}

// ValidateClubRecord validates the required fields of a club entity.
func ValidateClubRecord(attrs map[string]string) ValidationResult {
	result := ValidationResult{EntityType: "club", Valid: true}
	if attrs["name"] == "" {
		result.Valid = false
		result.Errors = append(result.Errors, "missing required field: name")
	}
	return result
}

// PublishSummary holds counts from a completed ingest-and-publish run.
type PublishSummary struct {
	Staged      int
	Normalized  int
	Resolved    int
	NewEntities int
	Errors      int
}
