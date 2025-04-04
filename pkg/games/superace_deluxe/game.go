package superace_deluxe

import (
    "fmt"
    "math/rand"
    "time"
)

type SymbolWeight struct {
    Symbol string
    Weight int
}

var (
    SymbolWeights = []SymbolWeight{
        {"ACE", 20},
        {"KING", 20},
        {"QUEEN", 20},
        {"JACK", 20},
        {"SPADE", 10},
        {"HEART", 10},
        {"DIAMOND", 10},
        {"CLUB", 10},
        {"SCATTER", 5},
    }
    NonScatterSymbolWeights = []SymbolWeight{
        {"ACE", 20},
        {"KING", 20},
        {"QUEEN", 20},
        {"JACK", 20},
        {"SPADE", 10},
        {"HEART", 10},
        {"DIAMOND", 10},
        {"CLUB", 10},
    }
    // Valid bet amounts as per the game interface
    ValidBetAmounts = []float64{0.5, 1, 2, 3, 5, 10, 20, 30, 40, 50, 80, 100, 200, 500, 1000}
    // Define paytables for different bet amounts
    Payouts05 = map[string]map[int]float64{ // For bet amounts 0.5
        "ACE":     {5: 1.25, 4: 0.75, 3: 0.25},
        "KING":    {5: 1, 4: 0.6, 3: 0.2},
        "QUEEN":   {5: 0.75, 4: 0.45, 3: 0.15},
        "JACK":    {5: 0.5, 4: 0.3, 3: 0.1},
        "SPADE":   {5: 0.25, 4: 0.15, 3: 0.05},
        "HEART":   {5: 0.25, 4: 0.15, 3: 0.05},
        "DIAMOND": {5: 0.13, 4: 0.08, 3: 0.03},
        "CLUB":    {5: 0.13, 4: 0.08, 3: 0.03},
    }
    Payouts1 = map[string]map[int]float64{ // For bet amounts 1
        "ACE":     {5: 2.5, 4: 1.5, 3: 0.5},
        "KING":    {5: 2, 4: 1.2, 3: 0.4},
        "QUEEN":   {5: 1.5, 4: 0.9, 3: 0.3},
        "JACK":    {5: 1, 4: 0.6, 3: 0.2},
        "SPADE":   {5: 0.5, 4: 0.3, 3: 0.1},
        "HEART":   {5: 0.5, 4: 0.3, 3: 0.1},
        "DIAMOND": {5: 0.25, 4: 0.15, 3: 0.05},
        "CLUB":    {5: 0.25, 4: 0.15, 3: 0.05},
    }
    Payouts2 = map[string]map[int]float64{ // For bet amounts 2
        "ACE":     {5: 5, 4: 3, 3: 1},
        "KING":    {5: 4, 4: 2.4, 3: 0.8},
        "QUEEN":   {5: 3, 4: 1.8, 3: 0.6},
        "JACK":    {5: 2, 4: 1.2, 3: 0.4},
        "SPADE":   {5: 1, 4: 0.6, 3: 0.2},
        "HEART":   {5: 1, 4: 0.6, 3: 0.2},
        "DIAMOND": {5: 0.5, 4: 0.3, 3: 0.1},
        "CLUB":    {5: 0.5, 4: 0.3, 3: 0.1},
    }
    Payouts3 = map[string]map[int]float64{ // For bet amounts 3
        "ACE":     {5: 7.5, 4: 4.5, 3: 1.5},
        "KING":    {5: 6, 4: 3.6, 3: 1.2},
        "QUEEN":   {5: 4.5, 4: 2.7, 3: 0.9},
        "JACK":    {5: 3, 4: 1.8, 3: 0.6},
        "SPADE":   {5: 1.5, 4: 0.9, 3: 0.3},
        "HEART":   {5: 1.5, 4: 0.9, 3: 0.3},
        "DIAMOND": {5: 0.75, 4: 0.45, 3: 0.15},
        "CLUB":    {5: 0.75, 4: 0.45, 3: 0.15},
    }
    Payouts5 = map[string]map[int]float64{ // For bet amounts 5
        "ACE":     {5: 12.5, 4: 7.5, 3: 2.5},
        "KING":    {5: 10, 4: 6, 3: 2},
        "QUEEN":   {5: 7.5, 4: 4.5, 3: 1.5},
        "JACK":    {5: 5, 4: 3, 3: 1},
        "SPADE":   {5: 2.5, 4: 1.5, 3: 0.5},
        "HEART":   {5: 2.5, 4: 1.5, 3: 0.5},
        "DIAMOND": {5: 1.25, 4: 0.75, 3: 0.25},
        "CLUB":    {5: 1.25, 4: 0.75, 3: 0.25},
    }
    Payouts10 = map[string]map[int]float64{ // For bet amounts 10
        "ACE":     {5: 25, 4: 15, 3: 5},
        "KING":    {5: 20, 4: 12, 3: 4},
        "QUEEN":   {5: 15, 4: 9, 3: 3},
        "JACK":    {5: 10, 4: 6, 3: 2},
        "SPADE":   {5: 5, 4: 3, 3: 1},
        "HEART":   {5: 5, 4: 3, 3: 1},
        "DIAMOND": {5: 2.5, 4: 1.5, 3: 0.5},
        "CLUB":    {5: 2.5, 4: 1.5, 3: 0.5},
    }
    Payouts20 = map[string]map[int]float64{ // For bet amounts 20
        "ACE":     {5: 50, 4: 30, 3: 10},
        "KING":    {5: 40, 4: 24, 3: 8},
        "QUEEN":   {5: 30, 4: 18, 3: 6},
        "JACK":    {5: 20, 4: 12, 3: 4},
        "SPADE":   {5: 10, 4: 6, 3: 2},
        "HEART":   {5: 10, 4: 6, 3: 2},
        "DIAMOND": {5: 5, 4: 3, 3: 1},
        "CLUB":    {5: 5, 4: 3, 3: 1},
    }
    Payouts30 = map[string]map[int]float64{ // For bet amounts 30
        "ACE":     {5: 75, 4: 45, 3: 15},
        "KING":    {5: 60, 4: 36, 3: 12},
        "QUEEN":   {5: 45, 4: 27, 3: 9},
        "JACK":    {5: 30, 4: 18, 3: 6},
        "SPADE":   {5: 15, 4: 9, 3: 3},
        "HEART":   {5: 15, 4: 9, 3: 3},
        "DIAMOND": {5: 7.5, 4: 4.5, 3: 1.5},
        "CLUB":    {5: 7.5, 4: 4.5, 3: 1.5},
    }
    Payouts40 = map[string]map[int]float64{ // For bet amounts 40
        "ACE":     {5: 100, 4: 60, 3: 20},
        "KING":    {5: 80, 4: 48, 3: 16},
        "QUEEN":   {5: 60, 4: 36, 3: 12},
        "JACK":    {5: 40, 4: 24, 3: 8},
        "SPADE":   {5: 20, 4: 12, 3: 4},
        "HEART":   {5: 20, 4: 12, 3: 4},
        "DIAMOND": {5: 10, 4: 6, 3: 2},
        "CLUB":    {5: 10, 4: 6, 3: 2},
    }
    Payouts50 = map[string]map[int]float64{ // For bet amounts 50
        "ACE":     {5: 125, 4: 75, 3: 25},
        "KING":    {5: 100, 4: 60, 3: 20},
        "QUEEN":   {5: 75, 4: 45, 3: 15},
        "JACK":    {5: 50, 4: 30, 3: 10},
        "SPADE":   {5: 25, 4: 15, 3: 5},
        "HEART":   {5: 25, 4: 15, 3: 5},
        "DIAMOND": {5: 12.5, 4: 7.5, 3: 2.5},
        "CLUB":    {5: 12.5, 4: 7.5, 3: 2.5},
    }
    Payouts80 = map[string]map[int]float64{ // For bet amounts 80
        "ACE":     {5: 200, 4: 120, 3: 40},
        "KING":    {5: 160, 4: 96, 3: 32},
        "QUEEN":   {5: 120, 4: 72, 3: 24},
        "JACK":    {5: 80, 4: 48, 3: 16},
        "SPADE":   {5: 40, 4: 24, 3: 8},
        "HEART":   {5: 40, 4: 24, 3: 8},
        "DIAMOND": {5: 20, 4: 12, 3: 4},
        "CLUB":    {5: 20, 4: 12, 3: 4},
    }
    Payouts100 = map[string]map[int]float64{ // For bet amounts 100
        "ACE":     {5: 250, 4: 150, 3: 50},
        "KING":    {5: 200, 4: 120, 3: 40},
        "QUEEN":   {5: 150, 4: 90, 3: 30},
        "JACK":    {5: 100, 4: 60, 3: 20},
        "SPADE":   {5: 50, 4: 30, 3: 10},
        "HEART":   {5: 50, 4: 30, 3: 10},
        "DIAMOND": {5: 25, 4: 15, 3: 5},
        "CLUB":    {5: 25, 4: 15, 3: 5},
    }
    Payouts200 = map[string]map[int]float64{ // For bet amounts 200
        "ACE":     {5: 500, 4: 300, 3: 100},
        "KING":    {5: 400, 4: 240, 3: 80},
        "QUEEN":   {5: 300, 4: 180, 3: 60},
        "JACK":    {5: 200, 4: 120, 3: 40},
        "SPADE":   {5: 100, 4: 60, 3: 20},
        "HEART":   {5: 100, 4: 60, 3: 20},
        "DIAMOND": {5: 50, 4: 30, 3: 10},
        "CLUB":    {5: 50, 4: 30, 3: 10},
    }
    Payouts500 = map[string]map[int]float64{ // For bet amounts 500
        "ACE":     {5: 1250, 4: 750, 3: 250},
        "KING":    {5: 1000, 4: 600, 3: 200},
        "QUEEN":   {5: 750, 4: 450, 3: 150},
        "JACK":    {5: 500, 4: 300, 3: 100},
        "SPADE":   {5: 250, 4: 150, 3: 50},
        "HEART":   {5: 250, 4: 150, 3: 50},
        "DIAMOND": {5: 125, 4: 75, 3: 25},
        "CLUB":    {5: 125, 4: 75, 3: 25},
    }
    Payouts1000 = map[string]map[int]float64{ // For bet amounts 1000
        "ACE":     {5: 2500, 4: 1500, 3: 500},
        "KING":    {5: 2000, 4: 1200, 3: 400},
        "QUEEN":   {5: 1500, 4: 900, 3: 300},
        "JACK":    {5: 1000, 4: 600, 3: 200},
        "SPADE":   {5: 500, 4: 300, 3: 100},
        "HEART":   {5: 500, 4: 300, 3: 100},
        "DIAMOND": {5: 250, 4: 150, 3: 50},
        "CLUB":    {5: 250, 4: 150, 3: 50},
    }
    ComboMultipliersNormal = []int{1, 2, 3, 5, 10}
    ComboMultipliersFree   = []int{2, 4, 6, 10, 20}
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

func getRandomSymbol() string {
    totalWeight := 0
    for _, sw := range SymbolWeights {
        totalWeight += sw.Weight
    }
    r := rand.Intn(totalWeight)
    currentWeight := 0
    for _, sw := range SymbolWeights {
        currentWeight += sw.Weight
        if r < currentWeight {
            return sw.Symbol
        }
    }
    return SymbolWeights[len(SymbolWeights)-1].Symbol
}

func getRandomNonScatterSymbol() string {
    totalWeight := 0
    for _, sw := range NonScatterSymbolWeights {
        totalWeight += sw.Weight
    }
    r := rand.Intn(totalWeight)
    currentWeight := 0
    for _, sw := range NonScatterSymbolWeights {
        currentWeight += sw.Weight
        if r < currentWeight {
            return sw.Symbol
        }
    }
    return NonScatterSymbolWeights[len(NonScatterSymbolWeights)-1].Symbol
}

// ValidateBetAmount checks if the bet amount is one of the valid values
func ValidateBetAmount(bet float64) bool {
    for _, validBet := range ValidBetAmounts {
        if bet == validBet {
            return true
        }
    }
    return false
}

// GetPayoutTable selects the appropriate paytable based on the bet amount
func (gs *GameState) GetPayoutTable() map[string]map[int]float64 {
    switch gs.BetAmount {
    case 0.5:
        return Payouts05
    case 1:
        return Payouts1
    case 2:
        return Payouts2
    case 3:
        return Payouts3
    case 5:
        return Payouts5
    case 10:
        return Payouts10
    case 20:
        return Payouts20
    case 30:
        return Payouts30
    case 40:
        return Payouts40
    case 50:
        return Payouts50
    case 80:
        return Payouts80
    case 100:
        return Payouts100
    case 200:
        return Payouts200
    case 500:
        return Payouts500
    case 1000:
        return Payouts1000
    default:
        // Default to Payouts05 if bet amount is invalid (shouldn't happen after validation)
        return Payouts05
    }
}

func NewGameState(bet float64, mode string) GameState {
    initialMultiplier := 1
    if mode == "FREE" {
        initialMultiplier = 2
    }
    fmt.Printf("NewGameState: Mode=%s, ComboMultiplier=%d\n", mode, initialMultiplier)
    return GameState{
        FreeSpins:       0,
        AmountWon:       0,
        ComboMultiplier: initialMultiplier,
        Cards:           [5][4]Card{},
        Mode:            mode,
        BetAmount:       bet,
    }
}

func (gs *GameState) Spin() float64 {
    gs.generateGrid()
    return gs.calculatePotentialWins()
}

func (gs *GameState) Transform() float64 {
    gs.removeTransformed()
    gs.applyGoldenTransformations()
    gs.fillEmptyPositions()
    gs.applyStarCard()
    return gs.calculatePotentialWins()
}

func (gs *GameState) ApplyRNGOutcome(rngResp RNGResponse, action string) {
    gs.clearTransformed()

    potentialWins := gs.calculatePotentialWins()
    payoutMultiplier := potentialWins / gs.BetAmount
    fmt.Printf("Applying RNG outcome - Potential wins: %v, Payout multiplier: %v\n", potentialWins, payoutMultiplier)

    if rngResp.WinAmount == 0 {
        gs.forceLoss()
        gs.AmountWon = 0
        gs.clearTransformed()
    } else {
        gs.AmountWon = rngResp.WinAmount
        if potentialWins > rngResp.WinAmount {
            gs.adjustWins(rngResp.WinAmount)
        }
    }

    if gs.AmountWon > 0 && gs.anyTransformed() {
        if gs.Mode == "NORMAL" && gs.ComboMultiplier < len(ComboMultipliersNormal) {
            gs.ComboMultiplier = ComboMultipliersNormal[gs.ComboMultiplier]
        } else if gs.Mode == "FREE" && gs.ComboMultiplier < len(ComboMultipliersFree) {
            gs.ComboMultiplier = ComboMultipliersFree[gs.ComboMultiplier-1]
        }
    } else {
        gs.ComboMultiplier = 1
        if gs.Mode == "FREE" {
            gs.ComboMultiplier = 2
        }
    }
    if action == "SPIN" {
        scatterCount := gs.countScatters()
        if scatterCount >= 3 {
            if gs.Mode == "NORMAL" {
                gs.FreeSpins = 10
                gs.Mode = "FREE"
                gs.ComboMultiplier = 2
            } else if gs.Mode == "FREE" {
                gs.FreeSpins += 5
            }
        }
    }
}

func (gs *GameState) generateGrid() {
    // Step 1: Pick a random non-Scatter symbol for the guaranteed win
    nonScatterSymbols := []string{"ACE", "KING", "QUEEN", "JACK", "SPADE", "HEART", "DIAMOND", "CLUB"}
    winSymbol := nonScatterSymbols[rand.Intn(len(nonScatterSymbols))]

    // Step 2: Randomly select a starting reel (0, 1, or 2) to ensure 3 adjacent reels
    startReel := rand.Intn(3)

    // Step 3: Place the symbol in 3 adjacent reels, randomizing the row for each
    positions := make(map[[2]int]bool)
    for reel := startReel; reel < startReel+3; reel++ {
        row := rand.Intn(4)
        gs.Cards[reel][row] = Card{Name: winSymbol}
        positions[[2]int{reel, row}] = true
    }

    // Step 4: Fill the remaining positions randomly
    for i := 0; i < 5; i++ {
        for j := 0; j < 4; j++ {
            if positions[[2]int{i, j}] {
                continue
            }
            if gs.Cards[i][j].Name == "" {
                gs.Cards[i][j] = Card{Name: getRandomSymbol()}
            }
        }
    }
    // Debug: Print the grid
    fmt.Println("Generated grid:")
    for j := 0; j < 4; j++ {
        row := ""
        for i := 0; i < 5; i++ {
            row += gs.Cards[i][j].Name + "\t"
        }
        fmt.Println(row)
    }
}

func (gs *GameState) removeTransformed() {
    for i := 0; i < 5; i++ {
        for j := 0; j < 4; j++ {
            if gs.Cards[i][j].Transformed {
                gs.Cards[i][j] = Card{}
            }
        }
    }
}

func (gs *GameState) applyGoldenTransformations() {
    for i := 1; i < 4; i++ {
        for j := 0; j < 4; j++ {
            if gs.Cards[i][j].Golden && gs.Cards[i][j].Transformed {
                if rand.Float32() < 0.5 {
                    gs.Cards[i][j] = Card{Substitute: "BIG_JOKER"}
                    gs.replicateBigJoker(i, j)
                } else {
                    gs.Cards[i][j] = Card{Substitute: "LITTLE_JOKER"}
                }
            }
        }
    }
}

func (gs *GameState) replicateBigJoker(row, col int) {
    replicates := rand.Intn(4) + 1
    for r := 0; r < replicates; r++ {
        i := rand.Intn(4) + 1
        j := rand.Intn(4)
        if gs.Cards[i][j].Substitute == "" && gs.Cards[i][j].Name != "SCATTER" {
            gs.Cards[i][j] = Card{Substitute: "BIG_JOKER"}
        }
    }
}
func (gs *GameState) fillEmptyPositions() {
    // First, find existing symbols in the first reel to potentially match
    firstReelSymbols := []string{}
    for j := 0; j < 4; j++ {
        if gs.Cards[0][j].Name != "" && gs.Cards[0][j].Name != "SCATTER" {
            firstReelSymbols = append(firstReelSymbols, gs.Cards[0][j].Name)
        }
    }
    
    if len(firstReelSymbols) == 0 {
        firstReelSymbols = []string{"ACE", "KING", "QUEEN", "JACK"}
    }
    
    targetSymbol := firstReelSymbols[rand.Intn(len(firstReelSymbols))]
    
    // Fill empty positions with a high chance of matching symbols
    for i := 0; i < 5; i++ {
        for j := 0; j < 4; j++ {
            if gs.Cards[i][j].Name == "" && gs.Cards[i][j].Substitute == "" {
                // First three reels have high chance of matching targetSymbol
                if i < 3 && rand.Float32() < 0.8 {
                    gs.Cards[i][j] = Card{Name: targetSymbol}
                } else {
                    gs.Cards[i][j] = Card{Name: getRandomSymbol()}
                }
            }
        }
    }
}

func (gs *GameState) applyStarCard() {
    cascadeCount := gs.ComboMultiplier - 1
    if (cascadeCount >= 1 && cascadeCount <= 3 && rand.Float32() < 0.3) || cascadeCount >= 4 {
        for i := 1; i < 4; i++ {
            for j := 0; j < 4; j++ {
                if gs.Cards[i][j].Name != "SCATTER" && gs.Cards[i][j].Substitute == "" && rand.Float32() < 0.2 {
                    gs.Cards[i][j].Golden = true
                }
            }
        }
    }
}

func (gs *GameState) calculatePotentialWins() float64 {
    var total float64
    ways := 0

    // Select the appropriate paytable based on the bet amount
    payouts := gs.GetPayoutTable()

    for r1 := 0; r1 < 4; r1++ {
        symbol := gs.Cards[0][r1].Name
        if symbol == "SCATTER" {
            continue
        }
        if gs.Cards[0][r1].Substitute != "" {
            symbol = "WILD"
        }

        positions := make([][]int, 5)
        positions[0] = []int{r1}

        for reel := 1; reel < 5; reel++ {
            positions[reel] = []int{}
            for r := 0; r < 4; r++ {
                if gs.Cards[reel][r].Name == symbol || gs.Cards[reel][r].Substitute != "" {
                    positions[reel] = append(positions[reel], r)
                }
            }
            if len(positions[reel]) == 0 {
                break
            }
        }

        count := 0
        for reel := 0; reel < 5; reel++ {
            if len(positions[reel]) > 0 {
                count++
            } else {
                break
            }
        }

        if count >= 3 && symbol != "WILD" {
            wayCount := 1
            for reel := 0; reel < count; reel++ {
                wayCount *= len(positions[reel])
            }
            ways += wayCount
            if payout, ok := payouts[symbol][count]; ok {
                fmt.Printf("Variables to calculate total: Symbol=%s, Count=%d, Ways=%d, gs.combomultiplier=%v, Payout=%v\n", symbol, count, wayCount, gs.ComboMultiplier, payout)
                winAmount := payout * float64(wayCount) * float64(gs.ComboMultiplier)
                total += winAmount
                gs.markTransformed(symbol, positions, count)
                fmt.Printf("Detected win: Symbol=%s, Count=%d, Ways=%d, Payout=%v, Total=%v\n", symbol, count, wayCount, payout, total)
            }
        }
    }
    fmt.Printf("Final total: %v\n", total)
    return total
}

func (gs *GameState) markTransformed(symbol string, positions [][]int, count int) {
    for reel := 0; reel < count; reel++ {
        for _, row := range positions[reel] {
            if gs.Cards[reel][row].Name == symbol || gs.Cards[reel][row].Substitute != "" {
                gs.Cards[reel][row].Transformed = true
            }
        }
    }
}

func (gs *GameState) forceLoss() {
    for i := 0; i < 5; i++ {
        for j := 0; j < 4; j++ {
            if gs.Cards[i][j].Transformed {
                newSymbol := gs.selectNonMatchingSymbol(i, j)
                gs.Cards[i][j] = Card{Name: newSymbol, Transformed: false}
            }
        }
    }
    gs.clearTransformed()
    potentialWins := gs.calculatePotentialWins()
    if potentialWins > 0 {
        gs.forceLoss()
    }
}

func (gs *GameState) selectNonMatchingSymbol(col, row int) string {
    newSymbol := getRandomNonScatterSymbol()
    attempts := 0
    maxAttempts := 10

    for attempts < maxAttempts {
        matches := false
        for k := 0; k < 5; k++ {
            if k != col {
                for l := 0; l < 4; l++ {
                    if (gs.Cards[k][l].Name == newSymbol || gs.Cards[k][l].Substitute != "") && abs(k-col) <= 1 {
                        matches = true
                        break
                    }
                }
            }
            if matches {
                break
            }
        }
        if !matches {
            return newSymbol
        }
        newSymbol = getRandomNonScatterSymbol()
        attempts++
    }
    return getRandomNonScatterSymbol()
}

func (gs *GameState) adjustWins(target float64) {
    current := gs.calculatePotentialWins()
    attempts := 0
    maxAttempts := 20

    for current > target && attempts < maxAttempts {
        transformedPositions := [][2]int{}
        for i := 0; i < 5; i++ {
            for j := 0; j < 4; j++ {
                if gs.Cards[i][j].Transformed {
                    transformedPositions = append(transformedPositions, [2]int{i, j})
                }
            }
        }

        if len(transformedPositions) == 0 {
            break
        }

        pos := transformedPositions[rand.Intn(len(transformedPositions))]
        i, j := pos[0], pos[1]
        newSymbol := gs.selectNonMatchingSymbol(i, j)
        gs.Cards[i][j] = Card{Name: newSymbol, Transformed: false}

        gs.clearTransformed()
        current = gs.calculatePotentialWins()
        attempts++
    }

    if current > target {
        gs.forceLoss()
        gs.AmountWon = 0
    }
}

func (gs *GameState) clearTransformed() {
    for i := 0; i < 5; i++ {
        for j := 0; j < 4; j++ {
            gs.Cards[i][j].Transformed = false
        }
    }
}

func (gs *GameState) anyTransformed() bool {
    for i := 0; i < 5; i++ {
        for j := 0; j < 4; j++ {
            if gs.Cards[i][j].Transformed {
                return true
            }
        }
    }
    return false
}

func (gs *GameState) countScatters() int {
    count := 0
    for i := 0; i < 5; i++ {
        for j := 0; j < 4; j++ {
            if gs.Cards[i][j].Name == "SCATTER" {
                count++
            }
        }
    }
    return count
}

func abs(x int) int {
    if x < 0 {
        return -x
    }
    return x
}