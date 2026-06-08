# Winning Mask Game - API Integration Guide for Unity Developers

## Overview

This document provides comprehensive guidelines for integrating the Winning Mask slot game backend API with a Unity frontend. The game features a 5x4 reel structure with 1024 ways to win, wild substitutions, and two bonus features: Free Spin Bonus with Mask Transformation and Mask Reel Bonus.

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

### Free Spin Bonus with Mask Transformation
- **Trigger**: 3 or more Bonus symbols appear on reels 1, 2, and 3 (one per reel maximum)
- **Awards**: 10 free spins
- **Bonus Symbol Payout**: 3 Bonus symbols pay 2 credits times the total bet amount
- **Retriggerable**: Up to maximum 150 free spins
- **Bet Consistency**: Bet amount remains the same throughout free spins

#### **🎭 Mask Transformation Feature (Free Spins Only)**
During free spins, each spin operates in **two stages**:

**Stage 1**: Normal spin calculation and wins are awarded

**Stage 2**: **Conditional Mask Transformation**
- **Condition**: If there is at least **one mask symbol OR wild symbol in EACH of the first 3 reels** (reels 1, 2, 3)
- **Wild Substitution**: Wild symbols count as masks for the transformation condition
- **Transformation**: ALL mask symbols in the ENTIRE 5x4 grid transform to one randomly selected mask type (Wild symbols remain unchanged)
- **Additional Wins**: New wins are calculated on the transformed grid and added to Stage 1 wins
- **RNG Compliance**: Both stages are pre-calculated and the total payout is RNG-approved before awarding

**Example Flow**:
1. Stage 1: Normal spin awards 5.00 with masks in reels 1 & 2, wild in reel 3
2. Transformation check: ✅ Reel 1 has mask, ✅ Reel 2 has mask, ✅ Reel 3 has wild (counts as mask)
3. Stage 2: All masks transform to "PurpleMask", awards additional 100.00
4. Total payout: 5.00 + 100.00 = 105.00 (if RNG approves)

**Transformation Condition Examples**:
- ✅ **TRIGGERS**: Reel 1: PurpleMask, Reel 2: Wild, Reel 3: BlueMask
- ✅ **TRIGGERS**: Reel 1: Wild, Reel 2: OrangeMask, Reel 3: Wild  
- ❌ **NO TRIGGER**: Reel 1: PurpleMask, Reel 2: A, Reel 3: BlueMask (missing mask/wild in reel 2)

### Mask Reel Bonus
- **Trigger**: 3 or more Mask Reel symbols appear on reels 3, 4, and 5 (one per reel maximum)
- **Availability**: Base game only (does NOT appear during Free Spin Bonus)
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

#### Base Game Response Format
```json
{
  "stage1_reels": [
    ["Symbol1", "Symbol2", "Symbol3", "Symbol4"],
    ["Symbol1", "Symbol2", "Symbol3", "Symbol4"],
    ["Symbol1", "Symbol2", "Symbol3", "Symbol4"],
    ["Symbol1", "Symbol2", "Symbol3", "Symbol4"],
    ["Symbol1", "Symbol2", "Symbol3", "Symbol4"]
  ],
  "stage1_win_amount": 3.50,
  "stage1_win_details": [
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
  "stage2_win_amount": 0,
  "total_win_amount": 3.50,
  "mask_transformation_used": false,
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

### 2. Free Spin with Mask Transformation

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

#### Free Spin Response with Transformation
```json
{
  "stage1_reels": [
    ["PurpleMask", "A", "K", "Q"],
    ["OrangeMask", "10", "J", "Wild"],
    ["BlueMask", "Wild", "A", "K"],
    ["GreenMask", "K", "Q", "10"],
    ["YellowMask", "J", "A", "Wild"]
  ],
  "stage1_win_amount": 5.00,
  "stage1_win_details": [
    {
      "symbol": "A",
      "count": 3,
      "payout": 3.00,
      "positions": [
        {"reel": 0, "row": 1},
        {"reel": 2, "row": 2},
        {"reel": 4, "row": 2}
      ]
    }
  ],
  "stage2_reels": [
    ["PurpleMask", "A", "K", "Q"],
    ["PurpleMask", "10", "J", "Wild"],
    ["PurpleMask", "Wild", "A", "K"],
    ["PurpleMask", "K", "Q", "10"],
    ["PurpleMask", "J", "A", "Wild"]
  ],
  "stage2_win_amount": 100.00,
  "stage2_win_details": [
    {
      "symbol": "PurpleMask",
      "count": 5,
      "payout": 100.00,
      "positions": [
        {"reel": 0, "row": 0},
        {"reel": 1, "row": 0},
        {"reel": 2, "row": 0},
        {"reel": 3, "row": 0},
        {"reel": 4, "row": 0}
      ]
    }
  ],
  "total_win_amount": 105.00,
  "mask_transformation_used": true,
  "selected_mask_type": "PurpleMask",
  "bonus_count": 0,
  "bonus_win_amount": 0,
  "bonus_positions": [],
  "mask_reel_count": 0,
  "mask_reel_positions": [],
  "free_spin_triggered": false,
  "free_spin_retriggered": false,
  "mask_reel_triggered": false,
  "is_free_spin": true,
  "remaining_free_spins": 9,
  "current_free_spin_index": 1,
  "total_free_spins_awarded": 10,
  "bet_amount": 1.0,
  "bet_multiplier": 2
}
```

#### Free Spin Response without Transformation
```json
{
  "stage1_reels": [
    ["PurpleMask", "A", "K", "Q"],
    ["OrangeMask", "10", "J", "Wild"],
    ["A", "Wild", "K", "J"],
    ["Q", "K", "J", "10"],
    ["BlueMask", "J", "Wild", "K"]
  ],
  "stage1_win_amount": 7.50,
  "stage1_win_details": [...],
  "stage2_win_amount": 0,
  "total_win_amount": 7.50,
  "mask_transformation_used": false,
  "bonus_count": 0,
  "bonus_win_amount": 0,
  "bonus_positions": [],
  "mask_reel_count": 0,
  "mask_reel_positions": [],
  "free_spin_triggered": false,
  "free_spin_retriggered": false,
  "mask_reel_triggered": false,
  "is_free_spin": true,
  "remaining_free_spins": 8,
  "current_free_spin_index": 2,
  "total_free_spins_awarded": 10,
  "bet_amount": 1.0,
  "bet_multiplier": 2
}
```

### 3. Mask Reel Bonus

When the Mask Reel Bonus is triggered in base game, call the separate bonus endpoint:

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

#### Stage Results
| Parameter | Type | Description |
|-----------|------|-------------|
| stage1_reels | [][]string | 5x4 grid of symbols from normal spin |
| stage1_win_amount | float | Win amount from Stage 1 (normal spin) |
| stage1_win_details | []WinDetail | Details of Stage 1 winning combinations |
| stage2_reels | [][]string | 5x4 grid after mask transformation (only if transformation occurred) |
| stage2_win_amount | float | Additional win amount from Stage 2 (transformation) |
| stage2_win_details | []WinDetail | Details of Stage 2 winning combinations |

#### Combined Results
| Parameter | Type | Description |
|-----------|------|-------------|
| total_win_amount | float | Combined win from both stages |
| mask_transformation_used | bool | Whether mask transformation occurred |
| selected_mask_type | string | Which mask type was selected for transformation |

#### Feature Information
| Parameter | Type | Description |
|-----------|------|-------------|
| bonus_count | int | Number of bonus symbols on reels 1-3 |
| bonus_win_amount | float | Win amount from bonus symbols (2x bet for 3 symbols) |
| bonus_positions | []Position | Positions of bonus symbols |
| mask_reel_count | int | Number of mask reel symbols on reels 3-5 |
| mask_reel_positions | []Position | Positions of mask reel symbols |
| free_spin_triggered | bool | Whether free spins were triggered |
| free_spin_retriggered | bool | Whether free spins were retriggered |
| mask_reel_triggered | bool | Whether mask reel bonus was triggered |

#### Game State
| Parameter | Type | Description |
|-----------|------|-------------|
| is_free_spin | bool | Whether we're in free spin mode |
| remaining_free_spins | int | Number of free spins remaining |
| current_free_spin_index | int | Current index of free spin (0-based) |
| total_free_spins_awarded | int | Total number of free spins awarded |
| bet_amount | float | The bet amount used for this spin |
| bet_multiplier | int | The bet multiplier derived from bet amount |

## Implementation Guidelines

### 1. Game Flow
1. **Base Game**: Player spins and can trigger either Free Spin Bonus or Mask Reel Bonus
2. **Free Spin Bonus**: 10 free spins with potential mask transformation, can retrigger
3. **Mask Reel Bonus**: Immediate multiplier win, only in base game

### 2. Handling Two-Stage Free Spins
- **Stage 1 Display**: Show `stage1_reels` with `stage1_win_amount` and celebrate wins
- **Check Transformation**: If `mask_transformation_used: true`, prepare for Stage 2
- **Stage 2 Display**: Show `stage2_reels` with mask transformation animation to `selected_mask_type`
- **Stage 2 Wins**: Display `stage2_win_amount` and `stage2_win_details`
- **Total Celebration**: Show combined `total_win_amount`

### 3. Mask Transformation Logic
**Transformation occurs when**:
- It's a free spin AND
- At least one mask OR wild exists in reel 1 AND
- At least one mask OR wild exists in reel 2 AND  
- At least one mask OR wild exists in reel 3

**Important**: Wild symbols count as masks for the transformation condition but are NOT transformed - they remain as Wild symbols.

**Visual Flow**:
1. Show Stage 1 result
2. Highlight mask and wild symbols in first 3 reels that triggered transformation
3. Transform ONLY mask symbols in entire grid to selected type (Wild symbols stay Wild)
4. Calculate and show additional wins
5. Display total combined payout

### 4. State Management
- Always use response values for next request
- Free spins maintain the same bet amount throughout
- Mask Reel Bonus does not appear during free spins
- Track both stage wins separately and combined total

### 5. Visual Presentation
- **Stage 1**: Highlight wins using `stage1_win_details[].positions`
- **Stage 2**: Show transformation animation and highlight `stage2_win_details[].positions`
- **Bonus Symbols**: Show using `bonus_positions` 
- **Bonus Payout**: Display `bonus_win_amount` separately
- **Mask Reel Symbols**: Show using `mask_reel_positions`
- **Total Win**: Always display `total_win_amount` as final result

## Bet Amounts and Multipliers

The game uses the following bet amounts and their corresponding multipliers:

| Bet Amount | Bet Multiplier | Total Credits |
|------------|----------------|---------------|
| 0.50       | 1              | 50            |
| 1.00       | 2              | 100           |
| 2.50       | 5              | 250           |
| 5.00       | 10             | 500           |
| 12.50      | 25             | 1250          |

## Error Handling

Common error responses:
- "Invalid bet amount" - Bet amount not in allowed values
- "client_id is required" - Missing required field
- "Failed to retrieve game settings" - Settings service issue
- "Failed to determine outcome" - RNG service issue
- "Failed to process mask transformation" - Two-stage logic error

All errors return:
```json
{
  "status": "error",
  "message": "Error description"
}
```

## Key Differences from Standard Slot Games

1. **Two-Stage Free Spins**: Each free spin can have two separate result stages
2. **Conditional Transformation**: Mask transformation only occurs when specific conditions are met
3. **Separate Win Tracking**: Stage 1 and Stage 2 wins are tracked separately but combined for total payout
4. **Enhanced Visual Flow**: Frontend must handle two-stage reveal animations
5. **RNG Compliance**: Total combined payout is pre-approved by RNG system