package onepiece

import (
	"fmt"
	"log"
	"math"

	"github.com/JILI-GAMES/b_backend_games/pkg/common/rng"
	"github.com/gofiber/fiber/v2"
)

var featureBuy bool = false

// / SpinHandler handles the /spin/onepiece endpoint
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
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.BetAmount, req.IsFreeSpin, req.FreeSpinOption, req.CurrentFreeSpinIndex, req.RemainingFreeSpins, req.FreeSpinMultiplier, req.ExtraBonusMultiplier); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Map bet amount to multiplier
	// betMultiplier := BetAmountToMultiplier[req.BetAmount]
	betAmount := req.BetAmount

	// Generate reels with a guaranteed win
	reels := GenerateReelsWithWin()

	// Calculate multiplier for free spins, including the extra bonus multiplier effect
	effectiveMultiplier := req.FreeSpinMultiplier

	if req.IsFreeSpin && req.ExtraBonusMultiplier > 0 {
		// Apply the extra bonus multiplier to the free spin multiplier
		effectiveMultiplier *= (req.ExtraBonusMultiplier)
		featureBuy = true
		log.Printf("Applied extra bonus multiplier: base=%d, extra=%d, effective=%d",
			req.FreeSpinMultiplier, req.ExtraBonusMultiplier, effectiveMultiplier)
	}

	if req.IsFreeSpin {
		featureBuy = true
	}

	log.Printf("Feature Buy✅✅✅: %v", featureBuy)

	// Calculate winnings using the effective multiplier
	totalWinnings, winDetails := CalculateWins(reels, betAmount, effectiveMultiplier, req.IsFreeSpin)
	log.Printf("Initial calculation: totalWinnings=%v, winDetails=%v", totalWinnings, winDetails)

	// Select correct clients for this request
	rngClient, settingsClient := rg.getClientsForRequest(c)

	// Calculate payout multiplier (corrected formula: total_win / bet_amount)
	payoutMultiplier := totalWinnings / req.BetAmount
	if math.IsNaN(payoutMultiplier) || math.IsInf(payoutMultiplier, 0) {
		payoutMultiplier = 0
		log.Printf("Payout multiplier is NaN or Inf, setting to 0")
	}
	log.Printf("Payout multiplier: %v", payoutMultiplier)

	// Get RTP
	rtp, settingsErr := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if settingsErr != nil {
		log.Printf("Failed to get RTP: %v", settingsErr)
	}

	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("✅IP: %v", ip)
	log.Printf("✅User-Agent: %v", userAgent)

	// Call RNG
	var rngResp rng.Response
	var rngErr error
	if settingsErr == nil {
		// Only call RNG if settings API succeeded
		rngResp, rngErr = rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount, ip, userAgent, featureBuy)
		if rngErr != nil {
			log.Printf("Failed to call RNG API: %v", rngErr)
		}
	} else {
		// If settings failed, we can't call RNG with proper RTP
		rngErr = fmt.Errorf("RNG call skipped due to settings API failure")
		log.Printf("Skipping RNG call due to settings API failure")
	}

	// Check if both APIs failed and send Telegram notification
	if settingsErr != nil && rngErr != nil {
		log.Printf("Both Settings and RNG APIs failed - sending Telegram notification")
		if rg.Telegram != nil {
			if telegramErr := rg.Telegram.SendErrorNotification(
				req.GameID, req.ClientID, req.PlayerID, req.BetID,
				settingsErr, rngErr,
			); telegramErr != nil {
				log.Printf("Failed to send Telegram notification: %v", telegramErr)
			}
		} else {
			log.Printf("Telegram client not configured - cannot send notification")
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Critical services unavailable - both settings and RNG APIs failed",
		})
	}

	// Handle individual API failures
	if settingsErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve game settings",
		})
	}

	if rngErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to determine outcome",
		})
	}

	log.Printf("RTP retrieved: %v", rtp)
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

	// Collect scatter positions
	var scatterPositions []Position
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			if reels[reel][row] == string(SymbolScatter) {
				scatterPositions = append(scatterPositions, Position{
					Reel: reel,
					Row:  row,
				})
			}
		}
	}

	freeSpinTriggered := false
	freeSpinRetriggered := false
	extraBonusMultiplier := 0
	if HasScatterOnEachReel(reels) {
		extraBonusMultiplier = CalculateExtraBonusMultiplier(scatterCount)
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
	if freeSpinTriggered {
		// Will be set after the player selects an option
		remainingFreeSpins = 0
		totalFreeSpinsAwarded = 0
	} else if freeSpinRetriggered {
		option := FreeSpinOptions[req.FreeSpinOption]
		remainingFreeSpins += option.TotalSpins
		totalFreeSpinsAwarded += option.TotalSpins
		if remainingFreeSpins > option.MaxSpins {
			remainingFreeSpins = option.MaxSpins
			totalFreeSpinsAwarded = option.MaxSpins
		}
	}

	// Update free spin index and remaining spins
	currentFreeSpinIndex := req.CurrentFreeSpinIndex
	freeSpinMultiplier := req.FreeSpinMultiplier
	if req.IsFreeSpin {
		currentFreeSpinIndex++
		remainingFreeSpins--

		// Update multiplier based on option
		if req.FreeSpinOption > 0 {
			option := FreeSpinOptions[req.FreeSpinOption]

			// For option 5, multiplier is fixed at the initial value (50)
			if req.FreeSpinOption == 5 {
				freeSpinMultiplier = option.InitialMultiplier
			} else {
				// For options 1-4, calculate the multiplier based on the current spin index
				freeSpinMultiplier = option.InitialMultiplier + (currentFreeSpinIndex)*option.MultiplierIncrease
			}
		}
	}
	log.Printf("Free Spin state updated: remainingFreeSpins=%d, totalFreeSpinsAwarded=%d, currentFreeSpinIndex=%d, freeSpinMultiplier=%d", remainingFreeSpins, totalFreeSpinsAwarded, currentFreeSpinIndex, freeSpinMultiplier)

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

	log.Printf("Spin completed: totalWin=%v, freeSpinTriggered=%v, freeSpinRetriggered=%v, isFreeSpin=%v", totalWinnings, freeSpinTriggered, freeSpinRetriggered, isFreeSpin)

	return c.JSON(SpinResponse{
		Reels:                 reels,
		WinAmount:             totalWinnings,
		WinDetails:            winDetails,
		ScatterCount:          scatterCount,
		ScatterPositions:      scatterPositions,
		FreeSpinTriggered:     freeSpinTriggered,
		FreeSpinRetriggered:   freeSpinRetriggered,
		ExtraBonusMultiplier:  extraBonusMultiplier,
		IsFreeSpin:            isFreeSpin,
		RemainingFreeSpins:    remainingFreeSpins,
		CurrentFreeSpinIndex:  currentFreeSpinIndex,
		FreeSpinMultiplier:    freeSpinMultiplier,
		TotalFreeSpinsAwarded: totalFreeSpinsAwarded,
		FreeSpinOption:        req.FreeSpinOption,
	})
}

// SelectFreeSpinOptionHandler handles the /select-free-spin-option/onepiece endpoint
func (rg *RouteGroup) SelectFreeSpinOptionHandler(c *fiber.Ctx) error {
	var req SelectFreeSpinOptionRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateSelectFreeSpinOptionRequest(req.ClientID, req.GameID, req.PlayerID, req.Option, req.ExtraBonusMultiplier); err != nil {
		log.Printf("Request validation failed: %v", err)
		// If option is invalid, default to Option 1 (as per paytable rules)
		if req.Option < 1 || req.Option > 5 {
			log.Printf("Invalid option %d, defaulting to Option 1", req.Option)
			req.Option = 1
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}
	}

	// Get the Free Spin Bonus option
	option, exists := FreeSpinOptions[req.Option]
	if !exists {
		log.Printf("Invalid Free Spin Bonus option: %d", req.Option)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid Free Spin Bonus option",
		})
	}

	// No longer apply extra bonus multiplier to the number of spins
	totalSpins := option.TotalSpins

	initialMultiplier := option.InitialMultiplier

	log.Printf("Free Spin Bonus option selected: option=%d, totalSpins=%d, initialMultiplier=%d, extraBonusMultiplier=%d",
		req.Option, totalSpins, initialMultiplier, req.ExtraBonusMultiplier)

	return c.JSON(SelectFreeSpinOptionResponse{
		TotalFreeSpins:       totalSpins,
		InitialMultiplier:    initialMultiplier,
		MaxFreeSpins:         option.MaxSpins,
		ExtraBonusMultiplier: req.ExtraBonusMultiplier, // Send this back to be used in free spins
	})
}

// validateRequest validates the /spin request fields
func validateRequest(clientID, gameID, playerID, betID string, betAmount float64, isFreeSpin bool, freeSpinOption, currentFreeSpinIndex, remainingFreeSpins, freeSpinMultiplier, extraBonusMultiplier int) error {
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
	if !isFreeSpin && betAmount == 0.0 {
		return fmt.Errorf("invalid bet amount")
	}
	if isFreeSpin {
		if freeSpinOption < 0 || freeSpinOption > 5 {
			return fmt.Errorf("invalid free spin option, allowed values are 0-5")
		}
		if currentFreeSpinIndex < 0 {
			return fmt.Errorf("current_free_spin_index must be non-negative")
		}
		if remainingFreeSpins < 0 {
			return fmt.Errorf("remaining_free_spins must be non-negative")
		}
		if freeSpinMultiplier < 0 {
			return fmt.Errorf("free_spin_multiplier must be non-negative")
		}
		if extraBonusMultiplier < 0 {
			return fmt.Errorf("extra_bonus_multiplier must be non-negative")
		}
	}
	return nil
}

// validateSelectFreeSpinOptionRequest validates the /select-free-spin-option request fields
func validateSelectFreeSpinOptionRequest(clientID, gameID, playerID string, option, extraBonusMultiplier int) error {
	if clientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if gameID == "" {
		return fmt.Errorf("game_id is required")
	}
	if playerID == "" {
		return fmt.Errorf("player_id is required")
	}
	if option < 1 || option > 5 {
		return fmt.Errorf("invalid option, allowed values are 1-5")
	}
	if extraBonusMultiplier < 0 || extraBonusMultiplier > 3 {
		return fmt.Errorf("invalid extra_bonus_multiplier, allowed values are 0-3")
	}
	return nil
}

// isValidBetAmount checks if the bet amount is valid
// func isValidBetAmount(amount float64) bool {
// 	validAmounts := []float64{10.0, 50.0, 100.0, 200.0, 500.0}
// 	for _, valid := range validAmounts {
// 		if amount == valid {
// 			return true
// 		}
// 	}
// 	return false
// }

// package onepiece

// import (
// 	"fmt"
// 	"log"
// 	"math"

// 	"github.com/gofiber/fiber/v2"
// )

// var featureBuy bool = false

// // / SpinHandler handles the /spin/onepiece endpoint
// func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
// 	var req SpinRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		log.Printf("Failed to parse request body: %v", err)
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"status":  "error",
// 			"message": "Invalid request body",
// 		})
// 	}

// 	// Validate request
// 	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.BetAmount, req.IsFreeSpin, req.FreeSpinOption, req.CurrentFreeSpinIndex, req.RemainingFreeSpins, req.FreeSpinMultiplier, req.ExtraBonusMultiplier); err != nil {
// 		log.Printf("Request validation failed: %v", err)
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"status":  "error",
// 			"message": err.Error(),
// 		})
// 	}

// 	// Map bet amount to multiplier
// 	betMultiplier := BetAmountToMultiplier[req.BetAmount]

// 	// Generate reels with a guaranteed win
// 	reels := GenerateReelsWithWin()

// 	// Calculate multiplier for free spins, including the extra bonus multiplier effect
// 	effectiveMultiplier := req.FreeSpinMultiplier

// 	if req.IsFreeSpin && req.ExtraBonusMultiplier > 0 {
// 		// Apply the extra bonus multiplier to the free spin multiplier
// 		effectiveMultiplier *= (req.ExtraBonusMultiplier)
// 		featureBuy = true
// 		log.Printf("Applied extra bonus multiplier: base=%d, extra=%d, effective=%d",
// 			req.FreeSpinMultiplier, req.ExtraBonusMultiplier, effectiveMultiplier)
// 	}

// 	if req.IsFreeSpin {
// 		featureBuy = true
// 	}

// 	log.Printf("Feature Buy✅✅✅: %v", featureBuy)

// 	// Calculate winnings using the effective multiplier
// 	totalWinnings, winDetails := CalculateWins(reels, betMultiplier, effectiveMultiplier, req.IsFreeSpin)
// 	log.Printf("Initial calculation: totalWinnings=%v, winDetails=%v", totalWinnings, winDetails)

// 	// Select correct clients for this request
// 	rngClient, settingsClient := rg.getClientsForRequest(c)

// 	// Calculate payout multiplier (corrected formula: total_win / bet_amount)
// 	payoutMultiplier := totalWinnings / req.BetAmount
// 	if math.IsNaN(payoutMultiplier) || math.IsInf(payoutMultiplier, 0) {
// 		payoutMultiplier = 0
// 		log.Printf("Payout multiplier is NaN or Inf, setting to 0")
// 	}
// 	log.Printf("Payout multiplier: %v", payoutMultiplier)

// 	// Get RTP
// 	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
// 	if err != nil {
// 		log.Printf("Failed to get RTP: %v", err)
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"status":  "error",
// 			"message": "Failed to retrieve game settings",
// 		})
// 	}
// 	log.Printf("RTP retrieved: %v", rtp)

// 	// Get IP address and user agent from request
// 	ip := c.IP()
// 	userAgent := c.Get("User-Agent")

// 	log.Printf("✅IP: %v", ip)
// 	log.Printf("✅User-Agent: %v", userAgent)

// 	// Call RNG
// 	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount, ip, userAgent, featureBuy)
// 	if err != nil {
// 		log.Printf("Failed to call RNG API: %v", err)
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"status":  "error",
// 			"message": "Failed to determine outcome",
// 		})
// 	}
// 	log.Printf("RNG response: %v", rngResp)

// 	// Adjust outcome based on RNG
// 	if rngResp.PrefOutcome == "loss" {
// 		log.Printf("RNG determined a loss outcome")
// 		reels = GenerateLossReels()
// 		totalWinnings = 0
// 		winDetails = nil
// 	}

// 	// Count Scatters and check for Free Spin Bonus trigger/retrigger
// 	scatterCount := CountScatters(reels)

// 	// Collect scatter positions
// 	var scatterPositions []Position
// 	for reel := 0; reel < Reels; reel++ {
// 		for row := 0; row < Rows; row++ {
// 			if reels[reel][row] == string(SymbolScatter) {
// 				scatterPositions = append(scatterPositions, Position{
// 					Reel: reel,
// 					Row:  row,
// 				})
// 			}
// 		}
// 	}

// 	freeSpinTriggered := false
// 	freeSpinRetriggered := false
// 	extraBonusMultiplier := 0
// 	if HasScatterOnEachReel(reels) {
// 		extraBonusMultiplier = CalculateExtraBonusMultiplier(scatterCount)
// 		if !req.IsFreeSpin {
// 			freeSpinTriggered = true
// 			log.Printf("Free Spin Bonus triggered: %d scatters", scatterCount)
// 		} else {
// 			freeSpinRetriggered = true
// 			log.Printf("Free Spin Bonus retriggered: %d scatters", scatterCount)
// 		}
// 	}

// 	// Update Free Spin state
// 	remainingFreeSpins := req.RemainingFreeSpins
// 	totalFreeSpinsAwarded := req.TotalFreeSpinsAwarded
// 	if freeSpinTriggered {
// 		// Will be set after the player selects an option
// 		remainingFreeSpins = 0
// 		totalFreeSpinsAwarded = 0
// 	} else if freeSpinRetriggered {
// 		option := FreeSpinOptions[req.FreeSpinOption]
// 		remainingFreeSpins += option.TotalSpins
// 		totalFreeSpinsAwarded += option.TotalSpins
// 		if remainingFreeSpins > option.MaxSpins {
// 			remainingFreeSpins = option.MaxSpins
// 			totalFreeSpinsAwarded = option.MaxSpins
// 		}
// 	}

// 	// Update free spin index and remaining spins
// 	currentFreeSpinIndex := req.CurrentFreeSpinIndex
// 	freeSpinMultiplier := req.FreeSpinMultiplier
// 	if req.IsFreeSpin {
// 		currentFreeSpinIndex++
// 		remainingFreeSpins--

// 		// Update multiplier based on option
// 		if req.FreeSpinOption > 0 {
// 			option := FreeSpinOptions[req.FreeSpinOption]

// 			// For option 5, multiplier is fixed at the initial value (50)
// 			if req.FreeSpinOption == 5 {
// 				freeSpinMultiplier = option.InitialMultiplier
// 			} else {
// 				// For options 1-4, calculate the multiplier based on the current spin index
// 				freeSpinMultiplier = option.InitialMultiplier + (currentFreeSpinIndex)*option.MultiplierIncrease
// 			}
// 		}
// 	}
// 	log.Printf("Free Spin state updated: remainingFreeSpins=%d, totalFreeSpinsAwarded=%d, currentFreeSpinIndex=%d, freeSpinMultiplier=%d", remainingFreeSpins, totalFreeSpinsAwarded, currentFreeSpinIndex, freeSpinMultiplier)

// 	// Check if Free Spin Bonus has ended
// 	isFreeSpin := req.IsFreeSpin
// 	if req.IsFreeSpin && remainingFreeSpins <= 0 {
// 		isFreeSpin = false
// 		currentFreeSpinIndex = 0
// 		freeSpinMultiplier = 0
// 		remainingFreeSpins = 0
// 		totalFreeSpinsAwarded = 0
// 		log.Printf("Free Spin Bonus ended")
// 	}

// 	log.Printf("Spin completed: totalWin=%v, freeSpinTriggered=%v, freeSpinRetriggered=%v, isFreeSpin=%v", totalWinnings, freeSpinTriggered, freeSpinRetriggered, isFreeSpin)

// 	return c.JSON(SpinResponse{
// 		Reels:                 reels,
// 		WinAmount:             totalWinnings,
// 		WinDetails:            winDetails,
// 		ScatterCount:          scatterCount,
// 		ScatterPositions:      scatterPositions,
// 		FreeSpinTriggered:     freeSpinTriggered,
// 		FreeSpinRetriggered:   freeSpinRetriggered,
// 		ExtraBonusMultiplier:  extraBonusMultiplier,
// 		IsFreeSpin:            isFreeSpin,
// 		RemainingFreeSpins:    remainingFreeSpins,
// 		CurrentFreeSpinIndex:  currentFreeSpinIndex,
// 		FreeSpinMultiplier:    freeSpinMultiplier,
// 		TotalFreeSpinsAwarded: totalFreeSpinsAwarded,
// 		FreeSpinOption:        req.FreeSpinOption,
// 	})
// }

// // SelectFreeSpinOptionHandler handles the /select-free-spin-option/onepiece endpoint
// func (rg *RouteGroup) SelectFreeSpinOptionHandler(c *fiber.Ctx) error {
// 	var req SelectFreeSpinOptionRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		log.Printf("Failed to parse request body: %v", err)
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"status":  "error",
// 			"message": "Invalid request body",
// 		})
// 	}

// 	// Validate request
// 	if err := validateSelectFreeSpinOptionRequest(req.ClientID, req.GameID, req.PlayerID, req.Option, req.ExtraBonusMultiplier); err != nil {
// 		log.Printf("Request validation failed: %v", err)
// 		// If option is invalid, default to Option 1 (as per paytable rules)
// 		if req.Option < 1 || req.Option > 5 {
// 			log.Printf("Invalid option %d, defaulting to Option 1", req.Option)
// 			req.Option = 1
// 		} else {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 				"status":  "error",
// 				"message": err.Error(),
// 			})
// 		}
// 	}

// 	// Get the Free Spin Bonus option
// 	option, exists := FreeSpinOptions[req.Option]
// 	if !exists {
// 		log.Printf("Invalid Free Spin Bonus option: %d", req.Option)
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"status":  "error",
// 			"message": "Invalid Free Spin Bonus option",
// 		})
// 	}

// 	// No longer apply extra bonus multiplier to the number of spins
// 	totalSpins := option.TotalSpins

// 	initialMultiplier := option.InitialMultiplier

// 	log.Printf("Free Spin Bonus option selected: option=%d, totalSpins=%d, initialMultiplier=%d, extraBonusMultiplier=%d",
// 		req.Option, totalSpins, initialMultiplier, req.ExtraBonusMultiplier)

// 	return c.JSON(SelectFreeSpinOptionResponse{
// 		TotalFreeSpins:       totalSpins,
// 		InitialMultiplier:    initialMultiplier,
// 		MaxFreeSpins:         option.MaxSpins,
// 		ExtraBonusMultiplier: req.ExtraBonusMultiplier, // Send this back to be used in free spins
// 	})
// }

// // validateRequest validates the /spin request fields
// func validateRequest(clientID, gameID, playerID, betID string, betAmount float64, isFreeSpin bool, freeSpinOption, currentFreeSpinIndex, remainingFreeSpins, freeSpinMultiplier, extraBonusMultiplier int) error {
// 	if clientID == "" {
// 		return fmt.Errorf("client_id is required")
// 	}
// 	if gameID == "" {
// 		return fmt.Errorf("game_id is required")
// 	}
// 	if playerID == "" {
// 		return fmt.Errorf("player_id is required")
// 	}
// 	if betID == "" {
// 		return fmt.Errorf("bet_id is required")
// 	}
// 	if !isFreeSpin && !isValidBetAmount(betAmount) {
// 		return fmt.Errorf("invalid bet amount, allowed values are 1.0, 5.0, 50.0, 200.0, 500.0")
// 	}
// 	if isFreeSpin {
// 		if freeSpinOption < 0 || freeSpinOption > 5 {
// 			return fmt.Errorf("invalid free spin option, allowed values are 0-5")
// 		}
// 		if currentFreeSpinIndex < 0 {
// 			return fmt.Errorf("current_free_spin_index must be non-negative")
// 		}
// 		if remainingFreeSpins < 0 {
// 			return fmt.Errorf("remaining_free_spins must be non-negative")
// 		}
// 		if freeSpinMultiplier < 0 {
// 			return fmt.Errorf("free_spin_multiplier must be non-negative")
// 		}
// 		if extraBonusMultiplier < 0 {
// 			return fmt.Errorf("extra_bonus_multiplier must be non-negative")
// 		}
// 	}
// 	return nil
// }

// // validateSelectFreeSpinOptionRequest validates the /select-free-spin-option request fields
// func validateSelectFreeSpinOptionRequest(clientID, gameID, playerID string, option, extraBonusMultiplier int) error {
// 	if clientID == "" {
// 		return fmt.Errorf("client_id is required")
// 	}
// 	if gameID == "" {
// 		return fmt.Errorf("game_id is required")
// 	}
// 	if playerID == "" {
// 		return fmt.Errorf("player_id is required")
// 	}
// 	if option < 1 || option > 5 {
// 		return fmt.Errorf("invalid option, allowed values are 1-5")
// 	}
// 	if extraBonusMultiplier < 0 || extraBonusMultiplier > 3 {
// 		return fmt.Errorf("invalid extra_bonus_multiplier, allowed values are 0-3")
// 	}
// 	return nil
// }

// // isValidBetAmount checks if the bet amount is valid
// func isValidBetAmount(amount float64) bool {
// 	validAmounts := []float64{1.0, 5.0, 50.0, 200.0, 500.0}
// 	for _, valid := range validAmounts {
// 		if amount == valid {
// 			return true
// 		}
// 	}
// 	return false
// }
