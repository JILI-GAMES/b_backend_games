package onepiece

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/JILI-GAMES/b_backend_games/pkg/common/rng"
	"github.com/JILI-GAMES/b_backend_games/pkg/common/security"
	"github.com/gofiber/fiber/v2"
)

var featureBuy bool = false

func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
	// Check if request body is empty
	body := c.Body()
	if len(body) == 0 {
		log.Printf("Empty request body received")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Request body is required and must be encrypted",
		})
	}

	// Decrypt request using Unity's CBC approach
	decryptedBody, err := security.DecryptRequestCBC(body)
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "Request must be encrypted",
		})
	}

	var req SpinRequest
	if err := json.Unmarshal(decryptedBody, &req); err != nil {
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

	betAmount := req.BetAmount
	reels := GenerateReelsWithWin()
	effectiveMultiplier := req.FreeSpinMultiplier

	if req.IsFreeSpin {
		featureBuy = true
		log.Printf("Feature Buy✅✅✅: %v (Free Spin Active)", featureBuy)

		if req.ExtraBonusMultiplier > 0 {
			effectiveMultiplier *= (req.ExtraBonusMultiplier)
			log.Printf("Applied extra bonus multiplier: base=%d, extra=%d, effective=%d",
				req.FreeSpinMultiplier, req.ExtraBonusMultiplier, effectiveMultiplier)
		}
	} else {
		featureBuy = false
		log.Printf("Feature Buy❌❌❌: %v (Not Free Spin)", featureBuy)
	}

	totalWinnings, winDetails := CalculateWins(reels, betAmount, effectiveMultiplier, req.IsFreeSpin)
	log.Printf("Initial calculation: totalWinnings=%v, winDetails=%v", totalWinnings, winDetails)

	rngClient, settingsClient := rg.getClientsForRequest(c)

	payoutMultiplier := totalWinnings / req.BetAmount
	if math.IsNaN(payoutMultiplier) || math.IsInf(payoutMultiplier, 0) {
		payoutMultiplier = 0
		log.Printf("Payout multiplier is NaN or Inf, setting to 0")
	}
	log.Printf("Payout multiplier: %v", payoutMultiplier)

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
				"source":       "one-piece",
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
🚨 <b>Settings API Failure Alert</b> 🚨

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

	ip := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("✅IP: %v", ip)
	log.Printf("✅User-Agent: %v", userAgent)

	var rngResp rng.Response
	var rngErr error
	if settingsErr == nil {
		rngResp, rngErr = rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.BetAmount, ip, userAgent, featureBuy)
		if rngErr != nil {
			log.Printf("Failed to call RNG API: %v", rngErr)
		}
	} else {
		rngResp = rng.Response{PrefOutcome: "loss"}
		rngErr = nil
		log.Printf("‼️‼️‼️Forcing loss outcome due to settings API failure")
	}

	if rngErr != nil {
		log.Printf("RNG API failed - sending Telegram notification")

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
				"source":       "one-piece",
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
🚨 <b>RNG API Failure Alert</b> 🚨

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
	}

	scatterCount := CountScatters(reels)

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

	remainingFreeSpins := req.RemainingFreeSpins
	totalFreeSpinsAwarded := req.TotalFreeSpinsAwarded
	if freeSpinTriggered {
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

	currentFreeSpinIndex := req.CurrentFreeSpinIndex
	freeSpinMultiplier := req.FreeSpinMultiplier
	if req.IsFreeSpin {
		currentFreeSpinIndex++
		remainingFreeSpins--

		if req.FreeSpinOption > 0 {
			option := FreeSpinOptions[req.FreeSpinOption]

			if req.FreeSpinOption == 5 {
				freeSpinMultiplier = option.InitialMultiplier
			} else {
				freeSpinMultiplier = option.InitialMultiplier + (currentFreeSpinIndex)*option.MultiplierIncrease
			}
		}
	}
	log.Printf("Free Spin state updated: remainingFreeSpins=%d, totalFreeSpinsAwarded=%d, currentFreeSpinIndex=%d, freeSpinMultiplier=%d", remainingFreeSpins, totalFreeSpinsAwarded, currentFreeSpinIndex, freeSpinMultiplier)

	isFreeSpin := req.IsFreeSpin
	if req.IsFreeSpin && remainingFreeSpins <= 0 {
		isFreeSpin = false
		currentFreeSpinIndex = 0
		freeSpinMultiplier = 0
		remainingFreeSpins = 0
		totalFreeSpinsAwarded = 0
		log.Printf("Free Spin Bonus ended")
	}

	// if its a win we update the win in mosomis endpoint https://admin-api.ibibe.africa/api/v1/update_bet
	var walletBalance float64 = 0
	if totalWinnings > 0 {
		origin := c.Get("Origin")
		log.Printf("🫠🫠Origin received: %v", origin)
		mosomiEndpoint := "https://admin-api.ibibe.africa/api/v1/update_bet"
		if len(origin) > 0 && (strings.Contains(strings.ToLower(origin), "test") || origin == "") {
			mosomiEndpoint = "https://admin-api3.ibibe.africa/api/v1/update_bet"
		}

		winPayload := map[string]interface{}{
			"bet_id":     req.BetID,
			"amount_won": totalWinnings,
			"client_id":  req.ClientID,
		}
		jsonData, err := json.Marshal(winPayload)
		if err != nil {
			log.Printf("Error marshaling win payload: %v", err)
		} else {
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Post(mosomiEndpoint, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Error sending win update to Mosomi's endpoint: %v", err)
			} else {
				defer resp.Body.Close()

				// Read entire response body once for logging/decoding
				bodyBytes, readErr := io.ReadAll(resp.Body)
				if readErr != nil {
					log.Printf("Error reading bet update response body: %v", readErr)
				}

				contentType := resp.Header.Get("Content-Type")
				statusCode := resp.StatusCode

				if statusCode != http.StatusOK {
					preview := string(bodyBytes)
					if len(preview) > 300 {
						preview = preview[:300] + "..."
					}
					log.Printf("Mosomi bet update non-200 status: %d, Content-Type: %s, Body: %s", statusCode, contentType, preview)
				}

				if !strings.Contains(strings.ToLower(contentType), "application/json") {
					preview := string(bodyBytes)
					if len(preview) > 300 {
						preview = preview[:300] + "..."
					}
					log.Printf("Mosomi bet update returned non-JSON response. Content-Type: %s, Body: %s", contentType, preview)
				}

				// Parse the response to get the new wallet balance
				var betUpdateResponse struct {
					StatusCode       int     `json:"status_code"`
					Message          string  `json:"message"`
					BetID            string  `json:"bet_id"`
					AmountWon        float64 `json:"amount_won"`
					NewWalletBalance float64 `json:"new_wallet_balance"`
					Status           string  `json:"status"`
				}

				if err := json.Unmarshal(bodyBytes, &betUpdateResponse); err != nil {
					preview := string(bodyBytes)
					if len(preview) > 300 {
						preview = preview[:300] + "..."
					}
					log.Printf("Error decoding bet update response: %v. Body: %s", err, preview)
				} else {
					walletBalance = betUpdateResponse.NewWalletBalance
					log.Printf("Bet update successful. New wallet balance: %v", walletBalance)
				}
			}
		}
	}

	log.Printf("Spin completed: totalWin=%v, freeSpinTriggered=%v, freeSpinRetriggered=%v, isFreeSpin=%v", totalWinnings, freeSpinTriggered, freeSpinRetriggered, isFreeSpin)

	// Build response
	response := SpinResponse{
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
		WalletBalance:         walletBalance,
	}

	// Encrypt response using Unity's CBC approach
	encryptedResponse, err := security.EncryptResponseCBC(response)
	if err != nil {
		log.Printf("Encryption failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encrypt response",
		})
	}

	return c.Send(encryptedResponse)
}

func (rg *RouteGroup) SelectFreeSpinOptionHandler(c *fiber.Ctx) error {
	// Check if request body is empty
	body := c.Body()
	if len(body) == 0 {
		log.Printf("Empty request body received")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Request body is required",
		})
	}

	// Decrypt request using Unity's CBC approach
	decryptedBody, err := security.DecryptRequestCBC(body)
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "Request must be encrypted.",
		})
	}

	var req SelectFreeSpinOptionRequest
	if err := json.Unmarshal(decryptedBody, &req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if err := validateSelectFreeSpinOptionRequest(req.ClientID, req.GameID, req.PlayerID, req.Option, req.ExtraBonusMultiplier); err != nil {
		log.Printf("Request validation failed: %v", err)
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

	option, exists := FreeSpinOptions[req.Option]
	if !exists {
		log.Printf("Invalid Free Spin Bonus option: %d", req.Option)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid Free Spin Bonus option",
		})
	}

	totalSpins := option.TotalSpins
	initialMultiplier := option.InitialMultiplier

	log.Printf("Free Spin Bonus option selected: playeID=%s, option=%d, totalSpins=%d, initialMultiplier=%d, extraBonusMultiplier=%d",
		req.PlayerID, req.Option, totalSpins, initialMultiplier, req.ExtraBonusMultiplier)

	response := SelectFreeSpinOptionResponse{
		TotalFreeSpins:       totalSpins,
		InitialMultiplier:    initialMultiplier,
		MaxFreeSpins:         option.MaxSpins,
		ExtraBonusMultiplier: req.ExtraBonusMultiplier,
	}

	// Encrypt response using Unity's CBC approach
	encryptedResponse, err := security.EncryptResponseCBC(response)
	if err != nil {
		log.Printf("Encryption failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encrypt response",
		})
	}

	return c.Send(encryptedResponse)
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
