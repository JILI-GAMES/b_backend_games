package magicaceoriginal

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
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
	SymbolACE:     0.12,
	SymbolKING:    0.12,
	SymbolQUEEN:   0.12,
	SymbolJACK:    0.12,
	SymbolHeart:   0.13,
	SymbolSpade:   0.13,
	SymbolClub:    0.13,
	SymbolDiamond: 0.13,
	SymbolTarget:  0.02,
}

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

// Booming Multipliers for Base Game
var BoomingMultipliers = []int{1, 2, 3, 5}

// Booming Multipliers for Free Spins (doubled values)
var BoomingMultipliersFreeSpins = []int{2, 4, 6, 10}

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

// CountSuperJokers counts the number of Super Joker cards
func CountSuperJokers(jokerCards []JokerCard) int {
	count := 0
	for _, joker := range jokerCards {
		if joker.Mode == ModeSuperJoker {
			count++
		}
	}
	return count
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
				// Add Golden Cards on reels 2, 3, 4 (indices 1, 2, 3) with probability
				if reel >= 1 && reel <= 3 && symbol != SymbolTarget && r.Float64() < GoldenCardProbability {
					specialSymbols.GoldenCards = append(specialSymbols.GoldenCards, pos)
					reels[reel][row] = fmt.Sprintf("golden_%s", string(symbol))
				} else {
					reels[reel][row] = string(symbol)
				}

				if symbol == SymbolTarget {
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

// GenerateLossReels generates a 5x4 grid with no wins
func GenerateLossReels(jokerCards []JokerCard, r *rand.Rand) ([][]string, SpecialSymbols) {
	var reels [][]string
	var specialSymbols SpecialSymbols

	// Keep generating until no wins are found
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

				// Avoid Golden Cards and try to scatter symbols to minimize wins
				availableSymbols := []Symbol{
					SymbolACE, SymbolKING, SymbolQUEEN, SymbolJACK,
					SymbolHeart, SymbolSpade, SymbolClub, SymbolDiamond, SymbolTarget,
				}
				symbol := availableSymbols[r.Intn(len(availableSymbols))]

				// Try to avoid matching adjacent symbols
				if reel > 0 {
					previousSymbol := reels[reel-1][row]
					attempts := 0
					for string(symbol) == previousSymbol && attempts < 5 {
						symbol = availableSymbols[r.Intn(len(availableSymbols))]
						attempts++
					}
				}

				reels[reel][row] = string(symbol)

				if symbol == SymbolTarget {
					specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
				}
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
func GenerateReelsForCascade(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard, lastWinDetails []WinDetail, newlyCreatedJokerPositions map[Position]bool, r *rand.Rand) ([][]string, SpecialSymbols, []JokerCard) {
	// Create a copy of the reels
	newReels := make([][]string, Reels)
	for reel := 0; reel < Reels; reel++ {
		newReels[reel] = make([]string, Rows)
		copy(newReels[reel], reels[reel])
	}

	// STEP 1: Remove joker cards that formed winning combinations (except newly created ones)
	var remainingJokers []JokerCard
	for _, joker := range jokerCards {
		// Check if this joker was newly created in this cascade
		isNewlyCreated := newlyCreatedJokerPositions[joker.Position]

		if isNewlyCreated {
			// Keep newly created jokers - they should persist to next cascade
			remainingJokers = append(remainingJokers, joker)
			log.Printf("Joker at %d,%d was newly created - keeping for next cascade", joker.Position.Reel, joker.Position.Row)
			continue
		}

		// For existing jokers, check if they were in any winning combination
		wasInWin := false
		for _, win := range lastWinDetails {
			for _, pos := range win.Payline {
				if pos.Reel == joker.Position.Reel && pos.Row == joker.Position.Row {
					wasInWin = true
					log.Printf("Existing joker at %d,%d was in win - will be removed", joker.Position.Reel, joker.Position.Row)
					break
				}
			}
			if wasInWin {
				break
			}
		}

		if !wasInWin {
			// Keep jokers that didn't form wins
			remainingJokers = append(remainingJokers, joker)
			log.Printf("Existing joker at %d,%d was not in win - keeping", joker.Position.Reel, joker.Position.Row)
		} else {
			// Replace joker position with random symbol in the reels
			randomSymbol := WeightedRandomSymbol(r)
			for randomSymbol == SymbolWild || randomSymbol == SymbolTarget ||
				strings.HasPrefix(string(randomSymbol), "golden_") {
				randomSymbol = WeightedRandomSymbol(r)
			}
			newReels[joker.Position.Reel][joker.Position.Row] = string(randomSymbol)
			log.Printf("Existing joker at %d,%d was in win - replaced with %s", joker.Position.Reel, joker.Position.Row, randomSymbol)
		}
	}

	// Initialize special symbols
	specialSymbols := SpecialSymbols{
		GoldenCards:   nil,
		JokerCards:    remainingJokers,
		TargetSymbols: nil,
	}

	// STEP 2: Mark remaining joker positions to preserve them
	jokerPositions := make(map[Position]bool)
	if remainingJokers != nil {
		for _, joker := range remainingJokers {
			jokerPositions[joker.Position] = true
			// ENSURE joker positions are set as wild in the new reels
			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
		}
	}

	// STEP 3: Preserve existing Target symbols (outside winning positions)
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			pos := Position{Reel: reel, Row: row}
			if newReels[reel][row] == string(SymbolTarget) && !winningPositions[pos] {
				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
			}
		}
	}

	// STEP 4: Keep generating until a potential win is found
	for {
		// Reset GoldenCards for this iteration
		specialSymbols.GoldenCards = nil

		// Replace only the winning positions that are NOT jokers
		for pos := range winningPositions {
			// CRITICAL: Skip if the position contains a remaining Joker Card
			if jokerPositions[pos] {
				// Ensure the joker position remains as wild
				newReels[pos.Reel][pos.Row] = string(SymbolWild)
				continue
			}

			// Generate new symbol for non-joker positions
			symbol := WeightedRandomSymbol(r)

			// Add Golden Cards on reels 2, 3, 4 with defined probability
			if pos.Reel >= 1 && pos.Reel <= 3 && symbol != SymbolTarget && r.Float64() < GoldenCardProbability {
				specialSymbols.GoldenCards = append(specialSymbols.GoldenCards, pos)
				newReels[pos.Reel][pos.Row] = fmt.Sprintf("golden_%s", string(symbol))
			} else {
				newReels[pos.Reel][pos.Row] = string(symbol)
			}

			// Track Target symbols
			if symbol == SymbolTarget {
				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
			}
		}

		// AFTER generation, ensure all remaining joker positions are still wild
		for _, joker := range remainingJokers {
			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
		}

		// Check for a potential win using the remaining jokers
		totalWinnings, _ := CalculateWins(newReels, 1, 1, remainingJokers)
		if totalWinnings > 0 {
			break
		}
	}

	return newReels, specialSymbols, remainingJokers
}

// GenerateLossForCascade generates new symbols for a cascade with no wins
func GenerateLossForCascade(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard, lastWinDetails []WinDetail, newlyCreatedJokerPositions map[Position]bool, r *rand.Rand) ([][]string, SpecialSymbols, []JokerCard) {
	newReels := make([][]string, Reels)
	for reel := 0; reel < Reels; reel++ {
		newReels[reel] = make([]string, Rows)
		copy(newReels[reel], reels[reel])
	}

	// STEP 1: Remove joker cards that formed winning combinations (except newly created ones)
	var remainingJokers []JokerCard
	for _, joker := range jokerCards {
		// Check if this joker was newly created in this cascade
		isNewlyCreated := newlyCreatedJokerPositions[joker.Position]

		if isNewlyCreated {
			// Keep newly created jokers - they should persist to next cascade
			remainingJokers = append(remainingJokers, joker)
			log.Printf("Joker at %d,%d was newly created - keeping for next cascade", joker.Position.Reel, joker.Position.Row)
			continue
		}

		// For existing jokers, check if they were in any winning combination
		wasInWin := false
		for _, win := range lastWinDetails {
			for _, pos := range win.Payline {
				if pos.Reel == joker.Position.Reel && pos.Row == joker.Position.Row {
					wasInWin = true
					log.Printf("Existing joker at %d,%d was in win - will be removed", joker.Position.Reel, joker.Position.Row)
					break
				}
			}
			if wasInWin {
				break
			}
		}

		if !wasInWin {
			// Keep jokers that didn't form wins
			remainingJokers = append(remainingJokers, joker)
			log.Printf("Existing joker at %d,%d was not in win - keeping", joker.Position.Reel, joker.Position.Row)
		} else {
			// Replace joker position with random symbol in the reels
			randomSymbol := WeightedRandomSymbol(r)
			for randomSymbol == SymbolWild || randomSymbol == SymbolTarget ||
				strings.HasPrefix(string(randomSymbol), "golden_") {
				randomSymbol = WeightedRandomSymbol(r)
			}
			newReels[joker.Position.Reel][joker.Position.Row] = string(randomSymbol)
			log.Printf("Existing joker at %d,%d was in win - replaced with %s", joker.Position.Reel, joker.Position.Row, randomSymbol)
		}
	}

	var specialSymbols SpecialSymbols
	specialSymbols.JokerCards = remainingJokers

	// STEP 2: Mark remaining joker positions to preserve them
	jokerPositions := make(map[Position]bool)
	if remainingJokers != nil {
		for _, joker := range remainingJokers {
			jokerPositions[joker.Position] = true
			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
		}
	}

	// STEP 3: Preserve existing Target symbols
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			pos := Position{Reel: reel, Row: row}
			if newReels[reel][row] == string(SymbolTarget) {
				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
			}
		}
	}

	// STEP 4: Keep generating until no wins are found
	for {
		specialSymbols.GoldenCards = nil

		// Replace only the winning positions
		for pos := range winningPositions {
			// CRITICAL: Don't replace remaining Joker Cards
			if jokerPositions[pos] {
				// Ensure the joker position remains as wild
				newReels[pos.Reel][pos.Row] = string(SymbolWild)
				continue
			}

			// Try to generate symbols that avoid wins
			availableSymbols := []Symbol{
				SymbolACE, SymbolKING, SymbolQUEEN, SymbolJACK,
				SymbolHeart, SymbolSpade, SymbolClub, SymbolDiamond,
			}
			symbol := availableSymbols[r.Intn(len(availableSymbols))]

			// Try to avoid matching patterns with adjacent positions
			if pos.Reel > 0 {
				adjacentSymbol := newReels[pos.Reel-1][pos.Row]
				attempts := 0
				for string(symbol) == adjacentSymbol && attempts < 3 {
					symbol = availableSymbols[r.Intn(len(availableSymbols))]
					attempts++
				}
			}

			newReels[pos.Reel][pos.Row] = string(symbol)

			if symbol == SymbolTarget {
				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
			}
		}

		// AFTER generation, ensure all remaining joker positions are still wild
		for _, joker := range remainingJokers {
			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
		}

		// Check for no wins using the remaining jokers
		totalWinnings, _ := CalculateWins(newReels, 1, 1, remainingJokers)
		if totalWinnings == 0 {
			break
		}
	}

	return newReels, specialSymbols, remainingJokers
}

// // GenerateReelsForCascade generates new symbols for a cascade
// func GenerateReelsForCascade(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard, r *rand.Rand) ([][]string, SpecialSymbols) {
// 	// Create a copy of the reels
// 	newReels := make([][]string, Reels)
// 	for reel := 0; reel < Reels; reel++ {
// 		newReels[reel] = make([]string, Rows)
// 		copy(newReels[reel], reels[reel])
// 	}

// 	// Initialize special symbols
// 	specialSymbols := SpecialSymbols{
// 		GoldenCards:   nil,
// 		JokerCards:    jokerCards,
// 		TargetSymbols: nil,
// 	}

// 	// Mark joker positions to preserve them - THIS IS CRITICAL
// 	jokerPositions := make(map[Position]bool)
// 	if jokerCards != nil {
// 		for _, joker := range jokerCards {
// 			jokerPositions[joker.Position] = true
// 			// ENSURE joker positions are set as wild in the new reels
// 			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
// 		}
// 	}

// 	// Preserve existing Target symbols (outside winning positions)
// 	for reel := 0; reel < Reels; reel++ {
// 		for row := 0; row < Rows; row++ {
// 			pos := Position{Reel: reel, Row: row}
// 			if newReels[reel][row] == string(SymbolTarget) && !winningPositions[pos] {
// 				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
// 			}
// 		}
// 	}

// 	// Keep generating until a potential win is found
// 	for {
// 		// Reset GoldenCards for this iteration
// 		specialSymbols.GoldenCards = nil

// 		// Replace only the winning positions that are NOT jokers
// 		for pos := range winningPositions {
// 			// CRITICAL: Skip if the position contains a Joker Card
// 			if jokerPositions[pos] {
// 				// Ensure the joker position remains as wild
// 				newReels[pos.Reel][pos.Row] = string(SymbolWild)
// 				continue
// 			}

// 			// Generate new symbol for non-joker positions
// 			symbol := WeightedRandomSymbol(r)

// 			// Add Golden Cards on reels 2, 3, 4 with defined probability
// 			if pos.Reel >= 1 && pos.Reel <= 3 && symbol != SymbolTarget && r.Float64() < GoldenCardProbability {
// 				specialSymbols.GoldenCards = append(specialSymbols.GoldenCards, pos)
// 				newReels[pos.Reel][pos.Row] = fmt.Sprintf("golden_%s", string(symbol))
// 			} else {
// 				newReels[pos.Reel][pos.Row] = string(symbol)
// 			}

// 			// Track Target symbols
// 			if symbol == SymbolTarget {
// 				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
// 			}
// 		}

// 		// AFTER generation, ensure all joker positions are still wild
// 		for _, joker := range jokerCards {
// 			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
// 		}

// 		// Check for a potential win
// 		totalWinnings, _ := CalculateWins(newReels, 1, 1, jokerCards)
// 		if totalWinnings > 0 {
// 			break
// 		}
// 	}

// 	return newReels, specialSymbols
// }

// // GenerateLossForCascade generates new symbols for a cascade with no wins
// func GenerateLossForCascade(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard, r *rand.Rand) ([][]string, SpecialSymbols) {
// 	newReels := make([][]string, Reels)
// 	for reel := 0; reel < Reels; reel++ {
// 		newReels[reel] = make([]string, Rows)
// 		copy(newReels[reel], reels[reel])
// 	}

// 	var specialSymbols SpecialSymbols
// 	specialSymbols.JokerCards = jokerCards

// 	// Mark joker positions to preserve them
// 	jokerPositions := make(map[Position]bool)
// 	if jokerCards != nil {
// 		for _, joker := range jokerCards {
// 			jokerPositions[joker.Position] = true
// 			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
// 		}
// 	}

// 	// Preserve existing Target symbols
// 	for reel := 0; reel < Reels; reel++ {
// 		for row := 0; row < Rows; row++ {
// 			pos := Position{Reel: reel, Row: row}
// 			if newReels[reel][row] == string(SymbolTarget) {
// 				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
// 			}
// 		}
// 	}

// 	// Keep generating until no wins are found
// 	for {
// 		specialSymbols.GoldenCards = nil

// 		// Replace only the winning positions
// 		for pos := range winningPositions {
// 			// CRITICAL: Don't replace Joker Cards
// 			if jokerPositions[pos] {
// 				// Ensure the joker position remains as wild
// 				newReels[pos.Reel][pos.Row] = string(SymbolWild)
// 				continue
// 			}

// 			// Try to generate symbols that avoid wins
// 			availableSymbols := []Symbol{
// 				SymbolACE, SymbolKING, SymbolQUEEN, SymbolJACK,
// 				SymbolHeart, SymbolSpade, SymbolClub, SymbolDiamond,
// 			}
// 			symbol := availableSymbols[r.Intn(len(availableSymbols))]

// 			// Try to avoid matching patterns with adjacent positions
// 			if pos.Reel > 0 {
// 				adjacentSymbol := newReels[pos.Reel-1][pos.Row]
// 				attempts := 0
// 				for string(symbol) == adjacentSymbol && attempts < 3 {
// 					symbol = availableSymbols[r.Intn(len(availableSymbols))]
// 					attempts++
// 				}
// 			}

// 			newReels[pos.Reel][pos.Row] = string(symbol)

// 			if symbol == SymbolTarget {
// 				specialSymbols.TargetSymbols = append(specialSymbols.TargetSymbols, pos)
// 			}
// 		}

// 		// AFTER generation, ensure all joker positions are still wild
// 		for _, joker := range jokerCards {
// 			newReels[joker.Position.Reel][joker.Position.Row] = string(SymbolWild)
// 		}

// 		// Check for no wins
// 		totalWinnings, _ := CalculateWins(newReels, 1, 1, jokerCards)
// 		if totalWinnings == 0 {
// 			break
// 		}
// 	}

// 	return newReels, specialSymbols
// }

// CalculateWins calculates the total payout and win details using 1024 ways
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
		if baseSymbol == string(SymbolTarget) {
			continue // Skip Target symbols as they don't form winning combinations
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
					// Round to 2 decimal places
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

	return totalPayout, winDetails
}

// CountTargets counts the number of Target symbols on the reels
func CountTargets(reels [][]string) (int, []Position) {
	count := 0
	var positions []Position
	for reel := 0; reel < Reels; reel++ {
		for row := 0; row < Rows; row++ {
			if reels[reel][row] == string(SymbolTarget) {
				count++
				positions = append(positions, Position{Reel: reel, Row: row})
			}
		}
	}
	return count, positions
}

// TransformGoldenCards transforms Golden Cards that are part of a win into Joker Cards
func TransformGoldenCards(reels [][]string, lastWinDetails []WinDetail, r *rand.Rand) []JokerCard {
	var newJokerCards []JokerCard
	occupiedPositions := make(map[Position]bool)

	// Track which golden card positions have already been transformed
	transformed := make(map[Position]bool)

	for _, win := range lastWinDetails {
		// Build a set of positions in the payline for this win
		paylinePositions := make(map[Position]bool)
		for _, p := range win.Payline {
			paylinePositions[p] = true
		}

		for _, pos := range win.GoldenCards {
			if transformed[pos] {
				continue // Already transformed this position
			}

			// Only transform if this golden card is in the payline
			if !paylinePositions[pos] {
				continue
			}

			symbol := reels[pos.Reel][pos.Row]
			if !strings.HasPrefix(symbol, "golden_") {
				continue // Not a golden card in the current grid
			}

			log.Printf("TRANSFORM: Golden card at position %d,%d will be transformed to a joker", pos.Reel, pos.Row)

			// Use weighted probability for joker types
			roll := r.Intn(100)
			var mode string

			// Probability distribution: Super 20%, Big 30%, Small 50%
			if roll < 50 {
				mode = ModeSuperJoker
			} else if roll < 30 {
				mode = ModeBigJoker
			} else {
				mode = ModeSmallJoker
			}

			joker := JokerCard{
				Position: pos,
				Mode:     mode,
			}

			newJokerCards = append(newJokerCards, joker)
			occupiedPositions[pos] = true
			transformed[pos] = true

			// Replace in reels
			reels[pos.Reel][pos.Row] = string(SymbolWild)
			log.Printf("Created %s joker at position %d,%d", mode, pos.Reel, pos.Row)

			// Big Joker creates a duplicate
			if mode == ModeBigJoker {
				// Try to find a valid position for duplicate
				const maxAttempts = 20
				var newReel, newRow int = -1, -1

				for attempt := 0; attempt < maxAttempts; attempt++ {
					candidateReel := r.Intn(Reels)
					candidateRow := r.Intn(Rows)
					candidatePos := Position{Reel: candidateReel, Row: candidateRow}

					// Check if position is available
					if !occupiedPositions[candidatePos] {
						symbol := reels[candidateReel][candidateRow]
						if symbol != string(SymbolTarget) && !strings.HasPrefix(symbol, "golden_") {
							newReel, newRow = candidateReel, candidateRow
							break
						}
					}
				}

				// If found a valid position, create the duplicate
				if newReel != -1 && newRow != -1 {
					newPos := Position{Reel: newReel, Row: newRow}
					duplicateJoker := JokerCard{
						Position: newPos,
						Mode:     ModeBigJoker,
					}

					newJokerCards = append(newJokerCards, duplicateJoker)
					reels[newReel][newRow] = string(SymbolWild)
					occupiedPositions[newPos] = true

					log.Printf("Created duplicate Big Joker at position %d,%d", newReel, newRow)
				}
			}
		}
	}

	return newJokerCards
}

// UpdateBaseGameCollector updates the base game collector system
func UpdateBaseGameCollector(collector *BaseGameCollector, superJokerCount int) {
	if superJokerCount > 0 && !collector.IsActive {
		// Add to collection count
		collector.CollectedCount += superJokerCount
		log.Printf("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@Base game collector: Added %d Super Jokers, total collected: %d/5", superJokerCount, collector.CollectedCount)

		// Check if we reached 5 to activate
		if collector.CollectedCount >= 5 {
			collector.IsActive = true
			collector.RemainingRounds = 3
			collector.CollectedCount = 5 // Cap at 5
			log.Printf("Base game collector ACTIVATED! 3 rounds of enhanced multipliers")
		}
	} else if superJokerCount > 0 && collector.IsActive {
		// Already active, Super Jokers have no effect
		log.Printf("Base game collector already active, Super Jokers ignored")
	}
}

// UpdateBaseGameCollectorRounds decrements collector rounds on winning spins (not cascades)
func UpdateBaseGameCollectorRounds(collector *BaseGameCollector, hasWins bool) {
	if collector.IsActive && collector.RemainingRounds > 0 && hasWins {
		collector.RemainingRounds--
		log.Printf("Base game collector: Round used, remaining: %d", collector.RemainingRounds)

		if collector.RemainingRounds <= 0 {
			// Reset collector completely
			collector.IsActive = false
			collector.CollectedCount = 0
			collector.RemainingRounds = 0
			log.Printf("Base game collector EXPIRED - completely reset")
		}
	}
}

// UpdateFreeSpinsCollector updates the free spins collector system
func UpdateFreeSpinsCollector(collector *FreeSpinsCollector, superJokerCount int, extraBetEnabled bool) {
	if superJokerCount > 0 {
		// Set max upgrades based on extra bet
		if extraBetEnabled {
			collector.MaxUpgrades = 999 // Unlimited
		} else {
			collector.MaxUpgrades = 10
		}

		// Check if we can still upgrade
		if collector.UpgradeCount < collector.MaxUpgrades {
			newUpgradeCount := collector.UpgradeCount + superJokerCount
			if newUpgradeCount > collector.MaxUpgrades {
				newUpgradeCount = collector.MaxUpgrades
			}

			actualUpgrades := newUpgradeCount - collector.UpgradeCount
			collector.UpgradeCount = newUpgradeCount

			log.Printf("********************Free spins collector: Added %d upgrades, total: %d/%d", actualUpgrades, collector.UpgradeCount, collector.MaxUpgrades)
		} else {
			log.Printf("Free spins collector: Max upgrades reached (%d), Super Jokers ignored", collector.MaxUpgrades)
		}
	}
}

// ResetFreeSpinsCollector resets the free spins collector when returning to base game
func ResetFreeSpinsCollector(collector *FreeSpinsCollector) {
	collector.UpgradeCount = 0
	collector.MaxUpgrades = 10
	log.Printf("Free spins collector RESET for base game")
}

// GetBoomingMultiplier calculates the current booming multiplier
func GetBoomingMultiplier(cascadeCount int, gameMode string, baseGameCollector BaseGameCollector, freeSpinsCollector FreeSpinsCollector) int {
	index := cascadeCount - 1
	if index < 0 {
		index = 0
	}

	var multipliers []int

	if gameMode == "freeSpins" {
		// Start with base free spins multipliers
		multipliers = make([]int, len(BoomingMultipliersFreeSpins))
		copy(multipliers, BoomingMultipliersFreeSpins)

		// Add free spins collector upgrades
		for i := range multipliers {
			multipliers[i] += freeSpinsCollector.UpgradeCount
		}
	} else {
		// Base game multipliers
		multipliers = make([]int, len(BoomingMultipliers))
		copy(multipliers, BoomingMultipliers)

		// Add base game collector upgrade (+1 when active)
		if baseGameCollector.IsActive {
			for i := range multipliers {
				multipliers[i] += 1
			}
		}
	}

	if index < len(multipliers) {
		return multipliers[index]
	} else {
		// For cascades beyond the array, always use the last value
		return multipliers[len(multipliers)-1]
	}
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
		if baseNextSymbol == string(symbol) || nextSymbol == string(SymbolWild) || (isJoker && string(symbol) != string(SymbolTarget)) {
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
