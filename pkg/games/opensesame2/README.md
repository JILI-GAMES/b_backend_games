# Open Sesame 2 Game - API Integration Guide for Unity Developers

## Overview

This document provides comprehensive guidelines for integrating the Open Sesame 2 slot game backend API with a Unity frontend. The game features a 5x3 reel structure with 243 ways to win, wild substitutions, scatter pays, and combination-based special symbol payouts with two types of free spin bonus features.

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
  - Free Spin (Open Sesame) - **Appears only on reels 1, 2, 3**
  - Mystery Box (treasure chest) - **Appears only on reel 3**

### Symbol Placement Rules
- **Wild**: Appears only on reels 2, 3, 4, and 5
- **Free Spin**: Appears only on reels 1, 2, and 3 (maximum one per reel)
- **Mystery Box**: Appears only on reel 3 (maximum one per reel)
- **Scatter**: Can appear anywhere on all reels

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

### Special Symbol Combination Payouts

#### **IMPORTANT: Combination-Based Payouts**
Free Spin and Mystery Box symbols use a **combination-based payout system**, not individual symbol counting:

| Combination | Credits | Description |
|-------------|---------|-------------|
| 3 Free Spin symbols (one each on reels 1, 2, 3) | **2** | Regular Free Spin combination |
| 2 Free Spin + 1 Mystery Box (Free Spin on reels 1&2, Mystery Box on reel 3) | **5** | Mystery Box combination |

**Calculation**: `credits × total_bet_amount`

#### **Examples**
- **3 Free Spin combo, bet 0.60**: 2 × 0.60 = **1.20 win**
- **Mystery Box combo, bet 3.00**: 5 × 3.00 = **15.00 win**

### Wild Symbol
- Appears on reels 2, 3, 4, and 5 only
- Substitutes for all symbols except Scatter, Free Spin, and Mystery Box symbols

### Scatter Symbol
- Pays anywhere on the reels
- Scatter payouts are multiplied by the total bet (bet amount)
- 3, 4, or 5 Scatter symbols pay 2x, 10x, or 40x the total bet respectively

### Free Spin Bonus Features

#### Regular Free Spin Bonus
- **Trigger**: One Free Spin symbol on each of reels 1, 2, and 3
- **Payout**: Also pays 2 credits × total bet for the combination
- Player selections:
  - **Lamp**: Reveals multiplier (2x, 3x, 4x, 5x, or 6x)
  - **Chest**: Reveals free spins (5, 8, 10, or 15)
- All wins during free spins are multiplied by the revealed multiplier, except 5 of a kind of the Woman symbol
- Free spins can be retriggered (maximum 250 free spins)

#### Extra Free Spin Bonus
- **Trigger**: Free Spin symbols on reels 1 and 2, Mystery Box on reel 3
- **Payout**: Also pays 5 credits × total bet for the combination
- Player selections:
  - **Lamp**: Reveals multiplier (2x, 3x, 4x, 5x, or 6x)
  - **Chest**: Reveals free spins (5, 8, 10, or 15)
  - **Treasure**: Reveals either extra multiplier (1x-5x) OR extra free spins (2-10)
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
  "combination_count": 3,
  "combination_win_amount": 1.2,
  "combination_type": "FreeSpinCombination",
  "combination_positions": [
    {"reel": 0, "row": 1},
    {"reel": 1, "row": 2},
    {"reel": 2, "row": 0}
  ],
  "combination_symbols": ["FreeSpins", "FreeSpins", "FreeSpins"],
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

### 2. Combination Payouts

#### Regular Free Spin Combination
- **Trigger**: Free Spin symbols on reels 1, 2, and 3
- **Response fields**:
```json
{
  "combination_count": 3,
  "combination_win_amount": 1.2,
  "combination_type": "FreeSpinCombination",
  "combination_symbols": ["FreeSpins", "FreeSpins", "FreeSpins"],
  "free_spin_triggered": true
}
```

#### Mystery Box Combination  
- **Trigger**: Free Spin symbols on reels 1&2, Mystery Box on reel 3
- **Response fields**:
```json
{
  "combination_count": 3,
  "combination_win_amount": 3.0,
  "combination_type": "MysteryBoxCombination", 
  "combination_symbols": ["FreeSpins", "FreeSpins", "MysteryBox"],
  "extra_free_spin_triggered": true
}
```

### 3. Free Spin Selection

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

### 4. Free Spin Game

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

#### Regular Free Spin Response Format
```json
{
  "reels": [
    ["Woman", "Horse", "A"],
    ["Wild", "K", "Q"],
    ["J", "Scatter", "10"],
    ["9", "A", "Woman"],
    ["Horse", "Pottery", "K"]
  ],
  "win_amount": 9.00,
  "win_details": [
    {
      "symbol": "Woman",
      "count": 3,
      "payout": 9.00,
      "positions": [
        {"reel": 0, "row": 0},
        {"reel": 1, "row": 0},
        {"reel": 4, "row": 2}
      ]
    }
  ],
  "scatter_count": 1,
  "scatter_win_amount": 0.0,
  "scatter_positions": [
    {"reel": 2, "row": 1}
  ],
  "combination_count": 0,
  "combination_win_amount": 0.0,
  "combination_type": "",
  "combination_positions": [],
  "combination_symbols": [],
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
  "extra_free_spin_multiplier": 2,
  "client_id": "client_id_here",
  "game_id": "53",
  "player_id": "player_id_here",
  "bet_id": "bet_id_here"
}
```

#### Extra Free Spin Response Format
```json
{
  "reels": [
    ["Man", "Q", "Horse"],
    ["A", "Wild", "K"],
    ["Pottery", "J", "10"],
    ["9", "Woman", "A"],
    ["K", "Horse", "Q"]
  ],
  "win_amount": 18.00,
  "win_details": [
    {
      "symbol": "Man",
      "count": 4,
      "payout": 18.00,
      "positions": [
        {"reel": 0, "row": 0},
        {"reel": 1, "row": 1},
        {"reel": 3, "row": 1},
        {"reel": 4, "row": 1}
      ]
    }
  ],
  "scatter_count": 0,
  "scatter_win_amount": 0.0,
  "scatter_positions": [],
  "combination_count": 0,
  "combination_win_amount": 0.0,
  "combination_type": "",
  "combination_positions": [],
  "combination_symbols": [],
  "free_spin_triggered": false,
  "extra_free_spin_triggered": false,
  "free_spin_retriggered": false,
  "is_free_spin": false,
  "is_extra_free_spin": true,
  "remaining_free_spins": 9,
  "current_free_spin_index": 1,
  "free_spin_multiplier": 3,
  "extra_free_spin_multiplier": 2,
  "total_free_spins_awarded": 10,
  "bet_amount": 0.60,
  "bet_multiplier": 1
}
```

#### When Free Spins Are Retriggered
```json
{
  "combination_type": "FreeSpinCombination",
  "combination_win_amount": 1.2,
  "free_spin_retriggered": true,
  "remaining_free_spins": 19,
  "total_free_spins_awarded": 20,
  "current_free_spin_index": 5
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

### Spin Response Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| reels | [][]string | 5x3 grid of symbols |
| win_amount | float | Total win amount for this spin |
| win_details | []WinDetail | Details of each winning combination |
| scatter_count | int | Number of scatter symbols on the reels |
| scatter_win_amount | float | Win amount from scatter symbols |
| scatter_positions | []Position | Positions of scatter symbols |
| combination_count | int | Number of symbols in special combination (3) |
| combination_win_amount | float | Win amount from special combination |
| combination_type | string | "FreeSpinCombination" or "MysteryBoxCombination" |
| combination_positions | []Position | Positions of symbols in combination |
| combination_symbols | []string | Symbol names in combination |
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

## Implementation Guidelines

### 1. Handling Combination Payouts

#### Visual Display
- Highlight all three positions when a combination win occurs
- Show different animations for Free Spin vs Mystery Box combinations
- Display combination payout separately from regular line wins

#### Win Calculation
```javascript
total_win = regular_wins + scatter_wins + combination_wins
```

#### Animation Sequence
1. Show regular line wins
2. Show scatter wins (if any)
3. Show combination wins with special animation
4. If bonus triggered, transition to selection screens

### 2. Bonus Trigger Detection

#### Regular Free Spin Bonus
```javascript
if (response.combination_type === "FreeSpinCombination" && response.free_spin_triggered) {
    // Show lamp and chest selection
}
```

#### Extra Free Spin Bonus  
```javascript
if (response.combination_type === "MysteryBoxCombination" && response.extra_free_spin_triggered) {
    // Show lamp, chest, and treasure selection
}
```

### 3. State Management
- Always use the state values from the most recent response
- Track combination payouts separately from regular wins
- Maintain bet amount consistency during free spins

## Troubleshooting

### Common Issues

1. **Combination payouts not displaying**
   - Check `combination_type` and `combination_win_amount` fields
   - Verify you're highlighting the correct `combination_positions`

2. **Wrong payout calculations**
   - Remember: combination payouts use `credits × total_bet`, not `credits × bet_multiplier × denomination`
   - Free Spin combo = 2 credits, Mystery Box combo = 5 credits

3. **Bonus not triggering**
   - Verify `free_spin_triggered` or `extra_free_spin_triggered` flags
   - Check that you're using correct selection indices

4. **Symbol positioning errors**
   - Free Spin symbols only appear on reels 1, 2, 3
   - Mystery Box symbols only appear on reel 3
   - Maximum one special symbol per reel

## Example Scenarios

### Scenario 1: Regular Free Spin Combination
- **Reels**: Free Spin on reel 1 (row 0), Free Spin on reel 2 (row 1), Free Spin on reel 3 (row 2)
- **Bet**: 0.60
- **Combination Payout**: 2 × 0.60 = 1.20
- **Bonus**: Regular Free Spin Bonus triggered
- **API Response**:
```json
{
  "combination_win_amount": 1.2,
  "combination_type": "FreeSpinCombination",
  "free_spin_triggered": true
}
```

### Scenario 2: Mystery Box Combination
- **Reels**: Free Spin on reel 1, Free Spin on reel 2, Mystery Box on reel 3
- **Bet**: 3.00
- **Combination Payout**: 5 × 3.00 = 15.00
- **Bonus**: Extra Free Spin Bonus triggered
- **API Response**:
```json
{
  "combination_win_amount": 15.0,
  "combination_type": "MysteryBoxCombination", 
  "extra_free_spin_triggered": true
}
```

### Scenario 3: Multiple Wins
- **Regular Wins**: 3 Horse symbols = 0.72
- **Scatter Wins**: 3 Scatter symbols = 1.20
- **Combination**: Free Spin combination = 1.20
- **Total Win**: 0.72 + 1.20 + 1.20 = **3.12**