package baseline

import "math"

type NaiveFrequencyModel struct {
	Prediction string  `json:"prediction"`
	HomeRate   float64 `json:"home_rate"`
	DrawRate   float64 `json:"draw_rate"`
	AwayRate   float64 `json:"away_rate"`
}

func NewNaiveFrequency(matches []Match) NaiveFrequencyModel {
	var h, d, a int
	for _, m := range matches {
		switch m.Result {
		case "H":
			h++
		case "D":
			d++
		case "A":
			a++
		}
	}
	total := float64(len(matches))
	if total == 0 {
		return NaiveFrequencyModel{Prediction: "H"}
	}
	prediction := "H"
	best := h
	if d > best {
		prediction = "D"
		best = d
	}
	if a > best {
		prediction = "A"
	}
	return NaiveFrequencyModel{
		Prediction: prediction,
		HomeRate:   float64(h) / total,
		DrawRate:   float64(d) / total,
		AwayRate:   float64(a) / total,
	}
}

func EvaluateNaiveFrequency(matches []Match, model NaiveFrequencyModel) Metrics {
	var correct int
	var logLoss float64
	var brier float64
	for _, m := range matches {
		if model.Prediction == m.Result {
			correct++
		}
		prob := probabilityForResultNaive(model, m.Result)
		if prob > 0 {
			logLoss += -math.Log(max(prob, 1e-12))
		}
		brier += brierTerm(model, m.Result)
	}
	total := float64(len(matches))
	return Metrics{
		Matches:          len(matches),
		Correct:          correct,
		Accuracy:         float64(correct) / total,
		LogLoss:          logLoss / total,
		BrierScore:       brier / total,
		PredictedHomeWin: ternary(model.Prediction == "H", 1.0, 0.0),
		PredictedDraw:    ternary(model.Prediction == "D", 1.0, 0.0),
		PredictedAwayWin: ternary(model.Prediction == "A", 1.0, 0.0),
	}
}

func probabilityForResultNaive(model NaiveFrequencyModel, result string) float64 {
	switch result {
	case "H":
		return model.HomeRate
	case "D":
		return model.DrawRate
	case "A":
		return model.AwayRate
	}
	return 0
}

func brierTerm(model NaiveFrequencyModel, result string) float64 {
	h := model.HomeRate
	d := model.DrawRate
	a := model.AwayRate
	switch result {
	case "H":
		return (1-h)*(1-h) + d*d + a*a
	case "D":
		return h*h + (1-d)*(1-d) + a*a
	case "A":
		return h*h + d*d + (1-a)*(1-a)
	}
	return 0
}

func ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
