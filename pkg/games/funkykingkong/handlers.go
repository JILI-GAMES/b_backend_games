package funkykingkong

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

// SpinHandler handles the slot machine spin
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

	// Validate the request
	if req.ClientID == "" || req.PlayerID == "" || req.BetID == "" || req.GameID == "" {
		log.Printf("Validation error: ClientID, PlayerID, BetID, GameID must not be empty")
		return c.Status(fiber.StatusBadRequest).JSON(SpinResponse{
			Status:  "error",
			Message: "ClientID, PlayerID, BetID, GameID must not be empty",
		})
	}

	if !ValidateBetLevel(req.BetLevel) {
		log.Printf("Validation error: Invalid bet level %d", req.BetLevel)
		return c.Status(fiber.StatusBadRequest).JSON(SpinResponse{
			Status:  "error",
			Message: "Invalid bet level, allowed values are 1, 2, 3",
		})
	}

	if !ValidateBetAmount(req.BetAmount, req.BetLevel) {
		validAmounts := GetValidBetAmounts(req.BetLevel)
		log.Printf("Validation error: Invalid bet amount %f for level %d", req.BetAmount, req.BetLevel)
		return c.Status(fiber.StatusBadRequest).JSON(SpinResponse{
			Status:  "error",
			Message: fmt.Sprintf("Invalid bet amount for level x%d, valid amounts: %v", req.BetLevel, validAmounts),
		})
	}

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

	// Generate guaranteed winning combination first
	internalMultiplier := GetInternalMultiplier(req.BetAmount, req.BetLevel)
	winningReels := GenerateWinningReels()
	potentialWin, winCombination := CalculateWin(winningReels, req.BetLevel, internalMultiplier)

	// Calculate payout multiplier for RNG
	payoutMultiplier := 0.0
	if req.BetAmount > 0 {
		payoutMultiplier = potentialWin / req.BetAmount
	}

	log.Printf("Generated winning reels: %v", winningReels)
	log.Printf("Potential win: %f", potentialWin)
	log.Printf("Win combination: %s", winCombination)
	log.Printf("Payout multiplier: %f", payoutMultiplier)

	// Call the RNG API
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("IP: %v", ip)
	log.Printf("User-Agent: %v", userAgent)

	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount, ip, userAgent)
	if err != nil {
		log.Printf("Error retrieving RNG outcome: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(SpinResponse{
			Status:  "error",
			Message: "Failed to retrieve RNG outcome: " + err.Error(),
		})
	}
	log.Printf("RNG outcome: %s", rngResp.PrefOutcome)

	// Determine final result based on RNG outcome
	var finalReels []string
	var finalWinAmount float64
	var finalWinCombination string

	if rngResp.PrefOutcome == "win" {
		// RNG says win - use the winning combination
		finalReels = winningReels
		finalWinAmount = potentialWin
		finalWinCombination = winCombination
		log.Printf("RNG Win - Using winning reels: %v, Win amount: %f", finalReels, finalWinAmount)
	} else {
		// RNG says loss - force a losing combination
		finalReels = GenerateLosingReels()
		finalWinAmount = 0
		finalWinCombination = ""
		log.Printf("RNG Loss - Using losing reels: %v", finalReels)
	}

	// Build the response
	response := SpinResponse{
		Status:             "success",
		Message:            "",
		Reels:              finalReels,
		WinAmount:          finalWinAmount,
		WinningCombination: finalWinCombination,
		PaytableUsed:       req.BetLevel,
		BetLevel:           req.BetLevel,
	}

	return c.JSON(response)
}
