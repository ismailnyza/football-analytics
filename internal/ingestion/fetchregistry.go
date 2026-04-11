package ingestion

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const IngestLedgerFile = "ingest_state.json"

var DefaultSourceOrder = []string{
	"fbref",
	"transfermarkt",
	"understat",
	"whoscored",
}

const DefaultCapPerRun = 2500

type SourceState struct {
	LastRunAt     time.Time `json:"last_run_at"`
	LastStaged    int       `json:"last_staged"`
	LastErrors    int       `json:"last_errors"`
	LastPublished int       `json:"last_published"`
	CapPerRun     int       `json:"cap_per_run"`
	LastError     string    `json:"last_error,omitempty"`
}

type fetchStatePayload struct {
	Version int                    `json:"version"`
	Sources map[string]SourceState `json:"sources"`
}

type FetchRegistry struct {
	path string
	mu   sync.Mutex
	data fetchStatePayload
}

func seedSources() map[string]SourceState {
	m := make(map[string]SourceState, len(DefaultSourceOrder))
	for _, name := range DefaultSourceOrder {
		m[name] = SourceState{CapPerRun: DefaultCapPerRun}
	}
	return m
}

func LoadOrCreateRegistry(stateDir string) (*FetchRegistry, error) {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(stateDir, IngestLedgerFile)
	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		reg := &FetchRegistry{
			path: path,
			data: fetchStatePayload{Version: 1, Sources: seedSources()},
		}
		if err := reg.saveUnlocked(); err != nil {
			return nil, err
		}
		return reg, nil
	}
	var raw fetchStatePayload
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	if raw.Version == 0 {
		raw.Version = 1
	}
	if raw.Sources == nil {
		raw.Sources = make(map[string]SourceState)
	}
	for _, name := range DefaultSourceOrder {
		if _, ok := raw.Sources[name]; !ok {
			raw.Sources[name] = SourceState{CapPerRun: DefaultCapPerRun}
		}
	}
	return &FetchRegistry{path: path, data: raw}, nil
}

func (r *FetchRegistry) Path() string {
	return r.path
}

func (r *FetchRegistry) CapFor(source string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	st, ok := r.data.Sources[source]
	if !ok || st.CapPerRun <= 0 {
		return DefaultCapPerRun
	}
	return st.CapPerRun
}

func (r *FetchRegistry) RecordRun(source string, at time.Time, staged, errs, published int, fetchErr error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	st, ok := r.data.Sources[source]
	if !ok {
		st = SourceState{CapPerRun: DefaultCapPerRun}
	}
	st.LastRunAt = at
	st.LastStaged = staged
	st.LastErrors = errs
	st.LastPublished = published
	if fetchErr != nil {
		st.LastError = fetchErr.Error()
	} else {
		st.LastError = ""
	}
	r.data.Sources[source] = st
}

func (r *FetchRegistry) Save() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.saveUnlocked()
}

func (r *FetchRegistry) saveUnlocked() error {
	r.data.Version = 1
	b, err := json.MarshalIndent(&r.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}

type SourceSnapshot struct {
	Name      string
	LastRunAt time.Time
	Records   int
	Errors    int
	Published int
	CapPerRun int
	LastError string
}

func (r *FetchRegistry) Snapshots() []SourceSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	seen := make(map[string]struct{})
	order := make([]string, 0, len(r.data.Sources))
	for _, name := range DefaultSourceOrder {
		if _, ok := r.data.Sources[name]; !ok {
			continue
		}
		seen[name] = struct{}{}
		order = append(order, name)
	}
	extra := make([]string, 0)
	for name := range r.data.Sources {
		if _, ok := seen[name]; ok {
			continue
		}
		extra = append(extra, name)
	}
	sort.Strings(extra)
	order = append(order, extra...)
	out := make([]SourceSnapshot, 0, len(order))
	for _, name := range order {
		st := r.data.Sources[name]
		cap := st.CapPerRun
		if cap <= 0 {
			cap = DefaultCapPerRun
		}
		out = append(out, SourceSnapshot{
			Name:      name,
			LastRunAt: st.LastRunAt,
			Records:   st.LastStaged,
			Errors:    st.LastErrors,
			Published: st.LastPublished,
			CapPerRun: cap,
			LastError: st.LastError,
		})
	}
	return out
}
