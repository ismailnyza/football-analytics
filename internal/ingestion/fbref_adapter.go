package ingestion

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/ismael/football-analytics/internal/ingestion/fbref"
)

const fbrefUserAgent = "football-analytics/0.1 (local ingestion; respect robots.txt)"

type fbrefPlayerTableAdapter struct {
	url    string
	client *http.Client
}

func NewFbrefPlayerTableAdapter(pageURL string) SourceAdapter {
	return &fbrefPlayerTableAdapter{
		url: strings.TrimSpace(pageURL),
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (a *fbrefPlayerTableAdapter) Name() string { return "fbref" }

func (a *fbrefPlayerTableAdapter) Fetch(ctx context.Context, store StagingStore) (int, error) {
	if a.url == "" {
		return 0, fmt.Errorf("fbref: empty page URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", fbrefUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := a.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fbref: request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return 0, fmt.Errorf("fbref: HTTP %d %s", resp.StatusCode, strings.TrimSpace(string(slurp)))
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("fbref: parse html: %w", err)
	}
	rows, err := fbref.ParsePlayerStatRows(doc)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	staged := 0
	for _, row := range rows {
		raw, err := fbref.RowToRawJSON(row)
		if err != nil {
			return staged, err
		}
		rec := SourceRecord{
			SourceName: a.Name(),
			EntityType: "player",
			ExternalID: row.ExternalID,
			RawJSON:    raw,
			FetchedAt:  now,
		}
		if err := store.StageRecord(ctx, rec); err != nil {
			return staged, fmt.Errorf("fbref: stage %s: %w", row.ExternalID, err)
		}
		staged++
	}
	return staged, nil
}
