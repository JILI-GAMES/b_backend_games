package opensesame2

import (
	"fmt"
	"log"
	"math"

	"github.com/gofiber/fiber/v2"
)

// SpinHandler handles the /spin/opensesame2 endpoint
func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
	var req SpinRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Get bet multiplier from bet amount
	betMultiplier, err := getBetMultiplierFromAmount(req.BetAmount)
	if err != nil {
		log.Printf("Invalid bet amount: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
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

	// Select correct clients for this request
	rngClient, settingsClient := rg.getClientsForRequest(c)

	// Generate reels with a guaranteed win
	reels := GenerateReelsWithWin()

	// Calculate winnings with the appropriate multiplier - Updated for combination payouts
	totalWinnings, winDetails, scatterWinAmount, combinationWinAmount, combinationType, combinationPositions, scatterPositions :=
		CalculateWins(reels, betMultiplier, req.FreeSpinMultiplier, req.ExtraFreeSpinMultiplier, req.IsFreeSpin, req.IsExtraFreeSpin)

	log.Printf("Initial calculation: totalWinnings=%v, scatterWinAmount=%v, combinationWinAmount=%v, combinationType=%s, winDetails=%v",
		totalWinnings, scatterWinAmount, combinationWinAmount, combinationType, winDetails)

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

	// Calculate total bet amount
	totalBetAmount := req.BetAmount

	// Calculate payout multiplier (total_win / bet_amount) - Updated to include combination wins
	payoutMultiplier := (totalWinnings + scatterWinAmount + combinationWinAmount) / totalBetAmount
	if math.IsNaN(payoutMultiplier) || math.IsInf(payoutMultiplier, 0) {
		payoutMultiplier = 0
		log.Printf("Payout multiplier is NaN or Inf, setting to 0")
	}
	log.Printf("Payout multiplier: %v", payoutMultiplier)

	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("✅IP: %v", ip)
	log.Printf("✅User-Agent: %v", userAgent)

	// Call RNG
	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, totalBetAmount, ip, userAgent)
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

		// Recalculate scatter and combination wins (they can still win on a "loss" outcome)
		_, _, scatterWinAmount, combinationWinAmount, combinationType, combinationPositions, scatterPositions =
			CalculateWins(reels, betMultiplier, req.FreeSpinMultiplier, req.ExtraFreeSpinMultiplier, req.IsFreeSpin, req.IsExtraFreeSpin)
	}

	// Check for free spin trigger/retrigger
	freeSpinTriggered := false
	freeSpinRetriggered := false
	extraFreeSpinTriggered := false

	if HasFreeSpinSymbolsOnFirstThreeReels(reels) {
		if !req.IsFreeSpin && !req.IsExtraFreeSpin {
			freeSpinTriggered = true
			log.Printf("Regular Free Spin Bonus triggered")
		} else {
			freeSpinRetriggered = true
			log.Printf("Free Spin Bonus retriggered")
		}
	}

	// Check for extra free spin trigger
	if CheckForExtraFreeSpinTrigger(reels) && !req.IsFreeSpin && !req.IsExtraFreeSpin {
		extraFreeSpinTriggered = true
		log.Printf("Extra Free Spin Bonus triggered")
	}

	// Update Free Spin state
	remainingFreeSpins := req.RemainingFreeSpins
	totalFreeSpinsAwarded := req.TotalFreeSpinsAwarded
	freeSpinMultiplier := req.FreeSpinMultiplier
	extraFreeSpinMultiplier := req.ExtraFreeSpinMultiplier
	isExtraFreeSpin := req.IsExtraFreeSpin

	// Handle free spin retriggering
	if freeSpinRetriggered {
		// Add the same number of free spins as initially awarded during the Gold Chest selection
		additionalSpins := req.TotalFreeSpinsAwarded

		remainingFreeSpins += additionalSpins
		totalFreeSpinsAwarded += additionalSpins

		// Cap at maximum free spins
		if remainingFreeSpins > MaxTotalFreeSpins {
			remainingFreeSpins = MaxTotalFreeSpins
		}
	}

	// Update free spin index and remaining spins
	currentFreeSpinIndex := req.CurrentFreeSpinIndex
	if req.IsFreeSpin || req.IsExtraFreeSpin {
		currentFreeSpinIndex++
		remainingFreeSpins--
	}

	// Check if Free Spin Bonus has ended
	isFreeSpin := req.IsFreeSpin
	if (req.IsFreeSpin || req.IsExtraFreeSpin) && remainingFreeSpins <= 0 {
		isFreeSpin = false
		isExtraFreeSpin = false
		currentFreeSpinIndex = 0
		freeSpinMultiplier = 0
		extraFreeSpinMultiplier = 0
		remainingFreeSpins = 0
		totalFreeSpinsAwarded = 0
		log.Printf("Free Spin Bonus ended")
	}

	// Calculate total win amount - Updated to include combination wins
	totalWinAmount := totalWinnings + scatterWinAmount + combinationWinAmount

	// Determine combination details for response
	var combinationCount int
	var combinationSymbols []string

	if combinationType == "FreeSpinCombination" {
		combinationCount = 3 // 3 Free Spin symbols
		combinationSymbols = []string{"FreeSpins", "FreeSpins", "FreeSpins"}
	} else if combinationType == "MysteryBoxCombination" {
		combinationCount = 3 // 2 Free Spin + 1 Mystery Box
		combinationSymbols = []string{"FreeSpins", "FreeSpins", "MysteryBox"}
	}

	log.Printf("Spin completed: regularWin=%v, scatterWin=%v, combinationWin=%v, combinationType=%s, totalWin=%v, freeSpinTriggered=%v, extraFreeSpinTriggered=%v, freeSpinRetriggered=%v, isFreeSpin=%v, isExtraFreeSpin=%v",
		totalWinnings, scatterWinAmount, combinationWinAmount, combinationType, totalWinAmount, freeSpinTriggered, extraFreeSpinTriggered, freeSpinRetriggered, isFreeSpin, isExtraFreeSpin)

	return c.JSON(SpinResponse{
		Reels:                   reels,
		WinAmount:               totalWinAmount,
		WinDetails:              winDetails,
		ScatterCount:            len(scatterPositions),
		ScatterWinAmount:        scatterWinAmount,
		ScatterPositions:        scatterPositions,
		CombinationCount:        combinationCount,
		CombinationWinAmount:    combinationWinAmount,
		CombinationType:         combinationType,
		CombinationPositions:    combinationPositions,
		CombinationSymbols:      combinationSymbols,
		FreeSpinTriggered:       freeSpinTriggered,
		FreeSpinRetriggered:     freeSpinRetriggered,
		ExtraFreeSpinTriggered:  extraFreeSpinTriggered,
		IsFreeSpin:              isFreeSpin,
		IsExtraFreeSpin:         isExtraFreeSpin,
		RemainingFreeSpins:      remainingFreeSpins,
		CurrentFreeSpinIndex:    currentFreeSpinIndex,
		FreeSpinMultiplier:      freeSpinMultiplier,
		ExtraFreeSpinMultiplier: extraFreeSpinMultiplier,
		TotalFreeSpinsAwarded:   totalFreeSpinsAwarded,
		BetAmount:               req.BetAmount,
		BetMultiplier:           betMultiplier,
	})
}

// SelectFreeSpinOptionHandler handles the /select-free-spin-option/opensesame2 endpoint
func (rg *RouteGroup) SelectFreeSpinOptionHandler(c *fiber.Ctx) error {
	var req SelectOptionRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Check if this is for extra free spin options
	if req.IsExtraBonus {
		// Validate request for extra bonus
		if err := validateExtraOptionRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.TreasureIndex); err != nil {
			log.Printf("Request validation failed: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		// Get extra free spin options based on player selection
		extraMultiplier, extraFreeSpins, isExtraMultiplier, err := GetSelectedExtraOption(req.TreasureIndex)
		if err != nil {
			log.Printf("Failed to get extra free spin options: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		log.Printf("Extra Free Spin options selected: extraMultiplier=%d, extraFreeSpins=%d, isExtraMultiplier=%v",
			extraMultiplier, extraFreeSpins, isExtraMultiplier)

		return c.JSON(SelectOptionResponse{
			FreeSpinCount:      0,
			FreeSpinMultiplier: 0,
			ExtraMultiplier:    extraMultiplier,
			ExtraFreeSpins:     extraFreeSpins,
			IsExtraMultiplier:  isExtraMultiplier,
		})
	} else {
		// Validate regular free spin option request
		if err := validateSelectOptionRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.ChestIndex, req.LampIndex); err != nil {
			log.Printf("Request validation failed: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		// Get free spin options based on player selection
		freeSpinCount, multiplier, err := GetSelectedFreeSpinOptions(req.ChestIndex, req.LampIndex)
		if err != nil {
			log.Printf("Failed to get free spin options: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		log.Printf("Free Spin options selected: count=%d, multiplier=%d", freeSpinCount, multiplier)

		return c.JSON(SelectOptionResponse{
			FreeSpinCount:      freeSpinCount,
			FreeSpinMultiplier: multiplier,
			ExtraMultiplier:    0,
			ExtraFreeSpins:     0,
			IsExtraMultiplier:  false,
		})
	}
}

// getBetMultiplierFromAmount converts bet amount to bet multiplier
func getBetMultiplierFromAmount(betAmount float64) (int, error) {
	multiplier, exists := BetAmountMap[betAmount]
	if !exists {
		return 0, fmt.Errorf("invalid bet amount, allowed values are 0.6, 1.2, 3.0, 6.0, 15.0")
	}
	return multiplier, nil
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
		return fmt.Errorf("invalid bet amount, allowed values are 0.6, 1.2, 3.0, 6.0, 15.0")
	}
	return nil
}

// validateSelectOptionRequest validates the /select-free-spin-option request fields
func validateSelectOptionRequest(clientID, gameID, playerID, betID string, chestIndex, lampIndex int) error {
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
	if chestIndex < 0 || chestIndex >= len(ChestOptions) {
		return fmt.Errorf("invalid chest index, must be between 0 and %d", len(ChestOptions)-1)
	}
	if lampIndex < 0 || lampIndex >= len(LampOptions) {
		return fmt.Errorf("invalid lamp index, must be between 0 and %d", len(LampOptions)-1)
	}
	return nil
}

// validateExtraOptionRequest validates the /select-free-spin-option request fields for extra bonus
func validateExtraOptionRequest(clientID, gameID, playerID, betID string, treasureIndex int) error {
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
	if treasureIndex < 0 || treasureIndex >= 10 { // 5 multiplier options + 5 free spin options
		return fmt.Errorf("invalid treasure index, must be between 0 and 9")
	}
	return nil
}

// isValidBetAmount checks if the bet amount is valid
func isValidBetAmount(amount float64) bool {
	validAmounts := []float64{0.6, 1.2, 3.0, 6.0, 15.0}
	for _, valid := range validAmounts {
		if amount == valid {
			return true
		}
	}
	return false
}
