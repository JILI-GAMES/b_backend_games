# Open Sesame 2 Game - API Integration Guide for Unity Developers

## Overview

This document provides comprehensive guidelines for integrating the Open Sesame 2 slot game backend API with a Unity frontend. The game features a 5x3 reel structure with 243 ways to win, wild substitutions, scatter pays, and two types of free spin bonus features.

## API Endpoints

- Base URL: `https://b.api.ibibe.africa`
- Spin endpoint: `POST /spin/opensesame2`
- Free Spin Option Selection: `POST /select-free-spin-option/opensesame2`
- Health check: `GET /status`

## Game Mechanics

### Core Game Rules
- 5x3 reel structure
- 243 ways to win
- Denomination: 0.01
- Bet amounts: 0.6, 1.2, 3.0, 6.0, 15.0 (corresponding to multipliers: 1, 2, 5, 10, 25)

### Symbols
- High-value symbols: Woman, Man with turban, Horse, Pottery
- Medium-value symbols: A, K
- Low-value symbols: Q, J, 10, 9
- Special symbols: 
  - Wild (man with white turban)
  - Scatter (gold coins)
  - Free Spin (Open Sesame)
  - Mystery Box (treasure chest)

### Paytable (for Bet Multiplier x1)
| Symbol      | 3 of a Kind | 4 of a Kind | 5 of a Kind |
|-------------|-------------|-------------|-------------|
| Woman       | 60          | 200         | 360         |
| Man         | 30          | 100         | 200         |
| Horse       | 24          | 48          | 120         |
| Pottery     | 24          | 48          | 120         |
| A           | 18          | 30          | 80          |
| K           | 18          | 30          | 80          |
| Q           | 18          | 24          | 72          |
| J           | 18          | 24          | 72          |
| 10          | 18          | 24          | 60          |
| 9           | 18          | 24          | 60          |
| Scatter     | 2           | 10          | 40          |
| Free Spin   | 2           | -           | -           |

### Wild Symbol
- Appears on reels 2, 3, 4, and 5 only
- Substitutes for all symbols except Scatter, Free Spin, and Mystery Box symbols

### Scatter Symbol
- Pays anywhere on the reels
- Scatter payouts are multiplied by the total bet (bet amount)
- 3, 4, or 5 Scatter symbols pay 2x, 10x, or 40x the total bet respectively

### Free Spin Symbol
- Appears only on reels 1, 2, and 3
- 3 Free Spin symbols (1 on each of the first three reels) triggers the Regular Free Spin Bonus
- Also pays 2 times the total bet for 3 symbols

### Mystery Box Symbol
- Appears only on reel 3
- Used to trigger the Extra Free Spin Bonus when combined with Free Spin symbols

### Free Spin Bonus Features

#### Regular Free Spin Bonus
- Triggered when at least one Free Spin symbol appears on each of the first three reels
- Player selects a lamp to reveal the multiplier (2x, 3x, 4x, 5x, or 6x)
- Player selects a chest to reveal the number of free spins (5, 8, 10, or 15)
- All wins during free spins are multiplied by the revealed multiplier, except 5 of a kind of the Woman symbol
- Free spins can be retriggered (maximum 250 free spins)
- Retriggering awards the same number of free spins and uses the same multiplier as the initial trigger

#### Extra Free Spin Bonus
- Triggered when 2 Free Spin symbols appear on reels 1 and 2, and a Mystery Box appears on reel 3
- Player selects a lamp to reveal the multiplier (2x, 3x, 4x, 5x, or 6x)
- Player selects a chest to reveal the number of free spins (5, 8, 10, or 15)
- Player selects a treasure to reveal either:
  - An extra multiplier (1x to 5x) that multiplies the initial multiplier
  - Additional free spins (2 to 10)
- All wins during free spins are multiplied by the combined multipliers, except 5 of a kind of the Woman symbol
- Free spins can be retriggered (maximum 250 free spins)

## API Interaction Flow

### 1. Base Game Spin

#### Request Format
```json
{
  "bet_amount": 0.6,
  "is_free_spin": false,
  "is_extra_free_spin": false,
  "current_free_spin_index": 0,
  "remaining_free_spins": 0,
  "total_free_spins_awarded": 0,
  "free_spin_multiplier": 0,
  "extra_free_spin_multiplier": 0,
  "client_id": "client_id_here",
  "game_id": "53",
  "player_id": "player_id_here",
  "bet_id": "bet_id_here"
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
  "win_amount": 1.5,
  "win_details": [
    {
      "symbol": "SymbolName",
      "count": 3,
      "payout": 1.5,
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
  "free_spin_win_amount": 1.2,
  "free_spin_positions": [
    {"reel": 0, "row": 1},
    {"reel": 1, "row": 2},
    {"reel": 2, "row": 0}
  ],
  "mystery_box_count": 1,
  "mystery_box_positions": [
    {"reel": 2, "row": 1}
  ],
  "free_spin_triggered": true,
  "extra_free_spin_triggered": false,
  "free_spin_retriggered": false,
  "is_free_spin": false,
  "is_extra_free_spin": false,
  "remaining_free_spins": 0,
  "current_free_spin_index": 0,
  "free_spin_multiplier": 0,
  "extra_free_spin_multiplier": 0,
  "total_free_spins_awarded": 0,
  "bet_amount": 0.6,
  "bet_multiplier": 1
}
```

### 2. Free Spin Trigger & Selection

When free spins are triggered, the response will include `free_spin_triggered: true` or `extra_free_spin_triggered: true`. The Unity client should then allow the player to make selections:

#### Regular Free Spin Selection Request
```json
{
  "chest_index": 2,
  "lamp_index": 1,
  "is_extra_bonus": false,
  "client_id": "client_id_here",
  "game_id": "53",
  "player_id": "player_id_here",
  "bet_id": "bet_id_here"
}
```

#### Extra Free Spin Selection Request (for treasure selection)
```json
{
  "treasure_index": 3,
  "is_extra_bonus": true,
  "client_id": "client_id_here",
  "game_id": "53",
  "player_id": "player_id_here",
  "bet_id": "bet_id_here"
}
```

#### Selection Response
```json
{
  "free_spin_count": 10,
  "free_spin_multiplier": 3,
  "extra_multiplier": 0,
  "extra_free_spins": 0,
  "is_extra_multiplier": false
}
```

#### Extra Bonus Selection Response
```json
{
  "free_spin_count": 0,
  "free_spin_multiplier": 0,
  "extra_multiplier": 3,
  "extra_free_spins": 0,
  "is_extra_multiplier": true
}
```

### 3. Free Spin Game

After receiving the selection responses, the Unity client should send the first free spin request:

#### Regular Free Spin Request Format
```json
{
  "bet_amount": 0.60,
  "is_free_spin": true,
  "is_extra_free_spin": false,
  "current_free_spin_index": 0,
  "remaining_free_spins": 10,
  "total_free_spins_awarded": 10,
  "free_spin_multiplier": 3,
  "extra_free_spin_multiplier": 0,
  "client_id": "client_id_here",
  "game_id": "53",
  "player_id": "player_id_here",
  "bet_id": "bet_id_here"
}
```

#### Extra Free Spin Request Format
```json
{
  "bet_amount": 0.60,
  "is_free_spin": false,
  "is_extra_free_spin": true,
  "current_free_spin_index": 0,
  "remaining_free_spins": 10,
  "total_free_spins_awarded": 10,
  "free_spin_multiplier": 3,
  "extra_free_spin_multiplier": 3,
  "client_id": "client_id_here",
  "game_id": "53",
  "player_id": "player_id_here",
  "bet_id": "bet_id_here"
}
```

#### Free Spin Response Format
```json
{
  "reels": [...],
  "win_amount": 9.00,
  "win_details": [...],
  "scatter_count": 0,
  "scatter_win_amount": 0.0,
  "scatter_positions": [],
  "free_spin_count": 0,
  "free_spin_win_amount": 0.0,
  "free_spin_positions": [],
  "mystery_box_count": 0,
  "mystery_box_positions": [],
  "free_spin_triggered": false,
  "extra_free_spin_triggered": false,
  "free_spin_retriggered": false,
  "is_free_spin": true,
  "is_extra_free_spin": false,
  "remaining_free_spins": 9,
  "current_free_spin_index": 1,
  "free_spin_multiplier": 3,
  "extra_free_spin_multiplier": 0,
  "total_free_spins_awarded": 10,
  "bet_amount": 0.60,
  "bet_multiplier": 1
}
```

#### When Free Spins Are Retriggered
```json
{
  "free_spin_retriggered": true,
  "remaining_free_spins": 19,
  "total_free_spins_awarded": 20
}
```

#### End of Free Spins
When the last free spin is completed, the response will have:
```json
{
  "is_free_spin": false,
  "is_extra_free_spin": false,
  "remaining_free_spins": 0,
  "current_free_spin_index": 0,
  "free_spin_multiplier": 0,
  "extra_free_spin_multiplier": 0,
  "total_free_spins_awarded": 0
}
```

## Detailed API Reference

### Spin Request Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| bet_amount | float | The selected bet amount (0.60, 1.20, 3.00, 6.00, 15.00) |
| is_free_spin | bool | Whether this is a regular free spin |
| is_extra_free_spin | bool | Whether this is an extra free spin |
| current_free_spin_index | int | Current index of free spin (0-based) |
| remaining_free_spins | int | Number of free spins remaining |
| total_free_spins_awarded | int | Total number of free spins awarded |
| free_spin_multiplier | int | Current free spin multiplier |
| extra_free_spin_multiplier | int | Extra multiplier for extra free spins |
| client_id | string | Client identifier |
| game_id | string | Game identifier |
| player_id | string | Player identifier |
| bet_id | string | Bet identifier |

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
| mystery_box_count | int | Number of mystery box symbols on reel 3 |
| mystery_box_positions | []Position | Positions of mystery box symbols |
| free_spin_triggered | bool | Whether regular free spins were triggered |
| extra_free_spin_triggered | bool | Whether extra free spins were triggered |
| free_spin_retriggered | bool | Whether free spins were retriggered |
| is_free_spin | bool | Whether we're in regular free spin mode |
| is_extra_free_spin | bool | Whether we're in extra free spin mode |
| remaining_free_spins | int | Number of free spins remaining |
| current_free_spin_index | int | Current index of free spin (0-based) |
| free_spin_multiplier | int | Current free spin multiplier |
| extra_free_spin_multiplier | int | Extra multiplier for extra free spins |
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
| treasure_index | int | Index of selected treasure for extra bonus (0-9) |
| is_extra_bonus | bool | Whether this is for an extra bonus selection |
| client_id | string | Client identifier |
| game_id | string | Game identifier |
| player_id | string | Player identifier |
| bet_id | string | Bet identifier |

### Select Option Response Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| free_spin_count | int | Number of free spins (5, 8, 10, or 15) |
| free_spin_multiplier | int | Free spin multiplier (2, 3, 4, 5, or 6) |
| extra_multiplier | int | Extra multiplier (1, 2, 3, 4, or 5) |
| extra_free_spins | int | Extra free spins (2, 4, 6, 8, or 10) |
| is_extra_multiplier | bool | Whether extra selection gave multiplier or free spins |

## Implementation Guidelines

### 1. Game Initialization
- Default to base game mode (`is_free_spin: false, is_extra_free_spin: false`)
- Set all free spin parameters to 0 initially
- Configure UI elements for displaying reels, wins, and free spin information

### 2. Sending a Spin Request
- For base game:
  - Set `bet_amount` to the player's selected amount (0.60, 1.20, 3.00, 6.00, 15.00)
  - Set `is_free_spin: false`, `is_extra_free_spin: false` and all free spin fields to 0
- For free spins:
  - Maintain the same `bet_amount` from the triggering spin
  - Set appropriate `is_free_spin` or `is_extra_free_spin` flags
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
  - For extra free spin bonus, also present the treasure chest for selection
  - After selections, show the number of free spins and multiplier(s)
  - Begin the free spin sequence

### 4. Free Spin Selection
- When regular free spins are triggered:
  - Allow the player to select one lamp (from 5 options) to determine the multiplier
  - Allow the player to select one chest (from 4 options) to determine the number of free spins
- When extra free spins are triggered:
  - Allow the player to select lamp and chest as above
  - Also allow the player to select one treasure (from 10 options) to determine extra multiplier or extra free spins
- Send appropriate selection requests to the server
- Use the returned values to start the free spin sequence

### 5. Handling Game State
- **CRITICAL**: Always use the state values from the most recent response in your next request
- Track these important fields:
  - `is_free_spin`
  - `is_extra_free_spin`
  - `remaining_free_spins`
  - `current_free_spin_index`
  - `free_spin_multiplier`
  - `extra_free_spin_multiplier`
  - `total_free_spins_awarded`
  - `bet_amount` (maintain the same bet amount during free spins)

## Bet Amounts and Multipliers

The game uses the following bet amounts and their corresponding multipliers:

| Bet Amount | Bet Multiplier | Total Credits |
|------------|----------------|---------------|
| 0.6        | 1              | 60            |
| 1.2        | 2              | 120           |
| 3.0        | 5              | 300           |
| 6.0        | 10             | 600           |
| 15.0       | 25             | 1500          |

When sending requests to the API, always use the `bet_amount` field with one of the values from the first column. The server will automatically calculate the appropriate bet multiplier.

## Troubleshooting

### Common Issues and Solutions

1. **No wins are displayed despite winning combinations**
   - Check if you're correctly interpreting the `win_details` positions
   - Ensure your symbol mapping matches the server-side mapping

2. **Free spin selection not working**
   - Verify you're sending the correct indices and setting the `is_extra_bonus` flag appropriately
   - Check that you're correctly handling the response values

3. **Free spins end prematurely**
   - Ensure you're correctly passing `remaining_free_spins` from the previous response
   - Verify that you're not resetting any state values between requests

4. **Extra free spin features not working**
   - Make sure you're setting `is_extra_free_spin: true` for extra free spins
   - Verify that you're passing the correct `extra_free_spin_multiplier` value

5. **API errors**
   - Validate all required fields are present in your request
   - Check that bet amount is one of the allowed values (0.60, 1.20, 3.00, 6.00, 15.00)
   - Ensure player ID, client ID, and bet ID are valid

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
- "invalid treasure index" - Treasure index is out of range (must be 0-9)
- "Failed to retrieve game settings" - Settings service is unavailable
- "Failed to determine outcome" - RNG service is unavailable

