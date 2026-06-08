package winningmask

// Symbol type
type Symbol string

const (
    SymbolPurpleMask   Symbol = "PurpleMask"   // Purple female mask - 5:1000, 4:200, 3:50
    SymbolOrangeMask   Symbol = "OrangeMask"   // Orange female mask - 5:400, 4:150, 3:25
    SymbolGreenMask    Symbol = "GreenMask"    // Green demon mask - 5:400, 4:150, 3:25
    SymbolYellowMask   Symbol = "YellowMask"   // Yellow male mask - 5:200, 4:75, 3:20
    SymbolBlueMask     Symbol = "BlueMask"     // Blue male mask - 5:200, 4:75, 3:20
    SymbolA            Symbol = "A"            // A - 5:150, 4:50, 3:10
    SymbolK            Symbol = "K"            // K - 5:150, 4:50, 3:10
    SymbolQ            Symbol = "Q"            // Q - 5:100, 4:20, 3:5
    SymbolJ            Symbol = "J"            // J - 5:100, 4:20, 3:5
    Symbol10           Symbol = "10"           // 10 - 5:100, 4:20, 3:5
    SymbolWild         Symbol = "Wild"         // Wild symbol
    SymbolBonus        Symbol = "Bonus"        // Free Spin Bonus symbol
    SymbolMaskReel     Symbol = "MaskReel"     // Mask Reel Bonus symbol
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

// Stage1Result represents the result of the first stage (normal spin)
type Stage1Result struct {
    Reels      [][]string  `json:"reels"`
    WinAmount  float64     `json:"win_amount"`
    WinDetails []WinDetail `json:"win_details"`
}

// TransformationResult represents the result of mask transformation
type TransformationResult struct {
    MaskType     Symbol      `json:"mask_type"`
    Reels        [][]string  `json:"reels"`
    WinAmount    float64     `json:"win_amount"`
    WinDetails   []WinDetail `json:"win_details"`
}

// CombinedScenario represents a complete two-stage scenario
type CombinedScenario struct {
    Stage1Win       float64              `json:"stage1_win"`
    Stage2Win       float64              `json:"stage2_win"`
    TotalWin        float64              `json:"total_win"`
    Stage1Reels     [][]string           `json:"stage1_reels"`
    Stage2Reels     [][]string           `json:"stage2_reels,omitempty"`
    Stage1Details   []WinDetail          `json:"stage1_details"`
    Stage2Details   []WinDetail          `json:"stage2_details,omitempty"`
    MaskType        string               `json:"mask_type,omitempty"`
    HasTransform    bool                 `json:"has_transform"`
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
    // Stage 1 (Normal Spin) Results
    Stage1Reels           [][]string  `json:"stage1_reels"`
    Stage1WinAmount       float64     `json:"stage1_win_amount"`
    Stage1WinDetails      []WinDetail `json:"stage1_win_details"`
    
    // Stage 2 (Transformation) Results - only present if transformation occurred
    Stage2Reels           [][]string  `json:"stage2_reels,omitempty"`
    Stage2WinAmount       float64     `json:"stage2_win_amount,omitempty"`
    Stage2WinDetails      []WinDetail `json:"stage2_win_details,omitempty"`
    
    // Combined Results
    TotalWinAmount        float64     `json:"total_win_amount"`
    MaskTransformationUsed bool       `json:"mask_transformation_used"`
    SelectedMaskType      string      `json:"selected_mask_type,omitempty"`
    
    // Bonus/Feature Information
    BonusCount            int         `json:"bonus_count"`
    BonusWinAmount        float64     `json:"bonus_win_amount"`
    BonusPositions        []Position  `json:"bonus_positions"`
    MaskReelCount         int         `json:"mask_reel_count"`
    MaskReelPositions     []Position  `json:"mask_reel_positions"`
    FreeSpinTriggered     bool        `json:"free_spin_triggered"`
    FreeSpinRetriggered   bool        `json:"free_spin_retriggered"`
    MaskReelTriggered     bool        `json:"mask_reel_triggered"`
    
    // Game State
    IsFreeSpin            bool        `json:"is_free_spin"`
    RemainingFreeSpins    int         `json:"remaining_free_spins"`
    CurrentFreeSpinIndex  int         `json:"current_free_spin_index"`
    TotalFreeSpinsAwarded int         `json:"total_free_spins_awarded"`
    BetAmount             float64     `json:"bet_amount"`
    BetMultiplier         int         `json:"bet_multiplier"`
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