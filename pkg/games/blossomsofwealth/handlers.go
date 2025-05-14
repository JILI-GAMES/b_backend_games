package blossomsofwealth

import (
    "fmt"
    "log"
    "math"

    "github.com/gofiber/fiber/v2"
)

// SpinHandler handles the /spin/blossomsofwealth endpoint
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
    if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.BetAmount, req.IsFreeSpin, req.CurrentFreeSpinIndex, req.RemainingFreeSpins, req.FreeSpinMultiplier); err != nil {
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

    // Count Silver Flowers and get positions
    silverFlowerPositions := GetSilverFlowerPositions(reels)
    silverFlowerCount := len(silverFlowerPositions)
    
    // Count Gold Flowers and get positions
    goldFlowerPositions := GetGoldFlowerPositions(reels)
    goldFlowerCount := len(goldFlowerPositions)
    
    // Check for free spin trigger - Silver Flowers on reels 0, 1, and 2
    freeSpinTriggered := HasSilverFlowerOnEachFirstReel(reels)
    if freeSpinTriggered {
        log.Printf("Free Spin Bonus triggered: %d silver flowers", silverFlowerCount)
    }
    
    // Check for bonus multiplier - requires free spins triggered AND Gold Flower on reel 3
    bonusMultiplier := 0
    if !req.IsFreeSpin && CheckBonusMultiplierCondition(reels) {
        bonusMultiplier = GenerateRandomBonusMultiplier()
        log.Printf("Bonus Multiplier triggered: %dx (free spins + gold flower on reel 3)", bonusMultiplier)
    }
    
    // Calculate winnings - if in free spin mode, apply the current free spin multiplier
    // Otherwise, if bonus multiplier is triggered, apply that
    effectiveMultiplier := 1
    
    if req.IsFreeSpin {
        // In free spin mode, use the free spin multiplier
        effectiveMultiplier = req.FreeSpinMultiplier
    } else if bonusMultiplier > 0 {
        // In main game with bonus multiplier, use that
        effectiveMultiplier = bonusMultiplier
    }
    
    // Apply appropriate multiplier to calculate total potential win
    totalWinnings, winDetails := CalculateWins(reels, betMultiplier, effectiveMultiplier, req.IsFreeSpin)
    log.Printf("Initial calculation: totalWinnings=%v, winDetails=%v, effectiveMultiplier=%v", 
        totalWinnings, winDetails, effectiveMultiplier)

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

    // Calculate payout multiplier with all multipliers already applied
    payoutMultiplier := totalWinnings / req.BetAmount
    if math.IsNaN(payoutMultiplier) || math.IsInf(payoutMultiplier, 0) {
        payoutMultiplier = 0
        log.Printf("Payout multiplier is NaN or Inf, setting to 0")
    }
    log.Printf("Payout multiplier: %v", payoutMultiplier)

    // Call RNG with the total potential win (including all multipliers)
    rngResp, err := rg.RNG.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount)
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
        
        // We still need to check if free spins can be triggered on a loss
        silverFlowerPositions = GetSilverFlowerPositions(reels)
        silverFlowerCount = len(silverFlowerPositions)
        
        goldFlowerPositions = GetGoldFlowerPositions(reels)
        goldFlowerCount = len(goldFlowerPositions)
        
        // Check for free spin trigger with strict Silver Flower requirements
        freeSpinTriggered = HasSilverFlowerOnEachFirstReel(reels)
        if freeSpinTriggered {
            log.Printf("Free Spin Bonus triggered on loss: %d silver flowers", silverFlowerCount)
        }
        
        // Check for bonus multiplier with strict conditions
        bonusMultiplier = 0
        if !req.IsFreeSpin && CheckBonusMultiplierCondition(reels) {
            bonusMultiplier = GenerateRandomBonusMultiplier()
            log.Printf("Bonus Multiplier triggered on loss: %dx (free spins + gold flower on reel 3)", bonusMultiplier)
        }
    }
    
    // Update Free Spin state
    remainingFreeSpins := req.RemainingFreeSpins
    totalFreeSpinsAwarded := req.TotalFreeSpinsAwarded
    currentFreeSpinIndex := req.CurrentFreeSpinIndex
    freeSpinMultiplier := req.FreeSpinMultiplier

    // Handle free spin trigger from main game
    if freeSpinTriggered && !req.IsFreeSpin {
        // Set initial multiplier based on bonus multiplier from gold flowers
        if bonusMultiplier > 0 {
            freeSpinMultiplier = bonusMultiplier
        } else {
            freeSpinMultiplier = 1  // Default to 1 if no gold flowers
        }
        
        // Calculate free spins based on Silver Flower count
        newFreeSpins := 0
        for i := 0; i < silverFlowerCount; i++ {
            newFreeSpins += GenerateRandomFreeSpins()
        }
        
        remainingFreeSpins = newFreeSpins
        totalFreeSpinsAwarded = newFreeSpins
        
        // Cap at maximum free spins
        if totalFreeSpinsAwarded > MaxTotalFreeSpins {
            remainingFreeSpins = MaxTotalFreeSpins
            totalFreeSpinsAwarded = MaxTotalFreeSpins
        }
        
        log.Printf("Free Spin Bonus started with %d spins and multiplier %d", 
            remainingFreeSpins, freeSpinMultiplier)
    // Handle free spin retrigger during free spins
    } else if freeSpinTriggered && req.IsFreeSpin { 
        // Just add more spins when retriggered during free spins
        newFreeSpins := 0
        for i := 0; i < silverFlowerCount; i++ {
            newFreeSpins += GenerateRandomFreeSpins()
        }
        
        remainingFreeSpins += newFreeSpins
        totalFreeSpinsAwarded += newFreeSpins
        
        // Cap at maximum free spins
        if totalFreeSpinsAwarded > MaxTotalFreeSpins {
            remainingFreeSpins = MaxTotalFreeSpins - (totalFreeSpinsAwarded - remainingFreeSpins)
            totalFreeSpinsAwarded = MaxTotalFreeSpins
        }
        
        log.Printf("Free Spin Bonus retriggered, adding %d spins. Total: %d", 
            newFreeSpins, remainingFreeSpins)
    }
    
    // For an active free spin, update the spin count and multiplier for next round
    if req.IsFreeSpin {
        currentFreeSpinIndex++
        remainingFreeSpins--
        
        // Increase multiplier by 1 for the next round
        freeSpinMultiplier = req.FreeSpinMultiplier + 1
        
        log.Printf("Free Spin in progress: spin=%d, remaining=%d, multiplier=%d", 
            currentFreeSpinIndex, remainingFreeSpins, freeSpinMultiplier)
    }
    
    // Check if Free Spin Bonus has ended
    isFreeSpin := req.IsFreeSpin
    if req.IsFreeSpin && remainingFreeSpins <= 0 {
        isFreeSpin = false
        currentFreeSpinIndex = 0
        freeSpinMultiplier = 0
        log.Printf("Free Spin Bonus ended")
    }
    
    log.Printf("Spin completed: totalWin=%v, freeSpinTriggered=%v, isFreeSpin=%v, bonusMultiplier=%v", 
        totalWinnings, freeSpinTriggered, isFreeSpin, bonusMultiplier)

    return c.JSON(SpinResponse{
        Reels:                 reels,
        WinAmount:             totalWinnings,
        WinDetails:            winDetails,
        SilverFlowerCount:     silverFlowerCount,
        SilverFlowerPositions: silverFlowerPositions,
        GoldFlowerCount:       goldFlowerCount,
        GoldFlowerPositions:   goldFlowerPositions,
        BonusMultiplier:       bonusMultiplier,
        FreeSpinTriggered:     freeSpinTriggered,
        IsFreeSpin:            isFreeSpin,
        RemainingFreeSpins:    remainingFreeSpins,
        CurrentFreeSpinIndex:  currentFreeSpinIndex,
        FreeSpinMultiplier:    freeSpinMultiplier,
        TotalFreeSpinsAwarded: totalFreeSpinsAwarded,
    })
}

// validateRequest validates the /spin request fields
func validateRequest(clientID, gameID, playerID, betID string, betAmount float64, isFreeSpin bool, currentFreeSpinIndex, remainingFreeSpins, freeSpinMultiplier int) error {
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
    if isFreeSpin {
        if currentFreeSpinIndex < 0 {
            return fmt.Errorf("current_free_spin_index must be non-negative")
        }
        if remainingFreeSpins < 0 {
            return fmt.Errorf("remaining_free_spins must be non-negative")
        }
        if freeSpinMultiplier < 0 {
            return fmt.Errorf("free_spin_multiplier must be non-negative")
        }
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