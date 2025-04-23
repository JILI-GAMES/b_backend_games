package magicace

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Constants
const (
    Denomination           = 0.01
    GoldenCardProbability  = 0.05
    Reels                  = 5
    Rows                   = 4
    MinBet                 = 20
)

// Symbol weights for random generation
var SymbolWeights = map[Symbol]float64{
    SymbolA:       0.1,
    SymbolK:       0.1,
    SymbolQ:       0.1,
    SymbolJ:       0.1,
    SymbolHeart:   0.1,
    SymbolSpade:   0.1,
    SymbolClub:    0.1,
    SymbolDiamond: 0.1,
    SymbolScatter: 0.08,
}
// global constant
var globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// Paytable (payouts for Bet Multiplier = 1)
var Paytable = map[Symbol]map[int]float64{
    SymbolA:       {3: 10, 4: 20, 5: 50},
    SymbolK:       {3: 8, 4: 16, 5: 40},
    SymbolQ:       {3: 6, 4: 12, 5: 30},
    SymbolJ:       {3: 4, 4: 8, 5: 20},
    SymbolHeart:   {3: 2, 4: 4, 5: 10},
    SymbolSpade:   {3: 2, 4: 4, 5: 10},
    SymbolClub:    {3: 1, 4: 2, 5: 5},
    SymbolDiamond: {3: 1, 4: 2, 5: 5},
}

// Booming Multipliers (Improvement #9 skipped, keeping original dynamic extension)
var (
    BoomingMultipliers           = []int{1, 2, 3, 5}
    BoomingMultipliersFreeSpins  = []int{2, 4, 6, 10}
    BoomingMultipliersExtraBet   = []int{1, 2, 3} // Will append dynamically for unlimited
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
    return SymbolA // Fallback
}

// GenerateReelsWithWin generates a 5x4 grid with a potential win 
func GenerateReelsWithWin(jokerCards []JokerCard) ([][]string, SpecialSymbols) {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
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
        for _, joker := range jokerCards {
            pos := Position{Reel: joker.Position.Reel, Row: joker.Position.Row}
            reelMap[pos] = string(SymbolWild)
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
                // Add Golden Cards on reels 2, 3, 4 (indices 1, 2, 3) with a 5% chance
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
func GenerateLossReels(jokerCards []JokerCard) ([][]string, SpecialSymbols) {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
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
        for _, joker := range jokerCards {
            pos := Position{Reel: joker.Position.Reel, Row: joker.Position.Row}
            reelMap[pos] = string(SymbolWild)
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
                    SymbolA, SymbolK, SymbolQ, SymbolJ,
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

func GenerateReelsForCascade(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard) ([][]string, SpecialSymbols) {
    // Create a copy of the reels
    newReels := make([][]string, Reels)
    for reel := 0; reel < Reels; reel++ {
        newReels[reel] = make([]string, Rows)
        copy(newReels[reel], reels[reel])
    }

    // Initialize special symbols
    specialSymbols := SpecialSymbols{
        GoldenCards:      nil,
        JokerCards:       jokerCards,
        TargetSymbols:    nil,
        NewTargetSymbols: nil,
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

        // Replace only the winning positions
        for pos := range winningPositions {
            // Skip if the position contains a Joker Card
            isJoker := false
            for _, joker := range jokerCards {
                if joker.Position.Reel == pos.Reel && joker.Position.Row == pos.Row {
                    isJoker = true
                    break
                }
            }
            if isJoker {
                continue
            }

            // Generate a new symbol
            symbol := WeightedRandomSymbol(globalRand)
            // Add Golden Cards on reels 2, 3, 4 with a 5% chance
            if pos.Reel >= 1 && pos.Reel <= 3 && symbol != SymbolScatter && globalRand.Float64() < GoldenCardProbability {
                specialSymbols.GoldenCards = append(specialSymbols.GoldenCards, pos)
                newReels[pos.Reel][pos.Row] = fmt.Sprintf("golden_%s", string(symbol))
            } else {
                newReels[pos.Reel][pos.Row] = string(symbol)
            }

            // Track new Scatters for Free Spin retriggers
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

    return newReels, specialSymbols
}

// GenerateLossForCascade generates new symbols for a cascade with no wins
func GenerateLossForCascade(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard) ([][]string, SpecialSymbols) {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
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

    // Keep generating until no wins are found
    for {
        specialSymbols.GoldenCards = nil
        specialSymbols.NewTargetSymbols = nil

        // Replace only the winning positions, skipping Jokers
        for pos := range winningPositions {
            isJoker := false
            for _, joker := range jokerCards {
                if joker.Position.Reel == pos.Reel && joker.Position.Row == pos.Row {
                    isJoker = true
                    break
                }
            }
            if isJoker {
                continue
            }

            // Avoid Golden Cards and Scatters to minimize wins
            availableSymbols := []Symbol{
                SymbolA, SymbolK, SymbolQ, SymbolJ,
                SymbolHeart, SymbolSpade, SymbolClub, SymbolDiamond,
            }
            symbol := availableSymbols[r.Intn(len(availableSymbols))]
            // Ensure the new symbol breaks potential wins involving Jokers
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
// func GenerateLossForCascade(reels [][]string, winningPositions map[Position]bool, jokerCards []JokerCard) ([][]string, SpecialSymbols) {
//     r := rand.New(rand.NewSource(time.Now().UnixNano()))
//     newReels := make([][]string, Reels)
//     for reel := 0; reel < Reels; reel++ {
//         newReels[reel] = make([]string, Rows)
//         copy(newReels[reel], reels[reel])
//     }

//     var specialSymbols SpecialSymbols
//     specialSymbols.JokerCards = jokerCards

//     // Keep generating until no wins are found (Improvement #6 skipped, no deterministic construction)
//     for {
//         specialSymbols.GoldenCards = nil
//         specialSymbols.TargetSymbols = nil

//         // Replace only the winning positions
//         for pos := range winningPositions {
//             // Don't replace Joker Cards
//             isJoker := false
//             for _, joker := range jokerCards {
//                 if joker.Position.Reel == pos.Reel && joker.Position.Row == pos.Row {
//                     isJoker = true
//                     break
//                 }
//             }
//             if isJoker {
//                 continue
//             }

//             // Avoid Golden Cards and Scatters
//             availableSymbols := []Symbol{
//                 SymbolA, SymbolK, SymbolQ, SymbolJ,
//                 SymbolHeart, SymbolSpade, SymbolClub, SymbolDiamond,
//             }
//             symbol := availableSymbols[r.Intn(len(availableSymbols))]
//             // Avoid matching the previous symbol
//             if pos.Reel > 0 {
//                 previousSymbol := newReels[pos.Reel-1][pos.Row]
//                 for {
//                     symbol = availableSymbols[r.Intn(len(availableSymbols))]
//                     if string(symbol) != previousSymbol {
//                         break
//                     }
//                 }
//             }
//             newReels[pos.Reel][pos.Row] = string(symbol)
//         }

//         // Check for no wins
//         totalWinnings, _ := CalculateWins(newReels, 1, 1, jokerCards)
//         if totalWinnings == 0 {
//             break
//         }
//     }

//     return newReels, specialSymbols
// }

// CalculateWins calculates the total payout and win details (Improvement #2: Track Golden Cards)
func CalculateWins(reels [][]string, betMultiplier int, boomingMultiplier int, jokerCards []JokerCard) (float64, []WinDetail) {
    totalPayout := 0.0
    var winDetails []WinDetail
    seenWins := make(map[string]bool)
    var seenWinsMu sync.Mutex // Mutex for thread-safe access to seenWins

    // Worker pool configuration
    const maxWorkers = 50
    jobs := make(chan []int, len(WaysToWin))
    results := make(chan WinDetail, len(WaysToWin))
    var wg sync.WaitGroup // WaitGroup to track worker completion

    // Start workers
    for w := 0; w < maxWorkers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for way := range jobs {
                // Preallocate slices to their expected sizes (5 reels)
                symbols := make([]string, 0, 5)
                payline := make([]Position, 0, 5)
                positions := make([]string, 0, 5)
                goldenCards := make([]Position, 0, 5)

                for reel, row := range way {
                    symbol := reels[reel][row]
                    // Check for Joker Cards
                    for _, joker := range jokerCards {
                        if joker.Position.Reel == reel && joker.Position.Row == row {
                            symbol = string(SymbolWild)
                        }
                    }
                    symbols = append(symbols, symbol)
                    pos := Position{Reel: reel, Row: row}
                    payline = append(payline, pos)
                    positions = append(positions, strconv.Itoa(reel)+","+strconv.Itoa(row))
                    if strings.HasPrefix(symbol, "golden_") {
                        goldenCards = append(goldenCards, pos)
                    }
                }

                // Count matching symbols
                firstSymbol := symbols[0]
                if firstSymbol == string(SymbolScatter) || strings.HasPrefix(firstSymbol, "golden_") {
                    firstSymbol = strings.TrimPrefix(firstSymbol, "golden_")
                }
                if firstSymbol == string(SymbolScatter) {
                    continue
                }
                matchCount := 1
                for i := 1; i < len(symbols); i++ {
                    currentSymbol := symbols[i]
                    if strings.HasPrefix(currentSymbol, "golden_") {
                        baseSymbol := strings.TrimPrefix(currentSymbol, "golden_")
                        if baseSymbol == firstSymbol || currentSymbol == string(SymbolWild) {
                            matchCount++
                        } else {
                            break
                        }
                    } else if currentSymbol == firstSymbol || currentSymbol == string(SymbolWild) {
                        matchCount++
                    } else {
                        break
                    }
                }

                if matchCount >= 3 {
                    // Create a unique key for the win
                    winKey := strings.Join(positions[:matchCount], "|")
                    seenWinsMu.Lock()
                    if seenWins[winKey] {
                        seenWinsMu.Unlock()
                        continue
                    }
                    seenWins[winKey] = true
                    seenWinsMu.Unlock()

                    payout := Paytable[Symbol(firstSymbol)][matchCount] * Denomination * float64(betMultiplier) * float64(boomingMultiplier)
                    results <- WinDetail{
                        Symbols:     symbols[:matchCount],
                        Payline:     payline[:matchCount],
                        Payout:      payout,
                        GoldenCards: goldenCards,
                    }
                }
            }
        }()
    }

    // Send jobs to workers
    for _, way := range WaysToWin {
        jobs <- way
    }
    close(jobs)

    // Close the results channel after all workers are done
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect results
    for win := range results {
        totalPayout += win.Payout
        winDetails = append(winDetails, win)
    }

    return totalPayout, winDetails
}
// func CalculateWins(reels [][]string, betMultiplier int, boomingMultiplier int, jokerCards []JokerCard) (float64, []WinDetail) {
//     totalPayout := 0.0
//     var winDetails []WinDetail
//     seenWins := make(map[string]bool) // To deduplicate wins based on positions

//     // Check for wins on all 1024 ways concurrently
//     var wg sync.WaitGroup
//     winChan := make(chan WinDetail, 1024)

//     for _, way := range WaysToWin {
//         wg.Add(1)
//         go func(way []int) {
//             defer wg.Done()
//             symbols := make([]string, 0)
//             payline := make([]Position, 0)
//             positions := make([]string, 0) // For deduplication
//             goldenCards := make([]Position, 0) // Track Golden Cards in this win
//             for reel, row := range way {
//                 symbol := reels[reel][row]
//                 // Check for Joker Cards
//                 for _, joker := range jokerCards {
//                     if joker.Position.Reel == reel && joker.Position.Row == row {
//                         symbol = string(SymbolWild)
//                     }
//                 }
//                 symbols = append(symbols, symbol)
//                 pos := Position{Reel: reel, Row: row}
//                 payline = append(payline, pos)
//                 positions = append(positions, fmt.Sprintf("%d,%d", reel, row))
//                 if strings.HasPrefix(symbol, "golden_") {
//                     goldenCards = append(goldenCards, pos)
//                 }
//             }

//             // Count matching symbols
//             firstSymbol := symbols[0]
//             if firstSymbol == string(SymbolScatter) || strings.HasPrefix(firstSymbol, "golden_") {
//                 firstSymbol = strings.TrimPrefix(firstSymbol, "golden_")
//             }
//             if firstSymbol == string(SymbolScatter) {
//                 return
//             }
//             matchCount := 1
//             for i := 1; i < len(symbols); i++ {
//                 currentSymbol := symbols[i]
//                 if strings.HasPrefix(currentSymbol, "golden_") {
//                     baseSymbol := strings.TrimPrefix(currentSymbol, "golden_")
//                     if baseSymbol == firstSymbol || currentSymbol == string(SymbolWild) {
//                         matchCount++
//                     } else {
//                         break
//                     }
//                 } else if currentSymbol == firstSymbol || currentSymbol == string(SymbolWild) {
//                     matchCount++
//                 } else {
//                     break
//                 }
//             }

//             if matchCount >= 3 {
//                 // Create a unique key for the win
//                 winKey := strings.Join(positions[:matchCount], "|")
//                 if seenWins[winKey] {
//                     return
//                 }
//                 seenWins[winKey] = true

//                 payout := Paytable[Symbol(firstSymbol)][matchCount] * Denomination * float64(betMultiplier) * float64(boomingMultiplier)
//                 winChan <- WinDetail{
//                     Symbols:     symbols[:matchCount],
//                     Payline:     payline[:matchCount],
//                     Payout:      payout,
//                     GoldenCards: goldenCards,
//                 }
//             }
//         }(way)
//     }

//     wg.Wait()
//     close(winChan)

//     for win := range winChan {
//         totalPayout += win.Payout
//         winDetails = append(winDetails, win)
//     }

//     return totalPayout, winDetails
// }

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

// TransformGoldenCards transforms Golden Cards that are part of a win into Joker Cards
func TransformGoldenCards(reels [][]string, lastWinDetails []WinDetail) []JokerCard {
    var newJokerCards []JokerCard
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    occupiedPositions := make(map[Position]bool)
    for _, joker := range newJokerCards {
        occupiedPositions[joker.Position] = true
    }
    for reel := 0; reel < Reels; reel++ {
        for row := 0; row < Rows; row++ {
            symbol := reels[reel][row]
            if strings.Contains(symbol, "Joker") || strings.HasPrefix(symbol, "golden_") || symbol == string(SymbolScatter) {
                occupiedPositions[Position{Reel: reel, Row: row}] = true
            }
        }
    }
    for _, win := range lastWinDetails {
        for _, pos := range win.GoldenCards {
            symbol := reels[pos.Reel][pos.Row]
            if strings.HasPrefix(symbol, "golden_") {
                mode := "smallJoker"
                if r.Intn(100) < 10 {
                    mode = "superJoker"
                } else if r.Intn(100) < 30 {
                    mode = "bigJoker"
                }
                newJokerCards = append(newJokerCards, JokerCard{
                    Position:        pos,
                    Mode:            mode,
                    RemainingRounds: 3,
                })
                reels[pos.Reel][pos.Row] = string(SymbolWild)
                occupiedPositions[pos] = true
                // If Big Joker, duplicate to a random position
                if mode == "bigJoker" {
                    newReel, newRow := getRandomPosition(reels, r, pos.Reel, pos.Row, occupiedPositions)
                    if newReel != -1 && newRow != -1 {
                        newPos := Position{Reel: newReel, Row: newRow}
                        newJokerCards = append(newJokerCards, JokerCard{
                            Position:        newPos,
                            Mode:            "bigJoker",
                            RemainingRounds: 3,
                        })
                        reels[newReel][newRow] = string(SymbolWild)
                        occupiedPositions[newPos] = true
                    }
                }
            }
        }
    }
    return newJokerCards
}

// getRandomPosition finds a random position on the reels, avoiding the given position and occupied positions
func getRandomPosition(reels [][]string, r *rand.Rand, excludeReel, excludeRow int, occupiedPositions map[Position]bool) (int, int) {
    for {
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
}