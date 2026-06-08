package moneybagsman

import (
	"fmt"
	"log"
	"math"

	"github.com/gofiber/fiber/v2"
)

// SpinHandler handles the /spin/moneybagsman endpoint
func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
	var req SpinRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.BetAmount, req.IsFreeSpin); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Map bet amount to multiplier
	betMultiplier := BetAmountToMultiplier[req.BetAmount]

	// Generate reels with a guaranteed win
	reels := GenerateReelsWithWin()

	// Calculate winnings with the appropriate multiplier
	totalWinnings, winDetails := CalculateWins(reels, betMultiplier, req.FreeSpinMultiplier, req.IsFreeSpin)
	log.Printf("Initial calculation: totalWinnings=%v, winDetails=%v", totalWinnings, winDetails)

	// Select correct clients for this request
	rngClient, settingsClient := rg.getClientsForRequest(c)

	// Get RTP
	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if err != nil {
		log.Printf("Failed to get RTP: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve game settings",
		})
	}
	log.Printf("RTP retrieved: %v", rtp)

	// Calculate payout multiplier (total_win / bet_amount)
	payoutMultiplier := totalWinnings / req.BetAmount
	if math.IsNaN(payoutMultiplier) || math.IsInf(payoutMultiplier, 0) {
		payoutMultiplier = 0
		log.Printf("Payout multiplier is NaN or Inf, setting to 0")
	}
	log.Printf("Payout multiplier: %v", payoutMultiplier)

	// Call RNG
	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("✅IP: %v", ip)
	log.Printf("✅User-Agent: %v", userAgent)
	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount, ip, userAgent, false)
	if err != nil {
		log.Printf("Failed to call RNG API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to determine outcome",
		})
	}
	log.Printf("RNG response: %v", rngResp)

	// Adjust outcome based on RNG
	if rngResp.PrefOutcome == "loss" {
		log.Printf("RNG determined a loss outcome")
		reels = GenerateLossReels()
		totalWinnings = 0
		winDetails = nil
	}

	// Count Scatters and check for Free Spin Bonus trigger/retrigger
	scatterCount := CountScatters(reels)
	freeSpinTriggered := false
	freeSpinRetriggered := false
	if HasScatterOnEachReel(reels) {
		if !req.IsFreeSpin {
			freeSpinTriggered = true
			log.Printf("Free Spin Bonus triggered: %d scatters", scatterCount)
		} else {
			freeSpinRetriggered = true
			log.Printf("Free Spin Bonus retriggered: %d scatters", scatterCount)
		}
	}

	// Update Free Spin state
	remainingFreeSpins := req.RemainingFreeSpins
	totalFreeSpinsAwarded := req.TotalFreeSpinsAwarded
	initialMultiplier := 0
	multiplierIncrement := 0
	maxMultiplier := 0

	if freeSpinTriggered {
		// Initial trigger of free spins
		remainingFreeSpins = FreeSpin
		totalFreeSpinsAwarded = FreeSpin

		// Set multiplier based on scatter count
		initialMultiplier, multiplierIncrement, maxMultiplier = GetFreeSpinMultiplierInfo(scatterCount)
		log.Printf("Free Spin Bonus triggered with %d scatters. Initial multiplier: %d, Increment: %d, Max: %d",
			scatterCount, initialMultiplier, multiplierIncrement, maxMultiplier)
	} else if freeSpinRetriggered {
		// Retrigger adds more free spins
		remainingFreeSpins += FreeSpin
		totalFreeSpinsAwarded += FreeSpin

		// Cap at maximum free spins
		if remainingFreeSpins > MaxFreeSpins {
			remainingFreeSpins = MaxFreeSpins
		}

		log.Printf("Free Spin Bonus retriggered. Remaining: %d, Total awarded: %d",
			remainingFreeSpins, totalFreeSpinsAwarded)
	}

	// Update free spin index and remaining spins
	currentFreeSpinIndex := req.CurrentFreeSpinIndex
	freeSpinMultiplier := req.FreeSpinMultiplier

	if req.IsFreeSpin {
		// Increment the free spin index
		currentFreeSpinIndex++
		remainingFreeSpins--

		// Calculate the new multiplier based on the index
		if freeSpinTriggered {
			// If we just triggered, use the initial multiplier
			freeSpinMultiplier = initialMultiplier
		} else {
			// Otherwise, calculate based on existing parameters
			initialMult, increment, maxMult := GetFreeSpinMultiplierInfo(req.ScatterCount)

			// Calculate the multiplier based on current index
			freeSpinMultiplier = initialMult + (currentFreeSpinIndex * increment)

			// Cap the multiplier at the maximum
			if freeSpinMultiplier > maxMult {
				freeSpinMultiplier = maxMult
			}
		}
	} else if freeSpinTriggered {
		// First free spin will use the initial multiplier
		freeSpinMultiplier = initialMultiplier
	}

	log.Printf("Free Spin state updated: remainingFreeSpins=%d, totalFreeSpinsAwarded=%d, currentFreeSpinIndex=%d, freeSpinMultiplier=%d",
		remainingFreeSpins, totalFreeSpinsAwarded, currentFreeSpinIndex, freeSpinMultiplier)

	// Check if Free Spin Bonus has ended
	isFreeSpin := req.IsFreeSpin
	if req.IsFreeSpin && remainingFreeSpins <= 0 {
		isFreeSpin = false
		currentFreeSpinIndex = 0
		freeSpinMultiplier = 0
		remainingFreeSpins = 0
		totalFreeSpinsAwarded = 0
		log.Printf("Free Spin Bonus ended")
	}

	// If free spins were just triggered, set isFreeSpin to true for the next round
	if freeSpinTriggered {
		isFreeSpin = true
	}

	log.Printf("Spin completed: totalWin=%v, freeSpinTriggered=%v, freeSpinRetriggered=%v, isFreeSpin=%v",
		totalWinnings, freeSpinTriggered, freeSpinRetriggered, isFreeSpin)

	// Get multiplier information for response
	initialMult, increment, maxMult := GetFreeSpinMultiplierInfo(scatterCount)
	if !freeSpinTriggered && req.ScatterCount > 0 {
		// Use the existing scatter count if we're in free spins
		initialMult, increment, maxMult = GetFreeSpinMultiplierInfo(req.ScatterCount)
	}
	log.Printf("Multiplier info: initial=%d, increment=%d, max=%d", initialMult, increment, maxMult)

	// Find scatter positions for animation
	scatterPositions := FindScatterPositions(reels)
	log.Printf("Found %d scatter positions for animation", len(scatterPositions))

	return c.JSON(SpinResponse{
		Reels:                 reels,
		WinAmount:             totalWinnings,
		WinDetails:            winDetails,
		ScatterCount:          scatterCount,
		ScatterPositions:      scatterPositions,
		FreeSpinTriggered:     freeSpinTriggered,
		FreeSpinRetriggered:   freeSpinRetriggered,
		IsFreeSpin:            isFreeSpin,
		RemainingFreeSpins:    remainingFreeSpins,
		CurrentFreeSpinIndex:  currentFreeSpinIndex,
		FreeSpinMultiplier:    freeSpinMultiplier,
		TotalFreeSpinsAwarded: totalFreeSpinsAwarded,
		MaxMultiplier:         maxMult,
		MultiplierIncrement:   increment,
	})
}

// validateRequest validates the /spin request fields
func validateRequest(clientID, gameID, playerID, betID string, betAmount float64, isFreeSpin bool) error {
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
	if !isFreeSpin && !isValidBetAmount(betAmount) {
		return fmt.Errorf("invalid bet amount, allowed values are 0.5, 1.0, 1.5, 2.5, 5.0")
	}
	return nil
}

// isValidBetAmount checks if the bet amount is valid
func isValidBetAmount(amount float64) bool {
	validAmounts := []float64{0.5, 1.0, 1.5, 2.5, 5.0}
	for _, valid := range validAmounts {
		if amount == valid {
			return true
		}
	}
	return false
}
