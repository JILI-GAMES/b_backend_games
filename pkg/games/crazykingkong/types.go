package crazykingkong

// BoulderType represents the different types of boulders
type BoulderType string

const (
	BoulderGold  BoulderType = "gold"
	BoulderBlue  BoulderType = "blue"
	BoulderRed   BoulderType = "red"
	BoulderWhite BoulderType = "white"
)

// StoneType represents the different types of bonus stones
type StoneType string

const (
	StoneGold   StoneType = "gold"
	StoneSilver StoneType = "silver"
	StoneBronze StoneType = "bronze"
)

// CrushRequest represents the request body for the /crush endpoint
type CrushRequest struct {
	ClientID     string      `json:"client_id"`
	GameID       string      `json:"game_id"`
	PlayerID     string      `json:"player_id"`
	BetID        string      `json:"bet_id"`
	BetAmount    float64     `json:"bet_amount"`
	BoulderType  BoulderType `json:"boulder_type"`
}

// BonusGameRequest represents the request body for the /bonus endpoint
type BonusGameRequest struct {
	ClientID  string    `json:"client_id"`
	GameID    string    `json:"game_id"`
	PlayerID  string    `json:"player_id"`
	BetID     string    `json:"bet_id"`
	BetAmount float64   `json:"bet_amount"`
	StoneType StoneType `json:"stone_type"`
}

// CrushResponse represents the response body for the /crush endpoint
type CrushResponse struct {
	Status           string  `json:"status"`
	Message          string  `json:"message"`
	BoulderType      string  `json:"boulder_type"`
	BoulderBroken    bool    `json:"boulder_broken"`
	Multiplier       float64 `json:"multiplier"`
	WinAmount        float64 `json:"win_amount"`
	BonusTriggered   bool    `json:"bonus_triggered"`
	AvailableStones  []Stone `json:"available_stones,omitempty"`
}

// BonusGameResponse represents the response body for the /bonus endpoint
type BonusGameResponse struct {
	Status     string  `json:"status"`
	Message    string  `json:"message"`
	StoneType  string  `json:"stone_type"`
	Multiplier float64 `json:"multiplier"`
	WinAmount  float64 `json:"win_amount"`
}

// Stone represents a bonus stone option
type Stone struct {
	Type        StoneType `json:"type"`
	DisplayName string    `json:"display_name"`
}