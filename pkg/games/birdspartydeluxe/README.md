# Birds Party Deluxe Game - API Integration Guide for Unity Developers

## Overview

This document provides comprehensive guidelines for integrating the Birds Party Deluxe game backend API with a Unity frontend. The game features a **dynamic grid progression system** (4x4 → 5x5 → 6x6), **connection-based clover mechanics**, **progressive booming reels multiplier system**, cluster-based connections, level advancement mechanics, cascading systems, and a Free Spin Bonus feature.

## API Endpoints

- Base URL: `https://b.api.ibibe.africa`
- Spin endpoint: `POST /spin/birdspartydeluxe`
- Stage-cleared processing: `POST /process-stage-cleared/birdspartydeluxe`
- Cascade endpoint: `POST /cascade/birdspartydeluxe`
- Health check: `GET /status`

## Game Mechanics

### Core Game Rules
- **Dynamic grid structure**: 4x4 → 5x5 → 6x6 based on level progression
- **Stage-cleared symbol priority removal**: Special symbols removed first before connections
- **Connection-based cluster system**: Horizontal/vertical adjacent symbols (birds + clovers)
- **Progressive booming reels**: X2 → X3 → X4 → X5 → X10 multiplier system
- 3-level progression system with automatic grid expansion
- Cascading mechanics with symbol removal and gravity
- Denomination: 0.01
- Bet amounts: 0.1, 0.2, 0.3, 0.5, 1.0 (corresponding to multipliers: 1, 2, 3, 5, 10)
- Minimum bet: 10 credits per bet multiplier

### Symbols

#### Regular Bird Symbols (Form Connections)
- Purple Owl, Green Owl, Yellow Owl, Blue Owl, Red Owl, **Four-leaf Clover** (19% each)
- **Connection-forming symbols** that create clusters and award payouts

#### Special Symbols (DELUXE: Rainbow Egg Only)
- **Rainbow Egg (`free_game`)** - Triggers 10 free spins (1% probability)

#### Stage-Cleared Symbols (Priority Removal)
- **Level 1**: `orange_slice` - Orange slice symbol (5% probability on 4x4 grid)
- **Level 2**: `honey_pot` - Honey pot symbol (5% probability on 5x5 grid)  
- **Level 3**: `strawberry` - Strawberry symbol (5% probability on 6x6 grid)

**IMPORTANT**: Stage-cleared symbols do NOT form connections. They are removed individually when they appear.

### DELUXE: Connection-Based Clover Mechanics

#### How Clovers Work
1. **Clovers form connections** like regular bird symbols (minimum 4/5/6 depending on level)
2. **When clovers form valid connections**:
   - They **upgrade the booming reels multiplier** (1x → 2x → 3x → 4x → 5x → 10x)
   - They **pay base value only** (same as Purple Owl payouts, NO multiplier applied)
   - **They do NOT use the multiplier for their own payouts**
3. **Subsequent bird connections** in the same cascade sequence use the upgraded multiplier
4. **Multiplier resets** when cascade sequence ends (no more connections)

#### Multiplier Progression Example
```
Initial: 1x multiplier
Clover connection found (4 symbols) → Upgrade to 2x → Clover pays: base_payout (no multiplier)
Bird connection found (5 symbols) → Bird pays: base_payout × 2x  
Another clover connection → Upgrade to 3x → Clover pays: base_payout (no multiplier)
Bird connection pays: base_payout × 3x
No more connections → Reset to 1x for next spin
```

### Dynamic Grid & Level System
- **Level 1**: 4x4 grid (16 positions), minimum 4 connected symbols required
- **Level 2**: 5x5 grid (25 positions), minimum 5 connected symbols required  
- **Level 3**: 6x6 grid (36 positions), minimum 6 connected symbols required
- **Progression**: Accumulate 15 stage-cleared symbols to advance to next level
- **Cycling**: After Level 3, returns to Level 1 (infinite progression)
- **Grid Expansion**: Grid automatically resizes when advancing levels

### Free Spin System (DELUXE Differences)

#### Free Spin Triggering
- **Rainbow Egg symbol** triggers 10 free spins
- **DELUXE**: Free spins do NOT re-trigger during free spins mode
- **Original**: Free spins could be re-triggered during bonus

#### Free Spin Features
- **Booming reels continue** during free spins
- **Stage-cleared symbols continue** to appear and advance levels
- **Clover connections continue** to upgrade multiplier
- **No cost** for free spin rounds

### Stage-Cleared Symbol Mechanics

#### Priority Removal System
1. **Stage-cleared symbols are detected** when grid is generated
2. **Priority removal**: Stage-cleared symbols are removed FIRST (ignore all connections)
3. **Gravity applied**: Symbols fall down, new symbols generated at top
4. **Then check connections**: Regular bird and clover symbol connections checked on new grid
5. **Count toward progress**: Each removed stage-cleared symbol counts toward 15

#### Level-Specific Appearance
- **Only level-appropriate symbols appear**: Orange slice only on Level 1, etc.
- **Individual removal**: Each stage-cleared symbol is removed separately
- **No connection rules**: Stage-cleared symbols don't need to be connected
- **Progress tracking**: Each removed symbol = +1 toward level advancement

### Three-Endpoint Game Flow

#### 1. Spin Phase - `/spin/birdspartydeluxe`
- Generates grid with potential connection-forming symbol connections (birds + clovers)
- **Identifies stage-cleared symbols** (does NOT remove them)
- **Processes clover connections first** → upgrades booming reels multiplier
- **Processes bird connections** → applies current multiplier to payouts
- **Checks for rainbow egg symbols** → triggers free spins if found
- **RNG validates bird connections only** (clovers always upgrade multiplier)
- **Sets cascading flag** if any connections exist
- Returns grid with stage-cleared symbol positions AND connection info

#### 2. Stage-Cleared Processing - `/process-stage-cleared/birdspartydeluxe`
- **Removes all stage-cleared symbols** from grid
- **Applies gravity** and fills with new symbols
- **Updates stage progress** count and checks for level advancement
- **Processes new clover connections** → upgrades multiplier AND pays out
- **Checks for NEW bird connections** → applies multiplier to payouts
- **RNG validates new paying connections** and sets cascading flag
- Returns updated grid with potential new connections

#### 3. Cascade Phase - `/cascade/birdspartydeluxe`
- **Processes connection-forming symbols** from previous steps
- **Processes clover connections first** → upgrades multiplier AND pays out
- **Processes bird connections** → applies current multiplier
- **ENHANCED**: Detects stage-cleared symbols that appear after gravity
- **Handles cascading mechanics** (remove → gravity → find new connections)
- **Resets multiplier** when no more connections exist (cascade sequence ends)
- **Returns stage-cleared detection info** for client to process via stage-cleared endpoint
- Continues until no more connections exist
- Handles RNG integration for subsequent paying connections

## API Interaction Flow

### 1. Basic Spin with DELUXE Mechanics

#### Initial Spin Request
```json
POST /spin/birdspartydeluxe
{
  "client_id": "client_id_here",
  "game_id": "birdspartydeluxe",
  "player_id": "player_id_here", 
  "bet_id": "bet_id_here",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [],
    "stageProgress": 5,
    "gameMode": "base",
    "freeSpins": { 
      "remaining": 0, 
      "totalAwarded": 0,
      "boomingReelsLevel": 0,
      "currentMultiplier": 1.0,
      "cloverConnectionsFound": 0
    },
    "totalWin": 0,
    "cascading": false,
    "lastConnections": [],
    "cascadeCount": 0,
    "stageClearedSymbols": []
  }
}
```

#### Spin Response with DELUXE Features
```json
{
  "status": "success",
  "message": "",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [
      ["orange_slice", "clover", "yellow_owl", "green_owl"],
      ["red_owl", "clover", "clover", "blue_owl"],
      ["blue_owl", "clover", "purple_owl", "red_owl"],
      ["purple_owl", "yellow_owl", "blue_owl", "green_owl"]
    ],
    "stageProgress": 5,
    "gameMode": "base",
    "freeSpins": { 
      "remaining": 0, 
      "totalAwarded": 0,
      "boomingReelsLevel": 1,
      "currentMultiplier": 2.0,
      "cloverConnectionsFound": 1
    },
    "totalWin": 0.08,
    "cascading": true,
    "lastConnections": [
      {
        "symbol": "clover",
        "positions": [
          {"x": 1, "y": 0}, {"x": 1, "y": 1}, {"x": 2, "y": 1}, {"x": 1, "y": 2}
        ],
        "count": 4,
        "payout": 0.02
      },
      {
        "symbol": "yellow_owl", 
        "positions": [
          {"x": 2, "y": 0}, {"x": 1, "y": 3}
        ],
        "count": 4,
        "payout": 0.08
      }
    ],
    "cascadeCount": 0,
    "stageClearedSymbols": [
      { "symbol": "orange_slice", "position": {"x": 0, "y": 0} }
    ]
  },
  "stageClearedSymbols": [
    { "symbol": "orange_slice", "position": {"x": 0, "y": 0} }
  ],
  "hasStageCleared": true,
  "totalCost": 0.1
}
```

### 2. Processing Stage-Cleared Symbols with Multiplier Continuity

#### Stage-Cleared Processing Request
```json
POST /process-stage-cleared/birdspartydeluxe
{
  "client_id": "client_id_here",
  "game_id": "birdspartydeluxe",
  "player_id": "player_id_here",
  "bet_id": "stage_cleared_001",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [
      ["orange_slice", "clover", "yellow_owl", "green_owl"],
      ["red_owl", "clover", "clover", "blue_owl"],
      ["blue_owl", "clover", "purple_owl", "red_owl"],
      ["purple_owl", "yellow_owl", "blue_owl", "green_owl"]
    ],
    "stageProgress": 5,
    "gameMode": "base",
    "freeSpins": { 
      "remaining": 0, 
      "totalAwarded": 0,
      "boomingReelsLevel": 1,
      "currentMultiplier": 2.0,
      "cloverConnectionsFound": 1
    },
    "totalWin": 0.08,
    "cascading": true,
    "lastConnections": [...],
    "cascadeCount": 0,
    "stageClearedSymbols": [
      { "symbol": "orange_slice", "position": {"x": 0, "y": 0} }
    ]
  }
}
```

#### Stage-Cleared Processing Response
```json
{
  "status": "success",
  "message": "",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [
      ["green_owl", "clover", "yellow_owl", "green_owl"],
      ["red_owl", "clover", "clover", "blue_owl"],
      ["blue_owl", "clover", "purple_owl", "red_owl"],
      ["purple_owl", "yellow_owl", "blue_owl", "green_owl"]
    ],
    "stageProgress": 6,
    "gameMode": "base",
    "freeSpins": { 
      "remaining": 0, 
      "totalAwarded": 0,
      "boomingReelsLevel": 1,
      "currentMultiplier": 2.0,
      "cloverConnectionsFound": 1
    },
    "totalWin": 0.08,
    "cascading": true,
    "lastConnections": [...],
    "cascadeCount": 1,
    "stageClearedSymbols": []
  },
  "stageClearedCount": 1,
  "levelAdvanced": false,
  "connections": [...],
  "totalCost": 0
}
```

### 3. Enhanced Cascade Processing with Booming Reels

#### Cascade Request
```json
POST /cascade/birdspartydeluxe
{
  "client_id": "client_id_here",
  "game_id": "birdspartydeluxe", 
  "player_id": "player_id_here",
  "bet_id": "cascade_001",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [...],
    "stageProgress": 6,
    "gameMode": "base",
    "freeSpins": { 
      "remaining": 0, 
      "totalAwarded": 0,
      "boomingReelsLevel": 1,
      "currentMultiplier": 2.0,
      "cloverConnectionsFound": 1
    },
    "totalWin": 0.08,
    "cascading": true,
    "lastConnections": [...],
    "cascadeCount": 1,
    "stageClearedSymbols": []
  }
}
```

#### Enhanced Cascade Response with Multiplier Reset
```json
{
  "status": "success",
  "message": "",
  "gameState": {
    "bet": { "amount": 0.1, "multiplier": 1 },
    "currentLevel": 1,
    "gridSize": 4,
    "grid": [...],
    "stageProgress": 6,
    "gameMode": "base",
    "freeSpins": { 
      "remaining": 0, 
      "totalAwarded": 0,
      "boomingReelsLevel": 0,
      "currentMultiplier": 1.0,
      "cloverConnectionsFound": 0
    },
    "totalWin": 0.0,
    "cascading": false,
    "lastConnections": [],
    "cascadeCount": 2
  },
  "connections": [],
  "stageClearedSymbols": [],
  "hasStageCleared": false,
  "totalCost": 0
}
```

## Error Handling

### Stage-Cleared Processing Errors
```json
{
  "status": "error",
  "message": "Invalid grid dimensions for level 2"
}
```

### Missing Stage-Cleared Symbols
```json
{
  "status": "error", 
  "message": "No stage-cleared symbols found to process"
}
```

### Common Errors
- "Invalid bet amount" - Bet amount not in allowed values (0.1, 0.2, 0.3, 0.5, 1.0)
- "client_id is required" - Missing required field
- "Failed to retrieve game settings" - Settings service issue
- "Failed to determine outcome" - RNG service issue

## DELUXE vs Original Differences

### Key DELUXE Features
1. **Connection-Based Clovers**: Clovers must form connections to upgrade multiplier
2. **Progressive Booming Reels**: X2 → X3 → X4 → X5 → X10 multiplier progression  
3. **Separate Special Symbols**: Rainbow egg (free spins) vs Clover (multiplier)
4. **No Free Spin Re-triggering**: Rainbow eggs don't trigger more free spins during free spins
5. **Multiplier Reset Logic**: Resets when cascade sequence ends
6. **Stage-Cleared During Free Spins**: Level progression continues during bonus

### Original vs DELUXE Comparison
| Feature | Original Birds Party | Birds Party DELUXE |
|---------|---------------------|-------------------|
| Free Spin Trigger | Free game symbol | Rainbow egg symbol |
| Multiplier System | Random 1.0-5.0x fixed | Progressive 2x-10x connection-based |
| Free Spin Re-trigger | Yes | No |
| Stage-Cleared in Free Spins | Forbidden | Allowed |
| Multiplier Progression | None | Clover connections upgrade |
| Special Symbol Count | 1 (multi-purpose) | 2 (separate functions) |

## Testing and Debugging

### Debug Information
- Monitor server logs for clover connection detection and multiplier upgrades
- Track booming reels progression through cascade sequences
- Verify multiplier reset timing at cascade sequence end
- Check separation between rainbow eggs and clovers
- **NEW**: Monitor connection-based clover mechanics
- **NEW**: Track multiplier application to bird connections only

### Test Scenarios
1. **Connection-Based Clover Mechanics**: Verify clovers must form connections to upgrade
2. **Progressive Multiplier**: Test X2 → X3 → X4 → X5 → X10 progression
3. **Multiplier Reset**: Verify reset when cascade sequence ends
4. **Rainbow Egg vs Clover**: Test separate handling of special symbols
5. **No Free Spin Re-trigger**: Verify rainbow eggs don't re-trigger during free spins
6. **Bird Connection Multipliers**: Verify bird payouts use current multiplier
7. **Clover Connection Payouts**: Verify clovers pay base value only (no multiplier applied)
8. **Stage-Cleared During Free Spins**: Test level progression continues
9. **Level Advancement with Multiplier**: Test multiplier preservation during level up
10. **Complex Cascade Sequences**: Test multiple clover upgrades in one sequence

### Performance Considerations
- **Three-Endpoint Flow**: Ensure smooth transitions between endpoints
- **Grid Resizing**: Optimize UI transitions when changing grid sizes  
- **Symbol Management**: Efficient loading/unloading of level-specific symbols
- **Animation Sequencing**: Coordinate stage-cleared removal with gravity effects
- **ENHANCED**: Multiplier Progression: Smooth visual feedback for booming reels upgrades
- **ENHANCED**: Connection Processing: Handle clover vs bird connection separation efficiently
- **ENHANCED**: Cascade Sequence Management: Track multiplier state throughout sequences

### Debug Output Examples

#### DELUXE Cascade Flow with Booming Reels
```
Starting DELUXE cascade sequence
Reset booming reels to 1.0x for new spin
Processing clover connections: 1 found
Booming reels upgraded to 2.0x
Clover pays base value only (no multiplier)
Processing bird connections: 2 found  
Bird payouts calculated with 2.0x multiplier
Processing cascade #1
Found 1 clover connection
Booming reels upgraded to 3.0x
Clover pays base value only (no multiplier)
Found 2 bird connections
Bird payouts calculated with 3.0x multiplier
Processing cascade #2
Found 0 connections
Cascade sequence ended, resetting booming reels to 1.0x
DELUXE cascade sequence completed
```

#### Level Advancement with Multiplier Preservation
```
Starting DELUXE cascade sequence
Booming reels at 3.0x from previous cascades
Stage-cleared processing: 4 symbols
Level advanced from 1 to 2, grid: 5x5
Found 1 clover connection on new level
Booming reels upgraded to 4.0x
Clover pays base value only (no multiplier)
Found 1 bird connection on new level
Bird payout calculated with 4.0x multiplier
DELUXE level advancement completed
```

## Key Enhancements in DELUXE Version

### 1. Connection-Based Clover System
- **Clover Connection Detection**: Clovers must form valid connections (4/5/6+ symbols)
- **Multiplier Upgrade Logic**: Each clover connection upgrades booming reels
- **Payout Separation**: Clovers pay base value only (no multiplier), birds receive multiplied payouts
- **RNG Integration**: Only bird connections validated by RNG

### 2. Progressive Booming Reels
- **Sequential Upgrades**: 1x → 2x → 3x → 4x → 5x → 10x progression
- **Cascade Continuity**: Multiplier persists throughout cascade sequence
- **Reset Logic**: Resets when no more connections found
- **Visual Feedback**: Clear indication of current multiplier level

### 3. Enhanced Symbol Management
- **Dual Special Symbols**: Rainbow egg and clover with distinct functions
- **Connection-Forming Logic**: Birds and clovers can form connections
- **Stage-Cleared Priority**: Stage-cleared symbols processed first
- **Free Spin Mechanics**: No re-triggering during bonus rounds
- **Clover Payout Logic**: Clovers pay base value only, never use multiplier

### 4. Architectural Benefits
- **Maintainable Code**: Each endpoint has clear, distinct responsibilities
- **Flexible Integration**: Client can orchestrate complex flows as needed
- **RNG Integrity**: Proper RNG validation for bird connections only
- **Scalable Design**: Easy to add new features or modify individual components

## Advanced Flow Patterns

### Pattern 1: Simple Cascade without Multiplier
```
Spin (1x) → Cascade (bird connections only) → Cascade (more birds) → End
```

### Pattern 2: Clover Progression Flow
```
Spin (1x) → Clover connection (2x, pays base) → Cascade (birds at 2x) → Clover (3x, pays base) → Birds (3x) → End
```

### Pattern 3: Complex Mixed Flow with Level Up
```
Spin (1x) → Clover (2x, pays base) → Process-Stage-Cleared (level up) → 
Clover (3x, pays base) → Cascade (birds at 3x) → End
```

### Pattern 4: Maximum Multiplier Flow
```
Spin (1x) → Clover (2x, pays base) → Clover (3x, pays base) → Clover (4x, pays base) → Clover (5x, pays base) → 
Clover (10x, pays base) → Cascade (birds at 10x) → End
```

## Implementation Notes

### Backend Responsibilities
1. **Spin Handler**: Grid generation, clover/bird separation, multiplier management
2. **Process-Stage-Cleared Handler**: Stage-cleared removal, level advancement, multiplier preservation
3. **Cascade Handler**: Connection processing, multiplier progression, sequence management

### Client Responsibilities
1. **Flow Orchestration**: Coordinate between endpoints based on response flags
2. **Animation Management**: Handle visual transitions for multiplier changes
3. **State Management**: Maintain booming reels state throughout sequences
4. **User Experience**: Provide clear feedback for multiplier progression

### Critical Success Factors
1. **Connection Detection**: Ensure clovers form valid connections before upgrading
2. **Multiplier Tracking**: Maintain accurate multiplier throughout cascade sequences
3. **Symbol Separation**: Never mix clover upgrades with bird payouts
4. **Clover Payout Logic**: Clovers must pay base value only, never use multiplier
5. **Reset Timing**: Properly reset multiplier when cascade sequences end
6. **Visual Clarity**: Clear indication of booming reels progression for players

This DELUXE implementation provides a sophisticated, engaging slot game experience with progressive multiplier mechanics that reward players for finding clover connections while maintaining the core Birds Party gameplay that players love!