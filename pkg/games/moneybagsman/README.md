# Money Bags Man Game - API Integration Guide for Unity Developers

## Overview

This document provides comprehensive guidelines for integrating the Money Bags Man slot game backend API with a Unity frontend. The game features a standard 5x3 reel structure with 243 ways to win, wild substitutions, scatter symbols, and a free spin bonus feature with progressive multipliers.

## API Endpoint

- Base URL: `https://api.example.com` (replace with actual production URL)
- Spin endpoint: `POST /spin/moneybagsman`
- Health check: `GET /status`

## Game Mechanics

### Core Game Rules
- 5x3 reel structure
- 243 ways to win
- Denomination: 0.01
- Bet multipliers: 1, 2, 3, 5, 10
- Bet amounts: 0.5, 1.0, 1.5, 2.5, 5.0

### Symbols
- High-value symbols: Airplane, Yacht, Car, Motorcycle
- Medium-value symbols: A, K
- Low-value symbols: Q, J
- Special symbols: Wild, Scatter

### Paytable (for Bet Multiplier x1)
| Symbol      | 3 of a Kind | 4 of a Kind | 5 of a Kind |
|-------------|-------------|-------------|-------------|
| Airplane    | 75          | 150         | 400         |
| Yacht       | 50          | 150         | 300         |
| Car         | 40          | 100         | 250         |
| Motorcycle  | 30          | 100         | 200         |
| A           | 15          | 30          | 125         |
| K           | 15          | 30          | 125         |
| Q           | 10          | 20          | 100         |
| J           | 10          | 20          | 100         |

### Wild Symbol
- Appears on reels 2, 3, 4, and 5 only
- Substitutes for all symbols except Scatter

### Free Spin Bonus
- Triggered when at least one Scatter appears on each of the 5 reels
- Initially awards 12 free spins
- Maximum of 50 free spins in a single game
- Free spins can be retriggered during the bonus round (+12 spins)
- Multiplier progression based on the number of triggering scatters:
  - 5 scatters: Starts at 1x, increases by 1 each spin, up to 50x
  - 6 scatters: Starts at 2x, increases by 2 each spin, up to 100x
  - 7+ scatters: Starts at 3x, increases by 3 each spin, up to 150x

## API Interaction Flow

### 1. Base Game Spin

#### Request Format
```json
{
  "bet_amount": 1.0,
  "is_free_spin": false,
  "current_free_spin_index": 0,
  "remaining_free_spins": 0,
  "total_free_spins_awarded": 0,
  "free_spin_multiplier": 0,
  "scatter_count": 0,
  "client_id": "client_id_here",
  "game_id": "52",
  "player_id": "player_id_here"
}
```

#### Response Format
```json
{
  "reels": [
    ["Symbol1", "Symbol2", "Symbol3"],
    ["Symbol1", "Symbol2", "Symbol3"],
    ["Symbol1", "Symbol2", "Symbol3"],
    ["Symbol1", "Symbol2", "Symbol3"],
    ["Symbol1", "Symbol2", "Symbol3"]
  ],
  "win_amount": 1.50,
  "win_details": [
    {
      "symbol": "SymbolName",
      "count": 3,
      "payout": 1.50,
      "positions": [
        {"reel": 0, "row": 0},
        {"reel": 1, "row": 0},
        {"reel": 2, "row": 0}
      ]
    }
  ],
  "scatter_count": 0,
  "free_spin_triggered": false,
  "free_spin_retriggered": false,
  "is_free_spin": false,
  "remaining_free_spins": 0,
  "current_free_spin_index": 0,
  "free_spin_multiplier": 0,
  "total_free_spins_awarded": 0,
  "max_multiplier": 0,
  "multiplier_increment": 0
}
```

### 2. Free Spin Game

#### When Free Spins Are Triggered

The response will contain these key fields:
```json
{
  "free_spin_triggered": true,
  "is_free_spin": true,
  "remaining_free_spins": 12,
  "scatter_count": 5,
  "free_spin_multiplier": 1,
  "max_multiplier": 50,
  "multiplier_increment": 1
}
```

#### Free Spin Request Format
```json
{
  "bet_amount": 1.0,
  "is_free_spin": true,
  "current_free_spin_index": 0,
  "remaining_free_spins": 12,
  "total_free_spins_awarded": 12,
  "free_spin_multiplier": 1,
  "scatter_count": 5,
  "client_id": "client_id_here",
  "game_id": "52",
  "player_id": "player_id_here"
}
```

#### Free Spin Response Format
```json
{
  "reels": [...],
  "win_amount": 3.00,
  "win_details": [...],
  "scatter_count": 0,
  "free_spin_triggered": false,
  "free_spin_retriggered": false,
  "is_free_spin": true,
  "remaining_free_spins": 11,
  "current_free_spin_index": 1,
  "free_spin_multiplier": 2,
  "total_free_spins_awarded": 12,
  "max_multiplier": 50,
  "multiplier_increment": 1
}
```

#### When Free Spins Are Retriggered
```json
{
  "free_spin_retriggered": true,
  "remaining_free_spins": 23,
  "total_free_spins_awarded": 24
}
```

#### End of Free Spins
When the last free spin is completed, the response will have:
```json
{
  "is_free_spin": false,
  "remaining_free_spins": 0,
  "current_free_spin_index": 0,
  "free_spin_multiplier": 0,
  "total_free_spins_awarded": 0
}
```

## Implementation Guidelines

### 1. Game Initialization
- Default to base game mode (`is_free_spin: false`)
- Set all free spin parameters to 0 initially
- Configure UI elements for displaying reels, wins, and free spin information

### 2. Sending a Spin Request
- For base game:
  - Set `bet_amount` to the player's selected amount
  - Set `is_free_spin: false` and all free spin fields to 0
- For free spins:
  - Maintain the same `bet_amount` from the triggering spin
  - Set `is_free_spin: true`
  - Use the state values received from the previous spin response

### 3. Processing Spin Response
- Update the reels display based on the `reels` array
- For wins:
  - Display win amount
  - Highlight winning symbol combinations using `win_details.positions`
- For free spin trigger:
  - Display appropriate animation/transition
  - Show the player the number of free spins awarded
  - Display the starting multiplier

### 4. Handling Free Spin State
- **CRITICAL**: Always use the state values from the most recent response in your next request
- Track these important fields:
  - `is_free_spin`
  - `remaining_free_spins`
  - `current_free_spin_index`
  - `free_spin_multiplier`
  - `scatter_count` (determines multiplier progression)
  - `total_free_spins_awarded`

### 5. Multiplier Progression
The `free_spin_multiplier` will automatically increase between spins based on the `scatter_count` that triggered free spins:
- 5 scatters: 1x → 2x → 3x → 4x → ... (increasing by 1)
- 6 scatters: 2x → 4x → 6x → 8x → ... (increasing by 2)
- 7+ scatters: 3x → 6x → 9x → 12x → ... (increasing by 3)

## Troubleshooting

### Common Issues and Solutions

1. **No wins are displayed despite winning combinations**
   - Check if you're correctly interpreting the `win_details` positions
   - Ensure your symbol mapping matches the server-side mapping

2. **Free spin multiplier not increasing correctly**
   - Verify you're passing the correct `scatter_count` in your requests
   - Ensure you're using the `free_spin_multiplier` value from the previous response

3. **Free spins end prematurely**
   - Check if you're correctly passing `remaining_free_spins` from the previous response
   - Verify that you're not resetting any state values between requests

4. **API errors**
   - Validate all required fields are present in your request
   - Check that bet amount is one of the allowed values (0.5, 1.0, 1.5, 2.5, 5.0)
   - Ensure player ID and client ID are valid

## API Error Responses

Error responses will have the following format:
```json
{
  "status": "error",
  "message": "Error description"
}
```

Common error messages:
- "Invalid request body" - Request JSON is malformed
- "client_id is required" - Missing client ID
- "invalid bet amount" - Bet amount is not one of the allowed values
- "Failed to retrieve game settings" - Settings service is unavailable
- "Failed to determine outcome" - RNG service is unavailable

