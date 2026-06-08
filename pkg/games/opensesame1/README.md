# Open Sesame Game - API Integration Guide for Unity Developers

## Overview

This document provides comprehensive guidelines for integrating the Open Sesame slot game backend API with a Unity frontend. The game features a 5x3 reel structure with 243 ways to win, wild substitutions, scatter pays, and a free spin bonus feature with player-selected multipliers and free spin counts.

## API Endpoints

- Base URL: `https://b.api.ibibe.africa`
- Spin endpoint: `POST /spin/opensesame`
- Free Spin Option Selection: `POST /select-free-spin-option/opensesame`
- Health check: `GET /status`

## Game Mechanics

### Core Game Rules
- 5x3 reel structure
- 243 ways to win
- Denomination: 0.01
- Bet amounts: 0.25, 0.5, 1.25, 2.5, 6.25 (corresponding to multipliers: 1, 2, 5, 10, 25)

### Symbols
- High-value symbols: Woman, Man with blue turban, Horse, Pottery
- Medium-value symbols: A, K
- Low-value symbols: Q, J, 10, 9
- Special symbols: Wild (man with white turban), Scatter (gold palace), Free Spin (Open Sesame)

### Paytable (for Bet Multiplier x1)
| Symbol      | 3 of a Kind | 4 of a Kind | 5 of a Kind |
|-------------|-------------|-------------|-------------|
| Woman       | 75          | 500         | 5000        |
| Man         | 50          | 200         | 500         |
| Horse       | 25          | 75          | 250         |
| Pottery     | 25          | 75          | 250         |
| A           | 15          | 50          | 150         |
| K           | 15          | 50          | 150         |
| Q           | 10          | 25          | 125         |
| J           | 10          | 25          | 125         |
| 10          | 5           | 25          | 100         |
| 9           | 5           | 25          | 100         |
| Scatter     | 2           | 10          | 50          |
| Free Spin   | 2           | -           | -           |

### Wild Symbol
- Appears on reels 2, 3, 4, and 5 only
- Substitutes for all symbols except Scatter and Free Spin symbols

### Scatter Symbol
- Pays anywhere on the reels
- Scatter payouts are multiplied by the total bet (bet amount)

### Free Spin Symbol
- Appears only on reels 1, 2, and 3
- 3 Free Spin symbols (1 on each of the first three reels) triggers the Free Spin Bonus
- Also pays 2 times the total bet for 3 symbols

### Free Spin Bonus
- Triggered when at least one Free Spin symbol appears on each of the first three reels
- Player selects a chest to reveal the number of free spins (8, 12, 16, or 20)
- Player selects a lamp to reveal the multiplier (2x, 3x, 4x, 5x, or 6x)
- All wins during free spins are multiplied by the revealed multiplier, except 5 of a kind of the Woman symbol
- Free spins can be retriggered (maximum 250 free spins)
- Retriggering awards the same number of free spins and uses the same multiplier as the initial trigger

## API Interaction Flow

### 1. Base Game Spin

#### Request Format
```json
{
  "bet_amount": 0.5,
  "is_free_spin": false,
  "current_free_spin_index": 0,
  "remaining_free_spins": 0,
  "total_free_spins_awarded": 0,
  "free_spin_multiplier": 0,
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
  "scatter_count": 2,
  "scatter_win_amount": 0.0,
  "scatter_positions": [
    {"reel": 0, "row": 2},
    {"reel": 3, "row": 1}
  ],
  "free_spin_count": 3,
  "free_spin_win_amount": 0.5,
  "free_spin_positions": [
    {"reel": 0, "row": 1},
    {"reel": 1, "row": 2},
    {"reel": 2, "row": 0}
  ],
  "free_spin_triggered": true,
  "free_spin_retriggered": false,
  "is_free_spin": false,
  "remaining_free_spins": 0,
  "current_free_spin_index": 0,
  "free_spin_multiplier": 0,
  "total_free_spins_awarded": 0,
  "bet_amount": 0.5,
  "bet_multiplier": 2
}
```

### 2. Free Spin Trigger & Selection

When free spins are triggered, the response will include `free_spin_triggered: true`. The Unity client should then allow the player to make selections:

#### Selection Request
```json
{
  "chest_index": 2,
  "lamp_index": 1,
  "client_id": "client_id_here",
  "game_id": "52",
  "player_id": "player_id_here"
}
```

#### Selection Response
```json
{
  "free_spin_count": 16,
  "free_spin_multiplier": 3
}
```

### 3. Free Spin Game

After receiving the selection response, the Unity client should send the first free spin request:

#### Free Spin Request Format
```json
{
  "bet_amount": 0.5,
  "is_free_spin": true,
  "current_free_spin_index": 0,
  "remaining_free_spins": 16,
  "total_free_spins_awarded": 16,
  "free_spin_multiplier": 3,
  "client_id": "client_id_here",
  "game_id": "52",
  "player_id": "player_id_here"
}
```

#### Free Spin Response Format
```json
{
  "reels": [...],
  "win_amount": 4.50,
  "win_details": [...],
  "scatter_count": 0,
  "scatter_win_amount": 0.0,
  "scatter_positions": [],
  "free_spin_count": 0,
  "free_spin_win_amount": 0.0,
  "free_spin_positions": [],
  "free_spin_triggered": false,
  "free_spin_retriggered": false,
  "is_free_spin": true,
  "remaining_free_spins": 15,
  "current_free_spin_index": 1,
  "free_spin_multiplier": 3,
  "total_free_spins_awarded": 16,
  "bet_amount": 0.5,
  "bet_multiplier": 2
}
```

#### When Free Spins Are Retriggered
```json
{
  "free_spin_retriggered": true,
  "remaining_free_spins": 28,
  "total_free_spins_awarded": 32
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

## Detailed API Reference

### Spin Request Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| bet_amount | float | The selected bet amount (0.25, 0.5, 1.25, 2.5, 6.25) |
| is_free_spin | bool | Whether this is a free spin |
| current_free_spin_index | int | Current index of free spin (0-based) |
| remaining_free_spins | int | Number of free spins remaining |
| total_free_spins_awarded | int | Total number of free spins awarded |
| free_spin_multiplier | int | Current free spin multiplier |
| client_id | string | Client identifier |
| game_id | string | Game identifier |
| player_id | string | Player identifier |

### Spin Response Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| reels | [][]string | 5x3 grid of symbols |
| win_amount | float | Total win amount for this spin |
| win_details | []WinDetail | Details of each winning combination |
| scatter_count | int | Number of scatter symbols on the reels |
| scatter_win_amount | float | Win amount from scatter symbols |
| scatter_positions | []Position | Positions of scatter symbols |
| free_spin_count | int | Number of free spin symbols on reels 1-3 |
| free_spin_win_amount | float | Win amount from free spin symbols |
| free_spin_positions | []Position | Positions of free spin symbols |
| free_spin_triggered | bool | Whether free spins were triggered |
| free_spin_retriggered | bool | Whether free spins were retriggered |
| is_free_spin | bool | Whether we're in free spin mode |
| remaining_free_spins | int | Number of free spins remaining |
| current_free_spin_index | int | Current index of free spin (0-based) |
| free_spin_multiplier | int | Current free spin multiplier |
| total_free_spins_awarded | int | Total number of free spins awarded |
| bet_amount | float | The bet amount used for this spin |
| bet_multiplier | int | The bet multiplier derived from bet amount |

### Win Detail Structure

| Field | Type | Description |
|-------|------|-------------|
| symbol | string | Symbol that formed the win |
| count | int | Number of matching symbols |
| payout | float | Win amount for this combination |
| positions | []Position | Positions of symbols in this win |

### Position Structure

| Field | Type | Description |
|-------|------|-------------|
| reel | int | Reel index (0-4) |
| row | int | Row index (0-2) |

### Select Option Request Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| chest_index | int | Index of selected chest (0-3) |
| lamp_index | int | Index of selected lamp (0-4) |
| client_id | string | Client identifier |
| game_id | string | Game identifier |
| player_id | string | Player identifier |

### Select Option Response Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| free_spin_count | int | Number of free spins (8, 12, 16, or 20) |
| free_spin_multiplier | int | Free spin multiplier (2, 3, 4, 5, or 6) |

## Implementation Guidelines

### 1. Game Initialization
- Default to base game mode (`is_free_spin: false`)
- Set all free spin parameters to 0 initially
- Configure UI elements for displaying reels, wins, and free spin information

### 2. Sending a Spin Request
- For base game:
  - Set `bet_amount` to the player's selected amount (0.25, 0.5, 1.25, 2.5, 6.25)
  - Set `is_free_spin: false` and all free spin fields to 0
- For free spins:
  - Maintain the same `bet_amount` from the triggering spin
  - Set `is_free_spin: true`
  - Use the state values received from the previous spin response

### 3. Processing Spin Response
- Update the reels display based on the `reels` array
- For wins:
  - Display win amount
  - Highlight winning symbol combinations using:
    - Regular wins: `win_details[].positions`
    - Scatter wins: `scatter_positions`
    - Free Spin wins: `free_spin_positions`
  - Show different animations for different win types
- For free spin trigger:
  - Display appropriate animation/transition
  - Present the player with chests and lamps to select
  - After selections, show the number of free spins and multiplier
  - Begin the free spin sequence

### 4. Free Spin Selection
- When free spins are triggered, allow the player to select:
  - One chest (from 4 options) to determine the number of free spins
  - One lamp (from 5 options) to determine the multiplier
- After selections, send a request to `/select-free-spin-option/opensesame`
- Use the returned values to start the free spin sequence

### 5. Handling Game State
- **CRITICAL**: Always use the state values from the most recent response in your next request
- Track these important fields:
  - `is_free_spin`
  - `remaining_free_spins`
  - `current_free_spin_index`
  - `free_spin_multiplier`
  - `total_free_spins_awarded`
  - `bet_amount` (maintain the same bet amount during free spins)

## Bet Amounts and Multipliers

The game uses the following bet amounts and their corresponding multipliers:

| Bet Amount | Bet Multiplier | Total Credits |
|------------|----------------|---------------|
| 0.25       | 1              | 25            |
| 0.50       | 2              | 50            |
| 1.25       | 5              | 125           |
| 2.50       | 10             | 250           |
| 6.25       | 25             | 625           |

When sending requests to the API, always use the `bet_amount` field with one of the values from the first column. The server will automatically calculate the appropriate bet multiplier.

## Troubleshooting

### Common Issues and Solutions

1. **No wins are displayed despite winning combinations**
   - Check if you're correctly interpreting the `win_details` positions
   - Ensure your symbol mapping matches the server-side mapping

2. **Free spin selection not working**
   - Verify you're sending the correct chest_index (0-3) and lamp_index (0-4)
   - Check that you're correctly handling the response values

3. **Free spins end prematurely**
   - Ensure you're correctly passing `remaining_free_spins` from the previous response
   - Verify that you're not resetting any state values between requests

4. **API errors**
   - Validate all required fields are present in your request
   - Check that bet amount is one of the allowed values (0.25, 0.5, 1.25, 2.5, 6.25)
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
- "invalid chest index" - Chest index is out of range (must be 0-3)
- "invalid lamp index" - Lamp index is out of range (must be 0-4)
- "Failed to retrieve game settings" - Settings service is unavailable
- "Failed to determine outcome" - RNG service is unavailable
