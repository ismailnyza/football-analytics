package baseline

type TeamFormRecord struct {
	Team            string
	RecentWins      int
	RecentDraws     int
	RecentLosses    int
	RecentGF        int
	RecentGA        int
	PointsPerGame   float64
	GoalDiffPerGame float64
	Window          int
}

type FormMap map[string]TeamFormRecord

func ExtractTeamForm(matches []Match, window int) FormMap {
	type result struct {
		goalsFor     int
		goalsAgainst int
		points       int
	}
	history := make(map[string][]result)

	for _, m := range matches {
		history[m.HomeTeam] = append(history[m.HomeTeam], result{
			goalsFor:     m.HomeGoals,
			goalsAgainst: m.AwayGoals,
			points:       pointsFromResult(m.Result, true),
		})
		history[m.AwayTeam] = append(history[m.AwayTeam], result{
			goalsFor:     m.AwayGoals,
			goalsAgainst: m.HomeGoals,
			points:       pointsFromResult(m.Result, false),
		})
	}

	form := make(FormMap)
	for team, results := range history {
		start := len(results) - window
		if start < 0 {
			start = 0
		}
		recent := results[start:]
		f := TeamFormRecord{Team: team, Window: len(recent)}
		for _, r := range recent {
			f.RecentGF += r.goalsFor
			f.RecentGA += r.goalsAgainst
			switch r.points {
			case 3:
				f.RecentWins++
			case 1:
				f.RecentDraws++
			case 0:
				f.RecentLosses++
			}
		}
		n := float64(len(recent))
		if n > 0 {
			f.PointsPerGame = float64(f.RecentWins*3+f.RecentDraws) / n
			f.GoalDiffPerGame = float64(f.RecentGF-f.RecentGA) / n
		}
		form[team] = f
	}
	return form
}

func CalculateFormDelta(homeForm, awayForm TeamFormRecord) float64 {
	return (homeForm.PointsPerGame-awayForm.PointsPerGame)*0.5 +
		(homeForm.GoalDiffPerGame-awayForm.GoalDiffPerGame)*0.3
}

func pointsFromResult(result string, isHome bool) int {
	switch result {
	case "H":
		if isHome {
			return 3
		}
		return 0
	case "D":
		return 1
	case "A":
		if isHome {
			return 0
		}
		return 3
	}
	return 0
}
