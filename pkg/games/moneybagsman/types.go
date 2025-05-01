package moneybagsman

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
    CurrentFreeSpinIndex  int     `json:"current_free_spin_index"`
    RemainingFreeSpins    int     `json:"remaining_free_spins"`
    TotalFreeSpinsAwarded int     `json:"total_free_spins_awarded"`
    FreeSpinMultiplier    int     `json:"free_spin_multiplier"`
    ScatterCount          int     `json:"scatter_count"`  // Count of scatters that triggered the free spins
    ClientID              string  `json:"client_id"`
    GameID                string  `json:"game_id"`
    PlayerID              string  `json:"player_id"`
}

// SpinResponse represents the response body for the /spin endpoint
type SpinResponse struct {
    Reels                 [][]string  `json:"reels"`
    WinAmount             float64     `json:"win_amount"`
    WinDetails            []WinDetail `json:"win_details"`
    ScatterCount          int         `json:"scatter_count"`
    FreeSpinTriggered     bool        `json:"free_spin_triggered"`
    FreeSpinRetriggered   bool        `json:"free_spin_retriggered"`
    IsFreeSpin            bool        `json:"is_free_spin"`
    RemainingFreeSpins    int         `json:"remaining_free_spins"`
    CurrentFreeSpinIndex  int         `json:"current_free_spin_index"`
    FreeSpinMultiplier    int         `json:"free_spin_multiplier"`
    TotalFreeSpinsAwarded int         `json:"total_free_spins_awarded"`
    MaxMultiplier         int         `json:"max_multiplier"`        // Maximum multiplier possible based on scatter count
    MultiplierIncrement   int         `json:"multiplier_increment"`  // How much the multiplier increases per spin
}