package ingestion

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadOrCreateRegistry_createsFile(t *testing.T) {
	dir := t.TempDir()
	reg, err := LoadOrCreateRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	if reg.Path() != filepath.Join(dir, IngestLedgerFile) {
		t.Fatalf("path %s", reg.Path())
	}
	reg2, err := LoadOrCreateRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(reg2.Snapshots()) < len(DefaultSourceOrder) {
		t.Fatalf("snapshots %d", len(reg2.Snapshots()))
	}
}

func TestFetchRegistry_RecordRun_roundTrip(t *testing.T) {
	dir := t.TempDir()
	reg, err := LoadOrCreateRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	reg.RecordRun("fbref", at, 10, 1, 9, nil)
	if err := reg.Save(); err != nil {
		t.Fatal(err)
	}
	reg3, err := LoadOrCreateRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	snaps := reg3.Snapshots()
	var fb *SourceSnapshot
	for i := range snaps {
		if snaps[i].Name == "fbref" {
			fb = &snaps[i]
			break
		}
	}
	if fb == nil {
		t.Fatal("missing fbref")
	}
	if !fb.LastRunAt.Equal(at) {
		t.Fatalf("time %v", fb.LastRunAt)
	}
	if fb.Records != 10 || fb.Errors != 1 || fb.Published != 9 {
		t.Fatalf("%+v", fb)
	}
}
