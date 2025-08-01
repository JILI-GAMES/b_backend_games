package hilo

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

// StartGameHandler handles the /start/hilo endpoint
func (rg *RouteGroup) StartGameHandler(c *fiber.Ctx) error {
	var req StartGameRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate bet amount
	if req.BetAmount <= 0 || req.BetAmount > 1000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid bet amount",
		})
	}

	// Round bet amount to 2 decimal places
	req.BetAmount = RoundToTwo(req.BetAmount)

	// Generate unique seed and deck
	seed := GenerateUniqueSeed()
	deck := GenerateDeck(seed)
	deckHash := HashDeck(deck)

	// Initialize game state
	gameState := GameState{
		Seed:           seed,
		DeckHash:       deckHash,
		CurrentCard:    deck[0],
		Position:       0,
		AccumulatedWin: 1.00, // Start with 1x multiplier
		BetAmount:      RoundToTwo(req.BetAmount),
		SkipsUsed:      0,
		SkipsRemaining: 5,
		MaxSkips:       5,
		GameHistory:    []Card{{Card: deck[0], Value: GetCardValue(deck[0]), Position: 0}},
		IsGameOver:     false,
		FinalWin:       0,
	}

	// Generate betting options for the starting card
	betOptions := GetHiloOptions(gameState.CurrentCard)

	// Generate signature
	signature := ComputeHMAC(seed, 0, gameState.AccumulatedWin, req.BetAmount)

	log.Printf("Started new Hilo game: seed=%s, currentCard=%s, betAmount=%.2f",
		seed, gameState.CurrentCard, req.BetAmount)

	return c.JSON(StartGameResponse{
		Status:     "success",
		Message:    "Game started successfully",
		GameState:  gameState,
		BetOptions: betOptions,
		Signature:  signature,
	})
}

// GuessHandler handles the /guess/hilo endpoint with RNG integration
func (rg *RouteGroup) GuessHandler(c *fiber.Ctx) error {
	rngClient, settingsClient := rg.getClientsForRequest(c)

	var req GuessRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Ensure monetary values are rounded
	req.GameState.BetAmount = RoundToTwo(req.GameState.BetAmount)
	req.GameState.AccumulatedWin = RoundToTwo(req.GameState.AccumulatedWin)

	// Verify signature
	if !VerifyHMAC(req.GameState.Seed, req.GameState.Position, req.GameState.AccumulatedWin, req.GameState.BetAmount, req.Signature) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid signature",
		})
	}

	// Validate bet choice
	currentValue := GetCardValue(req.GameState.CurrentCard)
	if !IsValidBetChoice(currentValue, req.BetChoice) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid bet choice for current card",
		})
	}

	// Check if game is over
	if req.GameState.IsGameOver {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Game is already over",
		})
	}

	// Regenerate deck and verify current card
	deck := GenerateDeck(req.GameState.Seed)
	if req.GameState.Position >= len(deck) || deck[req.GameState.Position] != req.GameState.CurrentCard {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid game state",
		})
	}

	// Check if we can draw next card
	if req.GameState.Position+1 >= len(deck) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "No more cards in deck",
		})
	}

	// Get pre-calculated multiplier for this bet
	payoutMultiplier := GetMultiplierForBet(currentValue, req.BetChoice)
	if payoutMultiplier == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid bet for current card",
		})
	}

	// Calculate potential win amount for RNG decision
	newAccumulated := RoundToTwo(req.GameState.AccumulatedWin * payoutMultiplier)
	totalWinAmount := RoundToTwo(newAccumulated * req.GameState.BetAmount)
	rngPayoutMultiplier := RoundToTwo(totalWinAmount / req.GameState.BetAmount)

	// Get RTP settings
	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if err != nil {
		log.Printf("Failed to get RTP: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve game settings",
		})
	}

	// CRITICAL: Call RNG to determine if house can afford this win
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("✅IP: %v", ip)
	log.Printf("✅User-Agent: %v", userAgent)
	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, rngPayoutMultiplier, req.GameState.BetAmount, ip, userAgent)
	if err != nil {
		log.Printf("Failed to call RNG API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to determine outcome",
		})
	}

	log.Printf("RNG Decision: %s for potential win of %.2f (multiplier: %.2fx)",
		rngResp.PrefOutcome, totalWinAmount, rngPayoutMultiplier)

	// Get the natural next card
	nextCard := deck[req.GameState.Position+1]
	naturalResult := CheckBetResult(currentValue, GetCardValue(nextCard), req.BetChoice)

	// Apply house edge control - override natural result if needed
	forced := false
	actualResult := naturalResult
	finalCard := nextCard

	if naturalResult && rngResp.PrefOutcome == "loss" {
		// Player would win naturally, but house can't afford it - FORCE LOSS
		remainingCards := deck[req.GameState.Position+1:]
		forcedCard := FindLosingCard(currentValue, req.BetChoice, remainingCards)
		finalCard = forcedCard
		actualResult = false
		forced = true
		log.Printf("RNG FORCED LOSS: Natural %s would win, forced %s instead (saving %.2f)",
			nextCard, forcedCard, totalWinAmount)
	} else if !naturalResult && rngResp.PrefOutcome == "win" {
		// Player would lose naturally, but house allows win (rare, for retention)
		remainingCards := deck[req.GameState.Position+1:]
		forcedCard := FindWinningCard(currentValue, req.BetChoice, remainingCards)
		if forcedCard != "" {
			finalCard = forcedCard
			actualResult = true
			forced = true
			log.Printf("RNG FORCED WIN: Natural %s would lose, forced %s for retention",
				nextCard, forcedCard)
		}
	}

	// Update game state
	newGameState := req.GameState
	newGameState.Position++
	newGameState.CurrentCard = finalCard

	// Add to history
	newGameState.GameHistory = append(newGameState.GameHistory, Card{
		Card:     finalCard,
		Value:    GetCardValue(finalCard),
		Position: newGameState.Position,
	})

	guessResult := GuessResult{
		NextCard:       finalCard,
		NextValue:      GetCardValue(finalCard),
		PayoutMultiple: payoutMultiplier,
		WasCorrect:     actualResult,
		Forced:         forced,
	}

	var newSignature string
	var betOptions []BetOption

	if actualResult {
		// WIN: Update accumulated win
		newGameState.AccumulatedWin = RoundToTwo(newAccumulated)
		newSignature = ComputeHMAC(newGameState.Seed, newGameState.Position, newGameState.AccumulatedWin, newGameState.BetAmount)
		guessResult.Success = true

		// Generate betting options for next round
		betOptions = GetHiloOptions(newGameState.CurrentCard)

		log.Printf("Player WON: accumulated=%.2f, multiplier=%.2f", newGameState.AccumulatedWin, payoutMultiplier)
	} else {
		// LOSS: Game over
		newGameState.AccumulatedWin = 0
		newGameState.IsGameOver = true
		newGameState.FinalWin = 0
		newSignature = ""
		guessResult.Success = false
		betOptions = []BetOption{} // No options when game is over

		log.Printf("Player LOST: game over")
	}

	return c.JSON(GuessResponse{
		Status:      "success",
		Message:     "",
		GameState:   newGameState,
		GuessResult: guessResult,
		BetOptions:  betOptions,
		Signature:   newSignature,
	})
}

// SkipHandler handles the /skip/hilo endpoint
func (rg *RouteGroup) SkipHandler(c *fiber.Ctx) error {
	var req SkipRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Ensure monetary values are rounded
	req.GameState.BetAmount = RoundToTwo(req.GameState.BetAmount)
	req.GameState.AccumulatedWin = RoundToTwo(req.GameState.AccumulatedWin)

	// Verify signature
	if !VerifyHMAC(req.GameState.Seed, req.GameState.Position, req.GameState.AccumulatedWin, req.GameState.BetAmount, req.Signature) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid signature",
		})
	}

	// Check if player can skip
	if req.GameState.SkipsRemaining <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "No skips remaining",
		})
	}

	// Check if game is over
	if req.GameState.IsGameOver {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Game is already over",
		})
	}

	// For first card (position 0), allow infinite skips
	// For other cards, check skip limit
	if req.GameState.Position > 0 && req.GameState.SkipsRemaining <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "No skips remaining",
		})
	}

	// Regenerate deck and verify
	deck := GenerateDeck(req.GameState.Seed)
	if req.GameState.Position >= len(deck) || deck[req.GameState.Position] != req.GameState.CurrentCard {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid game state",
		})
	}

	// Check if we can draw next card
	if req.GameState.Position+1 >= len(deck) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "No more cards in deck",
		})
	}

	// Skip to next card
	nextCard := deck[req.GameState.Position+1]

	// Update game state
	newGameState := req.GameState
	newGameState.Position++
	newGameState.CurrentCard = nextCard

	// Only count skips for cards after the first one
	if req.GameState.Position > 0 {
		newGameState.SkipsUsed++
		newGameState.SkipsRemaining--
	}

	// Add to history
	newGameState.GameHistory = append(newGameState.GameHistory, Card{
		Card:     nextCard,
		Value:    GetCardValue(nextCard),
		Position: newGameState.Position,
	})

	// Generate betting options for the new card
	betOptions := GetHiloOptions(newGameState.CurrentCard)

	// Generate new signature
	newSignature := ComputeHMAC(newGameState.Seed, newGameState.Position, newGameState.AccumulatedWin, newGameState.BetAmount)

	log.Printf("Player skipped: newCard=%s, skipsRemaining=%d", nextCard, newGameState.SkipsRemaining)

	return c.JSON(SkipResponse{
		Status:     "success",
		Message:    fmt.Sprintf("Skipped to next card: %s", nextCard),
		GameState:  newGameState,
		BetOptions: betOptions,
		Signature:  newSignature,
	})
}

// CashoutHandler handles the /cashout/hilo endpoint
func (rg *RouteGroup) CashoutHandler(c *fiber.Ctx) error {
	var req CashoutRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Ensure monetary values are rounded
	req.GameState.BetAmount = RoundToTwo(req.GameState.BetAmount)
	req.GameState.AccumulatedWin = RoundToTwo(req.GameState.AccumulatedWin)

	// Verify signature
	if !VerifyHMAC(req.GameState.Seed, req.GameState.Position, req.GameState.AccumulatedWin, req.GameState.BetAmount, req.Signature) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid signature",
		})
	}

	// Check if player can cash out
	if req.GameState.Position == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Cannot cash out before making any guesses",
		})
	}

	if req.GameState.IsGameOver {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Game is already over",
		})
	}

	// Calculate final win amount
	finalWin := RoundToTwo(req.GameState.AccumulatedWin * req.GameState.BetAmount)

	// Update game state
	newGameState := req.GameState
	newGameState.IsGameOver = true
	newGameState.FinalWin = finalWin

	log.Printf("Player cashed out: finalWin=%.2f, multiplier=%.2f", finalWin, req.GameState.AccumulatedWin)

	return c.JSON(CashoutResponse{
		Status:    "success",
		Message:   fmt.Sprintf("Cashed out successfully: %.2f", finalWin),
		FinalWin:  finalWin,
		GameState: newGameState,
	})
}

// VerifyDeckHandler handles the /verify/hilo endpoint
func (rg *RouteGroup) VerifyDeckHandler(c *fiber.Ctx) error {
	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if req.Seed == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Seed is required",
		})
	}

	// Regenerate deck and hash
	deck := GenerateDeck(req.Seed)
	deckHash := HashDeck(deck)

	return c.JSON(VerifyResponse{
		Status:   "success",
		Seed:     req.Seed,
		Deck:     deck,
		DeckHash: deckHash,
	})
}

// validateRequest validates common request fields
func validateRequest(clientID, gameID, playerID, betID string) error {
	if clientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if gameID == "" {
		return fmt.Errorf("game_id is required")
	}
	if playerID == "" {
		return fmt.Errorf("player_id is required")
	}
	if betID == "" {
		return fmt.Errorf("bet_id is required")
	}
	return nil
}
