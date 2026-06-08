# Crazy King Kong Game

A boulder-crushing multiplier game featuring King Kong where players select boulders for Kong to crush. Each crush attempt is a bet - if the boulder doesn't break, the player loses their bet and Kong tries again. When the boulder finally breaks, the player wins based on the revealed multiplier.

## Game Overview

Crazy King Kong is a progressive crushing game where:
1. **Player selects boulder type** and places bet
2. **Kong attempts to crush** the selected boulder
3. **Each crush attempt = 1 bet** (player loses bet if boulder doesn't break)
4. **Boulder breaks = Win** (multiplier revealed, proceed to next boulder)
5. **Boulder doesn't break = Loss** (try crushing same boulder again)
6. **Bonus rounds** randomly triggered when boulders break

## Game Mechanics

### Core Crushing System
- **RNG Determines Breaking**: External RNG service decides if boulder breaks
- **Progressive Attempts**: Same boulder until it breaks, each attempt costs a bet
- **No State Management**: Backend is stateless, Unity handles boulder progression
- **Win Only on Break**: Multiplier only revealed and paid when boulder breaks

### Boulder Breaking Logic
- **RNG "Win"** → Boulder breaks → Reveal multiplier → Player wins
- **RNG "Loss"** → Boulder doesn't break → Player loses bet → Try again

## Boulder Types & Multipliers

| Boulder Type | Multiplier Range | Risk Level | Description |
|-------------|------------------|------------|-------------|
| **Gold Boulder** | 1.5x ~ 100x | Highest | Hardest to break, highest rewards |
| **Blue Boulder** | 1.4x ~ 50x | High | Good balance of difficulty/reward |
| **Red Boulder** | 1.3x ~ 20x | Medium | Moderate difficulty and rewards |
| **White Boulder** | 1.2x ~ 10x | Lowest | Easiest to break, lower rewards |

*Note: Breaking difficulty is controlled by RNG service based on player RTP, not fixed probabilities*

## Bonus Game Features

### Trigger Conditions
- **Triggers when boulder breaks** (RNG win outcome)
- **10% chance** on successful boulder crush
- **Stone selection** from 3 randomly presented options
- **Bet amount locked** during bonus round

### Bonus Stone Types & Multipliers

| Stone Type | Multiplier Range | Description |
|-----------|------------------|-------------|
| **Gold Stone** | 28x ~ 888x | Variable high multipliers |
| **Silver Stone** | 18x (Fixed) | Consistent medium multiplier |
| **Bronze Stone** | 8x (Fixed) | Guaranteed lower multiplier |

## Betting Options

**Available Bet Amounts**: 0.5, 1, 2, 4, 5, 10, 20, 25, 50, 100 credits

**Payout Calculation**: `Win Amount = Bet Amount × Revealed Multiplier`

## API Endpoints

### Main Game - Boulder Crushing

**Endpoint**: `POST /crush/crazykingkong`

#### Request Body
```json
{
  "client_id": "1",
  "game_id": "45",
  "player_id": "22",
  "bet_id": "unique_per_crush_attempt",
  "bet_amount": 5.0,
  "boulder_type": "gold"
}
```

#### Response - Boulder Doesn't Break (RNG Loss)
```json
{
  "status": "success",
  "message": "Boulder didn't break, try again!",
  "boulder_type": "gold",
  "boulder_broken": false,
  "multiplier": 0,
  "win_amount": 0,
  "bonus_triggered": false
}
```

#### Response - Boulder Breaks (RNG Win)
```json
{
  "status": "success",
  "message": "Boulder crushed! Choose next boulder.",
  "boulder_type": "gold",
  "boulder_broken": true,
  "multiplier": 25.6,
  "win_amount": 128.0,
  "bonus_triggered": false
}
```

#### Response - Boulder Breaks + Bonus Triggered
```json
{
  "status": "success",
  "message": "Boulder crushed! Bonus game triggered - choose a stone!",
  "boulder_type": "gold", 
  "boulder_broken": true,
  "multiplier": 25.6,
  "win_amount": 128.0,
  "bonus_triggered": true,
  "available_stones": [
    {
      "type": "gold",
      "display_name": "gold stone"
    },
    {
      "type": "silver",
      "display_name": "silver stone"
    },
    {
      "type": "bronze",
      "display_name": "bronze stone"
    }
  ]
}
```

### Bonus Game - Stone Selection

**Endpoint**: `POST /bonus/crazykingkong`

#### Request Body
```json
{
  "client_id": "1",
  "game_id": "45",
  "player_id": "22",
  "bet_id": "bonus_unique_id",
  "bet_amount": 5.0,
  "stone_type": "gold"
}
```

#### Response Body
```json
{
  "status": "success",
  "message": "Bonus stone revealed!",
  "stone_type": "gold",
  "multiplier": 156.0,
  "win_amount": 780.0
}
```

## Game Flow

### Standard Crushing Flow
1. **Player Selection**: Choose boulder type and bet amount in Unity
2. **Crush Attempt**: Unity calls `/crush/crazykingkong` API
3. **RNG Decision**: Backend calls RNG service to determine outcome
4. **Boulder Doesn't Break**: 
   - Unity shows failed crush animation
   - Player loses bet amount
   - Same boulder remains for next attempt
5. **Boulder Breaks**:
   - Unity shows successful crush animation
   - Multiplier revealed and paid
   - Unity moves to next boulder selection
   - Possible bonus game trigger

### Bonus Game Flow
1. **Bonus Trigger**: When boulder breaks, 10% chance for bonus
2. **Stone Presentation**: Unity displays 3 available stones
3. **Player Choice**: Select one stone type
4. **Stone Reveal**: Backend processes bonus with RNG validation
5. **Bonus Payout**: Additional win amount based on stone multiplier

### Client (Unity) Responsibilities
- **Boulder State**: Track which boulder is currently being crushed
- **Attempt Counter**: Count crush attempts for UI feedback
- **Animation Management**: Show crush attempts, successes, failures
- **Boulder Progression**: Move to next boulder after successful break
- **Bonus Flow**: Handle bonus stone selection and animations

### Backend Responsibilities
- **Stateless Processing**: Each API call is independent
- **RNG Integration**: Determine break/no-break outcomes
- **Multiplier Generation**: Create fair weighted multipliers
- **Bonus Triggers**: Random bonus game activation
- **Validation**: Ensure all inputs are valid

## Technical Implementation

### Breaking Determination
```go
// Generate potential multiplier for boulder type
multiplier := GenerateBoulderMultiplier(boulderType)
potentialWin := betAmount * multiplier

// Let RNG service decide if boulder breaks
rngResponse := rngClient.GetOutcome(...)

if rngResponse.PrefOutcome == "win" {
    // Boulder breaks - reveal multiplier and pay
    boulderBroken = true
    winAmount = potentialWin
} else {
    // Boulder doesn't break - player loses bet
    boulderBroken = false
    winAmount = 0
}
```

### Multiplier Generation Algorithm
**Boulder Multipliers**: Uses exponential distribution favoring lower multipliers
```go
exponentialValue := 1.0 - (randomValue * randomValue * randomValue)
multiplier := min + (exponentialValue * (max - min))
```

**Stone Multipliers**: 
- Gold stones: Weighted distribution (28x-888x)
- Silver/Bronze: Fixed multipliers (18x/8x)

### RNG Integration Flow
1. **Generate Potential Outcome**: Calculate what player could win
2. **RNG Validation**: External service determines actual outcome based on RTP
3. **Apply Result**: 
   - **Win**: Use generated multiplier, boulder breaks
   - **Loss**: No multiplier, boulder doesn't break

### Error Handling & Validation
- **Request Validation**: Required fields, valid bet amounts, valid boulder/stone types
- **RNG Service Integration**: Graceful handling of external service failures
- **Settings Service**: Player RTP retrieval with retry logic
- **Comprehensive Logging**: All outcomes logged for monitoring

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
LOG_FILE=crazykingkong.log
```

### Client Selection Logic
- **Production Environment**: Default for all requests
- **Test Environment**: Activated when Origin header contains "test"

## File Structure

```
cmd/crazykingkong/
├── main.go                 # Main application entry point

pkg/games/crazykingkong/
├── types.go               # Request/response structures
├── game.go                # Core game logic and multiplier generation
├── handlers.go            # HTTP handlers for crush and bonus endpoints
├── routes.go              # Route registration and client selection
└── README.md              # This documentation

pkg/common/
├── config/config.go       # Environment configuration (shared)
├── rng/client.go          # RNG service client (shared)
└── settings/client.go     # Settings service client (shared)
```

## Game Balance Configuration

### Bonus Trigger Rate
**Current**: 10% chance when boulder breaks
**Adjustment**: Modify `ShouldTriggerBonus()` function

### Multiplier Distribution
**Current**: Exponential curves favor lower multipliers
**Adjustment**: Modify exponential power in generation functions
- **More Conservative**: Increase power (favor lower multipliers)
- **More Aggressive**: Decrease power (favor higher multipliers)

### Boulder Risk Profiles
**Current**: Gold hardest/highest, White easiest/lowest
**Adjustment**: Modify ranges in `BoulderMultipliers` map

## Development & Testing

### Running Locally
```bash
# Set environment variables
cp .env.example .env

# Run the application
go run cmd/crazykingkong/main.go
```

### Testing Boulder Crushing
```bash
# Test successful boulder break
curl -X POST http://localhost:11401/crush/crazykingkong \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "1",
    "game_id": "45",
    "player_id": "test123",
    "bet_id": "bet123",
    "bet_amount": 1.0,
    "boulder_type": "gold"
  }'

# Test bonus game
curl -X POST http://localhost:11401/bonus/crazykingkong \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "1", 
    "game_id": "45",
    "player_id": "test123",
    "bet_id": "bonus123",
    "bet_amount": 1.0,
    "stone_type": "gold"
  }'
```

### Integration Notes
- **Stateless Design**: No session management required
- **Unity State Management**: Client handles all boulder progression
- **Consistent Architecture**: Same structure as Kong slot game
- **Shared Services**: Reuses RNG and Settings service clients
- **Environment Separation**: Production and test environment support

## Expected Player Experience

### Typical Game Session
1. **Boulder Selection**: Player chooses gold boulder, bets 5.0 credits
2. **First Crush**: Kong swings, boulder doesn't break → Lose 5.0 credits
3. **Second Crush**: Kong swings, boulder doesn't break → Lose 5.0 credits  
4. **Third Crush**: Kong swings, boulder breaks → Win 5.0 × 15.6 = 78.0 credits
5. **Next Boulder**: Player selects new boulder type, process repeats
6. **Bonus Trigger**: Occasionally get bonus stone selection for extra wins

### Strategic Considerations
- **Boulder Type Selection**: Balance risk vs reward based on budget
- **Bet Management**: Consider multiple attempts may be needed per boulder
- **Bonus Opportunities**: Higher multiplier boulders may trigger more bonuses

## Monitoring & Analytics

### Key Metrics to Track
- **Break Rate per Boulder Type**: Actual vs expected breaking frequency
- **Average Attempts per Break**: How many crushes typically needed
- **Multiplier Distribution**: Verify weighted randomization working correctly
- **Bonus Trigger Rate**: Confirm 10% trigger rate in practice
- **Player Behavior**: Boulder type preferences and bet patterns
- **Win/Loss Ratios**: Monitor against expected RTP values

### Important Logs
- **Crush Attempts**: Every boulder crush attempt with outcome
- **RNG Responses**: All RNG service calls and responses
- **Bonus Triggers**: When and how bonus games activate
- **Error Conditions**: API failures and validation errors
- **Player Patterns**: Betting behavior and boulder selection trends