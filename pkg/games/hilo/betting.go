package hilo

import (
	"math"
)

// Base multipliers for each card (from JDB screenshot)
// All cards of the same rank have the same multipliers regardless of suit
var baseMultipliers = map[string]map[string]float64{
	// ACE cards
	"ACE_SPADES": {
		"higher":         1.04,
		"lower":          0.0, // NONE
		"same":           16.69,
		"higher_or_same": 0.0, // NONE
		"lower_or_same":  0.0, // NONE
	},
	"ACE_HEARTS": {
		"higher":         1.04,
		"lower":          0.0, // NONE
		"same":           16.69,
		"higher_or_same": 0.0, // NONE
		"lower_or_same":  0.0, // NONE
	},
	"ACE_DIAMONDS": {
		"higher":         1.04,
		"lower":          0.0, // NONE
		"same":           16.69,
		"higher_or_same": 0.0, // NONE
		"lower_or_same":  0.0, // NONE
	},
	"ACE_CLUBS": {
		"higher":         1.04,
		"lower":          0.0, // NONE
		"same":           16.69,
		"higher_or_same": 0.0, // NONE
		"lower_or_same":  0.0, // NONE
	},

	// TWO cards
	"TWO_SPADES": {
		"higher":         1.07,
		"lower":          7.15,
		"same":           16.69,
		"higher_or_same": 1.07,
		"lower_or_same":  7.15,
	},
	"TWO_HEARTS": {
		"higher":         1.07,
		"lower":          7.15,
		"same":           16.69,
		"higher_or_same": 1.07,
		"lower_or_same":  7.15,
	},
	"TWO_DIAMONDS": {
		"higher":         1.07,
		"lower":          7.15,
		"same":           16.69,
		"higher_or_same": 1.07,
		"lower_or_same":  7.15,
	},
	"TWO_CLUBS": {
		"higher":         1.07,
		"lower":          7.15,
		"same":           16.69,
		"higher_or_same": 1.07,
		"lower_or_same":  7.15,
	},

	// THREE cards
	"THREE_SPADES": {
		"higher":         1.10,
		"lower":          6.69,
		"same":           16.69,
		"higher_or_same": 1.10,
		"lower_or_same":  6.69,
	},
	"THREE_HEARTS": {
		"higher":         1.10,
		"lower":          6.69,
		"same":           16.69,
		"higher_or_same": 1.10,
		"lower_or_same":  6.69,
	},
	"THREE_DIAMONDS": {
		"higher":         1.10,
		"lower":          6.69,
		"same":           16.69,
		"higher_or_same": 1.10,
		"lower_or_same":  6.69,
	},
	"THREE_CLUBS": {
		"higher":         1.10,
		"lower":          6.69,
		"same":           16.69,
		"higher_or_same": 1.10,
		"lower_or_same":  6.69,
	},

	// FOUR cards
	"FOUR_SPADES": {
		"higher":         1.13,
		"lower":          6.23,
		"same":           16.69,
		"higher_or_same": 1.13,
		"lower_or_same":  6.23,
	},
	"FOUR_HEARTS": {
		"higher":         1.13,
		"lower":          6.23,
		"same":           16.69,
		"higher_or_same": 1.13,
		"lower_or_same":  6.23,
	},
	"FOUR_DIAMONDS": {
		"higher":         1.13,
		"lower":          6.23,
		"same":           16.69,
		"higher_or_same": 1.13,
		"lower_or_same":  6.23,
	},
	"FOUR_CLUBS": {
		"higher":         1.13,
		"lower":          6.23,
		"same":           16.69,
		"higher_or_same": 1.13,
		"lower_or_same":  6.23,
	},

	// FIVE cards
	"FIVE_SPADES": {
		"higher":         1.17,
		"lower":          5.77,
		"same":           16.69,
		"higher_or_same": 1.17,
		"lower_or_same":  5.77,
	},
	"FIVE_HEARTS": {
		"higher":         1.17,
		"lower":          5.77,
		"same":           16.69,
		"higher_or_same": 1.17,
		"lower_or_same":  5.77,
	},
	"FIVE_DIAMONDS": {
		"higher":         1.17,
		"lower":          5.77,
		"same":           16.69,
		"higher_or_same": 1.17,
		"lower_or_same":  5.77,
	},
	"FIVE_CLUBS": {
		"higher":         1.17,
		"lower":          5.77,
		"same":           16.69,
		"higher_or_same": 1.17,
		"lower_or_same":  5.77,
	},

	// SIX cards
	"SIX_SPADES": {
		"higher":         1.21,
		"lower":          5.31,
		"same":           16.69,
		"higher_or_same": 1.21,
		"lower_or_same":  5.31,
	},
	"SIX_HEARTS": {
		"higher":         1.21,
		"lower":          5.31,
		"same":           16.69,
		"higher_or_same": 1.21,
		"lower_or_same":  5.31,
	},
	"SIX_DIAMONDS": {
		"higher":         1.21,
		"lower":          5.31,
		"same":           16.69,
		"higher_or_same": 1.21,
		"lower_or_same":  5.31,
	},
	"SIX_CLUBS": {
		"higher":         1.21,
		"lower":          5.31,
		"same":           16.69,
		"higher_or_same": 1.21,
		"lower_or_same":  5.31,
	},

	// SEVEN cards
	"SEVEN_SPADES": {
		"higher":         1.26,
		"lower":          4.85,
		"same":           16.69,
		"higher_or_same": 1.26,
		"lower_or_same":  4.85,
	},
	"SEVEN_HEARTS": {
		"higher":         1.26,
		"lower":          4.85,
		"same":           16.69,
		"higher_or_same": 1.26,
		"lower_or_same":  4.85,
	},
	"SEVEN_DIAMONDS": {
		"higher":         1.26,
		"lower":          4.85,
		"same":           16.69,
		"higher_or_same": 1.26,
		"lower_or_same":  4.85,
	},
	"SEVEN_CLUBS": {
		"higher":         1.26,
		"lower":          4.85,
		"same":           16.69,
		"higher_or_same": 1.26,
		"lower_or_same":  4.85,
	},

	// EIGHT cards
	"EIGHT_SPADES": {
		"higher":         1.32,
		"lower":          4.39,
		"same":           16.69,
		"higher_or_same": 1.32,
		"lower_or_same":  4.39,
	},
	"EIGHT_HEARTS": {
		"higher":         1.32,
		"lower":          4.39,
		"same":           16.69,
		"higher_or_same": 1.32,
		"lower_or_same":  4.39,
	},
	"EIGHT_DIAMONDS": {
		"higher":         1.32,
		"lower":          4.39,
		"same":           16.69,
		"higher_or_same": 1.32,
		"lower_or_same":  4.39,
	},
	"EIGHT_CLUBS": {
		"higher":         1.32,
		"lower":          4.39,
		"same":           16.69,
		"higher_or_same": 1.32,
		"lower_or_same":  4.39,
	},

	// NINE cards
	"NINE_SPADES": {
		"higher":         2.79,
		"lower":          1.43,
		"same":           16.69,
		"higher_or_same": 2.79,
		"lower_or_same":  1.43,
	},
	"NINE_HEARTS": {
		"higher":         2.79,
		"lower":          1.43,
		"same":           16.69,
		"higher_or_same": 2.79,
		"lower_or_same":  1.43,
	},
	"NINE_DIAMONDS": {
		"higher":         2.79,
		"lower":          1.43,
		"same":           16.69,
		"higher_or_same": 2.79,
		"lower_or_same":  1.43,
	},
	"NINE_CLUBS": {
		"higher":         2.79,
		"lower":          1.43,
		"same":           16.69,
		"higher_or_same": 2.79,
		"lower_or_same":  1.43,
	},

	// TEN cards
	"TEN_SPADES": {
		"higher":         3.34,
		"lower":          1.28,
		"same":           16.69,
		"higher_or_same": 3.34,
		"lower_or_same":  1.28,
	},
	"TEN_HEARTS": {
		"higher":         3.34,
		"lower":          1.28,
		"same":           16.69,
		"higher_or_same": 3.34,
		"lower_or_same":  1.28,
	},
	"TEN_DIAMONDS": {
		"higher":         3.34,
		"lower":          1.28,
		"same":           16.69,
		"higher_or_same": 3.34,
		"lower_or_same":  1.28,
	},
	"TEN_CLUBS": {
		"higher":         3.34,
		"lower":          1.28,
		"same":           16.69,
		"higher_or_same": 3.34,
		"lower_or_same":  1.28,
	},

	// JACK cards
	"JACK_SPADES": {
		"higher":         4.55,
		"lower":          1.16,
		"same":           16.69,
		"higher_or_same": 4.55,
		"lower_or_same":  1.16,
	},
	"JACK_HEARTS": {
		"higher":         4.55,
		"lower":          1.16,
		"same":           16.69,
		"higher_or_same": 4.55,
		"lower_or_same":  1.16,
	},
	"JACK_DIAMONDS": {
		"higher":         4.55,
		"lower":          1.16,
		"same":           16.69,
		"higher_or_same": 4.55,
		"lower_or_same":  1.16,
	},
	"JACK_CLUBS": {
		"higher":         4.55,
		"lower":          1.16,
		"same":           16.69,
		"higher_or_same": 4.55,
		"lower_or_same":  1.16,
	},

	// QUEEN cards
	"QUEEN_SPADES": {
		"higher":         7.15,
		"lower":          1.07,
		"same":           16.69,
		"higher_or_same": 7.15,
		"lower_or_same":  1.07,
	},
	"QUEEN_HEARTS": {
		"higher":         7.15,
		"lower":          1.07,
		"same":           16.69,
		"higher_or_same": 7.15,
		"lower_or_same":  1.07,
	},
	"QUEEN_DIAMONDS": {
		"higher":         7.15,
		"lower":          1.07,
		"same":           16.69,
		"higher_or_same": 7.15,
		"lower_or_same":  1.07,
	},
	"QUEEN_CLUBS": {
		"higher":         7.15,
		"lower":          1.07,
		"same":           16.69,
		"higher_or_same": 7.15,
		"lower_or_same":  1.07,
	},

	// KING cards
	"KING_SPADES": {
		"higher":         0.0, // NONE
		"lower":          1.04,
		"same":           16.69,
		"higher_or_same": 0.0, // NONE
		"lower_or_same":  0.0, // NONE
	},
	"KING_HEARTS": {
		"higher":         0.0, // NONE
		"lower":          1.04,
		"same":           16.69,
		"higher_or_same": 0.0, // NONE
		"lower_or_same":  0.0, // NONE
	},
	"KING_DIAMONDS": {
		"higher":         0.0, // NONE
		"lower":          1.04,
		"same":           16.69,
		"higher_or_same": 0.0, // NONE
		"lower_or_same":  0.0, // NONE
	},
	"KING_CLUBS": {
		"higher":         0.0, // NONE
		"lower":          1.04,
		"same":           16.69,
		"higher_or_same": 0.0, // NONE
		"lower_or_same":  0.0, // NONE
	},
}

// getCardRank extracts the rank from a card string (e.g., "SEVEN_SPADES" -> "SEVEN_SPADES")
func getCardRank(card string) string {
	if len(card) == 0 {
		return ""
	}

	// For NINE_CLUBS format, return the card as-is
	return card
}

// getBaseMultipliers returns the base multipliers for a given card
func getBaseMultipliers(card string) map[string]float64 {
	rank := getCardRank(card)
	if multipliers, exists := baseMultipliers[rank]; exists {
		return multipliers
	}
	// Fallback to default multipliers if card not found
	return map[string]float64{
		"higher":         1.0,
		"lower":          1.0,
		"same":           16.69,
		"higher_or_same": 1.0,
		"lower_or_same":  1.0,
	}
}

// getJDBMultipliers returns multipliers using JDB formula
func getJDBMultipliers(card string, multiplierModifier float64, previousWinningMultiplier float64) map[string]float64 {
	baseMultipliers := getBaseMultipliers(card)
	jdbMultipliers := make(map[string]float64)

	// JDB Formula:
	// Bonus Streak = 1 + (Current Card Base Multiplier × 0.00588)
	// New Multiplier = Previous Winning Multiplier × Current Card Base Multiplier × Bonus Streak

	for betType, baseMultiplier := range baseMultipliers {
		if baseMultiplier > 0 { // Only apply to enabled bets
			// Calculate bonus streak
			bonusStreak := 1.0 + (baseMultiplier * 0.00588)

			// Calculate new multiplier using JDB formula
			newMultiplier := previousWinningMultiplier * baseMultiplier * bonusStreak

			jdbMultipliers[betType] = roundToTwo(newMultiplier)
		} else {
			jdbMultipliers[betType] = 0.0 // Keep disabled bets disabled
		}
	}

	return jdbMultipliers
}

const (
	HouseEdge = 0.95
)

// GetHiloOptions generates all betting options for the current card
// Uses JDB formula for progressive multipliers
// GetBaseHiloOptions returns base multipliers only (for start, preview, preview-skip)
func GetBaseHiloOptions(currentCard string) []BetOption {
	// Get base multipliers only (no JDB formula)
	baseMultipliers := getBaseMultipliers(currentCard)

	options := []BetOption{}

	// HIGHER button
	options = append(options, BetOption{
		ID:         "higher",
		Name:       "HIGH",
		Multiplier: roundToTwo(baseMultipliers["higher"]),
		IsEnabled:  baseMultipliers["higher"] > 0,
	})

	// LOWER button
	options = append(options, BetOption{
		ID:         "lower",
		Name:       "LOW",
		Multiplier: roundToTwo(baseMultipliers["lower"]),
		IsEnabled:  baseMultipliers["lower"] > 0,
	})

	// SAME button (always available)
	options = append(options, BetOption{
		ID:         "same",
		Name:       "SAME",
		Multiplier: roundToTwo(baseMultipliers["same"]),
		IsEnabled:  true,
	})

	// HIGHER OR SAME button
	options = append(options, BetOption{
		ID:         "higher_or_same",
		Name:       "HIGH OR SAME",
		Multiplier: roundToTwo(baseMultipliers["higher_or_same"]),
		IsEnabled:  baseMultipliers["higher_or_same"] > 0,
	})

	// LOWER OR SAME button
	options = append(options, BetOption{
		ID:         "lower_or_same",
		Name:       "LOW OR SAME",
		Multiplier: roundToTwo(baseMultipliers["lower_or_same"]),
		IsEnabled:  baseMultipliers["lower_or_same"] > 0,
	})

	return options
}

// GetHiloOptions returns JDB progressive multipliers (for guess, skip during game)
func GetHiloOptions(currentCard string, position int, multiplierModifier float64, previousWinningMultiplier float64) []BetOption {
	// Get JDB multipliers using the correct formula
	jdbMultipliers := getJDBMultipliers(currentCard, multiplierModifier, previousWinningMultiplier)

	options := []BetOption{}

	// HIGHER button
	options = append(options, BetOption{
		ID:         "higher",
		Name:       "HIGH",
		Multiplier: roundToTwo(jdbMultipliers["higher"]),
		IsEnabled:  jdbMultipliers["higher"] > 0,
	})

	// LOWER button
	options = append(options, BetOption{
		ID:         "lower",
		Name:       "LOW",
		Multiplier: roundToTwo(jdbMultipliers["lower"]),
		IsEnabled:  jdbMultipliers["lower"] > 0,
	})

	// SAME button (always available)
	options = append(options, BetOption{
		ID:         "same",
		Name:       "SAME",
		Multiplier: roundToTwo(jdbMultipliers["same"]),
		IsEnabled:  true,
	})

	// HIGHER OR SAME button
	options = append(options, BetOption{
		ID:         "higher_or_same",
		Name:       "HIGH OR SAME",
		Multiplier: roundToTwo(jdbMultipliers["higher_or_same"]),
		IsEnabled:  jdbMultipliers["higher_or_same"] > 0,
	})

	// LOWER OR SAME button
	options = append(options, BetOption{
		ID:         "lower_or_same",
		Name:       "LOW OR SAME",
		Multiplier: roundToTwo(jdbMultipliers["lower_or_same"]),
		IsEnabled:  jdbMultipliers["lower_or_same"] > 0,
	})

	return options
}

// GetBaseMultiplierForBet returns the base multiplier for a specific bet choice (no JDB formula)
func GetBaseMultiplierForBet(currentCard string, betChoice string) float64 {
	baseMultipliers := getBaseMultipliers(currentCard)
	return roundToTwo(baseMultipliers[betChoice])
}

// GetMultiplierForBet returns the multiplier for a specific bet choice
// Uses JDB formula for progressive multipliers
func GetMultiplierForBet(currentCard string, betChoice string, multiplierModifier float64, previousWinningMultiplier float64) float64 {
	jdbMultipliers := getJDBMultipliers(currentCard, multiplierModifier, previousWinningMultiplier)
	return roundToTwo(jdbMultipliers[betChoice])
}

// IsValidBetChoice validates if a bet choice is valid for the current card
func IsValidBetChoice(currentValue int, betChoice string) bool {
	switch betChoice {
	case "higher":
		return currentValue < 13
	case "lower":
		return currentValue > 1
	case "same", "higher_or_same", "lower_or_same":
		return true
	default:
		return false
	}
}

// CheckBetResult determines if a bet was correct
func CheckBetResult(currentValue, nextValue int, betChoice string) bool {
	switch betChoice {
	case "higher":
		return nextValue > currentValue
	case "lower":
		return nextValue < currentValue
	case "same":
		return nextValue == currentValue
	case "higher_or_same":
		return nextValue >= currentValue
	case "lower_or_same":
		return nextValue <= currentValue
	default:
		return false
	}
}

// FindLosingCard finds a card that would make the bet lose (for RNG forced losses)
func FindLosingCard(currentValue int, betChoice string, remainingCards []string) string {
	for _, card := range remainingCards {
		cardValue := GetCardValue(card)
		if !CheckBetResult(currentValue, cardValue, betChoice) {
			return card
		}
	}
	return remainingCards[0] // Fallback
}

// FindWinningCard finds a card that would make the bet win (for RNG forced wins)
func FindWinningCard(currentValue int, betChoice string, remainingCards []string) string {
	for _, card := range remainingCards {
		cardValue := GetCardValue(card)
		if CheckBetResult(currentValue, cardValue, betChoice) {
			return card
		}
	}
	return "" // No winning card available
}

func roundToTwo(val float64) float64 {
	return math.Round(val*100) / 100
}
