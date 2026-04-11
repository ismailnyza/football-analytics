// Package mlbridge provides the interface between the Go simulation engine and
// the Python calibration layer. It defines the data exchange contracts and a
// file-based bridge implementation.
package mlbridge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CalibrationRequest is the payload written to disk for the Python process.
type CalibrationRequest struct {
	RequestID  string          `json:"request_id"`
	CreatedAt  time.Time       `json:"created_at"`
	EntityType string          `json:"entity_type"` // "match_model", "player_ratings"
	InputData  json.RawMessage `json:"input_data"`
}

// CalibrationResult is the response produced by the Python process.
type CalibrationResult struct {
	RequestID  string          `json:"request_id"`
	CompletedAt time.Time      `json:"completed_at"`
	OutputData json.RawMessage `json:"output_data"`
	Error      string          `json:"error,omitempty"`
}

// Bridge writes calibration requests to a directory and reads responses back.
type Bridge struct {
	dir string
}

// NewBridge creates a bridge that uses the given directory for I/O files.
func NewBridge(dir string) *Bridge {
	return &Bridge{dir: dir}
}

// WriteRequest serialises a calibration request to a JSON file.
func (b *Bridge) WriteRequest(req CalibrationRequest) error {
	path := filepath.Join(b.dir, req.RequestID+".req.json")
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("bridge: marshal request: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("bridge: write request: %w", err)
	}
	return nil
}

// ReadResult reads and parses a calibration result file if it exists.
// Returns (nil, nil) when the result file is not yet available.
func (b *Bridge) ReadResult(requestID string) (*CalibrationResult, error) {
	path := filepath.Join(b.dir, requestID+".res.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("bridge: read result: %w", err)
	}
	var result CalibrationResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("bridge: parse result: %w", err)
	}
	return &result, nil
}

// PendingRequests returns the IDs of request files that have no matching result.
func (b *Bridge) PendingRequests() ([]string, error) {
	entries, err := os.ReadDir(b.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("bridge: read dir: %w", err)
	}
	var pending []string
	for _, e := range entries {
		name := e.Name()
		if filepath.Ext(name) == ".json" && len(name) > 9 && name[len(name)-9:] == ".req.json" {
			id := name[:len(name)-9]
			resPath := filepath.Join(b.dir, id+".res.json")
			if _, err := os.Stat(resPath); os.IsNotExist(err) {
				pending = append(pending, id)
			}
		}
	}
	return pending, nil
}
