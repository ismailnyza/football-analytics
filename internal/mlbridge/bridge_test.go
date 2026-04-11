package mlbridge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteAndReadRequest(t *testing.T) {
	dir := t.TempDir()
	bridge := NewBridge(dir)

	req := CalibrationRequest{
		RequestID:  "test-001",
		CreatedAt:  time.Now(),
		EntityType: "match_model",
		InputData:  json.RawMessage(`{"seed":42}`),
	}
	if err := bridge.WriteRequest(req); err != nil {
		t.Fatalf("write request: %v", err)
	}

	path := filepath.Join(dir, "test-001.req.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("request file not found: %v", err)
	}
}

func TestReadResult_notReady(t *testing.T) {
	dir := t.TempDir()
	bridge := NewBridge(dir)

	result, err := bridge.ReadResult("no-such-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result when file not present")
	}
}

func TestReadResult_ready(t *testing.T) {
	dir := t.TempDir()
	bridge := NewBridge(dir)

	res := CalibrationResult{
		RequestID:   "r1",
		CompletedAt: time.Now(),
		OutputData:  json.RawMessage(`{"score":0.85}`),
	}
	data, _ := json.Marshal(res)
	_ = os.WriteFile(filepath.Join(dir, "r1.res.json"), data, 0o644)

	result, err := bridge.ReadResult("r1")
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.RequestID != "r1" {
		t.Fatalf("request ID = %s, want r1", result.RequestID)
	}
}

func TestPendingRequests(t *testing.T) {
	dir := t.TempDir()
	bridge := NewBridge(dir)

	req := CalibrationRequest{RequestID: "pending-1", CreatedAt: time.Now(), InputData: json.RawMessage("{}")}
	_ = bridge.WriteRequest(req)

	pending, err := bridge.PendingRequests()
	if err != nil {
		t.Fatalf("pending: %v", err)
	}
	if len(pending) != 1 || pending[0] != "pending-1" {
		t.Fatalf("pending = %v, want [pending-1]", pending)
	}

	// Write result → no longer pending.
	res := CalibrationResult{RequestID: "pending-1", CompletedAt: time.Now(), OutputData: json.RawMessage("{}")}
	data, _ := json.Marshal(res)
	_ = os.WriteFile(filepath.Join(dir, "pending-1.res.json"), data, 0o644)

	pending, _ = bridge.PendingRequests()
	if len(pending) != 0 {
		t.Fatalf("expected 0 pending after result written, got %d", len(pending))
	}
}
