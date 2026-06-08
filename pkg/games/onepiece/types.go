package onepiece

// Symbol type
type Symbol string

const (
	SymbolAirplane   Symbol = "Airplane"
	SymbolYacht      Symbol = "Yacht"
	SymbolCar        Symbol = "Car"
	SymbolMotorcycle Symbol = "Motorcycle"
	SymbolA          Symbol = "A"
	SymbolK          Symbol = "K"
	SymbolQ          Symbol = "Q"
	SymbolJ          Symbol = "J"
	SymbolScatter    Symbol = "Scatter"
	SymbolWild       Symbol = "Wild"
)

// Position represents a position on the reel grid
type Position struct {
	Reel int `json:"reel"`
	Row  int `json:"row"`
}

// WinDetail represents a single winning combination
type WinDetail struct {
	Symbol    string     `json:"symbol"`
	Count     int        `json:"count"`
	Payout    float64    `json:"payout"`
	Positions []Position `json:"positions"`
}

// SpinRequest represents the request body for the /spin endpoint
type SpinRequest struct {
	BetAmount             float64 `json:"bet_amount"`
	IsFreeSpin            bool    `json:"is_free_spin"`
	FreeSpinOption        int     `json:"free_spin_option"`
	CurrentFreeSpinIndex  int     `json:"current_free_spin_index"`
	RemainingFreeSpins    int     `json:"remaining_free_spins"`
	TotalFreeSpinsAwarded int     `json:"total_free_spins_awarded"`
	FreeSpinMultiplier    int     `json:"free_spin_multiplier"`
	ExtraBonusMultiplier  int     `json:"extra_bonus_multiplier"` // Added field to track extra bonus multiplier
	ClientID              string  `json:"client_id"`
	GameID                string  `json:"game_id"`
	PlayerID              string  `json:"player_id"`
	BetID                 string  `json:"bet_id"`
}

// SpinResponse represents the response body for the /spin endpoint
type SpinResponse struct {
	Reels                 [][]string  `json:"reels"`
	WinAmount             float64     `json:"win_amount"`
	WinDetails            []WinDetail `json:"win_details"`
	ScatterCount          int         `json:"scatter_count"`
	ScatterPositions      []Position  `json:"scatter_positions"`
	FreeSpinTriggered     bool        `json:"free_spin_triggered"`
	FreeSpinRetriggered   bool        `json:"free_spin_retriggered"`
	ExtraBonusMultiplier  int         `json:"extra_bonus_multiplier"`
	IsFreeSpin            bool        `json:"is_free_spin"`
	RemainingFreeSpins    int         `json:"remaining_free_spins"`
	CurrentFreeSpinIndex  int         `json:"current_free_spin_index"`
	FreeSpinMultiplier    int         `json:"free_spin_multiplier"`
	TotalFreeSpinsAwarded int         `json:"total_free_spins_awarded"`
	FreeSpinOption        int         `json:"free_spin_option"`
	WalletBalance         float64     `json:"wallet_balance"`
}

// SelectFreeSpinOptionRequest represents the request body for the /select-free-spin-option endpoint
type SelectFreeSpinOptionRequest struct {
	Option               int    `json:"option"`
	ExtraBonusMultiplier int    `json:"extra_bonus_multiplier"`
	ClientID             string `json:"client_id"`
	GameID               string `json:"game_id"`
	PlayerID             string `json:"player_id"`
	BetID                string `json:"bet_id"`
}

// SelectFreeSpinOptionResponse represents the response body for the /select-free-spin-option endpoint
type SelectFreeSpinOptionResponse struct {
	TotalFreeSpins       int `json:"total_free_spins"`
	InitialMultiplier    int `json:"initial_multiplier"`
	MaxFreeSpins         int `json:"max_free_spins"`
	ExtraBonusMultiplier int `json:"extra_bonus_multiplier"` // Added field to include in response
}
