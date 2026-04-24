package domain

type MatchEventKind string

const (
	EventGoal         MatchEventKind = "goal"
	EventShot         MatchEventKind = "shot"
	EventCard         MatchEventKind = "card"
	EventSubstitution MatchEventKind = "substitution"
	EventAssist       MatchEventKind = "assist"
)

type MatchEvent struct {
	ID       string         `json:"id"`
	MatchID  string         `json:"match_id"`
	Minute   int            `json:"minute"`
	Stoppage int            `json:"stoppage"`
	Kind     MatchEventKind `json:"kind"`
	Goal     *Goal          `json:"goal,omitempty"`
	Shot     *Shot          `json:"shot,omitempty"`
	Card     *Card          `json:"card,omitempty"`
	Sub      *Substitution  `json:"sub,omitempty"`
}

type Goal struct {
	ScorerID   string `json:"scorer_id"`
	ScorerName string `json:"scorer_name"`
	GoalType   string `json:"goal_type"`
	IsOwnGoal  bool   `json:"is_own_goal"`
	IsPenalty  bool   `json:"is_penalty"`
}

type Shot struct {
	PlayerID string  `json:"player_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	XG       float64 `json:"xg"`
	Outcome  string  `json:"outcome"`
	BodyPart string  `json:"body_part"`
}

type Card struct {
	PlayerID string `json:"player_id"`
	CardType string `json:"card_type"`
	Reason   string `json:"reason"`
}

type Substitution struct {
	PlayerIn  string `json:"player_in"`
	PlayerOut string `json:"player_out"`
	Reason    string `json:"reason"`
}

type Assist struct {
	EventID    string `json:"event_id"`
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	AssistType string `json:"assist_type"`
}
