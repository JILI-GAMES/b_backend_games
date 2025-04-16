package moneybagsman2

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Constants
const (
    Denomination = 0.01
    Reels       = 5
    Rows        = 3
)

// Symbol weights for random generation
var SymbolWeights = map[Symbol]float64{
    SymbolAirplane:   0.1,
    SymbolYacht:      0.1,
    SymbolCar:        0.1,
    SymbolMotorcycle: 0.1,
    SymbolA:          0.1,
    SymbolK:          0.1,
    SymbolQ:          0.1,
    SymbolJ:          0.1,
    SymbolWild:       0.08,
    SymbolScatter:    0.08,
}

// Paytable (payouts for Bet Multiplier = 1)
var Paytable = map[Symbol]map[int]float64{
    SymbolAirplane:   {3: 95, 4: 200, 5: 500},
    SymbolYacht:      {3: 60, 4: 180, 5: 350},
    SymbolCar:        {3: 50, 4: 120, 5: 300},
    SymbolMotorcycle: {3: 35, 4: 120, 5: 240},
    SymbolA:          {3: 20, 4: 35, 5: 150},
    SymbolK:          {3: 20, 4: 30, 5: 130},
    SymbolQ:          {3: 15, 4: 25, 5: 120},
    SymbolJ:          {3: 15, 4: 25, 5: 120},
}

// BetAmountToMultiplier maps bet amounts to multipliers
var BetAmountToMultiplier = map[float64]int{
    0.6: 1,
    1.2: 2,
    1.8: 3,
    3.0: 5,
    6.0: 10,
}

// FreeSpinOptions defines the Free Spin Bonus options
var FreeSpinOptions = map[int]struct {
    TotalSpins      int
    InitialMultiplier int
    MaxSpins        int
    MultiplierIncrease int
}{
    1: {TotalSpins: 12, InitialMultiplier: 1, MaxSpins: 50, MultiplierIncrease: 1},
    2: {TotalSpins: 8, InitialMultiplier: 2, MaxSpins: 24, MultiplierIncrease: 2},
    3: {TotalSpins: 5, InitialMultiplier: 5, MaxSpins: 15, MultiplierIncrease: 5},
    4: {TotalSpins: 3, InitialMultiplier: 12, MaxSpins: 9, MultiplierIncrease: 12},
    5: {TotalSpins: 1, InitialMultiplier: 50, MaxSpins: 3, MultiplierIncrease: 0}, // No Increase for option 5
}

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

// GenerateReelsWithWin generates a 5x3 grid with a guaranteed win
func GenerateReelsWithWin() [][]string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    var reels [][]string

    // Keep generating until a win is found
    for {
        reels = make([][]string, Reels)
        for reel := 0; reel < Reels; reel++ {
            reels[reel] = make([]string, Rows)
            for row := 0; row < Rows; row++ {
                symbol := WeightedRandomSymbol(r)
                // Wilds only on reels 2-5 (indices 1-4)
                if reel == 0 && symbol == SymbolWild {
                    symbol = WeightedRandomSymbol(r)
                    for symbol == SymbolWild {
                        symbol = WeightedRandomSymbol(r)
                    }
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

// GenerateLossReels generates a 5x3 grid with no wins but allows Free Spin Bonus triggers
func GenerateLossReels() [][]string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    var reels [][]string

    // Keep generating until no wins are found
    for {
        reels = make([][]string, Reels)
        for reel := 0; reel < Reels; reel++ {
            reels[reel] = make([]string, Rows)
            for row := 0; row < Rows; row++ {
                symbol := WeightedRandomSymbol(r)
                // Wilds only on reels 2-5 (indices 1-4)
                if reel == 0 && symbol == SymbolWild {
                    symbol = WeightedRandomSymbol(r)
                    for symbol == SymbolWild {
                        symbol = WeightedRandomSymbol(r)
                    }
                }
                // Avoid matching symbols to prevent wins
                if reel > 0 {
                    previousSymbol := reels[reel-1][row]
                    for string(symbol) == previousSymbol && symbol != SymbolScatter {
                        symbol = WeightedRandomSymbol(r)
                        if reel == 0 && symbol == SymbolWild {
                            symbol = WeightedRandomSymbol(r)
                            for symbol == SymbolWild {
                                symbol = WeightedRandomSymbol(r)
                            }
                        }
                    }
                }
                reels[reel][row] = string(symbol)
            }
        }

        // Check for no wins (base game, so isFreeSpin=false)
        totalWinnings, _ := CalculateWins(reels, 1, 1, false)
        if totalWinnings == 0 {
            break
        }
    }
    // Log the reels for debugging
    log.Printf("Generated reels with no wins: %v", reels)

    return reels
}

// CalculateWins calculates the total payout and win details using Go concurrency
func CalculateWins(reels [][]string, betMultiplier int, freeSpinMultiplier int, isFreeSpin bool) (float64, []WinDetail) {
    totalPayout := 0.0
    var winDetails []WinDetail

    // Map to store the MAXIMUM win for each payline start and symbol: map[startKey+symbol]WinDetail
    maxWinsByStartAndSymbol := make(map[string]WinDetail)
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
                if firstSymbol == string(SymbolScatter) {
                    continue
                }

                // If the first symbol is Wild, find the first non-Wild, non-Scatter symbol
                if firstSymbol == string(SymbolWild) {
                    for i := 1; i < len(symbols); i++ {
                        if symbols[i] != string(SymbolWild) && symbols[i] != string(SymbolScatter) {
                            firstSymbol = symbols[i]
                            break
                        }
                    }
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
                    // Apply freeSpinMultiplier only during free spins; otherwise, use 1
                    effectiveMultiplier := 1
                    if isFreeSpin {
                        effectiveMultiplier = freeSpinMultiplier
                    }
                    payout := Paytable[Symbol(firstSymbol)][matchCount] * float64(betMultiplier) * float64(effectiveMultiplier) * Denomination
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

    // Process wins, keeping only the longest match for each starting positions and symbol
    for win := range results {
        if len(win.Positions) < 2 {
            continue
        }
        
        // Create a key from the first two positions and symbol to group related paths
        startKey := fmt.Sprintf("%s|%d,%d|%d,%d", 
            win.Symbol,
            win.Positions[0].Reel, win.Positions[0].Row,
            win.Positions[1].Reel, win.Positions[1].Row)
        
        winsMu.Lock()
        
        // Only keep the win with the highest match count (longest path)
        existingWin, exists := maxWinsByStartAndSymbol[startKey]
        if !exists || win.Count > existingWin.Count || 
           (win.Count == existingWin.Count && win.Payout > existingWin.Payout) {
            maxWinsByStartAndSymbol[startKey] = win
        }
        
        winsMu.Unlock()
    }

    // Collect all the maximum wins
    winsMu.Lock()
    for _, win := range maxWinsByStartAndSymbol {
        log.Printf("Calculating payout: symbol=%s, matchCount=%d, betMultiplier=%d, effectiveMultiplier=%d, payout=%v", 
            win.Symbol, win.Count, betMultiplier, 
            func() int {
                if isFreeSpin {
                    return freeSpinMultiplier
                }
                return 1
            }(), win.Payout)
        
        totalPayout += win.Payout
        totalPayout = math.Round(totalPayout*100) / 100
        winDetails = append(winDetails, win)
    }
    winsMu.Unlock()

    return totalPayout, winDetails
}

// // CalculateWins calculates the total payout and win details using Go concurrency
// func CalculateWins(reels [][]string, betMultiplier int, freeSpinMultiplier int, isFreeSpin bool) (float64, []WinDetail) {
//     totalPayout := 0.0
//     var winDetails []WinDetail

//     // Map to store the MAXIMUM win for each unique payline path: map[pathKey]WinDetail
//     maxWinsByPath := make(map[string]WinDetail)
//     var winsMu sync.Mutex // Mutex for thread-safe access to our maps

//     // Worker pool configuration
//     const maxWorkers = 50
//     jobs := make(chan struct {
//         wayIndex int
//         way      []int
//     }, len(WaysToWin))
//     results := make(chan WinDetail, len(WaysToWin))
//     var wg sync.WaitGroup

//     // Start workers to identify potential wins
//     for w := 0; w < maxWorkers; w++ {
//         wg.Add(1)
//         go func() {
//             defer wg.Done()
//             for job := range jobs {
//                 way := job.way

//                 // Preallocate slices
//                 symbols := make([]string, 0, 5)
//                 payline := make([]Position, 0, 5)

//                 for reel, row := range way {
//                     symbol := reels[reel][row]
//                     symbols = append(symbols, symbol)
//                     pos := Position{Reel: reel, Row: row}
//                     payline = append(payline, pos)
//                 }

//                 // Count matching symbols
//                 firstSymbol := symbols[0]
//                 if firstSymbol == string(SymbolScatter) {
//                     continue
//                 }

//                 // If the first symbol is Wild, find the first non-Wild, non-Scatter symbol
//                 if firstSymbol == string(SymbolWild) {
//                     for i := 1; i < len(symbols); i++ {
//                         if symbols[i] != string(SymbolWild) && symbols[i] != string(SymbolScatter) {
//                             firstSymbol = symbols[i]
//                             break
//                         }
//                     }
//                 }
//                 matchCount := 1
//                 for i := 1; i < len(symbols); i++ {
//                     currentSymbol := symbols[i]
//                     if currentSymbol == firstSymbol || currentSymbol == string(SymbolWild) {
//                         matchCount++
//                     } else {
//                         break
//                     }
//                 }

//                 if matchCount >= 3 {
//                     // Apply freeSpinMultiplier only during free spins; otherwise, use 1
//                     effectiveMultiplier := 1
//                     if isFreeSpin {
//                         effectiveMultiplier = freeSpinMultiplier
//                     }
//                     payout := Paytable[Symbol(firstSymbol)][matchCount] * float64(betMultiplier) * float64(effectiveMultiplier) * Denomination
//                     win := WinDetail{
//                         Symbol:    firstSymbol,
//                         Count:     matchCount,
//                         Payout:    payout,
//                         Positions: payline[:matchCount],
//                     }
//                     results <- win
//                 }
//             }
//         }()
//     }

//     // Send jobs to workers
//     for wayIndex, way := range WaysToWin {
//         jobs <- struct {
//             wayIndex int
//             way      []int
//         }{wayIndex: wayIndex, way: way}
//     }
//     close(jobs)

//     // Close the results channel after all workers are done
//     go func() {
//         wg.Wait()
//         close(results)
//     }()

//     // Process wins, keeping all unique payline paths
//     for win := range results {
//         if len(win.Positions) < 2 {
//             continue
//         }
        
//         // Create a key from the entire payline path to uniquely identify it
//         var pathKey strings.Builder
//         pathKey.WriteString(win.Symbol) // Include symbol in the key
//         for _, pos := range win.Positions {
//             pathKey.WriteString(fmt.Sprintf("|%d,%d", pos.Reel, pos.Row))
//         }
        
//         winsMu.Lock()
        
//         // Store this unique payline win
//         maxWinsByPath[pathKey.String()] = win
        
//         winsMu.Unlock()
//     }

//     // Collect all the maximum wins
//     winsMu.Lock()
//     for _, win := range maxWinsByPath {
//         log.Printf("Calculating payout: symbol=%s, matchCount=%d, betMultiplier=%d, effectiveMultiplier=%d, payout=%v", 
//             win.Symbol, win.Count, betMultiplier, 
//             func() int {
//                 if isFreeSpin {
//                     return freeSpinMultiplier
//                 }
//                 return 1
//             }(), win.Payout)
        
//         totalPayout += win.Payout
//         winDetails = append(winDetails, win)
//     }
//     winsMu.Unlock()

//     return totalPayout, winDetails
// }



// CountScatters counts the number of Scatter symbols on the reels
func CountScatters(reels [][]string) int {
    count := 0
    for reel := 0; reel < Reels; reel++ {
        for row := 0; row < Rows; row++ {
            if reels[reel][row] == string(SymbolScatter) {
                count++
            }
        }
    }
    return count
}

// HasScatterOnEachReel checks if there’s at least one Scatter on each reel
func HasScatterOnEachReel(reels [][]string) bool {
    for reel := 0; reel < Reels; reel++ {
        hasScatter := false
        for row := 0; row < Rows; row++ {
            if reels[reel][row] == string(SymbolScatter) {
                hasScatter = true
                break
            }
        }
        if !hasScatter {
            return false
        }
    }
    return true
}

// CalculateExtraBonusMultiplier calculates the extra bonus multiplier based on Scatter count
func CalculateExtraBonusMultiplier(scatterCount int) int {
    if scatterCount >= 7 {
        return 3
    } else if scatterCount == 6 {
        return 2
    } else if scatterCount == 5 {
        return 1
    }
    return 0
}