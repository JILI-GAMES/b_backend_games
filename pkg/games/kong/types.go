package kong

// WinningPosition represents a winning symbol position on the reels
type WinningPosition struct {
	Symbol   string `json:"symbol"`
	Reel     int    `json:"reel"`
	Row      int    `json:"row"`
	Count    int    `json:"count"`    // Number of consecutive symbols in this win
	Ways     int    `json:"ways"`     // Number of ways for this winning combination
	WinValue float64 `json:"win_value"` // Win amount for this specific win
}

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
	WinningPositions  []WinningPosition `json:"winning_positions"` 
}
