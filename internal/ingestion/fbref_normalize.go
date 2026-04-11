package ingestion

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ismael/football-analytics/internal/domain"
)

func NormalizeFbrefPlayerRecord(raw SourceRecord) (NormalizedRecord, error) {
	if raw.EntityType != "" && raw.EntityType != "player" {
		return NormalizedRecord{}, fmt.Errorf("fbref normalize: entity %q", raw.EntityType)
	}
	var stats map[string]string
	if err := json.Unmarshal([]byte(raw.RawJSON), &stats); err != nil {
		return NormalizedRecord{}, fmt.Errorf("fbref normalize json: %w", err)
	}
	name := strings.TrimSpace(stats["player"])
	if name == "" {
		return NormalizedRecord{}, fmt.Errorf("fbref normalize: missing player name")
	}
	posCode := strings.TrimSpace(stats["pos"])
	posStr := mapFBrefPosition(posCode)
	if _, err := domain.ParsePosition(posStr); err != nil {
		posStr = "CM"
	}
	nat := strings.TrimSpace(stats["nation"])
	if idx := strings.IndexByte(nat, ' '); idx > 0 && len(nat) > idx+1 {
		nat = strings.TrimSpace(nat[idx+1:])
	}
	attrs := map[string]string{
		"name":        name,
		"position":    posStr,
		"nationality": nat,
	}
	if age := strings.TrimSpace(stats["age"]); age != "" {
		attrs["age_display"] = age
	}
	if overall := approximateOverallFromFBrefStats(stats); overall > 0 {
		attrs["overall_approx"] = fmt.Sprintf("%d", overall)
	}
	return NormalizedRecord{
		SourceName: raw.SourceName,
		ExternalID: raw.ExternalID,
		EntityType: "player",
		Name:       name,
		Attributes: attrs,
	}, nil
}

func mapFBrefPosition(code string) string {
	u := strings.ToUpper(strings.TrimSpace(code))
	if u == "" {
		return "CM"
	}
	if p, err := domain.ParsePosition(u); err == nil {
		return string(p)
	}
	switch {
	case strings.Contains(u, "GK"):
		return "GK"
	case strings.Contains(u, "LB"):
		return "LB"
	case strings.Contains(u, "RB"):
		return "RB"
	case strings.Contains(u, "CB") || strings.Contains(u, "DF"):
		return "CB"
	case strings.Contains(u, "DM"):
		return "DM"
	case strings.Contains(u, "AM"):
		return "AM"
	case strings.Contains(u, "LW"):
		return "LW"
	case strings.Contains(u, "RW"):
		return "RW"
	case strings.Contains(u, "FW") || strings.Contains(u, "ST") || u == "F":
		return "ST"
	case strings.Contains(u, "MF") || strings.Contains(u, "CM"):
		return "CM"
	default:
		return "CM"
	}
}

func approximateOverallFromFBrefStats(stats map[string]string) int {
	if x := firstStatInt(stats, "goals", "gls"); x > 15 {
		return clampInt(75+x/3, 60, 92)
	}
	if x := firstStatInt(stats, "assists", "ast"); x > 12 {
		return clampInt(72+x/2, 58, 90)
	}
	if mp := firstStatInt(stats, "minutes", "min"); mp > 2000 {
		return clampInt(68+mp/500, 62, 84)
	}
	return 0
}

func firstStatInt(stats map[string]string, keys ...string) int {
	for _, k := range keys {
		if n := statIntField(stats, k); n > 0 {
			return n
		}
	}
	return 0
}

func statIntField(stats map[string]string, key string) int {
	var n int
	_, _ = fmt.Sscanf(strings.TrimSpace(stats[key]), "%d", &n)
	return n
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
