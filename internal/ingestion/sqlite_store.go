package ingestion

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	storagesqlite "github.com/ismael/football-analytics/internal/storage/sqlite"
	_ "modernc.org/sqlite"
)

// SQLiteStagingStore persists staged ingest records in the app SQLite database.
type SQLiteStagingStore struct {
	db *sql.DB
}

type PublishedEntity struct {
	SourceCode  string
	ExternalID  string
	EntityType  string
	ResolvedID  int64
	Name        string
	Attributes  map[string]string
	Validation  ValidationResult
	PublishedAt time.Time
}

func NewSQLiteStagingStore(stateDir string) (*SQLiteStagingStore, func() error, error) {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create state dir: %w", err)
	}
	db, err := sql.Open("sqlite", storagesqlite.DefaultPath(stateDir))
	if err != nil {
		return nil, nil, fmt.Errorf("open sqlite staging store: %w", err)
	}
	if err := storagesqlite.NewStore(db).Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("migrate sqlite staging store: %w", err)
	}
	return &SQLiteStagingStore{db: db}, db.Close, nil
}

func (s *SQLiteStagingStore) StageRecord(ctx context.Context, record SourceRecord) error {
	if record.SourceName == "" || record.ExternalID == "" {
		return fmt.Errorf("staging: source name and external ID are required")
	}
	sourceID, err := s.ensureSource(ctx, record.SourceName)
	if err != nil {
		return err
	}
	fetchedAt := record.FetchedAt
	if fetchedAt.IsZero() {
		fetchedAt = time.Now().UTC()
	}
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO raw_payloads(source_id, external_id, fetched_at, payload_json, checksum, processed_at)
		VALUES (?, ?, ?, ?, ?, NULL)`,
		sourceID,
		record.ExternalID,
		fetchedAt.Format(time.RFC3339Nano),
		record.RawJSON,
		record.EntityType,
	)
	if err != nil {
		return fmt.Errorf("insert raw payload: %w", err)
	}
	return nil
}

func (s *SQLiteStagingStore) ListStaged(ctx context.Context, source, entityType string) ([]SourceRecord, error) {
	query := `SELECT rs.code, COALESCE(rp.checksum, ''), rp.external_id, rp.payload_json, rp.fetched_at
		FROM raw_payloads rp
		INNER JOIN raw_sources rs ON rs.id = rp.source_id
		WHERE rp.processed_at IS NULL`
	args := make([]any, 0, 2)
	if source != "" {
		query += ` AND rs.code = ?`
		args = append(args, source)
	}
	if entityType != "" {
		query += ` AND COALESCE(rp.checksum, '') = ?`
		args = append(args, entityType)
	}
	query += ` ORDER BY rp.id`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query staged payloads: %w", err)
	}
	defer rows.Close()

	var out []SourceRecord
	for rows.Next() {
		var record SourceRecord
		var fetchedAt string
		if err := rows.Scan(&record.SourceName, &record.EntityType, &record.ExternalID, &record.RawJSON, &fetchedAt); err != nil {
			return nil, fmt.Errorf("scan staged payload: %w", err)
		}
		if fetchedAt != "" {
			if parsed, err := time.Parse(time.RFC3339Nano, fetchedAt); err == nil {
				record.FetchedAt = parsed
			}
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate staged payloads: %w", err)
	}
	return out, nil
}

func (s *SQLiteStagingStore) MarkProcessed(ctx context.Context, source, externalID string) error {
	_, err := s.db.ExecContext(
		ctx,
		`UPDATE raw_payloads
		SET processed_at = ?
		WHERE external_id = ? AND source_id = (SELECT id FROM raw_sources WHERE code = ? LIMIT 1) AND processed_at IS NULL`,
		time.Now().UTC().Format(time.RFC3339Nano),
		externalID,
		source,
	)
	if err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}
	return nil
}

func (s *SQLiteStagingStore) ensureSource(ctx context.Context, sourceName string) (int64, error) {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO raw_sources(code, description) VALUES (?, ?) ON CONFLICT(code) DO NOTHING`,
		sourceName,
		sourceName+" staging source",
	)
	if err != nil {
		return 0, fmt.Errorf("ensure source: %w", err)
	}
	var sourceID int64
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM raw_sources WHERE code = ?`, sourceName).Scan(&sourceID); err != nil {
		return 0, fmt.Errorf("lookup source id: %w", err)
	}
	return sourceID, nil
}

func resolveSQLiteStateStore(stateDir string) (*SQLiteStagingStore, func() error, error) {
	store, closeFn, err := NewSQLiteStagingStore(stateDir)
	if err != nil {
		return nil, nil, err
	}
	return store, closeFn, nil
}

func (s *SQLiteStagingStore) SavePublished(ctx context.Context, entity PublishedEntity) error {
	attrsJSON, err := json.Marshal(entity.Attributes)
	if err != nil {
		return fmt.Errorf("marshal attributes: %w", err)
	}
	validationJSON, err := json.Marshal(entity.Validation)
	if err != nil {
		return fmt.Errorf("marshal validation: %w", err)
	}
	publishedAt := entity.PublishedAt
	if publishedAt.IsZero() {
		publishedAt = time.Now().UTC()
	}
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO published_entities(source_code, external_id, entity_type, resolved_id, name, attributes_json, validation_json, published_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		entity.SourceCode,
		entity.ExternalID,
		entity.EntityType,
		entity.ResolvedID,
		entity.Name,
		string(attrsJSON),
		string(validationJSON),
		publishedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("insert published entity: %w", err)
	}
	return nil
}

func (s *SQLiteStagingStore) ListPublishedBySource(ctx context.Context, source string) ([]PublishedEntity, error) {
	query := `SELECT source_code, external_id, entity_type, resolved_id, name, attributes_json, validation_json, published_at
		FROM published_entities`
	args := make([]any, 0, 1)
	if source != "" {
		query += ` WHERE source_code = ?`
		args = append(args, source)
	}
	query += ` ORDER BY id DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query published entities: %w", err)
	}
	defer rows.Close()

	var out []PublishedEntity
	for rows.Next() {
		var entity PublishedEntity
		var attrsJSON string
		var validationJSON string
		var publishedAt string
		if err := rows.Scan(
			&entity.SourceCode,
			&entity.ExternalID,
			&entity.EntityType,
			&entity.ResolvedID,
			&entity.Name,
			&attrsJSON,
			&validationJSON,
			&publishedAt,
		); err != nil {
			return nil, fmt.Errorf("scan published entity: %w", err)
		}
		if err := json.Unmarshal([]byte(attrsJSON), &entity.Attributes); err != nil {
			return nil, fmt.Errorf("unmarshal published attributes: %w", err)
		}
		if err := json.Unmarshal([]byte(validationJSON), &entity.Validation); err != nil {
			return nil, fmt.Errorf("unmarshal published validation: %w", err)
		}
		if publishedAt != "" {
			if parsed, err := time.Parse(time.RFC3339Nano, publishedAt); err == nil {
				entity.PublishedAt = parsed
			}
		}
		out = append(out, entity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate published entities: %w", err)
	}
	return out, nil
}

func PublishSourceRecords(ctx context.Context, store *SQLiteStagingStore, source string) (int, error) {
	records, err := store.ListStaged(ctx, source, "")
	if err != nil {
		return 0, err
	}
	if len(records) == 0 {
		return 0, nil
	}
	existing, err := store.ListPublishedBySource(ctx, "")
	if err != nil {
		return 0, err
	}
	known := make(map[string]int64, len(existing))
	for _, entity := range existing {
		known[entity.Name] = entity.ResolvedID
	}
	resolver := NewEntityResolver(known)
	published := 0
	for _, record := range records {
		norm, validation, ok := normalizeForPublish(record)
		if !ok {
			continue
		}
		match := resolver.Resolve(norm)
		if err := store.SavePublished(ctx, PublishedEntity{
			SourceCode:  record.SourceName,
			ExternalID:  record.ExternalID,
			EntityType:  norm.EntityType,
			ResolvedID:  match.ResolvedID,
			Name:        norm.Name,
			Attributes:  norm.Attributes,
			Validation:  validation,
			PublishedAt: time.Now().UTC(),
		}); err != nil {
			return published, err
		}
		if err := store.MarkProcessed(ctx, record.SourceName, record.ExternalID); err != nil {
			return published, err
		}
		published++
	}
	return published, nil
}

func normalizeForPublish(record SourceRecord) (NormalizedRecord, ValidationResult, bool) {
	switch record.SourceName {
	case "fbref":
		norm, err := NormalizeFbrefPlayerRecord(record)
		if err != nil {
			return NormalizedRecord{}, ValidationResult{}, false
		}
		return norm, ValidatePlayerRecord(norm.Attributes), true
	default:
		norm := NormalizedRecord{
			SourceName: record.SourceName,
			ExternalID: record.ExternalID,
			EntityType: record.EntityType,
			Name:       record.ExternalID,
			Attributes: map[string]string{"name": record.ExternalID},
		}
		validation := ValidationResult{EntityType: record.EntityType, ExternalID: record.ExternalID, Valid: true}
		return norm, validation, true
	}
}
