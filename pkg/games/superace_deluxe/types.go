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

// RTPRequest for external RTP API
type RTPRequest struct {
    ClientID string `json:"client_id"`
    GameID   string `json:"game_id"`
    PlayerID string `json:"player_id"`
}

// RTPResponse from external RTP API
type RTPResponse struct {
    Data struct {
        GameBets string `json:"game_bets"`
        GameRTP  string `json:"game_rtp"`
        GameWins string `json:"game_wins"`
    } `json:"data"`
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