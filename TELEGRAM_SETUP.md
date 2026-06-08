# Telegram Error Notification Setup

This document explains how to set up Telegram error notifications for the OnePiece game when both the Settings API and RNG API fail.

## Environment Variables

Add the following environment variables to your `.env` file or set them in your deployment environment:

```bash
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=7775154815:AAH1miw3u9spNOGzqRRLYPvOyXDKnhLhomw
TELEGRAM_CHAT_ID=1099100013
```

## How It Works

1. **Error Detection**: The OnePiece game handlers now monitor both the Settings API and RNG API calls
2. **Dual Failure Detection**: When both APIs fail, a Telegram notification is automatically sent
3. **Notification Content**: The notification includes:
   - Game ID (onepiece)
   - Client ID, Player ID, Bet ID
   - Timestamp
   - Specific error messages from both APIs
   - Alert status indicating immediate attention is required

## Implementation Details

### Files Modified/Created:

1. **`pkg/common/telegram/client.go`** - New Telegram service client
2. **`pkg/common/config/env.go`** - Added Telegram configuration
3. **`pkg/games/onepiece/routes.go`** - Updated to include Telegram client
4. **`pkg/games/onepiece/handlers.go`** - Modified SpinHandler to detect dual API failures
5. **`main.go`** - Updated to initialize Telegram client

### Error Flow:

```
Settings API Call → Success/Failure
     ↓
RNG API Call → Success/Failure (only if Settings succeeded)
     ↓
If both failed → Send Telegram notification
     ↓
Return appropriate error response to client
```

## Testing

To test the Telegram integration, you can run the test script:

```bash
# Set environment variables
export TELEGRAM_BOT_TOKEN="7775154815:AAH1miw3u9spNOGzqRRLYPvOyXDKnhLhomw"
export TELEGRAM_CHAT_ID="1099100013"

# Run the test
go run test_telegram.go
```

## Message Format

The Telegram notification will look like this:

```
🚨 API Failure Alert 🚨

Game: onepiece
Client ID: client_123
Player ID: player_456
Bet ID: bet_789
Timestamp: 2024-01-15 14:30:25 UTC

Errors:
• Settings API: connection timeout
• RNG API: 500 internal server error

Status: Both critical APIs failed - immediate attention required!
```

## Security Notes

- The Telegram bot token and chat ID are loaded from environment variables
- The notification includes game context but no sensitive player data
- Failed Telegram notifications are logged but don't affect the main error response

## Future Enhancements

- Add rate limiting to prevent spam notifications
- Include more detailed error context
- Add retry logic for failed Telegram notifications
- Extend to other games in the system
