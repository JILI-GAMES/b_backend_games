package kong

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
	// Parse the request
	var req SpinRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(SpinResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Define allowed bet amounts and their corresponding multipliers
	betMultipliers := map[float64]int{
		0.3: 1,
		0.6: 2,
		0.9: 3,
		1.5: 5,
		3.0: 10,
	}

	// Validate the request
	if req.BetAmount <= 0 && !req.IsFreeSpin {
		log.Printf("Validation error: Bet amount must be positive")
		return c.Status(fiber.StatusBadRequest).JSON(SpinResponse{
			Status:  "error",
			Message: "Bet amount must be positive",
		})
	}
	// validate clientid, playerid, betid, gameid
	if req.ClientID == "" || req.PlayerID == "" || req.BetID == "" || req.GameID == "" {
		log.Printf("Validation error: ClientID, PlayerID, BetID, GameID must not be empty")
		return c.Status(fiber.StatusBadRequest).JSON(SpinResponse{
			Status:  "error",
			Message: "ClientID, PlayerID, BetID, GameID must not be empty",
		})
	}

	if req.IsFreeSpin && req.FreeSpinCount <= 0 {
		log.Printf("Validation error: Free spin count must be positive")
		return c.Status(fiber.StatusBadRequest).JSON(SpinResponse{
			Status:  "error",
			Message: "Free spin count must be positive",
		})
	}
	if req.IsFreeSpin && req.OriginalBetAmount <= 0 {
		log.Printf("Validation error: Original bet amount must be positive")
		return c.Status(fiber.StatusBadRequest).JSON(SpinResponse{
			Status:  "error",
			Message: "Original bet amount must be positive",
		})
	}

	// Validate bonus multiplier during free spins
	if req.IsFreeSpin && req.BonusMultiplier <= 0 {
		log.Printf("Warning: Bonus multiplier is 0 during free spins, setting to default value (3)")
		req.BonusMultiplier = 3
	}

	// Validate bet amount for normal spins
	if !req.IsFreeSpin {
		if _, exists := betMultipliers[req.BetAmount]; !exists {
			log.Printf("Validation error: Invalid bet amount %f, allowed values are 0.3, 0.6, 0.9, 1.5, 3.0", req.BetAmount)
			return c.Status(fiber.StatusBadRequest).JSON(SpinResponse{
				Status:  "error",
				Message: "Invalid bet amount, allowed values are 0.3, 0.6, 0.9, 1.5, 3.0",
			})
		}
	}

	// Get the bet multiplier for normal spins
	betMultiplier := 1 // Default for free spins (not used)
	if !req.IsFreeSpin {
		betMultiplier = betMultipliers[req.BetAmount]
	}
	log.Printf("Bet multiplier: %d", betMultiplier)

	// Select correct clients for this request
	rngClient, settingsClient := rg.getClientsForRequest(c)

	// Call the Settings API to get RTP
	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if err != nil {
		log.Printf("Error retrieving game settings: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(SpinResponse{
			Status:  "error",
			Message: "Failed to retrieve game settings: " + err.Error(),
		})
	}
	log.Printf("Retrieved RTP: %f", rtp)

	// Generate reels with a guaranteed win
	reels := GenerateReels(true)
	log.Printf("Generated reels: %v", reels)

	// Calculate potential win and winning positions
	potentialWin, winningPositions := CalculateRegularWin(reels, req.IsFreeSpin, betMultiplier, req.BonusMultiplier)
	log.Printf("Potential win: %f", potentialWin)
	log.Printf("Winning positions: %v", winningPositions)

	// Calculate payout multiplier
	betAmountForPayout := req.BetAmount
	if req.IsFreeSpin {
		betAmountForPayout = req.OriginalBetAmount
	}
	payoutMultiplier := potentialWin / betAmountForPayout
	log.Printf("Payout multiplier: %f", payoutMultiplier)

	// Call the RNG API
	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("✅IP: %v", ip)
	log.Printf("✅User-Agent: %v", userAgent)
	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, betAmountForPayout, ip, userAgent)
	if err != nil {
		log.Printf("Error retrieving RNG outcome: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(SpinResponse{
			Status:  "error",
			Message: "Failed to retrieve RNG outcome: " + err.Error(),
		})
	}
	log.Printf("RNG outcome: %s", rngResp.PrefOutcome)

	// Apply the RNG outcome
	winAmount := potentialWin
	if rngResp.PrefOutcome == "loss" {
		// Force a sensical loss (no winning combinations)
		reels = GenerateReels(false)
		log.Printf("Forced loss, new reels: %v", reels)
		winAmount = 0
		winningPositions = []WinningPosition{}
	}

	// Check for Free Spin Bonus trigger/retrigger
	freeSpinTriggered := false
	newFreeSpinCount := req.FreeSpinCount
	newBonusMultiplier := req.BonusMultiplier
	newIsFreeSpin := req.IsFreeSpin

	if HasScatterOnEachReel(reels) {
		freeSpinTriggered = true
		scatterCount := CountScatters(reels)
		newBonusMultiplier = GetBonusMultiplier(scatterCount)
		if newIsFreeSpin {
			// Retrigger
			newFreeSpinCount += 13
			if newFreeSpinCount > 100 {
				newFreeSpinCount = 100
			}
		} else {
			// Initial trigger
			newFreeSpinCount = 13
			newIsFreeSpin = true
		}
		log.Printf("Free spin triggered/retriggered: scatterCount=%d, newFreeSpinCount=%d, newBonusMultiplier=%d", scatterCount, newFreeSpinCount, newBonusMultiplier)
	}

	// // Add scatter positions to winning positions when free spins are triggered
	// for reel := 0; reel < 5; reel++ {
	// 	for row := 0; row < 3; row++ {
	// 		if reels[reel][row] == string(SymbolScatter) {
	// 			winningPositions = append(winningPositions, WinningPosition{
	// 				Symbol:   string(SymbolScatter),
	// 				Reel:     reel,
	// 				Row:      row,
	// 				Count:    scatterCount,
	// 				Ways:     1,
	// 				WinValue: 0, // Scatters don't have direct win value, but trigger free spins
	// 			})
	// 		}
	// 	}
	// }

	// Update free spin state
	if newIsFreeSpin {
		// Only decrement if the player was already in free spin mode
		if req.IsFreeSpin {
			newFreeSpinCount--
			if newFreeSpinCount <= 0 {
				newIsFreeSpin = false
				newBonusMultiplier = 0
			}
		}

	}

	// Build the response
	response := SpinResponse{
		Status:            "success",
		Message:           "",
		Reels:             reels,
		WinAmount:         winAmount,
		IsFreeSpin:        newIsFreeSpin,
		FreeSpinCount:     newFreeSpinCount,
		BonusMultiplier:   newBonusMultiplier,
		FreeSpinTriggered: freeSpinTriggered,
		WinningPositions:  winningPositions,
	}

	return c.JSON(response)
}
