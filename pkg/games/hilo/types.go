package hilo

// GameState represents the current state of the Hilo game
type GameState struct {
	Seed            string  `json:"seed"`
	DeckHash        string  `json:"deck_hash"`
	CurrentCard     string  `json:"current_card"`
	Position        int     `json:"position"`
	AccumulatedWin  float64 `json:"accumulated_win"`
	BetAmount       float64 `json:"bet_amount"`
	SkipsUsed       int     `json:"skips_used"`
	SkipsRemaining  int     `json:"skips_remaining"`
	MaxSkips        int     `json:"max_skips"`
	GameHistory     []Card  `json:"game_history"`
	IsGameOver      bool    `json:"is_game_over"`
	FinalWin        float64 `json:"final_win"`
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