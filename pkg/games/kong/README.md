# Treasure Hunt(Kong) Game

A 5-reel, 3-row slot machine-style game featuring various treasuring hunting symbols with a 243-ways-to-win mechanism.

## Game Features

- **5x3 Grid Layout**: 5 reels with 3 symbols each
- **243 Ways to Win**: No fixed paylines, matching symbols on adjacent reels win
- **Wild Symbols**: Substitute for any symbol except Scatter
- **Scatter Bonus**: Scatter symbols trigger free spins with multipliers
- **Increasing Multipliers**: More Scatter symbols provide higher multipliers
- **Retrigger Mechanism**: Free spins can be retriggered

## Symbols

| Symbol | Description | 3 of a Kind | 4 of a Kind | 5 of a Kind |
|--------|-------------|-------------|-------------|-------------|
| Treasure Chest | Highest value symbol | 40 | 100 | 250 |
| Explorer | High value symbol | 30 | 80 | 200 |
| Compass | Medium-high value symbol | 25 | 60 | 175 |
| Binoculars | Medium value symbol | 20 | 50 | 150 |
| A | Lower value symbol | 10 | 20 | 100 |
| K | Lower value symbol | 8 | 15 | 90 |
| Q | Lower value symbol | 6 | 12 | 80 |
| J | Lowest value symbol | 5 | 10 | 70 |
| Wild | Substitutes for all symbols except Scatter | - | - | - |
| Scatter | Triggers free spins | - | - | - |

## Special Features

### Free Spins Bonus

- **Trigger**: Scatter on each of the 5 reels
- **Initial Award**: 13 free spins
- **Multipliers**: Based on total Scatter count
  - 5 Scatters: 3x multiplier
  - 6 Scatters: 6x multiplier
  - 7 Scatters: 9x multiplier
  - 8 Scatters: 18x multiplier
  - 9 Scatters: 36x multiplier
  - 10 Scatters: 72x multiplier
- **Retrigger**: Free spins can be retriggered (max 100 free spins)

## Betting Options

Fixed 243 ways with multiple bet levels:
- 0.3
- 0.6
- 0.9
- 1.5
- 3.0

## API Endpoints

### Spin (`POST /spin/kong`)

#### Request
```json
{
  "client_id": "1",
  "game_id": "44",
  "player_id": "22",
  "bet_amount": 0.9,
  "is_free_spin": false,
  "free_spin_count": 0,
  "bonus_multiplier": 0,
  "original_bet_amount": 0
}
```

#### Response
```json
{
  "status": "success",
  "message": "",
  "reels": [
    ["Explorer", "Q", "A"],
    ["Wild", "K", "Binoculars"],
    ["Explorer", "Wild", "J"],
    ["Explorer", "Compass", "K"],
    ["Q", "Explorer", "J"]
  ],
  "win_amount": 8.1,
  "is_free_spin": false,
  "free_spin_count": 0,
  "bonus_multiplier": 0,
  "free_spin_triggered": false
}
```

### Free Spin Request

```json
{
  "client_id": "1",
  "game_id": "44",
  "player_id": "22",
  "bet_amount": 0,
  "is_free_spin": true,
  "free_spin_count": 10,
  "bonus_multiplier": 6,
  "original_bet_amount": 0.9
}
```

## Game Logic Details

### Win Calculation

1. **243 Ways to Win**: The game checks all possible combinations across the reels
2. **Adjacent Reels**: Winning combinations must start from the leftmost reel
3. **Minimum 3 Symbols**: At least 3 matching symbols are needed for a win
4. **Ways Calculation**: Number of matching symbols on each reel are multiplied together
5. **Multiple Wins**: All winning symbol combinations are paid

### RNG Integration

The game integrates with an external RNG service to ensure fair outcomes based on:
- Player RTP settings
- Potential win amounts
- Bet size

### Concurrency Optimization

The game uses Goroutines to parallelize reel generation, optimizing performance by:
- Using multiple CPU cores
- Trying different combinations in parallel
- Ensuring at least one potential win in every spin

## Running as Standalone Service

The Treasure Hunt game can run as a standalone service:

```bash
go run cmd/kong/main.go
```

## Integration

This game can be:
1. Run as a standalone service
2. Integrated into the main games server
3. Imported as a package into other Go applications

## Environmental Configuration

Configuration via `.env` file:
```
RNG_API_URL=http://159.89.235.166:17003/api/proxy/rng/1
SETTINGS_API_URL=https://t2.ibibe.africa/get-game-settings
PORT=11400
LOG_FILE=app.log
```