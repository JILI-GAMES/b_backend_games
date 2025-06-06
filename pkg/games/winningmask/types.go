package winningmask

// Symbol type
type Symbol string

const (
    SymbolPurpleMask   Symbol = "PurpleMask"   
    SymbolOrangeMask   Symbol = "OrangeMask"   
    SymbolGreenMask    Symbol = "GreenMask"    
    SymbolYellowMask   Symbol = "YellowMask"   
    SymbolBlueMask     Symbol = "BlueMask"     
    SymbolA            Symbol = "A"            
    SymbolK            Symbol = "K"            
    SymbolQ            Symbol = "Q"            
    SymbolJ            Symbol = "J"            
    Symbol10           Symbol = "10"           
    SymbolWild         Symbol = "Wild"         
    SymbolBonus        Symbol = "Bonus"        
    SymbolMaskReel     Symbol = "MaskReel"     
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
    ClientID             string  `json:"client_id"`
    GameID               string  `json:"game_id"`
    PlayerID             string  `json:"player_id"`
    BetID                string  `json:"bet_id"`
}

// SpinResponse represents the response body for the /spin endpoint
type SpinResponse struct {
    Reels                [][]string  `json:"reels"`
    WinAmount            float64     `json:"win_amount"`
    WinDetails           []WinDetail `json:"win_details"`
    BonusCount           int         `json:"bonus_count"`
    BonusWinAmount       float64     `json:"bonus_win_amount"`
    BonusPositions       []Position  `json:"bonus_positions"`
    MaskReelCount        int         `json:"mask_reel_count"`
    MaskReelPositions    []Position  `json:"mask_reel_positions"`
    FreeSpinTriggered    bool        `json:"free_spin_triggered"`
    FreeSpinRetriggered  bool        `json:"free_spin_retriggered"`
    MaskReelTriggered    bool        `json:"mask_reel_triggered"`
    IsFreeSpin           bool        `json:"is_free_spin"`
    RemainingFreeSpins   int         `json:"remaining_free_spins"`
    CurrentFreeSpinIndex int         `json:"current_free_spin_index"`
    TotalFreeSpinsAwarded int        `json:"total_free_spins_awarded"`
    BetAmount            float64     `json:"bet_amount"`
    BetMultiplier        int         `json:"bet_multiplier"`
}

// MaskReelBonusRequest represents the request body for the mask reel bonus
type MaskReelBonusRequest struct {
    ClientID    string `json:"client_id"`
    GameID      string `json:"game_id"`
    PlayerID    string `json:"player_id"`
    BetID       string `json:"bet_id"`
    BetAmount   float64 `json:"bet_amount"`
}

// MaskReelBonusResponse represents the response body for the mask reel bonus
type MaskReelBonusResponse struct {
    Multiplier  int     `json:"multiplier"`
    WinAmount   float64 `json:"win_amount"`
}