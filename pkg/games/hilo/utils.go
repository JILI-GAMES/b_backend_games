package hilo

import (
	"fmt"
	"math"
)

// FormatCurrency formats a float as currency with 2 decimal places
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// RoundToTwo rounds a float to 2 decimal places
func RoundToTwo(amount float64) float64 {
	return math.Round(amount*100) / 100
}

// GetCardDisplayName returns a user-friendly card name
func GetCardDisplayName(card string) string {
	return card
}

// GetProbabilityForBet returns the win probability for a bet choice
func GetProbabilityForBet(currentValue int, betChoice string, position int) float64 {
	remainingCards := 52.0 - float64(position) - 1.0

	switch betChoice {
	case "higher":
		if currentValue >= 13 {
			return 0
		}
		higherCards := float64((13 - currentValue) * 4)
		return higherCards / remainingCards

	case "lower":
		if currentValue <= 1 {
			return 0
		}
		lowerCards := float64((currentValue - 1) * 4)
		return lowerCards / remainingCards

	case "same":
		return 3.0 / remainingCards

	case "higher_or_same":
		higherOrSameCards := 3.0 // Same cards
		if currentValue < 13 {
			higherOrSameCards += float64((13 - currentValue) * 4)
		}
		return higherOrSameCards / remainingCards

	case "lower_or_same":
		lowerOrSameCards := 3.0 // Same cards
		if currentValue > 1 {
			lowerOrSameCards += float64((currentValue - 1) * 4)
		}
		return lowerOrSameCards / remainingCards

	default:
		return 0
	}
}

// GetProbabilityText returns human-readable probability text
func GetProbabilityText(currentValue int, betChoice string, position int) string {
	probability := GetProbabilityForBet(currentValue, betChoice, position)
	if probability == 0 {
		return "Impossible"
	}
	return fmt.Sprintf("%.1f%% chance", probability*100)
}

// GetPayoutText returns human-readable payout text
func GetPayoutText(multiplier float64) string {
	if multiplier <= 0 {
		return "No payout"
	}
	return fmt.Sprintf("%.2fx multiplier", multiplier)
}

// IsGameWinnable checks if the game can still be won
func IsGameWinnable(gameState GameState) bool {
	if gameState.IsGameOver {
		return false
	}

	// Check if we have cards remaining
	deck := GenerateDeck(gameState.Seed)
	return gameState.Position < len(deck)-1
}

// GetGameProgress returns the progress through the deck
func GetGameProgress(gameState GameState) (int, int, float64) {
	deck := GenerateDeck(gameState.Seed)
	totalCards := len(deck)
	currentPosition := gameState.Position
	progressPercent := float64(currentPosition) / float64(totalCards) * 100

	return currentPosition, totalCards, progressPercent
}

// GetRiskLevel returns a risk assessment for the current win amount
func GetRiskLevel(winAmount float64) string {
	if winAmount <= 5 {
		return "Low Risk"
	} else if winAmount <= 20 {
		return "Medium Risk"
	} else if winAmount <= 50 {
		return "High Risk"
	} else {
		return "Very High Risk"
	}
}

// GetCardString helper function to convert value back to card string for testing
func GetCardString(value int) string {
	if value < 1 || value > 13 {
		return ""
	}
	return ranks[value-1] + "♠" // Default to spades for value conversion
}

// getHighestMultiplier returns the highest multiplier from a slice of BetOptions
func getHighestMultiplier(options []BetOption) float64 {
	highest := 0.0
	for _, option := range options {
		if option.Multiplier > highest {
			highest = option.Multiplier
		}
	}
	return highest
}
