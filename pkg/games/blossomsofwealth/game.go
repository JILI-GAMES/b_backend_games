package blossomsofwealth

import (
    "fmt"
    "log"
    "math"
    "math/rand"
    "strings"
    "sync"
    "time"
)

const (
    Denomination = 0.01
    Reels       = 5
    Rows        = 3
    MinBet      = 20 // Minimum bet multiplier
)

// Symbol weights for random generation
var SymbolWeights = map[Symbol]float64{
    SymbolGoldTreasure:   0.08,
    SymbolPurpleTreasure: 0.08,
    SymbolBlueTreasure:   0.08,
    SymbolCompass:        0.08,
    SymbolA:              0.10,
    SymbolK:              0.10,
    SymbolQ:              0.10,
    SymbolJ:              0.10,
    SymbolWild:           0.06,
    SymbolSilverFlower:   0.02,
    SymbolGoldFlower:     0.02,
}

// Paytable (payouts for Bet Multiplier = 1)
var Paytable = map[Symbol]map[int]float64{
    SymbolGoldTreasure:   {3: 75, 4: 150, 5: 400},
    SymbolPurpleTreasure: {3: 50, 4: 150, 5: 300},
    SymbolBlueTreasure:   {3: 40, 4: 100, 5: 250},
    SymbolCompass:        {3: 30, 4: 100, 5: 200},
    SymbolA:              {3: 15, 4: 25, 5: 80},
    SymbolK:              {3: 15, 4: 25, 5: 80},
    SymbolQ:              {3: 10, 4: 15, 5: 60},
    SymbolJ:              {3: 10, 4: 15, 5: 60},
}

// BetAmountToMultiplier maps bet amounts to multipliers
var BetAmountToMultiplier = map[float64]int{
    10:  1,
    15:  2,
    20:  3,
    250:  5,
}

// Range for Gold Flower bonus multipliers
const (
    MinGoldFlowerMultiplier = 10
    MaxGoldFlowerMultiplier = 25
)

// Range for Silver Flower free spins
const (
    MinSilverFlowerSpins = 3
    MaxSilverFlowerSpins = 10
    MaxTotalFreeSpins    = 100
)

// WaysToWin generates all possible 243 ways to win (3^5)
var WaysToWin = generateWaysToWin()

func generateWaysToWin() [][]int {
    ways := make([][]int, 0, 243)
    for r0 := 0; r0 < 3; r0++ {
        for r1 := 0; r1 < 3; r1++ {
            for r2 := 0; r2 < 3; r2++ {
                for r3 := 0; r3 < 3; r3++ {
                    for r4 := 0; r4 < 3; r4++ {
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
    return SymbolA // Fallback
}

// WeightedRandomSymbolExcluding selects a symbol based on weights, excluding specified symbols
func WeightedRandomSymbolExcluding(r *rand.Rand, excludeSymbols []Symbol) Symbol {
    // Create a temporary weight map excluding the specified symbols
    tempWeights := make(map[Symbol]float64)
    for symbol, weight := range SymbolWeights {
        excluded := false
        for _, excludeSymbol := range excludeSymbols {
            if symbol == excludeSymbol {
                excluded = true
                break
            }
        }
        if !excluded {
            tempWeights[symbol] = weight
        }
    }
    
    totalWeight := 0.0
    for _, weight := range tempWeights {
        totalWeight += weight
    }

    roll := r.Float64() * totalWeight
    currentWeight := 0.0
    for symbol, weight := range tempWeights {
        currentWeight += weight
        if roll <= currentWeight {
            return symbol
        }
    }
    return SymbolA // Fallback
}

// Updated GenerateReelsWithWin function
func GenerateReelsWithWin() [][]string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    var reels [][]string

    // Keep generating until a win is found
    for {
        reels = make([][]string, Reels)
        for reel := 0; reel < Reels; reel++ {
            reels[reel] = make([]string, Rows)
            for row := 0; row < Rows; row++ {
                var symbol Symbol
                
                if reel == 0 {
                    // First reel: exclude Wild, SilverFlower, and GoldFlower
                    symbol = WeightedRandomSymbolExcluding(r, []Symbol{SymbolWild, SymbolSilverFlower, SymbolGoldFlower})
                } else if reel <= 2 {
                    // Reels 1-3: exclude GoldFlower only
                    symbol = WeightedRandomSymbolExcluding(r, []Symbol{SymbolGoldFlower})
                } else {
                    // Reels 4-5: exclude SilverFlower only
                    symbol = WeightedRandomSymbolExcluding(r, []Symbol{SymbolSilverFlower})
                }
                
                reels[reel][row] = string(symbol)
            }
        }

        // Check for a win (base game, so isFreeSpin=false)
        totalWinnings, _ := CalculateWins(reels, 1, 1, false)
        if totalWinnings > 0 {
            log.Printf("Generated reels with win: %v", reels)
            break
        }
    }

    return reels
}

// // GenerateReelsWithWin generates a 5x3 grid with a guaranteed win
// func GenerateReelsWithWin() [][]string {
//     r := rand.New(rand.NewSource(time.Now().UnixNano()))
//     var reels [][]string

//     // Keep generating until a win is found
//     for {
//         reels = make([][]string, Reels)
//         for reel := 0; reel < Reels; reel++ {
//             reels[reel] = make([]string, Rows)
//             for row := 0; row < Rows; row++ {
//                 symbol := WeightedRandomSymbol(r)
                
//                 // Wilds only on reels 2-5 (indices 1-4)
//                 if reel == 0 && symbol == SymbolWild {
//                     // symbol = WeightedRandomSymbol(r)
//                     for symbol == SymbolWild {
//                         symbol = WeightedRandomSymbol(r)
//                     }
//                 }
                
//                 // Silver Flowers only on reels 1-3 (indices 0-2)
//                 if reel > 2 && symbol == SymbolSilverFlower {
//                     symbol = WeightedRandomSymbol(r)
//                     for symbol == SymbolSilverFlower {
//                         symbol = WeightedRandomSymbol(r)
//                     }
//                 }
                
//                 // Gold Flowers only on reels 4-5 (indices 3-4)
//                 if reel < 3 && symbol == SymbolGoldFlower {
//                     symbol = WeightedRandomSymbol(r)
//                     for symbol == SymbolGoldFlower {
//                         symbol = WeightedRandomSymbol(r)
//                     }
//                 }
                
//                 reels[reel][row] = string(symbol)
//             }
//         }

//         // Check for a win (base game, so isFreeSpin=false)
//         totalWinnings, _ := CalculateWins(reels, 1, 1, false)
//         if totalWinnings > 0 {
//             log.Printf("Generated reels with win: %v", reels)
//             break
//         }
//     }

//     return reels
// }

// Updated GenerateLossReels function
func GenerateLossReels() [][]string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    var reels [][]string

    // Keep generating until no wins are found
    for {
        reels = make([][]string, Reels)
        for reel := 0; reel < Reels; reel++ {
            reels[reel] = make([]string, Rows)
            for row := 0; row < Rows; row++ {
                var symbol Symbol
                
                if reel == 0 {
                    // First reel: exclude Wild, SilverFlower, and GoldFlower
                    symbol = WeightedRandomSymbolExcluding(r, []Symbol{SymbolWild, SymbolSilverFlower, SymbolGoldFlower})
                } else if reel <= 2 {
                    // Reels 1-3: exclude GoldFlower only
                    symbol = WeightedRandomSymbolExcluding(r, []Symbol{SymbolGoldFlower})
                } else {
                    // Reels 4-5: exclude SilverFlower only
                    symbol = WeightedRandomSymbolExcluding(r, []Symbol{SymbolSilverFlower})
                }
                
                // Additional logic to avoid consecutive matching symbols for loss generation
                if reel > 0 {
                    previousSymbol := reels[reel-1][row]
                    attempts := 0
                    for string(symbol) == previousSymbol && symbol != SymbolSilverFlower && symbol != SymbolGoldFlower && attempts < 10 {
                        if reel == 0 {
                            symbol = WeightedRandomSymbolExcluding(r, []Symbol{SymbolWild, SymbolSilverFlower, SymbolGoldFlower})
                        } else if reel <= 2 {
                            symbol = WeightedRandomSymbolExcluding(r, []Symbol{SymbolGoldFlower})
                        } else {
                            symbol = WeightedRandomSymbolExcluding(r, []Symbol{SymbolSilverFlower})
                        }
                        attempts++
                    }
                }
                
                reels[reel][row] = string(symbol)
            }
        }

        // Check for no wins (base game, so isFreeSpin=false)
        totalWinnings, _ := CalculateWins(reels, 1, 1, false)
        if totalWinnings == 0 {
            log.Printf("Generated reels with no wins: %v", reels)
            break
        }
    }

    return reels
}

// // GenerateLossReels generates a 5x3 grid with no wins but allows Free Spin Bonus triggers
// func GenerateLossReels() [][]string {
//     r := rand.New(rand.NewSource(time.Now().UnixNano()))
//     var reels [][]string

//     // Keep generating until no wins are found
//     for {
//         reels = make([][]string, Reels)
//         for reel := 0; reel < Reels; reel++ {
//             reels[reel] = make([]string, Rows)
//             for row := 0; row < Rows; row++ {
//                 symbol := WeightedRandomSymbol(r)
                
//                 // Wilds only on reels 2-5 (indices 1-4)
//                 if reel == 0 && symbol == SymbolWild {
//                     // symbol = WeightedRandomSymbol(r)
//                     for symbol == SymbolWild {
//                         symbol = WeightedRandomSymbol(r)
//                     }
//                 }
                
//                 // Silver Flowers only on reels 1-3 (indices 0-2)
//                 if reel > 2 && symbol == SymbolSilverFlower {
//                     symbol = WeightedRandomSymbol(r)
//                     for symbol == SymbolSilverFlower {
//                         symbol = WeightedRandomSymbol(r)
//                     }
//                 }
                
//                 // Gold Flowers only on reels 4-5 (indices 3-4)
//                 if reel < 3 && symbol == SymbolGoldFlower {
//                     symbol = WeightedRandomSymbol(r)
//                     for symbol == SymbolGoldFlower {
//                         symbol = WeightedRandomSymbol(r)
//                     }
//                 }
                
//                 // Avoid matching symbols to prevent wins
//                 if reel > 0 {
//                     previousSymbol := reels[reel-1][row]
//                     for string(symbol) == previousSymbol && symbol != SymbolSilverFlower && symbol != SymbolGoldFlower {
//                         symbol = WeightedRandomSymbol(r)
//                         if reel == 0 && symbol == SymbolWild {
//                             symbol = WeightedRandomSymbol(r)
//                             for symbol == SymbolWild {
//                                 symbol = WeightedRandomSymbol(r)
//                             }
//                         }
                        
//                         if reel > 2 && symbol == SymbolSilverFlower {
//                             symbol = WeightedRandomSymbol(r)
//                             for symbol == SymbolSilverFlower {
//                                 symbol = WeightedRandomSymbol(r)
//                             }
//                         }
                        
//                         if reel < 3 && symbol == SymbolGoldFlower {
//                             symbol = WeightedRandomSymbol(r)
//                             for symbol == SymbolGoldFlower {
//                                 symbol = WeightedRandomSymbol(r)
//                             }
//                         }
//                     }
//                 }
                
//                 reels[reel][row] = string(symbol)
//             }
//         }

//         // Check for no wins (base game, so isFreeSpin=false)
//         totalWinnings, _ := CalculateWins(reels, 1, 1, false)
//         if totalWinnings == 0 {
//             log.Printf("Generated reels with no wins: %v", reels)
//             break
//         }
//     }

//     return reels
// }

// CalculateWins calculates the total payout and win details using Go concurrency
// Updated to include freeSpinMultiplier parameter for applying multiplier during free spins
func CalculateWins(reels [][]string, betMultiplier int, freeSpinMultiplier int, isFreeSpin bool) (float64, []WinDetail) {
    totalPayout := 0.0
    var winDetails []WinDetail

    // Map to store all unique winning paths: map[pathKey]WinDetail
    uniqueWins := make(map[string]WinDetail)
    // Keep track of which symbol+start combinations we've seen with which counts
    seenCombinations := make(map[string]map[int]bool)
    var winsMu sync.Mutex // Mutex for thread-safe access to our maps

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
                if firstSymbol == string(SymbolSilverFlower) || firstSymbol == string(SymbolGoldFlower) {
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
                    
                    // Calculate payout based on bet multiplier, free spin multiplier, and denomination
                    effectiveMultiplier := 1
                    if isFreeSpin {
                        effectiveMultiplier = freeSpinMultiplier
                    }
                    
                    payout := payoutValue * float64(betMultiplier) * float64(effectiveMultiplier) * Denomination
                    payout = math.Round(payout*100) / 100 // Round to 2 decimal places
                    
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
        log.Printf("Calculating payout: symbol=%s, matchCount=%d, betMultiplier=%d, freeSpinMultiplier=%d, isFreeSpin=%v, payout=%v", 
            win.Symbol, win.Count, betMultiplier, freeSpinMultiplier, isFreeSpin, win.Payout)
        
        totalPayout += win.Payout
        totalPayout = math.Round(totalPayout*100) / 100 // Round to 2 decimal places
        winDetails = append(winDetails, win)
    }
    winsMu.Unlock()

    return totalPayout, winDetails
}

// CountSilverFlowers counts the number of Silver Flower symbols on the reels
func CountSilverFlowers(reels [][]string) int {
    count := 0
    for reel := 0; reel < Reels; reel++ {
        for row := 0; row < Rows; row++ {
            if reels[reel][row] == string(SymbolSilverFlower) {
                count++
            }
        }
    }
    return count
}

// CountGoldFlowers counts the number of Gold Flower symbols on the reels
func CountGoldFlowers(reels [][]string) int {
    count := 0
    for reel := 0; reel < Reels; reel++ {
        for row := 0; row < Rows; row++ {
            if reels[reel][row] == string(SymbolGoldFlower) {
                count++
            }
        }
    }
    return count
}

// GetSilverFlowerPositions returns the positions of all Silver Flower symbols
func GetSilverFlowerPositions(reels [][]string) []Position {
    var positions []Position
    for reel := 0; reel < Reels; reel++ {
        for row := 0; row < Rows; row++ {
            if reels[reel][row] == string(SymbolSilverFlower) {
                positions = append(positions, Position{Reel: reel, Row: row})
            }
        }
    }
    return positions
}

// GetGoldFlowerPositions returns the positions of all Gold Flower symbols
func GetGoldFlowerPositions(reels [][]string) []Position {
    var positions []Position
    for reel := 0; reel < Reels; reel++ {
        for row := 0; row < Rows; row++ {
            if reels[reel][row] == string(SymbolGoldFlower) {
                positions = append(positions, Position{Reel: reel, Row: row})
            }
        }
    }
    return positions
}

// HasSilverFlowerOnEachFirstReel checks if there's at least one Silver Flower on each of reels 0, 1, and 2
func HasSilverFlowerOnEachFirstReel(reels [][]string) bool {
    // Check each of the first three reels for Silver Flowers
    for reel := 0; reel < 3; reel++ {
        hasSilverFlower := false
        for row := 0; row < Rows; row++ {
            if reels[reel][row] == string(SymbolSilverFlower) {
                hasSilverFlower = true
                break
            }
        }
        if !hasSilverFlower {
            // If any of the first three reels doesn't have a Silver Flower, return false
            return false
        }
    }
    return true
}

// CheckBonusMultiplierCondition checks if the bonus multiplier condition is met:
// 1. There must be Silver Flowers on each of reels 0, 1, and 2 (to trigger free spins)
// 2. There must be a Gold Flower on reel 3 or 4
func CheckBonusMultiplierCondition(reels [][]string) bool {
    // First condition: Check for Silver Flowers on reels 0, 1, and 2
    if !HasSilverFlowerOnEachFirstReel(reels) {
        return false
    }
    
    // Check for Gold Flower on reel 4 OR reel 5 (indices 3 OR 4)
    for reel := 3; reel < 5; reel++ {
        for row := 0; row < Rows; row++ {
            if reels[reel][row] == string(SymbolGoldFlower) {
                return true
            }
        }
    }
    return false
}

// HasThreeSilverFlowers checks if there are at least 3 Silver Flower symbols on reels 1-3
// Now updated to specifically check for at least one Silver Flower on EACH of reels 0, 1, and 2
func HasThreeSilverFlowers(reels [][]string) bool {
    return HasSilverFlowerOnEachFirstReel(reels)
}

// GenerateRandomBonusMultiplier generates a random bonus multiplier for Gold Flowers
func GenerateRandomBonusMultiplier() int {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    return r.Intn(MaxGoldFlowerMultiplier-MinGoldFlowerMultiplier+1) + MinGoldFlowerMultiplier
}

// GenerateRandomFreeSpins generates a random number of free spins for Silver Flowers
func GenerateRandomFreeSpins() int {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    return r.Intn(MaxSilverFlowerSpins-MinSilverFlowerSpins+1) + MinSilverFlowerSpins
}