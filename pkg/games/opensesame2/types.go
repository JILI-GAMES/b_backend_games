package opensesame2

// Symbol type
type Symbol string

const (
    SymbolWoman       Symbol = "Woman"
    SymbolMan         Symbol = "Man"
    SymbolHorse       Symbol = "Horse"
    SymbolPottery     Symbol = "Pottery"
    SymbolA           Symbol = "A"
    SymbolK           Symbol = "K"
    SymbolQ           Symbol = "Q"
    SymbolJ           Symbol = "J"
    Symbol10          Symbol = "10"
    Symbol9           Symbol = "9"
    SymbolScatter     Symbol = "Scatter"
    SymbolWild        Symbol = "Wild"
    SymbolFreeSpins   Symbol = "FreeSpins"
    SymbolMysteryBox  Symbol = "MysteryBox"
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
    BetAmount            float64 `json:"bet_amount"`
    IsFreeSpin           bool    `json:"is_free_spin"`
    CurrentFreeSpinIndex int     `json:"current_free_spin_index"`
    RemainingFreeSpins   int     `json:"remaining_free_spins"`
    TotalFreeSpinsAwarded int    `json:"total_free_spins_awarded"`
    FreeSpinMultiplier   int     `json:"free_spin_multiplier"`
    ExtraFreeSpinMultiplier int  `json:"extra_free_spin_multiplier"` 
    IsExtraFreeSpin      bool    `json:"is_extra_free_spin"`        
    ClientID             string  `json:"client_id"`
    GameID               string  `json:"game_id"`
    PlayerID             string  `json:"player_id"`
    BetID                string  `json:"bet_id"`
}

// SpinResponse represents the response body for the /spin endpoint
type SpinResponse struct {
    Reels                 [][]string  `json:"reels"`
    WinAmount             float64     `json:"win_amount"`
    WinDetails            []WinDetail `json:"win_details"`
    ScatterCount          int         `json:"scatter_count"`
    ScatterWinAmount      float64     `json:"scatter_win_amount"`
    ScatterPositions      []Position  `json:"scatter_positions"`
    FreeSpinCount         int         `json:"free_spin_count"`
    FreeSpinWinAmount     float64     `json:"free_spin_win_amount"`
    FreeSpinPositions     []Position  `json:"free_spin_positions"`
    MysteryBoxCount       int         `json:"mystery_box_count"`     
    MysteryBoxPositions   []Position  `json:"mystery_box_positions"` 
    FreeSpinTriggered     bool        `json:"free_spin_triggered"`
    FreeSpinRetriggered   bool        `json:"free_spin_retriggered"`
    ExtraFreeSpinTriggered bool       `json:"extra_free_spin_triggered"` 
    IsFreeSpin            bool        `json:"is_free_spin"`
    IsExtraFreeSpin       bool        `json:"is_extra_free_spin"`        
    RemainingFreeSpins    int         `json:"remaining_free_spins"`
    CurrentFreeSpinIndex  int         `json:"current_free_spin_index"`
    FreeSpinMultiplier    int         `json:"free_spin_multiplier"`
    ExtraFreeSpinMultiplier int       `json:"extra_free_spin_multiplier"` 
    TotalFreeSpinsAwarded int         `json:"total_free_spins_awarded"`
    BetAmount             float64     `json:"bet_amount"`
    BetMultiplier         int         `json:"bet_multiplier"`
}

// SelectOptionRequest represents the request body for the /select-free-spin-option endpoint
type SelectOptionRequest struct {
    ChestIndex    int    `json:"chest_index"`
    LampIndex     int    `json:"lamp_index"`
    TreasureIndex int    `json:"treasure_index"` 
    IsExtraBonus  bool   `json:"is_extra_bonus"` 
    ClientID      string `json:"client_id"`
    GameID        string `json:"game_id"`
    PlayerID      string `json:"player_id"`
    BetID         string `json:"bet_id"`
}

// SelectOptionResponse represents the response body for the /select-free-spin-option endpoint
type SelectOptionResponse struct {
    FreeSpinCount        int  `json:"free_spin_count"`
    FreeSpinMultiplier   int  `json:"free_spin_multiplier"`
    ExtraMultiplier      int  `json:"extra_multiplier"`      
    ExtraFreeSpins       int  `json:"extra_free_spins"`      
    IsExtraMultiplier    bool `json:"is_extra_multiplier"`   
}