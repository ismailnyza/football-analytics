package ingestion

import (
	"context"
	"time"
)

func RunAdapters(ctx context.Context, store StagingStore, reg *FetchRegistry, adapters []SourceAdapter, now time.Time) error {
	for _, ad := range adapters {
		capN := reg.CapFor(ad.Name())
		wrapped := NewCappedStagingStore(store, capN)
		staged, err := ad.Fetch(ctx, wrapped)
		published := staged
		if err != nil {
			reg.RecordRun(ad.Name(), now, staged, 1, 0, err)
			continue
		}
		reg.RecordRun(ad.Name(), now, staged, 0, published, nil)
	}
	return reg.Save()
}

func DefaultDemoAdapters() []SourceAdapter {
	return []SourceAdapter{
		NewStaticAdapter("fbref", []SourceRecord{
			{EntityType: "player", ExternalID: "demo-fb-1", RawJSON: `{"name":"Demo Striker","nation":"ZZ"}`},
			{EntityType: "player", ExternalID: "demo-fb-2", RawJSON: `{"name":"Demo Mid","nation":"ZZ"}`},
			{EntityType: "club", ExternalID: "demo-fb-c1", RawJSON: `{"name":"Demo United"}`},
		}),
		NewStaticAdapter("transfermarkt", []SourceRecord{
			{EntityType: "player", ExternalID: "demo-tm-1", RawJSON: `{"name":"TM Demo","pos":"RW"}`},
		}),
		NewStaticAdapter("understat", []SourceRecord{
			{EntityType: "player", ExternalID: "demo-us-1", RawJSON: `{"xg":"0.42"}`},
			{EntityType: "player", ExternalID: "demo-us-2", RawJSON: `{"xg":"0.11"}`},
		}),
		NewStaticAdapter("whoscored", []SourceRecord{
			{EntityType: "player", ExternalID: "demo-ws-1", RawJSON: `{"rating":"7.1"}`},
		}),
	}
}

func RunDemoFetch(ctx context.Context, stateDir string, now time.Time) error {
	reg, err := LoadOrCreateRegistry(stateDir)
	if err != nil {
		return err
	}
	store, closeFn, err := resolveSQLiteStateStore(stateDir)
	if err != nil {
		return err
	}
	defer closeFn()
	return RunAdapters(ctx, store, reg, DefaultDemoAdapters(), now)
}
