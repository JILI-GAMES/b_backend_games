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

    // Generate reels with a guaranteed win
    reels := GenerateReelsWithWin()

    // Calculate winnings
    totalWinnings, winDetails, bonusCount, maskReelCount, allSpecialPositions := CalculateWins(reels, betMultiplier, req.IsFreeSpin)
    log.Printf("Initial calculation: totalWinnings=%v, bonusCount=%v, maskReelCount=%v, winDetails=%v, allSpecialPositions=%v",
        totalWinnings, bonusCount, maskReelCount, winDetails, allSpecialPositions)

    // Get positions of bonus and mask reel symbols
    bonusPositions := GetSymbolPositions(reels, string(SymbolBonus))
    maskReelPositions := GetMaskReelPositions(reels)

    // Get RTP
    rtp, err := rg.Settings.GetRTP(req.ClientID, req.GameID, req.PlayerID)
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

    // Calculate payout multiplier (total_win / bet_amount)
    payoutMultiplier := totalWinnings / totalBetAmount
    if math.IsNaN(payoutMultiplier) || math.IsInf(payoutMultiplier, 0) {
        payoutMultiplier = 0
        log.Printf("Payout multiplier is NaN or Inf, setting to 0")
    }
    log.Printf("Payout multiplier: %v", payoutMultiplier)

    // Call RNG
    rngResp, err := rg.RNG.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, totalBetAmount)
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

        // Update positions for bonus and mask reel symbols
        bonusPositions = GetSymbolPositions(reels, string(SymbolBonus))
        maskReelPositions = GetMaskReelPositions(reels)
    }

    // Check for bonus triggers
    freeSpinTriggered := false
    freeSpinRetriggered := false
    maskReelTriggered := false

    // Check for Free Spin Bonus trigger (3+ bonus symbols on reels 1-3)
    if HasBonusOnFirstThreeReels(reels) {
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
    if !req.IsFreeSpin && HasMaskReelOnLastThreeReels(reels) {
        maskReelTriggered = true
        log.Printf("Mask Reel Bonus triggered")
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

    // Calculate total win amount
    totalWinAmount := totalWinnings

    log.Printf("Spin completed: regularWin=%v, totalWin=%v, freeSpinTriggered=%v, freeSpinRetriggered=%v, maskReelTriggered=%v, isFreeSpin=%v",
        totalWinnings, totalWinAmount, freeSpinTriggered, freeSpinRetriggered, maskReelTriggered, isFreeSpin)

    return c.JSON(SpinResponse{
        Reels:                 reels,
        WinAmount:             totalWinAmount,
        WinDetails:            winDetails,
        BonusCount:            len(bonusPositions),
        BonusPositions:        bonusPositions,
        MaskReelCount:         len(maskReelPositions),
        MaskReelPositions:     maskReelPositions,
        FreeSpinTriggered:     freeSpinTriggered,
        FreeSpinRetriggered:   freeSpinRetriggered,
        MaskReelTriggered:     maskReelTriggered,
        IsFreeSpin:            isFreeSpin,
        RemainingFreeSpins:    remainingFreeSpins,
        CurrentFreeSpinIndex:  currentFreeSpinIndex,
        TotalFreeSpinsAwarded: totalFreeSpinsAwarded,
        BetAmount:             req.BetAmount,
        BetMultiplier:         betMultiplier,
    })
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

    // Get RTP for mask bonus
    rtp, err := rg.Settings.GetRTP(req.ClientID, req.GameID, req.PlayerID)
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

    // Call RNG to determine if we can award the full multiplier
    rngResp, err := rg.RNG.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount)
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
        // Award minimum multiplier (2x or 3x)
        minMultipliers := []int{2, 3}
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
        return 0, fmt.Errorf("invalid bet amount, allowed values are 0.5, 1.0, 2.5, 5.0, 12.5")
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
        return fmt.Errorf("invalid bet amount, allowed values are 0.5, 1.0, 2.5, 5.0, 12.5")
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
        return fmt.Errorf("invalid bet amount, allowed values are 0.5, 1.0, 2.5, 5.0, 12.5")
    }
    return nil
}

// isValidBetAmount checks if the bet amount is valid
func isValidBetAmount(amount float64) bool {
    validAmounts := []float64{0.5, 1.0, 2.5, 5.0, 12.5}
    for _, valid := range validAmounts {
        if amount == valid {
            return true
        }
    }
    return false
}