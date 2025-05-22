package magicace

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"
)

// Constants
const (
	Denomination          = 0.01
	GoldenCardProbability = 0.20
	Reels                 = 5
	Rows                  = 4
	MinBet                = 20
)

// Symbol weights for random generation
var SymbolWeights = map[Symbol]float64{
	SymbolACE:     0.1,
	SymbolKING:    0.1,
	SymbolQUEEN:   0.1,
	SymbolJACK:    0.1,
	SymbolHeart:   0.1,
	SymbolSpade:   0.1,
	SymbolClub:    0.1,
	SymbolDiamond: 0.1,
	SymbolScatter: 0.04,
}

// global constant
var globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// Paytable (payouts for Bet Multiplier = 1)
var Paytable = map[Symbol]map[int]float64{
	SymbolACE:     {3: 10, 4: 20, 5: 50},
	SymbolKING:    {3: 8, 4: 16, 5: 40},
	SymbolQUEEN:   {3: 6, 4: 12, 5: 30},
	SymbolJACK:    {3: 4, 4: 8, 5: 20},
	SymbolHeart:   {3: 2, 4: 4, 5: 10},
	SymbolSpade:   {3: 2, 4: 4, 5: 10},
	SymbolClub:    {3: 1, 4: 2, 5: 5},
	SymbolDiamond: {3: 1, 4: 2, 5: 5},
}

// Booming Multipliers
var (
	BoomingMultipliers                  = []int{1, 2, 3, 5}
	BoomingMultipliersFreeSpins         = []int{2, 4, 6, 10}
	BoomingMultipliersExtraBet          = []int{1, 2, 3} // Will append dynamically for unlimited
	BoomingMultipliersFreeSpinsExtraBet = []int{2, 4, 6} // Will append dynamically
)

// BetAmountToMultiplier maps bet amounts to multipliers
var BetAmountToMultiplier = map[float64]int{
	0.2: 1,
	0.4: 2,
	0.6: 3,
	1.0: 5,
	2.0: 10,
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

// WeightedRandomSymbol selects a symbol based on weights
func WeightedRandomSymbol(r *rand.Rand) Symbol {
	totalWeight := 0.0
	for _, weight := range SymbolWeights {
		totalWeight += weight
	}

	roll := r.Float64() * totalWeight
	currentWeight := 0.0
	for symbol, weight := range SymbolWeights {
		currentWeight += weight
		if roll <= currentWeight {
			return symbol
		}
	}
	return SymbolACE // Fallback
}

// GenerateReelsWithWin generates a 5x4 grid with a potential win
func GenerateReelsWithWin(jokerCards []JokerCard, r *rand.Rand) ([][]string, SpecialSymbols) {
	var reels [][]string
	var specialSymbols SpecialSymbols

	// Keep generating until a potential win is found
	for {
		reels = make([][]string, Reels)
		specialSymbols = SpecialSymbols{
			GoldenCards:   nil,
			JokerCards:    jokerCards,
			TargetSymbols: nil,
		}

		// Place Joker Cards as Wilds
		reelMap := make(map[Position]string)
		if jokerCards != nil {
			for _, joker := range jokerCards {
				pos := Position{Reel: joker.Position.Reel, Row: joker.Position.Row}
				reelMap[pos] = string(SymbolWild)
			}
		}

		// Generate reels
		for reel := 0; reel < Reels; reel++ {
			reels[reel] = make([]string, Rows)
			for row := 0; row < Rows; row++ {
				pos := Position{Reel: reel, Row: row}
				if symbol, exists := reelMap[pos]; exists {
					reels[reel][row] = symbol
					continue
				}

				symbol := WeightedRandomSymbol(r)
				// Add Golden Cards on reels 2, 3, 4 (indices 1, 2, 3) with probability defined in constants
				if reel >= 1 && reel <= 3 && symbol != SymbolScatter && r.Float64() < GoldenCardProbability {
					specialSymbols.GoldenCards = append(specialSymbols.GoldenCards, pos)
					reels[reel][row] = fmt.Sprintf("golden_%s", string(symbol))
				} else {
					reels[reel][row] = string(symbol)
				}

				if symbol == SymbolScatter {
					specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
				}
			}
		}

		// Check for a potential win
		betMultiplier := 1 // Use a default multiplier for checking
		totalWinnings, _ := CalculateWins(reels, betMultiplier, 1, jokerCards)
		if totalWinnings > 0 {
			break
		}
	}

	return reels, specialSymbols
}

// GenerateLossReels generates a 5x4 grid with no wins (Improvement #1: Removed RNG calls)
func GenerateLossReels(jokerCards []JokerCard, r *rand.Rand) ([][]string, SpecialSymbols) {
	var reels [][]string
	var specialSymbols SpecialSymbols

	// Keep generating until no wins are found (Improvement #6 skipped, no deterministic construction)
	for {
		reels = make([][]string, Reels)
		specialSymbols = SpecialSymbols{
			GoldenCards:   nil,
			JokerCards:    jokerCards,
			TargetSymbols: nil,
		}

		// Place Joker Cards as Wilds
		reelMap := make(map[Position]string)
		if jokerCards != nil {
			for _, joker := range jokerCards {
				pos := Position{Reel: joker.Position.Reel, Row: joker.Position.Row}
				reelMap[pos] = string(SymbolWild)
			}
		}

		// Generate reels to avoid wins
		for reel := 0; reel < Reels; reel++ {
			reels[reel] = make([]string, Rows)
			for row := 0; row < Rows; row++ {
				pos := Position{Reel: reel, Row: row}
				if symbol, exists := reelMap[pos]; exists {
					reels[reel][row] = symbol
					continue
				}

				// Avoid Golden Cards and Scatters to minimize wins
				availableSymbols := []Symbol{
					SymbolACE, SymbolKING, SymbolQUEEN, SymbolJACK,
					SymbolHeart, SymbolSpade, SymbolClub, SymbolDiamond,
				}
				symbol := availableSymbols[r.Intn(len(availableSymbols))]
				// Scatter symbols to avoid matches
				if reel > 0 {
					previousSymbol := reels[reel-1][row]
					for {
						symbol = availableSymbols[r.Intn(len(availableSymbols))]
						if string(symbol) != previousSymbol {
							break
						}
					}
				}
				reels[reel][row] = string(symbol)
			}
		}

		// Check for no wins
		totalWinnings, _ := CalculateWins(reels, 1, 1, jokerCards)
		if totalWinnings == 0 {
			break
		}
	}

	return reels, specialSymbols
}

// GenerateReelsForCascade generates new symbols for a cascade
func GenerateReelsForCascade(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard, r *rand.Rand) ([][]string, SpecialSymbols) {
	// Create a copy of the reels
	newReels := make([][]string, Reels)
	for reel := 0; reel < Reels; reel++ {
		newReels[reel] = make([]string, Rows)
		copy(newReels[reel], reels[reel])
	}

	log.Printf("NEW REELS ENTRY: %v", newReels)

	// Initialize special symbols
	specialSymbols := SpecialSymbols{
		GoldenCards:      nil,
		JokerCards:       jokerCards,
		TargetSymbols:    nil,
		NewTargetSymbols: nil,
	}

	// Mark joker positions to preserve them
	jokerPositions := make(map[Position]bool)
	if jokerCards != nil {
		for _, joker := range jokerCards {
			jokerPositions[joker.Position] = true
			// Make sure jokers are shown as wild in the reels
			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
		}
	}

	// Preserve existing Scatters (outside winning positions)
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			pos := Position{Reel: reel, Row: row}
			if newReels[reel][row] == string(SymbolScatter) && !winningPositions[pos] {
				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
			}
		}
	}

	// Keep generating until a potential win is found
	for {
		// Reset GoldenCards and NewTargetSymbols for this iteration
		specialSymbols.GoldenCards = nil
		specialSymbols.NewTargetSymbols = nil

		// Replace only the winning positions that are not jokers
		for pos := range winningPositions {
			// Skip if the position contains a Joker Card
			if jokerPositions[pos] {
				continue
			}

			// Only allow regular symbols (no wild, scatter, or golden)
			var symbol Symbol
			for {
				symbol = WeightedRandomSymbol(r)
				if symbol != SymbolWild && symbol != SymbolScatter && !strings.HasPrefix(string(symbol), "golden_") {
					break
				}
			}
			// Add Golden Cards on reels 2, 3, 4 with defined probability
			if pos.Reel >= 1 && pos.Reel <= 3 && r.Float64() < GoldenCardProbability {
				specialSymbols.GoldenCards = append(specialSymbols.GoldenCards, pos)
				newReels[pos.Reel][pos.Row] = fmt.Sprintf("golden_%s", string(symbol))
			} else {
				newReels[pos.Reel][pos.Row] = string(symbol)
			}

			// Track new Scatters for Free Spin retriggers (should not happen here, but keep for consistency)
			if symbol == SymbolScatter {
				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
				specialSymbols.NewTargetSymbols = append(specialSymbols.NewTargetSymbols, pos)
			}
		}

		// Check for a potential win
		totalWinnings, _ := CalculateWins(newReels, 1, 1, jokerCards)
		if totalWinnings > 0 {
			break
		}
	}
	log.Printf("NEW REELS EXIT: %v", newReels)

	return newReels, specialSymbols
}

// GenerateLossForCascade generates new symbols for a cascade with no wins (Improvement #1: Removed RNG calls)
func GenerateLossForCascade(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard, r *rand.Rand) ([][]string, SpecialSymbols) {
	newReels := make([][]string, Reels)
	for reel := 0; reel < Reels; reel++ {
		newReels[reel] = make([]string, Rows)
		copy(newReels[reel], reels[reel])
	}

	var specialSymbols SpecialSymbols
	specialSymbols.JokerCards = jokerCards

	// Preserve existing Scatters and Jokers
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			pos := Position{Reel: reel, Row: row}
			if newReels[reel][row] == string(SymbolScatter) {
				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
			}
		}
	}

	// Keep generating until no wins are found (Improvement #6 skipped, no deterministic construction)
	for {
		specialSymbols.GoldenCards = nil
		specialSymbols.NewTargetSymbols = nil

		// Replace only the winning positions
		for pos := range winningPositions {
			// Don't replace Joker Cards
			isJoker := false
			if jokerCards != nil {
				for _, joker := range jokerCards {
					if joker.Position.Reel == pos.Reel && joker.Position.Row == pos.Row {
						isJoker = true
						break
					}
				}
			}
			if isJoker {
				continue
			}

			// Avoid Golden Cards and Scatters
			availableSymbols := []Symbol{
				SymbolACE, SymbolKING, SymbolQUEEN, SymbolJACK,
				SymbolHeart, SymbolSpade, SymbolClub, SymbolDiamond,
			}
			symbol := availableSymbols[r.Intn(len(availableSymbols))]
			// Avoid matching the previous symbol
			if pos.Reel > 0 {
				previousSymbol := newReels[pos.Reel-1][pos.Row]
				for _, joker := range jokerCards {
					if joker.Position.Reel == pos.Reel-1 && joker.Position.Row == pos.Row {
						previousSymbol = string(SymbolWild)
						break
					}
				}
				for {
					symbol = availableSymbols[r.Intn(len(availableSymbols))]
					if string(symbol) != previousSymbol {
						break
					}
				}
			}
			newReels[pos.Reel][pos.Row] = string(symbol)
		}

		// Check for no wins
		totalWinnings, _ := CalculateWins(newReels, 1, 1, jokerCards)
		if totalWinnings == 0 {
			break
		}
	}

	return newReels, specialSymbols
}

// GenerateLossForCascadePreservingJokers that ensures jokers are preserved
func GenerateLossForCascadePreservingJokers(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard, r *rand.Rand) ([][]string, SpecialSymbols, bool) {
	newReels := make([][]string, Reels)
	for reel := 0; reel < Reels; reel++ {
		newReels[reel] = make([]string, Rows)
		copy(newReels[reel], reels[reel])
	}

	var specialSymbols SpecialSymbols
	specialSymbols.JokerCards = jokerCards

	// Track positions of jokers for fast lookup
	jokerPositions := make(map[Position]bool)
	if jokerCards != nil {
		for _, joker := range jokerCards {
			jokerPositions[joker.Position] = true
			// Make sure jokers are represented as Wild in the reels
			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
		}
	}

	// Clear out the target symbols before collecting them
	specialSymbols.TargetSymbols = nil
	specialSymbols.NewTargetSymbols = nil

	// Preserve existing Scatters
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			if newReels[reel][row] == string(SymbolScatter) {
				pos := Position{Reel: reel, Row: row}
				// If this position was a winning position, it's a new scatter
				if winningPositions[pos] {
					specialSymbols.NewTargetSymbols = append(specialSymbols.NewTargetSymbols, pos)
				}
				// Add to the overall target symbols
				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
			}
		}
	}

	// Reduced attempts for better performance
	maxAttempts := 100
	attempts := 0

	for attempts < maxAttempts {
		attempts++
		specialSymbols.GoldenCards = nil
		specialSymbols.NewTargetSymbols = nil

		// Replace only the winning positions that are not jokers
		for pos := range winningPositions {
			// Skip positions that are now jokers
			if jokerPositions[pos] {
				continue
			}

			// For non-joker positions, try to generate symbols that won't form wins
			availableSymbols := []Symbol{
				SymbolACE, SymbolKING, SymbolQUEEN, SymbolJACK,
				SymbolHeart, SymbolSpade, SymbolClub, SymbolDiamond, SymbolScatter,
			}

			hasNearbyJoker := false
			if jokerCards != nil {
				for _, joker := range jokerCards {
					if (joker.Position.Reel == pos.Reel-1 || joker.Position.Reel == pos.Reel+1) &&
						(joker.Position.Row == pos.Row) {
						hasNearbyJoker = true
						break
					}
				}
			}

			if hasNearbyJoker {
				// Try multiple symbols until we find one unlikely to form a win
				symbolsToTry := make([]Symbol, len(availableSymbols))
				copy(symbolsToTry, availableSymbols)
				r.Shuffle(len(symbolsToTry), func(i, j int) { symbolsToTry[i], symbolsToTry[j] = symbolsToTry[j], symbolsToTry[i] })

				foundNonMatchingSymbol := false
				for _, trySymbol := range symbolsToTry {
					newReels[pos.Reel][pos.Row] = string(trySymbol)
					tempWinnings, _ := CalculateWins(newReels, 1, 1, jokerCards)
					if tempWinnings == 0 {
						foundNonMatchingSymbol = true
						break
					}
				}

				if !foundNonMatchingSymbol {
					newReels[pos.Reel][pos.Row] = string(availableSymbols[r.Intn(len(availableSymbols))])
				}
			} else {
				newReels[pos.Reel][pos.Row] = string(availableSymbols[r.Intn(len(availableSymbols))])
			}
		}

		// Check if we have a valid loss pattern
		totalWinnings, _ := CalculateWins(newReels, 1, 1, jokerCards)
		if totalWinnings == 0 {
			log.Printf("LOSS GENERATED: Successfully created loss after %d attempts", attempts)
			return newReels, specialSymbols, false // false = no win override
		}
	}

	// If we reach here, accept the win instead of forcing impossible loss
	log.Printf("✅✅✅✅WIN ACCEPTED: Could not generate loss after %d attempts, accepting natural win", maxAttempts)
	
	// Return original reels with jokers properly placed
	for reel := 0; reel < Reels; reel++ {
		copy(newReels[reel], reels[reel])
	}
	
	// Ensure jokers are in their positions
	if jokerCards != nil {
		for _, joker := range jokerCards {
			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
		}
	}
	
	return newReels, specialSymbols, true // true = win override applied
}
// func GenerateLossForCascadePreservingJokers(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard, r *rand.Rand) ([][]string, SpecialSymbols) {
// 	newReels := make([][]string, Reels)
// 	for reel := 0; reel < Reels; reel++ {
// 		newReels[reel] = make([]string, Rows)
// 		copy(newReels[reel], reels[reel])
// 	}

// 	var specialSymbols SpecialSymbols
// 	specialSymbols.JokerCards = jokerCards

// 	// Track positions of jokers for fast lookup
// 	jokerPositions := make(map[Position]bool)
// 	if jokerCards != nil {
// 		for _, joker := range jokerCards {
// 			jokerPositions[joker.Position] = true
// 			// Make sure jokers are represented as Wild in the reels
// 			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
// 		}
// 	}

// 	// Clear out the target symbols before collecting them
// 	specialSymbols.TargetSymbols = nil
// 	specialSymbols.NewTargetSymbols = nil

// 	// Preserve existing Scatters
// 	for reel := 0; reel < Reels; reel++ {
// 		for row := 0; row < Rows; row++ {
// 			if newReels[reel][row] == string(SymbolScatter) {
// 				pos := Position{Reel: reel, Row: row}
// 				// If this position was a winning position, it's a new scatter
// 				if winningPositions[pos] {
// 					specialSymbols.NewTargetSymbols = append(specialSymbols.NewTargetSymbols, pos)
// 				}
// 				// Add to the overall target symbols
// 				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
// 			}
// 		}
// 	}

// 	// Maximum attempts to prevent infinite loop
// 	maxAttempts := 500
// 	attempts := 0

// 	var lastWinnings float64
// 	var lastGrid [][]string
// 	var lastSpecialSymbols SpecialSymbols

// 	for attempts < maxAttempts {
// 		attempts++
// 		specialSymbols.GoldenCards = nil
// 		specialSymbols.NewTargetSymbols = nil

// 		// Replace only the winning positions that are not jokers
// 		for pos := range winningPositions {
// 			// Skip positions that are now jokers
// 			if jokerPositions[pos] {
// 				continue
// 			}

// 			// For non-joker positions, try to generate symbols that won't form wins
// 			// Prioritize positions around jokers to break potential winning patterns
// 			availableSymbols := []Symbol{
// 				SymbolACE, SymbolKING, SymbolQUEEN, SymbolJACK,
// 				SymbolHeart, SymbolSpade, SymbolClub, SymbolDiamond, SymbolScatter,
// 			}

// 			hasNearbyJoker := false
// 			if jokerCards != nil {
// 				for _, joker := range jokerCards {
// 					if (joker.Position.Reel == pos.Reel-1 || joker.Position.Reel == pos.Reel+1) &&
// 						(joker.Position.Row == pos.Row) {
// 						hasNearbyJoker = true
// 						break
// 					}
// 				}
// 			}

// 			if hasNearbyJoker {
// 				// Try multiple symbols until we find one unlikely to form a win
// 				symbolsToTry := make([]Symbol, len(availableSymbols))
// 				copy(symbolsToTry, availableSymbols)
// 				r.Shuffle(len(symbolsToTry), func(i, j int) { symbolsToTry[i], symbolsToTry[j] = symbolsToTry[j], symbolsToTry[i] })

// 				foundNonMatchingSymbol := false
// 				for _, trySymbol := range symbolsToTry {
// 					newReels[pos.Reel][pos.Row] = string(trySymbol)
// 					tempWinnings, _ := CalculateWins(newReels, 1, 1, jokerCards)
// 					if tempWinnings == 0 {
// 						foundNonMatchingSymbol = true
// 						break
// 					}
// 				}

// 				if !foundNonMatchingSymbol {
// 					newReels[pos.Reel][pos.Row] = string(availableSymbols[r.Intn(len(availableSymbols))])
// 				}
// 			} else {
// 				newReels[pos.Reel][pos.Row] = string(availableSymbols[r.Intn(len(availableSymbols))])
// 			}
// 		}

// 		// Check if we have a valid loss pattern
// 		totalWinnings, _ := CalculateWins(newReels, 1, 1, jokerCards)
// 		if totalWinnings == 0 {
// 			return newReels, specialSymbols
// 		}
// 		// Save the last attempted grid and win details
// 		lastWinnings = totalWinnings
// 		// Deep copy the grid and specialSymbols for override return
// 		lastGrid = make([][]string, Reels)
// 		for i := range newReels {
// 			lastGrid[i] = make([]string, Rows)
// 			copy(lastGrid[i], newReels[i])
// 		}
// 		lastSpecialSymbols = specialSymbols
// 	}

// 	// If we reach here, it's impossible to avoid a win with Jokers present
// 	log.Printf("HYBRID OVERRIDE: Could not generate a loss grid with Jokers after %d attempts. Allowing unavoidable win (payout=%.2f)", maxAttempts, lastWinnings)
// 	// Optionally, annotate specialSymbols for auditing (add a field if needed)
// 	// lastSpecialSymbols.HybridOverride = true
// 	return lastGrid, lastSpecialSymbols
// }

// CalculateWins calculates the total payout and win details, counting only the longest paths
func CalculateWins(reels [][]string, betMultiplier int, boomingMultiplier int, jokerCards []JokerCard) (float64, []WinDetail) {
	var winDetails []WinDetail
	seenWins := make(map[string]bool) // Track unique wins to avoid duplicates

	// Iterate over each starting position in Reel 0
	for row0 := 0; row0 < Rows; row0++ {
		symbol := reels[0][row0]
		baseSymbol := symbol
		if strings.HasPrefix(symbol, "golden_") {
			baseSymbol = strings.TrimPrefix(symbol, "golden_")
		}
		if baseSymbol == string(SymbolScatter) {
			continue // Skip Scatters as they don't form winning combinations
		}

		// Start building paths from this position
		initialPath := Path{
			Positions:   []Position{{Reel: 0, Row: row0}},
			Symbols:     []string{symbol},
			GoldenCards: nil,
		}
		if strings.HasPrefix(symbol, "golden_") {
			initialPath.GoldenCards = append(initialPath.GoldenCards, Position{Reel: 0, Row: row0})
		}

		// Find all possible paths for this symbol starting at this position
		paths := findAllPaths(reels, jokerCards, Symbol(baseSymbol), 0, row0, initialPath)

		// Filter to keep only the longest paths
		longestPaths := filterLongestPaths(paths)

		// Process each longest path
		for _, path := range longestPaths {
			// Create a unique key for the path to avoid duplicates
			var positions []string
			for _, pos := range path.Positions {
				positions = append(positions, fmt.Sprintf("%d,%d", pos.Reel, pos.Row))
			}
			winKey := strings.Join(positions, "|")
			if seenWins[winKey] {
				continue // Skip if this path has already been counted
			}
			seenWins[winKey] = true

			// Calculate payout
			matchCount := len(path.Positions)
			if payoutMap, exists := Paytable[Symbol(baseSymbol)]; exists {
				if payout, found := payoutMap[matchCount]; found {
					payout = payout * Denomination * float64(betMultiplier) * float64(boomingMultiplier)
					// payout to be rounded to 2 decimal places
					payout = math.Round(payout*100) / 100

					// Add to win details
					winDetails = append(winDetails, WinDetail{
						Symbols:     path.Symbols,
						Payline:     path.Positions,
						Payout:      payout,
						GoldenCards: path.GoldenCards,
					})
				}
			}
		}
	}

	// Calculate total payout
	totalPayout := 0.0
	for _, win := range winDetails {
		totalPayout += win.Payout
	}

	// log.Printf("WIN DETAILS: %v", winDetails)
	// log.Printf("TOTAL PAYOUT: %v", totalPayout)

	return totalPayout, winDetails
}

// CountScatters counts the number of Scatter symbols on the reels
func CountScatters(reels [][]string) (int, []Position) {
	count := 0
	var positions []Position
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			if reels[reel][row] == string(SymbolScatter) {
				count++
				positions = append(positions, Position{Reel: reel, Row: row})
			}
		}
	}
	return count, positions
}

// Helper function to count golden cards in win details
func countGoldenCards(winDetails []WinDetail) int {
	count := 0
	for _, win := range winDetails {
		count += len(win.GoldenCards)
	}
	return count
}

// TransformGoldenCards transforms Golden Cards that are part of a win into Joker Cards
func TransformGoldenCards(reels [][]string, lastWinDetails []WinDetail, existingJokerCards []JokerCard, r *rand.Rand) []JokerCard {
	log.Printf("TransformGoldenCards: Examining %d golden cards from %d win details",
		countGoldenCards(lastWinDetails), len(lastWinDetails))

	var newJokerCards []JokerCard
	occupiedPositions := make(map[Position]bool)

	// Initialize with existing joker positions
	if existingJokerCards != nil {
		for _, joker := range existingJokerCards {
			occupiedPositions[joker.Position] = true
			log.Printf("Existing joker at position %d,%d marked as occupied",
				joker.Position.Reel, joker.Position.Row)
		}
	}

	// Track other occupied positions
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			symbol := reels[reel][row]
			if strings.Contains(symbol, "wild") || symbol == string(SymbolScatter) {
				occupiedPositions[Position{Reel: reel, Row: row}] = true
				if strings.HasPrefix(symbol, "golden_") {
					log.Printf("Golden card at position %d,%d marked as occupied", reel, row)
				}
			}
		}
	}

	// Track which golden card positions have already been transformed
	transformed := make(map[Position]bool)

	for winIndex, win := range lastWinDetails {
		log.Printf("Processing win #%d with %d golden cards", winIndex, len(win.GoldenCards))

		// Build a set of positions in the payline for this win
		paylinePositions := make(map[Position]bool)
		for _, p := range win.Payline {
			paylinePositions[p] = true
			log.Printf("Payline includes position %d,%d", p.Reel, p.Row)
		}

		for _, pos := range win.GoldenCards {
			log.Printf("Checking golden card at position %d,%d", pos.Reel, pos.Row)

			if transformed[pos] {
				log.Printf("Position %d,%d already transformed - skipping", pos.Reel, pos.Row)
				continue // Already transformed this position
			}

			// Only transform if this golden card is also in the payline of this win
			if !paylinePositions[pos] {
				log.Printf("Position %d,%d not in current payline - skipping", pos.Reel, pos.Row)
				continue
			}

			symbol := reels[pos.Reel][pos.Row]
			log.Printf("Symbol at position %d,%d is %s", pos.Reel, pos.Row, symbol)

			if !strings.HasPrefix(symbol, "golden_") {
				log.Printf("Symbol %s is not a golden card - skipping", symbol)
				continue // Not a golden card in the current grid
			}

			// Check if a joker already exists at this position
			if occupiedPositions[pos] {
				log.Printf("Position %d,%d already occupied - skipping", pos.Reel, pos.Row)
				continue // Already a joker here
			}

			log.Printf("TRANSFORM: Golden card at position %d,%d will be transformed to a joker", pos.Reel, pos.Row)

			// Use a weighted probability distribution for joker types
			roll := r.Intn(100)
			var mode string
			var remainingRounds int

			// Probability distribution:
			// - Super Joker: 20% chance (rare but powerful)
			// - Big Joker: 30% chance (medium rarity)
			// - Small Joker: 50% chance (most common)
			if roll < 20 {
				mode = ModeSuperJoker
				remainingRounds = 3 // Start with 3 rounds for Super Joker
				log.Printf("Selected Super Joker (roll: %d)", roll)
			} else if roll < 80 {
				mode = ModeBigJoker
				remainingRounds = 1
				log.Printf("Selected Big Joker (roll: %d)", roll)
			} else {
				mode = ModeSmallJoker
				remainingRounds = 1
				log.Printf("Selected Small Joker (roll: %d)", roll)
			}

			joker := JokerCard{
				Position:        pos,
				Mode:            mode,
				RemainingRounds: remainingRounds,
			}

			newJokerCards = append(newJokerCards, joker)
			occupiedPositions[pos] = true
			transformed[pos] = true

			// Mutate the reels to show the joker (wild)
			reels[pos.Reel][pos.Row] = string(SymbolWild)
			log.Printf("Replaced golden_%s with wild at position %d,%d",
				strings.TrimPrefix(symbol, "golden_"), pos.Reel, pos.Row)

			// Add duplication logic for Big Joker
			if mode == ModeBigJoker {
				// Try to find a valid position for duplicate in reels 2-5
				const maxAttempts = 20
				var newReel, newRow int = -1, -1

				for attempt := 0; attempt < maxAttempts; attempt++ {
					// Generate reel number from 1-4 and add 1 to get reels 2-5
					candidateReel := r.Intn(4) + 1
					candidateRow := r.Intn(Rows)
					candidatePos := Position{Reel: candidateReel, Row: candidateRow}

					// Check if position is available (not occupied and not a special symbol)
					if !occupiedPositions[candidatePos] {
						symbol := reels[candidateReel][candidateRow]
						if symbol != string(SymbolScatter) && !strings.HasPrefix(symbol, "golden_") {
							newReel, newRow = candidateReel, candidateRow
							break
						}
					}
				}

				// If found a valid position, create the duplicate
				if newReel != -1 && newRow != -1 {
					newPos := Position{Reel: newReel, Row: newRow}
					duplicateJoker := JokerCard{
						Position:        newPos,
						Mode:            ModeBigJoker,
						RemainingRounds: remainingRounds, // Same as original
					}

					newJokerCards = append(newJokerCards, duplicateJoker)
					reels[newReel][newRow] = string(SymbolWild)
					occupiedPositions[newPos] = true

					log.Printf("Created duplicate Big Joker at position %d,%d",
						newReel, newRow)
				} else {
					log.Printf("Could not find position for duplicate Big Joker after %d attempts",
						maxAttempts)
				}
			}
		}
	}

	log.Printf("TransformGoldenCards: Created %d new joker cards", len(newJokerCards))
	return newJokerCards
}

// getRandomPosition finds a random position on the reels, avoiding the given position and occupied positions (Improvement #3)
func getRandomPosition(reels [][]string, r *rand.Rand, excludeReel, excludeRow int, occupiedPositions map[Position]bool) (int, int) {
	// Maximum attempts to prevent infinite loop
	maxAttempts := 100
	attempts := 0

	for attempts < maxAttempts {
		attempts++
		reel := r.Intn(Reels)
		row := r.Intn(Rows)
		pos := Position{Reel: reel, Row: row}
		symbol := reels[reel][row]
		if (reel != excludeReel || row != excludeRow) &&
			!occupiedPositions[pos] &&
			symbol != string(SymbolScatter) &&
			!strings.HasPrefix(symbol, "golden_") {
			return reel, row
		}
	}

	// If we couldn't find a suitable position after maximum attempts, log and return invalid
	log.Printf("Warning: Could not find suitable random position after %d attempts", maxAttempts)
	return -1, -1
}

// Path represents a potential winning path
type Path struct {
	Positions   []Position
	Symbols     []string
	GoldenCards []Position
}

// findAllPaths recursively finds all possible paths for a symbol starting at a given position
func findAllPaths(reels [][]string, jokerCards []JokerCard, symbol Symbol, currentReel, currentRow int, currentPath Path) []Path {
	var paths []Path

	// Base case: if we're at the last reel, stop recursion
	if currentReel == Reels-1 {
		if len(currentPath.Positions) >= 3 { // Minimum 3 symbols for a win
			paths = append(paths, currentPath)
		}
		return paths
	}

	// Try each row in the next reel
	nextReel := currentReel + 1
	for nextRow := 0; nextRow < Rows; nextRow++ {
		nextPos := Position{Reel: nextReel, Row: nextRow}
		nextSymbol := reels[nextReel][nextRow]

		// Check for Joker Cards
		isJoker := false
		if jokerCards != nil {
			for _, joker := range jokerCards {
				if joker.Position.Reel == nextReel && joker.Position.Row == nextRow {
					nextSymbol = string(SymbolWild)
					isJoker = true
					break
				}
			}
		}

		// Check if the symbol matches (including wilds)
		baseNextSymbol := nextSymbol
		if strings.HasPrefix(nextSymbol, "golden_") {
			baseNextSymbol = strings.TrimPrefix(nextSymbol, "golden_")
		}
		if baseNextSymbol == string(symbol) || nextSymbol == string(SymbolWild) || (isJoker && string(symbol) != string(SymbolScatter)) {
			newPath := Path{
				Positions:   append([]Position{}, currentPath.Positions...),
				Symbols:     append([]string{}, currentPath.Symbols...),
				GoldenCards: append([]Position{}, currentPath.GoldenCards...),
			}
			newPath.Positions = append(newPath.Positions, nextPos)
			newPath.Symbols = append(newPath.Symbols, nextSymbol)
			if strings.HasPrefix(nextSymbol, "golden_") {
				newPath.GoldenCards = append(newPath.GoldenCards, nextPos)
			}
			subPaths := findAllPaths(reels, jokerCards, symbol, nextReel, nextRow, newPath)
			paths = append(paths, subPaths...)
		}
	}

	// If the current path has at least 3 symbols, it might be a valid win
	if len(currentPath.Positions) >= 3 {
		paths = append(paths, currentPath)
	}

	return paths
}

// filterLongestPaths keeps only the paths with the maximum length
func filterLongestPaths(paths []Path) []Path {
	if len(paths) == 0 {
		return nil
	}

	// Find the maximum length
	maxLength := 0
	for _, path := range paths {
		if len(path.Positions) > maxLength {
			maxLength = len(path.Positions)
		}
	}

	// Keep only paths with the maximum length
	var longestPaths []Path
	for _, path := range paths {
		if len(path.Positions) == maxLength {
			longestPaths = append(longestPaths, path)
		}
	}

	return longestPaths
}
