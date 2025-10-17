package magicaceoriginal

import (
	"fmt"
	"math/rand"

	// "strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SpinHandler handles the /spin/magicaceoriginal endpoint
func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
	rngClient, settingsClient := rg.getClientsForRequest(c)

	var req SpinRequest
	if err := c.BodyParser(&req); err != nil {
		rg.GameLogger.Error("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		rg.GameLogger.Error("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Initialize game state
	if req.GameState.GameMode == "" {
		req.GameState.GameMode = "base"
	}

	// Clear any existing joker cards (they don't persist across spins)
	req.GameState.JokerCards = []JokerCard{}

	// Create a single rand.Rand instance for this request
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate reels with a potential win
	req.GameState.Bet.Multiplier = BetAmountToMultiplier[req.GameState.Bet.Amount]
	reels, specialSymbols := GenerateReelsWithWin(req.GameState.JokerCards, r)

	// Reset cascade count for new spin
	req.GameState.CascadeCount = 0

	// Calculate initial booming multiplier
	req.GameState.BoomingMultiplier = GetBoomingMultiplier(
		1, // First cascade
		req.GameState.GameMode,
		req.GameState.BaseGameCollector,
		req.GameState.FreeSpinsCollector,
	)

	// Calculate winnings
	totalWinnings, winDetails := CalculateWins(reels, req.GameState.Bet.Multiplier, req.GameState.BoomingMultiplier, req.GameState.JokerCards)

	if len(winDetails) > 0 {
		req.GameState.CascadeCount = 1 // First win is Cascade 1
	}

	// // Update collector rounds if there are wins (BEFORE RNG call, only for initial spin wins)
	// hasWins := len(winDetails) > 0
	// log.Printf("✅Has wins: %v", hasWins)
	// log.Printf("✅Game mode: %v", req.GameState.GameMode)
	// if hasWins && req.GameState.GameMode == "base" {
	// 	UpdateBaseGameCollectorRounds(&req.GameState.BaseGameCollector, true)
	// }

	// Get RTP
	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if err != nil {
		rg.GameLogger.Info("Failed to get RTP: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve game settings",
		})
	}

	// Call RNG
	payoutMultiplier := totalWinnings / req.GameState.Bet.Amount

	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	rg.GameLogger.Debug("IP: %v", ip)
	rg.GameLogger.Debug("User-Agent: %v", userAgent)

	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.GameState.Bet.Amount, ip, userAgent, false)
	if err != nil {
		rg.GameLogger.Info("Failed to call RNG API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to determine outcome",
		})
	}

	// Adjust outcome based on RNG
	if rngResp.PrefOutcome == "loss" {
		rg.GameLogger.Info("RNG determined a loss outcome")
		reels, specialSymbols = GenerateLossReels(req.GameState.JokerCards, r)
		totalWinnings = 0
		winDetails = nil
		req.GameState.CascadeCount = 0

	}

	req.GameState.Reels = reels
	req.GameState.SpecialSymbols = specialSymbols
	req.GameState.TotalWin = totalWinnings
	req.GameState.LastWinDetails = winDetails
	req.GameState.Cascading = len(winDetails) > 0

	// Update collector rounds if there are wins (BEFORE RNG call, only for initial spin wins)
	hasWins := len(winDetails) > 0
	rg.GameLogger.Info("✅Has wins: %v", hasWins)
	rg.GameLogger.Info("✅Game mode: %v", req.GameState.GameMode)
	if hasWins && req.GameState.GameMode == "base" {
		UpdateBaseGameCollectorRounds(&req.GameState.BaseGameCollector, true)
	}

	// Count Target symbols for Free Spins trigger only
	req.GameState.TargetCount, req.GameState.SpecialSymbols.TargetSymbols = CountTargets(reels)

	// Check for Free Spins trigger (3+ targets)
	if req.GameState.GameMode == "base" && req.GameState.TargetCount >= 3 {
		rg.GameLogger.Info("Free Spins triggered: %d target(s)", req.GameState.TargetCount)
		req.GameState.GameMode = "freeSpins"
		req.GameState.FreeSpins.Remaining = 10
		req.GameState.FreeSpins.TotalAwarded = 10
		req.GameState.FreeSpins.Retriggers = 0
		// Free spins collector is already initialized, no need to reset
	}

	// Update Free Spins
	if req.GameState.GameMode == "freeSpins" {
		req.GameState.FreeSpins.Remaining--
		if req.GameState.FreeSpins.Remaining <= 0 {
			rg.GameLogger.Info("Free Spins ended")
			req.GameState.GameMode = "base"
			req.GameState.FreeSpins = struct {
				Remaining    int `json:"remaining"`
				TotalAwarded int `json:"totalAwarded"`
				Retriggers   int `json:"retriggers"`
			}{0, 0, 0}
			// Reset free spins collector when returning to base game
			ResetFreeSpinsCollector(&req.GameState.FreeSpinsCollector)
		}
	}

	// Calculate total cost
	var totalCost float64
	if req.GameState.GameMode == "freeSpins" {
		totalCost = 0
	} else {
		totalCost = req.GameState.Bet.Amount
		if req.GameState.Bet.ExtraBetEnabled {
			totalCost *= 1.5 // Add 50% for Extra Bet
		}
	}

	rg.GameLogger.Info("Spin completed: totalWin=%v, cascading=%v, gameMode=%s, totalCost=%v", totalWinnings, req.GameState.Cascading, req.GameState.GameMode, totalCost)

	return c.JSON(SpinResponse{
		Status:     "success",
		Message:    "",
		GameState:  req.GameState,
		WinDetails: winDetails,
		TotalCost:  totalCost,
	})
}

// CascadeHandler handles the /cascade/magicaceoriginal endpoint
func (rg *RouteGroup) CascadeHandler(c *fiber.Ctx) error {
	rngClient, settingsClient := rg.getClientsForRequest(c)

	var req CascadeRequest
	if err := c.BodyParser(&req); err != nil {
		rg.GameLogger.Error("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		rg.GameLogger.Error("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate reels
	if len(req.GameState.Reels) != Reels || len(req.GameState.Reels[0]) != Rows {
		rg.GameLogger.Info("Invalid reels dimensions")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid reels",
		})
	}

	// Increment CascadeCount
	req.GameState.CascadeCount++

	// Update Booming Multiplier based on CascadeCount
	req.GameState.BoomingMultiplier = GetBoomingMultiplier(
		req.GameState.CascadeCount,
		req.GameState.GameMode,
		req.GameState.BaseGameCollector,
		req.GameState.FreeSpinsCollector,
	)

	rg.GameLogger.Info("Updated Booming Multiplier: %d (Cascade %d)", req.GameState.BoomingMultiplier, req.GameState.CascadeCount)

	// Create a single rand.Rand instance for this request
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// STEP 1: Check if there are golden cards in the last win details
	hasGoldenCards := false
	for _, win := range req.GameState.LastWinDetails {
		if len(win.GoldenCards) > 0 {
			hasGoldenCards = true
			break
		}
	}

	// STEP 2: Track which joker positions were newly created from golden cards in THIS cascade
	newlyCreatedJokerPositions := make(map[Position]bool)
	if hasGoldenCards {
		// Get the positions of golden cards that were transformed
		for _, win := range req.GameState.LastWinDetails {
			for _, goldenPos := range win.GoldenCards {
				newlyCreatedJokerPositions[goldenPos] = true
			}
		}
	}

	// STEP 3: Transform Golden Cards if present in the last win
	if hasGoldenCards {
		rg.GameLogger.Info("Processing golden cards: Transforming golden cards to jokers")
		newJokerCards := TransformGoldenCards(req.GameState.Reels, req.GameState.LastWinDetails, r)

		// Add new jokers to the game state
		req.GameState.JokerCards = append(req.GameState.JokerCards, newJokerCards...)
		req.GameState.SpecialSymbols.JokerCards = req.GameState.JokerCards

		rg.GameLogger.Info("Transformed %d Golden Cards into Joker Cards", len(newJokerCards))

		// CRITICAL: Ensure reels show 'wild' at joker positions immediately
		for _, joker := range newJokerCards {
			req.GameState.Reels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
			rg.GameLogger.Info("Set position %d,%d to 'wild' for new joker", joker.Position.Reel, joker.Position.Row)
		}

		rg.GameLogger.Info("✅New joker cards: %v", newJokerCards)

		// Count Super Jokers and update collectors
		superJokerCount := CountSuperJokers(newJokerCards)
		rg.GameLogger.Info("✅Super Jokers count: %d", superJokerCount)
		if superJokerCount > 0 {
			if req.GameState.GameMode == "base" {
				UpdateBaseGameCollector(&req.GameState.BaseGameCollector, superJokerCount)
			} else if req.GameState.GameMode == "freeSpins" {
				UpdateFreeSpinsCollector(&req.GameState.FreeSpinsCollector, superJokerCount, req.GameState.Bet.ExtraBetEnabled)
			}
		}
	}

	// STEP 4: Track positions to replace (AFTER golden card transformation)
	winningPositions := make(map[Position]bool)
	seenPositions := make(map[Position]bool)
	for _, win := range req.GameState.LastWinDetails {
		for i, pos := range win.Payline {
			if !seenPositions[pos] {
				rg.GameLogger.Info("Symbol to replace: %s at position %d,%d", win.Symbols[i], pos.Reel, pos.Row)
				winningPositions[pos] = true
				seenPositions[pos] = true
			}
		}
	}

	rg.GameLogger.Info("Winning positions before generation: %v", winningPositions)

	// STEP 5: Generate new symbols for the cascade (this now handles joker removal internally)
	newReels, specialSymbols, remainingJokers := GenerateReelsForCascade(
		req.GameState.Reels,
		winningPositions,
		req.GameState.JokerCards,
		req.GameState.LastWinDetails,
		newlyCreatedJokerPositions,
		r,
	)

	rg.GameLogger.Info("New reels after generation: %v", newReels)
	rg.GameLogger.Info("Remaining jokers after generation: %v", remainingJokers)

	// STEP 6: Calculate new wins using the UPDATED joker array and reels
	payout, winDetails := CalculateWins(newReels, req.GameState.Bet.Multiplier, req.GameState.BoomingMultiplier, remainingJokers)

	rg.GameLogger.Info("Calculated payout: %v, winDetails: %v", payout, winDetails)

	// NOTE: Do NOT update collector rounds during cascades - only on initial spin wins

	// STEP 7: Get RTP
	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if err != nil {
		rg.GameLogger.Info("Failed to get RTP: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve game settings",
		})
	}

	// STEP 8: Call RNG
	payoutMultiplier := payout / req.GameState.Bet.Amount
	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	rg.GameLogger.Debug("IP: %v", ip)
	rg.GameLogger.Debug("User-Agent: %v", userAgent)
	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.GameState.Bet.Amount, ip, userAgent, false)
	if err != nil {
		rg.GameLogger.Info("Failed to call RNG API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to determine outcome",
		})
	}

	// STEP 9: Adjust outcome based on RNG
	if rngResp.PrefOutcome == "loss" {
		rg.GameLogger.Info("RNG determined a loss outcome")
		newReels, specialSymbols, remainingJokers = GenerateLossForCascade(
			req.GameState.Reels,
			winningPositions,
			req.GameState.JokerCards,
			req.GameState.LastWinDetails,
			newlyCreatedJokerPositions,
			r,
		)
		payout = 0
		winDetails = nil
		rg.GameLogger.Info("Loss reels: %v", newReels)
		rg.GameLogger.Info("Loss remaining jokers: %v", remainingJokers)
	}

	// STEP 10: Update game state with final values
	req.GameState.Reels = newReels
	req.GameState.JokerCards = remainingJokers
	specialSymbols.JokerCards = remainingJokers // Ensure jokers are included in special symbols
	req.GameState.SpecialSymbols = specialSymbols
	req.GameState.TotalWin = payout
	req.GameState.LastWinDetails = winDetails
	req.GameState.Cascading = len(winDetails) > 0

	// STEP 11: Count Target symbols for Free Spins trigger only
	req.GameState.TargetCount, req.GameState.SpecialSymbols.TargetSymbols = CountTargets(newReels)

	// STEP 12: Check for Free Spins triggering from base game during a cascade
	if req.GameState.GameMode == "base" && req.GameState.TargetCount >= 3 {
		rg.GameLogger.Info("Free Spins triggered during cascade: %d target(s)", req.GameState.TargetCount)
		req.GameState.GameMode = "freeSpins"
		req.GameState.FreeSpins.Remaining = 10
		req.GameState.FreeSpins.TotalAwarded = 10
		req.GameState.FreeSpins.Retriggers = 0
	}

	// STEP 13: Check for Free Spins retrigger in Free Spins mode
	if req.GameState.GameMode == "freeSpins" && req.GameState.TargetCount >= 3 {
		additionalSpins := 5
		req.GameState.FreeSpins.Remaining += additionalSpins
		req.GameState.FreeSpins.TotalAwarded += additionalSpins
		req.GameState.FreeSpins.Retriggers++
		rg.GameLogger.Info("Free Spins retriggered in cascade: %d target(s), %d additional spins", req.GameState.TargetCount, additionalSpins)
	}

	rg.GameLogger.Info("Cascade completed: totalWin=%v, cascading=%v, gameMode=%s",
		req.GameState.TotalWin, req.GameState.Cascading, req.GameState.GameMode)

	return c.JSON(CascadeResponse{
		Status:     "success",
		Message:    "",
		GameState:  req.GameState,
		WinDetails: winDetails,
		TotalCost:  0,
	})
}

// FeatureBuyHandler handles the /featureBuy/magicaceoriginal endpoint
func (rg *RouteGroup) FeatureBuyHandler(c *fiber.Ctx) error {
	var req FeatureBuyRequest
	if err := c.BodyParser(&req); err != nil {
		rg.GameLogger.Error("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	//validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		rg.GameLogger.Error("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Initialize fields to avoid null in JSON response
	if req.GameState.JokerCards == nil {
		req.GameState.JokerCards = []JokerCard{}
	}
	if req.GameState.SpecialSymbols.GoldenCards == nil {
		req.GameState.SpecialSymbols.GoldenCards = []Position{}
	}
	if req.GameState.SpecialSymbols.JokerCards == nil {
		req.GameState.SpecialSymbols.JokerCards = []JokerCard{}
	}
	if req.GameState.SpecialSymbols.TargetSymbols == nil {
		req.GameState.SpecialSymbols.TargetSymbols = []Position{}
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		rg.GameLogger.Error("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Create a single rand.Rand instance for this request
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Calculate cost (50x total bet)
	totalBet := float64(MinBet) * float64(req.GameState.Bet.Multiplier) * Denomination
	cost := 50.0 * totalBet

	// Generate reels with guaranteed 3+ Target symbols
	targetCount := 0
	var specialSymbols SpecialSymbols
	for {
		req.GameState.Reels, specialSymbols = GenerateReelsWithWin(req.GameState.JokerCards, r)
		targetCount, specialSymbols.TargetSymbols = CountTargets(req.GameState.Reels)
		if targetCount >= 3 {
			break
		}
	}

	// Set Free Spins
	req.GameState.GameMode = "freeSpins"
	req.GameState.FreeSpins.Remaining = 10
	req.GameState.FreeSpins.TotalAwarded = 10
	req.GameState.FreeSpins.Retriggers = 0
	// Free spins collector is already initialized, no need to reset

	// Update target count
	req.GameState.TargetCount = targetCount

	req.GameState.SpecialSymbols = specialSymbols
	req.GameState.TotalWin = 0
	req.GameState.Cascading = false
	req.GameState.LastWinDetails = []WinDetail{}

	rg.GameLogger.Info("Feature Buy completed: cost=%.2f", cost)

	return c.JSON(FeatureBuyResponse{
		Status:     "success",
		Message:    fmt.Sprintf("Feature Buy successful, cost: %.2f", cost),
		GameState:  req.GameState,
		WinDetails: req.GameState.LastWinDetails,
		TotalCost:  cost,
	})
}

// validateRequest validates the request fields
func validateRequest(clientID, gameID, playerID, betID string, betAmount float64) error {
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
	if !isValidBetAmount(betAmount) {
		return fmt.Errorf("invalid bet amount, allowed values are 0.2, 0.4, 0.6, 1.0, 2.0")
	}
	return nil
}

// isValidBetAmount checks if the bet amount is valid
func isValidBetAmount(amount float64) bool {
	validAmounts := []float64{0.2, 0.4, 0.6, 1.0, 2.0}
	for _, valid := range validAmounts {
		if amount == valid {
			return true
		}
	}
	return false
}
