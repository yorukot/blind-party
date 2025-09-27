# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the backend for "Color Rush Survival" (also known as "Blind Party"), a real-time multiplayer elimination game where players must navigate to correct colored tiles before time expires. The backend is built in Go using Chi router and WebSockets for real-time communication.

## Core Commands

### Development
```bash
go run cmd/main.go                    # Run the server locally
go build -o bin/server cmd/main.go    # Build binary
go mod download                       # Download dependencies
go mod tidy                          # Clean up dependencies
```

### Testing & Quality
```bash
go test ./...                        # Run all tests
go test -v ./...                     # Run tests with verbose output
go test ./internal/handler/game      # Run specific package tests
go vet ./...                         # Run Go vet (linting)
go fmt ./...                         # Format code
```

## Architecture Overview

### Project Structure
- `cmd/main.go` - Application entry point with router setup
- `internal/` - Private application code
  - `config/` - Environment configuration management
  - `handler/game/` - Game-specific HTTP and WebSocket handlers
  - `middleware/` - HTTP middleware (logging, etc.)
  - `router/` - Route definitions
  - `schema/` - Core data structures and game state
- `pkg/` - Reusable packages
  - `logger/` - Zap logger configuration
  - `response/` - Standardized HTTP response utilities

### Core Game Architecture

The game follows a phase-based architecture with three main phases:
1. **PreGame** - Player lobby and joining
2. **InGame** - Active gameplay with rounds
3. **Settlement** - End game statistics and cleanup

### Key Data Structures

- `Game` (`internal/schema/game.go:108`) - Central game state with player management, WebSocket clients, and round tracking
- `Player` (`internal/schema/game.go:56`) - Player state including position, stats, and elimination status
- `Round` (`internal/schema/game.go:78`) - Round-specific data with timing and elimination tracking
- `WebSocketClient` (`internal/schema/game.go:92`) - WebSocket connection management per client

### Game State Management

The game uses a centralized `GameHandler` that maintains a map of active games (`internal/handler/game/handler.go:6`). Each game runs independently with:

- **Real-time Updates**: WebSocket-based position updates and game state broadcasts
- **Phase Management**: Automatic progression through game phases with timer-based round management
- **Player Elimination**: Position validation at round end with configurable timing
- **Map Evolution**: Dynamic color removal and tile redistribution

### WebSocket Communication

Real-time communication is handled through WebSocket connections (`internal/router/game.go:21`):
- Each game maintains its own client registry
- Broadcast channels for game-wide updates
- Per-client send channels for individual messages

## Game Configuration

The game follows the detailed specifications in `game.md`, including:
- 256x256 tile map with 16 wool colors
- Progressive round timing (4.0s → 1.2s)
- Special round types (multi-color, fake-out, speed rounds)
- Dynamic map changes (color removal at specific rounds)

## Environment Variables

Configuration is managed through environment variables (`internal/config/env.go`):
- `PORT` - Server port (default: 8080)
- `APP_ENV` - Environment (dev/prod, default: prod)
- `DEBUG` - Debug mode (default: false)
- `APP_NAME` - Application name (default: stargo)

## Development Notes

- Uses Chi router for HTTP routing with WebSocket upgrade capability
- Zap for structured logging with middleware integration
- Swagger documentation available in dev mode at `/swagger/`
- Health check endpoint at `/health`
- Game API routes under `/api/game/`

## WebSocket Endpoints

- `POST /api/game/` - Create new game
- `WS /api/game/{gameID}/ws` - Connect to game WebSocket

## Key Dependencies

- `github.com/go-chi/chi/v5` - HTTP router
- `golang.org/x/net/websocket` - WebSocket support
- `go.uber.org/zap` - Structured logging
- `github.com/google/uuid` - UUID generation
- `github.com/caarlos0/env/v10` - Environment variable parsing