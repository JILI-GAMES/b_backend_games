package opensesame2

import (
    "fmt"
    "log"
    "math"
    "math/rand"
    "strings"
    "sync"
    "time"
)

// Constants
const (
    Denomination       = 0.01
    Reels              = 5
    Rows               = 3
    CreditMultiplier   = 60     // Minimum bet is 60 credits per bet multiplier 
    MinFreeSpins       = 5      // Minimum free spins awarded
    MaxFreeSpins       = 15     // Maximum free spins awarded
    MinFreeSpinMult    = 2      // Minimum free spin multiplier
    MaxFreeSpinMult    = 6      // Maximum free spin multiplier
    MinExtraFreeSpins  = 2      // Minimum extra free spins
    MaxExtraFreeSpins  = 10     // Maximum extra free spins
    MinExtraMultiplier = 1      // Minimum extra multiplier
    MaxExtraMultiplier = 5      // Maximum extra multiplier
    MaxTotalFreeSpins  = 250    // Maximum total free spins possible
)

// Symbol weights for random generation
var SymbolWeights = map[Symbol]float64{
    SymbolWoman:      0.05,
    SymbolMan:        0.08,
    SymbolHorse:      0.08,
    SymbolPottery:    0.08,
    SymbolA:          0.09,
    SymbolK:          0.09,
    SymbolQ:          0.11,
    SymbolJ:          0.11,
    Symbol10:         0.11,
    Symbol9:          0.11,
    SymbolWild:       0.05,
    SymbolScatter:    0.04,
    SymbolFreeSpins:  0.03,
    SymbolMysteryBox: 0.02, // New symbol with lower weight
}

// Paytable (payouts for Bet Multiplier = 1) - Updated for Open Sesame 2
var Paytable = map[Symbol]map[int]int{
    SymbolWoman:      {3: 60, 4: 200, 5: 360},
    SymbolMan:        {3: 30, 4: 100, 5: 200},
    SymbolHorse:      {3: 24, 4: 48, 5: 120},
    SymbolPottery:    {3: 24, 4: 48, 5: 120},
    SymbolA:          {3: 18, 4: 30, 5: 80},
    SymbolK:          {3: 18, 4: 30, 5: 80},
    SymbolQ:          {3: 18, 4: 24, 5: 72},
    SymbolJ:          {3: 18, 4: 24, 5: 72},
    Symbol10:         {3: 18, 4: 24, 5: 60},
    Symbol9:          {3: 18, 4: 24, 5: 60},
    SymbolScatter:    {3: 2, 4: 10, 5: 40},
    SymbolFreeSpins:  {3: 2},
    SymbolMysteryBox: {},
}

// BetAmountMap maps bet amounts to multipliers
var BetAmountMap = map[float64]int{
    0.6: 1,
    1.2: 2,
    3.0: 5,
    6.0: 10,
    15.0: 25,
}

// Chest options for revealing free spin counts
var ChestOptions = []int{5, 8, 10, 15}

// Lamp options for revealing multipliers
var LampOptions = []int{2, 3, 4, 5, 6}

// Treasure options for extra feature
var TreasureMultiplierOptions = []int{1, 2, 3, 4, 5}
var TreasureFreeSpinOptions = []int{2, 4, 6, 8, 10}

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

// WeightedRandomSymbol selects a symbol based on weights, with reel-specific restrictions
func WeightedRandomSymbol(r *rand.Rand, reelIndex int) Symbol {
    totalWeight := 0.0
    for symbol, weight := range SymbolWeights {
        // Wild only appears on reels 2-5
        if symbol == SymbolWild && reelIndex == 0 {
            continue
        }
        // Free Spin symbols only appear on reels 1-3
        if symbol == SymbolFreeSpins && reelIndex > 2 {
            continue
        }
        // Mystery Box only appears on reel 3
        if symbol == SymbolMysteryBox && reelIndex != 2 {
            continue
        }
        totalWeight += weight
    }

    roll := r.Float64() * totalWeight
    currentWeight := 0.0
    for symbol, weight := range SymbolWeights {
        // Skip symbols that can't appear on this reel
        if symbol == SymbolWild && reelIndex == 0 {
            continue
        }
        if symbol == SymbolFreeSpins && reelIndex > 2 {
            continue
        }
        if symbol == SymbolMysteryBox && reelIndex != 2 {
            continue
        }

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
                symbol := WeightedRandomSymbol(r, reel)
                reels[reel][row] = string(symbol)
            }
        }

        // Check for a win (base game, so isFreeSpin=false and multiplier=1)
        totalWinnings, _, _, _, _, _ := CalculateWins(reels, 1, 1, 1, false, false)
        if totalWinnings > 0 {
            log.Printf("Generated reels with win: %v", reels)
            break
        }
    }

    return reels
}

// GenerateLossReels generates a 5x3 grid with no wins but allows triggering features
func GenerateLossReels() [][]string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    var reels [][]string

    // Keep generating until no regular wins are found
    for {
        reels = make([][]string, Reels)
        for reel := 0; reel < Reels; reel++ {
            reels[reel] = make([]string, Rows)
            for row := 0; row < Rows; row++ {
                symbol := WeightedRandomSymbol(r, reel)
                reels[reel][row] = string(symbol)
            }
        }

        // Check for no regular wins (exclude scatter wins)
        totalWinnings, _, _, _, _, _ := CalculateWins(reels, 1, 1, 1, false, false)
        if totalWinnings == 0 {
            log.Printf("Generated reels with no wins: %v", reels)
            break
        }
    }

    return reels
}

// CalculateWins calculates the total payout and win details
func CalculateWins(reels [][]string, betMultiplier int, freeSpinMultiplier int, extraMultiplier int, isFreeSpin bool, isExtraFreeSpin bool) (float64, []WinDetail, float64, int, float64, []Position) {
    totalPayout := 0.0
    var winDetails []WinDetail
    scatterPayout := 0.0
    freeSpinCount := 0
    freeSpinPayout := 0.0
    
    // Get Mystery Box positions
    mysteryBoxPositions := GetSymbolPositions(reels, string(SymbolMysteryBox))

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
                if firstSymbol == string(SymbolScatter) || 
                   firstSymbol == string(SymbolFreeSpins) || 
                   firstSymbol == string(SymbolMysteryBox) {
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
                    // Apply free spin multiplier but exclude 5 Woman symbols
                    effectiveMultiplier := 1
                    if (isFreeSpin || isExtraFreeSpin) && (firstSymbol != string(SymbolWoman) || matchCount < 5) {
                        effectiveMultiplier = freeSpinMultiplier
                        
                        // Apply extra multiplier for extra free spins
                        if isExtraFreeSpin {
                            effectiveMultiplier *= extraMultiplier
                        }
                    }
                    
                    // Check if this symbol/count has a payout in the paytable
                    payoutValue, exists := Paytable[Symbol(firstSymbol)][matchCount]
                    if !exists {
                        continue // Skip if no payout for this combination
                    }
                    
                    // Calculate payout based on payline, bet multiplier, and free spin multiplier
                    payout := float64(payoutValue) * float64(betMultiplier) * float64(effectiveMultiplier) * Denomination
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
        effectiveMultiplier := 1
        if (isFreeSpin || isExtraFreeSpin) && (win.Symbol != string(SymbolWoman) || win.Count < 5) {
            effectiveMultiplier = freeSpinMultiplier
            
            // Apply extra multiplier for extra free spins
            if isExtraFreeSpin {
                effectiveMultiplier *= extraMultiplier
            }
        }
        
        log.Printf("Calculating payout: symbol=%s, matchCount=%d, betMultiplier=%d, effectiveMultiplier=%d, payout=%v", 
            win.Symbol, win.Count, betMultiplier, effectiveMultiplier, win.Payout)
        
        totalPayout += win.Payout
        totalPayout = math.Round(totalPayout*100) / 100
        winDetails = append(winDetails, win)
    }
    winsMu.Unlock()

    // Calculate Scatter payouts separately
    scatterPositions := GetSymbolPositions(reels, string(SymbolScatter))
    scatterCount := len(scatterPositions)
    if scatterCount >= 3 {
        // Scatter pay is the credit/odds value multiplied by the bet amount
        scatterPayValue := float64(Paytable[SymbolScatter][scatterCount])
        totalBetAmount := float64(betMultiplier * CreditMultiplier) * Denomination
        scatterPayout = scatterPayValue * totalBetAmount
        scatterPayout = math.Round(scatterPayout*100) / 100
        log.Printf("Scatter payout: count=%d, odds=%v, betMultiplier=%d, totalBetAmount=%v, payout=%v", 
            scatterCount, scatterPayValue, betMultiplier, totalBetAmount, scatterPayout)
    }
    
    // Calculate Free Spin symbol payouts separately
    freeSpinPositions := GetFreeSpinSymbolPositions(reels)
    freeSpinCount = len(freeSpinPositions)
    if freeSpinCount >= 3 {
        // Free Spin symbol pay is also the credit/odds value multiplied by the bet amount
        freeSpinPayValue := float64(Paytable[SymbolFreeSpins][freeSpinCount])
        totalBetAmount := float64(betMultiplier * CreditMultiplier) * Denomination
        freeSpinPayout = freeSpinPayValue * totalBetAmount
        freeSpinPayout = math.Round(freeSpinPayout*100) / 100
        log.Printf("Free Spin symbol payout: count=%d, odds=%v, betMultiplier=%d, totalBetAmount=%v, payout=%v", 
            freeSpinCount, freeSpinPayValue, betMultiplier, totalBetAmount, freeSpinPayout)
    }

    return totalPayout, winDetails, scatterPayout, freeSpinCount, freeSpinPayout, mysteryBoxPositions
}

// GetSymbolPositions finds positions of a specific symbol on the reels
func GetSymbolPositions(reels [][]string, symbol string) []Position {
    var positions []Position
    
    for reel := 0; reel < len(reels); reel++ {
        for row := 0; row < len(reels[reel]); row++ {
            if reels[reel][row] == symbol {
                positions = append(positions, Position{
                    Reel: reel,
                    Row: row,
                })
            }
        }
    }
    
    return positions
}

// GetFreeSpinSymbolPositions finds positions of Free Spin symbols on the first three reels
func GetFreeSpinSymbolPositions(reels [][]string) []Position {
    var positions []Position
    
    // Only check the first three reels
    for reel := 0; reel < 3 && reel < len(reels); reel++ {
        for row := 0; row < len(reels[reel]); row++ {
            if reels[reel][row] == string(SymbolFreeSpins) {
                positions = append(positions, Position{
                    Reel: reel,
                    Row: row,
                })
            }
        }
    }
    
    return positions
}

// CountFreeSpinSymbolsOnFirstThreeReels counts free spin symbols on the first three reels
func CountFreeSpinSymbolsOnFirstThreeReels(reels [][]string) int {
    return len(GetFreeSpinSymbolPositions(reels))
}

// HasFreeSpinSymbolsOnFirstThreeReels checks if there's at least one Free Spin symbol on each of the first three reels
func HasFreeSpinSymbolsOnFirstThreeReels(reels [][]string) bool {
    for reel := 0; reel < 3; reel++ {
        hasFreeSpinSymbol := false
        for row := 0; row < Rows; row++ {
            if reels[reel][row] == string(SymbolFreeSpins) {
                hasFreeSpinSymbol = true
                break
            }
        }
        if !hasFreeSpinSymbol {
            return false
        }
    }
    return true
}

// CheckForExtraFreeSpinTrigger checks if the extra free spin bonus is triggered
// This requires 2 Free Spin symbols on reels 1 and 2, plus a Mystery Box on reel 3
func CheckForExtraFreeSpinTrigger(reels [][]string) bool {
    // Check for Free Spin symbols on reels 1 and 2
    hasReel1FreeSpins := false
    hasReel2FreeSpins := false
    
    for row := 0; row < Rows; row++ {
        if reels[0][row] == string(SymbolFreeSpins) {
            hasReel1FreeSpins = true
        }
        if reels[1][row] == string(SymbolFreeSpins) {
            hasReel2FreeSpins = true
        }
    }
    
    // Check for Mystery Box on reel 3
    hasMysteryBox := false
    for row := 0; row < Rows; row++ {
        if reels[2][row] == string(SymbolMysteryBox) {
            hasMysteryBox = true
            break
        }
    }
    
    return hasReel1FreeSpins && hasReel2FreeSpins && hasMysteryBox
}

// GenerateFreeSpinOptions randomly selects free spin options
func GenerateFreeSpinOptions() (int, int) {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    
    // Randomly select number of free spins (5, 8, 10, or 15)
    freeSpinCount := ChestOptions[r.Intn(len(ChestOptions))]
    
    // Randomly select multiplier (2x to 6x)
    multiplier := LampOptions[r.Intn(len(LampOptions))]
    
    return freeSpinCount, multiplier
}

// GenerateExtraFreeSpinOption randomly selects either extra multiplier or extra free spins
func GenerateExtraFreeSpinOption() (int, int, bool) {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    
    // 50/50 chance for extra multiplier or extra free spins
    isMultiplier := r.Intn(2) == 0
    
    if isMultiplier {
        // Extra multiplier (1x to 5x)
        extraMultiplier := TreasureMultiplierOptions[r.Intn(len(TreasureMultiplierOptions))]
        return extraMultiplier, 0, true
    } else {
        // Extra free spins (2 to 10)
        extraFreeSpins := TreasureFreeSpinOptions[r.Intn(len(TreasureFreeSpinOptions))]
        return 0, extraFreeSpins, false
    }
}

// GetSelectedFreeSpinOptions returns the free spin options based on player selection
func GetSelectedFreeSpinOptions(chestIndex, lampIndex int) (int, int, error) {
    // Validate indices
    if chestIndex < 0 || chestIndex >= len(ChestOptions) {
        return 0, 0, fmt.Errorf("invalid chest index: %d", chestIndex)
    }
    if lampIndex < 0 || lampIndex >= len(LampOptions) {
        return 0, 0, fmt.Errorf("invalid lamp index: %d", lampIndex)
    }
    
    return ChestOptions[chestIndex], LampOptions[lampIndex], nil
}

// GetSelectedExtraOption returns the extra feature option based on player selection
func GetSelectedExtraOption(treasureIndex int) (int, int, bool, error) {
    // Validate index
    if treasureIndex < 0 || treasureIndex >= 10 { // Total of 10 options: 5 multipliers and 5 free spin options
        return 0, 0, false, fmt.Errorf("invalid treasure index: %d", treasureIndex)
    }
    
    // First 5 options are multipliers, next 5 are free spins
    if treasureIndex < 5 {
        // It's a multiplier
        return TreasureMultiplierOptions[treasureIndex], 0, true, nil
    } else {
        // It's extra free spins
        return 0, TreasureFreeSpinOptions[treasureIndex-5], false, nil
    }
}