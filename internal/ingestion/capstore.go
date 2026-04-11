package ingestion

import (
	"context"
)

type CappedStagingStore struct {
	inner StagingStore
	max   int
	n     int
}

func NewCappedStagingStore(inner StagingStore, maxPerRun int) *CappedStagingStore {
	return &CappedStagingStore{inner: inner, max: maxPerRun}
}

func (c *CappedStagingStore) StagedCount() int { return c.n }

func (c *CappedStagingStore) StageRecord(ctx context.Context, r SourceRecord) error {
	if c.max > 0 && c.n >= c.max {
		return ErrCapReached
	}
	if err := c.inner.StageRecord(ctx, r); err != nil {
		return err
	}
	c.n++
	return nil
}

func (c *CappedStagingStore) ListStaged(ctx context.Context, source, entityType string) ([]SourceRecord, error) {
	return c.inner.ListStaged(ctx, source, entityType)
}

func (c *CappedStagingStore) MarkProcessed(ctx context.Context, source, externalID string) error {
	return c.inner.MarkProcessed(ctx, source, externalID)
}
