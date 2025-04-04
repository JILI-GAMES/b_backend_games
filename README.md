# JILI Games Backend

This repository contains a modular, scalable backend for multiple casino games. The backend is designed to allow each game to function both independently and as part of a unified service.

## Project Structure

```
b_backend_games/
├── cmd/                           # Command-line applications
│   ├── gamesserver/               # Main server application
│   │   └── main.go                # Entry point that aggregates games
│   ├── superace_deluxe/           # Standalone superace deluxe game
│   │   └── main.go                # Entry point for standalone game
│   └── kong/              # Standalone treasure hunt game
│       └── main.go                # Entry point for standalone game
│
├── pkg/                           # Shared packages
│   ├── common/                    # Common functionality
│   │   ├── rng/                   # RNG client code
│   │   │   └── client.go
│   │   ├── settings/              # Settings client code
│   │   │   └── client.go
│   │   └── config/                # Configuration code
│   │       └── env.go
│   │
│   ├── games/                     # Individual game packages
│       ├── superace_deluxe/       # Superace Deluxe game package
│       │   ├── game.go            # Game logic
│       │   ├── handlers.go        # HTTP handlers
│       │   ├── routes.go          # Route definitions
│       │   ├── types.go           # Game-specific types
│       │   └── utils.go           # Helper functions
│       │
│       └── kong/          # Treasure Hunt game package
│           ├── game.go            # Game logic
│           ├── handlers.go        # HTTP handlers
│           ├── routes.go          # Route definitions
│           ├── types.go           # Game-specific types
│           └── utils.go           # Helper functions
```

## Features

- **Modular Design**: Each game is implemented as a self-contained package
- **Independence**: Games can run as standalone services or as part of the main server
- **Shared Components**: Common functionality like RNG and settings are centralized
- **Scalability**: New games can be added without modifying existing code
- **Environment Configuration**: Easy configuration via environment variables

## Endpoints

### Combined Server

- `POST /spin/superace/deluxe` - SuperAce Deluxe game spin endpoint
- `POST /spin/kong` - Treasure Hunt game spin endpoint
- `GET /status` - Server status showing available games

### Standalone Game Servers

Each game can be run as a standalone server with only its specific endpoints:

- SuperAce Deluxe:
  - `POST /spin/superace/deluxe`
  - `GET /status`

- Treasure Hunt (Kong):
  - `POST /spin/kong`
  - `GET /status`

## Getting Started

### Prerequisites

- Go 1.18 or higher
- Make (optional, for using Makefile commands)

### Environment Setup

Create a `.env` file in the root directory with the following variables:

```
RNG_API_URL=http://159.89.235.166:17003/api/proxy/rng/1
SETTINGS_API_URL=https://t2.ibibe.africa/get-game-settings
PORT=11400
LOG_FILE=app.log
```

### Running the Servers

#### Combined Server (All Games)

```bash
go run cmd/gamesserver/main.go
```

#### Individual Game Servers

SuperAce Deluxe:
```bash
go run cmd/superace_deluxe/main.go
```

Treasure Hunt:
```bash
go run cmd/treasurehunt/main.go
```

## Adding a New Game

1. Create a new package in `pkg/games/` with your game's name
2. Implement the game logic, handlers, routes, and types
3. Create a standalone main file in `cmd/` (optional)
4. Register the game's routes in the main server (`cmd/gamesserver/main.go`)

## External Services

This backend integrates with the following external services:

1. **RNG Service**: Provides randomized outcomes based on RTP
2. **Settings Service**: Provides game settings configuration

## Logging

All server activities are logged to the file specified in the `LOG_FILE` environment variable.