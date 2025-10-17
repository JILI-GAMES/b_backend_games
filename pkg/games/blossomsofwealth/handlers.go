package blossomsofwealth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/JILI-GAMES/b_backend_games/pkg/common/rng"
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

	// Select correct clients for this request
	rngClient, settingsClient := rg.getClientsForRequest(c)

	// Get RTP
	rtp, settingsErr := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
	if settingsErr != nil {
		log.Printf("Failed to get RTP: %v", settingsErr)
		rtp = 0.9
		log.Printf("Using default RTP: %v", rtp)

		// Capture values from context before starting goroutine
		origin := c.Get("Origin")
		log.Printf("🫠🫠Origin received: %v", origin)

		// Send to Joe's endpoint with environment-aware label
		go func() {
			label := "rng"
			if len(origin) > 0 && (strings.Contains(strings.ToLower(origin), "test") || origin == "") {
				label = "rng-test"
			}

			joePayload := map[string]interface{}{
				"endpoint":     "https://t2.ibibe.africa/get-game-settings",
				"label":        label,
				"status":       "fail",
				"other_status": "settings api fail",
				"source":       "blossomsofwealth",
				"time":         time.Now().Format("2006-01-02T15:04"),
			}

			jsonData, err := json.Marshal(joePayload)
			if err != nil {
				log.Printf("Error marshaling Joe notification payload: %v", err)
				return
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Post("https://queue.ibibe.africa/proxy/queue/manageFails", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Error sending notification to Joe's endpoint: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Joe's endpoint returned non-200 status: %d", resp.StatusCode)
			} else {
				log.Printf("Successfully sent notification to Joe's endpoint")
			}
		}()

		if rg.Telegram != nil {
			notificationText := fmt.Sprintf(`
🚨 <b>Settings API Failure Alert Blossoms of Wealth</b> 🚨

<b>Game:</b> %s
<b>Client ID:</b> %s
<b>Player ID:</b> %s
<b>Bet ID:</b> %s
<b>Timestamp:</b> %s

<b>Error:</b> Settings(RTP) API failed: %v

<b>Action Taken:</b> Using default RTP (0.9) and forcing loss outcome
			`, req.GameID, req.ClientID, req.PlayerID, req.BetID, time.Now().Format("2006-01-02 15:04:05 UTC"), settingsErr)

			if telegramErr := rg.Telegram.SendMessage(notificationText); telegramErr != nil {
				log.Printf("Failed to send Telegram notification: %v", telegramErr)
			}
		} else {
			log.Printf("Telegram client not configured - cannot send notification")
		}
	}

	// Calculate payout multiplier with all multipliers already applied
	payoutMultiplier := totalWinnings / req.BetAmount
	if math.IsNaN(payoutMultiplier) || math.IsInf(payoutMultiplier, 0) {
		payoutMultiplier = 0
		log.Printf("Payout multiplier is NaN or Inf, setting to 0")
	}
	log.Printf("Payout multiplier: %v", payoutMultiplier)

	// Call RNG with the total potential win (including all multipliers)
	// Get IP address and user agent from request
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("✅IP: %v", ip)
	log.Printf("✅User-Agent: %v", userAgent)

	var rngResp rng.Response
	var rngErr error
	if settingsErr == nil {
		rngResp, rngErr = rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount, ip, userAgent, false)
		if rngErr != nil {
			log.Printf("Failed to call RNG API: %v", rngErr)
		}
	} else {
		rngResp = rng.Response{PrefOutcome: "loss"}
		rngErr = nil
		log.Printf("‼️‼️‼️Forcing loss outcome due to settings API failure")
	}

	if rngErr != nil {
		log.Printf("RNG API failed - sending Joe notification")

		// Capture values from context before starting goroutine
		origin := c.Get("Origin")
		log.Printf("🫠🫠Origin received: %v", origin)

		// Send to Joe's endpoint with environment-aware label
		go func() {
			label := "rng"
			if len(origin) > 0 && (strings.Contains(strings.ToLower(origin), "test") || origin == "") {
				label = "rng-test"
			}

			log.Printf("Joe endpoint: %v", label)

			joePayload := map[string]interface{}{
				"endpoint":     "http://159.89.235.166:17003/api/proxy/rng/1",
				"label":        label,
				"status":       "fail",
				"other_status": "rng api fail",
				"source":       "blossomsofwealth",
				"time":         time.Now().Format("2006-01-02T15:04"),
			}

			jsonData, err := json.Marshal(joePayload)
			if err != nil {
				log.Printf("Error marshaling Joe notification payload: %v", err)
				return
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Post("https://queue.ibibe.africa/proxy/queue/manageFails", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Error sending notification to Joe's endpoint: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Joe's endpoint returned non-200 status: %d", resp.StatusCode)
			} else {
				log.Printf("Successfully sent notification to Joe's endpoint")
			}
		}()

		if rg.Telegram != nil {
			notificationText := fmt.Sprintf(`
🚨 <b>RNG API Failure Alert Blossoms of Wealth</b> 🚨

<b>Game:</b> %s
<b>Client ID:</b> %s
<b>Player ID:</b> %s
<b>Bet ID:</b> %s
<b>Timestamp:</b> %s

<b>Error:</b> RNG(Outcome) API failed: %v

<b>Action Taken:</b> Forcing loss outcome to continue game
			`, req.GameID, req.ClientID, req.PlayerID, req.BetID, time.Now().Format("2006-01-02 15:04:05 UTC"), rngErr)

			if telegramErr := rg.Telegram.SendMessage(notificationText); telegramErr != nil {
				log.Printf("Failed to send Telegram notification: %v", telegramErr)
			}
		} else {
			log.Printf("Telegram client not configured - cannot send notification")
		}

		rngResp = rng.Response{PrefOutcome: "loss"}
		rngErr = nil
		log.Printf("‼️‼️‼️Forcing loss outcome due to RNG API failure")
	}

	log.Printf("RTP retrieved: %v", rtp)
	log.Printf("RNG response: %v", rngResp)

	// Adjust outcome based on RNG
	if rngResp.PrefOutcome == "loss" {
		log.Printf("RNG determined a loss outcome")
		reels = GenerateLossReels()
		totalWinnings = 0
		winDetails = nil

		// Capture values from context before starting goroutine
		origin := c.Get("Origin")
		log.Printf("🫠🫠Origin received: %v", origin)

		// Send loss update to Mosomi's endpoint with environment-aware URL
		go func() {
			mosomiEndpoint := "https://admin-api.ibibe.africa/api/v1/update_loss"
			if len(origin) > 0 && (strings.Contains(strings.ToLower(origin), "test") || origin == "") {
				mosomiEndpoint = "https://admin-api3.ibibe.africa/api/v1/update_loss"
			}

			log.Printf("Mosomi endpoint: %v", mosomiEndpoint)

			lossPayload := map[string]interface{}{
				"bet_id":     req.BetID,
				"bet_status": "lost",
				"client_id":  req.ClientID,
			}
			jsonData, err := json.Marshal(lossPayload)
			if err != nil {
				log.Printf("Error marshaling loss payload: %v", err)
				return
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Post(mosomiEndpoint, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Error sending loss update to Mosomi's endpoint: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Mosomi's endpoint returned non-200 status: %d", resp.StatusCode)
			} else {
				log.Printf("Successfully sent loss update to Mosomi's endpoint")
			}
		}()

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
			freeSpinMultiplier = 1 // Default to 1 if no gold flowers
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
	isFreeSpin := req.IsFreeSpin
	if req.IsFreeSpin {
		currentFreeSpinIndex++
		remainingFreeSpins--

		// Prepare multiplier for NEXT spin (only if there are more spins)
		if remainingFreeSpins > 0 {
			freeSpinMultiplier = req.FreeSpinMultiplier + 1
		}

		log.Printf("Free Spin in progress: spin=%d, remaining=%d, multiplier=%d",
			currentFreeSpinIndex, remainingFreeSpins, freeSpinMultiplier)
	}

	// Check if Free Spin Bonus has ended
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
