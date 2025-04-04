package kong

// SpinRequest represents the request body for the /spin endpoint
type SpinRequest struct {
	ClientID          string  `json:"client_id"`
	GameID            string  `json:"game_id"`
	PlayerID          string  `json:"player_id"`
	BetAmount         float64 `json:"bet_amount"`
	IsFreeSpin        bool    `json:"is_free_spin"`
	FreeSpinCount     int     `json:"free_spin_count"`
	BonusMultiplier   int     `json:"bonus_multiplier"`
	OriginalBetAmount float64 `json:"original_bet_amount"`
}

// SpinResponse represents the response body for the /spin endpoint
type SpinResponse struct {
	Status            string     `json:"status"`
	Message           string     `json:"message"`
	Reels             [][]string `json:"reels"`
	WinAmount         float64    `json:"win_amount"`
	IsFreeSpin        bool       `json:"is_free_spin"`
	FreeSpinCount     int        `json:"free_spin_count"`
	BonusMultiplier   int        `json:"bonus_multiplier"`
	FreeSpinTriggered bool       `json:"free_spin_triggered"`
}
