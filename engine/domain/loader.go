package domain

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type PlayerStatsLoader interface {
	LoadMatches(pattern string) ([]PlayerMatchStats, []string, error)
	LoadMatchesFromPaths(paths []string) ([]PlayerMatchStats, []string, error)
}

func LoadPlayerStatsFromGlob(pattern string) ([]PlayerMatchStats, []string, error) {
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, nil, fmt.Errorf("glob player stats files: %w", err)
	}
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("no files matched pattern %q", pattern)
	}
	return LoadPlayerStatsFromPaths(paths)
}

func LoadPlayerStatsFromPaths(paths []string) ([]PlayerMatchStats, []string, error) {
	sort.Strings(paths)
	var stats []PlayerMatchStats
	var used []string
	for _, path := range paths {
		fileStats, err := loadPlayerStatsCSV(path)
		if err != nil {
			return nil, nil, err
		}
		stats = append(stats, fileStats...)
		used = append(used, path)
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Date == stats[j].Date {
			if stats[i].Season == stats[j].Season {
				return stats[i].PlayerID < stats[j].PlayerID
			}
			return stats[i].Season < stats[j].Season
		}
		return stats[i].Date < stats[j].Date
	})
	return stats, used, nil
}

func loadPlayerStatsCSV(path string) ([]PlayerMatchStats, error) {
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
		headers[strings.TrimSpace(h)] = i
	}
	required := []string{"player_id", "match_id", "season", "date", "minutes", "goals", "assists"}
	for _, key := range required {
		if _, ok := headers[key]; !ok {
			return nil, fmt.Errorf("csv %s missing required column %s", path, key)
		}
	}

	stats := make([]PlayerMatchStats, 0, len(records)-1)
	for _, row := range records[1:] {
		s, err := parsePlayerStatsRow(row, headers, filepath.Base(path))
		if err != nil {
			return nil, fmt.Errorf("parse row in %s: %w", path, err)
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func parsePlayerStatsRow(row []string, headers map[string]int, source string) (PlayerMatchStats, error) {
	get := func(key string) string { return strings.TrimSpace(row[headers[key]]) }
	getInt := func(key string) (int, error) { return strconv.Atoi(get(key)) }

	minutes, err := getInt("minutes")
	if err != nil {
		return PlayerMatchStats{}, fmt.Errorf("minutes: %w", err)
	}
	goals, err := getInt("goals")
	if err != nil {
		return PlayerMatchStats{}, fmt.Errorf("goals: %w", err)
	}
	assists, err := getInt("assists")
	if err != nil {
		return PlayerMatchStats{}, fmt.Errorf("assists: %w", err)
	}

	s := PlayerMatchStats{
		PlayerID:       get("player_id"),
		MatchID:        get("match_id"),
		Season:         get("season"),
		Date:           get("date"),
		Minutes:        minutes,
		Goals:          goals,
		Assists:        assists,
		PositionPlayed: safeGet(row, headers, "position_played"),
		SourceFile:     source,
	}

	if idx, ok := headers["shots"]; ok {
		s.Shots, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
	}
	if idx, ok := headers["shots_on_target"]; ok {
		s.ShotsOnTarget, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
	}
	if idx, ok := headers["xg"]; ok {
		s.XG, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
	}
	if idx, ok := headers["xa"]; ok {
		s.XA, _ = strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
	}
	if idx, ok := headers["passes"]; ok {
		s.Passes, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
	}
	if idx, ok := headers["key_passes"]; ok {
		s.KeyPasses, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
	}
	if idx, ok := headers["tackles"]; ok {
		s.Tackles, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
	}
	if idx, ok := headers["interceptions"]; ok {
		s.Interceptions, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
	}
	if idx, ok := headers["fouls"]; ok {
		s.Fouls, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
	}
	if idx, ok := headers["fouled"]; ok {
		s.Fouled, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
	}
	if idx, ok := headers["yellow_cards"]; ok {
		s.YellowCards, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
	}
	if idx, ok := headers["red_cards"]; ok {
		s.RedCards, _ = strconv.Atoi(strings.TrimSpace(row[idx]))
	}

	return s, nil
}

func safeGet(row []string, headers map[string]int, key string) string {
	if idx, ok := headers[key]; ok && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}
