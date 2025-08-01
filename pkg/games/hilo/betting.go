package hilo

import "math"

const (
	HouseEdge      = 0.95
	RemainingCards = 51.0
)

// GetHiloOptions generates all betting options for the current card
func GetHiloOptions(currentCard string) []BetOption {
	currentValue := GetCardValue(currentCard)
	options := []BetOption{}

	// HIGHER button
	if currentValue < 13 {
		higherCards := float64((13 - currentValue) * 4)
		probability := higherCards / RemainingCards
		multiplier := (1.0 / probability) * HouseEdge

		options = append(options, BetOption{
			ID:         "higher",
			Name:       "HIGH",
			Multiplier: roundToTwo(multiplier),
			IsEnabled:  true,
		})
	} else {
		options = append(options, BetOption{
			ID:         "higher",
			Name:       "HIGH",
			Multiplier: 0,
			IsEnabled:  false,
		})
	}

	// LOWER button
	if currentValue > 1 {
		lowerCards := float64((currentValue - 1) * 4)
		probability := lowerCards / RemainingCards
		multiplier := (1.0 / probability) * HouseEdge

		options = append(options, BetOption{
			ID:         "lower",
			Name:       "LOW",
			Multiplier: roundToTwo(multiplier),
			IsEnabled:  true,
		})
	} else {
		options = append(options, BetOption{
			ID:         "lower",
			Name:       "LOW",
			Multiplier: 0,
			IsEnabled:  false,
		})
	}

	// SAME button (always available)
	sameCards := 3.0
	probability := sameCards / RemainingCards
	multiplier := (1.0 / probability) * HouseEdge

	options = append(options, BetOption{
		ID:         "same",
		Name:       "SAME",
		Multiplier: roundToTwo(multiplier),
		IsEnabled:  true,
	})

	// HIGHER OR SAME button
	higherOrSameCards := 3.0 // Same cards
	if currentValue < 13 {
		higherOrSameCards += float64((13 - currentValue) * 4)
	}
	probability = higherOrSameCards / RemainingCards
	multiplier = (1.0 / probability) * HouseEdge

	options = append(options, BetOption{
		ID:         "higher_or_same",
		Name:       "HIGH OR SAME",
		Multiplier: roundToTwo(multiplier),
		IsEnabled:  true,
	})

	// LOWER OR SAME button
	lowerOrSameCards := 3.0 // Same cards
	if currentValue > 1 {
		lowerOrSameCards += float64((currentValue - 1) * 4)
	}
	probability = lowerOrSameCards / RemainingCards
	multiplier = (1.0 / probability) * HouseEdge

	options = append(options, BetOption{
		ID:         "lower_or_same",
		Name:       "LOW OR SAME",
		Multiplier: roundToTwo(multiplier),
		IsEnabled:  true,
	})

	return options
}

// GetMultiplierForBet returns the multiplier for a specific bet choice
func GetMultiplierForBet(currentValue int, betChoice string) float64 {
	switch betChoice {
	case "higher":
		if currentValue >= 13 {
			return 0
		}
		higherCards := float64((13 - currentValue) * 4)
		probability := higherCards / RemainingCards
		return roundToTwo((1.0 / probability) * HouseEdge)

	case "lower":
		if currentValue <= 1 {
			return 0
		}
		lowerCards := float64((currentValue - 1) * 4)
		probability := lowerCards / RemainingCards
		return roundToTwo((1.0 / probability) * HouseEdge)

	case "same":
		probability := 3.0 / RemainingCards
		return roundToTwo((1.0 / probability) * HouseEdge)

	case "higher_or_same":
		higherOrSameCards := 3.0 // Same cards
		if currentValue < 13 {
			higherOrSameCards += float64((13 - currentValue) * 4)
		}
		probability := higherOrSameCards / RemainingCards
		return roundToTwo((1.0 / probability) * HouseEdge)

	case "lower_or_same":
		lowerOrSameCards := 3.0 // Same cards
		if currentValue > 1 {
			lowerOrSameCards += float64((currentValue - 1) * 4)
		}
		probability := lowerOrSameCards / RemainingCards
		return roundToTwo((1.0 / probability) * HouseEdge)

	default:
		return 0
	}
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