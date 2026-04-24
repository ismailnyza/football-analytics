package predict

import (
	"fmt"
	"math"

	"github.com/ismailnyza/football-analytics/engine/baseline"
	"github.com/ismailnyza/football-analytics/engine/feature"
)

type Pipeline struct {
	Registry    *feature.Registry
	PlayerPool  *PlayerPool
	TeamRatings map[string]float64
	TeamForm    baseline.FormMap
	Config      baseline.Config
	FormWindow  int
}

func NewPipeline(teamRatings map[string]float64, form baseline.FormMap, config baseline.Config) *Pipeline {
	pool := NewPlayerPool(teamRatings)
	reg := feature.NewRegistry()
	return &Pipeline{
		Registry:    reg,
		PlayerPool:  pool,
		TeamRatings: teamRatings,
		TeamForm:    form,
		Config:      config,
		FormWindow:  5,
	}
}

func (p *Pipeline) Predict(homeTeam, awayTeam string) (MatchPrediction, error) {
	homeRating := p.getRating(homeTeam)
	awayRating := p.getRating(awayTeam)

	probs := p.computeProbabilities(homeRating, awayRating)
	predictedResult := argmax(probs)

	homeForm := p.getFormSnapshot(homeTeam)
	awayForm := p.getFormSnapshot(awayTeam)

	expectedHomeGoals, expectedAwayGoals := p.expectedGoals(homeRating, awayRating, probs)

	homeGoals := int(math.Round(expectedHomeGoals))
	awayGoals := int(math.Round(expectedAwayGoals))
	if predictedResult == "H" && homeGoals <= awayGoals {
		homeGoals = awayGoals + 1
	}
	if predictedResult == "A" && awayGoals <= homeGoals {
		awayGoals = homeGoals + 1
	}
	predictedScoreline := fmt.Sprintf("%d-%d", homeGoals, awayGoals)

	homeScorers := p.PlayerPool.ScorerProbabilities(homeTeam, int(math.Round(expectedHomeGoals)))
	awayScorers := p.PlayerPool.ScorerProbabilities(awayTeam, int(math.Round(expectedAwayGoals)))
	homeAssists := p.PlayerPool.AssistProbabilities(homeTeam, int(math.Round(expectedHomeGoals)))
	awayAssists := p.PlayerPool.AssistProbabilities(awayTeam, int(math.Round(expectedAwayGoals)))

	teamEvents := p.expectedTeamEvents(homeRating, awayRating, probs, homeTeam, awayTeam)

	homeStats := teamEventsToRichStats(teamEvents[0])
	awayStats := teamEventsToRichStats(teamEvents[1])
	homeStats.FullTimeGoals = homeGoals
	awayStats.FullTimeGoals = awayGoals
	reconstructed := PredictMatchWithEvents(homeTeam, awayTeam, homeGoals, awayGoals, p.PlayerPool, homeStats, awayStats)

	return MatchPrediction{
		HomeTeam: homeTeam,
		AwayTeam: awayTeam,
		ResultProbabilities: ResultProbabilities{
			HomeWin: probs.HomeWin,
			Draw:    probs.Draw,
			AwayWin: probs.AwayWin,
		},
		PredictedResult:    predictedResult,
		ExpectedHomeGoals:  expectedHomeGoals,
		ExpectedAwayGoals:  expectedAwayGoals,
		PredictedScoreline: predictedScoreline,
		HomeScorers:        limitSlice(homeScorers, 5),
		AwayScorers:        limitSlice(awayScorers, 5),
		HomeAssists:        limitSlice(homeAssists, 5),
		AwayAssists:        limitSlice(awayAssists, 5),
		TeamEvents:         teamEvents,
		Reconstructed:      &reconstructed,
		HomeForm:           homeForm,
		AwayForm:           awayForm,
		PlayerPoolSize:     len(p.PlayerPool.Players),
	}, nil
}

func (p *Pipeline) getRating(team string) float64 {
	if r, ok := p.TeamRatings[team]; ok {
		return r
	}
	return p.Config.InitialRating
}

func (p *Pipeline) computeProbabilities(homeRating, awayRating float64) baseline.Probabilities {
	diff := (homeRating + p.Config.HomeAdvantage - awayRating) / p.Config.Scale

	if p.Config.ModelFamily == baseline.ModelFamilyDrawDecay {
		pDraw := p.Config.BaseDraw * math.Exp(-math.Abs(diff*p.Config.Scale)/p.Config.DrawScale)
		pDraw = clamp(pDraw, 0.01, 0.60)
		pHomeNoDraw := 1.0 / (1.0 + math.Pow(10, -diff))
		return baseline.Probabilities{
			Draw:    pDraw,
			HomeWin: (1.0 - pDraw) * pHomeNoDraw,
			AwayWin: (1.0 - pDraw) * (1.0 - pHomeNoDraw),
		}
	}

	if p.Config.ModelFamily == baseline.ModelFamilyOrderedProbit {
		c := p.Config.DrawCut
		pAway := normCDF(-c - diff)
		pHome := 1.0 - normCDF(c-diff)
		pDraw := clamp(1.0-pHome-pAway, 0.01, 0.99)
		sum := pHome + pDraw + pAway
		return baseline.Probabilities{
			HomeWin: pHome / sum,
			Draw:    pDraw / sum,
			AwayWin: pAway / sum,
		}
	}

	homeStrength := math.Pow(10, (homeRating+p.Config.HomeAdvantage)/p.Config.Scale)
	awayStrength := math.Pow(10, awayRating/p.Config.Scale)
	drawStrength := p.Config.DrawFactor * math.Sqrt(homeStrength*awayStrength)
	denom := homeStrength + awayStrength + drawStrength
	return baseline.Probabilities{
		HomeWin: homeStrength / denom,
		Draw:    drawStrength / denom,
		AwayWin: awayStrength / denom,
	}
}

func (p *Pipeline) expectedGoals(homeRating, awayRating float64, probs baseline.Probabilities) (float64, float64) {
	baseHome := homeRating / 1000.0
	baseAway := awayRating / 1000.0
	homeMult := probs.HomeWin*1.5 + probs.Draw*1.05 + probs.AwayWin*0.6
	awayMult := probs.HomeWin*0.6 + probs.Draw*1.05 + probs.AwayWin*1.5
	return baseHome * homeMult, baseAway * awayMult
}

func (p *Pipeline) expectedTeamEvents(homeRating, awayRating float64, probs baseline.Probabilities, homeTeam, awayTeam string) []TeamEventSummary {
	homeMult := homeRating / 1500.0
	awayMult := awayRating / 1500.0
	return []TeamEventSummary{
		{
			Team:          homeTeam,
			Shots:         math.Round(12*homeMult*10) / 10,
			ShotsOnTarget: math.Round(4.5*homeMult*10) / 10,
			Corners:       math.Round(5.5*homeMult*10) / 10,
			Fouls:         math.Round(11*homeMult*10) / 10,
			YellowCards:   math.Round(1.8*homeMult*10) / 10,
			RedCards:      math.Round(0.1*homeMult*10) / 10,
		},
		{
			Team:          awayTeam,
			Shots:         math.Round(10*awayMult*10) / 10,
			ShotsOnTarget: math.Round(3.5*awayMult*10) / 10,
			Corners:       math.Round(4.5*awayMult*10) / 10,
			Fouls:         math.Round(10*awayMult*10) / 10,
			YellowCards:   math.Round(1.5*awayMult*10) / 10,
			RedCards:      math.Round(0.0*10) / 10,
		},
	}
}

func (p *Pipeline) getFormSnapshot(team string) *TeamFormSnapshot {
	f, ok := p.TeamForm[team]
	if !ok {
		return nil
	}
	return &TeamFormSnapshot{
		Team:          f.Team,
		RecentWins:    f.RecentWins,
		RecentDraws:   f.RecentDraws,
		RecentLosses:  f.RecentLosses,
		PointsPerGame: f.PointsPerGame,
		GoalDiff:      f.GoalDiffPerGame,
		RecentGF:      f.RecentGF,
		RecentGA:      f.RecentGA,
	}
}

func argmax(probs baseline.Probabilities) string {
	result := "H"
	best := probs.HomeWin
	if probs.Draw > best {
		result = "D"
		best = probs.Draw
	}
	if probs.AwayWin > best {
		result = "A"
	}
	return result
}

func normCDF(x float64) float64 {
	return 0.5 * math.Erfc(-x/math.Sqrt2)
}

func clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func limitSlice[T any](s []T, n int) []T {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func teamEventsToRichStats(evt TeamEventSummary) RichTeamStats {
	return RichTeamStats{
		Team:          evt.Team,
		Shots:         int(evt.Shots),
		ShotsOnTarget: int(evt.ShotsOnTarget),
		Fouls:         int(evt.Fouls),
		Corners:       int(evt.Corners),
		YellowCards:   int(evt.YellowCards),
		RedCards:      int(evt.RedCards),
	}
}

var _ = fmt.Sprintf
