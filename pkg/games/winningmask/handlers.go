package winningmask

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/gofiber/fiber/v2"
)

// SpinHandler handles the /spin/winningmask endpoint
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

	var response SpinResponse

	if req.IsFreeSpin {
		log.Printf("Processing free spin %d with two-stage mask transformation logic", req.CurrentFreeSpinIndex+1)

		// Handle two-stage mask transformation for free spins
		selectedScenario, err := HandleTwoStageMaskTransformation(betMultiplier, req, rngClient, rtp)
		if err != nil {
			log.Printf("Error in two-stage mask transformation: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to process mask transformation",
			})
		}

		// Build response from selected scenario
		response = SpinResponse{
			// Stage 1 Results
			Stage1Reels:      selectedScenario.Stage1Reels,
			Stage1WinAmount:  selectedScenario.Stage1Win,
			Stage1WinDetails: selectedScenario.Stage1Details,

			// Stage 2 Results (if transformation occurred)
			Stage2WinAmount:  selectedScenario.Stage2Win,
			Stage2WinDetails: selectedScenario.Stage2Details,

			// Combined Results
			TotalWinAmount:         selectedScenario.TotalWin,
			MaskTransformationUsed: selectedScenario.HasTransform,

			// Basic game info
			BetAmount:     req.BetAmount,
			BetMultiplier: betMultiplier,
		}

		// Add Stage 2 reels and mask type if transformation occurred
		if selectedScenario.HasTransform {
			response.Stage2Reels = selectedScenario.Stage2Reels
			response.SelectedMaskType = selectedScenario.MaskType

			// Filter Stage2WinDetails and Stage2WinAmount to only include wins for the selected mask type
			filteredDetails := []WinDetail{}
			total := 0.0
			for _, wd := range selectedScenario.Stage2Details {
				if wd.Symbol == selectedScenario.MaskType {
					filteredDetails = append(filteredDetails, wd)
					total += wd.Payout
				}
			}
			response.Stage2WinDetails = filteredDetails
			response.Stage2WinAmount = total
			if response.MaskTransformationUsed {
				response.TotalWinAmount = response.Stage1WinAmount + response.Stage2WinAmount
			}
		}

		log.Printf("Free spin result: Stage1=%v, Stage2=%v, Total=%v, Transform=%v, Mask=%s",
			selectedScenario.Stage1Win, selectedScenario.Stage2Win, selectedScenario.TotalWin,
			selectedScenario.HasTransform, selectedScenario.MaskType)

	} else {
		log.Printf("Processing regular base game spin")

		// Generate reels with a guaranteed win for base game
		reels := GenerateReelsWithWin(false)

		// Calculate winnings for base game (no mask transformation in base game)
		totalWinnings, winDetails, _, _, _ := CalculateWins(reels, betMultiplier, false)
		log.Printf("Base game calculation: totalWinnings=%v, winDetails=%v", totalWinnings, winDetails)

		// Calculate payout multiplier (total_win / bet_amount)
		payoutMultiplier := totalWinnings / req.BetAmount
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

		// Call RNG for base game
		rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount, ip, userAgent, false)
		if err != nil {
			log.Printf("Failed to call RNG API: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to determine outcome",
			})
		}
		log.Printf("RNG response: %v", rngResp)

		// Adjust outcome based on RNG for base game
		if rngResp.PrefOutcome == "loss" {
			log.Printf("RNG determined a loss outcome")
			reels = GenerateLossReels(false)
			totalWinnings = 0
			winDetails = nil
		}

		// Build response for base game
		response = SpinResponse{
			// For base game, Stage1 is the only stage
			Stage1Reels:      reels,
			Stage1WinAmount:  totalWinnings,
			Stage1WinDetails: winDetails,

			// No Stage2 for base game
			Stage2WinAmount:  0,
			Stage2WinDetails: nil,

			// Combined is same as Stage1 for base game
			TotalWinAmount:         totalWinnings,
			MaskTransformationUsed: false,

			// Basic game info
			BetAmount:     req.BetAmount,
			BetMultiplier: betMultiplier,
		}
	}

	// Get positions of bonus and mask reel symbols from Stage1 reels
	bonusPositions := GetSymbolPositions(response.Stage1Reels, string(SymbolBonus))
	maskReelPositions := GetMaskReelPositions(response.Stage1Reels)

	// Check for bonus triggers (based on Stage1 reels)
	freeSpinTriggered := false
	freeSpinRetriggered := false
	maskReelTriggered := false

	// Check for Free Spin Bonus trigger (3+ bonus symbols on reels 1-3)
	if HasBonusOnFirstThreeReels(response.Stage1Reels) {
		if !req.IsFreeSpin {
			freeSpinTriggered = true
			log.Printf("Free Spin Bonus triggered")
		} else {
			freeSpinRetriggered = true
			log.Printf("Free Spin Bonus retriggered")
		}
	}

	// Check for Mask Reel Bonus trigger (3+ mask symbols on reels 3-5)
	// Note: Mask Reel Bonus does not appear in Free Spin Bonus
	if !req.IsFreeSpin && HasMaskReelOnLastThreeReels(response.Stage1Reels) {
		maskReelTriggered = true
		log.Printf("Mask Reel Bonus triggered")
	}

	// Calculate bonus payout from Stage1 reels
	bonusWinAmount := 0.0
	if len(bonusPositions) >= 3 {
		bonusPayValue := float64(Paytable[SymbolBonus][3])
		totalBetAmount := float64(betMultiplier*CreditMultiplier) * Denomination
		bonusWinAmount = bonusPayValue * totalBetAmount
		bonusWinAmount = math.Round(bonusWinAmount*100) / 100
		log.Printf("Bonus payout: count=%d, odds=%v, betMultiplier=%d, totalBetAmount=%v, payout=%v",
			len(bonusPositions), bonusPayValue, betMultiplier, totalBetAmount, bonusWinAmount)
	}

	// Update Free Spin state
	remainingFreeSpins := req.RemainingFreeSpins
	totalFreeSpinsAwarded := req.TotalFreeSpinsAwarded

	// Handle free spin triggering
	if freeSpinTriggered {
		remainingFreeSpins = FreeSpinCount
		totalFreeSpinsAwarded = FreeSpinCount
	}

	// Handle free spin retriggering
	if freeSpinRetriggered {
		remainingFreeSpins += FreeSpinCount
		totalFreeSpinsAwarded += FreeSpinCount

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
		remainingFreeSpins = 0
		totalFreeSpinsAwarded = 0
		log.Printf("Free Spin Bonus ended")
	}

	// Complete the response with bonus and state information
	response.BonusCount = len(bonusPositions)
	response.BonusWinAmount = bonusWinAmount
	response.BonusPositions = bonusPositions
	response.MaskReelCount = len(maskReelPositions)
	response.MaskReelPositions = maskReelPositions
	response.FreeSpinTriggered = freeSpinTriggered
	response.FreeSpinRetriggered = freeSpinRetriggered
	response.MaskReelTriggered = maskReelTriggered
	response.IsFreeSpin = isFreeSpin
	response.RemainingFreeSpins = remainingFreeSpins
	response.CurrentFreeSpinIndex = currentFreeSpinIndex
	response.TotalFreeSpinsAwarded = totalFreeSpinsAwarded

	log.Printf("Spin completed: Stage1=%v, Stage2=%v, Total=%v, Transform=%v, freeSpinTriggered=%v, freeSpinRetriggered=%v, maskReelTriggered=%v",
		response.Stage1WinAmount, response.Stage2WinAmount, response.TotalWinAmount, response.MaskTransformationUsed,
		freeSpinTriggered, freeSpinRetriggered, maskReelTriggered)

	return c.JSON(response)
}

// MaskReelBonusHandler handles the /mask-reel-bonus/winningmask endpoint
func (rg *RouteGroup) MaskReelBonusHandler(c *fiber.Ctx) error {
	var req MaskReelBonusRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateMaskReelBonusRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.BetAmount); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Select correct clients for this request
	rngClient, settingsClient := rg.getClientsForRequest(c)

	// Get RTP for mask bonus
	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if err != nil {
		log.Printf("Failed to get RTP for mask bonus: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve game settings",
		})
	}

	// Generate potential mask multiplier (2x to 200x)
	potentialMultiplier := GenerateMaskReelMultiplier()
	potentialWinAmount := float64(potentialMultiplier) * req.BetAmount
	potentialWinAmount = math.Round(potentialWinAmount*100) / 100

	// Calculate payout multiplier for RNG
	payoutMultiplier := potentialWinAmount / req.BetAmount

	log.Printf("Mask Reel Bonus RNG call: potentialMultiplier=%d, potentialWin=%v, payoutMultiplier=%v",
		potentialMultiplier, potentialWinAmount, payoutMultiplier)

	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("✅IP: %v", ip)
	log.Printf("✅User-Agent: %v", userAgent)	

	// Call RNG to determine if we can award the full multiplier
	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount, ip, userAgent, false)
	if err != nil {
		log.Printf("Failed to call RNG API for mask bonus: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to determine mask bonus outcome",
		})
	}

	log.Printf("Mask Bonus RNG response: %v", rngResp)

	var actualMultiplier int
	var actualWinAmount float64

	if rngResp.PrefOutcome == "win" {
		// Award the full potential multiplier
		actualMultiplier = potentialMultiplier
		actualWinAmount = potentialWinAmount
		log.Printf("RNG approved full mask bonus: multiplier=%d, winAmount=%v", actualMultiplier, actualWinAmount)
	} else {
		// Award minimum multiplier (5x or 10x)
		minMultipliers := []int{5, 10}
		actualMultiplier = minMultipliers[rand.Intn(len(minMultipliers))]
		actualWinAmount = float64(actualMultiplier) * req.BetAmount
		actualWinAmount = math.Round(actualWinAmount*100) / 100
		log.Printf("RNG declined full bonus, awarding minimum: multiplier=%d, winAmount=%v", actualMultiplier, actualWinAmount)
	}

	return c.JSON(MaskReelBonusResponse{
		Multiplier: actualMultiplier,
		WinAmount:  actualWinAmount,
	})
}

// getBetMultiplierFromAmount converts bet amount to bet multiplier
func getBetMultiplierFromAmount(betAmount float64) (int, error) {
	multiplier, exists := BetAmountMap[betAmount]
	if !exists {
		return 0, fmt.Errorf("invalid bet amount, allowed values are 10,15,20,250")
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
		return fmt.Errorf("invalid bet amount, allowed values are 10,15,20,250")
	}
	return nil
}

// validateMaskReelBonusRequest validates the /mask-reel-bonus request fields
func validateMaskReelBonusRequest(clientID, gameID, playerID, betID string, betAmount float64) error {
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
		return fmt.Errorf("invalid bet amount, allowed values are 10,15,20,250")
	}
	return nil
}

// isValidBetAmount checks if the bet amount is valid
func isValidBetAmount(amount float64) bool {
	validAmounts := []float64{10,15,20,250}
	for _, valid := range validAmounts {
		if amount == valid {
			return true
		}
	}
	return false
}
