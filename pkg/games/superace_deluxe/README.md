# Super Ace Deluxe Game

A slot machine-style game featuring a 5x4 grid of symbols with cascading wins, multipliers, and free spins. The game offers multiple betting tiers with corresponding payout tables and special features like transformations, golden symbols, and joker substitutions.

## Game Features

- **5x4 Symbol Grid**: 5 reels with 4 symbols each
- **Two Game Modes**: 
  - NORMAL: Regular play
  - FREE: Free spins mode with higher multipliers
- **Cascading Wins**: Winning symbols transform and are replaced with new symbols
- **Progressive Multipliers**: Multipliers increase with consecutive wins
- **Special Symbols**:
  - SCATTER: 3+ triggers free spins
  - Golden Symbols: Can transform into special jokers
  - BIG_JOKER: Substitutes for any symbol and can replicate
  - LITTLE_JOKER: Substitutes for any symbol

## Symbols

| Symbol | Description | 3 of a Kind | 4 of a Kind | 5 of a Kind |
|--------|-------------|-------------|-------------|-------------|
| ACE | Highest value symbol | 0.25-500* | 0.75-1500* | 1.25-2500* |
| KING | High value symbol | 0.2-400* | 0.6-1200* | 1-2000* |
| QUEEN | Medium-high value symbol | 0.15-300* | 0.45-900* | 0.75-1500* |
| JACK | Medium value symbol | 0.1-200* | 0.3-600* | 0.5-1000* |
| SPADE | Medium-low value symbol | 0.05-100* | 0.15-300* | 0.25-500* |
| HEART | Medium-low value symbol | 0.05-100* | 0.15-300* | 0.25-500* |
| DIAMOND | Low value symbol | 0.03-50* | 0.08-150* | 0.13-250* |
| CLUB | Low value symbol | 0.03-50* | 0.08-150* | 0.13-250* |
| SCATTER | Triggers free spins | - | - | - |

*Payouts depend on bet amount (0.5-1000)

## Special Features

### Cascading Wins
- Winning combinations are removed
- New symbols cascade down
- Consecutive cascades increase multipliers

### Multiplier Progression
- **Normal Mode**: 1x → 2x → 3x → 5x → 10x
- **Free Spins Mode**: 2x → 4x → 6x → 10x → 20x

### Free Spins
- **Trigger**: 3+ SCATTER symbols
- **Initial Award**: 10 free spins
- **Retrigger**: 5 additional free spins
- **Enhanced Multipliers**: Higher multipliers during free spins
- **Starting Multiplier**: Free spins start at 2x multiplier

### Golden Transformations
- Golden symbols appear after consecutive wins
- When part of a win, golden symbols transform into jokers:
  - 50% chance of BIG_JOKER
  - 50% chance of LITTLE_JOKER
- BIG_JOKER replicates to 1-4 random positions

## Betting Options

Multiple betting levels, each with its own payout table:
- 0.5, 1, 2, 3, 5, 10, 20, 30, 40, 50, 80, 100, 200, 500, 1000

## API Endpoints

### Spin (`POST /spin/superace/deluxe`)

#### SPIN Action Request
```json
{
    "game": {
        "id": "31",
        "name": "SUPER_ACE DELUXE",
        "mode": "NORMAL"
    },
    "betAmount": 20.0,
    "clientId": "1",
    "playerId": "22",
    "action": "SPIN"
}
```

#### TRANSFORM Action Request
```json
{
    "game": {
        "id": "31",
        "name": "SUPER_ACE DELUXE",
        "mode": "NORMAL"
    },
    "betAmount": 20.0,
    "clientId": "1",
    "playerId": "22",
    "action": "TRANSFORM",
    "cards": [...],
    "comboMultiplier": 2,
    "freeSpins": 0
}
```

#### Response
```json
{
  "status": 200,
  "message": "Success",
  "data": {
    "freeSpins": 0,
    "amountWon": 25,
    "comboMultiplier": 2,
    "cards": [...],
    "mode": "NORMAL",
    "betAmount": 20.0
  }
}
```

## Game Flow

### SPIN Action
1. Generates a new grid with at least one guaranteed potential win
2. Calculates potential wins
3. Requests RNG outcome
4. Applies RNG decision (win or loss)
5. Returns updated game state

### TRANSFORM Action (after a win)
1. Removes transformed symbols
2. Replaces with new symbols (high chance of creating winning patterns)
3. Applies golden transformations
4. Calculates potential wins
5. Requests RNG outcome
6. Applies RNG decision
7. Returns updated game state

## Game Logic Details

### Grid Generation
- Every spin guarantees at least one potential winning combination
- Win combinations start from the leftmost reel
- Minimum 3 matching symbols required for a win

### Win Calculation
1. Win combinations are identified based on adjacent reels
2. Payouts are based on symbol type, number of matches, and bet amount
3. Symbols are marked as transformed when part of a win
4. Cumulative wins are tracked across cascades

### RNG Integration
The game integrates with an external RNG service to ensure fair outcomes based on:
- Player RTP settings
- Potential win amounts
- Bet size

## Integration with External Services

The game integrates with two external services:

1. **RTP Service**: Provides Return-to-Player settings for each player
2. **RNG Service**: Determines win outcomes based on potential wins and RTP

## Running as Standalone Service


## Environmental Configuration

Configuration via `.env` file:
```
RNG_API_URL=http://159.89.235.166:17003/api/proxy/rng/1
SETTINGS_API_URL=https://t2.ibibe.africa/get-game-settings
PORT=8080
LOG_FILE=app.log
```

## Code Organization

```
pkg/games/superace_deluxe/
├── game.go      # Core game logic
├── handlers.go  # API handlers
├── routes.go    # Route definitions
├── types.go     # Game-specific types
└── utils.go     # Helper functions
```