package winningmask

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Constants
const (
	Denomination      = 0.01
	Reels             = 5
	Rows              = 4
	CreditMultiplier  = 50  // Minimum bet is 50 credits per bet multiplier
	FreeSpinCount     = 10  // Free spins awarded
	MaxTotalFreeSpins = 150 // Maximum total free spins possible
	MaxMaskMultiplier = 200 // Maximum mask reel bonus multiplier
)

// Symbol weights for random generation
var SymbolWeights = map[Symbol]float64{
	SymbolPurpleMask: 0.04,
	SymbolOrangeMask: 0.06,
	SymbolGreenMask:  0.06,
	SymbolYellowMask: 0.08,
	SymbolBlueMask:   0.08,
	SymbolA:          0.10,
	SymbolK:          0.10,
	SymbolQ:          0.12,
	SymbolJ:          0.12,
	Symbol10:         0.12,
	SymbolWild:       0.04,
	SymbolBonus:      0.04,
	SymbolMaskReel:   0.04,
}

// Paytable (payouts for Bet Multiplier = 1)
var Paytable = map[Symbol]map[int]int{
	SymbolPurpleMask: {3: 50, 4: 200, 5: 1000},
	SymbolOrangeMask: {3: 25, 4: 150, 5: 400},
	SymbolGreenMask:  {3: 25, 4: 150, 5: 400},
	SymbolYellowMask: {3: 20, 4: 75, 5: 200},
	SymbolBlueMask:   {3: 20, 4: 75, 5: 200},
	SymbolA:          {3: 10, 4: 50, 5: 150},
	SymbolK:          {3: 10, 4: 50, 5: 150},
	SymbolQ:          {3: 5, 4: 20, 5: 100},
	SymbolJ:          {3: 5, 4: 20, 5: 100},
	Symbol10:         {3: 5, 4: 20, 5: 100},
	SymbolBonus:      {3: 2}, // Free Spin Bonus symbol pays 2 credits times bet amount
}

// BetAmountMap maps bet amounts to multipliers
var BetAmountMap = map[float64]int{
	0.5:  1,
	1.0:  2,
	1.5:  3,
	2.5:  5,
	5.0:  10,
}

// WaysToWin generates all possible 1024 ways to win (4^5)
var WaysToWin = generateWaysToWin()

func generateWaysToWin() [][]int {
	ways := make([][]int, 0, 1024)
	for r0 := 0; r0 < 4; r0++ {
		for r1 := 0; r1 < 4; r1++ {
			for r2 := 0; r2 < 4; r2++ {
				for r3 := 0; r3 < 4; r3++ {
					for r4 := 0; r4 < 4; r4++ {
						ways = append(ways, []int{r0, r1, r2, r3, r4})
					}
				}
			}
		}
	}
	return ways
}

// WeightedRandomSymbol selects a symbol based on weights, with reel-specific restrictions
func WeightedRandomSymbol(r *rand.Rand, reelIndex int, bonusPlacedInReel bool, maskReelPlacedInReel bool) Symbol {
	totalWeight := 0.0
	for symbol, weight := range SymbolWeights {
		// Wild only appears on reels 2-5 (indices 1-4)
		if symbol == SymbolWild && reelIndex == 0 {
			continue
		}
		// Bonus symbols only appear on reels 1-3 (indices 0-2) and only once per reel
		if symbol == SymbolBonus && (reelIndex > 2 || bonusPlacedInReel) {
			continue
		}
		// Mask Reel symbols only appear on reels 3-5 (indices 2-4) and only once per reel
		if symbol == SymbolMaskReel && (reelIndex < 2 || maskReelPlacedInReel) {
			continue
		}
		totalWeight += weight
	}

	roll := r.Float64() * totalWeight
	currentWeight := 0.0
	for symbol, weight := range SymbolWeights {
		// Skip symbols that can't appear on this reel
		if symbol == SymbolWild && reelIndex == 0 {
			continue
		}
		if symbol == SymbolBonus && (reelIndex > 2 || bonusPlacedInReel) {
			continue
		}
		if symbol == SymbolMaskReel && (reelIndex < 2 || maskReelPlacedInReel) {
			continue
		}

		currentWeight += weight
		if roll <= currentWeight {
			return symbol
		}
	}
	return SymbolA // Fallback
}

// GenerateReelsWithWin generates a 5x4 grid with a guaranteed win
func GenerateReelsWithWin() [][]string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var reels [][]string

	// Keep generating until a win is found
	for {
		reels = make([][]string, Reels)
		for reel := 0; reel < Reels; reel++ {
			reels[reel] = make([]string, Rows)
			bonusPlaced := false
			maskReelPlaced := false
			for row := 0; row < Rows; row++ {
				symbol := WeightedRandomSymbol(r, reel, bonusPlaced, maskReelPlaced)
				if symbol == SymbolBonus {
					bonusPlaced = true
				}
				if symbol == SymbolMaskReel {
					maskReelPlaced = true
				}
				reels[reel][row] = string(symbol)
			}
		}

		// Check for a win
		totalWinnings, _, bonusPayout, _, _ := CalculateWins(reels, 1, false)
		if totalWinnings > 0 || bonusPayout > 0 {
			log.Printf("Generated reels with win: %v", reels)
			break
		}
	}

	return reels
}

// GenerateLossReels generates a 5x4 grid with no wins but allows triggering features
func GenerateLossReels() [][]string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var reels [][]string

	// Keep generating until no regular wins are found
	for {
		reels = make([][]string, Reels)
		for reel := 0; reel < Reels; reel++ {
			reels[reel] = make([]string, Rows)
			bonusPlaced := false
			maskReelPlaced := false
			for row := 0; row < Rows; row++ {
				symbol := WeightedRandomSymbol(r, reel, bonusPlaced, maskReelPlaced)
				if symbol == SymbolBonus {
					bonusPlaced = true
				}
				if symbol == SymbolMaskReel {
					maskReelPlaced = true
				}
				reels[reel][row] = string(symbol)
			}
		}

		// Check for no regular wins (but allow bonus payouts)
		totalWinnings, _, _, _, _ := CalculateWins(reels, 1, false)
		if totalWinnings == 0 {
			log.Printf("Generated reels with no wins: %v", reels)
			break
		}
	}

	return reels
}

// CalculateWins calculates the total payout and win details
func CalculateWins(reels [][]string, betMultiplier int, isFreeSpin bool) (float64, []WinDetail, float64, int, int) {
	totalPayout := 0.0
	var winDetails []WinDetail
	bonusPayout := 0.0

	// Map to store all unique winning paths
	uniqueWins := make(map[string]WinDetail)
	// Keep track of which symbol+start combinations we've seen with which counts
	seenCombinations := make(map[string]map[int]bool)
	var winsMu sync.Mutex

	// Worker pool configuration
	const maxWorkers = 50
	jobs := make(chan struct {
		wayIndex int
		way      []int
	}, len(WaysToWin))
	results := make(chan WinDetail, len(WaysToWin))
	var wg sync.WaitGroup

	// Start workers to identify potential wins
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				way := job.way

				// Preallocate slices
				symbols := make([]string, 0, 5)
				payline := make([]Position, 0, 5)

				for reel, row := range way {
					symbol := reels[reel][row]
					symbols = append(symbols, symbol)
					pos := Position{Reel: reel, Row: row}
					payline = append(payline, pos)
				}

				// Count matching symbols
				firstSymbol := symbols[0]
				if firstSymbol == string(SymbolBonus) || firstSymbol == string(SymbolMaskReel) {
					continue
				}

				matchCount := 1
				for i := 1; i < len(symbols); i++ {
					currentSymbol := symbols[i]
					if currentSymbol == firstSymbol || currentSymbol == string(SymbolWild) {
						matchCount++
					} else {
						break
					}
				}

				if matchCount >= 3 {
					// Check if this symbol/count has a payout in the paytable
					payoutValue, exists := Paytable[Symbol(firstSymbol)][matchCount]
					if !exists {
						continue // Skip if no payout for this combination
					}

					// Calculate payout based on payline and bet multiplier
					payout := float64(payoutValue) * float64(betMultiplier) * Denomination
					payout = math.Round(payout*100) / 100
					win := WinDetail{
						Symbol:    firstSymbol,
						Count:     matchCount,
						Payout:    payout,
						Positions: payline[:matchCount],
					}
					results <- win
				}
			}
		}()
	}

	// Send jobs to workers
	for wayIndex, way := range WaysToWin {
		jobs <- struct {
			wayIndex int
			way      []int
		}{wayIndex: wayIndex, way: way}
	}
	close(jobs)

	// Close the results channel after all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process all wins
	for win := range results {
		if len(win.Positions) < 3 {
			continue
		}

		// Create a key for the starting position and symbol
		startKey := fmt.Sprintf("%s|%d,%d",
			win.Symbol,
			win.Positions[0].Reel, win.Positions[0].Row)

		// Create a key for the complete path
		var pathKey strings.Builder
		pathKey.WriteString(win.Symbol)
		for _, pos := range win.Positions {
			pathKey.WriteString(fmt.Sprintf("|%d,%d", pos.Reel, pos.Row))
		}

		winsMu.Lock()

		// Initialize the count map if it doesn't exist
		if _, exists := seenCombinations[startKey]; !exists {
			seenCombinations[startKey] = make(map[int]bool)
		}

		// If we've seen a longer match for this start position, skip this one
		longerMatchExists := false
		for count := range seenCombinations[startKey] {
			if count > win.Count {
				longerMatchExists = true
				break
			}
		}

		if !longerMatchExists {
			// If we've seen a shorter match for this start position, remove it
			for count := range seenCombinations[startKey] {
				if count < win.Count {
					delete(seenCombinations[startKey], count)

					// Also remove any wins with this start position and shorter count
					for existingPathKey, existingWin := range uniqueWins {
						if existingWin.Symbol == win.Symbol &&
							existingWin.Positions[0].Reel == win.Positions[0].Reel &&
							existingWin.Positions[0].Row == win.Positions[0].Row &&
							existingWin.Count < win.Count {
							delete(uniqueWins, existingPathKey)
						}
					}
				}
			}

			// Add this count to the seen combinations
			seenCombinations[startKey][win.Count] = true

			// Add this win to the unique wins
			uniqueWins[pathKey.String()] = win
		}

		winsMu.Unlock()
	}

	// Collect all the unique wins
	winsMu.Lock()
	for _, win := range uniqueWins {
		log.Printf("Calculating payout: symbol=%s, matchCount=%d, betMultiplier=%d, payout=%v",
			win.Symbol, win.Count, betMultiplier, win.Payout)

		totalPayout += win.Payout
		totalPayout = math.Round(totalPayout*100) / 100
		winDetails = append(winDetails, win)
	}
	winsMu.Unlock()

	// Calculate Free Spin Bonus payout separately
	bonusPositions := GetBonusPositionsOnFirstThreeReels(reels)
	bonusCount := len(bonusPositions)
	if bonusCount >= 3 {
		// Bonus pay is the credit/odds value multiplied by the bet amount
		bonusPayValue := float64(Paytable[SymbolBonus][3])
		totalBetAmount := float64(betMultiplier*CreditMultiplier) * Denomination
		bonusPayout = bonusPayValue * totalBetAmount
		bonusPayout = math.Round(bonusPayout*100) / 100
		log.Printf("Bonus payout: count=%d, odds=%v, betMultiplier=%d, totalBetAmount=%v, payout=%v",
			bonusCount, bonusPayValue, betMultiplier, totalBetAmount, bonusPayout)
	}

	// Get mask reel positions
	maskReelPositions := GetMaskReelPositions(reels)

	return totalPayout, winDetails, bonusPayout, len(bonusPositions), len(maskReelPositions)
}

// GetSymbolPositions finds positions of a specific symbol on the reels
func GetSymbolPositions(reels [][]string, symbol string) []Position {
	var positions []Position

	for reel := 0; reel < len(reels); reel++ {
		for row := 0; row < len(reels[reel]); row++ {
			if reels[reel][row] == symbol {
				positions = append(positions, Position{
					Reel: reel,
					Row:  row,
				})
			}
		}
	}

	return positions
}

// GetBonusPositionsOnFirstThreeReels finds positions of Bonus symbols on reels 1-3 only
func GetBonusPositionsOnFirstThreeReels(reels [][]string) []Position {
	var positions []Position

	// Only check the first three reels (indices 0-2)
	for reel := 0; reel < 3 && reel < len(reels); reel++ {
		for row := 0; row < len(reels[reel]); row++ {
			if reels[reel][row] == string(SymbolBonus) {
				positions = append(positions, Position{
					Reel: reel,
					Row:  row,
				})
			}
		}
	}

	return positions
}

// GetMaskReelPositions finds positions of Mask Reel symbols on reels 3-5
func GetMaskReelPositions(reels [][]string) []Position {
	var positions []Position

	// Only check reels 3-5 (indices 2-4)
	for reel := 2; reel < 5 && reel < len(reels); reel++ {
		for row := 0; row < len(reels[reel]); row++ {
			if reels[reel][row] == string(SymbolMaskReel) {
				positions = append(positions, Position{
					Reel: reel,
					Row:  row,
				})
			}
		}
	}

	return positions
}

// HasBonusOnFirstThreeReels checks if there are at least 3 bonus symbols on the first three reels
func HasBonusOnFirstThreeReels(reels [][]string) bool {
	bonusCount := 0
	for reel := 0; reel < 3; reel++ {
		for row := 0; row < Rows; row++ {
			if reels[reel][row] == string(SymbolBonus) {
				bonusCount++
				break // Only count one per reel
			}
		}
	}
	return bonusCount >= 3
}

// HasMaskReelOnLastThreeReels checks if there are at least 3 mask reel symbols on reels 3-5
func HasMaskReelOnLastThreeReels(reels [][]string) bool {
	maskCount := 0
	for reel := 2; reel < 5; reel++ {
		for row := 0; row < Rows; row++ {
			if reels[reel][row] == string(SymbolMaskReel) {
				maskCount++
				break // Only count one per reel
			}
		}
	}
	return maskCount >= 3
}

// GenerateMaskReelMultiplier generates a random multiplier for the mask reel bonus (2x to 200x)
func GenerateMaskReelMultiplier() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Define possible multipliers with weights
	multipliers := []int{2, 3, 4, 5, 10, 15, 20, 25, 50, 75, 100, 150, 200}
	weights := []float64{0.2, 0.15, 0.15, 0.1, 0.1, 0.08, 0.07, 0.05, 0.04, 0.03, 0.02, 0.01, 0.005}
	
	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}
	
	roll := r.Float64() * totalWeight
	currentWeight := 0.0
	
	for i, weight := range weights {
		currentWeight += weight
		if roll <= currentWeight {
			return multipliers[i]
		}
	}
	
	return 2 // Fallback
}