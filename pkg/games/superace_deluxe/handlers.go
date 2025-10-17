package superace_deluxe

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
	var rowReq RowBasedSpinRequest
	if err := c.BodyParser(&rowReq); err != nil {
		rg.GameLogger.Info("Error parsing row-based request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid request format",
		})
	}

	// Validate request parameters
	if rowReq.BetAmount < 0.5 || (rowReq.Action != "SPIN" && rowReq.Action != "TRANSFORM") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid bet amount or action",
		})
	}
	// validate gameid,clientid,playerid,betid
	if rowReq.Game.ID == "" || rowReq.ClientID == "" || rowReq.PlayerID == "" || rowReq.BetID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid game ID, client ID, player ID or bet ID",
		})
	}

	// Create a game state using the internal reel-based format
	gs := NewGameState(rowReq.BetAmount, rowReq.Game.Mode)
	gs.FreeSpins = rowReq.FreeSpins

	// Only set ComboMultiplier and convert Cards for TRANSFORM actions
	if rowReq.Action == "TRANSFORM" {
		gs.ComboMultiplier = rowReq.ComboMultiplier
		// Convert row-based cards from request to reel-based format for internal processing
		gs.Cards = ConvertToReelBased(rowReq.Cards)
	}

	// Select correct clients for this request
	rngClient, settingsClient := rg.getClientsForRequest(c)

	// Use the shared settings service
	rtp, err := settingsClient.GetRTP(rowReq.ClientID, rowReq.Game.ID, rowReq.PlayerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  500,
			"message": "Failed to fetch RTP",
		})
	}

	fmt.Printf("RTP: %v\n", rtp)

	// Generate grid and calculate potential wins inside Spin/Transform
	var potentialWins float64
	if rowReq.Action == "SPIN" {
		potentialWins = gs.Spin()
	} else if rowReq.Action == "TRANSFORM" {
		potentialWins = gs.Transform()
	}

	payoutMultiplier := potentialWins / rowReq.BetAmount
	fmt.Printf("Potential wins: %v, Payout multiplier: %v\n", potentialWins, payoutMultiplier)

	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	rg.GameLogger.Debug("IP: %v", ip)
	rg.GameLogger.Debug("User-Agent: %v", userAgent)

	// Use the shared RNG service
	rngResp, err := rngClient.GetOutcome(rowReq.ClientID, rowReq.Game.ID, rowReq.PlayerID, rowReq.BetID, rtp, payoutMultiplier, rowReq.BetAmount, ip, userAgent, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  500,
			"message": "Failed to fetch RNG outcome",
		})
	}

	fmt.Printf("RNG response: {pref_outcome: %s, win_amount: %v, win_prob: %v}\n", rngResp.PrefOutcome, rngResp.WinAmount, rngResp.WinProb)

	// Convert RNG response to game-specific format and apply outcome
	gameRNGResp := RNGResponse{
		PrefOutcome: rngResp.PrefOutcome,
		WinAmount:   rngResp.WinAmount,
		WinProb:     rngResp.WinProb,
	}

	// Apply RNG outcome
	gs.ApplyRNGOutcome(gameRNGResp, rowReq.Action)

	maxPayout := rowReq.BetAmount * 10000
	if gs.AmountWon > maxPayout {
		gs.AmountWon = maxPayout
		rg.GameLogger.Info("Max payout reached: %v", maxPayout)
	}

	// Convert the internal reel-based game state to row-based for response
	rowBasedResponse := RowBasedSpinResponse{
		Status:  200,
		Message: "Success",
		Data:    gs.ConvertToRowBased(),
	}

	// Return row-based response
	return c.JSON(rowBasedResponse)
}
