package magicace

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"log"


	"github.com/gofiber/fiber/v2"
	
)

// SpinHandler handles the /spin/magicace endpoint
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
    if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.GameState.Bet.Amount); err != nil {
        log.Printf("Request validation failed: %v", err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": err.Error(),
        })
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

    log.Printf("##_____________##Game Mode: %s, Booming Multiplier: %d", req.GameState.GameMode, req.GameState.BoomingMultiplier)

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
    rtp, err := rg.Settings.GetRTP(req.ClientID, req.GameID, req.PlayerID)
    if err != nil {
        log.Printf("Failed to get RTP: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Failed to retrieve game settings",
        })
    }
    log.Printf("RTP retrieved: %v", rtp)

    // Call RNG
    payoutMultiplier := totalWinnings / req.GameState.Bet.Amount
    rngResp, err := rg.RNG.GetOutcome(req.ClientID, req.GameID, req.PlayerID, rtp, payoutMultiplier, req.GameState.Bet.Amount)
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
        reels, specialSymbols = GenerateLossReels(req.GameState.JokerCards, r)
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
        log.Printf("Free Spins triggered: %d scatter(s)", req.GameState.ScatterCount)
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
        log.Printf("Free Spins retriggered: %d scatter(s), %d additional spins, total %d spins", req.GameState.ScatterCount, additionalSpins, req.GameState.FreeSpins.TotalAwarded)
    }

    // Update Free Spins
    if req.GameState.GameMode == "freeSpins" {
        req.GameState.FreeSpins.Remaining--
        if req.GameState.FreeSpins.Remaining <= 0 {
            log.Printf("Free Spins ended")
            req.GameState.GameMode = "base"
            req.GameState.FreeSpins = struct {
                Remaining        int `json:"remaining"`
                TotalAwarded     int `json:"totalAwarded"`
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
        if wasInWin {
            if joker.RemainingRounds > 1 {
                joker.RemainingRounds--
                updatedJokerCards = append(updatedJokerCards, joker)
            }
            // If RemainingRounds == 1, remove it (already handled by not appending)
        } else {
            joker.RemainingRounds--
            if joker.RemainingRounds > 0 {
                updatedJokerCards = append(updatedJokerCards, joker)
            }
        }
    }
    req.GameState.JokerCards = updatedJokerCards
    req.GameState.SpecialSymbols.JokerCards = updatedJokerCards
    log.Printf("Updated Joker Cards: %d", len(req.GameState.JokerCards))

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

    log.Printf("Spin completed: totalWin=%v, cascading=%v, gameMode=%s, totalCost=%v", totalWinnings, req.GameState.Cascading, req.GameState.GameMode, totalCost)

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
    var req CascadeRequest
    if err := c.BodyParser(&req); err != nil {
        log.Printf("Failed to parse request body: %v", err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid request body",
        })
    }

    // Validate request
    if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.GameState.Bet.Amount); err != nil {
        log.Printf("Request validation failed: %v", err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": err.Error(),
        })
    }

    // Validate reels
    if len(req.GameState.Reels) != Reels || len(req.GameState.Reels[0]) != Rows {
        log.Printf("Invalid reels dimensions")
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": "Invalid reels",
        })
    }

    // Fix 3: Increment CascadeCount
    req.GameState.CascadeCount++

    // Fix 1: Update Booming Multiplier based on CascadeCount
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
        req.GameState.BoomingMultiplier = lastMultiplier + (index - len(boomingMultipliers) + 1) * step
    }
    log.Printf("Updated Booming Multiplier: %d (Cascade %d)", req.GameState.BoomingMultiplier, req.GameState.CascadeCount)

    // Create a single rand.Rand instance for this request.
    r := rand.New(rand.NewSource(time.Now().UnixNano()))

    // Transform Golden Cards if present in the last win
    newJokerCards := TransformGoldenCards(req.GameState.Reels, req.GameState.LastWinDetails, r)
    req.GameState.JokerCards = append(req.GameState.JokerCards, newJokerCards...)
    req.GameState.SpecialSymbols.JokerCards = req.GameState.JokerCards
    log.Printf("Transformed Golden Cards into %d Joker Cards", len(newJokerCards))

    // Determine positions to replace
    winningPositions := make(map[Position]bool)
    for _, win := range req.GameState.LastWinDetails {
        for i, pos := range win.Payline {
            symbol := win.Symbols[i]
            fmt.Println("Symbol to replace:", symbol)
            winningPositions[pos] = true
        }
    }

    // Generate new symbols for the cascade
    newReels, specialSymbols := GenerateReelsForCascade(req.GameState.Reels, winningPositions, req.GameState.JokerCards, r)

    // Calculate new wins
    payout, winDetails := CalculateWins(newReels, req.GameState.Bet.Multiplier, req.GameState.BoomingMultiplier, req.GameState.JokerCards)

    // Call RNG (as per your requirement, we keep this unchanged)
    rtp, err := rg.Settings.GetRTP(req.ClientID, req.GameID, req.PlayerID)
    if err != nil {
        log.Printf("Failed to get RTP: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Failed to retrieve game settings",
        })
    }
    log.Printf("RTP retrieved: %v", rtp)
    
    payoutMultiplier := payout / req.GameState.Bet.Amount
    rngResp, err := rg.RNG.GetOutcome(req.ClientID, req.GameID, req.PlayerID, rtp, payoutMultiplier, req.GameState.Bet.Amount)
    if err != nil {
        log.Printf("Failed to call RNG API: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Failed to determine outcome",
        })
    }
    log.Printf("RNG response: %v", rngResp)

    // Adjust outcome based on RNG (unchanged as per your requirement)
    if rngResp.PrefOutcome == "loss" {
        log.Printf("RNG determined a loss outcome")
        newReels, specialSymbols = GenerateLossForCascade(req.GameState.Reels, winningPositions, req.GameState.JokerCards, r)
        payout = 0
        winDetails = nil
    }

    // Update game state
    req.GameState.Reels = newReels
    req.GameState.SpecialSymbols = specialSymbols
    req.GameState.TotalWin = payout
    req.GameState.LastWinDetails = winDetails
    req.GameState.Cascading = len(winDetails) > 0

    // Update ScatterCount (includes all Scatters)
    req.GameState.ScatterCount = len(specialSymbols.TargetSymbols)

    // Fix 2: Check for Free Spins retrigger in Free Spins mode (only with new Scatters)
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
        log.Printf("Free Spins retriggered in cascade: %d new scatter(s), %d additional spins, total %d spins", newScatterCount, additionalSpins, req.GameState.FreeSpins.TotalAwarded)
    }

    // Update Joker Cards after the cascade
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
        if wasInWin {
            if joker.RemainingRounds > 1 {
                joker.RemainingRounds--
                updatedJokerCards = append(updatedJokerCards, joker)
            }
            // If RemainingRounds == 1, remove it (already handled by not appending)
        } else {
            joker.RemainingRounds--
            if joker.RemainingRounds > 0 {
                updatedJokerCards = append(updatedJokerCards, joker)
            }
        }
    }
    req.GameState.JokerCards = updatedJokerCards
    req.GameState.SpecialSymbols.JokerCards = updatedJokerCards
    log.Printf("Updated Joker Cards: %d", len(req.GameState.JokerCards))

    log.Printf("Cascade completed: totalWin=%v, cascading=%v, gameMode=%s", req.GameState.TotalWin, req.GameState.Cascading, req.GameState.GameMode)

    // Fix 4: Use the correct response type (CascadeResponse)
    return c.JSON(CascadeResponse{
        Status:     "success",
        Message:    "",
        GameState:  req.GameState,
        WinDetails: winDetails,
        TotalCost:  req.GameState.Bet.Amount, // No Extra Bet cost in cascade
    })
}

// FeatureBuyHandler handles the /featureBuy/magicace endpoint
func (rg *RouteGroup) FeatureBuyHandler(c *fiber.Ctx) error {
    var req FeatureBuyRequest
    if err := c.BodyParser(&req); err != nil {
        log.Printf("Failed to parse request body: %v", err)
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
    if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.GameState.Bet.Amount); err != nil {
        log.Printf("Request validation failed: %v", err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "error",
            "message": err.Error(),
        })
    }

    // Validate option
    if req.Option != 1 && req.Option != 2 {
        log.Printf("Invalid option, allowed values are 1 or 2")
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
        for _, joker := range req.GameState.JokerCards {
            occupiedPositions[joker.Position] = true
        }
        for reel := 0; reel < Reels; reel++ {
            for row := 0; row < Rows; row++ {
                symbol := req.GameState.Reels[reel][row]
                if strings.Contains(symbol, "Joker") || strings.HasPrefix(symbol, "golden_") || symbol == string(SymbolScatter) {
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
                RemainingRounds: 3,
            }
            if err := newJoker.ValidateMode(); err != nil {
                log.Printf("Invalid Joker Card mode in FeatureBuy: %v", err)
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "status":  "error",
                    "message": "Failed to create Super Joker",
                })
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

    log.Printf("Feature Buy completed: option=%d, cost=%.2f", req.Option, cost)

    return c.JSON(FeatureBuyResponse{
        Status:     "success",
        Message:    fmt.Sprintf("Feature Buy successful, cost: %.2f", cost),
        GameState:  req.GameState,
        WinDetails: req.GameState.LastWinDetails,
        TotalCost:  cost,
    })
}

// validateRequest validates the request fields
func validateRequest(clientID, gameID, playerID string, betAmount float64) error {
    if clientID == "" {
        return fmt.Errorf("client_id is required")
    }
    if gameID == "" {
        return fmt.Errorf("game_id is required")
    }
    if playerID == "" {
        return fmt.Errorf("player_id is required")
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
