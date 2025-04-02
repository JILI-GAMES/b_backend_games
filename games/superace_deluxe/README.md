# Super Ace Deluxe Game

## Overview

Super Ace Deluxe is a slot machine-style game featuring a 5x4 grid of symbols with cascading wins, multipliers, and free spins. The game offers multiple betting tiers with corresponding payout tables and special features like transformations, golden symbols, and joker substitutions.

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
- **Multiple Bet Levels**: From 0.5 to 1000 with corresponding payout tables

## Technical Implementation

### API Endpoints

- `POST /spin/superace/deluxe`: Main endpoint for both SPIN and TRANSFORM actions

### Request Format

#### Normal Spin
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

#### Transform
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

### Response Format

#### Normal Spin
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

#### Transform
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

### Game Flow

1. **SPIN Action**:
   - Generates a new grid with at least one guaranteed potential win
   - Calculates potential wins
   - Requests RNG outcome
   - Applies RNG decision (win or loss)
   - Returns updated game state

2. **TRANSFORM Action** (after a win):
   - Removes transformed symbols
   - Replaces with new symbols (high chance of creating winning patterns)
   - Applies golden transformations
   - Calculates potential wins
   - Requests RNG outcome
   - Applies RNG decision
   - Returns updated game state

### Symbol Values (Highest to Lowest)

1. ACE
2. KING
3. QUEEN
4. JACK
5. SPADE
6. HEART
7. DIAMOND
8. CLUB
9. SCATTER (special symbol)

## Integration with External Services

The game integrates with two external services:

1. **RTP Service**: Provides Return-to-Player settings for each player
2. **RNG Service**: Determines win outcomes based on potential wins and RTP

## Code Structure

```
games/superace/
├── routes.go    # Game route definitions
├── handlers.go  # API handlers
├── game.go      # Core game logic
├── types.go     # Game-specific types
└── utils.go     # Helper functions
```

## Development Guidelines

1. All game logic is contained within the `superace` package
2. The game uses shared platform services for RNG and settings
3. Maintain existing functionality when making changes to ensure game integrity
4. New features should follow the established pattern for consistency

## Future Enhancements

- Additional special symbol types
- Bonus rounds or mini-games
- Enhanced visual indicators for golden symbols
- Progressive jackpots

## Deployment

The game is part of a multi-game server with a single entry point. Configuration is managed through environment variables specified in the `.env` file.