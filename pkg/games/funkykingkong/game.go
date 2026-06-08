package funkykingkong

import (
	"fmt"
	"math/rand"
)

// Paytable defines the payouts for each symbol combination
// Index 0 = x1 multiplier, Index 1 = x2 multiplier, Index 2 = x3 multiplier
var Paytable = map[string][]int{
	"Kong Kong Kong":          {800, 1600, 2500},
	"Sun Sun Sun":             {400, 800, 1200},
	"Palm Palm Palm":          {200, 400, 600},
	"Coconut Coconut Coconut": {100, 200, 300},
	"Banana Banana Banana":    {80, 160, 240},
	"3BAR 3BAR 3BAR":          {60, 120, 180},
	"2BAR 2BAR 2BAR":          {40, 80, 120},
	"1BAR 1BAR 1BAR":          {20, 40, 60},
	"ANY_3X_BAR":              {10, 20, 30}, // Any combination of BAR symbols
}

// BetAmountToMultiplier maps bet amounts to internal multipliers for each paytable
var BetAmountToMultiplier = map[int]map[float64]int{
	1: { // x1 paytable
		0.01: 1,
		0.05: 5,
		0.1:  10,
		0.2:  20,
		0.25: 25,
	},
	2: { // x2 paytable
		0.02: 1,
		0.1:  5,
		0.2:  10,
		0.4:  20,
		0.5:  25,
	},
	3: { // x3 paytable
		0.03: 1,
		0.15: 5,
		0.3:  10,
		0.6:  20,
		0.75: 25,
	},
}

// AllSymbols contains all possible symbols that can appear on reels
var AllSymbols = []Symbol{
	SymbolKong,
	SymbolSun,
	SymbolPalm,
	SymbolCoconut,
	SymbolBanana,
	Symbol3BAR,
	Symbol2BAR,
	Symbol1BAR,
}

// ValidBetLevels contains valid bet level values
var ValidBetLevels = []int{1, 2, 3}

// GenerateWinningReels generates reels that guarantee a win
func GenerateWinningReels() []string {
	// Randomly select a winning combination
	winningCombinations := [][]string{
		{"Kong", "Kong", "Kong"},
		{"Sun", "Sun", "Sun"},
		{"Palm", "Palm", "Palm"},
		{"Coconut", "Coconut", "Coconut"},
		{"Banana", "Banana", "Banana"},
		{"3BAR", "3BAR", "3BAR"},
		{"2BAR", "2BAR", "2BAR"},
		{"1BAR", "1BAR", "1BAR"},
		// Mixed BAR combinations for ANY_3X_BAR
		{"1BAR", "2BAR", "3BAR"},
		{"3BAR", "1BAR", "2BAR"},
		{"2BAR", "3BAR", "1BAR"},
		{"1BAR", "3BAR", "2BAR"},
		{"2BAR", "1BAR", "3BAR"},
		{"3BAR", "2BAR", "1BAR"},
	}

	selectedCombination := winningCombinations[rand.Intn(len(winningCombinations))]
	return selectedCombination
}

// GenerateLosingReels generates reels that guarantee no win
func GenerateLosingReels() []string {
	// Various losing combinations including empty positions
	losingCombinations := [][]string{
		// Different symbols (no match)
		{"Kong", "Sun", "Palm"},
		{"Banana", "Coconut", "1BAR"},
		{"3BAR", "Sun", "Banana"},
		{"Palm", "Kong", "2BAR"},
		{"Sun", "3BAR", "Coconut"},
		{"1BAR", "Palm", "Kong"},

		// With empty positions (guaranteed no win)
		{"Kong", "EMPTY", "Kong"},
		{"Sun", "EMPTY", "Sun"},
		{"EMPTY", "Kong", "Sun"},
		{"Kong", "Sun", "EMPTY"},
		{"EMPTY", "EMPTY", "Kong"},
		{"Kong", "EMPTY", "EMPTY"},
		{"EMPTY", "EMPTY", "EMPTY"},
		{"Palm", "EMPTY", "Banana"},
		{"EMPTY", "3BAR", "EMPTY"},

		// Two same + one different (no win)
		{"Kong", "Kong", "Sun"},
		{"Banana", "Palm", "Banana"},
		{"1BAR", "2BAR", "1BAR"},
		{"Sun", "Sun", "Kong"},
		{"3BAR", "Coconut", "3BAR"},
	}

	selectedCombination := losingCombinations[rand.Intn(len(losingCombinations))]
	return selectedCombination
}

// CalculateWin determines the win amount and winning combination
// Only considers actual symbols, ignores EMPTY positions
func CalculateWin(reels []string, betLevel int, internalMultiplier int) (float64, string) {
	// Filter out empty positions
	actualSymbols := []string{}
	for _, symbol := range reels {
		if symbol != "EMPTY" {
			actualSymbols = append(actualSymbols, symbol)
		}
	}

	// Need exactly 3 symbols for any win (no wins with empty positions)
	if len(actualSymbols) != 3 {
		return 0, ""
	}

	// Check for exact matches (all three symbols must be the same)
	if actualSymbols[0] == actualSymbols[1] && actualSymbols[1] == actualSymbols[2] {
		combinationKey := fmt.Sprintf("%s %s %s", actualSymbols[0], actualSymbols[1], actualSymbols[2])
		if payout, exists := Paytable[combinationKey]; exists {
			winAmount := float64(payout[betLevel-1]*internalMultiplier) * 0.01
			return winAmount, combinationKey
		}
	}

	// Check for ANY 3X BAR (any combination of BAR symbols, all 3 must be different BARs)
	if isAnyBarCombination(actualSymbols) {
		payout := Paytable["ANY_3X_BAR"]
		winAmount := float64(payout[betLevel-1]*internalMultiplier) * 0.01
		return winAmount, "ANY 3X BAR"
	}

	// No win
	return 0, ""
}

// isAnyBarCombination checks if all three symbols are BAR variants and not all the same
func isAnyBarCombination(reels []string) bool {
	barSymbols := map[string]bool{
		"1BAR": true,
		"2BAR": true,
		"3BAR": true,
	}

	// All symbols must be BAR variants
	for _, symbol := range reels {
		if !barSymbols[symbol] {
			return false
		}
	}

	// Make sure it's not already an exact match (which would be handled above)
	if reels[0] == reels[1] && reels[1] == reels[2] {
		return false
	}

	return true
}

// ValidateBetAmount checks if the bet amount is valid for the given bet level
func ValidateBetAmount(betAmount float64, betLevel int) bool {
	if multiplierMap, exists := BetAmountToMultiplier[betLevel]; exists {
		_, valid := multiplierMap[betAmount]
		return valid
	}
	return false
}

// ValidateBetLevel checks if the bet level is valid
func ValidateBetLevel(betLevel int) bool {
	for _, valid := range ValidBetLevels {
		if betLevel == valid {
			return true
		}
	}
	return false
}

// GetInternalMultiplier gets the internal multiplier for the bet amount and bet level
func GetInternalMultiplier(betAmount float64, betLevel int) int {
	if multiplierMap, exists := BetAmountToMultiplier[betLevel]; exists {
		if multiplier, exists := multiplierMap[betAmount]; exists {
			return multiplier
		}
	}
	return 1 // Default fallback
}

// GetValidBetAmounts returns all valid bet amounts for a given bet level
func GetValidBetAmounts(betLevel int) []float64 {
	var amounts []float64
	if multiplierMap, exists := BetAmountToMultiplier[betLevel]; exists {
		for amount := range multiplierMap {
			amounts = append(amounts, amount)
		}
	}
	return amounts
}
