# Winning Mask Game - API Integration Guide for Unity Developers

## Overview

This document provides comprehensive guidelines for integrating the Winning Mask slot game backend API with a Unity frontend. The game features a 5x4 reel structure with 1024 ways to win, wild substitutions, and two bonus features: Free Spin Bonus and Mask Reel Bonus.

## API Endpoints

- Base URL: `https://b.api.ibibe.africa`
- Spin endpoint: `POST /spin/winningmask`
- Mask Reel Bonus: `POST /mask-reel-bonus/winningmask`
- Health check: `GET /status`

## Game Mechanics

### Core Game Rules
- 5x4 reel structure
- 1024 ways to win
- Denomination: 0.01
- Bet amounts: 0.5, 1.0, 2.5, 5.0, 12.5 (corresponding to multipliers: 1, 2, 5, 10, 25)
- Minimum bet: 50 credits per bet multiplier

### Symbols
- High-value symbols: Purple Mask, Orange Mask, Green Mask, Yellow Mask, Blue Mask
- Low-value symbols: A, K, Q, J, 10
- Special symbols: Wild, Bonus (Free Spin), Mask Reel

### Paytable (for Bet Multiplier x1)
| Symbol       | 3 of a Kind | 4 of a Kind | 5 of a Kind |
|--------------|-------------|-------------|-------------|
| Purple Mask  | 50          | 200         | 1000        |
| Orange Mask  | 25          | 150         | 400         |
| Green Mask   | 25          | 150         | 400         |
| Yellow Mask  | 20          | 75          | 200         |
| Blue Mask    | 20          | 75          | 200         |
| A            | 10          | 50          | 150         |
| K            | 10          | 50          | 150         |
| Q            | 5           | 20          | 100         |
| J            | 5           | 20          | 100         |
| 10           | 5           | 20          | 100         |

### Wild Symbol
- Appears on reels 2, 3, 4, and 5 only
- Substitutes for all symbols except Bonus and Mask Reel symbols

### Free Spin Bonus
- Triggered when 3 or more Bonus symbols appear on reels 1, 2, and 3 (one per reel maximum)
- Awards 10 free spins
- **Bonus Symbol Payout**: 3 Bonus symbols pay 2 credits times the total bet amount
- Free spins can be retriggered (maximum 150 free spins)
- During free spins, bet amount remains the same as the triggering spin

### Mask Reel Bonus
- Triggered when 3 or more Mask Reel symbols appear on reels 3, 4, and 5 (one per reel maximum)
- Does NOT appear during Free Spin Bonus
- **RNG-Controlled**: A potential multiplier (2x-200x) is generated and submitted to RNG
- **If RNG approves**: Full multiplier is awarded
- **If RNG declines**: Minimum multiplier (2x-3x) is awarded instead
- **Always awards some bonus** when triggered for good user experience

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
    ["Symbol1", "Symbol2", "Symbol3", "Symbol4"],
    ["Symbol1", "Symbol2", "Symbol3", "Symbol4"],
    ["Symbol1", "Symbol2", "Symbol3", "Symbol4"],
    ["Symbol1", "Symbol2", "Symbol3", "Symbol4"],
    ["Symbol1", "Symbol2", "Symbol3", "Symbol4"]
  ],
  "win_amount": 3.50,
  "win_details": [
    {
      "symbol": "SymbolName",
      "count": 3,
      "payout": 2.50,
      "positions": [
        {"reel": 0, "row": 0},
        {"reel": 1, "row": 0},
        {"reel": 2, "row": 0}
      ]
    }
  ],
  "bonus_count": 3,
  "bonus_win_amount": 1.00,
  "bonus_positions": [
    {"reel": 0, "row": 1},
    {"reel": 1, "row": 2},
    {"reel": 2, "row": 0}
  ],
  "mask_reel_count": 0,
  "mask_reel_positions": [],
  "free_spin_triggered": true,
  "free_spin_retriggered": false,
  "mask_reel_triggered": false,
  "is_free_spin": false,
  "remaining_free_spins": 10,
  "current_free_spin_index": 0,
  "total_free_spins_awarded": 10,
  "bet_amount": 1.0,
  "bet_multiplier": 2
}
```

### 2. Free Spin Game

When free spins are triggered, use the same spin endpoint with updated parameters:

#### Free Spin Request Format
```json
{
  "bet_amount": 1.0,
  "is_free_spin": true,
  "current_free_spin_index": 0,
  "remaining_free_spins": 10,
  "total_free_spins_awarded": 10,
  "client_id": "client_id_here",
  "game_id": "53",
  "player_id": "player_id_here",
  "bet_id": "bet_id_here"
}
```

### 3. Mask Reel Bonus

When the Mask Reel Bonus is triggered, call the separate bonus endpoint which will:
1. Generate a potential multiplier (2x-200x)
2. Submit to RNG for approval
3. Award full multiplier if approved, or minimum (2x-3x) if declined

#### Mask Reel Bonus Request
```json
{
  "client_id": "client_id_here",
  "game_id": "53",
  "player_id": "player_id_here",
  "bet_id": "bet_id_here",
  "bet_amount": 1.0
}
```

#### Mask Reel Bonus Response
```json
{
  "multiplier": 25,
  "win_amount": 25.0
}
```

**Note**: The multiplier returned is the RNG-approved amount, which could be the full potential multiplier or a reduced minimum multiplier.

## Detailed API Reference

### Spin Request Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| bet_amount | float | The selected bet amount (0.5, 1.0, 2.5, 5.0, 12.5) |
| is_free_spin | bool | Whether this is a free spin |
| current_free_spin_index | int | Current index of free spin (0-based) |
| remaining_free_spins | int | Number of free spins remaining |
| total_free_spins_awarded | int | Total number of free spins awarded |
| client_id | string | Client identifier |
| game_id | string | Game identifier |
| player_id | string | Player identifier |
| bet_id | string | Bet identifier |

### Spin Response Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| reels | [][]string | 5x4 grid of symbols |
| win_amount | float | Total win amount for this spin (includes bonus payout) |
| win_details | []WinDetail | Details of each winning combination |
| bonus_count | int | Number of bonus symbols on reels 1-3 |
| bonus_win_amount | float | Win amount from bonus symbols (2x bet for 3 symbols) |
| bonus_positions | []Position | Positions of bonus symbols |
| mask_reel_count | int | Number of mask reel symbols on reels 3-5 |
| mask_reel_positions | []Position | Positions of mask reel symbols |
| free_spin_triggered | bool | Whether free spins were triggered |
| free_spin_retriggered | bool | Whether free spins were retriggered |
| mask_reel_triggered | bool | Whether mask reel bonus was triggered |
| is_free_spin | bool | Whether we're in free spin mode |
| remaining_free_spins | int | Number of free spins remaining |
| current_free_spin_index | int | Current index of free spin (0-based) |
| total_free_spins_awarded | int | Total number of free spins awarded |
| bet_amount | float | The bet amount used for this spin |
| bet_multiplier | int | The bet multiplier derived from bet amount |

## Bet Amounts and Multipliers

The game uses the following bet amounts and their corresponding multipliers:

| Bet Amount | Bet Multiplier | Total Credits |
|------------|----------------|---------------|
| 0.50       | 1              | 50            |
| 1.00       | 2              | 100           |
| 2.50       | 5              | 250           |
| 5.00       | 10             | 500           |
| 12.50      | 25             | 1250          |

## Implementation Guidelines

### 1. Game Flow
1. **Base Game**: Player spins and can trigger either Free Spin Bonus or Mask Reel Bonus
2. **Free Spin Bonus**: 10 free spins with same bet amount, can retrigger
3. **Mask Reel Bonus**: Immediate multiplier win, only in base game

### 2. Handling Bonuses
- **Free Spin Trigger**: When `free_spin_triggered: true`, start free spin sequence
- **Mask Reel Trigger**: When `mask_reel_triggered: true`, call mask reel bonus endpoint
- **Retriggering**: During free spins, check for `free_spin_retriggered: true`

### 3. State Management
- Always use response values for next request
- Free spins maintain the same bet amount throughout
- Mask Reel Bonus does not appear during free spins

### 4. Visual Presentation
- Highlight winning combinations using `win_details[].positions`
- Show bonus symbols using `bonus_positions` 
- **Display bonus payout separately** using `bonus_win_amount`
- Show mask reel symbols using `mask_reel_positions`
- Display appropriate animations for each bonus type
- **Total win amount includes both regular wins and bonus payouts**

## Error Handling

Common error responses:
- "Invalid bet amount" - Bet amount not in allowed values
- "client_id is required" - Missing required field
- "Failed to retrieve game settings" - Settings service issue
- "Failed to determine outcome" - RNG service issue

All errors return:
```json
{
  "status": "error",
  "message": "Error description"
}
```