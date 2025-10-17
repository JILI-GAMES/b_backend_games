package magicace

import (
	"fmt"
	// "log"
	"math/rand"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SpinHandler handles the /spin/magicace endpoint
func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
	// Select correct clients for this request
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

	// Process existing jokers at the start of a new spin
	if req.GameState.JokerCards != nil && len(req.GameState.JokerCards) > 0 {
		var retainedJokers []JokerCard

		for _, joker := range req.GameState.JokerCards {
			rg.GameLogger.Debug("Start of spin - processing existing joker: %s at %d,%d with %d remaining rounds",
				joker.Mode, joker.Position.Reel, joker.Position.Row, joker.RemainingRounds)

			// Super Joker with counter = 1 is removed at start of new spin
			if joker.Mode == ModeSuperJoker && joker.RemainingRounds == 1 {
				rg.GameLogger.Debug("Removing Super Joker at %d,%d with counter 1 at start of new spin",
					joker.Position.Reel, joker.Position.Row)
				continue // Don't retain this joker
			}

			// Only Super Jokers are retained for the next spin
			// Big and Small Jokers should have been removed in previous spin/cascade
			if joker.Mode == ModeSuperJoker {
				retainedJokers = append(retainedJokers, joker)
			} else {
				rg.GameLogger.Debug("Unexpected %s Joker at start of spin - should have been removed earlier",
					joker.Mode)
				// Don't retain non-Super Jokers
			}
		}

		// Update the game state with retained jokers
		req.GameState.JokerCards = retainedJokers
		if req.GameState.SpecialSymbols.JokerCards == nil {
			req.GameState.SpecialSymbols.JokerCards = []JokerCard{}
		}
		req.GameState.SpecialSymbols.JokerCards = retainedJokers

		rg.GameLogger.Debug("Start of spin - retained %d jokers from previous spin", len(retainedJokers))
	}

	// Initialize game state
	if req.GameState.GameMode == "" {
		req.GameState.GameMode = "base"
	}
	if req.GameState.BoomingMultiplier == 0 {
		req.GameState.BoomingMultiplier = 1
		if req.GameState.GameMode == "freeSpins" {
			if req.GameState.Bet.ExtraBetEnabled {
				req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpinsExtraBet[0]
			} else {
				req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpins[0]
			}
		}
	}

	// Reset Booming Multiplier if Extra Bet is disabled in Base Game
	if !req.GameState.Bet.ExtraBetEnabled && req.GameState.GameMode == "base" {
		req.GameState.BoomingMultiplier = BoomingMultipliers[0] // Reset to 1
	}

	// Reset Booming Multiplier if Extra Bet is disabled in Free Spins
	if !req.GameState.Bet.ExtraBetEnabled && req.GameState.GameMode == "freeSpins" {
		req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpins[0] // Reset to 2
	}

	rg.GameLogger.Debug("##_____________##Game Mode: %s, Booming Multiplier: %d", req.GameState.GameMode, req.GameState.BoomingMultiplier)

	// Create a single rand.Rand instance for this request (Issue 1)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate reels with a potential win
	req.GameState.Bet.Multiplier = BetAmountToMultiplier[req.GameState.Bet.Amount]
	reels, specialSymbols := GenerateReelsWithWin(req.GameState.JokerCards, r)

	// Calculate winnings
	totalWinnings, winDetails := CalculateWins(reels, req.GameState.Bet.Multiplier, req.GameState.BoomingMultiplier, req.GameState.JokerCards)

	// Reset CascadeCount for a new spin sequence
	req.GameState.CascadeCount = 0
	if len(winDetails) > 0 {
		req.GameState.CascadeCount = 1 // First win is Cascade 1
	}

	// Get RTP
	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if err != nil {
		rg.GameLogger.Debug("Failed to get RTP: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve game settings",
		})
	}
	rg.GameLogger.Debug("RTP retrieved: %v", rtp)

	// Call RNG
	payoutMultiplier := totalWinnings / req.GameState.Bet.Amount
	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	rg.GameLogger.Debug("IP: %v", ip)
	rg.GameLogger.Debug("User-Agent: %v", userAgent)
	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.GameState.Bet.Amount, ip, userAgent, false)
	if err != nil {
		rg.GameLogger.Debug("Failed to call RNG API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to determine outcome",
		})
	}
	rg.GameLogger.Debug("RNG response: %v", rngResp)

	// Adjust outcome based on RNG
	if rngResp.PrefOutcome == "loss" {
		rg.GameLogger.Debug("RNG determined a loss outcome")

		// Make sure jokerCards is initialized
		if req.GameState.JokerCards == nil {
			req.GameState.JokerCards = []JokerCard{}
		}

		reels, specialSymbols = GenerateLossReels(req.GameState.JokerCards, r)

		// Important: Make sure joker cards are properly set in specialSymbols
		specialSymbols.JokerCards = req.GameState.JokerCards

		totalWinnings = 0
		winDetails = nil
	}

	req.GameState.Reels = reels
	req.GameState.SpecialSymbols = specialSymbols
	req.GameState.TotalWin = totalWinnings
	req.GameState.LastWinDetails = winDetails
	req.GameState.Cascading = len(winDetails) > 0

	// Count Scatters and trigger Free Spins
	req.GameState.ScatterCount, req.GameState.SpecialSymbols.TargetSymbols = CountScatters(reels)
	if req.GameState.GameMode == "base" && req.GameState.ScatterCount >= 3 {
		rg.GameLogger.Debug("Free Spins triggered: %d scatter(s)", req.GameState.ScatterCount)
		req.GameState.GameMode = "freeSpins"
		req.GameState.FreeSpins.ScattersTriggered = req.GameState.ScatterCount
		if req.GameState.ScatterCount == 3 {
			req.GameState.FreeSpins.Remaining = 10
			req.GameState.FreeSpins.TotalAwarded = 10
		} else if req.GameState.ScatterCount == 4 {
			req.GameState.FreeSpins.Remaining = 12
			req.GameState.FreeSpins.TotalAwarded = 12
		} else {
			req.GameState.FreeSpins.Remaining = 15
			req.GameState.FreeSpins.TotalAwarded = 15
		}
		if req.GameState.Bet.ExtraBetEnabled {
			req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpinsExtraBet[0]
		} else {
			req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpins[0]
		}
	} else if req.GameState.GameMode == "freeSpins" && req.GameState.ScatterCount >= 3 {
		additionalSpins := 0
		if req.GameState.ScatterCount == 3 {
			additionalSpins = 5
		} else if req.GameState.ScatterCount == 4 {
			additionalSpins = 8
		} else {
			additionalSpins = 10
		}
		req.GameState.FreeSpins.Remaining += additionalSpins
		req.GameState.FreeSpins.TotalAwarded += additionalSpins
		req.GameState.FreeSpins.ScattersTriggered = req.GameState.ScatterCount
		rg.GameLogger.Debug("Free Spins retriggered: %d scatter(s), %d additional spins, total %d spins", req.GameState.ScatterCount, additionalSpins, req.GameState.FreeSpins.TotalAwarded)
	}

	// Update Free Spins
	if req.GameState.GameMode == "freeSpins" {
		req.GameState.FreeSpins.Remaining--
		if req.GameState.FreeSpins.Remaining <= 0 {
			rg.GameLogger.Debug("Free Spins ended")
			req.GameState.GameMode = "base"
			req.GameState.FreeSpins = struct {
				Remaining         int `json:"remaining"`
				TotalAwarded      int `json:"totalAwarded"`
				ScattersTriggered int `json:"scattersTriggered"`
			}{0, 0, 0}
			req.GameState.BoomingMultiplier = 1
		}
	}

	// Update Joker Cards after the spin
	var updatedJokerCards []JokerCard
	for _, joker := range req.GameState.JokerCards {
		wasInWin := false
		for _, win := range winDetails {
			for _, pos := range win.Payline {
				if pos.Reel == joker.Position.Reel && pos.Row == joker.Position.Row {
					wasInWin = true
					break
				}
			}
			if wasInWin {
				break
			}
		}

		rg.GameLogger.Debug("Processing joker %s at %d,%d with %d remaining rounds - was in win: %v",
			joker.Mode, joker.Position.Reel, joker.Position.Row, joker.RemainingRounds, wasInWin)

		// If joker formed a winning combination, remove it
		if wasInWin {
			rg.GameLogger.Debug("Joker was in win - removing")
			// Skip adding to updatedJokerCards to remove the joker

			// For Super Joker, log the counter decrement before removal
			if joker.Mode == ModeSuperJoker {
				rg.GameLogger.Debug("Super Joker was in win - counter decreased before removal")
			}
			continue
		} else {
			// Joker wasn't in a win
			if joker.Mode == ModeSuperJoker {
				// Super Joker still decreases counter after each spin
				joker.RemainingRounds--
				rg.GameLogger.Debug("Super Joker was not in win - counter decreased to %d", joker.RemainingRounds)

				// Keep only if counter is still above 0
				if joker.RemainingRounds > 0 {
					updatedJokerCards = append(updatedJokerCards, joker)
					rg.GameLogger.Debug("Super Joker retained with counter %d", joker.RemainingRounds)
				} else {
					rg.GameLogger.Debug("Super Joker counter reached 0 - removing")
				}
			} else {
				// IMPORTANT: Big and Small Jokers don't persist to next spin
				rg.GameLogger.Debug("%s Joker was not in win but doesn't persist - removing", joker.Mode)
				// Don't add to updatedJokerCards to remove

				// Replace the wild symbol with a random regular symbol
				randomSymbol := WeightedRandomSymbol(r)
				for randomSymbol == SymbolWild || randomSymbol == SymbolScatter ||
					strings.HasPrefix(string(randomSymbol), "golden_") {
					randomSymbol = WeightedRandomSymbol(r)
				}
				req.GameState.Reels[joker.Position.Reel][joker.Position.Row] = string(randomSymbol)
				rg.GameLogger.Debug("Replaced %s Joker at %d,%d with random symbol %s",
					joker.Mode, joker.Position.Reel, joker.Position.Row, randomSymbol)
			}
		}
	}

	req.GameState.JokerCards = updatedJokerCards
	req.GameState.SpecialSymbols.JokerCards = updatedJokerCards
	rg.GameLogger.Debug("Updated Joker Cards: %d", len(req.GameState.JokerCards))

	// PATCH: Remove wilds at positions where jokers were removed due to being in a win
	for _, win := range winDetails {
		for _, pos := range win.Payline {
			// If this position was a joker that was just removed, and is a wild, replace it with a random symbol
			found := false
			for _, joker := range updatedJokerCards {
				if joker.Position.Reel == pos.Reel && joker.Position.Row == pos.Row {
					found = true
					break
				}
			}
			if !found && req.GameState.Reels[pos.Reel][pos.Row] == string(SymbolWild) {
				// Use WeightedRandomSymbol to get a random non-special symbol
				randomSymbol := WeightedRandomSymbol(rand.New(rand.NewSource(time.Now().UnixNano())))
				// Avoid wild, scatter, and golden symbols
				for randomSymbol == SymbolWild || randomSymbol == SymbolScatter || strings.HasPrefix(string(randomSymbol), "golden_") {
					randomSymbol = WeightedRandomSymbol(rand.New(rand.NewSource(time.Now().UnixNano())))
				}
				req.GameState.Reels[pos.Reel][pos.Row] = string(randomSymbol)
				rg.GameLogger.Debug("PATCH: Replaced wild at %d,%d after joker removal with random symbol %s", pos.Reel, pos.Row, randomSymbol)
			}
		}
	}

	// Calculate total cost
	// total cost should be 0 if free game
	var totalCost float64
	if req.GameState.GameMode == "freeSpins" {
		totalCost = 0
	} else {
		totalCost = req.GameState.Bet.Amount
		if req.GameState.Bet.ExtraBetEnabled {
			totalCost *= 1.5 // Add 50% for Extra Bet
		}
	}

	// Validation to ensure all wild symbols have corresponding joker cards
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			if req.GameState.Reels[reel][row] == string(SymbolWild) {
				// Check if this position is in the joker cards
				found := false
				if req.GameState.JokerCards != nil {
					for _, joker := range req.GameState.JokerCards {
						if joker.Position.Reel == reel && joker.Position.Row == row {
							found = true
							break
						}
					}
				}

				// If wild symbol doesn't have a corresponding joker card, add one
				if !found {
					rg.GameLogger.Debug("WARNING: Found wild symbol at %d,%d without a corresponding joker card - adding one", reel, row)

					// Create a new joker card (small joker by default)
					newJoker := JokerCard{
						Position: Position{
							Reel: reel,
							Row:  row,
						},
						Mode:            ModeSmallJoker, // Default to small joker
						RemainingRounds: 1,              // Will be replaced next spin
					}

					// Initialize arrays if needed
					if req.GameState.JokerCards == nil {
						req.GameState.JokerCards = []JokerCard{}
					}
					if req.GameState.SpecialSymbols.JokerCards == nil {
						req.GameState.SpecialSymbols.JokerCards = []JokerCard{}
					}

					// Add the joker card to both arrays
					req.GameState.JokerCards = append(req.GameState.JokerCards, newJoker)
					req.GameState.SpecialSymbols.JokerCards = req.GameState.JokerCards
				}
			}
		}
	}

	rg.GameLogger.Debug("Spin completed: totalWin=%v, cascading=%v, gameMode=%s, totalCost=%v", totalWinnings, req.GameState.Cascading, req.GameState.GameMode, totalCost)

	return c.JSON(SpinResponse{
		Status:     "success",
		Message:    "",
		GameState:  req.GameState,
		WinDetails: winDetails,
		TotalCost:  totalCost,
	})
}

// CascadeHandler handles the /cascade/magicace endpoint
func (rg *RouteGroup) CascadeHandler(c *fiber.Ctx) error {
	// Select correct clients for this request
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
		rg.GameLogger.Debug("Invalid reels dimensions")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid reels",
		})
	}

	// Increment CascadeCount
	req.GameState.CascadeCount++

	// Update Booming Multiplier based on CascadeCount
	var boomingMultipliers []int
	var step int
	if req.GameState.GameMode == "freeSpins" {
		if req.GameState.Bet.ExtraBetEnabled {
			boomingMultipliers = BoomingMultipliersFreeSpinsExtraBet // [2, 4, 6]
			step = 2
		} else {
			boomingMultipliers = BoomingMultipliersFreeSpins // [2, 4, 6, 10]
			step = 2
		}
	} else {
		if req.GameState.Bet.ExtraBetEnabled {
			boomingMultipliers = BoomingMultipliersExtraBet // [1, 2, 3]
			step = 1
		} else {
			boomingMultipliers = BoomingMultipliers // [1, 2, 3, 5]
			step = 1
		}
	}

	index := req.GameState.CascadeCount - 1
	if index < len(boomingMultipliers) {
		req.GameState.BoomingMultiplier = boomingMultipliers[index]
	} else {
		lastMultiplier := boomingMultipliers[len(boomingMultipliers)-1]
		req.GameState.BoomingMultiplier = lastMultiplier + (index-len(boomingMultipliers)+1)*step
	}
	rg.GameLogger.Debug("Updated Booming Multiplier: %d (Cascade %d)", req.GameState.BoomingMultiplier, req.GameState.CascadeCount)

	// Create a single rand.Rand instance for this request
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Check if there are existing jokers in the input
	hasExistingJokers := req.GameState.JokerCards != nil && len(req.GameState.JokerCards) > 0
	rg.GameLogger.Debug("Request contains %d existing jokers", len(req.GameState.JokerCards))

	// Track positions to replace

	// Track unique positions to replace
	winningPositions := make(map[Position]bool)
	seenPositions := make(map[Position]bool)
	for _, win := range req.GameState.LastWinDetails {
		for i, pos := range win.Payline {
			if !seenPositions[pos] {
				rg.GameLogger.Debug("Symbol to replace: %s at position %d,%d", win.Symbols[i], pos.Reel, pos.Row)
				winningPositions[pos] = true
				seenPositions[pos] = true
			}
		}
	}

	// FIRST CASCADE CASE: Check if there are golden cards in the last win details
	hasGoldenCards := false
	for _, win := range req.GameState.LastWinDetails {
		if len(win.GoldenCards) > 0 {
			hasGoldenCards = true
			break
		}
	}

	rg.GameLogger.Debug("$$$$$$$$$$$ hasGoldenCards: %v", hasGoldenCards)

	// First cascade case: Transform golden cards to jokers
	if hasGoldenCards {
		rg.GameLogger.Debug("Processing golden cards: Transforming golden cards to jokers")
		// Transform Golden Cards if present in the last win
		rg.GameLogger.Debug("Transforming Golden Cards from last win details")
		newJokerCards := TransformGoldenCards(req.GameState.Reels, req.GameState.LastWinDetails, req.GameState.JokerCards, r)

		rg.GameLogger.Debug("<--------> newJokerCards: %v", newJokerCards)
		// Log joker cards created and remove their positions from winningPositions
		for _, joker := range newJokerCards {
			rg.GameLogger.Debug("New joker created: mode=%s, position=%d,%d, rounds=%d",
				joker.Mode, joker.Position.Reel, joker.Position.Row, joker.RemainingRounds)
			// Remove transformed positions from winningPositions to preserve the jokers
			delete(winningPositions, joker.Position)
			rg.GameLogger.Debug("Removed position %d,%d from winningPositions to preserve new joker",
				joker.Position.Reel, joker.Position.Row)
		}

		// Initialize jokerCards array if needed
		if req.GameState.JokerCards == nil {
			req.GameState.JokerCards = []JokerCard{}
		}

		// Add new jokers to the game state
		req.GameState.JokerCards = append(req.GameState.JokerCards, newJokerCards...)

		// Ensure specialSymbols.JokerCards is also properly initialized and updated
		if req.GameState.SpecialSymbols.JokerCards == nil {
			req.GameState.SpecialSymbols.JokerCards = []JokerCard{}
		}
		req.GameState.SpecialSymbols.JokerCards = req.GameState.JokerCards

		rg.GameLogger.Debug("Transformed %d Golden Cards into Joker Cards. Total jokers now: %d",
			len(newJokerCards), len(req.GameState.JokerCards))
	}

	// Second cascade case: Process existing jokers
	if hasExistingJokers {
		rg.GameLogger.Debug("Processing existing jokers")

		// First check which jokers are in winning positions
		jokersInWins := make(map[Position]bool)
		for pos := range winningPositions {
			for _, joker := range req.GameState.JokerCards {
				if joker.Position.Reel == pos.Reel && joker.Position.Row == pos.Row {
					jokersInWins[joker.Position] = true
					rg.GameLogger.Debug("Joker at position %d,%d is in a win", pos.Reel, pos.Row)
				}
			}
		}

		// Process jokers based on whether they're in wins
		var updatedJokerCards []JokerCard
		for _, joker := range req.GameState.JokerCards {
			wasInWin := jokersInWins[joker.Position]
			rg.GameLogger.Debug("Processing existing joker %s at %d,%d with %d remaining rounds - was in win: %v",
				joker.Mode, joker.Position.Reel, joker.Position.Row, joker.RemainingRounds, wasInWin)

			if wasInWin {
				// If joker is in a win
				if joker.Mode == ModeSuperJoker {
					// Super Joker in win - decrement counter
					joker.RemainingRounds--
					rg.GameLogger.Debug("Super Joker in win - counter decreased to %d", joker.RemainingRounds)

					if joker.RemainingRounds > 0 {
						// Keep Super Joker with decreased counter
						updatedJokerCards = append(updatedJokerCards, joker)
						rg.GameLogger.Debug("Super Joker retained with counter %d", joker.RemainingRounds)
					} else {
						// Counter reached 0, remove Super Joker
						rg.GameLogger.Debug("Super Joker counter reached 0 - removing")
						// Replace with random symbol in winningPositions for replacement
					}
				} else {
					// Big or Small Joker in win - they're removed
					rg.GameLogger.Debug("%s Joker in win - will be removed", joker.Mode)
					// Will be replaced with new symbols in winningPositions
				}
			} else {
				// If joker is not in a win
				if joker.Mode == ModeSuperJoker {
					// Super Joker not in win - keep with same counter
					updatedJokerCards = append(updatedJokerCards, joker)
					rg.GameLogger.Debug("Super Joker not in win - retained with counter %d", joker.RemainingRounds)

					// Remove position from winningPositions to preserve it
					delete(winningPositions, joker.Position)
				} else {
					// Big or Small Joker not in win - still removed after one cascade
					rg.GameLogger.Debug("%s Joker not in win but doesn't persist - removing", joker.Mode)
					// Will be replaced with random symbols in next step

					// // First replace the joker with a random symbol
					// randomSymbol := WeightedRandomSymbol(r)
					// for randomSymbol == SymbolWild || randomSymbol == SymbolScatter ||
					// 	strings.HasPrefix(string(randomSymbol), "golden_") {
					// 	randomSymbol = WeightedRandomSymbol(r)
					// }

					// Keep the joker symbol in the reels
					req.GameState.Reels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)

					updatedJokerCards = append(updatedJokerCards, joker)

					rg.GameLogger.Debug("'''''''''''''''''''''''''''updatedJokerCards: %v", updatedJokerCards)

					// log.Printf("Replaced %s Joker with %s", joker.Mode, randomSymbol)
				}
			}
		}

		// Update joker cards
		req.GameState.JokerCards = updatedJokerCards
		if req.GameState.SpecialSymbols.JokerCards == nil {
			req.GameState.SpecialSymbols.JokerCards = []JokerCard{}
		}
		req.GameState.SpecialSymbols.JokerCards = updatedJokerCards
	}

	rg.GameLogger.Debug("winningPositions: %v", winningPositions)
	// Generate new symbols for the cascade, only for non-joker positions
	newReels, specialSymbols := GenerateReelsForCascade(req.GameState.Reels, winningPositions, req.GameState.JokerCards, r)

	rg.GameLogger.Debug("@@@@@@@@@@@@@@@@@@@@@@@@@newReels: %v", newReels)

	// Calculate new wins
	payout, winDetails := CalculateWins(newReels, req.GameState.Bet.Multiplier, req.GameState.BoomingMultiplier, req.GameState.JokerCards)

	rg.GameLogger.Debug("TOTAL PAYOUT: %v", payout)
	rg.GameLogger.Debug("WIN DETAILS: %v", winDetails)

	// Call RNG
	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if err != nil {
		rg.GameLogger.Debug("Failed to get RTP: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve game settings",
		})
	}
	rg.GameLogger.Debug("RTP retrieved: %v", rtp)

	payoutMultiplier := payout / req.GameState.Bet.Amount
	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	rg.GameLogger.Debug("IP: %v", ip)
	rg.GameLogger.Debug("User-Agent: %v", userAgent)
	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.GameState.Bet.Amount, ip, userAgent, false)
	if err != nil {
		rg.GameLogger.Debug("Failed to call RNG API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to determine outcome",
		})
	}
	rg.GameLogger.Debug("RNG response: %v", rngResp)

	// Adjust outcome based on RNG
	if rngResp.PrefOutcome == "loss" {
		rg.GameLogger.Debug("RNG determined a loss outcome")

		// Check if there are any Super Jokers that need processing
		hasSuperJokers := false
		for _, joker := range req.GameState.JokerCards {
			if joker.Mode == ModeSuperJoker {
				hasSuperJokers = true
				break
			}
		}

		// Generate a loss grid that preserves jokers
		lossMakingPositions := make(map[Position]bool)

		// Only add positions that were actually part of the winning combination
		for pos := range winningPositions {
			// Skip positions with jokers
			isJoker := false
			for _, joker := range req.GameState.JokerCards {
				if joker.Position.Reel == pos.Reel && joker.Position.Row == pos.Row {
					isJoker = true
					break
				}
			}

			if !isJoker && newReels[pos.Reel][pos.Row] != string(SymbolScatter) &&
				!strings.HasPrefix(newReels[pos.Reel][pos.Row], "golden_") {
				lossMakingPositions[pos] = true
			}
		}

		// for reel := 0; reel < Reels; reel++ {
		// 	for row := 0; row < Rows; row++ {
		// 		pos := Position{Reel: reel, Row: row}

		// 		// Skip positions with jokers
		// 		isJoker := false
		// 		for _, joker := range req.GameState.JokerCards {
		// 			if joker.Position.Reel == reel && joker.Position.Row == row {
		// 				isJoker = true
		// 				break
		// 			}
		// 		}

		// 		if !isJoker && newReels[reel][row] != string(SymbolScatter) &&
		// 			!strings.HasPrefix(newReels[reel][row], "golden_") {
		// 			lossMakingPositions[pos] = true
		// 		}
		// 	}
		// }

		rg.GameLogger.Debug("############Loss making positions: %v", lossMakingPositions)

		rg.GameLogger.Debug("*************NEW REELS BEFORE FORCING LOSS: %v", newReels)

		lossReels, lossSpecialSymbols, winOverride := GenerateLossForCascadePreservingJokers(
			newReels, lossMakingPositions, req.GameState.JokerCards, r)

		if winOverride {
			// RNG wanted loss but we're accepting the win
			rg.GameLogger.Debug("RNG OVERRIDE: RNG requested loss but jokers make it impossible, accepting win")

			// Calculate the actual win with current reels and jokers
			actualPayout, actualWinDetails := CalculateWins(
				newReels,
				req.GameState.Bet.Multiplier,
				req.GameState.BoomingMultiplier,
				req.GameState.JokerCards,
			)

			// Update game state with the win
			req.GameState.Reels = newReels
			specialSymbols.JokerCards = req.GameState.JokerCards
			req.GameState.SpecialSymbols = specialSymbols
			req.GameState.TotalWin = actualPayout
			req.GameState.LastWinDetails = actualWinDetails
			req.GameState.Cascading = actualPayout > 0

			rg.GameLogger.Debug("WIN OVERRIDE RESULT: Payout=%.2f, Cascading=%t", actualPayout, req.GameState.Cascading)

			// Count scatters
			scatterPositions := make([]Position, 0)
			for reel := 0; reel < Reels; reel++ {
				for row := 0; row < Rows; row++ {
					if req.GameState.Reels[reel][row] == string(SymbolScatter) {
						scatterPositions = append(scatterPositions, Position{Reel: reel, Row: row})
					}
				}
			}
			req.GameState.SpecialSymbols.TargetSymbols = scatterPositions
			req.GameState.ScatterCount = len(scatterPositions)

			// Handle free spins if needed
			if req.GameState.GameMode == "base" && req.GameState.ScatterCount >= 3 {
				rg.GameLogger.Debug("Free Spins triggered during win override: %d scatter(s)", req.GameState.ScatterCount)
				req.GameState.GameMode = "freeSpins"
				req.GameState.FreeSpins.ScattersTriggered = req.GameState.ScatterCount
				if req.GameState.ScatterCount == 3 {
					req.GameState.FreeSpins.Remaining = 10
					req.GameState.FreeSpins.TotalAwarded = 10
				} else if req.GameState.ScatterCount == 4 {
					req.GameState.FreeSpins.Remaining = 12
					req.GameState.FreeSpins.TotalAwarded = 12
				} else {
					req.GameState.FreeSpins.Remaining = 15
					req.GameState.FreeSpins.TotalAwarded = 15
				}
				if req.GameState.Bet.ExtraBetEnabled {
					req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpinsExtraBet[0]
				} else {
					req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpins[0]
				}
			}

			// Ensure wild symbols have corresponding joker cards
			for reel := 0; reel < Reels; reel++ {
				for row := 0; row < Rows; row++ {
					if req.GameState.Reels[reel][row] == string(SymbolWild) {
						found := false
						for _, joker := range req.GameState.JokerCards {
							if joker.Position.Reel == reel && joker.Position.Row == row {
								found = true
								break
							}
						}
						if !found {
							rg.GameLogger.Debug("WARNING: Found wild symbol at %d,%d without joker card in win override - adding one", reel, row)
							newJoker := JokerCard{
								Position:        Position{Reel: reel, Row: row},
								Mode:            ModeSmallJoker,
								RemainingRounds: 1,
							}
							req.GameState.JokerCards = append(req.GameState.JokerCards, newJoker)
							req.GameState.SpecialSymbols.JokerCards = req.GameState.JokerCards
						}
					}
				}
			}

			return c.JSON(CascadeResponse{
				Status:                    "success",
				Message:                   "",
				GameState:                 req.GameState,
				WinDetails:                actualWinDetails,
				TotalCost:                 0,
				NeedsSuperJokerProcessing: false,
			})
		} else {
			// Successfully generated loss
			rg.GameLogger.Debug("**************NEW REELS AFTER FORCING LOSS: %v", lossReels)

			// Update the game state with loss
			lossSpecialSymbols.JokerCards = req.GameState.JokerCards
			req.GameState.Reels = lossReels
			req.GameState.SpecialSymbols = lossSpecialSymbols
			req.GameState.TotalWin = 0
			req.GameState.Cascading = false
			req.GameState.LastWinDetails = nil

			// Count scatters
			scatterPositions := make([]Position, 0)
			for reel := 0; reel < Reels; reel++ {
				for row := 0; row < Rows; row++ {
					if req.GameState.Reels[reel][row] == string(SymbolScatter) {
						scatterPositions = append(scatterPositions, Position{Reel: reel, Row: row})
					}
				}
			}
			req.GameState.SpecialSymbols.TargetSymbols = scatterPositions
			req.GameState.ScatterCount = len(scatterPositions)

			return c.JSON(CascadeResponse{
				Status:                    "success",
				Message:                   "",
				GameState:                 req.GameState,
				WinDetails:                nil,
				TotalCost:                 0,
				NeedsSuperJokerProcessing: hasSuperJokers,
			})
		}
	}

	// 	// Use preserving jokers version
	// 	newReels, specialSymbols = GenerateLossForCascadePreservingJokers(
	// 		newReels, lossMakingPositions, req.GameState.JokerCards, r)

	// 	log.Printf("**************NEW REELS AFTER FORCING LOSS: %v", newReels)

	// 	// Update the game state
	// 	specialSymbols.JokerCards = req.GameState.JokerCards
	// 	req.GameState.Reels = newReels
	// 	req.GameState.SpecialSymbols = specialSymbols
	// 	req.GameState.TotalWin = 0
	// 	req.GameState.Cascading = false
	// 	req.GameState.LastWinDetails = nil

	// 	// Count scatters
	// 	scatterPositions := make([]Position, 0)
	// 	for reel := 0; reel < Reels; reel++ {
	// 		for row := 0; row < Rows; row++ {
	// 			if req.GameState.Reels[reel][row] == string(SymbolScatter) {
	// 				scatterPositions = append(scatterPositions, Position{Reel: reel, Row: row})
	// 			}
	// 		}
	// 	}
	// 	req.GameState.SpecialSymbols.TargetSymbols = scatterPositions
	// 	req.GameState.ScatterCount = len(scatterPositions)

	// 	return c.JSON(CascadeResponse{
	// 		Status:                    "success",
	// 		Message:                   "",
	// 		GameState:                 req.GameState,
	// 		WinDetails:                nil,
	// 		TotalCost:                 0,
	// 		NeedsSuperJokerProcessing: hasSuperJokers,
	// 	})
	// }

	rg.GameLogger.Debug("*************New reels in win: %v", newReels)

	// Update game state
	req.GameState.Reels = newReels
	specialSymbols.JokerCards = req.GameState.JokerCards // Ensure jokers are included
	req.GameState.SpecialSymbols = specialSymbols
	req.GameState.TotalWin = payout
	req.GameState.LastWinDetails = winDetails
	req.GameState.Cascading = len(winDetails) > 0

	// Count scatters properly by scanning the final reels
	scatterPositions := make([]Position, 0)
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			if req.GameState.Reels[reel][row] == string(SymbolScatter) {
				scatterPositions = append(scatterPositions, Position{Reel: reel, Row: row})
			}
		}
	}

	// Update the special symbols and scatter count
	req.GameState.SpecialSymbols.TargetSymbols = scatterPositions
	req.GameState.ScatterCount = len(scatterPositions)

	rg.GameLogger.Debug("Final scatter count: %d", req.GameState.ScatterCount)

	// Checks for free spins triggering from base game during a cascade
	if req.GameState.GameMode == "base" && req.GameState.ScatterCount >= 3 {
		rg.GameLogger.Debug("Free Spins triggered during cascade: %d scatter(s)", req.GameState.ScatterCount)
		req.GameState.GameMode = "freeSpins"
		req.GameState.FreeSpins.ScattersTriggered = req.GameState.ScatterCount
		if req.GameState.ScatterCount == 3 {
			req.GameState.FreeSpins.Remaining = 10
			req.GameState.FreeSpins.TotalAwarded = 10
		} else if req.GameState.ScatterCount == 4 {
			req.GameState.FreeSpins.Remaining = 12
			req.GameState.FreeSpins.TotalAwarded = 12
		} else {
			req.GameState.FreeSpins.Remaining = 15
			req.GameState.FreeSpins.TotalAwarded = 15
		}
		if req.GameState.Bet.ExtraBetEnabled {
			req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpinsExtraBet[0]
		} else {
			req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpins[0]
		}
	}

	// Check for Free Spins retrigger in Free Spins mode (only with new Scatters)
	newScatterCount := len(specialSymbols.NewTargetSymbols)
	if req.GameState.GameMode == "freeSpins" && newScatterCount >= 3 {
		additionalSpins := 0
		if newScatterCount == 3 {
			additionalSpins = 5
		} else if newScatterCount == 4 {
			additionalSpins = 8
		} else {
			additionalSpins = 10
		}
		req.GameState.FreeSpins.Remaining += additionalSpins
		req.GameState.FreeSpins.TotalAwarded += additionalSpins
		req.GameState.FreeSpins.ScattersTriggered = newScatterCount
		rg.GameLogger.Debug("Free Spins retriggered in cascade: %d new scatter(s), %d additional spins, total %d spins", newScatterCount, additionalSpins, req.GameState.FreeSpins.TotalAwarded)
	}

	// Ensure all wild symbols have corresponding joker cards
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			if req.GameState.Reels[reel][row] == string(SymbolWild) {
				// Check if this position is in the joker cards
				found := false
				for _, joker := range req.GameState.JokerCards {
					if joker.Position.Reel == reel && joker.Position.Row == row {
						found = true
						break
					}
				}

				// If wild symbol doesn't have a corresponding joker card, add one
				if !found {
					rg.GameLogger.Debug("WARNING: Found wild symbol at %d,%d without a corresponding joker card - adding one", reel, row)

					// Create a new joker card (small joker by default)
					newJoker := JokerCard{
						Position: Position{
							Reel: reel,
							Row:  row,
						},
						Mode:            ModeSmallJoker, // Default to small joker
						RemainingRounds: 1,              // Will be replaced next spin
					}

					// Add the joker card
					req.GameState.JokerCards = append(req.GameState.JokerCards, newJoker)
					req.GameState.SpecialSymbols.JokerCards = req.GameState.JokerCards
				}
			}
		}
	}

	rg.GameLogger.Debug("Cascade completed: totalWin=%v, cascading=%v, gameMode=%s",
		req.GameState.TotalWin, req.GameState.Cascading, req.GameState.GameMode)

	return c.JSON(CascadeResponse{
		Status:                    "success",
		Message:                   "",
		GameState:                 req.GameState,
		WinDetails:                winDetails,
		TotalCost:                 0,
		NeedsSuperJokerProcessing: false,
	})
}

// ProcessSuperJokersHandler handles Super Jokers after a loss in cascade
func (rg *RouteGroup) ProcessSuperJokersHandler(c *fiber.Ctx) error {
	// Remove unused rngClient, settingsClient
	var req SpinRequest
	if err := c.BodyParser(&req); err != nil {
		rg.GameLogger.Error("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Create a rand instance
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Flag to indicate if should return to spin
	shouldReturnToSpin := false

	// Process each Super Joker
	var updatedJokerCards []JokerCard
	for _, joker := range req.GameState.JokerCards {
		if joker.Mode == ModeSuperJoker {
			// Decrement Super Joker counter
			joker.RemainingRounds--
			rg.GameLogger.Debug("Super Joker counter decreased to %d", joker.RemainingRounds)

			if joker.RemainingRounds == 0 {
				// Counter reached 0, replace with random symbol
				randomSymbol := WeightedRandomSymbol(r)
				for randomSymbol == SymbolWild || randomSymbol == SymbolScatter ||
					strings.HasPrefix(string(randomSymbol), "golden_") {
					randomSymbol = WeightedRandomSymbol(r)
				}
				req.GameState.Reels[joker.Position.Reel][joker.Position.Row] = string(randomSymbol)
				rg.GameLogger.Debug("Super Joker removed (counter=0) - replaced with %s", randomSymbol)
			} else {
				// Keep joker with decreased counter
				updatedJokerCards = append(updatedJokerCards, joker)
				rg.GameLogger.Debug("Super Joker retained with counter %d", joker.RemainingRounds)

				// If counter is 1, signal to return to spin
				if joker.RemainingRounds == 1 {
					shouldReturnToSpin = true
					rg.GameLogger.Debug("Super Joker has counter=1 - should return to spin")
				}
			}
		} else {
			// Shouldn't have non-Super Jokers here
			rg.GameLogger.Debug("WARNING: Unexpected %s Joker in ProcessSuperJokersHandler", joker.Mode)

			// Replace with random symbol
			randomSymbol := WeightedRandomSymbol(r)
			for randomSymbol == SymbolWild || randomSymbol == SymbolScatter ||
				strings.HasPrefix(string(randomSymbol), "golden_") {
				randomSymbol = WeightedRandomSymbol(r)
			}
			req.GameState.Reels[joker.Position.Reel][joker.Position.Row] = string(randomSymbol)
			rg.GameLogger.Debug("Replaced unexpected %s Joker with %s", joker.Mode, randomSymbol)
		}
	}

	// Update joker cards
	req.GameState.JokerCards = updatedJokerCards
	if req.GameState.JokerCards == nil {
		req.GameState.JokerCards = []JokerCard{}
	}
	req.GameState.SpecialSymbols.JokerCards = req.GameState.JokerCards

	// After updating jokers, replace all non-joker positions with random regular symbols
	jokerPositions := make(map[Position]bool)
	for _, joker := range req.GameState.JokerCards {
		jokerPositions[joker.Position] = true
	}

	// Build a set of all non-joker positions to randomize
	positionsToRandomize := make(map[Position]bool)
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			pos := Position{Reel: reel, Row: row}
			if !jokerPositions[pos] {
				positionsToRandomize[pos] = true
			}
		}
	}

	// Guarantee a loss by randomizing all non-joker positions until no win exists
	override := false
	req.GameState.Reels, req.GameState.SpecialSymbols, override = GenerateLossForCascadePreservingJokers(
		req.GameState.Reels, positionsToRandomize, req.GameState.JokerCards, rand.New(rand.NewSource(time.Now().UnixNano())),
	)
	if override {
		rg.GameLogger.Debug("RNG OVERRIDE: RNG requested loss but jokers make it impossible, accepting win")
	}

	// Update game state
	req.GameState.TotalWin = 0
	req.GameState.Cascading = false
	req.GameState.LastWinDetails = nil
	req.GameState.ScatterCount = 0
	req.GameState.SpecialSymbols.TargetSymbols = []Position{}
	req.GameState.SpecialSymbols.NewTargetSymbols = []Position{}

	return c.JSON(SpinResponse{
		Status:       "success",
		Message:      "",
		GameState:    req.GameState,
		WinDetails:   nil,
		TotalCost:    0,
		ReturnToSpin: shouldReturnToSpin,
	})
}

// FeatureBuyHandler handles the /featureBuy/magicace endpoint
func (rg *RouteGroup) FeatureBuyHandler(c *fiber.Ctx) error {
	// Remove unused rngClient, settingsClient
	var req FeatureBuyRequest
	if err := c.BodyParser(&req); err != nil {
		rg.GameLogger.Error("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
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
	if req.GameState.SpecialSymbols.NewTargetSymbols == nil {
		req.GameState.SpecialSymbols.NewTargetSymbols = []Position{}
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		rg.GameLogger.Error("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate option
	if req.Option != 1 && req.Option != 2 {
		rg.GameLogger.Debug("Invalid option, allowed values are 1 or 2")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid option, allowed values are 1 or 2",
		})
	}

	// Create a single rand.Rand instance for this request
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Calculate cost
	totalBet := float64(MinBet) * float64(req.GameState.Bet.Multiplier) * Denomination
	costMultiplier := 50
	if req.Option == 2 {
		costMultiplier = 75
	}
	cost := float64(costMultiplier) * totalBet

	// Generate reels with guaranteed Scatters
	scatterCount := 0
	var specialSymbols SpecialSymbols
	for {
		req.GameState.Reels, specialSymbols = GenerateReelsWithWin(req.GameState.JokerCards, r)
		scatterCount, specialSymbols.TargetSymbols = CountScatters(req.GameState.Reels)
		if scatterCount >= 3 {
			break
		}
	}

	// If option 2, ensure a Super Joker
	if req.Option == 2 {
		occupiedPositions := make(map[Position]bool)
		if req.GameState.JokerCards != nil {
			for _, joker := range req.GameState.JokerCards {
				occupiedPositions[joker.Position] = true
			}
		}
		for reel := 0; reel < Reels; reel++ {
			for row := 0; row < Rows; row++ {
				symbol := req.GameState.Reels[reel][row]
				if strings.Contains(symbol, "wild") || strings.HasPrefix(symbol, "golden_") || symbol == string(SymbolScatter) {
					occupiedPositions[Position{Reel: reel, Row: row}] = true
				}
			}
		}
		reel, row := getRandomPosition(req.GameState.Reels, r, -1, -1, occupiedPositions)
		if reel != -1 && row != -1 {
			newJoker := JokerCard{
				Position: Position{
					Reel: reel,
					Row:  row,
				},
				Mode:            ModeSuperJoker,
				RemainingRounds: 3, // Start with 3 rounds
			}
			if err := newJoker.ValidateMode(); err != nil {
				rg.GameLogger.Debug("Invalid Joker Card mode in FeatureBuy: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": "Failed to create Super Joker",
				})
			}
			if req.GameState.JokerCards == nil {
				req.GameState.JokerCards = []JokerCard{}
			}
			req.GameState.JokerCards = append(req.GameState.JokerCards, newJoker)
			req.GameState.Reels[reel][row] = string(SymbolWild)
			specialSymbols.JokerCards = req.GameState.JokerCards
		}
	}

	// Set Free Spins
	req.GameState.GameMode = "freeSpins"
	req.GameState.FreeSpins.ScattersTriggered = scatterCount
	if scatterCount == 3 {
		req.GameState.FreeSpins.Remaining = 10
		req.GameState.FreeSpins.TotalAwarded = 10
	} else if scatterCount == 4 {
		req.GameState.FreeSpins.Remaining = 12
		req.GameState.FreeSpins.TotalAwarded = 12
	} else {
		req.GameState.FreeSpins.Remaining = 15
		req.GameState.FreeSpins.TotalAwarded = 15
	}

	// Set Booming Multiplier
	if req.GameState.Bet.ExtraBetEnabled {
		req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpinsExtraBet[0]
	} else {
		req.GameState.BoomingMultiplier = BoomingMultipliersFreeSpins[0]
	}

	// Update scatterCount
	req.GameState.ScatterCount = scatterCount

	req.GameState.SpecialSymbols = specialSymbols
	req.GameState.TotalWin = 0
	req.GameState.Cascading = false
	req.GameState.LastWinDetails = []WinDetail{}

	// Validation to ensure all wild symbols have corresponding joker cards
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			if req.GameState.Reels[reel][row] == string(SymbolWild) {
				// Check if this position is in the joker cards
				found := false
				if req.GameState.JokerCards != nil {
					for _, joker := range req.GameState.JokerCards {
						if joker.Position.Reel == reel && joker.Position.Row == row {
							found = true
							break
						}
					}
				}

				// If wild symbol doesn't have a corresponding joker card, add one
				if !found {
					rg.GameLogger.Debug("WARNING: Found wild symbol at %d,%d without a corresponding joker card in FeatureBuy - adding one", reel, row)

					// Create a new joker card (small joker by default)
					newJoker := JokerCard{
						Position: Position{
							Reel: reel,
							Row:  row,
						},
						Mode:            ModeSmallJoker, // Default to small joker
						RemainingRounds: 1,              // Will be replaced next spin
					}

					// Initialize arrays if needed
					if req.GameState.JokerCards == nil {
						req.GameState.JokerCards = []JokerCard{}
					}
					if req.GameState.SpecialSymbols.JokerCards == nil {
						req.GameState.SpecialSymbols.JokerCards = []JokerCard{}
					}

					// Add the joker card to both arrays
					req.GameState.JokerCards = append(req.GameState.JokerCards, newJoker)
					req.GameState.SpecialSymbols.JokerCards = req.GameState.JokerCards
				}
			}
		}
	}

	rg.GameLogger.Debug("Feature Buy completed: option=%d, cost=%.2f", req.Option, cost)

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
