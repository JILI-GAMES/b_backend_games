package crazykingkong

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// CrushHandler handles the boulder crushing game
func (rg *RouteGroup) CrushHandler(c *fiber.Ctx) error {
	// Parse the request
	var req CrushRequest
	if err := c.BodyParser(&req); err != nil {
		rg.GameLogger.Debug("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(CrushResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate the request
	if req.ClientID == "" || req.PlayerID == "" || req.BetID == "" || req.GameID == "" {
		rg.GameLogger.Debug("Validation error: ClientID, PlayerID, BetID, GameID must not be empty")
		return c.Status(fiber.StatusBadRequest).JSON(CrushResponse{
			Status:  "error",
			Message: "ClientID, PlayerID, BetID, GameID must not be empty",
		})
	}

	if !ValidateBetAmount(req.BetAmount) {
		rg.GameLogger.Debug("Validation error: Invalid bet amount %f", req.BetAmount)
		return c.Status(fiber.StatusBadRequest).JSON(CrushResponse{
			Status:  "error",
			Message: "Invalid bet amount, allowed values are 0.5, 1, 2, 4, 5, 10, 20, 25, 50, 100",
		})
	}

	if !ValidateBoulderType(req.BoulderType) {
		rg.GameLogger.Debug("Validation error: Invalid boulder type %s", req.BoulderType)
		return c.Status(fiber.StatusBadRequest).JSON(CrushResponse{
			Status:  "error",
			Message: "Invalid boulder type, allowed values are gold, blue, red, white",
		})
	}

	// Select correct clients for this request
	rngClient, settingsClient := rg.getClientsForRequest(c)

	// Call the Settings API to get RTP
	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if err != nil {
		rg.GameLogger.Debug("Error retrieving game settings: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(CrushResponse{
			Status:  "error",
			Message: "Failed to retrieve game settings: " + err.Error(),
		})
	}
	rg.GameLogger.Debug("Retrieved RTP: %f", rtp)

	// Generate potential multiplier for the chosen boulder
	multiplier := GenerateBoulderMultiplier(req.BoulderType)
	potentialWin := req.BetAmount * multiplier
	payoutMultiplier := multiplier

	log.Printf("Generated multiplier: %f for boulder type: %s", multiplier, req.BoulderType)
	log.Printf("Potential win: %f", potentialWin)

	// Call the RNG API to determine if boulder breaks
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("IP: %v", ip)
	log.Printf("User-Agent: %v", userAgent)

	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount, ip, userAgent, false)
	if err != nil {
		rg.GameLogger.Debug("Error retrieving RNG outcome: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(CrushResponse{
			Status:  "error",
			Message: "Failed to retrieve RNG outcome: " + err.Error(),
		})
	}
	log.Printf("RNG outcome: %s", rngResp.PrefOutcome)

	// Determine boulder break result based on RNG outcome
	var boulderBroken bool
	var winAmount float64
	var responseMessage string
	var bonusTriggered bool
	var availableStones []Stone

	if rngResp.PrefOutcome == "win" {
		// Boulder breaks - player wins
		boulderBroken = true
		winAmount = potentialWin
		responseMessage = "Boulder crushed! Choose next boulder."

		// Check if bonus game should be triggered (only when boulder breaks)
		bonusTriggered = ShouldTriggerBonus()
		if bonusTriggered {
			availableStones = GenerateAvailableStones()
			responseMessage = "Boulder crushed! Bonus game triggered - choose a stone!"
			rg.GameLogger.Debug("Bonus game triggered! Available stones: %v", availableStones)
		}

		rg.GameLogger.Debug("Boulder broken! Multiplier: %f, Win amount: %f", multiplier, winAmount)
	} else {
		// Boulder doesn't break - player loses bet
		boulderBroken = false
		multiplier = 0
		winAmount = 0
		responseMessage = "Boulder didn't break, try again!"
		bonusTriggered = false
		availableStones = nil

		rg.GameLogger.Debug("Boulder didn't break - no win")
	}

	// Build the response
	response := CrushResponse{
		Status:          "success",
		Message:         responseMessage,
		BoulderType:     string(req.BoulderType),
		BoulderBroken:   boulderBroken,
		Multiplier:      multiplier,
		WinAmount:       winAmount,
		BonusTriggered:  bonusTriggered,
		AvailableStones: availableStones,
	}

	return c.JSON(response)
}

// BonusGameHandler handles the bonus stone selection game
func (rg *RouteGroup) BonusGameHandler(c *fiber.Ctx) error {
	// Parse the request
	var req BonusGameRequest
	if err := c.BodyParser(&req); err != nil {
		rg.GameLogger.Debug("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(BonusGameResponse{
			Status:  "error",
			Message: "Invalid request body",
		})
	}

	// Validate the request
	if req.ClientID == "" || req.PlayerID == "" || req.BetID == "" || req.GameID == "" {
		rg.GameLogger.Debug("Validation error: ClientID, PlayerID, BetID, GameID must not be empty")
		return c.Status(fiber.StatusBadRequest).JSON(BonusGameResponse{
			Status:  "error",
			Message: "ClientID, PlayerID, BetID, GameID must not be empty",
		})
	}

	if !ValidateBetAmount(req.BetAmount) {
		rg.GameLogger.Debug("Validation error: Invalid bet amount %f", req.BetAmount)
		return c.Status(fiber.StatusBadRequest).JSON(BonusGameResponse{
			Status:  "error",
			Message: "Invalid bet amount, allowed values are 0.5, 1, 2, 4, 5, 10, 20, 25, 50, 100",
		})
	}

	if !ValidateStoneType(req.StoneType) {
		rg.GameLogger.Debug("Validation error: Invalid stone type %s", req.StoneType)
		return c.Status(fiber.StatusBadRequest).JSON(BonusGameResponse{
			Status:  "error",
			Message: "Invalid stone type, allowed values are gold, silver, bronze",
		})
	}

	// Select correct clients for this request
	rngClient, settingsClient := rg.getClientsForRequest(c)

	// Call the Settings API to get RTP
	rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if err != nil {
		rg.GameLogger.Debug("Error retrieving game settings: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(BonusGameResponse{
			Status:  "error",
			Message: "Failed to retrieve game settings: " + err.Error(),
		})
	}

	// Generate multiplier for the chosen stone
	multiplier := GenerateStoneMultiplier(req.StoneType)
	potentialWin := req.BetAmount * multiplier
	payoutMultiplier := multiplier

	log.Printf("Generated stone multiplier: %f for stone type: %s", multiplier, req.StoneType)
	log.Printf("Potential bonus win: %f", potentialWin)

	// Call the RNG API for bonus game outcome
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount, ip, userAgent, false)
	if err != nil {
		rg.GameLogger.Debug("Error retrieving RNG outcome: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(BonusGameResponse{
			Status:  "error",
			Message: "Failed to retrieve RNG outcome: " + err.Error(),
		})
	}
	// print full rng response
	log.Printf("RNG outcome: %+v", rngResp)

	// Apply the RNG outcome for bonus game
	winAmount := potentialWin
	if rngResp.PrefOutcome == "loss" {
		// For bonus game loss, give minimum multiplier for this stone type
		multiplierRange := StoneMultipliers[req.StoneType]
		multiplier = multiplierRange.Min
		winAmount = req.BetAmount * multiplier
		rg.GameLogger.Debug("Bonus game loss - reduced to minimum multiplier: %f", multiplier)
	} else {
		rg.GameLogger.Debug("Bonus game win - full multiplier: %f", multiplier)

	}

	// Build the response
	response := BonusGameResponse{
		Status:     "success",
		Message:    "Bonus stone revealed!",
		StoneType:  string(req.StoneType),
		Multiplier: multiplier,
		WinAmount:  winAmount,
	}

	return c.JSON(response)
}
