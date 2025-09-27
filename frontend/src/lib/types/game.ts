import type { PlayerSummary } from './player.js';
import { WoolColor } from '../constants/block-textures.js';

/**
 * Game phase enumeration representing the current state of the game
 */
export type GamePhase = 'pre-game' | 'in-game' | 'settlement';

/**
 * Player position coordinates in the game grid
 */
export interface GamePlayerPosition {
    pos_x: number;
    pos_y: number;
}

/**
 * Player object as received from the game state API
 */
export interface GamePlayer {
    name: string;
    position: GamePlayerPosition;
    is_spectator: boolean;
    is_eliminated: boolean;
}

/**
 * Main game state response structure from the API
 */
export interface GameStateResponse {
    game_id: string;
    phase: GamePhase;
    map: number[][];
    players: GamePlayer[];
    countdown_seconds: number | null;
    target_color: number;
}

/**
 * Extended game state with additional client-side properties
 */
export interface GameState extends GameStateResponse {
    /** Timestamp when the state was last updated */
    lastUpdated?: number;
    /** Whether this is the current player's game */
    isCurrentPlayerGame?: boolean;
}

/**
 * Game events that can occur during gameplay
 */
export type GameEvent =
    | { type: 'player_joined'; player: GamePlayer }
    | { type: 'player_left'; playerName: string }
    | { type: 'player_moved'; player: GamePlayer }
    | { type: 'player_eliminated'; playerName: string }
    | { type: 'phase_changed'; newPhase: GamePhase; countdown?: number }
    | { type: 'game_ended'; winners: string[] };

/**
 * Convert API GamePlayer to internal PlayerSummary format
 */
export function gamePlayerToPlayerSummary(gamePlayer: GamePlayer): PlayerSummary {
    return {
        id: gamePlayer.name, // Using name as ID since API doesn't provide separate ID
        name: gamePlayer.name,
        status: gamePlayer.is_spectator
            ? 'spectating'
            : gamePlayer.is_eliminated
              ? 'eliminated'
              : 'ingame',
        accent: '', // Default accent, can be set elsewhere
        position: {
            x: gamePlayer.position.pos_x,
            y: gamePlayer.position.pos_y
        }
    };
}

/**
 * Convert internal PlayerSummary to API GamePlayer format
 */
export function playerSummaryToGamePlayer(player: PlayerSummary): GamePlayer {
    return {
        name: player.name,
        position: {
            pos_x: player.position?.x ?? 0,
            pos_y: player.position?.y ?? 0
        },
        is_spectator: player.status === 'spectating',
        is_eliminated: player.status === 'eliminated'
    };
}

/**
 * Type guard to check if a phase is valid
 */
export function isValidGamePhase(phase: string): phase is GamePhase {
    return ['pre-game', 'in-game', 'settlement'].includes(phase);
}

/**
 * Get display name for game phase
 */
export function getPhaseDisplayName(phase: GamePhase): string {
    switch (phase) {
        case 'pre-game':
            return 'Waiting to Start';
        case 'in-game':
            return 'Playing';
        case 'settlement':
            return 'Game Over';
    }
}

/**
 * Get block name from wool color number
 */
export function getBlockName(woolColor: number): string {
    const colorName = WoolColor[woolColor];
    if (colorName === undefined) {
        return 'Unknown';
    }

    // Convert enum name to display format (e.g., "LightBlue" -> "Light Blue")
    return colorName.replace(/([A-Z])/g, ' $1').trim();
}


/**
 * Get CSS color value for wool color number
 */
export function getBlockColor(woolColor: number): string {
    const colorMap: Record<number, string> = {
        [WoolColor.White]: '#f8f8ff',
        [WoolColor.Orange]: '#ff8c00',
        [WoolColor.Magenta]: '#ff00ff',
        [WoolColor.LightBlue]: '#87ceeb',
        [WoolColor.Yellow]: '#ffff00',
        [WoolColor.Lime]: '#32cd32',
        [WoolColor.Pink]: '#ffc0cb',
        [WoolColor.Gray]: '#808080',
        [WoolColor.LightGray]: '#d3d3d3',
        [WoolColor.Cyan]: '#00ffff',
        [WoolColor.Purple]: '#800080',
        [WoolColor.Blue]: '#0000ff',
        [WoolColor.Brown]: '#a0522d',
        [WoolColor.Green]: '#008000',
        [WoolColor.Red]: '#ff0000',
        [WoolColor.Black]: '#000000',
        [WoolColor.Air]: '#transparent'
    };

    return colorMap[woolColor] || '#f8f8ff';
}
