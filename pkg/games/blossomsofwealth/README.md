# Blossoms of Wealth - API Documentation

## Overview

This document provides technical documentation for integrating the Blossoms of Wealth slot game backend API with your Unity frontend. The API provides endpoints for spinning the reels and managing the game state, including free spins and bonus features.

## API Endpoints

### Base URL

```
https://b.api.ibibe.africa
```

### Spin Endpoint

**Endpoint:** `/spin/blossomsofwealth`

**Method:** POST

**Description:** Spins the reels and returns the result, including any wins, free spin triggers, or bonus multipliers.

**Request Body:**

```json
{
  "bet_amount": 1.0,
  "is_free_spin": false,
  "current_free_spin_index": 0,
  "remaining_free_spins": 0,
  "total_free_spins_awarded": 0,
  "free_spin_multiplier": 0,
  "client_id": "client123",
  "game_id": "blossomsofwealth",
  "player_id": "player456",
  "bet_id": "bet789"
}
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| bet_amount | float | The amount bet by the player. Valid values: 0.5, 1.0, 1.5, 2.5, 5.0 |
| is_free_spin | boolean | Whether this spin is a free spin |
| current_free_spin_index | integer | The index of the current free spin (0 for the first spin) |
| remaining_free_spins | integer | The number of free spins remaining |
| total_free_spins_awarded | integer | The total number of free spins awarded in the current session |
| free_spin_multiplier | integer | The current multiplier for free spins |
| client_id | string | Unique identifier for the client/operator |
| game_id | string | Identifier for the game ("blossomsofwealth") |
| player_id | string | Unique identifier for the player |
| bet_id | string | Unique identifier for this specific bet/spin |

**Response:**

```json
{
  "reels": [
    ["A", "K", "SilverFlower"],
    ["Wild", "Q", "J"],
    ["GoldTreasure", "PurpleTreasure", "BlueTreasure"],
    ["Compass", "A", "GoldFlower"],
    ["K", "Q", "J"]
  ],
  "win_amount": 2.5,
  "win_details": [
    {
      "symbol": "A",
      "count": 3,
      "payout": 1.5,
      "positions": [
        {"reel": 0, "row": 0},
        {"reel": 1, "row": 2},
        {"reel": 3, "row": 1}
      ]
    }
  ],
  "silver_flower_count": 3,
  "silver_flower_positions": [
    {"reel": 0, "row": 2},
    {"reel": 1, "row": 1},
    {"reel": 2, "row": 0}
  ],
  "gold_flower_count": 1,
  "gold_flower_positions": [
    {"reel": 3, "row": 2}
  ],
  "bonus_multiplier": 15,
  "free_spin_triggered": true,
  "is_free_spin": false,
  "remaining_free_spins": 25,
  "current_free_spin_index": 0,
  "free_spin_multiplier": 15,
  "total_free_spins_awarded": 25
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| reels | array | 5x3 grid of symbols representing the final state of the reels |
| win_amount | float | Total payout amount for this spin |
| win_details | array | Details of individual winning combinations |
| silver_flower_count | integer | Number of Silver Flower symbols on the reels |
| silver_flower_positions | array | Positions of all Silver Flower symbols |
| gold_flower_count | integer | Number of Gold Flower symbols on the reels |
| gold_flower_positions | array | Positions of all Gold Flower symbols |
| bonus_multiplier | integer | Bonus multiplier (10-25x) if gold flowers trigger a multiplier, otherwise 0 |
| free_spin_triggered | boolean | Whether free spins were triggered on this spin |
| is_free_spin | boolean | Whether this spin was a free spin |
| remaining_free_spins | integer | Number of free spins remaining |
| current_free_spin_index | integer | Index of the current free spin |
| free_spin_multiplier | integer | Current multiplier for free spins |
| total_free_spins_awarded | integer | Total number of free spins awarded in the current session |

### Status Endpoint

**Endpoint:** `/status`

**Method:** GET

**Description:** Checks if the API is up and running.

**Response:**

```json
{
  "status": "ok",
  "game": "blossomsofwealth"
}
```

## Game Mechanics

### Symbols

The game features the following symbols:

| Symbol | Description |
|--------|-------------|
| GoldTreasure | High value symbol (3: 75, 4: 150, 5: 400) |
| PurpleTreasure | High value symbol (3: 50, 4: 150, 5: 300) |
| BlueTreasure | High value symbol (3: 40, 4: 100, 5: 250) |
| Compass | Medium value symbol (3: 30, 4: 100, 5: 200) |
| A | Low value symbol (3: 15, 4: 25, 5: 80) |
| K | Low value symbol (3: 15, 4: 25, 5: 80) |
| Q | Low value symbol (3: 10, 4: 15, 5: 60) |
| J | Low value symbol (3: 10, 4: 15, 5: 60) |
| Wild | Substitutes for all symbols except Silver Flower (appears on reels 2-5 only) |
| SilverFlower | Scatter symbol that triggers free spins (appears on reels 1-3 only) |
| GoldFlower | Bonus symbol that awards multipliers (appears on reels 4-5 only in main game) |

### Winning Combinations

- The game has 243 ways to win (all possible combinations across adjacent reels)
- All wins pay from leftmost to rightmost on adjacent reels
- Only the highest winning combination is paid on each pay way
- Minimum 3 matching symbols are required for a win

### Free Spin Bonus Feature

- Triggered when 3 or more Silver Flower symbols appear on reels 1-3
- Each Silver Flower awards 3-10 free spins at random
- The multiplier of the Free Spin Bonus feature starts with:
  - The bonus multiplier awarded from Gold Flowers in the main game, OR
  - 1x if no Gold Flowers appeared in the main game
- The multiplier increases by 1 at the start of each round of the free spins
- All winning combinations during free spins are multiplied by the current multiplier
- The multiplier accumulates until the end of the free spin bonus feature
- Free spins can be retriggered during the feature
- A maximum of 100 free spins can be awarded in a single bought game

### Bonus Multiplier Feature

- Gold Flowers award a bonus multiplier of 10x to 25x at random
- The multiplier is only awarded when Gold Flowers and Silver Flowers appear on consecutive reels from leftmost to rightmost
- This feature is only available in the main game (not during free spins)
- When free spins are triggered, this bonus multiplier becomes the initial free spin multiplier

## Integration Guidelines

### Game Flow

1. **Initial Load**:
   - Set up the game with initial bet amount (default 0.5)
   - Initialize all game state variables to zero/false

2. **Spin**:
   - Send a spin request with the current game state
   - Process the response and update the UI accordingly
   - If `free_spin_triggered` is true, prepare to transition to free spin mode:
     - Note that `free_spin_multiplier` will be set to the bonus multiplier (if Gold Flowers appeared) or 1
     - Use `remaining_free_spins` to display the number of free spins awarded
   - If `bonus_multiplier` is greater than 0, display the multiplier effect on wins

3. **Free Spins**:
   - When free spins are triggered, set `is_free_spin` to true for subsequent spins
   - Display the current `free_spin_multiplier` which increases with each spin
   - After each free spin, the server will automatically:
     - Decrement `remaining_free_spins`
     - Increment `free_spin_multiplier` by 1 for the next spin
   - If more Silver Flowers appear during free spins, additional free spins may be awarded
   - When `remaining_free_spins` reaches 0, return to the main game

### Error Handling

The API will return appropriate HTTP status codes and error messages:

- **400 Bad Request**: Invalid parameters or request body
- **500 Internal Server Error**: Server-side error

Always check the HTTP status code before processing the response. If an error occurs, display an appropriate message to the user and allow them to retry.

### Connection Management

- Implement proper timeout handling (recommended: 10-second timeout)
- Add retry logic for temporary connection issues
- Handle cases where the connection is lost during a spin

## Testing

### Test Environment

A test environment is available at:
```
https://b.api.ibibe.africa
```

### Test Credentials

For testing purposes, use the following credentials:
- client_id: "test_client"
- player_id: "test_player"
- game_id: "blossomsofwealth"
- bet_id: "test_bet_id"

### Test Cases

1. **Basic Spin**: Test a basic spin with no special features
2. **Free Spin Trigger**: Test a spin that triggers free spins
3. **Bonus Multiplier**: Test a spin that activates the bonus multiplier
4. **Free Spin with Initial Multiplier**: Test that the free spin multiplier is correctly set to the bonus multiplier or 1
5. **Free Spin Multiplier Increment**: Test that the free spin multiplier increases by 1 with each spin
6. **Free Spin Retrigger**: Test retriggering free spins during free spins
7. **Maximum Free Spins**: Test the cap of 100 free spins
8. **Maximum Bet**: Test a spin with the maximum bet amount
9. **Minimum Bet**: Test a spin with the minimum bet amount


