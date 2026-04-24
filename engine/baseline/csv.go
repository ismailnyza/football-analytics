package baseline

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "02/01/2006"

var competitionName = map[string]string{
	"E0":  "EPL",
	"SP1": "LaLiga",
	"D1":  "Bundesliga",
	"I1":  "SerieA",
	"F1":  "Ligue1",
}

func CompetitionFromCode(code string) string {
	if name, ok := competitionName[code]; ok {
		return name
	}
	return code
}

func LoadMatchesFromGlob(pattern string) ([]Match, []string, error) {
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, nil, fmt.Errorf("glob match files: %w", err)
	}
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("no files matched pattern %q", pattern)
	}
	return LoadMatches(paths)
}

func LoadMatches(paths []string) ([]Match, []string, error) {
	sort.Strings(paths)
	var matches []Match
	seen := make(map[string]struct{})
	var used []string
	for _, path := range paths {
		fileMatches, err := loadCSV(path)
		if err != nil {
			return nil, nil, err
		}
		matches = append(matches, fileMatches...)
		used = append(used, path)
		for _, m := range fileMatches {
			key := fmt.Sprintf("%s|%s|%s|%s", m.Date.Format(time.RFC3339), m.HomeTeam, m.AwayTeam, m.SourceFile)
			seen[key] = struct{}{}
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Date.Equal(matches[j].Date) {
			if matches[i].Season == matches[j].Season {
				if matches[i].HomeTeam == matches[j].HomeTeam {
					return matches[i].AwayTeam < matches[j].AwayTeam
				}
				return matches[i].HomeTeam < matches[j].HomeTeam
			}
			return matches[i].Season < matches[j].Season
		}
		return matches[i].Date.Before(matches[j].Date)
	})
	return matches, used, nil
}

func loadCSV(path string) ([]Match, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv %s: %w", path, err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("csv %s has no rows", path)
	}

	headers := make(map[string]int)
	for i, h := range records[0] {
		headers[h] = i
	}
	required := []string{"Date", "HomeTeam", "AwayTeam", "FTHG", "FTAG", "FTR"}
	for _, key := range required {
		if _, ok := headers[key]; !ok {
			return nil, fmt.Errorf("csv %s missing required column %s", path, key)
		}
	}

	season := seasonFromFilename(path)
	competition := competitionFromPath(path)
	matches := make([]Match, 0, len(records)-1)
	for _, row := range records[1:] {
		date, err := time.Parse(dateLayout, row[headers["Date"]])
		if err != nil {
			return nil, fmt.Errorf("parse date %q in %s: %w", row[headers["Date"]], path, err)
		}
		hg, err := strconv.Atoi(row[headers["FTHG"]])
		if err != nil {
			return nil, fmt.Errorf("parse FTHG in %s: %w", path, err)
		}
		ag, err := strconv.Atoi(row[headers["FTAG"]])
		if err != nil {
			return nil, fmt.Errorf("parse FTAG in %s: %w", path, err)
		}
		result := strings.TrimSpace(row[headers["FTR"]])
		matches = append(matches, Match{
			Competition: competition,
			Season:      season,
			Date:        date,
			HomeTeam:    strings.TrimSpace(row[headers["HomeTeam"]]),
			AwayTeam:    strings.TrimSpace(row[headers["AwayTeam"]]),
			HomeGoals:   hg,
			AwayGoals:   ag,
			Result:      result,
			SourceFile:  filepath.Base(path),
		})
	}
	return matches, nil
}

func competitionFromPath(path string) string {
	base := filepath.Base(path)
	parts := strings.SplitN(base, "_", 2)
	if len(parts) > 0 {
		return CompetitionFromCode(parts[0])
	}
	return "Unknown"
}

func seasonFromFilename(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	parts := strings.Split(base, "_")
	return parts[len(parts)-1]
}
