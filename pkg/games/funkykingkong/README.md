# Funky King Kong Slot Machine Game

A tropical-themed slot machine game featuring King Kong where players spin three reels to match symbols and win based on the selected paytable level. The game features a unique "Bet One" button that cycles through three different paytable levels (x1, x2, x3) with increasing payouts and corresponding bet amounts.

## Game Overview

Funky King Kong is a classic 3-reel slot machine where:
1. **Player selects bet level** (x1, x2, or x3) using the "Bet One" button
2. **Player chooses bet amount** from available options for the selected level
3. **Player spins the reels** to match symbols across the payline
4. **Winning combinations** pay out based on the selected paytable level
5. **RNG determines** whether the spin results in a win or loss
6. **Single payline** with adjacent reel matching from left to right

## Game Mechanics

### Core Slot System
- **3-Reel Layout**: Three reels with various tropical symbols
- **Single Payline**: Center horizontal line for winning combinations
- **Adjacent Matching**: Symbols must match on adjacent reels starting from leftmost
- **RNG Determines Outcome**: External RNG service decides win/loss
- **Stateless Design**: Each spin is independent, no session management

### Bet Level Cycling
- **"Bet One" Button**: Cycles through three paytable levels
- **Level Progression**: x1 → x2 → x3 → x1 (cycles back to x1)
- **Increasing Payouts**: Higher levels offer better payouts for same symbols
- **Corresponding Bet Amounts**: Each level has specific valid bet amounts

### Winning Logic
- **RNG "Win"** → Generate winning symbol combination → Pay out
- **RNG "Loss"** → Generate losing symbol combination → No payout

## Symbol Types & Payouts

### Paytable Structure (3 Levels)
Each symbol combination has three payout values corresponding to the three bet levels:

| Symbol Combination | Level x1 | Level x2 | Level x3 | Description |
|-------------------|----------|----------|----------|-------------|
| **Kong Kong Kong** | 800 | 1600 | 2500 | Three King Kong symbols |
| **Sun Sun Sun** | 400 | 800 | 1200 | Three sun symbols |
| **Palm Palm Palm** | 200 | 400 | 600 | Three palm tree symbols |
| **Coconut Coconut Coconut** | 100 | 200 | 300 | Three coconut symbols |
| **Banana Banana Banana** | 80 | 160 | 240 | Three banana symbols |
| **3BAR 3BAR 3BAR** | 60 | 120 | 180 | Three 3BAR symbols |
| **2BAR 2BAR 2BAR** | 40 | 80 | 120 | Three 2BAR symbols |
| **1BAR 1BAR 1BAR** | 20 | 40 | 60 | Three 1BAR symbols |
| **ANY 3X BAR** | 10 | 20 | 30 | Any combination of three BAR symbols |

### Symbol Types
- **High Value**: Kong (highest paying symbol)
- **Medium Value**: Sun, Palm, Coconut, Banana
- **Low Value**: 3BAR, 2BAR, 1BAR
- **Special**: ANY 3X BAR (any mix of BAR symbols)

## Betting Options

### Bet Level 1 (x1)
**Available Bet Amounts**: 0.01, 0.05, 0.1, 0.2, 0.25 credits
**Internal Multipliers**: 1, 5, 10, 20, 25

### Bet Level 2 (x2)
**Available Bet Amounts**: 0.02, 0.1, 0.2, 0.4, 0.5 credits
**Internal Multipliers**: 1, 5, 10, 20, 25

### Bet Level 3 (x3)
**Available Bet Amounts**: 0.03, 0.15, 0.3, 0.6, 0.75 credits
**Internal Multipliers**: 1, 5, 10, 20, 25

**Payout Calculation**: `Win Amount = (Paytable Value × Internal Multiplier) × 0.01`

## API Endpoints

### Main Game - Slot Spin

**Endpoint**: `POST /spin/funkykingkong`

#### Request Body
```json
{
  "client_id": "1",
  "game_id": "funkykingkong",
  "player_id": "22",
  "bet_id": "unique_per_spin",
  "bet_amount": 0.1,
  "bet_level": 1
}
```

#### Response - Winning Spin
```json
{
  "status": "success",
  "message": "",
  "reels": ["Kong", "Kong", "Kong"],
  "win_amount": 8.0,
  "winning_combination": "Kong Kong Kong",
  "paytable_used": 1,
  "bet_level": 1
}
```

#### Response - Losing Spin
```json
{
  "status": "success",
  "message": "",
  "reels": ["Kong", "Sun", "Palm"],
  "win_amount": 0.0,
  "winning_combination": "",
  "paytable_used": 1,
  "bet_level": 1
}
```

#### Response - ANY 3X BAR Win
```json
{
  "status": "success",
  "message": "",
  "reels": ["1BAR", "2BAR", "3BAR"],
  "win_amount": 0.1,
  "winning_combination": "ANY 3X BAR",
  "paytable_used": 1,
  "bet_level": 1
}
```

## Game Flow

### Standard Spin Flow
1. **Bet Level Selection**: Player clicks "Bet One" to cycle through x1, x2, x3
2. **Bet Amount Selection**: Player chooses valid bet amount for current level
3. **Spin Initiation**: Unity calls `/spin/funkykingkong` API
4. **RNG Decision**: Backend calls RNG service to determine outcome
5. **Reel Generation**: 
   - **Win**: Generate winning symbol combination
   - **Loss**: Generate losing symbol combination
6. **Result Display**: Unity shows final reel positions and payout

### Client (Unity) Responsibilities
- **Bet Level Cycling**: Handle "Bet One" button to cycle through paytables
- **Bet Amount Validation**: Show only valid amounts for current level
- **Reel Animation**: Display spinning animation and final positions
- **Win/Loss Display**: Show appropriate animations and payouts
- **Paytable Display**: Update visible paytable based on current level

### Backend Responsibilities
- **Stateless Processing**: Each spin is independent
- **RNG Integration**: Determine win/loss outcomes based on RTP
- **Symbol Generation**: Create appropriate winning or losing combinations
- **Validation**: Ensure bet amounts match selected level
- **Payout Calculation**: Calculate correct win amounts

## Technical Implementation

### Win Determination
```go
// Generate potential winning combination
winningReels := GenerateWinningReels()
potentialWin, winCombination := CalculateWin(winningReels, betLevel, internalMultiplier)

// Calculate payout multiplier for RNG
payoutMultiplier := potentialWin / betAmount

// Let RNG service decide outcome
rngResponse := rngClient.GetOutcome(...)

if rngResponse.PrefOutcome == "win" {
    // Use winning combination
    finalReels = winningReels
    finalWinAmount = potentialWin
} else {
    // Use losing combination
    finalReels = GenerateLosingReels()
    finalWinAmount = 0
}
```

### Symbol Generation Algorithm
**Winning Combinations**: Pre-defined winning symbol sets
- Exact matches: Three identical symbols
- ANY 3X BAR: Any combination of three different BAR symbols

**Losing Combinations**: Various non-winning combinations
- Different symbols: No matching symbols
- Partial matches: Two same + one different
- Empty positions: Symbols with empty spaces

### RNG Integration Flow
1. **Generate Potential Outcome**: Calculate what player could win
2. **RNG Validation**: External service determines actual outcome based on RTP
3. **Apply Result**: 
   - **Win**: Use generated winning combination
   - **Loss**: Use generated losing combination

### Bet Level Validation
```go
// Validate bet level (1, 2, or 3)
if !ValidateBetLevel(req.BetLevel) {
    return error("Invalid bet level")
}

// Validate bet amount for selected level
if !ValidateBetAmount(req.BetAmount, req.BetLevel) {
    return error("Invalid bet amount for level")
}
```

## Configuration

### Environment Variables
```env
# Production Configuration
PROD_RNG_API_URL=http://159.89.235.166:17003/api/proxy/rng/1
PROD_SETTINGS_API_URL=https://t2.ibibe.africa/get-game-settings

# Test Configuration  
TEST_RNG_API_URL=https://rng2.ibibe.africa/api/proxy/rng/1
TEST_SETTINGS_API_URL=https://t3.ibibe.africa/get-game-settings

# Server Configuration
PORT=11401
LOG_FILE=funkykingkong.log
```

### Client Selection Logic
- **Production Environment**: Default for all requests
- **Test Environment**: Activated when Origin header contains "test"

## File Structure

```
cmd/funkykingkong/
├── main.go                 # Main application entry point

pkg/games/funkykingkong/
├── types.go               # Request/response structures
├── game.go                # Core game logic and symbol generation
├── handlers.go            # HTTP handlers for spin endpoint
├── routes.go              # Route registration and client selection
└── utils.go               # Utility functions

pkg/common/
├── config/config.go       # Environment configuration (shared)
├── rng/client.go          # RNG service client (shared)
└── settings/client.go     # Settings service client (shared)
```

## Game Balance Configuration

### Paytable Adjustments
**Current**: Three-tier system with x1, x2, x3 multipliers
**Modification**: Edit `Paytable` map in `game.go`

### Bet Amount Ranges
**Current**: Scaled amounts for each level
**Modification**: Edit `BetAmountToMultiplier` map

### Symbol Distribution
**Current**: Balanced mix of high/medium/low value symbols
**Modification**: Adjust `GenerateWinningReels()` and `GenerateLosingReels()`

## Development & Testing

### Running Locally
```bash
# Set environment variables
cp .env.example .env

# Run the application
go run cmd/funkykingkong/main.go
```

### Testing Slot Spins
```bash
# Test winning spin with level 1
curl -X POST http://localhost:11401/spin/funkykingkong \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "1",
    "game_id": "funkykingkong",
    "player_id": "test123",
    "bet_id": "spin123",
    "bet_amount": 0.1,
    "bet_level": 1
  }'

# Test level 2 spin
curl -X POST http://localhost:11401/spin/funkykingkong \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "1",
    "game_id": "funkykingkong", 
    "player_id": "test123",
    "bet_id": "spin124",
    "bet_amount": 0.2,
    "bet_level": 2
  }'

# Test level 3 spin
curl -X POST http://localhost:11401/spin/funkykingkong \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "1",
    "game_id": "funkykingkong",
    "player_id": "test123", 
    "bet_id": "spin125",
    "bet_amount": 0.3,
    "bet_level": 3
  }'
```

### Integration Notes
- **Stateless Design**: No session management required
- **Unity State Management**: Client handles bet level cycling
- **Consistent Architecture**: Same structure as other games
- **Shared Services**: Reuses RNG and Settings service clients
- **Environment Separation**: Production and test environment support

## Expected Player Experience

### Typical Game Session
1. **Level Selection**: Player clicks "Bet One" to select x1 level
2. **Bet Amount**: Player chooses 0.1 credits
3. **First Spin**: Reels spin, no win → Lose 0.1 credits
4. **Second Spin**: Reels spin, Kong Kong Kong → Win 8.0 credits
5. **Level Change**: Player clicks "Bet One" to switch to x2 level
6. **Higher Bet**: Player chooses 0.2 credits for x2 level
7. **Third Spin**: Reels spin, Sun Sun Sun → Win 16.0 credits

### Strategic Considerations
- **Level Selection**: Higher levels offer better payouts but require larger bets
- **Bet Management**: Balance risk vs reward based on bankroll
- **Symbol Recognition**: Learn which combinations pay the most
- **ANY 3X BAR**: Special combination for mixed BAR symbols

## Monitoring & Analytics

### Key Metrics to Track
- **Win Rate per Level**: Actual vs expected win frequency
- **Bet Level Distribution**: Which levels players prefer
- **Symbol Distribution**: Verify random generation working correctly
- **Average Bet Amounts**: Player betting patterns per level
- **RTP Verification**: Monitor against expected return-to-player values
- **Session Length**: How long players stay engaged

### Important Logs
- **Spin Attempts**: Every spin with bet level and amount
- **RNG Responses**: All RNG service calls and outcomes
- **Win/Loss Patterns**: Symbol combinations and payouts
- **Error Conditions**: API failures and validation errors
- **Player Behavior**: Bet level preferences and betting patterns

## Game Rules Summary

### Core Rules
- **Single Payline**: Center horizontal line only
- **Left to Right**: Symbols must match starting from leftmost reel
- **Adjacent Reels**: Winning combinations must be on adjacent reels
- **Exact Matches**: Three identical symbols for standard wins
- **ANY 3X BAR**: Special rule for mixed BAR symbol combinations

### Bet Level System
- **Three Levels**: x1, x2, x3 paytables
- **Cycling**: "Bet One" button cycles through levels
- **Bet Amounts**: Each level has specific valid bet amounts
- **Payout Scaling**: Higher levels offer better payouts

### Winning Combinations
- **Kong Kong Kong**: Highest paying combination (800/1600/2500)
- **Sun Sun Sun**: High value combination (400/800/1200)
- **Palm Palm Palm**: Medium-high value (200/400/600)
- **Coconut Coconut Coconut**: Medium value (100/200/300)
- **Banana Banana Banana**: Medium-low value (80/160/240)
- **BAR Combinations**: Low value combinations (20-180)
- **ANY 3X BAR**: Special mixed BAR rule (10/20/30)