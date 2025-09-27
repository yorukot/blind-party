# In-Game Handler Specification

## Overview
The `in_game.go` module handles the core gameplay mechanics for the Color Rush Survival game, managing round-based elimination gameplay with real-time WebSocket communication.

## Core Functions

### 1. Random Color Generation
```go
func getRandomColor() schema.WoolColor
```
**Purpose**: Generates a random wool color from the 16 available colors (0-15)
**Returns**: Random `schema.WoolColor` value
**Colors Available**: White, Orange, Magenta, LightBlue, Yellow, Lime, Pink, Gray, LightGray, Cyan, Purple, Blue, Brown, Green, Red, Black

### 2. Map Generation
```go
func (h *GameHandler) generateRandomMap(game *schema.Game)
```
**Purpose**: Creates a new random map filled with colored blocks
**Parameters**:
- `game`: Game instance containing map dimensions and state
**Behavior**:
- Fills entire map (MapHeight × MapWidth) with random colors
- Uses `getRandomColor()` for each tile
- Logs map generation completion

### 3. Map Filtering
```go
func (h *GameHandler) removeNonTargetColors(game *schema.Game, targetColor schema.WoolColor)
```
**Purpose**: Removes all blocks except the target color, converting them to Air
**Parameters**:
- `game`: Game instance containing the map
- `targetColor`: The color to preserve
**Behavior**:
- Iterates through entire map
- Converts non-matching colors to `schema.Air`
- Logs color removal completion

### 4. Round Duration Calculation
```go
func (h *GameHandler) calculateRoundDuration(roundNumber int) float64
```
**Purpose**: Calculates round duration with exponential decay
**Parameters**:
- `roundNumber`: Current round number (1-based)
**Algorithm**:
- Base duration: 20.0 seconds
- Each round: 80% of previous round's duration
- Minimum duration: 1.2 seconds
**Returns**: Duration in seconds as float64
**Timing Progression**:
- Round 1: 20.0s
- Round 2: 16.0s
- Round 3: 12.8s
- Round 4: 10.24s
- Continue until minimum (1.2s)

### 5. Player Elimination
```go
func (h *GameHandler) eliminatePlayer(game *schema.Game, player *schema.Player)
```
**Purpose**: Handles player elimination with statistics tracking
**Parameters**:
- `game`: Game instance for context
- `player`: Player to eliminate
**Behavior**:
- Sets `player.IsEliminated = true`
- Records elimination timestamp
- Updates rounds survived count
- Calculates final position based on remaining alive players
- Prevents double elimination

### 6. Round Management
```go
func (h *GameHandler) startNewRound(game *schema.Game)
```
**Purpose**: Initializes and starts a new game round
**Process**:
1. Increment round number
2. Generate new random map
3. Select random target color
4. Calculate round duration
5. Create new Round object
6. Set countdown timer
7. Broadcast round start event

**WebSocket Event**: `game_update` with data:
- `round_number`: Current round
- `target_color`: Color players must reach
- `countdown`: Round duration in seconds
- `map`: 2D array representation of the map

### 7. Phase Handling
```go
func (h *GameHandler) handleInGamePhase(game *schema.Game)
```
**Purpose**: Main dispatcher for in-game phase management
**Behavior**: Routes to appropriate phase handler based on current round phase

### 8. Color Call Phase
```go
func (h *GameHandler) handleColorCallPhase(game *schema.Game)
```
**Purpose**: Manages the active gameplay phase where players navigate to target color
**Process**:
1. Update countdown timer based on elapsed time
2. Broadcast countdown updates via WebSocket
3. When countdown reaches 0:
   - Remove non-target colored blocks
   - Broadcast map update
   - Transition to elimination check phase

**WebSocket Events**:
- `game_update` (countdown): Contains `countdown_seconds` and `target_color`
- `game_update` (map change): Contains updated `map` and `blocks_removed: true`

### 9. Elimination Check Phase
```go
func (h *GameHandler) handleEliminationCheckPhase(game *schema.Game)
```
**Purpose**: Validates player positions and eliminates those on incorrect tiles
**Process**:
1. Check each non-eliminated player's position
2. Convert player coordinates to map indices
3. Validate position bounds
4. Check if player is on correct color block
5. Eliminate players on Air or wrong color
6. Broadcast elimination results
7. Check win condition or continue to next round

**Position Validation**:
- Player positions are 1-based (1.5, 2.5, etc.)
- Map coordinates are 0-based
- Conversion: `x = int(player.Position.X - 1)`
- Out-of-bounds players are eliminated
- Players on Air or wrong color blocks are eliminated

**WebSocket Events**:
- `game_update` (elimination): Contains `eliminated_players`, `round_number`, `target_color`
- `game_update` (game end): Contains `winner_id`, `end_time`, `total_rounds`, `alive_count`
- `game_update` (round end): Contains `round_number`, `alive_count`, `next_round_in`

### 10. Map Conversion Utility
```go
func (h *GameHandler) convertMapToArray(game *schema.Game) [][]int
```
**Purpose**: Converts internal map representation to JSON-serializable format
**Returns**: 2D integer array where each int represents a wool color value
**Usage**: Used in WebSocket broadcasts to send map state to clients

## Game Flow

### Round Lifecycle
1. **Round Start**: New map generated, target color selected, countdown initiated
2. **Color Call Phase**: Players navigate while countdown decreases
3. **Map Transformation**: Non-target blocks removed when countdown expires
4. **Elimination Check**: Player positions validated, eliminations processed
5. **Round End**: Statistics updated, next round scheduled or game ended

### Game End Conditions
- **Single Winner**: When only 1 player remains alive
- **No Winners**: When 0 players remain alive (all eliminated simultaneously)

### Inter-Round Timing
- 2-second break between rounds
- Next round starts automatically via goroutine

## WebSocket Communication

All game updates use the unified `game_update` event type with different data payloads:

### Event Data Structures
- **Round Start**: `round_number`, `target_color`, `countdown`, `map`
- **Countdown Update**: `countdown_seconds`, `target_color`
- **Map Update**: `map`, `blocks_removed`
- **Player Elimination**: `eliminated_players`, `round_number`, `target_color`
- **Game End**: `winner_id`, `end_time`, `total_rounds`, `alive_count`
- **Round End**: `round_number`, `alive_count`, `next_round_in`

## Dependencies

### Internal Packages
- `github.com/yorukot/blind-party/internal/schema`: Core data structures
- Standard library: `log`, `math/rand`, `time`

### Key Schema Types
- `schema.Game`: Central game state container
- `schema.Player`: Player state and statistics
- `schema.Round`: Round-specific data and timing
- `schema.WoolColor`: Enum for block colors (0-15)
- `schema.ColorCall`, `schema.EliminationCheck`: Round phase constants

## Configuration Constants

- **Base Round Duration**: 20.0 seconds
- **Duration Multiplier**: 0.8 (80% decay per round)
- **Minimum Duration**: 1.2 seconds
- **Inter-Round Delay**: 2.0 seconds
- **Available Colors**: 16 wool colors (0-15)

## Error Handling

- Bounds checking for player positions
- Nil checks for game state objects
- Elimination status validation to prevent double-elimination
- Countdown timer validation

## Logging

Comprehensive logging for:
- Map generation events
- Color removal operations
- Round state transitions
- Player eliminations
- Game completion events

All log messages include game ID for debugging and monitoring purposes.