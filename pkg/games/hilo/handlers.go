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

	// Determine starting card and generate appropriate deck
	var startingCard string
	var seed string
	var deck []string
	var deckHash string

	if req.Card != "" {
		// Use card from Unity frontend
		startingCard = req.Card
		log.Printf("Using card from Unity frontend: %s", startingCard)

		// Generate a deck that starts with the Unity-specified card
		seed = GenerateUniqueSeed()
		deck = GenerateDeckWithFirstCard(seed, startingCard)
		deckHash = HashDeck(deck)
		log.Printf("Generated deck with Unity card: firstCard=%s, deck[0]=%s", startingCard, deck[0])
	} else {
		// Generate random card and deck
		seed = GenerateUniqueSeed()
		deck = GenerateDeck(seed)
		startingCard = deck[0]
		deckHash = HashDeck(deck)
		log.Printf("Generated random card: %s", startingCard)
	}

	// Initialize game state
	gameState := GameState{
		Seed:                      seed,
		DeckHash:                  deckHash,
		CurrentCard:               startingCard,
		UnityCard:                 req.Card, // Track if this was a Unity-specified card
		Position:                  0,
		BetAmount:                 RoundToTwo(req.BetAmount),
		SkipsUsed:                 0,
		SkipsRemaining:            5,
		MaxSkips:                  5,
		MultiplierModifier:        1.0, // Start with 1.0 modifier
		PreviousWinningMultiplier: 1.0, // Start with 1.0 (no previous win)
		GameHistory:               []Card{{Card: startingCard, Value: GetCardValue(startingCard), Position: 0}},
		IsGameOver:                false,
		FinalWin:                  0,
	}

	// Generate betting options for the starting card (base multipliers)
	betOptions := GetBaseHiloOptions(gameState.CurrentCard)

	// Generate signature (use base multiplier for start)
	currentMultiplier := GetBaseMultiplierForBet(gameState.CurrentCard, "higher_or_same")
	signature := ComputeHMAC(seed, 0, currentMultiplier, req.BetAmount)

	cardSource := "Random"
	if req.Card != "" {
		cardSource = "Unity"
	}
	log.Printf("Started new Hilo game: seed=%s, currentCard=%s, betAmount=%.2f, cardSource=%s",
		seed, gameState.CurrentCard, req.BetAmount, cardSource)

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

	// Verify signature
	// For first guess (no previous wins), use base multiplier
	// For subsequent guesses, use progressive multiplier (same as what was used to generate the signature)
	var verificationMultiplier float64
	if req.GameState.MultiplierModifier == 1.0 && req.GameState.PreviousWinningMultiplier == 1.0 {
		// First guess - verify with base multiplier
		verificationMultiplier = GetBaseMultiplierForBet(req.GameState.CurrentCard, "higher_or_same")
	} else {
		// Subsequent guesses - verify with progressive multiplier (same as what generated the signature)
		verificationMultiplier = GetMultiplierForBet(req.GameState.CurrentCard, "higher_or_same", req.GameState.MultiplierModifier, req.GameState.PreviousWinningMultiplier)
	}
	if !VerifyHMAC(req.GameState.Seed, req.GameState.Position, verificationMultiplier, req.GameState.BetAmount, req.Signature) {
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
	var deck []string
	if req.GameState.UnityCard != "" {
		// Use the same deck generation logic as StartGameHandler for Unity cards
		deck = GenerateDeckWithFirstCard(req.GameState.Seed, req.GameState.UnityCard)
		log.Printf("Using Unity deck generation: unityCard=%s", req.GameState.UnityCard)
	} else {
		// Use regular deck generation for random cards
		deck = GenerateDeck(req.GameState.Seed)
	}

	// Apply RNG modifications if any
	if len(req.GameState.RNGModifications) > 0 {
		deck = ApplyRNGModifications(deck, req.GameState.RNGModifications)
		log.Printf("Applied %d RNG modifications to deck", len(req.GameState.RNGModifications))
	}

	// Verify deck hash matches
	regeneratedDeckHash := HashDeck(deck)
	log.Printf("Deck hash verification: expected=%s, actual=%s", req.GameState.DeckHash, regeneratedDeckHash)
	if regeneratedDeckHash != req.GameState.DeckHash {
		log.Printf("❌ Deck hash verification failed: expected=%s, actual=%s", req.GameState.DeckHash, regeneratedDeckHash)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid deck hash",
		})
	}

	log.Printf("Verifying game state: position=%d, currentCard=%s, deckCard=%s, deckLength=%d, seed=%s, unityCard=%s",
		req.GameState.Position, req.GameState.CurrentCard, deck[req.GameState.Position], len(deck), req.GameState.Seed, req.GameState.UnityCard)

	if req.GameState.Position >= len(deck) || deck[req.GameState.Position] != req.GameState.CurrentCard {
		log.Printf("❌ Game state verification failed: expected=%s, actual=%s", req.GameState.CurrentCard, deck[req.GameState.Position])
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

	// Get multiplier for this bet
	// For first guess (no previous wins), use base multiplier
	// For subsequent guesses, use JDB progressive formula
	var payoutMultiplier float64
	if req.GameState.MultiplierModifier == 1.0 && req.GameState.PreviousWinningMultiplier == 1.0 {
		// First guess - use base multiplier
		payoutMultiplier = GetBaseMultiplierForBet(req.GameState.CurrentCard, req.BetChoice)
	} else {
		// Subsequent guesses - use JDB progressive formula
		payoutMultiplier = GetMultiplierForBet(req.GameState.CurrentCard, req.BetChoice, req.GameState.MultiplierModifier, req.GameState.PreviousWinningMultiplier)
	}
	if payoutMultiplier == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid bet for current card",
		})
	}

	// Calculate potential win amount for RNG decision (JDB style: current multiplier × bet amount)
	totalWinAmount := RoundToTwo(payoutMultiplier * req.GameState.BetAmount)
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
		// Update deck with forced card
		deck[req.GameState.Position+1] = forcedCard
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
			// Update deck with forced card
			deck[req.GameState.Position+1] = forcedCard
			log.Printf("RNG FORCED WIN: Natural %s would lose, forced %s for retention",
				nextCard, forcedCard)
		}
	}

	// Update game state
	newGameState := req.GameState
	newGameState.Position++
	newGameState.CurrentCard = finalCard

	// Update deck hash if deck was modified by RNG
	if forced {
		newGameState.DeckHash = HashDeck(deck)
		// Track RNG modification
		rngMod := RNGModification{
			Position:     req.GameState.Position + 1,
			OriginalCard: nextCard,
			ForcedCard:   finalCard,
		}
		newGameState.RNGModifications = append(req.GameState.RNGModifications, rngMod)
		log.Printf("Updated deck hash due to RNG modification: %s", newGameState.DeckHash)
	}

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
		TotalWinAmount: totalWinAmount, // Total amount player would win if they cash out now
		WasCorrect:     actualResult,
		Forced:         forced,
	}

	var newSignature string
	var betOptions []BetOption

	if actualResult {
		// WIN: Update JDB multipliers
		newGameState.PreviousWinningMultiplier = payoutMultiplier // Store current winning multiplier
		newGameState.MultiplierModifier = payoutMultiplier        // Update modifier for next card

		// Normal win - continue game
		// Generate signature using the same logic as verification
		var signatureMultiplier float64
		if newGameState.MultiplierModifier == 1.0 && newGameState.PreviousWinningMultiplier == 1.0 {
			// First guess - use base multiplier for signature
			signatureMultiplier = GetBaseMultiplierForBet(newGameState.CurrentCard, "higher_or_same")
		} else {
			// Subsequent guesses - use progressive multiplier for signature
			signatureMultiplier = GetMultiplierForBet(newGameState.CurrentCard, "higher_or_same", newGameState.MultiplierModifier, newGameState.PreviousWinningMultiplier)
		}
		newSignature = ComputeHMAC(newGameState.Seed, newGameState.Position, signatureMultiplier, newGameState.BetAmount)
		guessResult.Success = true

		// Generate betting options for next round with JDB multipliers
		betOptions = GetHiloOptions(newGameState.CurrentCard, newGameState.Position, newGameState.MultiplierModifier, newGameState.PreviousWinningMultiplier)

		log.Printf("Player WON: current multiplier=%.2f, previous winning=%.2f", payoutMultiplier, newGameState.PreviousWinningMultiplier)
	} else {
		// LOSS: Game over, reset multipliers but show next card with base multipliers
		newGameState.PreviousWinningMultiplier = 1.0 // Reset to base
		newGameState.MultiplierModifier = 1.0        // Reset to base
		newGameState.IsGameOver = true
		newGameState.FinalWin = 0
		guessResult.Success = false
		guessResult.TotalWinAmount = 0.0 // Player loses, so total win amount is 0

		// Show next card with base multipliers for next game
		betOptions = GetBaseHiloOptions(newGameState.CurrentCard)
		newSignature = ""
		// Generate signature for next game (using base multipliers)
		// newSignature = ComputeHMAC(newGameState.Seed, newGameState.Position, GetBaseMultiplierForBet(newGameState.CurrentCard, "higher_or_same"), newGameState.BetAmount)

		log.Printf("Player LOST: game over, showing next card with base multipliers for new game")
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

	// Verify signature
	// For first guess (no previous wins), use base multiplier
	// For subsequent guesses, use progressive multiplier (same as what was used to generate the signature)
	var verificationMultiplier float64
	if req.GameState.MultiplierModifier == 1.0 && req.GameState.PreviousWinningMultiplier == 1.0 {
		// First guess - verify with base multiplier
		verificationMultiplier = GetBaseMultiplierForBet(req.GameState.CurrentCard, "higher_or_same")
	} else {
		// Subsequent guesses - verify with progressive multiplier (same as what generated the signature)
		verificationMultiplier = GetMultiplierForBet(req.GameState.CurrentCard, "higher_or_same", req.GameState.MultiplierModifier, req.GameState.PreviousWinningMultiplier)
	}
	if !VerifyHMAC(req.GameState.Seed, req.GameState.Position, verificationMultiplier, req.GameState.BetAmount, req.Signature) {
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
	var deck []string
	if req.GameState.UnityCard != "" {
		// Use the same deck generation logic as StartGameHandler for Unity cards
		deck = GenerateDeckWithFirstCard(req.GameState.Seed, req.GameState.UnityCard)
	} else {
		// Use regular deck generation for random cards
		deck = GenerateDeck(req.GameState.Seed)
	}

	// Apply RNG modifications if any
	if len(req.GameState.RNGModifications) > 0 {
		deck = ApplyRNGModifications(deck, req.GameState.RNGModifications)
		log.Printf("Applied %d RNG modifications to deck", len(req.GameState.RNGModifications))
	}

	// Verify deck hash matches
	regeneratedDeckHash := HashDeck(deck)
	if regeneratedDeckHash != req.GameState.DeckHash {
		log.Printf("❌ Deck hash verification failed: expected=%s, actual=%s", req.GameState.DeckHash, regeneratedDeckHash)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid deck hash",
		})
	}

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

	// Generate betting options for the new card (maintain JDB multipliers)
	betOptions := GetHiloOptions(newGameState.CurrentCard, newGameState.Position, newGameState.MultiplierModifier, newGameState.PreviousWinningMultiplier)

	// Generate new signature (JDB: use current multiplier)
	newCurrentMultiplier := GetMultiplierForBet(newGameState.CurrentCard, "higher_or_same", newGameState.MultiplierModifier, newGameState.PreviousWinningMultiplier)
	newSignature := ComputeHMAC(newGameState.Seed, newGameState.Position, newCurrentMultiplier, newGameState.BetAmount)

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

	// Verify signature
	// For first guess (no previous wins), use base multiplier
	// For subsequent guesses, use progressive multiplier (same as what was used to generate the signature)
	var verificationMultiplier float64
	if req.GameState.MultiplierModifier == 1.0 && req.GameState.PreviousWinningMultiplier == 1.0 {
		// First guess - verify with base multiplier
		verificationMultiplier = GetBaseMultiplierForBet(req.GameState.CurrentCard, "higher_or_same")
	} else {
		// Subsequent guesses - verify with progressive multiplier (same as what generated the signature)
		verificationMultiplier = GetMultiplierForBet(req.GameState.CurrentCard, "higher_or_same", req.GameState.MultiplierModifier, req.GameState.PreviousWinningMultiplier)
	}
	if !VerifyHMAC(req.GameState.Seed, req.GameState.Position, verificationMultiplier, req.GameState.BetAmount, req.Signature) {
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

	// Calculate final win amount (JDB: use multiplier_modifier × bet amount)
	// Apply 1000x cap for cashout

	multiplier := req.GameState.MultiplierModifier
	if multiplier > MaxMultiplier {
		multiplier = MaxMultiplier
		log.Printf("CASHOUT CAP APPLIED: Original multiplier %.2f capped at %.0fx", req.GameState.MultiplierModifier, MaxMultiplier)
	} else {
		multiplier = req.GameState.MultiplierModifier
	}

	finalWin := RoundToTwo(multiplier * req.GameState.BetAmount)

	// Update game state
	newGameState := req.GameState
	newGameState.IsGameOver = true
	newGameState.FinalWin = finalWin
	newGameState.MultiplierModifier = multiplier
	newGameState.PreviousWinningMultiplier = multiplier

	log.Printf("Player cashed out: finalWin=%.2f, multiplier=%.2f (capped from %.2f)", finalWin, multiplier, req.GameState.MultiplierModifier)

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

// PreviewHandler handles the /preview/hilo endpoint (pre-game phase)
func (rg *RouteGroup) PreviewHandler(c *fiber.Ctx) error {
	var req PreviewRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request (no bet_id required for preview)
	if req.ClientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "client_id is required",
		})
	}
	if req.GameID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "game_id is required",
		})
	}
	if req.PlayerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "player_id is required",
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

	// Determine starting card
	var currentCard string

	if req.Card != "" {
		// Use card from Unity frontend
		currentCard = req.Card
		log.Printf("Preview: Using card from Unity frontend: %s", currentCard)
	} else {
		// Generate random card for preview
		seed := GenerateUniqueSeed()
		deck := GenerateDeck(seed)
		currentCard = deck[0]
		log.Printf("Preview: Generated random card: %s", currentCard)
	}

	// Generate betting options for the card (base multipliers for preview)
	betOptions := GetBaseHiloOptions(currentCard)

	cardSource := "Random"
	if req.Card != "" {
		cardSource = "Unity"
	}
	log.Printf("Preview: Showing card %s (source: %s) with bet amount %.2f",
		currentCard, cardSource, req.BetAmount)

	return c.JSON(PreviewResponse{
		Status:      "success",
		Message:     "Card preview ready",
		CurrentCard: currentCard,
		BetOptions:  betOptions,
	})
}

// PreviewSkipHandler handles the /preview-skip/hilo endpoint (pre-game phase)
func (rg *RouteGroup) PreviewSkipHandler(c *fiber.Ctx) error {
	var req PreviewSkipRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if req.ClientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "client_id is required",
		})
	}
	if req.GameID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "game_id is required",
		})
	}
	if req.PlayerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "player_id is required",
		})
	}
	if req.CurrentCard == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "current_card is required",
		})
	}

	// Generate a new random card for preview skip
	seed := GenerateUniqueSeed()
	deck := GenerateDeck(seed)
	nextCard := deck[0]

	// Generate betting options for the new card (base multipliers for preview skip)
	betOptions := GetBaseHiloOptions(nextCard)

	log.Printf("Preview Skip: Skipped from %s to %s", req.CurrentCard, nextCard)

	return c.JSON(PreviewSkipResponse{
		Status:      "success",
		Message:     fmt.Sprintf("Skipped to next card: %s", nextCard),
		CurrentCard: nextCard,
		BetOptions:  betOptions,
	})
}
