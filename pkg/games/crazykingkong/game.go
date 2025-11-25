package crazykingkong

import (
	"math/rand"
)

// Boulder multiplier ranges
var BoulderMultipliers = map[BoulderType]struct {
	Min float64
	Max float64
}{
	BoulderGold:  {Min: 1.5, Max: 100.0},
	BoulderBlue:  {Min: 1.4, Max: 50.0},
	BoulderRed:   {Min: 1.3, Max: 20.0},
	BoulderWhite: {Min: 1.2, Max: 10.0},
}

// Stone multiplier ranges
var StoneMultipliers = map[StoneType]struct {
	Min float64
	Max float64
}{
	StoneGold:   {Min: 28.0, Max: 888.0},
	StoneSilver: {Min: 18.0, Max: 18.0}, // Fixed multiplier
	StoneBronze: {Min: 8.0, Max: 8.0},   // Fixed multiplier
}

// Valid bet amounts
var ValidBetAmounts = []float64{10,15,20,250}

// GenerateBoulderMultiplier generates a random multiplier for a given boulder type
func GenerateBoulderMultiplier(boulderType BoulderType) float64 {
	multiplierRange := BoulderMultipliers[boulderType]

	// Generate weighted random multiplier (lower values more likely)
	// Using exponential distribution to favor lower multipliers
	randomValue := rand.Float64()

	// Apply exponential curve to favor lower multipliers
	exponentialValue := 1.0 - (randomValue * randomValue * randomValue)

	multiplier := multiplierRange.Min + (exponentialValue * (multiplierRange.Max - multiplierRange.Min))

	// Round to 1 decimal place
	return float64(int(multiplier*10)) / 10
}

// GenerateStoneMultiplier generates a multiplier for a given stone type
func GenerateStoneMultiplier(stoneType StoneType) float64 {
	multiplierRange := StoneMultipliers[stoneType]

	if stoneType == StoneSilver || stoneType == StoneBronze {
		return multiplierRange.Min // Fixed multipliers
	}

	// For gold stone, use weighted random
	randomValue := rand.Float64()
	exponentialValue := 1.0 - (randomValue * randomValue)

	multiplier := multiplierRange.Min + (exponentialValue * (multiplierRange.Max - multiplierRange.Min))

	// Round to nearest integer for bonus multipliers
	return float64(int(multiplier))
}

// ShouldTriggerBonus determines if bonus game should be triggered
// Only triggers when boulder is broken (RNG says win)
func ShouldTriggerBonus() bool {
	// 10% chance to trigger bonus game when boulder breaks
	return rand.Float64() < 0.10
}

// GenerateAvailableStones creates 3 random stones for bonus selection
func GenerateAvailableStones() []Stone {
	stoneTypes := []StoneType{StoneGold, StoneSilver, StoneBronze}
	stones := make([]Stone, 3)

	// Shuffle and pick 3 (could include duplicates for variety)
	for i := 0; i < 3; i++ {
		randomType := stoneTypes[rand.Intn(len(stoneTypes))]
		stones[i] = Stone{
			Type:        randomType,
			DisplayName: string(randomType) + " stone",
		}
	}

	return stones
}

// ValidateBetAmount checks if the bet amount is valid
func ValidateBetAmount(betAmount float64) bool {
	for _, validAmount := range ValidBetAmounts {
		if betAmount == validAmount {
			return true
		}
	}
	return false
}

// ValidateBoulderType checks if the boulder type is valid
func ValidateBoulderType(boulderType BoulderType) bool {
	_, exists := BoulderMultipliers[boulderType]
	return exists
}

// ValidateStoneType checks if the stone type is valid
func ValidateStoneType(stoneType StoneType) bool {
	_, exists := StoneMultipliers[stoneType]
	return exists
}
