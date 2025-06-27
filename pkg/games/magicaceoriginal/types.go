package magicaceoriginal

import "fmt"

// Symbol type
type Symbol string

const (
    SymbolACE       Symbol = "Ace"
    SymbolKING      Symbol = "King"
    SymbolQUEEN     Symbol = "Queen"
    SymbolJACK      Symbol = "Jack"
    SymbolHeart     Symbol = "Heart"
    SymbolSpade     Symbol = "Spade"
    SymbolClub      Symbol = "Club"
    SymbolDiamond   Symbol = "Diamond"
    SymbolTarget    Symbol = "target"
    SymbolWild      Symbol = "wild"
)

// Joker Card modes
const (
    ModeSuperJoker = "super_joker"
    ModeBigJoker   = "big_joker"
    ModeSmallJoker = "small_joker"
)

// Position represents a position on the reel grid
type Position struct {
    Reel int `json:"reel"`
    Row  int `json:"row"`
}

// JokerCard represents a Joker Card on the reels
type JokerCard struct {
    Position Position `json:"position"`
    Mode     string   `json:"mode"` // "super_joker", "big_joker", "small_joker"
}

// ValidateMode validates the JokerCard's mode
func (jc *JokerCard) ValidateMode() error {
    switch jc.Mode {
    case ModeSuperJoker, ModeBigJoker, ModeSmallJoker:
        return nil
    default:
        return fmt.Errorf("invalid joker card mode: %s", jc.Mode)
    }
}

// WinDetail represents a single winning combination
type WinDetail struct {
    Symbols     []string   `json:"symbols"`
    Payline     []Position `json:"payline"`
    Payout      float64    `json:"payout"`
    GoldenCards []Position `json:"goldenCards"` // Track Golden Cards in the win
}

// SpecialSymbols tracks special symbols on the reels
type SpecialSymbols struct {
    GoldenCards   []Position  `json:"goldenCards"`
    JokerCards    []JokerCard `json:"jokerCards"`
    TargetSymbols []Position  `json:"targetSymbols"`   // All Target symbols
}

// BaseGameCollector tracks the base game collector system (collect 5 → activate for 3 rounds)
type BaseGameCollector struct {
    CollectedCount  int  `json:"collectedCount"`  // Super Jokers collected (0-5)
    IsActive        bool `json:"isActive"`        // Whether feature is active (only when collected 5)
    RemainingRounds int  `json:"remainingRounds"` // Rounds left (3 when activated, decreases on wins)
}

// FreeSpinsCollector tracks the free spins collector system (immediate +1 per Super Joker)
type FreeSpinsCollector struct {
    UpgradeCount int `json:"upgradeCount"` // Current upgrade level (0-10), adds +N to all multipliers
    MaxUpgrades  int `json:"maxUpgrades"`  // Maximum upgrades allowed (10 without extra bet, unlimited with)
}

// GameState represents the current state of the game
type GameState struct {
    Bet struct {
        Amount          float64 `json:"amount"`
        Multiplier      int     `json:"multiplier"`
        ExtraBetEnabled bool    `json:"extraBetEnabled"`
    } `json:"bet"`
    GameMode         string    `json:"gameMode"` // "base" or "freeSpins"
    FreeSpins        struct {
        Remaining     int `json:"remaining"`
        TotalAwarded  int `json:"totalAwarded"`
        Retriggers    int `json:"retriggers"`
    } `json:"freeSpins"`
    Reels                [][]string         `json:"reels"` // 5x4 grid
    JokerCards           []JokerCard        `json:"jokerCards"`
    BoomingMultiplier    int                `json:"boomingMultiplier"`
    BaseGameCollector    BaseGameCollector  `json:"baseGameCollector"`
    FreeSpinsCollector   FreeSpinsCollector `json:"freeSpinsCollector"`
    TotalWin             float64            `json:"totalWin"`
    Cascading            bool               `json:"cascading"`
    TargetCount          int                `json:"targetCount"`
    LastWinDetails       []WinDetail        `json:"lastWinDetails"`
    SpecialSymbols       SpecialSymbols     `json:"specialSymbols"`
    CascadeCount         int                `json:"cascadeCount"`
}

// SpinRequest represents the request body for the /spin endpoint
type SpinRequest struct {
    GameState GameState `json:"gameState"`
    ClientID  string    `json:"client_id"`
    GameID    string    `json:"game_id"`
    PlayerID  string    `json:"player_id"`
    BetID     string    `json:"bet_id"`
}

// CascadeRequest represents the request body for the /cascade endpoint
type CascadeRequest struct {
    GameState GameState `json:"gameState"`
    ClientID  string    `json:"client_id"`
    GameID    string    `json:"game_id"`
    PlayerID  string    `json:"player_id"`
    BetID     string    `json:"bet_id"`
}

// FeatureBuyRequest represents the request body for the /featureBuy endpoint
type FeatureBuyRequest struct {
    GameState GameState `json:"gameState"`
    ClientID  string    `json:"client_id"`
    GameID    string    `json:"game_id"`
    PlayerID  string    `json:"player_id"`
    BetID     string    `json:"bet_id"`
}

// SpinResponse represents the response body for the /spin endpoint
type SpinResponse struct {
    Status     string      `json:"status"`
    Message    string      `json:"message"`
    GameState  GameState   `json:"gameState"`
    WinDetails []WinDetail `json:"winDetails"`
    TotalCost  float64     `json:"totalCost"`
}

// CascadeResponse represents the response body for the /cascade endpoint
type CascadeResponse struct {
    Status     string      `json:"status"`
    Message    string      `json:"message"`
    GameState  GameState   `json:"gameState"`
    WinDetails []WinDetail `json:"winDetails"`
    TotalCost  float64     `json:"totalCost"`
}

// FeatureBuyResponse represents the response body for the /featureBuy endpoint
type FeatureBuyResponse struct {
    Status     string      `json:"status"`
    Message    string      `json:"message"`
    GameState  GameState   `json:"gameState"`
    WinDetails []WinDetail `json:"winDetails"`
    TotalCost  float64     `json:"totalCost"`
}