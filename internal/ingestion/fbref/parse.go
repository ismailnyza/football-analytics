package fbref

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PlayerStatRow struct {
	ExternalID string
	Stats      map[string]string
}

func ParsePlayerStatRows(doc *goquery.Document) ([]PlayerStatRow, error) {
	var collected []PlayerStatRow
	doc.Find("table.stats_table").Each(func(_ int, table *goquery.Selection) {
		table.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
			th := tr.Find("th[data-stat='player']")
			if th.Length() == 0 {
				return
			}
			link := th.Find("a[href*='/players/']")
			if link.Length() == 0 {
				return
			}
			href, _ := link.Attr("href")
			id := playerIDFromHref(href)
			if id == "" {
				return
			}
			name := strings.TrimSpace(link.Text())
			stats := make(map[string]string)
			stats["player"] = name
			tr.Find("td[data-stat]").Each(func(_ int, td *goquery.Selection) {
				k, _ := td.Attr("data-stat")
				if k == "" {
					return
				}
				stats[k] = strings.TrimSpace(td.Text())
			})
			collected = append(collected, PlayerStatRow{ExternalID: id, Stats: stats})
		})
	})
	if len(collected) == 0 {
		return nil, fmt.Errorf("fbref: no player rows found in stats_table")
	}
	return collected, nil
}

func playerIDFromHref(href string) string {
	href = strings.TrimSpace(href)
	parts := strings.Split(strings.Trim(href, "/"), "/")
	for i, p := range parts {
		if p == "players" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func RowToRawJSON(row PlayerStatRow) (string, error) {
	b, err := json.Marshal(row.Stats)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
