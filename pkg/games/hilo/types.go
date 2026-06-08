package hilo

// RNGModification represents a card change made by the RNG system
type RNGModification struct {
	Position     int    `json:"position"`      // Position where card was changed
	OriginalCard string `json:"original_card"` // Original card that was supposed to be there
	ForcedCard   string `json:"forced_card"`   // Card that was forced by RNG
}

// GameState represents the current state of the Hilo game
type GameState struct {
	Seed                      string            `json:"seed"`
	DeckHash                  string            `json:"deck_hash"`
	CurrentCard               string            `json:"current_card"`
	UnityCard                 string            `json:"unity_card,omitempty"` // Track if this was a Unity-specified card
	Position                  int               `json:"position"`
	BetAmount                 float64           `json:"bet_amount"`
	SkipsUsed                 int               `json:"skips_used"`
	SkipsRemaining            int               `json:"skips_remaining"`
	MaxSkips                  int               `json:"max_skips"`
	MultiplierModifier        float64           `json:"multiplier_modifier"`         // JDB: Modifier for current card multipliers
	PreviousWinningMultiplier float64           `json:"previous_winning_multiplier"` // JDB: Last winning multiplier
	GameHistory               []Card            `json:"game_history"`
	RNGModifications          []RNGModification `json:"rng_modifications,omitempty"` // Track RNG card modifications
	IsGameOver                bool              `json:"is_game_over"`
	FinalWin                  float64           `json:"final_win"`
}

// Card represents a playing card
type Card struct {
	Card     string `json:"card"`
	Value    int    `json:"value"`
	Position int    `json:"position"`
}

// BetOption represents each betting button in the UI
type BetOption struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Multiplier float64 `json:"multiplier"`
	IsEnabled  bool    `json:"is_enabled"`
}

// GuessResult represents the result of a guess
type GuessResult struct {
	Success        bool    `json:"success"`
	NextCard       string  `json:"next_card"`
	NextValue      int     `json:"next_value"`
	PayoutMultiple float64 `json:"payout_multiple"`
	TotalWinAmount float64 `json:"total_win_amount"` // Total amount player would win if they cash out now
	WasCorrect     bool    `json:"was_correct"`
	Forced         bool    `json:"forced"`
}

// StartGameRequest represents the request to start a new game
type StartGameRequest struct {
	ClientID  string  `json:"client_id"`
	GameID    string  `json:"game_id"`
	PlayerID  string  `json:"player_id"`
	BetID     string  `json:"bet_id"`
	BetAmount float64 `json:"bet_amount"`
	Card      string  `json:"card,omitempty"` // Optional card from Unity frontend
}

// StartGameResponse represents the response when starting a new game
type StartGameResponse struct {
	Status     string      `json:"status"`
	Message    string      `json:"message"`
	GameState  GameState   `json:"game_state"`
	BetOptions []BetOption `json:"bet_options"`
	Signature  string      `json:"signature"`
}

// GuessRequest represents the request to make a guess
type GuessRequest struct {
	ClientID  string    `json:"client_id"`
	GameID    string    `json:"game_id"`
	PlayerID  string    `json:"player_id"`
	BetID     string    `json:"bet_id"`
	GameState GameState `json:"game_state"`
	BetChoice string    `json:"bet_choice"` // "higher", "lower", "same", "higher_or_same", "lower_or_same"
	Signature string    `json:"signature"`
}

// GuessResponse represents the response after making a guess
type GuessResponse struct {
	Status      string      `json:"status"`
	Message     string      `json:"message"`
	GameState   GameState   `json:"game_state"`
	GuessResult GuessResult `json:"guess_result"`
	BetOptions  []BetOption `json:"bet_options"`
	Signature   string      `json:"signature"`
}

// SkipRequest represents the request to skip a card
type SkipRequest struct {
	ClientID  string    `json:"client_id"`
	GameID    string    `json:"game_id"`
	PlayerID  string    `json:"player_id"`
	BetID     string    `json:"bet_id"`
	GameState GameState `json:"game_state"`
	Signature string    `json:"signature"`
}

// SkipResponse represents the response after skipping a card
type SkipResponse struct {
	Status     string      `json:"status"`
	Message    string      `json:"message"`
	GameState  GameState   `json:"game_state"`
	BetOptions []BetOption `json:"bet_options"`
	Signature  string      `json:"signature"`
}

// CashoutRequest represents the request to cashout
type CashoutRequest struct {
	ClientID  string    `json:"client_id"`
	GameID    string    `json:"game_id"`
	PlayerID  string    `json:"player_id"`
	BetID     string    `json:"bet_id"`
	GameState GameState `json:"game_state"`
	Signature string    `json:"signature"`
}

// CashoutResponse represents the response after cashing out
type CashoutResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	FinalWin  float64   `json:"final_win"`
	GameState GameState `json:"game_state"`
}

// VerifyRequest represents the request to verify deck
type VerifyRequest struct {
	Seed string `json:"seed"`
}

// VerifyResponse represents the response for deck verification
type VerifyResponse struct {
	Status   string   `json:"status"`
	Seed     string   `json:"seed"`
	Deck     []string `json:"deck"`
	DeckHash string   `json:"deck_hash"`
}

// PreviewRequest represents the request to preview a card (pre-game)
type PreviewRequest struct {
	ClientID  string  `json:"client_id"`
	GameID    string  `json:"game_id"`
	PlayerID  string  `json:"player_id"`
	BetAmount float64 `json:"bet_amount"`
	Card      string  `json:"card,omitempty"` // Optional card from Unity frontend
}

// PreviewResponse represents the response for card preview (pre-game)
type PreviewResponse struct {
	Status      string      `json:"status"`
	Message     string      `json:"message"`
	CurrentCard string      `json:"current_card"`
	BetOptions  []BetOption `json:"bet_options"`
}

// PreviewSkipRequest represents the request to skip in pre-game phase
type PreviewSkipRequest struct {
	ClientID    string `json:"client_id"`
	GameID      string `json:"game_id"`
	PlayerID    string `json:"player_id"`
	CurrentCard string `json:"current_card"`
}

// PreviewSkipResponse represents the response for pre-game skip
type PreviewSkipResponse struct {
	Status      string      `json:"status"`
	Message     string      `json:"message"`
	CurrentCard string      `json:"current_card"`
	BetOptions  []BetOption `json:"bet_options"`
}
