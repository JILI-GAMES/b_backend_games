package birdspartydeluxe

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SpinHandler handles the /spin/birdspartydeluxe endpoint
// DELUXE: Modified to support connection-based clover mechanics and separate handling of special symbols
func (rg *RouteGroup) SpinHandler(c *fiber.Ctx) error {
	rngClient, settingsClient := rg.getClientsForRequest(c)

	var req SpinRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Initialize game state if needed
	if req.GameState.CurrentLevel == 0 {
		req.GameState = InitializeGameState()
	}
	if req.GameState.GameMode == "" {
		req.GameState.GameMode = "base"
	}

	// Ensure grid size matches current level
	expectedGridSize := req.GameState.CurrentLevel.GetGridSize()
	if req.GameState.GridSize != expectedGridSize {
		req.GameState.GridSize = expectedGridSize
		log.Printf("Corrected grid size to %d for level %d", expectedGridSize, req.GameState.CurrentLevel)
	}

	// Create rand instance
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Set bet multiplier
	req.GameState.Bet.Multiplier = BetAmountToMultiplier[req.GameState.Bet.Amount]

	// DELUXE: Reset booming reels multiplier for new spin (cascade sequence resets)
	ResetBoomingReels(&req.GameState)

	// DELUXE: Generate grid with potential connection-forming symbol connections (birds + clovers)
	req.GameState.Grid = GenerateGridWithWin(req.GameState.CurrentLevel, r, req.GameState.GameMode)

	// Find stage-cleared symbols (do NOT remove them yet)
	stageClearedSymbols := FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
	req.GameState.StageClearedSymbols = stageClearedSymbols

	// DELUXE: Find all connections and separate clover vs bird connections
	allConnections := FindAllConnections(req.GameState.Grid, req.GameState.CurrentLevel)
	cloverConnections, birdConnections := SeparateConnections(allConnections)

	// DELUXE: Process clover connections first - they upgrade multiplier but pay base value only
	totalWinnings := 0.0
	for i, cloverConnection := range cloverConnections {
		// Upgrade multiplier first
		UpgradeBoomingReels(&req.GameState)
		log.Printf("Clover connection found (%d symbols), multiplier upgraded to %.1fx",
			cloverConnection.Count, req.GameState.FreeSpins.CurrentMultiplier)

		// Calculate clover payout - BASE VALUE ONLY, NO MULTIPLIER
		payout := calculatePayout(cloverConnection.Symbol, cloverConnection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		// Clovers pay base value only - no booming reels multiplier applied
		cloverConnections[i].Payout = payout
		totalWinnings += payout
	}

	// Calculate total winnings from bird connections using current multiplier
	for i, connection := range birdConnections {
		payout := calculatePayout(connection.Symbol, connection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		// Apply booming reels multiplier to bird connections
		payout *= req.GameState.FreeSpins.CurrentMultiplier
		birdConnections[i].Payout = payout
		totalWinnings += payout
	}
	totalWinnings = round(totalWinnings)

	// Get RTP and call RNG for ALL paying connections (birds + clovers)
	allPayingConnections := append(cloverConnections, birdConnections...)
	if len(allPayingConnections) > 0 {
		rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
		if err != nil {
			log.Printf("Failed to get RTP: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve game settings",
			})
		}

		// Call RNG
		payoutMultiplier := totalWinnings / req.GameState.Bet.Amount
		ip := c.IP()
		userAgent := c.Get("User-Agent")

		log.Printf("✅IP: %v", ip)
		log.Printf("✅User-Agent: %v", userAgent)
		rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.GameState.Bet.Amount, ip, userAgent)
		if err != nil {
			log.Printf("Failed to call RNG API: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to determine outcome",
			})
		}

		// Adjust outcome based on RNG
		if rngResp.PrefOutcome == "loss" {
			log.Printf("RNG determined a loss outcome")
			req.GameState.Grid = GenerateLossGrid(req.GameState.CurrentLevel, r, req.GameState.GameMode)

			// Re-find stage-cleared symbols in loss grid
			stageClearedSymbols = FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
			req.GameState.StageClearedSymbols = stageClearedSymbols

			// Reset connections and winnings
			allConnections = nil
			cloverConnections = nil
			birdConnections = nil
			totalWinnings = 0

			// Reset booming reels since we regenerated the grid
			ResetBoomingReels(&req.GameState)
		}
	}

	// Reset cascade count for new spin
	req.GameState.CascadeCount = 0
	req.GameState.TotalWin = totalWinnings
	req.GameState.LastConnections = allConnections // Store all connections for cascade processing
	req.GameState.Cascading = len(allConnections) > 0

	// DELUXE: Check for free game symbols (rainbow eggs) - separate from clovers
	freeGameCount := CountFreeGameSymbols(req.GameState.Grid)
	if req.GameState.GameMode == "base" && freeGameCount > 0 {
		// Trigger free spins with rainbow egg
		req.GameState.GameMode = "freeSpins"
		req.GameState.FreeSpins.Remaining = FreeSpinsAwarded
		req.GameState.FreeSpins.TotalAwarded = FreeSpinsAwarded
		log.Printf("Free Spins triggered by rainbow egg, current booming reels multiplier: %.1fx", req.GameState.FreeSpins.CurrentMultiplier)
	}

	// Update free spins count (DELUXE: no re-triggering during free spins)
	if req.GameState.GameMode == "freeSpins" {
		req.GameState.FreeSpins.Remaining--
		if req.GameState.FreeSpins.Remaining <= 0 {
			req.GameState.GameMode = "base"
			// Reset free spins data but keep booming reels for this cascade sequence
			req.GameState.FreeSpins.Remaining = 0
			req.GameState.FreeSpins.TotalAwarded = 0
			log.Printf("Free Spins ended, booming reels multiplier continues: %.1fx", req.GameState.FreeSpins.CurrentMultiplier)
		}
	}

	// Calculate total cost
	var totalCost float64
	if req.GameState.GameMode == "freeSpins" {
		totalCost = 0
	} else {
		totalCost = req.GameState.Bet.Amount
	}

	// Determine if we have stage-cleared symbols
	hasStageCleared := len(stageClearedSymbols) > 0

	log.Printf("Spin completed: level=%d, gridSize=%dx%d, stageClearedSymbols=%d, hasStageCleared=%v, cascading=%v, boomingReels=%.1fx, cloverConnections=%d, birdConnections=%d",
		req.GameState.CurrentLevel, req.GameState.GridSize, req.GameState.GridSize,
		len(stageClearedSymbols), hasStageCleared, req.GameState.Cascading, req.GameState.FreeSpins.CurrentMultiplier, len(cloverConnections), len(birdConnections))

	return c.JSON(SpinResponse{
		Status:              "success",
		Message:             "",
		GameState:           req.GameState,
		StageClearedSymbols: stageClearedSymbols,
		HasStageCleared:     hasStageCleared,
		TotalCost:           totalCost,
	})
}

// ProcessStageClearedHandler handles the /process-stage-cleared/birdspartydeluxe endpoint
// DELUXE: Modified to support booming reels continuity and connection-based clover mechanics
func (rg *RouteGroup) ProcessStageClearedHandler(c *fiber.Ctx) error {
	rngClient, settingsClient := rg.getClientsForRequest(c)

	var req ProcessStageClearedRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate grid dimensions
	if !ValidateGridDimensions(req.GameState.Grid, req.GameState.CurrentLevel) {
		log.Printf("Invalid grid dimensions for level %d", req.GameState.CurrentLevel)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid grid dimensions",
		})
	}

	// Create rand instance
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get stage-cleared symbols from the current grid
	stageClearedSymbols := req.GameState.StageClearedSymbols
	if len(stageClearedSymbols) == 0 {
		// If none provided, find them from the grid
		stageClearedSymbols = FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
	}

	stageClearedCount := len(stageClearedSymbols)

	// PRESERVE ORIGINAL GRID before processing
	originalGrid := make([][]string, len(req.GameState.Grid))
	for i := range req.GameState.Grid {
		originalGrid[i] = make([]string, len(req.GameState.Grid[i]))
		copy(originalGrid[i], req.GameState.Grid[i])
	}

	// Prepare variables for level advancement
	var (
		levelAdvanced bool
		oldLevel      = req.GameState.CurrentLevel
		newLevel      = req.GameState.CurrentLevel
	)

	// Process stage-cleared symbols (remove, apply gravity, check level advancement)
	var newPositions []Position
	{
		// Remove stage-cleared symbols from grid surgically
		RemoveStageClearedSymbolsSurgical(req.GameState.Grid, stageClearedSymbols)
		// Apply gravity surgically and get new positions
		newPositions = ApplyGravitySurgical(req.GameState.Grid, stageClearedSymbols, req.GameState.CurrentLevel, r, req.GameState.GameMode)
		// Update stage progress
		req.GameState.StageProgress += len(stageClearedSymbols)
		log.Printf("Added %d stage-cleared symbols to progress, total: %d/15", len(stageClearedSymbols), req.GameState.StageProgress)
		// Check for level advancement
		if req.GameState.StageProgress >= StageProgressTarget {
			newLevel = AdvanceLevel(oldLevel)
			excessProgress := req.GameState.StageProgress - StageProgressTarget
			UpdateGameStateForLevel(&req.GameState, newLevel)

			if oldLevel == Level3 {
				req.GameState.StageProgress = 0 // Reset progress when looping from level 3 to 1
			} else {
				req.GameState.StageProgress = excessProgress // Carry over excess progress otherwise
			}

			// Generate new grid for the new level
			req.GameState.Grid = GenerateGrid(newLevel, r, req.GameState.GameMode)
			levelAdvanced = true
			log.Printf("Level advanced from %d to %d, excess progress: %d", oldLevel, newLevel, excessProgress)

			// --- NEW: Analyze the brand new grid for wins and special symbols ---
			allConnections := FindAllConnections(req.GameState.Grid, req.GameState.CurrentLevel)
			cloverConnections, birdConnections := SeparateConnections(allConnections)
			stageClearedSymbolsAfterLevelUp := FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
			req.GameState.StageClearedSymbols = stageClearedSymbolsAfterLevelUp

			// DELUXE: Process clover connections first - upgrade multiplier but pay base value only
			totalWinnings := 0.0
			for i, cloverConnection := range cloverConnections {
				// Upgrade multiplier first
				UpgradeBoomingReels(&req.GameState)
				log.Printf("Clover connection found on new level (%d symbols), multiplier upgraded to %.1fx",
					cloverConnection.Count, req.GameState.FreeSpins.CurrentMultiplier)

				// Calculate clover payout - BASE VALUE ONLY, NO MULTIPLIER
				payout := calculatePayout(cloverConnection.Symbol, cloverConnection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
				// Clovers pay base value only - no booming reels multiplier applied
				cloverConnections[i].Payout = payout
				totalWinnings += payout
			}

			// Calculate winnings from bird connections using current multiplier
			for i, connection := range birdConnections {
				payout := calculatePayout(connection.Symbol, connection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
				payout *= req.GameState.FreeSpins.CurrentMultiplier
				birdConnections[i].Payout = payout
				totalWinnings += payout
			}

			// DELUXE: Check for and trigger free spins on the new grid (rainbow eggs)
			freeGameCount := CountFreeGameSymbols(req.GameState.Grid)
			if req.GameState.GameMode == "base" && freeGameCount > 0 {
				req.GameState.GameMode = "freeSpins"
				req.GameState.FreeSpins.Remaining = FreeSpinsAwarded
				req.GameState.FreeSpins.TotalAwarded = FreeSpinsAwarded
				log.Printf("Free Spins triggered on new level by rainbow egg, current booming reels multiplier: %.1fx", req.GameState.FreeSpins.CurrentMultiplier)
			}

			// Update game state for the response
			req.GameState.TotalWin = round(totalWinnings)
			req.GameState.LastConnections = allConnections
			req.GameState.Cascading = len(allConnections) > 0
			req.GameState.CascadeCount = 0 // Reset for new level

			return c.JSON(ProcessStageClearedResponse{
				Status:            "success",
				Message:           "",
				GameState:         req.GameState,
				StageClearedCount: stageClearedCount,
				LevelAdvanced:     levelAdvanced,
				OldLevel:          oldLevel,
				NewLevel:          newLevel,
				Connections:       allConnections, // New connections from the new grid
				TotalCost:         0,
			})
		}
	}
	// Clear the stage-cleared symbols from game state since they've been processed
	req.GameState.StageClearedSymbols = []StageClearedSymbol{}

	// NOW check for connection-forming symbol connections in the new grid after gravity
	allConnections := FindAllConnections(req.GameState.Grid, req.GameState.CurrentLevel)
	cloverConnections, birdConnections := SeparateConnections(allConnections)

	// DELUXE: Process clover connections first - upgrade multiplier but pay base value only
	totalWinnings := 0.0
	for i, cloverConnection := range cloverConnections {
		// Upgrade multiplier first
		UpgradeBoomingReels(&req.GameState)
		log.Printf("Clover connection found after stage-cleared processing (%d symbols), multiplier upgraded to %.1fx",
			cloverConnection.Count, req.GameState.FreeSpins.CurrentMultiplier)

		// Calculate clover payout - BASE VALUE ONLY, NO MULTIPLIER
		payout := calculatePayout(cloverConnection.Symbol, cloverConnection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		// Clovers pay base value only - no booming reels multiplier applied
		cloverConnections[i].Payout = payout
		totalWinnings += payout
	}

	// Calculate total winnings from bird connections using current multiplier
	for i, connection := range birdConnections {
		payout := calculatePayout(connection.Symbol, connection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		// Apply booming reels multiplier to bird connections
		payout *= req.GameState.FreeSpins.CurrentMultiplier
		birdConnections[i].Payout = payout
		totalWinnings += payout
	}
	totalWinnings = round(totalWinnings)

	// Handle RNG for ALL paying connections (birds + clovers) with surgical loss approach
	rngBypassed := false
	allPayingConnections := append(cloverConnections, birdConnections...)
	if len(allPayingConnections) > 0 {
		rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
		if err != nil {
			log.Printf("Failed to get RTP: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve game settings",
			})
		}

		// Call RNG
		payoutMultiplier := totalWinnings / req.GameState.Bet.Amount
		ip := c.IP()
		userAgent := c.Get("User-Agent")

		log.Printf("✅IP: %v", ip)
		log.Printf("✅User-Agent: %v", userAgent)
		rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.GameState.Bet.Amount, ip, userAgent)
		if err != nil {
			log.Printf("Failed to call RNG API: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to determine outcome",
			})
		}

		// SURGICAL LOSS: Adjust outcome based on RNG while preserving grid structure
		if rngResp.PrefOutcome == "loss" {
			log.Printf("RNG determined a loss outcome for stage-cleared processing")
			// Try surgical loss approach first (only new positions)
			success := ApplySurgicalLoss(&req.GameState, originalGrid, stageClearedSymbols, req.GameState.CurrentLevel, r, newPositions)
			if !success {
				// If surgical loss is impossible, bypass RNG and allow the win
				log.Printf("⚠️  RNG BYPASS: Surgical loss impossible after stage-cleared processing - preserving natural outcome")
				log.Printf("⚠️  GRID PRESERVATION: Maintaining grid structure as surgical loss would break game mechanics")
				log.Printf("⚠️  REASON: Stage-cleared symbol removal at positions %+v made loss impossible", stageClearedSymbols)
				rngBypassed = true
				// Keep the original connections and winnings
				// Grid remains as it is after stage-cleared processing
			} else {
				// Surgical loss successful - remove ALL paying connections but keep multiplier upgrades
				cloverConnections = nil
				birdConnections = nil
				allConnections = nil // IMPORTANT: Clear allConnections when surgical loss is successful
				totalWinnings = 0
				log.Printf("Surgical loss applied successfully after stage-cleared processing (all paying connections)")

				// IMPORTANT: Re-find connections in the modified grid to ensure consistency
				allConnections = FindAllConnections(req.GameState.Grid, req.GameState.CurrentLevel)
			}
		}
	}

	// Update game state with connection results
	req.GameState.TotalWin = totalWinnings
	req.GameState.LastConnections = allConnections // Store all connections for cascade processing
	req.GameState.Cascading = len(allConnections) > 0

	// Reset cascade count since this is after stage-cleared processing
	// if there is a win then cascade count to be 1 else 0
	if len(allConnections) > 0 {
		req.GameState.CascadeCount = 1
	} else {
		req.GameState.CascadeCount = 0
	}

	logMessage := fmt.Sprintf("ProcessStageCleared completed: stageClearedCount=%d, levelAdvanced=%v, oldLevel=%d, newLevel=%d, progress=%d, cascading=%v, boomingReels=%.1fx, cloverConnections=%d, birdConnections=%d",
		stageClearedCount, levelAdvanced, oldLevel, newLevel, req.GameState.StageProgress, req.GameState.Cascading, req.GameState.FreeSpins.CurrentMultiplier, len(cloverConnections), len(birdConnections))

	if rngBypassed {
		logMessage += " [RNG BYPASSED - Surgical loss impossible]"
	}

	log.Printf(logMessage)

	return c.JSON(ProcessStageClearedResponse{
		Status:            "success",
		Message:           "",
		GameState:         req.GameState,
		StageClearedCount: stageClearedCount,
		LevelAdvanced:     levelAdvanced,
		OldLevel:          oldLevel,
		NewLevel:          newLevel,
		Connections:       allConnections,
		TotalCost:         0,
	})
}

// CascadeHandler handles the /cascade/birdspartydeluxe endpoint
// DELUXE: Modified to support connection-based clover mechanics and booming reels progression
func (rg *RouteGroup) CascadeHandler(c *fiber.Ctx) error {
	rngClient, settingsClient := rg.getClientsForRequest(c)

	var req CascadeRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validate request
	if err := validateRequest(req.ClientID, req.GameID, req.PlayerID, req.BetID, req.GameState.Bet.Amount); err != nil {
		log.Printf("Request validation failed: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Validate grid dimensions
	if !ValidateGridDimensions(req.GameState.Grid, req.GameState.CurrentLevel) {
		log.Printf("Invalid grid dimensions for level %d", req.GameState.CurrentLevel)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid grid dimensions",
		})
	}

	// Increment cascade count
	req.GameState.CascadeCount++

	// Create rand instance
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// PRESERVE ORIGINAL GRID before processing for surgical loss capability
	originalGrid := make([][]string, len(req.GameState.Grid))
	for i := range req.GameState.Grid {
		originalGrid[i] = make([]string, len(req.GameState.Grid[i]))
		copy(originalGrid[i], req.GameState.Grid[i])
	}

	var allConnections []Connection
	var cloverConnections []Connection
	var birdConnections []Connection
	var totalWinnings float64
	var affectedPositions []Position
	var newPositions []Position

	// Process cascade surgically
	if req.GameState.CascadeCount >= 1 && len(req.GameState.LastConnections) > 0 {
		// SURGICAL: Remove previous connections and apply gravity surgically
		affectedPositions = RemoveConnectionsSurgical(req.GameState.Grid, req.GameState.LastConnections)
		newPositions = ApplyGravitySurgicalForCascade(req.GameState.Grid, affectedPositions, req.GameState.CurrentLevel, r, req.GameState.GameMode)
	} else {
		// First cascade call - find existing connections
		allConnections = FindAllConnections(req.GameState.Grid, req.GameState.CurrentLevel)
		if len(allConnections) > 0 {
			// Extract positions that will be affected for surgical processing
			for _, connection := range allConnections {
				affectedPositions = append(affectedPositions, connection.Positions...)
			}
			newPositions = ApplyGravitySurgicalForCascade(req.GameState.Grid, affectedPositions, req.GameState.CurrentLevel, r, req.GameState.GameMode)
		}
	}

	// Find connection-forming symbol connections after cascade processing
	if req.GameState.CascadeCount >= 1 || len(allConnections) == 0 {
		allConnections = FindAllConnections(req.GameState.Grid, req.GameState.CurrentLevel)
	}

	// DELUXE: Separate clover and bird connections
	cloverConnections, birdConnections = SeparateConnections(allConnections)

	// DELUXE: Process clover connections first - upgrade multiplier but pay base value only
	totalWinnings = 0.0
	for i, cloverConnection := range cloverConnections {
		// Upgrade multiplier first
		UpgradeBoomingReels(&req.GameState)
		log.Printf("Clover connection found in cascade (%d symbols), multiplier upgraded to %.1fx",
			cloverConnection.Count, req.GameState.FreeSpins.CurrentMultiplier)

		// Calculate clover payout - BASE VALUE ONLY, NO MULTIPLIER
		payout := calculatePayout(cloverConnection.Symbol, cloverConnection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		// Clovers pay base value only - no booming reels multiplier applied
		cloverConnections[i].Payout = payout
		totalWinnings += payout
	}

	// Calculate total winnings from bird connections using current multiplier
	for i, connection := range birdConnections {
		payout := calculatePayout(connection.Symbol, connection.Count, req.GameState.CurrentLevel, req.GameState.Bet.Multiplier)
		// Apply booming reels multiplier to bird connections
		payout *= req.GameState.FreeSpins.CurrentMultiplier
		birdConnections[i].Payout = payout
		totalWinnings += payout
	}
	totalWinnings = round(totalWinnings)

	// Handle RNG for ALL paying connections (birds + clovers) with surgical loss approach
	rngBypassed := false
	allPayingConnections := append(cloverConnections, birdConnections...)
	if len(allPayingConnections) > 0 {
		rtp, err := settingsClient.GetRTP(req.ClientID, req.GameID, req.PlayerID)
		if err != nil {
			log.Printf("Failed to get RTP: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to retrieve game settings",
			})
		}

		// Call RNG
		payoutMultiplier := totalWinnings / req.GameState.Bet.Amount
		ip := c.IP()
		userAgent := c.Get("User-Agent")

		log.Printf("✅IP: %v", ip)
		log.Printf("✅User-Agent: %v", userAgent)
		rngResp, err := rngClient.GetOutcome(req.ClientID, req.GameID, req.PlayerID, req.BetID, rtp, payoutMultiplier, req.GameState.Bet.Amount, ip, userAgent)
		if err != nil {
			log.Printf("Failed to call RNG API: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to determine outcome",
			})
		}

		// SURGICAL LOSS: Adjust outcome based on RNG while preserving grid structure
		if rngResp.PrefOutcome == "loss" {
			log.Printf("RNG determined a loss outcome for cascade")

			// Try surgical loss approach first (only new positions)
			success := ApplySurgicalLossForCascade(&req.GameState, originalGrid, newPositions, req.GameState.CurrentLevel, r)

			if !success {
				// If surgical loss is impossible, bypass RNG and allow the win
				log.Printf("⚠️  RNG BYPASS: Surgical loss impossible after cascade processing - preserving natural outcome")
				log.Printf("⚠️  GRID PRESERVATION: Maintaining grid structure as surgical loss would break game mechanics")
				log.Printf("⚠️  REASON: Cascade processing at %d positions made loss impossible", len(affectedPositions))
				rngBypassed = true

				// Keep the original connections and winnings
				// Grid remains as it is after cascade processing
			} else {
				// Surgical loss successful - remove ALL paying connections but keep multiplier upgrades
				cloverConnections = nil
				birdConnections = nil
				allConnections = nil // IMPORTANT: Clear allConnections when surgical loss is successful
				totalWinnings = 0
				log.Printf("Surgical loss applied successfully after cascade processing (all paying connections)")

				// IMPORTANT: Re-find connections in the modified grid to ensure consistency
				allConnections = FindAllConnections(req.GameState.Grid, req.GameState.CurrentLevel)
			}
		}
	}

	// IMPORTANT: After all processing, check for stage-cleared symbols that may have appeared
	stageClearedSymbols := FindStageClearedSymbols(req.GameState.Grid, req.GameState.CurrentLevel)
	hasStageCleared := len(stageClearedSymbols) > 0

	// Store stage-cleared symbols in game state for potential next call to process-stage-cleared
	req.GameState.StageClearedSymbols = stageClearedSymbols

	// DELUXE: Check for free game symbols (rainbow eggs) only if no connections and no RNG bypass
	if len(allConnections) == 0 && !rngBypassed {
		freeGameCount := CountFreeGameSymbols(req.GameState.Grid)
		if req.GameState.GameMode == "base" && freeGameCount > 0 {
			req.GameState.GameMode = "freeSpins"
			req.GameState.FreeSpins.Remaining = FreeSpinsAwarded
			req.GameState.FreeSpins.TotalAwarded = FreeSpinsAwarded
			log.Printf("Free Spins triggered during cascade by rainbow egg, current booming reels multiplier: %.1fx", req.GameState.FreeSpins.CurrentMultiplier)
		}
	}

	// DELUXE: If no more connections, reset booming reels (cascade sequence ends)
	if len(allConnections) == 0 {
		log.Printf("Cascade sequence ended, resetting booming reels from %.1fx to 1.0x", req.GameState.FreeSpins.CurrentMultiplier)
		ResetBoomingReels(&req.GameState)
	}

	// Update game state
	req.GameState.TotalWin = totalWinnings
	req.GameState.LastConnections = allConnections
	req.GameState.Cascading = len(allConnections) > 0

	logMessage := fmt.Sprintf("Cascade completed: level=%d, gridSize=%dx%d, totalWin=%.2f, cascading=%v, cascadeCount=%d, stageClearedDetected=%v, boomingReels=%.1fx, cloverConnections=%d, birdConnections=%d",
		req.GameState.CurrentLevel, req.GameState.GridSize, req.GameState.GridSize,
		totalWinnings, req.GameState.Cascading, req.GameState.CascadeCount, hasStageCleared, req.GameState.FreeSpins.CurrentMultiplier, len(cloverConnections), len(birdConnections))

	if rngBypassed {
		logMessage += " [RNG BYPASSED - Surgical loss impossible]"
	}

	log.Printf(logMessage)

	return c.JSON(CascadeResponse{
		Status:              "success",
		Message:             "",
		GameState:           req.GameState,
		Connections:         allConnections,
		StageClearedSymbols: stageClearedSymbols, // Include detected stage-cleared symbols
		HasStageCleared:     hasStageCleared,     // Flag to indicate stage-cleared symbols found
		TotalCost:           0,
	})
}

// validateRequest validates the request fields
func validateRequest(clientID, gameID, playerID, betID string, betAmount float64) error {
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
		return fmt.Errorf("invalid bet amount, allowed values are 0.1, 0.2, 0.3, 0.5, 1.0")
	}
	return nil
}

// isValidBetAmount checks if the bet amount is valid
func isValidBetAmount(amount float64) bool {
	validAmounts := []float64{0.1, 0.2, 0.3, 0.5, 1.0}
	for _, valid := range validAmounts {
		if amount == valid {
			return true
		}
	}
	return false
}
