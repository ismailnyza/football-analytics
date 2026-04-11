package ingestion

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const FbrefURLFile = "fbref_url.txt"

func WriteFbrefURLFile(stateDir, pageURL string) error {
	pageURL = strings.TrimSpace(pageURL)
	if pageURL == "" {
		return fmt.Errorf("fbref URL must not be empty")
	}
	path := filepath.Join(stateDir, FbrefURLFile)
	return os.WriteFile(path, []byte(pageURL+"\n"), 0o644)
}

func ReadFbrefURLFile(stateDir string) (string, error) {
	path := filepath.Join(stateDir, FbrefURLFile)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line, nil
	}
	return "", fmt.Errorf("no non-empty URL line in %s", FbrefURLFile)
}

func RunFbrefScrape(ctx context.Context, stateDir, pageURL string, now time.Time) error {
	reg, err := LoadOrCreateRegistry(stateDir)
	if err != nil {
		return err
	}
	store, closeFn, err := resolveSQLiteStateStore(stateDir)
	if err != nil {
		return err
	}
	defer closeFn()
	ad := NewFbrefPlayerTableAdapter(pageURL)
	return RunAdapters(ctx, store, reg, []SourceAdapter{ad}, now)
}

func RunFbrefScrapeFromStateFile(ctx context.Context, stateDir string, now time.Time) error {
	u, err := ReadFbrefURLFile(stateDir)
	if err != nil {
		return err
	}
	return RunFbrefScrape(ctx, stateDir, u, now)
}
