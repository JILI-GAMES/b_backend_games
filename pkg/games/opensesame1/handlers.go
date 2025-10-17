package opensesame1

import (
	"fmt"
	// "log"
	"math"

	"github.com/gofiber/fiber/v2"
)

// SpinHandler handles the /spin/opensesame endpoint
func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
	var req SpinRequest
	if err := c.BodyParser(&req); err != nil {
		rg.GameLogger.Error("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Get bet multiplier from bet amount
	betMultiplier, err := getBetMultiplierFromAmount(req.BetAmount)
	if err != nil {
		rg.GameLogger.Debug("Invalid bet amount: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.BetAmount, req.IsFreeSpin); err != nil {
		rg.GameLogger.Error("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Generate reels with a guaranteed win
	reels := GenerateReelsWithWin()

	// Calculate winnings with the appropriate multiplier
	totalWinnings, winDetails, scatterWinAmount, freeSpinCount, freeSpinWinAmount := CalculateWins(reels, betMultiplier, req.FreeSpinMultiplier, req.IsFreeSpin)
	rg.GameLogger.Debug("Initial calculation: totalWinnings=%v, scatterWinAmount=%v, freeSpinWinAmount=%v, winDetails=%v",
		totalWinnings, scatterWinAmount, freeSpinWinAmount, winDetails)

	// Get positions of scatter and free spin symbols
	scatterPositions := GetSymbolPositions(reels, string(SymbolScatter))
	freeSpinPositions := GetFreeSpinSymbolPositions(reels)

	// Select correct clients for this request
	rngClient, settingsClient := rg.getClientsForRequest(c)

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

	// Calculate total bet amount
	totalBetAmount := req.BetAmount

	// Calculate payout multiplier (total_win / bet_amount)
	payoutMultiplier := (totalWinnings + scatterWinAmount + freeSpinWinAmount) / totalBetAmount
	if math.IsNaN(payoutMultiplier) || math.IsInf(payoutMultiplier, 0) {
		payoutMultiplier = 0
		rg.GameLogger.Debug("Payout multiplier is NaN or Inf, setting to 0")
	}
	rg.GameLogger.Debug("Payout multiplier: %v", payoutMultiplier)

	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	rg.GameLogger.Debug("IP: %v", ip)
	rg.GameLogger.Debug("User-Agent: %v", userAgent)

	// Call RNG
	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, totalBetAmount, ip, userAgent, false)
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
		reels = GenerateLossReels()
		totalWinnings = 0
		winDetails = nil

		// Recalculate scatter and free spin wins (they can still win on a "loss" outcome)
		_, _, scatterWinAmount, freeSpinCount, freeSpinWinAmount = CalculateWins(reels, betMultiplier, req.FreeSpinMultiplier, req.IsFreeSpin)

		// Update positions for scatter and free spin symbols
		scatterPositions = GetSymbolPositions(reels, string(SymbolScatter))
		freeSpinPositions = GetFreeSpinSymbolPositions(reels)
	}

	// Check for free spin trigger/retrigger
	freeSpinTriggered := false
	freeSpinRetriggered := false

	if HasFreeSpinSymbolsOnFirstThreeReels(reels) {
		if !req.IsFreeSpin {
			freeSpinTriggered = true
			rg.GameLogger.Debug("Free Spin Bonus triggered")
		} else {
			freeSpinRetriggered = true
			rg.GameLogger.Debug("Free Spin Bonus retriggered")
		}
	}

	// Update Free Spin state
	remainingFreeSpins := req.RemainingFreeSpins
	totalFreeSpinsAwarded := req.TotalFreeSpinsAwarded
	freeSpinMultiplier := req.FreeSpinMultiplier

	// Handle free spin retriggering
	if freeSpinRetriggered {
		// Add the same number of free spins as initially awarded
		additionalSpins := 0

		// If this is the first retrigger, use the initial number of spins
		if req.CurrentFreeSpinIndex > 0 {
			additionalSpins = req.TotalFreeSpinsAwarded / req.CurrentFreeSpinIndex
		} else {
			// Default to FreeSpin if we can't determine the original count
			additionalSpins = 12
		}

		remainingFreeSpins += additionalSpins
		totalFreeSpinsAwarded += additionalSpins

		// Cap at maximum free spins
		if remainingFreeSpins > MaxTotalFreeSpins {
			remainingFreeSpins = MaxTotalFreeSpins
		}
	}

	// Update free spin index and remaining spins
	currentFreeSpinIndex := req.CurrentFreeSpinIndex
	if req.IsFreeSpin {
		currentFreeSpinIndex++
		remainingFreeSpins--
	}

	// Check if Free Spin Bonus has ended
	isFreeSpin := req.IsFreeSpin
	if req.IsFreeSpin && remainingFreeSpins <= 0 {
		isFreeSpin = false
		currentFreeSpinIndex = 0
		freeSpinMultiplier = 0
		remainingFreeSpins = 0
		totalFreeSpinsAwarded = 0
		rg.GameLogger.Debug("Free Spin Bonus ended")
	}

	// Calculate total win amount
	totalWinAmount := totalWinnings + scatterWinAmount + freeSpinWinAmount

	rg.GameLogger.Debug("Spin completed: regularWin=%v, scatterWin=%v, freeSpinSymbolWin=%v, totalWin=%v, freeSpinTriggered=%v, freeSpinRetriggered=%v, isFreeSpin=%v",
		totalWinnings, scatterWinAmount, freeSpinWinAmount, totalWinAmount, freeSpinTriggered, freeSpinRetriggered, isFreeSpin)

	return c.JSON(SpinResponse{
		Reels:                 reels,
		WinAmount:             totalWinAmount,
		WinDetails:            winDetails,
		ScatterCount:          len(scatterPositions),
		ScatterWinAmount:      scatterWinAmount,
		ScatterPositions:      scatterPositions,
		FreeSpinCount:         freeSpinCount,
		FreeSpinWinAmount:     freeSpinWinAmount,
		FreeSpinPositions:     freeSpinPositions,
		FreeSpinTriggered:     freeSpinTriggered,
		FreeSpinRetriggered:   freeSpinRetriggered,
		IsFreeSpin:            isFreeSpin,
		RemainingFreeSpins:    remainingFreeSpins,
		CurrentFreeSpinIndex:  currentFreeSpinIndex,
		FreeSpinMultiplier:    freeSpinMultiplier,
		TotalFreeSpinsAwarded: totalFreeSpinsAwarded,
		BetAmount:             req.BetAmount,
		BetMultiplier:         betMultiplier,
	})
}

// SelectFreeSpinOptionHandler handles the /select-free-spin-option/opensesame endpoint
func (rg *RouteGroup) SelectFreeSpinOptionHandler(c *fiber.Ctx) error {
	var req SelectOptionRequest
	if err := c.BodyParser(&req); err != nil {
		rg.GameLogger.Error("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateSelectOptionRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.ChestIndex, req.LampIndex); err != nil {
		rg.GameLogger.Error("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Get free spin options based on player selection
	freeSpinCount, multiplier, err := GetSelectedFreeSpinOptions(req.ChestIndex, req.LampIndex)
	if err != nil {
		rg.GameLogger.Debug("Failed to get free spin options: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	rg.GameLogger.Debug("Free Spin options selected: count=%d, multiplier=%d", freeSpinCount, multiplier)

	return c.JSON(SelectOptionResponse{
		FreeSpinCount:      freeSpinCount,
		FreeSpinMultiplier: multiplier,
	})
}

// getBetMultiplierFromAmount converts bet amount to bet multiplier
func getBetMultiplierFromAmount(betAmount float64) (int, error) {
	multiplier, exists := BetAmountMap[betAmount]
	if !exists {
		return 0, fmt.Errorf("invalid bet amount, allowed values are 0.25, 0.5, 1.25, 2.5, 6.25")
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
		return fmt.Errorf("invalid bet amount, allowed values are 0.25, 0.5, 1.25, 2.5, 6.25")
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

// isValidBetAmount checks if the bet amount is valid
func isValidBetAmount(amount float64) bool {
	validAmounts := []float64{0.25, 0.5, 1.25, 2.5, 6.25}
	for _, valid := range validAmounts {
		if amount == valid {
			return true
		}
	}
	return false
}
