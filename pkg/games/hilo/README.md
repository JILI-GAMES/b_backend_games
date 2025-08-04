# Hilo Game API - Enhanced Implementation

## Overview

This is an enhanced implementation of the Hilo card prediction game with pre-calculated betting options, full RNG integration for house edge control, and Unity frontend integration.

## Key Features

### ✅ Unity Frontend Integration
- **Card Input Support**: Unity can specify starting cards
- **Fallback Safety**: Random card generation if no card provided
- **Consistent Format**: Returns exactly what Unity sends
- **Logging**: Clear tracking of card source (Unity vs Random)
- **Deck Consistency**: Proper deck generation for Unity cards
- **State Verification**: Fixed "Invalid game state" errors

### ✅ Decimal Precision Control
- **2 Decimal Places**: All monetary values consistently rounded
- **No Floating Point Errors**: Eliminates precision issues in JSON responses
- **Consistent Formatting**: Clean, professional financial display

### ✅ Enhanced Skip Logic
- **First Card Infinite Skips**: Players can skip first card unlimited times
- **5-Skip Limit**: Other cards limited to 5 skips per game
- **Smart Counting**: Only counts skips for cards after position 0

### ✅ Pre-calculated Betting System
- **Exact UI Match**: Five buttons (HIGH, LOW, SAME, HIGH OR SAME, LOW OR SAME)
- **Pre-calculated Multipliers**: Consistent calculations matching your UI
- **Smart Button States**: Automatically disabled when impossible
- **Performance Optimized**: No real-time calculations during gameplay

### ✅ Full RNG Integration
- **House Protection**: RNG can force losses when payouts are too high
- **Player Retention**: Occasional forced wins for engagement
- **Real-time Risk Management**: Every bet checked against house limits
- **Audit Trail**: All RNG decisions logged with IP and User-Agent

### ✅ Enhanced Security
- **HMAC Signatures**: Prevent game state tampering
- **Provable Fairness**: Deterministic deck generation
- **Request Validation**: Comprehensive input validation
- **Signature Verification**: All game state changes verified

## Project Structure

```
cmd/hilo/
├── main.go                    # Application entry point

pkg/games/hilo/
├── types.go                   # Data structures and request/response types
├── game.go                    # Core game logic (deck generation, HMAC, etc.)
├── betting.go                 # Pre-calculated betting system
├── handlers.go                # HTTP handlers for all endpoints
├── routes.go                  # Route registration and client selection
└── utils.go                   # Utility functions

pkg/common/
├── config/
│   └── config.go             # Configuration management
├── rng/
│   └── client.go             # RNG service client with IP/User-Agent
└── settings/
    └── client.go             # Settings service client
```

## API Endpoints

### Main Game Endpoints
- `POST /start/hilo` - Start a new game (with optional Unity card input)
- `POST /guess/hilo` - Make a guess (with RNG control)
- `POST /skip/hilo` - Skip current card (enhanced logic)
- `POST /cashout/hilo` - Cash out current winnings
- `POST /verify/hilo` - Verify deck fairness

### Utility Endpoints
- `POST /options/hilo` - Get betting options for a card
- `GET /status` - Server status

## Installation & Setup

1. **Install Dependencies**
```bash
go mod tidy
```

2. **Configure Environment**
```bash
cp .env.example .env
# Edit .env with your settings
```

3. **Run the Server**
```bash
go run cmd/hilo/main.go
```

## API Usage Examples

### 1. Start Game (with Unity Card Input)
```bash
curl -X POST http://localhost:11400/start/hilo \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "demo_client",
    "game_id": "hilo_001",
    "player_id": "player_123",
    "bet_id": "bet_456",
    "bet_amount": 5.00,
    "card": "ACE_SPADES"
  }'
```

**Response:**
```json
{
  "status": "success",
  "message": "Game started successfully",
  "game_state": {
    "seed": "1720000123456_987654321",
    "current_card": "ACE_SPADES",
    "unity_card": "ACE_SPADES",
    "position": 0,
    "accumulated_win": 1.00,
    "bet_amount": 5.00,
    "skips_remaining": 5
  },
  "bet_options": [
    {"id": "higher", "name": "HIGH", "multiplier": 1.85, "is_enabled": true},
    {"id": "lower", "name": "LOW", "multiplier": 0.00, "is_enabled": false},
    {"id": "same", "name": "SAME", "multiplier": 16.69, "is_enabled": true},
    {"id": "higher_or_same", "name": "HIGH OR SAME", "multiplier": 1.79, "is_enabled": true},
    {"id": "lower_or_same", "name": "LOW OR SAME", "multiplier": 16.69, "is_enabled": true}
  ],
  "signature": "hmac_signature..."
}
```

### 2. Start Game (without card - random generation)
```bash
curl -X POST http://localhost:11400/start/hilo \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "demo_client",
    "game_id": "hilo_001",
    "player_id": "player_123",
    "bet_id": "bet_456",
    "bet_amount": 5.00
  }'
```

### 3. Make Guess
```bash
curl -X POST http://localhost:11400/guess/hilo \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "demo_client",
    "game_id": "hilo_001",
    "player_id": "player_123",
    "bet_id": "bet_789",
    "game_state": { /* from previous response */ },
    "bet_choice": "higher_or_same",
    "signature": "hmac_signature_from_previous_response"
  }'
```

**Response:**
```json
{
  "status": "success",
  "game_state": {
    "current_card": "J♦",
    "position": 1,
    "accumulated_win": 1.79,
    "is_game_over": false
  },
  "guess_result": {
    "success": true,
    "next_card": "J♦",
    "next_value": 11,
    "payout_multiple": 1.79,
    "was_correct": true,
    "forced": false
  },
  "bet_options": [ /* options for next round */ ],
  "signature": "new_hmac_signature..."
}
```

### 4. Skip Card (Enhanced Logic)
```bash
curl -X POST http://localhost:11400/skip/hilo \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "demo_client",
    "game_id": "hilo_001",
    "player_id": "player_123",
    "bet_id": "bet_890",
    "game_state": { /* current game state */ },
    "signature": "hmac_signature_from_previous_response"
  }'
```

**Response:**
```json
{
  "status": "success",
  "message": "Skipped to next card: K♥",
  "game_state": {
    "current_card": "K♥",
    "position": 2,
    "accumulated_win": 1.79,
    "skips_used": 1,
    "skips_remaining": 4,
    "game_history": [
      {"card": "ACE_SPADES", "value": 1, "position": 0},
      {"card": "J♦", "value": 11, "position": 1},
      {"card": "K♥", "value": 13, "position": 2}
    ],
    "is_game_over": false
  },
  "bet_options": [
    {"id": "higher", "name": "HIGH", "multiplier": 0.00, "is_enabled": false},
    {"id": "lower", "name": "LOW", "multiplier": 1.04, "is_enabled": true},
    {"id": "same", "name": "SAME", "multiplier": 16.69, "is_enabled": true},
    {"id": "higher_or_same", "name": "HIGH OR SAME", "multiplier": 16.69, "is_enabled": true},
    {"id": "lower_or_same", "name": "LOW OR SAME", "multiplier": 1.03, "is_enabled": true}
  ],
  "signature": "new_hmac_signature..."
}
```

### 5. Cash Out (with 2 decimal precision)
```bash
curl -X POST http://localhost:11400/cashout/hilo \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "demo_client",
    "game_id": "hilo_001",
    "player_id": "player_123",
    "bet_id": "bet_901",
    "game_state": { /* current game state */ },
    "signature": "hmac_signature_from_previous_response"
  }'
```

**Response:**
```json
{
  "status": "success",
  "message": "Cashed out successfully: 8.95",
  "final_win": 8.95,
  "game_state": {
    "current_card": "K♥",
    "position": 2,
    "accumulated_win": 1.79,
    "bet_amount": 5.00,
    "is_game_over": true,
    "final_win": 8.95
  }
}
```

### 6. Verify Deck
```bash
curl -X POST http://localhost:11400/verify/hilo \
  -H "Content-Type: application/json" \
  -d '{
    "seed": "1720000123456_987654321"
  }'
```

**Response:**
```json
{
  "status": "success",
  "seed": "1720000123456_987654321",
  "deck": [
    "ACE_SPADES", "J♦", "K♥", "4♣", "2♦", "A♠", "Q♥", "8♣", "3♠", "10♦",
    "5♥", "9♣", "6♠", "A♦", "K♣", "7♥", "J♠", "4♦", "2♣", "Q♠",
    "8♥", "3♦", "10♣", "5♠", "9♥", "6♦", "A♣", "K♦", "7♣", "J♥",
    "4♠", "2♥", "Q♦", "8♠", "3♣", "10♥", "5♦", "9♠", "6♣", "A♥",
    "K♠", "7♦", "J♣", "4♥", "2♠", "Q♣", "8♦", "3♥", "10♠", "5♣",
    "9♦", "6♥"
  ],
  "deck_hash": "abc123def456789..."
}
```

### 7. Get Betting Options (Utility Endpoint)
```bash
curl -X POST http://localhost:11400/options/hilo \
  -H "Content-Type: application/json" \
  -d '{
    "current_card": "ACE_SPADES"
  }'
```

**Response:**
```json
{
  "status": "success",
  "current_card": "ACE_SPADES",
  "current_value": 1,
  "bet_options": [
    {"id": "higher", "name": "HIGH", "multiplier": 1.85, "is_enabled": true},
    {"id": "lower", "name": "LOW", "multiplier": 0.00, "is_enabled": false},
    {"id": "same", "name": "SAME", "multiplier": 16.69, "is_enabled": true},
    {"id": "higher_or_same", "name": "HIGH OR SAME", "multiplier": 1.79, "is_enabled": true},
    {"id": "lower_or_same", "name": "LOW OR SAME", "multiplier": 16.69, "is_enabled": true}
  ]
}
```

### 8. Server Status
```bash
curl -X GET http://localhost:11400/status
```

**Response:**
```json
{
  "status": "ok",
  "game": "hilo",
  "version": "2.0"
}
```

## Unity Integration

### Card Format
Cards should be in the format: `"RANK_SUIT"` where:
- **Ranks**: `ACE`, `TWO`, `THREE`, ..., `TEN`, `JACK`, `QUEEN`, `KING`
- **Suits**: `SPADES`, `CLUBS`, `DIAMONDS`, `HEARTS`

**Examples:**
- `"ACE_SPADES"`
- `"KING_HEARTS"` 
- `"TEN_DIAMONDS"`
- `"JACK_CLUBS"`

### Important Notes
- ✅ **Fixed**: "Invalid game state" errors when using Unity cards
- ✅ **Fixed**: Card value parsing for `"RANK_SUIT"` format
- ✅ **Added**: Unity card tracking in game state
- ✅ **Enhanced**: Debug logging for troubleshooting

### Unity C# Example
```csharp
[System.Serializable]
public class StartGameRequest
{
    public string client_id;
    public string game_id;
    public string player_id;
    public string bet_id;
    public float bet_amount;
    public string card; // Optional - Unity can specify starting card
}

[System.Serializable]
public class GameState
{
    public string seed;
    public string deck_hash;
    public string current_card;
    public string unity_card; // Tracks if this was a Unity-specified card
    public int position;
    public float accumulated_win;
    public float bet_amount;
    public int skips_used;
    public int skips_remaining;
    public int max_skips;
    public Card[] game_history;
    public bool is_game_over;
    public float final_win;
}

[System.Serializable]
public class StartGameResponse
{
    public string status;
    public string message;
    public GameState game_state;
    public BetOption[] bet_options;
    public string signature;
}

public class HiloGameController : MonoBehaviour
{
    private string apiUrl = "http://localhost:11400";
    private GameState currentGameState;
    private string currentSignature;

    public async Task<StartGameResponse> StartGame(string clientId, string gameId, string playerId, float betAmount, string card = null)
    {
        var request = new StartGameRequest
        {
            client_id = clientId,
            game_id = gameId,
            player_id = playerId,
            bet_id = $"bet_{DateTimeOffset.UtcNow.ToUnixTimeMilliseconds()}",
            bet_amount = betAmount,
            card = card // Unity can specify the starting card
        };

        var json = JsonUtility.ToJson(request);
        var response = await PostRequest($"{apiUrl}/start/hilo", json);
        var gameResponse = JsonUtility.FromJson<StartGameResponse>(response);
        
        currentGameState = gameResponse.game_state;
        currentSignature = gameResponse.signature;
        
        return gameResponse;
    }

    // Example usage with Unity-specified card
    public async void StartGameWithCard()
    {
        var response = await StartGame("demo_client", "hilo_001", "player_123", 5.0f, "ACE_SPADES");
        Debug.Log($"Game started with card: {response.game_state.current_card}");
    }

    // Example usage with random card
    public async void StartGameRandom()
    {
        var response = await StartGame("demo_client", "hilo_001", "player_123", 5.0f);
        Debug.Log($"Game started with random card: {response.game_state.current_card}");
    }
}
```

## Frontend Integration

### JavaScript Example
```javascript
class HiloGame {
  constructor(apiUrl) {
    this.apiUrl = apiUrl;
    this.gameState = null;
    this.signature = null;
  }

  async startGame(clientId, gameId, playerId, betAmount, card = null) {
    const requestBody = {
      client_id: clientId,
      game_id: gameId,
      player_id: playerId,
      bet_id: `bet_${Date.now()}`,
      bet_amount: betAmount
    };

    // Add card if provided by Unity
    if (card) {
      requestBody.card = card;
    }

    const response = await fetch(`${this.apiUrl}/start/hilo`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify(requestBody)
    });

    const data = await response.json();
    this.gameState = data.game_state;
    this.signature = data.signature;

    // Log Unity card information for debugging
    if (data.game_state.unity_card) {
      console.log(`🎮 Unity card detected: ${data.game_state.unity_card}`);
    }

    // Update UI with betting options
    this.updateBettingButtons(data.bet_options);
    
    return data;
  }

  async makeGuess(betChoice) {
    const response = await fetch(`${this.apiUrl}/guess/hilo`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        client_id: this.gameState.client_id,
        game_id: this.gameState.game_id,
        player_id: this.gameState.player_id,
        bet_id: `bet_${Date.now()}`,
        game_state: this.gameState,
        bet_choice: betChoice,
        signature: this.signature
      })
    });

    const data = await response.json();
    
    if (data.status === 'success') {
      this.gameState = data.game_state;
      this.signature = data.signature;
      
      // Update UI
      this.updateGameState(data.game_state);
      this.showGuessResult(data.guess_result);
      
      if (!data.game_state.is_game_over) {
        this.updateBettingButtons(data.bet_options);
      }
    }
    
    return data;
  }

  async skipCard() {
    const response = await fetch(`${this.apiUrl}/skip/hilo`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        client_id: this.gameState.client_id,
        game_id: this.gameState.game_id,
        player_id: this.gameState.player_id,
        bet_id: `bet_${Date.now()}`,
        game_state: this.gameState,
        signature: this.signature
      })
    });

    const data = await response.json();
    
    if (data.status === 'success') {
      this.gameState = data.game_state;
      this.signature = data.signature;
      
      // Update UI
      this.updateGameState(data.game_state);
      this.updateBettingButtons(data.bet_options);
      this.updateSkipsRemaining(data.game_state.skips_remaining);
    }
    
    return data;
  }

  async cashOut() {
    const response = await fetch(`${this.apiUrl}/cashout/hilo`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        client_id: this.gameState.client_id,
        game_id: this.gameState.game_id,
        player_id: this.gameState.player_id,
        bet_id: `bet_${Date.now()}`,
        game_state: this.gameState,
        signature: this.signature
      })
    });

    const data = await response.json();
    
    if (data.status === 'success') {
      this.gameState = data.game_state;
      this.signature = null; // Game is over
      
      // Show final win amount (now properly formatted to 2 decimals)
      this.showCashOutResult(data.final_win);
      this.disableAllButtons();
    }
    
    return data;
  }

  async verifyDeck(seed) {
    const response = await fetch(`${this.apiUrl}/verify/hilo`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ seed: seed })
    });

    const data = await response.json();
    
    if (data.status === 'success') {
      // Show deck verification results
      this.showDeckVerification(data.deck, data.deck_hash);
    }
    
    return data;
  }

  async getBettingOptions(currentCard) {
    const response = await fetch(`${this.apiUrl}/options/hilo`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ current_card: currentCard })
    });

    const data = await response.json();
    
    if (data.status === 'success') {
      this.updateBettingButtons(data.bet_options);
    }
    
    return data;
  }

  updateBettingButtons(betOptions) {
    const buttonMapping = {
      "higher": "HIGH",
      "lower": "LOW", 
      "same": "SAME",
      "higher_or_same": "HIGH OR SAME",
      "lower_or_same": "LOW OR SAME"
    };

    betOptions.forEach(option => {
      const button = document.getElementById(`${option.id}_button`);
      if (button) {
        button.textContent = `${buttonMapping[option.id]}\n${option.multiplier.toFixed(2)}x`;
        button.disabled = !option.is_enabled;
        button.onclick = () => this.makeGuess(option.id);
      }
    });
  }

  updateGameState(gameState) {
    // Update current card display
    document.getElementById('current_card').textContent = gameState.current_card;
    
    // Update accumulated win multiplier (now properly formatted)
    document.getElementById('multiplier').textContent = `${gameState.accumulated_win.toFixed(2)}x`;
    
    // Update potential win amount (now properly formatted)
    const potentialWin = gameState.accumulated_win * gameState.bet_amount;
    document.getElementById('potential_win').textContent = potentialWin.toFixed(2);
    
    // Update game history
    this.updateGameHistory(gameState.game_history);
  }

  updateGameHistory(history) {
    const historyContainer = document.getElementById('game_history');
    historyContainer.innerHTML = '';
    
    history.forEach(card => {
      const cardElement = document.createElement('div');
      cardElement.className = 'history-card';
      cardElement.textContent = card.card;
      historyContainer.appendChild(cardElement);
    });
  }

  updateSkipsRemaining(skipsRemaining) {
    const skipButton = document.getElementById('skip_button');
    
    // Enhanced skip logic - show different text for first card
    if (this.gameState.position === 0) {
      skipButton.textContent = `SKIP (∞ infinite)`;
      skipButton.disabled = false;
    } else {
      skipButton.textContent = `SKIP (${skipsRemaining} left)`;
      skipButton.disabled = skipsRemaining <= 0;
    }
  }

  showGuessResult(result) {
    const resultDiv = document.getElementById('guess_result');
    resultDiv.innerHTML = `
      <p>Next Card: ${result.next_card}</p>
      <p>Result: ${result.success ? 'WIN' : 'LOSE'}</p>
      <p>Multiplier: ${result.payout_multiple.toFixed(2)}x</p>
      ${result.forced ? '<p class="forced">⚠️ House Intervention</p>' : ''}
    `;
    
    // Show animation or visual feedback
    if (result.success) {
      this.showWinAnimation();
    } else {
      this.showLoseAnimation();
    }
  }

  showCashOutResult(finalWin) {
    const resultDiv = document.getElementById('final_result');
    resultDiv.innerHTML = `
      <h2>Cashed Out!</h2>
      <p>Final Win: ${finalWin.toFixed(2)}</p>
    `;
  }

  showDeckVerification(deck, deckHash) {
    const verificationDiv = document.getElementById('deck_verification');
    verificationDiv.innerHTML = `
      <h3>Deck Verification</h3>
      <p>Hash: ${deckHash}</p>
      <div class="deck-cards">
        ${deck.map(card => `<span class="card">${card}</span>`).join('')}
      </div>
    `;
  }

  disableAllButtons() {
    ['higher_button', 'lower_button', 'same_button', 
     'higher_or_same_button', 'lower_or_same_button', 
     'skip_button', 'cashout_button'].forEach(id => {
      const button = document.getElementById(id);
      if (button) button.disabled = true;
    });
  }

  showWinAnimation() {
    // Add your win animation logic here
    document.body.classList.add('win-animation');
    setTimeout(() => document.body.classList.remove('win-animation'), 2000);
  }

  showLoseAnimation() {
    // Add your lose animation logic here
    document.body.classList.add('lose-animation');
    setTimeout(() => document.body.classList.remove('lose-animation'), 2000);
  }
}

// Usage Example
const game = new HiloGame('http://localhost:11400');

// Initialize game with Unity card
document.getElementById('start_button').onclick = async () => {
  const betAmount = parseFloat(document.getElementById('bet_amount').value);
  const card = document.getElementById('unity_card').value; // Optional Unity card input
  
  if (card) {
    console.log(`🎮 Starting game with Unity card: ${card}`);
    await game.startGame('demo_client', 'hilo_001', 'player_123', betAmount, card);
  } else {
    console.log('🎲 Starting game with random card');
    await game.startGame('demo_client', 'hilo_001', 'player_123', betAmount);
  }
};

// Skip button (now with enhanced logic)
document.getElementById('skip_button').onclick = () => game.skipCard();

// Cash out button  
document.getElementById('cashout_button').onclick = () => game.cashOut();

// Verify deck button
document.getElementById('verify_button').onclick = () => {
  if (game.gameState && game.gameState.seed) {
    game.verifyDeck(game.gameState.seed);
  }
};
```

## Key Implementation Benefits

### 1. **Unity Integration**
- Unity can specify starting cards
- Fallback to random generation if no card provided
- Consistent card format and validation
- Clear logging of card source
- **Fixed "Invalid game state" errors** with proper deck generation
- **Deck consistency** between StartGame and Guess handlers
- **Unity card tracking** in game state for verification
- **Enhanced debugging** with detailed logging for troubleshooting

### 2. **Decimal Precision Control**
- All monetary values rounded to exactly 2 decimal places
- Eliminates floating-point precision errors in JSON
- Professional financial display
- Consistent formatting across all endpoints

### 3. **Enhanced Skip Logic**
- First card: Infinite skips (no counting)
- Other cards: 5-skip limit with proper counting
- Smart UI updates showing skip status
- Improved user experience

### 4. **Perfect UI Match**
- Exact multiplier calculations from your screenshots
- Button states match game logic perfectly
- Simple string-based bet choices
- Pre-calculated options in every response

### 5. **Business Protection**
- RNG controls every potential payout
- House can force losses when needed
- Complete audit trail of decisions with IP/User-Agent
- Player-specific RTP control

### 6. **Performance & Scalability**
- Pre-calculated multipliers (no runtime math)
- Stateless design for horizontal scaling
- Efficient RNG integration
- Minimal database requirements

### 7. **Maintainability**
- Clean separation of concerns
- Single source of truth for calculations
- Easy to test individual components
- Clear API contracts

## RNG Integration Details

The system integrates with your existing RNG service to maintain house edge:

1. **Calculate Potential Win**: `accumulated_win * multiplier * bet_amount`
2. **RNG Decision**: Ask if house can afford this payout
3. **Force Outcome**: Override natural result if needed
4. **Log Decision**: Record all interventions with IP and User-Agent for audit

This ensures:
- ✅ Players see fair, mathematical odds
- ✅ House maintains complete payout control
- ✅ System appears transparent and fair
- ✅ Business profitability is protected
- ✅ Complete audit trail with client information

## Recent Fixes & Improvements

### ✅ **Fixed "Invalid Game State" Error**
- **Problem**: Unity cards caused deck mismatch between StartGame and Guess handlers
- **Solution**: Added `UnityCard` field to track Unity-specified cards
- **Result**: Proper deck generation for both Unity and random cards

### ✅ **Enhanced Card Value Parsing**
- **Problem**: `GetCardValue()` couldn't parse `"RANK_SUIT"` format
- **Solution**: Updated to use `strings.Split()` for proper parsing
- **Result**: Correct card values for all betting calculations

### ✅ **Improved Debug Logging**
- **Added**: Detailed logging for deck generation process
- **Added**: Unity card tracking in game state
- **Added**: Verification step logging for troubleshooting
- **Result**: Easy debugging and monitoring of Unity integration

## Testing

Run tests with:
```bash
go test ./pkg/games/hilo/...
```

## Production Deployment

1. **Set Production Environment Variables**
2. **Configure RNG and Settings Service URLs**
3. **Set Up Logging and Monitoring**
4. **Deploy with Process Manager (PM2, systemd, etc.)**

## Support

For questions or issues:
- Check server logs for detailed error information
- Ensure all required fields are included in requests
- Verify HMAC signatures are properly generated
- Contact the development team for integration assistance

---

**This implementation provides the exact functionality shown in your screenshots while maintaining complete house control through RNG integration, with enhanced Unity support, decimal precision control, and fixed "Invalid game state" errors.**