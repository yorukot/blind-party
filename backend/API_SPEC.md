# Backend API Specification

This document outlines the API for the "Color Rush Survival" game backend.

## 1. HTTP API

### 1.1. Create a New Game

Creates a new game instance and returns a unique game ID.

-   **Endpoint:** `POST /api/game/`
-   **Request Body:** None
-   **Success Response (200 OK):**

    ```json
    {
      "game_id": "123456"
    }
    ```

## 2. WebSocket API

The primary communication for gameplay is handled via WebSockets.

### 2.1. Connection

-   **Endpoint:** `ws://<host>/api/game/{gameID}/ws?username={username}`
-   **Parameters:**
    -   `gameID` (string, required): The ID of the game to join, obtained from the "Create a New Game" endpoint.
    -   `username` (string, required): The display name for the player.

### 2.2. Coordinate System

The game uses a 20x20 block-based coordinate system:

-   **Map Size:** 20x20 blocks (400 total blocks)
-   **Coordinate Range:** 1.0 to 21.0 for both X and Y axes
-   **Block Centers:** Players spawn at block centers (1.5, 2.5, 3.5, ..., 20.5)
-   **Precision:** Maximum 2 decimal places (e.g., 1.25, 10.99, 15.33)
-   **Boundaries:** Players are eliminated if they move outside the 1.0-21.0 range

**Examples:**
-   Block (1,1): Player coordinates 1.0-2.0 (X) and 1.0-2.0 (Y)
-   Block (10,10): Player coordinates 10.0-11.0 (X) and 10.0-11.0 (Y)
-   Map center: (10.5, 10.5)

### 2.3. Client-to-Server Messages

Messages sent from the frontend client to the backend server.

#### `player_update`

Sent frequently to update the player's position on the map. The server validates the movement and will reject it if it's too fast or out of bounds.

-   **Type:** `player_update`
-   **Payload:**

    ```json
    {
      "type": "player_update",
      "data": {
        "pos_x": 10.5,
        "pos_y": 7.25
      }
    }
    ```

#### `ping`

Sent to keep the connection alive. The server will respond with a `pong` message.

-   **Type:** `ping`
-   **Payload:**
    ```json
    {
        "type": "ping"
    }
    ```

### 2.4. Server-to-Client Messages

Messages broadcast from the backend server to connected clients.

#### `game_state`

Sent to a newly connected client to provide the complete current state of the game.

-   **Type:** `game_state`
-   **Payload:** A full `Game` object. See [Data Models](#3-data-models) for the structure.

#### `player_joined`

Broadcast when a new player joins the game lobby.

-   **Type:** `player_joined`
-   **Payload:**
    ```json
    {
      "type": "player_joined",
      "data": {
        "player": { ...Player Object... },
        "player_count": 5
      }
    }
    ```

#### `preparation_started`

Broadcast when the game is about to start, initiating a 5-second countdown.

-   **Type:** `preparation_started`
-   **Payload:**
    ```json
    {
      "type": "preparation_started",
      "data": {
        "game_id": "123456",
        "players": [ ...Array of Player Objects... ],
        "preparation_time": 5
      }
    }
    ```

#### `game_started`

Broadcast when the game officially begins and the first round starts.

-   **Type:** `game_started`
-   **Payload:**
    ```json
    {
      "type": "game_started",
      "data": {
        "game_id": "123456",
        "players": [ ...Array of Player Objects... ],
        "map": [ ...2D Array of WoolColor IDs... ],
        "game_config": { ...GameConfig Object... }
      }
    }
    ```

#### `color_called`

Broadcast at the start of a round to announce the target color.

-   **Type:** `color_called`
-   **Payload:**
    ```json
    {
      "type": "color_called",
      "data": {
        "round_number": 1,
        "color_to_show": 14, // WoolColor ID (e.g., 14 is Red)
        "phase": "color-call",
        "phase_duration": 1.0
      }
    }
    ```

#### `rush_phase_started`

Broadcast after the `color_called` phase, indicating that players must now move to the correct color.

-   **Type:** `rush_phase_started`
-   **Payload:**
    ```json
    {
      "type": "rush_phase_started",
      "data": {
        "phase": "rush-phase",
        "rush_duration": 4.0,
        "round_number": 1
      }
    }
    ```

#### `rush_timer_update`

Broadcast periodically during the `rush-phase` to update the remaining time.

-   **Type:** `rush_timer_update`
-   **Payload:**
    ```json
    {
      "type": "rush_timer_update",
      "data": {
        "remaining_time": 3.25,
        "round_number": 1
      }
    }
    ```

#### `elimination_check_started`

Broadcast at the end of the rush phase, indicating that the server is now checking player positions.

-   **Type:** `elimination_check_started`
-   **Payload:**
    ```json
    {
      "type": "elimination_check_started",
      "data": {
        "phase": "elimination-check",
        "round_number": 1
      }
    }
    ```

#### `players_eliminated`

Broadcast if any players were eliminated during the round.

-   **Type:** `players_eliminated`
-   **Payload:**
    ```json
    {
      "type": "players_eliminated",
      "data": {
        "eliminated_players": [ ...Array of Player Objects... ],
        "remaining_count": 12,
        "round_number": 1
      }
    }
    ```

#### `round_results`

Broadcast after the elimination check, summarizing the round's outcome.

-   **Type:** `round_results`
-   **Payload:**
    ```json
    {
      "type": "round_results",
      "data": {
        "phase": "round-transition",
        "round_number": 1,
        "eliminated_count": 2,
        "remaining_count": 12
      }
    }
    ```

#### `round_finished`

Broadcast at the end of the round transition, before the next round begins.

-   **Type:** `round_finished`
-   **Payload:**
    ```json
    {
      "type": "round_finished",
      "data": {
        "round_number": 1,
        "eliminated_count": 2,
        "remaining_count": 12,
        "next_round_in": 3
      }
    }
    ```

#### `game_ended`

Broadcast when the game's win/loss conditions are met.

-   **Type:** `game_ended`
-   **Payload:**
    ```json
    {
      "type": "game_ended",
      "data": {
        "game_id": "123456",
        "total_rounds": 22,
        "duration": 185.7,
        "players": [ ...Array of Player Objects with final stats... ]
      }
    }
    ```

#### `settlement_started`

Broadcast after `game_ended` to transition to the final scoreboard/settlement screen.

-   **Type:** `settlement_started`
-   **Payload:**
    ```json
    {
      "type": "settlement_started",
      "data": {
        "game_id": "123456",
        "settlement_duration": 300, // 5 minutes in seconds
        "final_leaderboard": [ ...Array of Player Objects sorted by rank... ]
      }
    }
    ```

#### `final_results`

Broadcast periodically during the settlement phase with detailed game statistics.

-   **Type:** `final_results`
-   **Payload:**
    ```json
    {
      "type": "final_results",
      "data": {
        "game_id": "123456",
        "total_rounds": 22,
        "duration": 185.7,
        "leaderboard": [ ...Array of Player Objects sorted by rank... ],
        "game_stats": {
          "total_players": 16,
          "rounds_played": 22,
          "average_survival": 15.4,
          "winner": { ...Winner Object... },
          "longest_survival": 22
        }
      }
    }
    ```

#### `game_cleanup`

Broadcast at the very end of the settlement period, indicating the game instance is being destroyed. The client should disconnect after receiving this.

-   **Type:** `game_cleanup`
-   **Payload:**
    ```json
    {
      "type": "game_cleanup",
      "data": {
        "game_id": "123456",
        "reason": "settlement_completed"
      }
    }
    ```

#### `movement_rejected`

Sent to a specific client if their movement update was invalid. The client should reset their position to the one provided.

-   **Type:** `movement_rejected`
-   **Payload:**
    ```json
    {
      "type": "movement_rejected",
      "data": {
        "reason": "movement_too_fast",
        "speed": 8.5,
        "max_speed": 5.0,
        "reset_position": { "pos_x": 12.1, "pos_y": 8.4 },
        "message": "Position reset due to invalid movement"
      }
    }
    ```

#### `player_positions_update`

Broadcast periodically during the game to update all player positions. Sent at 10Hz (every 100ms) during active gameplay.

-   **Type:** `player_positions_update`
-   **Payload:**
    ```json
    {
      "type": "player_positions_update",
      "data": {
        "players": [
          {
            "user_id": "player_123",
            "name": "PlayerName",
            "pos_x": 10.5,
            "pos_y": 7.25,
            "is_spectator": false
          }
        ],
        "round_number": 1,
        "timestamp": 1693728000000
      }
    }
    ```

#### `pong`

Sent in response to a client's `ping` message.

-   **Type:** `pong`
-   **Payload:**
    ```json
    {
        "type": "pong"
    }
    ```

## 3. Data Models

Core data structures used in the WebSocket messages.

### `Game`

```typescript
interface Game {
  game_id: string;
  created_at: string; // ISO 8601
  started_at?: string; // ISO 8601
  ended_at?: string; // ISO 8601
  phase: 'pre-game' | 'in-game' | 'settlement';
  current_round?: Round;
  rounds: Round[];
  map: number[][]; // 20x20 grid of WoolColor IDs
  players: Player[];
  player_count: number;
  alive_count: number;
  config: GameConfig;
}
```

### `Player`

```typescript
interface Player {
  user_id: string;
  name: string;
  position: {
    pos_x: number;
    pos_y: number;
  };
  is_spectator: boolean;
  is_eliminated: boolean;
  joined_round: number;
  stats: PlayerStats;
}
```

### `PlayerStats`

```typescript
interface PlayerStats {
  rounds_survived: number;
  total_distance: number;
  eliminated_at?: string; // ISO 8601
  final_position: number;
  score: number;
  survival_points: number;
  elimination_bonus: number;
  speed_bonuses: number;
  streak_bonuses: number;
  current_streak: number;
  longest_streak: number;
  three_streak_count: number;
  five_streak_count: number;
  ten_streak_count: number;
  average_response_time: number;
  perfect_rounds: number;
}
```

### `Round`

```typescript
interface Round {
  round_number: number;
  phase: 'color-call' | 'rush-phase' | 'elimination-check' | 'round-transition';
  countdown_time: number;
  start_time: string; // ISO 8601
  end_time?: string; // ISO 8601
  color_to_show: number; // WoolColor ID
  rush_duration: number;
  eliminated_count: number;
}
```

### `GameConfig`

```typescript
interface GameConfig {
  map_width: number;
  map_height: number;
  countdown_sequence: number[];
  spectator_only_rounds: number;
  timing_progression: {
    start_round: number;
    end_round: number;
    duration: number;
  }[];
  survival_points_per_round: number;
  elimination_bonus_multiplier: number;
  speed_bonus_threshold: number;
  perfect_bonus_threshold: number;
  speed_bonus_points: number;
  perfect_bonus_points: number;
  final_winner_bonus: number;
  endurance_bonus: number;
  streak_bonuses: { [key: number]: number };
  base_movement_speed: number;
  max_movement_speed: number;
  lag_compensation_ms: number;
  position_update_hz: number;
  timer_update_hz: number;
}
```

### `WoolColor` Enum

A mapping of color names to their corresponding integer IDs.

```typescript
enum WoolColor {
  White = 0,
  Orange = 1,
  Magenta = 2,
  LightBlue = 3,
  Yellow = 4,
  Lime = 5,
  Pink = 6,
  Gray = 7,
  LightGray = 8,
  Cyan = 9,
  Purple = 10,
  Blue = 11,
  Brown = 12,
  Green = 13,
  Red = 14,
  Black = 15,
  Air = 16,
}
```
