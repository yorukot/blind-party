import { WoolColor } from '$lib/constants/blockTextures';

// ============================================================================
// Core Data Models
// ============================================================================

export interface Game {
    game_id: string;
    created_at: string; // ISO 8601
    started_at?: string; // ISO 8601
    ended_at?: string; // ISO 8601
    phase: 'pre-game' | 'in-game' | 'settlement';
    current_round?: Round;
    rounds: Round[];
    map: number[][]; // 20x20 grid of WoolColor IDs
    players: APIPlayer[];
    player_count: number;
    alive_count: number;
    config: GameConfig;
}

export interface APIPlayer {
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

export interface PlayerStats {
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

export interface Round {
    round_number: number;
    phase: 'color-call' | 'rush-phase' | 'elimination-check' | 'round-transition';
    countdown_time: number;
    start_time: string; // ISO 8601
    end_time?: string; // ISO 8601
    color_to_show: WoolColor; // WoolColor ID
    rush_duration: number;
    eliminated_count: number;
}

export interface GameConfig {
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

// ============================================================================
// HTTP API Types
// ============================================================================

export interface CreateGameResponse {
    game_id: string;
}

// ============================================================================
// WebSocket Message Types - Client to Server
// ============================================================================

export interface PlayerUpdateMessage {
    type: 'player_update';
    data: {
        pos_x: number;
        pos_y: number;
    };
}

export interface PingMessage {
    type: 'ping';
}

export type ClientMessage = PlayerUpdateMessage | PingMessage;

// ============================================================================
// WebSocket Message Types - Server to Client
// ============================================================================

export interface GameStateMessage {
    type: 'game_state';
    data: Game;
}

export interface PlayerJoinedMessage {
    type: 'player_joined';
    data: {
        player: APIPlayer;
        player_count: number;
    };
}

export interface PreparationStartedMessage {
    type: 'preparation_started';
    data: {
        game_id: string;
        players: APIPlayer[];
        preparation_time: number;
    };
}

export interface GameStartedMessage {
    type: 'game_started';
    data: {
        game_id: string;
        players: APIPlayer[];
        map: number[][];
        game_config: GameConfig;
    };
}

export interface ColorCalledMessage {
    type: 'color_called';
    data: {
        round_number: number;
        color_to_show: WoolColor;
        phase: 'color-call';
        phase_duration: number;
    };
}

export interface RushPhaseStartedMessage {
    type: 'rush_phase_started';
    data: {
        phase: 'rush-phase';
        rush_duration: number;
        round_number: number;
    };
}

export interface RushTimerUpdateMessage {
    type: 'rush_timer_update';
    data: {
        remaining_time: number;
        round_number: number;
    };
}

export interface EliminationCheckStartedMessage {
    type: 'elimination_check_started';
    data: {
        phase: 'elimination-check';
        round_number: number;
    };
}

export interface PlayersEliminatedMessage {
    type: 'players_eliminated';
    data: {
        eliminated_players: APIPlayer[];
        remaining_count: number;
        round_number: number;
    };
}

export interface RoundResultsMessage {
    type: 'round_results';
    data: {
        phase: 'round-transition';
        round_number: number;
        eliminated_count: number;
        remaining_count: number;
    };
}

export interface RoundFinishedMessage {
    type: 'round_finished';
    data: {
        round_number: number;
        eliminated_count: number;
        remaining_count: number;
        next_round_in: number;
    };
}

export interface GameEndedMessage {
    type: 'game_ended';
    data: {
        game_id: string;
        total_rounds: number;
        duration: number;
        players: APIPlayer[];
    };
}

export interface SettlementStartedMessage {
    type: 'settlement_started';
    data: {
        game_id: string;
        settlement_duration: number;
        final_leaderboard: APIPlayer[];
    };
}

export interface FinalResultsMessage {
    type: 'final_results';
    data: {
        game_id: string;
        total_rounds: number;
        duration: number;
        leaderboard: APIPlayer[];
        game_stats: {
            total_players: number;
            rounds_played: number;
            average_survival: number;
            winner: APIPlayer;
            longest_survival: number;
        };
    };
}

export interface GameCleanupMessage {
    type: 'game_cleanup';
    data: {
        game_id: string;
        reason: string;
    };
}

export interface MovementRejectedMessage {
    type: 'movement_rejected';
    data: {
        reason: string;
        speed?: number;
        max_speed?: number;
        reset_position: {
            pos_x: number;
            pos_y: number;
        };
        message: string;
    };
}

export interface PlayerPositionsUpdateMessage {
    type: 'player_positions_update';
    data: {
        players: {
            user_id: string;
            name: string;
            pos_x: number;
            pos_y: number;
            is_spectator: boolean;
        }[];
        round_number: number;
        timestamp: number;
    };
}

export interface PongMessage {
    type: 'pong';
}

export type ServerMessage =
    | GameStateMessage
    | PlayerJoinedMessage
    | PreparationStartedMessage
    | GameStartedMessage
    | ColorCalledMessage
    | RushPhaseStartedMessage
    | RushTimerUpdateMessage
    | EliminationCheckStartedMessage
    | PlayersEliminatedMessage
    | RoundResultsMessage
    | RoundFinishedMessage
    | GameEndedMessage
    | SettlementStartedMessage
    | FinalResultsMessage
    | GameCleanupMessage
    | MovementRejectedMessage
    | PlayerPositionsUpdateMessage
    | PongMessage;

// ============================================================================
// WebSocket Connection States
// ============================================================================

export type WebSocketConnectionState =
    | 'disconnected'
    | 'connecting'
    | 'connected'
    | 'reconnecting'
    | 'error';

// ============================================================================
// API Client Configuration
// ============================================================================

export interface APIClientConfig {
    apiBaseUrl: string;
    wsBaseUrl: string;
    reconnectAttempts?: number;
    reconnectDelay?: number;
    pingInterval?: number;
}