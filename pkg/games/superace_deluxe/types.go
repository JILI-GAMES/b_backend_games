package superace_deluxe

// Card represents a single symbol on the grid
type Card struct {
    Name       string `json:"name"`
    Substitute string `json:"substitute"` // "" or "BIG_JOKER" or "LITTLE_JOKER"
    Golden     bool   `json:"golden"`
    Transformed bool  `json:"transformed"`
}

// GameState holds the current state of the game
type GameState struct {
    FreeSpins       int        `json:"freeSpins"`
    AmountWon       float64    `json:"amountWon"`
    ComboMultiplier int        `json:"comboMultiplier"`
    Cards           [5][4]Card `json:"cards"` // 5x4 grid
    Mode            string     `json:"mode"`  // "NORMAL" or "FREE"
    BetAmount       float64    `json:"betAmount"`
}

// SpinRequest represents the client request
type SpinRequest struct {
    Game struct {
        ID     string `json:"id"`
        Name   string `json:"name"`
        Mode   string `json:"mode"`
    } `json:"game"`
    BetAmount       float64    `json:"betAmount"`
    ClientID        string     `json:"clientId"`
    PlayerID        string     `json:"playerId"`
    Action          string     `json:"action"` // "SPIN" or "TRANSFORM"
    Cards           [5][4]Card `json:"cards,omitempty"`
    ComboMultiplier int        `json:"comboMultiplier,omitempty"`
    FreeSpins       int        `json:"freeSpins,omitempty"`	
}

// SpinResponse represents the server response
type SpinResponse struct {
    Status  int       `json:"status"`
    Message string    `json:"message"`
    Data    GameState `json:"data"`
}

// RNGRequest for external RNG API
type RNGRequest struct {
    ClientID        string  `json:"client_id"`
    GameID          string  `json:"game_id"`
    PlayerID        string  `json:"player_id"`
    RTP             float64 `json:"rtp"`
    PayoutMultiplier float64 `json:"payout_multiplier"`
    RequestSalt     string  `json:"request_salt"`
    BetAmount       float64 `json:"bet_amount"`
}

// RNGResponse from external RNG API
type RNGResponse struct {
    PrefOutcome string  `json:"pref_outcome"`
    WinAmount   float64 `json:"win_amount"`
    WinProb     float64 `json:"win_prob"`
}

//-----------------------------------------------------------------------------------//

// RowBasedCard represents the same card data but organized for row-based representation
type RowBasedCard struct {
    Name        string `json:"name"`
    Substitute  string `json:"substitute"`
    Golden      bool   `json:"golden"`
    Transformed bool   `json:"transformed"`
}

// RowBasedGameState represents the game state with cards organized by rows instead of reels
type RowBasedGameState struct {
    FreeSpins       int              `json:"freeSpins"`
    AmountWon       float64          `json:"amountWon"`
    ComboMultiplier int              `json:"comboMultiplier"`
    Cards           [4][5]RowBasedCard `json:"cards"` // 4x5 grid (row-based)
    Mode            string           `json:"mode"`
    BetAmount       float64          `json:"betAmount"`
}

// RowBasedSpinResponse represents the server response with row-based cards
type RowBasedSpinResponse struct {
    Status  int            `json:"status"`
    Message string         `json:"message"`
    Data    RowBasedGameState `json:"data"`
}

// RowBasedSpinRequest represents the client request with row-based cards
type RowBasedSpinRequest struct {
    Game struct {
        ID     string `json:"id"`
        Name   string `json:"name"`
        Mode   string `json:"mode"`
    } `json:"game"`
    BetAmount       float64         `json:"betAmount"`
    ClientID        string          `json:"clientId"`
    PlayerID        string          `json:"playerId"`
    Action          string          `json:"action"` // "SPIN" or "TRANSFORM"
    Cards           [4][5]RowBasedCard `json:"cards,omitempty"`
    ComboMultiplier int             `json:"comboMultiplier,omitempty"`
    FreeSpins       int             `json:"freeSpins,omitempty"`    
}

// ConvertToRowBased converts a reel-based GameState to a row-based RowBasedGameState
func (gs *GameState) ConvertToRowBased() RowBasedGameState {
    rowBased := RowBasedGameState{
        FreeSpins:       gs.FreeSpins,
        AmountWon:       gs.AmountWon,
        ComboMultiplier: gs.ComboMultiplier,
        Mode:            gs.Mode,
        BetAmount:       gs.BetAmount,
    }
    
    // Transform the 5x4 reel-based grid to a 4x5 row-based grid
    for row := 0; row < 4; row++ {
        for reel := 0; reel < 5; reel++ {
            rowBased.Cards[row][reel] = RowBasedCard{
                Name:        gs.Cards[reel][row].Name,
                Substitute:  gs.Cards[reel][row].Substitute,
                Golden:      gs.Cards[reel][row].Golden,
                Transformed: gs.Cards[reel][row].Transformed,
            }
        }
    }
    
    return rowBased
}

// ConvertToReelBased converts row-based cards to reel-based cards
func ConvertToReelBased(rowBasedCards [4][5]RowBasedCard) [5][4]Card {
    var reelBased [5][4]Card
    
    // Transform the 4x5 row-based grid to a 5x4 reel-based grid
    for row := 0; row < 4; row++ {
        for reel := 0; reel < 5; reel++ {
            reelBased[reel][row] = Card{
                Name:        rowBasedCards[row][reel].Name,
                Substitute:  rowBasedCards[row][reel].Substitute,
                Golden:      rowBasedCards[row][reel].Golden,
                Transformed: rowBasedCards[row][reel].Transformed,
            }
        }
    }
    
    return reelBased
}