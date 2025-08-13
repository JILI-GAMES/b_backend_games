package funkykingkong

// Symbol represents a symbol on the reels
type Symbol string

const (
	SymbolKong    Symbol = "Kong"
	SymbolSun     Symbol = "Sun"
	SymbolPalm    Symbol = "Palm"
	SymbolCoconut Symbol = "Coconut"
	SymbolBanana  Symbol = "Banana"
	Symbol3BAR    Symbol = "3BAR"
	Symbol2BAR    Symbol = "2BAR"
	Symbol1BAR    Symbol = "1BAR"
)

// SpinRequest represents the request body for the /spin endpoint
type SpinRequest struct {
	ClientID  string  `json:"client_id"`
	GameID    string  `json:"game_id"`
	PlayerID  string  `json:"player_id"`
	BetID     string  `json:"bet_id"`
	BetAmount float64 `json:"bet_amount"`
	BetLevel  int     `json:"bet_level"` // 1, 2, or 3 (paytable selection)
}

// SpinResponse represents the response body for the /spin endpoint
type SpinResponse struct {
	Status             string   `json:"status"`
	Message            string   `json:"message"`
	Reels              []string `json:"reels"`
	WinAmount          float64  `json:"win_amount"`
	WinningCombination string   `json:"winning_combination"`
	PaytableUsed       int      `json:"paytable_used"`
	BetLevel           int      `json:"bet_level"`
}
